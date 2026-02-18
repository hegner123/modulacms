package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
	"github.com/hegner123/modulacms/internal/tui"
)

// FormDialogAction identifies the type of form dialog
type FormDialogAction string

// FormDialogAction constants identify dialog types for operations across entities.
const (
	FORMDIALOGCREATEDATATYPE         FormDialogAction = "create_datatype"
	FORMDIALOGEDITDATATYPE           FormDialogAction = "edit_datatype"
	FORMDIALOGCREATEFIELD            FormDialogAction = "create_field"
	FORMDIALOGEDITFIELD              FormDialogAction = "edit_field"
	FORMDIALOGCREATEROUTE            FormDialogAction = "create_route"
	FORMDIALOGEDITROUTE              FormDialogAction = "edit_route"
	FORMDIALOGCREATEROUTEWITHCONTENT FormDialogAction = "create_route_with_content"
	FORMDIALOGINITIALIZEROUTECONTENT FormDialogAction = "initialize_route_content"
	FORMDIALOGCHILDDATATYPE          FormDialogAction = "child_datatype"
	FORMDIALOGCREATECONTENT          FormDialogAction = "create_content"
	FORMDIALOGEDITCONTENT            FormDialogAction = "edit_content"
	FORMDIALOGMOVECONTENT            FormDialogAction = "move_content"
	FORMDIALOGCREATEUSER             FormDialogAction = "create_user"
	FORMDIALOGEDITUSER               FormDialogAction = "edit_user"
	FORMDIALOGEDIITSINGLEFIELD       FormDialogAction = "edit_single_field"
	FORMDIALOGADDCONTENTFIELD        FormDialogAction = "add_content_field"
	FORMDIALOGCREATEADMINROUTE       FormDialogAction = "create_admin_route"
	FORMDIALOGEDITADMINROUTE         FormDialogAction = "edit_admin_route"
	FORMDIALOGCREATEADMINDATATYPE    FormDialogAction = "create_admin_datatype"
	FORMDIALOGEDITADMINDATATYPE      FormDialogAction = "edit_admin_datatype"
	FORMDIALOGCREATEADMINFIELD       FormDialogAction = "create_admin_field"
	FORMDIALOGEDITADMINFIELD         FormDialogAction = "edit_admin_field"
	FORMDIALOGDBINSERT               FormDialogAction = "db_insert"
	FORMDIALOGDBUPDATE               FormDialogAction = "db_update"
	FORMDIALOGCONFIGEDIT             FormDialogAction = "config_edit"
)

// FormDialogField constants define focus indices for dialog fields.
const (
	FormDialogFieldLabel = iota
	FormDialogFieldType
	FormDialogFieldParent
	FormDialogButtonCancel
	FormDialogButtonConfirm
)

// ParentOption represents a selectable parent datatype
type ParentOption struct {
	Label string
	Value string // DatatypeID or empty for ROOT
}

// TypeOption represents a selectable field type from the registry
type TypeOption struct {
	Label string
	Value string // Registry key (e.g., "text")
}

// RoleOption represents a selectable role for user forms.
type RoleOption struct {
	Label string
	Value string // RoleID ULID
}

// TypeOptionsFromRegistry builds TypeOption slice from the field input registry.
func TypeOptionsFromRegistry() []TypeOption {
	entries := FieldInputTypes()
	opts := make([]TypeOption, len(entries))
	for i, e := range entries {
		opts[i] = TypeOption{Label: e.Label, Value: e.Key}
	}
	return opts
}

// dialogStyles holds the shared styles used by all dialog types.
type dialogStyles struct {
	borderStyle        lipgloss.Style
	titleStyle         lipgloss.Style
	labelStyle         lipgloss.Style
	inputStyle         lipgloss.Style
	focusedInputStyle  lipgloss.Style
	buttonStyle        lipgloss.Style
	cancelButtonStyle  lipgloss.Style
	confirmButtonStyle lipgloss.Style
}

func newDialogStyles() dialogStyles {
	return dialogStyles{
		borderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(1, 2),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(config.DefaultStyle.Accent).
			MarginBottom(1),
		labelStyle: lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Secondary).
			MarginBottom(0),
		inputStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(config.DefaultStyle.Tertiary).
			Padding(0, 1).
			MarginBottom(1),
		focusedInputStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(config.DefaultStyle.Accent).
			Padding(0, 1).
			MarginBottom(1),
		buttonStyle: lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(2),
		cancelButtonStyle: lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Secondary).
			Background(config.DefaultStyle.Tertiary).
			Padding(0, 2).
			MarginRight(2),
		confirmButtonStyle: lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Accent).
			Padding(0, 2),
	}
}

// FormDialogModel represents a form dialog with text inputs and buttons
type FormDialogModel struct {
	dialogStyles

	Title  string
	Width  int
	Action FormDialogAction

	// EntityID is the ID of the entity being edited (empty for create operations)
	EntityID string

	// Text input fields
	LabelInput textinput.Model
	TypeInput  textinput.Model

	// Type selection (populated from registry for field dialogs)
	TypeOptions []TypeOption
	TypeIndex   int

	// Parent selection
	ParentOptions []ParentOption
	ParentIndex   int

	// Focus management: 0=Label, 1=Type, 2=Parent, 3=Cancel, 4=Confirm
	focusIndex int
}

// NewFormDialog creates a new form dialog for datatype creation
func NewFormDialog(title string, action FormDialogAction, parents []db.Datatypes) FormDialogModel {
	// Create label input
	labelInput := textinput.New()
	labelInput.Placeholder = "Display name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.Focus()

	// Create type input
	typeInput := textinput.New()
	typeInput.Placeholder = "ROOT"
	typeInput.CharLimit = 32
	typeInput.Width = 40

	// Build parent options
	parentOptions := []ParentOption{
		{Label: "ROOT (no parent)", Value: ""},
	}
	for _, p := range parents {
		parentOptions = append(parentOptions, ParentOption{
			Label: p.Label,
			Value: string(p.DatatypeID),
		})
	}

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        action,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parentOptions,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
	}
}

// NewFieldFormDialog creates a form dialog for field creation (no parent selector)
func NewFieldFormDialog(title string, action FormDialogAction) FormDialogModel {
	// Create label input
	labelInput := textinput.New()
	labelInput.Placeholder = "Field name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.Focus()

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		LabelInput:   labelInput,
		TypeOptions:  TypeOptionsFromRegistry(),
		TypeIndex:    0,
		focusIndex:   FormDialogFieldLabel,
	}
}

// NewRouteFormDialog creates a form dialog for route creation (Title and Slug inputs)
func NewRouteFormDialog(title string, action FormDialogAction) FormDialogModel {
	// Create title input (uses LabelInput field)
	titleInput := textinput.New()
	titleInput.Placeholder = "Page title"
	titleInput.CharLimit = 128
	titleInput.Width = 40
	titleInput.Focus()

	// Create slug input (uses TypeInput field)
	slugInput := textinput.New()
	slugInput.Placeholder = "url-slug"
	slugInput.CharLimit = 128
	slugInput.Width = 40

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		LabelInput:   titleInput,
		TypeInput:    slugInput,
		focusIndex:   FormDialogFieldLabel,
	}
}

