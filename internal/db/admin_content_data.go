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

type AdminContentData struct {
	AdminContentDataID types.AdminContentID          `json:"admin_content_data_id"`
	ParentID           types.NullableContentID       `json:"parent_id"`
	FirstChildID       sql.NullString                `json:"first_child_id"`
	NextSiblingID      sql.NullString                `json:"next_sibling_id"`
	PrevSiblingID      sql.NullString                `json:"prev_sibling_id"`
	AdminRouteID       string                        `json:"admin_route_id"`
	AdminDatatypeID    types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           types.NullableUserID          `json:"author_id"`
	Status             types.ContentStatus           `json:"status"`
	DateCreated        types.Timestamp               `json:"date_created"`
	DateModified       types.Timestamp               `json:"date_modified"`
}
type CreateAdminContentDataParams struct {
	ParentID        types.NullableContentID       `json:"parent_id"`
	FirstChildID    sql.NullString                `json:"first_child_id"`
	NextSiblingID   sql.NullString                `json:"next_sibling_id"`
	PrevSiblingID   sql.NullString                `json:"prev_sibling_id"`
	AdminRouteID    string                        `json:"admin_route_id"`
	AdminDatatypeID types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID        types.NullableUserID          `json:"author_id"`
	Status          types.ContentStatus           `json:"status"`
	DateCreated     types.Timestamp               `json:"date_created"`
	DateModified    types.Timestamp               `json:"date_modified"`
}
type UpdateAdminContentDataParams struct {
	ParentID           types.NullableContentID       `json:"parent_id"`
	FirstChildID       sql.NullString                `json:"first_child_id"`
	NextSiblingID      sql.NullString                `json:"next_sibling_id"`
	PrevSiblingID      sql.NullString                `json:"prev_sibling_id"`
	AdminRouteID       string                        `json:"admin_route_id"`
	AdminDatatypeID    types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           types.NullableUserID          `json:"author_id"`
	Status             types.ContentStatus           `json:"status"`
	DateCreated        types.Timestamp               `json:"date_created"`
	DateModified       types.Timestamp               `json:"date_modified"`
	AdminContentDataID types.AdminContentID          `json:"admin_content_data_id"`
}
// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapAdminContentDataJSON converts AdminContentData to ContentDataJSON for tree building.
// Maps admin IDs into the public ContentDataJSON shape so BuildNodes works unchanged.
func MapAdminContentDataJSON(a AdminContentData) ContentDataJSON {
	firstChildID := ""
	if a.FirstChildID.Valid {
		firstChildID = a.FirstChildID.String
	}
	nextSiblingID := ""
	if a.NextSiblingID.Valid {
		nextSiblingID = a.NextSiblingID.String
	}
	prevSiblingID := ""
	if a.PrevSiblingID.Valid {
		prevSiblingID = a.PrevSiblingID.String
	}
	return ContentDataJSON{
		ContentDataID: a.AdminContentDataID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  firstChildID,
		NextSiblingID: nextSiblingID,
		PrevSiblingID: prevSiblingID,
		RouteID:       a.AdminRouteID,
		DatatypeID:    a.AdminDatatypeID.String(),
		AuthorID:      a.AuthorID.String(),
		Status:        string(a.Status),
		DateCreated:   a.DateCreated.String(),
		DateModified:  a.DateModified.String(),
	}
}

