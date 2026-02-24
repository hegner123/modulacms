package modula

import (
	"context"
	"encoding/json"
)

// DeployHealthResponse is returned by GET /api/v1/deploy/health.
type DeployHealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	NodeID  string `json:"node_id"`
}

// DeployExportRequest is the optional body for POST /api/v1/deploy/export.
type DeployExportRequest struct {
	Tables []string `json:"tables,omitempty"`
}

// DeploySyncResult is returned by POST /api/v1/deploy/import.
type DeploySyncResult struct {
	Success        bool              `json:"success"`
	DryRun         bool              `json:"dry_run"`
	Strategy       string            `json:"strategy"`
	TablesAffected []string          `json:"tables_affected"`
	RowCounts      map[string]int    `json:"row_counts"`
	BackupPath     string            `json:"backup_path"`
	SnapshotID     string            `json:"snapshot_id"`
	Duration       string            `json:"duration"`
	Errors         []DeploySyncError `json:"errors,omitempty"`
	Warnings       []string          `json:"warnings,omitempty"`
}

// DeploySyncError describes a specific failure during a sync operation.
type DeploySyncError struct {
	Table   string `json:"table"`
	Phase   string `json:"phase"`
	Message string `json:"message"`
	RowID   string `json:"row_id,omitempty"`
}

// DeployResource provides deploy sync operations.
type DeployResource struct {
	http *httpClient
}

// Health returns the deploy health status of the server.
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
