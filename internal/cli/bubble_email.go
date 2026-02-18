package cli

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "email",
		Label:       "Email",
		Description: "Email address input",
		NewBubble:   func() FieldBubble { return NewEmailBubble() },
	})
}

// EmailBubble wraps textinput.Model as a FieldBubble for email input.
type EmailBubble struct {
	input textinput.Model
}

// NewEmailBubble creates a new EmailBubble with default configuration.
func NewEmailBubble() *EmailBubble {
	ti := textinput.New()
	ti.Placeholder = "user@example.com"
	ti.CharLimit = 256
	ti.Width = 50
	return &EmailBubble{input: ti}
}

// Update handles Bubble Tea messages for the email bubble.
func (b *EmailBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

// View returns the rendered representation of the email bubble.
func (b *EmailBubble) View() string { return b.input.View() }

// Value returns the current value of the email input.
func (b *EmailBubble) Value() string { return b.input.Value() }

// SetValue sets the value of the email input.
func (b *EmailBubble) SetValue(v string) { b.input.SetValue(v) }

// Focus sets focus to the email input.
func (b *EmailBubble) Focus() tea.Cmd { return b.input.Focus() }

// Blur removes focus from the email input.
func (b *EmailBubble) Blur() { b.input.Blur() }

// Focused returns whether the email input is currently focused.
func (b *EmailBubble) Focused() bool { return b.input.Focused() }

// SetWidth sets the email input width for layout.
func (b *EmailBubble) SetWidth(w int) { b.input.Width = w }
