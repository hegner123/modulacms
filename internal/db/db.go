// Package db provides a multi-database abstraction layer for Modula supporting SQLite, MySQL, and PostgreSQL.
// It defines the DbDriver interface and wrapper types that convert sqlc-generated types to application-level types with custom validation and auditing.
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	utility "github.com/hegner123/modulacms/internal/utility"
)

//go:embed sql
var sqlFiles embed.FS

// Historied defines the interface for entities that track their modification history.
type Historied interface {
	GetHistory() string
	MapHistoryEntry() string
	UpdateHistory([]byte) error
}

// Database represents a SQLite database connection
type Database struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

// MysqlDatabase represents a MySQL database connection
type MysqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

// PsqlDatabase represents a PostgreSQL database connection
type PsqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

// DbStatus represents the status of a database connection
type DbStatus string

// Database connection status constants.
const (
	Open   DbStatus = "open"
	Closed DbStatus = "closed"
	Err    DbStatus = "error"
)

// DbDriver is the interface for all database drivers. It composes 22 focused
// repository interfaces defined in repositories.go. Consumers that need only
// a subset of methods should accept the narrower repository interface instead.
type DbDriver interface {
	SchemaRepository
	ConnectionRepository
	ContentDataRepository
	ContentFieldRepository
	AdminContentDataRepository
	AdminContentFieldRepository
	DatatypeRepository
	AdminDatatypeRepository
	FieldRepository
	AdminFieldRepository
	RouteRepository
	AdminRouteRepository
	MediaRepository
	UserRepository
	AuthRepository
	RBACRepository
	BackupRepository
	ChangeEventRepository
	TableRepository
	PluginRepository
	LocaleRepository
	WebhookRepository
	MediaFolderRepository
	FieldPluginConfigRepository
}

// GetConnection returns the database connection and context
func (d Database) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

// GetConnection returns the MySQL database connection and context
func (d MysqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

// GetConnection returns the PostgreSQL database connection and context
func (d PsqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

// Ping checks if the database connection is still alive
func (d Database) Ping() error {
	if d.Connection == nil {
		return fmt.Errorf("SQLite connection not established")
	}
	return d.Connection.Ping()
}

// Ping checks if the MySQL database connection is still alive
func (d MysqlDatabase) Ping() error {
	if d.Connection == nil {
		return fmt.Errorf("MySQL connection not established")
	}
	return d.Connection.Ping()
}

// Ping checks if the PostgreSQL database connection is still alive
func (d PsqlDatabase) Ping() error {
	if d.Connection == nil {
		return fmt.Errorf("PostgreSQL connection not established")
	}
	return d.Connection.Ping()
}

// ExecuteQuery executes a raw SQL query on the database
func (d Database) ExecuteQuery(query string, table DBTable) (*sql.Rows, error) {
	q := fmt.Sprintf("%s %s;", query, DBTableString(table))
	return d.Connection.Query(q)
}

// ExecuteQuery executes a raw SQL query on the MySQL database
func (d MysqlDatabase) ExecuteQuery(query string, table DBTable) (*sql.Rows, error) {
	q := fmt.Sprintf("%s %s;", query, DBTableString(table))
	return d.Connection.Query(q)
}

// ExecuteQuery executes a raw SQL query on the PostgreSQL database
func (d PsqlDatabase) ExecuteQuery(query string, table DBTable) (*sql.Rows, error) {
	q := fmt.Sprintf("%s %s;", query, DBTableString(table))
	return d.Connection.Query(q)
}

// CreateAllTables creates all database tables
func (d Database) CreateAllTables() error {
	// CRITICAL: Tables must be created in dependency order to satisfy FK constraints
	// See: ai/reference/TABLE_CREATION_ORDER.md for complete dependency graph

	// Infrastructure tables (no dependencies on application tables)
	err := d.CreateChangeEventsTable()
	if err != nil {
		return err
	}

	err = d.CreateBackupTables()
	if err != nil {
		return err
	}

	// Tier 0: Foundation tables (no dependencies)
	err = d.CreatePermissionTable()
	if err != nil {
		return err
	}

	err = d.CreateRoleTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaDimensionTable()
	if err != nil {
		return err
	}

	err = d.CreateFieldTypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTypeTable()
	if err != nil {
		return err
	}

	err = d.CreateLocaleTable()
	if err != nil {
		return err
	}

	// Tier 1: User management (depends on roles)
	err = d.CreateUserTable()
	if err != nil {
		return err
	}

	// Tier 2: User-related tables and core content tables (depend on users)
	err = d.CreateTokenTable()
	if err != nil {
		return err
	}

	err = d.CreateUserOauthTable()
	if err != nil {
		return err
	}

	err = d.CreateUserSshKeyTable()
	if err != nil {
		return err
	}

	err = d.CreateSessionTable()
	if err != nil {
		return err
	}

	err = d.CreateTableTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaFolderTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeTable()
	if err != nil {
		return err
	}

	// Tier 3: Field definition tables (depend on datatypes)
	err = d.CreateFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTable()
	if err != nil {
		return err
	}

	// Tier 4: Content data tables (depend on routes, datatypes, users)
	err = d.CreateContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentDataTable()
	if err != nil {
		return err
	}

	// Tier 5: Content field values (depend on content_data and fields)
	err = d.CreateContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentFieldTable()
	if err != nil {
		return err
	}

	// Tier 5.5: Content relation tables (depend on content_data and fields)
	err = d.CreateContentRelationTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentRelationTable()
	if err != nil {
		return err
	}

	// Tier 5.5b: Content version tables (depend on content_data + users)
	err = d.CreateContentVersionTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentVersionTable()
	if err != nil {
		return err
	}

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateRolePermissionsTable()
	if err != nil {
		return err
	}

	// Tier 7: Plugin system tables (plugins before pipelines for FK)
	err = d.CreatePluginTable()
	if err != nil {
		return err
	}

	err = d.CreatePipelineTable()
	if err != nil {
		return err
	}

	// Tier 7.5: Webhook system tables (webhooks before deliveries for FK)
	err = d.CreateWebhookTable()
	if err != nil {
		return err
	}

	err = d.CreateWebhookDeliveryTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 28 tables to verify successful table creation.
// If any table failed to create, this function will catch it immediately during install rather than later during operation.
func (d Database) CreateBootstrapData(adminHash string) error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap", "system")

	// 1. Create system admin permission (permission_id = 1)
	permission, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
		Label:           "admin",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create system admin permission: %w", err)
	}
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 1b. Create RBAC system permissions
	rbacPermissionLabels := []string{
		"content:read", "content:create", "content:update", "content:delete", "content:publish", "content:admin",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete", "datatypes:admin",
		"fields:read", "fields:create", "fields:update", "fields:delete", "fields:admin",
		"media:read", "media:create", "media:update", "media:delete", "media:admin",
		"routes:read", "routes:create", "routes:update", "routes:delete", "routes:admin",
		"users:read", "users:create", "users:update", "users:delete", "users:admin",
		"roles:read", "roles:create", "roles:update", "roles:delete", "roles:admin",
		"permissions:read", "permissions:create", "permissions:update", "permissions:delete", "permissions:admin",
		"sessions:read", "sessions:delete", "sessions:admin",
		"ssh_keys:read", "ssh_keys:create", "ssh_keys:delete", "ssh_keys:admin",
		"config:read", "config:update", "config:admin",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete", "admin_tree:admin",
		"field_types:read", "field_types:create", "field_types:update", "field_types:delete", "field_types:admin",
		"admin_field_types:read", "admin_field_types:create", "admin_field_types:update", "admin_field_types:delete", "admin_field_types:admin",
		"deploy:read", "deploy:create",
		"webhook:read", "webhook:create", "webhook:update", "webhook:delete", "webhook:admin",
		"plugins:read", "plugins:admin",
		"tables:read", "tables:create", "tables:update", "tables:delete", "tables:admin",
		"import:read", "import:create", "import:admin",
		"tokens:read", "tokens:create", "tokens:delete", "tokens:admin",
		"locale:read", "locale:create", "locale:update", "locale:delete", "locale:admin",
		"audit:read", "audit:admin",
		"backup:read", "backup:create", "backup:update", "backup:delete", "backup:admin",
		"search:read", "search:update", "search:admin",
	}
	rbacPermissions := make(map[string]types.PermissionID, len(rbacPermissionLabels))
	for _, label := range rbacPermissionLabels {
		p, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
			Label:           label,
			SystemProtected: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create RBAC permission %q: %w", label, err)
		}
		rbacPermissions[label] = p.PermissionID
	}

	// 2. Create system admin role (role_id = 1)
	adminRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "admin",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create system admin role: %w", err)
	}
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 2b. Create editor role
	editorRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "editor",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create editor role: %w", err)
	}
	if editorRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create editor role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "viewer",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create viewer role: %w", err)
	}
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 3b. Create role_permissions junction rows
	// Admin gets all permissions
	for _, permID := range rbacPermissions {
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       adminRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create admin role_permission: %w", err)
		}
	}

	// Editor permissions: CRUD on content, datatypes, fields, media, routes, admin_tree, field_types, admin_field_types; read-only on users, sessions, ssh_keys
	editorPermLabels := []string{
		"content:read", "content:create", "content:update", "content:delete",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete",
		"fields:read", "fields:create", "fields:update", "fields:delete",
		"media:read", "media:create", "media:update", "media:delete",
		"routes:read", "routes:create", "routes:update", "routes:delete",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete",
		"field_types:read", "field_types:create", "field_types:update", "field_types:delete",
		"admin_field_types:read", "admin_field_types:create", "admin_field_types:update", "admin_field_types:delete",
		"users:read",
		"sessions:read",
		"ssh_keys:read",
		"config:read",
		"webhook:read", "webhook:create", "webhook:update", "webhook:delete",
		"tokens:read", "tokens:create", "tokens:delete",
		"locale:read", "locale:create", "locale:update", "locale:delete",
		"audit:read",
		"import:read",
		"plugins:read",
		"backup:read",
		"search:read",
	}
	for _, label := range editorPermLabels {
		permID, ok := rbacPermissions[label]
		if !ok {
			return fmt.Errorf("missing RBAC permission %q for editor role", label)
		}
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       editorRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create editor role_permission for %q: %w", label, err)
		}
	}

	// Viewer permissions: read-only on content, media, routes, field_types, admin_field_types
	viewerPermLabels := []string{
		"content:read",
		"media:read",
		"routes:read",
		"field_types:read",
		"admin_field_types:read",
	}
	for _, label := range viewerPermLabels {
		permID, ok := rbacPermissions[label]
		if !ok {
			return fmt.Errorf("missing RBAC permission %q for viewer role", label)
		}
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       viewerRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create viewer role_permission for %q: %w", label, err)
		}
	}

	// 4. Create system user with well-known SystemUserID
	systemUser, err := d.CreateSystemUser(CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modula.local"),
		Hash:         adminHash,
		Role:         adminRole.RoleID.String(),
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create system user: %v", err)
	}
	if systemUser.UserID.IsZero() {
		return fmt.Errorf("failed to create system user: user_id is empty")
	}

	// 5. Create default home route (route_id = 1) - Recommended
	homeRoute, err := d.CreateRoute(ctx, ac, CreateRouteParams{
		Slug:         types.Slug("/"),
		Title:        "Home",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default home route: %w", err)
	}
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		Name:         "page",
		SortOrder:    0,
		Label:        "Page",
		Type:         string(types.DatatypeTypeRoot),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default page datatype: %w", err)
	}
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 6b. Create _reference system datatype
	refDatatype, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		SortOrder:    1,
		Label:        "Reference",
		Type:         string(types.DatatypeTypeReference),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create _reference datatype: %w", err)
	}
	if refDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create _reference datatype")
	}

	// 6c. Create "Target" field for _reference datatype (linked via parent_id)
	refField, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: refDatatype.DatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Target",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeIDRef,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create _reference Target field: %w", err)
	}
	if refField.FieldID.IsZero() {
		return fmt.Errorf("failed to create _reference Target field")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute, err := d.CreateAdminRoute(ctx, ac, CreateAdminRouteParams{
		Slug:         types.Slug("admin"),
		Title:        "Admin",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin route: %w", err)
	}
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype, err := d.CreateAdminDatatype(ctx, ac, CreateAdminDatatypeParams{
		ParentID:     types.NullableAdminDatatypeID{},
		SortOrder:    0,
		Label:        "Admin Page",
		Type:         string(types.DatatypeTypeRoot),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin datatype: %w", err)
	}
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1) -- linked to admin datatype via parent_id
	adminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{ID: adminDatatype.AdminDatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Content",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin field: %w", err)
	}
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1) -- linked to page datatype via parent_id
	field, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: pageDatatype.DatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Content",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default field: %w", err)
	}
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       types.NullableRouteID{Valid: true, ID: homeRoute.RouteID},
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		AuthorID:      systemUser.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default content_data: %w", err)
	}
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 11.5. Set root_id to self for the root content node
	contentRootID := types.NullableContentID{Valid: true, ID: contentData.ContentDataID}
	_, err = d.UpdateContentData(ctx, ac, UpdateContentDataParams{
		ContentDataID: contentData.ContentDataID,
		RootID:        contentRootID,
		RouteID:       contentData.RouteID,
		ParentID:      contentData.ParentID,
		FirstChildID:  contentData.FirstChildID,
		NextSiblingID: contentData.NextSiblingID,
		PrevSiblingID: contentData.PrevSiblingID,
		DatatypeID:    contentData.DatatypeID,
		AuthorID:      contentData.AuthorID,
		Status:        contentData.Status,
		DateCreated:   contentData.DateCreated,
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to set root_id on default content_data: %w", err)
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{
		ParentID:        types.NullableAdminContentID{},
		FirstChildID:    types.NullableAdminContentID{},
		NextSiblingID:   types.NullableAdminContentID{},
		PrevSiblingID:   types.NullableAdminContentID{},
		AdminRouteID:    types.NullableAdminRouteID{Valid: true, ID: adminRoute.AdminRouteID},
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AuthorID:        systemUser.UserID,
		Status:          types.ContentStatusDraft,
		DateCreated:     types.TimestampNow(),
		DateModified:    types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_content_data: %w", err)
	}
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 12.5. Set root_id to self for the root admin content node
	adminContentRootID := types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID}
	_, err = d.UpdateAdminContentData(ctx, ac, UpdateAdminContentDataParams{
		AdminContentDataID: adminContentData.AdminContentDataID,
		RootID:             adminContentRootID,
		AdminRouteID:       adminContentData.AdminRouteID,
		ParentID:           adminContentData.ParentID,
		FirstChildID:       adminContentData.FirstChildID,
		NextSiblingID:      adminContentData.NextSiblingID,
		PrevSiblingID:      adminContentData.PrevSiblingID,
		AdminDatatypeID:    adminContentData.AdminDatatypeID,
		AuthorID:           adminContentData.AuthorID,
		Status:             adminContentData.Status,
		DateCreated:        adminContentData.DateCreated,
		DateModified:       types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to set root_id on default admin_content_data: %w", err)
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		RootID:        contentRootID,
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      systemUser.UserID,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default content_field: %w", err)
	}
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField, err := d.CreateAdminContentField(ctx, ac, CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{},
		RootID:             adminContentRootID,
		AdminContentDataID: types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID},
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           systemUser.UserID,
		DateCreated:        types.TimestampNow(),
		DateModified:       types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_content_field: %w", err)
	}
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension, err := d.CreateMediaDimension(ctx, ac, CreateMediaDimensionParams{
		Label:       NewNullString("Default"),
		Width:       types.NewNullableInt64(1920),
		Height:      types.NewNullableInt64(1080),
		AspectRatio: NewNullString("16:9"),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media_dimension: %w", err)
	}
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         NewNullString("default"),
		DisplayName:  NewNullString("Default Media"),
		Alt:          NewNullString("Default"),
		Caption:      NullString{},
		Description:  NullString{},
		Class:        NullString{},
		URL:          types.URL("https://placeholder.local/default"),
		Mimetype:     NullString{},
		Dimensions:   NullString{},
		Srcset:       NullString{},
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media: %w", err)
	}
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token, err := d.CreateToken(ctx, ac, CreateTokenParams{
		UserID:    types.NullableUserID{Valid: true, ID: systemUser.UserID},
		TokenType: "validation",
		Token:     utility.HashToken("bootstrap_validation_token"),
		IssuedAt:  types.TimestampNow(),
		ExpiresAt: types.TimestampNow(),
		Revoked:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create default token: %w", err)
	}
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(ctx, ac, CreateSessionParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated: types.TimestampNow(),
		ExpiresAt:   types.TimestampNow(),
		LastAccess:  types.TimestampNow(),
		IpAddress:   NewNullString("127.0.0.1"),
		UserAgent:   NewNullString("bootstrap"),
		SessionData: NullString{},
	})
	if err != nil {
		return fmt.Errorf("failed to create default session: %v", err)
	}
	if session.SessionID.IsZero() {
		return fmt.Errorf("failed to create default session: session_id is 0")
	}

	// 19. Create default user_oauth record (user_oauth_id = 1) - Validation record
	userOauth, err := d.CreateUserOauth(ctx, ac, CreateUserOauthParams{
		UserID:              types.NullableUserID{Valid: true, ID: systemUser.UserID},
		OauthProvider:       "bootstrap",
		OauthProviderUserID: "bootstrap_user",
		AccessToken:         "bootstrap_access_token",
		RefreshToken:        "bootstrap_refresh_token",
		TokenExpiresAt:      types.TimestampNow(),
		DateCreated:         types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default user_oauth: %v", err)
	}
	if userOauth.UserOauthID.IsZero() {
		return fmt.Errorf("failed to create default user_oauth: user_oauth_id is 0")
	}

	// 19A. Create default user_ssh_key record (ssh_key_id = 1) - Validation record
	userSshKey, err := d.CreateUserSshKey(ctx, ac, CreateUserSshKeyParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBootstrapValidationKey bootstrap@modula",
		KeyType:     "ssh-ed25519",
		Fingerprint: "SHA256:bootstrap_validation_fingerprint",
		Label:       "Bootstrap Validation Key",
		DateCreated: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default user_ssh_key: %v", err)
	}
	if userSshKey.SshKeyID == "" {
		return fmt.Errorf("failed to create default user_ssh_key: ssh_key_id is empty")
	}

	// 20. Seed field_types and admin_field_types with the built-in field types
	fieldTypeSeedData := []struct{ Type, Label string }{
		{"text", "Text Input"}, {"textarea", "Text Area"}, {"number", "Number"},
		{"date", "Date"}, {"datetime", "Date & Time"}, {"boolean", "Boolean"},
		{"select", "Select"}, {"media", "Media"},
		{"json", "JSON"}, {"richtext", "Rich Text"}, {"slug", "Slug"},
		{"email", "Email"}, {"url", "URL"},
		{"_id", "ID Reference"}, {"_title", "Title"},
	}
	for _, ft := range fieldTypeSeedData {
		created, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{Type: ft.Type, Label: ft.Label})
		if err != nil {
			return fmt.Errorf("failed to seed field_type %q: %w", ft.Type, err)
		}
		if created.FieldTypeID.IsZero() {
			return fmt.Errorf("failed to seed field_type %q: id is zero", ft.Type)
		}
	}
	for _, ft := range fieldTypeSeedData {
		created, err := d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{Type: ft.Type, Label: ft.Label})
		if err != nil {
			return fmt.Errorf("failed to seed admin_field_type %q: %w", ft.Type, err)
		}
		if created.AdminFieldTypeID.IsZero() {
			return fmt.Errorf("failed to seed admin_field_type %q: id is zero", ft.Type)
		}
	}

	// 21. Register all 29 Modula tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
		"change_events",
		"backups",
		"backup_verifications",
		"backup_sets",
		"permissions",
		"roles",
		"media_dimensions",
		"media_folders",
		"field_types",
		"admin_field_types",
		"users",
		"tokens",
		"sessions",
		"routes",
		"media",
		"tables",
		"datatypes",
		"fields",
		"admin_fields",
		"content_data",
		"admin_content_data",
		"content_fields",
		"admin_content_fields",
		"admin_routes",
		"admin_datatypes",
		"user_oauth",
		"user_ssh_keys",
	}

	for _, tableName := range tableNames {
		table, err := d.CreateTable(ctx, ac, CreateTableParams{Label: tableName})
		if err != nil {
			return fmt.Errorf("failed to register table in tables registry: %s: %w", tableName, err)
		}
		if table.ID == "" {
			return fmt.Errorf("failed to register table in tables registry: %s", tableName)
		}
	}

	utility.DefaultLogger.Finfo("Bootstrap data created successfully: all tables validated with bootstrap records + complete table registry populated")
	return nil
}

