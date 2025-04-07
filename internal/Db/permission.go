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
type Permissions struct {
	PermissionID int64  `json:"permission_id"`
	TableID      int64  `json:"table_id"`
	Mode         int64  `json:"mode"`
	Label        string `json:"label"`
}

type CreatePermissionParams struct {
	TableID int64  `json:"table_id"`
	Mode    int64  `json:"mode"`
	Label   string `json:"label"`
}

type UpdatePermissionParams struct {
	TableID      int64  `json:"table_id"`
	Mode         int64  `json:"mode"`
	Label        string `json:"label"`
	PermissionID int64  `json:"permission_id"`
}

type PermissionsHistoryEntry struct {
	PermissionID int64  `json:"permission_id"`
	TableID      int64  `json:"table_id"`
	Mode         int64  `json:"mode"`
	Label        string `json:"label"`
}

type CreatePermissionFormParams struct {
	TableID string `json:"table_id"`
	Mode    string `json:"mode"`
	Label   string `json:"label"`
}

type UpdatePermissionFormParams struct {
	TableID      string `json:"table_id"`
	Mode         string `json:"mode"`
	Label        string `json:"label"`
	PermissionID string `json:"permission_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreatePermissionParams(a CreatePermissionFormParams) CreatePermissionParams {
	return CreatePermissionParams{
		TableID: StringToInt64(a.TableID),
		Mode:    StringToInt64(a.Mode),
		Label:   a.Label,
	}
}

func MapUpdatePermissionParams(a UpdatePermissionFormParams) UpdatePermissionParams {
	return UpdatePermissionParams{
		TableID:      StringToInt64(a.TableID),
		Mode:         StringToInt64(a.Mode),
		Label:        a.Label,
		PermissionID: StringToInt64(a.PermissionID),
	}
}

func MapStringPermission(a Permissions) StringPermissions {
	return StringPermissions{
		PermissionID: strconv.FormatInt(a.PermissionID, 10),
		TableID:      strconv.FormatInt(a.TableID, 10),
		Mode:         strconv.FormatInt(a.Mode, 10),
		Label:        a.Label,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

///MAPS
func (d Database) MapPermission(a mdb.Permissions) Permissions {
	return Permissions{
		PermissionID: a.PermissionID,
		TableID:      a.TableID,
		Mode:         a.Mode,
		Label:        a.Label,
	}
}

func (d Database) MapCreatePermissionParams(a CreatePermissionParams) mdb.CreatePermissionParams {
	return mdb.CreatePermissionParams{
		TableID: a.TableID,
		Mode:    a.Mode,
		Label:   a.Label,
	}
}

func (d Database) MapUpdatePermissionParams(a UpdatePermissionParams) mdb.UpdatePermissionParams {
	return mdb.UpdatePermissionParams{
		TableID:      a.TableID,
		Mode:         a.Mode,
		Label:        a.Label,
		PermissionID: a.PermissionID,
	}
}

///QUERIES
func (d Database) CountPermissions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreatePermissionTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

func (d Database) CreatePermission(s CreatePermissionParams) Permissions {
	params := d.MapCreatePermissionParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreatePermission(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreatePermission: %v\n", err)
	}
	return d.MapPermission(row)
}

func (d Database) DeletePermission(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeletePermission(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Permission: %v ", id)
	}
	return nil
}

func (d Database) GetPermission(id int64) (*Permissions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPermission(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapPermission(row)
	return &res, nil
}

func (d Database) ListPermissions() (*[]Permissions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Permissions: %v\n", err)
	}
	res := []Permissions{}
	for _, v := range rows {
		m := d.MapPermission(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdatePermission(s UpdatePermissionParams) (*string, error) {
	params := d.MapUpdatePermissionParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdatePermission(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

///MAPS
func (d MysqlDatabase) MapPermission(a mdbm.Permissions) Permissions {
	return Permissions{
		PermissionID: int64(a.PermissionID),
		TableID:      int64(a.TableID),
		Mode:         int64(a.Mode),
		Label:        a.Label,
	}
}

func (d MysqlDatabase) MapCreatePermissionParams(a CreatePermissionParams) mdbm.CreatePermissionParams {
	return mdbm.CreatePermissionParams{
		TableID: int32(a.TableID),
		Mode:    int32(a.Mode),
		Label:   a.Label,
	}
}

func (d MysqlDatabase) MapUpdatePermissionParams(a UpdatePermissionParams) mdbm.UpdatePermissionParams {
	return mdbm.UpdatePermissionParams{
		TableID:      int32(a.TableID),
		Mode:         int32(a.Mode),
		Label:        a.Label,
		PermissionID: int32(a.PermissionID),
	}
}

///QUERIES
func (d MysqlDatabase) CountPermissions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreatePermissionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

func (d MysqlDatabase) CreatePermission(s CreatePermissionParams) Permissions {
	params := d.MapCreatePermissionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreatePermission(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreatePermission: %v\n", err)
	}
	row, err := queries.GetLastPermission(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Permission: %v\n", err)
	}
	return d.MapPermission(row)
}

func (d MysqlDatabase) DeletePermission(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeletePermission(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Permission: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetPermission(id int64) (*Permissions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPermission(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapPermission(row)
	return &res, nil
}

func (d MysqlDatabase) ListPermissions() (*[]Permissions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Permissions: %v\n", err)
	}
	res := []Permissions{}
	for _, v := range rows {
		m := d.MapPermission(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdatePermission(s UpdatePermissionParams) (*string, error) {
	params := d.MapUpdatePermissionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdatePermission(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

///MAPS
func (d PsqlDatabase) MapPermission(a mdbp.Permissions) Permissions {
	return Permissions{
		PermissionID: int64(a.PermissionID),
		TableID:      int64(a.TableID),
		Mode:         int64(a.Mode),
		Label:        a.Label,
	}
}

func (d PsqlDatabase) MapCreatePermissionParams(a CreatePermissionParams) mdbp.CreatePermissionParams {
	return mdbp.CreatePermissionParams{
		TableID: int32(a.TableID),
		Mode:    int32(a.Mode),
		Label:   a.Label,
	}
}

func (d PsqlDatabase) MapUpdatePermissionParams(a UpdatePermissionParams) mdbp.UpdatePermissionParams {
	return mdbp.UpdatePermissionParams{
		TableID:      int32(a.TableID),
		Mode:         int32(a.Mode),
		Label:        a.Label,
		PermissionID: int32(a.PermissionID),
	}
}

///QUERIES
func (d PsqlDatabase) CountPermissions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreatePermissionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

func (d PsqlDatabase) CreatePermission(s CreatePermissionParams) Permissions {
	params := d.MapCreatePermissionParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreatePermission(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreatePermission: %v\n", err)
	}
	return d.MapPermission(row)
}

func (d PsqlDatabase) DeletePermission(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeletePermission(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Permission: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetPermission(id int64) (*Permissions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPermission(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapPermission(row)
	return &res, nil
}

func (d PsqlDatabase) ListPermissions() (*[]Permissions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Permissions: %v\n", err)
	}
	res := []Permissions{}
	for _, v := range rows {
		m := d.MapPermission(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdatePermission(s UpdatePermissionParams) (*string, error) {
	params := d.MapUpdatePermissionParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdatePermission(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
