# cmd

ModulaCMS command-line interface implementation using cobra for subcommand routing and configuration management.

## Overview

The cmd package provides the CLI entry point for ModulaCMS. It implements commands for server management, installation, database operations, backup and restore, certificate generation, TUI access, configuration validation, and updates. All commands respect global flags for config file path and verbose logging.

The package uses cobra.Command structures with RunE handlers that load configuration, initialize database connections, and delegate to internal packages. Commands handle graceful shutdown via context cancellation and signal handling.

### Global Flags

Persistent flags available to all commands:

```go
--config <path>    Path to configuration file (default: config.json)
-v, --verbose      Enable debug logging
```

## Commands

### Root Command

```go
var rootCmd = &cobra.Command{
	Use:   "modulacms",
	Short: "ModulaCMS - A headless CMS written in Go",
	Long:  "ModulaCMS serves content over HTTP/HTTPS and provides SSH access for backend management.",
	SilenceUsage:  true,
	SilenceErrors: true,
}
```

The root command is the parent for all subcommands. It configures persistent flags and registers all child commands in init().

Subcommands registered: serve, install, version, update, tui, cert, db, config, backup.

#### func Execute

`func Execute() error`

Execute runs the root cobra command and returns any execution error. Called from main() to start the CLI application.

Returns errors from cobra command execution.

```go
func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
```

#### func init

`func init()`

Registers global flags and adds all subcommands to rootCmd. Runs automatically at package initialization.

### serve Command

```go
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP, HTTPS, and SSH servers",
	Long:  "Start all ModulaCMS servers. Use --wizard for interactive setup.",
}
```

Starts HTTP, HTTPS, and SSH servers concurrently. Handles configuration loading, database initialization, plugin system startup, and graceful shutdown on SIGINT or SIGTERM.

If config file is missing and --wizard is not set, runs non-interactive install with generated admin password. If --wizard is true, launches interactive installation wizard.

The command runs install checks before starting servers. If tables are missing, attempts automatic database setup with generated credentials.

Servers run in separate goroutines. Shutdown context has 30-second timeout. Database pools and plugin manager are closed on exit.

Returns errors from server startup, configuration loading, or database initialization.

```go
modulacms serve
modulacms serve --wizard
modulacms serve --config=/etc/modulacms/config.json
```

#### Wizard Flag

```go
var wizard bool
```

Boolean flag for serve command. When true, runs interactive configuration wizard before starting servers.

```go
modulacms serve --wizard
```

#### func newHTTPServer

`func newHTTPServer(addr string, handler http.Handler, tlsConfig *tls.Config) *http.Server`

Creates an http.Server with explicit timeouts and optional TLS configuration. Sets ReadTimeout, WriteTimeout, and IdleTimeout to prevent resource exhaustion.

Timeouts: 15s read, 15s write, 60s idle.

```go
httpServer := newHTTPServer("localhost:8080", mux, nil)
httpsServer := newHTTPServer("localhost:8443", mux, tlsConfig)
```

#### func sanitizeCertDir

`func sanitizeCertDir(configCertDir string) (string, error)`

Validates and normalizes the certificate directory path. Converts to absolute path and verifies the path exists and is a directory.

Returns errors if path is empty, conversion to absolute fails, path does not exist, or path is not a directory.

```go
certDir, err := sanitizeCertDir(cfg.Cert_Dir)
if err != nil {
	return fmt.Errorf("invalid cert dir: %w", err)
}
```

### install Command

```go
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run the installation wizard",
	Long:  "Run the interactive installation process. Use --yes to accept all defaults.",
}
```

Runs the interactive installation wizard. Collects database settings, admin credentials, OAuth configuration, and S3 storage settings. Creates config file and initializes database tables.

With --yes flag, accepts all defaults and requires --admin-password flag for non-interactive installation.

Returns errors from install.RunInstall().

```go
modulacms install
modulacms install --yes --admin-password=SecurePass123
```

#### Install Flags

```go
var installYes bool
var installAdminPassword string
```

Flags for install command:

- installYes: Skip prompts and accept all defaults
- installAdminPassword: System admin password (required when --yes is set)

Registered in init() with BoolVarP and StringVar.

### version Command

```go
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
}
```

Prints full version information to stdout and exits. Version string includes build time, commit hash, and Go version from utility.GetFullVersionInfo().

```go
modulacms version
```

### update Command

```go
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and apply updates",
}
```

Checks for ModulaCMS updates from GitHub releases API. Compares current version against latest stable release. Downloads platform-specific binary if update is available, then replaces the running executable.

Uses runtime.GOOS and runtime.GOARCH to select correct binary asset. Requires restart after update completes.

Returns errors from update.CheckForUpdates(), update.GetDownloadURL(), update.DownloadUpdate(), or update.ApplyUpdate().

```go
modulacms update
```

### tui Command

```go
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal UI without the server",
}
```

Launches the terminal user interface without starting HTTP/HTTPS/SSH servers. Delegates to tuiDefaultCmd by default. Two subcommands available: default (new TUI) and v1 (legacy TUI).

