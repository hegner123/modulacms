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

type Datatypes struct {
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type CreateDatatypeParams struct {
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type UpdateDatatypeParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
}

// DatatypeJSON provides a string-based representation for JSON serialization
type DatatypeJSON struct {
	DatatypeID   string `json:"datatype_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// MapDatatypeJSON converts Datatypes to DatatypeJSON for JSON serialization
func MapDatatypeJSON(a Datatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID:   a.DatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringDatatype converts Datatypes to StringDatatypes for table display
func MapStringDatatype(a Datatypes) StringDatatypes {
	return StringDatatypes{
		DatatypeID:   a.DatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
		History:      "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapDatatype(a mdb.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateDatatypeParams(a CreateDatatypeParams) mdb.CreateDatatypeParams {
	id := a.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	return mdb.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdb.UpdateDatatypeParams {
	return mdb.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		DatatypeID:   a.DatatypeID,
	}
}

// QUERIES

func (d Database) CountDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d Database) CreateDatatype(ctx context.Context, ac audited.AuditContext, s CreateDatatypeParams) (*Datatypes, error) {
	cmd := d.NewDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatype: %w", err)
	}
	r := d.MapDatatype(result)
	return &r, nil
}

func (d Database) DeleteDatatype(ctx context.Context, ac audited.AuditContext, id types.DatatypeID) error {
	cmd := d.DeleteDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdb.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d Database) ListDatatypes() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListDatatypeChildren(parentID types.DatatypeID) (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	// Convert DatatypeID to NullableContentID for the query (sqlc generates this param type)
	params := mdb.ListDatatypeChildrenParams{
		ParentID: types.NullableContentID{ID: types.ContentID(parentID), Valid: true},
	}
	rows, err := queries.ListDatatypeChildren(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateDatatype(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeParams) (*string, error) {
	cmd := d.UpdateDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapDatatype(a mdbm.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbm.CreateDatatypeParams {
	id := a.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	return mdbm.CreateDatatypeParams{
		DatatypeID: id,
		ParentID:   a.ParentID,
		Label:    a.Label,
		Type:     a.Type,
		AuthorID: a.AuthorID,
	}
}

func (d MysqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbm.UpdateDatatypeParams {
	return mdbm.UpdateDatatypeParams{
		ParentID:   a.ParentID,
		Label:      a.Label,
		Type:       a.Type,
		AuthorID:   a.AuthorID,
		DatatypeID: a.DatatypeID,
	}
}

// QUERIES

func (d MysqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateDatatype(ctx context.Context, ac audited.AuditContext, s CreateDatatypeParams) (*Datatypes, error) {
	cmd := d.NewDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatype: %w", err)
	}
	r := d.MapDatatype(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteDatatype(ctx context.Context, ac audited.AuditContext, id types.DatatypeID) error {
	cmd := d.DeleteDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdbm.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d MysqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListDatatypeChildren(parentID types.DatatypeID) (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	params := mdbm.ListDatatypeChildrenParams{
		ParentID: types.NullableContentID{ID: types.ContentID(parentID), Valid: true},
	}
	rows, err := queries.ListDatatypeChildren(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateDatatype(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeParams) (*string, error) {
	cmd := d.UpdateDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapDatatype(a mdbp.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbp.CreateDatatypeParams {
	id := a.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	return mdbp.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbp.UpdateDatatypeParams {
	return mdbp.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		DatatypeID:   a.DatatypeID,
	}
}

// QUERIES

func (d PsqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateDatatype(ctx context.Context, ac audited.AuditContext, s CreateDatatypeParams) (*Datatypes, error) {
	cmd := d.NewDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatype: %w", err)
	}
	r := d.MapDatatype(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteDatatype(ctx context.Context, ac audited.AuditContext, id types.DatatypeID) error {
	cmd := d.DeleteDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdbp.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListDatatypeChildren(parentID types.DatatypeID) (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	params := mdbp.ListDatatypeChildrenParams{
		ParentID: types.NullableContentID{ID: types.ContentID(parentID), Valid: true},
	}
	rows, err := queries.ListDatatypeChildren(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateDatatype(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeParams) (*string, error) {
	cmd := d.UpdateDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

type NewDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewDatatypeCmd) Context() context.Context              { return c.ctx }
func (c NewDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewDatatypeCmd) Connection() *sql.DB                   { return c.conn }
func (c NewDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewDatatypeCmd) TableName() string                     { return "datatypes" }
func (c NewDatatypeCmd) Params() any                           { return c.params }
func (c NewDatatypeCmd) GetID(d mdb.Datatypes) string          { return string(d.DatatypeID) }

func (c NewDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Datatypes, error) {
	id := c.params.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	queries := mdb.New(tx)
	return queries.CreateDatatype(ctx, mdb.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d Database) NewDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeParams) NewDatatypeCmd {
	return NewDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

type UpdateDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeCmd) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeCmd) TableName() string                     { return "datatypes" }
func (c UpdateDatatypeCmd) Params() any                           { return c.params }
func (c UpdateDatatypeCmd) GetID() string                         { return string(c.params.DatatypeID) }

func (c UpdateDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Datatypes, error) {
	queries := mdb.New(tx)
	return queries.GetDatatype(ctx, mdb.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

func (c UpdateDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateDatatype(ctx, mdb.UpdateDatatypeParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		DatatypeID:   c.params.DatatypeID,
	})
}

func (d Database) UpdateDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeParams) UpdateDatatypeCmd {
	return UpdateDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.DatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteDatatypeCmd) Context() context.Context              { return c.ctx }
func (c DeleteDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteDatatypeCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteDatatypeCmd) TableName() string                     { return "datatypes" }
func (c DeleteDatatypeCmd) GetID() string                         { return string(c.id) }

func (c DeleteDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Datatypes, error) {
	queries := mdb.New(tx)
	return queries.GetDatatype(ctx, mdb.GetDatatypeParams{DatatypeID: c.id})
}

func (c DeleteDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteDatatype(ctx, mdb.DeleteDatatypeParams{DatatypeID: c.id})
}

func (d Database) DeleteDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.DatatypeID) DeleteDatatypeCmd {
	return DeleteDatatypeCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewDatatypeCmdMysql) Context() context.Context              { return c.ctx }
func (c NewDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewDatatypeCmdMysql) TableName() string                     { return "datatypes" }
func (c NewDatatypeCmdMysql) Params() any                           { return c.params }
func (c NewDatatypeCmdMysql) GetID(d mdbm.Datatypes) string         { return string(d.DatatypeID) }

func (c NewDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Datatypes, error) {
	id := c.params.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	queries := mdbm.New(tx)
	params := mdbm.CreateDatatypeParams{
		DatatypeID: id,
		ParentID:   c.params.ParentID,
		Label:      c.params.Label,
		Type:       c.params.Type,
		AuthorID:   c.params.AuthorID,
	}
	if err := queries.CreateDatatype(ctx, params); err != nil {
		return mdbm.Datatypes{}, err
	}
	return queries.GetDatatype(ctx, mdbm.GetDatatypeParams{DatatypeID: params.DatatypeID})
}

func (d MysqlDatabase) NewDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeParams) NewDatatypeCmdMysql {
	return NewDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

type UpdateDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeCmdMysql) TableName() string                     { return "datatypes" }
func (c UpdateDatatypeCmdMysql) Params() any                           { return c.params }
func (c UpdateDatatypeCmdMysql) GetID() string                         { return string(c.params.DatatypeID) }

func (c UpdateDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Datatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatype(ctx, mdbm.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

func (c UpdateDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateDatatype(ctx, mdbm.UpdateDatatypeParams{
		ParentID:   c.params.ParentID,
		Label:      c.params.Label,
		Type:       c.params.Type,
		AuthorID:   c.params.AuthorID,
		DatatypeID: c.params.DatatypeID,
	})
}

func (d MysqlDatabase) UpdateDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeParams) UpdateDatatypeCmdMysql {
	return UpdateDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.DatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteDatatypeCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteDatatypeCmdMysql) TableName() string                     { return "datatypes" }
func (c DeleteDatatypeCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Datatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatype(ctx, mdbm.GetDatatypeParams{DatatypeID: c.id})
}

func (c DeleteDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteDatatype(ctx, mdbm.DeleteDatatypeParams{DatatypeID: c.id})
}

func (d MysqlDatabase) DeleteDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.DatatypeID) DeleteDatatypeCmdMysql {
	return DeleteDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewDatatypeCmdPsql) Context() context.Context              { return c.ctx }
func (c NewDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewDatatypeCmdPsql) TableName() string                     { return "datatypes" }
func (c NewDatatypeCmdPsql) Params() any                           { return c.params }
func (c NewDatatypeCmdPsql) GetID(d mdbp.Datatypes) string         { return string(d.DatatypeID) }

func (c NewDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Datatypes, error) {
	id := c.params.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	queries := mdbp.New(tx)
	return queries.CreateDatatype(ctx, mdbp.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d PsqlDatabase) NewDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeParams) NewDatatypeCmdPsql {
	return NewDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

type UpdateDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateDatatypeCmdPsql) TableName() string                     { return "datatypes" }
func (c UpdateDatatypeCmdPsql) Params() any                           { return c.params }
func (c UpdateDatatypeCmdPsql) GetID() string                         { return string(c.params.DatatypeID) }

func (c UpdateDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Datatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetDatatype(ctx, mdbp.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

func (c UpdateDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateDatatype(ctx, mdbp.UpdateDatatypeParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		DatatypeID:   c.params.DatatypeID,
	})
}

func (d PsqlDatabase) UpdateDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeParams) UpdateDatatypeCmdPsql {
	return UpdateDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.DatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteDatatypeCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteDatatypeCmdPsql) TableName() string                     { return "datatypes" }
func (c DeleteDatatypeCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Datatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetDatatype(ctx, mdbp.GetDatatypeParams{DatatypeID: c.id})
}

func (c DeleteDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteDatatype(ctx, mdbp.DeleteDatatypeParams{DatatypeID: c.id})
}

func (d PsqlDatabase) DeleteDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.DatatypeID) DeleteDatatypeCmdPsql {
	return DeleteDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
