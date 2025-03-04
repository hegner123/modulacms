package db

import (
	"context"
	"database/sql"
)

type TestDbDriver interface {
	GetConnection() (*sql.DB, context.Context)
	CreateRole(*sql.DB, context.Context, CreateRoleParams) *Roles
}

type DbDriver interface {
    InitDb(v *bool) error
	Ping() error

	GetConnection() (*sql.DB, context.Context)
	CountAdminDatatypes() (*int64, error)
	CountAdminFields() (*int64, error)
	CountAdminRoutes() (*int64, error)
	CountContentData() (*int64, error)
	CountContentFields() (*int64, error)
	CountDatatypes() (*int64, error)
	CountFields() (*int64, error)
	CountMedia() (*int64, error)
	CountMediaDimensions() (*int64, error)
	CountRoles() (*int64, error)
	CountRoutes() (*int64, error)
	CountTables() (*int64, error)
	CountTokens() (*int64, error)
	CountUsers() (*int64, error)

	CreateAdminDatatype(CreateAdminDatatypeParams) AdminDatatypes
	CreateAdminField(CreateAdminFieldParams) AdminFields
	CreateAdminRoute(CreateAdminRouteParams) AdminRoutes
	CreateContentData(CreateContentDataParams) ContentData
	CreateContentField(CreateContentFieldParams) ContentFields
	CreateDatatype(CreateDatatypeParams) Datatypes
	CreateField(CreateFieldParams) Fields
	CreateMedia(CreateMediaParams) Media
	CreateMediaDimension(CreateMediaDimensionParams) MediaDimensions
	CreateRole(CreateRoleParams) Roles
	CreateRoute(CreateRouteParams) Routes
	CreateTable(string) Tables
	CreateToken(CreateTokenParams) Tokens
	CreateUser(CreateUserParams) Users

    CreateAllTables() error

	CreateAdminDatatypeTable() error
	CreateAdminFieldTable() error
	CreateAdminRouteTable() error
	CreateContentDataTable() error
	CreateContentFieldTable() error
	CreateDatatypeTable() error
	CreateFieldTable() error
	CreateMediaTable() error
	CreateMediaDimensionTable() error
	CreateRoleTable() error
	CreateRouteTable() error
	CreateTableTable() error
	CreateTokenTable() error
	CreateUserTable() error

	DeleteAdminDatatype(int64) error
	DeleteAdminField(int64) error
	DeleteAdminRoute(string) error
	DeleteContentData(int64) error
	DeleteContentField(int64) error
	DeleteDatatype(int64) error
	DeleteField(int64) error
	DeleteMedia(int64) error
	DeleteMediaDimension(int64) error
	DeleteRole(int64) error
	DeleteRoute(string) error
	DeleteTable(int64) error
	DeleteToken(int64) error
	DeleteUser(int64) error

	GetAdminDatatypeById(int64) (*AdminDatatypes, error)
	GetAdminField(int64) (*AdminFields, error)
	GetAdminRoute(string) (*AdminRoutes, error)
	GetContentData(int64) (*ContentData, error)
	GetContentField(int64) (*ContentFields, error)
	GetDatatype(int64) (*Datatypes, error)
	GetField(int64) (*Fields, error)
	GetMedia(int64) (*Media, error)
	GetMediaDimension(int64) (*MediaDimensions, error)
	GetRole(int64) (*Roles, error)
	GetRoute(string) (*Routes, error)
	GetTable(int64) (*Tables, error)
	GetToken(int64) (*Tokens, error)
	GetTokenByUserId(int64) (*[]Tokens, error)
	GetUser(int64) (*Users, error)

	ListAdminDatatypes() (*[]AdminDatatypes, error)
	ListAdminFields() (*[]AdminFields, error)
	ListAdminRoutes() (*[]AdminRoutes, error)
	ListContentData() (*[]ContentData, error)
	ListContentFields() (*[]ContentFields, error)
	ListDatatypes() (*[]Datatypes, error)
	ListFields() (*[]Fields, error)
	ListMedia() (*[]Media, error)
	ListMediaDimensions() (*[]MediaDimensions, error)
	ListRoles() (*[]Roles, error)
	ListRoutes() (*[]Routes, error)
	ListTables() (*[]Tables, error)
	ListTokens() (*[]Tokens, error)
	ListUsers() (*[]Users, error)

	UpdateAdminDatatype(UpdateAdminDatatypeParams) (*string, error)
	UpdateAdminField(UpdateAdminFieldParams) (*string, error)
	UpdateAdminRoute(UpdateAdminRouteParams) (*string, error)
	UpdateContentData(UpdateContentDataParams) (*string, error)
	UpdateContentField(UpdateContentFieldParams) (*string, error)
	UpdateDatatype(UpdateDatatypeParams) (*string, error)
	UpdateField(UpdateFieldParams) (*string, error)
	UpdateMedia(UpdateMediaParams) (*string, error)
	UpdateMediaDimension(UpdateMediaDimensionParams) (*string, error)
	UpdateRole(UpdateRoleParams) (*string, error)
	UpdateRoute(UpdateRouteParams) (*string, error)
	UpdateTable(UpdateTableParams) (*string, error)
	UpdateToken(UpdateTokenParams) (*string, error)
	UpdateUser(UpdateUserParams) (*string, error)
}
type AdminDatatypes struct {
	AdminDtID    int64          `json:"admin_dt_id"`
	AdminRouteID sql.NullInt64  `json:"admin_route_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Author       string         `json:"author"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type AdminFields struct {
	AdminFieldID int64          `json:"admin_field_id"`
	AdminRouteID sql.NullInt64  `json:"admin_route_id"`
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
	AdminDtID     int64          `json:"admin_dt_id"`
	History       sql.NullString `json:"history"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
}

type ContentFields struct {
	ContentFieldID int64          `json:"content_field_id"`
	ContentDataID  int64          `json:"content_data_id"`
	AdminFieldID   int64          `json:"admin_field_id"`
	FieldValue     string         `json:"field_value"`
	History        sql.NullString `json:"history"`
	DateCreated    sql.NullString `json:"date_created"`
	DateModified   sql.NullString `json:"date_modified"`
}

type Datatypes struct {
	DatatypeID   int64          `json:"datatype_id"`
	RouteID      sql.NullInt64  `json:"route_id"`
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
	RouteID      sql.NullInt64  `json:"route_id"`
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
	MediaID            int64          `json:"media_id"`
	Name               sql.NullString `json:"name"`
	DisplayName        sql.NullString `json:"display_name"`
	Alt                sql.NullString `json:"alt"`
	Caption            sql.NullString `json:"caption"`
	Description        sql.NullString `json:"description"`
	Class              sql.NullString `json:"class"`
	Author             any            `json:"author"`
	AuthorID           int64          `json:"author_id"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
	Mimetype           sql.NullString `json:"mimetype"`
	Dimensions         sql.NullString `json:"dimensions"`
	Url                sql.NullString `json:"url"`
	OptimizedMobile    sql.NullString `json:"optimized_mobile"`
	OptimizedTablet    sql.NullString `json:"optimized_tablet"`
	OptimizedDesktop   sql.NullString `json:"optimized_desktop"`
	OptimizedUltraWide sql.NullString `json:"optimized_ultra_wide"`
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
	Author       any            `json:"author"`
	AuthorID     int64          `json:"author_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
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
