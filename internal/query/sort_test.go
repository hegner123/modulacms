package query

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestParseSort(t *testing.T) {
	tests := []struct {
		input     string
		wantField string
		wantDesc  bool
	}{
		{"", "", false},
		{"title", "title", false},
		{"-title", "title", true},
		{"-published_at", "published_at", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			spec := ParseSort(tt.input)
			if spec.Field != tt.wantField {
				t.Errorf("field = %q, want %q", spec.Field, tt.wantField)
			}
			if spec.Desc != tt.wantDesc {
				t.Errorf("desc = %v, want %v", spec.Desc, tt.wantDesc)
			}
		})
	}
}

func TestCompareForSort_Number(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"5", "10", -1},
		{"10", "5", 1},
		{"10", "10", 0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareForSort(tt.a, tt.b, types.FieldTypeNumber)
			if got != tt.want {
				t.Errorf("compareForSort(%q, %q, number) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCompareForSort_Date(t *testing.T) {
	got := compareForSort("2026-01-10", "2026-01-15", types.FieldTypeDate)
	if got != -1 {
		t.Errorf("date compare: got %d, want -1", got)
	}
	got = compareForSort("2026-01-15", "2026-01-10", types.FieldTypeDate)
	if got != 1 {
		t.Errorf("date compare: got %d, want 1", got)
	}
}

func TestApplySort_MissingValues(t *testing.T) {
	typeIndex := map[string]types.FieldType{
		"views": types.FieldTypeNumber,
	}
	items := []QueryItem{
		{Fields: map[string]string{"views": ""}},
		{Fields: map[string]string{"views": "5"}},
		{Fields: map[string]string{"views": "10"}},
	}
	ApplySort(items, SortSpec{Field: "views"}, typeIndex)

	// Items with values should come first, empty last.
	if items[0].Fields["views"] != "10" && items[0].Fields["views"] != "5" {
		t.Errorf("expected non-empty first, got %q", items[0].Fields["views"])
	}
	if items[2].Fields["views"] != "" {
		t.Errorf("expected empty last, got %q", items[2].Fields["views"])
	}
}

func TestApplySort_Descending(t *testing.T) {
	typeIndex := map[string]types.FieldType{
		"views": types.FieldTypeNumber,
	}
	items := []QueryItem{
		{Fields: map[string]string{"views": "5"}},
		{Fields: map[string]string{"views": "20"}},
		{Fields: map[string]string{"views": "10"}},
	}
	ApplySort(items, SortSpec{Field: "views", Desc: true}, typeIndex)

	if items[0].Fields["views"] != "20" {
		t.Errorf("expected 20 first, got %q", items[0].Fields["views"])
	}
	if items[1].Fields["views"] != "10" {
		t.Errorf("expected 10 second, got %q", items[1].Fields["views"])
	}
	if items[2].Fields["views"] != "5" {
		t.Errorf("expected 5 third, got %q", items[2].Fields["views"])
	}
}