Bare tui command runs the new TUI. Use tui v1 to access the original implementation.

Loads configuration and initializes database before starting TUI. On exit, sends SIGTERM to current process for clean shutdown.

Returns errors from configuration loading, database initialization, or TUI execution.

```go
modulacms tui
modulacms tui default
modulacms tui v1
```

#### tuiDefaultCmd

```go
var tuiDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Launch the new TUI (default)",
}
```

Runs the new Bubbletea-based TUI from internal/tui package. Initializes model with tui.InitialModel() and runs with tui.Run().

#### tuiV1Cmd

```go
var tuiV1Cmd = &cobra.Command{
	Use:   "v1",
	Short: "Launch the original v1 TUI",
}
```

Runs the original TUI from internal/cli package. Initializes model with cli.InitialModel() and runs with cli.CliRun(). Requires DbDriver instance unlike the new TUI.

### cert Command

```go
var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Certificate management commands",
}
```

Parent command for certificate operations. Contains generate subcommand for creating self-signed SSL certificates.

#### certGenerateCmd

```go
var certGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate self-signed SSL certificates for local development",
}
```

Generates self-signed SSL certificate and private key for local HTTPS testing. Reads cert_dir and client_site from config file if available, otherwise uses defaults (./certs and localhost).

Creates certificate files: localhost.crt and localhost.key. Attempts to trust the certificate in the system keychain on macOS.

Returns errors from utility.GenerateSelfSignedCert() or utility.TrustCertificate().

```go
modulacms cert generate
```

### db Command

```go
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}
```

Parent command for database operations. Contains subcommands: init, wipe, wipe-redeploy, reset, export.

#### dbInitCmd

```go
var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create database tables and bootstrap data",
}
```

Creates all database tables and inserts bootstrap data (system admin user, default roles, permissions). Prompts for system admin password with confirmation using huh.NewInput() forms.

Validates password length (minimum 8 characters) and confirms passwords match before hashing. Calls install.CreateDbSimple() to execute schema and bootstrap operations.

Returns errors from configuration loading, password validation, auth.HashPassword(), or install.CreateDbSimple().

```go
modulacms db init
```

#### dbWipeCmd

```go
var dbWipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Drop all database tables (data is lost)",
}
```

Drops all database tables without recreating them. Prompts for confirmation before proceeding. Calls DbDriver.DropAllTables() to execute DROP TABLE statements.

Warning: All data is permanently deleted. Use wipe-redeploy to drop and recreate.

Returns errors from configuration loading, database initialization, or DbDriver.DropAllTables().

```go
modulacms db wipe
```

#### dbWipeRedeployCmd

```go
var dbWipeRedeployCmd = &cobra.Command{
	Use:   "wipe-redeploy",
	Short: "Drop all tables and recreate schema with bootstrap data",
}
```

Drops all tables, recreates schema, and inserts bootstrap data. Prompts for confirmation and new system admin password. Executes DropAllTables(), CreateAllTables(), CreateBootstrapData(), and ValidateBootstrapData() in sequence.

Use this for clean database resets during development or when schema migrations fail.

Returns errors from any step in the wipe-redeploy sequence.

```go
modulacms db wipe-redeploy
```

#### dbResetCmd

```go
var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete the database file (SQLite only)",
}
```

Deletes the SQLite database file. Does not drop tables or validate database state. Only works with SQLite driver.

For MySQL and PostgreSQL, use db wipe or db wipe-redeploy instead.

Returns errors from configuration loading or os.Remove().

```go
modulacms db reset
```

#### dbExportCmd

```go
var dbExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Dump database SQL to file",
}
```

Exports database schema and data to SQL file. Calls DbDriver.DumpSql() which writes to a timestamped file in the backups directory.

Output format depends on database driver (SQLite .dump, MySQL mysqldump, PostgreSQL pg_dump).

Returns errors from configuration loading, database initialization, or DbDriver.DumpSql().

```go
modulacms db export
```

### backup Command

```go
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore commands",
}
```

Parent command for backup operations. Contains subcommands: create, restore, list.

#### backupCreateCmd

```go
var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a full backup of the database and configured paths",
}
```

Creates a full backup archive containing database dump, config file, and configured media paths. Generates backup ID with types.NewBackupID() and records backup metadata to database.

Creates ZIP archive in backups directory. Updates backup record with completion status, duration, and file size. On failure, marks backup as failed with error message.

Returns errors from configuration loading, backup.CreateFullBackup(), or database operations.

```go
modulacms backup create
```

#### backupRestoreCmd

```go
var backupRestoreCmd = &cobra.Command{
	Use:   "restore <path>",
	Short: "Restore from a backup archive",
	Args:  cobra.ExactArgs(1),
}
```

Restores database and files from a backup archive. Reads manifest from ZIP, displays backup metadata, and prompts for confirmation before overwriting current database.

Extracts database dump and restores via driver-specific import. Restores config and media files to original locations.

Returns errors from backup.ReadManifest(), backup.RestoreFromBackup(), or confirmation handling.

```go
modulacms backup restore backups/backup-20240213T120000Z.zip
```

