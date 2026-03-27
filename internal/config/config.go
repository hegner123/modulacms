// Package config provides configuration management for Modula, including
// database drivers, OAuth endpoints, S3-compatible storage buckets, CORS
// settings, SSL/TLS configuration, plugin runtime options, and observability.
package config

import "strings"

// Endpoint identifies OAuth provider endpoint types.
type Endpoint string

// DbDriver specifies which database backend to use.
type DbDriver string

// OutputFormat defines the API response structure for content endpoints.
type OutputFormat string

// EmailProvider specifies which email sending backend to use.
type EmailProvider string

// Environment identifies the runtime stage and deployment mode.
// Format is "{stage}" or "{stage}-docker" where stage is one of:
// local, development, staging, production.
type Environment string

// OAuth endpoint keys used in the Oauth_Endpoint configuration map.
const (
	OauthAuthURL     Endpoint = "oauth_auth_url"
	OauthTokenURL    Endpoint = "oauth_token_url"
	OauthUserInfoURL Endpoint = "oauth_userinfo_url"
)

// Supported database drivers for Modula.
const (
	Sqlite DbDriver = "sqlite"
	Mysql  DbDriver = "mysql"
	Psql   DbDriver = "postgres"
)

// Supported email providers for transactional email.
const (
	EmailDisabled EmailProvider = ""         // Disabled (default, zero value)
	EmailSmtp     EmailProvider = "smtp"     // Standard SMTP relay
	EmailSendGrid EmailProvider = "sendgrid" // SendGrid HTTP API
	EmailSES      EmailProvider = "ses"      // AWS SES HTTP API
	EmailPostmark EmailProvider = "postmark" // Postmark HTTP API
)

// Output formats for content API responses mimicking popular CMS structures.
const (
	FormatContentful OutputFormat = "contentful"
	FormatSanity     OutputFormat = "sanity"
	FormatStrapi     OutputFormat = "strapi"
	FormatWordPress  OutputFormat = "wordpress"
	FormatClean      OutputFormat = "clean"
	FormatRaw        OutputFormat = "raw"
	FormatDefault    OutputFormat = "" // Empty string defaults to raw
)

// Environment constants. Each stage may run natively or in Docker.
// Docker variants bind to 0.0.0.0; native variants use localhost or client_site.
const (
	EnvLocal             Environment = "local"
	EnvLocalDocker       Environment = "local-docker"
	EnvDevelopment       Environment = "development"
	EnvDevelopmentDocker Environment = "development-docker"
	EnvStaging           Environment = "staging"
	EnvStagingDocker     Environment = "staging-docker"
	EnvProduction        Environment = "production"
	EnvProductionDocker  Environment = "production-docker"
)

// Stage returns the base stage name without the "-docker" suffix.
func (e Environment) Stage() string {
	return strings.TrimSuffix(string(e), "-docker")
}

// IsDocker returns true if the environment is a Docker variant.
func (e Environment) IsDocker() bool {
	return strings.HasSuffix(string(e), "-docker")
}

// IsLocal returns true for local and local-docker environments.
func (e Environment) IsLocal() bool {
	return e.Stage() == "local"
}

// UsesTLS returns true if the environment should start an HTTPS server.
// Local environments run HTTP only.
func (e Environment) UsesTLS() bool {
	return !e.IsLocal()
}

// UsesAutocert returns true if the environment should use Let's Encrypt
// for automatic TLS certificates. Development uses self-signed certs.
func (e Environment) UsesAutocert() bool {
	stage := e.Stage()
	return stage == "staging" || stage == "production"
}

// HTTPHost returns the HTTP server bind address.
// Docker variants bind 0.0.0.0, local binds localhost,
// all others return empty string (caller should use client_site).
func (e Environment) HTTPHost() string {
	if e.IsDocker() {
		return "0.0.0.0"
	}
	if e.IsLocal() {
		return "localhost"
	}
	return ""
}

// UsesHTTPScheme returns true if bucket/internal URLs should use http://
// instead of https://. True for local environments.
func (e Environment) UsesHTTPScheme() bool {
	return e.IsLocal()
}

// IsValid returns true if the environment is a recognized value.
func (e Environment) IsValid() bool {
	switch e {
	case EnvLocal, EnvLocalDocker, EnvDevelopment, EnvDevelopmentDocker,
		EnvStaging, EnvStagingDocker, EnvProduction, EnvProductionDocker:
		return true
	default:
		return false
	}
}

