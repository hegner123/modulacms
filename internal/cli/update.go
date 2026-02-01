package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle user provisioning first if needed
	if m, cmd := m.UpdateProvisioning(msg); cmd != nil {
		return m, cmd
	}

	if m, cmd := m.UpdateLog(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateTea(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateState(msg); cmd != nil {
		return m, cmd
	}

	if m, cmd := m.UpdateNavigation(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateFetch(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateForm(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateDialog(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateDatabase(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateCms(msg); cmd != nil {
		return m, cmd
	}
	// When dialog is active, route all key input to the dialog and stop.
	if m.DialogActive && m.Dialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			dialog, cmd := m.Dialog.Update(keyMsg)
			m.Dialog = &dialog
			return m, cmd
		}
	}

	return m.PageSpecificMsgHandlers(nil, msg)
}