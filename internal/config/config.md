# config

Package config manages application configuration for ModulaCMS. It provides a provider-based configuration system supporting JSON file loading, comprehensive TUI keybinding management, and extensive deployment settings including database drivers, OAuth providers, S3-compatible storage, CORS policies, observability integration, and plugin runtime configuration.

## Overview

The config package implements a provider abstraction for loading configuration from various sources. Currently supports JSON file loading via FileProvider. Configuration includes database connection parameters for SQLite, MySQL, and PostgreSQL drivers, OAuth endpoints, S3 bucket credentials, cookie settings, backup paths, CORS policy, update behavior, observability provider integration, and plugin runtime limits.

All configuration is managed through a Manager that handles thread-safe loading and access. Keybindings use a semantic action-based system allowing runtime customization while preserving default bindings for unspecified actions.

## Constants

### Endpoint Constants

```go
const (
    OauthAuthURL     Endpoint = "oauth_auth_url"
    OauthTokenURL    Endpoint = "oauth_token_url"
    OauthUserInfoURL Endpoint = "oauth_userinfo_url"
)
```

Endpoint constants define OAuth flow URL identifiers used in the Oauth_Endpoint map. OauthAuthURL identifies the authorization endpoint, OauthTokenURL the token exchange endpoint, and OauthUserInfoURL the user information retrieval endpoint.

### DbDriver Constants

```go
const (
    Sqlite DbDriver = "sqlite"
    Mysql  DbDriver = "mysql"
    Psql   DbDriver = "postgres"
)
```

DbDriver constants specify supported database backend drivers. Sqlite uses embedded SQLite, Mysql connects to MySQL servers, and Psql connects to PostgreSQL servers. The Db_Driver field in Config determines which driver the application uses at runtime.

### EmailProvider Constants

```go
const (
    EmailDisabled EmailProvider = ""         // Disabled (default, zero value)
    EmailSmtp     EmailProvider = "smtp"     // Standard SMTP relay
    EmailSendGrid EmailProvider = "sendgrid" // SendGrid HTTP API
    EmailSES      EmailProvider = "ses"      // AWS SES HTTP API
    EmailPostmark EmailProvider = "postmark" // Postmark HTTP API
)
```

EmailProvider constants specify supported email sending backends. EmailDisabled (empty string) is the default zero value. EmailSmtp uses standard SMTP relay. EmailSendGrid, EmailSES, and EmailPostmark use their respective HTTP APIs.

### OutputFormat Constants

```go
const (
    FormatContentful OutputFormat = "contentful"
    FormatSanity     OutputFormat = "sanity"
    FormatStrapi     OutputFormat = "strapi"
    FormatWordPress  OutputFormat = "wordpress"
    FormatClean      OutputFormat = "clean"
    FormatRaw        OutputFormat = "raw"
    FormatDefault    OutputFormat = ""
)
```

OutputFormat constants define content API response formatting options. FormatContentful, FormatSanity, FormatStrapi, and FormatWordPress match their respective CMS API structures. FormatClean removes metadata, FormatRaw returns unprocessed data. FormatDefault (empty string) defaults to raw output.

### Action Constants

