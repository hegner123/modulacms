// Integration tests for the admin_content_relation entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires 2 AdminContentData records created in-test, plus AdminField, AdminRoute, User seeds).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminContentRelation(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	adminRouteID := new(types.NullableAdminRouteID)
	adminRouteID.ID = seed.AdminRoute.AdminRouteID
	adminRouteID.Valid = true
	adminDatatypeID := types.NullableAdminDatatypeID{ID: seed.AdminDatatype.AdminDatatypeID, Valid: true}
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Create 2 prerequisite AdminContentData records (source and target) ---
	source, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{
		ParentID:        types.NullableAdminContentID{},
		FirstChildID:    types.NullableAdminContentID{},
		NextSiblingID:   types.NullableAdminContentID{},
		PrevSiblingID:   types.NullableAdminContentID{},
		AdminRouteID:    *adminRouteID,
		AdminDatatypeID: adminDatatypeID,
		AuthorID:        authorID,
		Status:          types.ContentStatusDraft,
		DateCreated:     now,
		DateModified:    now,
	})
	if err != nil {
		t.Fatalf("prerequisite CreateAdminContentData (source): %v", err)
	}

	target, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{
		ParentID:        types.NullableAdminContentID{},
		FirstChildID:    types.NullableAdminContentID{},
		NextSiblingID:   types.NullableAdminContentID{},
		PrevSiblingID:   types.NullableAdminContentID{},
		AdminRouteID:    *adminRouteID,
		AdminDatatypeID: adminDatatypeID,
		AuthorID:        authorID,
		Status:          types.ContentStatusDraft,
		DateCreated:     now,
		DateModified:    now,
	})
	if err != nil {
		t.Fatalf("prerequisite CreateAdminContentData (target): %v", err)
	}

	adminFieldID := seed.AdminField.AdminFieldID

	// --- Count: starts at zero ---
	count, err := d.CountAdminContentRelations()
	if err != nil {
		t.Fatalf("initial CountAdminContentRelations: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountAdminContentRelations = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminContentRelation(ctx, ac, CreateAdminContentRelationParams{
		SourceContentID: source.AdminContentDataID,
		TargetContentID: target.AdminContentDataID,
		AdminFieldID:    adminFieldID,
		SortOrder:       1,
		DateCreated:     now,
	})
	if err != nil {
		t.Fatalf("CreateAdminContentRelation: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminContentRelation returned nil")
	}
	if created.AdminContentRelationID.IsZero() {
		t.Fatal("CreateAdminContentRelation returned zero AdminContentRelationID")
	}
	if created.SourceContentID != source.AdminContentDataID {
		t.Errorf("SourceContentID = %v, want %v", created.SourceContentID, source.AdminContentDataID)
	}
	if created.TargetContentID != target.AdminContentDataID {
		t.Errorf("TargetContentID = %v, want %v", created.TargetContentID, target.AdminContentDataID)
	}
	if created.AdminFieldID != adminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", created.AdminFieldID, adminFieldID)
	}
	if created.SortOrder != 1 {
		t.Errorf("SortOrder = %d, want %d", created.SortOrder, 1)
	}

	// --- Get ---
	got, err := d.GetAdminContentRelation(created.AdminContentRelationID)
	if err != nil {
		t.Fatalf("GetAdminContentRelation: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminContentRelation returned nil")
	}
	if got.AdminContentRelationID != created.AdminContentRelationID {
		t.Errorf("GetAdminContentRelation ID = %v, want %v", got.AdminContentRelationID, created.AdminContentRelationID)
	}
	if got.SourceContentID != created.SourceContentID {
		t.Errorf("GetAdminContentRelation SourceContentID = %v, want %v", got.SourceContentID, created.SourceContentID)
	}

	// --- ListAdminContentRelationsBySource ---
	bySource, err := d.ListAdminContentRelationsBySource(source.AdminContentDataID)
	if err != nil {
		t.Fatalf("ListAdminContentRelationsBySource: %v", err)
	}
	if bySource == nil {
		t.Fatal("ListAdminContentRelationsBySource returned nil")
	}
	if len(*bySource) != 1 {
		t.Fatalf("ListAdminContentRelationsBySource len = %d, want 1", len(*bySource))
	}
	if (*bySource)[0].AdminContentRelationID != created.AdminContentRelationID {
		t.Errorf("ListAdminContentRelationsBySource[0].AdminContentRelationID = %v, want %v", (*bySource)[0].AdminContentRelationID, created.AdminContentRelationID)
	}

	// --- ListAdminContentRelationsByTarget ---
	byTarget, err := d.ListAdminContentRelationsByTarget(target.AdminContentDataID)
	if err != nil {
		t.Fatalf("ListAdminContentRelationsByTarget: %v", err)
	}
	if byTarget == nil {
		t.Fatal("ListAdminContentRelationsByTarget returned nil")
	}
	if len(*byTarget) != 1 {
		t.Fatalf("ListAdminContentRelationsByTarget len = %d, want 1", len(*byTarget))
	}
	if (*byTarget)[0].AdminContentRelationID != created.AdminContentRelationID {
		t.Errorf("ListAdminContentRelationsByTarget[0].AdminContentRelationID = %v, want %v", (*byTarget)[0].AdminContentRelationID, created.AdminContentRelationID)
	}

	// --- ListAdminContentRelationsBySourceAndField ---
	bySourceAndField, err := d.ListAdminContentRelationsBySourceAndField(source.AdminContentDataID, adminFieldID)
	if err != nil {
		t.Fatalf("ListAdminContentRelationsBySourceAndField: %v", err)
	}
	if bySourceAndField == nil {
		t.Fatal("ListAdminContentRelationsBySourceAndField returned nil")
	}
	if len(*bySourceAndField) != 1 {
		t.Fatalf("ListAdminContentRelationsBySourceAndField len = %d, want 1", len(*bySourceAndField))
	}
	if (*bySourceAndField)[0].AdminContentRelationID != created.AdminContentRelationID {
		t.Errorf("ListAdminContentRelationsBySourceAndField[0].AdminContentRelationID = %v, want %v", (*bySourceAndField)[0].AdminContentRelationID, created.AdminContentRelationID)
	}

	// --- ListAdminContentRelationsBySource with non-matching source ---
	noMatch, err := d.ListAdminContentRelationsBySource(types.NewAdminContentID())
	if err != nil {
		t.Fatalf("ListAdminContentRelationsBySource (no match): %v", err)
	}
	if noMatch != nil && len(*noMatch) != 0 {
		t.Errorf("ListAdminContentRelationsBySource (no match) len = %d, want 0", len(*noMatch))
	}

	// --- Count: now 1 ---
	count, err = d.CountAdminContentRelations()
	if err != nil {
		t.Fatalf("CountAdminContentRelations after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminContentRelations after create = %d, want 1", *count)
	}

	// --- UpdateAdminContentRelationSortOrder ---
	err = d.UpdateAdminContentRelationSortOrder(ctx, ac, UpdateAdminContentRelationSortOrderParams{
		AdminContentRelationID: created.AdminContentRelationID,
		SortOrder:              42,
	})
	if err != nil {
		t.Fatalf("UpdateAdminContentRelationSortOrder: %v", err)
	}

	// --- Get after sort order update ---
	afterSortUpdate, err := d.GetAdminContentRelation(created.AdminContentRelationID)
	if err != nil {
		t.Fatalf("GetAdminContentRelation after sort order update: %v", err)
	}
	if afterSortUpdate.SortOrder != 42 {
		t.Errorf("SortOrder after update = %d, want %d", afterSortUpdate.SortOrder, 42)
	}

	// --- Delete ---
	err = d.DeleteAdminContentRelation(ctx, ac, created.AdminContentRelationID)
	if err != nil {
		t.Fatalf("DeleteAdminContentRelation: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetAdminContentRelation(created.AdminContentRelationID)
	if err == nil {
		t.Fatal("GetAdminContentRelation after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountAdminContentRelations()
	if err != nil {
		t.Fatalf("CountAdminContentRelations after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountAdminContentRelations after delete = %d, want 0", *count)
	}
}
