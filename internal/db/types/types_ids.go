// Package types provides ULID-based identity types and custom type wrappers for database operations.
// All ID types are thread-safe, database-compatible (via driver.Valuer and sql.Scanner), and JSON-serializable.
// Each ID type validates 26-character ULID format and embeds a sortable timestamp component.
package types

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// Thread-safe entropy source for ULID generation
var (
	entropyMu sync.Mutex
	entropy   = ulid.Monotonic(rand.Reader, 0)
)

// NewULID generates a new ULID (thread-safe)
func NewULID() ulid.ULID {
	entropyMu.Lock()
	defer entropyMu.Unlock()
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
}

// validateULID checks if a string is a valid 26-char ULID
func validateULID(s string, typeName string) error {
	if s == "" {
		return fmt.Errorf("%s: cannot be empty", typeName)
	}
	if len(s) != 26 {
		return fmt.Errorf("%s: invalid length %d (expected 26)", typeName, len(s))
	}
	_, err := ulid.Parse(s)
	if err != nil {
		return fmt.Errorf("%s: invalid ULID format %q: %w", typeName, s, err)
	}
	return nil
}

// DatatypeID uniquely identifies a datatype.
type DatatypeID string

// NewDatatypeID generates a new ULID-based DatatypeID.
func NewDatatypeID() DatatypeID              { return DatatypeID(NewULID().String()) }

// String returns the string representation of the DatatypeID.
func (id DatatypeID) String() string         { return string(id) }

// IsZero returns true if the DatatypeID is empty.
func (id DatatypeID) IsZero() bool           { return id == "" }

// Validate checks if the DatatypeID is a valid ULID.
func (id DatatypeID) Validate() error        { return validateULID(string(id), "DatatypeID") }

