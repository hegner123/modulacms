# Configuration

All settings live in a single `modula.config.json` file, created automatically on first run with development defaults.

## File Location

ModulaCMS looks for `modula.config.json` in the working directory. Override the path with the `--config` flag:

```bash
modula serve --config /etc/modulacms/modula.config.json
```

## Config Layering (Multi-Environment)

When you maintain multiple environments (dev, staging, prod), you don't need to duplicate the entire config file. Create a small overlay file containing only the fields that differ, then merge it on top of the base config with `--overlay`:

```bash
modula serve --overlay modula.config.prod.json
```

The overlay file only needs the fields you want to override:

```json
{
  "environment": "prod",
  "db_driver": "postgres",
  "db_url": "db.example.com:5432",
  "db_password": "${POSTGRES_PASSWORD}",
  "port": ":8080",
  "bucket_secret_key": "${MINIO_ROOT_PASSWORD}"
}
```

At load time, overlay keys overwrite base keys. Fields absent from the overlay keep their base values. Maps and slices are replaced entirely (not deep-merged).

### Scaffold an Overlay

Generate a minimal overlay file for a new environment:

```bash
modula config overlay --env staging
# Creates modula.config.staging.json with {"environment": "staging"}
```

Then add only the fields that differ from your base config.

### Viewing Layered Config

```bash
# Show the merged result (base + overlay)
modula config show --overlay modula.config.prod.json

# Show the overlay file contents only
modula config show --overlay modula.config.prod.json --raw
```

### Updating Layered Config

When `--overlay` is set, `config set` writes to the overlay file by default (your intent is to override). Use `--base` to write to the base config instead:

```bash
# Write to overlay (default when layered)
modula config set db_password "newpass" --overlay modula.config.prod.json

# Write to base config
modula config set port ":8080" --overlay modula.config.prod.json --base
```

## Viewing and Validating

```bash
# Show current configuration (sensitive fields are redacted)
modula config show

# Validate configuration
modula config validate

# Machine-readable validation for CI scripts
modula config validate --json
# Output: {"valid": true, "errors": []}
```

Validation checks that required fields (`db_driver`, `db_url`, `port`, `ssh_port`) are present and that `db_driver` and `output_format` contain valid values.

## Project Registry

When you run `modula init`, the project is registered in `~/.modula/configs.json`. This lets you start any project's server or TUI from any directory:

```bash
modula serve mysite              # start mysite's default environment
modula serve mysite production   # start a specific environment
modula tui mysite                # launch TUI for mysite
```

Manage environments with `modula connect`:

```bash
modula connect set mysite staging /path/to/modula.config.staging.json
modula connect default mysite staging
modula connect list
```

> **Good to know**: The project registry and config overlays serve different purposes. Overlays merge fields within a single config file for per-environment deltas. The registry maps project names to entirely separate config files, potentially on different machines. Use overlays for small differences (different DB password), use the registry for separate projects or servers.

## Server Settings

| Field | Type | Default | Restart Required | Description |
|-------|------|---------|-----------------|-------------|
| `environment` | string | `"development"` | Yes | Runtime environment: `local`, `development`, `staging`, `production`, `docker`, `http-only` |
| `port` | string | `":8080"` | Yes | HTTP listen address |
| `ssl_port` | string | `":4000"` | Yes | HTTPS listen address |
| `ssh_host` | string | `"localhost"` | Yes | SSH server bind host |
| `ssh_port` | string | `"2233"` | Yes | SSH server port |
| `cert_dir` | string | `"./"` | Yes | Path to TLS certificate directory |
| `client_site` | string | `"localhost"` | No | Client site hostname |
| `admin_site` | string | `"localhost"` | No | Admin site hostname |
| `log_path` | string | `"./"` | No | Path for log files |
| `output_format` | string | `""` (raw) | No | Default API response format |
| `max_upload_size` | integer | `10485760` | No | Maximum file upload size in bytes (default 10 MB) |
| `node_id` | string | (auto-generated) | Yes | Unique node identifier (ULID). Auto-generated if empty. |