// HasParentSelector returns true if the dialog has a parent selector
func (d *FormDialogModel) HasParentSelector() bool {
	return len(d.ParentOptions) > 0
}

// HasTypeSelector returns true if the dialog has a type selector carousel
func (d *FormDialogModel) HasTypeSelector() bool {
	return len(d.TypeOptions) > 0
}

// HasSecondField returns true if the dialog should render the second (Type/Slug) field.
func (d *FormDialogModel) HasSecondField() bool {
	return d.Action != FORMDIALOGCONFIGEDIT
}

// Update handles input for the form dialog
func (d *FormDialogModel) Update(msg tea.Msg) (FormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Special handling for child datatype selection and move content - simple vertical list
		if d.Action == FORMDIALOGCHILDDATATYPE || d.Action == FORMDIALOGMOVECONTENT {
			return d.updateChildDatatypeSelection(msg)
		}

		switch msg.String() {
		case "tab", "down":
			d.focusNext()
			return *d, nil
		case "shift+tab", "up":
			d.focusPrev()
			return *d, nil
		case "left":
			if d.focusIndex == FormDialogFieldType && d.HasTypeSelector() {
				if d.TypeIndex > 0 {
					d.TypeIndex--
				}
				return *d, nil
			}
			if d.focusIndex == FormDialogFieldParent && d.HasParentSelector() {
				if d.ParentIndex > 0 {
					d.ParentIndex--
				}
				return *d, nil
			}
			if d.focusIndex == FormDialogButtonConfirm {
				d.focusIndex = FormDialogButtonCancel
				return *d, nil
			}
		case "right":
			if d.focusIndex == FormDialogFieldType && d.HasTypeSelector() {
				if d.TypeIndex < len(d.TypeOptions)-1 {
					d.TypeIndex++
				}
				return *d, nil
			}
			if d.focusIndex == FormDialogFieldParent && d.HasParentSelector() {
				if d.ParentIndex < len(d.ParentOptions)-1 {
					d.ParentIndex++
				}
				return *d, nil
			}
			if d.focusIndex == FormDialogButtonCancel {
				d.focusIndex = FormDialogButtonConfirm
				return *d, nil
			}
		case "enter":
			if d.focusIndex == FormDialogButtonCancel {
				return *d, func() tea.Msg { return FormDialogCancelMsg{} }
			}
			if d.focusIndex == FormDialogButtonConfirm {
				// Safely get parent ID (empty for dialogs without parent selector)
				parentID := ""
				if d.HasParentSelector() && d.ParentIndex < len(d.ParentOptions) {
					parentID = d.ParentOptions[d.ParentIndex].Value
				}

				// Get type value from selector or text input
				typeValue := d.TypeInput.Value()
				if d.HasTypeSelector() && d.TypeIndex < len(d.TypeOptions) {
					typeValue = d.TypeOptions[d.TypeIndex].Value
				}

				return *d, func() tea.Msg {
					return FormDialogAcceptMsg{
						Action:   d.Action,
						EntityID: d.EntityID,
						Label:    d.LabelInput.Value(),
						Type:     typeValue,
						ParentID: parentID,
					}
				}
			}
			// On text fields, enter moves to next field
			d.focusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return FormDialogCancelMsg{} }
		}

		// Update the focused text input
		var cmd tea.Cmd
		switch d.focusIndex {
		case FormDialogFieldLabel:
			d.LabelInput, cmd = d.LabelInput.Update(msg)
		case FormDialogFieldType:
			if !d.HasTypeSelector() {
				d.TypeInput, cmd = d.TypeInput.Update(msg)
			}
		}
		return *d, cmd
	}

	return *d, nil
}

// updateChildDatatypeSelection handles input for the child datatype selection dialog
func (d *FormDialogModel) updateChildDatatypeSelection(msg tea.KeyMsg) (FormDialogModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if d.ParentIndex > 0 {
			d.ParentIndex--
		}
		return *d, nil
	case "down", "j":
		if d.ParentIndex < len(d.ParentOptions)-1 {
			d.ParentIndex++
		}
		return *d, nil
	case "enter":
		// Accept the selected child datatype
		if len(d.ParentOptions) > 0 && d.ParentIndex < len(d.ParentOptions) {
			parentID := d.ParentOptions[d.ParentIndex].Value
			return *d, func() tea.Msg {
				return FormDialogAcceptMsg{
					Action:   d.Action,
					EntityID: d.EntityID,
					ParentID: parentID,
				}
			}
		}
		return *d, nil
	case "esc", "q":
		return *d, func() tea.Msg { return FormDialogCancelMsg{} }
	}
	return *d, nil
}

// focusNext advances focus to the next focusable element, wrapping at the end.
func (d *FormDialogModel) focusNext() {
	d.focusIndex++
	// Skip type field if not applicable
	if d.focusIndex == FormDialogFieldType && !d.HasSecondField() {
		d.focusIndex = FormDialogButtonCancel
	}
	// Skip parent field if no parent options
	if d.focusIndex == FormDialogFieldParent && !d.HasParentSelector() {
		d.focusIndex = FormDialogButtonCancel
	}
	if d.focusIndex > FormDialogButtonConfirm {
		d.focusIndex = FormDialogFieldLabel
	}
	d.updateFocus()
}

// focusPrev moves focus to the previous focusable element, wrapping at the start.
func (d *FormDialogModel) focusPrev() {
	d.focusIndex--
	// Skip parent field if no parent options
	if d.focusIndex == FormDialogFieldParent && !d.HasParentSelector() {
		d.focusIndex = FormDialogFieldType
	}
	// Skip type field if not applicable
	if d.focusIndex == FormDialogFieldType && !d.HasSecondField() {
		d.focusIndex = FormDialogFieldLabel
	}
	if d.focusIndex < FormDialogFieldLabel {
		d.focusIndex = FormDialogButtonConfirm
	}
	d.updateFocus()
}

// updateFocus applies focus styling to the currently focused element.
func (d *FormDialogModel) updateFocus() {
	d.LabelInput.Blur()
	if !d.HasTypeSelector() {
		d.TypeInput.Blur()
	}

	switch d.focusIndex {
	case FormDialogFieldLabel:
		d.LabelInput.Focus()
	case FormDialogFieldType:
		if !d.HasTypeSelector() {
			d.TypeInput.Focus()
		}
	}
}

