// Integration tests for the content_field entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires ContentData created in-test, plus Route, Field, User seeds).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_ContentField(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	routeID := types.NullableRouteID{ID: seed.Route.RouteID, Valid: true}
	datatypeID := types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true}
	cdAuthorID := seed.User.UserID // ContentData uses types.UserID (non-nullable)
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}
	fieldID := types.NullableFieldID{ID: seed.Field.FieldID, Valid: true}

	// --- Create prerequisite ContentData record ---
	contentData, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      cdAuthorID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("prerequisite CreateContentData: %v", err)
	}
	contentDataID := types.NullableContentID{ID: contentData.ContentDataID, Valid: true}

	// --- Count: starts at zero ---
	count, err := d.CountContentFields()
	if err != nil {
		t.Fatalf("initial CountContentFields: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountContentFields = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       routeID,
		ContentDataID: contentDataID,
		FieldID:       fieldID,
		FieldValue:    "test field value",
		AuthorID:      authorID,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentField: %v", err)
	}
	if created == nil {
		t.Fatal("CreateContentField returned nil")
	}
	if created.ContentFieldID.IsZero() {
		t.Fatal("CreateContentField returned zero ContentFieldID")
	}
	if created.FieldValue != "test field value" {
		t.Errorf("FieldValue = %q, want %q", created.FieldValue, "test field value")
	}
	if created.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", created.RouteID, routeID)
	}
	if created.ContentDataID != contentDataID {
		t.Errorf("ContentDataID = %v, want %v", created.ContentDataID, contentDataID)
	}
	if created.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", created.FieldID, fieldID)
	}
	if created.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", created.AuthorID, authorID)
	}

	// --- Get ---
	got, err := d.GetContentField(created.ContentFieldID)
	if err != nil {
		t.Fatalf("GetContentField: %v", err)
	}
	if got == nil {
		t.Fatal("GetContentField returned nil")
	}
	if got.ContentFieldID != created.ContentFieldID {
		t.Errorf("GetContentField ID = %v, want %v", got.ContentFieldID, created.ContentFieldID)
	}
	if got.FieldValue != created.FieldValue {
		t.Errorf("GetContentField FieldValue = %q, want %q", got.FieldValue, created.FieldValue)
	}

	// --- List ---
	list, err := d.ListContentFields()
	if err != nil {
		t.Fatalf("ListContentFields: %v", err)
	}
	if list == nil {
		t.Fatal("ListContentFields returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListContentFields len = %d, want 1", len(*list))
	}
	if (*list)[0].ContentFieldID != created.ContentFieldID {
		t.Errorf("ListContentFields[0].ContentFieldID = %v, want %v", (*list)[0].ContentFieldID, created.ContentFieldID)
	}

	// --- ListContentFieldsByRoute ---
	byRoute, err := d.ListContentFieldsByRoute(routeID)
	if err != nil {
		t.Fatalf("ListContentFieldsByRoute: %v", err)
	}
	if byRoute == nil {
		t.Fatal("ListContentFieldsByRoute returned nil")
	}
	if len(*byRoute) != 1 {
		t.Fatalf("ListContentFieldsByRoute len = %d, want 1", len(*byRoute))
	}
	if (*byRoute)[0].ContentFieldID != created.ContentFieldID {
		t.Errorf("ListContentFieldsByRoute[0].ContentFieldID = %v, want %v", (*byRoute)[0].ContentFieldID, created.ContentFieldID)
	}

	// --- ListContentFieldsByContentData ---
	byContentData, err := d.ListContentFieldsByContentData(contentDataID)
	if err != nil {
		t.Fatalf("ListContentFieldsByContentData: %v", err)
	}
	if byContentData == nil {
		t.Fatal("ListContentFieldsByContentData returned nil")
	}
	if len(*byContentData) != 1 {
		t.Fatalf("ListContentFieldsByContentData len = %d, want 1", len(*byContentData))
	}
	if (*byContentData)[0].ContentFieldID != created.ContentFieldID {
		t.Errorf("ListContentFieldsByContentData[0].ContentFieldID = %v, want %v", (*byContentData)[0].ContentFieldID, created.ContentFieldID)
	}

	// --- Count: now 1 ---
	count, err = d.CountContentFields()
	if err != nil {
		t.Fatalf("CountContentFields after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountContentFields after create = %d, want 1", *count)
	}

	// --- Update ---
	updateResult, err := d.UpdateContentField(ctx, ac, UpdateContentFieldParams{
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "updated field value",
		AuthorID:       authorID,
		DateCreated:    now,
		DateModified:   types.TimestampNow(),
		ContentFieldID: created.ContentFieldID,
	})
	if err != nil {
		t.Fatalf("UpdateContentField: %v", err)
	}
	if updateResult == nil {
		t.Error("UpdateContentField returned nil message (expected success message)")
	}

	// --- Get after update ---
	updated, err := d.GetContentField(created.ContentFieldID)
	if err != nil {
		t.Fatalf("GetContentField after update: %v", err)
	}
	if updated.FieldValue != "updated field value" {
		t.Errorf("updated FieldValue = %q, want %q", updated.FieldValue, "updated field value")
	}

	// --- Delete ---
	err = d.DeleteContentField(ctx, ac, created.ContentFieldID)
	if err != nil {
		t.Fatalf("DeleteContentField: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetContentField(created.ContentFieldID)
	if err == nil {
		t.Fatal("GetContentField after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountContentFields()
	if err != nil {
		t.Fatalf("CountContentFields after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountContentFields after delete = %d, want 0", *count)
	}
}
