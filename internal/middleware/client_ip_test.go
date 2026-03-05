package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		wantIP     string
	}{
		{
			name:       "plain RemoteAddr with port",
			remoteAddr: "192.168.1.1:12345",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			wantIP:     "203.0.113.50",
		},
		{
			name:       "X-Forwarded-For multiple IPs takes first",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18, 150.172.238.178"},
			wantIP:     "203.0.113.50",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Real-IP": "203.0.113.99"},
			wantIP:     "203.0.113.99",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "10.0.0.1:1234",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.50",
				"X-Real-IP":       "203.0.113.99",
			},
			wantIP: "203.0.113.50",
		},
		{
			name:       "IPv6 RemoteAddr",
			remoteAddr: "[::1]:12345",
			wantIP:     "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotIP string
			handler := ClientIPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotIP = ClientIPFromContext(r.Context())
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			handler.ServeHTTP(httptest.NewRecorder(), req)

			if gotIP != tt.wantIP {
				t.Errorf("ClientIPFromContext() = %q, want %q", gotIP, tt.wantIP)
			}
		})
	}
}

func TestClientIPFromContext_Missing(t *testing.T) {
	ip := ClientIPFromContext(context.Background())
	if ip != "" {
		t.Errorf("ClientIPFromContext(empty) = %q, want empty", ip)
	}
}
