package main

import (
	"fmt"
	"testing"
)

func TestReflection(t *testing.T) {
	user := User{UserName: "example", Name: "example", Email: "example@mail.com", Hash: "laksdbgoiabjkb", Role: "admin", DateCreated: "1730634309", DateModified: "1730634309"}
	result, fieldsLength := formatSQLColumns(user)
	expected := "(username, name, email, hash, role, dateCreated, dateModified)"
	expectedLength := 7
	if result != expected && fieldsLength != int64(expectedLength) {
		t.Errorf("Reflect the user struct into SQL column syntax. \nLen: %d Result:%s \nWant\nLen:%d SQL:%s\n", fieldsLength, result, expectedLength, expected)
	}
}


func TestTableCreation(t *testing.T) {
    var media Media
    table := formatCreateTable(media, "media")
    fmt.Print(table)

}

/*
type User struct {
	ID       int
	UserName string
	Name     string
	Email    string
	Hash     string
	Role     string
}
*/
