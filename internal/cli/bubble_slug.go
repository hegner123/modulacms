package cli

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "slug",
		Label:       "Slug",
		Description: "URL slug input",
		NewBubble:   func() FieldBubble { return NewSlugBubble() },
	})
}

// SlugBubble wraps textinput.Model as a FieldBubble for slug input.
type SlugBubble struct {
	input textinput.Model
}

// NewSlugBubble creates a new SlugBubble with default configuration.
func NewSlugBubble() *SlugBubble {
	ti := textinput.New()
	ti.Placeholder = "my-page-slug"
	ti.CharLimit = 256
	ti.Width = 50
	return &SlugBubble{input: ti}
}

// Update handles Bubble Tea messages for the slug bubble.
func (b *SlugBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

// View returns the rendered representation of the slug bubble.
func (b *SlugBubble) View() string { return b.input.View() }

// Value returns the current value of the slug input.
func (b *SlugBubble) Value() string { return b.input.Value() }

// SetValue sets the value of the slug input.
func (b *SlugBubble) SetValue(v string) { b.input.SetValue(v) }

// Focus sets focus to the slug input.
func (b *SlugBubble) Focus() tea.Cmd { return b.input.Focus() }

// Blur removes focus from the slug input.
func (b *SlugBubble) Blur() { b.input.Blur() }

// Focused returns whether the slug input is currently focused.
func (b *SlugBubble) Focused() bool { return b.input.Focused() }

// SetWidth sets the slug input width for layout.
func (b *SlugBubble) SetWidth(w int) { b.input.Width = w }
