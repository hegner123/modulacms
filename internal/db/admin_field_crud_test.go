// Integration tests for the admin_field entity CRUD lifecycle.
// Uses testSeededDB (Tier 1b: requires user for author_id FK).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminField(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at 1 (seed admin field) ---
	count, err := d.CountAdminFields()
	if err != nil {
		t.Fatalf("initial CountAdminFields: %v", err)
	}
	if *count != 1 {
		t.Fatalf("initial CountAdminFields = %d, want 1", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{},
		Label:        "crud-test-admin-field",
		Data:         "admin field data",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateAdminField: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminField returned nil")
	}
	if created.AdminFieldID.IsZero() {
		t.Fatal("CreateAdminField returned zero AdminFieldID")
	}
	if created.Label != "crud-test-admin-field" {
		t.Errorf("Label = %q, want %q", created.Label, "crud-test-admin-field")
	}
	if created.Data != "admin field data" {
		t.Errorf("Data = %q, want %q", created.Data, "admin field data")
	}
	if created.Type != types.FieldTypeText {
		t.Errorf("Type = %q, want %q", created.Type, types.FieldTypeText)
	}

	// --- Get ---
	got, err := d.GetAdminField(created.AdminFieldID)
	if err != nil {
		t.Fatalf("GetAdminField: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminField returned nil")
	}
	if got.AdminFieldID != created.AdminFieldID {
		t.Errorf("GetAdminField ID = %v, want %v", got.AdminFieldID, created.AdminFieldID)
	}
	if got.Label != created.Label {
		t.Errorf("GetAdminField Label = %q, want %q", got.Label, created.Label)
	}
	if got.Data != created.Data {
		t.Errorf("GetAdminField Data = %q, want %q", got.Data, created.Data)
	}
	if got.Type != created.Type {
		t.Errorf("GetAdminField Type = %q, want %q", got.Type, created.Type)
	}

	// --- List ---
	list, err := d.ListAdminFields()
	if err != nil {
		t.Fatalf("ListAdminFields: %v", err)
	}
	if list == nil {
		t.Fatal("ListAdminFields returned nil")
	}
	if len(*list) != 2 {
		t.Fatalf("ListAdminFields len = %d, want 2", len(*list))
	}

	// --- Count: now 2 ---
	count, err = d.CountAdminFields()
	if err != nil {
		t.Fatalf("CountAdminFields after create: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountAdminFields after create = %d, want 2", *count)
	}

	// --- Update ---
	updatedNow := types.TimestampNow()
	updateResult, err := d.UpdateAdminField(ctx, ac, UpdateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{},
		Label:        "crud-test-admin-field-updated",
		Data:         "updated admin data",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: updatedNow,
		AdminFieldID: created.AdminFieldID,
	})
	if err != nil {
		t.Fatalf("UpdateAdminField: %v", err)
	}
	// UpdateAdminField returns a success message on success
	if updateResult == nil {
		t.Error("UpdateAdminField returned nil message, expected success message")
	}

	// --- Get after update ---
	updated, err := d.GetAdminField(created.AdminFieldID)
	if err != nil {
		t.Fatalf("GetAdminField after update: %v", err)
	}
	if updated.Label != "crud-test-admin-field-updated" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "crud-test-admin-field-updated")
	}
	if updated.Data != "updated admin data" {
		t.Errorf("updated Data = %q, want %q", updated.Data, "updated admin data")
	}

	// --- Delete (only our created admin field, not the seed) ---
	err = d.DeleteAdminField(ctx, ac, created.AdminFieldID)
	if err != nil {
		t.Fatalf("DeleteAdminField: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetAdminField(created.AdminFieldID)
	if err == nil {
		t.Fatal("GetAdminField after delete: expected error, got nil")
	}

	// --- Count: back to 1 (seed admin field remains) ---
	count, err = d.CountAdminFields()
	if err != nil {
		t.Fatalf("CountAdminFields after delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminFields after delete = %d, want 1", *count)
	}
}
