package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ContentStatus represents the publication status of content
type ContentStatus string

// Valid ContentStatus values.
const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusArchived  ContentStatus = "archived"
	ContentStatusPending   ContentStatus = "pending"
)

// Validate checks that the ContentStatus is one of the allowed values.
func (s ContentStatus) Validate() error {
	switch s {
	case ContentStatusDraft, ContentStatusPublished, ContentStatusArchived, ContentStatusPending:
		return nil
	case "":
		return fmt.Errorf("ContentStatus: cannot be empty")
	default:
		return fmt.Errorf("ContentStatus: invalid value %q (valid: draft, published, archived, pending)", s)
	}
}

// String returns the string representation of ContentStatus.
func (s ContentStatus) String() string {
	return string(s)
}

// Value returns the database driver value for ContentStatus.
func (s ContentStatus) Value() (driver.Value, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return string(s), nil
}

// Scan reads a ContentStatus from a database value.
func (s *ContentStatus) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("ContentStatus: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*s = ContentStatus(v)
	case []byte:
		*s = ContentStatus(string(v))
	default:
		return fmt.Errorf("ContentStatus: cannot scan %T", value)
	}
	return s.Validate()
}

// MarshalJSON marshals ContentStatus to JSON.
func (s ContentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON unmarshals ContentStatus from JSON.
func (s *ContentStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("ContentStatus: %w", err)
	}
	*s = ContentStatus(str)
	return s.Validate()
}

// FieldType represents the type of a content field
type FieldType string

// Valid FieldType values.
const (
	FieldTypeText     FieldType = "text"
	FieldTypeTextarea FieldType = "textarea"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeDatetime FieldType = "datetime"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeSelect   FieldType = "select"
	FieldTypeMedia    FieldType = "media"
	FieldTypeRelation FieldType = "relation"
	FieldTypeJSON     FieldType = "json"
	FieldTypeRichText FieldType = "richtext"
	FieldTypeSlug     FieldType = "slug"
	FieldTypeEmail    FieldType = "email"
	FieldTypeURL      FieldType = "url"
)

// Validate checks that the FieldType is one of the allowed values.
func (t FieldType) Validate() error {
	switch t {
	case FieldTypeText, FieldTypeTextarea, FieldTypeNumber, FieldTypeDate,
		FieldTypeDatetime, FieldTypeBoolean, FieldTypeSelect, FieldTypeMedia,
		FieldTypeRelation, FieldTypeJSON, FieldTypeRichText, FieldTypeSlug,
		FieldTypeEmail, FieldTypeURL:
		return nil
	case "":
		return fmt.Errorf("FieldType: cannot be empty")
	default:
		return fmt.Errorf("FieldType: invalid value %q (valid: text, textarea, number, date, datetime, boolean, select, media, relation, json, richtext, slug, email, url)", t)
	}
}

// String returns the string representation of FieldType.
func (t FieldType) String() string {
	return string(t)
}

// Value returns the database driver value for FieldType.
func (t FieldType) Value() (driver.Value, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return string(t), nil
}

// Scan reads a FieldType from a database value.
func (t *FieldType) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("FieldType: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*t = FieldType(v)
	case []byte:
		*t = FieldType(string(v))
	default:
		return fmt.Errorf("FieldType: cannot scan %T", value)
	}
	return t.Validate()
}

// MarshalJSON marshals FieldType to JSON.
func (t FieldType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

// UnmarshalJSON unmarshals FieldType from JSON.
func (t *FieldType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("FieldType: %w", err)
	}
	*t = FieldType(str)
	return t.Validate()
}

// RouteType represents the type of a route
type RouteType string

// Valid RouteType values.
const (
	RouteTypeStatic   RouteType = "static"
	RouteTypeDynamic  RouteType = "dynamic"
	RouteTypeAPI      RouteType = "api"
	RouteTypeRedirect RouteType = "redirect"
)

// Validate checks that the RouteType is one of the allowed values.
func (r RouteType) Validate() error {
	switch r {
	case RouteTypeStatic, RouteTypeDynamic, RouteTypeAPI, RouteTypeRedirect:
		return nil
	case "":
		return fmt.Errorf("RouteType: cannot be empty")
	default:
		return fmt.Errorf("RouteType: invalid value %q (valid: static, dynamic, api, redirect)", r)
	}
}

// String returns the string representation of RouteType.
func (r RouteType) String() string {
	return string(r)
}

// Value returns the database driver value for RouteType.
func (r RouteType) Value() (driver.Value, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return string(r), nil
}

// Scan reads a RouteType from a database value.
func (r *RouteType) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("RouteType: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*r = RouteType(v)
	case []byte:
		*r = RouteType(string(v))
	default:
		return fmt.Errorf("RouteType: cannot scan %T", value)
	}
	return r.Validate()
}

// MarshalJSON marshals RouteType to JSON.
func (r RouteType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

// UnmarshalJSON unmarshals RouteType from JSON.
func (r *RouteType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("RouteType: %w", err)
	}
	*r = RouteType(str)
	return r.Validate()
}

