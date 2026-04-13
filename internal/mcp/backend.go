package mcp

import (
	"context"
	"encoding/json"
	"io"
)

// Backend interfaces abstract MCP tool operations from their underlying
// implementation. Each domain gets an interface that both the Go SDK adapter
// (remote mode) and the service adapter (direct mode) implement.
//
// Design: methods accept and return json.RawMessage for complex objects,
// avoiding import of either SDK or db types. Simple parameters (IDs, strings,
// numbers, booleans) use native Go types.

// ContentBackend abstracts content operations for MCP tools.
type ContentBackend interface {
	ListContent(ctx context.Context, limit, offset int64) (json.RawMessage, error)
	GetContent(ctx context.Context, id string) (json.RawMessage, error)
	CreateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteContent(ctx context.Context, id string) error
	GetPage(ctx context.Context, slug, format, locale string) (json.RawMessage, error)
	GetContentTree(ctx context.Context, slug, format string) (json.RawMessage, error)
	ListContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error)
	GetContentField(ctx context.Context, id string) (json.RawMessage, error)
	CreateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteContentField(ctx context.Context, id string) error
	ReorderContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	MoveContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	SaveContentTree(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	HealContent(ctx context.Context, dryRun bool) (json.RawMessage, error)
	BatchUpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	QueryContent(ctx context.Context, datatype string, params json.RawMessage) (json.RawMessage, error)
	GetGlobals(ctx context.Context, format string) (json.RawMessage, error)
	GetContentFull(ctx context.Context, id string) (json.RawMessage, error)
	GetContentByRoute(ctx context.Context, routeID string) (json.RawMessage, error)
	CreateContentComposite(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
}

// AdminContentBackend abstracts admin content operations for MCP tools.
type AdminContentBackend interface {
	ListAdminContent(ctx context.Context, limit, offset int64) (json.RawMessage, error)
	GetAdminContent(ctx context.Context, id string) (json.RawMessage, error)
	CreateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminContent(ctx context.Context, id string) error
	ReorderAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	MoveAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	ListAdminContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error)
	GetAdminContentField(ctx context.Context, id string) (json.RawMessage, error)
	CreateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminContentField(ctx context.Context, id string) error
	AdminGetContentFull(ctx context.Context, limit, offset int64) (json.RawMessage, error)
	GetAdminTree(ctx context.Context, slug string) (json.RawMessage, error)
}

// SchemaBackend abstracts schema operations (datatypes, fields, field types).
type SchemaBackend interface {
	ListDatatypes(ctx context.Context, full bool) (json.RawMessage, error)
	GetDatatype(ctx context.Context, id string) (json.RawMessage, error)
	CreateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteDatatype(ctx context.Context, id string) error
	ListFields(ctx context.Context) (json.RawMessage, error)
	GetField(ctx context.Context, id string) (json.RawMessage, error)
	CreateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteField(ctx context.Context, id string) error
	GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error)
	ListFieldTypes(ctx context.Context) (json.RawMessage, error)
	GetFieldType(ctx context.Context, id string) (json.RawMessage, error)
	CreateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteFieldType(ctx context.Context, id string) error
	GetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error)
	UpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error
	GetFieldMaxSortOrder(ctx context.Context) (json.RawMessage, error)
	UpdateFieldSortOrder(ctx context.Context, id string, sortOrder int64) error
}

// AdminSchemaBackend abstracts admin schema operations.
type AdminSchemaBackend interface {
	ListAdminDatatypes(ctx context.Context, full bool) (json.RawMessage, error)
	GetAdminDatatype(ctx context.Context, id string) (json.RawMessage, error)
	CreateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminDatatype(ctx context.Context, id string) error
	ListAdminFields(ctx context.Context) (json.RawMessage, error)
	GetAdminField(ctx context.Context, id string) (json.RawMessage, error)
	CreateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminField(ctx context.Context, id string) error
	AdminGetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error)
	AdminUpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error
}

