package deploy

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// coerceRows edge cases
// ---------------------------------------------------------------------------

func TestCoerceRows_Float64ToInt64(t *testing.T) {
	t.Parallel()

	// Fields has SortOrder int64 -- the canonical coercion target.
	td := TableData{
		Columns: []string{"field_id", "parent_id", "name", "sort_order"},
		Rows: [][]any{
			{"abc", "DT01", "F01", float64(42)},
			{"def", "DT02", "F02", float64(0)},
			{"ghi", "DT03", "F03", float64(-7)},
		},
	}

	rows, err := coerceRows(db.Field, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}

	for i, row := range rows {
		val, ok := row[3].(int64)
		if !ok {
			t.Errorf("row %d: sort_order type = %T, want int64", i, row[3])
			continue
		}
		expected := td.Rows[i][3].(float64)
		if val != int64(expected) {
			t.Errorf("row %d: sort_order = %d, want %d", i, val, int64(expected))
		}
	}
}

func TestCoerceRows_NilInIntegerColumn(t *testing.T) {
	t.Parallel()

	// nil values in an integer column should pass through unchanged.
	td := TableData{
		Columns: []string{"field_id", "parent_id", "name", "sort_order"},
		Rows: [][]any{
			{"abc", "DT01", "F01", nil},
		},
	}

	rows, err := coerceRows(db.Field, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}
	if rows[0][3] != nil {
		t.Errorf("expected nil for sort_order, got %v (%T)", rows[0][3], rows[0][3])
	}
}

func TestCoerceRows_MixedTypes(t *testing.T) {
	t.Parallel()

	// Row with a mix of float64 (needs coercion), string, and nil.
	td := TableData{
		Columns: []string{"field_id", "parent_id", "name", "sort_order"},
		Rows: [][]any{
			{"abc", "DT01", "F01", float64(100)},
			{"def", nil, "F02", float64(200)},
			{"ghi", "DT03", nil, nil},
		},
	}

	rows, err := coerceRows(db.Field, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}

	// First row: sort_order should be int64(100)
	if v, ok := rows[0][3].(int64); !ok || v != 100 {
		t.Errorf("row 0 sort_order = %v (%T), want int64(100)", rows[0][3], rows[0][3])
	}

	// Second row: parent_id is nil (non-integer column), sort_order coerced
	if rows[1][1] != nil {
		t.Errorf("row 1 parent_id = %v, want nil", rows[1][1])
	}
	if v, ok := rows[1][3].(int64); !ok || v != 200 {
		t.Errorf("row 1 sort_order = %v (%T), want int64(200)", rows[1][3], rows[1][3])
	}

	// Third row: sort_order is nil
	if rows[2][3] != nil {
		t.Errorf("row 2 sort_order = %v, want nil", rows[2][3])
	}
}

func TestCoerceRows_StringInIntegerColumn(t *testing.T) {
	t.Parallel()

	// A string value in an integer column should NOT be coerced (only float64 is).
	td := TableData{
		Columns: []string{"field_id", "parent_id", "name", "sort_order"},
		Rows: [][]any{
			{"abc", "DT01", "F01", "not-a-number"},
		},
	}

	rows, err := coerceRows(db.Field, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}

	if s, ok := rows[0][3].(string); !ok || s != "not-a-number" {
		t.Errorf("expected string passthrough, got %v (%T)", rows[0][3], rows[0][3])
	}
}

func TestCoerceRows_NoIntegerColumns(t *testing.T) {
	t.Parallel()

	// Permissions has no int fields -- coerceRows should return rows as-is.
	td := TableData{
		Columns: []string{"permission_id", "label"},
		Rows: [][]any{
			{"abc", "content:read"},
			{"def", "content:write"},
		},
	}

	rows, err := coerceRows(db.Permission, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}

	// When no int columns exist, the original slice is returned.
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "abc" {
		t.Errorf("row[0][0] = %v, want abc", rows[0][0])
	}
}

func TestCoerceRows_MissingStructMapping(t *testing.T) {
	t.Parallel()

	// A table not in TableStructMap should return rows unchanged.
	unknownTable := db.DBTable("nonexistent_table")
	td := TableData{
		Columns: []string{"a", "b"},
		Rows: [][]any{
			{float64(1), "hello"},
		},
	}

	rows, err := coerceRows(unknownTable, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}

	// float64 should remain float64 since no struct mapping exists.
	if _, ok := rows[0][0].(float64); !ok {
		t.Errorf("expected float64 passthrough for unknown table, got %T", rows[0][0])
	}
}

func TestCoerceRows_ShortRow(t *testing.T) {
	t.Parallel()

	// A row shorter than the column list. The integer column is past the
	// end of the row -- coerceRows must not panic.
	td := TableData{
		Columns: []string{"field_id", "parent_id", "name", "sort_order"},
		Rows: [][]any{
			{"abc", "DT01"}, // only 2 elements, sort_order index (3) out of range
		},
	}

	rows, err := coerceRows(db.Field, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}
	if len(rows[0]) != 2 {
		t.Errorf("expected row length 2, got %d", len(rows[0]))
	}
}

