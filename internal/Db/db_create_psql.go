package db

import (
	_ "embed"
	"fmt"

	mdbp "github.com/hegner123/modulacms/db-psql"
)

func (d PsqlDatabase) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
    params := d.MapCreateAdminDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminDatatype(row)
}
func (d PsqlDatabase) CreateAdminField(s CreateAdminFieldParams) AdminFields {
    params := d.MapCreateAdminFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminField(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminField  %v \n", err)
	}

	return d.MapAdminField(row)
}

func (d PsqlDatabase) CreateAdminRoute(s CreateAdminRouteParams) AdminRoutes {
    params:= d.MapCreateAdminRouteParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminRoute(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return d.MapAdminRoute(row)
}

func (d PsqlDatabase) CreateContentData(s CreateContentDataParams) ContentData {
    params := d.MapCreateContentDataParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return d.MapContentData(row)
}

func (d PsqlDatabase) CreateContentField(s CreateContentFieldParams) ContentFields {
    params:= d.MapCreateContentFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateContentField(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return d.MapContentField(row)
}

func (d PsqlDatabase) CreateDatatype(s CreateDatatypeParams) Datatypes {
    params := d.MapCreateDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateDatatype  %v \n", err)
	}

	return d.MapDatatype(row)
}

func (d PsqlDatabase) CreateField(s CreateFieldParams) Fields {
    params := d.MapCreateFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateField(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateField  %v \n", err)
	}

	return d.MapField(row)
}

func (d PsqlDatabase) CreateMedia(s CreateMediaParams) Media {
    params := d.MapCreateMediaParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateMedia(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateMedia.\n%v \n", err)
	}

	return d.MapMedia(row)
}

func (d PsqlDatabase) CreateMediaDimension(s CreateMediaDimensionParams) MediaDimensions {
    params := d.MapCreateMediaDimensionParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateMediaDimension(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateMediaDimension.\n%v \n", err)
	}

	return d.MapMediaDimension(row)
}
func (d PsqlDatabase) CreateRole(s CreateRoleParams) Roles {
    params := d.MapCreateRoleParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateRoute.\n %v\n", err)
	}

	return d.MapRoles(row)
}

func (d PsqlDatabase) CreateRoute(s CreateRouteParams) Routes {
    params := d.MapCreateRouteParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateRoute(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateRoute.\n %v\n", err)
	}

	return d.MapRoute(row)
}

func (d PsqlDatabase) CreateTable(s string) Tables {
    params := ns(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateTable(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateTable.\n %v\n", err)
	}

	return d.MapTables(row)
}

func (d PsqlDatabase) CreateToken(s CreateTokenParams) Tokens {
    params := d.MapCreateTokenParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateToken(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateToken.\n %v\n", err)
	}

	return d.MapToken(row)
}

func (d PsqlDatabase) CreateUser(s CreateUserParams) Users {
    params := d.MapCreateUserParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateUser(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateUser,\n %v\n", err)
	}

	return d.MapUser(row)
}