// MediaBackend abstracts media operations.
type MediaBackend interface {
	ListMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error)
	GetMedia(ctx context.Context, id string) (json.RawMessage, error)
	UpdateMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteMedia(ctx context.Context, id string) error
	UploadMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error)
	MediaHealth(ctx context.Context) (json.RawMessage, error)
	MediaCleanup(ctx context.Context) (json.RawMessage, error)
	MediaCleanupCheck(ctx context.Context) (json.RawMessage, error)
	ListMediaDimensions(ctx context.Context) (json.RawMessage, error)
	GetMediaDimension(ctx context.Context, id string) (json.RawMessage, error)
	CreateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteMediaDimension(ctx context.Context, id string) error
	DownloadMedia(ctx context.Context, id string) (json.RawMessage, error)
	GetMediaFull(ctx context.Context) (json.RawMessage, error)
	GetMediaReferences(ctx context.Context, id string) (json.RawMessage, error)
	ReprocessMedia(ctx context.Context) (json.RawMessage, error)
}

// MediaFolderBackend abstracts media folder operations.
type MediaFolderBackend interface {
	ListMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error)
	GetMediaFolder(ctx context.Context, id string) (json.RawMessage, error)
	CreateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteMediaFolder(ctx context.Context, id string) (json.RawMessage, error)
	MoveMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	GetMediaFolderTree(ctx context.Context) (json.RawMessage, error)
	ListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error)
}

// RouteBackend abstracts public route operations.
type RouteBackend interface {
	ListRoutes(ctx context.Context) (json.RawMessage, error)
	GetRoute(ctx context.Context, id string) (json.RawMessage, error)
	CreateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteRoute(ctx context.Context, id string) error
	ListRoutesFull(ctx context.Context) (json.RawMessage, error)
}

// AdminRouteBackend abstracts admin route and admin field type operations.
type AdminRouteBackend interface {
	ListAdminRoutes(ctx context.Context) (json.RawMessage, error)
	GetAdminRoute(ctx context.Context, slug string) (json.RawMessage, error)
	CreateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminRoute(ctx context.Context, id string) error
	ListAdminFieldTypes(ctx context.Context) (json.RawMessage, error)
	GetAdminFieldType(ctx context.Context, id string) (json.RawMessage, error)
	CreateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminFieldType(ctx context.Context, id string) error
}

// UserBackend abstracts user operations.
type UserBackend interface {
	Whoami(ctx context.Context) (json.RawMessage, error)
	ListUsers(ctx context.Context) (json.RawMessage, error)
	GetUser(ctx context.Context, id string) (json.RawMessage, error)
	CreateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsersFull(ctx context.Context) (json.RawMessage, error)
	GetUserFull(ctx context.Context, id string) (json.RawMessage, error)
	ReassignAndDeleteUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	ListUserSessions(ctx context.Context) (json.RawMessage, error)
}

// RBACBackend abstracts role-based access control operations.
type RBACBackend interface {
	ListRoles(ctx context.Context) (json.RawMessage, error)
	GetRole(ctx context.Context, id string) (json.RawMessage, error)
	CreateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteRole(ctx context.Context, id string) error
	ListPermissions(ctx context.Context) (json.RawMessage, error)
	GetPermission(ctx context.Context, id string) (json.RawMessage, error)
	CreatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeletePermission(ctx context.Context, id string) error
	AssignRolePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	RemoveRolePermission(ctx context.Context, id string) error
	ListRolePermissions(ctx context.Context) (json.RawMessage, error)
	GetRolePermission(ctx context.Context, id string) (json.RawMessage, error)
	ListRolePermissionsByRole(ctx context.Context, roleID string) (json.RawMessage, error)
}

// SessionBackend abstracts session operations.
type SessionBackend interface {
	ListSessions(ctx context.Context) (json.RawMessage, error)
	GetSession(ctx context.Context, id string) (json.RawMessage, error)
	UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteSession(ctx context.Context, id string) error
}

