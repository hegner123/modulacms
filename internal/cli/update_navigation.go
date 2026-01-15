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
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PageMenuSetCmd(m.CmsMenuInit()))

			return m, tea.Batch(cmds...)
		case ADMINCMSPAGE:
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, PageMenuSetCmd(m.CmsMenuInit()))

			return m, tea.Batch(cmds...)
		case DATABASEPAGE:
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))

			return m, tea.Batch(cmds...)
		case TABLEPAGE:
			cmds = append(cmds, PageMenuSetCmd(m.DatabaseMenuInit()))
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, GetColumnsCmd(*m.Config, m.TableState.Table))

			return m, tea.Batch(cmds...)
		case CREATEPAGE:
			cmds = append(cmds, FormNewCmd(DATABASECREATE))
			cmds = append(cmds, FocusSetCmd(FORMFOCUS))
			cmds = append(cmds, PageSetCmd(m.PageMap[CREATEPAGE]))
			cmds = append(cmds, StatusSetCmd(EDITING))

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
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, &page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case UPDATEFORMPAGE:
			page := m.PageMap[UPDATEFORMPAGE]
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, &page))
			cmds = append(cmds, FormNewCmd(DATABASEUPDATE))
			cmds = append(cmds, StatusSetCmd(EDITING))

			return m, tea.Batch(cmds...)
		case DELETEPAGE:
			page := m.PageMap[DELETEPAGE]
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
			cmds = append(cmds, DatatypesFetchCmd())
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))

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
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case USERSADMIN:
			page := m.PageMap[USERSADMIN]
			cmds = append(cmds, PageSetCmd(page))
			cmds = append(cmds, StatusSetCmd(OK))

			return m, tea.Batch(cmds...)
		case CONFIGPAGE:
			content, err := formatJSON(m.Config)
			if err == nil {
				cmds = append(cmds, SetViewportContentCmd(content))
			} else {
				cmds = append(cmds, SetViewportContentCmd(m.Content))

			}
			cmds = append(cmds, ReadyTrueCmd())

			if len(m.PageMenu) > 0 && m.Cursor < len(m.PageMenu) {
				cmds = append(cmds, PageSetCmd(m.PageMap[CONFIGPAGE]))
			}

			return m, tea.Batch(cmds...)
		}

		return m, tea.Batch()
	case SelectTable:
		cmds := make([]tea.Cmd, 0)
		cmds = append(cmds, NavigateToPageCmd(m.PageMap[TABLEPAGE]))
		cmds = append(cmds, TableSetCmd(m.Tables[m.Cursor]))
		cmds = append(cmds, PageMenuSetCmd(m.DatabaseMenuInit()))

		return m, tea.Batch(cmds...)
	case HistoryPop:
		cmds := make([]tea.Cmd, 0)
		newModel := m
		entry := newModel.PopHistory()
		cmds = append(cmds, PageSetCmd(entry.Page))
		cmds = append(cmds, PageMenuSetCmd(entry.Menu))
		cmds = append(cmds, CursorSetCmd(entry.Cursor))

		return newModel, tea.Batch(cmds...)

	default:
		return m, nil
	}
}
