package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/tui"
)

// FormDialogAction identifies the type of form dialog
type FormDialogAction string

const (
	FORMDIALOGCREATEDATATYPE FormDialogAction = "create_datatype"
	FORMDIALOGCREATEFIELD    FormDialogAction = "create_field"
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

	// Text input fields
	LabelInput textinput.Model
	TypeInput  textinput.Model

	// Parent selection
	ParentOptions []ParentOption
	ParentIndex   int

	// Focus management: 0=Label, 1=Type, 2=Parent, 3=Cancel, 4=Confirm
	focusIndex int

	// Styles
	borderStyle       lipgloss.Style
	titleStyle        lipgloss.Style
	labelStyle        lipgloss.Style
	inputStyle        lipgloss.Style
	focusedInputStyle lipgloss.Style
	buttonStyle       lipgloss.Style
	cancelButtonStyle lipgloss.Style
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
		Width:         50,
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
		Width:         50,
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

// HasParentSelector returns true if the dialog has a parent selector
func (d *FormDialogModel) HasParentSelector() bool {
	return len(d.ParentOptions) > 0
}

// Update handles input for the form dialog
func (d *FormDialogModel) Update(msg tea.Msg) (FormDialogModel, tea.Cmd) {
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
				return *d, func() tea.Msg {
					return FormDialogAcceptMsg{
						Action:   d.Action,
						Label:    d.LabelInput.Value(),
						Type:     d.TypeInput.Value(),
						ParentID: d.ParentOptions[d.ParentIndex].Value,
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

	// Label field
	labelLabel := d.labelStyle.Render("Label")
	content.WriteString(labelLabel)
	content.WriteString("\n")
	if d.focusIndex == FormDialogFieldLabel {
		content.WriteString(d.focusedInputStyle.Render(d.LabelInput.View()))
	} else {
		content.WriteString(d.inputStyle.Render(d.LabelInput.View()))
	}
	content.WriteString("\n")

	// Type field
	typeLabel := d.labelStyle.Render("Type")
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