// TokenBackend abstracts token operations.
type TokenBackend interface {
	ListTokens(ctx context.Context) (json.RawMessage, error)
	GetToken(ctx context.Context, id string) (json.RawMessage, error)
	CreateToken(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteToken(ctx context.Context, id string) error
}

// SSHKeyBackend abstracts SSH key operations.
type SSHKeyBackend interface {
	ListSSHKeys(ctx context.Context) (json.RawMessage, error)
	CreateSSHKey(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteSSHKey(ctx context.Context, id string) error
}

// OAuthBackend abstracts user OAuth connection operations.
type OAuthBackend interface {
	ListUsersOAuth(ctx context.Context) (json.RawMessage, error)
	GetUserOAuth(ctx context.Context, id string) (json.RawMessage, error)
	CreateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteUserOAuth(ctx context.Context, id string) error
}

// TableBackend abstracts CMS metadata table operations.
type TableBackend interface {
	ListTables(ctx context.Context) (json.RawMessage, error)
	GetTable(ctx context.Context, id string) (json.RawMessage, error)
	CreateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteTable(ctx context.Context, id string) error
}

// PluginBackend abstracts plugin management operations.
type PluginBackend interface {
	ListPlugins(ctx context.Context) (json.RawMessage, error)
	GetPlugin(ctx context.Context, name string) (json.RawMessage, error)
	ReloadPlugin(ctx context.Context, name string) (json.RawMessage, error)
	EnablePlugin(ctx context.Context, name string) (json.RawMessage, error)
	DisablePlugin(ctx context.Context, name string) (json.RawMessage, error)
	PluginCleanupDryRun(ctx context.Context) (json.RawMessage, error)
	PluginCleanupDrop(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	ListPluginRoutes(ctx context.Context) (json.RawMessage, error)
	ApprovePluginRoutes(ctx context.Context, params json.RawMessage) error
	RevokePluginRoutes(ctx context.Context, params json.RawMessage) error
	ListPluginHooks(ctx context.Context) (json.RawMessage, error)
	ApprovePluginHooks(ctx context.Context, params json.RawMessage) error
	RevokePluginHooks(ctx context.Context, params json.RawMessage) error
}

// ConfigBackend abstracts server configuration operations.
type ConfigBackend interface {
	GetConfig(ctx context.Context, category string) (json.RawMessage, error)
	GetConfigMeta(ctx context.Context) (json.RawMessage, error)
	UpdateConfig(ctx context.Context, updates map[string]any) (json.RawMessage, error)
}

// ImportBackend abstracts content import operations.
type ImportBackend interface {
	ImportContent(ctx context.Context, format string, data any) (json.RawMessage, error)
	ImportBulk(ctx context.Context, format string, data any) (json.RawMessage, error)
}

// DeployBackend abstracts deploy sync operations.
type DeployBackend interface {
	DeployHealth(ctx context.Context) (json.RawMessage, error)
	DeployExport(ctx context.Context, tables []string) (json.RawMessage, error)
	DeployImport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)
	DeployDryRun(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)
}

// AdminMediaBackend abstracts admin media operations.
type AdminMediaBackend interface {
	ListAdminMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error)
	GetAdminMedia(ctx context.Context, id string) (json.RawMessage, error)
	UpdateAdminMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminMedia(ctx context.Context, id string) error
	UploadAdminMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error)
	// ListMediaDimensions returns shared media dimension presets (same as public).
	ListMediaDimensions(ctx context.Context) (json.RawMessage, error)
}

// AdminMediaFolderBackend abstracts admin media folder operations.
type AdminMediaFolderBackend interface {
	ListAdminMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error)
	GetAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error)
	CreateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error)
	MoveAdminMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminGetMediaFolderTree(ctx context.Context) (json.RawMessage, error)
	AdminListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error)
}

// HealthBackend abstracts server health checks.
type HealthBackend interface {
	Health(ctx context.Context) (json.RawMessage, error)
	GetMetrics(ctx context.Context) (json.RawMessage, error)
	GetEnvironment(ctx context.Context) (json.RawMessage, error)
}

// ActivityBackend abstracts activity feed operations.
type ActivityBackend interface {
	ListRecentActivity(ctx context.Context, limit int64) (json.RawMessage, error)
}

// AuthBackend abstracts non-interactive auth operations.
type AuthBackend interface {
	RegisterUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	RequestPasswordReset(ctx context.Context, email string) (json.RawMessage, error)
}

