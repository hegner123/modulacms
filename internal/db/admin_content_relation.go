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

type AdminContentRelations struct {
	AdminContentRelationID types.AdminContentRelationID `json:"admin_content_relation_id"`
	SourceContentID        types.AdminContentID         `json:"source_content_id"`
	TargetContentID        types.AdminContentID         `json:"target_content_id"`
	AdminFieldID           types.AdminFieldID           `json:"admin_field_id"`
	SortOrder              int64                        `json:"sort_order"`
	DateCreated            types.Timestamp              `json:"date_created"`
}

type CreateAdminContentRelationParams struct {
	SourceContentID types.AdminContentID `json:"source_content_id"`
	TargetContentID types.AdminContentID `json:"target_content_id"`
	AdminFieldID    types.AdminFieldID   `json:"admin_field_id"`
	SortOrder       int64                `json:"sort_order"`
	DateCreated     types.Timestamp      `json:"date_created"`
}

type UpdateAdminContentRelationSortOrderParams struct {
	AdminContentRelationID types.AdminContentRelationID `json:"admin_content_relation_id"`
	SortOrder              int64                        `json:"sort_order"`
}

// StringAdminContentRelations is the string representation for TUI table display
type StringAdminContentRelations struct {
	AdminContentRelationID string `json:"admin_content_relation_id"`
	SourceContentID        string `json:"source_content_id"`
	TargetContentID        string `json:"target_content_id"`
	AdminFieldID           string `json:"admin_field_id"`
	SortOrder              string `json:"sort_order"`
	DateCreated            string `json:"date_created"`
}

// MapStringAdminContentRelation converts AdminContentRelations to StringAdminContentRelations for table display
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

func (d Database) MapAdminContentRelation(a mdb.AdminContentRelations) AdminContentRelations {
	return AdminContentRelations{
		AdminContentRelationID: a.AdminContentRelationID,
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID.ID,
		SortOrder:              a.SortOrder,
		DateCreated:            a.DateCreated,
	}
}

func (d Database) MapCreateAdminContentRelationParams(a CreateAdminContentRelationParams) mdb.CreateAdminContentRelationParams {
	return mdb.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           types.NullableAdminFieldID{ID: a.AdminFieldID, Valid: !a.AdminFieldID.IsZero()},
		SortOrder:              a.SortOrder,
		DateCreated:            a.DateCreated,
	}
}

func (d Database) MapUpdateAdminContentRelationSortOrderParams(a UpdateAdminContentRelationSortOrderParams) mdb.UpdateAdminContentRelationSortOrderParams {
	return mdb.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              a.SortOrder,
		AdminContentRelationID: a.AdminContentRelationID,
	}
}

// QUERIES

func (d Database) CountAdminContentRelations() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content relations: %w", err)
	}
	return &c, nil
}

func (d Database) CreateAdminContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateAdminContentRelationTable(d.Context)
}

func (d Database) DropAdminContentRelationTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropAdminContentRelationTable(d.Context)
}

func (d Database) CreateAdminContentRelation(ctx context.Context, ac audited.AuditContext, s CreateAdminContentRelationParams) (*AdminContentRelations, error) {
	cmd := d.NewAdminContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content relation: %w", err)
	}
	r := d.MapAdminContentRelation(result)
	return &r, nil
}

