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

type DatatypeFields struct {
	ID         string                   `json:"id"`
	DatatypeID types.NullableDatatypeID `json:"datatype_id"`
	FieldID    types.NullableFieldID    `json:"field_id"`
	SortOrder  int64                    `json:"sort_order"`
}

type CreateDatatypeFieldParams struct {
	ID         string                   `json:"id"`
	DatatypeID types.NullableDatatypeID `json:"datatype_id"`
	FieldID    types.NullableFieldID    `json:"field_id"`
	SortOrder  int64                    `json:"sort_order"`
}

type UpdateDatatypeFieldParams struct {
	DatatypeID types.NullableDatatypeID `json:"datatype_id"`
	FieldID    types.NullableFieldID    `json:"field_id"`
	SortOrder  int64                    `json:"sort_order"`
	ID         string                   `json:"id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringDatatypeField converts DatatypeFields to StringDatatypeFields for table display
func MapStringDatatypeField(a DatatypeFields) StringDatatypeFields {
	return StringDatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID.String(),
		FieldID:    a.FieldID.String(),
		SortOrder:  fmt.Sprintf("%d", a.SortOrder),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapDatatypeField(a mdb.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  a.SortOrder,
	}
}

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

func (d Database) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdb.UpdateDatatypeFieldParams {
	return mdb.UpdateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  a.SortOrder,
		ID:         a.ID,
	}
}

// QUERIES

func (d Database) CountDatatypeFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateDatatypeFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateDatatypesFieldsTable(d.Context)
	return err
}

func (d Database) CreateDatatypeField(s CreateDatatypeFieldParams) DatatypeFields {
	params := d.MapCreateDatatypeFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatypeField: %v\n", err)
	}
	return d.MapDatatypeField(row)
}

func (d Database) DeleteDatatypeField(id string) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteDatatypeField(d.Context, mdb.DeleteDatatypeFieldParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete DatatypeField: %v", id)
	}
	return nil
}

func (d Database) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListDatatypeFieldByDatatypeID(id types.NullableDatatypeID) (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByDatatypeID(d.Context, mdb.ListDatatypeFieldByDatatypeIDParams{DatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListDatatypeFieldByFieldID(id types.NullableFieldID) (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByFieldID(d.Context, mdb.ListDatatypeFieldByFieldIDParams{FieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateDatatypeField(s UpdateDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateDatatypeFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

func (d Database) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdb.UpdateDatatypeFieldSortOrderParams{
		SortOrder: sortOrder,
		ID:        id,
	})
}

func (d Database) GetMaxSortOrderByDatatypeID(datatypeID types.NullableDatatypeID) (int64, error) {
	queries := mdb.New(d.Connection)
	result, err := queries.GetMaxSortOrderByDatatypeID(d.Context, mdb.GetMaxSortOrderByDatatypeIDParams{
		DatatypeID: datatypeID,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to get max sort order: %v", err)
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
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapDatatypeField(a mdbm.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int64(a.SortOrder),
	}
}

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

func (d MysqlDatabase) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdbm.UpdateDatatypeFieldParams {
	return mdbm.UpdateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int32(a.SortOrder),
		ID:         a.ID,
	}
}

// QUERIES

func (d MysqlDatabase) CountDatatypeFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateDatatypeFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypesFieldsTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateDatatypeField(s CreateDatatypeFieldParams) DatatypeFields {
	params := d.MapCreateDatatypeFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatypeField: %v\n", err)
	}
	row, err := queries.GetDatatypeField(d.Context, mdbm.GetDatatypeFieldParams{ID: params.ID})
	if err != nil {
		fmt.Printf("Failed to get last inserted DatatypeField: %v\n", err)
	}
	return d.MapDatatypeField(row)
}

func (d MysqlDatabase) DeleteDatatypeField(id string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteDatatypeField(d.Context, mdbm.DeleteDatatypeFieldParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete DatatypeField: %v", id)
	}
	return nil
}

func (d MysqlDatabase) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListDatatypeFieldByFieldID(id types.NullableFieldID) (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByFieldID(d.Context, mdbm.ListDatatypeFieldByFieldIDParams{FieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListDatatypeFieldByDatatypeID(id types.NullableDatatypeID) (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByDatatypeID(d.Context, mdbm.ListDatatypeFieldByDatatypeIDParams{DatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateDatatypeField(s UpdateDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateDatatypeFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

func (d MysqlDatabase) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdbm.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(sortOrder),
		ID:        id,
	})
}

func (d MysqlDatabase) GetMaxSortOrderByDatatypeID(datatypeID types.NullableDatatypeID) (int64, error) {
	queries := mdbm.New(d.Connection)
	result, err := queries.GetMaxSortOrderByDatatypeID(d.Context, mdbm.GetMaxSortOrderByDatatypeIDParams{
		DatatypeID: datatypeID,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to get max sort order: %v", err)
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
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapDatatypeField(a mdbp.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int64(a.SortOrder),
	}
}

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

func (d PsqlDatabase) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdbp.UpdateDatatypeFieldParams {
	return mdbp.UpdateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		SortOrder:  int32(a.SortOrder),
		ID:         a.ID,
	}
}

// QUERIES

func (d PsqlDatabase) CountDatatypeFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateDatatypeFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateDatatypesFieldsTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateDatatypeField(s CreateDatatypeFieldParams) DatatypeFields {
	params := d.MapCreateDatatypeFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatypeField: %v\n", err)
	}
	return d.MapDatatypeField(row)
}

func (d PsqlDatabase) DeleteDatatypeField(id string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteDatatypeField(d.Context, mdbp.DeleteDatatypeFieldParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete DatatypeField: %v", id)
	}
	return nil
}

func (d PsqlDatabase) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListDatatypeFieldByDatatypeID(id types.NullableDatatypeID) (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByDatatypeID(d.Context, mdbp.ListDatatypeFieldByDatatypeIDParams{DatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListDatatypeFieldByFieldID(id types.NullableFieldID) (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByFieldID(d.Context, mdbp.ListDatatypeFieldByFieldIDParams{FieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateDatatypeField(s UpdateDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateDatatypeFieldParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

func (d PsqlDatabase) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdbp.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(sortOrder),
		ID:        id,
	})
}

func (d PsqlDatabase) GetMaxSortOrderByDatatypeID(datatypeID types.NullableDatatypeID) (int64, error) {
	queries := mdbp.New(d.Connection)
	result, err := queries.GetMaxSortOrderByDatatypeID(d.Context, mdbp.GetMaxSortOrderByDatatypeIDParams{
		DatatypeID: datatypeID,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to get max sort order: %v", err)
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
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewDatatypeFieldCmd is an audited create command for datatypes_fields (SQLite).
// Note: Update and Delete commands are not implemented because no GetDatatypeField
// query exists in the sqlc-generated code, which is required for GetBefore.
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

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

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

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

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
