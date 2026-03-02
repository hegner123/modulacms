package tui

import (
	"embed"
	"encoding/json"

	"github.com/charmbracelet/lipgloss"
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

	if isCMSPanelPage(m.Page.Index) {
		return renderCMSPanelLayout(m)
	}

	// Fallback for any unrecognized page index
	return m.renderFallback()
}

// renderFallback renders a minimal UI with status bar and dialog overlay
// for any page not handled by the panel layout.
func (m Model) renderFallback() string {
	docStyle := lipgloss.NewStyle().Width(m.Width).Height(m.Height)
	doc := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Padding(0, 2).Render(),
		m.RenderStatusBar(),
	)
	ui := docStyle.Render(doc)

	if m.ActiveOverlay != nil {
		return RenderOverlay(ui, m.ActiveOverlay, m.Width, m.Height)
	}
	return ui
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
