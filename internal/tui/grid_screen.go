package tui

import (
	"github.com/hegner123/modulacms/internal/config"
)

// GridScreen is the base struct for Screen implementations using the
// 12-column grid layout. Replaces PanelScreen for new and migrated screens.
type GridScreen struct {
	Grid       Grid
	FocusIndex int
	Cursor     int
	CursorMax  int
}

// HandleFocusNav processes cell focus cycling (tab/shift-tab).
// Returns true if the key was handled.
func (g *GridScreen) HandleFocusNav(key string, km config.KeyMap) bool {
	cellCount := g.Grid.CellCount()
	if cellCount < 2 {
		return false
	}
	if km.Matches(key, config.ActionNextPanel) {
		g.FocusIndex = (g.FocusIndex + 1) % cellCount
		return true
	}
	if km.Matches(key, config.ActionPrevPanel) {
		g.FocusIndex = (g.FocusIndex + cellCount - 1) % cellCount
		return true
	}
	return false
}

// RenderGrid renders the grid with the given cell contents.
// In ScreenFull mode, only the focused cell is rendered at full size.
// The active accent color from AppContext is passed through to all panels.
func (g *GridScreen) RenderGrid(ctx AppContext, cells []CellContent) string {
	if ctx.ScreenMode == ScreenFull {
		var cc CellContent
		if g.FocusIndex < len(cells) {
			cc = cells[g.FocusIndex]
		}
		pan := Panel{
			Title:        g.Grid.CellTitle(g.FocusIndex),
			Width:        ctx.Width,
			Height:       ctx.Height,
			Content:      cc.Content,
			Focused:      true,
			TotalLines:   cc.TotalLines,
			ScrollOffset: cc.ScrollOffset,
			Accent:       ctx.ActiveAccent,
		}
		return pan.Render()
	}
	return g.Grid.Render(cells, ctx.Width, ctx.Height, g.FocusIndex, ctx.ActiveAccent)
}
