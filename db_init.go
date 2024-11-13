package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var Times = timestamp()
var insertHomeRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status, datecreated, datemodified, content,  template) VALUES 
    ('/','system',0 ,'home',0,%s,%s,"content",'default.html');
    `, Times, Times)

var insertPagesRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status, datecreated, datemodified, content,  template) VALUES 
    ('/pages','system',0 ,'pages', 0, %s, %s, "content", 'default.html');
    `, Times, Times)

var insertTypesRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status, datecreated, datemodified, content,  template) VALUES 
    ('/types','system',0 ,'types', 0, %s, %s, "content", 'default.html');
    `, Times, Times)

var insertFieldsRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status, datecreated, datemodified, content,  template) VALUES 
    ('/fields','system',0, 'fields', 0, %s, %s, "content", 'default.html');
    `, Times, Times)

var insertMenusRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status, datecreated, datemodified, content,  template) VALUES 
    ('/menus','system',0 ,'menus', 0, %s, %s, "content", 'default.html');
    `, Times, Times)

var insertUsersRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status, datecreated, datemodified, content,  template) VALUES 
    ('/users','system',0, 'users', 0, %s, %s, "content", 'default.html');
    `, Times, Times)

var insertMediaRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status, datecreated, datemodified, content,  template) VALUES 
    ('/media','system',0, 'media', 0, %s, %s, "content", 'default.html');
    `, Times, Times)

var insertTestField string = fmt.Sprintf(`
    INSERT INTO fields (routeId, author, authorId, key, type, data, datecreated, datemodified, component, tags,parent) VALUES
    (4,'system','0','link_url','text','https://example.com',%s, %s,'link.html','','');
    `, "1730634309", "1730634309")

var insertSystemUser string = fmt.Sprintf(`
    INSERT INTO users (datecreated, datemodified, username, name, email, hash, role) VALUES 
    ('%s','%s','system', 'system', 'system@system.com', 'hash', 'root');
    `, Times, Times)

const insertDefaultTables string = `
    INSERT INTO tables (label) VALUES ('tables');
    INSERT INTO tables (label) VALUES ('fields');
    INSERT INTO tables (label) VALUES ('media');
    INSERT INTO tables (label) VALUES ('routes');
    INSERT INTO tables (label) VALUES ('adminroutes');
    INSERT INTO tables (label) VALUES ('users');
    `

func getDb(dbName Database) (*sql.DB, error) {
	if dbName.DB == "" {
		dbName.DB = "./modula.db"
	}
	db, err := sql.Open("sqlite3", dbName.DB)
	if err != nil {
		fmt.Printf("db exec err db_init 007 : %s", err)
		return nil, err
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s", err)
		return nil, err
	}
	return db, nil
}

func initializeDatabase(db *sql.DB, reset bool) error {
	if reset {
		res, err := db.Exec(`
            DROP TABLE IF EXISTS users;
            DROP TABLE IF EXISTS routes;
            DROP TABLE IF EXISTS adminroutes;
            DROP TABLE IF EXISTS fields;
            DROP TABLE IF EXISTS media;
            DROP TABLE IF EXISTS media_dimensions;
            DROP TABLE IF EXISTS tables;

            `)
		if err != nil {
			fmt.Printf("db exec err db_init 006 : %s", err)
			log.Fatal("I CAN'T FIND THE DATABASE CAPTIN!!!!\n Oh GOD IT'S GOT MY LEG!!!!")
		}
		if res != nil {
			log.Print(res)
		}
	}
    u:= User{}
    ar:=AdminRoute{}
    r := Routes{}
    f := Field{}
    m := Media{}
    md := MediaDimension{}
    userTable:= formatCreateTable(u,"users")
    adminRoutesTable:= formatCreateTable(ar,"adminroutes")
    routesTable:= formatCreateTable(r,"routes")
    fieldsTable := formatCreateTable(f,"fields")
    mediaTable := formatCreateTable(m,"media")
    mediaDimensionTable := formatCreateTable(md,"media_dimensions")
    

    
	statements := []string{tables, insertDefaultTables, userTable, adminRoutesTable, routesTable, fieldsTable, mediaTable, mediaDimensionTable}
	routes := []string{insertHomeRoute, insertPagesRoute, insertTypesRoute, insertFieldsRoute, insertMenusRoute, insertUsersRoute, insertMediaRoute, insertTestField}
	systemUser := []string{insertSystemUser}

	err := forEachStatement(db, statements, "tables")
	if err != nil {
		fmt.Printf("db exec err db_init 001 : %s", err)
	}
	err = forEachStatement(db, routes, "routes")
	if err != nil {
		fmt.Printf("db exec err db_init 002  : %s", err)
	}
	err = forEachStatement(db, systemUser, "systemUser")
	if err != nil {
		fmt.Printf("db exec err db_init 003 : %s", err)
	}
	res := dbCreateField(createFieldField)
	fmt.Printf("\n%v\n", res)
	return err
}

func initializeClientDatabase(clientDB string, clientReset bool) (*sql.DB, error) {
	dbName := RemoveTLD(clientDB)
	db, err := sql.Open("sqlite3", "./"+dbName+".db")
	if err != nil {
		fmt.Printf("db exec err db_init 004 : %s", err)
		return nil, err
	}
	res, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 005 : %s", err)
		return nil, err
	}
	fmt.Print(res)
	return db, nil
}
