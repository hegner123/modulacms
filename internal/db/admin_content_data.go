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

type AdminContentData struct {
	AdminContentDataID types.AdminContentID          `json:"admin_content_data_id"`
	ParentID           types.NullableContentID       `json:"parent_id"`
	FirstChildID       sql.NullString                `json:"first_child_id"`
	NextSiblingID      sql.NullString                `json:"next_sibling_id"`
	PrevSiblingID      sql.NullString                `json:"prev_sibling_id"`
	AdminRouteID       string                        `json:"admin_route_id"`
	AdminDatatypeID    types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           types.NullableUserID          `json:"author_id"`
	DateCreated        types.Timestamp               `json:"date_created"`
	DateModified       types.Timestamp               `json:"date_modified"`
}
type CreateAdminContentDataParams struct {
	ParentID        types.NullableContentID       `json:"parent_id"`
	FirstChildID    sql.NullString                `json:"first_child_id"`
	NextSiblingID   sql.NullString                `json:"next_sibling_id"`
	PrevSiblingID   sql.NullString                `json:"prev_sibling_id"`
	AdminRouteID    string                        `json:"admin_route_id"`
	AdminDatatypeID types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID        types.NullableUserID          `json:"author_id"`
	DateCreated     types.Timestamp               `json:"date_created"`
	DateModified    types.Timestamp               `json:"date_modified"`
}
type UpdateAdminContentDataParams struct {
	ParentID           types.NullableContentID       `json:"parent_id"`
	FirstChildID       sql.NullString                `json:"first_child_id"`
	NextSiblingID      sql.NullString                `json:"next_sibling_id"`
	PrevSiblingID      sql.NullString                `json:"prev_sibling_id"`
	AdminRouteID       string                        `json:"admin_route_id"`
	AdminDatatypeID    types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           types.NullableUserID          `json:"author_id"`
	DateCreated        types.Timestamp               `json:"date_created"`
	DateModified       types.Timestamp               `json:"date_modified"`
	AdminContentDataID types.AdminContentID          `json:"admin_content_data_id"`
}
// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapAdminContentDataJSON converts AdminContentData to ContentDataJSON for tree building.
// Maps admin IDs into the public ContentDataJSON shape so BuildNodes works unchanged.
func MapAdminContentDataJSON(a AdminContentData) ContentDataJSON {
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
		ContentDataID: a.AdminContentDataID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  firstChildID,
		NextSiblingID: nextSiblingID,
		PrevSiblingID: prevSiblingID,
		RouteID:       a.AdminRouteID,
		DatatypeID:    a.AdminDatatypeID.String(),
		AuthorID:      a.AuthorID.String(),
		DateCreated:   a.DateCreated.String(),
		DateModified:  a.DateModified.String(),
	}
}

// MapStringAdminContentData converts AdminContentData to StringAdminContentData for table display
func MapStringAdminContentData(a AdminContentData) StringAdminContentData {
	return StringAdminContentData{
		AdminContentDataID: a.AdminContentDataID.String(),
		ParentID:           a.ParentID.String(),
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID.String(),
		AuthorID:           a.AuthorID.String(),
		DateCreated:        a.DateCreated.String(),
		DateModified:       a.DateModified.String(),
		History:            "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminContentData(a mdb.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}

func (d Database) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdb.CreateAdminContentDataParams {
	return mdb.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}

func (d Database) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdb.UpdateAdminContentDataParams {
	return mdb.UpdateAdminContentDataParams{
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		AdminContentDataID: a.AdminContentDataID,
	}
}


// QUERIES

func (d Database) CountAdminContentData() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d Database) CreateAdminContentData(s CreateAdminContentDataParams) AdminContentData {
	params := d.MapCreateAdminContentDataParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminContentData(row)
}
func (d Database) CreateAdminContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d Database) DeleteAdminContentData(id types.AdminContentID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, mdb.DeleteAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin content data: %v\n ", id)
	}

	return nil
}
func (d Database) GetAdminContentData(id types.AdminContentID) (*AdminContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, mdb.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d Database) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListAdminContentDataByRoute(id string) (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, mdb.ListAdminContentDataByRouteParams{AdminRouteID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) UpdateAdminContentData(s UpdateAdminContentDataParams) (*string, error) {
	params := d.MapUpdateAdminContentDataParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.AdminDatatypeID)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminContentData(a mdbm.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d MysqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbm.CreateAdminContentDataParams {
	return mdbm.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d MysqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbm.UpdateAdminContentDataParams {
	return mdbm.UpdateAdminContentDataParams{
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		AdminContentDataID: a.AdminContentDataID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminContentData() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateAdminContentData(s CreateAdminContentDataParams) AdminContentData {
	params := d.MapCreateAdminContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminContentData: %v\n", err)
	}
	row, err := queries.GetLastAdminContentData(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminContentData: %v\n", err)
	}
	return d.MapAdminContentData(row)
}
func (d MysqlDatabase) CreateAdminContentDataTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d MysqlDatabase) DeleteAdminContentData(id types.AdminContentID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, mdbm.DeleteAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin content data: %v ", id)
	}

	return nil
}
func (d MysqlDatabase) GetAdminContentData(id types.AdminContentID) (*AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, mdbm.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentDataByRoute(id string) (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, mdbm.ListAdminContentDataByRouteParams{AdminRouteID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateAdminContentData(s UpdateAdminContentDataParams) (*string, error) {
	params := d.MapUpdateAdminContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.AdminDatatypeID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS

func (d PsqlDatabase) MapAdminContentData(a mdbp.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d PsqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbp.CreateAdminContentDataParams {
	return mdbp.CreateAdminContentDataParams{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}
func (d PsqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbp.UpdateAdminContentDataParams {
	return mdbp.UpdateAdminContentDataParams{
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		AdminContentDataID: a.AdminContentDataID,
	}
}

// /QUERIES

func (d PsqlDatabase) CountAdminContentData() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateAdminContentData(s CreateAdminContentDataParams) AdminContentData {
	params := d.MapCreateAdminContentDataParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminContentData(row)
}
func (d PsqlDatabase) CreateAdminContentDataTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminContentData(id types.AdminContentID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, mdbp.DeleteAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin content data: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) GetAdminContentData(id types.AdminContentID) (*AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, mdbp.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentDataByRoute(id string) (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, mdbp.ListAdminContentDataByRouteParams{AdminRouteID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) UpdateAdminContentData(s UpdateAdminContentDataParams) (*string, error) {
	params := d.MapUpdateAdminContentDataParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.AdminDatatypeID)
	return &u, nil
}
