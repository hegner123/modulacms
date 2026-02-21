package db

import (
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
