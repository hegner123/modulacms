package mcp

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/middleware"
)

// publicTools lists MCP tools that do not require authentication.
var publicTools = map[string]bool{
	"register_user":          true,
	"request_password_reset": true,
}

// toolPermissions maps each MCP tool name to its required resource:operation
// permission label. The mapping mirrors what the HTTP router uses for the
// equivalent API endpoint.
//
// Tools not in this map and not in publicTools are allowed unconditionally
// (connection tools registered only in registry mode).
var toolPermissions = map[string]string{
	// Content tools
	"list_content":              "content:read",
	"get_content":               "content:read",
	"create_content":            "content:create",
	"update_content":            "content:update",
	"delete_content":            "content:delete",
	"get_page":                  "content:read",
	"get_content_tree":          "content:read",
	"list_content_fields":       "content:read",
	"get_content_field":         "content:read",
	"create_content_field":      "content:create",
	"update_content_field":      "content:update",
	"delete_content_field":      "content:delete",
	"reorder_content":           "content:update",
	"move_content":              "content:update",
	"save_content_tree":         "content:update",
	"heal_content":              "content:update",
	"batch_update_content":      "content:update",
	"query_content":             "content:read",
	"get_globals":               "content:read",
	"get_content_full":          "content:read",
	"get_content_by_route":      "content:read",
	"create_content_composite":  "content:create",

	// Admin content tools
	"admin_list_content":        "content:read",
	"admin_get_content":         "content:read",
	"admin_create_content":      "content:create",
	"admin_update_content":      "content:update",
	"admin_delete_content":      "content:delete",
	"admin_reorder_content":     "content:update",
	"admin_move_content":        "content:update",
	"admin_get_content_full":    "content:read",
	"get_admin_tree":            "admin_tree:read",
	"admin_list_content_fields": "content:read",
	"admin_get_content_field":   "content:read",
	"admin_create_content_field": "content:create",
	"admin_update_content_field": "content:update",
	"admin_delete_content_field": "content:delete",

	// Schema tools
	"list_datatypes":              "datatypes:read",
	"get_datatype":                "datatypes:read",
	"create_datatype":             "datatypes:create",
	"update_datatype":             "datatypes:update",
	"delete_datatype":             "datatypes:delete",
	"list_fields":                 "fields:read",
	"get_field":                   "fields:read",
	"create_field":                "fields:create",
	"update_field":                "fields:update",
	"delete_field":                "fields:delete",
	"get_datatype_full":           "datatypes:read",
	"list_datatypes_full":         "datatypes:read",
	"get_datatype_max_sort_order": "datatypes:read",
	"update_datatype_sort_order":  "datatypes:update",
	"get_field_max_sort_order":    "fields:read",
	"update_field_sort_order":     "fields:update",
	"list_field_types":            "field_types:read",
	"get_field_type":              "field_types:read",
	"create_field_type":           "field_types:create",
	"update_field_type":           "field_types:update",
	"delete_field_type":           "field_types:delete",

	// Admin schema tools
	"admin_list_datatypes":              "datatypes:read",
	"admin_get_datatype":                "datatypes:read",
	"admin_create_datatype":             "datatypes:create",
	"admin_update_datatype":             "datatypes:update",
	"admin_delete_datatype":             "datatypes:delete",
	"admin_get_datatype_max_sort_order": "datatypes:read",
	"admin_update_datatype_sort_order":  "datatypes:update",
	"admin_list_fields":                 "fields:read",
	"admin_get_field":                   "fields:read",
	"admin_create_field":                "fields:create",
	"admin_update_field":                "fields:update",
	"admin_delete_field":                "fields:delete",

	// Media tools
	"list_media":             "media:read",
	"get_media":              "media:read",
	"update_media":           "media:update",
	"delete_media":           "media:delete",
	"upload_media":           "media:create",
	"media_health":           "media:read",
	"media_cleanup_check":    "media:read",
	"media_cleanup_apply":    "media:delete",
	"list_media_dimensions":  "media:read",
	"get_media_dimension":    "media:read",
	"create_media_dimension": "media:create",
	"update_media_dimension": "media:update",
	"delete_media_dimension": "media:delete",
	"download_media":         "media:read",
	"get_media_full":         "media:read",
	"get_media_references":   "media:read",
	"reprocess_media":        "media:update",

	// Media folder tools
	"list_media_folders":    "media:read",
	"get_media_folder":      "media:read",
	"create_media_folder":   "media:create",
	"update_media_folder":   "media:update",
	"delete_media_folder":   "media:delete",
	"move_media_to_folder":  "media:update",
	"get_media_folder_tree": "media:read",
	"list_media_in_folder":  "media:read",

	// Admin media tools
	"admin_list_media":   "media:read",
	"admin_get_media":    "media:read",
	"admin_update_media": "media:update",
	"admin_delete_media": "media:delete",
	"admin_upload_media": "media:create",

	// Admin media folder tools
	"admin_list_media_folders":    "media:read",
	"admin_get_media_folder":      "media:read",
	"admin_create_media_folder":   "media:create",
	"admin_update_media_folder":   "media:update",
	"admin_delete_media_folder":   "media:delete",
	"admin_move_media_to_folder":  "media:update",
	"admin_get_media_folder_tree": "media:read",
	"admin_list_media_in_folder":  "media:read",

	// Route tools
	"list_routes":      "routes:read",
	"get_route":        "routes:read",
	"list_routes_full": "routes:read",
	"create_route":     "routes:create",
	"update_route":     "routes:update",
	"delete_route":     "routes:delete",

	// Admin route tools
	"admin_list_routes":        "routes:read",
	"admin_get_route_by_slug":  "routes:read",
	"admin_create_route":       "routes:create",
	"admin_update_route":       "routes:update",
	"admin_delete_route":       "routes:delete",
	"admin_list_field_types":   "field_types:read",
	"admin_get_field_type":     "field_types:read",
	"admin_create_field_type":  "field_types:create",
	"admin_update_field_type":  "field_types:update",
	"admin_delete_field_type":  "field_types:delete",

	// User tools
	"whoami":                   "users:read",
	"list_users":               "users:read",
	"get_user":                 "users:read",
	"create_user":              "users:create",
	"update_user":              "users:update",
	"delete_user":              "users:delete",
	"list_users_full":          "users:read",
	"get_user_full":            "users:read",
	"reassign_and_delete_user": "users:delete",
	"list_user_sessions":       "users:read",

	// RBAC tools
	"list_roles":                   "roles:read",
	"get_role":                     "roles:read",
	"create_role":                  "roles:create",
	"update_role":                  "roles:update",
	"delete_role":                  "roles:delete",
	"list_permissions":             "permissions:read",
	"get_permission":               "permissions:read",
	"create_permission":            "permissions:create",
	"update_permission":            "permissions:update",
	"delete_permission":            "permissions:delete",
	"assign_role_permission":       "roles:create",
	"remove_role_permission":       "roles:delete",
	"list_role_permissions":        "roles:read",
	"get_role_permission":          "roles:read",
	"list_role_permissions_by_role": "roles:read",

	// Session tools
	"list_sessions":  "sessions:read",
	"get_session":    "sessions:read",
	"update_session": "sessions:update",
	"delete_session": "sessions:delete",

	// Token tools
	"list_tokens":  "tokens:read",
	"get_token":    "tokens:read",
	"create_token": "tokens:create",
	"delete_token": "tokens:delete",

	// SSH key tools
	"list_ssh_keys":  "ssh_keys:read",
	"create_ssh_key": "ssh_keys:create",
	"delete_ssh_key": "ssh_keys:delete",

	// OAuth tools
	"list_users_oauth":  "users:read",
	"get_user_oauth":    "users:read",
	"create_user_oauth": "users:create",
	"update_user_oauth": "users:update",
	"delete_user_oauth": "users:delete",

	// Table tools
	"list_tables":  "tables:read",
	"get_table":    "tables:read",
	"create_table": "tables:create",
	"update_table": "tables:update",
	"delete_table": "tables:delete",

	// Plugin tools
	"list_plugins":          "plugins:read",
	"get_plugin":            "plugins:read",
	"reload_plugin":         "plugins:update",
	"enable_plugin":         "plugins:update",
	"disable_plugin":        "plugins:update",
	"plugin_cleanup_dry_run": "plugins:read",
	"plugin_cleanup_drop":   "plugins:delete",
	"list_plugin_routes":    "plugins:read",
	"approve_plugin_routes": "plugins:update",
	"revoke_plugin_routes":  "plugins:update",
	"list_plugin_hooks":     "plugins:read",
	"approve_plugin_hooks":  "plugins:update",
	"revoke_plugin_hooks":   "plugins:update",

	// Config tools
	"get_config":      "config:read",
	"get_config_meta": "config:read",
	"update_config":   "config:update",

	// Publishing tools (content:publish, matching router)
	"publish_content":        "content:publish",
	"unpublish_content":      "content:publish",
	"schedule_content":       "content:publish",
	"admin_publish_content":  "publishing:create",
	"admin_unpublish_content": "publishing:delete",
	"admin_schedule_content": "publishing:create",

	// Version tools (content:read/update/delete, matching router)
	"list_content_versions":        "content:read",
	"get_content_version":          "content:read",
	"create_content_version":       "content:update",
	"delete_content_version":       "content:delete",
	"restore_content_version":      "content:update",
	"admin_list_content_versions":  "versions:read",
	"admin_get_content_version":    "versions:read",
	"admin_create_content_version": "versions:create",
	"admin_delete_content_version": "versions:delete",
	"admin_restore_content_version": "versions:update",

	// Webhook tools
	"list_webhooks":           "webhook:read",
	"get_webhook":             "webhook:read",
	"create_webhook":          "webhook:create",
	"update_webhook":          "webhook:update",
	"delete_webhook":          "webhook:delete",
	"test_webhook":            "webhook:update",
	"list_webhook_deliveries": "webhook:read",
	"retry_webhook_delivery":  "webhook:update",

	// Locale tools
	"list_locales":            "locale:read",
	"list_admin_locales":      "locale:read",
	"get_locale":              "locale:read",
	"create_locale":           "locale:create",
	"update_locale":           "locale:update",
	"delete_locale":           "locale:delete",
	"create_translation":      "content:create",
	"admin_create_translation": "locales:create",

	// Validation tools
	"list_validations":         "validations:read",
	"get_validation":           "validations:read",
	"create_validation":        "validations:create",
	"update_validation":        "validations:update",
	"delete_validation":        "validations:delete",
	"search_validations":       "validations:read",
	"admin_list_validations":   "validations:read",
	"admin_get_validation":     "validations:read",
	"admin_create_validation":  "validations:create",
	"admin_update_validation":  "validations:update",
	"admin_delete_validation":  "validations:delete",
	"admin_search_validations": "validations:read",

	// Search tools
	"search_content":       "search:read",
	"rebuild_search_index": "search:update",

	// Health/activity tools
	"health":               "health:read",
	"get_metrics":          "health:read",
	"get_environment":      "health:read",
	"list_recent_activity": "audit:read",

	// Import tools
	"import_content": "import:create",
	"import_bulk":    "import:create",

	// Deploy tools
	"sync_health":  "deploy:read",
	"sync_export":  "deploy:read",
	"sync_import":  "deploy:create",
	"sync_preview": "deploy:read",
}

func init() {
	for tool, perm := range toolPermissions {
		if err := middleware.ValidatePermissionLabel(perm); err != nil {
			panic(fmt.Sprintf("mcp: invalid permission label %q for tool %q: %v", perm, tool, err))
		}
	}
}