func TestCoerceRows_EmptyRows(t *testing.T) {
	t.Parallel()

	td := TableData{
		Columns: []string{"field_id", "parent_id", "name", "sort_order"},
		Rows:    [][]any{},
	}

	rows, err := coerceRows(db.Field, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestCoerceRows_IntAlreadyInt64(t *testing.T) {
	t.Parallel()

	// If the value is already int64, it should remain int64.
	td := TableData{
		Columns: []string{"field_id", "parent_id", "name", "sort_order"},
		Rows: [][]any{
			{"abc", "DT01", "F01", int64(99)},
		},
	}

	rows, err := coerceRows(db.Field, td)
	if err != nil {
		t.Fatalf("coerceRows error: %v", err)
	}

	if v, ok := rows[0][3].(int64); !ok || v != 99 {
		t.Errorf("expected int64(99), got %v (%T)", rows[0][3], rows[0][3])
	}
}

// ---------------------------------------------------------------------------
// Gzip round-trip at threshold boundary
// ---------------------------------------------------------------------------

func TestGzipRoundTrip_BelowThreshold(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outPath := filepath.Join(dir, "export.json")

	// Create a small payload that stays below gzipThreshold.
	payload := &SyncPayload{
		Manifest: SyncManifest{
			SchemaVersion: "abc123",
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Strategy:      StrategyOverwrite,
			Tables:        []string{"test_table"},
			RowCounts:     map[string]int{"test_table": 1},
		},
		Tables: map[string]TableData{
			"test_table": {
				Columns: []string{"id", "name"},
				Rows:    [][]any{{"1", "hello"}},
			},
		},
		UserRefs: map[string]string{},
	}

	// Compute real hashes.
	hash, err := computePayloadHash(payload.Tables)
	if err != nil {
		t.Fatalf("computePayloadHash: %v", err)
	}
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	// Write to file directly (simulating what ExportToFile does for small payloads).
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(outPath, data, 0640); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read back via ReadPayloadFile.
	readBack, err := ReadPayloadFile(outPath)
	if err != nil {
		t.Fatalf("ReadPayloadFile: %v", err)
	}

	var decoded SyncPayload
	if err := json.Unmarshal(readBack, &decoded); err != nil {
		t.Fatalf("unmarshal read-back: %v", err)
	}

	if len(decoded.Tables["test_table"].Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(decoded.Tables["test_table"].Rows))
	}
	if decoded.Manifest.PayloadHash != payload.Manifest.PayloadHash {
		t.Error("payload hash mismatch after round-trip")
	}
}

func TestGzipRoundTrip_Compressed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gzPath := filepath.Join(dir, "export.json.gz")

	// Create a payload, write it with gzip, read it back.
	payload := &SyncPayload{
		Manifest: SyncManifest{
			SchemaVersion: "abc123",
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Strategy:      StrategyOverwrite,
			Tables:        []string{"test_table"},
			RowCounts:     map[string]int{"test_table": 2},
		},
		Tables: map[string]TableData{
			"test_table": {
				Columns: []string{"id", "value"},
				Rows: [][]any{
					{"row1", "data1"},
					{"row2", "data2"},
				},
			},
		},
		UserRefs: map[string]string{},
	}

	hash, err := computePayloadHash(payload.Tables)
	if err != nil {
		t.Fatalf("computePayloadHash: %v", err)
	}
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Write as gzip.
	if err := writeGzipFile(gzPath, data); err != nil {
		t.Fatalf("writeGzipFile: %v", err)
	}

	// Read back using ReadPayloadFile which detects .gz extension.
	readBack, err := ReadPayloadFile(gzPath)
	if err != nil {
		t.Fatalf("ReadPayloadFile: %v", err)
	}

	var decoded SyncPayload
	if err := json.Unmarshal(readBack, &decoded); err != nil {
		t.Fatalf("unmarshal round-trip: %v", err)
	}

	if len(decoded.Tables["test_table"].Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(decoded.Tables["test_table"].Rows))
	}
	if decoded.Manifest.PayloadHash != payload.Manifest.PayloadHash {
		t.Error("payload hash mismatch after gzip round-trip")
	}
}

func TestGzipDetection_MagicBytes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write a gzip file WITHOUT .gz extension.
	path := filepath.Join(dir, "export.json")

	original := []byte(`{"manifest":{},"tables":{},"user_refs":{}}`)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(original); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0640); err != nil {
		t.Fatalf("write: %v", err)
	}

	// ReadPayloadFile should detect gzip magic bytes even without .gz extension.
	data, err := ReadPayloadFile(path)
	if err != nil {
		t.Fatalf("ReadPayloadFile: %v", err)
	}

	if !bytes.Equal(data, original) {
		t.Errorf("decompressed data mismatch:\n  got:  %s\n  want: %s", string(data), string(original))
	}
}

func TestIsGzipped_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		data []byte
		want bool
	}{
		{
			name: "gz_extension",
			path: "foo.json.gz",
			data: []byte("not actually gzip"),
			want: true,
		},
		{
			name: "magic_bytes",
			path: "foo.json",
			data: []byte{0x1f, 0x8b, 0x00},
			want: true,
		},
		{
			name: "plain_json",
			path: "foo.json",
			data: []byte(`{"key":"value"}`),
			want: false,
		},
		{
			name: "empty_data",
			path: "foo.json",
			data: []byte{},
			want: false,
		},
		{
			name: "single_byte",
			path: "foo.json",
			data: []byte{0x1f},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isGzipped(tt.path, tt.data)
			if got != tt.want {
				t.Errorf("isGzipped(%q, %v) = %v, want %v", tt.path, tt.data, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Client buildError with oversized response body
// ---------------------------------------------------------------------------

func TestBuildError_OversizedBody(t *testing.T) {
	t.Parallel()

	client := NewDeployClient("http://example.com", "key")

	// Build a response with a body larger than maxErrorBodySize (1 MB).
	bigBody := strings.Repeat("x", maxErrorBodySize+1024)
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader(bigBody)),
		Header:     make(http.Header),
	}

	ce := client.buildError(resp)
	if ce.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", ce.StatusCode)
	}

	// The body should be truncated to maxErrorBodySize.
	if len(ce.Body) > maxErrorBodySize {
		t.Errorf("Body length = %d, exceeds maxErrorBodySize (%d)", len(ce.Body), maxErrorBodySize)
	}
}

func TestBuildError_JSONError(t *testing.T) {
	t.Parallel()

	client := NewDeployClient("http://example.com", "key")

	body := `{"error": "something went wrong"}`
	resp := &http.Response{
		StatusCode: 422,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	ce := client.buildError(resp)
	if ce.StatusCode != 422 {
		t.Errorf("StatusCode = %d, want 422", ce.StatusCode)
	}
	if ce.Message != "something went wrong" {
		t.Errorf("Message = %q, want %q", ce.Message, "something went wrong")
	}
}

func TestBuildError_JSONMessage(t *testing.T) {
	t.Parallel()

	client := NewDeployClient("http://example.com", "key")

	body := `{"message": "alt error field"}`
	resp := &http.Response{
		StatusCode: 403,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	ce := client.buildError(resp)
	if ce.Message != "alt error field" {
		t.Errorf("Message = %q, want %q", ce.Message, "alt error field")
	}
}

func TestBuildError_NonJSON(t *testing.T) {
	t.Parallel()

	client := NewDeployClient("http://example.com", "key")

	resp := &http.Response{
		StatusCode: 502,
		Body:       io.NopCloser(strings.NewReader("Bad Gateway")),
		Header:     make(http.Header),
	}

	ce := client.buildError(resp)
	if ce.StatusCode != 502 {
		t.Errorf("StatusCode = %d, want 502", ce.StatusCode)
	}
	if ce.Message != "" {
		t.Errorf("Message = %q, want empty for non-JSON body", ce.Message)
	}
	if ce.Body != "Bad Gateway" {
		t.Errorf("Body = %q, want %q", ce.Body, "Bad Gateway")
	}
}

func TestBuildError_EmptyBody(t *testing.T) {
	t.Parallel()

	client := NewDeployClient("http://example.com", "key")

	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}

	ce := client.buildError(resp)
	if ce.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", ce.StatusCode)
	}
	if ce.Message != "" {
		t.Errorf("Message = %q, want empty", ce.Message)
	}
}

// ---------------------------------------------------------------------------
// ClientError.Error() formatting
// ---------------------------------------------------------------------------

func TestClientError_Error_WithMessage(t *testing.T) {
	t.Parallel()
	ce := &ClientError{StatusCode: 409, Message: "conflict"}
	want := "deploy remote: 409 conflict"
	if ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}
}