// MapStringAdminContentData converts AdminContentData to StringAdminContentData for table display
func MapStringAdminContentData(a AdminContentData) StringAdminContentData {
	return StringAdminContentData{
		AdminContentDataID: a.AdminContentDataID.String(),
		ParentID:           a.ParentID.String(),
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID.String(),
		AuthorID:           a.AuthorID.String(),
		Status:             string(a.Status),
		DateCreated:        a.DateCreated.String(),
		DateModified:       a.DateModified.String(),
		History:            "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminContentData(a mdb.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}

func (d Database) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdb.CreateAdminContentDataParams {
	return mdb.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}

func (d Database) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdb.UpdateAdminContentDataParams {
	return mdb.UpdateAdminContentDataParams{
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		AdminContentDataID: a.AdminContentDataID,
	}
}


// QUERIES

func (d Database) CountAdminContentData() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d Database) CreateAdminContentData(ctx context.Context, ac audited.AuditContext, s CreateAdminContentDataParams) (*AdminContentData, error) {
	cmd := d.NewAdminContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentData: %w", err)
	}
	r := d.MapAdminContentData(result)
	return &r, nil
}
func (d Database) CreateAdminContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d Database) DeleteAdminContentData(ctx context.Context, ac audited.AuditContext, id types.AdminContentID) error {
	cmd := d.DeleteAdminContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d Database) GetAdminContentData(id types.AdminContentID) (*AdminContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, mdb.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d Database) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListAdminContentDataByRoute(id string) (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, mdb.ListAdminContentDataByRouteParams{AdminRouteID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) UpdateAdminContentData(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentDataParams) (*string, error) {
	cmd := d.UpdateAdminContentDataCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentData: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentDataID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminContentData(a mdbm.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d MysqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbm.CreateAdminContentDataParams {
	return mdbm.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d MysqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbm.UpdateAdminContentDataParams {
	return mdbm.UpdateAdminContentDataParams{
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		AdminContentDataID: a.AdminContentDataID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminContentData() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateAdminContentData(ctx context.Context, ac audited.AuditContext, s CreateAdminContentDataParams) (*AdminContentData, error) {
	cmd := d.NewAdminContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentData: %w", err)
	}
	r := d.MapAdminContentData(result)
	return &r, nil
}
func (d MysqlDatabase) CreateAdminContentDataTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d MysqlDatabase) DeleteAdminContentData(ctx context.Context, ac audited.AuditContext, id types.AdminContentID) error {
	cmd := d.DeleteAdminContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d MysqlDatabase) GetAdminContentData(id types.AdminContentID) (*AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, mdbm.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentDataByRoute(id string) (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, mdbm.ListAdminContentDataByRouteParams{AdminRouteID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateAdminContentData(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentDataParams) (*string, error) {
	cmd := d.UpdateAdminContentDataCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentData: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentDataID)
	return &msg, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS

func (d PsqlDatabase) MapAdminContentData(a mdbp.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d PsqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbp.CreateAdminContentDataParams {
	return mdbp.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d PsqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbp.UpdateAdminContentDataParams {
	return mdbp.UpdateAdminContentDataParams{
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		AdminContentDataID: a.AdminContentDataID,
	}
}

// /QUERIES

func (d PsqlDatabase) CountAdminContentData() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateAdminContentData(ctx context.Context, ac audited.AuditContext, s CreateAdminContentDataParams) (*AdminContentData, error) {
	cmd := d.NewAdminContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentData: %w", err)
	}
	r := d.MapAdminContentData(result)
	return &r, nil
}
func (d PsqlDatabase) CreateAdminContentDataTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminContentData(ctx context.Context, ac audited.AuditContext, id types.AdminContentID) error {
	cmd := d.DeleteAdminContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d PsqlDatabase) GetAdminContentData(id types.AdminContentID) (*AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, mdbp.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentDataByRoute(id string) (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, mdbp.ListAdminContentDataByRouteParams{AdminRouteID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) UpdateAdminContentData(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentDataParams) (*string, error) {
	cmd := d.UpdateAdminContentDataCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentData: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentDataID)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewAdminContentDataCmd is an audited create command for admin_content_data (SQLite).
type NewAdminContentDataCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentDataCmd) Context() context.Context              { return c.ctx }
func (c NewAdminContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentDataCmd) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentDataCmd) TableName() string                     { return "admin_content_data" }
func (c NewAdminContentDataCmd) Params() any                           { return c.params }
func (c NewAdminContentDataCmd) GetID(row mdb.AdminContentData) string {
	return string(row.AdminContentDataID)
}

func (c NewAdminContentDataCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminContentData, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminContentData(ctx, mdb.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           c.params.ParentID,
		FirstChildID:       c.params.FirstChildID,
		NextSiblingID:      c.params.NextSiblingID,
		PrevSiblingID:      c.params.PrevSiblingID,
		AdminRouteID:       c.params.AdminRouteID,
		AdminDatatypeID:    c.params.AdminDatatypeID,
		AuthorID:           c.params.AuthorID,
		Status:             c.params.Status,
		DateCreated:        c.params.DateCreated,
		DateModified:       c.params.DateModified,
	})
}

func (d Database) NewAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentDataParams) NewAdminContentDataCmd {
	return NewAdminContentDataCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateAdminContentDataCmd is an audited update command for admin_content_data (SQLite).
type UpdateAdminContentDataCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentDataCmd) Context() context.Context              { return c.ctx }
func (c UpdateAdminContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminContentDataCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminContentDataCmd) TableName() string                     { return "admin_content_data" }
func (c UpdateAdminContentDataCmd) Params() any                           { return c.params }
func (c UpdateAdminContentDataCmd) GetID() string {
	return string(c.params.AdminContentDataID)
}

func (c UpdateAdminContentDataCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentData, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentData(ctx, mdb.GetAdminContentDataParams{AdminContentDataID: c.params.AdminContentDataID})
}

func (c UpdateAdminContentDataCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminContentData(ctx, mdb.UpdateAdminContentDataParams{
		ParentID:           c.params.ParentID,
		FirstChildID:       c.params.FirstChildID,
		NextSiblingID:      c.params.NextSiblingID,
		PrevSiblingID:      c.params.PrevSiblingID,
		AdminRouteID:       c.params.AdminRouteID,
		AdminDatatypeID:    c.params.AdminDatatypeID,
		AuthorID:           c.params.AuthorID,
		Status:             c.params.Status,
		DateCreated:        c.params.DateCreated,
		DateModified:       c.params.DateModified,
		AdminContentDataID: c.params.AdminContentDataID,
	})
}

func (d Database) UpdateAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentDataParams) UpdateAdminContentDataCmd {
	return UpdateAdminContentDataCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// DeleteAdminContentDataCmd is an audited delete command for admin_content_data (SQLite).
type DeleteAdminContentDataCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentDataCmd) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentDataCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentDataCmd) TableName() string                     { return "admin_content_data" }
func (c DeleteAdminContentDataCmd) GetID() string                         { return string(c.id) }

func (c DeleteAdminContentDataCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentData, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentData(ctx, mdb.GetAdminContentDataParams{AdminContentDataID: c.id})
}

func (c DeleteAdminContentDataCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminContentData(ctx, mdb.DeleteAdminContentDataParams{AdminContentDataID: c.id})
}

func (d Database) DeleteAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentID) DeleteAdminContentDataCmd {
	return DeleteAdminContentDataCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

// NewAdminContentDataCmdMysql is an audited create command for admin_content_data (MySQL).
type NewAdminContentDataCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentDataCmdMysql) Context() context.Context              { return c.ctx }
func (c NewAdminContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentDataCmdMysql) TableName() string                     { return "admin_content_data" }
func (c NewAdminContentDataCmdMysql) Params() any                           { return c.params }
func (c NewAdminContentDataCmdMysql) GetID(row mdbm.AdminContentData) string {
	return string(row.AdminContentDataID)
}

func (c NewAdminContentDataCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentData, error) {
	id := types.NewAdminContentID()
	queries := mdbm.New(tx)
	err := queries.CreateAdminContentData(ctx, mdbm.CreateAdminContentDataParams{
		AdminContentDataID: id,
		ParentID:           c.params.ParentID,
		FirstChildID:       c.params.FirstChildID,
		NextSiblingID:      c.params.NextSiblingID,
		PrevSiblingID:      c.params.PrevSiblingID,
		AdminRouteID:       c.params.AdminRouteID,
		AdminDatatypeID:    c.params.AdminDatatypeID,
		AuthorID:           c.params.AuthorID,
		Status:             c.params.Status,
		DateCreated:        c.params.DateCreated,
		DateModified:       c.params.DateModified,
	})
	if err != nil {
		return mdbm.AdminContentData{}, fmt.Errorf("execute create admin_content_data: %w", err)
	}
	return queries.GetAdminContentData(ctx, mdbm.GetAdminContentDataParams{AdminContentDataID: id})
}

func (d MysqlDatabase) NewAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentDataParams) NewAdminContentDataCmdMysql {
	return NewAdminContentDataCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateAdminContentDataCmdMysql is an audited update command for admin_content_data (MySQL).
type UpdateAdminContentDataCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentDataCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateAdminContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminContentDataCmdMysql) TableName() string                     { return "admin_content_data" }
func (c UpdateAdminContentDataCmdMysql) Params() any                           { return c.params }
func (c UpdateAdminContentDataCmdMysql) GetID() string {
	return string(c.params.AdminContentDataID)
}

func (c UpdateAdminContentDataCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentData, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentData(ctx, mdbm.GetAdminContentDataParams{AdminContentDataID: c.params.AdminContentDataID})
}

func (c UpdateAdminContentDataCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminContentData(ctx, mdbm.UpdateAdminContentDataParams{
		ParentID:           c.params.ParentID,
		FirstChildID:       c.params.FirstChildID,
		NextSiblingID:      c.params.NextSiblingID,
		PrevSiblingID:      c.params.PrevSiblingID,
		AdminRouteID:       c.params.AdminRouteID,
		AdminDatatypeID:    c.params.AdminDatatypeID,
		AuthorID:           c.params.AuthorID,
		Status:             c.params.Status,
		DateCreated:        c.params.DateCreated,
		DateModified:       c.params.DateModified,
		AdminContentDataID: c.params.AdminContentDataID,
	})
}

func (d MysqlDatabase) UpdateAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentDataParams) UpdateAdminContentDataCmdMysql {
	return UpdateAdminContentDataCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// DeleteAdminContentDataCmdMysql is an audited delete command for admin_content_data (MySQL).
type DeleteAdminContentDataCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentDataCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentDataCmdMysql) TableName() string                     { return "admin_content_data" }
func (c DeleteAdminContentDataCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteAdminContentDataCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentData, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentData(ctx, mdbm.GetAdminContentDataParams{AdminContentDataID: c.id})
}

func (c DeleteAdminContentDataCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminContentData(ctx, mdbm.DeleteAdminContentDataParams{AdminContentDataID: c.id})
}

func (d MysqlDatabase) DeleteAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentID) DeleteAdminContentDataCmdMysql {
	return DeleteAdminContentDataCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

// NewAdminContentDataCmdPsql is an audited create command for admin_content_data (PostgreSQL).
type NewAdminContentDataCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentDataCmdPsql) Context() context.Context              { return c.ctx }
func (c NewAdminContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentDataCmdPsql) TableName() string                     { return "admin_content_data" }
func (c NewAdminContentDataCmdPsql) Params() any                           { return c.params }
func (c NewAdminContentDataCmdPsql) GetID(row mdbp.AdminContentData) string {
	return string(row.AdminContentDataID)
}