// PublishingBackend abstracts content publishing operations.
type PublishingBackend interface {
	PublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	ScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminPublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminUnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
}

// VersionBackend abstracts content version operations.
type VersionBackend interface {
	ListVersions(ctx context.Context, contentID string) (json.RawMessage, error)
	GetVersion(ctx context.Context, versionID string) (json.RawMessage, error)
	CreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteVersion(ctx context.Context, versionID string) error
	RestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminListVersions(ctx context.Context, contentID string) (json.RawMessage, error)
	AdminGetVersion(ctx context.Context, versionID string) (json.RawMessage, error)
	AdminCreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminDeleteVersion(ctx context.Context, versionID string) error
	AdminRestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
}

// WebhookBackend abstracts webhook management operations.
type WebhookBackend interface {
	ListWebhooks(ctx context.Context) (json.RawMessage, error)
	GetWebhook(ctx context.Context, id string) (json.RawMessage, error)
	CreateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteWebhook(ctx context.Context, id string) error
	TestWebhook(ctx context.Context, id string) (json.RawMessage, error)
	ListWebhookDeliveries(ctx context.Context, webhookID string) (json.RawMessage, error)
	RetryWebhookDelivery(ctx context.Context, deliveryID string) error
}

// LocaleBackend abstracts locale management operations.
type LocaleBackend interface {
	ListLocales(ctx context.Context) (json.RawMessage, error)
	ListAdminLocales(ctx context.Context) (json.RawMessage, error)
	GetLocale(ctx context.Context, id string) (json.RawMessage, error)
	CreateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteLocale(ctx context.Context, id string) error
	CreateTranslation(ctx context.Context, contentDataID string, params json.RawMessage) (json.RawMessage, error)
	AdminCreateTranslation(ctx context.Context, adminContentDataID string, params json.RawMessage) (json.RawMessage, error)
}

// ValidationBackend abstracts validation rule operations.
type ValidationBackend interface {
	ListValidations(ctx context.Context) (json.RawMessage, error)
	GetValidation(ctx context.Context, id string) (json.RawMessage, error)
	CreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	UpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	DeleteValidation(ctx context.Context, id string) error
	SearchValidations(ctx context.Context, query string) (json.RawMessage, error)
	AdminListValidations(ctx context.Context) (json.RawMessage, error)
	AdminGetValidation(ctx context.Context, id string) (json.RawMessage, error)
	AdminCreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminUpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
	AdminDeleteValidation(ctx context.Context, id string) error
	AdminSearchValidations(ctx context.Context, query string) (json.RawMessage, error)
}

// SearchBackend abstracts content search operations.
type SearchBackend interface {
	SearchContent(ctx context.Context, query string, limit, offset int64) (json.RawMessage, error)
	RebuildSearchIndex(ctx context.Context) (json.RawMessage, error)
}

// Backends holds all domain backends for MCP tool registration.
// Each field can be satisfied by either an SDK adapter (remote mode)
// or a service adapter (direct mode).
type Backends struct {
	Content           ContentBackend
	AdminContent      AdminContentBackend
	Schema            SchemaBackend
	AdminSchema       AdminSchemaBackend
	Media             MediaBackend
	MediaFolders      MediaFolderBackend
	AdminMedia        AdminMediaBackend
	AdminMediaFolders AdminMediaFolderBackend
	Routes            RouteBackend
	AdminRoutes       AdminRouteBackend
	Users             UserBackend
	RBAC              RBACBackend
	Sessions          SessionBackend
	Tokens            TokenBackend
	SSHKeys           SSHKeyBackend
	OAuth             OAuthBackend
	Tables            TableBackend
	Plugins           PluginBackend
	Config            ConfigBackend
	Import            ImportBackend
	Deploy            DeployBackend
	Health            HealthBackend
	Publishing        PublishingBackend
	Versions          VersionBackend
	Webhooks          WebhookBackend
	Locales           LocaleBackend
	Validations       ValidationBackend
	Search            SearchBackend
	Activity          ActivityBackend
	Auth              AuthBackend
}
