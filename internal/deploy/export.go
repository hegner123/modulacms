package deploy

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// tableListFunc fetches all rows for a single table from the driver.
type tableListFunc func(db.DbDriver) (any, error)

// tableListFuncs maps syncable DBTables to their parameterless List* methods.
// Tables not in this map fall back to QueryAllRows in ExportPayload.
var tableListFuncs = map[db.DBTable]tableListFunc{
	// Schema
	db.Datatype:          func(d db.DbDriver) (any, error) { return d.ListDatatypes() },
	db.Admin_datatype:    func(d db.DbDriver) (any, error) { return d.ListAdminDatatypes() },
	db.Field:             func(d db.DbDriver) (any, error) { return d.ListFields() },
	db.Admin_field:       func(d db.DbDriver) (any, error) { return d.ListAdminFields() },
	db.Field_types:       func(d db.DbDriver) (any, error) { return d.ListFieldTypes() },
	db.Admin_field_types: func(d db.DbDriver) (any, error) { return d.ListAdminFieldTypes() },
	db.Route:             func(d db.DbDriver) (any, error) { return d.ListRoutes() },
	db.Admin_route:       func(d db.DbDriver) (any, error) { return d.ListAdminRoutes() },
	db.ValidationT:       func(d db.DbDriver) (any, error) { return d.ListValidations() },
	db.Admin_validation:  func(d db.DbDriver) (any, error) { return d.ListAdminValidations() },
	// Content
	db.Content_data:            func(d db.DbDriver) (any, error) { return d.ListContentData() },
	db.Admin_content_data:      func(d db.DbDriver) (any, error) { return d.ListAdminContentData() },
	db.Content_fields:          func(d db.DbDriver) (any, error) { return d.ListContentFields() },
	db.Admin_content_fields:    func(d db.DbDriver) (any, error) { return d.ListAdminContentFields() },
	db.Content_relations:       func(d db.DbDriver) (any, error) { return d.ListContentRelations() },
	db.Admin_content_relations: func(d db.DbDriver) (any, error) { return d.ListAdminContentRelations() },
	// Media
	db.MediaT:             func(d db.DbDriver) (any, error) { return d.ListMedia() },
	db.Admin_media:        func(d db.DbDriver) (any, error) { return d.ListAdminMedia() },
	db.Media_dimension:    func(d db.DbDriver) (any, error) { return d.ListMediaDimensions() },
	db.Media_folder:       func(d db.DbDriver) (any, error) { return d.ListMediaFolders() },
	db.Admin_media_folder: func(d db.DbDriver) (any, error) { return d.ListAdminMediaFolders() },
	// Identity
	db.User:             func(d db.DbDriver) (any, error) { return d.ListUsers() },
	db.Role:             func(d db.DbDriver) (any, error) { return d.ListRoles() },
	db.Permission:       func(d db.DbDriver) (any, error) { return d.ListPermissions() },
	db.Role_permissions: func(d db.DbDriver) (any, error) { return d.ListRolePermissions() },
	db.Session:          func(d db.DbDriver) (any, error) { return d.ListSessions() },
	db.Token:            func(d db.DbDriver) (any, error) { return d.ListTokens() },
	// System
	db.LocaleT:             func(d db.DbDriver) (any, error) { return d.ListLocales() },
	db.WebhookT:            func(d db.DbDriver) (any, error) { return d.ListWebhooks() },
	db.Webhook_deliveries:  func(d db.DbDriver) (any, error) { return d.ListWebhookDeliveries() },
	db.PipelineT:           func(d db.DbDriver) (any, error) { return d.ListPipelines() },
	db.Table:               func(d db.DbDriver) (any, error) { return d.ListTables() },
}

