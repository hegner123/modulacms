// White-box tests for content_relation.go: wrapper structs, mapper methods
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

// crTestFixture returns a fully populated ContentRelations and its individual parts.
func crTestFixture() (ContentRelations, types.ContentRelationID, types.ContentID, types.ContentID, types.FieldID, types.Timestamp) {
	relationID := types.NewContentRelationID()
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()
	ts := types.NewTimestamp(time.Date(2025, 7, 20, 14, 30, 0, 0, time.UTC))
	cr := ContentRelations{
		ContentRelationID: relationID,
		SourceContentID:   sourceID,
		TargetContentID:   targetID,
		FieldID:           fieldID,
		SortOrder:         5,
		DateCreated:       ts,
	}
	return cr, relationID, sourceID, targetID, fieldID, ts
}

// --- MapStringContentRelation tests ---

func TestMapStringContentRelation_AllFields(t *testing.T) {
	t.Parallel()
	cr, relationID, sourceID, targetID, fieldID, ts := crTestFixture()

	got := MapStringContentRelation(cr)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ContentRelationID", got.ContentRelationID, relationID.String()},
		{"SourceContentID", got.SourceContentID, sourceID.String()},
		{"TargetContentID", got.TargetContentID, targetID.String()},
		{"FieldID", got.FieldID, fieldID.String()},
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

func TestMapStringContentRelation_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringContentRelation(ContentRelations{})

	if got.ContentRelationID != "" {
		t.Errorf("ContentRelationID = %q, want empty string", got.ContentRelationID)
	}
	if got.SourceContentID != "" {
		t.Errorf("SourceContentID = %q, want empty string", got.SourceContentID)
	}
	if got.TargetContentID != "" {
		t.Errorf("TargetContentID = %q, want empty string", got.TargetContentID)
	}
	if got.FieldID != "" {
		t.Errorf("FieldID = %q, want empty string", got.FieldID)
	}
	if got.SortOrder != "0" {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, "0")
	}
}

func TestMapStringContentRelation_NegativeSortOrder(t *testing.T) {
	t.Parallel()
	cr := ContentRelations{SortOrder: -1}
	got := MapStringContentRelation(cr)
	if got.SortOrder != "-1" {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, "-1")
	}
}

func TestMapStringContentRelation_LargeSortOrder(t *testing.T) {
	t.Parallel()
	cr := ContentRelations{SortOrder: math.MaxInt64}
	got := MapStringContentRelation(cr)
	want := fmt.Sprintf("%d", int64(math.MaxInt64))
	if got.SortOrder != want {
		t.Errorf("SortOrder = %q, want %q", got.SortOrder, want)
	}
}

// --- SQLite Database.MapContentRelation tests ---

func TestDatabase_MapContentRelation_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	relationID := types.NewContentRelationID()
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()
	ts := types.NewTimestamp(time.Date(2025, 5, 10, 8, 0, 0, 0, time.UTC))

	input := mdb.ContentRelations{
		ContentRelationID: relationID,
		SourceContentID:   sourceID,
		TargetContentID:   targetID,
		FieldID:           fieldID,
		SortOrder:         42,
		DateCreated:       ts,
	}

	got := d.MapContentRelation(input)

	if got.ContentRelationID != relationID {
		t.Errorf("ContentRelationID = %v, want %v", got.ContentRelationID, relationID)
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.SortOrder != 42 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 42)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestDatabase_MapContentRelation_ZeroFieldID(t *testing.T) {
	t.Parallel()
	d := Database{}

	// When FieldID is zero, the mapped result should also be zero
	input := mdb.ContentRelations{
		FieldID: types.FieldID(""),
	}

	got := d.MapContentRelation(input)

	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
}

func TestDatabase_MapContentRelation_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapContentRelation(mdb.ContentRelations{})

	if got.ContentRelationID != "" {
		t.Errorf("ContentRelationID = %v, want zero value", got.ContentRelationID)
	}
	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
}

// --- SQLite Database.MapCreateContentRelationParams tests ---

func TestDatabase_MapCreateContentRelationParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()

	input := CreateContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		FieldID:         fieldID,
		SortOrder:       3,
		DateCreated:     ts,
	}

	got := d.MapCreateContentRelationParams(input)

	// A new ID should always be generated
	if got.ContentRelationID.IsZero() {
		t.Fatal("expected non-zero ContentRelationID to be generated")
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.SortOrder != 3 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 3)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestDatabase_MapCreateContentRelationParams_ZeroFieldID(t *testing.T) {
	t.Parallel()
	d := Database{}

	// When FieldID is zero, the mapped FieldID should also be zero
	input := CreateContentRelationParams{
		FieldID: types.FieldID(""),
	}

	got := d.MapCreateContentRelationParams(input)

	if !got.FieldID.IsZero() {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
}

func TestDatabase_MapCreateContentRelationParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}

	// Two calls should produce different IDs
	got1 := d.MapCreateContentRelationParams(CreateContentRelationParams{})
	got2 := d.MapCreateContentRelationParams(CreateContentRelationParams{})

	if got1.ContentRelationID == got2.ContentRelationID {
		t.Error("two calls produced identical ContentRelationIDs")
	}
}

