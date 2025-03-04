package db

import (
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func (d Database) DeleteAdminDatatype(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", id)
	}

	return nil
}

func (d Database) DeleteAdminField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", id)
	}

	return nil
}

func (d Database) DeleteAdminRoute(slug string) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", slug)
	}

	return nil
}

func (d Database) DeleteContentData(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentData(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete content data %v ", id)
	}

	return nil
}

func (d Database) DeleteContentField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentField(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete content field %v ", id)
	}

	return nil
}

func (d Database) DeleteDatatype(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete datatype %v ", id)
	}

	return nil
}

func (d Database) DeleteField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteField(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("failed to delete Field %v ", id)
	}

	return nil
}

func (d Database) DeleteMedia(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("failed to delete Media %v ", id)
	}

	return nil
}

func (d Database) DeleteMediaDimension(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("failed to delete MediaDimension %v ", id)
	}

	return nil
}

func (d Database) DeleteRole(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRole(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete  Route %v ", id)
	}

	return nil
}

func (d Database) DeleteRoute(slug string) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("failed to delete  Route %v ", slug)
	}

	return nil
}

func (d Database) DeleteTable(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteTable(d.Context, id)
	if err != nil {
		return err
	}

	return nil
}

func (d Database) DeleteToken(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteToken(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete Token %v ", id)
	}

	return nil
}

func (d Database) DeleteUser(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUser(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete User %v ", id)
	}

	return nil
}