// Render renders the form dialog
func (d FormDialogModel) Render(windowWidth, windowHeight int) string {
	var content strings.Builder
	// Inner width available for fields (dialog width minus border and padding)
	innerW := d.Width - 6

	// Title
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Special rendering for child datatype selection and move content - vertical list
	if d.Action == FORMDIALOGCHILDDATATYPE || d.Action == FORMDIALOGMOVECONTENT {
		return d.renderChildDatatypeSelection(windowWidth, windowHeight)
	}

	// Determine field labels based on action type
	firstFieldLabel := "Label"
	secondFieldLabel := "Type"
	if d.Action == FORMDIALOGCREATEROUTE || d.Action == FORMDIALOGEDITROUTE ||
		d.Action == FORMDIALOGCREATEADMINROUTE || d.Action == FORMDIALOGEDITADMINROUTE ||
		d.Action == FORMDIALOGCREATEROUTEWITHCONTENT {
		firstFieldLabel = "Title"
		secondFieldLabel = "Slug"
	}
	if d.Action == FORMDIALOGCONFIGEDIT {
		firstFieldLabel = "Value"
	}

	// First field (Label, Title, or Value)
	labelLabel := d.labelStyle.Render(firstFieldLabel)
	content.WriteString(labelLabel)
	content.WriteString("\n")
	if d.focusIndex == FormDialogFieldLabel {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.LabelInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.LabelInput.View()))
	}
	content.WriteString("\n")

	// Second field (Type or Slug) — skip for single-field dialogs
	if d.HasSecondField() {
		typeLabel := d.labelStyle.Render(secondFieldLabel)
		content.WriteString(typeLabel)
		content.WriteString("\n")
		if d.HasTypeSelector() {
			optLabel := d.TypeOptions[d.TypeIndex].Label
			if d.focusIndex == FormDialogFieldType {
				selector := lipgloss.NewStyle().
					Foreground(config.DefaultStyle.Primary).
					Background(config.DefaultStyle.Accent).
					Padding(0, 1).
					Render("◀ " + optLabel + " ▶")
				content.WriteString(selector)
			} else {
				selector := lipgloss.NewStyle().
					Foreground(config.DefaultStyle.Secondary).
					Padding(0, 1).
					Render("  " + optLabel + "  ")
				content.WriteString(selector)
			}
		} else {
			if d.focusIndex == FormDialogFieldType {
				content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.TypeInput.View()))
			} else {
				content.WriteString(d.inputStyle.Width(innerW).Render(d.TypeInput.View()))
			}
		}
		content.WriteString("\n")
	}

	// Parent selector (only for dialogs with parent options)
	if d.HasParentSelector() {
		parentLabelText := "Parent"
		if d.Action == FORMDIALOGCREATEROUTEWITHCONTENT {
			parentLabelText = "Datatype"
		}
		parentLabel := d.labelStyle.Render(parentLabelText)
		content.WriteString(parentLabel)
		content.WriteString("\n")
		parentValue := d.ParentOptions[d.ParentIndex].Label
		if d.focusIndex == FormDialogFieldParent {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Primary).
				Background(config.DefaultStyle.Accent).
				Padding(0, 1).
				Render("◀ " + parentValue + " ▶")
			content.WriteString(selector)
		} else {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Secondary).
				Padding(0, 1).
				Render("  " + parentValue + "  ")
			content.WriteString(selector)
		}
		content.WriteString("\n\n")
	} else {
		content.WriteString("\n")
	}

	// Buttons
	cancelBtn := d.renderCancelButton()
	confirmBtn := d.renderConfirmButton()
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, confirmBtn)
	content.WriteString(buttonRow)

	// Apply border
	dialogBox := d.borderStyle.Width(d.Width).Render(content.String())

	return dialogBox
}

// renderChildDatatypeSelection renders a vertical list for selecting a child datatype
func (d FormDialogModel) renderChildDatatypeSelection(windowWidth, windowHeight int) string {
	var content strings.Builder

	// Title
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Render vertical list of child datatypes
	selectedStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Primary).
		Background(config.DefaultStyle.Accent).
		Padding(0, 2)
	normalStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary).
		Padding(0, 2)

	for i, opt := range d.ParentOptions {
		if i == d.ParentIndex {
			content.WriteString(selectedStyle.Render("▸ " + opt.Label))
		} else {
			content.WriteString(normalStyle.Render("  " + opt.Label))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Italic(true)
	content.WriteString(helpStyle.Render("↑/↓ select • enter confirm • esc cancel"))

	// Apply border
	dialogBox := d.borderStyle.Width(d.Width).Render(content.String())

	return dialogBox
}

// renderCancelButton returns the styled cancel button view.
func (d FormDialogModel) renderCancelButton() string {
	style := d.cancelButtonStyle
	if d.focusIndex == FormDialogButtonCancel {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Tertiary).
			Bold(true)
	}
	return style.Render("Cancel")
}

// renderConfirmButton returns the styled confirm button view.
func (d FormDialogModel) renderConfirmButton() string {
	style := d.confirmButtonStyle
	if d.focusIndex == FormDialogButtonConfirm {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Accent).
			Bold(true)
	}
	return style.Render("Confirm")
}

// FormDialogOverlay positions a form dialog over existing content
func FormDialogOverlay(content string, dialog FormDialogModel, width, height int) string {
	dialogContent := dialog.Render(width, height)
	dialogW := lipgloss.Width(dialogContent)
	dialogH := lipgloss.Height(dialogContent)

	x := (width - dialogW) / 2
	y := (height - dialogH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return tui.Composite(content, tui.Overlay{
		Content: dialogContent,
		X:       x,
		Y:       y,
		Width:   dialogW,
		Height:  dialogH,
	})
}

// FormDialogAcceptMsg carries form dialog acceptance data.
type FormDialogAcceptMsg struct {
	Action   FormDialogAction
	EntityID string // ID of entity being edited (empty for create)
	Label    string
	Type     string
	ParentID string
}

// FormDialogCancelMsg is sent when a form dialog is cancelled.
type FormDialogCancelMsg struct{}

// ShowFormDialogMsg triggers display of a form dialog.
type ShowFormDialogMsg struct {
	Action  FormDialogAction
	Title   string
	Parents []db.Datatypes
}

// ShowFormDialogCmd creates a command to display a form dialog.
func ShowFormDialogCmd(action FormDialogAction, title string, parents []db.Datatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowFormDialogMsg{
			Action:  action,
			Title:   title,
			Parents: parents,
		}
	}
}

// ShowFieldFormDialogCmd shows a field creation dialog (no parent selector)
func ShowFieldFormDialogCmd(action FormDialogAction, title string) tea.Cmd {
	return func() tea.Msg {
		return ShowFieldFormDialogMsg{
			Action: action,
			Title:  title,
		}
	}
}

// ShowFieldFormDialogMsg is the message for showing a field form dialog
type ShowFieldFormDialogMsg struct {
	Action FormDialogAction
	Title  string
}

// ShowRouteFormDialogCmd shows a route creation dialog
func ShowRouteFormDialogCmd(action FormDialogAction, title string) tea.Cmd {
	return func() tea.Msg {
		return ShowRouteFormDialogMsg{
			Action: action,
			Title:  title,
		}
	}
}

// ShowRouteFormDialogMsg is the message for showing a route form dialog
type ShowRouteFormDialogMsg struct {
	Action FormDialogAction
	Title  string
}

func FormDialogSetCmd(dialog *FormDialogModel) tea.Cmd {
	return func() tea.Msg {
		return FormDialogSetMsg{Dialog: dialog}
	}
}

// FormDialogSetMsg carries a form dialog model to update.
type FormDialogSetMsg struct {
	Dialog *FormDialogModel
}

// FormDialogActiveSetMsg carries the active state for a form dialog.
type FormDialogActiveSetMsg struct {
	Active bool
}

// FormDialogActiveSetCmd creates a command to set the form dialog active state.
func FormDialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg {
		return FormDialogActiveSetMsg{Active: active}
	}
}

