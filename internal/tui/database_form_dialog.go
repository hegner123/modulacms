package tui

import (
	"database/sql"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// =============================================================================
// DATABASE FORM DIALOG
// =============================================================================

// DatabaseColumnInput holds column metadata and input state for a database form.
type DatabaseColumnInput struct {
	Column   string          // Column name
	TypeName string          // Database type (TEXT, INTEGER, etc.)
	Nullable bool            // Whether column allows NULL
	Input    textinput.Model // Text input for user entry
	Hidden   bool            // true for auto-filled columns (ID, dates, history)
}

// DatabaseFormDialogModel represents a form dialog for database table INSERT/UPDATE operations.
type DatabaseFormDialogModel struct {
	dialogStyles

	Title      string
	Width      int
	Action     FormDialogAction
	Table      db.DBTable
	RowID      string // For UPDATE: first column value of the selected row
	Fields     []DatabaseColumnInput
	focusIndex int

	// Scroll state for dialogs that exceed terminal height
	scroll ScrollState
}

// autoFillColumns maps column names that should be hidden and auto-filled.
var autoFillColumns = map[string]bool{
	"id":            true,
	"date_created":  true,
	"date_modified": true,
	"history":       true,
}

// isAutoFillColumn returns true if the column should be hidden/auto-filled.
func isAutoFillColumn(name string) bool {
	return autoFillColumns[strings.ToLower(name)]
}

// NewDatabaseInsertDialog creates a dialog for inserting a new row.
func NewDatabaseInsertDialog(title string, table db.DBTable, columns []string, columnTypes []*sql.ColumnType) DatabaseFormDialogModel {
	fields := make([]DatabaseColumnInput, len(columns))
	firstVisible := -1

	for i, col := range columns {
		ti := textinput.New()
		ti.CharLimit = 256
		ti.SetWidth(40)

		typeName := ""
		nullable := false
		if i < len(columnTypes) && columnTypes[i] != nil {
			typeName = columnTypes[i].DatabaseTypeName()
			n, ok := columnTypes[i].Nullable()
			if ok {
				nullable = n
			}
		}

		hidden := isAutoFillColumn(col)
		if hidden {
			ti.Placeholder = "(auto)"
		} else {
			ti.Placeholder = typeName
			if firstVisible == -1 {
				firstVisible = i
				ti.Focus()
			}
		}

		fields[i] = DatabaseColumnInput{
			Column:   col,
			TypeName: typeName,
			Nullable: nullable,
			Input:    ti,
			Hidden:   hidden,
		}
	}

	focusIdx := 0
	if firstVisible >= 0 {
		// Map to visible index
		vis := 0
		for i := range fields {
			if fields[i].Hidden {
				continue
			}
			if i == firstVisible {
				focusIdx = vis
				break
			}
			vis++
		}
	}

	return DatabaseFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       FORMDIALOGDBINSERT,
		Table:        table,
		Fields:       fields,
		focusIndex:   focusIdx,
	}
}

// NewDatabaseUpdateDialog creates a dialog pre-filled with current row values.
func NewDatabaseUpdateDialog(title string, table db.DBTable, columns []string, columnTypes []*sql.ColumnType, currentRow []string) DatabaseFormDialogModel {
	fields := make([]DatabaseColumnInput, len(columns))
	firstVisible := -1

	rowID := ""
	if len(currentRow) > 0 {
		rowID = currentRow[0]
	}

	for i, col := range columns {
		ti := textinput.New()
		ti.CharLimit = 256
		ti.SetWidth(40)

		typeName := ""
		nullable := false
		if i < len(columnTypes) && columnTypes[i] != nil {
			typeName = columnTypes[i].DatabaseTypeName()
			n, ok := columnTypes[i].Nullable()
			if ok {
				nullable = n
			}
		}

		hidden := isAutoFillColumn(col)
		if hidden {
			ti.Placeholder = "(auto)"
		} else {
			ti.Placeholder = typeName
			if firstVisible == -1 {
				firstVisible = i
				ti.Focus()
			}
		}

		// Pre-fill with current value
		if i < len(currentRow) {
			ti.SetValue(currentRow[i])
		}

		fields[i] = DatabaseColumnInput{
			Column:   col,
			TypeName: typeName,
			Nullable: nullable,
			Input:    ti,
			Hidden:   hidden,
		}
	}

	focusIdx := 0
	if firstVisible >= 0 {
		vis := 0
		for i := range fields {
			if fields[i].Hidden {
				continue
			}
			if i == firstVisible {
				focusIdx = vis
				break
			}
			vis++
		}
	}

	return DatabaseFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       FORMDIALOGDBUPDATE,
		Table:        table,
		RowID:        rowID,
		Fields:       fields,
		focusIndex:   focusIdx,
	}
}

