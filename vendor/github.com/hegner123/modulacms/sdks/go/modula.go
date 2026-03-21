// Package modula provides a Go client for the ModulaCMS REST API.
//
// The SDK offers typed access to all CMS resources through a single [Client]
// instance. Most resources use the generic [Resource] type, which provides
// List, Get, Create, Update, Delete, ListPaginated, Count, and RawList methods
// with compile-time type safety.
//
// All entity IDs are branded string types (e.g. [ContentID], [UserID],
// [DatatypeID]) so you cannot accidentally pass a UserID where a ContentID
// is expected.
//
// # Quick start
//
//	client, err := modula.NewClient(modula.ClientConfig{
//	    BaseURL: "https://cms.example.com",
//	    APIKey:  "your-api-key",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List all datatypes
//	datatypes, err := client.Datatypes.List(ctx)
//
//	// Get a single content item by ID
//	item, err := client.ContentData.Get(ctx, contentID)
//
//	// Paginated listing
//	page, err := client.Users.ListPaginated(ctx, modula.PaginationParams{
//	    Limit: 25, Offset: 0,
//	})
//
// # Error handling
//
// All API errors are returned as [*ApiError] and can be classified with
// helper functions like [IsNotFound] and [IsUnauthorized]:
//
//	item, err := client.ContentData.Get(ctx, id)
//	if modula.IsNotFound(err) {
//	    // handle 404
//	}
//
// # Resource categories
//
// The Client exposes resources in several groups:
//   - Standard CRUD resources (ContentData, Datatypes, Fields, etc.) for the
//     public-facing content API.
//   - Admin CRUD resources (AdminContentData, AdminDatatypes, etc.) for the
//     administrative API which operates on draft/working content.
//   - Specialized resources (Auth, MediaUpload, Content delivery, Deploy, etc.)
//     with domain-specific methods beyond simple CRUD.
//   - Composite resources (ContentComposite, UserComposite, etc.) for cascade
//     operations that span multiple tables atomically.
package modula

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

// ClientConfig holds the configuration needed to create a new [Client].
//
// Only BaseURL is required. If HTTPClient is nil, a default client with a
// 30-second timeout is used. If APIKey is empty, requests are sent without
// authentication (suitable for public read-only endpoints).
type ClientConfig struct {
	// BaseURL is the root URL of the ModulaCMS server, including scheme
	// (e.g. "https://cms.example.com"). A trailing slash is stripped
	// automatically. Required.
	BaseURL string

	// APIKey is the Bearer token sent in the Authorization header on every
	// request. Leave empty for unauthenticated access to public endpoints.
	APIKey string

	// HTTPClient is the underlying [http.Client] used for all requests.
	// When nil, a default client with a 30-second timeout is created.
	// Supply a custom client to configure TLS settings, proxies, or
	// transport-level middleware.
	HTTPClient *http.Client
}

