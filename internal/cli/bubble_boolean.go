package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "boolean",
		Label:       "Boolean",
		Description: "True/false toggle",
		NewBubble:   func() FieldBubble { return NewBooleanBubble() },
	})
}

// BooleanBubble is a FieldBubble that toggles between true and false.
type BooleanBubble struct {
	value   bool
	focused bool
}

// NewBooleanBubble creates a new BooleanBubble defaulting to false.
func NewBooleanBubble() *BooleanBubble {
	return &BooleanBubble{}
}

// Update handles Bubble Tea messages for the boolean bubble.
func (b *BooleanBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	if !b.focused {
		return b, nil
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left", "right", " ":
			b.value = !b.value
		}
	}
	return b, nil
}

// View returns the rendered representation of the boolean toggle.
func (b *BooleanBubble) View() string {
	trueLabel := "true"
	falseLabel := "false"

	accent := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Primary).
		Background(config.DefaultStyle.Accent).
		Padding(0, 1)
	dim := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Padding(0, 1)

	if b.value {
		return accent.Render(trueLabel) + " " + dim.Render(falseLabel)
	}
	return dim.Render(trueLabel) + " " + accent.Render(falseLabel)
}

// Value returns "true" or "false".
func (b *BooleanBubble) Value() string {
	if b.value {
		return "true"
	}
	return "false"
}

// SetValue parses a string to set the boolean value.
func (b *BooleanBubble) SetValue(v string) {
	b.value = v == "true" || v == "1"
}

// Focus gives the bubble input focus.
func (b *BooleanBubble) Focus() tea.Cmd {
	b.focused = true
	return nil
}

// Blur removes input focus.
func (b *BooleanBubble) Blur() { b.focused = false }

// Focused returns whether the bubble currently has focus.
func (b *BooleanBubble) Focused() bool { return b.focused }

// SetWidth is a no-op for boolean toggles.
func (b *BooleanBubble) SetWidth(_ int) {}