// ExportPayload exports data from the driver into a SyncPayload.
// If opts.Tables is nil or empty, DefaultTableSet is used.
// If opts.IncludePlugins is true, registered plugin tables are discovered and included.
func ExportPayload(ctx context.Context, driver db.DbDriver, opts ExportOptions) (*SyncPayload, error) {
	tables := opts.Tables
	if len(tables) == 0 {
		tables = DefaultTableSet
	}

	tableDataMap := make(map[string]TableData, len(tables))
	tableNames := make([]string, 0, len(tables))
	rowCounts := make(map[string]int, len(tables))

	// DeployOps is lazily created only when needed for QueryAllRows fallback.
	var ops db.DeployOps

	for _, t := range tables {
		name := string(t)

		if listFn, ok := tableListFuncs[t]; ok {
			// Typed export via struct serialization.
			slicePtr, err := listFn(driver)
			if err != nil {
				return nil, fmt.Errorf("export %s: %w", t, err)
			}
			td, err := structSliceToTableData(slicePtr)
			if err != nil {
				return nil, fmt.Errorf("export serialize %s: %w", t, err)
			}
			tableDataMap[name] = td
			tableNames = append(tableNames, name)
			rowCounts[name] = len(td.Rows)
		} else {
			// Fallback to catalog-based export (same as plugin tables).
			if ops == nil {
				var err error
				ops, err = db.NewDeployOps(driver)
				if err != nil {
					return nil, fmt.Errorf("export: create deploy ops: %w", err)
				}
			}
			cols, rows, err := ops.QueryAllRows(ctx, t)
			if err != nil {
				return nil, fmt.Errorf("export %s (QueryAllRows): %w", t, err)
			}
			tableDataMap[name] = TableData{Columns: cols, Rows: rows}
			tableNames = append(tableNames, name)
			rowCounts[name] = len(rows)
		}
	}

	// Export plugin tables via catalog introspection (no struct type needed).
	var pluginNames []string
	if opts.IncludePlugins {
		ops, err := db.NewDeployOps(driver)
		if err != nil {
			return nil, fmt.Errorf("export: create deploy ops for plugin tables: %w", err)
		}
		pluginTables, err := discoverPluginTables(driver)
		if err != nil {
			return nil, fmt.Errorf("export: discover plugin tables: %w", err)
		}

		for _, pt := range pluginTables {
			cols, rows, err := ops.QueryAllRows(ctx, pt)
			if err != nil {
				utility.DefaultLogger.Warn("export: skipping plugin table "+string(pt), err)
				continue
			}

			td := TableData{Columns: cols, Rows: rows}
			name := string(pt)
			tableDataMap[name] = td
			tableNames = append(tableNames, name)
			rowCounts[name] = len(rows)
			pluginNames = append(pluginNames, name)
		}
	}

	userRefs := collectUserRefs(tableDataMap, driver)

	schemaVersion := computeSchemaVersion(tableDataMap)
	payloadHash, err := computePayloadHash(tableDataMap)
	if err != nil {
		return nil, fmt.Errorf("export compute hash: %w", err)
	}

	manifest := SyncManifest{
		SchemaVersion: schemaVersion,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Version:       utility.GetCurrentVersion(),
		Strategy:      StrategyOverwrite,
		Tables:        tableNames,
		RowCounts:     rowCounts,
		PayloadHash:   payloadHash,
		PluginTables:  pluginNames,
	}

	return &SyncPayload{
		Manifest: manifest,
		Tables:   tableDataMap,
		UserRefs: userRefs,
	}, nil
}

// isPluginTable reports whether a table name follows the plugin table naming convention.
func isPluginTable(name string) bool {
	return db.IsValidPluginTableName(name)
}

