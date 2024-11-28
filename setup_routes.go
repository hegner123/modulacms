package main

import mdb "github.com/hegner123/modulacms/db-sqlite"

func createBaseAdminRoutes() {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get db", err)
	}
	defer db.Close()

	homePage := mdb.CreateAdminRouteParams{
		Author:       "system",
		Authorid:     1,
		Slug:         "/",
		Title:        "ModulaCMS",
		Status:       0,
		Datecreated:  timestampS(),
		Datemodified: timestampS(),
		Template:     "modula_base.html",
	}
	dbCreateAdminRoute(db, ctx, homePage)
}

func createSystemTableEntries() {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get db", err)
	}
	defer db.Close()
	systemTables := []string{
		"adminroute", "datatype", "field", "media", "media_dimension", "route",
		"table", "token", "user",
	}
    for _,v := range systemTables{
        table := mdb.Tables{Label: ns(v)}
        dbCreateTable(db,ctx, table)
    }
}

func createSystemUser() {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get db", err)
	}
	defer db.Close()

	systemUser := mdb.CreateUserParams{
		Datecreated:  timestampS(),
		Datemodified: timestampS(),
		Username:     "system",
		Email:        "system@modulacms.com",
		Name:         "system",
		Hash:         "",
		Role:         "Admin",
	}
	dbCreateUser(db, ctx, systemUser)
}
