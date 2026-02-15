// White-box tests for restore.go in the backup package.
//
// Why white-box: restoreMySQL, restorePostgres, and restoreExtraFiles are
// unexported functions with non-trivial error handling and file I/O that
// cannot be fully exercised through the exported RestoreFromBackup alone
// (which requires matching driver manifests and zip extraction).
package backup

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// ---------------------------------------------------------------------------
// ReadManifest -- additional edge cases
// ---------------------------------------------------------------------------

func TestReadManifest_MinimalValidJSON(t *testing.T) {
	// A manifest with only the driver field populated should still parse.
	// All other fields are zero-valued strings, which is valid JSON.
	t.Parallel()
	tmpDir := t.TempDir()

	manifest := BackupManifest{
		Driver: "sqlite",
	}
	zipPath := createTestZipWithManifest(t, filepath.Join(tmpDir, "minimal.zip"), manifest, false, "")

	got, err := ReadManifest(zipPath)
	if err != nil {
		t.Fatalf("ReadManifest with minimal fields: %v", err)
	}
	if got.Driver != "sqlite" {
		t.Errorf("Driver = %q, want %q", got.Driver, "sqlite")
	}
	if got.Timestamp != "" {
		t.Errorf("Timestamp should be empty for minimal manifest, got %q", got.Timestamp)
	}
	if got.Version != "" {
		t.Errorf("Version should be empty for minimal manifest, got %q", got.Version)
	}
	if got.NodeID != "" {
		t.Errorf("NodeID should be empty for minimal manifest, got %q", got.NodeID)
	}
	if got.DbName != "" {
		t.Errorf("DbName should be empty for minimal manifest, got %q", got.DbName)
	}
}

func TestReadManifest_EmptyZip(t *testing.T) {
	// A valid zip with zero entries should produce the "not found" error.
	t.Parallel()
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "empty.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(f)
	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing file: %v", err)
	}

	_, err = ReadManifest(zipPath)
	if err == nil {
		t.Fatal("expected error for empty zip, got nil")
	}
	if !strings.Contains(err.Error(), "manifest.json not found") {
		t.Errorf("expected 'manifest.json not found' error, got: %v", err)
	}
}

func TestReadManifest_ManifestWithExtraFields(t *testing.T) {
	// JSON with unknown fields should still parse successfully.
	// json.Unmarshal silently ignores unknown fields.
	t.Parallel()
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "extra-fields.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(f)

	manifestJSON := `{
		"driver": "mysql",
		"timestamp": "2025-01-01T00:00:00Z",
		"version": "2.0.0",
		"node_id": "node-x",
		"db_name": "mydb",
		"unknown_field": "should be ignored",
		"another_unknown": 42
	}`
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("creating manifest entry: %v", err)
	}
	if _, err := mw.Write([]byte(manifestJSON)); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing file: %v", err)
	}

	got, err := ReadManifest(zipPath)
	if err != nil {
		t.Fatalf("ReadManifest with extra fields: %v", err)
	}
	if got.Driver != "mysql" {
		t.Errorf("Driver = %q, want %q", got.Driver, "mysql")
	}
	if got.DbName != "mydb" {
		t.Errorf("DbName = %q, want %q", got.DbName, "mydb")
	}
}

func TestReadManifest_ManifestIsNotFirstEntry(t *testing.T) {
	// manifest.json may not be the first file in the zip. ReadManifest
	// iterates all entries, so it should find it regardless of position.
	t.Parallel()
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "manifest-last.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(f)

	// Add several files before the manifest
	for _, name := range []string{"database.db", "extra/config.toml", "extra/data.json"} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("creating %s: %v", name, err)
		}
		if _, err := w.Write([]byte("content-of-" + name)); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	// Add manifest last
	manifest := BackupManifest{
		Driver:    "postgres",
		Timestamp: "2025-06-15T12:00:00Z",
		Version:   "3.0.0",
		NodeID:    "node-last",
		DbName:    "lastdb",
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshaling manifest: %v", err)
	}
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("creating manifest entry: %v", err)
	}
	if _, err := mw.Write(manifestJSON); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing file: %v", err)
	}

	got, err := ReadManifest(zipPath)
	if err != nil {
		t.Fatalf("ReadManifest with manifest as last entry: %v", err)
	}
	if got.Driver != "postgres" {
		t.Errorf("Driver = %q, want %q", got.Driver, "postgres")
	}
	if got.NodeID != "node-last" {
		t.Errorf("NodeID = %q, want %q", got.NodeID, "node-last")
	}
}

