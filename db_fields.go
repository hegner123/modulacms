package main

import (
	"database/sql"
	"fmt"
)

func dbGetRouteFields(routeID int, db *sql.DB) ([]Field, error) {
	var fields []Field
	var fieldElement string
	rows, err := db.Query("SELECT * FROM fields WHERE routeid = ?;", routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		field := Field{}
		// Only scan into the selected fields
		if err := rows.Scan(&field.ID, &field.RouteID, &field.Author,
			&field.AuthorID, &field.Key, &field.Type, &field.Data, &field.DateCreated,
			&field.DateModified, &fieldElement, &field.Tags, &field.Parent); err != nil {
			return nil, err
		}
		field.Component = parseFieldElement(fieldElement)
		fields = append(fields, field)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fields, nil

}

func dbCreateField(field Field) sql.Result {
	fmt.Print("create field\n")
	db, err := getDb(Database{})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	insertSatement := FormatSqlInsertStatement(field, "fields")
	fmt.Printf("Insert Statement %s\n", insertSatement)
	res, err := db.Exec(insertSatement)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return res

}
