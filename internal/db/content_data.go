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

type ContentData struct {
	ContentDataID types.ContentID          `json:"content_data_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  sql.NullString           `json:"first_child_id"`
	NextSiblingID sql.NullString           `json:"next_sibling_id"`
	PrevSiblingID sql.NullString           `json:"prev_sibling_id"`
	RouteID       types.NullableRouteID    `json:"route_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	AuthorID      types.NullableUserID     `json:"author_id"`
	Status        types.ContentStatus      `json:"status"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
}

type CreateContentDataParams struct {
	RouteID       types.NullableRouteID    `json:"route_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  sql.NullString           `json:"first_child_id"`
	NextSiblingID sql.NullString           `json:"next_sibling_id"`
	PrevSiblingID sql.NullString           `json:"prev_sibling_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	AuthorID      types.NullableUserID     `json:"author_id"`
	Status        types.ContentStatus      `json:"status"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
}

type UpdateContentDataParams struct {
	RouteID       types.NullableRouteID    `json:"route_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  sql.NullString           `json:"first_child_id"`
	NextSiblingID sql.NullString           `json:"next_sibling_id"`
	PrevSiblingID sql.NullString           `json:"prev_sibling_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	AuthorID      types.NullableUserID     `json:"author_id"`
	Status        types.ContentStatus      `json:"status"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
	ContentDataID types.ContentID          `json:"content_data_id"`
}

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

// MapContentDataJSON converts ContentData to ContentDataJSON for JSON serialization
func MapContentDataJSON(a ContentData) ContentDataJSON {
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
		ContentDataID: a.ContentDataID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  firstChildID,
		NextSiblingID: nextSiblingID,
		PrevSiblingID: prevSiblingID,
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
	return StringContentData{
		ContentDataID: a.ContentDataID.String(),
		RouteID:       a.RouteID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  firstChildID,
		NextSiblingID: nextSiblingID,
		PrevSiblingID: prevSiblingID,
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

func (d Database) CountContentData() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d Database) CreateContentData(ctx context.Context, ac audited.AuditContext, s CreateContentDataParams) (*ContentData, error) {
	cmd := d.NewContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentData: %w", err)
	}
	r := d.MapContentData(result)
	return &r, nil
}

func (d Database) DeleteContentData(ctx context.Context, ac audited.AuditContext, id types.ContentID) error {
	cmd := d.DeleteContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetContentData(id types.ContentID) (*ContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentData(d.Context, mdb.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

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

func (d MysqlDatabase) CountContentData() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateContentDataTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateContentData(ctx context.Context, ac audited.AuditContext, s CreateContentDataParams) (*ContentData, error) {
	cmd := d.NewContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentData: %w", err)
	}
	r := d.MapContentData(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteContentData(ctx context.Context, ac audited.AuditContext, id types.ContentID) error {
	cmd := d.DeleteContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetContentData(id types.ContentID) (*ContentData, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentData(d.Context, mdbm.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

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

func (d PsqlDatabase) CountContentData() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateContentDataTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateContentData(ctx context.Context, ac audited.AuditContext, s CreateContentDataParams) (*ContentData, error) {
	cmd := d.NewContentDataCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contentData: %w", err)
	}
	r := d.MapContentData(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteContentData(ctx context.Context, ac audited.AuditContext, id types.ContentID) error {
	cmd := d.DeleteContentDataCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetContentData(id types.ContentID) (*ContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentData(d.Context, mdbp.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

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

func (c NewContentDataCmd) Context() context.Context              { return c.ctx }
func (c NewContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentDataCmd) Connection() *sql.DB                   { return c.conn }
func (c NewContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentDataCmd) TableName() string                     { return "content_data" }
func (c NewContentDataCmd) Params() any                           { return c.params }
func (c NewContentDataCmd) GetID(row mdb.ContentData) string      { return string(row.ContentDataID) }

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

func (c UpdateContentDataCmd) Context() context.Context              { return c.ctx }
func (c UpdateContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateContentDataCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateContentDataCmd) TableName() string                     { return "content_data" }
func (c UpdateContentDataCmd) Params() any                           { return c.params }
func (c UpdateContentDataCmd) GetID() string                         { return string(c.params.ContentDataID) }

func (c UpdateContentDataCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentData, error) {
	queries := mdb.New(tx)
	return queries.GetContentData(ctx, mdb.GetContentDataParams{ContentDataID: c.params.ContentDataID})
}

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

func (c DeleteContentDataCmd) Context() context.Context              { return c.ctx }
func (c DeleteContentDataCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentDataCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentDataCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentDataCmd) TableName() string                     { return "content_data" }
func (c DeleteContentDataCmd) GetID() string                         { return string(c.id) }

func (c DeleteContentDataCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentData, error) {
	queries := mdb.New(tx)
	return queries.GetContentData(ctx, mdb.GetContentDataParams{ContentDataID: c.id})
}

func (c DeleteContentDataCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteContentData(ctx, mdb.DeleteContentDataParams{ContentDataID: c.id})
}

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

func (c NewContentDataCmdMysql) Context() context.Context              { return c.ctx }
func (c NewContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentDataCmdMysql) TableName() string                     { return "content_data" }
func (c NewContentDataCmdMysql) Params() any                           { return c.params }
func (c NewContentDataCmdMysql) GetID(row mdbm.ContentData) string     { return string(row.ContentDataID) }

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

func (c UpdateContentDataCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateContentDataCmdMysql) TableName() string                     { return "content_data" }
func (c UpdateContentDataCmdMysql) Params() any                           { return c.params }
func (c UpdateContentDataCmdMysql) GetID() string                         { return string(c.params.ContentDataID) }

func (c UpdateContentDataCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentData, error) {
	queries := mdbm.New(tx)
	return queries.GetContentData(ctx, mdbm.GetContentDataParams{ContentDataID: c.params.ContentDataID})
}

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

func (c DeleteContentDataCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteContentDataCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentDataCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentDataCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentDataCmdMysql) TableName() string                     { return "content_data" }
func (c DeleteContentDataCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteContentDataCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentData, error) {
	queries := mdbm.New(tx)
	return queries.GetContentData(ctx, mdbm.GetContentDataParams{ContentDataID: c.id})
}

func (c DeleteContentDataCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteContentData(ctx, mdbm.DeleteContentDataParams{ContentDataID: c.id})
}

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

func (c NewContentDataCmdPsql) Context() context.Context              { return c.ctx }
func (c NewContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewContentDataCmdPsql) TableName() string                     { return "content_data" }
func (c NewContentDataCmdPsql) Params() any                           { return c.params }
func (c NewContentDataCmdPsql) GetID(row mdbp.ContentData) string     { return string(row.ContentDataID) }

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

func (c UpdateContentDataCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateContentDataCmdPsql) TableName() string                     { return "content_data" }
func (c UpdateContentDataCmdPsql) Params() any                           { return c.params }
func (c UpdateContentDataCmdPsql) GetID() string                         { return string(c.params.ContentDataID) }

func (c UpdateContentDataCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentData, error) {
	queries := mdbp.New(tx)
	return queries.GetContentData(ctx, mdbp.GetContentDataParams{ContentDataID: c.params.ContentDataID})
}

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

func (c DeleteContentDataCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteContentDataCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteContentDataCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteContentDataCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteContentDataCmdPsql) TableName() string                     { return "content_data" }
func (c DeleteContentDataCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteContentDataCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentData, error) {
	queries := mdbp.New(tx)
	return queries.GetContentData(ctx, mdbp.GetContentDataParams{ContentDataID: c.id})
}

func (c DeleteContentDataCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteContentData(ctx, mdbp.DeleteContentDataParams{ContentDataID: c.id})
}

func (d PsqlDatabase) DeleteContentDataCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentID) DeleteContentDataCmdPsql {
	return DeleteContentDataCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
