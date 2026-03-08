package modula

import (
	"context"
	"encoding/json"
)

// DeployHealthResponse is returned by [DeployResource.Health] and reports the
// deploy subsystem's readiness.
type DeployHealthResponse struct {
	// Status indicates the deploy subsystem state (e.g., "ok", "degraded").
	Status string `json:"status"`
	// Version is the ModulaCMS server version running on this node.
	Version string `json:"version"`
	// NodeID is the unique identifier of this CMS instance, used for
	// distinguishing nodes in multi-instance deployments.
	NodeID string `json:"node_id"`
}

// DeployExportRequest configures which tables to include in a deploy export.
// When Tables is empty or nil, all tables are exported.
type DeployExportRequest struct {
	// Tables is an optional list of table names to include in the export.
	// When empty, all tables are exported.
	Tables []string `json:"tables,omitempty"`
}

// DeploySyncResult is returned by [DeployResource.Import] and [DeployResource.DryRunImport].
// It summarizes the outcome of a content sync operation.
type DeploySyncResult struct {
	// Success is true if the import completed without errors.
	Success bool `json:"success"`
	// DryRun is true if the operation was a preview that did not write changes.
	DryRun bool `json:"dry_run"`
	// Strategy describes the merge strategy used (e.g., "upsert", "replace").
	Strategy string `json:"strategy"`
	// TablesAffected lists the database tables that were modified.
	TablesAffected []string `json:"tables_affected"`
	// RowCounts maps each affected table name to the number of rows written.
	RowCounts map[string]int `json:"row_counts"`
	// BackupPath is the filesystem path of the pre-import backup, if created.
	BackupPath string `json:"backup_path"`
	// SnapshotID identifies this sync snapshot for auditing and rollback.
	SnapshotID string `json:"snapshot_id"`
	// Duration is the human-readable elapsed time for the operation.
	Duration string `json:"duration"`
	// Errors lists any per-table or per-row failures encountered during sync.
	Errors []DeploySyncError `json:"errors,omitempty"`
	// Warnings lists non-fatal issues discovered during sync.
	Warnings []string `json:"warnings,omitempty"`
}

// DeploySyncError describes a specific failure during a sync operation,
// pinpointing the table, phase, and optionally the exact row that failed.
type DeploySyncError struct {
	// Table is the database table where the error occurred.
	Table string `json:"table"`
	// Phase indicates which stage of sync failed (e.g., "validate", "insert", "update").
	Phase string `json:"phase"`
	// Message is a human-readable description of the error.
	Message string `json:"message"`
	// RowID is the ULID of the specific row that failed, if applicable.
	RowID string `json:"row_id,omitempty"`
}

// DeployResource provides content synchronization operations between CMS environments.
// The typical workflow is: export from a source environment, then import into a target
// environment (optionally with a dry-run first to preview changes).
//
// Deploy operations create automatic backups before writing, and each sync is
// tracked with a snapshot ID for auditing and potential rollback.
// It is accessed via [Client].Deploy.
type DeployResource struct {
	http *httpClient
}

// Health returns the deploy subsystem's health status, including the server version
// and node ID. Use this to verify connectivity and compatibility before syncing.
func (d *DeployResource) Health(ctx context.Context) (*DeployHealthResponse, error) {
	var result DeployHealthResponse
	if err := d.http.get(ctx, "/api/v1/deploy/health", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Export exports a sync payload from the server. If tables is non-empty, only those
// tables are included; otherwise the full database is exported.
// The returned json.RawMessage is the opaque SyncPayload suitable for Import.
func (d *DeployResource) Export(ctx context.Context, tables []string) (json.RawMessage, error) {
	var body any
	if len(tables) > 0 {
		body = DeployExportRequest{Tables: tables}
	}
	var result json.RawMessage
	if err := d.http.post(ctx, "/api/v1/deploy/export", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Import imports a sync payload into the server.
// The payload should be the json.RawMessage returned by Export.
func (d *DeployResource) Import(ctx context.Context, payload json.RawMessage) (*DeploySyncResult, error) {
	var result DeploySyncResult
	if err := d.http.post(ctx, "/api/v1/deploy/import", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DryRunImport performs a dry-run import, returning what would change without writing.
// The payload should be the json.RawMessage returned by Export.
func (d *DeployResource) DryRunImport(ctx context.Context, payload json.RawMessage) (*DeploySyncResult, error) {
	var result DeploySyncResult
	if err := d.http.post(ctx, "/api/v1/deploy/import?dry_run=true", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
