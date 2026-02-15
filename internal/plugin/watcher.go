package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hegner123/modulacms/internal/utility"
)

// Watcher limits for security (S4).
const (
	maxLuaFilesPerPlugin = 100          // max .lua files per plugin directory
	maxBytesPerChecksum  = 10 << 20     // 10 MB total bytes per checksumming pass
	maxSlowReloads       = 3            // consecutive slow reloads before pausing watcher
	slowReloadThreshold  = 10 * time.Second
)

// pendingReload tracks a detected change awaiting debounce stability.
type pendingReload struct {
	checksum  string
	firstSeen time.Time
}

// Watcher performs file-polling hot reload with debounce and blue-green restart.
type Watcher struct {
	manager       *Manager
	pollInterval  time.Duration // default 2s
	debounceDelay time.Duration // default 1s -- wait for file stability before reload

	mu             sync.Mutex
	reloadMu       sync.Mutex // serialize reloads (one at a time)
	checksums      map[string]string         // pluginName -> SHA-256 of all .lua files
	pendingReloads map[string]pendingReload   // pluginName -> checksum + first seen time
	slowCounts     map[string]int             // pluginName -> consecutive slow reload count
	pausedPlugins  map[string]bool            // pluginName -> true if watcher paused for this plugin
	reloadCooldowns map[string]time.Time      // pluginName -> last reload time (S9: 10s cooldown)
	logger         *utility.Logger
}

// NewWatcher creates a new file-polling watcher for the given manager.
func NewWatcher(manager *Manager, pollInterval time.Duration) *Watcher {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	return &Watcher{
		manager:         manager,
		pollInterval:    pollInterval,
		debounceDelay:   1 * time.Second,
		checksums:       make(map[string]string),
		pendingReloads:  make(map[string]pendingReload),
		slowCounts:      make(map[string]int),
		pausedPlugins:   make(map[string]bool),
		reloadCooldowns: make(map[string]time.Time),
		logger:          utility.DefaultLogger,
	}
}

// InitialChecksums computes baseline checksums for all loaded plugins.
// Must be called after Manager.LoadAll().
func (w *Watcher) InitialChecksums() {
	w.mu.Lock()
	defer w.mu.Unlock()

	plugins := w.manager.ListPlugins()
	for _, inst := range plugins {
		checksum, err := computeChecksum(inst.Dir)
		if err != nil {
			w.logger.Warn(
				fmt.Sprintf("watcher: initial checksum failed for plugin %q: %s", inst.Info.Name, err.Error()),
				nil,
			)
			continue
		}
		w.checksums[inst.Info.Name] = checksum
	}
}

// Run is the blocking poll loop. Cancel via ctx.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.pollTick(ctx)
		}
	}
}

// pollTick performs one poll iteration: detect changes, manage debounce, trigger reloads.
func (w *Watcher) pollTick(ctx context.Context) {
	// Snapshot VM availability for metrics.
	w.manager.mu.RLock()
	plugins := make(map[string]*PluginInstance, len(w.manager.plugins))
	for k, v := range w.manager.plugins {
		plugins[k] = v
	}
	w.manager.mu.RUnlock()

	SnapshotVMAvailability(plugins)

	// Detect changes for each plugin.
	w.mu.Lock()

	for name, inst := range plugins {
		// Skip plugins paused due to consecutive slow reloads (S3).
		if w.pausedPlugins[name] {
			continue
		}

		checksum, err := computeChecksum(inst.Dir)
		if err != nil {
			w.logger.Warn(
				fmt.Sprintf("watcher: checksum error for plugin %q: %s", name, err.Error()),
				nil,
			)
			continue
		}

		baseline, hasBaseline := w.checksums[name]
		if !hasBaseline {
			// New plugin appeared since initial checksums (should not normally happen).
			w.checksums[name] = checksum
			continue
		}

		if checksum == baseline {
			// No change -- clear any pending reload for this plugin.
			delete(w.pendingReloads, name)
			continue
		}

		// Change detected. Check if we already have a pending reload.
		pending, hasPending := w.pendingReloads[name]
		if !hasPending {
			// First detection of this change.
			w.pendingReloads[name] = pendingReload{
				checksum:  checksum,
				firstSeen: time.Now(),
			}
			continue
		}

		if pending.checksum != checksum {
			// Checksum changed again (mid-save) -- reset the debounce timer.
			w.pendingReloads[name] = pendingReload{
				checksum:  checksum,
				firstSeen: time.Now(),
			}
			continue
		}

		// Same checksum as pending -- check debounce.
		if time.Since(pending.firstSeen) < w.debounceDelay {
			// Still within debounce window, wait more.
			continue
		}

		// Debounce satisfied -- trigger reload.
		delete(w.pendingReloads, name)
		w.checksums[name] = checksum

		// Check cooldown (S9).
		if lastReload, ok := w.reloadCooldowns[name]; ok {
			if time.Since(lastReload) < 10*time.Second {
				w.logger.Warn(
					fmt.Sprintf("watcher: skipping reload for plugin %q (within 10s cooldown)", name),
					nil,
				)
				continue
			}
		}

		w.mu.Unlock()
		w.triggerReload(ctx, name)
		w.mu.Lock()
	}

	w.mu.Unlock()
}

