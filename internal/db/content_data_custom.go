package db

import (
	"context"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

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
	PublishedAt   string `json:"published_at"`
	PublishedBy   string `json:"published_by"`
	PublishAt     string `json:"publish_at"`
	Revision      int64  `json:"revision"`
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
		PublishedAt:   a.PublishedAt.String(),
		PublishedBy:   a.PublishedBy.String(),
		PublishAt:     a.PublishAt.String(),
		Revision:      a.Revision,
	}
}

///////////////////////////////
// LIST BY DATATYPE
//////////////////////////////

// ListContentDataByDatatypeID returns all content data for a given datatype (SQLite).
func (d Database) ListContentDataByDatatypeID(datatypeID types.DatatypeID) (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	dtID := types.NullableDatatypeID{ID: datatypeID, Valid: true}
	rows, err := queries.ListContentDataByDatatypeID(d.Context, mdb.ListContentDataByDatatypeIDParams{DatatypeID: dtID})
	if err != nil {
		return nil, fmt.Errorf("failed to list content data by datatype %s: %w", datatypeID, err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// ListContentDataByDatatypeID returns all content data for a given datatype (MySQL).
func (d MysqlDatabase) ListContentDataByDatatypeID(datatypeID types.DatatypeID) (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	dtID := types.NullableDatatypeID{ID: datatypeID, Valid: true}
	rows, err := queries.ListContentDataByDatatypeID(d.Context, mdbm.ListContentDataByDatatypeIDParams{DatatypeID: dtID})
	if err != nil {
		return nil, fmt.Errorf("failed to list content data by datatype %s: %w", datatypeID, err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// ListContentDataByDatatypeID returns all content data for a given datatype (PostgreSQL).
func (d PsqlDatabase) ListContentDataByDatatypeID(datatypeID types.DatatypeID) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	dtID := types.NullableDatatypeID{ID: datatypeID, Valid: true}
	rows, err := queries.ListContentDataByDatatypeID(d.Context, mdbp.ListContentDataByDatatypeIDParams{DatatypeID: dtID})
	if err != nil {
		return nil, fmt.Errorf("failed to list content data by datatype %s: %w", datatypeID, err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// ListContentDataGlobal returns root-level content data whose datatype is _global (SQLite).
func (d Database) ListContentDataGlobal() (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataGlobal(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list global content data: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// ListContentDataGlobal returns root-level content data whose datatype is _global (MySQL).
func (d MysqlDatabase) ListContentDataGlobal() (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataGlobal(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list global content data: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// ListContentDataGlobal returns root-level content data whose datatype is _global (PostgreSQL).
func (d PsqlDatabase) ListContentDataGlobal() (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataGlobal(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list global content data: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

///////////////////////////////
// _root CONTENT SUMMARY
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

// SQLITE - MapRootContentSummary converts a sqlc-generated type to the wrapper type.
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

// MYSQL - MapRootContentSummary converts a sqlc-generated type to the wrapper type.
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

// PSQL - MapRootContentSummary converts a sqlc-generated type to the wrapper type.
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
// TOP-LEVEL CONTENT DATA
//////////////////////////////

// ContentDataTopLevel extends ContentData with resolved display names from JOINs.
type ContentDataTopLevel struct {
	ContentData
	AuthorName    string     `json:"author_name"`
	RouteSlug     types.Slug `json:"route_slug"`
	RouteTitle    string     `json:"route_title"`
	DatatypeLabel string     `json:"datatype_label"`
}

// SQLITE

func (d Database) mapContentDataTopLevel(a mdb.ListContentDataTopLevelPaginatedRow) ContentDataTopLevel {
	return ContentDataTopLevel{
		ContentData: d.MapContentData(mdb.ContentData{
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
			PublishedAt:   a.PublishedAt,
			PublishedBy:   a.PublishedBy,
			PublishAt:     a.PublishAt,
			Revision:      a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
	}
}

// ListContentDataTopLevelPaginated returns paginated content entries that have a route or _root datatype.
func (d Database) ListContentDataTopLevelPaginated(params PaginationParams) (*[]ContentDataTopLevel, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataTopLevelPaginated(d.Context, mdb.ListContentDataTopLevelPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level ContentData: %v", err)
	}
	res := make([]ContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapContentDataTopLevel(v))
	}
	return &res, nil
}

// CountContentDataTopLevel returns the count of content entries that have a route or _root datatype.
func (d Database) CountContentDataTopLevel() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentDataTopLevel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level ContentData: %v", err)
	}
	return &c, nil
}

func (d Database) mapContentDataTopLevelByStatus(a mdb.ListContentDataTopLevelPaginatedByStatusRow) ContentDataTopLevel {
	return ContentDataTopLevel{
		ContentData: d.MapContentData(mdb.ContentData{
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
			PublishedAt:   a.PublishedAt,
			PublishedBy:   a.PublishedBy,
			PublishAt:     a.PublishAt,
			Revision:      a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
	}
}

// ListContentDataTopLevelPaginatedByStatus returns paginated top-level content filtered by status.
func (d Database) ListContentDataTopLevelPaginatedByStatus(params PaginationParams, status types.ContentStatus) (*[]ContentDataTopLevel, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataTopLevelPaginatedByStatus(d.Context, mdb.ListContentDataTopLevelPaginatedByStatusParams{
		Status: status,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level ContentData by status: %v", err)
	}
	res := make([]ContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapContentDataTopLevelByStatus(v))
	}
	return &res, nil
}

// CountContentDataTopLevelByStatus returns the count of top-level content with a given status.
func (d Database) CountContentDataTopLevelByStatus(status types.ContentStatus) (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentDataTopLevelByStatus(d.Context, mdb.CountContentDataTopLevelByStatusParams{
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level ContentData by status: %v", err)
	}
	return &c, nil
}

// MYSQL

func (d MysqlDatabase) mapContentDataTopLevel(a mdbm.ListContentDataTopLevelPaginatedRow) ContentDataTopLevel {
	return ContentDataTopLevel{
		ContentData: d.MapContentData(mdbm.ContentData{
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
			PublishedAt:   a.PublishedAt,
			PublishedBy:   a.PublishedBy,
			PublishAt:     a.PublishAt,
			Revision:      a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
	}
}

// ListContentDataTopLevelPaginated returns paginated content entries that have a route or _root datatype.
func (d MysqlDatabase) ListContentDataTopLevelPaginated(params PaginationParams) (*[]ContentDataTopLevel, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataTopLevelPaginated(d.Context, mdbm.ListContentDataTopLevelPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level ContentData: %v", err)
	}
	res := make([]ContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapContentDataTopLevel(v))
	}
	return &res, nil
}

// CountContentDataTopLevel returns the count of content entries that have a route or _root datatype.
func (d MysqlDatabase) CountContentDataTopLevel() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentDataTopLevel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level ContentData: %v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) mapContentDataTopLevelByStatus(a mdbm.ListContentDataTopLevelPaginatedByStatusRow) ContentDataTopLevel {
	return ContentDataTopLevel{
		ContentData: d.MapContentData(mdbm.ContentData{
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
			PublishedAt:   a.PublishedAt,
			PublishedBy:   a.PublishedBy,
			PublishAt:     a.PublishAt,
			Revision:      a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
	}
}

// ListContentDataTopLevelPaginatedByStatus returns paginated top-level content filtered by status.
func (d MysqlDatabase) ListContentDataTopLevelPaginatedByStatus(params PaginationParams, status types.ContentStatus) (*[]ContentDataTopLevel, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataTopLevelPaginatedByStatus(d.Context, mdbm.ListContentDataTopLevelPaginatedByStatusParams{
		Status: status,
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level ContentData by status: %v", err)
	}
	res := make([]ContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapContentDataTopLevelByStatus(v))
	}
	return &res, nil
}

// CountContentDataTopLevelByStatus returns the count of top-level content with a given status.
func (d MysqlDatabase) CountContentDataTopLevelByStatus(status types.ContentStatus) (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentDataTopLevelByStatus(d.Context, mdbm.CountContentDataTopLevelByStatusParams{
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level ContentData by status: %v", err)
	}
	return &c, nil
}

// PSQL

func (d PsqlDatabase) mapContentDataTopLevel(a mdbp.ListContentDataTopLevelPaginatedRow) ContentDataTopLevel {
	return ContentDataTopLevel{
		ContentData: d.MapContentData(mdbp.ContentData{
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
			PublishedAt:   a.PublishedAt,
			PublishedBy:   a.PublishedBy,
			PublishAt:     a.PublishAt,
			Revision:      a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
	}
}

// ListContentDataTopLevelPaginated returns paginated content entries that have a route or _root datatype.
func (d PsqlDatabase) ListContentDataTopLevelPaginated(params PaginationParams) (*[]ContentDataTopLevel, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataTopLevelPaginated(d.Context, mdbp.ListContentDataTopLevelPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level ContentData: %v", err)
	}
	res := make([]ContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapContentDataTopLevel(v))
	}
	return &res, nil
}

// CountContentDataTopLevel returns the count of content entries that have a route or _root datatype.
func (d PsqlDatabase) CountContentDataTopLevel() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentDataTopLevel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level ContentData: %v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) mapContentDataTopLevelByStatus(a mdbp.ListContentDataTopLevelPaginatedByStatusRow) ContentDataTopLevel {
	return ContentDataTopLevel{
		ContentData: d.MapContentData(mdbp.ContentData{
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
			PublishedAt:   a.PublishedAt,
			PublishedBy:   a.PublishedBy,
			PublishAt:     a.PublishAt,
			Revision:      a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
	}
}

// ListContentDataTopLevelPaginatedByStatus returns paginated top-level content filtered by status.
func (d PsqlDatabase) ListContentDataTopLevelPaginatedByStatus(params PaginationParams, status types.ContentStatus) (*[]ContentDataTopLevel, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataTopLevelPaginatedByStatus(d.Context, mdbp.ListContentDataTopLevelPaginatedByStatusParams{
		Status: status,
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level ContentData by status: %v", err)
	}
	res := make([]ContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapContentDataTopLevelByStatus(v))
	}
	return &res, nil
}

// CountContentDataTopLevelByStatus returns the count of top-level content with a given status.
func (d PsqlDatabase) CountContentDataTopLevelByStatus(status types.ContentStatus) (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentDataTopLevelByStatus(d.Context, mdbp.CountContentDataTopLevelByStatusParams{
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level ContentData by status: %v", err)
	}
	return &c, nil
}

///////////////////////////////
// CONTENT DATA DESCENDANTS
//////////////////////////////

// GetContentDataDescendants returns a node and all its descendants via recursive CTE.
func (d Database) GetContentDataDescendants(ctx context.Context, id types.ContentID) (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetContentDataDescendants(ctx, mdb.GetContentDataDescendantsParams{ContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content data descendants: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// GetContentDataDescendants returns a node and all its descendants via recursive CTE.
func (d MysqlDatabase) GetContentDataDescendants(ctx context.Context, id types.ContentID) (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetContentDataDescendants(ctx, mdbm.GetContentDataDescendantsParams{ContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content data descendants: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// GetContentDataDescendants returns a node and all its descendants via recursive CTE.
func (d PsqlDatabase) GetContentDataDescendants(ctx context.Context, id types.ContentID) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetContentDataDescendants(ctx, mdbp.GetContentDataDescendantsParams{ContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content data descendants: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}
