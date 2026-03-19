package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/utility"
)

// NavigationUpdated signals that page navigation has been processed.
type NavigationUpdated struct{}

// NewNavUpdate returns a command that creates a NavigationUpdated message.
func NewNavUpdate() tea.Cmd {
	return func() tea.Msg {
		return NavigationUpdated{}
	}
}

// UpdateNavigation handles page navigation transitions and initialization.
func (m Model) UpdateNavigation(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateToPage:
		var cmds []tea.Cmd
		cmds = append(cmds, HistoryPushCmd(PageHistory{Page: m.Page, Cursor: m.Cursor, Menu: m.PageMenu, Screen: m.ActiveScreen}))
		cmds = append(cmds, CursorResetCmd())

		// Set ActiveScreen for Screen-based pages (nil = legacy path)
		m.ActiveScreen = m.screenForPage(msg.Page)

		switch msg.Page.Index {
		case HOMEPAGE:
			utility.DefaultLogger.Fdebug("[home] NavigateToPage HOMEPAGE: dispatching PageSetCmd + HomeDashboardFetchCmd")
			cmds = append(cmds, PageSetCmd(msg.Page))
			return m, tea.Batch(cmds...)
		case CMSPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			return m, tea.Batch(cmds...)
		case ADMINCMSPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			return m, tea.Batch(cmds...)
		case DATABASEPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			return m, tea.Batch(cmds...)
		case READPAGE:
			f := MakeFilter("Spinner", "Pages", "History", "Pages", "Viewport", "Paginator")
			page := m.PageMap[READPAGE]
			cmds = append(cmds, LogModelCMD(nil, &f))
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, &page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case DATATYPES:
			page := m.PageMap[DATATYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AllDatatypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case CONTENT:
			page := m.PageMap[CONTENT]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, RootContentSummaryFetchCmd())
			cmds = append(cmds, RootDatatypesFetchCmd())
			cmds = append(cmds, UsersFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			if !m.PageRouteId.IsZero() {
				cmds = append(cmds, ReloadContentTreeCmd(m.Config, m.PageRouteId))
			}
			return m, tea.Batch(cmds...)
		case MEDIA:
			page := m.PageMap[MEDIA]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, MediaFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case USERSADMIN:
			page := m.PageMap[USERSADMIN]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, UsersFetchCmd())
			cmds = append(cmds, RolesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case CONFIGPAGE:
			cmds = append(cmds, ReadyTrueCmd())
			cmds = append(cmds, PageSetCmd(m.PageMap[CONFIGPAGE]))
			return m, tea.Batch(cmds...)
		case ROUTES:
			page := m.PageMap[ROUTES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, RoutesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case ACTIONSPAGE:
			cmds = append(cmds, PageSetCmd(m.PageMap[ACTIONSPAGE]))
			cmds = append(cmds, UpdateCheckCmd())
			return m, tea.Batch(cmds...)
		case QUICKSTARTPAGE:
			cmds = append(cmds, PageSetCmd(m.PageMap[QUICKSTARTPAGE]))
			return m, tea.Batch(cmds...)
		case ADMINROUTES:
			page := m.PageMap[ADMINROUTES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminRoutesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case ADMINDATATYPES:
			page := m.PageMap[ADMINDATATYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminAllDatatypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case ADMINCONTENT:
			page := m.PageMap[ADMINCONTENT]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminContentDataFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case FIELDTYPES:
			page := m.PageMap[FIELDTYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FieldTypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case ADMINFIELDTYPES:
			page := m.PageMap[ADMINFIELDTYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminFieldTypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case VALIDATIONS:
			page := m.PageMap[VALIDATIONS]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, ValidationsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case ADMINVALIDATIONS:
			page := m.PageMap[ADMINVALIDATIONS]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminValidationsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case PLUGINSPAGE:
			page := m.PageMap[PLUGINSPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, PluginsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case PLUGINDETAILPAGE:
			page := m.PageMap[PLUGINDETAILPAGE]
			cmds = append(cmds, PluginsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, CursorResetCmd())
			return m, tea.Batch(cmds...)
		case PLUGINTUIPAGE:
			// Plugin TUI screens are set up via NavigateToPluginScreenMsg,
			// which directly creates the screen and sends PluginScreenSetupCmd.
			// This case handles navigation via the page map (e.g., history pop).
			page := m.PageMap[PLUGINTUIPAGE]
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case DEPLOYPAGE:
			page := m.PageMap[DEPLOYPAGE]
			cmds = append(cmds, DeployEnvsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, CursorResetCmd())
			return m, tea.Batch(cmds...)
		case PIPELINESPAGE:
			page := m.PageMap[PIPELINESPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, PipelinesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case PIPELINEDETAILPAGE:
			page := m.PageMap[PIPELINEDETAILPAGE]
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, CursorResetCmd())
			return m, tea.Batch(cmds...)
		case WEBHOOKSPAGE:
			page := m.PageMap[WEBHOOKSPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, WebhooksFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case TOKENSPAGE:
			page := m.PageMap[TOKENSPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TokensFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case ROLESPAGE:
			page := m.PageMap[ROLESPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, RolesScreenFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case IMPORTPAGE:
			page := m.PageMap[IMPORTPAGE]
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case MEDIADIMENSIONSPAGE:
			page := m.PageMap[MEDIADIMENSIONSPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, MediaDimensionsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		case SESSIONSPAGE:
			page := m.PageMap[SESSIONSPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, SessionsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			return m, tea.Batch(cmds...)
		}

		return m, nil
	case FormCompletedMsg:
		cmds := make([]tea.Cmd, 0)
		newModel := m

		// Priority 1: Use specified destination page
		if msg.DestinationPage != nil {
			cmds = append(cmds, NavigateToPageCmd(*msg.DestinationPage))
			return newModel, tea.Batch(cmds...)
		}

		// Priority 2: Try to pop history
		entry := newModel.PopHistory()
		if entry != nil {
			if entry.Screen != nil {
				newModel.Page = entry.Page
				newModel.PageMenu = entry.Menu
				newModel.Cursor = entry.Cursor
				newModel.ActiveScreen = entry.Screen
				return newModel, nil
			}
			cmds = append(cmds, PageSetCmd(entry.Page))
			cmds = append(cmds, PageMenuSetCmd(entry.Menu))
			cmds = append(cmds, CursorSetCmd(entry.Cursor))
			return newModel, tea.Batch(cmds...)
		}

		// Priority 3: Fallback to home page
		cmds = append(cmds, NavigateToPageCmd(m.PageMap[HOMEPAGE]))
		return newModel, tea.Batch(cmds...)

	case HistoryPop:
		newModel := m
		entry := newModel.PopHistory()

		// Check if history was empty
		if entry == nil {
			return newModel, nil
		}

		// Restore saved screen directly to preserve cursor/focus/data.
		if entry.Screen != nil {
			newModel.Page = entry.Page
			newModel.PageMenu = entry.Menu
			newModel.Cursor = entry.Cursor
			newModel.ActiveScreen = entry.Screen
			return newModel, nil
		}

		// Fallback: no saved screen, create fresh one.
		cmds := []tea.Cmd{
			PageSetCmd(entry.Page),
			PageMenuSetCmd(entry.Menu),
			CursorSetCmd(entry.Cursor),
		}
		return newModel, tea.Batch(cmds...)

	default:
		return m, nil
	}
}