// NewEditDatatypeDialog creates a form dialog for editing a datatype with pre-populated values
func NewEditDatatypeDialog(title string, action FormDialogAction, parents []db.Datatypes, datatype db.Datatypes) FormDialogModel {
	// Create label input with current value
	labelInput := textinput.New()
	labelInput.Placeholder = "Display name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.SetValue(datatype.Label)
	labelInput.Focus()

	// Create type input with current value
	typeInput := textinput.New()
	typeInput.Placeholder = "ROOT"
	typeInput.CharLimit = 32
	typeInput.Width = 40
	typeInput.SetValue(datatype.Type)

	// Build parent options
	parentOptions := []ParentOption{
		{Label: "ROOT (no parent)", Value: ""},
	}
	selectedParentIndex := 0
	for _, p := range parents {
		// Skip self to prevent circular reference
		if p.DatatypeID == datatype.DatatypeID {
			continue
		}
		parentOptions = append(parentOptions, ParentOption{
			Label: p.Label,
			Value: string(p.DatatypeID),
		})
		// Check if this is the current parent
		if datatype.ParentID.Valid && string(datatype.ParentID.ID) == string(p.DatatypeID) {
			selectedParentIndex = len(parentOptions) - 1
		}
	}

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        action,
		EntityID:      string(datatype.DatatypeID),
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parentOptions,
		ParentIndex:   selectedParentIndex,
		focusIndex:    FormDialogFieldLabel,
	}
}

// NewEditFieldDialog creates a form dialog for editing a field with pre-populated values
func NewEditFieldDialog(title string, action FormDialogAction, field db.Fields) FormDialogModel {
	// Create label input with current value
	labelInput := textinput.New()
	labelInput.Placeholder = "Field name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.SetValue(field.Label)
	labelInput.Focus()

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		EntityID:     string(field.FieldID),
		LabelInput:   labelInput,
		TypeOptions:  TypeOptionsFromRegistry(),
		TypeIndex:    FieldInputTypeIndex(string(field.Type)),
		focusIndex:   FormDialogFieldLabel,
	}
}

// NewEditRouteDialog creates a form dialog for editing a route with pre-populated values
func NewEditRouteDialog(title string, action FormDialogAction, route db.Routes) FormDialogModel {
	// Create title input with current value (uses LabelInput field)
	titleInput := textinput.New()
	titleInput.Placeholder = "Page title"
	titleInput.CharLimit = 128
	titleInput.Width = 40
	titleInput.SetValue(route.Title)
	titleInput.Focus()

	// Create slug input with current value (uses TypeInput field)
	slugInput := textinput.New()
	slugInput.Placeholder = "url-slug"
	slugInput.CharLimit = 128
	slugInput.Width = 40
	slugInput.SetValue(string(route.Slug))

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		EntityID:     string(route.RouteID),
		LabelInput:   titleInput,
		TypeInput:    slugInput,
		focusIndex:   FormDialogFieldLabel,
	}
}

// ShowEditDatatypeDialogMsg is the message for showing an edit datatype dialog
type ShowEditDatatypeDialogMsg struct {
	Datatype db.Datatypes
	Parents  []db.Datatypes
}

// ShowEditDatatypeDialogCmd shows an edit dialog for a datatype
func ShowEditDatatypeDialogCmd(datatype db.Datatypes, parents []db.Datatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowEditDatatypeDialogMsg{
			Datatype: datatype,
			Parents:  parents,
		}
	}
}

// ShowEditFieldDialogMsg is the message for showing an edit field dialog
type ShowEditFieldDialogMsg struct {
	Field db.Fields
}

// ShowEditFieldDialogCmd shows an edit dialog for a field
func ShowEditFieldDialogCmd(field db.Fields) tea.Cmd {
	return func() tea.Msg {
		return ShowEditFieldDialogMsg{
			Field: field,
		}
	}
}

// ShowEditRouteDialogMsg is the message for showing an edit route dialog
type ShowEditRouteDialogMsg struct {
	Route db.Routes
}

// ShowEditRouteDialogCmd shows an edit dialog for a route
func ShowEditRouteDialogCmd(route db.Routes) tea.Cmd {
	return func() tea.Msg {
		return ShowEditRouteDialogMsg{
			Route: route,
		}
	}
}

// NewRouteWithContentDialog creates a form dialog for creating a new route with initial content
func NewRouteWithContentDialog(title string, action FormDialogAction, rootDatatypes []db.Datatypes) FormDialogModel {
	// Create title input (uses LabelInput field)
	titleInput := textinput.New()
	titleInput.Placeholder = "Page title"
	titleInput.CharLimit = 128
	titleInput.Width = 40
	titleInput.Focus()

	// Create slug input (uses TypeInput field)
	slugInput := textinput.New()
	slugInput.Placeholder = "url-slug"
	slugInput.CharLimit = 128
	slugInput.Width = 40

	// Build datatype options using ParentOptions carousel
	parentOptions := make([]ParentOption, len(rootDatatypes))
	for i, dt := range rootDatatypes {
		parentOptions[i] = ParentOption{
			Label: dt.Label,
			Value: string(dt.DatatypeID),
		}
	}

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        action,
		LabelInput:    titleInput,
		TypeInput:     slugInput,
		ParentOptions: parentOptions,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
	}
}

// ShowCreateRouteWithContentDialogMsg is the message for showing a create route with content dialog
type ShowCreateRouteWithContentDialogMsg struct {
	RootDatatypes []db.Datatypes
}

// ShowCreateRouteWithContentDialogCmd shows a dialog to create a new route with initial content
func ShowCreateRouteWithContentDialogCmd(rootDatatypes []db.Datatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowCreateRouteWithContentDialogMsg{
			RootDatatypes: rootDatatypes,
		}
	}
}

// ShowInitializeRouteContentDialogMsg is the message for initializing content on an existing route
type ShowInitializeRouteContentDialogMsg struct {
	Route      db.Routes
	DatatypeID string
}

// ShowInitializeRouteContentDialogCmd shows a confirmation dialog to initialize content for a route
func ShowInitializeRouteContentDialogCmd(route db.Routes, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		return ShowInitializeRouteContentDialogMsg{
			Route:      route,
			DatatypeID: datatypeID,
		}
	}
}

// ShowChildDatatypeDialogMsg is the message for showing a child datatype selection dialog
type ShowChildDatatypeDialogMsg struct {
	ParentDatatypeID string
	RouteID          string
	ChildDatatypes   []db.Datatypes
}

// ShowChildDatatypeDialogCmd fetches child datatypes and shows a selection dialog
func ShowChildDatatypeDialogCmd(parentDatatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return FetchChildDatatypesMsg{
			ParentDatatypeID: parentDatatypeID,
			RouteID:          routeID,
		}
	}
}

// FetchChildDatatypesMsg triggers fetching child datatypes for a parent
type FetchChildDatatypesMsg struct {
	ParentDatatypeID types.DatatypeID
	RouteID          types.RouteID
}

// ChildDatatypeSelectedMsg is sent when a child datatype is selected from the dialog
type ChildDatatypeSelectedMsg struct {
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
}

// ChildDatatypeSelectedCmd creates a command that returns a ChildDatatypeSelectedMsg
func ChildDatatypeSelectedCmd(datatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return ChildDatatypeSelectedMsg{
			DatatypeID: datatypeID,
			RouteID:    routeID,
		}
	}
}