// Operation represents database operations (for change_events)
type Operation string

// Valid Operation values.
const (
	OpInsert Operation = "INSERT"
	OpUpdate Operation = "UPDATE"
	OpDelete Operation = "DELETE"
)

// Validate checks that the Operation is one of the allowed values.
func (o Operation) Validate() error {
	switch o {
	case OpInsert, OpUpdate, OpDelete:
		return nil
	case "":
		return fmt.Errorf("Operation: cannot be empty")
	default:
		return fmt.Errorf("Operation: invalid value %q (valid: INSERT, UPDATE, DELETE)", o)
	}
}

// String returns the string representation of Operation.
func (o Operation) String() string {
	return string(o)
}

// Value returns the database driver value for Operation.
func (o Operation) Value() (driver.Value, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	return string(o), nil
}

// Scan reads an Operation from a database value.
func (o *Operation) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("Operation: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*o = Operation(v)
	case []byte:
		*o = Operation(string(v))
	default:
		return fmt.Errorf("Operation: cannot scan %T", value)
	}
	return o.Validate()
}

// MarshalJSON marshals Operation to JSON.
func (o Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(o))
}

// UnmarshalJSON unmarshals Operation from JSON.
func (o *Operation) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Operation: %w", err)
	}
	*o = Operation(str)
	return o.Validate()
}

// Action represents business-level actions (more specific than Operation)
type Action string

// Valid Action values.
const (
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionPublish Action = "publish"
	ActionArchive Action = "archive"
)

// Validate checks that the Action is one of the allowed values.
func (a Action) Validate() error {
	switch a {
	case ActionCreate, ActionUpdate, ActionDelete, ActionPublish, ActionArchive:
		return nil
	case "":
		return fmt.Errorf("Action: cannot be empty")
	default:
		return fmt.Errorf("Action: invalid value %q (valid: create, update, delete, publish, archive)", a)
	}
}

// String returns the string representation of Action.
func (a Action) String() string {
	return string(a)
}

// Value returns the database driver value for Action.
func (a Action) Value() (driver.Value, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return string(a), nil
}

// Scan reads an Action from a database value.
func (a *Action) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("Action: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*a = Action(v)
	case []byte:
		*a = Action(string(v))
	default:
		return fmt.Errorf("Action: cannot scan %T", value)
	}
	return a.Validate()
}

// MarshalJSON marshals Action to JSON.
func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

// UnmarshalJSON unmarshals Action from JSON.
func (a *Action) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Action: %w", err)
	}
	*a = Action(str)
	return a.Validate()
}

// ConflictPolicy defines how conflicts are resolved for a datatype (for distributed conflict resolution)
type ConflictPolicy string

// Valid ConflictPolicy values.
const (
	ConflictLWW    ConflictPolicy = "lww"    // Last Write Wins (simple, possible data loss)
	ConflictManual ConflictPolicy = "manual" // Flag conflicts for human resolution
)

// Validate checks that the ConflictPolicy is one of the allowed values.
func (c ConflictPolicy) Validate() error {
	switch c {
	case ConflictLWW, ConflictManual:
		return nil
	case "":
		return fmt.Errorf("ConflictPolicy: cannot be empty")
	default:
		return fmt.Errorf("ConflictPolicy: invalid value %q (valid: lww, manual)", c)
	}
}

// String returns the string representation of ConflictPolicy.
func (c ConflictPolicy) String() string {
	return string(c)
}

// Value returns the database driver value for ConflictPolicy.
func (c ConflictPolicy) Value() (driver.Value, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return string(c), nil
}

// Scan reads a ConflictPolicy from a database value.
func (c *ConflictPolicy) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("ConflictPolicy: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*c = ConflictPolicy(v)
	case []byte:
		*c = ConflictPolicy(string(v))
	default:
		return fmt.Errorf("ConflictPolicy: cannot scan %T", value)
	}
	return c.Validate()
}

// MarshalJSON marshals ConflictPolicy to JSON.
func (c ConflictPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

// UnmarshalJSON unmarshals ConflictPolicy from JSON.
func (c *ConflictPolicy) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("ConflictPolicy: %w", err)
	}
	*c = ConflictPolicy(str)
	return c.Validate()
}

// ========================================
// BACKUP SYSTEM ENUMS
// ========================================

// BackupType represents the type of backup
type BackupType string

// Valid BackupType values.
const (
	BackupTypeFull         BackupType = "full"
	BackupTypeIncremental  BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
)

// Validate checks that the BackupType is one of the allowed values.
func (b BackupType) Validate() error {
	switch b {
	case BackupTypeFull, BackupTypeIncremental, BackupTypeDifferential:
		return nil
	case "":
		return fmt.Errorf("BackupType: cannot be empty")
	default:
		return fmt.Errorf("BackupType: invalid value %q (valid: full, incremental, differential)", b)
	}
}

// String returns the string representation of BackupType.
func (b BackupType) String() string { return string(b) }

