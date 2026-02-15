package db

import (
	"context"
	"encoding/json"
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

// adminRouteFixture returns a fully populated AdminRoutes and its parts for testing.
func adminRouteFixture() (AdminRoutes, types.AdminRouteID, types.Slug, types.Timestamp) {
	routeID := types.NewAdminRouteID()
	slug := types.Slug("/test-admin-route")
	ts := types.NewTimestamp(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	ar := AdminRoutes{
		AdminRouteID: routeID,
		Slug:         slug,
		Title:        "Test Admin Route",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}
	return ar, routeID, slug, ts
}

// adminRouteFixtureNulls returns an AdminRoutes where all nullable fields are null/invalid.
func adminRouteFixtureNulls() AdminRoutes {
	routeID := types.NewAdminRouteID()
	ts := types.NewTimestamp(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	return AdminRoutes{
		AdminRouteID: routeID,
		Slug:         types.Slug("/null-route"),
		Title:        "Minimal",
		Status:       0,
		AuthorID:     types.NullableUserID{Valid: false},
		DateCreated:  ts,
		DateModified: ts,
	}
}

// --- MapStringAdminRoute tests ---

func TestMapStringAdminRoute_AllFields(t *testing.T) {
	t.Parallel()
	ar, routeID, slug, ts := adminRouteFixture()

	got := MapStringAdminRoute(ar)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AdminRouteID", got.AdminRouteID, routeID.String()},
		{"Slug", got.Slug, string(slug)},
		{"Title", got.Title, "Test Admin Route"},
		{"Status", got.Status, fmt.Sprintf("%d", ar.Status)},
		{"AuthorID", got.AuthorID, ar.AuthorID.String()},
		{"DateCreated", got.DateCreated, ts.String()},
		{"DateModified", got.DateModified, ts.String()},
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

func TestMapStringAdminRoute_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// History field is hard-coded to "" per the comment in the source.
	ar, _, _, _ := adminRouteFixture()

	got := MapStringAdminRoute(ar)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (History field removed)", got.History)
	}
}

func TestMapStringAdminRoute_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringAdminRoute(AdminRoutes{})

	if got.AdminRouteID != "" {
		t.Errorf("AdminRouteID = %q, want empty string", got.AdminRouteID)
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
}

func TestMapStringAdminRoute_NullableFieldsShowNull(t *testing.T) {
	t.Parallel()
	ar := adminRouteFixtureNulls()

	got := MapStringAdminRoute(ar)

	// Nullable UserID with Valid=false produces "null" via its String() method
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null", got.AuthorID, "null")
	}
}

func TestMapStringAdminRoute_StatusFormatting(t *testing.T) {
	t.Parallel()
	// Verify different status values produce the correct string representation.
	tests := []struct {
		name   string
		status int64
		want   string
	}{
		{"zero", 0, "0"},
		{"positive", 42, "42"},
		{"negative", -1, "-1"},
		{"large", 9999999, "9999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ar := AdminRoutes{Status: tt.status}
			got := MapStringAdminRoute(ar)
			if got.Status != tt.want {
				t.Errorf("Status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

func TestMapStringAdminRoute_SpecialCharacters(t *testing.T) {
	t.Parallel()
	// Titles with special characters should pass through unchanged.
	tests := []struct {
		name  string
		title string
		slug  types.Slug
	}{
		{"unicode", "Seite \u00e4\u00f6\u00fc", types.Slug("seite-aou")},
		{"spaces", "My Admin Route", types.Slug("my-admin-route")},
		{"empty_strings", "", types.Slug("")},
		{"special_chars", "Route <script>", types.Slug("route-script")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ar := AdminRoutes{
				AdminRouteID: types.NewAdminRouteID(),
				Title:        tt.title,
				Slug:         tt.slug,
			}
			got := MapStringAdminRoute(ar)
			if got.Title != tt.title {
				t.Errorf("Title = %q, want %q", got.Title, tt.title)
			}
			if got.Slug != string(tt.slug) {
				t.Errorf("Slug = %q, want %q", got.Slug, string(tt.slug))
			}
		})
	}
}

func TestMapStringAdminRoute_TimestampFormatting(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ts   time.Time
	}{
		{"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"recent", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"midnight", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := types.NewTimestamp(tt.ts)
			ar := AdminRoutes{
				AdminRouteID: types.NewAdminRouteID(),
				Slug:         types.Slug("ts-test"),
				DateCreated:  ts,
				DateModified: ts,
			}
			got := MapStringAdminRoute(ar)
			if got.DateCreated != ts.String() {
				t.Errorf("DateCreated = %q, want %q", got.DateCreated, ts.String())
			}
			if got.DateModified != ts.String() {
				t.Errorf("DateModified = %q, want %q", got.DateModified, ts.String())
			}
		})
	}
}

// --- SQLite Database.MapAdminRoute tests ---

func TestDatabase_MapAdminRoute_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	routeID := types.NewAdminRouteID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdb.AdminRoutes{
		AdminRouteID: routeID,
		Slug:         types.Slug("sqlite-route"),
		Title:        "SQLite Route",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapAdminRoute(input)

	if got.AdminRouteID != routeID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, routeID)
	}
	if got.Slug != types.Slug("sqlite-route") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("sqlite-route"))
	}
	if got.Title != "SQLite Route" {
		t.Errorf("Title = %q, want %q", got.Title, "SQLite Route")
	}
	if got.Status != 1 {
		t.Errorf("Status = %d, want %d", got.Status, 1)
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

func TestDatabase_MapAdminRoute_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminRoute(mdb.AdminRoutes{})

	if got.AdminRouteID != "" {
		t.Errorf("AdminRouteID = %v, want zero value", got.AdminRouteID)
	}
	if got.Slug != "" {
		t.Errorf("Slug = %v, want zero value", got.Slug)
	}
	if got.Title != "" {
		t.Errorf("Title = %q, want empty string", got.Title)
	}
	if got.Status != 0 {
		t.Errorf("Status = %d, want 0", got.Status)
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false for zero value")
	}
}

