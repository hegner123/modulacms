package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func initializeDatabase(reset bool) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./modula.db")
	if err != nil {
		return nil, err
	}
	if reset {
		res, err := db.Exec(`
            DROP TABLE IF EXISTS users;
            DROP TABLE IF EXISTS posts;
            DROP TABLE IF EXISTS fields;
            DROP TABLE IF EXISTS media;
            `)
		if err != nil {
			log.Fatal("I CAN'T FIND THE DATABASE CAPTIN!!!!\n Oh GOD IT'S GOT MY LEG!!!!")
		}
		if res != nil {
			log.Print(res)
		}
	}

	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT,
		name TEXT,
		email TEXT UNIQUE,
        hash TEXT,
        role TEXT
	);`

	postsTable := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL,
		title TEXT,
		status INTEGER NOT NULL,
        dateCreated TEXT NOT NULL,
        dateModified TEXT NOT NULL,
        content TEXT NOT NULL,
        template TEXT NOT NULL 
	);`

	contentTable := `
    CREATE TABLE IF NOT EXISTS fields(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        postId INTEGER NOT NULL,
        data TEXT);`

	mediaTable := `
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
    optimizedUltrawide TEXT
);
`

	statements := [...]string{userTable, postsTable, contentTable, mediaTable}

	for i := 0; i < len(statements); i++ {
		_, err := db.Exec(statements[i])
		if err != nil {
			return db, err
		}
	}
	test := User{UserName: "triXXy", Name: "Alice", Email: "alice@example.com", Hash: "ajsdbfkjhe8b86v0s", Role: "Admin"}
	fmt.Print(queryCreateBuilder(test, "users"))

	if reset {

		seedDB(db)
        registerAdminRoutes(db)
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

		//user.Name = "Alice Updated"
		//if err := updateUserById(db, user); err != nil {
		//  log.Fatal("Error updating user:", err)
		//}
		//fmt.Println("User updated successfully")

		//if err := deleteUserById(db, int(userID)); err != nil {
		//	log.Fatal("Error deleting user:", err)
		//}
		//fmt.Println("User deleted successfully")
	}
	return db, err
}
