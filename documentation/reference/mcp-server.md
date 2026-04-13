# MCP Server

Connect AI tools to your ModulaCMS instance through the Model Context Protocol (MCP). The built-in MCP server exposes your entire CMS as structured tools that any MCP-compatible client can call.

## Quick Start

```bash
modula mcp                          # use ./modula.config.json
modula mcp mysite                   # resolve from project registry
modula mcp mysite production        # specific environment
```

The server runs over stdio transport. Register it with your MCP client:

```bash
claude mcp add --transport stdio modula -- ./modula mcp mysite
```

## Connection Management

Manage which CMS instance you're connected to.

| Tool | Description |
|------|-------------|
| `list_projects` | List registered projects and environments from the project registry. Shows which project is active. |
| `switch_project` | Switch the active connection to a different project or environment. |
| `get_connection` | Show the current project, environment, and CMS URL. |

## Content

Create, read, update, and delete content entries and their field values.

| Tool | Description |
|------|-------------|
| `list_content` | List content entries with pagination. Returns structural metadata without field values. |
| `get_content` | Get a single content entry by ID. |
| `create_content` | Create a content entry. Set its status, parent, route, datatype, and author. |
| `update_content` | Replace a content entry's properties. |
| `delete_content` | Delete a content entry by ID. |
| `get_page` | Get an assembled page by slug via the public delivery endpoint. Returns the full content tree with resolved field values. Supports format conversion (contentful, sanity, strapi, wordpress, clean, raw). |
| `get_content_tree` | Get the admin content tree for a slug with field values. |

### Content Fields

| Tool | Description |
|------|-------------|
| `list_content_fields` | List content field records with pagination. |
| `get_content_field` | Get a single content field by ID. |
| `create_content_field` | Create a field value for a content entry. Requires the content ID, field definition ID, and the value. |
| `update_content_field` | Update a field value. |
| `delete_content_field` | Delete a field value by ID. |

### Tree Operations

| Tool | Description |
|------|-------------|
| `reorder_content` | Atomically reorder sibling content nodes under a parent. Pass an ordered array of IDs. |
| `move_content` | Move a content node to a new parent at a specific position. |
| `save_content_tree` | Apply multiple tree structure changes (creates, updates, deletes) in a single atomic request. |
| `batch_update_content` | Atomically update a content entry's data and field values together. |
| `heal_content` | Scan content for malformed IDs and repair them. Supports dry run. |

## Admin Content

The admin content system mirrors the public content system. It powers the admin panel UI itself, not the site's published content.

| Tool | Description |
|------|-------------|
| `admin_list_content` | List admin content entries with pagination. |
| `admin_get_content` | Get a single admin content entry by ID. |
| `admin_create_content` | Create an admin content entry. |
| `admin_update_content` | Update an admin content entry. |
| `admin_delete_content` | Delete an admin content entry by ID. |
| `admin_reorder_content` | Reorder admin content siblings under a parent. |
| `admin_move_content` | Move an admin content node to a new parent. |

### Admin Content Fields

| Tool | Description |
|------|-------------|
| `admin_list_content_fields` | List admin content field records with pagination. |
| `admin_get_content_field` | Get a single admin content field by ID. |
| `admin_create_content_field` | Create an admin content field value. |
| `admin_update_content_field` | Update an admin content field value. |
| `admin_delete_content_field` | Delete an admin content field by ID. |

## Datatypes

Datatypes define the schema for content entries. Each datatype specifies which fields appear when creating or editing content.

| Tool | Description |
|------|-------------|
| `list_datatypes` | List all datatypes. Set `full=true` to include linked fields. |
| `get_datatype` | Get a single datatype by ID. |
| `get_datatype_full` | Get a datatype with its linked fields. Omit the ID to list all datatypes with fields. |
| `create_datatype` | Create a datatype with a name, label, and type. |
| `update_datatype` | Update a datatype (full replacement). |
| `delete_datatype` | Delete a datatype by ID. |