func (c NewAdminContentDataCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentData, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminContentData(ctx, mdbp.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           c.params.ParentID,
		FirstChildID:       c.params.FirstChildID,
		NextSiblingID:      c.params.NextSiblingID,
		PrevSiblingID:      c.params.PrevSiblingID,
		AdminRouteID:       c.params.AdminRouteID,
		AdminDatatypeID:    c.params.AdminDatatypeID,
		AuthorID:           c.params.AuthorID,
		Status:             c.params.Status,
		DateCreated:        c.params.DateCreated,
		DateModified:       c.params.DateModified,
	})
}

func (d PsqlDatabase) NewAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentDataParams) NewAdminContentDataCmdPsql {
	return NewAdminContentDataCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateAdminContentDataCmdPsql is an audited update command for admin_content_data (PostgreSQL).
type UpdateAdminContentDataCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentDataCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateAdminContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateAdminContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateAdminContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateAdminContentDataCmdPsql) TableName() string                     { return "admin_content_data" }
func (c UpdateAdminContentDataCmdPsql) Params() any                           { return c.params }
func (c UpdateAdminContentDataCmdPsql) GetID() string {
	return string(c.params.AdminContentDataID)
}

func (c UpdateAdminContentDataCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentData, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentData(ctx, mdbp.GetAdminContentDataParams{AdminContentDataID: c.params.AdminContentDataID})
}

