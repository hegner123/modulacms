package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tui"
)

// FormDialogAction identifies the type of form dialog
type FormDialogAction string

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
)

// FormDialogField indices for focus navigation
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

// FormDialogModel represents a form dialog with text inputs and buttons
type FormDialogModel struct {
	Title  string
	Width  int
	Action FormDialogAction

	// EntityID is the ID of the entity being edited (empty for create operations)
	EntityID string

	// Text input fields
	LabelInput textinput.Model
	TypeInput  textinput.Model

	// Parent selection
	ParentOptions []ParentOption
	ParentIndex   int

	// Focus management: 0=Label, 1=Type, 2=Parent, 3=Cancel, 4=Confirm
	focusIndex int

	// Styles
	borderStyle        lipgloss.Style
	titleStyle         lipgloss.Style
	labelStyle         lipgloss.Style
	inputStyle         lipgloss.Style
	focusedInputStyle  lipgloss.Style
	buttonStyle        lipgloss.Style
	cancelButtonStyle  lipgloss.Style
	confirmButtonStyle lipgloss.Style
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
		Title:         title,
		Width:         60,
		Action:        action,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parentOptions,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
		borderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent).MarginBottom(1),
		labelStyle:    lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).MarginBottom(0),
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

// NewFieldFormDialog creates a form dialog for field creation (no parent selector)
func NewFieldFormDialog(title string, action FormDialogAction) FormDialogModel {
	// Create label input
	labelInput := textinput.New()
	labelInput.Placeholder = "Field name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.Focus()

	// Create type input
	typeInput := textinput.New()
	typeInput.Placeholder = "text"
	typeInput.CharLimit = 32
	typeInput.Width = 40

	return FormDialogModel{
		Title:         title,
		Width:         60,
		Action:        action,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: nil, // No parent selection for fields
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
		borderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent).MarginBottom(1),
		labelStyle:    lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).MarginBottom(0),
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
		Title:         title,
		Width:         60,
		Action:        action,
		LabelInput:    titleInput,
		TypeInput:     slugInput,
		ParentOptions: nil, // No parent selection for routes
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
		borderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent).MarginBottom(1),
		labelStyle:    lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).MarginBottom(0),
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

// HasParentSelector returns true if the dialog has a parent selector
func (d *FormDialogModel) HasParentSelector() bool {
	return len(d.ParentOptions) > 0
}

