package cli

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "select",
		Label:       "Select",
		Description: "Single option selector",
		NewBubble:   func() FieldBubble { return NewSelectBubble() },
	})
}

// SelectOption represents a selectable option in the select bubble.
type SelectOption struct {
	Label string
	Value string
}

// SelectBubble is a FieldBubble that presents a vertical list of options.
type SelectBubble struct {
	options []SelectOption
	cursor  int
	focused bool
}

// NewSelectBubble creates a new SelectBubble with no options.
func NewSelectBubble() *SelectBubble {
	return &SelectBubble{}
}

// SetOptions sets the available options for selection.
func (b *SelectBubble) SetOptions(opts []SelectOption) {
	b.options = opts
	if b.cursor >= len(opts) {
		b.cursor = 0
	}
}

// ParseOptionsFromData parses options from a field's data JSON column.
// Expected format: {"options": ["a", "b", "c"]}
func (b *SelectBubble) ParseOptionsFromData(data string) {
	if data == "" || data == "{}" {
		return
	}
	var parsed struct {
		Options []string `json:"options"`
	}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return
	}
	opts := make([]SelectOption, len(parsed.Options))
	for i, o := range parsed.Options {
		opts[i] = SelectOption{Label: o, Value: o}
	}
	b.SetOptions(opts)
}

// Update handles Bubble Tea messages for the select bubble.
func (b *SelectBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	if !b.focused || len(b.options) == 0 {
		return b, nil
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if b.cursor > 0 {
				b.cursor--
			}
		case "down", "j":
			if b.cursor < len(b.options)-1 {
				b.cursor++
			}
		}
	}
	return b, nil
}

// View returns the rendered representation of the select list.
func (b *SelectBubble) View() string {
	if len(b.options) == 0 {
		dim := lipgloss.NewStyle().Foreground(config.DefaultStyle.Tertiary)
		return dim.Render("(no options)")
	}

	selected := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Primary).
		Background(config.DefaultStyle.Accent).
		Padding(0, 1)
	normal := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary).
		Padding(0, 1)

	var out string
	for i, opt := range b.options {
		if i > 0 {
			out += "\n"
		}
		if i == b.cursor {
			out += selected.Render("▸ " + opt.Label)
		} else {
			out += normal.Render("  " + opt.Label)
		}
	}
	return out
}

// Value returns the selected option's Value string.
func (b *SelectBubble) Value() string {
	if len(b.options) == 0 || b.cursor >= len(b.options) {
		return ""
	}
	return b.options[b.cursor].Value
}

// SetValue finds a matching option by value and sets the cursor.
func (b *SelectBubble) SetValue(v string) {
	for i, opt := range b.options {
		if opt.Value == v {
			b.cursor = i
			return
		}
	}
}

// Focus gives the bubble input focus.
func (b *SelectBubble) Focus() tea.Cmd {
	b.focused = true
	return nil
}

// Blur removes input focus.
func (b *SelectBubble) Blur() { b.focused = false }

// Focused returns whether the bubble currently has focus.
func (b *SelectBubble) Focused() bool { return b.focused }

// SetWidth is a no-op for the select list.
func (b *SelectBubble) SetWidth(_ int) {}
