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
}

// FieldRegistry enumerates all Config fields with their metadata.
var FieldRegistry = []FieldMeta{
	// Server
	{JSONKey: "environment", Label: "Environment", Category: CategoryServer, HotReloadable: false, Description: "Runtime environment (local, development, staging, production, docker, http-only)"},
	{JSONKey: "os", Label: "Operating System", Category: CategoryServer, HotReloadable: false, Description: "Target OS"},
	{JSONKey: "port", Label: "HTTP Port", Category: CategoryServer, HotReloadable: false, Required: true, Description: "HTTP listen address (e.g. :8080)"},
	{JSONKey: "ssl_port", Label: "HTTPS Port", Category: CategoryServer, HotReloadable: false, Description: "HTTPS listen address (e.g. :4000)"},
	{JSONKey: "cert_dir", Label: "Certificate Directory", Category: CategoryServer, HotReloadable: false, Description: "Path to TLS certificate directory"},
	{JSONKey: "ssh_host", Label: "SSH Host", Category: CategoryServer, HotReloadable: false, Description: "SSH server bind host"},
	{JSONKey: "ssh_port", Label: "SSH Port", Category: CategoryServer, HotReloadable: false, Required: true, Description: "SSH server port"},
	{JSONKey: "client_site", Label: "Client Site", Category: CategoryServer, HotReloadable: true, Description: "Client site hostname"},
	{JSONKey: "admin_site", Label: "Admin Site", Category: CategoryServer, HotReloadable: true, Description: "Admin site hostname"},
	{JSONKey: "log_path", Label: "Log Path", Category: CategoryServer, HotReloadable: true, Description: "Path for log files"},
	{JSONKey: "auth_salt", Label: "Auth Salt", Category: CategoryServer, HotReloadable: false, Sensitive: true, Description: "Salt for password hashing"},
	{JSONKey: "node_id", Label: "Node ID", Category: CategoryServer, HotReloadable: false, Description: "Unique node identifier (ULID)"},
	{JSONKey: "space_id", Label: "Space ID", Category: CategoryServer, HotReloadable: false, Description: "Space identifier"},
	{JSONKey: "output_format", Label: "Output Format", Category: CategoryServer, HotReloadable: true, Description: "Default API response format (contentful, sanity, strapi, wordpress, clean, raw)"},
	{JSONKey: "custom_style_path", Label: "Custom Style Path", Category: CategoryServer, HotReloadable: true, Description: "Path to custom TUI style file"},
	{JSONKey: "max_upload_size", Label: "Max Upload Size", Category: CategoryServer, HotReloadable: true, Description: "Maximum file upload size in bytes"},

	// Database
	{JSONKey: "db_driver", Label: "DB Driver", Category: CategoryDatabase, HotReloadable: false, Required: true, Description: "Database driver (sqlite, mysql, postgres)"},
	{JSONKey: "db_url", Label: "DB URL", Category: CategoryDatabase, HotReloadable: false, Required: true, Description: "Database connection URL"},
	{JSONKey: "db_name", Label: "DB Name", Category: CategoryDatabase, HotReloadable: false, Description: "Database name"},
	{JSONKey: "db_username", Label: "DB Username", Category: CategoryDatabase, HotReloadable: false, Description: "Database username"},
	{JSONKey: "db_password", Label: "DB Password", Category: CategoryDatabase, HotReloadable: false, Sensitive: true, Description: "Database password"},

	// Storage (S3)
	{JSONKey: "bucket_region", Label: "Bucket Region", Category: CategoryStorage, HotReloadable: true, Description: "S3 bucket region"},
	{JSONKey: "bucket_media", Label: "Media Bucket", Category: CategoryStorage, HotReloadable: true, Description: "S3 bucket name for media"},
	{JSONKey: "bucket_backup", Label: "Backup Bucket", Category: CategoryStorage, HotReloadable: true, Description: "S3 bucket name for backups"},
	{JSONKey: "bucket_endpoint", Label: "Bucket Endpoint", Category: CategoryStorage, HotReloadable: true, Description: "S3-compatible endpoint (without scheme)"},
	{JSONKey: "bucket_access_key", Label: "Bucket Access Key", Category: CategoryStorage, HotReloadable: true, Sensitive: true, Description: "S3 access key"},
	{JSONKey: "bucket_secret_key", Label: "Bucket Secret Key", Category: CategoryStorage, HotReloadable: true, Sensitive: true, Description: "S3 secret key"},
	{JSONKey: "bucket_public_url", Label: "Bucket Public URL", Category: CategoryStorage, HotReloadable: true, Description: "Public-facing URL for media assets"},
	{JSONKey: "bucket_default_acl", Label: "Bucket Default ACL", Category: CategoryStorage, HotReloadable: true, Description: "Default ACL for uploaded objects"},
	{JSONKey: "bucket_force_path_style", Label: "Force Path Style", Category: CategoryStorage, HotReloadable: true, Description: "Use path-style S3 URLs"},
	{JSONKey: "backup_option", Label: "Backup Option", Category: CategoryStorage, HotReloadable: true, Description: "Backup storage location"},
	{JSONKey: "backup_paths", Label: "Backup Paths", Category: CategoryStorage, HotReloadable: true, Description: "Additional backup paths"},

	// CORS
	{JSONKey: "cors_origins", Label: "Allowed Origins", Category: CategoryCORS, HotReloadable: true, Description: "CORS allowed origins"},
	{JSONKey: "cors_methods", Label: "Allowed Methods", Category: CategoryCORS, HotReloadable: true, Description: "CORS allowed HTTP methods"},
	{JSONKey: "cors_headers", Label: "Allowed Headers", Category: CategoryCORS, HotReloadable: true, Description: "CORS allowed request headers"},
	{JSONKey: "cors_credentials", Label: "Allow Credentials", Category: CategoryCORS, HotReloadable: true, Description: "Allow credentials in CORS requests"},

	// Cookie
	{JSONKey: "cookie_name", Label: "Cookie Name", Category: CategoryCookie, HotReloadable: true, Description: "Authentication cookie name"},
	{JSONKey: "cookie_duration", Label: "Cookie Duration", Category: CategoryCookie, HotReloadable: true, Description: "Cookie lifetime (e.g. 1w, 24h)"},
	{JSONKey: "cookie_secure", Label: "Cookie Secure", Category: CategoryCookie, HotReloadable: true, Description: "Set Secure flag on cookies"},
	{JSONKey: "cookie_samesite", Label: "Cookie SameSite", Category: CategoryCookie, HotReloadable: true, Description: "SameSite policy (strict, lax, none)"},

	// OAuth
	{JSONKey: "oauth_client_id", Label: "OAuth Client ID", Category: CategoryOAuth, HotReloadable: true, Sensitive: true, Description: "OAuth client ID"},
	{JSONKey: "oauth_client_secret", Label: "OAuth Client Secret", Category: CategoryOAuth, HotReloadable: true, Sensitive: true, Description: "OAuth client secret"},
	{JSONKey: "oauth_scopes", Label: "OAuth Scopes", Category: CategoryOAuth, HotReloadable: true, Description: "OAuth scopes"},
	{JSONKey: "oauth_provider_name", Label: "OAuth Provider", Category: CategoryOAuth, HotReloadable: true, Description: "OAuth provider name"},
	{JSONKey: "oauth_redirect_url", Label: "OAuth Redirect URL", Category: CategoryOAuth, HotReloadable: true, Description: "OAuth redirect callback URL"},
	{JSONKey: "oauth_success_redirect", Label: "OAuth Success Redirect", Category: CategoryOAuth, HotReloadable: true, Description: "URL to redirect after OAuth success"},

	// Observability
	{JSONKey: "observability_enabled", Label: "Enabled", Category: CategoryObservability, HotReloadable: true, Description: "Enable observability"},
	{JSONKey: "observability_provider", Label: "Provider", Category: CategoryObservability, HotReloadable: true, Description: "Observability provider (sentry, datadog, etc.)"},
	{JSONKey: "observability_dsn", Label: "DSN", Category: CategoryObservability, HotReloadable: true, Sensitive: true, Description: "Connection string / DSN"},
	{JSONKey: "observability_environment", Label: "Environment", Category: CategoryObservability, HotReloadable: true, Description: "Observability environment label"},
	{JSONKey: "observability_release", Label: "Release", Category: CategoryObservability, HotReloadable: true, Description: "Version / release identifier"},
	{JSONKey: "observability_sample_rate", Label: "Sample Rate", Category: CategoryObservability, HotReloadable: true, Description: "Event sample rate (0.0–1.0)"},
	{JSONKey: "observability_traces_rate", Label: "Traces Rate", Category: CategoryObservability, HotReloadable: true, Description: "Trace sample rate (0.0–1.0)"},
	{JSONKey: "observability_send_pii", Label: "Send PII", Category: CategoryObservability, HotReloadable: true, Description: "Send personally identifiable info"},
	{JSONKey: "observability_debug", Label: "Debug", Category: CategoryObservability, HotReloadable: true, Description: "Enable observability debug logging"},
	{JSONKey: "observability_server_name", Label: "Server Name", Category: CategoryObservability, HotReloadable: true, Description: "Server/instance name"},
	{JSONKey: "observability_flush_interval", Label: "Flush Interval", Category: CategoryObservability, HotReloadable: true, Description: "Metrics flush interval (e.g. 30s)"},

	// Plugin
	{JSONKey: "plugin_enabled", Label: "Plugin Enabled", Category: CategoryPlugin, HotReloadable: false, Description: "Enable plugin system"},
	{JSONKey: "plugin_directory", Label: "Plugin Directory", Category: CategoryPlugin, HotReloadable: false, Description: "Path to plugins directory"},
	{JSONKey: "plugin_max_vms", Label: "Max VMs", Category: CategoryPlugin, HotReloadable: true, Description: "Max Lua VMs per plugin"},
	{JSONKey: "plugin_timeout", Label: "Timeout (s)", Category: CategoryPlugin, HotReloadable: true, Description: "Plugin execution timeout in seconds"},
	{JSONKey: "plugin_max_ops", Label: "Max Ops", Category: CategoryPlugin, HotReloadable: true, Description: "Max operations per VM checkout"},
	{JSONKey: "plugin_hot_reload", Label: "Hot Reload", Category: CategoryPlugin, HotReloadable: false, Description: "Enable plugin hot reload (file watcher)"},
	{JSONKey: "plugin_max_failures", Label: "Max Failures", Category: CategoryPlugin, HotReloadable: true, Description: "Circuit breaker failure threshold"},
	{JSONKey: "plugin_reset_interval", Label: "Reset Interval", Category: CategoryPlugin, HotReloadable: true, Description: "Circuit breaker reset interval"},
	{JSONKey: "plugin_rate_limit", Label: "Rate Limit", Category: CategoryPlugin, HotReloadable: true, Description: "Plugin HTTP rate limit (req/sec/IP)"},
	{JSONKey: "plugin_max_routes", Label: "Max Routes", Category: CategoryPlugin, HotReloadable: true, Description: "Max HTTP routes per plugin"},
	{JSONKey: "plugin_max_request_body", Label: "Max Request Body", Category: CategoryPlugin, HotReloadable: true, Description: "Max request body size (bytes)"},
	{JSONKey: "plugin_max_response_body", Label: "Max Response Body", Category: CategoryPlugin, HotReloadable: true, Description: "Max response body size (bytes)"},

	// Update
	{JSONKey: "update_auto_enabled", Label: "Auto Update", Category: CategoryUpdate, HotReloadable: true, Description: "Enable automatic updates"},
	{JSONKey: "update_check_interval", Label: "Check Interval", Category: CategoryUpdate, HotReloadable: true, Description: "Update check interval (e.g. startup, 24h)"},
	{JSONKey: "update_channel", Label: "Channel", Category: CategoryUpdate, HotReloadable: true, Description: "Update channel (stable, beta)"},
	{JSONKey: "update_notify_only", Label: "Notify Only", Category: CategoryUpdate, HotReloadable: true, Description: "Only notify about updates, don't auto-install"},
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
