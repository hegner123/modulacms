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
    ActionCopy        Action = "copy"
    ActionPublish     Action = "publish"
    ActionArchive     Action = "archive"
    ActionGoParent    Action = "go_parent"
    ActionGoChild     Action = "go_child"
)
```

Action constants represent semantic TUI operations independent of physical keys. Each action maps to one or more key strings via KeyMap. Examples: ActionQuit exits the application, ActionSelect confirms a choice, ActionReorderUp moves items up in lists.

## Color Variables

```go
var (
    White       = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "15"}
    LightGray   = lipgloss.CompleteColor{TrueColor: "#c0c0c0", ANSI256: "254", ANSI: "7"}
    Gray        = lipgloss.CompleteColor{TrueColor: "#808080", ANSI256: "250", ANSI: "8"}
    Black       = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
    Purple      = lipgloss.CompleteColor{TrueColor: "#6612e3", ANSI256: "129", ANSI: "5"}
    LightPurple = lipgloss.CompleteColor{TrueColor: "#8347de", ANSI256: "98", ANSI: "13"}
    Emerald     = lipgloss.CompleteColor{TrueColor: "#00CC66", ANSI256: "41", ANSI: "2"}
    Rose        = lipgloss.CompleteColor{TrueColor: "#D90368", ANSI256: "161", ANSI: "1"}
    Yellow      = lipgloss.CompleteColor{TrueColor: "#F1C40F", ANSI256: "220", ANSI: "11"}
    Orange      = lipgloss.CompleteColor{TrueColor: "#F75C03", ANSI256: "202", ANSI: "3"}
    Blue        = lipgloss.CompleteColor{TrueColor: "#5f5fff", ANSI256: "63", ANSI: "4"}
)
```

Color variables define lipgloss.CompleteColor values with TrueColor, ANSI256, and ANSI fallback codes. Used in DefaultStyle and available for custom styling. Each color supports terminal environments from basic 16-color ANSI to full truecolor RGB.

```go
var DefaultStyle Color = Color{
    Primary: lipgloss.CompleteAdaptiveColor{Light: Black, Dark: White},
    PrimaryBG: lipgloss.CompleteAdaptiveColor{Light: White, Dark: Black},
    Secondary: lipgloss.CompleteAdaptiveColor{Light: Gray, Dark: LightGray},
    SecondaryBG: lipgloss.CompleteAdaptiveColor{Light: White, Dark: Black},
    Tertiary: lipgloss.CompleteAdaptiveColor{Light: LightGray, Dark: Gray},
    TertiaryBG: lipgloss.CompleteAdaptiveColor{Light: Gray, Dark: Black},
    Accent: lipgloss.CompleteAdaptiveColor{Light: Purple, Dark: Purple},
    AccentBG: lipgloss.CompleteAdaptiveColor{Light: White, Dark: Blue},
    Accent2: lipgloss.CompleteAdaptiveColor{Light: Rose, Dark: Rose},
    Accent2BG: lipgloss.CompleteAdaptiveColor{Light: White, Dark: Black},
    Active: lipgloss.CompleteAdaptiveColor{Light: Black, Dark: Black},
    ActiveBG: lipgloss.CompleteAdaptiveColor{Light: Gray, Dark: LightGray},
    Status1: lipgloss.CompleteAdaptiveColor{Light: Black, Dark: White},
    Status1BG: lipgloss.CompleteAdaptiveColor{Light: LightGray, Dark: Black},
    Status2: lipgloss.CompleteAdaptiveColor{Light: Gray, Dark: Black},
    Status2BG: lipgloss.CompleteAdaptiveColor{Light: Black, Dark: Gray},
    Status3: lipgloss.CompleteAdaptiveColor{Light: LightPurple, Dark: LightPurple},
    Status3BG: lipgloss.CompleteAdaptiveColor{Light: Black, Dark: Black},
    PrimaryBorder: lipgloss.CompleteAdaptiveColor{Light: Purple, Dark: Purple},
    Warn: lipgloss.CompleteAdaptiveColor{Light: Orange, Dark: Orange},
    WarnBG: lipgloss.CompleteAdaptiveColor{Light: White, Dark: White},
}
```

DefaultStyle provides the built-in TUI color scheme with adaptive light and dark variants. Each semantic role (Primary, Secondary, Accent, Active, Status, Warn) has foreground and background colors. Adapts to terminal background automatically.

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
    Primary       lipgloss.CompleteAdaptiveColor
    PrimaryBG     lipgloss.CompleteAdaptiveColor
    Secondary     lipgloss.CompleteAdaptiveColor
    SecondaryBG   lipgloss.CompleteAdaptiveColor
    Tertiary      lipgloss.CompleteAdaptiveColor
    TertiaryBG    lipgloss.CompleteAdaptiveColor
    Accent        lipgloss.CompleteAdaptiveColor
    AccentBG      lipgloss.CompleteAdaptiveColor
    Accent2       lipgloss.CompleteAdaptiveColor
    Accent2BG     lipgloss.CompleteAdaptiveColor
    Active        lipgloss.CompleteAdaptiveColor
    ActiveBG      lipgloss.CompleteAdaptiveColor
    Status1       lipgloss.CompleteAdaptiveColor
    Status1BG     lipgloss.CompleteAdaptiveColor
    Status2       lipgloss.CompleteAdaptiveColor
    Status2BG     lipgloss.CompleteAdaptiveColor
    Status3       lipgloss.CompleteAdaptiveColor
    Status3BG     lipgloss.CompleteAdaptiveColor
    PrimaryBorder lipgloss.CompleteAdaptiveColor
    Warn          lipgloss.CompleteAdaptiveColor
    WarnBG        lipgloss.CompleteAdaptiveColor
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
    Db_Driver             DbDriver
    Db_URL                string
    Db_Name               string
    Db_User               string
    Db_Password           string
    Bucket_Region         string
    Bucket_Media          string
    Bucket_Backup         string
    Bucket_Endpoint       string
    Bucket_Access_Key     string
    Bucket_Secret_Key     string
    Bucket_Default_ACL    string
    Bucket_Force_Path_Style bool
    Backup_Option         string
    Backup_Paths          []string
    Oauth_Client_Id       string
    Oauth_Client_Secret   string
    Oauth_Scopes          []string
    Oauth_Endpoint        map[Endpoint]string
    Oauth_Provider_Name   string
    Oauth_Redirect_URL    string
    Oauth_Success_Redirect string
    Cors_Origins          []string
    Cors_Methods          []string
    Cors_Headers          []string
    Cors_Credentials      bool
    Custom_Style_Path     string
    Update_Auto_Enabled   bool
    Update_Check_Interval string
    Update_Channel        string
    Update_Notify_Only    bool
    Output_Format         OutputFormat
    Space_ID              string
    Node_ID               string
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
    Plugin_Enabled        bool
    Plugin_Directory      string
    Plugin_Max_VMs        int
    Plugin_Timeout        int
    Plugin_Max_Ops        int
    Plugin_DB_MaxOpenConns    int
    Plugin_DB_MaxIdleConns    int
    Plugin_DB_ConnMaxLifetime string
    KeyBindings           KeyMap
}
```

Config holds all application configuration including server ports, database credentials, OAuth settings, S3 bucket configuration, CORS policy, observability provider integration, plugin runtime limits, and keybinding overrides. Loaded from JSON via Provider implementations.

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

FileProvider loads configuration from a JSON file. The path field specifies the file location. If path is empty, defaults to config.json. Implements the Provider interface via the Get method.

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

NewFileProvider creates a file-based configuration provider. If path is empty, defaults to config.json in the current directory. Returns a FileProvider ready to load configuration via the Get method.

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