```go
const (
    ActionQuit        Action = "quit"
    ActionDismiss     Action = "dismiss"
    ActionUp          Action = "up"
    ActionDown        Action = "down"
    ActionBack        Action = "back"
    ActionSelect      Action = "select"
    ActionNextPanel   Action = "next_panel"
    ActionPrevPanel   Action = "prev_panel"
    ActionNew         Action = "new"
    ActionEdit        Action = "edit"
    ActionDelete      Action = "delete"
    ActionMove        Action = "move"
    ActionTitlePrev   Action = "title_prev"
    ActionTitleNext   Action = "title_next"
    ActionPagePrev    Action = "page_prev"
    ActionPageNext    Action = "page_next"
    ActionExpand      Action = "expand"
    ActionCollapse    Action = "collapse"
    ActionReorderUp   Action = "reorder_up"
    ActionReorderDown Action = "reorder_down"
    ActionCopy         Action = "copy"
    ActionPublish      Action = "publish"
    ActionGoParent     Action = "go_parent"
    ActionGoChild      Action = "go_child"
    ActionVersions     Action = "versions"
    ActionLocale       Action = "locale"
    ActionScreenNext   Action = "screen_next"
    ActionScreenToggle Action = "screen_toggle"
    ActionScreenReset  Action = "screen_reset"
    ActionAccordion    Action = "accordion"
    ActionTabPrev      Action = "tab_prev"
    ActionTabNext      Action = "tab_next"
    ActionAdminToggle  Action = "admin_toggle"
    ActionSearch       Action = "search"
)
```

Action constants represent semantic TUI operations independent of physical keys. Each action maps to one or more key strings via KeyMap. Examples: ActionQuit exits the application, ActionSelect confirms a choice, ActionReorderUp moves items up in lists.

## Color Variables

```go
var (
    White       = lipgloss.Color("#FFFFFF")
    LightGray   = lipgloss.Color("#c0c0c0")
    Gray        = lipgloss.Color("#808080")
    Black       = lipgloss.Color("#000000")
    Purple      = lipgloss.Color("#6612e3")
    LightPurple = lipgloss.Color("#8347de")
    Emerald     = lipgloss.Color("#00CC66")
    Rose        = lipgloss.Color("#D90368")
    Yellow      = lipgloss.Color("#F1C40F")
    Orange      = lipgloss.Color("#F75C03")
    Blue        = lipgloss.Color("#5f5fff")
    Amber       = lipgloss.Color("#FFBF00")
)
```

Color variables define lipgloss.Color values (truecolor hex). Bubble Tea v2 handles color downsampling automatically. Used in DefaultStyle and available for custom styling.

```go
var DefaultStyle Color = Color{
    Primary:       compat.AdaptiveColor{Light: Black, Dark: White},
    PrimaryBG:     compat.AdaptiveColor{Light: White, Dark: Black},
    Secondary:     compat.AdaptiveColor{Light: Gray, Dark: LightGray},
    SecondaryBG:   compat.AdaptiveColor{Light: White, Dark: Black},
    Tertiary:      compat.AdaptiveColor{Light: LightGray, Dark: Gray},
    TertiaryBG:    compat.AdaptiveColor{Light: Gray, Dark: Black},
    Accent:        compat.AdaptiveColor{Light: Blue, Dark: Blue},
    AccentBG:      compat.AdaptiveColor{Light: White, Dark: Blue},
    Accent2:       compat.AdaptiveColor{Light: Rose, Dark: Rose},
    Accent2BG:     compat.AdaptiveColor{Light: White, Dark: Black},
    Active:        compat.AdaptiveColor{Light: Black, Dark: Black},
    ActiveBG:      compat.AdaptiveColor{Light: Gray, Dark: LightGray},
    Status1:       compat.AdaptiveColor{Light: Black, Dark: White},
    Status1BG:     compat.AdaptiveColor{Light: LightGray, Dark: Black},
    Status2:       compat.AdaptiveColor{Light: Gray, Dark: Black},
    Status2BG:     compat.AdaptiveColor{Light: Black, Dark: Gray},
    Status3:       compat.AdaptiveColor{Light: LightPurple, Dark: LightPurple},
    Status3BG:     compat.AdaptiveColor{Light: Black, Dark: Black},
    PrimaryBorder: compat.AdaptiveColor{Light: Purple, Dark: Purple},
    Warn:          compat.AdaptiveColor{Light: Orange, Dark: Orange},
    WarnBG:        compat.AdaptiveColor{Light: White, Dark: White},
}
```

