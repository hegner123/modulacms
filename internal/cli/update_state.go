package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/utility"
)

type StateUpdated struct{}

func NewStateUpdate() tea.Cmd {
	return func() tea.Msg {
		return StateUpdated{}
	}

}

func (m Model) UpdateState(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case LoadingTrue:
		newModel := m
		newModel.Loading = true
		return newModel, NewStateUpdate()
	case LoadingFalse:
		newModel := m
		newModel.Loading = false
		return newModel, NewStateUpdate()
	case CursorUp:
		newModel := m
		newModel.Cursor = m.Cursor - 1
		return newModel, NewStateUpdate()
	case CursorDown:
		newModel := m
		newModel.Cursor = m.Cursor + 1
		return newModel, NewStateUpdate()
	case CursorReset:
		newModel := m
		newModel.Cursor = 0
		return newModel, NewStateUpdate()
	case CursorSet:
		newModel := m
		newModel.Cursor = msg.Index
		return newModel, NewStateUpdate()
	case PageModNext:
		newModel := m
		newModel.PageMod = m.PageMod + 1
		return newModel, NewStateUpdate()
	case PageModPrevious:
		newModel := m
		newModel.PageMod = m.PageMod - 1
		return newModel, NewStateUpdate()
	case FocusSet:
		newModel := m
		newModel.Focus = msg.Focus
		return newModel, NewStateUpdate()
	case TitleFontNext:
		newModel := m
		if newModel.TitleFont < len(m.Titles)-1 {
			newModel.TitleFont++
		}
		return newModel, NewStateUpdate()
	case TitleFontPrevious:
		newModel := m
		if newModel.TitleFont > 0 {
			newModel.TitleFont--
		}
		return newModel, NewStateUpdate()
	case PageSet:
		newModel := m
		newModel.Page = msg.Page
		return newModel, NewStateUpdate()
	case HistoryPush:
		newModel := m
		newModel.History = append(newModel.History, msg.Page)
		return newModel, NewStateUpdate()
	case TableSet:
		newModel := m
		newModel.Table = msg.Table
		return newModel, NewStateUpdate()
	case TablesSet:
		newModel := m
		newModel.Tables = msg.Tables
		return newModel, NewStateUpdate()
	case ColumnsSet:
		newModel := m
		newModel.Columns = msg.Columns
		return newModel, NewStateUpdate()
	case ColumnTypesSet:
		newModel := m
		newModel.ColumnTypes = msg.ColumnTypes
		return newModel, NewStateUpdate()
	case HeadersSet:
		newModel := m
		newModel.Headers = msg.Headers
		return newModel, NewStateUpdate()
	case RowsSet:
		newModel := m
		newModel.Rows = msg.Rows
		return newModel, NewStateUpdate()
	case LogMsg:
		utility.DefaultLogger.Finfo(msg.Message)
		return m, NewStateUpdate()
	case ClearScreen:
		return m, tea.ClearScreen
	case SetPageContent:
		newModel := m
		newModel.Content = msg.Content
		return newModel, NewStateUpdate()
	case SetViewportContent:
		newModel := m
		newModel.Viewport.SetContent(msg.Content)
		return newModel, NewStateUpdate()
	case CursorMaxSet:
		newModel := m
		newModel.CursorMax = msg.CursorMax
		return newModel, NewStateUpdate()
	case PaginatorUpdate:
		newModel := m
		newModel.Paginator.PerPage = msg.PerPage
		newModel.Paginator.SetTotalPages(msg.TotalPages)
		return newModel, NewStateUpdate()
	case ErrorSet:
		newModel := m
		newModel.Err = msg.Err
		return newModel, LogMessageCmd(msg.Err.Error())
	case StatusSet:
		newModel := m
		newModel.Status = msg.Status
		return newModel, NewStateUpdate()
	case DialogSet:
		newModel := m
		newModel.Dialog = msg.Dialog
		return newModel, NewStateUpdate()
	case DialogActiveSet:
		newModel := m
		newModel.DialogActive = msg.DialogActive
		return newModel, NewStateUpdate()
	case DatatypeMenuSet:
		newModel := m
		newModel.DatatypeMenu = msg.DatatypeMenu
		return newModel, NewStateUpdate()
	case PageMenuSet:
		newModel := m
		newModel.PageMenu = msg.PageMenu
		return newModel, NewStateUpdate()
	case UpdatePagination:
		p, cmd := m.Paginator.Update(msg)
		newModel := m
		newModel.Paginator = p
		return newModel, cmd
	case FormLenSet:
		newModel := m
		newModel.FormLen = msg.FormLen
		return newModel, NewStateUpdate()
	case FormSet:
		newModel := m
		newModel.Form = &msg.Form
		newModel.FormValues = msg.Values
		return newModel, NewStateUpdate()
	case FormValuesSet:
		newModel := m
		newModel.FormValues = msg.Values
		return newModel, NewStateUpdate()
	case FormOptionsSet:
		newModel := m
		return newModel, NewStateUpdate()

	case UpdateMaxCursorMsg:
		cursorUpdate := func() tea.Msg {
			if m.Cursor > msg.cursorMax-1 {
				return CursorSet{Index: msg.cursorMax - 1}
			}
			return nil
		}()

		cmds := []tea.Cmd{
			CursorMaxSetCmd(msg.cursorMax),
		}

		if cursorUpdate != nil {
			cmds = append(cmds, func() tea.Msg { return cursorUpdate })
		}

		return m, tea.Batch(cmds...)
	default:
		return m, nil
	}
}
