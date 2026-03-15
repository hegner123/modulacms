package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// ColumnMeta describes a single column from database catalog introspection.
type ColumnMeta struct {
	Name      string
	IsInteger bool // INTEGER, INT, BIGINT, SMALLINT, TINYINT, SERIAL
}

// FKViolation describes a single foreign key constraint violation.
type FKViolation struct {
	Table   string
	RowID   string
	Parent  string
	FKIndex int
}

// DeployOps defines low-level database operations used exclusively by the
// deploy sync engine. These bypass auditing intentionally — the deploy
// orchestrator records its own audit trail via SyncManifest/SyncResult.
//
// Implementations are backend-specific because FK toggling, truncation,
// bulk insertion, and FK verification use different SQL across SQLite,
// MySQL, and PostgreSQL.
type DeployOps interface {
	// ImportAtomic runs fn inside a backend-appropriate atomic context:
	//   - SQLite:     pinned *sql.Conn with PRAGMA foreign_keys=OFF, wrapped in a transaction
	//   - PostgreSQL: transaction with session_replication_role='replica'
	//   - MySQL:      transaction with FOREIGN_KEY_CHECKS=0 (uses DELETE FROM, not TRUNCATE)
	//
	// If fn returns an error the entire import is rolled back (SQLite/PostgreSQL)
	// or left partially applied (MySQL TRUNCATE — callers should pre-snapshot).
	ImportAtomic(ctx context.Context, fn ImportFunc) error

	// TruncateTable removes all rows from table. Must be called inside ImportAtomic.
	// The table name is validated against allTables before execution.
	TruncateTable(ctx context.Context, ex Executor, table DBTable) error

	// BulkInsert writes rows into table using multi-row INSERT statements.
	// Automatically batches based on the backend's parameter limit and column count.
	// Must be called inside ImportAtomic.
	BulkInsert(ctx context.Context, ex Executor, table DBTable, columns []string, rows [][]any) error

	// VerifyForeignKeys checks all FK constraints and returns any violations.
	// For SQLite this uses PRAGMA foreign_key_check.
	// For MySQL/PostgreSQL this generates LEFT JOIN queries from FK metadata.
	VerifyForeignKeys(ctx context.Context, ex Executor) ([]FKViolation, error)

	// Placeholder returns the parameter placeholder for the nth parameter (1-based).
	// SQLite/MySQL return "?", PostgreSQL returns "$n".
	Placeholder(n int) string

	// IntrospectColumns queries the database catalog for column names and types.
	// Returns an error if the table does not exist.
	IntrospectColumns(ctx context.Context, table DBTable) ([]ColumnMeta, error)

	// QueryAllRows returns all rows from a table as column names and untyped values.
	// Integer columns (detected via IntrospectColumns) are scanned as int64;
	// all others are scanned as string (or nil for NULL).
	QueryAllRows(ctx context.Context, table DBTable) ([]string, [][]any, error)
}

// ImportFunc is the callback executed inside ImportAtomic.
// The Executor is either a *sql.Tx or a txOnConn wrapper — callers
// should use it for all SQL operations to stay on the correct
// connection and transaction.
//
// Executor is defined in query_builder.go and provides ExecContext/QueryContext.
type ImportFunc func(ctx context.Context, ex Executor) error

// Backend parameter limits for multi-row INSERT batching.
const (
	// sqliteMaxVars is the conservative SQLite SQLITE_LIMIT_VARIABLE_NUMBER.
	// Newer builds (3.32.0+) raise this to 32766, but 999 is the safe floor
	// for any SQLite linked by mattn/go-sqlite3.
	sqliteMaxVars = 999

	// psqlMaxVars is PostgreSQL's practical parameter limit (uint16 max).
	psqlMaxVars = 65535

	// mysqlDefaultBatchRows is a row-count ceiling for MySQL whose real
	// constraint is max_allowed_packet (byte-based, not var-count).
	// 1000 rows per batch keeps packet sizes well under the 4 MB default.
	mysqlDefaultBatchRows = 1000
)

