package main

import (
	"fmt"
	"testing"
)

func TestInsertUser(t *testing.T) {
	_, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
func TestGetUser(t *testing.T) {
	_, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}

}
func TestAuthUser(t *testing.T) {
	_, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}

}
func TestUpdateUser(t *testing.T) {
	_, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}

}
func TestDeleteUser(t *testing.T) {
	_, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}

}
