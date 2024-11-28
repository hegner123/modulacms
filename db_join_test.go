package main

import (
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func TestRecursiveSetup(t *testing.T){
    db,ctx,err:=getDb(Database{DB: "modula_test.db"})
    if err != nil { 
        logError("failed to open database", err)
    }
    times := timestampS()
    entry1:=mdb.CreateDatatypeParams{
        Routeid: ni(1),
        Label: "About Card 1",
        Type: "ui-element-1583",
        Author: "system",
        Authorid: ni(1),
        Datecreated: times,
        Datemodified: times,
    }
    entry2:=mdb.CreateDatatypeParams{
        Routeid: ni(1),
        Parentid: ni(1),
        Label: "Card Heading",
        Type: "ui-element-1584",
        Author: "system",
        Authorid: ni(1),
        Datecreated: times,
        Datemodified: times,
    }
    entry3:=mdb.CreateDatatypeParams{
        Routeid: ni(1),
        Parentid: ni(1),
        Label: "Card Body",
        Type: "ui-element-1584",
        Author: "system",
        Authorid: ni(1),
        Datecreated: times,
        Datemodified: times,
    }
    
    field1:=mdb.CreateFieldParams{}
    field2:=mdb.CreateFieldParams{}
    field3:=mdb.CreateFieldParams{}
    field4:=mdb.CreateFieldParams{}

    dbCreateDataType(db,ctx,entry1)
    dbCreateDataType(db,ctx,entry2)
    dbCreateDataType(db,ctx,entry3)

    dbCreateField(db,ctx,field1)
    dbCreateField(db,ctx,field2)
    dbCreateField(db,ctx,field3)
    dbCreateField(db,ctx,field4)


}

func TestRecursiveJoinByRoute(t *testing.T){


}