// --- SQLite Database.MapCreateAdminRouteParams tests ---

func TestDatabase_MapCreateAdminRouteParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminRouteParams{
		Slug:         types.Slug("new-route"),
		Title:        "New Route",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: types.NewUserID(), Valid: true},
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminRouteParams(input)

	if got.AdminRouteID.IsZero() {
		t.Fatal("expected non-zero AdminRouteID to be generated")
	}
	if got.Slug != types.Slug("new-route") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("new-route"))
	}
	if got.Title != "New Route" {
		t.Errorf("Title = %q, want %q", got.Title, "New Route")
	}
	if got.Status != 1 {
		t.Errorf("Status = %d, want %d", got.Status, 1)
	}
	if got.AuthorID != input.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, input.AuthorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapCreateAdminRouteParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateAdminRouteParams{Slug: types.Slug("unique-test")}

	got1 := d.MapCreateAdminRouteParams(input)
	got2 := d.MapCreateAdminRouteParams(input)
	if got1.AdminRouteID == got2.AdminRouteID {
		t.Error("two calls produced the same AdminRouteID -- each call should generate a unique ID")
	}
}

func TestDatabase_MapCreateAdminRouteParams_PreservesNullableFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateAdminRouteParams{
		Slug:     types.Slug("preserve-test"),
		Title:    "Test",
		AuthorID: authorID,
	}

	got := d.MapCreateAdminRouteParams(input)

	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

func TestDatabase_MapCreateAdminRouteParams_NullAuthorID(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateAdminRouteParams{
		Slug:     types.Slug("null-author"),
		AuthorID: types.NullableUserID{Valid: false},
	}

	got := d.MapCreateAdminRouteParams(input)

	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false")
	}
}

// --- SQLite Database.MapUpdateAdminRouteParams tests ---

func TestDatabase_MapUpdateAdminRouteParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateAdminRouteParams{
		Slug:         types.Slug("updated-route"),
		Title:        "Updated Route",
		Status:       2,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("original-route"),
	}

	got := d.MapUpdateAdminRouteParams(input)

	if got.Slug != types.Slug("updated-route") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("updated-route"))
	}
	if got.Title != "Updated Route" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated Route")
	}
	if got.Status != 2 {
		t.Errorf("Status = %d, want %d", got.Status, int64(2))
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
	// Slug_2 is the lookup key (WHERE clause)
	if got.Slug_2 != types.Slug("original-route") {
		t.Errorf("Slug_2 = %v, want %v", got.Slug_2, types.Slug("original-route"))
	}
}

