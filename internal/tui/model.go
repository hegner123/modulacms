// Package cli implements the terminal user interface for ModulaCMS using Charmbracelet Bubbletea.
// It provides an SSH-accessible TUI for managing content, datatypes, media, routes, and users
// through a Model-Update-View architecture with typed message flows and database abstraction.
package tui

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
	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/tree"

	"github.com/hegner123/modulacms/internal/utility"
)

// FocusKey represents which UI component has keyboard focus.
type FocusKey int

// ApplicationState represents the current operational state of the TUI.
type ApplicationState int

// FormOptionsMap maps form field names to their available select options.
type FormOptionsMap map[string][]huh.Option[string]

// Focus key constants define which UI component has keyboard focus.
const (
	PAGEFOCUS FocusKey = iota
	TABLEFOCUS
	FORMFOCUS
	DIALOGFOCUS
)

// Application state constants define the current operational state of the TUI.
const (
	OK ApplicationState = iota
	EDITING
	DELETING
	WARN
	ERROR
)

// CliInterface represents the type of CLI interface being used.
type CliInterface string

// InputType represents the type of input field in a form.
type InputType string

// FilePickerPurpose indicates the intended use of the file picker dialog.
type FilePickerPurpose int

// File picker purpose constants define how the file picker will be used.
const (
	FILEPICKER_MEDIA FilePickerPurpose = iota
	FILEPICKER_RESTORE
)

// DatabaseMode represents the current database operation mode.
type DatabaseMode int

const (
	DBModeRead DatabaseMode = iota
	DBModeUpdate
	DBModeDelete
)

// ModelInterface defines the interface for interacting with CLI model
type ModelInterface interface {
	GetConfig() *config.Config
	GetRoot() model.Root
	SetRoot(root model.Root)
	SetError(err error)
}

// Model is the root Bubbletea model for the ModulaCMS TUI, containing all application state, UI components, and database connections.
type Model struct {
	DB                     db.DbDriver
	Config                 *config.Config
	Logger                 Logger
	Status                 ApplicationState
	TitleFont              int
	Titles                 []string
	Term                   string
	Profile                string
	Width                  int
	Height                 int
	Bg                     string
	PageRouteId            types.RouteID
	TxtStyle               lipgloss.Style
	QuitStyle              lipgloss.Style
	Loading                bool
	Cursor                 int
	CursorMax              int
	FocusIndex             int
	Paginator              paginator.Model
	PageMod                int
	MaxRows                int
	Page                   Page
	PageMenu               []Page
	Pages                  []Page
	PageMap                map[PageIndex]Page
	DatatypeMenu           []string
	Tables                 []string
	FormState              *FormModel
	TableState             *TableModel
	DatabaseMode           DatabaseMode
	Focus                  FocusKey
	Verbose                bool
	Content                string
	Ready                  bool
	Err                    error
	Spinner                spinner.Model
	Viewport               viewport.Model
	History                []PageHistory
	QueryResults           []sql.Row
	Time                   time.Time
	ActiveOverlay          ModalOverlay // nil = no dialog active
	Root                   tree.Root
	PanelFocus             FocusPanel
	Routes                 []db.Routes
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
	RootContentSummary     []db.ContentDataTopLevel
	UsersList              []db.UserWithRoleLabelRow
	RolesList              []db.Roles

	// Admin CMS state
	AdminRoutes                 []db.AdminRoutes
	AdminAllDatatypes           []db.AdminDatatypes
	AdminSelectedDatatypeFields []db.AdminFields
	AdminRootContentSummary     []db.AdminContentDataTopLevel
	AdminSelectedContentFields  []AdminContentFieldDisplay
	AdminFieldCursor            int

	// Field types state
	FieldTypesList      []db.FieldTypes
	AdminFieldTypesList []db.AdminFieldTypes

	// Cursor tracking across tree reloads (e.g., after sibling reorder)
	PendingCursorContentID types.ContentID

	// Plugin management
	PluginManager  *plugin.Manager
	PluginsList    []PluginDisplay
	SelectedPlugin string
	AdminUsername  string

	// Pipeline management
	PipelinesList       []PipelineDisplay
	PipelineEntries     []PipelineEntryDisplay
	SelectedPipelineKey string

	// Config management
	ConfigManager        *config.Manager
	ConfigCategory       config.FieldCategory
	ConfigCategoryFields []config.FieldMeta
	ConfigFieldCursor    int

	// Deploy state
	DeployEnvironments    []config.DeployEnvironmentConfig
	DeployLastResult      *DeploySyncResult
	DeployLastHealth      *DeployHealthResult
	DeployStatusMessage   string
	DeployOperationActive bool

	// SSH User Provisioning
	NeedsProvisioning bool
	SSHFingerprint    string
	SSHKeyType        string
	SSHPublicKey      string
	UserID            types.UserID

	// Webhook management
	WebhooksList []db.Webhook
	Dispatcher   publishing.WebhookDispatcher // nil when webhooks disabled

	// i18n locale state
	ActiveLocale string // Current locale code; "" means i18n disabled / default behavior

	// Version list state
	Versions         []db.ContentVersion
	ShowVersionList  bool
	VersionContentID types.ContentID
	VersionRouteID   types.RouteID
	VersionCursor    int

	// IsRemote is true when connected to a remote CMS server via Go SDK.
	// Used to guard operations that require local database access.
	IsRemote bool

	// RemoteURL is the base URL of the remote CMS server (e.g., "https://cms.example.com").
	// Empty when running in local mode. Displayed in the status bar.
	RemoteURL string

	// DBReadyCh is signalled after DB init/redeploy so serve can reload
	// the permission cache and start HTTP/HTTPS servers.
	DBReadyCh chan struct{}

	// DCtx holds per-session dialog context, replacing former package-level vars.
	DCtx DialogContext
}

