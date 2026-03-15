package deploy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
)

// crockfordBase32 is the valid character set for Crockford base32 encoding used by ULIDs.
const crockfordBase32 = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ValidatePayload performs pre-import validation on the payload.
// Returns a list of errors; an empty list means the payload is valid.
func ValidatePayload(payload *SyncPayload, targetDriver db.DbDriver) []SyncError {
	var errs []SyncError

	// 1. Verify payload hash matches recomputed hash.
	recomputedHash, err := computePayloadHash(payload.Tables)
	if err != nil {
		errs = append(errs, SyncError{
			Phase:   "validate",
			Message: fmt.Sprintf("failed to recompute payload hash: %v", err),
		})
	} else if recomputedHash != payload.Manifest.PayloadHash {
		errs = append(errs, SyncError{
			Phase:   "validate",
			Message: "payload hash mismatch: data may have been tampered with",
		})
	}

	// 2. Verify row counts match actual row counts in tables.
	for tableName, expectedCount := range payload.Manifest.RowCounts {
		td, ok := payload.Tables[tableName]
		if !ok {
			errs = append(errs, SyncError{
				Table:   tableName,
				Phase:   "validate",
				Message: fmt.Sprintf("table listed in manifest row_counts but missing from payload"),
			})
			continue
		}
		if len(td.Rows) != expectedCount {
			errs = append(errs, SyncError{
				Table:   tableName,
				Phase:   "validate",
				Message: fmt.Sprintf("row count mismatch: manifest says %d, payload has %d", expectedCount, len(td.Rows)),
			})
		}
	}

	// 3. Compute target schema version and compare.
	targetSchemaVersion := computeSchemaVersion(payload.Tables)
	if targetSchemaVersion != payload.Manifest.SchemaVersion {
		errs = append(errs, SyncError{
			Phase:   "validate",
			Message: "schema version mismatch: source and target schemas differ",
		})
	}

	// 4. Validate ULID strings in ID columns (skip plugin tables — they may use non-ULID IDs).
	for tableName, td := range payload.Tables {
		if isPluginTable(tableName) {
			continue
		}
		idCols := findIDColumns(td.Columns)
		for _, row := range td.Rows {
			for _, colIdx := range idCols {
				if colIdx >= len(row) || row[colIdx] == nil {
					continue
				}
				idStr, ok := row[colIdx].(string)
				if !ok || idStr == "" {
					continue
				}
				if !isValidULID(idStr) {
					rowID := extractRowID(row, td.Columns)
					errs = append(errs, SyncError{
						Table:   tableName,
						Phase:   "validate",
						Message: fmt.Sprintf("invalid ULID in column %q: %q", td.Columns[colIdx], idStr),
						RowID:   rowID,
					})
				}
			}
		}
	}

	// 5. Intra-payload FK check: content_data.datatype_id references datatypes.
	errs = append(errs, validateContentDatatypeFK(payload)...)

	// 6. Content tree pointer check.
	errs = append(errs, validateContentTreePointers(payload)...)

	// 7. All author_id values present in UserRefs.
	errs = append(errs, validateUserRefs(payload)...)

	// 8. Row width consistency for plugin tables.
	for tableName, td := range payload.Tables {
		if !isPluginTable(tableName) {
			continue
		}
		for i, row := range td.Rows {
			if len(row) != len(td.Columns) {
				errs = append(errs, SyncError{
					Table:   tableName,
					Phase:   "validate",
					Message: fmt.Sprintf("row %d has %d values but %d columns", i, len(row), len(td.Columns)),
				})
			}
		}
	}

	return errs
}

// VerifyImport performs post-import validation.
// The expected map uses db.DBTable keys which are validated against the table whitelist
// before being used in SQL queries.
func VerifyImport(ctx context.Context, ops db.DeployOps, ex db.Executor, expected map[db.DBTable]int) []SyncError {
	var errs []SyncError

	// 1. FK integrity check.
	violations, err := ops.VerifyForeignKeys(ctx, ex)
	if err != nil {
		errs = append(errs, SyncError{
			Phase:   "verify",
			Message: fmt.Sprintf("FK verification failed: %v", err),
		})
	}
	for _, v := range violations {
		errs = append(errs, SyncError{
			Table:   v.Table,
			Phase:   "verify",
			Message: fmt.Sprintf("FK violation: row references missing parent in %s", v.Parent),
			RowID:   v.RowID,
		})
	}

	// 2. Row count verification.
	for table, expectedCount := range expected {
		if !db.IsValidTable(table) {
			errs = append(errs, SyncError{
				Table:   string(table),
				Phase:   "verify",
				Message: "unknown table name, skipping count verification",
			})
			continue
		}

		var count int64
		row, qErr := ex.QueryContext(ctx, "SELECT COUNT(*) FROM "+string(table)+";")
		if qErr != nil {
			errs = append(errs, SyncError{
				Table:   string(table),
				Phase:   "verify",
				Message: fmt.Sprintf("count query failed: %v", qErr),
			})
			continue
		}
		if row.Next() {
			if sErr := row.Scan(&count); sErr != nil {
				row.Close()
				errs = append(errs, SyncError{
					Table:   string(table),
					Phase:   "verify",
					Message: fmt.Sprintf("count scan failed: %v", sErr),
				})
				continue
			}
		}
		row.Close()

		if int(count) != expectedCount {
			errs = append(errs, SyncError{
				Table:   string(table),
				Phase:   "verify",
				Message: fmt.Sprintf("row count mismatch after import: expected %d, got %d", expectedCount, count),
			})
		}
	}

	return errs
}

