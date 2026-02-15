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

// AdminContentRelations represents a relation between admin content items.
type AdminContentRelations struct {
	AdminContentRelationID types.AdminContentRelationID `json:"admin_content_relation_id"`
	SourceContentID        types.AdminContentID         `json:"source_content_id"`
	TargetContentID        types.AdminContentID         `json:"target_content_id"`
	AdminFieldID           types.AdminFieldID           `json:"admin_field_id"`
	SortOrder              int64                        `json:"sort_order"`
	DateCreated            types.Timestamp              `json:"date_created"`
}

// CreateAdminContentRelationParams contains fields for creating a new admin content relation.
type CreateAdminContentRelationParams struct {
	SourceContentID types.AdminContentID `json:"source_content_id"`
	TargetContentID types.AdminContentID `json:"target_content_id"`
	AdminFieldID    types.AdminFieldID   `json:"admin_field_id"`
	SortOrder       int64                `json:"sort_order"`
	DateCreated     types.Timestamp      `json:"date_created"`
}

// UpdateAdminContentRelationSortOrderParams contains fields for updating the sort order of an admin content relation.
type UpdateAdminContentRelationSortOrderParams struct {
	AdminContentRelationID types.AdminContentRelationID `json:"admin_content_relation_id"`
	SortOrder              int64                        `json:"sort_order"`
}

// StringAdminContentRelations is the string representation for TUI table display.
type StringAdminContentRelations struct {
	AdminContentRelationID string `json:"admin_content_relation_id"`
	SourceContentID        string `json:"source_content_id"`
	TargetContentID        string `json:"target_content_id"`
	AdminFieldID           string `json:"admin_field_id"`
	SortOrder              string `json:"sort_order"`
	DateCreated            string `json:"date_created"`
}

// MapStringAdminContentRelation converts AdminContentRelations to StringAdminContentRelations for table display.
func MapStringAdminContentRelation(a AdminContentRelations) StringAdminContentRelations {
	return StringAdminContentRelations{
		AdminContentRelationID: a.AdminContentRelationID.String(),
		SourceContentID:        a.SourceContentID.String(),
		TargetContentID:        a.TargetContentID.String(),
		AdminFieldID:           a.AdminFieldID.String(),
		SortOrder:              fmt.Sprintf("%d", a.SortOrder),
		DateCreated:            a.DateCreated.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapAdminContentRelation converts a sqlc-generated type to the wrapper type.
func (d Database) MapAdminContentRelation(a mdb.AdminContentRelations) AdminContentRelations {
	return AdminContentRelations{
		AdminContentRelationID: a.AdminContentRelationID,
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID,
		SortOrder:              a.SortOrder,
		DateCreated:            a.DateCreated,
	}
}

// MapCreateAdminContentRelationParams converts a sqlc-generated type to the wrapper type.
func (d Database) MapCreateAdminContentRelationParams(a CreateAdminContentRelationParams) mdb.CreateAdminContentRelationParams {
	return mdb.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID,
		SortOrder:              a.SortOrder,
		DateCreated:            a.DateCreated,
	}
}

// MapUpdateAdminContentRelationSortOrderParams converts a sqlc-generated type to the wrapper type.
func (d Database) MapUpdateAdminContentRelationSortOrderParams(a UpdateAdminContentRelationSortOrderParams) mdb.UpdateAdminContentRelationSortOrderParams {
	return mdb.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              a.SortOrder,
		AdminContentRelationID: a.AdminContentRelationID,
	}
}

// QUERIES

// CountAdminContentRelations returns the total count of admin content relations.
func (d Database) CountAdminContentRelations() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content relations: %w", err)
	}
	return &c, nil
}

// CreateAdminContentRelationTable creates the admin_content_relations table.
func (d Database) CreateAdminContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateAdminContentRelationTable(d.Context)
}

// DropAdminContentRelationTable drops the admin_content_relations table.
func (d Database) DropAdminContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropAdminContentRelationTable(d.Context)
}

// CreateAdminContentRelation inserts a new admin content relation.
func (d Database) CreateAdminContentRelation(ctx context.Context, ac audited.AuditContext, s CreateAdminContentRelationParams) (*AdminContentRelations, error) {
	cmd := d.NewAdminContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content relation: %w", err)
	}
	r := d.MapAdminContentRelation(result)
	return &r, nil
}

