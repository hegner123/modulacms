package tui

import (
	tea "github.com/charmbracelet/bubbletea"
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
		cmds = append(cmds, HistoryPushCmd(PageHistory{Page: m.Page, Cursor: m.Cursor, Menu: m.PageMenu}))
		cmds = append(cmds, CursorResetCmd())

		// Set ActiveScreen for Screen-based pages (nil = legacy path)
		m.ActiveScreen = m.screenForPage(msg.Page)

		switch msg.Page.Index {
		case HOMEPAGE:
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PanelFocusResetCmd())
			cmds = append(cmds, HomeDashboardFetchCmd(m.DB))
			return m, tea.Batch(cmds...)
		case CMSPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PanelFocusResetCmd())
			return m, tea.Batch(cmds...)
		case ADMINCMSPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PanelFocusResetCmd())
			return m, tea.Batch(cmds...)
		case DATABASEPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case READPAGE:
			f := MakeFilter("Spinner", "Pages", "History", "Pages", "Viewport", "Paginator")
			page := m.PageMap[READPAGE]
			cmds = append(cmds, LogModelCMD(nil, &f))
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, &page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case DATATYPES:
			page := m.PageMap[DATATYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AllDatatypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case CONTENT:
			page := m.PageMap[CONTENT]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, RootContentSummaryFetchCmd())
			cmds = append(cmds, RootDatatypesFetchCmd())
			cmds = append(cmds, UsersFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			// Load content tree if PageRouteId is set
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
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case USERSADMIN:
			page := m.PageMap[USERSADMIN]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, UsersFetchCmd())
			cmds = append(cmds, RolesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case CONFIGPAGE:
			cmds = append(cmds, ReadyTrueCmd())
			cmds = append(cmds, PageSetCmd(m.PageMap[CONFIGPAGE]))
			cmds = append(cmds, PanelFocusResetCmd())
			return m, tea.Batch(cmds...)
		case ROUTES:
			page := m.PageMap[ROUTES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, RoutesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case ACTIONSPAGE:
			cmds = append(cmds, PageSetCmd(m.PageMap[ACTIONSPAGE]))
			cmds = append(cmds, PanelFocusResetCmd())
			cmds = append(cmds, UpdateCheckCmd())

			return m, tea.Batch(cmds...)
		case QUICKSTARTPAGE:
			cmds = append(cmds, PageSetCmd(m.PageMap[QUICKSTARTPAGE]))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case ADMINROUTES:
			page := m.PageMap[ADMINROUTES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminRoutesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case ADMINDATATYPES:
			page := m.PageMap[ADMINDATATYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminAllDatatypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case ADMINCONTENT:
			page := m.PageMap[ADMINCONTENT]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminContentDataFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case FIELDTYPES:
			page := m.PageMap[FIELDTYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FieldTypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case ADMINFIELDTYPES:
			page := m.PageMap[ADMINFIELDTYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AdminFieldTypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case PLUGINSPAGE:
			page := m.PageMap[PLUGINSPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, PluginsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case PLUGINDETAILPAGE:
			page := m.PageMap[PLUGINDETAILPAGE]
			cmds = append(cmds, PluginsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, CursorResetCmd())
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case DEPLOYPAGE:
			page := m.PageMap[DEPLOYPAGE]
			cmds = append(cmds, DeployEnvsFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())
			cmds = append(cmds, CursorResetCmd())

			return m, tea.Batch(cmds...)
		case PIPELINESPAGE:
			page := m.PageMap[PIPELINESPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, PipelinesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

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
			cmds = append(cmds, PanelFocusResetCmd())

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
			cmds = append(cmds, PageSetCmd(entry.Page))
			cmds = append(cmds, PageMenuSetCmd(entry.Menu))
			cmds = append(cmds, CursorSetCmd(entry.Cursor))
			return newModel, tea.Batch(cmds...)
		}

		// Priority 3: Fallback to home page
		cmds = append(cmds, NavigateToPageCmd(m.PageMap[HOMEPAGE]))
		return newModel, tea.Batch(cmds...)

	case HistoryPop:
		cmds := make([]tea.Cmd, 0)
		newModel := m
		entry := newModel.PopHistory()

		// Check if history was empty
		if entry == nil {
			return newModel, nil
		}

		cmds = append(cmds, PageSetCmd(entry.Page))
		cmds = append(cmds, PageMenuSetCmd(entry.Menu))
		cmds = append(cmds, CursorSetCmd(entry.Cursor))

		return newModel, tea.Batch(cmds...)

	default:
		return m, nil
	}
}
