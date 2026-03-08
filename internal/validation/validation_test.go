package validation

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func floatPtr(f float64) *float64 { return &f }

// validULID is a 26-character ULID for test inputs.
const validULID = "01ARYZ6S410000000000000000"

// ============================================================
// Type Validators
// ============================================================

func TestValidateType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		fieldType types.FieldType
		value     string
		data      string
		wantMsg   string // empty means pass
	}{
		// text/textarea/richtext: no type validation
		{name: "text valid", fieldType: types.FieldTypeText, value: "anything", wantMsg: ""},
		{name: "textarea valid", fieldType: types.FieldTypeTextarea, value: "<p>html</p>", wantMsg: ""},
		{name: "richtext valid", fieldType: types.FieldTypeRichText, value: "<b>bold</b>", wantMsg: ""},

		// number
		{name: "number valid int", fieldType: types.FieldTypeNumber, value: "42", wantMsg: ""},
		{name: "number valid float", fieldType: types.FieldTypeNumber, value: "3.14", wantMsg: ""},
		{name: "number valid negative", fieldType: types.FieldTypeNumber, value: "-10.5", wantMsg: ""},
		{name: "number valid zero", fieldType: types.FieldTypeNumber, value: "0", wantMsg: ""},
		{name: "number invalid", fieldType: types.FieldTypeNumber, value: "abc", wantMsg: "must be a valid number"},
		{name: "number empty spaces", fieldType: types.FieldTypeNumber, value: "  ", wantMsg: "must be a valid number"},

		// _id
		{name: "_id valid", fieldType: types.FieldTypeIDRef, value: validULID, wantMsg: ""},
		{name: "_id invalid", fieldType: types.FieldTypeIDRef, value: "not-a-ulid", wantMsg: "must be a valid content reference (ULID)"},
		{name: "_id too short", fieldType: types.FieldTypeIDRef, value: "01ARYZ6S41", wantMsg: "must be a valid content reference (ULID)"},

		// date
		{name: "date valid", fieldType: types.FieldTypeDate, value: "2024-01-15", wantMsg: ""},
		{name: "date invalid format", fieldType: types.FieldTypeDate, value: "01/15/2024", wantMsg: "must be a valid date (YYYY-MM-DD)"},
		{name: "date invalid text", fieldType: types.FieldTypeDate, value: "yesterday", wantMsg: "must be a valid date (YYYY-MM-DD)"},

		// datetime
		{name: "datetime valid rfc3339", fieldType: types.FieldTypeDatetime, value: "2024-01-15T10:30:00Z", wantMsg: ""},
		{name: "datetime valid rfc3339 offset", fieldType: types.FieldTypeDatetime, value: "2024-01-15T10:30:00+05:00", wantMsg: ""},
		{name: "datetime valid T format", fieldType: types.FieldTypeDatetime, value: "2024-01-15T10:30:00", wantMsg: ""},
		{name: "datetime valid space format", fieldType: types.FieldTypeDatetime, value: "2024-01-15 10:30:00", wantMsg: ""},
		{name: "datetime invalid", fieldType: types.FieldTypeDatetime, value: "not a datetime", wantMsg: "must be a valid datetime"},

		// boolean
		{name: "boolean true", fieldType: types.FieldTypeBoolean, value: "true", wantMsg: ""},
		{name: "boolean false", fieldType: types.FieldTypeBoolean, value: "false", wantMsg: ""},
		{name: "boolean 1", fieldType: types.FieldTypeBoolean, value: "1", wantMsg: ""},
		{name: "boolean 0", fieldType: types.FieldTypeBoolean, value: "0", wantMsg: ""},
		{name: "boolean invalid", fieldType: types.FieldTypeBoolean, value: "yes", wantMsg: "must be a boolean (true, false, 1, or 0)"},
		{name: "boolean True", fieldType: types.FieldTypeBoolean, value: "True", wantMsg: "must be a boolean (true, false, 1, or 0)"},

		// select
		{
			name:      "select valid",
			fieldType: types.FieldTypeSelect,
			value:     "red",
			data:      `[{"label":"Red","value":"red"},{"label":"Blue","value":"blue"}]`,
			wantMsg:   "",
		},
		{
			name:      "select invalid value",
			fieldType: types.FieldTypeSelect,
			value:     "green",
			data:      `[{"label":"Red","value":"red"},{"label":"Blue","value":"blue"}]`,
			wantMsg:   "must be one of the allowed options",
		},
		{
			name:      "select no options configured",
			fieldType: types.FieldTypeSelect,
			value:     "anything",
			data:      "",
			wantMsg:   "no select options configured",
		},
		{
			name:      "select empty JSON object",
			fieldType: types.FieldTypeSelect,
			value:     "anything",
			data:      "{}",
			wantMsg:   "no select options configured",
		},
		{
			name:      "select invalid data JSON",
			fieldType: types.FieldTypeSelect,
			value:     "anything",
			data:      "{bad json",
			wantMsg:   "invalid select options configuration",
		},

		// email
		{name: "email valid", fieldType: types.FieldTypeEmail, value: "user@example.com", wantMsg: ""},
		{name: "email invalid", fieldType: types.FieldTypeEmail, value: "not-an-email", wantMsg: "must be a valid email address"},
		{name: "email no domain", fieldType: types.FieldTypeEmail, value: "user@", wantMsg: "must be a valid email address"},

		// url
		{name: "url valid", fieldType: types.FieldTypeURL, value: "https://example.com", wantMsg: ""},
		{name: "url invalid no scheme", fieldType: types.FieldTypeURL, value: "example.com", wantMsg: "must be a valid URL"},
		{name: "url invalid no host", fieldType: types.FieldTypeURL, value: "https://", wantMsg: "must be a valid URL"},

		// slug
		{name: "slug valid root", fieldType: types.FieldTypeSlug, value: "/", wantMsg: ""},
		{name: "slug valid path", fieldType: types.FieldTypeSlug, value: "/about", wantMsg: ""},
		{name: "slug valid nested", fieldType: types.FieldTypeSlug, value: "/blog/my-post", wantMsg: ""},
		{name: "slug invalid no slash", fieldType: types.FieldTypeSlug, value: "about", wantMsg: "must be a valid slug"},
		{name: "slug invalid uppercase", fieldType: types.FieldTypeSlug, value: "/About", wantMsg: "must be a valid slug"},

		// media
		{name: "media valid", fieldType: types.FieldTypeMedia, value: validULID, wantMsg: ""},
		{name: "media invalid", fieldType: types.FieldTypeMedia, value: "not-ulid", wantMsg: "must be a valid media reference (ULID)"},

		// json
		{name: "json valid object", fieldType: types.FieldTypeJSON, value: `{"key":"value"}`, wantMsg: ""},
		{name: "json valid array", fieldType: types.FieldTypeJSON, value: `[1,2,3]`, wantMsg: ""},
		{name: "json valid string", fieldType: types.FieldTypeJSON, value: `"hello"`, wantMsg: ""},
		{name: "json invalid", fieldType: types.FieldTypeJSON, value: `{bad json`, wantMsg: "must be valid JSON"},

		// unknown field type: skip type validation
		{name: "unknown type passes", fieldType: "custom_widget", value: "anything", wantMsg: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validateType(tt.fieldType, tt.value, tt.data)
			if got != tt.wantMsg {
				t.Errorf("validateType(%q, %q, %q) = %q, want %q", tt.fieldType, tt.value, tt.data, got, tt.wantMsg)
			}
		})
	}
}

