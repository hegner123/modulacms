package validation

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hegner123/modulacms/internal/db/types"
)

// EvaluateRules evaluates the composable rule tree against the given value.
// Returns a list of error messages (empty if all rules pass).
// The required op is skipped since it is handled before rule evaluation.
func EvaluateRules(value string, entries []types.RuleEntry) []string {
	var messages []string
	for _, entry := range entries {
		msgs := evaluateEntry(value, entry)
		messages = append(messages, msgs...)
	}
	return messages
}

// evaluateEntry evaluates a single RuleEntry (either a rule or a group).
func evaluateEntry(value string, entry types.RuleEntry) []string {
	if entry.Rule != nil {
		return evaluateRule(value, *entry.Rule)
	}
	if entry.Group != nil {
		return evaluateGroup(value, *entry.Group)
	}
	return nil
}

// evaluateRule evaluates a single ValidationRule against the value.
// Returns error messages if the rule fails, or nil if it passes.
func evaluateRule(value string, rule types.ValidationRule) []string {
	// Skip required — handled before rule evaluation.
	if rule.Op == types.RuleRequired {
		return nil
	}

	passed := false

	switch rule.Op {
	case types.RuleContains:
		passed = evalContains(value, rule)
	case types.RuleStartsWith:
		passed = evalStartsWith(value, rule)
	case types.RuleEndsWith:
		passed = evalEndsWith(value, rule)
	case types.RuleEquals:
		passed = value == rule.Value
	case types.RuleLength:
		passed = evalLength(value, rule)
	case types.RuleCount:
		passed = evalCount(value, rule)
	case types.RuleRange:
		msg := evalRange(value, rule)
		if msg != "" {
			return []string{msg}
		}
		return nil
	case types.RuleItemCount:
		passed = evalItemCount(value, rule)
	case types.RuleOneOf:
		passed = evalOneOf(value, rule)
	default:
		// Unknown op: skip silently (forward compatibility).
		return nil
	}

	// Apply negate.
	if rule.Negate {
		passed = !passed
	}

	if passed {
		return nil
	}

	msg := rule.Message
	if msg == "" {
		msg = defaultMessage(rule)
	}
	return []string{msg}
}

// evaluateGroup evaluates a RuleGroup (AllOf or AnyOf) against the value.
func evaluateGroup(value string, group types.RuleGroup) []string {
	if len(group.AllOf) > 0 {
		// AllOf: collect ALL error messages from any failing entry.
		var messages []string
		for _, entry := range group.AllOf {
			msgs := evaluateEntry(value, entry)
			messages = append(messages, msgs...)
		}
		return messages
	}

	if len(group.AnyOf) > 0 {
		// AnyOf: if ALL fail, return the first failing entry's messages.
		// If any passes, return no errors.
		var firstFailure []string
		for _, entry := range group.AnyOf {
			msgs := evaluateEntry(value, entry)
			if len(msgs) == 0 {
				return nil // at least one passed
			}
			if firstFailure == nil {
				firstFailure = msgs
			}
		}
		return firstFailure
	}

	return nil
}

// evalContains checks if value contains the literal or character class.
func evalContains(value string, rule types.ValidationRule) bool {
	if rule.Value != "" {
		return strings.Contains(value, rule.Value)
	}
	// Character class: iterate runes, return true if any matches.
	for _, r := range value {
		if types.ClassifyChar(r, rule.Class) {
			return true
		}
	}
	return false
}

// evalStartsWith checks if value starts with the literal or character class.
func evalStartsWith(value string, rule types.ValidationRule) bool {
	if rule.Value != "" {
		return strings.HasPrefix(value, rule.Value)
	}
	// Character class: check first rune.
	r, size := utf8.DecodeRuneInString(value)
	if size == 0 {
		return false
	}
	return types.ClassifyChar(r, rule.Class)
}

// evalEndsWith checks if value ends with the literal or character class.
func evalEndsWith(value string, rule types.ValidationRule) bool {
	if rule.Value != "" {
		return strings.HasSuffix(value, rule.Value)
	}
	// Character class: check last rune.
	r, size := utf8.DecodeLastRuneInString(value)
	if size == 0 {
		return false
	}
	return types.ClassifyChar(r, rule.Class)
}

// evalLength compares the rune count of value against N using Cmp.
func evalLength(value string, rule types.ValidationRule) bool {
	length := utf8.RuneCountInString(value)
	n := int(*rule.N)
	return compareCmp(length, n, rule.Cmp)
}

// evalCount counts occurrences of the literal or character class and compares against N.
func evalCount(value string, rule types.ValidationRule) bool {
	var count int
	if rule.Value != "" {
		count = strings.Count(value, rule.Value)
	} else {
		for _, r := range value {
			if types.ClassifyChar(r, rule.Class) {
				count++
			}
		}
	}
	n := int(*rule.N)
	return compareCmp(count, n, rule.Cmp)
}

