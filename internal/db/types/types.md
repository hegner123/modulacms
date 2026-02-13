# types

Package types provides custom database types for ModulaCMS with built-in validation, cross-database compatibility, and JSON marshaling. All types implement `sql.driver.Valuer`, `sql.Scanner`, `json.Marshaler`, and `json.Unmarshaler`.

## ULID Generation

### NewULID

Generates a thread-safe 26-character ULID. Uses monotonic entropy source to guarantee ordering within milliseconds and prevent collisions across goroutines. Returns `ulid.ULID` value.

### validateULID

Validates a string as a properly formatted 26-character ULID. Checks length and format. Returns error if invalid. Used internally by all ID type `Validate` methods.

## ID Types

All ID types are string-based ULIDs with consistent method interfaces. Primary keys for database entities.

### DatatypeID

Uniquely identifies a datatype. Represents a content type definition such as Page, Blog Post, or Product.

#### NewDatatypeID

Returns a new ULID-based DatatypeID. Thread-safe generation with monotonic ordering guarantees.

#### String

Returns the DatatypeID as a string value. Returns the underlying 26-character ULID.

#### IsZero

Returns true if the DatatypeID is an empty string. Used to check for uninitialized values.

#### Validate

Validates the DatatypeID format. Ensures 26-character length and valid ULID encoding. Returns error if validation fails.

#### ULID

Parses the DatatypeID into a `ulid.ULID` value. Returns error if parsing fails due to invalid format.

#### Time

Extracts the timestamp embedded in the ULID. Returns `time.Time` representing when this ID was generated. Returns error if ULID parsing fails.

#### Value

Implements `driver.Valuer` for database writes. Returns string value or error if DatatypeID is empty. Empty IDs cannot be written to database.

#### Scan

Implements `sql.Scanner` for database reads. Accepts string or byte slice from database. Validates after scanning. Returns error for nil values or invalid types.

#### MarshalJSON

Marshals DatatypeID to JSON string. Returns JSON-encoded 26-character ULID.

#### UnmarshalJSON

Unmarshals DatatypeID from JSON string. Validates format after unmarshaling. Returns error for invalid ULID format.

#### ParseDatatypeID

Creates a DatatypeID from a string with validation. Returns error if string is not a valid ULID. Use this for parsing user input.

### UserID

Uniquely identifies a user. Represents a system user account with authentication and permissions.

Methods: `NewUserID`, `String`, `IsZero`, `Validate`, `ULID`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`, `ParseUserID`. Identical semantics to DatatypeID.

### RoleID

Uniquely identifies a role. Represents a permission set that can be assigned to users.

Methods: `NewRoleID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### PermissionID

Uniquely identifies a permission. Represents a single access control rule for a table or resource.

Methods: `NewPermissionID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### FieldID

Uniquely identifies a field definition. Represents a field configuration that can be attached to datatypes.

Methods: `NewFieldID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### ContentID

Uniquely identifies content data. Represents a single content node in the tree structure.

Methods: `NewContentID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### ContentFieldID

Uniquely identifies a content field value. Represents a field value attached to specific content.

Methods: `NewContentFieldID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### MediaID

Uniquely identifies media. Represents an uploaded image, video, or document in S3-compatible storage.

Methods: `NewMediaID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### MediaDimensionID

Uniquely identifies a media dimension preset. Represents configured image sizes for responsive images.

Methods: `NewMediaDimensionID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### SessionID

Uniquely identifies a session. Represents an active user session with authentication state.

Methods: `NewSessionID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### TokenID

Uniquely identifies a token. Represents an API token, refresh token, or verification token.

Methods: `NewTokenID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### RouteID

Uniquely identifies a route. Represents a URL route that serves content trees to clients.

Methods: `NewRouteID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### AdminRouteID

Uniquely identifies an admin route. Represents a route for admin panel content.

Methods: `NewAdminRouteID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### TableID

Uniquely identifies a table registry entry. Represents a tracked database table for plugin access.

Methods: `NewTableID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### UserOauthID

Uniquely identifies a user OAuth record. Links user accounts to external OAuth providers like Google or GitHub.

Methods: `NewUserOauthID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### UserSshKeyID