The `environment` field controls TLS behavior. `local` binds to `localhost` and uses `http://` for S3. `docker` binds to `0.0.0.0` and also uses `http://`. All other values use `https://` for S3 and enable Let's Encrypt autocert.

The `output_format` field sets the default response structure for content delivery endpoints. Valid values: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`. Defaults to `raw` when empty. See the [Routing guide](../building-content/routing.md) for details on output formats.

### Environment Hosts

The `environment_hosts` field maps environment names to hostnames for TLS certificate provisioning:

```json
{
  "environment_hosts": {
    "local": "localhost",
    "development": "localhost",
    "staging": "staging.example.com",
    "production": "cms.example.com"
  }
}
```

## Database Settings

| Field | Type | Default | Restart Required | Description |
|-------|------|---------|-----------------|-------------|
| `db_driver` | string | `"sqlite"` | Yes | Database driver: `sqlite`, `mysql`, `postgres` |
| `db_url` | string | `"./modula.db"` | Yes | Database connection URL or file path |
| `db_name` | string | `"modula.db"` | Yes | Database name |
| `db_username` | string | `""` | Yes | Database username (MySQL/PostgreSQL) |
| `db_password` | string | `""` | Yes | Database password (MySQL/PostgreSQL) |

SQLite uses a file path for `db_url`. MySQL and PostgreSQL use a `host:port` format.

```json
{
  "db_driver": "postgres",
  "db_url": "db.example.com:5432",
  "db_name": "modulacms",
  "db_username": "modula",
  "db_password": "your-password"
}
```

## S3 Storage Settings

ModulaCMS stores media assets and backups in S3-compatible storage. Any S3-compatible provider works: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, Cloudflare R2. See the [Media Management guide](../building-content/media.md) for upload and optimization details.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bucket_region` | string | `"us-east-1"` | S3 region |
| `bucket_media` | string | `""` | Bucket name for media assets |
| `bucket_backup` | string | `""` | Bucket name for backups |
| `bucket_endpoint` | string | `""` | S3 API endpoint hostname (without scheme) |
| `bucket_access_key` | string | `""` | S3 access key ID |
| `bucket_secret_key` | string | `""` | S3 secret access key |
| `bucket_public_url` | string | (falls back to endpoint) | Public-facing base URL for media links |
| `bucket_default_acl` | string | `""` | ACL applied to uploaded objects |
| `bucket_force_path_style` | bool | `true` | Use path-style URLs instead of virtual-hosted |
| `max_upload_size` | integer | `10485760` | Maximum upload size in bytes (10 MB) |

All S3 storage fields are hot-reloadable. You can change them without restarting the server.

Example for MinIO running locally:

```json
{
  "bucket_region": "us-east-1",
  "bucket_media": "media",
  "bucket_backup": "backups",
  "bucket_endpoint": "localhost:9000",
  "bucket_access_key": "minioadmin",
  "bucket_secret_key": "minioadmin",
  "bucket_public_url": "http://localhost:9000",
  "bucket_force_path_style": true
}
```

> **Good to know**: The `environment` field determines the S3 scheme. `http-only` and `docker` use `http://`; all other values use `https://`. Do not include the scheme in `bucket_endpoint`.

> **Good to know**: In Docker, `bucket_endpoint` typically points to a container hostname (e.g., `minio:9000`) that browsers cannot resolve. Set `bucket_public_url` to the externally reachable address so media URLs work in the browser.

## CORS Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `cors_origins` | string[] | `["http://localhost:3000"]` | Allowed origins |
| `cors_methods` | string[] | `["GET","POST","PUT","DELETE","OPTIONS"]` | Allowed HTTP methods |
| `cors_headers` | string[] | `["Content-Type","Authorization"]` | Allowed request headers |
| `cors_credentials` | bool | `true` | Allow credentials in CORS requests |

All CORS fields are hot-reloadable.

```json
{
  "cors_origins": ["https://app.example.com", "https://staging.example.com"],
  "cors_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
  "cors_headers": ["Content-Type", "Authorization"],
  "cors_credentials": true
}
```