// CleanupBootstrapData removes verification-only records after
// ValidateBootstrapData has confirmed all tables are working. These records
// exist solely to prove inserts succeed during install — they are not needed
// for CMS operation. Records that persist: permissions, roles, role_permissions,
// system user, field_types, admin_field_types, tables registry.
func (d Database) CleanupBootstrapData() error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap-cleanup", "system")

	// Public content tables (dependency order: leaf → root)

	contentFields, err := d.ListContentFields()
	if err != nil {
		return fmt.Errorf("cleanup: list content_fields: %w", err)
	}
	for _, cf := range *contentFields {
		if err := d.DeleteContentField(ctx, ac, cf.ContentFieldID); err != nil {
			return fmt.Errorf("cleanup: delete content_field %s: %w", cf.ContentFieldID, err)
		}
	}

	contentData, err := d.ListContentData()
	if err != nil {
		return fmt.Errorf("cleanup: list content_data: %w", err)
	}
	for _, cd := range *contentData {
		if err := d.DeleteContentData(ctx, ac, cd.ContentDataID); err != nil {
			return fmt.Errorf("cleanup: delete content_data %s: %w", cd.ContentDataID, err)
		}
	}

	fields, err := d.ListFields()
	if err != nil {
		return fmt.Errorf("cleanup: list fields: %w", err)
	}
	for _, f := range *fields {
		if err := d.DeleteField(ctx, ac, f.FieldID); err != nil {
			return fmt.Errorf("cleanup: delete field %s: %w", f.FieldID, err)
		}
	}

	datatypes, err := d.ListDatatypes()
	if err != nil {
		return fmt.Errorf("cleanup: list datatypes: %w", err)
	}
	for _, dt := range *datatypes {
		if err := d.DeleteDatatype(ctx, ac, dt.DatatypeID); err != nil {
			return fmt.Errorf("cleanup: delete datatype %s: %w", dt.DatatypeID, err)
		}
	}

	routes, err := d.ListRoutes()
	if err != nil {
		return fmt.Errorf("cleanup: list routes: %w", err)
	}
	for _, r := range *routes {
		if err := d.DeleteRoute(ctx, ac, r.RouteID); err != nil {
			return fmt.Errorf("cleanup: delete route %s: %w", r.RouteID, err)
		}
	}

	// Admin content tables (same dependency order)

	adminContentFields, err := d.ListAdminContentFields()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_content_fields: %w", err)
	}
	for _, acf := range *adminContentFields {
		if err := d.DeleteAdminContentField(ctx, ac, acf.AdminContentFieldID); err != nil {
			return fmt.Errorf("cleanup: delete admin_content_field %s: %w", acf.AdminContentFieldID, err)
		}
	}

	adminContentData, err := d.ListAdminContentData()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_content_data: %w", err)
	}
	for _, acd := range *adminContentData {
		if err := d.DeleteAdminContentData(ctx, ac, acd.AdminContentDataID); err != nil {
			return fmt.Errorf("cleanup: delete admin_content_data %s: %w", acd.AdminContentDataID, err)
		}
	}

	adminFields, err := d.ListAdminFields()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_fields: %w", err)
	}
	for _, af := range *adminFields {
		if err := d.DeleteAdminField(ctx, ac, af.AdminFieldID); err != nil {
			return fmt.Errorf("cleanup: delete admin_field %s: %w", af.AdminFieldID, err)
		}
	}

	adminDatatypes, err := d.ListAdminDatatypes()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_datatypes: %w", err)
	}
	for _, adt := range *adminDatatypes {
		if err := d.DeleteAdminDatatype(ctx, ac, adt.AdminDatatypeID); err != nil {
			return fmt.Errorf("cleanup: delete admin_datatype %s: %w", adt.AdminDatatypeID, err)
		}
	}

	adminRoutes, err := d.ListAdminRoutes()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_routes: %w", err)
	}
	for _, ar := range *adminRoutes {
		if err := d.DeleteAdminRoute(ctx, ac, ar.AdminRouteID); err != nil {
			return fmt.Errorf("cleanup: delete admin_route %s: %w", ar.AdminRouteID, err)
		}
	}

	// Media tables

	media, err := d.ListMedia()
	if err != nil {
		return fmt.Errorf("cleanup: list media: %w", err)
	}
	for _, m := range *media {
		if err := d.DeleteMedia(ctx, ac, m.MediaID); err != nil {
			return fmt.Errorf("cleanup: delete media %s: %w", m.MediaID, err)
		}
	}

	mediaDimensions, err := d.ListMediaDimensions()
	if err != nil {
		return fmt.Errorf("cleanup: list media_dimensions: %w", err)
	}
	for _, md := range *mediaDimensions {
		if err := d.DeleteMediaDimension(ctx, ac, md.MdID); err != nil {
			return fmt.Errorf("cleanup: delete media_dimension %s: %w", md.MdID, err)
		}
	}

	// Auth verification records (token, session, user_oauth, user_ssh_key)

	tokens, err := d.ListTokens()
	if err != nil {
		return fmt.Errorf("cleanup: list tokens: %w", err)
	}
	for _, tk := range *tokens {
		if err := d.DeleteToken(ctx, ac, tk.ID); err != nil {
			return fmt.Errorf("cleanup: delete token %s: %w", tk.ID, err)
		}
	}

	sessions, err := d.ListSessions()
	if err != nil {
		return fmt.Errorf("cleanup: list sessions: %w", err)
	}
	for _, s := range *sessions {
		if err := d.DeleteSession(ctx, ac, s.SessionID); err != nil {
			return fmt.Errorf("cleanup: delete session %s: %w", s.SessionID, err)
		}
	}

	userOauths, err := d.ListUserOauths()
	if err != nil {
		return fmt.Errorf("cleanup: list user_oauth: %w", err)
	}
	for _, uo := range *userOauths {
		if err := d.DeleteUserOauth(ctx, ac, uo.UserOauthID); err != nil {
			return fmt.Errorf("cleanup: delete user_oauth %s: %w", uo.UserOauthID, err)
		}
	}

	systemUserID := types.NullableUserID{Valid: true, ID: types.SystemUserID}
	sshKeys, err := d.ListUserSshKeys(systemUserID)
	if err != nil {
		return fmt.Errorf("cleanup: list user_ssh_keys: %w", err)
	}
	for _, sk := range *sshKeys {
		if err := d.DeleteUserSshKey(ctx, ac, sk.SshKeyID); err != nil {
			return fmt.Errorf("cleanup: delete user_ssh_key %s: %w", sk.SshKeyID, err)
		}
	}

	utility.DefaultLogger.Finfo("Bootstrap cleanup: removed all verification-only records")
	return nil
}