// batchSize returns the maximum number of rows per INSERT batch given the
// column count and the backend's parameter variable limit.
func batchSize(numColumns, maxVars int) int {
	if numColumns <= 0 {
		return 0
	}
	bs := maxVars / numColumns
	if bs <= 0 {
		bs = 1
	}
	return bs
}

// ---------- SQLite ----------

type sqliteDeployOps struct {
	pool *sql.DB
}

// ImportAtomic pins a single connection from the pool so that
// PRAGMA foreign_keys=OFF is guaranteed to apply to every subsequent
// statement, then wraps the work in a transaction for rollback safety.
func (s *sqliteDeployOps) ImportAtomic(ctx context.Context, fn ImportFunc) error {
	conn, err := s.pool.Conn(ctx)
	if err != nil {
		return fmt.Errorf("pin sqlite connection: %w", err)
	}
	defer conn.Close()

	// Disable FK checks on this specific connection.
	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys=OFF;"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}

	// Begin transaction on the pinned connection.
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // no-op after commit

	if err := fn(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Re-enable FK checks.
	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys=ON;"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	return nil
}

func (s *sqliteDeployOps) TruncateTable(ctx context.Context, ex Executor, table DBTable) error {
	if !IsValidTable(table) {
		return fmt.Errorf("truncate: unknown table %q", string(table))
	}
	// SQLite has no TRUNCATE — use DELETE FROM.
	_, err := ex.ExecContext(ctx, "DELETE FROM "+string(table)+";")
	if err != nil {
		return fmt.Errorf("truncate %s: %w", table, err)
	}
	return nil
}

func (s *sqliteDeployOps) BulkInsert(ctx context.Context, ex Executor, table DBTable, columns []string, rows [][]any) error {
	if !IsValidTable(table) {
		return fmt.Errorf("bulk insert: unknown table %q", string(table))
	}
	if len(rows) == 0 {
		return nil
	}

	bs := batchSize(len(columns), sqliteMaxVars)
	colList := strings.Join(columns, ", ")

	for i := 0; i < len(rows); i += bs {
		end := i + bs
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[i:end]

		placeholderRow := "(" + strings.Repeat("?, ", len(columns)-1) + "?)"
		allPlaceholders := strings.Repeat(placeholderRow+", ", len(chunk)-1) + placeholderRow

		query := "INSERT INTO " + string(table) + " (" + colList + ") VALUES " + allPlaceholders + ";"

		args := make([]any, 0, len(chunk)*len(columns))
		for _, row := range chunk {
			args = append(args, row...)
		}

		if _, err := ex.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("bulk insert %s (batch starting at row %d): %w", table, i, err)
		}
	}
	return nil
}

func (s *sqliteDeployOps) Placeholder(_ int) string { return "?" }

func (s *sqliteDeployOps) VerifyForeignKeys(ctx context.Context, ex Executor) ([]FKViolation, error) {
	rows, err := ex.QueryContext(ctx, "PRAGMA foreign_key_check;")
	if err != nil {
		return nil, fmt.Errorf("foreign key check: %w", err)
	}
	defer rows.Close()

	var violations []FKViolation
	for rows.Next() {
		var v FKViolation
		if err := rows.Scan(&v.Table, &v.RowID, &v.Parent, &v.FKIndex); err != nil {
			return nil, fmt.Errorf("scan FK violation: %w", err)
		}
		violations = append(violations, v)
	}
	return violations, rows.Err()
}

func (s *sqliteDeployOps) IntrospectColumns(ctx context.Context, table DBTable) ([]ColumnMeta, error) {
	if !IsValidTable(table) {
		return nil, fmt.Errorf("introspect: unknown table %q", string(table))
	}
	rows, err := s.pool.QueryContext(ctx, "PRAGMA table_info("+string(table)+");")
	if err != nil {
		return nil, fmt.Errorf("introspect %s: %w", table, err)
	}
	defer rows.Close()

	var cols []ColumnMeta
	for rows.Next() {
		var cid int
		var name, colType string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("scan column info for %s: %w", table, err)
		}
		upper := strings.ToUpper(colType)
		isInt := strings.Contains(upper, "INT") || upper == "SERIAL"
		cols = append(cols, ColumnMeta{Name: name, IsInteger: isInt})
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q does not exist or has no columns", string(table))
	}
	return cols, rows.Err()
}