// ============================================================
// Required Rule Behavior
// ============================================================

func TestValidateField_Required(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      FieldInput
		wantNil    bool
		wantMsgSub string // substring expected in first message
	}{
		{
			name: "required and empty value",
			input: FieldInput{
				FieldID:    types.FieldID(validULID),
				Label:      "Title",
				FieldType:  types.FieldTypeText,
				Value:      "",
				Validation: `{"rules":[{"rule":{"op":"required"}}]}`,
			},
			wantNil:    false,
			wantMsgSub: "is required",
		},
		{
			name: "required and non-empty value",
			input: FieldInput{
				FieldID:    types.FieldID(validULID),
				Label:      "Title",
				FieldType:  types.FieldTypeText,
				Value:      "hello",
				Validation: `{"rules":[{"rule":{"op":"required"}}]}`,
			},
			wantNil: true,
		},
		{
			name: "no required rule and empty value skips all checks",
			input: FieldInput{
				FieldID:    types.FieldID(validULID),
				Label:      "Title",
				FieldType:  types.FieldTypeText,
				Value:      "",
				Validation: `{"rules":[{"rule":{"op":"length","cmp":"gte","n":5}}]}`,
			},
			wantNil: true,
		},
		{
			name: "no rules at all and empty value passes",
			input: FieldInput{
				FieldID:    types.FieldID(validULID),
				Label:      "Title",
				FieldType:  types.FieldTypeText,
				Value:      "",
				Validation: `{}`,
			},
			wantNil: true,
		},
		{
			name: "empty validation string and empty value passes",
			input: FieldInput{
				FieldID:   types.FieldID(validULID),
				Label:     "Title",
				FieldType: types.FieldTypeText,
				Value:     "",
			},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fe := ValidateField(tt.input)
			if tt.wantNil {
				if fe != nil {
					t.Errorf("expected nil FieldError, got: %v", fe.Messages)
				}
				return
			}
			if fe == nil {
				t.Fatal("expected FieldError, got nil")
			}
			if tt.wantMsgSub != "" {
				found := false
				for _, msg := range fe.Messages {
					if strings.Contains(msg, tt.wantMsgSub) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected message containing %q, got %v", tt.wantMsgSub, fe.Messages)
				}
			}
		})
	}
}

// ============================================================
// Batch Validation
// ============================================================

func TestValidateBatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		inputs     []FieldInput
		wantErrors int
	}{
		{
			name: "all valid",
			inputs: []FieldInput{
				{FieldID: types.FieldID(validULID), Label: "Title", FieldType: types.FieldTypeText, Value: "hello", Validation: `{"rules":[{"rule":{"op":"required"}}]}`},
				{FieldID: types.FieldID("01ARYZ6S410000000000000001"), Label: "Count", FieldType: types.FieldTypeNumber, Value: "42", Validation: `{}`},
			},
			wantErrors: 0,
		},
		{
			name: "mixed valid and invalid",
			inputs: []FieldInput{
				{FieldID: types.FieldID(validULID), Label: "Title", FieldType: types.FieldTypeText, Value: "", Validation: `{"rules":[{"rule":{"op":"required"}}]}`},
				{FieldID: types.FieldID("01ARYZ6S410000000000000001"), Label: "Count", FieldType: types.FieldTypeNumber, Value: "42", Validation: `{}`},
			},
			wantErrors: 1,
		},
		{
			name: "all invalid",
			inputs: []FieldInput{
				{FieldID: types.FieldID(validULID), Label: "Title", FieldType: types.FieldTypeText, Value: "", Validation: `{"rules":[{"rule":{"op":"required"}}]}`},
				{FieldID: types.FieldID("01ARYZ6S410000000000000001"), Label: "Count", FieldType: types.FieldTypeNumber, Value: "abc", Validation: `{"rules":[{"rule":{"op":"required"}}]}`},
			},
			wantErrors: 2,
		},
		{
			name:       "empty inputs",
			inputs:     nil,
			wantErrors: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ve := ValidateBatch(tt.inputs)
			if len(ve.Fields) != tt.wantErrors {
				t.Errorf("ValidateBatch() errors = %d, want %d", len(ve.Fields), tt.wantErrors)
			}
		})
	}
}

