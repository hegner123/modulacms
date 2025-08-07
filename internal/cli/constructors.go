package cli

import (
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

// Basic action constructors
func ClearScreenCmd() tea.Cmd          { return func() tea.Msg { return ClearScreen{} } }
func TitleFontNextCmd() tea.Cmd        { return func() tea.Msg { return TitleFontNext{} } }
func TitleFontPreviousCmd() tea.Cmd    { return func() tea.Msg { return TitleFontPrevious{} } }
func LogMessageCmd(msg string) tea.Cmd { return func() tea.Msg { return LogMsg{Message: msg} } }

func ReadyTrueCmd() tea.Cmd  { return func() tea.Msg { return ReadyTrue{} } }
func ReadyFalseCmd() tea.Cmd { return func() tea.Msg { return ReadyFalse{} } }

// Data fetching constructors
func TablesFetchCmd() tea.Cmd { return func() tea.Msg { return TablesFetch{} } }
func TablesSetCmd(tables []string) tea.Cmd {
	return func() tea.Msg { return TablesSet{Tables: tables} }
}

// Loading state constructors
func LoadingStartCmd() tea.Cmd { return func() tea.Msg { return LoadingTrue{} } }
func LoadingStopCmd() tea.Cmd  { return func() tea.Msg { return LoadingFalse{} } }

// Cursor control constructors
func CursorUpCmd() tea.Cmd           { return func() tea.Msg { return CursorUp{} } }
func CursorDownCmd() tea.Cmd         { return func() tea.Msg { return CursorDown{} } }
func CursorResetCmd() tea.Cmd        { return func() tea.Msg { return CursorReset{} } }
func CursorSetCmd(index int) tea.Cmd { return func() tea.Msg { return CursorSet{Index: index} } }
func CursorMaxSetCmd(cursorMax int) tea.Cmd {
	return func() tea.Msg { return CursorMaxSet{CursorMax: cursorMax} }
}
func PageModNextCmd() tea.Cmd      { return func() tea.Msg { return PageModNext{} } }
func PageModPreviousCmd() tea.Cmd  { return func() tea.Msg { return PageModPrevious{} } }
func PaginationUpdateCmd() tea.Cmd { return func() tea.Msg { return UpdatePagination{} } }

// Page and navigation constructors
func PageSetCmd(page Page) tea.Cmd     { return func() tea.Msg { return PageSet{Page: page} } }
func TableSetCmd(table string) tea.Cmd { return func() tea.Msg { return TableSet{Table: table} } }
func NavigateToPageCmd(page Page) tea.Cmd {
	return func() tea.Msg { return NavigateToPage{Page: page} }
}

func SelectTableCmd(table string) tea.Cmd { return func() tea.Msg { return SelectTable{Table: table} } }

func SetPageContentCmd(content string) tea.Cmd {
	return func() tea.Msg { return SetPageContent{Content: content} }
}
func SetViewportContentCmd(content string) tea.Cmd {
	return func() tea.Msg { return SetViewportContent{Content: content} }
}

// Focus control constructors
func FocusSetCmd(focus FocusKey) tea.Cmd { return func() tea.Msg { return FocusSet{Focus: focus} } }

// Form control constructors

func FormNewCmd(f FormIndex) tea.Cmd       { return func() tea.Msg { return FormCreate{FormType: f} } }
func FormSetCmd(form huh.Form) tea.Cmd     { return func() tea.Msg { return FormSet{Form: form} } }
func FormSetValuesCmd(v []*string) tea.Cmd { return func() tea.Msg { return FormValuesSet{Values: v} } }
func FormLenSetCmd(formLen int) tea.Cmd {
	return func() tea.Msg { return FormLenSet{FormLen: formLen} }
}
func FormSubmitCmd() tea.Cmd { return func() tea.Msg { return FormSubmitMsg{} } }
func FormCancelCmd() tea.Cmd { return func() tea.Msg { return FormCancelMsg{} } }
func FormActionCmd(action DatabaseCMD, table string, columns []string, values []*string) tea.Cmd {
	return func() tea.Msg {
		// Debug log to trace FormActionCmd execution
		return tea.Batch(
			LogMessageCmd(fmt.Sprintf("FormActionCmd executed: %s action on table %s", action, table)),
			func() tea.Msg {
				return FormActionMsg{
					Action:  action,
					Table:   table,
					Columns: columns,
					Values:  values,
				}
			},
		)()
	}
}
func FormAbortCmd(action DatabaseCMD, table string) tea.Cmd {
	return func() tea.Msg {
		return FormAborted{
			Action: action,
			Table:  table,
		}
	}
}

// History management constructors
func HistoryPopCmd() tea.Cmd { return func() tea.Msg { return HistoryPop{} } }
func HistoryPushCmd(page PageHistory) tea.Cmd {
	return func() tea.Msg { return HistoryPush{Page: page} }
}

// Cms constructors
func DatatypesFetchCmd() tea.Cmd { return func() tea.Msg { return DatatypesFetchMsg{} } }
func DatatypesFetchResultCmd(data []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return DatatypesFetchResultsMsg{Data: data} }
}