func (s *sqliteDeployOps) QueryAllRows(ctx context.Context, table DBTable) ([]string, [][]any, error) {
	if !IsValidTable(table) {
		return nil, nil, fmt.Errorf("query all rows: unknown table %q", string(table))
	}
	colMeta, err := s.IntrospectColumns(ctx, table)
	if err != nil {
		return nil, nil, err
	}
	return queryAllRowsGeneric(ctx, s.pool, table, colMeta)
}

// ---------- PostgreSQL ----------

type psqlDeployOps struct {
	pool *sql.DB
}

// ImportAtomic starts a transaction and disables FK constraint checking
// via session_replication_role. PostgreSQL's TRUNCATE is transactional
// so the entire import rolls back cleanly on error.
func (p *psqlDeployOps) ImportAtomic(ctx context.Context, fn ImportFunc) error {
	tx, err := p.pool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Disable FK triggers for this transaction's session.
	if _, err := tx.ExecContext(ctx, "SET session_replication_role = 'replica';"); err != nil {
		return fmt.Errorf("disable FK triggers: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		return err
	}

	// Re-enable FK triggers before commit.
	if _, err := tx.ExecContext(ctx, "SET session_replication_role = 'origin';"); err != nil {
		return fmt.Errorf("enable FK triggers: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (p *psqlDeployOps) TruncateTable(ctx context.Context, ex Executor, table DBTable) error {
	if !IsValidTable(table) {
		return fmt.Errorf("truncate: unknown table %q", string(table))
	}
	_, err := ex.ExecContext(ctx, "TRUNCATE TABLE "+string(table)+" CASCADE;")
	if err != nil {
		return fmt.Errorf("truncate %s: %w", table, err)
	}
	return nil
}

func (p *psqlDeployOps) BulkInsert(ctx context.Context, ex Executor, table DBTable, columns []string, rows [][]any) error {
	if !IsValidTable(table) {
		return fmt.Errorf("bulk insert: unknown table %q", string(table))
	}
	if len(rows) == 0 {
		return nil
	}

	bs := batchSize(len(columns), psqlMaxVars)
	colList := strings.Join(columns, ", ")

	for i := 0; i < len(rows); i += bs {
		end := i + bs
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[i:end]

		// Build $1, $2, ... placeholders.
		var sb strings.Builder
		paramIdx := 1
		for rowIdx, row := range chunk {
			if rowIdx > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(")
			for colIdx := range row {
				if colIdx > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("$%d", paramIdx))
				paramIdx++
			}
			sb.WriteString(")")
		}

		query := "INSERT INTO " + string(table) + " (" + colList + ") VALUES " + sb.String() + ";"

		args := make([]any, 0, len(chunk)*len(columns))
		for _, row := range chunk {
			args = append(args, row...)
		}

		if _, err := ex.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("bulk insert %s (batch starting at row %d): %w", table, i, err)
		}
	}
	return nil
}

func (p *psqlDeployOps) Placeholder(n int) string { return fmt.Sprintf("$%d", n) }

func (p *psqlDeployOps) VerifyForeignKeys(ctx context.Context, ex Executor) ([]FKViolation, error) {
	// Query FK constraints with the child table's PK column name.
	fkRows, err := ex.QueryContext(ctx, `
		SELECT
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			pk.column_name  AS pk_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		JOIN (
			SELECT tco.table_name, kcup.column_name, tco.table_schema
			FROM information_schema.table_constraints AS tco
			JOIN information_schema.key_column_usage AS kcup
				ON tco.constraint_name = kcup.constraint_name
				AND tco.table_schema = kcup.table_schema
			WHERE tco.constraint_type = 'PRIMARY KEY'
				AND kcup.ordinal_position = 1
		) AS pk
			ON pk.table_name = tc.table_name
			AND pk.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = 'public';
	`)
	if err != nil {
		return nil, fmt.Errorf("list FK constraints: %w", err)
	}
	defer fkRows.Close()

	var fks []fkWithPK
	for fkRows.Next() {
		var fk fkWithPK
		if err := fkRows.Scan(&fk.table, &fk.column, &fk.refTable, &fk.refColumn, &fk.pkColumn); err != nil {
			return nil, fmt.Errorf("scan FK def: %w", err)
		}
		fks = append(fks, fk)
	}
	if err := fkRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate FK defs: %w", err)
	}

	return verifyFKsViaJoin(ctx, ex, fks)
}

func (p *psqlDeployOps) IntrospectColumns(ctx context.Context, table DBTable) ([]ColumnMeta, error) {
	if !IsValidTable(table) {
		return nil, fmt.Errorf("introspect: unknown table %q", string(table))
	}
	rows, err := p.pool.QueryContext(ctx,
		"SELECT column_name, data_type FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 ORDER BY ordinal_position",
		string(table),
	)
	if err != nil {
		return nil, fmt.Errorf("introspect %s: %w", table, err)
	}
	defer rows.Close()

	var cols []ColumnMeta
	for rows.Next() {
		var name, dataType string
		if err := rows.Scan(&name, &dataType); err != nil {
			return nil, fmt.Errorf("scan column info for %s: %w", table, err)
		}
		upper := strings.ToUpper(dataType)
		isInt := strings.Contains(upper, "INT") || upper == "SERIAL" || upper == "BIGSERIAL" || upper == "SMALLSERIAL"
		cols = append(cols, ColumnMeta{Name: name, IsInteger: isInt})
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q does not exist or has no columns", string(table))
	}
	return cols, rows.Err()
}

func (p *psqlDeployOps) QueryAllRows(ctx context.Context, table DBTable) ([]string, [][]any, error) {
	if !IsValidTable(table) {
		return nil, nil, fmt.Errorf("query all rows: unknown table %q", string(table))
	}
	colMeta, err := p.IntrospectColumns(ctx, table)
	if err != nil {
		return nil, nil, err
	}
	return queryAllRowsGeneric(ctx, p.pool, table, colMeta)
}

// ---------- Shared FK verification ----------

// fkWithPK holds a foreign key definition together with the child table's
// primary key column name — needed because Modula tables use typed PK
// names (datatype_id, content_data_id, etc.), not a uniform "id" column.
type fkWithPK struct {
	table     string // child table containing the FK column
	column    string // FK column in the child table
	refTable  string // parent table being referenced
	refColumn string // referenced column in the parent table
	pkColumn  string // primary key column of the child table
}

// verifyFKsViaJoin checks each FK constraint via a LEFT JOIN, selecting the
// child table's actual PK column to identify orphaned rows.
func verifyFKsViaJoin(ctx context.Context, ex Executor, fks []fkWithPK) ([]FKViolation, error) {
	var violations []FKViolation
	for idx, fk := range fks {
		q := fmt.Sprintf(
			"SELECT CAST(t.%s AS CHAR) FROM %s t LEFT JOIN %s p ON t.%s = p.%s WHERE p.%s IS NULL AND t.%s IS NOT NULL;",
			fk.pkColumn, fk.table, fk.refTable, fk.column, fk.refColumn, fk.refColumn, fk.column,
		)
		rows, err := ex.QueryContext(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("verify FK %s.%s -> %s.%s: %w", fk.table, fk.column, fk.refTable, fk.refColumn, err)
		}
		for rows.Next() {
			var rowID string
			if err := rows.Scan(&rowID); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan orphaned row in %s: %w", fk.table, err)
			}
			violations = append(violations, FKViolation{
				Table:   fk.table,
				RowID:   rowID,
				Parent:  fk.refTable,
				FKIndex: idx,
			})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate orphaned rows in %s: %w", fk.table, err)
		}
	}
	return violations, nil
}

