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

// ContentData represents a content entry in the tree-based content structure.
type ContentData struct {
	ContentDataID types.ContentID          `json:"content_data_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  types.NullableContentID  `json:"first_child_id"`
	NextSiblingID types.NullableContentID  `json:"next_sibling_id"`
	PrevSiblingID types.NullableContentID  `json:"prev_sibling_id"`
	RouteID       types.NullableRouteID    `json:"route_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	AuthorID      types.NullableUserID     `json:"author_id"`
	Status        types.ContentStatus      `json:"status"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
}

// CreateContentDataParams holds parameters for creating a new content_data record.
type CreateContentDataParams struct {
	RouteID       types.NullableRouteID    `json:"route_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  types.NullableContentID  `json:"first_child_id"`
	NextSiblingID types.NullableContentID  `json:"next_sibling_id"`
	PrevSiblingID types.NullableContentID  `json:"prev_sibling_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	AuthorID      types.NullableUserID     `json:"author_id"`
	Status        types.ContentStatus      `json:"status"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
}

// UpdateContentDataParams holds parameters for updating an existing content_data record.
type UpdateContentDataParams struct {
	RouteID       types.NullableRouteID    `json:"route_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  types.NullableContentID  `json:"first_child_id"`
	NextSiblingID types.NullableContentID  `json:"next_sibling_id"`
	PrevSiblingID types.NullableContentID  `json:"prev_sibling_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	AuthorID      types.NullableUserID     `json:"author_id"`
	Status        types.ContentStatus      `json:"status"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
	ContentDataID types.ContentID          `json:"content_data_id"`
}

// ListContentDataByRoutePaginatedParams holds parameters for paginated listing of content_data by route.
type ListContentDataByRoutePaginatedParams struct {
	RouteID types.NullableRouteID
	Limit   int64
	Offset  int64
}

// ContentDataJSON provides a string-based representation for JSON serialization
type ContentDataJSON struct {
	ContentDataID string `json:"content_data_id"`
	ParentID      string `json:"parent_id"`
	FirstChildID  string `json:"first_child_id"`
	NextSiblingID string `json:"next_sibling_id"`
	PrevSiblingID string `json:"prev_sibling_id"`
	RouteID       string `json:"route_id"`
	DatatypeID    string `json:"datatype_id"`
	AuthorID      string `json:"author_id"`
	Status        string `json:"status"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
}

// nullableContentIDStringEmpty returns "" when the nullable ID is invalid,
// and the ID string when valid. Used for sibling pointer fields that the
// tree builder checks with == "".
func nullableContentIDStringEmpty(n types.NullableContentID) string {
	if !n.Valid {
		return ""
	}
	return n.ID.String()
}

// MapContentDataJSON converts ContentData to ContentDataJSON for JSON serialization
func MapContentDataJSON(a ContentData) ContentDataJSON {
	return ContentDataJSON{
		ContentDataID: a.ContentDataID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  nullableContentIDStringEmpty(a.FirstChildID),
		NextSiblingID: nullableContentIDStringEmpty(a.NextSiblingID),
		PrevSiblingID: nullableContentIDStringEmpty(a.PrevSiblingID),
		RouteID:       a.RouteID.String(),
		DatatypeID:    a.DatatypeID.String(),
		AuthorID:      a.AuthorID.String(),
		Status:        string(a.Status),
		DateCreated:   a.DateCreated.String(),
		DateModified:  a.DateModified.String(),
	}
}

