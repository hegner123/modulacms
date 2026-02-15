// White-box tests for datatype_field.go: wrapper structs, mapper methods
// across all three database drivers (SQLite, MySQL, PostgreSQL), string mapping
// for TUI display, and audited command struct accessors.
//
// White-box access is needed because:
//   - Audited command structs have unexported fields (ctx, auditCtx, params, conn,
//     recorder) that can only be constructed through the Database/MysqlDatabase/
//     PsqlDatabase factory methods, which require access to the package internals.
//   - We verify that the SQLiteRecorder, MysqlRecorder, and PsqlRecorder package-level
//     vars are correctly wired into command constructors.
package db

import (
	"context"
	"fmt"
	"math"
	"testing"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// dfTestFixture returns a fully populated DatatypeFields and its individual parts.
func dfTestFixture() (DatatypeFields, types.DatatypeID, types.FieldID) {
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()
	df := DatatypeFields{
		ID:         "test-df-id-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  5,
	}
	return df, datatypeID, fieldID
}

// --- MapStringDatatypeField tests ---

func TestMapStringDatatypeField_AllFields(t *testing.T) {
	t.Parallel()
	df, datatypeID, fieldID := dfTestFixture()

	got := MapStringDatatypeField(df)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ID", got.ID, df.ID},
		{"DatatypeID", got.DatatypeID, string(datatypeID)},
		{"FieldID", got.FieldID, string(fieldID)},
		{"SortOrder", got.SortOrder, "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestMapStringDatatypeField_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringDatatypeField(DatatypeFields{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	// Zero-value DatatypeID and FieldID are empty strings,
	// so string() returns ""
	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %q, want %q", got.DatatypeID, "")
	}
	if got.FieldID != "" {
		t.Errorf("FieldID = %q, want %q", got.FieldID, "")
	}
	if got.SortOrder != "0" {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, "0")
	}
}

func TestMapStringDatatypeField_NegativeSortOrder(t *testing.T) {
	t.Parallel()
	df := DatatypeFields{SortOrder: -1}
	got := MapStringDatatypeField(df)
	if got.SortOrder != "-1" {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, "-1")
	}
}

func TestMapStringDatatypeField_LargeSortOrder(t *testing.T) {
	t.Parallel()
	df := DatatypeFields{SortOrder: math.MaxInt64}
	got := MapStringDatatypeField(df)
	want := fmt.Sprintf("%d", int64(math.MaxInt64))
	if got.SortOrder != want {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, want)
	}
}

// --- SQLite Database.MapDatatypeField tests ---

func TestDatabase_MapDatatypeField_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	input := mdb.DatatypesFields{
		ID:         "sqlite-df-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  42,
	}

	got := d.MapDatatypeField(input)

	if got.ID != "sqlite-df-001" {
		t.Errorf("ID = %v, want %v", got.ID, "sqlite-df-001")
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.SortOrder != 42 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 42)
	}
}

func TestDatabase_MapDatatypeField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapDatatypeField(mdb.DatatypesFields{})

	if got.ID != "" {
		t.Errorf("ID = %v, want zero value", got.ID)
	}
	if !got.DatatypeID.IsZero() {
		t.Errorf("DatatypeID.IsZero() = false, want true")
	}
	if !got.FieldID.IsZero() {
		t.Errorf("FieldID.IsZero() = false, want true")
	}
	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
}

// --- SQLite Database.MapCreateDatatypeFieldParams tests ---

func TestDatabase_MapCreateDatatypeFieldParams_WithExplicitID(t *testing.T) {
	t.Parallel()
	d := Database{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	input := CreateDatatypeFieldParams{
		ID:         "explicit-id-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  10,
	}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.ID != "explicit-id-001" {
		t.Errorf("ID = %q, want %q", got.ID, "explicit-id-001")
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.SortOrder != 10 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 10)
	}
}

func TestDatabase_MapCreateDatatypeFieldParams_GeneratesIDWhenEmpty(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateDatatypeFieldParams{
		ID:        "",
		SortOrder: 3,
	}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated when input ID is empty")
	}
	if got.SortOrder != 3 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 3)
	}
}

