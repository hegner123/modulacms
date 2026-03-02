package types

import (
	"testing"
)

// ============================================================
// Cardinality
// ============================================================

func TestCardinality_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		c       Cardinality
		wantErr bool
	}{
		{name: "one", c: CardinalityOne, wantErr: false},
		{name: "many", c: CardinalityMany, wantErr: false},
		{name: "empty", c: "", wantErr: true},
		{name: "invalid", c: "both", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cardinality(%q).Validate() error = %v, wantErr %v", tt.c, err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// ParseValidationConfig
// ============================================================

func TestParseValidationConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(t *testing.T, vc ValidationConfig)
	}{
		{
			name:    "empty string returns zero",
			input:   "",
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				if len(vc.Rules) != 0 {
					t.Errorf("empty input: Rules = %v, want nil/empty", vc.Rules)
				}
			},
		},
		{
			name:    "empty JSON object returns zero",
			input:   "{}",
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				if len(vc.Rules) != 0 {
					t.Errorf("{} input: Rules = %v, want nil/empty", vc.Rules)
				}
			},
		},
		{
			name:    "valid rules JSON",
			input:   `{"rules":[{"rule":{"op":"required"}},{"rule":{"op":"length","cmp":"gte","n":5}}]}`,
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				if len(vc.Rules) != 2 {
					t.Fatalf("Rules length = %d, want 2", len(vc.Rules))
				}
				if vc.Rules[0].Rule == nil || vc.Rules[0].Rule.Op != RuleRequired {
					t.Error("Rules[0] is not a required rule")
				}
				r1 := vc.Rules[1].Rule
				if r1 == nil || r1.Op != RuleLength || r1.Cmp != CmpGte || r1.N == nil || *r1.N != 5 {
					t.Error("Rules[1] is not a length>=5 rule")
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   "{bad",
			wantErr: true,
		},
		{
			name:    "old flat format parses to empty rules",
			input:   `{"required":true,"min_length":5,"max_length":100}`,
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				// Old flat fields are not on the new struct, so they are ignored.
				// Rules should be nil/empty.
				if len(vc.Rules) != 0 {
					t.Errorf("old flat format: Rules = %v, want nil/empty", vc.Rules)
				}
			},
		},
		{
			name:    "group with all_of",
			input:   `{"rules":[{"group":{"all_of":[{"rule":{"op":"contains","class":"uppercase"}},{"rule":{"op":"contains","class":"digits"}}]}}]}`,
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				if len(vc.Rules) != 1 {
					t.Fatalf("Rules length = %d, want 1", len(vc.Rules))
				}
				g := vc.Rules[0].Group
				if g == nil {
					t.Fatal("Rules[0].Group is nil")
				}
				if len(g.AllOf) != 2 {
					t.Fatalf("AllOf length = %d, want 2", len(g.AllOf))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vc, err := ParseValidationConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseValidationConfig(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, vc)
			}
		})
	}
}

// ============================================================
// ValidateRuleDefinition
// ============================================================

func floatPtr(f float64) *float64 { return &f }

