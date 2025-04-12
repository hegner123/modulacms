package db

import (
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// /////////////////////////////
// STRUCTS
// ////////////////////////////
type Tables struct {
	ID       int64  `json:"id"`
	Label    string `json:"label"`
	AuthorID int64  `json:"author_id"`
}

type CreateTableParams struct {
	Label    string `json:"label"`
	AuthorID int64  `json:"author_id"`
}

type UpdateTableParams struct {
	Label string `json:"label"`
	ID    int64  `json:"id"`
}

type TablesHistoryEntry struct {
	ID       int64  `json:"id"`
	Label    string `json:"label"`
	AuthorID int64  `json:"author_id"`
}

type CreateTableFormParams struct {
	Label    string `json:"label"`
	AuthorID string `json:"author_id"`
}

type UpdateTableFormParams struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateTableParams(a CreateTableFormParams) CreateTableParams {
	return CreateTableParams{
		Label:    a.Label,
		AuthorID: StringToInt64(a.AuthorID),
	}
}

func MapUpdateTableParams(a UpdateTableFormParams) UpdateTableParams {
	return UpdateTableParams{
		Label: a.Label,
		ID:    StringToInt64(a.ID),
	}
}

func MapStringTable(a Tables) StringTables {
	return StringTables{
		ID:       strconv.FormatInt(a.ID, 10),
		Label:    a.Label,
		AuthorID: strconv.FormatInt(a.AuthorID, 10),
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapTable(a mdb.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

func (d Database) MapUpdateTableParams(a UpdateTableParams) mdb.UpdateTableParams {
	return mdb.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// /QUERIES
func (d Database) CountTables() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateTableTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	if err != nil {
		return err
	}
	return nil
}

func (d Database) CreateTable(label string) Tables {
	queries := mdb.New(d.Connection)
	row, err := queries.CreateTable(d.Context, label)
	if err != nil {
		fmt.Printf("Failed to CreateTable: %v\n", err)
	}
	err = d.SortTables()
	if err != nil {
		fmt.Println("SORTING FAILED")
	}
	return d.MapTable(row)
}

func (d Database) DeleteTable(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteTable(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Table: %v ", id)
	}
	return nil
}

func (d Database) GetTable(id int64) (*Tables, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTable(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

func (d Database) ListTables() (*[]Tables, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListTable(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	res := []Tables{}
	for _, v := range rows {
		m := d.MapTable(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateTable(s UpdateTableParams) (*string, error) {
	params := d.MapUpdateTableParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateTable(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update table, %v", err)
	}
	u := fmt.Sprintf("Successfully updated table %v\n", s.ID)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapTable(a mdbm.Tables) Tables {
	return Tables{
		ID:       int64(a.ID),
		Label:    a.Label,
		AuthorID: int64(a.AuthorID),
	}
}

func (d MysqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbm.UpdateTableParams {
	return mdbm.UpdateTableParams{
		Label: a.Label,
		ID:    int32(a.ID),
	}
}

// /QUERIES
func (d MysqlDatabase) CountTables() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateTableTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateTable(label string) Tables {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTable(d.Context, label)
	if err != nil {
		fmt.Printf("Failed to CreateTable: %v\n", err)
	}
	row, err := queries.GetLastTable(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Table: %v\n", err)
	}
	err = d.SortTables()
	if err != nil {
		fmt.Println("SORTING FAILED")
	}
	return d.MapTable(row)
}

func (d MysqlDatabase) DeleteTable(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteTable(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Table: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetTable(id int64) (*Tables, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTable(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

func (d MysqlDatabase) ListTables() (*[]Tables, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListTable(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	res := []Tables{}
	for _, v := range rows {
		m := d.MapTable(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateTable(s UpdateTableParams) (*string, error) {
	params := d.MapUpdateTableParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateTable(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update table, %v", err)
	}
	u := fmt.Sprintf("Successfully updated table %v\n", s.ID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapTable(a mdbp.Tables) Tables {
	return Tables{
		ID:       int64(a.ID),
		Label:    a.Label,
		AuthorID: int64(a.AuthorID),
	}
}

func (d PsqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbp.UpdateTableParams {
	return mdbp.UpdateTableParams{
		Label: a.Label,
		ID:    int32(a.ID),
	}
}

// /QUERIES
func (d PsqlDatabase) CountTables() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateTableTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateTable(label string) Tables {
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateTable(d.Context, label)
	if err != nil {
		fmt.Printf("Failed to CreateTable: %v\n", err)
	}
	err = d.SortTables()
	if err != nil {
		fmt.Println("SORTING FAILED")
	}
	return d.MapTable(row)
}

func (d PsqlDatabase) DeleteTable(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteTable(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Table: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetTable(id int64) (*Tables, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetTable(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapTable(row)
	return &res, nil
}

func (d PsqlDatabase) ListTables() (*[]Tables, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListTable(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	res := []Tables{}
	for _, v := range rows {
		m := d.MapTable(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateTable(s UpdateTableParams) (*string, error) {
	params := d.MapUpdateTableParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateTable(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update table, %v", err)
	}
	u := fmt.Sprintf("Successfully updated table %v\n", s.ID)
	return &u, nil
}
