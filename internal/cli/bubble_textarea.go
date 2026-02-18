package cli

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "textarea",
		Label:       "Textarea",
		Description: "Multi-line text input",
		NewBubble:   func() FieldBubble { return NewTextareaBubble() },
	})
}

// TextareaBubble wraps textarea.Model as a FieldBubble for multi-line text input.
type TextareaBubble struct {
	input textarea.Model
}

// NewTextareaBubble creates a new TextareaBubble with default configuration.
func NewTextareaBubble() *TextareaBubble {
	ta := textarea.New()
	ta.Placeholder = "Enter text..."
	ta.CharLimit = 256
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	return &TextareaBubble{input: ta}
}

// Update handles Bubble Tea messages for the textarea bubble.
func (b *TextareaBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

// View returns the rendered representation of the textarea bubble.
func (b *TextareaBubble) View() string { return b.input.View() }

// Value returns the current value of the textarea input.
func (b *TextareaBubble) Value() string { return b.input.Value() }

// SetValue sets the value of the textarea input.
func (b *TextareaBubble) SetValue(v string) { b.input.SetValue(v) }

// Focus sets focus to the textarea input.
func (b *TextareaBubble) Focus() tea.Cmd { return b.input.Focus() }

// Blur removes focus from the textarea input.
func (b *TextareaBubble) Blur() { b.input.Blur() }

// Focused returns whether the textarea input is currently focused.
func (b *TextareaBubble) Focused() bool { return b.input.Focused() }

// SetWidth sets the textarea input width for layout.
func (b *TextareaBubble) SetWidth(w int) { b.input.SetWidth(w) }
