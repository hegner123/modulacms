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

// contentDataFixture returns a fully populated ContentData and its parts for testing.
func contentDataFixture() (ContentData, types.ContentID, types.Timestamp) {
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	cd := ContentData{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "first-child-001", Valid: true},
		NextSiblingID: sql.NullString{String: "next-sibling-001", Valid: true},
		PrevSiblingID: sql.NullString{String: "prev-sibling-001", Valid: true},
		RouteID:       routeID,
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatus("published"),
		DateCreated:   ts,
		DateModified:  ts,
	}
	return cd, contentID, ts
}

// contentDataFixtureNulls returns a ContentData where all nullable fields are null/invalid.
func contentDataFixtureNulls() ContentData {
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	return ContentData{
		ContentDataID: contentID,
		ParentID:      types.NullableContentID{Valid: false},
		FirstChildID:  sql.NullString{Valid: false},
		NextSiblingID: sql.NullString{Valid: false},
		PrevSiblingID: sql.NullString{Valid: false},
		RouteID:       types.NullableRouteID{Valid: false},
		DatatypeID:    types.NullableDatatypeID{Valid: false},
		AuthorID:      types.NullableUserID{Valid: false},
		Status:        types.ContentStatus("draft"),
		DateCreated:   ts,
		DateModified:  ts,
	}
}

// --- MapContentDataJSON tests ---

