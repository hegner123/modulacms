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

type Tables struct {
	ID       string               `json:"id"`
	Label    string               `json:"label"`
	AuthorID types.NullableUserID `json:"author_id"`
}

type CreateTableParams struct {
	Label string `json:"label"`
}

type UpdateTableParams struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

// FormParams and HistoryEntry variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringTable converts Tables to StringTables for table display
func MapStringTable(a Tables) StringTables {
	return StringTables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapTable(a mdb.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

func (d Database) MapCreateTableParams(a CreateTableParams) mdb.CreateTableParams {
	return mdb.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

func (d Database) MapUpdateTableParams(a UpdateTableParams) mdb.UpdateTableParams {
	return mdb.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

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
	return err
}

func (d Database) CreateTable(s CreateTableParams) Tables {
	params := d.MapCreateTableParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateTable(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateTable: %v\n", err)
	}
	return d.MapTable(row)
}

func (d Database) DeleteTable(id string) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteTable(d.Context, mdb.DeleteTableParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete table: %v", id)
	}
	return nil
}

func (d Database) GetTable(id string) (*Tables, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdb.GetTableParams{ID: id})
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
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapTable(a mdbm.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

func (d MysqlDatabase) MapCreateTableParams(a CreateTableParams) mdbm.CreateTableParams {
	return mdbm.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

func (d MysqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbm.UpdateTableParams {
	return mdbm.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

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

func (d MysqlDatabase) CreateTable(s CreateTableParams) Tables {
	params := d.MapCreateTableParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateTable(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateTable: %v\n", err)
	}
	row, err := queries.GetLastTable(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Table: %v\n", err)
	}
	return d.MapTable(row)
}

func (d MysqlDatabase) DeleteTable(id string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteTable(d.Context, mdbm.DeleteTableParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete table: %v", id)
	}
	return nil
}

func (d MysqlDatabase) GetTable(id string) (*Tables, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdbm.GetTableParams{ID: id})
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
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapTable(a mdbp.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}

func (d PsqlDatabase) MapCreateTableParams(a CreateTableParams) mdbp.CreateTableParams {
	return mdbp.CreateTableParams{
		ID:    string(types.NewTableID()),
		Label: a.Label,
	}
}

func (d PsqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbp.UpdateTableParams {
	return mdbp.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

// QUERIES

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

func (d PsqlDatabase) CreateTable(s CreateTableParams) Tables {
	params := d.MapCreateTableParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateTable(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateTable: %v\n", err)
	}
	return d.MapTable(row)
}

func (d PsqlDatabase) DeleteTable(id string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteTable(d.Context, mdbp.DeleteTableParams{ID: id})
	if err != nil {
		return fmt.Errorf("failed to delete table: %v", id)
	}
	return nil
}

func (d PsqlDatabase) GetTable(id string) (*Tables, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetTable(d.Context, mdbp.GetTableParams{ID: id})
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
