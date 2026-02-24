package cli

import (
	"fmt"
	"strings"
)

// renderDeployEnvsList renders the deploy environments list for the left panel.
func renderDeployEnvsList(m Model) string {
	if len(m.DeployEnvironments) == 0 {
		return "(no environments configured)\n\nAdd deploy_environments\nto config.json"
	}

	lines := make([]string, 0, len(m.DeployEnvironments))
	for i, env := range m.DeployEnvironments {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, env.Name))
	}
	return strings.Join(lines, "\n")
}

// renderDeployDetail renders the selected environment details and last operation
// result for the center panel.
func renderDeployDetail(m Model) string {
	if len(m.DeployEnvironments) == 0 || m.Cursor >= len(m.DeployEnvironments) {
		return "No environment selected"
	}

	env := m.DeployEnvironments[m.Cursor]

	apiKeyDisplay := "(not set)"
	if env.APIKey != "" {
		apiKeyDisplay = "********"
	}

	lines := []string{
		fmt.Sprintf("Name:    %s", env.Name),
		fmt.Sprintf("URL:     %s", env.URL),
		fmt.Sprintf("API Key: %s", apiKeyDisplay),
	}

	// Show health check result if available and matches this env
	if m.DeployLastHealth != nil && m.DeployLastHealth.EnvName == env.Name {
		h := m.DeployLastHealth
		lines = append(lines, "")
		if h.Err != "" {
			lines = append(lines, "Health: FAILED")
			lines = append(lines, fmt.Sprintf("  Error: %s", h.Err))
		} else {
			lines = append(lines, fmt.Sprintf("Health:  %s", h.Status))
			lines = append(lines, fmt.Sprintf("Version: %s", h.Version))
			lines = append(lines, fmt.Sprintf("Node:    %s", h.NodeID))
		}
	}

	// Show last operation result if available and matches this env
	if m.DeployLastResult != nil && m.DeployLastResult.EnvName == env.Name {
		r := m.DeployLastResult
		lines = append(lines, "")
		lines = append(lines, "--- Last Operation ---")

		status := "SUCCESS"
		if !r.Success {
			status = "FAILED"
		}
		if r.DryRun {
			status += " (dry run)"
		}

		lines = append(lines, fmt.Sprintf("  %s: %s", r.Operation, status))
		if r.Duration != "" {
			lines = append(lines, fmt.Sprintf("  Duration: %s", r.Duration))
		}
		lines = append(lines, fmt.Sprintf("  Tables:   %d", len(r.TablesAffected)))

		totalRows := 0
		for _, count := range r.RowCounts {
			totalRows += count
		}
		lines = append(lines, fmt.Sprintf("  Rows:     %d", totalRows))

		for _, w := range r.Warnings {
			lines = append(lines, fmt.Sprintf("  WARN: %s", w))
		}
		for _, e := range r.Errors {
			lines = append(lines, fmt.Sprintf("  ERR:  %s", e))
		}
	}

	// Show status message (errors, progress)
	if m.DeployStatusMessage != "" {
		lines = append(lines, "")
		lines = append(lines, m.DeployStatusMessage)
	}

	if m.DeployOperationActive {
		lines = append(lines, "")
		lines = append(lines, "  Operation in progress...")
	}

	return strings.Join(lines, "\n")
}

// renderDeployActions renders available actions for the right panel.
func renderDeployActions(m Model) string {
	lines := []string{
		"Actions",
		"",
		"  t: Test Connection",
		"  p: Pull (remote -> local)",
		"  s: Push (local -> remote)",
		"",
		"  Dry Run:",
		"  P: Dry Run Pull",
		"  S: Dry Run Push",
		"",
		fmt.Sprintf("Environments: %d", len(m.DeployEnvironments)),
	}

	if m.DeployOperationActive {
		lines = append(lines, "")
		lines = append(lines, "  (operation running)")
	}

	return strings.Join(lines, "\n")
}
