package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	var ui string
	if m.Loading {
		str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", m.Spinner.View())
		return str
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
	case CMSPAGE:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("CMS")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case ADMINCMSPAGE:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Admin CMS")
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
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
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
			columns[i].Value = m.TableState.Rows[m.Cursor][i]
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
	case CONTENT:
		menu := make([]string, 0, len(m.DatatypeMenu))
		menu = append(menu, m.DatatypeMenu...)
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Content")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case MEDIA:
		p := NewStaticPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Content")
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	default:
		ui = m.RenderUI()
	}
	return ui
}
