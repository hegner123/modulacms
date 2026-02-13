# utility

Package utility provides shared utilities for the ModulaCMS application including logging, metrics, observability, timestamp handling, certificate generation, and helper functions.

## Overview

The utility package contains common functionality used across ModulaCMS. Key components include a structured logger with level-based filtering, thread-safe metrics collection, database timestamp formatting for multiple database drivers, self-signed certificate generation for local development, nullable type handlers, and string manipulation helpers.

## Constants

### Storage Units

StorageUnit represents byte-based storage sizes. Constants KB, MB, GB, TB provide powers of two for byte calculations.

```go
const (
    KB StorageUnit = 1 << 10  // 1024 bytes
    MB StorageUnit = 1 << 20  // 1048576 bytes
    GB StorageUnit = 1 << 30  // 1073741824 bytes
    TB StorageUnit = 1 << 40  // 1099511627776 bytes
)
```

AppJson constant holds the MIME type string for JSON responses.

```go
const AppJson string = "application/json"
```

### Database Driver Types

DbDriverType constants identify supported database backends.

```go
const (
    DbSqlite DbDriverType = "sqlite"
    DbMysql  DbDriverType = "mysql"
    DbPsql   DbDriverType = "postgres"
)
```

### Metric Names

Predefined metric name constants ensure consistency across the application.

```go
const (
    MetricHTTPRequests      = "http.requests"
    MetricHTTPDuration      = "http.duration"
    MetricHTTPErrors        = "http.errors"
    MetricDBQueries         = "db.queries"
    MetricDBDuration        = "db.duration"
    MetricDBErrors          = "db.errors"
    MetricSSHConnections    = "ssh.connections"
    MetricSSHErrors         = "ssh.errors"
    MetricCacheHits         = "cache.hits"
    MetricCacheMisses       = "cache.misses"
    MetricActiveConnections = "connections.active"
    MetricMemoryUsage       = "memory.usage"
    MetricGoroutines        = "goroutines.count"
)
```

## Variables

### DefaultLogger

DefaultLogger is the package-level logger instance initialized at INFO level. All application logging should use this instance.

```go
var DefaultLogger = NewLogger(INFO)
```

### GlobalMetrics

GlobalMetrics is the default metrics instance for application-wide metric collection.

```go
var GlobalMetrics = NewMetrics()
```

### GlobalObservability

GlobalObservability is the global observability client instance for external provider integration.

```go
var GlobalObservability *ObservabilityClient
```

### Version Variables

Version, GitCommit, and BuildDate are set at compile time via ldflags. Default to development values.

```go
var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
)
```

## Types

### DbDriverType

DbDriverType represents database driver types to avoid import cycles. String-based type with predefined constants.

```go
type DbDriverType string
```

### StorageUnit

StorageUnit represents byte-based storage sizes as int64. Used with SizeInBytes function.

```go
type StorageUnit int64
```

### LogLevel

LogLevel defines the severity of a log message. Values: BLANK, DEBUG, INFO, WARN, ERROR, FATAL.

```go
type LogLevel int

const (
    BLANK LogLevel = iota
    DEBUG
    INFO
    WARN
    ERROR
    FATAL
)
```

### Logger

Logger represents a structured logger with levels, prefix support, and file output. Implements level-based filtering and styled console output.

```go
type Logger struct {
    level   LogLevel
    prefix  string
    logFile *os.File
}
```

### LogLevelStyle

LogLevelStyle defines the style for a specific log level. Contains level name and styling function.

```go
type LogLevelStyle struct {
    LevelName string
    Style     func(string) string
}
```

### Nullable

Nullable is a constraint for all sql.Null types. Used with generic nullable handlers.

```go
type Nullable interface {
    sql.NullInt64 | sql.NullInt32 | sql.NullInt16 | sql.NullString |
    sql.NullByte | sql.NullFloat64 | sql.NullTime | sql.NullBool
}
```

### MetricType

MetricType represents the type of metric. Values: counter, gauge, histogram.

```go
type MetricType string

const (
    MetricTypeCounter   MetricType = "counter"
    MetricTypeGauge     MetricType = "gauge"
    MetricTypeHistogram MetricType = "histogram"
)
```

### Labels

Labels are key-value pairs for dimensional metrics. Used to tag metrics with additional context.

```go
type Labels map[string]string
```

### Metric

Metric represents a single metric data point with name, type, value, labels, and timestamp.

```go
type Metric struct {
    Name      string
    Type      MetricType
    Value     float64
    Labels    Labels
    Timestamp time.Time
}
```

### Metrics

Metrics holds all application metrics with thread-safe operations. Supports counters, gauges, and histograms.