func TestDatabase_MapUpdateAdminRouteParams_Slug2PreservedSeparately(t *testing.T) {
	t.Parallel()
	// Slug_2 should be preserved independently from Slug, allowing slug changes.
	d := Database{}
	input := UpdateAdminRouteParams{
		Slug:   types.Slug("new-slug"),
		Slug_2: types.Slug("old-slug"),
	}

	got := d.MapUpdateAdminRouteParams(input)

	if got.Slug == got.Slug_2 {
		t.Errorf("Slug and Slug_2 should differ: Slug=%v, Slug_2=%v", got.Slug, got.Slug_2)
	}
	if got.Slug != types.Slug("new-slug") {
		t.Errorf("Slug = %v, want new-slug", got.Slug)
	}
	if got.Slug_2 != types.Slug("old-slug") {
		t.Errorf("Slug_2 = %v, want old-slug", got.Slug_2)
	}
}

// --- MySQL MysqlDatabase.MapAdminRoute tests ---

func TestMysqlDatabase_MapAdminRoute_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	routeID := types.NewAdminRouteID()
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbm.AdminRoutes{
		AdminRouteID: routeID,
		Slug:         types.Slug("mysql-route"),
		Title:        "MySQL Route",
		Status:       int32(1),
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapAdminRoute(input)

	if got.AdminRouteID != routeID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, routeID)
	}
	if got.Slug != types.Slug("mysql-route") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("mysql-route"))
	}
	if got.Title != "MySQL Route" {
		t.Errorf("Title = %q, want %q", got.Title, "MySQL Route")
	}
	if got.Status != 1 {
		t.Errorf("Status = %d, want %d", got.Status, 1)
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

func TestMysqlDatabase_MapAdminRoute_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminRoute(mdbm.AdminRoutes{})

	if got.AdminRouteID != "" {
		t.Errorf("AdminRouteID = %v, want zero value", got.AdminRouteID)
	}
	if got.Slug != "" {
		t.Errorf("Slug = %v, want zero value", got.Slug)
	}
	if got.Title != "" {
		t.Errorf("Title = %q, want empty string", got.Title)
	}
	if got.Status != 0 {
		t.Errorf("Status = %d, want 0", got.Status)
	}
}

// int32->int64 conversion: verify MySQL Status (int32) is widened correctly
func TestMysqlDatabase_MapAdminRoute_StatusInt32Conversion(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	tests := []struct {
		name     string
		input    int32
		expected int64
	}{
		{"max_int32", math.MaxInt32, math.MaxInt32},
		{"min_int32", math.MinInt32, math.MinInt32},
		{"zero", 0, 0},
		{"negative_one", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbm.AdminRoutes{Status: tt.input}
			got := d.MapAdminRoute(input)
			if got.Status != tt.expected {
				t.Errorf("Status = %d, want %d", got.Status, tt.expected)
			}
		})
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminRouteParams tests ---

func TestMysqlDatabase_MapCreateAdminRouteParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminRouteParams{
		Slug:         types.Slug("mysql-new"),
		Title:        "MySQL New",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminRouteParams(input)

	if got.AdminRouteID.IsZero() {
		t.Fatal("expected non-zero AdminRouteID to be generated")
	}
	if got.Slug != types.Slug("mysql-new") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("mysql-new"))
	}
	if got.Title != "MySQL New" {
		t.Errorf("Title = %q, want %q", got.Title, "MySQL New")
	}
}

// Verify int64->int32 narrowing for MySQL create params
func TestMysqlDatabase_MapCreateAdminRouteParams_StatusInt32Narrowing(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreateAdminRouteParams{
		Slug:   types.Slug("narrow-test"),
		Status: 42,
	}

	got := d.MapCreateAdminRouteParams(input)

	if got.Status != int32(42) {
		t.Errorf("Status = %d, want %d", got.Status, int32(42))
	}
}

// --- MySQL MysqlDatabase.MapUpdateAdminRouteParams tests ---