// Client is the top-level ModulaCMS API client. Create one with [NewClient]
// and use the exported resource fields to interact with the API.
//
// Resources are grouped into four categories:
//
//   - Standard CRUD: public-facing content resources (ContentData, Datatypes,
//     Fields, Media, Routes, Users, etc.). These use the generic [Resource] type.
//   - Admin CRUD: administrative resources that operate on draft/working content
//     before publishing (AdminContentData, AdminDatatypes, etc.).
//   - Specialized: domain-specific resources with custom methods (Auth,
//     MediaUpload, Content delivery, Deploy, ContentTree, etc.).
//   - Composite: cascade operations that atomically span multiple related
//     tables (ContentComposite, UserComposite, etc.).
//
// A Client is safe to use from multiple goroutines if the underlying
// [http.Client] is safe for concurrent use (the default client is).
type Client struct {
	// --- Standard CRUD resources ---

	// ContentData provides CRUD for published content items.
	ContentData *Resource[ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID]

	// ContentFields provides CRUD for content field values.
	ContentFields *Resource[ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID]

	// ContentRelations provides CRUD for content-to-content relations.
	ContentRelations *Resource[ContentRelation, CreateContentRelationParams, UpdateContentRelationParams, ContentRelationID]

	// Datatypes provides CRUD for content type definitions (schemas).
	Datatypes *Resource[Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID]

	// Fields provides CRUD for field definitions within datatypes.
	Fields *Resource[Field, CreateFieldParams, UpdateFieldParams, FieldID]

	// FieldTypes provides CRUD for field type metadata (text, number, etc.).
	FieldTypes *Resource[FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID]

	// DatatypeFields provides CRUD for the junction linking fields to datatypes.
	DatatypeFields *Resource[DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams, DatatypeFieldID]

	// Media provides read and update access to media items. For uploads, use MediaUpload.
	Media *Resource[Media, any, UpdateMediaParams, MediaID]

	// MediaDimensions provides CRUD for image dimension presets (thumbnail, medium, etc.).
	MediaDimensions *Resource[MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams, MediaDimensionID]

	// Routes provides CRUD for URL routing rules.
	Routes *Resource[Route, CreateRouteParams, UpdateRouteParams, RouteID]

	// Roles provides CRUD for RBAC roles.
	Roles *Resource[Role, CreateRoleParams, UpdateRoleParams, RoleID]

	// Permissions provides CRUD for RBAC permission definitions.
	Permissions *Resource[Permission, CreatePermissionParams, UpdatePermissionParams, PermissionID]

	// Users provides CRUD for user accounts.
	Users *Resource[User, CreateUserParams, UpdateUserParams, UserID]

	// Tokens provides CRUD for API tokens.
	Tokens *Resource[Token, CreateTokenParams, UpdateTokenParams, TokenID]

	// UsersOauth provides CRUD for OAuth provider connections on user accounts.
	UsersOauth *Resource[UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID]

	// Tables provides CRUD for custom database tables.
	Tables *Resource[Table, CreateTableParams, UpdateTableParams, TableID]

	// --- Admin CRUD resources ---

	// AdminContentData provides CRUD for draft/working content in the admin context.
	AdminContentData *Resource[AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID]

	// AdminContentFields provides CRUD for admin content field values.
	AdminContentFields *Resource[AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID]

	// AdminDatatypes provides CRUD for admin datatype definitions.
	AdminDatatypes *Resource[AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID]

	// AdminFields provides CRUD for admin field definitions.
	AdminFields *Resource[AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID]

	// AdminFieldTypes provides CRUD for admin field type metadata.
	AdminFieldTypes *Resource[AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams, AdminFieldTypeID]

	// AdminDatatypeFields provides CRUD for admin datatype-field junctions.
	AdminDatatypeFields *Resource[AdminDatatypeField, CreateAdminDatatypeFieldParams, UpdateAdminDatatypeFieldParams, AdminDatatypeFieldID]

	// AdminRoutes provides CRUD for admin routing rules.
	AdminRoutes *Resource[AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, AdminRouteID]

	// --- Specialized resources ---

	// Auth provides authentication operations (login, logout, register, password reset).
	Auth *AuthResource

	// MediaUpload provides multipart file upload for media items.
	MediaUpload *MediaUploadResource

	// AdminTree provides admin content tree traversal and manipulation.
	AdminTree *AdminTreeResource

	// Content provides public slug-based content delivery (read-only).
	Content *ContentDeliveryResource

	// Globals provides access to published global content trees.
	Globals *GlobalsResource

	// SSHKeys provides SSH public key management for user accounts.
	SSHKeys *SSHKeysResource

	// Sessions provides session listing and revocation.
	Sessions *SessionsResource

	// Import provides CMS data import from various formats.
	Import *ImportResource

	// ContentBatch provides batch content update operations.
	ContentBatch *ContentBatchResource

	// ContentTree provides content tree traversal operations.
	ContentTree *ContentTreeResource

	// ContentHeal provides content tree integrity repair operations.
	ContentHeal *ContentHealResource

	// Deploy provides content synchronization between environments (push/pull).
	Deploy *DeployResource

	// --- RBAC resources ---

	// RolePermissions provides role-to-permission assignment management.
	RolePermissions *RolePermissionsResource

	// --- Plugin resources ---

	// Plugins provides plugin listing, enable/disable, and lifecycle management.
	Plugins *PluginsResource

	// PluginRoutes provides custom routes registered by plugins.
	PluginRoutes *PluginRoutesResource

	// PluginHooks provides hook registration for plugin event handling.
	PluginHooks *PluginHooksResource

	// --- Config ---

	// Config provides CMS configuration reading and updating.
	Config *ConfigResource

	// --- Health ---

	// Health provides server health check endpoints.
	Health *HealthResource

	// --- Users full ---

	// UsersFull provides extended user operations including role details.
	UsersFull *UsersFullResource

	// --- Media folders ---

	// MediaFoldersData provides standard CRUD for media folders.
	MediaFoldersData *Resource[MediaFolder, CreateMediaFolderParams, UpdateMediaFolderParams, MediaFolderID]

	// MediaFolders provides specialized media folder operations (tree, list media, move).
	MediaFolders *MediaFoldersResource

	// --- Media admin ---

	// MediaAdmin provides administrative media operations (listing with filters, bulk actions).
	MediaAdmin *MediaAdminResource

	// --- Admin media ---

	// AdminMediaData provides standard CRUD for admin media assets (list, get, update, delete).
	AdminMediaData *Resource[AdminMedia, CreateAdminMediaParams, UpdateAdminMediaParams, AdminMediaID]

	// AdminMediaUpload provides multipart file upload for admin media assets.
	AdminMediaUpload *AdminMediaResource

	// AdminMediaFoldersData provides standard CRUD for admin media folders.
	AdminMediaFoldersData *Resource[AdminMediaFolder, CreateAdminMediaFolderParams, UpdateAdminMediaFolderParams, AdminMediaFolderID]

	// AdminMediaFolders provides specialized admin media folder operations (tree, list media, move).
	AdminMediaFolders *AdminMediaFoldersResource

	// --- Content reorder ---

	// ContentReorder provides sibling reordering for published content tree nodes.
	ContentReorder *ContentReorderResource

	// AdminContentReorder provides sibling reordering for admin content tree nodes.
	AdminContentReorder *AdminContentReorderResource

	// --- Publishing ---

	// Publishing provides publish/unpublish operations for public content.
	Publishing *PublishingResource

	// AdminPublishing provides publish/unpublish operations for admin content.
	AdminPublishing *PublishingResource

	// --- Locales ---

	// Locales provides locale (language/region) management.
	Locales *LocaleResource

	// --- Webhooks ---

	// Webhooks provides webhook registration, listing, and delivery history.
	Webhooks *WebhookResource

	// --- Content Query ---

	// Query provides advanced content querying with filters and projections.
	Query *QueryResource

	// --- Search ---

	// Search provides full-text search over published content (no auth required).
	Search *SearchResource

	// --- Content Versions ---

	// ContentVersions provides version history browsing and restoration for content items.
	ContentVersions *ContentVersionsResource

	// --- Fields extra ---

	// FieldsExtra provides sort order and max sort order operations for fields.
	FieldsExtra *FieldsExtraResource

	// --- Datatypes extra ---

	// DatatypesExtra provides sort order and max sort order operations for datatypes.
	DatatypesExtra *DatatypesExtraResource

	// AdminDatatypesExtra provides sort order and max sort order operations for admin datatypes.
	AdminDatatypesExtra *AdminDatatypesExtraResource

	// --- Validations ---

	// Validations provides CRUD and search for public validation configurations.
	Validations *ValidationResource

	// AdminValidations provides CRUD and search for admin validation configurations.
	AdminValidations *AdminValidationResource

	// --- Composite / cascade operations ---

	// ContentComposite provides atomic cascade operations across content and its related data.
	ContentComposite *ContentCompositeResource

	// UserComposite provides atomic cascade operations for users and their associated resources.
	UserComposite *UserCompositeResource

	// DatatypeComposite provides atomic cascade operations for datatypes, their fields, and content.
	DatatypeComposite *DatatypeCompositeResource

	// MediaComposite provides atomic cascade operations for media and its dimensions/references.
	MediaComposite *MediaCompositeResource
}

