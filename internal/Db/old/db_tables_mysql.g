package db

import mdbm "github.com/hegner123/modulacms/db-mysql"

func (d MysqlDatabase) CreateAllTables() error {
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
	if err := d.CreateUserOauthTable(); err != nil {
		return err
	}
	if err := d.CreateSessionTable(); err != nil {
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

func (d MysqlDatabase) CreateContentDataTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateContentFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateMediaTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateMediaTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateMediaDimensionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateMediaDimensionTable(d.Context)
	return err
}
func (d MysqlDatabase) CreatePermissionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateRoleTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateRouteTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateRouteTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateTableTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateTokenTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateUserTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}
func (d MysqlDatabase) CreateSessionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}
func (d MysqlDatabase) CreateUserOauthTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserOauthTable(d.Context)
	return err
}
