package tui

import (
	"embed"
	"encoding/json"

	config "github.com/hegner123/modulacms/internal/config"
)

func (m Model) View() string {
	// Show user provisioning form if needed
	if m.NeedsProvisioning {
		if m.FormState != nil && m.FormState.Form != nil {
			return m.FormState.Form.View()
		}
		return "Initializing user provisioning..."
	}

	return renderCMSPanelLayout(m)
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
