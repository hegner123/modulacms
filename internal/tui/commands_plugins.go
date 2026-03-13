package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/plugin"
)

// HandlePluginAction dispatches a plugin action request to the appropriate handler.
func (m Model) HandlePluginAction(msg PluginActionRequestMsg) tea.Cmd {
	mgr := m.PluginManager
	adminUser := m.AdminUsername
	if mgr == nil {
		return func() tea.Msg {
			return PluginActionResultMsg{Title: "Error", Message: "Plugin manager not available"}
		}
	}

	switch msg.Action {
	case PluginActionEnable:
		return func() tea.Msg {
			if err := mgr.ActivatePlugin(context.Background(), msg.Name, adminUser); err != nil {
				return PluginActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to enable plugin %q: %v", msg.Name, err),
				}
			}
			return PluginActionCompleteMsg{Name: msg.Name, Action: PluginActionEnable}
		}
	case PluginActionDisable:
		return func() tea.Msg {
			if err := mgr.DeactivatePlugin(context.Background(), msg.Name); err != nil {
				return PluginActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to disable plugin %q: %v", msg.Name, err),
				}
			}
			return PluginActionCompleteMsg{Name: msg.Name, Action: PluginActionDisable}
		}
	case PluginActionReload:
		return func() tea.Msg {
			if err := mgr.ReloadPlugin(context.Background(), msg.Name); err != nil {
				return PluginActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to reload plugin %q: %v", msg.Name, err),
				}
			}
			return PluginActionCompleteMsg{Name: msg.Name, Action: PluginActionReload}
		}
	case PluginActionApproveRoutes:
		return func() tea.Msg {
			bridge := mgr.Bridge()
			if bridge == nil {
				return PluginActionResultMsg{Title: "Error", Message: "HTTP bridge not available"}
			}
			routes := bridge.ListRoutes()
			approved := 0
			for _, r := range routes {
				if r.PluginName != msg.Name || r.Approved {
					continue
				}
				if err := bridge.ApproveRoute(context.Background(), r.PluginName, r.Method, r.Path, adminUser); err != nil {
					return PluginActionResultMsg{
						Title:   "Error",
						Message: fmt.Sprintf("Failed to approve route %s %s: %v", r.Method, r.Path, err),
					}
				}
				approved++
			}
			return PluginRoutesApprovedMsg{Name: msg.Name, Count: approved}
		}
	case PluginActionApproveHooks:
		return func() tea.Msg {
			engine := mgr.HookEngine()
			if engine == nil {
				return PluginActionResultMsg{Title: "Error", Message: "Hook engine not available"}
			}
			hooks := engine.ListHooks()
			approved := 0
			for _, h := range hooks {
				if h.PluginName != msg.Name || h.Approved {
					continue
				}
				if err := engine.ApproveHook(context.Background(), h.PluginName, h.Event, h.Table, adminUser); err != nil {
					return PluginActionResultMsg{
						Title:   "Error",
						Message: fmt.Sprintf("Failed to approve hook %s:%s: %v", h.Event, h.Table, err),
					}
				}
				approved++
			}
			return PluginHooksApprovedMsg{Name: msg.Name, Count: approved}
		}
	default:
		return nil
	}
}

// FetchPendingRoutesForApprovalScreenCmd is a free-function variant for Screen
// implementations that don't have access to Model.
func FetchPendingRoutesForApprovalScreenCmd(mgr *plugin.Manager, pluginName string) tea.Cmd {
	if mgr == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Plugin manager not available"}
		}
	}
	return func() tea.Msg {
		bridge := mgr.Bridge()
		if bridge == nil {
			return ActionResultMsg{Title: "Error", Message: "HTTP bridge not available"}
		}
		allRoutes := bridge.ListRoutes()
		var pending []string
		for _, r := range allRoutes {
			if !r.Approved && r.PluginName == pluginName {
				pending = append(pending, fmt.Sprintf("%s %s", r.Method, r.Path))
			}
		}
		if len(pending) == 0 {
			return ActionResultMsg{
				Title:   "No Pending Routes",
				Message: fmt.Sprintf("Plugin '%s' has no unapproved routes.", pluginName),
			}
		}
		return ShowApproveAllRoutesDialogMsg{PluginName: pluginName, PendingRoutes: pending}
	}
}

// FetchPendingHooksForApprovalScreenCmd is a free-function variant for Screen
// implementations that don't have access to Model.
func FetchPendingHooksForApprovalScreenCmd(mgr *plugin.Manager, pluginName string) tea.Cmd {
	if mgr == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Plugin manager not available"}
		}
	}
	return func() tea.Msg {
		engine := mgr.HookEngine()
		if engine == nil {
			return ActionResultMsg{Title: "Error", Message: "Hook engine not available"}
		}
		allHooks := engine.ListHooks()
		var pending []string
		for _, h := range allHooks {
			if !h.Approved && h.PluginName == pluginName {
				pending = append(pending, fmt.Sprintf("%s.%s", h.Event, h.Table))
			}
		}
		if len(pending) == 0 {
			return ActionResultMsg{
				Title:   "No Pending Hooks",
				Message: fmt.Sprintf("Plugin '%s' has no unapproved hooks.", pluginName),
			}
		}
		return ShowApproveAllHooksDialogMsg{PluginName: pluginName, PendingHooks: pending}
	}
}

// FetchPendingRoutesForApprovalCmd fetches unapproved routes for a plugin and shows a confirmation dialog.
func (m Model) FetchPendingRoutesForApprovalCmd(pluginName string) tea.Cmd {
	return FetchPendingRoutesForApprovalScreenCmd(m.PluginManager, pluginName)
}

// FetchPendingHooksForApprovalCmd fetches unapproved hooks for a plugin and shows a confirmation dialog.
func (m Model) FetchPendingHooksForApprovalCmd(pluginName string) tea.Cmd {
	return FetchPendingHooksForApprovalScreenCmd(m.PluginManager, pluginName)
}
