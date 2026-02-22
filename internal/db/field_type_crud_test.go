// Integration tests for the field_type entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0: no FK dependencies).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_FieldType(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// --- Count: starts at zero ---
	count, err := d.CountFieldTypes()
	if err != nil {
		t.Fatalf("initial CountFieldTypes: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountFieldTypes = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{
		Type:  "text",
		Label: "Text Input",
	})
	if err != nil {
		t.Fatalf("CreateFieldType: %v", err)
	}
	if created == nil {
		t.Fatal("CreateFieldType returned nil")
	}
	if created.FieldTypeID.IsZero() {
		t.Fatal("CreateFieldType returned zero FieldTypeID")
	}
	if created.Type != "text" {
		t.Errorf("Type = %q, want %q", created.Type, "text")
	}
	if created.Label != "Text Input" {
		t.Errorf("Label = %q, want %q", created.Label, "Text Input")
	}

	// --- Get ---
	got, err := d.GetFieldType(created.FieldTypeID)
	if err != nil {
		t.Fatalf("GetFieldType: %v", err)
	}
	if got == nil {
		t.Fatal("GetFieldType returned nil")
	}
	if got.FieldTypeID != created.FieldTypeID {
		t.Errorf("GetFieldType ID = %v, want %v", got.FieldTypeID, created.FieldTypeID)
	}
	if got.Type != created.Type {
		t.Errorf("GetFieldType Type = %q, want %q", got.Type, created.Type)
	}
	if got.Label != created.Label {
		t.Errorf("GetFieldType Label = %q, want %q", got.Label, created.Label)
	}

	// --- List ---
	list, err := d.ListFieldTypes()
	if err != nil {
		t.Fatalf("ListFieldTypes: %v", err)
	}
	if list == nil {
		t.Fatal("ListFieldTypes returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListFieldTypes len = %d, want 1", len(*list))
	}
	if (*list)[0].FieldTypeID != created.FieldTypeID {
		t.Errorf("ListFieldTypes[0].FieldTypeID = %v, want %v", (*list)[0].FieldTypeID, created.FieldTypeID)
	}

	// --- Count: now 1 ---
	count, err = d.CountFieldTypes()
	if err != nil {
		t.Fatalf("CountFieldTypes after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountFieldTypes after create = %d, want 1", *count)
	}

	// --- Update ---
	_, err = d.UpdateFieldType(ctx, ac, UpdateFieldTypeParams{
		Type:        "textarea",
		Label:       "Text Area",
		FieldTypeID: created.FieldTypeID,
	})
	if err != nil {
		t.Fatalf("UpdateFieldType: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetFieldType(created.FieldTypeID)
	if err != nil {
		t.Fatalf("GetFieldType after update: %v", err)
	}
	if updated.Type != "textarea" {
		t.Errorf("updated Type = %q, want %q", updated.Type, "textarea")
	}
	if updated.Label != "Text Area" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "Text Area")
	}

	// --- Delete ---
	err = d.DeleteFieldType(ctx, ac, created.FieldTypeID)
	if err != nil {
		t.Fatalf("DeleteFieldType: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetFieldType(created.FieldTypeID)
	if err == nil {
		t.Fatal("GetFieldType after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountFieldTypes()
	if err != nil {
		t.Fatalf("CountFieldTypes after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountFieldTypes after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_FieldType_MultipleRecords verifies that multiple
// field types can coexist and be listed independently.
func TestDatabase_CRUD_FieldType_MultipleRecords(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	entries := []struct{ Type, Label string }{
		{"text", "Text Input"},
		{"number", "Number"},
		{"boolean", "Boolean"},
	}
	ids := make([]types.FieldTypeID, len(entries))

	for i, e := range entries {
		ft, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{
			Type:  e.Type,
			Label: e.Label,
		})
		if err != nil {
			t.Fatalf("CreateFieldType(%s): %v", e.Type, err)
		}
		ids[i] = ft.FieldTypeID
	}

	count, err := d.CountFieldTypes()
	if err != nil {
		t.Fatalf("CountFieldTypes: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountFieldTypes = %d, want 3", *count)
	}

	list, err := d.ListFieldTypes()
	if err != nil {
		t.Fatalf("ListFieldTypes: %v", err)
	}
	if len(*list) != 3 {
		t.Fatalf("ListFieldTypes len = %d, want 3", len(*list))
	}

	// Delete one and verify count drops
	err = d.DeleteFieldType(ctx, ac, ids[1])
	if err != nil {
		t.Fatalf("DeleteFieldType: %v", err)
	}

	count, err = d.CountFieldTypes()
	if err != nil {
		t.Fatalf("CountFieldTypes after delete: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountFieldTypes after delete = %d, want 2", *count)
	}
}

// TestDatabase_CRUD_FieldType_GetByType verifies the GetFieldTypeByType query.
func TestDatabase_CRUD_FieldType_GetByType(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	created, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{
		Type:  "richtext",
		Label: "Rich Text",
	})
	if err != nil {
		t.Fatalf("CreateFieldType: %v", err)
	}

	got, err := d.GetFieldTypeByType("richtext")
	if err != nil {
		t.Fatalf("GetFieldTypeByType: %v", err)
	}
	if got == nil {
		t.Fatal("GetFieldTypeByType returned nil")
	}
	if got.FieldTypeID != created.FieldTypeID {
		t.Errorf("GetFieldTypeByType ID = %v, want %v", got.FieldTypeID, created.FieldTypeID)
	}
	if got.Type != "richtext" {
		t.Errorf("GetFieldTypeByType Type = %q, want %q", got.Type, "richtext")
	}
	if got.Label != "Rich Text" {
		t.Errorf("GetFieldTypeByType Label = %q, want %q", got.Label, "Rich Text")
	}

	// Non-existent type should return error
	_, err = d.GetFieldTypeByType("nonexistent")
	if err == nil {
		t.Fatal("GetFieldTypeByType for nonexistent type: expected error, got nil")
	}
}