// NewClient creates a new [Client] configured with the given [ClientConfig].
//
// It returns an error if BaseURL is empty. The returned Client is ready to use
// immediately. All resource fields are initialized and bound to the same
// underlying HTTP transport, so authentication and timeout settings apply
// uniformly across all API calls.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("modula: BaseURL is required")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	httpCl := cfg.HTTPClient
	if httpCl == nil {
		httpCl = &http.Client{Timeout: 30 * time.Second}
	}

	h := &httpClient{
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		httpClient: httpCl,
	}

	return &Client{
		// Standard CRUD
		ContentData:      newResource[ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID](h, "/api/v1/contentdata"),
		ContentFields:    newResource[ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID](h, "/api/v1/contentfields"),
		ContentRelations: newResource[ContentRelation, CreateContentRelationParams, UpdateContentRelationParams, ContentRelationID](h, "/api/v1/contentrelations"),
		Datatypes:        newResource[Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID](h, "/api/v1/datatype"),
		Fields:           newResource[Field, CreateFieldParams, UpdateFieldParams, FieldID](h, "/api/v1/fields"),
		FieldTypes:       newResource[FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID](h, "/api/v1/fieldtypes"),
		DatatypeFields:   newResource[DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams, DatatypeFieldID](h, "/api/v1/datatypefields"),
		Media:            newResource[Media, any, UpdateMediaParams, MediaID](h, "/api/v1/media"),
		MediaDimensions:  newResource[MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams, MediaDimensionID](h, "/api/v1/mediadimensions"),
		Routes:           newResource[Route, CreateRouteParams, UpdateRouteParams, RouteID](h, "/api/v1/routes"),
		Roles:            newResource[Role, CreateRoleParams, UpdateRoleParams, RoleID](h, "/api/v1/roles"),
		Permissions:      newResource[Permission, CreatePermissionParams, UpdatePermissionParams, PermissionID](h, "/api/v1/permissions"),
		Users:            newResource[User, CreateUserParams, UpdateUserParams, UserID](h, "/api/v1/users"),
		Tokens:           newResource[Token, CreateTokenParams, UpdateTokenParams, TokenID](h, "/api/v1/tokens"),
		UsersOauth:       newResource[UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID](h, "/api/v1/usersoauth"),
		Tables:           newResource[Table, CreateTableParams, UpdateTableParams, TableID](h, "/api/v1/tables"),

		// Admin CRUD
		AdminContentData:    newResource[AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID](h, "/api/v1/admincontentdatas"),
		AdminContentFields:  newResource[AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID](h, "/api/v1/admincontentfields"),
		AdminDatatypes:      newResource[AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID](h, "/api/v1/admindatatypes"),
		AdminFields:         newResource[AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID](h, "/api/v1/adminfields"),
		AdminFieldTypes:     newResource[AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams, AdminFieldTypeID](h, "/api/v1/adminfieldtypes"),
		AdminDatatypeFields: newResource[AdminDatatypeField, CreateAdminDatatypeFieldParams, UpdateAdminDatatypeFieldParams, AdminDatatypeFieldID](h, "/api/v1/admindatatypefields"),
		AdminRoutes:         newResource[AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, AdminRouteID](h, "/api/v1/adminroutes"),

		// Specialized
		Auth:         &AuthResource{http: h},
		MediaUpload:  &MediaUploadResource{http: h},
		AdminTree:    &AdminTreeResource{http: h},
		Content:      &ContentDeliveryResource{http: h},
		Globals:      &GlobalsResource{http: h},
		SSHKeys:      &SSHKeysResource{http: h},
		Sessions:     &SessionsResource{http: h},
		Import:       &ImportResource{http: h},
		ContentBatch: &ContentBatchResource{http: h},
		ContentTree:  &ContentTreeResource{http: h},
		ContentHeal:  &ContentHealResource{http: h},
		Deploy:       &DeployResource{http: h},

		// RBAC
		RolePermissions: &RolePermissionsResource{http: h},

		// Plugin
		Plugins:      &PluginsResource{http: h},
		PluginRoutes: &PluginRoutesResource{http: h},
		PluginHooks:  &PluginHooksResource{http: h},

		// Config
		Config: &ConfigResource{http: h},

		// Health
		Health: &HealthResource{http: h},

		// Users full
		UsersFull: &UsersFullResource{http: h},

		// Media folders
		MediaFoldersData: newResource[MediaFolder, CreateMediaFolderParams, UpdateMediaFolderParams, MediaFolderID](h, "/api/v1/media-folders"),
		MediaFolders:     &MediaFoldersResource{http: h},

		// Media admin
		MediaAdmin: &MediaAdminResource{http: h},

		// Admin media
		AdminMediaData:        newResource[AdminMedia, CreateAdminMediaParams, UpdateAdminMediaParams, AdminMediaID](h, "/api/v1/adminmedia"),
		AdminMediaUpload:      &AdminMediaResource{http: h},
		AdminMediaFoldersData: newResource[AdminMediaFolder, CreateAdminMediaFolderParams, UpdateAdminMediaFolderParams, AdminMediaFolderID](h, "/api/v1/adminmedia-folders"),
		AdminMediaFolders:     &AdminMediaFoldersResource{http: h},

		// Content reorder
		ContentReorder:      &ContentReorderResource{http: h},
		AdminContentReorder: &AdminContentReorderResource{http: h},

		// Publishing
		Publishing:      &PublishingResource{http: h, prefix: "content"},
		AdminPublishing: &PublishingResource{http: h, prefix: "admin/content"},

		// Locales
		Locales: newLocaleResource(h),

		// Webhooks
		Webhooks: newWebhookResource(h),

		// Validations
		Validations:      newValidationResource(h),
		AdminValidations: newAdminValidationResource(h),

		// Content Query
		Query: &QueryResource{http: h},

		// Search
		Search: &SearchResource{http: h},

		// Content Versions
		ContentVersions: &ContentVersionsResource{http: h},

		// Fields extra
		FieldsExtra: &FieldsExtraResource{http: h},

		// Datatypes extra
		DatatypesExtra:      &DatatypesExtraResource{http: h},
		AdminDatatypesExtra: &AdminDatatypesExtraResource{http: h},

		// Composite / cascade
		ContentComposite:  &ContentCompositeResource{http: h},
		UserComposite:     &UserCompositeResource{http: h},
		DatatypeComposite: &DatatypeCompositeResource{http: h},
		MediaComposite:    &MediaCompositeResource{http: h},
	}, nil
}

func newResource[E any, C any, U any, ID ~string](h *httpClient, path string) *Resource[E, C, U, ID] {
	return &Resource[E, C, U, ID]{
		path: path,
		http: h,
	}
}
