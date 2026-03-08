package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/validation"
)

// =============================================================================
// ContentFormDialogModel - Dynamic content fields dialog
// =============================================================================

// ContentFieldInput represents a single field input in the content form.
type ContentFieldInput struct {
	FieldID        types.FieldID
	Label          string
	Type           string // field type (text, textarea, number, etc.)
	Widget         string // UI widget override from UIConfig (e.g. "markdown", "code-editor")
	Bubble         FieldBubble
	ValidationJSON string // raw JSON from fields.validation
	DataJSON       string // raw JSON from fields.data
	HelpText       string // help text from UIConfig, shown below label
	Hidden         bool   // hidden from UIConfig, skip in render and focus
}

// ContentFormDialogModel represents a form dialog with dynamic content fields.
type ContentFormDialogModel struct {
	dialogStyles

	Title      string
	Width      int
	Action     FormDialogAction
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	ContentID  types.ContentID         // For edit mode, empty for create
	ParentID   types.NullableContentID // Parent content for child creation

	// Dynamic field inputs
	Fields []ContentFieldInput

	// Validation errors from pre-submit validation (nil when no validation has run or all fields are valid)
	ValidationErrors *validation.ValidationErrors

	// Logger for editor and dialog operations (nil-safe; callers should set after construction)
	Logger Logger

	// Focus management: 0 to len(Fields)-1 for fields, then Cancel, then Confirm
	focusIndex int

	// Scroll state for dialogs that exceed terminal height
	scroll ScrollState
}

// NewContentFormDialog creates a new content form dialog with dynamic fields
func NewContentFormDialog(title string, action FormDialogAction, datatypeID types.DatatypeID, routeID types.RouteID, fields []db.Fields) ContentFormDialogModel {
	contentFields := make([]ContentFieldInput, 0, len(fields))
	for _, f := range fields {
		contentFields = append(contentFields, resolveFieldInput(f))
	}
	// Focus first visible field after all inputs are created
	firstVisible := 0
	for i := range contentFields {
		if !contentFields[i].Hidden {
			contentFields[i].Bubble.Focus()
			firstVisible = i
			break
		}
	}

	return ContentFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        65,
		Action:       action,
		DatatypeID:   datatypeID,
		RouteID:      routeID,
		Fields:       contentFields,
		focusIndex:   firstVisible,
	}
}

// NewContentFormDialogWithParent creates a content form for child content creation
func NewContentFormDialogWithParent(title string, action FormDialogAction, datatypeID types.DatatypeID, routeID types.RouteID, parentID types.ContentID, fields []db.Fields) ContentFormDialogModel {
	dialog := NewContentFormDialog(title, action, datatypeID, routeID, fields)
	dialog.ParentID = types.NullableContentID{ID: parentID, Valid: true}
	return dialog
}

// ButtonCancelIndex returns the index of the Cancel button
func (d *ContentFormDialogModel) ButtonCancelIndex() int {
	return len(d.Fields)
}

// ButtonConfirmIndex returns the index of the Confirm button
func (d *ContentFormDialogModel) ButtonConfirmIndex() int {
	return len(d.Fields) + 1
}

