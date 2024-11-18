package main

import "testing"

func TestCreateRoutes(t *testing.T) {
	db, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to create database dump in archive: ", err)
	}
	r := Routes{Author: "system", AuthorID: "0", Slug: "/test", Title: "TestRoute",Status: 0, DateCreated: times, DateModified: times,Content: "page",Template: "page.html"}
	res, err := dbCreateRoute(db,r)
    if err!=nil {
        return
    }
    if res == int64(1){
        t.Failed()
    }

}
