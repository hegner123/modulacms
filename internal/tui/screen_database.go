package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// 3/9 grid: left = tables list, right = rows (top) + row detail (bottom)
var databaseGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Tables"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.55, Title: "Rows"},
			{Height: 0.45, Title: "Detail"},
		}},
	},
}

// DatabaseScreen implements Screen for DATABASEPAGE.
// Single unified view: left=table list, top-right=paginated rows,
// bottom-right=selected row detail (key-value). Actions via keys:
// enter=focus detail, e=edit dialog, d=delete confirm, n=create dialog.
type DatabaseScreen struct {
	GridScreen
	Tables     []string
	TableState *TableModel

	// Row cursor (within current page)
	RowCursor int

	// Detail cursor (for scrolling column values when detail is focused)
	DetailCursor int

	// Pagination
	PageMod int
	MaxRows int

	Loading bool
}

// NewDatabaseScreen creates a DatabaseScreen.
func NewDatabaseScreen(tables []string, tableState *TableModel) *DatabaseScreen {
	cursorMax := len(tables) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &DatabaseScreen{
		GridScreen: GridScreen{
			Grid:      databaseGrid,
			CursorMax: cursorMax,
		},
		Tables:     tables,
		TableState: tableState,
		MaxRows:    10,
	}
}

func (s *DatabaseScreen) PageIndex() PageIndex { return DATABASEPAGE }

