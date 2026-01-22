package db

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

// DatatypeID uniquely identifies a datatype (26-char ULID string)
type DatatypeID string

func NewDatatypeID() DatatypeID              { return DatatypeID(NewULID().String()) }
func (id DatatypeID) String() string         { return string(id) }
func (id DatatypeID) IsZero() bool           { return id == "" }
func (id DatatypeID) Validate() error        { return validateULID(string(id), "DatatypeID") }
func (id DatatypeID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

func (id DatatypeID) Time() (time.Time, error) {
	u, err := id.ULID()
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(u.Time()), nil
}

func (id DatatypeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("DatatypeID: cannot be empty")
	}
	return string(id), nil
}

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

func (id DatatypeID) MarshalJSON() ([]byte, error)  { return json.Marshal(string(id)) }
func (id *DatatypeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("DatatypeID: %w", err)
	}
	*id = DatatypeID(s)
	return id.Validate()
}

func ParseDatatypeID(s string) (DatatypeID, error) {
	id := DatatypeID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// UserID uniquely identifies a user
type UserID string

func NewUserID() UserID              { return UserID(NewULID().String()) }
func (id UserID) String() string     { return string(id) }
func (id UserID) IsZero() bool       { return id == "" }
func (id UserID) Validate() error    { return validateULID(string(id), "UserID") }
func (id UserID) ULID() (ulid.ULID, error) { return ulid.Parse(string(id)) }

func (id UserID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("UserID: cannot be empty")
	}
	return string(id), nil
}

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

func (id UserID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *UserID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("UserID: %w", err)
	}
	*id = UserID(s)
	return id.Validate()
}

func ParseUserID(s string) (UserID, error) {
	id := UserID(s)
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

// RoleID uniquely identifies a role
type RoleID string

func NewRoleID() RoleID              { return RoleID(NewULID().String()) }
func (id RoleID) String() string     { return string(id) }
func (id RoleID) IsZero() bool       { return id == "" }
func (id RoleID) Validate() error    { return validateULID(string(id), "RoleID") }

func (id RoleID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("RoleID: cannot be empty")
	}
	return string(id), nil
}

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

func (id RoleID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *RoleID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("RoleID: %w", err)
	}
	*id = RoleID(s)
	return id.Validate()
}

// PermissionID uniquely identifies a permission
type PermissionID string

func NewPermissionID() PermissionID      { return PermissionID(NewULID().String()) }
func (id PermissionID) String() string   { return string(id) }
func (id PermissionID) IsZero() bool     { return id == "" }
func (id PermissionID) Validate() error  { return validateULID(string(id), "PermissionID") }

func (id PermissionID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("PermissionID: cannot be empty")
	}
	return string(id), nil
}

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

func (id PermissionID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *PermissionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("PermissionID: %w", err)
	}
	*id = PermissionID(s)
	return id.Validate()
}

// FieldID uniquely identifies a field
type FieldID string

func NewFieldID() FieldID            { return FieldID(NewULID().String()) }
func (id FieldID) String() string    { return string(id) }
func (id FieldID) IsZero() bool      { return id == "" }
func (id FieldID) Validate() error   { return validateULID(string(id), "FieldID") }

func (id FieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("FieldID: cannot be empty")
	}
	return string(id), nil
}

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

func (id FieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *FieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("FieldID: %w", err)
	}
	*id = FieldID(s)
	return id.Validate()
}

// ContentID uniquely identifies content data
type ContentID string

func NewContentID() ContentID        { return ContentID(NewULID().String()) }
func (id ContentID) String() string  { return string(id) }
func (id ContentID) IsZero() bool    { return id == "" }
func (id ContentID) Validate() error { return validateULID(string(id), "ContentID") }

func (id ContentID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("ContentID: cannot be empty")
	}
	return string(id), nil
}

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

func (id ContentID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *ContentID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ContentID: %w", err)
	}
	*id = ContentID(s)
	return id.Validate()
}

// ContentFieldID uniquely identifies a content field
type ContentFieldID string

func NewContentFieldID() ContentFieldID    { return ContentFieldID(NewULID().String()) }
func (id ContentFieldID) String() string   { return string(id) }
func (id ContentFieldID) IsZero() bool     { return id == "" }
func (id ContentFieldID) Validate() error  { return validateULID(string(id), "ContentFieldID") }

