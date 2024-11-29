package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

type GlobalTestingState struct {
	Initialized bool
	Db          *sql.DB
}

var globalTestingState GlobalTestingState

func setup() {
    resetFlag:=false
	fmt.Printf("TestMain setup\n")
	db, ctx, err := getDb(Database{DB: "./modula_test.db"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	globalTestingState.Initialized = true
	globalTestingState.Db = db

	err = initDb(db, ctx, &resetFlag , "modula_test.db")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
    //createSetupInserts(db,ctx)
   // insertPlaceholders(db, ctx, "")
}

func teardown() {
	fmt.Printf("TestMain teardown\n")
	globalTestingState.Initialized = false
	globalTestingState.Db.Close()
}

func TestMain(m *testing.M) {
	fmt.Printf("TestMain init\n")
	globalTestingState.Initialized = false
	setup()
	code := m.Run()
	teardown()
	fmt.Printf("TestMain exit\n")
	os.Exit(code)
}