```go
type Metrics struct {
    mu         sync.RWMutex
    counters   map[string]float64
    gauges     map[string]float64
    histograms map[string][]float64
    labels     map[string]Labels
}
```

### ObservabilityProvider

ObservabilityProvider defines the interface for different observability backends. Implement to integrate with external monitoring services.

```go
type ObservabilityProvider interface {
    SendMetric(metric Metric) error
    SendError(err error, context map[string]any) error
    Flush(timeout time.Duration) error
    Close() error
}
```

### ObservabilityClient

ObservabilityClient manages metrics export to external providers. Handles periodic flushing and lifecycle.

```go
type ObservabilityClient struct {
    provider      ObservabilityProvider
    metrics       *Metrics
    flushInterval time.Duration
    stopChan      chan struct{}
    wg            sync.WaitGroup
    enabled       bool
}
```

### ObservabilityConfig

ObservabilityConfig holds configuration for observability setup. Includes provider details, sampling rates, and metadata.

```go
type ObservabilityConfig struct {
    Enabled       bool
    Provider      string
    DSN           string
    Environment   string
    Release       string
    SampleRate    float64
    TracesRate    float64
    SendPII       bool
    Debug         bool
    ServerName    string
    FlushInterval string
    Tags          map[string]string
}
```

### ConsoleProvider

ConsoleProvider logs metrics to console. Useful for development and debugging.

```go
type ConsoleProvider struct{}
```

### SentryProvider

SentryProvider integrates with Sentry for error tracking and performance monitoring. Placeholder for production Sentry SDK integration.

```go
type SentryProvider struct {
    config ObservabilityConfig
}
```

## Functions - File System

### FileExists

FileExists checks if a file exists at the specified path. Returns false for directories.

```go
func FileExists(path string) bool
```

### DirExists

DirExists checks if a directory exists at the specified path. Returns false for files.

```go
func DirExists(path string) bool
```

## Functions - Timestamps

### TimestampI

TimestampI returns the current Unix timestamp as int64.

```go
func TimestampI() int64
```

### TimestampS

TimestampS returns the current Unix timestamp as string.

```go
func TimestampS() string
```

### TimestampReadable

TimestampReadable returns a time string in RFC3339 format. Generic format that works across systems.

```go
func TimestampReadable() string
```

### FormatTimestampForDB

FormatTimestampForDB returns a timestamp string formatted for the specified database. SQLite uses RFC3339, MySQL and PostgreSQL use YYYY-MM-DD HH:MM:SS.

```go
func FormatTimestampForDB(t time.Time, dbDriverType DbDriverType) string
```

### FormatTimestampForDriverString

FormatTimestampForDriverString handles a driver name as string. Wrapper around FormatTimestampForDB.

```go
func FormatTimestampForDriverString(t time.Time, dbDriver string) string
```

### CurrentTimestampForDB

CurrentTimestampForDB returns the current time formatted for the specified database.

```go
func CurrentTimestampForDB(dbDriverType DbDriverType) string
```

### CurrentTimestampForDriverString

CurrentTimestampForDriverString returns the current time formatted for the specified database driver string.

```go
func CurrentTimestampForDriverString(dbDriver string) string
```

### ParseDBTimestamp

ParseDBTimestamp parses a timestamp string from the database based on the database driver type. Returns pointer to time.Time or error.

```go
func ParseDBTimestamp(timestamp string, dbDriverType DbDriverType) (*time.Time, error)
```

### ParseDBTimestampString

ParseDBTimestampString parses a timestamp with a driver name as string. Wrapper around ParseDBTimestamp.

```go
func ParseDBTimestampString(timestamp string, dbDriver string) (*time.Time, error)
```

### ParseTimeReadable

ParseTimeReadable parses an RFC3339 formatted string. Returns nil on error.

```go
func ParseTimeReadable(s string) *time.Time
```

### TokenExpiredTime

TokenExpiredTime returns token expiration timestamp 168 hours from now. Returns both string and int64.

```go
func TokenExpiredTime() (string, int64)
```

### TimestampLessThan

TimestampLessThan checks if the given timestamp string is less than current time. Returns false on parse error.

```go
func TimestampLessThan(a string) bool
```

## Functions - Certificates

### GenerateSelfSignedCert

GenerateSelfSignedCert generates a self-signed certificate for local development. Creates 4096-bit RSA key, valid for 1 year. Writes localhost.crt and localhost.key to certDir.

```go
func GenerateSelfSignedCert(certDir string, domain string) error
```

### TrustCertificate

TrustCertificate provides OS-specific instructions or attempts to trust the certificate. Supports macOS, Linux, and Windows. Prompts user for confirmation before executing trust commands.

```go
func TrustCertificate(certPath string) error
```

## Functions - Storage

