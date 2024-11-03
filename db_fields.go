package main

import (
	"database/sql"
)

func getPostFields(postId Post, db *sql.DB) ([]Field, error) {
	var fields []Field

	rows, err := db.Query("SELECT * FROM fields WHERE postid = ?;", 4)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		field := Field{}
		// Only scan into the selected fields
		if err := rows.Scan(&field.ID, &field.PostID, &field.Author,
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
