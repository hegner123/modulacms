package cli

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/config"
)

// ReadSingleRow represents a single row displayed in read single page view.
type ReadSingleRow struct {
	Index int
	Key   string
	Value string
}

// ViewPageMenus renders the page menu bar from model menu items.
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
	case CONFIGPAGE:
		menu := ConfigCategoryMenuInit()
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case CONFIGCATEGORYPAGE:
		ui = m.renderConfigCategoryPage()
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
		if m.Loading {
			p.AddBody(fmt.Sprintf("\n   %s Loading...\n", m.Spinner.View()))
		} else {
			p.AddHeaders(m.TableState.Headers)
			p.AddRows(m.TableState.Rows)
		}
		p.AddStatus(m.RenderStatusBar())
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
		if m.Loading {
			p.AddBody(fmt.Sprintf("\n   %s Loading...\n", m.Spinner.View()))
		} else {
			p.AddHeaders(m.TableState.Headers)
			p.AddRows(m.TableState.Rows)
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
		if m.Loading {
			p.AddBody(fmt.Sprintf("\n   %s Loading...\n", m.Spinner.View()))
		} else {
			p.AddHeaders(m.TableState.Headers)
			p.AddRows(m.TableState.Rows)
		}
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Delete")
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
	case DATATYPE:
		p := NewFormPage()
		p.AddTitle(m.Titles[m.TitleFont])
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
	case QUICKSTARTPAGE:
		menu := QuickstartMenuLabels()
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Quickstart")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case PLUGINDETAILPAGE:
		menu := []string{
			"Enable Plugin",
			"Disable Plugin",
			"Reload Plugin",
			"Approve Routes",
			"Approve Hooks",
		}
		header := "Plugin Detail"
		if m.SelectedPlugin != "" {
			header = fmt.Sprintf("Plugin: %s", m.SelectedPlugin)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader(header)
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	default:
		ui = m.RenderUI()
	}

	if m.FilePickerActive {
		return FilePickerOverlay(ui, m.FilePicker, m.Width, m.Height)
	}

	if m.DialogActive && m.Dialog != nil {
		return DialogOverlay(ui, *m.Dialog, m.Width, m.Height)
	}

	if m.FormDialogActive && m.FormDialog != nil {
		return FormDialogOverlay(ui, *m.FormDialog, m.Width, m.Height)
	}

	if m.ContentFormDialogActive && m.ContentFormDialog != nil {
		return ContentFormDialogOverlay(ui, *m.ContentFormDialog, m.Width, m.Height)
	}

	if m.UserFormDialogActive && m.UserFormDialog != nil {
		return UserFormDialogOverlay(ui, *m.UserFormDialog, m.Width, m.Height)
	}

	if m.DatabaseFormDialogActive && m.DatabaseFormDialog != nil {
		return DatabaseFormDialogOverlay(ui, *m.DatabaseFormDialog, m.Width, m.Height)
	}

	if m.UIConfigFormDialogActive && m.UIConfigFormDialog != nil {
		return UIConfigFormDialogOverlay(ui, *m.UIConfigFormDialog, m.Width, m.Height)
	}

	return ui
}

// Rendering utilities

// TitleFile embeds title graphics from the titles directory.
//go:embed titles
var TitleFile embed.FS

// RenderUI renders the main UI for the model.
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

// formatJSON marshals config to formatted JSON.
func formatJSON(b *config.Config) (string, error) {
	formatted, err := json.MarshalIndent(*b, "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

// renderConfigCategoryPage renders the config category detail page or raw JSON view.
func (m Model) renderConfigCategoryPage() string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)

	if m.ConfigCategory == "raw_json" {
		body := lipgloss.JoinVertical(
			lipgloss.Left,
			RenderTitle(m.Titles[m.TitleFont]),
			m.headerView(),
			m.Viewport.View(),
			m.footerView(),
		)
		controls := RenderFooter("↑↓/pgup/pgdn:Scroll │ h/backspace:Back │ q:Quit")
		h := m.RenderSpace(docStyle.Render(body) + controls)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			docStyle.Render(body),
			h,
			controls,
			m.RenderStatusBar(),
		)
	}

	// Category field list view
	title := config.CategoryLabel(m.ConfigCategory)
	var rows []string

	labelStyle := lipgloss.NewStyle().Width(25).Bold(true)
	valueStyle := lipgloss.NewStyle().Width(40)
	cursorStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	restartMark := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Render(" [restart]")

	for i, field := range m.ConfigCategoryFields {
		value := config.ConfigFieldString(*m.Config, field.JSONKey)
		if field.Sensitive && value != "" {
			value = "********"
		}

		label := labelStyle.Render(field.Label)
		val := valueStyle.Render(value)

		suffix := ""
		if !field.HotReloadable {
			suffix = restartMark
		}

		row := fmt.Sprintf("  %s  %s%s", label, val, suffix)
		if i == m.ConfigFieldCursor {
			row = cursorStyle.Render("> ") + lipgloss.NewStyle().Bold(true).Render(
				fmt.Sprintf("%s  %s%s", label, val, suffix),
			)
		}
		rows = append(rows, row)
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTitle(m.Titles[m.TitleFont]),
		"  "+lipgloss.NewStyle().Bold(true).Underline(true).Render(title),
		"",
		strings.Join(rows, "\n"),
	)

	controls := RenderFooter("↑↓:Navigate │ enter/e:Edit │ h/backspace:Back │ q:Quit")
	h := m.RenderSpace(docStyle.Render(body) + controls)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		docStyle.Render(body),
		h,
		controls,
		m.RenderStatusBar(),
	)
}
