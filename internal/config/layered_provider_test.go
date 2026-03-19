package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// Compile-time checks: *LayeredFileProvider must satisfy Provider and Saver.
var _ config.Provider = (*config.LayeredFileProvider)(nil)
var _ config.Saver = (*config.LayeredFileProvider)(nil)

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal JSON for %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func TestLayeredFileProvider_MergedResult(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "base.json")
	overlayPath := filepath.Join(dir, "overlay.json")

	writeJSON(t, basePath, map[string]any{
		"environment": "dev",
		"port":        ":8080",
		"ssh_port":    "2233",
		"db_driver":   "sqlite",
		"db_url":      "local.db",
	})
	writeJSON(t, overlayPath, map[string]any{
		"environment": "prod",
		"db_driver":   "postgres",
		"db_url":      "postgres://prod:5432/cms",
	})

	lp := config.NewLayeredFileProvider(basePath, overlayPath)
	cfg, err := lp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	// Overridden by overlay
	if cfg.Environment != "prod" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "prod")
	}
	if string(cfg.Db_Driver) != "postgres" {
		t.Errorf("Db_Driver = %q, want %q", cfg.Db_Driver, "postgres")
	}
	if cfg.Db_URL != "postgres://prod:5432/cms" {
		t.Errorf("Db_URL = %q, want %q", cfg.Db_URL, "postgres://prod:5432/cms")
	}

	// Preserved from base
	if cfg.Port != ":8080" {
		t.Errorf("Port = %q, want %q (from base)", cfg.Port, ":8080")
	}
	if cfg.SSH_Port != "2233" {
		t.Errorf("SSH_Port = %q, want %q (from base)", cfg.SSH_Port, "2233")
	}
}

func TestLayeredFileProvider_EmptyOverlay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "base.json")
	overlayPath := filepath.Join(dir, "overlay.json")

	writeJSON(t, basePath, map[string]any{
		"environment": "dev",
		"port":        ":8080",
		"db_driver":   "sqlite",
	})
	writeJSON(t, overlayPath, map[string]any{})

	lp := config.NewLayeredFileProvider(basePath, overlayPath)
	cfg, err := lp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if cfg.Environment != "dev" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "dev")
	}
	if cfg.Port != ":8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, ":8080")
	}
	if string(cfg.Db_Driver) != "sqlite" {
		t.Errorf("Db_Driver = %q, want %q", cfg.Db_Driver, "sqlite")
	}
}

func TestLayeredFileProvider_MissingOverlay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "base.json")
	overlayPath := filepath.Join(dir, "does-not-exist.json")

	writeJSON(t, basePath, map[string]any{"environment": "dev"})

	lp := config.NewLayeredFileProvider(basePath, overlayPath)
	_, err := lp.Get()

	if err == nil {
		t.Fatal("Get() expected error for missing overlay, got nil")
	}
	if !strings.Contains(err.Error(), "reading overlay config") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "reading overlay config")
	}
}

func TestLayeredFileProvider_MissingBase(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "does-not-exist.json")
	overlayPath := filepath.Join(dir, "overlay.json")

	writeJSON(t, overlayPath, map[string]any{"environment": "prod"})

	lp := config.NewLayeredFileProvider(basePath, overlayPath)
	_, err := lp.Get()

	if err == nil {
		t.Fatal("Get() expected error for missing base, got nil")
	}
	if !strings.Contains(err.Error(), "reading base config") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "reading base config")
	}
}

func TestLayeredFileProvider_SaveWritesToOverlay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "base.json")
	overlayPath := filepath.Join(dir, "overlay.json")

	writeJSON(t, basePath, map[string]any{"environment": "dev", "port": ":8080"})
	writeJSON(t, overlayPath, map[string]any{"environment": "prod"})

	lp := config.NewLayeredFileProvider(basePath, overlayPath)

	// Save a full config — it should write to the overlay path
	saveCfg := &config.Config{
		Environment: "staging",
		Port:        ":9090",
	}
	if err := lp.Save(saveCfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read back the overlay file directly
	data, err := os.ReadFile(overlayPath)
	if err != nil {
		t.Fatalf("reading overlay after save: %v", err)
	}
	if !strings.Contains(string(data), "staging") {
		t.Errorf("overlay file should contain 'staging', got: %s", string(data))
	}

	// Base should be unchanged
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatalf("reading base after save: %v", err)
	}
	if strings.Contains(string(baseData), "staging") {
		t.Error("base file should NOT contain 'staging' — save should target overlay only")
	}
}

func TestLayeredFileProvider_PathReturnsOverlay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "base.json")
	overlayPath := filepath.Join(dir, "overlay.json")

	lp := config.NewLayeredFileProvider(basePath, overlayPath)

	if lp.Path() != overlayPath {
		t.Errorf("Path() = %q, want %q", lp.Path(), overlayPath)
	}
	if lp.BasePath() != basePath {
		t.Errorf("BasePath() = %q, want %q", lp.BasePath(), basePath)
	}
	if lp.OverlayPath() != overlayPath {
		t.Errorf("OverlayPath() = %q, want %q", lp.OverlayPath(), overlayPath)
	}
}

func TestLayeredFileProvider_OverlayOnlyKeys(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "base.json")
	overlayPath := filepath.Join(dir, "overlay.json")

	// Base has no bucket_secret_key
	writeJSON(t, basePath, map[string]any{
		"environment": "dev",
		"port":        ":8080",
	})
	// Overlay adds a key not in base
	writeJSON(t, overlayPath, map[string]any{
		"bucket_secret_key": "s3cr3t",
	})

	lp := config.NewLayeredFileProvider(basePath, overlayPath)
	cfg, err := lp.Get()
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if cfg.Environment != "dev" {
		t.Errorf("Environment = %q, want %q (from base)", cfg.Environment, "dev")
	}
	if cfg.Port != ":8080" {
		t.Errorf("Port = %q, want %q (from base)", cfg.Port, ":8080")
	}
	if cfg.Bucket_Secret_Key != "s3cr3t" {
		t.Errorf("Bucket_Secret_Key = %q, want %q (from overlay)", cfg.Bucket_Secret_Key, "s3cr3t")
	}
}

func TestLayeredFileProvider_ManagerIntegration(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	basePath := filepath.Join(dir, "base.json")
	overlayPath := filepath.Join(dir, "overlay.json")

	writeJSON(t, basePath, map[string]any{
		"environment": "dev",
		"port":        ":8080",
		"db_driver":   "sqlite",
		"db_url":      "local.db",
	})
	writeJSON(t, overlayPath, map[string]any{
		"environment": "prod",
		"db_driver":   "postgres",
	})

	lp := config.NewLayeredFileProvider(basePath, overlayPath)
	mgr := config.NewManager(lp)

	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cfg, err := mgr.Config()
	if err != nil {
		t.Fatalf("Config() error: %v", err)
	}

	if cfg.Environment != "prod" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "prod")
	}
	if cfg.Port != ":8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, ":8080")
	}
	if string(cfg.Db_Driver) != "postgres" {
		t.Errorf("Db_Driver = %q, want %q", cfg.Db_Driver, "postgres")
	}
}
