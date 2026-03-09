package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/service"
)

type healthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]bool   `json:"checks"`
	Details map[string]string `json:"details,omitempty"`
}

// PluginHealthChecker is a function that returns plugin subsystem health.
// Nil when the plugin system is disabled.
type PluginHealthChecker func() PluginHealthResult

// PluginHealthResult mirrors plugin.PluginHealthStatus without importing the plugin package.
// This avoids tight coupling between the health endpoint and the plugin package.
type PluginHealthResult struct {
	Healthy             bool
	FailedPlugins       []string
	OpenCircuitBreakers []string
}

// HealthHandler reports the health of critical dependencies (database, S3 storage, plugins).
// Returns 200 when all checks pass, 503 when any critical check fails.
func HealthHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry, pluginHealthFn PluginHealthChecker) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	resp := healthResponse{
		Status:  "ok",
		Checks:  make(map[string]bool),
		Details: make(map[string]string),
	}

	// Database check
	resp.Checks["database"] = checkDatabase(ctx, svc, resp.Details)

	// S3 storage check (skip if not configured)
	if cfg.Bucket_Endpoint != "" {
		resp.Checks["storage"] = checkStorage(ctx, svc, resp.Details)
	}

	// Plugin health check (skip if plugin system is disabled)
	if pluginHealthFn != nil {
		status := pluginHealthFn()
		resp.Checks["plugins"] = status.Healthy
		if !status.Healthy {
			resp.Details["plugins"] = fmt.Sprintf("failed: %v, open_cb: %v", status.FailedPlugins, status.OpenCircuitBreakers)
		}
	}

	// Determine overall status
	for _, ok := range resp.Checks {
		if !ok {
			resp.Status = "degraded"
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(resp)
}

// checkDatabase pings the database and records the result.
func checkDatabase(_ context.Context, svc *service.Registry, details map[string]string) bool {
	driver := svc.Driver()
	if driver == nil {
		details["database"] = "driver not initialized"
		return false
	}
	if err := driver.Ping(); err != nil {
		details["database"] = err.Error()
		return false
	}
	return true
}

// checkStorage verifies the S3 media bucket is reachable via HeadBucket.
func checkStorage(_ context.Context, svc *service.Registry, details map[string]string) bool {
	cfg, err := svc.Config()
	if err != nil {
		details["storage"] = "configuration unavailable"
		return false
	}

	creds := bucket.S3Credentials{
		AccessKey:      cfg.Bucket_Access_Key,
		SecretKey:      cfg.Bucket_Secret_Key,
		URL:            cfg.BucketEndpointURL(),
		Region:         cfg.Bucket_Region,
		ForcePathStyle: cfg.Bucket_Force_Path_Style,
	}

	s3svc, err := creds.GetBucket()
	if err != nil {
		details["storage"] = err.Error()
		return false
	}

	_, err = s3svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket_Media),
	})
	if err != nil {
		details["storage"] = err.Error()
		return false
	}
	return true
}
