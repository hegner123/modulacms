package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// FormDialogAction identifies the type of form dialog
type FormDialogAction string

// FormDialogAction constants identify dialog types for operations across entities.
const (
	FORMDIALOGCREATEDATATYPE              FormDialogAction = "create_datatype"
	FORMDIALOGEDITDATATYPE                FormDialogAction = "edit_datatype"
	FORMDIALOGCREATEFIELD                 FormDialogAction = "create_field"
	FORMDIALOGEDITFIELD                   FormDialogAction = "edit_field"
	FORMDIALOGCREATEROUTE                 FormDialogAction = "create_route"
	FORMDIALOGEDITROUTE                   FormDialogAction = "edit_route"
	FORMDIALOGCREATEROUTEWITHCONTENT      FormDialogAction = "create_route_with_content"
	FORMDIALOGINITIALIZEROUTECONTENT      FormDialogAction = "initialize_route_content"
	FORMDIALOGCHILDDATATYPE               FormDialogAction = "child_datatype"
	FORMDIALOGCREATECONTENT               FormDialogAction = "create_content"
	FORMDIALOGEDITCONTENT                 FormDialogAction = "edit_content"
	FORMDIALOGMOVECONTENT                 FormDialogAction = "move_content"
	FORMDIALOGCREATEUSER                  FormDialogAction = "create_user"
	FORMDIALOGEDITUSER                    FormDialogAction = "edit_user"
	FORMDIALOGEDIITSINGLEFIELD            FormDialogAction = "edit_single_field"
	FORMDIALOGADDCONTENTFIELD             FormDialogAction = "add_content_field"
	FORMDIALOGCREATEADMINROUTE            FormDialogAction = "create_admin_route"
	FORMDIALOGEDITADMINROUTE              FormDialogAction = "edit_admin_route"
	FORMDIALOGCREATEADMINDATATYPE         FormDialogAction = "create_admin_datatype"
	FORMDIALOGEDITADMINDATATYPE           FormDialogAction = "edit_admin_datatype"
	FORMDIALOGCREATEADMINFIELD            FormDialogAction = "create_admin_field"
	FORMDIALOGEDITADMINFIELD              FormDialogAction = "edit_admin_field"
	FORMDIALOGDBINSERT                    FormDialogAction = "db_insert"
	FORMDIALOGDBUPDATE                    FormDialogAction = "db_update"
	FORMDIALOGCONFIGEDIT                  FormDialogAction = "config_edit"
	FORMDIALOGCREATEFIELDTYPE             FormDialogAction = "create_field_type"
	FORMDIALOGEDITFIELDTYPE               FormDialogAction = "edit_field_type"
	FORMDIALOGCREATEADMINFIELDTYPE        FormDialogAction = "create_admin_field_type"
	FORMDIALOGEDITADMINFIELDTYPE          FormDialogAction = "edit_admin_field_type"
	FORMDIALOGCREATEADMINROUTEWITHCONTENT FormDialogAction = "create_admin_route_with_content"
	FORMDIALOGCREATEADMINCONTENT          FormDialogAction = "create_admin_content"
	FORMDIALOGEDITADMINCONTENT            FormDialogAction = "edit_admin_content"
	FORMDIALOGMOVEADMINCONTENT            FormDialogAction = "move_admin_content"
	FORMDIALOGCHILDADMINDATATYPE          FormDialogAction = "child_admin_datatype"
	FORMDIALOGADDADMINCONTENTFIELD        FormDialogAction = "add_admin_content_field"
	FORMDIALOGEDITADMINSINGLEFIELD        FormDialogAction = "edit_admin_single_field"
)

// FormDialogField constants define focus indices for dialog fields.
const (
	FormDialogFieldName = iota
	FormDialogFieldLabel
	FormDialogFieldType
	FormDialogFieldParent
	FormDialogButtonCancel
	FormDialogButtonConfirm
)