Uniquely identifies a user SSH key. Represents a public key for SSH-based TUI authentication.

Methods: `NewUserSshKeyID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### AdminDatatypeID

Uniquely identifies an admin datatype. Represents content types for admin panel content.

Methods: `NewAdminDatatypeID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### AdminFieldID

Uniquely identifies an admin field definition. Represents field configurations for admin content.

Methods: `NewAdminFieldID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### AdminContentID

Uniquely identifies admin content data. Represents content nodes in admin panel trees.

Methods: `NewAdminContentID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### AdminContentFieldID

Uniquely identifies an admin content field value. Represents field values for admin content.

Methods: `NewAdminContentFieldID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### DatatypeFieldID

Uniquely identifies a datatype-field relationship. Represents the junction table linking datatypes to their field definitions.

Methods: `NewDatatypeFieldID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### AdminDatatypeFieldID

Uniquely identifies an admin datatype-field relationship. Junction for admin datatypes and admin fields.

Methods: `NewAdminDatatypeFieldID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### EventID

Uniquely identifies a change event. Represents an entry in the change_events audit log for replication and webhooks.

Methods: `NewEventID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### NodeID

Uniquely identifies a node in distributed deployment. Represents a single ModulaCMS instance in a multi-node cluster.

Methods: `NewNodeID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### BackupID

Uniquely identifies a backup. Represents a database backup snapshot on S3-compatible storage.

Methods: `NewBackupID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### VerificationID

Uniquely identifies a backup verification. Represents a verification attempt for a backup file.

