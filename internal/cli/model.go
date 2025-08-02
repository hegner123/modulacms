package cli

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/model"
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

// ModelInterface defines the interface for interacting with CLI model
type ModelInterface interface {
	GetConfig() *config.Config
	GetRoot() model.Root
	SetRoot(root model.Root)
	SetError(err error)
}

type Model struct {
	Config       *config.Config
	Status       ApplicationState
	TitleFont    int
	Titles       []string
	Term         string
	Profile      string
	Width        int
	Height       int
	Bg           string
	TxtStyle     lipgloss.Style
	QuitStyle    lipgloss.Style
	Loading      bool
	Cursor       int
	CursorMax    int
	FocusIndex   int
	Page         Page
	Paginator    paginator.Model
	PageMod      int
	MaxRows      int
	Table        string
	PageMenu     []*Page
	Pages        []Page
	DatatypeMenu []string
	Tables       []string
	Columns      *[]string
	ColumnTypes  *[]*sql.ColumnType
	Selected     map[int]struct{}
	Headers      []string
	Rows         [][]string
	Row          *[]string
	Form         *huh.Form
	FormLen      int
	FormMap      []string
	FormValues   []*string
	FormSubmit   bool
	FormGroups   []huh.Group
	FormFields   []huh.Field
	Focus        FocusKey
	Title        string
	Header       string
	Body         string
	Footer       string
	Verbose      bool
	Content      string
	Ready        bool
	Err          error
	Spinner      spinner.Model
	Viewport     viewport.Model
	History      []PageHistory
	QueryResults []sql.Row
	Time         time.Time
	Dialog       *DialogModel
	DialogActive bool
	Root         model.Root
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
		Config:     c,
		Status:     OK,
		TitleFont:  0,
		Titles:     LoadTitles(fonts),
		FocusIndex: 0,
		Page:       *homePage,
		Paginator:  p,
		Loading:    false,
		Spinner:    s,
		PageMod:    0,
		CursorMax:  0,
		MaxRows:    10,
		Table:      "",
		Viewport:   viewport.Model{},
		PageMenu: []*Page{
			developmentPage,
			cmsPage,
			selectTablePage,
			bucketPage,
			oauthPage,
			configPage,
		},
		Pages: []Page{
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
		Selected:   make(map[int]struct{}),
		FormMap:    make([]string, 0),
		Focus:      PAGEFOCUS,
		History:    []PageHistory{},
		Verbose:    verbose,
	}, GetTablesCMD(c)
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
	switch m.Status {
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

// Implement cms.ModelInterface for Model
func (m *Model) GetConfig() *config.Config {
	return m.Config
}

func (m *Model) GetRoot() model.Root {
	return m.Root
}

func (m *Model) SetRoot(root model.Root) {
	m.Root = root
}

func (m *Model) SetError(err error) {
	m.Err = err
	m.Status = ERROR
}
