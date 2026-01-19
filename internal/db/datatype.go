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

type Datatypes struct {
	DatatypeID   int64          `json:"datatype_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type CreateDatatypeParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type UpdateDatatypeParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
	DatatypeID   int64          `json:"datatype_id"`
}

type DatatypesHistoryEntry struct {
	DatatypeID   int64          `json:"datatype_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateDatatypeFormParams struct {
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type UpdateDatatypeFormParams struct {
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
	DatatypeID   string `json:"datatype_id"`
}

type DatatypeJSON struct {
	DatatypeID   int64      `json:"datatype_id"`
	ParentID     NullInt64  `json:"parent_id"`
	Label        string     `json:"label"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
}

type CreateDatatypeParamsJSON struct {
	ParentID     NullInt64  `json:"parent_id"`
	Label        string     `json:"label"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
}

type UpdateDatatypeParamsJSON struct {
	ParentID     NullInt64  `json:"parent_id"`
	Label        string     `json:"label"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
	DatatypeID   int64      `json:"datatype_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////


func MapCreateDatatypeParams(a CreateDatatypeFormParams) CreateDatatypeParams {
	return CreateDatatypeParams{
		ParentID:     StringToNullInt64(a.ParentID),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		History:      StringToNullString(a.History),
	}
}

func MapUpdateDatatypeParams(a UpdateDatatypeFormParams) UpdateDatatypeParams {
	return UpdateDatatypeParams{
		ParentID:     StringToNullInt64(a.ParentID),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		History:      StringToNullString(a.History),
		DatatypeID:   StringToInt64(a.DatatypeID),
	}
}

func MapStringDatatype(a Datatypes) StringDatatypes {
	return StringDatatypes{
		DatatypeID:   strconv.FormatInt(a.DatatypeID, 10),
		ParentID:     utility.NullToString(a.ParentID),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  utility.NullToString(a.DateCreated),
		DateModified: utility.NullToString(a.DateModified),
		History:      utility.NullToString(a.History),
	}
}

func MapJSONDatatypes(a DatatypeJSON)Datatypes{
    return Datatypes{
        DatatypeID: a.DatatypeID,
        ParentID: a.ParentID.NullInt64,
        Label: a.Label,
        Type: a.Type,
        AuthorID: a.AuthorID,
        DateCreated: a.DateCreated.NullString,
        DateModified: a.DateModified.NullString,
        History: a.History.NullString,
    }
}

func MapDatatypeJSON(a Datatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID: a.DatatypeID,
		ParentID: NullInt64{
			NullInt64: a.ParentID,
		},
        Label: a.Label,
        Type: a.Type,
        AuthorID: a.AuthorID,
        DateCreated: NullString{
            NullString: a.DateCreated,
        },
        DateModified: NullString{
            NullString: a.DateModified,
        },
        History: NullString{
            NullString: a.History,
        },
	}

}
func MapCreateDatatypeJSONParams(a CreateDatatypeParamsJSON)CreateDatatypeParams{
    return CreateDatatypeParams{
        ParentID: a.ParentID.NullInt64,
        Label: a.Label,
        Type: a.Type,
        AuthorID: a.AuthorID,
        DateCreated: a.DateCreated.NullString,
        DateModified: a.DateModified.NullString,
        History: a.History.NullString,
    }
}

func MapUpdateDatatypeJSONParams(a UpdateDatatypeParamsJSON)UpdateDatatypeParams{
    return UpdateDatatypeParams{
        ParentID: a.ParentID.NullInt64,
        Label: a.Label,
        Type: a.Type,
        AuthorID: a.AuthorID,
        DateCreated: a.DateCreated.NullString,
        DateModified: a.DateModified.NullString,
        History: a.History.NullString,
        DatatypeID: a.DatatypeID,
    }
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapDatatype(a mdb.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}

func (d Database) MapCreateDatatypeParams(a CreateDatatypeParams) mdb.CreateDatatypeParams {
	return mdb.CreateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}

func (d Database) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdb.UpdateDatatypeParams {
	return mdb.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
		DatatypeID:   a.DatatypeID,
	}
}

// /QUERIES
func (d Database) CountDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d Database) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype: %v\n", err)
	}
	return d.MapDatatype(row)
}

func (d Database) DeleteDatatype(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Datatype: %v ", id)
	}
	return nil
}

func (d Database) GetDatatype(id int64) (*Datatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d Database) ListDatatypes() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapDatatype(a mdbm.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   int64(a.DatatypeID),
		ParentID:     Int64ToNullInt64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated.String()),
		DateModified: StringToNullString(a.DateModified.String()),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbm.CreateDatatypeParams {
	return mdbm.CreateDatatypeParams{
		ParentID: Int64ToNullInt32(a.ParentID.Int64),
		Label:    a.Label,
		Type:     a.Type,
		AuthorID: int32(a.AuthorID),
		History:  a.History,
	}
}

func (d MysqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbm.UpdateDatatypeParams {
	return mdbm.UpdateDatatypeParams{
		ParentID:   Int64ToNullInt32(a.ParentID.Int64),
		Label:      a.Label,
		Type:       a.Type,
		AuthorID:   int32(a.AuthorID),
		History:    a.History,
		DatatypeID: int32(a.DatatypeID),
	}
}

// /QUERIES
func (d MysqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype: %v\n", err)
	}
	row, err := queries.GetLastDatatype(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Datatype: %v\n", err)
	}
	return d.MapDatatype(row)
}

func (d MysqlDatabase) DeleteDatatype(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Datatype: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetDatatype(id int64) (*Datatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d MysqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}


func (d MysqlDatabase) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapDatatype(a mdbp.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   int64(a.DatatypeID),
		ParentID:     Int64ToNullInt64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified: StringToNullString(NullTimeToString(a.DateModified)),
		History:      a.History,
	}
}

func (d PsqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbp.CreateDatatypeParams {
	return mdbp.CreateDatatypeParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
	}
}

func (d PsqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbp.UpdateDatatypeParams {
	return mdbp.UpdateDatatypeParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
		DatatypeID:   int32(a.DatatypeID),
	}
}

// /QUERIES
func (d PsqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype: %v\n", err)
	}
	return d.MapDatatype(row)
}

func (d PsqlDatabase) DeleteDatatype(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Datatype: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetDatatype(id int64) (*Datatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}


func (d PsqlDatabase) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
