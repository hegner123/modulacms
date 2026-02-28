# Configuration

ModulaCMS is configured through a single `config.json` file. The file is created automatically on first run with development defaults, or you can create it manually. Most fields have sensible defaults; you only need to set the fields relevant to your deployment.

## File Location

By default, ModulaCMS looks for `config.json` in the working directory. Override the path with the `--config` flag:

```bash
./modula-x86 serve --config /etc/modulacms/config.json
```

## Viewing and Validating

```bash
# Show current configuration (sensitive fields are redacted)
./modula-x86 config show

# Validate configuration
./modula-x86 config validate
```

Validation checks that required fields (`db_driver`, `db_url`, `port`, `ssh_port`) are present and that values like `db_driver` and `output_format` are valid options.

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

The `environment` field controls TLS behavior. When set to `local` or `docker`, the HTTP server binds to `localhost` or `0.0.0.0` respectively, and S3 connections use `http://`. All other values use `https://` for S3 and enable Let's Encrypt autocert for HTTPS.

The `output_format` field sets the default response structure for content delivery endpoints. Valid values: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`. When empty, defaults to `raw`.

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

S3-compatible storage is used for media assets and backups. Any S3-compatible provider works: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, Cloudflare R2.

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

All S3 storage fields are hot-reloadable -- you can change them without restarting the server.

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

The scheme (`http://` or `https://`) for the S3 API endpoint is determined by the `environment` field. When set to `http-only` or `docker`, the scheme is `http`; all other environments use `https`. The `bucket_endpoint` value should not include the scheme.

When running in Docker, `bucket_endpoint` typically points to a container hostname (e.g., `minio:9000`), which browsers cannot resolve. Set `bucket_public_url` to the externally reachable address so that media URLs in API responses work in the browser.

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

ModulaCMS supports OAuth with any OpenID Connect-compatible provider (Google, GitHub, Azure AD, etc.). Configure a single provider per instance.

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

All OAuth fields are hot-reloadable. OAuth is optional -- the CMS functions with local authentication when OAuth is not configured.

## Email Settings

Email is used for password reset flows. Four providers are supported: SMTP, SendGrid, AWS SES, and Postmark.

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

All email fields are hot-reloadable. When SES is configured without explicit AWS credentials, the system falls back to the default AWS credential chain (environment variables, IAM role).

## Plugin System Settings

ModulaCMS has a Lua-based plugin system. Plugins run in sandboxed VMs with configurable resource limits.

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

ModulaCMS supports external error tracking and metrics via providers like Sentry, Datadog, and New Relic.

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

Many configuration fields can be changed at runtime without restarting the server. The "Restart Required" column in each section above indicates which fields need a restart.

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
- S3 storage settings
- Email settings
- Observability settings
- Plugin runtime limits (VMs, timeout, rate limits)
- Update settings
- Output format and upload size

You can update hot-reloadable fields through the admin panel at `/admin/settings`, through the REST API, or by editing `config.json` directly. When updating via the API or admin panel, changes are saved to `config.json` and applied immediately. Changes made by editing the file directly require the config to be reloaded.

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

The API validates proposed changes before applying them. If a change affects a restart-required field, the response includes a warning listing the fields that need a server restart to take effect.

The config meta endpoint returns field metadata (categories, descriptions, hot-reloadable status):

```bash
curl http://localhost:8080/api/v1/admin/config/meta \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Both config endpoints require the `config:read` or `config:update` permission.

## Sensitive Fields

The following fields are treated as sensitive and are redacted (replaced with `********`) when the configuration is returned through the API or the `config show` command:

- `auth_salt`
- `db_password`
- `bucket_access_key`
- `bucket_secret_key`
- `oauth_client_id`
- `oauth_client_secret`
- `observability_dsn`
- `email_password`
- `email_api_key`
- `email_aws_access_key_id`
- `email_aws_secret_access_key`

When updating configuration via the API, redacted values (`********`) are automatically skipped to prevent accidentally overwriting secrets with the placeholder.

## Full Default Configuration

The default `config.json` created on first run uses these values:

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

Non-default fields (database credentials, S3 credentials, OAuth endpoints, email credentials) are empty strings or zero values and need to be configured for their respective features.
