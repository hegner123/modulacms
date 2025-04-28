package db

import (
	"context"
	"database/sql"

	"github.com/hegner123/modulacms/internal/config"
)

type Transaction struct {
	sql.Tx
}

type AdminContentDataRepository interface {
	Count(ctx context.Context) (*int64, error)
	CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateAdminContentDataParams) (*AdminContentData, error)
	GetByID(ctx context.Context, id int64) (*AdminContentData, error)
	List(ctx context.Context) ([]*AdminContentData, error)
	ListByRoute(ctx context.Context, id int64) ([]*AdminContentData, error)
	Update(ctx context.Context, params UpdateAdminContentDataParams) (*AdminContentData, error)
	Delete(ctx context.Context, id int64) error
}

type AdminContentFieldRepository interface {
	Count(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateAdminContentFieldParams) AdminContentFields
	CreateTable(ctx context.Context) error
	Delete(ctx context.Context, id int64) error
	GetById(ctx context.Context, id int64) (*AdminContentFields, error)
	List(ctx context.Context) (*[]AdminContentFields, error)
	ListsByRoute(ctx context.Context, id int64) (*[]AdminContentFields, error)
	Update(ctx context.Context, params UpdateAdminContentFieldParams) (*string, error)
}

type AdminDatatypeFieldRepository interface {
	Count(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateAdminDatatypeFieldParams) (*AdminDatatypeFields, error)
	GetByID(ctx context.Context, id int64) (*AdminDatatypeFields, error)
	List(ctx context.Context) ([]*AdminDatatypeFields, error)
	ListByParentID(ctx context.Context, id int64) ([]*AdminDatatypeFields, error)
	Update(ctx context.Context, params UpdateAdminDatatypeFieldParams) (*AdminDatatypeFields, error)
	Delete(ctx context.Context, id int64) error
}

type AdminDatatypeRepository interface {
	Count(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateAdminDatatypeParams) (*AdminDatatypes, error)
	GetByID(ctx context.Context, id int64) (*AdminDatatypes, error)
	List(ctx context.Context) ([]*AdminDatatypes, error)
	ListByParentID(ctx context.Context, id int64) ([]*AdminDatatypes, error)
	Update(ctx context.Context, params UpdateAdminDatatypeParams) (*AdminDatatypes, error)
	Delete(ctx context.Context, id int64) error
}

type AdminFieldRepository interface {
	Count(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateAdminFieldParams) (*AdminFields, error)
	GetByID(ctx context.Context, id int64) (*AdminFields, error)
	List(ctx context.Context) ([]*AdminFields, error)
	Update(ctx context.Context, params UpdateAdminFieldParams) (*AdminFields, error)
	Delete(ctx context.Context, id int64) error
}

type AdminRoutesRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateAdminRouteParams) (*AdminRoutes, error)
	GetByID(ctx context.Context, id int64) (*AdminRoutes, error)
	GetBySlug(ctx context.Context, slug string) (*AdminRoutes, error)
	List(ctx context.Context) ([]*AdminRoutes, error)
	Update(ctx context.Context, params UpdateAdminRouteParams) (*AdminRoutes, error)
	Delete(ctx context.Context, id int64) error
}

type ContentDataRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateContentDataParams) (*ContentData, error)
	GetByID(ctx context.Context, id int64) (*ContentData, error)
	List(ctx context.Context) ([]*ContentData, error)
	ListByRouteID(ctx context.Context, id int64) ([]*ContentData, error)
	Update(ctx context.Context, params UpdateContentDataParams) (*ContentData, error)
	Delete(ctx context.Context, id int64) error
}

type ContentFieldRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateContentFieldParams) (*ContentFields, error)
	GetByID(ctx context.Context, id int64) (*ContentFields, error)
	List(ctx context.Context) ([]*ContentFields, error)
	ListByContentDataID(ctx context.Context, id int64) ([]*ContentFields, error)
	Update(ctx context.Context, params UpdateContentFieldParams) (*ContentFields, error)
	Delete(ctx context.Context, id int64) error
}

type DatatypeFieldsRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateDatatypeFieldParams) (*DatatypeFields, error)
	GetByID(ctx context.Context, id int64) (*DatatypeFields, error)
	List(ctx context.Context) ([]*DatatypeFields, error)
	Update(ctx context.Context, params UpdateDatatypeFieldParams) (*DatatypeFields, error)
	Delete(ctx context.Context, id int64) error
}
type DatatypeRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateDatatypeParams) (*Datatypes, error)
	GetByID(ctx context.Context, id int64) (*Datatypes, error)
	List(ctx context.Context) ([]*Datatypes, error)
	ListByParentID(ctx context.Context, id int64) ([]*Datatypes, error)
	Update(ctx context.Context, params UpdateDatatypeParams) (*Datatypes, error)
	Delete(ctx context.Context, id int64) error
}

type FieldRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateFieldParams) (*Fields, error)
	GetByID(ctx context.Context, id int64) (*Fields, error)
	List(ctx context.Context) ([]*Fields, error)
	Update(ctx context.Context, params UpdateFieldParams) (*Fields, error)
	Delete(ctx context.Context, id int64) error
}

type MediaRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateMediaParams) (*Media, error)
	GetByID(ctx context.Context, id int64) (*Media, error)
	List(ctx context.Context) ([]*Media, error)
	Update(ctx context.Context, params UpdateMediaParams) (*Media, error)
	Delete(ctx context.Context, id int64) error
}

type MediaDimensionRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateMediaDimensionParams) (*MediaDimensions, error)
	GetByID(ctx context.Context, id int64) (*MediaDimensions, error)
	List(ctx context.Context) ([]*MediaDimensions, error)
	Update(ctx context.Context, params UpdateMediaDimensionParams) (*MediaDimensions, error)
	Delete(ctx context.Context, id int64) error
}

type PermissionRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreatePermissionParams) (*Permissions, error)
	GetByID(ctx context.Context, id int64) (*Permissions, error)
	List(ctx context.Context) ([]*Permissions, error)
	Update(ctx context.Context, params UpdatePermissionParams) (*Permissions, error)
	Delete(ctx context.Context, id int64) error
}

type RoleRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateRoleParams) (*Roles, error)
	GetByID(ctx context.Context, id int64) (*Roles, error)
	List(ctx context.Context) ([]*Roles, error)
	Update(ctx context.Context, params UpdateRoleParams) (*Roles, error)
	Delete(ctx context.Context, id int64) error
}

type RouteRepository interface {
	Count(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateRouteParams) (*Routes, error)
	CreateTable(ctx context.Context) error
	GetByID(ctx context.Context, id int64) (*Routes, error)
	GetBySlug(ctx context.Context, slug string) (*Routes, error)
	List(ctx context.Context) ([]*Routes, error)
	Update(ctx context.Context, params UpdateRouteParams) (*Routes, error)
	Delete(ctx context.Context, id int64) error
}

type SessionRepository interface {
	Count(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateSessionParams) (*Sessions, error)
	CreateTable(ctx context.Context) error
	GetByID(ctx context.Context, id int64) (*Sessions, error)
	GetByUserID(ctx context.Context, id int64) (*Sessions, error)
	List(ctx context.Context) ([]*Sessions, error)
	Update(ctx context.Context, params UpdateSessionParams) (*Sessions, error)
	Delete(ctx context.Context, id int64) error
}

type TablesRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateTableParams) (*Tables, error)
	GetByID(ctx context.Context, id int64) (*Tables, error)
	List(ctx context.Context) ([]*Tables, error)
	Update(ctx context.Context, params UpdateTableParams) (*Tables, error)
	Delete(ctx context.Context, id int64) error
}

type TokenRepository interface {
	Count(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateTokenParams) (*Tokens, error)
	CreateTable(ctx context.Context) error
	GetByID(ctx context.Context, id int64) (*Tokens, error)
	List(ctx context.Context) ([]*Tokens, error)
	Update(ctx context.Context, params UpdateTokenParams) (*Tokens, error)
	Delete(ctx context.Context, id int64) error
}

type UserOauthRepository interface {
	Count(ctx context.Context) (*int64, error)
    CreateTable(ctx context.Context) error
	Create(ctx context.Context, params CreateUserOauthParams) (*UserOauth, error)
	GetByID(ctx context.Context, id int64) (*UserOauth, error)
	List(ctx context.Context) ([]*UserOauth, error)
	Update(ctx context.Context, params UpdateUserOauthParams) (*UserOauth, error)
	Delete(ctx context.Context, id int64) error
}

type UserRepository interface {
	Counte(ctx context.Context) (*int64, error)
	Create(ctx context.Context, params CreateUserParams) (*Users, error)
	CreateTable(ctx context.Context) error
	GetByID(ctx context.Context, id int64) (*Users, error)
	GetByEmail(ctx context.Context, email string) (*Users, error)
	List(ctx context.Context) ([]*Users, error)
	Update(ctx context.Context, params UpdateUserParams) error
	Delete(ctx context.Context, id int64) error
}

type DatabaseProvider interface {
	// Core connection methods
	Connect(ctx context.Context) error
	Close() error
	Ping() error
	CreateAllTables() error
	InitDB(v *bool) error
	GetConnection() (*sql.DB, context.Context, error)
	ExecuteQuery(string, DBTable) (*sql.Rows, error)
	SortTables() error
	DumpSql(config.Config) error
	GetForeignKeys(args []string) *sql.Rows
	ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow
	SelectColumnFromTable(table string, column string)

	// Transaction support
	BeginTx(ctx context.Context) (Transaction, error)

	AdminContentDatatype() AdminContentDataRepository
	AdminContentFields() AdminContentFieldRepository
	AdminDatatypes() AdminDatatypeRepository
    AdminDatatypeFields() AdminDatatypeFieldRepository
	AdminFields() AdminFieldRepository
	AdminRoutes() AdminRoutesRepository
	ContentData() ContentDataRepository
	ContentFields() ContentFieldRepository
	Datatypes() DatatypeRepository
    DatatypeFields() DatatypeFieldsRepository
	Fields() FieldRepository
	Media() MediaRepository
	MediaDimensions() MediaDimensionRepository
	Permissions() PermissionRepository
	Roles() RoleRepository
	Routes() RouteRepository
	Sessions() SessionRepository
	Tables() TablesRepository
	UserOauth() UserOauthRepository
	Users() UserRepository
}
