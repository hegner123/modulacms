package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// DatabaseScreen implements Screen for both DATABASEPAGE and READPAGE.
// DATABASEPAGE: 2-panel layout — left=tables list, center=CRUD actions.
// READPAGE: 3-panel layout — left=mode selector, center=paginated rows, right=row detail.
type DatabaseScreen struct {
	// Which sub-page is active
	ReadMode bool // false = DATABASEPAGE, true = READPAGE

	// Shared state
	PanelFocus   FocusPanel
	Tables       []string
	TableState   *TableModel
	DatabaseMode DatabaseMode

	// DATABASEPAGE cursors
	Cursor      int // tables list cursor (TreePanel) or row cursor (READPAGE ContentPanel)
	FieldCursor int // CRUD actions cursor (DATABASEPAGE ContentPanel) or mode selector (READPAGE TreePanel)

	// READPAGE pagination
	PageMod int
	MaxRows int

	// Loading indicator
	Loading bool

	// PageMap reference for navigation
	PageMap map[PageIndex]Page
}

// NewDatabaseScreen creates a DatabaseScreen in DATABASEPAGE mode.
func NewDatabaseScreen(tables []string, tableState *TableModel, databaseMode DatabaseMode, pageMap map[PageIndex]Page) *DatabaseScreen {
	return &DatabaseScreen{
		ReadMode:     false,
		PanelFocus:   TreePanel,
		Tables:       tables,
		TableState:   tableState,
		DatabaseMode: databaseMode,
		Cursor:       0,
		FieldCursor:  0,
		PageMod:      0,
		MaxRows:      10,
		Loading:      false,
		PageMap:      pageMap,
	}
}

// NewDatabaseReadScreen creates a DatabaseScreen in READPAGE mode.
func NewDatabaseReadScreen(tables []string, tableState *TableModel, databaseMode DatabaseMode, pageMap map[PageIndex]Page) *DatabaseScreen {
	return &DatabaseScreen{
		ReadMode:     true,
		PanelFocus:   TreePanel,
		Tables:       tables,
		TableState:   tableState,
		DatabaseMode: databaseMode,
		Cursor:       0,
		FieldCursor:  int(databaseMode),
		PageMod:      0,
		MaxRows:      10,
		Loading:      false,
		PageMap:      pageMap,
	}
}

func (s *DatabaseScreen) PageIndex() PageIndex {
	if s.ReadMode {
		return READPAGE
	}
	return DATABASEPAGE
}

func (s *DatabaseScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	if s.ReadMode {
		return s.updateReadPage(ctx, msg)
	}
	return s.updateDatabasePage(ctx, msg)
}

