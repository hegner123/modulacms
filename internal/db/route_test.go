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

// routeTestFixture returns a fully populated Routes struct for testing.
func routeTestFixture() (Routes, types.RouteID, types.NullableUserID, types.Timestamp) {
	routeID := types.NewRouteID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	route := Routes{
		RouteID:      routeID,
		Slug:         types.Slug("test-route"),
		Title:        "Test Route",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}
	return route, routeID, authorID, ts
}

// --- MapStringRoute tests ---

func TestMapStringRoute_AllFields(t *testing.T) {
	t.Parallel()
	route, routeID, authorID, ts := routeTestFixture()

	got := MapStringRoute(route)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"RouteID", got.RouteID, routeID.String()},
		{"Slug", got.Slug, "test-route"},
		{"Title", got.Title, "Test Route"},
		{"Status", got.Status, "1"},
		{"AuthorID", got.AuthorID, authorID.String()},
		{"DateCreated", got.DateCreated, ts.String()},
		{"DateModified", got.DateModified, ts.String()},
		{"History", got.History, ""},
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

func TestMapStringRoute_ZeroValueRoute(t *testing.T) {
	t.Parallel()
	// Zero-value Routes: all fields should convert without panic
	got := MapStringRoute(Routes{})

	if got.RouteID != "" {
		t.Errorf("RouteID = %q, want empty string", got.RouteID)
	}
	if got.Slug != "" {
		t.Errorf("Slug = %q, want empty string", got.Slug)
	}
	if got.Title != "" {
		t.Errorf("Title = %q, want empty string", got.Title)
	}
	if got.Status != "0" {
		t.Errorf("Status = %q, want %q", got.Status, "0")
	}
	if got.History != "" {
		t.Errorf("History = %q, want empty string", got.History)
	}
}

func TestMapStringRoute_NegativeStatus(t *testing.T) {
	t.Parallel()
	// Negative status values should be formatted correctly
	route := Routes{Status: -1}
	got := MapStringRoute(route)
	if got.Status != "-1" {
		t.Errorf("Status = %q, want %q", got.Status, "-1")
	}
}

func TestMapStringRoute_LargeStatus(t *testing.T) {
	t.Parallel()
	// Ensure large int64 values format correctly
	route := Routes{Status: math.MaxInt64}
	got := MapStringRoute(route)
	want := fmt.Sprintf("%d", int64(math.MaxInt64))
	if got.Status != want {
		t.Errorf("Status = %q, want %q", got.Status, want)
	}
}

func TestMapStringRoute_NullAuthorID(t *testing.T) {
	t.Parallel()
	// AuthorID with Valid=false should still convert without panic
	route := Routes{
		AuthorID: types.NullableUserID{Valid: false},
	}
	got := MapStringRoute(route)
	// Should produce whatever NullableUserID.String() returns for invalid
	if got.AuthorID != route.AuthorID.String() {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, route.AuthorID.String())
	}
}

// --- SQLite Database.MapRoute tests ---

func TestDatabase_MapRoute_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	routeID := types.NewRouteID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	input := mdb.Routes{
		RouteID:      routeID,
		Slug:         types.Slug("sqlite-route"),
		Title:        "SQLite Route",
		Status:       42,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapRoute(input)

	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.Slug != types.Slug("sqlite-route") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("sqlite-route"))
	}
	if got.Title != "SQLite Route" {
		t.Errorf("Title = %q, want %q", got.Title, "SQLite Route")
	}
	if got.Status != 42 {
		t.Errorf("Status = %d, want %d", got.Status, 42)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapRoute_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapRoute(mdb.Routes{})

	if got.RouteID != "" {
		t.Errorf("RouteID = %v, want zero value", got.RouteID)
	}
	if got.Status != 0 {
		t.Errorf("Status = %d, want 0", got.Status)
	}
}

// --- SQLite Database.MapCreateRouteParams tests ---

