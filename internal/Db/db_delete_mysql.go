package db

import (
	_ "embed"
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
	_ "github.com/mattn/go-sqlite3"
)

func (d MysqlDatabase) DeleteAdminDatatype(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteAdminField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteAdminRoute(slug string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", slug)
	}
	
	return nil
}

func (d MysqlDatabase) DeleteContentData(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete content data %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteContentField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete content field %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteDatatype(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete datatype %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Field %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteMedia(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Media %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteMediaDimension(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete MediaDimension %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteRole(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRole(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete  Route %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteRoute(slug string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("failed to delete  Route %v ", slug)
	}
	
	return nil
}

func (d MysqlDatabase) DeleteTable(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteTable(d.Context, int32(id))
	if err != nil {
		return err
	}
	
	return nil
}

func (d MysqlDatabase) DeleteToken(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteToken(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Token %v ", int32(id))
	}
	
	return nil
}

func (d MysqlDatabase) DeleteUser(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete User %v ", int32(id))
	}
	
	return nil
}
