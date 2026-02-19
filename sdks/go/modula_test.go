package modula

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClient_RequiresBaseURL(t *testing.T) {
	_, err := NewClient(ClientConfig{})
	if err == nil {
		t.Fatal("expected error for empty BaseURL, got nil")
	}
	want := "modula: BaseURL is required"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "single trailing slash",
			baseURL: "https://cms.example.com/",
			want:    "https://cms.example.com",
		},
		{
			name:    "multiple trailing slashes",
			baseURL: "https://cms.example.com///",
			want:    "https://cms.example.com",
		},
		{
			name:    "no trailing slash",
			baseURL: "https://cms.example.com",
			want:    "https://cms.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ClientConfig{BaseURL: tt.baseURL})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Verify the baseURL was trimmed by checking that a resource's
			// internal httpClient has the correct baseURL.
			// We access via the Datatypes resource's http field.
			got := client.Datatypes.http.baseURL
			if got != tt.want {
				t.Errorf("baseURL = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	client, err := NewClient(ClientConfig{BaseURL: "https://cms.example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	httpCl := client.Datatypes.http.httpClient
	if httpCl == nil {
		t.Fatal("expected default HTTP client to be set, got nil")
	}
	if httpCl.Timeout != 30*time.Second {
		t.Errorf("default client timeout = %v, want %v", httpCl.Timeout, 30*time.Second)
	}
}

func TestNewClient_CustomHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 60 * time.Second}
	client, err := NewClient(ClientConfig{
		BaseURL:    "https://cms.example.com",
		HTTPClient: custom,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := client.Datatypes.http.httpClient
	if got != custom {
		t.Error("expected custom HTTP client to be used")
	}
	if got.Timeout != 60*time.Second {
		t.Errorf("custom client timeout = %v, want %v", got.Timeout, 60*time.Second)
	}
}

func TestNewClient_AllResourcesInitialized(t *testing.T) {
	client, err := NewClient(ClientConfig{BaseURL: "https://cms.example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		isNil bool
	}{
		// Standard CRUD resources
		{"ContentData", client.ContentData == nil},
		{"ContentFields", client.ContentFields == nil},
		{"ContentRelations", client.ContentRelations == nil},
		{"Datatypes", client.Datatypes == nil},
		{"DatatypeFields", client.DatatypeFields == nil},
		{"Fields", client.Fields == nil},
		{"Media", client.Media == nil},
		{"MediaDimensions", client.MediaDimensions == nil},
		{"Routes", client.Routes == nil},
		{"Roles", client.Roles == nil},
		{"Users", client.Users == nil},
		{"Tokens", client.Tokens == nil},
		{"UsersOauth", client.UsersOauth == nil},
		{"Tables", client.Tables == nil},

		// Admin CRUD resources
		{"AdminContentData", client.AdminContentData == nil},
		{"AdminContentFields", client.AdminContentFields == nil},
		{"AdminDatatypes", client.AdminDatatypes == nil},
		{"AdminDatatypeFields", client.AdminDatatypeFields == nil},
		{"AdminFields", client.AdminFields == nil},
		{"AdminRoutes", client.AdminRoutes == nil},

		// Specialized resources
		{"Auth", client.Auth == nil},
		{"MediaUpload", client.MediaUpload == nil},
		{"AdminTree", client.AdminTree == nil},
		{"Content", client.Content == nil},
		{"SSHKeys", client.SSHKeys == nil},
		{"Sessions", client.Sessions == nil},
		{"Import", client.Import == nil},
		{"ContentBatch", client.ContentBatch == nil},

		// RBAC resources
		{"RolePermissions", client.RolePermissions == nil},

		// Plugin resources
		{"Plugins", client.Plugins == nil},
		{"PluginRoutes", client.PluginRoutes == nil},
		{"PluginHooks", client.PluginHooks == nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isNil {
				t.Errorf("%s resource is nil, want non-nil", tt.name)
			}
		})
	}
}