func TestDatabase_MapCreateRouteParams_Table(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	tests := []struct {
		name  string
		input CreateRouteParams
	}{
		{
			name: "always generates a new RouteID",
			input: CreateRouteParams{
				Slug:         types.Slug("auto-slug"),
				Title:        "Auto Route",
				Status:       1,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
		},
		{
			name: "maps all fields correctly",
			input: CreateRouteParams{
				Slug:         types.Slug("explicit-slug"),
				Title:        "Explicit Route",
				Status:       0,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.MapCreateRouteParams(tt.input)

			if got.RouteID.IsZero() {
				t.Fatal("expected non-zero RouteID to be generated")
			}

			if got.Slug != tt.input.Slug {
				t.Errorf("Slug = %v, want %v", got.Slug, tt.input.Slug)
			}
			if got.Title != tt.input.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.input.Title)
			}
			if got.Status != tt.input.Status {
				t.Errorf("Status = %d, want %d", got.Status, tt.input.Status)
			}
			if got.AuthorID != tt.input.AuthorID {
				t.Errorf("AuthorID = %v, want %v", got.AuthorID, tt.input.AuthorID)
			}
			if got.DateCreated != tt.input.DateCreated {
				t.Errorf("DateCreated = %v, want %v", got.DateCreated, tt.input.DateCreated)
			}
			if got.DateModified != tt.input.DateModified {
				t.Errorf("DateModified = %v, want %v", got.DateModified, tt.input.DateModified)
			}
		})
	}
}

// --- SQLite Database.MapUpdateRouteParams tests ---

func TestDatabase_MapUpdateRouteParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateRouteParams{
		Slug:         types.Slug("new-slug"),
		Title:        "New Title",
		Status:       3,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("old-slug"),
	}

	got := d.MapUpdateRouteParams(input)

	if got.Slug != types.Slug("new-slug") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("new-slug"))
	}
	if got.Title != "New Title" {
		t.Errorf("Title = %q, want %q", got.Title, "New Title")
	}
	if got.Status != 3 {
		t.Errorf("Status = %d, want %d", got.Status, 3)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.Slug_2 != types.Slug("old-slug") {
		t.Errorf("Slug_2 = %v, want %v", got.Slug_2, types.Slug("old-slug"))
	}
}

// --- MySQL MysqlDatabase.MapRoute tests ---

func TestMysqlDatabase_MapRoute_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	routeID := types.NewRouteID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))

	tests := []struct {
		name         string
		inputStatus  int32
		wantStatus64 int64
	}{
		{"positive status", 5, 5},
		{"zero status", 0, 0},
		{"negative status", -1, -1},
		{"max int32", math.MaxInt32, int64(math.MaxInt32)},
		{"min int32", math.MinInt32, int64(math.MinInt32)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbm.Routes{
				RouteID:      routeID,
				Slug:         types.Slug("mysql-route"),
				Title:        "MySQL Route",
				Status:       tt.inputStatus,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			}

			got := d.MapRoute(input)

			if got.Status != tt.wantStatus64 {
				t.Errorf("Status = %d, want %d", got.Status, tt.wantStatus64)
			}
			if got.RouteID != routeID {
				t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
			}
			if got.Slug != types.Slug("mysql-route") {
				t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("mysql-route"))
			}
			if got.Title != "MySQL Route" {
				t.Errorf("Title = %q, want %q", got.Title, "MySQL Route")
			}
		})
	}
}

func TestMysqlDatabase_MapRoute_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapRoute(mdbm.Routes{})

	if got.Status != 0 {
		t.Errorf("Status = %d, want 0", got.Status)
	}
}

// --- MySQL MysqlDatabase.MapCreateRouteParams tests ---

