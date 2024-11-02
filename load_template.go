package main

import (
	"database/sql"
)

func getPostFields(postId Post, db *sql.DB) ([]Field, error) {
	var fields []Field

	// Query only the fields we need (slug, title, and template)
	rows, err := db.Query("SELECT slug, title, template FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		field := Field{}
		// Only scan into the selected fields
		if err := rows.Scan(&field.Data, &field.Parent); err != nil {
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