// Update handles input for the content form dialog
func (d *ContentFormDialogModel) Update(msg tea.Msg) (ContentFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			d.focusNext()
			return *d, nil
		case "shift+tab", "up":
			d.focusPrev()
			return *d, nil
		case "left":
			if d.focusIndex == d.ButtonConfirmIndex() {
				d.focusIndex = d.ButtonCancelIndex()
				return *d, nil
			}
		case "right":
			if d.focusIndex == d.ButtonCancelIndex() {
				d.focusIndex = d.ButtonConfirmIndex()
				return *d, nil
			}
		case "enter":
			if d.focusIndex == d.ButtonCancelIndex() {
				return *d, func() tea.Msg { return ContentFormDialogCancelMsg{} }
			}
			if d.focusIndex == d.ButtonConfirmIndex() {
				// Pre-submit validation: build inputs and validate
				inputs := make([]validation.FieldInput, 0, len(d.Fields))
				for _, f := range d.Fields {
					inputs = append(inputs, validation.FieldInput{
						FieldID:    f.FieldID,
						Label:      f.Label,
						FieldType:  types.FieldType(f.Type),
						Value:      f.Bubble.Value(),
						Validation: f.ValidationJSON,
						Data:       f.DataJSON,
					})
				}
				ve := validation.ValidateBatch(inputs)
				if ve.HasErrors() {
					d.ValidationErrors = &ve
					return *d, nil
				}
				d.ValidationErrors = nil

				// Collect all field values
				fieldValues := make(map[types.FieldID]string)
				for _, f := range d.Fields {
					fieldValues[f.FieldID] = f.Bubble.Value()
				}
				return *d, func() tea.Msg {
					return ContentFormDialogAcceptMsg{
						Action:      d.Action,
						DatatypeID:  d.DatatypeID,
						RouteID:     d.RouteID,
						ContentID:   d.ContentID,
						ParentID:    d.ParentID,
						FieldValues: fieldValues,
					}
				}
			}
			// On text fields, enter moves to next field
			d.focusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return ContentFormDialogCancelMsg{} }
		case "ctrl+e":
			// Launch $EDITOR for fields with editor-capable widgets
			if d.focusIndex < len(d.Fields) {
				f := d.Fields[d.focusIndex]
				if d.Logger != nil {
					d.Logger.Finfo(fmt.Sprintf("[editor] ctrl+e pressed on field %d (%s), widget=%q, type=%q", d.focusIndex, f.Label, f.Widget, f.Type))
				}
				if isEditorWidget(f.Widget) {
					fieldIdx := d.focusIndex
					currentValue := f.Bubble.Value()
					if d.Logger != nil {
						d.Logger.Finfo(fmt.Sprintf("[editor] widget %q is editor-capable, preparing editor cmd for field %d, value length=%d", f.Widget, fieldIdx, len(currentValue)))
					}
					cmd := prepareEditorCmd(fieldIdx, currentValue, f.Widget, d.Logger)
					if cmd != nil {
						if d.Logger != nil {
							d.Logger.Finfo(fmt.Sprintf("[editor] editor cmd prepared successfully for field %d", fieldIdx))
						}
						return *d, cmd
					}
					if d.Logger != nil {
						d.Logger.Finfo(fmt.Sprintf("[editor] prepareEditorCmd returned nil for field %d — temp file creation likely failed", fieldIdx))
					}
				} else {
					if d.Logger != nil {
						d.Logger.Finfo(fmt.Sprintf("[editor] widget %q is not editor-capable, ignoring ctrl+e", f.Widget))
					}
				}
			} else {
				if d.Logger != nil {
					d.Logger.Finfo(fmt.Sprintf("[editor] ctrl+e pressed but focusIndex %d is out of field range (%d fields)", d.focusIndex, len(d.Fields)))
				}
			}
		}

		// Update the focused field bubble
		if d.focusIndex < len(d.Fields) {
			var cmd tea.Cmd
			d.Fields[d.focusIndex].Bubble, cmd = d.Fields[d.focusIndex].Bubble.Update(msg)
			// Clear validation errors for this field when the user edits it
			if d.ValidationErrors != nil {
				d.ValidationErrors.ClearField(d.Fields[d.focusIndex].FieldID)
			}
			return *d, cmd
		}
	}

	return *d, nil
}

// focusNext advances focus to the next focusable element in the content form,
// skipping hidden fields while keeping indices stable.
func (d *ContentFormDialogModel) focusNext() {
	total := d.ButtonConfirmIndex() + 1
	for range total {
		d.focusIndex++
		if d.focusIndex > d.ButtonConfirmIndex() {
			d.focusIndex = 0
		}
		if d.focusIndex >= len(d.Fields) || !d.Fields[d.focusIndex].Hidden {
			break
		}
	}
	d.updateFocus()
}

