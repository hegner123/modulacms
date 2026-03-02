package tui

import "github.com/hegner123/modulacms/internal/config"

// DeploySyncResult holds the result of a pull/push operation for display in the TUI.
type DeploySyncResult struct {
	Success        bool
	DryRun         bool
	Operation      string // "pull" or "push"
	EnvName        string
	TablesAffected []string
	RowCounts      map[string]int
	Duration       string
	Warnings       []string
	Errors         []string
}

// DeployHealthResult holds the result of a health check for display in the TUI.
type DeployHealthResult struct {
	EnvName string
	Status  string
	Version string
	NodeID  string
	Err     string
}

// =============================================================================
// DEPLOY FETCH MESSAGES
// =============================================================================

// DeployEnvsFetchMsg requests loading environments from config.
type DeployEnvsFetchMsg struct{}

// DeployEnvsSet sets the deploy environments list in model state.
type DeployEnvsSet struct {
	Envs []config.DeployEnvironmentConfig
}

// =============================================================================
// DEPLOY OPERATION REQUEST MESSAGES
// =============================================================================

// DeployTestConnectionRequestMsg requests a health check against a remote environment.
type DeployTestConnectionRequestMsg struct {
	EnvName string
}

// DeployPullRequestMsg requests pulling data from a remote environment.
type DeployPullRequestMsg struct {
	EnvName string
	DryRun  bool
}

// DeployPushRequestMsg requests pushing data to a remote environment.
type DeployPushRequestMsg struct {
	EnvName string
	DryRun  bool
}

// =============================================================================
// DEPLOY OPERATION RESULT MESSAGES
// =============================================================================

// DeployTestConnectionResultMsg returns the health check result.
type DeployTestConnectionResultMsg struct {
	Health *DeployHealthResult
}

// DeployPullResultMsg returns the pull operation result.
type DeployPullResultMsg struct {
	Result *DeploySyncResult
	Err    string
}

// DeployPushResultMsg returns the push operation result.
type DeployPushResultMsg struct {
	Result *DeploySyncResult
	Err    string
}

// =============================================================================
// DEPLOY CONFIRMATION MESSAGES
// =============================================================================

// DeployConfirmPullMsg requests confirmation before pulling (overwrites local data).
type DeployConfirmPullMsg struct {
	EnvName string
}

// DeployConfirmPushMsg requests confirmation before pushing (overwrites remote data).
type DeployConfirmPushMsg struct {
	EnvName string
}