// ParentOption represents a selectable parent datatype
type ParentOption struct {
	Label string
	Value string // DatatypeID or empty for _root
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

// buttonLabel prefixes text with a cursor indicator when focused,
// ensuring buttons are distinguishable on monochrome terminals.
func buttonLabel(text string, focused bool) string {
	if focused {
		return "\u25b8 " + text
	}
	return "  " + text
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
	NameInput  textinput.Model
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
	// Create name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Machine name"
	nameInput.CharLimit = 64
	nameInput.SetWidth(40)
	nameInput.Focus()

	// Create label input
	labelInput := textinput.New()
	labelInput.Placeholder = "Display name"
	labelInput.CharLimit = 64
	labelInput.SetWidth(40)

	// Create type input
	typeInput := textinput.New()
	typeInput.Placeholder = "_root"
	typeInput.CharLimit = 32
	typeInput.SetWidth(40)

	// Build parent options
	parentOptions := []ParentOption{
		{Label: "_root (no parent)", Value: ""},
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
		NameInput:     nameInput,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parentOptions,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldName,
	}
}

// NewFieldFormDialog creates a form dialog for field creation (no parent selector)
func NewFieldFormDialog(title string, action FormDialogAction) FormDialogModel {
	// Create name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Machine name"
	nameInput.CharLimit = 64
	nameInput.SetWidth(40)
	nameInput.Focus()

	// Create label input
	labelInput := textinput.New()
	labelInput.Placeholder = "Display name"
	labelInput.CharLimit = 64
	labelInput.SetWidth(40)

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		NameInput:    nameInput,
		LabelInput:   labelInput,
		TypeOptions:  TypeOptionsFromRegistry(),
		TypeIndex:    0,
		focusIndex:   FormDialogFieldName,
	}
}

