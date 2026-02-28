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

// tableListFuncs maps each syncable DBTable to its parameterless List* method.
var tableListFuncs = map[db.DBTable]tableListFunc{
	db.Datatype:                func(d db.DbDriver) (any, error) { return d.ListDatatypes() },
	db.Admin_datatype:          func(d db.DbDriver) (any, error) { return d.ListAdminDatatypes() },
	db.Field:                   func(d db.DbDriver) (any, error) { return d.ListFields() },
	db.Admin_field:             func(d db.DbDriver) (any, error) { return d.ListAdminFields() },
	db.Route:                   func(d db.DbDriver) (any, error) { return d.ListRoutes() },
	db.Admin_route:             func(d db.DbDriver) (any, error) { return d.ListAdminRoutes() },
	db.Content_data:            func(d db.DbDriver) (any, error) { return d.ListContentData() },
	db.Admin_content_data:      func(d db.DbDriver) (any, error) { return d.ListAdminContentData() },
	db.Content_fields:          func(d db.DbDriver) (any, error) { return d.ListContentFields() },
	db.Admin_content_fields:    func(d db.DbDriver) (any, error) { return d.ListAdminContentFields() },
	db.Content_relations:       func(d db.DbDriver) (any, error) { return d.ListContentRelations() },
	db.Admin_content_relations: func(d db.DbDriver) (any, error) { return d.ListAdminContentRelations() },
}

// ExportPayload exports data from the driver into a SyncPayload.
// If tables is nil or empty, DefaultTableSet is used.
func ExportPayload(_ context.Context, driver db.DbDriver, tables []db.DBTable) (*SyncPayload, error) {
	if len(tables) == 0 {
		tables = DefaultTableSet
	}

	tableDataMap := make(map[string]TableData, len(tables))
	tableNames := make([]string, 0, len(tables))
	rowCounts := make(map[string]int, len(tables))

	for _, t := range tables {
		listFn, ok := tableListFuncs[t]
		if !ok {
			return nil, fmt.Errorf("export: no list function for table %q", string(t))
		}

		slicePtr, err := listFn(driver)
		if err != nil {
			return nil, fmt.Errorf("export %s: %w", t, err)
		}

		td, err := structSliceToTableData(slicePtr)
		if err != nil {
			return nil, fmt.Errorf("export serialize %s: %w", t, err)
		}

		name := string(t)
		tableDataMap[name] = td
		tableNames = append(tableNames, name)
		rowCounts[name] = len(td.Rows)
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
	}

	return &SyncPayload{
		Manifest: manifest,
		Tables:   tableDataMap,
		UserRefs: userRefs,
	}, nil
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

	// Columns that contain user IDs.
	userCols := map[string]bool{
		"author_id": true,
		"user_id":   true,
	}

	// Collect all user ID values from the exported data.
	userIDs := make(map[string]bool)
	for _, td := range tables {
		for colIdx, col := range td.Columns {
			if !userCols[col] {
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
