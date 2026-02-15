package types

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
)

// --- helpers ---

// validULID returns a fixed, valid 26-char ULID string for deterministic tests.
func validULID() string {
	return "01HXYZ1234567890ABCDEFGHJ0"
}

// generateValidULID creates a fresh ULID via the production generator so we
// know it round-trips correctly.
func generateValidULID() string {
	return NewULID().String()
}

// --- NewULID ---

func TestNewULID_ReturnsValid26CharString(t *testing.T) {
	t.Parallel()
	id := NewULID()
	s := id.String()
	if len(s) != 26 {
		t.Fatalf("NewULID() length = %d, want 26", len(s))
	}
	if _, err := ulid.Parse(s); err != nil {
		t.Fatalf("NewULID() produced unparseable ULID %q: %v", s, err)
	}
}

func TestNewULID_Uniqueness(t *testing.T) {
	t.Parallel()
	seen := make(map[string]struct{}, 1000)
	for range 1000 {
		s := NewULID().String()
		if _, exists := seen[s]; exists {
			t.Fatalf("NewULID() produced duplicate: %s", s)
		}
		seen[s] = struct{}{}
	}
}

func TestNewULID_ThreadSafety(t *testing.T) {
	t.Parallel()
	const goroutines = 50
	const perGoroutine = 100
	results := make(chan string, goroutines*perGoroutine)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range perGoroutine {
				results <- NewULID().String()
			}
		}()
	}
	wg.Wait()
	close(results)

	seen := make(map[string]struct{}, goroutines*perGoroutine)
	for s := range results {
		if _, exists := seen[s]; exists {
			t.Fatalf("concurrent NewULID() produced duplicate: %s", s)
		}
		seen[s] = struct{}{}
	}
}

// --- validateULID (unexported, tested indirectly via ID types) ---

func TestValidateULID_Indirect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   DatatypeID
		wantErr bool
	}{
		{name: "valid generated", input: DatatypeID(generateValidULID()), wantErr: false},
		{name: "empty string", input: DatatypeID(""), wantErr: true},
		{name: "too short", input: DatatypeID("01HX"), wantErr: true},
		{name: "too long", input: DatatypeID("01HXYZ1234567890ABCDEFGHJ0X"), wantErr: true},
		// Note: oklog/ulid uses a permissive decoder that accepts any ASCII in parse,
		// so 26-char strings with "invalid" Crockford chars still parse successfully.
		// Validation only catches empty and wrong-length strings.
		{name: "all zeros 26 chars", input: DatatypeID("00000000000000000000000000"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DatatypeID(%q).Validate() error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// --- DatatypeID (comprehensive; used as the reference for all other ID types) ---

func TestDatatypeID_New(t *testing.T) {
	t.Parallel()
	id := NewDatatypeID()
	if id.IsZero() {
		t.Fatal("NewDatatypeID() returned zero value")
	}
	if err := id.Validate(); err != nil {
		t.Fatalf("NewDatatypeID().Validate() = %v", err)
	}
	if id.String() == "" {
		t.Fatal("NewDatatypeID().String() is empty")
	}
}

func TestDatatypeID_IsZero(t *testing.T) {
	t.Parallel()
	var zero DatatypeID
	if !zero.IsZero() {
		t.Error("zero DatatypeID.IsZero() = false, want true")
	}
	id := NewDatatypeID()
	if id.IsZero() {
		t.Error("NewDatatypeID().IsZero() = true, want false")
	}
}

func TestDatatypeID_ULID(t *testing.T) {
	t.Parallel()
	id := NewDatatypeID()
	u, err := id.ULID()
	if err != nil {
		t.Fatalf("ULID() error = %v", err)
	}
	if u.String() != id.String() {
		t.Errorf("ULID().String() = %q, want %q", u.String(), id.String())
	}
}

func TestDatatypeID_Time(t *testing.T) {
	t.Parallel()
	before := time.Now().Add(-time.Second)
	id := NewDatatypeID()
	after := time.Now().Add(time.Second)

	ts, err := id.Time()
	if err != nil {
		t.Fatalf("Time() error = %v", err)
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Time() = %v, want between %v and %v", ts, before, after)
	}
}

func TestDatatypeID_Time_InvalidID(t *testing.T) {
	t.Parallel()
	id := DatatypeID("not-a-ulid")
	_, err := id.Time()
	if err == nil {
		t.Error("Time() on invalid ID expected error, got nil")
	}
}

func TestDatatypeID_Value(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		id      DatatypeID
		wantNil bool
		wantErr bool
	}{
		{name: "valid", id: NewDatatypeID(), wantNil: false, wantErr: false},
		{name: "empty", id: DatatypeID(""), wantNil: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v, err := tt.id.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil && v != nil {
				t.Errorf("Value() = %v, want nil", v)
			}
			if !tt.wantNil && v == nil {
				t.Error("Value() = nil, want non-nil")
			}
		})
	}
}

func TestDatatypeID_Scan(t *testing.T) {
	t.Parallel()
	valid := generateValidULID()
	tests := []struct {
		name    string
		input   any
		wantErr bool
		wantID  DatatypeID
	}{
		{name: "scan string", input: valid, wantErr: false, wantID: DatatypeID(valid)},
		{name: "scan bytes", input: []byte(valid), wantErr: false, wantID: DatatypeID(valid)},
		{name: "scan nil", input: nil, wantErr: true},
		{name: "scan int", input: 42, wantErr: true},
		{name: "scan invalid string", input: "too-short", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var id DatatypeID
			err := id.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && id != tt.wantID {
				t.Errorf("Scan(%v) = %q, want %q", tt.input, id, tt.wantID)
			}
		})
	}
}