// ============================================================
// Error Type Methods
// ============================================================

func TestFieldError_Error(t *testing.T) {
	t.Parallel()
	fe := FieldError{
		FieldID:  types.FieldID(validULID),
		Label:    "Title",
		Messages: []string{"is required", "must be at least 5 characters"},
	}
	got := fe.Error()
	want := "Title: is required; must be at least 5 characters"
	if got != want {
		t.Errorf("FieldError.Error() = %q, want %q", got, want)
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	t.Parallel()
	empty := ValidationErrors{}
	if empty.HasErrors() {
		t.Error("empty ValidationErrors.HasErrors() = true, want false")
	}
	withErrors := ValidationErrors{Fields: []FieldError{{Label: "x", Messages: []string{"err"}}}}
	if !withErrors.HasErrors() {
		t.Error("non-empty ValidationErrors.HasErrors() = false, want true")
	}
}

func TestValidationErrors_ForField(t *testing.T) {
	t.Parallel()
	id1 := types.FieldID(validULID)
	id2 := types.FieldID("01ARYZ6S410000000000000001")
	ve := ValidationErrors{Fields: []FieldError{
		{FieldID: id1, Label: "Title", Messages: []string{"is required"}},
	}}

	got := ve.ForField(id1)
	if got == nil {
		t.Fatal("ForField(id1) = nil, want non-nil")
	}
	if got.Label != "Title" {
		t.Errorf("ForField(id1).Label = %q, want %q", got.Label, "Title")
	}

	got2 := ve.ForField(id2)
	if got2 != nil {
		t.Errorf("ForField(id2) = %v, want nil", got2)
	}
}

func TestValidationErrors_ClearField(t *testing.T) {
	t.Parallel()
	id1 := types.FieldID(validULID)
	id2 := types.FieldID("01ARYZ6S410000000000000001")
	ve := ValidationErrors{Fields: []FieldError{
		{FieldID: id1, Label: "Title", Messages: []string{"is required"}},
		{FieldID: id2, Label: "Count", Messages: []string{"must be a number"}},
	}}

	ve.ClearField(id1)
	if len(ve.Fields) != 1 {
		t.Fatalf("after ClearField: len = %d, want 1", len(ve.Fields))
	}
	if ve.Fields[0].FieldID != id2 {
		t.Errorf("remaining field ID = %q, want %q", ve.Fields[0].FieldID, id2)
	}

	// Clearing a non-existent field is a no-op.
	ve.ClearField(types.FieldID("01ARYZ6S410000000000000099"))
	if len(ve.Fields) != 1 {
		t.Errorf("ClearField non-existent: len = %d, want 1", len(ve.Fields))
	}
}

func TestValidationErrors_Error(t *testing.T) {
	t.Parallel()
	ve := ValidationErrors{}
	if ve.Error() != "validation passed" {
		t.Errorf("empty Error() = %q, want %q", ve.Error(), "validation passed")
	}
	ve.Fields = []FieldError{
		{FieldID: types.FieldID(validULID), Label: "Title", Messages: []string{"is required"}},
	}
	got := ve.Error()
	if !strings.Contains(got, "Title: is required") {
		t.Errorf("Error() = %q, want to contain %q", got, "Title: is required")
	}
}

func TestValidationErrors_JSON(t *testing.T) {
	t.Parallel()
	ve := ValidationErrors{Fields: []FieldError{
		{FieldID: types.FieldID(validULID), Label: "Title", Messages: []string{"is required"}},
	}}
	data, err := json.Marshal(ve)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var decoded ValidationErrors
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(decoded.Fields) != 1 {
		t.Fatalf("decoded fields = %d, want 1", len(decoded.Fields))
	}
	if decoded.Fields[0].Label != "Title" {
		t.Errorf("decoded label = %q, want %q", decoded.Fields[0].Label, "Title")
	}
}

// ============================================================
// Contains Rule
// ============================================================

func TestRule_Contains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
	}{
		{name: "literal match", value: "hello world", rule: types.ValidationRule{Op: types.RuleContains, Value: "world"}, wantPass: true},
		{name: "literal no match", value: "hello", rule: types.ValidationRule{Op: types.RuleContains, Value: "world"}, wantPass: false},
		{name: "class uppercase present", value: "helloWorld", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}, wantPass: true},
		{name: "class uppercase absent", value: "helloworld", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}, wantPass: false},
		{name: "class lowercase present", value: "Hello", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassLowercase}, wantPass: true},
		{name: "class digits present", value: "abc123", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}, wantPass: true},
		{name: "class digits absent", value: "abcdef", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}, wantPass: false},
		{name: "class symbols present", value: "hello!", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassSymbols}, wantPass: true},
		{name: "class symbols absent", value: "hello world", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassSymbols}, wantPass: false},
		{name: "class spaces present", value: "hello world", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassSpaces}, wantPass: true},
		{name: "class spaces absent", value: "helloworld", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassSpaces}, wantPass: false},
		{name: "negate literal match becomes fail", value: "hello world", rule: types.ValidationRule{Op: types.RuleContains, Value: "world", Negate: true}, wantPass: false},
		{name: "negate literal no match becomes pass", value: "hello", rule: types.ValidationRule{Op: types.RuleContains, Value: "world", Negate: true}, wantPass: true},
		{name: "negate class present becomes fail", value: "ABC", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase, Negate: true}, wantPass: false},
		{name: "negate class absent becomes pass", value: "abc", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase, Negate: true}, wantPass: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q, contains) pass=%v, want %v, msgs=%v", tt.value, gotPass, tt.wantPass, msgs)
			}
		})
	}
}

