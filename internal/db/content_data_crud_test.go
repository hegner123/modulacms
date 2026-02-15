// Integration tests for the content_data entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires Route, Datatype, User seed records).
package db

import (
	"database/sql"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_ContentData(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	routeID := types.NullableRouteID{ID: seed.Route.RouteID, Valid: true}
	datatypeID := types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true}
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at zero ---
	count, err := d.CountContentData()
	if err != nil {
		t.Fatalf("initial CountContentData: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountContentData = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  sql.NullString{},
		NextSiblingID: sql.NullString{},
		PrevSiblingID: sql.NullString{},
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}
	if created == nil {
		t.Fatal("CreateContentData returned nil")
	}
	if created.ContentDataID.IsZero() {
		t.Fatal("CreateContentData returned zero ContentDataID")
	}
	if created.Status != types.ContentStatusDraft {
		t.Errorf("Status = %q, want %q", created.Status, types.ContentStatusDraft)
	}
	if created.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", created.RouteID, routeID)
	}
	if created.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", created.DatatypeID, datatypeID)
	}
	if created.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", created.AuthorID, authorID)
	}

	// --- Get ---
	got, err := d.GetContentData(created.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData: %v", err)
	}
	if got == nil {
		t.Fatal("GetContentData returned nil")
	}
	if got.ContentDataID != created.ContentDataID {
		t.Errorf("GetContentData ID = %v, want %v", got.ContentDataID, created.ContentDataID)
	}
	if got.Status != created.Status {
		t.Errorf("GetContentData Status = %q, want %q", got.Status, created.Status)
	}

	// --- List ---
	list, err := d.ListContentData()
	if err != nil {
		t.Fatalf("ListContentData: %v", err)
	}
	if list == nil {
		t.Fatal("ListContentData returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListContentData len = %d, want 1", len(*list))
	}
	if (*list)[0].ContentDataID != created.ContentDataID {
		t.Errorf("ListContentData[0].ContentDataID = %v, want %v", (*list)[0].ContentDataID, created.ContentDataID)
	}

	// --- ListContentDataByRoute ---
	byRoute, err := d.ListContentDataByRoute(routeID)
	if err != nil {
		t.Fatalf("ListContentDataByRoute: %v", err)
	}
	if byRoute == nil {
		t.Fatal("ListContentDataByRoute returned nil")
	}
	if len(*byRoute) != 1 {
		t.Fatalf("ListContentDataByRoute len = %d, want 1", len(*byRoute))
	}
	if (*byRoute)[0].ContentDataID != created.ContentDataID {
		t.Errorf("ListContentDataByRoute[0].ContentDataID = %v, want %v", (*byRoute)[0].ContentDataID, created.ContentDataID)
	}

	// --- ListContentDataByRoute with non-matching route ---
	noMatch, err := d.ListContentDataByRoute(types.NullableRouteID{ID: types.NewRouteID(), Valid: true})
	if err != nil {
		t.Fatalf("ListContentDataByRoute (no match): %v", err)
	}
	if noMatch != nil && len(*noMatch) != 0 {
		t.Errorf("ListContentDataByRoute (no match) len = %d, want 0", len(*noMatch))
	}

	// --- Count: now 1 ---
	count, err = d.CountContentData()
	if err != nil {
		t.Fatalf("CountContentData after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountContentData after create = %d, want 1", *count)
	}

	// --- Update ---
	updateResult, err := d.UpdateContentData(ctx, ac, UpdateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  sql.NullString{},
		NextSiblingID: sql.NullString{},
		PrevSiblingID: sql.NullString{},
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  types.TimestampNow(),
		ContentDataID: created.ContentDataID,
	})
	if err != nil {
		t.Fatalf("UpdateContentData: %v", err)
	}
	if updateResult == nil {
		t.Error("UpdateContentData returned nil message (expected success message)")
	}

	// --- Get after update ---
	updated, err := d.GetContentData(created.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData after update: %v", err)
	}
	if updated.Status != types.ContentStatusPublished {
		t.Errorf("updated Status = %q, want %q", updated.Status, types.ContentStatusPublished)
	}

	// --- Delete ---
	err = d.DeleteContentData(ctx, ac, created.ContentDataID)
	if err != nil {
		t.Fatalf("DeleteContentData: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetContentData(created.ContentDataID)
	if err == nil {
		t.Fatal("GetContentData after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountContentData()
	if err != nil {
		t.Fatalf("CountContentData after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountContentData after delete = %d, want 0", *count)
	}
}
