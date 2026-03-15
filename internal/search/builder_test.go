package search

import (
	"strings"
	"testing"
)

func TestSplitByHeadings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		html             string
		expectedSections []Section
	}{
		{
			name: "two h2 sections",
			html: "<h2>Setup</h2><p>Install it.</p><h2>Usage</h2><p>Run it.</p>",
			expectedSections: []Section{
				{Heading: "Setup", Anchor: "setup"},
				{Heading: "Usage", Anchor: "usage"},
			},
		},
		{
			name:             "no headings",
			html:             "<p>No headings here.</p>",
			expectedSections: nil,
		},
		{
			name: "nested inline tags inside heading",
			html: "<h3>Nested <em>Heading</em></h3><p>Body</p>",
			expectedSections: []Section{
				// StripHTML replaces </em> with a space, producing "Nested  Heading"
				// which TrimSpace does not collapse interior whitespace.
				{Heading: "Nested  Heading", Anchor: "nested-heading"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sections := SplitByHeadings(tt.html)

			if tt.expectedSections == nil {
				if len(sections) != 0 {
					t.Errorf("expected no sections, got %d", len(sections))
				}
				return
			}

			if len(sections) != len(tt.expectedSections) {
				t.Fatalf("expected %d sections, got %d: %+v",
					len(tt.expectedSections), len(sections), sections)
			}

			for i, expected := range tt.expectedSections {
				got := sections[i]
				if got.Heading != expected.Heading {
					t.Errorf("section[%d].Heading = %q, want %q", i, got.Heading, expected.Heading)
				}
				if got.Anchor != expected.Anchor {
					t.Errorf("section[%d].Anchor = %q, want %q", i, got.Anchor, expected.Anchor)
				}
			}
		})
	}
}

func TestSplitByHeadingsSectionBodies(t *testing.T) {
	t.Parallel()

	html := "<h2>Setup</h2><p>Install it.</p><h2>Usage</h2><p>Run it.</p>"
	sections := SplitByHeadings(html)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// Body text should have HTML stripped and contain the paragraph text
	if !strings.Contains(sections[0].Body, "Install it.") {
		t.Errorf("section[0].Body = %q, should contain 'Install it.'", sections[0].Body)
	}
	if !strings.Contains(sections[1].Body, "Run it.") {
		t.Errorf("section[1].Body = %q, should contain 'Run it.'", sections[1].Body)
	}
}

func TestSlugify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Setup",
			expected: "setup",
		},
		{
			name:     "multiple words",
			input:    "Nested Heading",
			expected: "nested-heading",
		},
		{
			name:     "special characters",
			input:    "What's New?",
			expected: "what-s-new",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slugify(tt.input)
			if got != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
