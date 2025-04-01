package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/db-mysql"
	mdbp "github.com/hegner123/modulacms/db-psql"
	mdb "github.com/hegner123/modulacms/db-sqlite"
)

///////////////////////////////
//STRUCTS
//////////////////////////////
type AdminRoutes struct {
	AdminRouteID int64          `json:"admin_route_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}
type CreateAdminRouteParams struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
    AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}
type UpdateAdminRouteParams struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
	Slug_2       string         `json:"slug_2"`
}
type UtilityGetAdminRoutesRow struct {
	AdminRouteID int64  `json:"admin_route_id"`
	Slug         string `json:"slug"`
}
type AdminRoutesHistoryEntry struct {
	AdminRouteID int64          `json:"admin_route_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}
type CreateAdminRouteFormParams struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
    AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}
type UpdateAdminRouteFormParams struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
	Slug_2       string `json:"slug_2"`
}

///////////////////////////////
//GENERIC
//////////////////////////////
func MapCreateAdminRouteParams(a CreateAdminRouteFormParams) CreateAdminRouteParams {
	return CreateAdminRouteParams{
		AuthorID:     Si(a.AuthorID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       Si(a.Status),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
	}
}
func MapUpdateAdminRouteParams(a UpdateAdminRouteFormParams) UpdateAdminRouteParams {
	return UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       Si(a.Status),
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
		Slug_2:       a.Slug_2,
	}
}
func MapStringAdminRoute(a AdminRoutes) StringAdminRoutes {
	return StringAdminRoutes{
		AdminRouteID: strconv.FormatInt(a.AdminRouteID, 10),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       strconv.FormatInt(a.Status, 10),
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
		History:      a.History.String,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

///MAPS
func (d Database) MapAdminRoute(a mdb.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}

func (d Database) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdb.CreateAdminRouteParams {
	return mdb.CreateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
        AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}
func (d Database) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdb.UpdateAdminRouteParams {
	return mdb.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
		Slug_2:       a.Slug_2,
	}
}
///QUERIES
func (d Database) CountAdminRoutes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminRoute(s CreateAdminRouteParams) AdminRoutes {
	params := d.MapCreateAdminRouteParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminRoute  %v \n", err)
	}

	return d.MapAdminRoute(row)
}

func (d Database) CreateAdminRouteTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminRouteTable(d.Context)
	return err
}

func (d Database) DeleteAdminRoute(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Route: %v ", id)
	}

	return nil
}

func (d Database) GetAdminRoute(slug string) (*AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}

func (d Database) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminRoute(s UpdateAdminRouteParams) (*string, error) {
	params := d.MapUpdateAdminRouteParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

///MAPS
func (d MysqlDatabase) MapAdminRoute(a mdbm.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: int64(a.AdminRouteID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(a.DateCreated.String()),
		DateModified: Ns(a.DateModified.String()),
		History:      a.History,
	}
}
func (d MysqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbm.CreateAdminRouteParams {
	return mdbm.CreateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
	}
}
func (d MysqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbm.UpdateAdminRouteParams {
	return mdbm.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
		Slug_2:       a.Slug_2,
	}
}
///QUERIES
func (d MysqlDatabase) CountAdminRoutes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminroute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminRoute(s CreateAdminRouteParams) AdminRoutes {
	params := d.MapCreateAdminRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminRoute: %v\n", err)
	}
	row, err := queries.GetLastAdminRoute(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminRoute: %v\n", err)
	}
	return d.MapAdminRoute(row)
}

func (d MysqlDatabase) CreateAdminRouteTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminRouteTable(d.Context)
	return err
}

func (d MysqlDatabase) DeleteAdminRoute(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Route: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) GetAdminRoute(slug string) (*AdminRoutes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateAdminRoute(s UpdateAdminRouteParams) (*string, error) {
	params := d.MapUpdateAdminRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

///MAPS
func (d PsqlDatabase) MapAdminRoute(a mdbp.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: int64(a.AdminRouteID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(Nt(a.DateCreated)),
		DateModified: Ns(Nt(a.DateModified)),
		History:      a.History,
	}
}
func (d PsqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbp.CreateAdminRouteParams {
	return mdbp.CreateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
	}
}
func (d PsqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbp.UpdateAdminRouteParams {
	return mdbp.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
		Slug_2:       a.Slug_2,
	}
}
///QUERIES
func (d PsqlDatabase) CountAdminRoutes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminroute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminRoute(s CreateAdminRouteParams) AdminRoutes {
	params := d.MapCreateAdminRouteParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminRoute  %v \n", err)
	}

	return d.MapAdminRoute(row)
}

func (d PsqlDatabase) CreateAdminRouteTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminRouteTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminRoute(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Route: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) GetAdminRoute(slug string) (*AdminRoutes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}

func (d PsqlDatabase) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminRoute(s UpdateAdminRouteParams) (*string, error) {
	params := d.MapUpdateAdminRouteParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}