### Admin Datatypes

| Tool | Description |
|------|-------------|
| `admin_list_datatypes` | List all admin datatypes. |
| `admin_get_datatype` | Get a single admin datatype by ID. |
| `admin_create_datatype` | Create an admin datatype. |
| `admin_update_datatype` | Update an admin datatype. |
| `admin_delete_datatype` | Delete an admin datatype by ID. |

## Fields

Fields are the individual data definitions within a datatype. Each field has a type that determines what kind of value it holds.

Supported field types: `text`, `textarea`, `number`, `date`, `datetime`, `boolean`, `select`, `media`, `_id`, `json`, `richtext`, `slug`, `email`, `url`.

| Tool | Description |
|------|-------------|
| `list_fields` | List all field definitions. |
| `get_field` | Get a single field definition by ID. |
| `create_field` | Create a field with a label, type, and optional validation, UI config, and data settings. |
| `update_field` | Update a field definition (full replacement). |
| `delete_field` | Delete a field definition by ID. |

### Admin Fields

Admin fields use the same types with the addition of `relation`.

| Tool | Description |
|------|-------------|
| `admin_list_fields` | List all admin field definitions. |
| `admin_get_field` | Get a single admin field by ID. |
| `admin_create_field` | Create an admin field. |
| `admin_update_field` | Update an admin field. |
| `admin_delete_field` | Delete an admin field by ID. |

## Field Types

Field type definitions describe the available field types in the system.

| Tool | Description |
|------|-------------|
| `list_field_types` | List all field type definitions. |
| `get_field_type` | Get a single field type by ID. |
| `create_field_type` | Create a field type with a type key and label. |
| `update_field_type` | Update a field type. |
| `delete_field_type` | Delete a field type by ID. |

### Admin Field Types

| Tool | Description |
|------|-------------|
| `admin_list_field_types` | List all admin field types. |
| `admin_get_field_type` | Get a single admin field type by ID. |
| `admin_create_field_type` | Create an admin field type. |
| `admin_update_field_type` | Update an admin field type. |
| `admin_delete_field_type` | Delete an admin field type by ID. |

## Routes

Routes map URL slugs to content. Each route has a slug, title, and status.

| Tool | Description |
|------|-------------|
| `list_routes` | List all routes. |
| `get_route` | Get a single route by ID. |
| `create_route` | Create a route with a slug, title, and status. |
| `update_route` | Update a route. |
| `delete_route` | Delete a route by ID. |

### Admin Routes

Admin routes use slug-based lookup instead of ID-based.

| Tool | Description |
|------|-------------|
| `admin_list_routes` | List all admin routes. |
| `admin_get_route` | Get a single admin route by slug. |
| `admin_create_route` | Create an admin route. |
| `admin_update_route` | Update an admin route by ID. |
| `admin_delete_route` | Delete an admin route by ID. |

## Media

Upload, organize, and manage media assets. Each asset has metadata including URL, dimensions, alt text, caption, and focal point coordinates.

| Tool | Description |
|------|-------------|
| `list_media` | List media assets with pagination. |
| `get_media` | Get a single media asset by ID. |
| `upload_media` | Upload a file from a local path. Optionally specify a filename and target folder. |
| `update_media` | Update media metadata (name, alt, caption, description, focal point). Only provided fields change. |
| `delete_media` | Delete a media asset by ID. |
| `media_health` | Check media storage health status. |
| `media_cleanup` | Remove orphaned media records that have no backing file. |

### Media Dimensions

Dimension presets define standard image sizes for your project.

| Tool | Description |
|------|-------------|
| `list_media_dimensions` | List all dimension presets. |
| `get_media_dimension` | Get a single dimension preset by ID. |
| `create_media_dimension` | Create a preset with a label, width, and height. |
| `update_media_dimension` | Update a dimension preset. |
| `delete_media_dimension` | Delete a dimension preset by ID. |

### Admin Media

