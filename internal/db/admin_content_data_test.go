package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// adminContentDataFixture returns a fully populated AdminContentData and its parts for testing.
func adminContentDataFixture() (AdminContentData, types.AdminContentID, types.Timestamp) {
	contentID := types.NewAdminContentID()
	ts := types.NewTimestamp(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC))
	parentID := types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	datatypeID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(types.NewAdminContentID()), Valid: true}

	acd := AdminContentData{
		AdminContentDataID: contentID,
		ParentID:           parentID,
		FirstChildID:       sql.NullString{String: "first-child-001", Valid: true},
		NextSiblingID:      sql.NullString{String: "next-sibling-001", Valid: true},
		PrevSiblingID:      sql.NullString{String: "prev-sibling-001", Valid: true},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("admin-route-001"), Valid: true},
		AdminDatatypeID:    datatypeID,
		AuthorID:           authorID,
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
	}
	return acd, contentID, ts
}

// adminContentDataFixtureNulls returns an AdminContentData where all nullable fields are null/invalid.
func adminContentDataFixtureNulls() AdminContentData {
	contentID := types.NewAdminContentID()
	ts := types.NewTimestamp(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	return AdminContentData{
		AdminContentDataID: contentID,
		ParentID:           types.NullableAdminContentID{Valid: false},
		FirstChildID:       sql.NullString{Valid: false},
		NextSiblingID:      sql.NullString{Valid: false},
		PrevSiblingID:      sql.NullString{Valid: false},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("route-null-test"), Valid: true},
		AdminDatatypeID:    types.NullableAdminDatatypeID{Valid: false},
		AuthorID:           types.NullableUserID{Valid: false},
		Status:             types.ContentStatus("draft"),
		DateCreated:        ts,
		DateModified:       ts,
	}
}

// --- MapAdminContentDataJSON tests ---

