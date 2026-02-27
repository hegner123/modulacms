package db

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"
)

// Dialect determines SQL placeholder and identifier quoting syntax.
type Dialect int

// Dialect constants define the supported database systems.
const (
	DialectSQLite   Dialect = iota
	DialectMySQL    Dialect = iota
	DialectPostgres Dialect = iota
)

// DialectFromString returns the Dialect matching a driver name string.
// Unrecognized values default to DialectSQLite.
func DialectFromString(driver string) Dialect {
	switch driver {
	case "mysql":
		return DialectMySQL
	case "postgres":
		return DialectPostgres
	default:
		return DialectSQLite
	}
}

// quoteIdent wraps an identifier in the dialect's quoting characters.
// SQLite and PostgreSQL use double quotes; MySQL uses backticks.
func quoteIdent(d Dialect, name string) string {
	if d == DialectMySQL {
		return "`" + name + "`"
	}
	return `"` + name + `"`
}

// quoteQualifiedIdent quotes a column name that may be qualified with a table name.
// "col" becomes `"col"`, "table.col" becomes `"table"."col"`.
// Returns an error if either part fails validation.
func quoteQualifiedIdent(d Dialect, name string) (string, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 1 {
		if err := ValidColumnName(parts[0]); err != nil {
			return "", fmt.Errorf("invalid column %q: %w", name, err)
		}
		return quoteIdent(d, parts[0]), nil
	}
	if err := ValidTableName(parts[0]); err != nil {
		return "", fmt.Errorf("invalid table in qualified name %q: %w", name, err)
	}
	if err := ValidColumnName(parts[1]); err != nil {
		return "", fmt.Errorf("invalid column in qualified name %q: %w", name, err)
	}
	return quoteIdent(d, parts[0]) + "." + quoteIdent(d, parts[1]), nil
}