## Cookie and Session Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `cookie_name` | string | `"modula_cms"` | Authentication cookie name |
| `cookie_duration` | string | `"1w"` | Cookie lifetime (e.g., `1w`, `24h`) |
| `cookie_secure` | bool | `false` | Set Secure flag on cookies (enable in production with HTTPS) |
| `cookie_samesite` | string | `"lax"` | SameSite policy: `strict`, `lax`, `none` |

All cookie fields are hot-reloadable.

## OAuth Settings

ModulaCMS supports OAuth with any OpenID Connect-compatible provider (Google, GitHub, Azure AD, etc.). Configure one provider per instance. See the [Authentication guide](../custom-admin/authentication.md) for setup instructions.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `oauth_client_id` | string | `""` | OAuth client ID |
| `oauth_client_secret` | string | `""` | OAuth client secret |
| `oauth_scopes` | string[] | `["openid","profile","email"]` | OAuth scopes to request |
| `oauth_provider_name` | string | `""` | Provider name (for display) |
| `oauth_redirect_url` | string | `""` | OAuth redirect callback URL |
| `oauth_success_redirect` | string | `"/"` | URL to redirect after successful login |
| `oauth_endpoint` | object | `{}` | Provider endpoint URLs (see below) |

The `oauth_endpoint` object requires three keys:

```json
{
  "oauth_endpoint": {
    "oauth_auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
    "oauth_token_url": "https://oauth2.googleapis.com/token",
    "oauth_userinfo_url": "https://openidconnect.googleapis.com/v1/userinfo"
  }
}
```

All OAuth fields are hot-reloadable. OAuth is optional -- the CMS works with local authentication when OAuth is not configured.

## Email Settings

ModulaCMS uses email for password reset flows. Four providers are supported: SMTP, SendGrid, AWS SES, and Postmark.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email_enabled` | bool | `false` | Enable email sending |
| `email_provider` | string | `""` | Provider: `smtp`, `sendgrid`, `ses`, `postmark` |
| `email_from_address` | string | `""` | Sender email address (required when enabled) |
| `email_from_name` | string | `""` | Sender display name |
| `email_reply_to` | string | `""` | Default reply-to address |
| `password_reset_url` | string | `""` | Base URL for password reset links |

### SMTP-specific Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email_host` | string | `""` | SMTP server hostname |
| `email_port` | integer | `587` | SMTP server port |
| `email_username` | string | `""` | SMTP auth username |
| `email_password` | string | `""` | SMTP auth password |
| `email_tls` | bool | `true` | Require TLS |

### API Provider Fields (SendGrid, Postmark, SES)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email_api_key` | string | `""` | API key (SendGrid, Postmark) |
| `email_api_endpoint` | string | `""` | Custom API endpoint URL |
| `email_aws_access_key_id` | string | `""` | AWS access key (SES only) |
| `email_aws_secret_access_key` | string | `""` | AWS secret key (SES only) |

All email fields are hot-reloadable.

> **Good to know**: When you configure SES without explicit AWS credentials, ModulaCMS falls back to the default AWS credential chain (environment variables, IAM role).

## Plugin System Settings

