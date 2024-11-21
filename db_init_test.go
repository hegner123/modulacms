package main

import (
	"embed"
	"fmt"
	"testing"

)

//go:embed sql/insert/*.sql
var f embed.FS

/*
func TestFS (t *testing.T){
    file,err:=sqlFiles.Open("insert/insert_media_dimension.sql")
    if err != nil {
        logError("failed to open embed fs file ", err)
    }
    db,err:=getDb(Database{DB: "modula_test.db"})
    if err != nil {
        logError("failed to creat connection to local db: ", err)
    }
    content, err := io.ReadAll(file)
    if err != nil {
        logError("failed to read contents: ", err)
    }
    query :=string(content)
    res,err := db.Exec(query,"test",100,100);
    if err != nil {
        logError("failed to insert row ", err)
    }
    _,err=res.LastInsertId()
    if err != nil {
        logError("test failed to find match ", err)
        t.Failed()
    }

}*/

func TestF(t *testing.T) {
	buffer, err := sqlFiles.ReadDir("sql/insert")
	if err != nil {
		t.Fatal()
		logError("failed to open embed fs file ", err)
	}
	for _, x := range buffer {
		fmt.Printf("%v\n", x.Name())
	}
}

func TestPreparedStatements(t *testing.T) {
	buffer, err := sqlFiles.ReadFile("sql/insert/insert_md.sql")
	if err != nil {
		t.Fatal()
		logError("failed to open embed fs file ", err)
	}
	db, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to get database ", err)
	}
    res,err := db.Exec(string(buffer),"test",100,100)
    if err != nil { 
        logError("failed to execute query ", err)
    }
    id,err :=res.LastInsertId()
    if err != nil { 
        logError("failed to retrieve last insert id ", err)
    }
    fmt.Printf("%v\n",id)

}
