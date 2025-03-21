package db

import (
	_ "embed"
	"fmt"

	mdbp "github.com/hegner123/modulacms/db-psql"
)



func (d PsqlDatabase) ListContentData() (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListContentDataByRoute(id int64) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataByRoute(d.Context, Ni32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFields() (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListContentFieldsByRoute(id int64) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, Ni32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
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

func (d PsqlDatabase) ListFields() (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListMedia() (*[]Media, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Medias: %v\n", err)
	}
	res := []Media{}
	for _, v := range rows {
		m := d.MapMedia(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListMediaDimensions() (*[]MediaDimensions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaDimensions: %v\n", err)
	}
	res := []MediaDimensions{}
	for _, v := range rows {
		m := d.MapMediaDimension(v)
		res = append(res, m)
	}
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
		m := d.MapPermissions(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListRoles() (*[]Roles, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRoles(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListRoutes() (*[]Routes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
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
		m := d.MapTables(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListTokens() (*[]Tokens, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListTokens(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListUsers() (*[]Users, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Users{}
	for _, v := range rows {
		m := d.MapUser(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListTokenDependencies(id int64) {
	// TODO implement dependency checking for delete candidate
}

func (d PsqlDatabase) ListAdminFieldsByDatatypeID(admin_datatype_id int64) (*[]ListAdminFieldsByDatatypeIDRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldsByDatatypeID(d.Context, Ni32(admin_datatype_id))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by AdminDatatypes id: %v\n ", err)
	}
	res := []ListAdminFieldsByDatatypeIDRow{}
	for _, v := range rows {
		m := d.MapListAdminFieldsByDatatypeIDRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypeChildren(parentId int64) (*[]AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeChildren(d.Context, Ni32(parentId))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes by AdminRouteId %v\n", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil

}
func (d PsqlDatabase) ListUserOauths() (*[]UserOauth, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListUserOauth(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []UserOauth{}
	for _, v := range rows {
		m := d.MapUserOauth(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListSessions() (*[]Sessions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListSessions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
	return &res, nil
}