// Label returns a human-readable display name for the environment.
func (e Environment) Label() string {
	switch e {
	case EnvLocal:
		return "Local"
	case EnvLocalDocker:
		return "Local (Docker)"
	case EnvDevelopment:
		return "Development"
	case EnvDevelopmentDocker:
		return "Development (Docker)"
	case EnvStaging:
		return "Staging"
	case EnvStagingDocker:
		return "Staging (Docker)"
	case EnvProduction:
		return "Production"
	case EnvProductionDocker:
		return "Production (Docker)"
	default:
		if e == "" {
			return "Default"
		}
		return string(e)
	}
}

// Config holds all runtime configuration for Modula including server settings,
// database credentials, OAuth providers, S3-compatible storage, CORS policies,
// plugin runtime limits, and observability integration.
type Config struct {
	Environment       Environment       `json:"environment"`
	OS                string            `json:"os"`
	Environment_Hosts map[string]string `json:"environment_hosts"`
	Port              string            `json:"port"`
	SSL_Port          string            `json:"ssl_port"`
	Cert_Dir          string            `json:"cert_dir"`
	Client_Site       string            `json:"client_site"`
	Admin_Site        string            `json:"admin_site"`
	SSH_Host          string            `json:"ssh_host"`
	SSH_Port          string            `json:"ssh_port"`
	Options           map[string][]any  `json:"options"`
	Log_Path          string            `json:"log_path"`
	Auth_Salt         string            `json:"auth_salt"`
	Cookie_Name       string            `json:"cookie_name"`
	Cookie_Duration   string            `json:"cookie_duration"`
	Cookie_Secure     bool              `json:"cookie_secure"`
	Cookie_SameSite   string            `json:"cookie_samesite"`
	Db_Driver         DbDriver          `json:"db_driver"`
	Db_URL            string            `json:"db_url"`
	Db_Name           string            `json:"db_name"`
	Db_User           string            `json:"db_username"`
	Db_Password       string            `json:"db_password"`

	// Remote connection (mutually exclusive with Db_Driver for connect command)
	Remote_URL              string              `json:"remote_url"`
	Remote_API_Key          string              `json:"remote_api_key"`
	Bucket_Region           string              `json:"bucket_region"`
	Bucket_Media            string              `json:"bucket_media"`
	Bucket_Backup           string              `json:"bucket_backup"`
	Bucket_Endpoint         string              `json:"bucket_endpoint"`
	Bucket_Access_Key       string              `json:"bucket_access_key"`
	Bucket_Secret_Key       string              `json:"bucket_secret_key"`
	Bucket_Public_URL       string              `json:"bucket_public_url"`
	Bucket_Default_ACL      string              `json:"bucket_default_acl"`
	Bucket_Force_Path_Style bool                `json:"bucket_force_path_style"`
	Bucket_Admin_Media      string              `json:"bucket_admin_media"`
	Bucket_Admin_Endpoint   string              `json:"bucket_admin_endpoint"`
	Bucket_Admin_Access_Key string              `json:"bucket_admin_access_key"`
	Bucket_Admin_Secret_Key string              `json:"bucket_admin_secret_key"`
	Bucket_Admin_Public_URL string              `json:"bucket_admin_public_url"`
	Max_Upload_Size         int64               `json:"max_upload_size"` // bytes, default 10MB (10485760)
	Backup_Option           string              `json:"backup_option"`
	Backup_Paths            []string            `json:"backup_paths"`
	Oauth_Client_Id         string              `json:"oauth_client_id"`
	Oauth_Client_Secret     string              `json:"oauth_client_secret"`
	Oauth_Scopes            []string            `json:"oauth_scopes"`
	Oauth_Endpoint          map[Endpoint]string `json:"oauth_endpoint"`
	Oauth_Provider_Name     string              `json:"oauth_provider_name"`
	Oauth_Redirect_URL      string              `json:"oauth_redirect_url"`
	Oauth_Success_Redirect  string              `json:"oauth_success_redirect"`
	Cors_Origins            []string            `json:"cors_origins"`
	Cors_Methods            []string            `json:"cors_methods"`
	Cors_Headers            []string            `json:"cors_headers"`
	Cors_Credentials        bool                `json:"cors_credentials"`
	Custom_Style_Path       string              `json:"custom_style_path"`
	Update_Auto_Enabled     bool                `json:"update_auto_enabled"`
	Update_Check_Interval   string              `json:"update_check_interval"`
	Update_Channel          string              `json:"update_channel"`
	Update_Notify_Only      bool                `json:"update_notify_only"`
	Output_Format           OutputFormat        `json:"output_format"`
	Space_ID                string              `json:"space_id"`
	Node_ID                 string              `json:"node_id"`

	// Observability - Metrics and Error Tracking
	Observability_Enabled        bool              `json:"observability_enabled"`
	Observability_Provider       string            `json:"observability_provider"`       // "sentry", "datadog", "newrelic", etc.
	Observability_DSN            string            `json:"observability_dsn"`            // Sentry DSN or equivalent connection string
	Observability_Environment    string            `json:"observability_environment"`    // "production", "staging", "development"
	Observability_Release        string            `json:"observability_release"`        // Version/release identifier
	Observability_Sample_Rate    float64           `json:"observability_sample_rate"`    // 0.0 to 1.0 - percentage of events to send
	Observability_Traces_Rate    float64           `json:"observability_traces_rate"`    // 0.0 to 1.0 - percentage of traces to send
	Observability_Send_PII       bool              `json:"observability_send_pii"`       // Whether to send personally identifiable info
	Observability_Debug          bool              `json:"observability_debug"`          // Enable debug logging for observability client
	Observability_Server_Name    string            `json:"observability_server_name"`    // Server/instance name
	Observability_Flush_Interval string            `json:"observability_flush_interval"` // How often to flush metrics (e.g., "30s", "1m")
	Observability_Tags           map[string]string `json:"observability_tags"`           // Global tags for all metrics/events

	// Email provider configuration
	Email_Enabled               bool          `json:"email_enabled"`
	Email_Provider              EmailProvider `json:"email_provider"`
	Email_From_Address          string        `json:"email_from_address"`
	Email_From_Name             string        `json:"email_from_name"`
	Email_Host                  string        `json:"email_host"`
	Email_Port                  int           `json:"email_port"`
	Email_Username              string        `json:"email_username"`
	Email_Password              string        `json:"email_password"`
	Email_TLS                   bool          `json:"email_tls"`
	Email_API_Key               string        `json:"email_api_key"`
	Email_API_Endpoint          string        `json:"email_api_endpoint"`
	Email_Reply_To              string        `json:"email_reply_to"`
	Email_AWS_Access_Key_ID     string        `json:"email_aws_access_key_id"`
	Email_AWS_Secret_Access_Key string        `json:"email_aws_secret_access_key"`
	Password_Reset_URL          string        `json:"password_reset_url"`

	// Plugin runtime configuration
	Plugin_Enabled   bool   `json:"plugin_enabled"`
	Plugin_Directory string `json:"plugin_directory"` // path to plugins dir, e.g. "./plugins/"
	Plugin_Max_VMs   int    `json:"plugin_max_vms"`   // per plugin, default 4
	Plugin_Timeout   int    `json:"plugin_timeout"`   // seconds, default 5
	Plugin_Max_Ops   int    `json:"plugin_max_ops"`   // per VM checkout, default 1000

	// Plugin database pool limits (zero values use defaults from db.DefaultPluginPoolConfig)
	Plugin_DB_MaxOpenConns    int    `json:"plugin_db_max_open_conns"`
	Plugin_DB_MaxIdleConns    int    `json:"plugin_db_max_idle_conns"`
	Plugin_DB_ConnMaxLifetime string `json:"plugin_db_conn_max_lifetime"`

	// Plugin HTTP integration configuration
	Plugin_Max_Request_Body  int64    `json:"plugin_max_request_body"`  // bytes, default 1MB
	Plugin_Max_Response_Body int64    `json:"plugin_max_response_body"` // bytes, default 5MB
	Plugin_Rate_Limit        int      `json:"plugin_rate_limit"`        // req/sec per IP, default 100
	Plugin_Max_Routes        int      `json:"plugin_max_routes"`        // per plugin, default 50
	Plugin_Trusted_Proxies   []string `json:"plugin_trusted_proxies"`   // CIDR list, empty = use RemoteAddr only

	// Plugin content hook configuration (Phase 3)
	Plugin_Hook_Reserve_VMs            int `json:"plugin_hook_reserve_vms"`            // VMs reserved for hooks per plugin, default 1
	Plugin_Hook_Max_Consecutive_Aborts int `json:"plugin_hook_max_consecutive_aborts"` // circuit breaker threshold, default 10
	Plugin_Hook_Max_Ops                int `json:"plugin_hook_max_ops"`                // reduced op budget for after-hooks, default 100
	Plugin_Hook_Max_Concurrent_After   int `json:"plugin_hook_max_concurrent_after"`   // max concurrent after-hook goroutines, default 10
	Plugin_Hook_Timeout_Ms             int `json:"plugin_hook_timeout_ms"`             // per-hook timeout in before-hooks (ms), default 2000
	Plugin_Hook_Event_Timeout_Ms       int `json:"plugin_hook_event_timeout_ms"`       // per-event total timeout for before-hook chain (ms), default 5000

	// Plugin outbound request engine configuration
	Plugin_Request_Timeout      int   `json:"plugin_request_timeout"`      // seconds, default 10
	Plugin_Request_Max_Response int64 `json:"plugin_request_max_response"` // bytes, default 1MB (1048576)
	Plugin_Request_Max_Body     int64 `json:"plugin_request_max_body"`     // bytes, default 1MB (1048576)
	Plugin_Request_Rate_Limit   int   `json:"plugin_request_rate_limit"`   // per plugin per domain per min, default 60
	Plugin_Request_Global_Rate  int   `json:"plugin_request_global_rate"`  // aggregate per min, default 600; 0 = unlimited
	Plugin_Request_CB_Failures  int   `json:"plugin_request_cb_failures"`  // consecutive failures to trip, default 5
	Plugin_Request_CB_Reset     int   `json:"plugin_request_cb_reset"`     // seconds before half-open probe, default 60
	Plugin_Request_Allow_Local  bool  `json:"plugin_request_allow_local"`  // allow HTTP to localhost (dev only), default false

	// Plugin production hardening (Phase 4)
	Plugin_Hot_Reload     bool   `json:"plugin_hot_reload"`     // default false (zero value) -- production opt-in only (S10)
	Plugin_Max_Failures   int    `json:"plugin_max_failures"`   // circuit breaker threshold, default 5
	Plugin_Reset_Interval string `json:"plugin_reset_interval"` // circuit breaker reset interval, default "60s"
	Plugin_Sync_Interval  string `json:"plugin_sync_interval"`  // DB state polling interval for multi-instance sync, default "10s"; "0" disables

	// Deploy sync configuration
	Deploy_Environments []DeployEnvironmentConfig `json:"deploy_environments"`
	Deploy_Snapshot_Dir string                    `json:"deploy_snapshot_dir"`

	// Tree composition depth limit
	Composition_Max_Depth int `json:"composition_max_depth"`

	// Publishing configuration
	Publish_Schedule_Interval int  `json:"publish_schedule_interval"` // seconds between scheduler ticks, default 60
	Version_Max_Per_Content   int  `json:"version_max_per_content"`   // max versions per content item, 0 = unlimited, default 50
	Node_Level_Publish        bool `json:"node_level_publish"`        // false (default): publish publishes root + all descendants; true: publish is per-node, "publish all" is separate action

	// Richtext editor toolbar configuration
	Richtext_Toolbar []string `json:"richtext_toolbar"`

	// Internationalization
	I18n_Enabled        bool   `json:"i18n_enabled"`        // default false
	I18n_Default_Locale string `json:"i18n_default_locale"` // default "en"

	// Webhooks
	Webhook_Enabled                 bool `json:"webhook_enabled"`
	Webhook_Timeout                 int  `json:"webhook_timeout"`
	Webhook_Max_Retries             int  `json:"webhook_max_retries"`
	Webhook_Workers                 int  `json:"webhook_workers"`
	Webhook_Allow_HTTP              bool `json:"webhook_allow_http"`
	Webhook_Delivery_Retention_Days int  `json:"webhook_delivery_retention_days"`

	// MCP server (Model Context Protocol for AI tooling)
	MCP_Enabled bool   `json:"mcp_enabled"`
	MCP_API_Key string `json:"mcp_api_key"` // API key for authenticating MCP clients

	// Search
	Search_Enabled bool   `json:"search_enabled"`
	Search_Path    string `json:"search_path"`

	KeyBindings KeyMap `json:"keybindings"`
}

