package main

import (
	"fmt"
    "reflect"
	"testing"
)

func TestInsertFormat(t *testing.T) {

	field := Field{PostID: 4, Key: "link_url", Data: "https://example.com", Component: "link.html"}
	result, fieldsLength := formatInsertColumns(field)
	expected := "(4, 'link_url','https://example.com','link.html')"
	expectedLength := 10
	if result != expected && fieldsLength != int64(expectedLength) {
		t.Errorf("Reflect the user struct into SQL column syntax. \nLen: %d Result:%s \nWant\nLen:%d SQL:%s\n", fieldsLength, result, expectedLength, expected)
	}
}

func TestFieldQuery(t *testing.T) {
	post := Post{ID: 4}
    db,err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fields, err := getPostFields(post, db)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
    expected := Field{ID: 1, PostID: 4, Author: "system", AuthorID: "0", Key: "link_url", Data: "https://example.com", DateCreated:"1730634309", DateModified:"1730634309",Component: "link.html", Tags: "", Parent: ""}
    if !reflect.DeepEqual(fields[0],expected){
        t.Errorf("Struct Results not deeply equal.\n expected: %v+ \n res: %v+",expected,fields[0])
    }
}
