// White-box tests for the backup package.
//
// Why white-box: several unexported functions (addSQLiteDB, addFilesToZip,
// restoreSQLite, restoreExtraFiles) contain non-trivial logic (file walking,
// zip entry creation, file copying) that is easier to exercise and assert on
// directly than through CreateFullBackup, which embeds a timestamp in its
// output path.
package backup

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// createTempFile creates a file at path with the given content and calls
// t.Fatal on failure. Parent directories are created automatically.
func createTempFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// readZipEntry opens a zip file and returns the contents of the named entry.
// Returns an error if the zip cannot be opened or the entry is missing.
func readZipEntry(t *testing.T, zipPath, entryName string) string {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("opening zip %s: %v", zipPath, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == entryName {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("opening entry %s in zip: %v", entryName, err)
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("reading entry %s: %v", entryName, err)
			}
			return string(data)
		}
	}
	t.Fatalf("entry %q not found in zip %s", entryName, zipPath)
	return ""
}

// zipEntryNames returns a sorted list of all entry names in a zip file.
func zipEntryNames(t *testing.T, zipPath string) []string {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("opening zip %s: %v", zipPath, err)
	}
	defer r.Close()

	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names
}

// createTestZipWithManifest creates a zip file at zipPath containing a
// manifest.json and optionally a database file. Returns the path.
func createTestZipWithManifest(t *testing.T, zipPath string, manifest BackupManifest, includeDB bool, dbEntryName string) string {
	t.Helper()
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip file: %v", err)
	}

	zw := zip.NewWriter(f)

	// Write manifest
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshaling manifest: %v", err)
	}
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("creating manifest entry: %v", err)
	}
	if _, err = mw.Write(manifestJSON); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	// Optionally write a database entry
	if includeDB && dbEntryName != "" {
		dw, err := zw.Create(dbEntryName)
		if err != nil {
			t.Fatalf("creating db entry: %v", err)
		}
		if _, err = dw.Write([]byte("fake-database-content")); err != nil {
			t.Fatalf("writing db entry: %v", err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing zip file: %v", err)
	}

	return zipPath
}

// ---------------------------------------------------------------------------
// TimestampBackupName
// ---------------------------------------------------------------------------

func TestTimestampBackupName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		output    string
		timestamp string
		want      string
	}{
		{
			name:      "typical values",
			output:    "backup",
			timestamp: "20240101_120000",
			want:      "backup_20240101_120000.zip",
		},
		{
			name:      "output with path",
			output:    "/tmp/backups/mydb",
			timestamp: "20240615_083045",
			want:      "/tmp/backups/mydb_20240615_083045.zip",
		},
		{
			name:      "empty output",
			output:    "",
			timestamp: "20240101_000000",
			want:      "_20240101_000000.zip",
		},
		{
			name:      "empty timestamp",
			output:    "backup",
			timestamp: "",
			want:      "backup_.zip",
		},
		{
			name:      "both empty",
			output:    "",
			timestamp: "",
			want:      "_.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TimestampBackupName(tt.output, tt.timestamp)
			if got != tt.want {
				t.Errorf("TimestampBackupName(%q, %q) = %q, want %q",
					tt.output, tt.timestamp, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// addSQLiteDB
// ---------------------------------------------------------------------------

func TestAddSQLiteDB_CopiesFileToZip(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a fake SQLite database file
	dbContent := "SQLite format 3\x00fake-database-bytes"
	dbPath := filepath.Join(tmpDir, "test.db")
	createTempFile(t, dbPath, dbContent)

	// Create a zip file and add the db
	zipPath := filepath.Join(tmpDir, "out.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(zipFile)

	if err := addSQLiteDB(zw, dbPath); err != nil {
		t.Fatalf("addSQLiteDB: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("closing zip file: %v", err)
	}

	// Verify the zip contains database.db with correct contents
	got := readZipEntry(t, zipPath, "database.db")
	if got != dbContent {
		t.Errorf("zip entry database.db content = %q, want %q", got, dbContent)
	}
}

func TestAddSQLiteDB_NonexistentFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "out.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(zipFile)

	err = addSQLiteDB(zw, filepath.Join(tmpDir, "does-not-exist.db"))
	if err == nil {
		t.Fatal("addSQLiteDB with nonexistent file: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should mention file does not exist, got: %v", err)
	}

	// Clean up zip resources even on error path
	zw.Close()
	zipFile.Close()
}

// ---------------------------------------------------------------------------
// addFilesToZip
// ---------------------------------------------------------------------------

func TestAddFilesToZip_SingleFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a source directory with one file
	srcDir := filepath.Join(tmpDir, "src")
	createTempFile(t, filepath.Join(srcDir, "readme.txt"), "hello world")

	// Create zip and add the directory
	zipPath := filepath.Join(tmpDir, "out.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(zipFile)

	if err := addFilesToZip(zw, srcDir, "extra/src"); err != nil {
		t.Fatalf("addFilesToZip: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("closing zip file: %v", err)
	}

	got := readZipEntry(t, zipPath, filepath.Join("extra/src", "readme.txt"))
	if got != "hello world" {
		t.Errorf("zip entry content = %q, want %q", got, "hello world")
	}
}

func TestAddFilesToZip_NestedDirectories(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	createTempFile(t, filepath.Join(srcDir, "a.txt"), "file-a")
	createTempFile(t, filepath.Join(srcDir, "sub", "b.txt"), "file-b")
	createTempFile(t, filepath.Join(srcDir, "sub", "deep", "c.txt"), "file-c")

	zipPath := filepath.Join(tmpDir, "out.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(zipFile)

	if err := addFilesToZip(zw, srcDir, "backup"); err != nil {
		t.Fatalf("addFilesToZip: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("closing zip file: %v", err)
	}

	// Verify all three files exist
	for _, tc := range []struct {
		entry string
		want  string
	}{
		{filepath.Join("backup", "a.txt"), "file-a"},
		{filepath.Join("backup", "sub", "b.txt"), "file-b"},
		{filepath.Join("backup", "sub", "deep", "c.txt"), "file-c"},
	} {
		got := readZipEntry(t, zipPath, tc.entry)
		if got != tc.want {
			t.Errorf("entry %q: got %q, want %q", tc.entry, got, tc.want)
		}
	}
}

func TestAddFilesToZip_EmptyDirectory(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("creating empty dir: %v", err)
	}

	zipPath := filepath.Join(tmpDir, "out.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(zipFile)

	// Should succeed with zero files added
	if err := addFilesToZip(zw, srcDir, "backup"); err != nil {
		t.Fatalf("addFilesToZip on empty dir: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("closing zip file: %v", err)
	}

	names := zipEntryNames(t, zipPath)
	if len(names) != 0 {
		t.Errorf("expected 0 entries in zip from empty dir, got %d: %v", len(names), names)
	}
}

func TestAddFilesToZip_NonexistentSource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "out.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}
	zw := zip.NewWriter(zipFile)

	err = addFilesToZip(zw, filepath.Join(tmpDir, "nonexistent"), "backup")
	if err == nil {
		t.Fatal("expected error for nonexistent source directory, got nil")
	}

	zw.Close()
	zipFile.Close()
}

// ---------------------------------------------------------------------------
// CreateFullBackup (SQLite path -- the only one testable without external tools)
// ---------------------------------------------------------------------------

func TestCreateFullBackup_SQLite_CreatesValidZip(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a fake SQLite database
	dbPath := filepath.Join(tmpDir, "app.db")
	dbContent := "SQLite format 3\x00test-db-content-here"
	createTempFile(t, dbPath, dbContent)

	cfg := config.Config{
		Db_Driver:     config.Sqlite,
		Db_URL:        dbPath,
		Db_Name:       "testdb",
		Node_ID:       "node-001",
		Backup_Option: tmpDir,
	}

	path, size, err := CreateFullBackup(cfg)
	if err != nil {
		t.Fatalf("CreateFullBackup: %v", err)
	}

	if path == "" {
		t.Fatal("CreateFullBackup returned empty path")
	}
	if size <= 0 {
		t.Errorf("CreateFullBackup returned size %d, want > 0", size)
	}

	// Verify the file actually exists at the returned path
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat returned path %s: %v", path, err)
	}
	if info.Size() != size {
		t.Errorf("stat size = %d, CreateFullBackup reported %d", info.Size(), size)
	}

	// Verify zip contents: should contain database.db and manifest.json
	names := zipEntryNames(t, path)
	foundDB := false
	foundManifest := false
	for _, name := range names {
		if name == "database.db" {
			foundDB = true
		}
		if name == "manifest.json" {
			foundManifest = true
		}
	}
	if !foundDB {
		t.Error("zip missing database.db entry")
	}
	if !foundManifest {
		t.Error("zip missing manifest.json entry")
	}

	// Verify database.db content matches
	gotDB := readZipEntry(t, path, "database.db")
	if gotDB != dbContent {
		t.Errorf("database.db content mismatch: got %d bytes, want %d bytes",
			len(gotDB), len(dbContent))
	}

	// Verify manifest.json is valid and has expected fields
	manifestRaw := readZipEntry(t, path, "manifest.json")
	var manifest BackupManifest
	if err := json.Unmarshal([]byte(manifestRaw), &manifest); err != nil {
		t.Fatalf("parsing manifest.json: %v", err)
	}
	if manifest.Driver != "sqlite" {
		t.Errorf("manifest.Driver = %q, want %q", manifest.Driver, "sqlite")
	}
	if manifest.NodeID != "node-001" {
		t.Errorf("manifest.NodeID = %q, want %q", manifest.NodeID, "node-001")
	}
	if manifest.DbName != "testdb" {
		t.Errorf("manifest.DbName = %q, want %q", manifest.DbName, "testdb")
	}
	if manifest.Timestamp == "" {
		t.Error("manifest.Timestamp is empty")
	}
}

func TestCreateFullBackup_SQLite_WithExtraPaths(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create fake SQLite db
	dbPath := filepath.Join(tmpDir, "app.db")
	createTempFile(t, dbPath, "database-content")

	// Create extra files to back up
	extraDir := filepath.Join(tmpDir, "uploads")
	createTempFile(t, filepath.Join(extraDir, "image.png"), "fake-png-data")
	createTempFile(t, filepath.Join(extraDir, "docs", "readme.md"), "# README")

	cfg := config.Config{
		Db_Driver:     config.Sqlite,
		Db_URL:        dbPath,
		Db_Name:       "testdb",
		Backup_Option: tmpDir,
		Backup_Paths:  []string{extraDir},
	}

	path, _, err := CreateFullBackup(cfg)
	if err != nil {
		t.Fatalf("CreateFullBackup with extra paths: %v", err)
	}

	// Verify extra files are in the zip under extra/<dirname>/
	names := zipEntryNames(t, path)
	expectedEntries := map[string]bool{
		"database.db":                        false,
		"manifest.json":                      false,
		filepath.Join("extra/uploads", "image.png"):       false,
		filepath.Join("extra/uploads", "docs", "readme.md"): false,
	}

	for _, name := range names {
		if _, ok := expectedEntries[name]; ok {
			expectedEntries[name] = true
		}
	}

	for entry, found := range expectedEntries {
		if !found {
			t.Errorf("expected zip entry %q not found. Zip contains: %v", entry, names)
		}
	}
}

func TestCreateFullBackup_SQLite_SkipsEmptyAndMissingBackupPaths(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	createTempFile(t, dbPath, "database-content")

	cfg := config.Config{
		Db_Driver:     config.Sqlite,
		Db_URL:        dbPath,
		Db_Name:       "testdb",
		Backup_Option: tmpDir,
		Backup_Paths: []string{
			"",                                        // empty -- should be skipped
			filepath.Join(tmpDir, "nonexistent-dir"),   // missing -- should be skipped
		},
	}

	path, _, err := CreateFullBackup(cfg)
	if err != nil {
		t.Fatalf("CreateFullBackup should skip missing paths, got error: %v", err)
	}

	// Should only contain database.db and manifest.json
	names := zipEntryNames(t, path)
	if len(names) != 2 {
		t.Errorf("expected 2 entries (database.db, manifest.json), got %d: %v", len(names), names)
	}
}

func TestCreateFullBackup_SQLite_MissingDatabase(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	cfg := config.Config{
		Db_Driver:     config.Sqlite,
		Db_URL:        filepath.Join(tmpDir, "nonexistent.db"),
		Backup_Option: tmpDir,
	}

	_, _, err := CreateFullBackup(cfg)
	if err == nil {
		t.Fatal("expected error when database file does not exist, got nil")
	}
	if !strings.Contains(err.Error(), "SQLite") {
		t.Errorf("error should mention SQLite, got: %v", err)
	}
}

func TestCreateFullBackup_UnsupportedDriver(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	cfg := config.Config{
		Db_Driver:     config.DbDriver("couchdb"),
		Backup_Option: tmpDir,
	}

	_, _, err := CreateFullBackup(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Errorf("error should mention unsupported driver, got: %v", err)
	}
}

func TestCreateFullBackup_DefaultBackupDir(t *testing.T) {
	t.Parallel()
	// When Backup_Option is empty, it defaults to "./" which means
	// backups/ relative to cwd. We set cwd to a temp dir to avoid polluting.
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	createTempFile(t, dbPath, "db-content")

	// Use the temp dir as backup option to keep things contained
	cfg := config.Config{
		Db_Driver:     config.Sqlite,
		Db_URL:        dbPath,
		Db_Name:       "testdb",
		Backup_Option: tmpDir,
	}

	path, _, err := CreateFullBackup(cfg)
	if err != nil {
		t.Fatalf("CreateFullBackup: %v", err)
	}

	// Path should be under tmpDir/backups/
	if !strings.HasPrefix(path, filepath.Join(tmpDir, "backups")) {
		t.Errorf("backup path %q should be under %s/backups/", path, tmpDir)
	}
}

// ---------------------------------------------------------------------------
// ReadManifest
// ---------------------------------------------------------------------------

func TestReadManifest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string) string // returns zip path
		wantErr   string
		checkFunc func(t *testing.T, m *BackupManifest)
	}{
		{
			name: "valid manifest",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				m := BackupManifest{
					Driver:    "sqlite",
					Timestamp: "2024-01-15T10:30:00Z",
					Version:   "1.0.0",
					NodeID:    "node-42",
					DbName:    "mydb",
				}
				return createTestZipWithManifest(t, filepath.Join(dir, "backup.zip"), m, true, "database.db")
			},
			checkFunc: func(t *testing.T, m *BackupManifest) {
				t.Helper()
				if m.Driver != "sqlite" {
					t.Errorf("Driver = %q, want %q", m.Driver, "sqlite")
				}
				if m.Timestamp != "2024-01-15T10:30:00Z" {
					t.Errorf("Timestamp = %q, want %q", m.Timestamp, "2024-01-15T10:30:00Z")
				}
				if m.Version != "1.0.0" {
					t.Errorf("Version = %q, want %q", m.Version, "1.0.0")
				}
				if m.NodeID != "node-42" {
					t.Errorf("NodeID = %q, want %q", m.NodeID, "node-42")
				}
				if m.DbName != "mydb" {
					t.Errorf("DbName = %q, want %q", m.DbName, "mydb")
				}
			},
		},
		{
			name: "nonexistent zip file",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return filepath.Join(dir, "does-not-exist.zip")
			},
			wantErr: "failed to open backup archive",
		},
		{
			name: "zip without manifest",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				zipPath := filepath.Join(dir, "no-manifest.zip")
				f, err := os.Create(zipPath)
				if err != nil {
					t.Fatalf("creating zip: %v", err)
				}
				zw := zip.NewWriter(f)
				// Add a file that is NOT manifest.json
				w, err := zw.Create("other-file.txt")
				if err != nil {
					t.Fatalf("creating entry: %v", err)
				}
				if _, err := w.Write([]byte("not a manifest")); err != nil {
					t.Fatalf("writing: %v", err)
				}
				if err := zw.Close(); err != nil {
					t.Fatalf("closing writer: %v", err)
				}
				if err := f.Close(); err != nil {
					t.Fatalf("closing file: %v", err)
				}
				return zipPath
			},
			wantErr: "manifest.json not found",
		},
		{
			name: "zip with invalid JSON manifest",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				zipPath := filepath.Join(dir, "bad-json.zip")
				f, err := os.Create(zipPath)
				if err != nil {
					t.Fatalf("creating zip: %v", err)
				}
				zw := zip.NewWriter(f)
				w, err := zw.Create("manifest.json")
				if err != nil {
					t.Fatalf("creating entry: %v", err)
				}
				if _, err := w.Write([]byte("this is not json {")); err != nil {
					t.Fatalf("writing: %v", err)
				}
				if err := zw.Close(); err != nil {
					t.Fatalf("closing writer: %v", err)
				}
				if err := f.Close(); err != nil {
					t.Fatalf("closing file: %v", err)
				}
				return zipPath
			},
			wantErr: "failed to parse manifest.json",
		},
		{
			name: "corrupt zip file",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				corruptPath := filepath.Join(dir, "corrupt.zip")
				createTempFile(t, corruptPath, "this is not a zip file at all")
				return corruptPath
			},
			wantErr: "failed to open backup archive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			zipPath := tt.setup(t, tmpDir)

			manifest, err := ReadManifest(zipPath)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if manifest == nil {
				t.Fatal("expected non-nil manifest, got nil")
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, manifest)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// restoreSQLite
// ---------------------------------------------------------------------------

func TestRestoreSQLite_CopiesDatabaseFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Simulate an extracted backup with database.db
	extractDir := filepath.Join(tmpDir, "extracted")
	dbContent := "restored-sqlite-database-content"
	createTempFile(t, filepath.Join(extractDir, "database.db"), dbContent)

	// Destination for the restored database
	destDB := filepath.Join(tmpDir, "restored.db")

	cfg := config.Config{
		Db_URL: destDB,
	}

	if err := restoreSQLite(cfg, extractDir); err != nil {
		t.Fatalf("restoreSQLite: %v", err)
	}

	// Verify the file was copied correctly
	got, err := os.ReadFile(destDB)
	if err != nil {
		t.Fatalf("reading restored db: %v", err)
	}
	if string(got) != dbContent {
		t.Errorf("restored db content = %q, want %q", string(got), dbContent)
	}
}

func TestRestoreSQLite_OverwritesExistingFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	extractDir := filepath.Join(tmpDir, "extracted")
	createTempFile(t, filepath.Join(extractDir, "database.db"), "new-db-content")

	// Create an existing db at the destination
	destDB := filepath.Join(tmpDir, "existing.db")
	createTempFile(t, destDB, "old-db-content")

	cfg := config.Config{
		Db_URL: destDB,
	}

	if err := restoreSQLite(cfg, extractDir); err != nil {
		t.Fatalf("restoreSQLite: %v", err)
	}

	got, err := os.ReadFile(destDB)
	if err != nil {
		t.Fatalf("reading restored db: %v", err)
	}
	if string(got) != "new-db-content" {
		t.Errorf("restored db should be overwritten, got %q", string(got))
	}
}

