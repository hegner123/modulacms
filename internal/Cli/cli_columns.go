package cli

import (
	"github.com/charmbracelet/lipgloss"
)

func NewHorizontalGroup(p lipgloss.Position, s []string) string {
	return lipgloss.JoinHorizontal(p, s...)
}

func NewVerticalGroup(p lipgloss.Position, s []string) string {
	return lipgloss.JoinVertical(p, s...)
}
