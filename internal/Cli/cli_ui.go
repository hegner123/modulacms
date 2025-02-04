package cli

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

type CliPage struct {
	Index    int
	Label    string
	Parent   *CliPage
	Children []*CliPage
}

type CliInterface string

type model struct {
	page       CliPage
	pages      []CliPage
	menu       []*CliPage
	table      string
	tables     []string
	cursor     int
	selected   map[int]struct{}
	header     string
	body       string
	footer     string
	controller CliInterface
}

var (
	homePage     *CliPage = &CliPage{Index: 0, Label: "Home", Parent: nil, Children: []*CliPage{databasePage, cmsPage, bucketPage, oauthPage, configPage}}
	databasePage *CliPage = &CliPage{Index: 1, Label: "Database", Parent: nil, Children: []*CliPage{createPage, readPage, updatePage, deletePage}}
	cmsPage      *CliPage = &CliPage{Index: 2, Label: "CMS", Parent: nil, Children: []*CliPage{}}
	bucketPage   *CliPage = &CliPage{Index: 3, Label: "Bucket", Parent: nil, Children: []*CliPage{}}
	oauthPage    *CliPage = &CliPage{Index: 4, Label: "oAuth", Parent: nil, Children: []*CliPage{}}
	configPage   *CliPage = &CliPage{Index: 5, Label: "Configuration", Parent: nil, Children: []*CliPage{}}
	createPage   *CliPage = &CliPage{Index: 6, Label: "Create", Parent: nil, Children: nil}
	readPage     *CliPage = &CliPage{Index: 7, Label: "Read", Parent: nil, Children: nil}
	updatePage   *CliPage = &CliPage{Index: 8, Label: "Update", Parent: nil, Children: nil}
	deletePage   *CliPage = &CliPage{Index: 9, Label: "Delete", Parent: nil, Children: nil}
	tablePage    *CliPage = &CliPage{Index: 10, Label: "Table", Parent: nil, Children: nil}
)
var (
	createInterface CliInterface = "CreateInterface"
	readInterface   CliInterface = "ReadInterface"
	updateInterface CliInterface = "UpdateInterface"
	deleteInterface CliInterface = "DeleteInterface"
	tableInterface  CliInterface = "TableInterface"
	pageInterface   CliInterface = "PageInterface"
)

func CliRun() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}

func initialModel() model {
	return model{
		page:   *homePage,
		tables: GetTables(""),
		menu:   []*CliPage{
            cmsPage, 
            databasePage,
            bucketPage,
            oauthPage,
            configPage,
        },
		pages: []CliPage{
			*homePage,
			*cmsPage,
			*databasePage,
            *bucketPage,
            *oauthPage,
            *configPage,
			*createPage,
			*readPage,
			*updatePage,
			*deletePage,
			*tablePage,
		},
		selected:   make(map[int]struct{}),
		controller: pageInterface,
	}
}

func (m model) Init() tea.Cmd {
	m.selected = make(map[int]struct{})
	return m.LaunchCms()
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch m.controller {
	case createInterface:
	case readInterface:
	case updateInterface:
	case deleteInterface:
	case pageInterface:
		return m.UpdatePageSelect(message)
	case tableInterface:
		return m.UpdateTableSelect(message)

	}
	return m, nil
}
func (m model) UpdateTableSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tables)-1 {
				m.cursor++
			}
		case "enter":
			m.table = m.tables[m.cursor]
		}
	}
	return m, nil
}

func (m model) UpdatePageSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
			m.page = *m.menu[m.cursor]
			m.menu = m.page.Children
		}
	}
	return m, nil
}

func (m model) View() string {
	var ui string
	switch m.page.Index {
	case homePage.Index:
		ui = m.PageHome()
	case databasePage.Index:
		ui = m.PageDatabase()
	case cmsPage.Index:
		ui = m.PageCMS()
	case createPage.Index:
		ui = m.PageCreate()
	case readPage.Index:
		ui = m.PageRead()
	case updatePage.Index:
		ui = m.PageUpdate()
	case deletePage.Index:
		ui = m.PageDelete()
	}
	return ui
}

func (m model) LaunchCms() tea.Cmd {
	m.tables = GetTables("")
	return func() tea.Msg {
		m.Update("Launch")
		return m.View()

	}

}

func (m model) RenderUI() string {
	m.footer += fmt.Sprintf("%v\nPress q to quit.\n%v", utility.REDF, utility.RESET)
	return m.header + m.body + m.footer
}
func (m model) PageCreate() string {
	m.header = "ModulaCMS\n\nCreate\n"

	if m.table == "" {
		m.controller = tableInterface
		m.SelectTableUI()
	} else {
		m.header += m.table
		m.controller = pageInterface
	}
	return m.RenderUI()
}
func (m model) PageRead() string {
	m.header = "ModulaCMS\n\nRead\n"
	return m.RenderUI()
}
func (m model) PageUpdate() string {
	m.header = "ModulaCMS\n\nUpdate\n"
	return m.RenderUI()
}
func (m model) PageDelete() string {
	m.header = "ModulaCMS\n\nDelete\n"
	return m.RenderUI()
}

func (m model) PageHome() string {
	m.header = "ModulaCMS\n\n MAIN MENU\n"

	for i, choice := range m.menu {

		cursor := " "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice.Label)
	}

	return m.RenderUI()

}

func (m model) SelectTableUI() string {
	m.tables = GetTables("")
	m.header = "Select table to edit?\n\n"

	for i, choice := range m.tables {

		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		m.body = fmt.Sprintf("%s  %s\n", cursor, choice)
	}

	return m.RenderUI()
}

func (m model) PageDatabase() string {
	m.header = "Select table to edit?\n\n"

	for i, choice := range m.menu {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice.Label)
	}

	return m.RenderUI()
}

func (m model) PageCMS() string {
	m.header = "ModulaCMS?\n\nEdit Your Content \n"
	for i, choice := range m.menu {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice.Label)
	}
	return m.RenderUI()
}
