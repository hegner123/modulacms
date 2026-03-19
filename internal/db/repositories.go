// Package db defines repository interfaces that decompose the DbDriver contract
// into focused, domain-specific groups. DbDriver embeds all repositories so existing
// consumers that accept DbDriver continue to work unchanged. Narrower consumers
// (like middleware.PermissionCache) can accept a single repository interface.
package db

import (
	"context"
	"database/sql"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// SchemaRepository manages database lifecycle operations: creating tables,
// seeding bootstrap data, dropping schema, and exporting SQL dumps.
type SchemaRepository interface {
	CreateAllTables() error
	CreateBootstrapData(adminHash string) error
	CleanupBootstrapData() error
	DropAllTables() error
	DumpSql(config.Config) error
	SortTables() error
	ValidateBootstrapData() error
}

// ConnectionRepository provides raw database access for queries, introspection,
// and health checks. Used by the table browser and diagnostic tools.
type ConnectionRepository interface {
	ExecuteQuery(string, DBTable) (*sql.Rows, error)
	GetConnection() (*sql.DB, context.Context, error)
	GetForeignKeys(args []string) *sql.Rows
	Ping() error
	Query(db *sql.DB, query string) (sql.Result, error)
	ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow
	SelectColumnFromTable(table string, column string)
}

// ContentDataRepository manages content data, content relations, and content
// versions. These entities share a lifecycle: versions and relations are
// meaningless without their parent content record.
type ContentDataRepository interface {
	// ContentData
	CountContentData() (*int64, error)
	CreateContentData(context.Context, audited.AuditContext, CreateContentDataParams) (*ContentData, error)
	CreateContentDataTable() error
	DeleteContentData(context.Context, audited.AuditContext, types.ContentID) error
	GetContentData(types.ContentID) (*ContentData, error)
	ListContentData() (*[]ContentData, error)
	ListContentDataByRoute(types.NullableRouteID) (*[]ContentData, error)
	ListContentDataByDatatypeID(types.DatatypeID) (*[]ContentData, error)
	ListContentDataByRootID(types.NullableContentID) (*[]ContentData, error)
	ListContentDataGlobal() (*[]ContentData, error)
	ListContentDataPaginated(PaginationParams) (*[]ContentData, error)
	ListContentDataTopLevelPaginated(PaginationParams) (*[]ContentDataTopLevel, error)
	ListContentDataTopLevelPaginatedByStatus(PaginationParams, types.ContentStatus) (*[]ContentDataTopLevel, error)
	ListContentDataByRoutePaginated(ListContentDataByRoutePaginatedParams) (*[]ContentData, error)
	CountContentDataTopLevel() (*int64, error)
	CountContentDataTopLevelByStatus(types.ContentStatus) (*int64, error)
	GetContentDataDescendants(context.Context, types.ContentID) (*[]ContentData, error)
	ListRootContentSummary() (*[]RootContentSummary, error)
	UpdateContentData(context.Context, audited.AuditContext, UpdateContentDataParams) (*string, error)
	UpdateContentDataPublishMeta(context.Context, UpdateContentDataPublishMetaParams) error
	UpdateContentDataWithRevision(context.Context, UpdateContentDataWithRevisionParams) error
	UpdateContentDataSchedule(context.Context, UpdateContentDataScheduleParams) error
	ClearContentDataSchedule(context.Context, ClearContentDataScheduleParams) error
	ListContentDataDueForPublish(types.Timestamp) (*[]ContentData, error)
	ReassignContentDataAuthor(context.Context, types.UserID, types.UserID) error
	CountContentDataByAuthor(context.Context, types.UserID) (int64, error)

	// ContentRelations
	CountContentRelations() (*int64, error)
	CreateContentRelation(context.Context, audited.AuditContext, CreateContentRelationParams) (*ContentRelations, error)
	CreateContentRelationTable() error
	DeleteContentRelation(context.Context, audited.AuditContext, types.ContentRelationID) error
	DropContentRelationTable() error
	GetContentRelation(types.ContentRelationID) (*ContentRelations, error)
	ListContentRelations() (*[]ContentRelations, error)
	ListContentRelationsBySource(types.ContentID) (*[]ContentRelations, error)
	ListContentRelationsByTarget(types.ContentID) (*[]ContentRelations, error)
	ListContentRelationsBySourceAndField(types.ContentID, types.FieldID) (*[]ContentRelations, error)
	UpdateContentRelationSortOrder(context.Context, audited.AuditContext, UpdateContentRelationSortOrderParams) error

	// ContentVersions
	CountContentVersions() (*int64, error)
	CountContentVersionsByContent(types.ContentID) (*int64, error)
	CreateContentVersion(context.Context, audited.AuditContext, CreateContentVersionParams) (*ContentVersion, error)
	CreateContentVersionTable() error
	DropContentVersionTable() error
	DeleteContentVersion(context.Context, audited.AuditContext, types.ContentVersionID) error
	GetContentVersion(types.ContentVersionID) (*ContentVersion, error)
	GetPublishedSnapshot(types.ContentID, string) (*ContentVersion, error)
	ListContentVersionsByContent(types.ContentID) (*[]ContentVersion, error)
	ListContentVersionsByContentLocale(types.ContentID, string) (*[]ContentVersion, error)
	ClearPublishedFlag(types.ContentID, string) error
	GetMaxVersionNumber(types.ContentID, string) (int64, error)
	PruneOldVersions(types.ContentID, string, int64) error
}

// ContentFieldRepository manages content field values.
type ContentFieldRepository interface {
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
	ListContentFieldsByContentDataAndLocale(types.NullableContentID, string) (*[]ContentFields, error)
	ListContentFieldsByContentDataIDs(context.Context, []types.ContentID, string) (*[]ContentFields, error)
	ListContentFieldsByRootID(types.NullableContentID) (*[]ContentFields, error)
	ListContentFieldsByRootIDAndLocale(types.NullableContentID, string) (*[]ContentFields, error)
	ListContentFieldsByRouteAndLocale(types.NullableRouteID, string) (*[]ContentFields, error)
	UpdateContentField(context.Context, audited.AuditContext, UpdateContentFieldParams) (*string, error)
}

// AdminContentDataRepository manages admin-side content data, relations, and
// versions. Mirrors ContentDataRepository for the admin schema.
type AdminContentDataRepository interface {
	// AdminContentData
	CountAdminContentData() (*int64, error)
	CreateAdminContentData(context.Context, audited.AuditContext, CreateAdminContentDataParams) (*AdminContentData, error)
	CreateAdminContentDataTable() error
	DeleteAdminContentData(context.Context, audited.AuditContext, types.AdminContentID) error
	GetAdminContentData(types.AdminContentID) (*AdminContentData, error)
	ListAdminContentData() (*[]AdminContentData, error)
	ListAdminContentDataByRoute(types.NullableAdminRouteID) (*[]AdminContentData, error)
	ListAdminContentDataByRootID(types.NullableAdminContentID) (*[]AdminContentData, error)
	ListAdminContentDataPaginated(PaginationParams) (*[]AdminContentData, error)
	ListAdminContentDataTopLevelPaginated(PaginationParams) (*[]AdminContentDataTopLevel, error)
	ListAdminContentDataByRoutePaginated(ListAdminContentDataByRoutePaginatedParams) (*[]AdminContentData, error)
	ListAdminContentDataWithDatatypeByRoute(types.NullableAdminRouteID) (*[]AdminContentDataWithDatatypeRow, error)
	ListAdminContentDataWithDatatypeByRootID(types.NullableAdminContentID) (*[]AdminContentDataWithDatatypeRow, error)
	CountAdminContentDataTopLevel() (*int64, error)
	UpdateAdminContentData(context.Context, audited.AuditContext, UpdateAdminContentDataParams) (*string, error)
	UpdateAdminContentDataPublishMeta(context.Context, UpdateAdminContentDataPublishMetaParams) error
	UpdateAdminContentDataWithRevision(context.Context, UpdateAdminContentDataWithRevisionParams) error
	UpdateAdminContentDataSchedule(context.Context, UpdateAdminContentDataScheduleParams) error
	ClearAdminContentDataSchedule(context.Context, ClearAdminContentDataScheduleParams) error
	ListAdminContentDataDueForPublish(types.Timestamp) (*[]AdminContentData, error)
	GetAdminContentDataDescendants(context.Context, types.AdminContentID) (*[]AdminContentData, error)
	ReassignAdminContentDataAuthor(context.Context, types.UserID, types.UserID) error
	CountAdminContentDataByAuthor(context.Context, types.UserID) (int64, error)

	// AdminContentRelations
	CountAdminContentRelations() (*int64, error)
	CreateAdminContentRelation(context.Context, audited.AuditContext, CreateAdminContentRelationParams) (*AdminContentRelations, error)
	CreateAdminContentRelationTable() error
	DeleteAdminContentRelation(context.Context, audited.AuditContext, types.AdminContentRelationID) error
	DropAdminContentRelationTable() error
	GetAdminContentRelation(types.AdminContentRelationID) (*AdminContentRelations, error)
	ListAdminContentRelations() (*[]AdminContentRelations, error)
	ListAdminContentRelationsBySource(types.AdminContentID) (*[]AdminContentRelations, error)
	ListAdminContentRelationsByTarget(types.AdminContentID) (*[]AdminContentRelations, error)
	ListAdminContentRelationsBySourceAndField(types.AdminContentID, types.AdminFieldID) (*[]AdminContentRelations, error)
	UpdateAdminContentRelationSortOrder(context.Context, audited.AuditContext, UpdateAdminContentRelationSortOrderParams) error

	// AdminContentVersions
	CountAdminContentVersions() (*int64, error)
	CountAdminContentVersionsByContent(types.AdminContentID) (*int64, error)
	CreateAdminContentVersion(context.Context, audited.AuditContext, CreateAdminContentVersionParams) (*AdminContentVersion, error)
	CreateAdminContentVersionTable() error
	DropAdminContentVersionTable() error
	DeleteAdminContentVersion(context.Context, audited.AuditContext, types.AdminContentVersionID) error
	GetAdminContentVersion(types.AdminContentVersionID) (*AdminContentVersion, error)
	GetAdminPublishedSnapshot(types.AdminContentID, string) (*AdminContentVersion, error)
	ListAdminContentVersionsByContent(types.AdminContentID) (*[]AdminContentVersion, error)
	ListAdminContentVersionsByContentLocale(types.AdminContentID, string) (*[]AdminContentVersion, error)
	ClearAdminPublishedFlag(types.AdminContentID, string) error
	GetAdminMaxVersionNumber(types.AdminContentID, string) (int64, error)
	PruneAdminOldVersions(types.AdminContentID, string, int64) error
}

// AdminContentFieldRepository manages admin-side content field values.
type AdminContentFieldRepository interface {
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
	ListAdminContentFieldsByContentData(types.NullableAdminContentID) (*[]AdminContentFields, error)
	ListAdminContentFieldsByContentDataAndLocale(types.NullableAdminContentID, string) (*[]AdminContentFields, error)
	ListAdminContentFieldsByRootID(types.NullableAdminContentID) (*[]AdminContentFields, error)
	ListAdminContentFieldsByRootIDAndLocale(types.NullableAdminContentID, string) (*[]AdminContentFields, error)
	ListAdminContentFieldsByRouteAndLocale(types.NullableAdminRouteID, string) (*[]AdminContentFields, error)
	ListAdminContentFieldsWithFieldByContentData(types.NullableAdminContentID) (*[]AdminContentFieldsWithFieldRow, error)
	ListAdminContentFieldsByContentDataIDs(context.Context, []types.AdminContentID, string) (*[]AdminContentFields, error)
	UpdateAdminContentField(context.Context, audited.AuditContext, UpdateAdminContentFieldParams) (*string, error)
}

// DatatypeRepository manages datatype definitions and their sort ordering.
type DatatypeRepository interface {
	CountDatatypes() (*int64, error)
	CreateDatatype(context.Context, audited.AuditContext, CreateDatatypeParams) (*Datatypes, error)
	CreateDatatypeTable() error
	DeleteDatatype(context.Context, audited.AuditContext, types.DatatypeID) error
	GetDatatype(types.DatatypeID) (*Datatypes, error)
	GetDatatypeByType(string) (*Datatypes, error)
	GetDatatypeByName(string) (*Datatypes, error)
	ListDatatypes() (*[]Datatypes, error)
	ListDatatypesRoot() (*[]Datatypes, error)
	ListDatatypesGlobal() (*[]Datatypes, error)
	ListDatatypeChildren(types.DatatypeID) (*[]Datatypes, error)
	ListDatatypesPaginated(PaginationParams) (*[]Datatypes, error)
	ListDatatypeChildrenPaginated(ListDatatypeChildrenPaginatedParams) (*[]Datatypes, error)
	UpdateDatatype(context.Context, audited.AuditContext, UpdateDatatypeParams) (*string, error)
	UpdateDatatypeSortOrder(context.Context, audited.AuditContext, UpdateDatatypeSortOrderParams) error
	GetMaxDatatypeSortOrder(types.NullableDatatypeID) (int64, error)
	ReassignDatatypeAuthor(context.Context, types.UserID, types.UserID) error
	CountDatatypesByAuthor(context.Context, types.UserID) (int64, error)
}

// AdminDatatypeRepository manages admin-side datatype definitions.
type AdminDatatypeRepository interface {
	CountAdminDatatypes() (*int64, error)
	CreateAdminDatatype(context.Context, audited.AuditContext, CreateAdminDatatypeParams) (*AdminDatatypes, error)
	CreateAdminDatatypeTable() error
	DeleteAdminDatatype(context.Context, audited.AuditContext, types.AdminDatatypeID) error
	GetAdminDatatypeById(types.AdminDatatypeID) (*AdminDatatypes, error)
	ListAdminDatatypes() (*[]AdminDatatypes, error)
	ListAdminDatatypesPaginated(PaginationParams) (*[]AdminDatatypes, error)
	ListAdminDatatypeChildrenPaginated(ListAdminDatatypeChildrenPaginatedParams) (*[]AdminDatatypes, error)
	UpdateAdminDatatype(context.Context, audited.AuditContext, UpdateAdminDatatypeParams) (*string, error)
	UpdateAdminDatatypeSortOrder(context.Context, audited.AuditContext, UpdateAdminDatatypeSortOrderParams) error
	GetMaxAdminDatatypeSortOrder(types.NullableAdminDatatypeID) (int64, error)
}

// FieldRepository manages field definitions, field types, and field sort ordering.
// FieldTypes are included because they are a lookup table for field validation
// and are never queried outside the field creation/update path.
type FieldRepository interface {
	// Fields
	ListFieldsWithSortOrderByDatatypeID(types.NullableDatatypeID) (*[]FieldWithSortOrderRow, error)
	CountFields() (*int64, error)
	CreateField(context.Context, audited.AuditContext, CreateFieldParams) (*Fields, error)
	CreateFieldTable() error
	DeleteField(context.Context, audited.AuditContext, types.FieldID) error
	GetField(types.FieldID) (*Fields, error)
	GetFieldsByIDs(ctx context.Context, ids []types.FieldID) ([]Fields, error)
	GetFieldDefinitionsByRoute(types.NullableRouteID) (*[]GetFieldDefinitionsByRouteRow, error)
	ListFields() (*[]Fields, error)
	ListFieldsByDatatypeID(types.NullableDatatypeID) (*[]Fields, error)
	ListFieldsPaginated(PaginationParams) (*[]Fields, error)
	UpdateField(context.Context, audited.AuditContext, UpdateFieldParams) (*string, error)
	UpdateFieldSortOrder(context.Context, audited.AuditContext, UpdateFieldSortOrderParams) error
	GetMaxSortOrderByParentID(types.NullableDatatypeID) (int64, error)

	// FieldTypes
	CountFieldTypes() (*int64, error)
	CreateFieldType(context.Context, audited.AuditContext, CreateFieldTypeParams) (*FieldTypes, error)
	CreateFieldTypeTable() error
	DeleteFieldType(context.Context, audited.AuditContext, types.FieldTypeID) error
	GetFieldType(types.FieldTypeID) (*FieldTypes, error)
	GetFieldTypeByType(string) (*FieldTypes, error)
	ListFieldTypes() (*[]FieldTypes, error)
	UpdateFieldType(context.Context, audited.AuditContext, UpdateFieldTypeParams) (*string, error)
}

// AdminFieldRepository manages admin-side field definitions and admin field types.
type AdminFieldRepository interface {
	// AdminFields
	CountAdminFields() (*int64, error)
	CreateAdminField(context.Context, audited.AuditContext, CreateAdminFieldParams) (*AdminFields, error)
	CreateAdminFieldTable() error
	DeleteAdminField(context.Context, audited.AuditContext, types.AdminFieldID) error
	GetAdminField(types.AdminFieldID) (*AdminFields, error)
	ListAdminFields() (*[]AdminFields, error)
	ListAdminFieldsPaginated(PaginationParams) (*[]AdminFields, error)
	ListAdminFieldsByParentIDPaginated(ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error)
	ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID) (*[]AdminFields, error)
	UpdateAdminField(context.Context, audited.AuditContext, UpdateAdminFieldParams) (*string, error)
	UpdateAdminFieldSortOrder(context.Context, audited.AuditContext, UpdateAdminFieldSortOrderParams) error
	GetMaxAdminSortOrderByParentID(types.NullableAdminDatatypeID) (int64, error)

	// AdminFieldTypes
	CountAdminFieldTypes() (*int64, error)
	CreateAdminFieldType(context.Context, audited.AuditContext, CreateAdminFieldTypeParams) (*AdminFieldTypes, error)
	CreateAdminFieldTypeTable() error
	DeleteAdminFieldType(context.Context, audited.AuditContext, types.AdminFieldTypeID) error
	GetAdminFieldType(types.AdminFieldTypeID) (*AdminFieldTypes, error)
	GetAdminFieldTypeByType(string) (*AdminFieldTypes, error)
	ListAdminFieldTypes() (*[]AdminFieldTypes, error)
	UpdateAdminFieldType(context.Context, audited.AuditContext, UpdateAdminFieldTypeParams) (*string, error)
}

