package cli

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "number",
		Label:       "Number",
		Description: "Numeric input",
		NewBubble:   func() FieldBubble { return NewNumberBubble() },
	})
}

// NumberBubble wraps textinput.Model as a FieldBubble that filters non-numeric input.
type NumberBubble struct {
	input textinput.Model
}

// NewNumberBubble creates a new NumberBubble with default configuration.
func NewNumberBubble() *NumberBubble {
	ti := textinput.New()
	ti.Placeholder = "0"
	ti.CharLimit = 32
	ti.Width = 50
	return &NumberBubble{input: ti}
}

// Update handles Bubble Tea messages for the number bubble.
// Intercepts key runes and drops non-numeric characters.
func (b *NumberBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyRunes && len(keyMsg.Runes) == 1 {
			r := keyMsg.Runes[0]
			switch {
			case r >= '0' && r <= '9':
				// allow digits
			case r == '-':
				// allow minus only at position 0
				if len(b.input.Value()) != 0 {
					return b, nil
				}
			case r == '.':
				// allow single decimal point
				for _, c := range b.input.Value() {
					if c == '.' {
						return b, nil
					}
				}
			default:
				return b, nil
			}
		}
	}
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

// View returns the rendered representation of the number bubble.
func (b *NumberBubble) View() string { return b.input.View() }

// Value returns the current value of the number input.
func (b *NumberBubble) Value() string { return b.input.Value() }

// SetValue sets the value of the number input.
func (b *NumberBubble) SetValue(v string) { b.input.SetValue(v) }

// Focus sets focus to the number input.
func (b *NumberBubble) Focus() tea.Cmd { return b.input.Focus() }

// Blur removes focus from the number input.
func (b *NumberBubble) Blur() { b.input.Blur() }

// Focused returns whether the number input is currently focused.
func (b *NumberBubble) Focused() bool { return b.input.Focused() }

// SetWidth sets the number input width for layout.
func (b *NumberBubble) SetWidth(w int) { b.input.Width = w }