func TestValidateRuleDefinition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		rule    ValidationRule
		wantErr bool
	}{
		// --- valid rules ---
		{name: "valid required", rule: ValidationRule{Op: RuleRequired}, wantErr: false},
		{name: "valid contains value", rule: ValidationRule{Op: RuleContains, Value: "abc"}, wantErr: false},
		{name: "valid contains class", rule: ValidationRule{Op: RuleContains, Class: ClassUppercase}, wantErr: false},
		{name: "valid contains negate", rule: ValidationRule{Op: RuleContains, Value: "x", Negate: true}, wantErr: false},
		{name: "valid starts_with value", rule: ValidationRule{Op: RuleStartsWith, Value: "http"}, wantErr: false},
		{name: "valid starts_with class", rule: ValidationRule{Op: RuleStartsWith, Class: ClassDigits}, wantErr: false},
		{name: "valid ends_with value", rule: ValidationRule{Op: RuleEndsWith, Value: ".com"}, wantErr: false},
		{name: "valid ends_with class", rule: ValidationRule{Op: RuleEndsWith, Class: ClassLowercase}, wantErr: false},
		{name: "valid equals", rule: ValidationRule{Op: RuleEquals, Value: "exact"}, wantErr: false},
		{name: "valid equals negate", rule: ValidationRule{Op: RuleEquals, Value: "exact", Negate: true}, wantErr: false},
		{name: "valid length", rule: ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(5)}, wantErr: false},
		{name: "valid count value", rule: ValidationRule{Op: RuleCount, Cmp: CmpEq, N: floatPtr(3), Value: "a"}, wantErr: false},
		{name: "valid count class", rule: ValidationRule{Op: RuleCount, Cmp: CmpGt, N: floatPtr(2), Class: ClassDigits}, wantErr: false},
		{name: "valid range", rule: ValidationRule{Op: RuleRange, Cmp: CmpGte, N: floatPtr(0)}, wantErr: false},
		{name: "valid range negative", rule: ValidationRule{Op: RuleRange, Cmp: CmpGte, N: floatPtr(-100)}, wantErr: false},
		{name: "valid range decimal", rule: ValidationRule{Op: RuleRange, Cmp: CmpLte, N: floatPtr(99.99)}, wantErr: false},
		{name: "valid item_count", rule: ValidationRule{Op: RuleItemCount, Cmp: CmpLte, N: floatPtr(5)}, wantErr: false},
		{name: "valid one_of", rule: ValidationRule{Op: RuleOneOf, Values: []string{"a", "b", "c"}}, wantErr: false},
		{name: "valid one_of negate", rule: ValidationRule{Op: RuleOneOf, Values: []string{"x"}, Negate: true}, wantErr: false},
		{name: "valid length zero", rule: ValidationRule{Op: RuleLength, Cmp: CmpEq, N: floatPtr(0)}, wantErr: false},

		// --- invalid op ---
		{name: "invalid op", rule: ValidationRule{Op: "bogus"}, wantErr: true},
		{name: "empty op", rule: ValidationRule{Op: ""}, wantErr: true},

		// --- required with extra fields ---
		{name: "required with value", rule: ValidationRule{Op: RuleRequired, Value: "x"}, wantErr: true},
		{name: "required with class", rule: ValidationRule{Op: RuleRequired, Class: ClassDigits}, wantErr: true},
		{name: "required with cmp", rule: ValidationRule{Op: RuleRequired, Cmp: CmpEq}, wantErr: true},
		{name: "required with n", rule: ValidationRule{Op: RuleRequired, N: floatPtr(1)}, wantErr: true},
		{name: "required with values", rule: ValidationRule{Op: RuleRequired, Values: []string{"a"}}, wantErr: true},
		{name: "required with negate", rule: ValidationRule{Op: RuleRequired, Negate: true}, wantErr: true},

		// --- contains/starts_with/ends_with errors ---
		{name: "contains no value or class", rule: ValidationRule{Op: RuleContains}, wantErr: true},
		{name: "contains both value and class", rule: ValidationRule{Op: RuleContains, Value: "x", Class: ClassDigits}, wantErr: true},
		{name: "contains invalid class", rule: ValidationRule{Op: RuleContains, Class: "bogus"}, wantErr: true},
		{name: "contains with cmp", rule: ValidationRule{Op: RuleContains, Value: "x", Cmp: CmpEq}, wantErr: true},
		{name: "starts_with no value or class", rule: ValidationRule{Op: RuleStartsWith}, wantErr: true},
		{name: "ends_with no value or class", rule: ValidationRule{Op: RuleEndsWith}, wantErr: true},

		// --- equals errors ---
		{name: "equals no value", rule: ValidationRule{Op: RuleEquals}, wantErr: true},
		{name: "equals with class", rule: ValidationRule{Op: RuleEquals, Value: "x", Class: ClassDigits}, wantErr: true},
		{name: "equals with cmp", rule: ValidationRule{Op: RuleEquals, Value: "x", Cmp: CmpEq}, wantErr: true},

		// --- length errors ---
		{name: "length no cmp", rule: ValidationRule{Op: RuleLength, N: floatPtr(5)}, wantErr: true},
		{name: "length no n", rule: ValidationRule{Op: RuleLength, Cmp: CmpGte}, wantErr: true},
		{name: "length with value", rule: ValidationRule{Op: RuleLength, Cmp: CmpEq, N: floatPtr(5), Value: "x"}, wantErr: true},
		{name: "length invalid cmp", rule: ValidationRule{Op: RuleLength, Cmp: "bad", N: floatPtr(5)}, wantErr: true},
		{name: "length fractional n", rule: ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(5.5)}, wantErr: true},
		{name: "length negative n", rule: ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(-1)}, wantErr: true},
		{name: "length n too large", rule: ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(2_000_000)}, wantErr: true},

		// --- count errors ---
		{name: "count no value or class", rule: ValidationRule{Op: RuleCount, Cmp: CmpEq, N: floatPtr(1)}, wantErr: true},
		{name: "count both value and class", rule: ValidationRule{Op: RuleCount, Cmp: CmpEq, N: floatPtr(1), Value: "a", Class: ClassDigits}, wantErr: true},
		{name: "count fractional n", rule: ValidationRule{Op: RuleCount, Cmp: CmpEq, N: floatPtr(1.5), Value: "a"}, wantErr: true},

		// --- range errors ---
		{name: "range no cmp", rule: ValidationRule{Op: RuleRange, N: floatPtr(5)}, wantErr: true},
		{name: "range no n", rule: ValidationRule{Op: RuleRange, Cmp: CmpGte}, wantErr: true},
		{name: "range n too large", rule: ValidationRule{Op: RuleRange, Cmp: CmpGte, N: floatPtr(2e15)}, wantErr: true},
		{name: "range n too small", rule: ValidationRule{Op: RuleRange, Cmp: CmpGte, N: floatPtr(-2e15)}, wantErr: true},
		{name: "range with value", rule: ValidationRule{Op: RuleRange, Cmp: CmpGte, N: floatPtr(0), Value: "x"}, wantErr: true},

		// --- item_count errors ---
		{name: "item_count fractional n", rule: ValidationRule{Op: RuleItemCount, Cmp: CmpLte, N: floatPtr(5.5)}, wantErr: true},
		{name: "item_count n too large", rule: ValidationRule{Op: RuleItemCount, Cmp: CmpLte, N: floatPtr(2_000_000)}, wantErr: true},

		// --- one_of errors ---
		{name: "one_of empty values", rule: ValidationRule{Op: RuleOneOf, Values: []string{}}, wantErr: true},
		{name: "one_of with value", rule: ValidationRule{Op: RuleOneOf, Values: []string{"a"}, Value: "x"}, wantErr: true},
		{name: "one_of with cmp", rule: ValidationRule{Op: RuleOneOf, Values: []string{"a"}, Cmp: CmpEq}, wantErr: true},

		// --- negate on non-negatable ops ---
		{name: "negate on length", rule: ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(5), Negate: true}, wantErr: true},
		{name: "negate on range", rule: ValidationRule{Op: RuleRange, Cmp: CmpGte, N: floatPtr(0), Negate: true}, wantErr: true},
		{name: "negate on count", rule: ValidationRule{Op: RuleCount, Cmp: CmpEq, N: floatPtr(1), Value: "a", Negate: true}, wantErr: true},
		{name: "negate on item_count", rule: ValidationRule{Op: RuleItemCount, Cmp: CmpLte, N: floatPtr(5), Negate: true}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRuleDefinition(tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRuleDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// ValidateRuleEntries
// ============================================================

func TestValidateRuleEntries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		entries []RuleEntry
		depth   int
		wantErr bool
	}{
		{
			name:    "nil entries",
			entries: nil,
			depth:   0,
			wantErr: false,
		},
		{
			name:    "empty entries",
			entries: []RuleEntry{},
			depth:   0,
			wantErr: false,
		},
		{
			name: "single valid rule",
			entries: []RuleEntry{
				{Rule: &ValidationRule{Op: RuleRequired}},
			},
			depth:   0,
			wantErr: false,
		},
		{
			name: "entry with neither rule nor group",
			entries: []RuleEntry{
				{},
			},
			depth:   0,
			wantErr: true,
		},
		{
			name: "entry with both rule and group",
			entries: []RuleEntry{
				{
					Rule:  &ValidationRule{Op: RuleRequired},
					Group: &RuleGroup{AllOf: []RuleEntry{{Rule: &ValidationRule{Op: RuleContains, Value: "x"}}}},
				},
			},
			depth:   0,
			wantErr: true,
		},
		{
			name: "valid all_of group",
			entries: []RuleEntry{
				{Group: &RuleGroup{AllOf: []RuleEntry{
					{Rule: &ValidationRule{Op: RuleContains, Value: "x"}},
					{Rule: &ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(5)}},
				}}},
			},
			depth:   0,
			wantErr: false,
		},
		{
			name: "valid any_of group",
			entries: []RuleEntry{
				{Group: &RuleGroup{AnyOf: []RuleEntry{
					{Rule: &ValidationRule{Op: RuleContains, Class: ClassUppercase}},
					{Rule: &ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(16)}},
				}}},
			},
			depth:   0,
			wantErr: false,
		},
		{
			name: "group with both all_of and any_of",
			entries: []RuleEntry{
				{Group: &RuleGroup{
					AllOf: []RuleEntry{{Rule: &ValidationRule{Op: RuleContains, Value: "x"}}},
					AnyOf: []RuleEntry{{Rule: &ValidationRule{Op: RuleContains, Value: "y"}}},
				}},
			},
			depth:   0,
			wantErr: true,
		},
		{
			name: "group with empty all_of and empty any_of",
			entries: []RuleEntry{
				{Group: &RuleGroup{}},
			},
			depth:   0,
			wantErr: true,
		},
		{
			name: "required inside group rejected",
			entries: []RuleEntry{
				{Rule: &ValidationRule{Op: RuleRequired}},
			},
			depth:   1,
			wantErr: true,
		},
		{
			name: "required at top level accepted",
			entries: []RuleEntry{
				{Rule: &ValidationRule{Op: RuleRequired}},
			},
			depth:   0,
			wantErr: false,
		},
		{
			name:    "depth exceeds max",
			entries: []RuleEntry{{Rule: &ValidationRule{Op: RuleContains, Value: "x"}}},
			depth:   11,
			wantErr: true,
		},
		{
			name: "nested groups valid",
			entries: []RuleEntry{
				{Group: &RuleGroup{AllOf: []RuleEntry{
					{Group: &RuleGroup{AnyOf: []RuleEntry{
						{Rule: &ValidationRule{Op: RuleContains, Value: "x"}},
						{Rule: &ValidationRule{Op: RuleContains, Value: "y"}},
					}}},
				}}},
			},
			depth:   0,
			wantErr: false,
		},
		{
			name: "invalid rule inside group propagates error",
			entries: []RuleEntry{
				{Group: &RuleGroup{AllOf: []RuleEntry{
					{Rule: &ValidationRule{Op: "bogus"}},
				}}},
			},
			depth:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRuleEntries(tt.entries, tt.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRuleEntries() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// ValidateValidationConfig
// ============================================================

func TestValidateValidationConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  ValidationConfig
		wantErr bool
	}{
		{
			name:    "empty config",
			config:  ValidationConfig{},
			wantErr: false,
		},
		{
			name: "valid password strength config",
			config: ValidationConfig{Rules: []RuleEntry{
				{Rule: &ValidationRule{Op: RuleRequired}},
				{Rule: &ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(8)}},
				{Rule: &ValidationRule{Op: RuleContains, Class: ClassUppercase}},
				{Rule: &ValidationRule{Op: RuleContains, Class: ClassLowercase}},
				{Rule: &ValidationRule{Op: RuleContains, Class: ClassDigits}},
				{Group: &RuleGroup{AnyOf: []RuleEntry{
					{Rule: &ValidationRule{Op: RuleCount, Class: ClassSymbols, Cmp: CmpGte, N: floatPtr(1)}},
					{Rule: &ValidationRule{Op: RuleLength, Cmp: CmpGte, N: floatPtr(16)}},
				}}},
			}},
			wantErr: false,
		},
		{
			name: "invalid rule in config",
			config: ValidationConfig{Rules: []RuleEntry{
				{Rule: &ValidationRule{Op: "invalid"}},
			}},
			wantErr: true,
		},
		{
			name: "required inside nested group rejected",
			config: ValidationConfig{Rules: []RuleEntry{
				{Group: &RuleGroup{AllOf: []RuleEntry{
					{Rule: &ValidationRule{Op: RuleRequired}},
				}}},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateValidationConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValidationConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// ClassifyChar
// ============================================================

func TestClassifyChar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		char  rune
		class CharClass
		want  bool
	}{
		{name: "A is uppercase", char: 'A', class: ClassUppercase, want: true},
		{name: "Z is uppercase", char: 'Z', class: ClassUppercase, want: true},
		{name: "a is not uppercase", char: 'a', class: ClassUppercase, want: false},
		{name: "a is lowercase", char: 'a', class: ClassLowercase, want: true},
		{name: "z is lowercase", char: 'z', class: ClassLowercase, want: true},
		{name: "A is not lowercase", char: 'A', class: ClassLowercase, want: false},
		{name: "0 is digit", char: '0', class: ClassDigits, want: true},
		{name: "9 is digit", char: '9', class: ClassDigits, want: true},
		{name: "a is not digit", char: 'a', class: ClassDigits, want: false},
		{name: "! is symbol", char: '!', class: ClassSymbols, want: true},
		{name: "@ is symbol", char: '@', class: ClassSymbols, want: true},
		{name: "a is not symbol", char: 'a', class: ClassSymbols, want: false},
		{name: "space is not symbol", char: ' ', class: ClassSymbols, want: false},
		{name: "space is space", char: ' ', class: ClassSpaces, want: true},
		{name: "tab is space", char: '\t', class: ClassSpaces, want: true},
		{name: "newline is space", char: '\n', class: ClassSpaces, want: true},
		{name: "carriage return is space", char: '\r', class: ClassSpaces, want: true},
		{name: "a is not space", char: 'a', class: ClassSpaces, want: false},
		{name: "invalid class returns false", char: 'a', class: "bogus", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ClassifyChar(tt.char, tt.class)
			if got != tt.want {
				t.Errorf("ClassifyChar(%q, %q) = %v, want %v", tt.char, tt.class, got, tt.want)
			}
		})
	}
}

// ============================================================
// ParseUIConfig
// ============================================================

func TestParseUIConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(t *testing.T, uc UIConfig)
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
			checkFn: func(t *testing.T, uc UIConfig) {
				t.Helper()
				if uc.Widget != "" {
					t.Error("empty: Widget not empty")
				}
			},
		},
		{
			name:    "empty JSON",
			input:   "{}",
			wantErr: false,
		},
		{
			name:    "full config",
			input:   `{"widget":"textarea","placeholder":"Enter text...","help_text":"Max 500 chars","hidden":true}`,
			wantErr: false,
			checkFn: func(t *testing.T, uc UIConfig) {
				t.Helper()
				if uc.Widget != "textarea" {
					t.Errorf("Widget = %q", uc.Widget)
				}
				if uc.Placeholder != "Enter text..." {
					t.Errorf("Placeholder = %q", uc.Placeholder)
				}
				if uc.HelpText != "Max 500 chars" {
					t.Errorf("HelpText = %q", uc.HelpText)
				}
				if !uc.Hidden {
					t.Error("Hidden = false")
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   "not json",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, err := ParseUIConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseUIConfig(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, uc)
			}
		})
	}
}

