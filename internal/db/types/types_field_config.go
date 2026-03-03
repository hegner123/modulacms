package types

import (
	"encoding/json"
	"fmt"
	"math"
)

// EmptyJSON is the canonical empty JSON object used as a default for JSON columns.
const EmptyJSON = "{}"

// Cardinality describes whether a relation targets one or many items.
type Cardinality string

// Cardinality constants for one-to-one and one-to-many relations.
const (
	CardinalityOne  Cardinality = "one"
	CardinalityMany Cardinality = "many"
)

// Validate checks whether c is a valid Cardinality value.
func (c Cardinality) Validate() error {
	switch c {
	case CardinalityOne, CardinalityMany:
		return nil
	default:
		return fmt.Errorf("Cardinality: invalid value %q (must be %q or %q)", c, CardinalityOne, CardinalityMany)
	}
}

// RelationConfig is the parsed form of a relation field's data_config JSON column.
type RelationConfig struct {
	TargetDatatypeID DatatypeID  `json:"target_datatype_id"`
	Cardinality      Cardinality `json:"cardinality"`
	MaxDepth         *int        `json:"max_depth,omitempty"`
}

// RuleOp identifies what a validation rule checks.
type RuleOp string

const (
	RuleRequired   RuleOp = "required"    // value must be non-empty
	RuleContains   RuleOp = "contains"    // value contains a substring or character class
	RuleStartsWith RuleOp = "starts_with" // value starts with a substring or character class
	RuleEndsWith   RuleOp = "ends_with"   // value ends with a substring or character class
	RuleEquals     RuleOp = "equals"      // value equals exactly
	RuleLength     RuleOp = "length"      // rune count of value
	RuleCount      RuleOp = "count"       // count occurrences of substring or class members
	RuleRange      RuleOp = "range"       // numeric value comparison (parse as float)
	RuleItemCount  RuleOp = "item_count"  // count items in comma-separated list or JSON array
	RuleOneOf      RuleOp = "one_of"      // value is one of a fixed set
)

// validRuleOps is the set of valid RuleOp values for validation.
var validRuleOps = map[RuleOp]bool{
	RuleRequired:   true,
	RuleContains:   true,
	RuleStartsWith: true,
	RuleEndsWith:   true,
	RuleEquals:     true,
	RuleLength:     true,
	RuleCount:      true,
	RuleRange:      true,
	RuleItemCount:  true,
	RuleOneOf:      true,
}

// Cmp is used for numeric comparisons in length/count/range rules.
type Cmp string

const (
	CmpEq  Cmp = "eq"  // ==
	CmpNeq Cmp = "neq" // !=
	CmpGt  Cmp = "gt"  // >
	CmpGte Cmp = "gte" // >=
	CmpLt  Cmp = "lt"  // <
	CmpLte Cmp = "lte" // <=
)

// validCmps is the set of valid Cmp values for validation.
var validCmps = map[Cmp]bool{
	CmpEq:  true,
	CmpNeq: true,
	CmpGt:  true,
	CmpGte: true,
	CmpLt:  true,
	CmpLte: true,
}

// CharClass identifies a predefined set of characters.
type CharClass string

const (
	ClassUppercase CharClass = "uppercase" // A-Z
	ClassLowercase CharClass = "lowercase" // a-z
	ClassDigits    CharClass = "digits"    // 0-9
	ClassSymbols   CharClass = "symbols"   // any char NOT in a-z, A-Z, 0-9, or whitespace
	ClassSpaces    CharClass = "spaces"    // whitespace (space, tab, newline)
)

// validCharClasses is the set of valid CharClass values for validation.
var validCharClasses = map[CharClass]bool{
	ClassUppercase: true,
	ClassLowercase: true,
	ClassDigits:    true,
	ClassSymbols:   true,
	ClassSpaces:    true,
}

// ClassifyChar returns whether the given rune belongs to the specified character class.
func ClassifyChar(c rune, class CharClass) bool {
	switch class {
	case ClassUppercase:
		return c >= 'A' && c <= 'Z'
	case ClassLowercase:
		return c >= 'a' && c <= 'z'
	case ClassDigits:
		return c >= '0' && c <= '9'
	case ClassSymbols:
		return !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == ' ' || c == '\t' || c == '\n' || c == '\r')
	case ClassSpaces:
		return c == ' ' || c == '\t' || c == '\n' || c == '\r'
	default:
		return false
	}
}

