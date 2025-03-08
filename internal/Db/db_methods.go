package db

import (
	"context"
	"database/sql"
)

func (d Database) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}
func (d MysqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}
func (d PsqlDatabase) GetConnection() (*sql.DB, context.Context, error) {
	return d.Connection, d.Context, nil
}

func (d Database) Ping() error {
	// Ping the database to ensure a connection is established.
	if err := d.Connection.Ping(); err != nil {
		return err
	}
	return nil
}
func (d MysqlDatabase) Ping() error {
	// Ping the database to ensure a connection is established.
	if err := d.Connection.Ping(); err != nil {
		return err
	}
	return nil

}
func (d PsqlDatabase) Ping() error {
	// Ping the database to ensure a connection is established.
	if err := d.Connection.Ping(); err != nil {
		return err
	}
	return nil

}