func TestMapAdminContentDataJSON_AllFieldsValid(t *testing.T) {
	t.Parallel()
	acd, contentID, ts := adminContentDataFixture()

	got := MapAdminContentDataJSON(acd)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ContentDataID", got.ContentDataID, contentID.String()},
		{"ParentID", got.ParentID, acd.ParentID.String()},
		{"FirstChildID", got.FirstChildID, "first-child-001"},
		{"NextSiblingID", got.NextSiblingID, "next-sibling-001"},
		{"PrevSiblingID", got.PrevSiblingID, "prev-sibling-001"},
		{"RouteID", got.RouteID, "admin-route-001"},
		{"DatatypeID", got.DatatypeID, acd.AdminDatatypeID.String()},
		{"AuthorID", got.AuthorID, acd.AuthorID.String()},
		{"Status", got.Status, "published"},
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

func TestMapAdminContentDataJSON_NullSiblingFields(t *testing.T) {
	t.Parallel()
	// When FirstChildID, NextSiblingID, PrevSiblingID are null (Valid=false),
	// MapAdminContentDataJSON should produce empty strings for those fields.
	acd := adminContentDataFixtureNulls()

	got := MapAdminContentDataJSON(acd)

	if got.FirstChildID != "" {
		t.Errorf("FirstChildID = %q, want empty string for null", got.FirstChildID)
	}
	if got.NextSiblingID != "" {
		t.Errorf("NextSiblingID = %q, want empty string for null", got.NextSiblingID)
	}
	if got.PrevSiblingID != "" {
		t.Errorf("PrevSiblingID = %q, want empty string for null", got.PrevSiblingID)
	}
}

func TestMapAdminContentDataJSON_NullableIDsProduceNullString(t *testing.T) {
	t.Parallel()
	// When ParentID, AdminDatatypeID, AuthorID are null (Valid=false),
	// their String() method returns "null".
	acd := adminContentDataFixtureNulls()

	got := MapAdminContentDataJSON(acd)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q for null NullableContentID", got.ParentID, "null")
	}
	if got.DatatypeID != "null" {
		t.Errorf("DatatypeID = %q, want %q for null NullableAdminDatatypeID", got.DatatypeID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null NullableUserID", got.AuthorID, "null")
	}
}

func TestMapAdminContentDataJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value AdminContentData should not panic
	got := MapAdminContentDataJSON(AdminContentData{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %q, want empty string for zero AdminContentID", got.ContentDataID)
	}
	// NullableXxx with zero value has Valid=false, so String() returns "null"
	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
	// sql.NullString zero value has Valid=false, so FirstChildID should be ""
	if got.FirstChildID != "" {
		t.Errorf("FirstChildID = %q, want empty string", got.FirstChildID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
}

func TestMapAdminContentDataJSON_MapsAdminIDsToPublicShape(t *testing.T) {
	t.Parallel()
	// Verify that admin-specific IDs are mapped into the public ContentDataJSON field names.
	// This is the core purpose of this function: AdminContentDataID -> ContentDataID,
	// AdminRouteID -> RouteID, AdminDatatypeID -> DatatypeID.
	acd, _, _ := adminContentDataFixture()
	got := MapAdminContentDataJSON(acd)

	if got.ContentDataID != acd.AdminContentDataID.String() {
		t.Errorf("ContentDataID = %q, want AdminContentDataID %q", got.ContentDataID, acd.AdminContentDataID.String())
	}
	if got.RouteID != acd.AdminRouteID.String() {
		t.Errorf("RouteID = %q, want AdminRouteID %q", got.RouteID, acd.AdminRouteID.String())
	}
	if got.DatatypeID != acd.AdminDatatypeID.String() {
		t.Errorf("DatatypeID = %q, want AdminDatatypeID %q", got.DatatypeID, acd.AdminDatatypeID.String())
	}
}

// --- MapStringAdminContentData tests ---

func TestMapStringAdminContentData_AllFields(t *testing.T) {
	t.Parallel()
	acd, contentID, ts := adminContentDataFixture()

	got := MapStringAdminContentData(acd)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AdminContentDataID", got.AdminContentDataID, contentID.String()},
		{"ParentID", got.ParentID, acd.ParentID.String()},
		{"AdminRouteID", got.AdminRouteID, "admin-route-001"},
		{"AdminDatatypeID", got.AdminDatatypeID, acd.AdminDatatypeID.String()},
		{"AuthorID", got.AuthorID, acd.AuthorID.String()},
		{"Status", got.Status, "published"},
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

func TestMapStringAdminContentData_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// History field is hard-coded to "" per the comment in the source.
	acd, _, _ := adminContentDataFixture()

	got := MapStringAdminContentData(acd)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (History field removed)", got.History)
	}
}

func TestMapStringAdminContentData_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringAdminContentData(AdminContentData{})

	if got.AdminContentDataID != "" {
		t.Errorf("AdminContentDataID = %q, want empty string", got.AdminContentDataID)
	}
	if got.AdminRouteID != "null" {
		t.Errorf("AdminRouteID = %q, want %q for zero NullableAdminRouteID", got.AdminRouteID, "null")
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
}

func TestMapStringAdminContentData_NullableFieldsShowNull(t *testing.T) {
	t.Parallel()
	acd := adminContentDataFixtureNulls()

	got := MapStringAdminContentData(acd)

	// Nullable IDs with Valid=false produce "null" via their String() method
	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q for null", got.ParentID, "null")
	}
	if got.AdminDatatypeID != "null" {
		t.Errorf("AdminDatatypeID = %q, want %q for null", got.AdminDatatypeID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null", got.AuthorID, "null")
	}
}

// --- SQLite Database.MapAdminContentData tests ---

func TestDatabase_MapAdminContentData_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	contentID := types.NewAdminContentID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	parentID := types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	datatypeID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(types.NewAdminContentID()), Valid: true}

	input := mdb.AdminContentData{
		AdminContentDataID: contentID,
		ParentID:           parentID,
		FirstChildID:       sql.NullString{String: "child-1", Valid: true},
		NextSiblingID:      sql.NullString{String: "next-1", Valid: true},
		PrevSiblingID:      sql.NullString{String: "prev-1", Valid: true},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("route-1"), Valid: true},
		AdminDatatypeID:    datatypeID,
		AuthorID:           authorID,
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
	}

	got := d.MapAdminContentData(input)

	if got.AdminContentDataID != contentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.NextSiblingID != input.NextSiblingID {
		t.Errorf("NextSiblingID = %v, want %v", got.NextSiblingID, input.NextSiblingID)
	}
	if got.PrevSiblingID != input.PrevSiblingID {
		t.Errorf("PrevSiblingID = %v, want %v", got.PrevSiblingID, input.PrevSiblingID)
	}
	if got.AdminRouteID.String() != "route-1" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "route-1")
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.Status != types.ContentStatus("published") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("published"))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapAdminContentData_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminContentData(mdb.AdminContentData{})

	if got.AdminContentDataID != "" {
		t.Errorf("AdminContentDataID = %v, want zero value", got.AdminContentDataID)
	}
	if got.AdminRouteID != (types.NullableAdminRouteID{}) {
		t.Errorf("AdminRouteID = %v, want zero value", got.AdminRouteID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
	if got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = true, want false for zero value")
	}
}