// findIDColumns returns indices of columns whose names end with "_id".
func findIDColumns(columns []string) []int {
	var idxs []int
	for i, col := range columns {
		if strings.HasSuffix(col, "_id") {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// isValidULID checks that s is a 26-character Crockford base32 string.
func isValidULID(s string) bool {
	if len(s) != 26 {
		return false
	}
	upper := strings.ToUpper(s)
	for _, c := range upper {
		if !strings.ContainsRune(crockfordBase32, c) {
			return false
		}
	}
	return true
}

// extractRowID returns the first column value as a string for error reporting.
func extractRowID(row []any, columns []string) string {
	if len(row) == 0 || len(columns) == 0 {
		return ""
	}
	if s, ok := row[0].(string); ok {
		return s
	}
	return fmt.Sprintf("%v", row[0])
}

// validateContentDatatypeFK checks that every content_data.datatype_id exists in datatypes.
func validateContentDatatypeFK(payload *SyncPayload) []SyncError {
	var errs []SyncError

	// Build set of known datatype IDs.
	dtIDs := make(map[string]bool)
	for _, tableName := range []string{string(db.Datatype), string(db.Admin_datatype)} {
		td, ok := payload.Tables[tableName]
		if !ok {
			continue
		}
		pkIdx := colIndex(td.Columns, "datatype_id")
		if pkIdx < 0 {
			pkIdx = colIndex(td.Columns, "admin_datatype_id")
		}
		if pkIdx < 0 {
			continue
		}
		for _, row := range td.Rows {
			if pkIdx < len(row) && row[pkIdx] != nil {
				if s, ok := row[pkIdx].(string); ok {
					dtIDs[s] = true
				}
			}
		}
	}

	// Check content_data and admin_content_data.
	for _, tableName := range []string{string(db.Content_data), string(db.Admin_content_data)} {
		td, ok := payload.Tables[tableName]
		if !ok {
			continue
		}
		dtColIdx := colIndex(td.Columns, "datatype_id")
		if dtColIdx < 0 {
			dtColIdx = colIndex(td.Columns, "admin_datatype_id")
		}
		if dtColIdx < 0 {
			continue
		}
		for _, row := range td.Rows {
			if dtColIdx >= len(row) || row[dtColIdx] == nil {
				continue
			}
			if s, ok := row[dtColIdx].(string); ok && s != "" {
				if !dtIDs[s] {
					errs = append(errs, SyncError{
						Table:   tableName,
						Phase:   "validate",
						Message: fmt.Sprintf("datatype_id %q not found in datatypes", s),
						RowID:   extractRowID(row, td.Columns),
					})
				}
			}
		}
	}

	return errs
}

// validateContentTreePointers checks that parent/child/sibling IDs reference existing content_data rows.
func validateContentTreePointers(payload *SyncPayload) []SyncError {
	var errs []SyncError

	pointerCols := []string{"parent_id", "first_child_id", "next_sibling_id", "prev_sibling_id"}

	for _, tableName := range []string{string(db.Content_data), string(db.Admin_content_data)} {
		td, ok := payload.Tables[tableName]
		if !ok {
			continue
		}

		// Build set of known content IDs for this table.
		contentIDs := make(map[string]bool)
		pkIdx := colIndex(td.Columns, "content_data_id")
		if pkIdx < 0 {
			pkIdx = colIndex(td.Columns, "admin_content_data_id")
		}
		if pkIdx < 0 {
			continue
		}
		for _, row := range td.Rows {
			if pkIdx < len(row) && row[pkIdx] != nil {
				if s, ok := row[pkIdx].(string); ok {
					contentIDs[s] = true
				}
			}
		}

		// Check pointer columns.
		for _, colName := range pointerCols {
			idx := colIndex(td.Columns, colName)
			if idx < 0 {
				continue
			}
			for _, row := range td.Rows {
				if idx >= len(row) || row[idx] == nil {
					continue
				}
				if s, ok := row[idx].(string); ok && s != "" {
					if !contentIDs[s] {
						errs = append(errs, SyncError{
							Table:   tableName,
							Phase:   "validate",
							Message: fmt.Sprintf("%s %q not found in %s", colName, s, tableName),
							RowID:   extractRowID(row, td.Columns),
						})
					}
				}
			}
		}
	}

	return errs
}

// validateUserRefs checks that all author_id values in exported tables are present in UserRefs.
func validateUserRefs(payload *SyncPayload) []SyncError {
	var errs []SyncError

	for tableName, td := range payload.Tables {
		idx := colIndex(td.Columns, "author_id")
		if idx < 0 {
			continue
		}
		for _, row := range td.Rows {
			if idx >= len(row) || row[idx] == nil {
				continue
			}
			if s, ok := row[idx].(string); ok && s != "" {
				if _, found := payload.UserRefs[s]; !found {
					errs = append(errs, SyncError{
						Table:   tableName,
						Phase:   "validate",
						Message: fmt.Sprintf("author_id %q not found in user_refs", s),
						RowID:   extractRowID(row, td.Columns),
					})
				}
			}
		}
	}

	return errs
}

// colIndex returns the index of a column name in the columns slice, or -1 if not found.
func colIndex(columns []string, name string) int {
	for i, c := range columns {
		if c == name {
			return i
		}
	}
	return -1
}
