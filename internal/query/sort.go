package query

import (
	"slices"
	"strconv"
	"strings"

	"github.com/hegner123/modulacms/internal/db/types"
)

// SortSpec describes a sort directive: field name and direction.
type SortSpec struct {
	Field string
	Desc  bool
}

// ParseSort parses a sort parameter string into a SortSpec.
// A leading "-" indicates descending order. Empty string means no sort.
func ParseSort(s string) SortSpec {
	if s == "" {
		return SortSpec{}
	}
	if strings.HasPrefix(s, "-") {
		return SortSpec{Field: s[1:], Desc: true}
	}
	return SortSpec{Field: s, Desc: false}
}

// ApplySort sorts items in place by the given spec using type-aware comparison.
// Items missing the sort field value sort to the end.
func ApplySort(items []QueryItem, spec SortSpec, typeIndex map[string]types.FieldType) {
	if spec.Field == "" {
		return
	}
	ft := typeIndex[spec.Field]
	slices.SortStableFunc(items, func(a, b QueryItem) int {
		av := a.Fields[spec.Field]
		bv := b.Fields[spec.Field]

		// Missing values sort to the end.
		aEmpty := av == ""
		bEmpty := bv == ""
		if aEmpty && bEmpty {
			return 0
		}
		if aEmpty {
			return 1
		}
		if bEmpty {
			return -1
		}

		cmp := compareForSort(av, bv, ft)
		if spec.Desc {
			cmp = -cmp
		}
		return cmp
	})
}

// compareForSort returns -1, 0, or 1 for type-aware ordering.
func compareForSort(a, b string, ft types.FieldType) int {
	switch ft {
	case types.FieldTypeNumber:
		fa, errA := strconv.ParseFloat(a, 64)
		fb, errB := strconv.ParseFloat(b, 64)
		if errA != nil || errB != nil {
			return strings.Compare(a, b)
		}
		if fa < fb {
			return -1
		}
		if fa > fb {
			return 1
		}
		return 0
	case types.FieldTypeDate, types.FieldTypeDatetime:
		ta := parseTime(a)
		tb := parseTime(b)
		if ta.IsZero() || tb.IsZero() {
			return strings.Compare(a, b)
		}
		if ta.Before(tb) {
			return -1
		}
		if ta.After(tb) {
			return 1
		}
		return 0
	default:
		return strings.Compare(a, b)
	}
}