ModulaCMS has a Lua-based plugin system with sandboxed VMs and configurable resource limits. See the [Managing Plugins guide](../extending/overview.md) for setup and the [Plugin Tutorial](../extending/tutorial.md) for building your own.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_enabled` | bool | `false` | Enable the plugin system |
| `plugin_directory` | string | `"./plugins/"` | Path to plugins directory |
| `plugin_hot_reload` | bool | `false` | Enable file watcher for live plugin reload |
| `plugin_max_vms` | integer | `4` | Max Lua VMs per plugin |
| `plugin_timeout` | integer | `5` | Execution timeout in seconds |
| `plugin_max_ops` | integer | `1000` | Max operations per VM checkout |
| `plugin_max_failures` | integer | `5` | Circuit breaker failure threshold |
| `plugin_reset_interval` | string | `"60s"` | Circuit breaker reset interval |
| `plugin_rate_limit` | integer | `100` | HTTP rate limit per IP (req/sec) |
| `plugin_max_routes` | integer | `50` | Max HTTP routes per plugin |
| `plugin_max_request_body` | integer | `1048576` | Max request body (bytes, default 1 MB) |
| `plugin_max_response_body` | integer | `5242880` | Max response body (bytes, default 5 MB) |

`plugin_enabled`, `plugin_directory`, and `plugin_hot_reload` require a server restart. All other plugin fields are hot-reloadable.

## Observability Settings

ModulaCMS supports external error tracking and metrics through Sentry, Datadog, and New Relic.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `observability_enabled` | bool | `false` | Enable observability |
| `observability_provider` | string | `"console"` | Provider: `sentry`, `datadog`, `newrelic`, `console` |
| `observability_dsn` | string | `""` | Connection string / DSN |
| `observability_environment` | string | `"development"` | Environment label |
| `observability_release` | string | `""` | Version/release identifier |
| `observability_sample_rate` | float | `1.0` | Event sample rate (0.0 to 1.0) |
| `observability_traces_rate` | float | `0.1` | Trace sample rate (0.0 to 1.0) |
| `observability_send_pii` | bool | `false` | Send personally identifiable info |
| `observability_debug` | bool | `false` | Enable debug logging for the observability client |
| `observability_server_name` | string | `""` | Server/instance name |
| `observability_flush_interval` | string | `"30s"` | How often to flush metrics |
| `observability_tags` | object | `{}` | Global key-value tags for all events |

All observability fields are hot-reloadable.

## Admin S3 Storage Settings

When admin media needs a separate bucket or S3 endpoint from public media, configure these fields. If left empty, they fall back to the shared S3 settings above.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bucket_admin_media` | string | (falls back to `bucket_media`) | Bucket name for admin media assets |
| `bucket_admin_endpoint` | string | (falls back to `bucket_endpoint`) | S3 endpoint for admin media |
| `bucket_admin_access_key` | string | (falls back to `bucket_access_key`) | S3 access key for admin media |
| `bucket_admin_secret_key` | string | (falls back to `bucket_secret_key`) | S3 secret key for admin media |
| `bucket_admin_public_url` | string | (falls back to `bucket_public_url`) | Public URL for admin media links |

All admin bucket fields are hot-reloadable.

## Remote Connection Settings

Used by the `connect` command to operate against a remote ModulaCMS instance over HTTPS instead of a local database.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `remote_url` | string | `""` | Base URL of the remote ModulaCMS instance |
| `remote_api_key` | string | `""` | API key for authenticating with the remote instance |

These fields are mutually exclusive with `db_driver` when using the `connect` command.

## Deploy Settings

Configure remote environments for content sync operations.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `deploy_environments` | array | `[]` | List of remote environments (each with `name`, `url`, `api_key`) |
| `deploy_snapshot_dir` | string | `""` | Local directory for deploy snapshots |

Each entry in `deploy_environments` is an object:

```json
{
  "deploy_environments": [
    { "name": "staging", "url": "https://staging.example.com", "api_key": "${STAGING_API_KEY}" },
    { "name": "production", "url": "https://cms.example.com", "api_key": "${PROD_API_KEY}" }
  ]
}
```

## Publishing Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `publish_schedule_interval` | integer | `60` | Seconds between scheduled publish ticks |
| `version_max_per_content` | integer | `50` | Maximum versions per content item (0 = unlimited) |
| `node_level_publish` | bool | `false` | When false, publish propagates to all descendants; when true, publish is per-node |

## Webhook Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `webhook_enabled` | bool | `false` | Enable webhook delivery |
| `webhook_timeout` | integer | `10` | HTTP timeout in seconds for deliveries |
| `webhook_max_retries` | integer | `3` | Maximum retry attempts per delivery |
| `webhook_workers` | integer | `4` | Concurrent delivery workers |
| `webhook_allow_http` | bool | `false` | Allow non-TLS webhook URLs (dev only) |
| `webhook_delivery_retention_days` | integer | `30` | Days to retain completed deliveries |

