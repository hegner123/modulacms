package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
)

// 3/9 grid: left = menu list, right = detail (top) + info (bottom)
var cmsMenuGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Navigation"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Details"},
			{Height: 0.4, Title: "Info"},
		}},
	},
}

// CMSMenuScreen implements Screen for CMSPAGE and ADMINCMSPAGE.
// It shows a navigable menu list in the left panel.
type CMSMenuScreen struct {
	GridScreen
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
	cursorMax := len(menu) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &CMSMenuScreen{
		GridScreen: GridScreen{
			Grid:      cmsMenuGrid,
			CursorMax: cursorMax,
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

	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
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
	}
}

func (s *CMSMenuScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderMenu(), TotalLines: len(s.Menu), ScrollOffset: ClampScroll(s.Cursor, len(s.Menu), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
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

func (s *CMSMenuScreen) renderDetail() string {
	if len(s.Menu) == 0 || s.Cursor >= len(s.Menu) {
		return " No item selected"
	}
	item := s.Menu[s.Cursor]
	return fmt.Sprintf(" %s", item.Label)
}

func (s *CMSMenuScreen) renderInfo() string {
	label := "CMS"
	if s.IsAdmin {
		label = "Admin CMS"
	}
	lines := []string{
		fmt.Sprintf(" %s Menu", label),
		"",
		fmt.Sprintf(" Items: %d", len(s.Menu)),
	}
	return strings.Join(lines, "\n")
}
