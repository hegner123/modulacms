package tui

// PluginAction represents the type of plugin operation.
type PluginAction int

const (
	PluginActionEnable PluginAction = iota
	PluginActionDisable
	PluginActionReload
	PluginActionApproveRoutes
	PluginActionApproveHooks
)

// PluginActionRequestMsg requests a plugin operation.
type PluginActionRequestMsg struct {
	Name   string
	Action PluginAction
}

// PluginActionCompleteMsg signals that a plugin action completed successfully.
type PluginActionCompleteMsg struct {
	Name   string
	Action PluginAction
}

// PluginActionResultMsg signals the result of a plugin action (enable/disable/reload/approve).
type PluginActionResultMsg struct {
	Title   string
	Message string
}

// PluginRoutesApprovedMsg signals that all plugin routes were approved.
type PluginRoutesApprovedMsg struct {
	Name  string
	Count int
}

// PluginHooksApprovedMsg signals that all plugin hooks were approved.
type PluginHooksApprovedMsg struct {
	Name  string
	Count int
}

// ShowApproveAllRoutesDialogMsg triggers the route approval confirmation dialog.
type ShowApproveAllRoutesDialogMsg struct {
	PluginName    string
	PendingRoutes []string // human-readable list for display
}

// ShowApproveAllHooksDialogMsg triggers the hook approval confirmation dialog.
type ShowApproveAllHooksDialogMsg struct {
	PluginName   string
	PendingHooks []string // human-readable list for display
}

// PluginSyncCapabilitiesRequestMsg requests a capability sync for a plugin.
type PluginSyncCapabilitiesRequestMsg struct {
	Name string
}

// PluginSyncCapabilitiesResultMsg carries the result of a capability sync.
type PluginSyncCapabilitiesResultMsg struct {
	Name string
	Err  error
}