// Database operation constructors
func DatabaseGetCmd(source FetchSource, table db.DBTable, id int64) tea.Cmd {
	return func() tea.Msg {
		return DatabaseGetMsg{
			Source: source,
			ID:     id,
			Table:  table,
		}
	}
}
func DatabaseListCmd(source FetchSource, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListMsg{
			Source: source,
			Table:  table,
		}
	}
}
func DatabaseListFilteredCmd(source FetchSource, table db.DBTable, columns []string, whereColumn string, value any) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListFilteredMsg{
			Source:      source,
			Table:       table,
			Columns:     columns,
			WhereColumn: whereColumn,
			Value:       value,
		}
	}
}
func DatabaseGetRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseGetRowMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
func DatabaseListRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListRowsMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
func DatabaseListFilteredRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListFilteredRowsMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
func DatabaseDeleteEntryCmd(id int, table string) tea.Cmd {
	return func() tea.Msg { return DatabaseDeleteEntry{Id: id, Table: table} }
}
func DatabaseInsertCmd(table db.DBTable, columns []string, values []*string) tea.Cmd {
	return func() tea.Msg {
		return DatabaseInsertEntry{
			Table:   table,
			Columns: columns,
			Values:  values,
		}
	}
}
func DatabaseUpdateEntryCmd(table db.DBTable, err error) tea.Cmd {
	return func() tea.Msg { return DatabaseUpdateEntry{Table: table} }
}

// Data management constructors
func ColumnsFetchedCmd(columns *[]string, columnTypes *[]*sql.ColumnType) tea.Cmd {
	return func() tea.Msg { return ColumnsFetched{Columns: columns, ColumnTypes: columnTypes} }
}
func ColumnsSetCmd(columns *[]string) tea.Cmd {
	return func() tea.Msg { return ColumnsSet{Columns: columns} }
}
func ColumnTypesSetCmd(columnTypes *[]*sql.ColumnType) tea.Cmd {
	return func() tea.Msg { return ColumnTypesSet{ColumnTypes: columnTypes} }
}
func HeadersSetCmd(columns []string) tea.Cmd {
	return func() tea.Msg { return HeadersSet{Headers: columns} }
}
func RowsSetCmd(rows [][]string) tea.Cmd {
	return func() tea.Msg { return RowsSet{Rows: rows} }
}
func PaginatorUpdateCmd(perPage, totalPages int) tea.Cmd {
	return func() tea.Msg { return PaginatorUpdate{PerPage: perPage, TotalPages: totalPages} }
}

// Error and status constructors
func ErrorSetCmd(err error) tea.Cmd { return func() tea.Msg { return ErrorSet{Err: err} } }
func StatusSetCmd(status ApplicationState) tea.Cmd {
	return func() tea.Msg { return StatusSet{Status: status} }
}

// Dialog constructors
func DialogSetCmd(dialog *DialogModel) tea.Cmd {
	return func() tea.Msg { return DialogSet{Dialog: dialog} }
}
func DialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg { return DialogActiveSet{DialogActive: active} }
}
func DialogReadyOKSetCmd(ready bool) tea.Cmd {
	return func() tea.Msg { return DialogReadyOKSet{Ready: ready} }
}

// Root and menu constructors
func RootSetCmd(root model.Root) tea.Cmd { return func() tea.Msg { return RootSet{Root: root} } }
func DatatypeMenuSetCmd(menu []string) tea.Cmd {
	return func() tea.Msg { return DatatypeMenuSet{DatatypeMenu: menu} }
}
func PageMenuSetCmd(menu []*Page) tea.Cmd {
	return func() tea.Msg { return PageMenuSet{PageMenu: menu} }
}

