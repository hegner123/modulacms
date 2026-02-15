// Integration tests for the admin_content_data entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires AdminRoute, AdminDatatype, User seed records).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminContentData(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	adminRouteID := types.NullableAdminRouteID{ID: seed.AdminRoute.AdminRouteID, Valid: true}
	adminDatatypeID := types.NullableAdminDatatypeID{ID: seed.AdminDatatype.AdminDatatypeID, Valid: true}
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at zero ---
	count, err := d.CountAdminContentData()
	if err != nil {
		t.Fatalf("initial CountAdminContentData: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountAdminContentData = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{
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
		t.Fatalf("CreateAdminContentData: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminContentData returned nil")
	}
	if created.AdminContentDataID.IsZero() {
		t.Fatal("CreateAdminContentData returned zero AdminContentDataID")
	}
	if created.Status != types.ContentStatusDraft {
		t.Errorf("Status = %q, want %q", created.Status, types.ContentStatusDraft)
	}
	if created.AdminRouteID != adminRouteID {
		t.Errorf("AdminRouteID = %q, want %q", created.AdminRouteID, adminRouteID)
	}
	if created.AdminDatatypeID != adminDatatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", created.AdminDatatypeID, adminDatatypeID)
	}
	if created.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", created.AuthorID, authorID)
	}

	// --- Get ---
	got, err := d.GetAdminContentData(created.AdminContentDataID)
	if err != nil {
		t.Fatalf("GetAdminContentData: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminContentData returned nil")
	}
	if got.AdminContentDataID != created.AdminContentDataID {
		t.Errorf("GetAdminContentData ID = %v, want %v", got.AdminContentDataID, created.AdminContentDataID)
	}
	if got.Status != created.Status {
		t.Errorf("GetAdminContentData Status = %q, want %q", got.Status, created.Status)
	}

	// --- List ---
	list, err := d.ListAdminContentData()
	if err != nil {
		t.Fatalf("ListAdminContentData: %v", err)
	}
	if list == nil {
		t.Fatal("ListAdminContentData returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListAdminContentData len = %d, want 1", len(*list))
	}
	if (*list)[0].AdminContentDataID != created.AdminContentDataID {
		t.Errorf("ListAdminContentData[0].AdminContentDataID = %v, want %v", (*list)[0].AdminContentDataID, created.AdminContentDataID)
	}

	// --- ListAdminContentDataByRoute ---
	byRoute, err := d.ListAdminContentDataByRoute(adminRouteID.String())
	if err != nil {
		t.Fatalf("ListAdminContentDataByRoute: %v", err)
	}
	if byRoute == nil {
		t.Fatal("ListAdminContentDataByRoute returned nil")
	}
	if len(*byRoute) != 1 {
		t.Fatalf("ListAdminContentDataByRoute len = %d, want 1", len(*byRoute))
	}
	if (*byRoute)[0].AdminContentDataID != created.AdminContentDataID {
		t.Errorf("ListAdminContentDataByRoute[0].AdminContentDataID = %v, want %v", (*byRoute)[0].AdminContentDataID, created.AdminContentDataID)
	}

	// --- ListAdminContentDataByRoute with non-matching route ---
	noMatch, err := d.ListAdminContentDataByRoute("nonexistent-route-id")
	if err != nil {
		t.Fatalf("ListAdminContentDataByRoute (no match): %v", err)
	}
	if noMatch != nil && len(*noMatch) != 0 {
		t.Errorf("ListAdminContentDataByRoute (no match) len = %d, want 0", len(*noMatch))
	}

	// --- Count: now 1 ---
	count, err = d.CountAdminContentData()
	if err != nil {
		t.Fatalf("CountAdminContentData after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminContentData after create = %d, want 1", *count)
	}

	// --- Update ---
	updateResult, err := d.UpdateAdminContentData(ctx, ac, UpdateAdminContentDataParams{
		ParentID:           types.NullableAdminContentID{},
		FirstChildID:       types.NullableAdminContentID{},
		NextSiblingID:      types.NullableAdminContentID{},
		PrevSiblingID:      types.NullableAdminContentID{},
		AdminRouteID:       adminRouteID,
		AdminDatatypeID:    adminDatatypeID,
		AuthorID:           authorID,
		Status:             types.ContentStatusPublished,
		DateCreated:        now,
		DateModified:       types.TimestampNow(),
		AdminContentDataID: created.AdminContentDataID,
	})
	if err != nil {
		t.Fatalf("UpdateAdminContentData: %v", err)
	}
	if updateResult == nil {
		t.Error("UpdateAdminContentData returned nil message (expected success message)")
	}

	// --- Get after update ---
	updated, err := d.GetAdminContentData(created.AdminContentDataID)
	if err != nil {
		t.Fatalf("GetAdminContentData after update: %v", err)
	}
	if updated.Status != types.ContentStatusPublished {
		t.Errorf("updated Status = %q, want %q", updated.Status, types.ContentStatusPublished)
	}

	// --- Delete ---
	err = d.DeleteAdminContentData(ctx, ac, created.AdminContentDataID)
	if err != nil {
		t.Fatalf("DeleteAdminContentData: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetAdminContentData(created.AdminContentDataID)
	if err == nil {
		t.Fatal("GetAdminContentData after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountAdminContentData()
	if err != nil {
		t.Fatalf("CountAdminContentData after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountAdminContentData after delete = %d, want 0", *count)
	}
}
