package cli

import (
	"fmt"
	
	"github.com/charmbracelet/lipgloss"
)

type PageIndex int

type Page struct {
	Index      PageIndex
	Controller CliInterface
	Label      string
	Parent     *Page
	Children   []*Page
	Next       *Page
}

const (
	HOMEPAGE PageIndex = iota
	CMSPAGE
	DATABASEPAGE
	BUCKETPAGE
	OAUTHPAGE
	CONFIGPAGE
	TABLEPAGE
	CREATEPAGE
	READPAGE
	UPDATEPAGE
	DELETEPAGE
	UPDATEFORMPAGE
	READSINGLEPAGE
	DYNAMICPAGE
	DEFINEDATATYPE
	DEVELOPMENT
)

var (
	homePage            *Page = &Page{Index: HOMEPAGE, Controller: pageInterface, Label: "Home", Parent: nil}
	cmsPage             *Page = &Page{Index: CMSPAGE, Controller: pageInterface, Label: "CMS", Parent: homePage}
	selectTablePage     *Page = &Page{Index: DATABASEPAGE, Controller: tableInterface, Label: "Database", Parent: homePage}
	bucketPage          *Page = &Page{Index: BUCKETPAGE, Controller: contentInterface, Label: "Bucket", Parent: homePage}
	oauthPage           *Page = &Page{Index: OAUTHPAGE, Controller: contentInterface, Label: "Oauth", Parent: homePage}
	configPage          *Page = &Page{Index: CONFIGPAGE, Controller: configInterface, Label: "Config", Parent: homePage, Children: nil}
	tableActionsPage    *Page = &Page{Index: TABLEPAGE, Controller: pageInterface, Label: "Table Actions", Parent: selectTablePage, Children: nil}
	createPage          *Page = &Page{Index: CREATEPAGE, Controller: createInterface, Label: "Create", Parent: tableActionsPage, Children: nil}
	readPage            *Page = &Page{Index: READPAGE, Controller: readInterface, Label: "Read", Parent: tableActionsPage, Children: nil}
	updatePage          *Page = &Page{Index: UPDATEPAGE, Controller: updateInterface, Label: "Update", Parent: tableActionsPage, Children: nil}
	deletePage          *Page = &Page{Index: DELETEPAGE, Controller: deleteInterface, Label: "Delete", Parent: tableActionsPage, Children: nil}
	updateFormPage      *Page = &Page{Index: UPDATEFORMPAGE, Controller: updateFormInterface, Label: "UpdateForm", Parent: nil, Children: nil}
	readSinglePage      *Page = &Page{Index: READSINGLEPAGE, Controller: readSingleInterface, Label: "ReadSingle", Parent: nil, Children: nil}
	dynamicPage         *Page = &Page{Index: DYNAMICPAGE, Controller: pageInterface, Label: "Dynamic", Parent: nil, Children: nil}
	definedDatatypePage *Page = &Page{Index: DEFINEDATATYPE, Controller: pageInterface, Label: "DefineDatatype", Parent: nil, Children: nil}
	developmentPage     *Page = &Page{Index: DEVELOPMENT, Controller: developmentInterface, Label: "Development", Parent: nil, Children: nil}
)

func (m Model) View() string {
	var ui string
	if m.Loading {
		str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", m.Spinner.View())
		return str
	}
	switch m.Page.Index {
	case homePage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.Titles[m.TitleFont], "MAIN MENU", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case selectTablePage.Index:
		menu := make([]string, 0, len(m.Tables))
		menu = append(menu, m.Tables...)
		p := NewMenuPage(menu, m.Titles[m.TitleFont], "TABLES", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case cmsPage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.Titles[m.TitleFont], "CMS", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case bucketPage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.Titles[m.TitleFont], "Bucket", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case oauthPage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.Titles[m.TitleFont], "OAUTH", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case configPage.Index:
		p := NewMenuPage([]string{}, m.Titles[m.TitleFont], "CONFIG", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case tableActionsPage.Index:
		menu := make([]string, 0, len(m.PageMenu))
		for _, v := range m.PageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.Titles[m.TitleFont], m.Table, []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case readPage.Index:
		p := NewTablePage(m.Headers, m.Rows, m.Table, m.Titles[m.TitleFont], "READ", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case readSinglePage.Index:
		var row []Row
		value := make(map[string]string, len(m.Headers))
		for i, v := range m.Headers {
			value[v] = m.Rows[m.Cursor][i]
		}
		for k, v := range value {
			col := NewColumn(lipgloss.Left, k, v); r := NewRow(lipgloss.Left, col)
			row = append(row, r)
		}
		body := []Row{}
		p := NewStaticPage(m.Titles[m.TitleFont], m.Table, append(body, row...), "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case updatePage.Index:
		p := NewTablePage(m.Headers, m.Rows, m.Table, m.Titles[m.TitleFont], "UPDATE", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case updateFormPage.Index:
		p := NewFormPage(m.Titles[m.TitleFont], m.Table, []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case deletePage.Index:
		p := NewTablePage(m.Headers, m.Rows, m.Table, m.Titles[m.TitleFont], "DELETE", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case createPage.Index:
		p := NewFormPage(m.Titles[m.TitleFont], m.Table, []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case developmentPage.Index:
		p := NewStaticPage(m.Titles[m.TitleFont], "DEVELOPMENT", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	default:
		ui = m.RenderUI()
	}
	return ui
}

// InitPages initializes page relationships that can't be set during declaration due to circular references
func init() {
	// Set Next pointers for page navigation
	selectTablePage.Next = tableActionsPage
	
	// Set up children for pages
	homePage.Children = []*Page{developmentPage, cmsPage, selectTablePage, bucketPage, oauthPage, configPage}
	tableActionsPage.Children = []*Page{createPage, readPage, updatePage, deletePage}
}