DefaultStyle provides the built-in TUI color scheme with adaptive light and dark variants using `compat.AdaptiveColor` from `charm.land/lipgloss/v2/compat`. Each semantic role (Primary, Secondary, Accent, Active, Status, Warn) has foreground and background colors. Adapts to terminal background automatically.

## Types

### Endpoint

```go
type Endpoint string
```

Endpoint represents OAuth endpoint URL identifiers. Used as map keys in Config.Oauth_Endpoint to store authorization, token exchange, and user info URLs. Values are the Oauth* constants defined above.

### DbDriver

```go
type DbDriver string
```

DbDriver identifies the database backend driver. Valid values are Sqlite, Mysql, and Psql constants. Determines which database abstraction implementation is loaded at startup. Affects connection string format and SQL dialect.

### EmailProvider

```go
type EmailProvider string
```

EmailProvider identifies the email sending backend. Valid values are the Email* constants. Empty string means email is disabled.

### OutputFormat

```go
type OutputFormat string
```

OutputFormat specifies content API response formatting style. Controls how content is serialized when returned via HTTP endpoints. Valid values are the Format* constants. Empty string defaults to FormatRaw.

### Action

```go
type Action string
```

Action represents a semantic TUI operation independent of physical key bindings. Actions map to one or more key strings via KeyMap. Allows keybinding customization without changing control logic. Values are the Action* constants.

### Color

```go
type Color struct {
    Primary       compat.AdaptiveColor
    PrimaryBG     compat.AdaptiveColor
    Secondary     compat.AdaptiveColor
    SecondaryBG   compat.AdaptiveColor
    Tertiary      compat.AdaptiveColor
    TertiaryBG    compat.AdaptiveColor
    Accent        compat.AdaptiveColor
    AccentBG      compat.AdaptiveColor
    Accent2       compat.AdaptiveColor
    Accent2BG     compat.AdaptiveColor
    Active        compat.AdaptiveColor
    ActiveBG      compat.AdaptiveColor
    Status1       compat.AdaptiveColor
    Status1BG     compat.AdaptiveColor
    Status2       compat.AdaptiveColor
    Status2BG     compat.AdaptiveColor
    Status3       compat.AdaptiveColor
    Status3BG     compat.AdaptiveColor
    PrimaryBorder compat.AdaptiveColor
    Warn          compat.AdaptiveColor
    WarnBG        compat.AdaptiveColor
}
```

Color defines the TUI color palette with semantic roles. Each role has foreground and background adaptive colors that switch based on terminal background detection. Used for consistent styling across all TUI screens and components.

### KeyMap

```go
type KeyMap map[Action][]string
```

KeyMap maps semantic actions to physical key strings as reported by bubbletea's KeyMsg.String(). Each action can have multiple key bindings. Supports JSON marshaling for configuration persistence. Implements custom UnmarshalJSON for loading overrides.

### Config

