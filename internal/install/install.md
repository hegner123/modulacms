# install

Package install provides interactive installation and setup for ModulaCMS. It handles configuration file creation, database initialization, credential validation, and interactive user prompts for all required settings.

## Overview

The install package orchestrates the ModulaCMS installation process through an interactive terminal UI built with Charmbracelet Huh. It collects configuration values, validates them, writes the config file, creates database tables, and verifies external service connections.

Key responsibilities include collecting database credentials for SQLite, MySQL, or PostgreSQL, configuring S3-compatible bucket storage, setting up OAuth providers, validating network ports and domains, creating initial admin user with password hashing, and providing retry logic for failed installations.

The package supports both interactive mode with forms and non-interactive mode for automation via command-line flags.

## Constants

### maxRetries

Maximum number of installation retry attempts allowed when installation fails. Set to 3.

## Types

### DatabaseDriver

DatabaseDriver is a string enum representing supported database systems.

```go
type DatabaseDriver string

const (
    SQLITE   DatabaseDriver = "sqlite"
    MYSQL    DatabaseDriver = "mysql"
    POSTGRES DatabaseDriver = "postgres"
)
```

Valid values are sqlite for SQLite3, mysql for MySQL 5.7 or later, and postgres for PostgreSQL 12 or later.

### InstallArguments

InstallArguments holds all configuration values collected during installation.

```go
type InstallArguments struct {
    UseDefaultConfig  bool
    ConfigPath        string
    Config            config.Config
    DB_Driver         DatabaseDriver
    Create_Tables     bool
    AdminPasswordHash string
}
```

The AdminPasswordHash field is omitted from JSON serialization for security. All other fields map to the final config.json structure.

### InstallError

InstallError provides structured error information with context and user hints.

```go
type InstallError struct {
    Operation string
    Cause     error
    Hint      string
}
```

The Operation field describes what was being attempted. Cause wraps the underlying error. Hint provides actionable guidance for resolving the issue.

### InstallProgress

InstallProgress manages a sequence of installation steps with progress indicators.

```go
type InstallProgress struct {
    steps        []Step
    successStyle lipgloss.Style
    failStyle    lipgloss.Style
    stepStyle    lipgloss.Style
}
```

Each step runs sequentially with a spinner. Styles control terminal output colors for success, failure, and in-progress states.

### ModulaInit

ModulaInit represents the installation verification status across all components.

```go
type ModulaInit struct {
    UseSSL          bool
    DbFileExists    bool
    Certificates    bool
    Key             bool
    ConfigExists    bool
    DBConnected     bool
    BucketConnected bool
    OauthConnected  bool
}
```

All boolean fields default to false. CheckInstall populates these based on actual connection tests and file existence checks.

### Step

Step represents a single installation step with a name, description, and action function.

```go
type Step struct {
    Name        string
    Description string
    Action      func() error
}
```

The Action function executes the step logic. Returning an error stops the installation sequence unless using RunWithWarnings.

### DBStatus

DBStatus reports database connection check results.

```go
type DBStatus struct {
    Driver string
    URL    string
    Err    error
}
```

Driver contains the database type string. URL contains the connection string or file path. Err is nil on success.

## Error Constructors

### ErrBucketConnect

```go
func ErrBucketConnect(cause error) *InstallError
```

ErrBucketConnect creates an error for S3 bucket connection failures. Hint recommends verifying access key, secret key, endpoint URL, and checking bucket existence and accessibility.

### ErrConfigWrite

```go
func ErrConfigWrite(cause error, path string) *InstallError
```

ErrConfigWrite creates an error for config file write failures. Hint suggests checking write permissions and directory existence for the given path.

### ErrDBBootstrap

```go
func ErrDBBootstrap(cause error) *InstallError
```

ErrDBBootstrap creates an error for bootstrap data insertion failures. Hint suggests the database may already contain data and recommends using a fresh database.

### ErrDBConnect

