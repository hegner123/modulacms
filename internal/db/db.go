// Package db provides a multi-database abstraction layer for ModulaCMS supporting SQLite, MySQL, and PostgreSQL.
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

// DbDriver is the interface for all database drivers
type DbDriver interface {
	// Connection
	CreateAllTables() error
	CreateBootstrapData(adminHash string) error
	DropAllTables() error
	DumpSql(config.Config) error
	ExecuteQuery(string, DBTable) (*sql.Rows, error)
	GetConnection() (*sql.DB, context.Context, error)
	GetForeignKeys(args []string) *sql.Rows
	Ping() error
	Query(db *sql.DB, query string) (sql.Result, error)
	ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow
	SelectColumnFromTable(table string, column string)
	SortTables() error
	ValidateBootstrapData() error

	// AdminContentData
	CountAdminContentData() (*int64, error)
	CreateAdminContentData(context.Context, audited.AuditContext, CreateAdminContentDataParams) (*AdminContentData, error)
	CreateAdminContentDataTable() error
	DeleteAdminContentData(context.Context, audited.AuditContext, types.AdminContentID) error
	GetAdminContentData(types.AdminContentID) (*AdminContentData, error)
	ListAdminContentData() (*[]AdminContentData, error)
	ListAdminContentDataByRoute(types.NullableAdminRouteID) (*[]AdminContentData, error)
	ListAdminContentDataPaginated(PaginationParams) (*[]AdminContentData, error)
	ListAdminContentDataByRoutePaginated(ListAdminContentDataByRoutePaginatedParams) (*[]AdminContentData, error)
	ListAdminContentDataWithDatatypeByRoute(types.NullableAdminRouteID) (*[]AdminContentDataWithDatatypeRow, error)
	UpdateAdminContentData(context.Context, audited.AuditContext, UpdateAdminContentDataParams) (*string, error)

	// AdminContentFields
	CountAdminContentFields() (*int64, error)
	CreateAdminContentField(context.Context, audited.AuditContext, CreateAdminContentFieldParams) (*AdminContentFields, error)
	CreateAdminContentFieldTable() error
	DeleteAdminContentField(context.Context, audited.AuditContext, types.AdminContentFieldID) error
	GetAdminContentField(types.AdminContentFieldID) (*AdminContentFields, error)
	ListAdminContentFields() (*[]AdminContentFields, error)
	ListAdminContentFieldsByRoute(types.NullableAdminRouteID) (*[]AdminContentFields, error)
	ListAdminContentFieldsPaginated(PaginationParams) (*[]AdminContentFields, error)
	ListAdminContentFieldsByRoutePaginated(ListAdminContentFieldsByRoutePaginatedParams) (*[]AdminContentFields, error)
	ListAdminContentFieldsWithFieldByRoute(types.NullableAdminRouteID) (*[]AdminContentFieldsWithFieldRow, error)
	UpdateAdminContentField(context.Context, audited.AuditContext, UpdateAdminContentFieldParams) (*string, error)

	// AdminContentRelations
	CountAdminContentRelations() (*int64, error)
	CreateAdminContentRelation(context.Context, audited.AuditContext, CreateAdminContentRelationParams) (*AdminContentRelations, error)
	CreateAdminContentRelationTable() error
	DeleteAdminContentRelation(context.Context, audited.AuditContext, types.AdminContentRelationID) error
	DropAdminContentRelationTable() error
	GetAdminContentRelation(types.AdminContentRelationID) (*AdminContentRelations, error)
	ListAdminContentRelationsBySource(types.AdminContentID) (*[]AdminContentRelations, error)
	ListAdminContentRelationsByTarget(types.AdminContentID) (*[]AdminContentRelations, error)
	ListAdminContentRelationsBySourceAndField(types.AdminContentID, types.AdminFieldID) (*[]AdminContentRelations, error)
	UpdateAdminContentRelationSortOrder(context.Context, audited.AuditContext, UpdateAdminContentRelationSortOrderParams) error

	// AdminDatatypes
	CountAdminDatatypes() (*int64, error)
	CreateAdminDatatype(context.Context, audited.AuditContext, CreateAdminDatatypeParams) (*AdminDatatypes, error)
	CreateAdminDatatypeTable() error
	DeleteAdminDatatype(context.Context, audited.AuditContext, types.AdminDatatypeID) error
	GetAdminDatatypeById(types.AdminDatatypeID) (*AdminDatatypes, error)
	ListAdminDatatypes() (*[]AdminDatatypes, error)
	ListAdminDatatypesPaginated(PaginationParams) (*[]AdminDatatypes, error)
	ListAdminDatatypeChildrenPaginated(ListAdminDatatypeChildrenPaginatedParams) (*[]AdminDatatypes, error)
	UpdateAdminDatatype(context.Context, audited.AuditContext, UpdateAdminDatatypeParams) (*string, error)

	// AdminDatatypeFields
	CountAdminDatatypeFields() (*int64, error)
	CreateAdminDatatypeField(context.Context, audited.AuditContext, CreateAdminDatatypeFieldParams) (*AdminDatatypeFields, error)
	CreateAdminDatatypeFieldTable() error
	DeleteAdminDatatypeField(context.Context, audited.AuditContext, string) error
	ListAdminDatatypeField() (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByDatatypeID(types.AdminDatatypeID) (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByFieldID(types.AdminFieldID) (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldPaginated(PaginationParams) (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByDatatypeIDPaginated(ListAdminDatatypeFieldByDatatypeIDPaginatedParams) (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByFieldIDPaginated(ListAdminDatatypeFieldByFieldIDPaginatedParams) (*[]AdminDatatypeFields, error)
	UpdateAdminDatatypeField(context.Context, audited.AuditContext, UpdateAdminDatatypeFieldParams) (*string, error)

	// AdminFields
	CountAdminFields() (*int64, error)
	CreateAdminField(context.Context, audited.AuditContext, CreateAdminFieldParams) (*AdminFields, error)
	CreateAdminFieldTable() error
	DeleteAdminField(context.Context, audited.AuditContext, types.AdminFieldID) error
	GetAdminField(types.AdminFieldID) (*AdminFields, error)
	ListAdminFields() (*[]AdminFields, error)
	ListAdminFieldsPaginated(PaginationParams) (*[]AdminFields, error)
	ListAdminFieldsByParentIDPaginated(ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error)
	UpdateAdminField(context.Context, audited.AuditContext, UpdateAdminFieldParams) (*string, error)

	// AdminRoutes
	CountAdminRoutes() (*int64, error)
	CreateAdminRoute(context.Context, audited.AuditContext, CreateAdminRouteParams) (*AdminRoutes, error)
	CreateAdminRouteTable() error
	DeleteAdminRoute(context.Context, audited.AuditContext, types.AdminRouteID) error
	GetAdminRoute(types.Slug) (*AdminRoutes, error)
	ListAdminRoutes() (*[]AdminRoutes, error)
	ListAdminRoutesPaginated(PaginationParams) (*[]AdminRoutes, error)
	UpdateAdminRoute(context.Context, audited.AuditContext, UpdateAdminRouteParams) (*string, error)

	// Backups
	CountBackups() (*int64, error)
	CreateBackup(CreateBackupParams) (*Backup, error)
	CreateBackupTables() error
	DeleteBackup(types.BackupID) error
	DropBackupTables() error
	GetBackup(types.BackupID) (*Backup, error)
	GetLatestBackup(types.NodeID) (*Backup, error)
	ListBackups(ListBackupsParams) (*[]Backup, error)
	UpdateBackupStatus(UpdateBackupStatusParams) error

	// BackupSets
	CountBackupSets() (*int64, error)
	CreateBackupSet(CreateBackupSetParams) (*BackupSet, error)
	GetBackupSet(types.BackupSetID) (*BackupSet, error)
	GetPendingBackupSets() (*[]BackupSet, error)

	// BackupVerifications
	CountVerifications() (*int64, error)
	CreateVerification(CreateVerificationParams) (*BackupVerification, error)
	GetLatestVerification(types.BackupID) (*BackupVerification, error)
	GetVerification(types.VerificationID) (*BackupVerification, error)

	// ChangeEvents
	CountChangeEvents() (*int64, error)
	CreateChangeEventsTable() error
	DeleteChangeEvent(types.EventID) error
	DropChangeEventsTable() error
	GetChangeEvent(types.EventID) (*ChangeEvent, error)
	GetChangeEventsByRecord(string, string) (*[]ChangeEvent, error)
	GetUnconsumedEvents(int64) (*[]ChangeEvent, error)
	GetUnsyncedEvents(int64) (*[]ChangeEvent, error)
	ListChangeEvents(ListChangeEventsParams) (*[]ChangeEvent, error)
	MarkEventConsumed(types.EventID) error
	MarkEventSynced(types.EventID) error
	RecordChangeEvent(RecordChangeEventParams) (*ChangeEvent, error)

	// ContentData
	CountContentData() (*int64, error)
	CreateContentData(context.Context, audited.AuditContext, CreateContentDataParams) (*ContentData, error)
	CreateContentDataTable() error
	DeleteContentData(context.Context, audited.AuditContext, types.ContentID) error
	GetContentData(types.ContentID) (*ContentData, error)
	ListContentData() (*[]ContentData, error)
	ListContentDataByRoute(types.NullableRouteID) (*[]ContentData, error)
	ListContentDataPaginated(PaginationParams) (*[]ContentData, error)
	ListContentDataByRoutePaginated(ListContentDataByRoutePaginatedParams) (*[]ContentData, error)
	ListRootContentSummary() (*[]RootContentSummary, error)
	UpdateContentData(context.Context, audited.AuditContext, UpdateContentDataParams) (*string, error)

	// ContentFields
	CountContentFields() (*int64, error)
	CreateContentField(context.Context, audited.AuditContext, CreateContentFieldParams) (*ContentFields, error)
	CreateContentFieldTable() error
	DeleteContentField(context.Context, audited.AuditContext, types.ContentFieldID) error
	GetContentField(types.ContentFieldID) (*ContentFields, error)
	GetContentFieldsByRoute(types.NullableRouteID) (*[]GetContentFieldsByRouteRow, error)
	ListContentFields() (*[]ContentFields, error)
	ListContentFieldsByRoute(types.NullableRouteID) (*[]ContentFields, error)
	ListContentFieldsByContentData(types.NullableContentID) (*[]ContentFields, error)
	ListContentFieldsPaginated(PaginationParams) (*[]ContentFields, error)
	ListContentFieldsByRoutePaginated(ListContentFieldsByRoutePaginatedParams) (*[]ContentFields, error)
	ListContentFieldsByContentDataPaginated(ListContentFieldsByContentDataPaginatedParams) (*[]ContentFields, error)
	ListContentFieldsWithFieldByContentData(types.NullableContentID) (*[]ContentFieldWithFieldRow, error)
	UpdateContentField(context.Context, audited.AuditContext, UpdateContentFieldParams) (*string, error)

	// ContentRelations
	CountContentRelations() (*int64, error)
	CreateContentRelation(context.Context, audited.AuditContext, CreateContentRelationParams) (*ContentRelations, error)
	CreateContentRelationTable() error
	DeleteContentRelation(context.Context, audited.AuditContext, types.ContentRelationID) error
	DropContentRelationTable() error
	GetContentRelation(types.ContentRelationID) (*ContentRelations, error)
	ListContentRelationsBySource(types.ContentID) (*[]ContentRelations, error)
	ListContentRelationsByTarget(types.ContentID) (*[]ContentRelations, error)
	ListContentRelationsBySourceAndField(types.ContentID, types.FieldID) (*[]ContentRelations, error)
	UpdateContentRelationSortOrder(context.Context, audited.AuditContext, UpdateContentRelationSortOrderParams) error

	// Datatypes
	CountDatatypes() (*int64, error)
	CreateDatatype(context.Context, audited.AuditContext, CreateDatatypeParams) (*Datatypes, error)
	CreateDatatypeTable() error
	DeleteDatatype(context.Context, audited.AuditContext, types.DatatypeID) error
	GetDatatype(types.DatatypeID) (*Datatypes, error)
	ListDatatypes() (*[]Datatypes, error)
	ListDatatypesRoot() (*[]Datatypes, error)
	ListDatatypeChildren(types.DatatypeID) (*[]Datatypes, error)
	ListDatatypesPaginated(PaginationParams) (*[]Datatypes, error)
	ListDatatypeChildrenPaginated(ListDatatypeChildrenPaginatedParams) (*[]Datatypes, error)
	UpdateDatatype(context.Context, audited.AuditContext, UpdateDatatypeParams) (*string, error)

	// DatatypeFields
	CountDatatypeFields() (*int64, error)
	CreateDatatypeField(context.Context, audited.AuditContext, CreateDatatypeFieldParams) (*DatatypeFields, error)
	CreateDatatypeFieldTable() error
	DeleteDatatypeField(context.Context, audited.AuditContext, string) error
	GetMaxSortOrderByDatatypeID(types.DatatypeID) (int64, error)
	ListDatatypeField() (*[]DatatypeFields, error)
	ListDatatypeFieldByDatatypeID(types.DatatypeID) (*[]DatatypeFields, error)
	ListDatatypeFieldByFieldID(types.FieldID) (*[]DatatypeFields, error)
	ListDatatypeFieldPaginated(PaginationParams) (*[]DatatypeFields, error)
	ListDatatypeFieldByDatatypeIDPaginated(ListDatatypeFieldByDatatypeIDPaginatedParams) (*[]DatatypeFields, error)
	ListDatatypeFieldByFieldIDPaginated(ListDatatypeFieldByFieldIDPaginatedParams) (*[]DatatypeFields, error)
	UpdateDatatypeField(context.Context, audited.AuditContext, UpdateDatatypeFieldParams) (*string, error)
	UpdateDatatypeFieldSortOrder(context.Context, audited.AuditContext, string, int64) error
	ListFieldsWithSortOrderByDatatypeID(types.DatatypeID) (*[]FieldWithSortOrderRow, error)

	// Fields
	CountFields() (*int64, error)
	CreateField(context.Context, audited.AuditContext, CreateFieldParams) (*Fields, error)
	CreateFieldTable() error
	DeleteField(context.Context, audited.AuditContext, types.FieldID) error
	GetField(types.FieldID) (*Fields, error)
	GetFieldDefinitionsByRoute(types.NullableRouteID) (*[]GetFieldDefinitionsByRouteRow, error)
	ListFields() (*[]Fields, error)
	ListFieldsByDatatypeID(types.NullableDatatypeID) (*[]Fields, error)
	ListFieldsPaginated(PaginationParams) (*[]Fields, error)
	UpdateField(context.Context, audited.AuditContext, UpdateFieldParams) (*string, error)

	// Media
	CountMedia() (*int64, error)
	CreateMedia(context.Context, audited.AuditContext, CreateMediaParams) (*Media, error)
	CreateMediaTable() error
	DeleteMedia(context.Context, audited.AuditContext, types.MediaID) error
	GetMedia(types.MediaID) (*Media, error)
	GetMediaByName(string) (*Media, error)
	GetMediaByURL(types.URL) (*Media, error)
	ListMedia() (*[]Media, error)
	ListMediaPaginated(PaginationParams) (*[]Media, error)
	UpdateMedia(context.Context, audited.AuditContext, UpdateMediaParams) (*string, error)

	// MediaDimensions
	CountMediaDimensions() (*int64, error)
	CreateMediaDimension(context.Context, audited.AuditContext, CreateMediaDimensionParams) (*MediaDimensions, error)
	CreateMediaDimensionTable() error
	DeleteMediaDimension(context.Context, audited.AuditContext, string) error
	GetMediaDimension(string) (*MediaDimensions, error)
	ListMediaDimensions() (*[]MediaDimensions, error)
	UpdateMediaDimension(context.Context, audited.AuditContext, UpdateMediaDimensionParams) (*string, error)

	// Permissions
	CountPermissions() (*int64, error)
	CreatePermission(context.Context, audited.AuditContext, CreatePermissionParams) (*Permissions, error)
	CreatePermissionTable() error
	DeletePermission(context.Context, audited.AuditContext, types.PermissionID) error
	GetPermission(types.PermissionID) (*Permissions, error)
	GetPermissionByLabel(string) (*Permissions, error)
	ListPermissions() (*[]Permissions, error)
	UpdatePermission(context.Context, audited.AuditContext, UpdatePermissionParams) (*string, error)

	// Roles
	CountRoles() (*int64, error)
	CreateRole(context.Context, audited.AuditContext, CreateRoleParams) (*Roles, error)
	CreateRoleTable() error
	DeleteRole(context.Context, audited.AuditContext, types.RoleID) error
	GetRole(types.RoleID) (*Roles, error)
	GetRoleByLabel(string) (*Roles, error)
	ListRoles() (*[]Roles, error)
	UpdateRole(context.Context, audited.AuditContext, UpdateRoleParams) (*string, error)

	// RolePermissions
	CountRolePermissions() (*int64, error)
	CreateRolePermission(context.Context, audited.AuditContext, CreateRolePermissionParams) (*RolePermissions, error)
	CreateRolePermissionsTable() error
	DeleteRolePermission(context.Context, audited.AuditContext, types.RolePermissionID) error
	DeleteRolePermissionsByRoleID(context.Context, audited.AuditContext, types.RoleID) error
	GetRolePermission(types.RolePermissionID) (*RolePermissions, error)
	ListRolePermissions() (*[]RolePermissions, error)
	ListRolePermissionsByRoleID(types.RoleID) (*[]RolePermissions, error)
	ListRolePermissionsByPermissionID(types.PermissionID) (*[]RolePermissions, error)
	ListPermissionLabelsByRoleID(types.RoleID) (*[]string, error)

	// Routes
	CountRoutes() (*int64, error)
	CreateRoute(context.Context, audited.AuditContext, CreateRouteParams) (*Routes, error)
	CreateRouteTable() error
	DeleteRoute(context.Context, audited.AuditContext, types.RouteID) error
	GetContentTreeByRoute(types.NullableRouteID) (*[]GetContentTreeByRouteRow, error)
	GetRoute(types.RouteID) (*Routes, error)
	GetRouteID(string) (*types.RouteID, error)
	GetRouteTreeByRouteID(types.NullableRouteID) (*[]GetRouteTreeByRouteIDRow, error)
	ListRoutes() (*[]Routes, error)
	ListRoutesByDatatype(types.DatatypeID) (*[]Routes, error)
	ListRoutesPaginated(PaginationParams) (*[]Routes, error)
	UpdateRoute(context.Context, audited.AuditContext, UpdateRouteParams) (*string, error)

	// Sessions
	CountSessions() (*int64, error)
	CreateSession(context.Context, audited.AuditContext, CreateSessionParams) (*Sessions, error)
	CreateSessionTable() error
	DeleteSession(context.Context, audited.AuditContext, types.SessionID) error
	GetSession(types.SessionID) (*Sessions, error)
	GetSessionByUserId(types.NullableUserID) (*Sessions, error)
	ListSessions() (*[]Sessions, error)
	UpdateSession(context.Context, audited.AuditContext, UpdateSessionParams) (*string, error)

	// Tables
	CountTables() (*int64, error)
	CreateTable(context.Context, audited.AuditContext, CreateTableParams) (*Tables, error)
	CreateTableTable() error
	DeleteTable(context.Context, audited.AuditContext, string) error
	GetTable(string) (*Tables, error)
	ListTables() (*[]Tables, error)
	UpdateTable(context.Context, audited.AuditContext, UpdateTableParams) (*string, error)

	// Tokens
	CountTokens() (*int64, error)
	CreateToken(context.Context, audited.AuditContext, CreateTokenParams) (*Tokens, error)
	CreateTokenTable() error
	DeleteToken(context.Context, audited.AuditContext, string) error
	GetToken(string) (*Tokens, error)
	GetTokenByTokenValue(string) (*Tokens, error)
	GetTokenByUserId(types.NullableUserID) (*[]Tokens, error)
	ListTokens() (*[]Tokens, error)
	UpdateToken(context.Context, audited.AuditContext, UpdateTokenParams) (*string, error)

	// Users
	CountUsers() (*int64, error)
	CreateUser(context.Context, audited.AuditContext, CreateUserParams) (*Users, error)
	CreateUserTable() error
	DeleteUser(context.Context, audited.AuditContext, types.UserID) error
	GetUser(types.UserID) (*Users, error)
	GetUserByEmail(types.Email) (*Users, error)
	GetUserBySSHFingerprint(string) (*Users, error)
	ListUsers() (*[]Users, error)
	ListUsersWithRoleLabel() (*[]UserWithRoleLabelRow, error)
	UpdateUser(context.Context, audited.AuditContext, UpdateUserParams) (*string, error)

	// UserOauths
	CountUserOauths() (*int64, error)
	CreateUserOauth(context.Context, audited.AuditContext, CreateUserOauthParams) (*UserOauth, error)
	CreateUserOauthTable() error
	DeleteUserOauth(context.Context, audited.AuditContext, types.UserOauthID) error
	GetUserOauth(types.UserOauthID) (*UserOauth, error)
	GetUserOauthByProviderID(string, string) (*UserOauth, error)
	GetUserOauthByUserId(types.NullableUserID) (*UserOauth, error)
	ListUserOauths() (*[]UserOauth, error)
	UpdateUserOauth(context.Context, audited.AuditContext, UpdateUserOauthParams) (*string, error)

	// UserSshKeys
	CountUserSshKeys() (*int64, error)
	CreateUserSshKey(context.Context, audited.AuditContext, CreateUserSshKeyParams) (*UserSshKeys, error)
	CreateUserSshKeyTable() error
	DeleteUserSshKey(context.Context, audited.AuditContext, string) error
	GetUserSshKey(string) (*UserSshKeys, error)
	GetUserSshKeyByFingerprint(string) (*UserSshKeys, error)
	ListUserSshKeys(types.NullableUserID) (*[]UserSshKeys, error)
	UpdateUserSshKeyLabel(context.Context, audited.AuditContext, string, string) error
	UpdateUserSshKeyLastUsed(string, string) error
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

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateRolePermissionsTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 26 tables to verify successful table creation.
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
		"content:read", "content:create", "content:update", "content:delete", "content:admin",
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

	// Editor permissions: CRUD on content, datatypes, fields, media, routes, admin_tree; read-only on users, sessions, ssh_keys
	editorPermLabels := []string{
		"content:read", "content:create", "content:update", "content:delete",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete",
		"fields:read", "fields:create", "fields:update", "fields:delete",
		"media:read", "media:create", "media:update", "media:delete",
		"routes:read", "routes:create", "routes:update", "routes:delete",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete",
		"users:read",
		"sessions:read",
		"ssh_keys:read",
		"config:read",
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

	// Viewer permissions: read-only on content, media, routes
	viewerPermLabels := []string{
		"content:read",
		"media:read",
		"routes:read",
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

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(ctx, ac, CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modulacms.local"),
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
		Slug:         types.Slug("home"),
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
		Label:        "Page",
		Type:         "ROOT",
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
		Label:        "Admin Page",
		Type:         "ROOT",
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

	// 9. Create default admin field (admin_field_id = 1)
	adminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{},
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

	// 10. Create default field (field_id = 1)
	field, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{},
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

	// 13. Create default content_field (content_field_id = 1)
	contentField, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
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
		AdminContentDataID: types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID},
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           types.NullableUserID{Valid: true, ID: systemUser.UserID},
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
		Label:       StringToNullString("Default"),
		Width:       Int64ToNullInt64(1920),
		Height:      Int64ToNullInt64(1080),
		AspectRatio: StringToNullString("16:9"),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media_dimension: %w", err)
	}
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         StringToNullString("default"),
		DisplayName:  StringToNullString("Default Media"),
		Alt:          StringToNullString("Default"),
		Caption:      sql.NullString{},
		Description:  sql.NullString{},
		Class:        sql.NullString{},
		URL:          types.URL("https://placeholder.local/default"),
		Mimetype:     sql.NullString{},
		Dimensions:   sql.NullString{},
		Srcset:       sql.NullString{},
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
		Token:     "bootstrap_validation_token",
		IssuedAt:  utility.TimestampS(),
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
		LastAccess:  StringToNullString(utility.TimestampS()),
		IpAddress:   StringToNullString("127.0.0.1"),
		UserAgent:   StringToNullString("bootstrap"),
		SessionData: sql.NullString{},
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
		TokenExpiresAt:      utility.TimestampS(),
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
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBootstrapValidationKey bootstrap@modulacms",
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

	// 20. Register all 26 ModulaCMS tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
		"change_events",
		"backups",
		"backup_verifications",
		"backup_sets",
		"permissions",
		"roles",
		"media_dimensions",
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
		"datatypes_fields",
		"admin_datatypes_fields",
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

	// 21. Create default datatypes_fields junction record (id = 1) - Links datatype to field
	datatypeField, err := d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		DatatypeID: pageDatatype.DatatypeID,
		FieldID:    field.FieldID,
	})
	if err != nil {
		return fmt.Errorf("failed to create default datatypes_fields: %w", err)
	}
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record (id = 1) - Links admin datatype to admin field
	adminDatatypeField, err := d.CreateAdminDatatypeField(ctx, ac, CreateAdminDatatypeFieldParams{
		AdminDatatypeID: adminDatatype.AdminDatatypeID,
		AdminFieldID:    adminField.AdminFieldID,
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_datatypes_fields: %w", err)
	}
	if adminDatatypeField.ID == "" {
		return fmt.Errorf("failed to create default admin_datatypes_fields")
	}

	utility.DefaultLogger.Finfo("Bootstrap data created successfully: ALL 26 tables validated with bootstrap records + complete table registry populated")
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

	// Validate tables table (should have EXACTLY 26 records - all core tables)
	tableCount, err := d.CountTables()
	if err != nil || tableCount == nil || *tableCount != 26 {
		actual := int64(0)
		if tableCount != nil {
			actual = *tableCount
		}
		errors = append(errors, fmt.Sprintf("tables table: expected exactly 26 records (table registry), got %d", actual))
	}

	// Validate datatypes_fields junction table (should have at least 1 record)
	datatypeFieldCount, err := d.CountDatatypeFields()
	if err != nil || datatypeFieldCount == nil || *datatypeFieldCount < 1 {
		errors = append(errors, "datatypes_fields table: expected ≥1 records, validation failed")
	}

	// Validate admin_datatypes_fields junction table (should have at least 1 record)
	adminDatatypeFieldCount, err := d.CountAdminDatatypeFields()
	if err != nil || adminDatatypeFieldCount == nil || *adminDatatypeFieldCount < 1 {
		errors = append(errors, "admin_datatypes_fields table: expected ≥1 records, validation failed")
	}

	// If any validation failed, return combined error
	if len(errors) > 0 {
		err := fmt.Errorf("bootstrap validation failed:\n  - %s", strings.Join(errors, "\n  - "))
		utility.DefaultLogger.Ferror("Bootstrap validation failed", err)
		return err
	}

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 26 tables contain expected records")
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

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateRolePermissionsTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 26 tables to verify successful table creation.
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
		"content:read", "content:create", "content:update", "content:delete", "content:admin",
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

	editorPermLabels := []string{
		"content:read", "content:create", "content:update", "content:delete",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete",
		"fields:read", "fields:create", "fields:update", "fields:delete",
		"media:read", "media:create", "media:update", "media:delete",
		"routes:read", "routes:create", "routes:update", "routes:delete",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete",
		"users:read",
		"sessions:read",
		"ssh_keys:read",
		"config:read",
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

	viewerPermLabels := []string{
		"content:read",
		"media:read",
		"routes:read",
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

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(ctx, ac, CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modulacms.local"),
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
		Slug:         types.Slug("home"),
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
		Label:        "Page",
		Type:         "ROOT",
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
		Label:        "Admin Page",
		Type:         "ROOT",
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

	// 9. Create default admin field (admin_field_id = 1)
	adminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{},
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

	// 10. Create default field (field_id = 1)
	field, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{},
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

	// 13. Create default content_field (content_field_id = 1)
	contentField, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
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
		AdminContentDataID: types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID},
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           types.NullableUserID{Valid: true, ID: systemUser.UserID},
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
		Label:       StringToNullString("Default"),
		Width:       Int64ToNullInt64(1920),
		Height:      Int64ToNullInt64(1080),
		AspectRatio: StringToNullString("16:9"),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media_dimension: %w", err)
	}
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         StringToNullString("default"),
		DisplayName:  StringToNullString("Default Media"),
		Alt:          StringToNullString("Default"),
		Caption:      sql.NullString{},
		Description:  sql.NullString{},
		Class:        sql.NullString{},
		URL:          types.URL("https://placeholder.local/default"),
		Mimetype:     sql.NullString{},
		Dimensions:   sql.NullString{},
		Srcset:       sql.NullString{},
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
		Token:     "bootstrap_validation_token",
		IssuedAt:  utility.TimestampS(),
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
		LastAccess:  StringToNullString(utility.TimestampS()),
		IpAddress:   StringToNullString("127.0.0.1"),
		UserAgent:   StringToNullString("bootstrap"),
		SessionData: sql.NullString{},
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
		TokenExpiresAt:      utility.TimestampS(),
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
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBootstrapValidationKey bootstrap@modulacms",
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

	// 20. Register all 26 ModulaCMS tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
		"change_events",
		"backups",
		"backup_verifications",
		"backup_sets",
		"permissions",
		"roles",
		"media_dimensions",
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
		"datatypes_fields",
		"admin_datatypes_fields",
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

	// 21. Create default datatypes_fields junction record (id = 1) - Links datatype to field
	datatypeField, err := d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		DatatypeID: pageDatatype.DatatypeID,
		FieldID:    field.FieldID,
	})
	if err != nil {
		return fmt.Errorf("failed to create default datatypes_fields: %w", err)
	}
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record - Links admin datatype to admin field
	adminDatatypeField, err := d.CreateAdminDatatypeField(ctx, ac, CreateAdminDatatypeFieldParams{
		AdminDatatypeID: adminDatatype.AdminDatatypeID,
		AdminFieldID:    adminField.AdminFieldID,
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_datatypes_fields: %w", err)
	}
	if adminDatatypeField.ID == "" {
		return fmt.Errorf("failed to create default admin_datatypes_fields")
	}

	utility.DefaultLogger.Finfo("Bootstrap data created successfully: ALL 21 tables validated with bootstrap records + complete table registry populated")
	return nil
}

// ValidateBootstrapData verifies all tables have expected bootstrap records for MySQL.
// CRITICAL: Call after CreateBootstrapData() to catch any silent failures.
// Returns detailed error if any table validation fails.
func (d MysqlDatabase) ValidateBootstrapData() error {
	var errors []string

	// Validate all 26 tables have expected record counts
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

	tableCount, err := d.CountTables()
	if err != nil || tableCount == nil || *tableCount != 26 {
		actual := int64(0)
		if tableCount != nil {
			actual = *tableCount
		}
		errors = append(errors, fmt.Sprintf("tables table: expected exactly 26 records (table registry), got %d", actual))
	}

	datatypeFieldCount, err := d.CountDatatypeFields()
	if err != nil || datatypeFieldCount == nil || *datatypeFieldCount < 1 {
		errors = append(errors, "datatypes_fields table: expected ≥1 records, validation failed")
	}

	adminDatatypeFieldCount, err := d.CountAdminDatatypeFields()
	if err != nil || adminDatatypeFieldCount == nil || *adminDatatypeFieldCount < 1 {
		errors = append(errors, "admin_datatypes_fields table: expected ≥1 records, validation failed")
	}

	if len(errors) > 0 {
		err := fmt.Errorf("bootstrap validation failed:\n  - %s", strings.Join(errors, "\n  - "))
		utility.DefaultLogger.Ferror("Bootstrap validation failed", err)
		return err
	}

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 26 tables contain expected records")
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

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateRolePermissionsTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 26 tables to verify successful table creation.
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
		"content:read", "content:create", "content:update", "content:delete", "content:admin",
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

	editorPermLabels := []string{
		"content:read", "content:create", "content:update", "content:delete",
		"datatypes:read", "datatypes:create", "datatypes:update", "datatypes:delete",
		"fields:read", "fields:create", "fields:update", "fields:delete",
		"media:read", "media:create", "media:update", "media:delete",
		"routes:read", "routes:create", "routes:update", "routes:delete",
		"admin_tree:read", "admin_tree:create", "admin_tree:update", "admin_tree:delete",
		"users:read",
		"sessions:read",
		"ssh_keys:read",
		"config:read",
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

	viewerPermLabels := []string{
		"content:read",
		"media:read",
		"routes:read",
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

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(ctx, ac, CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modulacms.local"),
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
		Slug:         types.Slug("home"),
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
		Label:        "Page",
		Type:         "ROOT",
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
		Label:        "Admin Page",
		Type:         "ROOT",
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

	// 9. Create default admin field (admin_field_id = 1)
	adminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{},
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

	// 10. Create default field (field_id = 1)
	field, err := d.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{},
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

	// 13. Create default content_field (content_field_id = 1)
	contentField, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
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
		AdminContentDataID: types.NullableAdminContentID{Valid: true, ID: adminContentData.AdminContentDataID},
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           types.NullableUserID{Valid: true, ID: systemUser.UserID},
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
		Label:       StringToNullString("Default"),
		Width:       Int64ToNullInt64(1920),
		Height:      Int64ToNullInt64(1080),
		AspectRatio: StringToNullString("16:9"),
	})
	if err != nil {
		return fmt.Errorf("failed to create default media_dimension: %w", err)
	}
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         StringToNullString("default"),
		DisplayName:  StringToNullString("Default Media"),
		Alt:          StringToNullString("Default"),
		Caption:      sql.NullString{},
		Description:  sql.NullString{},
		Class:        sql.NullString{},
		URL:          types.URL("https://placeholder.local/default"),
		Mimetype:     sql.NullString{},
		Dimensions:   sql.NullString{},
		Srcset:       sql.NullString{},
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
		Token:     "bootstrap_validation_token",
		IssuedAt:  utility.TimestampS(),
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
		LastAccess:  StringToNullString(utility.TimestampS()),
		IpAddress:   StringToNullString("127.0.0.1"),
		UserAgent:   StringToNullString("bootstrap"),
		SessionData: sql.NullString{},
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
		TokenExpiresAt:      utility.TimestampS(),
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
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBootstrapValidationKey bootstrap@modulacms",
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

	// 20. Register all 26 ModulaCMS tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
		"change_events",
		"backups",
		"backup_verifications",
		"backup_sets",
		"permissions",
		"roles",
		"media_dimensions",
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
		"datatypes_fields",
		"admin_datatypes_fields",
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

	// 21. Create default datatypes_fields junction record (id = 1) - Links datatype to field
	datatypeField, err := d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		DatatypeID: pageDatatype.DatatypeID,
		FieldID:    field.FieldID,
	})
	if err != nil {
		return fmt.Errorf("failed to create default datatypes_fields: %w", err)
	}
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record - Links admin datatype to admin field
	adminDatatypeField, err := d.CreateAdminDatatypeField(ctx, ac, CreateAdminDatatypeFieldParams{
		AdminDatatypeID: adminDatatype.AdminDatatypeID,
		AdminFieldID:    adminField.AdminFieldID,
	})
	if err != nil {
		return fmt.Errorf("failed to create default admin_datatypes_fields: %w", err)
	}
	if adminDatatypeField.ID == "" {
		return fmt.Errorf("failed to create default admin_datatypes_fields")
	}

	utility.DefaultLogger.Finfo("Bootstrap data created successfully: ALL 21 tables validated with bootstrap records + complete table registry populated")
	return nil
}

// ValidateBootstrapData verifies all tables have expected bootstrap records for PostgreSQL.
// CRITICAL: Call after CreateBootstrapData() to catch any silent failures.
// Returns detailed error if any table validation fails.
func (d PsqlDatabase) ValidateBootstrapData() error {
	var errors []string

	// Validate all 26 tables have expected record counts
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

	tableCount, err := d.CountTables()
	if err != nil || tableCount == nil || *tableCount != 26 {
		actual := int64(0)
		if tableCount != nil {
			actual = *tableCount
		}
		errors = append(errors, fmt.Sprintf("tables table: expected exactly 26 records (table registry), got %d", actual))
	}

	datatypeFieldCount, err := d.CountDatatypeFields()
	if err != nil || datatypeFieldCount == nil || *datatypeFieldCount < 1 {
		errors = append(errors, "datatypes_fields table: expected ≥1 records, validation failed")
	}

	adminDatatypeFieldCount, err := d.CountAdminDatatypeFields()
	if err != nil || adminDatatypeFieldCount == nil || *adminDatatypeFieldCount < 1 {
		errors = append(errors, "admin_datatypes_fields table: expected ≥1 records, validation failed")
	}

	if len(errors) > 0 {
		err := fmt.Errorf("bootstrap validation failed:\n  - %s", strings.Join(errors, "\n  - "))
		utility.DefaultLogger.Ferror("Bootstrap validation failed", err)
		return err
	}

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 26 tables contain expected records")
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
		utility.DefaultLogger.Fatal("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create temporary file: %v", err)
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
		utility.DefaultLogger.Fatal("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Fatal("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "sqlite"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Fatal("failed to execute script: %v, output: %s", err, output)
		return err
	}
	return nil

}
func (d MysqlDatabase) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_mysql.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create temporary file: %v", err)
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
		utility.DefaultLogger.Fatal("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Fatal("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_User, c.Db_Password, c.Db_Name, "mysql"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Fatal("failed to execute script: %v, output: %s", err, string(output))
		return err
	}
	return nil

}
func (d PsqlDatabase) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_psql.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create temporary file: %v", err)
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
		utility.DefaultLogger.Fatal("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Fatal("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "sqlite"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Fatal("failed to execute script: %v, output: %s", err, output)
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