func TestMysqlDatabase_MapUpdateAdminRouteParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := UpdateAdminRouteParams{
		Slug:         types.Slug("mysql-updated"),
		Title:        "MySQL Updated",
		Status:       3,
		AuthorID:     types.NullableUserID{Valid: false},
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("mysql-original"),
	}

	got := d.MapUpdateAdminRouteParams(input)

	if got.Slug != types.Slug("mysql-updated") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("mysql-updated"))
	}
	if got.Title != "MySQL Updated" {
		t.Errorf("Title = %q, want %q", got.Title, "MySQL Updated")
	}
	if got.Status != int32(3) {
		t.Errorf("Status = %d, want %d", got.Status, int32(3))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
	if got.Slug_2 != types.Slug("mysql-original") {
		t.Errorf("Slug_2 = %v, want %v", got.Slug_2, types.Slug("mysql-original"))
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminRoute tests ---

func TestPsqlDatabase_MapAdminRoute_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	routeID := types.NewAdminRouteID()
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbp.AdminRoutes{
		AdminRouteID: routeID,
		Slug:         types.Slug("psql-route"),
		Title:        "Psql Route",
		Status:       int32(1),
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapAdminRoute(input)

	if got.AdminRouteID != routeID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, routeID)
	}
	if got.Slug != types.Slug("psql-route") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("psql-route"))
	}
	if got.Title != "Psql Route" {
		t.Errorf("Title = %q, want %q", got.Title, "Psql Route")
	}
	if got.Status != 1 {
		t.Errorf("Status = %d, want %d", got.Status, 1)
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

func TestPsqlDatabase_MapAdminRoute_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminRoute(mdbp.AdminRoutes{})

	if got.AdminRouteID != "" {
		t.Errorf("AdminRouteID = %v, want zero value", got.AdminRouteID)
	}
	if got.Slug != "" {
		t.Errorf("Slug = %v, want zero value", got.Slug)
	}
	if got.Title != "" {
		t.Errorf("Title = %q, want empty string", got.Title)
	}
	if got.Status != 0 {
		t.Errorf("Status = %d, want 0", got.Status)
	}
}

// int32->int64 conversion: verify PostgreSQL Status (int32) is widened correctly
func TestPsqlDatabase_MapAdminRoute_StatusInt32Conversion(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	tests := []struct {
		name     string
		input    int32
		expected int64
	}{
		{"max_int32", math.MaxInt32, math.MaxInt32},
		{"min_int32", math.MinInt32, math.MinInt32},
		{"zero", 0, 0},
		{"negative_one", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbp.AdminRoutes{Status: tt.input}
			got := d.MapAdminRoute(input)
			if got.Status != tt.expected {
				t.Errorf("Status = %d, want %d", got.Status, tt.expected)
			}
		})
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminRouteParams tests ---

func TestPsqlDatabase_MapCreateAdminRouteParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminRouteParams{
		Slug:         types.Slug("psql-new"),
		Title:        "Psql New",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminRouteParams(input)

	if got.AdminRouteID.IsZero() {
		t.Fatal("expected non-zero AdminRouteID to be generated")
	}
	if got.Slug != types.Slug("psql-new") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("psql-new"))
	}
	if got.Title != "Psql New" {
		t.Errorf("Title = %q, want %q", got.Title, "Psql New")
	}
}

// Verify int64->int32 narrowing for PostgreSQL create params
func TestPsqlDatabase_MapCreateAdminRouteParams_StatusInt32Narrowing(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateAdminRouteParams{
		Slug:   types.Slug("narrow-test"),
		Status: 42,
	}

	got := d.MapCreateAdminRouteParams(input)

	if got.Status != int32(42) {
		t.Errorf("Status = %d, want %d", got.Status, int32(42))
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminRouteParams tests ---

func TestPsqlDatabase_MapUpdateAdminRouteParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := UpdateAdminRouteParams{
		Slug:         types.Slug("psql-updated"),
		Title:        "Psql Updated",
		Status:       3,
		AuthorID:     types.NullableUserID{Valid: false},
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("psql-original"),
	}

	got := d.MapUpdateAdminRouteParams(input)

	if got.Slug != types.Slug("psql-updated") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("psql-updated"))
	}
	if got.Title != "Psql Updated" {
		t.Errorf("Title = %q, want %q", got.Title, "Psql Updated")
	}
	if got.Status != int32(3) {
		t.Errorf("Status = %d, want %d", got.Status, int32(3))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
	if got.Slug_2 != types.Slug("psql-original") {
		t.Errorf("Slug_2 = %v, want %v", got.Slug_2, types.Slug("psql-original"))
	}
}

// --- Cross-database mapper consistency ---
// Verifies that all three database mappers produce identical AdminRoutes from equivalent input.

