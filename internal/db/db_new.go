package db

import (
	"context"
	"database/sql"
	"fmt"

	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
)




// SQLiteProvider implements DatabaseProvider interface
type SQLiteProvider struct {
	db     *sql.DB
	ctx    context.Context
	logger *utility.Logger
	config config.Config
}

type MysqlProvider struct {
	db     *sql.DB
	ctx    context.Context
	logger *utility.Logger
	config config.Config
}
type PsqlProvider struct {
	db     *sql.DB
	ctx    context.Context
	logger *utility.Logger
	config config.Config
}

func NewSQLiteProvider(config config.Config) *SQLiteProvider {
	return &SQLiteProvider{
		ctx:    context.Background(),
		logger: utility.DefaultLogger,
		config: config,
	}
}
func NewMysqlProvider(config config.Config) *MysqlProvider {
	return &MysqlProvider{
		ctx:    context.Background(),
		logger: utility.DefaultLogger,
		config: config,
	}
}
func NewPsqlProvider(config config.Config) *PsqlProvider {
	return &PsqlProvider{
		ctx:    context.Background(),
		logger: utility.DefaultLogger,
		config: config,
	}
}

// Connect establishes a connection to the SQLite database
func (p *SQLiteProvider) Connect(ctx context.Context) error {
	p.ctx = ctx
	dbPath := p.config.Db_URL

	if dbPath == "" {
		err := fmt.Errorf("dbPath is empty")
		return err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	p.db = db
	return p.Ping()
}

// GetConnection returns the database connection and context
func (p SQLiteProvider) GetConnection() (*sql.DB, context.Context, error) {
	return p.db, p.ctx, nil
}

// GetConnection returns the MySQL database connection and context
func (p MysqlProvider) GetConnection() (*sql.DB, context.Context, error) {
	return p.db, p.ctx, nil
}

// GetConnection returns the PostgreSQL database connection and context
func (p PsqlProvider) GetConnection() (*sql.DB, context.Context, error) {
	return p.db, p.ctx, nil
}

// Ping checks if the database connection is still alive
func (p SQLiteProvider) Ping() error {
	return p.db.Ping()
}

// Ping checks if the MySQL database connection is still alive
func (p MysqlProvider) Ping() error {
	return p.db.Ping()
}

// Ping checks if the PostgreSQL database connection is still alive
func (p PsqlProvider) Ping() error {
	return p.db.Ping()
}