// NewChildDatatypeDialog creates a dialog for selecting a child datatype
func NewChildDatatypeDialog(title string, childDatatypes []db.Datatypes, routeID string) FormDialogModel {
	// Build parent options from child datatypes
	parents := make([]ParentOption, 0, len(childDatatypes))
	for _, dt := range childDatatypes {
		parents = append(parents, ParentOption{
			Label: dt.Label,
			Value: string(dt.DatatypeID),
		})
	}

	// Initialize text inputs even though they're not displayed
	// This prevents nil pointer panics when updateFocus is called
	labelInput := textinput.New()
	labelInput.Placeholder = ""
	typeInput := textinput.New()
	typeInput.Placeholder = ""

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         50,
		Action:        FORMDIALOGCHILDDATATYPE,
		EntityID:      routeID,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parents,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldParent, // Start on selection
	}
}

// =============================================================================
// MOVE CONTENT DIALOG
// =============================================================================

// ShowMoveContentDialogMsg triggers showing the move content dialog
type ShowMoveContentDialogMsg struct {
	SourceNode   *tree.Node
	RouteID      types.RouteID
	ValidTargets []ParentOption
}

// MoveContentRequestMsg triggers the actual content move operation
type MoveContentRequestMsg struct {
	SourceContentID types.ContentID
	TargetContentID types.ContentID
	RouteID         types.RouteID
}

// ContentMovedMsg is sent after content is successfully moved
type ContentMovedMsg struct {
	SourceContentID types.ContentID
	TargetContentID types.ContentID
	RouteID         types.RouteID
}

// ShowMoveContentDialogCmd creates a command to show the move content dialog
func ShowMoveContentDialogCmd(node *tree.Node, routeID types.RouteID, targets []ParentOption) tea.Cmd {
	return func() tea.Msg {
		return ShowMoveContentDialogMsg{
			SourceNode:   node,
			RouteID:      routeID,
			ValidTargets: targets,
		}
	}
}

// MoveContentCmd creates a command that returns a MoveContentRequestMsg
func MoveContentCmd(sourceID, targetID types.ContentID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return MoveContentRequestMsg{
			SourceContentID: sourceID,
			TargetContentID: targetID,
			RouteID:         routeID,
		}
	}
}

// NewMoveContentDialog creates a dialog for selecting a move target
func NewMoveContentDialog(title string, sourceContentID string, routeID string, targets []ParentOption) FormDialogModel {
	// Initialize text inputs even though they're not displayed
	// This prevents nil pointer panics when updateFocus is called
	labelInput := textinput.New()
	labelInput.Placeholder = ""
	typeInput := textinput.New()
	typeInput.Placeholder = ""

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         50,
		Action:        FORMDIALOGMOVECONTENT,
		EntityID:      sourceContentID + "|" + routeID,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: targets,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldParent, // Start on selection
	}
}

// =============================================================================
// ContentFormDialogModel - Dynamic content fields dialog
// =============================================================================

// ContentFieldInput represents a single field input in the content form.
type ContentFieldInput struct {
	FieldID types.FieldID
	Label   string
	Type    string // field type (text, textarea, number, etc.)
	Widget  string // UI widget override from UIConfig (e.g. "markdown", "code-editor")
	Bubble  FieldBubble
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

	// Logger for editor and dialog operations (nil-safe; callers should set after construction)
	Logger Logger

	// Focus management: 0 to len(Fields)-1 for fields, then Cancel, then Confirm
	focusIndex int
}

// NewContentFormDialog creates a new content form dialog with dynamic fields
func NewContentFormDialog(title string, action FormDialogAction, datatypeID types.DatatypeID, routeID types.RouteID, fields []db.Fields) ContentFormDialogModel {
	contentFields := make([]ContentFieldInput, 0, len(fields))
	for _, f := range fields {
		bubble := FieldBubbleForType(string(f.Type))

		// For select fields, parse options from the Data JSON column
		if string(f.Type) == "select" {
			if sb, ok := bubble.(*SelectBubble); ok {
				sb.ParseOptionsFromData(f.Data)
			}
		}

		// Parse UIConfig to extract widget override
		var widget string
		if uc, err := types.ParseUIConfig(f.UIConfig); err == nil {
			widget = uc.Widget
		}

		contentFields = append(contentFields, ContentFieldInput{
			FieldID: f.FieldID,
			Label:   f.Label,
			Type:    string(f.Type),
			Widget:  widget,
			Bubble:  bubble,
		})
	}
	// Focus first field after all inputs are created
	if len(contentFields) > 0 {
		contentFields[0].Bubble.Focus()
	}

	return ContentFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        65,
		Action:       action,
		DatatypeID:   datatypeID,
		RouteID:      routeID,
		Fields:       contentFields,
		focusIndex:   0,
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
			return *d, cmd
		}
	}

	return *d, nil
}

// focusNext advances focus to the next focusable element in the content form, wrapping at the end.
func (d *ContentFormDialogModel) focusNext() {
	d.focusIndex++
	if d.focusIndex > d.ButtonConfirmIndex() {
		d.focusIndex = 0
	}
	d.updateFocus()
}

