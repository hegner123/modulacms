package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ContentStatus represents the publication status of content
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusArchived  ContentStatus = "archived"
	ContentStatusPending   ContentStatus = "pending"
)

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

func (s ContentStatus) String() string {
	return string(s)
}

func (s ContentStatus) Value() (driver.Value, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return string(s), nil
}

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

func (s ContentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

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

func (t FieldType) String() string {
	return string(t)
}

func (t FieldType) Value() (driver.Value, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return string(t), nil
}

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

func (t FieldType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

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

const (
	RouteTypeStatic   RouteType = "static"
	RouteTypeDynamic  RouteType = "dynamic"
	RouteTypeAPI      RouteType = "api"
	RouteTypeRedirect RouteType = "redirect"
)

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

func (r RouteType) String() string {
	return string(r)
}

func (r RouteType) Value() (driver.Value, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return string(r), nil
}

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

func (r RouteType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

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

const (
	OpInsert Operation = "INSERT"
	OpUpdate Operation = "UPDATE"
	OpDelete Operation = "DELETE"
)

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

func (o Operation) String() string {
	return string(o)
}

func (o Operation) Value() (driver.Value, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	return string(o), nil
}

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

func (o Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(o))
}

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

const (
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionPublish Action = "publish"
	ActionArchive Action = "archive"
)

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

func (a Action) String() string {
	return string(a)
}

func (a Action) Value() (driver.Value, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return string(a), nil
}

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

func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

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

const (
	ConflictLWW    ConflictPolicy = "lww"    // Last Write Wins (simple, possible data loss)
	ConflictManual ConflictPolicy = "manual" // Flag conflicts for human resolution
)

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

func (c ConflictPolicy) String() string {
	return string(c)
}

func (c ConflictPolicy) Value() (driver.Value, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return string(c), nil
}

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

func (c ConflictPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

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

const (
	BackupTypeFull         BackupType = "full"
	BackupTypeIncremental  BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
)

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

func (b BackupType) String() string { return string(b) }

func (b BackupType) Value() (driver.Value, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return string(b), nil
}

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

func (b BackupType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

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

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
)

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

func (b BackupStatus) String() string { return string(b) }

func (b BackupStatus) Value() (driver.Value, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return string(b), nil
}

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

func (b BackupStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

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

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusFailed   VerificationStatus = "failed"
)

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

func (v VerificationStatus) String() string { return string(v) }

func (v VerificationStatus) Value() (driver.Value, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return string(v), nil
}

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

func (v VerificationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(v))
}

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

const (
	BackupSetStatusPending  BackupSetStatus = "pending"
	BackupSetStatusComplete BackupSetStatus = "complete"
	BackupSetStatusPartial  BackupSetStatus = "partial"
)

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

func (b BackupSetStatus) String() string { return string(b) }

func (b BackupSetStatus) Value() (driver.Value, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return string(b), nil
}

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

func (b BackupSetStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

func (b *BackupSetStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("BackupSetStatus: %w", err)
	}
	*b = BackupSetStatus(str)
	return b.Validate()
}
