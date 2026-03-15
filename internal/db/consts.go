package db

import (
	"fmt"
	"reflect"
)

// DBTable represents a database table name.
type DBTable string

// Database table name constants.
const (
	Admin_content_data      DBTable = "admin_content_data"
	Admin_content_fields    DBTable = "admin_content_fields"
	Admin_content_relations DBTable = "admin_content_relations"
	Admin_content_versions  DBTable = "admin_content_versions"
	Admin_datatype          DBTable = "admin_datatypes"
	Admin_field             DBTable = "admin_fields"
	Admin_field_types       DBTable = "admin_field_types"
	Admin_route             DBTable = "admin_routes"
	BackupT                 DBTable = "backups"
	Backup_set              DBTable = "backup_sets"
	Backup_verification     DBTable = "backup_verifications"
	Change_event            DBTable = "change_events"
	Content_data            DBTable = "content_data"
	Content_fields          DBTable = "content_fields"
	Content_relations       DBTable = "content_relations"
	Content_versions        DBTable = "content_versions"
	Datatype                DBTable = "datatypes"
	Field                   DBTable = "fields"
	Field_plugin_config     DBTable = "field_plugin_config"
	Field_types             DBTable = "field_types"
	LocaleT                 DBTable = "locales"
	MediaT                  DBTable = "media"
	Media_dimension         DBTable = "media_dimensions"
	Media_folder            DBTable = "media_folders"
	Permission              DBTable = "permissions"
	PipelineT               DBTable = "pipelines"
	Role                    DBTable = "roles"
	Role_permissions        DBTable = "role_permissions"
	Route                   DBTable = "routes"
	Session                 DBTable = "sessions"
	Table                   DBTable = "tables"
	Token                   DBTable = "tokens"
	User                    DBTable = "users"
	User_oauth              DBTable = "user_oauth"
	User_ssh_keys           DBTable = "user_ssh_keys"
	WebhookT                DBTable = "webhooks"
	Webhook_deliveries      DBTable = "webhook_deliveries"
)

// allTables is the exhaustive set of known table names for validation.
var allTables = map[DBTable]struct{}{
	Admin_content_data:      {},
	Admin_content_fields:    {},
	Admin_content_relations: {},
	Admin_content_versions:  {},
	Admin_datatype:          {},
	Admin_field:             {},
	Admin_field_types:       {},
	Admin_route:             {},
	BackupT:                 {},
	Backup_set:              {},
	Backup_verification:     {},
	Change_event:            {},
	Content_data:            {},
	Content_fields:          {},
	Content_relations:       {},
	Content_versions:        {},
	Datatype:                {},
	Field:                   {},
	Field_plugin_config:     {},
	Field_types:             {},
	LocaleT:                 {},
	MediaT:                  {},
	Media_dimension:         {},
	Media_folder:            {},
	Permission:              {},
	PipelineT:               {},
	Role:                    {},
	Role_permissions:        {},
	Route:                   {},
	Session:                 {},
	Table:                   {},
	Token:                   {},
	User:                    {},
	User_oauth:              {},
	User_ssh_keys:           {},
	WebhookT:                {},
	Webhook_deliveries:      {},
}

// SystemPluginTables are CMS-managed plugin infrastructure tables that should
// not be included in plugin data export/import.
var SystemPluginTables = map[string]bool{
	"plugin_routes":   true,
	"plugin_hooks":    true,
	"plugin_requests": true,
}

// IsValidPluginTableName reports whether name follows the plugin table naming
// convention (plugin_ prefix) and is safe for use in SQL queries.
func IsValidPluginTableName(name string) bool {
	return len(name) >= 9 && name[:7] == "plugin_" && ValidTableName(name) == nil
}

// ValidateTableName checks that name corresponds to a known table or a valid
// plugin table. Returns the typed DBTable on success or an error for unknown names.
func ValidateTableName(name string) (DBTable, error) {
	t := DBTable(name)
	if _, ok := allTables[t]; ok {
		return t, nil
	}
	if IsValidPluginTableName(name) {
		return t, nil
	}
	return "", fmt.Errorf("unknown table name: %q", name)
}

// IsValidTable reports whether t is a known table name or a valid plugin table.
func IsValidTable(t DBTable) bool {
	if _, ok := allTables[t]; ok {
		return true
	}
	return IsValidPluginTableName(string(t))
}