```go
type Config struct {
    // Server
    Environment           string
    OS                    string
    Environment_Hosts     map[string]string
    Port                  string
    SSL_Port              string
    Cert_Dir              string
    Client_Site           string
    Admin_Site            string
    SSH_Host              string
    SSH_Port              string
    Options               map[string][]any
    Log_Path              string
    Auth_Salt             string
    Cookie_Name           string
    Cookie_Duration       string
    Cookie_Secure         bool
    Cookie_SameSite       string

    // Database
    Db_Driver             DbDriver
    Db_URL                string
    Db_Name               string
    Db_User               string
    Db_Password           string

    // Remote connection (mutually exclusive with Db_Driver for connect command)
    Remote_URL            string
    Remote_API_Key        string

    // S3 storage (public media)
    Bucket_Region           string
    Bucket_Media            string
    Bucket_Backup           string
    Bucket_Endpoint         string
    Bucket_Access_Key       string
    Bucket_Secret_Key       string
    Bucket_Public_URL       string
    Bucket_Default_ACL      string
    Bucket_Force_Path_Style bool

    // S3 storage (admin media — falls back to shared bucket if unset)
    Bucket_Admin_Media      string
    Bucket_Admin_Endpoint   string
    Bucket_Admin_Access_Key string
    Bucket_Admin_Secret_Key string
    Bucket_Admin_Public_URL string

    // Upload limits
    Max_Upload_Size       int64  // bytes, default 10MB (10485760)

    // Backup
    Backup_Option         string
    Backup_Paths          []string

    // OAuth
    Oauth_Client_Id        string
    Oauth_Client_Secret    string
    Oauth_Scopes           []string
    Oauth_Endpoint         map[Endpoint]string
    Oauth_Provider_Name    string
    Oauth_Redirect_URL     string
    Oauth_Success_Redirect string

    // CORS
    Cors_Origins          []string
    Cors_Methods          []string
    Cors_Headers          []string
    Cors_Credentials      bool

    // TUI
    Custom_Style_Path     string
    KeyBindings           KeyMap

    // Update
    Update_Auto_Enabled   bool
    Update_Check_Interval string
    Update_Channel        string
    Update_Notify_Only    bool

    // Output
    Output_Format         OutputFormat
    Space_ID              string
    Node_ID               string

    // Observability
    Observability_Enabled        bool
    Observability_Provider       string
    Observability_DSN            string
    Observability_Environment    string
    Observability_Release        string
    Observability_Sample_Rate    float64
    Observability_Traces_Rate    float64
    Observability_Send_PII       bool
    Observability_Debug          bool
    Observability_Server_Name    string
    Observability_Flush_Interval string
    Observability_Tags           map[string]string

    // Email
    Email_Enabled               bool
    Email_Provider              EmailProvider
    Email_From_Address          string
    Email_From_Name             string
    Email_Host                  string
    Email_Port                  int
    Email_Username              string
    Email_Password              string
    Email_TLS                   bool
    Email_API_Key               string
    Email_API_Endpoint          string
    Email_Reply_To              string
    Email_AWS_Access_Key_ID     string
    Email_AWS_Secret_Access_Key string
    Password_Reset_URL          string

    // Plugin runtime
    Plugin_Enabled        bool
    Plugin_Directory      string
    Plugin_Max_VMs        int
    Plugin_Timeout        int
    Plugin_Max_Ops        int

    // Plugin database pool limits
    Plugin_DB_MaxOpenConns    int
    Plugin_DB_MaxIdleConns    int
    Plugin_DB_ConnMaxLifetime string

    // Plugin HTTP integration
    Plugin_Max_Request_Body  int64
    Plugin_Max_Response_Body int64
    Plugin_Rate_Limit        int
    Plugin_Max_Routes        int
    Plugin_Trusted_Proxies   []string

    // Plugin content hooks (Phase 3)
    Plugin_Hook_Reserve_VMs            int
    Plugin_Hook_Max_Consecutive_Aborts int
    Plugin_Hook_Max_Ops                int
    Plugin_Hook_Max_Concurrent_After   int
    Plugin_Hook_Timeout_Ms             int
    Plugin_Hook_Event_Timeout_Ms       int

    // Plugin outbound requests
    Plugin_Request_Timeout      int
    Plugin_Request_Max_Response int64
    Plugin_Request_Max_Body     int64
    Plugin_Request_Rate_Limit   int
    Plugin_Request_Global_Rate  int
    Plugin_Request_CB_Failures  int
    Plugin_Request_CB_Reset     int
    Plugin_Request_Allow_Local  bool

    // Plugin production hardening (Phase 4)
    Plugin_Hot_Reload     bool
    Plugin_Max_Failures   int
    Plugin_Reset_Interval string
    Plugin_Sync_Interval  string

    // Deploy sync
    Deploy_Environments []DeployEnvironmentConfig
    Deploy_Snapshot_Dir string

    // Tree composition
    Composition_Max_Depth int

    // Publishing
    Publish_Schedule_Interval int
    Version_Max_Per_Content   int
    Node_Level_Publish        bool

    // Richtext editor
    Richtext_Toolbar []string

    // Internationalization
    I18n_Enabled        bool
    I18n_Default_Locale string

    // Webhooks
    Webhook_Enabled                 bool
    Webhook_Timeout                 int
    Webhook_Max_Retries             int
    Webhook_Workers                 int
    Webhook_Allow_HTTP              bool
    Webhook_Delivery_Retention_Days int

    // MCP server
    MCP_Enabled bool
    MCP_API_Key string

    // Search
    Search_Enabled bool
    Search_Path    string
}
```

