package cli

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/config"
)

//go:embed titles
var TitleFile embed.FS

func (m Model) RenderUI() string {
	docStyle := lipgloss.NewStyle()
	docStyle = docStyle.Width(m.Width).Height(m.Height)

	doc := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Padding(0, 2).Render(),
		m.RenderStatusBar(),
	)

	renderedDoc := docStyle.Render(doc)

	// If dialog is active, render dialog over the UI
	if m.DialogActive && m.Dialog != nil {
		return DialogOverlay(renderedDoc, *m.Dialog, m.Width, m.Height)
	}

	return renderedDoc
}

func formatJSON(b *config.Config) (string, error) {
	formatted, err := json.MarshalIndent(*b, "", "    ")
	if err != nil {
		return "", err
	}
	nulled := strings.ReplaceAll(string(formatted), "\"\",", "null")
	trimmed := strings.ReplaceAll(nulled, "\"", "")
	return string(trimmed), nil
}