// ValidateBootstrapData verifies all tables have expected bootstrap records.
// CRITICAL: Call after CreateBootstrapData() to catch any silent failures.
// Returns detailed error if any table validation fails.
func (d Database) ValidateBootstrapData() error {
	var errors []string

	// Validate permissions table (should have at least 1 record)
	permCount, err := d.CountPermissions()
	if err != nil || permCount == nil || *permCount < 1 {
		errors = append(errors, "permissions table: expected ≥1 records, validation failed")
	}

	// Validate roles table (should have at least 2 records: admin + viewer)
	roleCount, err := d.CountRoles()
	if err != nil || roleCount == nil || *roleCount < 2 {
		errors = append(errors, "roles table: expected ≥2 records, validation failed")
	}

	// Validate users table (should have at least 1 record: system user)
	userCount, err := d.CountUsers()
	if err != nil || userCount == nil || *userCount < 1 {
		errors = append(errors, "users table: expected ≥1 records, validation failed")
	}

	// Validate routes table (should have at least 1 record)
	routeCount, err := d.CountRoutes()
	if err != nil || routeCount == nil || *routeCount < 1 {
		errors = append(errors, "routes table: expected ≥1 records, validation failed")
	}

	// Validate datatypes table (should have at least 1 record)
	datatypeCount, err := d.CountDatatypes()
	if err != nil || datatypeCount == nil || *datatypeCount < 1 {
		errors = append(errors, "datatypes table: expected ≥1 records, validation failed")
	}

	// Validate fields table (should have at least 1 record)
	fieldCount, err := d.CountFields()
	if err != nil || fieldCount == nil || *fieldCount < 1 {
		errors = append(errors, "fields table: expected ≥1 records, validation failed")
	}

	// Validate admin_routes table (should have at least 1 record)
	adminRouteCount, err := d.CountAdminRoutes()
	if err != nil || adminRouteCount == nil || *adminRouteCount < 1 {
		errors = append(errors, "admin_routes table: expected ≥1 records, validation failed")
	}

	// Validate admin_datatypes table (should have at least 1 record)
	adminDatatypeCount, err := d.CountAdminDatatypes()
	if err != nil || adminDatatypeCount == nil || *adminDatatypeCount < 1 {
		errors = append(errors, "admin_datatypes table: expected ≥1 records, validation failed")
	}

	// Validate admin_fields table (should have at least 1 record)
	adminFieldCount, err := d.CountAdminFields()
	if err != nil || adminFieldCount == nil || *adminFieldCount < 1 {
		errors = append(errors, "admin_fields table: expected ≥1 records, validation failed")
	}

	// Validate content_data table (should have at least 1 record)
	contentDataCount, err := d.CountContentData()
	if err != nil || contentDataCount == nil || *contentDataCount < 1 {
		errors = append(errors, "content_data table: expected ≥1 records, validation failed")
	}

	// Validate admin_content_data table (should have at least 1 record)
	adminContentDataCount, err := d.CountAdminContentData()
	if err != nil || adminContentDataCount == nil || *adminContentDataCount < 1 {
		errors = append(errors, "admin_content_data table: expected ≥1 records, validation failed")
	}

	// Validate content_fields table (should have at least 1 record)
	contentFieldCount, err := d.CountContentFields()
	if err != nil || contentFieldCount == nil || *contentFieldCount < 1 {
		errors = append(errors, "content_fields table: expected ≥1 records, validation failed")
	}

	// Validate admin_content_fields table (should have at least 1 record)
	adminContentFieldCount, err := d.CountAdminContentFields()
	if err != nil || adminContentFieldCount == nil || *adminContentFieldCount < 1 {
		errors = append(errors, "admin_content_fields table: expected ≥1 records, validation failed")
	}

	// Validate media_dimensions table (should have at least 1 record)
	mediaDimCount, err := d.CountMediaDimensions()
	if err != nil || mediaDimCount == nil || *mediaDimCount < 1 {
		errors = append(errors, "media_dimensions table: expected ≥1 records, validation failed")
	}

	// Validate media table (should have at least 1 record)
	mediaCount, err := d.CountMedia()
	if err != nil || mediaCount == nil || *mediaCount < 1 {
		errors = append(errors, "media table: expected ≥1 records, validation failed")
	}

	// Validate tokens table (should have at least 1 record)
	tokenCount, err := d.CountTokens()
	if err != nil || tokenCount == nil || *tokenCount < 1 {
		errors = append(errors, "tokens table: expected ≥1 records, validation failed")
	}

	// Validate sessions table (should have at least 1 record)
	sessionCount, err := d.CountSessions()
	if err != nil || sessionCount == nil || *sessionCount < 1 {
		errors = append(errors, "sessions table: expected ≥1 records, validation failed")
	}

	// Validate user_oauth table (should have at least 1 record)
	userOauthCount, err := d.CountUserOauths()
	if err != nil || userOauthCount == nil || *userOauthCount < 1 {
		errors = append(errors, "user_oauth table: expected ≥1 records, validation failed")
	}

	// Validate user_ssh_keys table (should have at least 1 record)
	userSshKeyCount, err := d.CountUserSshKeys()
	if err != nil || userSshKeyCount == nil || *userSshKeyCount < 1 {
		errors = append(errors, "user_ssh_keys table: expected ≥1 records, validation failed")
	}

	// Validate field_types table (should have at least 1 record)
	fieldTypeCount, err := d.CountFieldTypes()
	if err != nil || fieldTypeCount == nil || *fieldTypeCount < 1 {
		errors = append(errors, "field_types table: expected ≥1 records, validation failed")
	}

	// Validate admin_field_types table (should have at least 1 record)
	adminFieldTypeCount, err := d.CountAdminFieldTypes()
	if err != nil || adminFieldTypeCount == nil || *adminFieldTypeCount < 1 {
		errors = append(errors, "admin_field_types table: expected ≥1 records, validation failed")
	}

	// Validate tables table (should have EXACTLY 27 records - all core tables)
	tableCount, err := d.CountTables()
	if err != nil || tableCount == nil || *tableCount < 27 {
		actual := int64(0)
		if tableCount != nil {
			actual = *tableCount
		}
		errors = append(errors, fmt.Sprintf("tables table: expected at least 27 records (table registry), got %d", actual))
	}

	// If any validation failed, return combined error
	if len(errors) > 0 {
		err := fmt.Errorf("bootstrap validation failed:\n  - %s", strings.Join(errors, "\n  - "))
		utility.DefaultLogger.Ferror("Bootstrap validation failed", err)
		return err
	}

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 29 tables contain expected records")
	return nil
}

