package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Overlay represents a positioned content rectangle to render on top of a base layer.
type Overlay struct {
	Content string // Pre-rendered content for this layer
	X       int    // Left column origin (0-based)
	Y       int    // Top row origin (0-based)
	Width   int    // Width in columns
	Height  int    // Height in rows
}

// Composite renders the overlay on top of the base string buffer.
// Rows outside the overlay's Y range pass through unchanged.
// Rows inside the overlay's Y range have columns [X, X+Width) replaced
// with the corresponding overlay line.
func Composite(base string, overlay Overlay) string {
	baseLines := strings.Split(base, "\n")
	topLines := strings.Split(overlay.Content, "\n")

	// Extend base if overlay extends beyond it
	needed := overlay.Y + overlay.Height
	for len(baseLines) < needed {
		baseLines = append(baseLines, "")
	}

	result := make([]string, len(baseLines))
	for y, baseLine := range baseLines {
		if y >= overlay.Y && y < overlay.Y+overlay.Height {
			topIdx := y - overlay.Y
			if topIdx < len(topLines) {
				result[y] = spliceLine(baseLine, topLines[topIdx], overlay.X, overlay.Width)
			} else {
				result[y] = baseLine
			}
		} else {
			result[y] = baseLine
		}
	}
	return strings.Join(result, "\n")
}

// spliceLine replaces columns [x, x+width) of base with top content.
func spliceLine(base, top string, x, width int) string {
	// Left: base content up to column x
	left := ansi.Truncate(base, x, "")
	leftW := ansi.StringWidth(left)
	if leftW < x {
		left += strings.Repeat(" ", x-leftW)
	}

	// Middle: top content, padded or truncated to exactly `width` columns
	middle := ansi.Truncate(top, width, "")
	middleW := ansi.StringWidth(middle)
	if middleW < width {
		middle += strings.Repeat(" ", width-middleW)
	}

	// Right: base content after column x+width
	right := ansi.TruncateLeft(base, x+width, "")

	return left + middle + right
}
