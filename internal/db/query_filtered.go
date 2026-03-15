package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Safety limits for filtered queries and bulk insert.
const (
	MaxOrderByCols    = 8    // max columns in ORDER BY
	MaxBulkInsertRows = 1000 // absolute max rows per QBulkInsert call
)

// OrderByColumn defines a single column in a multi-column ORDER BY for filtered queries.
type OrderByColumn struct {
	Column string
	Desc   bool
}

// FilteredSelectParams configures a SELECT query using the Condition system.
type FilteredSelectParams struct {
	Table       string
	Columns     []string           // nil = SELECT *
	Aggregates  []AggregateColumn  // aggregate expressions; appended after Columns in SELECT
	Filter      Condition          // required unless Aggregates or GroupBy are present
	GroupBy     []string           // GROUP BY column names (validated)
	Having      Condition          // HAVING condition; requires GroupBy
	OrderByCols []OrderByColumn    // empty = no ORDER BY
	Distinct    bool
	Limit       int64 // 0 = default (maxLimit); negative = no limit
	Offset      int64
}

// FilteredUpdateParams configures an UPDATE query using the Condition system.
type FilteredUpdateParams struct {
	Table  string
	Set    map[string]any // must be non-empty
	Filter Condition      // required
}

// FilteredDeleteParams configures a DELETE query using the Condition system.
type FilteredDeleteParams struct {
	Table  string
	Filter Condition // required
}

// BulkInsertParams configures a multi-row INSERT query.
type BulkInsertParams struct {
	Table   string
	Columns []string
	Rows    [][]any // each inner slice has len(Columns) values; nil elements = SQL NULL
}

// QSelectFiltered executes a SELECT query using the Condition system.
func QSelectFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredSelectParams) ([]Row, error) {
	query, args, err := buildFilteredSelectQuery(d, p)
	if err != nil {
		return nil, err
	}
	return execQuery(ctx, exec, query, args)
}

// QSelectOneFiltered executes a SELECT query and returns the first matching row, or nil if no match.
func QSelectOneFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredSelectParams) (Row, error) {
	p.Limit = 1
	rows, err := QSelectFiltered(ctx, exec, d, p)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// QUpdateFiltered executes an UPDATE query using the Condition system.
// Filter must be non-nil and contain at least one value binding.
func QUpdateFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredUpdateParams) (sql.Result, error) {
	if err := ValidTableName(p.Table); err != nil {
		return nil, fmt.Errorf("invalid table: %w", err)
	}
	if len(p.Set) == 0 {
		return nil, fmt.Errorf("update requires non-empty set")
	}
	if p.Filter == nil {
		return nil, fmt.Errorf("update requires a non-nil filter condition")
	}
	if err := ValidateCondition(p.Filter); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}
	if !HasValueBinding(p.Filter) {
		return nil, fmt.Errorf("update requires a filter with at least one value binding (e.g., Compare, In, Between)")
	}

	// Build SET clause
	setKeys := sortedKeys(p.Set)
	setClauses := make([]string, len(setKeys))
	var args []any
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

	// Build WHERE clause from filter, threading argOffset through SET
	bctx := NewBuildContext()
	whereSQL, whereArgs, _, err := p.Filter.Build(bctx, d, argIdx)
	if err != nil {
		return nil, fmt.Errorf("filter build: %w", err)
	}

	query := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`,
		quoteIdent(d, p.Table),
		strings.Join(setClauses, ", "),
		whereSQL,
	)
	args = append(args, whereArgs...)

	return exec.ExecContext(ctx, query, args...)
}

// QDeleteFiltered executes a DELETE query using the Condition system.
// Filter must be non-nil and contain at least one value binding.
func QDeleteFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredDeleteParams) (sql.Result, error) {
	if err := ValidTableName(p.Table); err != nil {
		return nil, fmt.Errorf("invalid table: %w", err)
	}
	if p.Filter == nil {
		return nil, fmt.Errorf("delete requires a non-nil filter condition")
	}
	if err := ValidateCondition(p.Filter); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}
	if !HasValueBinding(p.Filter) {
		return nil, fmt.Errorf("delete requires a filter with at least one value binding (e.g., Compare, In, Between)")
	}

	bctx := NewBuildContext()
	whereSQL, whereArgs, _, err := p.Filter.Build(bctx, d, 1)
	if err != nil {
		return nil, fmt.Errorf("filter build: %w", err)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE %s`, quoteIdent(d, p.Table), whereSQL)
	return exec.ExecContext(ctx, query, whereArgs...)
}