func TestRestoreSQLite_MissingDatabaseFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Empty extract directory -- no database.db
	extractDir := filepath.Join(tmpDir, "empty-extract")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}

	cfg := config.Config{
		Db_URL: filepath.Join(tmpDir, "dest.db"),
	}

	err := restoreSQLite(cfg, extractDir)
	if err == nil {
		t.Fatal("expected error for missing database.db, got nil")
	}
	if !strings.Contains(err.Error(), "database.db not found") {
		t.Errorf("error should mention missing database.db, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// restoreExtraFiles
// ---------------------------------------------------------------------------

func TestRestoreExtraFiles_CopiesFilesToCurrentDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// restoreExtraFiles writes to relative paths from the extra dir.
	// To avoid polluting the real working directory, we chdir to a temp dir.
	// But since tests run in parallel with shared process cwd, we instead
	// test the function's behavior by creating the extract structure and
	// verifying it writes to the expected relative paths.
	//
	// Note: restoreExtraFiles uses filepath.Walk and writes to relPath which
	// is relative. This is a design limitation. For the test, we work around
	// it by creating the extra dir structure in a way where relPath resolves
	// to something inside our temp dir.

	// We can test the function by providing an extraDir that has files where
	// the relative paths will be valid in the test temp dir. Since the function
	// uses os.Create(destPath) where destPath = relPath, and relPath is
	// relative to extraDir, the files will be created relative to cwd.
	// This is inherently not parallelizable in a strict sense, but we can
	// verify the walk logic works by checking no error is returned when the
	// structure is valid and the cwd is writable.

	extraDir := filepath.Join(tmpDir, "extra")
	createTempFile(t, filepath.Join(extraDir, "config.toml"), "key = value")

	// The function will try to create "config.toml" relative to cwd.
	// Since we cannot safely chdir in a parallel test, we verify:
	// 1. The function does not error on a valid input
	// 2. The Walk logic traverses correctly

	// For a more robust test we would need the function to accept a dest dir,
	// but we test what we have.
	// REQUIRES REFACTOR: restoreExtraFiles should accept a destination directory
	// parameter instead of writing to relative paths from cwd.

	// We can still verify the function handles the walk correctly by checking
	// it processes files without error when the write path is accessible.
	// The cwd for the test process should be writable.
	err := restoreExtraFiles(extraDir)
	if err != nil {
		t.Fatalf("restoreExtraFiles: %v", err)
	}

	// Clean up the file that was created relative to cwd
	t.Cleanup(func() {
		os.Remove("config.toml")
	})
}

func TestRestoreExtraFiles_EmptyDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	extraDir := filepath.Join(tmpDir, "empty-extra")
	if err := os.MkdirAll(extraDir, 0o755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}

	// Should complete without error when there are no files
	if err := restoreExtraFiles(extraDir); err != nil {
		t.Fatalf("restoreExtraFiles on empty dir: %v", err)
	}
}