// ============================================================
// StartsWith / EndsWith Rule
// ============================================================

func TestRule_StartsWithEndsWith(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
	}{
		// starts_with literal
		{name: "starts_with literal match", value: "hello", rule: types.ValidationRule{Op: types.RuleStartsWith, Value: "hel"}, wantPass: true},
		{name: "starts_with literal no match", value: "hello", rule: types.ValidationRule{Op: types.RuleStartsWith, Value: "xyz"}, wantPass: false},
		{name: "starts_with literal negated match", value: "hello", rule: types.ValidationRule{Op: types.RuleStartsWith, Value: "hel", Negate: true}, wantPass: false},
		{name: "starts_with literal negated no match", value: "hello", rule: types.ValidationRule{Op: types.RuleStartsWith, Value: "xyz", Negate: true}, wantPass: true},
		// starts_with class
		{name: "starts_with class uppercase match", value: "Hello", rule: types.ValidationRule{Op: types.RuleStartsWith, Class: types.ClassUppercase}, wantPass: true},
		{name: "starts_with class uppercase no match", value: "hello", rule: types.ValidationRule{Op: types.RuleStartsWith, Class: types.ClassUppercase}, wantPass: false},
		{name: "starts_with class digit match", value: "1abc", rule: types.ValidationRule{Op: types.RuleStartsWith, Class: types.ClassDigits}, wantPass: true},
		{name: "starts_with empty value", value: "", rule: types.ValidationRule{Op: types.RuleStartsWith, Class: types.ClassUppercase}, wantPass: false},
		// ends_with literal
		{name: "ends_with literal match", value: "hello.com", rule: types.ValidationRule{Op: types.RuleEndsWith, Value: ".com"}, wantPass: true},
		{name: "ends_with literal no match", value: "hello.org", rule: types.ValidationRule{Op: types.RuleEndsWith, Value: ".com"}, wantPass: false},
		{name: "ends_with literal negated match", value: "hello.com", rule: types.ValidationRule{Op: types.RuleEndsWith, Value: ".com", Negate: true}, wantPass: false},
		{name: "ends_with literal negated no match", value: "hello.org", rule: types.ValidationRule{Op: types.RuleEndsWith, Value: ".com", Negate: true}, wantPass: true},
		// ends_with class
		{name: "ends_with class digit match", value: "abc9", rule: types.ValidationRule{Op: types.RuleEndsWith, Class: types.ClassDigits}, wantPass: true},
		{name: "ends_with class digit no match", value: "abcx", rule: types.ValidationRule{Op: types.RuleEndsWith, Class: types.ClassDigits}, wantPass: false},
		{name: "ends_with class lowercase match", value: "ABCz", rule: types.ValidationRule{Op: types.RuleEndsWith, Class: types.ClassLowercase}, wantPass: true},
		{name: "ends_with empty value", value: "", rule: types.ValidationRule{Op: types.RuleEndsWith, Class: types.ClassLowercase}, wantPass: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q) pass=%v, want %v, msgs=%v", tt.value, gotPass, tt.wantPass, msgs)
			}
		})
	}
}

// ============================================================
// Equals Rule
// ============================================================

func TestRule_Equals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
	}{
		{name: "exact match", value: "hello", rule: types.ValidationRule{Op: types.RuleEquals, Value: "hello"}, wantPass: true},
		{name: "case mismatch", value: "Hello", rule: types.ValidationRule{Op: types.RuleEquals, Value: "hello"}, wantPass: false},
		{name: "no match", value: "world", rule: types.ValidationRule{Op: types.RuleEquals, Value: "hello"}, wantPass: false},
		{name: "negated match", value: "hello", rule: types.ValidationRule{Op: types.RuleEquals, Value: "hello", Negate: true}, wantPass: false},
		{name: "negated no match", value: "world", rule: types.ValidationRule{Op: types.RuleEquals, Value: "hello", Negate: true}, wantPass: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q, equals) pass=%v, want %v", tt.value, gotPass, tt.wantPass)
			}
		})
	}
}

// ============================================================
// Length Rule
// ============================================================

