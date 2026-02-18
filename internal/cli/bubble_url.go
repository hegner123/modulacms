package cli

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "url",
		Label:       "URL",
		Description: "URL input",
		NewBubble:   func() FieldBubble { return NewURLBubble() },
	})
}

// URLBubble wraps textinput.Model as a FieldBubble for URL input.
type URLBubble struct {
	input textinput.Model
}

// NewURLBubble creates a new URLBubble with default configuration.
func NewURLBubble() *URLBubble {
	ti := textinput.New()
	ti.Placeholder = "https://..."
	ti.CharLimit = 512
	ti.Width = 50
	return &URLBubble{input: ti}
}

// Update handles Bubble Tea messages for the URL bubble.
func (b *URLBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

// View returns the rendered representation of the URL bubble.
func (b *URLBubble) View() string { return b.input.View() }

// Value returns the current value of the URL input.
func (b *URLBubble) Value() string { return b.input.Value() }

// SetValue sets the value of the URL input.
func (b *URLBubble) SetValue(v string) { b.input.SetValue(v) }

// Focus sets focus to the URL input.
func (b *URLBubble) Focus() tea.Cmd { return b.input.Focus() }

// Blur removes focus from the URL input.
func (b *URLBubble) Blur() { b.input.Blur() }

// Focused returns whether the URL input is currently focused.
func (b *URLBubble) Focused() bool { return b.input.Focused() }

// SetWidth sets the URL input width for layout.
func (b *URLBubble) SetWidth(w int) { b.input.Width = w }
