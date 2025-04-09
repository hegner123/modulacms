package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
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
)

var (
	homePage       *Page = &Page{Index: HOMEPAGE, Controller: pageInterface, Label: "Home", Parent: nil, Children: homepageMenu}
	cmsPage        *Page = &Page{Index: CMSPAGE, Controller: pageInterface, Label: "CMS", Parent: nil, Children: cmsMenu}
	databasePage   *Page = &Page{Index: DATABASEPAGE, Controller: tableInterface, Label: "Database", Parent: nil, Children: nil, Next: tablePage}
	bucketPage     *Page = &Page{Index: BUCKETPAGE, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: nil}
	oauthPage      *Page = &Page{Index: OAUTHPAGE, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: nil}
	configPage     *Page = &Page{Index: CONFIGPAGE, Controller: configInterface, Label: "Configuration", Parent: nil, Children: nil}
	tablePage      *Page = &Page{Index: TABLEPAGE, Controller: pageInterface, Label: "Table", Parent: nil, Children: tableMenu}
	createPage     *Page = &Page{Index: CREATEPAGE, Controller: createInterface, Label: "Create", Parent: nil, Children: nil}
	readPage       *Page = &Page{Index: READPAGE, Controller: readInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage     *Page = &Page{Index: UPDATEPAGE, Controller: updateInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage     *Page = &Page{Index: DELETEPAGE, Controller: deleteInterface, Label: "Delete", Parent: nil, Children: nil}
	updateFormPage *Page = &Page{Index: UPDATEFORMPAGE, Controller: updateFormInterface, Label: "UpdateForm", Parent: nil, Children: nil}
	readSinglePage *Page = &Page{Index: READSINGLEPAGE, Controller: readSingleInterface, Label: "ReadSingle", Parent: nil, Children: nil}
	contentPage    *Page = &Page{Index: CONTENTPAGE, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage      *Page = &Page{Index: MEDIAPAGE, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage      *Page = &Page{Index: USERSPAGE, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
	dynamicPage    *Page = &Page{Index: DYNAMICPAGE, Controller: pageInterface, Label: "Dynamic", Parent: nil, Children: nil}
)

func (m model) View() string {
	var ui string
	switch m.page.Index {
	case homePage.Index:
		p := NewMenuPage(m.pageMenu, m.titles[m.titleFont], "MAIN MENU", []Row{}, "q quit", m.RenderStatusBar())
		ui = p.Render(m.cursor, m)
	case databasePage.Index:
		ui = m.PageDatabase()
	case cmsPage.Index:
		ui = m.PageCMS()
	case bucketPage.Index:
		ui = m.PageBucket()
	case configPage.Index:
		ui = m.PageConfig()
	case tablePage.Index:
		ui = m.PageTable()
	case createPage.Index:
		ui = m.PageCreate()
	case readPage.Index:
		ui = m.PageRead()
	case updatePage.Index:
		ui = m.PageUpdate()
	case updateFormPage.Index:
		ui = m.PageUpdateForm()
	case deletePage.Index:
		ui = m.PageDelete()
	case contentPage.Index:
		ui = m.PageContent()
	case readSinglePage.Index:
		ui = m.PageReadSingle()
	case defineDatatype.Index:
		ui = m.PageDefineDatatype()
	default:
		ui = m.Page404()
	}
	return ui
}

func (m model) PageHome() string {
	m.header = "MAIN MENU"
	var b strings.Builder
	for i, choice := range m.pageMenu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		b.WriteString(fmt.Sprintf("%s%s  \n", cursor, choice.Label))
	}
	l := lipgloss.JoinHorizontal(
		lipgloss.Top,
		RenderBorderFixed(b.String()),
		RenderBlockFixed(StatusBlock()),
	)
	m.body = l
	return m.RenderUI()
}

func (m model) PageDatabase() string {
	m.header = "Select table"
	var group string
	var row []string
	var column []string

	for i, choice := range m.tables {

		cursor := "   "
		if m.cursor == i {
			cursor = " ->"
		}

		fs := fmt.Sprintf("%s%s   ", cursor, choice)
		column = append(column, fs)
		if (i+1)%6 == 0 || i == len(m.tables)-1 {
			c := NewVerticalGroup(lipgloss.Left, column)
			row = append(row, c)
			column = []string{}
		}
	}
	group = NewHorizontalGroup(lipgloss.Top, row)
	m.body = RenderBorderFlex(group)
	return m.RenderUI()
}

func (m model) PageCMS() string {
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

func (m model) PageTable() string {
	m.header = "Database Method"
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

func (m model) PageContent() string {
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

func (m model) PageCreate() string {
	m.header = m.header + fmt.Sprintf("Create %s", m.table)
	m.body = "\n"
	if m.form == nil {
		m.body += "Form == nil"
	}
	m.body += m.form.View()
	return m.RenderUI()
}

func (m model) PageRead() string {
	m.header = m.header + fmt.Sprintf("Read %s", m.table)
	h, r, err := GetColumnsRows(m.table)
	if err != nil {
		utility.DefaultLogger.Finfo("", err)
	}
	start, end := m.paginator.GetSliceBounds(len(m.rows))

	t := StyledTable(h, r[start:end], m.cursor)
	m.body = t.Render()
	if len(r) > m.maxRows {
		m.body += "\n\n" + m.paginator.View()
	}
	return m.RenderUI()
}

func (m model) PageReadSingle() string {
	m.header = m.header + fmt.Sprintf("Read %s", m.table)
	h, c, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	m.body = RenderEntry(h, c[m.cursor])
	return m.RenderUI()
}

func (m model) PageUpdate() string {
	m.header = m.header + fmt.Sprintf("Select Row to Update %s", m.table)
	h, r, err := GetColumnsRows(m.table)
	if err != nil {
		utility.DefaultLogger.Finfo("", err)
	}
	start, end := m.paginator.GetSliceBounds(len(m.rows))

	t := StyledTable(h, r[start:end], m.cursor)

	m.body = t.Render()
	if len(r) > m.maxRows {
		m.body += "\n\n" + m.paginator.View()
	}
	return m.RenderUI()
}

func (m *model) PageUpdateForm() string {
	m.header = m.header + fmt.Sprintf("Update Row  %s", m.table)
	m.body = "\n"
	if m.form == nil {
		m.body += "Form == nil"
	}
	m.body += m.form.View()
	return m.RenderUI()
}

func (m model) PageDelete() string {
	m.header = m.header + fmt.Sprintf("Delete %s", m.table)
	h, r, err := GetColumnsRows(m.table)
	if err != nil {
		utility.DefaultLogger.Finfo("", err)
	}
	start, end := m.paginator.GetSliceBounds(len(m.rows))

	t := StyledTable(h, r[start:end], m.cursor)

	m.body = t.Render()
	if len(r) > m.maxRows {
		m.body += "\n\n" + m.paginator.View()
	}
	return m.RenderUI()
}

func (m model) PageBucket() string {
	m.header = "Bucket Settings"
	b1 := config.Env.Bucket_Url
	b2 := config.Env.Bucket_Endpoint
	b3 := config.Env.Bucket_Access_Key
	b4 := config.Env.Bucket_Secret_Key
	m.body += "\nBucket_Url: " + b1
	m.body += "\nBucket_Endpoint: " + b2
	m.body += "\nBucket_Access_Key: " + b3
	m.body += "\nBucket_Secret_Key: " + b4
	m.body = RenderBorderFixed(m.body)
	return m.RenderUI()
}

func (m model) PageConfig() string {
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

func (m model) PageDynamic() string {
	m.header = "Content Editor"
	m.body = "\n"
	if m.form == nil {
		m.body += "Form == nil"
	}
	m.body += m.form.View()
	return m.RenderUI()
}

func (m model) Page404() string {
	m.header = "PAGE NOT FOUND"

	return m.RenderUI()
}
