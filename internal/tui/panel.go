package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// PageLayout declares the default panel configuration for a page.
type PageLayout struct {
	Panels int        // 1, 2, or 3 visible panels
	Ratios [3]float64 // {left, center, right} proportions (must sum to 1.0)
	Titles [3]string  // panel titles
}

// ScreenMode determines the panel layout mode.
type ScreenMode int

const (
	ScreenNormal ScreenMode = iota // 3 panels: proportional split
	ScreenWide                     // 2 panels: focused + gutters
	ScreenFull                     // 1 panel: focused takes 100%
)

// String returns the display name of the screen mode.
func (s ScreenMode) String() string {
	switch s {
	case ScreenNormal:
		return "Normal"
	case ScreenWide:
		return "Wide"
	case ScreenFull:
		return "Full"
	default:
		return "Unknown"
	}
}

// FocusPanel identifies which panel currently has focus.
type FocusPanel int

const (
	TreePanel FocusPanel = iota
	ContentPanel
	RoutePanel
)

// String returns the display name of the focused panel.
func (f FocusPanel) String() string {
	switch f {
	case TreePanel:
		return "Tree"
	case ContentPanel:
		return "Content"
	case RoutePanel:
		return "Route"
	default:
		return "Unknown"
	}
}

// PanelTab describes a single tab within a panel. The Render function is a
// closure that captures the owning screen's state so it can render the tab
// content with access to cursors, loaded data, etc.
type PanelTab struct {
	Label  string
	Render func(ctx AppContext, w, h int) string
}

// Panel represents a bordered UI section with a title.
type Panel struct {
	Title        string
	Width        int
	Height       int
	Content      string
	Focused      bool
	TotalLines   int      // total content lines; 0 = no scrollbar
	ScrollOffset int      // first visible line index
	TabLabels    []string // tab labels for tab bar (shown when len > 1)
	ActiveTab    int      // index of the currently active tab
	Accent       lipgloss.CompleteAdaptiveColor // override accent; zero value uses DefaultStyle.Accent
}

// Render draws the panel as a bordered box with a title bar.
// When TotalLines > 0 and exceeds the inner height, a scrollbar track is
// rendered in the rightmost column and a position indicator is appended
// to the title.
func (p Panel) Render() string {
	accent := config.DefaultStyle.Accent
	if p.Accent != (lipgloss.CompleteAdaptiveColor{}) {
		accent = p.Accent
	}

	borderColor := config.DefaultStyle.Tertiary
	if p.Focused {
		borderColor = accent
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent)

	// Inner width is panel width minus border columns (2 chars for left+right border)
	innerWidth := p.Width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	// Inner height is panel height minus border rows (2 for top+bottom) minus title row
	innerHeight := p.Height - 3
	if innerHeight < 0 {
		innerHeight = 0
	}

	showTabs := len(p.TabLabels) > 1
	if showTabs && innerHeight > 0 {
		innerHeight-- // tab bar consumes 1 line
	}

	showScrollbar := p.TotalLines > 0 && p.TotalLines > innerHeight && innerHeight > 0

	contentWidth := innerWidth
	if showScrollbar && contentWidth > 1 {
		contentWidth-- // reserve 1 column for scrollbar
	}

	// Build title line with optional position indicator
	title := titleStyle.Render(p.Title)
	if showScrollbar {
		posStyle := lipgloss.NewStyle().Faint(true)
		title += posStyle.Render(fmt.Sprintf(" %d/%d", p.ScrollOffset+1, p.TotalLines))
	}

	// Slice and pad content to fill the inner area
	content := padContentWithScroll(p.Content, contentWidth, innerHeight, p.ScrollOffset, showScrollbar)

	if showScrollbar {
		content = appendScrollbar(content, innerHeight, p.TotalLines, p.ScrollOffset)
	}

	tabBar := ""
	if showTabs {
		tabBar = renderTabBar(p.TabLabels, p.ActiveTab, innerWidth) + "\n"
	}

	body := title + "\n" + tabBar + content

	boxHeight := innerHeight + 1 // +1 for the title line
	if showTabs {
		boxHeight++ // +1 for the tab bar line
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(boxHeight).
		Render(body)

	return box
}

// padContent ensures content fills exactly the given width and height.
func padContent(content string, width, height int) string {
	return padContentWithScroll(content, width, height, 0, false)
}

// padContentWithScroll slices content starting at scrollOffset (when scrolling
// is active) and pads/truncates to fill the given width and height.
func padContentWithScroll(content string, width, height, scrollOffset int, scrolling bool) string {
	lines := strings.Split(content, "\n")

	// Apply scroll offset when scrolling is active
	if scrolling && scrollOffset > 0 {
		if scrollOffset >= len(lines) {
			lines = []string{}
		} else {
			lines = lines[scrollOffset:]
		}
	}

	// Truncate or pad lines to fill height
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	// Truncate each line to width
	for i, line := range lines {
		runeCount := len([]rune(line))
		if runeCount > width {
			lines[i] = string([]rune(line)[:width])
		}
	}

	return strings.Join(lines, "\n")
}

// appendScrollbar adds a scrollbar track character to the right edge of each
// content line. Uses "▓" for the thumb and "░" for the track.
func appendScrollbar(content string, viewportHeight, totalLines, scrollOffset int) string {
	lines := strings.Split(content, "\n")

	thumbHeight := viewportHeight * viewportHeight / totalLines
	if thumbHeight < 1 {
		thumbHeight = 1
	}

	maxScroll := totalLines - viewportHeight
	if maxScroll < 1 {
		maxScroll = 1
	}

	thumbStart := scrollOffset * (viewportHeight - thumbHeight) / maxScroll

	trackStyle := lipgloss.NewStyle().Faint(true)

	for i := range lines {
		if i >= thumbStart && i < thumbStart+thumbHeight {
			lines[i] += "▓"
		} else {
			lines[i] += trackStyle.Render("░")
		}
	}

	return strings.Join(lines, "\n")
}

// PanelInnerHeight returns the content area height inside a panel.
func PanelInnerHeight(panelHeight int) int {
	h := panelHeight - 3 // 2 border rows + 1 title row
	if h < 0 {
		return 0
	}
	return h
}

// PanelInnerHeightWithTabs returns the content area height inside a panel
// that has a visible tab bar (len(tabs) > 1).
func PanelInnerHeightWithTabs(panelHeight int) int {
	h := PanelInnerHeight(panelHeight)
	if h > 0 {
		h--
	}
	return h
}

// renderTabBar builds a single-line tab bar from the given labels, highlighting
// the active tab with the accent color. The bar is truncated to fit maxWidth.
func renderTabBar(labels []string, active, maxWidth int) string {
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent).
		Underline(true)
	inactiveStyle := lipgloss.NewStyle().
		Faint(true)
	sepStyle := lipgloss.NewStyle().Faint(true)

	var parts []string
	for i, label := range labels {
		if i > 0 {
			parts = append(parts, sepStyle.Render(" | "))
		}
		if i == active {
			parts = append(parts, activeStyle.Render(label))
		} else {
			parts = append(parts, inactiveStyle.Render(label))
		}
	}
	return strings.Join(parts, "")
}

// ClampScroll computes a scroll offset that keeps the cursor visible
// within a viewport of the given height.
func ClampScroll(cursor, totalItems, viewportHeight int) int {
	if totalItems <= viewportHeight {
		return 0
	}
	offset := cursor - viewportHeight/2
	if offset < 0 {
		offset = 0
	}
	maxOffset := totalItems - viewportHeight
	if offset > maxOffset {
		offset = maxOffset
	}
	return offset
}
