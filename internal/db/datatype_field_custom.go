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
// SQLITE — MAPPERS
//////////////////////////////

// MapDatatypeField converts a sqlc-generated SQLite type to the wrapper type.
func (d Database) MapDatatypeField(a mdb.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  a.SortOrder,
	}
}

// MapCreateDatatypeFieldParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateDatatypeFieldParams(a CreateDatatypeFieldParams) mdb.CreateDatatypeFieldParams {
	id := a.ID
	if id == "" {
		id = string(types.NewDatatypeFieldID())
	}
	return mdb.CreateDatatypeFieldParams{
		ID:         id,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  a.SortOrder,
	}
}

// MapUpdateDatatypeFieldParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdb.UpdateDatatypeFieldParams {
	return mdb.UpdateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  a.SortOrder,
		ID:         a.ID,
	}
}

///////////////////////////////
// SQLITE — AUDITED COMMANDS
//////////////////////////////

// ----- SQLite CREATE -----

// NewDatatypeFieldCmd is an audited create command for datatypes_fields (SQLite).
type NewDatatypeFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewDatatypeFieldCmd) Context() context.Context              { return c.ctx }
func (c NewDatatypeFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewDatatypeFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c NewDatatypeFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewDatatypeFieldCmd) TableName() string                     { return "datatypes_fields" }
func (c NewDatatypeFieldCmd) Params() any                           { return c.params }
func (c NewDatatypeFieldCmd) GetID(row mdb.DatatypesFields) string  { return row.ID }

func (c NewDatatypeFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.DatatypesFields, error) {
	id := c.params.ID
	if id == "" {
		id = string(types.NewDatatypeFieldID())
	}
	queries := mdb.New(tx)
	return queries.CreateDatatypeField(ctx, mdb.CreateDatatypeFieldParams{
		ID:         id,
		DatatypeID: c.params.DatatypeID,
		FieldID:    c.params.FieldID,
		SortOrder:  c.params.SortOrder,
	})
}

func (d Database) NewDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeFieldParams) NewDatatypeFieldCmd {
	return NewDatatypeFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateDatatypeFieldCmd is an audited update command for datatypes_fields (SQLite).
type UpdateDatatypeFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeFieldCmd) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeFieldCmd) TableName() string                     { return "datatypes_fields" }
func (c UpdateDatatypeFieldCmd) Params() any                           { return c.params }
func (c UpdateDatatypeFieldCmd) GetID() string                         { return c.params.ID }

func (c UpdateDatatypeFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.DatatypesFields, error) {
	queries := mdb.New(tx)
	rows, err := queries.ListDatatypeField(ctx)
	if err != nil {
		return mdb.DatatypesFields{}, fmt.Errorf("list datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.params.ID {
			return v, nil
		}
	}
	return mdb.DatatypesFields{}, fmt.Errorf("datatypes_fields not found: %v", c.params.ID)
}

func (c UpdateDatatypeFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateDatatypeField(ctx, mdb.UpdateDatatypeFieldParams{
		DatatypeID: c.params.DatatypeID,
		FieldID:    c.params.FieldID,
		SortOrder:  c.params.SortOrder,
		ID:         c.params.ID,
	})
}

func (d Database) UpdateDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeFieldParams) UpdateDatatypeFieldCmd {
	return UpdateDatatypeFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE SORT ORDER -----

// UpdateDatatypeFieldSortOrderCmd is an audited update command for sort order of datatypes_fields (SQLite).
type UpdateDatatypeFieldSortOrderCmd struct {
	ctx       context.Context
	auditCtx  audited.AuditContext
	id        string
	sortOrder int64
	conn      *sql.DB
	recorder  audited.ChangeEventRecorder
}

func (c UpdateDatatypeFieldSortOrderCmd) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeFieldSortOrderCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeFieldSortOrderCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeFieldSortOrderCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeFieldSortOrderCmd) TableName() string                     { return "datatypes_fields" }
func (c UpdateDatatypeFieldSortOrderCmd) GetID() string                         { return c.id }

