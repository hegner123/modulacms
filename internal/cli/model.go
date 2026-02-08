package cli

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/tree"
	"github.com/hegner123/modulacms/internal/tui"
	"github.com/hegner123/modulacms/internal/utility"
)

type FocusKey int
type ApplicationState int
type FormOptionsMap map[string][]huh.Option[string]

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

type FilePickerPurpose int

const (
	FILEPICKER_MEDIA   FilePickerPurpose = iota
	FILEPICKER_RESTORE
)

// ModelInterface defines the interface for interacting with CLI model
type ModelInterface interface {
	GetConfig() *config.Config
	GetRoot() model.Root
	SetRoot(root model.Root)
	SetError(err error)
}

type Model struct {
	DB           db.DbDriver
	Config       *config.Config
	Logger       Logger
	Status       ApplicationState
	TitleFont    int
	Titles       []string
	Term         string
	Profile      string
	Width        int
	Height       int
	Bg           string
	PageRouteId  types.RouteID
	TxtStyle     lipgloss.Style
	QuitStyle    lipgloss.Style
	Loading      bool
	Cursor       int
	CursorMax    int
	FocusIndex   int
	Paginator    paginator.Model
	PageMod      int
	MaxRows      int
	Page         Page
	PageMenu     []Page
	Pages        []Page
	PageMap      map[PageIndex]Page
	DatatypeMenu []string
	Tables       []string
	FormState    *FormModel
	TableState   *TableModel
	Focus        FocusKey
	Verbose      bool
	Content      string
	Ready        bool
	Err          error
	Spinner      spinner.Model
	Viewport     viewport.Model
	History      []PageHistory
	QueryResults []sql.Row
	Time         time.Time
	Dialog           *DialogModel
	DialogActive     bool
	FormDialog              *FormDialogModel
	FormDialogActive        bool
	ContentFormDialog       *ContentFormDialogModel
	ContentFormDialogActive bool
	Root                    tree.Root
	PanelFocus        tui.FocusPanel
	Routes            []db.Routes
	RootDatatypes          []db.Datatypes
	AllDatatypes           []db.Datatypes
	SelectedDatatype       types.DatatypeID
	SelectedDatatypeFields []db.Fields
	FieldCursor            int // Cursor for fields panel (datatypes: center, content: right)
	SelectedContentFields  []ContentFieldDisplay
	MediaList              []db.Media
	FilePicker             filepicker.Model
	FilePickerActive       bool
	FilePickerPurpose      FilePickerPurpose
	RootContentSummary     []db.RootContentSummary
	UsersList              []db.Users
	UserFormDialog             *UserFormDialogModel
	UserFormDialogActive       bool
	DatabaseFormDialog         *DatabaseFormDialogModel
	DatabaseFormDialogActive   bool

	// Admin CMS state
	AdminRoutes                 []db.AdminRoutes
	AdminAllDatatypes           []db.AdminDatatypes
	AdminSelectedDatatypeFields []db.AdminFields
	AdminRootContentSummary     []db.AdminContentData
	AdminSelectedContentFields  []AdminContentFieldDisplay
	AdminFieldCursor            int

	// Cursor tracking across tree reloads (e.g., after sibling reorder)
	PendingCursorContentID types.ContentID

	// SSH User Provisioning
	NeedsProvisioning bool
	SSHFingerprint    string
	SSHKeyType        string
	SSHPublicKey      string
	UserID            types.UserID
}

// ContentFieldDisplay represents a content field for right panel display.
type ContentFieldDisplay struct {
	ContentFieldID types.ContentFieldID
	DatatypeFieldID string
	FieldID        types.FieldID
	Label          string
	Type           string
	Value          string
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

func InitialModel(v *bool, c *config.Config, driver db.DbDriver, logger Logger) (Model, tea.Cmd) {
	// Use provided logger or fall back to utility.DefaultLogger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	verbose := false
	if v != nil {
		verbose = *v
	}

	// TODO add conditional to check ui config for custom titles
	fs, err := TitleFile.ReadDir("titles")
	if err != nil {
		logger.Fatal("", err)
	}
	fonts := ParseTitles(fs)

	// paginator
	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

	// spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Find system user for CLI mode (non-SSH sessions)
	var systemAdminID types.UserID
	if users, err := driver.ListUsers(); err == nil && users != nil {
		for _, u := range *users {
			if strings.EqualFold(u.Username, "system") {
				systemAdminID = u.UserID
				logger.Finfo(fmt.Sprintf("CLI mode: using system user %s (%s)", u.Username, u.UserID))
				break
			}
		}
	}

	m := Model{
		DB:          driver,
		Config:      c,
		Logger:      logger,
		Status:      OK,
		TitleFont:   0,
		Titles:      LoadTitles(fonts),
		FocusIndex:  0,
		Page:        NewPage(HOMEPAGE, "Home"),
		Paginator:   p,
		Loading:     false,
		Spinner:     s,
		PageMod:     0,
		CursorMax:   0,
		MaxRows:     10,
		Viewport:    viewport.Model{},
		PageMap:     *InitPages(),
		FormState:   NewFormModel(),
		TableState:  NewTableModel(),
		Focus:       PAGEFOCUS,
		PanelFocus:  tui.TreePanel,
		History:     []PageHistory{},
		Verbose:     verbose,
		PageRouteId: types.RouteID(""), // TODO: Implement route selection UI
		UserID:      systemAdminID,     // Set system admin for CLI mode
	}
	m.PageMenu = m.HomepageMenuInit()
	return m, tea.Batch(
		GetTablesCMD(m.Config),
	)
}

func ModelPostInit(m Model) tea.Cmd {
	return tea.Batch(
		LogMessageCmd("Test Menu Init"),
		PageMenuSetCmd(m.HomepageMenuInit()),
	)
}

func ParseTitles(f []fs.DirEntry) []string {
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