// Value returns the database driver value for BackupType.
func (b BackupType) Value() (driver.Value, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan reads a BackupType from a database value.
func (b *BackupType) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("BackupType: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*b = BackupType(v)
	case []byte:
		*b = BackupType(string(v))
	default:
		return fmt.Errorf("BackupType: cannot scan %T", value)
	}
	return b.Validate()
}

// MarshalJSON marshals BackupType to JSON.
func (b BackupType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

// UnmarshalJSON unmarshals BackupType from JSON.
func (b *BackupType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("BackupType: %w", err)
	}
	*b = BackupType(str)
	return b.Validate()
}

// BackupStatus represents the status of a backup operation
type BackupStatus string

// Valid BackupStatus values.
const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
)

// Validate checks that the BackupStatus is one of the allowed values.
func (b BackupStatus) Validate() error {
	switch b {
	case BackupStatusPending, BackupStatusInProgress, BackupStatusCompleted, BackupStatusFailed:
		return nil
	case "":
		return fmt.Errorf("BackupStatus: cannot be empty")
	default:
		return fmt.Errorf("BackupStatus: invalid value %q (valid: pending, in_progress, completed, failed)", b)
	}
}

// String returns the string representation of BackupStatus.
func (b BackupStatus) String() string { return string(b) }

// Value returns the database driver value for BackupStatus.
func (b BackupStatus) Value() (driver.Value, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan reads a BackupStatus from a database value.
func (b *BackupStatus) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("BackupStatus: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*b = BackupStatus(v)
	case []byte:
		*b = BackupStatus(string(v))
	default:
		return fmt.Errorf("BackupStatus: cannot scan %T", value)
	}
	return b.Validate()
}

// MarshalJSON marshals BackupStatus to JSON.
func (b BackupStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

// UnmarshalJSON unmarshals BackupStatus from JSON.
func (b *BackupStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("BackupStatus: %w", err)
	}
	*b = BackupStatus(str)
	return b.Validate()
}

// VerificationStatus represents the status of a backup verification
type VerificationStatus string

// Valid VerificationStatus values.
const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusFailed   VerificationStatus = "failed"
)

// Validate checks that the VerificationStatus is one of the allowed values.
func (v VerificationStatus) Validate() error {
	switch v {
	case VerificationStatusPending, VerificationStatusVerified, VerificationStatusFailed:
		return nil
	case "":
		return fmt.Errorf("VerificationStatus: cannot be empty")
	default:
		return fmt.Errorf("VerificationStatus: invalid value %q (valid: pending, verified, failed)", v)
	}
}

// String returns the string representation of VerificationStatus.
func (v VerificationStatus) String() string { return string(v) }

// Value returns the database driver value for VerificationStatus.
func (v VerificationStatus) Value() (driver.Value, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan reads a VerificationStatus from a database value.
func (vs *VerificationStatus) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("VerificationStatus: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*vs = VerificationStatus(v)
	case []byte:
		*vs = VerificationStatus(string(v))
	default:
		return fmt.Errorf("VerificationStatus: cannot scan %T", value)
	}
	return vs.Validate()
}

// MarshalJSON marshals VerificationStatus to JSON.
func (v VerificationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(v))
}

// UnmarshalJSON unmarshals VerificationStatus from JSON.
func (vs *VerificationStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("VerificationStatus: %w", err)
	}
	*vs = VerificationStatus(str)
	return vs.Validate()
}

// BackupSetStatus represents the status of a backup set (collection of backups)
type BackupSetStatus string

// Valid BackupSetStatus values.
const (
	BackupSetStatusPending  BackupSetStatus = "pending"
	BackupSetStatusComplete BackupSetStatus = "complete"
	BackupSetStatusPartial  BackupSetStatus = "partial"
)

// Validate checks that the BackupSetStatus is one of the allowed values.
func (b BackupSetStatus) Validate() error {
	switch b {
	case BackupSetStatusPending, BackupSetStatusComplete, BackupSetStatusPartial:
		return nil
	case "":
		return fmt.Errorf("BackupSetStatus: cannot be empty")
	default:
		return fmt.Errorf("BackupSetStatus: invalid value %q (valid: pending, complete, partial)", b)
	}
}

// String returns the string representation of BackupSetStatus.
func (b BackupSetStatus) String() string { return string(b) }

// Value returns the database driver value for BackupSetStatus.
func (b BackupSetStatus) Value() (driver.Value, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan reads a BackupSetStatus from a database value.
func (b *BackupSetStatus) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("BackupSetStatus: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*b = BackupSetStatus(v)
	case []byte:
		*b = BackupSetStatus(string(v))
	default:
		return fmt.Errorf("BackupSetStatus: cannot scan %T", value)
	}
	return b.Validate()
}

// MarshalJSON marshals BackupSetStatus to JSON.
func (b BackupSetStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

// UnmarshalJSON unmarshals BackupSetStatus from JSON.
func (b *BackupSetStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("BackupSetStatus: %w", err)
	}
	*b = BackupSetStatus(str)
	return b.Validate()
}
