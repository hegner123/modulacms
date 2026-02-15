// Integration tests for the field entity CRUD lifecycle.
// Uses testSeededDB (Tier 1b: requires user for author_id FK).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Field(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at 1 (seed field) ---
	count, err := d.CountFields()
	if err != nil {
		t.Fatalf("initial CountFields: %v", err)
	}
	if *count != 1 {
		t.Fatalf("initial CountFields = %d, want 1", *count)
	}

	// --- Create ---
	created, err := d.CreateField(ctx, ac, CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "crud-test-field",
		Data:         "test data",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateField: %v", err)
	}
	if created == nil {
		t.Fatal("CreateField returned nil")
	}
	if created.FieldID.IsZero() {
		t.Fatal("CreateField returned zero FieldID")
	}
	if created.Label != "crud-test-field" {
		t.Errorf("Label = %q, want %q", created.Label, "crud-test-field")
	}
	if created.Data != "test data" {
		t.Errorf("Data = %q, want %q", created.Data, "test data")
	}
	if created.Type != types.FieldTypeText {
		t.Errorf("Type = %q, want %q", created.Type, types.FieldTypeText)
	}

	// --- Get ---
	got, err := d.GetField(created.FieldID)
	if err != nil {
		t.Fatalf("GetField: %v", err)
	}
	if got == nil {
		t.Fatal("GetField returned nil")
	}
	if got.FieldID != created.FieldID {
		t.Errorf("GetField ID = %v, want %v", got.FieldID, created.FieldID)
	}
	if got.Label != created.Label {
		t.Errorf("GetField Label = %q, want %q", got.Label, created.Label)
	}
	if got.Data != created.Data {
		t.Errorf("GetField Data = %q, want %q", got.Data, created.Data)
	}
	if got.Type != created.Type {
		t.Errorf("GetField Type = %q, want %q", got.Type, created.Type)
	}

	// --- List ---
	list, err := d.ListFields()
	if err != nil {
		t.Fatalf("ListFields: %v", err)
	}
	if list == nil {
		t.Fatal("ListFields returned nil")
	}
	if len(*list) != 2 {
		t.Fatalf("ListFields len = %d, want 2", len(*list))
	}

	// --- Count: now 2 ---
	count, err = d.CountFields()
	if err != nil {
		t.Fatalf("CountFields after create: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountFields after create = %d, want 2", *count)
	}

	// --- Update ---
	updatedNow := types.TimestampNow()
	updateResult, err := d.UpdateField(ctx, ac, UpdateFieldParams{
		ParentID:     types.NullableDatatypeID{},
		Label:        "crud-test-field-updated",
		Data:         "updated data",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: updatedNow,
		FieldID:      created.FieldID,
	})
	if err != nil {
		t.Fatalf("UpdateField: %v", err)
	}
	// UpdateField returns a success message on success
	if updateResult == nil {
		t.Error("UpdateField returned nil message, expected success message")
	}

	// --- Get after update ---
	updated, err := d.GetField(created.FieldID)
	if err != nil {
		t.Fatalf("GetField after update: %v", err)
	}
	if updated.Label != "crud-test-field-updated" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "crud-test-field-updated")
	}
	if updated.Data != "updated data" {
		t.Errorf("updated Data = %q, want %q", updated.Data, "updated data")
	}

	// --- Delete (only our created field, not the seed) ---
	err = d.DeleteField(ctx, ac, created.FieldID)
	if err != nil {
		t.Fatalf("DeleteField: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetField(created.FieldID)
	if err == nil {
		t.Fatal("GetField after delete: expected error, got nil")
	}

	// --- Count: back to 1 (seed field remains) ---
	count, err = d.CountFields()
	if err != nil {
		t.Fatalf("CountFields after delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountFields after delete = %d, want 1", *count)
	}
}

// TestDatabase_CRUD_Field_ListFieldsByDatatypeID tests listing fields filtered by datatype.
func TestDatabase_CRUD_Field_ListFieldsByDatatypeID(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// Create a field with ParentID referencing the seed datatype
	parentID := types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true}
	_, err := d.CreateField(ctx, ac, CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     parentID,
		Label:        "typed-field",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateField (with parent): %v", err)
	}

	// List fields by this datatype ID
	fields, err := d.ListFieldsByDatatypeID(parentID)
	if err != nil {
		t.Fatalf("ListFieldsByDatatypeID: %v", err)
	}
	if fields == nil {
		t.Fatal("ListFieldsByDatatypeID returned nil")
	}
	if len(*fields) != 1 {
		t.Fatalf("ListFieldsByDatatypeID len = %d, want 1", len(*fields))
	}
	if (*fields)[0].Label != "typed-field" {
		t.Errorf("ListFieldsByDatatypeID[0].Label = %q, want %q", (*fields)[0].Label, "typed-field")
	}

	// Listing by a non-existent datatype returns empty list (not an error)
	nonExistentParent := types.NullableDatatypeID{
		ID:    types.DatatypeID(types.NewDatatypeID()),
		Valid: true,
	}
	emptyFields, err := d.ListFieldsByDatatypeID(nonExistentParent)
	if err != nil {
		t.Fatalf("ListFieldsByDatatypeID (non-existent): %v", err)
	}
	if emptyFields == nil {
		t.Fatal("ListFieldsByDatatypeID (non-existent) returned nil")
	}
	if len(*emptyFields) != 0 {
		t.Fatalf("ListFieldsByDatatypeID (non-existent) len = %d, want 0", len(*emptyFields))
	}
}