// CreateAllTables creates all MySQL database tables
func (d MysqlDatabase) CreateAllTables() error {
	// CRITICAL: Tables must be created in dependency order to satisfy FK constraints
	// See: ai/reference/TABLE_CREATION_ORDER.md for complete dependency graph

	// Infrastructure tables (no dependencies on application tables)
	err := d.CreateChangeEventsTable()
	if err != nil {
		return err
	}

	err = d.CreateBackupTables()
	if err != nil {
		return err
	}

	// Tier 0: Foundation tables (no dependencies)
	err = d.CreatePermissionTable()
	if err != nil {
		return err
	}

	err = d.CreateRoleTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaDimensionTable()
	if err != nil {
		return err
	}

	err = d.CreateFieldTypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTypeTable()
	if err != nil {
		return err
	}

	err = d.CreateLocaleTable()
	if err != nil {
		return err
	}

	// Tier 1: User management (depends on roles)
	err = d.CreateUserTable()
	if err != nil {
		return err
	}

	// Tier 2: User-related tables and core content tables (depend on users)
	err = d.CreateTokenTable()
	if err != nil {
		return err
	}

	err = d.CreateUserOauthTable()
	if err != nil {
		return err
	}

	err = d.CreateUserSshKeyTable()
	if err != nil {
		return err
	}

	err = d.CreateSessionTable()
	if err != nil {
		return err
	}

	err = d.CreateTableTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaFolderTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeTable()
	if err != nil {
		return err
	}

	// Tier 3: Field definition tables (depend on datatypes)
	err = d.CreateFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTable()
	if err != nil {
		return err
	}

	// Tier 4: Content data tables (depend on routes, datatypes, users)
	err = d.CreateContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentDataTable()
	if err != nil {
		return err
	}

	// Tier 5: Content field values (depend on content_data and fields)
	err = d.CreateContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentFieldTable()
	if err != nil {
		return err
	}

	// Tier 5.5: Content relation tables (depend on content_data and fields)
	err = d.CreateContentRelationTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentRelationTable()
	if err != nil {
		return err
	}

	// Tier 5.5b: Content version tables (depend on content_data + users)
	err = d.CreateContentVersionTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentVersionTable()
	if err != nil {
		return err
	}

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateRolePermissionsTable()
	if err != nil {
		return err
	}

	// Tier 7: Plugin system tables (plugins before pipelines for FK)
	err = d.CreatePluginTable()
	if err != nil {
		return err
	}

	err = d.CreatePipelineTable()
	if err != nil {
		return err
	}

	// Tier 7.5: Webhook system tables (webhooks before deliveries for FK)
	err = d.CreateWebhookTable()
	if err != nil {
		return err
	}

	err = d.CreateWebhookDeliveryTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 28 tables to verify successful table creation.
// If any table failed to create, this function will catch it immediately during install rather than later during operation.
func (d MysqlDatabase) CreateBootstrapData(adminHash string) error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap", "system")

	// 1. Create system admin permission (permission_id = 1)
	permission, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
		Label:           "admin",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create system admin permission: %w", err)
	}
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 1b. Create RBAC system permissions
	rbacPermissionLabels := []string{
		"content:read", "content:create", "content:update", "content:delete", "content:publish", "content:admin",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete", "datatypes:admin",
		"fields:read", "fields:create", "fields:update", "fields:delete", "fields:admin",
		"media:read", "media:create", "media:update", "media:delete", "media:admin",
		"routes:read", "routes:create", "routes:update", "routes:delete", "routes:admin",
		"users:read", "users:create", "users:update", "users:delete", "users:admin",
		"roles:read", "roles:create", "roles:update", "roles:delete", "roles:admin",
		"permissions:read", "permissions:create", "permissions:update", "permissions:delete", "permissions:admin",
		"sessions:read", "sessions:delete", "sessions:admin",
		"ssh_keys:read", "ssh_keys:create", "ssh_keys:delete", "ssh_keys:admin",
		"config:read", "config:update", "config:admin",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete", "admin_tree:admin",
		"field_types:read", "field_types:create", "field_types:update", "field_types:delete", "field_types:admin",
		"admin_field_types:read", "admin_field_types:create", "admin_field_types:update", "admin_field_types:delete", "admin_field_types:admin",
		"deploy:read", "deploy:create",
		"webhook:read", "webhook:create", "webhook:update", "webhook:delete", "webhook:admin",
		"plugins:read", "plugins:admin",
		"tables:read", "tables:create", "tables:update", "tables:delete", "tables:admin",
		"import:read", "import:create", "import:admin",
		"tokens:read", "tokens:create", "tokens:delete", "tokens:admin",
		"locale:read", "locale:create", "locale:update", "locale:delete", "locale:admin",
		"audit:read", "audit:admin",
		"backup:read", "backup:create", "backup:update", "backup:delete", "backup:admin",
		"search:read", "search:update", "search:admin",
	}
	rbacPermissions := make(map[string]types.PermissionID, len(rbacPermissionLabels))
	for _, label := range rbacPermissionLabels {
		p, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
			Label:           label,
			SystemProtected: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create RBAC permission %q: %w", label, err)
		}
		rbacPermissions[label] = p.PermissionID
	}

	// 2. Create system admin role (role_id = 1)
	adminRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "admin",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create system admin role: %w", err)
	}
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 2b. Create editor role
	editorRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "editor",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create editor role: %w", err)
	}
	if editorRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create editor role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "viewer",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create viewer role: %w", err)
	}
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 3b. Create role_permissions junction rows
	for _, permID := range rbacPermissions {
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       adminRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create admin role_permission: %w", err)
		}
	}

	// Editor permissions: CRUD on content, datatypes, fields, media, routes, admin_tree, field_types, admin_field_types; read-only on users, sessions, ssh_keys
	editorPermLabels := []string{
		"content:read", "content:create", "content:update", "content:delete",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete",
		"fields:read", "fields:create", "fields:update", "fields:delete",
		"media:read", "media:create", "media:update", "media:delete",
		"routes:read", "routes:create", "routes:update", "routes:delete",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete",
		"field_types:read", "field_types:create", "field_types:update", "field_types:delete",
		"admin_field_types:read", "admin_field_types:create", "admin_field_types:update", "admin_field_types:delete",
		"users:read",
		"sessions:read",
		"ssh_keys:read",
		"config:read",
		"webhook:read", "webhook:create", "webhook:update", "webhook:delete",
		"tokens:read", "tokens:create", "tokens:delete",
		"locale:read", "locale:create", "locale:update", "locale:delete",
		"audit:read",
		"import:read",
		"plugins:read",
		"backup:read",
		"search:read",
	}
	for _, label := range editorPermLabels {
		permID, ok := rbacPermissions[label]
		if !ok {
			return fmt.Errorf("missing RBAC permission %q for editor role", label)
		}
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       editorRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create editor role_permission for %q: %w", label, err)
		}
	}

	// Viewer permissions: read-only on content, media, routes, field_types, admin_field_types
	viewerPermLabels := []string{
		"content:read",
		"media:read",
		"routes:read",
		"field_types:read",
		"admin_field_types:read",
	}
	for _, label := range viewerPermLabels {
		permID, ok := rbacPermissions[label]
		if !ok {
			return fmt.Errorf("missing RBAC permission %q for viewer role", label)
		}
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       viewerRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create viewer role_permission for %q: %w", label, err)
		}
	}

	// 4. Create system user with well-known SystemUserID
	systemUser, err := d.CreateSystemUser(CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modula.local"),
		Hash:         adminHash,
		Role:         adminRole.RoleID.String(),
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create system user: %v", err)
	}
	if systemUser.UserID.IsZero() {
		return fmt.Errorf("failed to create system user: user_id is empty")
	}

	// 5. Create default home route (route_id = 1) - Recommended
	homeRoute, err := d.CreateRoute(ctx, ac, CreateRouteParams{
		Slug:         types.Slug("/"),
		Title:        "Home",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default home route: %w", err)
	}
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		Name:         "page",
		SortOrder:    0,
		Label:        "Page",
		Type:         string(types.DatatypeTypeRoot),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default page datatype: %w", err)
	}
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 6b. Create _reference system datatype
	refDatatype, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		SortOrder:    1,
		Label:        "Reference",
		Type:         string(types.DatatypeTypeReference),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create _reference datatype: %w", err)
	}
	if refDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create _reference datatype")
	}

	// 6c. Create "Target" field for _reference datatype (linked via parent_id)
	refField, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: refDatatype.DatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Target",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeIDRef,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create _reference Target field: %w", err)
	}
	if refField.FieldID.IsZero() {
		return fmt.Errorf("failed to create _reference Target field")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute, err := d.CreateAdminRoute(ctx, ac, CreateAdminRouteParams{
		Slug:         types.Slug("admin"),
		Title:        "Admin",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin route: %w", err)
	}
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype, err := d.CreateAdminDatatype(ctx, ac, CreateAdminDatatypeParams{
		ParentID:     types.NullableAdminDatatypeID{},
		SortOrder:    0,
		Label:        "Admin Page",
		Type:         string(types.DatatypeTypeRoot),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin datatype: %w", err)
	}
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1) -- linked to admin datatype via parent_id
	adminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{ID: adminDatatype.AdminDatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Content",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin field: %w", err)
	}
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1) -- linked to page datatype via parent_id
	field, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: pageDatatype.DatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Content",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default field: %w", err)
	}
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       types.NullableRouteID{Valid: true, ID: homeRoute.RouteID},
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		AuthorID:      systemUser.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default content_data: %w", err)
	}
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 11.5. Set root_id to self for the root content node
	contentRootID := types.NullableContentID{Valid: true, ID: contentData.ContentDataID}
	_, err = d.UpdateContentData(ctx, ac, UpdateContentDataParams{
		ContentDataID: contentData.ContentDataID,
		RootID:        contentRootID,
		RouteID:       contentData.RouteID,
		ParentID:      contentData.ParentID,
		FirstChildID:  contentData.FirstChildID,
		NextSiblingID: contentData.NextSiblingID,
		PrevSiblingID: contentData.PrevSiblingID,
		DatatypeID:    contentData.DatatypeID,
		AuthorID:      contentData.AuthorID,
		Status:        contentData.Status,
		DateCreated:   contentData.DateCreated,
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to set root_id on default content_data: %w", err)
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{
		ParentID:        types.NullableAdminContentID{},
		FirstChildID:    types.NullableAdminContentID{},
		NextSiblingID:   types.NullableAdminContentID{},
		PrevSiblingID:   types.NullableAdminContentID{},
		AdminRouteID:    types.NullableAdminRouteID{Valid: true, ID: adminRoute.AdminRouteID},
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AuthorID:        systemUser.UserID,
		Status:          types.ContentStatusDraft,
		DateCreated:     types.TimestampNow(),
		DateModified:    types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_content_data: %w", err)
	}
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 12.5. Set root_id to self for the root admin content node
	adminContentRootID := types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID}
	_, err = d.UpdateAdminContentData(ctx, ac, UpdateAdminContentDataParams{
		AdminContentDataID: adminContentData.AdminContentDataID,
		RootID:             adminContentRootID,
		AdminRouteID:       adminContentData.AdminRouteID,
		ParentID:           adminContentData.ParentID,
		FirstChildID:       adminContentData.FirstChildID,
		NextSiblingID:      adminContentData.NextSiblingID,
		PrevSiblingID:      adminContentData.PrevSiblingID,
		AdminDatatypeID:    adminContentData.AdminDatatypeID,
		AuthorID:           adminContentData.AuthorID,
		Status:             adminContentData.Status,
		DateCreated:        adminContentData.DateCreated,
		DateModified:       types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to set root_id on default admin_content_data: %w", err)
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		RootID:        contentRootID,
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      systemUser.UserID,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default content_field: %w", err)
	}
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField, err := d.CreateAdminContentField(ctx, ac, CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{},
		RootID:             adminContentRootID,
		AdminContentDataID: types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID},
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           systemUser.UserID,
		DateCreated:        types.TimestampNow(),
		DateModified:       types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_content_field: %w", err)
	}
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension, err := d.CreateMediaDimension(ctx, ac, CreateMediaDimensionParams{
		Label:       NewNullString("Default"),
		Width:       types.NewNullableInt64(1920),
		Height:      types.NewNullableInt64(1080),
		AspectRatio: NewNullString("16:9"),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media_dimension: %w", err)
	}
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         NewNullString("default"),
		DisplayName:  NewNullString("Default Media"),
		Alt:          NewNullString("Default"),
		Caption:      NullString{},
		Description:  NullString{},
		Class:        NullString{},
		URL:          types.URL("https://placeholder.local/default"),
		Mimetype:     NullString{},
		Dimensions:   NullString{},
		Srcset:       NullString{},
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media: %w", err)
	}
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token, err := d.CreateToken(ctx, ac, CreateTokenParams{
		UserID:    types.NullableUserID{Valid: true, ID: systemUser.UserID},
		TokenType: "validation",
		Token:     utility.HashToken("bootstrap_validation_token"),
		IssuedAt:  types.TimestampNow(),
		ExpiresAt: types.TimestampNow(),
		Revoked:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create default token: %w", err)
	}
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(ctx, ac, CreateSessionParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated: types.TimestampNow(),
		ExpiresAt:   types.TimestampNow(),
		LastAccess:  types.TimestampNow(),
		IpAddress:   NewNullString("127.0.0.1"),
		UserAgent:   NewNullString("bootstrap"),
		SessionData: NullString{},
	})
	if err != nil {
		return fmt.Errorf("failed to create default session: %v", err)
	}
	if session.SessionID.IsZero() {
		return fmt.Errorf("failed to create default session: session_id is 0")
	}

	// 19. Create default user_oauth record (user_oauth_id = 1) - Validation record
	userOauth, err := d.CreateUserOauth(ctx, ac, CreateUserOauthParams{
		UserID:              types.NullableUserID{Valid: true, ID: systemUser.UserID},
		OauthProvider:       "bootstrap",
		OauthProviderUserID: "bootstrap_user",
		AccessToken:         "bootstrap_access_token",
		RefreshToken:        "bootstrap_refresh_token",
		TokenExpiresAt:      types.TimestampNow(),
		DateCreated:         types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default user_oauth: %v", err)
	}
	if userOauth.UserOauthID.IsZero() {
		return fmt.Errorf("failed to create default user_oauth: user_oauth_id is 0")
	}

	// 19A. Create default user_ssh_key record (ssh_key_id = 1) - Validation record
	userSshKey, err := d.CreateUserSshKey(ctx, ac, CreateUserSshKeyParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBootstrapValidationKey bootstrap@modula",
		KeyType:     "ssh-ed25519",
		Fingerprint: "SHA256:bootstrap_validation_fingerprint",
		Label:       "Bootstrap Validation Key",
		DateCreated: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default user_ssh_key: %v", err)
	}
	if userSshKey.SshKeyID == "" {
		return fmt.Errorf("failed to create default user_ssh_key: ssh_key_id is empty")
	}

	// 20. Seed field_types and admin_field_types with the built-in field types
	fieldTypeSeedData := []struct{ Type, Label string }{
		{"text", "Text Input"}, {"textarea", "Text Area"}, {"number", "Number"},
		{"date", "Date"}, {"datetime", "Date & Time"}, {"boolean", "Boolean"},
		{"select", "Select"}, {"media", "Media"},
		{"json", "JSON"}, {"richtext", "Rich Text"}, {"slug", "Slug"},
		{"email", "Email"}, {"url", "URL"},
		{"_id", "ID Reference"}, {"_title", "Title"},
	}
	for _, ft := range fieldTypeSeedData {
		created, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{Type: ft.Type, Label: ft.Label})
		if err != nil {
			return fmt.Errorf("failed to seed field_type %q: %w", ft.Type, err)
		}
		if created.FieldTypeID.IsZero() {
			return fmt.Errorf("failed to seed field_type %q: id is zero", ft.Type)
		}
	}
	for _, ft := range fieldTypeSeedData {
		created, err := d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{Type: ft.Type, Label: ft.Label})
		if err != nil {
			return fmt.Errorf("failed to seed admin_field_type %q: %w", ft.Type, err)
		}
		if created.AdminFieldTypeID.IsZero() {
			return fmt.Errorf("failed to seed admin_field_type %q: id is zero", ft.Type)
		}
	}

	// 21. Register all 29 Modula tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
		"change_events",
		"backups",
		"backup_verifications",
		"backup_sets",
		"permissions",
		"roles",
		"media_dimensions",
		"media_folders",
		"field_types",
		"admin_field_types",
		"users",
		"tokens",
		"sessions",
		"routes",
		"media",
		"tables",
		"datatypes",
		"fields",
		"admin_fields",
		"content_data",
		"admin_content_data",
		"content_fields",
		"admin_content_fields",
		"admin_routes",
		"admin_datatypes",
		"user_oauth",
		"user_ssh_keys",
	}

	for _, tableName := range tableNames {
		table, err := d.CreateTable(ctx, ac, CreateTableParams{Label: tableName})
		if err != nil {
			return fmt.Errorf("failed to register table in tables registry: %s: %w", tableName, err)
		}
		if table.ID == "" {
			return fmt.Errorf("failed to register table in tables registry: %s", tableName)
		}
	}

	utility.DefaultLogger.Finfo("Bootstrap data created successfully: all tables validated with bootstrap records + complete table registry populated")
	return nil
}

