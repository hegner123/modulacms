package modula

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPClient_Get(t *testing.T) {
	type testPayload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/things" {
			t.Errorf("path = %q, want /api/v1/things", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "test-id" {
			t.Errorf("query param q = %q, want test-id", r.URL.Query().Get("q"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testPayload{Name: "thing1", Value: 42})
	}))
	defer srv.Close()

	h := &httpClient{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}

	params := url.Values{}
	params.Set("q", "test-id")

	var result testPayload
	err := h.get(context.Background(), "/api/v1/things", params, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "thing1" {
		t.Errorf("Name = %q, want %q", result.Name, "thing1")
	}
	if result.Value != 42 {
		t.Errorf("Value = %d, want %d", result.Value, 42)
	}
}

func TestHTTPClient_Post(t *testing.T) {
	type reqBody struct {
		Label string `json:"label"`
	}
	type respBody struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var body reqBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Label != "test-label" {
			t.Errorf("body.Label = %q, want %q", body.Label, "test-label")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(respBody{ID: "abc123", Label: body.Label})
	}))
	defer srv.Close()

	h := &httpClient{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}

	var result respBody
	err := h.post(context.Background(), "/api/v1/things", reqBody{Label: "test-label"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc123" {
		t.Errorf("ID = %q, want %q", result.ID, "abc123")
	}
	if result.Label != "test-label" {
		t.Errorf("Label = %q, want %q", result.Label, "test-label")
	}
}

func TestHTTPClient_Put(t *testing.T) {
	type reqBody struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	}
	type respBody struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var body reqBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respBody{ID: body.ID, Label: body.Label})
	}))
	defer srv.Close()

	h := &httpClient{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}

	var result respBody
	err := h.put(context.Background(), "/api/v1/things/", reqBody{ID: "abc123", Label: "updated"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Label != "updated" {
		t.Errorf("Label = %q, want %q", result.Label, "updated")
	}
}

func TestHTTPClient_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Query().Get("q") != "del-id" {
			t.Errorf("query param q = %q, want del-id", r.URL.Query().Get("q"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h := &httpClient{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}

	params := url.Values{}
	params.Set("q", "del-id")

	err := h.del(context.Background(), "/api/v1/things/", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClient_AuthHeader(t *testing.T) {
	t.Run("with API key", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer test-api-key" {
				t.Errorf("Authorization = %q, want %q", auth, "Bearer test-api-key")
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"ok":true}`)
		}))
		defer srv.Close()

		h := &httpClient{
			baseURL:    srv.URL,
			apiKey:     "test-api-key",
			httpClient: srv.Client(),
		}

		var result map[string]any
		err := h.get(context.Background(), "/api/v1/test", nil, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("without API key", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "" {
				t.Errorf("Authorization = %q, want empty", auth)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"ok":true}`)
		}))
		defer srv.Close()

		h := &httpClient{
			baseURL:    srv.URL,
			apiKey:     "",
			httpClient: srv.Client(),
		}

		var result map[string]any
		err := h.get(context.Background(), "/api/v1/test", nil, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestHTTPClient_ErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "400 Bad Request",
			statusCode: 400,
			body:       "bad request",
		},
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			body:       "unauthorized",
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			body:       "not found",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			body:       "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				io.WriteString(w, tt.body)
			}))
			defer srv.Close()

			h := &httpClient{
				baseURL:    srv.URL,
				httpClient: srv.Client(),
			}

			var result map[string]any
			err := h.get(context.Background(), "/api/v1/test", nil, &result)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			apiErr, ok := err.(*ApiError)
			if !ok {
				t.Fatalf("error type = %T, want *ApiError", err)
			}
			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}
			if apiErr.Body != tt.body {
				t.Errorf("Body = %q, want %q", apiErr.Body, tt.body)
			}
		})
	}
}

func TestHTTPClient_ErrorWithJSONMessage(t *testing.T) {
	t.Run("message field", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "field is required"})
		}))
		defer srv.Close()

		h := &httpClient{
			baseURL:    srv.URL,
			httpClient: srv.Client(),
		}

		var result map[string]any
		err := h.get(context.Background(), "/api/v1/test", nil, &result)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		apiErr, ok := err.(*ApiError)
		if !ok {
			t.Fatalf("error type = %T, want *ApiError", err)
		}
		if apiErr.Message != "field is required" {
			t.Errorf("Message = %q, want %q", apiErr.Message, "field is required")
		}
	})

	t.Run("error field", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "access denied"})
		}))
		defer srv.Close()

		h := &httpClient{
			baseURL:    srv.URL,
			httpClient: srv.Client(),
		}

		var result map[string]any
		err := h.get(context.Background(), "/api/v1/test", nil, &result)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		apiErr, ok := err.(*ApiError)
		if !ok {
			t.Fatalf("error type = %T, want *ApiError", err)
		}
		if apiErr.Message != "access denied" {
			t.Errorf("Message = %q, want %q", apiErr.Message, "access denied")
		}
	})

	t.Run("message takes precedence over error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "primary message",
				"error":   "secondary error",
			})
		}))
		defer srv.Close()

		h := &httpClient{
			baseURL:    srv.URL,
			httpClient: srv.Client(),
		}

		var result map[string]any
		err := h.get(context.Background(), "/api/v1/test", nil, &result)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		apiErr, ok := err.(*ApiError)
		if !ok {
			t.Fatalf("error type = %T, want *ApiError", err)
		}
		if apiErr.Message != "primary message" {
			t.Errorf("Message = %q, want %q", apiErr.Message, "primary message")
		}
	})
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	h := &httpClient{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var result map[string]any
	err := h.get(ctx, "/api/v1/test", nil, &result)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
