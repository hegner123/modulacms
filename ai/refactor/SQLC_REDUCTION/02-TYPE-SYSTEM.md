# Custom Type System

This document defines all custom Go types that will be shared across all three database engines (SQLite, MySQL, PostgreSQL) via sqlc type overrides.

---

## Primary ID Types (ULID-based)

Using [github.com/oklog/ulid/v2](https://github.com/oklog/ulid) for ULID generation.

```go
// internal/db/types_ids.go
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

// DatatypeID uniquely identifies a datatype (26-char ULID string)
type DatatypeID string

// NewDatatypeID generates a new DatatypeID
func NewDatatypeID() DatatypeID {
    return DatatypeID(NewULID().String())
}

func (id DatatypeID) Validate() error {
    if id == "" {
        return fmt.Errorf("DatatypeID: cannot be empty")
    }
    if len(id) != 26 {
        return fmt.Errorf("DatatypeID: invalid length %d (expected 26)", len(id))
    }
    // Validate ULID format
    _, err := ulid.Parse(string(id))
    if err != nil {
        return fmt.Errorf("DatatypeID: invalid ULID format %q: %w", id, err)
    }
    return nil
}

func (id DatatypeID) String() string {
    return string(id)
}

// ParseDatatypeID parses a string into a DatatypeID (validates format)
func ParseDatatypeID(s string) (DatatypeID, error) {
    id := DatatypeID(s)
    if err := id.Validate(); err != nil {
        return "", err
    }
    return id, nil
}

// ULID returns the underlying ulid.ULID (for time extraction, comparison)
func (id DatatypeID) ULID() (ulid.ULID, error) {
    return ulid.Parse(string(id))
}

// Time returns the timestamp encoded in the ULID
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

func (id DatatypeID) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(id))
}

func (id *DatatypeID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return fmt.Errorf("DatatypeID: %w", err)
    }
    *id = DatatypeID(s)
    return id.Validate()
}

// IsZero returns true if the ID is empty
func (id DatatypeID) IsZero() bool {
    return id == ""
}

// Repeat pattern for all ID types:
// - UserID, RoleID, PermissionID, FieldID
// - ContentID, ContentFieldID, MediaID, MediaDimensionID
// - SessionID, TokenID, RouteID, AdminRouteID, TableID
// - UserOauthID, UserSshKeyID
// - AdminDatatypeID, AdminFieldID, AdminContentID, AdminContentFieldID
// - DatatypeFieldID, AdminDatatypeFieldID
// - EventID (for change_events)
// - NodeID (for multi-node support)
```

---

## Nullable ID Types (ULID-based)

```go
// internal/db/types_nullable_ids.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
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
        return "NullableDatatypeID(null)"
    }
    return fmt.Sprintf("NullableDatatypeID(%s)", n.ID)
}

func (n NullableDatatypeID) Value() (driver.Value, error) {
    if !n.Valid {
        return nil, nil
    }
    return string(n.ID), nil  // Store as string
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

// IsZero returns true if null or empty
func (n NullableDatatypeID) IsZero() bool {
    return !n.Valid || n.ID == ""
}

// Repeat pattern for all nullable ID types:
// - NullableUserID
// - NullableRoleID
// - NullableContentID
// - NullableFieldID
// - NullableMediaID
// - NullableDatatypeID
// etc.
```

---

## Unified Timestamp

**Standardization rule:** All timestamps stored as UTC. One input format accepted (RFC3339 with timezone).

```go
// internal/db/types_timestamp.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
)

// Strict formats for API input - only RFC3339 with timezone
var strictTimestampFormats = []string{
    time.RFC3339,     // "2006-01-02T15:04:05Z07:00" - PRIMARY FORMAT
    time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
}

// Legacy formats for database reads only (historical data compatibility)
// These are NOT accepted for API input
var legacyTimestampFormats = []string{
    "2006-01-02 15:04:05",     // MySQL without TZ (assume UTC)
    "2006-01-02T15:04:05Z",    // UTC shorthand
    "2006-01-02T15:04:05",     // No TZ (assume UTC)
    "2006-01-02",              // Date only (assume 00:00:00 UTC)
}

// Timestamp handles datetime columns across SQLite (TEXT), MySQL (DATETIME), PostgreSQL (TIMESTAMP)
// All times are stored and returned in UTC.
type Timestamp struct {
    Time  time.Time
    Valid bool
}

func NewTimestamp(t time.Time) Timestamp {
    return Timestamp{Time: t.UTC(), Valid: true}
}

func TimestampNow() Timestamp {
    return Timestamp{Time: time.Now().UTC(), Valid: true}
}

func (t Timestamp) String() string {
    if !t.Valid {
        return "Timestamp(null)"
    }
    return fmt.Sprintf("Timestamp(%s)", t.Time.UTC().Format(time.RFC3339))
}

func (t Timestamp) Value() (driver.Value, error) {
    if !t.Valid {
        return nil, nil
    }
    return t.Time.UTC(), nil  // Always store as UTC
}

// Scan reads from database - accepts legacy formats for compatibility
func (t *Timestamp) Scan(value any) error {
    if value == nil {
        t.Valid = false
        return nil
    }
    switch v := value.(type) {
    case time.Time:
        t.Time, t.Valid = v.UTC(), true
        return nil
    case string:
        if v == "" {
            t.Valid = false
            return nil
        }
        // Try strict formats first
        for _, format := range strictTimestampFormats {
            if parsed, err := time.Parse(format, v); err == nil {
                t.Time, t.Valid = parsed.UTC(), true
                return nil
            }
        }
        // Fall back to legacy formats for database reads
        for _, format := range legacyTimestampFormats {
            if parsed, err := time.Parse(format, v); err == nil {
                t.Time, t.Valid = parsed.UTC(), true
                return nil
            }
        }
        return fmt.Errorf("Timestamp: cannot parse %q", v)
    case []byte:
        return t.Scan(string(v))
    default:
        return fmt.Errorf("Timestamp: cannot scan %T", value)
    }
}

// MarshalJSON always outputs RFC3339 in UTC
func (t Timestamp) MarshalJSON() ([]byte, error) {
    if !t.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(t.Time.UTC().Format(time.RFC3339))
}

// UnmarshalJSON ONLY accepts RFC3339 format with timezone - strict API input validation
func (t *Timestamp) UnmarshalJSON(data []byte) error {
    if string(data) == "null" {
        t.Valid = false
        return nil
    }
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return fmt.Errorf("Timestamp: expected string, got %s", string(data))
    }

    // STRICT: Only accept RFC3339 formats for API input
    for _, format := range strictTimestampFormats {
        if parsed, err := time.Parse(format, s); err == nil {
            t.Time, t.Valid = parsed.UTC(), true
            return nil
        }
    }

    // Reject legacy formats for API input
    return fmt.Errorf("Timestamp: invalid format %q (must be RFC3339: 2006-01-02T15:04:05Z or 2006-01-02T15:04:05-07:00)", s)
}

// IsZero returns true if timestamp is null or zero time
func (t Timestamp) IsZero() bool {
    return !t.Valid || t.Time.IsZero()
}

// Before reports whether t is before u
func (t Timestamp) Before(u Timestamp) bool {
    if !t.Valid || !u.Valid {
        return false
    }
    return t.Time.Before(u.Time)
}

// After reports whether t is after u
func (t Timestamp) After(u Timestamp) bool {
    if !t.Valid || !u.Valid {
        return false
    }
    return t.Time.After(u.Time)
}

// UTC returns the time in UTC (already stored as UTC, but explicit for clarity)
func (t Timestamp) UTC() time.Time {
    return t.Time.UTC()
}
```

---

## Hybrid Logical Clock (HLC)

For distributed ordering of events across nodes with clock skew.

```go
// internal/db/types_hlc.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

// HLC represents a Hybrid Logical Clock timestamp
// Format: (wall_time_ms << 16) | logical_counter
// - Upper 48 bits: milliseconds since Unix epoch
// - Lower 16 bits: logical counter (for ordering within same millisecond)
type HLC int64

var (
    hlcMu      sync.Mutex
    hlcLast    HLC
    hlcCounter uint16
)

// Now returns a new HLC timestamp greater than any previously issued
func HLCNow() HLC {
    hlcMu.Lock()
    defer hlcMu.Unlock()

    wallMs := time.Now().UnixMilli()
    physical := HLC(wallMs << 16)

    if physical > hlcLast {
        hlcLast = physical
        hlcCounter = 0
    } else {
        hlcCounter++
        if hlcCounter == 0 {
            // Counter overflow, wait for wall clock to advance
            time.Sleep(time.Millisecond)
            return HLCNow()
        }
    }

    hlcLast = physical | HLC(hlcCounter)
    return hlcLast
}

// Update merges a received HLC with local time (for receiving events from other nodes)
func HLCUpdate(received HLC) HLC {
    hlcMu.Lock()
    defer hlcMu.Unlock()

    wallMs := time.Now().UnixMilli()
    physical := HLC(wallMs << 16)

    maxPhysical := physical
    if received > maxPhysical {
        maxPhysical = received & ^HLC(0xFFFF) // Extract physical part
    }
    if hlcLast > maxPhysical {
        maxPhysical = hlcLast & ^HLC(0xFFFF)
    }

    if maxPhysical == hlcLast&^HLC(0xFFFF) {
        hlcCounter++
    } else {
        hlcCounter = 0
    }

    hlcLast = maxPhysical | HLC(hlcCounter)
    return hlcLast
}

func (h HLC) Physical() time.Time {
    ms := int64(h >> 16)
    return time.UnixMilli(ms)
}

func (h HLC) Logical() uint16 {
    return uint16(h & 0xFFFF)
}

func (h HLC) String() string {
    return fmt.Sprintf("HLC(%d:%d)", h>>16, h&0xFFFF)
}

func (h HLC) Value() (driver.Value, error) {
    return int64(h), nil
}

func (h *HLC) Scan(value any) error {
    if value == nil {
        *h = 0
        return nil
    }
    switch v := value.(type) {
    case int64:
        *h = HLC(v)
    case int:
        *h = HLC(v)
    default:
        return fmt.Errorf("HLC: cannot scan %T", value)
    }
    return nil
}

func (h HLC) MarshalJSON() ([]byte, error) {
    return json.Marshal(int64(h))
}

func (h *HLC) UnmarshalJSON(data []byte) error {
    var v int64
    if err := json.Unmarshal(data, &v); err != nil {
        return fmt.Errorf("HLC: %w", err)
    }
    *h = HLC(v)
    return nil
}

// Before returns true if h happened before other
func (h HLC) Before(other HLC) bool {
    return h < other
}

// After returns true if h happened after other
func (h HLC) After(other HLC) bool {
    return h > other
}
```

---

## Domain Enums

```go
// internal/db/types_enums.go
package db

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
        return fmt.Errorf("ContentStatus: invalid value %q", s)
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
        return fmt.Errorf("FieldType: invalid value %q", t)
    }
}

// Similar implementation for Scan, Value, MarshalJSON, UnmarshalJSON...

// RouteType represents the type of a route
type RouteType string

const (
    RouteTypeStatic  RouteType = "static"
    RouteTypeDynamic RouteType = "dynamic"
    RouteTypeAPI     RouteType = "api"
    RouteTypeRedirect RouteType = "redirect"
)

func (t RouteType) Validate() error {
    switch t {
    case RouteTypeStatic, RouteTypeDynamic, RouteTypeAPI, RouteTypeRedirect:
        return nil
    case "":
        return fmt.Errorf("RouteType: cannot be empty")
    default:
        return fmt.Errorf("RouteType: invalid value %q", t)
    }
}

// Similar implementation for Scan, Value, MarshalJSON, UnmarshalJSON...

// Operation represents database operations
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
    default:
        return fmt.Errorf("Operation: invalid value %q", o)
    }
}

// Action represents business-level actions (optional, more specific than Operation)
type Action string

const (
    ActionCreate  Action = "create"
    ActionUpdate  Action = "update"
    ActionDelete  Action = "delete"
    ActionPublish Action = "publish"
    ActionArchive Action = "archive"
)

// ConflictPolicy defines how conflicts are resolved for a datatype
type ConflictPolicy string

const (
    ConflictLWW    ConflictPolicy = "lww"    // Last Write Wins (simple, possible data loss)
    ConflictManual ConflictPolicy = "manual" // Flag conflicts for human resolution
)

func (c ConflictPolicy) Validate() error {
    switch c {
    case ConflictLWW, ConflictManual:
        return nil
    default:
        return fmt.Errorf("ConflictPolicy: invalid value %q", c)
    }
}
```

---

## Validation Types

```go
// internal/db/types_validation.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "net/url"
    "regexp"
    "strings"
    "unicode"
)

// Slug represents a URL-safe identifier (lowercase, hyphens, no spaces)
type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s Slug) Validate() error {
    if s == "" {
        return fmt.Errorf("Slug: cannot be empty")
    }
    if len(s) > 255 {
        return fmt.Errorf("Slug: too long (max 255 chars)")
    }
    if !slugRegex.MatchString(string(s)) {
        return fmt.Errorf("Slug: invalid format %q (must be lowercase alphanumeric with hyphens)", s)
    }
    return nil
}

func (s Slug) String() string {
    return string(s)
}

// Slugify converts a string to a valid slug
func Slugify(input string) Slug {
    // Lowercase
    result := strings.ToLower(input)
    // Replace spaces and underscores with hyphens
    result = strings.ReplaceAll(result, " ", "-")
    result = strings.ReplaceAll(result, "_", "-")
    // Remove non-alphanumeric except hyphens
    var sb strings.Builder
    for _, r := range result {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
            sb.WriteRune(r)
        }
    }
    result = sb.String()
    // Collapse multiple hyphens
    for strings.Contains(result, "--") {
        result = strings.ReplaceAll(result, "--", "-")
    }
    // Trim hyphens from ends
    result = strings.Trim(result, "-")
    return Slug(result)
}

func (s Slug) Value() (driver.Value, error) {
    return string(s), nil
}

func (s *Slug) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("Slug: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *s = Slug(v)
    case []byte:
        *s = Slug(string(v))
    default:
        return fmt.Errorf("Slug: cannot scan %T", value)
    }
    return s.Validate()
}

func (s Slug) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(s))
}

func (s *Slug) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return fmt.Errorf("Slug: %w", err)
    }
    *s = Slug(str)
    return s.Validate()
}

// Email represents a validated email address
type Email string

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (e Email) Validate() error {
    if e == "" {
        return fmt.Errorf("Email: cannot be empty")
    }
    if len(e) > 254 {
        return fmt.Errorf("Email: too long (max 254 chars)")
    }
    if !emailRegex.MatchString(string(e)) {
        return fmt.Errorf("Email: invalid format %q", e)
    }
    return nil
}

func (e Email) String() string {
    return string(e)
}

func (e Email) Domain() string {
    parts := strings.Split(string(e), "@")
    if len(parts) != 2 {
        return ""
    }
    return parts[1]
}

func (e Email) Value() (driver.Value, error) {
    return string(e), nil
}

func (e *Email) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("Email: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *e = Email(v)
    case []byte:
        *e = Email(string(v))
    default:
        return fmt.Errorf("Email: cannot scan %T", value)
    }
    return e.Validate()
}

func (e Email) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(e))
}

func (e *Email) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return fmt.Errorf("Email: %w", err)
    }
    *e = Email(str)
    return e.Validate()
}

// URL represents a validated URL
type URL string

func (u URL) Validate() error {
    if u == "" {
        return fmt.Errorf("URL: cannot be empty")
    }
    parsed, err := url.Parse(string(u))
    if err != nil {
        return fmt.Errorf("URL: invalid format %q: %w", u, err)
    }
    if parsed.Scheme == "" {
        return fmt.Errorf("URL: missing scheme in %q", u)
    }
    if parsed.Host == "" {
        return fmt.Errorf("URL: missing host in %q", u)
    }
    return nil
}

func (u URL) String() string {
    return string(u)
}

func (u URL) Parse() (*url.URL, error) {
    return url.Parse(string(u))
}

func (u URL) Value() (driver.Value, error) {
    return string(u), nil
}

func (u *URL) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("URL: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *u = URL(v)
    case []byte:
        *u = URL(string(v))
    default:
        return fmt.Errorf("URL: cannot scan %T", value)
    }
    return u.Validate()
}

func (u URL) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(u))
}

func (u *URL) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return fmt.Errorf("URL: %w", err)
    }
    *u = URL(str)
    return u.Validate()
}

// NullableSlug, NullableEmail, NullableURL for optional fields
type NullableSlug struct {
    Slug  Slug
    Valid bool
}

type NullableEmail struct {
    Email Email
    Valid bool
}

type NullableURL struct {
    URL   URL
    Valid bool
}

// Similar Scan, Value, MarshalJSON, UnmarshalJSON implementations...
```

---

## Change Events (Audit + Replication + Webhooks)

**Note:** The `change_events` table replaces JSON `history` columns and serves as:
1. **Audit trail** - Who changed what, when
2. **Replication log** - What to sync to other nodes
3. **Webhook source** - What events to fire

```go
// internal/db/types_change_events.go
package db

import (
    "context"
)

// EventID uniquely identifies a change event (ULID)
type EventID string

func NewEventID() EventID {
    return EventID(NewULID().String())
}

// (Same methods as other ULID-based ID types)

// NodeID identifies a node in a distributed deployment (ULID)
type NodeID string

func NewNodeID() NodeID {
    return NodeID(NewULID().String())
}

// (Same methods as other ULID-based ID types)

// ChangeEvent represents a row in change_events table
type ChangeEvent struct {
    EventID       EventID        `json:"event_id"`
    HLCTimestamp  HLC            `json:"hlc_timestamp"`
    WallTimestamp Timestamp      `json:"wall_timestamp"`
    NodeID        NodeID         `json:"node_id"`
    TableName     string         `json:"table_name"`
    RecordID      string         `json:"record_id"`  // ULID of affected record
    Operation     Operation      `json:"operation"`
    Action        Action         `json:"action,omitempty"`
    UserID        NullableUserID `json:"user_id,omitempty"`
    OldValues     any            `json:"old_values,omitempty"`
    NewValues     any            `json:"new_values,omitempty"`
    Metadata      any            `json:"metadata,omitempty"`
    SyncedAt      Timestamp      `json:"synced_at,omitempty"`
    ConsumedAt    Timestamp      `json:"consumed_at,omitempty"`
}

// EventLogger interface for change event operations
type EventLogger interface {
    // LogEvent records a change event
    LogEvent(ctx context.Context, event ChangeEvent) error

    // GetEventsByRecord retrieves events for a specific record
    GetEventsByRecord(ctx context.Context, tableName string, recordID string) ([]ChangeEvent, error)

    // GetEventsSince retrieves events after an HLC timestamp (for replication)
    GetEventsSince(ctx context.Context, hlc HLC, limit int) ([]ChangeEvent, error)

    // GetUnsyncedEvents retrieves events not yet synced to other nodes
    GetUnsyncedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

    // GetUnconsumedEvents retrieves events not yet processed by webhooks
    GetUnconsumedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

    // MarkSynced marks events as synced
    MarkSynced(ctx context.Context, eventIDs []EventID) error

    // MarkConsumed marks events as consumed by webhooks
    MarkConsumed(ctx context.Context, eventIDs []EventID) error
}

// NewChangeEvent creates a change event for the current node
func NewChangeEvent(nodeID NodeID, tableName string, recordID string, op Operation, action Action, userID UserID) ChangeEvent {
    return ChangeEvent{
        EventID:       NewEventID(),
        HLCTimestamp:  HLCNow(),
        WallTimestamp: TimestampNow(),
        NodeID:        nodeID,
        TableName:     tableName,
        RecordID:      recordID,
        Operation:     op,
        Action:        action,
        UserID:        NullableUserID{ID: userID, Valid: userID != ""},
    }
}

// WithChanges adds old/new values to the event
func (e ChangeEvent) WithChanges(oldVal, newVal any) ChangeEvent {
    e.OldValues = oldVal
    e.NewValues = newVal
    return e
}

// WithMetadata adds metadata to the event
func (e ChangeEvent) WithMetadata(meta any) ChangeEvent {
    e.Metadata = meta
    return e
}
```

### SQLC Queries for change_events

```sql
-- name: LogEvent :exec
INSERT INTO change_events (event_id, hlc_timestamp, node_id, table_name, record_id, operation, action, user_id, old_values, new_values, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetEventsByRecord :many
SELECT * FROM change_events
WHERE table_name = $1 AND record_id = $2
ORDER BY hlc_timestamp DESC;

-- name: GetEventsSince :many
SELECT * FROM change_events
WHERE hlc_timestamp > $1
ORDER BY hlc_timestamp ASC
LIMIT $2;

-- name: GetUnsyncedEvents :many
SELECT * FROM change_events
WHERE synced_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT $1;

-- name: GetUnconsumedEvents :many
SELECT * FROM change_events
WHERE consumed_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT $1;

-- name: MarkSynced :exec
UPDATE change_events SET synced_at = CURRENT_TIMESTAMP WHERE event_id = ANY($1);

-- name: MarkConsumed :exec
UPDATE change_events SET consumed_at = CURRENT_TIMESTAMP WHERE event_id = ANY($1);
```

---

## Transaction Helper

```go
// internal/db/transaction.go
package db

import (
    "context"
    "database/sql"
    "fmt"
)

// TxFunc executes within a transaction
type TxFunc func(tx *sql.Tx) error

// WithTransaction executes fn within a transaction with automatic commit/rollback
func WithTransaction(ctx context.Context, db *sql.DB, fn TxFunc) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback() // no-op if already committed

    if err := fn(tx); err != nil {
        return err
    }
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}

// WithTransactionResult executes fn within a transaction and returns a result
func WithTransactionResult[T any](ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) (T, error)) (T, error) {
    var result T
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return result, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    result, err = fn(tx)
    if err != nil {
        return result, err
    }
    if err := tx.Commit(); err != nil {
        return result, fmt.Errorf("commit transaction: %w", err)
    }
    return result, nil
}
```