func TestDatabase_MapCreateDatatypeFieldParams_UniqueGeneratedIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateDatatypeFieldParams{ID: ""}

	got1 := d.MapCreateDatatypeFieldParams(input)
	got2 := d.MapCreateDatatypeFieldParams(input)

	if got1.ID == got2.ID {
		t.Error("two calls with empty ID produced identical generated IDs")
	}
}

func TestDatabase_MapCreateDatatypeFieldParams_PreservesExplicitID(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateDatatypeFieldParams{ID: "keep-this-id"}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.ID != "keep-this-id" {
		t.Errorf("ID = %q, want %q -- explicit ID should be preserved", got.ID, "keep-this-id")
	}
}

// --- SQLite Database.MapUpdateDatatypeFieldParams tests ---

func TestDatabase_MapUpdateDatatypeFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	input := UpdateDatatypeFieldParams{
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  99,
		ID:         "update-id-001",
	}

	got := d.MapUpdateDatatypeFieldParams(input)

	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.SortOrder != 99 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 99)
	}
	if got.ID != "update-id-001" {
		t.Errorf("ID = %q, want %q", got.ID, "update-id-001")
	}
}

func TestDatabase_MapUpdateDatatypeFieldParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateDatatypeFieldParams(UpdateDatatypeFieldParams{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if !got.DatatypeID.IsZero() {
		t.Errorf("DatatypeID.IsZero() = false, want true")
	}
	if !got.FieldID.IsZero() {
		t.Errorf("FieldID.IsZero() = false, want true")
	}
}

// --- MySQL MysqlDatabase.MapDatatypeField tests ---

func TestMysqlDatabase_MapDatatypeField_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	tests := []struct {
		name            string
		inputSortOrder  int32
		wantSortOrder64 int64
	}{
		{"positive sort order", 5, 5},
		{"zero sort order", 0, 0},
		{"negative sort order", -1, -1},
		{"max int32", math.MaxInt32, int64(math.MaxInt32)},
		{"min int32", math.MinInt32, int64(math.MinInt32)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbm.DatatypesFields{
				ID:         "mysql-df-001",
				DatatypeID: datatypeID,
				FieldID:    fieldID,
				SortOrder:  tt.inputSortOrder,
			}

			got := d.MapDatatypeField(input)

			if got.SortOrder != tt.wantSortOrder64 {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.wantSortOrder64)
			}
			if got.ID != "mysql-df-001" {
				t.Errorf("ID = %v, want %v", got.ID, "mysql-df-001")
			}
			if got.DatatypeID != datatypeID {
				t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
			}
			if got.FieldID != fieldID {
				t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
			}
		})
	}
}

func TestMysqlDatabase_MapDatatypeField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapDatatypeField(mdbm.DatatypesFields{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.ID != "" {
		t.Errorf("ID = %v, want zero value", got.ID)
	}
}

// --- MySQL MysqlDatabase.MapCreateDatatypeFieldParams tests ---

func TestMysqlDatabase_MapCreateDatatypeFieldParams_WithExplicitID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	input := CreateDatatypeFieldParams{
		ID:         "mysql-explicit-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  77,
	}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.ID != "mysql-explicit-001" {
		t.Errorf("ID = %q, want %q", got.ID, "mysql-explicit-001")
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	// MySQL uses int32 for SortOrder
	if got.SortOrder != int32(77) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(77))
	}
}

func TestMysqlDatabase_MapCreateDatatypeFieldParams_GeneratesIDWhenEmpty(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := CreateDatatypeFieldParams{ID: ""}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated when input ID is empty")
	}
}

func TestMysqlDatabase_MapCreateDatatypeFieldParams_Int64ToInt32SortOrder(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := CreateDatatypeFieldParams{
		ID:        "mysql-conv-001",
		SortOrder: 12345,
	}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.SortOrder != int32(12345) {
		t.Errorf("SortOrder = %d, want %d (int64 to int32 conversion)", got.SortOrder, int32(12345))
	}
}

