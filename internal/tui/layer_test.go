package tui

import (
	"strings"
	"testing"
)

// ============================================================
// spliceLine
// ============================================================

func TestSpliceLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		base  string
		top   string
		x     int
		width int
		want  string
	}{
		{
			name:  "replace middle of line",
			base:  "AAAAAAAAAA",
			top:   "BBB",
			x:     3,
			width: 3,
			want:  "AAABBBAAAA",
		},
		{
			name:  "replace at start",
			base:  "AAAAAAAAAA",
			top:   "XX",
			x:     0,
			width: 2,
			want:  "XXAAAAAAAA",
		},
		{
			name:  "replace at end",
			base:  "AAAAA",
			top:   "ZZ",
			x:     3,
			width: 2,
			want:  "AAAZZ",
		},
		{
			name:  "top shorter than width pads with spaces",
			base:  "AAAAAAAAAA",
			top:   "B",
			x:     2,
			width: 4,
			want:  "AAB   AAAA",
		},
		{
			name:  "top longer than width gets truncated",
			base:  "AAAAAAAAAA",
			top:   "BBBBBBB",
			x:     2,
			width: 3,
			want:  "AABBBAAAAA",
		},
		{
			name:  "base shorter than x pads left",
			base:  "AA",
			top:   "BB",
			x:     5,
			width: 2,
			want:  "AA   BB",
		},
		{
			name:  "empty base",
			base:  "",
			top:   "HI",
			x:     3,
			width: 2,
			want:  "   HI",
		},
		{
			name:  "empty top fills with spaces",
			base:  "AAAAAAAAAA",
			top:   "",
			x:     2,
			width: 3,
			want:  "AA   AAAAA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := spliceLine(tt.base, tt.top, tt.x, tt.width)
			if got != tt.want {
				t.Errorf("spliceLine(%q, %q, %d, %d) = %q, want %q",
					tt.base, tt.top, tt.x, tt.width, got, tt.want)
			}
		})
	}
}

// ============================================================
// Composite
// ============================================================

func TestComposite(t *testing.T) {
	t.Parallel()

	t.Run("overlay in center", func(t *testing.T) {
		t.Parallel()
		base := "AAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA"
		overlay := Overlay{
			Content: "BB\nBB",
			X:       2,
			Y:       1,
			Width:   2,
			Height:  2,
		}
		got := Composite(base, overlay)
		lines := strings.Split(got, "\n")

		if lines[0] != "AAAAAAAAAA" {
			t.Errorf("line 0 unchanged: got %q", lines[0])
		}
		if lines[1] != "AABBAAAAAA" {
			t.Errorf("line 1: got %q, want %q", lines[1], "AABBAAAAAA")
		}
		if lines[2] != "AABBAAAAAA" {
			t.Errorf("line 2: got %q, want %q", lines[2], "AABBAAAAAA")
		}
		if lines[3] != "AAAAAAAAAA" {
			t.Errorf("line 3 unchanged: got %q", lines[3])
		}
	})

	t.Run("overlay extends past base height", func(t *testing.T) {
		t.Parallel()
		base := "AAA"
		overlay := Overlay{
			Content: "BB\nBB\nBB",
			X:       0,
			Y:       0,
			Width:   2,
			Height:  3,
		}
		got := Composite(base, overlay)
		lines := strings.Split(got, "\n")

		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}
		if lines[0] != "BBA" {
			t.Errorf("line 0: got %q, want %q", lines[0], "BBA")
		}
	})

	t.Run("overlay at origin", func(t *testing.T) {
		t.Parallel()
		base := "XXXXX\nXXXXX"
		overlay := Overlay{
			Content: "OO\nOO",
			X:       0,
			Y:       0,
			Width:   2,
			Height:  2,
		}
		got := Composite(base, overlay)
		lines := strings.Split(got, "\n")

		if lines[0] != "OOXXX" {
			t.Errorf("line 0: got %q, want %q", lines[0], "OOXXX")
		}
		if lines[1] != "OOXXX" {
			t.Errorf("line 1: got %q, want %q", lines[1], "OOXXX")
		}
	})

	t.Run("empty overlay content", func(t *testing.T) {
		t.Parallel()
		base := "AAAAA\nAAAAA"
		overlay := Overlay{
			Content: "",
			X:       1,
			Y:       0,
			Width:   3,
			Height:  2,
		}
		got := Composite(base, overlay)
		lines := strings.Split(got, "\n")

		// First overlay line is empty string → padded to 3 spaces
		if lines[0] != "A   A" {
			t.Errorf("line 0: got %q, want %q", lines[0], "A   A")
		}
		// Second overlay line is out of range of topLines → baseLine unchanged
		if lines[1] != "AAAAA" {
			t.Errorf("line 1: got %q, want %q", lines[1], "AAAAA")
		}
	})
}

// ============================================================
// runeWidth
// ============================================================

func TestRuneWidth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{name: "ASCII letter", r: 'A', want: 1},
		{name: "ASCII space", r: ' ', want: 1},
		{name: "Latin accent", r: 'é', want: 1},
		{name: "CJK ideograph", r: '中', want: 2},
		{name: "Hangul syllable", r: '한', want: 2},
		{name: "Katakana", r: 'カ', want: 2},
		{name: "Fullwidth exclamation", r: '！', want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := runeWidth(tt.r)
			if got != tt.want {
				t.Errorf("runeWidth(%q) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

// ============================================================
// truncateToVisualWidth
// ============================================================

func TestTruncateToVisualWidth(t *testing.T) {
	t.Parallel()

	t.Run("ASCII within limit", func(t *testing.T) {
		t.Parallel()
		got := truncateToVisualWidth("hello", 10)
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("ASCII truncated", func(t *testing.T) {
		t.Parallel()
		got := truncateToVisualWidth("hello world", 5)
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("CJK truncated at boundary", func(t *testing.T) {
		t.Parallel()
		// Each CJK char is 2 wide. "中文字" = 6 visual width.
		got := truncateToVisualWidth("中文字", 4)
		if got != "中文" {
			t.Errorf("got %q, want %q", got, "中文")
		}
	})

	t.Run("CJK truncated mid-char", func(t *testing.T) {
		t.Parallel()
		// Width 3 can't fit "中文" (4 wide), so only "中" (2 wide) fits
		got := truncateToVisualWidth("中文字", 3)
		if got != "中" {
			t.Errorf("got %q, want %q", got, "中")
		}
	})

	t.Run("ANSI escapes preserved", func(t *testing.T) {
		t.Parallel()
		// "\033[31m" is red color, "\033[0m" is reset — both are zero-width
		input := "\033[31mhello\033[0m"
		got := truncateToVisualWidth(input, 3)
		if got != "\033[31mhel" {
			t.Errorf("got %q, want %q", got, "\033[31mhel")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		got := truncateToVisualWidth("", 5)
		if got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})

	t.Run("zero width", func(t *testing.T) {
		t.Parallel()
		got := truncateToVisualWidth("hello", 0)
		if got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})
}