// NewRouteFormDialog creates a form dialog for route creation (Title and Slug inputs)
func NewRouteFormDialog(title string, action FormDialogAction) FormDialogModel {
	// Create title input (uses LabelInput field)
	titleInput := textinput.New()
	titleInput.Placeholder = "Page title"
	titleInput.CharLimit = 128
	titleInput.SetWidth(40)
	titleInput.Focus()

	// Create slug input (uses TypeInput field)
	slugInput := textinput.New()
	slugInput.Placeholder = "url-slug"
	slugInput.CharLimit = 128
	slugInput.SetWidth(40)

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		LabelInput:   titleInput,
		TypeInput:    slugInput,
		focusIndex:   FormDialogFieldLabel, // Skip name field for routes
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

// HasNameField returns true if the dialog should render the Name input field.
// Datatype and field dialogs include Name; route, config, and other dialogs do not.
func (d *FormDialogModel) HasNameField() bool {
	switch d.Action {
	case FORMDIALOGCREATEDATATYPE, FORMDIALOGEDITDATATYPE,
		FORMDIALOGCREATEFIELD, FORMDIALOGEDITFIELD,
		FORMDIALOGCREATEADMINDATATYPE, FORMDIALOGEDITADMINDATATYPE,
		FORMDIALOGCREATEADMINFIELD, FORMDIALOGEDITADMINFIELD:
		return true
	}
	return false
}

// HasSecondField returns true if the dialog should render the second (Type/Slug) field.
func (d *FormDialogModel) HasSecondField() bool {
	return d.Action != FORMDIALOGCONFIGEDIT
}

// Update handles input for the form dialog
func (d *FormDialogModel) Update(msg tea.Msg) (FormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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
						Name:     d.NameInput.Value(),
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
		case FormDialogFieldName:
			d.NameInput, cmd = d.NameInput.Update(msg)
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
func (d *FormDialogModel) updateChildDatatypeSelection(msg tea.KeyPressMsg) (FormDialogModel, tea.Cmd) {
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
	// Skip name field if not applicable
	if d.focusIndex == FormDialogFieldName && !d.HasNameField() {
		d.focusIndex = FormDialogFieldLabel
	}
	// Skip type field if not applicable
	if d.focusIndex == FormDialogFieldType && !d.HasSecondField() {
		d.focusIndex = FormDialogButtonCancel
	}
	// Skip parent field if no parent options
	if d.focusIndex == FormDialogFieldParent && !d.HasParentSelector() {
		d.focusIndex = FormDialogButtonCancel
	}
	if d.focusIndex > FormDialogButtonConfirm {
		if d.HasNameField() {
			d.focusIndex = FormDialogFieldName
		} else {
			d.focusIndex = FormDialogFieldLabel
		}
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
	// Skip name field if not applicable
	if d.focusIndex == FormDialogFieldName && !d.HasNameField() {
		d.focusIndex = FormDialogButtonConfirm
	}
	if d.focusIndex < FormDialogFieldName {
		d.focusIndex = FormDialogButtonConfirm
	}
	d.updateFocus()
}

// updateFocus applies focus styling to the currently focused element.
func (d *FormDialogModel) updateFocus() {
	d.NameInput.Blur()
	d.LabelInput.Blur()
	if !d.HasTypeSelector() {
		d.TypeInput.Blur()
	}

	switch d.focusIndex {
	case FormDialogFieldName:
		d.NameInput.Focus()
	case FormDialogFieldLabel:
		d.LabelInput.Focus()
	case FormDialogFieldType:
		if !d.HasTypeSelector() {
			d.TypeInput.Focus()
		}
	}
}

// OverlayUpdate implements ModalOverlay for FormDialogModel.
func (d *FormDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayView implements ModalOverlay for FormDialogModel.
func (d *FormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the form dialog
func (d FormDialogModel) Render(windowWidth, windowHeight int) string {
	var content strings.Builder
	// Inner width available for fields (dialog width minus border and padding)
	innerW := d.Width - dialogBorderPadding

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
		d.Action == FORMDIALOGCREATEROUTEWITHCONTENT ||
		d.Action == FORMDIALOGCREATEADMINROUTEWITHCONTENT {
		firstFieldLabel = "Title"
		secondFieldLabel = "Slug"
	}
	if d.Action == FORMDIALOGCONFIGEDIT {
		firstFieldLabel = "Value"
	}

	// Name field (only for datatype and field dialogs)
	if d.HasNameField() {
		nameLabel := d.labelStyle.Render("Name")
		content.WriteString(nameLabel)
		content.WriteString("\n")
		if d.focusIndex == FormDialogFieldName {
			content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.NameInput.View()))
		} else {
			content.WriteString(d.inputStyle.Width(innerW).Render(d.NameInput.View()))
		}
		content.WriteString("\n")
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
		if d.Action == FORMDIALOGCREATEROUTEWITHCONTENT || d.Action == FORMDIALOGCREATEADMINROUTEWITHCONTENT {
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
	focused := d.focusIndex == FormDialogButtonCancel
	style := d.cancelButtonStyle
	if focused {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Tertiary).
			Bold(true)
	}
	return style.Render(buttonLabel("Cancel", focused))
}

// renderConfirmButton returns the styled confirm button view.
func (d FormDialogModel) renderConfirmButton() string {
	focused := d.focusIndex == FormDialogButtonConfirm
	style := d.confirmButtonStyle
	if focused {
		style = style.
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Accent).
			Bold(true)
	}
	return style.Render(buttonLabel("Confirm", focused))
}

// FormDialogAcceptMsg carries form dialog acceptance data.
type FormDialogAcceptMsg struct {
	Action   FormDialogAction
	EntityID string // ID of entity being edited (empty for create)
	Name     string
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

// ShowFieldFormDialogCmd shows a field creation dialog (no parent selector).
// contextID is the parent datatype ID that the new field will be linked to.
func ShowFieldFormDialogCmd(action FormDialogAction, title string, contextID string) tea.Cmd {
	return func() tea.Msg {
		return ShowFieldFormDialogMsg{
			Action:    action,
			Title:     title,
			ContextID: contextID,
		}
	}
}

// ShowFieldFormDialogMsg is the message for showing a field form dialog.
// ContextID carries the parent datatype ID for field creation.
type ShowFieldFormDialogMsg struct {
	Action    FormDialogAction
	Title     string
	ContextID string // parent datatype ID (stored as EntityID on the dialog model)
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
