package db

import (
	"context"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// nullableAdminContentIDStringEmpty returns "" when the nullable ID is invalid,
// and the ID string when valid. Used for sibling pointer fields that the
// tree builder checks with == "".
func nullableAdminContentIDStringEmpty(n types.NullableAdminContentID) string {
	if !n.Valid {
		return ""
	}
	return n.ID.String()
}

// MapAdminContentDataJSON converts AdminContentData to ContentDataJSON for tree building.
// Maps admin IDs into the public ContentDataJSON shape so BuildNodes works unchanged.
func MapAdminContentDataJSON(a AdminContentData) ContentDataJSON {
	return ContentDataJSON{
		ContentDataID: a.AdminContentDataID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  nullableAdminContentIDStringEmpty(a.FirstChildID),
		NextSiblingID: nullableAdminContentIDStringEmpty(a.NextSiblingID),
		PrevSiblingID: nullableAdminContentIDStringEmpty(a.PrevSiblingID),
		RouteID:       a.AdminRouteID.String(),
		DatatypeID:    a.AdminDatatypeID.String(),
		AuthorID:      a.AuthorID.String(),
		Status:        string(a.Status),
		DateCreated:   a.DateCreated.String(),
		DateModified:  a.DateModified.String(),
	}
}

// ListAdminContentDataByRootID returns all admin content data for a given root_id (SQLite).
func (d Database) ListAdminContentDataByRootID(rootID types.NullableAdminContentID) (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRootID(d.Context, mdb.ListAdminContentDataByRootIDParams{RootID: rootID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data by root_id: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

///////////////////////////////
// TOP-LEVEL ADMIN CONTENT DATA
//////////////////////////////

// AdminContentDataTopLevel extends AdminContentData with resolved display names from JOINs.
type AdminContentDataTopLevel struct {
	AdminContentData
	AuthorName    string     `json:"author_name"`
	RouteSlug     types.Slug `json:"route_slug"`
	RouteTitle    string     `json:"route_title"`
	DatatypeLabel string     `json:"datatype_label"`
	DatatypeType  string     `json:"datatype_type"`
}

// SQLITE

func (d Database) mapAdminContentDataTopLevel(a mdb.ListAdminContentDataTopLevelPaginatedRow) AdminContentDataTopLevel {
	return AdminContentDataTopLevel{
		AdminContentData: d.MapAdminContentData(mdb.AdminContentData{
			AdminContentDataID: a.AdminContentDataID,
			ParentID:           a.ParentID,
			FirstChildID:       a.FirstChildID,
			NextSiblingID:      a.NextSiblingID,
			PrevSiblingID:      a.PrevSiblingID,
			RootID:             a.RootID,
			AdminRouteID:       a.AdminRouteID,
			AdminDatatypeID:    a.AdminDatatypeID,
			AuthorID:           a.AuthorID,
			Status:             a.Status,
			DateCreated:        a.DateCreated,
			DateModified:       a.DateModified,
			PublishedAt:        a.PublishedAt,
			PublishedBy:        a.PublishedBy,
			PublishAt:          a.PublishAt,
			Revision:           a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// ListAdminContentDataTopLevelPaginated returns paginated admin content entries that have a route or _root datatype.
func (d Database) ListAdminContentDataTopLevelPaginated(params PaginationParams) (*[]AdminContentDataTopLevel, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentDataTopLevelPaginated(d.Context, mdb.ListAdminContentDataTopLevelPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level AdminContentData: %v", err)
	}
	res := make([]AdminContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapAdminContentDataTopLevel(v))
	}
	return &res, nil
}

// CountAdminContentDataTopLevel returns the count of admin content entries that have a route or _root datatype.
func (d Database) CountAdminContentDataTopLevel() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentDataTopLevel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level AdminContentData: %v", err)
	}
	return &c, nil
}

///////////////////////////////
// ADMIN CONTENT DATA DESCENDANTS
//////////////////////////////

// GetAdminContentDataDescendants returns a node and all its descendants via recursive CTE (SQLite).
func (d Database) GetAdminContentDataDescendants(ctx context.Context, id types.AdminContentID) (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetAdminContentDataDescendants(ctx, mdb.GetAdminContentDataDescendantsParams{AdminContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content data descendants: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

// MYSQL

// ListAdminContentDataByRootID returns all admin content data for a given root_id (MySQL).
func (d MysqlDatabase) ListAdminContentDataByRootID(rootID types.NullableAdminContentID) (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRootID(d.Context, mdbm.ListAdminContentDataByRootIDParams{RootID: rootID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data by root_id: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

func (d MysqlDatabase) mapAdminContentDataTopLevel(a mdbm.ListAdminContentDataTopLevelPaginatedRow) AdminContentDataTopLevel {
	return AdminContentDataTopLevel{
		AdminContentData: d.MapAdminContentData(mdbm.AdminContentData{
			AdminContentDataID: a.AdminContentDataID,
			ParentID:           a.ParentID,
			FirstChildID:       a.FirstChildID,
			NextSiblingID:      a.NextSiblingID,
			PrevSiblingID:      a.PrevSiblingID,
			RootID:             a.RootID,
			AdminRouteID:       a.AdminRouteID,
			AdminDatatypeID:    a.AdminDatatypeID,
			AuthorID:           a.AuthorID,
			Status:             a.Status,
			DateCreated:        a.DateCreated,
			DateModified:       a.DateModified,
			PublishedAt:        a.PublishedAt,
			PublishedBy:        a.PublishedBy,
			PublishAt:          a.PublishAt,
			Revision:           a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// CountAdminContentDataTopLevel returns the count of admin content entries that have a route or _root datatype.
func (d MysqlDatabase) CountAdminContentDataTopLevel() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentDataTopLevel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level AdminContentData: %v", err)
	}
	return &c, nil
}

// GetAdminContentDataDescendants returns a node and all its descendants via recursive CTE (MySQL).
func (d MysqlDatabase) GetAdminContentDataDescendants(ctx context.Context, id types.AdminContentID) (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetAdminContentDataDescendants(ctx, mdbm.GetAdminContentDataDescendantsParams{AdminContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content data descendants: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

// ListAdminContentDataTopLevelPaginated returns paginated admin content entries that have a route or _root datatype.
func (d MysqlDatabase) ListAdminContentDataTopLevelPaginated(params PaginationParams) (*[]AdminContentDataTopLevel, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentDataTopLevelPaginated(d.Context, mdbm.ListAdminContentDataTopLevelPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level AdminContentData: %v", err)
	}
	res := make([]AdminContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapAdminContentDataTopLevel(v))
	}
	return &res, nil
}

// PSQL

// ListAdminContentDataByRootID returns all admin content data for a given root_id (PostgreSQL).
func (d PsqlDatabase) ListAdminContentDataByRootID(rootID types.NullableAdminContentID) (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRootID(d.Context, mdbp.ListAdminContentDataByRootIDParams{RootID: rootID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data by root_id: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

func (d PsqlDatabase) mapAdminContentDataTopLevel(a mdbp.ListAdminContentDataTopLevelPaginatedRow) AdminContentDataTopLevel {
	return AdminContentDataTopLevel{
		AdminContentData: d.MapAdminContentData(mdbp.AdminContentData{
			AdminContentDataID: a.AdminContentDataID,
			ParentID:           a.ParentID,
			FirstChildID:       a.FirstChildID,
			NextSiblingID:      a.NextSiblingID,
			PrevSiblingID:      a.PrevSiblingID,
			RootID:             a.RootID,
			AdminRouteID:       a.AdminRouteID,
			AdminDatatypeID:    a.AdminDatatypeID,
			AuthorID:           a.AuthorID,
			Status:             a.Status,
			DateCreated:        a.DateCreated,
			DateModified:       a.DateModified,
			PublishedAt:        a.PublishedAt,
			PublishedBy:        a.PublishedBy,
			PublishAt:          a.PublishAt,
			Revision:           a.Revision,
		}),
		AuthorName:    a.AuthorName.String,
		RouteSlug:     a.RouteSlug,
		RouteTitle:    a.RouteTitle,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// CountAdminContentDataTopLevel returns the count of admin content entries that have a route or _root datatype.
func (d PsqlDatabase) CountAdminContentDataTopLevel() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentDataTopLevel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count top-level AdminContentData: %v", err)
	}
	return &c, nil
}

// GetAdminContentDataDescendants returns a node and all its descendants via recursive CTE (PostgreSQL).
func (d PsqlDatabase) GetAdminContentDataDescendants(ctx context.Context, id types.AdminContentID) (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetAdminContentDataDescendants(ctx, mdbp.GetAdminContentDataDescendantsParams{AdminContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content data descendants: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

// ListAdminContentDataTopLevelPaginated returns paginated admin content entries that have a route or _root datatype.
func (d PsqlDatabase) ListAdminContentDataTopLevelPaginated(params PaginationParams) (*[]AdminContentDataTopLevel, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentDataTopLevelPaginated(d.Context, mdbp.ListAdminContentDataTopLevelPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level AdminContentData: %v", err)
	}
	res := make([]AdminContentDataTopLevel, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.mapAdminContentDataTopLevel(v))
	}
	return &res, nil
}
