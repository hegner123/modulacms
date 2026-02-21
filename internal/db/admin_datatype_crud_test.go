// Integration tests for the admin_datatype entity CRUD lifecycle.
// Uses testSeededDB (Tier 1b: requires user for author_id FK).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminDatatype(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := seed.User.UserID

	// --- Count: starts at 1 (seed admin datatype) ---
	count, err := d.CountAdminDatatypes()
	if err != nil {
		t.Fatalf("initial CountAdminDatatypes: %v", err)
	}
	if *count != 1 {
		t.Fatalf("initial CountAdminDatatypes = %d, want 1", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminDatatype(ctx, ac, CreateAdminDatatypeParams{
		ParentID:     types.NullableAdminDatatypeID{},
		Label:        "crud-test-admin-datatype",
		Type:         "widget",
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateAdminDatatype: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminDatatype returned nil")
	}
	if created.AdminDatatypeID.IsZero() {
		t.Fatal("CreateAdminDatatype returned zero AdminDatatypeID")
	}
	if created.Label != "crud-test-admin-datatype" {
		t.Errorf("Label = %q, want %q", created.Label, "crud-test-admin-datatype")
	}
	if created.Type != "widget" {
		t.Errorf("Type = %q, want %q", created.Type, "widget")
	}

	// --- Get (by ID) ---
	got, err := d.GetAdminDatatypeById(created.AdminDatatypeID)
	if err != nil {
		t.Fatalf("GetAdminDatatypeById: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminDatatypeById returned nil")
	}
	if got.AdminDatatypeID != created.AdminDatatypeID {
		t.Errorf("GetAdminDatatypeById ID = %v, want %v", got.AdminDatatypeID, created.AdminDatatypeID)
	}
	if got.Label != created.Label {
		t.Errorf("GetAdminDatatypeById Label = %q, want %q", got.Label, created.Label)
	}
	if got.Type != created.Type {
		t.Errorf("GetAdminDatatypeById Type = %q, want %q", got.Type, created.Type)
	}

	// --- List ---
	list, err := d.ListAdminDatatypes()
	if err != nil {
		t.Fatalf("ListAdminDatatypes: %v", err)
	}
	if list == nil {
		t.Fatal("ListAdminDatatypes returned nil")
	}
	if len(*list) != 2 {
		t.Fatalf("ListAdminDatatypes len = %d, want 2", len(*list))
	}

	// --- Count: now 2 ---
	count, err = d.CountAdminDatatypes()
	if err != nil {
		t.Fatalf("CountAdminDatatypes after create: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountAdminDatatypes after create = %d, want 2", *count)
	}

	// --- Update ---
	updatedNow := types.TimestampNow()
	updateResult, err := d.UpdateAdminDatatype(ctx, ac, UpdateAdminDatatypeParams{
		ParentID:        types.NullableAdminDatatypeID{},
		Label:           "crud-test-admin-datatype-updated",
		Type:            "panel",
		AuthorID:        authorID,
		DateCreated:     now,
		DateModified:    updatedNow,
		AdminDatatypeID: created.AdminDatatypeID,
	})
	if err != nil {
		t.Fatalf("UpdateAdminDatatype: %v", err)
	}
	// UpdateAdminDatatype returns a success message on success
	if updateResult == nil {
		t.Error("UpdateAdminDatatype returned nil message, expected success message")
	}

	// --- Get after update ---
	updated, err := d.GetAdminDatatypeById(created.AdminDatatypeID)
	if err != nil {
		t.Fatalf("GetAdminDatatypeById after update: %v", err)
	}
	if updated.Label != "crud-test-admin-datatype-updated" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "crud-test-admin-datatype-updated")
	}
	if updated.Type != "panel" {
		t.Errorf("updated Type = %q, want %q", updated.Type, "panel")
	}

	// --- Delete (only our created admin datatype, not the seed) ---
	err = d.DeleteAdminDatatype(ctx, ac, created.AdminDatatypeID)
	if err != nil {
		t.Fatalf("DeleteAdminDatatype: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetAdminDatatypeById(created.AdminDatatypeID)
	if err == nil {
		t.Fatal("GetAdminDatatypeById after delete: expected error, got nil")
	}

	// --- Count: back to 1 (seed admin datatype remains) ---
	count, err = d.CountAdminDatatypes()
	if err != nil {
		t.Fatalf("CountAdminDatatypes after delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminDatatypes after delete = %d, want 1", *count)
	}
}