// ============================================================
// ParseRelationConfig
// ============================================================

func TestParseRelationConfig(t *testing.T) {
	t.Parallel()
	validDatatypeID := generateValidULID()
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(t *testing.T, rc RelationConfig)
	}{
		{
			name:    "empty string errors",
			input:   "",
			wantErr: true,
		},
		{
			name:    "empty JSON object errors",
			input:   "{}",
			wantErr: true,
		},
		{
			name:    "missing target_datatype_id",
			input:   `{"cardinality":"one"}`,
			wantErr: true,
		},
		{
			name:    "invalid cardinality",
			input:   `{"target_datatype_id":"` + validDatatypeID + `","cardinality":"both"}`,
			wantErr: true,
		},
		{
			name:    "valid one",
			input:   `{"target_datatype_id":"` + validDatatypeID + `","cardinality":"one"}`,
			wantErr: false,
			checkFn: func(t *testing.T, rc RelationConfig) {
				t.Helper()
				if string(rc.TargetDatatypeID) != validDatatypeID {
					t.Errorf("TargetDatatypeID = %q", rc.TargetDatatypeID)
				}
				if rc.Cardinality != CardinalityOne {
					t.Errorf("Cardinality = %q", rc.Cardinality)
				}
				if rc.MaxDepth != nil {
					t.Errorf("MaxDepth = %v, want nil", rc.MaxDepth)
				}
			},
		},
		{
			name:    "valid many with max_depth",
			input:   `{"target_datatype_id":"` + validDatatypeID + `","cardinality":"many","max_depth":3}`,
			wantErr: false,
			checkFn: func(t *testing.T, rc RelationConfig) {
				t.Helper()
				if rc.Cardinality != CardinalityMany {
					t.Errorf("Cardinality = %q", rc.Cardinality)
				}
				if rc.MaxDepth == nil || *rc.MaxDepth != 3 {
					t.Error("MaxDepth != 3")
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   "{bad",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rc, err := ParseRelationConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRelationConfig(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, rc)
			}
		})
	}
}

