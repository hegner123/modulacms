package db

import (
	"context"
	"database/sql"
)

func (d Database) GetConnection() (*sql.DB, context.Context) {
	return d.Connection, d.Context
}
func (d MysqlDatabase) GetConnection() (*sql.DB, context.Context) {
	return d.Connection, d.Context
}
func (d PsqlDatabase) GetConnection() (*sql.DB, context.Context) {
	return d.Connection, d.Context
}