// ValidationRule is a single validation predicate.
type ValidationRule struct {
	Op      RuleOp    `json:"op"`                // operation to perform
	Value   string    `json:"value,omitempty"`   // literal string (for contains, starts_with, ends_with, equals)
	Values  []string  `json:"values,omitempty"`  // set of values (for one_of)
	Class   CharClass `json:"class,omitempty"`   // character class (alternative to Value for contains/count)
	Cmp     Cmp       `json:"cmp,omitempty"`     // comparison operator (required for length, count, range, item_count)
	N       *float64  `json:"n,omitempty"`       // numeric operand (for length, count, range, item_count)
	Negate  bool      `json:"negate,omitempty"`  // invert the result (contains -> not_contains)
	Message string    `json:"message,omitempty"` // custom error message (overrides default)
}

// RuleEntry is a single entry in a validation rule list, containing either a
// flat rule or a group of rules. Exactly one of Rule or Group must be set.
type RuleEntry struct {
	Rule  *ValidationRule `json:"rule,omitempty"`
	Group *RuleGroup      `json:"group,omitempty"`
}

// RuleGroup applies a logical operator across its children. Exactly one of
// AllOf or AnyOf must be populated, with at least one entry.
type RuleGroup struct {
	AllOf []RuleEntry `json:"all_of,omitempty"` // AND -- all must pass
	AnyOf []RuleEntry `json:"any_of,omitempty"` // OR  -- at least one must pass
}

// ValidationConfig is the parsed form of a field's validation JSON column.
// All constraints are expressed as composable rules.
type ValidationConfig struct {
	Rules []RuleEntry `json:"rules,omitempty"`
}

// N bounds for validation.
const (
	maxIntegerN = 1_000_000
	maxRangeN   = 1e15
	minRangeN   = -1e15
)

// opsAllowingNegate lists the ops where Negate is valid.
var opsAllowingNegate = map[RuleOp]bool{
	RuleContains:   true,
	RuleStartsWith: true,
	RuleEndsWith:   true,
	RuleEquals:     true,
	RuleOneOf:      true,
}

// maxRuleNestingDepth is the maximum allowed nesting depth for rule groups.
const maxRuleNestingDepth = 10