// focusPrev moves focus to the previous focusable element in the content form,
// skipping hidden fields while keeping indices stable.
func (d *ContentFormDialogModel) focusPrev() {
	total := d.ButtonConfirmIndex() + 1
	for range total {
		d.focusIndex--
		if d.focusIndex < 0 {
			d.focusIndex = d.ButtonConfirmIndex()
		}
		if d.focusIndex >= len(d.Fields) || !d.Fields[d.focusIndex].Hidden {
			break
		}
	}
	d.updateFocus()
}

// updateFocus applies focus styling to the currently focused field in the content form.
func (d *ContentFormDialogModel) updateFocus() {
	// Blur all fields
	for i := range d.Fields {
		d.Fields[i].Bubble.Blur()
	}
	// Focus the current field
	if d.focusIndex < len(d.Fields) {
		d.Fields[d.focusIndex].Bubble.Focus()
	}
}

// OverlayUpdate implements ModalOverlay for ContentFormDialogModel.
func (d *ContentFormDialogModel) OverlayUpdate(msg tea.KeyMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayView implements ModalOverlay for ContentFormDialogModel.
func (d *ContentFormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the content form dialog with scrolling support.
// Uses a pointer receiver so scrollableBody can persist offset changes.
func (d *ContentFormDialogModel) Render(windowWidth, windowHeight int) string {
	// Inner width available for fields (dialog width minus border and padding)
	innerW := d.Width - dialogBorderPadding

	// --- Header ---
	header := d.titleStyle.Render(d.Title)

	// --- Field items (one string per visible field) ---
	editorHintStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Italic(true)
	helpTextStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Italic(true)
	validationErrStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent2).
		Italic(true)

	var fieldItems []string
	focusItemIdx := 0
	visIdx := 0
	for i, f := range d.Fields {
		if f.Hidden {
			continue
		}
		f.Bubble.SetWidth(innerW)

		var item strings.Builder
		label := d.labelStyle.Render(f.Label)
		if isEditorWidget(f.Widget) {
			label += " " + editorHintStyle.Render("ctrl+e: $EDITOR")
		}
		item.WriteString(label)
		item.WriteString("\n")

		if f.HelpText != "" {
			item.WriteString(helpTextStyle.Render("  " + f.HelpText))
			item.WriteString("\n")
		}

		if d.focusIndex == i {
			item.WriteString(d.focusedInputStyle.Width(innerW).Render(f.Bubble.View()))
			focusItemIdx = visIdx
		} else {
			item.WriteString(d.inputStyle.Width(innerW).Render(f.Bubble.View()))
		}

		// Validation errors below the field input
		if d.ValidationErrors != nil {
			if fe := d.ValidationErrors.ForField(f.FieldID); fe != nil {
				for _, errMsg := range fe.Messages {
					item.WriteString("\n")
					item.WriteString(validationErrStyle.Render("  " + errMsg))
				}
			}
		}

		fieldItems = append(fieldItems, item.String())
		visIdx++
	}

	// If focus is on a button, keep the last field visible
	if d.focusIndex >= len(d.Fields) {
		focusItemIdx = len(fieldItems) - 1
		if focusItemIdx < 0 {
			focusItemIdx = 0
		}
	}

	// --- Footer (buttons) ---
	cancelBtn := d.renderCancelButton()
	confirmBtn := d.renderConfirmButton()
	footer := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, confirmBtn)

	// --- Compute available body lines ---
	// Border + padding overhead: top border (1) + top padding (1) + bottom padding (1) + bottom border (1) = 4
	// Plus header gap (\n\n = 2 lines after header) and footer gap (\n before buttons)
	borderOverhead := 4
	headerH := lipgloss.Height(header) + 1 // +1 for blank line after header
	footerH := lipgloss.Height(footer) + 1 // +1 for blank line before footer
	indicatorH := 2                        // room for both scroll indicators (1 line each)
	maxDialogH := windowHeight - 4         // leave margin around dialog
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

	// Apply border
	return d.borderStyle.Width(d.Width).Render(content.String())
}

// renderCancelButton returns the styled cancel button view for the content form.
func (d ContentFormDialogModel) renderCancelButton() string {
	focused := d.focusIndex == d.ButtonCancelIndex()
	style := d.cancelButtonStyle
	if focused {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Tertiary).
			Bold(true)
	}
	return style.Render(buttonLabel("Cancel", focused))
}

