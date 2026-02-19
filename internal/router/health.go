package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

type healthResponse struct {
	Status  string          `json:"status"`
	Checks  map[string]bool `json:"checks"`
	Details map[string]string `json:"details,omitempty"`
}

// HealthHandler reports the health of critical dependencies (database, S3 storage).
// Returns 200 when all checks pass, 503 when any critical check fails.
func HealthHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := healthResponse{
		Status:  "ok",
		Checks:  make(map[string]bool),
		Details: make(map[string]string),
	}

	// Database check
	resp.Checks["database"] = checkDatabase(ctx, c, resp.Details)

	// S3 storage check (skip if not configured)
	if c.Bucket_Endpoint != "" {
		resp.Checks["storage"] = checkStorage(ctx, c, resp.Details)
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
func checkDatabase(_ context.Context, c config.Config, details map[string]string) bool {
	driver := db.ConfigDB(c)
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
func checkStorage(_ context.Context, c config.Config, details map[string]string) bool {
	creds := bucket.S3Credentials{
		AccessKey:      c.Bucket_Access_Key,
		SecretKey:      c.Bucket_Secret_Key,
		URL:            c.BucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}

	svc, err := creds.GetBucket()
	if err != nil {
		details["storage"] = err.Error()
		return false
	}

	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(c.Bucket_Media),
	})
	if err != nil {
		details["storage"] = err.Error()
		return false
	}
	return true
}