// ---------- MySQL ----------

type mysqlDeployOps struct {
	pool *sql.DB
}

// ImportAtomic starts a transaction with FK checks disabled.
// Uses DELETE FROM (not TRUNCATE) so the operation is transactional.
func (m *mysqlDeployOps) ImportAtomic(ctx context.Context, fn ImportFunc) error {
	tx, err := m.pool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
		return fmt.Errorf("disable FK checks: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;"); err != nil {
		return fmt.Errorf("enable FK checks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (m *mysqlDeployOps) TruncateTable(ctx context.Context, ex Executor, table DBTable) error {
	if !IsValidTable(table) {
		return fmt.Errorf("truncate: unknown table %q", string(table))
	}
	// DELETE FROM instead of TRUNCATE TABLE so the operation stays
	// inside the transaction (TRUNCATE is DDL in MySQL and auto-commits).
	_, err := ex.ExecContext(ctx, "DELETE FROM "+string(table)+";")
	if err != nil {
		return fmt.Errorf("truncate %s: %w", table, err)
	}
	return nil
}

func (m *mysqlDeployOps) BulkInsert(ctx context.Context, ex Executor, table DBTable, columns []string, rows [][]any) error {
	if !IsValidTable(table) {
		return fmt.Errorf("bulk insert: unknown table %q", string(table))
	}
	if len(rows) == 0 {
		return nil
	}

	bs := mysqlDefaultBatchRows
	// Also respect a per-row variable count so we don't exceed MySQL's
	// 65535 prepared statement parameter limit.
	if varBs := batchSize(len(columns), 65535); varBs < bs {
		bs = varBs
	}
	colList := strings.Join(columns, ", ")

	for i := 0; i < len(rows); i += bs {
		end := i + bs
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[i:end]

		placeholderRow := "(" + strings.Repeat("?, ", len(columns)-1) + "?)"
		allPlaceholders := strings.Repeat(placeholderRow+", ", len(chunk)-1) + placeholderRow

		query := "INSERT INTO " + string(table) + " (" + colList + ") VALUES " + allPlaceholders + ";"

		args := make([]any, 0, len(chunk)*len(columns))
		for _, row := range chunk {
			args = append(args, row...)
		}

		if _, err := ex.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("bulk insert %s (batch starting at row %d): %w", table, i, err)
		}
	}
	return nil
}

func (m *mysqlDeployOps) Placeholder(_ int) string { return "?" }

func (m *mysqlDeployOps) VerifyForeignKeys(ctx context.Context, ex Executor) ([]FKViolation, error) {
	// Query FK definitions along with the child table's PK column.
	fkRows, err := ex.QueryContext(ctx, `
		SELECT
			fk.TABLE_NAME,
			fk.COLUMN_NAME,
			fk.REFERENCED_TABLE_NAME,
			fk.REFERENCED_COLUMN_NAME,
			pk.COLUMN_NAME AS PK_COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE AS fk
		JOIN (
			SELECT TABLE_NAME, COLUMN_NAME, TABLE_SCHEMA
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE CONSTRAINT_NAME = 'PRIMARY'
				AND TABLE_SCHEMA = DATABASE()
				AND ORDINAL_POSITION = 1
		) AS pk
			ON pk.TABLE_NAME = fk.TABLE_NAME
			AND pk.TABLE_SCHEMA = fk.TABLE_SCHEMA
		WHERE fk.TABLE_SCHEMA = DATABASE()
			AND fk.REFERENCED_TABLE_NAME IS NOT NULL;
	`)
	if err != nil {
		return nil, fmt.Errorf("list FK constraints: %w", err)
	}
	defer fkRows.Close()

	var fks []fkWithPK
	for fkRows.Next() {
		var fk fkWithPK
		if err := fkRows.Scan(&fk.table, &fk.column, &fk.refTable, &fk.refColumn, &fk.pkColumn); err != nil {
			return nil, fmt.Errorf("scan FK def: %w", err)
		}
		fks = append(fks, fk)
	}
	if err := fkRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate FK defs: %w", err)
	}

	return verifyFKsViaJoin(ctx, ex, fks)
}

