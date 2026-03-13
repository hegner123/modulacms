package tui

import (
	"embed"
	"encoding/json"

	tea "charm.land/bubbletea/v2"
	config "github.com/hegner123/modulacms/internal/config"
)

func (m Model) View() tea.View {
	var content string
	// Show user provisioning form if needed
	if m.NeedsProvisioning {
		if m.FormState != nil && m.FormState.Form != nil {
			content = m.FormState.Form.View()
		} else {
			content = "Initializing user provisioning..."
		}
	} else {
		content = renderCMSPanelLayout(m)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// Rendering utilities

// TitleFile embeds title graphics from the titles directory.
//
//go:embed titles
var TitleFile embed.FS

// formatJSON marshals config to formatted JSON.
func formatJSON(b *config.Config) (string, error) {
	formatted, err := json.MarshalIndent(*b, "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}
