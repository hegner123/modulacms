// Integration tests for the admin_datatype_field entity CRUD lifecycle.
// Uses testSeededDB (Tier 2: requires AdminDatatype and AdminField seed records).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminDatatypeField(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	adminDatatypeID := seed.AdminDatatype.AdminDatatypeID
	adminFieldID := seed.AdminField.AdminFieldID

	// --- Count: starts at zero ---
	count, err := d.CountAdminDatatypeFields()
	if err != nil {
		t.Fatalf("initial CountAdminDatatypeFields: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountAdminDatatypeFields = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminDatatypeField(ctx, ac, CreateAdminDatatypeFieldParams{
		AdminDatatypeID: adminDatatypeID,
		AdminFieldID:    adminFieldID,
	})
	if err != nil {
		t.Fatalf("CreateAdminDatatypeField: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminDatatypeField returned nil")
	}
	if created.ID == "" {
		t.Fatal("CreateAdminDatatypeField returned empty ID (expected auto-generated)")
	}
	if created.AdminDatatypeID != adminDatatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", created.AdminDatatypeID, adminDatatypeID)
	}
	if created.AdminFieldID != adminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", created.AdminFieldID, adminFieldID)
	}

	// --- ListAdminDatatypeField ---
	list, err := d.ListAdminDatatypeField()
	if err != nil {
		t.Fatalf("ListAdminDatatypeField: %v", err)
	}
	if list == nil {
		t.Fatal("ListAdminDatatypeField returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListAdminDatatypeField len = %d, want 1", len(*list))
	}
	if (*list)[0].ID != created.ID {
		t.Errorf("ListAdminDatatypeField[0].ID = %q, want %q", (*list)[0].ID, created.ID)
	}

	// --- ListAdminDatatypeFieldByDatatypeID ---
	byDatatype, err := d.ListAdminDatatypeFieldByDatatypeID(adminDatatypeID)
	if err != nil {
		t.Fatalf("ListAdminDatatypeFieldByDatatypeID: %v", err)
	}
	if byDatatype == nil {
		t.Fatal("ListAdminDatatypeFieldByDatatypeID returned nil")
	}
	if len(*byDatatype) != 1 {
		t.Fatalf("ListAdminDatatypeFieldByDatatypeID len = %d, want 1", len(*byDatatype))
	}
	if (*byDatatype)[0].ID != created.ID {
		t.Errorf("ListAdminDatatypeFieldByDatatypeID[0].ID = %q, want %q", (*byDatatype)[0].ID, created.ID)
	}

	// --- ListAdminDatatypeFieldByFieldID ---
	byField, err := d.ListAdminDatatypeFieldByFieldID(adminFieldID)
	if err != nil {
		t.Fatalf("ListAdminDatatypeFieldByFieldID: %v", err)
	}
	if byField == nil {
		t.Fatal("ListAdminDatatypeFieldByFieldID returned nil")
	}
	if len(*byField) != 1 {
		t.Fatalf("ListAdminDatatypeFieldByFieldID len = %d, want 1", len(*byField))
	}
	if (*byField)[0].ID != created.ID {
		t.Errorf("ListAdminDatatypeFieldByFieldID[0].ID = %q, want %q", (*byField)[0].ID, created.ID)
	}

	// --- ListAdminDatatypeFieldByDatatypeID with non-matching ID ---
	noMatchDT, err := d.ListAdminDatatypeFieldByDatatypeID(types.NewAdminDatatypeID())
	if err != nil {
		t.Fatalf("ListAdminDatatypeFieldByDatatypeID (no match): %v", err)
	}
	if noMatchDT != nil && len(*noMatchDT) != 0 {
		t.Errorf("ListAdminDatatypeFieldByDatatypeID (no match) len = %d, want 0", len(*noMatchDT))
	}

	// --- Count: now 1 ---
	count, err = d.CountAdminDatatypeFields()
	if err != nil {
		t.Fatalf("CountAdminDatatypeFields after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminDatatypeFields after create = %d, want 1", *count)
	}

	// --- Create a second admin field to use for update ---
	secondAdminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{},
		Label:        "second-admin-field",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     types.NullableUserID{ID: seed.User.UserID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("prerequisite CreateAdminField: %v", err)
	}
	updatedAdminFieldID := secondAdminField.AdminFieldID

	// --- Update ---
	updateResult, err := d.UpdateAdminDatatypeField(ctx, ac, UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: adminDatatypeID,
		AdminFieldID:    updatedAdminFieldID,
		ID:              created.ID,
	})
	if err != nil {
		t.Fatalf("UpdateAdminDatatypeField: %v", err)
	}
	if updateResult == nil {
		t.Error("UpdateAdminDatatypeField returned nil message (expected success message)")
	}

	// --- Verify update via list ---
	listAfterUpdate, err := d.ListAdminDatatypeField()
	if err != nil {
		t.Fatalf("ListAdminDatatypeField after update: %v", err)
	}
	if len(*listAfterUpdate) != 1 {
		t.Fatalf("ListAdminDatatypeField after update len = %d, want 1", len(*listAfterUpdate))
	}
	if (*listAfterUpdate)[0].AdminFieldID != updatedAdminFieldID {
		t.Errorf("updated AdminFieldID = %v, want %v", (*listAfterUpdate)[0].AdminFieldID, updatedAdminFieldID)
	}

	// --- Delete ---
	err = d.DeleteAdminDatatypeField(ctx, ac, created.ID)
	if err != nil {
		t.Fatalf("DeleteAdminDatatypeField: %v", err)
	}

	// --- Count: back to zero ---
	count, err = d.CountAdminDatatypeFields()
	if err != nil {
		t.Fatalf("CountAdminDatatypeFields after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountAdminDatatypeFields after delete = %d, want 0", *count)
	}
}
