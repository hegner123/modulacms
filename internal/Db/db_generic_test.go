package db

import (
	"fmt"
	"testing"
)

func TestFormatSqlColumns(t *testing.T) {
	source1 := []string{"test1", "test2", "test3"}
	source2 := []string{"test1"}
	s1 := FormatSqlColumns(source1, false)
	s2 := FormatSqlColumns(source2, false)
	e1 := "test1,test2,test3"
	e2 := "test1"

	if s1 != e1 {
		fmt.Printf("\nSource1 test: %s\n", s1)
		t.Fatal("s1 fail")
	}
	if s2 != e2 {
		fmt.Printf("\nSource2 test: %s\n", s2)
		t.Fatal("s2 fail")
	}

}
/*
func TestFormatSqlFilter(t *testing.T) {

}
*/

func TestInsertQuery(t *testing.T) {
	table := "users"
	columnSlice := []string{"username", "email", "hash", "created", "2fa"}
	valueSlice := []string{"petty76", "email@email.com", "89a7s6fdf69ss86f9f690e87efhf", "2025-01-15", "0"}
	e1 := "INSERT INTO users (username,email,hash,created,2fa) VALUES ('petty76','email@email.com','89a7s6fdf69ss86f9f690e87efhf','2025-01-15',0);"
	columns := FormatSqlColumns(columnSlice, false)
	values := FormatSqlColumns(valueSlice, true)
	q := InsertQuery(table, columns, values)
	fmt.Printf("\nColumns: %v\n", columns)
	fmt.Printf("\nValues: %v\n", values)
	fmt.Printf("\nQuery: %v\n", q)
	fmt.Printf("\nExpected: %v\n", e1)

	if q != e1 {
		t.Fail()
	}

}
func TestSelectQuery(t *testing.T) {
	var fs []WhereKeyValue
	table := "users"
	columnSlice := []string{"username", "email", "hash", }
	f1 := WhereKeyValue{key: "username", value: "petty76"}
	f2 := WhereKeyValue{key: "email", value: "email@email.com", method: &and}
	fs = append(fs, f1)
	fs = append(fs, f2)
	e1 := "SELECT (username,email,hash) FROM users WHERE username='petty76' AND email='email@email.com';"
	columns := FormatSqlColumns(columnSlice, false)
	filter := FormatSqlFU(fs,false)

	q := SelectQuery(table, columns, filter)
	fmt.Printf("\nColumns: %v\n", columns)
	fmt.Printf("\nQuery: %v\n", q)
	fmt.Printf("\nExpected: %v\n", e1)

	if q != e1 {
		t.Fail()
	}

}
