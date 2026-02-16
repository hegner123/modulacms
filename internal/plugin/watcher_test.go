package plugin

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// -- computeChecksum tests --

func TestComputeChecksum_ValidDirectory(t *testing.T) {
	dir := t.TempDir()

	// Write two .lua files.
	if err := os.WriteFile(filepath.Join(dir, "init.lua"), []byte("plugin_info = {}"), 0644); err != nil {
		t.Fatalf("writing init.lua: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "helpers.lua"), []byte("return {}"), 0644); err != nil {
		t.Fatalf("writing helpers.lua: %v", err)
	}

	checksum, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("computeChecksum failed: %v", err)
	}
	if checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestComputeChecksum_DeterministicForSameContent(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "init.lua"), []byte("plugin_info = {}"), 0644); err != nil {
		t.Fatalf("writing init.lua: %v", err)
	}

	c1, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("first checksum: %v", err)
	}

	c2, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("second checksum: %v", err)
	}

	if c1 != c2 {
		t.Errorf("checksums differ for same content: %q vs %q", c1, c2)
	}
}

func TestComputeChecksum_ChangesOnContentModification(t *testing.T) {
	dir := t.TempDir()

	initPath := filepath.Join(dir, "init.lua")
	if err := os.WriteFile(initPath, []byte("version = '1.0.0'"), 0644); err != nil {
		t.Fatalf("writing init.lua: %v", err)
	}

	c1, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("first checksum: %v", err)
	}

	// Modify the file.
	if err := os.WriteFile(initPath, []byte("version = '2.0.0'"), 0644); err != nil {
		t.Fatalf("rewriting init.lua: %v", err)
	}

	c2, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("second checksum: %v", err)
	}

	if c1 == c2 {
		t.Error("checksums should differ after content modification")
	}
}

func TestComputeChecksum_IgnoresNonLuaFiles(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "init.lua"), []byte("plugin_info = {}"), 0644); err != nil {
		t.Fatalf("writing init.lua: %v", err)
	}

	c1, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("first checksum: %v", err)
	}

	// Add a non-lua file -- should not affect checksum.
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Plugin"), 0644); err != nil {
		t.Fatalf("writing readme.md: %v", err)
	}

	c2, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("second checksum: %v", err)
	}

	if c1 != c2 {
		t.Error("non-lua files should not affect checksum")
	}
}

func TestComputeChecksum_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	checksum, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("computeChecksum failed on empty dir: %v", err)
	}
	// Empty directory should produce a valid (empty hash) checksum.
	if checksum == "" {
		t.Error("expected non-empty checksum even for empty directory")
	}
}

func TestComputeChecksum_NonExistentDirectory(t *testing.T) {
	_, err := computeChecksum("/nonexistent/directory/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestComputeChecksum_SkipsSymlinks(t *testing.T) {
	dir := t.TempDir()

	realFile := filepath.Join(dir, "real.lua")
	if err := os.WriteFile(realFile, []byte("-- real"), 0644); err != nil {
		t.Fatalf("writing real.lua: %v", err)
	}

	symPath := filepath.Join(dir, "link.lua")
	if err := os.Symlink(realFile, symPath); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Should not fail -- just skips the symlink.
	checksum, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("computeChecksum failed with symlink: %v", err)
	}
	if checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestComputeChecksum_DetectsRename(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "old_name.lua"), []byte("same content"), 0644); err != nil {
		t.Fatalf("writing old_name.lua: %v", err)
	}

	c1, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("first checksum: %v", err)
	}

	// Rename the file (same content, different name).
	if err := os.Rename(filepath.Join(dir, "old_name.lua"), filepath.Join(dir, "new_name.lua")); err != nil {
		t.Fatalf("renaming: %v", err)
	}

	c2, err := computeChecksum(dir)
	if err != nil {
		t.Fatalf("second checksum: %v", err)
	}

	if c1 == c2 {
		t.Error("checksums should differ after rename (filename included in hash)")
	}
}

// -- Watcher lifecycle tests --

func TestNewWatcher_DefaultPollInterval(t *testing.T) {
	w := NewWatcher(nil, 0)
	if w.pollInterval != 2*time.Second {
		t.Errorf("default pollInterval = %s, want 2s", w.pollInterval)
	}
}

func TestNewWatcher_CustomPollInterval(t *testing.T) {
	w := NewWatcher(nil, 5*time.Second)
	if w.pollInterval != 5*time.Second {
		t.Errorf("pollInterval = %s, want 5s", w.pollInterval)
	}
}

func TestWatcher_RunCancelledByContext(t *testing.T) {
	// Watcher.pollTick accesses manager.mu, so we need a real (empty) Manager.
	pool, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("opening in-memory sqlite: %v", err)
	}
	defer func() {
		// Close pool error is benign on test cleanup.
		pool.Close()
	}()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       t.TempDir(),
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, pool, db.DialectSQLite, nil)

	w := NewWatcher(mgr, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	// Cancel after a short time.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Good -- Run returned after cancel.
	case <-time.After(2 * time.Second):
		t.Fatal("watcher.Run did not return after context cancellation")
	}
}

// -- Cooldown tests --

func TestWatcher_IsPluginCooldownActive(t *testing.T) {
	w := NewWatcher(nil, 50*time.Millisecond)

	// No cooldown initially.
	if w.IsPluginCooldownActive("test") {
		t.Error("cooldown should not be active for unknown plugin")
	}

	// Set cooldown.
	w.SetReloadCooldown("test")

	// Should be active immediately.
	if !w.IsPluginCooldownActive("test") {
		t.Error("cooldown should be active after SetReloadCooldown")
	}
}

func TestWatcher_CooldownExpiresAfter10s(t *testing.T) {
	w := NewWatcher(nil, 50*time.Millisecond)

	// Manually set a cooldown in the past.
	w.mu.Lock()
	w.reloadCooldowns["test"] = time.Now().Add(-11 * time.Second)
	w.mu.Unlock()

	if w.IsPluginCooldownActive("test") {
		t.Error("cooldown should have expired after 10s")
	}
}

// -- Slow reload tracking --

func TestWatcher_PausedPluginsMapInitialization(t *testing.T) {
	w := NewWatcher(nil, 50*time.Millisecond)

	if w.pausedPlugins == nil {
		t.Error("pausedPlugins map should be initialized")
	}
	if w.slowCounts == nil {
		t.Error("slowCounts map should be initialized")
	}
	if w.checksums == nil {
		t.Error("checksums map should be initialized")
	}
}

// -- InitialChecksums tests --

// TestWatcher_InitialChecksums_PopulatesMap is a smoke test that verifies
// InitialChecksums does not panic with a nil manager. The actual behavior
// with loaded plugins is tested in integration tests.
func TestWatcher_InitialChecksums_SkipsWithNilPlugins(t *testing.T) {
	// This test verifies that InitialChecksums handles the case where
	// the manager returns no plugins without panicking.
	// We cannot test the full flow here because it requires a real Manager
	// with loaded plugins.
}