func (c UpdateDatatypeFieldSortOrderCmd) Params() any {
	return map[string]any{"id": c.id, "sort_order": c.sortOrder}
}

func (c UpdateDatatypeFieldSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.DatatypesFields, error) {
	queries := mdb.New(tx)
	rows, err := queries.ListDatatypeField(ctx)
	if err != nil {
		return mdb.DatatypesFields{}, fmt.Errorf("list datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.id {
			return v, nil
		}
	}
	return mdb.DatatypesFields{}, fmt.Errorf("datatypes_fields not found: %v", c.id)
}

func (c UpdateDatatypeFieldSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateDatatypeFieldSortOrder(ctx, mdb.UpdateDatatypeFieldSortOrderParams{
		SortOrder: c.sortOrder,
		ID:        c.id,
	})
}

func (d Database) UpdateDatatypeFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, id string, sortOrder int64) UpdateDatatypeFieldSortOrderCmd {
	return UpdateDatatypeFieldSortOrderCmd{ctx: ctx, auditCtx: auditCtx, id: id, sortOrder: sortOrder, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteDatatypeFieldCmd is an audited delete command for datatypes_fields (SQLite).
type DeleteDatatypeFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteDatatypeFieldCmd) Context() context.Context              { return c.ctx }
func (c DeleteDatatypeFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteDatatypeFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteDatatypeFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteDatatypeFieldCmd) TableName() string                     { return "datatypes_fields" }
func (c DeleteDatatypeFieldCmd) GetID() string                         { return c.id }

func (c DeleteDatatypeFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.DatatypesFields, error) {
	queries := mdb.New(tx)
	rows, err := queries.ListDatatypeField(ctx)
	if err != nil {
		return mdb.DatatypesFields{}, fmt.Errorf("list datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.id {
			return v, nil
		}
	}
	return mdb.DatatypesFields{}, fmt.Errorf("datatypes_fields not found: %v", c.id)
}

func (c DeleteDatatypeFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteDatatypeField(ctx, mdb.DeleteDatatypeFieldParams{ID: c.id})
}

func (d Database) DeleteDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteDatatypeFieldCmd {
	return DeleteDatatypeFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// SQLITE — CRUD METHODS
//////////////////////////////

// CreateDatatypeField creates a new audited datatype field.
func (d Database) CreateDatatypeField(ctx context.Context, ac audited.AuditContext, s CreateDatatypeFieldParams) (*DatatypeFields, error) {
	cmd := d.NewDatatypeFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatypeField: %w", err)
	}
	r := d.MapDatatypeField(result)
	return &r, nil
}

// UpdateDatatypeField updates an audited datatype field.
func (d Database) UpdateDatatypeField(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeFieldParams) (*string, error) {
	cmd := d.UpdateDatatypeFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatypeField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

// UpdateDatatypeFieldSortOrder updates the sort order for a datatype field.
func (d Database) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {
	cmd := d.UpdateDatatypeFieldSortOrderCmd(ctx, ac, id, sortOrder)
	return audited.Update(cmd)
}

// DeleteDatatypeField deletes an audited datatype field by ID.
func (d Database) DeleteDatatypeField(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteDatatypeFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetMaxSortOrderByDatatypeID returns the maximum sort order for fields in a datatype.
func (d Database) GetMaxSortOrderByDatatypeID(datatypeID types.DatatypeID) (int64, error) {
	queries := mdb.New(d.Connection)
	result, err := queries.GetMaxSortOrderByDatatypeID(d.Context, mdb.GetMaxSortOrderByDatatypeIDParams{
		DatatypeID: datatypeID,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to get max sort order: %w", err)
	}
	switch v := result.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return -1, nil
	}
}

///////////////////////////////
// MYSQL — MAPPERS
//////////////////////////////

// MapDatatypeField converts a sqlc-generated MySQL type to the wrapper type.
func (d MysqlDatabase) MapDatatypeField(a mdbm.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int64(a.SortOrder),
	}
}

// MapCreateDatatypeFieldParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateDatatypeFieldParams(a CreateDatatypeFieldParams) mdbm.CreateDatatypeFieldParams {
	id := a.ID
	if id == "" {
		id = string(types.NewDatatypeFieldID())
	}
	return mdbm.CreateDatatypeFieldParams{
		ID:         id,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int32(a.SortOrder),
	}
}

// MapUpdateDatatypeFieldParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdbm.UpdateDatatypeFieldParams {
	return mdbm.UpdateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int32(a.SortOrder),
		ID:         a.ID,
	}
}

