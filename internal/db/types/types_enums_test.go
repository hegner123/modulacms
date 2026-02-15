package types

import (
	"encoding/json"
	"testing"
)

// enumContract captures the common behavior every enum type must satisfy.
type enumContract struct {
	name        string
	validValues []string
	validateFn  func(string) error
	valueFn     func(string) (any, error)
	scanFn      func(any) (string, error) // scan input -> resulting string
}

func allEnumContracts() []enumContract {
	return []enumContract{
		{
			name:        "ContentStatus",
			validValues: []string{"draft", "published", "archived", "pending"},
			validateFn:  func(s string) error { return ContentStatus(s).Validate() },
			valueFn:     func(s string) (any, error) { return ContentStatus(s).Value() },
			scanFn: func(v any) (string, error) {
				var cs ContentStatus
				err := cs.Scan(v)
				return string(cs), err
			},
		},
		{
			name:        "FieldType",
			validValues: []string{"text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "relation", "json", "richtext", "slug", "email", "url"},
			validateFn:  func(s string) error { return FieldType(s).Validate() },
			valueFn:     func(s string) (any, error) { return FieldType(s).Value() },
			scanFn: func(v any) (string, error) {
				var ft FieldType
				err := ft.Scan(v)
				return string(ft), err
			},
		},
		{
			name:        "RouteType",
			validValues: []string{"static", "dynamic", "api", "redirect"},
			validateFn:  func(s string) error { return RouteType(s).Validate() },
			valueFn:     func(s string) (any, error) { return RouteType(s).Value() },
			scanFn: func(v any) (string, error) {
				var rt RouteType
				err := rt.Scan(v)
				return string(rt), err
			},
		},
		{
			name:        "Operation",
			validValues: []string{"INSERT", "UPDATE", "DELETE"},
			validateFn:  func(s string) error { return Operation(s).Validate() },
			valueFn:     func(s string) (any, error) { return Operation(s).Value() },
			scanFn: func(v any) (string, error) {
				var o Operation
				err := o.Scan(v)
				return string(o), err
			},
		},
		{
			name:        "Action",
			validValues: []string{"create", "update", "delete", "publish", "archive"},
			validateFn:  func(s string) error { return Action(s).Validate() },
			valueFn:     func(s string) (any, error) { return Action(s).Value() },
			scanFn: func(v any) (string, error) {
				var a Action
				err := a.Scan(v)
				return string(a), err
			},
		},
		{
			name:        "ConflictPolicy",
			validValues: []string{"lww", "manual"},
			validateFn:  func(s string) error { return ConflictPolicy(s).Validate() },
			valueFn:     func(s string) (any, error) { return ConflictPolicy(s).Value() },
			scanFn: func(v any) (string, error) {
				var cp ConflictPolicy
				err := cp.Scan(v)
				return string(cp), err
			},
		},
		{
			name:        "BackupType",
			validValues: []string{"full", "incremental", "differential"},
			validateFn:  func(s string) error { return BackupType(s).Validate() },
			valueFn:     func(s string) (any, error) { return BackupType(s).Value() },
			scanFn: func(v any) (string, error) {
				var bt BackupType
				err := bt.Scan(v)
				return string(bt), err
			},
		},
		{
			name:        "BackupStatus",
			validValues: []string{"pending", "in_progress", "completed", "failed"},
			validateFn:  func(s string) error { return BackupStatus(s).Validate() },
			valueFn:     func(s string) (any, error) { return BackupStatus(s).Value() },
			scanFn: func(v any) (string, error) {
				var bs BackupStatus
				err := bs.Scan(v)
				return string(bs), err
			},
		},
		{
			name:        "VerificationStatus",
			validValues: []string{"pending", "verified", "failed"},
			validateFn:  func(s string) error { return VerificationStatus(s).Validate() },
			valueFn:     func(s string) (any, error) { return VerificationStatus(s).Value() },
			scanFn: func(v any) (string, error) {
				var vs VerificationStatus
				err := vs.Scan(v)
				return string(vs), err
			},
		},
		{
			name:        "BackupSetStatus",
			validValues: []string{"pending", "complete", "partial"},
			validateFn:  func(s string) error { return BackupSetStatus(s).Validate() },
			valueFn:     func(s string) (any, error) { return BackupSetStatus(s).Value() },
			scanFn: func(v any) (string, error) {
				var bss BackupSetStatus
				err := bss.Scan(v)
				return string(bss), err
			},
		},
	}
}

func TestEnums_Validate_ValidValues(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, v := range c.validValues {
				t.Run(v, func(t *testing.T) {
					t.Parallel()
					if err := c.validateFn(v); err != nil {
						t.Errorf("%s(%q).Validate() = %v", c.name, v, err)
					}
				})
			}
		})
	}
}

func TestEnums_Validate_Empty(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if err := c.validateFn(""); err == nil {
				t.Errorf("%s(\"\").Validate() = nil, want error", c.name)
			}
		})
	}
}

