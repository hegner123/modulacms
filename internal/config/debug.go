package config

import "github.com/charmbracelet/lipgloss"

func (c Config) Stringify() string {
	out := make([]string, 0)
	return lipgloss.JoinVertical(lipgloss.Center, out...)
}
