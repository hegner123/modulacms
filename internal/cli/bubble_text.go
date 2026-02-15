package cli

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "text",
		Label:       "Text",
		Description: "Single-line text input",
		NewBubble:   func() FieldBubble { return NewTextBubble() },
	})
}

// TextBubble wraps textinput.Model as a FieldBubble for text input.
type TextBubble struct {
	input textinput.Model
}

// NewTextBubble creates a new TextBubble with default configuration.
func NewTextBubble() *TextBubble {
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.CharLimit = 256
	ti.Width = 50
	return &TextBubble{input: ti}
}

// Update handles Bubble Tea messages for the text bubble.
func (b *TextBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

// View returns the rendered representation of the text bubble.
func (b *TextBubble) View() string     { return b.input.View() }

// Value returns the current value of the text input.
func (b *TextBubble) Value() string    { return b.input.Value() }

// SetValue sets the value of the text input.
func (b *TextBubble) SetValue(v string) { b.input.SetValue(v) }

// Focus sets focus to the text input.
func (b *TextBubble) Focus() tea.Cmd   { return b.input.Focus() }

// Blur removes focus from the text input.
func (b *TextBubble) Blur()            { b.input.Blur() }

// Focused returns whether the text input is currently focused.
func (b *TextBubble) Focused() bool    { return b.input.Focused() }