// DeleteAdminContentRelation removes a record.
func (d Database) DeleteAdminContentRelation(ctx context.Context, ac audited.AuditContext, id types.AdminContentRelationID) error {
	cmd := d.DeleteAdminContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateAdminContentRelationSortOrder modifies the sort order of an admin content relation.
func (d Database) UpdateAdminContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentRelationSortOrderParams) error {
	cmd := d.UpdateAdminContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetAdminContentRelation retrieves a record by ID.
func (d Database) GetAdminContentRelation(id types.AdminContentRelationID) (*AdminContentRelations, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentRelation(d.Context, mdb.GetAdminContentRelationParams{AdminContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content relation: %w", err)
	}
	res := d.MapAdminContentRelation(row)
	return &res, nil
}

// ListAdminContentRelationsBySource returns all admin content relations for a source content item.
func (d Database) ListAdminContentRelationsBySource(id types.AdminContentID) (*[]AdminContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySource(d.Context, mdb.ListAdminContentRelationsBySourceParams{SourceContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by source: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

// ListAdminContentRelationsByTarget returns all admin content relations for a target content item.
func (d Database) ListAdminContentRelationsByTarget(id types.AdminContentID) (*[]AdminContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsByTarget(d.Context, mdb.ListAdminContentRelationsByTargetParams{TargetContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by target: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

// ListAdminContentRelationsBySourceAndField returns all admin content relations for a source content item and field.
func (d Database) ListAdminContentRelationsBySourceAndField(contentID types.AdminContentID, fieldID types.AdminFieldID) (*[]AdminContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySourceAndField(d.Context, mdb.ListAdminContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		AdminFieldID:    fieldID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by source and field: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapAdminContentRelation converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapAdminContentRelation(a mdbm.AdminContentRelations) AdminContentRelations {
	return AdminContentRelations{
		AdminContentRelationID: a.AdminContentRelationID,
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID,
		SortOrder:              int64(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

// MapCreateAdminContentRelationParams converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapCreateAdminContentRelationParams(a CreateAdminContentRelationParams) mdbm.CreateAdminContentRelationParams {
	return mdbm.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID,
		SortOrder:              int32(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

// MapUpdateAdminContentRelationSortOrderParams converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapUpdateAdminContentRelationSortOrderParams(a UpdateAdminContentRelationSortOrderParams) mdbm.UpdateAdminContentRelationSortOrderParams {
	return mdbm.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(a.SortOrder),
		AdminContentRelationID: a.AdminContentRelationID,
	}
}

// QUERIES

// CountAdminContentRelations returns the total count of admin content relations.
func (d MysqlDatabase) CountAdminContentRelations() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content relations: %w", err)
	}
	return &c, nil
}

// CreateAdminContentRelationTable creates the admin_content_relations table.
func (d MysqlDatabase) CreateAdminContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateAdminContentRelationTable(d.Context)
}

// DropAdminContentRelationTable drops the admin_content_relations table.
func (d MysqlDatabase) DropAdminContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropAdminContentRelationTable(d.Context)
}

// CreateAdminContentRelation inserts a new admin content relation.
func (d MysqlDatabase) CreateAdminContentRelation(ctx context.Context, ac audited.AuditContext, s CreateAdminContentRelationParams) (*AdminContentRelations, error) {
	cmd := d.NewAdminContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content relation: %w", err)
	}
	r := d.MapAdminContentRelation(result)
	return &r, nil
}

// DeleteAdminContentRelation removes a record.
func (d MysqlDatabase) DeleteAdminContentRelation(ctx context.Context, ac audited.AuditContext, id types.AdminContentRelationID) error {
	cmd := d.DeleteAdminContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateAdminContentRelationSortOrder modifies the sort order of an admin content relation.
func (d MysqlDatabase) UpdateAdminContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentRelationSortOrderParams) error {
	cmd := d.UpdateAdminContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetAdminContentRelation retrieves a record by ID.
func (d MysqlDatabase) GetAdminContentRelation(id types.AdminContentRelationID) (*AdminContentRelations, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentRelation(d.Context, mdbm.GetAdminContentRelationParams{AdminContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content relation: %w", err)
	}
	res := d.MapAdminContentRelation(row)
	return &res, nil
}

// ListAdminContentRelationsBySource returns all admin content relations for a source content item.
func (d MysqlDatabase) ListAdminContentRelationsBySource(id types.AdminContentID) (*[]AdminContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySource(d.Context, mdbm.ListAdminContentRelationsBySourceParams{SourceContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by source: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

// ListAdminContentRelationsByTarget returns all admin content relations for a target content item.
func (d MysqlDatabase) ListAdminContentRelationsByTarget(id types.AdminContentID) (*[]AdminContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsByTarget(d.Context, mdbm.ListAdminContentRelationsByTargetParams{TargetContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by target: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

// ListAdminContentRelationsBySourceAndField returns all admin content relations for a source content item and field.
func (d MysqlDatabase) ListAdminContentRelationsBySourceAndField(contentID types.AdminContentID, fieldID types.AdminFieldID) (*[]AdminContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySourceAndField(d.Context, mdbm.ListAdminContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		AdminFieldID:    fieldID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by source and field: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapAdminContentRelation converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapAdminContentRelation(a mdbp.AdminContentRelations) AdminContentRelations {
	return AdminContentRelations{
		AdminContentRelationID: a.AdminContentRelationID,
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID,
		SortOrder:              int64(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

// MapCreateAdminContentRelationParams converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapCreateAdminContentRelationParams(a CreateAdminContentRelationParams) mdbp.CreateAdminContentRelationParams {
	return mdbp.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID,
		SortOrder:              int32(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

// MapUpdateAdminContentRelationSortOrderParams converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapUpdateAdminContentRelationSortOrderParams(a UpdateAdminContentRelationSortOrderParams) mdbp.UpdateAdminContentRelationSortOrderParams {
	return mdbp.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(a.SortOrder),
		AdminContentRelationID: a.AdminContentRelationID,
	}
}

// QUERIES

// CountAdminContentRelations returns the total count of admin content relations.
func (d PsqlDatabase) CountAdminContentRelations() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content relations: %w", err)
	}
	return &c, nil
}

// CreateAdminContentRelationTable creates the admin_content_relations table.
func (d PsqlDatabase) CreateAdminContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateAdminContentRelationTable(d.Context)
}

// DropAdminContentRelationTable drops the admin_content_relations table.
func (d PsqlDatabase) DropAdminContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropAdminContentRelationTable(d.Context)
}

// CreateAdminContentRelation inserts a new admin content relation.
func (d PsqlDatabase) CreateAdminContentRelation(ctx context.Context, ac audited.AuditContext, s CreateAdminContentRelationParams) (*AdminContentRelations, error) {
	cmd := d.NewAdminContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content relation: %w", err)
	}
	r := d.MapAdminContentRelation(result)
	return &r, nil
}

// DeleteAdminContentRelation removes a record.
func (d PsqlDatabase) DeleteAdminContentRelation(ctx context.Context, ac audited.AuditContext, id types.AdminContentRelationID) error {
	cmd := d.DeleteAdminContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateAdminContentRelationSortOrder modifies the sort order of an admin content relation.
func (d PsqlDatabase) UpdateAdminContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentRelationSortOrderParams) error {
	cmd := d.UpdateAdminContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetAdminContentRelation retrieves a record by ID.
func (d PsqlDatabase) GetAdminContentRelation(id types.AdminContentRelationID) (*AdminContentRelations, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentRelation(d.Context, mdbp.GetAdminContentRelationParams{AdminContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content relation: %w", err)
	}
	res := d.MapAdminContentRelation(row)
	return &res, nil
}

// ListAdminContentRelationsBySource returns all admin content relations for a source content item.
func (d PsqlDatabase) ListAdminContentRelationsBySource(id types.AdminContentID) (*[]AdminContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySource(d.Context, mdbp.ListAdminContentRelationsBySourceParams{SourceContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by source: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

// ListAdminContentRelationsByTarget returns all admin content relations for a target content item.
func (d PsqlDatabase) ListAdminContentRelationsByTarget(id types.AdminContentID) (*[]AdminContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsByTarget(d.Context, mdbp.ListAdminContentRelationsByTargetParams{TargetContentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by target: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

// ListAdminContentRelationsBySourceAndField returns all admin content relations for a source content item and field.
func (d PsqlDatabase) ListAdminContentRelationsBySourceAndField(contentID types.AdminContentID, fieldID types.AdminFieldID) (*[]AdminContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySourceAndField(d.Context, mdbp.ListAdminContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		AdminFieldID:    fieldID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content relations by source and field: %w", err)
	}
	res := []AdminContentRelations{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentRelation(v))
	}
	return &res, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

// NewAdminContentRelationCmd is an audited command for create on admin_content_relations.
type NewAdminContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c NewAdminContentRelationCmd) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c NewAdminContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c NewAdminContentRelationCmd) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c NewAdminContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c NewAdminContentRelationCmd) TableName() string { return "admin_content_relations" }
// Params returns the command parameters.
func (c NewAdminContentRelationCmd) Params() any       { return c.params }
// GetID returns the record ID.
func (c NewAdminContentRelationCmd) GetID(r mdb.AdminContentRelations) string {
	return string(r.AdminContentRelationID)
}

// Execute performs the create operation.
func (c NewAdminContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminContentRelations, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminContentRelation(ctx, mdb.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        c.params.SourceContentID,
		TargetContentID:        c.params.TargetContentID,
		AdminFieldID:           c.params.AdminFieldID,
		SortOrder:              c.params.SortOrder,
		DateCreated:            c.params.DateCreated,
	})
}

// NewAdminContentRelationCmd creates a new audited command for create.
func (d Database) NewAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentRelationParams) NewAdminContentRelationCmd {
	return NewAdminContentRelationCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE (SortOrder) -----

// UpdateAdminContentRelationSortOrderCmd is an audited command for update on admin_content_relations.
type UpdateAdminContentRelationSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c UpdateAdminContentRelationSortOrderCmd) Context() context.Context { return c.ctx }
// AuditContext returns the audit context.
func (c UpdateAdminContentRelationSortOrderCmd) AuditContext() audited.AuditContext {
	return c.auditCtx
}
// Connection returns the database connection.
func (c UpdateAdminContentRelationSortOrderCmd) Connection() *sql.DB { return c.conn }
// Recorder returns the change event recorder.
func (c UpdateAdminContentRelationSortOrderCmd) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
// TableName returns the table name.
func (c UpdateAdminContentRelationSortOrderCmd) TableName() string {
	return "admin_content_relations"
}
// Params returns the command parameters.
func (c UpdateAdminContentRelationSortOrderCmd) Params() any { return c.params }
// GetID returns the record ID.
func (c UpdateAdminContentRelationSortOrderCmd) GetID() string {
	return string(c.params.AdminContentRelationID)
}

// GetBefore retrieves the record before modification.
func (c UpdateAdminContentRelationSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentRelation(ctx, mdb.GetAdminContentRelationParams{AdminContentRelationID: c.params.AdminContentRelationID})
}

// Execute performs the update operation.
func (c UpdateAdminContentRelationSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminContentRelationSortOrder(ctx, mdb.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              c.params.SortOrder,
		AdminContentRelationID: c.params.AdminContentRelationID,
	})
}

// UpdateAdminContentRelationSortOrderCmd creates a new audited command for update.
func (d Database) UpdateAdminContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentRelationSortOrderParams) UpdateAdminContentRelationSortOrderCmd {
	return UpdateAdminContentRelationSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteAdminContentRelationCmd is an audited command for delete on admin_content_relations.
type DeleteAdminContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c DeleteAdminContentRelationCmd) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c DeleteAdminContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c DeleteAdminContentRelationCmd) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c DeleteAdminContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c DeleteAdminContentRelationCmd) TableName() string { return "admin_content_relations" }
// GetID returns the record ID.
func (c DeleteAdminContentRelationCmd) GetID() string     { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminContentRelationCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentRelation(ctx, mdb.GetAdminContentRelationParams{AdminContentRelationID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminContentRelation(ctx, mdb.DeleteAdminContentRelationParams{AdminContentRelationID: c.id})
}

// DeleteAdminContentRelationCmd creates a new audited command for delete.
func (d Database) DeleteAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentRelationID) DeleteAdminContentRelationCmd {
	return DeleteAdminContentRelationCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewAdminContentRelationCmdMysql is an audited command for create on admin_content_relations (MySQL).
type NewAdminContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c NewAdminContentRelationCmdMysql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c NewAdminContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c NewAdminContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c NewAdminContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c NewAdminContentRelationCmdMysql) TableName() string { return "admin_content_relations" }
// Params returns the command parameters.
func (c NewAdminContentRelationCmdMysql) Params() any       { return c.params }
// GetID returns the record ID.
func (c NewAdminContentRelationCmdMysql) GetID(r mdbm.AdminContentRelations) string {
	return string(r.AdminContentRelationID)
}

// Execute performs the create operation.
func (c NewAdminContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentRelations, error) {
	id := types.NewAdminContentRelationID()
	queries := mdbm.New(tx)
	params := mdbm.CreateAdminContentRelationParams{
		AdminContentRelationID: id,
		SourceContentID:        c.params.SourceContentID,
		TargetContentID:        c.params.TargetContentID,
		AdminFieldID:           c.params.AdminFieldID,
		SortOrder:              int32(c.params.SortOrder),
		DateCreated:            c.params.DateCreated,
	}
	if err := queries.CreateAdminContentRelation(ctx, params); err != nil {
		return mdbm.AdminContentRelations{}, err
	}
	return queries.GetAdminContentRelation(ctx, mdbm.GetAdminContentRelationParams{AdminContentRelationID: id})
}

// NewAdminContentRelationCmd creates a new audited command for create.
func (d MysqlDatabase) NewAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentRelationParams) NewAdminContentRelationCmdMysql {
	return NewAdminContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE (SortOrder) -----

// UpdateAdminContentRelationSortOrderCmdMysql is an audited command for update on admin_content_relations (MySQL).
type UpdateAdminContentRelationSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c UpdateAdminContentRelationSortOrderCmdMysql) Context() context.Context { return c.ctx }
// AuditContext returns the audit context.
func (c UpdateAdminContentRelationSortOrderCmdMysql) AuditContext() audited.AuditContext {
	return c.auditCtx
}
// Connection returns the database connection.
func (c UpdateAdminContentRelationSortOrderCmdMysql) Connection() *sql.DB { return c.conn }
// Recorder returns the change event recorder.
func (c UpdateAdminContentRelationSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
// TableName returns the table name.
func (c UpdateAdminContentRelationSortOrderCmdMysql) TableName() string {
	return "admin_content_relations"
}
// Params returns the command parameters.
func (c UpdateAdminContentRelationSortOrderCmdMysql) Params() any { return c.params }
// GetID returns the record ID.
func (c UpdateAdminContentRelationSortOrderCmdMysql) GetID() string {
	return string(c.params.AdminContentRelationID)
}

// GetBefore retrieves the record before modification.
func (c UpdateAdminContentRelationSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbm.GetAdminContentRelationParams{AdminContentRelationID: c.params.AdminContentRelationID})
}

// Execute performs the update operation.
func (c UpdateAdminContentRelationSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminContentRelationSortOrder(ctx, mdbm.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(c.params.SortOrder),
		AdminContentRelationID: c.params.AdminContentRelationID,
	})
}

// UpdateAdminContentRelationSortOrderCmd creates a new audited command for update.
func (d MysqlDatabase) UpdateAdminContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentRelationSortOrderParams) UpdateAdminContentRelationSortOrderCmdMysql {
	return UpdateAdminContentRelationSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteAdminContentRelationCmdMysql is an audited command for delete on admin_content_relations (MySQL).
type DeleteAdminContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c DeleteAdminContentRelationCmdMysql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c DeleteAdminContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c DeleteAdminContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c DeleteAdminContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c DeleteAdminContentRelationCmdMysql) TableName() string { return "admin_content_relations" }
// GetID returns the record ID.
func (c DeleteAdminContentRelationCmdMysql) GetID() string     { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminContentRelationCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbm.GetAdminContentRelationParams{AdminContentRelationID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminContentRelation(ctx, mdbm.DeleteAdminContentRelationParams{AdminContentRelationID: c.id})
}

// DeleteAdminContentRelationCmd creates a new audited command for delete.
func (d MysqlDatabase) DeleteAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentRelationID) DeleteAdminContentRelationCmdMysql {
	return DeleteAdminContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewAdminContentRelationCmdPsql is an audited command for create on admin_content_relations (PostgreSQL).
type NewAdminContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c NewAdminContentRelationCmdPsql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c NewAdminContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c NewAdminContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c NewAdminContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c NewAdminContentRelationCmdPsql) TableName() string { return "admin_content_relations" }
// Params returns the command parameters.
func (c NewAdminContentRelationCmdPsql) Params() any       { return c.params }
// GetID returns the record ID.
func (c NewAdminContentRelationCmdPsql) GetID(r mdbp.AdminContentRelations) string {
	return string(r.AdminContentRelationID)
}

// Execute performs the create operation.
func (c NewAdminContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminContentRelation(ctx, mdbp.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        c.params.SourceContentID,
		TargetContentID:        c.params.TargetContentID,
		AdminFieldID:           c.params.AdminFieldID,
		SortOrder:              int32(c.params.SortOrder),
		DateCreated:            c.params.DateCreated,
	})
}

// NewAdminContentRelationCmd creates a new audited command for create.
func (d PsqlDatabase) NewAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentRelationParams) NewAdminContentRelationCmdPsql {
	return NewAdminContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE (SortOrder) -----

// UpdateAdminContentRelationSortOrderCmdPsql is an audited command for update on admin_content_relations (PostgreSQL).
type UpdateAdminContentRelationSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c UpdateAdminContentRelationSortOrderCmdPsql) Context() context.Context { return c.ctx }
// AuditContext returns the audit context.
func (c UpdateAdminContentRelationSortOrderCmdPsql) AuditContext() audited.AuditContext {
	return c.auditCtx
}
// Connection returns the database connection.
func (c UpdateAdminContentRelationSortOrderCmdPsql) Connection() *sql.DB { return c.conn }
// Recorder returns the change event recorder.
func (c UpdateAdminContentRelationSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
// TableName returns the table name.
func (c UpdateAdminContentRelationSortOrderCmdPsql) TableName() string {
	return "admin_content_relations"
}
// Params returns the command parameters.
func (c UpdateAdminContentRelationSortOrderCmdPsql) Params() any { return c.params }
// GetID returns the record ID.
func (c UpdateAdminContentRelationSortOrderCmdPsql) GetID() string {
	return string(c.params.AdminContentRelationID)
}

// GetBefore retrieves the record before modification.
func (c UpdateAdminContentRelationSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbp.GetAdminContentRelationParams{AdminContentRelationID: c.params.AdminContentRelationID})
}

// Execute performs the update operation.
func (c UpdateAdminContentRelationSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminContentRelationSortOrder(ctx, mdbp.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(c.params.SortOrder),
		AdminContentRelationID: c.params.AdminContentRelationID,
	})
}

// UpdateAdminContentRelationSortOrderCmd creates a new audited command for update.
func (d PsqlDatabase) UpdateAdminContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentRelationSortOrderParams) UpdateAdminContentRelationSortOrderCmdPsql {
	return UpdateAdminContentRelationSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteAdminContentRelationCmdPsql is an audited command for delete on admin_content_relations (PostgreSQL).
type DeleteAdminContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c DeleteAdminContentRelationCmdPsql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c DeleteAdminContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c DeleteAdminContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c DeleteAdminContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c DeleteAdminContentRelationCmdPsql) TableName() string { return "admin_content_relations" }
// GetID returns the record ID.
func (c DeleteAdminContentRelationCmdPsql) GetID() string     { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminContentRelationCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbp.GetAdminContentRelationParams{AdminContentRelationID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminContentRelation(ctx, mdbp.DeleteAdminContentRelationParams{AdminContentRelationID: c.id})
}

// DeleteAdminContentRelationCmd creates a new audited command for delete.
func (d PsqlDatabase) DeleteAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentRelationID) DeleteAdminContentRelationCmdPsql {
	return DeleteAdminContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
