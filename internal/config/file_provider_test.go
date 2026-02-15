package config_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// Compile-time check: *FileProvider must satisfy the Provider interface.
var _ config.Provider = (*config.FileProvider)(nil)

func TestNewFileProvider_CustomPath(t *testing.T) {
	t.Parallel()

	// A nonexistent absolute path: Get should fail with an "opening config file" error,
	// confirming NewFileProvider stored the path we gave it.
	fp := config.NewFileProvider("/some/custom/path-that-does-not-exist.json")
	if fp == nil {
		t.Fatal("NewFileProvider returned nil")
	}

	_, err := fp.Get()
	if err == nil {
		t.Fatal("expected error from Get on nonexistent path, got nil")
	}
	if !strings.Contains(err.Error(), "opening config file") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "opening config file")
	}
}

func TestNewFileProvider_EmptyStringDefaultPath(t *testing.T) {
	t.Parallel()

	// When given an empty string, NewFileProvider defaults to "config.json".
	// We cannot control whether that file exists in the test working directory,
	// so we only verify the constructor returns a non-nil provider and that
	// Get does not panic.
	fp := config.NewFileProvider("")
	if fp == nil {
		t.Fatal("NewFileProvider(\"\") returned nil")
	}

	// Call Get -- if config.json exists it returns a Config, otherwise an error.
	// Either outcome is acceptable; the important thing is no panic.
	_, _ = fp.Get()
}

func TestNewFileProvider_ExplicitPath(t *testing.T) {
	t.Parallel()

	// Write a known config file and confirm NewFileProvider uses that exact path.
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "custom.json")

	if err := os.WriteFile(cfgPath, []byte(`{"environment":"custom-path-test"}`), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if got.Environment != "custom-path-test" {
		t.Errorf("Environment = %q, want %q", got.Environment, "custom-path-test")
	}
}

func TestFileProvider_Get_ValidConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test-config.json")

	// Write a minimal valid config JSON
	cfgJSON := `{
		"environment": "production",
		"port": ":9090",
		"ssl_port": ":4433",
		"db_driver": "postgres",
		"db_name": "mydb",
		"db_url": "localhost:5432",
		"ssh_host": "ssh.example.com",
		"ssh_port": "2222",
		"bucket_region": "eu-west-1",
		"cookie_name": "test_cookie"
	}`

	if err := os.WriteFile(cfgPath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "Environment", got: got.Environment, want: "production"},
		{name: "Port", got: got.Port, want: ":9090"},
		{name: "SSL_Port", got: got.SSL_Port, want: ":4433"},
		{name: "Db_Driver", got: string(got.Db_Driver), want: "postgres"},
		{name: "Db_Name", got: got.Db_Name, want: "mydb"},
		{name: "Db_URL", got: got.Db_URL, want: "localhost:5432"},
		{name: "SSH_Host", got: got.SSH_Host, want: "ssh.example.com"},
		{name: "SSH_Port", got: got.SSH_Port, want: "2222"},
		{name: "Bucket_Region", got: got.Bucket_Region, want: "eu-west-1"},
		{name: "Cookie_Name", got: got.Cookie_Name, want: "test_cookie"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestFileProvider_Get_FileNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	noSuchFile := filepath.Join(dir, "does-not-exist.json")

	fp := config.NewFileProvider(noSuchFile)
	_, err := fp.Get()

	if err == nil {
		t.Fatal("Get() expected error for missing file, got nil")
	}

	// Error should wrap os.ErrNotExist via the fmt.Errorf %w chain
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected error to wrap os.ErrNotExist, got: %v", err)
	}

	if !strings.Contains(err.Error(), "opening config file") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "opening config file")
	}
}

func TestFileProvider_Get_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(cfgPath, []byte(`{not valid json`), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	_, err := fp.Get()

	if err == nil {
		t.Fatal("Get() expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "parsing config JSON") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "parsing config JSON")
	}
}

func TestFileProvider_Get_EmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "empty.json")

	// An empty file is not valid JSON -- json.Unmarshal returns an error
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	_, err := fp.Get()

	if err == nil {
		t.Fatal("Get() expected error for empty file, got nil")
	}

	if !strings.Contains(err.Error(), "parsing config JSON") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "parsing config JSON")
	}
}