// RouteRepository manages route definitions and route-based tree queries.
type RouteRepository interface {
	CountRoutes() (*int64, error)
	CreateRoute(context.Context, audited.AuditContext, CreateRouteParams) (*Routes, error)
	CreateRouteTable() error
	DeleteRoute(context.Context, audited.AuditContext, types.RouteID) error
	GetContentTreeByRoute(types.NullableRouteID) (*[]GetContentTreeByRouteRow, error)
	GetContentTreeByRootID(types.NullableContentID) (*[]GetContentTreeByRouteRow, error)
	GetRoute(types.RouteID) (*Routes, error)
	GetRouteID(string) (*types.RouteID, error)
	GetRouteTreeByRouteID(types.NullableRouteID) (*[]GetRouteTreeByRouteIDRow, error)
	ListRoutes() (*[]Routes, error)
	ListRoutesByDatatype(types.DatatypeID) (*[]Routes, error)
	ListRoutesPaginated(PaginationParams) (*[]Routes, error)
	UpdateRoute(context.Context, audited.AuditContext, UpdateRouteParams) (*string, error)
}

// AdminRouteRepository manages admin-side route definitions.
type AdminRouteRepository interface {
	CountAdminRoutes() (*int64, error)
	CreateAdminRoute(context.Context, audited.AuditContext, CreateAdminRouteParams) (*AdminRoutes, error)
	CreateAdminRouteTable() error
	DeleteAdminRoute(context.Context, audited.AuditContext, types.AdminRouteID) error
	GetAdminRoute(types.Slug) (*AdminRoutes, error)
	GetAdminRouteByID(types.AdminRouteID) (*AdminRoutes, error)
	ListAdminRoutes() (*[]AdminRoutes, error)
	ListAdminRoutesPaginated(PaginationParams) (*[]AdminRoutes, error)
	UpdateAdminRoute(context.Context, audited.AuditContext, UpdateAdminRouteParams) (*string, error)
}