// ---------------------------------------------------------------------------
// restoreMySQL -- error paths (no mysql binary needed)
// ---------------------------------------------------------------------------

func TestRestoreMySQL_MissingSQLFile(t *testing.T) {
	// When the extracted temp directory does not contain database.sql,
	// restoreMySQL should return a clear error before attempting to
	// invoke the mysql CLI.
	t.Parallel()
	tmpDir := t.TempDir()

	// Empty extract directory -- no database.sql
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}

	cfg := config.Config{
		Db_User:     "testuser",
		Db_Password: "testpass",
		Db_Name:     "testdb",
	}

	err := restoreMySQL(cfg, extractDir)
	if err == nil {
		t.Fatal("expected error for missing database.sql, got nil")
	}
	if !strings.Contains(err.Error(), "database.sql not found") {
		t.Errorf("error should mention 'database.sql not found', got: %v", err)
	}
}

func TestRestoreMySQL_SQLFileExistsButMysqlBinaryMissing(t *testing.T) {
	// When database.sql exists but the mysql binary is not available,
	// exec.Command should fail. This tests the error wrapping path.
	// Cannot use t.Parallel because t.Setenv modifies process-wide state.
	tmpDir := t.TempDir()

	extractDir := filepath.Join(tmpDir, "extracted")
	createTempFile(t, filepath.Join(extractDir, "database.sql"), "DROP TABLE IF EXISTS test;")

	cfg := config.Config{
		Db_User:     "testuser",
		Db_Password: "testpass",
		Db_Name:     "testdb",
	}

	// Manipulate PATH to ensure mysql is not found
	t.Setenv("PATH", tmpDir)

	err := restoreMySQL(cfg, extractDir)
	if err == nil {
		t.Fatal("expected error when mysql binary is not available, got nil")
	}
	if !strings.Contains(err.Error(), "mysql restore failed") {
		t.Errorf("error should mention 'mysql restore failed', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// restorePostgres -- error paths (no psql binary needed)
// ---------------------------------------------------------------------------

func TestRestorePostgres_MissingSQLFile(t *testing.T) {
	// When the extracted temp directory does not contain database.sql,
	// restorePostgres should return a clear error.
	t.Parallel()
	tmpDir := t.TempDir()

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}

	cfg := config.Config{
		Db_User:     "pguser",
		Db_Password: "pgpass",
		Db_Name:     "pgdb",
	}

	err := restorePostgres(cfg, extractDir)
	if err == nil {
		t.Fatal("expected error for missing database.sql, got nil")
	}
	if !strings.Contains(err.Error(), "database.sql not found") {
		t.Errorf("error should mention 'database.sql not found', got: %v", err)
	}
}

func TestRestorePostgres_SQLFileExistsButPsqlBinaryMissing(t *testing.T) {
	// When database.sql exists but psql is not available, the command
	// should fail and the error should be wrapped.
	// Cannot use t.Parallel because t.Setenv modifies process-wide state.
	tmpDir := t.TempDir()

	extractDir := filepath.Join(tmpDir, "extracted")
	createTempFile(t, filepath.Join(extractDir, "database.sql"), "SELECT 1;")

	cfg := config.Config{
		Db_User:     "pguser",
		Db_Password: "pgpass",
		Db_Name:     "pgdb",
	}

	t.Setenv("PATH", tmpDir)

	err := restorePostgres(cfg, extractDir)
	if err == nil {
		t.Fatal("expected error when psql binary is not available, got nil")
	}
	if !strings.Contains(err.Error(), "psql restore failed") {
		t.Errorf("error should mention 'psql restore failed', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// restoreSQLite -- additional edge cases
// ---------------------------------------------------------------------------

func TestRestoreSQLite_BinaryContent(t *testing.T) {
	// SQLite database files contain binary content. Verify that binary
	// data is preserved through the copy without corruption.
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a byte sequence covering all 256 byte values
	var binaryContent []byte
	for i := range 256 {
		binaryContent = append(binaryContent, byte(i))
	}
	// Repeat to make it non-trivial size
	fullContent := make([]byte, 0, 256*10)
	for range 10 {
		fullContent = append(fullContent, binaryContent...)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	dbPath := filepath.Join(extractDir, "database.db")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}
	if err := os.WriteFile(dbPath, fullContent, 0o644); err != nil {
		t.Fatalf("writing binary db: %v", err)
	}

	destDB := filepath.Join(tmpDir, "restored.db")
	cfg := config.Config{
		Db_URL: destDB,
	}

	if err := restoreSQLite(cfg, extractDir); err != nil {
		t.Fatalf("restoreSQLite with binary content: %v", err)
	}

	got, err := os.ReadFile(destDB)
	if err != nil {
		t.Fatalf("reading restored db: %v", err)
	}
	if len(got) != len(fullContent) {
		t.Fatalf("restored db size = %d bytes, want %d bytes", len(got), len(fullContent))
	}
	for i := range fullContent {
		if got[i] != fullContent[i] {
			t.Fatalf("byte mismatch at position %d: got 0x%02x, want 0x%02x", i, got[i], fullContent[i])
		}
	}
}

func TestRestoreSQLite_DestinationDirDoesNotExist(t *testing.T) {
	// When the destination directory does not exist, os.Create should fail.
	// This tests the "failed to create destination database" error path.
	t.Parallel()
	tmpDir := t.TempDir()

	extractDir := filepath.Join(tmpDir, "extracted")
	createTempFile(t, filepath.Join(extractDir, "database.db"), "db-content")

	cfg := config.Config{
		Db_URL: filepath.Join(tmpDir, "nonexistent-parent", "deep", "restored.db"),
	}

	err := restoreSQLite(cfg, extractDir)
	if err == nil {
		t.Fatal("expected error when destination directory does not exist, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create destination database") {
		t.Errorf("error should mention 'failed to create destination database', got: %v", err)
	}
}

func TestRestoreSQLite_EmptyDatabaseFile(t *testing.T) {
	// An empty database.db should still be copied without error.
	// This covers the edge case of a zero-byte file.
	t.Parallel()
	tmpDir := t.TempDir()

	extractDir := filepath.Join(tmpDir, "extracted")
	createTempFile(t, filepath.Join(extractDir, "database.db"), "")

	destDB := filepath.Join(tmpDir, "restored.db")
	cfg := config.Config{
		Db_URL: destDB,
	}

	if err := restoreSQLite(cfg, extractDir); err != nil {
		t.Fatalf("restoreSQLite with empty db: %v", err)
	}

	info, err := os.Stat(destDB)
	if err != nil {
		t.Fatalf("stat restored db: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("restored db size = %d, want 0", info.Size())
	}
}

// ---------------------------------------------------------------------------
// restoreExtraFiles -- nested directories and edge cases
// ---------------------------------------------------------------------------

func TestRestoreExtraFiles_NestedDirectories(t *testing.T) {
	// restoreExtraFiles creates directories as needed for nested paths.
	// Since the function writes relative to cwd, we use an explicit
	// working directory approach: create the nested structure and verify
	// that os.MkdirAll is called for subdirectories.
	//
	// REQUIRES REFACTOR: restoreExtraFiles should accept a destination
	// directory parameter. This test works within the limitation by
	// cleaning up files written to cwd.
	t.Parallel()
	tmpDir := t.TempDir()

	extraDir := filepath.Join(tmpDir, "extra")
	createTempFile(t, filepath.Join(extraDir, "configs", "app.toml"), "app-config")
	createTempFile(t, filepath.Join(extraDir, "configs", "nested", "deep.yaml"), "deep-config")

	err := restoreExtraFiles(extraDir)
	if err != nil {
		t.Fatalf("restoreExtraFiles with nested dirs: %v", err)
	}

	// Clean up files created relative to cwd
	t.Cleanup(func() {
		os.RemoveAll("configs")
	})
}

func TestRestoreExtraFiles_MultipleFiles(t *testing.T) {
	// Verify that restoreExtraFiles walks all files in the extra directory.
	t.Parallel()
	tmpDir := t.TempDir()

	extraDir := filepath.Join(tmpDir, "extra")
	files := []struct {
		relPath string
		content string
	}{
		{"restore_test_a.txt", "content-a"},
		{"restore_test_b.txt", "content-b"},
		{"restore_test_sub/c.txt", "content-c"},
	}

	for _, f := range files {
		createTempFile(t, filepath.Join(extraDir, f.relPath), f.content)
	}

	err := restoreExtraFiles(extraDir)
	if err != nil {
		t.Fatalf("restoreExtraFiles: %v", err)
	}

	// Verify files were created (relative to cwd)
	for _, f := range files {
		content, readErr := os.ReadFile(f.relPath)
		if readErr != nil {
			t.Errorf("failed to read restored file %s: %v", f.relPath, readErr)
			continue
		}
		if string(content) != f.content {
			t.Errorf("file %s: got %q, want %q", f.relPath, string(content), f.content)
		}
	}

	// Clean up
	t.Cleanup(func() {
		os.Remove("restore_test_a.txt")
		os.Remove("restore_test_b.txt")
		os.RemoveAll("restore_test_sub")
	})
}

// ---------------------------------------------------------------------------
// RestoreFromBackup -- extra files integration path
// ---------------------------------------------------------------------------

func TestRestoreFromBackup_SQLite_WithExtraFiles(t *testing.T) {
	// This tests the branch in RestoreFromBackup where the extracted
	// archive contains an "extra" directory, triggering restoreExtraFiles.
	t.Parallel()
	tmpDir := t.TempDir()

	// Build a zip with manifest, database.db, and extra/ files
	zipPath := filepath.Join(tmpDir, "backup-with-extras.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	zw := zip.NewWriter(f)

	// Write manifest
	manifest := BackupManifest{
		Driver:    "sqlite",
		Timestamp: "2025-01-15T10:00:00Z",
		Version:   "1.0.0",
		NodeID:    "node-extra",
		DbName:    "extradb",
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshaling manifest: %v", err)
	}
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("creating manifest: %v", err)
	}
	if _, err := mw.Write(manifestJSON); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	// Write database.db
	dbContent := "sqlite-db-with-extras"
	dw, err := zw.Create("database.db")
	if err != nil {
		t.Fatalf("creating db entry: %v", err)
	}
	if _, err := dw.Write([]byte(dbContent)); err != nil {
		t.Fatalf("writing db: %v", err)
	}

	// Write extra files
	ew, err := zw.Create("extra/restore_integration_test_config.toml")
	if err != nil {
		t.Fatalf("creating extra entry: %v", err)
	}
	if _, err := ew.Write([]byte("extra-config-content")); err != nil {
		t.Fatalf("writing extra: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing zip file: %v", err)
	}

	// Restore
	destDB := filepath.Join(tmpDir, "restored.db")
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    destDB,
	}

	err = RestoreFromBackup(cfg, zipPath)
	if err != nil {
		t.Fatalf("RestoreFromBackup with extra files: %v", err)
	}

	// Verify database was restored
	got, readErr := os.ReadFile(destDB)
	if readErr != nil {
		t.Fatalf("reading restored db: %v", readErr)
	}
	if string(got) != dbContent {
		t.Errorf("restored db content = %q, want %q", string(got), dbContent)
	}

	// The extra file should have been restored relative to cwd
	t.Cleanup(func() {
		os.Remove("restore_integration_test_config.toml")
	})
}

func TestRestoreFromBackup_SQLite_NoExtraDir(t *testing.T) {
	// When the backup archive does not contain an "extra" directory,
	// RestoreFromBackup should complete without error and skip the
	// restoreExtraFiles call. This is the most common case.
	t.Parallel()
	tmpDir := t.TempDir()

	manifest := BackupManifest{
		Driver:    "sqlite",
		Timestamp: "2025-02-01T00:00:00Z",
		Version:   "1.0.0",
		NodeID:    "node-noextra",
		DbName:    "nodb",
	}
	zipPath := createTestZipWithManifest(t,
		filepath.Join(tmpDir, "no-extra.zip"),
		manifest, true, "database.db")

	destDB := filepath.Join(tmpDir, "restored.db")
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    destDB,
	}

	if err := RestoreFromBackup(cfg, zipPath); err != nil {
		t.Fatalf("RestoreFromBackup without extra dir: %v", err)
	}

	// Verify database was restored
	got, err := os.ReadFile(destDB)
	if err != nil {
		t.Fatalf("reading restored db: %v", err)
	}
	if string(got) != "fake-database-content" {
		t.Errorf("restored db = %q, want %q", string(got), "fake-database-content")
	}
}

func TestRestoreFromBackup_SQLite_MissingDatabaseInArchive(t *testing.T) {
	// The zip has a manifest but no database.db file. The restoreSQLite
	// function should fail with a clear error.
	t.Parallel()
	tmpDir := t.TempDir()

	manifest := BackupManifest{
		Driver:    "sqlite",
		Timestamp: "2025-03-01T00:00:00Z",
		Version:   "1.0.0",
		NodeID:    "node-nodb",
		DbName:    "nodb",
	}
	// includeDB=false means no database.db in the zip
	zipPath := createTestZipWithManifest(t,
		filepath.Join(tmpDir, "no-db.zip"),
		manifest, false, "")

	destDB := filepath.Join(tmpDir, "restored.db")
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    destDB,
	}

	err := RestoreFromBackup(cfg, zipPath)
	if err == nil {
		t.Fatal("expected error for missing database.db in archive, got nil")
	}
	if !strings.Contains(err.Error(), "SQLite restore failed") {
		t.Errorf("error should mention 'SQLite restore failed', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RestoreFromBackup -- driver matching edge cases
// ---------------------------------------------------------------------------

func TestRestoreFromBackup_DriverMismatch_AllCombinations(t *testing.T) {
	// Verify that every pairing of mismatched drivers produces a
	// driver mismatch error. This catches regressions where a new
	// driver constant is added but the check is bypassed.
	t.Parallel()

	drivers := []struct {
		name   string
		driver config.DbDriver
	}{
		{"sqlite", config.Sqlite},
		{"mysql", config.Mysql},
		{"postgres", config.Psql},
	}

	for _, backupDriver := range drivers {
		for _, configDriver := range drivers {
			if backupDriver.name == configDriver.name {
				continue // skip matching pairs
			}

			t.Run(backupDriver.name+"_backup_vs_"+configDriver.name+"_config", func(t *testing.T) {
				t.Parallel()
				tmpDir := t.TempDir()

				manifest := BackupManifest{
					Driver:    backupDriver.name,
					Timestamp: "2025-01-01T00:00:00Z",
					Version:   "1.0.0",
					NodeID:    "node-1",
					DbName:    "test",
				}
				zipPath := createTestZipWithManifest(t,
					filepath.Join(tmpDir, "backup.zip"),
					manifest, true, "database.db")

				cfg := config.Config{
					Db_Driver: configDriver.driver,
					Db_URL:    filepath.Join(tmpDir, "dest.db"),
				}

				err := RestoreFromBackup(cfg, zipPath)
				if err == nil {
					t.Fatalf("expected driver mismatch error for backup=%s config=%s, got nil",
						backupDriver.name, configDriver.name)
				}
				if !strings.Contains(err.Error(), "driver mismatch") {
					t.Errorf("error should mention 'driver mismatch', got: %v", err)
				}
				// Verify both driver names appear in the error message
				if !strings.Contains(err.Error(), backupDriver.name) {
					t.Errorf("error should mention backup driver %q, got: %v",
						backupDriver.name, err)
				}
				if !strings.Contains(err.Error(), string(configDriver.driver)) {
					t.Errorf("error should mention config driver %q, got: %v",
						configDriver.driver, err)
				}
			})
		}
	}
}

// ---------------------------------------------------------------------------
// RestoreFromBackup -- MySQL and PostgreSQL missing-binary error paths
// ---------------------------------------------------------------------------

func TestRestoreFromBackup_MySQL_MissingBinary(t *testing.T) {
	// When the backup archive driver is mysql and the mysql CLI is not
	// available, RestoreFromBackup should fail with a meaningful error.
	// Cannot use t.Parallel because t.Setenv modifies process-wide state.
	tmpDir := t.TempDir()

	// Create zip with mysql manifest and database.sql
	zipPath := filepath.Join(tmpDir, "mysql-backup.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(f)

	manifest := BackupManifest{
		Driver:    "mysql",
		Timestamp: "2025-01-01T00:00:00Z",
		Version:   "1.0.0",
		NodeID:    "node-1",
		DbName:    "testdb",
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshaling manifest: %v", err)
	}
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("creating manifest: %v", err)
	}
	if _, err := mw.Write(manifestJSON); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	sw, err := zw.Create("database.sql")
	if err != nil {
		t.Fatalf("creating sql entry: %v", err)
	}
	if _, err := sw.Write([]byte("CREATE TABLE test (id INT);")); err != nil {
		t.Fatalf("writing sql: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing file: %v", err)
	}

	// Remove mysql from PATH
	t.Setenv("PATH", tmpDir)

	cfg := config.Config{
		Db_Driver:   config.Mysql,
		Db_User:     "user",
		Db_Password: "pass",
		Db_Name:     "testdb",
	}

	err = RestoreFromBackup(cfg, zipPath)
	if err == nil {
		t.Fatal("expected error when mysql binary is not available, got nil")
	}
	if !strings.Contains(err.Error(), "MySQL restore failed") {
		t.Errorf("error should mention 'MySQL restore failed', got: %v", err)
	}
}

func TestRestoreFromBackup_Postgres_MissingBinary(t *testing.T) {
	// Same as above but for PostgreSQL.
	// Cannot use t.Parallel because t.Setenv modifies process-wide state.
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "pg-backup.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(f)

	manifest := BackupManifest{
		Driver:    "postgres",
		Timestamp: "2025-01-01T00:00:00Z",
		Version:   "1.0.0",
		NodeID:    "node-1",
		DbName:    "pgdb",
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshaling manifest: %v", err)
	}
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("creating manifest: %v", err)
	}
	if _, err := mw.Write(manifestJSON); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	sw, err := zw.Create("database.sql")
	if err != nil {
		t.Fatalf("creating sql entry: %v", err)
	}
	if _, err := sw.Write([]byte("CREATE TABLE test (id SERIAL);")); err != nil {
		t.Fatalf("writing sql: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing file: %v", err)
	}

	// Remove psql from PATH
	t.Setenv("PATH", tmpDir)

	cfg := config.Config{
		Db_Driver:   config.Psql,
		Db_User:     "pguser",
		Db_Password: "pgpass",
		Db_Name:     "pgdb",
	}

	err = RestoreFromBackup(cfg, zipPath)
	if err == nil {
		t.Fatal("expected error when psql binary is not available, got nil")
	}
	if !strings.Contains(err.Error(), "PostgreSQL restore failed") {
		t.Errorf("error should mention 'PostgreSQL restore failed', got: %v", err)
	}
}