| Tool | Description |
|------|-------------|
| `admin_list_media` | List admin media assets with pagination. |
| `admin_get_media` | Get a single admin media asset by ID. |
| `admin_upload_media` | Upload a file to the admin media library. |
| `admin_update_media` | Update admin media metadata. |
| `admin_delete_media` | Delete an admin media asset by ID. |
| `admin_list_media_dimensions` | List admin dimension presets. |

## Media Folders

Organize media into nested folders up to 10 levels deep.

| Tool | Description |
|------|-------------|
| `list_media_folders` | List root folders, or children of a given parent. |
| `get_media_folder` | Get a folder by ID. |
| `create_media_folder` | Create a folder with a name and optional parent. |
| `update_media_folder` | Rename or move a folder. |
| `delete_media_folder` | Delete a folder. Fails if it contains children or media items. |
| `move_media_to_folder` | Move up to 100 media items to a folder in one batch. |

### Admin Media Folders

| Tool | Description |
|------|-------------|
| `admin_list_media_folders` | List admin media folders. |
| `admin_get_media_folder` | Get an admin media folder by ID. |
| `admin_create_media_folder` | Create an admin media folder. |
| `admin_update_media_folder` | Rename or move an admin media folder. |
| `admin_delete_media_folder` | Delete an admin media folder. |
| `admin_move_media_to_folder` | Batch move admin media items to a folder. |

## Users

Manage user accounts and retrieve profile information.

| Tool | Description |
|------|-------------|
| `whoami` | Get the authenticated user's profile (user ID, username, name, email, role). Use this to get your user ID for content authoring. |
| `list_users` | List all users. |
| `get_user` | Get a single user by ID. |
| `create_user` | Create a user with username, name, email, password, and role. |
| `update_user` | Update a user (full replacement). Omit `password` to keep the current one. |
| `delete_user` | Delete a user by ID. |
| `list_users_full` | List all users with roles, permissions, and session data. |
| `get_user_full` | Get a single user with full associated data. |

## Roles and Permissions

Role-based access control with granular `resource:operation` permissions. Three default roles: **admin** (all permissions), **editor** (60 permissions), **viewer** (5 read-only permissions).

### Roles

| Tool | Description |
|------|-------------|
| `list_roles` | List all roles. |
| `get_role` | Get a single role by ID. |
| `create_role` | Create a role with a label. |
| `update_role` | Update a role's label. |
| `delete_role` | Delete a role. System-protected roles cannot be deleted. |

### Permissions

| Tool | Description |
|------|-------------|
| `list_permissions` | List all permissions. |
| `get_permission` | Get a single permission by ID. |
| `create_permission` | Create a permission. Label must follow `resource:operation` format. |
| `update_permission` | Update a permission's label. |
| `delete_permission` | Delete a permission. System-protected permissions cannot be deleted. |

### Role-Permission Assignments

| Tool | Description |
|------|-------------|
| `list_role_permissions` | List all role-permission associations. |
| `get_role_permission` | Get a single association by ID. |
| `list_role_permissions_by_role` | List all permissions assigned to a specific role. |
| `assign_role_permission` | Assign a permission to a role. |
| `remove_role_permission` | Remove a permission from a role. |

## Configuration

Read and update server configuration at runtime.

| Tool | Description |
|------|-------------|
| `get_config` | Get current configuration. Sensitive values are redacted. Optionally filter by category. |
| `get_config_meta` | Get metadata for all config fields: key, label, category, whether it's hot-reloadable, sensitive, or required. |
| `update_config` | Update configuration values. Pass a JSON object of key-value pairs. |

## Plugins

Manage Lua plugins, their lifecycle, and their route/hook approvals.

| Tool | Description |
|------|-------------|
| `list_plugins` | List all installed plugins with status. |
| `get_plugin` | Get detailed info for a plugin by name. |
| `reload_plugin` | Reload a plugin from disk. |
| `enable_plugin` | Enable a disabled plugin. |
| `disable_plugin` | Disable an active plugin. |
| `plugin_cleanup_dry_run` | List orphaned plugin tables without dropping them. |
| `plugin_cleanup_drop` | Drop orphaned plugin tables. Requires explicit confirmation and table list. |

