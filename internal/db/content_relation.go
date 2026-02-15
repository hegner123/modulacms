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

// ContentRelations represents a content relation connecting two content items through a field.
type ContentRelations struct {
	ContentRelationID types.ContentRelationID `json:"content_relation_id"`
	SourceContentID   types.ContentID         `json:"source_content_id"`
	TargetContentID   types.ContentID         `json:"target_content_id"`
	FieldID           types.FieldID           `json:"field_id"`
	SortOrder         int64                   `json:"sort_order"`
	DateCreated       types.Timestamp         `json:"date_created"`
}

// CreateContentRelationParams specifies parameters for creating a content relation.
type CreateContentRelationParams struct {
	SourceContentID types.ContentID `json:"source_content_id"`
	TargetContentID types.ContentID `json:"target_content_id"`
	FieldID         types.FieldID   `json:"field_id"`
	SortOrder       int64           `json:"sort_order"`
	DateCreated     types.Timestamp `json:"date_created"`
}

// UpdateContentRelationSortOrderParams specifies parameters for updating a content relation's sort order.
type UpdateContentRelationSortOrderParams struct {
	ContentRelationID types.ContentRelationID `json:"content_relation_id"`
	SortOrder         int64                   `json:"sort_order"`
}

// StringContentRelations is the string representation of ContentRelations for TUI table display.
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

// MapContentRelation converts a sqlc-generated SQLite type to the wrapper type.
func (d Database) MapContentRelation(a mdb.ContentRelations) ContentRelations {
	return ContentRelations{
		ContentRelationID: a.ContentRelationID,
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID,
		SortOrder:         a.SortOrder,
		DateCreated:       a.DateCreated,
	}
}

// MapCreateContentRelationParams converts wrapper params to a sqlc-generated SQLite type.
func (d Database) MapCreateContentRelationParams(a CreateContentRelationParams) mdb.CreateContentRelationParams {
	return mdb.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID,
		SortOrder:         a.SortOrder,
		DateCreated:       a.DateCreated,
	}
}

// MapUpdateContentRelationSortOrderParams converts wrapper params to a sqlc-generated SQLite type.
func (d Database) MapUpdateContentRelationSortOrderParams(a UpdateContentRelationSortOrderParams) mdb.UpdateContentRelationSortOrderParams {
	return mdb.UpdateContentRelationSortOrderParams{
		SortOrder:         a.SortOrder,
		ContentRelationID: a.ContentRelationID,
	}
}

// QUERIES

// CountContentRelations returns the total count of content relations.
func (d Database) CountContentRelations() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content relations: %w", err)
	}
	return &c, nil
}

// CreateContentRelationTable creates the content_relations table.
func (d Database) CreateContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateContentRelationTable(d.Context)
}

// DropContentRelationTable drops the content_relations table.
func (d Database) DropContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropContentRelationTable(d.Context)
}

// CreateContentRelation creates a new content relation with audit trail.
func (d Database) CreateContentRelation(ctx context.Context, ac audited.AuditContext, s CreateContentRelationParams) (*ContentRelations, error) {
	cmd := d.NewContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content relation: %w", err)
	}
	r := d.MapContentRelation(result)
	return &r, nil
}

