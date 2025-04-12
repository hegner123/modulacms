package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/hegner123/modulacms/internal/utility"
)

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
