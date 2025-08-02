package cli

import (
	"fmt"

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
		cmds = append(cmds, HistoryPushCmd(PageHistory{Page: m.Page, Cursor: m.Cursor}))
		cmds = append(cmds, CursorResetCmd())
		cmds = append(cmds, LogMessage(fmt.Sprintln("cursor:", m.Cursor)))
		cmds = append(cmds, LogMessage(fmt.Sprintln("pages", ViewPageMenus(m))))
		switch msg.Page.Index {
		case DATABASEPAGE:
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
			return m, tea.Batch(cmds...)
		case TABLEPAGE:
			cmds = append(cmds, PageMenuSetCmd(TableMenu))
			cmds = append(cmds, PageSetCmd(msg.Page))
			cmds = append(cmds, GetColumnsCmd(*m.Config, m.Table))
			return m, tea.Batch(cmds...)
		case CREATEPAGE:

			return m, tea.Batch(
				FormNewCmd(DATABASECREATE),
				FocusSetCmd(FORMFOCUS),
				PageSetCmd(m.Pages[CREATEPAGE]),
				StatusSetCmd(EDITING),
			)
		case UPDATEPAGE:
			return m, tea.Batch(
				CursorResetCmd(),
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[UPDATEPAGE]),
				StatusSetCmd(OK),
			)
		case READPAGE:
			return m, tea.Batch(
				CursorResetCmd(),
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[READPAGE]),
				StatusSetCmd(OK),
			)
		case DELETEPAGE:
			return m, tea.Batch(
				CursorResetCmd(),
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[DELETEPAGE]),
				StatusSetCmd(DELETING),
			)
		case UPDATEFORMPAGE:
			return m, tea.Batch(
				FormNewCmd(DATABASEUPDATE),
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[UPDATEFORMPAGE]),
				StatusSetCmd(EDITING),
			)
		case READSINGLEPAGE:
			return m, tea.Batch(
				PageSetCmd(m.Pages[READSINGLEPAGE]),
				StatusSetCmd(OK),
			)
		case CONFIGPAGE:
			form, err := formatJSON(m.Config)
			if err == nil {
				m.Content = form
			}
			m.Viewport.SetContent(m.Content)
			m.Ready = true

			if len(m.PageMenu) > 0 && m.Cursor < len(m.PageMenu) {
				cmds = append(cmds, PageSetCmd(*m.PageMenu[m.Cursor]))
			}

			return m, tea.Batch(cmds...)
		}

		return m, nil
	case SelectTable:
		return m, tea.Batch(
			NavigateToPageCmd(m.Pages[TABLEPAGE]),
			TableSetCmd(m.Tables[m.Cursor]),
			PageMenuSetCmd(TableMenu),
		)
	case PageSet:
		newModel := m
		newModel.Page = msg.Page
		return newModel, NewNavUpdate()
	case HistoryPush:
		newModel := m
		newModel.History = append(newModel.History, msg.Page)
		return newModel, NewNavUpdate()

	case HistoryPop:
		newModel := m
		entry := m.PopHistory()
		newModel.PageMenu = m.Page.Children
		return newModel, tea.Batch(
			PageSetCmd(entry.Page),
			CursorSetCmd(entry.Cursor),
		)

	default:
		return m, nil

	}
}