// DeployEnvironmentConfig describes a remote Modula instance for deploy operations.
// APIKey supports ${VAR} expansion via the existing config system.
type DeployEnvironmentConfig struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	APIKey string `json:"api_key"`
}

// BucketEndpointURL returns Bucket_Endpoint prefixed with the scheme
// determined by Environment. Non-TLS environments get http, all others https.
// This is used for S3 API calls (internal/Docker hostname is fine).
func (c Config) BucketEndpointURL() string {
	if c.Bucket_Endpoint == "" {
		return ""
	}
	scheme := "https"
	if c.Environment.UsesHTTPScheme() {
		scheme = "http"
	}
	return scheme + "://" + c.Bucket_Endpoint
}

// BucketPublicURL returns the public-facing base URL for constructing
// browser-accessible media URLs. Falls back to BucketEndpointURL if
// Bucket_Public_URL is not configured.
//
// In Docker environments, Bucket_Endpoint is typically a container hostname
// (e.g. minio:9000) which browsers cannot resolve. Set Bucket_Public_URL
// to the externally reachable address (e.g. http://localhost:9000).
func (c Config) BucketPublicURL() string {
	if c.Bucket_Public_URL != "" {
		return c.Bucket_Public_URL
	}
	return c.BucketEndpointURL()
}

