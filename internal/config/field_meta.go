package config

// FieldCategory groups related configuration fields for display and filtering.
type FieldCategory string

const (
	CategoryServer        FieldCategory = "server"
	CategoryDatabase      FieldCategory = "database"
	CategoryStorage       FieldCategory = "storage"
	CategoryCORS          FieldCategory = "cors"
	CategoryCookie        FieldCategory = "cookie"
	CategoryOAuth         FieldCategory = "oauth"
	CategoryObservability FieldCategory = "observability"
	CategoryEmail         FieldCategory = "email"
	CategoryPlugin        FieldCategory = "plugin"
	CategoryUpdate        FieldCategory = "update"
	CategoryMisc          FieldCategory = "misc"
)

// AllCategories returns all field categories in display order.
func AllCategories() []FieldCategory {
	return []FieldCategory{
		CategoryServer,
		CategoryDatabase,
		CategoryStorage,
		CategoryCORS,
		CategoryCookie,
		CategoryOAuth,
		CategoryObservability,
		CategoryEmail,
		CategoryPlugin,
		CategoryUpdate,
		CategoryMisc,
	}
}

// CategoryLabel returns a human-readable label for a category.
func CategoryLabel(c FieldCategory) string {
	switch c {
	case CategoryServer:
		return "Server Settings"
	case CategoryDatabase:
		return "Database Settings"
	case CategoryStorage:
		return "Storage (S3) Settings"
	case CategoryCORS:
		return "CORS Settings"
	case CategoryCookie:
		return "Cookie Settings"
	case CategoryOAuth:
		return "OAuth Settings"
	case CategoryObservability:
		return "Observability Settings"
	case CategoryEmail:
		return "Email Settings"
	case CategoryPlugin:
		return "Plugin Settings"
	case CategoryUpdate:
		return "Update Settings"
	case CategoryMisc:
		return "Misc Settings"
	default:
		return string(c)
	}
}

// FieldMeta describes a single configuration field for display and validation.
type FieldMeta struct {
	JSONKey       string
	Label         string
	Category      FieldCategory
	HotReloadable bool
	Sensitive     bool
	Required      bool
	Description   string
	Example       string
}

