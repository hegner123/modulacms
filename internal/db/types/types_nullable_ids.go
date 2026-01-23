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

func (n NullableDatatypeID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableDatatypeID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableDatatypeID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableDatatypeID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableDatatypeID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableDatatypeID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableUserID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableUserID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableUserID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableUserID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableUserID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableUserID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableContentID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableContentID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableContentID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableContentID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableContentID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableContentID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableFieldID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableFieldID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableFieldID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableFieldID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableFieldID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableFieldID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableMediaID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableMediaID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableMediaID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableMediaID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableMediaID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableMediaID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableRouteID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableRouteID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableRouteID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableRouteID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableRouteID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableRouteID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableAdminRouteID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableAdminRouteID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableAdminRouteID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableAdminRouteID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableAdminRouteID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableAdminRouteID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableAdminContentID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableAdminContentID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableAdminContentID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableAdminContentID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableAdminContentID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableAdminContentID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableNodeID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableNodeID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableNodeID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableNodeID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableNodeID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableNodeID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

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

func (n NullableRoleID) Validate() error {
	if n.Valid {
		return n.ID.Validate()
	}
	return nil
}

func (n NullableRoleID) String() string {
	if !n.Valid {
		return "null"
	}
	return n.ID.String()
}

func (n NullableRoleID) IsZero() bool { return !n.Valid || n.ID == "" }

func (n NullableRoleID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.ID), nil
}

func (n *NullableRoleID) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return n.ID.Scan(value)
}

func (n NullableRoleID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID)
}

func (n *NullableRoleID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.ID = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.ID)
}
