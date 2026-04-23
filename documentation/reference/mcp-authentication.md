# MCP Authentication

Authenticate MCP tool calls with API tokens and enforce per-tool permissions using the same RBAC system that protects the REST API.

## How it works

The MCP server runs on the `/mcp` HTTP endpoint alongside your REST API. Every MCP request passes through the same authentication middleware as API requests. Your MCP client sends a Bearer token in the `Authorization` header, ModulaCMS resolves it to a user, and the MCP server checks that user's permissions before executing each tool.

This means:

- MCP users have exactly the same access as API users with the same role.
- Audit trails record the real user, not a generic "mcp" identity.
- Permission changes take effect within 60 seconds, the same as the REST API.

## Set up authentication

### 1. Enable the MCP server

Set `mcp_enabled` to `true` in your configuration:

```json
{
  "mcp_enabled": true
}
```

Or update it through the API:

```bash
curl -X PUT http://localhost:8080/api/v1/config \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"mcp_enabled": true}'
```

### 2. Create an API token

Create a token for the user who will connect via MCP:

```bash
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Authorization: Bearer YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "token_type": "api_key",
    "token": "mcms_my-mcp-token",
    "issued_at": "2026-01-15T10:00:00Z",
    "expires_at": "2027-01-15T10:00:00Z",
    "revoked": false
  }'
```

You can also create tokens through the admin panel at `/admin/settings`.

### 3. Configure your MCP client

Pass the token as a Bearer token. How you do this depends on your MCP client.

For Claude Code, set the token in an environment variable and pass it via headers:

```bash
export MODULA_MCP_TOKEN="mcms_my-mcp-token"
```

For clients that support custom headers on Streamable HTTP transport, set the `Authorization` header to `Bearer <token>`.

## Permission enforcement

Every MCP tool maps to a `resource:operation` permission, the same labels used by the REST API. When a tool is called, the MCP server:

1. Looks up the required permission for the tool.
2. Checks whether the authenticated user's role has that permission.
3. Returns a permission error if the check fails.

Admin users bypass permission checks entirely, the same as the REST API.

### Permission mapping

MCP tools map directly to API permissions. The pattern is consistent: tools that read data require `:read`, tools that create require `:create`, and so on.

| Tool pattern | Permission |
|--------------|------------|
| `list_content`, `get_content` | `content:read` |
| `create_content` | `content:create` |
| `update_content`, `reorder_content`, `move_content` | `content:update` |
| `delete_content` | `content:delete` |
| `list_datatypes`, `get_datatype` | `datatypes:read` |
| `create_datatype` | `datatypes:create` |
| `list_media`, `get_media` | `media:read` |
| `upload_media` | `media:create` |
| `publish_content` | `content:publish` |
| `unpublish_content` | `content:publish` |

Admin-prefixed tools (`admin_list_content`, `admin_create_datatype`) use the same permissions as their public counterparts.

