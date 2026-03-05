package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// HomeScreen implements Screen for the HOMEPAGE.
type HomeScreen struct {
	PanelScreen
	Menu     []Page
	Username string
}

// NewHomeScreen creates a HomeScreen with the given menu and username.
func NewHomeScreen(menu []Page, username string) *HomeScreen {
	return &HomeScreen{
		PanelScreen: PanelScreen{
			Layout:     layoutForPage(HOMEPAGE),
			PanelFocus: ContentPanel,
			CursorMax:  len(menu) - 1,
		},
		Menu:     menu,
		Username: username,
	}
}

func (s *HomeScreen) PageIndex() PageIndex { return HOMEPAGE }

func (s *HomeScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case PageMenuSet:
		s.Menu = msg.PageMenu
		s.CursorMax = len(s.Menu) - 1
		if s.Cursor > s.CursorMax && s.CursorMax >= 0 {
			s.Cursor = s.CursorMax
		}
		return s, nil

	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandlePanelNav(key, km) {
			return s, nil
		}

		if km.Matches(key, config.ActionTitlePrev) {
			return s, TitleFontPreviousCmd()
		}
		if km.Matches(key, config.ActionTitleNext) {
			return s, TitleFontNextCmd()
		}

		if km.Matches(key, config.ActionSelect) {
			if s.Cursor < len(s.Menu) {
				return s, NavigateToPageCmd(s.Menu[s.Cursor])
			}
		}

		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}
	}
	return s, nil
}

func (s *HomeScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionSelect), "select"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *HomeScreen) View(ctx AppContext) string {
	left := s.renderSystem(ctx)
	center := s.renderNavigation()
	right := s.renderInfo()
	return s.RenderPanels(ctx, left, center, right)
}

func (s *HomeScreen) renderSystem(ctx AppContext) string {
	dbDriver := ""
	if ctx.Config != nil {
		dbDriver = string(ctx.Config.Db_Driver)
	}
	lines := []string{
		"System Info",
		"",
		fmt.Sprintf("  Version:  %s", utility.Version),
		fmt.Sprintf("  Database: %s", dbDriver),
		fmt.Sprintf("  User:     %s", s.Username),
	}
	return strings.Join(lines, "\n")
}

func (s *HomeScreen) renderNavigation() string {
	if len(s.Menu) == 0 {
		return "(no items)"
	}
	lines := make([]string, 0, len(s.Menu))
	for i, item := range s.Menu {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, item.Label))
	}
	return strings.Join(lines, "\n")
}

func (s *HomeScreen) renderInfo() string {
	lines := []string{
		"Keyboard",
		"",
		"  up/down   Navigate",
		"  enter     Select",
		"  tab       Panel",
		"  q         Quit",
	}
	return strings.Join(lines, "\n")
}
