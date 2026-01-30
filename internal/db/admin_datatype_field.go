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

type AdminDatatypeFields struct {
	ID              string                        `json:"id"`
	AdminDatatypeID types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AdminFieldID    types.NullableAdminFieldID    `json:"admin_field_id"`
}

type CreateAdminDatatypeFieldParams struct {
	AdminDatatypeID types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AdminFieldID    types.NullableAdminFieldID    `json:"admin_field_id"`
}

type UpdateAdminDatatypeFieldParams struct {
	AdminDatatypeID types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AdminFieldID    types.NullableAdminFieldID    `json:"admin_field_id"`
	ID              string                        `json:"id"`
}

// MapStringAdminDatatypeField converts AdminDatatypeFields to StringAdminDatatypeFields for display purposes
func MapStringAdminDatatypeField(a AdminDatatypeFields) StringAdminDatatypeFields {
	return StringAdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID.String(),
		AdminFieldID:    a.AdminFieldID.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminDatatypeField(a mdb.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d Database) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdb.CreateAdminDatatypeFieldParams {
	return mdb.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d Database) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdb.UpdateAdminDatatypeFieldParams {
	return mdb.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
		ID:              a.ID,
	}
}

// QUERIES

func (d Database) CountAdminDatatypeFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminDatatypeFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminDatatypesFieldsTable(d.Context)
	return err
}

func (d Database) CreateAdminDatatypeField(s CreateAdminDatatypeFieldParams) AdminDatatypeFields {
	params := d.MapCreateAdminDatatypeFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatypeField: %v\n", err)
	}
	return d.MapAdminDatatypeField(row)
}

func (d Database) DeleteAdminDatatypeField(id string) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminDatatypeField(d.Context, mdb.DeleteAdminDatatypeFieldParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete AdminDatatypeField: %v", id)
	}
	return nil
}

func (d Database) GetAdminDatatypeField(id string) (*AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeField: %v", err)
	}
	for _, v := range rows {
		if v.ID == id {
			m := d.MapAdminDatatypeField(v)
			return &m, nil
		}
	}
	return nil, fmt.Errorf("AdminDatatypeField not found: %v", id)
}

func (d Database) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypeFieldByDatatypeID(id types.NullableAdminDatatypeID) (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, mdb.ListAdminDatatypeFieldByDatatypeIDParams{AdminDatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypeFieldByFieldID(id types.NullableAdminFieldID) (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, mdb.ListAdminDatatypeFieldByFieldIDParams{AdminFieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminDatatypeField(s UpdateAdminDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminDatatypeField(a mdbm.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d MysqlDatabase) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdbm.CreateAdminDatatypeFieldParams {
	return mdbm.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d MysqlDatabase) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdbm.UpdateAdminDatatypeFieldParams {
	return mdbm.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
		ID:              a.ID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminDatatypeFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminDatatypeFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatypesFieldsTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateAdminDatatypeField(s CreateAdminDatatypeFieldParams) AdminDatatypeFields {
	params := d.MapCreateAdminDatatypeFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatypeField: %v\n", err)
	}
	row, err := queries.GetLastAdminDatatypeField(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminDatatypeField: %v\n", err)
	}
	return d.MapAdminDatatypeField(row)
}

func (d MysqlDatabase) DeleteAdminDatatypeField(id string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminDatatypeField(d.Context, mdbm.DeleteAdminDatatypeFieldParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete AdminDatatypeField: %v", id)
	}
	return nil
}

func (d MysqlDatabase) GetAdminDatatypeField(id string) (*AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeField: %v", err)
	}
	for _, v := range rows {
		if v.ID == id {
			m := d.MapAdminDatatypeField(v)
			return &m, nil
		}
	}
	return nil, fmt.Errorf("AdminDatatypeField not found: %v", id)
}

func (d MysqlDatabase) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypeFieldByFieldID(id types.NullableAdminFieldID) (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, mdbm.ListAdminDatatypeFieldByFieldIDParams{AdminFieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypeFieldByDatatypeID(id types.NullableAdminDatatypeID) (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, mdbm.ListAdminDatatypeFieldByDatatypeIDParams{AdminDatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateAdminDatatypeField(s UpdateAdminDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapAdminDatatypeField(a mdbp.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              a.ID,
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d PsqlDatabase) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdbp.CreateAdminDatatypeFieldParams {
	return mdbp.CreateAdminDatatypeFieldParams{
		ID:              string(types.NewAdminDatatypeFieldID()),
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
	}
}

func (d PsqlDatabase) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdbp.UpdateAdminDatatypeFieldParams {
	return mdbp.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: a.AdminDatatypeID,
		AdminFieldID:    a.AdminFieldID,
		ID:              a.ID,
	}
}

// QUERIES

func (d PsqlDatabase) CountAdminDatatypeFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminDatatypeFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminDatatypesFieldsTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateAdminDatatypeField(s CreateAdminDatatypeFieldParams) AdminDatatypeFields {
	params := d.MapCreateAdminDatatypeFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatypeField: %v\n", err)
	}
	return d.MapAdminDatatypeField(row)
}

func (d PsqlDatabase) DeleteAdminDatatypeField(id string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminDatatypeField(d.Context, mdbp.DeleteAdminDatatypeFieldParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete AdminDatatypeField: %v", id)
	}
	return nil
}

func (d PsqlDatabase) GetAdminDatatypeField(id string) (*AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeField: %v", err)
	}
	for _, v := range rows {
		if v.ID == id {
			m := d.MapAdminDatatypeField(v)
			return &m, nil
		}
	}
	return nil, fmt.Errorf("AdminDatatypeField not found: %v", id)
}

func (d PsqlDatabase) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypeFieldByDatatypeID(id types.NullableAdminDatatypeID) (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, mdbp.ListAdminDatatypeFieldByDatatypeIDParams{AdminDatatypeID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypeFieldByFieldID(id types.NullableAdminFieldID) (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, mdbp.ListAdminDatatypeFieldByFieldIDParams{AdminFieldID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminDatatypeField(s UpdateAdminDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeFieldParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}