func TestDatatypeID_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	id := NewDatatypeID()
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}

	var got DatatypeID
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if got != id {
		t.Errorf("JSON round-trip: got %q, want %q", got, id)
	}
}

func TestDatatypeID_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{name: "invalid json", input: `{}`},
		{name: "too short", input: `"abc"`},
		{name: "number", input: `42`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var id DatatypeID
			if err := json.Unmarshal([]byte(tt.input), &id); err == nil {
				t.Errorf("UnmarshalJSON(%s) expected error, got nil (id=%q)", tt.input, id)
			}
		})
	}
}

func TestDatatypeID_Parse(t *testing.T) {
	t.Parallel()
	valid := generateValidULID()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: valid, wantErr: false},
		{name: "empty", input: "", wantErr: true},
		{name: "invalid", input: "not-a-ulid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, err := ParseDatatypeID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDatatypeID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && id.String() != tt.input {
				t.Errorf("ParseDatatypeID(%q) = %q", tt.input, id)
			}
		})
	}
}

// --- Conformance tests: every other ID type follows the same contract ---
// These ensure no copy-paste error broke any particular ID type.

// idContract defines the common operations every ID type must support.
type idContract struct {
	name       string
	newFn      func() string            // generates a new valid ID string
	validateFn func(string) error        // validates a string
	valueFn    func(string) (any, error) // driver.Value
}