// renderConfirmButton returns the styled confirm button view for the content form.
func (d ContentFormDialogModel) renderConfirmButton() string {
	focused := d.focusIndex == d.ButtonConfirmIndex()
	style := d.confirmButtonStyle
	if focused {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Accent).
			Bold(true)
	}
	return style.Render(buttonLabel("Confirm", focused))
}

// ContentFormDialogAcceptMsg carries content form dialog acceptance data.
type ContentFormDialogAcceptMsg struct {
	Action      FormDialogAction
	DatatypeID  types.DatatypeID
	RouteID     types.RouteID
	ContentID   types.ContentID         // For edit mode
	ParentID    types.NullableContentID // For child creation
	FieldValues map[types.FieldID]string
}

// ContentFormDialogCancelMsg is sent when a content form dialog is cancelled.
type ContentFormDialogCancelMsg struct{}

// ShowContentFormDialogMsg triggers display of a content form dialog.
type ShowContentFormDialogMsg struct {
	Action     FormDialogAction
	Title      string
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	ParentID   types.NullableContentID
	Fields     []db.Fields
}

// ShowContentFormDialogCmd creates a command to display a content form dialog.
func ShowContentFormDialogCmd(action FormDialogAction, title string, datatypeID types.DatatypeID, routeID types.RouteID, parentID types.NullableContentID, fields []db.Fields) tea.Cmd {
	return func() tea.Msg {
		return ShowContentFormDialogMsg{
			Action:     action,
			Title:      title,
			DatatypeID: datatypeID,
			RouteID:    routeID,
			ParentID:   parentID,
			Fields:     fields,
		}
	}
}

// CreateContentFromDialogRequestMsg triggers content creation from the dialog.
type CreateContentFromDialogRequestMsg struct {
	DatatypeID  types.DatatypeID
	RouteID     types.RouteID
	ParentID    types.NullableContentID
	FieldValues map[types.FieldID]string
}

// CreateContentFromDialogCmd creates a command to create content from dialog values.
func CreateContentFromDialogCmd(datatypeID types.DatatypeID, routeID types.RouteID, parentID types.NullableContentID, fieldValues map[types.FieldID]string) tea.Cmd {
	return func() tea.Msg {
		return CreateContentFromDialogRequestMsg{
			DatatypeID:  datatypeID,
			RouteID:     routeID,
			ParentID:    parentID,
			FieldValues: fieldValues,
		}
	}
}

// ContentCreatedFromDialogMsg is sent after content is successfully created from dialog.
type ContentCreatedFromDialogMsg struct {
	ContentID  types.ContentID
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	FieldCount int
}

// FetchContentFieldsMsg triggers fetching fields for a datatype to show the content form.
type FetchContentFieldsMsg struct {
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	ParentID   types.NullableContentID
	Title      string
}

// FetchContentFieldsCmd creates a command to fetch fields for content form.
func FetchContentFieldsCmd(datatypeID types.DatatypeID, routeID types.RouteID, parentID types.NullableContentID, title string) tea.Cmd {
	return func() tea.Msg {
		return FetchContentFieldsMsg{
			DatatypeID: datatypeID,
			RouteID:    routeID,
			ParentID:   parentID,
			Title:      title,
		}
	}
}

// =============================================================================
// EDIT CONTENT FORM DIALOG
// =============================================================================

// ExistingContentField represents a field with its current value for editing.
type ExistingContentField struct {
	ContentFieldID types.ContentFieldID
	FieldID        types.FieldID
	Label          string
	Type           string
	Widget         string // UI widget override from UIConfig
	Placeholder    string // placeholder from UIConfig
	Value          string
	ValidationJSON string // raw JSON from fields.validation
	DataJSON       string // raw JSON from fields.data
	HelpText       string // help text from UIConfig, shown below label
	Hidden         bool   // hidden from UIConfig, skip in render and focus
}

