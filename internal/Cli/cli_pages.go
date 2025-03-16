package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	config "github.com/hegner123/modulacms/internal/Config"
)

type CliPage struct {
	Index      int
	Controller CliInterface
	Label      string
	Parent     *CliPage
	Children   []*CliPage
	Next       *CliPage
}

const (
	Home = iota
	CMS
	Database
	Bucket
	OAuth
	Config
	Table
	Create
	Read
	Update
	Delete
	Form
	Content
	Media
	Users
)

var (
	homePage     *CliPage = &CliPage{Index: Home, Controller: pageInterface, Label: "Home", Parent: nil, Children: homepageMenu}
	cmsPage      *CliPage = &CliPage{Index: CMS, Controller: pageInterface, Label: "CMS", Parent: nil, Children: cmsMenu}
	databasePage *CliPage = &CliPage{Index: Database, Controller: tableInterface, Label: "Database", Parent: nil, Children: nil, Next: tablePage}
	bucketPage   *CliPage = &CliPage{Index: Bucket, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: []*CliPage{}}
	oauthPage    *CliPage = &CliPage{Index: OAuth, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: []*CliPage{}}
	configPage   *CliPage = &CliPage{Index: Config, Controller: pageInterface, Label: "Configuration", Parent: nil, Children: []*CliPage{}}
	tablePage    *CliPage = &CliPage{Index: Table, Controller: pageInterface, Label: "Table", Parent: nil, Children: tableMenu}
	createPage   *CliPage = &CliPage{Index: Create, Controller: tableInterface, Label: "Create", Parent: nil, Children: nil}
	readPage     *CliPage = &CliPage{Index: Read, Controller: tableInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage   *CliPage = &CliPage{Index: Update, Controller: tableInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage   *CliPage = &CliPage{Index: Delete, Controller: tableInterface, Label: "Delete", Parent: nil, Children: nil}
	formPage     *CliPage = &CliPage{Index: Form, Controller: createInterface, Label: "Form", Parent: nil, Children: nil}
	contentPage  *CliPage = &CliPage{Index: Content, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage    *CliPage = &CliPage{Index: Media, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage    *CliPage = &CliPage{Index: Users, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
)

func (m model) PageHome() string {
	m.header = "MAIN MENU"
	m.body = "\n"
	for i, choice := range m.menu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s\n", cursor, choice.Label)
	}
	m.body += "\n"
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageDatabase() string {
	m.header = "Select table"
	for i, choice := range m.tables {
		cursor := "  "
		if m.cursor == i {
			cursor = "->" // cursor!
		}

		m.body += fmt.Sprintf("%s%s\n", cursor, choice)
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageCMS() string {
	m.header = "Edit Your Content \n"
	for i, choice := range m.menu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s\n", cursor, choice.Label)
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageTable() string {
	m.header = m.header + "\nDatabase Method\n"
	for i, choice := range m.menu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s\n", cursor, choice.Label)
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageCreate() string {
	m.header = m.header + fmt.Sprintf("\nCreate %s\n", m.table)
	cs, ct, err := GetColumns(m.table)
	if err != nil {
		return err.Error()
	}
	for i, v := range *cs {
		c := *ct
		m.body += fmt.Sprintf("%s%s\n", v, c[i].DatabaseTypeName())
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageRead() string {
	m.header = m.header + fmt.Sprintf("\n\nRead %s\n", m.table)
	hdrs, collection, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	t := StyledTable(*hdrs, *collection, m.cursor)

	m.body = t.Render()
	return m.RenderUI()
}

func (m model) PageUpdate() string {
	m.header = m.header + fmt.Sprintf("\n\nUpdate %s\n", m.table)
	hdrs, collection, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	t := StyledTable(*hdrs, *collection, m.cursor)

	m.body = t.Render()
	return m.RenderUI()
}

func (m model) PageDelete() string {
	m.header = m.header + fmt.Sprintf("\n\nDelete%s\n", m.table)
	hdrs, collection, err := GetColumnsRows(m.table)
	if err != nil {
		fmt.Println("err", err)
	}

	t := StyledTable(*hdrs, *collection, m.cursor)

	m.body = t.Render()
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

func (m model) FormView() string {
	m.header = m.header + "\n\nCreate\n"
	if m.form.State == huh.StateCompleted {
		class := m.form.GetString("class")
		level := m.form.GetInt("level")
		return fmt.Sprintf("You selected: %s, Lvl. %d", class, level)
	}
	m.body = m.form.View()
	return m.RenderUI()
}

func (m model) PageBucket() string {
	m.header = "\nBucket Settings\n"
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