```go
func ErrDBConnect(cause error, driver string) *InstallError
```

ErrDBConnect creates an error for database connection failures. For sqlite driver, hint checks path writability and .db extension. For other drivers, hint checks credentials and server running status.

### ErrDBTables

```go
func ErrDBTables(cause error) *InstallError
```

ErrDBTables creates an error for table creation failures. Hint recommends checking CREATE TABLE privileges for MySQL and PostgreSQL.

### ErrMaxRetries

```go
func ErrMaxRetries(attempts int) *InstallError
```

ErrMaxRetries creates an error when max install retries exceeded. Hint instructs user to review errors and fix configuration before running install again.

### ErrUserAborted

```go
func ErrUserAborted() *InstallError
```

ErrUserAborted creates an error when user cancels the installation. No hint provided.

### ErrValidation

```go
func ErrValidation(field string, cause error) *InstallError
```

ErrValidation creates an error for input validation failures. The field parameter identifies which input failed validation.

## Installation Checks

### CheckBackupTools

```go
func CheckBackupTools(driver config.DbDriver) (warning string, err error)
```

CheckBackupTools verifies that the required database client tool is available in PATH for the configured database driver. Returns empty warning and nil error if OK. Returns error if pg_dump is missing for PostgreSQL or mysqldump is missing for MySQL. SQLite requires no external tool.

### CheckBucket

```go
func CheckBucket(v *bool, c *config.Config) (string, error)
```

CheckBucket validates S3 bucket connectivity using credentials in config. Returns status string and error. If credentials are empty, returns "Not configured" with nil error since bucket is optional. If verbose is true, logs warnings or success.

### CheckCerts

```go
func CheckCerts(path string) bool
```

CheckCerts verifies that localhost.crt and localhost.key exist in the given path. Returns true only if both files exist.

### CheckConfigExists

```go
func CheckConfigExists(path string) error
```

CheckConfigExists checks if config file exists at the given path. Uses config.json as default if path is empty. Logs success and returns nil if file exists.

### CheckDb

```go
func CheckDb(v *bool, c config.Config) (DBStatus, error)
```

CheckDb attempts to connect to the database using credentials in config. Returns DBStatus with driver, URL, and error status. If verbose is true, logs connection attempts and results.

### CheckInstall

```go
func CheckInstall(c *config.Config, v *bool) (ModulaInit, error)
```

CheckInstall runs all pre-flight checks and returns ModulaInit status. Checks config existence, database connection, bucket connection, OAuth configuration, and SSL certificates. Bucket and OAuth failures are logged as warnings but do not fail the overall check since they are optional.

### CheckOauth

```go
func CheckOauth(v *bool, c *config.Config) (string, error)
```

CheckOauth validates OAuth provider configuration by checking for required client ID, client secret, auth URL, and token URL. Returns "Not configured" with nil error if fields are empty since OAuth is optional. If verbose is true, logs warnings or success.

## Database Creation

### CreateDb

```go
func CreateDb(path string, c *config.Config, adminHash string) error
```

CreateDb creates database tables and bootstrap data with progress indicators. Runs three steps in sequence: create all tables, insert bootstrap data including admin user with provided password hash, and validate bootstrap data integrity. Uses InstallProgress for visual feedback.

### CreateDbSimple

```go
func CreateDbSimple(path string, c *config.Config, adminHash string) error
```

CreateDbSimple creates database without progress indicators for programmatic use. Executes the same operations as CreateDb but without spinner UI. Suitable for automated scripts and testing.

### CreateDefaultConfig

```go
func CreateDefaultConfig(path string) error
```

CreateDefaultConfig writes a default config.json file at the given path. Creates the file if it does not exist. Truncates and overwrites if it does exist. Returns ErrConfigWrite on failure.

## Interactive Forms

### GetAdminPassword

```go
func GetAdminPassword(i *InstallArguments) error
```