// visibleFields returns indices of non-hidden fields.
func (d *DatabaseFormDialogModel) visibleFields() []int {
	var indices []int
	for i, f := range d.Fields {
		if !f.Hidden {
			indices = append(indices, i)
		}
	}
	return indices
}

// totalFocusable returns the number of focusable elements (visible fields + 2 buttons).
func (d *DatabaseFormDialogModel) totalFocusable() int {
	return len(d.visibleFields()) + 2
}

// cancelButtonIndex returns the focus index for the cancel button.
func (d *DatabaseFormDialogModel) cancelButtonIndex() int {
	return len(d.visibleFields())
}

// confirmButtonIndex returns the focus index for the confirm button.
func (d *DatabaseFormDialogModel) confirmButtonIndex() int {
	return len(d.visibleFields()) + 1
}

// Update handles user input for the database form dialog.
func (d *DatabaseFormDialogModel) Update(msg tea.Msg) (DatabaseFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			d.dbFormFocusNext()
			return *d, nil
		case "shift+tab", "up":
			d.dbFormFocusPrev()
			return *d, nil
		case "enter":
			if d.focusIndex == d.confirmButtonIndex() {
				return *d, d.buildAcceptCmd()
			}
			if d.focusIndex == d.cancelButtonIndex() {
				return *d, func() tea.Msg { return DatabaseFormDialogCancelMsg{} }
			}
			// On text fields, move to next
			d.dbFormFocusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return DatabaseFormDialogCancelMsg{} }
		}
	}

	// Update the focused text input
	visible := d.visibleFields()
	if d.focusIndex < len(visible) {
		fieldIdx := visible[d.focusIndex]
		var cmd tea.Cmd
		d.Fields[fieldIdx].Input, cmd = d.Fields[fieldIdx].Input.Update(msg)
		return *d, cmd
	}
	return *d, nil
}

// buildAcceptCmd creates a command returning the accept message with column/value pairs.
func (d *DatabaseFormDialogModel) buildAcceptCmd() tea.Cmd {
	columns := make([]string, 0, len(d.Fields))
	values := make([]string, 0, len(d.Fields))
	for _, f := range d.Fields {
		columns = append(columns, f.Column)
		values = append(values, f.Input.Value())
	}
	return func() tea.Msg {
		return DatabaseFormDialogAcceptMsg{
			Action:  d.Action,
			Table:   d.Table,
			RowID:   d.RowID,
			Columns: columns,
			Values:  values,
		}
	}
}

// dbFormFocusNext advances focus to the next focusable element, wrapping at the end.
func (d *DatabaseFormDialogModel) dbFormFocusNext() {
	total := d.totalFocusable()
	d.focusIndex = (d.focusIndex + 1) % total
	d.dbFormUpdateFocus()
}

// dbFormFocusPrev moves focus to the previous focusable element, wrapping at the start.
func (d *DatabaseFormDialogModel) dbFormFocusPrev() {
	total := d.totalFocusable()
	d.focusIndex = (d.focusIndex + total - 1) % total
	d.dbFormUpdateFocus()
}

// dbFormUpdateFocus applies focus styling to the currently focused input field.
func (d *DatabaseFormDialogModel) dbFormUpdateFocus() {
	visible := d.visibleFields()
	// Blur all
	for i := range d.Fields {
		d.Fields[i].Input.Blur()
	}
	// Focus the active visible field
	if d.focusIndex < len(visible) {
		d.Fields[visible[d.focusIndex]].Input.Focus()
	}
}