func (id ContentFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("ContentFieldID: cannot be empty")
	}
	return string(id), nil
}

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

func (id ContentFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *ContentFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ContentFieldID: %w", err)
	}
	*id = ContentFieldID(s)
	return id.Validate()
}

// MediaID uniquely identifies media
type MediaID string

func NewMediaID() MediaID            { return MediaID(NewULID().String()) }
func (id MediaID) String() string    { return string(id) }
func (id MediaID) IsZero() bool      { return id == "" }
func (id MediaID) Validate() error   { return validateULID(string(id), "MediaID") }

func (id MediaID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("MediaID: cannot be empty")
	}
	return string(id), nil
}

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

func (id MediaID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *MediaID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("MediaID: %w", err)
	}
	*id = MediaID(s)
	return id.Validate()
}

// MediaDimensionID uniquely identifies a media dimension
type MediaDimensionID string

func NewMediaDimensionID() MediaDimensionID    { return MediaDimensionID(NewULID().String()) }
func (id MediaDimensionID) String() string     { return string(id) }
func (id MediaDimensionID) IsZero() bool       { return id == "" }
func (id MediaDimensionID) Validate() error    { return validateULID(string(id), "MediaDimensionID") }

func (id MediaDimensionID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("MediaDimensionID: cannot be empty")
	}
	return string(id), nil
}

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

func (id MediaDimensionID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *MediaDimensionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("MediaDimensionID: %w", err)
	}
	*id = MediaDimensionID(s)
	return id.Validate()
}

// SessionID uniquely identifies a session
type SessionID string

func NewSessionID() SessionID        { return SessionID(NewULID().String()) }
func (id SessionID) String() string  { return string(id) }
func (id SessionID) IsZero() bool    { return id == "" }
func (id SessionID) Validate() error { return validateULID(string(id), "SessionID") }

func (id SessionID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("SessionID: cannot be empty")
	}
	return string(id), nil
}

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

func (id SessionID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *SessionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("SessionID: %w", err)
	}
	*id = SessionID(s)
	return id.Validate()
}

// TokenID uniquely identifies a token
type TokenID string

func NewTokenID() TokenID            { return TokenID(NewULID().String()) }
func (id TokenID) String() string    { return string(id) }
func (id TokenID) IsZero() bool      { return id == "" }
func (id TokenID) Validate() error   { return validateULID(string(id), "TokenID") }

func (id TokenID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("TokenID: cannot be empty")
	}
	return string(id), nil
}

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

func (id TokenID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *TokenID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TokenID: %w", err)
	}
	*id = TokenID(s)
	return id.Validate()
}

// RouteID uniquely identifies a route
type RouteID string

func NewRouteID() RouteID            { return RouteID(NewULID().String()) }
func (id RouteID) String() string    { return string(id) }
func (id RouteID) IsZero() bool      { return id == "" }
func (id RouteID) Validate() error   { return validateULID(string(id), "RouteID") }

func (id RouteID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("RouteID: cannot be empty")
	}
	return string(id), nil
}

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

func (id RouteID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *RouteID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("RouteID: %w", err)
	}
	*id = RouteID(s)
	return id.Validate()
}

// AdminRouteID uniquely identifies an admin route
type AdminRouteID string

func NewAdminRouteID() AdminRouteID      { return AdminRouteID(NewULID().String()) }
func (id AdminRouteID) String() string   { return string(id) }
func (id AdminRouteID) IsZero() bool     { return id == "" }
func (id AdminRouteID) Validate() error  { return validateULID(string(id), "AdminRouteID") }

func (id AdminRouteID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminRouteID: cannot be empty")
	}
	return string(id), nil
}

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

func (id AdminRouteID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *AdminRouteID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminRouteID: %w", err)
	}
	*id = AdminRouteID(s)
	return id.Validate()
}

// TableID uniquely identifies a table
type TableID string

func NewTableID() TableID            { return TableID(NewULID().String()) }
func (id TableID) String() string    { return string(id) }
func (id TableID) IsZero() bool      { return id == "" }
func (id TableID) Validate() error   { return validateULID(string(id), "TableID") }

func (id TableID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("TableID: cannot be empty")
	}
	return string(id), nil
}

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

func (id TableID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *TableID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TableID: %w", err)
	}
	*id = TableID(s)
	return id.Validate()
}