GetAdminPassword prompts for the system admin password with confirmation, validates it meets requirements, and stores the bcrypt hash in InstallArguments. Requires minimum 8 characters. Returns error if passwords do not match or hashing fails.

### GetBuckets

```go
func GetBuckets(i *InstallArguments) error
```

GetBuckets prompts for S3-compatible bucket configuration including access key, secret key, region, endpoint URL, media path, backup path, and path-style addressing preference. All fields are optional. Sets Config bucket fields.

### GetCertDir

```go
func GetCertDir(i *InstallArguments) error
```

GetCertDir prompts for certificate directory path where localhost.crt and localhost.key will be stored or read. Validates that path is an accessible directory. Defaults to current directory.

### GetConfigPath

```go
func GetConfigPath(i *InstallArguments) error
```

GetConfigPath prompts for config file save location. Validates that parent directory exists and is writable. Defaults to config.json in current directory.

### GetCookie

```go
func GetCookie(i *InstallArguments) error
```

GetCookie prompts for cookie name used for session management. Validates name contains only alphanumeric characters, underscores, and hyphens. Defaults to modula_cms.

### GetCORS

```go
func GetCORS(i *InstallArguments) error
```

GetCORS prompts for CORS allowed origins as comma-separated list and whether to allow credentials. Automatically sets methods to GET, POST, PUT, DELETE, OPTIONS and headers to Content-Type and Authorization.

### GetDbDriver

```go
func GetDbDriver(i *InstallArguments) error
```

GetDbDriver prompts user to select database driver from Sqlite, MySql, or Postgres. Sets both DB_Driver and Config.Db_Driver based on selection.

### GetDomains

```go
func GetDomains(i *InstallArguments) error
```

GetDomains prompts for client site domain and admin site domain. Validates URL format. Defaults both to localhost.

### GetEnvironments

```go
func GetEnvironments(i *InstallArguments) error
```

GetEnvironments prompts for development, staging, and production URLs. Stores them in Config.Environment_Hosts map with keys development, staging, and production.

### GetFullSqlSetup

```go
func GetFullSqlSetup(i *InstallArguments) error
```

GetFullSqlSetup prompts for MySQL or PostgreSQL connection parameters including host URL, database name, username, and password. Generates a random password as default. Validates URL and database name format.

### GetLiteSqlSetup

```go
func GetLiteSqlSetup(i *InstallArguments) error
```

GetLiteSqlSetup prompts for SQLite database file path and database name. Validates file path has .db extension. Defaults to modula.db in current directory.

### GetOAuth

```go
func GetOAuth(i *InstallArguments) error
```

GetOAuth prompts for full OAuth provider configuration including provider name, client ID and secret, authorization URL, token URL, user info URL, redirect URL, scopes as comma-separated list, and success redirect path. Validates all URL fields.

### GetOAuthOptional

```go
func GetOAuthOptional(i *InstallArguments) error
```

GetOAuthOptional asks if user wants to configure OAuth. If yes, calls GetOAuth. If no, sets all OAuth config fields to empty values.

### GetOutputFormat

```go
func GetOutputFormat(i *InstallArguments) error
```

GetOutputFormat prompts user to select API output format from raw, clean, contentful, sanity, strapi, or wordpress. Sets Config.Output_Format.

### GetPorts

```go
func GetPorts(i *InstallArguments) error
```

GetPorts prompts for HTTP port, HTTPS port, and SSH port. Validates each port is numeric and in range 1024 to 65535. Defaults to 1234, 4000, and 2233 respectively.

### GetUseDefault

```go
func GetUseDefault(i *InstallArguments) error
```

GetUseDefault prompts user whether to use default configuration. Sets InstallArguments.UseDefaultConfig to true or false based on selection.

## Installation Entry Points

### RunInstall

```go
func RunInstall(v *bool, yes *bool, adminPassword *string) error
```

