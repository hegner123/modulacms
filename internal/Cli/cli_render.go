package cli

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

func (m model) RenderUI() string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	physicalWidth, _, _ := term.GetSize(os.Stdout.Fd())

	docStyle = docStyle.MaxWidth(physicalWidth)

	doc := strings.Builder{}
	m.footer += "\n\nPress q to quit.\n"
	header := RenderHeading(m.header)

	doc.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		header,
        m.body,
		m.footer,
		m.RenderStatusTable(),
	    ),
	)

	return docStyle.Render(doc.String())
}
