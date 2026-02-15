package types

import (
	"encoding/json"
	"testing"
)

// nullableIDContract captures the common behavior of all nullable ID types.
type nullableIDContract struct {
	name       string
	newValidFn func() (string, any) // returns (raw ULID string, NullableXxxID{ID: xxx, Valid: true})
	scanNilFn  func() (bool, error) // scan nil -> returns (Valid, error)
	scanValFn  func(string) (string, bool, error)
	valueFn    func(string, bool) (any, error)
	marshalFn  func(string, bool) ([]byte, error)
	unmarshalFn func([]byte) (string, bool, error)
}

func allNullableIDContracts() []nullableIDContract {
	return []nullableIDContract{
		{
			name: "NullableDatatypeID",
			newValidFn: func() (string, any) {
				id := NewDatatypeID()
				return string(id), NullableDatatypeID{ID: id, Valid: true}
			},
			scanNilFn: func() (bool, error) {
				var n NullableDatatypeID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableDatatypeID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableDatatypeID{ID: DatatypeID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableDatatypeID{ID: DatatypeID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableDatatypeID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableUserID",
			newValidFn: func() (string, any) {
				id := NewUserID()
				return string(id), NullableUserID{ID: id, Valid: true}
			},
			scanNilFn: func() (bool, error) {
				var n NullableUserID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableUserID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableUserID{ID: UserID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableUserID{ID: UserID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableUserID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableContentID",
			newValidFn: func() (string, any) {
				id := NewContentID()
				return string(id), NullableContentID{ID: id, Valid: true}
			},
			scanNilFn: func() (bool, error) {
				var n NullableContentID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableContentID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableContentID{ID: ContentID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableContentID{ID: ContentID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableContentID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableFieldID",
			scanNilFn: func() (bool, error) {
				var n NullableFieldID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableFieldID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableFieldID{ID: FieldID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableFieldID{ID: FieldID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableFieldID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableMediaID",
			scanNilFn: func() (bool, error) {
				var n NullableMediaID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableMediaID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableMediaID{ID: MediaID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableMediaID{ID: MediaID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableMediaID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableRouteID",
			scanNilFn: func() (bool, error) {
				var n NullableRouteID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableRouteID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableRouteID{ID: RouteID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableRouteID{ID: RouteID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableRouteID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableAdminRouteID",
			scanNilFn: func() (bool, error) {
				var n NullableAdminRouteID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableAdminRouteID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableAdminRouteID{ID: AdminRouteID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableAdminRouteID{ID: AdminRouteID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableAdminRouteID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableAdminContentID",
			scanNilFn: func() (bool, error) {
				var n NullableAdminContentID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableAdminContentID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableAdminContentID{ID: AdminContentID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableAdminContentID{ID: AdminContentID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableAdminContentID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableNodeID",
			scanNilFn: func() (bool, error) {
				var n NullableNodeID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableNodeID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableNodeID{ID: NodeID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableNodeID{ID: NodeID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableNodeID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableRoleID",
			scanNilFn: func() (bool, error) {
				var n NullableRoleID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableRoleID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableRoleID{ID: RoleID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableRoleID{ID: RoleID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableRoleID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableAdminDatatypeID",
			scanNilFn: func() (bool, error) {
				var n NullableAdminDatatypeID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableAdminDatatypeID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableAdminDatatypeID{ID: AdminDatatypeID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableAdminDatatypeID{ID: AdminDatatypeID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableAdminDatatypeID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
		{
			name: "NullableAdminFieldID",
			scanNilFn: func() (bool, error) {
				var n NullableAdminFieldID
				err := n.Scan(nil)
				return n.Valid, err
			},
			scanValFn: func(s string) (string, bool, error) {
				var n NullableAdminFieldID
				err := n.Scan(s)
				return string(n.ID), n.Valid, err
			},
			valueFn: func(s string, valid bool) (any, error) {
				n := NullableAdminFieldID{ID: AdminFieldID(s), Valid: valid}
				return n.Value()
			},
			marshalFn: func(s string, valid bool) ([]byte, error) {
				n := NullableAdminFieldID{ID: AdminFieldID(s), Valid: valid}
				return json.Marshal(n)
			},
			unmarshalFn: func(data []byte) (string, bool, error) {
				var n NullableAdminFieldID
				err := json.Unmarshal(data, &n)
				return string(n.ID), n.Valid, err
			},
		},
	}
}

func TestNullableIDs_ScanNil_SetsInvalid(t *testing.T) {
	t.Parallel()
	for _, c := range allNullableIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			valid, err := c.scanNilFn()
			if err != nil {
				t.Fatalf("Scan(nil) error = %v", err)
			}
			if valid {
				t.Error("Scan(nil) -> Valid = true, want false")
			}
		})
	}
}

func TestNullableIDs_ScanValidString(t *testing.T) {
	t.Parallel()
	validULID := generateValidULID()
	for _, c := range allNullableIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			idStr, isValid, err := c.scanValFn(validULID)
			if err != nil {
				t.Fatalf("Scan(%q) error = %v", validULID, err)
			}
			if !isValid {
				t.Error("Scan(valid) -> Valid = false")
			}
			if idStr != validULID {
				t.Errorf("Scan(valid) ID = %q, want %q", idStr, validULID)
			}
		})
	}
}

func TestNullableIDs_ValueNull(t *testing.T) {
	t.Parallel()
	for _, c := range allNullableIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			v, err := c.valueFn("", false)
			if err != nil {
				t.Fatalf("Value(null) error = %v", err)
			}
			if v != nil {
				t.Errorf("Value(null) = %v, want nil", v)
			}
		})
	}
}

func TestNullableIDs_ValueValid(t *testing.T) {
	t.Parallel()
	validULID := generateValidULID()
	for _, c := range allNullableIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			v, err := c.valueFn(validULID, true)
			if err != nil {
				t.Fatalf("Value(valid) error = %v", err)
			}
			str, ok := v.(string)
			if !ok {
				t.Fatalf("Value() type = %T, want string", v)
			}
			if str != validULID {
				t.Errorf("Value() = %q, want %q", str, validULID)
			}
		})
	}
}

func TestNullableIDs_JSON_Null(t *testing.T) {
	t.Parallel()
	for _, c := range allNullableIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			data, err := c.marshalFn("", false)
			if err != nil {
				t.Fatalf("MarshalJSON(null) error = %v", err)
			}
			if string(data) != "null" {
				t.Errorf("MarshalJSON(null) = %s, want null", data)
			}

			idStr, valid, err := c.unmarshalFn([]byte("null"))
			if err != nil {
				t.Fatalf("UnmarshalJSON(null) error = %v", err)
			}
			if valid {
				t.Error("UnmarshalJSON(null) -> Valid = true")
			}
			if idStr != "" {
				t.Errorf("UnmarshalJSON(null) -> ID = %q, want empty", idStr)
			}
		})
	}
}

func TestNullableIDs_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	validULID := generateValidULID()
	for _, c := range allNullableIDContracts() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			data, err := c.marshalFn(validULID, true)
			if err != nil {
				t.Fatalf("MarshalJSON error = %v", err)
			}

			idStr, valid, err := c.unmarshalFn(data)
			if err != nil {
				t.Fatalf("UnmarshalJSON error = %v", err)
			}
			if !valid {
				t.Error("UnmarshalJSON -> Valid = false")
			}
			if idStr != validULID {
				t.Errorf("UnmarshalJSON -> ID = %q, want %q", idStr, validULID)
			}
		})
	}
}

// --- Validate and String/IsZero for representative types ---

func TestNullableDatatypeID_Validate(t *testing.T) {
	t.Parallel()
	// Null -> valid
	n := NullableDatatypeID{Valid: false}
	if err := n.Validate(); err != nil {
		t.Errorf("null.Validate() = %v", err)
	}

	// Valid with good ID -> valid
	n = NullableDatatypeID{ID: NewDatatypeID(), Valid: true}
	if err := n.Validate(); err != nil {
		t.Errorf("valid.Validate() = %v", err)
	}

	// Valid with bad ID -> error
	n = NullableDatatypeID{ID: DatatypeID("bad"), Valid: true}
	if err := n.Validate(); err == nil {
		t.Error("valid+bad.Validate() = nil, want error")
	}
}

func TestNullableUserID_StringAndIsZero(t *testing.T) {
	t.Parallel()
	// Null
	n := NullableUserID{Valid: false}
	if n.String() != "null" {
		t.Errorf("null.String() = %q", n.String())
	}
	if !n.IsZero() {
		t.Error("null.IsZero() = false")
	}

	// Valid
	id := NewUserID()
	n = NullableUserID{ID: id, Valid: true}
	if n.String() != id.String() {
		t.Errorf("valid.String() = %q, want %q", n.String(), id.String())
	}
	if n.IsZero() {
		t.Error("valid.IsZero() = true")
	}

	// Valid but empty ID
	n = NullableUserID{ID: "", Valid: true}
	if !n.IsZero() {
		t.Error("valid+empty.IsZero() = false")
	}
}