// UserOauthID uniquely identifies a user OAuth record
type UserOauthID string

func NewUserOauthID() UserOauthID      { return UserOauthID(NewULID().String()) }
func (id UserOauthID) String() string  { return string(id) }
func (id UserOauthID) IsZero() bool    { return id == "" }
func (id UserOauthID) Validate() error { return validateULID(string(id), "UserOauthID") }

func (id UserOauthID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("UserOauthID: cannot be empty")
	}
	return string(id), nil
}

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

func (id UserOauthID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *UserOauthID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("UserOauthID: %w", err)
	}
	*id = UserOauthID(s)
	return id.Validate()
}

// UserSshKeyID uniquely identifies a user SSH key
type UserSshKeyID string

func NewUserSshKeyID() UserSshKeyID      { return UserSshKeyID(NewULID().String()) }
func (id UserSshKeyID) String() string   { return string(id) }
func (id UserSshKeyID) IsZero() bool     { return id == "" }
func (id UserSshKeyID) Validate() error  { return validateULID(string(id), "UserSshKeyID") }

func (id UserSshKeyID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("UserSshKeyID: cannot be empty")
	}
	return string(id), nil
}

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

func (id UserSshKeyID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *UserSshKeyID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("UserSshKeyID: %w", err)
	}
	*id = UserSshKeyID(s)
	return id.Validate()
}

// AdminDatatypeID uniquely identifies an admin datatype
type AdminDatatypeID string

func NewAdminDatatypeID() AdminDatatypeID    { return AdminDatatypeID(NewULID().String()) }
func (id AdminDatatypeID) String() string    { return string(id) }
func (id AdminDatatypeID) IsZero() bool      { return id == "" }
func (id AdminDatatypeID) Validate() error   { return validateULID(string(id), "AdminDatatypeID") }

func (id AdminDatatypeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminDatatypeID: cannot be empty")
	}
	return string(id), nil
}

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

func (id AdminDatatypeID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *AdminDatatypeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminDatatypeID: %w", err)
	}
	*id = AdminDatatypeID(s)
	return id.Validate()
}

// AdminFieldID uniquely identifies an admin field
type AdminFieldID string

func NewAdminFieldID() AdminFieldID      { return AdminFieldID(NewULID().String()) }
func (id AdminFieldID) String() string   { return string(id) }
func (id AdminFieldID) IsZero() bool     { return id == "" }
func (id AdminFieldID) Validate() error  { return validateULID(string(id), "AdminFieldID") }

func (id AdminFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminFieldID: cannot be empty")
	}
	return string(id), nil
}

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

func (id AdminFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *AdminFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminFieldID: %w", err)
	}
	*id = AdminFieldID(s)
	return id.Validate()
}

// AdminContentID uniquely identifies admin content data
type AdminContentID string

func NewAdminContentID() AdminContentID    { return AdminContentID(NewULID().String()) }
func (id AdminContentID) String() string   { return string(id) }
func (id AdminContentID) IsZero() bool     { return id == "" }
func (id AdminContentID) Validate() error  { return validateULID(string(id), "AdminContentID") }

func (id AdminContentID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminContentID: cannot be empty")
	}
	return string(id), nil
}

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

func (id AdminContentID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *AdminContentID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminContentID: %w", err)
	}
	*id = AdminContentID(s)
	return id.Validate()
}

// AdminContentFieldID uniquely identifies an admin content field
type AdminContentFieldID string

func NewAdminContentFieldID() AdminContentFieldID    { return AdminContentFieldID(NewULID().String()) }
func (id AdminContentFieldID) String() string        { return string(id) }
func (id AdminContentFieldID) IsZero() bool          { return id == "" }
func (id AdminContentFieldID) Validate() error       { return validateULID(string(id), "AdminContentFieldID") }

func (id AdminContentFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminContentFieldID: cannot be empty")
	}
	return string(id), nil
}

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

func (id AdminContentFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *AdminContentFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminContentFieldID: %w", err)
	}
	*id = AdminContentFieldID(s)
	return id.Validate()
}

// DatatypeFieldID uniquely identifies a datatype-field relationship
type DatatypeFieldID string

