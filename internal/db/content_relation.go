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

type ContentRelations struct {
	ContentRelationID types.ContentRelationID `json:"content_relation_id"`
	SourceContentID   types.ContentID         `json:"source_content_id"`
	TargetContentID   types.ContentID         `json:"target_content_id"`
	FieldID           types.FieldID           `json:"field_id"`
	SortOrder         int64                   `json:"sort_order"`
	DateCreated       types.Timestamp         `json:"date_created"`
}

type CreateContentRelationParams struct {
	SourceContentID types.ContentID `json:"source_content_id"`
	TargetContentID types.ContentID `json:"target_content_id"`
	FieldID         types.FieldID   `json:"field_id"`
	SortOrder       int64           `json:"sort_order"`
	DateCreated     types.Timestamp `json:"date_created"`
}

type UpdateContentRelationSortOrderParams struct {
	ContentRelationID types.ContentRelationID `json:"content_relation_id"`
	SortOrder         int64                   `json:"sort_order"`
}

// StringContentRelations is the string representation for TUI table display
type StringContentRelations struct {
	ContentRelationID string `json:"content_relation_id"`
	SourceContentID   string `json:"source_content_id"`
	TargetContentID   string `json:"target_content_id"`
	FieldID           string `json:"field_id"`
	SortOrder         string `json:"sort_order"`
	DateCreated       string `json:"date_created"`
}

// MapStringContentRelation converts ContentRelations to StringContentRelations for table display
func MapStringContentRelation(a ContentRelations) StringContentRelations {
	return StringContentRelations{
		ContentRelationID: a.ContentRelationID.String(),
		SourceContentID:   a.SourceContentID.String(),
		TargetContentID:   a.TargetContentID.String(),
		FieldID:           a.FieldID.String(),
		SortOrder:         fmt.Sprintf("%d", a.SortOrder),
		DateCreated:       a.DateCreated.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapContentRelation(a mdb.ContentRelations) ContentRelations {
	return ContentRelations{
		ContentRelationID: a.ContentRelationID,
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID.ID,
		SortOrder:         a.SortOrder,
		DateCreated:       a.DateCreated,
	}
}

func (d Database) MapCreateContentRelationParams(a CreateContentRelationParams) mdb.CreateContentRelationParams {
	return mdb.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           types.NullableFieldID{ID: a.FieldID, Valid: !a.FieldID.IsZero()},
		SortOrder:         a.SortOrder,
		DateCreated:       a.DateCreated,
	}
}

func (d Database) MapUpdateContentRelationSortOrderParams(a UpdateContentRelationSortOrderParams) mdb.UpdateContentRelationSortOrderParams {
	return mdb.UpdateContentRelationSortOrderParams{
		SortOrder:         a.SortOrder,
		ContentRelationID: a.ContentRelationID,
	}
}

// QUERIES

func (d Database) CountContentRelations() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content relations: %w", err)
	}
	return &c, nil
}

func (d Database) CreateContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateContentRelationTable(d.Context)
}

func (d Database) DropContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropContentRelationTable(d.Context)
}

func (d Database) CreateContentRelation(ctx context.Context, ac audited.AuditContext, s CreateContentRelationParams) (*ContentRelations, error) {
	cmd := d.NewContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content relation: %w", err)
	}
	r := d.MapContentRelation(result)
	return &r, nil
}

