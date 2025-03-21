package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)





func (d MysqlDatabase) CountContentData() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountContentFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountMedia() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountMediaDimensions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountPermissions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountRoles() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountRoutes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountTables() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountTokens() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountTokens(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CountUsers() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountUsers(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountSessions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountSessions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountUserOauths() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountUserOauths(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
