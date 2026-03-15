# deploy

The deploy package provides content synchronization between ModulaCMS instances. It exports content tables as a JSON payload, transmits it to a target instance, and imports it with validation, backup, and audit trail support.

## Overview

The sync engine exports 12 core content tables (datatypes, fields, routes, content_data, content_fields, content_relations and their admin counterparts) as a self-contained JSON payload. The payload includes a manifest with schema version, row counts, and integrity hash. Import truncates target tables and bulk-inserts all rows inside an atomic transaction with foreign key checks disabled.

Plugin tables (created by Lua plugins, prefixed `plugin_`) can optionally be included via `ExportOptions.IncludePlugins`. Plugin tables are discovered from the `tables` registry, exported via catalog introspection, and imported with schema-match validation.

## Types

### ExportOptions

```go
type ExportOptions struct {
    Tables         []db.DBTable
    IncludePlugins bool
}
```

Controls what is included in an export. `Tables` defaults to `DefaultTableSet` if nil. `IncludePlugins` discovers and includes registered plugin tables.

### SyncPayload

```go
type SyncPayload struct {
    Manifest SyncManifest         `json:"manifest"`
    Tables   map[string]TableData `json:"tables"`
    UserRefs map[string]string    `json:"user_refs"`
}
```

The complete wire format for deploy sync. `UserRefs` maps user IDs to usernames for placeholder user creation on import.

### SyncManifest

```go
type SyncManifest struct {
    SchemaVersion string         `json:"schema_version"`
    Timestamp     string         `json:"timestamp"`
    SourceNodeID  string         `json:"source_node_id"`
    SourceURL     string         `json:"source_url"`
    Version       string         `json:"version"`
    Strategy      MergeStrategy  `json:"strategy"`
    Tables        []string       `json:"tables"`
    RowCounts     map[string]int `json:"row_counts"`
    PayloadHash   string         `json:"payload_hash"`
    PluginTables  []string       `json:"plugin_tables,omitempty"`
}
```

Metadata for validation. `SchemaVersion` is a SHA256 of sorted table:columns pairs. `PayloadHash` is a SHA256 of the JSON-encoded tables map. `PluginTables` lists which entries in `Tables` are plugin tables.

### TableData

```go
type TableData struct {
    Columns []string `json:"columns"`
    Rows    [][]any  `json:"rows"`
}
```

Per-table wire format. Columns are ordered; rows are positional arrays matching column order.

### SyncResult

```go
type SyncResult struct {
    Success        bool           `json:"success"`
    DryRun         bool           `json:"dry_run"`
    Strategy       MergeStrategy  `json:"strategy"`
    TablesAffected []string       `json:"tables_affected"`
    RowCounts      map[string]int `json:"row_counts"`
    BackupPath     string         `json:"backup_path"`
    SnapshotID     string         `json:"snapshot_id"`
    Duration       string         `json:"duration"`
    Errors         []SyncError    `json:"errors,omitempty"`
    Warnings       []string       `json:"warnings,omitempty"`
}
```

Returned after import or dry run.

### SyncError

```go
type SyncError struct {
    Table   string `json:"table"`
    Phase   string `json:"phase"`
    Message string `json:"message"`
    RowID   string `json:"row_id,omitempty"`
}
```

Describes a specific failure. Phase is one of: `validate`, `import`, `verify`.

### MergeStrategy

```go
type MergeStrategy string
const StrategyOverwrite MergeStrategy = "overwrite"
```

Currently only `overwrite` is supported (truncate + insert).

### DefaultTableSet

```go
var DefaultTableSet = []db.DBTable{ ... }
```

The 12 core tables exported by default, ordered by insertion tier (tier 1 first, tier 6 last). Truncation happens in reverse order.

### DefaultTimeout

```go
const DefaultTimeout = 5 * time.Minute
```

Maximum duration for an import operation.

## Functions

### ExportPayload

```go
func ExportPayload(ctx context.Context, driver db.DbDriver, opts ExportOptions) (*SyncPayload, error)
```

Exports data from the driver into a SyncPayload. Core tables use typed List methods via `tableListFuncs`. Plugin tables (when `opts.IncludePlugins` is true) use `DeployOps.QueryAllRows` for catalog-based export without requiring Go struct types.

### ExportToFile

```go
func ExportToFile(ctx context.Context, driver db.DbDriver, opts ExportOptions, outPath string) (*SyncManifest, string, error)
```

Exports to JSON file. Gzip-compresses if output exceeds 1 GB. Returns the manifest and actual file path written.

### ImportPayload