// OverlayUpdate implements ModalOverlay for DatabaseFormDialogModel.
func (d *DatabaseFormDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayView implements ModalOverlay for DatabaseFormDialogModel.
func (d *DatabaseFormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the database form dialog with scrolling support.
// Uses a pointer receiver so scrollableBody can persist offset changes.
func (d *DatabaseFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	// --- Header ---
	header := d.titleStyle.Render(d.Title)

	// --- Field items ---
	visible := d.visibleFields()
	var fieldItems []string
	focusItemIdx := 0
	for visIdx, fieldIdx := range visible {
		f := d.Fields[fieldIdx]
		label := d.labelStyle.Render(f.Column)
		style := d.inputStyle
		if d.focusIndex == visIdx {
			style = d.focusedInputStyle
			focusItemIdx = visIdx
		}
		input := style.Width(innerW).Render(f.Input.View())
		fieldItems = append(fieldItems, label+"\n"+input)
	}

	// If focus is on a button, keep the last field visible
	if d.focusIndex >= len(visible) {
		focusItemIdx = len(fieldItems) - 1
		if focusItemIdx < 0 {
			focusItemIdx = 0
		}
	}

	// --- Footer (buttons) ---
	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == d.cancelButtonIndex() {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == d.confirmButtonIndex() {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render(buttonLabel("Cancel", d.focusIndex == d.cancelButtonIndex()))
	confirmBtn := confirmStyle.Render(buttonLabel("Save", d.focusIndex == d.confirmButtonIndex()))
	footer := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)

	// --- Compute available body lines ---
	borderOverhead := 4
	headerH := lipgloss.Height(header) + 1
	footerH := lipgloss.Height(footer) + 1
	indicatorH := 2
	maxDialogH := windowHeight - 4
	maxBodyLines := maxDialogH - borderOverhead - headerH - footerH - indicatorH
	if maxBodyLines < 3 {
		maxBodyLines = 3
	}

	// --- Scroll ---
	visibleBody, topClip, bottomClip := d.scroll.scrollableBody(fieldItems, focusItemIdx, maxBodyLines)

	// --- Assemble ---
	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n")

	if topClip {
		content.WriteString(scrollUpIndicator(innerW))
		content.WriteString("\n")
	}

	content.WriteString(visibleBody)
	content.WriteString("\n")

	if bottomClip {
		content.WriteString(scrollDownIndicator(innerW))
		content.WriteString("\n")
	}

	content.WriteString(footer)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// =============================================================================
// MESSAGES
// =============================================================================

// DatabaseFormDialogAcceptMsg carries acceptance data from a database form dialog.
type DatabaseFormDialogAcceptMsg struct {
	Action  FormDialogAction
	Table   db.DBTable
	RowID   string
	Columns []string
	Values  []string
}

// DatabaseFormDialogCancelMsg is sent when a database form dialog is cancelled.
type DatabaseFormDialogCancelMsg struct{}

// ShowDatabaseFormDialogMsg triggers display of a database form dialog.
type ShowDatabaseFormDialogMsg struct {
	Action FormDialogAction
	Title  string
	Table  db.DBTable
	RowID  string // For UPDATE: selected row ID
}

// =============================================================================
// COMMANDS
// =============================================================================

// ShowDatabaseInsertDialogCmd creates a command to show an insert dialog.
func ShowDatabaseInsertDialogCmd(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return ShowDatabaseFormDialogMsg{
			Action: FORMDIALOGDBINSERT,
			Title:  "Insert Row",
			Table:  table,
		}
	}
}

// ShowDatabaseUpdateDialogCmd creates a command to show an update dialog.
func ShowDatabaseUpdateDialogCmd(table db.DBTable, rowID string) tea.Cmd {
	return func() tea.Msg {
		return ShowDatabaseFormDialogMsg{
			Action: FORMDIALOGDBUPDATE,
			Title:  "Update Row",
			Table:  table,
			RowID:  rowID,
		}
	}
}