///////////////////////////////
// MYSQL — AUDITED COMMANDS
//////////////////////////////

// ----- MySQL CREATE -----

// NewDatatypeFieldCmdMysql is an audited create command for datatypes_fields (MySQL).
type NewDatatypeFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewDatatypeFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c NewDatatypeFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewDatatypeFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewDatatypeFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewDatatypeFieldCmdMysql) TableName() string                     { return "datatypes_fields" }
func (c NewDatatypeFieldCmdMysql) Params() any                           { return c.params }
func (c NewDatatypeFieldCmdMysql) GetID(row mdbm.DatatypesFields) string { return row.ID }

func (c NewDatatypeFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.DatatypesFields, error) {
	id := c.params.ID
	if id == "" {
		id = string(types.NewDatatypeFieldID())
	}
	queries := mdbm.New(tx)
	err := queries.CreateDatatypeField(ctx, mdbm.CreateDatatypeFieldParams{
		ID:         id,
		DatatypeID: c.params.DatatypeID,
		FieldID:    c.params.FieldID,
		SortOrder:  int32(c.params.SortOrder),
	})
	if err != nil {
		return mdbm.DatatypesFields{}, fmt.Errorf("execute create datatypes_fields: %w", err)
	}
	return queries.GetDatatypeField(ctx, mdbm.GetDatatypeFieldParams{ID: id})
}

func (d MysqlDatabase) NewDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeFieldParams) NewDatatypeFieldCmdMysql {
	return NewDatatypeFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateDatatypeFieldCmdMysql is an audited update command for datatypes_fields (MySQL).
type UpdateDatatypeFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeFieldCmdMysql) TableName() string                     { return "datatypes_fields" }
func (c UpdateDatatypeFieldCmdMysql) Params() any                           { return c.params }
func (c UpdateDatatypeFieldCmdMysql) GetID() string                         { return c.params.ID }

func (c UpdateDatatypeFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.DatatypesFields, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatypeField(ctx, mdbm.GetDatatypeFieldParams{ID: c.params.ID})
}

func (c UpdateDatatypeFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateDatatypeField(ctx, mdbm.UpdateDatatypeFieldParams{
		DatatypeID: c.params.DatatypeID,
		FieldID:    c.params.FieldID,
		SortOrder:  int32(c.params.SortOrder),
		ID:         c.params.ID,
	})
}

func (d MysqlDatabase) UpdateDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeFieldParams) UpdateDatatypeFieldCmdMysql {
	return UpdateDatatypeFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE SORT ORDER -----

// UpdateDatatypeFieldSortOrderCmdMysql is an audited update command for sort order of datatypes_fields (MySQL).
type UpdateDatatypeFieldSortOrderCmdMysql struct {
	ctx       context.Context
	auditCtx  audited.AuditContext
	id        string
	sortOrder int64
	conn      *sql.DB
	recorder  audited.ChangeEventRecorder
}

func (c UpdateDatatypeFieldSortOrderCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeFieldSortOrderCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeFieldSortOrderCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeFieldSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeFieldSortOrderCmdMysql) TableName() string                     { return "datatypes_fields" }
func (c UpdateDatatypeFieldSortOrderCmdMysql) GetID() string                         { return c.id }

func (c UpdateDatatypeFieldSortOrderCmdMysql) Params() any {
	return map[string]any{"id": c.id, "sort_order": c.sortOrder}
}

func (c UpdateDatatypeFieldSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.DatatypesFields, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatypeField(ctx, mdbm.GetDatatypeFieldParams{ID: c.id})
}

func (c UpdateDatatypeFieldSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateDatatypeFieldSortOrder(ctx, mdbm.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(c.sortOrder),
		ID:        c.id,
	})
}