func TestClientError_Error_WithoutMessage(t *testing.T) {
	t.Parallel()
	ce := &ClientError{StatusCode: 500}
	want := "deploy remote: HTTP 500"
	if ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}
}

// ---------------------------------------------------------------------------
// Snapshot ID collision
// ---------------------------------------------------------------------------

func TestSaveSnapshot_CollisionWithinSameSecond(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	payload := &SyncPayload{
		Manifest: SyncManifest{
			Tables:    []string{},
			RowCounts: map[string]int{},
		},
		Tables:   map[string]TableData{},
		UserRefs: map[string]string{},
	}

	// Call SaveSnapshot rapidly in sequence. Since the ID is based on
	// time.Now().UTC().Format("20060102_150405"), two calls within the
	// same second will produce the same ID and thus the same filename.
	// This test documents the current behavior: the second write
	// overwrites the first file silently (os.WriteFile with O_CREATE).
	id1, err := SaveSnapshot(dir, payload)
	if err != nil {
		t.Fatalf("first SaveSnapshot: %v", err)
	}

	id2, err := SaveSnapshot(dir, payload)
	if err != nil {
		t.Fatalf("second SaveSnapshot: %v", err)
	}

	// Current behavior: both IDs are identical because they share the
	// same second. This IS a known limitation that should be documented.
	if id1 != id2 {
		// If they differ, the implementation already handles sub-second
		// uniqueness, which is even better.
		t.Logf("IDs differ (good): %q vs %q", id1, id2)
		return
	}

	t.Logf("WARNING: snapshot IDs collide within same second: %q == %q", id1, id2)

	// Verify the file exists (one of them overwrote the other).
	path := filepath.Join(dir, id1+".json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected snapshot file at %s, got error: %v", path, err)
	}
}

func TestSaveSnapshot_RoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	payload := &SyncPayload{
		Manifest: SyncManifest{
			SchemaVersion: "test",
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Strategy:      StrategyOverwrite,
			Tables:        []string{"datatypes"},
			RowCounts:     map[string]int{"datatypes": 1},
		},
		Tables: map[string]TableData{
			"datatypes": {
				Columns: []string{"datatype_id", "label"},
				Rows:    [][]any{{"DT01", "Page"}},
			},
		},
		UserRefs: map[string]string{},
	}

	id, err := SaveSnapshot(dir, payload)
	if err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	loaded, err := LoadSnapshot(dir, id)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}

	if len(loaded.Tables["datatypes"].Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(loaded.Tables["datatypes"].Rows))
	}
	if loaded.Manifest.Strategy != StrategyOverwrite {
		t.Errorf("strategy = %q, want %q", loaded.Manifest.Strategy, StrategyOverwrite)
	}
}

// ---------------------------------------------------------------------------
// VerifyImport with crafted table names (SQL injection prevention)
// ---------------------------------------------------------------------------

func TestVerifyImport_CraftedTableName(t *testing.T) {
	t.Parallel()

	// A crafted table name should be rejected by IsValidTable. The
	// db.DBTable type constrains names to known constants, but we verify
	// that VerifyImport checks with IsValidTable before constructing SQL.
	malicious := db.DBTable("users; DROP TABLE users; --")

	// We use a real in-memory SQLite for the executor so sql.Rows is real.
	pool := openTestDB(t)
	fakeOps := &stubDeployOps{}

	expected := map[db.DBTable]int{
		malicious: 5,
	}

	errs := VerifyImport(context.Background(), fakeOps, pool, expected)

	// Should have an error for the unknown table name.
	found := false
	for _, e := range errs {
		if e.Table == string(malicious) && e.Phase == "verify" && strings.Contains(e.Message, "unknown table") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'unknown table' error for crafted name %q, got: %v", malicious, errs)
	}
}

