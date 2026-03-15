package deploy

import (
	"time"

	"github.com/hegner123/modulacms/internal/db"
)

// MergeStrategy defines how data is applied to the target during import.
type MergeStrategy string

const (
	// StrategyOverwrite nukes the target tables and replaces with source data.
	StrategyOverwrite MergeStrategy = "overwrite"
)

// TableData is the per-table wire format inside SyncPayload.
// Columns defines the ordered column names. Rows are positional, matching Columns order.
// Typed IDs serialize as their 26-char ULID string. Timestamps serialize as RFC3339 strings.
// Nullable fields serialize as JSON null or their string value.
type TableData struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

// SyncPayload is the complete wire format for deploy sync operations.
type SyncPayload struct {
	Manifest SyncManifest         `json:"manifest"`
	Tables   map[string]TableData `json:"tables"`
	UserRefs map[string]string    `json:"user_refs"` // user_id -> username
}

// SyncManifest contains metadata about the sync payload for validation.
type SyncManifest struct {
	SchemaVersion string         `json:"schema_version"` // SHA256 of sorted table:columns
	Timestamp     string         `json:"timestamp"`      // RFC3339 UTC
	SourceNodeID  string         `json:"source_node_id"`
	SourceURL     string         `json:"source_url"`
	Version       string         `json:"version"` // Modula version
	Strategy      MergeStrategy  `json:"strategy"`
	Tables        []string       `json:"tables"`                  // table names included
	RowCounts     map[string]int `json:"row_counts"`              // table -> count
	PayloadHash   string         `json:"payload_hash"`            // SHA256 of Tables map JSON
	PluginTables  []string       `json:"plugin_tables,omitempty"` // subset of Tables that are plugin tables
}

// ExportOptions controls what is included in a deploy export.
type ExportOptions struct {
	Tables         []db.DBTable // core tables; nil = DefaultTableSet
	IncludePlugins bool         // discover and include registered plugin tables
}

// SyncResult is returned after a sync operation completes.
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

// SyncError describes a specific failure during a sync operation.
// Every error includes the table and phase so the user knows exactly what failed and where.
type SyncError struct {
	Table   string `json:"table"`
	Phase   string `json:"phase"` // "export", "validate", "truncate", "insert", "verify"
	Message string `json:"message"`
	RowID   string `json:"row_id,omitempty"`
}

// SyncConfig controls the behavior of a sync operation.
type SyncConfig struct {
	Source     string // "local" or environment name
	Target     string // "local" or environment name
	Strategy   MergeStrategy
	Tables     []db.DBTable // empty = DefaultTableSet
	DryRun     bool
	SkipBackup bool
	Timeout    time.Duration // default 5 minutes
}

// DefaultTableSet is the set of tables synced by default, ordered by insertion tier.
// Truncate order is the reverse of this slice.
var DefaultTableSet = []db.DBTable{
	// Tier 1: foundation
	db.Datatype,
	db.Admin_datatype,
	// Tier 2: dependent on tier 1
	db.Field,
	db.Admin_field,
	db.Route,
	db.Admin_route,
	// Tier 3: content
	db.Content_data,
	db.Admin_content_data,
	// Tier 5: content fields
	db.Content_fields,
	db.Admin_content_fields,
	// Tier 6: relations
	db.Content_relations,
	db.Admin_content_relations,
}

// DefaultTimeout is the default maximum duration for an import operation.
const DefaultTimeout = 5 * time.Minute
