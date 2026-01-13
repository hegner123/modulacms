package cli

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Stringify() string {
	out := make([]string, 0)

	//Config       *config.Config
	dRefConfig := *m.Config
	c := dRefConfig.Stringify()
	out = append(out, c)

	//Status       ApplicationState
	status := fmt.Sprintf("ApplicationState: %d", m.Status)
	out = append(out, status)

	//TitleFont    int
	tf := fmt.Sprintf("TitleFont: %d", m.TitleFont)
	out = append(out, tf)

	//Titles []string
	tl := fmt.Sprintf("Titles(length): %d", len(m.Titles))
	out = append(out, tl)

	//Term         string
	t := ValidString(m.Term)
	term := fmt.Sprintf("Term: %s", t)
	out = append(out, term)

	//Profile      string
	p := ValidString(m.Profile)
	prof := fmt.Sprintf("Profile: %s", p)
	out = append(out, prof)

	//Width       int
	width := fmt.Sprintf("Width: %d", m.Width)
	out = append(out, width)

	//Height       int
	height := fmt.Sprintf("Height: %d", m.Height)
	out = append(out, height)

	//Bg           string
	b := ValidString(m.Bg)
	bg := fmt.Sprintf("Bg: %s", b)
	out = append(out, bg)

	//PageRouteId int64
	pri := fmt.Sprintf("PageRouteId: %d", m.PageRouteId)
	out = append(out, pri)

	//TxtStyle     lipgloss.Style
	txt := fmt.Sprintf("TxtStyle: %v", m.TxtStyle)
	out = append(out, txt)

	//QuitStyle    lipgloss.Style
	qts := fmt.Sprintf("QuitStyle: %v", m.QuitStyle)
	out = append(out, qts)

	//Loading      bool
	ld := fmt.Sprintf("Loading: %v", m.Loading)
	out = append(out, ld)

	//Cursor       int
	crs := fmt.Sprintf("Cursor: %d", m.Cursor)
	out = append(out, crs)

	//CursorMax    int
	crsm := fmt.Sprintf("CursorMax: %d", m.CursorMax)
	out = append(out, crsm)

	//FocusIndex   int
	fi := fmt.Sprintf("FocusIndex: %d", m.FocusIndex)
	out = append(out, fi)

	//Page         Page
	pg := fmt.Sprintf("Page: %v", m.Page.DebugString())
	out = append(out, pg)

	//Paginator    paginator.Model
	pag := fmt.Sprintf("Paginator: %v", m.Paginator)
	out = append(out, pag)

	//PageMod      int
	pm := fmt.Sprintf("PageMode: %d", m.PageMod)
	out = append(out, pm)

	//MaxRows      int
	mr := fmt.Sprintf("MaxRows: %d", m.MaxRows)
	out = append(out, mr)

	//Table        string
	table := fmt.Sprintf("Table: %s", m.Table)
	out = append(out, table)

	/*
		//PageMenu     []*Page
		pageMenuDebug := fmt.Sprint("PageMenu:\n")
		out = append(out, pageMenuDebug)

		//Pages        []Page
		pagesDebug := fmt.Sprintf("Page Slices\n %s", PageValueSliceDebugString(m.Pages))
		out = append(out, pagesDebug)
	*/

	//DatatypeMenu []string
	datatypeMenu := fmt.Sprintf("DatatypeMenu(length): %d", len(m.DatatypeMenu))
	out = append(out, datatypeMenu)

	//Tables       []string
	tables := fmt.Sprintf("Tables(length): %d", len(m.Tables))
	out = append(out, tables)

	//Columns      *[]string
	var columnsStr string
	if m.Columns != nil {
		columnsStr = fmt.Sprintf("length: %d", len(*m.Columns))
	} else {
		columnsStr = "(nil)"
	}
	columns := fmt.Sprintf("Columns: %s", columnsStr)
	out = append(out, columns)

	//ColumnTypes  *[]*sql.ColumnType
	var columnTypesStr string
	if m.ColumnTypes != nil {
		columnTypesStr = fmt.Sprintf("length: %d", len(*m.ColumnTypes))
	} else {
		columnTypesStr = "(nil)"
	}
	columnTypes := fmt.Sprintf("ColumnTypes: %s", columnTypesStr)
	out = append(out, columnTypes)

	//Selected     map[int]struct{}
	selected := fmt.Sprintf("Selected(length): %d", len(m.Selected))
	out = append(out, selected)
	//Headers      []string
	headers := fmt.Sprintf("Headers(length): %d", len(m.Headers))
	out = append(out, headers)
	//Rows         [][]string
	rows := fmt.Sprintf("Rows(length): %d", len(m.Rows))
	out = append(out, rows)
	//Row          *[]string
	var rowStr string
	if m.Row != nil {
		rowStr = fmt.Sprintf("length: %d", len(*m.Row))
	} else {
		rowStr = "(nil)"
	}
	row := fmt.Sprintf("Row: %s", rowStr)
	out = append(out, row)
	//Form         *huh.Form
	var formStr string
	if m.FormState.Form != nil {
		formStr = "initialized"
	} else {
		formStr = "(nil)"
	}
	form := fmt.Sprintf("Form: %s", formStr)
	out = append(out, form)

	//FormLen      int
	formLen := fmt.Sprintf("FormLen: %d", m.FormState.FormLen)
	out = append(out, formLen)

	//FormMap      []string
	formMap := fmt.Sprintf("FormMap(length): %d", len(m.FormState.FormMap))
	out = append(out, formMap)

	//FormValues   []*string
	formValues := fmt.Sprintf("FormValues(length): %d", len(m.FormState.FormValues))
	out = append(out, formValues)

	//FormSubmit   bool
	formSubmit := fmt.Sprintf("FormSubmit: %v", m.FormState.FormSubmit)
	out = append(out, formSubmit)

	//FormGroups   []huh.Group
	formGroupsDebug := HuhGroupSliceDebugString(m.FormState.FormGroups)
	out = append(out, formGroupsDebug)

	//FormFields   []huh.Field
	formFieldsDebug := HuhFieldSliceDebugString(m.FormState.FormFields)
	out = append(out, formFieldsDebug)


	//Verbose      bool
	verb := fmt.Sprintf("Verbose: %v", m.Verbose)
	out = append(out, verb)

	//Content      string
	content := fmt.Sprintf("Content: %s", ValidString(m.Content))
	out = append(out, content)

	//Ready        bool
	ready := fmt.Sprintf("Ready: %v", m.Ready)
	out = append(out, ready)

	//Err          error
	var errStr string
	if m.Err != nil {
		errStr = m.Err.Error()
	} else {
		errStr = "(nil)"
	}
	err := fmt.Sprintf("Err: %s", errStr)
	out = append(out, err)

	//Time         time.Time
	timeStr := fmt.Sprintf("Time: %s", m.Time.Format("2006-01-02 15:04:05"))
	out = append(out, timeStr)

	//DialogActive bool
	dialogActive := fmt.Sprintf("DialogActive: %v", m.DialogActive)
	out = append(out, dialogActive)

	//Focus        FocusKey
	focus := fmt.Sprintf("Focus: %v", m.Focus)
	out = append(out, focus)

	//Spinner      spinner.Model
	spinner := fmt.Sprintf("Spinner: %v", m.Spinner)
	out = append(out, spinner)

	//Viewport     viewport.Model
	viewport := fmt.Sprintf("Viewport: %v", m.Viewport)
	out = append(out, viewport)

	//History      []PageHistory
	historyDebug := PageHistorySliceDebugString(m.History)
	hst := fmt.Sprintf("History:%s", historyDebug)
	out = append(out, hst)

	//QueryResults []sql.Row
	queryResultsDebug := SqlRowSliceDebugString(m.QueryResults)
	qrd := fmt.Sprintf("QueryResults:%s", queryResultsDebug)
	out = append(out, qrd)

	//Dialog       *DialogModel
	var dialogStr string
	if m.Dialog != nil {
		dialogStr = "initialized"
	} else {
		dialogStr = "(nil)"
	}
	dialog := fmt.Sprintf("Dialog: %s", dialogStr)
	out = append(out, dialog)

	//Root         TreeRoot
	root := fmt.Sprintf("Root: %v", m.Root)
	out = append(out, root)

	return lipgloss.JoinVertical(lipgloss.Top, out...)
}