func allIDContracts() []idContract {
	return []idContract{
		{
			name:       "DatatypeID",
			newFn:      func() string { return string(NewDatatypeID()) },
			validateFn: func(s string) error { return DatatypeID(s).Validate() },
			valueFn:    func(s string) (any, error) { return DatatypeID(s).Value() },
		},
		{
			name:       "UserID",
			newFn:      func() string { return string(NewUserID()) },
			validateFn: func(s string) error { return UserID(s).Validate() },
			valueFn:    func(s string) (any, error) { return UserID(s).Value() },
		},
		{
			name:       "RoleID",
			newFn:      func() string { return string(NewRoleID()) },
			validateFn: func(s string) error { return RoleID(s).Validate() },
			valueFn:    func(s string) (any, error) { return RoleID(s).Value() },
		},
		{
			name:       "PermissionID",
			newFn:      func() string { return string(NewPermissionID()) },
			validateFn: func(s string) error { return PermissionID(s).Validate() },
			valueFn:    func(s string) (any, error) { return PermissionID(s).Value() },
		},
		{
			name:       "FieldID",
			newFn:      func() string { return string(NewFieldID()) },
			validateFn: func(s string) error { return FieldID(s).Validate() },
			valueFn:    func(s string) (any, error) { return FieldID(s).Value() },
		},
		{
			name:       "ContentID",
			newFn:      func() string { return string(NewContentID()) },
			validateFn: func(s string) error { return ContentID(s).Validate() },
			valueFn:    func(s string) (any, error) { return ContentID(s).Value() },
		},
		{
			name:       "ContentFieldID",
			newFn:      func() string { return string(NewContentFieldID()) },
			validateFn: func(s string) error { return ContentFieldID(s).Validate() },
			valueFn:    func(s string) (any, error) { return ContentFieldID(s).Value() },
		},
		{
			name:       "MediaID",
			newFn:      func() string { return string(NewMediaID()) },
			validateFn: func(s string) error { return MediaID(s).Validate() },
			valueFn:    func(s string) (any, error) { return MediaID(s).Value() },
		},
		{
			name:       "MediaDimensionID",
			newFn:      func() string { return string(NewMediaDimensionID()) },
			validateFn: func(s string) error { return MediaDimensionID(s).Validate() },
			valueFn:    func(s string) (any, error) { return MediaDimensionID(s).Value() },
		},
		{
			name:       "SessionID",
			newFn:      func() string { return string(NewSessionID()) },
			validateFn: func(s string) error { return SessionID(s).Validate() },
			valueFn:    func(s string) (any, error) { return SessionID(s).Value() },
		},
		{
			name:       "TokenID",
			newFn:      func() string { return string(NewTokenID()) },
			validateFn: func(s string) error { return TokenID(s).Validate() },
			valueFn:    func(s string) (any, error) { return TokenID(s).Value() },
		},
		{
			name:       "RouteID",
			newFn:      func() string { return string(NewRouteID()) },
			validateFn: func(s string) error { return RouteID(s).Validate() },
			valueFn:    func(s string) (any, error) { return RouteID(s).Value() },
		},
		{
			name:       "AdminRouteID",
			newFn:      func() string { return string(NewAdminRouteID()) },
			validateFn: func(s string) error { return AdminRouteID(s).Validate() },
			valueFn:    func(s string) (any, error) { return AdminRouteID(s).Value() },
		},
		{
			name:       "TableID",
			newFn:      func() string { return string(NewTableID()) },
			validateFn: func(s string) error { return TableID(s).Validate() },
			valueFn:    func(s string) (any, error) { return TableID(s).Value() },
		},
		{
			name:       "UserOauthID",
			newFn:      func() string { return string(NewUserOauthID()) },
			validateFn: func(s string) error { return UserOauthID(s).Validate() },
			valueFn:    func(s string) (any, error) { return UserOauthID(s).Value() },
		},
		{
			name:       "UserSshKeyID",
			newFn:      func() string { return string(NewUserSshKeyID()) },
			validateFn: func(s string) error { return UserSshKeyID(s).Validate() },
			valueFn:    func(s string) (any, error) { return UserSshKeyID(s).Value() },
		},
		{
			name:       "AdminDatatypeID",
			newFn:      func() string { return string(NewAdminDatatypeID()) },
			validateFn: func(s string) error { return AdminDatatypeID(s).Validate() },
			valueFn:    func(s string) (any, error) { return AdminDatatypeID(s).Value() },
		},
		{
			name:       "AdminFieldID",
			newFn:      func() string { return string(NewAdminFieldID()) },
			validateFn: func(s string) error { return AdminFieldID(s).Validate() },
			valueFn:    func(s string) (any, error) { return AdminFieldID(s).Value() },
		},
		{
			name:       "AdminContentID",
			newFn:      func() string { return string(NewAdminContentID()) },
			validateFn: func(s string) error { return AdminContentID(s).Validate() },
			valueFn:    func(s string) (any, error) { return AdminContentID(s).Value() },
		},
		{
			name:       "AdminContentFieldID",
			newFn:      func() string { return string(NewAdminContentFieldID()) },
			validateFn: func(s string) error { return AdminContentFieldID(s).Validate() },
			valueFn:    func(s string) (any, error) { return AdminContentFieldID(s).Value() },
		},
		{
			name:       "DatatypeFieldID",
			newFn:      func() string { return string(NewDatatypeFieldID()) },
			validateFn: func(s string) error { return DatatypeFieldID(s).Validate() },
			valueFn:    func(s string) (any, error) { return DatatypeFieldID(s).Value() },
		},
		{
			name:       "AdminDatatypeFieldID",
			newFn:      func() string { return string(NewAdminDatatypeFieldID()) },
			validateFn: func(s string) error { return AdminDatatypeFieldID(s).Validate() },
			valueFn:    func(s string) (any, error) { return AdminDatatypeFieldID(s).Value() },
		},
		{
			name:       "EventID",
			newFn:      func() string { return string(NewEventID()) },
			validateFn: func(s string) error { return EventID(s).Validate() },
			valueFn:    func(s string) (any, error) { return EventID(s).Value() },
		},
		{
			name:       "NodeID",
			newFn:      func() string { return string(NewNodeID()) },
			validateFn: func(s string) error { return NodeID(s).Validate() },
			valueFn:    func(s string) (any, error) { return NodeID(s).Value() },
		},
		{
			name:       "BackupID",
			newFn:      func() string { return string(NewBackupID()) },
			validateFn: func(s string) error { return BackupID(s).Validate() },
			valueFn:    func(s string) (any, error) { return BackupID(s).Value() },
		},
		{
			name:       "VerificationID",
			newFn:      func() string { return string(NewVerificationID()) },
			validateFn: func(s string) error { return VerificationID(s).Validate() },
			valueFn:    func(s string) (any, error) { return VerificationID(s).Value() },
		},
		{
			name:       "BackupSetID",
			newFn:      func() string { return string(NewBackupSetID()) },
			validateFn: func(s string) error { return BackupSetID(s).Validate() },
			valueFn:    func(s string) (any, error) { return BackupSetID(s).Value() },
		},
		{
			name:       "ContentRelationID",
			newFn:      func() string { return string(NewContentRelationID()) },
			validateFn: func(s string) error { return ContentRelationID(s).Validate() },
			valueFn:    func(s string) (any, error) { return ContentRelationID(s).Value() },
		},
		{
			name:       "AdminContentRelationID",
			newFn:      func() string { return string(NewAdminContentRelationID()) },
			validateFn: func(s string) error { return AdminContentRelationID(s).Validate() },
			valueFn:    func(s string) (any, error) { return AdminContentRelationID(s).Value() },
		},
	}
}

