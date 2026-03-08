package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func namedKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// --- NumberBubble ---

func TestNumberBubble_AllowsDigits(t *testing.T) {
	b := NewNumberBubble()
	b.Focus()
	b.Update(runeKey('5'))
	b.Update(runeKey('3'))
	if got := b.Value(); got != "53" {
		t.Errorf("expected %q, got %q", "53", got)
	}
}

func TestNumberBubble_RejectsNonNumeric(t *testing.T) {
	b := NewNumberBubble()
	b.Focus()
	b.Update(runeKey('a'))
	b.Update(runeKey('!'))
	b.Update(runeKey('z'))
	if got := b.Value(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestNumberBubble_MinusPosition(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"minus on empty input", "", "-"},
		{"minus blocked on non-empty", "5", "5"},
		{"minus blocked on existing minus", "-", "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewNumberBubble()
			b.Focus()
			if tt.initial != "" {
				b.SetValue(tt.initial)
			}
			b.Update(runeKey('-'))
			if got := b.Value(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestNumberBubble_DecimalFiltering(t *testing.T) {
	tests := []struct {
		name     string
		sequence []rune
		expected string
	}{
		{"single decimal allowed", []rune{'1', '.', '5'}, "1.5"},
		{"duplicate decimal blocked", []rune{'1', '.', '2', '.', '3'}, "1.23"},
		{"leading decimal", []rune{'.', '5'}, ".5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewNumberBubble()
			b.Focus()
			for _, r := range tt.sequence {
				b.Update(runeKey(r))
			}
			if got := b.Value(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// --- DatePickerBubble ---

func TestDatePickerBubble_MonthRollover(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		key      rune
		expected string
	}{
		{"dec to jan", "2024-12-15", 'L', "2025-01-15"},
		{"jan to dec", "2024-01-15", 'H', "2023-12-15"},
		{"jan 31 to feb leap year clamp", "2024-01-31", 'L', "2024-02-29"},
		{"jan 31 to feb non-leap clamp", "2025-01-31", 'L', "2025-02-28"},
		{"mar 31 to feb clamp backward", "2024-03-31", 'H', "2024-02-29"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewDatePickerBubble()
			b.SetValue(tt.initial)
			b.Focus()
			b.Update(runeKey(tt.key))
			if got := b.Value(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestDatePickerBubble_HourWrap(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		key      tea.KeyType
		expected string
	}{
		{"23 wraps to 0", "23:00", tea.KeyUp, "00:00"},
		{"0 wraps to 23", "00:00", tea.KeyDown, "23:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewTimePickerBubble()
			b.SetValue(tt.initial)
			b.Focus()
			b.Update(namedKey(tt.key))
			if got := b.Value(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestDatePickerBubble_MinuteWrap(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		key      tea.KeyType
		expected string
	}{
		{"59 wraps to 0", "00:59", tea.KeyUp, "00:00"},
		{"0 wraps to 59", "00:00", tea.KeyDown, "00:59"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewTimePickerBubble()
			b.SetValue(tt.initial)
			b.Focus()
			// Move sub-focus from hour to minute
			b.Update(namedKey(tea.KeyRight))
			b.Update(namedKey(tt.key))
			if got := b.Value(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// --- SelectBubble ---

func TestSelectBubble_EmptyOptions(t *testing.T) {
	b := NewSelectBubble()
	if got := b.Value(); got != "" {
		t.Errorf("expected empty value, got %q", got)
	}
}

func TestSelectBubble_CursorNavigation(t *testing.T) {
	b := NewSelectBubble()
	b.SetOptions([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	})
	b.Focus()

	if got := b.Value(); got != "a" {
		t.Errorf("initial: expected %q, got %q", "a", got)
	}

	b.Update(namedKey(tea.KeyDown))
	if got := b.Value(); got != "b" {
		t.Errorf("after down: expected %q, got %q", "b", got)
	}

	b.Update(namedKey(tea.KeyDown))
	if got := b.Value(); got != "c" {
		t.Errorf("after second down: expected %q, got %q", "c", got)
	}

	// Cursor should not go past last option
	b.Update(namedKey(tea.KeyDown))
	if got := b.Value(); got != "c" {
		t.Errorf("past end: expected %q, got %q", "c", got)
	}

	b.Update(namedKey(tea.KeyUp))
	if got := b.Value(); got != "b" {
		t.Errorf("after up: expected %q, got %q", "b", got)
	}
}

func TestSelectBubble_CursorBoundsOnUp(t *testing.T) {
	b := NewSelectBubble()
	b.SetOptions([]SelectOption{
		{Label: "X", Value: "x"},
	})
	b.Focus()

	// Cursor at 0, up should stay at 0
	b.Update(namedKey(tea.KeyUp))
	if got := b.Value(); got != "x" {
		t.Errorf("expected %q, got %q", "x", got)
	}
}

func TestSelectBubble_OptionReplacement(t *testing.T) {
	b := NewSelectBubble()
	b.SetOptions([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	})
	b.Focus()

	// Move cursor to index 2
	b.Update(namedKey(tea.KeyDown))
	b.Update(namedKey(tea.KeyDown))
	if got := b.Value(); got != "c" {
		t.Fatalf("setup: expected %q, got %q", "c", got)
	}

	// Replace with shorter list — cursor should reset
	b.SetOptions([]SelectOption{
		{Label: "X", Value: "x"},
	})
	if got := b.Value(); got != "x" {
		t.Errorf("after replacement: expected %q, got %q", "x", got)
	}
}

func TestSelectBubble_SetValue(t *testing.T) {
	b := NewSelectBubble()
	b.SetOptions([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	})
	b.SetValue("b")
	if got := b.Value(); got != "b" {
		t.Errorf("expected %q, got %q", "b", got)
	}
}

// --- BooleanBubble ---

func TestBooleanBubble_DefaultFalse(t *testing.T) {
	b := NewBooleanBubble()
	if got := b.Value(); got != "false" {
		t.Errorf("expected %q, got %q", "false", got)
	}
}

func TestBooleanBubble_Toggle(t *testing.T) {
	b := NewBooleanBubble()
	b.Focus()

	// Space toggles to true
	b.Update(runeKey(' '))
	if got := b.Value(); got != "true" {
		t.Errorf("after space: expected %q, got %q", "true", got)
	}

	// Space toggles back to false
	b.Update(runeKey(' '))
	if got := b.Value(); got != "false" {
		t.Errorf("after second space: expected %q, got %q", "false", got)
	}

	// Left toggles to true
	b.Update(namedKey(tea.KeyLeft))
	if got := b.Value(); got != "true" {
		t.Errorf("after left: expected %q, got %q", "true", got)
	}

	// Right toggles to false
	b.Update(namedKey(tea.KeyRight))
	if got := b.Value(); got != "false" {
		t.Errorf("after right: expected %q, got %q", "false", got)
	}
}

func TestBooleanBubble_IgnoresWhenUnfocused(t *testing.T) {
	b := NewBooleanBubble()
	// Not focused — space should be ignored
	b.Update(runeKey(' '))
	if got := b.Value(); got != "false" {
		t.Errorf("expected %q, got %q", "false", got)
	}
}

func TestBooleanBubble_SetValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"true", "true"},
		{"1", "true"},
		{"false", "false"},
		{"0", "false"},
		{"", "false"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			b := NewBooleanBubble()
			b.SetValue(tt.input)
			if got := b.Value(); got != tt.expected {
				t.Errorf("SetValue(%q): expected %q, got %q", tt.input, tt.expected, got)
			}
		})
	}
}
