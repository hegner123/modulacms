package db

import (
	"fmt"
	"strings"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/utility"
)

// dropOp pairs a table name with its drop function for use in runDropOps.
type dropOp struct {
	name string
	fn   func() error
}

// DropAllTables drops all database tables in reverse dependency order (SQLite).
// Each table is dropped individually because sqlc :exec only executes one statement.
// Tables that don't exist are skipped with a warning.
func (d Database) DropAllTables() error {
	queries := mdb.New(d.Connection)

	ops := []dropOp{
		// Tier 7.5: Webhook system tables (deliveries before webhooks for FK)
		{"webhook_deliveries", func() error { return queries.DropWebhookDeliveryTable(d.Context) }},
		{"webhooks", func() error { return queries.DropWebhookTable(d.Context) }},
		// Tier 7: Plugin system tables (pipelines before plugins for FK)
		{"pipelines", func() error { return queries.DropPipelinesTable(d.Context) }},
		{"plugins", func() error { return queries.DropPluginsTable(d.Context) }},
		// Tier 6: Junction tables
		{"role_permissions", func() error { return queries.DropRolePermissionsTable(d.Context) }},
		// Tier 5.5: Content relation tables (depend on content_data and fields)
		{"admin_content_relations", func() error { return queries.DropAdminContentRelationTable(d.Context) }},
		{"content_relations", func() error { return queries.DropContentRelationTable(d.Context) }},
		// Tier 5.5b: Content version tables (depend on content_data + users)
		{"admin_content_versions", func() error { return queries.DropAdminContentVersionTable(d.Context) }},
		{"content_versions", func() error { return queries.DropContentVersionTable(d.Context) }},
		// Tier 5: Content field values
		{"admin_content_fields", func() error { return queries.DropAdminContentFieldTable(d.Context) }},
		{"content_fields", func() error { return queries.DropContentFieldTable(d.Context) }},
		// Tier 4: Content data tables
		{"admin_content_data", func() error { return queries.DropAdminContentDataTable(d.Context) }},
		{"content_data", func() error { return queries.DropContentDataTable(d.Context) }},
		// Tier 3: Field definition tables (field_plugin_config references fields)
		{"field_plugin_config", func() error { return queries.DropFieldPluginConfigTable(d.Context) }},
		{"admin_fields", func() error { return queries.DropAdminFieldTable(d.Context) }},
		{"fields", func() error { return queries.DropFieldTable(d.Context) }},
		// Tier 2.5: Validation tables (referenced by fields)
		{"admin_validations", func() error { return queries.DropAdminValidationTable(d.Context) }},
		{"validations", func() error { return queries.DropValidationTable(d.Context) }},
		// Tier 2: Core content tables and user-related tables
		{"admin_datatypes", func() error { return queries.DropAdminDatatypeTable(d.Context) }},
		{"datatypes", func() error { return queries.DropDatatypeTable(d.Context) }},
		{"routes", func() error { return queries.DropRouteTable(d.Context) }},
		{"admin_routes", func() error { return queries.DropAdminRouteTable(d.Context) }},
		{"media", func() error { return queries.DropMediaTable(d.Context) }},
		{"media_folders", func() error { return queries.DropMediaFolderTable(d.Context) }},
		{"admin_media", func() error { return queries.DropAdminMediaTable(d.Context) }},
		{"admin_media_folders", func() error { return queries.DropAdminMediaFolderTable(d.Context) }},
		{"tables", func() error { return queries.DropTableTable(d.Context) }},
		{"sessions", func() error { return queries.DropSessionTable(d.Context) }},
		{"user_ssh_keys", func() error { return queries.DropUserSshKeyTable(d.Context) }},
		{"user_oauth", func() error { return queries.DropUserOauthTable(d.Context) }},
		{"tokens", func() error { return queries.DropTokenTable(d.Context) }},
		// Tier 1: User management
		{"users", func() error { return queries.DropUserTable(d.Context) }},
		// Tier 0: Foundation tables
		{"media_dimensions", func() error { return queries.DropMediaDimensionTable(d.Context) }},
		{"field_types", func() error { return queries.DropFieldTypeTable(d.Context) }},
		{"admin_field_types", func() error { return queries.DropAdminFieldTypeTable(d.Context) }},
		{"locales", func() error { return queries.DropLocaleTable(d.Context) }},
		{"roles", func() error { return queries.DropRoleTable(d.Context) }},
		{"permissions", func() error { return queries.DropPermissionTable(d.Context) }},
		// Infrastructure tables
		{"backup_sets", func() error { return queries.DropBackupSetsTable(d.Context) }},
		{"backup_verifications", func() error { return queries.DropBackupVerificationsTable(d.Context) }},
		{"backups", func() error { return queries.DropBackupsTable(d.Context) }},
		{"change_events", func() error { return queries.DropChangeEventsTable(d.Context) }},
	}

	return runDropOps(ops)
}

