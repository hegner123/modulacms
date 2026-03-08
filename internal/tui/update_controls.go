package tui

import (
	"github.com/hegner123/modulacms/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

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