func (d Database) DeleteContentRelation(ctx context.Context, ac audited.AuditContext, id types.ContentRelationID) error {
	cmd := d.DeleteContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) UpdateContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateContentRelationSortOrderParams) error {
	cmd := d.UpdateContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d Database) GetContentRelation(id types.ContentRelationID) (*ContentRelations, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentRelation(d.Context, mdb.GetContentRelationParams{ContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content relation: %w", err)
	}
	res := d.MapContentRelation(row)
	return &res, nil
}

func (d Database) ListContentRelationsBySource(id types.ContentID) (*[]ContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentRelationsBySource(d.Context, mdb.ListContentRelationsBySourceParams{SourceContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by source: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

func (d Database) ListContentRelationsByTarget(id types.ContentID) (*[]ContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentRelationsByTarget(d.Context, mdb.ListContentRelationsByTargetParams{TargetContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by target: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

func (d Database) ListContentRelationsBySourceAndField(contentID types.ContentID, fieldID types.FieldID) (*[]ContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentRelationsBySourceAndField(d.Context, mdb.ListContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		FieldID:         types.NullableFieldID{ID: fieldID, Valid: !fieldID.IsZero()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by source and field: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapContentRelation(a mdbm.ContentRelations) ContentRelations {
	return ContentRelations{
		ContentRelationID: a.ContentRelationID,
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID.ID,
		SortOrder:         int64(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

func (d MysqlDatabase) MapCreateContentRelationParams(a CreateContentRelationParams) mdbm.CreateContentRelationParams {
	return mdbm.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           types.NullableFieldID{ID: a.FieldID, Valid: !a.FieldID.IsZero()},
		SortOrder:         int32(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

func (d MysqlDatabase) MapUpdateContentRelationSortOrderParams(a UpdateContentRelationSortOrderParams) mdbm.UpdateContentRelationSortOrderParams {
	return mdbm.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(a.SortOrder),
		ContentRelationID: a.ContentRelationID,
	}
}

// QUERIES

func (d MysqlDatabase) CountContentRelations() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content relations: %w", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateContentRelationTable(d.Context)
}

func (d MysqlDatabase) DropContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropContentRelationTable(d.Context)
}

func (d MysqlDatabase) CreateContentRelation(ctx context.Context, ac audited.AuditContext, s CreateContentRelationParams) (*ContentRelations, error) {
	cmd := d.NewContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content relation: %w", err)
	}
	r := d.MapContentRelation(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteContentRelation(ctx context.Context, ac audited.AuditContext, id types.ContentRelationID) error {
	cmd := d.DeleteContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) UpdateContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateContentRelationSortOrderParams) error {
	cmd := d.UpdateContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d MysqlDatabase) GetContentRelation(id types.ContentRelationID) (*ContentRelations, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentRelation(d.Context, mdbm.GetContentRelationParams{ContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content relation: %w", err)
	}
	res := d.MapContentRelation(row)
	return &res, nil
}

func (d MysqlDatabase) ListContentRelationsBySource(id types.ContentID) (*[]ContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentRelationsBySource(d.Context, mdbm.ListContentRelationsBySourceParams{SourceContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by source: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentRelationsByTarget(id types.ContentID) (*[]ContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentRelationsByTarget(d.Context, mdbm.ListContentRelationsByTargetParams{TargetContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by target: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentRelationsBySourceAndField(contentID types.ContentID, fieldID types.FieldID) (*[]ContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentRelationsBySourceAndField(d.Context, mdbm.ListContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		FieldID:         types.NullableFieldID{ID: fieldID, Valid: !fieldID.IsZero()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by source and field: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapContentRelation(a mdbp.ContentRelations) ContentRelations {
	return ContentRelations{
		ContentRelationID: a.ContentRelationID,
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID.ID,
		SortOrder:         int64(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

func (d PsqlDatabase) MapCreateContentRelationParams(a CreateContentRelationParams) mdbp.CreateContentRelationParams {
	return mdbp.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           types.NullableFieldID{ID: a.FieldID, Valid: !a.FieldID.IsZero()},
		SortOrder:         int32(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

func (d PsqlDatabase) MapUpdateContentRelationSortOrderParams(a UpdateContentRelationSortOrderParams) mdbp.UpdateContentRelationSortOrderParams {
	return mdbp.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(a.SortOrder),
		ContentRelationID: a.ContentRelationID,
	}
}

// QUERIES

func (d PsqlDatabase) CountContentRelations() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content relations: %w", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateContentRelationTable(d.Context)
}

func (d PsqlDatabase) DropContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropContentRelationTable(d.Context)
}

func (d PsqlDatabase) CreateContentRelation(ctx context.Context, ac audited.AuditContext, s CreateContentRelationParams) (*ContentRelations, error) {
	cmd := d.NewContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content relation: %w", err)
	}
	r := d.MapContentRelation(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteContentRelation(ctx context.Context, ac audited.AuditContext, id types.ContentRelationID) error {
	cmd := d.DeleteContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) UpdateContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateContentRelationSortOrderParams) error {
	cmd := d.UpdateContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d PsqlDatabase) GetContentRelation(id types.ContentRelationID) (*ContentRelations, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentRelation(d.Context, mdbp.GetContentRelationParams{ContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content relation: %w", err)
	}
	res := d.MapContentRelation(row)
	return &res, nil
}

func (d PsqlDatabase) ListContentRelationsBySource(id types.ContentID) (*[]ContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentRelationsBySource(d.Context, mdbp.ListContentRelationsBySourceParams{SourceContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by source: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentRelationsByTarget(id types.ContentID) (*[]ContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentRelationsByTarget(d.Context, mdbp.ListContentRelationsByTargetParams{TargetContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by target: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentRelationsBySourceAndField(contentID types.ContentID, fieldID types.FieldID) (*[]ContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentRelationsBySourceAndField(d.Context, mdbp.ListContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		FieldID:         types.NullableFieldID{ID: fieldID, Valid: !fieldID.IsZero()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content relations by source and field: %w", err)
	}
	res := []ContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapContentRelation(v))
	}
	return &res, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

type NewContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewContentRelationCmd) Context() context.Context              { return c.ctx }
func (c NewContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentRelationCmd) Connection() *sql.DB                   { return c.conn }
func (c NewContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentRelationCmd) TableName() string                     { return "content_relations" }
func (c NewContentRelationCmd) Params() any                           { return c.params }
func (c NewContentRelationCmd) GetID(r mdb.ContentRelations) string {
	return string(r.ContentRelationID)
}

func (c NewContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.ContentRelations, error) {
	queries := mdb.New(tx)
	return queries.CreateContentRelation(ctx, mdb.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   c.params.SourceContentID,
		TargetContentID:   c.params.TargetContentID,
		FieldID:           types.NullableFieldID{ID: c.params.FieldID, Valid: !c.params.FieldID.IsZero()},
		SortOrder:         c.params.SortOrder,
		DateCreated:       c.params.DateCreated,
	})
}

func (d Database) NewContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentRelationParams) NewContentRelationCmd {
	return NewContentRelationCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE (SortOrder) -----

type UpdateContentRelationSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateContentRelationSortOrderCmd) Context() context.Context  { return c.ctx }
func (c UpdateContentRelationSortOrderCmd) AuditContext() audited.AuditContext {
	return c.auditCtx
}
func (c UpdateContentRelationSortOrderCmd) Connection() *sql.DB { return c.conn }
func (c UpdateContentRelationSortOrderCmd) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateContentRelationSortOrderCmd) TableName() string { return "content_relations" }
func (c UpdateContentRelationSortOrderCmd) Params() any       { return c.params }
func (c UpdateContentRelationSortOrderCmd) GetID() string {
	return string(c.params.ContentRelationID)
}

func (c UpdateContentRelationSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetContentRelation(ctx, mdb.GetContentRelationParams{ContentRelationID: c.params.ContentRelationID})
}

func (c UpdateContentRelationSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateContentRelationSortOrder(ctx, mdb.UpdateContentRelationSortOrderParams{
		SortOrder:         c.params.SortOrder,
		ContentRelationID: c.params.ContentRelationID,
	})
}

func (d Database) UpdateContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentRelationSortOrderParams) UpdateContentRelationSortOrderCmd {
	return UpdateContentRelationSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteContentRelationCmd) Context() context.Context              { return c.ctx }
func (c DeleteContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentRelationCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentRelationCmd) TableName() string                     { return "content_relations" }
func (c DeleteContentRelationCmd) GetID() string                         { return string(c.id) }

func (c DeleteContentRelationCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetContentRelation(ctx, mdb.GetContentRelationParams{ContentRelationID: c.id})
}

func (c DeleteContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteContentRelation(ctx, mdb.DeleteContentRelationParams{ContentRelationID: c.id})
}

func (d Database) DeleteContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentRelationID) DeleteContentRelationCmd {
	return DeleteContentRelationCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewContentRelationCmdMysql) Context() context.Context              { return c.ctx }
func (c NewContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentRelationCmdMysql) TableName() string                     { return "content_relations" }
func (c NewContentRelationCmdMysql) Params() any                           { return c.params }
func (c NewContentRelationCmdMysql) GetID(r mdbm.ContentRelations) string {
	return string(r.ContentRelationID)
}

func (c NewContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.ContentRelations, error) {
	id := types.NewContentRelationID()
	queries := mdbm.New(tx)
	params := mdbm.CreateContentRelationParams{
		ContentRelationID: id,
		SourceContentID:   c.params.SourceContentID,
		TargetContentID:   c.params.TargetContentID,
		FieldID:           types.NullableFieldID{ID: c.params.FieldID, Valid: !c.params.FieldID.IsZero()},
		SortOrder:         int32(c.params.SortOrder),
		DateCreated:       c.params.DateCreated,
	}
	if err := queries.CreateContentRelation(ctx, params); err != nil {
		return mdbm.ContentRelations{}, err
	}
	return queries.GetContentRelation(ctx, mdbm.GetContentRelationParams{ContentRelationID: id})
}

func (d MysqlDatabase) NewContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentRelationParams) NewContentRelationCmdMysql {
	return NewContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE (SortOrder) -----

type UpdateContentRelationSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateContentRelationSortOrderCmdMysql) Context() context.Context { return c.ctx }
func (c UpdateContentRelationSortOrderCmdMysql) AuditContext() audited.AuditContext {
	return c.auditCtx
}
func (c UpdateContentRelationSortOrderCmdMysql) Connection() *sql.DB { return c.conn }
func (c UpdateContentRelationSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateContentRelationSortOrderCmdMysql) TableName() string { return "content_relations" }
func (c UpdateContentRelationSortOrderCmdMysql) Params() any       { return c.params }
func (c UpdateContentRelationSortOrderCmdMysql) GetID() string {
	return string(c.params.ContentRelationID)
}

func (c UpdateContentRelationSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetContentRelation(ctx, mdbm.GetContentRelationParams{ContentRelationID: c.params.ContentRelationID})
}

func (c UpdateContentRelationSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateContentRelationSortOrder(ctx, mdbm.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(c.params.SortOrder),
		ContentRelationID: c.params.ContentRelationID,
	})
}

func (d MysqlDatabase) UpdateContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentRelationSortOrderParams) UpdateContentRelationSortOrderCmdMysql {
	return UpdateContentRelationSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteContentRelationCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentRelationCmdMysql) TableName() string { return "content_relations" }
func (c DeleteContentRelationCmdMysql) GetID() string     { return string(c.id) }

func (c DeleteContentRelationCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetContentRelation(ctx, mdbm.GetContentRelationParams{ContentRelationID: c.id})
}

func (c DeleteContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteContentRelation(ctx, mdbm.DeleteContentRelationParams{ContentRelationID: c.id})
}

func (d MysqlDatabase) DeleteContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentRelationID) DeleteContentRelationCmdMysql {
	return DeleteContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewContentRelationCmdPsql) Context() context.Context              { return c.ctx }
func (c NewContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentRelationCmdPsql) TableName() string                     { return "content_relations" }
func (c NewContentRelationCmdPsql) Params() any                           { return c.params }
func (c NewContentRelationCmdPsql) GetID(r mdbp.ContentRelations) string {
	return string(r.ContentRelationID)
}

func (c NewContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.ContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.CreateContentRelation(ctx, mdbp.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   c.params.SourceContentID,
		TargetContentID:   c.params.TargetContentID,
		FieldID:           types.NullableFieldID{ID: c.params.FieldID, Valid: !c.params.FieldID.IsZero()},
		SortOrder:         int32(c.params.SortOrder),
		DateCreated:       c.params.DateCreated,
	})
}

func (d PsqlDatabase) NewContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentRelationParams) NewContentRelationCmdPsql {
	return NewContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE (SortOrder) -----

type UpdateContentRelationSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateContentRelationSortOrderCmdPsql) Context() context.Context { return c.ctx }
func (c UpdateContentRelationSortOrderCmdPsql) AuditContext() audited.AuditContext {
	return c.auditCtx
}
func (c UpdateContentRelationSortOrderCmdPsql) Connection() *sql.DB { return c.conn }
func (c UpdateContentRelationSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateContentRelationSortOrderCmdPsql) TableName() string { return "content_relations" }
func (c UpdateContentRelationSortOrderCmdPsql) Params() any       { return c.params }
func (c UpdateContentRelationSortOrderCmdPsql) GetID() string {
	return string(c.params.ContentRelationID)
}

func (c UpdateContentRelationSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetContentRelation(ctx, mdbp.GetContentRelationParams{ContentRelationID: c.params.ContentRelationID})
}

func (c UpdateContentRelationSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateContentRelationSortOrder(ctx, mdbp.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(c.params.SortOrder),
		ContentRelationID: c.params.ContentRelationID,
	})
}

func (d PsqlDatabase) UpdateContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentRelationSortOrderParams) UpdateContentRelationSortOrderCmdPsql {
	return UpdateContentRelationSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteContentRelationCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentRelationCmdPsql) TableName() string { return "content_relations" }
func (c DeleteContentRelationCmdPsql) GetID() string     { return string(c.id) }

func (c DeleteContentRelationCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetContentRelation(ctx, mdbp.GetContentRelationParams{ContentRelationID: c.id})
}

func (c DeleteContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteContentRelation(ctx, mdbp.DeleteContentRelationParams{ContentRelationID: c.id})
}

func (d PsqlDatabase) DeleteContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentRelationID) DeleteContentRelationCmdPsql {
	return DeleteContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
