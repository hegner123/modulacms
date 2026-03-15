package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/deploy"
)

// =============================================================================
// DEPLOY DIALOG CONTEXTS
// =============================================================================

// DeployPullContext stores context for a pending deploy pull confirmation.
type DeployPullContext struct {
	EnvName string
}

// DeployPushContext stores context for a pending deploy push confirmation.
type DeployPushContext struct {
	EnvName string
}

// =============================================================================
// DEPLOY UPDATE HANDLERS
// =============================================================================

// UpdateDeployFetch handles fetch messages for deploy environments.
func (m Model) UpdateDeployFetch(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case DeployEnvsFetchMsg:
		if m.Config == nil {
			return m, nil
		}
		envs := m.Config.Deploy_Environments
		return m, tea.Batch(
			DeployEnvsSetCmd(envs),
			LoadingStopCmd(),
		)
	}
	return m, nil
}

// UpdateDeployCms routes deploy operation request and result messages.
func (m Model) UpdateDeployCms(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	// =========================================================================
	// TEST CONNECTION
	// =========================================================================
	case DeployTestConnectionRequestMsg:
		cfg := m.Config
		if cfg == nil {
			return m, nil
		}
		envName := msg.EnvName
		// Screen manages its own OperationActive/StatusMessage state
		return m, func() tea.Msg {
			ctx := context.Background()
			health, err := deploy.TestEnvConnection(ctx, *cfg, envName)
			if err != nil {
				return DeployTestConnectionResultMsg{
					Health: &DeployHealthResult{
						EnvName: envName,
						Err:     err.Error(),
					},
				}
			}
			return DeployTestConnectionResultMsg{
				Health: &DeployHealthResult{
					EnvName: envName,
					Status:  health.Status,
					Version: health.Version,
					NodeID:  health.NodeID,
				},
			}
		}

	case DeployTestConnectionResultMsg:
		// Screen manages display state; log the result at Model level
		var logMsg string
		if msg.Health.Err != "" {
			logMsg = fmt.Sprintf("Connection failed: %s", msg.Health.Err)
		} else {
			logMsg = fmt.Sprintf("Connected to %s (v%s)", msg.Health.EnvName, msg.Health.Version)
		}
		return m, LogMessageCmd(logMsg)

	// =========================================================================
	// PULL REQUEST → execute
	// =========================================================================
	case DeployPullRequestMsg:
		cfg := m.Config
		driver := m.DB
		if cfg == nil || driver == nil {
			return m, nil
		}
		envName := msg.EnvName
		dryRun := msg.DryRun
		// Screen manages its own OperationActive/StatusMessage state
		opLabel := "pull"
		if dryRun {
			opLabel = "dry-run pull"
		}
		return m, func() tea.Msg {
			ctx := context.Background()
			result, err := deploy.Pull(ctx, *cfg, driver, envName, deploy.ExportOptions{}, false, dryRun)
			if err != nil {
				return DeployPullResultMsg{
					Err: err.Error(),
					Result: &DeploySyncResult{
						Success:   false,
						DryRun:    dryRun,
						Operation: opLabel,
						EnvName:   envName,
					},
				}
			}
			return DeployPullResultMsg{
				Result: syncResultToDisplay(result, opLabel, envName),
			}
		}

	case DeployPullResultMsg:
		// Screen manages display state; log the result at Model level
		var logMsg string
		if msg.Err != "" {
			logMsg = fmt.Sprintf("Pull failed: %s", msg.Err)
		} else {
			logMsg = fmt.Sprintf("Pull completed: %d tables", len(msg.Result.TablesAffected))
		}
		return m, LogMessageCmd(logMsg)

	// =========================================================================
	// PUSH REQUEST → execute
	// =========================================================================
	case DeployPushRequestMsg:
		cfg := m.Config
		driver := m.DB
		if cfg == nil || driver == nil {
			return m, nil
		}
		envName := msg.EnvName
		dryRun := msg.DryRun
		// Screen manages its own OperationActive/StatusMessage state
		opLabel := "push"
		if dryRun {
			opLabel = "dry-run push"
		}
		return m, func() tea.Msg {
			ctx := context.Background()
			result, err := deploy.Push(ctx, *cfg, driver, envName, deploy.ExportOptions{}, dryRun)
			if err != nil {
				return DeployPushResultMsg{
					Err: err.Error(),
					Result: &DeploySyncResult{
						Success:   false,
						DryRun:    dryRun,
						Operation: opLabel,
						EnvName:   envName,
					},
				}
			}
			return DeployPushResultMsg{
				Result: syncResultToDisplay(result, opLabel, envName),
			}
		}

	case DeployPushResultMsg:
		// Screen manages display state; log the result at Model level
		var logMsg string
		if msg.Err != "" {
			logMsg = fmt.Sprintf("Push failed: %s", msg.Err)
		} else {
			logMsg = fmt.Sprintf("Push completed: %d tables", len(msg.Result.TablesAffected))
		}
		return m, LogMessageCmd(logMsg)
	}

	return m, nil
}

// syncResultToDisplay converts a deploy.SyncResult to the TUI display type.
func syncResultToDisplay(r *deploy.SyncResult, operation, envName string) *DeploySyncResult {
	var errs []string
	for _, e := range r.Errors {
		errs = append(errs, fmt.Sprintf("[%s/%s] %s", e.Table, e.Phase, e.Message))
	}
	return &DeploySyncResult{
		Success:        r.Success,
		DryRun:         r.DryRun,
		Operation:      operation,
		EnvName:        envName,
		TablesAffected: r.TablesAffected,
		RowCounts:      r.RowCounts,
		Duration:       r.Duration,
		Warnings:       r.Warnings,
		Errors:         errs,
	}
}
