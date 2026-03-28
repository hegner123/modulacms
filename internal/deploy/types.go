package deploy

import (
	"time"

	"github.com/hegner123/modulacms/internal/config"
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
	Phase   string `json:"phase"` // "validate", "import", "verify"
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

// ContentTableSet is the conservative default for admin panel push/pull.
// Only syncs content rows — no schema, no users, no media.
var ContentTableSet = []db.DBTable{
	db.Content_data,
	db.Admin_content_data,
	db.Content_fields,
	db.Admin_content_fields,
	db.Content_relations,
	db.Admin_content_relations,
}

// FullTableSet is every syncable table ordered by FK dependency tier.
// Truncate order is the reverse of this slice.
var FullTableSet = []db.DBTable{
	// Tier 0: standalone (no FK dependencies)
	db.Role,
	db.Permission,
	db.LocaleT,
	db.Field_types,
	db.Admin_field_types,
	db.Table,
	db.PipelineT,
	// Tier 1: depends on tier 0
	db.Role_permissions,
	db.User,
	db.Datatype,
	db.Admin_datatype,
	db.ValidationT,
	db.Admin_validation,
	// Tier 2: depends on tier 1
	db.User_oauth,
	db.User_ssh_keys,
	db.Session,
	db.Token,
	db.Field,
	db.Admin_field,
	db.Route,
	db.Admin_route,
	db.WebhookT,
	db.Media_folder,
	db.Admin_media_folder,
	// Tier 3: depends on tier 2
	db.Content_data,
	db.Admin_content_data,
	db.MediaT,
	db.Admin_media,
	db.Media_dimension,
	db.Webhook_deliveries,
	db.Field_plugin_config,
	// Tier 4: depends on tier 3
	db.Content_fields,
	db.Admin_content_fields,
	db.Content_versions,
	db.Admin_content_versions,
	// Tier 5: depends on tier 4
	db.Content_relations,
	db.Admin_content_relations,
	// Tier 6: append-only / audit
	db.Change_event,
	db.BackupT,
	db.Backup_set,
	db.Backup_verification,
}

// TableGroup is a display grouping for the deploy table selection UI.
type TableGroup struct {
	Label  string
	Tables []db.DBTable
}

// SyncableTableGroups organizes all syncable tables into categories for the UI.
var SyncableTableGroups = []TableGroup{
	{Label: "Content", Tables: []db.DBTable{
		db.Content_data, db.Admin_content_data,
		db.Content_fields, db.Admin_content_fields,
		db.Content_relations, db.Admin_content_relations,
		db.Content_versions, db.Admin_content_versions,
	}},
	{Label: "Schema", Tables: []db.DBTable{
		db.Datatype, db.Admin_datatype,
		db.Field, db.Admin_field,
		db.Field_types, db.Admin_field_types,
		db.Route, db.Admin_route,
		db.ValidationT, db.Admin_validation,
	}},
	{Label: "Media", Tables: []db.DBTable{
		db.MediaT, db.Admin_media,
		db.Media_dimension,
		db.Media_folder, db.Admin_media_folder,
	}},
	{Label: "Identity", Tables: []db.DBTable{
		db.User, db.User_oauth, db.User_ssh_keys,
		db.Role, db.Permission, db.Role_permissions,
		db.Session, db.Token,
	}},
	{Label: "System", Tables: []db.DBTable{
		db.LocaleT,
		db.WebhookT, db.Webhook_deliveries,
		db.PipelineT, db.Field_plugin_config,
		db.Change_event, db.Table,
		db.BackupT, db.Backup_set, db.Backup_verification,
	}},
}

// ResolveTables validates a set of table name strings, orders them by
// FullTableSet insertion tier, and returns the result. Unknown names are
// silently ignored. Returns ContentTableSet if input is empty or yields
// no valid tables.
func ResolveTables(names []string) []db.DBTable {
	if len(names) == 0 {
		return ContentTableSet
	}
	wanted := make(map[string]bool, len(names))
	for _, n := range names {
		wanted[n] = true
	}
	var result []db.DBTable
	for _, t := range FullTableSet {
		if wanted[string(t)] {
			result = append(result, t)
		}
	}
	if len(result) == 0 {
		return ContentTableSet
	}
	return result
}

// TablesForEnv resolves the table set for a deploy environment config.
// If the config has explicit tables, those are used (ordered by FullTableSet).
// Otherwise returns ContentTableSet.
func TablesForEnv(env config.DeployEnvironmentConfig) []db.DBTable {
	return ResolveTables(env.Tables)
}

// ContentTableNames returns the string names of ContentTableSet for use
// as default checkbox values in the UI.
func ContentTableNames() []string {
	names := make([]string, len(ContentTableSet))
	for i, t := range ContentTableSet {
		names[i] = string(t)
	}
	return names
}

// DefaultTimeout is the default maximum duration for an import operation.
const DefaultTimeout = 5 * time.Minute
