package cli

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tui"
)

// WidgetOption represents a selectable widget override for a field's UI.
type WidgetOption struct {
	Label string // Display label
	Value string // Stored value (empty string = default/none)
}

// widgetOptions is the hardcoded list of available widget overrides.
var widgetOptions = []WidgetOption{
	{Label: "(default)", Value: ""},
	{Label: "Markdown", Value: "markdown"},
	{Label: "Rich Text", Value: "rich-text"},
	{Label: "Code Editor", Value: "code-editor"},
	{Label: "Color Picker", Value: "color-picker"},
	{Label: "Date Picker", Value: "date-picker"},
	{Label: "Date-Time Picker", Value: "datetime-picker"},
	{Label: "Time Picker", Value: "time-picker"},
	{Label: "Toggle", Value: "toggle"},
	{Label: "Slider", Value: "slider"},
	{Label: "Range", Value: "range"},
	{Label: "Radio", Value: "radio"},
	{Label: "Checkbox Group", Value: "checkbox-group"},
	{Label: "Tags", Value: "tags"},
	{Label: "Password", Value: "password"},
	{Label: "JSON Editor", Value: "json-editor"},
	{Label: "File Upload", Value: "file-upload"},
	{Label: "Image Upload", Value: "image-upload"},
	{Label: "Map", Value: "map"},
}

// widgetOptionIndex returns the index matching value, or 0 if not found.
func widgetOptionIndex(value string) int {
	for i, opt := range widgetOptions {
		if opt.Value == value {
			return i
		}
	}
	return 0
}

// UIConfigFormDialogModel represents a form dialog for editing field UI configuration.
type UIConfigFormDialogModel struct {
	dialogStyles

	Title   string
	Width   int
	FieldID string

	WidgetOptions []WidgetOption // available widgets
	WidgetIndex   int            // currently selected widget
	PlaceholderInput textinput.Model // placeholder text
	HelpTextInput    textinput.Model // help text
	HiddenToggle     bool            // boolean toggle

	// Focus: 0=widget, 1=placeholder, 2=helptext, 3=hidden, 4=cancel, 5=confirm
	focusIndex int
}

// NewUIConfigFormDialog creates a blank UIConfig form dialog.
func NewUIConfigFormDialog(title, fieldID string) UIConfigFormDialogModel {
	placeholder := textinput.New()
	placeholder.Placeholder = "Placeholder text shown to users"
	placeholder.CharLimit = 128
	placeholder.Width = 40

	helpText := textinput.New()
	helpText.Placeholder = "Help text for this field"
	helpText.CharLimit = 256
	helpText.Width = 40

	return UIConfigFormDialogModel{
		dialogStyles:     newDialogStyles(),
		Title:            title,
		Width:            60,
		FieldID:          fieldID,
		WidgetOptions:    widgetOptions,
		WidgetIndex:      0,
		PlaceholderInput: placeholder,
		HelpTextInput:    helpText,
		HiddenToggle:     false,
		focusIndex:       0,
	}
}

// NewEditUIConfigFormDialog creates a pre-populated UIConfig form dialog.
func NewEditUIConfigFormDialog(title, fieldID string, existing types.UIConfig) UIConfigFormDialogModel {
	placeholder := textinput.New()
	placeholder.Placeholder = "Placeholder text shown to users"
	placeholder.CharLimit = 128
	placeholder.Width = 40
	placeholder.SetValue(existing.Placeholder)

	helpText := textinput.New()
	helpText.Placeholder = "Help text for this field"
	helpText.CharLimit = 256
	helpText.Width = 40
	helpText.SetValue(existing.HelpText)

	return UIConfigFormDialogModel{
		dialogStyles:     newDialogStyles(),
		Title:            title,
		Width:            60,
		FieldID:          fieldID,
		WidgetOptions:    widgetOptions,
		WidgetIndex:      widgetOptionIndex(existing.Widget),
		PlaceholderInput: placeholder,
		HelpTextInput:    helpText,
		HiddenToggle:     existing.Hidden,
		focusIndex:       0,
	}
}

