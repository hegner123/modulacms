// Package backup provides backup and restore functionality for Modula databases.
// It supports SQLite, MySQL, and PostgreSQL, and can include additional paths
// in backup archives. Backups are created as zip files containing database dumps
// and metadata.
package backup

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// splitHostPort splits a "host:port" string. If no port is present, defaultPort is used.
func splitHostPort(addr, defaultPort string) (host, port string) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		// No port in addr
		return addr, defaultPort
	}
	return h, p
}

type backupName func(string, string) string

// BackupManifest records metadata about a backup archive.
type BackupManifest struct {
	Driver    string `json:"driver"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	NodeID    string `json:"node_id"`
	DbName    string `json:"db_name"`
}

// TimestampBackupName returns a backup filename with the given output directory and timestamp.
func TimestampBackupName(output string, timestamp string) string {
	return fmt.Sprintf("%s_%s.zip", output, timestamp)
}

// CreateFullBackup creates a zip archive containing the database and any
// configured backup paths. It returns the file path, size in bytes, or an error.
// For MySQL and PostgreSQL, the driver parameter is used for a pure-Go export
// via DeployOps. For SQLite, the database file is copied directly and driver
// may be nil.
func CreateFullBackup(cfg config.Config, driver db.DbDriver) (path string, sizeBytes int64, err error) {
	// Determine output directory
	outputDir := cfg.Backup_Option
	if outputDir == "" {
		outputDir = "./"
	}
	backupDir := filepath.Join(outputDir, "backups")
	if mkErr := os.MkdirAll(backupDir, 0o755); mkErr != nil {
		return "", 0, fmt.Errorf("failed to create backup directory: %w", mkErr)
	}

	// Generate filename with timestamp
	ts := time.Now().UTC().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.zip", ts)
	fullPath := filepath.Join(backupDir, filename)

	// Create the zip file
	zipFile, err := os.Create(fullPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create backup file: %w", err)
	}

	zipClosed := false
	defer func() {
		if !zipClosed {
			zipFile.Close()
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	writerClosed := false
	defer func() {
		if !writerClosed {
			zipWriter.Close()
		}
	}()

	// Add database to zip based on driver type
	switch cfg.Db_Driver {
	case config.Sqlite:
		if addErr := addSQLiteDB(zipWriter, cfg.Db_URL); addErr != nil {
			return "", 0, fmt.Errorf("failed to add SQLite database: %w", addErr)
		}
	case config.Mysql, config.Psql:
		if driver == nil {
			return "", 0, fmt.Errorf("driver required for %s backup", cfg.Db_Driver)
		}
		if addErr := addGoSQLDump(zipWriter, driver); addErr != nil {
			return "", 0, fmt.Errorf("failed to add database dump: %w", addErr)
		}
	default:
		return "", 0, fmt.Errorf("unsupported database driver: %s", cfg.Db_Driver)
	}

	// Write manifest
	manifest := BackupManifest{
		Driver:    string(cfg.Db_Driver),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   utility.GetCurrentVersion(),
		NodeID:    cfg.Node_ID,
		DbName:    cfg.Db_Name,
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	manifestWriter, err := zipWriter.Create("manifest.json")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create manifest in archive: %w", err)
	}
	if _, err = manifestWriter.Write(manifestJSON); err != nil {
		return "", 0, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Add extra backup paths
	for _, p := range cfg.Backup_Paths {
		if p == "" {
			continue
		}
		if !utility.DirExists(p) && !utility.FileExists(p) {
			continue
		}
		if addErr := addFilesToZip(zipWriter, p, filepath.Join("extra", filepath.Base(p))); addErr != nil {
			return "", 0, fmt.Errorf("failed to add backup path %s: %w", p, addErr)
		}
	}

	// Close writers before stat
	if err = zipWriter.Close(); err != nil {
		return "", 0, fmt.Errorf("failed to finalize zip: %w", err)
	}
	writerClosed = true
	if err = zipFile.Close(); err != nil {
		return "", 0, fmt.Errorf("failed to close zip file: %w", err)
	}
	zipClosed = true

	// Get file size
	info, err := os.Stat(fullPath)
	if err != nil {
		return fullPath, 0, fmt.Errorf("backup created but failed to stat: %w", err)
	}

	return fullPath, info.Size(), nil
}

func addSQLiteDB(zw *zip.Writer, dbPath string) error {
	if !utility.FileExists(dbPath) {
		return fmt.Errorf("database file does not exist: %s", dbPath)
	}

	f, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database file: %w", err)
	}
	defer f.Close()

	w, err := zw.Create("database.db")
	if err != nil {
		return fmt.Errorf("failed to create database entry in archive: %w", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		return fmt.Errorf("failed to copy database to archive: %w", err)
	}

	return nil
}

// addGoSQLDump exports all CMS tables as SQL INSERT statements using DeployOps.
// Pure Go — no external tools (pg_dump, mysqldump) required.
func addGoSQLDump(zw *zip.Writer, driver db.DbDriver) error {
	ops, err := db.NewDeployOps(driver)
	if err != nil {
		return fmt.Errorf("failed to create deploy ops: %w", err)
	}

	ctx := context.Background()

	w, err := zw.Create("database.sql")
	if err != nil {
		return fmt.Errorf("failed to create database.sql in archive: %w", err)
	}

	// Write header
	fmt.Fprintf(w, "-- ModulaCMS backup\n")
	fmt.Fprintf(w, "-- Generated: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "-- Method: pure-go DeployOps export\n\n")

	// Export each table
	for table := range db.AllTablesExported() {
		columns, rows, queryErr := ops.QueryAllRows(ctx, table)
		if queryErr != nil {
			// Table might not exist yet — skip silently
			fmt.Fprintf(w, "-- Skipped %s: %v\n\n", string(table), queryErr)
			continue
		}

		if len(rows) == 0 {
			fmt.Fprintf(w, "-- %s: 0 rows\n\n", string(table))
			continue
		}

		fmt.Fprintf(w, "-- %s: %d rows\n", string(table), len(rows))

		for _, row := range rows {
			fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (", string(table), strings.Join(columns, ", "))
			for j, val := range row {
				if j > 0 {
					fmt.Fprint(w, ", ")
				}
				if val == nil {
					fmt.Fprint(w, "NULL")
				} else {
					switch v := val.(type) {
					case int64:
						fmt.Fprintf(w, "%d", v)
					case string:
						fmt.Fprintf(w, "'%s'", escapeSQLString(v))
					default:
						fmt.Fprintf(w, "'%s'", escapeSQLString(fmt.Sprintf("%v", v)))
					}
				}
			}
			fmt.Fprint(w, ");\n")
		}
		fmt.Fprint(w, "\n")
	}

	return nil
}

// escapeSQLString escapes single quotes in SQL string literals.
func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func addFilesToZip(zipWriter *zip.Writer, dir, baseInZip string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(baseInZip, relPath)
		fileWriter, err := zipWriter.Create(zipPath)
		if err != nil {
			return fmt.Errorf("failed to add file %s to zip: %w", path, err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		_, err = io.Copy(fileWriter, file)
		if err != nil {
			return fmt.Errorf("failed to copy file %s to zip: %w", path, err)
		}

		return nil
	})
}