// triggerReload attempts a blue-green reload with try-lock and timing.
func (w *Watcher) triggerReload(ctx context.Context, pluginName string) {
	// S3: Try-lock reloadMu (non-blocking). If already held, skip.
	acquired := w.tryLockReload()
	if !acquired {
		w.logger.Warn(
			fmt.Sprintf("watcher: reload already in progress for another plugin, skipping %q", pluginName),
			nil,
		)
		return
	}
	defer w.reloadMu.Unlock()

	start := time.Now()

	w.logger.Info(
		fmt.Sprintf("watcher: reloading plugin %q", pluginName),
	)

	err := w.manager.ReloadPlugin(ctx, pluginName)
	duration := time.Since(start)

	w.mu.Lock()
	w.reloadCooldowns[pluginName] = time.Now()

	if err != nil {
		w.logger.Warn(
			fmt.Sprintf("watcher: reload failed for plugin %q: %s (old version still running)",
				pluginName, err.Error()),
			nil,
		)
		w.mu.Unlock()
		return
	}

	// S3: Track consecutive slow reloads.
	if duration > slowReloadThreshold {
		w.slowCounts[pluginName]++
		if w.slowCounts[pluginName] >= maxSlowReloads {
			w.pausedPlugins[pluginName] = true
			w.logger.Warn(
				fmt.Sprintf("watcher: pausing file polling for plugin %q (%d consecutive slow reloads > %s). Admin reload still available.",
					pluginName, w.slowCounts[pluginName], slowReloadThreshold),
				nil,
			)
		}
	} else {
		w.slowCounts[pluginName] = 0
	}

	w.mu.Unlock()

	w.logger.Info(
		fmt.Sprintf("watcher: plugin %q reloaded in %s", pluginName, duration),
	)
}

// tryLockReload attempts a non-blocking acquisition of reloadMu.
// Returns true if the lock was acquired, false if it was already held.
func (w *Watcher) tryLockReload() bool {
	// Go does not have a native TryLock on sync.Mutex prior to Go 1.18.
	// Since Go 1.18+ is our minimum: use TryLock().
	return w.reloadMu.TryLock()
}

// IsPluginCooldownActive returns whether the per-plugin reload cooldown is active.
// Used by the admin reload endpoint (S9).
func (w *Watcher) IsPluginCooldownActive(pluginName string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	lastReload, ok := w.reloadCooldowns[pluginName]
	if !ok {
		return false
	}
	return time.Since(lastReload) < 10*time.Second
}

// SetReloadCooldown records a reload time for cooldown tracking.
// Used by the admin reload handler.
func (w *Watcher) SetReloadCooldown(pluginName string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.reloadCooldowns[pluginName] = time.Now()
}

// computeChecksum computes a SHA-256 checksum of all .lua files in a plugin directory.
//
// S4: Uses os.Lstat (not Stat) and skips symlinks and non-regular files.
// Enforces max 100 .lua files per directory and max 10 MB total bytes per
// checksumming pass. Returns error if limits exceeded.
func computeChecksum(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("reading directory %q: %w", dir, err)
	}

	h := sha256.New()
	fileCount := 0
	totalBytes := int64(0)

	for _, entry := range entries {
		name := entry.Name()
		if len(name) < 5 || name[len(name)-4:] != ".lua" {
			continue
		}

		fullPath := filepath.Join(dir, name)

		// S4: Use Lstat to detect symlinks.
		info, statErr := os.Lstat(fullPath)
		if statErr != nil {
			return "", fmt.Errorf("lstat %q: %w", fullPath, statErr)
		}

		// Skip symlinks and non-regular files.
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			continue
		}

		fileCount++
		if fileCount > maxLuaFilesPerPlugin {
			return "", fmt.Errorf("plugin directory %q exceeds max %d .lua files", dir, maxLuaFilesPerPlugin)
		}

		totalBytes += info.Size()
		if totalBytes > maxBytesPerChecksum {
			return "", fmt.Errorf("plugin directory %q exceeds max %d bytes for checksumming", dir, maxBytesPerChecksum)
		}

		// Include the filename in the hash so renames are detected.
		h.Write([]byte(name))

		f, openErr := os.Open(fullPath)
		if openErr != nil {
			return "", fmt.Errorf("opening %q: %w", fullPath, openErr)
		}

		if _, copyErr := io.Copy(h, f); copyErr != nil {
			// Close file before returning error.
			f.Close()
			return "", fmt.Errorf("reading %q: %w", fullPath, copyErr)
		}

		if cerr := f.Close(); cerr != nil {
			return "", fmt.Errorf("closing %q: %w", fullPath, cerr)
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
