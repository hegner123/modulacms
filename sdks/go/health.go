package modula

import "context"

// HealthResponse is returned by GET /api/v1/health.
type HealthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]bool   `json:"checks,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// HealthResource provides health check operations.
type HealthResource struct {
	http *httpClient
}

// Check returns the server health status.
func (h *HealthResource) Check(ctx context.Context) (*HealthResponse, error) {
	var result HealthResponse
	if err := h.http.get(ctx, "/api/v1/health", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