### SizeInBytes

SizeInBytes converts a storage value and unit to bytes. Returns int64 byte count.

```go
func SizeInBytes(value int64, unit StorageUnit) int64
```

## Functions - Database

### HandleRowsCloseDeferErr

HandleRowsCloseDeferErr closes database rows and logs any errors. Use with defer.

```go
func HandleRowsCloseDeferErr(r *sql.Rows)
```

### HandleConnectionCloseDeferErr

HandleConnectionCloseDeferErr closes database connection and logs any errors. Use with defer.

```go
func HandleConnectionCloseDeferErr(r *sql.DB)
```

## Functions - Logging

### NewLogFile

NewLogFile creates a log file at debug.log. Returns file handle for Logger initialization.

```go
func NewLogFile() *os.File
```

### NewLogger

NewLogger creates a new logger with the specified minimum level. Initializes with log file.

```go
func NewLogger(level LogLevel) *Logger
```

### Logger.WithPrefix

WithPrefix creates a new logger with the same level but a custom prefix. Useful for component-specific logging.

```go
func (l *Logger) WithPrefix(prefix string) *Logger
```

### Logger.SetLevel

SetLevel changes the logger minimum level. Messages below this level are filtered.

```go
func (l *Logger) SetLevel(level LogLevel)
```

### Logger.Blank

Blank logs a raw message without level decoration. Only logged if level allows BLANK.

```go
func (l *Logger) Blank(message string, args ...any)
```

### Logger.Debug

Debug logs a debug message. Only logged if level allows DEBUG.

```go
func (l *Logger) Debug(message string, args ...any)
```

### Logger.Info

Info logs an informational message. Only logged if level allows INFO.

```go
func (l *Logger) Info(message string, args ...any)
```

### Logger.Warn

Warn logs a warning message with optional error. Only logged if level allows WARN.

```go
func (l *Logger) Warn(message string, err error, args ...any)
```

### Logger.Error

Error logs an error message with optional error. Only logged if level allows ERROR.

```go
func (l *Logger) Error(message string, err error, args ...any)
```

### Logger.Fatal

Fatal logs an error message and exits with code 1. Always executed when level allows FATAL.

```go
func (l *Logger) Fatal(message string, err error, args ...any)
```

### Logger.Fblank

Fblank logs a raw message to file without level decoration.

```go
func (l *Logger) Fblank(message string, args ...any)
```

### Logger.Fdebug

Fdebug logs a debug message to file.

```go
func (l *Logger) Fdebug(message string, args ...any)
```

### Logger.Finfo

Finfo logs an informational message to file.

```go
func (l *Logger) Finfo(message string, args ...any)
```

### Logger.Fwarn

Fwarn logs a warning message to file with optional error.

```go
func (l *Logger) Fwarn(message string, err error, args ...any)
```

### Logger.Ferror

Ferror logs an error message to file with optional error.

```go
func (l *Logger) Ferror(message string, err error, args ...any)
```

### Logger.Ffatal

Ffatal logs an error message to file and exits with code 1.

```go
func (l *Logger) Ffatal(message string, err error, args ...any)
```

## Functions - Metrics

### NewMetrics

NewMetrics creates a new Metrics instance with initialized maps.

```go
func NewMetrics() *Metrics
```

### Metrics.Counter

Counter increments a counter metric by the specified value. Thread-safe.

```go
func (m *Metrics) Counter(name string, value float64, labels Labels)
```

### Metrics.Increment

Increment increments a counter by 1. Convenience wrapper for Counter.

```go
func (m *Metrics) Increment(name string, labels Labels)
```

### Metrics.Gauge

Gauge sets a gauge metric to a specific value. Thread-safe.

```go
func (m *Metrics) Gauge(name string, value float64, labels Labels)
```

### Metrics.Histogram

Histogram records a value in a histogram. Appends to slice of values. Thread-safe.

```go
func (m *Metrics) Histogram(name string, value float64, labels Labels)
```

### Metrics.Timing

Timing records a duration as a histogram value in milliseconds. Convenience wrapper for Histogram.

```go
func (m *Metrics) Timing(name string, duration time.Duration, labels Labels)
```

### Metrics.GetSnapshot

GetSnapshot returns a snapshot of all metrics. For histograms, calculates average. Thread-safe read.

```go
func (m *Metrics) GetSnapshot() map[string]Metric
```

### Metrics.Reset

Reset clears all metrics. Useful for testing or periodic resets. Thread-safe.

```go
func (m *Metrics) Reset()
```

### MeasureTime

MeasureTime executes a function and records its duration to GlobalMetrics. Convenience function for timing operations.

```go
func MeasureTime(name string, labels Labels, fn func())
```

