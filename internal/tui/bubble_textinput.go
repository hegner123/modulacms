package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// TextInputBubble wraps textinput.Model as a generic FieldBubble for single-line text input.
// Specialized inputs (text, email, URL, slug) are thin constructors over this type.
type TextInputBubble struct {
	input    textinput.Model
	label    string
	validate func(string) error
}

// NewTextInputBubble creates a new TextInputBubble with the given configuration.
// Pass nil for validate if no validation is needed.
func NewTextInputBubble(label, placeholder string, charLimit int, validate func(string) error) *TextInputBubble {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	ti.SetWidth(50)
	return &TextInputBubble{input: ti, label: label, validate: validate}
}

func (b *TextInputBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)
	return b, cmd
}

func (b *TextInputBubble) View() string            { return b.input.View() }
func (b *TextInputBubble) Value() string           { return b.input.Value() }
func (b *TextInputBubble) SetValue(v string)       { b.input.SetValue(v) }
func (b *TextInputBubble) Focus() tea.Cmd          { return b.input.Focus() }
func (b *TextInputBubble) Blur()                   { b.input.Blur() }
func (b *TextInputBubble) Focused() bool           { return b.input.Focused() }
func (b *TextInputBubble) SetWidth(w int)          { b.input.SetWidth(w) }
func (b *TextInputBubble) SetPlaceholder(p string) { b.input.Placeholder = p }

// Validate runs the validation function if set, returning nil if no validator is configured.
func (b *TextInputBubble) Validate() error {
	if b.validate == nil {
		return nil
	}
	return b.validate(b.Value())
}
