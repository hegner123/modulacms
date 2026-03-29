package tui

import (
	"fmt"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/update"
	"github.com/hegner123/modulacms/internal/utility"
)

// Actions grid: 3 columns
//
//	Col 0 (span 3): Actions menu
//	Col 1 (span 6): Details (top), Help (bottom)
//	Col 2 (span 3): System (top), Updates (bottom)
var actionsGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1, Title: "Actions"},
		}},
		{Span: 6, Cells: []GridCell{
			{Height: 0.70, Title: "Details"},
			{Height: 0.30, Title: "Help"},
		}},
		{Span: 3, Cells: []GridCell{
			{Height: 0.55, Title: "System"},
			{Height: 0.45, Title: "Updates"},
		}},
	},
}

// UpdateCheckMsg delivers the result of an async update check.
type UpdateCheckMsg struct {
	Available  bool
	NewVersion string
	Err        error
}

// UpdateCheckCmd runs an async update check against the GitHub releases API.
func UpdateCheckCmd() tea.Cmd {
	return func() tea.Msg {
		current := utility.GetCurrentVersion()
		release, available, err := update.CheckForUpdates(current, "stable")
		if err != nil {
			return UpdateCheckMsg{Err: err}
		}
		version := ""
		if release != nil {
			version = release.TagName
		}
		return UpdateCheckMsg{Available: available, NewVersion: version}
	}
}

// ActionsScreen implements Screen for the actions page.
type ActionsScreen struct {
	GridScreen
	IsRemote       bool
	UpdateChecked  bool
	UpdateAvail    bool
	UpdateVersion  string
	UpdateCheckErr error
}

// NewActionsScreen creates an ActionsScreen.
func NewActionsScreen(isRemote bool) *ActionsScreen {
	actions := ActionsMenuForMode(isRemote)
	cursorMax := len(actions) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &ActionsScreen{
		GridScreen: GridScreen{
			Grid:      actionsGrid,
			CursorMax: cursorMax,
		},
		IsRemote: isRemote,
	}
}

func (s *ActionsScreen) PageIndex() PageIndex { return ACTIONSPAGE }

func (s *ActionsScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	actions := ActionsMenuForMode(s.IsRemote)
	s.CursorMax = len(actions) - 1

	switch msg := msg.(type) {
	case UpdateCheckMsg:
		s.UpdateChecked = true
		s.UpdateAvail = msg.Available
		s.UpdateVersion = msg.NewVersion
		s.UpdateCheckErr = msg.Err
		return s, nil

	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Select action
		if km.Matches(key, config.ActionSelect) {
			if s.FocusIndex != 0 || s.Cursor >= len(actions) {
				return s, nil
			}
			action := actions[s.Cursor]
			if action.Destructive {
				return s, func() tea.Msg {
					return ActionConfirmMsg{ActionIndex: action.Index}
				}
			}
			return s, tea.Batch(
				LoadingStartCmd(),
				RunActionCmd(ActionParams{
					Config:         ctx.Config,
					UserID:         ctx.UserID,
					SSHFingerprint: ctx.SSHFingerprint,
					SSHKeyType:     ctx.SSHKeyType,
					SSHPublicKey:   ctx.SSHPublicKey,
				}, action.Index),
			)
		}

		// Common keys (quit, back, cursor)
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}
	}

	return s, nil
}

func (s *ActionsScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionSelect), "run"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

func (s *ActionsScreen) View(ctx AppContext) string {
	actions := ActionsMenuForMode(s.IsRemote)

	cells := []CellContent{
		{Content: s.renderMenu(actions)},
		{Content: s.renderDetail(actions)},
		{Content: s.renderHelp(actions)},
		{Content: s.renderSystem(ctx)},
		{Content: s.renderUpdates()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *ActionsScreen) renderMenu(actions []ActionItem) string {
	if len(actions) == 0 {
		return "(no actions)"
	}
	warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
	lines := make([]string, 0, len(actions))
	for i, action := range actions {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		label := action.Label
		if action.Destructive {
			label = warnStyle.Render(label)
		}
		lines = append(lines, fmt.Sprintf(" %s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

func (s *ActionsScreen) renderDetail(actions []ActionItem) string {
	if len(actions) == 0 || s.Cursor >= len(actions) {
		return " No action selected"
	}
	action := actions[s.Cursor]

	accent := lipgloss.NewStyle().Bold(true)
	lines := []string{
		accent.Render(" " + action.Label),
		"",
		" " + action.Description,
	}

	if action.Destructive {
		warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
		lines = append(lines, "")
		lines = append(lines, warnStyle.Render(" Requires confirmation before execution"))
	}

	return strings.Join(lines, "\n")
}

func (s *ActionsScreen) renderHelp(actions []ActionItem) string {
	lines := []string{
		" Press Enter to run the selected action.",
		" Use Tab to switch between panels.",
	}
	if s.IsRemote {
		lines = append(lines, "")
		lines = append(lines, " Remote mode: limited actions available.")
	}
	return strings.Join(lines, "\n")
}

func (s *ActionsScreen) renderSystem(ctx AppContext) string {
	faint := lipgloss.NewStyle().Faint(true)

	version := utility.Version
	if utility.IsDevBuild() {
		version += " " + faint.Render("(dev)")
	}

	lines := []string{
		fmt.Sprintf(" Version  %s", version),
		fmt.Sprintf(" Commit   %s", shortenHash(utility.GitCommit, 8)),
		fmt.Sprintf(" Built    %s", utility.BuildDate),
		fmt.Sprintf(" Go       %s", runtime.Version()),
		fmt.Sprintf(" OS/Arch  %s/%s", runtime.GOOS, runtime.GOARCH),
	}

	if ctx.Config != nil {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf(" DB       %s", ctx.Config.Db_Driver))
		lines = append(lines, fmt.Sprintf(" Env      %s", ctx.Config.Environment))
		if ctx.Config.Plugin_Enabled {
			lines = append(lines, fmt.Sprintf(" Plugins  %s", "enabled"))
		}
	}

	return strings.Join(lines, "\n")
}

func (s *ActionsScreen) renderUpdates() string {
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	warn := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
	faint := lipgloss.NewStyle().Faint(true)

	if !s.UpdateChecked {
		return faint.Render(" Checking for updates...")
	}

	if s.UpdateCheckErr != nil {
		return warn.Render(fmt.Sprintf(" Check failed: %s", s.UpdateCheckErr))
	}

	current := utility.GetCurrentVersion()
	lines := []string{
		fmt.Sprintf(" Current  %s", current),
	}

	if s.UpdateAvail {
		lines = append(lines, accent.Render(fmt.Sprintf(" Latest   %s", s.UpdateVersion)))
		lines = append(lines, "")
		lines = append(lines, accent.Render(" Update available!"))
		lines = append(lines, " Select \"Check for Updates\" to install.")
	} else {
		lines = append(lines, "")
		lines = append(lines, accent.Render(" Up to date"))
	}

	return strings.Join(lines, "\n")
}

// shortenHash returns the first n characters of s, or s if shorter.
func shortenHash(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
