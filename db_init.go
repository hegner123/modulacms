package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

const fieldsTable string = `CREATE TABLE fields (
    id INTEGER PRIMARY KEY,
    routeid INTEGER,
    author TEXT,
    authorid TEXT,
    key TEXT,
    type TEXT,
    data TEXT,
    datecreated TEXT,
    datemodified TEXT,
    componentid INTEGER,
    tags TEXT,
    parent TEXT,
    FOREIGN KEY (componentid) REFERENCES elements(id)
);`

const elementsTable string = `CREATE TABLE elements (
    id INTEGER PRIMARY KEY,
    tag TEXT
);`

const attributesTable string = `CREATE TABLE attributes (
    id INTEGER PRIMARY KEY,
    elementid INTEGER,
    key TEXT,
    value TEXT,
    FOREIGN KEY (elementid) REFERENCES elements(id)
);`

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

var insertFields string = fmt.Sprintf(`INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent)
VALUES (1, 'Alice', 'alice123', 'field_key_1', 'text', 'Sample data for field 1', '%s', '%s', 1, 'tag1,tag2', 'parent1');

INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent)
VALUES (2, 'Bob', 'bob456', 'field_key_2', 'number', '42', '%s', '%s', 2, 'tag3,tag4', 'parent2');

INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent)
VALUES (3, 'Charlie', 'charlie789', 'field_key_3', 'text', 'Another example of field data', '%s', '%s', 3, 'tag5', 'parent3');`, Times, Times, Times, Times, Times, Times)

var insertElement string = `INSERT INTO elements (id, tag) VALUES
(1, 'div'),
(2, 'span'),
(3, 'section');`

var insertAttribute string = `INSERT INTO attributes (elementid, key, value) VALUES
(1, 'class', 'container'),
(1, 'id', 'main-div'),
(2, 'style', 'color: red; font-size: 14px;'),
(2, 'data-role', 'user-info'),
(3, 'class', 'content-section'),
(3, 'data-id', 'section-123');
`

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
    INSERT INTO tables (label) VALUES ('elements');
    INSERT INTO tables (label) VALUES ('attributes');
    `

func getDb(dbName Database) (*sql.DB, error) {
	if dbName.DB == "" {
		dbName.DB = "./modula.db"
	}
	db, err := sql.Open("sqlite3", dbName.DB)
	if err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
		return nil, err
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
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
            DROP TABLE IF EXISTS elements;
            DROP TABLE IF EXISTS attributes;`)
		if err != nil {
			fmt.Printf("db exec err db_init 006 : %s\n", err)
			log.Fatal("I CAN'T FIND THE DATABASE CAPTIN!!!!\n Oh GOD IT'S GOT MY LEG!!!!")
		}
		if res != nil {
			log.Print(res)
		}
	}
	u := User{}
	ar := AdminRoute{}
	r := Routes{}
	m := Media{}
	md := MediaDimension{}
	userTable := formatCreateTable(u, "users")
	adminRoutesTable := formatCreateTable(ar, "adminroutes")
	routesTable := formatCreateTable(r, "routes")
	mediaTable := formatCreateTable(m, "media")
	mediaDimensionTable := formatCreateTable(md, "media_dimensions")

	statements := []string{tables, insertDefaultTables, userTable, adminRoutesTable, routesTable, fieldsTable, mediaTable, mediaDimensionTable, elementsTable, attributesTable}
	routes := []string{insertHomeRoute, insertPagesRoute, insertTypesRoute, insertFieldsRoute, insertMenusRoute, insertUsersRoute, insertMediaRoute,insertElement, insertFields, insertAttribute}
	systemUser := []string{insertSystemUser}

	err := forEachStatement(db, statements, "tables")
	if err != nil {
		fmt.Printf("db exec err db_init 001 : %s\n", err)
	}
	err = forEachStatement(db, routes, "routes")
	if err != nil {
		fmt.Printf("db exec err db_init 002  : %s\n", err)
	}
	err = forEachStatement(db, systemUser, "systemUser")
	if err != nil {
		fmt.Printf("db exec err db_init 003 : %s\n", err)
	}
	return err
}

func initializeClientDatabase(clientDB string, clientReset bool) (*sql.DB, error) {
	dbName := RemoveTLD(clientDB)
	db, err := sql.Open("sqlite3", "./"+dbName+".db")
	if err != nil {
		fmt.Printf("db exec err db_init 004 : %s\n", err)
		return nil, err
	}
	res, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 005 : %s\n", err)
		return nil, err
	}
	fmt.Print(res)
	return db, nil
}
