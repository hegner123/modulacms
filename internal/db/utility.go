package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	utility "github.com/hegner123/modulacms/internal/utility"
)

type ColumnNameType map[string]string
type ColumnIndexName map[int]string


func CopyDb(dbName string, useDefault bool) (string, error) {
	times := utility.TimestampS()
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
		srcSQLName = backup + "tests.sql"
	}

	dstCmd := exec.Command("sqlite3", dstDbName, ".read "+srcSQLName)
	_, err = dstCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command failed: %s\n", err)
	}

	fmt.Printf("dbName: %v\n", dbName)
	fmt.Printf("useDefault: %v\n", useDefault)
	fmt.Printf("times: %v\n", times)
	fmt.Printf("srcSQLName: %v\n", srcSQLName)
	fmt.Printf("dstDbName: %v\n", dstDbName)

	return dstDbName, nil
}

// GetTableColumns retrieves the columns and their types for a given table.
func GetTableColumns(ctx context.Context, db *sql.DB, tableName string) (ColumnNameType, ColumnIndexName, error) {
	// Optionally: validate tableName here to avoid SQL injection.
	if tableName == "" {
		return nil, nil, fmt.Errorf("table name cannot be empty")
	}

	// Construct the PRAGMA query.
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	// Execute the query using the provided context.
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Prepare a map to hold the column names and their types.
	columnsNT := make(ColumnNameType)
	columnsIN := make(ColumnIndexName)

	// PRAGMA table_info returns the following columns
	// cid, name, type, notnull, dflt_value, pk
	index := 0
	for rows.Next() {
		var cid int
		var name, colType string
		var notnull, pk int
		var dfltValue sql.NullString

		// Scan the row into local variables.
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Map the column name to its type.
		columnsNT[name] = colType
		columnsIN[index] = name
		index++

	}

	// Check for any errors encountered during iteration.
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	return columnsNT, columnsIN, nil
}

func GetColumnRowsString(dbc *sql.DB, ctx context.Context, tableName string, columnName string) ([]string, error) {
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
	rows, err := dbc.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Prepare a map to hold the column names and their types.
	rowValues := make([]string, 0)

	for rows.Next() {
		var value any
		var ap string

		// Scan the row into local variables.
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		s, ok := value.(string)
		ap = s
		if !ok {
			i, ok := value.(int64)
			if !ok {
				return nil, nil
			}
			ap = strconv.FormatInt(i, 10)
		}
		// Map the column name to its type.
		rowValues = append(rowValues, ap)
	}

	// Check for any errors encountered during iteration.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return rowValues, nil
}

