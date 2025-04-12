package db

import (
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

///////////////////////////////
//STRUCTS
//////////////////////////////

type AdminDatatypeFields struct {
	ID              int64 `json:"id"`
	AdminDatatypeID int64 `json:"admin_datatype_id"`
	AdminFieldID    int64 `json:"admin_field_id"`
}

type CreateAdminDatatypeFieldParams struct {
	AdminDatatypeID int64 `json:"admin_datatype_id"`
	AdminFieldID    int64 `json:"admin_field_id"`
}

type UpdateAdminDatatypeFieldParams struct {
	AdminDatatypeID int64 `json:"admin_datatype_id"`
	AdminFieldID    int64 `json:"admin_field_id"`
	ID              int64 `json:"id"`
}

type AdminDatatypeFieldHistoryEntry struct {
	AdminDatatypeID int64 `json:"admin_datatype_id"`
	AdminFieldID    int64 `json:"admin_field_id"`
}

type CreateAdminDatatypeFieldFormParams struct {
	AdminDatatypeID string `json:"admin_datatype_id"`
	AdminFieldID    string `json:"admin_field_id"`
}

type UpdateAdminDatatypeFieldFormParams struct {
	AdminDatatypeID string `json:"admin_datatype_id"`
	AdminFieldID    string `json:"admin_field_id"`
	ID              string `json:"id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldFormParams) CreateAdminDatatypeFieldParams {
	return CreateAdminDatatypeFieldParams{
		AdminDatatypeID: StringToInt64(a.AdminDatatypeID),
		AdminFieldID:    StringToInt64(a.AdminFieldID),
	}
}

func MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldFormParams) UpdateAdminDatatypeFieldParams {
	return UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: StringToInt64(a.AdminDatatypeID),
		AdminFieldID:    StringToInt64(a.AdminFieldID),
	}
}

func MapStringAdminDatatypeField(a AdminDatatypeFields) StringAdminDatatypeFields {
	return StringAdminDatatypeFields{
		ID:              strconv.FormatInt(a.ID, 10),
		AdminDatatypeID: strconv.FormatInt(a.AdminDatatypeID, 10),
		AdminFieldID:    strconv.FormatInt(a.AdminFieldID, 10),
	}
}

///////////////////////////////
//SQLITE
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

func (d Database) DeleteAdminDatatypeField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminDatatypeField(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete AdminDatatypeField: %v ", id)
	}
	return nil
}

func (d Database) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListAdminDatatypeFieldByDatatypeID(id int64) (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypeFieldByFieldID(id int64) (*[]AdminDatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
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
//MYSQL
//////////////////////////////

// MAPS
func (d MysqlDatabase) MapAdminDatatypeField(a mdbm.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              int64(a.ID),
		AdminDatatypeID: int64(a.AdminDatatypeID),
		AdminFieldID:    int64(a.AdminFieldID),
	}
}

func (d MysqlDatabase) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdbm.CreateAdminDatatypeFieldParams {
	return mdbm.CreateAdminDatatypeFieldParams{

		AdminDatatypeID: int32(a.AdminDatatypeID),
		AdminFieldID:    int32(a.AdminFieldID),
	}
}

func (d MysqlDatabase) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdbm.UpdateAdminDatatypeFieldParams {
	return mdbm.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: int32(a.AdminDatatypeID),
		AdminFieldID:    int32(a.AdminFieldID),
		ID:              int32(a.ID),
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

func (d MysqlDatabase) DeleteAdminDatatypeField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminDatatypeField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete AdminDatatypeField: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypeFieldByFieldID(id int64) (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListAdminDatatypeFieldByDatatypeID(id int64) (*[]AdminDatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
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
//POSTGRES
//////////////////////////////

// MAPS
func (d PsqlDatabase) MapAdminDatatypeField(a mdbp.AdminDatatypesFields) AdminDatatypeFields {
	return AdminDatatypeFields{
		ID:              int64(a.ID),
		AdminDatatypeID: int64(a.AdminDatatypeID),
		AdminFieldID:    int64(a.AdminFieldID),
	}
}

func (d PsqlDatabase) MapCreateAdminDatatypeFieldParams(a CreateAdminDatatypeFieldParams) mdbp.CreateAdminDatatypeFieldParams {
	return mdbp.CreateAdminDatatypeFieldParams{

		AdminDatatypeID: int32(a.AdminDatatypeID),
		AdminFieldID:    int32(a.AdminFieldID),
	}
}

func (d PsqlDatabase) MapUpdateAdminDatatypeFieldParams(a UpdateAdminDatatypeFieldParams) mdbp.UpdateAdminDatatypeFieldParams {
	return mdbp.UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: int32(a.AdminDatatypeID),
		AdminFieldID:    int32(a.AdminFieldID),
		ID:              int32(a.ID),
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

func (d PsqlDatabase) DeleteAdminDatatypeField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminDatatypeField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete AdminDatatypeField: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) ListAdminDatatypeField() (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypeFieldByDatatypeID(id int64) (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByDatatypeID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
	}
	res := []AdminDatatypeFields{}
	for _, v := range rows {
		m := d.MapAdminDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypeFieldByFieldID(id int64) (*[]AdminDatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeFieldByFieldID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypeFields: %v\n", err)
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