func TestMysqlDatabase_MapCreateRouteParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	tests := []struct {
		name  string
		input CreateRouteParams
	}{
		{
			name: "always generates a new RouteID",
			input: CreateRouteParams{
				Slug:         types.Slug("mysql-auto"),
				Title:        "MySQL Auto",
				Status:       10,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
		},
		{
			name: "maps all fields correctly",
			input: CreateRouteParams{
				Slug:         types.Slug("mysql-explicit"),
				Title:        "MySQL Explicit",
				Status:       20,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.MapCreateRouteParams(tt.input)

			if got.RouteID.IsZero() {
				t.Fatal("expected non-zero RouteID to be generated")
			}

			// MySQL uses int32 for Status
			if got.Status != int32(tt.input.Status) {
				t.Errorf("Status = %d, want %d", got.Status, int32(tt.input.Status))
			}
			if got.Slug != tt.input.Slug {
				t.Errorf("Slug = %v, want %v", got.Slug, tt.input.Slug)
			}
			if got.Title != tt.input.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.input.Title)
			}
		})
	}
}

// --- MySQL MysqlDatabase.MapUpdateRouteParams tests ---

func TestMysqlDatabase_MapUpdateRouteParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateRouteParams{
		Slug:         types.Slug("mysql-updated"),
		Title:        "MySQL Updated",
		Status:       7,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("mysql-original"),
	}

	got := d.MapUpdateRouteParams(input)

	if got.Slug != types.Slug("mysql-updated") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("mysql-updated"))
	}
	if got.Title != "MySQL Updated" {
		t.Errorf("Title = %q, want %q", got.Title, "MySQL Updated")
	}
	// MySQL uses int32
	if got.Status != int32(7) {
		t.Errorf("Status = %d, want %d", got.Status, int32(7))
	}
	if got.Slug_2 != types.Slug("mysql-original") {
		t.Errorf("Slug_2 = %v, want %v", got.Slug_2, types.Slug("mysql-original"))
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

// --- PostgreSQL PsqlDatabase.MapRoute tests ---

func TestPsqlDatabase_MapRoute_Int32ToInt64Conversion(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	routeID := types.NewRouteID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))

	tests := []struct {
		name         string
		inputStatus  int32
		wantStatus64 int64
	}{
		{"positive status", 3, 3},
		{"zero status", 0, 0},
		{"negative status", -100, -100},
		{"max int32", math.MaxInt32, int64(math.MaxInt32)},
		{"min int32", math.MinInt32, int64(math.MinInt32)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbp.Routes{
				RouteID:      routeID,
				Slug:         types.Slug("psql-route"),
				Title:        "Psql Route",
				Status:       tt.inputStatus,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			}

			got := d.MapRoute(input)

			if got.Status != tt.wantStatus64 {
				t.Errorf("Status = %d, want %d", got.Status, tt.wantStatus64)
			}
			if got.RouteID != routeID {
				t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
			}
			if got.Title != "Psql Route" {
				t.Errorf("Title = %q, want %q", got.Title, "Psql Route")
			}
		})
	}
}

