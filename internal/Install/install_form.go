package install

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type DBDriver int

const (
	Sqlite DBDriver = iota
	Mysql
	Postgres
)

type installModel struct {
	CustomConfig    bool
	DbDriver        DBDriver
	UseSSL          bool
	DbFileExists    bool
	ContentVersion  bool
	Certificates    bool
	Key             bool
	ConfigExists    bool
	DBConnected     bool
	BucketConnected bool
	OauthConnected  bool
	InstallCerts    bool
	page            string
	form            *huh.Form
	msg             string
}

func (m installModel) InstallForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Setup Config").
				Description("ModulaCMS requires a config file. Would you like to customize it now or use the default? You can always update the config later.").
				Options(
					huh.NewOption("Default", false),
					huh.NewOption("Custom", true),
				).Value(&m.CustomConfig),
			huh.NewSelect[DBDriver]().
				Title("Select Database Driver").
				Description("ModulaCMS supports three sql drivers (Sqlite,Mysql,Postgres).").
				Options(
					huh.NewOption("Sqlite(default)", Sqlite),
					huh.NewOption("MySql", Mysql),
					huh.NewOption("Postgres", Postgres),
				).Value(&m.DbDriver),
			huh.NewSelect[bool]().
				Title("Would you like to create ssl certificates?").
				Options(
					huh.NewOption("Yes", true),
					huh.NewOption("No", false),
				).Value(&m.InstallCerts),
		),
	)
}

func initialModel(init ModulaInit) installModel {
	return installModel{
		UseSSL:          init.UseSSL,
		DbFileExists:    init.DbFileExists,
		ContentVersion:  init.ContentVersion,
		Certificates:    init.Certificates,
		Key:             init.Key,
		ConfigExists:    init.ConfigExists,
		DBConnected:     init.DBConnected,
		BucketConnected: init.BucketConnected,
		OauthConnected:  init.OauthConnected,
	}
}

func defaultModel() installModel {
	return installModel{
		UseSSL:          false,
		DbFileExists:    false,
		ContentVersion:  false,
		Certificates:    false,
		Key:             false,
		ConfigExists:    false,
		DBConnected:     false,
		BucketConnected: false,
		OauthConnected:  false,
	}

}
func (m *installModel) Init() tea.Cmd {
	m.form = m.InstallForm()
	return nil
}

func (m *installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.page {
	case "form":
		m.FormControls(msg)
	default:
		ms, _ := msg.(tea.KeyMsg)
		switch ms.String() {
		case "q":
			os.Exit(0)
		case "enter":
			if m.form != nil {
				m.page = "form"
			}
		}
	}

	return m, cmd
}

func (m *installModel) FormControls(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	k, _ := msg.(tea.KeyMsg)
	m.msg = k.String()
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}
	if m.form.State == huh.StateAborted {
		os.Exit(1)
	}
	if m.form.State == huh.StateCompleted {
		os.Exit(0)
	}
	return m, tea.Batch(cmds...)
}

func (m *installModel) View() string {
	var ui string
	switch m.page {
	case "form":
		ui = m.FormView()
	default:
		ui = ""

	}
	return ui
}

func (m *installModel) FormView() string {
	var columnA string
	if m.form != nil {
		columnA = m.form.View()
	} else {
		columnA = "loading"
	}
	columnB := m.RenderStatus()

	columns := lipgloss.JoinHorizontal(lipgloss.Center, columnA, columnB)
	row := lipgloss.JoinVertical(lipgloss.Top, columns)

	return row
}

func (m *installModel) RenderStatus() string {
	style := lipgloss.NewStyle().Padding(0, 4)
	doc := lipgloss.JoinVertical(lipgloss.Top,
		fmt.Sprintf("Customize Config %v", m.CustomConfig),
		fmt.Sprintf("DB Driver %v", m.DbDriver),
		fmt.Sprintf("Install Certs %v", m.InstallCerts),
		fmt.Sprintf("message %v", m.msg),
	)
	return style.Render(doc)
}

func InstallUI(init *ModulaInit) bool {
	var m installModel
	if init != nil {
		m = initialModel(*init)
	} else {
		m = defaultModel()
	}
	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
	}

	return false
}