```go
func ImportPayload(ctx context.Context, cfg config.Config, driver db.DbDriver, payload *SyncPayload, skipBackup bool) (*SyncResult, error)
```

Imports a SyncPayload into the target database. Validates the payload, acquires an import lock, creates a pre-import snapshot and optional backup, then runs truncate + insert inside an atomic transaction. Plugin tables in the payload are imported after core tables; missing plugin tables on the destination are skipped with a warning.

### ImportFromFile

```go
func ImportFromFile(ctx context.Context, cfg config.Config, driver db.DbDriver, inPath string, skipBackup bool) (*SyncResult, error)
```

Reads a JSON/gzip export file and calls ImportPayload.

### Pull

```go
func Pull(ctx context.Context, cfg config.Config, driver db.DbDriver, envName string, opts ExportOptions, skipBackup bool, dryRun bool) (*SyncResult, error)
```

Exports from a remote instance and imports locally. Environment is resolved from `cfg.Deploy_Environments`.

### Push

```go
func Push(ctx context.Context, cfg config.Config, driver db.DbDriver, envName string, opts ExportOptions, dryRun bool) (*SyncResult, error)
```

Exports locally and imports on a remote instance.

### BuildDryRunResult

```go
func BuildDryRunResult(payload *SyncPayload, driver db.DbDriver) *SyncResult
```

Validates the payload and returns an impact report without modifying the database.

### TestEnvConnection

```go
func TestEnvConnection(ctx context.Context, cfg config.Config, envName string) (*HealthResponse, error)
```

Tests connectivity and authentication to a configured deploy environment.

### ValidatePayload

```go
func ValidatePayload(payload *SyncPayload, targetDriver db.DbDriver) []SyncError
```

Pre-import validation. Checks: payload hash, row counts, schema version, ULID format in ID columns (skipped for plugin tables), content datatype FKs, content tree pointers, user refs completeness, and plugin table row width consistency.

### VerifyImport

```go
func VerifyImport(ctx context.Context, ops db.DeployOps, ex db.Executor, expected map[db.DBTable]int) []SyncError
```

Post-import validation. Checks FK integrity and verifies row counts match expectations.

### ReadPayloadFile

```go
func ReadPayloadFile(path string) ([]byte, error)
```

Reads a file, auto-detecting and decompressing gzip.

## HTTP Handlers

### DeployHealthHandler

`GET /api/v1/deploy/health` -- Returns status, version, and node ID.

### DeployExportHandler

`POST /api/v1/deploy/export` -- Accepts optional JSON body with `tables` (string array) and `include_plugins` (bool). Returns the SyncPayload JSON. Gzip-compresses if large and client accepts gzip.

### DeployImportHandler

`POST /api/v1/deploy/import` -- Accepts SyncPayload JSON (supports gzip request body). With `?dry_run=true`, validates without writing.

## DeployClient

```go
func NewDeployClient(baseURL, apiKey string) *DeployClient
```

HTTP client for remote Modula deploy API. Methods:

- `Health(ctx) (*HealthResponse, error)`
- `Export(ctx, ExportOptions) (*SyncPayload, error)`
- `Import(ctx, *SyncPayload) (*SyncResult, error)`
- `DryRunImport(ctx, *SyncPayload) (*SyncResult, error)`

## DeployClient Types

### HealthResponse

```go
type HealthResponse struct {
    Status  string `json:"status"`
    Version string `json:"version"`
    NodeID  string `json:"node_id"`
}
```

### ClientError

```go
type ClientError struct {
    StatusCode int
    Message    string
    Body       string
}
```

Implements `error`. Returned when the remote API responds with a non-2xx status.

## Snapshots

Pre-import snapshots are saved to a configurable directory.

### SnapshotInfo

```go
type SnapshotInfo struct {
    ID        string
    Timestamp time.Time
    Tables    []string
    RowCounts map[string]int
    SizeBytes int64
}
```

### Functions

- `SaveSnapshot(dir string, payload *SyncPayload) (string, error)` -- Saves payload JSON with a ULID-based filename.
- `ListSnapshots(dir string) ([]SnapshotInfo, error)` -- Lists all snapshots in the directory.
- `LoadSnapshot(dir, id string) (*SyncPayload, error)` -- Loads a snapshot by ID.
- `RestoreSnapshot(ctx context.Context, cfg config.Config, driver db.DbDriver, dir, id string) (*SyncResult, error)` -- Loads and imports a snapshot.
- `SnapshotDir(cfg config.Config) string` -- Returns the snapshot directory path.