#### backupListCmd

```go
var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup history from the database",
}
```

Lists the 50 most recent backup records from the database. Displays backup ID, type, status, start time, size, and storage path in tabular format.

Calls DbDriver.ListBackups() with limit 50 and offset 0. Returns early if no backups exist.

Returns errors from configuration loading, database initialization, or DbDriver.ListBackups().

```go
modulacms backup list
```

#### func formatBytes

`func formatBytes(b int64) string`

Formats byte count as human-readable size string with appropriate unit (B, KB, MB, GB, TB, PB, EB). Uses 1024-byte units for conversion.

```go
fmt.Println(formatBytes(1536))       // "1.5 KB"
fmt.Println(formatBytes(1048576))    // "1.0 MB"
```

### config Command

```go
var configParentCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
}
```

Parent command for configuration operations. Contains subcommands: show, validate.

#### configShowCmd

```go
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print the loaded configuration as JSON",
}
```

Loads configuration from --config path and prints formatted JSON to stdout. Uses utility.FormatJSON() for pretty-printing.

Returns errors from loadConfig() or utility.FormatJSON().

```go
modulacms config show
modulacms config show --config=/etc/modulacms/config.json
```

#### configValidateCmd

```go
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
}
```

Validates configuration file structure and required fields. Checks for presence of db_driver, db_url, port, and ssh_port. Reports all validation errors with field names.

Returns errors listing missing required fields or loadConfig() errors.

```go
modulacms config validate
```

## Helper Functions

### func configureLogger

`func configureLogger()`

Sets global logger level based on verbose flag. If verbose is true, sets level to DEBUG, otherwise sets to INFO. Called at the start of most command handlers.

Uses utility.DefaultLogger.SetLevel().

```go
func init() {
	installCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}
// Later in command handler
configureLogger()
```

### func loadConfig

`func loadConfig() (*config.Config, error)`

Loads configuration from the path specified by cfgPath global variable. Creates FileProvider and Manager, calls Load(), and returns parsed Config struct.

Returns errors from config loading or parsing.

```go
cfg, err := loadConfig()
if err != nil {
	return fmt.Errorf("loading configuration: %w", err)
}
```

### func loadConfigAndDB

`func loadConfigAndDB() (*config.Config, db.DbDriver, error)`

Loads configuration and initializes the singleton database connection pool. Returns both Config and DbDriver instances.

Caller must defer db.CloseDB() to close the pool when done.

Returns errors from loadConfig() or db.InitDB().

```go
cfg, driver, err := loadConfigAndDB()
if err != nil {
	return err
}
defer func() {
	if cerr := db.CloseDB(); cerr != nil {
		utility.DefaultLogger.Error("Database pool close error", cerr)
	}
}()
```

### func initObservability

`func initObservability(ctx context.Context, cfg *config.Config) func()`

Initializes observability client if observability_enabled is true in configuration. Creates client with Sentry, Datadog, or custom provider. Starts metrics collection and error reporting.

Returns cleanup function that stops the client. If observability is disabled or initialization fails, returns no-op cleanup function.

Caller must defer the returned function to ensure graceful shutdown.

```go
obsCleanup := initObservability(ctx, cfg)
defer obsCleanup()
```

### func initPluginPool

`func initPluginPool(cfg *config.Config) (*sql.DB, func(), error)`

Creates an isolated database connection pool for plugin use. Reads plugin_db_max_open_conns, plugin_db_max_idle_conns, and plugin_db_conn_max_lifetime from config to configure pool limits.

Returns the pool, a cleanup function that closes the pool, and any initialization error. Caller must defer the cleanup function.

Uses db.OpenPool() with db.DefaultPluginPoolConfig() as base configuration.

Returns errors from time.ParseDuration() or db.OpenPool().

```go
pluginPool, pluginPoolCleanup, err := initPluginPool(cfg)
if err != nil {
	return fmt.Errorf("plugin pool init failed: %w", err)
}
defer pluginPoolCleanup()
```

### func initPluginManager

`func initPluginManager(ctx context.Context, cfg *config.Config, pool *sql.DB) *plugin.Manager`

Creates and initializes the Lua plugin manager if plugin_enabled is true. Reads plugin_directory, plugin_max_vms, plugin_timeout, and plugin_max_ops from config.

Calls Manager.LoadAll() to discover and load all .lua files from the plugin directory. Returns nil if plugins are disabled.

Caller must defer Manager.Shutdown() when the returned manager is non-nil. Note that Manager.Shutdown() closes the sql.DB pool internally.

```go
pluginManager := initPluginManager(ctx, cfg, pluginPool)
if pluginManager != nil {
	defer pluginManager.Shutdown(ctx)
}
```

### func logConfigSummary

`func logConfigSummary(cfg *config.Config)`

Logs configuration details to INFO level. Outputs database driver and URL, client and admin site domains, HTTP/HTTPS/SSH ports, environment settings, OAuth provider, S3 storage endpoints, and CORS configuration.

Called after configuration is loaded and before servers start.

```go
logConfigSummary(cfg)
```
