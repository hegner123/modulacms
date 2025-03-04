package db

import (
	"fmt"

	mdbp "github.com/hegner123/modulacms/db-psql"
)

func (d PsqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CountAdminFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CountAdminRoutes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminroute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CountContentData() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CountContentFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CountFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CountMedia() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CountMediaDimensions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CountRoles() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CountRoutes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CountTables() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountTables(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CountTokens() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountTokens(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CountUsers() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountUsers(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