func (c UpdateAdminContentDataCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminContentData(ctx, mdbp.UpdateAdminContentDataParams{
		ParentID:           c.params.ParentID,
		FirstChildID:       c.params.FirstChildID,
		NextSiblingID:      c.params.NextSiblingID,
		PrevSiblingID:      c.params.PrevSiblingID,
		AdminRouteID:       c.params.AdminRouteID,
		AdminDatatypeID:    c.params.AdminDatatypeID,
		AuthorID:           c.params.AuthorID,
		Status:             c.params.Status,
		DateCreated:        c.params.DateCreated,
		DateModified:       c.params.DateModified,
		AdminContentDataID: c.params.AdminContentDataID,
	})
}

func (d PsqlDatabase) UpdateAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentDataParams) UpdateAdminContentDataCmdPsql {
	return UpdateAdminContentDataCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// DeleteAdminContentDataCmdPsql is an audited delete command for admin_content_data (PostgreSQL).
type DeleteAdminContentDataCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentDataCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentDataCmdPsql) TableName() string                     { return "admin_content_data" }
func (c DeleteAdminContentDataCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteAdminContentDataCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentData, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentData(ctx, mdbp.GetAdminContentDataParams{AdminContentDataID: c.id})
}

func (c DeleteAdminContentDataCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminContentData(ctx, mdbp.DeleteAdminContentDataParams{AdminContentDataID: c.id})
}

func (d PsqlDatabase) DeleteAdminContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentID) DeleteAdminContentDataCmdPsql {
	return DeleteAdminContentDataCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