// DropAllTables drops all database tables in reverse dependency order (MySQL).
// Each table is dropped individually because sqlc :exec only executes one statement.
// Tables that don't exist are skipped with a warning.
func (d MysqlDatabase) DropAllTables() error {
	queries := mdbm.New(d.Connection)

	ops := []dropOp{
		// Tier 7.5: Webhook system tables (deliveries before webhooks for FK)
		{"webhook_deliveries", func() error { return queries.DropWebhookDeliveryTable(d.Context) }},
		{"webhooks", func() error { return queries.DropWebhookTable(d.Context) }},
		// Tier 7: Plugin system tables (pipelines before plugins for FK)
		{"pipelines", func() error { return queries.DropPipelinesTable(d.Context) }},
		{"plugins", func() error { return queries.DropPluginsTable(d.Context) }},
		// Tier 6: Junction tables
		{"role_permissions", func() error { return queries.DropRolePermissionsTable(d.Context) }},
		// Tier 5.5: Content relation tables (depend on content_data and fields)
		{"admin_content_relations", func() error { return queries.DropAdminContentRelationTable(d.Context) }},
		{"content_relations", func() error { return queries.DropContentRelationTable(d.Context) }},
		// Tier 5.5b: Content version tables (depend on content_data + users)
		{"admin_content_versions", func() error { return queries.DropAdminContentVersionTable(d.Context) }},
		{"content_versions", func() error { return queries.DropContentVersionTable(d.Context) }},
		// Tier 5: Content field values
		{"admin_content_fields", func() error { return queries.DropAdminContentFieldTable(d.Context) }},
		{"content_fields", func() error { return queries.DropContentFieldTable(d.Context) }},
		// Tier 4: Content data tables
		{"admin_content_data", func() error { return queries.DropAdminContentDataTable(d.Context) }},
		{"content_data", func() error { return queries.DropContentDataTable(d.Context) }},
		// Tier 3: Field definition tables (field_plugin_config references fields)
		{"field_plugin_config", func() error { return queries.DropFieldPluginConfigTable(d.Context) }},
		{"admin_fields", func() error { return queries.DropAdminFieldTable(d.Context) }},
		{"fields", func() error { return queries.DropFieldTable(d.Context) }},
		// Tier 2.5: Validation tables (referenced by fields)
		{"admin_validations", func() error { return queries.DropAdminValidationTable(d.Context) }},
		{"validations", func() error { return queries.DropValidationTable(d.Context) }},
		// Tier 2: Core content tables and user-related tables
		{"admin_datatypes", func() error { return queries.DropAdminDatatypeTable(d.Context) }},
		{"datatypes", func() error { return queries.DropDatatypeTable(d.Context) }},
		{"routes", func() error { return queries.DropRouteTable(d.Context) }},
		{"admin_routes", func() error { return queries.DropAdminRouteTable(d.Context) }},
		{"media", func() error { return queries.DropMediaTable(d.Context) }},
		{"media_folders", func() error { return queries.DropMediaFolderTable(d.Context) }},
		{"admin_media", func() error { return queries.DropAdminMediaTable(d.Context) }},
		{"admin_media_folders", func() error { return queries.DropAdminMediaFolderTable(d.Context) }},
		{"tables", func() error { return queries.DropTableTable(d.Context) }},
		{"sessions", func() error { return queries.DropSessionTable(d.Context) }},
		{"user_ssh_keys", func() error { return queries.DropUserSshKeyTable(d.Context) }},
		{"user_oauth", func() error { return queries.DropUserOauthTable(d.Context) }},
		{"tokens", func() error { return queries.DropTokenTable(d.Context) }},
		// Tier 1: User management
		{"users", func() error { return queries.DropUserTable(d.Context) }},
		// Tier 0: Foundation tables
		{"media_dimensions", func() error { return queries.DropMediaDimensionTable(d.Context) }},
		{"field_types", func() error { return queries.DropFieldTypeTable(d.Context) }},
		{"admin_field_types", func() error { return queries.DropAdminFieldTypeTable(d.Context) }},
		{"locales", func() error { return queries.DropLocaleTable(d.Context) }},
		{"roles", func() error { return queries.DropRoleTable(d.Context) }},
		{"permissions", func() error { return queries.DropPermissionTable(d.Context) }},
		// Infrastructure tables
		{"backup_sets", func() error { return queries.DropBackupSetsTable(d.Context) }},
		{"backup_verifications", func() error { return queries.DropBackupVerificationsTable(d.Context) }},
		{"backups", func() error { return queries.DropBackupsTable(d.Context) }},
		{"change_events", func() error { return queries.DropChangeEventsTable(d.Context) }},
	}

	return runDropOps(ops)
}