// --- SQLite Database.MapUpdateContentRelationSortOrderParams tests ---

func TestDatabase_MapUpdateContentRelationSortOrderParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	relationID := types.NewContentRelationID()

	input := UpdateContentRelationSortOrderParams{
		ContentRelationID: relationID,
		SortOrder:         10,
	}

	got := d.MapUpdateContentRelationSortOrderParams(input)

	if got.ContentRelationID != relationID {
		t.Errorf("ContentRelationID = %v, want %v", got.ContentRelationID, relationID)
	}
	if got.SortOrder != 10 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 10)
	}
}

func TestDatabase_MapUpdateContentRelationSortOrderParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateContentRelationSortOrderParams(UpdateContentRelationSortOrderParams{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.ContentRelationID != "" {
		t.Errorf("ContentRelationID = %v, want zero value", got.ContentRelationID)
	}
}

// --- MySQL MysqlDatabase.MapContentRelation tests ---

func TestMysqlDatabase_MapContentRelation_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	relationID := types.NewContentRelationID()
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()
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
			input := mdbm.ContentRelations{
				ContentRelationID: relationID,
				SourceContentID:   sourceID,
				TargetContentID:   targetID,
				FieldID:           fieldID,
				SortOrder:         tt.inputSortOrder,
				DateCreated:       ts,
			}

			got := d.MapContentRelation(input)

			if got.SortOrder != tt.wantSortOrder64 {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.wantSortOrder64)
			}
			if got.ContentRelationID != relationID {
				t.Errorf("ContentRelationID = %v, want %v", got.ContentRelationID, relationID)
			}
			if got.SourceContentID != sourceID {
				t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
			}
			if got.TargetContentID != targetID {
				t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
			}
			if got.FieldID != fieldID {
				t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
			}
		})
	}
}

func TestMysqlDatabase_MapContentRelation_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapContentRelation(mdbm.ContentRelations{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
}

// --- MySQL MysqlDatabase.MapCreateContentRelationParams tests ---

func TestMysqlDatabase_MapCreateContentRelationParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()

	input := CreateContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		FieldID:         fieldID,
		SortOrder:       7,
		DateCreated:     ts,
	}

	got := d.MapCreateContentRelationParams(input)

	if got.ContentRelationID.IsZero() {
		t.Fatal("expected non-zero ContentRelationID to be generated")
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	// MySQL uses int32 for SortOrder
	if got.SortOrder != int32(7) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(7))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestMysqlDatabase_MapCreateContentRelationParams_ZeroFieldID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreateContentRelationParams{
		FieldID: types.FieldID(""),
	}

	got := d.MapCreateContentRelationParams(input)

	if !got.FieldID.IsZero() {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
}

// --- MySQL MysqlDatabase.MapUpdateContentRelationSortOrderParams tests ---

func TestMysqlDatabase_MapUpdateContentRelationSortOrderParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	relationID := types.NewContentRelationID()

	input := UpdateContentRelationSortOrderParams{
		ContentRelationID: relationID,
		SortOrder:         15,
	}

	got := d.MapUpdateContentRelationSortOrderParams(input)

	if got.ContentRelationID != relationID {
		t.Errorf("ContentRelationID = %v, want %v", got.ContentRelationID, relationID)
	}
	// MySQL uses int32
	if got.SortOrder != int32(15) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(15))
	}
}

// --- PostgreSQL PsqlDatabase.MapContentRelation tests ---

func TestPsqlDatabase_MapContentRelation_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	relationID := types.NewContentRelationID()
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()
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
			input := mdbp.ContentRelations{
				ContentRelationID: relationID,
				SourceContentID:   sourceID,
				TargetContentID:   targetID,
				FieldID:           fieldID,
				SortOrder:         tt.inputSortOrder,
				DateCreated:       ts,
			}

			got := d.MapContentRelation(input)

			if got.SortOrder != tt.wantSortOrder64 {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.wantSortOrder64)
			}
			if got.ContentRelationID != relationID {
				t.Errorf("ContentRelationID = %v, want %v", got.ContentRelationID, relationID)
			}
			if got.SourceContentID != sourceID {
				t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
			}
			if got.TargetContentID != targetID {
				t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
			}
			if got.FieldID != fieldID {
				t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
			}
		})
	}
}

