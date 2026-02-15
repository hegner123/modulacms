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
		name       string
		input      string
		wantErr    bool
		checkFn    func(t *testing.T, vc ValidationConfig)
	}{
		{
			name:    "empty string returns zero",
			input:   "",
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				if vc.Required {
					t.Error("empty input: Required = true")
				}
			},
		},
		{
			name:    "empty JSON object returns zero",
			input:   "{}",
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				if vc.Required {
					t.Error("{} input: Required = true")
				}
			},
		},
		{
			name:    "full config",
			input:   `{"required":true,"min_length":1,"max_length":100,"min":0,"max":999,"pattern":"^[a-z]+$","max_items":10}`,
			wantErr: false,
			checkFn: func(t *testing.T, vc ValidationConfig) {
				t.Helper()
				if !vc.Required {
					t.Error("Required = false")
				}
				if vc.MinLength == nil || *vc.MinLength != 1 {
					t.Error("MinLength != 1")
				}
				if vc.MaxLength == nil || *vc.MaxLength != 100 {
					t.Error("MaxLength != 100")
				}
				if vc.Min == nil || *vc.Min != 0 {
					t.Error("Min != 0")
				}
				if vc.Max == nil || *vc.Max != 999 {
					t.Error("Max != 999")
				}
				if vc.Pattern != "^[a-z]+$" {
					t.Errorf("Pattern = %q", vc.Pattern)
				}
				if vc.MaxItems == nil || *vc.MaxItems != 10 {
					t.Error("MaxItems != 10")
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
