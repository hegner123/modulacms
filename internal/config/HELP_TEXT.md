# Configuration Help Text

Help text for each configuration field, organized by category. These descriptions are intended for admin panel tooltips, settings page help icons, and configuration documentation.

## Server

### `environment`
The runtime environment determines TLS behavior and logging verbosity. Use `development` for local work (verbose logs, no TLS requirement), `staging` for pre-production, `production` for live deployments (strict TLS, reduced logging), `http-only` to explicitly disable TLS, or `docker` for containerized environments that terminate TLS at a load balancer.

**Default:** `development`

### `os`
The operating system the server is running on. Set automatically at startup. Used internally for OS-specific behavior like file paths and signal handling.

**Default:** detected at runtime

### `environment_hosts`
A map of environment names to hostnames. Used to resolve the correct hostname for the current environment when constructing URLs (e.g., autocert domains). Each key is an environment name (`local`, `development`, `staging`, `production`, `http-only`) and each value is its hostname.

**Default:** all set to `localhost`

### `port`
The address and port the HTTP server listens on. Include the colon prefix (e.g., `:8080`). To bind to a specific interface, include the IP (e.g., `127.0.0.1:8080`). This is the primary API and admin panel port.

**Default:** `:8080`

### `ssl_port`
The address and port the HTTPS server listens on. Uses Let's Encrypt autocert for automatic TLS certificates. Set to empty to disable the HTTPS listener. Requires a valid domain in `environment_hosts` for certificate issuance.

**Default:** `:4000`

### `cert_dir`
Directory where TLS certificates are stored. Used by the autocert manager to cache Let's Encrypt certificates. Must be writable by the server process.

**Default:** `./`

### `client_site`
The hostname of the frontend client site. Used in CORS configuration and link generation. Set this to your public-facing domain.

**Default:** `localhost`

### `admin_site`
The hostname of the admin panel. Used for admin-specific CORS configuration and generating admin URLs in emails or notifications.

**Default:** `localhost`

### `ssh_host`
The address the SSH server binds to. Set to `0.0.0.0` to accept connections on all interfaces, or `127.0.0.1` to restrict to localhost. The SSH server provides the Bubbletea TUI.

**Default:** `localhost`

### `ssh_port`
The port the SSH server listens on. Users connect to the TUI management interface via `ssh -p <port> user@host`.

**Default:** `2233`

### `options`
A generic key-value map for custom options. Reserved for future use and plugin configuration. Values are arrays of any type.

### `log_path`
Directory where log files are written. The server writes structured JSON logs here. Set to `./` for the current directory or provide an absolute path. Ensure the directory exists and is writable.

**Default:** `./`

### `auth_salt`
A secret string used as salt for password hashing. Generated automatically on first setup. Changing this value after users have been created will invalidate all existing passwords. Store securely and never expose publicly.

**Default:** auto-generated from timestamp (change in production)

### `cookie_name`
The name of the HTTP session cookie. Change this if running multiple ModulaCMS instances on the same domain to prevent cookie collisions.

**Default:** `modula_cms`

### `cookie_duration`
How long the session cookie remains valid. Accepts duration strings: `1w` (one week), `24h` (24 hours), `30m` (30 minutes). After expiration, users must log in again.

**Default:** `1w`

### `cookie_secure`
When enabled, the session cookie is only sent over HTTPS connections. Must be `true` in production with TLS. Set to `false` only for local development over HTTP.

**Default:** `false`

### `cookie_samesite`
Controls when the browser sends the session cookie in cross-site requests. `strict` blocks all cross-site cookie sending (most secure). `lax` allows cookies on top-level navigations (recommended). `none` allows all cross-site sending (requires `cookie_secure: true`).

**Default:** `lax`

### `output_format`
The default content API response format. Determines how content trees are serialized when no `?format=` query parameter is provided. Options: `contentful`, `sanity`, `strapi`, `wordpress`, `clean` (metadata stripped), `raw` (unprocessed).

**Default:** empty (defaults to `raw`)

### `custom_style_path`
Path to a custom TUI style JSON file. Overrides the default color scheme for the SSH terminal interface. Leave empty to use the built-in theme.

**Default:** empty (built-in theme)

