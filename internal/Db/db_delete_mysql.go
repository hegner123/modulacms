package db

import (
	_ "embed"
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)

func (d MysqlDatabase) DeleteAdminContentData(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Data: %v ", id)
	}

	return nil
}
func (d MysqlDatabase) DeleteAdminContentField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Field: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteAdminDatatype(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Datatype: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteAdminField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Field: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteAdminRoute(slug string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Route: %v ", slug)
	}

	return nil
}

func (d MysqlDatabase) DeleteContentData(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Content Data: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteContentField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Content Field: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteDatatype(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Datatype: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Field: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteMedia(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Media: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteMediaDimension(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete MediaDimension: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteRole(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRole(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Role: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteRoute(slug string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", slug)
	}

	return nil
}

func (d MysqlDatabase) DeleteTable(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteTable(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Table: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteToken(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteToken(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteUser(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}

	return nil
}
func (d MysqlDatabase) DeleteSession(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteSession(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) DeleteUserOauth(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}

	return nil
}