func (d Database) DeleteAdminContentRelation(ctx context.Context, ac audited.AuditContext, id types.AdminContentRelationID) error {
	cmd := d.DeleteAdminContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) UpdateAdminContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentRelationSortOrderParams) error {
	cmd := d.UpdateAdminContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d Database) GetAdminContentRelation(id types.AdminContentRelationID) (*AdminContentRelations, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentRelation(d.Context, mdb.GetAdminContentRelationParams{AdminContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content relation: %w", err)
	}
	res := d.MapAdminContentRelation(row)
	return &res, nil
}

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

func (d Database) ListAdminContentRelationsBySourceAndField(contentID types.AdminContentID, fieldID types.AdminFieldID) (*[]AdminContentRelations, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySourceAndField(d.Context, mdb.ListAdminContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		AdminFieldID:    types.NullableAdminFieldID{ID: fieldID, Valid: !fieldID.IsZero()},
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

func (d MysqlDatabase) MapAdminContentRelation(a mdbm.AdminContentRelations) AdminContentRelations {
	return AdminContentRelations{
		AdminContentRelationID: a.AdminContentRelationID,
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID.ID,
		SortOrder:              int64(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

func (d MysqlDatabase) MapCreateAdminContentRelationParams(a CreateAdminContentRelationParams) mdbm.CreateAdminContentRelationParams {
	return mdbm.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           types.NullableAdminFieldID{ID: a.AdminFieldID, Valid: !a.AdminFieldID.IsZero()},
		SortOrder:              int32(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

func (d MysqlDatabase) MapUpdateAdminContentRelationSortOrderParams(a UpdateAdminContentRelationSortOrderParams) mdbm.UpdateAdminContentRelationSortOrderParams {
	return mdbm.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(a.SortOrder),
		AdminContentRelationID: a.AdminContentRelationID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminContentRelations() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content relations: %w", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateAdminContentRelationTable(d.Context)
}

func (d MysqlDatabase) DropAdminContentRelationTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropAdminContentRelationTable(d.Context)
}

func (d MysqlDatabase) CreateAdminContentRelation(ctx context.Context, ac audited.AuditContext, s CreateAdminContentRelationParams) (*AdminContentRelations, error) {
	cmd := d.NewAdminContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content relation: %w", err)
	}
	r := d.MapAdminContentRelation(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteAdminContentRelation(ctx context.Context, ac audited.AuditContext, id types.AdminContentRelationID) error {
	cmd := d.DeleteAdminContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) UpdateAdminContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentRelationSortOrderParams) error {
	cmd := d.UpdateAdminContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d MysqlDatabase) GetAdminContentRelation(id types.AdminContentRelationID) (*AdminContentRelations, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentRelation(d.Context, mdbm.GetAdminContentRelationParams{AdminContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content relation: %w", err)
	}
	res := d.MapAdminContentRelation(row)
	return &res, nil
}

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

func (d MysqlDatabase) ListAdminContentRelationsBySourceAndField(contentID types.AdminContentID, fieldID types.AdminFieldID) (*[]AdminContentRelations, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySourceAndField(d.Context, mdbm.ListAdminContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		AdminFieldID:    types.NullableAdminFieldID{ID: fieldID, Valid: !fieldID.IsZero()},
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

func (d PsqlDatabase) MapAdminContentRelation(a mdbp.AdminContentRelations) AdminContentRelations {
	return AdminContentRelations{
		AdminContentRelationID: a.AdminContentRelationID,
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           a.AdminFieldID.ID,
		SortOrder:              int64(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

func (d PsqlDatabase) MapCreateAdminContentRelationParams(a CreateAdminContentRelationParams) mdbp.CreateAdminContentRelationParams {
	return mdbp.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        a.SourceContentID,
		TargetContentID:        a.TargetContentID,
		AdminFieldID:           types.NullableAdminFieldID{ID: a.AdminFieldID, Valid: !a.AdminFieldID.IsZero()},
		SortOrder:              int32(a.SortOrder),
		DateCreated:            a.DateCreated,
	}
}

func (d PsqlDatabase) MapUpdateAdminContentRelationSortOrderParams(a UpdateAdminContentRelationSortOrderParams) mdbp.UpdateAdminContentRelationSortOrderParams {
	return mdbp.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(a.SortOrder),
		AdminContentRelationID: a.AdminContentRelationID,
	}
}

// QUERIES

func (d PsqlDatabase) CountAdminContentRelations() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentRelation(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content relations: %w", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateAdminContentRelationTable(d.Context)
}

func (d PsqlDatabase) DropAdminContentRelationTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropAdminContentRelationTable(d.Context)
}

func (d PsqlDatabase) CreateAdminContentRelation(ctx context.Context, ac audited.AuditContext, s CreateAdminContentRelationParams) (*AdminContentRelations, error) {
	cmd := d.NewAdminContentRelationCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content relation: %w", err)
	}
	r := d.MapAdminContentRelation(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteAdminContentRelation(ctx context.Context, ac audited.AuditContext, id types.AdminContentRelationID) error {
	cmd := d.DeleteAdminContentRelationCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) UpdateAdminContentRelationSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentRelationSortOrderParams) error {
	cmd := d.UpdateAdminContentRelationSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d PsqlDatabase) GetAdminContentRelation(id types.AdminContentRelationID) (*AdminContentRelations, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentRelation(d.Context, mdbp.GetAdminContentRelationParams{AdminContentRelationID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content relation: %w", err)
	}
	res := d.MapAdminContentRelation(row)
	return &res, nil
}

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

func (d PsqlDatabase) ListAdminContentRelationsBySourceAndField(contentID types.AdminContentID, fieldID types.AdminFieldID) (*[]AdminContentRelations, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentRelationsBySourceAndField(d.Context, mdbp.ListAdminContentRelationsBySourceAndFieldParams{
		SourceContentID: contentID,
		AdminFieldID:    types.NullableAdminFieldID{ID: fieldID, Valid: !fieldID.IsZero()},
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

type NewAdminContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentRelationCmd) Context() context.Context              { return c.ctx }
func (c NewAdminContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentRelationCmd) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentRelationCmd) TableName() string { return "admin_content_relations" }
func (c NewAdminContentRelationCmd) Params() any       { return c.params }
func (c NewAdminContentRelationCmd) GetID(r mdb.AdminContentRelations) string {
	return string(r.AdminContentRelationID)
}

func (c NewAdminContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminContentRelations, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminContentRelation(ctx, mdb.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        c.params.SourceContentID,
		TargetContentID:        c.params.TargetContentID,
		AdminFieldID:           types.NullableAdminFieldID{ID: c.params.AdminFieldID, Valid: !c.params.AdminFieldID.IsZero()},
		SortOrder:              c.params.SortOrder,
		DateCreated:            c.params.DateCreated,
	})
}

func (d Database) NewAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentRelationParams) NewAdminContentRelationCmd {
	return NewAdminContentRelationCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE (SortOrder) -----

type UpdateAdminContentRelationSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentRelationSortOrderCmd) Context() context.Context { return c.ctx }
func (c UpdateAdminContentRelationSortOrderCmd) AuditContext() audited.AuditContext {
	return c.auditCtx
}
func (c UpdateAdminContentRelationSortOrderCmd) Connection() *sql.DB { return c.conn }
func (c UpdateAdminContentRelationSortOrderCmd) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateAdminContentRelationSortOrderCmd) TableName() string {
	return "admin_content_relations"
}
func (c UpdateAdminContentRelationSortOrderCmd) Params() any { return c.params }
func (c UpdateAdminContentRelationSortOrderCmd) GetID() string {
	return string(c.params.AdminContentRelationID)
}

func (c UpdateAdminContentRelationSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentRelation(ctx, mdb.GetAdminContentRelationParams{AdminContentRelationID: c.params.AdminContentRelationID})
}

func (c UpdateAdminContentRelationSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminContentRelationSortOrder(ctx, mdb.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              c.params.SortOrder,
		AdminContentRelationID: c.params.AdminContentRelationID,
	})
}

func (d Database) UpdateAdminContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentRelationSortOrderParams) UpdateAdminContentRelationSortOrderCmd {
	return UpdateAdminContentRelationSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteAdminContentRelationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentRelationCmd) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentRelationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentRelationCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentRelationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentRelationCmd) TableName() string { return "admin_content_relations" }
func (c DeleteAdminContentRelationCmd) GetID() string     { return string(c.id) }

func (c DeleteAdminContentRelationCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentRelations, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentRelation(ctx, mdb.GetAdminContentRelationParams{AdminContentRelationID: c.id})
}

func (c DeleteAdminContentRelationCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminContentRelation(ctx, mdb.DeleteAdminContentRelationParams{AdminContentRelationID: c.id})
}

func (d Database) DeleteAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentRelationID) DeleteAdminContentRelationCmd {
	return DeleteAdminContentRelationCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewAdminContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentRelationCmdMysql) Context() context.Context              { return c.ctx }
func (c NewAdminContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentRelationCmdMysql) TableName() string { return "admin_content_relations" }
func (c NewAdminContentRelationCmdMysql) Params() any       { return c.params }
func (c NewAdminContentRelationCmdMysql) GetID(r mdbm.AdminContentRelations) string {
	return string(r.AdminContentRelationID)
}

func (c NewAdminContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentRelations, error) {
	id := types.NewAdminContentRelationID()
	queries := mdbm.New(tx)
	params := mdbm.CreateAdminContentRelationParams{
		AdminContentRelationID: id,
		SourceContentID:        c.params.SourceContentID,
		TargetContentID:        c.params.TargetContentID,
		AdminFieldID:           types.NullableAdminFieldID{ID: c.params.AdminFieldID, Valid: !c.params.AdminFieldID.IsZero()},
		SortOrder:              int32(c.params.SortOrder),
		DateCreated:            c.params.DateCreated,
	}
	if err := queries.CreateAdminContentRelation(ctx, params); err != nil {
		return mdbm.AdminContentRelations{}, err
	}
	return queries.GetAdminContentRelation(ctx, mdbm.GetAdminContentRelationParams{AdminContentRelationID: id})
}

func (d MysqlDatabase) NewAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentRelationParams) NewAdminContentRelationCmdMysql {
	return NewAdminContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE (SortOrder) -----

type UpdateAdminContentRelationSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentRelationSortOrderCmdMysql) Context() context.Context { return c.ctx }
func (c UpdateAdminContentRelationSortOrderCmdMysql) AuditContext() audited.AuditContext {
	return c.auditCtx
}
func (c UpdateAdminContentRelationSortOrderCmdMysql) Connection() *sql.DB { return c.conn }
func (c UpdateAdminContentRelationSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateAdminContentRelationSortOrderCmdMysql) TableName() string {
	return "admin_content_relations"
}
func (c UpdateAdminContentRelationSortOrderCmdMysql) Params() any { return c.params }
func (c UpdateAdminContentRelationSortOrderCmdMysql) GetID() string {
	return string(c.params.AdminContentRelationID)
}

func (c UpdateAdminContentRelationSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbm.GetAdminContentRelationParams{AdminContentRelationID: c.params.AdminContentRelationID})
}

func (c UpdateAdminContentRelationSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminContentRelationSortOrder(ctx, mdbm.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(c.params.SortOrder),
		AdminContentRelationID: c.params.AdminContentRelationID,
	})
}

func (d MysqlDatabase) UpdateAdminContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentRelationSortOrderParams) UpdateAdminContentRelationSortOrderCmdMysql {
	return UpdateAdminContentRelationSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteAdminContentRelationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentRelationCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentRelationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentRelationCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentRelationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentRelationCmdMysql) TableName() string { return "admin_content_relations" }
func (c DeleteAdminContentRelationCmdMysql) GetID() string     { return string(c.id) }

func (c DeleteAdminContentRelationCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentRelations, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbm.GetAdminContentRelationParams{AdminContentRelationID: c.id})
}

