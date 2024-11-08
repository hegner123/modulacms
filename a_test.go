package main

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	db, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		return
	}
	err = initializeDatabase(db, true)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
