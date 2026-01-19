package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/utility"
)

///////////////////////////////
//STRUCTS
//////////////////////////////
type Routes struct {
	RouteID      int64          `json:"route_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
    AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
    History      sql.NullString `json:"history"`
}

type CreateRouteParams struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
    AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
    History      sql.NullString `json:"history"`
}

type UpdateRouteParams struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
    History      sql.NullString `json:"history"`
	Slug_2       string         `json:"slug_2"`
}

type RoutesHistoryEntry struct {
	RouteID      any            `json:"route_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
    AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateRouteFormParams struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
    AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
    History      string `json:"history"`
}

type UpdateRouteFormParams struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
    History      string `json:"history"`
	Slug_2       string `json:"slug_2"`
}
type RoutesJSON struct {
	RouteID      int64          `json:"route_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
    AuthorID     int64          `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
    History      NullString `json:"history"`
}

type CreateRouteParamsJSON struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
    AuthorID     int64          `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
    History      NullString `json:"history"`
}

type UpdateRouteParamsJSON struct {
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
    History      NullString `json:"history"`
	Slug_2       string         `json:"slug_2"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateRouteParams(a CreateRouteFormParams) CreateRouteParams {
	return CreateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       StringToInt64(a.Status),
        AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
        History:      StringToNullString(a.History),
	}
}

func MapUpdateRouteParams(a UpdateRouteFormParams) UpdateRouteParams {
	return UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       StringToInt64(a.Status),
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
        History:      StringToNullString(a.History),
		Slug_2:       a.Slug_2,
	}
}

func MapStringRoute(a Routes) StringRoutes {
	return StringRoutes{
		RouteID:      strconv.FormatInt(a.RouteID, 10),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       strconv.FormatInt(a.Status, 10),
        AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  utility.NullToString(a.DateCreated),
		DateModified: utility.NullToString(a.DateModified),
        History:      utility.NullToString(a.History),
	}
}
func MapCreateRouteJSONParams(a CreateRouteParamsJSON) CreateRouteParams {
	return CreateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
        AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
        History:      a.History.NullString,
	}
}

func MapUpdateRouteJSONParams(a UpdateRouteParamsJSON) UpdateRouteParams {
	return UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
        History:      a.History.NullString,
		Slug_2:       a.Slug_2,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

///MAPS
func (d Database) MapRoute(a mdb.Routes) Routes {
	return Routes{
		RouteID:      a.RouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
        AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
        History:      a.History,
	}
}

func (d Database) MapCreateRouteParams(a CreateRouteParams) mdb.CreateRouteParams {
	return mdb.CreateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
        AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
        History:      a.History,
	}
}

func (d Database) MapUpdateRouteParams(a UpdateRouteParams) mdb.UpdateRouteParams {
	return mdb.UpdateRouteParams{
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
func (d Database) CountRoutes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateRouteTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateRouteTable(d.Context)
	return err
}

func (d Database) CreateRoute(s CreateRouteParams) Routes {
	params := d.MapCreateRouteParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRoute: %v\n", err)
	}
	return d.MapRoute(row)
}

func (d Database) DeleteRoute(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRoute(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", id)
	}
	return nil
}

func (d Database) GetRoute(id int64) (*Routes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRoute(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}

func (d Database) GetRouteID(slug string) (*int64, error) {
	queries := mdb.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (d Database) ListRoutes() (*[]Routes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateRoute(s UpdateRouteParams) (*string, error) {
	params := d.MapUpdateRouteParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

///MAPS
func (d MysqlDatabase) MapRoute(a mdbm.Routes) Routes {
	return Routes{
		RouteID:      int64(a.RouteID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
        AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated.String()),
		DateModified: StringToNullString(a.DateModified.String()),
        History:      a.History,
	}
}

func (d MysqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbm.CreateRouteParams {
	return mdbm.CreateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
        AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
        History:      a.History,
	}
}

func (d MysqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbm.UpdateRouteParams {
	return mdbm.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		History:      a.History,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		Slug_2:       a.Slug_2,
	}
}

///QUERIES
func (d MysqlDatabase) CountRoutes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateRouteTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateRouteTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateRoute(s CreateRouteParams) Routes {
	params := d.MapCreateRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRoute: %v\n", err)
	}
	row, err := queries.GetLastRoute(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Route: %v\n", err)
	}
	return d.MapRoute(row)
}

func (d MysqlDatabase) DeleteRoute(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRoute(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetRoute(id int64) (*Routes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRoute(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}

func (d MysqlDatabase) GetRouteID(slug string) (*int64, error) {
	queries := mdbm.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	i := int64(id)
	return &i, nil
}

func (d MysqlDatabase) ListRoutes() (*[]Routes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateRoute(s UpdateRouteParams) (*string, error) {
	params := d.MapUpdateRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

///MAPS
func (d PsqlDatabase) MapRoute(a mdbp.Routes) Routes {
	return Routes{
		RouteID:      int64(a.RouteID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
        AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated.String),
		DateModified: StringToNullString(a.DateModified.String),
        History:      a.History,
	}
}

func (d PsqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbp.CreateRouteParams {
	return mdbp.CreateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
        AuthorID:     int32(a.AuthorID),
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
        History:      a.History,
	}
}

func (d PsqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbp.UpdateRouteParams {
	return mdbp.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
        History:      a.History,
		Slug_2:       a.Slug_2,
	}
}

///QUERIES
func (d PsqlDatabase) CountRoutes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateRouteTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateRouteTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateRoute(s CreateRouteParams) Routes {
	params := d.MapCreateRouteParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRoute: %v\n", err)
	}
	return d.MapRoute(row)
}

func (d PsqlDatabase) DeleteRoute(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRoute(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetRoute(id int64) (*Routes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRoute(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}


func (d PsqlDatabase) GetRouteID(slug string) (*int64, error) {
	queries := mdbp.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	i := int64(id)
	return &i, nil
}

func (d PsqlDatabase) ListRoutes() (*[]Routes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateRoute(s UpdateRouteParams) (*string, error) {
	params := d.MapUpdateRouteParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}
