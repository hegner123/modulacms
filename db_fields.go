package main

import (
	"database/sql"
)

func dbGetField(db *sql.DB, fieldID int) (Field, error) {
    var field Field

    // Query fields table
    fieldQuery := `
        SELECT id, routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent
        FROM fields WHERE id = ?`
    row := db.QueryRow(fieldQuery, fieldID)
    var componentID int
    err := row.Scan(&field.ID, &field.RouteID, &field.Author, &field.AuthorID, &field.Key, &field.Type,
        &field.Data, &field.DateCreated, &field.DateModified, &componentID, &field.Tags, &field.Parent)
    if err != nil {
        return field, err
    }

    // Query elements table
    elementQuery := `SELECT tag FROM elements WHERE id = ?`
    var element Element
    element.ID = componentID
    err = db.QueryRow(elementQuery, componentID).Scan(&element.Tag)
    if err != nil {
        return field, err
    }

    // Query attributes table
    attrQuery := `SELECT key, value FROM attributes WHERE elementid = ?`
    rows, err := db.Query(attrQuery, componentID)
    if err != nil {
        return field, err
    }
    defer rows.Close()

    element.Attributes = make(map[string]string)
    for rows.Next() {
        var key, value string
        if err := rows.Scan(&key, &value); err != nil {
            return field, err
        }
        element.Attributes[key] = value
    }

    field.Component = element
    return field, nil
}

func dbInsertField(db *sql.DB, field Field) error {
    // Insert into elements table
    elementQuery := `INSERT INTO elements (tag) VALUES (?)`
    res, err := db.Exec(elementQuery, field.Component.Tag)
    if err != nil {
        return err
    }
    elementID, err := res.LastInsertId()
    if err != nil {
        return err
    }

    // Insert into attributes table
    attrQuery := `INSERT INTO attributes (elementid, key, value) VALUES (?, ?, ?)`
    for key, value := range field.Component.Attributes {
        _, err := db.Exec(attrQuery, elementID, key, value)
        if err != nil {
            return err
        }
    }

    // Insert into fields table
    fieldQuery := `
        INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    _, err = db.Exec(fieldQuery, field.RouteID, field.Author, field.AuthorID, field.Key, field.Type,
        field.Data, field.DateCreated, field.DateModified, elementID, field.Tags, field.Parent)
    return err
}