// evalRange parses the value as float64 and compares against N.
// Returns an error message if parse fails or comparison fails, empty string on success.
// Range is special because a parse failure produces its own error message.
func evalRange(value string, rule types.ValidationRule) string {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		msg := rule.Message
		if msg == "" {
			msg = "must be a number"
		}
		return msg
	}
	if compareFloat(f, *rule.N, rule.Cmp) {
		return ""
	}
	msg := rule.Message
	if msg == "" {
		msg = defaultMessage(rule)
	}
	return msg
}

// evalItemCount counts items in the value and compares against N.
func evalItemCount(value string, rule types.ValidationRule) bool {
	count := countItems(value)
	n := int(*rule.N)
	return compareCmp(count, n, rule.Cmp)
}

// evalOneOf checks if value is in the Values set.
func evalOneOf(value string, rule types.ValidationRule) bool {
	for _, v := range rule.Values {
		if value == v {
			return true
		}
	}
	return false
}

// countItems counts items in a value using item_count semantics:
// - Empty string: 0
// - Starts with "[": try JSON array, fall back to comma-separated
// - Otherwise: comma-separated, trimming whitespace, skipping empty segments
func countItems(value string) int {
	if value == "" {
		return 0
	}

	if strings.HasPrefix(value, "[") {
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(value), &arr); err == nil {
			return len(arr)
		}
		// Fall through to comma-separated on unmarshal failure.
	}

	parts := strings.Split(value, ",")
	count := 0
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			count++
		}
	}
	return count
}

// compareCmp compares two integers using the given Cmp operator.
func compareCmp(actual, expected int, cmp types.Cmp) bool {
	switch cmp {
	case types.CmpEq:
		return actual == expected
	case types.CmpNeq:
		return actual != expected
	case types.CmpGt:
		return actual > expected
	case types.CmpGte:
		return actual >= expected
	case types.CmpLt:
		return actual < expected
	case types.CmpLte:
		return actual <= expected
	default:
		return false
	}
}

// compareFloat compares two float64 values using the given Cmp operator.
func compareFloat(actual, expected float64, cmp types.Cmp) bool {
	switch cmp {
	case types.CmpEq:
		return actual == expected
	case types.CmpNeq:
		return actual != expected
	case types.CmpGt:
		return actual > expected
	case types.CmpGte:
		return actual >= expected
	case types.CmpLt:
		return actual < expected
	case types.CmpLte:
		return actual <= expected
	default:
		return false
	}
}

// cmpLabel returns a human-readable label for a comparison operator.
func cmpLabel(cmp types.Cmp) string {
	switch cmp {
	case types.CmpEq:
		return "exactly"
	case types.CmpNeq:
		return "not"
	case types.CmpGt:
		return "more than"
	case types.CmpGte:
		return "at least"
	case types.CmpLt:
		return "less than"
	case types.CmpLte:
		return "at most"
	default:
		return string(cmp)
	}
}

// formatN formats a float64 for display in error messages.
// Whole numbers are displayed without a decimal point.
func formatN(n float64) string {
	if n == float64(int64(n)) {
		return strconv.FormatInt(int64(n), 10)
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// defaultMessage generates a default error message for a failing rule.
func defaultMessage(rule types.ValidationRule) string {
	switch rule.Op {
	case types.RuleRequired:
		return "is required"

	case types.RuleContains:
		target := targetLabel(rule)
		if rule.Negate {
			return fmt.Sprintf("must not contain %s", target)
		}
		return fmt.Sprintf("must contain %s", target)

	case types.RuleStartsWith:
		target := targetLabel(rule)
		if rule.Negate {
			return fmt.Sprintf("must not start with %s", target)
		}
		return fmt.Sprintf("must start with %s", target)

	case types.RuleEndsWith:
		target := targetLabel(rule)
		if rule.Negate {
			return fmt.Sprintf("must not end with %s", target)
		}
		return fmt.Sprintf("must end with %s", target)

	case types.RuleEquals:
		if rule.Negate {
			return fmt.Sprintf("must not equal %q", rule.Value)
		}
		return fmt.Sprintf("must equal %q", rule.Value)

	case types.RuleLength:
		return fmt.Sprintf("must be %s %s characters", cmpLabel(rule.Cmp), formatN(*rule.N))

	case types.RuleCount:
		target := targetLabel(rule)
		return fmt.Sprintf("must have %s %s occurrences of %s", cmpLabel(rule.Cmp), formatN(*rule.N), target)

	case types.RuleRange:
		return fmt.Sprintf("value must be %s %s", cmpLabel(rule.Cmp), formatN(*rule.N))

	case types.RuleItemCount:
		return fmt.Sprintf("must have %s %s items", cmpLabel(rule.Cmp), formatN(*rule.N))

	case types.RuleOneOf:
		joined := strings.Join(rule.Values, ", ")
		if rule.Negate {
			return fmt.Sprintf("must not be one of: %s", joined)
		}
		return fmt.Sprintf("must be one of: %s", joined)

	default:
		return "validation failed"
	}
}

// targetLabel returns the display label for a rule's value or class target.
func targetLabel(rule types.ValidationRule) string {
	if rule.Value != "" {
		return fmt.Sprintf("%q", rule.Value)
	}
	return string(rule.Class) + " characters"
}