func ValidString(s string) string {
	if len(s) < 1 {
		return "(empty)"
	} else {
		return s
	}
}

func (p Page) DebugString() string {
	out := make([]string, 0)

	index := fmt.Sprintf("Index: %d", p.Index)
	out = append(out, index)

	label := fmt.Sprintf("Label: %s", ValidString(p.Label))
	out = append(out, label)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

func (p *Page) DebugStringPtr() string {
	if p == nil {
		return "(nil)"
	}
	return p.DebugString()
}

func HuhGroupDebugString(g huh.Group) string {
	out := make([]string, 0)

	group := fmt.Sprintf("Group: %s", ValidString(g.View()))
	out = append(out, group)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

func HuhFieldDebugString(f huh.Field) string {
	out := make([]string, 0)

	key := fmt.Sprintf("Key: %s", ValidString(f.GetKey()))
	out = append(out, key)

	value := fmt.Sprintf("Value: %v", f.GetValue())
	out = append(out, value)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

func PageHistoryDebugString(ph PageHistory) string {
	out := make([]string, 0)

	page := fmt.Sprintf("Page: %s", ph.Page.DebugString())
	out = append(out, page)

	cursor := fmt.Sprintf("Cursor: %d", ph.Cursor)
	out = append(out, cursor)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

func SqlRowDebugString(row sql.Row) string {
	out := make([]string, 0)

	err := fmt.Sprintf("Err: %v", row.Err())
	out = append(out, err)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}
/*
func DbContentFieldsDebugString(cf db.ContentFields) string {
	out := make([]string, 0)

	id := fmt.Sprintf("ID: %d", cf.ContentFieldID)
	out = append(out, id)

	contentID := fmt.Sprintf("ContentID: %d", cf.ContentDataID)
	out = append(out, contentID)

	fieldID := fmt.Sprintf("FieldID: %d", cf.FieldID)
	out = append(out, fieldID)

	value := fmt.Sprintf("Value: %s", ValidString(cf.FieldValue))
	out = append(out, value)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

func DbFieldsDebugString(f db.Fields) string {
	out := make([]string, 0)

	id := fmt.Sprintf("ID: %d", f.FieldID)
	out = append(out, id)

	ParentID := fmt.Sprintf("ParentID: %d", f.ParentID.Int64)
	out = append(out, ParentID)

	name := fmt.Sprintf("Name: %s", ValidString(f.Label))
	out = append(out, name)

	fieldType := fmt.Sprintf("Type: %s", ValidString(f.Type))
	out = append(out, fieldType)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}*/

/*func TreeNodeDebugString(tn TreeNode) string {
	out := make([]string, 0)

	var nodeStr string
	if tn.Node != nil {
		nodeStr = fmt.Sprintf("ID: %d", tn.Node.ContentDataID)
	} else {
		nodeStr = "(nil)"
	}
	node := fmt.Sprintf("Node: %s", nodeStr)
	out = append(out, node)

	nodeFields := fmt.Sprintf("NodeFields(length): %d", len(tn.NodeFields))
	out = append(out, nodeFields)

	datatype := fmt.Sprintf("NodeDatatype: %s", ValidString(tn.NodeDatatype.Label))
	out = append(out, datatype)

	nodeFieldTypes := fmt.Sprintf("NodeFieldTypes(length): %d", len(tn.NodeFieldTypes))
	out = append(out, nodeFieldTypes)

	var nodesStr string
	if tn.Nodes != nil {
		nodesStr = fmt.Sprintf("%d", len(*tn.Nodes))
	} else {
		nodesStr = "(nil)"
	}
	nodes := fmt.Sprintf("Nodes(length): %s", nodesStr)
	out = append(out, nodes)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}*/

/*func DbDatatypesDebugString(dt db.Datatypes) string {
	out := make([]string, 0)

	datatypeID := fmt.Sprintf("DatatypeID: %d", dt.DatatypeID)
	out = append(out, datatypeID)

	var parentIDStr string
	if dt.ParentID.Valid {
		parentIDStr = fmt.Sprintf("%d", dt.ParentID.Int64)
	} else {
		parentIDStr = "(null)"
	}
	parentID := fmt.Sprintf("ParentID: %s", parentIDStr)
	out = append(out, parentID)

	label := fmt.Sprintf("Label: %s", ValidString(dt.Label))
	out = append(out, label)

	typeStr := fmt.Sprintf("Type: %s", ValidString(dt.Type))
	out = append(out, typeStr)

	authorID := fmt.Sprintf("AuthorID: %d", dt.AuthorID)
	out = append(out, authorID)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}*/

/*func DbContentDataDebugString(cd db.ContentData) string {
	out := make([]string, 0)

	contentDataID := fmt.Sprintf("ContentDataID: %d", cd.ContentDataID)
	out = append(out, contentDataID)

	var parentIDStr string
	if cd.ParentID.Valid {
		parentIDStr = fmt.Sprintf("%d", cd.ParentID.Int64)
	} else {
		parentIDStr = "(null)"
	}
	parentID := fmt.Sprintf("ParentID: %s", parentIDStr)
	out = append(out, parentID)

	routeID := fmt.Sprintf("RouteID: %d", cd.RouteID)
	out = append(out, routeID)

	datatypeID := fmt.Sprintf("DatatypeID: %d", cd.DatatypeID)
	out = append(out, datatypeID)

	authorID := fmt.Sprintf("AuthorID: %d", cd.AuthorID)
	out = append(out, authorID)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}*/

/*func PageSliceDebugString(pages []*Page) string {
	if len(pages) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, page := range pages {
		header := fmt.Sprintf("--- Page %d ---", i)
		var content string
		if page != nil {
			content = page.DebugString()
		} else {
			content = "(nil)"
		}
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}
	// join horizontal by three
	//join verticle by end

	return lipgloss.JoinHorizontal(lipgloss.Top)
}*/

/*func PageValueSliceDebugString(pages []Page) string {
	if len(pages) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, page := range pages {
		header := fmt.Sprintf("--- Page %d ---", i)
		content := page.DebugString()
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}*/

func HuhGroupSliceDebugString(groups []huh.Group) string {
	if len(groups) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, group := range groups {
		header := fmt.Sprintf("--- Group %d ---", i)
		content := HuhGroupDebugString(group)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func HuhFieldSliceDebugString(fields []huh.Field) string {
	if len(fields) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, field := range fields {
		header := fmt.Sprintf("--- Field %d ---", i)
		content := HuhFieldDebugString(field)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func PageHistorySliceDebugString(history []PageHistory) string {
	if len(history) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, entry := range history {
		header := fmt.Sprintf("--- History %d ---", i)
		content := PageHistoryDebugString(entry)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func SqlRowSliceDebugString(rows []sql.Row) string {
	if len(rows) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, row := range rows {
		header := fmt.Sprintf("--- Row %d ---", i)
		content := SqlRowDebugString(row)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

/*func DbContentFieldsSliceDebugString(fields []db.ContentFields) string {
	if len(fields) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, field := range fields {
		header := fmt.Sprintf("--- ContentField %d ---", i)
		content := DbContentFieldsDebugString(field)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}*/

/*func DbFieldsSliceDebugString(fields []db.Fields) string {
	if len(fields) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, field := range fields {
		header := fmt.Sprintf("--- Field %d ---", i)
		content := DbFieldsDebugString(field)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}*/

/*func TreeNodeSliceDebugString(nodes []*TreeNode) string {
	if len(nodes) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, node := range nodes {
		header := fmt.Sprintf("--- TreeNode %d ---", i)
		var content string
		if node != nil {
			content = TreeNodeDebugString(*node)
		} else {
			content = "(nil)"
		}
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}*/

/*func DbDatatypesSliceDebugString(datatypes []db.Datatypes) string {
	if len(datatypes) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, datatype := range datatypes {
		header := fmt.Sprintf("--- Datatype %d ---", i)
		content := DbDatatypesDebugString(datatype)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}*/

/*func DbContentDataSliceDebugString(contentData []db.ContentData) string {
	if len(contentData) == 0 {
		return "(empty slice)"
	}

	columns := make([]string, 0)
	for i, data := range contentData {
		header := fmt.Sprintf("--- ContentData %d ---", i)
		content := DbContentDataDebugString(data)
		column := lipgloss.JoinVertical(lipgloss.Top, header, content)
		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}*/

/*func DebugViewPortString(v viewport.Model) string {
	src := make([]string, 0)
	src = append(src, fmt.Sprintf("Height: %d", v.Height))
	src = append(src, fmt.Sprintf("Width: %d", v.Width))
	src = append(src, fmt.Sprintf("Mouse Wheel Enabled: %v", v.MouseWheelEnabled))
	src = append(src, fmt.Sprintf("Mouse Wheel Delta: %d", v.MouseWheelDelta))

	return lipgloss.JoinVertical(lipgloss.Top, src...)
}*/