// Composite constructors for common operations
func ShowDialogCmd(title, message string, showCancel bool, action DialogAction) tea.Cmd {
	dialog := NewDialog(title, message, showCancel, action)
	return func() tea.Msg {
		return tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)()
	}
}

func HideDialogCmd() tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			DialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)()
	}
}

func SetErrorWithStatusCmd(err error, status ApplicationState) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			ErrorSetCmd(err),
			StatusSetCmd(status),
			LoadingStopCmd(),
		)()
	}
}

func SetFormDataCmd(form huh.Form, formLen int, values []*string) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			FormSetCmd(form),
			FormLenSetCmd(formLen),
			FormSetValuesCmd(values),
			LoadingStopCmd(),
		)()
	}
}

func SetTableDataCmd(headers []string, rows [][]string, maxRows int) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			HeadersSetCmd(headers),
			RowsSetCmd(rows),
			PaginatorUpdateCmd(maxRows, len(rows)),
			LoadingStopCmd(),
		)()
	}
}

func FetchTableHeadersRowsCmd(c config.Config, t string) tea.Cmd {
	return func() tea.Msg {
		return FetchHeadersRows{
			Config: c,
			Table:  t,
		}
	}
}

func TableHeadersRowsFetchedCmd(headers []string, rows [][]string) tea.Cmd {
	return func() tea.Msg {
		return TableHeadersRowsFetchedMsg{
			Headers: headers,
			Rows:    rows,
		}
	}
}

func GetColumnsCmd(c config.Config, t string) tea.Cmd {
	return func() tea.Msg {
		return GetColumns{
			Config: c,
			Table:  t,
		}
	}
}

func DbResultCmd(res sql.Result, table string) tea.Cmd {
	return func() tea.Msg {
		return DbResMsg{
			Result: res,
			Table:  table,
		}
	}
}

func GetFullTreeResCMD(s string, rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	return func() tea.Msg {
		return GetFullTreeResMsg{
			Rows:    rows,
			Content: s,
		}
	}

}
func BuildTreeFromRowsCmd(rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRows{
			Rows: rows,
		}
	}
}

func DatabaseTreeCMD() tea.Cmd {
	return func() tea.Msg {
		return DatabaseTreeMsg{}
	}
}

// Tree node state management constructors
func SetExpandCmd(nodeID int64, expand bool) tea.Cmd {
	return func() tea.Msg {
		return SetExpandMsg{
			NodeID: nodeID,
			Expand: expand,
		}
	}
}

func SetWrappedCmd(nodeID int64, wrapped int) tea.Cmd {
	return func() tea.Msg {
		return SetWrappedMsg{
			NodeID:  nodeID,
			Wrapped: wrapped,
		}
	}
}

func SetIndentCmd(nodeID int64, indent int) tea.Cmd {
	return func() tea.Msg {
		return SetIndentMsg{
			NodeID: nodeID,
			Indent: indent,
		}
	}
}

func ExpandNodeCmd(nodeID int64, recursive bool) tea.Cmd {
	return func() tea.Msg {
		return ExpandNodeMsg{
			NodeID:    nodeID,
			Recursive: recursive,
		}
	}
}

func CollapseNodeCmd(nodeID int64, recursive bool) tea.Cmd {
	return func() tea.Msg {
		return CollapseNodeMsg{
			NodeID:    nodeID,
			Recursive: recursive,
		}
	}
}

func ToggleExpandCmd(nodeID int64) tea.Cmd {
	return func() tea.Msg {
		return ToggleExpandMsg{
			NodeID: nodeID,
		}
	}
}

func SetExpandForTypeCmd(nodeType string, expand bool) tea.Cmd {
	return func() tea.Msg {
		return SetExpandForTypeMsg{
			NodeType: nodeType,
			Expand:   expand,
		}
	}
}

func CalculateDepthsCmd() tea.Cmd {
	return func() tea.Msg {
		return CalculateDepthsMsg{}
	}
}

func SetDepthsFromParentCmd(parentID int64, startDepth int) tea.Cmd {
	return func() tea.Msg {
		return SetDepthsFromParentMsg{
			ParentID:   parentID,
			StartDepth: startDepth,
		}
	}
}

func GetNodeStateCmd(nodeID int64) tea.Cmd {
	return func() tea.Msg {
		return GetNodeStateMsg{
			NodeID: nodeID,
		}
	}
}