func TestPsqlDatabase_MapRoute_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapRoute(mdbp.Routes{})

	if got.Status != 0 {
		t.Errorf("Status = %d, want 0", got.Status)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateRouteParams tests ---

func TestPsqlDatabase_MapCreateRouteParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	tests := []struct {
		name  string
		input CreateRouteParams
	}{
		{
			name: "always generates a new RouteID",
			input: CreateRouteParams{
				Slug:         types.Slug("psql-auto"),
				Title:        "Psql Auto",
				Status:       10,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
		},
		{
			name: "maps all fields correctly",
			input: CreateRouteParams{
				Slug:         types.Slug("psql-explicit"),
				Title:        "Psql Explicit",
				Status:       20,
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.MapCreateRouteParams(tt.input)

			if got.RouteID.IsZero() {
				t.Fatal("expected non-zero RouteID to be generated")
			}

			// PostgreSQL uses int32 for Status
			if got.Status != int32(tt.input.Status) {
				t.Errorf("Status = %d, want %d", got.Status, int32(tt.input.Status))
			}
			if got.Slug != tt.input.Slug {
				t.Errorf("Slug = %v, want %v", got.Slug, tt.input.Slug)
			}
			if got.Title != tt.input.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.input.Title)
			}
		})
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateRouteParams tests ---

func TestPsqlDatabase_MapUpdateRouteParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateRouteParams{
		Slug:         types.Slug("psql-updated"),
		Title:        "Psql Updated",
		Status:       9,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("psql-original"),
	}

	got := d.MapUpdateRouteParams(input)

	if got.Slug != types.Slug("psql-updated") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("psql-updated"))
	}
	if got.Title != "Psql Updated" {
		t.Errorf("Title = %q, want %q", got.Title, "Psql Updated")
	}
	// PostgreSQL uses int32
	if got.Status != int32(9) {
		t.Errorf("Status = %d, want %d", got.Status, int32(9))
	}
	if got.Slug_2 != types.Slug("psql-original") {
		t.Errorf("Slug_2 = %v, want %v", got.Slug_2, types.Slug("psql-original"))
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

// --- Cross-database mapper consistency ---
// This test verifies that all three database mappers produce identical Routes
// from equivalent input, proving the abstraction layer works correctly.

func TestCrossDatabaseMapRoute_Consistency(t *testing.T) {
	t.Parallel()
	routeID := types.NewRouteID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))

	sqliteInput := mdb.Routes{
		RouteID: routeID, Slug: types.Slug("cross-db"),
		Title: "Cross DB", Status: 2,
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.Routes{
		RouteID: routeID, Slug: types.Slug("cross-db"),
		Title: "Cross DB", Status: 2,
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.Routes{
		RouteID: routeID, Slug: types.Slug("cross-db"),
		Title: "Cross DB", Status: 2,
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapRoute(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapRoute(mysqlInput)
	psqlResult := PsqlDatabase{}.MapRoute(psqlInput)

	// All three should produce identical wrapper Routes
	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateRouteParams consistency ---
// Verify all three auto-generate an ID when none is provided.

func TestCrossDatabaseMapCreateRouteParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateRouteParams{
		Slug:         types.Slug("cross-create"),
		Title:        "Cross Create",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateRouteParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateRouteParams(input)
	psqlResult := PsqlDatabase{}.MapCreateRouteParams(input)

	if sqliteResult.RouteID.IsZero() {
		t.Error("SQLite: expected non-zero generated RouteID")
	}
	if mysqlResult.RouteID.IsZero() {
		t.Error("MySQL: expected non-zero generated RouteID")
	}
	if psqlResult.RouteID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated RouteID")
	}

	// Each call should generate a unique ID
	if sqliteResult.RouteID == mysqlResult.RouteID {
		t.Error("SQLite and MySQL generated the same RouteID -- each call should be unique")
	}
	if sqliteResult.RouteID == psqlResult.RouteID {
		t.Error("SQLite and PostgreSQL generated the same RouteID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewRouteCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-123",
		IP:        "10.0.0.1",
	}
	params := CreateRouteParams{
		Slug:         types.Slug("cmd-test"),
		Title:        "Cmd Test",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	p, ok := cmd.Params().(CreateRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateRouteParams", cmd.Params())
	}
	if p.Title != "Cmd Test" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "Cmd Test")
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewRouteCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	routeID := types.NewRouteID()
	cmd := NewRouteCmd{}

	row := mdb.Routes{RouteID: routeID}
	got := cmd.GetID(row)
	if got != string(routeID) {
		t.Errorf("GetID() = %q, want %q", got, string(routeID))
	}
}

func TestNewRouteCmd_ExecuteAutoIDGeneration(t *testing.T) {
	t.Parallel()
	// Verify that the cmd stores params correctly through the constructor.
	// ID generation always happens inside Execute, not in params.
	params := CreateRouteParams{
		Slug: types.Slug("exec-auto"),
	}
	cmd := Database{}.NewRouteCmd(context.Background(), audited.AuditContext{}, params)

	p, ok := cmd.Params().(CreateRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateRouteParams", cmd.Params())
	}
	if p.Slug != types.Slug("exec-auto") {
		t.Errorf("Params().Slug = %q, want %q", p.Slug, "exec-auto")
	}
}

func TestUpdateRouteCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateRouteParams{
		Slug:         types.Slug("updated"),
		Title:        "Updated",
		Status:       2,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("original"),
	}

	cmd := Database{}.UpdateRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	// GetID returns the Slug_2 value (the identifier for the WHERE clause)
	if cmd.GetID() != "original" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "original")
	}
	p, ok := cmd.Params().(UpdateRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateRouteParams", cmd.Params())
	}
	if p.Title != "Updated" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "Updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteRouteCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	routeID := types.NewRouteID()

	cmd := Database{}.DeleteRouteCmd(ctx, ac, routeID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.GetID() != string(routeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(routeID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewRouteCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateRouteParams{
		Slug:         types.Slug("mysql-cmd"),
		Title:        "MySQL Cmd",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	p, ok := cmd.Params().(CreateRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateRouteParams", cmd.Params())
	}
	if p.Title != "MySQL Cmd" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "MySQL Cmd")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewRouteCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	routeID := types.NewRouteID()
	cmd := NewRouteCmdMysql{}

	row := mdbm.Routes{RouteID: routeID}
	got := cmd.GetID(row)
	if got != string(routeID) {
		t.Errorf("GetID() = %q, want %q", got, string(routeID))
	}
}

func TestUpdateRouteCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateRouteParams{
		Slug:         types.Slug("mysql-updated"),
		Title:        "MySQL Updated",
		Status:       4,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("mysql-original"),
	}

	cmd := MysqlDatabase{}.UpdateRouteCmd(ctx, ac, params)

	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.GetID() != "mysql-original" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "mysql-original")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateRouteParams", cmd.Params())
	}
	if p.Status != 4 {
		t.Errorf("Params().Status = %d, want %d", p.Status, 4)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteRouteCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	routeID := types.NewRouteID()

	cmd := MysqlDatabase{}.DeleteRouteCmd(ctx, ac, routeID)

	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.GetID() != string(routeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(routeID))
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

func TestNewRouteCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateRouteParams{
		Slug:         types.Slug("psql-cmd"),
		Title:        "Psql Cmd",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	p, ok := cmd.Params().(CreateRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateRouteParams", cmd.Params())
	}
	if p.Title != "Psql Cmd" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "Psql Cmd")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewRouteCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	routeID := types.NewRouteID()
	cmd := NewRouteCmdPsql{}

	row := mdbp.Routes{RouteID: routeID}
	got := cmd.GetID(row)
	if got != string(routeID) {
		t.Errorf("GetID() = %q, want %q", got, string(routeID))
	}
}

func TestUpdateRouteCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateRouteParams{
		Slug:         types.Slug("psql-updated"),
		Title:        "Psql Updated",
		Status:       5,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("psql-original"),
	}

	cmd := PsqlDatabase{}.UpdateRouteCmd(ctx, ac, params)

	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.GetID() != "psql-original" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "psql-original")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateRouteParams", cmd.Params())
	}
	if p.Status != 5 {
		t.Errorf("Params().Status = %d, want %d", p.Status, 5)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteRouteCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	routeID := types.NewRouteID()

	cmd := PsqlDatabase{}.DeleteRouteCmd(ctx, ac, routeID)

	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.GetID() != string(routeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(routeID))
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
// Verify that all three database types produce commands with the correct table
// name and the correct recorder type.

func TestAuditedRouteCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateRouteParams{}
	updateParams := UpdateRouteParams{Slug_2: types.Slug("target")}
	routeID := types.NewRouteID()

	// SQLite
	sqliteCreate := Database{}.NewRouteCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateRouteCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteRouteCmd(ctx, ac, routeID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewRouteCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateRouteCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteRouteCmd(ctx, ac, routeID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewRouteCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateRouteCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteRouteCmd(ctx, ac, routeID)

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", sqliteCreate.TableName()},
		{"SQLite Update", sqliteUpdate.TableName()},
		{"SQLite Delete", sqliteDelete.TableName()},
		{"MySQL Create", mysqlCreate.TableName()},
		{"MySQL Update", mysqlUpdate.TableName()},
		{"MySQL Delete", mysqlDelete.TableName()},
		{"PostgreSQL Create", psqlCreate.TableName()},
		{"PostgreSQL Update", psqlUpdate.TableName()},
		{"PostgreSQL Delete", psqlDelete.TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "routes" {
				t.Errorf("TableName() = %q, want %q", c.name, "routes")
			}
		})
	}
}

func TestAuditedRouteCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateRouteParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewRouteCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewRouteCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewRouteCmd(ctx, ac, createParams).Recorder()},
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
// Verify GetID returns the correct identifier across all database variants.

func TestAuditedRouteCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	routeID := types.NewRouteID()
	updateSlug := types.Slug("target-slug")

	t.Run("UpdateCmd GetID returns Slug_2", func(t *testing.T) {
		t.Parallel()
		params := UpdateRouteParams{Slug_2: updateSlug}

		sqliteCmd := Database{}.UpdateRouteCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateRouteCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateRouteCmd(ctx, ac, params)

		wantID := string(updateSlug)
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

	t.Run("DeleteCmd GetID returns route ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteRouteCmd(ctx, ac, routeID)
		mysqlCmd := MysqlDatabase{}.DeleteRouteCmd(ctx, ac, routeID)
		psqlCmd := PsqlDatabase{}.DeleteRouteCmd(ctx, ac, routeID)

		wantID := string(routeID)
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
		testRouteID := types.NewRouteID()

		sqliteCmd := NewRouteCmd{}
		mysqlCmd := NewRouteCmdMysql{}
		psqlCmd := NewRouteCmdPsql{}

		wantID := string(testRouteID)

		sqliteRow := mdb.Routes{RouteID: testRouteID}
		mysqlRow := mdbm.Routes{RouteID: testRouteID}
		psqlRow := mdbp.Routes{RouteID: testRouteID}

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

// --- Edge case: UpdateRouteCmd with empty Slug_2 ---

func TestUpdateRouteCmd_GetID_EmptySlug(t *testing.T) {
	t.Parallel()
	// When Slug_2 is empty, GetID should return empty string
	params := UpdateRouteParams{Slug_2: ""}

	sqliteCmd := Database{}.UpdateRouteCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateRouteCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateRouteCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteRouteCmd with empty ID ---

func TestDeleteRouteCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.RouteID("")

	sqliteCmd := Database{}.DeleteRouteCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteRouteCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteRouteCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.Routes]  = NewRouteCmd{}
	_ audited.UpdateCommand[mdb.Routes]  = UpdateRouteCmd{}
	_ audited.DeleteCommand[mdb.Routes]  = DeleteRouteCmd{}
	_ audited.CreateCommand[mdbm.Routes] = NewRouteCmdMysql{}
	_ audited.UpdateCommand[mdbm.Routes] = UpdateRouteCmdMysql{}
	_ audited.DeleteCommand[mdbm.Routes] = DeleteRouteCmdMysql{}
	_ audited.CreateCommand[mdbp.Routes] = NewRouteCmdPsql{}
	_ audited.UpdateCommand[mdbp.Routes] = UpdateRouteCmdPsql{}
	_ audited.DeleteCommand[mdbp.Routes] = DeleteRouteCmdPsql{}
)