// TableStructMap maps each DBTable to its associated struct type
var TableStructMap = map[DBTable]reflect.Type{
	Admin_content_data:      reflect.TypeFor[AdminContentData](),
	Admin_content_fields:    reflect.TypeFor[AdminContentFields](),
	Admin_content_relations: reflect.TypeFor[AdminContentRelations](),
	Admin_content_versions:  reflect.TypeFor[AdminContentVersion](),
	Admin_datatype:          reflect.TypeFor[AdminDatatypes](),
	Admin_field:             reflect.TypeFor[AdminFields](),
	Admin_field_types:       reflect.TypeFor[AdminFieldTypes](),
	Admin_route:             reflect.TypeFor[AdminRoutes](),
	BackupT:                 reflect.TypeFor[Backup](),
	Backup_set:              reflect.TypeFor[BackupSet](),
	Backup_verification:     reflect.TypeFor[BackupVerification](),
	Change_event:            reflect.TypeFor[ChangeEvent](),
	Content_data:            reflect.TypeFor[ContentData](),
	Content_fields:          reflect.TypeFor[ContentFields](),
	Content_relations:       reflect.TypeFor[ContentRelations](),
	Content_versions:        reflect.TypeFor[ContentVersion](),
	Datatype:                reflect.TypeFor[Datatypes](),
	Field:                   reflect.TypeFor[Fields](),
	Field_plugin_config:     reflect.TypeFor[FieldPluginConfig](),
	Field_types:             reflect.TypeFor[FieldTypes](),
	LocaleT:                 reflect.TypeFor[Locale](),
	MediaT:                  reflect.TypeFor[Media](),
	Media_dimension:         reflect.TypeFor[MediaDimensions](),
	Media_folder:            reflect.TypeFor[MediaFolder](),
	Permission:              reflect.TypeFor[Permissions](),
	PipelineT:               reflect.TypeFor[Pipeline](),
	Role:                    reflect.TypeFor[Roles](),
	Role_permissions:        reflect.TypeFor[RolePermissions](),
	Route:                   reflect.TypeFor[Routes](),
	Session:                 reflect.TypeFor[Sessions](),
	Table:                   reflect.TypeFor[Tables](),
	Token:                   reflect.TypeFor[Tokens](),
	User:                    reflect.TypeFor[Users](),
	User_oauth:              reflect.TypeFor[UserOauth](),
	User_ssh_keys:           reflect.TypeFor[UserSshKeys](),
	WebhookT:                reflect.TypeFor[Webhook](),
	Webhook_deliveries:      reflect.TypeFor[WebhookDelivery](),
}

// CastToTypedSlice casts an any return from Parse to a typed slice based on the DBTable
func CastToTypedSlice(result any, table DBTable) any {
	if result == nil {
		return nil
	}

	switch table {
	case Admin_content_data:
		if slice, ok := result.([]AdminContentData); ok {
			return slice
		}
	case Admin_content_fields:
		if slice, ok := result.([]AdminContentFields); ok {
			return slice
		}
	case Admin_content_relations:
		if slice, ok := result.([]AdminContentRelations); ok {
			return slice
		}
	case Admin_content_versions:
		if slice, ok := result.([]AdminContentVersion); ok {
			return slice
		}
	case Admin_datatype:
		if slice, ok := result.([]AdminDatatypes); ok {
			return slice
		}
	case Admin_field:
		if slice, ok := result.([]AdminFields); ok {
			return slice
		}
	case Admin_field_types:
		if slice, ok := result.([]AdminFieldTypes); ok {
			return slice
		}
	case Admin_route:
		if slice, ok := result.([]AdminRoutes); ok {
			return slice
		}
	case BackupT:
		if slice, ok := result.([]Backup); ok {
			return slice
		}
	case Backup_set:
		if slice, ok := result.([]BackupSet); ok {
			return slice
		}
	case Backup_verification:
		if slice, ok := result.([]BackupVerification); ok {
			return slice
		}
	case Change_event:
		if slice, ok := result.([]ChangeEvent); ok {
			return slice
		}
	case Content_data:
		if slice, ok := result.([]ContentData); ok {
			return slice
		}
	case Content_fields:
		if slice, ok := result.([]ContentFields); ok {
			return slice
		}
	case Content_relations:
		if slice, ok := result.([]ContentRelations); ok {
			return slice
		}
	case Content_versions:
		if slice, ok := result.([]ContentVersion); ok {
			return slice
		}
	case Datatype:
		if slice, ok := result.([]Datatypes); ok {
			return slice
		}
	case Field:
		if slice, ok := result.([]Fields); ok {
			return slice
		}
	case Field_plugin_config:
		if slice, ok := result.([]FieldPluginConfig); ok {
			return slice
		}
	case Field_types:
		if slice, ok := result.([]FieldTypes); ok {
			return slice
		}
	case LocaleT:
		if slice, ok := result.([]Locale); ok {
			return slice
		}
	case MediaT:
		if slice, ok := result.([]Media); ok {
			return slice
		}
	case Media_dimension:
		if slice, ok := result.([]MediaDimensions); ok {
			return slice
		}
	case Media_folder:
		if slice, ok := result.([]MediaFolder); ok {
			return slice
		}
	case Permission:
		if slice, ok := result.([]Permissions); ok {
			return slice
		}
	case PipelineT:
		if slice, ok := result.([]Pipeline); ok {
			return slice
		}
	case Role:
		if slice, ok := result.([]Roles); ok {
			return slice
		}
	case Role_permissions:
		if slice, ok := result.([]RolePermissions); ok {
			return slice
		}
	case Route:
		if slice, ok := result.([]Routes); ok {
			return slice
		}
	case Session:
		if slice, ok := result.([]Sessions); ok {
			return slice
		}
	case Table:
		if slice, ok := result.([]Tables); ok {
			return slice
		}
	case Token:
		if slice, ok := result.([]Tokens); ok {
			return slice
		}
	case User:
		if slice, ok := result.([]Users); ok {
			return slice
		}
	case User_oauth:
		if slice, ok := result.([]UserOauth); ok {
			return slice
		}
	case User_ssh_keys:
		if slice, ok := result.([]UserSshKeys); ok {
			return slice
		}
	case WebhookT:
		if slice, ok := result.([]Webhook); ok {
			return slice
		}
	case Webhook_deliveries:
		if slice, ok := result.([]WebhookDelivery); ok {
			return slice
		}
	}

	// Return as-is if no match (could be []map[string]any from parseGeneric)
	return result
}