// --- SQLite Database.MapCreateAdminContentDataParams tests ---

func TestDatabase_MapCreateAdminContentDataParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminContentDataParams{
		ParentID:        types.NullableAdminContentID{Valid: false},
		FirstChildID:    sql.NullString{Valid: false},
		NextSiblingID:   sql.NullString{Valid: false},
		PrevSiblingID:   sql.NullString{Valid: false},
		AdminRouteID:    types.NullableAdminRouteID{ID: types.AdminRouteID("new-route"), Valid: true},
		AdminDatatypeID: types.NullableAdminDatatypeID{Valid: false},
		AuthorID:        types.NullableUserID{ID: types.NewUserID(), Valid: true},
		Status:          types.ContentStatus("draft"),
		DateCreated:     ts,
		DateModified:    ts,
	}

	got := d.MapCreateAdminContentDataParams(input)

	if got.AdminContentDataID.IsZero() {
		t.Fatal("expected non-zero AdminContentDataID to be generated")
	}
	if got.AdminRouteID.String() != "new-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "new-route")
	}
	if got.Status != types.ContentStatus("draft") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("draft"))
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

func TestDatabase_MapCreateAdminContentDataParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateAdminContentDataParams{AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("test"), Valid: true}}

	got1 := d.MapCreateAdminContentDataParams(input)
	got2 := d.MapCreateAdminContentDataParams(input)
	if got1.AdminContentDataID == got2.AdminContentDataID {
		t.Error("two calls produced the same AdminContentDataID -- each call should generate a unique ID")
	}
}

func TestDatabase_MapCreateAdminContentDataParams_PreservesNullableFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	parentID := types.NullableAdminContentID{ID: types.AdminContentID("parent-123"), Valid: true}
	firstChild := sql.NullString{String: "child-abc", Valid: true}
	nextSibling := sql.NullString{String: "next-def", Valid: true}
	prevSibling := sql.NullString{String: "prev-ghi", Valid: true}
	datatypeID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID("dt-456"), Valid: true}

	input := CreateAdminContentDataParams{
		ParentID:        parentID,
		FirstChildID:    firstChild,
		NextSiblingID:   nextSibling,
		PrevSiblingID:   prevSibling,
		AdminDatatypeID: datatypeID,
	}

	got := d.MapCreateAdminContentDataParams(input)

	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != firstChild {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, firstChild)
	}
	if got.NextSiblingID != nextSibling {
		t.Errorf("NextSiblingID = %v, want %v", got.NextSiblingID, nextSibling)
	}
	if got.PrevSiblingID != prevSibling {
		t.Errorf("PrevSiblingID = %v, want %v", got.PrevSiblingID, prevSibling)
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
}

// --- SQLite Database.MapUpdateAdminContentDataParams tests ---

func TestDatabase_MapUpdateAdminContentDataParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	contentID := types.NewAdminContentID()
	parentID := types.NullableAdminContentID{ID: types.AdminContentID("parent-updated"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateAdminContentDataParams{
		ParentID:           parentID,
		FirstChildID:       sql.NullString{String: "updated-child", Valid: true},
		NextSiblingID:      sql.NullString{String: "updated-next", Valid: true},
		PrevSiblingID:      sql.NullString{String: "updated-prev", Valid: true},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("updated-route"), Valid: true},
		AdminDatatypeID:    types.NullableAdminDatatypeID{Valid: false},
		AuthorID:           authorID,
		Status:             types.ContentStatus("archived"),
		DateCreated:        ts,
		DateModified:       ts,
		AdminContentDataID: contentID,
	}

	got := d.MapUpdateAdminContentDataParams(input)

	if got.AdminContentDataID != contentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID.String != "updated-child" {
		t.Errorf("FirstChildID.String = %q, want %q", got.FirstChildID.String, "updated-child")
	}
	if got.AdminRouteID.String() != "updated-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "updated-route")
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.Status != types.ContentStatus("archived") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("archived"))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

// --- MySQL MysqlDatabase.MapAdminContentData tests ---

func TestMysqlDatabase_MapAdminContentData_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	contentID := types.NewAdminContentID()
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))
	parentID := types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	datatypeID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(types.NewAdminContentID()), Valid: true}

	input := mdbm.AdminContentData{
		AdminContentDataID: contentID,
		ParentID:           parentID,
		FirstChildID:       sql.NullString{String: "mysql-child", Valid: true},
		NextSiblingID:      sql.NullString{String: "mysql-next", Valid: true},
		PrevSiblingID:      sql.NullString{String: "mysql-prev", Valid: true},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-route"), Valid: true},
		AdminDatatypeID:    datatypeID,
		AuthorID:           authorID,
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
	}

	got := d.MapAdminContentData(input)

	if got.AdminContentDataID != contentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.AdminRouteID.String() != "mysql-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "mysql-route")
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.Status != types.ContentStatus("published") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("published"))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestMysqlDatabase_MapAdminContentData_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminContentData(mdbm.AdminContentData{})

	if got.AdminContentDataID != "" {
		t.Errorf("AdminContentDataID = %v, want zero value", got.AdminContentDataID)
	}
	if got.AdminRouteID != (types.NullableAdminRouteID{}) {
		t.Errorf("AdminRouteID = %v, want zero value", got.AdminRouteID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminContentDataParams tests ---

func TestMysqlDatabase_MapCreateAdminContentDataParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminContentDataParams{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-new-route"), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminContentDataParams(input)

	if got.AdminContentDataID.IsZero() {
		t.Fatal("expected non-zero AdminContentDataID to be generated")
	}
	if got.AdminRouteID.String() != "mysql-new-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "mysql-new-route")
	}
	if got.Status != types.ContentStatus("draft") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("draft"))
	}
}

// --- MySQL MysqlDatabase.MapUpdateAdminContentDataParams tests ---

