// Integration tests for the content_relation entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires 2 ContentData records created in-test, plus Field, Route, User seeds).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_ContentRelation(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	routeID := types.NullableRouteID{ID: seed.Route.RouteID, Valid: true}
	datatypeID := types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true}
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Create 2 prerequisite ContentData records (source and target) ---
	source, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("prerequisite CreateContentData (source): %v", err)
	}

	target, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("prerequisite CreateContentData (target): %v", err)
	}

	fieldID := seed.Field.FieldID

	// --- Count: starts at zero ---
	count, err := d.CountContentRelations()
	if err != nil {
		t.Fatalf("initial CountContentRelations: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountContentRelations = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateContentRelation(ctx, ac, CreateContentRelationParams{
		SourceContentID: source.ContentDataID,
		TargetContentID: target.ContentDataID,
		FieldID:         fieldID,
		SortOrder:       1,
		DateCreated:     now,
	})
	if err != nil {
		t.Fatalf("CreateContentRelation: %v", err)
	}
	if created == nil {
		t.Fatal("CreateContentRelation returned nil")
	}
	if created.ContentRelationID.IsZero() {
		t.Fatal("CreateContentRelation returned zero ContentRelationID")
	}
	if created.SourceContentID != source.ContentDataID {
		t.Errorf("SourceContentID = %v, want %v", created.SourceContentID, source.ContentDataID)
	}
	if created.TargetContentID != target.ContentDataID {
		t.Errorf("TargetContentID = %v, want %v", created.TargetContentID, target.ContentDataID)
	}
	if created.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", created.FieldID, fieldID)
	}
	if created.SortOrder != 1 {
		t.Errorf("SortOrder = %d, want %d", created.SortOrder, 1)
	}

	// --- Get ---
	got, err := d.GetContentRelation(created.ContentRelationID)
	if err != nil {
		t.Fatalf("GetContentRelation: %v", err)
	}
	if got == nil {
		t.Fatal("GetContentRelation returned nil")
	}
	if got.ContentRelationID != created.ContentRelationID {
		t.Errorf("GetContentRelation ID = %v, want %v", got.ContentRelationID, created.ContentRelationID)
	}
	if got.SourceContentID != created.SourceContentID {
		t.Errorf("GetContentRelation SourceContentID = %v, want %v", got.SourceContentID, created.SourceContentID)
	}

	// --- ListContentRelationsBySource ---
	bySource, err := d.ListContentRelationsBySource(source.ContentDataID)
	if err != nil {
		t.Fatalf("ListContentRelationsBySource: %v", err)
	}
	if bySource == nil {
		t.Fatal("ListContentRelationsBySource returned nil")
	}
	if len(*bySource) != 1 {
		t.Fatalf("ListContentRelationsBySource len = %d, want 1", len(*bySource))
	}
	if (*bySource)[0].ContentRelationID != created.ContentRelationID {
		t.Errorf("ListContentRelationsBySource[0].ContentRelationID = %v, want %v", (*bySource)[0].ContentRelationID, created.ContentRelationID)
	}

	// --- ListContentRelationsByTarget ---
	byTarget, err := d.ListContentRelationsByTarget(target.ContentDataID)
	if err != nil {
		t.Fatalf("ListContentRelationsByTarget: %v", err)
	}
	if byTarget == nil {
		t.Fatal("ListContentRelationsByTarget returned nil")
	}
	if len(*byTarget) != 1 {
		t.Fatalf("ListContentRelationsByTarget len = %d, want 1", len(*byTarget))
	}
	if (*byTarget)[0].ContentRelationID != created.ContentRelationID {
		t.Errorf("ListContentRelationsByTarget[0].ContentRelationID = %v, want %v", (*byTarget)[0].ContentRelationID, created.ContentRelationID)
	}

	// --- ListContentRelationsBySourceAndField ---
	bySourceAndField, err := d.ListContentRelationsBySourceAndField(source.ContentDataID, fieldID)
	if err != nil {
		t.Fatalf("ListContentRelationsBySourceAndField: %v", err)
	}
	if bySourceAndField == nil {
		t.Fatal("ListContentRelationsBySourceAndField returned nil")
	}
	if len(*bySourceAndField) != 1 {
		t.Fatalf("ListContentRelationsBySourceAndField len = %d, want 1", len(*bySourceAndField))
	}
	if (*bySourceAndField)[0].ContentRelationID != created.ContentRelationID {
		t.Errorf("ListContentRelationsBySourceAndField[0].ContentRelationID = %v, want %v", (*bySourceAndField)[0].ContentRelationID, created.ContentRelationID)
	}

	// --- ListContentRelationsBySource with non-matching source ---
	noMatch, err := d.ListContentRelationsBySource(types.NewContentID())
	if err != nil {
		t.Fatalf("ListContentRelationsBySource (no match): %v", err)
	}
	if noMatch != nil && len(*noMatch) != 0 {
		t.Errorf("ListContentRelationsBySource (no match) len = %d, want 0", len(*noMatch))
	}

	// --- Count: now 1 ---
	count, err = d.CountContentRelations()
	if err != nil {
		t.Fatalf("CountContentRelations after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountContentRelations after create = %d, want 1", *count)
	}

	// --- UpdateContentRelationSortOrder ---
	err = d.UpdateContentRelationSortOrder(ctx, ac, UpdateContentRelationSortOrderParams{
		ContentRelationID: created.ContentRelationID,
		SortOrder:         42,
	})
	if err != nil {
		t.Fatalf("UpdateContentRelationSortOrder: %v", err)
	}

	// --- Get after sort order update ---
	afterSortUpdate, err := d.GetContentRelation(created.ContentRelationID)
	if err != nil {
		t.Fatalf("GetContentRelation after sort order update: %v", err)
	}
	if afterSortUpdate.SortOrder != 42 {
		t.Errorf("SortOrder after update = %d, want %d", afterSortUpdate.SortOrder, 42)
	}

	// --- Delete ---
	err = d.DeleteContentRelation(ctx, ac, created.ContentRelationID)
	if err != nil {
		t.Fatalf("DeleteContentRelation: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetContentRelation(created.ContentRelationID)
	if err == nil {
		t.Fatal("GetContentRelation after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountContentRelations()
	if err != nil {
		t.Fatalf("CountContentRelations after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountContentRelations after delete = %d, want 0", *count)
	}
}
