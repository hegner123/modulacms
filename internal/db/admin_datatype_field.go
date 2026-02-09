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

type AdminDatatypeFields struct {
	ID              string                `json:"id"`
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	AdminFieldID    types.AdminFieldID    `json:"admin_field_id"`
}

type CreateAdminDatatypeFieldParams struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	AdminFieldID    types.AdminFieldID    `json:"admin_field_id"`
}

type UpdateAdminDatatypeFieldParams struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	AdminFieldID    types.AdminFieldID    `json:"admin_field_id"`
	ID              string                `json:"id"`
}

// MapStringAdminDatatypeField converts AdminDatatypeFields to StringAdminDatatypeFields for display purposes
func MapStringAdminDatatypeField(a AdminDatatypeFields) StringAdminDatatypeFields {
	return StringAdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID.String(),
		AdminFieldID:    a.AdminFieldID.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminDatatypeField(a mdb.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d Database) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdb.CreateAdminDatatypeFieldParams {
	return mdb.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d Database) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdb.UpdateAdminDatatypeFieldParams {
	return mdb.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
		ID:              a.ID,
	}
}

// QUERIES

func (d Database) CountAdminDatatypeFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminDatatypeFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminDatatypesFieldsTable(d.Context)
	return err
}

func (d Database) CreateAdminDatatypeField(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeFieldParams) (*AdminDatatypeFields, error) {
	cmd := d.NewAdminDatatypeFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatypeField: %w", err)
	}
	r := d.MapAdminDatatypeField(result)
	return &r, nil
}

func (d Database) DeleteAdminDatatypeField(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteAdminDatatypeFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetAdminDatatypeField(id string) (*AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeField: %v", err)
	}
	for _, v := range rows {
		if v.ID == id {
			m := d.MapAdminDatatypeField(v)
			return &m, nil
		}
	}
	return nil, fmt.Errorf("AdminDatatypeField not found: %v", id)
}