func (c DeleteAdminContentRelationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminContentRelation(ctx, mdbm.DeleteAdminContentRelationParams{AdminContentRelationID: c.id})
}

func (d MysqlDatabase) DeleteAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentRelationID) DeleteAdminContentRelationCmdMysql {
	return DeleteAdminContentRelationCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewAdminContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentRelationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewAdminContentRelationCmdPsql) Context() context.Context              { return c.ctx }
func (c NewAdminContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewAdminContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewAdminContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewAdminContentRelationCmdPsql) TableName() string { return "admin_content_relations" }
func (c NewAdminContentRelationCmdPsql) Params() any       { return c.params }
func (c NewAdminContentRelationCmdPsql) GetID(r mdbp.AdminContentRelations) string {
	return string(r.AdminContentRelationID)
}

func (c NewAdminContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminContentRelation(ctx, mdbp.CreateAdminContentRelationParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
		SourceContentID:        c.params.SourceContentID,
		TargetContentID:        c.params.TargetContentID,
		AdminFieldID:           types.NullableAdminFieldID{ID: c.params.AdminFieldID, Valid: !c.params.AdminFieldID.IsZero()},
		SortOrder:              int32(c.params.SortOrder),
		DateCreated:            c.params.DateCreated,
	})
}

