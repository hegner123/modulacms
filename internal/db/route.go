package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
//STRUCTS
//////////////////////////////
type Routes struct {
	RouteID      types.RouteID        `json:"route_id"`
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}

type CreateRouteParams struct {
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}

type UpdateRouteParams struct {
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
	Slug_2       types.Slug           `json:"slug_2"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - use typed params directly

// MapStringRoute converts Routes to StringRoutes for table display
func MapStringRoute(a Routes) StringRoutes {
	return StringRoutes{
		RouteID:      a.RouteID.String(),
		Slug:         string(a.Slug),
		Title:        a.Title,
		Status:       fmt.Sprintf("%d", a.Status),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
		History:      "", // History field removed
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
	}
}

func (d Database) MapCreateRouteParams(a CreateRouteParams) mdb.CreateRouteParams {
	return mdb.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
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

func (d Database) DeleteRoute(id types.RouteID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRoute(d.Context, mdb.DeleteRouteParams{RouteID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", id)
	}
	return nil
}

func (d Database) GetRoute(id types.RouteID) (*Routes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRoute(d.Context, mdb.GetRouteParams{RouteID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}

func (d Database) GetRouteID(slug string) (*types.RouteID, error) {
	queries := mdb.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, mdb.GetRouteIDBySlugParams{Slug: types.Slug(slug)})
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
		RouteID:      a.RouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbm.CreateRouteParams {
	return mdbm.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbm.UpdateRouteParams {
	return mdbm.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
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

func (d MysqlDatabase) DeleteRoute(id types.RouteID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRoute(d.Context, mdbm.DeleteRouteParams{RouteID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetRoute(id types.RouteID) (*Routes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRoute(d.Context, mdbm.GetRouteParams{RouteID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}

func (d MysqlDatabase) GetRouteID(slug string) (*types.RouteID, error) {
	queries := mdbm.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, mdbm.GetRouteIDBySlugParams{Slug: types.Slug(slug)})
	if err != nil {
		return nil, err
	}
	return &id, nil
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
		RouteID:      a.RouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbp.CreateRouteParams {
	return mdbp.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbp.UpdateRouteParams {
	return mdbp.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
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

func (d PsqlDatabase) DeleteRoute(id types.RouteID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRoute(d.Context, mdbp.DeleteRouteParams{RouteID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetRoute(id types.RouteID) (*Routes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRoute(d.Context, mdbp.GetRouteParams{RouteID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}

func (d PsqlDatabase) GetRouteID(slug string) (*types.RouteID, error) {
	queries := mdbp.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, mdbp.GetRouteIDBySlugParams{Slug: types.Slug(slug)})
	if err != nil {
		return nil, err
	}
	return &id, nil
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