// focusPrev moves focus to the previous focusable element in the content form, wrapping at the start.
func (d *ContentFormDialogModel) focusPrev() {
	d.focusIndex--
	if d.focusIndex < 0 {
		d.focusIndex = d.ButtonConfirmIndex()
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

// Render renders the content form dialog
func (d ContentFormDialogModel) Render(windowWidth, windowHeight int) string {
	var content strings.Builder
	// Inner width available for fields (dialog width minus border and padding)
	innerW := d.Width - 6

	// Title
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Render each field
	editorHintStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Italic(true)
	for i, f := range d.Fields {
		f.Bubble.SetWidth(innerW)
		label := d.labelStyle.Render(f.Label)
		if isEditorWidget(f.Widget) {
			label += " " + editorHintStyle.Render("ctrl+e: $EDITOR")
		}
		content.WriteString(label)
		content.WriteString("\n")

		if d.focusIndex == i {
			content.WriteString(d.focusedInputStyle.Width(innerW).Render(f.Bubble.View()))
		} else {
			content.WriteString(d.inputStyle.Width(innerW).Render(f.Bubble.View()))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Buttons
	cancelBtn := d.renderCancelButton()
	confirmBtn := d.renderConfirmButton()
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, confirmBtn)
	content.WriteString(buttonRow)

	// Apply border
	dialogBox := d.borderStyle.Width(d.Width).Render(content.String())

	return dialogBox
}

// renderCancelButton returns the styled cancel button view for the content form.
func (d ContentFormDialogModel) renderCancelButton() string {
	style := d.cancelButtonStyle
	if d.focusIndex == d.ButtonCancelIndex() {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Tertiary).
			Bold(true)
	}
	return style.Render("Cancel")
}

// renderConfirmButton returns the styled confirm button view for the content form.
func (d ContentFormDialogModel) renderConfirmButton() string {
	style := d.confirmButtonStyle
	if d.focusIndex == d.ButtonConfirmIndex() {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Accent).
			Bold(true)
	}
	return style.Render("Confirm")
}

// ContentFormDialogOverlay positions a content form dialog over existing content
func ContentFormDialogOverlay(content string, dialog ContentFormDialogModel, width, height int) string {
	dialogContent := dialog.Render(width, height)
	dialogW := lipgloss.Width(dialogContent)
	dialogH := lipgloss.Height(dialogContent)

	x := (width - dialogW) / 2
	y := (height - dialogH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return tui.Composite(content, tui.Overlay{
		Content: dialogContent,
		X:       x,
		Y:       y,
		Width:   dialogW,
		Height:  dialogH,
	})
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

// ContentFormDialogSetCmd creates a command to set the content form dialog model.
func ContentFormDialogSetCmd(dialog *ContentFormDialogModel) tea.Cmd {
	return func() tea.Msg {
		return ContentFormDialogSetMsg{Dialog: dialog}
	}
}

// ContentFormDialogSetMsg carries a content form dialog model to update.
type ContentFormDialogSetMsg struct {
	Dialog *ContentFormDialogModel
}

// ContentFormDialogActiveSetMsg carries the active state for a content form dialog.
type ContentFormDialogActiveSetMsg struct {
	Active bool
}

// ContentFormDialogActiveSetCmd creates a command to set the content form dialog active state.
func ContentFormDialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg {
		return ContentFormDialogActiveSetMsg{Active: active}
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
	Value          string
}

// NewEditContentFormDialog creates a content form dialog pre-populated with existing values
func NewEditContentFormDialog(title string, contentID types.ContentID, datatypeID types.DatatypeID, routeID types.RouteID, existingFields []ExistingContentField) ContentFormDialogModel {
	contentFields := make([]ContentFieldInput, 0, len(existingFields))
	for _, f := range existingFields {
		bubble := FieldBubbleForType(f.Type)
		bubble.SetValue(f.Value)
		contentFields = append(contentFields, ContentFieldInput{
			FieldID: f.FieldID,
			Label:   f.Label,
			Type:    f.Type,
			Widget:  f.Widget,
			Bubble:  bubble,
		})
	}
	// Focus first field after all inputs are created
	if len(contentFields) > 0 {
		contentFields[0].Bubble.Focus()
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
		focusIndex:   0,
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

// =============================================================================
// USER FORM DIALOG
// =============================================================================

// UserFormDialogModel represents a form dialog for user CRUD operations.
type UserFormDialogModel struct {
	dialogStyles

	Title    string
	Width    int
	Action   FormDialogAction
	EntityID string // UserID for edit operations

	UsernameInput textinput.Model
	NameInput     textinput.Model
	EmailInput    textinput.Model
	PasswordInput textinput.Model

	// Role selection carousel
	RoleOptions []RoleOption
	RoleIndex   int

	// Create mode focus: 0=username, 1=name, 2=email, 3=password, 4=role, 5=cancel, 6=confirm
	// Edit mode focus:   0=username, 1=name, 2=email, 3=role, 4=cancel, 5=confirm
	focusIndex int
}

// isCreateMode returns true if the dialog is for creating a new user.
func (d *UserFormDialogModel) isCreateMode() bool {
	return d.Action == FORMDIALOGCREATEUSER
}

// maxFocusIndex returns the maximum valid focus index.
func (d *UserFormDialogModel) maxFocusIndex() int {
	if d.isCreateMode() {
		return 6
	}
	return 5
}

// roleFocusIndex returns the focus index for the role carousel.
func (d *UserFormDialogModel) roleFocusIndex() int {
	if d.isCreateMode() {
		return 4
	}
	return 3
}

// cancelFocusIndex returns the focus index for the cancel button.
func (d *UserFormDialogModel) cancelFocusIndex() int {
	if d.isCreateMode() {
		return 5
	}
	return 4
}

// confirmFocusIndex returns the focus index for the confirm button.
func (d *UserFormDialogModel) confirmFocusIndex() int {
	if d.isCreateMode() {
		return 6
	}
	return 5
}

// NewUserFormDialog creates a user form dialog for creating a new user
func NewUserFormDialog(title string, roles []db.Roles) UserFormDialogModel {
	username := textinput.New()
	username.Placeholder = "username"
	username.CharLimit = 64
	username.Width = 40
	username.Focus()

	name := textinput.New()
	name.Placeholder = "Full Name"
	name.CharLimit = 128
	name.Width = 40

	email := textinput.New()
	email.Placeholder = "user@example.com"
	email.CharLimit = 128
	email.Width = 40

	password := textinput.New()
	password.Placeholder = "password"
	password.CharLimit = 72
	password.Width = 40
	password.EchoMode = textinput.EchoPassword

	roleOptions := make([]RoleOption, 0, len(roles))
	for _, r := range roles {
		roleOptions = append(roleOptions, RoleOption{
			Label: r.Label,
			Value: string(r.RoleID),
		})
	}

	return UserFormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        FORMDIALOGCREATEUSER,
		UsernameInput: username,
		NameInput:     name,
		EmailInput:    email,
		PasswordInput: password,
		RoleOptions:   roleOptions,
		RoleIndex:     0,
		focusIndex:    0,
	}
}

// NewEditUserFormDialog creates a user form dialog pre-populated for editing
func NewEditUserFormDialog(title string, user db.UserWithRoleLabelRow, roles []db.Roles) UserFormDialogModel {
	username := textinput.New()
	username.Placeholder = "username"
	username.CharLimit = 64
	username.Width = 40
	username.SetValue(user.Username)
	username.Focus()

	name := textinput.New()
	name.Placeholder = "Full Name"
	name.CharLimit = 128
	name.Width = 40
	name.SetValue(user.Name)

	email := textinput.New()
	email.Placeholder = "user@example.com"
	email.CharLimit = 128
	email.Width = 40
	email.SetValue(user.Email.String())

	// Password input exists but is not shown in edit mode
	password := textinput.New()
	password.Placeholder = ""
	password.CharLimit = 72
	password.Width = 40
	password.EchoMode = textinput.EchoPassword

	roleOptions := make([]RoleOption, 0, len(roles))
	selectedRoleIndex := 0
	for i, r := range roles {
		roleOptions = append(roleOptions, RoleOption{
			Label: r.Label,
			Value: string(r.RoleID),
		})
		if string(r.RoleID) == user.Role {
			selectedRoleIndex = i
		}
	}

	return UserFormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        FORMDIALOGEDITUSER,
		EntityID:      user.UserID.String(),
		UsernameInput: username,
		NameInput:     name,
		EmailInput:    email,
		PasswordInput: password,
		RoleOptions:   roleOptions,
		RoleIndex:     selectedRoleIndex,
		focusIndex:    0,
	}
}

// Update handles user input for the user form dialog
func (d *UserFormDialogModel) Update(msg tea.Msg) (UserFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			d.userFormFocusNext()
			return *d, nil
		case "shift+tab", "up":
			d.userFormFocusPrev()
			return *d, nil
		case "left":
			if d.focusIndex == d.roleFocusIndex() && len(d.RoleOptions) > 0 {
				if d.RoleIndex > 0 {
					d.RoleIndex--
				}
				return *d, nil
			}
			if d.focusIndex == d.confirmFocusIndex() {
				d.focusIndex = d.cancelFocusIndex()
				return *d, nil
			}
		case "right":
			if d.focusIndex == d.roleFocusIndex() && len(d.RoleOptions) > 0 {
				if d.RoleIndex < len(d.RoleOptions)-1 {
					d.RoleIndex++
				}
				return *d, nil
			}
			if d.focusIndex == d.cancelFocusIndex() {
				d.focusIndex = d.confirmFocusIndex()
				return *d, nil
			}
		case "enter":
			if d.focusIndex == d.confirmFocusIndex() {
				roleValue := ""
				if len(d.RoleOptions) > 0 && d.RoleIndex < len(d.RoleOptions) {
					roleValue = d.RoleOptions[d.RoleIndex].Value
				}
				password := ""
				if d.isCreateMode() {
					password = d.PasswordInput.Value()
				}
				return *d, func() tea.Msg {
					return UserFormDialogAcceptMsg{
						Action:   d.Action,
						EntityID: d.EntityID,
						Username: d.UsernameInput.Value(),
						Name:     d.NameInput.Value(),
						Email:    d.EmailInput.Value(),
						Password: password,
						Role:     roleValue,
					}
				}
			}
			if d.focusIndex == d.cancelFocusIndex() {
				return *d, func() tea.Msg { return UserFormDialogCancelMsg{} }
			}
			// On text fields, move to next
			d.userFormFocusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return UserFormDialogCancelMsg{} }
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	switch d.focusIndex {
	case 0:
		d.UsernameInput, cmd = d.UsernameInput.Update(msg)
	case 1:
		d.NameInput, cmd = d.NameInput.Update(msg)
	case 2:
		d.EmailInput, cmd = d.EmailInput.Update(msg)
	default:
		if d.isCreateMode() && d.focusIndex == 3 {
			d.PasswordInput, cmd = d.PasswordInput.Update(msg)
		}
	}
	return *d, cmd
}

// userFormFocusNext advances focus to the next focusable element in the user form, wrapping at the end.
func (d *UserFormDialogModel) userFormFocusNext() {
	max := d.maxFocusIndex()
	d.focusIndex = (d.focusIndex + 1) % (max + 1)
	d.userFormUpdateFocus()
}

// userFormFocusPrev moves focus to the previous focusable element in the user form, wrapping at the start.
func (d *UserFormDialogModel) userFormFocusPrev() {
	max := d.maxFocusIndex()
	d.focusIndex = (d.focusIndex + max) % (max + 1)
	d.userFormUpdateFocus()
}

// userFormUpdateFocus applies focus styling to the currently focused field in the user form.
func (d *UserFormDialogModel) userFormUpdateFocus() {
	d.UsernameInput.Blur()
	d.NameInput.Blur()
	d.EmailInput.Blur()
	d.PasswordInput.Blur()
	switch d.focusIndex {
	case 0:
		d.UsernameInput.Focus()
	case 1:
		d.NameInput.Focus()
	case 2:
		d.EmailInput.Focus()
	default:
		if d.isCreateMode() && d.focusIndex == 3 {
			d.PasswordInput.Focus()
		}
	}
}

// Render renders the user form dialog
func (d UserFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - 6

	var content strings.Builder
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Username field
	content.WriteString(d.labelStyle.Render("Username"))
	content.WriteString("\n")
	if d.focusIndex == 0 {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.UsernameInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.UsernameInput.View()))
	}
	content.WriteString("\n")

	// Name field
	content.WriteString(d.labelStyle.Render("Name"))
	content.WriteString("\n")
	if d.focusIndex == 1 {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.NameInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.NameInput.View()))
	}
	content.WriteString("\n")

	// Email field
	content.WriteString(d.labelStyle.Render("Email"))
	content.WriteString("\n")
	if d.focusIndex == 2 {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.EmailInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.EmailInput.View()))
	}
	content.WriteString("\n")

	// Password field (create mode only)
	if d.isCreateMode() {
		content.WriteString(d.labelStyle.Render("Password"))
		content.WriteString("\n")
		if d.focusIndex == 3 {
			content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.PasswordInput.View()))
		} else {
			content.WriteString(d.inputStyle.Width(innerW).Render(d.PasswordInput.View()))
		}
		content.WriteString("\n")
	}

	// Role carousel
	content.WriteString(d.labelStyle.Render("Role"))
	content.WriteString("\n")
	if len(d.RoleOptions) > 0 {
		optLabel := d.RoleOptions[d.RoleIndex].Label
		if d.focusIndex == d.roleFocusIndex() {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Primary).
				Background(config.DefaultStyle.Accent).
				Padding(0, 1).
				Render("◀ " + optLabel + " ▶")
			content.WriteString(selector)
		} else {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Secondary).
				Padding(0, 1).
				Render("  " + optLabel + "  ")
			content.WriteString(selector)
		}
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render("No roles available"))
	}
	content.WriteString("\n\n")

	// Buttons
	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == d.cancelFocusIndex() {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == d.confirmFocusIndex() {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render("Cancel")
	confirmBtn := confirmStyle.Render("Save")
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)
	content.WriteString(buttonBar)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// UserFormDialogAcceptMsg is sent when the user form dialog is confirmed.
type UserFormDialogAcceptMsg struct {
	Action   FormDialogAction
	EntityID string
	Username string
	Name     string
	Email    string
	Password string
	Role     string
}

// UserFormDialogCancelMsg is sent when the user form dialog is cancelled.
type UserFormDialogCancelMsg struct{}

// ShowUserFormDialogMsg triggers showing a user form dialog.
type ShowUserFormDialogMsg struct {
	Title string
	Roles []db.Roles
}

// ShowEditUserDialogMsg triggers showing an edit user dialog.
type ShowEditUserDialogMsg struct {
	User  db.UserWithRoleLabelRow
	Roles []db.Roles
}

// UserFormDialogSetCmd creates a command to set the user form dialog model.
func UserFormDialogSetCmd(dialog *UserFormDialogModel) tea.Cmd {
	return func() tea.Msg { return UserFormDialogSetMsg{Dialog: dialog} }
}

// UserFormDialogActiveSetCmd creates a command to set the user form dialog active state.
func UserFormDialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg { return UserFormDialogActiveSetMsg{Active: active} }
}

