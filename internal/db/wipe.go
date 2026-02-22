package db

import (
	"fmt"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
)

// DropAllTables drops all database tables in reverse dependency order (SQLite).
// Each table is dropped individually because sqlc :exec only executes one statement.
func (d Database) DropAllTables() error {
	queries := mdb.New(d.Connection)

	// Tier 6: Junction tables
	if err := queries.DropAdminDatatypesFieldsTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_datatypes_fields: %w", err)
	}
	if err := queries.DropDatatypesFieldsTable(d.Context); err != nil {
		return fmt.Errorf("drop datatypes_fields: %w", err)
	}
	if err := queries.DropRolePermissionsTable(d.Context); err != nil {
		return fmt.Errorf("drop role_permissions: %w", err)
	}

	// Tier 5.5: Content relation tables (depend on content_data and fields)
	if err := queries.DropAdminContentRelationTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_relations: %w", err)
	}
	if err := queries.DropContentRelationTable(d.Context); err != nil {
		return fmt.Errorf("drop content_relations: %w", err)
	}

	// Tier 5: Content field values
	if err := queries.DropAdminContentField(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_fields: %w", err)
	}
	if err := queries.DropContentFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop content_fields: %w", err)
	}

	// Tier 4: Content data tables
	if err := queries.DropAdminContentData(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_data: %w", err)
	}
	if err := queries.DropContentDataTable(d.Context); err != nil {
		return fmt.Errorf("drop content_data: %w", err)
	}

	// Tier 3: Field definition tables
	if err := queries.DropAdminFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_fields: %w", err)
	}
	if err := queries.DropFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop fields: %w", err)
	}

	// Tier 2: Core content tables and user-related tables
	if err := queries.DropAdminDatatypeTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_datatypes: %w", err)
	}
	if err := queries.DropDatatypeTable(d.Context); err != nil {
		return fmt.Errorf("drop datatypes: %w", err)
	}
	if err := queries.DropRouteTable(d.Context); err != nil {
		return fmt.Errorf("drop routes: %w", err)
	}
	if err := queries.DropAdminRouteTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_routes: %w", err)
	}
	if err := queries.DropMediaTable(d.Context); err != nil {
		return fmt.Errorf("drop media: %w", err)
	}
	if err := queries.DropTableTable(d.Context); err != nil {
		return fmt.Errorf("drop tables: %w", err)
	}
	if err := queries.DropSessionTable(d.Context); err != nil {
		return fmt.Errorf("drop sessions: %w", err)
	}
	if err := queries.DropUserSshKeyTable(d.Context); err != nil {
		return fmt.Errorf("drop user_ssh_keys: %w", err)
	}
	if err := queries.DropUserOauthTable(d.Context); err != nil {
		return fmt.Errorf("drop user_oauth: %w", err)
	}
	if err := queries.DropTokenTable(d.Context); err != nil {
		return fmt.Errorf("drop tokens: %w", err)
	}

	// Tier 1: User management
	if err := queries.DropUserTable(d.Context); err != nil {
		return fmt.Errorf("drop users: %w", err)
	}

	// Tier 0: Foundation tables
	if err := queries.DropMediaDimensionTable(d.Context); err != nil {
		return fmt.Errorf("drop media_dimensions: %w", err)
	}
	if err := queries.DropFieldTypesTable(d.Context); err != nil {
		return fmt.Errorf("drop field_types: %w", err)
	}
	if err := queries.DropAdminFieldTypesTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_field_types: %w", err)
	}
	if err := queries.DropRoleTable(d.Context); err != nil {
		return fmt.Errorf("drop roles: %w", err)
	}
	if err := queries.DropPermissionTable(d.Context); err != nil {
		return fmt.Errorf("drop permissions: %w", err)
	}

	// Infrastructure tables
	if err := queries.DropBackupSetsTable(d.Context); err != nil {
		return fmt.Errorf("drop backup_sets: %w", err)
	}
	if err := queries.DropBackupVerificationsTable(d.Context); err != nil {
		return fmt.Errorf("drop backup_verifications: %w", err)
	}
	if err := queries.DropBackupsTable(d.Context); err != nil {
		return fmt.Errorf("drop backups: %w", err)
	}
	if err := queries.DropChangeEventsTable(d.Context); err != nil {
		return fmt.Errorf("drop change_events: %w", err)
	}

	return nil
}

