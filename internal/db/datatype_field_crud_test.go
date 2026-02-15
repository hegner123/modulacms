// Integration tests for the datatype_field entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires Datatype and Field seed records).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_DatatypeField(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	datatypeID := seed.Datatype.DatatypeID
	fieldID := seed.Field.FieldID

	// --- Count: starts at zero ---
	count, err := d.CountDatatypeFields()
	if err != nil {
		t.Fatalf("initial CountDatatypeFields: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountDatatypeFields = %d, want 0", *count)
	}

	// --- Create (with auto-generated ID via empty string) ---
	created, err := d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		ID:         "",
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  1,
	})
	if err != nil {
		t.Fatalf("CreateDatatypeField: %v", err)
	}
	if created == nil {
		t.Fatal("CreateDatatypeField returned nil")
	}
	if created.ID == "" {
		t.Fatal("CreateDatatypeField returned empty ID (expected auto-generated)")
	}
	if created.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", created.DatatypeID, datatypeID)
	}
	if created.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", created.FieldID, fieldID)
	}
	if created.SortOrder != 1 {
		t.Errorf("SortOrder = %d, want %d", created.SortOrder, 1)
	}

	// --- Create with pre-specified ID ---
	// Create a second field to use as FK for this record
	secondField, err := d.CreateField(ctx, ac, CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "second-test-field",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     types.NullableUserID{ID: seed.User.UserID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("prerequisite CreateField: %v", err)
	}
	preID := string(types.NewDatatypeFieldID())
	created2, err := d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		ID:         preID,
		DatatypeID: datatypeID,
		FieldID:    secondField.FieldID,
		SortOrder:  2,
	})
	if err != nil {
		t.Fatalf("CreateDatatypeField (pre-specified ID): %v", err)
	}
	if created2.ID != preID {
		t.Errorf("pre-specified ID = %q, want %q", created2.ID, preID)
	}

	// --- ListDatatypeField ---
	list, err := d.ListDatatypeField()
	if err != nil {
		t.Fatalf("ListDatatypeField: %v", err)
	}
	if list == nil {
		t.Fatal("ListDatatypeField returned nil")
	}
	if len(*list) != 2 {
		t.Fatalf("ListDatatypeField len = %d, want 2", len(*list))
	}

	// --- ListDatatypeFieldByDatatypeID ---
	byDatatype, err := d.ListDatatypeFieldByDatatypeID(datatypeID)
	if err != nil {
		t.Fatalf("ListDatatypeFieldByDatatypeID: %v", err)
	}
	if byDatatype == nil {
		t.Fatal("ListDatatypeFieldByDatatypeID returned nil")
	}
	if len(*byDatatype) != 2 {
		t.Fatalf("ListDatatypeFieldByDatatypeID len = %d, want 2", len(*byDatatype))
	}

	// --- ListDatatypeFieldByFieldID ---
	byField, err := d.ListDatatypeFieldByFieldID(fieldID)
	if err != nil {
		t.Fatalf("ListDatatypeFieldByFieldID: %v", err)
	}
	if byField == nil {
		t.Fatal("ListDatatypeFieldByFieldID returned nil")
	}
	if len(*byField) != 1 {
		t.Fatalf("ListDatatypeFieldByFieldID len = %d, want 1", len(*byField))
	}
	if (*byField)[0].ID != created.ID {
		t.Errorf("ListDatatypeFieldByFieldID[0].ID = %q, want %q", (*byField)[0].ID, created.ID)
	}

	// --- Count: now 2 ---
	count, err = d.CountDatatypeFields()
	if err != nil {
		t.Fatalf("CountDatatypeFields after creates: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountDatatypeFields after creates = %d, want 2", *count)
	}

	// --- Update ---
	updateResult, err := d.UpdateDatatypeField(ctx, ac, UpdateDatatypeFieldParams{
		DatatypeID: datatypeID,
		FieldID:    fieldID,
		SortOrder:  10,
		ID:         created.ID,
	})
	if err != nil {
		t.Fatalf("UpdateDatatypeField: %v", err)
	}
	if updateResult == nil {
		t.Error("UpdateDatatypeField returned nil message (expected success message)")
	}

	// --- Verify update via list ---
	listAfterUpdate, err := d.ListDatatypeField()
	if err != nil {
		t.Fatalf("ListDatatypeField after update: %v", err)
	}
	var found bool
	for _, item := range *listAfterUpdate {
		if item.ID == created.ID {
			found = true
			if item.SortOrder != 10 {
				t.Errorf("updated SortOrder = %d, want %d", item.SortOrder, 10)
			}
			break
		}
	}
	if !found {
		t.Error("updated record not found in ListDatatypeField")
	}

	// --- UpdateDatatypeFieldSortOrder ---
	err = d.UpdateDatatypeFieldSortOrder(ctx, ac, created.ID, 99)
	if err != nil {
		t.Fatalf("UpdateDatatypeFieldSortOrder: %v", err)
	}

	// --- Verify sort order update via list ---
	listAfterSortUpdate, err := d.ListDatatypeField()
	if err != nil {
		t.Fatalf("ListDatatypeField after sort order update: %v", err)
	}
	for _, item := range *listAfterSortUpdate {
		if item.ID == created.ID {
			if item.SortOrder != 99 {
				t.Errorf("SortOrder after UpdateDatatypeFieldSortOrder = %d, want %d", item.SortOrder, 99)
			}
			break
		}
	}

	// --- Delete ---
	err = d.DeleteDatatypeField(ctx, ac, created.ID)
	if err != nil {
		t.Fatalf("DeleteDatatypeField: %v", err)
	}

	// --- Count: should be 1 (one remaining) ---
	count, err = d.CountDatatypeFields()
	if err != nil {
		t.Fatalf("CountDatatypeFields after delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountDatatypeFields after delete = %d, want 1", *count)
	}

	// --- Delete second record ---
	err = d.DeleteDatatypeField(ctx, ac, created2.ID)
	if err != nil {
		t.Fatalf("DeleteDatatypeField (second): %v", err)
	}

	// --- Count: back to zero ---
	count, err = d.CountDatatypeFields()
	if err != nil {
		t.Fatalf("CountDatatypeFields after all deletes: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountDatatypeFields after all deletes = %d, want 0", *count)
	}
}