The full mapping covers all 246 tools. Each tool requires exactly one permission. See the complete table in [Permission reference](#permission-reference).

### Public tools

Two tools do not require authentication:

| Tool | Description |
|------|-------------|
| `register_user` | Register a new user account |
| `request_password_reset` | Request a password reset email |

### Permission errors

When a tool call is denied, the MCP server returns a tool result with `IsError: true` and one of these messages:

| Message | Meaning |
|---------|---------|
| `authentication required` | No valid token was provided, or the token is expired or revoked. |
| `forbidden: requires permission 'content:create'` | The user's role does not have the required permission. The specific permission label is included in the message. |

These are MCP tool errors, not HTTP errors. Your MCP client receives them as structured tool results.

## Stdio mode

When you run `modula mcp` for stdio transport (connecting via a local pipe), no HTTP authentication is available. Stdio mode is designed for local development where the MCP client runs on the same machine as the CMS.

In stdio mode:

- Permission checks are not enforced.
- Audit trails record the identity as "mcp-local".
- All tools are available regardless of role.

> **Good to know**: Stdio mode connects to a remote CMS instance using the project registry. The remote server enforces its own authentication and authorization on the API calls the MCP proxy makes. The `mcp_proxy_token` config field provides the token for this outbound connection.

## Proxy mode configuration

When the MCP server connects to a remote CMS instance (registry mode), it uses the `mcp_proxy_token` config field to authenticate against the remote server:

```json
{
  "mcp_proxy_token": "mcms_remote-server-token"
}
```

This token authenticates the outbound connection from your local MCP server to the remote CMS. It is not used for authenticating inbound MCP client requests (that uses the standard API token system described in this page).

## Audit trails

Every mutation made through MCP records the authenticated user in the audit log. This includes:

- The user ID of the API token owner
- The HTTP request ID
- The client IP address

Use the `list_recent_activity` tool or the REST API to review audit entries.

## Permission reference

The complete tool-to-permission mapping, grouped by domain.

### Content

| Tool | Permission |
|------|------------|
| `list_content` | `content:read` |
| `get_content` | `content:read` |
| `create_content` | `content:create` |
| `update_content` | `content:update` |
| `delete_content` | `content:delete` |
| `get_page` | `content:read` |
| `get_content_tree` | `content:read` |
| `list_content_fields` | `content:read` |
| `get_content_field` | `content:read` |
| `create_content_field` | `content:create` |
| `update_content_field` | `content:update` |
| `delete_content_field` | `content:delete` |
| `reorder_content` | `content:update` |
| `move_content` | `content:update` |
| `save_content_tree` | `content:update` |
| `heal_content` | `content:update` |
| `batch_update_content` | `content:update` |
| `query_content` | `content:read` |
| `get_globals` | `content:read` |
| `get_content_full` | `content:read` |
| `get_content_by_route` | `content:read` |
| `create_content_composite` | `content:create` |

### Admin content

| Tool | Permission |
|------|------------|
| `admin_list_content` | `content:read` |
| `admin_get_content` | `content:read` |
| `admin_create_content` | `content:create` |
| `admin_update_content` | `content:update` |
| `admin_delete_content` | `content:delete` |
| `admin_reorder_content` | `content:update` |
| `admin_move_content` | `content:update` |
| `admin_get_content_full` | `content:read` |
| `get_admin_tree` | `admin_tree:read` |
| `admin_list_content_fields` | `content:read` |
| `admin_get_content_field` | `content:read` |
| `admin_create_content_field` | `content:create` |
| `admin_update_content_field` | `content:update` |
| `admin_delete_content_field` | `content:delete` |

### Schema

| Tool | Permission |
|------|------------|
| `list_datatypes` | `datatypes:read` |
| `get_datatype` | `datatypes:read` |
| `create_datatype` | `datatypes:create` |
| `update_datatype` | `datatypes:update` |
| `delete_datatype` | `datatypes:delete` |
| `list_fields` | `fields:read` |
| `get_field` | `fields:read` |
| `create_field` | `fields:create` |
| `update_field` | `fields:update` |
| `delete_field` | `fields:delete` |
| `get_datatype_full` | `datatypes:read` |
| `list_datatypes_full` | `datatypes:read` |
| `get_datatype_max_sort_order` | `datatypes:read` |
| `update_datatype_sort_order` | `datatypes:update` |
| `get_field_max_sort_order` | `fields:read` |
| `update_field_sort_order` | `fields:update` |
| `list_field_types` | `field_types:read` |
| `get_field_type` | `field_types:read` |
| `create_field_type` | `field_types:create` |
| `update_field_type` | `field_types:update` |
| `delete_field_type` | `field_types:delete` |

### Admin schema

| Tool | Permission |
|------|------------|
| `admin_list_datatypes` | `datatypes:read` |
| `admin_get_datatype` | `datatypes:read` |
| `admin_create_datatype` | `datatypes:create` |
| `admin_update_datatype` | `datatypes:update` |
| `admin_delete_datatype` | `datatypes:delete` |
| `admin_get_datatype_max_sort_order` | `datatypes:read` |
| `admin_update_datatype_sort_order` | `datatypes:update` |
| `admin_list_fields` | `fields:read` |
| `admin_get_field` | `fields:read` |
| `admin_create_field` | `fields:create` |
| `admin_update_field` | `fields:update` |
| `admin_delete_field` | `fields:delete` |

### Media

| Tool | Permission |
|------|------------|
| `list_media` | `media:read` |
| `get_media` | `media:read` |
| `update_media` | `media:update` |
| `delete_media` | `media:delete` |
| `upload_media` | `media:create` |
| `media_health` | `media:read` |
| `media_cleanup_check` | `media:read` |
| `media_cleanup_apply` | `media:delete` |
| `list_media_dimensions` | `media:read` |
| `get_media_dimension` | `media:read` |
| `create_media_dimension` | `media:create` |
| `update_media_dimension` | `media:update` |
| `delete_media_dimension` | `media:delete` |
| `download_media` | `media:read` |
| `get_media_full` | `media:read` |
| `get_media_references` | `media:read` |
| `reprocess_media` | `media:update` |

### Media folders

| Tool | Permission |
|------|------------|
| `list_media_folders` | `media:read` |
| `get_media_folder` | `media:read` |
| `create_media_folder` | `media:create` |
| `update_media_folder` | `media:update` |
| `delete_media_folder` | `media:delete` |
| `move_media_to_folder` | `media:update` |
| `get_media_folder_tree` | `media:read` |
| `list_media_in_folder` | `media:read` |

### Admin media

| Tool | Permission |
|------|------------|
| `admin_list_media` | `media:read` |
| `admin_get_media` | `media:read` |
| `admin_update_media` | `media:update` |
| `admin_delete_media` | `media:delete` |
| `admin_upload_media` | `media:create` |

### Admin media folders

| Tool | Permission |
|------|------------|
| `admin_list_media_folders` | `media:read` |
| `admin_get_media_folder` | `media:read` |
| `admin_create_media_folder` | `media:create` |
| `admin_update_media_folder` | `media:update` |
| `admin_delete_media_folder` | `media:delete` |
| `admin_move_media_to_folder` | `media:update` |
| `admin_get_media_folder_tree` | `media:read` |
| `admin_list_media_in_folder` | `media:read` |

### Routes

| Tool | Permission |
|------|------------|
| `list_routes` | `routes:read` |
| `get_route` | `routes:read` |
| `list_routes_full` | `routes:read` |
| `create_route` | `routes:create` |
| `update_route` | `routes:update` |
| `delete_route` | `routes:delete` |

### Admin routes and field types

| Tool | Permission |
|------|------------|
| `admin_list_routes` | `routes:read` |
| `admin_get_route_by_slug` | `routes:read` |
| `admin_create_route` | `routes:create` |
| `admin_update_route` | `routes:update` |
| `admin_delete_route` | `routes:delete` |
| `admin_list_field_types` | `field_types:read` |
| `admin_get_field_type` | `field_types:read` |
| `admin_create_field_type` | `field_types:create` |
| `admin_update_field_type` | `field_types:update` |
| `admin_delete_field_type` | `field_types:delete` |

### Users

| Tool | Permission |
|------|------------|
| `whoami` | `users:read` |
| `list_users` | `users:read` |
| `get_user` | `users:read` |
| `create_user` | `users:create` |
| `update_user` | `users:update` |
| `delete_user` | `users:delete` |
| `list_users_full` | `users:read` |
| `get_user_full` | `users:read` |
| `reassign_and_delete_user` | `users:delete` |
| `list_user_sessions` | `users:read` |

### RBAC

| Tool | Permission |
|------|------------|
| `list_roles` | `roles:read` |
| `get_role` | `roles:read` |
| `create_role` | `roles:create` |
| `update_role` | `roles:update` |
| `delete_role` | `roles:delete` |
| `list_permissions` | `permissions:read` |
| `get_permission` | `permissions:read` |
| `create_permission` | `permissions:create` |
| `update_permission` | `permissions:update` |
| `delete_permission` | `permissions:delete` |
| `assign_role_permission` | `roles:create` |
| `remove_role_permission` | `roles:delete` |
| `list_role_permissions` | `roles:read` |
| `get_role_permission` | `roles:read` |
| `list_role_permissions_by_role` | `roles:read` |

### Sessions

| Tool | Permission |
|------|------------|
| `list_sessions` | `sessions:read` |
| `get_session` | `sessions:read` |
| `update_session` | `sessions:update` |
| `delete_session` | `sessions:delete` |

### Tokens

| Tool | Permission |
|------|------------|
| `list_tokens` | `tokens:read` |
| `get_token` | `tokens:read` |
| `create_token` | `tokens:create` |
| `delete_token` | `tokens:delete` |

### SSH keys

| Tool | Permission |
|------|------------|
| `list_ssh_keys` | `ssh_keys:read` |
| `create_ssh_key` | `ssh_keys:create` |
| `delete_ssh_key` | `ssh_keys:delete` |

### OAuth

| Tool | Permission |
|------|------------|
| `list_users_oauth` | `users:read` |
| `get_user_oauth` | `users:read` |
| `create_user_oauth` | `users:create` |
| `update_user_oauth` | `users:update` |
| `delete_user_oauth` | `users:delete` |

### Tables

| Tool | Permission |
|------|------------|
| `list_tables` | `tables:read` |
| `get_table` | `tables:read` |
| `create_table` | `tables:create` |
| `update_table` | `tables:update` |
| `delete_table` | `tables:delete` |

### Plugins

| Tool | Permission |
|------|------------|
| `list_plugins` | `plugins:read` |
| `get_plugin` | `plugins:read` |
| `reload_plugin` | `plugins:update` |
| `enable_plugin` | `plugins:update` |
| `disable_plugin` | `plugins:update` |
| `plugin_cleanup_dry_run` | `plugins:read` |
| `plugin_cleanup_drop` | `plugins:delete` |
| `list_plugin_routes` | `plugins:read` |
| `approve_plugin_routes` | `plugins:update` |
| `revoke_plugin_routes` | `plugins:update` |
| `list_plugin_hooks` | `plugins:read` |
| `approve_plugin_hooks` | `plugins:update` |
| `revoke_plugin_hooks` | `plugins:update` |

### Configuration

| Tool | Permission |
|------|------------|
| `get_config` | `config:read` |
| `get_config_meta` | `config:read` |
| `update_config` | `config:update` |

### Publishing

| Tool | Permission |
|------|------------|
| `publish_content` | `content:publish` |
| `unpublish_content` | `content:publish` |
| `schedule_content` | `content:publish` |
| `admin_publish_content` | `publishing:create` |
| `admin_unpublish_content` | `publishing:delete` |
| `admin_schedule_content` | `publishing:create` |

### Versions

| Tool | Permission |
|------|------------|
| `list_content_versions` | `content:read` |
| `get_content_version` | `content:read` |
| `create_content_version` | `content:update` |
| `delete_content_version` | `content:delete` |
| `restore_content_version` | `content:update` |
| `admin_list_content_versions` | `versions:read` |
| `admin_get_content_version` | `versions:read` |
| `admin_create_content_version` | `versions:create` |
| `admin_delete_content_version` | `versions:delete` |
| `admin_restore_content_version` | `versions:update` |

### Webhooks

| Tool | Permission |
|------|------------|
| `list_webhooks` | `webhook:read` |
| `get_webhook` | `webhook:read` |
| `create_webhook` | `webhook:create` |
| `update_webhook` | `webhook:update` |
| `delete_webhook` | `webhook:delete` |
| `test_webhook` | `webhook:update` |
| `list_webhook_deliveries` | `webhook:read` |
| `retry_webhook_delivery` | `webhook:update` |

### Locales

| Tool | Permission |
|------|------------|
| `list_locales` | `locale:read` |
| `list_admin_locales` | `locale:read` |
| `get_locale` | `locale:read` |
| `create_locale` | `locale:create` |
| `update_locale` | `locale:update` |
| `delete_locale` | `locale:delete` |
| `create_translation` | `content:create` |
| `admin_create_translation` | `locales:create` |

### Validations

| Tool | Permission |
|------|------------|
| `list_validations` | `validations:read` |
| `get_validation` | `validations:read` |
| `create_validation` | `validations:create` |
| `update_validation` | `validations:update` |
| `delete_validation` | `validations:delete` |
| `search_validations` | `validations:read` |
| `admin_list_validations` | `validations:read` |
| `admin_get_validation` | `validations:read` |
| `admin_create_validation` | `validations:create` |
| `admin_update_validation` | `validations:update` |
| `admin_delete_validation` | `validations:delete` |
| `admin_search_validations` | `validations:read` |

### Search

| Tool | Permission |
|------|------------|
| `search_content` | `search:read` |
| `rebuild_search_index` | `search:update` |

### Health and activity

| Tool | Permission |
|------|------------|
| `health` | `health:read` |
| `get_metrics` | `health:read` |
| `get_environment` | `health:read` |
| `list_recent_activity` | `audit:read` |

### Import

| Tool | Permission |
|------|------------|
| `import_content` | `import:create` |
| `import_bulk` | `import:create` |

### Deploy sync

| Tool | Permission |
|------|------------|
| `sync_health` | `deploy:read` |
| `sync_export` | `deploy:read` |
| `sync_import` | `deploy:create` |
| `sync_preview` | `deploy:read` |

## Next steps

- [MCP Server tool reference](/docs/reference/mcp-server) -- full list of tools with descriptions and parameters
- [Authentication and access control](/docs/custom-admin/authentication) -- create API tokens and manage roles
- [Configuration](/docs/getting-started/configuration) -- MCP and other server settings
