package config

import "github.com/charmbracelet/lipgloss"

// Stringify returns a formatted string representation of the Config for display.
func (c Config) Stringify() string {
	out := make([]string, 0)
	return lipgloss.JoinVertical(lipgloss.Center, out...)
}