// AdminBucketMedia returns the admin media bucket name, falling back to the shared media bucket.
func (c Config) AdminBucketMedia() string {
	if c.Bucket_Admin_Media != "" {
		return c.Bucket_Admin_Media
	}
	return c.Bucket_Media
}

// AdminBucketEndpoint returns the admin media endpoint, falling back to the shared endpoint.
func (c Config) AdminBucketEndpoint() string {
	if c.Bucket_Admin_Endpoint != "" {
		return c.Bucket_Admin_Endpoint
	}
	return c.Bucket_Endpoint
}

// AdminBucketEndpointURL returns the full admin bucket endpoint URL with scheme.
func (c Config) AdminBucketEndpointURL() string {
	endpoint := c.AdminBucketEndpoint()
	if endpoint == "" {
		return ""
	}
	scheme := "https"
	if c.Environment.UsesHTTPScheme() {
		scheme = "http"
	}
	return scheme + "://" + endpoint
}

// AdminBucketAccessKey returns the admin media access key, falling back to the shared key.
func (c Config) AdminBucketAccessKey() string {
	if c.Bucket_Admin_Access_Key != "" {
		return c.Bucket_Admin_Access_Key
	}
	return c.Bucket_Access_Key
}

// AdminBucketSecretKey returns the admin media secret key, falling back to the shared key.
func (c Config) AdminBucketSecretKey() string {
	if c.Bucket_Admin_Secret_Key != "" {
		return c.Bucket_Admin_Secret_Key
	}
	return c.Bucket_Secret_Key
}