## Internationalization Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `i18n_enabled` | bool | `false` | Enable internationalization |
| `i18n_default_locale` | string | `"en"` | Default locale code |

## Search Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `search_enabled` | bool | `false` | Enable full-text search |
| `search_path` | string | `""` | Path for the search index directory |

## MCP Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mcp_enabled` | bool | `false` | Enable the MCP server (Model Context Protocol for AI tooling) |
| `mcp_api_key` | string | `""` | API key for authenticating MCP clients |

## Composition Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `composition_max_depth` | integer | `10` | Maximum depth for reference datatype composition |

## Richtext Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `richtext_toolbar` | string[] | `["bold","italic","h1","h2","h3","link","ul","ol","preview"]` | Default toolbar buttons for richtext fields |

## Extended Plugin Settings

In addition to the base plugin fields above, these control database pooling, content hooks, outbound requests, and production hardening.

### Plugin Database Pool

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_db_max_open_conns` | integer | (driver default) | Max open database connections for plugins |
| `plugin_db_max_idle_conns` | integer | (driver default) | Max idle database connections for plugins |
| `plugin_db_conn_max_lifetime` | string | (driver default) | Max connection lifetime (e.g., `"5m"`) |

### Plugin Hook Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_hook_reserve_vms` | integer | `1` | VMs reserved for hooks per plugin |
| `plugin_hook_max_consecutive_aborts` | integer | `10` | Circuit breaker threshold for hook failures |
| `plugin_hook_max_ops` | integer | `100` | Reduced op budget for after-hooks |
| `plugin_hook_max_concurrent_after` | integer | `10` | Max concurrent after-hook goroutines |
| `plugin_hook_timeout_ms` | integer | `2000` | Per-hook timeout in before-hooks (ms) |
| `plugin_hook_event_timeout_ms` | integer | `5000` | Total timeout for before-hook chain per event (ms) |

### Plugin Outbound Request Engine

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_request_timeout` | integer | `10` | Outbound request timeout in seconds |
| `plugin_request_max_response` | integer | `1048576` | Max response body size in bytes (1 MB) |
| `plugin_request_max_body` | integer | `1048576` | Max request body size in bytes (1 MB) |
| `plugin_request_rate_limit` | integer | `60` | Requests per plugin per domain per minute |
| `plugin_request_global_rate` | integer | `600` | Aggregate requests per minute (0 = unlimited) |
| `plugin_request_cb_failures` | integer | `5` | Consecutive failures to trip circuit breaker |
| `plugin_request_cb_reset` | integer | `60` | Seconds before half-open probe |
| `plugin_request_allow_local` | bool | `false` | Allow HTTP to localhost (dev only) |

### Plugin Production Hardening

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_sync_interval` | string | `"10s"` | DB state polling interval for multi-instance sync (`"0"` disables) |
| `plugin_trusted_proxies` | string[] | `[]` | CIDR list for trusted proxies (empty = use RemoteAddr only) |

## Update Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `update_auto_enabled` | bool | `false` | Enable automatic updates |
| `update_check_interval` | string | `"startup"` | Check frequency: `startup`, or a duration like `24h` |
| `update_channel` | string | `"stable"` | Update channel: `stable`, `beta` |
| `update_notify_only` | bool | `false` | Only notify about updates, do not auto-install |

## Backup Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `backup_option` | string | `"./"` | Local backup storage directory |
| `backup_paths` | string[] | `[""]` | Additional backup paths |

When `bucket_backup` is configured, backups can also be stored in S3.

## Hot Reloading

You can change many configuration fields at runtime without restarting the server. The "Restart Required" column in each section indicates which fields need a restart.

Fields that require a restart:
- Server ports and bind addresses (`port`, `ssl_port`, `ssh_host`, `ssh_port`)
- Environment and certificate settings (`environment`, `cert_dir`)
- Database connection settings (`db_driver`, `db_url`, `db_name`, `db_username`, `db_password`)
- Authentication salt (`auth_salt`)
- Plugin enable/disable and directory (`plugin_enabled`, `plugin_directory`, `plugin_hot_reload`)
- Node ID (`node_id`)

