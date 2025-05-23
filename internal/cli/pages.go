package cli

import (
	"fmt"
	"strings"

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
	CONTENTPAGE
	MEDIAPAGE
	USERSPAGE
	DYNAMICPAGE
	DEFINEDATATYPE
	DEVELOPMENT
)

var (
	homePage            *Page = &Page{Index: HOMEPAGE, Controller: pageInterface, Label: "Home", Parent: nil, Children: homepageMenu}
	cmsPage             *Page = &Page{Index: CMSPAGE, Controller: pageInterface, Label: "CMS", Parent: nil, Children: cmsMenu}
	selectTablePage     *Page = &Page{Index: DATABASEPAGE, Controller: tableInterface, Label: "Database", Parent: nil, Children: nil, Next: tableActionsPage}
	bucketPage          *Page = &Page{Index: BUCKETPAGE, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: nil}
	oauthPage           *Page = &Page{Index: OAUTHPAGE, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: nil}
	configPage          *Page = &Page{Index: CONFIGPAGE, Controller: configInterface, Label: "Configuration", Parent: nil, Children: nil}
	tableActionsPage    *Page = &Page{Index: TABLEPAGE, Controller: pageInterface, Label: "Table", Parent: nil, Children: tableMenu}
	createPage          *Page = &Page{Index: CREATEPAGE, Controller: createInterface, Label: "Create", Parent: nil, Children: nil}
	readPage            *Page = &Page{Index: READPAGE, Controller: readInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage          *Page = &Page{Index: UPDATEPAGE, Controller: updateInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage          *Page = &Page{Index: DELETEPAGE, Controller: deleteInterface, Label: "Delete", Parent: nil, Children: nil}
	updateFormPage      *Page = &Page{Index: UPDATEFORMPAGE, Controller: updateFormInterface, Label: "UpdateForm", Parent: nil, Children: nil}
	readSinglePage      *Page = &Page{Index: READSINGLEPAGE, Controller: readSingleInterface, Label: "ReadSingle", Parent: nil, Children: nil}
	contentPage         *Page = &Page{Index: CONTENTPAGE, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage           *Page = &Page{Index: MEDIAPAGE, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage           *Page = &Page{Index: USERSPAGE, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
	dynamicPage         *Page = &Page{Index: DYNAMICPAGE, Controller: pageInterface, Label: "Dynamic", Parent: nil, Children: nil}
	definedDatatypePage *Page = &Page{Index: DEFINEDATATYPE, Controller: pageInterface, Label: "DefineDatatype", Parent: nil, Children: nil}
	developmentPage     *Page = &Page{Index: DEVELOPMENT, Controller: developmentInterface, Label: "Development", Parent: nil, Children: nil}
)

func (m Model) View() string {
	var ui string
	if m.loading {
		str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", m.spinner.View())
		return str
	}
	switch m.page.Index {
	case homePage.Index:
		menu := make([]string, 0, len(m.pageMenu))
		for _, v := range m.pageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.titles[m.titleFont], "MAIN MENU", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case selectTablePage.Index:
		menu := make([]string, 0, len(m.tables))
		menu = append(menu, m.tables...)
		p := NewMenuPage(menu, m.titles[m.titleFont], "TABLES", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case cmsPage.Index:
		menu := make([]string, 0, len(m.pageMenu))
		for _, v := range m.pageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.titles[m.titleFont], "CMS", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case bucketPage.Index:
		p := NewStaticPage(m.titles[m.titleFont], "BUCKET SETTINGS", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case configPage.Index:
		ui = m.PageConfig()
	case tableActionsPage.Index:
		menu := make([]string, 0, len(m.pageMenu))
		for _, v := range m.pageMenu {
			menu = append(menu, v.Label)
		}
		p := NewMenuPage(menu, m.titles[m.titleFont], fmt.Sprintf("USING %s", m.table), []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m)
	case createPage.Index:
		ui = m.PageCreate()
	case readPage.Index:
		p := NewTablePage(m.headers, m.rows, m.table, m.titles[m.titleFont], "READ", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case updatePage.Index:
		p := NewTablePage(m.headers, m.rows, m.table, m.titles[m.titleFont], "UPDATE", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case updateFormPage.Index:
		ui = m.PageUpdateForm()
	case deletePage.Index:
		p := NewTablePage(m.headers, m.rows, m.table, m.titles[m.titleFont], "DELETE", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(&m)
	case contentPage.Index:
		ui = m.PageContent()
	case readSinglePage.Index:
		ui = m.PageReadSingle()
	case definedDatatypePage.Index:
		ui = ""
	case developmentPage.Index:
		p := NewStaticPage(m.titles[m.titleFont], "Development", []Row{}, "q quit", "")
		ui = p.Render(m)
	default:
		ui = m.Page404()
	}
	return ui
}

func (m Model) PageCMS() string {
	m.header = "Edit Your Content"
	for i, choice := range m.pageMenu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice.Label)
	}
	m.body = RenderBorderFixed(m.body)
	return m.RenderUI()
}

func (m Model) PageContent() string {
	m.header = "Content"
	for i, choice := range m.pageMenu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice.Label)
	}
	m.body = RenderBorderFixed(m.body)
	return m.RenderUI()
}

func (m Model) PageCreate() string {
	m.header = m.header + fmt.Sprintf("Create %s", m.table)
	m.body = "\n"
	if m.form == nil {
		m.body += "Form == nil"
	}
	m.body += m.form.View()
	return m.RenderUI()
}

func (m Model) PageReadSingle() string {
	m.header = m.header + fmt.Sprintf("Read %s", m.table)

	m.body = RenderEntry(m.headers, m.rows[m.cursor])
	return m.RenderUI()
}

func (m *Model) PageUpdateForm() string {
	m.header = m.header + fmt.Sprintf("Update Row  %s", m.table)
	m.body = "\n"
	if m.form == nil {
		m.body += "Form == nil"
	}
	m.body += m.form.View()
	return m.RenderUI()
}

func (m Model) PageBucket() string {
	c := m.config
	m.header = "Bucket Settings"
	b1 := c.Bucket_Url
	b2 := c.Bucket_Endpoint
	b3 := c.Bucket_Access_Key
	b4 := c.Bucket_Secret_Key
	m.body += "\nBucket_Url: " + b1
	m.body += "\nBucket_Endpoint: " + b2
	m.body += "\nBucket_Access_Key: " + b3
	m.body += "\nBucket_Secret_Key: " + b4
	m.body = RenderBorderFixed(m.body)
	return m.RenderUI()
}

func (m Model) PageConfig() string {
	var b strings.Builder
	if !m.ready {
		return m.RenderUI()
	}
	b.WriteString(m.headerView() + "\n")
	b.WriteString(m.viewport.View() + "\n")
	b.WriteString(m.footerView() + "\n")
	m.body += b.String()
	return m.RenderUI()

}

func (m Model) PageDynamic() string {
	m.header = "Content Editor"
	m.body = "\n"
	if m.form == nil {
		m.body += "Form == nil"
	}
	m.body += m.form.View()
	return m.RenderUI()
}

func (m Model) Page404() string {
	m.header = "PAGE NOT FOUND"

	return m.RenderUI()
}
