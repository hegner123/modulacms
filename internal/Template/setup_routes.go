package main

import mdb "github.com/hegner123/modulacms/db-sqlite"

func createBaseAdminRoutes(dbName string) error {
	db, ctx, err := getDb(Database{src: dbName})
	if err != nil {
		return err
	}
	defer db.Close()

	homePage := mdb.CreateAdminRouteParams{
		Author:       "system",
		AuthorID:     1,
		Slug:         "/",
		Title:        "ModulaCMS",
		Status:       0,
		DateCreated:  ns(timestampS()),
		DateModified: ns(timestampS()),
		Template:     "modula_base.html",
	}
	dbCreateAdminRoute(db, ctx, homePage)
	return nil
}

func createSystemTableEntries() error {
	db, ctx, err := getDb(Database{})
	if err != nil {
        return err
	}
	defer db.Close()
	systemTables := []string{
		"adminroute", "datatype", "field", "media", "media_dimension", "route",
		"table", "token", "user",
	}
	for _, v := range systemTables {
		table := mdb.Tables{Label: ns(v)}
		dbCreateTable(db, ctx, table)
	}
    return nil
}

func createSystemUser(name string)error {
	db, ctx, err := getDb(Database{src: name})
	if err != nil {
        return err
	}

	defer db.Close()

	systemUser := mdb.CreateUserParams{
		DateCreated:  ns(timestampS()),
		DateModified: ns(timestampS()),
		Username:     "system",
		Email:        "system@modulacms.com",
		Name:         "system",
		Hash:         "",
		Role:         "Admin",
	}
	dbCreateUser(db, ctx, systemUser)
    return err
}
