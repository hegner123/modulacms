// Integration tests for the admin_content_field entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires AdminContentData created in-test, plus AdminRoute, AdminField, User seeds).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminContentField(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	adminRouteID := types.NullableAdminRouteID{ID: seed.AdminRoute.AdminRouteID, Valid: true}
	adminDatatypeID := types.NullableAdminDatatypeID{ID: seed.AdminDatatype.AdminDatatypeID, Valid: true}
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}
	adminFieldID := types.NullableAdminFieldID{ID: seed.AdminField.AdminFieldID, Valid: true}

	// --- Create prerequisite AdminContentData record ---
	adminContentData, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{
		ParentID:        types.NullableAdminContentID{},
		FirstChildID:    types.NullableAdminContentID{},
		NextSiblingID:   types.NullableAdminContentID{},
		PrevSiblingID:   types.NullableAdminContentID{},
		AdminRouteID:    adminRouteID,
		AdminDatatypeID: adminDatatypeID,
		AuthorID:        authorID,
		Status:          types.ContentStatusDraft,
		DateCreated:     now,
		DateModified:    now,
	})
	if err != nil {
		t.Fatalf("prerequisite CreateAdminContentData: %v", err)
	}
	adminContentDataID := types.NullableAdminContentID{ID: adminContentData.AdminContentDataID, Valid: true}

	// --- Count: starts at zero ---
	count, err := d.CountAdminContentFields()
	if err != nil {
		t.Fatalf("initial CountAdminContentFields: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountAdminContentFields = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminContentField(ctx, ac, CreateAdminContentFieldParams{
		AdminRouteID:       adminRouteID,
		AdminContentDataID: adminContentDataID,
		AdminFieldID:       adminFieldID,
		AdminFieldValue:    "test admin field value",
		AuthorID:           authorID,
		DateCreated:        now,
		DateModified:       now,
	})
	if err != nil {
		t.Fatalf("CreateAdminContentField: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminContentField returned nil")
	}
	if created.AdminContentFieldID.IsZero() {
		t.Fatal("CreateAdminContentField returned zero AdminContentFieldID")
	}
	if created.AdminFieldValue != "test admin field value" {
		t.Errorf("AdminFieldValue = %q, want %q", created.AdminFieldValue, "test admin field value")
	}
	if created.AdminContentDataID != adminContentDataID {
		t.Errorf("AdminContentDataID = %v, want %v", created.AdminContentDataID, adminContentDataID)
	}
	if created.AdminFieldID != adminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", created.AdminFieldID, adminFieldID)
	}
	if created.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", created.AuthorID, authorID)
	}

	// --- Get ---
	got, err := d.GetAdminContentField(created.AdminContentFieldID)
	if err != nil {
		t.Fatalf("GetAdminContentField: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminContentField returned nil")
	}
	if got.AdminContentFieldID != created.AdminContentFieldID {
		t.Errorf("GetAdminContentField ID = %v, want %v", got.AdminContentFieldID, created.AdminContentFieldID)
	}
	if got.AdminFieldValue != created.AdminFieldValue {
		t.Errorf("GetAdminContentField AdminFieldValue = %q, want %q", got.AdminFieldValue, created.AdminFieldValue)
	}

	// --- List ---
	list, err := d.ListAdminContentFields()
	if err != nil {
		t.Fatalf("ListAdminContentFields: %v", err)
	}
	if list == nil {
		t.Fatal("ListAdminContentFields returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListAdminContentFields len = %d, want 1", len(*list))
	}
	if (*list)[0].AdminContentFieldID != created.AdminContentFieldID {
		t.Errorf("ListAdminContentFields[0].AdminContentFieldID = %v, want %v", (*list)[0].AdminContentFieldID, created.AdminContentFieldID)
	}

	// --- ListAdminContentFieldsByRoute ---
	byRoute, err := d.ListAdminContentFieldsByRoute(adminRouteID)
	if err != nil {
		t.Fatalf("ListAdminContentFieldsByRoute: %v", err)
	}
	if byRoute == nil {
		t.Fatal("ListAdminContentFieldsByRoute returned nil")
	}
	if len(*byRoute) != 1 {
		t.Fatalf("ListAdminContentFieldsByRoute len = %d, want 1", len(*byRoute))
	}
	if (*byRoute)[0].AdminContentFieldID != created.AdminContentFieldID {
		t.Errorf("ListAdminContentFieldsByRoute[0].AdminContentFieldID = %v, want %v", (*byRoute)[0].AdminContentFieldID, created.AdminContentFieldID)
	}

	// --- ListAdminContentFieldsByRoute with non-matching route ---
	noMatch, err := d.ListAdminContentFieldsByRoute(types.NullableAdminRouteID{ID: types.AdminRouteID("nonexistent-route-id"), Valid: true})
	if err != nil {
		t.Fatalf("ListAdminContentFieldsByRoute (no match): %v", err)
	}
	if noMatch != nil && len(*noMatch) != 0 {
		t.Errorf("ListAdminContentFieldsByRoute (no match) len = %d, want 0", len(*noMatch))
	}

	// --- Count: now 1 ---
	count, err = d.CountAdminContentFields()
	if err != nil {
		t.Fatalf("CountAdminContentFields after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminContentFields after create = %d, want 1", *count)
	}

	// --- Update ---
	updateResult, err := d.UpdateAdminContentField(ctx, ac, UpdateAdminContentFieldParams{
		AdminRouteID:        adminRouteID,
		AdminContentDataID:  adminContentDataID,
		AdminFieldID:        adminFieldID,
		AdminFieldValue:     "updated admin field value",
		AuthorID:            authorID,
		DateCreated:         now,
		DateModified:        types.TimestampNow(),
		AdminContentFieldID: created.AdminContentFieldID,
	})
	if err != nil {
		t.Fatalf("UpdateAdminContentField: %v", err)
	}
	if updateResult == nil {
		t.Error("UpdateAdminContentField returned nil message (expected success message)")
	}

	// --- Get after update ---
	updated, err := d.GetAdminContentField(created.AdminContentFieldID)
	if err != nil {
		t.Fatalf("GetAdminContentField after update: %v", err)
	}
	if updated.AdminFieldValue != "updated admin field value" {
		t.Errorf("updated AdminFieldValue = %q, want %q", updated.AdminFieldValue, "updated admin field value")
	}

	// --- Delete ---
	err = d.DeleteAdminContentField(ctx, ac, created.AdminContentFieldID)
	if err != nil {
		t.Fatalf("DeleteAdminContentField: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetAdminContentField(created.AdminContentFieldID)
	if err == nil {
		t.Fatal("GetAdminContentField after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountAdminContentFields()
	if err != nil {
		t.Fatalf("CountAdminContentFields after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountAdminContentFields after delete = %d, want 0", *count)
	}
}
