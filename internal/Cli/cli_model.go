package cli

import (
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)


type OptionList struct {
	Key  string
	List []Option
}

type Option struct {
	Index     int
	InputType InputType
	Key       string
	Value     string
}

type CliInterface string
type InputType string
type errMsg error

type model struct {
	cursor     int
	page       CliPage
	table      string
	menu       []*CliPage
	pages      []CliPage
	tables     []string
	Options    []OptionList
	selected   map[int]struct{}
	header     string
	body       string
	footer     string
	textarea   textarea.Model
	controller CliInterface
	history    []CliPage
	err        error
}

func CliRun() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}

func initialModel() model {
	t1 := textarea.New()
	t1.Placeholder = ""
	return model{
		page:   *homePage,
		tables: GetTables(""),
		table:  "",
		menu: []*CliPage{
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
		textarea:   t1,
		history:    []CliPage{},
	}
}
