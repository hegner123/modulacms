package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

type backupName func(string, string) string

// BackupManifest records metadata about a backup archive.
type BackupManifest struct {
	Driver    string `json:"driver"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	NodeID    string `json:"node_id"`
	DbName    string `json:"db_name"`
}

func TimestampBackupName(output string, timestamp string) string {
	return fmt.Sprintf("%s_%s.zip", output, timestamp)
}

// CreateFullBackup creates a zip archive containing the database and any
// configured backup paths. It returns the file path, size in bytes, or an error.
func CreateFullBackup(cfg config.Config) (path string, sizeBytes int64, err error) {
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
	case config.Mysql:
		if addErr := addMySQLDump(zipWriter, cfg); addErr != nil {
			return "", 0, fmt.Errorf("failed to add MySQL dump: %w", addErr)
		}
	case config.Psql:
		if addErr := addPostgresDump(zipWriter, cfg); addErr != nil {
			return "", 0, fmt.Errorf("failed to add PostgreSQL dump: %w", addErr)
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

func addMySQLDump(zw *zip.Writer, cfg config.Config) error {
	cmd := exec.Command("mysqldump",
		"-u", cfg.Db_User,
		"-p"+cfg.Db_Password,
		cfg.Db_Name,
	)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}

	w, err := zw.Create("database.sql")
	if err != nil {
		return fmt.Errorf("failed to create database.sql in archive: %w", err)
	}

	if _, err = w.Write(output); err != nil {
		return fmt.Errorf("failed to write MySQL dump to archive: %w", err)
	}

	return nil
}

func addPostgresDump(zw *zip.Writer, cfg config.Config) error {
	cmd := exec.Command("pg_dump",
		"-U", cfg.Db_User,
		"-d", cfg.Db_Name,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Db_Password)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}

	w, err := zw.Create("database.sql")
	if err != nil {
		return fmt.Errorf("failed to create database.sql in archive: %w", err)
	}

	if _, err = w.Write(output); err != nil {
		return fmt.Errorf("failed to write PostgreSQL dump to archive: %w", err)
	}

	return nil
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
