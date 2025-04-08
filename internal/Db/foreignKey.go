package db

import (
	"database/sql"
	"fmt"

	utility "github.com/hegner123/modulacms/internal/utility"
)

type SqliteForeignKeyQueryRow struct {
	id        int
	seq       int
	tableName string
	fromCol   string
	toCol     string
	onUpdate  string
	onDelete  string
	match     string
}

const (
	sqliteQuery = "PRAGMA foreign_key_list('%s');"
	mysqlQuery  = `
    SELECT 
        TABLE_NAME, 
        COLUMN_NAME, 
        CONSTRAINT_NAME, 
        REFERENCED_TABLE_NAME, 
        REFERENCED_COLUMN_NAME
    FROM 
        INFORMATION_SCHEMA.KEY_COLUMN_USAGE
    WHERE 
        TABLE_SCHEMA = '%s' 
        AND TABLE_NAME = '%s'
        AND COLUMN_NAME = '%s'
        AND REFERENCED_TABLE_NAME IS NOT NULL;`

	psqlQuery = `
    SELECT
        tc.constraint_name, 
        tc.table_name, 
        kcu.column_name, 
        ccu.table_name AS foreign_table_name,
        ccu.column_name AS foreign_column_name 
    FROM 
        information_schema.table_constraints AS tc 
    JOIN 
        information_schema.key_column_usage AS kcu
          ON tc.constraint_name = kcu.constraint_name
          AND tc.table_schema = kcu.table_schema
    JOIN 
        information_schema.constraint_column_usage AS ccu
          ON ccu.constraint_name = tc.constraint_name
          AND ccu.table_schema = tc.table_schema
    WHERE 
        tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_name = '%s'
        AND kcu.column_name = '%s';`
)

func (d Database) GetForeignKeys(args []string) *sql.Rows {
	if len(args) != 1 {
		return nil
	}
	con, ctx, err := d.GetConnection()
	if err != nil {
		return nil
	}
	q := fmt.Sprintf(sqliteQuery, args[0])
	s, err := con.QueryContext(ctx, q)
	if err != nil {
		return nil
	}
	return s
}

func (d MysqlDatabase) GetForeignKeys(args []string) *sql.Rows {
	if len(args) != 3 {
		return nil
	}
	con, ctx, err := d.GetConnection()
	if err != nil {
		return nil
	}
	q := fmt.Sprintf(mysqlQuery, args[0], args[1], args[2])
	s, err := con.QueryContext(ctx, q)
	if err != nil {
		return nil
	}
	return s

}

func (d PsqlDatabase) GetForeignKeys(args []string) *sql.Rows {
	if len(args) != 3 {
		return nil
	}
	con, ctx, err := d.GetConnection()
	if err != nil {
		return nil
	}
	q := fmt.Sprintf(psqlQuery, args[0], args[1])
	s, err := con.QueryContext(ctx, q)
	if err != nil {
		return nil
	}
	return s

}

func (d Database) ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow {
	row := SqliteForeignKeyQueryRow{}
	foreignKeys := make([]SqliteForeignKeyQueryRow, 0)

	for rows.Next() {
		err := rows.Scan(&row.id, &row.seq, &row.tableName, &row.fromCol, &row.toCol, &row.onUpdate, &row.onDelete, &row.match)
		if err != nil {
			utility.DefaultLogger.Fatal("", err)
		}
		foreignKeys = append(foreignKeys, row)
	}

	return foreignKeys

}

func (d Database) SelectColumnFromTable(table string, column string) {
	t := StringDBTable(table)
	s, err := GenericList(t, d)
	if err != nil {
		return
	}
	for _, v := range s {
		utility.DefaultLogger.Info("", v[:len(v)-3])
	}
}

func (d MysqlDatabase) ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow {
	row := SqliteForeignKeyQueryRow{}
	foreignKeys := make([]SqliteForeignKeyQueryRow, 0)

	for rows.Next() {
		err := rows.Scan(&row.id, &row.seq, &row.tableName, &row.fromCol, &row.toCol, &row.onUpdate, &row.onDelete, &row.match)
		if err != nil {
			// handle error
			utility.DefaultLogger.Fatal("", err)
		}
		foreignKeys = append(foreignKeys, row)
	}

	return foreignKeys

}

func (d MysqlDatabase) SelectColumnFromTable(table string, column string) {
	t := StringDBTable(table)
	s, err := GenericList(t, d)
	if err != nil {
		return
	}
	for _, v := range s {
		utility.DefaultLogger.Info("", v[:len(v)-3])
	}
}

func (d PsqlDatabase) ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow {
	row := SqliteForeignKeyQueryRow{}
	foreignKeys := make([]SqliteForeignKeyQueryRow, 0)

	for rows.Next() {
		err := rows.Scan(&row.id, &row.seq, &row.tableName, &row.fromCol, &row.toCol, &row.onUpdate, &row.onDelete, &row.match)
		if err != nil {
			// handle error
			utility.DefaultLogger.Fatal("", err)
		}
		foreignKeys = append(foreignKeys, row)
	}

	return foreignKeys

}

func (d PsqlDatabase) SelectColumnFromTable(table string, column string) {
	t := StringDBTable(table)
	s, err := GenericList(t, d)
	if err != nil {
		return
	}
	for _, v := range s {
		utility.DefaultLogger.Info("", v[:len(v)-3])
	}
}