func TestMysqlDatabase_MapUpdateAdminContentDataParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	contentID := types.NewAdminContentID()

	input := UpdateAdminContentDataParams{
		ParentID:           types.NullableAdminContentID{Valid: false},
		FirstChildID:       sql.NullString{String: "mysql-updated-child", Valid: true},
		NextSiblingID:      sql.NullString{Valid: false},
		PrevSiblingID:      sql.NullString{Valid: false},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-updated-route"), Valid: true},
		AdminDatatypeID:    types.NullableAdminDatatypeID{Valid: false},
		AuthorID:           types.NullableUserID{Valid: false},
		Status:             types.ContentStatus("archived"),
		DateCreated:        ts,
		DateModified:       ts,
		AdminContentDataID: contentID,
	}

	got := d.MapUpdateAdminContentDataParams(input)

	if got.AdminContentDataID != contentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, contentID)
	}
	if got.AdminRouteID.String() != "mysql-updated-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "mysql-updated-route")
	}
	if got.Status != types.ContentStatus("archived") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("archived"))
	}
	if got.FirstChildID.String != "mysql-updated-child" {
		t.Errorf("FirstChildID.String = %q, want %q", got.FirstChildID.String, "mysql-updated-child")
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminContentData tests ---

func TestPsqlDatabase_MapAdminContentData_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	contentID := types.NewAdminContentID()
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))
	parentID := types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	datatypeID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(types.NewAdminContentID()), Valid: true}

	input := mdbp.AdminContentData{
		AdminContentDataID: contentID,
		ParentID:           parentID,
		FirstChildID:       sql.NullString{String: "psql-child", Valid: true},
		NextSiblingID:      sql.NullString{String: "psql-next", Valid: true},
		PrevSiblingID:      sql.NullString{String: "psql-prev", Valid: true},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("psql-route"), Valid: true},
		AdminDatatypeID:    datatypeID,
		AuthorID:           authorID,
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
	}

	got := d.MapAdminContentData(input)

	if got.AdminContentDataID != contentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.AdminRouteID.String() != "psql-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "psql-route")
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.Status != types.ContentStatus("published") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("published"))
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestPsqlDatabase_MapAdminContentData_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminContentData(mdbp.AdminContentData{})

	if got.AdminContentDataID != "" {
		t.Errorf("AdminContentDataID = %v, want zero value", got.AdminContentDataID)
	}
	if got.AdminRouteID != (types.NullableAdminRouteID{}) {
		t.Errorf("AdminRouteID = %v, want zero value", got.AdminRouteID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminContentDataParams tests ---

func TestPsqlDatabase_MapCreateAdminContentDataParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminContentDataParams{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("psql-new-route"), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminContentDataParams(input)

	if got.AdminContentDataID.IsZero() {
		t.Fatal("expected non-zero AdminContentDataID to be generated")
	}
	if got.AdminRouteID.String() != "psql-new-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "psql-new-route")
	}
	if got.Status != types.ContentStatus("draft") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("draft"))
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminContentDataParams tests ---

func TestPsqlDatabase_MapUpdateAdminContentDataParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	contentID := types.NewAdminContentID()

	input := UpdateAdminContentDataParams{
		ParentID:           types.NullableAdminContentID{Valid: false},
		FirstChildID:       sql.NullString{String: "psql-updated-child", Valid: true},
		NextSiblingID:      sql.NullString{Valid: false},
		PrevSiblingID:      sql.NullString{Valid: false},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("psql-updated-route"), Valid: true},
		AdminDatatypeID:    types.NullableAdminDatatypeID{Valid: false},
		AuthorID:           types.NullableUserID{Valid: false},
		Status:             types.ContentStatus("pending"),
		DateCreated:        ts,
		DateModified:       ts,
		AdminContentDataID: contentID,
	}

	got := d.MapUpdateAdminContentDataParams(input)

	if got.AdminContentDataID != contentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, contentID)
	}
	if got.AdminRouteID.String() != "psql-updated-route" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID.String(), "psql-updated-route")
	}
	if got.Status != types.ContentStatus("pending") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("pending"))
	}
	if got.FirstChildID.String != "psql-updated-child" {
		t.Errorf("FirstChildID.String = %q, want %q", got.FirstChildID.String, "psql-updated-child")
	}
}

// --- Cross-database mapper consistency ---
// Verifies that all three database mappers produce identical AdminContentData from equivalent input.

