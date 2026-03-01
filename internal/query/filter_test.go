package query

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestParseFilterKey(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		wantOp  FilterOp
	}{
		{"title", "title", OpEq},
		{"title[eq]", "title", OpEq},
		{"views[gt]", "views", OpGt},
		{"views[gte]", "views", OpGte},
		{"views[lt]", "views", OpLt},
		{"views[lte]", "views", OpLte},
		{"name[neq]", "name", OpNeq},
		{"title[like]", "title", OpLike},
		{"field[banana]", "field", OpEq},
		{"noclose[eq", "noclose[eq", OpEq},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, op := parseFilterKey(tt.input)
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
			if op != tt.wantOp {
				t.Errorf("op = %q, want %q", op, tt.wantOp)
			}
		})
	}
}

func TestParseFilters(t *testing.T) {
	params := map[string][]string{
		"title":     {"hello"},
		"sort":      {"-date"},
		"limit":     {"10"},
		"views[gt]": {"100"},
		"status":    {"published"},
	}
	filters := ParseFilters(params)

	// Should have title and views[gt] but not sort, limit, status (reserved).
	foundTitle := false
	foundViews := false
	for _, f := range filters {
		switch f.Field {
		case "title":
			foundTitle = true
			if f.Operator != OpEq || f.Value != "hello" {
				t.Errorf("title filter: op=%q val=%q", f.Operator, f.Value)
			}
		case "views":
			foundViews = true
			if f.Operator != OpGt || f.Value != "100" {
				t.Errorf("views filter: op=%q val=%q", f.Operator, f.Value)
			}
		case "sort", "limit", "status":
			t.Errorf("reserved key %q should not appear as filter", f.Field)
		}
	}
	if !foundTitle {
		t.Error("expected title filter")
	}
	if !foundViews {
		t.Error("expected views filter")
	}
}

func TestCompareFieldValue_Number(t *testing.T) {
	tests := []struct {
		a, b string
		op   FilterOp
		want bool
	}{
		{"10", "5", OpGt, true},
		{"5", "10", OpGt, false},
		{"10", "10", OpEq, true},
		{"10", "10", OpGte, true},
		{"5", "10", OpLt, true},
		{"5", "5", OpLte, true},
		{"10", "5", OpNeq, true},
		{"10", "10", OpNeq, false},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+string(tt.op)+"_"+tt.b, func(t *testing.T) {
			got := compareFieldValue(tt.a, tt.b, tt.op, types.FieldTypeNumber)
			if got != tt.want {
				t.Errorf("compareFieldValue(%q, %q, %q, number) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}

func TestCompareFieldValue_Boolean(t *testing.T) {
	tests := []struct {
		a, b string
		op   FilterOp
		want bool
	}{
		{"true", "true", OpEq, true},
		{"1", "true", OpEq, true},
		{"false", "0", OpEq, true},
		{"true", "false", OpEq, false},
		{"true", "false", OpNeq, true},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+string(tt.op)+"_"+tt.b, func(t *testing.T) {
			got := compareFieldValue(tt.a, tt.b, tt.op, types.FieldTypeBoolean)
			if got != tt.want {
				t.Errorf("compareFieldValue(%q, %q, %q, boolean) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}

func TestCompareFieldValue_Date(t *testing.T) {
	tests := []struct {
		a, b string
		op   FilterOp
		want bool
	}{
		{"2026-01-15", "2026-01-10", OpGt, true},
		{"2026-01-10", "2026-01-15", OpLt, true},
		{"2026-01-15", "2026-01-15", OpEq, true},
		{"2026-01-15T00:00:00Z", "2026-01-15", OpEq, true},
		{"2026-01-15", "2026-01-10", OpGte, true},
		{"2026-01-10", "2026-01-15", OpLte, true},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+string(tt.op)+"_"+tt.b, func(t *testing.T) {
			got := compareFieldValue(tt.a, tt.b, tt.op, types.FieldTypeDate)
			if got != tt.want {
				t.Errorf("compareFieldValue(%q, %q, %q, date) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}

func TestCompareFieldValue_String(t *testing.T) {
	tests := []struct {
		a, b string
		op   FilterOp
		want bool
	}{
		{"hello", "hello", OpEq, true},
		{"hello", "world", OpEq, false},
		{"hello", "world", OpNeq, true},
		{"Hello World", "world", OpLike, true},
		{"Hello", "HELLO", OpLike, true},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+string(tt.op)+"_"+tt.b, func(t *testing.T) {
			got := compareFieldValue(tt.a, tt.b, tt.op, types.FieldTypeText)
			if got != tt.want {
				t.Errorf("compareFieldValue(%q, %q, %q, text) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}
