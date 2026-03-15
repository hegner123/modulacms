package dbmetrics

import (
	"database/sql"

	mysql "github.com/go-sql-driver/mysql"
	pq "github.com/lib/pq"
	sqlite3 "github.com/mattn/go-sqlite3"
)

func init() {
	sql.Register("sqlite3-metrics", &metricsDriver{
		inner:      &sqlite3.SQLiteDriver{},
		driverName: "sqlite",
	})
	sql.Register("mysql-metrics", &metricsDriver{
		inner:      &mysql.MySQLDriver{},
		driverName: "mysql",
	})
	sql.Register("postgres-metrics", &metricsDriver{
		inner:      &pq.Driver{},
		driverName: "postgres",
	})
}
