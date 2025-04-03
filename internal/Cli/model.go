package cli

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

type formCompletedMsg struct{}
type formCancelledMsg struct{}

type FocusKey int

const (
	PAGEFOCUS FocusKey = iota
	FORMFOCUS
	DIALOGFOCUS
)

type CliInterface string
type InputType string

type model struct {
	titleFont    int
	titles       []string
	term         string
	profile      string
	width        int
	height       int
	bg           string
	txtStyle     lipgloss.Style
	quitStyle    lipgloss.Style
	cursor       int
	focusIndex   int
	page         Page
	table        string
	pageMenu     []*Page
	pages        []Page
    datatypeMenu []string
	tables       []string
	selected     map[int]struct{}
	headers      []string
	rows         [][]string
	row          *[]string
	form         *huh.Form
	formLen      int
	formMap      []string
	formValues   []*string
	formSubmit   bool
	formGroups   []huh.Group
	formFields   []huh.Field
	focus        FocusKey
	title        string
	header       string
	body         string
	footer       string
	verbose      bool
	controller   CliInterface
	history      []Page
	QueryResults []sql.Row
	time         time.Time
}

var CliContinue bool = false

func InitialModel(v *bool) model {
	verbose := false
	if v != nil {
		verbose = *v
	}
	fs, err := TitleFile.ReadDir("titles")
	if err != nil {
		utility.DefaultLogger.Fatal("", err)
	}
	fonts := ParseTitleFonts(fs)
	return model{
		titleFont:  0,
		titles:     LoadTitles(fonts),
		focusIndex: 0,
		page:       *homePage,
		tables:     GetTables(),
		table:      "",
		pageMenu: []*Page{
			cmsPage,
			databasePage,
			bucketPage,
			oauthPage,
			configPage,
		},
		pages: []Page{
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
			*dynamicPage,
			*defineDatatype,
		},
		selected:   make(map[int]struct{}),
		formMap:    make([]string, 0),
		controller: pageInterface,
		focus:      PAGEFOCUS,
		history:    []Page{},
		verbose:    verbose,
	}
}

func (m model) GetIDRow() int64 {
	logFile, _ := tea.LogToFile("debug.log", "debug")
	defer logFile.Close()
	rows := m.rows
	row := rows[m.cursor]
	rowCol := row[0]
	utility.DefaultLogger.Finfo(logFile, "rowCOl", rowCol)
	id, err := strconv.ParseInt(rowCol, 10, 64)
	if err != nil {
		utility.DefaultLogger.Ferror(logFile, "", err)
	}
	return id

}

func ParseTitleFonts(f []fs.DirEntry) []string {
	var fonts []string

	for _, file := range f {
		rmExt := strings.TrimSuffix(file.Name(), ".txt")
		name := strings.Split(rmExt, "_")
		if len(name) < 1 {
			err := fmt.Errorf("font name not correctly formated %v", file.Name())
			utility.DefaultLogger.Fatal("", err)
		}
		fonts = append(fonts, name[1])
	}
	return fonts
}

func LoadTitles(f []string) []string {
	var titles []string
	for _, font := range f {
		aTitle, err := TitleFile.ReadFile("titles/title_" + font + ".txt")
		if err != nil {
			aTitle = []byte("ModulaCMS")
		}
		t := string(aTitle)
		titles = append(titles, t)
	}

	return titles
}