func TestVerifyImport_ValidTable(t *testing.T) {
	t.Parallel()

	// Create a real in-memory SQLite and a table with a known row count.
	pool := openTestDB(t)
	if _, err := pool.ExecContext(context.Background(), "CREATE TABLE datatypes (datatype_id TEXT PRIMARY KEY, label TEXT);"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	for i := range 3 {
		if _, err := pool.ExecContext(context.Background(), "INSERT INTO datatypes (datatype_id, label) VALUES (?, ?);", fmt.Sprintf("DT%02d", i), fmt.Sprintf("Label%d", i)); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	fakeOps := &stubDeployOps{}
	expected := map[db.DBTable]int{
		db.Datatype: 3,
	}

	errs := VerifyImport(context.Background(), fakeOps, pool, expected)

	for _, e := range errs {
		if e.Table == string(db.Datatype) {
			t.Errorf("unexpected error for valid table: %v", e)
		}
	}
}

func TestVerifyImport_CountMismatch(t *testing.T) {
	t.Parallel()

	pool := openTestDB(t)
	if _, err := pool.ExecContext(context.Background(), "CREATE TABLE datatypes (datatype_id TEXT PRIMARY KEY, label TEXT);"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	// Insert 5 rows but expect 10.
	for i := range 5 {
		if _, err := pool.ExecContext(context.Background(), "INSERT INTO datatypes (datatype_id, label) VALUES (?, ?);", fmt.Sprintf("DT%02d", i), fmt.Sprintf("Label%d", i)); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	fakeOps := &stubDeployOps{}
	expected := map[db.DBTable]int{
		db.Datatype: 10,
	}

	errs := VerifyImport(context.Background(), fakeOps, pool, expected)

	found := false
	for _, e := range errs {
		if e.Table == string(db.Datatype) && strings.Contains(e.Message, "row count mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected row count mismatch error")
	}
}

// ---------------------------------------------------------------------------
// ExportPayload with canceled context
// ---------------------------------------------------------------------------

func TestExportPayload_CanceledContext(t *testing.T) {
	t.Parallel()

	// ExportPayload currently takes a context but does not check it
	// directly (the context parameter has underscore). This test documents
	// that behavior and ensures it does not panic with a canceled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// We need a minimal mock driver that returns empty data.
	driver := &fakeDbDriver{}

	// ExportPayload should still succeed because it does not use the
	// context for anything besides passing through.
	payload, err := ExportPayload(ctx, driver, ExportOptions{Tables: []db.DBTable{db.Datatype}})
	if err != nil {
		t.Fatalf("ExportPayload with canceled context: %v", err)
	}
	if payload == nil {
		t.Fatal("payload is nil")
	}
}

// ---------------------------------------------------------------------------
// Client do() with server returning various statuses
// ---------------------------------------------------------------------------

func TestClientDo_GzipResponse(t *testing.T) {
	t.Parallel()

	// Test that the client transparently decompresses gzip responses.
	wantResp := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
		NodeID:  "test-node",
	}
	respJSON, err := json.Marshal(wantResp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(respJSON); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}))
	defer srv.Close()

	client := NewDeployClient(srv.URL, "test-key")
	got, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("Status = %q, want %q", got.Status, "ok")
	}
	if got.NodeID != "test-node" {
		t.Errorf("NodeID = %q, want %q", got.NodeID, "test-node")
	}
}

func TestClientDo_AuthHeader(t *testing.T) {
	t.Parallel()

	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer srv.Close()

	client := NewDeployClient(srv.URL, "my-secret-key")
	if _, err := client.Health(context.Background()); err != nil {
		t.Fatalf("Health: %v", err)
	}

	if receivedAuth != "Bearer my-secret-key" {
		t.Errorf("Authorization = %q, want %q", receivedAuth, "Bearer my-secret-key")
	}
}

func TestClientDo_Non2xxReturnsClientError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
	}))
	defer srv.Close()

	client := NewDeployClient(srv.URL, "key")
	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for 403 response")
	}

	var ce *ClientError
	if errors.As(err, &ce) {
		if ce.StatusCode != 403 {
			t.Errorf("StatusCode = %d, want 403", ce.StatusCode)
		}
		if ce.Message != "forbidden" {
			t.Errorf("Message = %q, want %q", ce.Message, "forbidden")
		}
	} else {
		// buildError returns *ClientError which satisfies error interface
		if !strings.Contains(err.Error(), "403") {
			t.Errorf("error should mention 403, got: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

func TestFindIDColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		columns []string
		want    []int
	}{
		{
			name:    "typical",
			columns: []string{"content_data_id", "parent_id", "label", "author_id"},
			want:    []int{0, 1, 3},
		},
		{
			name:    "no_id_columns",
			columns: []string{"label", "data", "value"},
			want:    nil,
		},
		{
			name:    "empty",
			columns: []string{},
			want:    nil,
		},
		{
			name:    "suffix_only",
			columns: []string{"_id"},
			want:    []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := findIDColumns(tt.columns)
			if len(got) != len(tt.want) {
				t.Fatalf("findIDColumns(%v) = %v, want %v", tt.columns, got, tt.want)
			}
			for i, idx := range got {
				if idx != tt.want[i] {
					t.Errorf("index %d: got %d, want %d", i, idx, tt.want[i])
				}
			}
		})
	}
}

func TestIsValidULID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "valid_uppercase", s: "01ARZ3NDEKTSV4RRFFQ69G5FAV", want: true},
		{name: "valid_lowercase", s: "01arz3ndektsv4rrffq69g5fav", want: true},
		{name: "too_short", s: "01ARZ3NDEKTSV4RRFFQ69G5FA", want: false},
		{name: "too_long", s: "01ARZ3NDEKTSV4RRFFQ69G5FAVX", want: false},
		{name: "empty", s: "", want: false},
		// I, L, O, U are excluded from Crockford base32
		{name: "contains_I", s: "01ARZ3NDEKTSV4RRFFQ69G5FAI", want: false},
		{name: "contains_L", s: "01ARZ3NDEKTSV4RRFFQ69G5FAL", want: false},
		{name: "contains_O", s: "01ARZ3NDEKTSV4RRFFQ69G5FAO", want: false},
		{name: "contains_U", s: "01ARZ3NDEKTSV4RRFFQ69G5FAU", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isValidULID(tt.s)
			if got != tt.want {
				t.Errorf("isValidULID(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestFindCharIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s    string
		c    byte
		want int
	}{
		{"hello,world", ',', 5},
		{"nope", ',', -1},
		{"", ',', -1},
		{",leading", ',', 0},
		{"trailing,", ',', 8},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q_%c", tt.s, tt.c), func(t *testing.T) {
			t.Parallel()
			got := findCharIndex(tt.s, tt.c)
			if got != tt.want {
				t.Errorf("findCharIndex(%q, %c) = %d, want %d", tt.s, tt.c, got, tt.want)
			}
		})
	}
}

