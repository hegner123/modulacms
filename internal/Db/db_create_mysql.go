package db

import (
	_ "embed"
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)

func (d MysqlDatabase) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
	params := d.MapCreateAdminDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminDatatype  %v \n", err)
	}
	row, err := queries.GetLastAdminDatatype(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last admin datatype\n")
	}

	return d.MapAdminDatatype(row)
}
func (d MysqlDatabase) CreateAdminField(s CreateAdminFieldParams) AdminFields {
	params := d.MapCreateAdminFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminField(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminField  %v \n", err)
	}
	row, err := queries.GetLastAdminField(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last admin field\n")
	}

	return d.MapAdminField(row)
}

func (d MysqlDatabase) CreateAdminRoute(s CreateAdminRouteParams) AdminRoutes {
	params := d.MapCreateAdminRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminRoute(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}
	row, err := queries.GetLastAdminRoute(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last admin route\n")
	}

	return d.MapAdminRoute(row)
}

func (d MysqlDatabase) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}
	row, err := queries.GetLastContentData(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last content data\n")
	}

	return d.MapContentData(row)
}

func (d MysqlDatabase) CreateContentField(s CreateContentFieldParams) ContentFields {
	params := d.MapCreateContentFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentField(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}
	row, err := queries.GetLastContentField(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last content field\n")
	}

	return d.MapContentField(row)
}

func (d MysqlDatabase) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateDatatype  %v \n", err)
	}
	row, err := queries.GetLastDatatype(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last  datatype\n")
	}

	return d.MapDatatype(row)
}

func (d MysqlDatabase) CreateField(s CreateFieldParams) Fields {
	params := d.MapCreateFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateField(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateField  %v \n", err)
	}
	row, err := queries.GetLastField(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last field\n")
	}

	return d.MapField(row)
}

func (d MysqlDatabase) CreateMedia(s CreateMediaParams) Media {
	params := d.MapCreateMediaParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateMedia(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateMedia.\n%v \n", err)
	}
	row, err := queries.GetLastMedia(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last media\n")
	}

	return d.MapMedia(row)
}

func (d MysqlDatabase) CreateMediaDimension(s CreateMediaDimensionParams) MediaDimensions {
	params := d.MapCreateMediaDimensionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateMediaDimension(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateMediaDimension.\n%v \n", err)
	}
	row, err := queries.GetLastMediaDimension(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last MediaDimensions\n")
	}

	return d.MapMediaDimension(row)
}
func (d MysqlDatabase) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateRoute.\n %v\n", err)
	}
	row, err := queries.GetLastRole(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last role\n")
	}

	return d.MapRoles(row)
}

func (d MysqlDatabase) CreateRoute(s CreateRouteParams) Routes {
	params := d.MapCreateRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateRoute(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateRoute.\n %v\n", err)
	}
	row, err := queries.GetLastRoute(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last route\n")
	}

	return d.MapRoute(row)
}

func (d MysqlDatabase) CreateTable(s string) Tables {
	params := ns(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateTable(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateTable.\n %v\n", err)
	}
	row, err := queries.GetLastTable(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last table\n")
	}

	return d.MapTables(row)
}

func (d MysqlDatabase) CreateToken(s CreateTokenParams) Tokens {
	params := d.MapCreateTokenParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateToken(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateToken.\n %v\n", err)
	}
	row, err := queries.GetLastToken(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last token\n")
	}

	return d.MapToken(row)
}

func (d MysqlDatabase) CreateUser(s CreateUserParams) Users {
	params := d.MapCreateUserParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateUser(d.Context, params)
	if err != nil {
		fmt.Printf("failed to CreateUser,\n %v\n", err)
	}
	row, err := queries.GetLastUser(d.Context)
	if err != nil {
		fmt.Printf("Couldn't Get last user\n")
	}

	return d.MapUser(row)
}
