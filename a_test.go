package main

import "testing"


func TestInit(t *testing.T){
    db,err := getDb(Database{DB: "modula_test.db"})
    if err!=nil {
        return
    }
    initializeDatabase(db,true)
}
