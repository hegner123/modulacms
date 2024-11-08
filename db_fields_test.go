package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestInsertFormat(t *testing.T) {
	times := timestamp()
	var element = Element{Tag: "a", Attributes: map[string]string{"href": "/", "Value": "Click here"}}
	j, err := json.Marshal(element)
	if err != nil {
		fmt.Printf("\nError: %s\n", err)
	}
	fmt.Printf("json:%v\n", string(j))
	field := Field{RouteID: 4, Author: "system", AuthorID: "0", Key: "logo_link", Type: "Link", Data: "", DateCreated: times, DateModified: times, Component: element, Tags: "menu", Parent: "root"}
	sql := FormatSqlInsertStatement(field, "fields")
    result := dbCreateField(field)
    fmt.Printf("result: %v", result)
	expected := `INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, component, tags, parent) VALUES ('4', 'system', '0', 'logo_link', 'Link', '', '` + times + `', '` + times + `', '{"tag":"a","Attributes":{"Value":"Click here","href":"/"}}', 'menu', 'root');`
	if sql != expected {
		t.Errorf("Reflect the user struct into SQL column syntax. \nResult:%s \nWant\nSQL:%s\n", result, expected)
	}
}

func TestFieldQuery(t *testing.T) {
	route := Routes{ID: 0}
	db, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fields, err := dbGetRouteFields(route.ID, db)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	var element = Element{Tag: "a", Attributes: map[string]string{"href": "/", "Value": "Click here"}}
	expected := Field{ID: 1, RouteID: 4, Author: "system", AuthorID: "0", Key: "link_url", Data: "https://example.com", DateCreated: "1730634309", DateModified: "1730634309", Component: element, Tags: "", Parent: ""}
	if !reflect.DeepEqual(fields[0], expected) {
		t.Errorf("Struct Results not deeply equal.\n expected: %v+ \n res: %v+", expected, fields[0])
	}
}
