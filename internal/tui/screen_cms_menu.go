package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// CMSMenuScreen implements Screen for CMSPAGE and ADMINCMSPAGE.
// It shows a navigable menu list in the left panel.
type CMSMenuScreen struct {
	PanelScreen
	Menu    []Page
	IsAdmin bool
	pageIdx PageIndex
}

// NewCMSMenuScreen creates a CMSMenuScreen.
func NewCMSMenuScreen(isAdmin bool, menu []Page) *CMSMenuScreen {
	idx := CMSPAGE
	if isAdmin {
		idx = ADMINCMSPAGE
	}
	return &CMSMenuScreen{
		PanelScreen: PanelScreen{
			Layout:     layoutForPage(idx),
			PanelFocus: TreePanel,
			CursorMax:  len(menu) - 1,
		},
		Menu:    menu,
		IsAdmin: isAdmin,
		pageIdx: idx,
	}
}

func (s *CMSMenuScreen) PageIndex() PageIndex { return s.pageIdx }

func (s *CMSMenuScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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

func (s *CMSMenuScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionSelect), "select"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *CMSMenuScreen) View(ctx AppContext) string {
	left := s.renderMenu()
	center := "Select an item"
	right := "Route\n\n  (none)"
	return s.RenderPanels(ctx, left, center, right)
}

func (s *CMSMenuScreen) renderMenu() string {
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
