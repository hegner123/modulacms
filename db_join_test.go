package main

import (
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func TestRecursiveSetup(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to open database", err)
	}
	c := countUsers(db, ctx)
	pLog(c)
	logDb("modula_test.db")
	/*
		times := timestampS()
		entry1 := mdb.CreateDatatypeParams{
				Routeid:      ni(1),
				Label:        "About Card 1",
				Type:         "ui-element-1583",
				Author:       "system",
				Datecreated:  times,
				Datemodified: times,
			}
		    logStruct(entry1)
			entry2 := mdb.CreateDatatypeParams{
				Routeid:      ni(1),
				Parentid:     ni(1),
				Label:        "Card Heading",
				Type:         "ui-element-1584",
				Author:       "system",
				Authorid:     int64(1),
				Datecreated:  times,
				Datemodified: times,
			}
			entry3 := mdb.CreateDatatypeParams{
				Routeid:      ni(1),
				Parentid:     ni(1),
				Label:        "Card Body",
				Type:         "ui-element-1584",
				Author:       "system",
				Authorid:     int64(1),
				Datecreated:  times,
				Datemodified: times,
			}

			field1 := mdb.CreateFieldParams{
				Routeid:      ni(1),
				Parentid:     ni(2),
				Label:        "Heading",
				Data:         "About us",
				Type:         "ui-element-1585",
				Author:       "system",
				Authorid:     int64(1),
				Datecreated:  times,
				Datemodified: times,
			}
			field2 := mdb.CreateFieldParams{
				Routeid:      ni(1),
				Parentid:     ni(2),
				Label:        "Sub-Heading",
				Data:         "Meet the team",
				Type:         "ui-element-1585",
				Author:       "system",
				Authorid:     int64(1),
				Datecreated:  times,
				Datemodified: times,
			}
			field3 := mdb.CreateFieldParams{
				Routeid:  ni(1),
				Parentid: ni(3),
				Label:    "Image",
				Data: `{
		            Image: {
		                id:1
		            }
		        }`,
				Type:         "ui-element-1585",
				Author:       "system",
				Authorid:     int64(1),
				Datecreated:  times,
				Datemodified: times,
			}
			field4 := mdb.CreateFieldParams{
				Routeid:      ni(1),
				Parentid:     ni(3),
				Label:        "Body",
				Data:         `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`,
				Type:         "ui-element-1585",
				Author:       "system",
				Authorid:     int64(1),
				Datecreated:  times,
				Datemodified: times,
			}
	*/
	_, err = dbCreateDataType(db, ctx, entry1)
	if err != nil {
		logError("failed to create dataType1: ", err)
	}
	/*
		_, err = dbCreateDataType(db, ctx, entry2)
		if err != nil {
			logError("failed to create dataType2: ", err)
		}
		_, err = dbCreateDataType(db, ctx, entry3)
		if err != nil {
			logError("failed to create dataType3: ", err)
		}

		_, err = dbCreateField(db, ctx, field1)
		if err != nil {
			logError("failed to create field1: ", err)
		}
		_, err = dbCreateField(db, ctx, field2)
		if err != nil {
			logError("failed to create field2: ", err)
		}
		_, err = dbCreateField(db, ctx, field3)
		if err != nil {
			logError("failed to create field3: ", err)
		}
		_, err = dbCreateField(db, ctx, field4)
		if err != nil {
			logError("failed to create field5: ", err)
		}
	*/
}

func TestRecursiveJoinByRoute(t *testing.T) {
}