func TestFileProvider_Get_EmptyJSONObject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "empty-obj.json")

	// An empty JSON object is valid -- should produce a zero-value Config
	if err := os.WriteFile(cfgPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()

	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	// All string fields should be zero values
	if got.Environment != "" {
		t.Errorf("Environment = %q, want empty", got.Environment)
	}
	if got.Port != "" {
		t.Errorf("Port = %q, want empty", got.Port)
	}
	if string(got.Db_Driver) != "" {
		t.Errorf("Db_Driver = %q, want empty", got.Db_Driver)
	}
}

func TestFileProvider_Get_RoundTripDefaultConfig(t *testing.T) {
	t.Parallel()

	// Serialize DefaultConfig to JSON, write to file, read back via FileProvider,
	// and verify key fields survive the round trip.
	original := config.DefaultConfig()
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal DefaultConfig: %v", err)
	}

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "roundtrip.json")

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "Environment", got: got.Environment, want: original.Environment},
		{name: "Port", got: got.Port, want: original.Port},
		{name: "SSL_Port", got: got.SSL_Port, want: original.SSL_Port},
		{name: "Db_Driver", got: string(got.Db_Driver), want: string(original.Db_Driver)},
		{name: "Db_Name", got: got.Db_Name, want: original.Db_Name},
		{name: "Db_URL", got: got.Db_URL, want: original.Db_URL},
		{name: "SSH_Host", got: got.SSH_Host, want: original.SSH_Host},
		{name: "SSH_Port", got: got.SSH_Port, want: original.SSH_Port},
		{name: "Cookie_Name", got: got.Cookie_Name, want: original.Cookie_Name},
		{name: "Cookie_Duration", got: got.Cookie_Duration, want: original.Cookie_Duration},
		{name: "Cookie_SameSite", got: got.Cookie_SameSite, want: original.Cookie_SameSite},
		{name: "Bucket_Region", got: got.Bucket_Region, want: original.Bucket_Region},
		{name: "Auth_Salt", got: got.Auth_Salt, want: original.Auth_Salt},
		{name: "Node_ID", got: got.Node_ID, want: original.Node_ID},
		{name: "Update_Channel", got: got.Update_Channel, want: original.Update_Channel},
		{name: "Observability_Provider", got: got.Observability_Provider, want: original.Observability_Provider},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("round-trip %s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	// Verify boolean fields
	if got.Cors_Credentials != original.Cors_Credentials {
		t.Errorf("round-trip Cors_Credentials = %v, want %v", got.Cors_Credentials, original.Cors_Credentials)
	}
	if got.Bucket_Force_Path_Style != original.Bucket_Force_Path_Style {
		t.Errorf("round-trip Bucket_Force_Path_Style = %v, want %v", got.Bucket_Force_Path_Style, original.Bucket_Force_Path_Style)
	}
}

func TestFileProvider_Get_PartialConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "partial.json")

	// Only set a few fields -- the rest should be Go zero values
	partialJSON := `{
		"environment": "staging",
		"db_driver": "mysql"
	}`

	if err := os.WriteFile(cfgPath, []byte(partialJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	// Explicitly set fields
	if got.Environment != "staging" {
		t.Errorf("Environment = %q, want %q", got.Environment, "staging")
	}
	if string(got.Db_Driver) != "mysql" {
		t.Errorf("Db_Driver = %q, want %q", got.Db_Driver, "mysql")
	}

	// Fields not in the JSON should be zero values
	if got.Port != "" {
		t.Errorf("Port = %q, want empty (zero value)", got.Port)
	}
	if got.SSH_Host != "" {
		t.Errorf("SSH_Host = %q, want empty (zero value)", got.SSH_Host)
	}
	if got.Cookie_Secure != false {
		t.Errorf("Cookie_Secure = %v, want false (zero value)", got.Cookie_Secure)
	}
	if got.Observability_Sample_Rate != 0 {
		t.Errorf("Observability_Sample_Rate = %f, want 0 (zero value)", got.Observability_Sample_Rate)
	}
}

func TestFileProvider_Get_UnknownFieldsIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "extra-fields.json")

	// JSON with fields not in the Config struct -- json.Unmarshal ignores them
	extraJSON := `{
		"environment": "test",
		"totally_unknown_field": "should be ignored",
		"another_fake": 12345
	}`

	if err := os.WriteFile(cfgPath, []byte(extraJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if got.Environment != "test" {
		t.Errorf("Environment = %q, want %q", got.Environment, "test")
	}
}

func TestFileProvider_Get_PermissionDenied(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "no-read.json")

	if err := os.WriteFile(cfgPath, []byte(`{"environment":"test"}`), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Remove read permission
	if err := os.Chmod(cfgPath, 0000); err != nil {
		t.Fatalf("failed to chmod test file: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so t.TempDir cleanup can remove the file
		os.Chmod(cfgPath, 0644) //nolint:errcheck // cleanup best-effort
	})

	fp := config.NewFileProvider(cfgPath)
	_, err := fp.Get()

	if err == nil {
		t.Fatal("Get() expected error for unreadable file, got nil")
	}

	if !strings.Contains(err.Error(), "opening config file") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "opening config file")
	}

	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("expected error to wrap os.ErrPermission, got: %v", err)
	}
}