func TestCrossDatabaseMapAdminContentData_Consistency(t *testing.T) {
	t.Parallel()
	contentID := types.NewAdminContentID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	parentID := types.NullableAdminContentID{ID: types.AdminContentID("parent-cross"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	datatypeID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID("dt-cross"), Valid: true}
	firstChild := sql.NullString{String: "cross-child", Valid: true}
	nextSibling := sql.NullString{String: "cross-next", Valid: true}
	prevSibling := sql.NullString{String: "cross-prev", Valid: true}

	sqliteInput := mdb.AdminContentData{
		AdminContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("cross-route"), Valid: true}, AdminDatatypeID: datatypeID, AuthorID: authorID,
		Status: types.ContentStatus("published"), DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.AdminContentData{
		AdminContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("cross-route"), Valid: true}, AdminDatatypeID: datatypeID, AuthorID: authorID,
		Status: types.ContentStatus("published"), DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.AdminContentData{
		AdminContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("cross-route"), Valid: true}, AdminDatatypeID: datatypeID, AuthorID: authorID,
		Status: types.ContentStatus("published"), DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapAdminContentData(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminContentData(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminContentData(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminContentDataParams consistency ---
// All three should auto-generate unique IDs.

func TestCrossDatabaseMapCreateAdminContentDataParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminContentDataParams{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("cross-create-route"), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateAdminContentDataParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminContentDataParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminContentDataParams(input)

	if sqliteResult.AdminContentDataID.IsZero() {
		t.Error("SQLite: expected non-zero generated AdminContentDataID")
	}
	if mysqlResult.AdminContentDataID.IsZero() {
		t.Error("MySQL: expected non-zero generated AdminContentDataID")
	}
	if psqlResult.AdminContentDataID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated AdminContentDataID")
	}

	// Each call should generate a unique ID
	if sqliteResult.AdminContentDataID == mysqlResult.AdminContentDataID {
		t.Error("SQLite and MySQL generated the same AdminContentDataID -- each call should be unique")
	}
	if sqliteResult.AdminContentDataID == psqlResult.AdminContentDataID {
		t.Error("SQLite and PostgreSQL generated the same AdminContentDataID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewAdminContentDataCmd_AllAccessors(t *testing.T) {
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
	params := CreateAdminContentDataParams{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("cmd-route"), Valid: true},
		Status:       types.ContentStatus("draft"),
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewAdminContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	p, ok := cmd.Params().(CreateAdminContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentDataParams", cmd.Params())
	}
	if p.AdminRouteID.String() != "cmd-route" {
		t.Errorf("Params().AdminRouteID = %q, want %q", p.AdminRouteID.String(), "cmd-route")
	}
	if p.Status != types.ContentStatus("draft") {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.ContentStatus("draft"))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminContentDataCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	contentID := types.NewAdminContentID()
	cmd := NewAdminContentDataCmd{}

	row := mdb.AdminContentData{AdminContentDataID: contentID}
	got := cmd.GetID(row)
	if got != string(contentID) {
		t.Errorf("GetID() = %q, want %q", got, string(contentID))
	}
}

func TestNewAdminContentDataCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewAdminContentDataCmd{}

	row := mdb.AdminContentData{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateAdminContentDataCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	contentID := types.NewAdminContentID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminContentDataParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("updated-route"), Valid: true},
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
		AdminContentDataID: contentID,
	}

	cmd := Database{}.UpdateAdminContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	if cmd.GetID() != string(contentID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(contentID))
	}
	p, ok := cmd.Params().(UpdateAdminContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentDataParams", cmd.Params())
	}
	if p.AdminRouteID.String() != "updated-route" {
		t.Errorf("Params().AdminRouteID = %q, want %q", p.AdminRouteID.String(), "updated-route")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentDataCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	contentID := types.NewAdminContentID()

	cmd := Database{}.DeleteAdminContentDataCmd(ctx, ac, contentID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	if cmd.GetID() != string(contentID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(contentID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewAdminContentDataCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateAdminContentDataParams{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-cmd-route"), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewAdminContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	p, ok := cmd.Params().(CreateAdminContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentDataParams", cmd.Params())
	}
	if p.AdminRouteID.String() != "mysql-cmd-route" {
		t.Errorf("Params().AdminRouteID = %q, want %q", p.AdminRouteID.String(), "mysql-cmd-route")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminContentDataCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	contentID := types.NewAdminContentID()
	cmd := NewAdminContentDataCmdMysql{}

	row := mdbm.AdminContentData{AdminContentDataID: contentID}
	got := cmd.GetID(row)
	if got != string(contentID) {
		t.Errorf("GetID() = %q, want %q", got, string(contentID))
	}
}

func TestUpdateAdminContentDataCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	contentID := types.NewAdminContentID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminContentDataParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-updated"), Valid: true},
		Status:             types.ContentStatus("archived"),
		DateCreated:        ts,
		DateModified:       ts,
		AdminContentDataID: contentID,
	}

	cmd := MysqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, params)

	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	if cmd.GetID() != string(contentID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(contentID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentDataParams", cmd.Params())
	}
	if p.AdminRouteID.String() != "mysql-updated" {
		t.Errorf("Params().AdminRouteID = %q, want %q", p.AdminRouteID.String(), "mysql-updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentDataCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	contentID := types.NewAdminContentID()

	cmd := MysqlDatabase{}.DeleteAdminContentDataCmd(ctx, ac, contentID)

	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	if cmd.GetID() != string(contentID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(contentID))
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

func TestNewAdminContentDataCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateAdminContentDataParams{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("psql-cmd-route"), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewAdminContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	p, ok := cmd.Params().(CreateAdminContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentDataParams", cmd.Params())
	}
	if p.AdminRouteID.String() != "psql-cmd-route" {
		t.Errorf("Params().AdminRouteID = %q, want %q", p.AdminRouteID.String(), "psql-cmd-route")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminContentDataCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	contentID := types.NewAdminContentID()
	cmd := NewAdminContentDataCmdPsql{}

	row := mdbp.AdminContentData{AdminContentDataID: contentID}
	got := cmd.GetID(row)
	if got != string(contentID) {
		t.Errorf("GetID() = %q, want %q", got, string(contentID))
	}
}

func TestUpdateAdminContentDataCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	contentID := types.NewAdminContentID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminContentDataParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("psql-updated"), Valid: true},
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
		AdminContentDataID: contentID,
	}

	cmd := PsqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, params)

	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	if cmd.GetID() != string(contentID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(contentID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentDataParams", cmd.Params())
	}
	if p.AdminRouteID.String() != "psql-updated" {
		t.Errorf("Params().AdminRouteID = %q, want %q", p.AdminRouteID.String(), "psql-updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentDataCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	contentID := types.NewAdminContentID()

	cmd := PsqlDatabase{}.DeleteAdminContentDataCmd(ctx, ac, contentID)

	if cmd.TableName() != "admin_content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_data")
	}
	if cmd.GetID() != string(contentID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(contentID))
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

func TestAuditedAdminContentDataCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateAdminContentDataParams{}
	updateParams := UpdateAdminContentDataParams{AdminContentDataID: types.NewAdminContentID()}
	contentID := types.NewAdminContentID()

	// SQLite
	sqliteCreate := Database{}.NewAdminContentDataCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateAdminContentDataCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteAdminContentDataCmd(ctx, ac, contentID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewAdminContentDataCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteAdminContentDataCmd(ctx, ac, contentID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewAdminContentDataCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteAdminContentDataCmd(ctx, ac, contentID)

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
			if c.name != "admin_content_data" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_content_data")
			}
		})
	}
}

func TestAuditedAdminContentDataCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminContentDataParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminContentDataCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminContentDataCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminContentDataCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminContentDataCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	contentID := types.NewAdminContentID()

	t.Run("UpdateCmd GetID returns AdminContentDataID", func(t *testing.T) {
		t.Parallel()
		params := UpdateAdminContentDataParams{AdminContentDataID: contentID}

		sqliteCmd := Database{}.UpdateAdminContentDataCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, params)

		wantID := string(contentID)
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

	t.Run("DeleteCmd GetID returns AdminContentDataID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminContentDataCmd(ctx, ac, contentID)
		mysqlCmd := MysqlDatabase{}.DeleteAdminContentDataCmd(ctx, ac, contentID)
		psqlCmd := PsqlDatabase{}.DeleteAdminContentDataCmd(ctx, ac, contentID)

		wantID := string(contentID)
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
		testContentID := types.NewAdminContentID()

		sqliteCmd := NewAdminContentDataCmd{}
		mysqlCmd := NewAdminContentDataCmdMysql{}
		psqlCmd := NewAdminContentDataCmdPsql{}

		wantID := string(testContentID)

		sqliteRow := mdb.AdminContentData{AdminContentDataID: testContentID}
		mysqlRow := mdbm.AdminContentData{AdminContentDataID: testContentID}
		psqlRow := mdbp.AdminContentData{AdminContentDataID: testContentID}

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

// --- Edge case: UpdateAdminContentDataCmd with empty AdminContentDataID ---

func TestUpdateAdminContentDataCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateAdminContentDataParams{AdminContentDataID: ""}

	sqliteCmd := Database{}.UpdateAdminContentDataCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminContentDataCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminContentDataCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteAdminContentDataCmd with empty ID ---

func TestDeleteAdminContentDataCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.AdminContentID("")

	sqliteCmd := Database{}.DeleteAdminContentDataCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminContentDataCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminContentDataCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.AdminContentData]  = NewAdminContentDataCmd{}
	_ audited.UpdateCommand[mdb.AdminContentData]  = UpdateAdminContentDataCmd{}
	_ audited.DeleteCommand[mdb.AdminContentData]  = DeleteAdminContentDataCmd{}
	_ audited.CreateCommand[mdbm.AdminContentData] = NewAdminContentDataCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminContentData] = UpdateAdminContentDataCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminContentData] = DeleteAdminContentDataCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminContentData] = NewAdminContentDataCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminContentData] = UpdateAdminContentDataCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminContentData] = DeleteAdminContentDataCmdPsql{}
)

