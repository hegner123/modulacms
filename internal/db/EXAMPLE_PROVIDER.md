# Database Provider Implementation Example

This example demonstrates how to implement the `DatabaseProvider` interface from the ModulaCMS project.

## Implementation Example

```go
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	
	_ "github.com/mattn/go-sqlite3"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// SQLiteProvider implements DatabaseProvider interface
type SQLiteProvider struct {
	db     *sql.DB
	ctx    context.Context
	logger *utility.Logger
	config config.Config
}

// NewSQLiteProvider creates a new SQLite database provider
func NewSQLiteProvider(config config.Config) *SQLiteProvider {
	return &SQLiteProvider{
		ctx:    context.Background(),
		logger: utility.DefaultLogger(),
		config: config,
	}
}

// Connect establishes a connection to the SQLite database
func (p *SQLiteProvider) Connect(ctx context.Context) error {
	p.ctx = ctx
	dbPath := p.config.Database.Path
	
	if dbPath == "" {
		return errors.New("database path is not specified in config")
	}
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	
	p.db = db
	return p.Ping()
}

// Close closes the database connection
func (p *SQLiteProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping verifies the connection to the database
func (p *SQLiteProvider) Ping() error {
	if p.db == nil {
		return errors.New("database not connected")
	}
	return p.db.PingContext(p.ctx)
}

// GetConnection returns the database connection
func (p *SQLiteProvider) GetConnection() (*sql.DB, context.Context, error) {
	if p.db == nil {
		return nil, nil, errors.New("database not connected")
	}
	return p.db, p.ctx, nil
}

// ExecuteQuery executes a SQL query
func (p *SQLiteProvider) ExecuteQuery(query string, table DBTable) (*sql.Rows, error) {
	if p.db == nil {
		return nil, errors.New("database not connected")
	}
	
	rows, err := p.db.QueryContext(p.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	
	return rows, nil
}

// BeginTx starts a new transaction
func (p *SQLiteProvider) BeginTx(ctx context.Context) (Transaction, error) {
	if p.db == nil {
		return Transaction{}, errors.New("database not connected")
	}
	
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	return Transaction{*tx}, nil
}

// CreateAllTables creates all database tables
func (p *SQLiteProvider) CreateAllTables() error {
	// Create tables in sequence to respect foreign key constraints
	tables := []func(context.Context) error{
		p.Permissions().CreateTable,
		p.Roles().CreateTable,
		p.MediaDimensions().CreateTable,
		p.Users().CreateTable,
		p.AdminRoutes().CreateTable,
		p.Routes().CreateTable,
		p.Datatypes().CreateTable,
		p.Fields().CreateTable,
		p.AdminDatatypes().CreateTable,
		p.AdminFields().CreateTable,
		p.Tables().CreateTable,
		p.Media().CreateTable,
		p.Sessions().CreateTable,
		p.ContentData().CreateTable,
		p.ContentFields().CreateTable,
		p.AdminContentDatatype().CreateTable,
		p.AdminContentFields().CreateTable,
		p.DatatypeFields().CreateTable,
		p.AdminDatatypeFields().CreateTable,
	}
	
	for _, createTable := range tables {
		if err := createTable(p.ctx); err != nil {
			return err
		}
	}
	
	return nil
}

// InitDB initializes the database with required data
func (p *SQLiteProvider) InitDB(initialized *bool) error {
	if *initialized {
		p.logger.Info("Database already initialized")
		return nil
	}
	
	// Create all tables
	if err := p.CreateAllTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	
	// Initialize with default data
	if err := p.setupDefaultData(); err != nil {
		return fmt.Errorf("failed to setup default data: %w", err)
	}
	
	*initialized = true
	p.logger.Info("Database initialized successfully")
	return nil
}

// setupDefaultData creates default required records
func (p *SQLiteProvider) setupDefaultData() error {
	// Create admin user
	_, err := p.Users().Create(p.ctx, CreateUserParams{
		Email:    "admin@example.com",
		Password: "hashed_password", // In real implementation, use proper password hashing
		Name:     "Admin",
		Active:   true,
		RoleID:   1, // Admin role
	})
	if err != nil {
		return err
	}
	
	// Create default roles
	_, err = p.Roles().Create(p.ctx, CreateRoleParams{
		Name:        "Admin",
		Description: "Administrator with all permissions",
	})
	if err != nil {
		return err
	}
	
	// Create other default data as needed
	
	return nil
}

// SortTables sorts tables by their dependency order
func (p *SQLiteProvider) SortTables() error {
	// Implementation for sorting tables by dependencies
	return nil
}

// DumpSql dumps the database schema to SQL files
func (p *SQLiteProvider) DumpSql(config config.Config) error {
	// Implementation for dumping database schema
	return nil
}

// GetForeignKeys retrieves foreign key relationships
func (p *SQLiteProvider) GetForeignKeys(args []string) *sql.Rows {
	// Query foreign key information
	query := `
	SELECT
		m.name as table_name,
		p.name as column_name,
		p."table" as referenced_table,
		p."to" as referenced_column
	FROM 
		sqlite_master m
	JOIN 
		pragma_foreign_key_list(m.name) p
	WHERE 
		m.type = 'table'
	`
	
	rows, err := p.db.QueryContext(p.ctx, query)
	if err != nil {
		p.logger.Error("Failed to get foreign keys:", err)
		return nil
	}
	
	return rows
}

// ScanForeignKeyQueryRows scans foreign key query results
func (p *SQLiteProvider) ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow {
	var result []SqliteForeignKeyQueryRow
	
	for rows.Next() {
		var row SqliteForeignKeyQueryRow
		if err := rows.Scan(&row.TableName, &row.ColumnName, &row.RefTable, &row.RefColumn); err != nil {
			p.logger.Error("Error scanning foreign key row:", err)
			continue
		}
		result = append(result, row)
	}
	
	return result
}

// SelectColumnFromTable selects a specific column from a table
func (p *SQLiteProvider) SelectColumnFromTable(table string, column string) {
	// Implementation for selecting a column from a table
	// This method is incomplete in the interface declaration
}

// Repository accessor methods
func (p *SQLiteProvider) AdminContentDatatype() AdminContentDataRepository {
	return &adminContentDataRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) AdminContentFields() AdminContentFieldRepository {
	return &adminContentFieldRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) AdminDatatypes() AdminDatatypeRepository {
	return &adminDatatypeRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) AdminDatatypeFields() AdminDatatypeFieldRepository {
	return &adminDatatypeFieldRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) AdminFields() AdminFieldRepository {
	return &adminFieldRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) AdminRoutes() AdminRoutesRepository {
	return &adminRoutesRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) ContentData() ContentDataRepository {
	return &contentDataRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) ContentFields() ContentFieldRepository {
	return &contentFieldRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Datatypes() DatatypeRepository {
	return &datatypeRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) DatatypeFields() DatatypeFieldsRepository {
	return &datatypeFieldsRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Fields() FieldRepository {
	return &fieldRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Media() MediaRepository {
	return &mediaRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) MediaDimensions() MediaDimensionRepository {
	return &mediaDimensionRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Permissions() PermissionRepository {
	return &permissionRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Roles() RoleRepository {
	return &roleRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Routes() RouteRepository {
	return &routeRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Sessions() SessionRepository {
	return &sessionRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Tables() TablesRepository {
	return &tablesRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Tokens() TokenRepository {
	return &tokenRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) UserOauth() UserOauthRepository {
	return &userOauthRepo{db: p.db, ctx: p.ctx}
}

func (p *SQLiteProvider) Users() UserRepository {
	return &userRepo{db: p.db, ctx: p.ctx}
}

// Example repository implementation
type userRepo struct {
	db  *sql.DB
	ctx context.Context
}

func (r *userRepo) Counte(ctx context.Context) (*int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

func (r *userRepo) Create(ctx context.Context, params CreateUserParams) (*Users, error) {
	query := `
	INSERT INTO users (email, password, name, active, role_id)
	VALUES (?, ?, ?, ?, ?)
	RETURNING id, email, password, name, active, role_id, created_at, updated_at
	`
	
	var user Users
	err := r.db.QueryRowContext(ctx, query,
		params.Email, params.Password, params.Name, params.Active, params.RoleID,
	).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Active, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *userRepo) CreateTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		name TEXT NOT NULL,
		active BOOLEAN NOT NULL DEFAULT true,
		role_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (role_id) REFERENCES roles(id)
	)
	`
	
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*Users, error) {
	query := `
	SELECT id, email, password, name, active, role_id, created_at, updated_at
	FROM users
	WHERE id = ?
	`
	
	var user Users
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Active, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*Users, error) {
	query := `
	SELECT id, email, password, name, active, role_id, created_at, updated_at
	FROM users
	WHERE email = ?
	`
	
	var user Users
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Active, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *userRepo) List(ctx context.Context) ([]*Users, error) {
	query := `
	SELECT id, email, password, name, active, role_id, created_at, updated_at
	FROM users
	ORDER BY id
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*Users
	for rows.Next() {
		var user Users
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.Name, &user.Active, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
}

func (r *userRepo) Update(ctx context.Context, params UpdateUserParams) error {
	query := `
	UPDATE users
	SET email = ?, password = ?, name = ?, active = ?, role_id = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`
	
	_, err := r.db.ExecContext(ctx, query,
		params.Email, params.Password, params.Name, params.Active, params.RoleID, params.ID,
	)
	
	return err
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
```

## Using the Provider

Here's an example of how to use the SQLite database provider:

```go
package main

import (
	"context"
	"log"
	
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create database provider
	provider := db.NewSQLiteProvider(cfg)
	
	// Connect to database
	ctx := context.Background()
	err = provider.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer provider.Close()
	
	// Initialize database if needed
	initialized := false
	err = provider.InitDB(&initialized)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Example: Create a new user
	user, err := provider.Users().Create(ctx, db.CreateUserParams{
		Email:    "user@example.com",
		Password: "secure_password", // Use proper password hashing in production
		Name:     "Example User",
		Active:   true,
		RoleID:   2, // Regular user role
	})
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	
	log.Printf("Created user ID: %d, Name: %s", user.ID, user.Name)
	
	// Example: List all users
	users, err := provider.Users().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}
	
	log.Printf("Found %d users:", len(users))
	for _, u := range users {
		log.Printf("- ID: %d, Email: %s, Name: %s", u.ID, u.Email, u.Name)
	}
}
```

## Transaction Example

Here's an example of using transactions:

```go
func createUserWithRole(provider db.DatabaseProvider, ctx context.Context) error {
	// Start a transaction
	tx, err := provider.BeginTx(ctx)
	if err != nil {
		return err
	}
	
	// Rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	
	// Create a role
	role, err := provider.Roles().Create(ctx, db.CreateRoleParams{
		Name:        "Editor",
		Description: "Content editor role",
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	
	// Create a user with the new role
	_, err = provider.Users().Create(ctx, db.CreateUserParams{
		Email:    "editor@example.com",
		Password: "secure_password", // Use proper password hashing in production
		Name:     "Editor User",
		Active:   true,
		RoleID:   role.ID,
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	
	// Commit the transaction
	return tx.Commit()
}
```