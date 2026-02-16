package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hegner123/modulacms/internal/utility"
)

// Saver defines an interface for persisting configuration to storage.
type Saver interface {
	Save(c *Config) error
}

// FileProvider loads configuration from a JSON file
type FileProvider struct {
	path string
}

// NewFileProvider creates a new file-based configuration provider
func NewFileProvider(path string) *FileProvider {
	if path == "" {
		path = "config.json"
	}
	return &FileProvider{path: path}
}

// Get implements the Provider interface
func (fp *FileProvider) Get() (*Config, error) {
	file, err := os.Open(fp.path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("parsing config JSON: %w", err)
	}

	return &config, nil
}

// Save persists the configuration to the file atomically.
// It writes to a temporary file first, then renames to avoid partial writes.
func (fp *FileProvider) Save(c *Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(fp.path)
	tmp, err := os.CreateTemp(dir, "config-*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temp config file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		// Write failed — attempt cleanup of the temp file.
		tmp.Close() // Close error is secondary to the write failure.
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp config file: %w", err)
	}

	if err := os.Rename(tmpPath, fp.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp config to %s: %w", fp.path, err)
	}

	return nil
}

// Path returns the file path this provider reads from and writes to.
func (fp *FileProvider) Path() string {
	return fp.path
}
