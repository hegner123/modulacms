package modula

import "time"

// Content IDs

// ContentID identifies a content data item.
type ContentID string

func (id ContentID) String() string { return string(id) }
func (id ContentID) IsZero() bool   { return id == "" }

// ContentFieldID identifies a content field.
type ContentFieldID string

func (id ContentFieldID) String() string { return string(id) }
func (id ContentFieldID) IsZero() bool   { return id == "" }

// ContentRelationID identifies a content relation.
type ContentRelationID string

func (id ContentRelationID) String() string { return string(id) }
func (id ContentRelationID) IsZero() bool   { return id == "" }

// Admin content IDs

// AdminContentID identifies an admin content data item.
type AdminContentID string

func (id AdminContentID) String() string { return string(id) }
func (id AdminContentID) IsZero() bool   { return id == "" }

// AdminContentFieldID identifies an admin content field.
type AdminContentFieldID string

func (id AdminContentFieldID) String() string { return string(id) }
func (id AdminContentFieldID) IsZero() bool   { return id == "" }

// AdminContentRelationID identifies an admin content relation.
type AdminContentRelationID string

func (id AdminContentRelationID) String() string { return string(id) }
func (id AdminContentRelationID) IsZero() bool   { return id == "" }

// Schema IDs

// DatatypeID identifies a datatype.
type DatatypeID string

func (id DatatypeID) String() string { return string(id) }
func (id DatatypeID) IsZero() bool   { return id == "" }

// DatatypeFieldID identifies a datatype-field mapping.
type DatatypeFieldID string

func (id DatatypeFieldID) String() string { return string(id) }
func (id DatatypeFieldID) IsZero() bool   { return id == "" }

// FieldID identifies a field definition.
type FieldID string

func (id FieldID) String() string { return string(id) }
func (id FieldID) IsZero() bool   { return id == "" }

// Admin schema IDs

// AdminDatatypeID identifies an admin datatype.
type AdminDatatypeID string

func (id AdminDatatypeID) String() string { return string(id) }
func (id AdminDatatypeID) IsZero() bool   { return id == "" }

// AdminDatatypeFieldID identifies an admin datatype-field mapping.
type AdminDatatypeFieldID string

func (id AdminDatatypeFieldID) String() string { return string(id) }
func (id AdminDatatypeFieldID) IsZero() bool   { return id == "" }

// AdminFieldID identifies an admin field definition.
type AdminFieldID string

func (id AdminFieldID) String() string { return string(id) }
func (id AdminFieldID) IsZero() bool   { return id == "" }

// Media IDs

// MediaID identifies a media item.
type MediaID string

func (id MediaID) String() string { return string(id) }
func (id MediaID) IsZero() bool   { return id == "" }

// MediaDimensionID identifies a media dimension preset.
type MediaDimensionID string

func (id MediaDimensionID) String() string { return string(id) }
func (id MediaDimensionID) IsZero() bool   { return id == "" }

// Auth IDs

// UserID identifies a user.
type UserID string

func (id UserID) String() string { return string(id) }
func (id UserID) IsZero() bool   { return id == "" }

// RoleID identifies a role.
type RoleID string

func (id RoleID) String() string { return string(id) }
func (id RoleID) IsZero() bool   { return id == "" }

// SessionID identifies a session.
type SessionID string

func (id SessionID) String() string { return string(id) }
func (id SessionID) IsZero() bool   { return id == "" }

// TokenID identifies a token.
type TokenID string

func (id TokenID) String() string { return string(id) }
func (id TokenID) IsZero() bool   { return id == "" }

// UserOauthID identifies a user OAuth connection.
type UserOauthID string

func (id UserOauthID) String() string { return string(id) }
func (id UserOauthID) IsZero() bool   { return id == "" }

// UserSshKeyID identifies a user SSH key.
type UserSshKeyID string

func (id UserSshKeyID) String() string { return string(id) }
func (id UserSshKeyID) IsZero() bool   { return id == "" }

// PermissionID identifies a permission.
type PermissionID string

func (id PermissionID) String() string { return string(id) }
func (id PermissionID) IsZero() bool   { return id == "" }

// RolePermissionID identifies a role-permission junction row.
type RolePermissionID string

func (id RolePermissionID) String() string { return string(id) }
func (id RolePermissionID) IsZero() bool   { return id == "" }

// Routing IDs

// RouteID identifies a route.
type RouteID string

func (id RouteID) String() string { return string(id) }
func (id RouteID) IsZero() bool   { return id == "" }

// AdminRouteID identifies an admin route.
type AdminRouteID string

func (id AdminRouteID) String() string { return string(id) }
func (id AdminRouteID) IsZero() bool   { return id == "" }

// Other IDs

// TableID identifies a table.
type TableID string

func (id TableID) String() string { return string(id) }
func (id TableID) IsZero() bool   { return id == "" }

// EventID identifies a change event.
type EventID string

func (id EventID) String() string { return string(id) }
func (id EventID) IsZero() bool   { return id == "" }

// BackupID identifies a backup.
type BackupID string

func (id BackupID) String() string { return string(id) }
func (id BackupID) IsZero() bool   { return id == "" }

// Value types

// Slug represents a URL slug.
type Slug string

func (s Slug) String() string { return string(s) }
func (s Slug) IsZero() bool   { return s == "" }

// Email represents an email address.
type Email string

func (e Email) String() string { return string(e) }
func (e Email) IsZero() bool   { return e == "" }

// URL represents a URL string.
type URL string

func (u URL) String() string { return string(u) }
func (u URL) IsZero() bool   { return u == "" }

// Timestamp represents an RFC3339 UTC timestamp from the API.
type Timestamp string

func (t Timestamp) String() string { return string(t) }
func (t Timestamp) IsZero() bool   { return t == "" }

// Time parses the timestamp into a time.Time value.
func (t Timestamp) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(t))
}

// NewTimestamp creates a Timestamp from a time.Time value.
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp(t.UTC().Format(time.RFC3339))
}

// TimestampNow returns the current time as a Timestamp.
func TimestampNow() Timestamp {
	return NewTimestamp(time.Now())
}