// PluginDisplay holds the display-ready fields for a plugin in the TUI list.
type PluginDisplay struct {
	Name             string
	Version          string
	State            string
	CBState          string
	Description      string
	ManifestDrift    bool
	CapabilityDrifts int
	SchemaDrifts     int
}

// PipelineDisplay holds the display-ready fields for a pipeline chain in the TUI list.
type PipelineDisplay struct {
	Key       string // "table.phase_operation"
	Table     string
	Operation string
	Phase     string // "before" or "after"
	Count     int    // entries in this chain
}

// PipelineEntryDisplay holds the display-ready fields for a single pipeline entry.
type PipelineEntryDisplay struct {
	PipelineID string
	PluginName string
	Handler    string
	Priority   int
	Enabled    bool
}

// ContentFieldDisplay represents a content field for right panel display.
type ContentFieldDisplay struct {
	ContentFieldID types.ContentFieldID
	FieldID        types.FieldID
	Label          string
	Type           string
	Value          string
	ValidationJSON string // raw JSON from fields.validation
	DataJSON       string // raw JSON from fields.data
}

// DialogContext holds per-session dialog/operation context that was previously stored
// in package-level variables. Moving these into Model ensures concurrent SSH sessions
// cannot overwrite each other's dialog state.
type DialogContext struct {
	// from update_dialog.go
	InitRouteContent    *InitializeRouteContentContext
	RestoreBackup       *RestoreBackupContext
	RestoreRequiresQuit bool
	DeleteContent       *DeleteContentContext
	PublishContent      *PublishContentContext
	RestoreVersion      *RestoreVersionContext
	ApprovePluginRoutes *ApprovePluginContext
	ApprovePluginHooks  *ApprovePluginContext
	DeleteContentField  *DeleteContentFieldContext
	EditSingleField     *editSingleFieldCtx
	AddContentField     *addContentFieldCtx
	DeleteDatatype      *DeleteDatatypeContext
	DeleteField         *DeleteFieldContext
	DeleteRoute         *DeleteRouteContext
	DeleteMedia         *DeleteMediaContext
	DeleteUser          *DeleteUserContext

	// from admin_update_dialog.go
	DeleteAdminRoute     *DeleteAdminRouteContext
	DeleteAdminDatatype  *DeleteAdminDatatypeContext
	DeleteAdminField     *DeleteAdminFieldContext
	DeleteFieldType      *DeleteFieldTypeContext
	DeleteAdminFieldType *DeleteAdminFieldTypeContext

	// from deploy_update.go
	DeployPull *DeployPullContext
	DeployPush *DeployPushContext
}

