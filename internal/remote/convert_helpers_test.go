package remote

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// nullStr / strPtr
// ---------------------------------------------------------------------------

func TestNullStr(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := nullStr(nil)
		if result.Valid {
			t.Error("expected invalid NullString for nil input")
		}
	})
	t.Run("non-nil produces valid", func(t *testing.T) {
		s := "hello"
		result := nullStr(&s)
		if !result.Valid {
			t.Fatal("expected valid NullString")
		}
		if result.String != "hello" {
			t.Errorf("String = %q, want %q", result.String, "hello")
		}
	})
	t.Run("empty string produces valid", func(t *testing.T) {
		s := ""
		result := nullStr(&s)
		if !result.Valid {
			t.Error("expected valid NullString for empty string")
		}
		if result.String != "" {
			t.Errorf("String = %q, want empty", result.String)
		}
	})
}

func TestStrPtr(t *testing.T) {
	t.Run("invalid produces nil", func(t *testing.T) {
		result := strPtr(sql.NullString{})
		if result != nil {
			t.Errorf("expected nil, got %q", *result)
		}
	})
	t.Run("valid produces pointer", func(t *testing.T) {
		result := strPtr(sql.NullString{String: "hello", Valid: true})
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != "hello" {
			t.Errorf("*result = %q, want %q", *result, "hello")
		}
	})
}

func TestNullStr_StrPtr_Roundtrip(t *testing.T) {
	s := "round-trip"
	result := strPtr(nullStr(&s))
	if result == nil || *result != s {
		t.Errorf("roundtrip failed: got %v, want %q", result, s)
	}

	resultNil := strPtr(nullStr(nil))
	if resultNil != nil {
		t.Errorf("nil roundtrip failed: got %v, want nil", resultNil)
	}
}

// ---------------------------------------------------------------------------
// dbNullStr / dbStrPtr
// ---------------------------------------------------------------------------

func TestDbNullStr(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := dbNullStr(nil)
		if result.Valid {
			t.Error("expected invalid db.NullString for nil input")
		}
	})
	t.Run("non-nil produces valid", func(t *testing.T) {
		s := "test"
		result := dbNullStr(&s)
		if !result.Valid {
			t.Fatal("expected valid db.NullString")
		}
		if result.String != "test" {
			t.Errorf("String = %q, want %q", result.String, "test")
		}
	})
}

func TestDbStrPtr(t *testing.T) {
	t.Run("invalid produces nil", func(t *testing.T) {
		result := dbStrPtr(db.NullString{})
		if result != nil {
			t.Errorf("expected nil, got %q", *result)
		}
	})
	t.Run("valid produces pointer", func(t *testing.T) {
		ns := db.NullString{NullString: sql.NullString{String: "test", Valid: true}}
		result := dbStrPtr(ns)
		if result == nil || *result != "test" {
			t.Errorf("expected %q, got %v", "test", result)
		}
	})
}

func TestDbNullStr_DbStrPtr_Roundtrip(t *testing.T) {
	s := "roundtrip"
	result := dbStrPtr(dbNullStr(&s))
	if result == nil || *result != s {
		t.Errorf("roundtrip failed: got %v, want %q", result, s)
	}
}

// ---------------------------------------------------------------------------
// nullInt64 / int64Ptr
// ---------------------------------------------------------------------------

func TestNullInt64(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := nullInt64(nil)
		if result.Valid {
			t.Error("expected invalid for nil input")
		}
	})
	t.Run("non-nil produces valid", func(t *testing.T) {
		v := int64(42)
		result := nullInt64(&v)
		if !result.Valid || result.Int64 != 42 {
			t.Errorf("got {%d, %v}, want {42, true}", result.Int64, result.Valid)
		}
	})
	t.Run("zero value preserved", func(t *testing.T) {
		v := int64(0)
		result := nullInt64(&v)
		if !result.Valid || result.Int64 != 0 {
			t.Errorf("zero value not preserved: got {%d, %v}", result.Int64, result.Valid)
		}
	})
}