func TestEnums_Validate_Invalid(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if err := c.validateFn("DEFINITELY_NOT_VALID"); err == nil {
				t.Errorf("%s(\"DEFINITELY_NOT_VALID\").Validate() = nil, want error", c.name)
			}
		})
	}
}

func TestEnums_Value_Valid(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, v := range c.validValues {
				t.Run(v, func(t *testing.T) {
					t.Parallel()
					val, err := c.valueFn(v)
					if err != nil {
						t.Fatalf("%s(%q).Value() error = %v", c.name, v, err)
					}
					str, ok := val.(string)
					if !ok {
						t.Fatalf("%s(%q).Value() type = %T, want string", c.name, v, val)
					}
					if str != v {
						t.Errorf("%s(%q).Value() = %q", c.name, v, str)
					}
				})
			}
		})
	}
}

func TestEnums_Value_Invalid(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.valueFn("")
			if err == nil {
				t.Errorf("%s(\"\").Value() = nil error, want error", c.name)
			}
			_, err = c.valueFn("INVALID")
			if err == nil {
				t.Errorf("%s(\"INVALID\").Value() = nil error, want error", c.name)
			}
		})
	}
}

func TestEnums_Scan_String(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, v := range c.validValues {
				t.Run(v, func(t *testing.T) {
					t.Parallel()
					got, err := c.scanFn(v)
					if err != nil {
						t.Fatalf("Scan(%q) error = %v", v, err)
					}
					if got != v {
						t.Errorf("Scan(%q) = %q", v, got)
					}
				})
			}
		})
	}
}

func TestEnums_Scan_Bytes(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			v := c.validValues[0]
			got, err := c.scanFn([]byte(v))
			if err != nil {
				t.Fatalf("Scan([]byte(%q)) error = %v", v, err)
			}
			if got != v {
				t.Errorf("Scan([]byte(%q)) = %q", v, got)
			}
		})
	}
}

func TestEnums_Scan_Nil(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.scanFn(nil)
			if err == nil {
				t.Errorf("%s.Scan(nil) = nil error, want error", c.name)
			}
		})
	}
}

func TestEnums_Scan_UnsupportedType(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.scanFn(42)
			if err == nil {
				t.Errorf("%s.Scan(42) = nil error, want error", c.name)
			}
		})
	}
}

func TestEnums_Scan_InvalidValue(t *testing.T) {
	t.Parallel()
	for _, c := range allEnumContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.scanFn("NOPE")
			if err == nil {
				t.Errorf("%s.Scan(\"NOPE\") = nil error, want error", c.name)
			}
		})
	}
}

// --- JSON tests for representative enums ---

func TestContentStatus_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	cs := ContentStatusPublished
	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	if string(data) != `"published"` {
		t.Errorf("MarshalJSON = %s, want %q", data, "published")
	}

	var got ContentStatus
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if got != cs {
		t.Errorf("JSON round-trip: got %q, want %q", got, cs)
	}
}

func TestContentStatus_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()
	var cs ContentStatus
	if err := json.Unmarshal([]byte(`"bogus"`), &cs); err == nil {
		t.Error("UnmarshalJSON(\"bogus\") expected error")
	}
	if err := json.Unmarshal([]byte(`42`), &cs); err == nil {
		t.Error("UnmarshalJSON(42) expected error")
	}
}

func TestFieldType_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	ft := FieldTypeRichText
	data, err := json.Marshal(ft)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	var got FieldType
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if got != ft {
		t.Errorf("JSON round-trip: got %q, want %q", got, ft)
	}
}

func TestOperation_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	op := OpInsert
	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	var got Operation
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if got != op {
		t.Errorf("JSON round-trip: got %q, want %q", got, op)
	}
}

func TestBackupType_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	bt := BackupTypeIncremental
	data, err := json.Marshal(bt)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	var got BackupType
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if got != bt {
		t.Errorf("JSON round-trip: got %q, want %q", got, bt)
	}
}

// --- String() tests ---

func TestEnums_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "ContentStatus", got: ContentStatusDraft.String(), want: "draft"},
		{name: "FieldType", got: FieldTypeText.String(), want: "text"},
		{name: "RouteType", got: RouteTypeAPI.String(), want: "api"},
		{name: "Operation", got: OpDelete.String(), want: "DELETE"},
		{name: "Action", got: ActionPublish.String(), want: "publish"},
		{name: "ConflictPolicy", got: ConflictLWW.String(), want: "lww"},
		{name: "BackupType", got: BackupTypeFull.String(), want: "full"},
		{name: "BackupStatus", got: BackupStatusFailed.String(), want: "failed"},
		{name: "VerificationStatus", got: VerificationStatusVerified.String(), want: "verified"},
		{name: "BackupSetStatus", got: BackupSetStatusComplete.String(), want: "complete"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s.String() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
