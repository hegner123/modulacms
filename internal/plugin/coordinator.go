package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// pluginSnapshot holds the DB state for a single plugin at the last poll.
type pluginSnapshot struct {
	Status       types.PluginStatus
	DateModified string
}

// Coordinator polls the plugins DB table at a configurable interval and
// reconciles local in-memory plugin state with the shared database state.
// This enables multi-instance awareness: when Instance A enables/disables
// a plugin, Instance B detects the change on its next poll.
//
// The Coordinator uses manager.driver (not a separate driver field) to avoid
// two references to the same thing. The manager.driver is guaranteed non-nil
// when the Coordinator is created (checked by StartCoordinator).
type Coordinator struct {
	manager      *Manager
	pollInterval time.Duration
	lastSeen     map[string]pluginSnapshot
	logger       *utility.Logger
}

// NewCoordinator creates a new Coordinator tied to the given Manager.
func NewCoordinator(manager *Manager, pollInterval time.Duration) *Coordinator {
	return &Coordinator{
		manager:      manager,
		pollInterval: pollInterval,
		lastSeen:     make(map[string]pluginSnapshot),
		logger:       utility.DefaultLogger,
	}
}

// Run is the blocking poll loop. Cancel via ctx.
func (c *Coordinator) Run(ctx context.Context) {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.reconcile(ctx)
		}
	}
}

// seed populates the lastSeen map with the current DB state. Called once at
// startup before the poll loop begins. This prevents false-positive
// reconciliation on the first tick.
func (c *Coordinator) seed(ctx context.Context) error {
	plugins, err := c.manager.driver.ListPlugins()
	if err != nil {
		return fmt.Errorf("listing plugins for seed: %w", err)
	}
	if plugins == nil {
		return nil
	}

	for _, p := range *plugins {
		c.lastSeen[p.Name] = pluginSnapshot{
			Status:       p.Status,
			DateModified: p.DateModified.String(),
		}
	}

	c.logger.Info(fmt.Sprintf("coordinator: seeded with %d plugin(s)", len(*plugins)))
	return nil
}

// reconcile performs one poll cycle: fetch DB state and compare against lastSeen.
// Acts when a plugin's status has changed (activate/deactivate) or when date_modified
// changes without a status change (registry reload for SyncCapabilities on other instances).
func (c *Coordinator) reconcile(ctx context.Context) {
	plugins, err := c.manager.driver.ListPlugins()
	if err != nil {
		c.logger.Warn(
			fmt.Sprintf("coordinator: failed to list plugins: %s", err.Error()),
			nil,
		)
		return
	}

	if plugins == nil {
		return
	}

	// Track which DB plugins we've seen this cycle.
	dbNames := make(map[string]bool, len(*plugins))
	needsRegistryReload := false

	for _, p := range *plugins {
		dbNames[p.Name] = true

		prev, hasPrev := c.lastSeen[p.Name]
		currentDateModified := p.DateModified.String()

		// Update snapshot regardless.
		c.lastSeen[p.Name] = pluginSnapshot{
			Status:       p.Status,
			DateModified: currentDateModified,
		}

		// Skip the first time we see this plugin (seed should have captured it,
		// but handle late-discovered plugins gracefully).
		if !hasPrev {
			continue
		}

		// Check for date_modified change without status change (e.g., SyncCapabilities
		// on another instance). Triggers a pipeline registry reload.
		if prev.Status == p.Status {
			if prev.DateModified != currentDateModified {
				needsRegistryReload = true
			}
			continue
		}

		// Status changed — reconcile.
		localState, localExists := c.manager.GetPluginState(p.Name)

		switch p.Status {
		case types.PluginStatusEnabled:
			// DB says enabled, local might be stopped/failed.
			if localExists && localState != StateRunning {
				c.logger.Info(fmt.Sprintf("coordinator: activating plugin %q (DB status: enabled, local state: %s)", p.Name, localState))
				if activateErr := c.manager.ActivatePlugin(ctx, p.Name, "coordinator-sync"); activateErr != nil {
					c.logger.Warn(
						fmt.Sprintf("coordinator: failed to activate plugin %q: %s", p.Name, activateErr.Error()),
						nil,
					)
				}
			}

		case types.PluginStatusInstalled:
			// DB says installed (disabled), local might be running.
			if localExists && localState == StateRunning {
				c.logger.Info(fmt.Sprintf("coordinator: deactivating plugin %q (DB status: installed, local state: %s)", p.Name, localState))
				if deactivateErr := c.manager.DeactivatePlugin(ctx, p.Name); deactivateErr != nil {
					c.logger.Warn(
						fmt.Sprintf("coordinator: failed to deactivate plugin %q: %s", p.Name, deactivateErr.Error()),
						nil,
					)
				}
			}
		}
	}

	// Reload pipeline registry if any plugin's date_modified changed without status change.
	// This picks up capability/pipeline changes made by SyncCapabilities on other instances.
	if needsRegistryReload {
		if regErr := c.manager.LoadRegistry(ctx); regErr != nil {
			c.logger.Warn(
				fmt.Sprintf("coordinator: failed to reload pipeline registry: %s", regErr.Error()),
				nil,
			)
		} else {
			c.logger.Info("coordinator: pipeline registry reloaded (date_modified change detected)")
		}
	}

	// Check for locally running plugins that no longer exist in the DB.
	localPlugins := c.manager.ListPlugins()
	for _, inst := range localPlugins {
		if !dbNames[inst.Info.Name] {
			// Plugin exists locally but not in DB — evict it.
			if inst.State == StateRunning || inst.State == StateLoading {
				c.logger.Info(fmt.Sprintf("coordinator: evicting plugin %q (removed from database)", inst.Info.Name))
				if evictErr := c.manager.EvictPlugin(ctx, inst.Info.Name, "plugin removed from database"); evictErr != nil {
					c.logger.Warn(
						fmt.Sprintf("coordinator: failed to evict plugin %q: %s", inst.Info.Name, evictErr.Error()),
						nil,
					)
				}
			}
			// Remove from lastSeen since it's gone from DB.
			delete(c.lastSeen, inst.Info.Name)
		}
	}
}