Methods: `NewVerificationID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### BackupSetID

Uniquely identifies a backup set. Represents a collection of related backups for restoration.

Methods: `NewBackupSetID`, `String`, `IsZero`, `Validate`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### ContentRelationID

Uniquely identifies a content relation. Represents a link between two content nodes via a relation field.

Methods: `NewContentRelationID`, `String`, `IsZero`, `Validate`, `ULID`, `Time`, `ParseContentRelationID`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

### AdminContentRelationID

Uniquely identifies an admin content relation. Links admin content nodes through relation fields.

Methods: `NewAdminContentRelationID`, `String`, `IsZero`, `Validate`, `ULID`, `Time`, `ParseAdminContentRelationID`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to DatatypeID.

## Nullable ID Types

Nullable versions of ID types for foreign keys that can be NULL. All nullable types have `ID` field and `Valid` bool field.

### NullableDatatypeID

Nullable foreign key to datatypes table. Used when parent datatype is optional.

#### Validate

Validates the ID if Valid is true. Returns nil for null values. Returns error if ID format is invalid.

#### String

Returns the ID as string or "null" if not valid. Used for logging and debugging.

#### IsZero

Returns true if not valid or ID is empty. Checks both validity flag and ID content.

#### Value

Implements `driver.Valuer`. Returns nil for invalid, string for valid. Allows NULL in database column.

#### Scan

Implements `sql.Scanner`. Sets Valid to false for nil values. Sets Valid to true and scans ID for non-nil values.

#### MarshalJSON

Returns JSON null for invalid, marshals ID for valid. Produces correct JSON nullable representation.

#### UnmarshalJSON

Handles JSON null values. Sets Valid to false for null, true for non-null. Validates ID after unmarshaling.

### NullableUserID

Nullable foreign key to users table. Used for optional author or owner references.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableContentID

Nullable foreign key to content table. Used for parent_id in content tree structure.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableFieldID

Nullable foreign key to fields table. Used for optional field references.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableMediaID

Nullable foreign key to media table. Used for optional media attachments.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableRouteID

Nullable foreign key to routes table. Used for optional route associations.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableAdminRouteID

Nullable foreign key to admin_routes table. Optional admin route references.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableAdminContentID

Nullable foreign key to admin_content_data table. Parent references in admin content trees.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableNodeID

Nullable node ID for distributed systems. Optional node identification.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableRoleID

Nullable foreign key to roles table. Optional role assignments.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableAdminDatatypeID

Nullable foreign key to admin_datatypes table. Optional admin datatype references.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

### NullableAdminFieldID

Nullable foreign key to admin_fields table. Optional admin field references.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to NullableDatatypeID.

## Enums

String-based enums with validation and predefined constants. All enums reject empty strings and invalid values.

### ContentStatus

Publication status of content. Tracks content lifecycle state.

Constants: `ContentStatusDraft`, `ContentStatusPublished`, `ContentStatusArchived`, `ContentStatusPending`.

#### Validate

Ensures value is one of the defined constants. Returns error for empty or invalid values.

#### String

Returns the enum value as a string. Returns underlying string content.

#### Value

Implements `driver.Valuer` for database writes. Validates before returning. Returns error for invalid values.

#### Scan

Implements `sql.Scanner` for database reads. Accepts string or byte slice. Validates after scanning.

#### MarshalJSON

Marshals to JSON string. Returns JSON-encoded enum value.

#### UnmarshalJSON

Unmarshals from JSON string. Validates after unmarshaling. Returns error for invalid values.

### FieldType

Type of content field. Determines validation rules and UI rendering.

Constants: `FieldTypeText`, `FieldTypeTextarea`, `FieldTypeNumber`, `FieldTypeDate`, `FieldTypeDatetime`, `FieldTypeBoolean`, `FieldTypeSelect`, `FieldTypeMedia`, `FieldTypeRelation`, `FieldTypeJSON`, `FieldTypeRichText`, `FieldTypeSlug`, `FieldTypeEmail`, `FieldTypeURL`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### RouteType

Type of route. Determines how route handles requests.

Constants: `RouteTypeStatic`, `RouteTypeDynamic`, `RouteTypeAPI`, `RouteTypeRedirect`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### Operation

Database operation type for change events. Maps to SQL operations.

Constants: `OpInsert`, `OpUpdate`, `OpDelete`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### Action

Business-level action for change events. More specific than Operation.

Constants: `ActionCreate`, `ActionUpdate`, `ActionDelete`, `ActionPublish`, `ActionArchive`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### ConflictPolicy

Conflict resolution strategy for distributed systems. Determines how to handle concurrent edits.

Constants: `ConflictLWW` (last write wins), `ConflictManual` (flag for human resolution).

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### BackupType

Type of backup operation. Determines backup strategy.

Constants: `BackupTypeFull`, `BackupTypeIncremental`, `BackupTypeDifferential`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### BackupStatus

Status of backup operation. Tracks backup lifecycle.

Constants: `BackupStatusPending`, `BackupStatusInProgress`, `BackupStatusCompleted`, `BackupStatusFailed`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### VerificationStatus

Status of backup verification. Tracks verification attempts.

Constants: `VerificationStatusPending`, `VerificationStatusVerified`, `VerificationStatusFailed`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

### BackupSetStatus

Status of backup set. Tracks collection completeness.

Constants: `BackupSetStatusPending`, `BackupSetStatusComplete`, `BackupSetStatusPartial`.

Methods: `Validate`, `String`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Identical semantics to ContentStatus.

## Timestamp

Handles datetime columns across SQLite TEXT, MySQL DATETIME, PostgreSQL TIMESTAMP. All times stored and returned in UTC.

### NewTimestamp

Creates a Timestamp from a `time.Time`. Converts to UTC automatically.

### TimestampNow

Returns current time as Timestamp in UTC. Uses `time.Now` internally.

### String

Returns RFC3339 format string in UTC or "null" for invalid timestamps. Format: `2006-01-02T15:04:05Z`.

### IsZero

Returns true if timestamp is null or zero time. Checks both validity flag and time value.

### Value

Implements `driver.Valuer`. Returns UTC time for database writes. Returns nil for invalid timestamps.

### Scan

Implements `sql.Scanner`. Accepts `time.Time`, string, or byte slice. Tries strict RFC3339 formats first, falls back to legacy formats for database compatibility. Always converts to UTC.

### MarshalJSON

Marshals to RFC3339 string in UTC. Returns JSON null for invalid timestamps. Format matches API standard.

### UnmarshalJSON

Unmarshals from RFC3339 format only. Strict validation for API input. Returns error for legacy formats. Always converts to UTC.

### Before

Compares timestamps. Returns true if this timestamp is before the argument. Returns false for invalid timestamps.

### After

Compares timestamps. Returns true if this timestamp is after the argument. Returns false for invalid timestamps.

### UTC

Returns the time in UTC timezone. Extracts `time.Time` from wrapper.

### Add

Adds a duration to timestamp. Returns new timestamp with added duration. Preserves validity flag.

### Sub

Subtracts timestamps. Returns duration between this and argument. Returns zero duration for invalid timestamps.

## Validation Types

Types with built-in validation for semantic correctness.

### Slug

URL path starting with slash. Lowercase alphanumeric segments with hyphens. Examples: `/`, `/about`, `/blog/2024/my-post`.

#### Validate

Validates slug format. Must start with `/`. Allows segments with `a-z`, `0-9`, `-`. Cannot end with `-` or `/`. Cannot have consecutive hyphens or slashes. Returns error with specific format violations.

#### String

Returns slug as string value. Returns underlying string content.

#### IsZero

Returns true if slug is empty string. Checks for uninitialized value.

#### Slugify

Converts any string to valid slug. Lowercases input, replaces spaces and underscores with hyphens, removes invalid characters, collapses duplicates, ensures leading slash. Input "About Us" becomes "/about-us".

#### Value

Implements `driver.Valuer`. Returns error if slug is empty. Cannot write empty slugs to database.

#### Scan

Implements `sql.Scanner`. Accepts string or byte slice. Skips validation for legacy data compatibility.

#### MarshalJSON

Marshals to JSON string. Returns JSON-encoded slug value.

#### UnmarshalJSON

Unmarshals from JSON string. Validates format after unmarshaling. Returns error for invalid format.

### Email

Validated email address with local and domain parts. Maximum 254 characters.

#### Validate

Validates email format. Requires exactly one `@` symbol. Validates local part allows `a-z`, `A-Z`, `0-9`, `._%+-`. Validates domain requires at least one dot and 2-character TLD. Returns error with specific validation failures.

#### String

Returns email as string value. Returns underlying string content.

#### IsZero

Returns true if email is empty string. Checks for uninitialized value.

#### Domain

Extracts domain part after `@` symbol. Returns empty string if email format is invalid.

#### Value

Implements `driver.Valuer`. Returns error if email is empty. Cannot write empty emails to database.

#### Scan

Implements `sql.Scanner`. Accepts string or byte slice. Validates after scanning.

#### MarshalJSON

Marshals to JSON string. Returns JSON-encoded email value.

#### UnmarshalJSON

Unmarshals from JSON string. Validates format after unmarshaling. Returns error for invalid format.

### URL

Validated URL with required scheme and host. Must be absolute URL.

#### Validate

Validates URL format using Go standard library. Requires non-empty scheme like `https`. Requires non-empty host. Returns error for relative URLs or missing components.

#### String

Returns URL as string value. Returns underlying string content.

#### IsZero

Returns true if URL is empty string. Checks for uninitialized value.

#### Parse

Parses URL into `*url.URL` from standard library. Returns parsed URL or error. Use for URL manipulation.

#### Value

Implements `driver.Valuer`. Returns error if URL is empty. Cannot write empty URLs to database.

#### Scan

Implements `sql.Scanner`. Accepts string or byte slice. Validates after scanning.

#### MarshalJSON

Marshals to JSON string. Returns JSON-encoded URL value.

#### UnmarshalJSON

Unmarshals from JSON string. Validates format after unmarshaling. Returns error for invalid format.

### NullableSlug

Nullable version of Slug for optional fields. Structure: `Slug` field, `Valid` bool field.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Same semantics as nullable ID types.

### NullableEmail

Nullable version of Email for optional fields. Structure: `Email` field, `Valid` bool field.

#### Domain

Extracts domain part from email or returns empty string if null. Adds domain extraction to nullable wrapper.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Same semantics as nullable ID types.

### NullableURL

Nullable version of URL for optional fields. Structure: `URL` field, `Valid` bool field.

#### Parse

Parses URL to `*url.URL` or returns nil if null. Adds parsing to nullable wrapper.

Methods: `Validate`, `String`, `IsZero`, `Value`, `Scan`, `MarshalJSON`, `UnmarshalJSON`. Same semantics as nullable ID types.

## Nullable Primitive Types

Nullable wrappers for standard Go types with cross-database compatibility.

### NullableString

Nullable string type. Handles NULL VARCHAR, TEXT across all databases.

#### NewNullableString

Creates a valid NullableString from string. Sets Valid flag to true.

#### Value

Returns nil for invalid, string for valid. Implements `driver.Valuer`.

#### Scan

Handles nil, string, byte slice from database. Sets Valid flag appropriately.

#### MarshalJSON

Returns JSON null for invalid, JSON string for valid. Produces correct JSON representation.

#### UnmarshalJSON

Handles JSON null values. Sets Valid to false for null, true and unmarshals for non-null.

#### IsZero

Returns true if not valid or string is empty. Checks both validity and content.

### NullableInt64

Nullable 64-bit integer. Handles NULL INTEGER, BIGINT across all databases.

#### NewNullableInt64

Creates a valid NullableInt64 from int64. Sets Valid flag to true.

#### Value

Returns nil for invalid, int64 for valid. Implements `driver.Valuer`.

#### Scan

Handles nil, int64, int32, int, float64 from database. Converts int32 to int64 for MySQL compatibility.

#### MarshalJSON

Returns JSON null for invalid, JSON number for valid. Produces correct JSON representation.

#### UnmarshalJSON

Handles JSON null values. Unmarshals numbers for non-null.

#### IsZero

Returns true if not valid or int64 is zero. Checks both validity and zero value.

### NullableBool

Nullable boolean. Handles NULL BOOLEAN, TINYINT across all databases.

#### NewNullableBool

Creates a valid NullableBool from bool. Sets Valid flag to true.

#### Value

Returns nil for invalid, bool for valid. Implements `driver.Valuer`.

#### Scan

Handles nil, bool, int64, int32, int from database. Converts integers to boolean where non-zero is true.

#### MarshalJSON

Returns JSON null for invalid, JSON boolean for valid. Produces correct JSON representation.

#### UnmarshalJSON

Handles JSON null values. Unmarshals boolean for non-null.

#### IsZero

Returns true if not valid. Boolean zero value is false which is a valid state.

### JSONData

Arbitrary JSON data stored across SQLite TEXT, MySQL JSON, PostgreSQL JSONB. Stores any marshalable Go value.

#### NewJSONData

Creates a valid JSONData from any value. Stores value for later marshaling.

#### Value

Marshals data to JSON bytes for database storage. Returns nil for invalid or nil data.

#### Scan

Unmarshals JSON from string or byte slice. Accepts any valid JSON. Returns error for invalid JSON.

#### MarshalJSON

Marshals underlying data to JSON. Returns JSON null for invalid or nil data.

#### UnmarshalJSON

Unmarshals JSON into underlying data field. Handles JSON null values.

#### IsZero

Returns true if not valid or data is nil. Checks both validity and content.

#### String

Returns JSON representation as string. Returns empty string for invalid data.

## Hybrid Logical Clock

HLC provides distributed timestamp ordering across nodes. Format: upper 48 bits wall time milliseconds, lower 16 bits logical counter.

### HLCNow

Generates new HLC timestamp greater than any previously issued. Thread-safe with mutex protection. Increments counter for same-millisecond calls. Waits for wall clock advance if counter overflows.

### HLCUpdate

Merges received HLC from remote node with local time. Used when receiving change events from other nodes. Ensures monotonicity across distributed system. Thread-safe with mutex protection.

### Physical

Extracts wall time from HLC timestamp. Returns `time.Time` from upper 48 bits. Represents milliseconds since Unix epoch.

### Logical

Extracts counter from HLC timestamp. Returns uint16 from lower 16 bits. Represents ordering within same millisecond.

### String

Formats HLC as human-readable string. Format: `HLC(wallms:counter)`. Example: `HLC(1234567890:5)`.

### Value

Implements `driver.Valuer` for database storage. Returns int64 representation.

### Scan

Implements `sql.Scanner` for database reads. Accepts int64 or int. Returns zero HLC for nil.

### MarshalJSON

Marshals HLC as JSON number. Returns int64 representation.

### UnmarshalJSON

Unmarshals HLC from JSON number. Accepts int64 value.

### Before

Compares HLC timestamps. Returns true if this HLC happened before argument. Simple integer comparison.

### After

Compares HLC timestamps. Returns true if this HLC happened after argument. Simple integer comparison.

## Change Events

Change event types for audit trail, replication, and webhooks.

### ChangeEvent

Represents a database change event. Used for audit trails, replication logs, and webhook triggers. Contains event metadata, operation details, and optional old/new values.

Fields: `EventID`, `HLCTimestamp`, `WallTimestamp`, `NodeID`, `TableName`, `RecordID`, `Operation`, `Action`, `UserID`, `OldValues`, `NewValues`, `Metadata`, `SyncedAt`, `ConsumedAt`.

### EventLogger

Interface for change event operations. Defines methods for logging events, querying by record, retrieving unsynced or unconsumed events, and marking events as processed.

#### LogEvent

Records a change event to database. Accepts context and event. Returns error on failure.

#### GetEventsByRecord

Retrieves all events for a specific table record. Accepts context, table name, record ID. Returns slice of events or error.

#### GetEventsSince

Retrieves events after an HLC timestamp for replication. Accepts context, HLC cutoff, limit. Returns events in HLC order.

#### GetUnsyncedEvents

Retrieves events not yet synced to other nodes. Accepts context and limit. Returns events ready for replication.

#### GetUnconsumedEvents

Retrieves events not yet processed by webhooks. Accepts context and limit. Returns events ready for webhook dispatch.

#### MarkSynced

Marks events as synced to other nodes. Accepts context and event ID slice. Returns error on failure.

#### MarkConsumed

Marks events as consumed by webhooks. Accepts context and event ID slice. Returns error on failure.

### NewChangeEvent

Creates a change event for current node. Generates new event ID and HLC timestamp. Accepts node ID, table name, record ID, operation, action, user ID. Returns initialized ChangeEvent.

### WithChanges

Adds old and new values to change event. Accepts old value, new value. Returns modified event. Use for UPDATE operations.

### WithMetadata

Adds metadata to change event. Accepts any metadata value. Returns modified event. Use for custom event annotations.

## Transactions

Transaction helpers for database operations with automatic commit and rollback.

### TxFunc

Function type that executes within a transaction. Signature: `func(tx *sql.Tx) error`. Used with WithTransaction helper.

### WithTransaction

Executes function within transaction with automatic rollback on error and commit on success. Accepts context, database connection, transaction function. Returns error from function or transaction operations. Always calls Rollback in defer as no-op after successful commit.

### WithTransactionResult

Generic transaction helper that returns a result value. Executes function within transaction. Accepts context, database connection, function returning result and error. Returns result and error. Type parameter T can be any type. Rolls back on error, commits on success.

## Field Configuration

JSON configuration types for field definitions.

### EmptyJSON

Canonical empty JSON object constant. Value: `"{}"`. Used as default for JSON columns.

### Cardinality

Relation field cardinality. Determines if relation targets one or many items.

Constants: `CardinalityOne`, `CardinalityMany`.

#### Validate

Ensures value is one or many. Returns error for invalid values.

### RelationConfig

Parsed form of relation field data config. Contains target datatype ID, cardinality, and optional max depth.

Fields: `TargetDatatypeID`, `Cardinality`, `MaxDepth`.

### ValidationConfig

Parsed form of field validation config. Contains validation rules for field values.

Fields: `Required`, `MinLength`, `MaxLength`, `Min`, `Max`, `Pattern`, `MaxItems`.

### UIConfig

Parsed form of field UI config. Contains rendering hints for admin interfaces.

Fields: `Widget`, `Placeholder`, `HelpText`, `Hidden`.

### ParseValidationConfig

Parses JSON string to ValidationConfig. Returns zero-value config for empty or `"{}"` input. Returns error for malformed JSON.

### ParseUIConfig

Parses JSON string to UIConfig. Returns zero-value config for empty or `"{}"` input. Returns error for malformed JSON.

### ParseRelationConfig

Parses JSON string to RelationConfig. Returns error for empty or `"{}"` input because relation fields require non-empty config. Validates target datatype ID and cardinality. Returns error for missing required fields.
