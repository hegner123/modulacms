package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
)

// Screen is the interface for self-contained TUI pages that manage their own
// state, update logic, and rendering. Screens receive an AppContext snapshot
// and return commands -- they never mutate the root Model directly.
//
// All pages implement Screen. The root Update dispatches messages to the
// ActiveScreen, except WindowSizeMsg and provisioning which are always
// handled by the root Model.
type Screen interface {
	// Update processes a message and returns the (possibly new) Screen plus
	// any commands to execute. The AppContext provides read-only access to
	// shared state; mutations to Model fields (like PanelFocus or overlays)
	// are performed via returned tea.Cmd messages.
	Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd)

	// View renders the screen content. For panel-based screens this is the
	// full panel layout; for single-panel screens it fills the available area.
	View(ctx AppContext) string

	// PageIndex returns the PageIndex constant for this screen, used to
	// keep Model.Page in sync during the coexistence period.
	PageIndex() PageIndex
}

// KeyHint represents a single key:label pair for the statusbar.
type KeyHint struct {
	Key   string // display key (e.g. "n", "enter", "tab")
	Label string // action label (e.g. "new", "select", "panel")
}

// KeyHinter is an optional interface that Screen implementations can satisfy
// to provide context-aware key hints for the statusbar. Screens that do not
// implement this interface will show no key hints in the statusbar.
type KeyHinter interface {
	KeyHints(km config.KeyMap) []KeyHint
}

// OpenFilePickerMsg is emitted by Screen implementations to activate the
// shared file picker on the root Model.
type OpenFilePickerMsg struct {
	Purpose FilePickerPurpose
}

// HandleCommonKeys processes shared keybinding actions (quit, back, cursor
// movement) that all Screen implementations handle the same way. Returns
// the updated cursor, an optional command, and whether the key was handled.
func HandleCommonKeys(key string, km config.KeyMap, cursor, cursorMax int) (newCursor int, cmd tea.Cmd, handled bool) {
	if km.Matches(key, config.ActionQuit) {
		return cursor, func() tea.Msg { return ShowQuitConfirmDialogMsg{} }, true
	}
	if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
		return cursor, HistoryPopCmd(), true
	}
	if km.Matches(key, config.ActionUp) {
		if cursor > 0 {
			return cursor - 1, nil, true
		}
		return cursor, nil, true
	}
	if km.Matches(key, config.ActionDown) {
		if cursor < cursorMax {
			return cursor + 1, nil, true
		}
		return cursor, nil, true
	}
	return cursor, nil, false
}

// screenForPage returns the Screen implementation for a given page.
func (m Model) screenForPage(page Page) Screen {
	// Screens are created with empty/nil initial data. Each screen's init
	// command (dispatched by update_navigation.go) fetches the real data,
	// which arrives as *Set messages handled by the screen's Update().
	switch page.Index {
	case HOMEPAGE:
		return NewHomeScreen(m.HomepageMenuInit(), m.AdminUsername)
	case CMSPAGE:
		return NewCMSMenuScreen(false, m.CmsMenuInit())
	case ADMINCMSPAGE:
		return NewCMSMenuScreen(true, m.AdminCmsMenuInit())
	case PLUGINDETAILPAGE:
		return NewPluginDetailScreen(m.SelectedPlugin, nil)
	case ACTIONSPAGE:
		return NewActionsScreen(m.IsRemote)
	case QUICKSTARTPAGE:
		return NewQuickstartScreen()
	case MEDIA:
		return NewMediaScreen(nil)
	case USERSADMIN:
		return NewUsersScreen(nil, nil)
	case FIELDTYPES:
		return NewFieldTypesScreen(false, nil, nil)
	case ADMINFIELDTYPES:
		return NewFieldTypesScreen(true, nil, nil)
	case PLUGINSPAGE:
		return NewPluginsScreen(nil)
	case ROUTES:
		return NewRoutesScreen(false, nil, nil, m.PageRouteId)
	case ADMINROUTES:
		return NewRoutesScreen(true, nil, nil, m.PageRouteId)
	case WEBHOOKSPAGE:
		return NewWebhooksScreen(nil)
	case DEPLOYPAGE:
		return NewDeployScreen(nil)
	case PIPELINESPAGE:
		return NewPipelinesScreen(nil)
	case PIPELINEDETAILPAGE:
		return NewPipelineDetailScreen(nil, nil, "")
	case DATABASEPAGE:
		return NewDatabaseScreen(m.Tables, m.TableState)
	case READPAGE:
		// READPAGE is deprecated; redirect to unified DATABASEPAGE
		return NewDatabaseScreen(m.Tables, m.TableState)
	case CONFIGPAGE:
		return NewConfigScreen("", nil, 0)
	case DATATYPES:
		return NewDatatypesScreen(false)
	case ADMINDATATYPES:
		return NewDatatypesScreen(true)
	case CONTENT:
		return NewContentScreen(false, nil, nil, nil, nil, m.PageRouteId)
	case ADMINCONTENT:
		return NewContentScreen(true, nil, nil, nil, nil, "")
	}
	return nil
}