// discoverPluginTables finds user-created plugin tables registered in the tables table,
// excluding system plugin tables (plugin_routes, plugin_hooks, plugin_requests).
func discoverPluginTables(driver db.DbDriver) ([]db.DBTable, error) {
	tables, err := driver.ListTables()
	if err != nil {
		return nil, err
	}

	system := db.SystemPluginTables

	var result []db.DBTable
	if tables != nil {
		for _, t := range *tables {
			if !isPluginTable(t.Label) || system[t.Label] {
				continue
			}
			result = append(result, db.DBTable(t.Label))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

// columnsFromType extracts ordered column names from struct json tags.
func columnsFromType(t reflect.Type) []string {
	cols := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		// Handle "name,omitempty" style tags.
		if idx := strings.IndexByte(tag, ','); idx != -1 {
			tag = tag[:idx]
		}
		cols = append(cols, tag)
	}
	return cols
}

// serializeField converts a typed struct field value to a JSON-friendly any.
func serializeField(v reflect.Value) any {
	iface := v.Interface()

	// Handle fmt.Stringer types (typed IDs, Timestamps, Nullable*, Slug, Email, etc.)
	if s, ok := iface.(fmt.Stringer); ok {
		str := s.String()
		if str == "null" {
			return nil
		}
		return str
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.String:
		return v.String()
	default:
		return iface
	}
}

// structSliceToTableData converts a *[]T (pointer to slice of structs) into TableData.
func structSliceToTableData(slicePtr any) (TableData, error) {
	if slicePtr == nil {
		return TableData{Columns: []string{}, Rows: [][]any{}}, nil
	}

	ptrVal := reflect.ValueOf(slicePtr)
	if ptrVal.Kind() != reflect.Ptr {
		return TableData{}, fmt.Errorf("expected pointer, got %T", slicePtr)
	}
	if ptrVal.IsNil() {
		return TableData{Columns: []string{}, Rows: [][]any{}}, nil
	}

	sliceVal := ptrVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return TableData{}, fmt.Errorf("expected pointer to slice, got pointer to %s", sliceVal.Kind())
	}

	elemType := sliceVal.Type().Elem()
	cols := columnsFromType(elemType)

	rows := make([][]any, 0, sliceVal.Len())
	for i := range sliceVal.Len() {
		structVal := sliceVal.Index(i)
		row := make([]any, 0, len(cols))
		fieldIdx := 0
		for j := range elemType.NumField() {
			f := elemType.Field(j)
			tag := f.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			row = append(row, serializeField(structVal.Field(j)))
			fieldIdx++
		}
		rows = append(rows, row)
	}

	return TableData{Columns: cols, Rows: rows}, nil
}

// collectUserRefs scans exported tables for user ID columns and resolves them
// to usernames via driver.ListUsers().
func collectUserRefs(tables map[string]TableData, driver db.DbDriver) map[string]string {
	refs := make(map[string]string)

	// Collect all user ID values from the exported data.
	userIDs := make(map[string]bool)
	for tableName, td := range tables {
		isPlugin := isPluginTable(tableName)
		for colIdx, col := range td.Columns {
			// For plugin tables, only collect author_id (user_id may reference external systems).
			if col == "user_id" && isPlugin {
				continue
			}
			if col != "author_id" && col != "user_id" {
				continue
			}
			for _, row := range td.Rows {
				if colIdx < len(row) && row[colIdx] != nil {
					if idStr, ok := row[colIdx].(string); ok && idStr != "" {
						userIDs[idStr] = true
					}
				}
			}
		}
	}

	if len(userIDs) == 0 {
		return refs
	}

	// Resolve user IDs to usernames.
	users, err := driver.ListUsers()
	if err != nil || users == nil {
		return refs
	}

	for _, u := range *users {
		uid := u.UserID.String()
		if userIDs[uid] {
			refs[uid] = u.Username
		}
	}

	return refs
}

// computeSchemaVersion produces a SHA256 hash of sorted "table:col1,col2,...\n" pairs.
func computeSchemaVersion(tables map[string]TableData) string {
	lines := make([]string, 0, len(tables))
	for name, td := range tables {
		lines = append(lines, name+":"+strings.Join(td.Columns, ",")+"\n")
	}
	sort.Strings(lines)

	h := sha256.New()
	for _, line := range lines {
		h.Write([]byte(line))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// computePayloadHash produces a SHA256 hash of the JSON-encoded tables map.
func computePayloadHash(tables map[string]TableData) (string, error) {
	data, err := json.Marshal(tables)
	if err != nil {
		return "", fmt.Errorf("marshal tables for hash: %w", err)
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h), nil
}
