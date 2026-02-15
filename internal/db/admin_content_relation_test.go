// White-box tests for admin_content_relation.go: wrapper structs, mapper methods
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
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// acrTestFixture returns fully populated test data for AdminContentRelations tests.
func acrTestFixture() (AdminContentRelations, types.AdminContentRelationID, types.AdminContentID, types.AdminContentID, types.AdminFieldID, types.Timestamp) {
	relationID := types.NewAdminContentRelationID()
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()
	ts := types.NewTimestamp(time.Date(2025, 7, 20, 14, 30, 0, 0, time.UTC))
	acr := AdminContentRelations{
		AdminContentRelationID: relationID,
		SourceContentID:        sourceID,
		TargetContentID:        targetID,
		AdminFieldID:           fieldID,
		SortOrder:              5,
		DateCreated:            ts,
	}
	return acr, relationID, sourceID, targetID, fieldID, ts
}

// --- MapStringAdminContentRelation tests ---

func TestMapStringAdminContentRelation_AllFields(t *testing.T) {
	t.Parallel()
	acr, relationID, sourceID, targetID, fieldID, ts := acrTestFixture()

	got := MapStringAdminContentRelation(acr)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AdminContentRelationID", got.AdminContentRelationID, relationID.String()},
		{"SourceContentID", got.SourceContentID, sourceID.String()},
		{"TargetContentID", got.TargetContentID, targetID.String()},
		{"AdminFieldID", got.AdminFieldID, fieldID.String()},
		{"SortOrder", got.SortOrder, "5"},
		{"DateCreated", got.DateCreated, ts.String()},
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

func TestMapStringAdminContentRelation_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringAdminContentRelation(AdminContentRelations{})

	if got.AdminContentRelationID != "" {
		t.Errorf("AdminContentRelationID = %q, want empty string", got.AdminContentRelationID)
	}
	if got.SourceContentID != "" {
		t.Errorf("SourceContentID = %q, want empty string", got.SourceContentID)
	}
	if got.TargetContentID != "" {
		t.Errorf("TargetContentID = %q, want empty string", got.TargetContentID)
	}
	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %q, want empty string", got.AdminFieldID)
	}
	if got.SortOrder != "0" {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, "0")
	}
}

func TestMapStringAdminContentRelation_NegativeSortOrder(t *testing.T) {
	t.Parallel()
	acr := AdminContentRelations{SortOrder: -1}
	got := MapStringAdminContentRelation(acr)
	if got.SortOrder != "-1" {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, "-1")
	}
}

func TestMapStringAdminContentRelation_LargeSortOrder(t *testing.T) {
	t.Parallel()
	acr := AdminContentRelations{SortOrder: math.MaxInt64}
	got := MapStringAdminContentRelation(acr)
	want := fmt.Sprintf("%d", int64(math.MaxInt64))
	if got.SortOrder != want {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, want)
	}
}

// --- SQLite Database.MapAdminContentRelation tests ---

func TestDatabase_MapAdminContentRelation_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	relationID := types.NewAdminContentRelationID()
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()
	ts := types.NewTimestamp(time.Date(2025, 5, 10, 8, 0, 0, 0, time.UTC))

	input := mdb.AdminContentRelations{
		AdminContentRelationID: relationID,
		SourceContentID:        sourceID,
		TargetContentID:        targetID,
		AdminFieldID:           fieldID,
		SortOrder:              42,
		DateCreated:            ts,
	}

	got := d.MapAdminContentRelation(input)

	if got.AdminContentRelationID != relationID {
		t.Errorf("AdminContentRelationID = %v, want %v", got.AdminContentRelationID, relationID)
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
	if got.SortOrder != 42 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 42)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestDatabase_MapAdminContentRelation_ZeroAdminFieldID(t *testing.T) {
	t.Parallel()
	d := Database{}

	// When AdminFieldID is zero, the mapped result should also be zero
	input := mdb.AdminContentRelations{
		AdminFieldID: types.AdminFieldID(""),
	}

	got := d.MapAdminContentRelation(input)

	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

func TestDatabase_MapAdminContentRelation_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminContentRelation(mdb.AdminContentRelations{})

	if got.AdminContentRelationID != "" {
		t.Errorf("AdminContentRelationID = %v, want zero value", got.AdminContentRelationID)
	}
	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
}

// --- SQLite Database.MapCreateAdminContentRelationParams tests ---

func TestDatabase_MapCreateAdminContentRelationParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()

	input := CreateAdminContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		AdminFieldID:    fieldID,
		SortOrder:       3,
		DateCreated:     ts,
	}

	got := d.MapCreateAdminContentRelationParams(input)

	// A new ID should always be generated
	if got.AdminContentRelationID.IsZero() {
		t.Fatal("expected non-zero AdminContentRelationID to be generated")
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
	if got.SortOrder != 3 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 3)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestDatabase_MapCreateAdminContentRelationParams_ZeroFieldID(t *testing.T) {
	t.Parallel()
	d := Database{}

	// When AdminFieldID is zero, the mapped AdminFieldID should also be zero
	input := CreateAdminContentRelationParams{
		AdminFieldID: types.AdminFieldID(""),
	}

	got := d.MapCreateAdminContentRelationParams(input)

	if !got.AdminFieldID.IsZero() {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

func TestDatabase_MapCreateAdminContentRelationParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}

	// Two calls should produce different IDs
	got1 := d.MapCreateAdminContentRelationParams(CreateAdminContentRelationParams{})
	got2 := d.MapCreateAdminContentRelationParams(CreateAdminContentRelationParams{})

	if got1.AdminContentRelationID == got2.AdminContentRelationID {
		t.Error("two calls produced identical AdminContentRelationIDs")
	}
}

// --- SQLite Database.MapUpdateAdminContentRelationSortOrderParams tests ---

func TestDatabase_MapUpdateAdminContentRelationSortOrderParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	relationID := types.NewAdminContentRelationID()

	input := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: relationID,
		SortOrder:              10,
	}

	got := d.MapUpdateAdminContentRelationSortOrderParams(input)

	if got.AdminContentRelationID != relationID {
		t.Errorf("AdminContentRelationID = %v, want %v", got.AdminContentRelationID, relationID)
	}
	if got.SortOrder != 10 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 10)
	}
}

func TestDatabase_MapUpdateAdminContentRelationSortOrderParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateAdminContentRelationSortOrderParams(UpdateAdminContentRelationSortOrderParams{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.AdminContentRelationID != "" {
		t.Errorf("AdminContentRelationID = %v, want zero value", got.AdminContentRelationID)
	}
}

// --- MySQL MysqlDatabase.MapAdminContentRelation tests ---

func TestMysqlDatabase_MapAdminContentRelation_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	relationID := types.NewAdminContentRelationID()
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()
	ts := types.NewTimestamp(time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC))

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
			input := mdbm.AdminContentRelations{
				AdminContentRelationID: relationID,
				SourceContentID:        sourceID,
				TargetContentID:        targetID,
				AdminFieldID:           fieldID,
				SortOrder:              tt.inputSortOrder,
				DateCreated:            ts,
			}

			got := d.MapAdminContentRelation(input)

			if got.SortOrder != tt.wantSortOrder64 {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.wantSortOrder64)
			}
			if got.AdminContentRelationID != relationID {
				t.Errorf("AdminContentRelationID = %v, want %v", got.AdminContentRelationID, relationID)
			}
			if got.SourceContentID != sourceID {
				t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
			}
			if got.TargetContentID != targetID {
				t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
			}
			if got.AdminFieldID != fieldID {
				t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
			}
		})
	}
}

func TestMysqlDatabase_MapAdminContentRelation_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminContentRelation(mdbm.AdminContentRelations{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminContentRelationParams tests ---

func TestMysqlDatabase_MapCreateAdminContentRelationParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()

	input := CreateAdminContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		AdminFieldID:    fieldID,
		SortOrder:       7,
		DateCreated:     ts,
	}

	got := d.MapCreateAdminContentRelationParams(input)

	if got.AdminContentRelationID.IsZero() {
		t.Fatal("expected non-zero AdminContentRelationID to be generated")
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
	// MySQL uses int32 for SortOrder
	if got.SortOrder != int32(7) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(7))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestMysqlDatabase_MapCreateAdminContentRelationParams_ZeroFieldID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreateAdminContentRelationParams{
		AdminFieldID: types.AdminFieldID(""),
	}

	got := d.MapCreateAdminContentRelationParams(input)

	if !got.AdminFieldID.IsZero() {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- MySQL MysqlDatabase.MapUpdateAdminContentRelationSortOrderParams tests ---

func TestMysqlDatabase_MapUpdateAdminContentRelationSortOrderParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	relationID := types.NewAdminContentRelationID()

	input := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: relationID,
		SortOrder:              15,
	}

	got := d.MapUpdateAdminContentRelationSortOrderParams(input)

	if got.AdminContentRelationID != relationID {
		t.Errorf("AdminContentRelationID = %v, want %v", got.AdminContentRelationID, relationID)
	}
	// MySQL uses int32
	if got.SortOrder != int32(15) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(15))
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminContentRelation tests ---

