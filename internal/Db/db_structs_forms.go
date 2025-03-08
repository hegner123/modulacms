package db

type CreateAdminContentDataFormParams struct {
	AdminRouteID    string `json:"admin_route_id"`
	AdminDatatypeID string `json:"admin_datatype_id"`
	History         string `json:"history"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
}
type CreateAdminContentFieldFormParams struct {
	AdminRouteID        string `json:"admin_route_id"`
	AdminContentFieldID string `json:"admin_content_field_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	History             string `json:"history"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
}
type CreateAdminDatatypeFormParams struct {
	AdminRouteID string `json:"admin_route_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}
type CreateAdminFieldFormParams struct {
	AdminRouteID string `json:"admin_route_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}
type CreateAdminRouteFormParams struct {
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}
type CreateContentDataFormParams struct {
	RouteID      string `json:"route_id"`
	DatatypeID   string `json:"datatype_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}
type CreateContentFieldFormParams struct {
	RouteID        string `json:"route_id"`
	ContentFieldID string `json:"content_field_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	FieldValue     string `json:"field_value"`
	History        string `json:"history"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
}
type CreateDatatypeFormParams struct {
	RouteID      string `json:"route_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}
type CreateFieldFormParams struct {
	RouteID      string `json:"route_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}
type CreateMediaFormParams struct {
	Name               string `json:"name"`
	DisplayName        string `json:"display_name"`
	Alt                string `json:"alt"`
	Caption            string `json:"caption"`
	Description        string `json:"description"`
	Class              string `json:"class"`
	Author             string `json:"author"`
	AuthorID           string `json:"author_id"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
	Url                string `json:"url"`
	Mimetype           string `json:"mimetype"`
	Dimensions         string `json:"dimensions"`
	OptimizedMobile    string `json:"optimized_mobile"`
	OptimizedTablet    string `json:"optimized_tablet"`
	OptimizedDesktop   string `json:"optimized_desktop"`
	OptimizedUltraWide string `json:"optimized_ultra_wide"`
}
type CreateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}
type CreateRoleFormParams struct {
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}
type CreateRouteFormParams struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	History      string `json:"history"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}
type CreateTokenFormParams struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
	Revoked   string `json:"revoked"`
}
type CreateUserFormParams struct {
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
}
type UpdateAdminContentDataFormParams struct {
	AdminRouteID       string `json:"admin_route_id"`
	AdminDatatypeID    string `json:"admin_datatype_id"`
	History            string `json:"history"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
	AdminContentDataID string `json:"admin_content_data_id"`
}
type UpdateAdminContentFieldFormParams struct {
	AdminRouteID          string `json:"admin_route_id"`
	AdminContentFieldID   string `json:"content_field_id"`
	AdminContentDataID    string `json:"content_data_id"`
	AdminFieldID          string `json:"admin_field_id"`
	AdminFieldValue       string `json:"admin_field_value"`
	History               string `json:"history"`
	DateCreated           string `json:"date_created"`
	DateModified          string `json:"date_modified"`
	AdminContentFieldID_2 string `json:"admin_content_field_id_2"`
}
type UpdateAdminDatatypeFormParams struct {
	AdminRouteID    string `json:"admin_route_id"`
	ParentID        string `json:"parent_id"`
	Label           string `json:"label"`
	Type            string `json:"type"`
	Author          string `json:"author"`
	AuthorID        string `json:"author_id"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
	History         string `json:"history"`
	AdminDatatypeID string `json:"admin_datatype_id"`
}
type UpdateAdminFieldFormParams struct {
	AdminRouteID string `json:"admin_route_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
	AdminFieldID string `json:"admin_field_id"`
}
type UpdateAdminRouteFormParams struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
	Slug_2       string `json:"slug_2"`
}
type UpdateContentDataFormParams struct {
	RouteID       string `json:"route_id"`
	DatatypeID    string `json:"admin_datatype_id"`
	History       string `json:"history"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
	ContentDataID string `json:"content_data_id"`
}
type UpdateContentFieldFormParams struct {
	RouteID          string `json:"route_id"`
	ContentFieldID   string `json:"content_field_id"`
	ContentDataID    string `json:"content_data_id"`
	FieldID          string `json:"field_id"`
	FieldValue       string `json:"field_value"`
	History          string `json:"history"`
	DateCreated      string `json:"date_created"`
	DateModified     string `json:"date_modified"`
	ContentFieldID_2 string `json:"content_field_id_2"`
}
type UpdateDatatypeFormParams struct {
	RouteID      string `json:"route_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	DatatypeID   string `json:"datatype_id"`
}
type UpdateFieldFormParams struct {
	RouteID      string `json:"route_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	FieldID      string `json:"field_id"`
}
type UpdateMediaFormParams struct {
	Name               string `json:"name"`
	DisplayName        string `json:"display_name"`
	Alt                string `json:"alt"`
	Caption            string `json:"caption"`
	Description        string `json:"description"`
	Class              string `json:"class"`
	Author             string `json:"author"`
	AuthorID           string `json:"author_id"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
	Url                string `json:"url"`
	Mimetype           string `json:"mimetype"`
	Dimensions         string `json:"dimensions"`
	OptimizedMobile    string `json:"optimized_mobile"`
	OptimizedTablet    string `json:"optimized_tablet"`
	OptimizedDesktop   string `json:"optimized_desktop"`
	OptimizedUltraWide string `json:"optimized_ultra_wide"`
	MediaID            string `json:"media_id"`
}
type UpdateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
	MdID        string `json:"md_id"`
}
type UpdateRoleFormParams struct {
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
	RoleID      string `json:"role_id"`
}
type UpdateRouteFormParams struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	History      string `json:"history"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	Slug_2       string `json:"slug_2"`
}
type UpdateTableFormParams struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}
type UpdateTokenFormParams struct {
	Token     string `json:"token"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
	Revoked   string `json:"revoked"`
	ID        string `json:"id"`
}
type UpdateUserFormParams struct {
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
	UserID       string `json:"user_id"`
}
