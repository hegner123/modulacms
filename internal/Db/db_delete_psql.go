package db

import (
	_ "embed"
	"fmt"

	mdbp "github.com/hegner123/modulacms/db-psql"
)

func (d PsqlDatabase) DeleteAdminDatatype(id int64) error {
    
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteAdminField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteAdminRoute(slug string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("failed to delete Admin Route %v ", slug)
	}
	
	return nil
}

func (d PsqlDatabase) DeleteContentData(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete content data %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteContentField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete content field %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteDatatype(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete datatype %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Field %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteMedia(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Media %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteMediaDimension(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete MediaDimension %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteRole(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRole(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete  Route %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteRoute(slug string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRoute(d.Context, slug)
	if err != nil {
		return fmt.Errorf("failed to delete  Route %v ", slug)
	}
	
	return nil
}

func (d PsqlDatabase) DeleteTable(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteTable(d.Context, int32(id))
	if err != nil {
		return err
	}
	
	return nil
}

func (d PsqlDatabase) DeleteToken(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteToken(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete Token %v ", int32(id))
	}
	
	return nil
}

func (d PsqlDatabase) DeleteUser(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete User %v ", int32(id))
	}
	
	return nil
}
