package modula

import "context"

// HealthResponse is returned by [HealthResource.Check] and provides an overview
// of the CMS server's operational status.
type HealthResponse struct {
	// Status is the aggregate health state (e.g., "healthy", "degraded", "unhealthy").
	Status string `json:"status"`
	// Checks maps individual subsystem names to their pass/fail status.
	// Common keys include "database", "storage", and "cache".
	Checks map[string]bool `json:"checks,omitempty"`
	// Details provides human-readable additional information for each subsystem,
	// such as version numbers or connection latencies.
	Details map[string]string `json:"details,omitempty"`
}

// HealthResource provides health check operations for monitoring the CMS server.
// The health endpoint is unauthenticated and suitable for load balancer probes,
// uptime monitoring, and readiness checks.
// It is accessed via [Client].Health.
type HealthResource struct {
	http *httpClient
}

// Check returns the server health status, including individual subsystem checks
// (database, storage, etc.) and their details. A non-nil error indicates a
// network or transport failure; check [HealthResponse.Status] for application-level health.
func (h *HealthResource) Check(ctx context.Context) (*HealthResponse, error) {
	var result HealthResponse
	if err := h.http.get(ctx, "/api/v1/health", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