// updateDatabasePage handles Update for DATABASEPAGE mode.
func (s *DatabaseScreen) updateDatabasePage(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Quit
		if km.Matches(key, config.ActionQuit) {
			return s, tea.Quit
		}

		// ESC: move focus to TreePanel first, then quit
		if km.Matches(key, config.ActionDismiss) {
			if s.PanelFocus != TreePanel {
				s.PanelFocus = TreePanel
				return s, nil
			}
			return s, tea.Quit
		}

		// Panel navigation
		if km.Matches(key, config.ActionNextPanel) {
			s.PanelFocus = (s.PanelFocus + 1) % 3
			if s.PanelFocus == ContentPanel {
				s.FieldCursor = 0
			}
			return s, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			s.PanelFocus = (s.PanelFocus + 2) % 3
			if s.PanelFocus == ContentPanel {
				s.FieldCursor = 0
			}
			return s, nil
		}

		switch s.PanelFocus {
		case TreePanel:
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
			if km.Matches(key, config.ActionBack) {
				return s, HistoryPopCmd()
			}
			if km.Matches(key, config.ActionSelect) {
				if len(s.Tables) > 0 && s.Cursor < len(s.Tables) {
					s.TableState.Table = s.Tables[s.Cursor]
					s.PanelFocus = ContentPanel
					s.FieldCursor = 0
					return s, GetColumnsCmd(*ctx.Config, s.TableState.Table)
				}
			}

		case ContentPanel:
			if km.Matches(key, config.ActionUp) {
				if s.FieldCursor > 0 {
					s.FieldCursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				// 4 CRUD actions: Create(0), Read(1), Update(2), Delete(3)
				if s.FieldCursor < 3 {
					s.FieldCursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionBack) {
				s.PanelFocus = TreePanel
				return s, nil
			}
			if km.Matches(key, config.ActionSelect) {
				if s.TableState.Table == "" {
					return s, nil
				}
				switch s.FieldCursor {
				case 0: // Create
					return s, ShowDatabaseInsertDialogCmd(db.DBTable(s.TableState.Table))
				case 1: // Read
					return s, tea.Sequence(
						setDatabaseModeCmd(DBModeRead),
						NavigateToPageCmd(s.PageMap[READPAGE]),
					)
				case 2: // Update
					return s, tea.Sequence(
						setDatabaseModeCmd(DBModeUpdate),
						NavigateToPageCmd(s.PageMap[READPAGE]),
					)
				case 3: // Delete
					return s, tea.Sequence(
						setDatabaseModeCmd(DBModeDelete),
						NavigateToPageCmd(s.PageMap[READPAGE]),
					)
				}
			}
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

// updateReadPage handles Update for READPAGE mode.
func (s *DatabaseScreen) updateReadPage(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Quit
		if km.Matches(key, config.ActionQuit) {
			return s, tea.Quit
		}

		// Panel navigation
		if km.Matches(key, config.ActionNextPanel) {
			s.PanelFocus = (s.PanelFocus + 1) % 3
			return s, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			s.PanelFocus = (s.PanelFocus + 2) % 3
			return s, nil
		}

		switch s.PanelFocus {
		case TreePanel:
			// Mode selector: 3 items (Read=0, Update=1, Delete=2)
			if km.Matches(key, config.ActionUp) {
				if s.FieldCursor > 0 {
					s.FieldCursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				if s.FieldCursor < 2 {
					s.FieldCursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionSelect) {
				s.DatabaseMode = DatabaseMode(s.FieldCursor)
				s.PanelFocus = ContentPanel
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				return s, HistoryPopCmd()
			}

		case ContentPanel:
			if km.Matches(key, config.ActionUp) {
				if s.Cursor > 0 {
					s.Cursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				// Bound cursor by visible rows on current page
				pageEnd := (s.PageMod + 1) * s.MaxRows
				if pageEnd > len(s.TableState.Rows) {
					pageEnd = len(s.TableState.Rows)
				}
				pageSize := pageEnd - (s.PageMod * s.MaxRows)
				if s.Cursor < pageSize-1 {
					s.Cursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionPagePrev) {
				if s.PageMod > 0 {
					s.PageMod--
					s.Cursor = 0
				}
				return s, nil
			}
			if km.Matches(key, config.ActionPageNext) {
				if s.PageMod < (len(s.TableState.Rows)-1)/s.MaxRows {
					s.PageMod++
					s.Cursor = 0
				}
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.PanelFocus = TreePanel
				return s, nil
			}
			if km.Matches(key, config.ActionSelect) {
				recordIndex := (s.PageMod * s.MaxRows) + s.Cursor
				if recordIndex < len(s.TableState.Rows) {
					switch s.DatabaseMode {
					case DBModeUpdate:
						return s, ShowDatabaseUpdateDialogCmd(db.DBTable(s.TableState.Table), s.TableState.Rows[recordIndex][0])
					case DBModeDelete:
						// Sync Model cursor to the actual record index before showing dialog,
						// because DIALOGDELETE handler uses m.GetCurrentRowId() which reads m.Cursor.
						return s, tea.Sequence(
							CursorSetCmd(recordIndex),
							ShowDialogCmd("Confirm Delete",
								"Are you sure you want to delete this record? This action cannot be undone.", true, DIALOGDELETE),
						)
					}
					// DBModeRead: enter is no-op (detail already visible in right panel)
				}
				return s, nil
			}

		case RoutePanel:
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.PanelFocus = ContentPanel
				return s, nil
			}
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

	// Data refresh messages
	case TablesSet:
		s.Tables = msg.Tables
		s.Loading = false
		return s, nil
	case HeadersSet:
		s.TableState.Headers = msg.Headers
		s.Loading = false
		return s, nil
	case RowsSet:
		s.TableState.Rows = msg.Rows
		s.Loading = false
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
	if s.ReadMode {
		switch s.PanelFocus {
		case ContentPanel:
			pageHint := km.HintString(config.ActionPagePrev) + "/" + km.HintString(config.ActionPageNext)
			switch s.DatabaseMode {
			case DBModeUpdate:
				return []KeyHint{
					{km.HintString(config.ActionSelect), "edit"},
					{pageHint, "page"},
					{km.HintString(config.ActionNextPanel), "panel"},
					{km.HintString(config.ActionBack), "back"},
					{km.HintString(config.ActionQuit), "quit"},
				}
			case DBModeDelete:
				return []KeyHint{
					{km.HintString(config.ActionSelect), "delete"},
					{pageHint, "page"},
					{km.HintString(config.ActionNextPanel), "panel"},
					{km.HintString(config.ActionBack), "back"},
					{km.HintString(config.ActionQuit), "quit"},
				}
			default:
				return []KeyHint{
					{pageHint, "page"},
					{km.HintString(config.ActionNextPanel), "panel"},
					{km.HintString(config.ActionBack), "back"},
					{km.HintString(config.ActionQuit), "quit"},
				}
			}
		default:
			return []KeyHint{
				{km.HintString(config.ActionSelect), "select"},
				{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
				{km.HintString(config.ActionNextPanel), "panel"},
				{km.HintString(config.ActionBack), "back"},
				{km.HintString(config.ActionQuit), "quit"},
			}
		}
	}
	// DATABASEPAGE
	switch s.PanelFocus {
	case ContentPanel:
		return []KeyHint{
			{km.HintString(config.ActionSelect), "run"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	default:
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
	if s.ReadMode {
		return s.viewReadPage(ctx)
	}
	return s.viewDatabasePage(ctx)
}

// viewDatabasePage renders the DATABASEPAGE 2-panel layout.
func (s *DatabaseScreen) viewDatabasePage(ctx AppContext) string {
	left := s.renderDatabaseTables()
	center := s.renderDatabaseActions()
	right := s.renderDatabaseInfo()

	layout := layoutForPage(DATABASEPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	listLen := len(s.Tables)

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel, TotalLines: listLen, ScrollOffset: ClampScroll(s.Cursor, listLen, innerH)}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[1], Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

// viewReadPage renders the READPAGE 3-panel layout.
func (s *DatabaseScreen) viewReadPage(ctx AppContext) string {
	left := s.renderReadMode()
	center := s.renderReadTable()
	right := s.renderReadRowDetail()

	layout := layoutForPage(READPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	// Dynamic center title: show mode and table name
	modeNames := []string{"Read", "Update", "Delete"}
	mode := "Read"
	if int(s.DatabaseMode) < len(modeNames) {
		mode = modeNames[s.DatabaseMode]
	}
	centerTitle := fmt.Sprintf("%s: %s", mode, s.TableState.Table)

	innerH := PanelInnerHeight(ctx.Height)
	rowCount := 0
	if s.TableState != nil {
		rowCount = len(s.TableState.Rows)
	}

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: centerTitle, Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel, TotalLines: rowCount, ScrollOffset: ClampScroll(s.Cursor, rowCount, innerH)}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

// setDatabaseModeCmd creates a command that sets the database mode on the Model.
func setDatabaseModeCmd(mode DatabaseMode) tea.Cmd {
	return func() tea.Msg {
		return SetDatabaseModeMsg{Mode: mode}
	}
}

// --- DATABASEPAGE render methods ---

// renderDatabaseTables renders the table list for the left panel.
func (s *DatabaseScreen) renderDatabaseTables() string {
	if len(s.Tables) == 0 {
		return "(no tables)"
	}
	lines := make([]string, 0, len(s.Tables))
	for i, tbl := range s.Tables {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, tbl))
	}
	return strings.Join(lines, "\n")
}

// renderDatabaseActions renders the CRUD action menu for the center panel.
func (s *DatabaseScreen) renderDatabaseActions() string {
	if s.TableState.Table == "" {
		return "Select a table"
	}
	actions := []string{"Create", "Read", "Update", "Delete"}
	lines := []string{
		fmt.Sprintf("Table: %s", s.TableState.Table),
		"",
	}
	for i, action := range actions {
		cursor := "   "
		if s.PanelFocus == ContentPanel && s.FieldCursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, action))
	}
	return strings.Join(lines, "\n")
}

// renderDatabaseInfo renders column metadata for the right panel.
func (s *DatabaseScreen) renderDatabaseInfo() string {
	if s.TableState.Table == "" {
		return "Database\n\n  Select a table to\n  view its columns."
	}
	lines := []string{
		fmt.Sprintf("Table: %s", s.TableState.Table),
		"",
	}
	if len(s.TableState.Headers) > 0 {
		lines = append(lines, "Columns:")
		lines = append(lines, "")
		for i, h := range s.TableState.Headers {
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, h))
		}
	} else {
		lines = append(lines, "  (no column info)")
	}
	return strings.Join(lines, "\n")
}

// --- READPAGE render methods ---

// renderReadMode renders the mode selector for the left panel.
func (s *DatabaseScreen) renderReadMode() string {
	modes := []struct {
		label string
		mode  DatabaseMode
	}{
		{"Read", DBModeRead},
		{"Update", DBModeUpdate},
		{"Delete", DBModeDelete},
	}

	lines := make([]string, 0, len(modes)+2)
	for i, mode := range modes {
		cursor := "   "
		if s.PanelFocus == TreePanel && s.FieldCursor == i {
			cursor = " ->"
		}
		active := ""
		if s.DatabaseMode == mode.mode {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, mode.label, active))
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Rows: %d", len(s.TableState.Rows)))
	if len(s.TableState.Rows) > s.MaxRows {
		totalPages := (len(s.TableState.Rows)-1)/s.MaxRows + 1
		lines = append(lines, fmt.Sprintf("  Page: %d/%d", s.PageMod+1, totalPages))
	}

	return strings.Join(lines, "\n")
}

// renderReadTable renders paginated table rows for the center panel.
func (s *DatabaseScreen) renderReadTable() string {
	if s.Loading {
		return "\n   Loading..."
	}
	if len(s.TableState.Headers) == 0 {
		return "No data loaded"
	}

	// Calculate page bounds
	start := s.PageMod * s.MaxRows
	end := start + s.MaxRows
	if end > len(s.TableState.Rows) {
		end = len(s.TableState.Rows)
	}
	if start >= len(s.TableState.Rows) {
		return "No rows on this page"
	}

	currentView := s.TableState.Rows[start:end]

	// Build simple text table with cursor
	lines := make([]string, 0, len(currentView)+2)

	// Header row (truncate each header to fit)
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
		if s.PanelFocus == ContentPanel && s.Cursor == i {
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
		lines = append(lines, fmt.Sprintf("  Page %d of %d", s.PageMod+1, totalPages))
	}

	return strings.Join(lines, "\n")
}

// renderReadRowDetail renders key-value detail of the selected row for the right panel.
func (s *DatabaseScreen) renderReadRowDetail() string {
	if len(s.TableState.Rows) == 0 || len(s.TableState.Headers) == 0 {
		return "No row selected"
	}

	// Calculate actual row index from page offset + cursor
	rowIndex := (s.PageMod * s.MaxRows) + s.Cursor
	if rowIndex >= len(s.TableState.Rows) {
		return "No row selected"
	}

	row := s.TableState.Rows[rowIndex]
	lines := make([]string, 0, len(s.TableState.Headers)+4)
	lines = append(lines, fmt.Sprintf("Row %d", rowIndex))
	lines = append(lines, "")

	for i, header := range s.TableState.Headers {
		value := ""
		if i < len(row) {
			value = row[i]
		}
		lines = append(lines, fmt.Sprintf("%s:", header))
		if len(value) > 40 {
			// Wrap long values
			for len(value) > 40 {
				lines = append(lines, fmt.Sprintf("  %s", value[:40]))
				value = value[40:]
			}
			if len(value) > 0 {
				lines = append(lines, fmt.Sprintf("  %s", value))
			}
		} else {
			lines = append(lines, fmt.Sprintf("  %s", value))
		}
	}

	// Mode-specific hint
	lines = append(lines, "")
	switch s.DatabaseMode {
	case DBModeUpdate:
		lines = append(lines, "  enter: Edit this row")
	case DBModeDelete:
		warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
		lines = append(lines, warnStyle.Render("  enter: Delete this row"))
	}

	return strings.Join(lines, "\n")
}
