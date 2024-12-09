package main

import (
	"fmt"
	"runtime"

)

func BuildTemplateStructFromRouteId(id int64) *Tree {
	dbName := "modula.db"
	db, ctx, err := getDb(Database{src: dbName})
	if err != nil {
		logError("failed to connect to db", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
	}
	defer db.Close()

    global := dbGetAdminDatatypeGlobalId(db, ctx)
	rows := dbGetChildren(make([]int64, 0), int(global.AdminDtID), dbName, "origin")

	t1 := NewTree(Scan)

	for i := 0; i < len(rows); i++ {
		dt := dbGetAdminDatatypeId(db, ctx, rows[i])
		t1.Add(dt)
	}
	t1.Root.AddTreeFields("")
	return t1
}
