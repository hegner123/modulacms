package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

const mediaTable string = `
    CREATE TABLE IF NOT EXISTS media (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    displayName TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    createdBy INTEGER,
    dateCreated TEXT,
    dateModified TEXT,
    url TEXT,
    mimeType TEXT,
    dimensions TEXT,
    optimizedMobile TEXT,
    optimizedTablet TEXT,
    optimizedDesktop TEXT,
    optimizedUltrawide TEXT);`

const userTable string = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY ,
        username TEXT,
		name TEXT,
		email TEXT UNIQUE,
        hash TEXT,
        role TEXT);`

const adminPostsTable string = `
	CREATE TABLE IF NOT EXISTS adminposts (
		id INTEGER PRIMARY KEY ,
        slug TEXT NOT NULL,
        author TEXT,
        authorId INTEGER,
		title TEXT,
		status INTEGER NOT NULL,
        dateCreated TEXT NOT NULL,
        dateModified TEXT NOT NULL,
        content TEXT NOT NULL,
        type TEXT NOT NULL,
        template TEXT  );`

const postsTable string = `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY ,
        slug TEXT NOT NULL,
        author TEXT,
        authorId INTEGER,
		title TEXT,
		status INTEGER NOT NULL,
        dateCreated TEXT NOT NULL,
        dateModified TEXT NOT NULL,
        content TEXT NOT NULL,
        type TEXT NOT NULL,
        template TEXT  );`

const fieldsTable string = `
    CREATE TABLE IF NOT EXISTS fields(
        id INTEGER PRIMARY KEY ,
        postId INTEGER NOT NULL,
        author TEXT,
        authorId TEXT,
        key TEXT,
        data TEXT,
        dateCreated TEXT,
        dateModified TEXT,
        component TEXT,
        tags TEXT,
        parent string);`

const tables string = `
    CREATE TABLE IF NOT EXISTS tables (
    id INTEGER PRIMARY KEY,
    label TEXT UNIQUE);
    `

var times = timestamp()
var insertHomeRoute string = fmt.Sprintf(`
    INSERT INTO adminposts (slug, title,status,dateCreated, dateModified, content, type,  template) VALUES 
    ('/','home',0,%s,%s,"content","page",'default.html');
    `, times, times)

var insertPagesRoute string = fmt.Sprintf(`
    INSERT INTO adminposts (slug, title, status, dateCreated, dateModified, content, type, template) VALUES 
    ('/pages', 'pages', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertTypesRoute string = fmt.Sprintf(`
    INSERT INTO adminposts (slug, title, status, dateCreated, dateModified, content, type, template) VALUES 
    ('/types', 'types', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertFieldsRoute string = fmt.Sprintf(`
    INSERT INTO adminposts (slug, title, status, dateCreated, dateModified, content, type, template) VALUES 
    ('/fields', 'fields', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertMenusRoute string = fmt.Sprintf(`
    INSERT INTO adminposts (slug, title, status, dateCreated, dateModified, content, type, template) VALUES 
    ('/menus', 'menus', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertUsersRoute string = fmt.Sprintf(`
    INSERT INTO adminposts (slug, title, status, dateCreated, dateModified, content, type, template) VALUES 
    ('/users', 'users', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertMediaRoute string = fmt.Sprintf(`
    INSERT INTO adminposts (slug, title, status, dateCreated, dateModified, content, type, template) VALUES 
    ('/media', 'media', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertTestField string = fmt.Sprintf(`
    INSERT INTO fields (postId, author, authorId, key, data, dateCreated, dateModified, component, tags,parent) VALUES
    (4,'system','0','link_url','https://example.com',%s, %s,'link.html','','');
    `, "1730634309", "1730634309")

/*
const fieldsTable string = `

	CREATE TABLE IF NOT EXISTS fields(
	    id INTEGER PRIMARY KEY ,
	    postId INTEGER NOT NULL,
	    author TEXT,
	    authorId TEXT,
	    key TEXT,
	    data TEXT,
	    dateCreated TEXT,
	    dateModified TEXT,
	    component TEXT,
	    tags TEXT,
	    parent INTEGER);`
*/
const insertDefaultTables string = `
    INSERT INTO tables (label) VALUES ('tables');
    INSERT INTO tables (label) VALUES ('fields');
    INSERT INTO tables (label) VALUES ('media');
    INSERT INTO tables (label) VALUES ('posts');
    INSERT INTO tables (label) VALUES ('adminposts');
    INSERT INTO tables (label) VALUES ('users');
    `

func getDb(dbName Database) (*sql.DB, error) {
    if dbName.DB == "" {
        dbName.DB = "./modula.db"
    }
	db, err := sql.Open("sqlite3", dbName.DB)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}
	return db, nil
}


func initializeDatabase(reset bool) (*sql.DB, error) {
	db, err := getDb(Database{})
	if err != nil {
		return nil, err
	}
	if reset {
		res, err := db.Exec(`
            DROP TABLE IF EXISTS users;
            DROP TABLE IF EXISTS posts;
            DROP TABLE IF EXISTS adminposts;
            DROP TABLE IF EXISTS fields;
            DROP TABLE IF EXISTS media;
            DROP TABLE IF EXISTS tables;

            `)
		if err != nil {
			log.Fatal("I CAN'T FIND THE DATABASE CAPTIN!!!!\n Oh GOD IT'S GOT MY LEG!!!!")
		}
		if res != nil {
			log.Print(res)
		}
	}

	statements := []string{tables, insertDefaultTables, userTable, adminPostsTable, postsTable, fieldsTable, mediaTable}
	routes := []string{insertHomeRoute, insertPagesRoute, insertTypesRoute, insertFieldsRoute, insertMenusRoute, insertUsersRoute, insertMediaRoute, insertTestField}
	err = forEachStatement(db, statements)
	if err != nil {
		fmt.Printf("db exec err: %s", err)
	}
    err = forEachStatement(db, routes)
    if err != nil {
        fmt.Printf("db exec err: %s", err)
    }

	return db, err
}

func initializeClientDatabase(clientDB string, clientReset bool) (*sql.DB, error) {
	dbName := RemoveTLD(clientDB)
	db, err := sql.Open("sqlite3", "./"+dbName+".db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}
	return db, nil
}