// --- MySQL MysqlDatabase.MapUpdateDatatypeFieldParams tests ---

func TestMysqlDatabase_MapUpdateDatatypeFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	input := UpdateDatatypeFieldParams{
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  50,
		ID:         "mysql-update-001",
	}

	got := d.MapUpdateDatatypeFieldParams(input)

	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	// MySQL uses int32
	if got.SortOrder != int32(50) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(50))
	}
	if got.ID != "mysql-update-001" {
		t.Errorf("ID = %q, want %q", got.ID, "mysql-update-001")
	}
}

func TestMysqlDatabase_MapUpdateDatatypeFieldParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapUpdateDatatypeFieldParams(UpdateDatatypeFieldParams{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
}

// --- PostgreSQL PsqlDatabase.MapDatatypeField tests ---

func TestPsqlDatabase_MapDatatypeField_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	tests := []struct {
		name            string
		inputSortOrder  int32
		wantSortOrder64 int64
	}{
		{"positive sort order", 8, 8},
		{"zero sort order", 0, 0},
		{"negative sort order", -50, -50},
		{"max int32", math.MaxInt32, int64(math.MaxInt32)},
		{"min int32", math.MinInt32, int64(math.MinInt32)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbp.DatatypesFields{
				ID:         "psql-df-001",
				DatatypeID: datatypeID,
				FieldID:    fieldID,
				SortOrder:  tt.inputSortOrder,
			}

			got := d.MapDatatypeField(input)

			if got.SortOrder != tt.wantSortOrder64 {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.wantSortOrder64)
			}
			if got.ID != "psql-df-001" {
				t.Errorf("ID = %v, want %v", got.ID, "psql-df-001")
			}
			if got.DatatypeID != datatypeID {
				t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
			}
			if got.FieldID != fieldID {
				t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
			}
		})
	}
}

func TestPsqlDatabase_MapDatatypeField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapDatatypeField(mdbp.DatatypesFields{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.ID != "" {
		t.Errorf("ID = %v, want zero value", got.ID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateDatatypeFieldParams tests ---

func TestPsqlDatabase_MapCreateDatatypeFieldParams_WithExplicitID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	input := CreateDatatypeFieldParams{
		ID:         "psql-explicit-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  33,
	}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.ID != "psql-explicit-001" {
		t.Errorf("ID = %q, want %q", got.ID, "psql-explicit-001")
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	// PostgreSQL uses int32
	if got.SortOrder != int32(33) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(33))
	}
}

func TestPsqlDatabase_MapCreateDatatypeFieldParams_GeneratesIDWhenEmpty(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := CreateDatatypeFieldParams{ID: ""}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated when input ID is empty")
	}
}

func TestPsqlDatabase_MapCreateDatatypeFieldParams_Int64ToInt32SortOrder(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := CreateDatatypeFieldParams{
		ID:        "psql-conv-001",
		SortOrder: 98765,
	}

	got := d.MapCreateDatatypeFieldParams(input)

	if got.SortOrder != int32(98765) {
		t.Errorf("SortOrder = %d, want %d (int64 to int32 conversion)", got.SortOrder, int32(98765))
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateDatatypeFieldParams tests ---

func TestPsqlDatabase_MapUpdateDatatypeFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	input := UpdateDatatypeFieldParams{
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  25,
		ID:         "psql-update-001",
	}

	got := d.MapUpdateDatatypeFieldParams(input)

	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	// PostgreSQL uses int32
	if got.SortOrder != int32(25) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(25))
	}
	if got.ID != "psql-update-001" {
		t.Errorf("ID = %q, want %q", got.ID, "psql-update-001")
	}
}

