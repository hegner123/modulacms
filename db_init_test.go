package main

import (
	"fmt"
	"testing"
)


func TestDbInit(t *testing.T) {
    _,err := getDb(Database{DB: "modula_test.db"})
    if err!=nil {
        fmt.Printf("%s\n",err)
    }
    
}

func TestDbCreateTable(t *testing.T) {
    sql := userTable
    times := timestamp()
    user := User{DateCreated: times, DateModified: times, UserName:"test", Name:"test", Email: "test@test.com", Hash:"test", Role: "test"}
    
    res := formatCreateTable(user, "user")
    if  res != sql {
        t.Errorf("sql statement does not match. \nexpected %s\nwant %s ", sql, res)
    }
    fmt.Printf("%v", res)

}