func TestRule_Length(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
	}{
		{name: "eq pass", value: "hello", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpEq, N: floatPtr(5)}, wantPass: true},
		{name: "eq fail", value: "hi", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpEq, N: floatPtr(5)}, wantPass: false},
		{name: "neq pass", value: "hi", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpNeq, N: floatPtr(5)}, wantPass: true},
		{name: "neq fail", value: "hello", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpNeq, N: floatPtr(5)}, wantPass: false},
		{name: "gt pass", value: "hello!", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGt, N: floatPtr(5)}, wantPass: true},
		{name: "gt boundary fail", value: "hello", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGt, N: floatPtr(5)}, wantPass: false},
		{name: "gte pass equal", value: "hello", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGte, N: floatPtr(5)}, wantPass: true},
		{name: "gte pass above", value: "hello!", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGte, N: floatPtr(5)}, wantPass: true},
		{name: "gte fail", value: "hi", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGte, N: floatPtr(5)}, wantPass: false},
		{name: "lt pass", value: "hi", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpLt, N: floatPtr(5)}, wantPass: true},
		{name: "lt boundary fail", value: "hello", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpLt, N: floatPtr(5)}, wantPass: false},
		{name: "lte pass equal", value: "hello", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpLte, N: floatPtr(5)}, wantPass: true},
		{name: "lte fail", value: "hello!", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpLte, N: floatPtr(5)}, wantPass: false},
		// Unicode: rune count, not byte count
		{name: "unicode rune count", value: "\u00e9\u00e9\u00e9", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpEq, N: floatPtr(3)}, wantPass: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q, length) pass=%v, want %v", tt.value, gotPass, tt.wantPass)
			}
		})
	}
}

// ============================================================
// Count Rule
// ============================================================

func TestRule_Count(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
	}{
		{name: "count literal match", value: "banana", rule: types.ValidationRule{Op: types.RuleCount, Value: "a", Cmp: types.CmpEq, N: floatPtr(3)}, wantPass: true},
		{name: "count literal no match", value: "banana", rule: types.ValidationRule{Op: types.RuleCount, Value: "a", Cmp: types.CmpEq, N: floatPtr(2)}, wantPass: false},
		{name: "count literal gte", value: "banana", rule: types.ValidationRule{Op: types.RuleCount, Value: "a", Cmp: types.CmpGte, N: floatPtr(2)}, wantPass: true},
		{name: "count class digits", value: "abc123def", rule: types.ValidationRule{Op: types.RuleCount, Class: types.ClassDigits, Cmp: types.CmpEq, N: floatPtr(3)}, wantPass: true},
		{name: "count class digits fail", value: "abc123def", rule: types.ValidationRule{Op: types.RuleCount, Class: types.ClassDigits, Cmp: types.CmpEq, N: floatPtr(2)}, wantPass: false},
		{name: "count class symbols", value: "a!b@c#", rule: types.ValidationRule{Op: types.RuleCount, Class: types.ClassSymbols, Cmp: types.CmpGte, N: floatPtr(3)}, wantPass: true},
		{name: "count class spaces", value: "a b c", rule: types.ValidationRule{Op: types.RuleCount, Class: types.ClassSpaces, Cmp: types.CmpEq, N: floatPtr(2)}, wantPass: true},
		{name: "count zero occurrences", value: "abc", rule: types.ValidationRule{Op: types.RuleCount, Value: "x", Cmp: types.CmpEq, N: floatPtr(0)}, wantPass: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q, count) pass=%v, want %v", tt.value, gotPass, tt.wantPass)
			}
		})
	}
}

// ============================================================
// Range Rule
// ============================================================

func TestRule_Range(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
		wantMsg  string // if specific message expected
	}{
		{name: "gte pass", value: "10", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(5)}, wantPass: true},
		{name: "gte boundary", value: "5", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(5)}, wantPass: true},
		{name: "gte fail", value: "3", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(5)}, wantPass: false},
		{name: "lte pass", value: "5", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpLte, N: floatPtr(10)}, wantPass: true},
		{name: "lte fail", value: "11", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpLte, N: floatPtr(10)}, wantPass: false},
		{name: "eq pass", value: "42", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpEq, N: floatPtr(42)}, wantPass: true},
		{name: "neq pass", value: "43", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpNeq, N: floatPtr(42)}, wantPass: true},
		{name: "gt pass", value: "6", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGt, N: floatPtr(5)}, wantPass: true},
		{name: "lt pass", value: "4", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpLt, N: floatPtr(5)}, wantPass: true},
		{name: "decimal comparison", value: "0.01", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(0.01)}, wantPass: true},
		{name: "negative value", value: "-5", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(-10)}, wantPass: true},
		{name: "not a number", value: "abc", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(0)}, wantPass: false, wantMsg: "must be a number"},
		{name: "not a number custom msg", value: "abc", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(0), Message: "enter a numeric value"}, wantPass: false, wantMsg: "enter a numeric value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q, range) pass=%v, want %v, msgs=%v", tt.value, gotPass, tt.wantPass, msgs)
			}
			if tt.wantMsg != "" && len(msgs) > 0 && msgs[0] != tt.wantMsg {
				t.Errorf("message = %q, want %q", msgs[0], tt.wantMsg)
			}
		})
	}
}

// ============================================================
// ItemCount Rule
// ============================================================