// ValidateBootstrapData verifies all tables have expected bootstrap records for MySQL.
// CRITICAL: Call after CreateBootstrapData() to catch any silent failures.
// Returns detailed error if any table validation fails.
func (d MysqlDatabase) ValidateBootstrapData() error {
	var errors []string

	// Validate all 28 tables have expected record counts
	permCount, err := d.CountPermissions()
	if err != nil || permCount == nil || *permCount < 1 {
		errors = append(errors, "permissions table: expected ≥1 records, validation failed")
	}

	roleCount, err := d.CountRoles()
	if err != nil || roleCount == nil || *roleCount < 2 {
		errors = append(errors, "roles table: expected ≥2 records, validation failed")
	}

	userCount, err := d.CountUsers()
	if err != nil || userCount == nil || *userCount < 1 {
		errors = append(errors, "users table: expected ≥1 records, validation failed")
	}

	routeCount, err := d.CountRoutes()
	if err != nil || routeCount == nil || *routeCount < 1 {
		errors = append(errors, "routes table: expected ≥1 records, validation failed")
	}

	datatypeCount, err := d.CountDatatypes()
	if err != nil || datatypeCount == nil || *datatypeCount < 1 {
		errors = append(errors, "datatypes table: expected ≥1 records, validation failed")
	}

	fieldCount, err := d.CountFields()
	if err != nil || fieldCount == nil || *fieldCount < 1 {
		errors = append(errors, "fields table: expected ≥1 records, validation failed")
	}

	adminRouteCount, err := d.CountAdminRoutes()
	if err != nil || adminRouteCount == nil || *adminRouteCount < 1 {
		errors = append(errors, "admin_routes table: expected ≥1 records, validation failed")
	}

	adminDatatypeCount, err := d.CountAdminDatatypes()
	if err != nil || adminDatatypeCount == nil || *adminDatatypeCount < 1 {
		errors = append(errors, "admin_datatypes table: expected ≥1 records, validation failed")
	}

	adminFieldCount, err := d.CountAdminFields()
	if err != nil || adminFieldCount == nil || *adminFieldCount < 1 {
		errors = append(errors, "admin_fields table: expected ≥1 records, validation failed")
	}

	contentDataCount, err := d.CountContentData()
	if err != nil || contentDataCount == nil || *contentDataCount < 1 {
		errors = append(errors, "content_data table: expected ≥1 records, validation failed")
	}

	adminContentDataCount, err := d.CountAdminContentData()
	if err != nil || adminContentDataCount == nil || *adminContentDataCount < 1 {
		errors = append(errors, "admin_content_data table: expected ≥1 records, validation failed")
	}

	contentFieldCount, err := d.CountContentFields()
	if err != nil || contentFieldCount == nil || *contentFieldCount < 1 {
		errors = append(errors, "content_fields table: expected ≥1 records, validation failed")
	}

	adminContentFieldCount, err := d.CountAdminContentFields()
	if err != nil || adminContentFieldCount == nil || *adminContentFieldCount < 1 {
		errors = append(errors, "admin_content_fields table: expected ≥1 records, validation failed")
	}

	mediaDimCount, err := d.CountMediaDimensions()
	if err != nil || mediaDimCount == nil || *mediaDimCount < 1 {
		errors = append(errors, "media_dimensions table: expected ≥1 records, validation failed")
	}

	mediaCount, err := d.CountMedia()
	if err != nil || mediaCount == nil || *mediaCount < 1 {
		errors = append(errors, "media table: expected ≥1 records, validation failed")
	}

	tokenCount, err := d.CountTokens()
	if err != nil || tokenCount == nil || *tokenCount < 1 {
		errors = append(errors, "tokens table: expected ≥1 records, validation failed")
	}

	sessionCount, err := d.CountSessions()
	if err != nil || sessionCount == nil || *sessionCount < 1 {
		errors = append(errors, "sessions table: expected ≥1 records, validation failed")
	}

	userOauthCount, err := d.CountUserOauths()
	if err != nil || userOauthCount == nil || *userOauthCount < 1 {
		errors = append(errors, "user_oauth table: expected ≥1 records, validation failed")
	}

	fieldTypeCount, err := d.CountFieldTypes()
	if err != nil || fieldTypeCount == nil || *fieldTypeCount < 1 {
		errors = append(errors, "field_types table: expected ≥1 records, validation failed")
	}

	adminFieldTypeCount, err := d.CountAdminFieldTypes()
	if err != nil || adminFieldTypeCount == nil || *adminFieldTypeCount < 1 {
		errors = append(errors, "admin_field_types table: expected ≥1 records, validation failed")
	}

	tableCount, err := d.CountTables()
	if err != nil || tableCount == nil || *tableCount < 27 {
		actual := int64(0)
		if tableCount != nil {
			actual = *tableCount
		}
		errors = append(errors, fmt.Sprintf("tables table: expected at least 27 records (table registry), got %d", actual))
	}

	if len(errors) > 0 {
		err := fmt.Errorf("bootstrap validation failed:\n  - %s", strings.Join(errors, "\n  - "))
		utility.DefaultLogger.Ferror("Bootstrap validation failed", err)
		return err
	}

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 28 tables contain expected records (MySQL)")
	return nil
}

// CleanupBootstrapData removes verification-only records after validation (MySQL).
func (d MysqlDatabase) CleanupBootstrapData() error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap-cleanup", "system")

	contentFields, err := d.ListContentFields()
	if err != nil {
		return fmt.Errorf("cleanup: list content_fields: %w", err)
	}
	for _, cf := range *contentFields {
		if err := d.DeleteContentField(ctx, ac, cf.ContentFieldID); err != nil {
			return fmt.Errorf("cleanup: delete content_field %s: %w", cf.ContentFieldID, err)
		}
	}

	contentData, err := d.ListContentData()
	if err != nil {
		return fmt.Errorf("cleanup: list content_data: %w", err)
	}
	for _, cd := range *contentData {
		if err := d.DeleteContentData(ctx, ac, cd.ContentDataID); err != nil {
			return fmt.Errorf("cleanup: delete content_data %s: %w", cd.ContentDataID, err)
		}
	}

	fields, err := d.ListFields()
	if err != nil {
		return fmt.Errorf("cleanup: list fields: %w", err)
	}
	for _, f := range *fields {
		if err := d.DeleteField(ctx, ac, f.FieldID); err != nil {
			return fmt.Errorf("cleanup: delete field %s: %w", f.FieldID, err)
		}
	}

	datatypes, err := d.ListDatatypes()
	if err != nil {
		return fmt.Errorf("cleanup: list datatypes: %w", err)
	}
	for _, dt := range *datatypes {
		if err := d.DeleteDatatype(ctx, ac, dt.DatatypeID); err != nil {
			return fmt.Errorf("cleanup: delete datatype %s: %w", dt.DatatypeID, err)
		}
	}

	routes, err := d.ListRoutes()
	if err != nil {
		return fmt.Errorf("cleanup: list routes: %w", err)
	}
	for _, r := range *routes {
		if err := d.DeleteRoute(ctx, ac, r.RouteID); err != nil {
			return fmt.Errorf("cleanup: delete route %s: %w", r.RouteID, err)
		}
	}

	adminContentFields, err := d.ListAdminContentFields()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_content_fields: %w", err)
	}
	for _, acf := range *adminContentFields {
		if err := d.DeleteAdminContentField(ctx, ac, acf.AdminContentFieldID); err != nil {
			return fmt.Errorf("cleanup: delete admin_content_field %s: %w", acf.AdminContentFieldID, err)
		}
	}

	adminContentData, err := d.ListAdminContentData()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_content_data: %w", err)
	}
	for _, acd := range *adminContentData {
		if err := d.DeleteAdminContentData(ctx, ac, acd.AdminContentDataID); err != nil {
			return fmt.Errorf("cleanup: delete admin_content_data %s: %w", acd.AdminContentDataID, err)
		}
	}

	adminFields, err := d.ListAdminFields()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_fields: %w", err)
	}
	for _, af := range *adminFields {
		if err := d.DeleteAdminField(ctx, ac, af.AdminFieldID); err != nil {
			return fmt.Errorf("cleanup: delete admin_field %s: %w", af.AdminFieldID, err)
		}
	}

	adminDatatypes, err := d.ListAdminDatatypes()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_datatypes: %w", err)
	}
	for _, adt := range *adminDatatypes {
		if err := d.DeleteAdminDatatype(ctx, ac, adt.AdminDatatypeID); err != nil {
			return fmt.Errorf("cleanup: delete admin_datatype %s: %w", adt.AdminDatatypeID, err)
		}
	}

	adminRoutes, err := d.ListAdminRoutes()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_routes: %w", err)
	}
	for _, ar := range *adminRoutes {
		if err := d.DeleteAdminRoute(ctx, ac, ar.AdminRouteID); err != nil {
			return fmt.Errorf("cleanup: delete admin_route %s: %w", ar.AdminRouteID, err)
		}
	}

	media, err := d.ListMedia()
	if err != nil {
		return fmt.Errorf("cleanup: list media: %w", err)
	}
	for _, m := range *media {
		if err := d.DeleteMedia(ctx, ac, m.MediaID); err != nil {
			return fmt.Errorf("cleanup: delete media %s: %w", m.MediaID, err)
		}
	}

	mediaDimensions, err := d.ListMediaDimensions()
	if err != nil {
		return fmt.Errorf("cleanup: list media_dimensions: %w", err)
	}
	for _, md := range *mediaDimensions {
		if err := d.DeleteMediaDimension(ctx, ac, md.MdID); err != nil {
			return fmt.Errorf("cleanup: delete media_dimension %s: %w", md.MdID, err)
		}
	}

	tokens, err := d.ListTokens()
	if err != nil {
		return fmt.Errorf("cleanup: list tokens: %w", err)
	}
	for _, tk := range *tokens {
		if err := d.DeleteToken(ctx, ac, tk.ID); err != nil {
			return fmt.Errorf("cleanup: delete token %s: %w", tk.ID, err)
		}
	}

	sessions, err := d.ListSessions()
	if err != nil {
		return fmt.Errorf("cleanup: list sessions: %w", err)
	}
	for _, s := range *sessions {
		if err := d.DeleteSession(ctx, ac, s.SessionID); err != nil {
			return fmt.Errorf("cleanup: delete session %s: %w", s.SessionID, err)
		}
	}

	userOauths, err := d.ListUserOauths()
	if err != nil {
		return fmt.Errorf("cleanup: list user_oauth: %w", err)
	}
	for _, uo := range *userOauths {
		if err := d.DeleteUserOauth(ctx, ac, uo.UserOauthID); err != nil {
			return fmt.Errorf("cleanup: delete user_oauth %s: %w", uo.UserOauthID, err)
		}
	}

	systemUserID := types.NullableUserID{Valid: true, ID: types.SystemUserID}
	sshKeys, err := d.ListUserSshKeys(systemUserID)
	if err != nil {
		return fmt.Errorf("cleanup: list user_ssh_keys: %w", err)
	}
	for _, sk := range *sshKeys {
		if err := d.DeleteUserSshKey(ctx, ac, sk.SshKeyID); err != nil {
			return fmt.Errorf("cleanup: delete user_ssh_key %s: %w", sk.SshKeyID, err)
		}
	}

	utility.DefaultLogger.Finfo("Bootstrap cleanup: removed all verification-only records (MySQL)")
	return nil
}