// DropAllTables drops all database tables in reverse dependency order (MySQL).
// Each table is dropped individually because sqlc :exec only executes one statement.
func (d MysqlDatabase) DropAllTables() error {
	queries := mdbm.New(d.Connection)

	// Tier 6: Junction tables
	if err := queries.DropAdminDatatypesFieldsTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_datatypes_fields: %w", err)
	}
	if err := queries.DropDatatypesFieldsTable(d.Context); err != nil {
		return fmt.Errorf("drop datatypes_fields: %w", err)
	}
	if err := queries.DropRolePermissionsTable(d.Context); err != nil {
		return fmt.Errorf("drop role_permissions: %w", err)
	}

	// Tier 5.5: Content relation tables (depend on content_data and fields)
	if err := queries.DropAdminContentRelationTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_relations: %w", err)
	}
	if err := queries.DropContentRelationTable(d.Context); err != nil {
		return fmt.Errorf("drop content_relations: %w", err)
	}

	// Tier 5: Content field values
	if err := queries.DropAdminContentFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_fields: %w", err)
	}
	if err := queries.DropContentFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop content_fields: %w", err)
	}

	// Tier 4: Content data tables
	if err := queries.DropAdminContentData(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_data: %w", err)
	}
	if err := queries.DropContentDataTable(d.Context); err != nil {
		return fmt.Errorf("drop content_data: %w", err)
	}

	// Tier 3: Field definition tables
	if err := queries.DropAdminFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_fields: %w", err)
	}
	if err := queries.DropFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop fields: %w", err)
	}

	// Tier 2: Core content tables and user-related tables
	if err := queries.DropAdminDatatypeTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_datatypes: %w", err)
	}
	if err := queries.DropDatatypeTable(d.Context); err != nil {
		return fmt.Errorf("drop datatypes: %w", err)
	}
	if err := queries.DropRouteTable(d.Context); err != nil {
		return fmt.Errorf("drop routes: %w", err)
	}
	if err := queries.DropAdminRouteTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_routes: %w", err)
	}
	if err := queries.DropMediaTable(d.Context); err != nil {
		return fmt.Errorf("drop media: %w", err)
	}
	if err := queries.DropTableTable(d.Context); err != nil {
		return fmt.Errorf("drop tables: %w", err)
	}
	if err := queries.DropSessionTable(d.Context); err != nil {
		return fmt.Errorf("drop sessions: %w", err)
	}
	if err := queries.DropUserSshKeyTable(d.Context); err != nil {
		return fmt.Errorf("drop user_ssh_keys: %w", err)
	}
	if err := queries.DropUserOauthTable(d.Context); err != nil {
		return fmt.Errorf("drop user_oauth: %w", err)
	}
	if err := queries.DropTokenTable(d.Context); err != nil {
		return fmt.Errorf("drop tokens: %w", err)
	}

	// Tier 1: User management
	if err := queries.DropUserTable(d.Context); err != nil {
		return fmt.Errorf("drop users: %w", err)
	}

	// Tier 0: Foundation tables
	if err := queries.DropMediaDimensionTable(d.Context); err != nil {
		return fmt.Errorf("drop media_dimensions: %w", err)
	}
	if err := queries.DropFieldTypesTable(d.Context); err != nil {
		return fmt.Errorf("drop field_types: %w", err)
	}
	if err := queries.DropAdminFieldTypesTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_field_types: %w", err)
	}
	if err := queries.DropRoleTable(d.Context); err != nil {
		return fmt.Errorf("drop roles: %w", err)
	}
	if err := queries.DropPermissionTable(d.Context); err != nil {
		return fmt.Errorf("drop permissions: %w", err)
	}

	// Infrastructure tables
	if err := queries.DropBackupSetsTable(d.Context); err != nil {
		return fmt.Errorf("drop backup_sets: %w", err)
	}
	if err := queries.DropBackupVerificationsTable(d.Context); err != nil {
		return fmt.Errorf("drop backup_verifications: %w", err)
	}
	if err := queries.DropBackupsTable(d.Context); err != nil {
		return fmt.Errorf("drop backups: %w", err)
	}
	if err := queries.DropChangeEventsTable(d.Context); err != nil {
		return fmt.Errorf("drop change_events: %w", err)
	}

	return nil
}

