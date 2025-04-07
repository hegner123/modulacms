package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

type CLIControls interface {
	update(tea.Msg) (tea.Model, tea.Cmd)
}

type CLIControl struct {
	model    *tea.Model
	cursor   int
	up       []string
	down     []string
	next     []string
	nextFunc func()
	prev     []string
	prevFunc func()
}

func (c CLIControl) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return *c.model, nil
}

func NewCLIControl(m tea.Model) CLIControls {
	return CLIControl{
		cursor:   0,
		up:       []string{"up", "k"},
		down:     []string{"down", "j"},
		next:     []string{"enter", "l"},
		prev:     []string{"shift+tab", "h"},
		nextFunc: func() {},
		prevFunc: func() {},
	}

}