func TestPsqlDatabase_MapUpdateDatatypeFieldParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapUpdateDatatypeFieldParams(UpdateDatatypeFieldParams{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
}

// --- Cross-database mapper consistency ---
// Verify that all three database mappers produce identical DatatypeFields
// from equivalent input.

func TestCrossDatabaseMapDatatypeField_Consistency(t *testing.T) {
	t.Parallel()
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()

	sqliteInput := mdb.DatatypesFields{
		ID:         "cross-db-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  7,
	}
	mysqlInput := mdbm.DatatypesFields{
		ID:         "cross-db-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  7,
	}
	psqlInput := mdbp.DatatypesFields{
		ID:         "cross-db-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  7,
	}

	sqliteResult := Database{}.MapDatatypeField(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapDatatypeField(mysqlInput)
	psqlResult := PsqlDatabase{}.MapDatatypeField(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

func TestCrossDatabaseMapDatatypeField_ZeroValueConsistency(t *testing.T) {
	t.Parallel()

	sqliteResult := Database{}.MapDatatypeField(mdb.DatatypesFields{})
	mysqlResult := MysqlDatabase{}.MapDatatypeField(mdbm.DatatypesFields{})
	psqlResult := PsqlDatabase{}.MapDatatypeField(mdbp.DatatypesFields{})

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL zero values differ:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL zero values differ:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateDatatypeFieldParams - ID generation ---

func TestCrossDatabaseMapCreateDatatypeFieldParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	input := CreateDatatypeFieldParams{
		ID:        "",
		SortOrder: 1,
	}

	sqliteResult := Database{}.MapCreateDatatypeFieldParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateDatatypeFieldParams(input)
	psqlResult := PsqlDatabase{}.MapCreateDatatypeFieldParams(input)

	if sqliteResult.ID == "" {
		t.Error("SQLite: expected non-empty generated ID")
	}
	if mysqlResult.ID == "" {
		t.Error("MySQL: expected non-empty generated ID")
	}
	if psqlResult.ID == "" {
		t.Error("PostgreSQL: expected non-empty generated ID")
	}

	// Each call should generate a unique ID
	if sqliteResult.ID == mysqlResult.ID {
		t.Error("SQLite and MySQL generated the same ID -- each call should be unique")
	}
	if sqliteResult.ID == psqlResult.ID {
		t.Error("SQLite and PostgreSQL generated the same ID -- each call should be unique")
	}
}

func TestCrossDatabaseMapCreateDatatypeFieldParams_ExplicitIDPreserved(t *testing.T) {
	t.Parallel()
	input := CreateDatatypeFieldParams{
		ID:        "explicit-cross-db",
		SortOrder: 5,
	}

	sqliteResult := Database{}.MapCreateDatatypeFieldParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateDatatypeFieldParams(input)
	psqlResult := PsqlDatabase{}.MapCreateDatatypeFieldParams(input)

	if sqliteResult.ID != "explicit-cross-db" {
		t.Errorf("SQLite ID = %q, want %q", sqliteResult.ID, "explicit-cross-db")
	}
	if mysqlResult.ID != "explicit-cross-db" {
		t.Errorf("MySQL ID = %q, want %q", mysqlResult.ID, "explicit-cross-db")
	}
	if psqlResult.ID != "explicit-cross-db" {
		t.Errorf("PostgreSQL ID = %q, want %q", psqlResult.ID, "explicit-cross-db")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewDatatypeFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-df-001",
		IP:        "10.0.0.1",
	}
	datatypeID := types.NewDatatypeID()
	fieldID := types.NewFieldID()
	params := CreateDatatypeFieldParams{
		ID:         "cmd-create-001",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  5,
	}

	cmd := Database{}.NewDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	p, ok := cmd.Params().(CreateDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateDatatypeFieldParams", cmd.Params())
	}
	if p.ID != "cmd-create-001" {
		t.Errorf("Params().ID = %v, want %v", p.ID, "cmd-create-001")
	}
	if p.SortOrder != 5 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 5)
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewDatatypeFieldCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	cmd := NewDatatypeFieldCmd{}

	row := mdb.DatatypesFields{ID: "row-id-001"}
	got := cmd.GetID(row)
	if got != "row-id-001" {
		t.Errorf("GetID() = %q, want %q", got, "row-id-001")
	}
}

func TestUpdateDatatypeFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateDatatypeFieldParams{
		DatatypeID: types.NewDatatypeID(),
		FieldID:    types.NewFieldID(),
		SortOrder:  42,
		ID:         "update-cmd-001",
	}

	cmd := Database{}.UpdateDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "update-cmd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "update-cmd-001")
	}
	p, ok := cmd.Params().(UpdateDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateDatatypeFieldParams", cmd.Params())
	}
	if p.SortOrder != 42 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 42)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestUpdateDatatypeFieldSortOrderCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}

	cmd := Database{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, "sort-order-cmd-001", 77)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "sort-order-cmd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "sort-order-cmd-001")
	}
	p, ok := cmd.Params().(map[string]any)
	if !ok {
		t.Fatalf("Params() returned %T, want map[string]any", cmd.Params())
	}
	if p["id"] != "sort-order-cmd-001" {
		t.Errorf("Params()[\"id\"] = %v, want %v", p["id"], "sort-order-cmd-001")
	}
	if p["sort_order"] != int64(77) {
		t.Errorf("Params()[\"sort_order\"] = %v, want %v", p["sort_order"], int64(77))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteDatatypeFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}

	cmd := Database{}.DeleteDatatypeFieldCmd(ctx, ac, "delete-cmd-001")

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "delete-cmd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "delete-cmd-001")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewDatatypeFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-df-001",
		IP:        "192.168.1.1",
	}
	params := CreateDatatypeFieldParams{
		ID:        "mysql-cmd-001",
		SortOrder: 3,
	}

	cmd := MysqlDatabase{}.NewDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	p, ok := cmd.Params().(CreateDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateDatatypeFieldParams", cmd.Params())
	}
	if p.SortOrder != 3 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 3)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewDatatypeFieldCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewDatatypeFieldCmdMysql{}

	row := mdbm.DatatypesFields{ID: "mysql-row-001"}
	got := cmd.GetID(row)
	if got != "mysql-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "mysql-row-001")
	}
}

func TestUpdateDatatypeFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateDatatypeFieldParams{
		SortOrder: 55,
		ID:        "mysql-update-cmd-001",
	}

	cmd := MysqlDatabase{}.UpdateDatatypeFieldCmd(ctx, ac, params)

	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "mysql-update-cmd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "mysql-update-cmd-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateDatatypeFieldParams", cmd.Params())
	}
	if p.SortOrder != 55 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 55)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestUpdateDatatypeFieldSortOrderCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}

	cmd := MysqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, "mysql-sort-001", 88)

	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "mysql-sort-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "mysql-sort-001")
	}
	p, ok := cmd.Params().(map[string]any)
	if !ok {
		t.Fatalf("Params() returned %T, want map[string]any", cmd.Params())
	}
	if p["id"] != "mysql-sort-001" {
		t.Errorf("Params()[\"id\"] = %v, want %v", p["id"], "mysql-sort-001")
	}
	if p["sort_order"] != int64(88) {
		t.Errorf("Params()[\"sort_order\"] = %v, want %v", p["sort_order"], int64(88))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteDatatypeFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}

	cmd := MysqlDatabase{}.DeleteDatatypeFieldCmd(ctx, ac, "mysql-delete-001")

	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "mysql-delete-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "mysql-delete-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Audited Command Accessor tests ---

func TestNewDatatypeFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-df-001",
		IP:        "172.16.0.1",
	}
	params := CreateDatatypeFieldParams{
		ID:        "psql-cmd-001",
		SortOrder: 8,
	}

	cmd := PsqlDatabase{}.NewDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	p, ok := cmd.Params().(CreateDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateDatatypeFieldParams", cmd.Params())
	}
	if p.SortOrder != 8 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 8)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewDatatypeFieldCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewDatatypeFieldCmdPsql{}

	row := mdbp.DatatypesFields{ID: "psql-row-001"}
	got := cmd.GetID(row)
	if got != "psql-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "psql-row-001")
	}
}

func TestUpdateDatatypeFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateDatatypeFieldParams{
		SortOrder: 66,
		ID:        "psql-update-cmd-001",
	}

	cmd := PsqlDatabase{}.UpdateDatatypeFieldCmd(ctx, ac, params)

	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "psql-update-cmd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "psql-update-cmd-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateDatatypeFieldParams", cmd.Params())
	}
	if p.SortOrder != 66 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 66)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestUpdateDatatypeFieldSortOrderCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}

	cmd := PsqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, "psql-sort-001", 99)

	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "psql-sort-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "psql-sort-001")
	}
	p, ok := cmd.Params().(map[string]any)
	if !ok {
		t.Fatalf("Params() returned %T, want map[string]any", cmd.Params())
	}
	if p["id"] != "psql-sort-001" {
		t.Errorf("Params()[\"id\"] = %v, want %v", p["id"], "psql-sort-001")
	}
	if p["sort_order"] != int64(99) {
		t.Errorf("Params()[\"sort_order\"] = %v, want %v", p["sort_order"], int64(99))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteDatatypeFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}

	cmd := PsqlDatabase{}.DeleteDatatypeFieldCmd(ctx, ac, "psql-delete-001")

	if cmd.TableName() != "datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes_fields")
	}
	if cmd.GetID() != "psql-delete-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "psql-delete-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- Cross-database audited command consistency ---

func TestAuditedDatatypeFieldCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateDatatypeFieldParams{}
	updateParams := UpdateDatatypeFieldParams{ID: "table-name-check"}

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewDatatypeFieldCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateDatatypeFieldCmd(ctx, ac, updateParams).TableName()},
		{"SQLite UpdateSortOrder", Database{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, "id", 0).TableName()},
		{"SQLite Delete", Database{}.DeleteDatatypeFieldCmd(ctx, ac, "id").TableName()},
		{"MySQL Create", MysqlDatabase{}.NewDatatypeFieldCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateDatatypeFieldCmd(ctx, ac, updateParams).TableName()},
		{"MySQL UpdateSortOrder", MysqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, "id", 0).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteDatatypeFieldCmd(ctx, ac, "id").TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewDatatypeFieldCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateDatatypeFieldCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL UpdateSortOrder", PsqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, "id", 0).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteDatatypeFieldCmd(ctx, ac, "id").TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "datatypes_fields" {
				t.Errorf("TableName() = %q, want %q", c.name, "datatypes_fields")
			}
		})
	}
}

func TestAuditedDatatypeFieldCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateDatatypeFieldParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewDatatypeFieldCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewDatatypeFieldCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewDatatypeFieldCmd(ctx, ac, createParams).Recorder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recorder == nil {
				t.Fatalf("%s recorder is nil", tt.name)
			}
		})
	}
}

// --- Audited Command GetID consistency ---

func TestAuditedDatatypeFieldCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	t.Run("UpdateCmd GetID returns params ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateDatatypeFieldParams{ID: "getid-check-update"}

		sqliteCmd := Database{}.UpdateDatatypeFieldCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateDatatypeFieldCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateDatatypeFieldCmd(ctx, ac, params)

		wantID := "getid-check-update"
		if sqliteCmd.GetID() != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), wantID)
		}
		if mysqlCmd.GetID() != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), wantID)
		}
		if psqlCmd.GetID() != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), wantID)
		}
	})

	t.Run("UpdateSortOrderCmd GetID returns id", func(t *testing.T) {
		t.Parallel()
		wantID := "getid-check-sort"

		sqliteCmd := Database{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, wantID, 1)
		mysqlCmd := MysqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, wantID, 1)
		psqlCmd := PsqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(ctx, ac, wantID, 1)

		if sqliteCmd.GetID() != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), wantID)
		}
		if mysqlCmd.GetID() != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), wantID)
		}
		if psqlCmd.GetID() != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), wantID)
		}
	})

	t.Run("DeleteCmd GetID returns id", func(t *testing.T) {
		t.Parallel()
		wantID := "getid-check-delete"

		sqliteCmd := Database{}.DeleteDatatypeFieldCmd(ctx, ac, wantID)
		mysqlCmd := MysqlDatabase{}.DeleteDatatypeFieldCmd(ctx, ac, wantID)
		psqlCmd := PsqlDatabase{}.DeleteDatatypeFieldCmd(ctx, ac, wantID)

		if sqliteCmd.GetID() != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), wantID)
		}
		if mysqlCmd.GetID() != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), wantID)
		}
		if psqlCmd.GetID() != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), wantID)
		}
	})

	t.Run("CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		wantID := "create-row-id-001"

		sqliteCmd := NewDatatypeFieldCmd{}
		mysqlCmd := NewDatatypeFieldCmdMysql{}
		psqlCmd := NewDatatypeFieldCmdPsql{}

		sqliteRow := mdb.DatatypesFields{ID: wantID}
		mysqlRow := mdbm.DatatypesFields{ID: wantID}
		psqlRow := mdbp.DatatypesFields{ID: wantID}

		if sqliteCmd.GetID(sqliteRow) != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), wantID)
		}
		if mysqlCmd.GetID(mysqlRow) != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), wantID)
		}
		if psqlCmd.GetID(psqlRow) != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), wantID)
		}
	})
}

// --- Edge cases: empty IDs ---

func TestUpdateDatatypeFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateDatatypeFieldParams{ID: ""}

	sqliteCmd := Database{}.UpdateDatatypeFieldCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateDatatypeFieldCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateDatatypeFieldCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestUpdateDatatypeFieldSortOrderCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()

	sqliteCmd := Database{}.UpdateDatatypeFieldSortOrderCmd(context.Background(), audited.AuditContext{}, "", 0)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(context.Background(), audited.AuditContext{}, "", 0)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateDatatypeFieldSortOrderCmd(context.Background(), audited.AuditContext{}, "", 0)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteDatatypeFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()

	sqliteCmd := Database{}.DeleteDatatypeFieldCmd(context.Background(), audited.AuditContext{}, "")
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteDatatypeFieldCmd(context.Background(), audited.AuditContext{}, "")
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteDatatypeFieldCmd(context.Background(), audited.AuditContext{}, "")
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	// SQLite
	_ audited.CreateCommand[mdb.DatatypesFields] = NewDatatypeFieldCmd{}
	_ audited.UpdateCommand[mdb.DatatypesFields]  = UpdateDatatypeFieldCmd{}
	_ audited.UpdateCommand[mdb.DatatypesFields]  = UpdateDatatypeFieldSortOrderCmd{}
	_ audited.DeleteCommand[mdb.DatatypesFields]  = DeleteDatatypeFieldCmd{}

	// MySQL
	_ audited.CreateCommand[mdbm.DatatypesFields] = NewDatatypeFieldCmdMysql{}
	_ audited.UpdateCommand[mdbm.DatatypesFields]  = UpdateDatatypeFieldCmdMysql{}
	_ audited.UpdateCommand[mdbm.DatatypesFields]  = UpdateDatatypeFieldSortOrderCmdMysql{}
	_ audited.DeleteCommand[mdbm.DatatypesFields]  = DeleteDatatypeFieldCmdMysql{}

	// PostgreSQL
	_ audited.CreateCommand[mdbp.DatatypesFields] = NewDatatypeFieldCmdPsql{}
	_ audited.UpdateCommand[mdbp.DatatypesFields]  = UpdateDatatypeFieldCmdPsql{}
	_ audited.UpdateCommand[mdbp.DatatypesFields]  = UpdateDatatypeFieldSortOrderCmdPsql{}
	_ audited.DeleteCommand[mdbp.DatatypesFields]  = DeleteDatatypeFieldCmdPsql{}
)