func TestPsqlDatabase_MapAdminContentRelation_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	relationID := types.NewAdminContentRelationID()
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()
	ts := types.NewTimestamp(time.Date(2025, 4, 15, 16, 0, 0, 0, time.UTC))

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
			input := mdbp.AdminContentRelations{
				AdminContentRelationID: relationID,
				SourceContentID:        sourceID,
				TargetContentID:        targetID,
				AdminFieldID:           fieldID,
				SortOrder:              tt.inputSortOrder,
				DateCreated:            ts,
			}

			got := d.MapAdminContentRelation(input)

			if got.SortOrder != tt.wantSortOrder64 {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.wantSortOrder64)
			}
			if got.AdminContentRelationID != relationID {
				t.Errorf("AdminContentRelationID = %v, want %v", got.AdminContentRelationID, relationID)
			}
			if got.SourceContentID != sourceID {
				t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
			}
			if got.TargetContentID != targetID {
				t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
			}
			if got.AdminFieldID != fieldID {
				t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
			}
		})
	}
}

func TestPsqlDatabase_MapAdminContentRelation_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminContentRelation(mdbp.AdminContentRelations{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminContentRelationParams tests ---

func TestPsqlDatabase_MapCreateAdminContentRelationParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()

	input := CreateAdminContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		AdminFieldID:    fieldID,
		SortOrder:       12,
		DateCreated:     ts,
	}

	got := d.MapCreateAdminContentRelationParams(input)

	if got.AdminContentRelationID.IsZero() {
		t.Fatal("expected non-zero AdminContentRelationID to be generated")
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
	// PostgreSQL uses int32 for SortOrder
	if got.SortOrder != int32(12) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(12))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestPsqlDatabase_MapCreateAdminContentRelationParams_ZeroFieldID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateAdminContentRelationParams{
		AdminFieldID: types.AdminFieldID(""),
	}

	got := d.MapCreateAdminContentRelationParams(input)

	if !got.AdminFieldID.IsZero() {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminContentRelationSortOrderParams tests ---

func TestPsqlDatabase_MapUpdateAdminContentRelationSortOrderParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	relationID := types.NewAdminContentRelationID()

	input := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: relationID,
		SortOrder:              20,
	}

	got := d.MapUpdateAdminContentRelationSortOrderParams(input)

	if got.AdminContentRelationID != relationID {
		t.Errorf("AdminContentRelationID = %v, want %v", got.AdminContentRelationID, relationID)
	}
	// PostgreSQL uses int32
	if got.SortOrder != int32(20) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(20))
	}
}

// --- Cross-database mapper consistency ---
// Verify that all three database mappers produce identical AdminContentRelations
// from equivalent input.