// FieldRegistry enumerates all Config fields with their metadata.
var FieldRegistry = []FieldMeta{
	// Server
	{JSONKey: "environment", Label: "Environment", Category: CategoryServer, HotReloadable: false, Description: "Runtime environment (local, development, staging, production, docker, http-only)", Example: "production"},
	{JSONKey: "os", Label: "Operating System", Category: CategoryServer, HotReloadable: false, Description: "Target OS", Example: "linux"},
	{JSONKey: "port", Label: "HTTP Port", Category: CategoryServer, HotReloadable: false, Required: true, Description: "HTTP listen address (e.g. :8080)", Example: ":8080"},
	{JSONKey: "ssl_port", Label: "HTTPS Port", Category: CategoryServer, HotReloadable: false, Description: "HTTPS listen address (e.g. :4000)", Example: ":4000"},
	{JSONKey: "cert_dir", Label: "Certificate Directory", Category: CategoryServer, HotReloadable: false, Description: "Path to TLS certificate directory", Example: "/etc/modula/certs"},
	{JSONKey: "ssh_host", Label: "SSH Host", Category: CategoryServer, HotReloadable: false, Description: "SSH server bind host", Example: "0.0.0.0"},
	{JSONKey: "ssh_port", Label: "SSH Port", Category: CategoryServer, HotReloadable: false, Required: true, Description: "SSH server port", Example: "2222"},
	{JSONKey: "client_site", Label: "Client Site", Category: CategoryServer, HotReloadable: true, Description: "Client site hostname", Example: "www.example.com"},
	{JSONKey: "admin_site", Label: "Admin Site", Category: CategoryServer, HotReloadable: true, Description: "Admin site hostname", Example: "admin.example.com"},
	{JSONKey: "log_path", Label: "Log Path", Category: CategoryServer, HotReloadable: true, Description: "Path for log files", Example: "/var/log/modula"},
	{JSONKey: "auth_salt", Label: "Auth Salt", Category: CategoryServer, HotReloadable: false, Sensitive: true, Description: "Salt for password hashing", Example: "a-random-32-char-string"},
	{JSONKey: "node_id", Label: "Node ID", Category: CategoryServer, HotReloadable: false, Description: "Unique node identifier (ULID)", Example: "01HY5N3K0G1JQXKM8V7Z4R6W9T"},
	{JSONKey: "space_id", Label: "Space ID", Category: CategoryServer, HotReloadable: false, Description: "Space identifier", Example: "my-space"},
	{JSONKey: "output_format", Label: "Output Format", Category: CategoryServer, HotReloadable: true, Description: "Default API response format (contentful, sanity, strapi, wordpress, clean, raw)", Example: "clean"},
	{JSONKey: "custom_style_path", Label: "Custom Style Path", Category: CategoryServer, HotReloadable: true, Description: "Path to custom TUI style file", Example: "./styles/dark.json"},
	{JSONKey: "max_upload_size", Label: "Max Upload Size", Category: CategoryServer, HotReloadable: true, Description: "Maximum file upload size in bytes", Example: "10485760"},

	// Database
	{JSONKey: "db_driver", Label: "DB Driver", Category: CategoryDatabase, HotReloadable: false, Required: true, Description: "Database driver (sqlite, mysql, postgres)", Example: "sqlite"},
	{JSONKey: "db_url", Label: "DB URL", Category: CategoryDatabase, HotReloadable: false, Required: true, Description: "Database connection URL", Example: "./modula.db"},
	{JSONKey: "db_name", Label: "DB Name", Category: CategoryDatabase, HotReloadable: false, Description: "Database name", Example: "modulacms"},
	{JSONKey: "db_username", Label: "DB Username", Category: CategoryDatabase, HotReloadable: false, Description: "Database username", Example: "modula"},
	{JSONKey: "db_password", Label: "DB Password", Category: CategoryDatabase, HotReloadable: false, Sensitive: true, Description: "Database password", Example: "secret"},

	// Storage (S3)
	{JSONKey: "bucket_region", Label: "Bucket Region", Category: CategoryStorage, HotReloadable: true, Description: "S3 bucket region", Example: "us-east-1"},
	{JSONKey: "bucket_media", Label: "Media Bucket", Category: CategoryStorage, HotReloadable: true, Description: "S3 bucket name for media", Example: "mysite-media"},
	{JSONKey: "bucket_backup", Label: "Backup Bucket", Category: CategoryStorage, HotReloadable: true, Description: "S3 bucket name for backups", Example: "mysite-backups"},
	{JSONKey: "bucket_endpoint", Label: "Bucket Endpoint", Category: CategoryStorage, HotReloadable: true, Description: "S3-compatible endpoint (without scheme)", Example: "s3.amazonaws.com"},
	{JSONKey: "bucket_access_key", Label: "Bucket Access Key", Category: CategoryStorage, HotReloadable: true, Sensitive: true, Description: "S3 access key", Example: "AKIAIOSFODNN7EXAMPLE"},
	{JSONKey: "bucket_secret_key", Label: "Bucket Secret Key", Category: CategoryStorage, HotReloadable: true, Sensitive: true, Description: "S3 secret key", Example: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"},
	{JSONKey: "bucket_public_url", Label: "Bucket Public URL", Category: CategoryStorage, HotReloadable: true, Description: "Public-facing URL for media assets", Example: "https://cdn.example.com"},
	{JSONKey: "bucket_default_acl", Label: "Bucket Default ACL", Category: CategoryStorage, HotReloadable: true, Description: "Default ACL for uploaded objects", Example: "public-read"},
	{JSONKey: "bucket_force_path_style", Label: "Force Path Style", Category: CategoryStorage, HotReloadable: true, Description: "Use path-style S3 URLs", Example: "true"},
	{JSONKey: "backup_option", Label: "Backup Option", Category: CategoryStorage, HotReloadable: true, Description: "Backup storage location", Example: "s3"},
	{JSONKey: "backup_paths", Label: "Backup Paths", Category: CategoryStorage, HotReloadable: true, Description: "Additional backup paths", Example: "/var/backups/modula"},

	// CORS
	{JSONKey: "cors_origins", Label: "Allowed Origins", Category: CategoryCORS, HotReloadable: true, Description: "CORS allowed origins", Example: "https://example.com,https://admin.example.com"},
	{JSONKey: "cors_methods", Label: "Allowed Methods", Category: CategoryCORS, HotReloadable: true, Description: "CORS allowed HTTP methods", Example: "GET,POST,PUT,DELETE,OPTIONS"},
	{JSONKey: "cors_headers", Label: "Allowed Headers", Category: CategoryCORS, HotReloadable: true, Description: "CORS allowed request headers", Example: "Content-Type,Authorization,X-CSRF-Token"},
	{JSONKey: "cors_credentials", Label: "Allow Credentials", Category: CategoryCORS, HotReloadable: true, Description: "Allow credentials in CORS requests", Example: "true"},

	// Cookie
	{JSONKey: "cookie_name", Label: "Cookie Name", Category: CategoryCookie, HotReloadable: true, Description: "Authentication cookie name", Example: "modula_session"},
	{JSONKey: "cookie_duration", Label: "Cookie Duration", Category: CategoryCookie, HotReloadable: true, Description: "Cookie lifetime (e.g. 1w, 24h)", Example: "1w"},
	{JSONKey: "cookie_secure", Label: "Cookie Secure", Category: CategoryCookie, HotReloadable: true, Description: "Set Secure flag on cookies", Example: "true"},
	{JSONKey: "cookie_samesite", Label: "Cookie SameSite", Category: CategoryCookie, HotReloadable: true, Description: "SameSite policy (strict, lax, none)", Example: "lax"},

	// OAuth
	{JSONKey: "oauth_client_id", Label: "OAuth Client ID", Category: CategoryOAuth, HotReloadable: true, Sensitive: true, Description: "OAuth client ID", Example: "123456789.apps.googleusercontent.com"},
	{JSONKey: "oauth_client_secret", Label: "OAuth Client Secret", Category: CategoryOAuth, HotReloadable: true, Sensitive: true, Description: "OAuth client secret", Example: "GOCSPX-xxxxxxxxxxxxxxxxxxxx"},
	{JSONKey: "oauth_scopes", Label: "OAuth Scopes", Category: CategoryOAuth, HotReloadable: true, Description: "OAuth scopes", Example: "openid,email,profile"},
	{JSONKey: "oauth_provider_name", Label: "OAuth Provider", Category: CategoryOAuth, HotReloadable: true, Description: "OAuth provider name", Example: "google"},
	{JSONKey: "oauth_redirect_url", Label: "OAuth Redirect URL", Category: CategoryOAuth, HotReloadable: true, Description: "OAuth redirect callback URL", Example: "https://example.com/auth/callback"},
	{JSONKey: "oauth_success_redirect", Label: "OAuth Success Redirect", Category: CategoryOAuth, HotReloadable: true, Description: "URL to redirect after OAuth success", Example: "https://admin.example.com/dashboard"},

	// Observability
	{JSONKey: "observability_enabled", Label: "Enabled", Category: CategoryObservability, HotReloadable: true, Description: "Enable observability", Example: "true"},
	{JSONKey: "observability_provider", Label: "Provider", Category: CategoryObservability, HotReloadable: true, Description: "Observability provider (sentry, datadog, etc.)", Example: "sentry"},
	{JSONKey: "observability_dsn", Label: "DSN", Category: CategoryObservability, HotReloadable: true, Sensitive: true, Description: "Connection string / DSN", Example: "https://key@sentry.io/12345"},
	{JSONKey: "observability_environment", Label: "Environment", Category: CategoryObservability, HotReloadable: true, Description: "Observability environment label", Example: "production"},
	{JSONKey: "observability_release", Label: "Release", Category: CategoryObservability, HotReloadable: true, Description: "Version / release identifier", Example: "v1.2.3"},
	{JSONKey: "observability_sample_rate", Label: "Sample Rate", Category: CategoryObservability, HotReloadable: true, Description: "Event sample rate (0.0–1.0)", Example: "1.0"},
	{JSONKey: "observability_traces_rate", Label: "Traces Rate", Category: CategoryObservability, HotReloadable: true, Description: "Trace sample rate (0.0–1.0)", Example: "0.2"},
	{JSONKey: "observability_send_pii", Label: "Send PII", Category: CategoryObservability, HotReloadable: true, Description: "Send personally identifiable info", Example: "false"},
	{JSONKey: "observability_debug", Label: "Debug", Category: CategoryObservability, HotReloadable: true, Description: "Enable observability debug logging", Example: "false"},
	{JSONKey: "observability_server_name", Label: "Server Name", Category: CategoryObservability, HotReloadable: true, Description: "Server/instance name", Example: "cms-prod-01"},
	{JSONKey: "observability_flush_interval", Label: "Flush Interval", Category: CategoryObservability, HotReloadable: true, Description: "Metrics flush interval (e.g. 30s)", Example: "30s"},

	// Email
	{JSONKey: "email_enabled", Label: "Email Enabled", Category: CategoryEmail, HotReloadable: true, Description: "Enable email sending", Example: "true"},
	{JSONKey: "email_provider", Label: "Email Provider", Category: CategoryEmail, HotReloadable: true, Description: "Email provider (smtp, sendgrid, ses, postmark)", Example: "smtp"},
	{JSONKey: "email_from_address", Label: "From Address", Category: CategoryEmail, HotReloadable: true, Description: "Sender email address", Example: "noreply@example.com"},
	{JSONKey: "email_from_name", Label: "From Name", Category: CategoryEmail, HotReloadable: true, Description: "Sender display name", Example: "ModulaCMS"},
	{JSONKey: "email_host", Label: "SMTP Host", Category: CategoryEmail, HotReloadable: true, Description: "SMTP server hostname", Example: "smtp.gmail.com"},
	{JSONKey: "email_port", Label: "SMTP Port", Category: CategoryEmail, HotReloadable: true, Description: "SMTP server port", Example: "587"},
	{JSONKey: "email_username", Label: "SMTP Username", Category: CategoryEmail, HotReloadable: true, Description: "SMTP authentication username", Example: "user@gmail.com"},
	{JSONKey: "email_password", Label: "SMTP Password", Category: CategoryEmail, HotReloadable: true, Sensitive: true, Description: "SMTP authentication password", Example: "app-specific-password"},
	{JSONKey: "email_tls", Label: "Require TLS", Category: CategoryEmail, HotReloadable: true, Description: "Require TLS for SMTP connections", Example: "true"},
	{JSONKey: "email_api_key", Label: "API Key", Category: CategoryEmail, HotReloadable: true, Sensitive: true, Description: "API key for HTTP email providers", Example: "SG.xxxxxxxxxxxxxxxxxxxx"},
	{JSONKey: "email_api_endpoint", Label: "API Endpoint", Category: CategoryEmail, HotReloadable: true, Description: "Custom API endpoint URL", Example: "https://api.sendgrid.com/v3/mail/send"},
	{JSONKey: "email_reply_to", Label: "Reply-To", Category: CategoryEmail, HotReloadable: true, Description: "Default reply-to address", Example: "support@example.com"},
	{JSONKey: "email_aws_access_key_id", Label: "AWS Access Key ID", Category: CategoryEmail, HotReloadable: true, Sensitive: true, Description: "AWS access key ID for SES", Example: "AKIAIOSFODNN7EXAMPLE"},
	{JSONKey: "email_aws_secret_access_key", Label: "AWS Secret Access Key", Category: CategoryEmail, HotReloadable: true, Sensitive: true, Description: "AWS secret access key for SES", Example: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"},

	// Plugin
	{JSONKey: "plugin_enabled", Label: "Plugin Enabled", Category: CategoryPlugin, HotReloadable: false, Description: "Enable plugin system", Example: "true"},
	{JSONKey: "plugin_directory", Label: "Plugin Directory", Category: CategoryPlugin, HotReloadable: false, Description: "Path to plugins directory", Example: "./plugins"},
	{JSONKey: "plugin_max_vms", Label: "Max VMs", Category: CategoryPlugin, HotReloadable: true, Description: "Max Lua VMs per plugin", Example: "4"},
	{JSONKey: "plugin_timeout", Label: "Timeout (s)", Category: CategoryPlugin, HotReloadable: true, Description: "Plugin execution timeout in seconds", Example: "5"},
	{JSONKey: "plugin_max_ops", Label: "Max Ops", Category: CategoryPlugin, HotReloadable: true, Description: "Max operations per VM checkout", Example: "1000"},
	{JSONKey: "plugin_hot_reload", Label: "Hot Reload", Category: CategoryPlugin, HotReloadable: false, Description: "Enable plugin hot reload (file watcher)", Example: "true"},
	{JSONKey: "plugin_max_failures", Label: "Max Failures", Category: CategoryPlugin, HotReloadable: true, Description: "Circuit breaker failure threshold", Example: "5"},
	{JSONKey: "plugin_reset_interval", Label: "Reset Interval", Category: CategoryPlugin, HotReloadable: true, Description: "Circuit breaker reset interval", Example: "60s"},
	{JSONKey: "plugin_rate_limit", Label: "Rate Limit", Category: CategoryPlugin, HotReloadable: true, Description: "Plugin HTTP rate limit (req/sec/IP)", Example: "100"},
	{JSONKey: "plugin_max_routes", Label: "Max Routes", Category: CategoryPlugin, HotReloadable: true, Description: "Max HTTP routes per plugin", Example: "10"},
	{JSONKey: "plugin_max_request_body", Label: "Max Request Body", Category: CategoryPlugin, HotReloadable: true, Description: "Max request body size (bytes)", Example: "1048576"},
	{JSONKey: "plugin_max_response_body", Label: "Max Response Body", Category: CategoryPlugin, HotReloadable: true, Description: "Max response body size (bytes)", Example: "5242880"},

	// Misc
	{JSONKey: "richtext_toolbar", Label: "Richtext Toolbar", Category: CategoryMisc, HotReloadable: true, Description: "Default toolbar buttons for richtext fields", Example: "bold,italic,link,heading,list,quote,code"},

	// Update
	{JSONKey: "update_auto_enabled", Label: "Auto Update", Category: CategoryUpdate, HotReloadable: true, Description: "Enable automatic updates", Example: "true"},
	{JSONKey: "update_check_interval", Label: "Check Interval", Category: CategoryUpdate, HotReloadable: true, Description: "Update check interval (e.g. startup, 24h)", Example: "24h"},
	{JSONKey: "update_channel", Label: "Channel", Category: CategoryUpdate, HotReloadable: true, Description: "Update channel (stable, beta)", Example: "stable"},
	{JSONKey: "update_notify_only", Label: "Notify Only", Category: CategoryUpdate, HotReloadable: true, Description: "Only notify about updates, don't auto-install", Example: "false"},
}

// FieldsByCategory returns the fields that belong to the given category.
func FieldsByCategory(category FieldCategory) []FieldMeta {
	var result []FieldMeta
	for _, f := range FieldRegistry {
		if f.Category == category {
			result = append(result, f)
		}
	}
	return result
}

// FieldByKey looks up a field by its JSON key. Returns the field and true if found.
func FieldByKey(key string) (FieldMeta, bool) {
	for _, f := range FieldRegistry {
		if f.JSONKey == key {
			return f, true
		}
	}
	return FieldMeta{}, false
}

// SensitiveKeys returns the set of JSON keys that are marked sensitive.
func SensitiveKeys() map[string]bool {
	keys := make(map[string]bool)
	for _, f := range FieldRegistry {
		if f.Sensitive {
			keys[f.JSONKey] = true
		}
	}
	return keys
}
