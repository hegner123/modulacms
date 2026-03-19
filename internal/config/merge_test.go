package config_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

func TestMergeMaps_EmptyOverlay(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"port":        ":8080",
		"environment": "dev",
	}
	overlay := map[string]any{}

	merged := config.MergeMaps(base, overlay)

	if merged["port"] != ":8080" {
		t.Errorf("port = %v, want %q", merged["port"], ":8080")
	}
	if merged["environment"] != "dev" {
		t.Errorf("environment = %v, want %q", merged["environment"], "dev")
	}
	if len(merged) != 2 {
		t.Errorf("len(merged) = %d, want 2", len(merged))
	}
}

func TestMergeMaps_OverlayOverwritesBase(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"port":        ":8080",
		"environment": "dev",
		"db_driver":   "sqlite",
	}
	overlay := map[string]any{
		"environment": "prod",
		"db_driver":   "postgres",
	}

	merged := config.MergeMaps(base, overlay)

	if merged["port"] != ":8080" {
		t.Errorf("port = %v, want %q (preserved from base)", merged["port"], ":8080")
	}
	if merged["environment"] != "prod" {
		t.Errorf("environment = %v, want %q (overwritten by overlay)", merged["environment"], "prod")
	}
	if merged["db_driver"] != "postgres" {
		t.Errorf("db_driver = %v, want %q (overwritten by overlay)", merged["db_driver"], "postgres")
	}
}

func TestMergeMaps_BaseKeysPreserved(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"port":     ":8080",
		"ssh_port": "2233",
		"db_url":   "local.db",
	}
	overlay := map[string]any{
		"port": ":9090",
	}

	merged := config.MergeMaps(base, overlay)

	if merged["port"] != ":9090" {
		t.Errorf("port = %v, want %q", merged["port"], ":9090")
	}
	if merged["ssh_port"] != "2233" {
		t.Errorf("ssh_port = %v, want %q (should be preserved)", merged["ssh_port"], "2233")
	}
	if merged["db_url"] != "local.db" {
		t.Errorf("db_url = %v, want %q (should be preserved)", merged["db_url"], "local.db")
	}
}

func TestMergeMaps_OverlayAddsNewKeys(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"port": ":8080",
	}
	overlay := map[string]any{
		"bucket_secret_key": "s3cr3t",
		"db_password":       "hunter2",
	}

	merged := config.MergeMaps(base, overlay)

	if merged["port"] != ":8080" {
		t.Errorf("port = %v, want %q", merged["port"], ":8080")
	}
	if merged["bucket_secret_key"] != "s3cr3t" {
		t.Errorf("bucket_secret_key = %v, want %q", merged["bucket_secret_key"], "s3cr3t")
	}
	if merged["db_password"] != "hunter2" {
		t.Errorf("db_password = %v, want %q", merged["db_password"], "hunter2")
	}
	if len(merged) != 3 {
		t.Errorf("len(merged) = %d, want 3", len(merged))
	}
}

func TestMergeMaps_BothEmpty(t *testing.T) {
	t.Parallel()

	merged := config.MergeMaps(map[string]any{}, map[string]any{})

	if len(merged) != 0 {
		t.Errorf("len(merged) = %d, want 0", len(merged))
	}
}

func TestMergeMaps_DoesNotMutateInputs(t *testing.T) {
	t.Parallel()

	base := map[string]any{"a": "1", "b": "2"}
	overlay := map[string]any{"b": "3", "c": "4"}

	config.MergeMaps(base, overlay)

	// base should be unchanged
	if base["a"] != "1" || base["b"] != "2" || len(base) != 2 {
		t.Errorf("base was mutated: %v", base)
	}
	// overlay should be unchanged
	if overlay["b"] != "3" || overlay["c"] != "4" || len(overlay) != 2 {
		t.Errorf("overlay was mutated: %v", overlay)
	}
}

func TestMergeMaps_SlicesReplacedEntirely(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"cors_origins": []any{"http://localhost:3000", "http://localhost:4000"},
	}
	overlay := map[string]any{
		"cors_origins": []any{"https://prod.example.com"},
	}

	merged := config.MergeMaps(base, overlay)

	origins, ok := merged["cors_origins"].([]any)
	if !ok {
		t.Fatalf("cors_origins type = %T, want []any", merged["cors_origins"])
	}
	if len(origins) != 1 {
		t.Fatalf("len(cors_origins) = %d, want 1 (replaced, not appended)", len(origins))
	}
	if origins[0] != "https://prod.example.com" {
		t.Errorf("cors_origins[0] = %v, want %q", origins[0], "https://prod.example.com")
	}
}

func TestMergeMaps_MapsReplacedEntirely(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"environment_hosts": map[string]any{
			"dev":     "localhost",
			"staging": "staging.example.com",
		},
	}
	overlay := map[string]any{
		"environment_hosts": map[string]any{
			"prod": "prod.example.com",
		},
	}

	merged := config.MergeMaps(base, overlay)

	hosts, ok := merged["environment_hosts"].(map[string]any)
	if !ok {
		t.Fatalf("environment_hosts type = %T, want map[string]any", merged["environment_hosts"])
	}
	// Only the overlay map should be present (shallow merge = replace)
	if len(hosts) != 1 {
		t.Fatalf("len(environment_hosts) = %d, want 1 (replaced, not deep-merged)", len(hosts))
	}
	if hosts["prod"] != "prod.example.com" {
		t.Errorf("environment_hosts[prod] = %v, want %q", hosts["prod"], "prod.example.com")
	}
}
