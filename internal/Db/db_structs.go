package db

import (
	"context"
	"database/sql"

	config "github.com/hegner123/modulacms/internal/Config"
)

type DbStatus string

const (
	Open   DbStatus = "open"
	Closed DbStatus = "closed"
	Err    DbStatus = "error"
)

type Database struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

type MysqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}
type PsqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

type AdminContentData struct {
	AdminContentDataID int64          `json:"admin_content_data_id"`
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	History            sql.NullString `json:"history"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
}

type AdminContentFields struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        int64          `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	History             sql.NullString `json:"history"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
}

type AdminDatatypes struct {
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	Author          string         `json:"author"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
	History         sql.NullString `json:"history"`
}

type AdminFields struct {
	AdminFieldID int64          `json:"admin_field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         any            `json:"data"`
	Type         any            `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type AdminRoutes struct {
	AdminRouteID int64          `json:"admin_route_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type ContentData struct {
	ContentDataID int64          `json:"content_data_id"`
	RouteID       int64          `json:"route_id"`
	DatatypeID    int64          `json:"datatype_id"`
	History       sql.NullString `json:"history"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
}

type ContentFields struct {
	ContentFieldID int64          `json:"content_field_id"`
	RouteID        int64          `json:"route_id"`
	ContentDataID  int64          `json:"content_data_id"`
	FieldID        int64          `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	History        sql.NullString `json:"history"`
	DateCreated    sql.NullString `json:"date_created"`
	DateModified   sql.NullString `json:"date_modified"`
}

type Datatypes struct {
	DatatypeID   int64          `json:"datatype_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type Fields struct {
	FieldID      int64          `json:"field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type Media struct {
	MediaID      int64          `json:"media_id"`
	Name         sql.NullString `json:"name"`
	DisplayName  sql.NullString `json:"display_name"`
	Alt          sql.NullString `json:"alt"`
	Caption      sql.NullString `json:"caption"`
	Description  sql.NullString `json:"description"`
	Class        sql.NullString `json:"class"`
	Mimetype     sql.NullString `json:"mimetype"`
	Dimensions   sql.NullString `json:"dimensions"`
	Url          sql.NullString `json:"url"`
	Srcset       sql.NullString `json:"srcset"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type MediaDimensions struct {
	MdID        int64          `json:"md_id"`
	Label       sql.NullString `json:"label"`
	Width       sql.NullInt64  `json:"width"`
	Height      sql.NullInt64  `json:"height"`
	AspectRatio sql.NullString `json:"aspect_ratio"`
}

type Roles struct {
	RoleID      int64  `json:"role_id"`
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}

type Routes struct {
	RouteID      int64          `json:"route_id"`
	Author       string         `json:"author"`
	AuthorID     int64          `json:"author_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type Sessions struct {
	SessionID   int64          `json:"session_id"`
	UserID      int64          `json:"user_id"`
	CreatedAt   sql.NullString `json:"created_at"`
	ExpiresAt   sql.NullString `json:"expires_at"`
	LastAccess  sql.NullString `json:"last_access"`
	IpAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	SessionData sql.NullString `json:"session_data"`
}

type Tables struct {
	ID       int64          `json:"id"`
	Label    sql.NullString `json:"label"`
	AuthorID int64          `json:"author_id"`
}

type Tokens struct {
	ID        int64        `json:"id"`
	UserID    int64        `json:"user_id"`
	TokenType string       `json:"token_type"`
	Token     string       `json:"token"`
	IssuedAt  string       `json:"issued_at"`
	ExpiresAt string       `json:"expires_at"`
	Revoked   sql.NullBool `json:"revoked"`
}

type Users struct {
	UserID       int64          `json:"user_id"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	References   any            `json:"references"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type UserOauth struct {
	UserOauthID         int64          `json:"user_oauth_id"`
	UserID              int64          `json:"user_id"`
	OauthProvider       string         `json:"oauth_provider"`
	OauthProviderUserID string         `json:"oauth_provider_user_id"`
	AccessToken         sql.NullString `json:"access_token"`
	RefreshToken        sql.NullString `json:"refresh_token"`
	TokenExpiresAt      sql.NullString `json:"token_expires_at"`
	DateCreated         sql.NullString `json:"date_created"`
}

type CreateAdminContentDataParams struct {
	AdminRouteID    int64          `json:"admin_route_id"`
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	History         sql.NullString `json:"history"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
}

type CreateAdminContentFieldParams struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        int64          `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	History             sql.NullString `json:"history"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
}

type CreateAdminDatatypeParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Author       string         `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type CreateAdminFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         any            `json:"data"`
	Type         any            `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type CreateAdminRouteParams struct {
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type CreateContentDataParams struct {
	RouteID      int64          `json:"route_id"`
	DatatypeID   int64          `json:"datatype_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateContentFieldParams struct {
	ContentFieldID int64          `json:"content_field_id"`
	RouteID        int64          `json:"route_id"`
	ContentDataID  int64          `json:"content_data_id"`
	FieldID        int64          `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	History        sql.NullString `json:"history"`
	DateCreated    sql.NullString `json:"date_created"`
	DateModified   sql.NullString `json:"date_modified"`
}
type CreateDatatypeParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateMediaParams struct {
	Name         sql.NullString `json:"name"`
	DisplayName  sql.NullString `json:"display_name"`
	Alt          sql.NullString `json:"alt"`
	Caption      sql.NullString `json:"caption"`
	Description  sql.NullString `json:"description"`
	Class        sql.NullString `json:"class"`
	Url          sql.NullString `json:"url"`
	Mimetype     sql.NullString `json:"mimetype"`
	Dimensions   sql.NullString `json:"dimensions"`
	Srcset       sql.NullString `json:"srcset"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateMediaDimensionParams struct {
	Label       sql.NullString `json:"label"`
	Width       sql.NullInt64  `json:"width"`
	Height      sql.NullInt64  `json:"height"`
	AspectRatio sql.NullString `json:"aspect_ratio"`
}
type CreateRoleParams struct {
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}
type CreateRouteParams struct {
	Author       string         `json:"author"`
	AuthorID     int64          `json:"author_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}
type CreateSessionParams struct {
	UserID      int64          `json:"user_id"`
	CreatedAt   sql.NullString `json:"created_at"`
	ExpiresAt   sql.NullString `json:"expires_at"`
	LastAccess  sql.NullString `json:"last_access"`
	IpAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	SessionData sql.NullString `json:"session_data"`
}

type CreateTokenParams struct {
	UserID    int64        `json:"user_id"`
	TokenType string       `json:"token_type"`
	Token     string       `json:"token"`
	IssuedAt  string       `json:"issued_at"`
	ExpiresAt string       `json:"expires_at"`
	Revoked   sql.NullBool `json:"revoked"`
}
type CreateUserParams struct {
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
}
type CreateUserOauthParams struct {
	UserID              int64          `json:"user_id"`
	OauthProvider       string         `json:"oauth_provider"`
	OauthProviderUserID string         `json:"oauth_provider_user_id"`
	AccessToken         sql.NullString `json:"access_token"`
	RefreshToken        sql.NullString `json:"refresh_token"`
	TokenExpiresAt      sql.NullString `json:"token_expires_at"`
	DateCreated         sql.NullString `json:"date_created"`
}

type ListAdminDatatypeByRouteIdRow struct {
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	AdminRouteID    sql.NullInt64  `json:"admin_route_id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	History         sql.NullString `json:"history"`
}
type ListAdminDatatypeTreeRow struct {
	ChildID     int64          `json:"child_id"`
	ChildLabel  string         `json:"child_label"`
	ParentID    sql.NullInt64  `json:"parent_id"`
	ParentLabel sql.NullString `json:"parent_label"`
}
type ListAdminFieldByRouteIdRow struct {
	AdminFieldID int64          `json:"admin_field_id"`
	AdminRouteID sql.NullInt64  `json:"admin_route_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         any            `json:"data"`
	Type         any            `json:"type"`
	History      sql.NullString `json:"history"`
}
type ListAdminFieldsByDatatypeIDRow struct {
	AdminFieldID int64          `json:"admin_field_id"`
	AdminRouteID sql.NullInt64  `json:"admin_route_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         any            `json:"data"`
	Type         any            `json:"type"`
	History      sql.NullString `json:"history"`
}
type ListDatatypeByRouteIdRow struct {
	DatatypeID int64         `json:"datatype_id"`
	RouteID    sql.NullInt64 `json:"route_id"`
	ParentID   sql.NullInt64 `json:"parent_id"`
	Label      string        `json:"label"`
	Type       string        `json:"type"`
}
type ListFieldByRouteIdRow struct {
	FieldID  int64         `json:"field_id"`
	RouteID  sql.NullInt64 `json:"route_id"`
	ParentID sql.NullInt64 `json:"parent_id"`
	Label    any           `json:"label"`
	Data     string        `json:"data"`
	Type     string        `json:"type"`
}

type UpdateAdminContentDataParams struct {
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	History            sql.NullString `json:"history"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
	AdminContentDataID int64          `json:"admin_content_data_id"`
}

type UpdateAdminContentFieldParams struct {
	AdminContentFieldID   int64          `json:"admin_content_field_id"`
	AdminRouteID          int64          `json:"admin_route_id"`
	AdminContentDataID    int64          `json:"admin_content_data_id"`
	AdminFieldID          int64          `json:"admin_field_id"`
	AdminFieldValue       string         `json:"admin_field_value"`
	History               sql.NullString `json:"history"`
	DateCreated           sql.NullString `json:"date_created"`
	DateModified          sql.NullString `json:"date_modified"`
	AdminContentFieldID_2 int64          `json:"admin_content_field_id_2"`
}

type UpdateAdminDatatypeParams struct {
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	Author          string         `json:"author"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
	History         sql.NullString `json:"history"`
	AdminDatatypeID int64          `json:"admin_datatype_id"`
}

type UpdateAdminFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         any            `json:"data"`
	Type         any            `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
	AdminFieldID int64          `json:"admin_field_id"`
}

type UpdateAdminRouteParams struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
	Slug_2       string         `json:"slug_2"`
}

type UpdateContentDataParams struct {
	RouteID       int64          `json:"route_id"`
	DatatypeID    int64          `json:"datatype_id"`
	History       sql.NullString `json:"history"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
	ContentDataID int64          `json:"content_data_id"`
}

type UpdateContentFieldParams struct {
	ContentFieldID   int64          `json:"content_field_id"`
	RouteID          int64          `json:"route_id"`
	ContentDataID    int64          `json:"content_data_id"`
	FieldID          int64          `json:"field_id"`
	FieldValue       string         `json:"field_value"`
	History          sql.NullString `json:"history"`
	DateCreated      sql.NullString `json:"date_created"`
	DateModified     sql.NullString `json:"date_modified"`
	ContentFieldID_2 int64          `json:"content_field_id_2"`
}

type UpdateDatatypeParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	DatatypeID   int64          `json:"datatype_id"`
}

type UpdateFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	FieldID      int64          `json:"field_id"`
}

type UpdateMediaParams struct {
	Name         sql.NullString `json:"name"`
	DisplayName  sql.NullString `json:"display_name"`
	Alt          sql.NullString `json:"alt"`
	Caption      sql.NullString `json:"caption"`
	Description  sql.NullString `json:"description"`
	Class        sql.NullString `json:"class"`
	Url          sql.NullString `json:"url"`
	Mimetype     sql.NullString `json:"mimetype"`
	Dimensions   sql.NullString `json:"dimensions"`
	Srcset       sql.NullString `json:"srcset"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	MediaID      int64          `json:"media_id"`
}

type UpdateMediaDimensionParams struct {
	Label       sql.NullString `json:"label"`
	Width       sql.NullInt64  `json:"width"`
	Height      sql.NullInt64  `json:"height"`
	AspectRatio sql.NullString `json:"aspect_ratio"`
	MdID        int64          `json:"md_id"`
}

type UpdateRoleParams struct {
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
	RoleID      int64  `json:"role_id"`
}

type UpdateRouteParams struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	History      sql.NullString `json:"history"`
	Author       string         `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	Slug_2       string         `json:"slug_2"`
}
type UpdateSessionParams struct {
	UserID      int64          `json:"user_id"`
	CreatedAt   sql.NullString `json:"created_at"`
	ExpiresAt   sql.NullString `json:"expires_at"`
	LastAccess  sql.NullString `json:"last_access"`
	IpAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	SessionData sql.NullString `json:"session_data"`
	SessionID   string         `json:"session_id"`
}

type UpdateTableParams struct {
	Label sql.NullString `json:"label"`
	ID    int64          `json:"id"`
}

type UpdateTokenParams struct {
	Token     string       `json:"token"`
	IssuedAt  string       `json:"issued_at"`
	ExpiresAt string       `json:"expires_at"`
	Revoked   sql.NullBool `json:"revoked"`
	ID        int64        `json:"id"`
}

type UpdateUserParams struct {
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	UserID       int64          `json:"user_id"`
}

type UpdateUserOauthParams struct {
	AccessToken    sql.NullString `json:"access_token"`
	RefreshToken   sql.NullString `json:"refresh_token"`
	TokenExpiresAt sql.NullString `json:"token_expires_at"`
	UserOauthID    int64          `json:"user_oauth_id"`
}

type UtilityGetAdminDatatypesRow struct {
	AdminDatatypeID int64  `json:"admin_datatype_id"`
	Label           string `json:"label"`
}
type UtilityGetAdminRoutesRow struct {
	AdminRouteID int64  `json:"admin_route_id"`
	Slug         string `json:"slug"`
}

type UtilityGetAdminfieldsRow struct {
	AdminFieldID int64 `json:"admin_field_id"`
	Label        any   `json:"label"`
}

type UtilityGetDatatypesRow struct {
	DatatypeID int64  `json:"datatype_id"`
	Label      string `json:"label"`
}
type UtilityGetFieldsRow struct {
	FieldID int64 `json:"field_id"`
	Label   any   `json:"label"`
}

type UtilityGetMediaRow struct {
	MediaID int64          `json:"media_id"`
	Name    sql.NullString `json:"name"`
}

type UtilityGetMediaDimensionRow struct {
	MdID  int64          `json:"md_id"`
	Label sql.NullString `json:"label"`
}

type UtilityGetRouteRow struct {
	RouteID any    `json:"route_id"`
	Slug    string `json:"slug"`
}

type UtilityGetTablesRow struct {
	ID    int64          `json:"id"`
	Label sql.NullString `json:"label"`
}

type UtilityGetTokenRow struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}

type UtilityGetUsersRow struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

type UtilityRecordCountRow struct {
	TableName string `json:"table_name"`
	RowCount  int64  `json:"row_count"`
}
type AdminContentDataHistoryEntry struct {
	AdminContentDataID int64          `json:"admin_content_data_id"`
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
}

type AdminContentFieldsHistoryEntry struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        int64          `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
}

type AdminDatatypesHistoryEntry struct {
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	Author          string         `json:"author"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
}

type AdminFieldsHistoryEntry struct {
	AdminFieldID int64          `json:"admin_field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         any            `json:"data"`
	Type         any            `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type AdminRoutesHistoryEntry struct {
	AdminRouteID int64          `json:"admin_route_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type ContentDataHistoryEntry struct {
	ContentDataID int64          `json:"content_data_id"`
	RouteID       int64          `json:"route_id"`
	DatatypeID    int64          `json:"datatype_id"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
}

type ContentFieldsHistoryEntry struct {
	ContentFieldID int64          `json:"content_field_id"`
	RouteID        int64          `json:"route_id"`
	ContentDataID  int64          `json:"content_data_id"`
	FieldID        int64          `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	DateCreated    sql.NullString `json:"date_created"`
	DateModified   sql.NullString `json:"date_modified"`
}

type DatatypesHistoryEntry struct {
	DatatypeID   int64          `json:"datatype_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type FieldsHistoryEntry struct {
	FieldID      int64          `json:"field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type RoutesHistoryEntry struct {
	RouteID      any            `json:"route_id"`
	Author       string         `json:"author"`
	AuthorID     int64          `json:"author_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}
