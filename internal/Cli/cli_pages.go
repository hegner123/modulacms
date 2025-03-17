package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/Config"
)

type PageIndex int

const (
	Home PageIndex = iota
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
	UpdateForm
	Content
	Media
	Users
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
	homePage       *CliPage = &CliPage{Index: Home, Controller: pageInterface, Label: "Home", Parent: nil, Children: homepageMenu}
	cmsPage        *CliPage = &CliPage{Index: CMS, Controller: pageInterface, Label: "CMS", Parent: nil, Children: cmsMenu}
	databasePage   *CliPage = &CliPage{Index: Database, Controller: tableInterface, Label: "Database", Parent: nil, Children: nil, Next: tablePage}
	bucketPage     *CliPage = &CliPage{Index: Bucket, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: nil}
	oauthPage      *CliPage = &CliPage{Index: OAuth, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: nil}
	configPage     *CliPage = &CliPage{Index: Config, Controller: pageInterface, Label: "Configuration", Parent: nil, Children: nil}
	tablePage      *CliPage = &CliPage{Index: Table, Controller: pageInterface, Label: "Table", Parent: nil, Children: tableMenu}
	createPage     *CliPage = &CliPage{Index: Create, Controller: createInterface, Label: "Create", Parent: nil, Children: nil}
	readPage       *CliPage = &CliPage{Index: Read, Controller: readInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage     *CliPage = &CliPage{Index: Update, Controller: updateInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage     *CliPage = &CliPage{Index: Delete, Controller: deleteInterface, Label: "Delete", Parent: nil, Children: nil}
	updateFormPage *CliPage = &CliPage{Index: UpdateForm, Controller: updateFormInterface, Label: "UpdateForm", Parent: nil, Children: nil}
	contentPage    *CliPage = &CliPage{Index: Content, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage      *CliPage = &CliPage{Index: Media, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage      *CliPage = &CliPage{Index: Users, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
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
	var group string
	var row []string
	var column []string

	for i, choice := range m.tables {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		fs := fmt.Sprintf("%s%s %d %d\n", cursor, choice, i, len(m.tables))
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
	m.header = "\nDatabase Method\n"
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
	m.body = m.form.View()
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
func (m model) PageUpdateForm() string {
	m.header = m.header + fmt.Sprintf("\n\nUpdate %s\n", m.table)
	doc := strings.Builder{}
	buttons := []string{}
	fStyle := lipgloss.NewStyle().Height(30)
	if m.form == nil {
		m.body = "Form == nil"
	}
	for _, action := range m.formActions {
		if m.formActions[m.cursor] == action {
			buttons = append(buttons, RenderActiveButton(string(action)))
		} else {
			buttons = append(buttons, RenderButton(string(action)))
		}
	}
    m.footer = " tab|right|down Next Field .. shift+tab|left|up = Previous Field"

	doc.WriteString(
		lipgloss.JoinVertical(lipgloss.Top,
			fStyle.Render(m.form.View()),
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				buttons...,
			)),
	)
	m.body = doc.String()

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
	m.body = "Dynamic Bubble Tea Inputs Example\n\n"
	for i, input := range m.textInputs {
		m.body += input.View() + "\n"
		if i < len(m.textInputs)-1 {
			m.body += "\n"
		}
	}
	m.footer = "\nPress tab/shift+tab or up/down to switch focus. Press esc or ctrl+c to quit."
	return m.RenderUI()
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

func (m model) PageConfig() string {
	m.header = "\nConfiguration\n"
	s, _ := json.Marshal(config.Env)
	m.body += "\n" + string(s)

	return m.RenderUI()
}

func (m model) Page404() string {
	m.header = "\nPAGE NOT FOUND\n"

	return m.RenderUI()
}