// MediaRepository manages media assets and their dimension presets.
type MediaRepository interface {
	// Media
	CountMedia() (*int64, error)
	CountMediaByFolder(types.NullableMediaFolderID) (*int64, error)
	CountMediaUnfiled() (*int64, error)
	CreateMedia(context.Context, audited.AuditContext, CreateMediaParams) (*Media, error)
	CreateMediaTable() error
	DeleteMedia(context.Context, audited.AuditContext, types.MediaID) error
	GetMedia(types.MediaID) (*Media, error)
	GetMediaByName(string) (*Media, error)
	GetMediaByURL(types.URL) (*Media, error)
	ListMedia() (*[]Media, error)
	ListMediaByFolder(types.NullableMediaFolderID) (*[]Media, error)
	ListMediaByFolderPaginated(ListMediaByFolderPaginatedParams) (*[]Media, error)
	ListMediaPaginated(PaginationParams) (*[]Media, error)
	ListMediaUnfiled() (*[]Media, error)
	ListMediaUnfiledPaginated(PaginationParams) (*[]Media, error)
	MoveMediaToFolder(context.Context, audited.AuditContext, MoveMediaToFolderParams) error
	UpdateMedia(context.Context, audited.AuditContext, UpdateMediaParams) (*string, error)

	// MediaDimensions
	CountMediaDimensions() (*int64, error)
	CreateMediaDimension(context.Context, audited.AuditContext, CreateMediaDimensionParams) (*MediaDimensions, error)
	CreateMediaDimensionTable() error
	DeleteMediaDimension(context.Context, audited.AuditContext, string) error
	GetMediaDimension(string) (*MediaDimensions, error)
	ListMediaDimensions() (*[]MediaDimensions, error)
	UpdateMediaDimension(context.Context, audited.AuditContext, UpdateMediaDimensionParams) (*string, error)
}

