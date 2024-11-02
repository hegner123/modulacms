package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
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
        data TEXT,
        dateCreated TEXT,
        dateModified TEXT,
        component TEXT,
        tags TEXT,
        parent INTEGER);`

const tables string = `
    CREATE TABLE IF NOT EXISTS tables (
    id INTEGER PRIMARY KEY,
    label TEXT);
    `
const insertDefaultTables string = `
    INSERT INTO tables (label) VALUES ('fields');
    INSERT INTO tables (label) VALUES ('media');
    INSERT INTO tables (label) VALUES ('posts');
    INSERT INTO tables (label) VALUES ('users');
    `

func getDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./modula.db")
	if err != nil {
		return nil, err
	}
    _, err = db.Exec("PRAGMA foreign_keys = ON;")
    if err!=nil {
        return nil, err
    }
    return db, nil
}

func initializeDatabase(reset bool) (*sql.DB, error) {
	db, err := getDb()
	if err != nil {
		return nil, err
	}
	if reset {
		res, err := db.Exec(`
            DROP TABLE IF EXISTS users;
            DROP TABLE IF EXISTS posts;
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

	statements := [...]string{tables, insertDefaultTables, userTable, postsTable, fieldsTable, mediaTable}

	for i := 0; i < len(statements); i++ {
		_, err := db.Exec(statements[i])
		if err != nil {
			return db, err
		}
	}

	if reset {

		seedDB(db)
		newUser := User{UserName: "admin", Name: "Admin", Email: "admin@admin.com", Hash: "", Role: "0"}
		userID, err := createUser(db, newUser)
		if err != nil {
			log.Fatal("Error creating user:", err)
		}

		fmt.Printf("New user ID: %d\n", userID)

		user, err := getUserById(db, int(userID))
		if err != nil {
			log.Fatal("Error fetching user:", err)
		}
		fmt.Printf("Fetched User: %+v\n", user)

		_, err = createPost(db, Post{Slug: "blog", Title: "Blog", Status: 0, DateCreated: time.Now().Unix(), DateModified: time.Now().Unix(), Content: "hello", Template: "page"})
		if err != nil {
			log.Fatal("I CAN'T MAKE THE POST CAPTIN!!!")
		}

	}
	return db, err
}
