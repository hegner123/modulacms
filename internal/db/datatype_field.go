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

type DatatypeFields struct {
	ID         int64 `json:"id"`
	DatatypeID int64 `json:"datatype_id"`
	FieldID    int64 `json:"field_id"`
}

type CreateDatatypeFieldParams struct {
	DatatypeID int64 `json:"datatype_id"`
	FieldID    int64 `json:"field_id"`
}

type UpdateDatatypeFieldParams struct {
	DatatypeID int64 `json:"datatype_id"`
	FieldID    int64 `json:"field_id"`
	ID         int64 `json:"id"`
}

type DatatypeFieldHistoryEntry struct {
	DatatypeID int64 `json:"datatype_id"`
	FieldID    int64 `json:"field_id"`
}

type CreateDatatypeFieldFormParams struct {
	DatatypeID string `json:"datatype_id"`
	FieldID    string `json:"field_id"`
}

type UpdateDatatypeFieldFormParams struct {
	DatatypeID string `json:"datatype_id"`
	FieldID    string `json:"field_id"`
	ID         string `json:"id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateDatatypeFieldParams(a CreateDatatypeFieldFormParams) CreateDatatypeFieldParams {
	return CreateDatatypeFieldParams{
		DatatypeID: StringToInt64(a.DatatypeID),
		FieldID:    StringToInt64(a.FieldID),
	}
}

func MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldFormParams) UpdateDatatypeFieldParams {
	return UpdateDatatypeFieldParams{
		DatatypeID: StringToInt64(a.DatatypeID),
		FieldID:    StringToInt64(a.FieldID),
	}
}

func MapStringDatatypeField(a DatatypeFields) StringDatatypeFields {
	return StringDatatypeFields{
		ID:         strconv.FormatInt(a.ID, 10),
		DatatypeID: strconv.FormatInt(a.DatatypeID, 10),
		FieldID:    strconv.FormatInt(a.FieldID, 10),
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapDatatypeField(a mdb.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         a.ID,
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
	}
}

func (d Database) MapCreateDatatypeFieldParams(a CreateDatatypeFieldParams) mdb.CreateDatatypeFieldParams {
	return mdb.CreateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
	}
}

func (d Database) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdb.UpdateDatatypeFieldParams {
	return mdb.UpdateDatatypeFieldParams{
		DatatypeID: a.DatatypeID,
		FieldID:    a.FieldID,
		ID:         a.ID,
	}
}

// /QUERIES
func (d Database) CountDatatypeFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateDatatypeFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateDatatypesFieldsTable(d.Context)
	return err
}

func (d Database) CreateDatatypeField(s CreateDatatypeFieldParams) DatatypeFields {
	params := d.MapCreateDatatypeFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatypeField: %v\n", err)
	}
	return d.MapDatatypeField(row)
}

func (d Database) DeleteDatatypeField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteDatatypeField(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete DatatypeField: %v ", id)
	}
	return nil
}

func (d Database) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListDatatypeFieldByDatatypeID(id int64) (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByDatatypeID(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListDatatypeFieldByFieldID(id int64) (*[]DatatypeFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByFieldID(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateDatatypeField(s UpdateDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateDatatypeFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapDatatypeField(a mdbm.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         int64(a.ID),
		DatatypeID: int64(a.DatatypeID),
		FieldID:    int64(a.FieldID),
	}
}

func (d MysqlDatabase) MapCreateDatatypeFieldParams(a CreateDatatypeFieldParams) mdbm.CreateDatatypeFieldParams {
	return mdbm.CreateDatatypeFieldParams{

		DatatypeID: int32(a.DatatypeID),
		FieldID:    int32(a.FieldID),
	}
}

func (d MysqlDatabase) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdbm.UpdateDatatypeFieldParams {
	return mdbm.UpdateDatatypeFieldParams{
		DatatypeID: int32(a.DatatypeID),
		FieldID:    int32(a.FieldID),
		ID:         int32(a.ID),
	}
}

// /QUERIES
func (d MysqlDatabase) CountDatatypeFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateDatatypeFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypesFieldsTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateDatatypeField(s CreateDatatypeFieldParams) DatatypeFields {
	params := d.MapCreateDatatypeFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatypeField: %v\n", err)
	}
	row, err := queries.GetLastDatatypeField(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted DatatypeField: %v\n", err)
	}
	return d.MapDatatypeField(row)
}

func (d MysqlDatabase) DeleteDatatypeField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteDatatypeField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete DatatypeField: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListDatatypeFieldByFieldID(id int64) (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByFieldID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListDatatypeFieldByDatatypeID(id int64) (*[]DatatypeFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByDatatypeID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateDatatypeField(s UpdateDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateDatatypeFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapDatatypeField(a mdbp.DatatypesFields) DatatypeFields {
	return DatatypeFields{
		ID:         int64(a.ID),
		DatatypeID: int64(a.DatatypeID),
		FieldID:    int64(a.FieldID),
	}
}

func (d PsqlDatabase) MapCreateDatatypeFieldParams(a CreateDatatypeFieldParams) mdbp.CreateDatatypeFieldParams {
	return mdbp.CreateDatatypeFieldParams{

		DatatypeID: int32(a.DatatypeID),
		FieldID:    int32(a.FieldID),
	}
}

func (d PsqlDatabase) MapUpdateDatatypeFieldParams(a UpdateDatatypeFieldParams) mdbp.UpdateDatatypeFieldParams {
	return mdbp.UpdateDatatypeFieldParams{
		DatatypeID: int32(a.DatatypeID),
		FieldID:    int32(a.FieldID),
		ID:         int32(a.ID),
	}
}

// /QUERIES
func (d PsqlDatabase) CountDatatypeFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateDatatypeFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateDatatypesFieldsTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateDatatypeField(s CreateDatatypeFieldParams) DatatypeFields {
	params := d.MapCreateDatatypeFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateDatatypeField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatypeField: %v\n", err)
	}
	return d.MapDatatypeField(row)
}

func (d PsqlDatabase) DeleteDatatypeField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteDatatypeField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete DatatypeField: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) ListDatatypeField() (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListDatatypeFieldByDatatypeID(id int64) (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByDatatypeID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListDatatypeFieldByFieldID(id int64) (*[]DatatypeFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeFieldByFieldID(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get DatatypeFields: %v\n", err)
	}
	res := []DatatypeFields{}
	for _, v := range rows {
		m := d.MapDatatypeField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateDatatypeField(s UpdateDatatypeFieldParams) (*string, error) {
	params := d.MapUpdateDatatypeFieldParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateDatatypeField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}
