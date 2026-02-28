package modula

import "encoding/json"

// ---------------------------------------------------------------------------
// Content Data
// ---------------------------------------------------------------------------

// ContentData represents a content entry in the tree-based content structure.
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

// CreateContentDataParams holds parameters for creating a new content_data record.
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

// UpdateContentDataParams holds parameters for updating an existing content_data record.
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

// ContentField represents a single content field record.
type ContentField struct {
	ContentFieldID ContentFieldID `json:"content_field_id"`
	RouteID        *RouteID       `json:"route_id"`
	ContentDataID  *ContentID     `json:"content_data_id"`
	FieldID        *FieldID       `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	AuthorID       *UserID        `json:"author_id"`
	DateCreated    Timestamp      `json:"date_created"`
	DateModified   Timestamp      `json:"date_modified"`
}

// CreateContentFieldParams contains fields for inserting a new content field.
type CreateContentFieldParams struct {
	RouteID       *RouteID   `json:"route_id"`
	ContentDataID *ContentID `json:"content_data_id"`
	FieldID       *FieldID   `json:"field_id"`
	FieldValue    string     `json:"field_value"`
	AuthorID      *UserID    `json:"author_id"`
}

// UpdateContentFieldParams contains fields for modifying an existing content field.
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

// ContentRelation represents a relation between two content items through a field.
type ContentRelation struct {
	ContentRelationID ContentRelationID `json:"content_relation_id"`
	SourceContentID   ContentID         `json:"source_content_id"`
	TargetContentID   ContentID         `json:"target_content_id"`
	FieldID           FieldID           `json:"field_id"`
	SortOrder         int64             `json:"sort_order"`
	DateCreated       Timestamp         `json:"date_created"`
}

// CreateContentRelationParams specifies parameters for creating a content relation.
type CreateContentRelationParams struct {
	SourceContentID ContentID `json:"source_content_id"`
	TargetContentID ContentID `json:"target_content_id"`
	FieldID         FieldID   `json:"field_id"`
	SortOrder       int64     `json:"sort_order"`
}

// UpdateContentRelationParams specifies parameters for updating a content relation's sort order.
type UpdateContentRelationParams struct {
	ContentRelationID ContentRelationID `json:"content_relation_id"`
	SortOrder         int64             `json:"sort_order"`
}

// ---------------------------------------------------------------------------
// Datatype
// ---------------------------------------------------------------------------

// Datatype represents a datatype record.
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
type CreateDatatypeParams struct {
	DatatypeID *DatatypeID `json:"datatype_id,omitempty"`
	ParentID   *DatatypeID `json:"parent_id"`
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Type       string      `json:"type"`
	AuthorID   *UserID     `json:"author_id"`
}

// UpdateDatatypeParams holds the parameters for updating an existing datatype.
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

// Field represents a field definition.
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
	AuthorID     *UserID     `json:"author_id"`
	DateCreated  Timestamp   `json:"date_created"`
	DateModified Timestamp   `json:"date_modified"`
}

// CreateFieldParams contains parameters for creating a field.
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
	AuthorID   *UserID     `json:"author_id"`
}

// UpdateFieldParams contains parameters for updating a field.
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
	AuthorID   *UserID     `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Media
// ---------------------------------------------------------------------------

// Media represents a media asset.
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

// UpdateMediaParams contains fields for updating an existing media entry.
// Media creation is handled via multipart upload (see MediaUploadResource).
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

// MediaDimension represents a media dimension preset.
type MediaDimension struct {
	MdID        MediaDimensionID `json:"md_id"`
	Label       *string          `json:"label"`
	Width       *int64           `json:"width"`
	Height      *int64           `json:"height"`
	AspectRatio *string          `json:"aspect_ratio"`
}

// CreateMediaDimensionParams contains parameters for creating a media dimension.
type CreateMediaDimensionParams struct {
	Label       *string `json:"label"`
	Width       *int64  `json:"width"`
	Height      *int64  `json:"height"`
	AspectRatio *string `json:"aspect_ratio"`
}

// UpdateMediaDimensionParams contains parameters for updating a media dimension.
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

// Route represents a URL route.
type Route struct {
	RouteID      RouteID   `json:"route_id"`
	Slug         Slug      `json:"slug"`
	Title        string    `json:"title"`
	Status       int64     `json:"status"`
	AuthorID     *UserID   `json:"author_id"`
	DateCreated  Timestamp `json:"date_created"`
	DateModified Timestamp `json:"date_modified"`
}

// CreateRouteParams contains parameters for creating a new route.
type CreateRouteParams struct {
	Slug     Slug    `json:"slug"`
	Title    string  `json:"title"`
	Status   int64   `json:"status"`
	AuthorID *UserID `json:"author_id"`
}

// UpdateRouteParams contains parameters for updating an existing route.
type UpdateRouteParams struct {
	Slug     Slug    `json:"slug"`
	Title    string  `json:"title"`
	Status   int64   `json:"status"`
	AuthorID *UserID `json:"author_id"`
	Slug2    Slug    `json:"slug_2"`
}

// ---------------------------------------------------------------------------
// User
// ---------------------------------------------------------------------------

// User represents a user record without sensitive fields (hash is omitted).
type User struct {
	UserID       UserID    `json:"user_id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Email        Email     `json:"email"`
	Role         string    `json:"role"`
	DateCreated  Timestamp `json:"date_created"`
	DateModified Timestamp `json:"date_modified"`
}

// CreateUserParams contains parameters for creating a new user.
type CreateUserParams struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    Email  `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// UpdateUserParams contains parameters for updating an existing user.
type UpdateUserParams struct {
	UserID   UserID `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    Email  `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

// ResetPasswordParams contains parameters for resetting a user's password.
// Deprecated: Use RequestPasswordResetParams and ConfirmPasswordResetParams instead.
type ResetPasswordParams struct {
	Email       Email  `json:"email"`
	NewPassword string `json:"new_password"`
	Token       string `json:"token"`
}

// RequestPasswordResetParams contains parameters for requesting a password reset email.
type RequestPasswordResetParams struct {
	Email string `json:"email"`
}

// ConfirmPasswordResetParams contains parameters for confirming a password reset with a token.
type ConfirmPasswordResetParams struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// MessageResponse represents a simple server message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// ---------------------------------------------------------------------------
// Role
// ---------------------------------------------------------------------------

// Role represents a role entity.
type Role struct {
	RoleID RoleID `json:"role_id"`
	Label  string `json:"label"`
}

// CreateRoleParams contains parameters for creating a role.
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

// Permission represents a permission entity with access control information.
type Permission struct {
	PermissionID PermissionID `json:"permission_id"`
	Label        string       `json:"label"`
}

// CreatePermissionParams contains parameters for creating a permission.
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

// RolePermission represents a junction between a role and a permission.
type RolePermission struct {
	ID           RolePermissionID `json:"id"`
	RoleID       RoleID           `json:"role_id"`
	PermissionID PermissionID     `json:"permission_id"`
}

// CreateRolePermissionParams contains parameters for creating a role-permission association.
type CreateRolePermissionParams struct {
	RoleID       RoleID       `json:"role_id"`
	PermissionID PermissionID `json:"permission_id"`
}

// ---------------------------------------------------------------------------
// Session
// ---------------------------------------------------------------------------

// Session represents a user session.
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

// Token represents an authentication token.
type Token struct {
	ID        TokenID   `json:"id"`
	UserID    *UserID   `json:"user_id"`
	TokenType string    `json:"token_type"`
	Token     string    `json:"token"`
	IssuedAt  string    `json:"issued_at"`
	ExpiresAt Timestamp `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// CreateTokenParams contains the parameters for creating a new token.
type CreateTokenParams struct {
	UserID    *UserID   `json:"user_id"`
	TokenType string    `json:"token_type"`
	Token     string    `json:"token"`
	IssuedAt  string    `json:"issued_at"`
	ExpiresAt Timestamp `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// UpdateTokenParams contains the parameters for updating an existing token.
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

// UserOauth represents an OAuth token record for a user.
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

// CreateUserOauthParams contains the parameters for creating a new user OAuth record.
type CreateUserOauthParams struct {
	UserID              *UserID   `json:"user_id"`
	OauthProvider       string    `json:"oauth_provider"`
	OauthProviderUserID string    `json:"oauth_provider_user_id"`
	AccessToken         string    `json:"access_token"`
	RefreshToken        string    `json:"refresh_token"`
	TokenExpiresAt      string    `json:"token_expires_at"`
	DateCreated         Timestamp `json:"date_created"`
}

// UpdateUserOauthParams contains the parameters for updating a user OAuth record.
type UpdateUserOauthParams struct {
	UserOauthID    UserOauthID `json:"user_oauth_id"`
	AccessToken    string      `json:"access_token"`
	RefreshToken   string      `json:"refresh_token"`
	TokenExpiresAt string      `json:"token_expires_at"`
}

// ---------------------------------------------------------------------------
// User SSH Key
// ---------------------------------------------------------------------------

// SshKey represents a user's SSH public key.
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

// SshKeyListItem represents an SSH key in list responses (without full public key).
type SshKeyListItem struct {
	SshKeyID    UserSshKeyID `json:"ssh_key_id"`
	KeyType     string       `json:"key_type"`
	Fingerprint string       `json:"fingerprint"`
	Label       string       `json:"label"`
	DateCreated Timestamp    `json:"date_created"`
	LastUsed    string       `json:"last_used"`
}

// CreateSSHKeyParams contains parameters for adding a new SSH key.
type CreateSSHKeyParams struct {
	PublicKey string `json:"public_key"`
	Label     string `json:"label"`
}

// ---------------------------------------------------------------------------
// Table
// ---------------------------------------------------------------------------

// Table represents a table record in the CMS metadata.
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

// AdminContentData represents a content data entry in the admin namespace.
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

// CreateAdminContentDataParams contains fields for creating a new admin content data record.
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

// UpdateAdminContentDataParams contains fields for updating an existing admin content data record.
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

// AdminContentField represents a content field in the admin namespace.
type AdminContentField struct {
	AdminContentFieldID AdminContentFieldID `json:"admin_content_field_id"`
	AdminRouteID        *AdminRouteID       `json:"admin_route_id"`
	AdminContentDataID  *AdminContentID     `json:"admin_content_data_id"`
	AdminFieldID        *AdminFieldID       `json:"admin_field_id"`
	AdminFieldValue     string              `json:"admin_field_value"`
	AuthorID            *UserID             `json:"author_id"`
	DateCreated         Timestamp           `json:"date_created"`
	DateModified        Timestamp           `json:"date_modified"`
}

// CreateAdminContentFieldParams contains fields for creating a new admin content field.
type CreateAdminContentFieldParams struct {
	AdminRouteID       *AdminRouteID   `json:"admin_route_id"`
	AdminContentDataID *AdminContentID `json:"admin_content_data_id"`
	AdminFieldID       *AdminFieldID   `json:"admin_field_id"`
	AdminFieldValue    string          `json:"admin_field_value"`
	AuthorID           *UserID         `json:"author_id"`
}

// UpdateAdminContentFieldParams contains fields for updating an existing admin content field.
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

// AdminContentRelation represents a relation between admin content items.
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

// AdminDatatype represents an admin datatype.
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

// CreateAdminDatatypeParams contains the parameters for creating an admin datatype.
type CreateAdminDatatypeParams struct {
	ParentID *AdminDatatypeID `json:"parent_id"`
	Name     string           `json:"name"`
	Label    string           `json:"label"`
	Type     string           `json:"type"`
	AuthorID *UserID          `json:"author_id"`
}

// UpdateAdminDatatypeParams contains the parameters for updating an admin datatype.
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

// AdminField represents an admin field definition.
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
	AuthorID     *UserID          `json:"author_id"`
	DateCreated  Timestamp        `json:"date_created"`
	DateModified Timestamp        `json:"date_modified"`
}

// CreateAdminFieldParams contains parameters for creating a new admin field.
type CreateAdminFieldParams struct {
	ParentID   *AdminDatatypeID `json:"parent_id"`
	SortOrder  int64            `json:"sort_order"`
	Name       string           `json:"name"`
	Label      string           `json:"label"`
	Data       string           `json:"data"`
	Validation string           `json:"validation"`
	UIConfig   string           `json:"ui_config"`
	Type       FieldType        `json:"type"`
	AuthorID   *UserID          `json:"author_id"`
}

// UpdateAdminFieldParams contains parameters for updating an existing admin field.
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
	AuthorID     *UserID          `json:"author_id"`
}

// ---------------------------------------------------------------------------
// Field Type
// ---------------------------------------------------------------------------

// FieldTypeInfo represents a field type definition.
type FieldTypeInfo struct {
	FieldTypeID FieldTypeID `json:"field_type_id"`
	Type        string      `json:"type"`
	Label       string      `json:"label"`
}

// CreateFieldTypeParams contains parameters for creating a field type.
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

// AdminFieldTypeInfo represents an admin field type definition.
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

// AdminRoute represents a CMS admin route.
type AdminRoute struct {
	AdminRouteID AdminRouteID `json:"admin_route_id"`
	Slug         Slug         `json:"slug"`
	Title        string       `json:"title"`
	Status       int64        `json:"status"`
	AuthorID     *UserID      `json:"author_id"`
	DateCreated  Timestamp    `json:"date_created"`
	DateModified Timestamp    `json:"date_modified"`
}

// CreateAdminRouteParams contains parameters for creating an admin route.
type CreateAdminRouteParams struct {
	Slug     Slug    `json:"slug"`
	Title    string  `json:"title"`
	Status   int64   `json:"status"`
	AuthorID *UserID `json:"author_id"`
}

// UpdateAdminRouteParams contains parameters for updating an admin route.
type UpdateAdminRouteParams struct {
	Slug     Slug    `json:"slug"`
	Title    string  `json:"title"`
	Status   int64   `json:"status"`
	AuthorID *UserID `json:"author_id"`
	Slug2    Slug    `json:"slug_2"`
}

// ---------------------------------------------------------------------------
// Auth Types
// ---------------------------------------------------------------------------

// LoginParams contains credentials for password-based authentication.
type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the response returned after successful authentication.
type LoginResponse struct {
	UserID      UserID    `json:"user_id"`
	Email       Email     `json:"email"`
	Username    string    `json:"username"`
	DateCreated Timestamp `json:"date_created"`
}

// MeResponse is the response returned by the /me endpoint.
// Note: The User type can be used directly since it matches the /me response shape.

// ---------------------------------------------------------------------------
// Import Types
// ---------------------------------------------------------------------------

// ImportResult represents the result of an import operation.
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

// ChangeEvent represents an audit trail entry for database mutations.
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

// Backup represents a backup record.
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

// ContentVersion represents a snapshot version of content data.
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

// AdminContentVersion represents a snapshot version of admin content data.
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

// PublishRequest is the request body for publishing content.
type PublishRequest struct {
	ContentDataID ContentID `json:"content_data_id"`
}

// AdminPublishRequest is the request body for publishing admin content.
type AdminPublishRequest struct {
	AdminContentDataID AdminContentID `json:"admin_content_data_id"`
}

// PublishResponse is the response from a publish or unpublish operation.
type PublishResponse struct {
	Status           string `json:"status"`
	VersionNumber    int64  `json:"version_number,omitempty"`
	ContentVersionID string `json:"content_version_id,omitempty"`
	ContentDataID    string `json:"content_data_id"`
}

// AdminPublishResponse is the response from an admin publish or unpublish operation.
type AdminPublishResponse struct {
	Status                string `json:"status"`
	VersionNumber         int64  `json:"version_number,omitempty"`
	AdminContentVersionID string `json:"admin_content_version_id,omitempty"`
	AdminContentDataID    string `json:"admin_content_data_id"`
}

// ScheduleRequest is the request body for scheduling content publication.
type ScheduleRequest struct {
	ContentDataID ContentID `json:"content_data_id"`
	PublishAt     string    `json:"publish_at"`
}

// AdminScheduleRequest is the request body for scheduling admin content publication.
type AdminScheduleRequest struct {
	AdminContentDataID AdminContentID `json:"admin_content_data_id"`
	PublishAt          string         `json:"publish_at"`
}

// ScheduleResponse is the response from a schedule operation.
type ScheduleResponse struct {
	Status        string `json:"status"`
	ContentDataID string `json:"content_data_id"`
	PublishAt     string `json:"publish_at"`
}

// AdminScheduleResponse is the response from an admin schedule operation.
type AdminScheduleResponse struct {
	Status             string `json:"status"`
	AdminContentDataID string `json:"admin_content_data_id"`
	PublishAt          string `json:"publish_at"`
}

// CreateVersionRequest is the request body for manually creating a content version.
type CreateVersionRequest struct {
	ContentDataID ContentID `json:"content_data_id"`
	Label         string    `json:"label,omitempty"`
}

// CreateAdminVersionRequest is the request body for manually creating an admin content version.
type CreateAdminVersionRequest struct {
	AdminContentDataID AdminContentID `json:"admin_content_data_id"`
	Label              string         `json:"label,omitempty"`
}

// RestoreRequest is the request body for restoring content to a previous version.
type RestoreRequest struct {
	ContentDataID    ContentID        `json:"content_data_id"`
	ContentVersionID ContentVersionID `json:"content_version_id"`
}

// AdminRestoreRequest is the request body for restoring admin content to a previous version.
type AdminRestoreRequest struct {
	AdminContentDataID    AdminContentID        `json:"admin_content_data_id"`
	AdminContentVersionID AdminContentVersionID `json:"admin_content_version_id"`
}

// RestoreResponse is the response from a restore operation.
type RestoreResponse struct {
	Status          string   `json:"status"`
	ContentDataID   string   `json:"content_data_id"`
	RestoredVersion string   `json:"restored_version_id"`
	FieldsRestored  int      `json:"fields_restored"`
	UnmappedFields  []string `json:"unmapped_fields,omitempty"`
}

// AdminRestoreResponse is the response from an admin restore operation.
type AdminRestoreResponse struct {
	Status             string   `json:"status"`
	AdminContentDataID string   `json:"admin_content_data_id"`
	RestoredVersion    string   `json:"restored_version_id"`
	FieldsRestored     int      `json:"fields_restored"`
	UnmappedFields     []string `json:"unmapped_fields,omitempty"`
}