// CreateAllTables creates all PostgreSQL database tables
func (d PsqlDatabase) CreateAllTables() error {
	// CRITICAL: Tables must be created in dependency order to satisfy FK constraints
	// See: ai/reference/TABLE_CREATION_ORDER.md for complete dependency graph

	// Infrastructure tables (no dependencies on application tables)
	err := d.CreateChangeEventsTable()
	if err != nil {
		return err
	}

	err = d.CreateBackupTables()
	if err != nil {
		return err
	}

	// Tier 0: Foundation tables (no dependencies)
	err = d.CreatePermissionTable()
	if err != nil {
		return err
	}

	err = d.CreateRoleTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaDimensionTable()
	if err != nil {
		return err
	}

	err = d.CreateFieldTypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTypeTable()
	if err != nil {
		return err
	}

	err = d.CreateLocaleTable()
	if err != nil {
		return err
	}

	// Tier 1: User management (depends on roles)
	err = d.CreateUserTable()
	if err != nil {
		return err
	}

	// Tier 2: User-related tables and core content tables (depend on users)
	err = d.CreateTokenTable()
	if err != nil {
		return err
	}

	err = d.CreateUserOauthTable()
	if err != nil {
		return err
	}

	err = d.CreateUserSshKeyTable()
	if err != nil {
		return err
	}

	err = d.CreateSessionTable()
	if err != nil {
		return err
	}

	err = d.CreateTableTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaFolderTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeTable()
	if err != nil {
		return err
	}

	// Tier 3: Field definition tables (depend on datatypes)
	err = d.CreateFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTable()
	if err != nil {
		return err
	}

	// Tier 4: Content data tables (depend on routes, datatypes, users)
	err = d.CreateContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentDataTable()
	if err != nil {
		return err
	}

	// Tier 5: Content field values (depend on content_data and fields)
	err = d.CreateContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentFieldTable()
	if err != nil {
		return err
	}

	// Tier 5.5: Content relation tables (depend on content_data and fields)
	err = d.CreateContentRelationTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentRelationTable()
	if err != nil {
		return err
	}

	// Tier 5.5b: Content version tables (depend on content_data + users)
	err = d.CreateContentVersionTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentVersionTable()
	if err != nil {
		return err
	}

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateRolePermissionsTable()
	if err != nil {
		return err
	}

	// Tier 7: Plugin system tables (plugins before pipelines for FK)
	err = d.CreatePluginTable()
	if err != nil {
		return err
	}

	err = d.CreatePipelineTable()
	if err != nil {
		return err
	}

	// Tier 7.5: Webhook system tables (webhooks before deliveries for FK)
	err = d.CreateWebhookTable()
	if err != nil {
		return err
	}

	err = d.CreateWebhookDeliveryTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 28 tables to verify successful table creation.
// If any table failed to create, this function will catch it immediately during install rather than later during operation.
func (d PsqlDatabase) CreateBootstrapData(adminHash string) error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap", "system")

	// 1. Create system admin permission (permission_id = 1)
	permission, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
		Label:           "admin",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create system admin permission: %w", err)
	}
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 1b. Create RBAC system permissions
	rbacPermissionLabels := []string{
		"content:read", "content:create", "content:update", "content:delete", "content:publish", "content:admin",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete", "datatypes:admin",
		"fields:read", "fields:create", "fields:update", "fields:delete", "fields:admin",
		"media:read", "media:create", "media:update", "media:delete", "media:admin",
		"routes:read", "routes:create", "routes:update", "routes:delete", "routes:admin",
		"users:read", "users:create", "users:update", "users:delete", "users:admin",
		"roles:read", "roles:create", "roles:update", "roles:delete", "roles:admin",
		"permissions:read", "permissions:create", "permissions:update", "permissions:delete", "permissions:admin",
		"sessions:read", "sessions:delete", "sessions:admin",
		"ssh_keys:read", "ssh_keys:create", "ssh_keys:delete", "ssh_keys:admin",
		"config:read", "config:update", "config:admin",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete", "admin_tree:admin",
		"field_types:read", "field_types:create", "field_types:update", "field_types:delete", "field_types:admin",
		"admin_field_types:read", "admin_field_types:create", "admin_field_types:update", "admin_field_types:delete", "admin_field_types:admin",
		"deploy:read", "deploy:create",
		"webhook:read", "webhook:create", "webhook:update", "webhook:delete", "webhook:admin",
		"plugins:read", "plugins:admin",
		"tables:read", "tables:create", "tables:update", "tables:delete", "tables:admin",
		"import:read", "import:create", "import:admin",
		"tokens:read", "tokens:create", "tokens:delete", "tokens:admin",
		"locale:read", "locale:create", "locale:update", "locale:delete", "locale:admin",
		"audit:read", "audit:admin",
		"backup:read", "backup:create", "backup:update", "backup:delete", "backup:admin",
		"search:read", "search:update", "search:admin",
	}
	rbacPermissions := make(map[string]types.PermissionID, len(rbacPermissionLabels))
	for _, label := range rbacPermissionLabels {
		p, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
			Label:           label,
			SystemProtected: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create RBAC permission %q: %w", label, err)
		}
		rbacPermissions[label] = p.PermissionID
	}

	// 2. Create system admin role (role_id = 1)
	adminRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "admin",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create system admin role: %w", err)
	}
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 2b. Create editor role
	editorRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "editor",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create editor role: %w", err)
	}
	if editorRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create editor role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:           "viewer",
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create viewer role: %w", err)
	}
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 3b. Create role_permissions junction rows
	for _, permID := range rbacPermissions {
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       adminRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create admin role_permission: %w", err)
		}
	}

	// Editor permissions: CRUD on content, datatypes, fields, media, routes, admin_tree, field_types, admin_field_types; read-only on users, sessions, ssh_keys
	editorPermLabels := []string{
		"content:read", "content:create", "content:update", "content:delete",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete",
		"fields:read", "fields:create", "fields:update", "fields:delete",
		"media:read", "media:create", "media:update", "media:delete",
		"routes:read", "routes:create", "routes:update", "routes:delete",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete",
		"field_types:read", "field_types:create", "field_types:update", "field_types:delete",
		"admin_field_types:read", "admin_field_types:create", "admin_field_types:update", "admin_field_types:delete",
		"users:read",
		"sessions:read",
		"ssh_keys:read",
		"config:read",
		"webhook:read", "webhook:create", "webhook:update", "webhook:delete",
		"tokens:read", "tokens:create", "tokens:delete",
		"locale:read", "locale:create", "locale:update", "locale:delete",
		"audit:read",
		"import:read",
		"plugins:read",
		"backup:read",
		"search:read",
	}
	for _, label := range editorPermLabels {
		permID, ok := rbacPermissions[label]
		if !ok {
			return fmt.Errorf("missing RBAC permission %q for editor role", label)
		}
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       editorRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create editor role_permission for %q: %w", label, err)
		}
	}

	// Viewer permissions: read-only on content, media, routes, field_types, admin_field_types
	viewerPermLabels := []string{
		"content:read",
		"media:read",
		"routes:read",
		"field_types:read",
		"admin_field_types:read",
	}
	for _, label := range viewerPermLabels {
		permID, ok := rbacPermissions[label]
		if !ok {
			return fmt.Errorf("missing RBAC permission %q for viewer role", label)
		}
		_, err := d.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
			RoleID:       viewerRole.RoleID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("failed to create viewer role_permission for %q: %w", label, err)
		}
	}

	// 4. Create system user with well-known SystemUserID
	systemUser, err := d.CreateSystemUser(CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modula.local"),
		Hash:         adminHash,
		Role:         adminRole.RoleID.String(),
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create system user: %v", err)
	}
	if systemUser.UserID.IsZero() {
		return fmt.Errorf("failed to create system user: user_id is empty")
	}

	// 5. Create default home route (route_id = 1) - Recommended
	homeRoute, err := d.CreateRoute(ctx, ac, CreateRouteParams{
		Slug:         types.Slug("/"),
		Title:        "Home",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default home route: %w", err)
	}
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		Name:         "page",
		SortOrder:    0,
		Label:        "Page",
		Type:         string(types.DatatypeTypeRoot),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default page datatype: %w", err)
	}
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 6b. Create _reference system datatype
	refDatatype, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		SortOrder:    1,
		Label:        "Reference",
		Type:         string(types.DatatypeTypeReference),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create _reference datatype: %w", err)
	}
	if refDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create _reference datatype")
	}

	// 6c. Create "Target" field for _reference datatype (linked via parent_id)
	refField, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: refDatatype.DatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Target",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeIDRef,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create _reference Target field: %w", err)
	}
	if refField.FieldID.IsZero() {
		return fmt.Errorf("failed to create _reference Target field")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute, err := d.CreateAdminRoute(ctx, ac, CreateAdminRouteParams{
		Slug:         types.Slug("admin"),
		Title:        "Admin",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin route: %w", err)
	}
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype, err := d.CreateAdminDatatype(ctx, ac, CreateAdminDatatypeParams{
		ParentID:     types.NullableAdminDatatypeID{},
		SortOrder:    0,
		Label:        "Admin Page",
		Type:         string(types.DatatypeTypeRoot),
		AuthorID:     systemUser.UserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin datatype: %w", err)
	}
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1) -- linked to admin datatype via parent_id
	adminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{ID: adminDatatype.AdminDatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Content",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin field: %w", err)
	}
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1) -- linked to page datatype via parent_id
	field, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: pageDatatype.DatatypeID, Valid: true},
		SortOrder:    0,
		Label:        "Content",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default field: %w", err)
	}
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       types.NullableRouteID{Valid: true, ID: homeRoute.RouteID},
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		AuthorID:      systemUser.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default content_data: %w", err)
	}
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 11.5. Set root_id to self for the root content node
	contentRootID := types.NullableContentID{Valid: true, ID: contentData.ContentDataID}
	_, err = d.UpdateContentData(ctx, ac, UpdateContentDataParams{
		ContentDataID: contentData.ContentDataID,
		RootID:        contentRootID,
		RouteID:       contentData.RouteID,
		ParentID:      contentData.ParentID,
		FirstChildID:  contentData.FirstChildID,
		NextSiblingID: contentData.NextSiblingID,
		PrevSiblingID: contentData.PrevSiblingID,
		DatatypeID:    contentData.DatatypeID,
		AuthorID:      contentData.AuthorID,
		Status:        contentData.Status,
		DateCreated:   contentData.DateCreated,
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to set root_id on default content_data: %w", err)
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{
		ParentID:        types.NullableAdminContentID{},
		FirstChildID:    types.NullableAdminContentID{},
		NextSiblingID:   types.NullableAdminContentID{},
		PrevSiblingID:   types.NullableAdminContentID{},
		AdminRouteID:    types.NullableAdminRouteID{Valid: true, ID: adminRoute.AdminRouteID},
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AuthorID:        systemUser.UserID,
		Status:          types.ContentStatusDraft,
		DateCreated:     types.TimestampNow(),
		DateModified:    types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_content_data: %w", err)
	}
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 12.5. Set root_id to self for the root admin content node
	adminContentRootID := types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID}
	_, err = d.UpdateAdminContentData(ctx, ac, UpdateAdminContentDataParams{
		AdminContentDataID: adminContentData.AdminContentDataID,
		RootID:             adminContentRootID,
		AdminRouteID:       adminContentData.AdminRouteID,
		ParentID:           adminContentData.ParentID,
		FirstChildID:       adminContentData.FirstChildID,
		NextSiblingID:      adminContentData.NextSiblingID,
		PrevSiblingID:      adminContentData.PrevSiblingID,
		AdminDatatypeID:    adminContentData.AdminDatatypeID,
		AuthorID:           adminContentData.AuthorID,
		Status:             adminContentData.Status,
		DateCreated:        adminContentData.DateCreated,
		DateModified:       types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to set root_id on default admin_content_data: %w", err)
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		RootID:        contentRootID,
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      systemUser.UserID,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default content_field: %w", err)
	}
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField, err := d.CreateAdminContentField(ctx, ac, CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{},
		RootID:             adminContentRootID,
		AdminContentDataID: types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID},
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           systemUser.UserID,
		DateCreated:        types.TimestampNow(),
		DateModified:       types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_content_field: %w", err)
	}
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension, err := d.CreateMediaDimension(ctx, ac, CreateMediaDimensionParams{
		Label:       NewNullString("Default"),
		Width:       types.NewNullableInt64(1920),
		Height:      types.NewNullableInt64(1080),
		AspectRatio: NewNullString("16:9"),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media_dimension: %w", err)
	}
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         NewNullString("default"),
		DisplayName:  NewNullString("Default Media"),
		Alt:          NewNullString("Default"),
		Caption:      NullString{},
		Description:  NullString{},
		Class:        NullString{},
		URL:          types.URL("https://placeholder.local/default"),
		Mimetype:     NullString{},
		Dimensions:   NullString{},
		Srcset:       NullString{},
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media: %w", err)
	}
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token, err := d.CreateToken(ctx, ac, CreateTokenParams{
		UserID:    types.NullableUserID{Valid: true, ID: systemUser.UserID},
		TokenType: "validation",
		Token:     utility.HashToken("bootstrap_validation_token"),
		IssuedAt:  types.TimestampNow(),
		ExpiresAt: types.TimestampNow(),
		Revoked:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create default token: %w", err)
	}
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(ctx, ac, CreateSessionParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated: types.TimestampNow(),
		ExpiresAt:   types.TimestampNow(),
		LastAccess:  types.TimestampNow(),
		IpAddress:   NewNullString("127.0.0.1"),
		UserAgent:   NewNullString("bootstrap"),
		SessionData: NullString{},
	})
	if err != nil {
		return fmt.Errorf("failed to create default session: %v", err)
	}
	if session.SessionID.IsZero() {
		return fmt.Errorf("failed to create default session: session_id is 0")
	}

	// 19. Create default user_oauth record (user_oauth_id = 1) - Validation record
	userOauth, err := d.CreateUserOauth(ctx, ac, CreateUserOauthParams{
		UserID:              types.NullableUserID{Valid: true, ID: systemUser.UserID},
		OauthProvider:       "bootstrap",
		OauthProviderUserID: "bootstrap_user",
		AccessToken:         "bootstrap_access_token",
		RefreshToken:        "bootstrap_refresh_token",
		TokenExpiresAt:      types.TimestampNow(),
		DateCreated:         types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default user_oauth: %v", err)
	}
	if userOauth.UserOauthID.IsZero() {
		return fmt.Errorf("failed to create default user_oauth: user_oauth_id is 0")
	}

	// 19A. Create default user_ssh_key record (ssh_key_id = 1) - Validation record
	userSshKey, err := d.CreateUserSshKey(ctx, ac, CreateUserSshKeyParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBootstrapValidationKey bootstrap@modula",
		KeyType:     "ssh-ed25519",
		Fingerprint: "SHA256:bootstrap_validation_fingerprint",
		Label:       "Bootstrap Validation Key",
		DateCreated: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default user_ssh_key: %v", err)
	}
	if userSshKey.SshKeyID == "" {
		return fmt.Errorf("failed to create default user_ssh_key: ssh_key_id is empty")
	}

	// 20. Seed field_types and admin_field_types with the built-in field types
	fieldTypeSeedData := []struct{ Type, Label string }{
		{"text", "Text Input"}, {"textarea", "Text Area"}, {"number", "Number"},
		{"date", "Date"}, {"datetime", "Date & Time"}, {"boolean", "Boolean"},
		{"select", "Select"}, {"media", "Media"},
		{"json", "JSON"}, {"richtext", "Rich Text"}, {"slug", "Slug"},
		{"email", "Email"}, {"url", "URL"},
		{"_id", "ID Reference"}, {"_title", "Title"},
	}
	for _, ft := range fieldTypeSeedData {
		created, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{Type: ft.Type, Label: ft.Label})
		if err != nil {
			return fmt.Errorf("failed to seed field_type %q: %w", ft.Type, err)
		}
		if created.FieldTypeID.IsZero() {
			return fmt.Errorf("failed to seed field_type %q: id is zero", ft.Type)
		}
	}
	for _, ft := range fieldTypeSeedData {
		created, err := d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{Type: ft.Type, Label: ft.Label})
		if err != nil {
			return fmt.Errorf("failed to seed admin_field_type %q: %w", ft.Type, err)
		}
		if created.AdminFieldTypeID.IsZero() {
			return fmt.Errorf("failed to seed admin_field_type %q: id is zero", ft.Type)
		}
	}

	// 21. Register all 29 Modula tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
		"change_events",
		"backups",
		"backup_verifications",
		"backup_sets",
		"permissions",
		"roles",
		"media_dimensions",
		"media_folders",
		"field_types",
		"admin_field_types",
		"users",
		"tokens",
		"sessions",
		"routes",
		"media",
		"tables",
		"datatypes",
		"fields",
		"admin_fields",
		"content_data",
		"admin_content_data",
		"content_fields",
		"admin_content_fields",
		"admin_routes",
		"admin_datatypes",
		"user_oauth",
		"user_ssh_keys",
	}

	for _, tableName := range tableNames {
		table, err := d.CreateTable(ctx, ac, CreateTableParams{Label: tableName})
		if err != nil {
			return fmt.Errorf("failed to register table in tables registry: %s: %w", tableName, err)
		}
		if table.ID == "" {
			return fmt.Errorf("failed to register table in tables registry: %s", tableName)
		}
	}

	utility.DefaultLogger.Finfo("Bootstrap data created successfully: all tables validated with bootstrap records + complete table registry populated")
	return nil
}

