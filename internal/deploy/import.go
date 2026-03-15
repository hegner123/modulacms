package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// importMu prevents concurrent imports.
var importMu sync.Mutex

// ImportPayload imports a SyncPayload into the target database using the overwrite strategy.
func ImportPayload(ctx context.Context, cfg config.Config, driver db.DbDriver, payload *SyncPayload, skipBackup bool) (*SyncResult, error) {
	start := time.Now()

	// Pre-import validation (read-only, safe before lock).
	validationErrs := ValidatePayload(payload, driver)
	if len(validationErrs) > 0 {
		return &SyncResult{
			Success:  false,
			Strategy: StrategyOverwrite,
			Errors:   validationErrs,
			Duration: time.Since(start).String(),
		}, fmt.Errorf("pre-import validation failed with %d errors", len(validationErrs))
	}

	// Acquire import lock (non-blocking) BEFORE snapshot/backup to avoid
	// wasted I/O and orphan files when a concurrent import is already running.
	if !importMu.TryLock() {
		return nil, fmt.Errorf("import already in progress")
	}
	defer importMu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	// Save pre-import snapshot.
	snapshotDir := SnapshotDir(cfg)
	snapshotID, err := SaveSnapshot(snapshotDir, payload)
	if err != nil {
		return nil, fmt.Errorf("save pre-import snapshot: %w", err)
	}

	// Optional full backup.
	var backupPath string
	if !skipBackup {
		path, _, bErr := backup.CreateFullBackup(cfg)
		if bErr != nil {
			return nil, fmt.Errorf("pre-import backup: %w", bErr)
		}
		backupPath = path
	}

	ops, err := db.NewDeployOps(driver)
	if err != nil {
		return nil, fmt.Errorf("create deploy ops: %w", err)
	}

	// Look up viewer role before entering atomic block.
	viewerRole, err := driver.GetRoleByLabel("viewer")
	if err != nil {
		return nil, fmt.Errorf("lookup viewer role: %w", err)
	}

	var warnings []string

	err = ops.ImportAtomic(ctx, func(ctx context.Context, ex db.Executor) error {
		// Truncate tables in reverse order (tier 6→1).
		for i := len(DefaultTableSet) - 1; i >= 0; i-- {
			t := DefaultTableSet[i]
			if _, ok := payload.Tables[string(t)]; !ok {
				continue
			}
			if tErr := ops.TruncateTable(ctx, ex, t); tErr != nil {
				return fmt.Errorf("truncate %s: %w", t, tErr)
			}
		}

		// Resolve user refs: create placeholder users for missing ones.
		if len(payload.UserRefs) > 0 {
			if rErr := resolveUserRefs(ctx, ex, ops, payload.UserRefs, viewerRole.RoleID.String()); rErr != nil {
				return fmt.Errorf("resolve user refs: %w", rErr)
			}
		}

		// Bulk insert tables in DefaultTableSet order (tier 1→6).
		for _, t := range DefaultTableSet {
			td, ok := payload.Tables[string(t)]
			if !ok {
				continue
			}
			if len(td.Rows) == 0 {
				continue
			}

			rows, cErr := coerceRows(t, td)
			if cErr != nil {
				return fmt.Errorf("coerce %s: %w", t, cErr)
			}

			if iErr := ops.BulkInsert(ctx, ex, t, td.Columns, rows); iErr != nil {
				return fmt.Errorf("insert %s: %w", t, iErr)
			}
		}

		// Import plugin tables (after core tables).
		var pluginTableNames []string
		for name := range payload.Tables {
			if isPluginTable(name) {
				pluginTableNames = append(pluginTableNames, name)
			}
		}
		sort.Strings(pluginTableNames)

		// Truncate existing plugin tables in reverse sorted order.
		for i := len(pluginTableNames) - 1; i >= 0; i-- {
			name := pluginTableNames[i]
			t := db.DBTable(name)
			if _, iErr := ops.IntrospectColumns(ctx, t); iErr != nil {
				warnings = append(warnings, fmt.Sprintf("plugin table %q not on destination (plugin not installed); skipping", name))
				continue
			}
			if tErr := ops.TruncateTable(ctx, ex, t); tErr != nil {
				return fmt.Errorf("truncate plugin table %s: %w", name, tErr)
			}
		}

		// Insert plugin table data.
		for _, name := range pluginTableNames {
			td := payload.Tables[name]
			if len(td.Rows) == 0 {
				continue
			}
			t := db.DBTable(name)

			destCols, iErr := ops.IntrospectColumns(ctx, t)
			if iErr != nil {
				continue // already warned above
			}

			destColNames := extractColNames(destCols)
			if !columnsMatch(td.Columns, destColNames) {
				warnings = append(warnings, fmt.Sprintf("plugin table %q schema mismatch; skipping", name))
				continue
			}

			intCols := buildIntColumnMap(td.Columns, destCols)
			rows := coerceWithIntMap(td.Rows, intCols)

			if iErr := ops.BulkInsert(ctx, ex, t, td.Columns, rows); iErr != nil {
				return fmt.Errorf("insert plugin table %s: %w", name, iErr)
			}
		}

		// Post-insert FK verification (collect as warnings, don't fail).
		violations, vErr := ops.VerifyForeignKeys(ctx, ex)
		if vErr != nil {
			warnings = append(warnings, fmt.Sprintf("FK verification error: %v", vErr))
		}
		for _, v := range violations {
			warnings = append(warnings, fmt.Sprintf("FK violation in %s (row %s -> %s)", v.Table, v.RowID, v.Parent))
		}

		return nil
	})
	if err != nil {
		return &SyncResult{
			Success:    false,
			Strategy:   StrategyOverwrite,
			BackupPath: backupPath,
			SnapshotID: snapshotID,
			Duration:   time.Since(start).String(),
			Errors: []SyncError{{
				Phase:   "import",
				Message: err.Error(),
			}},
		}, err
	}

	// Record synthetic change event for audit trail.
	manifestJSON, mErr := json.Marshal(payload.Manifest)
	if mErr != nil {
		warnings = append(warnings, fmt.Sprintf("marshal audit manifest: %v", mErr))
	} else {
		if _, ceErr := driver.RecordChangeEvent(db.RecordChangeEventParams{
			EventID:      types.NewEventID(),
			HlcTimestamp: types.HLCNow(),
			NodeID:       types.NodeID(cfg.Node_ID),
			TableName:    "_deploy_sync",
			RecordID:     snapshotID,
			Operation:    types.OpInsert,
			Action:       types.ActionCreate,
			NewValues:    types.JSONData{Data: json.RawMessage(manifestJSON), Valid: true},
		}); ceErr != nil {
			utility.DefaultLogger.Error("record deploy change event", ceErr)
			warnings = append(warnings, fmt.Sprintf("audit event recording failed: %v", ceErr))
		}
	}

	// Build result.
	tablesAffected := make([]string, 0, len(payload.Tables))
	rowCounts := make(map[string]int, len(payload.Tables))
	for name, td := range payload.Tables {
		tablesAffected = append(tablesAffected, name)
		rowCounts[name] = len(td.Rows)
	}

	return &SyncResult{
		Success:        true,
		Strategy:       StrategyOverwrite,
		TablesAffected: tablesAffected,
		RowCounts:      rowCounts,
		BackupPath:     backupPath,
		SnapshotID:     snapshotID,
		Duration:       time.Since(start).String(),
		Warnings:       warnings,
	}, nil
}

