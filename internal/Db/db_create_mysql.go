package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)

func (d MysqlDatabase) CreateRole(dbc *sql.DB, ctx context.Context, s CreateRoleParams) *Roles {
	queries := mdbm.New(dbc)
	params := mdbm.CreateRoleParams{Label: s.Label, Permissions: s.Permissions}
	err := queries.CreateRole(ctx, params)
	row, err := queries.GetLastRole(ctx)
	if err != nil {
		return nil
	}
	res := Roles{RoleID: int64(row.RoleID), Label: row.Label, Permissions: row.Permissions}

	return &res
}
func (d MysqlDatabase) CreateRoute(dbc *sql.DB, ctx context.Context, s CreateRouteParams) *Routes {
	queries := mdbm.New(dbc)
	params := mdbm.CreateRouteParams{}
	err := queries.CreateRoute(ctx, params)
	row, err := queries.GetLastRoute(ctx)
	if err != nil {
		fmt.Printf("failed to CreateRoute.\n %v\n", err)
		return nil
	}
	res := Routes{
		RouteID:      int64(row.RouteID),
		Author:       row.Author,
		AuthorID:     int64(row.AuthorID),
		Slug:         row.Slug,
		Title:        row.Title,
		Status:       int64(row.Status),
		History:      row.History,
		DateCreated:  ns(row.DateCreated.String()),
		DateModified: ns(row.DateModified.String()),
	}
	return &res
}

/*
func CreateAdminDatatype(db *sql.DB, ctx context.Context, s mdbs.CreateAdminDatatypeParams) mdbs.AdminDatatypes {
	queries := mdbs.New(db)
	insertedAdminDatatype, err := queries.CreateAdminDatatype(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminDatatype  %v \n", err)
	}

	return insertedAdminDatatype
}
func CreateAdminField(db *sql.DB, ctx context.Context, s mdbs.CreateAdminFieldParams) mdbs.AdminFields {
	queries := mdbs.New(db)
	insertedAdminField, err := queries.CreateAdminField(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminField  %v \n", err)
	}

	return insertedAdminField
}

func CreateAdminRoute(db *sql.DB, ctx context.Context, s mdbs.CreateAdminRouteParams) mdbs.AdminRoutes {
	queries := mdbs.New(db)
	insertedAdminRoute, err := queries.CreateAdminRoute(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return insertedAdminRoute
}

func CreateContentData(db *sql.DB, ctx context.Context, s mdbs.CreateContentDataParams) mdbs.ContentData {
	queries := mdbs.New(db)
	insertedContentData, err := queries.CreateContentData(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return insertedContentData
}

func CreateContentField(db *sql.DB, ctx context.Context, s mdbs.CreateContentFieldParams) mdbs.ContentFields {
	queries := mdbs.New(db)
	insertedContentField, err := queries.CreateContentField(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return insertedContentField
}

func CreateDataType(db *sql.DB, ctx context.Context, s mdbs.CreateDatatypeParams) (mdbs.Datatypes, error) {
	queries := mdbs.New(db)
	insertedDatatypes, err := queries.CreateDatatype(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateDatatype  %v \n", err)
		return insertedDatatypes, err
	}

	return insertedDatatypes, nil
}

func CreateField(db *sql.DB, ctx context.Context, s mdbs.CreateFieldParams) (mdbs.Fields, error) {
	queries := mdbs.New(db)
	insertedFields, err := queries.CreateField(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateField  %v \n", err)
		return insertedFields, err
	}

	return insertedFields, nil
}

func CreateMedia(db *sql.DB, ctx context.Context, s mdbs.CreateMediaParams) mdbs.Media {
	queries := mdbs.New(db)
	insertedMedia, err := queries.CreateMedia(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateMedia.\n%v \n", err)
	}

	return insertedMedia
}

func CreateMediaDimension(db *sql.DB, ctx context.Context, s mdbs.CreateMediaDimensionParams) mdbs.MediaDimensions {
	queries := mdbs.New(db)
	insertedMediaDimension, err := queries.CreateMediaDimension(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateMediaDimension.\n%v \n", err)
	}

	return insertedMediaDimension
}


func CreateTable(db *sql.DB, ctx context.Context, s mdbs.Tables) mdbs.Tables {
	queries := mdbs.New(db)
	insertedTable, err := queries.CreateTable(ctx, s.Label)
	if err != nil {
		fmt.Printf("failed to CreateTable.\n %v\n", err)
	}

	return insertedTable
}

func CreateToken(db *sql.DB, ctx context.Context, s mdbs.CreateTokenParams) mdbs.Tokens {
	queries := mdbs.New(db)
	insertedToken, err := queries.CreateToken(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateToken.\n %v\n", err)
	}

	return insertedToken
}

func CreateUser(db *sql.DB, ctx context.Context, s mdbs.CreateUserParams) mdbs.Users {
	queries := mdbs.New(db)
	insertedUser, err := queries.CreateUser(ctx, s)
	if err != nil {
		splitErr := strings.Split(err.Error(), ".")
		property := splitErr[len(splitErr)-1]
		v := getColumnValue(property, s)
		fmt.Printf("failed to CreateUser.\n %v\n %v\n %v\n", err, property, v)
	}

	return insertedUser
}
*/
