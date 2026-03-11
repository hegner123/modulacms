package modula

import "encoding/json"

// ---------------------------------------------------------------------------
// Content Data
// ---------------------------------------------------------------------------

// ContentData represents a content entry in the CMS tree-based content structure.
// Content nodes form a hierarchy using sibling pointers (ParentID, FirstChildID,
// NextSiblingID, PrevSiblingID) for O(1) navigation and reordering.
// Each node is associated with a Datatype schema and optionally linked to a Route for URL resolution.
// Returned by content CRUD endpoints and the content tree API.
type ContentData struct {
	ContentDataID ContentID     `json:"content_data_id"`
	ParentID      *ContentID    `json:"parent_id"`
	FirstChildID  *string       `json:"first_child_id"`
	NextSiblingID *string       `json:"next_sibling_id"`
	PrevSiblingID *string       `json:"prev_sibling_id"`
	RouteID       *RouteID      `json:"route_id"`
	DatatypeID    *DatatypeID   `json:"datatype_id"`
	AuthorID      *UserID       `json:"author_id"`
	Status        ContentStatus `json:"status"`
	PublishedAt   *Timestamp    `json:"published_at,omitempty"`
	PublishedBy   *UserID       `json:"published_by,omitempty"`
	PublishAt     *Timestamp    `json:"publish_at,omitempty"`
	Revision      int64         `json:"revision"`
	DateCreated   Timestamp     `json:"date_created"`
	DateModified  Timestamp     `json:"date_modified"`
}

// CreateContentDataParams holds parameters for creating a new content node.
// DatatypeID and Status are required. ParentID determines placement in the content tree;
// sibling pointers (FirstChildID, NextSiblingID, PrevSiblingID) are typically managed
// by the server but can be set explicitly for tree operations.
type CreateContentDataParams struct {
	ParentID      *ContentID    `json:"parent_id"`
	FirstChildID  *string       `json:"first_child_id"`
	NextSiblingID *string       `json:"next_sibling_id"`
	PrevSiblingID *string       `json:"prev_sibling_id"`
	RouteID       *RouteID      `json:"route_id"`
	DatatypeID    *DatatypeID   `json:"datatype_id"`
	AuthorID      *UserID       `json:"author_id"`
	Status        ContentStatus `json:"status"`
}

// UpdateContentDataParams holds parameters for updating an existing content node.
// ContentDataID identifies the record to update. All other fields are set to their
// new values; use the current values for fields that should remain unchanged.
type UpdateContentDataParams struct {
	ContentDataID ContentID     `json:"content_data_id"`
	ParentID      *ContentID    `json:"parent_id"`
	FirstChildID  *string       `json:"first_child_id"`
	NextSiblingID *string       `json:"next_sibling_id"`
	PrevSiblingID *string       `json:"prev_sibling_id"`
	RouteID       *RouteID      `json:"route_id"`
	DatatypeID    *DatatypeID   `json:"datatype_id"`
	AuthorID      *UserID       `json:"author_id"`
	Status        ContentStatus `json:"status"`
}

// ---------------------------------------------------------------------------
// Content Field
// ---------------------------------------------------------------------------

