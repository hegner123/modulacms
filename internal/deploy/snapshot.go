package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// SnapshotInfo describes a saved snapshot without loading its full data.
type SnapshotInfo struct {
	ID        string         `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Tables    []string       `json:"tables"`
	RowCounts map[string]int `json:"row_counts"`
	SizeBytes int64          `json:"size_bytes"`
}

// snapshotPrefix is used for snapshot filenames.
const snapshotPrefix = "snap_"

// SaveSnapshot writes the payload to a JSON file in dir and returns the snapshot ID.
// If the serialized JSON exceeds gzipThreshold (1 GB), the file is gzip-compressed
// with a .json.gz extension.
func SaveSnapshot(dir string, payload *SyncPayload) (string, error) {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("create snapshot directory: %w", err)
	}

	id := snapshotPrefix + time.Now().UTC().Format("20060102_150405")

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal snapshot: %w", err)
	}

	if len(data) > gzipThreshold {
		path := filepath.Join(dir, id+".json.gz")
		if err := writeGzipFile(path, data); err != nil {
			return "", fmt.Errorf("write gzip snapshot: %w", err)
		}
	} else {
		path := filepath.Join(dir, id+".json")
		if err := os.WriteFile(path, data, 0640); err != nil {
			return "", fmt.Errorf("write snapshot: %w", err)
		}
	}

	return id, nil
}

// ListSnapshots reads the snapshot directory and returns metadata for each snapshot.
func ListSnapshots(dir string) ([]SnapshotInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read snapshot directory: %w", err)
	}

	var snapshots []SnapshotInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, snapshotPrefix) {
			continue
		}

		var id string
		switch {
		case strings.HasSuffix(name, ".json.gz"):
			id = strings.TrimSuffix(name, ".json.gz")
		case strings.HasSuffix(name, ".json"):
			id = strings.TrimSuffix(name, ".json")
		default:
			continue
		}

		ts := parseSnapshotTimestamp(id)

		info, fErr := entry.Info()
		if fErr != nil {
			continue
		}

		// Read manifest from the file to get table/row info.
		path := filepath.Join(dir, name)
		payload, lErr := loadSnapshotFile(path)
		if lErr != nil {
			continue
		}

		snapshots = append(snapshots, SnapshotInfo{
			ID:        id,
			Timestamp: ts,
			Tables:    payload.Manifest.Tables,
			RowCounts: payload.Manifest.RowCounts,
			SizeBytes: info.Size(),
		})
	}

	// Sort by timestamp descending (newest first).
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	return snapshots, nil
}

// LoadSnapshot reads and decodes a snapshot file by ID from the given directory.
// Tries .json.gz first, then .json.
func LoadSnapshot(dir, id string) (*SyncPayload, error) {
	gzPath := filepath.Join(dir, id+".json.gz")
	if _, err := os.Stat(gzPath); err == nil {
		return loadSnapshotFile(gzPath)
	}

	path := filepath.Join(dir, id+".json")
	return loadSnapshotFile(path)
}

// RestoreSnapshot loads a snapshot and imports it into the target database.
func RestoreSnapshot(ctx context.Context, cfg config.Config, driver db.DbDriver, dir, id string) (*SyncResult, error) {
	payload, err := LoadSnapshot(dir, id)
	if err != nil {
		return nil, fmt.Errorf("load snapshot %s: %w", id, err)
	}

	return ImportPayload(ctx, cfg, driver, payload, false)
}

// loadSnapshotFile reads and JSON-decodes a snapshot file at the given path.
// Handles gzip-compressed files based on extension or magic bytes.
func loadSnapshotFile(path string) (*SyncPayload, error) {
	data, err := ReadPayloadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot file: %w", err)
	}

	var payload SyncPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}

	return &payload, nil
}

// parseSnapshotTimestamp extracts a time.Time from a snapshot ID like "snap_20260223_143022".
func parseSnapshotTimestamp(id string) time.Time {
	// Strip prefix.
	dateStr := strings.TrimPrefix(id, snapshotPrefix)
	t, err := time.Parse("20060102_150405", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
