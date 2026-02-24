package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// DeployControls handles keyboard navigation for the deploy page.
func (m Model) DeployControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionNextPanel) {
			m.PanelFocus = (m.PanelFocus + 1) % 3
			return m, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			m.PanelFocus = (m.PanelFocus + 2) % 3
			return m, nil
		}
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				m.DeployLastHealth = nil
				m.DeployLastResult = nil
				m.DeployStatusMessage = ""
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.DeployEnvironments)-1 {
				m.DeployLastHealth = nil
				m.DeployLastResult = nil
				m.DeployStatusMessage = ""
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionTitlePrev) {
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		}
		if km.Matches(key, config.ActionTitleNext) {
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		}

		// Block actions while operation is running
		if m.DeployOperationActive {
			return m, nil
		}

		if len(m.DeployEnvironments) == 0 || m.Cursor >= len(m.DeployEnvironments) {
			return m, nil
		}

		env := m.DeployEnvironments[m.Cursor]

		switch key {
		case "t":
			return m, DeployTestConnectionCmd(env.Name)
		case "p":
			return m, ShowDeployConfirmPullCmd(env.Name)
		case "s":
			return m, ShowDeployConfirmPushCmd(env.Name)
		case "P":
			return m, DeployPullCmd(env.Name, true)
		case "S":
			return m, DeployPushCmd(env.Name, true)
		}
	}
	return m, nil
}
