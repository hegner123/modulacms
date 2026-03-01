package query

import (
	"strconv"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/db/types"
)

// FilterOp represents a comparison operator for filtering.
type FilterOp string

const (
	OpEq   FilterOp = "eq"
	OpNeq  FilterOp = "neq"
	OpGt   FilterOp = "gt"
	OpGte  FilterOp = "gte"
	OpLt   FilterOp = "lt"
	OpLte  FilterOp = "lte"
	OpLike FilterOp = "like"
)

// Filter represents a single field filter.
type Filter struct {
	Field    string
	Operator FilterOp
	Value    string
}

// validOps contains the set of recognized filter operators.
var validOps = map[FilterOp]bool{
	OpEq: true, OpNeq: true, OpGt: true, OpGte: true,
	OpLt: true, OpLte: true, OpLike: true,
}

// reservedKeys are query parameter names that are not field filters.
var reservedKeys = map[string]bool{
	"sort": true, "limit": true, "offset": true,
	"locale": true, "status": true, "format": true,
}

// ParseFilters extracts field filters from raw query parameters.
// Keys in the reservedKeys set are skipped. Operator syntax: field[op]=value.
// Bare keys default to "eq". Unrecognized operators fall back to "eq".
func ParseFilters(params map[string][]string) []Filter {
	var filters []Filter
	for rawKey, vals := range params {
		if len(vals) == 0 {
			continue
		}
		field, op := parseFilterKey(rawKey)
		if reservedKeys[field] {
			continue
		}
		filters = append(filters, Filter{
			Field:    field,
			Operator: op,
			Value:    vals[0],
		})
	}
	return filters
}

// parseFilterKey splits "field[op]" into field name and operator.
// "title[eq]" -> ("title", "eq"). Bare "title" -> ("title", "eq").
func parseFilterKey(key string) (string, FilterOp) {
	bracketIdx := strings.IndexByte(key, '[')
	if bracketIdx < 0 {
		return key, OpEq
	}
	field := key[:bracketIdx]
	closeBracket := strings.IndexByte(key[bracketIdx:], ']')
	if closeBracket < 0 {
		return key, OpEq
	}
	opStr := key[bracketIdx+1 : bracketIdx+closeBracket]
	op := FilterOp(opStr)
	if !validOps[op] {
		return field, OpEq
	}
	return field, op
}

// applyFilters returns only items whose fields match all filters.
func applyFilters(items []QueryItem, filters []Filter, typeIndex map[string]types.FieldType) []QueryItem {
	if len(filters) == 0 {
		return items
	}
	result := make([]QueryItem, 0, len(items))
	for _, item := range items {
		if matchesAllFilters(item, filters, typeIndex) {
			result = append(result, item)
		}
	}
	return result
}

func matchesAllFilters(item QueryItem, filters []Filter, typeIndex map[string]types.FieldType) bool {
	for _, f := range filters {
		fieldValue, ok := item.Fields[f.Field]
		if !ok {
			return false
		}
		ft := typeIndex[f.Field]
		if !compareFieldValue(fieldValue, f.Value, f.Operator, ft) {
			return false
		}
	}
	return true
}

// compareFieldValue performs a type-aware comparison.
func compareFieldValue(fieldValue, filterValue string, op FilterOp, ft types.FieldType) bool {
	switch ft {
	case types.FieldTypeNumber:
		return compareNumber(fieldValue, filterValue, op)
	case types.FieldTypeBoolean:
		return compareBoolean(fieldValue, filterValue, op)
	case types.FieldTypeDate, types.FieldTypeDatetime:
		return compareDate(fieldValue, filterValue, op)
	default:
		return compareString(fieldValue, filterValue, op)
	}
}

func compareNumber(a, b string, op FilterOp) bool {
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA != nil || errB != nil {
		return compareString(a, b, op)
	}
	switch op {
	case OpEq:
		return fa == fb
	case OpNeq:
		return fa != fb
	case OpGt:
		return fa > fb
	case OpGte:
		return fa >= fb
	case OpLt:
		return fa < fb
	case OpLte:
		return fa <= fb
	case OpLike:
		return strings.Contains(a, b)
	default:
		return fa == fb
	}
}

func compareBoolean(a, b string, op FilterOp) bool {
	ba := normalizeBool(a)
	bb := normalizeBool(b)
	switch op {
	case OpEq:
		return ba == bb
	case OpNeq:
		return ba != bb
	default:
		return ba == bb
	}
}

func normalizeBool(s string) bool {
	lower := strings.ToLower(s)
	return lower == "true" || lower == "1"
}

// dateLayouts are the time layouts tried when parsing date filter values.
var dateLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02",
}

func compareDate(a, b string, op FilterOp) bool {
	ta := parseTime(a)
	tb := parseTime(b)
	if ta.IsZero() || tb.IsZero() {
		return compareString(a, b, op)
	}
	switch op {
	case OpEq:
		return ta.Equal(tb)
	case OpNeq:
		return !ta.Equal(tb)
	case OpGt:
		return ta.After(tb)
	case OpGte:
		return ta.After(tb) || ta.Equal(tb)
	case OpLt:
		return ta.Before(tb)
	case OpLte:
		return ta.Before(tb) || ta.Equal(tb)
	case OpLike:
		return strings.Contains(a, b)
	default:
		return ta.Equal(tb)
	}
}

func parseTime(s string) time.Time {
	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

func compareString(a, b string, op FilterOp) bool {
	switch op {
	case OpEq:
		return a == b
	case OpNeq:
		return a != b
	case OpGt:
		return a > b
	case OpGte:
		return a >= b
	case OpLt:
		return a < b
	case OpLte:
		return a <= b
	case OpLike:
		return strings.Contains(strings.ToLower(a), strings.ToLower(b))
	default:
		return a == b
	}
}
