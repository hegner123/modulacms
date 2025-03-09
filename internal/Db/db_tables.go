package db

import mdb "github.com/hegner123/modulacms/db-sqlite"

func (d Database) CreateAllTables() error {
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
	if err := d.CreateSessionTable(); err != nil {
		return err
	}
	if err := d.CreateUserOauthTable(); err != nil {
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
func (d Database) CreateAdminContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}

func (d Database) CreateAdminContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}

func (d Database) CreateAdminDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d Database) CreateAdminFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

func (d Database) CreateAdminRouteTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminRouteTable(d.Context)
	return err
}

func (d Database) CreateContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d Database) CreateContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d Database) CreateDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d Database) CreateFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}

func (d Database) CreateMediaTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateMediaTable(d.Context)
	return err
}

func (d Database) CreateMediaDimensionTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateMediaDimensionTable(d.Context)
	return err
}

func (d Database) CreateRoleTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d Database) CreateRouteTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateRouteTable(d.Context)
	return err
}

func (d Database) CreateTableTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateTablesTable(d.Context)
	return err
}

func (d Database) CreateTokenTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

func (d Database) CreateUserTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

func (d Database) CreateSessionTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

func (d Database) CreateUserOauthTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserOauthTable(d.Context)
	return err
}
