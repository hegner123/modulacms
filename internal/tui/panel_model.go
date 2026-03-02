package tui

import (
	"time"

	"github.com/hegner123/modulacms/internal/config"
)

// PanelModel is the top-level Bubbletea model for the new TUI.
type PanelModel struct {
	Config  *config.Config
	Width   int
	Height  int
	Term    string
	Time    time.Time
	Verbose bool
	Focus   FocusPanel
}

// NewPanelModel creates a PanelModel with the given config wired in.
// No initial command is returned.
func NewPanelModel(v *bool, c *config.Config) (PanelModel, error) {
	verbose := false
	if v != nil {
		verbose = *v
	}

	m := PanelModel{
		Config:  c,
		Verbose: verbose,
		Focus:   TreePanel,
	}

	return m, nil
}
