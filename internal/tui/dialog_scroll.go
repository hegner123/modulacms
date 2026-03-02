package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// ScrollState tracks the vertical scroll offset for dialogs whose body
// can exceed the available terminal height.
type ScrollState struct {
	offset int
}

// scrollableBody takes a list of rendered items (one per logical entry,
// e.g. one field label+input), the index of the currently focused item,
// and the maximum number of visible lines. It adjusts the scroll offset
// so the focused item is visible, then returns the visible slice of
// lines plus booleans indicating whether content is clipped above/below.
//
// If the total content fits within maxLines, everything is returned
// unchanged and both clipped flags are false.
func (s *ScrollState) scrollableBody(items []string, focusIdx int, maxLines int) (visible string, topClipped bool, bottomClipped bool) {
	if len(items) == 0 || maxLines <= 0 {
		return "", false, false
	}

	// Measure each item's rendered height and build a span table.
	// spans[i] = {startLine, endLine} (exclusive end).
	type span struct {
		start int
		end   int
	}
	spans := make([]span, len(items))
	totalLines := 0
	for i, item := range items {
		h := lipgloss.Height(item)
		if h < 1 {
			h = 1
		}
		spans[i] = span{start: totalLines, end: totalLines + h}
		totalLines += h
	}

	// No scrolling needed — everything fits.
	if totalLines <= maxLines {
		s.offset = 0
		return strings.Join(items, "\n"), false, false
	}

	// Follow focus: ensure the focused item is fully visible.
	if focusIdx >= 0 && focusIdx < len(spans) {
		fs := spans[focusIdx]
		// If focused item starts above viewport, scroll up to it.
		if fs.start < s.offset {
			s.offset = fs.start
		}
		// If focused item ends below viewport, scroll down.
		if fs.end > s.offset+maxLines {
			s.offset = fs.end - maxLines
		}
	}

	// Clamp offset to valid range.
	maxOffset := totalLines - maxLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.offset < 0 {
		s.offset = 0
	}
	if s.offset > maxOffset {
		s.offset = maxOffset
	}

	// Join all items, split into individual lines, slice the viewport.
	allLines := strings.Split(strings.Join(items, "\n"), "\n")
	end := s.offset + maxLines
	if end > len(allLines) {
		end = len(allLines)
	}
	visibleLines := allLines[s.offset:end]

	return strings.Join(visibleLines, "\n"), s.offset > 0, end < len(allLines)
}

// scrollUpIndicator returns a centered "▲ more" hint styled with the
// tertiary color. width is the available content width.
func scrollUpIndicator(width int) string {
	style := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Italic(true).
		Width(width).
		Align(lipgloss.Center)
	return style.Render("▲ more")
}

// scrollDownIndicator returns a centered "▼ more" hint styled with the
// tertiary color. width is the available content width.
func scrollDownIndicator(width int) string {
	style := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Italic(true).
		Width(width).
		Align(lipgloss.Center)
	return style.Render("▼ more")
}
