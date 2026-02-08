package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

type NavigationUpdated struct{}

func NewNavUpdate() tea.Cmd {
	return func() tea.Msg {
		return NavigationUpdated{}
	}
}

func (m Model) UpdateNavigation(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateToPage:
		var cmds []tea.Cmd
		cmds = append(cmds, HistoryPushCmd(PageHistory{Page: m.Page, Cursor: m.Cursor, Menu: m.PageMenu}))
		cmds = append(cmds, CursorResetCmd())
		switch msg.Page.Index {
		case CMSPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PageMenuSetCmd(m.CmsMenuInit()))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case ADMINCMSPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PageMenuSetCmd(m.AdminCmsMenuInit()))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case DATABASEPAGE:
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))

			return m, tea.Batch(cmds...)
		case TABLEPAGE:
			cmds = append(cmds, PageMenuSetCmd(m.DatabaseMenuInit()))
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, GetColumnsCmd(*m.Config, m.TableState.Table))

			return m, tea.Batch(cmds...)
		case READPAGE:
			f := MakeFilter("Spinner", "Pages", "History", "Pages", "Viewport", "Paginator")
			page := m.PageMap[READPAGE]
			cmds = append(cmds, LogModelCMD(nil, &f))
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, &page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case READSINGLEPAGE:
			page := m.PageMap[READSINGLEPAGE]
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case UPDATEPAGE:
			page := m.PageMap[UPDATEPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, &page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case DELETEPAGE:
			page := m.PageMap[DELETEPAGE]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, &page))
			cmds = append(cmds, StatusSetCmd(DELETING))

			return m, tea.Batch(cmds...)
		case DYNAMICPAGE:
			page := m.PageMap[DYNAMICPAGE]
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case DATATYPES:
			page := m.PageMap[DATATYPES]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, AllDatatypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case DATATYPE:
			page := m.PageMap[DATATYPE]
			cmds = append(cmds, FocusSetCmd(FORMFOCUS))
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, LoadingStopCmd())

			return m, tea.Batch(cmds...)
		case FIELDS:
			page := m.PageMap[FIELDS]
			cmds = append(cmds, FocusSetCmd(FORMFOCUS))
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, LoadingStopCmd())

			return m, tea.Batch(cmds...)

		case CONTENT:
			page := m.PageMap[CONTENT]
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, RootContentSummaryFetchCmd())
			cmds = append(cmds, RootDatatypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			// If a datatype is already selected, fetch its routes
			if !m.SelectedDatatype.IsZero() {
				cmds = append(cmds, RoutesByDatatypeFetchCmd(m.SelectedDatatype))
			}

			// Load content tree if PageRouteId is set
			if !m.PageRouteId.IsZero() {
				cmds = append(cmds, ReloadContentTreeCmd(m.Config, m.PageRouteId))
			}

			return m, tea.Batch(cmds...)
		case PICKCONTENT:
			page := NewPickContentPage("Pick")
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case EDITCONTENT:
			page := m.PageMap[EDITCONTENT]
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(EDITING))

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
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))
			cmds = append(cmds, PanelFocusResetCmd())

			return m, tea.Batch(cmds...)
		case CONFIGPAGE:
			content, err := formatJSON(m.Config)
			if err == nil {
				cmds = append(cmds, SetViewportContentCmd(content))
			} else {
				cmds = append(cmds, SetViewportContentCmd(m.Content))
			}
			cmds = append(cmds, ReadyTrueCmd())
			cmds = append(cmds, PageSetCmd(m.PageMap[CONFIGPAGE]))

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
		}

		return m, nil
	case SelectTable:
		// Set table synchronously to avoid race with NavigateToPageCmd
		newModel := m
		newModel.TableState.Table = m.Tables[m.Cursor]
		cmds := make([]tea.Cmd, 0)
		cmds = append(cmds, NavigateToPageCmd(m.PageMap[TABLEPAGE]))
		cmds = append(cmds, PageMenuSetCmd(m.DatabaseMenuInit()))

		return newModel, tea.Batch(cmds...)
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
