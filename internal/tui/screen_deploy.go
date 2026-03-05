package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// DeployScreen implements Screen for the deploy page.
type DeployScreen struct {
	Cursor          int
	PanelFocus      FocusPanel
	Environments    []config.DeployEnvironmentConfig
	LastResult      *DeploySyncResult
	LastHealth      *DeployHealthResult
	StatusMessage   string
	OperationActive bool
}

// NewDeployScreen creates a DeployScreen with the given environments.
func NewDeployScreen(envs []config.DeployEnvironmentConfig) *DeployScreen {
	return &DeployScreen{
		Cursor:       0,
		PanelFocus:   TreePanel,
		Environments: envs,
	}
}

func (s *DeployScreen) PageIndex() PageIndex { return DEPLOYPAGE }

func (s *DeployScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Panel navigation
		if km.Matches(key, config.ActionNextPanel) {
			s.PanelFocus = (s.PanelFocus + 1) % 3
			return s, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			s.PanelFocus = (s.PanelFocus + 2) % 3
			return s, nil
		}

		// Cursor movement with status field clearing
		if km.Matches(key, config.ActionUp) {
			if s.Cursor > 0 {
				s.LastHealth = nil
				s.LastResult = nil
				s.StatusMessage = ""
				s.Cursor--
			}
			return s, nil
		}
		if km.Matches(key, config.ActionDown) {
			if s.Cursor < len(s.Environments)-1 {
				s.LastHealth = nil
				s.LastResult = nil
				s.StatusMessage = ""
				s.Cursor++
			}
			return s, nil
		}

		// Block all action keys while an operation is running
		if s.OperationActive {
			// Still allow quit/back via common keys
			_, cmd, handled := HandleCommonKeys(key, km, s.Cursor, len(s.Environments)-1)
			if handled {
				return s, cmd
			}
			return s, nil
		}

		// Guard: no environments or cursor out of range
		if len(s.Environments) == 0 || s.Cursor >= len(s.Environments) {
			_, cmd, handled := HandleCommonKeys(key, km, s.Cursor, 0)
			if handled {
				return s, cmd
			}
			return s, nil
		}

		env := s.Environments[s.Cursor]

		// Hardcoded deploy action keys
		switch key {
		case "t":
			return s, DeployTestConnectionCmd(env.Name)
		case "p":
			return s, ShowDeployConfirmPullCmd(env.Name)
		case "s":
			return s, ShowDeployConfirmPushCmd(env.Name)
		case "P":
			return s, DeployPullCmd(env.Name, true)
		case "S":
			return s, DeployPushCmd(env.Name, true)
		}

		// Common keys LAST (quit, back)
		// Note: cursor is handled above with status clearing, so HandleCommonKeys
		// will not re-handle up/down since we already returned above.
		_, cmd, handled := HandleCommonKeys(key, km, s.Cursor, len(s.Environments)-1)
		if handled {
			return s, cmd
		}

	// Data refresh messages
	case DeployEnvsSet:
		s.Environments = msg.Envs
		if s.Cursor >= len(s.Environments) && len(s.Environments) > 0 {
			s.Cursor = len(s.Environments) - 1
		}
		return s, nil

	case DeployTestConnectionResultMsg:
		s.OperationActive = false
		s.LastHealth = msg.Health
		if msg.Health.Err != "" {
			s.StatusMessage = fmt.Sprintf("Connection failed: %s", msg.Health.Err)
		} else {
			s.StatusMessage = fmt.Sprintf("Connected to %s (v%s)", msg.Health.EnvName, msg.Health.Version)
		}
		return s, nil

	case DeployTestConnectionRequestMsg:
		s.OperationActive = true
		s.StatusMessage = fmt.Sprintf("Testing connection to %s...", msg.EnvName)
		return s, nil

	case DeployPullRequestMsg:
		s.OperationActive = true
		opLabel := "pull"
		if msg.DryRun {
			opLabel = "dry-run pull"
		}
		s.StatusMessage = fmt.Sprintf("Running %s from %s...", opLabel, msg.EnvName)
		return s, nil

	case DeployPullResultMsg:
		s.OperationActive = false
		if msg.Err != "" {
			s.StatusMessage = fmt.Sprintf("Pull failed: %s", msg.Err)
			if msg.Result != nil {
				s.LastResult = msg.Result
				s.LastResult.Errors = append(s.LastResult.Errors, msg.Err)
			}
		} else {
			s.LastResult = msg.Result
			s.StatusMessage = fmt.Sprintf("Pull completed: %d tables", len(msg.Result.TablesAffected))
		}
		return s, nil

	case DeployPushRequestMsg:
		s.OperationActive = true
		opLabel := "push"
		if msg.DryRun {
			opLabel = "dry-run push"
		}
		s.StatusMessage = fmt.Sprintf("Running %s to %s...", opLabel, msg.EnvName)
		return s, nil

	case DeployPushResultMsg:
		s.OperationActive = false
		if msg.Err != "" {
			s.StatusMessage = fmt.Sprintf("Push failed: %s", msg.Err)
			if msg.Result != nil {
				s.LastResult = msg.Result
				s.LastResult.Errors = append(s.LastResult.Errors, msg.Err)
			}
		} else {
			s.LastResult = msg.Result
			s.StatusMessage = fmt.Sprintf("Push completed: %d tables", len(msg.Result.TablesAffected))
		}
		return s, nil
	}

	return s, nil
}