func TestFileProvider_Get_NestedObjects(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nested.json")

	// Config with map and slice fields populated
	nestedJSON := `{
		"environment": "production",
		"environment_hosts": {
			"production": "prod.example.com",
			"staging": "staging.example.com"
		},
		"cors_origins": ["https://app.example.com", "https://admin.example.com"],
		"cors_methods": ["GET", "POST"],
		"oauth_endpoint": {
			"oauth_auth_url": "https://auth.example.com/authorize",
			"oauth_token_url": "https://auth.example.com/token",
			"oauth_userinfo_url": "https://auth.example.com/userinfo"
		},
		"observability_tags": {
			"service": "modulacms",
			"team": "platform"
		}
	}`

	if err := os.WriteFile(cfgPath, []byte(nestedJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	// Environment hosts
	if got.Environment_Hosts == nil {
		t.Fatal("Environment_Hosts is nil")
	}
	if got.Environment_Hosts["production"] != "prod.example.com" {
		t.Errorf("Environment_Hosts[production] = %q, want %q", got.Environment_Hosts["production"], "prod.example.com")
	}
	if got.Environment_Hosts["staging"] != "staging.example.com" {
		t.Errorf("Environment_Hosts[staging] = %q, want %q", got.Environment_Hosts["staging"], "staging.example.com")
	}

	// CORS origins
	if len(got.Cors_Origins) != 2 {
		t.Fatalf("Cors_Origins length = %d, want 2", len(got.Cors_Origins))
	}
	if got.Cors_Origins[0] != "https://app.example.com" {
		t.Errorf("Cors_Origins[0] = %q, want %q", got.Cors_Origins[0], "https://app.example.com")
	}
	if got.Cors_Origins[1] != "https://admin.example.com" {
		t.Errorf("Cors_Origins[1] = %q, want %q", got.Cors_Origins[1], "https://admin.example.com")
	}

	// OAuth endpoints
	if got.Oauth_Endpoint == nil {
		t.Fatal("Oauth_Endpoint is nil")
	}
	if got.Oauth_Endpoint[config.OauthAuthURL] != "https://auth.example.com/authorize" {
		t.Errorf("Oauth_Endpoint[OauthAuthURL] = %q, want %q", got.Oauth_Endpoint[config.OauthAuthURL], "https://auth.example.com/authorize")
	}

	// Observability tags
	if got.Observability_Tags == nil {
		t.Fatal("Observability_Tags is nil")
	}
	if got.Observability_Tags["service"] != "modulacms" {
		t.Errorf("Observability_Tags[service] = %q, want %q", got.Observability_Tags["service"], "modulacms")
	}
}

func TestFileProvider_Get_BooleanFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bools.json")

	boolJSON := `{
		"cookie_secure": true,
		"cors_credentials": false,
		"bucket_force_path_style": false,
		"update_auto_enabled": true,
		"observability_enabled": true,
		"observability_send_pii": true,
		"plugin_enabled": true
	}`

	if err := os.WriteFile(cfgPath, []byte(boolJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "Cookie_Secure", got: got.Cookie_Secure, want: true},
		{name: "Cors_Credentials", got: got.Cors_Credentials, want: false},
		{name: "Bucket_Force_Path_Style", got: got.Bucket_Force_Path_Style, want: false},
		{name: "Update_Auto_Enabled", got: got.Update_Auto_Enabled, want: true},
		{name: "Observability_Enabled", got: got.Observability_Enabled, want: true},
		{name: "Observability_Send_PII", got: got.Observability_Send_PII, want: true},
		{name: "Plugin_Enabled", got: got.Plugin_Enabled, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestFileProvider_Get_NumericFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nums.json")

	numJSON := `{
		"observability_sample_rate": 0.5,
		"observability_traces_rate": 0.25,
		"plugin_max_vms": 8,
		"plugin_timeout": 10,
		"plugin_max_ops": 5000
	}`

	if err := os.WriteFile(cfgPath, []byte(numJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	got, err := fp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if got.Observability_Sample_Rate != 0.5 {
		t.Errorf("Observability_Sample_Rate = %f, want 0.5", got.Observability_Sample_Rate)
	}
	if got.Observability_Traces_Rate != 0.25 {
		t.Errorf("Observability_Traces_Rate = %f, want 0.25", got.Observability_Traces_Rate)
	}
	if got.Plugin_Max_VMs != 8 {
		t.Errorf("Plugin_Max_VMs = %d, want 8", got.Plugin_Max_VMs)
	}
	if got.Plugin_Timeout != 10 {
		t.Errorf("Plugin_Timeout = %d, want 10", got.Plugin_Timeout)
	}
	if got.Plugin_Max_Ops != 5000 {
		t.Errorf("Plugin_Max_Ops = %d, want 5000", got.Plugin_Max_Ops)
	}
}

func TestFileProvider_Get_TypeMismatchInJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "type-mismatch.json")

	// port expects a string, but we give it a number -- json.Unmarshal will fail
	badTypeJSON := `{
		"port": 8080
	}`

	if err := os.WriteFile(cfgPath, []byte(badTypeJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)
	_, err := fp.Get()

	if err == nil {
		t.Fatal("Get() expected error for type mismatch, got nil")
	}

	if !strings.Contains(err.Error(), "parsing config JSON") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "parsing config JSON")
	}
}

func TestFileProvider_Get_CalledMultipleTimes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "multi.json")

	initialJSON := `{"environment": "first"}`
	if err := os.WriteFile(cfgPath, []byte(initialJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)

	// First call
	got1, err := fp.Get()
	if err != nil {
		t.Fatalf("first Get() unexpected error: %v", err)
	}
	if got1.Environment != "first" {
		t.Errorf("first Get() Environment = %q, want %q", got1.Environment, "first")
	}

	// Overwrite the file with different content
	updatedJSON := `{"environment": "second"}`
	if err := os.WriteFile(cfgPath, []byte(updatedJSON), 0644); err != nil {
		t.Fatalf("failed to overwrite test file: %v", err)
	}

	// Second call should read the updated file (FileProvider has no caching)
	got2, err := fp.Get()
	if err != nil {
		t.Fatalf("second Get() unexpected error: %v", err)
	}
	if got2.Environment != "second" {
		t.Errorf("second Get() Environment = %q, want %q", got2.Environment, "second")
	}
}

func TestFileProvider_Get_ReturnsDistinctPointers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "distinct.json")

	if err := os.WriteFile(cfgPath, []byte(`{"environment": "test"}`), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fp := config.NewFileProvider(cfgPath)

	cfg1, err := fp.Get()
	if err != nil {
		t.Fatalf("first Get() error: %v", err)
	}

	cfg2, err := fp.Get()
	if err != nil {
		t.Fatalf("second Get() error: %v", err)
	}

	// Each call should return a new Config allocation -- mutating one
	// must not affect the other.
	cfg1.Environment = "mutated"
	if cfg2.Environment == "mutated" {
		t.Error("Get() returned the same pointer on two calls -- mutations leak between callers")
	}
}