func TestMapContentDataJSON_AllFieldsValid(t *testing.T) {
	t.Parallel()
	cd, contentID, ts := contentDataFixture()

	got := MapContentDataJSON(cd)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ContentDataID", got.ContentDataID, contentID.String()},
		{"ParentID", got.ParentID, cd.ParentID.String()},
		{"FirstChildID", got.FirstChildID, "first-child-001"},
		{"NextSiblingID", got.NextSiblingID, "next-sibling-001"},
		{"PrevSiblingID", got.PrevSiblingID, "prev-sibling-001"},
		{"RouteID", got.RouteID, cd.RouteID.String()},
		{"DatatypeID", got.DatatypeID, cd.DatatypeID.String()},
		{"AuthorID", got.AuthorID, cd.AuthorID.String()},
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

func TestMapContentDataJSON_NullSiblingFields(t *testing.T) {
	t.Parallel()
	// When FirstChildID, NextSiblingID, PrevSiblingID are null (Valid=false),
	// MapContentDataJSON should produce empty strings for those fields.
	cd := contentDataFixtureNulls()

	got := MapContentDataJSON(cd)

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

func TestMapContentDataJSON_NullableIDsProduceNullString(t *testing.T) {
	t.Parallel()
	// When ParentID, RouteID, DatatypeID, AuthorID are null (Valid=false),
	// their String() method returns "null".
	cd := contentDataFixtureNulls()

	got := MapContentDataJSON(cd)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q for null NullableContentID", got.ParentID, "null")
	}
	if got.RouteID != "null" {
		t.Errorf("RouteID = %q, want %q for null NullableRouteID", got.RouteID, "null")
	}
	if got.DatatypeID != "null" {
		t.Errorf("DatatypeID = %q, want %q for null NullableDatatypeID", got.DatatypeID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null NullableUserID", got.AuthorID, "null")
	}
}

func TestMapContentDataJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value ContentData should not panic
	got := MapContentDataJSON(ContentData{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %q, want empty string for zero ContentID", got.ContentDataID)
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

func TestMapContentDataJSON_ValidButEmptyNullString(t *testing.T) {
	t.Parallel()
	// sql.NullString{String: "", Valid: true} means the column is NOT NULL
	// but contains an empty string. MapContentDataJSON should produce "".
	cd := ContentData{
		ContentDataID: types.NewContentID(),
		FirstChildID:  sql.NullString{String: "", Valid: true},
		NextSiblingID: sql.NullString{String: "", Valid: true},
		PrevSiblingID: sql.NullString{String: "", Valid: true},
		Status:        types.ContentStatus("draft"),
	}

	got := MapContentDataJSON(cd)

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

func TestMapContentDataJSON_StatusVariants(t *testing.T) {
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
			cd := ContentData{Status: tt.status}
			got := MapContentDataJSON(cd)
			if got.Status != string(tt.status) {
				t.Errorf("Status = %q, want %q", got.Status, string(tt.status))
			}
		})
	}
}

// --- MapStringContentData tests ---

func TestMapStringContentData_AllFields(t *testing.T) {
	t.Parallel()
	cd, contentID, ts := contentDataFixture()

	got := MapStringContentData(cd)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ContentDataID", got.ContentDataID, contentID.String()},
		{"ParentID", got.ParentID, cd.ParentID.String()},
		{"FirstChildID", got.FirstChildID, "first-child-001"},
		{"NextSiblingID", got.NextSiblingID, "next-sibling-001"},
		{"PrevSiblingID", got.PrevSiblingID, "prev-sibling-001"},
		{"RouteID", got.RouteID, cd.RouteID.String()},
		{"DatatypeID", got.DatatypeID, cd.DatatypeID.String()},
		{"AuthorID", got.AuthorID, cd.AuthorID.String()},
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

func TestMapStringContentData_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// History field is hard-coded to "" per the comment in the source.
	cd, _, _ := contentDataFixture()

	got := MapStringContentData(cd)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (History field removed)", got.History)
	}
}

func TestMapStringContentData_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringContentData(ContentData{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %q, want empty string", got.ContentDataID)
	}
	if got.RouteID != "null" {
		t.Errorf("RouteID = %q, want %q for null NullableRouteID", got.RouteID, "null")
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
}

func TestMapStringContentData_NullableFieldsShowNull(t *testing.T) {
	t.Parallel()
	cd := contentDataFixtureNulls()

	got := MapStringContentData(cd)

	// Nullable IDs with Valid=false produce "null" via their String() method
	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q for null", got.ParentID, "null")
	}
	if got.RouteID != "null" {
		t.Errorf("RouteID = %q, want %q for null", got.RouteID, "null")
	}
	if got.DatatypeID != "null" {
		t.Errorf("DatatypeID = %q, want %q for null", got.DatatypeID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null", got.AuthorID, "null")
	}
}

func TestMapStringContentData_NullSiblingFieldsEmpty(t *testing.T) {
	t.Parallel()
	cd := contentDataFixtureNulls()

	got := MapStringContentData(cd)

	// sql.NullString with Valid=false should produce empty string
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

// --- SQLite Database.MapContentData tests ---

func TestDatabase_MapContentData_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdb.ContentData{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "child-1", Valid: true},
		NextSiblingID: sql.NullString{String: "next-1", Valid: true},
		PrevSiblingID: sql.NullString{String: "prev-1", Valid: true},
		RouteID:       routeID,
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatus("published"),
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapContentData(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
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
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
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

func TestDatabase_MapContentData_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapContentData(mdb.ContentData{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
	if got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = true, want false for zero value")
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
}

// --- SQLite Database.MapCreateContentDataParams tests ---

func TestDatabase_MapCreateContentDataParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateContentDataParams{
		ParentID:      types.NullableContentID{Valid: false},
		FirstChildID:  sql.NullString{Valid: false},
		NextSiblingID: sql.NullString{Valid: false},
		PrevSiblingID: sql.NullString{Valid: false},
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		DatatypeID:    types.NullableDatatypeID{Valid: false},
		AuthorID:      types.NullableUserID{ID: types.NewUserID(), Valid: true},
		Status:        types.ContentStatus("draft"),
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapCreateContentDataParams(input)

	if got.ContentDataID.IsZero() {
		t.Fatal("expected non-zero ContentDataID to be generated")
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
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

func TestDatabase_MapCreateContentDataParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateContentDataParams{Status: types.ContentStatus("draft")}

	got1 := d.MapCreateContentDataParams(input)
	got2 := d.MapCreateContentDataParams(input)
	if got1.ContentDataID == got2.ContentDataID {
		t.Error("two calls produced the same ContentDataID -- each call should generate a unique ID")
	}
}

func TestDatabase_MapCreateContentDataParams_PreservesNullableFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	parentID := types.NullableContentID{ID: types.ContentID("parent-123"), Valid: true}
	firstChild := sql.NullString{String: "child-abc", Valid: true}
	nextSibling := sql.NullString{String: "next-def", Valid: true}
	prevSibling := sql.NullString{String: "prev-ghi", Valid: true}
	routeID := types.NullableRouteID{ID: types.RouteID("route-456"), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.DatatypeID("dt-789"), Valid: true}

	input := CreateContentDataParams{
		ParentID:      parentID,
		FirstChildID:  firstChild,
		NextSiblingID: nextSibling,
		PrevSiblingID: prevSibling,
		RouteID:       routeID,
		DatatypeID:    datatypeID,
	}

	got := d.MapCreateContentDataParams(input)

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
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
}

// --- SQLite Database.MapUpdateContentDataParams tests ---

func TestDatabase_MapUpdateContentDataParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	contentID := types.NewContentID()
	parentID := types.NullableContentID{ID: types.ContentID("parent-updated"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}

	input := UpdateContentDataParams{
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "updated-child", Valid: true},
		NextSiblingID: sql.NullString{String: "updated-next", Valid: true},
		PrevSiblingID: sql.NullString{String: "updated-prev", Valid: true},
		RouteID:       routeID,
		DatatypeID:    types.NullableDatatypeID{Valid: false},
		AuthorID:      authorID,
		Status:        types.ContentStatus("archived"),
		DateCreated:   ts,
		DateModified:  ts,
		ContentDataID: contentID,
	}

	got := d.MapUpdateContentDataParams(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID.String != "updated-child" {
		t.Errorf("FirstChildID.String = %q, want %q", got.FirstChildID.String, "updated-child")
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
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

// --- MySQL MysqlDatabase.MapContentData tests ---

func TestMysqlDatabase_MapContentData_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbm.ContentData{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "mysql-child", Valid: true},
		NextSiblingID: sql.NullString{String: "mysql-next", Valid: true},
		PrevSiblingID: sql.NullString{String: "mysql-prev", Valid: true},
		RouteID:       routeID,
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatus("published"),
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapContentData(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
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

func TestMysqlDatabase_MapContentData_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapContentData(mdbm.ContentData{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
	if got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = true, want false for zero value")
	}
}

// --- MySQL MysqlDatabase.MapCreateContentDataParams tests ---

func TestMysqlDatabase_MapCreateContentDataParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateContentDataParams(input)

	if got.ContentDataID.IsZero() {
		t.Fatal("expected non-zero ContentDataID to be generated")
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.Status != types.ContentStatus("draft") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("draft"))
	}
}

// --- MySQL MysqlDatabase.MapUpdateContentDataParams tests ---

func TestMysqlDatabase_MapUpdateContentDataParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	contentID := types.NewContentID()

	input := UpdateContentDataParams{
		ParentID:      types.NullableContentID{Valid: false},
		FirstChildID:  sql.NullString{String: "mysql-updated-child", Valid: true},
		NextSiblingID: sql.NullString{Valid: false},
		PrevSiblingID: sql.NullString{Valid: false},
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		DatatypeID:    types.NullableDatatypeID{Valid: false},
		AuthorID:      types.NullableUserID{Valid: false},
		Status:        types.ContentStatus("archived"),
		DateCreated:   ts,
		DateModified:  ts,
		ContentDataID: contentID,
	}

	got := d.MapUpdateContentDataParams(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.Status != types.ContentStatus("archived") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("archived"))
	}
	if got.FirstChildID.String != "mysql-updated-child" {
		t.Errorf("FirstChildID.String = %q, want %q", got.FirstChildID.String, "mysql-updated-child")
	}
}

// --- PostgreSQL PsqlDatabase.MapContentData tests ---

func TestPsqlDatabase_MapContentData_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbp.ContentData{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "psql-child", Valid: true},
		NextSiblingID: sql.NullString{String: "psql-next", Valid: true},
		PrevSiblingID: sql.NullString{String: "psql-prev", Valid: true},
		RouteID:       routeID,
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatus("published"),
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapContentData(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
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

func TestPsqlDatabase_MapContentData_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapContentData(mdbp.ContentData{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
	if got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = true, want false for zero value")
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateContentDataParams tests ---

func TestPsqlDatabase_MapCreateContentDataParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateContentDataParams(input)

	if got.ContentDataID.IsZero() {
		t.Fatal("expected non-zero ContentDataID to be generated")
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.Status != types.ContentStatus("draft") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("draft"))
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateContentDataParams tests ---

func TestPsqlDatabase_MapUpdateContentDataParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	contentID := types.NewContentID()

	input := UpdateContentDataParams{
		ParentID:      types.NullableContentID{Valid: false},
		FirstChildID:  sql.NullString{String: "psql-updated-child", Valid: true},
		NextSiblingID: sql.NullString{Valid: false},
		PrevSiblingID: sql.NullString{Valid: false},
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		DatatypeID:    types.NullableDatatypeID{Valid: false},
		AuthorID:      types.NullableUserID{Valid: false},
		Status:        types.ContentStatus("pending"),
		DateCreated:   ts,
		DateModified:  ts,
		ContentDataID: contentID,
	}

	got := d.MapUpdateContentDataParams(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.Status != types.ContentStatus("pending") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("pending"))
	}
	if got.FirstChildID.String != "psql-updated-child" {
		t.Errorf("FirstChildID.String = %q, want %q", got.FirstChildID.String, "psql-updated-child")
	}
}

// --- Cross-database mapper consistency ---
// Verifies that all three database mappers produce identical ContentData from equivalent input.

func TestCrossDatabaseMapContentData_Consistency(t *testing.T) {
	t.Parallel()
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.ContentID("parent-cross"), Valid: true}
	routeID := types.NullableRouteID{ID: types.RouteID("route-cross"), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.DatatypeID("dt-cross"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	firstChild := sql.NullString{String: "cross-child", Valid: true}
	nextSibling := sql.NullString{String: "cross-next", Valid: true}
	prevSibling := sql.NullString{String: "cross-prev", Valid: true}

	sqliteInput := mdb.ContentData{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		RouteID: routeID, DatatypeID: datatypeID, AuthorID: authorID,
		Status: types.ContentStatus("published"), DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.ContentData{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		RouteID: routeID, DatatypeID: datatypeID, AuthorID: authorID,
		Status: types.ContentStatus("published"), DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.ContentData{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		RouteID: routeID, DatatypeID: datatypeID, AuthorID: authorID,
		Status: types.ContentStatus("published"), DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapContentData(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapContentData(mysqlInput)
	psqlResult := PsqlDatabase{}.MapContentData(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateContentDataParams consistency ---
// All three should auto-generate unique IDs.

func TestCrossDatabaseMapCreateContentDataParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateContentDataParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateContentDataParams(input)
	psqlResult := PsqlDatabase{}.MapCreateContentDataParams(input)

	if sqliteResult.ContentDataID.IsZero() {
		t.Error("SQLite: expected non-zero generated ContentDataID")
	}
	if mysqlResult.ContentDataID.IsZero() {
		t.Error("MySQL: expected non-zero generated ContentDataID")
	}
	if psqlResult.ContentDataID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated ContentDataID")
	}

	// Each call should generate a unique ID
	if sqliteResult.ContentDataID == mysqlResult.ContentDataID {
		t.Error("SQLite and MySQL generated the same ContentDataID -- each call should be unique")
	}
	if sqliteResult.ContentDataID == psqlResult.ContentDataID {
		t.Error("SQLite and PostgreSQL generated the same ContentDataID -- each call should be unique")
	}
}

// --- RootContentSummary mapper tests ---

func TestDatabase_MapRootContentSummary_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	contentID := types.NewContentID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC))

	input := mdb.ListRootContentSummaryRow{
		ContentDataID: contentID,
		RouteID:       routeID,
		DatatypeID:    datatypeID,
		RouteSlug:     types.Slug("root-page"),
		RouteTitle:    "Root Page",
		DatatypeLabel: "Article",
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapRootContentSummary(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.RouteSlug != types.Slug("root-page") {
		t.Errorf("RouteSlug = %v, want %v", got.RouteSlug, types.Slug("root-page"))
	}
	if got.RouteTitle != "Root Page" {
		t.Errorf("RouteTitle = %q, want %q", got.RouteTitle, "Root Page")
	}
	if got.DatatypeLabel != "Article" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Article")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestMysqlDatabase_MapRootContentSummary_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	contentID := types.NewContentID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC))

	input := mdbm.ListRootContentSummaryRow{
		ContentDataID: contentID,
		RouteID:       routeID,
		DatatypeID:    datatypeID,
		RouteSlug:     types.Slug("mysql-root"),
		RouteTitle:    "MySQL Root",
		DatatypeLabel: "Page",
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapRootContentSummary(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.RouteSlug != types.Slug("mysql-root") {
		t.Errorf("RouteSlug = %v, want %v", got.RouteSlug, types.Slug("mysql-root"))
	}
	if got.RouteTitle != "MySQL Root" {
		t.Errorf("RouteTitle = %q, want %q", got.RouteTitle, "MySQL Root")
	}
	if got.DatatypeLabel != "Page" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Page")
	}
}

func TestPsqlDatabase_MapRootContentSummary_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	contentID := types.NewContentID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC))

	input := mdbp.ListRootContentSummaryRow{
		ContentDataID: contentID,
		RouteID:       routeID,
		DatatypeID:    datatypeID,
		RouteSlug:     types.Slug("psql-root"),
		RouteTitle:    "PSQL Root",
		DatatypeLabel: "Blog",
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapRootContentSummary(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.RouteSlug != types.Slug("psql-root") {
		t.Errorf("RouteSlug = %v, want %v", got.RouteSlug, types.Slug("psql-root"))
	}
	if got.RouteTitle != "PSQL Root" {
		t.Errorf("RouteTitle = %q, want %q", got.RouteTitle, "PSQL Root")
	}
	if got.DatatypeLabel != "Blog" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Blog")
	}
}

// Cross-database RootContentSummary consistency
func TestCrossDatabaseMapRootContentSummary_Consistency(t *testing.T) {
	t.Parallel()
	contentID := types.NewContentID()
	routeID := types.NullableRouteID{ID: types.RouteID("route-cross"), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.DatatypeID("dt-cross"), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC))

	sqliteInput := mdb.ListRootContentSummaryRow{
		ContentDataID: contentID, RouteID: routeID, DatatypeID: datatypeID,
		RouteSlug: types.Slug("cross"), RouteTitle: "Cross", DatatypeLabel: "Type",
		DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.ListRootContentSummaryRow{
		ContentDataID: contentID, RouteID: routeID, DatatypeID: datatypeID,
		RouteSlug: types.Slug("cross"), RouteTitle: "Cross", DatatypeLabel: "Type",
		DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.ListRootContentSummaryRow{
		ContentDataID: contentID, RouteID: routeID, DatatypeID: datatypeID,
		RouteSlug: types.Slug("cross"), RouteTitle: "Cross", DatatypeLabel: "Type",
		DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapRootContentSummary(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapRootContentSummary(mysqlInput)
	psqlResult := PsqlDatabase{}.MapRootContentSummary(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different RootContentSummary:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different RootContentSummary:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewContentDataCmd_AllAccessors(t *testing.T) {
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
	params := CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:       types.ContentStatus("draft"),
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
	}
	p, ok := cmd.Params().(CreateContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentDataParams", cmd.Params())
	}
	if p.Status != types.ContentStatus("draft") {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.ContentStatus("draft"))
	}
	if p.RouteID != params.RouteID {
		t.Errorf("Params().RouteID = %v, want %v", p.RouteID, params.RouteID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewContentDataCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	contentID := types.NewContentID()
	cmd := NewContentDataCmd{}

	row := mdb.ContentData{ContentDataID: contentID}
	got := cmd.GetID(row)
	if got != string(contentID) {
		t.Errorf("GetID() = %q, want %q", got, string(contentID))
	}
}

func TestNewContentDataCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewContentDataCmd{}

	row := mdb.ContentData{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateContentDataCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	contentID := types.NewContentID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateContentDataParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:        types.ContentStatus("published"),
		DateCreated:   ts,
		DateModified:  ts,
		ContentDataID: contentID,
	}

	cmd := Database{}.UpdateContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
	}
	if cmd.GetID() != string(contentID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(contentID))
	}
	p, ok := cmd.Params().(UpdateContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentDataParams", cmd.Params())
	}
	if p.Status != types.ContentStatus("published") {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.ContentStatus("published"))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentDataCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	contentID := types.NewContentID()

	cmd := Database{}.DeleteContentDataCmd(ctx, ac, contentID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
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

func TestNewContentDataCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
	}
	p, ok := cmd.Params().(CreateContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentDataParams", cmd.Params())
	}
	if p.Status != types.ContentStatus("draft") {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.ContentStatus("draft"))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewContentDataCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	contentID := types.NewContentID()
	cmd := NewContentDataCmdMysql{}

	row := mdbm.ContentData{ContentDataID: contentID}
	got := cmd.GetID(row)
	if got != string(contentID) {
		t.Errorf("GetID() = %q, want %q", got, string(contentID))
	}
}

func TestUpdateContentDataCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	contentID := types.NewContentID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateContentDataParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:        types.ContentStatus("archived"),
		DateCreated:   ts,
		DateModified:  ts,
		ContentDataID: contentID,
	}

	cmd := MysqlDatabase{}.UpdateContentDataCmd(ctx, ac, params)

	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
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
	p, ok := cmd.Params().(UpdateContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentDataParams", cmd.Params())
	}
	if p.Status != types.ContentStatus("archived") {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.ContentStatus("archived"))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentDataCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	contentID := types.NewContentID()

	cmd := MysqlDatabase{}.DeleteContentDataCmd(ctx, ac, contentID)

	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
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

func TestNewContentDataCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:       types.ContentStatus("draft"),
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewContentDataCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
	}
	p, ok := cmd.Params().(CreateContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentDataParams", cmd.Params())
	}
	if p.Status != types.ContentStatus("draft") {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.ContentStatus("draft"))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewContentDataCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	contentID := types.NewContentID()
	cmd := NewContentDataCmdPsql{}

	row := mdbp.ContentData{ContentDataID: contentID}
	got := cmd.GetID(row)
	if got != string(contentID) {
		t.Errorf("GetID() = %q, want %q", got, string(contentID))
	}
}

func TestUpdateContentDataCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	contentID := types.NewContentID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateContentDataParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:        types.ContentStatus("published"),
		DateCreated:   ts,
		DateModified:  ts,
		ContentDataID: contentID,
	}

	cmd := PsqlDatabase{}.UpdateContentDataCmd(ctx, ac, params)

	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
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
	p, ok := cmd.Params().(UpdateContentDataParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentDataParams", cmd.Params())
	}
	if p.Status != types.ContentStatus("published") {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.ContentStatus("published"))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentDataCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	contentID := types.NewContentID()

	cmd := PsqlDatabase{}.DeleteContentDataCmd(ctx, ac, contentID)

	if cmd.TableName() != "content_data" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_data")
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

func TestAuditedContentDataCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateContentDataParams{}
	updateParams := UpdateContentDataParams{ContentDataID: types.NewContentID()}
	contentID := types.NewContentID()

	// SQLite
	sqliteCreate := Database{}.NewContentDataCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateContentDataCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteContentDataCmd(ctx, ac, contentID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewContentDataCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateContentDataCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteContentDataCmd(ctx, ac, contentID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewContentDataCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateContentDataCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteContentDataCmd(ctx, ac, contentID)

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
			if c.name != "content_data" {
				t.Errorf("TableName() = %q, want %q", c.name, "content_data")
			}
		})
	}
}

func TestAuditedContentDataCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateContentDataParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewContentDataCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewContentDataCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewContentDataCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedContentDataCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	contentID := types.NewContentID()

	t.Run("UpdateCmd GetID returns ContentDataID", func(t *testing.T) {
		t.Parallel()
		params := UpdateContentDataParams{ContentDataID: contentID}

		sqliteCmd := Database{}.UpdateContentDataCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateContentDataCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateContentDataCmd(ctx, ac, params)

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

	t.Run("DeleteCmd GetID returns ContentDataID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteContentDataCmd(ctx, ac, contentID)
		mysqlCmd := MysqlDatabase{}.DeleteContentDataCmd(ctx, ac, contentID)
		psqlCmd := PsqlDatabase{}.DeleteContentDataCmd(ctx, ac, contentID)

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
		testContentID := types.NewContentID()

		sqliteCmd := NewContentDataCmd{}
		mysqlCmd := NewContentDataCmdMysql{}
		psqlCmd := NewContentDataCmdPsql{}

		wantID := string(testContentID)

		sqliteRow := mdb.ContentData{ContentDataID: testContentID}
		mysqlRow := mdbm.ContentData{ContentDataID: testContentID}
		psqlRow := mdbp.ContentData{ContentDataID: testContentID}

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

// --- Edge case: UpdateContentDataCmd with empty ContentDataID ---

func TestUpdateContentDataCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateContentDataParams{ContentDataID: ""}

	sqliteCmd := Database{}.UpdateContentDataCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateContentDataCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateContentDataCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteContentDataCmd with empty ID ---

func TestDeleteContentDataCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.ContentID("")

	sqliteCmd := Database{}.DeleteContentDataCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteContentDataCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteContentDataCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.ContentData]  = NewContentDataCmd{}
	_ audited.UpdateCommand[mdb.ContentData]  = UpdateContentDataCmd{}
	_ audited.DeleteCommand[mdb.ContentData]  = DeleteContentDataCmd{}
	_ audited.CreateCommand[mdbm.ContentData] = NewContentDataCmdMysql{}
	_ audited.UpdateCommand[mdbm.ContentData] = UpdateContentDataCmdMysql{}
	_ audited.DeleteCommand[mdbm.ContentData] = DeleteContentDataCmdMysql{}
	_ audited.CreateCommand[mdbp.ContentData] = NewContentDataCmdPsql{}
	_ audited.UpdateCommand[mdbp.ContentData] = UpdateContentDataCmdPsql{}
	_ audited.DeleteCommand[mdbp.ContentData] = DeleteContentDataCmdPsql{}
)

// --- Struct field correctness ---
// Verify that the wrapper ContentData struct and param structs hold values correctly via JSON.

func TestContentDataStruct_JSONTags(t *testing.T) {
	t.Parallel()
	cd, _, _ := contentDataFixture()

	data, err := json.Marshal(cd)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"content_data_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id", "route_id",
		"datatype_id", "author_id", "status",
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

func TestCreateContentDataParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	p := CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
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
		"route_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id", "datatype_id",
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

func TestUpdateContentDataParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	contentID := types.NewContentID()
	p := UpdateContentDataParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		Status:        types.ContentStatus("published"),
		DateCreated:   ts,
		DateModified:  ts,
		ContentDataID: contentID,
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
		"route_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id", "datatype_id",
		"author_id", "status", "date_created", "date_modified",
		"content_data_id",
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

func TestContentDataJSONStruct_JSONTags(t *testing.T) {
	t.Parallel()
	cdj := ContentDataJSON{
		ContentDataID: "id-1",
		ParentID:      "parent-1",
		FirstChildID:  "child-1",
		NextSiblingID: "next-1",
		PrevSiblingID: "prev-1",
		RouteID:       "route-1",
		DatatypeID:    "dt-1",
		AuthorID:      "author-1",
		Status:        "published",
		DateCreated:   "2025-01-01T00:00:00Z",
		DateModified:  "2025-01-01T00:00:00Z",
	}

	data, err := json.Marshal(cdj)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"content_data_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id", "route_id",
		"datatype_id", "author_id", "status",
		"date_created", "date_modified",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("ContentDataJSON JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("ContentDataJSON JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestRootContentSummaryStruct_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	rcs := RootContentSummary{
		ContentDataID: types.NewContentID(),
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		DatatypeID:    types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true},
		RouteSlug:     types.Slug("test-slug"),
		RouteTitle:    "Test Title",
		DatatypeLabel: "Article",
		DateCreated:   ts,
		DateModified:  ts,
	}

	data, err := json.Marshal(rcs)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"content_data_id", "route_id", "datatype_id",
		"route_slug", "route_title", "datatype_label",
		"date_created", "date_modified",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("RootContentSummary JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("RootContentSummary JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

// --- MapContentDataJSON and MapStringContentData consistency ---
// Both functions process the same input; verify sibling pointer fields are handled identically.

func TestMapContentDataJSON_MapStringContentData_SiblingFieldConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		firstChild    sql.NullString
		nextSibling   sql.NullString
		prevSibling   sql.NullString
		wantFirstJSON string
		wantNextJSON  string
		wantPrevJSON  string
	}{
		{
			name:          "all valid",
			firstChild:    sql.NullString{String: "fc-1", Valid: true},
			nextSibling:   sql.NullString{String: "ns-1", Valid: true},
			prevSibling:   sql.NullString{String: "ps-1", Valid: true},
			wantFirstJSON: "fc-1",
			wantNextJSON:  "ns-1",
			wantPrevJSON:  "ps-1",
		},
		{
			name:          "all null",
			firstChild:    sql.NullString{Valid: false},
			nextSibling:   sql.NullString{Valid: false},
			prevSibling:   sql.NullString{Valid: false},
			wantFirstJSON: "",
			wantNextJSON:  "",
			wantPrevJSON:  "",
		},
		{
			name:          "mixed",
			firstChild:    sql.NullString{String: "fc-mix", Valid: true},
			nextSibling:   sql.NullString{Valid: false},
			prevSibling:   sql.NullString{String: "ps-mix", Valid: true},
			wantFirstJSON: "fc-mix",
			wantNextJSON:  "",
			wantPrevJSON:  "ps-mix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cd := ContentData{
				ContentDataID: types.NewContentID(),
				FirstChildID:  tt.firstChild,
				NextSiblingID: tt.nextSibling,
				PrevSiblingID: tt.prevSibling,
			}

			jsonResult := MapContentDataJSON(cd)
			stringResult := MapStringContentData(cd)

			// Both should produce the same values for sibling fields
			if jsonResult.FirstChildID != tt.wantFirstJSON {
				t.Errorf("JSON FirstChildID = %q, want %q", jsonResult.FirstChildID, tt.wantFirstJSON)
			}
			if stringResult.FirstChildID != tt.wantFirstJSON {
				t.Errorf("String FirstChildID = %q, want %q", stringResult.FirstChildID, tt.wantFirstJSON)
			}
			if jsonResult.NextSiblingID != tt.wantNextJSON {
				t.Errorf("JSON NextSiblingID = %q, want %q", jsonResult.NextSiblingID, tt.wantNextJSON)
			}
			if stringResult.NextSiblingID != tt.wantNextJSON {
				t.Errorf("String NextSiblingID = %q, want %q", stringResult.NextSiblingID, tt.wantNextJSON)
			}
			if jsonResult.PrevSiblingID != tt.wantPrevJSON {
				t.Errorf("JSON PrevSiblingID = %q, want %q", jsonResult.PrevSiblingID, tt.wantPrevJSON)
			}
			if stringResult.PrevSiblingID != tt.wantPrevJSON {
				t.Errorf("String PrevSiblingID = %q, want %q", stringResult.PrevSiblingID, tt.wantPrevJSON)
			}
		})
	}
}
