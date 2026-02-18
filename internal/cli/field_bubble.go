package cli

import tea "github.com/charmbracelet/bubbletea"

// FieldBubble is the interface that field input components must satisfy for content entry forms.
type FieldBubble interface {
	// Update handles a Bubbletea message, returning the updated bubble and command.
	Update(msg tea.Msg) (FieldBubble, tea.Cmd)

	// View renders the bubble's current state.
	View() string

	// Value returns the current field value as a string for storage.
	Value() string

	// SetValue sets the field's value (for editing existing content).
	SetValue(value string)

	// Focus gives the bubble input focus.
	Focus() tea.Cmd

	// Blur removes input focus.
	Blur()

	// Focused returns whether the bubble currently has focus.
	Focused() bool

	// SetWidth sets the input width for layout within the dialog.
	SetWidth(w int)
}