func TestRule_ItemCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
	}{
		{name: "json array 3 items eq", value: `["a","b","c"]`, rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(3)}, wantPass: true},
		{name: "json array 3 items lte 5", value: `["a","b","c"]`, rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpLte, N: floatPtr(5)}, wantPass: true},
		{name: "json array too many", value: `[1,2,3,4,5,6]`, rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpLte, N: floatPtr(5)}, wantPass: false},
		{name: "comma separated", value: "red, green, blue", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(3)}, wantPass: true},
		{name: "comma separated no spaces", value: "a,b,c", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(3)}, wantPass: true},
		{name: "empty value is 0", value: "", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(0)}, wantPass: true},
		{name: "empty value not 1", value: "", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(1)}, wantPass: false},
		{name: "malformed json fallback to csv", value: "[broken json, with commas", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(2)}, wantPass: true},
		{name: "whitespace only segments skipped", value: "a, , b, ,c", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(3)}, wantPass: true},
		{name: "single item", value: "single", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpEq, N: floatPtr(1)}, wantPass: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q, item_count) pass=%v, want %v, msgs=%v", tt.value, gotPass, tt.wantPass, msgs)
			}
		})
	}
}

// ============================================================
// OneOf Rule
// ============================================================

func TestRule_OneOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		rule     types.ValidationRule
		wantPass bool
	}{
		{name: "in set", value: "b", rule: types.ValidationRule{Op: types.RuleOneOf, Values: []string{"a", "b", "c"}}, wantPass: true},
		{name: "not in set", value: "d", rule: types.ValidationRule{Op: types.RuleOneOf, Values: []string{"a", "b", "c"}}, wantPass: false},
		{name: "case sensitive", value: "A", rule: types.ValidationRule{Op: types.RuleOneOf, Values: []string{"a", "b", "c"}}, wantPass: false},
		{name: "negated in set", value: "b", rule: types.ValidationRule{Op: types.RuleOneOf, Values: []string{"a", "b", "c"}, Negate: true}, wantPass: false},
		{name: "negated not in set", value: "d", rule: types.ValidationRule{Op: types.RuleOneOf, Values: []string{"a", "b", "c"}, Negate: true}, wantPass: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateRule(tt.value, tt.rule)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateRule(%q, one_of) pass=%v, want %v", tt.value, gotPass, tt.wantPass)
			}
		})
	}
}

// ============================================================
// AllOf / AnyOf Groups
// ============================================================

func TestRule_AllOfGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		value      string
		group      types.RuleGroup
		wantMsgCnt int
	}{
		{
			name:  "all pass",
			value: "Hello123",
			group: types.RuleGroup{AllOf: []types.RuleEntry{
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}},
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}},
			}},
			wantMsgCnt: 0,
		},
		{
			name:  "one fails",
			value: "hello",
			group: types.RuleGroup{AllOf: []types.RuleEntry{
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassLowercase}},
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}},
			}},
			wantMsgCnt: 1,
		},
		{
			name:  "all fail",
			value: "hello",
			group: types.RuleGroup{AllOf: []types.RuleEntry{
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}},
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}},
			}},
			wantMsgCnt: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateGroup(tt.value, tt.group)
			if len(msgs) != tt.wantMsgCnt {
				t.Errorf("evaluateGroup AllOf: msgs=%d, want %d; msgs=%v", len(msgs), tt.wantMsgCnt, msgs)
			}
		})
	}
}

func TestRule_AnyOfGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		group    types.RuleGroup
		wantPass bool
	}{
		{
			name:  "all pass",
			value: "Hello123!",
			group: types.RuleGroup{AnyOf: []types.RuleEntry{
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}},
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}},
			}},
			wantPass: true,
		},
		{
			name:  "one passes",
			value: "hello",
			group: types.RuleGroup{AnyOf: []types.RuleEntry{
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}},
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassLowercase}},
			}},
			wantPass: true,
		},
		{
			name:  "all fail returns first failure",
			value: "hello",
			group: types.RuleGroup{AnyOf: []types.RuleEntry{
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}},
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}},
			}},
			wantPass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msgs := evaluateGroup(tt.value, tt.group)
			gotPass := len(msgs) == 0
			if gotPass != tt.wantPass {
				t.Errorf("evaluateGroup AnyOf: pass=%v, want %v, msgs=%v", gotPass, tt.wantPass, msgs)
			}
		})
	}
}

// ============================================================
// Nested Groups
// ============================================================

func TestRule_NestedGroups(t *testing.T) {
	t.Parallel()
	// allOf(contains uppercase, anyOf(contains digits, length >= 16))
	entries := []types.RuleEntry{
		{Group: &types.RuleGroup{AllOf: []types.RuleEntry{
			{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassUppercase}},
			{Group: &types.RuleGroup{AnyOf: []types.RuleEntry{
				{Rule: &types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}},
				{Rule: &types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGte, N: floatPtr(16)}},
			}}},
		}}},
	}

	// "Hello123" -> uppercase present, digits present -> pass
	msgs := EvaluateRules("Hello123", entries)
	if len(msgs) != 0 {
		t.Errorf("Hello123: want pass, got msgs=%v", msgs)
	}

	// "Helloworldtest!!" (16 chars, uppercase, no digits) -> pass via length >= 16
	msgs = EvaluateRules("Helloworldtest!!", entries)
	if len(msgs) != 0 {
		t.Errorf("Helloworldtest!!: want pass, got msgs=%v", msgs)
	}

	// "hello" -> no uppercase -> fail
	msgs = EvaluateRules("hello", entries)
	if len(msgs) == 0 {
		t.Error("hello: want fail, got pass")
	}

	// "Hello" -> uppercase present, no digits, length < 16 -> anyOf fails
	msgs = EvaluateRules("Hello", entries)
	if len(msgs) == 0 {
		t.Error("Hello: want fail (anyOf fails), got pass")
	}
}