Fields that are hot-reloadable (take effect immediately):
- CORS settings
- Cookie settings
- OAuth settings
- S3 storage settings (including admin bucket fields)
- Email settings
- Observability settings
- Plugin runtime limits (VMs, timeout, rate limits, hooks, requests)
- Update settings
- Output format and upload size
- Webhook settings
- Search settings
- Internationalization settings
- Composition settings
- Richtext toolbar
- Publishing settings
- MCP settings

Update hot-reloadable fields through the admin panel at `/admin/settings`, through the REST API, or by editing `modula.config.json` directly. Changes made via the API or admin panel save to `modula.config.json` and take effect immediately. Changes made by editing the file directly require the config to be reloaded.

### API-Based Configuration

Read the current configuration (sensitive fields are redacted):

```bash
curl http://localhost:8080/api/v1/admin/config \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Update configuration fields:

```bash
curl -X PATCH http://localhost:8080/api/v1/admin/config \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "cors_origins": ["https://app.example.com"],
    "max_upload_size": 20971520
  }'
```

The API validates changes before applying them. If a change affects a restart-required field, the response includes a warning listing the fields that need a server restart.

The config meta endpoint returns field metadata (categories, descriptions, hot-reloadable status):

```bash
curl http://localhost:8080/api/v1/admin/config/meta \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Both config endpoints require the `config:read` or `config:update` permission.

## Sensitive Fields

ModulaCMS redacts the following fields (replacing them with `********`) when returning configuration through the API or the `config show` command:

- `auth_salt`
- `db_password`
- `bucket_access_key`
- `bucket_secret_key`
- `bucket_admin_access_key`
- `bucket_admin_secret_key`
- `remote_api_key`
- `oauth_client_id`
- `oauth_client_secret`
- `observability_dsn`
- `email_password`
- `email_api_key`
- `email_aws_access_key_id`
- `email_aws_secret_access_key`
- `mcp_api_key`

> **Good to know**: When you update configuration via the API, redacted values (`********`) are skipped automatically to prevent overwriting secrets with the placeholder.

## Full Default Configuration

`modula init` creates a `modula.config.json` with these defaults:

```json
{
  "environment": "development",
  "port": ":8080",
  "ssl_port": ":4000",
  "ssh_host": "localhost",
  "ssh_port": "2233",
  "cert_dir": "./",
  "client_site": "localhost",
  "admin_site": "localhost",
  "db_driver": "sqlite",
  "db_url": "./modula.db",
  "db_name": "modula.db",
  "cookie_name": "modula_cms",
  "cookie_duration": "1w",
  "cookie_secure": false,
  "cookie_samesite": "lax",
  "cors_origins": ["http://localhost:3000"],
  "cors_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
  "cors_headers": ["Content-Type", "Authorization"],
  "cors_credentials": true,
  "bucket_region": "us-east-1",
  "bucket_force_path_style": true,
  "max_upload_size": 10485760,
  "output_format": "",
  "observability_enabled": false,
  "observability_provider": "console",
  "observability_sample_rate": 1.0,
  "observability_traces_rate": 0.1,
  "email_enabled": false,
  "email_port": 587,
  "email_tls": true,
  "plugin_enabled": false,
  "update_auto_enabled": false,
  "update_check_interval": "startup",
  "update_channel": "stable"
}
```

Fields not shown here (database credentials, S3 credentials, OAuth endpoints, email credentials, admin bucket settings, remote connection, deploy environments, MCP, search, webhook, i18n, richtext, composition) default to empty strings, empty arrays, zero values, or false and need to be set for their respective features.

## Next steps

- [Authentication guide](../custom-admin/authentication.md) -- set up OAuth, sessions, and API tokens
- [Media Management guide](../building-content/media.md) -- configure S3 storage and upload files
- [Managing Plugins guide](../extending/overview.md) -- enable and configure the plugin system
