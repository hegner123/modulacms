package main

import (
	"fmt"
	"reflect"
	"testing"
)

var times = timestamp()
var element = Element{Tag: "div", Attributes: map[string]string{"test": "test"}}
var test_field Field = Field{RouteID: 4, Author: "system", AuthorID: "0", Key: "logo_link", Type: "Link", Data: "", DateCreated: times, DateModified: times, Component: element, Tags: "menu", Parent: "root"}

func TestInsertField(t *testing.T) {
	db, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	element := Element{Tag: "div", Attributes: map[string]string{"test": "test"}}
	times := timestamp()
	field := Field{RouteID: 4, Author: "system", AuthorID: "0", Key: "logo_link", Type: "Link", Data: "", DateCreated: times, DateModified: times, Component: element, Tags: "menu", Parent: "root"}
	err = dbInsertField(db, field)
	if err != nil {
		t.Errorf("Reflect the user struct into SQL column syntax. \nResult:%s \nWant\n", err)
	} else {
		t.Name()
		t.Log("PASS")
		fmt.Printf("Test InsertField PASS\n")
	}
}

func TestFieldQuery(t *testing.T) {
	route := Routes{ID: 4}
	db, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fields, err := dbGetField(db, route.ID)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	var expected = test_field
	expected.ID = 4
	expected.Component.ID = 4
	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("Struct Results not deeply equal.\n expected: %v+ \n res: %v+", expected, fields)
	} else {
		t.Name()
		t.Log("PASS")
		fmt.Printf("Test get Field PASS\n")
	}
}