// ============================================================
// Custom Message Override
// ============================================================

func TestRule_CustomMessage(t *testing.T) {
	t.Parallel()
	rule := types.ValidationRule{
		Op:      types.RuleContains,
		Value:   "required-text",
		Message: "you forgot the magic word",
	}
	msgs := evaluateRule("hello", rule)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0] != "you forgot the magic word" {
		t.Errorf("message = %q, want %q", msgs[0], "you forgot the magic word")
	}
}

// ============================================================
// Malformed Validation JSON
// ============================================================

func TestValidateField_MalformedJSON(t *testing.T) {
	t.Parallel()
	input := FieldInput{
		FieldID:    types.FieldID(validULID),
		Label:      "Title",
		FieldType:  types.FieldTypeText,
		Value:      "hello",
		Validation: "{bad json",
	}
	fe := ValidateField(input)
	if fe == nil {
		t.Fatal("expected FieldError for malformed JSON, got nil")
	}
	if len(fe.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	if !strings.Contains(fe.Messages[0], "invalid validation configuration") {
		t.Errorf("message = %q, want to contain %q", fe.Messages[0], "invalid validation configuration")
	}
}

// ============================================================
// Unknown Field Type
// ============================================================

func TestValidateField_UnknownFieldType(t *testing.T) {
	t.Parallel()
	// Unknown field type should skip type validation but still run composable rules.
	input := FieldInput{
		FieldID:    types.FieldID(validULID),
		Label:      "Custom",
		FieldType:  "custom_widget",
		Value:      "short",
		Validation: `{"rules":[{"rule":{"op":"length","cmp":"gte","n":10}}]}`,
	}
	fe := ValidateField(input)
	if fe == nil {
		t.Fatal("expected FieldError (length rule fails), got nil")
	}
	// Passes type validation, fails length rule.
	found := false
	for _, msg := range fe.Messages {
		if strings.Contains(msg, "characters") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected length error message, got %v", fe.Messages)
	}

	// Unknown field type with passing rules.
	input2 := FieldInput{
		FieldID:    types.FieldID(validULID),
		Label:      "Custom",
		FieldType:  "custom_widget",
		Value:      "long enough value",
		Validation: `{"rules":[{"rule":{"op":"length","cmp":"gte","n":5}}]}`,
	}
	fe2 := ValidateField(input2)
	if fe2 != nil {
		t.Errorf("expected nil for passing unknown type, got %v", fe2.Messages)
	}
}

// ============================================================
// Empty Rules Slice
// ============================================================

func TestEvaluateRules_Empty(t *testing.T) {
	t.Parallel()
	msgs := EvaluateRules("anything", nil)
	if len(msgs) != 0 {
		t.Errorf("empty rules: msgs=%v, want empty", msgs)
	}
	msgs = EvaluateRules("anything", []types.RuleEntry{})
	if len(msgs) != 0 {
		t.Errorf("zero-length rules: msgs=%v, want empty", msgs)
	}
}

// ============================================================
// Required Op Is Skipped in EvaluateRules
// ============================================================

func TestEvaluateRules_SkipsRequired(t *testing.T) {
	t.Parallel()
	entries := []types.RuleEntry{
		{Rule: &types.ValidationRule{Op: types.RuleRequired}},
		{Rule: &types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGte, N: floatPtr(5)}},
	}
	// Even with empty value, EvaluateRules skips required; only length is evaluated.
	msgs := EvaluateRules("hi", entries)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message (length fail), got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "characters") {
		t.Errorf("expected length error, got %q", msgs[0])
	}
}

// ============================================================
// Default Error Messages
// ============================================================

func TestDefaultMessages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		rule    types.ValidationRule
		wantSub string
	}{
		{name: "required", rule: types.ValidationRule{Op: types.RuleRequired}, wantSub: "is required"},
		{name: "contains literal", rule: types.ValidationRule{Op: types.RuleContains, Value: "abc"}, wantSub: `must contain "abc"`},
		{name: "contains class", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits}, wantSub: "must contain digits characters"},
		{name: "contains negated literal", rule: types.ValidationRule{Op: types.RuleContains, Value: "abc", Negate: true}, wantSub: `must not contain "abc"`},
		{name: "contains negated class", rule: types.ValidationRule{Op: types.RuleContains, Class: types.ClassDigits, Negate: true}, wantSub: "must not contain digits characters"},
		{name: "starts_with literal", rule: types.ValidationRule{Op: types.RuleStartsWith, Value: "http"}, wantSub: `must start with "http"`},
		{name: "starts_with negated", rule: types.ValidationRule{Op: types.RuleStartsWith, Value: "http", Negate: true}, wantSub: `must not start with "http"`},
		{name: "ends_with literal", rule: types.ValidationRule{Op: types.RuleEndsWith, Value: ".com"}, wantSub: `must end with ".com"`},
		{name: "ends_with negated", rule: types.ValidationRule{Op: types.RuleEndsWith, Value: ".com", Negate: true}, wantSub: `must not end with ".com"`},
		{name: "equals", rule: types.ValidationRule{Op: types.RuleEquals, Value: "exact"}, wantSub: `must equal "exact"`},
		{name: "equals negated", rule: types.ValidationRule{Op: types.RuleEquals, Value: "exact", Negate: true}, wantSub: `must not equal "exact"`},
		{name: "length gte", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpGte, N: floatPtr(10)}, wantSub: "must be at least 10 characters"},
		{name: "length lte", rule: types.ValidationRule{Op: types.RuleLength, Cmp: types.CmpLte, N: floatPtr(100)}, wantSub: "must be at most 100 characters"},
		{name: "count literal", rule: types.ValidationRule{Op: types.RuleCount, Value: "x", Cmp: types.CmpGt, N: floatPtr(2)}, wantSub: `must have more than 2 occurrences of "x"`},
		{name: "count class", rule: types.ValidationRule{Op: types.RuleCount, Class: types.ClassDigits, Cmp: types.CmpGte, N: floatPtr(3)}, wantSub: "must have at least 3 occurrences of digits characters"},
		{name: "range gte", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpGte, N: floatPtr(0)}, wantSub: "value must be at least 0"},
		{name: "range lte decimal", rule: types.ValidationRule{Op: types.RuleRange, Cmp: types.CmpLte, N: floatPtr(99.99)}, wantSub: "value must be at most 99.99"},
		{name: "item_count lte", rule: types.ValidationRule{Op: types.RuleItemCount, Cmp: types.CmpLte, N: floatPtr(5)}, wantSub: "must have at most 5 items"},
		{name: "one_of", rule: types.ValidationRule{Op: types.RuleOneOf, Values: []string{"a", "b", "c"}}, wantSub: "must be one of: a, b, c"},
		{name: "one_of negated", rule: types.ValidationRule{Op: types.RuleOneOf, Values: []string{"x"}, Negate: true}, wantSub: "must not be one of: x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := defaultMessage(tt.rule)
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("defaultMessage() = %q, want to contain %q", got, tt.wantSub)
			}
		})
	}
}