func TestExtractRowID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		row     []any
		columns []string
		want    string
	}{
		{
			name:    "string_first_column",
			row:     []any{"ULID123", "data"},
			columns: []string{"id", "data"},
			want:    "ULID123",
		},
		{
			name:    "numeric_first_column",
			row:     []any{42, "data"},
			columns: []string{"id", "data"},
			want:    "42",
		},
		{
			name:    "empty_row",
			row:     []any{},
			columns: []string{},
			want:    "",
		},
		{
			name:    "nil_first_value",
			row:     []any{nil},
			columns: []string{"id"},
			want:    "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractRowID(tt.row, tt.columns)
			if got != tt.want {
				t.Errorf("extractRowID = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColIndex(t *testing.T) {
	t.Parallel()

	cols := []string{"id", "name", "author_id", "date_created"}

	tests := []struct {
		name string
		want int
	}{
		{"id", 0},
		{"name", 1},
		{"author_id", 2},
		{"date_created", 3},
		{"nonexistent", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := colIndex(cols, tt.name)
			if got != tt.want {
				t.Errorf("colIndex(cols, %q) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// resolveContext behavior
// ---------------------------------------------------------------------------

func TestResolveContext_ParentHasDeadline(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer parentCancel()

	ctx, cancel := resolveContext(parent)
	defer cancel()

	// resolveContext should return the parent as-is when it already has a deadline.
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}
	if time.Until(deadline) > 31*time.Second {
		t.Errorf("deadline too far in the future: %v", deadline)
	}
}

func TestResolveContext_NoDeadline(t *testing.T) {
	t.Parallel()

	ctx, cancel := resolveContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected resolveContext to add a deadline")
	}
	// Should be approximately DefaultTimeout (5 min) from now.
	remaining := time.Until(deadline)
	if remaining < 4*time.Minute || remaining > 6*time.Minute {
		t.Errorf("expected deadline ~5m from now, got %v", remaining)
	}
}

// ---------------------------------------------------------------------------
// Concurrent ImportPayload lock test (verifies the loser gets clean error)
// ---------------------------------------------------------------------------

func TestImportMu_ConcurrentLock(t *testing.T) {
	t.Parallel()

	// We test the sync.Mutex TryLock behavior directly to verify our
	// understanding of the concurrency contract used in ImportPayload.
	// The real importMu is a package-level var; we test the pattern
	// with a local mutex to avoid interfering with other tests.
	var testMu sync.Mutex
	testMu.Lock()

	// Channel to synchronize: second goroutine attempts lock after first holds it.
	got := make(chan bool, 1)
	go func() {
		got <- testMu.TryLock()
	}()

	// Give the goroutine time to attempt.
	time.Sleep(50 * time.Millisecond)

	locked := <-got
	if locked {
		t.Error("expected TryLock to fail when mutex is already held")
		testMu.Unlock()
	}

	testMu.Unlock()

	// Now it should succeed.
	if !testMu.TryLock() {
		t.Error("expected TryLock to succeed after unlock")
	} else {
		testMu.Unlock()
	}
}

// ---------------------------------------------------------------------------
// Validate helpers
// ---------------------------------------------------------------------------

func TestValidatePayload_HashMismatch(t *testing.T) {
	t.Parallel()

	payload := &SyncPayload{
		Manifest: SyncManifest{
			PayloadHash: "wrong-hash",
			RowCounts:   map[string]int{"datatypes": 1},
		},
		Tables: map[string]TableData{
			"datatypes": {
				Columns: []string{"datatype_id", "label"},
				Rows:    [][]any{{"01ARZ3NDEKTSV4RRFFQ69G5FAV", "Page"}},
			},
		},
		UserRefs: map[string]string{},
	}
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	errs := ValidatePayload(payload, nil)

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "payload hash mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected payload hash mismatch error")
	}
}

func TestValidatePayload_RowCountMismatch(t *testing.T) {
	t.Parallel()

	payload := &SyncPayload{
		Manifest: SyncManifest{
			RowCounts: map[string]int{"datatypes": 99},
		},
		Tables: map[string]TableData{
			"datatypes": {
				Columns: []string{"datatype_id"},
				Rows:    [][]any{{"01ARZ3NDEKTSV4RRFFQ69G5FAV"}},
			},
		},
		UserRefs: map[string]string{},
	}
	hash, _ := computePayloadHash(payload.Tables)
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	errs := ValidatePayload(payload, nil)

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "row count mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected row count mismatch error")
	}
}

func TestValidatePayload_ValidPayload(t *testing.T) {
	t.Parallel()

	payload := &SyncPayload{
		Manifest: SyncManifest{
			RowCounts: map[string]int{"datatypes": 1},
		},
		Tables: map[string]TableData{
			"datatypes": {
				Columns: []string{"datatype_id", "label"},
				Rows:    [][]any{{"01ARZ3NDEKTSV4RRFFQ69G5FAV", "Page"}},
			},
		},
		UserRefs: map[string]string{},
	}
	hash, _ := computePayloadHash(payload.Tables)
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	errs := ValidatePayload(payload, nil)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for valid payload, got %d: %v", len(errs), errs)
	}
}

func TestValidatePayload_InvalidULID(t *testing.T) {
	t.Parallel()

	payload := &SyncPayload{
		Manifest: SyncManifest{
			RowCounts: map[string]int{"datatypes": 1},
		},
		Tables: map[string]TableData{
			"datatypes": {
				Columns: []string{"datatype_id", "label"},
				Rows:    [][]any{{"INVALID_ULID!!", "Page"}},
			},
		},
		UserRefs: map[string]string{},
	}
	hash, _ := computePayloadHash(payload.Tables)
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	errs := ValidatePayload(payload, nil)

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid ULID") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected invalid ULID error, got: %v", errs)
	}
}

func TestValidateUserRefs_MissingAuthor(t *testing.T) {
	t.Parallel()

	payload := &SyncPayload{
		Tables: map[string]TableData{
			"content_data": {
				Columns: []string{"content_data_id", "author_id"},
				Rows: [][]any{
					{"01ARZ3NDEKTSV4RRFFQ69G5FAV", "MISSING_USER_ULID_XXXXXXXXX"},
				},
			},
		},
		UserRefs: map[string]string{}, // empty: no refs
	}

	errs := validateUserRefs(payload)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "author_id") && strings.Contains(e.Message, "not found in user_refs") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected missing author_id in user_refs error")
	}
}

// ---------------------------------------------------------------------------
// structSliceToTableData and serializeField
// ---------------------------------------------------------------------------

// testStruct is a minimal struct for structSliceToTableData tests.
type testStruct struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestStructSliceToTableData_NilInput(t *testing.T) {
	t.Parallel()

	td, err := structSliceToTableData(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Columns) != 0 {
		t.Errorf("expected 0 columns, got %d", len(td.Columns))
	}
	if len(td.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(td.Rows))
	}
}

