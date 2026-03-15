// Package deploy provides the sync engine for exporting and importing
// Modula content data between instances via JSON files or the deploy API.
package deploy

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// gzipThreshold is the uncompressed JSON size above which files are gzip-compressed.
// 1 GB in bytes.
const gzipThreshold = 1 << 30

// ExportToFile exports data from the driver, marshals it to JSON, and writes it to outPath.
// If the serialized JSON exceeds gzipThreshold (1 GB), the file is gzip-compressed and
// ".gz" is appended to the path. Returns the manifest and the actual path written.
func ExportToFile(ctx context.Context, driver db.DbDriver, opts ExportOptions, outPath string) (*SyncManifest, string, error) {
	payload, err := ExportPayload(ctx, driver, opts)
	if err != nil {
		return nil, "", fmt.Errorf("export payload: %w", err)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("marshal payload: %w", err)
	}

	actualPath := outPath
	if len(data) > gzipThreshold {
		if !strings.HasSuffix(actualPath, ".gz") {
			actualPath = actualPath + ".gz"
		}
		if err := writeGzipFile(actualPath, data); err != nil {
			return nil, "", fmt.Errorf("write gzip export file: %w", err)
		}
	} else {
		if err := os.WriteFile(actualPath, data, 0640); err != nil {
			return nil, "", fmt.Errorf("write export file: %w", err)
		}
	}

	return &payload.Manifest, actualPath, nil
}

// ImportFromFile reads a JSON export file and imports it into the target database.
// Automatically detects gzip-compressed files by extension (.gz) or magic bytes.
func ImportFromFile(ctx context.Context, cfg config.Config, driver db.DbDriver, inPath string, skipBackup bool) (*SyncResult, error) {
	data, err := ReadPayloadFile(inPath)
	if err != nil {
		return nil, fmt.Errorf("read import file: %w", err)
	}

	var payload SyncPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal import file: %w", err)
	}

	return ImportPayload(ctx, cfg, driver, &payload, skipBackup)
}

// Pull exports data from a remote Modula instance and imports it into the local database.
// The environment name is resolved from cfg.Deploy_Environments.
// If dryRun is true, the payload is validated locally without modifying the database.
func Pull(ctx context.Context, cfg config.Config, driver db.DbDriver, envName string, opts ExportOptions, skipBackup bool, dryRun bool) (*SyncResult, error) {
	ctx, cancel := resolveContext(ctx)
	defer cancel()

	client, err := clientForEnv(cfg, envName)
	if err != nil {
		return nil, err
	}

	// Verify remote is reachable.
	if _, err := client.Health(ctx); err != nil {
		return nil, fmt.Errorf("remote health check failed: %w", err)
	}

	// Export from remote.
	payload, err := client.Export(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("remote export: %w", err)
	}

	if dryRun {
		return BuildDryRunResult(payload, driver), nil
	}

	// Import locally.
	return ImportPayload(ctx, cfg, driver, payload, skipBackup)
}

// Push exports data from the local database and sends it to a remote Modula instance.
// The environment name is resolved from cfg.Deploy_Environments.
// If dryRun is true, the payload is validated on the remote without modifying its database.
func Push(ctx context.Context, cfg config.Config, driver db.DbDriver, envName string, opts ExportOptions, dryRun bool) (*SyncResult, error) {
	ctx, cancel := resolveContext(ctx)
	defer cancel()

	client, err := clientForEnv(cfg, envName)
	if err != nil {
		return nil, err
	}

	// Verify remote is reachable.
	if _, err := client.Health(ctx); err != nil {
		return nil, fmt.Errorf("remote health check failed: %w", err)
	}

	// Export locally.
	payload, err := ExportPayload(ctx, driver, opts)
	if err != nil {
		return nil, fmt.Errorf("local export: %w", err)
	}

	if dryRun {
		return client.DryRunImport(ctx, payload)
	}

	// Import on remote.
	return client.Import(ctx, payload)
}

// TestEnvConnection tests connectivity and authentication to a configured deploy environment.
// Returns the HealthResponse on success or an error describing the failure.
func TestEnvConnection(ctx context.Context, cfg config.Config, envName string) (*HealthResponse, error) {
	client, err := clientForEnv(cfg, envName)
	if err != nil {
		return nil, err
	}

	return client.Health(ctx)
}

// BuildDryRunResult validates the payload and returns a SyncResult with the impact report.
func BuildDryRunResult(payload *SyncPayload, driver db.DbDriver) *SyncResult {
	validationErrs := ValidatePayload(payload, driver)

	tablesAffected := make([]string, 0, len(payload.Tables))
	rowCounts := make(map[string]int, len(payload.Tables))
	for name, td := range payload.Tables {
		tablesAffected = append(tablesAffected, name)
		rowCounts[name] = len(td.Rows)
	}

	// Check for users that would need placeholders.
	var warnings []string
	if driver != nil {
		users, uErr := driver.ListUsers()
		if uErr == nil && users != nil {
			existingUsers := make(map[string]bool, len(*users))
			for _, u := range *users {
				existingUsers[u.UserID.String()] = true
			}
			missingCount := 0
			for uid := range payload.UserRefs {
				if !existingUsers[uid] {
					missingCount++
				}
			}
			if missingCount > 0 {
				warnings = append(warnings, fmt.Sprintf("Would create %d placeholder user(s)", missingCount))
			}
		}
	}

	return &SyncResult{
		Success:        len(validationErrs) == 0,
		DryRun:         true,
		Strategy:       StrategyOverwrite,
		TablesAffected: tablesAffected,
		RowCounts:      rowCounts,
		Errors:         validationErrs,
		Warnings:       warnings,
	}
}

// clientForEnv looks up the named environment in cfg.Deploy_Environments and returns a DeployClient.
func clientForEnv(cfg config.Config, envName string) (*DeployClient, error) {
	for _, env := range cfg.Deploy_Environments {
		if env.Name == envName {
			if env.URL == "" {
				return nil, fmt.Errorf("environment %q has no URL configured", envName)
			}
			if env.APIKey == "" {
				return nil, fmt.Errorf("environment %q has no api_key configured", envName)
			}
			return NewDeployClient(env.URL, env.APIKey), nil
		}
	}
	return nil, fmt.Errorf("unknown deploy environment %q", envName)
}

// writeGzipFile compresses data with gzip and writes to path.
func writeGzipFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}

	gw, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		f.Close()
		return err
	}

	if _, err := gw.Write(data); err != nil {
		gw.Close()
		f.Close()
		return err
	}

	if err := gw.Close(); err != nil {
		f.Close()
		return err
	}

	return f.Close()
}

// ReadPayloadFile reads a file, decompressing gzip if the file has a .gz extension
// or starts with the gzip magic bytes (0x1f 0x8b).
func ReadPayloadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if isGzipped(path, data) {
		return decompressGzip(data)
	}

	return data, nil
}

// isGzipped returns true if the file path ends with .gz or the data starts with gzip magic bytes.
func isGzipped(path string, data []byte) bool {
	if strings.HasSuffix(path, ".gz") {
		return true
	}
	// gzip magic bytes: 0x1f 0x8b
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// decompressGzip decompresses gzip-encoded data.
func decompressGzip(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	return io.ReadAll(gr)
}
