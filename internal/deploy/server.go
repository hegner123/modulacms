package deploy

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// maxImportBodySize is the maximum request body size for import (100 MB).
const maxImportBodySize = 100 << 20

// DeployHealthHandler returns deploy endpoint health status and version info.
// Requires deploy:read permission.
func DeployHealthHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	cfg, _ := svc.Config()
	c := *cfg
	w.Header().Set("Content-Type", "application/json")
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"version": utility.GetCurrentVersion(),
		"node_id": c.Node_ID,
	})
}

// exportRequest is the JSON body for POST /api/v1/deploy/export.
type exportRequest struct {
	Tables         []string `json:"tables"`
	IncludePlugins bool     `json:"include_plugins"`
}

// DeployExportHandler exports CMS data as a SyncPayload JSON response.
// Requires deploy:read permission.
func DeployExportHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	driver := svc.Driver()

	// Parse optional table list and plugin flag from request body.
	var opts ExportOptions
	if r.Body != nil && r.ContentLength != 0 {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit for request body

		var req exportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		for _, name := range req.Tables {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			t, vErr := db.ValidateTableName(name)
			if vErr != nil {
				writeDeployError(w, http.StatusBadRequest, fmt.Sprintf("invalid table name %q", name), nil)
				return
			}
			opts.Tables = append(opts.Tables, t)
		}
		opts.IncludePlugins = req.IncludePlugins
	}

	ctx := r.Context()
	payload, err := ExportPayload(ctx, driver, opts)
	if err != nil {
		utility.DefaultLogger.Error("deploy export failed", err)
		writeDeployError(w, http.StatusInternalServerError, "export failed", nil)
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		utility.DefaultLogger.Error("deploy export marshal failed", err)
		writeDeployError(w, http.StatusInternalServerError, "marshal failed", nil)
		return
	}

	// Gzip the response if it's large and the client accepts gzip.
	if len(data) > gzipThreshold && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		gw.Write(data)
		gw.Close()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// DeployImportHandler applies a SyncPayload to the local database.
// Requires deploy:create permission.
func DeployImportHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	cfg, _ := svc.Config()
	c := *cfg
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImportBodySize)

	// Support gzip-encoded request bodies.
	var bodyReader io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gr, gErr := gzip.NewReader(r.Body)
		if gErr != nil {
			utility.DefaultLogger.Error("deploy import: invalid gzip body", gErr)
			writeDeployError(w, http.StatusBadRequest, "invalid gzip encoding in request body", nil)
			return
		}
		defer gr.Close()
		bodyReader = gr
	}

	var payload SyncPayload
	if err := json.NewDecoder(bodyReader).Decode(&payload); err != nil {
		utility.DefaultLogger.Error("deploy import: invalid request body", err)
		writeDeployError(w, http.StatusBadRequest, "invalid or malformed JSON in request body", nil)
		return
	}

	// Check for dry-run query parameter.
	dryRun := r.URL.Query().Get("dry_run") == "true"

	driver := svc.Driver()
	ctx := r.Context()

	if dryRun {
		result := BuildDryRunResult(&payload, driver)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	utility.DefaultLogger.Info("deploy import: starting",
		"tables", len(payload.Tables),
		"manifest_version", payload.Manifest.Version,
		"manifest_schema", payload.Manifest.SchemaVersion[:12]+"...",
		"dry_run", false)

	result, err := ImportPayload(ctx, c, driver, &payload, false)
	if err != nil {
		utility.DefaultLogger.Error("deploy import failed", err)

		status := http.StatusInternalServerError
		msg := err.Error() // propagate the actual error to the client

		if strings.Contains(msg, "import already in progress") {
			status = http.StatusConflict
		}

		if result != nil {
			// Return the structured result (contains validation errors, not internal details).
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(result)
			return
		}

		writeDeployError(w, status, msg, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(result)
}

// writeDeployError writes a structured JSON error response for deploy endpoints.
func writeDeployError(w http.ResponseWriter, status int, message string, details []SyncError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]any{"error": message}
	if len(details) > 0 {
		resp["details"] = details
	}
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(resp)
}

// SnapshotDir returns the configured snapshot directory, falling back to the default.
func SnapshotDir(c config.Config) string {
	if c.Deploy_Snapshot_Dir != "" {
		return c.Deploy_Snapshot_Dir
	}
	return "./deploy/snapshots"
}

// resolveContext creates a context with the default deploy timeout if the parent has none.
func resolveContext(parent context.Context) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		return parent, func() {}
	}
	return context.WithTimeout(parent, DefaultTimeout)
}