Config holds all application configuration including server ports, database credentials, remote connection, OAuth settings, S3 bucket configuration for both public and admin media, CORS policy, email provider settings, observability integration, plugin runtime limits (HTTP, hooks, requests, hardening), deploy sync, publishing, i18n, webhooks, MCP server, search, and keybinding overrides. Loaded from JSON via Provider implementations.

### DeployEnvironmentConfig

```go
type DeployEnvironmentConfig struct {
    Name   string
    URL    string
    APIKey string
}
```

DeployEnvironmentConfig describes a remote Modula instance for deploy operations. APIKey supports `${VAR}` expansion via the existing config system.

### Provider

```go
type Provider interface {
    Get() (*Config, error)
}
```

Provider defines the interface for retrieving configuration. Implementations load Config from various sources. Get returns a populated Config or an error if loading fails. Currently implemented by FileProvider.

### FileProvider

```go
type FileProvider struct {
    path string
}
```

FileProvider loads configuration from a JSON file. The path field specifies the file location. If path is empty, defaults to modula.config.json. Implements the Provider interface via the Get method.

### Manager

```go
type Manager struct {
    provider  Provider
    config    *Config
    loaded    bool
    loadMutex sync.Mutex
}
```

Manager handles configuration loading and access with thread-safe lazy initialization. Uses a Provider to load Config on first access. The loadMutex ensures concurrent calls to Load or Config block until loading completes. The loaded flag tracks initialization state.

## Functions

#### DefaultKeyMap

```go
func DefaultKeyMap() KeyMap
```

DefaultKeyMap returns the built-in keybindings matching the original hardcoded TUI control handlers. Maps each Action constant to one or more key strings. Example mappings: ActionQuit to "q" and "ctrl+c", ActionUp to "up" and "k", ActionSelect to "enter", "l", and "right".

#### DefaultConfig

```go
func DefaultConfig() Config
```

DefaultConfig returns a Config populated with sensible development defaults. Generates unique Auth_Salt from current Unix timestamp. Sets Environment to development, Db_Driver to sqlite, ports to 8080 HTTP and 4000 SSL, localhost hosts, empty OAuth credentials, permissive CORS for localhost:3000, disabled auto-updates, console observability provider, and default keybindings. Returns a Config ready for use or JSON serialization.

#### IsValidOutputFormat

```go
func IsValidOutputFormat(format string) bool
```

IsValidOutputFormat checks if the given format string matches a valid OutputFormat constant. Returns true for contentful, sanity, strapi, wordpress, clean, raw, or empty string. Returns false otherwise. Use before setting Config.Output_Format from user input.

#### GetValidOutputFormats

```go
func GetValidOutputFormats() []string
```

GetValidOutputFormats returns a slice of all valid output format strings: contentful, sanity, strapi, wordpress, clean, raw. Useful for validation messages or displaying available options to users. Does not include the empty default format.

#### NewFileProvider

```go
func NewFileProvider(path string) *FileProvider
```

NewFileProvider creates a file-based configuration provider. If path is empty, defaults to modula.config.json in the current directory. Returns a FileProvider ready to load configuration via the Get method.

#### NewManager

```go
func NewManager(provider Provider) *Manager
```

NewManager creates a configuration manager with the specified provider. Returns a Manager in unloaded state. Call Load to load configuration immediately, or call Config which loads on first access. Thread-safe for concurrent Load or Config calls.