// NewEditContentFormDialog creates a content form dialog pre-populated with existing values
func NewEditContentFormDialog(title string, contentID types.ContentID, datatypeID types.DatatypeID, routeID types.RouteID, existingFields []ExistingContentField) ContentFormDialogModel {
	contentFields := make([]ContentFieldInput, 0, len(existingFields))
	for _, f := range existingFields {
		bubble := resolveBubble(f.Type, f.Widget)
		applyPlaceholder(bubble, f.Placeholder)

		// For select/radio fields, parse options from the DataJSON column
		if f.Type == "select" || f.Widget == "radio" {
			if sb, ok := bubble.(*SelectBubble); ok {
				sb.ParseOptionsFromData(f.DataJSON)
			}
		}

		bubble.SetValue(f.Value)

		contentFields = append(contentFields, ContentFieldInput{
			FieldID:        f.FieldID,
			Label:          f.Label,
			Type:           f.Type,
			Widget:         f.Widget,
			Bubble:         bubble,
			ValidationJSON: f.ValidationJSON,
			DataJSON:       f.DataJSON,
			HelpText:       f.HelpText,
			Hidden:         f.Hidden,
		})
	}
	// Focus first visible field after all inputs are created
	firstVisible := 0
	for i := range contentFields {
		if !contentFields[i].Hidden {
			contentFields[i].Bubble.Focus()
			firstVisible = i
			break
		}
	}

	return ContentFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        65,
		Action:       FORMDIALOGEDITCONTENT,
		DatatypeID:   datatypeID,
		RouteID:      routeID,
		ContentID:    contentID,
		Fields:       contentFields,
		focusIndex:   firstVisible,
	}
}

// ShowEditContentFormDialogMsg triggers display of an edit content form dialog.
type ShowEditContentFormDialogMsg struct {
	Title          string
	ContentID      types.ContentID
	DatatypeID     types.DatatypeID
	RouteID        types.RouteID
	ExistingFields []ExistingContentField
}

// ShowEditContentFormDialogCmd creates a command to show a content form dialog pre-populated for editing.
func ShowEditContentFormDialogCmd(title string, contentID types.ContentID, datatypeID types.DatatypeID, routeID types.RouteID, existingFields []ExistingContentField) tea.Cmd {
	return func() tea.Msg {
		return ShowEditContentFormDialogMsg{
			Title:          title,
			ContentID:      contentID,
			DatatypeID:     datatypeID,
			RouteID:        routeID,
			ExistingFields: existingFields,
		}
	}
}

// FetchContentForEditMsg triggers fetching content fields for editing.
type FetchContentForEditMsg struct {
	ContentID  types.ContentID
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	Title      string
}

// FetchContentForEditCmd creates a command to fetch content fields for editing.
func FetchContentForEditCmd(contentID types.ContentID, datatypeID types.DatatypeID, routeID types.RouteID, title string) tea.Cmd {
	return func() tea.Msg {
		return FetchContentForEditMsg{
			ContentID:  contentID,
			DatatypeID: datatypeID,
			RouteID:    routeID,
			Title:      title,
		}
	}
}

// UpdateContentFromDialogRequestMsg triggers content update from the dialog.
type UpdateContentFromDialogRequestMsg struct {
	ContentID   types.ContentID
	DatatypeID  types.DatatypeID
	RouteID     types.RouteID
	FieldValues map[types.FieldID]string
}

// UpdateContentFromDialogCmd creates a command to update content from dialog values.
func UpdateContentFromDialogCmd(contentID types.ContentID, datatypeID types.DatatypeID, routeID types.RouteID, fieldValues map[types.FieldID]string) tea.Cmd {
	return func() tea.Msg {
		return UpdateContentFromDialogRequestMsg{
			ContentID:   contentID,
			DatatypeID:  datatypeID,
			RouteID:     routeID,
			FieldValues: fieldValues,
		}
	}
}

// ContentUpdatedFromDialogMsg is sent after content is successfully updated from dialog.
type ContentUpdatedFromDialogMsg struct {
	ContentID    types.ContentID
	DatatypeID   types.DatatypeID
	RouteID      types.RouteID
	UpdatedCount int
}
