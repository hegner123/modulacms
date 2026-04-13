package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hegner123/modulacms/internal/install"
)

func TestCreateDefaultConfig(t *testing.T) {
	t.Parallel()

	t.Run("creates config file at path", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "modula.config.json")

		err := install.CreateDefaultConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("config file not created: %v", statErr)
		}
		if info.Size() == 0 {
			t.Fatal("config file is empty")
		}
	})

	t.Run("output is valid JSON", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "modula.config.json")

		err := install.CreateDefaultConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("failed to read config: %v", readErr)
		}

		var parsed map[string]any
		if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
			t.Fatalf("invalid JSON: %v", jsonErr)
		}

		// Verify key fields exist
		if _, ok := parsed["port"]; !ok {
			t.Error("missing 'port' field in config")
		}
		if _, ok := parsed["db_driver"]; !ok {
			t.Error("missing 'db_driver' field in config")
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "modula.config.json")

		// Write dummy content first
		if err := os.WriteFile(path, []byte("old content"), 0644); err != nil {
			t.Fatal(err)
		}

		err := install.CreateDefaultConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("failed to read config: %v", readErr)
		}

		if string(data) == "old content" {
			t.Error("file was not overwritten")
		}

		// Verify new content is valid JSON
		var parsed map[string]any
		if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
			t.Fatalf("overwritten file is not valid JSON: %v", jsonErr)
		}
	})

	t.Run("returns error for unwritable path", func(t *testing.T) {
		t.Parallel()
		err := install.CreateDefaultConfig("/nonexistent/dir/config.json")
		if err == nil {
			t.Fatal("expected error for unwritable path, got nil")
		}
	})
}