RunInstall runs the interactive installation process with retry support. When yes is non-nil and true, all prompts are skipped and defaults are used. adminPassword provides the system admin password and is required for non-interactive mode. Uses maxRetries for retry limit.

### RunInstallIO

```go
func RunInstallIO() (*InstallArguments, error)
```

RunInstallIO executes the full interactive installation form sequence. Calls all Get functions in order: default config choice, config path, environments, ports, domains, CORS, cert dir, cookie, output format, database driver, database setup, buckets, OAuth, and admin password. Returns populated InstallArguments or error if user aborts.

## Progress Display

### NewInstallProgress

```go
func NewInstallProgress() *InstallProgress
```

NewInstallProgress creates a new progress tracker with predefined color styles. Success style is green, fail style is red, and step style is gray.

### AddStep

```go
func (p *InstallProgress) AddStep(name, description string, action func() error) *InstallProgress
```

AddStep adds a step to the progress tracker. Returns the InstallProgress pointer for method chaining.

### Run

```go
func (p *InstallProgress) Run() error
```

Run executes all steps with spinner feedback. Displays progress as [current/total] prefix. On success, prints checkmark. On failure, prints X and returns error immediately. Does not continue to next step after failure.

### RunWithWarnings

```go
func (p *InstallProgress) RunWithWarnings() (warnings []string, err error)
```

RunWithWarnings executes all steps, collecting warnings but not stopping on non-critical failures. Continues to next step even if current step fails. Returns slice of warning strings and error. Only returns error if spinner itself fails.

### PrintError

```go
func PrintError(msg string)
```

PrintError prints an error message in red with X prefix.

### PrintSuccess

```go
func PrintSuccess(msg string)
```

PrintSuccess prints a success message in green with checkmark prefix.

### PrintWarning

```go
func PrintWarning(msg string)
```

PrintWarning prints a warning message in orange with exclamation point prefix.

## Validation Functions

### ValidateCookieName

```go
func ValidateCookieName(s string) error
```

ValidateCookieName checks that a cookie name contains only alphanumeric characters, underscores, and hyphens. Returns error if empty or contains invalid characters.

### ValidateConfigPath

```go
func ValidateConfigPath(s string) error
```

ValidateConfigPath checks if the config path is valid and writable. Validates parent directory exists and is accessible. Allows overwriting existing file. Returns error if path is empty or directory does not exist.

### ValidateDBName

```go
func ValidateDBName(s string) error
```

ValidateDBName checks database name contains only alphanumeric characters and underscores. Returns error if empty or contains special characters.

### ValidateDBPath

```go
func ValidateDBPath(s string) error
```

ValidateDBPath checks that the path has a .db extension for SQLite databases. Returns error if empty or extension is not .db.

### ValidateDirPath

```go
func ValidateDirPath(s string) error
```

ValidateDirPath checks that a directory path exists and is accessible. Returns error if path is empty, does not exist, or is not a directory.

### ValidateNotEmpty

```go
func ValidateNotEmpty(fieldName string) func(string) error
```

ValidateNotEmpty returns a validation function that checks a field is not empty or whitespace-only. The fieldName parameter customizes the error message.

### ValidatePassword

```go
func ValidatePassword(s string) error
```

ValidatePassword checks that a password meets minimum requirements. Requires at least 8 characters and no more than 72 bytes due to bcrypt limit.

### ValidatePort

```go
func ValidatePort(s string) error
```

ValidatePort checks if the port is numeric and within the valid range 1024 to 65535. Returns error if empty, non-numeric, or out of range.

### ValidatePortOrEmpty

```go
func ValidatePortOrEmpty(s string) error
```

ValidatePortOrEmpty allows empty string for optional ports or validates port if provided. Returns nil for empty string.

### ValidateURL

```go
func ValidateURL(s string) error
```

ValidateURL performs basic URL format validation. Allows simple hostnames like localhost without scheme. If scheme is present, validates host is not empty. Returns error if empty or invalid format.