func (d MysqlDatabase) UpdateDatatypeFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, id string, sortOrder int64) UpdateDatatypeFieldSortOrderCmdMysql {
	return UpdateDatatypeFieldSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, sortOrder: sortOrder, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteDatatypeFieldCmdMysql is an audited delete command for datatypes_fields (MySQL).
type DeleteDatatypeFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteDatatypeFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteDatatypeFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteDatatypeFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteDatatypeFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteDatatypeFieldCmdMysql) TableName() string                     { return "datatypes_fields" }
func (c DeleteDatatypeFieldCmdMysql) GetID() string                         { return c.id }

func (c DeleteDatatypeFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.DatatypesFields, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatypeField(ctx, mdbm.GetDatatypeFieldParams{ID: c.id})
}

func (c DeleteDatatypeFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteDatatypeField(ctx, mdbm.DeleteDatatypeFieldParams{ID: c.id})
}

func (d MysqlDatabase) DeleteDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteDatatypeFieldCmdMysql {
	return DeleteDatatypeFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// MYSQL — CRUD METHODS
//////////////////////////////

// CreateDatatypeField creates a new audited datatype field.
func (d MysqlDatabase) CreateDatatypeField(ctx context.Context, ac audited.AuditContext, s CreateDatatypeFieldParams) (*DatatypeFields, error) {
	cmd := d.NewDatatypeFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatypeField: %w", err)
	}
	r := d.MapDatatypeField(result)
	return &r, nil
}

// UpdateDatatypeField updates an audited datatype field.
func (d MysqlDatabase) UpdateDatatypeField(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeFieldParams) (*string, error) {
	cmd := d.UpdateDatatypeFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatypeField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

// UpdateDatatypeFieldSortOrder updates the sort order for a datatype field.
func (d MysqlDatabase) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {
	cmd := d.UpdateDatatypeFieldSortOrderCmd(ctx, ac, id, sortOrder)
	return audited.Update(cmd)
}

// DeleteDatatypeField deletes an audited datatype field by ID.
func (d MysqlDatabase) DeleteDatatypeField(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteDatatypeFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetMaxSortOrderByDatatypeID returns the maximum sort order for fields in a datatype.
func (d MysqlDatabase) GetMaxSortOrderByDatatypeID(datatypeID types.DatatypeID) (int64, error) {
	queries := mdbm.New(d.Connection)
	result, err := queries.GetMaxSortOrderByDatatypeID(d.Context, mdbm.GetMaxSortOrderByDatatypeIDParams{
		DatatypeID: datatypeID,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to get max sort order: %w", err)
	}
	switch v := result.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return -1, nil
	}
}

///////////////////////////////
// POSTGRES — MAPPERS
//////////////////////////////

// MapDatatypeField converts a sqlc-generated PostgreSQL type to the wrapper type.
func (d PsqlDatabase) MapDatatypeField(a mdbp.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int64(a.SortOrder),
	}
}

// MapCreateDatatypeFieldParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateDatatypeFieldParams(a CreateDatatypeFieldParams) mdbp.CreateDatatypeFieldParams {
	id := a.ID
	if id == "" {
		id = string(types.NewDatatypeFieldID())
	}
	return mdbp.CreateDatatypeFieldParams{
		ID:         id,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int32(a.SortOrder),
	}
}

// MapUpdateDatatypeFieldParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdbp.UpdateDatatypeFieldParams {
	return mdbp.UpdateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int32(a.SortOrder),
		ID:         a.ID,
	}
}

///////////////////////////////
// POSTGRES — AUDITED COMMANDS
//////////////////////////////

// ----- PostgreSQL CREATE -----