// ============================================================
// ParseRichTextConfig
// ============================================================

func TestParseRichTextConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(t *testing.T, rc RichTextConfig)
	}{
		{
			name:    "empty string returns zero value",
			input:   "",
			wantErr: false,
			checkFn: func(t *testing.T, rc RichTextConfig) {
				t.Helper()
				if len(rc.Toolbar) != 0 {
					t.Errorf("empty: Toolbar = %v, want nil/empty", rc.Toolbar)
				}
			},
		},
		{
			name:    "empty JSON object returns zero value",
			input:   "{}",
			wantErr: false,
			checkFn: func(t *testing.T, rc RichTextConfig) {
				t.Helper()
				if len(rc.Toolbar) != 0 {
					t.Errorf("{}: Toolbar = %v, want nil/empty", rc.Toolbar)
				}
			},
		},
		{
			name:    "valid toolbar config",
			input:   `{"toolbar":["bold","italic","link"]}`,
			wantErr: false,
			checkFn: func(t *testing.T, rc RichTextConfig) {
				t.Helper()
				want := []string{"bold", "italic", "link"}
				if len(rc.Toolbar) != len(want) {
					t.Fatalf("Toolbar length = %d, want %d", len(rc.Toolbar), len(want))
				}
				for i, v := range want {
					if rc.Toolbar[i] != v {
						t.Errorf("Toolbar[%d] = %q, want %q", i, rc.Toolbar[i], v)
					}
				}
			},
		},
		{
			name:    "empty toolbar array",
			input:   `{"toolbar":[]}`,
			wantErr: false,
			checkFn: func(t *testing.T, rc RichTextConfig) {
				t.Helper()
				if len(rc.Toolbar) != 0 {
					t.Errorf("Toolbar = %v, want empty", rc.Toolbar)
				}
			},
		},
		{
			name:    "extra fields are ignored",
			input:   `{"toolbar":["bold"],"other":"ignored"}`,
			wantErr: false,
			checkFn: func(t *testing.T, rc RichTextConfig) {
				t.Helper()
				if len(rc.Toolbar) != 1 || rc.Toolbar[0] != "bold" {
					t.Errorf("Toolbar = %v, want [bold]", rc.Toolbar)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   "{bad",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rc, err := ParseRichTextConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRichTextConfig(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, rc)
			}
		})
	}
}
