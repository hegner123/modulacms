package main

import (
	"database/sql"
	"fmt"
)

func getRouteFields(routeId Routes, db *sql.DB) ([]Field, error) {
	var fields []Field

	rows, err := db.Query("SELECT * FROM fields WHERE routeid = ?;", 4)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		field := Field{}
		// Only scan into the selected fields
		if err := rows.Scan(&field.ID, &field.RouteID, &field.Author,
			&field.AuthorID, &field.Key, &field.Data, &field.DateCreated,
			&field.DateModified, &field.Component, &field.Tags, &field.Parent); err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fields, nil

}

func dbCreateField(field Field) string {
    fmt.Print("create field")
	db, err := getDb(Database{})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	insertSatement := FormatSqlInsertStatement(field, "fields")
    fmt.Printf("Insert Statement %s",insertSatement)
	res, err := db.Exec(insertSatement)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	id, err := res.LastInsertId()
    if err!=nil {
        fmt.Printf("%s\n",err)
    }

	if rows < 1 {
		return "Insert Failed"
	} else {
		return fmt.Sprintf("Successfully created field as %v", id)
	}

}
