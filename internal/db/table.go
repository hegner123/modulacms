package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

// Tables represents a table in the database.
type Tables struct {
	ID       string               `json:"id"`
	Label    string               `json:"label"`
	AuthorID types.NullableUserID `json:"author_id"`
}

// CreateTableParams holds parameters for creating a table.
type CreateTableParams struct {
	Label string `json:"label"`
}

// UpdateTableParams holds parameters for updating a table.
type UpdateTableParams struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

// FormParams and HistoryEntry variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringTable converts Tables to StringTables for table display.
func MapStringTable(a Tables) StringTables {
	return StringTables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapTable converts a sqlc-generated SQLite table to the wrapper type.
func (d Database) MapTable(a mdb.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

// MapCreateTableParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateTableParams(a CreateTableParams) mdb.CreateTableParams {
	return mdb.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

// MapUpdateTableParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdateTableParams(a UpdateTableParams) mdb.UpdateTableParams {
	return mdb.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

// CountTables returns the total number of tables in the database.
func (d Database) CountTables() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateTableTable creates the tables table.
func (d Database) CreateTableTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

// CreateTable creates a new table with audit tracking.
func (d Database) CreateTable(ctx context.Context, ac audited.AuditContext, s CreateTableParams) (*Tables, error) {
	cmd := d.NewTableCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	r := d.MapTable(result)
	return &r, nil
}

// DeleteTable deletes a table with audit tracking.
func (d Database) DeleteTable(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTableCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetTable retrieves a single table by ID.
func (d Database) GetTable(id string) (*Tables, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdb.GetTableParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

// ListTables retrieves all tables.
func (d Database) ListTables() (*[]Tables, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListTable(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	res := []Tables{}
	for _, v := range rows {
		m := d.MapTable(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateTable updates a table with audit tracking.
func (d Database) UpdateTable(ctx context.Context, ac audited.AuditContext, s UpdateTableParams) (*string, error) {
	cmd := d.UpdateTableCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update table: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapTable converts a sqlc-generated MySQL table to the wrapper type.
func (d MysqlDatabase) MapTable(a mdbm.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

// MapCreateTableParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateTableParams(a CreateTableParams) mdbm.CreateTableParams {
	return mdbm.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

// MapUpdateTableParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbm.UpdateTableParams {
	return mdbm.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

// CountTables returns the total number of tables in the database.
func (d MysqlDatabase) CountTables() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateTableTable creates the tables table.
func (d MysqlDatabase) CreateTableTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

// CreateTable creates a new table with audit tracking.
func (d MysqlDatabase) CreateTable(ctx context.Context, ac audited.AuditContext, s CreateTableParams) (*Tables, error) {
	cmd := d.NewTableCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	r := d.MapTable(result)
	return &r, nil
}

// DeleteTable deletes a table with audit tracking.
func (d MysqlDatabase) DeleteTable(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTableCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetTable retrieves a single table by ID.
func (d MysqlDatabase) GetTable(id string) (*Tables, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdbm.GetTableParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

// ListTables retrieves all tables.
func (d MysqlDatabase) ListTables() (*[]Tables, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListTable(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	res := []Tables{}
	for _, v := range rows {
		m := d.MapTable(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateTable updates a table with audit tracking.
func (d MysqlDatabase) UpdateTable(ctx context.Context, ac audited.AuditContext, s UpdateTableParams) (*string, error) {
	cmd := d.UpdateTableCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update table: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapTable converts a sqlc-generated PostgreSQL table to the wrapper type.
func (d PsqlDatabase) MapTable(a mdbp.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

// MapCreateTableParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateTableParams(a CreateTableParams) mdbp.CreateTableParams {
	return mdbp.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

// MapUpdateTableParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbp.UpdateTableParams {
	return mdbp.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

// CountTables returns the total number of tables in the database.
func (d PsqlDatabase) CountTables() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateTableTable creates the tables table.
func (d PsqlDatabase) CreateTableTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

// CreateTable creates a new table with audit tracking.
func (d PsqlDatabase) CreateTable(ctx context.Context, ac audited.AuditContext, s CreateTableParams) (*Tables, error) {
	cmd := d.NewTableCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	r := d.MapTable(result)
	return &r, nil
}

// DeleteTable deletes a table with audit tracking.
func (d PsqlDatabase) DeleteTable(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTableCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetTable retrieves a single table by ID.
func (d PsqlDatabase) GetTable(id string) (*Tables, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdbp.GetTableParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

// ListTables retrieves all tables.
func (d PsqlDatabase) ListTables() (*[]Tables, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListTable(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	res := []Tables{}
	for _, v := range rows {
		m := d.MapTable(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateTable updates a table with audit tracking.
func (d PsqlDatabase) UpdateTable(ctx context.Context, ac audited.AuditContext, s UpdateTableParams) (*string, error) {
	cmd := d.UpdateTableCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update table: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewTableCmd is an audited command for creating tables in SQLite.
type NewTableCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTableParams
	conn     *sql.DB
}

// Context returns the context for the command.
func (c NewTableCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewTableCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c NewTableCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c NewTableCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the name of the table being operated on.
func (c NewTableCmd) TableName() string                           { return "tables" }

// Params returns the parameters for the command.
func (c NewTableCmd) Params() any                                 { return c.params }

// GetID extracts the ID from the created table.
func (c NewTableCmd) GetID(x mdb.Tables) string {
	return x.ID
}

// Execute creates a table in the database.
func (c NewTableCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Tables, error) {
	queries := mdb.New(tx)
	return queries.CreateTable(ctx, mdb.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: c.params.Label,
	})
}

// NewTableCmd creates a new command for creating a table.
func (d Database) NewTableCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTableParams) NewTableCmd {
	return NewTableCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTableCmd is an audited command for updating tables in SQLite.
type UpdateTableCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTableParams
	conn     *sql.DB
}

// Context returns the context for the command.
func (c UpdateTableCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateTableCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c UpdateTableCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c UpdateTableCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the name of the table being operated on.
func (c UpdateTableCmd) TableName() string                           { return "tables" }

// Params returns the parameters for the command.
func (c UpdateTableCmd) Params() any                                 { return c.params }

// GetID returns the ID of the table being updated.
func (c UpdateTableCmd) GetID() string                               { return c.params.ID }

// GetBefore retrieves the table state before the update.
func (c UpdateTableCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tables, error) {
	queries := mdb.New(tx)
	return queries.GetTable(ctx, mdb.GetTableParams{ID: c.params.ID})
}

// Execute updates a table in the database.
func (c UpdateTableCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateTable(ctx, mdb.UpdateTableParams{
		Label: c.params.Label,
		ID:    c.params.ID,
	})
}

// UpdateTableCmd creates a new command for updating a table.
func (d Database) UpdateTableCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTableParams) UpdateTableCmd {
	return UpdateTableCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTableCmd is an audited command for deleting tables in SQLite.
type DeleteTableCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

// Context returns the context for the command.
func (c DeleteTableCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteTableCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c DeleteTableCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c DeleteTableCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the name of the table being operated on.
func (c DeleteTableCmd) TableName() string                           { return "tables" }

// GetID returns the ID of the table being deleted.
func (c DeleteTableCmd) GetID() string                               { return c.id }

// GetBefore retrieves the table state before the deletion.
func (c DeleteTableCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tables, error) {
	queries := mdb.New(tx)
	return queries.GetTable(ctx, mdb.GetTableParams{ID: c.id})
}

// Execute deletes a table from the database.
func (c DeleteTableCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteTable(ctx, mdb.DeleteTableParams{ID: c.id})
}

// DeleteTableCmd creates a new command for deleting a table.
func (d Database) DeleteTableCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTableCmd {
	return DeleteTableCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewTableCmdMysql is an audited command for creating tables in MySQL.
type NewTableCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTableParams
	conn     *sql.DB
}

// Context returns the context for the command.
func (c NewTableCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewTableCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c NewTableCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c NewTableCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the name of the table being operated on.
func (c NewTableCmdMysql) TableName() string                           { return "tables" }

// Params returns the parameters for the command.
func (c NewTableCmdMysql) Params() any                                 { return c.params }

// GetID extracts the ID from the created table.
func (c NewTableCmdMysql) GetID(x mdbm.Tables) string {
	return x.ID
}

// Execute creates a table in the database.
func (c NewTableCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Tables, error) {
	id := string(types.NewTableID())
	queries := mdbm.New(tx)
	err := queries.CreateTable(ctx, mdbm.CreateTableParams{
		ID:    id,
		Label: c.params.Label,
	})
	if err != nil {
		return mdbm.Tables{}, fmt.Errorf("Failed to CreateTable: %w", err)
	}
	return queries.GetTable(ctx, mdbm.GetTableParams{ID: id})
}

// NewTableCmd creates a new command for creating a table.
func (d MysqlDatabase) NewTableCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTableParams) NewTableCmdMysql {
	return NewTableCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTableCmdMysql is an audited command for updating tables in MySQL.
type UpdateTableCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTableParams
	conn     *sql.DB
}

// Context returns the context for the command.
func (c UpdateTableCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateTableCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c UpdateTableCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c UpdateTableCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the name of the table being operated on.
func (c UpdateTableCmdMysql) TableName() string                           { return "tables" }

// Params returns the parameters for the command.
func (c UpdateTableCmdMysql) Params() any                                 { return c.params }

// GetID returns the ID of the table being updated.
func (c UpdateTableCmdMysql) GetID() string                               { return c.params.ID }

// GetBefore retrieves the table state before the update.
func (c UpdateTableCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tables, error) {
	queries := mdbm.New(tx)
	return queries.GetTable(ctx, mdbm.GetTableParams{ID: c.params.ID})
}

// Execute updates a table in the database.
func (c UpdateTableCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateTable(ctx, mdbm.UpdateTableParams{
		Label: c.params.Label,
		ID:    c.params.ID,
	})
}

// UpdateTableCmd creates a new command for updating a table.
func (d MysqlDatabase) UpdateTableCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTableParams) UpdateTableCmdMysql {
	return UpdateTableCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTableCmdMysql is an audited command for deleting tables in MySQL.
type DeleteTableCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

// Context returns the context for the command.
func (c DeleteTableCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteTableCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c DeleteTableCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c DeleteTableCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the name of the table being operated on.
func (c DeleteTableCmdMysql) TableName() string                           { return "tables" }

// GetID returns the ID of the table being deleted.
func (c DeleteTableCmdMysql) GetID() string                               { return c.id }

// GetBefore retrieves the table state before the deletion.
func (c DeleteTableCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tables, error) {
	queries := mdbm.New(tx)
	return queries.GetTable(ctx, mdbm.GetTableParams{ID: c.id})
}

// Execute deletes a table from the database.
func (c DeleteTableCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteTable(ctx, mdbm.DeleteTableParams{ID: c.id})
}

// DeleteTableCmd creates a new command for deleting a table.
func (d MysqlDatabase) DeleteTableCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTableCmdMysql {
	return DeleteTableCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewTableCmdPsql is an audited command for creating tables in PostgreSQL.
type NewTableCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTableParams
	conn     *sql.DB
}

// Context returns the context for the command.
func (c NewTableCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewTableCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c NewTableCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c NewTableCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the name of the table being operated on.
func (c NewTableCmdPsql) TableName() string                           { return "tables" }

// Params returns the parameters for the command.
func (c NewTableCmdPsql) Params() any                                 { return c.params }

// GetID extracts the ID from the created table.
func (c NewTableCmdPsql) GetID(x mdbp.Tables) string {
	return x.ID
}

// Execute creates a table in the database.
func (c NewTableCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Tables, error) {
	queries := mdbp.New(tx)
	return queries.CreateTable(ctx, mdbp.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: c.params.Label,
	})
}

// NewTableCmd creates a new command for creating a table.
func (d PsqlDatabase) NewTableCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTableParams) NewTableCmdPsql {
	return NewTableCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTableCmdPsql is an audited command for updating tables in PostgreSQL.
type UpdateTableCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTableParams
	conn     *sql.DB
}

// Context returns the context for the command.
func (c UpdateTableCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateTableCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c UpdateTableCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c UpdateTableCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the name of the table being operated on.
func (c UpdateTableCmdPsql) TableName() string                           { return "tables" }

// Params returns the parameters for the command.
func (c UpdateTableCmdPsql) Params() any                                 { return c.params }

// GetID returns the ID of the table being updated.
func (c UpdateTableCmdPsql) GetID() string                               { return c.params.ID }

// GetBefore retrieves the table state before the update.
func (c UpdateTableCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tables, error) {
	queries := mdbp.New(tx)
	return queries.GetTable(ctx, mdbp.GetTableParams{ID: c.params.ID})
}

// Execute updates a table in the database.
func (c UpdateTableCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateTable(ctx, mdbp.UpdateTableParams{
		Label: c.params.Label,
		ID:    c.params.ID,
	})
}

// UpdateTableCmd creates a new command for updating a table.
func (d PsqlDatabase) UpdateTableCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTableParams) UpdateTableCmdPsql {
	return UpdateTableCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTableCmdPsql is an audited command for deleting tables in PostgreSQL.
type DeleteTableCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

// Context returns the context for the command.
func (c DeleteTableCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteTableCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection for the command.
func (c DeleteTableCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c DeleteTableCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the name of the table being operated on.
func (c DeleteTableCmdPsql) TableName() string                           { return "tables" }

// GetID returns the ID of the table being deleted.
func (c DeleteTableCmdPsql) GetID() string                               { return c.id }

// GetBefore retrieves the table state before the deletion.
func (c DeleteTableCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tables, error) {
	queries := mdbp.New(tx)
	return queries.GetTable(ctx, mdbp.GetTableParams{ID: c.id})
}

// Execute deletes a table from the database.
func (c DeleteTableCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteTable(ctx, mdbp.DeleteTableParams{ID: c.id})
}

// DeleteTableCmd creates a new command for deleting a table.
func (d PsqlDatabase) DeleteTableCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTableCmdPsql {
	return DeleteTableCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
