package remote

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// Nullable string helpers
// ---------------------------------------------------------------------------

// nullStr converts a *string to sql.NullString.
func nullStr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// strPtr converts a sql.NullString to *string.
func strPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// ---------------------------------------------------------------------------
// db.NullString helpers (wraps sql.NullString)
// ---------------------------------------------------------------------------

// dbNullStr converts a *string to db.NullString.
func dbNullStr(s *string) db.NullString {
	if s == nil {
		return db.NullString{}
	}
	return db.NullString{NullString: sql.NullString{String: *s, Valid: true}}
}

// dbStrPtr converts a db.NullString to *string.
func dbStrPtr(ns db.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// ---------------------------------------------------------------------------
// Nullable int64 helpers
// ---------------------------------------------------------------------------

// nullInt64 converts a *int64 to types.NullableInt64.
func nullInt64(v *int64) types.NullableInt64 {
	if v == nil {
		return types.NullableInt64{}
	}
	return types.NullableInt64{Int64: *v, Valid: true}
}

// int64Ptr converts a types.NullableInt64 to *int64.
func int64Ptr(n types.NullableInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

// ---------------------------------------------------------------------------
// Nullable float64 helpers
// ---------------------------------------------------------------------------

// nullFloat64 converts a *float64 to types.NullableFloat64.
func nullFloat64(v *float64) types.NullableFloat64 {
	if v == nil {
		return types.NullableFloat64{}
	}
	return types.NullableFloat64{Float64: *v, Valid: true}
}

// float64Ptr converts a types.NullableFloat64 to *float64.
func float64Ptr(n types.NullableFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	return &n.Float64
}

// ---------------------------------------------------------------------------
// Nullable typed ID helpers
// ---------------------------------------------------------------------------

func nullContentID(id *modula.ContentID) types.NullableContentID {
	if id == nil {
		return types.NullableContentID{}
	}
	return types.NullableContentID{ID: types.ContentID(string(*id)), Valid: true}
}

func contentIDPtr(n types.NullableContentID) *modula.ContentID {
	if !n.Valid {
		return nil
	}
	id := modula.ContentID(string(n.ID))
	return &id
}

func nullRouteID(id *modula.RouteID) types.NullableRouteID {
	if id == nil {
		return types.NullableRouteID{}
	}
	return types.NullableRouteID{ID: types.RouteID(string(*id)), Valid: true}
}

func routeIDPtr(n types.NullableRouteID) *modula.RouteID {
	if !n.Valid {
		return nil
	}
	id := modula.RouteID(string(n.ID))
	return &id
}

func nullDatatypeID(id *modula.DatatypeID) types.NullableDatatypeID {
	if id == nil {
		return types.NullableDatatypeID{}
	}
	return types.NullableDatatypeID{ID: types.DatatypeID(string(*id)), Valid: true}
}

func datatypeIDPtr(n types.NullableDatatypeID) *modula.DatatypeID {
	if !n.Valid {
		return nil
	}
	id := modula.DatatypeID(string(n.ID))
	return &id
}

func nullUserID(id *modula.UserID) types.NullableUserID {
	if id == nil {
		return types.NullableUserID{}
	}
	return types.NullableUserID{ID: types.UserID(string(*id)), Valid: true}
}

func userIDPtr(n types.NullableUserID) *modula.UserID {
	if !n.Valid {
		return nil
	}
	id := modula.UserID(string(n.ID))
	return &id
}

func nullFieldID(id *modula.FieldID) types.NullableFieldID {
	if id == nil {
		return types.NullableFieldID{}
	}
	return types.NullableFieldID{ID: types.FieldID(string(*id)), Valid: true}
}

func fieldIDPtr(n types.NullableFieldID) *modula.FieldID {
	if !n.Valid {
		return nil
	}
	id := modula.FieldID(string(n.ID))
	return &id
}

func nullAdminContentID(id *modula.AdminContentID) types.NullableAdminContentID {
	if id == nil {
		return types.NullableAdminContentID{}
	}
	return types.NullableAdminContentID{ID: types.AdminContentID(string(*id)), Valid: true}
}

func adminContentIDPtr(n types.NullableAdminContentID) *modula.AdminContentID {
	if !n.Valid {
		return nil
	}
	id := modula.AdminContentID(string(n.ID))
	return &id
}

func nullAdminRouteID(id *modula.AdminRouteID) types.NullableAdminRouteID {
	if id == nil {
		return types.NullableAdminRouteID{}
	}
	return types.NullableAdminRouteID{ID: types.AdminRouteID(string(*id)), Valid: true}
}

func adminRouteIDPtr(n types.NullableAdminRouteID) *modula.AdminRouteID {
	if !n.Valid {
		return nil
	}
	id := modula.AdminRouteID(string(n.ID))
	return &id
}

func nullAdminDatatypeID(id *modula.AdminDatatypeID) types.NullableAdminDatatypeID {
	if id == nil {
		return types.NullableAdminDatatypeID{}
	}
	return types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(string(*id)), Valid: true}
}

func adminDatatypeIDPtr(n types.NullableAdminDatatypeID) *modula.AdminDatatypeID {
	if !n.Valid {
		return nil
	}
	id := modula.AdminDatatypeID(string(n.ID))
	return &id
}

func nullAdminFieldID(id *modula.AdminFieldID) types.NullableAdminFieldID {
	if id == nil {
		return types.NullableAdminFieldID{}
	}
	return types.NullableAdminFieldID{ID: types.AdminFieldID(string(*id)), Valid: true}
}

func adminFieldIDPtr(n types.NullableAdminFieldID) *modula.AdminFieldID {
	if !n.Valid {
		return nil
	}
	id := modula.AdminFieldID(string(n.ID))
	return &id
}

// nullNullableString converts a *string to types.NullableString.
func nullNullableString(s *string) types.NullableString {
	if s == nil {
		return types.NullableString{}
	}
	return types.NullableString{String: *s, Valid: true}
}

// nullableStringPtr converts a types.NullableString to *string.
func nullableStringPtr(n types.NullableString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

// ---------------------------------------------------------------------------
// Timestamp helpers
// ---------------------------------------------------------------------------

// sdkTimestampToDb converts a modula.Timestamp to types.Timestamp.
func sdkTimestampToDb(ts modula.Timestamp) types.Timestamp {
	if ts.IsZero() {
		return types.Timestamp{}
	}
	t, err := ts.Time()
	if err != nil {
		return types.Timestamp{}
	}
	return types.NewTimestamp(t)
}

// dbTimestampToSdk converts a types.Timestamp to modula.Timestamp.
func dbTimestampToSdk(ts types.Timestamp) modula.Timestamp {
	if !ts.Valid {
		return ""
	}
	return modula.Timestamp(ts.Time.UTC().Format(time.RFC3339))
}

// sdkTimestampPtrToDb converts a *modula.Timestamp to types.Timestamp.
func sdkTimestampPtrToDb(ts *modula.Timestamp) types.Timestamp {
	if ts == nil || ts.IsZero() {
		return types.Timestamp{}
	}
	t, err := ts.Time()
	if err != nil {
		return types.Timestamp{}
	}
	return types.NewTimestamp(t)
}

// dbTimestampToSdkPtr converts a types.Timestamp to *modula.Timestamp (nil if zero).
func dbTimestampToSdkPtr(ts types.Timestamp) *modula.Timestamp {
	if !ts.Valid {
		return nil
	}
	s := modula.Timestamp(ts.Time.UTC().Format(time.RFC3339))
	return &s
}

// ---------------------------------------------------------------------------
// Roles / NullableString <-> []string helpers
// ---------------------------------------------------------------------------

// rolesToNullableString converts []string roles to types.NullableString.
// nil or empty slice -> invalid (NULL), otherwise comma-joined.
func rolesToNullableString(roles []string) types.NullableString {
	if len(roles) == 0 {
		return types.NullableString{}
	}
	return types.NullableString{String: strings.Join(roles, ","), Valid: true}
}

// nullableStringToRoles converts types.NullableString to []string.
func nullableStringToRoles(ns types.NullableString) []string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	return strings.Split(ns.String, ",")
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

// jsonDataToRaw converts types.JSONData to json.RawMessage.
func jsonDataToRaw(jd types.JSONData) json.RawMessage {
	if !jd.Valid || jd.Data == nil {
		return nil
	}
	// JSONData.Data is any; attempt to marshal it to raw JSON.
	raw, err := json.Marshal(jd.Data)
	if err != nil {
		return nil
	}
	return raw
}

// rawToJSONData converts json.RawMessage to types.JSONData.
func rawToJSONData(raw json.RawMessage) types.JSONData {
	if raw == nil {
		return types.JSONData{}
	}
	return types.NewJSONData(raw)
}

// ---------------------------------------------------------------------------
// ContentID <-> NullableContentID via *string (for SDK *string sibling pointers)
// ---------------------------------------------------------------------------

// nullContentIDFromString converts a *string to types.NullableContentID.
func nullContentIDFromString(s *string) types.NullableContentID {
	if s == nil {
		return types.NullableContentID{}
	}
	return types.NullableContentID{ID: types.ContentID(*s), Valid: true}
}

// nullableContentIDToString converts types.NullableContentID to *string.
func nullableContentIDToString(n types.NullableContentID) *string {
	if !n.Valid {
		return nil
	}
	s := string(n.ID)
	return &s
}

// nullAdminContentIDFromString converts a *string to types.NullableAdminContentID.
func nullAdminContentIDFromString(s *string) types.NullableAdminContentID {
	if s == nil {
		return types.NullableAdminContentID{}
	}
	return types.NullableAdminContentID{ID: types.AdminContentID(*s), Valid: true}
}

// nullableAdminContentIDToString converts types.NullableAdminContentID to *string.
func nullableAdminContentIDToString(n types.NullableAdminContentID) *string {
	if !n.Valid {
		return nil
	}
	s := string(n.ID)
	return &s
}

// userIDToDb converts a modula.UserID to types.UserID for non-nullable fields.
func userIDToDb(id modula.UserID) types.UserID {
	return types.UserID(string(id))
}

// userIDFromDb converts a types.UserID to modula.UserID.
func userIDFromDb(id types.UserID) modula.UserID {
	return modula.UserID(string(id))
}

// userIDPtrToDb converts a *modula.UserID to types.UserID for fields where the
// SDK uses a pointer but the db uses a non-nullable UserID. If nil, returns zero value.
func userIDPtrToDb(id *modula.UserID) types.UserID {
	if id == nil {
		return ""
	}
	return types.UserID(string(*id))
}

// userIDToSdkPtr converts a non-nullable types.UserID to *modula.UserID.
// Returns nil for zero-valued IDs.
func userIDToSdkPtr(id types.UserID) *modula.UserID {
	if id == "" {
		return nil
	}
	uid := modula.UserID(string(id))
	return &uid
}

// nullableUserIDToDb converts a types.NullableUserID to types.UserID for
// bridging nullable->non-nullable. Returns zero value if not valid.
func nullableUserIDToDb(n types.NullableUserID) types.UserID {
	if !n.Valid {
		return ""
	}
	return n.ID
}