// Update handles input for the UIConfig form dialog.
func (d *UIConfigFormDialogModel) Update(msg tea.Msg) (UIConfigFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			d.uiConfigFocusNext()
			return *d, nil
		case "shift+tab", "up":
			d.uiConfigFocusPrev()
			return *d, nil
		case "enter":
			if d.focusIndex == 5 {
				// Confirm
				widget := ""
				if d.WidgetIndex < len(d.WidgetOptions) {
					widget = d.WidgetOptions[d.WidgetIndex].Value
				}
				return *d, func() tea.Msg {
					return UIConfigFormDialogAcceptMsg{
						FieldID:     d.FieldID,
						Widget:      widget,
						Placeholder: d.PlaceholderInput.Value(),
						HelpText:    d.HelpTextInput.Value(),
						Hidden:      d.HiddenToggle,
					}
				}
			}
			if d.focusIndex == 4 {
				// Cancel
				return *d, func() tea.Msg { return UIConfigFormDialogCancelMsg{} }
			}
			// On other fields, move to next
			d.uiConfigFocusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return UIConfigFormDialogCancelMsg{} }
		case "left":
			// Widget carousel
			if d.focusIndex == 0 {
				if d.WidgetIndex > 0 {
					d.WidgetIndex--
				}
				return *d, nil
			}
			// Hidden toggle
			if d.focusIndex == 3 {
				d.HiddenToggle = !d.HiddenToggle
				return *d, nil
			}
			// Buttons
			if d.focusIndex == 5 {
				d.focusIndex = 4
				return *d, nil
			}
		case "right":
			// Widget carousel
			if d.focusIndex == 0 {
				if d.WidgetIndex < len(d.WidgetOptions)-1 {
					d.WidgetIndex++
				}
				return *d, nil
			}
			// Hidden toggle
			if d.focusIndex == 3 {
				d.HiddenToggle = !d.HiddenToggle
				return *d, nil
			}
			// Buttons
			if d.focusIndex == 4 {
				d.focusIndex = 5
				return *d, nil
			}
		case " ":
			// Space toggles hidden
			if d.focusIndex == 3 {
				d.HiddenToggle = !d.HiddenToggle
				return *d, nil
			}
		}

		// Update the focused text input
		var cmd tea.Cmd
		switch d.focusIndex {
		case 1:
			d.PlaceholderInput, cmd = d.PlaceholderInput.Update(msg)
		case 2:
			d.HelpTextInput, cmd = d.HelpTextInput.Update(msg)
		}
		return *d, cmd
	}

	return *d, nil
}

// uiConfigFocusNext advances focus to the next element, wrapping at the end.
func (d *UIConfigFormDialogModel) uiConfigFocusNext() {
	d.focusIndex = (d.focusIndex + 1) % 6
	d.uiConfigUpdateFocus()
}

// uiConfigFocusPrev moves focus to the previous element, wrapping at the start.
func (d *UIConfigFormDialogModel) uiConfigFocusPrev() {
	d.focusIndex = (d.focusIndex + 5) % 6
	d.uiConfigUpdateFocus()
}

// uiConfigUpdateFocus applies focus styling to the currently focused element.
func (d *UIConfigFormDialogModel) uiConfigUpdateFocus() {
	d.PlaceholderInput.Blur()
	d.HelpTextInput.Blur()
	switch d.focusIndex {
	case 1:
		d.PlaceholderInput.Focus()
	case 2:
		d.HelpTextInput.Focus()
	}
}

// Render renders the UIConfig form dialog.
func (d UIConfigFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - 6

	titleText := d.titleStyle.Render(d.Title)

	var fieldRows []string

	// Widget selector (carousel)
	widgetLabel := d.labelStyle.Render("Widget")
	optLabel := d.WidgetOptions[d.WidgetIndex].Label
	var widgetView string
	if d.focusIndex == 0 {
		widgetView = lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Accent).
			Padding(0, 1).
			Render("\u25c0 " + optLabel + " \u25b6")
	} else {
		widgetView = lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Secondary).
			Padding(0, 1).
			Render("  " + optLabel + "  ")
	}
	fieldRows = append(fieldRows, widgetLabel+"\n"+widgetView)

	// Text input fields
	textFields := []struct {
		label string
		input textinput.Model
		idx   int
	}{
		{"Placeholder", d.PlaceholderInput, 1},
		{"Help Text", d.HelpTextInput, 2},
	}

	for _, f := range textFields {
		label := d.labelStyle.Render(f.label)
		style := d.inputStyle
		if d.focusIndex == f.idx {
			style = d.focusedInputStyle
		}
		input := style.Width(innerW).Render(f.input.View())
		fieldRows = append(fieldRows, label+"\n"+input)
	}

	// Hidden toggle
	hiddenLabel := d.labelStyle.Render("Hidden")
	checkmark := "[ ]"
	if d.HiddenToggle {
		checkmark = "[x]"
	}
	var hiddenView string
	if d.focusIndex == 3 {
		hiddenView = lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Primary).
			Background(config.DefaultStyle.Accent).
			Padding(0, 1).
			Render(checkmark + " Hidden")
	} else {
		hiddenView = lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Secondary).
			Padding(0, 1).
			Render(checkmark + " Hidden")
	}
	fieldRows = append(fieldRows, hiddenLabel+"\n"+hiddenView)

	// Buttons
	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == 4 {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == 5 {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render("Cancel")
	confirmBtn := confirmStyle.Render("Save")
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)

	// Assemble
	content := titleText + "\n\n"
	content += strings.Join(fieldRows, "\n")
	content += "\n\n" + buttonBar

	return d.borderStyle.Width(contentWidth).Render(content)
}

