package db

import (
	"context"
	"database/sql"
	"encoding/json"

	utility "github.com/hegner123/modulacms/internal/Utility"
)

// Database Methods
func (d Database) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}
func (d MysqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}
func (d PsqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

func (d Database) Ping() error {
	// Ping the database to ensure a connection is established.
	if err := d.Connection.Ping(); err != nil {
		return err
	}
	return nil
}
func (d MysqlDatabase) Ping() error {
	// Ping the database to ensure a connection is established.
	if err := d.Connection.Ping(); err != nil {
		return err
	}
	return nil

}
func (d PsqlDatabase) Ping() error {
	// Ping the database to ensure a connection is established.
	if err := d.Connection.Ping(); err != nil {
		return err
	}
	return nil

}

//Struct Methods

func (a AdminContentData) GetHistory() string {
	return a.History.String
}

func (a AdminContentFields) GetHistory() string {
	return a.History.String
}

func (a AdminDatatypes) GetHistory() string {
	return a.History.String
}

func (a AdminFields) GetHistory() string {
	return a.History.String
}

func (a AdminRoutes) GetHistory() string {
	return a.History.String
}

func (c ContentData) GetHistory() string {
	return c.History.String
}

func (c ContentFields) GetHistory() string {
	return c.History.String
}

func (d Datatypes) GetHistory() string {
	return d.History.String
}

func (f Fields) GetHistory() string {
	return f.History.String
}

func (r Routes) GetHistory() string {
	return r.History.String
}

func (a AdminContentData) MapHistoryEntry() string {
	entry := AdminContentDataHistoryEntry{
		AdminContentDataID: a.AdminContentDataID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminContentFields) MapHistoryEntry() string {
	entry := AdminContentFieldsHistoryEntry{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminRouteID:        a.AdminRouteID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminDatatypes) MapHistoryEntry() string {
	entry := AdminDatatypesHistoryEntry{
		AdminDatatypeID: a.AdminDatatypeID,
		AdminRouteID:    a.AdminRouteID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		Author:          a.Author,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminFields) MapHistoryEntry() string {
	entry := AdminFieldsHistoryEntry{
		AdminFieldID: a.AdminFieldID,
		AdminRouteID: a.AdminRouteID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		Data:         a.Data,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminRoutes) MapHistoryEntry() string {
	entry := AdminRoutesHistoryEntry{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (c ContentData) MapHistoryEntry() string {
	entry := ContentDataHistoryEntry{
		ContentDataID: c.ContentDataID,
		RouteID:       c.RouteID,
		DatatypeID:    c.DatatypeID,
		DateCreated:   c.DateCreated,
		DateModified:  c.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (c ContentFields) MapHistoryEntry() string {
	entry := ContentFieldsHistoryEntry{
		ContentFieldID: c.ContentFieldID,
		RouteID:        c.RouteID,
		ContentDataID:  c.ContentDataID,
		FieldID:        c.FieldID,
		FieldValue:     c.FieldValue,
		DateCreated:    c.DateCreated,
		DateModified:   c.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (d Datatypes) MapHistoryEntry() string {
	entry := DatatypesHistoryEntry{
		DatatypeID:   d.DatatypeID,
		RouteID:      d.RouteID,
		ParentID:     d.ParentID,
		Label:        d.Label,
		Type:         d.Type,
		Author:       d.Author,
		AuthorID:     d.AuthorID,
		DateCreated:  d.DateCreated,
		DateModified: d.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (f Fields) MapHistoryEntry() string {
	entry := FieldsHistoryEntry{
		FieldID:      f.FieldID,
		RouteID:      f.RouteID,
		ParentID:     f.ParentID,
		Label:        f.Label,
		Data:         f.Data,
		Type:         f.Type,
		Author:       f.Author,
		AuthorID:     f.AuthorID,
		DateCreated:  f.DateCreated,
		DateModified: f.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (r Routes) MapHistoryEntry() string {
	entry := RoutesHistoryEntry{
		RouteID:      r.RouteID,
		Slug:         r.Slug,
		Title:        r.Title,
		Status:       r.Status,
		Author:       r.Author,
		AuthorID:     r.AuthorID,
		DateCreated:  r.DateCreated,
		DateModified: r.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}
