package main

import (
	"fmt"
	"testing"
)

func TestDbCreateTable(t *testing.T) {
    sql := userTableConst
    times := timestamp()
    user := User{DateCreated: times, DateModified: times, UserName:"test", Name:"test", Email: "test@test.com", Hash:"test", Role: "test"}
    
    res := formatCreateTable(user, "users")
    if  res != sql {
        t.Errorf("sql statement does not match. \nexpected %s\nwant %s ", sql, res)
    }
    fmt.Printf("%v", res)

}
func TestReflection(t *testing.T) {
	user := User{UserName: "example", Name: "example", Email: "example@mail.com", Hash: "laksdbgoiabjkb", Role: "admin", DateCreated: "1730634309", DateModified: "1730634309"}
	result, fieldsLength := formatSQLColumns(user)
	expected := "(id, dateCreated, dateModified, username, name, email, hash, role)"
	expectedLength := 8
	if result != expected && fieldsLength != int64(expectedLength) {
		t.Errorf("Reflect the user struct into SQL column syntax. \nLen: %d Result:%s \nWant\nLen:%d SQL:%s\n", fieldsLength, result, expectedLength, expected)
	}
}