// UserRepository manages user accounts, OAuth links, and SSH keys.
type UserRepository interface {
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

// AuthRepository manages sessions and API tokens.
type AuthRepository interface {
	// Sessions
	CountSessions() (*int64, error)
	CreateSession(context.Context, audited.AuditContext, CreateSessionParams) (*Sessions, error)
	CreateSessionTable() error
	DeleteSession(context.Context, audited.AuditContext, types.SessionID) error
	GetSession(types.SessionID) (*Sessions, error)
	GetSessionByUserId(types.NullableUserID) (*Sessions, error)
	GetSessionByToken(string) (*Sessions, error)
	ListSessions() (*[]Sessions, error)
	UpdateSession(context.Context, audited.AuditContext, UpdateSessionParams) (*string, error)

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
}

// RBACRepository manages roles, permissions, and the role-permission junction.
type RBACRepository interface {
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
}

// BackupRepository manages backups, backup sets, and backup verifications.
type BackupRepository interface {
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
}

// ChangeEventRepository manages the change event audit log for replication.
type ChangeEventRepository interface {
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
}

// TableRepository manages the table registry for plugin-created tables.
type TableRepository interface {
	CountTables() (*int64, error)
	CreateTable(context.Context, audited.AuditContext, CreateTableParams) (*Tables, error)
	CreateTableTable() error
	DeleteTable(context.Context, audited.AuditContext, string) error
	GetTable(string) (*Tables, error)
	ListTables() (*[]Tables, error)
	UpdateTable(context.Context, audited.AuditContext, UpdateTableParams) (*string, error)
}

// PluginRepository manages plugins and their associated pipelines.
// Pipelines have a FK to plugins and share the plugin lifecycle.
type PluginRepository interface {
	// Plugins
	CountPlugins() (*int64, error)
	CreatePlugin(context.Context, audited.AuditContext, CreatePluginParams) (*Plugin, error)
	CreatePluginTable() error
	DeletePlugin(context.Context, audited.AuditContext, types.PluginID) error
	GetPlugin(types.PluginID) (*Plugin, error)
	GetPluginByName(string) (*Plugin, error)
	ListPlugins() (*[]Plugin, error)
	ListPluginsByStatus(types.PluginStatus) (*[]Plugin, error)
	UpdatePlugin(context.Context, audited.AuditContext, UpdatePluginParams) error
	UpdatePluginStatus(context.Context, audited.AuditContext, types.PluginID, types.PluginStatus) error

	// Pipelines
	CountPipelines() (*int64, error)
	CreatePipeline(context.Context, audited.AuditContext, CreatePipelineParams) (*Pipeline, error)
	CreatePipelineTable() error
	DeletePipeline(context.Context, audited.AuditContext, types.PipelineID) error
	DeletePipelinesByPluginID(context.Context, audited.AuditContext, types.PluginID) error
	GetPipeline(types.PipelineID) (*Pipeline, error)
	ListPipelines() (*[]Pipeline, error)
	ListPipelinesByTable(string) (*[]Pipeline, error)
	ListPipelinesByPluginID(types.PluginID) (*[]Pipeline, error)
	ListPipelinesByTableOperation(string, string) (*[]Pipeline, error)
	ListEnabledPipelines() (*[]Pipeline, error)
	UpdatePipeline(context.Context, audited.AuditContext, UpdatePipelineParams) error
	UpdatePipelineEnabled(context.Context, audited.AuditContext, types.PipelineID, bool) error
}

// LocaleRepository manages locale definitions for internationalization.
type LocaleRepository interface {
	CountLocales() (*int64, error)
	CreateLocale(context.Context, audited.AuditContext, CreateLocaleParams) (*Locale, error)
	CreateLocaleTable() error
	DeleteLocale(context.Context, audited.AuditContext, types.LocaleID) error
	GetLocale(types.LocaleID) (*Locale, error)
	GetLocaleByCode(string) (*Locale, error)
	GetDefaultLocale() (*Locale, error)
	ListLocales() (*[]Locale, error)
	ListEnabledLocales() (*[]Locale, error)
	ListLocalesPaginated(PaginationParams) (*[]Locale, error)
	UpdateLocale(context.Context, audited.AuditContext, UpdateLocaleParams) error
	ClearDefaultLocale(context.Context) error
}

// WebhookRepository manages webhooks and their delivery records.
// WebhookDeliveries have a FK to webhooks and share the webhook lifecycle.
type WebhookRepository interface {
	// Webhooks
	CountWebhooks() (*int64, error)
	CreateWebhook(context.Context, audited.AuditContext, CreateWebhookParams) (*Webhook, error)
	CreateWebhookTable() error
	DeleteWebhook(context.Context, audited.AuditContext, types.WebhookID) error
	GetWebhook(types.WebhookID) (*Webhook, error)
	ListWebhooks() (*[]Webhook, error)
	ListActiveWebhooks() (*[]Webhook, error)
	ListWebhooksPaginated(PaginationParams) (*[]Webhook, error)
	UpdateWebhook(context.Context, audited.AuditContext, UpdateWebhookParams) error

	// WebhookDeliveries
	CountWebhookDeliveries() (*int64, error)
	CreateWebhookDelivery(context.Context, CreateWebhookDeliveryParams) (*WebhookDelivery, error)
	CreateWebhookDeliveryTable() error
	DeleteWebhookDelivery(context.Context, types.WebhookDeliveryID) error
	GetWebhookDelivery(types.WebhookDeliveryID) (*WebhookDelivery, error)
	ListWebhookDeliveries() (*[]WebhookDelivery, error)
	ListWebhookDeliveriesByWebhook(types.WebhookID) (*[]WebhookDelivery, error)
	ListPendingRetries(types.Timestamp, int64) (*[]WebhookDelivery, error)
	UpdateWebhookDeliveryStatus(context.Context, UpdateWebhookDeliveryStatusParams) error
	PruneOldDeliveries(context.Context, types.Timestamp) error
}

// MediaFolderRepository manages media folder hierarchy for organizing media assets.
type MediaFolderRepository interface {
	CountMediaFolders() (*int64, error)
	CreateMediaFolder(context.Context, audited.AuditContext, CreateMediaFolderParams) (*MediaFolder, error)
	CreateMediaFolderTable() error
	DeleteMediaFolder(context.Context, audited.AuditContext, types.MediaFolderID) error
	GetMediaFolder(types.MediaFolderID) (*MediaFolder, error)
	GetMediaFolderBreadcrumb(types.MediaFolderID) ([]MediaFolder, error)
	GetMediaFolderByNameAndParent(types.MediaFolderID, string) (*MediaFolder, error)
	GetMediaFolderByNameAtRoot(string) (*MediaFolder, error)
	ListMediaFolders() (*[]MediaFolder, error)
	ListMediaFoldersByParent(types.MediaFolderID) (*[]MediaFolder, error)
	ListMediaFoldersAtRoot() (*[]MediaFolder, error)
	ListMediaFoldersPaginated(PaginationParams) (*[]MediaFolder, error)
	UpdateMediaFolder(context.Context, audited.AuditContext, UpdateMediaFolderParams) (*string, error)
	ValidateMediaFolderName(string, types.NullableMediaFolderID) error
	ValidateMediaFolderMove(types.MediaFolderID, types.NullableMediaFolderID) error
}

// ValidationRepository manages reusable validation configs referenced by fields.
type ValidationRepository interface {
	// Validations
	CountValidations() (*int64, error)
	CreateValidation(context.Context, audited.AuditContext, CreateValidationParams) (*Validation, error)
	CreateValidationTable() error
	DeleteValidation(context.Context, audited.AuditContext, types.ValidationID) error
	GetValidation(types.ValidationID) (*Validation, error)
	ListValidations() (*[]Validation, error)
	ListValidationsPaginated(PaginationParams) (*[]Validation, error)
	UpdateValidation(context.Context, audited.AuditContext, UpdateValidationParams) (*string, error)
	ListValidationsByName(string) (*[]Validation, error)

	// AdminValidations
	CountAdminValidations() (*int64, error)
	CreateAdminValidation(context.Context, audited.AuditContext, CreateAdminValidationParams) (*AdminValidation, error)
	CreateAdminValidationTable() error
	DeleteAdminValidation(context.Context, audited.AuditContext, types.AdminValidationID) error
	GetAdminValidation(types.AdminValidationID) (*AdminValidation, error)
	ListAdminValidations() (*[]AdminValidation, error)
	ListAdminValidationsPaginated(PaginationParams) (*[]AdminValidation, error)
	UpdateAdminValidation(context.Context, audited.AuditContext, UpdateAdminValidationParams) (*string, error)
	ListAdminValidationsByName(string) (*[]AdminValidation, error)
}

// FieldPluginConfigRepository manages the field_plugin_config extension table
// that binds plugin field types to their plugin and interface definitions.
type FieldPluginConfigRepository interface {
	CreateFieldPluginConfigTable() error
	CreateAdminFieldPluginConfigTable() error
	GetFieldPluginConfig(context.Context, types.FieldID) (*FieldPluginConfig, error)
	GetAdminFieldPluginConfig(context.Context, types.FieldID) (*FieldPluginConfig, error)
	CreateFieldPluginConfig(context.Context, CreateFieldPluginConfigParams) error
	CreateAdminFieldPluginConfig(context.Context, CreateFieldPluginConfigParams) error
	UpdateFieldPluginConfig(context.Context, UpdateFieldPluginConfigParams) error
	UpdateAdminFieldPluginConfig(context.Context, UpdateFieldPluginConfigParams) error
	DeleteFieldPluginConfig(context.Context, types.FieldID) error
	DeleteAdminFieldPluginConfig(context.Context, types.FieldID) error
}