func NewDatatypeFieldID() DatatypeFieldID    { return DatatypeFieldID(NewULID().String()) }
func (id DatatypeFieldID) String() string    { return string(id) }
func (id DatatypeFieldID) IsZero() bool      { return id == "" }
func (id DatatypeFieldID) Validate() error   { return validateULID(string(id), "DatatypeFieldID") }

func (id DatatypeFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("DatatypeFieldID: cannot be empty")
	}
	return string(id), nil
}

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

func (id DatatypeFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *DatatypeFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("DatatypeFieldID: %w", err)
	}
	*id = DatatypeFieldID(s)
	return id.Validate()
}

// AdminDatatypeFieldID uniquely identifies an admin datatype-field relationship
type AdminDatatypeFieldID string

func NewAdminDatatypeFieldID() AdminDatatypeFieldID    { return AdminDatatypeFieldID(NewULID().String()) }
func (id AdminDatatypeFieldID) String() string         { return string(id) }
func (id AdminDatatypeFieldID) IsZero() bool           { return id == "" }
func (id AdminDatatypeFieldID) Validate() error        { return validateULID(string(id), "AdminDatatypeFieldID") }

func (id AdminDatatypeFieldID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("AdminDatatypeFieldID: cannot be empty")
	}
	return string(id), nil
}

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

func (id AdminDatatypeFieldID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *AdminDatatypeFieldID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminDatatypeFieldID: %w", err)
	}
	*id = AdminDatatypeFieldID(s)
	return id.Validate()
}

// EventID uniquely identifies a change event
type EventID string

func NewEventID() EventID            { return EventID(NewULID().String()) }
func (id EventID) String() string    { return string(id) }
func (id EventID) IsZero() bool      { return id == "" }
func (id EventID) Validate() error   { return validateULID(string(id), "EventID") }

func (id EventID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("EventID: cannot be empty")
	}
	return string(id), nil
}

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

func (id EventID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *EventID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("EventID: %w", err)
	}
	*id = EventID(s)
	return id.Validate()
}

// NodeID uniquely identifies a node in a distributed deployment
type NodeID string

func NewNodeID() NodeID              { return NodeID(NewULID().String()) }
func (id NodeID) String() string     { return string(id) }
func (id NodeID) IsZero() bool       { return id == "" }
func (id NodeID) Validate() error    { return validateULID(string(id), "NodeID") }

func (id NodeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("NodeID: cannot be empty")
	}
	return string(id), nil
}

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

func (id NodeID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *NodeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("NodeID: %w", err)
	}
	*id = NodeID(s)
	return id.Validate()
}

// BackupID uniquely identifies a backup
type BackupID string

func NewBackupID() BackupID          { return BackupID(NewULID().String()) }
func (id BackupID) String() string   { return string(id) }
func (id BackupID) IsZero() bool     { return id == "" }
func (id BackupID) Validate() error  { return validateULID(string(id), "BackupID") }

func (id BackupID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("BackupID: cannot be empty")
	}
	return string(id), nil
}

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

func (id BackupID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *BackupID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("BackupID: %w", err)
	}
	*id = BackupID(s)
	return id.Validate()
}

// VerificationID uniquely identifies a backup verification
type VerificationID string

func NewVerificationID() VerificationID    { return VerificationID(NewULID().String()) }
func (id VerificationID) String() string   { return string(id) }
func (id VerificationID) IsZero() bool     { return id == "" }
func (id VerificationID) Validate() error  { return validateULID(string(id), "VerificationID") }

func (id VerificationID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("VerificationID: cannot be empty")
	}
	return string(id), nil
}

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

func (id VerificationID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *VerificationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("VerificationID: %w", err)
	}
	*id = VerificationID(s)
	return id.Validate()
}

// BackupSetID uniquely identifies a backup set
type BackupSetID string

func NewBackupSetID() BackupSetID      { return BackupSetID(NewULID().String()) }
func (id BackupSetID) String() string  { return string(id) }
func (id BackupSetID) IsZero() bool    { return id == "" }
func (id BackupSetID) Validate() error { return validateULID(string(id), "BackupSetID") }

func (id BackupSetID) Value() (driver.Value, error) {
	if id == "" {
		return nil, fmt.Errorf("BackupSetID: cannot be empty")
	}
	return string(id), nil
}

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

func (id BackupSetID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }
func (id *BackupSetID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("BackupSetID: %w", err)
	}
	*id = BackupSetID(s)
	return id.Validate()
}