## Methods

#### Config.BucketEndpointURL

```go
func (c Config) BucketEndpointURL() string
```

BucketEndpointURL returns Bucket_Endpoint prefixed with the scheme determined by Environment. Non-TLS environments (http-only and docker) get http scheme. All other environments get https scheme. Returns empty string if Bucket_Endpoint is empty.

#### Config.BucketPublicURL

```go
func (c Config) BucketPublicURL() string
```

BucketPublicURL returns the public-facing base URL for constructing browser-accessible media URLs. Falls back to BucketEndpointURL if Bucket_Public_URL is not configured. In Docker environments, Bucket_Endpoint is typically a container hostname that browsers cannot resolve.

#### Config.AdminBucketMedia

```go
func (c Config) AdminBucketMedia() string
```

AdminBucketMedia returns the admin media bucket name. Falls back to Bucket_Media if Bucket_Admin_Media is not configured.

#### Config.AdminBucketEndpoint

```go
func (c Config) AdminBucketEndpoint() string
```

AdminBucketEndpoint returns the admin media endpoint. Falls back to Bucket_Endpoint if Bucket_Admin_Endpoint is not configured.

#### Config.AdminBucketEndpointURL

```go
func (c Config) AdminBucketEndpointURL() string
```

AdminBucketEndpointURL returns the full admin bucket endpoint URL with scheme. Uses AdminBucketEndpoint and applies the same scheme logic as BucketEndpointURL.

#### Config.AdminBucketAccessKey

```go
func (c Config) AdminBucketAccessKey() string
```

AdminBucketAccessKey returns the admin media access key. Falls back to Bucket_Access_Key if Bucket_Admin_Access_Key is not configured.

#### Config.AdminBucketSecretKey

```go
func (c Config) AdminBucketSecretKey() string
```

AdminBucketSecretKey returns the admin media secret key. Falls back to Bucket_Secret_Key if Bucket_Admin_Secret_Key is not configured.

#### Config.AdminBucketPublicURL

```go
func (c Config) AdminBucketPublicURL() string
```

AdminBucketPublicURL returns the admin media public URL. Falls back to BucketPublicURL if Bucket_Admin_Public_URL is not configured.

#### Config.CompositionMaxDepth

```go
func (c Config) CompositionMaxDepth() int
```

CompositionMaxDepth returns the configured maximum composition depth. Falls back to 10 if no positive value is configured.

#### Config.PublishScheduleInterval

```go
func (c Config) PublishScheduleInterval() int
```

PublishScheduleInterval returns the configured interval in seconds between scheduler ticks for scheduled publishing. Falls back to 60 if not configured.

#### Config.VersionMaxPerContent

```go
func (c Config) VersionMaxPerContent() int
```

VersionMaxPerContent returns the maximum number of versions to retain per content item. Falls back to 50 if not configured. 0 means unlimited.

#### Config.RichtextToolbar

```go
func (c Config) RichtextToolbar() []string
```

RichtextToolbar returns the configured default toolbar buttons for richtext fields. Falls back to a sensible default set (bold, italic, h1, h2, h3, link, ul, ol, preview) if not configured.

#### Config.WebhookEnabled

```go
func (c Config) WebhookEnabled() bool
```

WebhookEnabled returns whether webhooks are active.

#### Config.WebhookTimeout

```go
func (c Config) WebhookTimeout() int
```

WebhookTimeout returns the HTTP timeout in seconds for webhook delivery. Falls back to 10 if not configured.

#### Config.WebhookMaxRetries

```go
func (c Config) WebhookMaxRetries() int
```

WebhookMaxRetries returns the maximum number of delivery retry attempts. Falls back to 3 if not configured.

#### Config.WebhookWorkers

```go
func (c Config) WebhookWorkers() int
```

WebhookWorkers returns the number of concurrent delivery workers. Falls back to 4 if not configured.

