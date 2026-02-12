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

type AdminFields struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type CreateAdminFieldParams struct {
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type UpdateAdminFieldParams struct {
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
	AdminFieldID types.AdminFieldID      `json:"admin_field_id"`
}

type ListAdminFieldByRouteIdRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
}

type ListAdminFieldsByParentIDPaginatedParams struct {
	ParentID types.AdminDatatypeID
	Limit    int64
	Offset   int64
}

type ListAdminFieldsByDatatypeIDRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Validation   string                  `json:"validation"`
	UIConfig     string                  `json:"ui_config"`
	Type         types.FieldType         `json:"type"`
}

type UtilityGetAdminfieldsRow struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	Label        string             `json:"label"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapAdminFieldJSON converts AdminFields to FieldsJSON for tree building.
// Maps admin field ID into the public FieldsJSON shape so BuildNodes works unchanged.
func MapAdminFieldJSON(a AdminFields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.AdminFieldID.String(),
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

// MapStringAdminField converts AdminFields to StringAdminFields for table display
func MapStringAdminField(a AdminFields) StringAdminFields {
	return StringAdminFields{
		AdminFieldID: a.AdminFieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UIConfig,
		Type:         string(a.Type),
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

func (d Database) MapAdminField(a mdb.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
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

func (d Database) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdb.CreateAdminFieldParams {
	return mdb.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
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

func (d Database) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdb.UpdateAdminFieldParams {
	return mdb.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

func (d Database) CountAdminFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminField(ctx context.Context, ac audited.AuditContext, s CreateAdminFieldParams) (*AdminFields, error) {
	cmd := d.NewAdminFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminField: %w", err)
	}
	r := d.MapAdminField(result)
	return &r, nil
}

func (d Database) CreateAdminFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

func (d Database) DeleteAdminField(ctx context.Context, ac audited.AuditContext, id types.AdminFieldID) error {
	cmd := d.DeleteAdminFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdb.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d Database) ListAdminFields() (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdb.ListAdminFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminFieldsByParentIDPaginated(params ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentIDPaginated(d.Context, mdb.ListAdminFieldByParentIDPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by parent paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminField(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldParams) (*string, error) {
	cmd := d.UpdateAdminFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminField(a mdbm.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
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

func (d MysqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbm.CreateAdminFieldParams {
	return mdbm.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
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

func (d MysqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbm.UpdateAdminFieldParams {
	return mdbm.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminField(ctx context.Context, ac audited.AuditContext, s CreateAdminFieldParams) (*AdminFields, error) {
	cmd := d.NewAdminFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminField: %w", err)
	}
	r := d.MapAdminField(result)
	return &r, nil
}

func (d MysqlDatabase) CreateAdminFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

func (d MysqlDatabase) DeleteAdminField(ctx context.Context, ac audited.AuditContext, id types.AdminFieldID) error {
	cmd := d.DeleteAdminFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdbm.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d MysqlDatabase) ListAdminFields() (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbm.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminFieldsByParentIDPaginated(params ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentIDPaginated(d.Context, mdbm.ListAdminFieldByParentIDPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by parent paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateAdminField(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldParams) (*string, error) {
	cmd := d.UpdateAdminFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapAdminField(a mdbp.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
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

func (d PsqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbp.CreateAdminFieldParams {
	return mdbp.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
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

func (d PsqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbp.UpdateAdminFieldParams {
	return mdbp.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

func (d PsqlDatabase) CountAdminFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminField(ctx context.Context, ac audited.AuditContext, s CreateAdminFieldParams) (*AdminFields, error) {
	cmd := d.NewAdminFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminField: %w", err)
	}
	r := d.MapAdminField(result)
	return &r, nil
}

func (d PsqlDatabase) CreateAdminFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

func (d PsqlDatabase) DeleteAdminField(ctx context.Context, ac audited.AuditContext, id types.AdminFieldID) error {
	cmd := d.DeleteAdminFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdbp.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d PsqlDatabase) ListAdminFields() (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbp.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminFieldsByParentIDPaginated(params ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentIDPaginated(d.Context, mdbp.ListAdminFieldByParentIDPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by parent paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminField(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldParams) (*string, error) {
	cmd := d.UpdateAdminFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

type NewAdminFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminFieldCmd) Context() context.Context              { return c.ctx }
func (c NewAdminFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c NewAdminFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminFieldCmd) TableName() string                     { return "admin_fields" }
func (c NewAdminFieldCmd) Params() any                           { return c.params }
func (c NewAdminFieldCmd) GetID(u mdb.AdminFields) string        { return string(u.AdminFieldID) }

func (c NewAdminFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminFields, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminField(ctx, mdb.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
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

func (d Database) NewAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminFieldParams) NewAdminFieldCmd {
	return NewAdminFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

type UpdateAdminFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminFieldCmd) Context() context.Context              { return c.ctx }
func (c UpdateAdminFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminFieldCmd) TableName() string                     { return "admin_fields" }
func (c UpdateAdminFieldCmd) Params() any                           { return c.params }
func (c UpdateAdminFieldCmd) GetID() string                         { return string(c.params.AdminFieldID) }

func (c UpdateAdminFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminField(ctx, mdb.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

func (c UpdateAdminFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminField(ctx, mdb.UpdateAdminFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		AdminFieldID: c.params.AdminFieldID,
	})
}

func (d Database) UpdateAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldParams) UpdateAdminFieldCmd {
	return UpdateAdminFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteAdminFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminFieldCmd) Context() context.Context              { return c.ctx }
func (c DeleteAdminFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminFieldCmd) TableName() string                     { return "admin_fields" }
func (c DeleteAdminFieldCmd) GetID() string                         { return string(c.id) }

func (c DeleteAdminFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminField(ctx, mdb.GetAdminFieldParams{AdminFieldID: c.id})
}

func (c DeleteAdminFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminField(ctx, mdb.DeleteAdminFieldParams{AdminFieldID: c.id})
}

func (d Database) DeleteAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminFieldID) DeleteAdminFieldCmd {
	return DeleteAdminFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewAdminFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c NewAdminFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminFieldCmdMysql) TableName() string                     { return "admin_fields" }
func (c NewAdminFieldCmdMysql) Params() any                           { return c.params }
func (c NewAdminFieldCmdMysql) GetID(u mdbm.AdminFields) string      { return string(u.AdminFieldID) }

func (c NewAdminFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminFields, error) {
	queries := mdbm.New(tx)
	id := types.NewAdminFieldID()
	err := queries.CreateAdminField(ctx, mdbm.CreateAdminFieldParams{
		AdminFieldID: id,
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
	if err != nil {
		return mdbm.AdminFields{}, err
	}
	return queries.GetAdminField(ctx, mdbm.GetAdminFieldParams{AdminFieldID: id})
}

func (d MysqlDatabase) NewAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminFieldParams) NewAdminFieldCmdMysql {
	return NewAdminFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

type UpdateAdminFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateAdminFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminFieldCmdMysql) TableName() string                     { return "admin_fields" }
func (c UpdateAdminFieldCmdMysql) Params() any                           { return c.params }
func (c UpdateAdminFieldCmdMysql) GetID() string                         { return string(c.params.AdminFieldID) }

func (c UpdateAdminFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminField(ctx, mdbm.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

func (c UpdateAdminFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminField(ctx, mdbm.UpdateAdminFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		AdminFieldID: c.params.AdminFieldID,
	})
}

func (d MysqlDatabase) UpdateAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldParams) UpdateAdminFieldCmdMysql {
	return UpdateAdminFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteAdminFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteAdminFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminFieldCmdMysql) TableName() string                     { return "admin_fields" }
func (c DeleteAdminFieldCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteAdminFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminField(ctx, mdbm.GetAdminFieldParams{AdminFieldID: c.id})
}

func (c DeleteAdminFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminField(ctx, mdbm.DeleteAdminFieldParams{AdminFieldID: c.id})
}

func (d MysqlDatabase) DeleteAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminFieldID) DeleteAdminFieldCmdMysql {
	return DeleteAdminFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewAdminFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c NewAdminFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminFieldCmdPsql) TableName() string                     { return "admin_fields" }
func (c NewAdminFieldCmdPsql) Params() any                           { return c.params }
func (c NewAdminFieldCmdPsql) GetID(u mdbp.AdminFields) string      { return string(u.AdminFieldID) }

func (c NewAdminFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminFields, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminField(ctx, mdbp.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
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

func (d PsqlDatabase) NewAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminFieldParams) NewAdminFieldCmdPsql {
	return NewAdminFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

type UpdateAdminFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateAdminFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminFieldCmdPsql) TableName() string                     { return "admin_fields" }
func (c UpdateAdminFieldCmdPsql) Params() any                           { return c.params }
func (c UpdateAdminFieldCmdPsql) GetID() string                         { return string(c.params.AdminFieldID) }

func (c UpdateAdminFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminField(ctx, mdbp.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

func (c UpdateAdminFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminField(ctx, mdbp.UpdateAdminFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		AdminFieldID: c.params.AdminFieldID,
	})
}

func (d PsqlDatabase) UpdateAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldParams) UpdateAdminFieldCmdPsql {
	return UpdateAdminFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteAdminFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteAdminFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminFieldCmdPsql) TableName() string                     { return "admin_fields" }
func (c DeleteAdminFieldCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteAdminFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminField(ctx, mdbp.GetAdminFieldParams{AdminFieldID: c.id})
}

func (c DeleteAdminFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminField(ctx, mdbp.DeleteAdminFieldParams{AdminFieldID: c.id})
}

func (d PsqlDatabase) DeleteAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminFieldID) DeleteAdminFieldCmdPsql {
	return DeleteAdminFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
