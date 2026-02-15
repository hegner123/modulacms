// Integration tests for the permission entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0: no FK dependencies).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Permission(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// --- Count: starts at zero ---
	count, err := d.CountPermissions()
	if err != nil {
		t.Fatalf("initial CountPermissions: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountPermissions = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
		TableID: "test_table",
		Mode:    7,
		Label:   "full_access",
	})
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}
	if created == nil {
		t.Fatal("CreatePermission returned nil")
	}
	if created.PermissionID.IsZero() {
		t.Fatal("CreatePermission returned zero PermissionID")
	}
	if created.TableID != "test_table" {
		t.Errorf("TableID = %q, want %q", created.TableID, "test_table")
	}
	if created.Mode != 7 {
		t.Errorf("Mode = %d, want %d", created.Mode, 7)
	}
	if created.Label != "full_access" {
		t.Errorf("Label = %q, want %q", created.Label, "full_access")
	}

	// --- Get ---
	got, err := d.GetPermission(created.PermissionID)
	if err != nil {
		t.Fatalf("GetPermission: %v", err)
	}
	if got == nil {
		t.Fatal("GetPermission returned nil")
	}
	if got.PermissionID != created.PermissionID {
		t.Errorf("GetPermission ID = %v, want %v", got.PermissionID, created.PermissionID)
	}
	if got.TableID != created.TableID {
		t.Errorf("GetPermission TableID = %q, want %q", got.TableID, created.TableID)
	}
	if got.Mode != created.Mode {
		t.Errorf("GetPermission Mode = %d, want %d", got.Mode, created.Mode)
	}
	if got.Label != created.Label {
		t.Errorf("GetPermission Label = %q, want %q", got.Label, created.Label)
	}

	// --- List ---
	list, err := d.ListPermissions()
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if list == nil {
		t.Fatal("ListPermissions returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListPermissions len = %d, want 1", len(*list))
	}
	if (*list)[0].PermissionID != created.PermissionID {
		t.Errorf("ListPermissions[0].PermissionID = %v, want %v", (*list)[0].PermissionID, created.PermissionID)
	}

	// --- Count: now 1 ---
	count, err = d.CountPermissions()
	if err != nil {
		t.Fatalf("CountPermissions after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountPermissions after create = %d, want 1", *count)
	}

	// --- Update ---
	_, err = d.UpdatePermission(ctx, ac, UpdatePermissionParams{
		TableID:      "updated_table",
		Mode:         3,
		Label:        "read_only",
		PermissionID: created.PermissionID,
	})
	if err != nil {
		t.Fatalf("UpdatePermission: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetPermission(created.PermissionID)
	if err != nil {
		t.Fatalf("GetPermission after update: %v", err)
	}
	if updated.TableID != "updated_table" {
		t.Errorf("updated TableID = %q, want %q", updated.TableID, "updated_table")
	}
	if updated.Mode != 3 {
		t.Errorf("updated Mode = %d, want %d", updated.Mode, 3)
	}
	if updated.Label != "read_only" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "read_only")
	}

	// --- Delete ---
	err = d.DeletePermission(ctx, ac, created.PermissionID)
	if err != nil {
		t.Fatalf("DeletePermission: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetPermission(created.PermissionID)
	if err == nil {
		t.Fatal("GetPermission after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountPermissions()
	if err != nil {
		t.Fatalf("CountPermissions after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountPermissions after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_Permission_MultipleRecords verifies that multiple
// permissions can coexist and be listed independently.
func TestDatabase_CRUD_Permission_MultipleRecords(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	labels := []string{"perm_alpha", "perm_beta", "perm_gamma"}
	ids := make([]types.PermissionID, len(labels))

	for i, label := range labels {
		p, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
			TableID: "multi_table",
			Mode:    int64(i + 1),
			Label:   label,
		})
		if err != nil {
			t.Fatalf("CreatePermission(%s): %v", label, err)
		}
		ids[i] = p.PermissionID
	}

	count, err := d.CountPermissions()
	if err != nil {
		t.Fatalf("CountPermissions: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountPermissions = %d, want 3", *count)
	}

	list, err := d.ListPermissions()
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(*list) != 3 {
		t.Fatalf("ListPermissions len = %d, want 3", len(*list))
	}

	// Delete one and verify count drops
	err = d.DeletePermission(ctx, ac, ids[1])
	if err != nil {
		t.Fatalf("DeletePermission: %v", err)
	}

	count, err = d.CountPermissions()
	if err != nil {
		t.Fatalf("CountPermissions after delete: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountPermissions after delete = %d, want 2", *count)
	}
}
