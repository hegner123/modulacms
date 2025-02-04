package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CopyDb(dbName string, useDefault bool) (string, error) {
	times := TimestampS()
	backup := "../../testdb/backups/"
	base := "../../testdb/"
	db := strings.TrimSuffix(dbName, ".db")
	srcSQLName := backup + db + ".sql"

	dstDbName := base + "testing" + times + dbName
	_, err := os.Create(dstDbName)
	if err != nil {
		fmt.Printf("Couldn't create file")
	}
	if useDefault {
		srcSQLName = backup + "test.sql"
	}

	dstCmd := exec.Command("sqlite3", dstDbName, ".read "+srcSQLName)
	_, err = dstCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command failed: %s\n", err)
	}

	if err != nil {
		return "", err
	}

	return dstDbName, nil
}

// GetTableColumns retrieves the columns and their types for a given table.
func GetTableColumns(ctx context.Context, db *sql.DB, tableName string) (map[string]string, error) {
	// Optionally: validate tableName here to avoid SQL injection.
	if tableName == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}

	// Construct the PRAGMA query.
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	// Execute the query using the provided context.
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Prepare a map to hold the column names and their types.
	columns := make(map[string]string)

	// PRAGMA table_info returns the following columns:
	// cid, name, type, notnull, dflt_value, pk
	for rows.Next() {
		var cid int
		var name, colType string
		var notnull, pk int
		var dfltValue sql.NullString

		// Scan the row into local variables.
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Map the column name to its type.
		columns[name] = colType
	}

	// Check for any errors encountered during iteration.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return columns, nil
}

func GetColumnRows(dbc Database, tableName string, columnName string) ([]any, error) {
	// Optionally: validate tableName here to avoid SQL injection.
	if tableName == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}
	if columnName == "" {
		return nil, fmt.Errorf("column name cannot be empty")
	}

	// Construct the PRAGMA query.
	query := fmt.Sprintf("SELECT %s FROM %s;", columnName, tableName)

	// Execute the query using the provided context.
	rows, err := dbc.Connection.QueryContext(dbc.Context, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Prepare a map to hold the column names and their types.
	rowValues := make([]any, 0)

	for rows.Next() {
		var value any

		// Scan the row into local variables.
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Map the column name to its type.
		rowValues = append(rowValues, value)
	}

	// Check for any errors encountered during iteration.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return rowValues, nil
}

func ReadNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	} else {
		return "NIll"
	}
}
