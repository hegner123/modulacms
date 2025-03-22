package cli

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/Config"
)

type PageIndex int

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
)

type CliPage struct {
	Index      PageIndex
	Controller CliInterface
	Label      string
	Parent     *CliPage
	Children   []*CliPage
	Next       *CliPage
}

var (
	homePage       *CliPage = &CliPage{Index: HOMEPAGE, Controller: pageInterface, Label: "Home", Parent: nil, Children: homepageMenu}
	cmsPage        *CliPage = &CliPage{Index: CMSPAGE, Controller: pageInterface, Label: "CMS", Parent: nil, Children: cmsMenu}
	databasePage   *CliPage = &CliPage{Index: DATABASEPAGE, Controller: tableInterface, Label: "Database", Parent: nil, Children: nil, Next: tablePage}
	bucketPage     *CliPage = &CliPage{Index: BUCKETPAGE, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: nil}
	oauthPage      *CliPage = &CliPage{Index: OAUTHPAGE, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: nil}
	configPage     *CliPage = &CliPage{Index: CONFIGPAGE, Controller: pageInterface, Label: "Configuration", Parent: nil, Children: nil}
	tablePage      *CliPage = &CliPage{Index: TABLEPAGE, Controller: pageInterface, Label: "Table", Parent: nil, Children: tableMenu}
	createPage     *CliPage = &CliPage{Index: CREATEPAGE, Controller: createInterface, Label: "Create", Parent: nil, Children: nil}
	readPage       *CliPage = &CliPage{Index: READPAGE, Controller: readInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage     *CliPage = &CliPage{Index: UPDATEPAGE, Controller: updateInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage     *CliPage = &CliPage{Index: DELETEPAGE, Controller: deleteInterface, Label: "Delete", Parent: nil, Children: nil}
	updateFormPage *CliPage = &CliPage{Index: UPDATEFORMPAGE, Controller: updateFormInterface, Label: "UpdateForm", Parent: nil, Children: nil}
	readSinglePage *CliPage = &CliPage{Index: READSINGLEPAGE, Controller: readSingleInterface, Label: "ReadSingle", Parent: nil, Children: nil}
	contentPage    *CliPage = &CliPage{Index: CONTENTPAGE, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage      *CliPage = &CliPage{Index: MEDIAPAGE, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage      *CliPage = &CliPage{Index: USERSPAGE, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
	dynamicPage    *CliPage = &CliPage{Index: DYNAMICPAGE, Controller: pageInterface, Label: "Dynamic", Parent: nil, Children: nil}
)

func (m model) PageHome() string {
	m.header = "MAIN MENU"
	for i, choice := range m.menu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice.Label)
	}
	m.body = RenderBorder(m.body)
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
	m.body = RenderBorder(group)
	return m.RenderUI()
}

func (m model) PageCMS() string {
	m.header = "Edit Your Content"
	for i, choice := range m.menu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice.Label)
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageTable() string {
	m.header = "Database Method"
	for i, choice := range m.menu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice.Label)
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageContent() string {
	m.header = "Content"
	for i, choice := range m.menu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice.Label)
	}
	m.body = RenderBorder(m.body)
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
	f, _ := tea.LogToFile("debug.log", "debug")
	fmt.Fprintln(f, "Help")
	h, c, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	t := StyledTable(h, c, m.cursor)

	m.body = t.Render()
	return m.RenderUI()
}

func (m model) PageReadSingle() string {
	m.header = m.header + fmt.Sprintf("Read %s", m.table)
	h, c, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	t := StyledTable(h, c, m.cursor)

	m.body = t.Render()
	return m.RenderUI()
}

func (m model) PageUpdate() string {
	m.header = m.header + fmt.Sprintf("Update %s", m.table)
	h, c, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	t := StyledTable(h, c, m.cursor)

	m.body = t.Render()
	return m.RenderUI()
}
func (m *model) PageUpdateForm() string {
	m.header = m.header + fmt.Sprintf("Update %s", m.table)
	m.body = "\n"
	if m.form == nil {
		m.body += "Form == nil"
	}
	m.body += m.form.View()
	return m.RenderUI()
}

func (m model) PageDelete() string {
	m.header = m.header + fmt.Sprintf("Delete %s", m.table)
	h, c, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	t := StyledTable(h, c, m.cursor)

	m.body = t.Render()
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
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageConfig() string {
	m.header = "Configuration"
	s, _ := json.Marshal(config.Env)
	m.body += "\n" + string(s)

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