func TestInt64Ptr(t *testing.T) {
	t.Run("invalid produces nil", func(t *testing.T) {
		result := int64Ptr(types.NullableInt64{})
		if result != nil {
			t.Errorf("expected nil, got %d", *result)
		}
	})
	t.Run("valid produces pointer", func(t *testing.T) {
		result := int64Ptr(types.NullableInt64{Int64: 99, Valid: true})
		if result == nil || *result != 99 {
			t.Errorf("expected 99, got %v", result)
		}
	})
}

func TestNullInt64_Int64Ptr_Roundtrip(t *testing.T) {
	v := int64(-7)
	result := int64Ptr(nullInt64(&v))
	if result == nil || *result != v {
		t.Errorf("roundtrip failed: got %v, want %d", result, v)
	}
}

// ---------------------------------------------------------------------------
// nullFloat64 / float64Ptr
// ---------------------------------------------------------------------------

func TestNullFloat64(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := nullFloat64(nil)
		if result.Valid {
			t.Error("expected invalid for nil input")
		}
	})
	t.Run("non-nil produces valid", func(t *testing.T) {
		v := 3.14
		result := nullFloat64(&v)
		if !result.Valid || result.Float64 != 3.14 {
			t.Errorf("got {%f, %v}, want {3.14, true}", result.Float64, result.Valid)
		}
	})
}

func TestFloat64Ptr(t *testing.T) {
	t.Run("invalid produces nil", func(t *testing.T) {
		result := float64Ptr(types.NullableFloat64{})
		if result != nil {
			t.Errorf("expected nil, got %f", *result)
		}
	})
	t.Run("valid produces pointer", func(t *testing.T) {
		result := float64Ptr(types.NullableFloat64{Float64: 2.718, Valid: true})
		if result == nil || *result != 2.718 {
			t.Errorf("expected 2.718, got %v", result)
		}
	})
}

// ---------------------------------------------------------------------------
// Typed ID helpers (one representative pair per ID type)
// ---------------------------------------------------------------------------

func TestNullContentID_ContentIDPtr_Roundtrip(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := nullContentID(nil)
		if result.Valid {
			t.Error("expected invalid for nil ContentID")
		}
	})
	t.Run("roundtrip", func(t *testing.T) {
		id := modula.ContentID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
		result := contentIDPtr(nullContentID(&id))
		if result == nil || *result != id {
			t.Errorf("roundtrip failed: got %v, want %v", result, id)
		}
	})
	t.Run("nil roundtrip", func(t *testing.T) {
		result := contentIDPtr(nullContentID(nil))
		if result != nil {
			t.Errorf("nil roundtrip failed: got %v", result)
		}
	})
}

