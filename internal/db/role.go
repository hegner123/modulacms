package db

import (
	"encoding/json"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/sqlc-dev/pqtype"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type Roles struct {
	RoleID      types.RoleID `json:"role_id"`
	Label       string       `json:"label"`
	Permissions string       `json:"permissions"`
}

type CreateRoleParams struct {
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}

type UpdateRoleParams struct {
	Label       string       `json:"label"`
	Permissions string       `json:"permissions"`
	RoleID      types.RoleID `json:"role_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapRole(a mdb.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func (d Database) MapCreateRoleParams(a CreateRoleParams) mdb.CreateRoleParams {
	return mdb.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func (d Database) MapUpdateRoleParams(a UpdateRoleParams) mdb.UpdateRoleParams {
	return mdb.UpdateRoleParams{
		Label:       a.Label,
		Permissions: a.Permissions,
		RoleID:      a.RoleID,
	}
}

// QUERIES

func (d Database) CountRoles() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateRoleTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d Database) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRole: %v\n", err)
	}
	return d.MapRole(row)
}

func (d Database) DeleteRole(id types.RoleID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRole(d.Context, mdb.DeleteRoleParams{RoleID: id})
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", id)
	}
	return nil
}

func (d Database) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdb.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

func (d Database) ListRoles() (*[]Roles, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRole(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update role, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapRole(a mdbm.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: a.Permissions.String,
	}
}

func (d MysqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbm.CreateRoleParams {
	return mdbm.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       a.Label,
		Permissions: StringToNullString(a.Permissions),
	}
}

func (d MysqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbm.UpdateRoleParams {
	return mdbm.UpdateRoleParams{
		Label:       a.Label,
		Permissions: StringToNullString(a.Permissions),
		RoleID:      a.RoleID,
	}
}

// QUERIES

func (d MysqlDatabase) CountRoles() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateRoleTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRole: %v\n", err)
	}
	row, err := queries.GetRole(d.Context, mdbm.GetRoleParams{RoleID: params.RoleID})
	if err != nil {
		fmt.Printf("Failed to get last inserted Role: %v\n", err)
	}
	return d.MapRole(row)
}

func (d MysqlDatabase) DeleteRole(id types.RoleID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRole(d.Context, mdbm.DeleteRoleParams{RoleID: id})
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", id)
	}
	return nil
}

func (d MysqlDatabase) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdbm.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

func (d MysqlDatabase) ListRoles() (*[]Roles, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRole(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update role, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapRole(a mdbp.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: string(a.Permissions.RawMessage),
	}
}

func (d PsqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbp.CreateRoleParams {
	return mdbp.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       a.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(a.Permissions)},
	}
}

func (d PsqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbp.UpdateRoleParams {
	return mdbp.UpdateRoleParams{
		Label:       a.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(a.Permissions)},
		RoleID:      a.RoleID,
	}
}

// QUERIES

func (d PsqlDatabase) CountRoles() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateRoleTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRole: %v\n", err)
	}
	return d.MapRole(row)
}

func (d PsqlDatabase) DeleteRole(id types.RoleID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRole(d.Context, mdbp.DeleteRoleParams{RoleID: id})
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", id)
	}
	return nil
}

func (d PsqlDatabase) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdbp.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

func (d PsqlDatabase) ListRoles() (*[]Roles, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRole(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update role, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