// QCountFiltered returns the count of rows matching the condition.
func QCountFiltered(ctx context.Context, exec Executor, d Dialect, table string, filter Condition) (int64, error) {
	if err := ValidTableName(table); err != nil {
		return 0, fmt.Errorf("invalid table: %w", err)
	}
	if filter == nil {
		return 0, fmt.Errorf("count requires a non-nil filter condition")
	}
	if err := ValidateCondition(filter); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}

	bctx := NewBuildContext()
	whereSQL, whereArgs, _, err := filter.Build(bctx, d, 1)
	if err != nil {
		return 0, fmt.Errorf("filter build: %w", err)
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, quoteIdent(d, table), whereSQL)
	rows, err := exec.QueryContext(ctx, query, whereArgs...)
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

// QExistsFiltered returns true if at least one row matches the condition.
// Uses SELECT 1 ... LIMIT 1 for early termination.
func QExistsFiltered(ctx context.Context, exec Executor, d Dialect, table string, filter Condition) (bool, error) {
	if err := ValidTableName(table); err != nil {
		return false, fmt.Errorf("invalid table: %w", err)
	}
	if filter == nil {
		return false, fmt.Errorf("exists requires a non-nil filter condition")
	}
	if err := ValidateCondition(filter); err != nil {
		return false, fmt.Errorf("invalid filter: %w", err)
	}

	bctx := NewBuildContext()
	whereSQL, whereArgs, _, err := filter.Build(bctx, d, 1)
	if err != nil {
		return false, fmt.Errorf("filter build: %w", err)
	}

	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE %s LIMIT 1`, quoteIdent(d, table), whereSQL)
	rows, err := exec.QueryContext(ctx, query, whereArgs...)
	if err != nil {
		return false, fmt.Errorf("exists query failed: %w", err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

// QBulkInsert executes a multi-row INSERT with dialect-aware dynamic batch sizing.
// When multiple batches are needed, wraps them in a transaction (if exec is *sql.DB)
// or executes within the existing transaction (if exec is *sql.Tx).
func QBulkInsert(ctx context.Context, exec Executor, d Dialect, p BulkInsertParams) (sql.Result, error) {
	if err := validateBulkInsertParams(p); err != nil {
		return nil, fmt.Errorf("bulk insert: %w", err)
	}

	batchSize := bulkBatchSize(d, len(p.Columns))
	totalRows := len(p.Rows)

	if totalRows <= batchSize {
		// Single batch: no transaction wrapping needed
		return execBulkInsertBatch(ctx, exec, d, p.Table, p.Columns, p.Rows, 0)
	}

	// Multi-batch: need transaction wrapping
	var txExec Executor
	var tx *sql.Tx
	var ownsTx bool

	switch e := exec.(type) {
	case *sql.Tx:
		txExec = e
	case *sql.DB:
		var err error
		tx, err = e.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("bulk insert begin tx: %w", err)
		}
		txExec = tx
		ownsTx = true
	default:
		return nil, fmt.Errorf("QBulkInsert with multiple batches requires *sql.DB or *sql.Tx, got %T", exec)
	}

	var totalAffected int64
	for offset := 0; offset < totalRows; offset += batchSize {
		end := offset + batchSize
		if end > totalRows {
			end = totalRows
		}
		result, err := execBulkInsertBatch(ctx, txExec, d, p.Table, p.Columns, p.Rows[offset:end], offset)
		if err != nil {
			if ownsTx {
				rbErr := tx.Rollback()
				if rbErr != nil {
					return nil, fmt.Errorf("bulk insert batch at offset %d: %w (rollback also failed: %v)", offset, err, rbErr)
				}
			}
			return nil, fmt.Errorf("bulk insert batch at offset %d: %w", offset, err)
		}
		affected, raErr := result.RowsAffected()
		if raErr == nil {
			totalAffected += affected
		}
	}

	if ownsTx {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("bulk insert commit: %w", err)
		}
	}

	return compositeResult{rowsAffected: totalAffected}, nil
}

// buildFilteredSelectQuery constructs the SQL and args for a filtered SELECT.
func buildFilteredSelectQuery(d Dialect, p FilteredSelectParams) (string, []any, error) {
	if err := ValidTableName(p.Table); err != nil {
		return "", nil, fmt.Errorf("invalid table: %w", err)
	}

	// Filter is required unless aggregates or GroupBy are present.
	hasAggregate := len(p.Aggregates) > 0 || len(p.GroupBy) > 0
	if p.Filter == nil && !hasAggregate {
		return "", nil, fmt.Errorf("select requires a non-nil filter condition")
	}
	if p.Filter != nil {
		if err := ValidateCondition(p.Filter); err != nil {
			return "", nil, fmt.Errorf("invalid filter: %w", err)
		}
	}
	if len(p.OrderByCols) > MaxOrderByCols {
		return "", nil, fmt.Errorf("too many order_by columns: %d (max %d)", len(p.OrderByCols), MaxOrderByCols)
	}

	// Column count cap
	if len(p.Columns)+len(p.Aggregates) > maxSelectColumns {
		return "", nil, fmt.Errorf("too many select columns: %d (max %d)", len(p.Columns)+len(p.Aggregates), maxSelectColumns)
	}

	// HAVING requires GROUP BY
	if p.Having != nil && len(p.GroupBy) == 0 {
		return "", nil, fmt.Errorf("having requires group_by")
	}

	// SELECT [DISTINCT] columns
	selectKw := "SELECT"
	if p.Distinct {
		selectKw = "SELECT DISTINCT"
	}

	cols := "*"
	if len(p.Columns) > 0 || len(p.Aggregates) > 0 {
		var parts []string
		for _, c := range p.Columns {
			if err := ValidColumnName(c); err != nil {
				return "", nil, fmt.Errorf("invalid column %q: %w", c, err)
			}
			parts = append(parts, quoteIdent(d, c))
		}
		for _, a := range p.Aggregates {
			expr, aerr := buildAggregateExpr(d, a)
			if aerr != nil {
				return "", nil, fmt.Errorf("invalid aggregate: %w", aerr)
			}
			parts = append(parts, expr)
		}
		cols = strings.Join(parts, ", ")
	}

	query := fmt.Sprintf(`%s %s FROM %s`, selectKw, cols, quoteIdent(d, p.Table))

	// WHERE (optional when aggregates/GroupBy are present)
	var args []any
	nextOffset := 1
	if p.Filter != nil {
		bctx := NewBuildContext()
		whereSQL, whereArgs, noff, err := p.Filter.Build(bctx, d, 1)
		if err != nil {
			return "", nil, fmt.Errorf("filter build: %w", err)
		}
		query += " WHERE " + whereSQL
		args = whereArgs
		nextOffset = noff
	}

	// GROUP BY
	if len(p.GroupBy) > 0 {
		groupParts := make([]string, len(p.GroupBy))
		for i, c := range p.GroupBy {
			if err := ValidColumnName(c); err != nil {
				return "", nil, fmt.Errorf("invalid group_by column %q: %w", c, err)
			}
			groupParts[i] = quoteIdent(d, c)
		}
		query += " GROUP BY " + strings.Join(groupParts, ", ")
	}

	// HAVING
	if p.Having != nil {
		if err := ValidateCondition(p.Having); err != nil {
			return "", nil, fmt.Errorf("invalid having: %w", err)
		}
		bctx := NewBuildContext()
		havingSQL, havingArgs, noff, err := p.Having.Build(bctx, d, nextOffset)
		if err != nil {
			return "", nil, fmt.Errorf("having build: %w", err)
		}
		query += " HAVING " + havingSQL
		args = append(args, havingArgs...)
		nextOffset = noff
	}

	// ORDER BY
	if len(p.OrderByCols) > 0 {
		orderParts := make([]string, len(p.OrderByCols))
		for i, o := range p.OrderByCols {
			if err := ValidColumnName(o.Column); err != nil {
				return "", nil, fmt.Errorf("invalid order_by column %q: %w", o.Column, err)
			}
			dir := "ASC"
			if o.Desc {
				dir = "DESC"
			}
			orderParts[i] = quoteIdent(d, o.Column) + " " + dir
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	}

	// LIMIT
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

	// OFFSET
	if p.Offset > 0 {
		query += fmt.Sprintf(` OFFSET %d`, p.Offset)
	}

	return query, args, nil
}

// bulkBatchSize returns the number of rows per INSERT statement for the given dialect and column count.
func bulkBatchSize(d Dialect, numCols int) int {
	if numCols <= 0 {
		return 1
	}
	switch d {
	case DialectSQLite:
		perCol := 999 / numCols
		if perCol > 100 {
			return 100
		}
		if perCol < 1 {
			return 1
		}
		return perCol
	case DialectMySQL:
		return 500
	case DialectPostgres:
		return 1000
	default:
		return 100
	}
}

// validateBulkInsertParams validates all fields of BulkInsertParams.
func validateBulkInsertParams(p BulkInsertParams) error {
	if err := ValidTableName(p.Table); err != nil {
		return fmt.Errorf("invalid table: %w", err)
	}
	if len(p.Columns) == 0 {
		return fmt.Errorf("columns cannot be empty")
	}
	for _, col := range p.Columns {
		if err := ValidColumnName(col); err != nil {
			return fmt.Errorf("invalid column %q: %w", col, err)
		}
	}
	if len(p.Rows) == 0 {
		return fmt.Errorf("rows cannot be empty")
	}
	if len(p.Rows) > MaxBulkInsertRows {
		return fmt.Errorf("too many rows: %d (max %d)", len(p.Rows), MaxBulkInsertRows)
	}
	numCols := len(p.Columns)
	for i, row := range p.Rows {
		if len(row) != numCols {
			return fmt.Errorf("row %d has %d values, expected %d", i, len(row), numCols)
		}
	}
	return nil
}

// execBulkInsertBatch executes a single INSERT batch.
func execBulkInsertBatch(ctx context.Context, exec Executor, d Dialect, table string, columns []string, rows [][]any, batchOffset int) (sql.Result, error) {
	numCols := len(columns)
	quotedCols := make([]string, numCols)
	for i, col := range columns {
		quotedCols[i] = quoteIdent(d, col)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES ", quoteIdent(d, table), strings.Join(quotedCols, ", ")))

	var args []any
	argIdx := 1

	for i, row := range rows {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteByte('(')
		for j, val := range row {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(placeholder(d, argIdx))
			args = append(args, val)
			argIdx++
		}
		sb.WriteByte(')')
	}

	return exec.ExecContext(ctx, sb.String(), args...)
}

// compositeResult implements sql.Result with an accumulated RowsAffected count.
type compositeResult struct {
	rowsAffected int64
}

func (r compositeResult) LastInsertId() (int64, error) {
	return 0, fmt.Errorf("LastInsertId not available for bulk insert")
}

func (r compositeResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