func TestIDTypes_Conformance_NewReturnsValid(t *testing.T) {
	t.Parallel()
	for _, c := range allIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.newFn()
			if s == "" {
				t.Fatalf("New%s() returned empty string", c.name)
			}
			if len(s) != 26 {
				t.Fatalf("New%s() length = %d, want 26", c.name, len(s))
			}
			if err := c.validateFn(s); err != nil {
				t.Fatalf("New%s().Validate() = %v", c.name, err)
			}
		})
	}
}

func TestIDTypes_Conformance_EmptyIsInvalid(t *testing.T) {
	t.Parallel()
	for _, c := range allIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if err := c.validateFn(""); err == nil {
				t.Fatalf("%s(\"\").Validate() = nil, want error", c.name)
			}
		})
	}
}

func TestIDTypes_Conformance_EmptyValueReturnsError(t *testing.T) {
	t.Parallel()
	for _, c := range allIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			v, err := c.valueFn("")
			if err == nil {
				t.Fatalf("%s(\"\").Value() = (%v, nil), want error", c.name, v)
			}
		})
	}
}

func TestIDTypes_Conformance_ValidValueReturnsString(t *testing.T) {
	t.Parallel()
	for _, c := range allIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.newFn()
			v, err := c.valueFn(s)
			if err != nil {
				t.Fatalf("%s.Value() error = %v", c.name, err)
			}
			str, ok := v.(string)
			if !ok {
				t.Fatalf("%s.Value() type = %T, want string", c.name, v)
			}
			if str != s {
				t.Fatalf("%s.Value() = %q, want %q", c.name, str, s)
			}
		})
	}
}

// --- Parse functions for types that have them ---

func TestParseUserID(t *testing.T) {
	t.Parallel()
	valid := generateValidULID()
	id, err := ParseUserID(valid)
	if err != nil {
		t.Fatalf("ParseUserID(%q) error = %v", valid, err)
	}
	if string(id) != valid {
		t.Errorf("ParseUserID(%q) = %q", valid, id)
	}

	_, err = ParseUserID("")
	if err == nil {
		t.Error("ParseUserID(\"\") expected error")
	}
}