// placeholder returns the parameter placeholder for the given 1-based index.
// SQLite and MySQL use ?; PostgreSQL uses $1, $2, etc.
func placeholder(d Dialect, index int) string {
	if d == DialectPostgres {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

// ===== COLUMN TYPE MAPPING =====

// ColumnType represents an abstract column type for cross-dialect DDL.
type ColumnType string

// ColumnType constants define the abstract column types supported across all database dialects.
const (
	ColText      ColumnType = "text"
	ColInteger   ColumnType = "integer"
	ColReal      ColumnType = "real"
	ColBlob      ColumnType = "blob"
	ColBoolean   ColumnType = "boolean"
	ColTimestamp ColumnType = "timestamp"
	ColJSON      ColumnType = "json"
)

// ValidColumnTypes lists all supported abstract column types.
var ValidColumnTypes = []ColumnType{
	ColText, ColInteger, ColReal, ColBlob, ColBoolean, ColTimestamp, ColJSON,
}

// ValidateColumnType returns an error if the given string is not a valid column type.
func ValidateColumnType(ct string) error {
	for _, valid := range ValidColumnTypes {
		if ColumnType(ct) == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid column type %q: must be one of text, integer, real, blob, boolean, timestamp, json", ct)
}

// typeMap maps each dialect and abstract column type to its concrete SQL type string.
var typeMap = map[Dialect]map[ColumnType]string{
	DialectSQLite: {
		ColText:      "TEXT",
		ColInteger:   "INTEGER",
		ColReal:      "REAL",
		ColBlob:      "BLOB",
		ColBoolean:   "INTEGER",
		ColTimestamp: "TEXT",
		ColJSON:      "TEXT",
	},
	DialectMySQL: {
		ColText:      "TEXT",
		ColInteger:   "INT",
		ColReal:      "DOUBLE",
		ColBoolean:   "TINYINT(1)",
		ColTimestamp: "TIMESTAMP",
		ColJSON:      "JSON",
		ColBlob:      "BLOB",
	},
	DialectPostgres: {
		ColText:      "TEXT",
		ColInteger:   "INTEGER",
		ColReal:      "DOUBLE PRECISION",
		ColBlob:      "BYTEA",
		ColBoolean:   "BOOLEAN",
		ColTimestamp: "TIMESTAMP",
		ColJSON:      "JSONB",
	},
}

// SQLType returns the concrete SQL type string for a given dialect and abstract column type.
// Panics if the column type is not in the type map (callers should validate first).
func SQLType(d Dialect, ct ColumnType) string {
	dialectMap, ok := typeMap[d]
	if !ok {
		return typeMap[DialectSQLite][ct]
	}
	sqlType, ok := dialectMap[ct]
	if !ok {
		return string(ct)
	}
	return sqlType
}

// ===== DDL TYPES =====

// CreateColumnDef defines a single column in a CREATE TABLE statement.
type CreateColumnDef struct {
	Name       string
	Type       ColumnType
	PrimaryKey bool
	NotNull    bool
	Default    string // empty = no default; value is wrapped in single quotes
	Unique     bool
}

// IndexDef defines an index to create alongside a table.
type IndexDef struct {
	Columns []string
	Unique  bool
}

// ForeignKeyDef defines a foreign key constraint in a CREATE TABLE statement.
type ForeignKeyDef struct {
	Column    string // local column
	RefTable  string // referenced table (full name, after any prefixing)
	RefColumn string // referenced column
	OnDelete  string // CASCADE, SET NULL, RESTRICT
}

// DDLCreateTableParams configures a CREATE TABLE statement.
// Named with DDL prefix to distinguish from the CMS entity CreateTableParams.
type DDLCreateTableParams struct {
	Table       string
	Columns     []CreateColumnDef
	Indexes     []IndexDef
	ForeignKeys []ForeignKeyDef
	IfNotExists bool
}

// DDLCreateIndexParams configures a standalone CREATE INDEX statement.
type DDLCreateIndexParams struct {
	Table       string
	Columns     []string
	Unique      bool
	IfNotExists bool
}

// validOnDeleteActions lists the allowed ON DELETE actions for foreign keys.
var validOnDeleteActions = []string{"CASCADE", "SET NULL", "RESTRICT"}

// maxColumns is the maximum number of columns allowed in a CREATE TABLE.
const maxColumns = 64

// ===== DDL FUNCTIONS =====

// DDLCreateTable executes a CREATE TABLE statement with dialect-aware type mapping,
// optional indexes, and optional foreign key constraints.
func DDLCreateTable(ctx context.Context, exec Executor, d Dialect, p DDLCreateTableParams) error {
	if err := validateDDLCreateTableParams(p); err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("CREATE TABLE ")
	if p.IfNotExists {
		sb.WriteString("IF NOT EXISTS ")
	}
	sb.WriteString(quoteIdent(d, p.Table))
	sb.WriteString(" (\n")

	for i, col := range p.Columns {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString("    ")
		sb.WriteString(quoteIdent(d, col.Name))
		sb.WriteString(" ")
		sb.WriteString(SQLType(d, col.Type))

		if col.NotNull {
			sb.WriteString(" NOT NULL")
		}
		if col.PrimaryKey {
			sb.WriteString(" PRIMARY KEY")
		}
		if col.Unique && !col.PrimaryKey {
			sb.WriteString(" UNIQUE")
		}
		if col.Default != "" {
			sb.WriteString(" DEFAULT '")
			sb.WriteString(escapeSingleQuotes(col.Default))
			sb.WriteString("'")
		}
	}

	for _, fk := range p.ForeignKeys {
		sb.WriteString(",\n    FOREIGN KEY (")
		sb.WriteString(quoteIdent(d, fk.Column))
		sb.WriteString(") REFERENCES ")
		sb.WriteString(quoteIdent(d, fk.RefTable))
		sb.WriteString(" (")
		sb.WriteString(quoteIdent(d, fk.RefColumn))
		sb.WriteString(")")
		if fk.OnDelete != "" {
			sb.WriteString(" ON DELETE ")
			sb.WriteString(fk.OnDelete)
		}
	}

	sb.WriteString("\n)")

	_, err := exec.ExecContext(ctx, sb.String())
	if err != nil {
		return fmt.Errorf("create table %q: %w", p.Table, err)
	}

	for _, idx := range p.Indexes {
		err := DDLCreateIndex(ctx, exec, d, DDLCreateIndexParams{
			Table:       p.Table,
			Columns:     idx.Columns,
			Unique:      idx.Unique,
			IfNotExists: p.IfNotExists,
		})
		if err != nil {
			return fmt.Errorf("create index on %q: %w", p.Table, err)
		}
	}

	return nil
}

// DDLCreateIndex executes a standalone CREATE INDEX statement.
// Index names are auto-generated as idx_<table>_<col1>_<col2>.
func DDLCreateIndex(ctx context.Context, exec Executor, d Dialect, p DDLCreateIndexParams) error {
	if err := ValidTableName(p.Table); err != nil {
		return fmt.Errorf("create index: invalid table: %w", err)
	}
	if len(p.Columns) == 0 {
		return fmt.Errorf("create index: columns cannot be empty")
	}
	for _, col := range p.Columns {
		if err := ValidColumnName(col); err != nil {
			return fmt.Errorf("create index: invalid column %q: %w", col, err)
		}
	}

	indexName := "idx_" + p.Table + "_" + strings.Join(p.Columns, "_")

	var sb strings.Builder
	sb.WriteString("CREATE ")
	if p.Unique {
		sb.WriteString("UNIQUE ")
	}
	sb.WriteString("INDEX ")
	if p.IfNotExists && d != DialectMySQL {
		sb.WriteString("IF NOT EXISTS ")
	}
	sb.WriteString(quoteIdent(d, indexName))
	sb.WriteString(" ON ")
	sb.WriteString(quoteIdent(d, p.Table))
	sb.WriteString(" (")

	quotedCols := make([]string, len(p.Columns))
	for i, col := range p.Columns {
		quotedCols[i] = quoteIdent(d, col)
	}
	sb.WriteString(strings.Join(quotedCols, ", "))
	sb.WriteString(")")

	_, err := exec.ExecContext(ctx, sb.String())
	if err != nil {
		// MySQL lacks IF NOT EXISTS for indexes; treat "already exists" as success
		if d == DialectMySQL && p.IfNotExists && strings.Contains(err.Error(), "Duplicate key name") {
			return nil
		}
		return fmt.Errorf("create index %q: %w", indexName, err)
	}

	return nil
}

// validateDDLCreateTableParams validates all fields of DDLCreateTableParams.
func validateDDLCreateTableParams(p DDLCreateTableParams) error {
	if err := ValidTableName(p.Table); err != nil {
		return fmt.Errorf("invalid table: %w", err)
	}

	if len(p.Columns) == 0 {
		return fmt.Errorf("columns cannot be empty")
	}
	if len(p.Columns) > maxColumns {
		return fmt.Errorf("too many columns: %d (max %d)", len(p.Columns), maxColumns)
	}

	pkCount := 0
	colNames := make(map[string]bool, len(p.Columns))
	for _, col := range p.Columns {
		if err := ValidColumnName(col.Name); err != nil {
			return fmt.Errorf("invalid column name %q: %w", col.Name, err)
		}
		if colNames[col.Name] {
			return fmt.Errorf("duplicate column name %q", col.Name)
		}
		colNames[col.Name] = true

		if err := ValidateColumnType(string(col.Type)); err != nil {
			return err
		}
		if col.PrimaryKey {
			pkCount++
		}
	}

	if pkCount == 0 {
		return fmt.Errorf("at least one primary key column is required")
	}
	if pkCount > 1 {
		return fmt.Errorf("only one primary key column is allowed (got %d)", pkCount)
	}

	for _, fk := range p.ForeignKeys {
		if err := ValidColumnName(fk.Column); err != nil {
			return fmt.Errorf("invalid FK column %q: %w", fk.Column, err)
		}
		if !colNames[fk.Column] {
			return fmt.Errorf("FK column %q not found in column definitions", fk.Column)
		}
		if err := ValidTableName(fk.RefTable); err != nil {
			return fmt.Errorf("invalid FK reference table %q: %w", fk.RefTable, err)
		}
		if err := ValidColumnName(fk.RefColumn); err != nil {
			return fmt.Errorf("invalid FK reference column %q: %w", fk.RefColumn, err)
		}
		if fk.OnDelete != "" && !slices.Contains(validOnDeleteActions, fk.OnDelete) {
			return fmt.Errorf("invalid FK ON DELETE action %q: must be one of CASCADE, SET NULL, RESTRICT", fk.OnDelete)
		}
	}

	for _, idx := range p.Indexes {
		if len(idx.Columns) == 0 {
			return fmt.Errorf("index columns cannot be empty")
		}
		for _, col := range idx.Columns {
			if err := ValidColumnName(col); err != nil {
				return fmt.Errorf("invalid index column %q: %w", col, err)
			}
			if !colNames[col] {
				return fmt.Errorf("index column %q not found in column definitions", col)
			}
		}
	}

	return nil
}

// escapeSingleQuotes doubles single quotes in a string for safe embedding in SQL literals.
func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// Executor abstracts *sql.DB and *sql.Tx for query execution.
type Executor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Row is a single result row mapping column name to value.
type Row map[string]any

// SelectParams configures a SELECT query.
type SelectParams struct {
	Table    string
	Columns  []string         // nil = SELECT *
	Where    map[string]any   // AND conditions; nil values produce IS NULL; Condition values use operators
	WhereOr  []map[string]any // OR groups; each inner map is AND-joined, groups are OR-joined
	Joins    []JoinClause     // JOIN clauses (INNER, LEFT, RIGHT)
	OrderBy  string           // empty = no ORDER BY (legacy; use Orders for multi-column)
	Desc     bool             // only used when OrderBy is set (legacy; use Orders for multi-column)
	Orders   []OrderByClause  // multi-column ORDER BY; takes precedence over OrderBy/Desc when non-empty
	GroupBy  []string         // GROUP BY column names
	Having   map[string]any   // HAVING conditions (same Condition system as Where); requires GroupBy
	Distinct bool             // SELECT DISTINCT
	Limit    int64            // 0 = default (maxLimit); negative = no limit; positive = capped at maxLimit
	Offset   int64            // 0 = no offset
}

// InsertParams configures an INSERT query.
type InsertParams struct {
	Table  string
	Values map[string]any // must be non-empty
}

// UpdateParams configures an UPDATE query.
type UpdateParams struct {
	Table   string
	Set     map[string]any   // must be non-empty
	Where   map[string]any   // AND conditions (safety: prevents full-table update)
	WhereOr []map[string]any // OR groups; each inner map is AND-joined, groups are OR-joined
}

// DeleteParams configures a DELETE query.
type DeleteParams struct {
	Table   string
	Where   map[string]any   // must be non-empty (safety: prevents full-table delete)
	WhereOr []map[string]any // OR groups; each inner map is AND-joined, groups are OR-joined
}

// ===== CONDITION TYPE =====

// Condition represents a comparison operator with a value for WHERE clauses.
// Use the constructor functions (Eq, Gt, Like, In, etc.) to create conditions.
type Condition struct {
	op    string
	value any
}

// Eq creates an explicit equality condition (=).
func Eq(v any) Condition { return Condition{op: "=", value: v} }

// Neq creates a not-equal condition (!=).
func Neq(v any) Condition { return Condition{op: "!=", value: v} }

// Gt creates a greater-than condition (>).
func Gt(v any) Condition { return Condition{op: ">", value: v} }

// Gte creates a greater-than-or-equal condition (>=).
func Gte(v any) Condition { return Condition{op: ">=", value: v} }

// Lt creates a less-than condition (<).
func Lt(v any) Condition { return Condition{op: "<", value: v} }

// Lte creates a less-than-or-equal condition (<=).
func Lte(v any) Condition { return Condition{op: "<=", value: v} }

// Like creates a LIKE pattern condition.
func Like(v string) Condition { return Condition{op: "LIKE", value: v} }

// NotLike creates a NOT LIKE pattern condition.
func NotLike(v string) Condition { return Condition{op: "NOT LIKE", value: v} }

// In creates an IN condition. Panics if no values provided.
func In(vals ...any) Condition { return Condition{op: "IN", value: vals} }

// NotIn creates a NOT IN condition. Panics if no values provided.
func NotIn(vals ...any) Condition { return Condition{op: "NOT IN", value: vals} }

// Between creates a BETWEEN low AND high condition.
func Between(low, high any) Condition { return Condition{op: "BETWEEN", value: [2]any{low, high}} }

// IsNull creates an IS NULL condition.
func IsNull() Condition { return Condition{op: "IS NULL"} }

// IsNotNull creates an IS NOT NULL condition.
func IsNotNull() Condition { return Condition{op: "IS NOT NULL"} }

// ===== JOIN TYPES =====

// JoinType represents the type of SQL JOIN.
type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
)

// JoinClause defines a single JOIN in a SELECT query.
type JoinClause struct {
	Type       JoinType // InnerJoin, LeftJoin, RightJoin
	Table      string   // table to join (validated)
	LocalCol   string   // column on the left side (may be "table.col")
	ForeignCol string   // column on the joined table (may be "table.col")
}

// ===== ORDER BY =====

// OrderByClause defines a single column in a multi-column ORDER BY.
type OrderByClause struct {
	Column string // may be qualified ("table.col")
	Desc   bool
}

const maxLimit int64 = 10000

// ===== VALIDATION =====

// isValidIdentifier validates that a string is a valid SQL identifier.
// Only allows alphanumeric characters and underscores, must start with letter or underscore.
func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	firstChar := rune(name[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return false
	}

	for _, char := range name[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}

	return true
}

// sqlKeywords is the set of SQL keywords rejected as identifiers.
var sqlKeywords = []string{
	"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "INDEX",
	"TABLE", "DATABASE", "SCHEMA", "VIEW", "PROCEDURE", "FUNCTION", "TRIGGER",
	"UNION", "WHERE", "ORDER", "GROUP", "HAVING", "FROM", "JOIN", "ON", "AS",
}

// ValidTableName validates that a table name is safe to use in queries.
// Only allows alphanumeric characters, underscores, and limited length.
func ValidTableName(tableName string) error {
	if len(tableName) == 0 {
		return fmt.Errorf("table name cannot be empty")
	}
	if len(tableName) > 64 {
		return fmt.Errorf("table name too long (max 64 characters)")
	}

	if !isValidIdentifier(tableName) {
		return fmt.Errorf("invalid table name format: must contain only letters, numbers, and underscores, and start with a letter or underscore")
	}

	upperTableName := strings.ToUpper(tableName)
	if slices.Contains(sqlKeywords, upperTableName) {
		return fmt.Errorf("table name cannot be a SQL keyword: %s", tableName)
	}

	return nil
}

// ValidColumnName validates that a column name is safe to use in queries.
func ValidColumnName(columnName string) error {
	if len(columnName) == 0 {
		return fmt.Errorf("column name cannot be empty")
	}
	if len(columnName) > 64 {
		return fmt.Errorf("column name too long (max 64 characters)")
	}

	if !isValidIdentifier(columnName) {
		return fmt.Errorf("invalid column name format: must contain only letters, numbers, and underscores, and start with a letter or underscore")
	}

	upperColumnName := strings.ToUpper(columnName)
	if slices.Contains(sqlKeywords, upperColumnName) {
		return fmt.Errorf("column name cannot be a SQL keyword: %s", columnName)
	}

	return nil
}

// ===== QUERY FUNCTIONS =====

// QSelect executes a SELECT query and returns all matching rows.
func QSelect(ctx context.Context, exec Executor, d Dialect, p SelectParams) ([]Row, error) {
	query, args, err := buildSelectQuery(d, p)
	if err != nil {
		return nil, err
	}
	return execQuery(ctx, exec, query, args)
}

// QSelectRows executes a SELECT query and returns the raw *sql.Rows.
// Caller is responsible for closing the returned rows.
func QSelectRows(ctx context.Context, exec Executor, d Dialect, p SelectParams) (*sql.Rows, error) {
	query, args, err := buildSelectQuery(d, p)
	if err != nil {
		return nil, err
	}
	return exec.QueryContext(ctx, query, args...)
}

// QSelectOne executes a SELECT query and returns the first matching row, or nil if no match.
func QSelectOne(ctx context.Context, exec Executor, d Dialect, p SelectParams) (Row, error) {
	p.Limit = 1
	rows, err := QSelect(ctx, exec, d, p)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// QInsert executes an INSERT query.
func QInsert(ctx context.Context, exec Executor, d Dialect, p InsertParams) (sql.Result, error) {
	if err := ValidTableName(p.Table); err != nil {
		return nil, fmt.Errorf("invalid table: %w", err)
	}
	if len(p.Values) == 0 {
		return nil, fmt.Errorf("insert requires non-empty values")
	}

	keys := sortedKeys(p.Values)
	quoted := make([]string, len(keys))
	placeholders := make([]string, len(keys))
	args := make([]any, 0, len(keys))

	for i, k := range keys {
		if err := ValidColumnName(k); err != nil {
			return nil, fmt.Errorf("invalid column %q: %w", k, err)
		}
		quoted[i] = quoteIdent(d, k)
		placeholders[i] = placeholder(d, i+1)
		args = append(args, p.Values[k])
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
		quoteIdent(d, p.Table),
		strings.Join(quoted, ", "),
		strings.Join(placeholders, ", "),
	)

	return exec.ExecContext(ctx, query, args...)
}

// QUpdate executes an UPDATE query. Where or WhereOr must be non-empty to prevent accidental full-table updates.
func QUpdate(ctx context.Context, exec Executor, d Dialect, p UpdateParams) (sql.Result, error) {
	if err := ValidTableName(p.Table); err != nil {
		return nil, fmt.Errorf("invalid table: %w", err)
	}
	if len(p.Set) == 0 {
		return nil, fmt.Errorf("update requires non-empty set")
	}
	if len(p.Where) == 0 && len(p.WhereOr) == 0 {
		return nil, fmt.Errorf("update requires non-empty where (safety: prevents full-table update)")
	}

	setKeys := sortedKeys(p.Set)
	setClauses := make([]string, len(setKeys))
	args := make([]any, 0, len(setKeys)+len(p.Where))
	argIdx := 1

	for i, k := range setKeys {
		if err := ValidColumnName(k); err != nil {
			return nil, fmt.Errorf("invalid set column %q: %w", k, err)
		}
		if p.Set[k] == nil {
			setClauses[i] = fmt.Sprintf(`%s = NULL`, quoteIdent(d, k))
		} else {
			setClauses[i] = fmt.Sprintf(`%s = %s`, quoteIdent(d, k), placeholder(d, argIdx))
			args = append(args, p.Set[k])
			argIdx++
		}
	}

	query := fmt.Sprintf(`UPDATE %s SET %s`, quoteIdent(d, p.Table), strings.Join(setClauses, ", "))

	whereClause, whereArgs, nextIdx, err := buildWhereClause(d, p.Where, p.WhereOr, argIdx)
	if err != nil {
		return nil, err
	}
	_ = nextIdx
	query += whereClause
	args = append(args, whereArgs...)

	return exec.ExecContext(ctx, query, args...)
}

// QDelete executes a DELETE query. Where or WhereOr must be non-empty to prevent accidental full-table deletes.
func QDelete(ctx context.Context, exec Executor, d Dialect, p DeleteParams) (sql.Result, error) {
	if err := ValidTableName(p.Table); err != nil {
		return nil, fmt.Errorf("invalid table: %w", err)
	}
	if len(p.Where) == 0 && len(p.WhereOr) == 0 {
		return nil, fmt.Errorf("delete requires non-empty where (safety: prevents full-table delete)")
	}

	query := fmt.Sprintf(`DELETE FROM %s`, quoteIdent(d, p.Table))

	whereClause, args, _, err := buildWhereClause(d, p.Where, p.WhereOr, 1)
	if err != nil {
		return nil, err
	}
	query += whereClause

	return exec.ExecContext(ctx, query, args...)
}

// QCount returns the number of rows matching the WHERE conditions. Pass nil where for total count.
func QCount(ctx context.Context, exec Executor, d Dialect, table string, where map[string]any) (int64, error) {
	if err := ValidTableName(table); err != nil {
		return 0, fmt.Errorf("invalid table: %w", err)
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, quoteIdent(d, table))
	whereClause, args, err := buildWhere(d, where, 1)
	if err != nil {
		return 0, err
	}
	query += whereClause

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("count query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, fmt.Errorf("count query returned no rows")
	}

	var count int64
	if err := rows.Scan(&count); err != nil {
		return 0, fmt.Errorf("count scan failed: %w", err)
	}
	return count, nil
}

// QExists returns true if at least one row matches the WHERE conditions.
// Uses SELECT 1 ... LIMIT 1 for early termination instead of COUNT(*).
func QExists(ctx context.Context, exec Executor, d Dialect, table string, where map[string]any) (bool, error) {
	if err := ValidTableName(table); err != nil {
		return false, fmt.Errorf("invalid table: %w", err)
	}

	query := fmt.Sprintf(`SELECT 1 FROM %s`, quoteIdent(d, table))
	whereClause, args, err := buildWhere(d, where, 1)
	if err != nil {
		return false, err
	}
	query += whereClause + " LIMIT 1"

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("exists query failed: %w", err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

// ===== INTERNAL HELPERS =====

// buildSelectQuery constructs a validated SELECT query string with args.
func buildSelectQuery(d Dialect, p SelectParams) (query string, args []any, err error) {
	if err := ValidTableName(p.Table); err != nil {
		return "", nil, fmt.Errorf("invalid table: %w", err)
	}

	// DISTINCT
	selectKw := "SELECT"
	if p.Distinct {
		selectKw = "SELECT DISTINCT"
	}

	// Columns
	cols := "*"
	if len(p.Columns) > 0 {
		quoted := make([]string, len(p.Columns))
		for i, c := range p.Columns {
			q, qerr := quoteQualifiedIdent(d, c)
			if qerr != nil {
				return "", nil, fmt.Errorf("invalid column %q: %w", c, qerr)
			}
			quoted[i] = q
		}
		cols = strings.Join(quoted, ", ")
	}

	query = fmt.Sprintf(`%s %s FROM %s`, selectKw, cols, quoteIdent(d, p.Table))
	argIdx := 1

	// JOINs
	for _, j := range p.Joins {
		if err := ValidTableName(j.Table); err != nil {
			return "", nil, fmt.Errorf("invalid join table %q: %w", j.Table, err)
		}
		localQ, lerr := quoteQualifiedIdent(d, j.LocalCol)
		if lerr != nil {
			return "", nil, fmt.Errorf("invalid join local column %q: %w", j.LocalCol, lerr)
		}
		foreignQ, ferr := quoteQualifiedIdent(d, j.ForeignCol)
		if ferr != nil {
			return "", nil, fmt.Errorf("invalid join foreign column %q: %w", j.ForeignCol, ferr)
		}
		joinType := string(j.Type)
		if joinType == "" {
			joinType = "INNER JOIN"
		}
		query += fmt.Sprintf(` %s %s ON %s = %s`, joinType, quoteIdent(d, j.Table), localQ, foreignQ)
	}

	// WHERE
	whereClause, whereArgs, nextIdx, werr := buildWhereClause(d, p.Where, p.WhereOr, argIdx)
	if werr != nil {
		return "", nil, werr
	}
	query += whereClause
	args = whereArgs
	argIdx = nextIdx

	// GROUP BY
	if len(p.GroupBy) > 0 {
		groupCols := make([]string, len(p.GroupBy))
		for i, c := range p.GroupBy {
			q, qerr := quoteQualifiedIdent(d, c)
			if qerr != nil {
				return "", nil, fmt.Errorf("invalid group_by column %q: %w", c, qerr)
			}
			groupCols[i] = q
		}
		query += " GROUP BY " + strings.Join(groupCols, ", ")
	}

	// HAVING (requires GROUP BY)
	if len(p.Having) > 0 {
		if len(p.GroupBy) == 0 {
			return "", nil, fmt.Errorf("having requires group_by")
		}
		havingClause, havingArgs, hNextIdx, herr := buildConditionMap(d, p.Having, argIdx)
		if herr != nil {
			return "", nil, fmt.Errorf("invalid having: %w", herr)
		}
		query += " HAVING " + strings.Join(havingClause, " AND ")
		args = append(args, havingArgs...)
		argIdx = hNextIdx
	}

	// ORDER BY: new Orders takes precedence over legacy OrderBy/Desc
	if len(p.Orders) > 0 {
		orderParts := make([]string, len(p.Orders))
		for i, o := range p.Orders {
			q, qerr := quoteQualifiedIdent(d, o.Column)
			if qerr != nil {
				return "", nil, fmt.Errorf("invalid order_by column %q: %w", o.Column, qerr)
			}
			dir := "ASC"
			if o.Desc {
				dir = "DESC"
			}
			orderParts[i] = q + " " + dir
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	} else if p.OrderBy != "" {
		if err := ValidColumnName(p.OrderBy); err != nil {
			return "", nil, fmt.Errorf("invalid order_by column %q: %w", p.OrderBy, err)
		}
		direction := "ASC"
		if p.Desc {
			direction = "DESC"
		}
		query += fmt.Sprintf(` ORDER BY %s %s`, quoteIdent(d, p.OrderBy), direction)
	}

	// Apply limit: 0 = default (maxLimit), negative = no limit, positive = capped at maxLimit
	limit := p.Limit
	if limit == 0 {
		limit = maxLimit
	}
	if limit > 0 {
		if limit > maxLimit {
			limit = maxLimit
		}
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	if p.Offset > 0 {
		query += fmt.Sprintf(` OFFSET %d`, p.Offset)
	}

	return query, args, nil
}

// buildWhere constructs a WHERE clause from a map of column=value conditions joined by AND.
// nil values produce "column" IS NULL. Condition values use their operator.
// Keys are sorted for deterministic output.
// argOffset is the 1-based starting index for placeholders (relevant for PostgreSQL).
// This is the thin wrapper used by QCount and QExists (which don't support WhereOr).
func buildWhere(d Dialect, where map[string]any, argOffset int) (clause string, args []any, err error) {
	if len(where) == 0 {
		return "", nil, nil
	}

	conditions, condArgs, _, cerr := buildConditionMap(d, where, argOffset)
	if cerr != nil {
		return "", nil, cerr
	}

	return " WHERE " + strings.Join(conditions, " AND "), condArgs, nil
}

// buildWhereClause constructs a full WHERE clause supporting both AND conditions (where)
// and OR groups (whereOr). Returns the clause string, args, next arg index, and error.
//
// Semantics:
//   - Where only   → WHERE a AND b
//   - WhereOr only → WHERE (g1) OR (g2)
//   - Both         → WHERE (a AND b) AND ((g1) OR (g2))
func buildWhereClause(d Dialect, where map[string]any, whereOr []map[string]any, argOffset int) (clause string, args []any, nextIdx int, err error) {
	argIdx := argOffset

	if len(where) == 0 && len(whereOr) == 0 {
		return "", nil, argIdx, nil
	}

	var parts []string

	// AND group from Where
	if len(where) > 0 {
		conditions, condArgs, nextArg, cerr := buildConditionMap(d, where, argIdx)
		if cerr != nil {
			return "", nil, 0, cerr
		}
		andPart := strings.Join(conditions, " AND ")
		if len(whereOr) > 0 {
			andPart = "(" + andPart + ")"
		}
		parts = append(parts, andPart)
		args = append(args, condArgs...)
		argIdx = nextArg
	}

	// OR groups from WhereOr
	if len(whereOr) > 0 {
		orParts := make([]string, 0, len(whereOr))
		for _, group := range whereOr {
			if len(group) == 0 {
				continue
			}
			conditions, condArgs, nextArg, cerr := buildConditionMap(d, group, argIdx)
			if cerr != nil {
				return "", nil, 0, cerr
			}
			orParts = append(orParts, "("+strings.Join(conditions, " AND ")+")")
			args = append(args, condArgs...)
			argIdx = nextArg
		}
		if len(orParts) > 0 {
			orClause := strings.Join(orParts, " OR ")
			if len(where) > 0 {
				orClause = "(" + orClause + ")"
			}
			parts = append(parts, orClause)
		}
	}

	if len(parts) == 0 {
		return "", nil, argIdx, nil
	}

	return " WHERE " + strings.Join(parts, " AND "), args, argIdx, nil
}

// buildConditionMap processes a map[string]any of conditions, returning SQL fragments,
// args, and the next arg index. Handles plain values, nil, and Condition types.
func buildConditionMap(d Dialect, m map[string]any, argOffset int) (conditions []string, args []any, nextIdx int, err error) {
	keys := sortedKeys(m)
	conditions = make([]string, 0, len(keys))
	argIdx := argOffset

	for _, k := range keys {
		quotedCol, qerr := quoteQualifiedIdent(d, k)
		if qerr != nil {
			return nil, nil, 0, fmt.Errorf("invalid where column %q: %w", k, qerr)
		}

		switch v := m[k].(type) {
		case Condition:
			frag, condArgs, nextArg, cerr := buildCondition(d, quotedCol, v, argIdx)
			if cerr != nil {
				return nil, nil, 0, fmt.Errorf("column %q: %w", k, cerr)
			}
			conditions = append(conditions, frag)
			args = append(args, condArgs...)
			argIdx = nextArg
		case nil:
			conditions = append(conditions, quotedCol+" IS NULL")
		default:
			conditions = append(conditions, quotedCol+" = "+placeholder(d, argIdx))
			args = append(args, v)
			argIdx++
		}
	}

	return conditions, args, argIdx, nil
}

// buildCondition generates a SQL fragment for a single Condition.
func buildCondition(d Dialect, quotedCol string, cond Condition, argIdx int) (frag string, args []any, nextIdx int, err error) {
	switch cond.op {
	case "=", "!=", ">", ">=", "<", "<=", "LIKE", "NOT LIKE":
		return fmt.Sprintf("%s %s %s", quotedCol, cond.op, placeholder(d, argIdx)),
			[]any{cond.value}, argIdx + 1, nil

	case "IN", "NOT IN":
		vals, ok := cond.value.([]any)
		if !ok || len(vals) == 0 {
			return "", nil, 0, fmt.Errorf("%s requires at least one value", cond.op)
		}
		phs := make([]string, len(vals))
		for i := range vals {
			phs[i] = placeholder(d, argIdx+i)
		}
		return fmt.Sprintf("%s %s (%s)", quotedCol, cond.op, strings.Join(phs, ", ")),
			vals, argIdx + len(vals), nil

	case "BETWEEN":
		pair, ok := cond.value.([2]any)
		if !ok {
			return "", nil, 0, fmt.Errorf("BETWEEN requires [2]any value")
		}
		return fmt.Sprintf("%s BETWEEN %s AND %s", quotedCol, placeholder(d, argIdx), placeholder(d, argIdx+1)),
			[]any{pair[0], pair[1]}, argIdx + 2, nil

	case "IS NULL":
		return quotedCol + " IS NULL", nil, argIdx, nil

	case "IS NOT NULL":
		return quotedCol + " IS NOT NULL", nil, argIdx, nil

	default:
		return "", nil, 0, fmt.Errorf("unsupported operator %q", cond.op)
	}
}

// execQuery runs a SELECT query and scans all result rows into []Row.
func execQuery(ctx context.Context, exec Executor, query string, args []any) ([]Row, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var result []Row
	for rows.Next() {
		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		row := make(Row, len(columns))
		for i, col := range columns {
			row[col] = values[i]
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}

// sortedKeys returns the keys of a map sorted alphabetically for deterministic query generation.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
