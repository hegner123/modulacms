package tui

import (
	"time"

	"github.com/hegner123/modulacms/internal/config"
)

// Model is the top-level Bubbletea model for the new TUI.
type Model struct {
	Config  *config.Config
	Width   int
	Height  int
	Term    string
	Time    time.Time
	Verbose bool
	Focus   FocusPanel
}

// InitialModel creates a Model with the given config wired in.
// No initial command is returned.
func InitialModel(v *bool, c *config.Config) (Model, error) {
	verbose := false
	if v != nil {
		verbose = *v
	}

	m := Model{
		Config:  c,
		Verbose: verbose,
		Focus:   TreePanel,
	}

	return m, nil
}
