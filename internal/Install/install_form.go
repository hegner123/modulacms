package install

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	cli "github.com/hegner123/modulacms/internal/Cli"
)

type installModel struct {
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
	form            *huh.Form
}

var ()

func InstallForm(m installModel) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
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
func (m installModel) Init() tea.Cmd {
	return nil
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.form = InstallForm(m)
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	return m, cmd
}

func (m installModel) View() string {
	m.form = InstallForm(m)
	columnA := m.form.View()
	columnBBuilder := m.RenderInstallModel()
	columnB := cli.RenderBlock(columnBBuilder)

	columns := lipgloss.JoinHorizontal(lipgloss.Center, columnA, columnB)
	row := lipgloss.JoinVertical(lipgloss.Top, columns)

	return row
}

func (m installModel) RenderInstallModel() string {
	doc := lipgloss.JoinVertical(lipgloss.Top,
		fmt.Sprintf("UseSSL %v", m.UseSSL),
		fmt.Sprintf("DbFileExists %v", m.DbFileExists),
		fmt.Sprintf("ContentVersion %v", m.ContentVersion),
		fmt.Sprintf("Certificates %v", m.Certificates),
		fmt.Sprintf("Key %v", m.Key),
		fmt.Sprintf("ConfigExists %v", m.ConfigExists),
		fmt.Sprintf("DBConnected %v", m.DBConnected),
		fmt.Sprintf("BucketConnected %v", m.BucketConnected),
		fmt.Sprintf("OauthConnected %v", m.OauthConnected),
	)
	return doc
}

func InstallUI(init ModulaInit) bool {
	p := tea.NewProgram(initialModel(init))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
	}
	return false

}
