// Integration tests for the admin_field_type entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0: no FK dependencies).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminFieldType(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// --- Count: starts at zero ---
	count, err := d.CountAdminFieldTypes()
	if err != nil {
		t.Fatalf("initial CountAdminFieldTypes: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountAdminFieldTypes = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{
		Type:  "text",
		Label: "Text Input",
	})
	if err != nil {
		t.Fatalf("CreateAdminFieldType: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminFieldType returned nil")
	}
	if created.AdminFieldTypeID.IsZero() {
		t.Fatal("CreateAdminFieldType returned zero AdminFieldTypeID")
	}
	if created.Type != "text" {
		t.Errorf("Type = %q, want %q", created.Type, "text")
	}
	if created.Label != "Text Input" {
		t.Errorf("Label = %q, want %q", created.Label, "Text Input")
	}

	// --- Get ---
	got, err := d.GetAdminFieldType(created.AdminFieldTypeID)
	if err != nil {
		t.Fatalf("GetAdminFieldType: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminFieldType returned nil")
	}
	if got.AdminFieldTypeID != created.AdminFieldTypeID {
		t.Errorf("GetAdminFieldType ID = %v, want %v", got.AdminFieldTypeID, created.AdminFieldTypeID)
	}
	if got.Type != created.Type {
		t.Errorf("GetAdminFieldType Type = %q, want %q", got.Type, created.Type)
	}
	if got.Label != created.Label {
		t.Errorf("GetAdminFieldType Label = %q, want %q", got.Label, created.Label)
	}

	// --- List ---
	list, err := d.ListAdminFieldTypes()
	if err != nil {
		t.Fatalf("ListAdminFieldTypes: %v", err)
	}
	if list == nil {
		t.Fatal("ListAdminFieldTypes returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListAdminFieldTypes len = %d, want 1", len(*list))
	}
	if (*list)[0].AdminFieldTypeID != created.AdminFieldTypeID {
		t.Errorf("ListAdminFieldTypes[0].AdminFieldTypeID = %v, want %v", (*list)[0].AdminFieldTypeID, created.AdminFieldTypeID)
	}

	// --- Count: now 1 ---
	count, err = d.CountAdminFieldTypes()
	if err != nil {
		t.Fatalf("CountAdminFieldTypes after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminFieldTypes after create = %d, want 1", *count)
	}

	// --- Update ---
	_, err = d.UpdateAdminFieldType(ctx, ac, UpdateAdminFieldTypeParams{
		Type:             "textarea",
		Label:            "Text Area",
		AdminFieldTypeID: created.AdminFieldTypeID,
	})
	if err != nil {
		t.Fatalf("UpdateAdminFieldType: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetAdminFieldType(created.AdminFieldTypeID)
	if err != nil {
		t.Fatalf("GetAdminFieldType after update: %v", err)
	}
	if updated.Type != "textarea" {
		t.Errorf("updated Type = %q, want %q", updated.Type, "textarea")
	}
	if updated.Label != "Text Area" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "Text Area")
	}

	// --- Delete ---
	err = d.DeleteAdminFieldType(ctx, ac, created.AdminFieldTypeID)
	if err != nil {
		t.Fatalf("DeleteAdminFieldType: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetAdminFieldType(created.AdminFieldTypeID)
	if err == nil {
		t.Fatal("GetAdminFieldType after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountAdminFieldTypes()
	if err != nil {
		t.Fatalf("CountAdminFieldTypes after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountAdminFieldTypes after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_AdminFieldType_MultipleRecords verifies that multiple
// admin field types can coexist and be listed independently.
func TestDatabase_CRUD_AdminFieldType_MultipleRecords(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	entries := []struct{ Type, Label string }{
		{"text", "Text Input"},
		{"number", "Number"},
		{"boolean", "Boolean"},
	}
	ids := make([]types.AdminFieldTypeID, len(entries))

	for i, e := range entries {
		ft, err := d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{
			Type:  e.Type,
			Label: e.Label,
		})
		if err != nil {
			t.Fatalf("CreateAdminFieldType(%s): %v", e.Type, err)
		}
		ids[i] = ft.AdminFieldTypeID
	}

	count, err := d.CountAdminFieldTypes()
	if err != nil {
		t.Fatalf("CountAdminFieldTypes: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountAdminFieldTypes = %d, want 3", *count)
	}

	list, err := d.ListAdminFieldTypes()
	if err != nil {
		t.Fatalf("ListAdminFieldTypes: %v", err)
	}
	if len(*list) != 3 {
		t.Fatalf("ListAdminFieldTypes len = %d, want 3", len(*list))
	}

	// Delete one and verify count drops
	err = d.DeleteAdminFieldType(ctx, ac, ids[1])
	if err != nil {
		t.Fatalf("DeleteAdminFieldType: %v", err)
	}

	count, err = d.CountAdminFieldTypes()
	if err != nil {
		t.Fatalf("CountAdminFieldTypes after delete: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountAdminFieldTypes after delete = %d, want 2", *count)
	}
}

// TestDatabase_CRUD_AdminFieldType_GetByType verifies the GetAdminFieldTypeByType query.
func TestDatabase_CRUD_AdminFieldType_GetByType(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	created, err := d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{
		Type:  "richtext",
		Label: "Rich Text",
	})
	if err != nil {
		t.Fatalf("CreateAdminFieldType: %v", err)
	}

	got, err := d.GetAdminFieldTypeByType("richtext")
	if err != nil {
		t.Fatalf("GetAdminFieldTypeByType: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminFieldTypeByType returned nil")
	}
	if got.AdminFieldTypeID != created.AdminFieldTypeID {
		t.Errorf("GetAdminFieldTypeByType ID = %v, want %v", got.AdminFieldTypeID, created.AdminFieldTypeID)
	}
	if got.Type != "richtext" {
		t.Errorf("GetAdminFieldTypeByType Type = %q, want %q", got.Type, "richtext")
	}
	if got.Label != "Rich Text" {
		t.Errorf("GetAdminFieldTypeByType Label = %q, want %q", got.Label, "Rich Text")
	}

	// Non-existent type should return error
	_, err = d.GetAdminFieldTypeByType("nonexistent")
	if err == nil {
		t.Fatal("GetAdminFieldTypeByType for nonexistent type: expected error, got nil")
	}
}
