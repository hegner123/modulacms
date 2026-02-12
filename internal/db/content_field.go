package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type ContentFields struct {
	ContentFieldID types.ContentFieldID    `json:"content_field_id"`
	RouteID        types.NullableRouteID   `json:"route_id"`
	ContentDataID  types.NullableContentID `json:"content_data_id"`
	FieldID        types.NullableFieldID   `json:"field_id"`
	FieldValue     string                  `json:"field_value"`
	AuthorID       types.NullableUserID    `json:"author_id"`
	DateCreated    types.Timestamp         `json:"date_created"`
	DateModified   types.Timestamp         `json:"date_modified"`
}

type CreateContentFieldParams struct {
	RouteID       types.NullableRouteID   `json:"route_id"`
	ContentDataID types.NullableContentID `json:"content_data_id"`
	FieldID       types.NullableFieldID   `json:"field_id"`
	FieldValue    string                  `json:"field_value"`
	AuthorID      types.NullableUserID    `json:"author_id"`
	DateCreated   types.Timestamp         `json:"date_created"`
	DateModified  types.Timestamp         `json:"date_modified"`
}

type UpdateContentFieldParams struct {
	RouteID        types.NullableRouteID   `json:"route_id"`
	ContentDataID  types.NullableContentID `json:"content_data_id"`
	FieldID        types.NullableFieldID   `json:"field_id"`
	FieldValue     string                  `json:"field_value"`
	AuthorID       types.NullableUserID    `json:"author_id"`
	DateCreated    types.Timestamp         `json:"date_created"`
	DateModified   types.Timestamp         `json:"date_modified"`
	ContentFieldID types.ContentFieldID    `json:"content_field_id"`
}

type ListContentFieldsByRoutePaginatedParams struct {
	RouteID types.NullableRouteID
	Limit   int64
	Offset  int64
}

type ListContentFieldsByContentDataPaginatedParams struct {
	ContentDataID types.NullableContentID
	Limit         int64
	Offset        int64
}

// FormParams and JSON variants removed - use typed params directly