// AdminBucketPublicURL returns the admin media public URL, falling back to the shared public URL.
func (c Config) AdminBucketPublicURL() string {
	if c.Bucket_Admin_Public_URL != "" {
		return c.Bucket_Admin_Public_URL
	}
	return c.BucketPublicURL()
}

// CompositionMaxDepth returns the configured maximum composition depth.
// Falls back to 10 if no positive value is configured.
func (c Config) CompositionMaxDepth() int {
	if c.Composition_Max_Depth <= 0 {
		return 10
	}
	return c.Composition_Max_Depth
}

// PublishScheduleInterval returns the configured interval in seconds between
// scheduler ticks for scheduled publishing. Falls back to 60 if not configured.
func (c Config) PublishScheduleInterval() int {
	if c.Publish_Schedule_Interval <= 0 {
		return 60
	}
	return c.Publish_Schedule_Interval
}

// VersionMaxPerContent returns the maximum number of versions to retain per
// content item. Falls back to 50 if not configured. 0 means unlimited.
func (c Config) VersionMaxPerContent() int {
	if c.Version_Max_Per_Content < 0 {
		return 50
	}
	if c.Version_Max_Per_Content == 0 {
		return 50
	}
	return c.Version_Max_Per_Content
}

// RichtextToolbar returns the configured default toolbar buttons for richtext fields.
// Falls back to a sensible default set if not configured.
func (c Config) RichtextToolbar() []string {
	if len(c.Richtext_Toolbar) == 0 {
		return []string{"bold", "italic", "h1", "h2", "h3", "link", "ul", "ol", "preview"}
	}
	return c.Richtext_Toolbar
}