// DeleteContentRelation deletes a content relation with audit trail.
func (d Database) DeleteContentRelation(ctx context.Context, ac audited.AuditContext, id types.ContentRelationID) error {
	cmd := d.DeleteContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateContentRelationSortOrder updates a content relation's sort order with audit trail.
func (d Database) UpdateContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateContentRelationSortOrderParams) error {
	cmd := d.UpdateContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetContentRelation retrieves a content relation by ID.
func (d Database) GetContentRelation(id types.ContentRelationID) (*ContentRelations, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentRelation(d.Context, mdb.GetContentRelationParams{ContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content relation: %w", err)
	}
	res := d.MapContentRelation(row)
	return &res, nil
}

// ListContentRelationsBySource retrieves all relations for a given source content ID.
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

// ListContentRelationsByTarget retrieves all relations for a given target content ID.
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

// ListContentRelationsBySourceAndField retrieves relations for a given source content ID and field ID.
func (d Database) ListContentRelationsBySourceAndField(contentID types.ContentID, fieldID types.FieldID) (*[]ContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentRelationsBySourceAndField(d.Context, mdb.ListContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		FieldID:         fieldID,
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

// MapContentRelation converts a sqlc-generated MySQL type to the wrapper type.
func (d MysqlDatabase) MapContentRelation(a mdbm.ContentRelations) ContentRelations {
	return ContentRelations{
		ContentRelationID: a.ContentRelationID,
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID,
		SortOrder:         int64(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

// MapCreateContentRelationParams converts wrapper params to a sqlc-generated MySQL type.
func (d MysqlDatabase) MapCreateContentRelationParams(a CreateContentRelationParams) mdbm.CreateContentRelationParams {
	return mdbm.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID,
		SortOrder:         int32(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

// MapUpdateContentRelationSortOrderParams converts wrapper params to a sqlc-generated MySQL type.
func (d MysqlDatabase) MapUpdateContentRelationSortOrderParams(a UpdateContentRelationSortOrderParams) mdbm.UpdateContentRelationSortOrderParams {
	return mdbm.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(a.SortOrder),
		ContentRelationID: a.ContentRelationID,
	}
}

// QUERIES

// CountContentRelations returns the total count of content relations.
func (d MysqlDatabase) CountContentRelations() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content relations: %w", err)
	}
	return &c, nil
}

// CreateContentRelationTable creates the content_relations table.
func (d MysqlDatabase) CreateContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateContentRelationTable(d.Context)
}

// DropContentRelationTable drops the content_relations table.
func (d MysqlDatabase) DropContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropContentRelationTable(d.Context)
}

// CreateContentRelation creates a new content relation with audit trail.
func (d MysqlDatabase) CreateContentRelation(ctx context.Context, ac audited.AuditContext, s CreateContentRelationParams) (*ContentRelations, error) {
	cmd := d.NewContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content relation: %w", err)
	}
	r := d.MapContentRelation(result)
	return &r, nil
}

// DeleteContentRelation deletes a content relation with audit trail.
func (d MysqlDatabase) DeleteContentRelation(ctx context.Context, ac audited.AuditContext, id types.ContentRelationID) error {
	cmd := d.DeleteContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateContentRelationSortOrder updates a content relation's sort order with audit trail.
func (d MysqlDatabase) UpdateContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateContentRelationSortOrderParams) error {
	cmd := d.UpdateContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetContentRelation retrieves a content relation by ID.
func (d MysqlDatabase) GetContentRelation(id types.ContentRelationID) (*ContentRelations, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentRelation(d.Context, mdbm.GetContentRelationParams{ContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content relation: %w", err)
	}
	res := d.MapContentRelation(row)
	return &res, nil
}

// ListContentRelationsBySource retrieves all relations for a given source content ID.
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

// ListContentRelationsByTarget retrieves all relations for a given target content ID.
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

// ListContentRelationsBySourceAndField retrieves relations for a given source content ID and field ID.
func (d MysqlDatabase) ListContentRelationsBySourceAndField(contentID types.ContentID, fieldID types.FieldID) (*[]ContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentRelationsBySourceAndField(d.Context, mdbm.ListContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		FieldID:         fieldID,
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

// MapContentRelation converts a sqlc-generated PostgreSQL type to the wrapper type.
func (d PsqlDatabase) MapContentRelation(a mdbp.ContentRelations) ContentRelations {
	return ContentRelations{
		ContentRelationID: a.ContentRelationID,
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID,
		SortOrder:         int64(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

// MapCreateContentRelationParams converts wrapper params to a sqlc-generated PostgreSQL type.
func (d PsqlDatabase) MapCreateContentRelationParams(a CreateContentRelationParams) mdbp.CreateContentRelationParams {
	return mdbp.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   a.SourceContentID,
		TargetContentID:   a.TargetContentID,
		FieldID:           a.FieldID,
		SortOrder:         int32(a.SortOrder),
		DateCreated:       a.DateCreated,
	}
}

// MapUpdateContentRelationSortOrderParams converts wrapper params to a sqlc-generated PostgreSQL type.
func (d PsqlDatabase) MapUpdateContentRelationSortOrderParams(a UpdateContentRelationSortOrderParams) mdbp.UpdateContentRelationSortOrderParams {
	return mdbp.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(a.SortOrder),
		ContentRelationID: a.ContentRelationID,
	}
}

// QUERIES

// CountContentRelations returns the total count of content relations.
func (d PsqlDatabase) CountContentRelations() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content relations: %w", err)
	}
	return &c, nil
}

// CreateContentRelationTable creates the content_relations table.
func (d PsqlDatabase) CreateContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateContentRelationTable(d.Context)
}

// DropContentRelationTable drops the content_relations table.
func (d PsqlDatabase) DropContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropContentRelationTable(d.Context)
}

// CreateContentRelation creates a new content relation with audit trail.
func (d PsqlDatabase) CreateContentRelation(ctx context.Context, ac audited.AuditContext, s CreateContentRelationParams) (*ContentRelations, error) {
	cmd := d.NewContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content relation: %w", err)
	}
	r := d.MapContentRelation(result)
	return &r, nil
}

// DeleteContentRelation deletes a content relation with audit trail.
func (d PsqlDatabase) DeleteContentRelation(ctx context.Context, ac audited.AuditContext, id types.ContentRelationID) error {
	cmd := d.DeleteContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateContentRelationSortOrder updates a content relation's sort order with audit trail.
func (d PsqlDatabase) UpdateContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateContentRelationSortOrderParams) error {
	cmd := d.UpdateContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetContentRelation retrieves a content relation by ID.
func (d PsqlDatabase) GetContentRelation(id types.ContentRelationID) (*ContentRelations, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentRelation(d.Context, mdbp.GetContentRelationParams{ContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content relation: %w", err)
	}
	res := d.MapContentRelation(row)
	return &res, nil
}

// ListContentRelationsBySource retrieves all relations for a given source content ID.
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

// ListContentRelationsByTarget retrieves all relations for a given target content ID.
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

// ListContentRelationsBySourceAndField retrieves relations for a given source content ID and field ID.
func (d PsqlDatabase) ListContentRelationsBySourceAndField(contentID types.ContentID, fieldID types.FieldID) (*[]ContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentRelationsBySourceAndField(d.Context, mdbp.ListContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		FieldID:         fieldID,
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

// NewContentRelationCmd is an audited command for creating a content relation.
type NewContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewContentRelationCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentRelationCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewContentRelationCmd) TableName() string                     { return "content_relations" }

// Params returns the command parameters.
func (c NewContentRelationCmd) Params() any                           { return c.params }

// GetID returns the ID from a content relation.
func (c NewContentRelationCmd) GetID(r mdb.ContentRelations) string {
	return string(r.ContentRelationID)
}

// Execute creates the content relation in the database.
func (c NewContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.ContentRelations, error) {
	queries := mdb.New(tx)
	return queries.CreateContentRelation(ctx, mdb.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   c.params.SourceContentID,
		TargetContentID:   c.params.TargetContentID,
		FieldID:           c.params.FieldID,
		SortOrder:         c.params.SortOrder,
		DateCreated:       c.params.DateCreated,
	})
}

// NewContentRelationCmd creates a new create command for a content relation.
func (d Database) NewContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentRelationParams) NewContentRelationCmd {
	return NewContentRelationCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE (SortOrder) -----

// UpdateContentRelationSortOrderCmd is an audited command for updating a content relation's sort order.
type UpdateContentRelationSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateContentRelationSortOrderCmd) Context() context.Context  { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateContentRelationSortOrderCmd) AuditContext() audited.AuditContext {
	return c.auditCtx
}

// Connection returns the database connection.
func (c UpdateContentRelationSortOrderCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateContentRelationSortOrderCmd) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}

// TableName returns the table name.
func (c UpdateContentRelationSortOrderCmd) TableName() string { return "content_relations" }

// Params returns the command parameters.
func (c UpdateContentRelationSortOrderCmd) Params() any       { return c.params }

// GetID returns the content relation ID.
func (c UpdateContentRelationSortOrderCmd) GetID() string {
	return string(c.params.ContentRelationID)
}

// GetBefore retrieves the content relation before the update.
func (c UpdateContentRelationSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetContentRelation(ctx, mdb.GetContentRelationParams{ContentRelationID: c.params.ContentRelationID})
}

// Execute updates the content relation's sort order in the database.
func (c UpdateContentRelationSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateContentRelationSortOrder(ctx, mdb.UpdateContentRelationSortOrderParams{
		SortOrder:         c.params.SortOrder,
		ContentRelationID: c.params.ContentRelationID,
	})
}

// UpdateContentRelationSortOrderCmd creates a new update command for a content relation's sort order.
func (d Database) UpdateContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentRelationSortOrderParams) UpdateContentRelationSortOrderCmd {
	return UpdateContentRelationSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteContentRelationCmd is an audited command for deleting a content relation.
type DeleteContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteContentRelationCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentRelationCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteContentRelationCmd) TableName() string                     { return "content_relations" }

// GetID returns the content relation ID.
func (c DeleteContentRelationCmd) GetID() string                         { return string(c.id) }

// GetBefore retrieves the content relation before deletion.
func (c DeleteContentRelationCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetContentRelation(ctx, mdb.GetContentRelationParams{ContentRelationID: c.id})
}

// Execute deletes the content relation from the database.
func (c DeleteContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteContentRelation(ctx, mdb.DeleteContentRelationParams{ContentRelationID: c.id})
}

// DeleteContentRelationCmd creates a new delete command for a content relation.
func (d Database) DeleteContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentRelationID) DeleteContentRelationCmd {
	return DeleteContentRelationCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewContentRelationCmdMysql is an audited command for creating a content relation on MySQL.
type NewContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewContentRelationCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewContentRelationCmdMysql) TableName() string                     { return "content_relations" }

// Params returns the command parameters.
func (c NewContentRelationCmdMysql) Params() any                           { return c.params }

// GetID returns the ID from a content relation.
func (c NewContentRelationCmdMysql) GetID(r mdbm.ContentRelations) string {
	return string(r.ContentRelationID)
}

// Execute creates the content relation in the database.
func (c NewContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.ContentRelations, error) {
	id := types.NewContentRelationID()
	queries := mdbm.New(tx)
	params := mdbm.CreateContentRelationParams{
		ContentRelationID: id,
		SourceContentID:   c.params.SourceContentID,
		TargetContentID:   c.params.TargetContentID,
		FieldID:           c.params.FieldID,
		SortOrder:         int32(c.params.SortOrder),
		DateCreated:       c.params.DateCreated,
	}
	if err := queries.CreateContentRelation(ctx, params); err != nil {
		return mdbm.ContentRelations{}, err
	}
	return queries.GetContentRelation(ctx, mdbm.GetContentRelationParams{ContentRelationID: id})
}

// NewContentRelationCmd creates a new create command for a content relation.
func (d MysqlDatabase) NewContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentRelationParams) NewContentRelationCmdMysql {
	return NewContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE (SortOrder) -----

// UpdateContentRelationSortOrderCmdMysql is an audited command for updating a content relation's sort order on MySQL.
type UpdateContentRelationSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateContentRelationSortOrderCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateContentRelationSortOrderCmdMysql) AuditContext() audited.AuditContext {
	return c.auditCtx
}

