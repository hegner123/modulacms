package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
)

// 3/9 grid: left = environments list, right = detail (top) + actions (bottom)
var deployGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Environments"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Details"},
			{Height: 0.4, Title: "Actions"},
		}},
	},
}

// DeployScreen implements Screen for the deploy page.
type DeployScreen struct {
	GridScreen
	Environments    []config.DeployEnvironmentConfig
	LastResult      *DeploySyncResult
	LastHealth      *DeployHealthResult
	StatusMessage   string
	OperationActive bool
}

// NewDeployScreen creates a DeployScreen with the given environments.
func NewDeployScreen(envs []config.DeployEnvironmentConfig) *DeployScreen {
	cursorMax := len(envs) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &DeployScreen{
		GridScreen: GridScreen{
			Grid:      deployGrid,
			CursorMax: cursorMax,
		},
		Environments: envs,
	}
}

func (s *DeployScreen) PageIndex() PageIndex { return DEPLOYPAGE }

func (s *DeployScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
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

		_, cmd, handled := HandleCommonKeys(key, km, s.Cursor, len(s.Environments)-1)
		if handled {
			return s, cmd
		}

	// Data refresh messages
	case DeployEnvsSet:
		s.Environments = msg.Envs
		s.CursorMax = len(s.Environments) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
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
	cells := []CellContent{
		{Content: s.renderEnvsList(), TotalLines: len(s.Environments), ScrollOffset: ClampScroll(s.Cursor, len(s.Environments), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderActions()},
	}
	return s.RenderGrid(ctx, cells)
}

// renderEnvsList renders the deploy environments list.
func (s *DeployScreen) renderEnvsList() string {
	if len(s.Environments) == 0 {
		return "(no environments configured)\n\nAdd deploy_environments\nto modula.config.json"
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

// renderDetail renders the selected environment details and last operation result.
func (s *DeployScreen) renderDetail() string {
	if len(s.Environments) == 0 || s.Cursor >= len(s.Environments) {
		return " No environment selected"
	}

	env := s.Environments[s.Cursor]

	apiKeyDisplay := "(not set)"
	if env.APIKey != "" {
		apiKeyDisplay = "********"
	}

	lines := []string{
		fmt.Sprintf(" Name     %s", env.Name),
		fmt.Sprintf(" URL      %s", env.URL),
		fmt.Sprintf(" API Key  %s", apiKeyDisplay),
	}

	if s.LastHealth != nil && s.LastHealth.EnvName == env.Name {
		h := s.LastHealth
		lines = append(lines, "")
		if h.Err != "" {
			lines = append(lines, " Health: FAILED")
			lines = append(lines, fmt.Sprintf("   Error: %s", h.Err))
		} else {
			lines = append(lines, fmt.Sprintf(" Health   %s", h.Status))
			lines = append(lines, fmt.Sprintf(" Version  %s", h.Version))
			lines = append(lines, fmt.Sprintf(" Node     %s", h.NodeID))
		}
	}

	if s.LastResult != nil && s.LastResult.EnvName == env.Name {
		r := s.LastResult
		lines = append(lines, "")
		status := "SUCCESS"
		if !r.Success {
			status = "FAILED"
		}
		if r.DryRun {
			status += " (dry run)"
		}
		lines = append(lines, fmt.Sprintf(" %s: %s", r.Operation, status))
		if r.Duration != "" {
			lines = append(lines, fmt.Sprintf("   Duration: %s", r.Duration))
		}
		lines = append(lines, fmt.Sprintf("   Tables:   %d", len(r.TablesAffected)))

		totalRows := 0
		for _, count := range r.RowCounts {
			totalRows += count
		}
		lines = append(lines, fmt.Sprintf("   Rows:     %d", totalRows))

		for _, w := range r.Warnings {
			lines = append(lines, fmt.Sprintf("   WARN: %s", w))
		}
		for _, e := range r.Errors {
			lines = append(lines, fmt.Sprintf("   ERR:  %s", e))
		}
	}

	if s.StatusMessage != "" {
		lines = append(lines, "", " "+s.StatusMessage)
	}

	if s.OperationActive {
		lines = append(lines, "", "   Operation in progress...")
	}

	return strings.Join(lines, "\n")
}

// renderActions renders available actions.
func (s *DeployScreen) renderActions() string {
	lines := []string{
		" Actions",
		"",
		"   t  Test Connection",
		"   p  Pull (remote -> local)",
		"   s  Push (local -> remote)",
		"",
		"   Dry Run:",
		"   P  Dry Run Pull",
		"   S  Dry Run Push",
		"",
		fmt.Sprintf(" Environments: %d", len(s.Environments)),
	}

	if s.OperationActive {
		lines = append(lines, "", "   (operation running)")
	}

	return strings.Join(lines, "\n")
}
