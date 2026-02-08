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

type AdminDatatypes struct {
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
	ParentID        types.NullableContentID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
	AuthorID        types.NullableUserID    `json:"author_id"`
	DateCreated     types.Timestamp         `json:"date_created"`
	DateModified    types.Timestamp         `json:"date_modified"`
}
type CreateAdminDatatypeParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}
type ListAdminDatatypeByRouteIdRow struct {
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
	AdminRouteID    types.NullableRouteID   `json:"admin_route_id"`
	ParentID        types.NullableContentID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
}
type UpdateAdminDatatypeParams struct {
	ParentID        types.NullableContentID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
	AuthorID        types.NullableUserID    `json:"author_id"`
	DateCreated     types.Timestamp         `json:"date_created"`
	DateModified    types.Timestamp         `json:"date_modified"`
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
}
type UtilityGetAdminDatatypesRow struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	Label           string                `json:"label"`
}

// FormParams and JSON variants removed - use typed params directly

// MapAdminDatatypeJSON converts AdminDatatypes to DatatypeJSON for tree building.
// Maps admin datatype ID into the public DatatypeJSON shape so BuildNodes works unchanged.
func MapAdminDatatypeJSON(a AdminDatatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID:   a.AdminDatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringAdminDatatype converts AdminDatatypes to StringAdminDatatypes for table display
func MapStringAdminDatatype(a AdminDatatypes) StringAdminDatatypes {
	return StringAdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID.String(),
		ParentID:        a.ParentID.String(),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID.String(),
		DateCreated:     a.DateCreated.String(),
		DateModified:    a.DateModified.String(),
		History:         "", // History field removed from AdminDatatypes
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminDatatype(a mdb.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

func (d Database) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdb.CreateAdminDatatypeParams {
	return mdb.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdb.UpdateAdminDatatypeParams {
	return mdb.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

func (d Database) CountAdminDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminDatatype(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeParams) (*AdminDatatypes, error) {
	cmd := d.NewAdminDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatype: %w", err)
	}
	r := d.MapAdminDatatype(result)
	return &r, nil
}
func (d Database) CreateAdminDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d Database) DeleteAdminDatatype(ctx context.Context, ac audited.AuditContext, id types.AdminDatatypeID) error {
	cmd := d.DeleteAdminDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdb.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d Database) ListAdminDatatypeGlobalId() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeGlobal(d.Context)
	if err != nil {
		return nil, err
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminDatatype(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminDatatype(a mdbm.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbm.CreateAdminDatatypeParams {
	return mdbm.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbm.UpdateAdminDatatypeParams {
	return mdbm.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminDatatype(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeParams) (*AdminDatatypes, error) {
	cmd := d.NewAdminDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatype: %w", err)
	}
	r := d.MapAdminDatatype(result)
	return &r, nil
}

func (d MysqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d MysqlDatabase) DeleteAdminDatatype(ctx context.Context, ac audited.AuditContext, id types.AdminDatatypeID) error {
	cmd := d.DeleteAdminDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdbm.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateAdminDatatype(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapAdminDatatype(a mdbp.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbp.CreateAdminDatatypeParams {
	return mdbp.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbp.UpdateAdminDatatypeParams {
	return mdbp.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

func (d PsqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminDatatype(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeParams) (*AdminDatatypes, error) {
	cmd := d.NewAdminDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatype: %w", err)
	}
	r := d.MapAdminDatatype(result)
	return &r, nil
}

func (d PsqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d PsqlDatabase) DeleteAdminDatatype(ctx context.Context, ac audited.AuditContext, id types.AdminDatatypeID) error {
	cmd := d.DeleteAdminDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdbp.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminDatatype(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

type NewAdminDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminDatatypeCmd) Context() context.Context              { return c.ctx }
func (c NewAdminDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminDatatypeCmd) Connection() *sql.DB                   { return c.conn }
func (c NewAdminDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminDatatypeCmd) TableName() string                     { return "admin_datatypes" }
func (c NewAdminDatatypeCmd) Params() any                           { return c.params }
func (c NewAdminDatatypeCmd) GetID(u mdb.AdminDatatypes) string     { return string(u.AdminDatatypeID) }

func (c NewAdminDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypes, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminDatatype(ctx, mdb.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
	})
}

func (d Database) NewAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeParams) NewAdminDatatypeCmd {
	return NewAdminDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

type UpdateAdminDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeCmd) Context() context.Context              { return c.ctx }
func (c UpdateAdminDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminDatatypeCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminDatatypeCmd) TableName() string                     { return "admin_datatypes" }
func (c UpdateAdminDatatypeCmd) Params() any                           { return c.params }
func (c UpdateAdminDatatypeCmd) GetID() string                         { return string(c.params.AdminDatatypeID) }

func (c UpdateAdminDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypes, error) {
	queries := mdb.New(tx)
	return queries.GetAdminDatatype(ctx, mdb.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

func (c UpdateAdminDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminDatatype(ctx, mdb.UpdateAdminDatatypeParams{
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

func (d Database) UpdateAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeParams) UpdateAdminDatatypeCmd {
	return UpdateAdminDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteAdminDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminDatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminDatatypeCmd) Context() context.Context              { return c.ctx }
func (c DeleteAdminDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminDatatypeCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminDatatypeCmd) TableName() string                     { return "admin_datatypes" }
func (c DeleteAdminDatatypeCmd) GetID() string                         { return string(c.id) }

func (c DeleteAdminDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypes, error) {
	queries := mdb.New(tx)
	return queries.GetAdminDatatype(ctx, mdb.GetAdminDatatypeParams{AdminDatatypeID: c.id})
}

func (c DeleteAdminDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminDatatype(ctx, mdb.DeleteAdminDatatypeParams{AdminDatatypeID: c.id})
}

func (d Database) DeleteAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminDatatypeID) DeleteAdminDatatypeCmd {
	return DeleteAdminDatatypeCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewAdminDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminDatatypeCmdMysql) Context() context.Context              { return c.ctx }
func (c NewAdminDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminDatatypeCmdMysql) TableName() string                     { return "admin_datatypes" }
func (c NewAdminDatatypeCmdMysql) Params() any                           { return c.params }
func (c NewAdminDatatypeCmdMysql) GetID(u mdbm.AdminDatatypes) string    { return string(u.AdminDatatypeID) }

func (c NewAdminDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypes, error) {
	queries := mdbm.New(tx)
	id := types.NewAdminDatatypeID()
	err := queries.CreateAdminDatatype(ctx, mdbm.CreateAdminDatatypeParams{
		AdminDatatypeID: id,
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
	})
	if err != nil {
		return mdbm.AdminDatatypes{}, err
	}
	return queries.GetAdminDatatype(ctx, mdbm.GetAdminDatatypeParams{AdminDatatypeID: id})
}

func (d MysqlDatabase) NewAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeParams) NewAdminDatatypeCmdMysql {
	return NewAdminDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

type UpdateAdminDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateAdminDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminDatatypeCmdMysql) TableName() string                     { return "admin_datatypes" }
func (c UpdateAdminDatatypeCmdMysql) Params() any                           { return c.params }
func (c UpdateAdminDatatypeCmdMysql) GetID() string                         { return string(c.params.AdminDatatypeID) }

func (c UpdateAdminDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminDatatype(ctx, mdbm.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

func (c UpdateAdminDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminDatatype(ctx, mdbm.UpdateAdminDatatypeParams{
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

func (d MysqlDatabase) UpdateAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeParams) UpdateAdminDatatypeCmdMysql {
	return UpdateAdminDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteAdminDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminDatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminDatatypeCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteAdminDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminDatatypeCmdMysql) TableName() string                     { return "admin_datatypes" }
func (c DeleteAdminDatatypeCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteAdminDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminDatatype(ctx, mdbm.GetAdminDatatypeParams{AdminDatatypeID: c.id})
}

func (c DeleteAdminDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminDatatype(ctx, mdbm.DeleteAdminDatatypeParams{AdminDatatypeID: c.id})
}

func (d MysqlDatabase) DeleteAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminDatatypeID) DeleteAdminDatatypeCmdMysql {
	return DeleteAdminDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewAdminDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminDatatypeCmdPsql) Context() context.Context              { return c.ctx }
func (c NewAdminDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminDatatypeCmdPsql) TableName() string                     { return "admin_datatypes" }
func (c NewAdminDatatypeCmdPsql) Params() any                           { return c.params }
func (c NewAdminDatatypeCmdPsql) GetID(u mdbp.AdminDatatypes) string    { return string(u.AdminDatatypeID) }

func (c NewAdminDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypes, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminDatatype(ctx, mdbp.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
	})
}

func (d PsqlDatabase) NewAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeParams) NewAdminDatatypeCmdPsql {
	return NewAdminDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

type UpdateAdminDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateAdminDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminDatatypeCmdPsql) TableName() string                     { return "admin_datatypes" }
func (c UpdateAdminDatatypeCmdPsql) Params() any                           { return c.params }
func (c UpdateAdminDatatypeCmdPsql) GetID() string                         { return string(c.params.AdminDatatypeID) }

func (c UpdateAdminDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminDatatype(ctx, mdbp.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

func (c UpdateAdminDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminDatatype(ctx, mdbp.UpdateAdminDatatypeParams{
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

func (d PsqlDatabase) UpdateAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeParams) UpdateAdminDatatypeCmdPsql {
	return UpdateAdminDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteAdminDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminDatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminDatatypeCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteAdminDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminDatatypeCmdPsql) TableName() string                     { return "admin_datatypes" }
func (c DeleteAdminDatatypeCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteAdminDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminDatatype(ctx, mdbp.GetAdminDatatypeParams{AdminDatatypeID: c.id})
}

func (c DeleteAdminDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminDatatype(ctx, mdbp.DeleteAdminDatatypeParams{AdminDatatypeID: c.id})
}

func (d PsqlDatabase) DeleteAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminDatatypeID) DeleteAdminDatatypeCmdPsql {
	return DeleteAdminDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
