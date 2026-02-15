package config_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

func TestBucketEndpointURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		env      string
		want     string
	}{
		{
			name:     "empty endpoint returns empty string",
			endpoint: "",
			env:      "production",
			want:     "",
		},
		{
			name:     "production environment uses https",
			endpoint: "s3.example.com",
			env:      "production",
			want:     "https://s3.example.com",
		},
		{
			name:     "staging environment uses https",
			endpoint: "s3.staging.example.com",
			env:      "staging",
			want:     "https://s3.staging.example.com",
		},
		{
			name:     "development environment uses https",
			endpoint: "localhost:9000",
			env:      "development",
			want:     "https://localhost:9000",
		},
		{
			name:     "http-only environment uses http",
			endpoint: "localhost:9000",
			env:      "http-only",
			want:     "http://localhost:9000",
		},
		{
			name:     "docker environment uses http",
			endpoint: "minio:9000",
			env:      "docker",
			want:     "http://minio:9000",
		},
		{
			name:     "unknown environment defaults to https",
			endpoint: "s3.example.com",
			env:      "custom-env",
			want:     "https://s3.example.com",
		},
		{
			name:     "empty environment defaults to https",
			endpoint: "s3.example.com",
			env:      "",
			want:     "https://s3.example.com",
		},
		{
			name:     "endpoint with port uses correct scheme",
			endpoint: "localhost:9000",
			env:      "production",
			want:     "https://localhost:9000",
		},
		{
			name:     "endpoint with path preserves path",
			endpoint: "s3.example.com/bucket",
			env:      "production",
			want:     "https://s3.example.com/bucket",
		},
		{
			name:     "empty endpoint with http-only still returns empty",
			endpoint: "",
			env:      "http-only",
			want:     "",
		},
		{
			name:     "empty endpoint with docker still returns empty",
			endpoint: "",
			env:      "docker",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Config{
				Bucket_Endpoint: tt.endpoint,
				Environment:     tt.env,
			}
			got := c.BucketEndpointURL()
			if got != tt.want {
				t.Errorf("BucketEndpointURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValidOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
		want   bool
	}{
		// All valid formats
		{
			name:   "contentful is valid",
			format: "contentful",
			want:   true,
		},
		{
			name:   "sanity is valid",
			format: "sanity",
			want:   true,
		},
		{
			name:   "strapi is valid",
			format: "strapi",
			want:   true,
		},
		{
			name:   "wordpress is valid",
			format: "wordpress",
			want:   true,
		},
		{
			name:   "clean is valid",
			format: "clean",
			want:   true,
		},
		{
			name:   "raw is valid",
			format: "raw",
			want:   true,
		},
		{
			// FormatDefault is "" -- empty string is valid because it defaults to raw
			name:   "empty string is valid (defaults to raw)",
			format: "",
			want:   true,
		},
		// Invalid formats
		{
			name:   "unknown format is invalid",
			format: "unknown",
			want:   false,
		},
		{
			name:   "uppercase variant is invalid",
			format: "Contentful",
			want:   false,
		},
		{
			name:   "mixed case is invalid",
			format: "WordPress",
			want:   false,
		},
		{
			name:   "whitespace is invalid",
			format: " raw ",
			want:   false,
		},
		{
			name:   "random string is invalid",
			format: "json",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := config.IsValidOutputFormat(tt.format)
			if got != tt.want {
				t.Errorf("IsValidOutputFormat(%q) = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestGetValidOutputFormats(t *testing.T) {
	t.Parallel()

	got := config.GetValidOutputFormats()

	// Verify the expected formats are present.
	// The function should return all named formats (not the empty-string default).
	expected := []string{"contentful", "sanity", "strapi", "wordpress", "clean", "raw"}

	if len(got) != len(expected) {
		t.Fatalf("GetValidOutputFormats() returned %d formats, want %d\ngot:  %v\nwant: %v",
			len(got), len(expected), got, expected)
	}

	// Build a set of expected values for lookup
	expectedSet := make(map[string]bool, len(expected))
	for _, e := range expected {
		expectedSet[e] = true
	}

	for _, format := range got {
		if !expectedSet[format] {
			t.Errorf("GetValidOutputFormats() contains unexpected format %q", format)
		}
	}

	// Verify every expected format is present in the result
	gotSet := make(map[string]bool, len(got))
	for _, g := range got {
		gotSet[g] = true
	}
	for _, e := range expected {
		if !gotSet[e] {
			t.Errorf("GetValidOutputFormats() missing expected format %q", e)
		}
	}
}

func TestGetValidOutputFormats_AllAreValid(t *testing.T) {
	t.Parallel()

	// Every format returned by GetValidOutputFormats should pass IsValidOutputFormat.
	// This catches drift between the two functions.
	formats := config.GetValidOutputFormats()
	for _, format := range formats {
		if !config.IsValidOutputFormat(format) {
			t.Errorf("GetValidOutputFormats() includes %q but IsValidOutputFormat(%q) returns false", format, format)
		}
	}
}

func TestDbDriverConstants(t *testing.T) {
	t.Parallel()

	// Verify the driver constants have the expected string values,
	// since downstream code (database selection) depends on these exact strings.
	tests := []struct {
		name string
		got  config.DbDriver
		want string
	}{
		{name: "sqlite driver", got: config.Sqlite, want: "sqlite"},
		{name: "mysql driver", got: config.Mysql, want: "mysql"},
		{name: "postgres driver", got: config.Psql, want: "postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.got) != tt.want {
				t.Errorf("DbDriver constant = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestEndpointConstants(t *testing.T) {
	t.Parallel()

	// Verify endpoint constants have the expected values since they are used
	// as map keys in Oauth_Endpoint.
	tests := []struct {
		name string
		got  config.Endpoint
		want string
	}{
		{name: "auth URL", got: config.OauthAuthURL, want: "oauth_auth_url"},
		{name: "token URL", got: config.OauthTokenURL, want: "oauth_token_url"},
		{name: "userinfo URL", got: config.OauthUserInfoURL, want: "oauth_userinfo_url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.got) != tt.want {
				t.Errorf("Endpoint constant = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestOutputFormatConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  config.OutputFormat
		want string
	}{
		{name: "contentful", got: config.FormatContentful, want: "contentful"},
		{name: "sanity", got: config.FormatSanity, want: "sanity"},
		{name: "strapi", got: config.FormatStrapi, want: "strapi"},
		{name: "wordpress", got: config.FormatWordPress, want: "wordpress"},
		{name: "clean", got: config.FormatClean, want: "clean"},
		{name: "raw", got: config.FormatRaw, want: "raw"},
		{name: "default is empty", got: config.FormatDefault, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.got) != tt.want {
				t.Errorf("OutputFormat constant = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