// ContentField represents a single field value within a content entry.
// Each ContentField stores the value for one Field definition (identified by FieldID)
// on one content node (identified by ContentDataID). The Locale field supports
// internationalized content where the same field can have different values per locale.
type ContentField struct {
	ContentFieldID ContentFieldID `json:"content_field_id"`
	RouteID        *RouteID       `json:"route_id"`
	ContentDataID  *ContentID     `json:"content_data_id"`
	FieldID        *FieldID       `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	Locale         string         `json:"locale"`
	AuthorID       *UserID        `json:"author_id"`
	DateCreated    Timestamp      `json:"date_created"`
	DateModified   Timestamp      `json:"date_modified"`
}

// CreateContentFieldParams contains fields for inserting a new content field value.
// ContentDataID and FieldID are required to associate the value with a content node
// and its schema field definition. FieldValue holds the string-encoded value.
type CreateContentFieldParams struct {
	RouteID       *RouteID   `json:"route_id"`
	ContentDataID *ContentID `json:"content_data_id"`
	FieldID       *FieldID   `json:"field_id"`
	FieldValue    string     `json:"field_value"`
	AuthorID      *UserID    `json:"author_id"`
}

// UpdateContentFieldParams contains fields for modifying an existing content field value.
// ContentFieldID identifies the record to update.
type UpdateContentFieldParams struct {
	ContentFieldID ContentFieldID `json:"content_field_id"`
	RouteID        *RouteID       `json:"route_id"`
	ContentDataID  *ContentID     `json:"content_data_id"`
	FieldID        *FieldID       `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	AuthorID       *UserID        `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Content Relation
// ---------------------------------------------------------------------------

// ContentRelation represents a directional relationship between two content items
// through a reference-type field. Used for "related content" features where one
// content node references another. SortOrder controls the display ordering of
// multiple relations on the same field.
type ContentRelation struct {
	ContentRelationID ContentRelationID `json:"content_relation_id"`
	SourceContentID   ContentID         `json:"source_content_id"`
	TargetContentID   ContentID         `json:"target_content_id"`
	FieldID           FieldID           `json:"field_id"`
	SortOrder         int64             `json:"sort_order"`
	DateCreated       Timestamp         `json:"date_created"`
}

// CreateContentRelationParams specifies parameters for creating a content relation.
// SourceContentID is the owning content node, TargetContentID is the referenced node,
// and FieldID identifies which reference field on the source holds this relation.
type CreateContentRelationParams struct {
	SourceContentID ContentID `json:"source_content_id"`
	TargetContentID ContentID `json:"target_content_id"`
	FieldID         FieldID   `json:"field_id"`
	SortOrder       int64     `json:"sort_order"`
}

// UpdateContentRelationParams specifies parameters for updating a content relation.
// Only SortOrder can be changed; to change the source or target, delete and recreate.
type UpdateContentRelationParams struct {
	ContentRelationID ContentRelationID `json:"content_relation_id"`
	SortOrder         int64             `json:"sort_order"`
}

// ---------------------------------------------------------------------------
// Datatype
// ---------------------------------------------------------------------------

// Datatype represents a content schema definition, analogous to a "post type" in WordPress
// or a "content type" in other CMS platforms. Datatypes define the structure of content
// by grouping Fields together. They support hierarchical organization via ParentID.
// The Type field categorizes the datatype (e.g., "collection", "single", "component").
// Returned by the /datatypes endpoints.
type Datatype struct {
	DatatypeID   DatatypeID  `json:"datatype_id"`
	ParentID     *DatatypeID `json:"parent_id"`
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	Type         string      `json:"type"`
	AuthorID     *UserID     `json:"author_id"`
	DateCreated  Timestamp   `json:"date_created"`
	DateModified Timestamp   `json:"date_modified"`
}

// CreateDatatypeParams holds the parameters for creating a new datatype.
// Name (machine-readable identifier) and Label (human-readable display name) are required.
// Type categorizes the datatype. DatatypeID is optional; the server generates one if omitted.
type CreateDatatypeParams struct {
	DatatypeID *DatatypeID `json:"datatype_id,omitempty"`
	ParentID   *DatatypeID `json:"parent_id"`
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Type       string      `json:"type"`
	AuthorID   *UserID     `json:"author_id"`
}

// UpdateDatatypeParams holds the parameters for updating an existing datatype.
// DatatypeID identifies the record to update. All fields are set to their new values.
type UpdateDatatypeParams struct {
	DatatypeID DatatypeID  `json:"datatype_id"`
	ParentID   *DatatypeID `json:"parent_id"`
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Type       string      `json:"type"`
	AuthorID   *UserID     `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Field
// ---------------------------------------------------------------------------

// Field represents a field definition within a Datatype schema. Fields define the
// individual properties that content entries can have (e.g., "title", "body", "featured_image").
// The Type field specifies the data type (text, number, media, reference, etc.).
// Data holds type-specific configuration as JSON, Validation holds validation rules as JSON,
// and UIConfig holds rendering hints for admin interfaces as JSON.
// When Translatable is true, the field supports per-locale values for i18n content.
// Roles restricts field visibility to specific roles; nil means unrestricted access.
type Field struct {
	FieldID      FieldID     `json:"field_id"`
	ParentID     *DatatypeID `json:"parent_id"`
	SortOrder    int64       `json:"sort_order"`
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	Data         string      `json:"data"`
	Validation   string      `json:"validation"`
	UIConfig     string      `json:"ui_config"`
	Type         FieldType   `json:"type"`
	Translatable bool        `json:"translatable"`
	Roles        []string    `json:"roles"` // nil = unrestricted
	AuthorID     *UserID     `json:"author_id"`
	DateCreated  Timestamp   `json:"date_created"`
	DateModified Timestamp   `json:"date_modified"`
}

// CreateFieldParams contains parameters for creating a new field definition.
// ParentID links the field to a Datatype. Name, Label, and Type are required.
// FieldID is optional; the server generates one if omitted.
type CreateFieldParams struct {
	FieldID    *FieldID    `json:"field_id,omitempty"`
	ParentID   *DatatypeID `json:"parent_id"`
	SortOrder  int64       `json:"sort_order"`
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Data       string      `json:"data"`
	Validation string      `json:"validation"`
	UIConfig   string      `json:"ui_config"`
	Type       FieldType   `json:"type"`
	Roles      []string    `json:"roles,omitempty"` // nil = unrestricted
	AuthorID   *UserID     `json:"author_id"`
}

// UpdateFieldParams contains parameters for updating an existing field definition.
// FieldID identifies the record to update. All fields are set to their new values.
type UpdateFieldParams struct {
	FieldID    FieldID     `json:"field_id"`
	ParentID   *DatatypeID `json:"parent_id"`
	SortOrder  int64       `json:"sort_order"`
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Data       string      `json:"data"`
	Validation string      `json:"validation"`
	UIConfig   string      `json:"ui_config"`
	Type       FieldType   `json:"type"`
	Roles      []string    `json:"roles,omitempty"` // nil = unrestricted
	AuthorID   *UserID     `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Datatype-Field Link
// ---------------------------------------------------------------------------

// DatatypeField represents a junction record linking a Field to a Datatype.
// This many-to-many relationship allows fields to be shared across multiple datatypes.
// SortOrder controls the display ordering of fields within a datatype's schema.
type DatatypeField struct {
	ID           DatatypeFieldID `json:"id"`
	DatatypeID   DatatypeID      `json:"datatype_id"`
	FieldID      FieldID         `json:"field_id"`
	SortOrder    int64           `json:"sort_order"`
	DateCreated  Timestamp       `json:"date_created"`
	DateModified Timestamp       `json:"date_modified"`
}

// CreateDatatypeFieldParams holds parameters for linking an existing Field to a Datatype.
// Both DatatypeID and FieldID must reference existing records.
type CreateDatatypeFieldParams struct {
	DatatypeID DatatypeID `json:"datatype_id"`
	FieldID    FieldID    `json:"field_id"`
	SortOrder  int64      `json:"sort_order"`
}

// UpdateDatatypeFieldParams holds parameters for updating a datatype-field link,
// typically to change the SortOrder or reassign the field to a different datatype.
type UpdateDatatypeFieldParams struct {
	ID         DatatypeFieldID `json:"id"`
	DatatypeID DatatypeID      `json:"datatype_id"`
	FieldID    FieldID         `json:"field_id"`
	SortOrder  int64           `json:"sort_order"`
}

// ---------------------------------------------------------------------------
// Media
// ---------------------------------------------------------------------------

// Media represents an uploaded media asset (image, document, video, etc.) stored in
// the CMS media library. Files are stored in S3-compatible object storage. The URL
// field contains the primary access URL. Images may have Srcset for responsive variants,
// Dimensions for size metadata, and FocalX/FocalY (0.0-1.0) for focal point cropping.
// Media creation is handled via multipart upload; see MediaUploadResource.
type Media struct {
	MediaID      MediaID   `json:"media_id"`
	Name         *string   `json:"name"`
	DisplayName  *string   `json:"display_name"`
	Alt          *string   `json:"alt"`
	Caption      *string   `json:"caption"`
	Description  *string   `json:"description"`
	Class        *string   `json:"class"`
	Mimetype     *string   `json:"mimetype"`
	Dimensions   *string   `json:"dimensions"`
	URL          URL       `json:"url"`
	Srcset       *string   `json:"srcset"`
	FocalX       *float64  `json:"focal_x"`
	FocalY       *float64  `json:"focal_y"`
	AuthorID     *UserID   `json:"author_id"`
	DateCreated  Timestamp `json:"date_created"`
	DateModified Timestamp `json:"date_modified"`
}

// UpdateMediaParams contains fields for updating metadata on an existing media asset.
// Only metadata fields (name, alt text, caption, etc.) can be updated; the underlying
// file cannot be replaced. Media creation is handled via multipart upload (see MediaUploadResource).
// Pointer fields are optional; nil means no change.
type UpdateMediaParams struct {
	MediaID     MediaID  `json:"media_id"`
	Name        *string  `json:"name"`
	DisplayName *string  `json:"display_name"`
	Alt         *string  `json:"alt"`
	Caption     *string  `json:"caption"`
	Description *string  `json:"description"`
	Class       *string  `json:"class"`
	FocalX      *float64 `json:"focal_x"`
	FocalY      *float64 `json:"focal_y"`
}

// ---------------------------------------------------------------------------
// Media Dimension
// ---------------------------------------------------------------------------

// MediaDimension represents a reusable image dimension preset for responsive images.
// When media is uploaded, the CMS generates resized variants matching each defined
// MediaDimension. Width and Height define the target pixel dimensions; AspectRatio
// (e.g., "16:9") can constrain cropping behavior.
type MediaDimension struct {
	MdID        MediaDimensionID `json:"md_id"`
	Label       *string          `json:"label"`
	Width       *int64           `json:"width"`
	Height      *int64           `json:"height"`
	AspectRatio *string          `json:"aspect_ratio"`
}

// CreateMediaDimensionParams contains parameters for creating a media dimension preset.
// At least one of Width or Height should be specified.
type CreateMediaDimensionParams struct {
	Label       *string `json:"label"`
	Width       *int64  `json:"width"`
	Height      *int64  `json:"height"`
	AspectRatio *string `json:"aspect_ratio"`
}

// UpdateMediaDimensionParams contains parameters for updating a media dimension preset.
// MdID identifies the record to update. Pointer fields are optional; nil means no change.
type UpdateMediaDimensionParams struct {
	MdID        MediaDimensionID `json:"md_id"`
	Label       *string          `json:"label"`
	Width       *int64           `json:"width"`
	Height      *int64           `json:"height"`
	AspectRatio *string          `json:"aspect_ratio"`
}

// ---------------------------------------------------------------------------
// Route
// ---------------------------------------------------------------------------

// Route represents a URL route that maps a slug to content in the CMS.
// Routes provide the public URL structure for content delivery. The Slug field
// is the URL path segment (e.g., "/about" or "/blog/my-post"). Status controls
// visibility (active, inactive, redirect, etc.). Content nodes reference routes
// via RouteID to become URL-addressable.
type Route struct {
	RouteID      RouteID   `json:"route_id"`
	Slug         Slug      `json:"slug"`
	Title        string    `json:"title"`
	Status       int64     `json:"status"`
	AuthorID     *UserID   `json:"author_id"`
	DateCreated  Timestamp `json:"date_created"`
	DateModified Timestamp `json:"date_modified"`
}

// CreateRouteParams contains parameters for creating a new URL route.
// Slug and Title are required. Status defaults to active if not specified.
type CreateRouteParams struct {
	Slug     Slug    `json:"slug"`
	Title    string  `json:"title"`
	Status   int64   `json:"status"`
	AuthorID *UserID `json:"author_id"`
}

// UpdateRouteParams contains parameters for updating an existing route.
// RouteID identifies the route to update. Slug is the new slug value.
type UpdateRouteParams struct {
	RouteID  RouteID `json:"route_id"`
	Slug     Slug    `json:"slug"`
	Title    string  `json:"title"`
	Status   int64   `json:"status"`
	AuthorID *UserID `json:"author_id"`
}

// ---------------------------------------------------------------------------
// User
// ---------------------------------------------------------------------------

// User represents a CMS user account. Sensitive fields such as the password hash
// are never returned from the API. Each user has a Role that determines their
// permissions via the RBAC system. Returned by user CRUD endpoints and /auth/me.
type User struct {
	UserID       UserID    `json:"user_id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Email        Email     `json:"email"`
	Role         string    `json:"role"`
	DateCreated  Timestamp `json:"date_created"`
	DateModified Timestamp `json:"date_modified"`
}

// CreateUserParams contains parameters for creating a new user account.
// All fields are required. Password is sent in plaintext and hashed server-side.
// Role must be a valid role label (e.g., "admin", "editor", "viewer").
type CreateUserParams struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    Email  `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// UpdateUserParams contains parameters for updating an existing user account.
// UserID identifies the record to update. Password is optional (omitempty);
// when empty, the existing password is preserved.
type UpdateUserParams struct {
	UserID   UserID `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    Email  `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

// RequestPasswordResetParams contains parameters for initiating a password reset flow.
// The server sends a reset token to the provided email address.
type RequestPasswordResetParams struct {
	Email string `json:"email"`
}

// ConfirmPasswordResetParams contains parameters for completing a password reset.
// Token is the reset token received via email; Password is the new plaintext password.
type ConfirmPasswordResetParams struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// MessageResponse represents a simple server message, typically returned for
// operations that do not produce a resource (e.g., delete confirmations, status messages).
type MessageResponse struct {
	Message string `json:"message"`
}

// ---------------------------------------------------------------------------
// Role
// ---------------------------------------------------------------------------

// Role represents a role in the RBAC authorization system.
// Roles group permissions and are assigned to users. The CMS ships with three
// bootstrap roles: "admin" (full access), "editor" (content management),
// and "viewer" (read-only). System-protected roles cannot be deleted or renamed.
type Role struct {
	RoleID RoleID `json:"role_id"`
	Label  string `json:"label"`
}

// CreateRoleParams contains parameters for creating a new role.
// Label is the human-readable role name (e.g., "contributor").
type CreateRoleParams struct {
	Label string `json:"label"`
}

// UpdateRoleParams contains parameters for updating a role.
type UpdateRoleParams struct {
	RoleID RoleID `json:"role_id"`
	Label  string `json:"label"`
}

// ---------------------------------------------------------------------------
// Permission
// ---------------------------------------------------------------------------

// Permission represents a granular access control permission in the RBAC system.
// Permission labels follow the "resource:operation" format (e.g., "content:read",
// "media:create", "users:delete"). System-protected permissions cannot be deleted or renamed.
type Permission struct {
	PermissionID PermissionID `json:"permission_id"`
	Label        string       `json:"label"`
}

// CreatePermissionParams contains parameters for creating a new permission.
// Label must follow the "resource:operation" format.
type CreatePermissionParams struct {
	Label string `json:"label"`
}

// UpdatePermissionParams contains parameters for updating a permission.
type UpdatePermissionParams struct {
	PermissionID PermissionID `json:"permission_id"`
	Label        string       `json:"label"`
}

// ---------------------------------------------------------------------------
// RolePermission
// ---------------------------------------------------------------------------

// RolePermission represents a junction record that grants a Permission to a Role.
// The role_permissions table maps roles to their allowed operations.
type RolePermission struct {
	ID           RolePermissionID `json:"id"`
	RoleID       RoleID           `json:"role_id"`
	PermissionID PermissionID     `json:"permission_id"`
}

// CreateRolePermissionParams contains parameters for granting a permission to a role.
// Both RoleID and PermissionID must reference existing records.
type CreateRolePermissionParams struct {
	RoleID       RoleID       `json:"role_id"`
	PermissionID PermissionID `json:"permission_id"`
}

// ---------------------------------------------------------------------------
// Session
// ---------------------------------------------------------------------------

// Session represents an authenticated user session. Sessions are created on login
// and track expiration, IP address, user agent, and optional session data.
// The SessionData field can hold arbitrary JSON for session-scoped state.
type Session struct {
	SessionID   SessionID `json:"session_id"`
	UserID      *UserID   `json:"user_id"`
	DateCreated Timestamp `json:"date_created"`
	ExpiresAt   Timestamp `json:"expires_at"`
	LastAccess  *string   `json:"last_access"`
	IpAddress   *string   `json:"ip_address"`
	UserAgent   *string   `json:"user_agent"`
	SessionData *string   `json:"session_data"`
}

// UpdateSessionParams holds parameters for updating an existing session.
// SessionID identifies the record to update.
type UpdateSessionParams struct {
	SessionID   SessionID `json:"session_id"`
	UserID      *UserID   `json:"user_id"`
	ExpiresAt   Timestamp `json:"expires_at"`
	LastAccess  *string   `json:"last_access"`
	IpAddress   *string   `json:"ip_address"`
	UserAgent   *string   `json:"user_agent"`
	SessionData *string   `json:"session_data"`
}

// ---------------------------------------------------------------------------
// Token
// ---------------------------------------------------------------------------

// Token represents an authentication token used for API access, password resets,
// or other token-based operations. TokenType distinguishes the purpose (e.g., "api",
// "reset", "refresh"). Tokens can be revoked individually without affecting the session.
type Token struct {
	ID        TokenID   `json:"id"`
	UserID    *UserID   `json:"user_id"`
	TokenType string    `json:"token_type"`
	Token     string    `json:"token"`
	IssuedAt  string    `json:"issued_at"`
	ExpiresAt Timestamp `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// CreateTokenParams contains the parameters for creating a new authentication token.
// TokenType and Token (the actual token string) are required.
type CreateTokenParams struct {
	UserID    *UserID   `json:"user_id"`
	TokenType string    `json:"token_type"`
	Token     string    `json:"token"`
	IssuedAt  string    `json:"issued_at"`
	ExpiresAt Timestamp `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// UpdateTokenParams contains the parameters for updating an existing token.
// ID identifies the token to update. Typically used to revoke tokens or extend expiration.
type UpdateTokenParams struct {
	ID        TokenID   `json:"id"`
	Token     string    `json:"token"`
	IssuedAt  string    `json:"issued_at"`
	ExpiresAt Timestamp `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// ---------------------------------------------------------------------------
// User OAuth
// ---------------------------------------------------------------------------

// UserOauth represents an OAuth provider link for a user account.
// Stores the provider-specific user ID and OAuth tokens (access, refresh)
// for providers like Google, GitHub, and Azure. One user can have multiple
// OAuth links across different providers.
type UserOauth struct {
	UserOauthID         UserOauthID `json:"user_oauth_id"`
	UserID              *UserID     `json:"user_id"`
	OauthProvider       string      `json:"oauth_provider"`
	OauthProviderUserID string      `json:"oauth_provider_user_id"`
	AccessToken         string      `json:"access_token"`
	RefreshToken        string      `json:"refresh_token"`
	TokenExpiresAt      string      `json:"token_expires_at"`
	DateCreated         Timestamp   `json:"date_created"`
}

// CreateUserOauthParams contains the parameters for linking an OAuth provider to a user.
// OauthProvider identifies the provider (e.g., "google", "github", "azure").
type CreateUserOauthParams struct {
	UserID              *UserID   `json:"user_id"`
	OauthProvider       string    `json:"oauth_provider"`
	OauthProviderUserID string    `json:"oauth_provider_user_id"`
	AccessToken         string    `json:"access_token"`
	RefreshToken        string    `json:"refresh_token"`
	TokenExpiresAt      string    `json:"token_expires_at"`
	DateCreated         Timestamp `json:"date_created"`
}

// UpdateUserOauthParams contains the parameters for refreshing OAuth tokens.
// Only token fields can be updated; the provider and provider user ID are immutable.
type UpdateUserOauthParams struct {
	UserOauthID    UserOauthID `json:"user_oauth_id"`
	AccessToken    string      `json:"access_token"`
	RefreshToken   string      `json:"refresh_token"`
	TokenExpiresAt string      `json:"token_expires_at"`
}

// ---------------------------------------------------------------------------
// User SSH Key
// ---------------------------------------------------------------------------

// SshKey represents a user's SSH public key registered for TUI access.
// The CMS runs an SSH server (via Charmbracelet Wish) that authenticates
// users by matching their SSH key fingerprint. Returned by the SSH key
// management endpoints.
type SshKey struct {
	SshKeyID    UserSshKeyID `json:"ssh_key_id"`
	UserID      *UserID      `json:"user_id"`
	PublicKey   string       `json:"public_key"`
	KeyType     string       `json:"key_type"`
	Fingerprint string       `json:"fingerprint"`
	Label       string       `json:"label"`
	DateCreated Timestamp    `json:"date_created"`
	LastUsed    string       `json:"last_used"`
}

// SshKeyListItem represents an SSH key in list responses, omitting the full public
// key material for brevity. Contains only the fingerprint, type, and metadata.
type SshKeyListItem struct {
	SshKeyID    UserSshKeyID `json:"ssh_key_id"`
	KeyType     string       `json:"key_type"`
	Fingerprint string       `json:"fingerprint"`
	Label       string       `json:"label"`
	DateCreated Timestamp    `json:"date_created"`
	LastUsed    string       `json:"last_used"`
}

// CreateSSHKeyParams contains parameters for registering a new SSH public key.
// PublicKey is the full OpenSSH-format public key string. Label is a user-defined
// name for identification (e.g., "work laptop").
type CreateSSHKeyParams struct {
	PublicKey string `json:"public_key"`
	Label     string `json:"label"`
}

// ---------------------------------------------------------------------------
// Table
// ---------------------------------------------------------------------------

// Table represents a CMS metadata table entry used for internal schema tracking.
// Tables are registered during installation and used by the backup and migration systems.
type Table struct {
	ID       TableID `json:"id"`
	Label    string  `json:"label"`
	AuthorID *UserID `json:"author_id"`
}

// CreateTableParams holds parameters for creating a table.
type CreateTableParams struct {
	Label string `json:"label"`
}

// UpdateTableParams holds parameters for updating a table.
type UpdateTableParams struct {
	ID    TableID `json:"id"`
	Label string  `json:"label"`
}

// ---------------------------------------------------------------------------
// Admin Content Data
// ---------------------------------------------------------------------------

// AdminContentData represents a content entry in the admin-managed content namespace.
// Admin content uses a separate set of tables from user-facing content, allowing the
// CMS to manage its own internal content (e.g., dashboard widgets, system pages) with
// the same tree-based structure and sibling pointer navigation as regular content.
type AdminContentData struct {
	AdminContentDataID AdminContentID   `json:"admin_content_data_id"`
	ParentID           *AdminContentID  `json:"parent_id"`
	FirstChildID       *string          `json:"first_child_id"`
	NextSiblingID      *string          `json:"next_sibling_id"`
	PrevSiblingID      *string          `json:"prev_sibling_id"`
	AdminRouteID       *AdminRouteID    `json:"admin_route_id"`
	AdminDatatypeID    *AdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           *UserID          `json:"author_id"`
	Status             ContentStatus    `json:"status"`
	PublishedAt        *Timestamp       `json:"published_at,omitempty"`
	PublishedBy        *UserID          `json:"published_by,omitempty"`
	PublishAt          *Timestamp       `json:"publish_at,omitempty"`
	Revision           int64            `json:"revision"`
	DateCreated        Timestamp        `json:"date_created"`
	DateModified       Timestamp        `json:"date_modified"`
}

// CreateAdminContentDataParams contains fields for creating a new admin content node.
// AdminDatatypeID and Status are required.
type CreateAdminContentDataParams struct {
	ParentID        *AdminContentID  `json:"parent_id"`
	FirstChildID    *string          `json:"first_child_id"`
	NextSiblingID   *string          `json:"next_sibling_id"`
	PrevSiblingID   *string          `json:"prev_sibling_id"`
	AdminRouteID    *AdminRouteID    `json:"admin_route_id"`
	AdminDatatypeID *AdminDatatypeID `json:"admin_datatype_id"`
	AuthorID        *UserID          `json:"author_id"`
	Status          ContentStatus    `json:"status"`
}

// UpdateAdminContentDataParams contains fields for updating an existing admin content node.
// AdminContentDataID identifies the record to update.
type UpdateAdminContentDataParams struct {
	AdminContentDataID AdminContentID   `json:"admin_content_data_id"`
	ParentID           *AdminContentID  `json:"parent_id"`
	FirstChildID       *string          `json:"first_child_id"`
	NextSiblingID      *string          `json:"next_sibling_id"`
	PrevSiblingID      *string          `json:"prev_sibling_id"`
	AdminRouteID       *AdminRouteID    `json:"admin_route_id"`
	AdminDatatypeID    *AdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           *UserID          `json:"author_id"`
	Status             ContentStatus    `json:"status"`
}

// ---------------------------------------------------------------------------
// Admin Content Field
// ---------------------------------------------------------------------------

// AdminContentField represents a field value within an admin content entry.
// Mirrors ContentField but operates in the admin content namespace with
// admin-specific field and route references.
type AdminContentField struct {
	AdminContentFieldID AdminContentFieldID `json:"admin_content_field_id"`
	AdminRouteID        *AdminRouteID       `json:"admin_route_id"`
	AdminContentDataID  *AdminContentID     `json:"admin_content_data_id"`
	AdminFieldID        *AdminFieldID       `json:"admin_field_id"`
	AdminFieldValue     string              `json:"admin_field_value"`
	Locale              string              `json:"locale"`
	AuthorID            *UserID             `json:"author_id"`
	DateCreated         Timestamp           `json:"date_created"`
	DateModified        Timestamp           `json:"date_modified"`
}

// CreateAdminContentFieldParams contains fields for creating a new admin content field value.
// AdminContentDataID and AdminFieldID are required.
type CreateAdminContentFieldParams struct {
	AdminRouteID       *AdminRouteID   `json:"admin_route_id"`
	AdminContentDataID *AdminContentID `json:"admin_content_data_id"`
	AdminFieldID       *AdminFieldID   `json:"admin_field_id"`
	AdminFieldValue    string          `json:"admin_field_value"`
	AuthorID           *UserID         `json:"author_id"`
}

// UpdateAdminContentFieldParams contains fields for updating an existing admin content field value.
// AdminContentFieldID identifies the record to update.
type UpdateAdminContentFieldParams struct {
	AdminContentFieldID AdminContentFieldID `json:"admin_content_field_id"`
	AdminRouteID        *AdminRouteID       `json:"admin_route_id"`
	AdminContentDataID  *AdminContentID     `json:"admin_content_data_id"`
	AdminFieldID        *AdminFieldID       `json:"admin_field_id"`
	AdminFieldValue     string              `json:"admin_field_value"`
	AuthorID            *UserID             `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Admin Content Relation
// ---------------------------------------------------------------------------

// AdminContentRelation represents a directional relationship between two admin
// content items through a reference-type admin field. Mirrors ContentRelation
// but operates in the admin content namespace.
type AdminContentRelation struct {
	AdminContentRelationID AdminContentRelationID `json:"admin_content_relation_id"`
	SourceContentID        AdminContentID         `json:"source_content_id"`
	TargetContentID        AdminContentID         `json:"target_content_id"`
	AdminFieldID           AdminFieldID           `json:"admin_field_id"`
	SortOrder              int64                  `json:"sort_order"`
	DateCreated            Timestamp              `json:"date_created"`
}

// ---------------------------------------------------------------------------
// Admin Datatype
// ---------------------------------------------------------------------------

// AdminDatatype represents a schema definition in the admin content namespace.
// Mirrors Datatype but defines schemas for CMS-internal content rather than
// user-facing content. Admin datatypes are managed separately to avoid
// namespace collisions with user-defined schemas.
type AdminDatatype struct {
	AdminDatatypeID AdminDatatypeID  `json:"admin_datatype_id"`
	ParentID        *AdminDatatypeID `json:"parent_id"`
	Name            string           `json:"name"`
	Label           string           `json:"label"`
	Type            string           `json:"type"`
	AuthorID        *UserID          `json:"author_id"`
	DateCreated     Timestamp        `json:"date_created"`
	DateModified    Timestamp        `json:"date_modified"`
}

// CreateAdminDatatypeParams contains the parameters for creating a new admin datatype.
// Name, Label, and Type are required.
type CreateAdminDatatypeParams struct {
	ParentID *AdminDatatypeID `json:"parent_id"`
	Name     string           `json:"name"`
	Label    string           `json:"label"`
	Type     string           `json:"type"`
	AuthorID *UserID          `json:"author_id"`
}

// UpdateAdminDatatypeParams contains the parameters for updating an existing admin datatype.
// AdminDatatypeID identifies the record to update.
type UpdateAdminDatatypeParams struct {
	AdminDatatypeID AdminDatatypeID  `json:"admin_datatype_id"`
	ParentID        *AdminDatatypeID `json:"parent_id"`
	Name            string           `json:"name"`
	Label           string           `json:"label"`
	Type            string           `json:"type"`
	AuthorID        *UserID          `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Admin Field
// ---------------------------------------------------------------------------

// AdminField represents a field definition in the admin content namespace.
// Mirrors Field but belongs to AdminDatatypes rather than user-facing Datatypes.
// See Field for documentation on Data, Validation, UIConfig, Translatable, and Roles.
type AdminField struct {
	AdminFieldID AdminFieldID     `json:"admin_field_id"`
	ParentID     *AdminDatatypeID `json:"parent_id"`
	SortOrder    int64            `json:"sort_order"`
	Name         string           `json:"name"`
	Label        string           `json:"label"`
	Data         string           `json:"data"`
	Validation   string           `json:"validation"`
	UIConfig     string           `json:"ui_config"`
	Type         FieldType        `json:"type"`
	Translatable bool             `json:"translatable"`
	Roles        []string         `json:"roles"` // nil = unrestricted
	AuthorID     *UserID          `json:"author_id"`
	DateCreated  Timestamp        `json:"date_created"`
	DateModified Timestamp        `json:"date_modified"`
}

// CreateAdminFieldParams contains parameters for creating a new admin field definition.
// ParentID links the field to an AdminDatatype. Name, Label, and Type are required.
type CreateAdminFieldParams struct {
	ParentID   *AdminDatatypeID `json:"parent_id"`
	SortOrder  int64            `json:"sort_order"`
	Name       string           `json:"name"`
	Label      string           `json:"label"`
	Data       string           `json:"data"`
	Validation string           `json:"validation"`
	UIConfig   string           `json:"ui_config"`
	Type       FieldType        `json:"type"`
	Roles      []string         `json:"roles,omitempty"` // nil = unrestricted
	AuthorID   *UserID          `json:"author_id"`
}

// UpdateAdminFieldParams contains parameters for updating an existing admin field definition.
// AdminFieldID identifies the record to update.
type UpdateAdminFieldParams struct {
	AdminFieldID AdminFieldID     `json:"admin_field_id"`
	ParentID     *AdminDatatypeID `json:"parent_id"`
	SortOrder    int64            `json:"sort_order"`
	Name         string           `json:"name"`
	Label        string           `json:"label"`
	Data         string           `json:"data"`
	Validation   string           `json:"validation"`
	UIConfig     string           `json:"ui_config"`
	Type         FieldType        `json:"type"`
	Roles        []string         `json:"roles,omitempty"` // nil = unrestricted
	AuthorID     *UserID          `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Admin Datatype-Field Link
// ---------------------------------------------------------------------------

// AdminDatatypeField represents a junction record linking an AdminField to an AdminDatatype.
// Mirrors DatatypeField but operates in the admin content namespace.
type AdminDatatypeField struct {
	ID              AdminDatatypeFieldID `json:"id"`
	AdminDatatypeID AdminDatatypeID      `json:"admin_datatype_id"`
	AdminFieldID    AdminFieldID         `json:"admin_field_id"`
	SortOrder       int64                `json:"sort_order"`
	DateCreated     Timestamp            `json:"date_created"`
	DateModified    Timestamp            `json:"date_modified"`
}

// CreateAdminDatatypeFieldParams holds parameters for linking an AdminField to an AdminDatatype.
// Both AdminDatatypeID and AdminFieldID must reference existing records.
type CreateAdminDatatypeFieldParams struct {
	AdminDatatypeID AdminDatatypeID `json:"admin_datatype_id"`
	AdminFieldID    AdminFieldID    `json:"admin_field_id"`
}

// UpdateAdminDatatypeFieldParams holds parameters for updating an admin datatype-field link.
type UpdateAdminDatatypeFieldParams struct {
	ID              AdminDatatypeFieldID `json:"id"`
	AdminDatatypeID AdminDatatypeID      `json:"admin_datatype_id"`
	AdminFieldID    AdminFieldID         `json:"admin_field_id"`
}

// ---------------------------------------------------------------------------
// Field Type
// ---------------------------------------------------------------------------

// FieldTypeInfo represents a registered field type in the CMS type system.
// Field types define the available data types for schema fields (e.g., "text",
// "number", "media", "reference", "boolean", "select"). Custom field types
// can be registered to extend the CMS.
type FieldTypeInfo struct {
	FieldTypeID FieldTypeID `json:"field_type_id"`
	Type        string      `json:"type"`
	Label       string      `json:"label"`
}

// CreateFieldTypeParams contains parameters for registering a new field type.
// Type is the machine-readable identifier; Label is the human-readable display name.
type CreateFieldTypeParams struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

// UpdateFieldTypeParams contains parameters for updating a field type.
type UpdateFieldTypeParams struct {
	FieldTypeID FieldTypeID `json:"field_type_id"`
	Type        string      `json:"type"`
	Label       string      `json:"label"`
}

// ---------------------------------------------------------------------------
// Admin Field Type
// ---------------------------------------------------------------------------

// AdminFieldTypeInfo represents a registered field type in the admin content type system.
// Mirrors FieldTypeInfo but operates in the admin namespace.
type AdminFieldTypeInfo struct {
	AdminFieldTypeID AdminFieldTypeID `json:"admin_field_type_id"`
	Type             string           `json:"type"`
	Label            string           `json:"label"`
}

// CreateAdminFieldTypeParams contains parameters for creating an admin field type.
type CreateAdminFieldTypeParams struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

// UpdateAdminFieldTypeParams contains parameters for updating an admin field type.
type UpdateAdminFieldTypeParams struct {
	AdminFieldTypeID AdminFieldTypeID `json:"admin_field_type_id"`
	Type             string           `json:"type"`
	Label            string           `json:"label"`
}

// ---------------------------------------------------------------------------
// Admin Route
// ---------------------------------------------------------------------------

// AdminRoute represents a URL route in the admin content namespace.
// Mirrors Route but provides URL mappings for admin-managed content rather
// than user-facing content.
type AdminRoute struct {
	AdminRouteID AdminRouteID `json:"admin_route_id"`
	Slug         Slug         `json:"slug"`
	Title        string       `json:"title"`
	Status       int64        `json:"status"`
	AuthorID     *UserID      `json:"author_id"`
	DateCreated  Timestamp    `json:"date_created"`
	DateModified Timestamp    `json:"date_modified"`
}

// CreateAdminRouteParams contains parameters for creating a new admin route.
// Slug and Title are required.
type CreateAdminRouteParams struct {
	Slug     Slug    `json:"slug"`
	Title    string  `json:"title"`
	Status   int64   `json:"status"`
	AuthorID *UserID `json:"author_id"`
}

// UpdateAdminRouteParams contains parameters for updating an existing admin route.
// AdminRouteID identifies the route to update. Slug is the new slug value.
type UpdateAdminRouteParams struct {
	AdminRouteID AdminRouteID `json:"route_id"`
	Slug         Slug         `json:"slug"`
	Title        string       `json:"title"`
	Status       int64        `json:"status"`
	AuthorID     *UserID      `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Auth Types
// ---------------------------------------------------------------------------

// LoginParams contains credentials for password-based authentication via POST /auth/login.
type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned by POST /auth/login after successful authentication.
// Contains the authenticated user's basic profile information.
type LoginResponse struct {
	UserID      UserID    `json:"user_id"`
	Email       Email     `json:"email"`
	Username    string    `json:"username"`
	DateCreated Timestamp `json:"date_created"`
}

// MeResponse: The GET /auth/me endpoint returns a User struct directly.
// No separate response type is needed since the User type matches the response shape.

// ---------------------------------------------------------------------------
// Import Types
// ---------------------------------------------------------------------------

// ImportResult represents the outcome of a CMS data import operation.
// Returned by the import endpoint after processing content from external
// CMS formats (WordPress, Contentful, Strapi, Sanity, etc.).
type ImportResult struct {
	Success          bool     `json:"success"`
	DatatypesCreated int      `json:"datatypes_created"`
	FieldsCreated    int      `json:"fields_created"`
	ContentCreated   int      `json:"content_created"`
	Errors           []string `json:"errors,omitempty"`
	Message          string   `json:"message"`
}

// ---------------------------------------------------------------------------
// Change Event (read-only)
// ---------------------------------------------------------------------------

// ChangeEvent represents a read-only audit trail entry for database mutations.
// Every audited write operation (create, update, delete) generates a ChangeEvent
// recording the table, record ID, operation type, old/new JSON values, and request
// metadata (user, IP, request ID). Used for audit logging and cross-node replication.
// ChangeEvents are immutable once created.
type ChangeEvent struct {
	EventID       EventID         `json:"event_id"`
	HlcTimestamp  int64           `json:"hlc_timestamp"`
	WallTimestamp Timestamp       `json:"wall_timestamp"`
	NodeID        string          `json:"node_id"`
	TableName     string          `json:"table_name"`
	RecordID      string          `json:"record_id"`
	Operation     string          `json:"operation"`
	Action        string          `json:"action"`
	UserID        *UserID         `json:"user_id"`
	OldValues     json.RawMessage `json:"old_values"`
	NewValues     json.RawMessage `json:"new_values"`
	Metadata      json.RawMessage `json:"metadata"`
	RequestID     *string         `json:"request_id"`
	IP            *string         `json:"ip"`
	SyncedAt      Timestamp       `json:"synced_at"`
	ConsumedAt    Timestamp       `json:"consumed_at"`
}

// ---------------------------------------------------------------------------
// Backup (read-only)
// ---------------------------------------------------------------------------

// Backup represents a read-only backup record that tracks database backup operations.
// Backups can be full or incremental, stored locally or in S3-compatible storage.
// Status tracks the backup lifecycle ("running", "completed", "failed").
// SizeBytes and RecordCount provide metrics; Checksum enables integrity verification.
type Backup struct {
	BackupID       BackupID        `json:"backup_id"`
	NodeID         string          `json:"node_id"`
	BackupType     string          `json:"backup_type"`
	Status         string          `json:"status"`
	StartedAt      Timestamp       `json:"started_at"`
	CompletedAt    Timestamp       `json:"completed_at"`
	DurationMs     *int64          `json:"duration_ms"`
	RecordCount    *int64          `json:"record_count"`
	SizeBytes      *int64          `json:"size_bytes"`
	ReplicationLsn *string         `json:"replication_lsn"`
	HlcTimestamp   int64           `json:"hlc_timestamp"`
	StoragePath    string          `json:"storage_path"`
	Checksum       *string         `json:"checksum"`
	TriggeredBy    *string         `json:"triggered_by"`
	ErrorMessage   *string         `json:"error_message"`
	Metadata       json.RawMessage `json:"metadata"`
}

// ---------------------------------------------------------------------------
// Content Version
// ---------------------------------------------------------------------------

// ContentVersion represents an immutable snapshot of a content node's field values
// at a point in time. Versions are created automatically on publish or manually via
// the versioning API. The Snapshot field contains a JSON representation of all field
// values. Trigger indicates what created the version (e.g., "publish", "manual", "restore").
// When Published is true, this version is the currently live version for the given Locale.
type ContentVersion struct {
	ContentVersionID ContentVersionID `json:"content_version_id"`
	ContentDataID    ContentID        `json:"content_data_id"`
	VersionNumber    int64            `json:"version_number"`
	Locale           string           `json:"locale"`
	Snapshot         string           `json:"snapshot"`
	Trigger          string           `json:"trigger"`
	Label            string           `json:"label"`
	Published        bool             `json:"published"`
	PublishedBy      *UserID          `json:"published_by,omitempty"`
	DateCreated      Timestamp        `json:"date_created"`
}

// AdminContentVersion represents an immutable snapshot of an admin content node's
// field values. Mirrors ContentVersion but operates in the admin content namespace.
type AdminContentVersion struct {
	AdminContentVersionID AdminContentVersionID `json:"admin_content_version_id"`
	AdminContentDataID    AdminContentID        `json:"admin_content_data_id"`
	VersionNumber         int64                 `json:"version_number"`
	Locale                string                `json:"locale"`
	Snapshot              string                `json:"snapshot"`
	Trigger               string                `json:"trigger"`
	Label                 string                `json:"label"`
	Published             bool                  `json:"published"`
	PublishedBy           *UserID               `json:"published_by,omitempty"`
	DateCreated           Timestamp             `json:"date_created"`
}

// ---------------------------------------------------------------------------
// Publishing Request/Response Types
// ---------------------------------------------------------------------------

// PublishRequest is the request body for POST /content/publish, which creates a
// published version snapshot and sets the content status to "published".
// Locale is optional; when omitted, the default locale is used.
type PublishRequest struct {
	ContentDataID ContentID `json:"content_data_id"`
	Locale        string    `json:"locale,omitempty"`
}

// AdminPublishRequest is the request body for publishing admin content.
// Mirrors PublishRequest but targets admin content namespace.
type AdminPublishRequest struct {
	AdminContentDataID AdminContentID `json:"admin_content_data_id"`
	Locale             string         `json:"locale,omitempty"`
}

// PublishResponse is returned by publish and unpublish operations.
// Status indicates the outcome (e.g., "published", "unpublished").
// VersionNumber and ContentVersionID are populated on successful publish.
type PublishResponse struct {
	Status           string `json:"status"`
	VersionNumber    int64  `json:"version_number,omitempty"`
	ContentVersionID string `json:"content_version_id,omitempty"`
	ContentDataID    string `json:"content_data_id"`
}

// AdminPublishResponse is returned by admin publish and unpublish operations.
// Mirrors PublishResponse but references admin content IDs.
type AdminPublishResponse struct {
	Status                string `json:"status"`
	VersionNumber         int64  `json:"version_number,omitempty"`
	AdminContentVersionID string `json:"admin_content_version_id,omitempty"`
	AdminContentDataID    string `json:"admin_content_data_id"`
}

// ScheduleRequest is the request body for scheduling future content publication.
// PublishAt is an ISO 8601 timestamp specifying when the content should go live.
type ScheduleRequest struct {
	ContentDataID ContentID `json:"content_data_id"`
	PublishAt     string    `json:"publish_at"`
}

// AdminScheduleRequest is the request body for scheduling future admin content publication.
// Mirrors ScheduleRequest but targets admin content namespace.
type AdminScheduleRequest struct {
	AdminContentDataID AdminContentID `json:"admin_content_data_id"`
	PublishAt          string         `json:"publish_at"`
}

// ScheduleResponse is returned after successfully scheduling content for future publication.
type ScheduleResponse struct {
	Status        string `json:"status"`
	ContentDataID string `json:"content_data_id"`
	PublishAt     string `json:"publish_at"`
}

// AdminScheduleResponse is returned after scheduling admin content for future publication.
type AdminScheduleResponse struct {
	Status             string `json:"status"`
	AdminContentDataID string `json:"admin_content_data_id"`
	PublishAt          string `json:"publish_at"`
}

// CreateVersionRequest is the request body for manually creating a content version
// snapshot without publishing. Label is an optional human-readable name for the version.
type CreateVersionRequest struct {
	ContentDataID ContentID `json:"content_data_id"`
	Label         string    `json:"label,omitempty"`
}

// CreateAdminVersionRequest is the request body for manually creating an admin content
// version snapshot. Mirrors CreateVersionRequest but targets admin content namespace.
type CreateAdminVersionRequest struct {
	AdminContentDataID AdminContentID `json:"admin_content_data_id"`
	Label              string         `json:"label,omitempty"`
}

// RestoreRequest is the request body for restoring content field values from a
// previous version snapshot. The content node's fields are overwritten with the
// values from the specified ContentVersionID.
type RestoreRequest struct {
	ContentDataID    ContentID        `json:"content_data_id"`
	ContentVersionID ContentVersionID `json:"content_version_id"`
}

// AdminRestoreRequest is the request body for restoring admin content to a previous version.
// Mirrors RestoreRequest but targets admin content namespace.
type AdminRestoreRequest struct {
	AdminContentDataID    AdminContentID        `json:"admin_content_data_id"`
	AdminContentVersionID AdminContentVersionID `json:"admin_content_version_id"`
}

// RestoreResponse is returned after restoring content to a previous version.
// FieldsRestored indicates how many fields were updated. UnmappedFields lists
// field names present in the version snapshot but not in the current schema.
type RestoreResponse struct {
	Status          string   `json:"status"`
	ContentDataID   string   `json:"content_data_id"`
	RestoredVersion string   `json:"restored_version_id"`
	FieldsRestored  int      `json:"fields_restored"`
	UnmappedFields  []string `json:"unmapped_fields,omitempty"`
}

// AdminRestoreResponse is returned after restoring admin content to a previous version.
// Mirrors RestoreResponse but references admin content IDs.
type AdminRestoreResponse struct {
	Status             string   `json:"status"`
	AdminContentDataID string   `json:"admin_content_data_id"`
	RestoredVersion    string   `json:"restored_version_id"`
	FieldsRestored     int      `json:"fields_restored"`
	UnmappedFields     []string `json:"unmapped_fields,omitempty"`
}

// ---------------------------------------------------------------------------
// Locale
// ---------------------------------------------------------------------------

// Locale represents a language/region configuration for internationalized content.
// The CMS supports multiple locales with fallback chains: when content is not available
// in the requested locale, the CMS falls back to FallbackCode (e.g., "en-US" falls back
// to "en"). One locale is marked IsDefault as the primary content language.
// SortOrder controls the display ordering in locale selection UI.
type Locale struct {
	LocaleID     LocaleID `json:"locale_id"`
	Code         string   `json:"code"`
	Label        string   `json:"label"`
	IsDefault    bool     `json:"is_default"`
	IsEnabled    bool     `json:"is_enabled"`
	FallbackCode string   `json:"fallback_code"`
	SortOrder    int64    `json:"sort_order"`
	DateCreated  string   `json:"date_created"`
}

// CreateLocaleRequest contains parameters for creating a new locale.
// Code (e.g., "en", "fr-CA") and Label are required. Setting IsDefault to true
// makes this the primary content locale, clearing the flag on any previous default.
type CreateLocaleRequest struct {
	Code         string `json:"code"`
	Label        string `json:"label"`
	IsDefault    bool   `json:"is_default"`
	IsEnabled    bool   `json:"is_enabled"`
	FallbackCode string `json:"fallback_code,omitempty"`
	SortOrder    int64  `json:"sort_order"`
}

// UpdateLocaleRequest contains parameters for updating an existing locale configuration.
// LocaleID identifies the record to update.
type UpdateLocaleRequest struct {
	LocaleID     LocaleID `json:"locale_id"`
	Code         string   `json:"code"`
	Label        string   `json:"label"`
	IsDefault    bool     `json:"is_default"`
	IsEnabled    bool     `json:"is_enabled"`
	FallbackCode string   `json:"fallback_code,omitempty"`
	SortOrder    int64    `json:"sort_order"`
}

// ---------------------------------------------------------------------------
// Webhook
// ---------------------------------------------------------------------------

// Webhook represents a registered webhook endpoint that receives HTTP POST
// notifications when CMS events occur (e.g., content published, media uploaded).
// Events lists the subscribed event types. Secret is used to sign payloads with
// HMAC-SHA256 for verification. Headers are custom HTTP headers sent with each delivery.
type Webhook struct {
	WebhookID    WebhookID         `json:"webhook_id"`
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Secret       string            `json:"secret"`
	Events       []string          `json:"events"`
	IsActive     bool              `json:"is_active"`
	Headers      map[string]string `json:"headers"`
	AuthorID     UserID            `json:"author_id"`
	DateCreated  Timestamp         `json:"date_created"`
	DateModified Timestamp         `json:"date_modified"`
}

// CreateWebhookRequest contains parameters for registering a new webhook endpoint.
// Name, URL, and Events are required. Secret is optional; when provided, payloads
// are signed for verification. Headers are optional custom HTTP headers.
type CreateWebhookRequest struct {
	Name     string            `json:"name"`
	URL      string            `json:"url"`
	Secret   string            `json:"secret,omitempty"`
	Events   []string          `json:"events"`
	IsActive bool              `json:"is_active"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// UpdateWebhookRequest contains parameters for updating an existing webhook endpoint.
// WebhookID identifies the record to update.
type UpdateWebhookRequest struct {
	WebhookID WebhookID         `json:"webhook_id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Secret    string            `json:"secret,omitempty"`
	Events    []string          `json:"events"`
	IsActive  bool              `json:"is_active"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// WebhookDelivery represents a single delivery attempt for a webhook event.
// Tracks delivery status, retry attempts, HTTP status codes, and errors.
// Failed deliveries are retried with exponential backoff up to a maximum attempt count.
type WebhookDelivery struct {
	DeliveryID     WebhookDeliveryID `json:"delivery_id"`
	WebhookID      WebhookID         `json:"webhook_id"`
	Event          string            `json:"event"`
	Payload        string            `json:"payload"`
	Status         string            `json:"status"`
	Attempts       int64             `json:"attempts"`
	LastStatusCode int64             `json:"last_status_code"`
	LastError      string            `json:"last_error"`
	NextRetryAt    string            `json:"next_retry_at"`
	CreatedAt      Timestamp         `json:"created_at"`
	CompletedAt    string            `json:"completed_at"`
}

// WebhookTestResponse is returned by the webhook test endpoint, which sends a
// test payload to the webhook URL and reports the result without creating a
// persistent delivery record.
type WebhookTestResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Translation
// ---------------------------------------------------------------------------

// CreateTranslationRequest is the request body for creating locale-specific field values
// for a content item. Copies all translatable fields from the default locale into the
// specified target locale as a starting point for translation.
type CreateTranslationRequest struct {
	Locale string `json:"locale"`
}

// CreateTranslationResponse is returned after creating translation field values.
// FieldsCreated indicates how many content fields were copied to the target locale.
type CreateTranslationResponse struct {
	Locale        string `json:"locale"`
	FieldsCreated int    `json:"fields_created"`
}

// ---------------------------------------------------------------------------
// Content Composite (create with fields)
// ---------------------------------------------------------------------------

// ContentCreateParams holds parameters for atomically creating a content node and its
// field values in a single request. This is a convenience wrapper over separate
// content data and content field creation. DatatypeID is required. Fields is a map
// of field names to values.
type ContentCreateParams struct {
	ParentID   string            `json:"parent_id,omitempty"`
	RouteID    string            `json:"route_id,omitempty"`
	DatatypeID string            `json:"datatype_id"`
	Status     string            `json:"status,omitempty"`
	Fields     map[string]string `json:"fields,omitempty"`
}

// ContentCreateResponse is returned by the composite content creation endpoint.
// Contains the created content node, all successfully created fields, and any
// errors encountered during field creation. Partial success is possible:
// FieldsCreated + FieldsFailed == len(Fields) from the request.
type ContentCreateResponse struct {
	ContentData   ContentData    `json:"content_data"`
	Fields        []ContentField `json:"fields"`
	FieldsCreated int            `json:"fields_created"`
	FieldsFailed  int            `json:"fields_failed"`
	Errors        []string       `json:"errors"`
}

// ---------------------------------------------------------------------------
// User Reassign-Delete
// ---------------------------------------------------------------------------

// UserReassignDeleteParams holds parameters for safely deleting a user by first
// reassigning all their owned content, datatypes, and admin content to another user.
// ReassignTo is optional; when omitted, a default admin user receives ownership.
type UserReassignDeleteParams struct {
	UserID     string `json:"user_id"`
	ReassignTo string `json:"reassign_to,omitempty"`
}

// UserReassignDeleteResponse is returned after reassigning and deleting a user.
// Reports the counts of reassigned records across content, datatypes, and admin content.
type UserReassignDeleteResponse struct {
	DeletedUserID              string `json:"deleted_user_id"`
	ReassignedTo               string `json:"reassigned_to"`
	ContentDataReassigned      int64  `json:"content_data_reassigned"`
	DatatypesReassigned        int64  `json:"datatypes_reassigned"`
	AdminContentDataReassigned int64  `json:"admin_content_data_reassigned"`
}

// ---------------------------------------------------------------------------
// Datatype Cascade Delete
// ---------------------------------------------------------------------------

// DatatypeCascadeDeleteResponse is returned after a cascade datatype deletion.
// This operation removes the datatype, all its field associations, and all content
// nodes that use the datatype. ContentDeleted reports how many content nodes were removed.
type DatatypeCascadeDeleteResponse struct {
	DeletedDatatypeID string   `json:"deleted_datatype_id"`
	ContentDeleted    int      `json:"content_deleted"`
	Errors            []string `json:"errors"`
}

// ---------------------------------------------------------------------------
// Recursive Content Delete
// ---------------------------------------------------------------------------

// RecursiveDeleteResponse is returned after recursively deleting a content tree branch.
// Removes the specified root node and all of its descendants (children, grandchildren, etc.).
// DeletedIDs lists every content node ID that was removed.
type RecursiveDeleteResponse struct {
	DeletedRoot  string   `json:"deleted_root"`
	TotalDeleted int      `json:"total_deleted"`
	DeletedIDs   []string `json:"deleted_ids"`
}

// ---------------------------------------------------------------------------
// Media References
// ---------------------------------------------------------------------------

// MediaReferenceInfo describes a single content field that holds a reference to a
// media asset. Used to track where media items are used across content, enabling
// safe deletion checks and usage audits.
type MediaReferenceInfo struct {
	ContentFieldID string `json:"content_field_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
}

// MediaReferenceScanResponse is returned by the media reference scan endpoint.
// Lists all content fields that reference a given media asset, enabling clients
// to check for usage before deleting media or to find all places a media item appears.
type MediaReferenceScanResponse struct {
	MediaID        string               `json:"media_id"`
	References     []MediaReferenceInfo `json:"references"`
	ReferenceCount int                  `json:"reference_count"`
}
