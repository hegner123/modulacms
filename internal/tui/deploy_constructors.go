package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
)

// DeployEnvsFetchCmd creates a command to fetch deploy environments from config.
func DeployEnvsFetchCmd() tea.Cmd {
	return func() tea.Msg { return DeployEnvsFetchMsg{} }
}

// DeployEnvsSetCmd creates a command to set the deploy environments list.
func DeployEnvsSetCmd(envs []config.DeployEnvironmentConfig) tea.Cmd {
	return func() tea.Msg { return DeployEnvsSet{Envs: envs} }
}

// DeployTestConnectionCmd creates a command to test a remote environment connection.
func DeployTestConnectionCmd(envName string) tea.Cmd {
	return func() tea.Msg { return DeployTestConnectionRequestMsg{EnvName: envName} }
}

// DeployPullCmd creates a command to pull data from a remote environment.
func DeployPullCmd(envName string, dryRun bool) tea.Cmd {
	return func() tea.Msg { return DeployPullRequestMsg{EnvName: envName, DryRun: dryRun} }
}

// DeployPushCmd creates a command to push data to a remote environment.
func DeployPushCmd(envName string, dryRun bool) tea.Cmd {
	return func() tea.Msg { return DeployPushRequestMsg{EnvName: envName, DryRun: dryRun} }
}

// ShowDeployConfirmPullCmd shows a confirmation dialog for pulling from a remote environment.
func ShowDeployConfirmPullCmd(envName string) tea.Cmd {
	return func() tea.Msg { return DeployConfirmPullMsg{EnvName: envName} }
}

// ShowDeployConfirmPushCmd shows a confirmation dialog for pushing to a remote environment.
func ShowDeployConfirmPushCmd(envName string) tea.Cmd {
	return func() tea.Msg { return DeployConfirmPushMsg{EnvName: envName} }
}