// WebhookEnabled returns whether webhooks are active.
func (c Config) WebhookEnabled() bool { return c.Webhook_Enabled }

// WebhookTimeout returns the HTTP timeout in seconds for webhook delivery.
// Falls back to 10 if not configured.
func (c Config) WebhookTimeout() int {
	if c.Webhook_Timeout <= 0 {
		return 10
	}
	return c.Webhook_Timeout
}

// WebhookMaxRetries returns the maximum number of delivery retry attempts.
// Falls back to 3 if not configured.
func (c Config) WebhookMaxRetries() int {
	if c.Webhook_Max_Retries <= 0 {
		return 3
	}
	return c.Webhook_Max_Retries
}

// WebhookWorkers returns the number of concurrent delivery workers.
// Falls back to 4 if not configured.
func (c Config) WebhookWorkers() int {
	if c.Webhook_Workers <= 0 {
		return 4
	}
	return c.Webhook_Workers
}

// WebhookAllowHTTP returns whether non-TLS webhook URLs are allowed (dev only).
func (c Config) WebhookAllowHTTP() bool { return c.Webhook_Allow_HTTP }

// WebhookDeliveryRetentionDays returns the number of days to retain completed deliveries.
// Falls back to 30 if not configured. 0 means unlimited retention.
func (c Config) WebhookDeliveryRetentionDays() int {
	if c.Webhook_Delivery_Retention_Days < 0 {
		return 30
	}
	if c.Webhook_Delivery_Retention_Days == 0 {
		// Distinguish "not set" (zero value) from explicit 0 (unlimited).
		// Since JSON unmarshaling defaults to 0, treat 0 as "use default 30".
		// Users must set to -1 or a positive value to override.
		return 30
	}
	return c.Webhook_Delivery_Retention_Days
}

// SearchEnabled returns whether search is enabled.
func (c Config) SearchEnabled() bool { return c.Search_Enabled }

// I18nEnabled returns whether internationalization is active.
func (c Config) I18nEnabled() bool { return c.I18n_Enabled }

// I18nDefaultLocale returns the configured default locale code.
// Falls back to "en" if not configured.
func (c Config) I18nDefaultLocale() string {
	if c.I18n_Default_Locale == "" {
		return "en"
	}
	return c.I18n_Default_Locale
}

// MaxUploadSize returns the configured maximum upload size in bytes.
// Falls back to 10 MB if no positive value is configured, ensuring
// backward compatibility with config files that omit this field.
func (c Config) MaxUploadSize() int64 {
	if c.Max_Upload_Size <= 0 {
		return 10 << 20 // 10 MB fallback
	}
	return c.Max_Upload_Size
}

// IsValidOutputFormat checks if the given format string is valid
func IsValidOutputFormat(format string) bool {
	switch OutputFormat(format) {
	case FormatContentful, FormatSanity, FormatStrapi, FormatWordPress, FormatClean, FormatRaw, FormatDefault:
		return true
	default:
		return false
	}
}

// GetValidOutputFormats returns a slice of all valid output formats
func GetValidOutputFormats() []string {
	return []string{
		string(FormatContentful),
		string(FormatSanity),
		string(FormatStrapi),
		string(FormatWordPress),
		string(FormatClean),
		string(FormatRaw),
	}
}

// IsValidEmailProvider checks if the given provider string is valid.
func IsValidEmailProvider(provider string) bool {
	switch EmailProvider(provider) {
	case EmailDisabled, EmailSmtp, EmailSendGrid, EmailSES, EmailPostmark:
		return true
	default:
		return false
	}
}

// GetValidEmailProviders returns a slice of all valid email provider values.
func GetValidEmailProviders() []string {
	return []string{
		string(EmailSmtp),
		string(EmailSendGrid),
		string(EmailSES),
		string(EmailPostmark),
	}
}
