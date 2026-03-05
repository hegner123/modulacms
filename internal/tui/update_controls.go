package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// PageSpecificMsgHandlers routes keyboard events to the appropriate page handler.
func (m Model) PageSpecificMsgHandlers(cmd tea.Cmd, msg tea.Msg) (Model, tea.Cmd) {

	// Handle screen mode keys globally before page-specific dispatch.
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		km := m.Config.KeyBindings
		key := keyMsg.String()

		if km.Matches(key, config.ActionScreenToggle) {
			if m.ScreenMode == ScreenFull {
				m.ScreenMode = ScreenNormal
			} else {
				m.ScreenMode = ScreenFull
			}
			m.ScreenModeManual = true
			return m, nil
		}
		if km.Matches(key, config.ActionScreenNext) {
			m.ScreenMode = (m.ScreenMode + 1) % 3
			m.ScreenModeManual = true
			return m, nil
		}
		if km.Matches(key, config.ActionAccordion) {
			m.AccordionEnabled = !m.AccordionEnabled
			return m, nil
		}
		if km.Matches(key, config.ActionScreenReset) {
			m.ScreenMode = ScreenNormal
			m.ScreenModeManual = false
			return m, nil
		}
	}

	return m, nil
}

// ShowConfigFieldEditDialogCmd returns a command that shows the config field edit dialog.
func ShowConfigFieldEditDialogCmd(field config.FieldMeta, currentValue string) tea.Cmd {
	return func() tea.Msg {
		return ShowConfigFieldEditMsg{
			Field:        field,
			CurrentValue: currentValue,
		}
	}
}

// ShowConfigFieldEditMsg triggers the config field edit dialog.
type ShowConfigFieldEditMsg struct {
	Field        config.FieldMeta
	CurrentValue string
}
