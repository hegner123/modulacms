package db

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/hegner123/modulacms/internal/utility"
)

// SecureQueryBuilder provides SQL injection safe query building for SQLite
type SecureQueryBuilder struct {
	db *sql.DB
}

// NewSecureQueryBuilder creates a new secure query builder
func NewSecureQueryBuilder(db *sql.DB) *SecureQueryBuilder {
	return &SecureQueryBuilder{db: db}
}

// isValidIdentifier validates that a string is a valid SQL identifier
// Only allows alphanumeric characters and underscores, must start with letter or underscore
func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character must be letter or underscore
	firstChar := rune(name[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return false
	}

	// Remaining characters must be alphanumeric or underscore
	for _, char := range name[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}

	return true
}

// ValidTableName validates that a table name is safe to use in queries
// Only allows alphanumeric characters, underscores, and limited length
func ValidTableName(tableName string) error {
	if len(tableName) == 0 {
		return fmt.Errorf("table name cannot be empty")
	}
	if len(tableName) > 64 {
		return fmt.Errorf("table name too long (max 64 characters)")
	}

	// Only allow alphanumeric and underscores, must start with letter or underscore
	if !isValidIdentifier(tableName) {
		return fmt.Errorf("invalid table name format: must contain only letters, numbers, and underscores, and start with a letter or underscore")
	}

	// Prevent SQL keywords as table names
	keywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "INDEX",
		"TABLE", "DATABASE", "SCHEMA", "VIEW", "PROCEDURE", "FUNCTION", "TRIGGER",
		"UNION", "WHERE", "ORDER", "GROUP", "HAVING", "FROM", "JOIN", "ON", "AS",
	}

	upperTableName := strings.ToUpper(tableName)
	if slices.Contains(keywords, upperTableName) {
		return fmt.Errorf("table name cannot be a SQL keyword: %s", tableName)

	}

	return nil
}

// ValidColumnName validates that a column name is safe to use in queries
func ValidColumnName(columnName string) error {
	if len(columnName) == 0 {
		return fmt.Errorf("column name cannot be empty")
	}
	if len(columnName) > 64 {
		return fmt.Errorf("column name too long (max 64 characters)")
	}

	// Only allow alphanumeric and underscores, must start with letter or underscore
	if !isValidIdentifier(columnName) {
		return fmt.Errorf("invalid column name format: must contain only letters, numbers, and underscores, and start with a letter or underscore")
	}

	return nil
}

// SecureBuildSelectQuery creates a secure SELECT query with parameterized queries
func (sqb *SecureQueryBuilder) SecureBuildSelectQuery(tableName string, id int64) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	// Use quoted identifiers for table names to prevent injection
	query := fmt.Sprintf(`SELECT * FROM "%s" WHERE id = ?`, tableName)
	return query, []any{id}, nil
}

// SecureBuildListQuery creates a secure SELECT query for listing all records
func (sqb *SecureQueryBuilder) SecureBuildListQuery(tableName string) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	// Use quoted identifiers for table names to prevent injection
	query := fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	return query, []any{}, nil
}

