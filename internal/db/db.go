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
	"github.com/hegner123/modulacms/internal/db/types"
	utility "github.com/hegner123/modulacms/internal/utility"
)

//go:embed sql
var sqlFiles embed.FS

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

const (
	Open   DbStatus = "open"
	Closed DbStatus = "closed"
	Err    DbStatus = "error"
)

// DbDriver is the interface for all database drivers
type DbDriver interface {
	// Connection
	CreateAllTables() error
	CreateBootstrapData() error
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
	CreateAdminContentData(CreateAdminContentDataParams) AdminContentData
	CreateAdminContentDataTable() error
	DeleteAdminContentData(types.AdminContentID) error
	GetAdminContentData(types.AdminContentID) (*AdminContentData, error)
	ListAdminContentData() (*[]AdminContentData, error)
	ListAdminContentDataByRoute(string) (*[]AdminContentData, error)
	UpdateAdminContentData(UpdateAdminContentDataParams) (*string, error)

	// AdminContentFields
	CountAdminContentFields() (*int64, error)
	CreateAdminContentField(CreateAdminContentFieldParams) AdminContentFields
	CreateAdminContentFieldTable() error
	DeleteAdminContentField(types.AdminContentFieldID) error
	GetAdminContentField(types.AdminContentFieldID) (*AdminContentFields, error)
	ListAdminContentFields() (*[]AdminContentFields, error)
	ListAdminContentFieldsByRoute(string) (*[]AdminContentFields, error)
	UpdateAdminContentField(UpdateAdminContentFieldParams) (*string, error)

	// AdminDatatypes
	CountAdminDatatypes() (*int64, error)
	CreateAdminDatatype(CreateAdminDatatypeParams) AdminDatatypes
	CreateAdminDatatypeTable() error
	DeleteAdminDatatype(types.AdminDatatypeID) error
	GetAdminDatatypeById(types.AdminDatatypeID) (*AdminDatatypes, error)
	ListAdminDatatypes() (*[]AdminDatatypes, error)
	UpdateAdminDatatype(UpdateAdminDatatypeParams) (*string, error)

	// AdminDatatypeFields
	CountAdminDatatypeFields() (*int64, error)
	CreateAdminDatatypeField(CreateAdminDatatypeFieldParams) AdminDatatypeFields
	CreateAdminDatatypeFieldTable() error
	DeleteAdminDatatypeField(string) error
	ListAdminDatatypeField() (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByDatatypeID(types.NullableAdminDatatypeID) (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByFieldID(types.NullableAdminFieldID) (*[]AdminDatatypeFields, error)
	UpdateAdminDatatypeField(UpdateAdminDatatypeFieldParams) (*string, error)

	// AdminFields
	CountAdminFields() (*int64, error)
	CreateAdminField(CreateAdminFieldParams) AdminFields
	CreateAdminFieldTable() error
	DeleteAdminField(types.AdminFieldID) error
	GetAdminField(types.AdminFieldID) (*AdminFields, error)
	ListAdminFields() (*[]AdminFields, error)
	UpdateAdminField(UpdateAdminFieldParams) (*string, error)

	// AdminRoutes
	CountAdminRoutes() (*int64, error)
	CreateAdminRoute(CreateAdminRouteParams) AdminRoutes
	CreateAdminRouteTable() error
	DeleteAdminRoute(types.AdminRouteID) error
	GetAdminRoute(types.Slug) (*AdminRoutes, error)
	ListAdminRoutes() (*[]AdminRoutes, error)
	UpdateAdminRoute(UpdateAdminRouteParams) (*string, error)

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
	CreateContentData(CreateContentDataParams) ContentData
	CreateContentDataTable() error
	DeleteContentData(types.ContentID) error
	GetContentData(types.ContentID) (*ContentData, error)
	ListContentData() (*[]ContentData, error)
	ListContentDataByRoute(types.NullableRouteID) (*[]ContentData, error)
	UpdateContentData(UpdateContentDataParams) (*string, error)

	// ContentFields
	CountContentFields() (*int64, error)
	CreateContentField(CreateContentFieldParams) ContentFields
	CreateContentFieldTable() error
	DeleteContentField(types.ContentFieldID) error
	GetContentField(types.ContentFieldID) (*ContentFields, error)
	GetContentFieldsByRoute(types.NullableRouteID) (*[]GetContentFieldsByRouteRow, error)
	ListContentFields() (*[]ContentFields, error)
	ListContentFieldsByRoute(types.NullableRouteID) (*[]ContentFields, error)
	UpdateContentField(UpdateContentFieldParams) (*string, error)

	// Datatypes
	CountDatatypes() (*int64, error)
	CreateDatatype(CreateDatatypeParams) Datatypes
	CreateDatatypeTable() error
	DeleteDatatype(types.DatatypeID) error
	GetDatatype(types.DatatypeID) (*Datatypes, error)
	ListDatatypes() (*[]Datatypes, error)
	ListDatatypesRoot() (*[]Datatypes, error)
	UpdateDatatype(UpdateDatatypeParams) (*string, error)

	// DatatypeFields
	CountDatatypeFields() (*int64, error)
	CreateDatatypeField(CreateDatatypeFieldParams) DatatypeFields
	CreateDatatypeFieldTable() error
	DeleteDatatypeField(string) error
	ListDatatypeField() (*[]DatatypeFields, error)
	ListDatatypeFieldByDatatypeID(types.NullableDatatypeID) (*[]DatatypeFields, error)
	ListDatatypeFieldByFieldID(types.NullableFieldID) (*[]DatatypeFields, error)
	UpdateDatatypeField(UpdateDatatypeFieldParams) (*string, error)

	// Fields
	CountFields() (*int64, error)
	CreateField(CreateFieldParams) Fields
	CreateFieldTable() error
	DeleteField(types.FieldID) error
	GetField(types.FieldID) (*Fields, error)
	GetFieldDefinitionsByRoute(types.NullableRouteID) (*[]GetFieldDefinitionsByRouteRow, error)
	ListFields() (*[]Fields, error)
	ListFieldsByDatatypeID(types.NullableContentID) (*[]Fields, error)
	UpdateField(UpdateFieldParams) (*string, error)

	// Media
	CountMedia() (*int64, error)
	CreateMedia(CreateMediaParams) Media
	CreateMediaTable() error
	DeleteMedia(types.MediaID) error
	GetMedia(types.MediaID) (*Media, error)
	GetMediaByName(string) (*Media, error)
	GetMediaByURL(types.URL) (*Media, error)
	ListMedia() (*[]Media, error)
	UpdateMedia(UpdateMediaParams) (*string, error)

	// MediaDimensions
	CountMediaDimensions() (*int64, error)
	CreateMediaDimension(CreateMediaDimensionParams) MediaDimensions
	CreateMediaDimensionTable() error
	DeleteMediaDimension(string) error
	GetMediaDimension(string) (*MediaDimensions, error)
	ListMediaDimensions() (*[]MediaDimensions, error)
	UpdateMediaDimension(UpdateMediaDimensionParams) (*string, error)

	// Permissions
	CountPermissions() (*int64, error)
	CreatePermission(CreatePermissionParams) Permissions
	CreatePermissionTable() error
	DeletePermission(types.PermissionID) error
	GetPermission(types.PermissionID) (*Permissions, error)
	ListPermissions() (*[]Permissions, error)
	UpdatePermission(UpdatePermissionParams) (*string, error)

	// Roles
	CountRoles() (*int64, error)
	CreateRole(CreateRoleParams) Roles
	CreateRoleTable() error
	DeleteRole(types.RoleID) error
	GetRole(types.RoleID) (*Roles, error)
	ListRoles() (*[]Roles, error)
	UpdateRole(UpdateRoleParams) (*string, error)

	// Routes
	CountRoutes() (*int64, error)
	CreateRoute(CreateRouteParams) Routes
	CreateRouteTable() error
	DeleteRoute(types.RouteID) error
	GetContentTreeByRoute(types.NullableRouteID) (*[]GetContentTreeByRouteRow, error)
	GetRoute(types.RouteID) (*Routes, error)
	GetRouteID(string) (*types.RouteID, error)
	GetRouteTreeByRouteID(types.NullableRouteID) (*[]GetRouteTreeByRouteIDRow, error)
	ListRoutes() (*[]Routes, error)
	UpdateRoute(UpdateRouteParams) (*string, error)

	// Sessions
	CountSessions() (*int64, error)
	CreateSession(CreateSessionParams) (*Sessions, error)
	CreateSessionTable() error
	DeleteSession(types.SessionID) error
	GetSession(types.SessionID) (*Sessions, error)
	GetSessionByUserId(types.NullableUserID) (*Sessions, error)
	ListSessions() (*[]Sessions, error)
	UpdateSession(UpdateSessionParams) (*string, error)

	// Tables
	CountTables() (*int64, error)
	CreateTable(CreateTableParams) Tables
	CreateTableTable() error
	DeleteTable(string) error
	GetTable(string) (*Tables, error)
	ListTables() (*[]Tables, error)
	UpdateTable(UpdateTableParams) (*string, error)

	// Tokens
	CountTokens() (*int64, error)
	CreateToken(CreateTokenParams) Tokens
	CreateTokenTable() error
	DeleteToken(string) error
	GetToken(string) (*Tokens, error)
	GetTokenByTokenValue(string) (*Tokens, error)
	GetTokenByUserId(types.NullableUserID) (*[]Tokens, error)
	ListTokens() (*[]Tokens, error)
	UpdateToken(UpdateTokenParams) (*string, error)

	// Users
	CountUsers() (*int64, error)
	CreateUser(CreateUserParams) (*Users, error)
	CreateUserTable() error
	DeleteUser(types.UserID) error
	GetUser(types.UserID) (*Users, error)
	GetUserByEmail(types.Email) (*Users, error)
	GetUserBySSHFingerprint(string) (*Users, error)
	ListUsers() (*[]Users, error)
	UpdateUser(UpdateUserParams) (*string, error)

	// UserOauths
	CountUserOauths() (*int64, error)
	CreateUserOauth(CreateUserOauthParams) (*UserOauth, error)
	CreateUserOauthTable() error
	DeleteUserOauth(types.UserOauthID) error
	GetUserOauth(types.UserOauthID) (*UserOauth, error)
	GetUserOauthByProviderID(string, string) (*UserOauth, error)
	GetUserOauthByUserId(types.NullableUserID) (*UserOauth, error)
	ListUserOauths() (*[]UserOauth, error)
	UpdateUserOauth(UpdateUserOauthParams) (*string, error)

	// UserSshKeys
	CountUserSshKeys() (*int64, error)
	CreateUserSshKey(CreateUserSshKeyParams) (*UserSshKeys, error)
	CreateUserSshKeyTable() error
	DeleteUserSshKey(string) error
	GetUserSshKey(string) (*UserSshKeys, error)
	GetUserSshKeyByFingerprint(string) (*UserSshKeys, error)
	ListUserSshKeys(types.NullableUserID) (*[]UserSshKeys, error)
	UpdateUserSshKeyLabel(string, string) error
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
	return d.Connection.Ping()
}

// Ping checks if the MySQL database connection is still alive
func (d MysqlDatabase) Ping() error {
	return d.Connection.Ping()
}

// Ping checks if the PostgreSQL database connection is still alive
func (d PsqlDatabase) Ping() error {
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

	// Tier 0: Foundation tables (no dependencies)
	err := d.CreatePermissionTable()
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

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeFieldTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 22 tables to verify successful table creation.
// If any table failed to create, this function will catch it immediately during install rather than later during operation.
func (d Database) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)
	permission := d.CreatePermission(CreatePermissionParams{
		TableID: "",
		Mode:    7,
		Label:   "system_admin",
	})
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 2. Create system admin role (role_id = 1)
	adminRole := d.CreateRole(CreateRoleParams{
		Label:       "system_admin",
		Permissions: `{"system_admin": true}`,
	})
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole := d.CreateRole(CreateRoleParams{
		Label:       "viewer",
		Permissions: `{"read": true}`,
	})
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modulacms.local"),
		Hash:         "",
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
	homeRoute := d.CreateRoute(CreateRouteParams{
		Slug:         types.Slug("home"),
		Title:        "Home",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype := d.CreateDatatype(CreateDatatypeParams{
		ParentID:     types.NullableContentID{},
		Label:        "Page",
		Type:         "page",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute := d.CreateAdminRoute(CreateAdminRouteParams{
		Slug:         types.Slug("admin"),
		Title:        "Admin",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype := d.CreateAdminDatatype(CreateAdminDatatypeParams{
		ParentID:     types.NullableContentID{},
		Label:        "Admin Page",
		Type:         "admin_page",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1)
	adminField := d.CreateAdminField(CreateAdminFieldParams{
		ParentID:     types.NullableContentID{},
		Label:        "Content",
		Data:         "",
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1)
	field := d.CreateField(CreateFieldParams{
		ParentID:     types.NullableContentID{},
		Label:        "Content",
		Data:         "",
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData := d.CreateContentData(CreateContentDataParams{
		RouteID:       types.NullableRouteID{Valid: true, ID: homeRoute.RouteID},
		ParentID:      types.NullableContentID{},
		FirstChildID:  sql.NullString{},
		NextSiblingID: sql.NullString{},
		PrevSiblingID: sql.NullString{},
		DatatypeID:    types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData := d.CreateAdminContentData(CreateAdminContentDataParams{
		ParentID:        types.NullableContentID{},
		FirstChildID:    sql.NullString{},
		NextSiblingID:   sql.NullString{},
		PrevSiblingID:   sql.NullString{},
		AdminRouteID:    adminRoute.AdminRouteID.String(),
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AuthorID:        types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:     types.TimestampNow(),
		DateModified:    types.TimestampNow(),
	})
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField := d.CreateContentField(CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField := d.CreateAdminContentField(CreateAdminContentFieldParams{
		AdminRouteID:       sql.NullString{},
		AdminContentDataID: adminContentData.AdminContentDataID.String(),
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:        types.TimestampNow(),
		DateModified:       types.TimestampNow(),
	})
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension := d.CreateMediaDimension(CreateMediaDimensionParams{
		Label:       StringToNullString("Default"),
		Width:       Int64ToNullInt64(1920),
		Height:      Int64ToNullInt64(1080),
		AspectRatio: StringToNullString("16:9"),
	})
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media := d.CreateMedia(CreateMediaParams{
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
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token := d.CreateToken(CreateTokenParams{
		UserID:    types.NullableUserID{Valid: true, ID: systemUser.UserID},
		TokenType: "validation",
		Token:     "bootstrap_validation_token",
		IssuedAt:  utility.TimestampS(),
		ExpiresAt: types.TimestampNow(),
		Revoked:   true,
	})
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(CreateSessionParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		CreatedAt:   types.TimestampNow(),
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
	userOauth, err := d.CreateUserOauth(CreateUserOauthParams{
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
	userSshKey, err := d.CreateUserSshKey(CreateUserSshKeyParams{
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

	// 20. Register all 22 ModulaCMS tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
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
		table := d.CreateTable(CreateTableParams{Label: tableName})
		if table.ID == "" {
			return fmt.Errorf("failed to register table in tables registry: %s", tableName)
		}
	}

	// 21. Create default datatypes_fields junction record (id = 1) - Links datatype to field
	datatypeField := d.CreateDatatypeField(CreateDatatypeFieldParams{
		DatatypeID: types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		FieldID:    types.NullableFieldID{Valid: true, ID: field.FieldID},
	})
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record (id = 1) - Links admin datatype to admin field
	adminDatatypeField := d.CreateAdminDatatypeField(CreateAdminDatatypeFieldParams{
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AdminFieldID:    types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
	})
	if adminDatatypeField.ID == "" {
		return fmt.Errorf("failed to create default admin_datatypes_fields")
	}

	utility.DefaultLogger.Finfo("Bootstrap data created successfully: ALL 22 tables validated with bootstrap records + complete table registry populated")
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

	// Validate roles table (should have at least 2 records: system_admin + viewer)
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

	// Validate tables table (should have EXACTLY 22 records - all core tables)
	tableCount, err := d.CountTables()
	if err != nil || tableCount == nil || *tableCount != 22 {
		errors = append(errors, fmt.Sprintf("tables table: expected exactly 22 records (table registry), got %v", tableCount))
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

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 22 tables contain expected records")
	return nil
}

// CreateAllTables creates all MySQL database tables
func (d MysqlDatabase) CreateAllTables() error {
	// CRITICAL: Tables must be created in dependency order to satisfy FK constraints
	// See: ai/reference/TABLE_CREATION_ORDER.md for complete dependency graph

	// Tier 0: Foundation tables (no dependencies)
	err := d.CreatePermissionTable()
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

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeFieldTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 22 tables to verify successful table creation.
// If any table failed to create, this function will catch it immediately during install rather than later during operation.
func (d MysqlDatabase) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)
	permission := d.CreatePermission(CreatePermissionParams{
		TableID: "",
		Mode:    7,
		Label:   "system_admin",
	})
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 2. Create system admin role (role_id = 1)
	adminRole := d.CreateRole(CreateRoleParams{
		Label:       "system_admin",
		Permissions: `{"system_admin": true}`,
	})
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole := d.CreateRole(CreateRoleParams{
		Label:       "viewer",
		Permissions: `{"read": true}`,
	})
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modulacms.local"),
		Hash:         "",
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
	homeRoute := d.CreateRoute(CreateRouteParams{
		Slug:         types.Slug("home"),
		Title:        "Home",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype := d.CreateDatatype(CreateDatatypeParams{
		ParentID:     types.NullableContentID{},
		Label:        "Page",
		Type:         "page",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute := d.CreateAdminRoute(CreateAdminRouteParams{
		Slug:         types.Slug("admin"),
		Title:        "Admin",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype := d.CreateAdminDatatype(CreateAdminDatatypeParams{
		ParentID:     types.NullableContentID{},
		Label:        "Admin Page",
		Type:         "admin_page",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1)
	adminField := d.CreateAdminField(CreateAdminFieldParams{
		ParentID:     types.NullableContentID{},
		Label:        "Content",
		Data:         "",
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1)
	field := d.CreateField(CreateFieldParams{
		ParentID:     types.NullableContentID{},
		Label:        "Content",
		Data:         "",
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData := d.CreateContentData(CreateContentDataParams{
		RouteID:       types.NullableRouteID{Valid: true, ID: homeRoute.RouteID},
		ParentID:      types.NullableContentID{},
		FirstChildID:  sql.NullString{},
		NextSiblingID: sql.NullString{},
		PrevSiblingID: sql.NullString{},
		DatatypeID:    types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData := d.CreateAdminContentData(CreateAdminContentDataParams{
		ParentID:        types.NullableContentID{},
		FirstChildID:    sql.NullString{},
		NextSiblingID:   sql.NullString{},
		PrevSiblingID:   sql.NullString{},
		AdminRouteID:    adminRoute.AdminRouteID.String(),
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AuthorID:        types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:     types.TimestampNow(),
		DateModified:    types.TimestampNow(),
	})
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField := d.CreateContentField(CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField := d.CreateAdminContentField(CreateAdminContentFieldParams{
		AdminRouteID:       sql.NullString{},
		AdminContentDataID: adminContentData.AdminContentDataID.String(),
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:        types.TimestampNow(),
		DateModified:       types.TimestampNow(),
	})
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension := d.CreateMediaDimension(CreateMediaDimensionParams{
		Label:       StringToNullString("Default"),
		Width:       Int64ToNullInt64(1920),
		Height:      Int64ToNullInt64(1080),
		AspectRatio: StringToNullString("16:9"),
	})
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media := d.CreateMedia(CreateMediaParams{
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
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token := d.CreateToken(CreateTokenParams{
		UserID:    types.NullableUserID{Valid: true, ID: systemUser.UserID},
		TokenType: "validation",
		Token:     "bootstrap_validation_token",
		IssuedAt:  utility.TimestampS(),
		ExpiresAt: types.TimestampNow(),
		Revoked:   true,
	})
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(CreateSessionParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		CreatedAt:   types.TimestampNow(),
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
	userOauth, err := d.CreateUserOauth(CreateUserOauthParams{
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
	userSshKey, err := d.CreateUserSshKey(CreateUserSshKeyParams{
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

	// 20. Register all 22 ModulaCMS tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
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
		table := d.CreateTable(CreateTableParams{Label: tableName})
		if table.ID == "" {
			return fmt.Errorf("failed to register table in tables registry: %s", tableName)
		}
	}

	// 21. Create default datatypes_fields junction record (id = 1) - Links datatype to field
	datatypeField := d.CreateDatatypeField(CreateDatatypeFieldParams{
		DatatypeID: types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		FieldID:    types.NullableFieldID{Valid: true, ID: field.FieldID},
	})
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record - Links admin datatype to admin field
	adminDatatypeField := d.CreateAdminDatatypeField(CreateAdminDatatypeFieldParams{
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AdminFieldID:    types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
	})
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

	// Validate all 21 tables have expected record counts
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
	if err != nil || tableCount == nil || *tableCount != 21 {
		errors = append(errors, fmt.Sprintf("tables table: expected exactly 21 records (table registry), got %v", tableCount))
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

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 22 tables contain expected records")
	return nil
}

// CreateAllTables creates all PostgreSQL database tables
func (d PsqlDatabase) CreateAllTables() error {
	// CRITICAL: Tables must be created in dependency order to satisfy FK constraints
	// See: ai/reference/TABLE_CREATION_ORDER.md for complete dependency graph

	// Tier 0: Foundation tables (no dependencies)
	err := d.CreatePermissionTable()
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

	// Tier 6: Junction tables (depend on both sides)
	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeFieldTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateBootstrapData inserts required system records for initial database setup.
// CRITICAL: Must be called after CreateAllTables() succeeds.
// Inserts validation/bootstrap data for ALL 22 tables to verify successful table creation.
// If any table failed to create, this function will catch it immediately during install rather than later during operation.
func (d PsqlDatabase) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)
	permission := d.CreatePermission(CreatePermissionParams{
		TableID: "",
		Mode:    7,
		Label:   "system_admin",
	})
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 2. Create system admin role (role_id = 1)
	adminRole := d.CreateRole(CreateRoleParams{
		Label:       "system_admin",
		Permissions: `{"system_admin": true}`,
	})
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole := d.CreateRole(CreateRoleParams{
		Label:       "viewer",
		Permissions: `{"read": true}`,
	})
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(CreateUserParams{
		Username:     "system",
		Name:         "System Administrator",
		Email:        types.Email("system@modulacms.local"),
		Hash:         "",
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
	homeRoute := d.CreateRoute(CreateRouteParams{
		Slug:         types.Slug("home"),
		Title:        "Home",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype := d.CreateDatatype(CreateDatatypeParams{
		ParentID:     types.NullableContentID{},
		Label:        "Page",
		Type:         "page",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute := d.CreateAdminRoute(CreateAdminRouteParams{
		Slug:         types.Slug("admin"),
		Title:        "Admin",
		Status:       1,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype := d.CreateAdminDatatype(CreateAdminDatatypeParams{
		ParentID:     types.NullableContentID{},
		Label:        "Admin Page",
		Type:         "admin_page",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1)
	adminField := d.CreateAdminField(CreateAdminFieldParams{
		ParentID:     types.NullableContentID{},
		Label:        "Content",
		Data:         "",
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1)
	field := d.CreateField(CreateFieldParams{
		ParentID:     types.NullableContentID{},
		Label:        "Content",
		Data:         "",
		Type:         "text",
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData := d.CreateContentData(CreateContentDataParams{
		RouteID:       types.NullableRouteID{Valid: true, ID: homeRoute.RouteID},
		ParentID:      types.NullableContentID{},
		FirstChildID:  sql.NullString{},
		NextSiblingID: sql.NullString{},
		PrevSiblingID: sql.NullString{},
		DatatypeID:    types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData := d.CreateAdminContentData(CreateAdminContentDataParams{
		ParentID:        types.NullableContentID{},
		FirstChildID:    sql.NullString{},
		NextSiblingID:   sql.NullString{},
		PrevSiblingID:   sql.NullString{},
		AdminRouteID:    adminRoute.AdminRouteID.String(),
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AuthorID:        types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:     types.TimestampNow(),
		DateModified:    types.TimestampNow(),
	})
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField := d.CreateContentField(CreateContentFieldParams{
		RouteID:       types.NullableRouteID{},
		ContentDataID: types.NullableContentID{Valid: true, ID: contentData.ContentDataID},
		FieldID:       types.NullableFieldID{Valid: true, ID: field.FieldID},
		FieldValue:    "Default content",
		AuthorID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField := d.CreateAdminContentField(CreateAdminContentFieldParams{
		AdminRouteID:       sql.NullString{},
		AdminContentDataID: adminContentData.AdminContentDataID.String(),
		AdminFieldID:       types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
		AdminFieldValue:    "Default admin content",
		AuthorID:           types.NullableUserID{Valid: true, ID: systemUser.UserID},
		DateCreated:        types.TimestampNow(),
		DateModified:       types.TimestampNow(),
	})
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension := d.CreateMediaDimension(CreateMediaDimensionParams{
		Label:       StringToNullString("Default"),
		Width:       Int64ToNullInt64(1920),
		Height:      Int64ToNullInt64(1080),
		AspectRatio: StringToNullString("16:9"),
	})
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media := d.CreateMedia(CreateMediaParams{
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
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token := d.CreateToken(CreateTokenParams{
		UserID:    types.NullableUserID{Valid: true, ID: systemUser.UserID},
		TokenType: "validation",
		Token:     "bootstrap_validation_token",
		IssuedAt:  utility.TimestampS(),
		ExpiresAt: types.TimestampNow(),
		Revoked:   true,
	})
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(CreateSessionParams{
		UserID:      types.NullableUserID{Valid: true, ID: systemUser.UserID},
		CreatedAt:   types.TimestampNow(),
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
	userOauth, err := d.CreateUserOauth(CreateUserOauthParams{
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
	userSshKey, err := d.CreateUserSshKey(CreateUserSshKeyParams{
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

	// 20. Register all 22 ModulaCMS tables in the tables registry
	// This tracks all tables in the system and is critical for plugin support
	tableNames := []string{
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
		table := d.CreateTable(CreateTableParams{Label: tableName})
		if table.ID == "" {
			return fmt.Errorf("failed to register table in tables registry: %s", tableName)
		}
	}

	// 21. Create default datatypes_fields junction record (id = 1) - Links datatype to field
	datatypeField := d.CreateDatatypeField(CreateDatatypeFieldParams{
		DatatypeID: types.NullableDatatypeID{Valid: true, ID: pageDatatype.DatatypeID},
		FieldID:    types.NullableFieldID{Valid: true, ID: field.FieldID},
	})
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record - Links admin datatype to admin field
	adminDatatypeField := d.CreateAdminDatatypeField(CreateAdminDatatypeFieldParams{
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: true, ID: adminDatatype.AdminDatatypeID},
		AdminFieldID:    types.NullableAdminFieldID{Valid: true, ID: adminField.AdminFieldID},
	})
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

	// Validate all 21 tables have expected record counts
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
	if err != nil || tableCount == nil || *tableCount != 21 {
		errors = append(errors, fmt.Sprintf("tables table: expected exactly 21 records (table registry), got %v", tableCount))
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

	utility.DefaultLogger.Finfo("Bootstrap validation passed: all 22 tables contain expected records")
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
	for _, tt := range st {
		d.CreateTable(CreateTableParams{Label: tt})
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
	for _, tt := range st {
		d.CreateTable(CreateTableParams{Label: tt})
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
	for _, tt := range st {
		d.CreateTable(CreateTableParams{Label: tt})
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
