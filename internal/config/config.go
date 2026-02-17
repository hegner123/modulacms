// Package config provides configuration management for ModulaCMS, including
// database drivers, OAuth endpoints, S3-compatible storage buckets, CORS
// settings, SSL/TLS configuration, plugin runtime options, and observability.
package config

// Endpoint identifies OAuth provider endpoint types.
type Endpoint string

// DbDriver specifies which database backend to use.
type DbDriver string

// OutputFormat defines the API response structure for content endpoints.
type OutputFormat string

// EmailProvider specifies which email sending backend to use.
type EmailProvider string

// OAuth endpoint keys used in the Oauth_Endpoint configuration map.
const (
	OauthAuthURL     Endpoint = "oauth_auth_url"
	OauthTokenURL    Endpoint = "oauth_token_url"
	OauthUserInfoURL Endpoint = "oauth_userinfo_url"
)

// Supported database drivers for ModulaCMS.
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

// Config holds all runtime configuration for ModulaCMS including server settings,
// database credentials, OAuth providers, S3-compatible storage, CORS policies,
// plugin runtime limits, and observability integration.
type Config struct {
	Environment         string              `json:"environment"`
	OS                  string              `json:"os"`
	Environment_Hosts   map[string]string   `json:"environment_hosts"`
	Port                string              `json:"port"`
	SSL_Port            string              `json:"ssl_port"`
	Cert_Dir            string              `json:"cert_dir"`
	Client_Site         string              `json:"client_site"`
	Admin_Site          string              `json:"admin_site"`
	SSH_Host            string              `json:"ssh_host"`
	SSH_Port            string              `json:"ssh_port"`
	Options             map[string][]any    `json:"options"`
	Log_Path            string              `json:"log_path"`
	Auth_Salt           string              `json:"auth_salt"`
	Cookie_Name         string              `json:"cookie_name"`
	Cookie_Duration     string              `json:"cookie_duration"`
	Cookie_Secure       bool                `json:"cookie_secure"`
	Cookie_SameSite     string              `json:"cookie_samesite"`
	Db_Driver           DbDriver            `json:"db_driver"`
	Db_URL              string              `json:"db_url"`
	Db_Name             string              `json:"db_name"`
	Db_User             string              `json:"db_username"`
	Db_Password         string              `json:"db_password"`
	Bucket_Region       string              `json:"bucket_region"`
	Bucket_Media        string              `json:"bucket_media"`
	Bucket_Backup       string              `json:"bucket_backup"`
	Bucket_Endpoint     string              `json:"bucket_endpoint"`
	Bucket_Access_Key   string              `json:"bucket_access_key"`
	Bucket_Secret_Key   string              `json:"bucket_secret_key"`
	Bucket_Public_URL        string `json:"bucket_public_url"`
	Bucket_Default_ACL       string `json:"bucket_default_acl"`
	Bucket_Force_Path_Style  bool   `json:"bucket_force_path_style"`
	Max_Upload_Size          int64  `json:"max_upload_size"` // bytes, default 10MB (10485760)
	Backup_Option       string              `json:"backup_option"`
	Backup_Paths        []string            `json:"backup_paths"`
	Oauth_Client_Id        string              `json:"oauth_client_id"`
	Oauth_Client_Secret    string              `json:"oauth_client_secret"`
	Oauth_Scopes           []string            `json:"oauth_scopes"`
	Oauth_Endpoint         map[Endpoint]string `json:"oauth_endpoint"`
	Oauth_Provider_Name    string              `json:"oauth_provider_name"`
	Oauth_Redirect_URL     string              `json:"oauth_redirect_url"`
	Oauth_Success_Redirect string              `json:"oauth_success_redirect"`
	Cors_Origins        []string            `json:"cors_origins"`
	Cors_Methods        []string            `json:"cors_methods"`
	Cors_Headers        []string            `json:"cors_headers"`
	Cors_Credentials    bool                `json:"cors_credentials"`
	Custom_Style_Path   string              `json:"custom_style_path"`
	Update_Auto_Enabled   bool         `json:"update_auto_enabled"`
	Update_Check_Interval string       `json:"update_check_interval"`
	Update_Channel        string       `json:"update_channel"`
	Update_Notify_Only    bool         `json:"update_notify_only"`
	Output_Format         OutputFormat `json:"output_format"`
	Space_ID              string       `json:"space_id"`
	Node_ID               string       `json:"node_id"`

	// Observability - Metrics and Error Tracking
	Observability_Enabled        bool    `json:"observability_enabled"`
	Observability_Provider       string  `json:"observability_provider"`        // "sentry", "datadog", "newrelic", etc.
	Observability_DSN            string  `json:"observability_dsn"`             // Sentry DSN or equivalent connection string
	Observability_Environment    string  `json:"observability_environment"`     // "production", "staging", "development"
	Observability_Release        string  `json:"observability_release"`         // Version/release identifier
	Observability_Sample_Rate    float64 `json:"observability_sample_rate"`     // 0.0 to 1.0 - percentage of events to send
	Observability_Traces_Rate    float64 `json:"observability_traces_rate"`     // 0.0 to 1.0 - percentage of traces to send
	Observability_Send_PII       bool    `json:"observability_send_pii"`        // Whether to send personally identifiable info
	Observability_Debug          bool    `json:"observability_debug"`           // Enable debug logging for observability client
	Observability_Server_Name    string  `json:"observability_server_name"`     // Server/instance name
	Observability_Flush_Interval string  `json:"observability_flush_interval"`  // How often to flush metrics (e.g., "30s", "1m")
	Observability_Tags           map[string]string `json:"observability_tags"` // Global tags for all metrics/events

	// Email provider configuration
	Email_Enabled      bool          `json:"email_enabled"`
	Email_Provider     EmailProvider `json:"email_provider"`
	Email_From_Address string        `json:"email_from_address"`
	Email_From_Name    string        `json:"email_from_name"`
	Email_Host         string        `json:"email_host"`
	Email_Port         int           `json:"email_port"`
	Email_Username     string        `json:"email_username"`
	Email_Password     string        `json:"email_password"`
	Email_TLS          bool          `json:"email_tls"`
	Email_API_Key              string        `json:"email_api_key"`
	Email_API_Endpoint         string        `json:"email_api_endpoint"`
	Email_Reply_To             string        `json:"email_reply_to"`
	Email_AWS_Access_Key_ID    string        `json:"email_aws_access_key_id"`
	Email_AWS_Secret_Access_Key string       `json:"email_aws_secret_access_key"`

	// Plugin runtime configuration
	Plugin_Enabled   bool   `json:"plugin_enabled"`
	Plugin_Directory string `json:"plugin_directory"`  // path to plugins dir, e.g. "./plugins/"
	Plugin_Max_VMs   int    `json:"plugin_max_vms"`    // per plugin, default 4
	Plugin_Timeout   int    `json:"plugin_timeout"`    // seconds, default 5
	Plugin_Max_Ops   int    `json:"plugin_max_ops"`    // per VM checkout, default 1000

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
	Plugin_Hook_Reserve_VMs          int `json:"plugin_hook_reserve_vms"`           // VMs reserved for hooks per plugin, default 1
	Plugin_Hook_Max_Consecutive_Aborts int `json:"plugin_hook_max_consecutive_aborts"` // circuit breaker threshold, default 10
	Plugin_Hook_Max_Ops              int `json:"plugin_hook_max_ops"`               // reduced op budget for after-hooks, default 100
	Plugin_Hook_Max_Concurrent_After int `json:"plugin_hook_max_concurrent_after"`  // max concurrent after-hook goroutines, default 10
	Plugin_Hook_Timeout_Ms           int `json:"plugin_hook_timeout_ms"`            // per-hook timeout in before-hooks (ms), default 2000
	Plugin_Hook_Event_Timeout_Ms     int `json:"plugin_hook_event_timeout_ms"`      // per-event total timeout for before-hook chain (ms), default 5000

	// Plugin production hardening (Phase 4)
	Plugin_Hot_Reload     bool   `json:"plugin_hot_reload"`      // default false (zero value) -- production opt-in only (S10)
	Plugin_Max_Failures   int    `json:"plugin_max_failures"`    // circuit breaker threshold, default 5
	Plugin_Reset_Interval string `json:"plugin_reset_interval"`  // circuit breaker reset interval, default "60s"

	KeyBindings KeyMap `json:"keybindings"`
}

// BucketEndpointURL returns Bucket_Endpoint prefixed with the scheme
// determined by Environment. Non-TLS environments get http, all others https.
// This is used for S3 API calls (internal/Docker hostname is fine).
func (c Config) BucketEndpointURL() string {
	if c.Bucket_Endpoint == "" {
		return ""
	}
	scheme := "https"
	if c.Environment == "http-only" || c.Environment == "docker" {
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
