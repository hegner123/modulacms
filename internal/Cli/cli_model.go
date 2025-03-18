package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/Db"
)

type formCompletedMsg struct{}
type formCancelledMsg struct{}

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

type FocusKey int

const (
	PAGEFOCUS FocusKey = iota
	FORMFOCUS
	DIALOGFOCUS
)

type CliInterface string
type InputType string

type model struct {
	cursor       int
	focusIndex   int
	page         CliPage
	table        string
	menu         []*CliPage
	pages        []CliPage
	tables       []string
	Options      []OptionList
	selected     map[int]struct{}
	headers      *[]string
	rows         *[][]string
	row          *[]string
	form         *huh.Form
	formLen      int
	formMap      map[string]string
	formValues   []*string
	formActions  []formAction
	formSubmit   bool
	formGroups   []huh.Group
	formFields   []huh.Field
	focus        FocusKey
	title        string
	header       string
	body         string
	footer       string
	controller   CliInterface
	history      []CliPage
	Query        db.SQLQuery
	QueryResults []sql.Row
}

var CliContinue bool = false

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
			*tablePage,
			*createPage,
			*readPage,
			*updatePage,
			*deletePage,
			*updateFormPage,
			*readSinglePage,
		},
		selected:    make(map[int]struct{}),
		formMap:     make(map[string]string),
		controller:  pageInterface,
		focus:       PAGEFOCUS,
		formActions: []formAction{edit, submit, reset, cancel},
		history:     []CliPage{},
	}
}

func (m model) GetIDRow() int64 {
	logFile, _ := tea.LogToFile("debug.log", "debug")
	defer logFile.Close()
	rows := *m.rows
	row := rows[m.cursor]
	rowCol := row[0]
	fmt.Fprintln(logFile, "RowCol", rowCol)
	id, err := strconv.ParseInt(rowCol, 10, 64)
	if err != nil {
		fmt.Fprintln(logFile, err)
	}
	return id

}
