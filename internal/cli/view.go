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
	case homePage.Index:
		menu := make([]string, 0, len(HomepageMenu))
		for _, v := range HomepageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Home")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())

		ui = p.Render(m)
	case selectTablePage.Index:
		menu := make([]string, 0, len(m.Tables))
		menu = append(menu, m.Tables...)
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Tables")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case cmsPage.Index:
		menu := make([]string, 0, len(CmsHomeMenu))
		for _, v := range CmsHomeMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
                p.AddHeader("CMS")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case adminCmsPage.Index:
		menu := make([]string, 0, len(CmsHomeMenu))
		for _, v := range CmsHomeMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
                p.AddHeader("Admin CMS")
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case bucketPage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case oauthPage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case configPage.Index:
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case tableActionsPage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddMenu(menu)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case readPage.Index:
		p := NewTablePage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader(fmt.Sprintf("Read %s", m.Table))
		if !m.Loading {
			p.AddHeaders(m.Headers)
			p.AddRows(m.Rows)
			p.AddStatus(m.RenderStatusBar())
		}

		ui = p.Render(m)
	case readSinglePage.Index:
		columns := make([]ReadSingleRow, 0, len(m.Headers))
		for i, v := range m.Headers {
			c := ReadSingleRow{
				Index: i,
				Key:   v,
				Value: "",
			}
			columns = append(columns, c)

		}
		//(m.Headers)
		//m.Rows[m.Cursor][i]
		for i := range columns {
			columns[i].Value = m.Rows[m.Cursor][i]
		}
		formatted := make([]string, 0)
		for _, v := range columns {
			row := fmt.Sprintf("%s: %s", v.Key, v.Value)
			formatted = append(formatted, row)
		}
		content := lipgloss.JoinVertical(lipgloss.Left, formatted...)
		p := NewStaticPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader(fmt.Sprintf("Read %s Row %d", m.Table, m.Cursor))
		p.AddBody(content)
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case updatePage.Index:
		p := NewTablePage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case updateFormPage.Index:
		p := NewFormPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case deletePage.Index:
		p := NewTablePage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case createPage.Index:
		p := NewFormPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case developmentPage.Index:
		p := NewStaticPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	case dynamicPage.Index:
		p := NewCMSPage()
		p.AddTitle(m.Titles[m.TitleFont])
		p.AddHeader("Dynamic")
		p.AddControls("q quit")
		p.AddStatus(m.RenderStatusBar())
		ui = p.Render(m)
	default:
		ui = m.RenderUI()
	}
	return ui
}