### `max_upload_size`
Maximum file upload size in bytes. Applies to media uploads via the API and admin panel. Set to `0` to use the default. Common values: `10485760` (10 MB), `52428800` (50 MB), `104857600` (100 MB).

**Default:** `10485760` (10 MB)

### `node_id`
A unique ULID identifying this server instance. Generated automatically on first setup. Used in multi-instance deployments to distinguish which node produced an audit event or held a lock.

**Default:** auto-generated ULID

### `space_id`
An optional identifier for this CMS installation. Useful for multi-tenant setups or when managing multiple independent CMS instances.

**Default:** empty

## Database

### `db_driver`
The database backend to use. `sqlite` stores data in a local file (simplest, no external dependencies). `mysql` connects to a MySQL server. `postgres` connects to a PostgreSQL server. Cannot be changed after initial setup without migrating data.

**Default:** `sqlite`

### `db_url`
The database connection string. For SQLite: a file path (e.g., `./modula.db`). For MySQL: `user:password@tcp(host:3306)/dbname`. For PostgreSQL: `postgres://user:password@host:5432/dbname?sslmode=disable`.

**Default:** `./modula.db`

### `db_name`
The database name. Used by MySQL and PostgreSQL drivers. Ignored by SQLite (the file path in `db_url` is used instead).

**Default:** `modula`

### `db_username`
The database username for MySQL or PostgreSQL connections. Not used with SQLite.

**Default:** empty

### `db_password`
The database password for MySQL or PostgreSQL connections. Not used with SQLite. Supports `${ENV_VAR}` syntax to reference environment variables instead of storing secrets in the config file.

**Default:** empty

## Remote

### `remote_url`
Base URL of a remote ModulaCMS instance. Used by the `connect` command to operate on a remote CMS via its API instead of a local database. Mutually exclusive with `db_driver` — set one or the other, not both.

**Default:** empty

### `remote_api_key`
API key for authenticating with the remote ModulaCMS instance. Required when `remote_url` is set. Supports `${ENV_VAR}` syntax.

**Default:** empty

## Storage (S3)

### `bucket_region`
The AWS region (or S3-compatible region) for your storage buckets. Must match the region where your buckets are created.

**Default:** `us-east-1`

### `bucket_media`
The name of the S3 bucket used for media file storage. All uploaded images, documents, and files are stored here. Must be created before use.

**Default:** empty

### `bucket_backup`
The name of the S3 bucket used for backup storage. Database dumps and site backups are stored here when `backup_option` is set to S3.

**Default:** empty

### `bucket_endpoint`
The S3-compatible API endpoint hostname, without the `https://` scheme prefix. The scheme is determined automatically from the `environment` setting. For AWS: `s3.amazonaws.com`. For MinIO: `minio:9000` or `localhost:9000`.

**Default:** empty

### `bucket_access_key`
The S3 access key ID for authenticating with the storage provider. Supports `${ENV_VAR}` syntax.

**Default:** empty

### `bucket_secret_key`
The S3 secret access key for authenticating with the storage provider. Supports `${ENV_VAR}` syntax. Never commit this value to version control.

**Default:** empty

### `bucket_public_url`
The public-facing base URL for media assets. Used to construct URLs that browsers can access. In Docker environments, `bucket_endpoint` may be an internal hostname (e.g., `minio:9000`) that browsers cannot resolve — set this to the externally reachable address (e.g., `http://localhost:9000`). If empty, falls back to the endpoint URL.

**Default:** empty (falls back to `bucket_endpoint`)

### `bucket_default_acl`
The default ACL applied to uploaded objects. Common values: `public-read` (objects are publicly accessible), `private` (objects require signed URLs). Depends on your storage provider's ACL support.

**Default:** empty

### `bucket_force_path_style`
Use path-style URLs for S3 requests (`endpoint/bucket/key`) instead of virtual-hosted-style (`bucket.endpoint/key`). Enable this for S3-compatible providers like MinIO, Wasabi, or local development.

**Default:** `true`

### `backup_option`
Where to store backups. Set to a local directory path (e.g., `./backups`) for local storage, or `s3` to store in the `bucket_backup` bucket.

**Default:** `./`

### `backup_paths`
Additional file paths to include in backups. These paths are archived alongside the database dump when creating backups.

**Default:** empty

## CORS

