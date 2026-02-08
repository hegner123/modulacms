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

type AdminContentFields struct {
	AdminContentFieldID types.AdminContentFieldID  `json:"admin_content_field_id"`
	AdminRouteID        sql.NullString             `json:"admin_route_id"`
	AdminContentDataID  string                     `json:"admin_content_data_id"`
	AdminFieldID        types.NullableAdminFieldID `json:"admin_field_id"`
	AdminFieldValue     string                     `json:"admin_field_value"`
	AuthorID            types.NullableUserID       `json:"author_id"`
	DateCreated         types.Timestamp            `json:"date_created"`
	DateModified        types.Timestamp            `json:"date_modified"`
}
type CreateAdminContentFieldParams struct {
	AdminRouteID       sql.NullString             `json:"admin_route_id"`
	AdminContentDataID string                     `json:"admin_content_data_id"`
	AdminFieldID       types.NullableAdminFieldID `json:"admin_field_id"`
	AdminFieldValue    string                     `json:"admin_field_value"`
	AuthorID           types.NullableUserID       `json:"author_id"`
	DateCreated        types.Timestamp            `json:"date_created"`
	DateModified       types.Timestamp            `json:"date_modified"`
}
type UpdateAdminContentFieldParams struct {
	AdminRouteID        sql.NullString             `json:"admin_route_id"`
	AdminContentDataID  string                     `json:"admin_content_data_id"`
	AdminFieldID        types.NullableAdminFieldID `json:"admin_field_id"`
	AdminFieldValue     string                     `json:"admin_field_value"`
	AuthorID            types.NullableUserID       `json:"author_id"`
	DateCreated         types.Timestamp            `json:"date_created"`
	DateModified        types.Timestamp            `json:"date_modified"`
	AdminContentFieldID types.AdminContentFieldID  `json:"admin_content_field_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapAdminContentFieldJSON converts AdminContentFields to ContentFieldsJSON for tree building.
// Maps admin field value into the public ContentFieldsJSON shape so BuildNodes works unchanged.
func MapAdminContentFieldJSON(a AdminContentFields) ContentFieldsJSON {
	return ContentFieldsJSON{
		ContentFieldID: 0,
		RouteID:        0,
		ContentDataID:  0,
		FieldID:        0,
		FieldValue:     a.AdminFieldValue,
		AuthorID:       0,
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}

// MapStringAdminContentField converts AdminContentFields to StringAdminContentFields for table display
func MapStringAdminContentField(a AdminContentFields) StringAdminContentFields {
	adminRouteID := ""
	if a.AdminRouteID.Valid {
		adminRouteID = a.AdminRouteID.String
	}
	return StringAdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID.String(),
		AdminRouteID:        adminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID.String(),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID.String(),
		DateCreated:         a.DateCreated.String(),
		DateModified:        a.DateModified.String(),
		History:             "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminContentField(a mdb.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}

func (d Database) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdb.CreateAdminContentFieldParams {
	return mdb.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}

func (d Database) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdb.UpdateAdminContentFieldParams {
	return mdb.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// QUERIES

func (d Database) CountAdminContentFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d Database) CreateAdminContentField(ctx context.Context, ac audited.AuditContext, s CreateAdminContentFieldParams) (*AdminContentFields, error) {
	cmd := d.NewAdminContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentField: %w", err)
	}
	r := d.MapAdminContentField(result)
	return &r, nil
}
func (d Database) CreateAdminContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d Database) DeleteAdminContentField(ctx context.Context, ac audited.AuditContext, id types.AdminContentFieldID) error {
	cmd := d.DeleteAdminContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d Database) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdb.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d Database) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdb.ListAdminContentFieldsByRouteParams{AdminRouteID: StringToNullString(id)})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) UpdateAdminContentField(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentFieldParams) (*string, error) {
	cmd := d.UpdateAdminContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentFieldID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminContentField(a mdbm.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d MysqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbm.CreateAdminContentFieldParams {
	return mdbm.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d MysqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbm.UpdateAdminContentFieldParams {
	return mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminContentFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateAdminContentField(ctx context.Context, ac audited.AuditContext, s CreateAdminContentFieldParams) (*AdminContentFields, error) {
	cmd := d.NewAdminContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentField: %w", err)
	}
	r := d.MapAdminContentField(result)
	return &r, nil
}

func (d MysqlDatabase) CreateAdminContentFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d MysqlDatabase) DeleteAdminContentField(ctx context.Context, ac audited.AuditContext, id types.AdminContentFieldID) error {
	cmd := d.DeleteAdminContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d MysqlDatabase) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdbm.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdbm.ListAdminContentFieldsByRouteParams{AdminRouteID: StringToNullString(id)})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateAdminContentField(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentFieldParams) (*string, error) {
	cmd := d.UpdateAdminContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentFieldID)
	return &msg, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS

func (d PsqlDatabase) MapAdminContentField(a mdbp.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d PsqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbp.CreateAdminContentFieldParams {
	return mdbp.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d PsqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbp.UpdateAdminContentFieldParams {
	return mdbp.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// /QUERIES

func (d PsqlDatabase) CountAdminContentFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateAdminContentField(ctx context.Context, ac audited.AuditContext, s CreateAdminContentFieldParams) (*AdminContentFields, error) {
	cmd := d.NewAdminContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentField: %w", err)
	}
	r := d.MapAdminContentField(result)
	return &r, nil
}
func (d PsqlDatabase) CreateAdminContentFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminContentField(ctx context.Context, ac audited.AuditContext, id types.AdminContentFieldID) error {
	cmd := d.DeleteAdminContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d PsqlDatabase) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdbp.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdbp.ListAdminContentFieldsByRouteParams{AdminRouteID: StringToNullString(id)})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) UpdateAdminContentField(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentFieldParams) (*string, error) {
	cmd := d.UpdateAdminContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentFieldID)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewAdminContentFieldCmd is an audited create command for admin_content_fields (SQLite).
type NewAdminContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentFieldCmd) Context() context.Context              { return c.ctx }
func (c NewAdminContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentFieldCmd) TableName() string                     { return "admin_content_fields" }
func (c NewAdminContentFieldCmd) Params() any                           { return c.params }
func (c NewAdminContentFieldCmd) GetID(row mdb.AdminContentFields) string {
	return string(row.AdminContentFieldID)
}

func (c NewAdminContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminContentFields, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminContentField(ctx, mdb.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
	})
}

func (d Database) NewAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentFieldParams) NewAdminContentFieldCmd {
	return NewAdminContentFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateAdminContentFieldCmd is an audited update command for admin_content_fields (SQLite).
type UpdateAdminContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentFieldCmd) Context() context.Context              { return c.ctx }
func (c UpdateAdminContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminContentFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminContentFieldCmd) TableName() string                     { return "admin_content_fields" }
func (c UpdateAdminContentFieldCmd) Params() any                           { return c.params }
func (c UpdateAdminContentFieldCmd) GetID() string {
	return string(c.params.AdminContentFieldID)
}

func (c UpdateAdminContentFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentField(ctx, mdb.GetAdminContentFieldParams{AdminContentFieldID: c.params.AdminContentFieldID})
}

func (c UpdateAdminContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminContentField(ctx, mdb.UpdateAdminContentFieldParams{
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
		AdminContentFieldID: c.params.AdminContentFieldID,
	})
}

func (d Database) UpdateAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentFieldParams) UpdateAdminContentFieldCmd {
	return UpdateAdminContentFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// DeleteAdminContentFieldCmd is an audited delete command for admin_content_fields (SQLite).
type DeleteAdminContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentFieldCmd) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentFieldCmd) TableName() string                     { return "admin_content_fields" }
func (c DeleteAdminContentFieldCmd) GetID() string                         { return string(c.id) }

func (c DeleteAdminContentFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentField(ctx, mdb.GetAdminContentFieldParams{AdminContentFieldID: c.id})
}

func (c DeleteAdminContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminContentField(ctx, mdb.DeleteAdminContentFieldParams{AdminContentFieldID: c.id})
}

func (d Database) DeleteAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentFieldID) DeleteAdminContentFieldCmd {
	return DeleteAdminContentFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

// NewAdminContentFieldCmdMysql is an audited create command for admin_content_fields (MySQL).
type NewAdminContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c NewAdminContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentFieldCmdMysql) TableName() string                     { return "admin_content_fields" }
func (c NewAdminContentFieldCmdMysql) Params() any                           { return c.params }
func (c NewAdminContentFieldCmdMysql) GetID(row mdbm.AdminContentFields) string {
	return string(row.AdminContentFieldID)
}

func (c NewAdminContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentFields, error) {
	id := types.NewAdminContentFieldID()
	queries := mdbm.New(tx)
	err := queries.CreateAdminContentField(ctx, mdbm.CreateAdminContentFieldParams{
		AdminContentFieldID: id,
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
	})
	if err != nil {
		return mdbm.AdminContentFields{}, fmt.Errorf("execute create admin_content_fields: %w", err)
	}
	return queries.GetAdminContentField(ctx, mdbm.GetAdminContentFieldParams{AdminContentFieldID: id})
}

func (d MysqlDatabase) NewAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentFieldParams) NewAdminContentFieldCmdMysql {
	return NewAdminContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateAdminContentFieldCmdMysql is an audited update command for admin_content_fields (MySQL).
type UpdateAdminContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateAdminContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminContentFieldCmdMysql) TableName() string                     { return "admin_content_fields" }
func (c UpdateAdminContentFieldCmdMysql) Params() any                           { return c.params }
func (c UpdateAdminContentFieldCmdMysql) GetID() string {
	return string(c.params.AdminContentFieldID)
}

func (c UpdateAdminContentFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentField(ctx, mdbm.GetAdminContentFieldParams{AdminContentFieldID: c.params.AdminContentFieldID})
}

func (c UpdateAdminContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminContentField(ctx, mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
		AdminContentFieldID: c.params.AdminContentFieldID,
	})
}

func (d MysqlDatabase) UpdateAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentFieldParams) UpdateAdminContentFieldCmdMysql {
	return UpdateAdminContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// DeleteAdminContentFieldCmdMysql is an audited delete command for admin_content_fields (MySQL).
type DeleteAdminContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentFieldCmdMysql) TableName() string                     { return "admin_content_fields" }
func (c DeleteAdminContentFieldCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteAdminContentFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentField(ctx, mdbm.GetAdminContentFieldParams{AdminContentFieldID: c.id})
}

func (c DeleteAdminContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminContentField(ctx, mdbm.DeleteAdminContentFieldParams{AdminContentFieldID: c.id})
}

func (d MysqlDatabase) DeleteAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentFieldID) DeleteAdminContentFieldCmdMysql {
	return DeleteAdminContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

// NewAdminContentFieldCmdPsql is an audited create command for admin_content_fields (PostgreSQL).
type NewAdminContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c NewAdminContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentFieldCmdPsql) TableName() string                     { return "admin_content_fields" }
func (c NewAdminContentFieldCmdPsql) Params() any                           { return c.params }
func (c NewAdminContentFieldCmdPsql) GetID(row mdbp.AdminContentFields) string {
	return string(row.AdminContentFieldID)
}

func (c NewAdminContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentFields, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminContentField(ctx, mdbp.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
	})
}

func (d PsqlDatabase) NewAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentFieldParams) NewAdminContentFieldCmdPsql {
	return NewAdminContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateAdminContentFieldCmdPsql is an audited update command for admin_content_fields (PostgreSQL).
type UpdateAdminContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateAdminContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminContentFieldCmdPsql) TableName() string                     { return "admin_content_fields" }
func (c UpdateAdminContentFieldCmdPsql) Params() any                           { return c.params }
func (c UpdateAdminContentFieldCmdPsql) GetID() string {
	return string(c.params.AdminContentFieldID)
}

func (c UpdateAdminContentFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentField(ctx, mdbp.GetAdminContentFieldParams{AdminContentFieldID: c.params.AdminContentFieldID})
}

func (c UpdateAdminContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminContentField(ctx, mdbp.UpdateAdminContentFieldParams{
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
		AdminContentFieldID: c.params.AdminContentFieldID,
	})
}

func (d PsqlDatabase) UpdateAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentFieldParams) UpdateAdminContentFieldCmdPsql {
	return UpdateAdminContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// DeleteAdminContentFieldCmdPsql is an audited delete command for admin_content_fields (PostgreSQL).
type DeleteAdminContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentFieldCmdPsql) TableName() string                     { return "admin_content_fields" }
func (c DeleteAdminContentFieldCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteAdminContentFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentField(ctx, mdbp.GetAdminContentFieldParams{AdminContentFieldID: c.id})
}

func (c DeleteAdminContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminContentField(ctx, mdbp.DeleteAdminContentFieldParams{AdminContentFieldID: c.id})
}

func (d PsqlDatabase) DeleteAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentFieldID) DeleteAdminContentFieldCmdPsql {
	return DeleteAdminContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
