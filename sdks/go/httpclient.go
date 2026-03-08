package modula

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// httpClient is the internal HTTP transport shared by all resources on a
// [Client]. It handles base URL construction, Bearer token authentication,
// JSON encoding/decoding, and HTTP error mapping to [*ApiError].
//
// Timeout behavior is inherited from the underlying [http.Client] provided
// in [ClientConfig]. The default is 30 seconds. Context deadlines on
// individual requests take precedence over the client-level timeout.
type httpClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// get performs a GET request to baseURL+path with optional query parameters.
// The response body is JSON-decoded into result. Pass nil for result to
// discard the body.
func (h *httpClient) get(ctx context.Context, path string, params url.Values, result any) error {
	fullURL := h.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("modula: %w", err)
	}

	return h.do(req, result)
}

// post performs a POST request with a JSON-encoded body. The response is
// JSON-decoded into result.
func (h *httpClient) post(ctx context.Context, path string, body any, result any) error {
	return h.jsonRequest(ctx, http.MethodPost, path, body, result)
}

// put performs a PUT request with a JSON-encoded body. The response is
// JSON-decoded into result.
func (h *httpClient) put(ctx context.Context, path string, body any, result any) error {
	return h.jsonRequest(ctx, http.MethodPut, path, body, result)
}

// del performs a DELETE request with optional query parameters. The response
// body is discarded. Use delBody when the server returns data on deletion.
func (h *httpClient) del(ctx context.Context, path string, params url.Values) error {
	fullURL := h.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return fmt.Errorf("modula: %w", err)
	}

	return h.do(req, nil)
}

// delBody performs a DELETE request and JSON-decodes the response body into
// result. Used for delete endpoints that return the deleted entity or a
// summary of cascaded deletions.
func (h *httpClient) delBody(ctx context.Context, path string, params url.Values, result any) error {
	fullURL := h.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return fmt.Errorf("modula: %w", err)
	}

	return h.do(req, result)
}

// doRaw executes a pre-built request and returns the raw [http.Response].
// Authentication headers are set automatically. The caller is responsible
// for closing the response Body. This is used by specialized resources
// (e.g. MediaUpload) that need direct access to the response.
func (h *httpClient) doRaw(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	h.setAuth(req)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("modula: %w", err)
	}

	return resp, nil
}

// jsonRequest builds and executes an HTTP request with a JSON-encoded body
// and Content-Type: application/json header. Used by post and put.
func (h *httpClient) jsonRequest(ctx context.Context, method string, path string, body any, result any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("modula: encoding request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, h.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("modula: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return h.do(req, result)
}

// do is the central request executor. It sets the Authorization header,
// sends the request, checks the HTTP status code (returning [*ApiError]
// for any non-2xx response), and optionally JSON-decodes the response body
// into result. Pass nil for result to discard the body.
func (h *httpClient) do(req *http.Request, result any) error {
	h.setAuth(req)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("modula: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return h.buildError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("modula: decoding response: %w", err)
		}
	}

	return nil
}

// setAuth adds the "Authorization: Bearer <apiKey>" header to the request.
// If no API key was configured, this is a no-op.
func (h *httpClient) setAuth(req *http.Request) {
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}
}

// buildError reads the response body and constructs an [*ApiError].
// It attempts to parse the body as JSON and extract a "message" or "error"
// field for a human-readable error description. If the body is not JSON or
// neither field is present, Message is left empty and the raw Body is still
// available for inspection.
func (h *httpClient) buildError(resp *http.Response) *ApiError {
	rawBody, readErr := io.ReadAll(resp.Body)

	apiErr := &ApiError{
		StatusCode: resp.StatusCode,
	}

	if readErr != nil {
		return apiErr
	}

	bodyStr := strings.TrimSpace(string(rawBody))
	apiErr.Body = bodyStr

	var parsed map[string]any
	if json.Unmarshal(rawBody, &parsed) == nil {
		if msg, ok := parsed["message"].(string); ok && msg != "" {
			apiErr.Message = msg
		} else if errMsg, ok := parsed["error"].(string); ok && errMsg != "" {
			apiErr.Message = errMsg
		}
	}

	return apiErr
}
