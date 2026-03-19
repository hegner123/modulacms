package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hegner123/modulacms/internal/utility"
)

// LayeredFileProvider reads a base config file and an overlay config file,
// merging them at load time. The overlay's keys overwrite the base's keys;
// absent keys in the overlay leave the base values intact.
//
// Save operations target the overlay file so that operator overrides are
// written to the correct layer.
type LayeredFileProvider struct {
	base    *FileProvider
	overlay *FileProvider
}

// NewLayeredFileProvider creates a provider that merges basePath and overlayPath
// at load time. basePath is the full config; overlayPath contains only the
// fields that differ per environment.
func NewLayeredFileProvider(basePath, overlayPath string) *LayeredFileProvider {
	return &LayeredFileProvider{
		base:    NewFileProvider(basePath),
		overlay: NewFileProvider(overlayPath),
	}
}

// Get reads both files, marshals each to map[string]any, merges them with
// MergeMaps, and unmarshals the result to *Config.
func (lp *LayeredFileProvider) Get() (*Config, error) {
	baseMap, err := readJSONMap(lp.base.Path())
	if err != nil {
		return nil, fmt.Errorf("reading base config: %w", err)
	}

	overlayMap, err := readJSONMap(lp.overlay.Path())
	if err != nil {
		return nil, fmt.Errorf("reading overlay config: %w", err)
	}

	merged := MergeMaps(baseMap, overlayMap)

	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshaling merged config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(mergedBytes, &cfg); err != nil {
		return nil, fmt.Errorf("parsing merged config: %w", err)
	}

	return &cfg, nil
}

// Save persists the configuration to the overlay file. When layered, the
// overlay is the write target because operator intent is to override.
func (lp *LayeredFileProvider) Save(c *Config) error {
	return lp.overlay.Save(c)
}

// SaveToBase persists the configuration to the base file.
func (lp *LayeredFileProvider) SaveToBase(c *Config) error {
	return lp.base.Save(c)
}

// Path returns the overlay file path (the primary write target).
func (lp *LayeredFileProvider) Path() string {
	return lp.overlay.Path()
}

// BasePath returns the base config file path.
func (lp *LayeredFileProvider) BasePath() string {
	return lp.base.Path()
}

// OverlayPath returns the overlay config file path.
func (lp *LayeredFileProvider) OverlayPath() string {
	return lp.overlay.Path()
}

// readJSONMap reads a JSON file and returns its contents as map[string]any.
func readJSONMap(path string) (map[string]any, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filepath.Base(path), err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filepath.Base(path), err)
	}

	var m map[string]any
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
	}

	return m, nil
}
