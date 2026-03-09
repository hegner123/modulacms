package tui

import (
	"strings"
	"testing"
)

// ============================================================
// PanelInnerHeight
// ============================================================

func TestPanelInnerHeight(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		panelHeight int
		want        int
	}{
		{name: "standard panel", panelHeight: 20, want: 17},
		{name: "minimum usable", panelHeight: 4, want: 1},
		{name: "exactly 3", panelHeight: 3, want: 0},
		{name: "too small", panelHeight: 2, want: 0},
		{name: "zero height", panelHeight: 0, want: 0},
		{name: "negative height", panelHeight: -5, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := PanelInnerHeight(tt.panelHeight)
			if got != tt.want {
				t.Errorf("PanelInnerHeight(%d) = %d, want %d", tt.panelHeight, got, tt.want)
			}
		})
	}
}

// ============================================================
// ClampScroll
// ============================================================

func TestClampScroll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		cursor         int
		totalItems     int
		viewportHeight int
		want           int
	}{
		{name: "items fit in viewport", cursor: 3, totalItems: 5, viewportHeight: 10, want: 0},
		{name: "cursor at start", cursor: 0, totalItems: 100, viewportHeight: 10, want: 0},
		{name: "cursor in middle", cursor: 50, totalItems: 100, viewportHeight: 10, want: 45},
		{name: "cursor near end", cursor: 98, totalItems: 100, viewportHeight: 10, want: 90},
		{name: "cursor at end", cursor: 99, totalItems: 100, viewportHeight: 10, want: 90},
		{name: "single item", cursor: 0, totalItems: 1, viewportHeight: 10, want: 0},
		{name: "viewport equals total", cursor: 5, totalItems: 10, viewportHeight: 10, want: 0},
		{name: "viewport one less than total", cursor: 9, totalItems: 10, viewportHeight: 9, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ClampScroll(tt.cursor, tt.totalItems, tt.viewportHeight)
			if got != tt.want {
				t.Errorf("ClampScroll(%d, %d, %d) = %d, want %d",
					tt.cursor, tt.totalItems, tt.viewportHeight, got, tt.want)
			}
		})
	}
}

// ============================================================
// ScreenMode.String
// ============================================================

func TestScreenMode_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mode ScreenMode
		want string
	}{
		{ScreenNormal, "Normal"},
		{ScreenFull, "Full"},
		{ScreenMode(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("ScreenMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

// ============================================================
// padContent
// ============================================================

func TestPadContent(t *testing.T) {
	t.Parallel()

	t.Run("pads short content to fill height", func(t *testing.T) {
		t.Parallel()
		got := padContent("line1\nline2", 10, 4)
		lines := strings.Split(got, "\n")
		if len(lines) != 4 {
			t.Fatalf("expected 4 lines, got %d", len(lines))
		}
	})

	t.Run("truncates tall content to height", func(t *testing.T) {
		t.Parallel()
		content := "a\nb\nc\nd\ne\nf"
		got := padContent(content, 5, 3)
		lines := strings.Split(got, "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}
	})

	t.Run("pads narrow lines to width", func(t *testing.T) {
		t.Parallel()
		got := padContent("hi", 8, 1)
		lines := strings.Split(got, "\n")
		// "hi" + 6 spaces = 8 chars
		if len(lines[0]) != 8 {
			t.Errorf("line width = %d, want 8", len(lines[0]))
		}
	})

	t.Run("empty content fills with blank lines", func(t *testing.T) {
		t.Parallel()
		got := padContent("", 5, 3)
		lines := strings.Split(got, "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}
		for i, line := range lines {
			if len(line) != 5 {
				t.Errorf("line %d width = %d, want 5", i, len(line))
			}
		}
	})

	t.Run("zero dimensions returns empty", func(t *testing.T) {
		t.Parallel()
		got := padContent("hello", 0, 0)
		// height 0 → should truncate to 0 lines, but strings.Split on empty gives [""]
		// The function joins with \n, so we check for empty or minimal output
		lines := strings.Split(got, "\n")
		if len(lines) > 0 && lines[0] != "" {
			t.Errorf("expected empty output for zero dimensions, got %q", got)
		}
	})
}

// ============================================================
// padContentWithScroll
// ============================================================

func TestPadContentWithScroll(t *testing.T) {
	t.Parallel()

	t.Run("scroll offset skips lines", func(t *testing.T) {
		t.Parallel()
		content := "line0\nline1\nline2\nline3\nline4"
		got := padContentWithScroll(content, 10, 2, 2, true)
		lines := strings.Split(got, "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		// After scrolling by 2, first visible line should start with "line2"
		if !strings.HasPrefix(lines[0], "line2") {
			t.Errorf("first visible line = %q, want prefix %q", lines[0], "line2")
		}
	})

	t.Run("scroll offset past end gives blank lines", func(t *testing.T) {
		t.Parallel()
		got := padContentWithScroll("a\nb", 5, 2, 10, true)
		lines := strings.Split(got, "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		// All lines should be blank (padded spaces)
		for i, line := range lines {
			if strings.TrimRight(line, " ") != "" {
				t.Errorf("line %d should be blank, got %q", i, line)
			}
		}
	})

	t.Run("no scroll when scrolling is false", func(t *testing.T) {
		t.Parallel()
		content := "line0\nline1\nline2"
		got := padContentWithScroll(content, 10, 2, 2, false)
		lines := strings.Split(got, "\n")
		// scrolling=false ignores scrollOffset, so first line is "line0"
		if !strings.HasPrefix(lines[0], "line0") {
			t.Errorf("first line = %q, want prefix %q", lines[0], "line0")
		}
	})
}
