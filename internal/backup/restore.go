package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// ReadManifest extracts and parses the manifest.json from a backup archive.
func ReadManifest(backupPath string) (*BackupManifest, error) {
	r, err := zip.OpenReader(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup archive: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open manifest.json: %w", err)
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest.json: %w", err)
		}

		var manifest BackupManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
		}

		return &manifest, nil
	}

	return nil, fmt.Errorf("manifest.json not found in backup archive")
}

// RestoreFromBackup restores a backup archive to the configured database.
// For SQLite, this replaces the database file.
// For MySQL/PostgreSQL, this pipes the SQL dump into the respective client.
func RestoreFromBackup(cfg config.Config, backupPath string) error {
	// Read and verify manifest
	manifest, err := ReadManifest(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup manifest: %w", err)
	}

	if manifest.Driver != string(cfg.Db_Driver) {
		return fmt.Errorf("backup driver mismatch: backup is %q but current config uses %q", manifest.Driver, cfg.Db_Driver)
	}

	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "modulacms-restore-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract zip to temp dir
	if err := unzip(backupPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract backup archive: %w", err)
	}

	// Restore database based on driver type
	switch cfg.Db_Driver {
	case config.Sqlite:
		if err := restoreSQLite(cfg, tempDir); err != nil {
			return fmt.Errorf("SQLite restore failed: %w", err)
		}
	case config.Mysql:
		if err := restoreMySQL(cfg, tempDir); err != nil {
			return fmt.Errorf("MySQL restore failed: %w", err)
		}
	case config.Psql:
		if err := restorePostgres(cfg, tempDir); err != nil {
			return fmt.Errorf("PostgreSQL restore failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Db_Driver)
	}

	// Restore extra files if present
	extraDir := filepath.Join(tempDir, "extra")
	if utility.DirExists(extraDir) {
		if err := restoreExtraFiles(extraDir); err != nil {
			return fmt.Errorf("failed to restore extra files: %w", err)
		}
	}

	return nil
}

func restoreSQLite(cfg config.Config, tempDir string) error {
	srcDB := filepath.Join(tempDir, "database.db")
	if !utility.FileExists(srcDB) {
		return fmt.Errorf("database.db not found in backup archive")
	}

	// Copy the database file to the configured location
	src, err := os.Open(srcDB)
	if err != nil {
		return fmt.Errorf("failed to open backup database: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(cfg.Db_URL)
	if err != nil {
		return fmt.Errorf("failed to create destination database: %w", err)
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy database file: %w", err)
	}

	return nil
}

func restoreMySQL(cfg config.Config, tempDir string) error {
	sqlFile := filepath.Join(tempDir, "database.sql")
	if !utility.FileExists(sqlFile) {
		return fmt.Errorf("database.sql not found in backup archive")
	}

	f, err := os.Open(sqlFile)
	if err != nil {
		return fmt.Errorf("failed to open SQL dump: %w", err)
	}
	defer f.Close()

	cmd := exec.Command("mysql",
		"-u", cfg.Db_User,
		"-p"+cfg.Db_Password,
		cfg.Db_Name,
	)
	cmd.Stdin = f
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore failed: %w", err)
	}

	return nil
}

func restorePostgres(cfg config.Config, tempDir string) error {
	sqlFile := filepath.Join(tempDir, "database.sql")
	if !utility.FileExists(sqlFile) {
		return fmt.Errorf("database.sql not found in backup archive")
	}

	f, err := os.Open(sqlFile)
	if err != nil {
		return fmt.Errorf("failed to open SQL dump: %w", err)
	}
	defer f.Close()

	cmd := exec.Command("psql",
		"-U", cfg.Db_User,
		"-d", cfg.Db_Name,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Db_Password)
	cmd.Stdin = f
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("psql restore failed: %w", err)
	}

	return nil
}

// unzip extracts a zip archive to dest. It validates that all extracted paths
// remain within dest to prevent zip slip (CWE-22) path traversal attacks.
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer r.Close()

	// Clean once for all path checks.
	cleanDest := filepath.Clean(dest) + string(os.PathSeparator)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Zip slip guard: resolved path must stay within dest.
		if !strings.HasPrefix(filepath.Clean(fpath)+string(os.PathSeparator), cleanDest) {
			return fmt.Errorf("illegal file path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if mkErr := os.MkdirAll(fpath, 0o755); mkErr != nil {
				return fmt.Errorf("creating directory %s: %w", f.Name, mkErr)
			}
			continue
		}

		if mkErr := os.MkdirAll(filepath.Dir(fpath), 0o755); mkErr != nil {
			return fmt.Errorf("creating parent directory for %s: %w", f.Name, mkErr)
		}

		rc, openErr := f.Open()
		if openErr != nil {
			return fmt.Errorf("opening archived file %s: %w", f.Name, openErr)
		}

		outFile, createErr := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if createErr != nil {
			rc.Close()
			return fmt.Errorf("creating %s: %w", f.Name, createErr)
		}

		_, copyErr := io.Copy(outFile, rc)
		closeErr := outFile.Close()
		rc.Close()

		if copyErr != nil {
			return fmt.Errorf("writing %s: %w", f.Name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("closing %s: %w", f.Name, closeErr)
		}
	}

	return nil
}

func restoreExtraFiles(extraDir string) error {
	return filepath.Walk(extraDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(extraDir, path)
		if err != nil {
			return err
		}

		destPath := relPath
		destDir := filepath.Dir(destPath)
		if destDir != "." {
			if mkErr := os.MkdirAll(destDir, 0o755); mkErr != nil {
				return fmt.Errorf("failed to create directory %s: %w", destDir, mkErr)
			}
		}

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer src.Close()

		dst, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", destPath, err)
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}

		return nil
	})
}
