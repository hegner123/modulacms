//go:build debug

package tui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Stringify returns a formatted debug string representation of the model.
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

	//PageRouteId types.RouteID
	pri := fmt.Sprintf("PageRouteId: %s", m.PageRouteId)
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

	//Page         Page
	pg := fmt.Sprintf("Page: %v", m.Page.DebugString())
	out = append(out, pg)

	//Paginator    paginator.Model
	pag := fmt.Sprintf("Paginator: %v", m.Paginator)
	out = append(out, pag)

	//PageMod      int
	pm := fmt.Sprintf("PageMode: %d", m.PageMod)
	out = append(out, pm)

	//Table        string
	table := fmt.Sprintf("Table: %s", m.TableState.Table)
	out = append(out, table)

	//DatatypeMenu []string
	datatypeMenu := fmt.Sprintf("DatatypeMenu(length): %d", len(m.DatatypeMenu))
	out = append(out, datatypeMenu)

	//Tables       []string
	tables := fmt.Sprintf("Tables(length): %d", len(m.Tables))
	out = append(out, tables)

	//Columns      *[]string
	var columnsStr string
	if m.TableState.Columns != nil {
		columnsStr = fmt.Sprintf("length: %d", len(*m.TableState.Columns))
	} else {
		columnsStr = "(nil)"
	}
	columns := fmt.Sprintf("Columns: %s", columnsStr)
	out = append(out, columns)

	//ColumnTypes  *[]*sql.ColumnType
	var columnTypesStr string
	if m.TableState.ColumnTypes != nil {
		columnTypesStr = fmt.Sprintf("length: %d", len(*m.TableState.ColumnTypes))
	} else {
		columnTypesStr = "(nil)"
	}
	columnTypes := fmt.Sprintf("ColumnTypes: %s", columnTypesStr)
	out = append(out, columnTypes)

	//Selected     map[int]struct{}
	selected := fmt.Sprintf("Selected(length): %d", len(m.TableState.Selected))
	out = append(out, selected)
	//Headers      []string
	headers := fmt.Sprintf("Headers(length): %d", len(m.TableState.Headers))
	out = append(out, headers)
	//Rows         [][]string
	rows := fmt.Sprintf("Rows(length): %d", len(m.TableState.Rows))
	out = append(out, rows)
	//Row          *[]string
	var rowStr string
	if m.TableState.Row != nil {
		rowStr = fmt.Sprintf("length: %d", len(*m.TableState.Row))
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

	//ActiveOverlay
	dialogActive := fmt.Sprintf("ActiveOverlay: %v", m.ActiveOverlay != nil)
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

	//ActiveOverlay
	var dialogStr string
	if m.ActiveOverlay != nil {
		dialogStr = fmt.Sprintf("%T", m.ActiveOverlay)
	} else {
		dialogStr = "(nil)"
	}
	dialog := fmt.Sprintf("ActiveOverlay: %s", dialogStr)
	out = append(out, dialog)

	return lipgloss.JoinVertical(lipgloss.Top, out...)
}

// ValidString returns the string if non-empty, otherwise returns "(empty)".
func ValidString(s string) string {
	if len(s) < 1 {
		return "(empty)"
	}
	return s
}

// DebugString returns a formatted debug string representation of the Page.
func (p Page) DebugString() string {
	out := make([]string, 0)

	index := fmt.Sprintf("Index: %d", p.Index)
	out = append(out, index)

	label := fmt.Sprintf("Label: %s", ValidString(p.Label))
	out = append(out, label)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

// DebugStringPtr returns a formatted debug string of a Page pointer, or "(nil)" if nil.
func (p *Page) DebugStringPtr() string {
	if p == nil {
		return "(nil)"
	}
	return p.DebugString()
}

// HuhGroupDebugString returns a formatted debug string representation of a huh.Group.
func HuhGroupDebugString(g huh.Group) string {
	out := make([]string, 0)

	group := fmt.Sprintf("Group: %s", ValidString(g.View()))
	out = append(out, group)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

// HuhFieldDebugString returns a formatted debug string representation of a huh.Field.
func HuhFieldDebugString(f huh.Field) string {
	out := make([]string, 0)

	key := fmt.Sprintf("Key: %s", ValidString(f.GetKey()))
	out = append(out, key)

	value := fmt.Sprintf("Value: %v", f.GetValue())
	out = append(out, value)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

// PageHistoryDebugString returns a formatted debug string representation of PageHistory.
func PageHistoryDebugString(ph PageHistory) string {
	out := make([]string, 0)

	page := fmt.Sprintf("Page: %s", ph.Page.DebugString())
	out = append(out, page)

	cursor := fmt.Sprintf("Cursor: %d", ph.Cursor)
	out = append(out, cursor)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

// SqlRowDebugString returns a formatted debug string representation of a sql.Row.
func SqlRowDebugString(row sql.Row) string {
	out := make([]string, 0)

	err := fmt.Sprintf("Err: %v", row.Err())
	out = append(out, err)

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

// HuhGroupSliceDebugString returns a formatted debug string representation of a huh.Group slice.
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

// HuhFieldSliceDebugString returns a formatted debug string representation of a huh.Field slice.
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

// PageHistorySliceDebugString returns a formatted debug string representation of a PageHistory slice.
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

// SqlRowSliceDebugString returns a formatted debug string representation of a sql.Row slice.
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