func TestPsqlDatabase_MapContentRelation_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapContentRelation(mdbp.ContentRelations{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateContentRelationParams tests ---

func TestPsqlDatabase_MapCreateContentRelationParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()

	input := CreateContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		FieldID:         fieldID,
		SortOrder:       12,
		DateCreated:     ts,
	}

	got := d.MapCreateContentRelationParams(input)

	if got.ContentRelationID.IsZero() {
		t.Fatal("expected non-zero ContentRelationID to be generated")
	}
	if got.SourceContentID != sourceID {
		t.Errorf("SourceContentID = %v, want %v", got.SourceContentID, sourceID)
	}
	if got.TargetContentID != targetID {
		t.Errorf("TargetContentID = %v, want %v", got.TargetContentID, targetID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	// PostgreSQL uses int32 for SortOrder
	if got.SortOrder != int32(12) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(12))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestPsqlDatabase_MapCreateContentRelationParams_ZeroFieldID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateContentRelationParams{
		FieldID: types.FieldID(""),
	}

	got := d.MapCreateContentRelationParams(input)

	if !got.FieldID.IsZero() {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateContentRelationSortOrderParams tests ---

func TestPsqlDatabase_MapUpdateContentRelationSortOrderParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	relationID := types.NewContentRelationID()

	input := UpdateContentRelationSortOrderParams{
		ContentRelationID: relationID,
		SortOrder:         20,
	}

	got := d.MapUpdateContentRelationSortOrderParams(input)

	if got.ContentRelationID != relationID {
		t.Errorf("ContentRelationID = %v, want %v", got.ContentRelationID, relationID)
	}
	// PostgreSQL uses int32
	if got.SortOrder != int32(20) {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int32(20))
	}
}

// --- Cross-database mapper consistency ---
// Verify that all three database mappers produce identical ContentRelations
// from equivalent input.

func TestCrossDatabaseMapContentRelation_Consistency(t *testing.T) {
	t.Parallel()
	relationID := types.NewContentRelationID()
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()
	ts := types.NewTimestamp(time.Date(2025, 8, 1, 9, 0, 0, 0, time.UTC))
	sqliteInput := mdb.ContentRelations{
		ContentRelationID: relationID,
		SourceContentID:   sourceID,
		TargetContentID:   targetID,
		FieldID:           fieldID,
		SortOrder:         7,
		DateCreated:       ts,
	}
	mysqlInput := mdbm.ContentRelations{
		ContentRelationID: relationID,
		SourceContentID:   sourceID,
		TargetContentID:   targetID,
		FieldID:           fieldID,
		SortOrder:         7,
		DateCreated:       ts,
	}
	psqlInput := mdbp.ContentRelations{
		ContentRelationID: relationID,
		SourceContentID:   sourceID,
		TargetContentID:   targetID,
		FieldID:           fieldID,
		SortOrder:         7,
		DateCreated:       ts,
	}

	sqliteResult := Database{}.MapContentRelation(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapContentRelation(mysqlInput)
	psqlResult := PsqlDatabase{}.MapContentRelation(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateContentRelationParams - ID generation ---

func TestCrossDatabaseMapCreateContentRelationParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	sourceID := types.NewContentID()
	targetID := types.NewContentID()

	input := CreateContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		SortOrder:       1,
		DateCreated:     ts,
	}

	sqliteResult := Database{}.MapCreateContentRelationParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateContentRelationParams(input)
	psqlResult := PsqlDatabase{}.MapCreateContentRelationParams(input)

	if sqliteResult.ContentRelationID.IsZero() {
		t.Error("SQLite: expected non-zero generated ContentRelationID")
	}
	if mysqlResult.ContentRelationID.IsZero() {
		t.Error("MySQL: expected non-zero generated ContentRelationID")
	}
	if psqlResult.ContentRelationID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated ContentRelationID")
	}

	// Each call should generate a unique ID
	if sqliteResult.ContentRelationID == mysqlResult.ContentRelationID {
		t.Error("SQLite and MySQL generated the same ContentRelationID -- each call should be unique")
	}
	if sqliteResult.ContentRelationID == psqlResult.ContentRelationID {
		t.Error("SQLite and PostgreSQL generated the same ContentRelationID -- each call should be unique")
	}
}

// --- Cross-database FieldID passthrough consistency ---

func TestCrossDatabaseMapCreateContentRelationParams_FieldIDPassthrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fieldID types.FieldID
	}{
		{"non-zero field ID", types.NewFieldID()},
		{"zero field ID", types.FieldID("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := CreateContentRelationParams{
				FieldID: tt.fieldID,
			}

			sqliteGot := Database{}.MapCreateContentRelationParams(input)
			mysqlGot := MysqlDatabase{}.MapCreateContentRelationParams(input)
			psqlGot := PsqlDatabase{}.MapCreateContentRelationParams(input)

			if sqliteGot.FieldID != tt.fieldID {
				t.Errorf("SQLite FieldID = %v, want %v", sqliteGot.FieldID, tt.fieldID)
			}
			if mysqlGot.FieldID != tt.fieldID {
				t.Errorf("MySQL FieldID = %v, want %v", mysqlGot.FieldID, tt.fieldID)
			}
			if psqlGot.FieldID != tt.fieldID {
				t.Errorf("PostgreSQL FieldID = %v, want %v", psqlGot.FieldID, tt.fieldID)
			}
		})
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewContentRelationCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-cr-001",
		IP:        "10.0.0.1",
	}
	sourceID := types.NewContentID()
	targetID := types.NewContentID()
	fieldID := types.NewFieldID()
	params := CreateContentRelationParams{
		SourceContentID: sourceID,
		TargetContentID: targetID,
		FieldID:         fieldID,
		SortOrder:       5,
		DateCreated:     ts,
	}

	cmd := Database{}.NewContentRelationCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
	}
	p, ok := cmd.Params().(CreateContentRelationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentRelationParams", cmd.Params())
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

func TestNewContentRelationCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	relationID := types.NewContentRelationID()
	cmd := NewContentRelationCmd{}

	row := mdb.ContentRelations{ContentRelationID: relationID}
	got := cmd.GetID(row)
	if got != string(relationID) {
		t.Errorf("GetID() = %q, want %q", got, string(relationID))
	}
}

func TestUpdateContentRelationSortOrderCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewContentRelationID()
	params := UpdateContentRelationSortOrderParams{
		ContentRelationID: relationID,
		SortOrder:         99,
	}

	cmd := Database{}.UpdateContentRelationSortOrderCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
	}
	if cmd.GetID() != string(relationID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(relationID))
	}
	p, ok := cmd.Params().(UpdateContentRelationSortOrderParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentRelationSortOrderParams", cmd.Params())
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

func TestDeleteContentRelationCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewContentRelationID()

	cmd := Database{}.DeleteContentRelationCmd(ctx, ac, relationID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
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

func TestNewContentRelationCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-cr-001",
		IP:        "192.168.1.1",
	}
	params := CreateContentRelationParams{
		SourceContentID: types.NewContentID(),
		TargetContentID: types.NewContentID(),
		FieldID:         types.NewFieldID(),
		SortOrder:       3,
		DateCreated:     ts,
	}

	cmd := MysqlDatabase{}.NewContentRelationCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
	}
	p, ok := cmd.Params().(CreateContentRelationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentRelationParams", cmd.Params())
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

func TestNewContentRelationCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	relationID := types.NewContentRelationID()
	cmd := NewContentRelationCmdMysql{}

	row := mdbm.ContentRelations{ContentRelationID: relationID}
	got := cmd.GetID(row)
	if got != string(relationID) {
		t.Errorf("GetID() = %q, want %q", got, string(relationID))
	}
}

func TestUpdateContentRelationSortOrderCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewContentRelationID()
	params := UpdateContentRelationSortOrderParams{
		ContentRelationID: relationID,
		SortOrder:         55,
	}

	cmd := MysqlDatabase{}.UpdateContentRelationSortOrderCmd(ctx, ac, params)

	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
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
	p, ok := cmd.Params().(UpdateContentRelationSortOrderParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentRelationSortOrderParams", cmd.Params())
	}
	if p.SortOrder != 55 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 55)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentRelationCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewContentRelationID()

	cmd := MysqlDatabase{}.DeleteContentRelationCmd(ctx, ac, relationID)

	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
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

func TestNewContentRelationCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-cr-001",
		IP:        "172.16.0.1",
	}
	params := CreateContentRelationParams{
		SourceContentID: types.NewContentID(),
		TargetContentID: types.NewContentID(),
		FieldID:         types.NewFieldID(),
		SortOrder:       8,
		DateCreated:     ts,
	}

	cmd := PsqlDatabase{}.NewContentRelationCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
	}
	p, ok := cmd.Params().(CreateContentRelationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentRelationParams", cmd.Params())
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

func TestNewContentRelationCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	relationID := types.NewContentRelationID()
	cmd := NewContentRelationCmdPsql{}

	row := mdbp.ContentRelations{ContentRelationID: relationID}
	got := cmd.GetID(row)
	if got != string(relationID) {
		t.Errorf("GetID() = %q, want %q", got, string(relationID))
	}
}

func TestUpdateContentRelationSortOrderCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewContentRelationID()
	params := UpdateContentRelationSortOrderParams{
		ContentRelationID: relationID,
		SortOrder:         77,
	}

	cmd := PsqlDatabase{}.UpdateContentRelationSortOrderCmd(ctx, ac, params)

	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
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
	p, ok := cmd.Params().(UpdateContentRelationSortOrderParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentRelationSortOrderParams", cmd.Params())
	}
	if p.SortOrder != 77 {
		t.Errorf("Params().SortOrder = %d, want %d", p.SortOrder, 77)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentRelationCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	relationID := types.NewContentRelationID()

	cmd := PsqlDatabase{}.DeleteContentRelationCmd(ctx, ac, relationID)

	if cmd.TableName() != "content_relations" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_relations")
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

func TestAuditedContentRelationCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateContentRelationParams{}
	updateParams := UpdateContentRelationSortOrderParams{
		ContentRelationID: types.NewContentRelationID(),
	}
	relationID := types.NewContentRelationID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewContentRelationCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateContentRelationSortOrderCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteContentRelationCmd(ctx, ac, relationID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewContentRelationCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateContentRelationSortOrderCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteContentRelationCmd(ctx, ac, relationID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewContentRelationCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateContentRelationSortOrderCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteContentRelationCmd(ctx, ac, relationID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "content_relations" {
				t.Errorf("TableName() = %q, want %q", c.name, "content_relations")
			}
		})
	}
}

func TestAuditedContentRelationCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateContentRelationParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewContentRelationCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewContentRelationCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewContentRelationCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedContentRelationCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	relationID := types.NewContentRelationID()

	t.Run("UpdateCmd GetID returns relation ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateContentRelationSortOrderParams{
			ContentRelationID: relationID,
			SortOrder:         1,
		}

		sqliteCmd := Database{}.UpdateContentRelationSortOrderCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateContentRelationSortOrderCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateContentRelationSortOrderCmd(ctx, ac, params)

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
		sqliteCmd := Database{}.DeleteContentRelationCmd(ctx, ac, relationID)
		mysqlCmd := MysqlDatabase{}.DeleteContentRelationCmd(ctx, ac, relationID)
		psqlCmd := PsqlDatabase{}.DeleteContentRelationCmd(ctx, ac, relationID)

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
		testRelationID := types.NewContentRelationID()
		wantID := string(testRelationID)

		sqliteCmd := NewContentRelationCmd{}
		mysqlCmd := NewContentRelationCmdMysql{}
		psqlCmd := NewContentRelationCmdPsql{}

		sqliteRow := mdb.ContentRelations{ContentRelationID: testRelationID}
		mysqlRow := mdbm.ContentRelations{ContentRelationID: testRelationID}
		psqlRow := mdbp.ContentRelations{ContentRelationID: testRelationID}

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

func TestUpdateContentRelationSortOrderCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateContentRelationSortOrderParams{
		ContentRelationID: "",
	}

	sqliteCmd := Database{}.UpdateContentRelationSortOrderCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateContentRelationSortOrderCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateContentRelationSortOrderCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteContentRelationCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.ContentRelationID("")

	sqliteCmd := Database{}.DeleteContentRelationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteContentRelationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteContentRelationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.ContentRelations]  = NewContentRelationCmd{}
	_ audited.UpdateCommand[mdb.ContentRelations]  = UpdateContentRelationSortOrderCmd{}
	_ audited.DeleteCommand[mdb.ContentRelations]  = DeleteContentRelationCmd{}
	_ audited.CreateCommand[mdbm.ContentRelations] = NewContentRelationCmdMysql{}
	_ audited.UpdateCommand[mdbm.ContentRelations] = UpdateContentRelationSortOrderCmdMysql{}
	_ audited.DeleteCommand[mdbm.ContentRelations] = DeleteContentRelationCmdMysql{}
	_ audited.CreateCommand[mdbp.ContentRelations] = NewContentRelationCmdPsql{}
	_ audited.UpdateCommand[mdbp.ContentRelations] = UpdateContentRelationSortOrderCmdPsql{}
	_ audited.DeleteCommand[mdbp.ContentRelations] = DeleteContentRelationCmdPsql{}
)
