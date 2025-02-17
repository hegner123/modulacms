package cli

import "fmt"

type CliPage struct {
	Index      int
	Controller CliInterface
	Label      string
	Parent     *CliPage
	Children   []*CliPage
	Next       *CliPage
}

var (
	homePage     *CliPage = &CliPage{Index: 0, Controller: pageInterface, Label: "Home", Parent: nil, Children: homepageMenu}
	cmsPage      *CliPage = &CliPage{Index: 1, Controller: pageInterface, Label: "CMS", Parent: nil, Children: cmsMenu}
	databasePage *CliPage = &CliPage{Index: 2, Controller: tableInterface, Label: "Database", Parent: nil, Children: nil, Next: tablePage}
	bucketPage   *CliPage = &CliPage{Index: 3, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: []*CliPage{}}
	oauthPage    *CliPage = &CliPage{Index: 4, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: []*CliPage{}}
	configPage   *CliPage = &CliPage{Index: 5, Controller: pageInterface, Label: "Configuration", Parent: nil, Children: []*CliPage{}}
	tablePage    *CliPage = &CliPage{Index: 6, Controller: pageInterface, Label: "Table", Parent: nil, Children: tableMenu}
	createPage   *CliPage = &CliPage{Index: 7, Controller: tableInterface, Label: "Create", Parent: nil, Children: nil}
	readPage     *CliPage = &CliPage{Index: 8, Controller: tableInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage   *CliPage = &CliPage{Index: 9, Controller: tableInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage   *CliPage = &CliPage{Index: 10, Controller: tableInterface, Label: "Delete", Parent: nil, Children: nil}
	contentPage  *CliPage = &CliPage{Index: 11, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage    *CliPage = &CliPage{Index: 12, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage    *CliPage = &CliPage{Index: 13, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
)


func (m model) PageHome() string {
	m.header = m.header + "ModulaCMS\n\n MAIN MENU\n"

	for i, choice := range m.menu {

		cursor := " "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice.Label)
	}

	return m.RenderUI()

}

func (m model) PageDatabase() string {
	m.header += "\nSelect table\n\n"

	for i, choice := range m.tables {

		cursor := " " // no cursor
		if m.cursor == i {
			cursor = "->" // cursor!
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice)
	}

	return m.RenderUI()
}

func (m model) PageCMS() string {
	m.header = m.header + "ModulaCMS\n\nEdit Your Content \n"
	for i, choice := range m.menu {

		cursor := " "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice.Label)
	}
	return m.RenderUI()
}

func (m model) PageTable() string {
	m.header = m.header + "ModulaCMS\n\nDatabase Method\n"
	for i, choice := range m.menu {

		cursor := " "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice.Label)
	}
	return m.RenderUI()
}

func (m model) PageCreate() string {
	m.header = m.header + fmt.Sprintf("ModulaCMS\n\nCreate %s\n", m.table)
	m.body = fmt.Sprintf("%v", GetFields(m.table, ""))

	return m.RenderUI()
}

func (m model) PageRead() string {
	m.header = m.header + "ModulaCMS\n\nRead\n"
	m.body = fmt.Sprintf("%v", GetFieldsString(m.table, ""))
	return m.RenderUI()
}

func (m model) PageUpdate() string {
	m.header = m.header + "ModulaCMS\n\nUpdate\n"
	m.body = fmt.Sprintf("%v", GetFieldsString(m.table, ""))
	return m.RenderUI()
}

func (m model) PageDelete() string {
	m.header = m.header + "ModulaCMS\n\nDelete\n"
	m.body = fmt.Sprintf("%v", GetFieldsString(m.table, ""))
	return m.RenderUI()
}

// View renders the UI.
func (m model) InputsView() string {
	s := "Dynamic Bubble Tea Inputs Example\n\n"
	for i, input := range m.textInputs {
		s += input.View() + "\n"
		if i < len(m.textInputs)-1 {
			s += "\n"
		}
	}
	s += "\nPress tab/shift+tab or up/down to switch focus. Press esc or ctrl+c to quit."
	return s
}
