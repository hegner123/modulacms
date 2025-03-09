package db

import (
	_ "embed"
	"fmt"

	mdbp "github.com/hegner123/modulacms/db-psql"
)

func (d PsqlDatabase) DeleteAdminContentData(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Data: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) DeleteAdminContentField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Field: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteAdminDatatype(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Datatype: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteAdminField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Field: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteAdminRoute(slug string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Route: %v ", slug)
	}

	return nil
}

func (d PsqlDatabase) DeleteContentData(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Content Data: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteContentField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Content Field: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteDatatype(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Datatype: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Field: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteMedia(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Media: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteMediaDimension(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete MediaDimension: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteRole(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRole(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Role: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteRoute(slug string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("Failed to Delete Route: %v ", slug)
	}

	return nil
}

func (d PsqlDatabase) DeleteTable(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteTable(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Table: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteToken(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteToken(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteUser(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) DeleteSession(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteSession(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}

	return nil
}

func (d PsqlDatabase) DeleteUserOauth(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}

	return nil
}