// MapStringContentData converts ContentData to StringContentData for table display
func MapStringContentData(a ContentData) StringContentData {
	return StringContentData{
		ContentDataID: a.ContentDataID.String(),
		RouteID:       a.RouteID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  nullableContentIDStringEmpty(a.FirstChildID),
		NextSiblingID: nullableContentIDStringEmpty(a.NextSiblingID),
		PrevSiblingID: nullableContentIDStringEmpty(a.PrevSiblingID),
		DatatypeID:    a.DatatypeID.String(),
		AuthorID:      a.AuthorID.String(),
		Status:        string(a.Status),
		DateCreated:   a.DateCreated.String(),
		DateModified:  a.DateModified.String(),
		History:       "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapContentData converts a sqlc-generated type to the wrapper type.
func (d Database) MapContentData(a mdb.ContentData) ContentData {
	return ContentData{
		ContentDataID: a.ContentDataID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// MapCreateContentDataParams converts wrapper params to sqlc-generated create params.
func (d Database) MapCreateContentDataParams(a CreateContentDataParams) mdb.CreateContentDataParams {
	return mdb.CreateContentDataParams{
		ContentDataID: types.NewContentID(),
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// MapUpdateContentDataParams converts wrapper params to sqlc-generated update params.
func (d Database) MapUpdateContentDataParams(a UpdateContentDataParams) mdb.UpdateContentDataParams {
	return mdb.UpdateContentDataParams{
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		ContentDataID: a.ContentDataID,
	}
}

// QUERIES

// CountContentData returns the total count of content_data records.
func (d Database) CountContentData() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateContentDataTable creates the content_data table.
func (d Database) CreateContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

// CreateContentData inserts a new content_data record with audit tracking.
func (d Database) CreateContentData(ctx context.Context, ac audited.AuditContext, s CreateContentDataParams) (*ContentData, error) {
	cmd := d.NewContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentData: %w", err)
	}
	r := d.MapContentData(result)
	return &r, nil
}

// DeleteContentData removes a content_data record with audit tracking.
func (d Database) DeleteContentData(ctx context.Context, ac audited.AuditContext, id types.ContentID) error {
	cmd := d.DeleteContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetContentData retrieves a content_data record by ID.
func (d Database) GetContentData(id types.ContentID) (*ContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentData(d.Context, mdb.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

// ListContentData returns all content_data records.
func (d Database) ListContentData() (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataByRoute returns all content_data records for a given route.
func (d Database) ListContentDataByRoute(routeID types.NullableRouteID) (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataByRoute(d.Context, mdb.ListContentDataByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataPaginated returns a paginated subset of content_data records.
func (d Database) ListContentDataPaginated(params PaginationParams) (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataPaginated(d.Context, mdb.ListContentDataPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData paginated: %v", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataByRoutePaginated returns a paginated subset of content_data records for a given route.
func (d Database) ListContentDataByRoutePaginated(params ListContentDataByRoutePaginatedParams) (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataByRoutePaginated(d.Context, mdb.ListContentDataByRoutePaginatedParams{
		RouteID: params.RouteID,
		Limit:   params.Limit,
		Offset:  params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route paginated: %v", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateContentData modifies an existing content_data record with audit tracking.
func (d Database) UpdateContentData(ctx context.Context, ac audited.AuditContext, s UpdateContentDataParams) (*string, error) {
	cmd := d.UpdateContentDataCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update contentData: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ContentDataID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapContentData converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapContentData(a mdbm.ContentData) ContentData {
	return ContentData{
		ContentDataID: a.ContentDataID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// MapCreateContentDataParams converts wrapper params to sqlc-generated create params.
func (d MysqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbm.CreateContentDataParams {
	return mdbm.CreateContentDataParams{
		ContentDataID: types.NewContentID(),
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// MapUpdateContentDataParams converts wrapper params to sqlc-generated update params.
func (d MysqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbm.UpdateContentDataParams {
	return mdbm.UpdateContentDataParams{
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		ContentDataID: a.ContentDataID,
	}
}

// QUERIES

// CountContentData returns the total count of content_data records.
func (d MysqlDatabase) CountContentData() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateContentDataTable creates the content_data table.
func (d MysqlDatabase) CreateContentDataTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

// CreateContentData inserts a new content_data record with audit tracking.
func (d MysqlDatabase) CreateContentData(ctx context.Context, ac audited.AuditContext, s CreateContentDataParams) (*ContentData, error) {
	cmd := d.NewContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentData: %w", err)
	}
	r := d.MapContentData(result)
	return &r, nil
}

// DeleteContentData removes a content_data record with audit tracking.
func (d MysqlDatabase) DeleteContentData(ctx context.Context, ac audited.AuditContext, id types.ContentID) error {
	cmd := d.DeleteContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetContentData retrieves a content_data record by ID.
func (d MysqlDatabase) GetContentData(id types.ContentID) (*ContentData, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentData(d.Context, mdbm.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

// ListContentData returns all content_data records.
func (d MysqlDatabase) ListContentData() (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataByRoute returns all content_data records for a given route.
func (d MysqlDatabase) ListContentDataByRoute(routeID types.NullableRouteID) (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataByRoute(d.Context, mdbm.ListContentDataByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataPaginated returns a paginated subset of content_data records.
func (d MysqlDatabase) ListContentDataPaginated(params PaginationParams) (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataPaginated(d.Context, mdbm.ListContentDataPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData paginated: %v", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataByRoutePaginated returns a paginated subset of content_data records for a given route.
func (d MysqlDatabase) ListContentDataByRoutePaginated(params ListContentDataByRoutePaginatedParams) (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataByRoutePaginated(d.Context, mdbm.ListContentDataByRoutePaginatedParams{
		RouteID: params.RouteID,
		Limit:   int32(params.Limit),
		Offset:  int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route paginated: %v", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateContentData modifies an existing content_data record with audit tracking.
func (d MysqlDatabase) UpdateContentData(ctx context.Context, ac audited.AuditContext, s UpdateContentDataParams) (*string, error) {
	cmd := d.UpdateContentDataCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update contentData: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ContentDataID)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapContentData converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapContentData(a mdbp.ContentData) ContentData {
	return ContentData{
		ContentDataID: a.ContentDataID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// MapCreateContentDataParams converts wrapper params to sqlc-generated create params.
func (d PsqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbp.CreateContentDataParams {
	return mdbp.CreateContentDataParams{
		ContentDataID: types.NewContentID(),
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// MapUpdateContentDataParams converts wrapper params to sqlc-generated update params.
func (d PsqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbp.UpdateContentDataParams {
	return mdbp.UpdateContentDataParams{
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		Status:        a.Status,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		ContentDataID: a.ContentDataID,
	}
}

// QUERIES

// CountContentData returns the total count of content_data records.
func (d PsqlDatabase) CountContentData() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateContentDataTable creates the content_data table.
func (d PsqlDatabase) CreateContentDataTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

// CreateContentData inserts a new content_data record with audit tracking.
func (d PsqlDatabase) CreateContentData(ctx context.Context, ac audited.AuditContext, s CreateContentDataParams) (*ContentData, error) {
	cmd := d.NewContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentData: %w", err)
	}
	r := d.MapContentData(result)
	return &r, nil
}

// DeleteContentData removes a content_data record with audit tracking.
func (d PsqlDatabase) DeleteContentData(ctx context.Context, ac audited.AuditContext, id types.ContentID) error {
	cmd := d.DeleteContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetContentData retrieves a content_data record by ID.
func (d PsqlDatabase) GetContentData(id types.ContentID) (*ContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentData(d.Context, mdbp.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

// ListContentData returns all content_data records.
func (d PsqlDatabase) ListContentData() (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataByRoute returns all content_data records for a given route.
func (d PsqlDatabase) ListContentDataByRoute(routeID types.NullableRouteID) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataByRoute(d.Context, mdbp.ListContentDataByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataPaginated returns a paginated subset of content_data records.
func (d PsqlDatabase) ListContentDataPaginated(params PaginationParams) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataPaginated(d.Context, mdbp.ListContentDataPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData paginated: %v", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentDataByRoutePaginated returns a paginated subset of content_data records for a given route.
func (d PsqlDatabase) ListContentDataByRoutePaginated(params ListContentDataByRoutePaginatedParams) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataByRoutePaginated(d.Context, mdbp.ListContentDataByRoutePaginatedParams{
		RouteID: params.RouteID,
		Limit:   int32(params.Limit),
		Offset:  int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route paginated: %v", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateContentData modifies an existing content_data record with audit tracking.
func (d PsqlDatabase) UpdateContentData(ctx context.Context, ac audited.AuditContext, s UpdateContentDataParams) (*string, error) {
	cmd := d.UpdateContentDataCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update contentData: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ContentDataID)
	return &msg, nil
}

///////////////////////////////
// ROOT CONTENT SUMMARY
//////////////////////////////

// RootContentSummary represents a root content data entry with route and datatype info
type RootContentSummary struct {
	ContentDataID types.ContentID          `json:"content_data_id"`
	RouteID       types.NullableRouteID    `json:"route_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	RouteSlug     types.Slug               `json:"route_slug"`
	RouteTitle    string                   `json:"route_title"`
	DatatypeLabel string                   `json:"datatype_label"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
}

// SQLITE - ListRootContentSummary

// MapRootContentSummary converts a sqlc-generated type to the wrapper type.
func (d Database) MapRootContentSummary(a mdb.ListRootContentSummaryRow) RootContentSummary {
	return RootContentSummary{
		ContentDataID: a.ContentDataID,
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// ListRootContentSummary returns all root content entries with route and datatype information.
func (d Database) ListRootContentSummary() (*[]RootContentSummary, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRootContentSummary(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get root content summary: %v", err)
	}
	res := []RootContentSummary{}
	for _, v := range rows {
		m := d.MapRootContentSummary(v)
		res = append(res, m)
	}
	return &res, nil
}

// MYSQL - ListRootContentSummary

// MapRootContentSummary converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapRootContentSummary(a mdbm.ListRootContentSummaryRow) RootContentSummary {
	return RootContentSummary{
		ContentDataID: a.ContentDataID,
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// ListRootContentSummary returns all root content entries with route and datatype information.
func (d MysqlDatabase) ListRootContentSummary() (*[]RootContentSummary, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRootContentSummary(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get root content summary: %v", err)
	}
	res := []RootContentSummary{}
	for _, v := range rows {
		m := d.MapRootContentSummary(v)
		res = append(res, m)
	}
	return &res, nil
}

// PSQL - ListRootContentSummary

// MapRootContentSummary converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapRootContentSummary(a mdbp.ListRootContentSummaryRow) RootContentSummary {
	return RootContentSummary{
		ContentDataID: a.ContentDataID,
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

// ListRootContentSummary returns all root content entries with route and datatype information.
func (d PsqlDatabase) ListRootContentSummary() (*[]RootContentSummary, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRootContentSummary(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get root content summary: %v", err)
	}
	res := []RootContentSummary{}
	for _, v := range rows {
		m := d.MapRootContentSummary(v)
		res = append(res, m)
	}
	return &res, nil
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewContentDataCmd is an audited create command for content_data (SQLite).
type NewContentDataCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewContentDataCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentDataCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c NewContentDataCmd) TableName() string                     { return "content_data" }

// Params returns the command parameters.
func (c NewContentDataCmd) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewContentDataCmd) GetID(row mdb.ContentData) string      { return string(row.ContentDataID) }

// Execute creates the record in the database.
func (c NewContentDataCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.ContentData, error) {
	queries := mdb.New(tx)
	return queries.CreateContentData(ctx, mdb.CreateContentDataParams{
		ContentDataID: types.NewContentID(),
		RouteID:       c.params.RouteID,
		ParentID:      c.params.ParentID,
		FirstChildID:  c.params.FirstChildID,
		NextSiblingID: c.params.NextSiblingID,
		PrevSiblingID: c.params.PrevSiblingID,
		DatatypeID:    c.params.DatatypeID,
		AuthorID:      c.params.AuthorID,
		Status:        c.params.Status,
		DateCreated:   c.params.DateCreated,
		DateModified:  c.params.DateModified,
	})
}

// NewContentDataCmd creates a new audited create command.
func (d Database) NewContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentDataParams) NewContentDataCmd {
	return NewContentDataCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateContentDataCmd is an audited update command for content_data (SQLite).
type UpdateContentDataCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateContentDataCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateContentDataCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c UpdateContentDataCmd) TableName() string                     { return "content_data" }

// Params returns the command parameters.
func (c UpdateContentDataCmd) Params() any                           { return c.params }

// GetID returns the record ID being updated.
func (c UpdateContentDataCmd) GetID() string                         { return string(c.params.ContentDataID) }

// GetBefore retrieves the record before modification.
func (c UpdateContentDataCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentData, error) {
	queries := mdb.New(tx)
	return queries.GetContentData(ctx, mdb.GetContentDataParams{ContentDataID: c.params.ContentDataID})
}

// Execute updates the record in the database.
func (c UpdateContentDataCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateContentData(ctx, mdb.UpdateContentDataParams{
		RouteID:       c.params.RouteID,
		ParentID:      c.params.ParentID,
		FirstChildID:  c.params.FirstChildID,
		NextSiblingID: c.params.NextSiblingID,
		PrevSiblingID: c.params.PrevSiblingID,
		DatatypeID:    c.params.DatatypeID,
		AuthorID:      c.params.AuthorID,
		Status:        c.params.Status,
		DateCreated:   c.params.DateCreated,
		DateModified:  c.params.DateModified,
		ContentDataID: c.params.ContentDataID,
	})
}

// UpdateContentDataCmd creates a new audited update command.
func (d Database) UpdateContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentDataParams) UpdateContentDataCmd {
	return UpdateContentDataCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// DeleteContentDataCmd is an audited delete command for content_data (SQLite).
type DeleteContentDataCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteContentDataCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentDataCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c DeleteContentDataCmd) TableName() string                     { return "content_data" }

// GetID returns the record ID being deleted.
func (c DeleteContentDataCmd) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteContentDataCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentData, error) {
	queries := mdb.New(tx)
	return queries.GetContentData(ctx, mdb.GetContentDataParams{ContentDataID: c.id})
}

// Execute deletes the record from the database.
func (c DeleteContentDataCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteContentData(ctx, mdb.DeleteContentDataParams{ContentDataID: c.id})
}

// DeleteContentDataCmd creates a new audited delete command.
func (d Database) DeleteContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentID) DeleteContentDataCmd {
	return DeleteContentDataCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

// NewContentDataCmdMysql is an audited create command for content_data (MySQL).
type NewContentDataCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewContentDataCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c NewContentDataCmdMysql) TableName() string                     { return "content_data" }

// Params returns the command parameters.
func (c NewContentDataCmdMysql) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewContentDataCmdMysql) GetID(row mdbm.ContentData) string     { return string(row.ContentDataID) }

// Execute creates the record in the database.
func (c NewContentDataCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.ContentData, error) {
	id := types.NewContentID()
	queries := mdbm.New(tx)
	err := queries.CreateContentData(ctx, mdbm.CreateContentDataParams{
		ContentDataID: id,
		RouteID:       c.params.RouteID,
		ParentID:      c.params.ParentID,
		FirstChildID:  c.params.FirstChildID,
		NextSiblingID: c.params.NextSiblingID,
		PrevSiblingID: c.params.PrevSiblingID,
		DatatypeID:    c.params.DatatypeID,
		AuthorID:      c.params.AuthorID,
		Status:        c.params.Status,
		DateCreated:   c.params.DateCreated,
		DateModified:  c.params.DateModified,
	})
	if err != nil {
		return mdbm.ContentData{}, fmt.Errorf("execute create content_data: %w", err)
	}
	return queries.GetContentData(ctx, mdbm.GetContentDataParams{ContentDataID: id})
}

// NewContentDataCmd creates a new audited create command.
func (d MysqlDatabase) NewContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentDataParams) NewContentDataCmdMysql {
	return NewContentDataCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateContentDataCmdMysql is an audited update command for content_data (MySQL).
type UpdateContentDataCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateContentDataCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c UpdateContentDataCmdMysql) TableName() string                     { return "content_data" }

// Params returns the command parameters.
func (c UpdateContentDataCmdMysql) Params() any                           { return c.params }

// GetID returns the record ID being updated.
func (c UpdateContentDataCmdMysql) GetID() string                         { return string(c.params.ContentDataID) }

// GetBefore retrieves the record before modification.
func (c UpdateContentDataCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentData, error) {
	queries := mdbm.New(tx)
	return queries.GetContentData(ctx, mdbm.GetContentDataParams{ContentDataID: c.params.ContentDataID})
}

// Execute updates the record in the database.
func (c UpdateContentDataCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateContentData(ctx, mdbm.UpdateContentDataParams{
		RouteID:       c.params.RouteID,
		ParentID:      c.params.ParentID,
		FirstChildID:  c.params.FirstChildID,
		NextSiblingID: c.params.NextSiblingID,
		PrevSiblingID: c.params.PrevSiblingID,
		DatatypeID:    c.params.DatatypeID,
		AuthorID:      c.params.AuthorID,
		Status:        c.params.Status,
		DateCreated:   c.params.DateCreated,
		DateModified:  c.params.DateModified,
		ContentDataID: c.params.ContentDataID,
	})
}

// UpdateContentDataCmd creates a new audited update command.
func (d MysqlDatabase) UpdateContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentDataParams) UpdateContentDataCmdMysql {
	return UpdateContentDataCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// DeleteContentDataCmdMysql is an audited delete command for content_data (MySQL).
type DeleteContentDataCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteContentDataCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c DeleteContentDataCmdMysql) TableName() string                     { return "content_data" }

// GetID returns the record ID being deleted.
func (c DeleteContentDataCmdMysql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteContentDataCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentData, error) {
	queries := mdbm.New(tx)
	return queries.GetContentData(ctx, mdbm.GetContentDataParams{ContentDataID: c.id})
}

// Execute deletes the record from the database.
func (c DeleteContentDataCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteContentData(ctx, mdbm.DeleteContentDataParams{ContentDataID: c.id})
}

// DeleteContentDataCmd creates a new audited delete command.
func (d MysqlDatabase) DeleteContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentID) DeleteContentDataCmdMysql {
	return DeleteContentDataCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

// NewContentDataCmdPsql is an audited create command for content_data (PostgreSQL).
type NewContentDataCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewContentDataCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c NewContentDataCmdPsql) TableName() string                     { return "content_data" }

// Params returns the command parameters.
func (c NewContentDataCmdPsql) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewContentDataCmdPsql) GetID(row mdbp.ContentData) string     { return string(row.ContentDataID) }

// Execute creates the record in the database.
func (c NewContentDataCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.ContentData, error) {
	queries := mdbp.New(tx)
	return queries.CreateContentData(ctx, mdbp.CreateContentDataParams{
		ContentDataID: types.NewContentID(),
		ParentID:      c.params.ParentID,
		FirstChildID:  c.params.FirstChildID,
		NextSiblingID: c.params.NextSiblingID,
		PrevSiblingID: c.params.PrevSiblingID,
		RouteID:       c.params.RouteID,
		DatatypeID:    c.params.DatatypeID,
		AuthorID:      c.params.AuthorID,
		Status:        c.params.Status,
		DateCreated:   c.params.DateCreated,
		DateModified:  c.params.DateModified,
	})
}

// NewContentDataCmd creates a new audited create command.
func (d PsqlDatabase) NewContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentDataParams) NewContentDataCmdPsql {
	return NewContentDataCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateContentDataCmdPsql is an audited update command for content_data (PostgreSQL).
type UpdateContentDataCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateContentDataParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateContentDataCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c UpdateContentDataCmdPsql) TableName() string                     { return "content_data" }

// Params returns the command parameters.
func (c UpdateContentDataCmdPsql) Params() any                           { return c.params }

// GetID returns the record ID being updated.
func (c UpdateContentDataCmdPsql) GetID() string                         { return string(c.params.ContentDataID) }

// GetBefore retrieves the record before modification.
func (c UpdateContentDataCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentData, error) {
	queries := mdbp.New(tx)
	return queries.GetContentData(ctx, mdbp.GetContentDataParams{ContentDataID: c.params.ContentDataID})
}

// Execute updates the record in the database.
func (c UpdateContentDataCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateContentData(ctx, mdbp.UpdateContentDataParams{
		RouteID:       c.params.RouteID,
		ParentID:      c.params.ParentID,
		FirstChildID:  c.params.FirstChildID,
		NextSiblingID: c.params.NextSiblingID,
		PrevSiblingID: c.params.PrevSiblingID,
		DatatypeID:    c.params.DatatypeID,
		AuthorID:      c.params.AuthorID,
		Status:        c.params.Status,
		DateCreated:   c.params.DateCreated,
		DateModified:  c.params.DateModified,
		ContentDataID: c.params.ContentDataID,
	})
}

// UpdateContentDataCmd creates a new audited update command.
func (d PsqlDatabase) UpdateContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateContentDataParams) UpdateContentDataCmdPsql {
	return UpdateContentDataCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// DeleteContentDataCmdPsql is an audited delete command for content_data (PostgreSQL).
type DeleteContentDataCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteContentDataCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the target table name.
func (c DeleteContentDataCmdPsql) TableName() string                     { return "content_data" }

// GetID returns the record ID being deleted.
func (c DeleteContentDataCmdPsql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteContentDataCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentData, error) {
	queries := mdbp.New(tx)
	return queries.GetContentData(ctx, mdbp.GetContentDataParams{ContentDataID: c.id})
}

// Execute deletes the record from the database.
func (c DeleteContentDataCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteContentData(ctx, mdbp.DeleteContentDataParams{ContentDataID: c.id})
}

// DeleteContentDataCmd creates a new audited delete command.
func (d PsqlDatabase) DeleteContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentID) DeleteContentDataCmdPsql {
	return DeleteContentDataCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