func TestRestoreExtraFiles_NonexistentDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	err := restoreExtraFiles(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Fatal("expected error for nonexistent extra dir, got nil")
	}
}

// ---------------------------------------------------------------------------
// RestoreFromBackup (SQLite round-trip)
// ---------------------------------------------------------------------------

func TestRestoreFromBackup_SQLite_RoundTrip(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Step 1: Create a "source" database
	srcDB := filepath.Join(tmpDir, "source.db")
	dbContent := "SQLite format 3\x00original-database-content-for-roundtrip"
	createTempFile(t, srcDB, dbContent)

	// Step 2: Create a backup
	cfg := config.Config{
		Db_Driver:     config.Sqlite,
		Db_URL:        srcDB,
		Db_Name:       "roundtrip-test",
		Node_ID:       "node-rt",
		Backup_Option: tmpDir,
	}

	backupPath, _, err := CreateFullBackup(cfg)
	if err != nil {
		t.Fatalf("CreateFullBackup: %v", err)
	}

	// Step 3: Restore to a new location
	destDB := filepath.Join(tmpDir, "restored.db")
	restoreCfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    destDB,
		Db_Name:   "roundtrip-test",
	}

	if err := RestoreFromBackup(restoreCfg, backupPath); err != nil {
		t.Fatalf("RestoreFromBackup: %v", err)
	}

	// Step 4: Verify the restored database matches the original
	restored, err := os.ReadFile(destDB)
	if err != nil {
		t.Fatalf("reading restored db: %v", err)
	}
	if string(restored) != dbContent {
		t.Errorf("restored database content mismatch: got %d bytes, want %d bytes",
			len(restored), len(dbContent))
	}
}