// UIConfigFormDialogOverlay positions a UIConfig form dialog over existing content.
func UIConfigFormDialogOverlay(content string, dialog UIConfigFormDialogModel, width, height int) string {
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

// --- Messages ---

// UIConfigFormDialogAcceptMsg is sent when the UIConfig form dialog is confirmed.
type UIConfigFormDialogAcceptMsg struct {
	FieldID     string
	Widget      string
	Placeholder string
	HelpText    string
	Hidden      bool
}

// UIConfigFormDialogCancelMsg is sent when the UIConfig form dialog is cancelled.
type UIConfigFormDialogCancelMsg struct{}

// ShowUIConfigFormDialogMsg triggers showing a blank UIConfig form dialog.
type ShowUIConfigFormDialogMsg struct {
	Title   string
	FieldID string
}

// ShowEditUIConfigFormDialogMsg triggers showing a pre-populated UIConfig form dialog.
type ShowEditUIConfigFormDialogMsg struct {
	Title    string
	FieldID  string
	Existing types.UIConfig
}

// UIConfigFormDialogSetMsg carries a UIConfig form dialog model to update.
type UIConfigFormDialogSetMsg struct {
	Dialog *UIConfigFormDialogModel
}

// UIConfigFormDialogActiveSetMsg carries the active state for a UIConfig form dialog.
type UIConfigFormDialogActiveSetMsg struct {
	Active bool
}

// --- Commands ---

// UIConfigFormDialogSetCmd creates a command to set the UIConfig form dialog model.
func UIConfigFormDialogSetCmd(dialog *UIConfigFormDialogModel) tea.Cmd {
	return func() tea.Msg { return UIConfigFormDialogSetMsg{Dialog: dialog} }
}

// UIConfigFormDialogActiveSetCmd creates a command to set the UIConfig form dialog active state.
func UIConfigFormDialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg { return UIConfigFormDialogActiveSetMsg{Active: active} }
}

// ShowUIConfigFormDialogCmd creates a command to show a blank UIConfig form dialog.
func ShowUIConfigFormDialogCmd(title, fieldID string) tea.Cmd {
	return func() tea.Msg {
		return ShowUIConfigFormDialogMsg{Title: title, FieldID: fieldID}
	}
}

// ShowEditUIConfigFormDialogCmd creates a command to show a pre-populated UIConfig form dialog.
func ShowEditUIConfigFormDialogCmd(title, fieldID string, existing types.UIConfig) tea.Cmd {
	return func() tea.Msg {
		return ShowEditUIConfigFormDialogMsg{Title: title, FieldID: fieldID, Existing: existing}
	}
}

// --- DB update messages ---

// UpdateFieldUIConfigRequestMsg triggers a field UIConfig update.
type UpdateFieldUIConfigRequestMsg struct {
	FieldID      string
	UIConfigJSON string
}

// UpdateFieldUIConfigCmd creates a command to update a field's UIConfig.
func UpdateFieldUIConfigCmd(fieldID, uiConfigJSON string) tea.Cmd {
	return func() tea.Msg {
		return UpdateFieldUIConfigRequestMsg{FieldID: fieldID, UIConfigJSON: uiConfigJSON}
	}
}

// FieldUIConfigUpdatedMsg is sent after a field's UIConfig is successfully updated.
type FieldUIConfigUpdatedMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
}

// --- Editor helpers ---

// editorWidgets is the set of widget types that support launching $EDITOR.
var editorWidgets = map[string]bool{
	"rich-text":   true,
	"code-editor": true,
	"json-editor": true,
	"markdown":    true,
}

// isEditorWidget returns true if the widget type supports launching $EDITOR.
func isEditorWidget(widget string) bool {
	return editorWidgets[widget]
}

// editorCommand returns the user's preferred editor command.
// Checks $VISUAL, then $EDITOR, then falls back to "vi".
func editorCommand() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	return "vi"
}

// --- Helper ---

// marshalUIConfig returns types.EmptyJSON if all fields are zero, otherwise marshals to JSON.
func marshalUIConfig(widget, placeholder, helpText string, hidden bool) string {
	if widget == "" && placeholder == "" && helpText == "" && !hidden {
		return types.EmptyJSON
	}
	uc := types.UIConfig{
		Widget:      widget,
		Placeholder: placeholder,
		HelpText:    helpText,
		Hidden:      hidden,
	}
	data, err := json.Marshal(uc)
	if err != nil {
		return types.EmptyJSON
	}
	return string(data)
}