func TestNullRouteID_RouteIDPtr_Roundtrip(t *testing.T) {
	id := modula.RouteID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := routeIDPtr(nullRouteID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if routeIDPtr(nullRouteID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

func TestNullDatatypeID_DatatypeIDPtr_Roundtrip(t *testing.T) {
	id := modula.DatatypeID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := datatypeIDPtr(nullDatatypeID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if datatypeIDPtr(nullDatatypeID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

func TestNullUserID_UserIDPtr_Roundtrip(t *testing.T) {
	id := modula.UserID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := userIDPtr(nullUserID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if userIDPtr(nullUserID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

func TestNullFieldID_FieldIDPtr_Roundtrip(t *testing.T) {
	id := modula.FieldID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := fieldIDPtr(nullFieldID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if fieldIDPtr(nullFieldID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

func TestNullAdminContentID_Roundtrip(t *testing.T) {
	id := modula.AdminContentID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := adminContentIDPtr(nullAdminContentID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if adminContentIDPtr(nullAdminContentID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

func TestNullAdminRouteID_Roundtrip(t *testing.T) {
	id := modula.AdminRouteID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := adminRouteIDPtr(nullAdminRouteID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if adminRouteIDPtr(nullAdminRouteID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

func TestNullAdminDatatypeID_Roundtrip(t *testing.T) {
	id := modula.AdminDatatypeID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := adminDatatypeIDPtr(nullAdminDatatypeID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if adminDatatypeIDPtr(nullAdminDatatypeID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

func TestNullAdminFieldID_Roundtrip(t *testing.T) {
	id := modula.AdminFieldID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := adminFieldIDPtr(nullAdminFieldID(&id))
	if result == nil || *result != id {
		t.Errorf("roundtrip failed: got %v, want %v", result, id)
	}
	if adminFieldIDPtr(nullAdminFieldID(nil)) != nil {
		t.Error("nil roundtrip should produce nil")
	}
}

// ---------------------------------------------------------------------------
// NullableString helpers
// ---------------------------------------------------------------------------

func TestNullNullableString_NullableStringPtr_Roundtrip(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := nullNullableString(nil)
		if result.Valid {
			t.Error("expected invalid for nil")
		}
	})
	t.Run("roundtrip", func(t *testing.T) {
		s := "test"
		result := nullableStringPtr(nullNullableString(&s))
		if result == nil || *result != s {
			t.Errorf("roundtrip failed: got %v, want %q", result, s)
		}
	})
	t.Run("nil roundtrip", func(t *testing.T) {
		result := nullableStringPtr(nullNullableString(nil))
		if result != nil {
			t.Errorf("nil roundtrip should produce nil, got %v", result)
		}
	})
}

// ---------------------------------------------------------------------------
// Timestamp helpers
// ---------------------------------------------------------------------------

func TestSdkTimestampToDb(t *testing.T) {
	t.Run("zero timestamp produces invalid", func(t *testing.T) {
		result := sdkTimestampToDb("")
		if result.Valid {
			t.Error("expected invalid for zero timestamp")
		}
	})
	t.Run("valid timestamp converts", func(t *testing.T) {
		ts := modula.Timestamp("2024-01-15T10:30:00Z")
		result := sdkTimestampToDb(ts)
		if !result.Valid {
			t.Fatal("expected valid timestamp")
		}
		expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		if !result.Time.Equal(expected) {
			t.Errorf("time = %v, want %v", result.Time, expected)
		}
	})
	t.Run("invalid format produces invalid", func(t *testing.T) {
		ts := modula.Timestamp("not-a-timestamp")
		result := sdkTimestampToDb(ts)
		if result.Valid {
			t.Error("expected invalid for unparseable timestamp")
		}
	})
}

func TestDbTimestampToSdk(t *testing.T) {
	t.Run("invalid produces empty", func(t *testing.T) {
		result := dbTimestampToSdk(types.Timestamp{})
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})
	t.Run("valid converts to RFC3339", func(t *testing.T) {
		ts := types.NewTimestamp(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
		result := dbTimestampToSdk(ts)
		want := modula.Timestamp("2024-06-01T12:00:00Z")
		if result != want {
			t.Errorf("result = %q, want %q", result, want)
		}
	})
}

func TestSdkTimestamp_Roundtrip(t *testing.T) {
	original := modula.Timestamp("2024-03-20T15:45:00Z")
	dbTs := sdkTimestampToDb(original)
	back := dbTimestampToSdk(dbTs)
	if back != original {
		t.Errorf("roundtrip failed: got %q, want %q", back, original)
	}
}

func TestSdkTimestampPtrToDb(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := sdkTimestampPtrToDb(nil)
		if result.Valid {
			t.Error("expected invalid for nil")
		}
	})
	t.Run("zero value produces invalid", func(t *testing.T) {
		ts := modula.Timestamp("")
		result := sdkTimestampPtrToDb(&ts)
		if result.Valid {
			t.Error("expected invalid for zero timestamp pointer")
		}
	})
	t.Run("valid pointer converts", func(t *testing.T) {
		ts := modula.Timestamp("2024-01-01T00:00:00Z")
		result := sdkTimestampPtrToDb(&ts)
		if !result.Valid {
			t.Fatal("expected valid timestamp")
		}
	})
}

func TestDbTimestampToSdkPtr(t *testing.T) {
	t.Run("invalid produces nil", func(t *testing.T) {
		result := dbTimestampToSdkPtr(types.Timestamp{})
		if result != nil {
			t.Errorf("expected nil, got %q", *result)
		}
	})
	t.Run("valid produces pointer", func(t *testing.T) {
		ts := types.NewTimestamp(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		result := dbTimestampToSdkPtr(ts)
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		want := modula.Timestamp("2024-01-01T00:00:00Z")
		if *result != want {
			t.Errorf("result = %q, want %q", *result, want)
		}
	})
}

// ---------------------------------------------------------------------------
// Roles helpers
// ---------------------------------------------------------------------------

func TestRolesToNullableString(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		valid bool
		want  string
	}{
		{name: "nil slice", roles: nil, valid: false},
		{name: "empty slice", roles: []string{}, valid: false},
		{name: "single role", roles: []string{"admin"}, valid: true, want: "admin"},
		{name: "multiple roles", roles: []string{"admin", "editor", "viewer"}, valid: true, want: "admin,editor,viewer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rolesToNullableString(tt.roles)
			if result.Valid != tt.valid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.valid)
			}
			if tt.valid && result.String != tt.want {
				t.Errorf("String = %q, want %q", result.String, tt.want)
			}
		})
	}
}

func TestNullableStringToRoles(t *testing.T) {
	tests := []struct {
		name string
		ns   types.NullableString
		want []string
	}{
		{name: "invalid produces nil", ns: types.NullableString{}, want: nil},
		{name: "empty string produces nil", ns: types.NullableString{String: "", Valid: true}, want: nil},
		{name: "single role", ns: types.NullableString{String: "admin", Valid: true}, want: []string{"admin"}},
		{name: "multiple roles", ns: types.NullableString{String: "admin,editor", Valid: true}, want: []string{"admin", "editor"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullableStringToRoles(tt.ns)
			if len(result) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(result), len(tt.want))
			}
			for i := range result {
				if result[i] != tt.want[i] {
					t.Errorf("result[%d] = %q, want %q", i, result[i], tt.want[i])
				}
			}
		})
	}
}

func TestRoles_Roundtrip(t *testing.T) {
	original := []string{"admin", "editor", "viewer"}
	result := nullableStringToRoles(rolesToNullableString(original))
	if len(result) != len(original) {
		t.Fatalf("roundtrip length = %d, want %d", len(result), len(original))
	}
	for i := range original {
		if result[i] != original[i] {
			t.Errorf("roundtrip[%d] = %q, want %q", i, result[i], original[i])
		}
	}
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

func TestJsonDataToRaw(t *testing.T) {
	t.Run("invalid produces nil", func(t *testing.T) {
		result := jsonDataToRaw(types.JSONData{})
		if result != nil {
			t.Errorf("expected nil, got %s", result)
		}
	})
	t.Run("nil data produces nil", func(t *testing.T) {
		result := jsonDataToRaw(types.JSONData{Data: nil, Valid: true})
		if result != nil {
			t.Errorf("expected nil for nil data, got %s", result)
		}
	})
	t.Run("valid data marshals", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		result := jsonDataToRaw(types.NewJSONData(data))
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		var decoded map[string]string
		if err := json.Unmarshal(result, &decoded); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if decoded["key"] != "value" {
			t.Errorf("decoded[key] = %q, want %q", decoded["key"], "value")
		}
	})
	t.Run("raw JSON bytes preserved", func(t *testing.T) {
		raw := json.RawMessage(`{"nested":true}`)
		result := jsonDataToRaw(types.NewJSONData(raw))
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	})
}

func TestRawToJSONData(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := rawToJSONData(nil)
		if result.Valid {
			t.Error("expected invalid for nil raw")
		}
	})
	t.Run("valid raw produces valid", func(t *testing.T) {
		raw := json.RawMessage(`{"hello":"world"}`)
		result := rawToJSONData(raw)
		if !result.Valid {
			t.Error("expected valid JSONData")
		}
		if result.Data == nil {
			t.Error("expected non-nil Data")
		}
	})
}

// ---------------------------------------------------------------------------
// ContentID string helpers
// ---------------------------------------------------------------------------

func TestNullContentIDFromString(t *testing.T) {
	t.Run("nil produces invalid", func(t *testing.T) {
		result := nullContentIDFromString(nil)
		if result.Valid {
			t.Error("expected invalid for nil")
		}
	})
	t.Run("non-nil produces valid", func(t *testing.T) {
		s := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
		result := nullContentIDFromString(&s)
		if !result.Valid {
			t.Fatal("expected valid")
		}
		if string(result.ID) != s {
			t.Errorf("ID = %q, want %q", result.ID, s)
		}
	})
}

func TestNullableContentIDToString(t *testing.T) {
	t.Run("invalid produces nil", func(t *testing.T) {
		result := nullableContentIDToString(types.NullableContentID{})
		if result != nil {
			t.Errorf("expected nil, got %q", *result)
		}
	})
	t.Run("valid produces string pointer", func(t *testing.T) {
		n := types.NullableContentID{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Valid: true}
		result := nullableContentIDToString(n)
		if result == nil || *result != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
			t.Errorf("expected %q, got %v", "01ARZ3NDEKTSV4RRFFQ69G5FAV", result)
		}
	})
}

func TestNullAdminContentIDFromString(t *testing.T) {
	s := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	result := nullAdminContentIDFromString(&s)
	if !result.Valid || string(result.ID) != s {
		t.Errorf("got {%v, %v}, want {%s, true}", result.ID, result.Valid, s)
	}
	if nullAdminContentIDFromString(nil).Valid {
		t.Error("nil should produce invalid")
	}
}

func TestNullableAdminContentIDToString(t *testing.T) {
	n := types.NullableAdminContentID{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Valid: true}
	result := nullableAdminContentIDToString(n)
	if result == nil || *result != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
		t.Errorf("expected %q, got %v", "01ARZ3NDEKTSV4RRFFQ69G5FAV", result)
	}
	if nullableAdminContentIDToString(types.NullableAdminContentID{}) != nil {
		t.Error("invalid should produce nil")
	}
}

// ---------------------------------------------------------------------------
// UserID helpers
// ---------------------------------------------------------------------------

func TestUserIDToDb(t *testing.T) {
	id := modula.UserID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := userIDToDb(id)
	if string(result) != string(id) {
		t.Errorf("got %q, want %q", result, id)
	}
}

func TestUserIDFromDb(t *testing.T) {
	id := types.UserID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	result := userIDFromDb(id)
	if string(result) != string(id) {
		t.Errorf("got %q, want %q", result, id)
	}
}

func TestUserIDPtrToDb(t *testing.T) {
	t.Run("nil produces zero value", func(t *testing.T) {
		result := userIDPtrToDb(nil)
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})
	t.Run("non-nil converts", func(t *testing.T) {
		id := modula.UserID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
		result := userIDPtrToDb(&id)
		if string(result) != string(id) {
			t.Errorf("got %q, want %q", result, id)
		}
	})
}

func TestUserIDToSdkPtr(t *testing.T) {
	t.Run("empty produces nil", func(t *testing.T) {
		result := userIDToSdkPtr("")
		if result != nil {
			t.Errorf("expected nil, got %q", *result)
		}
	})
	t.Run("non-empty produces pointer", func(t *testing.T) {
		id := types.UserID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
		result := userIDToSdkPtr(id)
		if result == nil || string(*result) != string(id) {
			t.Errorf("got %v, want %q", result, id)
		}
	})
}

func TestNullableUserIDToDb(t *testing.T) {
	t.Run("invalid produces zero value", func(t *testing.T) {
		result := nullableUserIDToDb(types.NullableUserID{})
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})
	t.Run("valid produces ID", func(t *testing.T) {
		n := types.NullableUserID{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Valid: true}
		result := nullableUserIDToDb(n)
		if string(result) != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
			t.Errorf("got %q, want %q", result, "01ARZ3NDEKTSV4RRFFQ69G5FAV")
		}
	})
}
