package modula

import "context"

// EnvironmentResponse is returned by [EnvironmentResource.Get] and describes the
// server's current deployment environment and stage classification.
type EnvironmentResponse struct {
	// Environment is the raw environment identifier (e.g. "development", "staging", "production").
	Environment string `json:"environment"`
	// Stage is a normalized classification derived from the environment
	// (e.g. "local", "preview", "production").
	Stage string `json:"stage"`
}

// EnvironmentResource provides access to the server's environment metadata.
// The environment endpoint is unauthenticated and useful for conditionally
// enabling client features based on the deployment target.
// It is accessed via [Client].Environment.
type EnvironmentResource struct {
	http *httpClient
}

// Get returns the server's current environment and stage classification.
// A non-nil error indicates a network or transport failure.
func (e *EnvironmentResource) Get(ctx context.Context) (*EnvironmentResponse, error) {
	var result EnvironmentResponse
	if err := e.http.get(ctx, "/api/v1/environment", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