// ContentFieldsJSON is used for JSON serialization in model package
// Deprecated: Will be removed in future version. Use typed ContentFields directly.
type ContentFieldsJSON struct {
	ContentFieldID int64  `json:"content_field_id"`
	RouteID        int64  `json:"route_id"`
	ContentDataID  int64  `json:"content_data_id"`
	FieldID        int64  `json:"field_id"`
	FieldValue     string `json:"field_value"`
	AuthorID       int64  `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
}

// MapContentFieldJSON converts ContentFields to ContentFieldsJSON for JSON serialization
// Deprecated: Will be removed in future version
func MapContentFieldJSON(a ContentFields) ContentFieldsJSON {
	return ContentFieldsJSON{
		ContentFieldID: 0,                       // Type conversion not available, set to 0
		RouteID:        0,                       // Type conversion not available, set to 0
		ContentDataID:  0,                       // Type conversion not available, set to 0
		FieldID:        0,                       // Type conversion not available, set to 0
		FieldValue:     a.FieldValue,
		AuthorID:       0,                       // Type conversion not available, set to 0
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

func MapStringContentField(a ContentFields) StringContentFields {
	return StringContentFields{
		ContentFieldID: a.ContentFieldID.String(),
		RouteID:        a.RouteID.String(),
		ContentDataID:  a.ContentDataID.String(),
		FieldID:        a.FieldID.String(),
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID.String(),
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
		History:        "", // History field removed from ContentFields
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapContentField(a mdb.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d Database) MapCreateContentFieldParams(a CreateContentFieldParams) mdb.CreateContentFieldParams {
	return mdb.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d Database) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdb.UpdateContentFieldParams {
	return mdb.UpdateContentFieldParams{
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
		ContentFieldID: a.ContentFieldID,
	}
}

// QUERIES

func (d Database) CountContentFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d Database) CreateContentField(ctx context.Context, ac audited.AuditContext, s CreateContentFieldParams) (*ContentFields, error) {
	cmd := d.NewContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentField: %w", err)
	}
	r := d.MapContentField(result)
	return &r, nil
}

func (d Database) DeleteContentField(ctx context.Context, ac audited.AuditContext, id types.ContentFieldID) error {
	cmd := d.DeleteContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetContentField(id types.ContentFieldID) (*ContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentField(d.Context, mdb.GetContentFieldParams{ContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d Database) ListContentFields() (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsByRoute(routeID types.NullableRouteID) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, mdb.ListContentFieldsByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsByContentData(contentDataID types.NullableContentID) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentData(d.Context, mdb.ListContentFieldsByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsPaginated(params PaginationParams) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsPaginated(d.Context, mdb.ListContentFieldsPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsByRoutePaginated(params ListContentFieldsByRoutePaginatedParams) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoutePaginated(d.Context, mdb.ListContentFieldsByRoutePaginatedParams{
		RouteID: params.RouteID,
		Limit:   params.Limit,
		Offset:  params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsByContentDataPaginated(params ListContentFieldsByContentDataPaginatedParams) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentDataPaginated(d.Context, mdb.ListContentFieldsByContentDataPaginatedParams{
		ContentDataID: params.ContentDataID,
		Limit:         params.Limit,
		Offset:        params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateContentField(ctx context.Context, ac audited.AuditContext, s UpdateContentFieldParams) (*string, error) {
	cmd := d.UpdateContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update contentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ContentFieldID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapContentField(a mdbm.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbm.CreateContentFieldParams {
	return mdbm.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
	}
}

func (d MysqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbm.UpdateContentFieldParams {
	return mdbm.UpdateContentFieldParams{
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		ContentFieldID: a.ContentFieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountContentFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateContentFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateContentField(ctx context.Context, ac audited.AuditContext, s CreateContentFieldParams) (*ContentFields, error) {
	cmd := d.NewContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentField: %w", err)
	}
	r := d.MapContentField(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteContentField(ctx context.Context, ac audited.AuditContext, id types.ContentFieldID) error {
	cmd := d.DeleteContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetContentField(id types.ContentFieldID) (*ContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentField(d.Context, mdbm.GetContentFieldParams{ContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d MysqlDatabase) ListContentFields() (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsByRoute(routeID types.NullableRouteID) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, mdbm.ListContentFieldsByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsByContentData(contentDataID types.NullableContentID) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentData(d.Context, mdbm.ListContentFieldsByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsPaginated(params PaginationParams) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsPaginated(d.Context, mdbm.ListContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsByRoutePaginated(params ListContentFieldsByRoutePaginatedParams) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoutePaginated(d.Context, mdbm.ListContentFieldsByRoutePaginatedParams{
		RouteID: params.RouteID,
		Limit:   int32(params.Limit),
		Offset:  int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsByContentDataPaginated(params ListContentFieldsByContentDataPaginatedParams) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentDataPaginated(d.Context, mdbm.ListContentFieldsByContentDataPaginatedParams{
		ContentDataID: params.ContentDataID,
		Limit:         int32(params.Limit),
		Offset:        int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateContentField(ctx context.Context, ac audited.AuditContext, s UpdateContentFieldParams) (*string, error) {
	cmd := d.UpdateContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update contentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ContentFieldID)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapContentField(a mdbp.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbp.CreateContentFieldParams {
	return mdbp.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbp.UpdateContentFieldParams {
	return mdbp.UpdateContentFieldParams{
		ContentFieldID:   a.ContentFieldID,
		RouteID:          a.RouteID,
		ContentDataID:    a.ContentDataID,
		FieldID:          a.FieldID,
		FieldValue:       a.FieldValue,
		AuthorID:         a.AuthorID,
		DateCreated:      a.DateCreated,
		DateModified:     a.DateModified,
		ContentFieldID_2: a.ContentFieldID,
	}
}

// QUERIES

func (d PsqlDatabase) CountContentFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateContentFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateContentField(ctx context.Context, ac audited.AuditContext, s CreateContentFieldParams) (*ContentFields, error) {
	cmd := d.NewContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentField: %w", err)
	}
	r := d.MapContentField(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteContentField(ctx context.Context, ac audited.AuditContext, id types.ContentFieldID) error {
	cmd := d.DeleteContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetContentField(id types.ContentFieldID) (*ContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentField(d.Context, mdbp.GetContentFieldParams{ContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d PsqlDatabase) ListContentFields() (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsByRoute(routeID types.NullableRouteID) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, mdbp.ListContentFieldsByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsByContentData(contentDataID types.NullableContentID) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentData(d.Context, mdbp.ListContentFieldsByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsPaginated(params PaginationParams) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsPaginated(d.Context, mdbp.ListContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsByRoutePaginated(params ListContentFieldsByRoutePaginatedParams) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoutePaginated(d.Context, mdbp.ListContentFieldsByRoutePaginatedParams{
		RouteID: params.RouteID,
		Limit:   int32(params.Limit),
		Offset:  int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsByContentDataPaginated(params ListContentFieldsByContentDataPaginatedParams) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentDataPaginated(d.Context, mdbp.ListContentFieldsByContentDataPaginatedParams{
		ContentDataID: params.ContentDataID,
		Limit:         int32(params.Limit),
		Offset:        int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data paginated: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateContentField(ctx context.Context, ac audited.AuditContext, s UpdateContentFieldParams) (*string, error) {
	cmd := d.UpdateContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update contentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ContentFieldID)
	return &msg, nil
}

// Utility function for backward compatibility
// Deprecated: Use types.NullableRouteID directly
func NullableRouteIDFromInt64(id int64) types.NullableRouteID {
	return types.NullableRouteID{ID: types.RouteID(strconv.FormatInt(id, 10)), Valid: id != 0}
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewContentFieldCmd is an audited create command for content_fields (SQLite).
type NewContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewContentFieldCmd) Context() context.Context              { return c.ctx }
func (c NewContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c NewContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentFieldCmd) TableName() string                     { return "content_fields" }
func (c NewContentFieldCmd) Params() any                           { return c.params }
func (c NewContentFieldCmd) GetID(row mdb.ContentFields) string    { return string(row.ContentFieldID) }

func (c NewContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.ContentFields, error) {
	queries := mdb.New(tx)
	return queries.CreateContentField(ctx, mdb.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        c.params.RouteID,
		ContentDataID:  c.params.ContentDataID,
		FieldID:        c.params.FieldID,
		FieldValue:     c.params.FieldValue,
		AuthorID:       c.params.AuthorID,
		DateCreated:    c.params.DateCreated,
		DateModified:   c.params.DateModified,
	})
}

func (d Database) NewContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentFieldParams) NewContentFieldCmd {
	return NewContentFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateContentFieldCmd is an audited update command for content_fields (SQLite).
type UpdateContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateContentFieldCmd) Context() context.Context              { return c.ctx }
func (c UpdateContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateContentFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateContentFieldCmd) TableName() string                     { return "content_fields" }
func (c UpdateContentFieldCmd) Params() any                           { return c.params }
func (c UpdateContentFieldCmd) GetID() string                         { return string(c.params.ContentFieldID) }

func (c UpdateContentFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentFields, error) {
	queries := mdb.New(tx)
	return queries.GetContentField(ctx, mdb.GetContentFieldParams{ContentFieldID: c.params.ContentFieldID})
}

func (c UpdateContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateContentField(ctx, mdb.UpdateContentFieldParams{
		RouteID:        c.params.RouteID,
		ContentDataID:  c.params.ContentDataID,
		FieldID:        c.params.FieldID,
		FieldValue:     c.params.FieldValue,
		AuthorID:       c.params.AuthorID,
		DateCreated:    c.params.DateCreated,
		DateModified:   c.params.DateModified,
		ContentFieldID: c.params.ContentFieldID,
	})
}

func (d Database) UpdateContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentFieldParams) UpdateContentFieldCmd {
	return UpdateContentFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// DeleteContentFieldCmd is an audited delete command for content_fields (SQLite).
type DeleteContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteContentFieldCmd) Context() context.Context              { return c.ctx }
func (c DeleteContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentFieldCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentFieldCmd) TableName() string                     { return "content_fields" }
func (c DeleteContentFieldCmd) GetID() string                         { return string(c.id) }

func (c DeleteContentFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentFields, error) {
	queries := mdb.New(tx)
	return queries.GetContentField(ctx, mdb.GetContentFieldParams{ContentFieldID: c.id})
}

func (c DeleteContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteContentField(ctx, mdb.DeleteContentFieldParams{ContentFieldID: c.id})
}

func (d Database) DeleteContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentFieldID) DeleteContentFieldCmd {
	return DeleteContentFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

// NewContentFieldCmdMysql is an audited create command for content_fields (MySQL).
type NewContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewContentFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c NewContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentFieldCmdMysql) TableName() string                     { return "content_fields" }
func (c NewContentFieldCmdMysql) Params() any                           { return c.params }
func (c NewContentFieldCmdMysql) GetID(row mdbm.ContentFields) string   { return string(row.ContentFieldID) }

func (c NewContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.ContentFields, error) {
	id := types.NewContentFieldID()
	queries := mdbm.New(tx)
	err := queries.CreateContentField(ctx, mdbm.CreateContentFieldParams{
		ContentFieldID: id,
		RouteID:        c.params.RouteID,
		ContentDataID:  c.params.ContentDataID,
		FieldID:        c.params.FieldID,
		FieldValue:     c.params.FieldValue,
		AuthorID:       c.params.AuthorID,
	})
	if err != nil {
		return mdbm.ContentFields{}, fmt.Errorf("execute create content_fields: %w", err)
	}
	return queries.GetContentField(ctx, mdbm.GetContentFieldParams{ContentFieldID: id})
}

func (d MysqlDatabase) NewContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentFieldParams) NewContentFieldCmdMysql {
	return NewContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateContentFieldCmdMysql is an audited update command for content_fields (MySQL).
type UpdateContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateContentFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateContentFieldCmdMysql) TableName() string                     { return "content_fields" }
func (c UpdateContentFieldCmdMysql) Params() any                           { return c.params }
func (c UpdateContentFieldCmdMysql) GetID() string                         { return string(c.params.ContentFieldID) }

func (c UpdateContentFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentFields, error) {
	queries := mdbm.New(tx)
	return queries.GetContentField(ctx, mdbm.GetContentFieldParams{ContentFieldID: c.params.ContentFieldID})
}

func (c UpdateContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateContentField(ctx, mdbm.UpdateContentFieldParams{
		RouteID:        c.params.RouteID,
		ContentDataID:  c.params.ContentDataID,
		FieldID:        c.params.FieldID,
		FieldValue:     c.params.FieldValue,
		AuthorID:       c.params.AuthorID,
		ContentFieldID: c.params.ContentFieldID,
	})
}

func (d MysqlDatabase) UpdateContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentFieldParams) UpdateContentFieldCmdMysql {
	return UpdateContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// DeleteContentFieldCmdMysql is an audited delete command for content_fields (MySQL).
type DeleteContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteContentFieldCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentFieldCmdMysql) TableName() string                     { return "content_fields" }
func (c DeleteContentFieldCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteContentFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentFields, error) {
	queries := mdbm.New(tx)
	return queries.GetContentField(ctx, mdbm.GetContentFieldParams{ContentFieldID: c.id})
}

func (c DeleteContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteContentField(ctx, mdbm.DeleteContentFieldParams{ContentFieldID: c.id})
}

func (d MysqlDatabase) DeleteContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentFieldID) DeleteContentFieldCmdMysql {
	return DeleteContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

// NewContentFieldCmdPsql is an audited create command for content_fields (PostgreSQL).
type NewContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewContentFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c NewContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentFieldCmdPsql) TableName() string                     { return "content_fields" }
func (c NewContentFieldCmdPsql) Params() any                           { return c.params }
func (c NewContentFieldCmdPsql) GetID(row mdbp.ContentFields) string   { return string(row.ContentFieldID) }

func (c NewContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.ContentFields, error) {
	queries := mdbp.New(tx)
	return queries.CreateContentField(ctx, mdbp.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        c.params.RouteID,
		ContentDataID:  c.params.ContentDataID,
		FieldID:        c.params.FieldID,
		FieldValue:     c.params.FieldValue,
		AuthorID:       c.params.AuthorID,
		DateCreated:    c.params.DateCreated,
		DateModified:   c.params.DateModified,
	})
}

func (d PsqlDatabase) NewContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentFieldParams) NewContentFieldCmdPsql {
	return NewContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateContentFieldCmdPsql is an audited update command for content_fields (PostgreSQL).
type UpdateContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateContentFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateContentFieldCmdPsql) TableName() string                     { return "content_fields" }
func (c UpdateContentFieldCmdPsql) Params() any                           { return c.params }
func (c UpdateContentFieldCmdPsql) GetID() string                         { return string(c.params.ContentFieldID) }

func (c UpdateContentFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentFields, error) {
	queries := mdbp.New(tx)
	return queries.GetContentField(ctx, mdbp.GetContentFieldParams{ContentFieldID: c.params.ContentFieldID})
}

func (c UpdateContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateContentField(ctx, mdbp.UpdateContentFieldParams{
		ContentFieldID:   c.params.ContentFieldID,
		RouteID:          c.params.RouteID,
		ContentDataID:    c.params.ContentDataID,
		FieldID:          c.params.FieldID,
		FieldValue:       c.params.FieldValue,
		AuthorID:         c.params.AuthorID,
		DateCreated:      c.params.DateCreated,
		DateModified:     c.params.DateModified,
		ContentFieldID_2: c.params.ContentFieldID,
	})
}

func (d PsqlDatabase) UpdateContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentFieldParams) UpdateContentFieldCmdPsql {
	return UpdateContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// DeleteContentFieldCmdPsql is an audited delete command for content_fields (PostgreSQL).
type DeleteContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteContentFieldCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentFieldCmdPsql) TableName() string                     { return "content_fields" }
func (c DeleteContentFieldCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteContentFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentFields, error) {
	queries := mdbp.New(tx)
	return queries.GetContentField(ctx, mdbp.GetContentFieldParams{ContentFieldID: c.id})
}

func (c DeleteContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteContentField(ctx, mdbp.DeleteContentFieldParams{ContentFieldID: c.id})
}

func (d PsqlDatabase) DeleteContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentFieldID) DeleteContentFieldCmdPsql {
	return DeleteContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