func TestCrossDatabaseMapAdminRoute_Consistency(t *testing.T) {
	t.Parallel()
	routeID := types.NewAdminRouteID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	sqliteInput := mdb.AdminRoutes{
		AdminRouteID: routeID, Slug: types.Slug("cross-test"),
		Title: "Cross Test", Status: 1,
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.AdminRoutes{
		AdminRouteID: routeID, Slug: types.Slug("cross-test"),
		Title: "Cross Test", Status: int32(1),
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.AdminRoutes{
		AdminRouteID: routeID, Slug: types.Slug("cross-test"),
		Title: "Cross Test", Status: int32(1),
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapAdminRoute(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminRoute(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminRoute(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminRouteParams consistency ---
// All three should auto-generate unique IDs.

func TestCrossDatabaseMapCreateAdminRouteParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminRouteParams{
		Slug:         types.Slug("cross-create"),
		Title:        "Cross Create",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateAdminRouteParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminRouteParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminRouteParams(input)

	if sqliteResult.AdminRouteID.IsZero() {
		t.Error("SQLite: expected non-zero generated AdminRouteID")
	}
	if mysqlResult.AdminRouteID.IsZero() {
		t.Error("MySQL: expected non-zero generated AdminRouteID")
	}
	if psqlResult.AdminRouteID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated AdminRouteID")
	}

	// Each call should generate a unique ID
	if sqliteResult.AdminRouteID == mysqlResult.AdminRouteID {
		t.Error("SQLite and MySQL generated the same AdminRouteID -- each call should be unique")
	}
	if sqliteResult.AdminRouteID == psqlResult.AdminRouteID {
		t.Error("SQLite and PostgreSQL generated the same AdminRouteID -- each call should be unique")
	}
}

// --- Cross-database MapUpdateAdminRouteParams: Slug_2 consistency ---

func TestCrossDatabaseMapUpdateAdminRouteParams_Slug2Preserved(t *testing.T) {
	t.Parallel()
	input := UpdateAdminRouteParams{
		Slug:   types.Slug("new-slug"),
		Slug_2: types.Slug("original-slug"),
	}

	sqliteResult := Database{}.MapUpdateAdminRouteParams(input)
	mysqlResult := MysqlDatabase{}.MapUpdateAdminRouteParams(input)
	psqlResult := PsqlDatabase{}.MapUpdateAdminRouteParams(input)

	if sqliteResult.Slug_2 != types.Slug("original-slug") {
		t.Errorf("SQLite Slug_2 = %v, want original-slug", sqliteResult.Slug_2)
	}
	if mysqlResult.Slug_2 != types.Slug("original-slug") {
		t.Errorf("MySQL Slug_2 = %v, want original-slug", mysqlResult.Slug_2)
	}
	if psqlResult.Slug_2 != types.Slug("original-slug") {
		t.Errorf("PostgreSQL Slug_2 = %v, want original-slug", psqlResult.Slug_2)
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewAdminRouteCmd_AllAccessors(t *testing.T) {
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
	params := CreateAdminRouteParams{
		Slug:         types.Slug("cmd-test"),
		Title:        "Cmd Test",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewAdminRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
	}
	p, ok := cmd.Params().(CreateAdminRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminRouteParams", cmd.Params())
	}
	if p.Slug != types.Slug("cmd-test") {
		t.Errorf("Params().Slug = %v, want cmd-test", p.Slug)
	}
	if p.Title != "Cmd Test" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "Cmd Test")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminRouteCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	routeID := types.NewAdminRouteID()
	cmd := NewAdminRouteCmd{}

	row := mdb.AdminRoutes{AdminRouteID: routeID}
	got := cmd.GetID(row)
	if got != string(routeID) {
		t.Errorf("GetID() = %q, want %q", got, string(routeID))
	}
}

func TestNewAdminRouteCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewAdminRouteCmd{}

	row := mdb.AdminRoutes{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateAdminRouteCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminRouteParams{
		Slug:         types.Slug("updated"),
		Title:        "Updated",
		Status:       2,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("original"),
	}

	cmd := Database{}.UpdateAdminRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
	}
	// UpdateAdminRouteCmd.GetID returns the Slug_2 (lookup key)
	if cmd.GetID() != string(types.Slug("original")) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(types.Slug("original")))
	}
	p, ok := cmd.Params().(UpdateAdminRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminRouteParams", cmd.Params())
	}
	if p.Title != "Updated" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "Updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminRouteCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	routeID := types.NewAdminRouteID()

	cmd := Database{}.DeleteAdminRouteCmd(ctx, ac, routeID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
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

func TestNewAdminRouteCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateAdminRouteParams{
		Slug:         types.Slug("mysql-cmd"),
		Title:        "MySQL Cmd",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewAdminRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
	}
	p, ok := cmd.Params().(CreateAdminRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminRouteParams", cmd.Params())
	}
	if p.Slug != types.Slug("mysql-cmd") {
		t.Errorf("Params().Slug = %v, want mysql-cmd", p.Slug)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminRouteCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	routeID := types.NewAdminRouteID()
	cmd := NewAdminRouteCmdMysql{}

	row := mdbm.AdminRoutes{AdminRouteID: routeID}
	got := cmd.GetID(row)
	if got != string(routeID) {
		t.Errorf("GetID() = %q, want %q", got, string(routeID))
	}
}

func TestUpdateAdminRouteCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminRouteParams{
		Slug:         types.Slug("mysql-updated"),
		Title:        "MySQL Updated",
		Status:       2,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("mysql-original"),
	}

	cmd := MysqlDatabase{}.UpdateAdminRouteCmd(ctx, ac, params)

	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
	}
	// MySQL UpdateAdminRouteCmd.GetID returns the Slug_2
	if cmd.GetID() != string(types.Slug("mysql-original")) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(types.Slug("mysql-original")))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminRouteParams", cmd.Params())
	}
	if p.Title != "MySQL Updated" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "MySQL Updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminRouteCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	routeID := types.NewAdminRouteID()

	cmd := MysqlDatabase{}.DeleteAdminRouteCmd(ctx, ac, routeID)

	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
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

func TestNewAdminRouteCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateAdminRouteParams{
		Slug:         types.Slug("psql-cmd"),
		Title:        "Psql Cmd",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewAdminRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
	}
	p, ok := cmd.Params().(CreateAdminRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminRouteParams", cmd.Params())
	}
	if p.Slug != types.Slug("psql-cmd") {
		t.Errorf("Params().Slug = %v, want psql-cmd", p.Slug)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminRouteCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	routeID := types.NewAdminRouteID()
	cmd := NewAdminRouteCmdPsql{}

	row := mdbp.AdminRoutes{AdminRouteID: routeID}
	got := cmd.GetID(row)
	if got != string(routeID) {
		t.Errorf("GetID() = %q, want %q", got, string(routeID))
	}
}

func TestUpdateAdminRouteCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminRouteParams{
		Slug:         types.Slug("psql-updated"),
		Title:        "Psql Updated",
		Status:       2,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("psql-original"),
	}

	cmd := PsqlDatabase{}.UpdateAdminRouteCmd(ctx, ac, params)

	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
	}
	if cmd.GetID() != string(types.Slug("psql-original")) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(types.Slug("psql-original")))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminRouteParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminRouteParams", cmd.Params())
	}
	if p.Title != "Psql Updated" {
		t.Errorf("Params().Title = %q, want %q", p.Title, "Psql Updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminRouteCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	routeID := types.NewAdminRouteID()

	cmd := PsqlDatabase{}.DeleteAdminRouteCmd(ctx, ac, routeID)

	if cmd.TableName() != "admin_routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_routes")
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
// Verify all three database types produce commands with the correct table name.

func TestAuditedAdminRouteCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateAdminRouteParams{}
	updateParams := UpdateAdminRouteParams{Slug_2: types.Slug("lookup")}
	routeID := types.NewAdminRouteID()

	// SQLite
	sqliteCreate := Database{}.NewAdminRouteCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateAdminRouteCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteAdminRouteCmd(ctx, ac, routeID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewAdminRouteCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateAdminRouteCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteAdminRouteCmd(ctx, ac, routeID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewAdminRouteCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateAdminRouteCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteAdminRouteCmd(ctx, ac, routeID)

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
			if c.name != "admin_routes" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_routes")
			}
		})
	}
}

func TestAuditedAdminRouteCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminRouteParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminRouteCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminRouteCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminRouteCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminRouteCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	routeID := types.NewAdminRouteID()

	t.Run("UpdateCmd GetID returns Slug_2", func(t *testing.T) {
		t.Parallel()
		// Admin route update commands use Slug_2 as the lookup identifier
		params := UpdateAdminRouteParams{Slug_2: types.Slug("update-lookup")}

		sqliteCmd := Database{}.UpdateAdminRouteCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminRouteCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminRouteCmd(ctx, ac, params)

		wantID := "update-lookup"
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

	t.Run("DeleteCmd GetID returns AdminRouteID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminRouteCmd(ctx, ac, routeID)
		mysqlCmd := MysqlDatabase{}.DeleteAdminRouteCmd(ctx, ac, routeID)
		psqlCmd := PsqlDatabase{}.DeleteAdminRouteCmd(ctx, ac, routeID)

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
		testRouteID := types.NewAdminRouteID()

		sqliteCmd := NewAdminRouteCmd{}
		mysqlCmd := NewAdminRouteCmdMysql{}
		psqlCmd := NewAdminRouteCmdPsql{}

		wantID := string(testRouteID)

		sqliteRow := mdb.AdminRoutes{AdminRouteID: testRouteID}
		mysqlRow := mdbm.AdminRoutes{AdminRouteID: testRouteID}
		psqlRow := mdbp.AdminRoutes{AdminRouteID: testRouteID}

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

// --- Edge case: UpdateAdminRouteCmd with empty Slug_2 ---

func TestUpdateAdminRouteCmd_GetID_EmptySlug(t *testing.T) {
	t.Parallel()
	params := UpdateAdminRouteParams{Slug_2: ""}

	sqliteCmd := Database{}.UpdateAdminRouteCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminRouteCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminRouteCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteAdminRouteCmd with empty ID ---

func TestDeleteAdminRouteCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.AdminRouteID("")

	sqliteCmd := Database{}.DeleteAdminRouteCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminRouteCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminRouteCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.AdminRoutes]  = NewAdminRouteCmd{}
	_ audited.UpdateCommand[mdb.AdminRoutes]  = UpdateAdminRouteCmd{}
	_ audited.DeleteCommand[mdb.AdminRoutes]  = DeleteAdminRouteCmd{}
	_ audited.CreateCommand[mdbm.AdminRoutes] = NewAdminRouteCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminRoutes] = UpdateAdminRouteCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminRoutes] = DeleteAdminRouteCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminRoutes] = NewAdminRouteCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminRoutes] = UpdateAdminRouteCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminRoutes] = DeleteAdminRouteCmdPsql{}
)

