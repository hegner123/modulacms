package db

import (
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
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
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
	ContentDataID types.ContentID          `json:"content_data_id"`
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

func (d Database) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentData: %v\n", err)
	}
	return d.MapContentData(row)
}

func (d Database) DeleteContentData(id types.ContentID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentData(d.Context, mdb.DeleteContentDataParams{ContentDataID: id})
	if err != nil {
		return fmt.Errorf("failed to delete content data: %v\n", id)
	}
	return nil
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

func (d Database) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content data %v\n", s.ContentDataID)
	return &u, nil
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

func (d MysqlDatabase) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentData: %v\n", err)
	}
	row, err := queries.GetLastContentData(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted ContentData: %v\n", err)
	}
	return d.MapContentData(row)
}

func (d MysqlDatabase) DeleteContentData(id types.ContentID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentData(d.Context, mdbm.DeleteContentDataParams{ContentDataID: id})
	if err != nil {
		return fmt.Errorf("failed to delete content data: %v", id)
	}
	return nil
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

func (d MysqlDatabase) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content data %v\n", s.ContentDataID)
	return &u, nil
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

func (d PsqlDatabase) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentData: %v\n", err)
	}
	return d.MapContentData(row)
}

func (d PsqlDatabase) DeleteContentData(id types.ContentID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentData(d.Context, mdbp.DeleteContentDataParams{ContentDataID: id})
	if err != nil {
		return fmt.Errorf("failed to delete content data: %v", id)
	}
	return nil
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

func (d PsqlDatabase) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content data %v\n", s.ContentDataID)
	return &u, nil
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
