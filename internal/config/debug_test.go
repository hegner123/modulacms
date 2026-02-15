package config_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

func TestConfig_Stringify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  config.Config
		want string
	}{
		{
			name: "zero-value config returns empty string",
			cfg:  config.Config{},
			want: "",
		},
		{
			name: "config with populated fields returns empty string",
			// Stringify currently builds an empty slice regardless of fields,
			// so even a fully populated Config produces "".
			// When the implementation is filled in, this test will break --
			// that is intentional to signal that new assertions are needed.
			cfg: config.Config{
				Environment:     "production",
				Port:            ":8080",
				SSL_Port:        ":4000",
				Client_Site:     "example.com",
				Admin_Site:      "admin.example.com",
				SSH_Host:        "ssh.example.com",
				SSH_Port:        "2233",
				Db_Driver:       config.Sqlite,
				Db_Name:         "test.db",
				Db_URL:          "./test.db",
				Bucket_Endpoint: "s3.example.com",
				Bucket_Region:   "us-east-1",
				Output_Format:   config.FormatRaw,
			},
			want: "",
		},
		{
			name: "default config returns empty string",
			cfg:  config.DefaultConfig(),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.cfg.Stringify()
			if got != tt.want {
				t.Errorf("Stringify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfig_Stringify_ReturnType(t *testing.T) {
	t.Parallel()

	// Verify Stringify always returns a string (not a panic) even on a
	// zero-value Config with nil maps and nil slices.
	var c config.Config
	got := c.Stringify()

	if got != "" {
		t.Errorf("Stringify() on zero-value Config = %q, want %q", got, "")
	}
}
