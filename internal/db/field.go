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

type Fields struct {
	FieldID      types.FieldID            `json:"field_id"`
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}
type CreateFieldParams struct {
	FieldID      types.FieldID            `json:"field_id"`
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}
type UpdateFieldParams struct {
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
	FieldID      types.FieldID           `json:"field_id"`
}
// FieldsJSON is used for JSON serialization in model package
type FieldsJSON struct {
	FieldID      string `json:"field_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Validation   string `json:"validation"`
	UIConfig     string `json:"ui_config"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// MapFieldJSON converts Fields to FieldsJSON for JSON serialization
func MapFieldJSON(a Fields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.FieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UIConfig,
		Type:         a.Type.String(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringField converts Fields to StringFields for table display
func MapStringField(a Fields) StringFields {
	return StringFields{
		FieldID:      a.FieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UIConfig,
		Type:         a.Type.String(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
		History:      "", // History field removed from schema
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapField(a mdb.Fields) Fields {
	return Fields{
		FieldID:      a.FieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UiConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateFieldParams(a CreateFieldParams) mdb.CreateFieldParams {
	id := a.FieldID
	if id.IsZero() {
		id = types.NewFieldID()
	}
	return mdb.CreateFieldParams{
		FieldID:      id,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateFieldParams(a UpdateFieldParams) mdb.UpdateFieldParams {
	return mdb.UpdateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		FieldID:      a.FieldID,
	}
}

// QUERIES

func (d Database) CountFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}
func (d Database) CreateField(ctx context.Context, ac audited.AuditContext, s CreateFieldParams) (*Fields, error) {
	cmd := d.NewFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create field: %w", err)
	}
	r := d.MapField(result)
	return &r, nil
}
func (d Database) DeleteField(ctx context.Context, ac audited.AuditContext, id types.FieldID) error {
	cmd := d.DeleteFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetField(id types.FieldID) (*Fields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetField(d.Context, mdb.GetFieldParams{FieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}

func (d Database) ListFields() (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListFieldsByDatatypeID(id types.NullableDatatypeID) (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, mdb.ListFieldByDatatypeIDParams{ParentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdb.ListFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %v", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateField(ctx context.Context, ac audited.AuditContext, s UpdateFieldParams) (*string, error) {
	cmd := d.UpdateFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update field: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapField(a mdbm.Fields) Fields {
	return Fields{
		FieldID:      a.FieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UiConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbm.CreateFieldParams {
	id := a.FieldID
	if id.IsZero() {
		id = types.NewFieldID()
	}
	return mdbm.CreateFieldParams{
		FieldID:      id,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbm.UpdateFieldParams {
	return mdbm.UpdateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		FieldID:      a.FieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}
func (d MysqlDatabase) CreateField(ctx context.Context, ac audited.AuditContext, s CreateFieldParams) (*Fields, error) {
	cmd := d.NewFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create field: %w", err)
	}
	r := d.MapField(result)
	return &r, nil
}
func (d MysqlDatabase) DeleteField(ctx context.Context, ac audited.AuditContext, id types.FieldID) error {
	cmd := d.DeleteFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d MysqlDatabase) GetField(id types.FieldID) (*Fields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetField(d.Context, mdbm.GetFieldParams{FieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}
func (d MysqlDatabase) ListFields() (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListFieldsByDatatypeID(id types.NullableDatatypeID) (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, mdbm.ListFieldByDatatypeIDParams{ParentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdbm.ListFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %v", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateField(ctx context.Context, ac audited.AuditContext, s UpdateFieldParams) (*string, error) {
	cmd := d.UpdateFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update field: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapField(a mdbp.Fields) Fields {
	return Fields{
		FieldID:      a.FieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UiConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbp.CreateFieldParams {
	id := a.FieldID
	if id.IsZero() {
		id = types.NewFieldID()
	}
	return mdbp.CreateFieldParams{
		FieldID:      id,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbp.UpdateFieldParams {
	return mdbp.UpdateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		FieldID:      a.FieldID,
	}
}

// QUERIES

func (d PsqlDatabase) CountFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}
func (d PsqlDatabase) CreateField(ctx context.Context, ac audited.AuditContext, s CreateFieldParams) (*Fields, error) {
	cmd := d.NewFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create field: %w", err)
	}
	r := d.MapField(result)
	return &r, nil
}
func (d PsqlDatabase) DeleteField(ctx context.Context, ac audited.AuditContext, id types.FieldID) error {
	cmd := d.DeleteFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d PsqlDatabase) GetField(id types.FieldID) (*Fields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetField(d.Context, mdbp.GetFieldParams{FieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}
func (d PsqlDatabase) ListFields() (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListFieldsByDatatypeID(id types.NullableDatatypeID) (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, mdbp.ListFieldByDatatypeIDParams{ParentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdbp.ListFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %v", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateField(ctx context.Context, ac audited.AuditContext, s UpdateFieldParams) (*string, error) {
	cmd := d.UpdateFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update field: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

type NewFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewFieldCmd) Context() context.Context              { return c.ctx }
func (c NewFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c NewFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewFieldCmd) TableName() string                     { return "fields" }
func (c NewFieldCmd) Params() any                           { return c.params }
func (c NewFieldCmd) GetID(f mdb.Fields) string             { return string(f.FieldID) }

func (c NewFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Fields, error) {
	id := c.params.FieldID
	if id.IsZero() {
		id = types.NewFieldID()
	}
	queries := mdb.New(tx)
	return queries.CreateField(ctx, mdb.CreateFieldParams{
		FieldID:      id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d Database) NewFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateFieldParams) NewFieldCmd {
	return NewFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

type UpdateFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateFieldCmd) Context() context.Context              { return c.ctx }
func (c UpdateFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateFieldCmd) TableName() string                     { return "fields" }
func (c UpdateFieldCmd) Params() any                           { return c.params }
func (c UpdateFieldCmd) GetID() string                         { return string(c.params.FieldID) }

func (c UpdateFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Fields, error) {
	queries := mdb.New(tx)
	return queries.GetField(ctx, mdb.GetFieldParams{FieldID: c.params.FieldID})
}

func (c UpdateFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateField(ctx, mdb.UpdateFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		FieldID:      c.params.FieldID,
	})
}

func (d Database) UpdateFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateFieldParams) UpdateFieldCmd {
	return UpdateFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.FieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteFieldCmd) Context() context.Context              { return c.ctx }
func (c DeleteFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteFieldCmd) TableName() string                     { return "fields" }
func (c DeleteFieldCmd) GetID() string                         { return string(c.id) }

func (c DeleteFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Fields, error) {
	queries := mdb.New(tx)
	return queries.GetField(ctx, mdb.GetFieldParams{FieldID: c.id})
}

func (c DeleteFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteField(ctx, mdb.DeleteFieldParams{FieldID: c.id})
}

func (d Database) DeleteFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.FieldID) DeleteFieldCmd {
	return DeleteFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c NewFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewFieldCmdMysql) TableName() string                     { return "fields" }
func (c NewFieldCmdMysql) Params() any                           { return c.params }
func (c NewFieldCmdMysql) GetID(f mdbm.Fields) string            { return string(f.FieldID) }

func (c NewFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Fields, error) {
	id := c.params.FieldID
	if id.IsZero() {
		id = types.NewFieldID()
	}
	queries := mdbm.New(tx)
	params := mdbm.CreateFieldParams{
		FieldID:      id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	}
	if err := queries.CreateField(ctx, params); err != nil {
		return mdbm.Fields{}, err
	}
	return queries.GetField(ctx, mdbm.GetFieldParams{FieldID: params.FieldID})
}

func (d MysqlDatabase) NewFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateFieldParams) NewFieldCmdMysql {
	return NewFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

type UpdateFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateFieldCmdMysql) TableName() string                     { return "fields" }
func (c UpdateFieldCmdMysql) Params() any                           { return c.params }
func (c UpdateFieldCmdMysql) GetID() string                         { return string(c.params.FieldID) }

func (c UpdateFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Fields, error) {
	queries := mdbm.New(tx)
	return queries.GetField(ctx, mdbm.GetFieldParams{FieldID: c.params.FieldID})
}

func (c UpdateFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateField(ctx, mdbm.UpdateFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		FieldID:      c.params.FieldID,
	})
}

func (d MysqlDatabase) UpdateFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateFieldParams) UpdateFieldCmdMysql {
	return UpdateFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.FieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteFieldCmdMysql) TableName() string                     { return "fields" }
func (c DeleteFieldCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Fields, error) {
	queries := mdbm.New(tx)
	return queries.GetField(ctx, mdbm.GetFieldParams{FieldID: c.id})
}

func (c DeleteFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteField(ctx, mdbm.DeleteFieldParams{FieldID: c.id})
}

func (d MysqlDatabase) DeleteFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.FieldID) DeleteFieldCmdMysql {
	return DeleteFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c NewFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewFieldCmdPsql) TableName() string                     { return "fields" }
func (c NewFieldCmdPsql) Params() any                           { return c.params }
func (c NewFieldCmdPsql) GetID(f mdbp.Fields) string            { return string(f.FieldID) }

func (c NewFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Fields, error) {
	id := c.params.FieldID
	if id.IsZero() {
		id = types.NewFieldID()
	}
	queries := mdbp.New(tx)
	return queries.CreateField(ctx, mdbp.CreateFieldParams{
		FieldID:      id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d PsqlDatabase) NewFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateFieldParams) NewFieldCmdPsql {
	return NewFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

type UpdateFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateFieldCmdPsql) TableName() string                     { return "fields" }
func (c UpdateFieldCmdPsql) Params() any                           { return c.params }
func (c UpdateFieldCmdPsql) GetID() string                         { return string(c.params.FieldID) }

func (c UpdateFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Fields, error) {
	queries := mdbp.New(tx)
	return queries.GetField(ctx, mdbp.GetFieldParams{FieldID: c.params.FieldID})
}

func (c UpdateFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateField(ctx, mdbp.UpdateFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		FieldID:      c.params.FieldID,
	})
}

func (d PsqlDatabase) UpdateFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateFieldParams) UpdateFieldCmdPsql {
	return UpdateFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.FieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteFieldCmdPsql) TableName() string                     { return "fields" }
func (c DeleteFieldCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Fields, error) {
	queries := mdbp.New(tx)
	return queries.GetField(ctx, mdbp.GetFieldParams{FieldID: c.id})
}

func (c DeleteFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteField(ctx, mdbp.DeleteFieldParams{FieldID: c.id})
}

func (d PsqlDatabase) DeleteFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.FieldID) DeleteFieldCmdPsql {
	return DeleteFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