// NewDatatypeFieldCmdPsql is an audited create command for datatypes_fields (PostgreSQL).
type NewDatatypeFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewDatatypeFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c NewDatatypeFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewDatatypeFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewDatatypeFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewDatatypeFieldCmdPsql) TableName() string                     { return "datatypes_fields" }
func (c NewDatatypeFieldCmdPsql) Params() any                           { return c.params }
func (c NewDatatypeFieldCmdPsql) GetID(row mdbp.DatatypesFields) string { return row.ID }

func (c NewDatatypeFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.DatatypesFields, error) {
	id := c.params.ID
	if id == "" {
		id = string(types.NewDatatypeFieldID())
	}
	queries := mdbp.New(tx)
	return queries.CreateDatatypeField(ctx, mdbp.CreateDatatypeFieldParams{
		ID:         id,
		DatatypeID: c.params.DatatypeID,
		FieldID:    c.params.FieldID,
		SortOrder:  int32(c.params.SortOrder),
	})
}

func (d PsqlDatabase) NewDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeFieldParams) NewDatatypeFieldCmdPsql {
	return NewDatatypeFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateDatatypeFieldCmdPsql is an audited update command for datatypes_fields (PostgreSQL).
type UpdateDatatypeFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeFieldCmdPsql) TableName() string                     { return "datatypes_fields" }
func (c UpdateDatatypeFieldCmdPsql) Params() any                           { return c.params }
func (c UpdateDatatypeFieldCmdPsql) GetID() string                         { return c.params.ID }

func (c UpdateDatatypeFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.DatatypesFields, error) {
	queries := mdbp.New(tx)
	rows, err := queries.ListDatatypeField(ctx)
	if err != nil {
		return mdbp.DatatypesFields{}, fmt.Errorf("list datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.params.ID {
			return v, nil
		}
	}
	return mdbp.DatatypesFields{}, fmt.Errorf("datatypes_fields not found: %v", c.params.ID)
}

func (c UpdateDatatypeFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateDatatypeField(ctx, mdbp.UpdateDatatypeFieldParams{
		DatatypeID: c.params.DatatypeID,
		FieldID:    c.params.FieldID,
		SortOrder:  int32(c.params.SortOrder),
		ID:         c.params.ID,
	})
}

func (d PsqlDatabase) UpdateDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeFieldParams) UpdateDatatypeFieldCmdPsql {
	return UpdateDatatypeFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE SORT ORDER -----

// UpdateDatatypeFieldSortOrderCmdPsql is an audited update command for sort order of datatypes_fields (PostgreSQL).
type UpdateDatatypeFieldSortOrderCmdPsql struct {
	ctx       context.Context
	auditCtx  audited.AuditContext
	id        string
	sortOrder int64
	conn      *sql.DB
	recorder  audited.ChangeEventRecorder
}

func (c UpdateDatatypeFieldSortOrderCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeFieldSortOrderCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeFieldSortOrderCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeFieldSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeFieldSortOrderCmdPsql) TableName() string                     { return "datatypes_fields" }
func (c UpdateDatatypeFieldSortOrderCmdPsql) GetID() string                         { return c.id }

func (c UpdateDatatypeFieldSortOrderCmdPsql) Params() any {
	return map[string]any{"id": c.id, "sort_order": c.sortOrder}
}

func (c UpdateDatatypeFieldSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.DatatypesFields, error) {
	queries := mdbp.New(tx)
	rows, err := queries.ListDatatypeField(ctx)
	if err != nil {
		return mdbp.DatatypesFields{}, fmt.Errorf("list datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.id {
			return v, nil
		}
	}
	return mdbp.DatatypesFields{}, fmt.Errorf("datatypes_fields not found: %v", c.id)
}

func (c UpdateDatatypeFieldSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateDatatypeFieldSortOrder(ctx, mdbp.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(c.sortOrder),
		ID:        c.id,
	})
}