// ULID parses the DatatypeID as a ulid.ULID.
func (id DatatypeID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

// Time extracts the timestamp embedded in the DatatypeID.
func (id DatatypeID) Time() (time.Time, error) {
	u, err := id.ULID()
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(u.Time()), nil
}

// Value implements driver.Valuer for database serialization.
func (id DatatypeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("DatatypeID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *DatatypeID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("DatatypeID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = DatatypeID(v)
	case []byte:
		*id = DatatypeID(string(v))
	default:
		return fmt.Errorf("DatatypeID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id DatatypeID) MarshalJSON() ([]byte, error)  { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *DatatypeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("DatatypeID: %w", err)
	}
	*id = DatatypeID(s)
	return id.Validate()
}

// ParseDatatypeID parses and validates a string as a DatatypeID.
func ParseDatatypeID(s string) (DatatypeID, error) {
	id := DatatypeID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// UserID uniquely identifies a user.
type UserID string

// NewUserID generates a new ULID-based UserID.
func NewUserID() UserID              { return UserID(NewULID().String()) }

// String returns the string representation of the UserID.
func (id UserID) String() string     { return string(id) }

// IsZero returns true if the UserID is empty.
func (id UserID) IsZero() bool       { return id == "" }

// Validate checks if the UserID is a valid ULID.
func (id UserID) Validate() error    { return validateULID(string(id), "UserID") }

// ULID parses the UserID as a ulid.ULID.
func (id UserID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

// Value implements driver.Valuer for database serialization.
func (id UserID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("UserID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *UserID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("UserID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = UserID(v)
	case []byte:
		*id = UserID(string(v))
	default:
		return fmt.Errorf("UserID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id UserID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *UserID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("UserID: %w", err)
	}
	*id = UserID(s)
	return id.Validate()
}

// ParseUserID parses and validates a string as a UserID.
func ParseUserID(s string) (UserID, error) {
	id := UserID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// RoleID uniquely identifies a role.
type RoleID string

// NewRoleID generates a new ULID-based RoleID.
func NewRoleID() RoleID              { return RoleID(NewULID().String()) }

// String returns the string representation of the RoleID.
func (id RoleID) String() string     { return string(id) }

// IsZero returns true if the RoleID is empty.
func (id RoleID) IsZero() bool       { return id == "" }

// Validate checks if the RoleID is a valid ULID.
func (id RoleID) Validate() error    { return validateULID(string(id), "RoleID") }

// Value implements driver.Valuer for database serialization.
func (id RoleID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("RoleID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *RoleID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("RoleID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = RoleID(v)
	case []byte:
		*id = RoleID(string(v))
	default:
		return fmt.Errorf("RoleID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id RoleID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *RoleID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("RoleID: %w", err)
	}
	*id = RoleID(s)
	return id.Validate()
}

// PermissionID uniquely identifies a permission.
type PermissionID string

// NewPermissionID generates a new ULID-based PermissionID.
func NewPermissionID() PermissionID      { return PermissionID(NewULID().String()) }

// String returns the string representation of the PermissionID.
func (id PermissionID) String() string   { return string(id) }

// IsZero returns true if the PermissionID is empty.
func (id PermissionID) IsZero() bool     { return id == "" }

// Validate checks if the PermissionID is a valid ULID.
func (id PermissionID) Validate() error  { return validateULID(string(id), "PermissionID") }

// Value implements driver.Valuer for database serialization.
func (id PermissionID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("PermissionID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *PermissionID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("PermissionID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = PermissionID(v)
	case []byte:
		*id = PermissionID(string(v))
	default:
		return fmt.Errorf("PermissionID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id PermissionID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *PermissionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("PermissionID: %w", err)
	}
	*id = PermissionID(s)
	return id.Validate()
}

// FieldID uniquely identifies a field.
type FieldID string

// NewFieldID generates a new ULID-based FieldID.
func NewFieldID() FieldID            { return FieldID(NewULID().String()) }

// String returns the string representation of the FieldID.
func (id FieldID) String() string    { return string(id) }

// IsZero returns true if the FieldID is empty.
func (id FieldID) IsZero() bool      { return id == "" }

// Validate checks if the FieldID is a valid ULID.
func (id FieldID) Validate() error   { return validateULID(string(id), "FieldID") }

// Value implements driver.Valuer for database serialization.
func (id FieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("FieldID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *FieldID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("FieldID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = FieldID(v)
	case []byte:
		*id = FieldID(string(v))
	default:
		return fmt.Errorf("FieldID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id FieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *FieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("FieldID: %w", err)
	}
	*id = FieldID(s)
	return id.Validate()
}

// ContentID uniquely identifies content data.
type ContentID string

// NewContentID generates a new ULID-based ContentID.
func NewContentID() ContentID        { return ContentID(NewULID().String()) }

// String returns the string representation of the ContentID.
func (id ContentID) String() string  { return string(id) }

// IsZero returns true if the ContentID is empty.
func (id ContentID) IsZero() bool    { return id == "" }

// Validate checks if the ContentID is a valid ULID.
func (id ContentID) Validate() error { return validateULID(string(id), "ContentID") }

// Value implements driver.Valuer for database serialization.
func (id ContentID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("ContentID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *ContentID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("ContentID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = ContentID(v)
	case []byte:
		*id = ContentID(string(v))
	default:
		return fmt.Errorf("ContentID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id ContentID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *ContentID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ContentID: %w", err)
	}
	*id = ContentID(s)
	return id.Validate()
}

// ContentFieldID uniquely identifies a content field.
type ContentFieldID string

// NewContentFieldID generates a new ULID-based ContentFieldID.
func NewContentFieldID() ContentFieldID    { return ContentFieldID(NewULID().String()) }

// String returns the string representation of the ContentFieldID.
func (id ContentFieldID) String() string   { return string(id) }

// IsZero returns true if the ContentFieldID is empty.
func (id ContentFieldID) IsZero() bool     { return id == "" }

// Validate checks if the ContentFieldID is a valid ULID.
func (id ContentFieldID) Validate() error  { return validateULID(string(id), "ContentFieldID") }

// Value implements driver.Valuer for database serialization.
func (id ContentFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("ContentFieldID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *ContentFieldID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("ContentFieldID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = ContentFieldID(v)
	case []byte:
		*id = ContentFieldID(string(v))
	default:
		return fmt.Errorf("ContentFieldID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id ContentFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *ContentFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ContentFieldID: %w", err)
	}
	*id = ContentFieldID(s)
	return id.Validate()
}

// MediaID uniquely identifies media.
type MediaID string

// NewMediaID generates a new ULID-based MediaID.
func NewMediaID() MediaID            { return MediaID(NewULID().String()) }

// String returns the string representation of the MediaID.
func (id MediaID) String() string    { return string(id) }

// IsZero returns true if the MediaID is empty.
func (id MediaID) IsZero() bool      { return id == "" }

// Validate checks if the MediaID is a valid ULID.
func (id MediaID) Validate() error   { return validateULID(string(id), "MediaID") }

// Value implements driver.Valuer for database serialization.
func (id MediaID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("MediaID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *MediaID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("MediaID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = MediaID(v)
	case []byte:
		*id = MediaID(string(v))
	default:
		return fmt.Errorf("MediaID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id MediaID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *MediaID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("MediaID: %w", err)
	}
	*id = MediaID(s)
	return id.Validate()
}

// MediaDimensionID uniquely identifies a media dimension.
type MediaDimensionID string

// NewMediaDimensionID generates a new ULID-based MediaDimensionID.
func NewMediaDimensionID() MediaDimensionID    { return MediaDimensionID(NewULID().String()) }

// String returns the string representation of the MediaDimensionID.
func (id MediaDimensionID) String() string     { return string(id) }

// IsZero returns true if the MediaDimensionID is empty.
func (id MediaDimensionID) IsZero() bool       { return id == "" }

// Validate checks if the MediaDimensionID is a valid ULID.
func (id MediaDimensionID) Validate() error    { return validateULID(string(id), "MediaDimensionID") }

// Value implements driver.Valuer for database serialization.
func (id MediaDimensionID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("MediaDimensionID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *MediaDimensionID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("MediaDimensionID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = MediaDimensionID(v)
	case []byte:
		*id = MediaDimensionID(string(v))
	default:
		return fmt.Errorf("MediaDimensionID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id MediaDimensionID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *MediaDimensionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("MediaDimensionID: %w", err)
	}
	*id = MediaDimensionID(s)
	return id.Validate()
}

// SessionID uniquely identifies a session.
type SessionID string

// NewSessionID generates a new ULID-based SessionID.
func NewSessionID() SessionID        { return SessionID(NewULID().String()) }

// String returns the string representation of the SessionID.
func (id SessionID) String() string  { return string(id) }

// IsZero returns true if the SessionID is empty.
func (id SessionID) IsZero() bool    { return id == "" }

// Validate checks if the SessionID is a valid ULID.
func (id SessionID) Validate() error { return validateULID(string(id), "SessionID") }

// Value implements driver.Valuer for database serialization.
func (id SessionID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("SessionID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *SessionID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("SessionID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = SessionID(v)
	case []byte:
		*id = SessionID(string(v))
	default:
		return fmt.Errorf("SessionID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id SessionID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *SessionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("SessionID: %w", err)
	}
	*id = SessionID(s)
	return id.Validate()
}

// TokenID uniquely identifies a token.
type TokenID string

// NewTokenID generates a new ULID-based TokenID.
func NewTokenID() TokenID            { return TokenID(NewULID().String()) }

// String returns the string representation of the TokenID.
func (id TokenID) String() string    { return string(id) }

// IsZero returns true if the TokenID is empty.
func (id TokenID) IsZero() bool      { return id == "" }

// Validate checks if the TokenID is a valid ULID.
func (id TokenID) Validate() error   { return validateULID(string(id), "TokenID") }

// Value implements driver.Valuer for database serialization.
func (id TokenID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("TokenID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *TokenID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("TokenID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = TokenID(v)
	case []byte:
		*id = TokenID(string(v))
	default:
		return fmt.Errorf("TokenID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id TokenID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *TokenID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TokenID: %w", err)
	}
	*id = TokenID(s)
	return id.Validate()
}

// RouteID uniquely identifies a route.
type RouteID string

// NewRouteID generates a new ULID-based RouteID.
func NewRouteID() RouteID            { return RouteID(NewULID().String()) }

// String returns the string representation of the RouteID.
func (id RouteID) String() string    { return string(id) }

// IsZero returns true if the RouteID is empty.
func (id RouteID) IsZero() bool      { return id == "" }

// Validate checks if the RouteID is a valid ULID.
func (id RouteID) Validate() error   { return validateULID(string(id), "RouteID") }

// Value implements driver.Valuer for database serialization.
func (id RouteID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("RouteID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *RouteID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("RouteID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = RouteID(v)
	case []byte:
		*id = RouteID(string(v))
	default:
		return fmt.Errorf("RouteID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id RouteID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *RouteID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("RouteID: %w", err)
	}
	*id = RouteID(s)
	return id.Validate()
}

// AdminRouteID uniquely identifies an admin route.
type AdminRouteID string

// NewAdminRouteID generates a new ULID-based AdminRouteID.
func NewAdminRouteID() AdminRouteID      { return AdminRouteID(NewULID().String()) }

// String returns the string representation of the AdminRouteID.
func (id AdminRouteID) String() string   { return string(id) }

// IsZero returns true if the AdminRouteID is empty.
func (id AdminRouteID) IsZero() bool     { return id == "" }

// Validate checks if the AdminRouteID is a valid ULID.
func (id AdminRouteID) Validate() error  { return validateULID(string(id), "AdminRouteID") }

// Value implements driver.Valuer for database serialization.
func (id AdminRouteID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminRouteID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminRouteID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminRouteID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminRouteID(v)
	case []byte:
		*id = AdminRouteID(string(v))
	default:
		return fmt.Errorf("AdminRouteID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminRouteID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminRouteID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminRouteID: %w", err)
	}
	*id = AdminRouteID(s)
	return id.Validate()
}

// TableID uniquely identifies a table.
type TableID string

// NewTableID generates a new ULID-based TableID.
func NewTableID() TableID            { return TableID(NewULID().String()) }

// String returns the string representation of the TableID.
func (id TableID) String() string    { return string(id) }

// IsZero returns true if the TableID is empty.
func (id TableID) IsZero() bool      { return id == "" }

// Validate checks if the TableID is a valid ULID.
func (id TableID) Validate() error   { return validateULID(string(id), "TableID") }

// Value implements driver.Valuer for database serialization.
func (id TableID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("TableID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *TableID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("TableID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = TableID(v)
	case []byte:
		*id = TableID(string(v))
	default:
		return fmt.Errorf("TableID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id TableID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *TableID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TableID: %w", err)
	}
	*id = TableID(s)
	return id.Validate()
}

// UserOauthID uniquely identifies a user OAuth record.
type UserOauthID string

// NewUserOauthID generates a new ULID-based UserOauthID.
func NewUserOauthID() UserOauthID      { return UserOauthID(NewULID().String()) }

// String returns the string representation of the UserOauthID.
func (id UserOauthID) String() string  { return string(id) }

// IsZero returns true if the UserOauthID is empty.
func (id UserOauthID) IsZero() bool    { return id == "" }

// Validate checks if the UserOauthID is a valid ULID.
func (id UserOauthID) Validate() error { return validateULID(string(id), "UserOauthID") }

// Value implements driver.Valuer for database serialization.
func (id UserOauthID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("UserOauthID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *UserOauthID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("UserOauthID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = UserOauthID(v)
	case []byte:
		*id = UserOauthID(string(v))
	default:
		return fmt.Errorf("UserOauthID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id UserOauthID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *UserOauthID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("UserOauthID: %w", err)
	}
	*id = UserOauthID(s)
	return id.Validate()
}

// UserSshKeyID uniquely identifies a user SSH key.
type UserSshKeyID string

// NewUserSshKeyID generates a new ULID-based UserSshKeyID.
func NewUserSshKeyID() UserSshKeyID      { return UserSshKeyID(NewULID().String()) }

// String returns the string representation of the UserSshKeyID.
func (id UserSshKeyID) String() string   { return string(id) }

// IsZero returns true if the UserSshKeyID is empty.
func (id UserSshKeyID) IsZero() bool     { return id == "" }

// Validate checks if the UserSshKeyID is a valid ULID.
func (id UserSshKeyID) Validate() error  { return validateULID(string(id), "UserSshKeyID") }

// Value implements driver.Valuer for database serialization.
func (id UserSshKeyID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("UserSshKeyID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *UserSshKeyID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("UserSshKeyID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = UserSshKeyID(v)
	case []byte:
		*id = UserSshKeyID(string(v))
	default:
		return fmt.Errorf("UserSshKeyID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id UserSshKeyID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *UserSshKeyID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("UserSshKeyID: %w", err)
	}
	*id = UserSshKeyID(s)
	return id.Validate()
}

// AdminDatatypeID uniquely identifies an admin datatype.
type AdminDatatypeID string

// NewAdminDatatypeID generates a new ULID-based AdminDatatypeID.
func NewAdminDatatypeID() AdminDatatypeID    { return AdminDatatypeID(NewULID().String()) }

// String returns the string representation of the AdminDatatypeID.
func (id AdminDatatypeID) String() string    { return string(id) }

// IsZero returns true if the AdminDatatypeID is empty.
func (id AdminDatatypeID) IsZero() bool      { return id == "" }

// Validate checks if the AdminDatatypeID is a valid ULID.
func (id AdminDatatypeID) Validate() error   { return validateULID(string(id), "AdminDatatypeID") }

// Value implements driver.Valuer for database serialization.
func (id AdminDatatypeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminDatatypeID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminDatatypeID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminDatatypeID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminDatatypeID(v)
	case []byte:
		*id = AdminDatatypeID(string(v))
	default:
		return fmt.Errorf("AdminDatatypeID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminDatatypeID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminDatatypeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminDatatypeID: %w", err)
	}
	*id = AdminDatatypeID(s)
	return id.Validate()
}

// AdminFieldID uniquely identifies an admin field.
type AdminFieldID string

// NewAdminFieldID generates a new ULID-based AdminFieldID.
func NewAdminFieldID() AdminFieldID      { return AdminFieldID(NewULID().String()) }

// String returns the string representation of the AdminFieldID.
func (id AdminFieldID) String() string   { return string(id) }

// IsZero returns true if the AdminFieldID is empty.
func (id AdminFieldID) IsZero() bool     { return id == "" }

// Validate checks if the AdminFieldID is a valid ULID.
func (id AdminFieldID) Validate() error  { return validateULID(string(id), "AdminFieldID") }

// Value implements driver.Valuer for database serialization.
func (id AdminFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminFieldID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminFieldID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminFieldID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminFieldID(v)
	case []byte:
		*id = AdminFieldID(string(v))
	default:
		return fmt.Errorf("AdminFieldID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminFieldID: %w", err)
	}
	*id = AdminFieldID(s)
	return id.Validate()
}

// AdminContentID uniquely identifies admin content data.
type AdminContentID string

// NewAdminContentID generates a new ULID-based AdminContentID.
func NewAdminContentID() AdminContentID    { return AdminContentID(NewULID().String()) }

// String returns the string representation of the AdminContentID.
func (id AdminContentID) String() string   { return string(id) }

// IsZero returns true if the AdminContentID is empty.
func (id AdminContentID) IsZero() bool     { return id == "" }

// Validate checks if the AdminContentID is a valid ULID.
func (id AdminContentID) Validate() error  { return validateULID(string(id), "AdminContentID") }

// Value implements driver.Valuer for database serialization.
func (id AdminContentID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminContentID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminContentID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminContentID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminContentID(v)
	case []byte:
		*id = AdminContentID(string(v))
	default:
		return fmt.Errorf("AdminContentID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminContentID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminContentID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminContentID: %w", err)
	}
	*id = AdminContentID(s)
	return id.Validate()
}

// AdminContentFieldID uniquely identifies an admin content field.
type AdminContentFieldID string

// NewAdminContentFieldID generates a new ULID-based AdminContentFieldID.
func NewAdminContentFieldID() AdminContentFieldID    { return AdminContentFieldID(NewULID().String()) }

// String returns the string representation of the AdminContentFieldID.
func (id AdminContentFieldID) String() string        { return string(id) }

// IsZero returns true if the AdminContentFieldID is empty.
func (id AdminContentFieldID) IsZero() bool          { return id == "" }

// Validate checks if the AdminContentFieldID is a valid ULID.
func (id AdminContentFieldID) Validate() error       { return validateULID(string(id), "AdminContentFieldID") }

// Value implements driver.Valuer for database serialization.
func (id AdminContentFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminContentFieldID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminContentFieldID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminContentFieldID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminContentFieldID(v)
	case []byte:
		*id = AdminContentFieldID(string(v))
	default:
		return fmt.Errorf("AdminContentFieldID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminContentFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminContentFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminContentFieldID: %w", err)
	}
	*id = AdminContentFieldID(s)
	return id.Validate()
}

// DatatypeFieldID uniquely identifies a datatype-field relationship.
type DatatypeFieldID string

// NewDatatypeFieldID generates a new ULID-based DatatypeFieldID.
func NewDatatypeFieldID() DatatypeFieldID    { return DatatypeFieldID(NewULID().String()) }

// String returns the string representation of the DatatypeFieldID.
func (id DatatypeFieldID) String() string    { return string(id) }

// IsZero returns true if the DatatypeFieldID is empty.
func (id DatatypeFieldID) IsZero() bool      { return id == "" }

// Validate checks if the DatatypeFieldID is a valid ULID.
func (id DatatypeFieldID) Validate() error   { return validateULID(string(id), "DatatypeFieldID") }

// Value implements driver.Valuer for database serialization.
func (id DatatypeFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("DatatypeFieldID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *DatatypeFieldID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("DatatypeFieldID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = DatatypeFieldID(v)
	case []byte:
		*id = DatatypeFieldID(string(v))
	default:
		return fmt.Errorf("DatatypeFieldID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id DatatypeFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *DatatypeFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("DatatypeFieldID: %w", err)
	}
	*id = DatatypeFieldID(s)
	return id.Validate()
}

// AdminDatatypeFieldID uniquely identifies an admin datatype-field relationship.
type AdminDatatypeFieldID string

// NewAdminDatatypeFieldID generates a new ULID-based AdminDatatypeFieldID.
func NewAdminDatatypeFieldID() AdminDatatypeFieldID    { return AdminDatatypeFieldID(NewULID().String()) }

// String returns the string representation of the AdminDatatypeFieldID.
func (id AdminDatatypeFieldID) String() string         { return string(id) }

// IsZero returns true if the AdminDatatypeFieldID is empty.
func (id AdminDatatypeFieldID) IsZero() bool           { return id == "" }

// Validate checks if the AdminDatatypeFieldID is a valid ULID.
func (id AdminDatatypeFieldID) Validate() error        { return validateULID(string(id), "AdminDatatypeFieldID") }

// Value implements driver.Valuer for database serialization.
func (id AdminDatatypeFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminDatatypeFieldID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminDatatypeFieldID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminDatatypeFieldID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminDatatypeFieldID(v)
	case []byte:
		*id = AdminDatatypeFieldID(string(v))
	default:
		return fmt.Errorf("AdminDatatypeFieldID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminDatatypeFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminDatatypeFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminDatatypeFieldID: %w", err)
	}
	*id = AdminDatatypeFieldID(s)
	return id.Validate()
}

// EventID uniquely identifies a change event.
type EventID string

// NewEventID generates a new ULID-based EventID.
func NewEventID() EventID            { return EventID(NewULID().String()) }

// String returns the string representation of the EventID.
func (id EventID) String() string    { return string(id) }

// IsZero returns true if the EventID is empty.
func (id EventID) IsZero() bool      { return id == "" }

// Validate checks if the EventID is a valid ULID.
func (id EventID) Validate() error   { return validateULID(string(id), "EventID") }

// Value implements driver.Valuer for database serialization.
func (id EventID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("EventID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *EventID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("EventID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = EventID(v)
	case []byte:
		*id = EventID(string(v))
	default:
		return fmt.Errorf("EventID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id EventID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *EventID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("EventID: %w", err)
	}
	*id = EventID(s)
	return id.Validate()
}

// NodeID uniquely identifies a node in a distributed deployment.
type NodeID string

// NewNodeID generates a new ULID-based NodeID.
func NewNodeID() NodeID              { return NodeID(NewULID().String()) }

// String returns the string representation of the NodeID.
func (id NodeID) String() string     { return string(id) }

// IsZero returns true if the NodeID is empty.
func (id NodeID) IsZero() bool       { return id == "" }

// Validate checks if the NodeID is a valid ULID.
func (id NodeID) Validate() error    { return validateULID(string(id), "NodeID") }

// Value implements driver.Valuer for database serialization.
func (id NodeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("NodeID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *NodeID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("NodeID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = NodeID(v)
	case []byte:
		*id = NodeID(string(v))
	default:
		return fmt.Errorf("NodeID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id NodeID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *NodeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("NodeID: %w", err)
	}
	*id = NodeID(s)
	return id.Validate()
}

// BackupID uniquely identifies a backup.
type BackupID string

// NewBackupID generates a new ULID-based BackupID.
func NewBackupID() BackupID          { return BackupID(NewULID().String()) }

// String returns the string representation of the BackupID.
func (id BackupID) String() string   { return string(id) }

// IsZero returns true if the BackupID is empty.
func (id BackupID) IsZero() bool     { return id == "" }

// Validate checks if the BackupID is a valid ULID.
func (id BackupID) Validate() error  { return validateULID(string(id), "BackupID") }

// Value implements driver.Valuer for database serialization.
func (id BackupID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("BackupID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *BackupID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("BackupID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = BackupID(v)
	case []byte:
		*id = BackupID(string(v))
	default:
		return fmt.Errorf("BackupID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id BackupID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *BackupID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("BackupID: %w", err)
	}
	*id = BackupID(s)
	return id.Validate()
}

// VerificationID uniquely identifies a backup verification.
type VerificationID string

// NewVerificationID generates a new ULID-based VerificationID.
func NewVerificationID() VerificationID    { return VerificationID(NewULID().String()) }

// String returns the string representation of the VerificationID.
func (id VerificationID) String() string   { return string(id) }

// IsZero returns true if the VerificationID is empty.
func (id VerificationID) IsZero() bool     { return id == "" }

// Validate checks if the VerificationID is a valid ULID.
func (id VerificationID) Validate() error  { return validateULID(string(id), "VerificationID") }

// Value implements driver.Valuer for database serialization.
func (id VerificationID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("VerificationID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *VerificationID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("VerificationID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = VerificationID(v)
	case []byte:
		*id = VerificationID(string(v))
	default:
		return fmt.Errorf("VerificationID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id VerificationID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *VerificationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("VerificationID: %w", err)
	}
	*id = VerificationID(s)
	return id.Validate()
}

// BackupSetID uniquely identifies a backup set.
type BackupSetID string

// NewBackupSetID generates a new ULID-based BackupSetID.
func NewBackupSetID() BackupSetID      { return BackupSetID(NewULID().String()) }

// String returns the string representation of the BackupSetID.
func (id BackupSetID) String() string  { return string(id) }

// IsZero returns true if the BackupSetID is empty.
func (id BackupSetID) IsZero() bool    { return id == "" }

// Validate checks if the BackupSetID is a valid ULID.
func (id BackupSetID) Validate() error { return validateULID(string(id), "BackupSetID") }

// Value implements driver.Valuer for database serialization.
func (id BackupSetID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("BackupSetID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *BackupSetID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("BackupSetID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = BackupSetID(v)
	case []byte:
		*id = BackupSetID(string(v))
	default:
		return fmt.Errorf("BackupSetID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id BackupSetID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *BackupSetID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("BackupSetID: %w", err)
	}
	*id = BackupSetID(s)
	return id.Validate()
}

// ContentRelationID uniquely identifies a content relation.
type ContentRelationID string

// NewContentRelationID generates a new ULID-based ContentRelationID.
func NewContentRelationID() ContentRelationID    { return ContentRelationID(NewULID().String()) }

// String returns the string representation of the ContentRelationID.
func (id ContentRelationID) String() string      { return string(id) }

// IsZero returns true if the ContentRelationID is empty.
func (id ContentRelationID) IsZero() bool        { return id == "" }

// Validate checks if the ContentRelationID is a valid ULID.
func (id ContentRelationID) Validate() error     { return validateULID(string(id), "ContentRelationID") }

// ULID parses the ContentRelationID as a ulid.ULID.
func (id ContentRelationID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

// Time extracts the timestamp embedded in the ContentRelationID.
func (id ContentRelationID) Time() (time.Time, error) {
	u, err := id.ULID()
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(u.Time()), nil
}

// ParseContentRelationID parses and validates a string as a ContentRelationID.
func ParseContentRelationID(s string) (ContentRelationID, error) {
	id := ContentRelationID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// Value implements driver.Valuer for database serialization.
func (id ContentRelationID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("ContentRelationID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *ContentRelationID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("ContentRelationID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = ContentRelationID(v)
	case []byte:
		*id = ContentRelationID(string(v))
	default:
		return fmt.Errorf("ContentRelationID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id ContentRelationID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *ContentRelationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ContentRelationID: %w", err)
	}
	*id = ContentRelationID(s)
	return id.Validate()
}

// AdminContentRelationID uniquely identifies an admin content relation.
type AdminContentRelationID string

// NewAdminContentRelationID generates a new ULID-based AdminContentRelationID.
func NewAdminContentRelationID() AdminContentRelationID { return AdminContentRelationID(NewULID().String()) }

// String returns the string representation of the AdminContentRelationID.
func (id AdminContentRelationID) String() string        { return string(id) }

// IsZero returns true if the AdminContentRelationID is empty.
func (id AdminContentRelationID) IsZero() bool          { return id == "" }

// Validate checks if the AdminContentRelationID is a valid ULID.
func (id AdminContentRelationID) Validate() error       { return validateULID(string(id), "AdminContentRelationID") }

// ULID parses the AdminContentRelationID as a ulid.ULID.
func (id AdminContentRelationID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

// Time extracts the timestamp embedded in the AdminContentRelationID.
func (id AdminContentRelationID) Time() (time.Time, error) {
	u, err := id.ULID()
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(u.Time()), nil
}

// ParseAdminContentRelationID parses and validates a string as a AdminContentRelationID.
func ParseAdminContentRelationID(s string) (AdminContentRelationID, error) {
	id := AdminContentRelationID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// Value implements driver.Valuer for database serialization.
func (id AdminContentRelationID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminContentRelationID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminContentRelationID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminContentRelationID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminContentRelationID(v)
	case []byte:
		*id = AdminContentRelationID(string(v))
	default:
		return fmt.Errorf("AdminContentRelationID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminContentRelationID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminContentRelationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminContentRelationID: %w", err)
	}
	*id = AdminContentRelationID(s)
	return id.Validate()
}

// RolePermissionID uniquely identifies a role-permission junction row.
type RolePermissionID string

// NewRolePermissionID generates a new ULID-based RolePermissionID.
func NewRolePermissionID() RolePermissionID { return RolePermissionID(NewULID().String()) }

// String returns the string representation of the RolePermissionID.
func (id RolePermissionID) String() string { return string(id) }

// IsZero returns true if the RolePermissionID is empty.
func (id RolePermissionID) IsZero() bool { return id == "" }

// Validate checks if the RolePermissionID is a valid ULID.
func (id RolePermissionID) Validate() error { return validateULID(string(id), "RolePermissionID") }

// ULID parses the RolePermissionID as a ulid.ULID.
func (id RolePermissionID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

// Time extracts the timestamp embedded in the RolePermissionID.
func (id RolePermissionID) Time() (time.Time, error) {
	u, err := id.ULID()
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(u.Time()), nil
}

// ParseRolePermissionID parses and validates a string as a RolePermissionID.
func ParseRolePermissionID(s string) (RolePermissionID, error) {
	id := RolePermissionID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// Value implements driver.Valuer for database serialization.
func (id RolePermissionID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("RolePermissionID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *RolePermissionID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("RolePermissionID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = RolePermissionID(v)
	case []byte:
		*id = RolePermissionID(string(v))
	default:
		return fmt.Errorf("RolePermissionID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id RolePermissionID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *RolePermissionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("RolePermissionID: %w", err)
	}
	*id = RolePermissionID(s)
	return id.Validate()
}

// FieldTypeID uniquely identifies a field type.
type FieldTypeID string

// NewFieldTypeID generates a new ULID-based FieldTypeID.
func NewFieldTypeID() FieldTypeID { return FieldTypeID(NewULID().String()) }

// String returns the string representation of the FieldTypeID.
func (id FieldTypeID) String() string { return string(id) }

// IsZero returns true if the FieldTypeID is empty.
func (id FieldTypeID) IsZero() bool { return id == "" }

// Validate checks if the FieldTypeID is a valid ULID.
func (id FieldTypeID) Validate() error { return validateULID(string(id), "FieldTypeID") }

// ULID parses the FieldTypeID as a ulid.ULID.
func (id FieldTypeID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

// Time extracts the timestamp embedded in the FieldTypeID.
func (id FieldTypeID) Time() (time.Time, error) {
	u, err := id.ULID()
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(u.Time()), nil
}

// ParseFieldTypeID parses and validates a string as a FieldTypeID.
func ParseFieldTypeID(s string) (FieldTypeID, error) {
	id := FieldTypeID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// Value implements driver.Valuer for database serialization.
func (id FieldTypeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("FieldTypeID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *FieldTypeID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("FieldTypeID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = FieldTypeID(v)
	case []byte:
		*id = FieldTypeID(string(v))
	default:
		return fmt.Errorf("FieldTypeID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id FieldTypeID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *FieldTypeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("FieldTypeID: %w", err)
	}
	*id = FieldTypeID(s)
	return id.Validate()
}

// AdminFieldTypeID uniquely identifies an admin field type.
type AdminFieldTypeID string

// NewAdminFieldTypeID generates a new ULID-based AdminFieldTypeID.
func NewAdminFieldTypeID() AdminFieldTypeID { return AdminFieldTypeID(NewULID().String()) }

// String returns the string representation of the AdminFieldTypeID.
func (id AdminFieldTypeID) String() string { return string(id) }

// IsZero returns true if the AdminFieldTypeID is empty.
func (id AdminFieldTypeID) IsZero() bool { return id == "" }

// Validate checks if the AdminFieldTypeID is a valid ULID.
func (id AdminFieldTypeID) Validate() error { return validateULID(string(id), "AdminFieldTypeID") }

// ULID parses the AdminFieldTypeID as a ulid.ULID.
func (id AdminFieldTypeID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

// Time extracts the timestamp embedded in the AdminFieldTypeID.
func (id AdminFieldTypeID) Time() (time.Time, error) {
	u, err := id.ULID()
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(u.Time()), nil
}

// ParseAdminFieldTypeID parses and validates a string as an AdminFieldTypeID.
func ParseAdminFieldTypeID(s string) (AdminFieldTypeID, error) {
	id := AdminFieldTypeID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// Value implements driver.Valuer for database serialization.
func (id AdminFieldTypeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminFieldTypeID: cannot be empty")
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database deserialization.
func (id *AdminFieldTypeID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("AdminFieldTypeID: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*id = AdminFieldTypeID(v)
	case []byte:
		*id = AdminFieldTypeID(string(v))
	default:
		return fmt.Errorf("AdminFieldTypeID: cannot scan %T", value)
	}
	return id.Validate()
}

// MarshalJSON implements json.Marshaler.
func (id AdminFieldTypeID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *AdminFieldTypeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminFieldTypeID: %w", err)
	}
	*id = AdminFieldTypeID(s)
	return id.Validate()
}
