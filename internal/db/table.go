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

type Tables struct {
	ID       string               `json:"id"`
	Label    string               `json:"label"`
	AuthorID types.NullableUserID `json:"author_id"`
}

type CreateTableParams struct {
	Label string `json:"label"`
}

type UpdateTableParams struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

// FormParams and HistoryEntry variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringTable converts Tables to StringTables for table display
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

func (d Database) MapTable(a mdb.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

func (d Database) MapCreateTableParams(a CreateTableParams) mdb.CreateTableParams {
	return mdb.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

func (d Database) MapUpdateTableParams(a UpdateTableParams) mdb.UpdateTableParams {
	return mdb.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

func (d Database) CountTables() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateTableTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d Database) CreateTable(ctx context.Context, ac audited.AuditContext, s CreateTableParams) (*Tables, error) {
	cmd := d.NewTableCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	r := d.MapTable(result)
	return &r, nil
}

func (d Database) DeleteTable(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTableCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetTable(id string) (*Tables, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdb.GetTableParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

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

func (d MysqlDatabase) MapTable(a mdbm.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

func (d MysqlDatabase) MapCreateTableParams(a CreateTableParams) mdbm.CreateTableParams {
	return mdbm.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

func (d MysqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbm.UpdateTableParams {
	return mdbm.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

func (d MysqlDatabase) CountTables() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateTableTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateTable(ctx context.Context, ac audited.AuditContext, s CreateTableParams) (*Tables, error) {
	cmd := d.NewTableCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	r := d.MapTable(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteTable(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTableCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetTable(id string) (*Tables, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdbm.GetTableParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

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

func (d PsqlDatabase) MapTable(a mdbp.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

func (d PsqlDatabase) MapCreateTableParams(a CreateTableParams) mdbp.CreateTableParams {
	return mdbp.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

func (d PsqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbp.UpdateTableParams {
	return mdbp.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

func (d PsqlDatabase) CountTables() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateTableTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateTable(ctx context.Context, ac audited.AuditContext, s CreateTableParams) (*Tables, error) {
	cmd := d.NewTableCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	r := d.MapTable(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteTable(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTableCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetTable(id string) (*Tables, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdbp.GetTableParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

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

// NewTableCmd implements audited.CreateCommand[mdb.Tables] for SQLite.
type NewTableCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTableParams
	conn     *sql.DB
}

func (c NewTableCmd) Context() context.Context                    { return c.ctx }
func (c NewTableCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewTableCmd) Connection() *sql.DB                         { return c.conn }
func (c NewTableCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c NewTableCmd) TableName() string                           { return "tables" }
func (c NewTableCmd) Params() any                                 { return c.params }

func (c NewTableCmd) GetID(x mdb.Tables) string {
	return x.ID
}

func (c NewTableCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Tables, error) {
	queries := mdb.New(tx)
	return queries.CreateTable(ctx, mdb.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: c.params.Label,
	})
}

func (d Database) NewTableCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTableParams) NewTableCmd {
	return NewTableCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTableCmd implements audited.UpdateCommand[mdb.Tables] for SQLite.
type UpdateTableCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTableParams
	conn     *sql.DB
}

func (c UpdateTableCmd) Context() context.Context                    { return c.ctx }
func (c UpdateTableCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateTableCmd) Connection() *sql.DB                         { return c.conn }
func (c UpdateTableCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c UpdateTableCmd) TableName() string                           { return "tables" }
func (c UpdateTableCmd) Params() any                                 { return c.params }
func (c UpdateTableCmd) GetID() string                               { return c.params.ID }

func (c UpdateTableCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tables, error) {
	queries := mdb.New(tx)
	return queries.GetTable(ctx, mdb.GetTableParams{ID: c.params.ID})
}

func (c UpdateTableCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateTable(ctx, mdb.UpdateTableParams{
		Label: c.params.Label,
		ID:    c.params.ID,
	})
}

func (d Database) UpdateTableCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTableParams) UpdateTableCmd {
	return UpdateTableCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTableCmd implements audited.DeleteCommand[mdb.Tables] for SQLite.
type DeleteTableCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

func (c DeleteTableCmd) Context() context.Context                    { return c.ctx }
func (c DeleteTableCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteTableCmd) Connection() *sql.DB                         { return c.conn }
func (c DeleteTableCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c DeleteTableCmd) TableName() string                           { return "tables" }
func (c DeleteTableCmd) GetID() string                               { return c.id }

func (c DeleteTableCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tables, error) {
	queries := mdb.New(tx)
	return queries.GetTable(ctx, mdb.GetTableParams{ID: c.id})
}

func (c DeleteTableCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteTable(ctx, mdb.DeleteTableParams{ID: c.id})
}

func (d Database) DeleteTableCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTableCmd {
	return DeleteTableCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewTableCmdMysql implements audited.CreateCommand[mdbm.Tables] for MySQL.
type NewTableCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTableParams
	conn     *sql.DB
}

func (c NewTableCmdMysql) Context() context.Context                    { return c.ctx }
func (c NewTableCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewTableCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c NewTableCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c NewTableCmdMysql) TableName() string                           { return "tables" }
func (c NewTableCmdMysql) Params() any                                 { return c.params }

func (c NewTableCmdMysql) GetID(x mdbm.Tables) string {
	return x.ID
}

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

func (d MysqlDatabase) NewTableCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTableParams) NewTableCmdMysql {
	return NewTableCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTableCmdMysql implements audited.UpdateCommand[mdbm.Tables] for MySQL.
type UpdateTableCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTableParams
	conn     *sql.DB
}

func (c UpdateTableCmdMysql) Context() context.Context                    { return c.ctx }
func (c UpdateTableCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateTableCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c UpdateTableCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c UpdateTableCmdMysql) TableName() string                           { return "tables" }
func (c UpdateTableCmdMysql) Params() any                                 { return c.params }
func (c UpdateTableCmdMysql) GetID() string                               { return c.params.ID }

func (c UpdateTableCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tables, error) {
	queries := mdbm.New(tx)
	return queries.GetTable(ctx, mdbm.GetTableParams{ID: c.params.ID})
}

func (c UpdateTableCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateTable(ctx, mdbm.UpdateTableParams{
		Label: c.params.Label,
		ID:    c.params.ID,
	})
}

func (d MysqlDatabase) UpdateTableCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTableParams) UpdateTableCmdMysql {
	return UpdateTableCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTableCmdMysql implements audited.DeleteCommand[mdbm.Tables] for MySQL.
type DeleteTableCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

func (c DeleteTableCmdMysql) Context() context.Context                    { return c.ctx }
func (c DeleteTableCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteTableCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c DeleteTableCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c DeleteTableCmdMysql) TableName() string                           { return "tables" }
func (c DeleteTableCmdMysql) GetID() string                               { return c.id }

func (c DeleteTableCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tables, error) {
	queries := mdbm.New(tx)
	return queries.GetTable(ctx, mdbm.GetTableParams{ID: c.id})
}

func (c DeleteTableCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteTable(ctx, mdbm.DeleteTableParams{ID: c.id})
}

func (d MysqlDatabase) DeleteTableCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTableCmdMysql {
	return DeleteTableCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewTableCmdPsql implements audited.CreateCommand[mdbp.Tables] for PostgreSQL.
type NewTableCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTableParams
	conn     *sql.DB
}

func (c NewTableCmdPsql) Context() context.Context                    { return c.ctx }
func (c NewTableCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewTableCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c NewTableCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c NewTableCmdPsql) TableName() string                           { return "tables" }
func (c NewTableCmdPsql) Params() any                                 { return c.params }

func (c NewTableCmdPsql) GetID(x mdbp.Tables) string {
	return x.ID
}

func (c NewTableCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Tables, error) {
	queries := mdbp.New(tx)
	return queries.CreateTable(ctx, mdbp.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: c.params.Label,
	})
}

func (d PsqlDatabase) NewTableCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTableParams) NewTableCmdPsql {
	return NewTableCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTableCmdPsql implements audited.UpdateCommand[mdbp.Tables] for PostgreSQL.
type UpdateTableCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTableParams
	conn     *sql.DB
}

func (c UpdateTableCmdPsql) Context() context.Context                    { return c.ctx }
func (c UpdateTableCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateTableCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c UpdateTableCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c UpdateTableCmdPsql) TableName() string                           { return "tables" }
func (c UpdateTableCmdPsql) Params() any                                 { return c.params }
func (c UpdateTableCmdPsql) GetID() string                               { return c.params.ID }

func (c UpdateTableCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tables, error) {
	queries := mdbp.New(tx)
	return queries.GetTable(ctx, mdbp.GetTableParams{ID: c.params.ID})
}

func (c UpdateTableCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateTable(ctx, mdbp.UpdateTableParams{
		Label: c.params.Label,
		ID:    c.params.ID,
	})
}

func (d PsqlDatabase) UpdateTableCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTableParams) UpdateTableCmdPsql {
	return UpdateTableCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTableCmdPsql implements audited.DeleteCommand[mdbp.Tables] for PostgreSQL.
type DeleteTableCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

func (c DeleteTableCmdPsql) Context() context.Context                    { return c.ctx }
func (c DeleteTableCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteTableCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c DeleteTableCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c DeleteTableCmdPsql) TableName() string                           { return "tables" }
func (c DeleteTableCmdPsql) GetID() string                               { return c.id }

func (c DeleteTableCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tables, error) {
	queries := mdbp.New(tx)
	return queries.GetTable(ctx, mdbp.GetTableParams{ID: c.id})
}

func (c DeleteTableCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteTable(ctx, mdbp.DeleteTableParams{ID: c.id})
}

func (d PsqlDatabase) DeleteTableCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTableCmdPsql {
	return DeleteTableCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
