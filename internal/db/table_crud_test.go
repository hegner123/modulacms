// Integration tests for the table entity CRUD lifecycle.
// Uses testSeededDB (Tier 1a: tables reference users via NullableUserID for author_id).
//
// Special notes:
// - CreateTableParams only has Label (very minimal).
// - ID is raw string, not a typed ID.
// - AuthorID is set by the audited command layer, not by CreateTableParams.
package db

import (
	"testing"
)

func TestDatabase_CRUD_Table(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// --- Count: starts at zero ---
	count, err := d.CountTables()
	if err != nil {
		t.Fatalf("initial CountTables: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountTables = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateTable(ctx, ac, CreateTableParams{
		Label: "test_content",
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	if created == nil {
		t.Fatal("CreateTable returned nil")
	}
	if created.ID == "" {
		t.Fatal("CreateTable returned empty ID")
	}
	if created.Label != "test_content" {
		t.Errorf("Label = %q, want %q", created.Label, "test_content")
	}

	// --- Get ---
	got, err := d.GetTable(created.ID)
	if err != nil {
		t.Fatalf("GetTable: %v", err)
	}
	if got == nil {
		t.Fatal("GetTable returned nil")
	}
	if got.ID != created.ID {
		t.Errorf("GetTable ID = %q, want %q", got.ID, created.ID)
	}
	if got.Label != created.Label {
		t.Errorf("GetTable Label = %q, want %q", got.Label, created.Label)
	}

	// --- List ---
	list, err := d.ListTables()
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	if list == nil {
		t.Fatal("ListTables returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListTables len = %d, want 1", len(*list))
	}
	if (*list)[0].ID != created.ID {
		t.Errorf("ListTables[0].ID = %q, want %q", (*list)[0].ID, created.ID)
	}

	// --- Count: now 1 ---
	count, err = d.CountTables()
	if err != nil {
		t.Fatalf("CountTables after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountTables after create = %d, want 1", *count)
	}

	// --- Update ---
	_, err = d.UpdateTable(ctx, ac, UpdateTableParams{
		ID:    created.ID,
		Label: "updated_content",
	})
	if err != nil {
		t.Fatalf("UpdateTable: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetTable(created.ID)
	if err != nil {
		t.Fatalf("GetTable after update: %v", err)
	}
	if updated.Label != "updated_content" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "updated_content")
	}

	// --- Delete ---
	err = d.DeleteTable(ctx, ac, created.ID)
	if err != nil {
		t.Fatalf("DeleteTable: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetTable(created.ID)
	if err == nil {
		t.Fatal("GetTable after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountTables()
	if err != nil {
		t.Fatalf("CountTables after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountTables after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_Table_MultipleRecords verifies that multiple tables
// can coexist and be listed independently.
func TestDatabase_CRUD_Table_MultipleRecords(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	labels := []string{"table_alpha", "table_beta", "table_gamma"}
	ids := make([]string, len(labels))

	for i, label := range labels {
		tbl, err := d.CreateTable(ctx, ac, CreateTableParams{
			Label: label,
		})
		if err != nil {
			t.Fatalf("CreateTable(%s): %v", label, err)
		}
		ids[i] = tbl.ID
	}

	count, err := d.CountTables()
	if err != nil {
		t.Fatalf("CountTables: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountTables = %d, want 3", *count)
	}

	list, err := d.ListTables()
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	if len(*list) != 3 {
		t.Fatalf("ListTables len = %d, want 3", len(*list))
	}

	// Delete one and verify count drops
	err = d.DeleteTable(ctx, ac, ids[1])
	if err != nil {
		t.Fatalf("DeleteTable: %v", err)
	}

	count, err = d.CountTables()
	if err != nil {
		t.Fatalf("CountTables after delete: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountTables after delete = %d, want 2", *count)
	}
}