// CliContinue controls whether the CLI should continue running after processing a command.
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

// InitialModel creates and initializes a new Model with the provided configuration, database driver, logger, and optional plugin manager.
// dbReadyCh is an optional channel signalled after DB init so the serve command can start HTTP.
func InitialModel(v *bool, c *config.Config, driver db.DbDriver, logger Logger, pluginMgr *plugin.Manager, mgr *config.Manager, dbReadyCh chan struct{}, dispatcher publishing.WebhookDispatcher) (Model, tea.Cmd) {
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
	var systemAdminUsername string
	if users, err := driver.ListUsers(); err == nil && users != nil {
		for _, u := range *users {
			if strings.EqualFold(u.Username, "system") {
				systemAdminID = u.UserID
				systemAdminUsername = u.Username
				logger.Finfo(fmt.Sprintf("CLI mode: using system user %s (%s)", u.Username, u.UserID))
				break
			}
		}
	}

	m := Model{
		DB:            driver,
		Config:        c,
		Logger:        logger,
		Status:        OK,
		TitleFont:     0,
		Titles:        LoadTitles(fonts),
		FocusIndex:    0,
		Page:          NewPage(HOMEPAGE, "Home"),
		Paginator:     p,
		Loading:       false,
		Spinner:       s,
		PageMod:       0,
		CursorMax:     0,
		MaxRows:       10,
		Viewport:      viewport.Model{},
		PageMap:       *InitPages(),
		FormState:     NewFormModel(),
		TableState:    NewTableModel(),
		Focus:         PAGEFOCUS,
		PanelFocus:    TreePanel,
		History:       []PageHistory{},
		Verbose:       verbose,
		PageRouteId:   types.RouteID(""), // TODO: Implement route selection UI
		UserID:        systemAdminID,     // Set system admin for CLI mode
		AdminUsername: systemAdminUsername,
		PluginManager: pluginMgr,
		ConfigManager: mgr,
		DBReadyCh:     dbReadyCh,
		Dispatcher:    dispatcher,
	}
	m.PageMenu = m.HomepageMenuInit()
	return m, tea.Batch(
		GetTablesCMD(m.Config),
	)
}

// ModelPostInit performs post-initialization setup for the model, initializing menus and logging.
func ModelPostInit(m Model) tea.Cmd {
	return tea.Batch(
		LogMessageCmd("Test Menu Init"),
		PageMenuSetCmd(m.HomepageMenuInit()),
	)
}

// ParseTitles extracts font names from title file entries by removing the .txt extension and splitting on underscores.
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

// LoadTitles reads ASCII art title files from the embedded filesystem for the given font names.
func LoadTitles(f []string) []string {
	var titles []string
	for _, font := range f {
		aTitle, err := TitleFile.ReadFile("titles/title_" + font + ".txt")
		if err != nil {
			aTitle = []byte("Modula")
		}
		t := string(aTitle)
		titles = append(titles, t)
	}

	return titles
}

// GetStatus returns a styled status string based on the current application state.
func (m Model) GetStatus() string {
	black := lipgloss.Color("#000000")
	switch m.Status {
	case EDITING:
		editStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent).Background(black).Bold(true).Padding(0, 1)
		return editStyle.Render(" EDIT ")
	case DELETING:
		deleteStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent2).Background(black).Bold(true).Blink(true).Padding(0, 1)
		return deleteStyle.Render("DELETE")
	case WARN:
		warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Background(black).Bold(true).Padding(0, 1)
		return warnStyle.Render(" WARN ")
	case ERROR:
		errorStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent2).Background(black).Bold(true).Blink(true).Padding(0, 1)
		return errorStyle.Render("ERROR ")
	default:
		okStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent).Background(black).Bold(true).Padding(0, 1)
		return okStyle.Render("  OK  ")
	}
}

// GetConfig returns the model's configuration, implementing the cms.ModelInterface.
func (m *Model) GetConfig() *config.Config {
	return m.Config
}