// ValidateBootstrapData verifies all tables have expected bootstrap records for PostgreSQL.
// CRITICAL: Call after CreateBootstrapData() to catch any silent failures.
// Returns detailed error if any table validation fails.
func (d PsqlDatabase) ValidateBootstrapData() error {
	var errors []string

	// Validate all 28 tables have expected record counts
	permCount, err := d.CountPermissions()
	if err != nil || permCount == nil || *permCount < 1 {
		errors = append(errors, "permissions table: expected ≥1 records, validation failed")
	}

	roleCount, err := d.CountRoles()
	if err != nil || roleCount == nil || *roleCount < 2 {
		errors = append(errors, "roles table: expected ≥2 records, validation failed")
	}

	userCount, err := d.CountUsers()
	if err != nil || userCount == nil || *userCount < 1 {
		errors = append(errors, "users table: expected ≥1 records, validation failed")
	}

	routeCount, err := d.CountRoutes()
	if err != nil || routeCount == nil || *routeCount < 1 {
		errors = append(errors, "routes table: expected ≥1 records, validation failed")
	}

	datatypeCount, err := d.CountDatatypes()
	if err != nil || datatypeCount == nil || *datatypeCount < 1 {
		errors = append(errors, "datatypes table: expected ≥1 records, validation failed")
	}

	fieldCount, err := d.CountFields()
	if err != nil || fieldCount == nil || *fieldCount < 1 {
		errors = append(errors, "fields table: expected ≥1 records, validation failed")
	}

	adminRouteCount, err := d.CountAdminRoutes()
	if err != nil || adminRouteCount == nil || *adminRouteCount < 1 {
		errors = append(errors, "admin_routes table: expected ≥1 records, validation failed")
	}

	adminDatatypeCount, err := d.CountAdminDatatypes()
	if err != nil || adminDatatypeCount == nil || *adminDatatypeCount < 1 {
		errors = append(errors, "admin_datatypes table: expected ≥1 records, validation failed")
	}

	adminFieldCount, err := d.CountAdminFields()
	if err != nil || adminFieldCount == nil || *adminFieldCount < 1 {
		errors = append(errors, "admin_fields table: expected ≥1 records, validation failed")
	}

	contentDataCount, err := d.CountContentData()
	if err != nil || contentDataCount == nil || *contentDataCount < 1 {
		errors = append(errors, "content_data table: expected ≥1 records, validation failed")
	}

	adminContentDataCount, err := d.CountAdminContentData()
	if err != nil || adminContentDataCount == nil || *adminContentDataCount < 1 {
		errors = append(errors, "admin_content_data table: expected ≥1 records, validation failed")
	}

	contentFieldCount, err := d.CountContentFields()
	if err != nil || contentFieldCount == nil || *contentFieldCount < 1 {
		errors = append(errors, "content_fields table: expected ≥1 records, validation failed")
	}

	adminContentFieldCount, err := d.CountAdminContentFields()
	if err != nil || adminContentFieldCount == nil || *adminContentFieldCount < 1 {
		errors = append(errors, "admin_content_fields table: expected ≥1 records, validation failed")
	}

	mediaDimCount, err := d.CountMediaDimensions()
	if err != nil || mediaDimCount == nil || *mediaDimCount < 1 {
		errors = append(errors, "media_dimensions table: expected ≥1 records, validation failed")
	}

	mediaCount, err := d.CountMedia()
	if err != nil || mediaCount == nil || *mediaCount < 1 {
		errors = append(errors, "media table: expected ≥1 records, validation failed")
	}

	tokenCount, err := d.CountTokens()
	if err != nil || tokenCount == nil || *tokenCount < 1 {
		errors = append(errors, "tokens table: expected ≥1 records, validation failed")
	}

	sessionCount, err := d.CountSessions()
	if err != nil || sessionCount == nil || *sessionCount < 1 {
		errors = append(errors, "sessions table: expected ≥1 records, validation failed")
	}

	userOauthCount, err := d.CountUserOauths()
	if err != nil || userOauthCount == nil || *userOauthCount < 1 {
		errors = append(errors, "user_oauth table: expected ≥1 records, validation failed")
	}

	fieldTypeCount, err := d.CountFieldTypes()
	if err != nil || fieldTypeCount == nil || *fieldTypeCount < 1 {
		errors = append(errors, "field_types table: expected ≥1 records, validation failed")
	}

	adminFieldTypeCount, err := d.CountAdminFieldTypes()
	if err != nil || adminFieldTypeCount == nil || *adminFieldTypeCount < 1 {
		errors = append(errors, "admin_field_types table: expected ≥1 records, validation failed")
	}

	tableCount, err := d.CountTables()
	if err != nil || tableCount == nil || *tableCount < 27 {
		actual := int64(0)
		if tableCount != nil {
			actual = *tableCount
		}
		errors = append(errors, fmt.Sprintf("tables table: expected at least 27 records (table registry), got %d", actual))
	}

	if len(errors) > 0 {
		err := fmt.Errorf("bootstrap validation failed:\n  - %s", strings.Join(errors, "\n  - "))
		utility.DefaultLogger.Ferror("Bootstrap validation failed", err)
		return err
	}

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 28 tables contain expected records (PostgreSQL)")
	return nil
}

