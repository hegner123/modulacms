package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// ActionsScreen implements Screen for the actions page.
type ActionsScreen struct {
	Cursor     int
	IsRemote   bool
	PanelFocus FocusPanel
}

// NewActionsScreen creates an ActionsScreen.
func NewActionsScreen(isRemote bool) *ActionsScreen {
	return &ActionsScreen{
		Cursor:     0,
		IsRemote:   isRemote,
		PanelFocus: ContentPanel,
	}
}

func (s *ActionsScreen) PageIndex() PageIndex { return ACTIONSPAGE }

func (s *ActionsScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	actions := ActionsMenuForMode(s.IsRemote)
	cursorMax := len(actions) - 1

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

		// Select action
		if km.Matches(key, config.ActionSelect) {
			if s.Cursor >= len(actions) {
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
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, cursorMax)
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
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *ActionsScreen) View(ctx AppContext) string {
	actions := ActionsMenuForMode(s.IsRemote)
	left := s.renderMenu(actions)
	center := s.renderDetail(actions)
	right := s.renderStatus(actions)

	layout := layoutForPage(ACTIONSPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	// Single-panel page: render center panel full width
	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: "Actions", Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: "Details", Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: "Status", Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

func (s *ActionsScreen) renderMenu(actions []ActionItem) string {
	if len(actions) == 0 {
		return "(no actions)"
	}
	warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
	lines := make([]string, 0, len(actions))
	for i, action := range actions {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		label := action.Label
		if action.Destructive {
			label = warnStyle.Render(label)
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

func (s *ActionsScreen) renderDetail(actions []ActionItem) string {
	if len(actions) == 0 || s.Cursor >= len(actions) {
		return "No action selected"
	}
	action := actions[s.Cursor]
	return fmt.Sprintf("%s\n\n%s", action.Label, action.Description)
}

func (s *ActionsScreen) renderStatus(actions []ActionItem) string {
	lines := []string{
		"Actions",
		"",
		fmt.Sprintf("  Total: %d", len(actions)),
	}
	if s.Cursor < len(actions) && actions[s.Cursor].Destructive {
		lines = append(lines, "")
		warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
		lines = append(lines, warnStyle.Render("  !! Destructive"))
	}
	return strings.Join(lines, "\n")
}