// resolveUserRefs ensures that every user_id in userRefs exists in the users table.
// Missing users are created as locked placeholder accounts with the viewer role.
func resolveUserRefs(ctx context.Context, ex db.Executor, ops db.DeployOps, userRefs map[string]string, viewerRoleID string) error {
	now := types.NewTimestamp(time.Now().UTC())

	for userID, username := range userRefs {
		// Check if user already exists. Use ops.Placeholder for backend-appropriate SQL.
		var exists int64
		rows, err := ex.QueryContext(ctx, "SELECT COUNT(*) FROM users WHERE user_id = "+ops.Placeholder(1)+";", userID)
		if err != nil {
			return fmt.Errorf("check user %s: %w", userID, err)
		}
		if rows.Next() {
			if sErr := rows.Scan(&exists); sErr != nil {
				rows.Close()
				return fmt.Errorf("scan user check: %w", sErr)
			}
		}
		rows.Close()

		if exists > 0 {
			continue
		}

		// Create placeholder user via BulkInsert.
		columns := []string{"user_id", "username", "name", "email", "hash", "role", "date_created", "date_modified"}
		row := []any{
			userID,
			"[sync] " + username,
			"",
			username + "@sync.placeholder",
			"", // empty hash = locked account
			viewerRoleID,
			now.String(),
			now.String(),
		}

		if err := ops.BulkInsert(ctx, ex, db.User, columns, [][]any{row}); err != nil {
			return fmt.Errorf("create placeholder user %s: %w", userID, err)
		}
	}

	return nil
}

// coerceRows converts JSON-decoded row values to database-compatible types.
// Specifically: float64 → int64 for integer columns, detected via TableStructMap.
func coerceRows(table db.DBTable, td TableData) ([][]any, error) {
	structType, ok := db.TableStructMap[table]
	if !ok {
		// No struct type info; return rows as-is.
		return td.Rows, nil
	}

	// Build a map of column index → true if the struct field is an integer type.
	intCols := make(map[int]bool)
	for i, col := range td.Columns {
		for j := range structType.NumField() {
			f := structType.Field(j)
			tag := f.Tag.Get("json")
			if idx := findCharIndex(tag, ','); idx != -1 {
				tag = tag[:idx]
			}
			if tag != col {
				continue
			}
			kind := f.Type.Kind()
			if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
				intCols[i] = true
			}
			break
		}
	}

	return coerceWithIntMap(td.Rows, intCols), nil
}

// coerceWithIntMap converts float64 values to int64 at the column indices in intCols.
func coerceWithIntMap(rows [][]any, intCols map[int]bool) [][]any {
	if len(intCols) == 0 {
		return rows
	}

	result := make([][]any, len(rows))
	for i, row := range rows {
		newRow := make([]any, len(row))
		copy(newRow, row)
		for colIdx := range intCols {
			if colIdx >= len(newRow) || newRow[colIdx] == nil {
				continue
			}
			if f, ok := newRow[colIdx].(float64); ok {
				newRow[colIdx] = int64(f)
			}
		}
		result[i] = newRow
	}

	return result
}

// extractColNames returns just the Name field from each ColumnMeta.
func extractColNames(cols []db.ColumnMeta) []string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names
}

// columnsMatch returns true if both slices contain the same column names in the same order.
func columnsMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// buildIntColumnMap returns a map of column indices that correspond to integer columns
// according to destMeta. Used for JSON float64 → int64 coercion on plugin tables.
func buildIntColumnMap(cols []string, destMeta []db.ColumnMeta) map[int]bool {
	metaByName := make(map[string]bool, len(destMeta))
	for _, cm := range destMeta {
		if cm.IsInteger {
			metaByName[cm.Name] = true
		}
	}
	intCols := make(map[int]bool)
	for i, col := range cols {
		if metaByName[col] {
			intCols[i] = true
		}
	}
	return intCols
}

// findCharIndex returns the index of a byte in a string, or -1 if not found.
func findCharIndex(s string, c byte) int {
	for i := range len(s) {
		if s[i] == c {
			return i
		}
	}
	return -1
}