// ============================================================
// Integration: ValidateField with Type + Rules
// ============================================================

func TestValidateField_TypeAndRules(t *testing.T) {
	t.Parallel()

	// Number field with range rules: 0 <= value <= 100
	input := FieldInput{
		FieldID:    types.FieldID(validULID),
		Label:      "Score",
		FieldType:  types.FieldTypeNumber,
		Value:      "abc",
		Validation: `{"rules":[{"rule":{"op":"range","cmp":"gte","n":0}},{"rule":{"op":"range","cmp":"lte","n":100}}]}`,
	}
	fe := ValidateField(input)
	if fe == nil {
		t.Fatal("expected errors for non-numeric value, got nil")
	}
	// Should have type validation error + range errors
	if len(fe.Messages) < 1 {
		t.Errorf("expected at least 1 message, got %d", len(fe.Messages))
	}

	// Valid number within range
	input2 := FieldInput{
		FieldID:    types.FieldID(validULID),
		Label:      "Score",
		FieldType:  types.FieldTypeNumber,
		Value:      "50",
		Validation: `{"rules":[{"rule":{"op":"range","cmp":"gte","n":0}},{"rule":{"op":"range","cmp":"lte","n":100}}]}`,
	}
	fe2 := ValidateField(input2)
	if fe2 != nil {
		t.Errorf("expected nil for valid score, got %v", fe2.Messages)
	}

	// Number out of range
	input3 := FieldInput{
		FieldID:    types.FieldID(validULID),
		Label:      "Score",
		FieldType:  types.FieldTypeNumber,
		Value:      "150",
		Validation: `{"rules":[{"rule":{"op":"range","cmp":"gte","n":0}},{"rule":{"op":"range","cmp":"lte","n":100}}]}`,
	}
	fe3 := ValidateField(input3)
	if fe3 == nil {
		t.Fatal("expected error for out-of-range value, got nil")
	}
}

// ============================================================
// Integration: Password Strength Example
// ============================================================

func TestValidateField_PasswordStrength(t *testing.T) {
	t.Parallel()
	validation := `{
		"rules": [
			{"rule": {"op": "required"}},
			{"rule": {"op": "length", "cmp": "gte", "n": 8}},
			{"rule": {"op": "contains", "class": "uppercase"}},
			{"rule": {"op": "contains", "class": "lowercase"}},
			{"rule": {"op": "contains", "class": "digits"}},
			{"group": {"any_of": [
				{"rule": {"op": "count", "class": "symbols", "cmp": "gte", "n": 1}},
				{"rule": {"op": "length", "cmp": "gte", "n": 16}}
			]}}
		]
	}`

	tests := []struct {
		name     string
		value    string
		wantPass bool
	}{
		{name: "strong with symbol", value: "MyP@ssw0rd", wantPass: true},
		{name: "strong with length", value: "MyPasswordIsVeryLong1", wantPass: true},
		{name: "too short", value: "Ab1!", wantPass: false},
		{name: "no uppercase", value: "myp@ssw0rd", wantPass: false},
		{name: "no lowercase", value: "MYP@SSW0RD", wantPass: false},
		{name: "no digit", value: "MyP@ssword", wantPass: false},
		{name: "no symbol and too short", value: "MyPassw1", wantPass: false},
		{name: "empty required", value: "", wantPass: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := FieldInput{
				FieldID:    types.FieldID(validULID),
				Label:      "Password",
				FieldType:  types.FieldTypeText,
				Value:      tt.value,
				Validation: validation,
			}
			fe := ValidateField(input)
			gotPass := fe == nil
			if gotPass != tt.wantPass {
				msgs := []string{"nil"}
				if fe != nil {
					msgs = fe.Messages
				}
				t.Errorf("password %q: pass=%v, want %v, msgs=%v", tt.value, gotPass, tt.wantPass, msgs)
			}
		})
	}
}
