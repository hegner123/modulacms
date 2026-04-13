package modula

import (
	"context"
	"encoding/json"
	"fmt"
)

// MetricsResource provides access to the admin metrics endpoint, which returns
// a snapshot of server runtime metrics (request counts, latencies, error rates, etc.).
// Requires config:read permission.
// It is accessed via [Client].Metrics.
type MetricsResource struct {
	http *httpClient
}

// Get returns the current server metrics snapshot as raw JSON. The shape of the
// response depends on which metrics the server is collecting and may change
// between versions. Parse into your own struct as needed.
func (m *MetricsResource) Get(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := m.http.get(ctx, "/api/v1/admin/metrics", nil, &result); err != nil {
		return nil, fmt.Errorf("get metrics: %w", err)
	}
	return result, nil
}