### MeasureTimeCtx

MeasureTimeCtx executes a function and records its duration, returns any error. Convenience function for timing operations with error handling.

```go
func MeasureTimeCtx(name string, labels Labels, fn func() error) error
```

## Functions - Observability

### NewObservabilityClient

NewObservabilityClient creates a new observability client from config. Initializes provider based on config.Provider.

```go
func NewObservabilityClient(config ObservabilityConfig) (*ObservabilityClient, error)
```

### ObservabilityClient.Start

Start begins periodic metric flushing. Runs in background goroutine until context cancelled or stopped.

```go
func (c *ObservabilityClient) Start(ctx context.Context)
```

### ObservabilityClient.Stop

Stop stops the observability client. Performs final flush and closes provider.

```go
func (c *ObservabilityClient) Stop() error
```

### ObservabilityClient.SendError

SendError sends an error to the observability provider. Returns nil if client disabled.

```go
func (c *ObservabilityClient) SendError(err error, context map[string]any) error
```

### NewConsoleProvider

NewConsoleProvider creates a console logging provider for development.

```go
func NewConsoleProvider() *ConsoleProvider
```

### ConsoleProvider.SendMetric

SendMetric logs a metric to console via DefaultLogger.

```go
func (p *ConsoleProvider) SendMetric(metric Metric) error
```

### ConsoleProvider.SendError

SendError logs an error to console via DefaultLogger.

```go
func (p *ConsoleProvider) SendError(err error, context map[string]any) error
```

### ConsoleProvider.Flush

Flush is a no-op for console provider. Returns nil.

```go
func (p *ConsoleProvider) Flush(timeout time.Duration) error
```

### ConsoleProvider.Close

Close is a no-op for console provider. Returns nil.

```go
func (p *ConsoleProvider) Close() error
```

### NewSentryProvider

NewSentryProvider creates a Sentry integration provider. Placeholder for production Sentry SDK.

```go
func NewSentryProvider(config ObservabilityConfig) (*SentryProvider, error)
```

### SentryProvider.SendMetric

SendMetric sends a metric to Sentry. Placeholder logs to debug.

```go
func (p *SentryProvider) SendMetric(metric Metric) error
```

### SentryProvider.SendError

SendError sends an error to Sentry. Placeholder logs to error.

```go
func (p *SentryProvider) SendError(err error, context map[string]any) error
```

### SentryProvider.Flush

Flush ensures all buffered data sent to Sentry. Placeholder returns nil.

```go
func (p *SentryProvider) Flush(timeout time.Duration) error
```

### SentryProvider.Close

Close shuts down Sentry provider. Placeholder returns nil.

```go
func (p *SentryProvider) Close() error
```

### CaptureError

CaptureError sends an error to GlobalObservability if initialized. Always logs via DefaultLogger.

```go
func CaptureError(err error, context map[string]any)
```

## Functions - Nullable Types

### IsNull

IsNull checks if a sql.Null type has a valid value. Generic function supporting all sql.Null types.

```go
func IsNull[T Nullable](value T) bool
```

### NullToString

NullToString converts any sql.Null type to string, returning null if invalid. Generic function supporting all sql.Null types.

```go
func NullToString[T Nullable](value T) string
```

## Functions - String Utilities

### TrimStringEnd

TrimStringEnd removes l characters from the end of str. Returns original if empty.

```go
func TrimStringEnd(str string, l int) string
```

### IsInt

IsInt checks if string s can be parsed as integer. Returns true if valid.

```go
func IsInt(s string) bool
```

### FormatJSON

FormatJSON formats any value as indented JSON string. Returns formatted string or error.

```go
func FormatJSON(b any) (string, error)
```

### MakeRandomString

MakeRandomString generates a cryptographically secure random string. Returns 32 random bytes encoded as base64, 43 characters.

```go
func MakeRandomString() (string, error)
```

### IsValidEmail

IsValidEmail checks if an email address is valid using regex. Returns true if matches email pattern.

```go
func IsValidEmail(email string) bool
```

## Functions - Version

### GetVersionInfo

GetVersionInfo returns the version string. Defaults to dev for development builds.

```go
func GetVersionInfo() string
```

### GetFullVersionInfo

GetFullVersionInfo returns formatted string with version, commit, and build date.

```go
func GetFullVersionInfo() string
```

### GetVersion

GetVersion returns pointer to full version info string. Always returns nil error.

```go
func GetVersion() (*string, error)
```

### GetCurrentVersion

GetCurrentVersion returns just the version string without metadata.

```go
func GetCurrentVersion() string
```

### IsDevBuild

IsDevBuild checks if running a development build. Returns true if version is dev or commit unknown.

```go
func IsDevBuild() bool
```
