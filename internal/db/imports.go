package db

// This file imports all resources

// Resource types

type StringAdminDatatypeFields struct {
	ID              string `json:"id"`
	AdminDatatypeID string `json:"admin_datatype_id"`
	AdminFieldID    string `json:"admin_field_id"`
}
type StringUsers struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type StringRoutes struct {
	RouteID      string `json:"route_id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type StringFields struct {
	FieldID      string `json:"field_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Validation   string `json:"validation"`
	UIConfig     string `json:"ui_config"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type StringMedia struct {
	MediaID      string `json:"media_id"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Alt          string `json:"alt"`
	Caption      string `json:"caption"`
	Description  string `json:"description"`
	Class        string `json:"class"`
	Mimetype     string `json:"mimetype"`
	Dimensions   string `json:"dimensions"`
	Url          string `json:"url"`
	Srcset       string `json:"srcset"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type StringMediaDimensions struct {
	MdID        string `json:"md_id"`
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}

type StringTokens struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
	Revoked   string `json:"revoked"`
}

type StringDatatypes struct {
	DatatypeID   string `json:"datatype_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type StringDatatypeFields struct {
	ID         string `json:"id"`
	DatatypeID string `json:"datatype_id"`
	FieldID    string `json:"field_id"`
	SortOrder  string `json:"sort_order"`
}

type StringSessions struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at"`
	LastAccess  string `json:"last_access"`
	IpAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	SessionData string `json:"session_data"`
}

type StringRoles struct {
	RoleID      string `json:"role_id"`
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}

type StringPermissions struct {
	PermissionID string `json:"permission_id"`
	TableID      string `json:"table_id"`
	Mode         string `json:"mode"`
	Label        string `json:"label"`
}

type StringContentData struct {
	ContentDataID string `json:"content_data_id"`
	RouteID       string `json:"route_id"`
	ParentID      string `json:"parent_id"`
	FirstChildID  string `json:"first_child_id"`
	NextSiblingID string `json:"next_sibling_id"`
	PrevSiblingID string `json:"prev_sibling_id"`
	DatatypeID    string `json:"datatype_id"`
	AuthorID      string `json:"author_id"`
	Status        string `json:"status"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
	History       string `json:"history"`
}

type StringContentFields struct {
	ContentFieldID string `json:"content_field_id"`
	RouteID        string `json:"route_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	FieldValue     string `json:"field_value"`
	AuthorID       string `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
	History        string `json:"history"`
}

type StringAdminRoutes struct {
	AdminRouteID string `json:"admin_route_id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type StringAdminFields struct {
	AdminFieldID string `json:"admin_field_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Validation   string `json:"validation"`
	UIConfig     string `json:"ui_config"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type StringAdminDatatypes struct {
	AdminDatatypeID string `json:"admin_datatype_id"`
	ParentID        string `json:"parent_id"`
	Label           string `json:"label"`
	Type            string `json:"type"`
	AuthorID        string `json:"author_id"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
	History         string `json:"history"`
}

type StringAdminContentData struct {
	AdminContentDataID string `json:"admin_content_data_id"`
	ParentID           string `json:"parent_id"`
	AdminRouteID       string `json:"admin_route_id"`
	AdminDatatypeID    string `json:"admin_datatype_id"`
	AuthorID           string `json:"author_id"`
	Status             string `json:"status"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
	History            string `json:"history"`
}

type StringAdminContentFields struct {
	AdminContentFieldID string `json:"admin_content_field_id"`
	AdminRouteID        string `json:"admin_route_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	AuthorID            string `json:"author_id"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
	History             string `json:"history"`
}

type StringTables struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	AuthorID string `json:"author_id"`
}

type StringUserOauth struct {
	UserOauthID         string `json:"user_oauth_id"`
	UserID              string `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}
