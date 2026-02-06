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

// TextBubble wraps textinput.Model as a FieldBubble.
type TextBubble struct {
	input textinput.Model
}

func NewTextBubble() *TextBubble {
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.CharLimit = 256
	ti.Width = 50
	return &TextBubble{input: ti}
}

func (b *TextBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

func (b *TextBubble) View() string     { return b.input.View() }
func (b *TextBubble) Value() string    { return b.input.Value() }
func (b *TextBubble) SetValue(v string) { b.input.SetValue(v) }
func (b *TextBubble) Focus() tea.Cmd   { return b.input.Focus() }
func (b *TextBubble) Blur()            { b.input.Blur() }
func (b *TextBubble) Focused() bool    { return b.input.Focused() }
