package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// PanelScrollInfo holds scroll state for a single panel.
type PanelScrollInfo struct {
	TotalLines   int // 0 = no scrollbar
	ScrollOffset int
}

// PanelScreen is the base struct for Screen implementations that use the
// standard 3-panel CMS layout. Embedding it provides RenderPanels and
// HandlePanelNav helpers that replace per-screen duplication of layout
// calculation and panel navigation logic.
type PanelScreen struct {
	Layout     PageLayout
	PanelFocus FocusPanel
	Cursor     int
	CursorMax  int
	TabSets    [3][]PanelTab // tabs per panel (left, center, right)
	ActiveTabs [3]int        // active tab index per panel
}

// HandlePanelNav processes panel navigation keys (tab/shift-tab). Returns true
// if the key was handled.
func (p *PanelScreen) HandlePanelNav(key string, km config.KeyMap) bool {
	if km.Matches(key, config.ActionNextPanel) {
		p.PanelFocus = (p.PanelFocus + 1) % 3
		return true
	}
	if km.Matches(key, config.ActionPrevPanel) {
		p.PanelFocus = (p.PanelFocus + 2) % 3
		return true
	}
	return false
}

// HandleTabNav processes tab cycling keys ([/]) within the focused panel.
// Returns true if the key was handled.
func (p *PanelScreen) HandleTabNav(key string, km config.KeyMap) bool {
	tabs := p.TabSets[p.PanelFocus]
	if len(tabs) < 2 {
		return false
	}
	if km.Matches(key, config.ActionTabPrev) {
		p.ActiveTabs[p.PanelFocus] = (p.ActiveTabs[p.PanelFocus] + len(tabs) - 1) % len(tabs)
		return true
	}
	if km.Matches(key, config.ActionTabNext) {
		p.ActiveTabs[p.PanelFocus] = (p.ActiveTabs[p.PanelFocus] + 1) % len(tabs)
		return true
	}
	return false
}

// RenderPanels renders the three-panel layout with proper ScreenMode, accordion,
// and gutter handling. The left/center/right arguments are the raw content
// strings for each panel; titles come from the embedded Layout. When a panel
// has tabs defined (via TabSets), the active tab's Render function is called
// instead of using the passed content string. Optional scroll info enables
// scroll indicators per panel.
func (p *PanelScreen) RenderPanels(ctx AppContext, left, center, right string, scroll ...PanelScrollInfo) string {
	leftW, centerW, rightW := p.panelWidths(ctx)

	type panelSpec struct {
		title   string
		width   int
		content string
		focus   FocusPanel
		scroll  PanelScrollInfo
	}

	var scrollInfo [3]PanelScrollInfo
	for i := range scroll {
		if i < 3 {
			scrollInfo[i] = scroll[i]
		}
	}

	contents := [3]string{left, center, right}
	specs := []panelSpec{
		{p.Layout.Titles[0], leftW, contents[0], TreePanel, scrollInfo[0]},
		{p.Layout.Titles[1], centerW, contents[1], ContentPanel, scrollInfo[1]},
		{p.Layout.Titles[2], rightW, contents[2], RoutePanel, scrollInfo[2]},
	}

	var panels []string
	for i, s := range specs {
		focused := p.PanelFocus == s.focus
		switch {
		case s.width <= 0:
			continue
		case ctx.ScreenMode == ScreenWide && s.focus != p.PanelFocus:
			panels = append(panels, renderGutterStrip(s.title, ctx.Height, focused))
		default:
			// Use tab content when tabs are defined for this panel
			tabContent := s.content
			tabs := p.TabSets[i]
			activeIdx := p.ActiveTabs[i]
			if len(tabs) > 0 {
				if activeIdx >= len(tabs) {
					activeIdx = 0
				}
				innerH := PanelInnerHeight(ctx.Height)
				if len(tabs) > 1 {
					innerH = PanelInnerHeightWithTabs(ctx.Height)
				}
				tabContent = tabs[activeIdx].Render(ctx, s.width-2, innerH)
			}

			var tabLabels []string
			if len(tabs) > 1 {
				tabLabels = make([]string, len(tabs))
				for j, t := range tabs {
					tabLabels[j] = t.Label
				}
			}

			pan := Panel{
				Title:        s.title,
				Width:        s.width,
				Height:       ctx.Height,
				Content:      tabContent,
				Focused:      focused,
				TotalLines:   s.scroll.TotalLines,
				ScrollOffset: s.scroll.ScrollOffset,
				TabLabels:    tabLabels,
				ActiveTab:    activeIdx,
			}
			panels = append(panels, pan.Render())
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

// panelWidths calculates the left, center, and right panel widths based on
// ScreenMode, Layout ratios, accordion state, and which panel has focus.
func (p *PanelScreen) panelWidths(ctx AppContext) (left, center, right int) {
	switch ctx.ScreenMode {
	case ScreenWide:
		const gutter = 4
		switch p.PanelFocus {
		case TreePanel:
			left = ctx.Width - 2*gutter
			center, right = gutter, gutter
		case ContentPanel:
			left, right = gutter, gutter
			center = ctx.Width - 2*gutter
		case RoutePanel:
			left, center = gutter, gutter
			right = ctx.Width - 2*gutter
		}
	case ScreenFull:
		switch p.PanelFocus {
		case TreePanel:
			left, center, right = ctx.Width, 0, 0
		case ContentPanel:
			left, center, right = 0, ctx.Width, 0
		case RoutePanel:
			left, center, right = 0, 0, ctx.Width
		}
	default: // ScreenNormal
		if ctx.AccordionEnabled {
			const focusFraction = 0.60
			remaining := 1.0 - focusFraction
			widths := [3]int{}

			var otherSum float64
			for i, r := range p.Layout.Ratios {
				if FocusPanel(i) != p.PanelFocus && r > 0 {
					otherSum += r
				}
			}

			used := 0
			for i := range p.Layout.Ratios {
				if p.Layout.Ratios[i] == 0 {
					widths[i] = 0
				} else if FocusPanel(i) == p.PanelFocus {
					widths[i] = int(focusFraction * float64(ctx.Width))
				} else {
					widths[i] = int(remaining * (p.Layout.Ratios[i] / otherSum) * float64(ctx.Width))
				}
				used += widths[i]
			}
			widths[p.PanelFocus] += ctx.Width - used

			left, center, right = widths[0], widths[1], widths[2]
		} else {
			left = int(float64(ctx.Width) * p.Layout.Ratios[0])
			center = int(float64(ctx.Width) * p.Layout.Ratios[1])
			right = ctx.Width - left - center
		}
	}
	return
}
