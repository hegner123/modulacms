package cli

import "fmt"
var (
	homePage     *CliPage = &CliPage{Index: 0, Controller: pageInterface, Label: "Home", Parent: nil, Children: []*CliPage{cmsPage, databasePage, bucketPage, oauthPage, configPage}}
	cmsPage      *CliPage = &CliPage{Index: 1, Controller: pageInterface, Label: "CMS", Parent: nil, Children: []*CliPage{contentPage, mediaPage, usersPage}}
	databasePage *CliPage = &CliPage{Index: 2, Controller: pageInterface, Label: "Database", Parent: nil, Children: []*CliPage{createPage, readPage, updatePage, deletePage}}
	bucketPage   *CliPage = &CliPage{Index: 3, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: []*CliPage{}}
	oauthPage    *CliPage = &CliPage{Index: 4, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: []*CliPage{}}
	configPage   *CliPage = &CliPage{Index: 5, Controller: pageInterface, Label: "Configuration", Parent: nil, Children: []*CliPage{}}
	createPage   *CliPage = &CliPage{Index: 6, Controller: tableInterface, Label: "Create", Parent: nil, Children: nil}
	readPage     *CliPage = &CliPage{Index: 7, Controller: tableInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage   *CliPage = &CliPage{Index: 8, Controller: tableInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage   *CliPage = &CliPage{Index: 9, Controller: tableInterface, Label: "Delete", Parent: nil, Children: nil}
	tablePage    *CliPage = &CliPage{Index: 10, Controller: tableInterface, Label: "Table", Parent: nil, Children: nil}
	contentPage  *CliPage = &CliPage{Index: 11, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage    *CliPage = &CliPage{Index: 12, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage    *CliPage = &CliPage{Index: 13, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
)

func (m model) PageCreate() string {
	m.header = m.header + fmt.Sprintf("ModulaCMS\n\nCreate %s\n", m.table)
	m.body = fmt.Sprintf("%v", GetFields(m.table, ""))

	return m.RenderUI()
}

func (m model) PageRead() string {
	m.header = m.header + "ModulaCMS\n\nRead\n"
	m.body = fmt.Sprintf("%v", GetFields(m.table, ""))
	return m.RenderUI()
}

func (m model) PageUpdate() string {
	m.header = m.header + "ModulaCMS\n\nUpdate\n"
	m.body = fmt.Sprintf("%v", GetFields(m.table, ""))
	return m.RenderUI()
}

func (m model) PageDelete() string {
	m.header = m.header + "ModulaCMS\n\nDelete\n"
	m.body = fmt.Sprintf("%v", GetFields(m.table, ""))
	return m.RenderUI()
}

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

func (m model) SelectTableUI(action string) string {
	m.header += fmt.Sprintf("\nSelect table to %s\n\n", action)

	for i, choice := range m.tables {

		cursor := " " // no cursor
		if m.cursor == i {
			cursor = "->" // cursor!
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice)
	}

	return m.RenderUI()
}

func (m model) PageDatabase() string {
	m.header = fmt.Sprintf("%v\nEditing %s\n\n", m.header, m.table)

	for i, choice := range m.menu {

		cursor := " "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice.Label)
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