// Connection returns the database connection.
func (c UpdateContentRelationSortOrderCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateContentRelationSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}

// TableName returns the table name.
func (c UpdateContentRelationSortOrderCmdMysql) TableName() string { return "content_relations" }

// Params returns the command parameters.
func (c UpdateContentRelationSortOrderCmdMysql) Params() any       { return c.params }

// GetID returns the content relation ID.
func (c UpdateContentRelationSortOrderCmdMysql) GetID() string {
	return string(c.params.ContentRelationID)
}

// GetBefore retrieves the content relation before the update.
func (c UpdateContentRelationSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetContentRelation(ctx, mdbm.GetContentRelationParams{ContentRelationID: c.params.ContentRelationID})
}

// Execute updates the content relation's sort order in the database.
func (c UpdateContentRelationSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateContentRelationSortOrder(ctx, mdbm.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(c.params.SortOrder),
		ContentRelationID: c.params.ContentRelationID,
	})
}

// UpdateContentRelationSortOrderCmd creates a new update command for a content relation's sort order.
func (d MysqlDatabase) UpdateContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentRelationSortOrderParams) UpdateContentRelationSortOrderCmdMysql {
	return UpdateContentRelationSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteContentRelationCmdMysql is an audited command for deleting a content relation on MySQL.
type DeleteContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteContentRelationCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteContentRelationCmdMysql) TableName() string { return "content_relations" }