### Plugin Routes

| Tool | Description |
|------|-------------|
| `list_plugin_routes` | List all plugin-registered HTTP routes with approval status. |
| `approve_plugin_routes` | Approve routes by specifying plugin, method, and path for each. |
| `revoke_plugin_routes` | Revoke approval for routes. |

### Plugin Hooks

| Tool | Description |
|------|-------------|
| `list_plugin_hooks` | List all plugin-registered hooks with approval status. |
| `approve_plugin_hooks` | Approve hooks by specifying plugin, event, and table for each. |
| `revoke_plugin_hooks` | Revoke approval for hooks. |

## Import

Import content from other CMS platforms into ModulaCMS.

| Tool | Description |
|------|-------------|
| `import_content` | Import content from another CMS format. Supported formats: contentful, sanity, strapi, wordpress, clean. Keep payloads under ~5MB. |
| `import_bulk` | Bulk import a raw JSON payload. |

## Deploy Sync

Synchronize content between ModulaCMS environments.

| Tool | Description |
|------|-------------|
| `deploy_health` | Check sync status, version, and node ID. |
| `deploy_export` | Export a sync payload. Optionally filter by table names. |
| `deploy_import` | Import a sync payload from another environment. |
| `deploy_dry_run` | Preview what an import would change without writing. |

## Sessions

| Tool | Description |
|------|-------------|
| `list_sessions` | List all active sessions. |
| `get_session` | Get a session by ID. |
| `update_session` | Update session properties (user, expiry, IP, user agent, data). |
| `delete_session` | Delete a session by ID. |

## Tokens

| Tool | Description |
|------|-------------|
| `list_tokens` | List all authentication tokens. |
| `get_token` | Get a token by ID. |
| `create_token` | Create a token with type, value, issued/expiry timestamps, and optional user association. |
| `delete_token` | Delete a token by ID. |

## SSH Keys

| Tool | Description |
|------|-------------|
| `list_ssh_keys` | List SSH keys for the authenticated user. |
| `create_ssh_key` | Add an SSH public key with a label. |
| `delete_ssh_key` | Delete an SSH key by ID. |

## OAuth

| Tool | Description |
|------|-------------|
| `list_users_oauth` | List all OAuth connections. |
| `get_user_oauth` | Get an OAuth connection by ID. |
| `create_user_oauth` | Create an OAuth connection (provider, provider user ID, tokens, expiry). |
| `update_user_oauth` | Refresh an OAuth connection's tokens. |
| `delete_user_oauth` | Delete an OAuth connection by ID. |

## Tables

CMS metadata table records.

| Tool | Description |
|------|-------------|
| `list_tables` | List all CMS tables. |
| `get_table` | Get a table by ID. |
| `create_table` | Create a table record with a label. |
| `update_table` | Update a table's label. |
| `delete_table` | Delete a table by ID. |

## Health

| Tool | Description |
|------|-------------|
| `health` | Check overall server health status. |

## Tool Count by Domain

| Domain | Tools |
|--------|-------|
| Connection | 3 |
| Content (public + admin) | 31 |
| Datatypes (public + admin) | 11 |
| Fields (public + admin) | 10 |
| Field Types (public + admin) | 10 |
| Routes (public + admin) | 10 |
| Media (public + admin) | 20 |
| Media Folders (public + admin) | 12 |
| Users | 8 |
| Roles & Permissions | 15 |
| Configuration | 3 |
| Plugins | 12 |
| Import | 2 |
| Deploy Sync | 4 |
| Sessions | 4 |
| Tokens | 4 |
| SSH Keys | 3 |
| OAuth | 5 |
| Tables | 5 |
| Health | 1 |
| **Total** | **170** |
