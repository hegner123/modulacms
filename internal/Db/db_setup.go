package db

import (
	"context"
	"database/sql"

	"fmt"
	mdb "github.com/hegner123/modulacms/db-sqlite"
	"time"
)

func TimestampI() int64 {
	return time.Now().Unix()
}

func TimestampS() string {
	return fmt.Sprint(time.Now().Unix())
}

func createSetupInserts(db *sql.DB, ctx context.Context) error {
	times := TimestampS()
	CreateUser(db, ctx, mdb.CreateUserParams{
		DateCreated:  ns(times),
		DateModified: ns(times),
		Username:     "system",
		Name:         "system",
		Email:        "system@modulacms.com",
		Hash:         "has",
		Role:         "admin",
	})
	CreateAdminRoute(db, ctx, mdb.CreateAdminRouteParams{
		Author:       "system",
		AuthorID:     1,
		Slug:         "/",
		Title:        "Admin",
		Status:       0,
		Template:     "modula_base.html",
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	CreateRoute(db, ctx, mdb.CreateRouteParams{
		Author:       "system",
		AuthorID:     1,
		Slug:         "/api/v1/",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	CreateMedia(db, ctx, mdb.CreateMediaParams{
		Name:               ns("test.png"),
		DisplayName:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "system",
		AuthorID:           1,
		DateCreated:        ns(times),
		DateModified:       ns(times),
		Url:                ns("public/2024/11/test1.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		OptimizedMobile:    ns("public/2024/11/test-mobile.png"),
		OptimizedTablet:    ns("public/2024/11/test-tablet.png"),
		OptimizedDesktop:   ns("public/2024/11/test-desktop.png"),
		OptimizedUltraWide: ns("public/2024/11/test-ultra.png"),
	})
	_, err := CreateDataType(db, ctx, mdb.CreateDatatypeParams{
		Label:        "Parent",
		Type:         "Navigation",
		Author:       "system",
		AuthorID:     1,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	if err != nil {
		return err
	}

	_, err = CreateField(db, ctx, mdb.CreateFieldParams{
		RouteID:      ni(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "system",
		AuthorID:     1,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	if err != nil {
		return err
	}
	CreateMediaDimension(db, ctx, mdb.CreateMediaDimensionParams{
		Label:       ns("Tablet"),
		Width:       ni(1920),
		Height:      ni(1080),
		AspectRatio: ns("16:9"),
	})
	return nil
}