// --- Struct field correctness ---
// Verify that the wrapper AdminRoutes struct and param structs hold values correctly via JSON.

func TestAdminRoutesStruct_JSONTags(t *testing.T) {
	t.Parallel()
	ar, _, _, _ := adminRouteFixture()

	data, err := json.Marshal(ar)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"admin_route_id", "slug", "title",
		"status", "author_id", "date_created", "date_modified",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestCreateAdminRouteParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	p := CreateAdminRouteParams{
		Slug:         types.Slug("json-test"),
		Title:        "JSON Test",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"slug", "title", "status",
		"author_id", "date_created", "date_modified",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestUpdateAdminRouteParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	p := UpdateAdminRouteParams{
		Slug:         types.Slug("json-update"),
		Title:        "JSON Update Test",
		Status:       2,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("json-original"),
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"slug", "title", "status",
		"author_id", "date_created", "date_modified",
		"slug_2",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestUtilityGetAdminRoutesRow_JSONTags(t *testing.T) {
	t.Parallel()
	row := UtilityGetAdminRoutesRow{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         types.Slug("utility-row"),
	}

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"admin_route_id", "slug",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

// --- JSON round-trip tests ---

func TestAdminRoutesStruct_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	ar, _, _, _ := adminRouteFixture()

	data, err := json.Marshal(ar)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got AdminRoutes
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.AdminRouteID != ar.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, ar.AdminRouteID)
	}
	if got.Slug != ar.Slug {
		t.Errorf("Slug = %v, want %v", got.Slug, ar.Slug)
	}
	if got.Title != ar.Title {
		t.Errorf("Title = %q, want %q", got.Title, ar.Title)
	}
	if got.Status != ar.Status {
		t.Errorf("Status = %d, want %d", got.Status, ar.Status)
	}
}