func (s *DeployScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{"t", "test"},
		{"p", "pull"},
		{"s", "push"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *DeployScreen) View(ctx AppContext) string {
	left := s.renderEnvsList()
	center := s.renderDetail()
	right := s.renderActions()

	layout := layoutForPage(DEPLOYPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	listLen := len(s.Environments)

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel, TotalLines: listLen, ScrollOffset: ClampScroll(s.Cursor, listLen, innerH)}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[1], Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

// renderEnvsList renders the deploy environments list for the left panel.
func (s *DeployScreen) renderEnvsList() string {
	if len(s.Environments) == 0 {
		return "(no environments configured)\n\nAdd deploy_environments\nto config.json"
	}

	lines := make([]string, 0, len(s.Environments))
	for i, env := range s.Environments {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, env.Name))
	}
	return strings.Join(lines, "\n")
}

// renderDetail renders the selected environment details and last operation
// result for the center panel.
func (s *DeployScreen) renderDetail() string {
	if len(s.Environments) == 0 || s.Cursor >= len(s.Environments) {
		return "No environment selected"
	}

	env := s.Environments[s.Cursor]

	apiKeyDisplay := "(not set)"
	if env.APIKey != "" {
		apiKeyDisplay = "********"
	}

	lines := []string{
		fmt.Sprintf("Name:    %s", env.Name),
		fmt.Sprintf("URL:     %s", env.URL),
		fmt.Sprintf("API Key: %s", apiKeyDisplay),
	}

	// Show health check result if available and matches this env
	if s.LastHealth != nil && s.LastHealth.EnvName == env.Name {
		h := s.LastHealth
		lines = append(lines, "")
		if h.Err != "" {
			lines = append(lines, "Health: FAILED")
			lines = append(lines, fmt.Sprintf("  Error: %s", h.Err))
		} else {
			lines = append(lines, fmt.Sprintf("Health:  %s", h.Status))
			lines = append(lines, fmt.Sprintf("Version: %s", h.Version))
			lines = append(lines, fmt.Sprintf("Node:    %s", h.NodeID))
		}
	}

	// Show last operation result if available and matches this env
	if s.LastResult != nil && s.LastResult.EnvName == env.Name {
		r := s.LastResult
		lines = append(lines, "")
		lines = append(lines, "--- Last Operation ---")

		status := "SUCCESS"
		if !r.Success {
			status = "FAILED"
		}
		if r.DryRun {
			status += " (dry run)"
		}

		lines = append(lines, fmt.Sprintf("  %s: %s", r.Operation, status))
		if r.Duration != "" {
			lines = append(lines, fmt.Sprintf("  Duration: %s", r.Duration))
		}
		lines = append(lines, fmt.Sprintf("  Tables:   %d", len(r.TablesAffected)))

		totalRows := 0
		for _, count := range r.RowCounts {
			totalRows += count
		}
		lines = append(lines, fmt.Sprintf("  Rows:     %d", totalRows))

		for _, w := range r.Warnings {
			lines = append(lines, fmt.Sprintf("  WARN: %s", w))
		}
		for _, e := range r.Errors {
			lines = append(lines, fmt.Sprintf("  ERR:  %s", e))
		}
	}

	// Show status message (errors, progress)
	if s.StatusMessage != "" {
		lines = append(lines, "")
		lines = append(lines, s.StatusMessage)
	}

	if s.OperationActive {
		lines = append(lines, "")
		lines = append(lines, "  Operation in progress...")
	}

	return strings.Join(lines, "\n")
}

// renderActions renders available actions for the right panel.
func (s *DeployScreen) renderActions() string {
	lines := []string{
		"Actions",
		"",
		"  t: Test Connection",
		"  p: Pull (remote -> local)",
		"  s: Push (local -> remote)",
		"",
		"  Dry Run:",
		"  P: Dry Run Pull",
		"  S: Dry Run Push",
		"",
		fmt.Sprintf("Environments: %d", len(s.Environments)),
	}

	if s.OperationActive {
		lines = append(lines, "")
		lines = append(lines, "  (operation running)")
	}

	return strings.Join(lines, "\n")
}