func (d PsqlDatabase) NewAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentRelationParams) NewAdminContentRelationCmdPsql {
	return NewAdminContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE (SortOrder) -----

type UpdateAdminContentRelationSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentRelationSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminContentRelationSortOrderCmdPsql) Context() context.Context { return c.ctx }
func (c UpdateAdminContentRelationSortOrderCmdPsql) AuditContext() audited.AuditContext {
	return c.auditCtx
}
func (c UpdateAdminContentRelationSortOrderCmdPsql) Connection() *sql.DB { return c.conn }
func (c UpdateAdminContentRelationSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateAdminContentRelationSortOrderCmdPsql) TableName() string {
	return "admin_content_relations"
}
func (c UpdateAdminContentRelationSortOrderCmdPsql) Params() any { return c.params }
func (c UpdateAdminContentRelationSortOrderCmdPsql) GetID() string {
	return string(c.params.AdminContentRelationID)
}

func (c UpdateAdminContentRelationSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbp.GetAdminContentRelationParams{AdminContentRelationID: c.params.AdminContentRelationID})
}

func (c UpdateAdminContentRelationSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminContentRelationSortOrder(ctx, mdbp.UpdateAdminContentRelationSortOrderParams{
		SortOrder:              int32(c.params.SortOrder),
		AdminContentRelationID: c.params.AdminContentRelationID,
	})
}

func (d PsqlDatabase) UpdateAdminContentRelationSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentRelationSortOrderParams) UpdateAdminContentRelationSortOrderCmdPsql {
	return UpdateAdminContentRelationSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteAdminContentRelationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentRelationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteAdminContentRelationCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteAdminContentRelationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteAdminContentRelationCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteAdminContentRelationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteAdminContentRelationCmdPsql) TableName() string { return "admin_content_relations" }
func (c DeleteAdminContentRelationCmdPsql) GetID() string     { return string(c.id) }

func (c DeleteAdminContentRelationCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentRelations, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentRelation(ctx, mdbp.GetAdminContentRelationParams{AdminContentRelationID: c.id})
}

func (c DeleteAdminContentRelationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminContentRelation(ctx, mdbp.DeleteAdminContentRelationParams{AdminContentRelationID: c.id})
}

func (d PsqlDatabase) DeleteAdminContentRelationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentRelationID) DeleteAdminContentRelationCmdPsql {
	return DeleteAdminContentRelationCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
