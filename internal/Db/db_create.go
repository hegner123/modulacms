package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func CreateAdminDatatype(db *sql.DB, ctx context.Context, s mdb.CreateAdminDatatypeParams) mdb.AdminDatatypes {
	queries := mdb.New(db)
	insertedAdminDatatype, err := queries.CreateAdminDatatype(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminDatatype  %v \n", err)
	}

	return insertedAdminDatatype
}
func CreateAdminField(db *sql.DB, ctx context.Context, s mdb.CreateAdminFieldParams) mdb.AdminFields {
	queries := mdb.New(db)
	insertedAdminField, err := queries.CreateAdminField(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminField  %v \n", err)
	}

	return insertedAdminField
}

func CreateAdminRoute(db *sql.DB, ctx context.Context, s mdb.CreateAdminRouteParams) mdb.AdminRoutes {
	queries := mdb.New(db)
	insertedAdminRoute, err := queries.CreateAdminRoute(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return insertedAdminRoute
}

func CreateContentData(db *sql.DB, ctx context.Context, s mdb.CreateContentDataParams) mdb.ContentData {
	queries := mdb.New(db)
	insertedContentData, err := queries.CreateContentData(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return insertedContentData
}

func CreateContentField(db *sql.DB, ctx context.Context, s mdb.CreateContentFieldParams) mdb.ContentFields {
	queries := mdb.New(db)
	insertedContentField, err := queries.CreateContentField(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateAdminRoute  %v \n", err)
	}

	return insertedContentField
}

func CreateDataType(db *sql.DB, ctx context.Context, s mdb.CreateDatatypeParams) (mdb.Datatypes, error) {
	queries := mdb.New(db)
	insertedDatatypes, err := queries.CreateDatatype(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateDatatype  %v \n", err)
		return insertedDatatypes, err
	}

	return insertedDatatypes, nil
}

func CreateField(db *sql.DB, ctx context.Context, s mdb.CreateFieldParams) (mdb.Fields, error) {
	queries := mdb.New(db)
	insertedFields, err := queries.CreateField(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateField  %v \n", err)
		return insertedFields, err
	}

	return insertedFields, nil
}

func CreateMedia(db *sql.DB, ctx context.Context, s mdb.CreateMediaParams) mdb.Media {
	queries := mdb.New(db)
	insertedMedia, err := queries.CreateMedia(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateMedia.\n%v \n", err)
	}

	return insertedMedia
}

func CreateMediaDimension(db *sql.DB, ctx context.Context, s mdb.CreateMediaDimensionParams) mdb.MediaDimensions {
	queries := mdb.New(db)
	insertedMediaDimension, err := queries.CreateMediaDimension(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateMediaDimension.\n%v \n", err)
	}

	return insertedMediaDimension
}
func (d Database) CreateRole(db *sql.DB, ctx context.Context, s CreateRoleParams) *Roles {
	queries := mdb.New(db)
	params := mdb.CreateRoleParams{Label: s.Label, Permissions: s.Permissions}
	row, err := queries.CreateRole(ctx, params)
	if err != nil {
		fmt.Printf("failed to CreateRoute.\n %v\n", err)
	}
	res := Roles{RoleID: row.RoleID, Label: row.Label, Permissions: row.Permissions}

	return &res
}

func CreateRoute(db *sql.DB, ctx context.Context, s mdb.CreateRouteParams) mdb.Routes {
	queries := mdb.New(db)
	insertedRoute, err := queries.CreateRoute(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateRoute.\n %v\n", err)
	}

	return insertedRoute
}

func CreateTable(db *sql.DB, ctx context.Context, s mdb.Tables) mdb.Tables {
	queries := mdb.New(db)
	insertedTable, err := queries.CreateTable(ctx, s.Label)
	if err != nil {
		fmt.Printf("failed to CreateTable.\n %v\n", err)
	}

	return insertedTable
}

func CreateToken(db *sql.DB, ctx context.Context, s mdb.CreateTokenParams) mdb.Tokens {
	queries := mdb.New(db)
	insertedToken, err := queries.CreateToken(ctx, s)
	if err != nil {
		fmt.Printf("failed to CreateToken.\n %v\n", err)
	}

	return insertedToken
}

func CreateUser(db *sql.DB, ctx context.Context, s mdb.CreateUserParams) mdb.Users {
	queries := mdb.New(db)
	insertedUser, err := queries.CreateUser(ctx, s)
	if err != nil {
		splitErr := strings.Split(err.Error(), ".")
		property := splitErr[len(splitErr)-1]
		v := getColumnValue(property, s)
		fmt.Printf("failed to CreateUser.\n %v\n %v\n %v\n", err, property, v)
	}

	return insertedUser
}
