package main

import (
	"fmt"
	"runtime"
	"testing"
)

var TreeTestTable string

func TestTreeDBCopy(t *testing.T) {
	testTable, err := createDbCopy("tree.db", true)
	if err != nil {
		logError("failed to create copy of the database, I have to hurry, I'm running out of time!!! ", err)
		t.FailNow()
	}
	TreeTestTable = testTable
}

func TestTree(t *testing.T) {
	db, ctx, err := getDb(Database{src: TreeTestTable})
	if err != nil {
		logError("failed to connect to db", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
	}
	defer db.Close()
	rows := dbGetChildren(make([]int64, 0), 1, TreeTestTable, "origin")

	t1 := NewTree(Scan)

	for i := 0; i < len(rows); i++ {
		dt := dbGetAdminDatatypeId(db, ctx, rows[i])
		t1.Add(dt)
	}
    t1.Root.AddTreeFields("")
    t1.Root.PrintTree(0)
}
