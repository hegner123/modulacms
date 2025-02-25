package cli

import (
	"database/sql"
	"log"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	db "github.com/hegner123/modulacms/internal/Db"
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
	cursor       int
	focusIndex   int
	page         CliPage
	table        string
	menu         []*CliPage
	pages        []CliPage
	tables       []string
	textInputs   []textinput.Model
	textAreas    []textarea.Model
	filePicker   []filepicker.Model
	Options      []OptionList
	selected     map[int]struct{}
	header       string
	body         string
	footer       string
	textarea     textarea.Model
	controller   CliInterface
	history      []CliPage
	Query        db.SQLQuery
	QueryResults []sql.Row
	err          error
}

var CliContinue bool = false

func CliRun() bool {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	return CliContinue
}

func initialModel() model {
	return model{
		focusIndex: 0,
		page:       *homePage,
		tables:     GetTables(""),
		table:      "",
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
		textInputs: make([]textinput.Model, 0),
		textAreas:  make([]textarea.Model, 0),
		filePicker: make([]filepicker.Model, 0),
		history:    []CliPage{},
	}
}
