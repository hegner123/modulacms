package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/tui"
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
		newModel.TableState.Table = msg.Table
		return newModel, NewStateUpdate()
	case TablesSet:
		newModel := m
		newModel.Tables = msg.Tables
		return newModel, tea.Batch(NewStateUpdate(), LoadingStopCmd())
	case ColumnsSet:
		newModel := m
		newModel.TableState.Columns = msg.Columns
		return newModel, NewStateUpdate()
	case ColumnTypesSet:
		newModel := m
		newModel.TableState.ColumnTypes = msg.ColumnTypes
		return newModel, NewStateUpdate()
	case HeadersSet:
		newModel := m
		newModel.TableState.Headers = msg.Headers
		return newModel, NewStateUpdate()
	case RowsSet:
		newModel := m
		newModel.TableState.Rows = msg.Rows
		return newModel, NewStateUpdate()
	case LogMsg:
		m.Logger.Finfo(msg.Message)
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
	case FormDialogSetMsg:
		newModel := m
		newModel.FormDialog = msg.Dialog
		return newModel, NewStateUpdate()
	case FormDialogActiveSetMsg:
		newModel := m
		newModel.FormDialogActive = msg.Active
		return newModel, NewStateUpdate()
	case ContentFormDialogSetMsg:
		newModel := m
		newModel.ContentFormDialog = msg.Dialog
		return newModel, NewStateUpdate()
	case ContentFormDialogActiveSetMsg:
		newModel := m
		newModel.ContentFormDialogActive = msg.Active
		return newModel, NewStateUpdate()
	case DatatypeMenuSet:
		newModel := m
		newModel.DatatypeMenu = msg.DatatypeMenu
		return newModel, NewStateUpdate()
	case RoutesSet:
		newModel := m
		newModel.Routes = msg.Routes
		return newModel, NewStateUpdate()
	case MediaListSet:
		newModel := m
		newModel.MediaList = msg.MediaList
		return newModel, NewStateUpdate()
	case UsersListSet:
		newModel := m
		newModel.UsersList = msg.UsersList
		return newModel, NewStateUpdate()
	case RootContentSummarySet:
		newModel := m
		newModel.RootContentSummary = msg.RootContentSummary
		return newModel, NewStateUpdate()
	case RootDatatypesSet:
		newModel := m
		newModel.RootDatatypes = msg.RootDatatypes
		return newModel, NewStateUpdate()
	case AllDatatypesSet:
		newModel := m
		newModel.AllDatatypes = msg.AllDatatypes
		return newModel, NewStateUpdate()
	case DatatypeFieldsSet:
		newModel := m
		newModel.SelectedDatatypeFields = msg.Fields
		return newModel, NewStateUpdate()
	case SelectedDatatypeSet:
		newModel := m
		newModel.SelectedDatatype = msg.DatatypeID
		return newModel, NewStateUpdate()
	case PanelFocusReset:
		newModel := m
		newModel.PanelFocus = tui.TreePanel
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
		newModel.FormState.FormLen = msg.FormLen
		return newModel, NewStateUpdate()
	case FormMapSet:
		newModel := m
		newModel.FormState.FormMap = msg.FormMap
		return newModel, NewStateUpdate()
	case FormSet:
		newModel := m
		newModel.FormState.Form = &msg.Form
		newModel.FormState.FormValues = msg.Values
		return newModel, NewStateUpdate()
	case FormValuesSet:
		newModel := m
		newModel.FormState.FormValues = msg.Values
		return newModel, NewStateUpdate()
	case FormOptionsSet:
		newModel := m
		newModel.FormState.FormOptions = msg.Options
		return newModel, NewStateUpdate()

	case UpdateMaxCursorMsg:
		cursorUpdate := func() tea.Msg {
			if m.Cursor > msg.CursorMax-1 {
				return CursorSet{Index: msg.CursorMax - 1}
			}
			return nil
		}()

		cmds := []tea.Cmd{
			CursorMaxSetCmd(msg.CursorMax),
		}

		if cursorUpdate != nil {
			cmds = append(cmds, func() tea.Msg { return cursorUpdate })
		}

		return m, tea.Batch(cmds...)
	default:
		return m, nil
	}
}