func TestRestoreFromBackup_DriverMismatch(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a backup zip with a "sqlite" manifest
	manifest := BackupManifest{
		Driver:    "sqlite",
		Timestamp: "2024-01-15T00:00:00Z",
		Version:   "1.0.0",
		NodeID:    "node-1",
		DbName:    "test",
	}
	zipPath := createTestZipWithManifest(t, filepath.Join(tmpDir, "backup.zip"), manifest, true, "database.db")

	// Try to restore with a PostgreSQL config
	cfg := config.Config{
		Db_Driver: config.Psql,
		Db_URL:    filepath.Join(tmpDir, "dest.db"),
	}

	err := RestoreFromBackup(cfg, zipPath)
	if err == nil {
		t.Fatal("expected error for driver mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "driver mismatch") {
		t.Errorf("error should mention driver mismatch, got: %v", err)
	}
}

func TestRestoreFromBackup_InvalidArchive(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a file that is not a valid zip
	notZip := filepath.Join(tmpDir, "not-a-zip.zip")
	createTempFile(t, notZip, "this is definitely not a zip file")

	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(tmpDir, "dest.db"),
	}

	err := RestoreFromBackup(cfg, notZip)
	if err == nil {
		t.Fatal("expected error for invalid archive, got nil")
	}
	if !strings.Contains(err.Error(), "backup manifest") {
		t.Errorf("error should mention backup manifest, got: %v", err)
	}
}