func (m *mysqlDeployOps) IntrospectColumns(ctx context.Context, table DBTable) ([]ColumnMeta, error) {
	if !IsValidTable(table) {
		return nil, fmt.Errorf("introspect: unknown table %q", string(table))
	}
	rows, err := m.pool.QueryContext(ctx,
		"SELECT COLUMN_NAME, DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? ORDER BY ORDINAL_POSITION",
		string(table),
	)
	if err != nil {
		return nil, fmt.Errorf("introspect %s: %w", table, err)
	}
	defer rows.Close()

	var cols []ColumnMeta
	for rows.Next() {
		var name, dataType string
		if err := rows.Scan(&name, &dataType); err != nil {
			return nil, fmt.Errorf("scan column info for %s: %w", table, err)
		}
		upper := strings.ToUpper(dataType)
		isInt := strings.Contains(upper, "INT") || upper == "SERIAL"
		cols = append(cols, ColumnMeta{Name: name, IsInteger: isInt})
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q does not exist or has no columns", string(table))
	}
	return cols, rows.Err()
}

func (m *mysqlDeployOps) QueryAllRows(ctx context.Context, table DBTable) ([]string, [][]any, error) {
	if !IsValidTable(table) {
		return nil, nil, fmt.Errorf("query all rows: unknown table %q", string(table))
	}
	colMeta, err := m.IntrospectColumns(ctx, table)
	if err != nil {
		return nil, nil, err
	}
	return queryAllRowsGeneric(ctx, m.pool, table, colMeta)
}

