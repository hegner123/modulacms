package db

import (
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)


func (d Database) DeleteContentData(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentData(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Content Data: %v ", id)
	}

	return nil
}

func (d Database) DeleteContentField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentField(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Content Field: %v ", id)
	}

	return nil
}

func (d Database) DeleteDatatype(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Datatype: %v ", id)
	}

	return nil
}

func (d Database) DeleteField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteField(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Field: %v ", id)
	}

	return nil
}

func (d Database) DeleteMedia(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Media: %v ", id)
	}

	return nil
}

func (d Database) DeleteMediaDimension(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete MediaDimension: %v ", id)
	}

	return nil
}

func (d Database) DeletePermission(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeletePermission(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Permission: %v ", id)
	}

	return nil
}

func (d Database) DeleteRole(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRole(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Role: %v ", id)
	}

	return nil
}

func (d Database) DeleteRoute(slug string) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", slug)
	}

	return nil
}

func (d Database) DeleteTable(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteTable(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Table: %v ", id)
	}

	return nil
}

func (d Database) DeleteToken(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteToken(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}

	return nil
}

func (d Database) DeleteUser(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUser(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}

	return nil
}

func (d Database) DeleteSession(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteSession(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}

	return nil
}

func (d Database) DeleteUserOauth(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}

	return nil
}
