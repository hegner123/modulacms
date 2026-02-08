package generic

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

// CreateTableParams configures a CREATE TABLE statement.
type CreateTableParams struct {
	Table       string
	Columns     []CreateColumnDef
	Indexes     []IndexDef
	ForeignKeys []ForeignKeyDef
	IfNotExists bool
}

// CreateIndexParams configures a standalone CREATE INDEX statement.
type CreateIndexParams struct {
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

// CreateTable executes a CREATE TABLE statement with dialect-aware type mapping,
// optional indexes, and optional foreign key constraints.
func CreateTable(ctx context.Context, exec Executor, d Dialect, p CreateTableParams) error {
	if err := validateCreateTableParams(p); err != nil {
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
		err := CreateIndex(ctx, exec, d, CreateIndexParams{
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

// CreateIndex executes a standalone CREATE INDEX statement.
// Index names are auto-generated as idx_<table>_<col1>_<col2>.
func CreateIndex(ctx context.Context, exec Executor, d Dialect, p CreateIndexParams) error {
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

// validateCreateTableParams validates all fields of CreateTableParams.
func validateCreateTableParams(p CreateTableParams) error {
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
	Table   string
	Columns []string       // nil = SELECT *
	Where   map[string]any // AND equality; nil values produce IS NULL
	OrderBy string         // empty = no ORDER BY
	Desc    bool           // only used when OrderBy is set
	Limit   int64          // 0 = default (maxLimit); negative = no limit; positive = capped at maxLimit
	Offset  int64          // 0 = no offset
}

// InsertParams configures an INSERT query.
type InsertParams struct {
	Table  string
	Values map[string]any // must be non-empty
}

// UpdateParams configures an UPDATE query.
type UpdateParams struct {
	Table string
	Set   map[string]any // must be non-empty
	Where map[string]any // must be non-empty (safety: prevents full-table update)
}

// DeleteParams configures a DELETE query.
type DeleteParams struct {
	Table string
	Where map[string]any // must be non-empty (safety: prevents full-table delete)
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

// Select executes a SELECT query and returns all matching rows.
func Select(ctx context.Context, exec Executor, d Dialect, p SelectParams) ([]Row, error) {
	query, args, err := buildSelectQuery(d, p)
	if err != nil {
		return nil, err
	}
	return execQuery(ctx, exec, query, args)
}

// SelectRows executes a SELECT query and returns the raw *sql.Rows.
// Caller is responsible for closing the returned rows.
func SelectRows(ctx context.Context, exec Executor, d Dialect, p SelectParams) (*sql.Rows, error) {
	query, args, err := buildSelectQuery(d, p)
	if err != nil {
		return nil, err
	}
	return exec.QueryContext(ctx, query, args...)
}

// SelectOne executes a SELECT query and returns the first matching row, or nil if no match.
func SelectOne(ctx context.Context, exec Executor, d Dialect, p SelectParams) (Row, error) {
	p.Limit = 1
	rows, err := Select(ctx, exec, d, p)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// Insert executes an INSERT query.
func Insert(ctx context.Context, exec Executor, d Dialect, p InsertParams) (sql.Result, error) {
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

// Update executes an UPDATE query. Where must be non-empty to prevent accidental full-table updates.
func Update(ctx context.Context, exec Executor, d Dialect, p UpdateParams) (sql.Result, error) {
	if err := ValidTableName(p.Table); err != nil {
		return nil, fmt.Errorf("invalid table: %w", err)
	}
	if len(p.Set) == 0 {
		return nil, fmt.Errorf("update requires non-empty set")
	}
	if len(p.Where) == 0 {
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

	whereClause, whereArgs, err := buildWhere(d, p.Where, argIdx)
	if err != nil {
		return nil, err
	}
	query += whereClause
	args = append(args, whereArgs...)

	return exec.ExecContext(ctx, query, args...)
}

// Delete executes a DELETE query. Where must be non-empty to prevent accidental full-table deletes.
func Delete(ctx context.Context, exec Executor, d Dialect, p DeleteParams) (sql.Result, error) {
	if err := ValidTableName(p.Table); err != nil {
		return nil, fmt.Errorf("invalid table: %w", err)
	}
	if len(p.Where) == 0 {
		return nil, fmt.Errorf("delete requires non-empty where (safety: prevents full-table delete)")
	}

	query := fmt.Sprintf(`DELETE FROM %s`, quoteIdent(d, p.Table))

	whereClause, args, err := buildWhere(d, p.Where, 1)
	if err != nil {
		return nil, err
	}
	query += whereClause

	return exec.ExecContext(ctx, query, args...)
}

// Count returns the number of rows matching the WHERE conditions. Pass nil where for total count.
func Count(ctx context.Context, exec Executor, d Dialect, table string, where map[string]any) (int64, error) {
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

// Exists returns true if at least one row matches the WHERE conditions.
// Uses SELECT 1 ... LIMIT 1 for early termination instead of COUNT(*).
func Exists(ctx context.Context, exec Executor, d Dialect, table string, where map[string]any) (bool, error) {
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

	cols := "*"
	if len(p.Columns) > 0 {
		quoted := make([]string, len(p.Columns))
		for i, c := range p.Columns {
			if err := ValidColumnName(c); err != nil {
				return "", nil, fmt.Errorf("invalid column %q: %w", c, err)
			}
			quoted[i] = quoteIdent(d, c)
		}
		cols = strings.Join(quoted, ", ")
	}

	query = fmt.Sprintf(`SELECT %s FROM %s`, cols, quoteIdent(d, p.Table))
	whereClause, whereArgs, err := buildWhere(d, p.Where, 1)
	if err != nil {
		return "", nil, err
	}
	query += whereClause
	args = whereArgs

	if p.OrderBy != "" {
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
// nil values produce "column" IS NULL. Keys are sorted for deterministic output.
// argOffset is the 1-based starting index for placeholders (relevant for PostgreSQL).
func buildWhere(d Dialect, where map[string]any, argOffset int) (clause string, args []any, err error) {
	if len(where) == 0 {
		return "", nil, nil
	}

	keys := sortedKeys(where)
	conditions := make([]string, 0, len(keys))
	argIdx := argOffset

	for _, k := range keys {
		if err := ValidColumnName(k); err != nil {
			return "", nil, fmt.Errorf("invalid where column %q: %w", k, err)
		}
		if where[k] == nil {
			conditions = append(conditions, fmt.Sprintf(`%s IS NULL`, quoteIdent(d, k)))
		} else {
			conditions = append(conditions, fmt.Sprintf(`%s = %s`, quoteIdent(d, k), placeholder(d, argIdx)))
			args = append(args, where[k])
			argIdx++
		}
	}

	return " WHERE " + strings.Join(conditions, " AND "), args, nil
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