// SecureBuildSelectWithColumnsQuery creates a secure SELECT query for specific columns
func (sqb *SecureQueryBuilder) SecureBuildSelectWithColumnsQuery(tableName string, columns []string, whereColumn string, value any) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	if err := ValidColumnName(whereColumn); err != nil {
		utility.DefaultLogger.Ferror("Invalid where column name: %v", err)
		return "", nil, err
	}

	// Validate all column names
	for _, col := range columns {
		if err := ValidColumnName(col); err != nil {
			utility.DefaultLogger.Ferror("Invalid column name: %v", err)
			return "", nil, err
		}
	}

	// Build column list with quoted identifiers
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = fmt.Sprintf(`"%s"`, col)
	}

	columnList := strings.Join(quotedColumns, ", ")
	query := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "%s" = ?`, columnList, tableName, whereColumn)

	return query, []any{value}, nil
}

// SecureBuildListWithLimitQuery creates a secure SELECT query with LIMIT and OFFSET
func (sqb *SecureQueryBuilder) SecureBuildListWithLimitQuery(tableName string, limit, offset int64) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	if limit < 0 || offset < 0 {
		return "", nil, fmt.Errorf("limit and offset must be non-negative")
	}

	// Reasonable limits to prevent resource exhaustion
	if limit > 10000 {
		return "", nil, fmt.Errorf("limit too large (max 10000)")
	}

	query := fmt.Sprintf(`SELECT * FROM "%s" LIMIT ? OFFSET ?`, tableName)
	return query, []any{limit, offset}, nil
}

// SecureBuildOrderedListQuery creates a secure SELECT query with ORDER BY
func (sqb *SecureQueryBuilder) SecureBuildOrderedListQuery(tableName string, orderByColumn string, ascending bool) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	if err := ValidColumnName(orderByColumn); err != nil {
		utility.DefaultLogger.Ferror("Invalid order by column name: %v", err)
		return "", nil, err
	}

	order := "ASC"
	if !ascending {
		order = "DESC"
	}

	query := fmt.Sprintf(`SELECT * FROM "%s" ORDER BY "%s" %s`, tableName, orderByColumn, order)
	return query, []any{}, nil
}

// SecureExecuteSelectQuery executes a secure SELECT query and returns rows
func (sqb *SecureQueryBuilder) SecureExecuteSelectQuery(query string, args []any) (*sql.Rows, error) {
	utility.DefaultLogger.Finfo("Executing secure query: %s with args: %v", query, args)

	rows, err := sqb.db.Query(query, args...)
	if err != nil {
		utility.DefaultLogger.Ferror("Query execution failed: %v", err)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return rows, nil
}

// SecureCountQuery creates a secure COUNT query
func (sqb *SecureQueryBuilder) SecureCountQuery(tableName string) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
	return query, []any{}, nil
}

// SecureExistsQuery creates a secure EXISTS query
func (sqb *SecureQueryBuilder) SecureExistsQuery(tableName string, whereColumn string, value any) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	if err := ValidColumnName(whereColumn); err != nil {
		utility.DefaultLogger.Ferror("Invalid where column name: %v", err)
		return "", nil, err
	}

	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM "%s" WHERE "%s" = ?)`, tableName, whereColumn)
	return query, []any{value}, nil
}

// SecureBuildInsertQuery creates a secure INSERT query with parameterized queries
func (sqb *SecureQueryBuilder) SecureBuildInsertQuery(tableName string, values map[string]any) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	if len(values) == 0 {
		return "", nil, fmt.Errorf("no values provided for insert")
	}

	// Validate all column names
	var columns []string
	var placeholders []string
	var args []any

	for column, value := range values {
		if err := ValidColumnName(column); err != nil {
			utility.DefaultLogger.Ferror("Invalid column name: %v", err)
			return "", nil, err
		}
		columns = append(columns, fmt.Sprintf(`"%s"`, column))
		placeholders = append(placeholders, "?")
		args = append(args, value)
	}

	columnList := strings.Join(columns, ", ")
	placeholderList := strings.Join(placeholders, ", ")
	query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, tableName, columnList, placeholderList)

	return query, args, nil
}

// SecureBuildUpdateQuery creates a secure UPDATE query with parameterized queries
func (sqb *SecureQueryBuilder) SecureBuildUpdateQuery(tableName string, id int64, values map[string]any) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	if len(values) == 0 {
		return "", nil, fmt.Errorf("no values provided for update")
	}

	// Validate all column names and build SET clause
	var setStmts []string
	var args []any

	for column, value := range values {
		if err := ValidColumnName(column); err != nil {
			utility.DefaultLogger.Ferror("Invalid column name: %v", err)
			return "", nil, err
		}
		setStmts = append(setStmts, fmt.Sprintf(`"%s" = ?`, column))
		args = append(args, value)
	}

	// Add the ID parameter at the end
	args = append(args, id)

	setClause := strings.Join(setStmts, ", ")
	query := fmt.Sprintf(`UPDATE "%s" SET %s WHERE "id" = ?`, tableName, setClause)

	return query, args, nil
}

// SecureBuildDeleteQuery creates a secure DELETE query with parameterized queries
func (sqb *SecureQueryBuilder) SecureBuildDeleteQuery(tableName string, id int64) (string, []any, error) {
	if err := ValidTableName(tableName); err != nil {
		utility.DefaultLogger.Ferror("Invalid table name: %v", err)
		return "", nil, err
	}

	query := fmt.Sprintf(`DELETE FROM "%s" WHERE "id" = ?`, tableName)
	return query, []any{id}, nil
}

// SecureExecuteModifyQuery executes a secure INSERT, UPDATE, or DELETE query
func (sqb *SecureQueryBuilder) SecureExecuteModifyQuery(query string, args []any) (sql.Result, error) {
	utility.DefaultLogger.Finfo("Executing secure modify query: %s with args: %v", query, args)

	result, err := sqb.db.Exec(query, args...)
	if err != nil {
		utility.DefaultLogger.Ferror("Modify query execution failed: %v", err)
		return nil, fmt.Errorf("modify query execution failed: %w", err)
	}

	return result, nil
}
