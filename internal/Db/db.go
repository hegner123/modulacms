package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"sort"

	config "github.com/hegner123/modulacms/internal/Config"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

//go:embed sql
var sqlFiles embed.FS

type Historied interface {
	GetHistory() string
	MapHistoryEntry() string
	UpdateHistory([]byte) error
}

// Database represents a SQLite database connection
type Database struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

// MysqlDatabase represents a MySQL database connection
type MysqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

// PsqlDatabase represents a PostgreSQL database connection
type PsqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}

// DbStatus represents the status of a database connection
type DbStatus string

const (
	Open   DbStatus = "open"
	Closed DbStatus = "closed"
	Err    DbStatus = "error"
)

// DbDriver is the interface for all database drivers
type DbDriver interface {
	// Database Connection
	CreateAllTables() error
	InitDB(v *bool) error
	Ping() error
	GetConnection() (*sql.DB, context.Context, error)
	ExecuteQuery(string, DBTable) (*sql.Rows, error)
	SortTables() error
	DumpSql(config.Config) error
	//RestoreDB() error

	// Count operations
	CountAdminContentData() (*int64, error)
	CountAdminContentFields() (*int64, error)
	CountAdminDatatypes() (*int64, error)
	CountAdminDatatypeFields() (*int64, error)
	CountAdminFields() (*int64, error)
	CountAdminRoutes() (*int64, error)
	CountContentData() (*int64, error)
	CountContentFields() (*int64, error)
	CountDatatypes() (*int64, error)
	CountDatatypeFields() (*int64, error)
	CountFields() (*int64, error)
	CountMedia() (*int64, error)
	CountMediaDimensions() (*int64, error)
	CountPermissions() (*int64, error)
	CountRoles() (*int64, error)
	CountRoutes() (*int64, error)
	CountSessions() (*int64, error)
	CountTables() (*int64, error)
	CountTokens() (*int64, error)
	CountUsers() (*int64, error)
	CountUserOauths() (*int64, error)

	// Create operations
	CreateAdminContentData(CreateAdminContentDataParams) AdminContentData
	CreateAdminContentField(CreateAdminContentFieldParams) AdminContentFields
	CreateAdminDatatype(CreateAdminDatatypeParams) AdminDatatypes
	CreateAdminDatatypeField(CreateAdminDatatypeFieldParams) AdminDatatypeFields
	CreateAdminField(CreateAdminFieldParams) AdminFields
	CreateAdminRoute(CreateAdminRouteParams) AdminRoutes
	CreateContentData(CreateContentDataParams) ContentData
	CreateContentField(CreateContentFieldParams) ContentFields
	CreateDatatype(CreateDatatypeParams) Datatypes
	CreateDatatypeField(CreateDatatypeFieldParams) DatatypeFields
	CreateField(CreateFieldParams) Fields
	CreateMedia(CreateMediaParams) Media
	CreateMediaDimension(CreateMediaDimensionParams) MediaDimensions
	CreatePermission(CreatePermissionParams) Permissions
	CreateRole(CreateRoleParams) Roles
	CreateRoute(CreateRouteParams) Routes
	CreateSession(CreateSessionParams) (*Sessions, error)
	CreateTable(string) Tables
	CreateToken(CreateTokenParams) Tokens
	CreateUser(CreateUserParams) (*Users, error)
	CreateUserOauth(CreateUserOauthParams) (*UserOauth, error)

	// Create table operations
	CreateAdminContentDataTable() error
	CreateAdminContentFieldTable() error
	CreateAdminDatatypeTable() error
	CreateAdminDatatypeFieldTable() error
	CreateAdminFieldTable() error
	CreateAdminRouteTable() error
	CreateContentDataTable() error
	CreateContentFieldTable() error
	CreateDatatypeTable() error
	CreateDatatypeFieldTable() error
	CreateFieldTable() error
	CreateMediaTable() error
	CreateMediaDimensionTable() error
	CreatePermissionTable() error
	CreateRoleTable() error
	CreateRouteTable() error
	CreateSessionTable() error
	CreateTableTable() error
	CreateTokenTable() error
	CreateUserTable() error
	CreateUserOauthTable() error

	// Delete operations
	DeleteAdminContentData(int64) error
	DeleteAdminContentField(int64) error
	DeleteAdminDatatype(int64) error
	DeleteAdminDatatypeField(int64) error
	DeleteAdminField(int64) error
	DeleteAdminRoute(int64) error
	DeleteContentData(int64) error
	DeleteContentField(int64) error
	DeleteDatatype(int64) error
	DeleteDatatypeField(int64) error
	DeleteField(int64) error
	DeleteMedia(int64) error
	DeleteMediaDimension(int64) error
	DeletePermission(int64) error
	DeleteRole(int64) error
	DeleteRoute(int64) error
	DeleteSession(int64) error
	DeleteTable(int64) error
	DeleteToken(int64) error
	DeleteUser(int64) error
	DeleteUserOauth(int64) error

	// Get operations
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
	GetMediaByName(string) (*Media, error)
	GetMediaByURL(string) (*Media, error)
	GetMediaDimension(int64) (*MediaDimensions, error)
	GetPermission(int64) (*Permissions, error)
	GetRole(int64) (*Roles, error)
	GetRoute(int64) (*Routes, error)
	GetRouteID(string) (*int64, error)
	GetSession(int64) (*Sessions, error)
	GetSessionsByUserId(int64) (*[]Sessions, error)
	GetTable(int64) (*Tables, error)
	GetToken(int64) (*Tokens, error)
	GetTokenByUserId(int64) (*[]Tokens, error)
	GetUser(int64) (*Users, error)
	GetUserOauth(int64) (*UserOauth, error)
	GetUserByEmail(string) (*Users, error)

	// List operations
	ListAdminContentData() (*[]AdminContentData, error)
	ListAdminContentDataByRoute(int64) (*[]AdminContentData, error)
	ListAdminContentFields() (*[]AdminContentFields, error)
	ListAdminContentFieldsByRoute(int64) (*[]AdminContentFields, error)
	ListAdminDatatypes() (*[]AdminDatatypes, error)
	ListAdminDatatypeField() (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByDatatypeID(int64) (*[]AdminDatatypeFields, error)
	ListAdminDatatypeFieldByFieldID(int64) (*[]AdminDatatypeFields, error)
	ListAdminFields() (*[]AdminFields, error)
	ListAdminRoutes() (*[]AdminRoutes, error)
	ListContentData() (*[]ContentData, error)
	ListContentDataByRoute(int64) (*[]ContentData, error)
	ListContentFields() (*[]ContentFields, error)
	ListContentFieldsByRoute(int64) (*[]ContentFields, error)
	ListDatatypes() (*[]Datatypes, error)
	ListDatatypeField() (*[]DatatypeFields, error)
	ListDatatypeFieldByDatatypeID(int64) (*[]DatatypeFields, error)
	ListDatatypeFieldByFieldID(int64) (*[]DatatypeFields, error)
	ListFields() (*[]Fields, error)
	ListFieldsByDatatypeID(int64) (*[]Fields, error)
	ListMedia() (*[]Media, error)
	ListMediaDimensions() (*[]MediaDimensions, error)
	ListPermissions() (*[]Permissions, error)
	ListRoles() (*[]Roles, error)
	ListRoutes() (*[]Routes, error)
	ListSessions() (*[]Sessions, error)
	ListTables() (*[]Tables, error)
	ListTokens() (*[]Tokens, error)
	ListUsers() (*[]Users, error)
	ListUserOauths() (*[]UserOauth, error)

	// Update operations
	UpdateAdminContentData(UpdateAdminContentDataParams) (*string, error)
	UpdateAdminContentField(UpdateAdminContentFieldParams) (*string, error)
	UpdateAdminDatatype(UpdateAdminDatatypeParams) (*string, error)
	UpdateAdminDatatypeField(UpdateAdminDatatypeFieldParams) (*string, error)
	UpdateAdminField(UpdateAdminFieldParams) (*string, error)
	UpdateAdminRoute(UpdateAdminRouteParams) (*string, error)
	UpdateContentData(UpdateContentDataParams) (*string, error)
	UpdateContentField(UpdateContentFieldParams) (*string, error)
	UpdateDatatype(UpdateDatatypeParams) (*string, error)
	UpdateDatatypeField(UpdateDatatypeFieldParams) (*string, error)
	UpdateField(UpdateFieldParams) (*string, error)
	UpdateMedia(UpdateMediaParams) (*string, error)
	UpdateMediaDimension(UpdateMediaDimensionParams) (*string, error)
	UpdatePermission(UpdatePermissionParams) (*string, error)
	UpdateRole(UpdateRoleParams) (*string, error)
	UpdateRoute(UpdateRouteParams) (*string, error)
	UpdateSession(UpdateSessionParams) (*string, error)
	UpdateTable(UpdateTableParams) (*string, error)
	UpdateToken(UpdateTokenParams) (*string, error)
	UpdateUser(UpdateUserParams) (*string, error)
	UpdateUserOauth(UpdateUserOauthParams) (*string, error)
}

// GetConnection returns the database connection and context
func (d Database) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

// GetConnection returns the MySQL database connection and context
func (d MysqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

// GetConnection returns the PostgreSQL database connection and context
func (d PsqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

// Ping checks if the database connection is still alive
func (d Database) Ping() error {
	return d.Connection.Ping()
}

// Ping checks if the MySQL database connection is still alive
func (d MysqlDatabase) Ping() error {
	return d.Connection.Ping()
}

// Ping checks if the PostgreSQL database connection is still alive
func (d PsqlDatabase) Ping() error {
	return d.Connection.Ping()
}

// ExecuteQuery executes a raw SQL query on the database
func (d Database) ExecuteQuery(query string, table DBTable) (*sql.Rows, error) {
	q := fmt.Sprintf("%s %s;", query, DBTableString(table))
	return d.Connection.Query(q)
}

// ExecuteQuery executes a raw SQL query on the MySQL database
func (d MysqlDatabase) ExecuteQuery(query string, table DBTable) (*sql.Rows, error) {
	q := fmt.Sprintf("%s %s;", query, DBTableString(table))
	return d.Connection.Query(q)
}

// ExecuteQuery executes a raw SQL query on the PostgreSQL database
func (d PsqlDatabase) ExecuteQuery(query string, table DBTable) (*sql.Rows, error) {
	q := fmt.Sprintf("%s %s;", query, DBTableString(table))
	return d.Connection.Query(q)
}

// CreateAllTables creates all database tables
func (d Database) CreateAllTables() error {
	// Create all tables
	err := d.CreateUserTable()
	if err != nil {
		return err
	}

	err = d.CreateRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaDimensionTable()
	if err != nil {
		return err
	}

	err = d.CreateTokenTable()
	if err != nil {
		return err
	}

	err = d.CreateSessionTable()
	if err != nil {
		return err
	}

	err = d.CreateRoleTable()
	if err != nil {
		return err
	}

	err = d.CreatePermissionTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateTableTable()
	if err != nil {
		return err
	}

	err = d.CreateUserOauthTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateAllTables creates all MySQL database tables
func (d MysqlDatabase) CreateAllTables() error {
	// Create all tables
	err := d.CreateUserTable()
	if err != nil {
		return err
	}

	err = d.CreateRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaDimensionTable()
	if err != nil {
		return err
	}

	err = d.CreateTokenTable()
	if err != nil {
		return err
	}

	err = d.CreateSessionTable()
	if err != nil {
		return err
	}

	err = d.CreateRoleTable()
	if err != nil {
		return err
	}

	err = d.CreatePermissionTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateTableTable()
	if err != nil {
		return err
	}

	err = d.CreateUserOauthTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateAllTables creates all PostgreSQL database tables
func (d PsqlDatabase) CreateAllTables() error {
	// Create all tables
	err := d.CreateUserTable()
	if err != nil {
		return err
	}

	err = d.CreateRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaTable()
	if err != nil {
		return err
	}

	err = d.CreateMediaDimensionTable()
	if err != nil {
		return err
	}

	err = d.CreateTokenTable()
	if err != nil {
		return err
	}

	err = d.CreateSessionTable()
	if err != nil {
		return err
	}

	err = d.CreateRoleTable()
	if err != nil {
		return err
	}

	err = d.CreatePermissionTable()
	if err != nil {
		return err
	}

	err = d.CreateDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminRouteTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminDatatypeTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentDataTable()
	if err != nil {
		return err
	}

	err = d.CreateAdminContentFieldTable()
	if err != nil {
		return err
	}

	err = d.CreateTableTable()
	if err != nil {
		return err
	}

	err = d.CreateUserOauthTable()
	if err != nil {
		return err
	}

	return nil
}

/*
// InitDb initializes the database
func (d Database) InitDb(v *bool) error {
	err := d.CreateAllTables()
	if err != nil {
		return err
	}
	return nil
}

// InitDb initializes the MySQL database
func (d MysqlDatabase) InitDb(v *bool) error {
	err := d.CreateAllTables()
	if err != nil {
		return err
	}
	return nil
}

// InitDb initializes the PostgreSQL database
func (d PsqlDatabase) InitDb(v *bool) error {
	err := d.CreateAllTables()
	if err != nil {
		return err
	}
	return nil
}
*/

func (d Database) SortTables() error {
	clearTables := "DELETE FROM tables;"
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}

	tb, err := d.ListTables()
	if err != nil {
		return err
	}
	_, err = con.Exec(clearTables)
	if err != nil {
		return err
	}
	tables := *tb
	st := []string{}
	for _, t := range tables {
		st = append(st, t.Label)
	}
	sort.Strings(st)
	for _, tt := range st {
		d.CreateTable(tt)
	}

	return nil
}
func (d MysqlDatabase) SortTables() error {
	clearTables := "DELETE FROM tables;"
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}

	tb, err := d.ListTables()
	if err != nil {
		return err
	}
	_, err = con.Exec(clearTables)
	if err != nil {
		return err
	}
	tables := *tb
	st := []string{}
	for _, t := range tables {
		st = append(st, t.Label)
	}
	sort.Strings(st)
	for _, tt := range st {
		d.CreateTable(tt)
	}

	return nil
}
func (d PsqlDatabase) SortTables() error {
	clearTables := "DELETE FROM tables;"
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}

	tb, err := d.ListTables()
	if err != nil {
		return err
	}
	_, err = con.Exec(clearTables)
	if err != nil {
		return err
	}
	tables := *tb
	st := []string{}
	for _, t := range tables {
		st = append(st, t.Label)
	}
	sort.Strings(st)
	for _, tt := range st {
		d.CreateTable(tt)
	}

	return nil
}

func (d Database) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_sql.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create temporary file: %v", err)
		return err
	}
	// Ensure the file is removed after execution.
	defer func() {
		if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Write the embedded script contents to the temporary file.
	if _, err := tmpFile.Write(script); err != nil {
		utility.DefaultLogger.Fatal("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
    err=tmpFile.Close()
    if err!=nil {
        return err
    }

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Fatal("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "sqlite"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Fatal("failed to execute script: %v, output: %s", err, output)
		return err
	}
	return nil

}
func (d MysqlDatabase) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_mysql.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create temporary file: %v", err)
		return err
	}
	// Ensure the file is removed after execution.
	defer func() {
		if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Write the embedded script contents to the temporary file.
	if _, err := tmpFile.Write(script); err != nil {
		utility.DefaultLogger.Fatal("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
    err=tmpFile.Close()
    if err!=nil {
        return err
    }

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Fatal("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_User, c.Db_Password, c.Db_Name, "mysql"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Fatal("failed to execute script: %v, output: %s", err, string(output))
		return err
	}
	return nil

}
func (d PsqlDatabase) DumpSql(c config.Config) error {

	// Read the embedded Bash script.
	script, err := sqlFiles.ReadFile("sql/dump_psql.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to read embedded script: %v", err)
		return err
	}

	// Create a temporary file for the script.
	tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create temporary file: %v", err)
		return err
	}
	// Ensure the file is removed after execution.
	defer func() {
		if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Write the embedded script contents to the temporary file.
	if _, err := tmpFile.Write(script); err != nil {
		utility.DefaultLogger.Fatal("failed to write script to file: %v", err)
		return err
	}
	// Close the file so that it can be executed.
    err=tmpFile.Close()
    if err!=nil {
        return err
    }

	// Make the temporary file executable.
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		utility.DefaultLogger.Fatal("failed to chmod the temporary file: %v", err)
		return err
	}

	// Execute the Bash script using /bin/bash.
	t := utility.TimestampReadable()
	cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "sqlite"+t+".sql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utility.DefaultLogger.Fatal("failed to execute script: %v, output: %s", err, output)
		return err
	}
	return nil

}