// --- Struct field correctness ---
// Verify that the wrapper AdminContentData struct and param structs hold values correctly via JSON.

func TestAdminContentDataStruct_JSONTags(t *testing.T) {
	t.Parallel()
	acd, _, _ := adminContentDataFixture()

	data, err := json.Marshal(acd)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"admin_content_data_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id", "admin_route_id",
		"admin_datatype_id", "author_id", "status",
		"date_created", "date_modified",
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

func TestCreateAdminContentDataParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	p := CreateAdminContentDataParams{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("route-json"), Valid: true},
		Status:       types.ContentStatus("draft"),
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
		"parent_id", "first_child_id", "next_sibling_id",
		"prev_sibling_id", "admin_route_id", "admin_datatype_id",
		"author_id", "status", "date_created", "date_modified",
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

func TestUpdateAdminContentDataParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	contentID := types.NewAdminContentID()
	p := UpdateAdminContentDataParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("route-json-update"), Valid: true},
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
		AdminContentDataID: contentID,
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
		"parent_id", "first_child_id", "next_sibling_id",
		"prev_sibling_id", "admin_route_id", "admin_datatype_id",
		"author_id", "status", "date_created", "date_modified",
		"admin_content_data_id",
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

// --- MapAdminContentDataJSON: NullString edge cases ---
// Verifies that sql.NullString with Valid=true but empty String is handled correctly.