// ---------- Shared helpers ----------

// queryAllRowsGeneric implements QueryAllRows for all backends.
// Scans every column as *sql.NullString then converts integer columns to int64.
func queryAllRowsGeneric(ctx context.Context, pool *sql.DB, table DBTable, colMeta []ColumnMeta) ([]string, [][]any, error) {
	intSet := make(map[int]bool, len(colMeta))
	colNames := make([]string, len(colMeta))
	for i, cm := range colMeta {
		colNames[i] = cm.Name
		if cm.IsInteger {
			intSet[i] = true
		}
	}

	rows, err := pool.QueryContext(ctx, "SELECT * FROM "+string(table)+";")
	if err != nil {
		return nil, nil, fmt.Errorf("query all rows %s: %w", table, err)
	}
	defer rows.Close()

	var result [][]any
	for rows.Next() {
		scanTargets := make([]any, len(colMeta))
		for i := range scanTargets {
			scanTargets[i] = new(sql.NullString)
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, nil, fmt.Errorf("scan row in %s: %w", table, err)
		}

		row := make([]any, len(colMeta))
		for i, target := range scanTargets {
			ns := target.(*sql.NullString)
			if !ns.Valid {
				row[i] = nil
				continue
			}
			if intSet[i] {
				var v int64
				fmt.Sscanf(ns.String, "%d", &v)
				row[i] = v
			} else {
				row[i] = ns.String
			}
		}
		result = append(result, row)
	}

	return colNames, result, rows.Err()
}

// ---------- Constructor ----------

// NewDeployOps creates a DeployOps implementation for the given DbDriver.
// It type-switches on the concrete driver to extract the *sql.DB pool.
func NewDeployOps(driver DbDriver) (DeployOps, error) {
	switch d := driver.(type) {
	case Database:
		return &sqliteDeployOps{pool: d.Connection}, nil
	case MysqlDatabase:
		return &mysqlDeployOps{pool: d.Connection}, nil
	case PsqlDatabase:
		return &psqlDeployOps{pool: d.Connection}, nil
	default:
		return nil, fmt.Errorf("unsupported driver type for deploy ops: %T", driver)
	}
}