// ValidateRuleDefinition checks that a single ValidationRule is well-formed.
func ValidateRuleDefinition(r ValidationRule) error {
	if !validRuleOps[r.Op] {
		return fmt.Errorf("ValidateRuleDefinition: invalid op %q", r.Op)
	}

	// Negate is only valid on certain ops.
	if r.Negate && !opsAllowingNegate[r.Op] {
		return fmt.Errorf("ValidateRuleDefinition: negate is not valid on op %q", r.Op)
	}

	switch r.Op {
	case RuleRequired:
		// All other fields must be empty.
		if r.Value != "" || r.Class != "" || r.Cmp != "" || r.N != nil || len(r.Values) > 0 || r.Negate {
			return fmt.Errorf("ValidateRuleDefinition: required op must have no other fields set")
		}

	case RuleContains, RuleStartsWith, RuleEndsWith:
		// Exactly one of Value or Class must be set.
		hasValue := r.Value != ""
		hasClass := r.Class != ""
		if hasValue == hasClass {
			return fmt.Errorf("ValidateRuleDefinition: %s requires exactly one of value or class", r.Op)
		}
		if hasClass {
			if !validCharClasses[r.Class] {
				return fmt.Errorf("ValidateRuleDefinition: invalid class %q", r.Class)
			}
		}
		if r.Cmp != "" || r.N != nil || len(r.Values) > 0 {
			return fmt.Errorf("ValidateRuleDefinition: %s does not accept cmp, n, or values", r.Op)
		}

	case RuleEquals:
		// Value must be set, Class must be empty.
		if r.Value == "" {
			return fmt.Errorf("ValidateRuleDefinition: equals requires value")
		}
		if r.Class != "" || r.Cmp != "" || r.N != nil || len(r.Values) > 0 {
			return fmt.Errorf("ValidateRuleDefinition: equals does not accept class, cmp, n, or values")
		}

	case RuleLength:
		// Cmp and N must be set, Value/Class must be empty.
		if err := validateCmpAndN(r, "length"); err != nil {
			return err
		}
		if r.Value != "" || r.Class != "" || len(r.Values) > 0 {
			return fmt.Errorf("ValidateRuleDefinition: length does not accept value, class, or values")
		}
		if err := validateIntegerN(*r.N, "length"); err != nil {
			return err
		}

	case RuleCount:
		// Cmp and N must be set, exactly one of Value or Class must be set.
		if err := validateCmpAndN(r, "count"); err != nil {
			return err
		}
		hasValue := r.Value != ""
		hasClass := r.Class != ""
		if hasValue == hasClass {
			return fmt.Errorf("ValidateRuleDefinition: count requires exactly one of value or class")
		}
		if hasClass {
			if !validCharClasses[r.Class] {
				return fmt.Errorf("ValidateRuleDefinition: invalid class %q", r.Class)
			}
		}
		if len(r.Values) > 0 {
			return fmt.Errorf("ValidateRuleDefinition: count does not accept values")
		}
		if err := validateIntegerN(*r.N, "count"); err != nil {
			return err
		}

	case RuleRange:
		// Cmp and N must be set, Value/Class must be empty.
		// Range allows negative N, so use validateCmpAndNAllowNegative.
		if err := validateCmpAndNAllowNegative(r, "range"); err != nil {
			return err
		}
		if r.Value != "" || r.Class != "" || len(r.Values) > 0 {
			return fmt.Errorf("ValidateRuleDefinition: range does not accept value, class, or values")
		}
		if *r.N < minRangeN || *r.N > maxRangeN {
			return fmt.Errorf("ValidateRuleDefinition: range n must be between %g and %g", minRangeN, maxRangeN)
		}

	case RuleItemCount:
		// Cmp and N must be set, Value/Class must be empty.
		if err := validateCmpAndN(r, "item_count"); err != nil {
			return err
		}
		if r.Value != "" || r.Class != "" || len(r.Values) > 0 {
			return fmt.Errorf("ValidateRuleDefinition: item_count does not accept value, class, or values")
		}
		if err := validateIntegerN(*r.N, "item_count"); err != nil {
			return err
		}

	case RuleOneOf:
		// Values must be non-empty, Value/Class must be empty.
		if len(r.Values) == 0 {
			return fmt.Errorf("ValidateRuleDefinition: one_of requires non-empty values")
		}
		if r.Value != "" || r.Class != "" || r.Cmp != "" || r.N != nil {
			return fmt.Errorf("ValidateRuleDefinition: one_of does not accept value, class, cmp, or n")
		}
	}

	return nil
}

// validateCmpAndN checks that both Cmp and N are set and valid.
func validateCmpAndN(r ValidationRule, opName string) error {
	if r.Cmp == "" {
		return fmt.Errorf("ValidateRuleDefinition: %s requires cmp", opName)
	}
	if !validCmps[r.Cmp] {
		return fmt.Errorf("ValidateRuleDefinition: invalid cmp %q", r.Cmp)
	}
	if r.N == nil {
		return fmt.Errorf("ValidateRuleDefinition: %s requires n", opName)
	}
	if *r.N < 0 {
		return fmt.Errorf("ValidateRuleDefinition: %s n must be non-negative", opName)
	}
	return nil
}

// validateCmpAndNAllowNegative checks that both Cmp and N are set and valid,
// without enforcing a non-negative constraint on N. Used by the range op.
func validateCmpAndNAllowNegative(r ValidationRule, opName string) error {
	if r.Cmp == "" {
		return fmt.Errorf("ValidateRuleDefinition: %s requires cmp", opName)
	}
	if !validCmps[r.Cmp] {
		return fmt.Errorf("ValidateRuleDefinition: invalid cmp %q", r.Cmp)
	}
	if r.N == nil {
		return fmt.Errorf("ValidateRuleDefinition: %s requires n", opName)
	}
	return nil
}

// validateIntegerN checks that N is a whole number within the integer-semantic
// bounds [0, 1_000_000].
func validateIntegerN(n float64, opName string) error {
	if n != math.Trunc(n) {
		return fmt.Errorf("ValidateRuleDefinition: %s n must be a whole number, got %g", opName, n)
	}
	if n > maxIntegerN {
		return fmt.Errorf("ValidateRuleDefinition: %s n must be at most %d", opName, maxIntegerN)
	}
	return nil
}