func TestMapAdminContentDataJSON_ValidButEmptyNullString(t *testing.T) {
	t.Parallel()
	// sql.NullString{String: "", Valid: true} means the column is NOT NULL
	// but contains an empty string. MapAdminContentDataJSON should produce "".
	acd := AdminContentData{
		AdminContentDataID: types.NewAdminContentID(),
		FirstChildID:       sql.NullString{String: "", Valid: true},
		NextSiblingID:      sql.NullString{String: "", Valid: true},
		PrevSiblingID:      sql.NullString{String: "", Valid: true},
		Status:             types.ContentStatus("draft"),
	}

	got := MapAdminContentDataJSON(acd)

	if got.FirstChildID != "" {
		t.Errorf("FirstChildID = %q, want empty string for Valid=true empty String", got.FirstChildID)
	}
	if got.NextSiblingID != "" {
		t.Errorf("NextSiblingID = %q, want empty string for Valid=true empty String", got.NextSiblingID)
	}
	if got.PrevSiblingID != "" {
		t.Errorf("PrevSiblingID = %q, want empty string for Valid=true empty String", got.PrevSiblingID)
	}
}

// --- MapAdminContentDataJSON: all status variants ---

func TestMapAdminContentDataJSON_StatusVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status types.ContentStatus
	}{
		{"draft", types.ContentStatus("draft")},
		{"published", types.ContentStatus("published")},
		{"archived", types.ContentStatus("archived")},
		{"pending", types.ContentStatus("pending")},
		{"empty", types.ContentStatus("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			acd := AdminContentData{Status: tt.status}
			got := MapAdminContentDataJSON(acd)
			if got.Status != string(tt.status) {
				t.Errorf("Status = %q, want %q", got.Status, string(tt.status))
			}
		})
	}
}
