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
	CountAdminDatatype(...any) int
	CountAdminFields() int
	CountAdminRoutes() int
	CountContentData() int
	CountContentField() int
	CountDatatype() int
	CountField() int
	CountMedia() int
	CountMediaDimension() int
	CountRole() int
	CountRoute() int
	CountTable() int
	CountToken() int
	CountUser() int
	CreateAdminDatatype() AdminDatatypes
	CreateAdminFields() AdminFields
	CreateAdminRoutes() AdminRoutes
	CreateContentData() ContentData
	CreateContentField() ContentFields
	CreateDatatype() Datatypes
	CreateField() Fields
	CreateMedia() Media
	CreateMediaDimension() MediaDimensions
	CreateRole() Roles
	CreateRoute() Routes
	CreateTable() Tables
	CreateToken() Tokens
	CreateUser() Users
	DeleteAdminDatatype() AdminDatatypes
	DeleteAdminFields() AdminFields
	DeleteAdminRoutes() AdminRoutes
	DeleteContentData() ContentData
	DeleteContentField() ContentFields
	DeleteDatatype() Datatypes
	DeleteField() Fields
	DeleteMedia() Media
	DeleteMediaDimension() MediaDimensions
	DeleteRole() Roles
	DeleteRoute() Routes
	DeleteTable() Tables
	DeleteToken() Tokens
	DeleteUser() Users
	GetAdminDatatype() AdminDatatypes
	GetAdminFields() AdminFields
	GetAdminRoutes() AdminRoutes
	GetContentData() ContentData
	GetContentField() ContentFields
	GetDatatype() Datatypes
	GetField() Fields
	GetMedia() Media
	GetMediaDimension() MediaDimensions
	GetRole() Roles
	GetRoute() Routes
	GetTable() Tables
	GetToken() Tokens
	GetUser() Users
	ListAdminDatatypes() []AdminDatatypes
	ListAdminFields() []AdminFields
	ListAdminRoutes() []AdminRoutes
	ListContentDatas() []ContentData
	ListContentFields() []ContentFields
	ListDatatypes() []Datatypes
	ListFields() []Fields
	ListMedia() []Media
	ListMediaDimensions() []MediaDimensions
	ListRoles() []Roles
	ListRoutes() []Routes
	ListTables() []Tables
	ListTokens() []Tokens
	ListUsers() []Users
	UpdateAdminDatatype() AdminDatatypes
	UpdateAdminFields() AdminFields
	UpdateAdminRoutes() AdminRoutes
	UpdateContentData() ContentData
	UpdateContentField() ContentFields
	UpdateDatatype() Datatypes
	UpdateField() Fields
	UpdateMedia() Media
	UpdateMediaDimension() MediaDimensions
	UpdateRole() Roles
	UpdateRoute() Routes
	UpdateTable() Tables
	UpdateToken() Tokens
	UpdateUser() Users
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
