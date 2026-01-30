package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type AdminRoutes struct {
	AdminRouteID types.AdminRouteID   `json:"admin_route_id"`
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}
type CreateAdminRouteParams struct {
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}
type UpdateAdminRouteParams struct {
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
	Slug_2       types.Slug           `json:"slug_2"`
}
type UtilityGetAdminRoutesRow struct {
	AdminRouteID types.AdminRouteID `json:"admin_route_id"`
	Slug         types.Slug         `json:"slug"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringAdminRoute converts AdminRoutes to StringAdminRoutes for table display
func MapStringAdminRoute(a AdminRoutes) StringAdminRoutes {
	return StringAdminRoutes{
		AdminRouteID: a.AdminRouteID.String(),
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
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminRoute(a mdb.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdb.CreateAdminRouteParams {
	return mdb.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
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
		Slug_2:       a.Slug_2,
	}
}

// QUERIES

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

func (d Database) DeleteAdminRoute(id types.AdminRouteID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, mdb.DeleteAdminRouteParams{AdminRouteID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin route: %v ", id)
	}

	return nil
}

func (d Database) GetAdminRoute(slug types.Slug) (*AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, mdb.GetAdminRouteBySlugParams{Slug: slug})
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
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminRoute(a mdbm.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbm.CreateAdminRouteParams {
	return mdbm.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbm.UpdateAdminRouteParams {
	return mdbm.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

// QUERIES

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

func (d MysqlDatabase) DeleteAdminRoute(id types.AdminRouteID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, mdbm.DeleteAdminRouteParams{AdminRouteID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin route: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) GetAdminRoute(slug types.Slug) (*AdminRoutes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, mdbm.GetAdminRouteBySlugParams{Slug: slug})
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
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapAdminRoute(a mdbp.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbp.CreateAdminRouteParams {
	return mdbp.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbp.UpdateAdminRouteParams {
	return mdbp.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

// QUERIES

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

func (d PsqlDatabase) DeleteAdminRoute(id types.AdminRouteID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, mdbp.DeleteAdminRouteParams{AdminRouteID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin route: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) GetAdminRoute(slug types.Slug) (*AdminRoutes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, mdbp.GetAdminRouteBySlugParams{Slug: slug})
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
