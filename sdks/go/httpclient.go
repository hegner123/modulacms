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

type httpClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// get performs a GET request. params are added as query string. result is JSON-decoded.
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

// post performs a POST request with JSON body. result is JSON-decoded.
func (h *httpClient) post(ctx context.Context, path string, body any, result any) error {
	return h.jsonRequest(ctx, http.MethodPost, path, body, result)
}

// put performs a PUT request with JSON body. result is JSON-decoded.
func (h *httpClient) put(ctx context.Context, path string, body any, result any) error {
	return h.jsonRequest(ctx, http.MethodPut, path, body, result)
}

// del performs a DELETE request. params are added as query string.
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

// doRaw executes a pre-built request and returns the raw response.
// Caller is responsible for closing Body.
func (h *httpClient) doRaw(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	h.setAuth(req)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("modula: %w", err)
	}

	return resp, nil
}

// jsonRequest builds and executes a request with a JSON-encoded body.
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

// do sets the auth header, executes the request, checks the status code,
// and optionally JSON-decodes the response body into result.
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

// setAuth adds the Authorization header if an API key is configured.
func (h *httpClient) setAuth(req *http.Request) {
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}
}

// buildError reads the response body and constructs an *ApiError.
// It attempts to extract a "message" or "error" field from a JSON response body.
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