// ValidateRuleEntries validates a slice of RuleEntry values recursively.
// The depth parameter tracks nesting level; pass 0 for top-level entries.
func ValidateRuleEntries(entries []RuleEntry, depth int) error {
	if depth > maxRuleNestingDepth {
		return fmt.Errorf("ValidateRuleEntries: maximum nesting depth of %d exceeded", maxRuleNestingDepth)
	}
	for i, entry := range entries {
		hasRule := entry.Rule != nil
		hasGroup := entry.Group != nil
		if hasRule == hasGroup {
			return fmt.Errorf("ValidateRuleEntries: entry[%d] must have exactly one of rule or group", i)
		}
		if hasRule {
			if depth > 0 && entry.Rule.Op == RuleRequired {
				return fmt.Errorf("ValidateRuleEntries: required op is not valid inside a group (entry[%d])", i)
			}
			if err := ValidateRuleDefinition(*entry.Rule); err != nil {
				return fmt.Errorf("ValidateRuleEntries: entry[%d]: %w", i, err)
			}
		}
		if hasGroup {
			g := entry.Group
			hasAllOf := len(g.AllOf) > 0
			hasAnyOf := len(g.AnyOf) > 0
			if hasAllOf == hasAnyOf {
				return fmt.Errorf("ValidateRuleEntries: entry[%d] group must have exactly one of all_of or any_of (non-empty)", i)
			}
			if hasAllOf {
				if err := ValidateRuleEntries(g.AllOf, depth+1); err != nil {
					return fmt.Errorf("ValidateRuleEntries: entry[%d] all_of: %w", i, err)
				}
			}
			if hasAnyOf {
				if err := ValidateRuleEntries(g.AnyOf, depth+1); err != nil {
					return fmt.Errorf("ValidateRuleEntries: entry[%d] any_of: %w", i, err)
				}
			}
		}
	}
	return nil
}

// ValidateValidationConfig validates an entire ValidationConfig.
func ValidateValidationConfig(vc ValidationConfig) error {
	return ValidateRuleEntries(vc.Rules, 0)
}

// UIConfig is the parsed form of a field's ui_config JSON column.
type UIConfig struct {
	Widget      string `json:"widget,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	HelpText    string `json:"help_text,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
}

// ParseValidationConfig parses a JSON string into a ValidationConfig.
// Returns a zero-value config for empty or "{}" input.
func ParseValidationConfig(s string) (ValidationConfig, error) {
	if s == "" || s == EmptyJSON {
		return ValidationConfig{}, nil
	}
	var vc ValidationConfig
	if err := json.Unmarshal([]byte(s), &vc); err != nil {
		return ValidationConfig{}, fmt.Errorf("ParseValidationConfig: %w", err)
	}
	return vc, nil
}

// ParseUIConfig parses a JSON string into a UIConfig.
// Returns a zero-value config for empty or "{}" input.
func ParseUIConfig(s string) (UIConfig, error) {
	if s == "" || s == EmptyJSON {
		return UIConfig{}, nil
	}
	var uc UIConfig
	if err := json.Unmarshal([]byte(s), &uc); err != nil {
		return UIConfig{}, fmt.Errorf("ParseUIConfig: %w", err)
	}
	return uc, nil
}

// RichTextConfig is the parsed form of a richtext field's data JSON column.
// An empty or zero-value config means "use global default toolbar".
type RichTextConfig struct {
	Toolbar []string `json:"toolbar,omitempty"`
}

// ParseRichTextConfig parses a JSON string into a RichTextConfig.
// Returns a zero-value config for empty or "{}" input (meaning "use global default").
func ParseRichTextConfig(s string) (RichTextConfig, error) {
	if s == "" || s == EmptyJSON {
		return RichTextConfig{}, nil
	}
	var rc RichTextConfig
	if err := json.Unmarshal([]byte(s), &rc); err != nil {
		return RichTextConfig{}, fmt.Errorf("ParseRichTextConfig: %w", err)
	}
	return rc, nil
}

// ParseRelationConfig parses a JSON string into a RelationConfig.
// Unlike the other parse functions, this returns an error for empty or "{}" input
// because relation fields require a non-empty data config.
func ParseRelationConfig(s string) (RelationConfig, error) {
	if s == "" || s == EmptyJSON {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: relation fields require a non-empty data config")
	}
	var rc RelationConfig
	if err := json.Unmarshal([]byte(s), &rc); err != nil {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: %w", err)
	}
	if rc.TargetDatatypeID.IsZero() {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: target_datatype_id is required")
	}
	if err := rc.Cardinality.Validate(); err != nil {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: %w", err)
	}
	return rc, nil
}