// Update handles input for the form dialog
func (d *FormDialogModel) Update(msg tea.Msg) (FormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Special handling for child datatype selection - simple vertical list
		if d.Action == FORMDIALOGCHILDDATATYPE {
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

				return *d, func() tea.Msg {
					return FormDialogAcceptMsg{
						Action:   d.Action,
						EntityID: d.EntityID,
						Label:    d.LabelInput.Value(),
						Type:     d.TypeInput.Value(),
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
			d.TypeInput, cmd = d.TypeInput.Update(msg)
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

func (d *FormDialogModel) focusNext() {
	d.focusIndex++
	// Skip parent field if no parent options
	if d.focusIndex == FormDialogFieldParent && !d.HasParentSelector() {
		d.focusIndex = FormDialogButtonCancel
	}
	if d.focusIndex > FormDialogButtonConfirm {
		d.focusIndex = FormDialogFieldLabel
	}
	d.updateFocus()
}

func (d *FormDialogModel) focusPrev() {
	d.focusIndex--
	// Skip parent field if no parent options
	if d.focusIndex == FormDialogFieldParent && !d.HasParentSelector() {
		d.focusIndex = FormDialogFieldType
	}
	if d.focusIndex < FormDialogFieldLabel {
		d.focusIndex = FormDialogButtonConfirm
	}
	d.updateFocus()
}

func (d *FormDialogModel) updateFocus() {
	d.LabelInput.Blur()
	d.TypeInput.Blur()

	switch d.focusIndex {
	case FormDialogFieldLabel:
		d.LabelInput.Focus()
	case FormDialogFieldType:
		d.TypeInput.Focus()
	}
}

// Render renders the form dialog
func (d FormDialogModel) Render(windowWidth, windowHeight int) string {
	var content strings.Builder

	// Title
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Special rendering for child datatype selection - vertical list
	if d.Action == FORMDIALOGCHILDDATATYPE {
		return d.renderChildDatatypeSelection(windowWidth, windowHeight)
	}

	// Determine field labels based on action type
	firstFieldLabel := "Label"
	secondFieldLabel := "Type"
	if d.Action == FORMDIALOGCREATEROUTE || d.Action == FORMDIALOGEDITROUTE {
		firstFieldLabel = "Title"
		secondFieldLabel = "Slug"
	}

	// First field (Label or Title)
	labelLabel := d.labelStyle.Render(firstFieldLabel)
	content.WriteString(labelLabel)
	content.WriteString("\n")
	if d.focusIndex == FormDialogFieldLabel {
		content.WriteString(d.focusedInputStyle.Render(d.LabelInput.View()))
	} else {
		content.WriteString(d.inputStyle.Render(d.LabelInput.View()))
	}
	content.WriteString("\n")

	// Second field (Type or Slug)
	typeLabel := d.labelStyle.Render(secondFieldLabel)
	content.WriteString(typeLabel)
	content.WriteString("\n")
	if d.focusIndex == FormDialogFieldType {
		content.WriteString(d.focusedInputStyle.Render(d.TypeInput.View()))
	} else {
		content.WriteString(d.inputStyle.Render(d.TypeInput.View()))
	}
	content.WriteString("\n")

	// Parent selector (only for dialogs with parent options)
	if d.HasParentSelector() {
		parentLabel := d.labelStyle.Render("Parent")
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

// Form dialog messages
type FormDialogAcceptMsg struct {
	Action   FormDialogAction
	EntityID string // ID of entity being edited (empty for create)
	Label    string
	Type     string
	ParentID string
}

type FormDialogCancelMsg struct{}

type ShowFormDialogMsg struct {
	Action  FormDialogAction
	Title   string
	Parents []db.Datatypes
}

// Commands
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

type FormDialogSetMsg struct {
	Dialog *FormDialogModel
}

type FormDialogActiveSetMsg struct {
	Active bool
}

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
		Title:         title,
		Width:         60,
		Action:        action,
		EntityID:      string(datatype.DatatypeID),
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parentOptions,
		ParentIndex:   selectedParentIndex,
		focusIndex:    FormDialogFieldLabel,
		borderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent).MarginBottom(1),
		labelStyle:    lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).MarginBottom(0),
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

// NewEditFieldDialog creates a form dialog for editing a field with pre-populated values
func NewEditFieldDialog(title string, action FormDialogAction, field db.Fields) FormDialogModel {
	// Create label input with current value
	labelInput := textinput.New()
	labelInput.Placeholder = "Field name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.SetValue(field.Label)
	labelInput.Focus()

	// Create type input with current value
	typeInput := textinput.New()
	typeInput.Placeholder = "text"
	typeInput.CharLimit = 32
	typeInput.Width = 40
	typeInput.SetValue(string(field.Type))

	return FormDialogModel{
		Title:         title,
		Width:         60,
		Action:        action,
		EntityID:      string(field.FieldID),
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: nil, // No parent selection for fields
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
		borderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent).MarginBottom(1),
		labelStyle:    lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).MarginBottom(0),
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
		Title:         title,
		Width:         60,
		Action:        action,
		EntityID:      string(route.RouteID),
		LabelInput:    titleInput,
		TypeInput:     slugInput,
		ParentOptions: nil, // No parent selection for routes
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
		borderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent).MarginBottom(1),
		labelStyle:    lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).MarginBottom(0),
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
func NewRouteWithContentDialog(title string, action FormDialogAction, datatypeID string) FormDialogModel {
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
		Title:         title,
		Width:         60,
		Action:        action,
		EntityID:      datatypeID, // Store datatypeID in EntityID for route+content creation
		LabelInput:    titleInput,
		TypeInput:     slugInput,
		ParentOptions: nil,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
		borderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent).MarginBottom(1),
		labelStyle:    lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).MarginBottom(0),
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

// ShowCreateRouteWithContentDialogMsg is the message for showing a create route with content dialog
type ShowCreateRouteWithContentDialogMsg struct {
	DatatypeID string
}

// ShowCreateRouteWithContentDialogCmd shows a dialog to create a new route with initial content
func ShowCreateRouteWithContentDialogCmd(datatypeID string) tea.Cmd {
	return func() tea.Msg {
		return ShowCreateRouteWithContentDialogMsg{
			DatatypeID: datatypeID,
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
		Title:         title,
		Width:         50,
		Action:        FORMDIALOGCHILDDATATYPE,
		EntityID:      routeID,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parents,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldParent, // Start on selection
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

// =============================================================================
// ContentFormDialogModel - Dynamic content fields dialog
// =============================================================================

// ContentFieldInput represents a single field input in the content form
type ContentFieldInput struct {
	FieldID types.FieldID
	Label   string
	Type    string // field type (text, textarea, number, etc.)
	Input   textinput.Model
}

// ContentFormDialogModel represents a form dialog with dynamic content fields
type ContentFormDialogModel struct {
	Title      string
	Width      int
	Action     FormDialogAction
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	ContentID  types.ContentID         // For edit mode, empty for create
	ParentID   types.NullableContentID // Parent content for child creation

	// Dynamic field inputs
	Fields []ContentFieldInput

	// Focus management: 0 to len(Fields)-1 for fields, then Cancel, then Confirm
	focusIndex int

	// Styles
	borderStyle        lipgloss.Style
	titleStyle         lipgloss.Style
	labelStyle         lipgloss.Style
	inputStyle         lipgloss.Style
	focusedInputStyle  lipgloss.Style
	buttonStyle        lipgloss.Style
	cancelButtonStyle  lipgloss.Style
	confirmButtonStyle lipgloss.Style
}

// NewContentFormDialog creates a new content form dialog with dynamic fields
func NewContentFormDialog(title string, action FormDialogAction, datatypeID types.DatatypeID, routeID types.RouteID, fields []db.Fields) ContentFormDialogModel {
	contentFields := make([]ContentFieldInput, 0, len(fields))
	for _, f := range fields {
		input := textinput.New()
		input.Placeholder = f.Label
		input.CharLimit = 256
		input.Width = 50
		contentFields = append(contentFields, ContentFieldInput{
			FieldID: f.FieldID,
			Label:   f.Label,
			Type:    string(f.Type),
			Input:   input,
		})
	}
	// Focus first field after all inputs are created
	if len(contentFields) > 0 {
		contentFields[0].Input.Focus()
	}

	return ContentFormDialogModel{
		Title:      title,
		Width:      65,
		Action:     action,
		DatatypeID: datatypeID,
		RouteID:    routeID,
		Fields:     contentFields,
		focusIndex: 0,
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
					fieldValues[f.FieldID] = f.Input.Value()
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
		}

		// Update the focused text input
		if d.focusIndex < len(d.Fields) {
			var cmd tea.Cmd
			d.Fields[d.focusIndex].Input, cmd = d.Fields[d.focusIndex].Input.Update(msg)
			return *d, cmd
		}
	}

	return *d, nil
}

func (d *ContentFormDialogModel) focusNext() {
	d.focusIndex++
	if d.focusIndex > d.ButtonConfirmIndex() {
		d.focusIndex = 0
	}
	d.updateFocus()
}

func (d *ContentFormDialogModel) focusPrev() {
	d.focusIndex--
	if d.focusIndex < 0 {
		d.focusIndex = d.ButtonConfirmIndex()
	}
	d.updateFocus()
}

func (d *ContentFormDialogModel) updateFocus() {
	// Blur all fields
	for i := range d.Fields {
		d.Fields[i].Input.Blur()
	}
	// Focus the current field if it's a text input
	if d.focusIndex < len(d.Fields) {
		d.Fields[d.focusIndex].Input.Focus()
	}
}

// Render renders the content form dialog
func (d ContentFormDialogModel) Render(windowWidth, windowHeight int) string {
	var content strings.Builder

	// Title
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Render each field
	for i, f := range d.Fields {
		label := d.labelStyle.Render(f.Label)
		content.WriteString(label)
		content.WriteString("\n")

		if d.focusIndex == i {
			content.WriteString(d.focusedInputStyle.Render(f.Input.View()))
		} else {
			content.WriteString(d.inputStyle.Render(f.Input.View()))
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

// Content form dialog messages
type ContentFormDialogAcceptMsg struct {
	Action      FormDialogAction
	DatatypeID  types.DatatypeID
	RouteID     types.RouteID
	ContentID   types.ContentID         // For edit mode
	ParentID    types.NullableContentID // For child creation
	FieldValues map[types.FieldID]string
}

type ContentFormDialogCancelMsg struct{}

type ShowContentFormDialogMsg struct {
	Action     FormDialogAction
	Title      string
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	ParentID   types.NullableContentID
	Fields     []db.Fields
}

// Commands
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

func ContentFormDialogSetCmd(dialog *ContentFormDialogModel) tea.Cmd {
	return func() tea.Msg {
		return ContentFormDialogSetMsg{Dialog: dialog}
	}
}

type ContentFormDialogSetMsg struct {
	Dialog *ContentFormDialogModel
}

type ContentFormDialogActiveSetMsg struct {
	Active bool
}

func ContentFormDialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg {
		return ContentFormDialogActiveSetMsg{Active: active}
	}
}

// CreateContentFromDialogRequestMsg is sent to create content from the dialog
type CreateContentFromDialogRequestMsg struct {
	DatatypeID  types.DatatypeID
	RouteID     types.RouteID
	ParentID    types.NullableContentID
	FieldValues map[types.FieldID]string
}

// CreateContentFromDialogCmd creates a command to create content from dialog values
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

// ContentCreatedFromDialogMsg is sent after content is successfully created
type ContentCreatedFromDialogMsg struct {
	ContentID  types.ContentID
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	FieldCount int
}

// FetchContentFieldsMsg triggers fetching fields for a datatype to show the content form
type FetchContentFieldsMsg struct {
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	ParentID   types.NullableContentID
	Title      string
}

// FetchContentFieldsCmd creates a command to fetch fields for content form
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

// ExistingContentField represents a field with its current value for editing
type ExistingContentField struct {
	ContentFieldID types.ContentFieldID
	FieldID        types.FieldID
	Label          string
	Type           string
	Value          string
}

// NewEditContentFormDialog creates a content form dialog pre-populated with existing values
func NewEditContentFormDialog(title string, contentID types.ContentID, datatypeID types.DatatypeID, routeID types.RouteID, existingFields []ExistingContentField) ContentFormDialogModel {
	contentFields := make([]ContentFieldInput, 0, len(existingFields))
	for _, f := range existingFields {
		input := textinput.New()
		input.Placeholder = f.Label
		input.CharLimit = 256
		input.Width = 50
		input.SetValue(f.Value) // Pre-populate with existing value
		contentFields = append(contentFields, ContentFieldInput{
			FieldID: f.FieldID,
			Label:   f.Label,
			Type:    f.Type,
			Input:   input,
		})
	}
	// Focus first field after all inputs are created
	if len(contentFields) > 0 {
		contentFields[0].Input.Focus()
	}

	return ContentFormDialogModel{
		Title:      title,
		Width:      65,
		Action:     FORMDIALOGEDITCONTENT,
		DatatypeID: datatypeID,
		RouteID:    routeID,
		ContentID:  contentID, // Set the content ID for edit mode
		Fields:     contentFields,
		focusIndex: 0,
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

// ShowEditContentFormDialogMsg is the message for showing an edit content form dialog
type ShowEditContentFormDialogMsg struct {
	Title          string
	ContentID      types.ContentID
	DatatypeID     types.DatatypeID
	RouteID        types.RouteID
	ExistingFields []ExistingContentField
}

// ShowEditContentFormDialogCmd shows a content form dialog pre-populated for editing
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

// FetchContentForEditMsg triggers fetching content fields for editing
type FetchContentForEditMsg struct {
	ContentID  types.ContentID
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
	Title      string
}

// FetchContentForEditCmd creates a command to fetch content fields for editing
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

// UpdateContentFromDialogRequestMsg is sent to update content from the dialog
type UpdateContentFromDialogRequestMsg struct {
	ContentID   types.ContentID
	DatatypeID  types.DatatypeID
	RouteID     types.RouteID
	FieldValues map[types.FieldID]string
}

// UpdateContentFromDialogCmd creates a command to update content from dialog values
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

// ContentUpdatedFromDialogMsg is sent after content is successfully updated
type ContentUpdatedFromDialogMsg struct {
	ContentID    types.ContentID
	DatatypeID   types.DatatypeID
	RouteID      types.RouteID
	UpdatedCount int
}