func (s *DatabaseScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		switch s.FocusIndex {
		case 0: // Tables list
			if km.Matches(key, config.ActionUp) {
				if s.Cursor > 0 {
					s.Cursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				if s.Cursor < len(s.Tables)-1 {
					s.Cursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				return s, HistoryPopCmd()
			}
			if km.Matches(key, config.ActionSelect) {
				if len(s.Tables) > 0 && s.Cursor < len(s.Tables) {
					s.TableState.Table = s.Tables[s.Cursor]
					s.RowCursor = 0
					s.PageMod = 0
					s.FocusIndex = 1
					return s, tea.Batch(
						GetColumnsCmd(*ctx.Config, s.TableState.Table),
						FetchTableHeadersRowsCmd(*ctx.Config, s.TableState.Table, nil),
					)
				}
			}

		case 1: // Rows
			pageSize := s.currentPageSize()
			if km.Matches(key, config.ActionUp) {
				if s.RowCursor > 0 {
					s.RowCursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				if s.RowCursor < pageSize-1 {
					s.RowCursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionPagePrev) {
				if s.PageMod > 0 {
					s.PageMod--
					s.RowCursor = 0
				}
				return s, nil
			}
			if km.Matches(key, config.ActionPageNext) {
				if len(s.TableState.Rows) > 0 && s.PageMod < (len(s.TableState.Rows)-1)/s.MaxRows {
					s.PageMod++
					s.RowCursor = 0
				}
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.FocusIndex = 0
				return s, nil
			}
			// Enter: focus detail panel for scrolling
			if km.Matches(key, config.ActionSelect) {
				s.DetailCursor = 0
				s.FocusIndex = 2
				return s, nil
			}
			// e: edit dialog
			if key == "e" {
				rowID := s.currentRowID()
				if rowID != "" {
					recordIndex := (s.PageMod * s.MaxRows) + s.RowCursor
					return s, tea.Sequence(
						CursorSetCmd(recordIndex),
						ShowDatabaseUpdateDialogCmd(db.DBTable(s.TableState.Table), rowID),
					)
				}
				return s, nil
			}
			// d: delete confirm
			if key == "d" {
				rowID := s.currentRowID()
				if rowID != "" {
					recordIndex := (s.PageMod * s.MaxRows) + s.RowCursor
					return s, tea.Sequence(
						CursorSetCmd(recordIndex),
						ShowDialogCmd("Confirm Delete",
							"Are you sure you want to delete this record? This action cannot be undone.", true, DIALOGDELETE),
					)
				}
				return s, nil
			}
			// n: create/insert dialog
			if key == "n" {
				if s.TableState.Table != "" {
					return s, ShowDatabaseInsertDialogCmd(db.DBTable(s.TableState.Table))
				}
				return s, nil
			}

		case 2: // Detail (scrollable key-value view)
			detailMax := s.detailLineCount() - 1
			if detailMax < 0 {
				detailMax = 0
			}
			if km.Matches(key, config.ActionUp) {
				if s.DetailCursor > 0 {
					s.DetailCursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				if s.DetailCursor < detailMax {
					s.DetailCursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.FocusIndex = 1
				return s, nil
			}
			// e/d/n work from detail panel too
			if key == "e" {
				rowID := s.currentRowID()
				if rowID != "" {
					recordIndex := (s.PageMod * s.MaxRows) + s.RowCursor
					return s, tea.Sequence(
						CursorSetCmd(recordIndex),
						ShowDatabaseUpdateDialogCmd(db.DBTable(s.TableState.Table), rowID),
					)
				}
				return s, nil
			}
			if key == "d" {
				rowID := s.currentRowID()
				if rowID != "" {
					recordIndex := (s.PageMod * s.MaxRows) + s.RowCursor
					return s, tea.Sequence(
						CursorSetCmd(recordIndex),
						ShowDialogCmd("Confirm Delete",
							"Are you sure you want to delete this record? This action cannot be undone.", true, DIALOGDELETE),
					)
				}
				return s, nil
			}
			if key == "n" {
				if s.TableState.Table != "" {
					return s, ShowDatabaseInsertDialogCmd(db.DBTable(s.TableState.Table))
				}
				return s, nil
			}
		}

		if km.Matches(key, config.ActionQuit) {
			return s, tea.Quit
		}

	// Fetch request messages
	case TablesFetch:
		return s, GetTablesCMD(ctx.Config)
	case FetchHeadersRows:
		t := msg.Table
		dbt := db.StringDBTable(t)
		d := db.ConfigDB(msg.Config)
		columns := db.GenericHeaders(dbt)
		if columns == nil {
			return s, ErrorSetCmd(fmt.Errorf("unknown table: %s", t))
		}
		listRows, err := db.GenericList(dbt, d)
		if err != nil {
			return s, ErrorSetCmd(err)
		}
		return s, tea.Batch(
			TableHeadersRowsFetchedCmd(columns, listRows, msg.Page),
			LogMessageCmd(fmt.Sprintf("Table %s headers fetched", s.TableState.Table)),
		)
	case TableHeadersRowsFetchedMsg:
		s.TableState.Headers = msg.Headers
		s.TableState.Rows = msg.Rows
		s.RowCursor = 0
		s.PageMod = 0
		s.Loading = false
		return s, nil
	case GetColumns:
		dbt := db.StringDBTable(msg.Table)
		clm := db.GenericHeaders(dbt)
		if clm == nil {
			return s, ErrorSetCmd(fmt.Errorf("unknown table: %s", msg.Table))
		}
		return s, ColumnInfoSetCmd(&clm, nil)
	case ColumnsFetched:
		s.TableState.Columns = msg.Columns
		s.TableState.ColumnTypes = msg.ColumnTypes
		s.Loading = false
		return s, nil
	case DatatypesFetchMsg:
		return s, tea.Batch(
			LoadingStartCmd(),
			DatabaseListCmd(DATATYPEMENU, db.Datatype),
		)

	// Data refresh messages
	case TablesSet:
		s.Tables = msg.Tables
		s.CursorMax = len(s.Tables) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		s.Loading = false
		return s, nil
	case HeadersSet:
		s.TableState.Headers = msg.Headers
		return s, nil
	case RowsSet:
		s.TableState.Rows = msg.Rows
		return s, nil
	case TableSet:
		s.TableState.Table = msg.Table
		return s, nil
	case ColumnInfoSetMsg:
		s.TableState.Columns = msg.Columns
		s.TableState.ColumnTypes = msg.ColumnTypes
		return s, nil
	case SetLoadingMsg:
		s.Loading = msg.Loading
		return s, nil
	}

	return s, nil
}

func (s *DatabaseScreen) KeyHints(km config.KeyMap) []KeyHint {
	switch s.FocusIndex {
	case 1: // Rows
		hints := []KeyHint{
			{km.HintString(config.ActionSelect), "detail"},
			{"e", "edit"},
			{"d", "delete"},
			{"n", "new"},
		}
		if len(s.TableState.Rows) > s.MaxRows {
			hints = append(hints, KeyHint{km.HintString(config.ActionPagePrev) + "/" + km.HintString(config.ActionPageNext), "page"})
		}
		hints = append(hints,
			KeyHint{km.HintString(config.ActionNextPanel), "panel"},
			KeyHint{km.HintString(config.ActionBack), "back"},
		)
		return hints
	case 2: // Detail
		return []KeyHint{
			{"e", "edit"},
			{"d", "delete"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "scroll"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "rows"},
		}
	default: // Tables
		return []KeyHint{
			{km.HintString(config.ActionSelect), "select"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	}
}

func (s *DatabaseScreen) View(ctx AppContext) string {
	tablesInnerH := s.Grid.CellInnerHeight(0, ctx.Height)
	rowsInnerH := s.Grid.CellInnerHeight(1, ctx.Height)
	detailInnerH := s.Grid.CellInnerHeight(2, ctx.Height)

	rowsContent := s.renderRows()
	rowsTotalLines := strings.Count(rowsContent, "\n") + 1

	detailContent := s.renderDetail()
	detailTotalLines := strings.Count(detailContent, "\n") + 1

	cells := []CellContent{
		{Content: s.renderTables(), TotalLines: len(s.Tables), ScrollOffset: ClampScroll(s.Cursor, len(s.Tables), tablesInnerH)},
		{Content: rowsContent, TotalLines: rowsTotalLines, ScrollOffset: ClampScroll(s.rowScrollLine(), rowsTotalLines, rowsInnerH)},
		{Content: detailContent, TotalLines: detailTotalLines, ScrollOffset: ClampScroll(s.detailScrollLine(), detailTotalLines, detailInnerH)},
	}
	return s.RenderGrid(ctx, cells)
}

// --- Helpers ---

// currentPageSize returns the number of rows visible on the current page.
func (s *DatabaseScreen) currentPageSize() int {
	if len(s.TableState.Rows) == 0 {
		return 0
	}
	start := s.PageMod * s.MaxRows
	end := start + s.MaxRows
	if end > len(s.TableState.Rows) {
		end = len(s.TableState.Rows)
	}
	return end - start
}

// currentRowID returns the first column (ID) of the currently selected row.
func (s *DatabaseScreen) currentRowID() string {
	recordIndex := (s.PageMod * s.MaxRows) + s.RowCursor
	if recordIndex >= len(s.TableState.Rows) || len(s.TableState.Rows[recordIndex]) == 0 {
		return ""
	}
	return s.TableState.Rows[recordIndex][0]
}

// detailLineCount returns the number of header columns (one line per column in detail).
func (s *DatabaseScreen) detailLineCount() int {
	return len(s.TableState.Headers)
}

// rowScrollLine converts RowCursor to a line offset for scroll.
// 2-line header + 1 line per row.
func (s *DatabaseScreen) rowScrollLine() int {
	return 2 + s.RowCursor
}

// detailScrollLine converts DetailCursor to a line offset for scroll.
// 2-line header + 2 lines per column (label + value).
func (s *DatabaseScreen) detailScrollLine() int {
	return 2 + s.DetailCursor*2
}

// --- Render methods ---

// renderTables renders the table list for the left panel.
func (s *DatabaseScreen) renderTables() string {
	if len(s.Tables) == 0 {
		return " (no tables)"
	}
	lines := make([]string, 0, len(s.Tables))
	for i, tbl := range s.Tables {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		active := ""
		if s.TableState.Table == tbl {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, tbl, active))
	}
	return strings.Join(lines, "\n")
}

// renderRows renders paginated table rows for the top-right cell.
func (s *DatabaseScreen) renderRows() string {
	if s.TableState.Table == "" {
		return " Select a table"
	}
	if s.Loading {
		return " Loading..."
	}
	if len(s.TableState.Headers) == 0 {
		return " No data loaded"
	}

	start := s.PageMod * s.MaxRows
	end := start + s.MaxRows
	if end > len(s.TableState.Rows) {
		end = len(s.TableState.Rows)
	}
	if start >= len(s.TableState.Rows) {
		return " (no rows)"
	}

	currentView := s.TableState.Rows[start:end]

	lines := make([]string, 0, len(currentView)+4)

	// Header row
	headerLine := "   "
	for _, h := range s.TableState.Headers {
		if len(h) > 15 {
			h = h[:12] + "..."
		}
		headerLine += fmt.Sprintf("%-16s", h)
	}
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(headerLine))
	lines = append(lines, "")

	for i, row := range currentView {
		cursor := "   "
		if s.RowCursor == i {
			cursor = " ->"
		}
		rowLine := cursor
		for _, cell := range row {
			if len(cell) > 15 {
				cell = cell[:12] + "..."
			}
			rowLine += fmt.Sprintf("%-16s", cell)
		}
		lines = append(lines, rowLine)
	}

	// Pagination indicator
	if len(s.TableState.Rows) > s.MaxRows {
		totalPages := (len(s.TableState.Rows)-1)/s.MaxRows + 1
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  Page %d/%d  (%d rows)", s.PageMod+1, totalPages, len(s.TableState.Rows)))
	}

	return strings.Join(lines, "\n")
}

// renderDetail renders the selected row as key-value pairs for the bottom-right cell.
func (s *DatabaseScreen) renderDetail() string {
	if s.TableState.Table == "" {
		return " Database\n\n  Select a table to browse rows."
	}
	if len(s.TableState.Rows) == 0 || len(s.TableState.Headers) == 0 {
		return " No row selected"
	}

	recordIndex := (s.PageMod * s.MaxRows) + s.RowCursor
	if recordIndex >= len(s.TableState.Rows) {
		return " No row selected"
	}

	row := s.TableState.Rows[recordIndex]
	lines := make([]string, 0, len(s.TableState.Headers)*2+4)
	lines = append(lines, fmt.Sprintf(" Row %d", recordIndex+1))
	lines = append(lines, "")

	for i, header := range s.TableState.Headers {
		value := ""
		if i < len(row) {
			value = row[i]
		}

		cursor := " "
		if s.FocusIndex == 2 && s.DetailCursor == i {
			cursor = ">"
		}
		lines = append(lines, fmt.Sprintf(" %s %s", cursor, lipgloss.NewStyle().Bold(true).Render(header)))
		lines = append(lines, fmt.Sprintf("     %s", value))
	}

	return strings.Join(lines, "\n")
}

// setDatabaseModeCmd creates a command that sets the database mode on the Model.
func setDatabaseModeCmd(mode DatabaseMode) tea.Cmd {
	return func() tea.Msg {
		return SetDatabaseModeMsg{Mode: mode}
	}
}