func TestStructSliceToTableData_NilPointer(t *testing.T) {
	t.Parallel()

	var s *[]testStruct
	td, err := structSliceToTableData(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Columns) != 0 {
		t.Errorf("expected 0 columns, got %d", len(td.Columns))
	}
}

func TestStructSliceToTableData_NonPointer(t *testing.T) {
	t.Parallel()

	s := []testStruct{{ID: "a"}}
	_, err := structSliceToTableData(s)
	if err == nil {
		t.Fatal("expected error for non-pointer input")
	}
	if !strings.Contains(err.Error(), "expected pointer") {
		t.Errorf("error = %q, want 'expected pointer'", err.Error())
	}
}

func TestStructSliceToTableData_EmptySlice(t *testing.T) {
	t.Parallel()

	s := &[]testStruct{}
	td, err := structSliceToTableData(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(td.Rows))
	}
	// Columns should still be populated from the struct type.
	if len(td.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(td.Columns))
	}
}

func TestStructSliceToTableData_WithData(t *testing.T) {
	t.Parallel()

	s := &[]testStruct{
		{ID: "abc", Name: "first"},
		{ID: "def", Name: "second"},
	}
	td, err := structSliceToTableData(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(td.Columns))
	}
	if td.Columns[0] != "id" || td.Columns[1] != "name" {
		t.Errorf("columns = %v, want [id name]", td.Columns)
	}
	if len(td.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(td.Rows))
	}
	if td.Rows[0][0] != "abc" || td.Rows[0][1] != "first" {
		t.Errorf("row 0 = %v, want [abc first]", td.Rows[0])
	}
}

// ---------------------------------------------------------------------------
// computeSchemaVersion determinism
// ---------------------------------------------------------------------------

func TestComputeSchemaVersion_Deterministic(t *testing.T) {
	t.Parallel()

	tables := map[string]TableData{
		"b_table": {Columns: []string{"x", "y"}},
		"a_table": {Columns: []string{"p", "q", "r"}},
	}

	v1 := computeSchemaVersion(tables)
	v2 := computeSchemaVersion(tables)
	if v1 != v2 {
		t.Errorf("non-deterministic: %q != %q", v1, v2)
	}
	if len(v1) != 64 {
		t.Errorf("expected SHA256 hex (64 chars), got %d chars", len(v1))
	}
}

func TestComputeSchemaVersion_DifferentColumnsProduceDifferentHash(t *testing.T) {
	t.Parallel()

	tables1 := map[string]TableData{
		"table": {Columns: []string{"a", "b"}},
	}
	tables2 := map[string]TableData{
		"table": {Columns: []string{"a", "c"}},
	}

	v1 := computeSchemaVersion(tables1)
	v2 := computeSchemaVersion(tables2)
	if v1 == v2 {
		t.Error("different columns should produce different schema versions")
	}
}

// ---------------------------------------------------------------------------
// NewDeployClient URL normalization
// ---------------------------------------------------------------------------

func TestNewDeployClient_TrimsTrailingSlash(t *testing.T) {
	t.Parallel()

	client := NewDeployClient("https://example.com/", "key")
	if client.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want trailing slash removed", client.baseURL)
	}
}

func TestNewDeployClient_NoSlash(t *testing.T) {
	t.Parallel()

	client := NewDeployClient("https://example.com", "key")
	if client.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q", client.baseURL)
	}
}

// ---------------------------------------------------------------------------
// SnapshotDir
// ---------------------------------------------------------------------------

func TestSnapshotDir_Custom(t *testing.T) {
	t.Parallel()

	cfg := configWithSnapshotDir("/custom/path")
	got := SnapshotDir(cfg)
	if got != "/custom/path" {
		t.Errorf("SnapshotDir = %q, want /custom/path", got)
	}
}

func TestSnapshotDir_Default(t *testing.T) {
	t.Parallel()

	cfg := configWithSnapshotDir("")
	got := SnapshotDir(cfg)
	if got != "./deploy/snapshots" {
		t.Errorf("SnapshotDir = %q, want ./deploy/snapshots", got)
	}
}

// ---------------------------------------------------------------------------
// ListSnapshots
// ---------------------------------------------------------------------------

func TestListSnapshots_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	snapshots, err := ListSnapshots(dir)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snapshots))
	}
}

func TestListSnapshots_NonexistentDir(t *testing.T) {
	t.Parallel()

	snapshots, err := ListSnapshots("/nonexistent/path/for/test")
	if err != nil {
		t.Fatalf("ListSnapshots should return nil for nonexistent dir, got error: %v", err)
	}
	if snapshots != nil {
		t.Errorf("expected nil snapshots for nonexistent dir, got %d", len(snapshots))
	}
}

func TestListSnapshots_SortedNewestFirst(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	payload := &SyncPayload{
		Manifest: SyncManifest{
			Tables:    []string{"test"},
			RowCounts: map[string]int{"test": 0},
		},
		Tables:   map[string]TableData{"test": {Columns: []string{"id"}, Rows: [][]any{}}},
		UserRefs: map[string]string{},
	}

	data, _ := json.Marshal(payload)
	// Older snapshot
	if err := os.WriteFile(filepath.Join(dir, "snap_20250101_120000.json"), data, 0640); err != nil {
		t.Fatalf("write old snapshot: %v", err)
	}
	// Newer snapshot
	if err := os.WriteFile(filepath.Join(dir, "snap_20260223_120000.json"), data, 0640); err != nil {
		t.Fatalf("write new snapshot: %v", err)
	}

	snapshots, err := ListSnapshots(dir)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}

	// Newest first
	if snapshots[0].ID != "snap_20260223_120000" {
		t.Errorf("first snapshot ID = %q, want snap_20260223_120000", snapshots[0].ID)
	}
	if snapshots[1].ID != "snap_20250101_120000" {
		t.Errorf("second snapshot ID = %q, want snap_20250101_120000", snapshots[1].ID)
	}
}

// ---------------------------------------------------------------------------
// parseSnapshotTimestamp
// ---------------------------------------------------------------------------

func TestParseSnapshotTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want time.Time
	}{
		{
			id:   "snap_20260223_143022",
			want: time.Date(2026, 2, 23, 14, 30, 22, 0, time.UTC),
		},
		{
			id:   "snap_invalid",
			want: time.Time{},
		},
		{
			id:   "snap_",
			want: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			got := parseSnapshotTimestamp(tt.id)
			if !got.Equal(tt.want) {
				t.Errorf("parseSnapshotTimestamp(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BuildDryRunResult
// ---------------------------------------------------------------------------

func TestBuildDryRunResult_ValidPayload(t *testing.T) {
	t.Parallel()

	payload := &SyncPayload{
		Manifest: SyncManifest{
			RowCounts: map[string]int{"datatypes": 2},
		},
		Tables: map[string]TableData{
			"datatypes": {
				Columns: []string{"datatype_id", "label"},
				Rows: [][]any{
					{"01ARZ3NDEKTSV4RRFFQ69G5FAV", "Page"},
					{"01ARZ3NDEKTSV4RRFFQ69G5FAW", "Post"},
				},
			},
		},
		UserRefs: map[string]string{},
	}
	hash, _ := computePayloadHash(payload.Tables)
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	result := BuildDryRunResult(payload, nil)
	if !result.Success {
		t.Errorf("expected success, got errors: %v", result.Errors)
	}
	if !result.DryRun {
		t.Error("expected DryRun = true")
	}
	if result.RowCounts["datatypes"] != 2 {
		t.Errorf("RowCounts[datatypes] = %d, want 2", result.RowCounts["datatypes"])
	}
}

// ---------------------------------------------------------------------------
// clientForEnv
// ---------------------------------------------------------------------------

func TestClientForEnv_Found(t *testing.T) {
	t.Parallel()

	cfg := configWithEnvs([]envDef{
		{name: "staging", url: "https://staging.example.com", key: "sk-123"},
	})

	client, err := clientForEnv(cfg, "staging")
	if err != nil {
		t.Fatalf("clientForEnv: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != "https://staging.example.com" {
		t.Errorf("baseURL = %q", client.baseURL)
	}
}

func TestClientForEnv_NotFound(t *testing.T) {
	t.Parallel()

	cfg := configWithEnvs(nil)
	_, err := clientForEnv(cfg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown environment")
	}
	if !strings.Contains(err.Error(), "unknown deploy environment") {
		t.Errorf("error = %q, want 'unknown deploy environment'", err.Error())
	}
}

func TestClientForEnv_NoURL(t *testing.T) {
	t.Parallel()

	cfg := configWithEnvs([]envDef{
		{name: "broken", url: "", key: "sk-123"},
	})

	_, err := clientForEnv(cfg, "broken")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "no URL configured") {
		t.Errorf("error = %q, want 'no URL configured'", err.Error())
	}
}

func TestClientForEnv_NoAPIKey(t *testing.T) {
	t.Parallel()

	cfg := configWithEnvs([]envDef{
		{name: "nokey", url: "https://example.com", key: ""},
	})

	_, err := clientForEnv(cfg, "nokey")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	if !strings.Contains(err.Error(), "no api_key configured") {
		t.Errorf("error = %q, want 'no api_key configured'", err.Error())
	}
}

// ---------------------------------------------------------------------------
// columnsFromType
// ---------------------------------------------------------------------------

func TestColumnsFromType_OmitsIgnoredFields(t *testing.T) {
	t.Parallel()

	type sample struct {
		Exported   string `json:"exported"`
		Ignored    string `json:"-"`
		Untagged   string
		WithOption string `json:"with_opt,omitempty"`
	}

	cols := columnsFromType(reflect.TypeOf(sample{}))
	expected := []string{"exported", "with_opt"}
	if len(cols) != len(expected) {
		t.Fatalf("columns = %v, want %v", cols, expected)
	}
	for i, c := range cols {
		if c != expected[i] {
			t.Errorf("column %d = %q, want %q", i, c, expected[i])
		}
	}
}

// ---------------------------------------------------------------------------
// resolveUserRefs remap -- SQLite backend verification
// ---------------------------------------------------------------------------

func TestResolveUserRefs_SQLite_RemapToAdmin(t *testing.T) {
	t.Parallel()

	pool := openTestDB(t)
	ctx := context.Background()

	// Create minimal schema.
	for _, ddl := range []string{
		`CREATE TABLE roles (role_id TEXT PRIMARY KEY, label TEXT);`,
		`CREATE TABLE users (user_id TEXT PRIMARY KEY, username TEXT, name TEXT, email TEXT, hash TEXT, role TEXT, date_created TEXT, date_modified TEXT);`,
		`CREATE TABLE content_data (content_data_id TEXT PRIMARY KEY, author_id TEXT, published_by TEXT);`,
		`CREATE TABLE content_fields (content_field_id TEXT PRIMARY KEY, author_id TEXT);`,
		`CREATE TABLE admin_content_data (content_data_id TEXT PRIMARY KEY, author_id TEXT, published_by TEXT);`,
		`CREATE TABLE admin_content_fields (content_field_id TEXT PRIMARY KEY, author_id TEXT);`,
	} {
		if _, err := pool.ExecContext(ctx, ddl); err != nil {
			t.Fatalf("ddl: %v", err)
		}
	}

	// Insert admin role and admin user.
	if _, err := pool.ExecContext(ctx,
		"INSERT INTO roles (role_id, label) VALUES (?, ?);",
		"ADMIN_ROLE_ID_XXXXXXXXXXXXXX", "admin",
	); err != nil {
		t.Fatalf("insert role: %v", err)
	}
	if _, err := pool.ExecContext(ctx,
		"INSERT INTO users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
		"ADMIN_USER_ID_XXXXXXXXXXXXXX", "admin", "Admin", "admin@example.com", "hash", "ADMIN_ROLE_ID_XXXXXXXXXXXXXX", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z",
	); err != nil {
		t.Fatalf("insert admin: %v", err)
	}

	// Insert content referencing a missing user.
	missingUserID := "MISSING_USER_IDXXXXXXXXXXXXXX"
	if _, err := pool.ExecContext(ctx,
		"INSERT INTO content_data (content_data_id, author_id, published_by) VALUES (?, ?, ?);",
		"CONTENT_ID_XXXXXXXXXXXXXXXX", missingUserID, missingUserID,
	); err != nil {
		t.Fatalf("insert content: %v", err)
	}

	ops := &sqliteTestDeployOps{pool: pool}
	userRefs := map[string]string{
		missingUserID: "ghost",
	}

	tx, err := pool.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	remapped, err := resolveUserRefs(ctx, tx, ops, userRefs)
	if err != nil {
		t.Fatalf("resolveUserRefs: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify remap was reported.
	if len(remapped) != 1 {
		t.Fatalf("expected 1 remapped user, got %d", len(remapped))
	}
	if remapped[missingUserID] != "ADMIN_USER_ID_XXXXXXXXXXXXXX" {
		t.Errorf("remapped target = %q, want admin user ID", remapped[missingUserID])
	}

	// Verify content_data was updated.
	var authorID, publishedBy string
	if err := pool.QueryRowContext(ctx,
		"SELECT author_id, published_by FROM content_data WHERE content_data_id = ?;",
		"CONTENT_ID_XXXXXXXXXXXXXXXX",
	).Scan(&authorID, &publishedBy); err != nil {
		t.Fatalf("query content: %v", err)
	}
	if authorID != "ADMIN_USER_ID_XXXXXXXXXXXXXX" {
		t.Errorf("author_id = %q, want admin user ID", authorID)
	}
	if publishedBy != "ADMIN_USER_ID_XXXXXXXXXXXXXX" {
		t.Errorf("published_by = %q, want admin user ID", publishedBy)
	}

	// Verify no placeholder user was created.
	var userCount int64
	if err := pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM users;").Scan(&userCount); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 1 {
		t.Errorf("expected 1 user (admin only), got %d", userCount)
	}
}

// ---------------------------------------------------------------------------
// Fakes and helpers
// ---------------------------------------------------------------------------

// openTestDB creates an in-memory SQLite database for testing.
// The database is closed when the test completes.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	pool, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

// stubDeployOps satisfies db.DeployOps for tests that only check validation paths.
type stubDeployOps struct{}

func (f *stubDeployOps) ImportAtomic(_ context.Context, fn db.ImportFunc) error {
	return nil
}

func (f *stubDeployOps) TruncateTable(_ context.Context, _ db.Executor, _ db.DBTable) error {
	return nil
}

func (f *stubDeployOps) BulkInsert(_ context.Context, _ db.Executor, _ db.DBTable, _ []string, _ [][]any) error {
	return nil
}

func (f *stubDeployOps) VerifyForeignKeys(_ context.Context, _ db.Executor) ([]db.FKViolation, error) {
	return nil, nil
}

func (f *stubDeployOps) Placeholder(n int) string { return fmt.Sprintf("$%d", n) }

func (f *stubDeployOps) IntrospectColumns(_ context.Context, _ db.DBTable) ([]db.ColumnMeta, error) {
	return nil, fmt.Errorf("table does not exist")
}

func (f *stubDeployOps) QueryAllRows(_ context.Context, _ db.DBTable) ([]string, [][]any, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// sqliteTestDeployOps wraps a *sql.DB for SQLite-compatible deploy ops in tests.
type sqliteTestDeployOps struct {
	pool *sql.DB
}

func (s *sqliteTestDeployOps) ImportAtomic(ctx context.Context, fn db.ImportFunc) error {
	tx, err := s.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *sqliteTestDeployOps) TruncateTable(ctx context.Context, ex db.Executor, table db.DBTable) error {
	_, err := ex.ExecContext(ctx, "DELETE FROM "+string(table)+";")
	return err
}

func (s *sqliteTestDeployOps) BulkInsert(ctx context.Context, ex db.Executor, table db.DBTable, columns []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}
	colList := strings.Join(columns, ", ")
	for _, row := range rows {
		placeholders := strings.Repeat("?, ", len(columns)-1) + "?"
		query := "INSERT INTO " + string(table) + " (" + colList + ") VALUES (" + placeholders + ");"
		if _, err := ex.ExecContext(ctx, query, row...); err != nil {
			return err
		}
	}
	return nil
}

func (s *sqliteTestDeployOps) VerifyForeignKeys(_ context.Context, _ db.Executor) ([]db.FKViolation, error) {
	return nil, nil
}

func (s *sqliteTestDeployOps) Placeholder(_ int) string { return "?" }

func (s *sqliteTestDeployOps) IntrospectColumns(ctx context.Context, table db.DBTable) ([]db.ColumnMeta, error) {
	rows, err := s.pool.QueryContext(ctx, "PRAGMA table_info("+string(table)+");")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []db.ColumnMeta
	for rows.Next() {
		var cid int
		var name, colType string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		upper := strings.ToUpper(colType)
		isInt := strings.Contains(upper, "INT") || upper == "SERIAL"
		cols = append(cols, db.ColumnMeta{Name: name, IsInteger: isInt})
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q does not exist", string(table))
	}
	return cols, rows.Err()
}

func (s *sqliteTestDeployOps) QueryAllRows(_ context.Context, _ db.DBTable) ([]string, [][]any, error) {
	return nil, nil, fmt.Errorf("not implemented in test stub")
}

// fakeDbDriver is a minimal DbDriver mock for ExportPayload tests.
// It only implements the methods actually called during export of Datatype table.
type fakeDbDriver struct {
	db.DbDriver
}

func (f *fakeDbDriver) ListDatatypes() (*[]db.Datatypes, error) {
	result := []db.Datatypes{}
	return &result, nil
}

func (f *fakeDbDriver) ListUsers() (*[]db.Users, error) {
	result := []db.Users{}
	return &result, nil
}

// configWithSnapshotDir builds a minimal config.Config with the given snapshot dir.
func configWithSnapshotDir(dir string) config.Config {
	return config.Config{
		Deploy_Snapshot_Dir: dir,
	}
}

type envDef struct {
	name string
	url  string
	key  string
}

func configWithEnvs(envs []envDef) config.Config {
	var result []config.DeployEnvironmentConfig
	for _, e := range envs {
		result = append(result, config.DeployEnvironmentConfig{
			Name:   e.name,
			URL:    e.url,
			APIKey: e.key,
		})
	}
	return config.Config{
		Deploy_Environments: result,
	}
}