func SetNodeStateCmd(nodeID int64, state NodeState) tea.Cmd {
	return func() tea.Msg {
		return SetNodeStateMsg{
			NodeID: nodeID,
			State:  state,
		}
	}
}

func ApplyStatesCmd(states []NodeState) tea.Cmd {
	return func() tea.Msg {
		return ApplyStatesMsg{
			States: states,
		}
	}
}

func GetAllNodeStatesCmd() tea.Cmd {
	return func() tea.Msg {
		return GetAllNodeStatesMsg{}
	}
}

func InitializeViewStatesCmd() tea.Cmd {
	return func() tea.Msg {
		return InitializeViewStatesMsg{}
	}
}

// Response constructors for tree operations
func NodeStateResponseCmd(state *NodeState, exists bool) tea.Cmd {
	return func() tea.Msg {
		return NodeStateResponseMsg{
			State:  state,
			Exists: exists,
		}
	}
}

func AllNodeStatesResponseCmd(states []NodeState) tea.Cmd {
	return func() tea.Msg {
		return AllNodeStatesResponseMsg{
			States: states,
		}
	}
}

func TreeOperationResultCmd(success bool, nodeID int64) tea.Cmd {
	return func() tea.Msg {
		return TreeOperationResultMsg{
			Success: success,
			NodeID:  nodeID,
		}
	}
}

// Bulk tree operation constructors
func BulkExpandCmd(nodeIDs []int64, expand bool, recursive bool) tea.Cmd {
	return func() tea.Msg {
		return BulkExpandMsg{
			NodeIDs:   nodeIDs,
			Expand:    expand,
			Recursive: recursive,
		}
	}
}

func BulkSetWrappedCmd(nodeUpdates []struct {
	NodeID  int64
	Wrapped int
}) tea.Cmd {
	return func() tea.Msg {
		return BulkSetWrappedMsg{
			NodeUpdates: nodeUpdates,
		}
	}
}

func BulkSetIndentCmd(nodeUpdates []struct {
	NodeID int64
	Indent int
}) tea.Cmd {
	return func() tea.Msg {
		return BulkSetIndentMsg{
			NodeUpdates: nodeUpdates,
		}
	}
}

// Composite tree operation constructors
func ExpandSectionCmd(nodeID int64) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			LogMessageCmd(fmt.Sprintf("Expanding section: %d", nodeID)),
			ExpandNodeCmd(nodeID, true),
		)()
	}
}

func CollapseSectionCmd(nodeID int64) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			LogMessageCmd(fmt.Sprintf("Collapsing section: %d", nodeID)),
			CollapseNodeCmd(nodeID, true),
		)()
	}
}

func InitializeTreeViewCmd() tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			LogMessageCmd("Initializing tree view states"),
			CalculateDepthsCmd(),
			InitializeViewStatesCmd(),
		)()
	}
}

func SetTreeViewModeCmd(mode string) tea.Cmd {
	return func() tea.Msg {
		switch mode {
		case "expanded":
			return tea.Batch(
				LogMessageCmd("Setting tree view mode: expanded"),
				SetExpandForTypeCmd("Container", true),
				SetExpandForTypeCmd("Row", true),
			)()
		case "collapsed":
			return tea.Batch(
				LogMessageCmd("Setting tree view mode: collapsed"),
				SetExpandForTypeCmd("Card", false),
				SetExpandForTypeCmd("Column", false),
				SetExpandForTypeCmd("RichText", false),
			)()
		case "structure":
			return tea.Batch(
				LogMessageCmd("Setting tree view mode: structure only"),
				SetExpandForTypeCmd("Navigation", true),
				SetExpandForTypeCmd("Hero", true),
				SetExpandForTypeCmd("Container", true),
				SetExpandForTypeCmd("Footer", true),
				SetExpandForTypeCmd("Card", false),
				SetExpandForTypeCmd("RichText", false),
			)()
		default:
			return LogMessageCmd(fmt.Sprintf("Unknown tree view mode: %s", mode))
		}
	}
}

func SaveTreeStateCmd(states []NodeState) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			LogMessageCmd(fmt.Sprintf("Saving tree state for %d nodes", len(states))),
			ApplyStatesCmd(states),
		)()
	}
}

func RestoreTreeStateCmd(states []NodeState) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			LogMessageCmd(fmt.Sprintf("Restoring tree state for %d nodes", len(states))),
			ApplyStatesCmd(states),
		)()
	}
}
