package main

import (
	"context"
	"database/sql"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func insertPlaceholders(db *sql.DB, ctx context.Context, modify string) {
	times := timestampS()
	dbCreateUser(db, ctx, mdb.CreateUserParams{
		Datecreated:  times,
		Datemodified: times,
		Username:     "systeminit" + modify,
		Name:         "system",
		Email:        "system@modulacms.com" + modify,
		Hash:         "has",
		Role:         "admin",
	})
	dbCreateAdminRoute(db, ctx, mdb.CreateAdminRouteParams{
		Author:       "systeminit" + modify,
		Authorid:     1,
		Slug:         "/test1" + modify,
		Title:        "Test",
		Status:       0,
		Template:     "page.html",
		Datecreated:  times,
		Datemodified: times,
	})
	dbCreateRoute(db, ctx, mdb.CreateRouteParams{
		Author:       "systeminit" + modify,
		Authorid:     1,
		Slug:         "/test1" + modify,
		Title:        "Test",
		Status:       0,
		Datecreated:  times,
		Datemodified: times,
	})
	dbCreateMedia(db, ctx, mdb.CreateMediaParams{
		Name:               ns("test.png"),
		Displayname:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "systeminit" + modify,
		Authorid:           1,
		Datecreated:        times,
		Datemodified:       times,
		Url:                ns("public/2024/11/test1.png" + modify),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		Optimizedmobile:    ns("public/2024/11/test-mobile.png" + modify),
		Optimizedtablet:    ns("public/2024/11/test-tablet.png" + modify),
		Optimizeddesktop:   ns("public/2024/11/test-desktop.png" + modify),
		Optimizedultrawide: ns("public/2024/11/test-ultra.png" + modify),
	})

	dbCreateField(db, ctx, mdb.CreateFieldParams{
		Routeid:      int64(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "systeminit" + modify,
		Authorid:     1,
		Datecreated:  ns(times),
		Datemodified: ns(times),
	})
	dbCreateMediaDimension(db, ctx, mdb.CreateMediaDimensionParams{
		Label:  ns("Tablet" + modify),
		Width:  ni(1920),
		Height: ni(1080),
	})

	dbCreateTable(db, ctx, mdb.Tables{Label: ns("Test1" + modify)})
}