// GetID returns the content relation ID.
func (c DeleteContentRelationCmdMysql) GetID() string     { return string(c.id) }

// GetBefore retrieves the content relation before deletion.
func (c DeleteContentRelationCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetContentRelation(ctx, mdbm.GetContentRelationParams{ContentRelationID: c.id})
}

// Execute deletes the content relation from the database.
func (c DeleteContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteContentRelation(ctx, mdbm.DeleteContentRelationParams{ContentRelationID: c.id})
}

// DeleteContentRelationCmd creates a new delete command for a content relation.
func (d MysqlDatabase) DeleteContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentRelationID) DeleteContentRelationCmdMysql {
	return DeleteContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewContentRelationCmdPsql is an audited command for creating a content relation on PostgreSQL.
type NewContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewContentRelationCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewContentRelationCmdPsql) TableName() string                     { return "content_relations" }

// Params returns the command parameters.
func (c NewContentRelationCmdPsql) Params() any                           { return c.params }

// GetID returns the ID from a content relation.
func (c NewContentRelationCmdPsql) GetID(r mdbp.ContentRelations) string {
	return string(r.ContentRelationID)
}

// Execute creates the content relation in the database.
func (c NewContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.ContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.CreateContentRelation(ctx, mdbp.CreateContentRelationParams{
		ContentRelationID: types.NewContentRelationID(),
		SourceContentID:   c.params.SourceContentID,
		TargetContentID:   c.params.TargetContentID,
		FieldID:           c.params.FieldID,
		SortOrder:         int32(c.params.SortOrder),
		DateCreated:       c.params.DateCreated,
	})
}

