package cli

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/config"
)

type ReadSingleRow struct {
	Index int
	Key   string
	Value string
}

func ViewPageMenus(m Model) string {
	out := strings.Builder{}
	for _, item := range m.PageMenu {
		out.WriteString(" " + item.Label + " ")

	}
	return out.String()
}

func (m Model) View() string {
	// Show user provisioning form if needed
	if m.NeedsProvisioning {
		if m.FormState != nil && m.FormState.Form != nil {
			return m.FormState.Form.View()
		}
		return "Initializing user provisioning..."
	}

	var ui string
	if m.Loading {
		str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", m.Spinner.View())
		return str
	}

	if isCMSPanelPage(m.Page.Index) {
		return renderCMSPanelLayout(m)
	}

	switch m.Page.Index {
	case HOMEPAGE:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Home")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())

		ui = p.Render(m)
	case DATABASEPAGE:
		menu := make([]string, 0, len(m.Tables))
		menu = append(menu, m.Tables...)
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Tables")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case BUCKETPAGE:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case OAUTHPAGE:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case CONFIGPAGE:
		docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
		body := lipgloss.JoinVertical(
			lipgloss.Left,
			RenderTitle(m.Titles[m.TitleFont]),
			m.headerView(),
			m.Viewport.View(),
			m.footerView(),
		)
		controls := RenderFooter("↑↓/pgup/pgdn:Scroll │ h/backspace:Back │ q:Quit")
		h := m.RenderSpace(docStyle.Render(body) + controls)
		ui = lipgloss.JoinVertical(
			lipgloss.Left,
			docStyle.Render(body),
			h,
			controls,
			m.RenderStatusBar(),
		)
	case TABLEPAGE:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case CREATEPAGE:
		p := NewFormPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Create")
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case READPAGE:
		p := NewTablePage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader(fmt.Sprintf("Read %s", m.TableState.Table))
		if !m.Loading {
			p.AddHeaders(m.TableState.Headers)
			p.AddRows(m.TableState.Rows)
			p.AddStatus(m.RenderStatusBar())
		}

		ui = p.Render(m)
	case READSINGLEPAGE:
		columns := make([]ReadSingleRow, 0, len(m.TableState.Headers))
		for i, v := range m.TableState.Headers {
			c := ReadSingleRow{
				Index: i,
				Key:   v,
				Value: "",
			}
			columns = append(columns, c)
		}
		for i := range columns {
			// Check if the row has enough columns to avoid panic
			if i < len(m.TableState.Rows[m.Cursor]) {
				columns[i].Value = m.TableState.Rows[m.Cursor][i]
			} else {
				columns[i].Value = "" // Empty value if column doesn't exist
			}
		}
		formatted := make([]string, 0)
		for _, v := range columns {
			row := fmt.Sprintf("%s: %s", v.Key, v.Value)
			formatted = append(formatted, row)
		}
		content := lipgloss.JoinVertical(lipgloss.Left, formatted...)
		p := NewStaticPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader(fmt.Sprintf("Read %s Row %d", m.TableState.Table, m.Cursor))
		p.AddBody(content)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case UPDATEPAGE:
		p := NewTablePage()
		if !m.Loading {
			p.AddHeaders(m.TableState.Headers)
			p.AddRows(m.TableState.Rows)
			p.AddStatus(m.RenderStatusBar())
		}
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case UPDATEFORMPAGE:
		p := NewFormPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Update FORM")
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case DELETEPAGE:
		p := NewTablePage()
		if !m.Loading {
			p.AddHeaders(m.TableState.Headers)
			p.AddRows(m.TableState.Rows)
			p.AddStatus(m.RenderStatusBar())
		}
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Delete")
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case DATATYPES:
		p := NewFormPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case DEVELOPMENT:
		p := NewStaticPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case DYNAMICPAGE:
		p := NewCMSPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Dynamic")
		p.AddControls("q quit")
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case ACTIONSPAGE:
		menu := ActionsMenuLabels()
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Actions")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	default:
		ui = m.RenderUI()
	}
	return ui
}

// Rendering utilities

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
	formatted, err := json.MarshalIndent(*b, "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}