// UserFormDialogSetMsg carries the dialog model to update.
type UserFormDialogSetMsg struct {
	Dialog *UserFormDialogModel
}

// UserFormDialogActiveSetMsg carries the active state for a user form dialog.
type UserFormDialogActiveSetMsg struct {
	Active bool
}

// UserFormDialogOverlay positions a user form dialog over existing content.
func UserFormDialogOverlay(content string, dialog UserFormDialogModel, width, height int) string {
	dialogContent := dialog.Render(width, height)
	dialogW := lipgloss.Width(dialogContent)
	dialogH := lipgloss.Height(dialogContent)

	x := (width - dialogW) / 2
	y := (height - dialogH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return tui.Composite(content, tui.Overlay{
		Content: dialogContent,
		X:       x,
		Y:       y,
		Width:   dialogW,
		Height:  dialogH,
	})
}

// =============================================================================
// SINGLE CONTENT FIELD EDIT DIALOG
// =============================================================================

// ShowEditSingleFieldDialogMsg triggers showing a single-field edit dialog.
type ShowEditSingleFieldDialogMsg struct {
	Field      ContentFieldDisplay
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

// ShowEditSingleFieldDialogCmd creates a command to show a single-field edit dialog.
func ShowEditSingleFieldDialogCmd(cf ContentFieldDisplay, contentID types.ContentID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowEditSingleFieldDialogMsg{
			Field:      cf,
			ContentID:  contentID,
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// EditSingleFieldAcceptMsg carries acceptance data from a single-field edit dialog.
type EditSingleFieldAcceptMsg struct {
	ContentFieldID types.ContentFieldID
	ContentID      types.ContentID
	FieldID        types.FieldID
	NewValue       string
	RouteID        types.RouteID
	DatatypeID     types.NullableDatatypeID
}

// =============================================================================
// ADD CONTENT FIELD DIALOG (picker for multiple missing fields)
// =============================================================================

// ShowAddContentFieldDialogMsg triggers showing an add-field picker dialog.
type ShowAddContentFieldDialogMsg struct {
	Options    []huh.Option[string]
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

// ShowAddContentFieldDialogCmd creates a command to show an add-field picker.
func ShowAddContentFieldDialogCmd(options []huh.Option[string], contentID types.ContentID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowAddContentFieldDialogMsg{
			Options:    options,
			ContentID:  contentID,
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// =============================================================================
// DELETE CONTENT FIELD DIALOG
// =============================================================================

// DeleteContentFieldContext stores context for a content field deletion operation.
type DeleteContentFieldContext struct {
	ContentFieldID types.ContentFieldID
	ContentID      types.ContentID
	RouteID        types.RouteID
	DatatypeID     types.NullableDatatypeID
}

// ShowDeleteContentFieldDialogMsg triggers showing a delete content field confirmation dialog.
type ShowDeleteContentFieldDialogMsg struct {
	Field      ContentFieldDisplay
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

// ShowDeleteContentFieldDialogCmd creates a command to show a delete content field dialog.
func ShowDeleteContentFieldDialogCmd(cf ContentFieldDisplay, contentID types.ContentID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteContentFieldDialogMsg{
			Field:      cf,
			ContentID:  contentID,
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// =============================================================================
// EXTERNAL EDITOR SUPPORT
// =============================================================================

// EditorFinishedMsg is sent after $EDITOR exits with the edited content.
type EditorFinishedMsg struct {
	FieldIndex int
	Content    string
	Err        error
}

// editorFileExtension returns a file extension hint for the given widget type.
func editorFileExtension(widget string) string {
	switch widget {
	case "markdown":
		return ".md"
	case "json-editor":
		return ".json"
	case "code-editor":
		return ".txt"
	case "rich-text":
		return ".html"
	default:
		return ".txt"
	}
}

// sshExecCommand wraps an exec.Cmd for use with tea.Exec, routing stderr to
// the same writer as stdout. This is necessary for SSH sessions where
// Bubbletea's default exec hardcodes stderr to os.Stderr (the server's
// stderr), which disconnects interactive editors like nvim from the SSH
// client's terminal.
type sshExecCommand struct {
	*exec.Cmd
}

func (c *sshExecCommand) SetStdin(r io.Reader) {
	if c.Stdin == nil {
		c.Stdin = r
	}
}

func (c *sshExecCommand) SetStdout(w io.Writer) {
	if c.Stdout == nil {
		c.Stdout = w
	}
	// Also route stderr to the SSH session writer so the editor's
	// full TUI output reaches the client.
	if c.Stderr == nil {
		c.Stderr = w
	}
}

func (c *sshExecCommand) SetStderr(w io.Writer) {
	// No-op: stderr was already wired to the SSH session in SetStdout.
	// This prevents Bubbletea's default of os.Stderr from overriding it.
}

// prepareEditorCmd synchronously creates a temp file with the current value,
// then returns a tea.Exec cmd that launches $EDITOR with stdin, stdout, and
// stderr all routed through the SSH session's PTY. When the editor exits,
// the callback reads the file and returns an EditorFinishedMsg.
// Returns nil if temp file creation fails. Logger may be nil.
func prepareEditorCmd(fieldIndex int, currentValue string, widget string, logger Logger) tea.Cmd {
	ext := editorFileExtension(widget)
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] file extension for widget %q: %s", widget, ext))
	}

	tmpFile, err := os.CreateTemp("", "modulacms-*"+ext)
	if err != nil {
		if logger != nil {
			logger.Ferror(fmt.Sprintf("[editor] failed to create temp file with extension %s", ext), err)
		}
		return nil
	}
	tmpPath := tmpFile.Name()
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] created temp file: %s", tmpPath))
	}

	if _, writeErr := tmpFile.WriteString(currentValue); writeErr != nil {
		if logger != nil {
			logger.Ferror(fmt.Sprintf("[editor] failed to write current value (%d bytes) to temp file %s", len(currentValue), tmpPath), writeErr)
		}
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil
	}
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] wrote %d bytes to temp file %s", len(currentValue), tmpPath))
	}

	if closeErr := tmpFile.Close(); closeErr != nil {
		if logger != nil {
			logger.Ferror(fmt.Sprintf("[editor] failed to close temp file %s", tmpPath), closeErr)
		}
		os.Remove(tmpPath)
		return nil
	}

	editor := editorCommand()
	editorParts := strings.Fields(editor)
	if len(editorParts) == 0 {
		if logger != nil {
			logger.Ferror("[editor] resolved editor command is empty after splitting", fmt.Errorf("empty editor command: %q", editor))
		}
		os.Remove(tmpPath)
		return nil
	}
	editorArgs := append(editorParts[1:], tmpPath)
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] resolved editor command: %q (binary: %q, args: %v)", editor, editorParts[0], editorArgs))
		logger.Finfo(fmt.Sprintf("[editor] launching (via sshExecCommand): %s %s", editorParts[0], strings.Join(editorArgs, " ")))
	}
	c := exec.Command(editorParts[0], editorArgs...)

	return tea.Exec(&sshExecCommand{Cmd: c}, func(procErr error) tea.Msg {
		defer os.Remove(tmpPath)
		if procErr != nil {
			if logger != nil {
				logger.Ferror(fmt.Sprintf("[editor] editor process exited with error for field %d, temp file %s", fieldIndex, tmpPath), procErr)
			}
			return EditorFinishedMsg{FieldIndex: fieldIndex, Err: procErr}
		}
		if logger != nil {
			logger.Finfo(fmt.Sprintf("[editor] editor process exited successfully for field %d, reading back temp file %s", fieldIndex, tmpPath))
		}
		data, readErr := os.ReadFile(tmpPath)
		if readErr != nil {
			if logger != nil {
				logger.Ferror(fmt.Sprintf("[editor] failed to read temp file %s after editor exit", tmpPath), readErr)
			}
			return EditorFinishedMsg{FieldIndex: fieldIndex, Err: readErr}
		}
		if logger != nil {
			logger.Finfo(fmt.Sprintf("[editor] read %d bytes from temp file %s, returning EditorFinishedMsg for field %d", len(data), tmpPath, fieldIndex))
		}
		return EditorFinishedMsg{FieldIndex: fieldIndex, Content: string(data)}
	})
}