func (d PsqlDatabase) UpdateDatatypeFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, id string, sortOrder int64) UpdateDatatypeFieldSortOrderCmdPsql {
	return UpdateDatatypeFieldSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, sortOrder: sortOrder, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteDatatypeFieldCmdPsql is an audited delete command for datatypes_fields (PostgreSQL).
type DeleteDatatypeFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteDatatypeFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteDatatypeFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteDatatypeFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteDatatypeFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteDatatypeFieldCmdPsql) TableName() string                     { return "datatypes_fields" }
func (c DeleteDatatypeFieldCmdPsql) GetID() string                         { return c.id }

func (c DeleteDatatypeFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.DatatypesFields, error) {
	queries := mdbp.New(tx)
	rows, err := queries.ListDatatypeField(ctx)
	if err != nil {
		return mdbp.DatatypesFields{}, fmt.Errorf("list datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.id {
			return v, nil
		}
	}
	return mdbp.DatatypesFields{}, fmt.Errorf("datatypes_fields not found: %v", c.id)
}

func (c DeleteDatatypeFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteDatatypeField(ctx, mdbp.DeleteDatatypeFieldParams{ID: c.id})
}

func (d PsqlDatabase) DeleteDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteDatatypeFieldCmdPsql {
	return DeleteDatatypeFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

///////////////////////////////
// POSTGRES — CRUD METHODS
//////////////////////////////

// CreateDatatypeField creates a new audited datatype field.
func (d PsqlDatabase) CreateDatatypeField(ctx context.Context, ac audited.AuditContext, s CreateDatatypeFieldParams) (*DatatypeFields, error) {
	cmd := d.NewDatatypeFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatypeField: %w", err)
	}
	r := d.MapDatatypeField(result)
	return &r, nil
}

// UpdateDatatypeField updates an audited datatype field.
func (d PsqlDatabase) UpdateDatatypeField(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeFieldParams) (*string, error) {
	cmd := d.UpdateDatatypeFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatypeField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

// UpdateDatatypeFieldSortOrder updates the sort order for a datatype field.
func (d PsqlDatabase) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {
	cmd := d.UpdateDatatypeFieldSortOrderCmd(ctx, ac, id, sortOrder)
	return audited.Update(cmd)
}

// DeleteDatatypeField deletes an audited datatype field by ID.
func (d PsqlDatabase) DeleteDatatypeField(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteDatatypeFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetMaxSortOrderByDatatypeID returns the maximum sort order for fields in a datatype.
func (d PsqlDatabase) GetMaxSortOrderByDatatypeID(datatypeID types.DatatypeID) (int64, error) {
	queries := mdbp.New(d.Connection)
	result, err := queries.GetMaxSortOrderByDatatypeID(d.Context, mdbp.GetMaxSortOrderByDatatypeIDParams{
		DatatypeID: datatypeID,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to get max sort order: %w", err)
	}
	switch v := result.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return -1, nil
	}
}

///////////////////////////////
// LIST (SINGULAR ALIAS) + PAGINATED
//////////////////////////////

// ListDatatypeField returns all datatypes_fields records (SQLite).
// This is a singular alias required by the DbDriver interface.
func (d Database) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %w", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeField returns all datatypes_fields records (MySQL).
func (d MysqlDatabase) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %w", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeField returns all datatypes_fields records (PostgreSQL).
func (d PsqlDatabase) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %w", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeFieldPaginated returns datatype fields with pagination (SQLite).
func (d Database) ListDatatypeFieldPaginated(params PaginationParams) (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeFieldPaginated(d.Context, mdb.ListDatatypeFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields paginated: %w", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeFieldPaginated returns datatype fields with pagination (MySQL).
func (d MysqlDatabase) ListDatatypeFieldPaginated(params PaginationParams) (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeFieldPaginated(d.Context, mdbm.ListDatatypeFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields paginated: %w", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeFieldPaginated returns datatype fields with pagination (PostgreSQL).
func (d PsqlDatabase) ListDatatypeFieldPaginated(params PaginationParams) (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeFieldPaginated(d.Context, mdbp.ListDatatypeFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields paginated: %w", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}
