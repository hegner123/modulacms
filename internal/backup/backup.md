# backup

Package backup provides backup and restore functionality for ModulaCMS databases. It supports SQLite, MySQL, and PostgreSQL database drivers, creating zip archives containing database dumps and optional extra file paths. Backup archives include a manifest file with metadata for verification during restore operations.

## Types

### BackupManifest

BackupManifest records metadata about a backup archive. It contains information about the database driver, timestamp, ModulaCMS version, node identifier, and database name. This metadata is stored as manifest.json inside each backup archive and is used during restore operations to verify compatibility between the backup and the target system.

```go
type BackupManifest struct {
	Driver    string `json:"driver"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	NodeID    string `json:"node_id"`
	DbName    string `json:"db_name"`
}
```

## Functions

#### CreateFullBackup

CreateFullBackup creates a zip archive containing the database and any configured backup paths. It returns the file path, size in bytes, or an error. The function generates a timestamped filename in the format backup_YYYYMMDD_HHMMSS.zip and stores it in the configured backup directory. The backup includes the database dump or file based on the driver type, a manifest.json file with backup metadata, and any additional paths specified in the configuration.

```go
func CreateFullBackup(cfg config.Config) (path string, sizeBytes int64, err error)
```

For SQLite databases, the entire database file is copied into the archive as database.db. For MySQL and PostgreSQL databases, the function executes mysqldump or pg_dump respectively and stores the SQL output as database.sql. Extra backup paths from the configuration are added to an extra subdirectory within the archive, preserving their relative structure.

#### ReadManifest

ReadManifest extracts and parses the manifest.json from a backup archive. It opens the zip file, locates the manifest entry, reads the JSON data, and returns a BackupManifest struct. This function is used during restore operations to verify that the backup is compatible with the current database configuration before attempting restoration.

```go
func ReadManifest(backupPath string) (*BackupManifest, error)
```

Returns an error if the archive cannot be opened, if manifest.json is not found, or if the JSON data cannot be parsed.

#### RestoreFromBackup

RestoreFromBackup restores a backup archive to the configured database. For SQLite, this replaces the database file. For MySQL and PostgreSQL, this pipes the SQL dump into the respective client. The function first reads and validates the manifest to ensure the backup driver matches the current configuration, then extracts the archive to a temporary directory and performs the driver-specific restore operation.

```go
func RestoreFromBackup(cfg config.Config, backupPath string) error
```

If the backup archive contains extra files in the extra directory, these are restored to their original relative paths. The temporary extraction directory is cleaned up after the restore completes or if an error occurs. Returns an error if the manifest cannot be read, if there is a driver mismatch, or if the restore operation fails.

#### TimestampBackupName

TimestampBackupName generates a backup filename by appending a timestamp and the .zip extension to the provided output prefix. The format is output_timestamp.zip where timestamp is provided as a string parameter.

```go
func TimestampBackupName(output string, timestamp string) string
```

## Internal Functions

#### addSQLiteDB

addSQLiteDB adds a SQLite database file to the zip archive. It opens the database file at the path specified in dbPath, creates an entry named database.db in the zip archive, and copies the database contents to the archive. Returns an error if the database file does not exist, cannot be opened, or if the copy operation fails.

```go
func addSQLiteDB(zw *zip.Writer, dbPath string) error
```

#### addMySQLDump

addMySQLDump executes mysqldump to create a database backup and adds the SQL output to the zip archive. It parses the database URL to extract host and port, defaulting to port 3306 if not specified. The function runs mysqldump with the configured credentials and database name, captures the output, and writes it to database.sql in the archive.

```go
func addMySQLDump(zw *zip.Writer, cfg config.Config) error
```

Returns an error if mysqldump execution fails or if the output cannot be written to the archive. Stderr output from mysqldump is included in the error message for debugging.

#### addPostgresDump

addPostgresDump executes pg_dump to create a database backup and adds the SQL output to the zip archive. It parses the database URL to extract host and port, defaulting to port 5432 if not specified. The function runs pg_dump with the configured credentials and database name, passing the password via the PGPASSWORD environment variable.

```go
func addPostgresDump(zw *zip.Writer, cfg config.Config) error
```

The SQL output is written to database.sql in the archive. Returns an error if pg_dump execution fails or if the output cannot be written. Stderr output from pg_dump is included in the error message for debugging.

#### addFilesToZip

addFilesToZip recursively walks a directory and adds all files to the zip archive under the specified base path. Files are added with their relative paths preserved under baseInZip. Directories themselves are not added as separate entries. The function skips any paths that encounter errors during traversal.

```go
func addFilesToZip(zipWriter *zip.Writer, dir, baseInZip string) error
```

Returns an error if file operations fail during the walk, including failures to determine relative paths, create zip entries, or copy file contents.

#### restoreSQLite

restoreSQLite restores a SQLite database from the extracted backup. It locates database.db in the temporary directory, opens it, and copies its contents to the configured database path in the configuration. The destination database file is created or overwritten during this operation.

```go
func restoreSQLite(cfg config.Config, tempDir string) error
```

Returns an error if database.db is not found in the backup, if the source or destination files cannot be opened, or if the copy operation fails.

#### restoreMySQL

restoreMySQL restores a MySQL database by piping the SQL dump into the mysql command-line client. It locates database.sql in the temporary directory, opens it, and streams its contents to mysql with the configured username, password, and database name. Stderr from mysql is directed to the process stderr for visibility.

```go
func restoreMySQL(cfg config.Config, tempDir string) error
```

Returns an error if database.sql is not found in the backup, if the SQL file cannot be opened, or if the mysql command execution fails.

#### restorePostgres

restorePostgres restores a PostgreSQL database by piping the SQL dump into the psql command-line client. It locates database.sql in the temporary directory, opens it, and streams its contents to psql with the configured username and database name. The password is provided via the PGPASSWORD environment variable.

```go
func restorePostgres(cfg config.Config, tempDir string) error
```

Stderr from psql is directed to the process stderr for visibility. Returns an error if database.sql is not found in the backup, if the SQL file cannot be opened, or if the psql command execution fails.

#### restoreExtraFiles

restoreExtraFiles recursively walks the extra directory from the backup archive and restores all files to their original relative paths in the current working directory. It creates necessary directories as it encounters them and copies each file from the extraction directory to its destination.

```go
func restoreExtraFiles(extraDir string) error
```

Returns an error if directory creation fails, if source or destination files cannot be opened, or if the copy operation fails. Directories themselves are not processed, only the files they contain.

#### splitHostPort

splitHostPort splits a host:port string into separate host and port components. If no port is present in the input address, the defaultPort parameter is used. This function wraps net.SplitHostPort and provides a fallback behavior for addresses without explicit ports.

```go
func splitHostPort(addr, defaultPort string) (host, port string)
```

Used internally for parsing database URLs before invoking mysqldump and pg_dump commands.
