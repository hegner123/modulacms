package db

import (
	"context"
	"database/sql"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func insertPlaceholders(db *sql.DB, ctx context.Context, modify string) {
	times := timestampS()
	dbCreateUser(db, ctx, mdb.CreateUserParams{
		DateCreated:  ns(times),
		DateModified: ns(times),
		Username:     "systeminit" + modify,
		Name:         "system",
		Email:        "system@modulacms.com" + modify,
		Hash:         "has",
		Role:         "admin",
	})
	dbCreateAdminRoute(db, ctx, mdb.CreateAdminRouteParams{
		Author:       "systeminit" + modify,
		AuthorID:     1,
		Slug:         "/test1" + modify,
		Title:        "Test",
		Status:       0,
		Template:     "page.html",
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	dbCreateRoute(db, ctx, mdb.CreateRouteParams{
		Author:       "systeminit" + modify,
		AuthorID:     1,
		Slug:         "/test1" + modify,
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	dbCreateMedia(db, ctx, mdb.CreateMediaParams{
		Name:               ns("test.png"),
		DisplayName:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "systeminit" + modify,
		AuthorID:           1,
		DateCreated:        ns(times),
		DateModified:       ns(times),
		Url:                ns("public/2024/11/test1.png" + modify),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		OptimizedMobile:    ns("public/2024/11/test-mobile.png" + modify),
		OptimizedTablet:    ns("public/2024/11/test-tablet.png" + modify),
		OptimizedDesktop:   ns("public/2024/11/test-desktop.png" + modify),
		OptimizedUltraWide: ns("public/2024/11/test-ultra.png" + modify),
	})

	_,_ =dbCreateField(db, ctx, mdb.CreateFieldParams{
		RouteID:      ni64(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "systeminit" + modify,
		AuthorID:     1,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	dbCreateMediaDimension(db, ctx, mdb.CreateMediaDimensionParams{
		Label:  ns("Tablet" + modify),
		Width:  ni(1920),
		Height: ni(1080),
        AspectRatio: ns("100x100"),
	})

	dbCreateTable(db, ctx, mdb.Tables{Label: ns("Test1" + modify)})
}