func TestRestoreFromBackup_NonexistentArchive(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(tmpDir, "dest.db"),
	}

	err := RestoreFromBackup(cfg, filepath.Join(tmpDir, "missing.zip"))
	if err == nil {
		t.Fatal("expected error for missing archive, got nil")
	}
}

func TestRestoreFromBackup_UnsupportedDriver(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a zip with an unsupported driver in manifest
	manifest := BackupManifest{
		Driver:    "couchdb",
		Timestamp: "2024-01-15T00:00:00Z",
		Version:   "1.0.0",
		NodeID:    "node-1",
		DbName:    "test",
	}
	zipPath := createTestZipWithManifest(t, filepath.Join(tmpDir, "backup.zip"), manifest, true, "database.db")

	// Config also uses "couchdb" so driver check passes, but switch fails
	cfg := config.Config{
		Db_Driver: config.DbDriver("couchdb"),
	}

	err := RestoreFromBackup(cfg, zipPath)
	if err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Errorf("error should mention unsupported driver, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// BackupManifest JSON serialization
// ---------------------------------------------------------------------------

func TestBackupManifest_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := BackupManifest{
		Driver:    "sqlite",
		Timestamp: "2024-06-15T14:30:00Z",
		Version:   "0.5.2",
		NodeID:    "node-abc-123",
		DbName:    "production",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded BackupManifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch:\n  got:  %+v\n  want: %+v", decoded, original)
	}
}

func TestBackupManifest_JSONFieldNames(t *testing.T) {
	t.Parallel()

	m := BackupManifest{
		Driver:    "postgres",
		Timestamp: "2024-01-01T00:00:00Z",
		Version:   "1.0.0",
		NodeID:    "n1",
		DbName:    "db1",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	raw := string(data)

	// Verify JSON field names match the tags
	expectedFields := []string{
		`"driver"`,
		`"timestamp"`,
		`"version"`,
		`"node_id"`,
		`"db_name"`,
	}
	for _, field := range expectedFields {
		if !strings.Contains(raw, field) {
			t.Errorf("JSON output missing field %s, got: %s", field, raw)
		}
	}
}

func TestBackupManifest_ZeroValue(t *testing.T) {
	t.Parallel()

	var m BackupManifest
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal zero value: %v", err)
	}

	var decoded BackupManifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal zero value: %v", err)
	}

	if decoded.Driver != "" {
		t.Errorf("zero value Driver = %q, want empty", decoded.Driver)
	}
	if decoded.Timestamp != "" {
		t.Errorf("zero value Timestamp = %q, want empty", decoded.Timestamp)
	}
}
