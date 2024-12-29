package mTemplate

import (
	mdb "github.com/hegner123/modulacms/db-sqlite"
	db "github.com/hegner123/modulacms/internal/Db"
)

func createBaseAdminRoutes(dbName string) error {
	dbc := db.GetDb(db.Database{})

	defer dbc.Connection.Close()
	homePage := mdb.CreateAdminRouteParams{
		Author:       "system",
		AuthorID:     1,
		Slug:         "/",
		Title:        "ModulaCMS",
		Status:       0,
		DateCreated:  db.Ns(db.TimestampS()),
		DateModified: db.Ns(db.TimestampS()),
		Template:     "modula_base.html",
	}
	db.CreateAdminRoute(dbc.Connection, dbc.Context, homePage)
	return nil
}

func createSystemTableEntries() error {
	dbc := db.GetDb(db.Database{})
	defer dbc.Connection.Close()
	systemTables := []string{
		"adminroute", "datatype", "field", "media", "media_dimension", "route",
		"table", "token", "user",
	}
	for _, v := range systemTables {
		table := mdb.Tables{Label: db.Ns(v)}
		db.CreateTable(dbc.Connection, dbc.Context, table)
	}
	return nil
}

func createSystemUser(name string) error {
	dbc := db.GetDb(db.Database{})
	defer dbc.Connection.Close()

	systemUser := mdb.CreateUserParams{
		DateCreated: db.Ns(db.TimestampS()),
		DateModified: db.Ns(db.TimestampS()),
		Username:     "system",
		Email:        "system@modulacms.com",
		Name:         "system",
		Hash:         "",
		Role:         "Admin",
	}
	db.CreateUser(dbc.Connection, dbc.Context, systemUser)
	return nil
}