### `cors_origins`
The list of origins allowed to make cross-origin requests to the API. Each entry should be a full origin including scheme (e.g., `https://example.com`). Set to `*` to allow all origins (not recommended for production). Separate multiple origins with commas in the config file.

**Default:** `["http://localhost:3000"]`

### `cors_methods`
The HTTP methods allowed in cross-origin requests. Typically includes all methods your frontend uses.

**Default:** `["GET", "POST", "PUT", "DELETE", "OPTIONS"]`

### `cors_headers`
The request headers allowed in cross-origin requests. Include any custom headers your frontend sends (e.g., `X-CSRF-Token` for the admin panel).

**Default:** `["Content-Type", "Authorization"]`

### `cors_credentials`
Whether to allow credentials (cookies, authorization headers) in cross-origin requests. Must be `true` if your frontend uses cookie-based authentication across origins.

**Default:** `true`

## Cookie

### `cookie_name`
See [Server > cookie_name](#cookie_name).

### `cookie_duration`
See [Server > cookie_duration](#cookie_duration).

### `cookie_secure`
See [Server > cookie_secure](#cookie_secure).

### `cookie_samesite`
See [Server > cookie_samesite](#cookie_samesite).

## OAuth

### `oauth_client_id`
The OAuth client ID from your identity provider (Google, GitHub, Azure, etc.). Obtained when registering your application with the provider.

**Default:** empty (OAuth disabled)

### `oauth_client_secret`
The OAuth client secret from your identity provider. Keep this value secret. Supports `${ENV_VAR}` syntax.

**Default:** empty

### `oauth_scopes`
The OAuth scopes to request during authentication. Most providers require at least `openid`, `profile`, and `email` for basic user information.

**Default:** `["openid", "profile", "email"]`

### `oauth_endpoint`
A map of OAuth endpoint URLs keyed by type: `oauth_auth_url` (authorization endpoint), `oauth_token_url` (token exchange endpoint), `oauth_userinfo_url` (user info endpoint). These vary by provider. Consult your provider's documentation for the correct URLs.

**Default:** all empty

### `oauth_provider_name`
A human-readable name for the OAuth provider, used in the login UI (e.g., "Google", "GitHub", "Azure AD").

**Default:** empty

### `oauth_redirect_url`
The callback URL that the OAuth provider redirects to after authentication. Must match exactly what you registered with your provider. Typically `https://yourdomain.com/api/v1/auth/oauth/callback`.

**Default:** empty

### `oauth_success_redirect`
The URL to redirect to after successful OAuth authentication. Typically your admin panel dashboard or frontend home page.

**Default:** `/`

## Observability

### `observability_enabled`
Enable the observability integration for error tracking and performance monitoring.

**Default:** `false`

### `observability_provider`
The observability backend to send data to. Options: `sentry`, `datadog`, `newrelic`, `console` (local logging only).

**Default:** `console`

### `observability_dsn`
The connection string for your observability provider (e.g., a Sentry DSN). Format varies by provider.

**Default:** empty

### `observability_environment`
The environment label sent with observability events. Helps filter events in your monitoring dashboard by environment.

**Default:** `development`

### `observability_release`
The version or release identifier sent with events. Typically set to the application version (e.g., `v1.2.3`). Helps correlate errors with specific releases.

**Default:** empty

### `observability_sample_rate`
The percentage of error events to send, from `0.0` (none) to `1.0` (all). Lower values reduce costs and noise in high-traffic environments while still catching representative errors.

**Default:** `1.0`

### `observability_traces_rate`
The percentage of performance traces to capture, from `0.0` (none) to `1.0` (all). Traces measure request latency and database query performance. Start low (0.1) in production and increase if needed.

**Default:** `0.1`

### `observability_send_pii`
Whether to include personally identifiable information (usernames, emails, IP addresses) in observability events. Disable in production environments with strict privacy requirements.

**Default:** `false`

### `observability_debug`
Enable debug logging for the observability client itself. Useful for troubleshooting connectivity issues with your monitoring provider.

**Default:** `false`

### `observability_server_name`
A name identifying this server instance in monitoring dashboards. Useful for distinguishing between multiple CMS instances.

**Default:** empty

### `observability_flush_interval`
How often to flush accumulated metrics and events to the provider. Accepts duration strings (e.g., `30s`, `1m`, `5m`).

**Default:** `30s`

### `observability_tags`
A map of key-value pairs attached to all events and metrics as global tags. Useful for adding custom metadata like team name, datacenter, or deployment slot.

**Default:** empty

## Email

### `email_enabled`
Enable transactional email sending. Required for password reset flows and email notifications. When disabled, password reset requests will fail.

**Default:** `false`

### `email_provider`
The email sending backend. Options: `smtp` (standard SMTP relay), `sendgrid` (SendGrid HTTP API), `ses` (AWS SES), `postmark` (Postmark HTTP API). Each provider requires its own set of credentials below.

**Default:** empty (disabled)

### `email_from_address`
The sender email address for all outgoing messages (e.g., `noreply@example.com`). Must be a verified address with your email provider.

**Default:** empty

### `email_from_name`
The sender display name shown in email clients alongside the from address (e.g., `ModulaCMS`).

**Default:** empty

### `email_host`
The SMTP server hostname. Only used when `email_provider` is `smtp`. Common values: `smtp.gmail.com`, `smtp.sendgrid.net`, `email-smtp.us-east-1.amazonaws.com`.

**Default:** empty

### `email_port`
The SMTP server port. Only used when `email_provider` is `smtp`. Common values: `587` (STARTTLS, recommended), `465` (implicit TLS), `25` (unencrypted, not recommended).

**Default:** `587`

### `email_username`
The SMTP authentication username. Only used when `email_provider` is `smtp`.

**Default:** empty

### `email_password`
The SMTP authentication password. Only used when `email_provider` is `smtp`. Supports `${ENV_VAR}` syntax.

**Default:** empty

### `email_tls`
Require TLS encryption for SMTP connections. Should be `true` for production use. Disable only for local development SMTP servers without TLS.

**Default:** `true`

### `email_api_key`
The API key for HTTP-based email providers (SendGrid, Postmark). Not used with SMTP or SES. Supports `${ENV_VAR}` syntax.

**Default:** empty

### `email_api_endpoint`
Custom API endpoint URL for email providers. Only set this if your provider uses a non-standard endpoint. Leave empty to use the provider's default.

**Default:** empty

### `email_reply_to`
The default reply-to address for outgoing emails. When recipients reply, their message goes to this address instead of the from address.

**Default:** empty

### `email_aws_access_key_id`
AWS access key ID for SES email sending. Only used when `email_provider` is `ses`. Supports `${ENV_VAR}` syntax.

**Default:** empty

### `email_aws_secret_access_key`
AWS secret access key for SES email sending. Only used when `email_provider` is `ses`. Supports `${ENV_VAR}` syntax.

**Default:** empty

### `password_reset_url`
The URL template for password reset links sent in emails. The server appends a `?token=` parameter. Should point to your admin panel's reset password page (e.g., `https://admin.example.com/reset-password`).

**Default:** empty

## Plugin

### `plugin_enabled`
Enable the Lua plugin system. When disabled, no plugins are loaded or executed. Requires a server restart to take effect.

**Default:** `false`

### `plugin_directory`
The directory containing plugin folders. Each subdirectory should contain a `manifest.json` and Lua source files. Relative paths are resolved from the server's working directory.

**Default:** `./plugins`

### `plugin_max_vms`
Maximum number of Lua virtual machines pooled per plugin. Higher values allow more concurrent plugin executions but use more memory. Each VM is approximately 1-2 MB.

**Default:** `4`

### `plugin_timeout`
Maximum execution time in seconds for a single plugin invocation. Plugins exceeding this limit are terminated. Protects against infinite loops in plugin code.

**Default:** `5`

### `plugin_max_ops`
Maximum number of Lua operations (instructions) per VM checkout. Acts as a CPU budget preventing runaway plugins from consuming excessive resources.

**Default:** `1000`

### `plugin_hot_reload`
Enable automatic plugin reloading when files change on disk. Uses a file watcher to detect changes and reload affected plugins without server restart. Recommended only for development.

**Default:** `false`

### `plugin_max_failures`
Number of consecutive failures before the circuit breaker trips for a plugin. A tripped circuit breaker stops all requests to the failing plugin until the reset interval elapses.

**Default:** `5`

### `plugin_reset_interval`
Duration before a tripped circuit breaker transitions to half-open state and allows a probe request. Accepts duration strings (e.g., `60s`, `5m`).

**Default:** `60s`

### `plugin_sync_interval`
How often to poll the database for plugin state changes in multi-instance deployments. Set to `0` to disable. Accepts duration strings.

**Default:** `10s`

### `plugin_rate_limit`
Maximum HTTP requests per second per IP address to plugin-served routes. Protects plugin endpoints from abuse.

**Default:** `100`

### `plugin_max_routes`
Maximum number of HTTP routes a single plugin can register. Prevents a single plugin from consuming the entire URL namespace.

**Default:** `50`

### `plugin_max_request_body`
Maximum request body size in bytes for plugin HTTP endpoints. Limits memory consumption from large payloads.

**Default:** `1048576` (1 MB)

### `plugin_max_response_body`
Maximum response body size in bytes that a plugin can return. Prevents plugins from sending excessively large responses.

**Default:** `5242880` (5 MB)

### `plugin_trusted_proxies`
List of trusted proxy CIDR ranges for extracting real client IPs from `X-Forwarded-For` headers. Empty means only `RemoteAddr` is used. Example: `["10.0.0.0/8", "172.16.0.0/12"]`.

**Default:** empty (use RemoteAddr only)

### `plugin_db_max_open_conns`
Maximum open database connections in the plugin connection pool. Zero uses the default from the database driver.

**Default:** `0` (driver default)

### `plugin_db_max_idle_conns`
Maximum idle database connections in the plugin connection pool. Zero uses the default.

**Default:** `0` (driver default)

### `plugin_db_conn_max_lifetime`
Maximum lifetime of a database connection in the plugin pool before it is closed and replaced. Accepts duration strings (e.g., `30m`, `1h`).

**Default:** empty (no limit)

### `plugin_hook_reserve_vms`
Number of Lua VMs reserved exclusively for content hooks per plugin. Ensures hooks can execute even when the general VM pool is exhausted.

**Default:** `1`

### `plugin_hook_max_consecutive_aborts`
Number of consecutive hook aborts before the hook circuit breaker trips, disabling that hook chain for the plugin.

**Default:** `10`

### `plugin_hook_max_ops`
Reduced Lua operation budget for after-hooks. Lower than `plugin_max_ops` because after-hooks run asynchronously and should be lightweight.

**Default:** `100`

### `plugin_hook_max_concurrent_after`
Maximum number of after-hook goroutines running concurrently across all plugins. Prevents hook storms from consuming all server resources.

**Default:** `10`

### `plugin_hook_timeout_ms`
Per-hook timeout in milliseconds for before-hooks. Before-hooks run synchronously in the request path, so timeouts must be short.

**Default:** `2000` (2 seconds)

### `plugin_hook_event_timeout_ms`
Total timeout in milliseconds for the entire before-hook chain on a single event. If multiple plugins have before-hooks, this is the maximum time for all of them combined.

**Default:** `5000` (5 seconds)

### `plugin_request_timeout`
Timeout in seconds for outbound HTTP requests made by plugins via the request API.

**Default:** `10`

### `plugin_request_max_response`
Maximum response body size in bytes that plugins can receive from outbound HTTP requests.

**Default:** `1048576` (1 MB)

### `plugin_request_max_body`
Maximum request body size in bytes that plugins can send in outbound HTTP requests.

**Default:** `1048576` (1 MB)

### `plugin_request_rate_limit`
Maximum outbound HTTP requests per plugin per domain per minute. Prevents plugins from overwhelming external services.

**Default:** `60`

### `plugin_request_global_rate`
Aggregate maximum outbound HTTP requests per minute across all plugins. Set to `0` for unlimited.

**Default:** `600`

### `plugin_request_cb_failures`
Number of consecutive failures to an external domain before the outbound request circuit breaker trips.

**Default:** `5`

### `plugin_request_cb_reset`
Seconds before a tripped outbound request circuit breaker allows a probe request.

**Default:** `60`

### `plugin_request_allow_local`
Allow plugins to make HTTP requests to localhost addresses. Enable only for development. In production, this should be `false` to prevent SSRF attacks.

**Default:** `false`

## Deploy

### `deploy_environments`
A list of remote ModulaCMS instances for content synchronization. Each entry has a `name`, `url`, and `api_key`. Used by the deploy push/pull commands to transfer content between environments. API keys support `${ENV_VAR}` syntax.

**Default:** empty

### `deploy_snapshot_dir`
Local directory for storing deploy snapshots. Snapshots are JSON exports of content trees used during push/pull operations.

**Default:** `./deploy/snapshots`

## Publishing

### `publish_schedule_interval`
Seconds between scheduler ticks for processing scheduled publications. Lower values mean scheduled content goes live faster but increase CPU usage. The scheduler checks for content whose scheduled publication time has passed.

**Default:** `60`

### `version_max_per_content`
Maximum number of version snapshots retained per content item. When a new version is created and the limit is reached, the oldest version is deleted. Set to `0` for unlimited retention (not recommended for large sites).

**Default:** `50`

### `composition_max_depth`
Maximum depth for recursive content tree composition. Limits how deeply reference datatypes are resolved to prevent infinite loops from circular references.

**Default:** `10`

## Internationalization

### `i18n_enabled`
Enable internationalization support. When enabled, content can have locale-specific field values and the API accepts `?locale=` parameters.

**Default:** `false`

### `i18n_default_locale`
The default locale code used when no locale is specified in API requests. Must be a valid BCP 47 language tag (e.g., `en`, `fr`, `de`, `ja`).

**Default:** `en`

## Webhooks

### `webhook_enabled`
Enable the webhook system for sending event notifications to external URLs when content is published, updated, or deleted.

**Default:** `false`

### `webhook_timeout`
HTTP timeout in seconds for individual webhook delivery attempts. If the target server doesn't respond within this time, the delivery is marked as failed and retried.

**Default:** `10`

### `webhook_max_retries`
Maximum number of retry attempts for failed webhook deliveries. Retries use exponential backoff. Set to `0` to disable retries.

**Default:** `3`

### `webhook_workers`
Number of concurrent webhook delivery workers. Higher values allow more simultaneous deliveries but use more resources.

**Default:** `4`

### `webhook_allow_http`
Allow webhook URLs with `http://` (non-TLS). Enable only for development. In production, all webhook targets should use HTTPS.

**Default:** `false`

### `webhook_delivery_retention_days`
Number of days to retain completed webhook delivery records. Older records are automatically cleaned up. Set to a higher value if you need longer delivery audit trails.

**Default:** `30`

## Search

### `search_enabled`
Enable the built-in full-text search index. When enabled, published content is automatically indexed and searchable via the `/api/v1/search` endpoint.

**Default:** `false`

### `search_path`
File path where the search index is persisted to disk. The index is saved periodically and on shutdown, then loaded on startup. Relative paths are resolved from the server's working directory. Use an absolute path for predictable behavior. Must not be empty when search is enabled.

**Default:** `search.idx`

## MCP

### `mcp_enabled`
Enable the Model Context Protocol server for AI tool integration. When enabled, AI assistants can manage CMS content through the MCP protocol.

**Default:** `false`

### `mcp_api_key`
API key for authenticating MCP client connections. Required when `mcp_enabled` is `true`. Supports `${ENV_VAR}` syntax.

**Default:** empty

## Update

### `update_auto_enabled`
Enable automatic update checking. When enabled, the server periodically checks for new versions according to the `update_check_interval`.

**Default:** `false`

### `update_check_interval`
How often to check for updates. Set to `startup` to check only on server start, or a duration string (e.g., `24h`, `1w`) for periodic checks.

**Default:** `startup`

### `update_channel`
The release channel to check for updates. `stable` receives only production-ready releases. `beta` includes pre-release versions.

**Default:** `stable`

### `update_notify_only`
When enabled, the server only logs a notification about available updates instead of automatically downloading them.

**Default:** `false`

## Misc

### `richtext_toolbar`
The default set of toolbar buttons for richtext editor fields in the admin panel. Controls which formatting options are available to content editors. Options include: `bold`, `italic`, `underline`, `strikethrough`, `h1`, `h2`, `h3`, `h4`, `link`, `image`, `ul`, `ol`, `blockquote`, `code`, `codeblock`, `hr`, `preview`.

**Default:** `["bold", "italic", "h1", "h2", "h3", "link", "ul", "ol", "preview"]`

### `keybindings`
Custom key bindings for the SSH TUI interface. Maps action names to arrays of key strings. Any action not specified uses its default binding. See the TUI documentation for the full list of available actions.

**Default:** built-in key map