// DropAllTables drops all database tables in reverse dependency order (PostgreSQL).
// Each table is dropped individually because sqlc :exec only executes one statement.
func (d PsqlDatabase) DropAllTables() error {
	queries := mdbp.New(d.Connection)

	// Tier 6: Junction tables
	if err := queries.DropAdminDatatypesFieldsTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_datatypes_fields: %w", err)
	}
	if err := queries.DropDatatypesFieldsTable(d.Context); err != nil {
		return fmt.Errorf("drop datatypes_fields: %w", err)
	}
	if err := queries.DropRolePermissionsTable(d.Context); err != nil {
		return fmt.Errorf("drop role_permissions: %w", err)
	}

	// Tier 5.5: Content relation tables (depend on content_data and fields)
	if err := queries.DropAdminContentRelationTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_relations: %w", err)
	}
	if err := queries.DropContentRelationTable(d.Context); err != nil {
		return fmt.Errorf("drop content_relations: %w", err)
	}

	// Tier 5: Content field values
	if err := queries.DropAdminContentFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_fields: %w", err)
	}
	if err := queries.DropContentFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop content_fields: %w", err)
	}

	// Tier 4: Content data tables
	if err := queries.DropAdminContentDataTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_content_data: %w", err)
	}
	if err := queries.DropContentDataTable(d.Context); err != nil {
		return fmt.Errorf("drop content_data: %w", err)
	}

	// Tier 3: Field definition tables
	if err := queries.DropAdminFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_fields: %w", err)
	}
	if err := queries.DropFieldTable(d.Context); err != nil {
		return fmt.Errorf("drop fields: %w", err)
	}

	// Tier 2: Core content tables and user-related tables
	if err := queries.DropAdminDatatypeTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_datatypes: %w", err)
	}
	if err := queries.DropDatatypeTable(d.Context); err != nil {
		return fmt.Errorf("drop datatypes: %w", err)
	}
	if err := queries.DropRouteTable(d.Context); err != nil {
		return fmt.Errorf("drop routes: %w", err)
	}
	if err := queries.DropAdminRouteTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_routes: %w", err)
	}
	if err := queries.DropMediaTable(d.Context); err != nil {
		return fmt.Errorf("drop media: %w", err)
	}
	if err := queries.DropTableTable(d.Context); err != nil {
		return fmt.Errorf("drop tables: %w", err)
	}
	if err := queries.DropSessionTable(d.Context); err != nil {
		return fmt.Errorf("drop sessions: %w", err)
	}
	if err := queries.DropUserSshKeyTable(d.Context); err != nil {
		return fmt.Errorf("drop user_ssh_keys: %w", err)
	}
	if err := queries.DropUserOauthTable(d.Context); err != nil {
		return fmt.Errorf("drop user_oauth: %w", err)
	}
	if err := queries.DropTokenTable(d.Context); err != nil {
		return fmt.Errorf("drop tokens: %w", err)
	}

	// Tier 1: User management
	if err := queries.DropUserTable(d.Context); err != nil {
		return fmt.Errorf("drop users: %w", err)
	}

	// Tier 0: Foundation tables
	if err := queries.DropMediaDimensionTable(d.Context); err != nil {
		return fmt.Errorf("drop media_dimensions: %w", err)
	}
	if err := queries.DropFieldTypesTable(d.Context); err != nil {
		return fmt.Errorf("drop field_types: %w", err)
	}
	if err := queries.DropAdminFieldTypesTable(d.Context); err != nil {
		return fmt.Errorf("drop admin_field_types: %w", err)
	}
	if err := queries.DropRoleTable(d.Context); err != nil {
		return fmt.Errorf("drop roles: %w", err)
	}
	if err := queries.DropPermissionTable(d.Context); err != nil {
		return fmt.Errorf("drop permissions: %w", err)
	}

	// Infrastructure tables
	if err := queries.DropBackupSetsTable(d.Context); err != nil {
		return fmt.Errorf("drop backup_sets: %w", err)
	}
	if err := queries.DropBackupVerificationsTable(d.Context); err != nil {
		return fmt.Errorf("drop backup_verifications: %w", err)
	}
	if err := queries.DropBackupsTable(d.Context); err != nil {
		return fmt.Errorf("drop backups: %w", err)
	}
	if err := queries.DropChangeEventsTable(d.Context); err != nil {
		return fmt.Errorf("drop change_events: %w", err)
	}

	return nil
}
