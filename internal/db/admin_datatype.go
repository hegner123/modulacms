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

type AdminDatatypes struct {
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
	ParentID        types.NullableContentID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
	AuthorID        types.NullableUserID    `json:"author_id"`
	DateCreated     types.Timestamp         `json:"date_created"`
	DateModified    types.Timestamp         `json:"date_modified"`
}
type CreateAdminDatatypeParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}
type ListAdminDatatypeByRouteIdRow struct {
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
	AdminRouteID    types.NullableRouteID   `json:"admin_route_id"`
	ParentID        types.NullableContentID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
}
type UpdateAdminDatatypeParams struct {
	ParentID        types.NullableContentID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
	AuthorID        types.NullableUserID    `json:"author_id"`
	DateCreated     types.Timestamp         `json:"date_created"`
	DateModified    types.Timestamp         `json:"date_modified"`
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
}
type UtilityGetAdminDatatypesRow struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	Label           string                `json:"label"`
}

// FormParams and JSON variants removed - use typed params directly

// MapAdminDatatypeJSON converts AdminDatatypes to DatatypeJSON for tree building.
// Maps admin datatype ID into the public DatatypeJSON shape so BuildNodes works unchanged.
func MapAdminDatatypeJSON(a AdminDatatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID:   a.AdminDatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringAdminDatatype converts AdminDatatypes to StringAdminDatatypes for table display
func MapStringAdminDatatype(a AdminDatatypes) StringAdminDatatypes {
	return StringAdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID.String(),
		ParentID:        a.ParentID.String(),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID.String(),
		DateCreated:     a.DateCreated.String(),
		DateModified:    a.DateModified.String(),
		History:         "", // History field removed from AdminDatatypes
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminDatatype(a mdb.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

func (d Database) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdb.CreateAdminDatatypeParams {
	return mdb.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdb.UpdateAdminDatatypeParams {
	return mdb.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

func (d Database) CountAdminDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
	params := d.MapCreateAdminDatatypeParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminDatatype(row)
}
func (d Database) CreateAdminDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d Database) DeleteAdminDatatype(id types.AdminDatatypeID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, mdb.DeleteAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin datatype: %v ", id)
	}

	return nil
}

func (d Database) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdb.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d Database) ListAdminDatatypeGlobalId() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeGlobal(d.Context)
	if err != nil {
		return nil, err
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminDatatype(s UpdateAdminDatatypeParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminDatatype(a mdbm.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbm.CreateAdminDatatypeParams {
	return mdbm.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbm.UpdateAdminDatatypeParams {
	return mdbm.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
	params := d.MapCreateAdminDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype: %v\n", err)
	}
	row, err := queries.GetAdminDatatype(d.Context, mdbm.GetAdminDatatypeParams{AdminDatatypeID: params.AdminDatatypeID})
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminDatatype: %v\n", err)
	}
	return d.MapAdminDatatype(row)
}

func (d MysqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d MysqlDatabase) DeleteAdminDatatype(id types.AdminDatatypeID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, mdbm.DeleteAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin datatype: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdbm.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateAdminDatatype(s UpdateAdminDatatypeParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapAdminDatatype(a mdbp.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbp.CreateAdminDatatypeParams {
	return mdbp.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbp.UpdateAdminDatatypeParams {
	return mdbp.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

func (d PsqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
	params := d.MapCreateAdminDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminDatatype(row)
}

func (d PsqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d PsqlDatabase) DeleteAdminDatatype(id types.AdminDatatypeID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, mdbp.DeleteAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin datatype: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdbp.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminDatatype(s UpdateAdminDatatypeParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