func TestCrossDatabaseMapAdminContentRelation_Consistency(t *testing.T) {
	t.Parallel()
	relationID := types.NewAdminContentRelationID()
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()
	ts := types.NewTimestamp(time.Date(2025, 8, 1, 9, 0, 0, 0, time.UTC))
	sqliteInput := mdb.AdminContentRelations{
		AdminContentRelationID: relationID,
		SourceContentID:        sourceID,
		TargetContentID:        targetID,
		AdminFieldID:           fieldID,
		SortOrder:              7,
		DateCreated:            ts,
	}
	mysqlInput := mdbm.AdminContentRelations{
		AdminContentRelationID: relationID,
		SourceContentID:        sourceID,
		TargetContentID:        targetID,
		AdminFieldID:           fieldID,
		SortOrder:              7,
		DateCreated:            ts,
	}
	psqlInput := mdbp.AdminContentRelations{
		AdminContentRelationID: relationID,
		SourceContentID:        sourceID,
		TargetContentID:        targetID,
		AdminFieldID:           fieldID,
		SortOrder:              7,
		DateCreated:            ts,
	}

	sqliteResult := Database{}.MapAdminContentRelation(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminContentRelation(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminContentRelation(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminContentRelationParams - ID generation ---

func TestCrossDatabaseMapCreateAdminContentRelationParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()

	input := CreateAdminContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		SortOrder:       1,
		DateCreated:     ts,
	}

	sqliteResult := Database{}.MapCreateAdminContentRelationParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminContentRelationParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminContentRelationParams(input)

	if sqliteResult.AdminContentRelationID.IsZero() {
		t.Error("SQLite: expected non-zero generated AdminContentRelationID")
	}
	if mysqlResult.AdminContentRelationID.IsZero() {
		t.Error("MySQL: expected non-zero generated AdminContentRelationID")
	}
	if psqlResult.AdminContentRelationID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated AdminContentRelationID")
	}

	// Each call should generate a unique ID
	if sqliteResult.AdminContentRelationID == mysqlResult.AdminContentRelationID {
		t.Error("SQLite and MySQL generated the same AdminContentRelationID -- each call should be unique")
	}
	if sqliteResult.AdminContentRelationID == psqlResult.AdminContentRelationID {
		t.Error("SQLite and PostgreSQL generated the same AdminContentRelationID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewAdminContentRelationCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-acr-001",
		IP:        "10.0.0.1",
	}
	sourceID := types.NewAdminContentID()
	targetID := types.NewAdminContentID()
	fieldID := types.NewAdminFieldID()
	params := CreateAdminContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		AdminFieldID:    fieldID,
		SortOrder:       5,
		DateCreated:     ts,
	}

	cmd := Database{}.NewAdminContentRelationCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	p, ok := cmd.Params().(CreateAdminContentRelationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentRelationParams", cmd.Params())
	}
	if p.SourceContentID != sourceID {
		t.Errorf("Params().SourceContentID = %v, want %v", p.SourceContentID, sourceID)
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

func TestNewAdminContentRelationCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	relationID := types.NewAdminContentRelationID()
	cmd := NewAdminContentRelationCmd{}

	row := mdb.AdminContentRelations{AdminContentRelationID: relationID}
	got := cmd.GetID(row)
	if got != string(relationID) {
		t.Errorf("GetID() = %q, want %q", got, string(relationID))
	}
}

func TestUpdateAdminContentRelationSortOrderCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewAdminContentRelationID()
	params := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: relationID,
		SortOrder:              99,
	}

	cmd := Database{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	if cmd.GetID() != string(relationID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(relationID))
	}
	p, ok := cmd.Params().(UpdateAdminContentRelationSortOrderParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentRelationSortOrderParams", cmd.Params())
	}
	if p.SortOrder != 99 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 99)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentRelationCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewAdminContentRelationID()

	cmd := Database{}.DeleteAdminContentRelationCmd(ctx, ac, relationID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	if cmd.GetID() != string(relationID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(relationID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewAdminContentRelationCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-acr-001",
		IP:        "192.168.1.1",
	}
	params := CreateAdminContentRelationParams{
		SourceContentID: types.NewAdminContentID(),
		TargetContentID: types.NewAdminContentID(),
		AdminFieldID:    types.NewAdminFieldID(),
		SortOrder:       3,
		DateCreated:     ts,
	}

	cmd := MysqlDatabase{}.NewAdminContentRelationCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	p, ok := cmd.Params().(CreateAdminContentRelationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentRelationParams", cmd.Params())
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

func TestNewAdminContentRelationCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	relationID := types.NewAdminContentRelationID()
	cmd := NewAdminContentRelationCmdMysql{}

	row := mdbm.AdminContentRelations{AdminContentRelationID: relationID}
	got := cmd.GetID(row)
	if got != string(relationID) {
		t.Errorf("GetID() = %q, want %q", got, string(relationID))
	}
}

func TestUpdateAdminContentRelationSortOrderCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewAdminContentRelationID()
	params := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: relationID,
		SortOrder:              55,
	}

	cmd := MysqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, params)

	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	if cmd.GetID() != string(relationID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(relationID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminContentRelationSortOrderParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentRelationSortOrderParams", cmd.Params())
	}
	if p.SortOrder != 55 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 55)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentRelationCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewAdminContentRelationID()

	cmd := MysqlDatabase{}.DeleteAdminContentRelationCmd(ctx, ac, relationID)

	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	if cmd.GetID() != string(relationID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(relationID))
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

func TestNewAdminContentRelationCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-acr-001",
		IP:        "172.16.0.1",
	}
	params := CreateAdminContentRelationParams{
		SourceContentID: types.NewAdminContentID(),
		TargetContentID: types.NewAdminContentID(),
		AdminFieldID:    types.NewAdminFieldID(),
		SortOrder:       8,
		DateCreated:     ts,
	}

	cmd := PsqlDatabase{}.NewAdminContentRelationCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	p, ok := cmd.Params().(CreateAdminContentRelationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentRelationParams", cmd.Params())
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

func TestNewAdminContentRelationCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	relationID := types.NewAdminContentRelationID()
	cmd := NewAdminContentRelationCmdPsql{}

	row := mdbp.AdminContentRelations{AdminContentRelationID: relationID}
	got := cmd.GetID(row)
	if got != string(relationID) {
		t.Errorf("GetID() = %q, want %q", got, string(relationID))
	}
}

func TestUpdateAdminContentRelationSortOrderCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewAdminContentRelationID()
	params := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: relationID,
		SortOrder:              77,
	}

	cmd := PsqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, params)

	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	if cmd.GetID() != string(relationID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(relationID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminContentRelationSortOrderParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentRelationSortOrderParams", cmd.Params())
	}
	if p.SortOrder != 77 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 77)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentRelationCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewAdminContentRelationID()

	cmd := PsqlDatabase{}.DeleteAdminContentRelationCmd(ctx, ac, relationID)

	if cmd.TableName() != "admin_content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_relations")
	}
	if cmd.GetID() != string(relationID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(relationID))
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

func TestAuditedAdminContentRelationCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminContentRelationParams{}
	updateParams := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: types.NewAdminContentRelationID(),
	}
	relationID := types.NewAdminContentRelationID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewAdminContentRelationCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteAdminContentRelationCmd(ctx, ac, relationID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewAdminContentRelationCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteAdminContentRelationCmd(ctx, ac, relationID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewAdminContentRelationCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteAdminContentRelationCmd(ctx, ac, relationID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "admin_content_relations" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_content_relations")
			}
		})
	}
}

func TestAuditedAdminContentRelationCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminContentRelationParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminContentRelationCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminContentRelationCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminContentRelationCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminContentRelationCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	relationID := types.NewAdminContentRelationID()

	t.Run("UpdateCmd GetID returns relation ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateAdminContentRelationSortOrderParams{
			AdminContentRelationID: relationID,
			SortOrder:              1,
		}

		sqliteCmd := Database{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(ctx, ac, params)

		wantID := string(relationID)
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

	t.Run("DeleteCmd GetID returns relation ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminContentRelationCmd(ctx, ac, relationID)
		mysqlCmd := MysqlDatabase{}.DeleteAdminContentRelationCmd(ctx, ac, relationID)
		psqlCmd := PsqlDatabase{}.DeleteAdminContentRelationCmd(ctx, ac, relationID)

		wantID := string(relationID)
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
		testRelationID := types.NewAdminContentRelationID()
		wantID := string(testRelationID)

		sqliteCmd := NewAdminContentRelationCmd{}
		mysqlCmd := NewAdminContentRelationCmdMysql{}
		psqlCmd := NewAdminContentRelationCmdPsql{}

		sqliteRow := mdb.AdminContentRelations{AdminContentRelationID: testRelationID}
		mysqlRow := mdbm.AdminContentRelations{AdminContentRelationID: testRelationID}
		psqlRow := mdbp.AdminContentRelations{AdminContentRelationID: testRelationID}

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

func TestUpdateAdminContentRelationSortOrderCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: "",
	}

	sqliteCmd := Database{}.UpdateAdminContentRelationSortOrderCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminContentRelationSortOrderCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteAdminContentRelationCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.AdminContentRelationID("")

	sqliteCmd := Database{}.DeleteAdminContentRelationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminContentRelationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminContentRelationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.AdminContentRelations]  = NewAdminContentRelationCmd{}
	_ audited.UpdateCommand[mdb.AdminContentRelations]  = UpdateAdminContentRelationSortOrderCmd{}
	_ audited.DeleteCommand[mdb.AdminContentRelations]  = DeleteAdminContentRelationCmd{}
	_ audited.CreateCommand[mdbm.AdminContentRelations] = NewAdminContentRelationCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminContentRelations] = UpdateAdminContentRelationSortOrderCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminContentRelations] = DeleteAdminContentRelationCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminContentRelations] = NewAdminContentRelationCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminContentRelations] = UpdateAdminContentRelationSortOrderCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminContentRelations] = DeleteAdminContentRelationCmdPsql{}
)
