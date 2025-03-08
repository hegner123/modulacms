package db

import (
	"context"
	"database/sql"
)

type DbDriver interface {
	CreateAllTables() error
	InitDb(v *bool) error
	Ping() error
	GetConnection() (*sql.DB, context.Context, error)

	CountAdminContentData() (*int64, error)
	CountAdminContentFields() (*int64, error)
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

	CreateAdminContentData(CreateAdminContentDataParams) AdminContentData
	CreateAdminContentField(CreateAdminContentFieldParams) AdminContentFields
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
	CreateUser(CreateUserParams) (*Users, error)

	CreateAdminContentDataTable() error
	CreateAdminContentFieldTable() error
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

	DeleteAdminContentData(int64) error
	DeleteAdminContentField(int64) error
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

	GetAdminContentData(int64) (*AdminContentData, error)
	GetAdminContentField(int64) (*AdminContentFields, error)
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

	ListAdminContentData() (*[]AdminContentData, error)
	ListAdminContentFields() (*[]AdminContentFields, error)
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

	UpdateAdminContentData(UpdateAdminContentDataParams) (*string, error)
	UpdateAdminContentField(UpdateAdminContentFieldParams) (*string, error)
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
