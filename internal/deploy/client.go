package deploy

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/db"
)

// DeployClient communicates with a remote Modula deploy API.
type DeployClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// HealthResponse is the JSON body returned by GET /api/v1/deploy/health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	NodeID  string `json:"node_id"`
}

// ClientError is returned when the remote deploy API returns a non-2xx status.
type ClientError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *ClientError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("deploy remote: %d %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("deploy remote: HTTP %d", e.StatusCode)
}

// NewDeployClient creates a client for the given remote URL and Bearer token.
// The URL should be the base URL of a Modula instance (e.g. "https://cms.example.com").
func NewDeployClient(baseURL, apiKey string) *DeployClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &DeployClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// Health calls GET /api/v1/deploy/health on the remote instance.
func (c *DeployClient) Health(ctx context.Context) (*HealthResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/deploy/health", nil)
	if err != nil {
		return nil, fmt.Errorf("build health request: %w", err)
	}

	var resp HealthResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Export calls POST /api/v1/deploy/export on the remote instance and returns the SyncPayload.
func (c *DeployClient) Export(ctx context.Context, tables []db.DBTable) (*SyncPayload, error) {
	var body any
	if len(tables) > 0 {
		names := make([]string, len(tables))
		for i, t := range tables {
			names[i] = string(t)
		}
		body = exportRequest{Tables: names}
	}

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encode export request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/deploy/export", &buf)
	if err != nil {
		return nil, fmt.Errorf("build export request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	var payload SyncPayload
	if err := c.do(req, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// Import calls POST /api/v1/deploy/import on the remote instance with the given payload.
// Gzip-compresses the request body if it exceeds gzipThreshold.
func (c *DeployClient) Import(ctx context.Context, payload *SyncPayload) (*SyncResult, error) {
	return c.doImport(ctx, payload, false)
}

// DryRunImport calls POST /api/v1/deploy/import?dry_run=true on the remote instance.
// Validates the payload without modifying the target database.
func (c *DeployClient) DryRunImport(ctx context.Context, payload *SyncPayload) (*SyncResult, error) {
	return c.doImport(ctx, payload, true)
}

// doImport sends the payload to the remote import endpoint, optionally as dry-run.
func (c *DeployClient) doImport(ctx context.Context, payload *SyncPayload, dryRun bool) (*SyncResult, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode import payload: %w", err)
	}

	url := c.baseURL + "/api/v1/deploy/import"
	if dryRun {
		url += "?dry_run=true"
	}

	var body io.Reader
	compressed := len(data) > gzipThreshold

	if compressed {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, wErr := gw.Write(data); wErr != nil {
			gw.Close()
			return nil, fmt.Errorf("gzip compress import payload: %w", wErr)
		}
		if err := gw.Close(); err != nil {
			return nil, fmt.Errorf("gzip close: %w", err)
		}
		body = &buf
	} else {
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("build import request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if compressed {
		req.Header.Set("Content-Encoding", "gzip")
	}

	var result SyncResult
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// do sets the auth header, executes the request, checks status, and decodes JSON into result.
// Requests gzip encoding and transparently decompresses gzip responses.
func (c *DeployClient) do(req *http.Request, result any) error {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("deploy remote request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.buildError(resp)
	}

	if result != nil {
		var bodyReader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gr, gErr := gzip.NewReader(resp.Body)
			if gErr != nil {
				return fmt.Errorf("decompress gzip response: %w", gErr)
			}
			defer gr.Close()
			bodyReader = gr
		}
		if err := json.NewDecoder(bodyReader).Decode(result); err != nil {
			return fmt.Errorf("decode deploy response: %w", err)
		}
	}

	return nil
}

// maxErrorBodySize limits how much of an error response body we read (1 MB).
const maxErrorBodySize = 1 << 20

// buildError reads the response body and constructs a *ClientError.
func (c *DeployClient) buildError(resp *http.Response) *ClientError {
	rawBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))

	ce := &ClientError{
		StatusCode: resp.StatusCode,
	}

	if readErr != nil {
		return ce
	}

	ce.Body = strings.TrimSpace(string(rawBody))

	var parsed map[string]any
	if json.Unmarshal(rawBody, &parsed) == nil {
		if msg, ok := parsed["error"].(string); ok && msg != "" {
			ce.Message = msg
		} else if msg, ok := parsed["message"].(string); ok && msg != "" {
			ce.Message = msg
		}
	}

	return ce
}