// DropAllTables drops all database tables in reverse dependency order (PostgreSQL).
// Each table is dropped individually because sqlc :exec only executes one statement.
// Tables that don't exist are skipped with a warning.
func (d PsqlDatabase) DropAllTables() error {
	queries := mdbp.New(d.Connection)

	ops := []dropOp{
		// Tier 7.5: Webhook system tables (deliveries before webhooks for FK)
		{"webhook_deliveries", func() error { return queries.DropWebhookDeliveryTable(d.Context) }},
		{"webhooks", func() error { return queries.DropWebhookTable(d.Context) }},
		// Tier 7: Plugin system tables (pipelines before plugins for FK)
		{"pipelines", func() error { return queries.DropPipelinesTable(d.Context) }},
		{"plugins", func() error { return queries.DropPluginsTable(d.Context) }},
		// Tier 6: Junction tables
		{"role_permissions", func() error { return queries.DropRolePermissionsTable(d.Context) }},
		// Tier 5.5: Content relation tables (depend on content_data and fields)
		{"admin_content_relations", func() error { return queries.DropAdminContentRelationTable(d.Context) }},
		{"content_relations", func() error { return queries.DropContentRelationTable(d.Context) }},
		// Tier 5.5b: Content version tables (depend on content_data + users)
		{"admin_content_versions", func() error { return queries.DropAdminContentVersionTable(d.Context) }},
		{"content_versions", func() error { return queries.DropContentVersionTable(d.Context) }},
		// Tier 5: Content field values
		{"admin_content_fields", func() error { return queries.DropAdminContentFieldTable(d.Context) }},
		{"content_fields", func() error { return queries.DropContentFieldTable(d.Context) }},
		// Tier 4: Content data tables
		{"admin_content_data", func() error { return queries.DropAdminContentDataTable(d.Context) }},
		{"content_data", func() error { return queries.DropContentDataTable(d.Context) }},
		// Tier 3: Field definition tables (field_plugin_config references fields)
		{"field_plugin_config", func() error { return queries.DropFieldPluginConfigTable(d.Context) }},
		{"admin_fields", func() error { return queries.DropAdminFieldTable(d.Context) }},
		{"fields", func() error { return queries.DropFieldTable(d.Context) }},
		// Tier 2.5: Validation tables (referenced by fields)
		{"admin_validations", func() error { return queries.DropAdminValidationTable(d.Context) }},
		{"validations", func() error { return queries.DropValidationTable(d.Context) }},
		// Tier 2: Core content tables and user-related tables
		{"admin_datatypes", func() error { return queries.DropAdminDatatypeTable(d.Context) }},
		{"datatypes", func() error { return queries.DropDatatypeTable(d.Context) }},
		{"routes", func() error { return queries.DropRouteTable(d.Context) }},
		{"admin_routes", func() error { return queries.DropAdminRouteTable(d.Context) }},
		{"media", func() error { return queries.DropMediaTable(d.Context) }},
		{"media_folders", func() error { return queries.DropMediaFolderTable(d.Context) }},
		{"admin_media", func() error { return queries.DropAdminMediaTable(d.Context) }},
		{"admin_media_folders", func() error { return queries.DropAdminMediaFolderTable(d.Context) }},
		{"tables", func() error { return queries.DropTableTable(d.Context) }},
		{"sessions", func() error { return queries.DropSessionTable(d.Context) }},
		{"user_ssh_keys", func() error { return queries.DropUserSshKeyTable(d.Context) }},
		{"user_oauth", func() error { return queries.DropUserOauthTable(d.Context) }},
		{"tokens", func() error { return queries.DropTokenTable(d.Context) }},
		// Tier 1: User management
		{"users", func() error { return queries.DropUserTable(d.Context) }},
		// Tier 0: Foundation tables
		{"media_dimensions", func() error { return queries.DropMediaDimensionTable(d.Context) }},
		{"field_types", func() error { return queries.DropFieldTypeTable(d.Context) }},
		{"admin_field_types", func() error { return queries.DropAdminFieldTypeTable(d.Context) }},
		{"locales", func() error { return queries.DropLocaleTable(d.Context) }},
		{"roles", func() error { return queries.DropRoleTable(d.Context) }},
		{"permissions", func() error { return queries.DropPermissionTable(d.Context) }},
		// Infrastructure tables
		{"backup_sets", func() error { return queries.DropBackupSetsTable(d.Context) }},
		{"backup_verifications", func() error { return queries.DropBackupVerificationsTable(d.Context) }},
		{"backups", func() error { return queries.DropBackupsTable(d.Context) }},
		{"change_events", func() error { return queries.DropChangeEventsTable(d.Context) }},
	}

	return runDropOps(ops)
}

// runDropOps executes drop operations sequentially, logging warnings for failures
// and continuing to the next table. Returns a combined error if any drops failed.
func runDropOps(ops []dropOp) error {
	var failed []string
	for _, op := range ops {
		if err := op.fn(); err != nil {
			utility.DefaultLogger.Warn("failed to drop table "+op.name+", continuing", err)
			failed = append(failed, op.name)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to drop %d table(s): %s", len(failed), strings.Join(failed, ", "))
	}
	return nil
}