func (d Database) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypeFieldByDatatypeID(id types.AdminDatatypeID) (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, mdb.ListAdminDatatypeFieldByDatatypeIDParams{AdminDatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypeFieldByFieldID(id types.AdminFieldID) (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, mdb.ListAdminDatatypeFieldByFieldIDParams{AdminFieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminDatatypeField(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeFieldParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatypeField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminDatatypeField(a mdbm.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d MysqlDatabase) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdbm.CreateAdminDatatypeFieldParams {
	return mdbm.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d MysqlDatabase) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdbm.UpdateAdminDatatypeFieldParams {
	return mdbm.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
		ID:              a.ID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminDatatypeFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminDatatypeFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatypesFieldsTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateAdminDatatypeField(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeFieldParams) (*AdminDatatypeFields, error) {
	cmd := d.NewAdminDatatypeFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatypeField: %w", err)
	}
	r := d.MapAdminDatatypeField(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteAdminDatatypeField(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteAdminDatatypeFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetAdminDatatypeField(id string) (*AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeField: %v", err)
	}
	for _, v := range rows {
		if v.ID == id {
			m := d.MapAdminDatatypeField(v)
			return &m, nil
		}
	}
	return nil, fmt.Errorf("AdminDatatypeField not found: %v", id)
}

func (d MysqlDatabase) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypeFieldByFieldID(id types.AdminFieldID) (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, mdbm.ListAdminDatatypeFieldByFieldIDParams{AdminFieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypeFieldByDatatypeID(id types.AdminDatatypeID) (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, mdbm.ListAdminDatatypeFieldByDatatypeIDParams{AdminDatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateAdminDatatypeField(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeFieldParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatypeField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapAdminDatatypeField(a mdbp.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d PsqlDatabase) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdbp.CreateAdminDatatypeFieldParams {
	return mdbp.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d PsqlDatabase) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdbp.UpdateAdminDatatypeFieldParams {
	return mdbp.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
		ID:              a.ID,
	}
}

// QUERIES

func (d PsqlDatabase) CountAdminDatatypeFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminDatatypeFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminDatatypesFieldsTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateAdminDatatypeField(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeFieldParams) (*AdminDatatypeFields, error) {
	cmd := d.NewAdminDatatypeFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatypeField: %w", err)
	}
	r := d.MapAdminDatatypeField(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteAdminDatatypeField(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteAdminDatatypeFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetAdminDatatypeField(id string) (*AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeField: %v", err)
	}
	for _, v := range rows {
		if v.ID == id {
			m := d.MapAdminDatatypeField(v)
			return &m, nil
		}
	}
	return nil, fmt.Errorf("AdminDatatypeField not found: %v", id)
}

func (d PsqlDatabase) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypeFieldByDatatypeID(id types.AdminDatatypeID) (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, mdbp.ListAdminDatatypeFieldByDatatypeIDParams{AdminDatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypeFieldByFieldID(id types.AdminFieldID) (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, mdbp.ListAdminDatatypeFieldByFieldIDParams{AdminFieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminDatatypeField(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeFieldParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatypeField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewAdminDatatypeFieldCmd is an audited create command for admin_datatypes_fields (SQLite).
type NewAdminDatatypeFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminDatatypeFieldCmd) Context() context.Context              { return c.ctx }
func (c NewAdminDatatypeFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminDatatypeFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c NewAdminDatatypeFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminDatatypeFieldCmd) TableName() string                     { return "admin_datatypes_fields" }
func (c NewAdminDatatypeFieldCmd) Params() any                           { return c.params }
func (c NewAdminDatatypeFieldCmd) GetID(row mdb.AdminDatatypesFields) string { return row.ID }

func (c NewAdminDatatypeFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypesFields, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminDatatypeField(ctx, mdb.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: c.params.AdminDatatypeID,
		AdminFieldID:    c.params.AdminFieldID,
	})
}

func (d Database) NewAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeFieldParams) NewAdminDatatypeFieldCmd {
	return NewAdminDatatypeFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

type UpdateAdminDatatypeFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeFieldCmd) Context() context.Context              { return c.ctx }
func (c UpdateAdminDatatypeFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminDatatypeFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminDatatypeFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminDatatypeFieldCmd) TableName() string                     { return "admin_datatypes_fields" }
func (c UpdateAdminDatatypeFieldCmd) Params() any                           { return c.params }
func (c UpdateAdminDatatypeFieldCmd) GetID() string                         { return c.params.ID }

func (c UpdateAdminDatatypeFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypesFields, error) {
	queries := mdb.New(tx)
	rows, err := queries.ListAdminDatatypeField(ctx)
	if err != nil {
		return mdb.AdminDatatypesFields{}, fmt.Errorf("list admin_datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.params.ID {
			return v, nil
		}
	}
	return mdb.AdminDatatypesFields{}, fmt.Errorf("admin_datatypes_fields not found: %v", c.params.ID)
}

func (c UpdateAdminDatatypeFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminDatatypeField(ctx, mdb.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: c.params.AdminDatatypeID,
		AdminFieldID:    c.params.AdminFieldID,
		ID:              c.params.ID,
	})
}

func (d Database) UpdateAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeFieldParams) UpdateAdminDatatypeFieldCmd {
	return UpdateAdminDatatypeFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteAdminDatatypeFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminDatatypeFieldCmd) Context() context.Context              { return c.ctx }
func (c DeleteAdminDatatypeFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminDatatypeFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminDatatypeFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminDatatypeFieldCmd) TableName() string                     { return "admin_datatypes_fields" }
func (c DeleteAdminDatatypeFieldCmd) GetID() string                         { return c.id }

func (c DeleteAdminDatatypeFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypesFields, error) {
	queries := mdb.New(tx)
	rows, err := queries.ListAdminDatatypeField(ctx)
	if err != nil {
		return mdb.AdminDatatypesFields{}, fmt.Errorf("list admin_datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.id {
			return v, nil
		}
	}
	return mdb.AdminDatatypesFields{}, fmt.Errorf("admin_datatypes_fields not found: %v", c.id)
}

func (c DeleteAdminDatatypeFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminDatatypeField(ctx, mdb.DeleteAdminDatatypeFieldParams{ID: c.id})
}

func (d Database) DeleteAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteAdminDatatypeFieldCmd {
	return DeleteAdminDatatypeFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

// NewAdminDatatypeFieldCmdMysql is an audited create command for admin_datatypes_fields (MySQL).
type NewAdminDatatypeFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminDatatypeFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c NewAdminDatatypeFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminDatatypeFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminDatatypeFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminDatatypeFieldCmdMysql) TableName() string                     { return "admin_datatypes_fields" }
func (c NewAdminDatatypeFieldCmdMysql) Params() any                           { return c.params }
func (c NewAdminDatatypeFieldCmdMysql) GetID(row mdbm.AdminDatatypesFields) string { return row.ID }

func (c NewAdminDatatypeFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypesFields, error) {
	id := string(types.NewAdminDatatypeFieldID())
	queries := mdbm.New(tx)
	err := queries.CreateAdminDatatypeField(ctx, mdbm.CreateAdminDatatypeFieldParams{
		ID:              id,
		AdminDatatypeID: c.params.AdminDatatypeID,
		AdminFieldID:    c.params.AdminFieldID,
	})
	if err != nil {
		return mdbm.AdminDatatypesFields{}, fmt.Errorf("execute create admin_datatypes_fields: %w", err)
	}
	return queries.GetAdminDatatypeField(ctx, mdbm.GetAdminDatatypeFieldParams{ID: id})
}

func (d MysqlDatabase) NewAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeFieldParams) NewAdminDatatypeFieldCmdMysql {
	return NewAdminDatatypeFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

type UpdateAdminDatatypeFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateAdminDatatypeFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminDatatypeFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminDatatypeFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminDatatypeFieldCmdMysql) TableName() string                     { return "admin_datatypes_fields" }
func (c UpdateAdminDatatypeFieldCmdMysql) Params() any                           { return c.params }
func (c UpdateAdminDatatypeFieldCmdMysql) GetID() string                         { return c.params.ID }

func (c UpdateAdminDatatypeFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypesFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminDatatypeField(ctx, mdbm.GetAdminDatatypeFieldParams{ID: c.params.ID})
}

func (c UpdateAdminDatatypeFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminDatatypeField(ctx, mdbm.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: c.params.AdminDatatypeID,
		AdminFieldID:    c.params.AdminFieldID,
		ID:              c.params.ID,
	})
}

func (d MysqlDatabase) UpdateAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeFieldParams) UpdateAdminDatatypeFieldCmdMysql {
	return UpdateAdminDatatypeFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteAdminDatatypeFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminDatatypeFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteAdminDatatypeFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminDatatypeFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminDatatypeFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminDatatypeFieldCmdMysql) TableName() string                     { return "admin_datatypes_fields" }
func (c DeleteAdminDatatypeFieldCmdMysql) GetID() string                         { return c.id }

func (c DeleteAdminDatatypeFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypesFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminDatatypeField(ctx, mdbm.GetAdminDatatypeFieldParams{ID: c.id})
}

func (c DeleteAdminDatatypeFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminDatatypeField(ctx, mdbm.DeleteAdminDatatypeFieldParams{ID: c.id})
}

func (d MysqlDatabase) DeleteAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteAdminDatatypeFieldCmdMysql {
	return DeleteAdminDatatypeFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

// NewAdminDatatypeFieldCmdPsql is an audited create command for admin_datatypes_fields (PostgreSQL).
type NewAdminDatatypeFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminDatatypeFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c NewAdminDatatypeFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminDatatypeFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminDatatypeFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminDatatypeFieldCmdPsql) TableName() string                     { return "admin_datatypes_fields" }
func (c NewAdminDatatypeFieldCmdPsql) Params() any                           { return c.params }
func (c NewAdminDatatypeFieldCmdPsql) GetID(row mdbp.AdminDatatypesFields) string { return row.ID }

func (c NewAdminDatatypeFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypesFields, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminDatatypeField(ctx, mdbp.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: c.params.AdminDatatypeID,
		AdminFieldID:    c.params.AdminFieldID,
	})
}

func (d PsqlDatabase) NewAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeFieldParams) NewAdminDatatypeFieldCmdPsql {
	return NewAdminDatatypeFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

type UpdateAdminDatatypeFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateAdminDatatypeFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminDatatypeFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminDatatypeFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminDatatypeFieldCmdPsql) TableName() string                     { return "admin_datatypes_fields" }
func (c UpdateAdminDatatypeFieldCmdPsql) Params() any                           { return c.params }
func (c UpdateAdminDatatypeFieldCmdPsql) GetID() string                         { return c.params.ID }

func (c UpdateAdminDatatypeFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypesFields, error) {
	queries := mdbp.New(tx)
	rows, err := queries.ListAdminDatatypeField(ctx)
	if err != nil {
		return mdbp.AdminDatatypesFields{}, fmt.Errorf("list admin_datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.params.ID {
			return v, nil
		}
	}
	return mdbp.AdminDatatypesFields{}, fmt.Errorf("admin_datatypes_fields not found: %v", c.params.ID)
}

func (c UpdateAdminDatatypeFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminDatatypeField(ctx, mdbp.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: c.params.AdminDatatypeID,
		AdminFieldID:    c.params.AdminFieldID,
		ID:              c.params.ID,
	})
}

func (d PsqlDatabase) UpdateAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeFieldParams) UpdateAdminDatatypeFieldCmdPsql {
	return UpdateAdminDatatypeFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteAdminDatatypeFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminDatatypeFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteAdminDatatypeFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminDatatypeFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminDatatypeFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminDatatypeFieldCmdPsql) TableName() string                     { return "admin_datatypes_fields" }
func (c DeleteAdminDatatypeFieldCmdPsql) GetID() string                         { return c.id }

func (c DeleteAdminDatatypeFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypesFields, error) {
	queries := mdbp.New(tx)
	rows, err := queries.ListAdminDatatypeField(ctx)
	if err != nil {
		return mdbp.AdminDatatypesFields{}, fmt.Errorf("list admin_datatypes_fields for before snapshot: %w", err)
	}
	for _, v := range rows {
		if v.ID == c.id {
			return v, nil
		}
	}
	return mdbp.AdminDatatypesFields{}, fmt.Errorf("admin_datatypes_fields not found: %v", c.id)
}

func (c DeleteAdminDatatypeFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminDatatypeField(ctx, mdbp.DeleteAdminDatatypeFieldParams{ID: c.id})
}

func (d PsqlDatabase) DeleteAdminDatatypeFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteAdminDatatypeFieldCmdPsql {
	return DeleteAdminDatatypeFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