// CleanupBootstrapData removes verification-only records after validation (PostgreSQL).
func (d PsqlDatabase) CleanupBootstrapData() error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap-cleanup", "system")

	contentFields, err := d.ListContentFields()
	if err != nil {
		return fmt.Errorf("cleanup: list content_fields: %w", err)
	}
	for _, cf := range *contentFields {
		if err := d.DeleteContentField(ctx, ac, cf.ContentFieldID); err != nil {
			return fmt.Errorf("cleanup: delete content_field %s: %w", cf.ContentFieldID, err)
		}
	}

	contentData, err := d.ListContentData()
	if err != nil {
		return fmt.Errorf("cleanup: list content_data: %w", err)
	}
	for _, cd := range *contentData {
		if err := d.DeleteContentData(ctx, ac, cd.ContentDataID); err != nil {
			return fmt.Errorf("cleanup: delete content_data %s: %w", cd.ContentDataID, err)
		}
	}

	fields, err := d.ListFields()
	if err != nil {
		return fmt.Errorf("cleanup: list fields: %w", err)
	}
	for _, f := range *fields {
		if err := d.DeleteField(ctx, ac, f.FieldID); err != nil {
			return fmt.Errorf("cleanup: delete field %s: %w", f.FieldID, err)
		}
	}

	datatypes, err := d.ListDatatypes()
	if err != nil {
		return fmt.Errorf("cleanup: list datatypes: %w", err)
	}
	for _, dt := range *datatypes {
		if err := d.DeleteDatatype(ctx, ac, dt.DatatypeID); err != nil {
			return fmt.Errorf("cleanup: delete datatype %s: %w", dt.DatatypeID, err)
		}
	}

	routes, err := d.ListRoutes()
	if err != nil {
		return fmt.Errorf("cleanup: list routes: %w", err)
	}
	for _, r := range *routes {
		if err := d.DeleteRoute(ctx, ac, r.RouteID); err != nil {
			return fmt.Errorf("cleanup: delete route %s: %w", r.RouteID, err)
		}
	}

	adminContentFields, err := d.ListAdminContentFields()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_content_fields: %w", err)
	}
	for _, acf := range *adminContentFields {
		if err := d.DeleteAdminContentField(ctx, ac, acf.AdminContentFieldID); err != nil {
			return fmt.Errorf("cleanup: delete admin_content_field %s: %w", acf.AdminContentFieldID, err)
		}
	}

	adminContentData, err := d.ListAdminContentData()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_content_data: %w", err)
	}
	for _, acd := range *adminContentData {
		if err := d.DeleteAdminContentData(ctx, ac, acd.AdminContentDataID); err != nil {
			return fmt.Errorf("cleanup: delete admin_content_data %s: %w", acd.AdminContentDataID, err)
		}
	}

	adminFields, err := d.ListAdminFields()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_fields: %w", err)
	}
	for _, af := range *adminFields {
		if err := d.DeleteAdminField(ctx, ac, af.AdminFieldID); err != nil {
			return fmt.Errorf("cleanup: delete admin_field %s: %w", af.AdminFieldID, err)
		}
	}

	adminDatatypes, err := d.ListAdminDatatypes()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_datatypes: %w", err)
	}
	for _, adt := range *adminDatatypes {
		if err := d.DeleteAdminDatatype(ctx, ac, adt.AdminDatatypeID); err != nil {
			return fmt.Errorf("cleanup: delete admin_datatype %s: %w", adt.AdminDatatypeID, err)
		}
	}

	adminRoutes, err := d.ListAdminRoutes()
	if err != nil {
		return fmt.Errorf("cleanup: list admin_routes: %w", err)
	}
	for _, ar := range *adminRoutes {
		if err := d.DeleteAdminRoute(ctx, ac, ar.AdminRouteID); err != nil {
			return fmt.Errorf("cleanup: delete admin_route %s: %w", ar.AdminRouteID, err)
		}
	}

	media, err := d.ListMedia()
	if err != nil {
		return fmt.Errorf("cleanup: list media: %w", err)
	}
	for _, m := range *media {
		if err := d.DeleteMedia(ctx, ac, m.MediaID); err != nil {
			return fmt.Errorf("cleanup: delete media %s: %w", m.MediaID, err)
		}
	}

	mediaDimensions, err := d.ListMediaDimensions()
	if err != nil {
		return fmt.Errorf("cleanup: list media_dimensions: %w", err)
	}
	for _, md := range *mediaDimensions {
		if err := d.DeleteMediaDimension(ctx, ac, md.MdID); err != nil {
			return fmt.Errorf("cleanup: delete media_dimension %s: %w", md.MdID, err)
		}
	}

	tokens, err := d.ListTokens()
	if err != nil {
		return fmt.Errorf("cleanup: list tokens: %w", err)
	}
	for _, tk := range *tokens {
		if err := d.DeleteToken(ctx, ac, tk.ID); err != nil {
			return fmt.Errorf("cleanup: delete token %s: %w", tk.ID, err)
		}
	}

	sessions, err := d.ListSessions()
	if err != nil {
		return fmt.Errorf("cleanup: list sessions: %w", err)
	}
	for _, s := range *sessions {
		if err := d.DeleteSession(ctx, ac, s.SessionID); err != nil {
			return fmt.Errorf("cleanup: delete session %s: %w", s.SessionID, err)
		}
	}

	userOauths, err := d.ListUserOauths()
	if err != nil {
		return fmt.Errorf("cleanup: list user_oauth: %w", err)
	}
	for _, uo := range *userOauths {
		if err := d.DeleteUserOauth(ctx, ac, uo.UserOauthID); err != nil {
			return fmt.Errorf("cleanup: delete user_oauth %s: %w", uo.UserOauthID, err)
		}
	}

	systemUserID := types.NullableUserID{Valid: true, ID: types.SystemUserID}
	sshKeys, err := d.ListUserSshKeys(systemUserID)
	if err != nil {
		return fmt.Errorf("cleanup: list user_ssh_keys: %w", err)
	}
	for _, sk := range *sshKeys {
		if err := d.DeleteUserSshKey(ctx, ac, sk.SshKeyID); err != nil {
			return fmt.Errorf("cleanup: delete user_ssh_key %s: %w", sk.SshKeyID, err)
		}
	}

	utility.DefaultLogger.Finfo("Bootstrap cleanup: removed all verification-only records (PostgreSQL)")
	return nil
}

/*
// InitDb initializes the database
func (d Database) InitDb(v *bool) error {
	err := d.CreateAllTables()
	if err != nil {
		return err
	}
	return nil
}

// InitDb initializes the MySQL database
func (d MysqlDatabase) InitDb(v *bool) error {
	err := d.CreateAllTables()
	if err != nil {
		return err
	}
	return nil
}

// InitDb initializes the PostgreSQL database
func (d PsqlDatabase) InitDb(v *bool) error {
	err := d.CreateAllTables()
	if err != nil {
		return err
	}
	return nil
}
*/

func (d Database) SortTables() error {
	clearTables := "DELETE FROM tables;"
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}

	tb, err := d.ListTables()
	if err != nil {
		return err
	}
	_, err = con.Exec(clearTables)
	if err != nil {
		return err
	}
	tables := *tb
	st := []string{}
	for _, t := range tables {
		st = append(st, t.Label)
	}
	sort.Strings(st)
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "sort-tables", "system")
	for _, tt := range st {
		if _, err := d.CreateTable(ctx, ac, CreateTableParams{Label: tt}); err != nil {
			return fmt.Errorf("failed to re-create table entry %s: %w", tt, err)
		}
	}

	return nil
}
func (d MysqlDatabase) SortTables() error {
	clearTables := "DELETE FROM tables;"
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}

	tb, err := d.ListTables()
	if err != nil {
		return err
	}
	_, err = con.Exec(clearTables)
	if err != nil {
		return err
	}
	tables := *tb
	st := []string{}
	for _, t := range tables {
		st = append(st, t.Label)
	}
	sort.Strings(st)
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "sort-tables", "system")
	for _, tt := range st {
		if _, err := d.CreateTable(ctx, ac, CreateTableParams{Label: tt}); err != nil {
			return fmt.Errorf("failed to re-create table entry %s: %w", tt, err)
		}
	}

	return nil
}
func (d PsqlDatabase) SortTables() error {
	clearTables := "DELETE FROM tables;"
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}

	tb, err := d.ListTables()
	if err != nil {
		return err
	}
	_, err = con.Exec(clearTables)
	if err != nil {
		return err
	}
	tables := *tb
	st := []string{}
	for _, t := range tables {
		st = append(st, t.Label)
	}
	sort.Strings(st)
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "sort-tables", "system")
	for _, tt := range st {
		if _, err := d.CreateTable(ctx, ac, CreateTableParams{Label: tt}); err != nil {
			return fmt.Errorf("failed to re-create table entry %s: %w", tt, err)
		}
	}

	return nil
}

func (d Database) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_sql.sh")
	if err != nil {
		utility.DefaultLogger.Ferror("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Ferror("failed to create temporary file: %v", err)
		return err
	}
	// Ensure the file is removed after execution.
	defer func() {
		if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Write the embedded script contents to the temporary file.
	if _, err := tmpFile.Write(script); err != nil {
		utility.DefaultLogger.Ferror("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Ferror("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "sqlite"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Ferror("failed to execute script: %v, output: %s", err, output)
		return err
	}
	return nil

}
func (d MysqlDatabase) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_mysql.sh")
	if err != nil {
		utility.DefaultLogger.Ferror("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Ferror("failed to create temporary file: %v", err)
		return err
	}
	// Ensure the file is removed after execution.
	defer func() {
		if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Write the embedded script contents to the temporary file.
	if _, err := tmpFile.Write(script); err != nil {
		utility.DefaultLogger.Ferror("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Ferror("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_User, c.Db_Password, c.Db_Name, "mysql"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Ferror("failed to execute script: %v, output: %s", err, string(output))
		return err
	}
	return nil

}
func (d PsqlDatabase) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_psql.sh")
	if err != nil {
		utility.DefaultLogger.Ferror("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Ferror("failed to create temporary file: %v", err)
		return err
	}
	// Ensure the file is removed after execution.
	defer func() {
		if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Write the embedded script contents to the temporary file.
	if _, err := tmpFile.Write(script); err != nil {
		utility.DefaultLogger.Ferror("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Ferror("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "sqlite"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Ferror("failed to execute script: %v, output: %s", err, output)
		return err
	}
	return nil

}

// Query methods for database implementations
func (d Database) Query(db *sql.DB, query string) (sql.Result, error) {
	utility.DefaultLogger.Finfo("Executing query:", query)
	return db.Exec(query)
}

func (d MysqlDatabase) Query(db *sql.DB, query string) (sql.Result, error) {
	utility.DefaultLogger.Finfo("Executing query:", query)
	return db.Exec(query)
}

func (d PsqlDatabase) Query(db *sql.DB, query string) (sql.Result, error) {
	utility.DefaultLogger.Finfo("Executing query:", query)
	return db.Exec(query)
}

// Helper functions for CRUD operations that use the Query method

// BuildInsertQuery creates an INSERT query
func BuildInsertQuery(table string, values map[string]string) string {
	var cols, vals []string

	for col, val := range values {
		cols = append(cols, col)
		vals = append(vals, fmt.Sprintf("'%s'", val))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(vals, ", "))

	return query
}

// BuildUpdateQuery creates an UPDATE query
func BuildUpdateQuery(table string, id int64, values map[string]string) string {
	var setStmts []string

	for col, val := range values {
		setStmts = append(setStmts, fmt.Sprintf("%s = '%s'", col, val))
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = %d",
		table,
		strings.Join(setStmts, ", "),
		id)

	return query
}

// BuildSelectQuery creates a SELECT query
func BuildSelectQuery(table string, id int64) string {
	query := fmt.Sprintf("SELECT * FROM \"%s\" WHERE id = %d", table, id)
	return query
}

// BuildListQuery creates a SELECT query
func BuildListQuery(table string) string {
	query := fmt.Sprintf("SELECT * FROM \"%s\"", table)
	return query
}

// BuildDeleteQuery creates a DELETE query
func BuildDeleteQuery(table string, id int64) string {
	query := fmt.Sprintf("DELETE FROM \"%s\" WHERE id = %d", table, id)
	return query
}
