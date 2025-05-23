package cli

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

type formCompletedMsg struct{}
type formCancelledMsg struct{}

type FocusKey int
type ApplicationState int

const (
	PAGEFOCUS FocusKey = iota
	TABLEFOCUS
	FORMFOCUS
	DIALOGFOCUS
)

const (
	OK ApplicationState = iota
	EDITING
	DELETING
	WARN
	ERROR
)

type CliInterface string
type InputType string

type Model struct {
	config       *config.Config
	status       ApplicationState
	titleFont    int
	titles       []string
	term         string
	profile      string
	width        int
	height       int
	bg           string
	txtStyle     lipgloss.Style
	quitStyle    lipgloss.Style
	loading      bool
	cursor       int
	cursorMax    int
	focusIndex   int
	page         Page
	paginator    paginator.Model
	pageMod      int
	maxRows      int
	table        string
	pageMenu     []*Page
	pages        []Page
	datatypeMenu []string
	tables       []string
	columns      *[]string
	columnTypes  *[]*sql.ColumnType
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
	content      string
	ready        bool
	err          error
	spinner      spinner.Model
	viewport     viewport.Model
	controller   CliInterface
	history      []PageHistory
	QueryResults []sql.Row
	time         time.Time
	dialog       *DialogModel
	dialogActive bool
}

var CliContinue bool = false

// ShowDialog creates a command to show a dialog
func ShowDialog(title, message string, showCancel bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDialogMsg{
			Title:      title,
			Message:    message,
			ShowCancel: showCancel,
		}
	}
}

func InitialModel(v *bool, c *config.Config) (Model, tea.Cmd) {

	verbose := false
	if v != nil {
		verbose = *v
	}

	// TODO add conditional to check ui config for custom titles
	fs, err := TitleFile.ReadDir("titles")
	if err != nil {
		utility.DefaultLogger.Fatal("", err)
	}
	fonts := ParseTitleFonts(fs)

	// paginator
	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

	// spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		config:     c,
		status:     OK,
		titleFont:  0,
		titles:     LoadTitles(fonts),
		focusIndex: 0,
		page:       *homePage,
		paginator:  p,
		loading:    false,
		spinner:    s,
		pageMod:    0,
		cursorMax:  0,
		maxRows:    10,
		table:      "",
		viewport:   viewport.Model{},
		pageMenu: []*Page{
			developmentPage,
			cmsPage,
			selectTablePage,
			bucketPage,
			oauthPage,
			configPage,
		},
		pages: []Page{
			*homePage,
			*cmsPage,
			*selectTablePage,
			*bucketPage,
			*oauthPage,
			*configPage,
			*tableActionsPage,
			*createPage,
			*readPage,
			*updatePage,
			*deletePage,
			*updateFormPage,
			*readSinglePage,
			*dynamicPage,
			*definedDatatypePage,
			*developmentPage,
		},
		selected:   make(map[int]struct{}),
		formMap:    make([]string, 0),
		controller: pageInterface,
		focus:      PAGEFOCUS,
		history:    []PageHistory{},
		verbose:    verbose,
	}, GetTablesCMD(c)
}

func (m Model) GetIDRow() int64 {
	rows := m.rows
	row := rows[m.cursor]
	rowCol := row[0]
	utility.DefaultLogger.Finfo("rowCOl", rowCol)
	id, err := strconv.ParseInt(rowCol, 10, 64)
	if err != nil {
		utility.DefaultLogger.Ferror("", err)
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

func (m Model) GetStatus() string {
	switch m.status {
	case EDITING:
		editStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent).Background(config.DefaultStyle.AccentBG).Bold(true).Padding(0, 1)
		return editStyle.Render(" EDIT ")
	case DELETING:
		deleteStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent2).Background(config.DefaultStyle.Accent2BG).Bold(true).Blink(true).Padding(0, 1)
		return deleteStyle.Render("DELETE")
	case WARN:
		warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Background(config.DefaultStyle.WarnBG).Bold(true).Padding(0, 1)
		return warnStyle.Render(" WARN ")
	case ERROR:
		errorStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent2).Background(config.DefaultStyle.Accent2BG).Bold(true).Blink(true).Padding(0, 1)
		return errorStyle.Render("ERROR ")
	default:
		okStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent).Background(config.DefaultStyle.AccentBG).Bold(true).Padding(0, 1)
		return okStyle.Render("  OK  ")
	}

}
