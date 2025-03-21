package db

import mdbp "github.com/hegner123/modulacms/db-psql"

func (d PsqlDatabase) CreateAllTables() error {
	if err := d.CreatePermissionTable(); err != nil {
		return err
	}
	if err := d.CreateRoleTable(); err != nil {
		return err
	}
	if err := d.CreateUserTable(); err != nil {
		return err
	}
	if err := d.CreateAdminRouteTable(); err != nil {
		return err
	}
	if err := d.CreateRouteTable(); err != nil {
		return err
	}
	if err := d.CreateAdminDatatypeTable(); err != nil {
		return err
	}
	if err := d.CreateAdminFieldTable(); err != nil {
		return err
	}
	if err := d.CreateContentDataTable(); err != nil {
		return err
	}
	if err := d.CreateAdminContentDataTable(); err != nil {
		return err
	}
	if err := d.CreateDatatypeTable(); err != nil {
		return err
	}
	if err := d.CreateFieldTable(); err != nil {
		return err
	}
	if err := d.CreateTokenTable(); err != nil {
		return err
	}
	if err := d.CreateContentFieldTable(); err != nil {
		return err
	}
	if err := d.CreateAdminContentFieldTable(); err != nil {
		return err
	}
	if err := d.CreateTableTable(); err != nil {
		return err
	}
	if err := d.CreateMediaTable(); err != nil {
		return err
	}
	if err := d.CreateMediaDimensionTable(); err != nil {
		return err
	}
	return nil
}


func (d PsqlDatabase) CreateContentDataTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateContentFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateMediaTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateMediaTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateMediaDimensionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateMediaDimensionTable(d.Context)
	return err
}

func (d PsqlDatabase) CreatePermissionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateRoleTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateRouteTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateRouteTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateTableTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateTokenTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateUserTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateSessionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateUserOauthTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserOauthTable(d.Context)
	return err
}