#### Config.WebhookAllowHTTP

```go
func (c Config) WebhookAllowHTTP() bool
```

WebhookAllowHTTP returns whether non-TLS webhook URLs are allowed (dev only).

#### Config.WebhookDeliveryRetentionDays

```go
func (c Config) WebhookDeliveryRetentionDays() int
```

WebhookDeliveryRetentionDays returns the number of days to retain completed deliveries. Falls back to 30 if not configured. 0 means unlimited retention.

#### Config.SearchEnabled

```go
func (c Config) SearchEnabled() bool
```

SearchEnabled returns whether search is enabled.

#### Config.I18nEnabled

```go
func (c Config) I18nEnabled() bool
```

I18nEnabled returns whether internationalization is active.

#### Config.I18nDefaultLocale

```go
func (c Config) I18nDefaultLocale() string
```

I18nDefaultLocale returns the configured default locale code. Falls back to "en" if not configured.

#### Config.MaxUploadSize

```go
func (c Config) MaxUploadSize() int64
```

MaxUploadSize returns the configured maximum upload size in bytes. Falls back to 10 MB if no positive value is configured.

#### Config.JSON

```go
func (c Config) JSON() []byte
```

JSON marshals the Config to JSON bytes. Returns empty byte slice if marshaling fails. Useful for serializing configuration to files or HTTP responses. Does not return error on failure.

#### Config.Stringify

```go
func (c Config) Stringify() string
```

Stringify formats the Config as a vertically joined lipgloss string. Currently returns empty string as implementation is incomplete. Intended for TUI display of configuration values.

#### FileProvider.Get

```go
func (fp *FileProvider) Get() (*Config, error)
```

Get implements the Provider interface. Opens the file at fp.path, reads all bytes, unmarshals JSON into a Config struct. Returns error if file cannot be opened, read, or parsed. Closes file even on error. Wraps errors with context.

#### KeyMap.Matches

```go
func (km KeyMap) Matches(key string, action Action) bool
```

Matches returns true when key is bound to the given action. Searches all keys mapped to action. Use in control handlers to test if a pressed key should trigger an action. Returns false if action has no bindings.

#### KeyMap.Merge

```go
func (km KeyMap) Merge(overrides KeyMap)
```

Merge replaces bindings in km with those from overrides. Actions not present in overrides keep their current bindings. Mutates km in place. Used by Manager.Load to apply user overrides while preserving default bindings for unspecified actions.

#### KeyMap.HintString

```go
func (km KeyMap) HintString(action Action) string
```

HintString returns the first key bound to action, suitable for display in status bars or help text. Returns "?" if the action has no bindings. Does not format or prettify the key string.

#### KeyMap.MarshalJSON

```go
func (km KeyMap) MarshalJSON() ([]byte, error)
```

MarshalJSON implements json.Marshaler. Converts Action-keyed map to string-keyed map for JSON serialization. All bindings are written. Use when saving KeyMap to configuration files.

#### KeyMap.UnmarshalJSON

```go
func (km *KeyMap) UnmarshalJSON(data []byte) error
```

UnmarshalJSON implements json.Unmarshaler. Reads string-keyed maps from JSON into Action-keyed KeyMap. Overwrites km contents. Use when loading KeyMap overrides from configuration files.

#### Manager.Load

```go
func (m *Manager) Load() error
```

Load loads configuration from the provider immediately. Acquires loadMutex, calls loadLocked, sets loaded flag on success. Returns error if provider.Get fails. Normalizes Bucket_Endpoint by stripping scheme prefixes. Merges KeyBindings with defaults so unspecified actions keep default bindings. Safe to call concurrently.

#### Manager.Config

```go
func (m *Manager) Config() (*Config, error)
```

Config returns the loaded configuration. If not already loaded, loads it first by calling loadLocked. Acquires loadMutex for thread-safe lazy initialization. Returns error if loading fails. Subsequent calls return cached config without re-loading.
