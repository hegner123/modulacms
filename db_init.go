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
        displayname TEXT,
        alt TEXT,
        caption TEXT,
        description TEXT,
        class TEXT,
        author TEXT,
        authorid INTEGER,
        datecreated TEXT,
        datemodified TEXT,
        url TEXT,
        mimeType TEXT,
        dimensions TEXT,
        optimizedmobile TEXT,
        optimizedtablet TEXT,
        optimizeddesktop TEXT,
        optimizedultrawide TEXT);`

const userTable string = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY ,
        datecreated TEXT ,
        datemodified TEXT,
        username TEXT,
		name TEXT,
		email TEXT UNIQUE ,
        hash TEXT,
        role TEXT);`

const adminRoutesTable string = `
	CREATE TABLE IF NOT EXISTS adminroutes (
		id INTEGER PRIMARY KEY ,
        slug TEXT NOT NULL,
        author TEXT,
        authorId INTEGER,
		title TEXT,
		status INTEGER NOT NULL,
        datecreated TEXT NOT NULL,
        datemodified TEXT NOT NULL,
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
        datecreated TEXT NOT NULL,
        datemodified TEXT NOT NULL,
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
        datecreated TEXT,
        datemodified TEXT,
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
    INSERT INTO adminroutes (slug, author, authorId, title,status,datecreated, datemodified, content, type,  template) VALUES 
    ('/','system',0 ,'home',0,%s,%s,"content","page",'default.html');
    `, times, times)

var insertPagesRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title,status,datecreated, datemodified, content, type,  template) VALUES 
    ('/pages','system',0 ,'pages', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertTypesRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title,status,datecreated, datemodified, content, type,  template) VALUES 
    ('/types','system',0 ,'types', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertFieldsRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title,status,datecreated, datemodified, content, type,  template) VALUES 
    ('/fields','system',0, 'fields', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertMenusRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title,status,datecreated, datemodified, content, type,  template) VALUES 
    ('/menus','system',0 ,'menus', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertUsersRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title,status,datecreated, datemodified, content, type,  template) VALUES 
    ('/users','system',0, 'users', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertMediaRoute string = fmt.Sprintf(`
    INSERT INTO adminroutes (slug, author, authorId, title, status,datecreated, datemodified, content, type,  template) VALUES 
    ('/media','system',0, 'media', 0, %s, %s, "content", "page", 'default.html');
    `, times, times)

var insertTestField string = fmt.Sprintf(`
    INSERT INTO fields (postId, author, authorId, key, data, datecreated, datemodified, component, tags,parent) VALUES
    (4,'system','0','link_url','https://example.com',%s, %s,'link.html','','');
    `, "1730634309", "1730634309")

/*
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY ,
	        datecreated TEXT ,
	        datemodified TEXT,
	        username TEXT,
			name TEXT,
			email TEXT UNIQUE ,
	        hash TEXT,
	        role TEXT);`
*/
var insertSystemUser string = fmt.Sprintf(`
    INSERT INTO users (datecreated, datemodified, username, name, email, hash, role) VALUES 
    ('%s','%s','system', 'system', 'system@system.com', 'hash', 'root');
    `, times, times)

/*
const fieldsTable string = `

	CREATE TABLE IF NOT EXISTS fields(
	    id INTEGER PRIMARY KEY ,
	    postId INTEGER NOT NULL,
	    author TEXT,
	    authorId TEXT,
	    key TEXT,
	    data TEXT,
	    datecreated TEXT,
	    datemodified TEXT,
	    component TEXT,
	    tags TEXT,
	    parent INTEGER);`
*/
const insertDefaultTables string = `
    INSERT INTO tables (label) VALUES ('tables');
    INSERT INTO tables (label) VALUES ('fields');
    INSERT INTO tables (label) VALUES ('media');
    INSERT INTO tables (label) VALUES ('posts');
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
            DROP TABLE IF EXISTS posts;
            DROP TABLE IF EXISTS adminroutes;
            DROP TABLE IF EXISTS fields;
            DROP TABLE IF EXISTS media;
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

	statements := []string{tables, insertDefaultTables, userTable, adminRoutesTable, postsTable, fieldsTable, mediaTable}
	routes := []string{insertHomeRoute, insertPagesRoute, insertTypesRoute, insertFieldsRoute, insertMenusRoute, insertUsersRoute, insertMediaRoute, insertTestField}
	systemUser := []string{insertSystemUser}

	err := forEachStatement(db, statements,"tables")
	if err != nil {
		fmt.Printf("db exec err db_init 001 : %s", err)
	}
	err = forEachStatement(db, routes,"routes")
	if err != nil {
		fmt.Printf("db exec err db_init 002  : %s", err)
	}
	err = forEachStatement(db, systemUser,"systemUser")
	if err != nil {
		fmt.Printf("db exec err db_init 003 : %s", err)
	}
	return  err
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