// NewContentRelationCmd creates a new create command for a content relation.
func (d PsqlDatabase) NewContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentRelationParams) NewContentRelationCmdPsql {
	return NewContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE (SortOrder) -----

// UpdateContentRelationSortOrderCmdPsql is an audited command for updating a content relation's sort order on PostgreSQL.
type UpdateContentRelationSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateContentRelationSortOrderCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateContentRelationSortOrderCmdPsql) AuditContext() audited.AuditContext {
	return c.auditCtx
}

// Connection returns the database connection.
func (c UpdateContentRelationSortOrderCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateContentRelationSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}

// TableName returns the table name.
func (c UpdateContentRelationSortOrderCmdPsql) TableName() string { return "content_relations" }

// Params returns the command parameters.
func (c UpdateContentRelationSortOrderCmdPsql) Params() any       { return c.params }

// GetID returns the content relation ID.
func (c UpdateContentRelationSortOrderCmdPsql) GetID() string {
	return string(c.params.ContentRelationID)
}

// GetBefore retrieves the content relation before the update.
func (c UpdateContentRelationSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetContentRelation(ctx, mdbp.GetContentRelationParams{ContentRelationID: c.params.ContentRelationID})
}

// Execute updates the content relation's sort order in the database.
func (c UpdateContentRelationSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateContentRelationSortOrder(ctx, mdbp.UpdateContentRelationSortOrderParams{
		SortOrder:         int32(c.params.SortOrder),
		ContentRelationID: c.params.ContentRelationID,
	})
}

// UpdateContentRelationSortOrderCmd creates a new update command for a content relation's sort order.
func (d PsqlDatabase) UpdateContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentRelationSortOrderParams) UpdateContentRelationSortOrderCmdPsql {
	return UpdateContentRelationSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteContentRelationCmdPsql is an audited command for deleting a content relation on PostgreSQL.
type DeleteContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteContentRelationCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteContentRelationCmdPsql) TableName() string { return "content_relations" }

// GetID returns the content relation ID.
func (c DeleteContentRelationCmdPsql) GetID() string     { return string(c.id) }

// GetBefore retrieves the content relation before deletion.
func (c DeleteContentRelationCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetContentRelation(ctx, mdbp.GetContentRelationParams{ContentRelationID: c.id})
}

// Execute deletes the content relation from the database.
func (c DeleteContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteContentRelation(ctx, mdbp.DeleteContentRelationParams{ContentRelationID: c.id})
}

// DeleteContentRelationCmd creates a new delete command for a content relation.
func (d PsqlDatabase) DeleteContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentRelationID) DeleteContentRelationCmdPsql {
	return DeleteContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