func TestParseContentRelationID(t *testing.T) {
	t.Parallel()
	valid := generateValidULID()
	id, err := ParseContentRelationID(valid)
	if err != nil {
		t.Fatalf("ParseContentRelationID(%q) error = %v", valid, err)
	}
	if string(id) != valid {
		t.Errorf("ParseContentRelationID(%q) = %q", valid, id)
	}

	_, err = ParseContentRelationID("nope")
	if err == nil {
		t.Error("ParseContentRelationID(\"nope\") expected error")
	}
}

func TestParseAdminContentRelationID(t *testing.T) {
	t.Parallel()
	valid := generateValidULID()
	id, err := ParseAdminContentRelationID(valid)
	if err != nil {
		t.Fatalf("ParseAdminContentRelationID(%q) error = %v", valid, err)
	}
	if string(id) != valid {
		t.Errorf("ParseAdminContentRelationID(%q) = %q", valid, id)
	}

	_, err = ParseAdminContentRelationID("")
	if err == nil {
		t.Error("ParseAdminContentRelationID(\"\") expected error")
	}
}

// --- ContentRelationID extra methods (ULID, Time) ---

func TestContentRelationID_ULID(t *testing.T) {
	t.Parallel()
	id := NewContentRelationID()
	u, err := id.ULID()
	if err != nil {
		t.Fatalf("ContentRelationID.ULID() error = %v", err)
	}
	if u.String() != id.String() {
		t.Errorf("ULID().String() = %q, want %q", u.String(), id.String())
	}
}

func TestContentRelationID_Time(t *testing.T) {
	t.Parallel()
	before := time.Now().Add(-time.Second)
	id := NewContentRelationID()
	after := time.Now().Add(time.Second)

	ts, err := id.Time()
	if err != nil {
		t.Fatalf("ContentRelationID.Time() error = %v", err)
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Time() = %v, want between %v and %v", ts, before, after)
	}
}

// --- AdminContentRelationID extra methods (ULID, Time) ---

func TestAdminContentRelationID_ULID(t *testing.T) {
	t.Parallel()
	id := NewAdminContentRelationID()
	u, err := id.ULID()
	if err != nil {
		t.Fatalf("AdminContentRelationID.ULID() error = %v", err)
	}
	if u.String() != id.String() {
		t.Errorf("ULID().String() = %q, want %q", u.String(), id.String())
	}
}

func TestAdminContentRelationID_Time(t *testing.T) {
	t.Parallel()
	before := time.Now().Add(-time.Second)
	id := NewAdminContentRelationID()
	after := time.Now().Add(time.Second)

	ts, err := id.Time()
	if err != nil {
		t.Fatalf("AdminContentRelationID.Time() error = %v", err)
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Time() = %v, want between %v and %v", ts, before, after)
	}
}

// --- UserID extra methods (ULID) ---

func TestUserID_ULID(t *testing.T) {
	t.Parallel()
	id := NewUserID()
	u, err := id.ULID()
	if err != nil {
		t.Fatalf("UserID.ULID() error = %v", err)
	}
	if u.String() != id.String() {
		t.Errorf("ULID().String() = %q, want %q", u.String(), id.String())
	}
}

// --- ContentID Scan with []byte (verify bytes path works for a non-DatatypeID type) ---

func TestContentID_Scan_Bytes(t *testing.T) {
	t.Parallel()
	valid := generateValidULID()
	var id ContentID
	if err := id.Scan([]byte(valid)); err != nil {
		t.Fatalf("ContentID.Scan([]byte) error = %v", err)
	}
	if string(id) != valid {
		t.Errorf("ContentID after Scan = %q, want %q", id, valid)
	}
}

// --- JSON round-trip for a non-DatatypeID type (catches copy-paste errors in JSON methods) ---

func TestRouteID_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	id := NewRouteID()
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("RouteID.MarshalJSON error = %v", err)
	}
	var got RouteID
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("RouteID.UnmarshalJSON error = %v", err)
	}
	if got != id {
		t.Errorf("RouteID JSON round-trip: got %q, want %q", got, id)
	}
}

func TestBackupSetID_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	id := NewBackupSetID()
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("BackupSetID.MarshalJSON error = %v", err)
	}
	var got BackupSetID
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("BackupSetID.UnmarshalJSON error = %v", err)
	}
	if got != id {
		t.Errorf("BackupSetID JSON round-trip: got %q, want %q", got, id)
	}
}
