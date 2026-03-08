package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// ScreenMode determines the panel layout mode.
type ScreenMode int

const (
	ScreenNormal ScreenMode = iota // grid: proportional column split
	ScreenFull                     // 1 panel: focused cell takes 100%
)

// String returns the display name of the screen mode.
func (s ScreenMode) String() string {
	switch s {
	case ScreenNormal:
		return "Normal"
	case ScreenFull:
		return "Full"
	default:
		return "Unknown"
	}
}

// Panel represents a bordered UI section with a title.
type Panel struct {
	Title        string
	Width        int
	Height       int
	Content      string
	Focused      bool
	TotalLines   int                            // total content lines; 0 = no scrollbar
	ScrollOffset int                            // first visible line index
	TabLabels    []string                       // tab labels for tab bar (shown when len > 1)
	ActiveTab    int                            // index of the currently active tab
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

	// Truncate or pad each line to exact visual width (accounts for ANSI escapes)
	for i, line := range lines {
		visWidth := lipgloss.Width(line)
		if visWidth > width {
			lines[i] = truncateToVisualWidth(line, width)
		} else if visWidth < width {
			lines[i] = line + strings.Repeat(" ", width-visWidth)
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

// truncateToVisualWidth truncates a string (possibly containing ANSI escape
// codes) to the given visual width. It walks rune-by-rune, skipping ANSI
// sequences, and stops once the visible width reaches the limit.
func truncateToVisualWidth(s string, maxWidth int) string {
	var visWidth int
	var result []byte
	bytes := []byte(s)
	i := 0
	for i < len(bytes) {
		// Skip ANSI escape sequences (ESC [ ... final byte)
		if bytes[i] == 0x1b && i+1 < len(bytes) && bytes[i+1] == '[' {
			start := i
			i += 2
			for i < len(bytes) && bytes[i] >= 0x20 && bytes[i] <= 0x3f {
				i++
			}
			if i < len(bytes) {
				i++ // consume final byte
			}
			result = append(result, bytes[start:i]...)
			continue
		}
		// Decode one rune
		r, size := rune(bytes[i]), 1
		if bytes[i] >= 0x80 {
			var n int
			r, n = decodeRune(bytes[i:])
			size = n
		}
		rw := runeWidth(r)
		if visWidth+rw > maxWidth {
			break
		}
		visWidth += rw
		result = append(result, bytes[i:i+size]...)
		i += size
	}
	return string(result)
}

// decodeRune decodes a single UTF-8 rune from the byte slice.
func decodeRune(b []byte) (rune, int) {
	if len(b) == 0 {
		return 0, 0
	}
	r := rune(b[0])
	switch {
	case r < 0x80:
		return r, 1
	case r < 0xC0:
		return r, 1
	case r < 0xE0 && len(b) >= 2:
		return rune(b[0]&0x1F)<<6 | rune(b[1]&0x3F), 2
	case r < 0xF0 && len(b) >= 3:
		return rune(b[0]&0x0F)<<12 | rune(b[1]&0x3F)<<6 | rune(b[2]&0x3F), 3
	case r < 0xF8 && len(b) >= 4:
		return rune(b[0]&0x07)<<18 | rune(b[1]&0x3F)<<12 | rune(b[2]&0x3F)<<6 | rune(b[3]&0x3F), 4
	default:
		return r, 1
	}
}

// runeWidth returns the visual column width of a rune (1 for most, 2 for CJK).
func runeWidth(r rune) int {
	if r >= 0x1100 &&
		(r <= 0x115F || r == 0x2329 || r == 0x232A ||
			(r >= 0x2E80 && r <= 0x303E) ||
			(r >= 0x3040 && r <= 0x33BF) ||
			(r >= 0x3400 && r <= 0x4DBF) ||
			(r >= 0x4E00 && r <= 0xA4CF) ||
			(r >= 0xA960 && r <= 0xA97C) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFE10 && r <= 0xFE6B) ||
			(r >= 0xFF01 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6) ||
			(r >= 0x1F000 && r <= 0x1FAFF) ||
			(r >= 0x20000 && r <= 0x2FA1F)) {
		return 2
	}
	return 1
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
