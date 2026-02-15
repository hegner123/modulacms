package types

import (
	"database/sql/driver"
	"encoding/json"
)

// NullableDatatypeID represents a nullable foreign key to datatypes
type NullableDatatypeID struct {
	ID    DatatypeID
	Valid bool
}

// Validate checks if the NullableDatatypeID is valid if set.
func (n NullableDatatypeID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableDatatypeID.
func (n NullableDatatypeID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableDatatypeID is not set or empty.
func (n NullableDatatypeID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableDatatypeID.
func (n NullableDatatypeID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableDatatypeID.
func (n *NullableDatatypeID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableDatatypeID.
func (n NullableDatatypeID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableDatatypeID.
func (n *NullableDatatypeID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableUserID represents a nullable foreign key to users
type NullableUserID struct {
	ID    UserID
	Valid bool
}

// Validate checks if the NullableUserID is valid if set.
func (n NullableUserID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableUserID.
func (n NullableUserID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableUserID is not set or empty.
func (n NullableUserID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableUserID.
func (n NullableUserID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableUserID.
func (n *NullableUserID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableUserID.
func (n NullableUserID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableUserID.
func (n *NullableUserID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableContentID represents a nullable foreign key to content
type NullableContentID struct {
	ID    ContentID
	Valid bool
}

// Validate checks if the NullableContentID is valid if set.
func (n NullableContentID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableContentID.
func (n NullableContentID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableContentID is not set or empty.
func (n NullableContentID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableContentID.
func (n NullableContentID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableContentID.
func (n *NullableContentID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableContentID.
func (n NullableContentID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableContentID.
func (n *NullableContentID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableFieldID represents a nullable foreign key to fields
type NullableFieldID struct {
	ID    FieldID
	Valid bool
}

// Validate checks if the NullableFieldID is valid if set.
func (n NullableFieldID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableFieldID.
func (n NullableFieldID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableFieldID is not set or empty.
func (n NullableFieldID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableFieldID.
func (n NullableFieldID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableFieldID.
func (n *NullableFieldID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableFieldID.
func (n NullableFieldID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableFieldID.
func (n *NullableFieldID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableMediaID represents a nullable foreign key to media
type NullableMediaID struct {
	ID    MediaID
	Valid bool
}

// Validate checks if the NullableMediaID is valid if set.
func (n NullableMediaID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableMediaID.
func (n NullableMediaID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableMediaID is not set or empty.
func (n NullableMediaID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableMediaID.
func (n NullableMediaID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableMediaID.
func (n *NullableMediaID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableMediaID.
func (n NullableMediaID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableMediaID.
func (n *NullableMediaID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableRouteID represents a nullable foreign key to routes
type NullableRouteID struct {
	ID    RouteID
	Valid bool
}

// Validate checks if the NullableRouteID is valid if set.
func (n NullableRouteID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableRouteID.
func (n NullableRouteID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableRouteID is not set or empty.
func (n NullableRouteID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableRouteID.
func (n NullableRouteID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableRouteID.
func (n *NullableRouteID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableRouteID.
func (n NullableRouteID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableRouteID.
func (n *NullableRouteID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableAdminRouteID represents a nullable foreign key to admin routes
type NullableAdminRouteID struct {
	ID    AdminRouteID
	Valid bool
}

// Validate checks if the NullableAdminRouteID is valid if set.
func (n NullableAdminRouteID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableAdminRouteID.
func (n NullableAdminRouteID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableAdminRouteID is not set or empty.
func (n NullableAdminRouteID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableAdminRouteID.
func (n NullableAdminRouteID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableAdminRouteID.
func (n *NullableAdminRouteID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableAdminRouteID.
func (n NullableAdminRouteID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableAdminRouteID.
func (n *NullableAdminRouteID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableAdminContentID represents a nullable foreign key to admin content
type NullableAdminContentID struct {
	ID    AdminContentID
	Valid bool
}

// Validate checks if the NullableAdminContentID is valid if set.
func (n NullableAdminContentID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableAdminContentID.
func (n NullableAdminContentID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableAdminContentID is not set or empty.
func (n NullableAdminContentID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableAdminContentID.
func (n NullableAdminContentID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableAdminContentID.
func (n *NullableAdminContentID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableAdminContentID.
func (n NullableAdminContentID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableAdminContentID.
func (n *NullableAdminContentID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableNodeID represents a nullable node ID
type NullableNodeID struct {
	ID    NodeID
	Valid bool
}

// Validate checks if the NullableNodeID is valid if set.
func (n NullableNodeID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableNodeID.
func (n NullableNodeID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableNodeID is not set or empty.
func (n NullableNodeID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableNodeID.
func (n NullableNodeID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableNodeID.
func (n *NullableNodeID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableNodeID.
func (n NullableNodeID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableNodeID.
func (n *NullableNodeID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableRoleID represents a nullable foreign key to roles
type NullableRoleID struct {
	ID    RoleID
	Valid bool
}

// Validate checks if the NullableRoleID is valid if set.
func (n NullableRoleID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableRoleID.
func (n NullableRoleID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableRoleID is not set or empty.
func (n NullableRoleID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableRoleID.
func (n NullableRoleID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableRoleID.
func (n *NullableRoleID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableRoleID.
func (n NullableRoleID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableRoleID.
func (n *NullableRoleID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableAdminDatatypeID represents a nullable foreign key to admin datatypes
type NullableAdminDatatypeID struct {
	ID    AdminDatatypeID
	Valid bool
}

// Validate checks if the NullableAdminDatatypeID is valid if set.
func (n NullableAdminDatatypeID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableAdminDatatypeID.
func (n NullableAdminDatatypeID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableAdminDatatypeID is not set or empty.
func (n NullableAdminDatatypeID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableAdminDatatypeID.
func (n NullableAdminDatatypeID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableAdminDatatypeID.
func (n *NullableAdminDatatypeID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableAdminDatatypeID.
func (n NullableAdminDatatypeID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableAdminDatatypeID.
func (n *NullableAdminDatatypeID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}

// NullableAdminFieldID represents a nullable foreign key to admin fields
type NullableAdminFieldID struct {
	ID    AdminFieldID
	Valid bool
}

// Validate checks if the NullableAdminFieldID is valid if set.
func (n NullableAdminFieldID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

// String returns the string representation of the NullableAdminFieldID.
func (n NullableAdminFieldID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

// IsZero returns true if the NullableAdminFieldID is not set or empty.
func (n NullableAdminFieldID) IsZero() bool { return !n.Valid || n.ID == "" }

// Value returns the database driver value for the NullableAdminFieldID.
func (n NullableAdminFieldID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

// Scan scans a value from the database into the NullableAdminFieldID.
func (n *NullableAdminFieldID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

// MarshalJSON returns the JSON representation of the NullableAdminFieldID.
func (n NullableAdminFieldID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

// UnmarshalJSON parses JSON data into the NullableAdminFieldID.
func (n *NullableAdminFieldID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}
