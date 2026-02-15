// Integration tests for the role entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0: no FK dependencies).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Role(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// --- Count: starts at zero ---
	count, err := d.CountRoles()
	if err != nil {
		t.Fatalf("initial CountRoles: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountRoles = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label:       "admin",
		Permissions: `{"read":true,"write":true,"delete":true}`,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if created == nil {
		t.Fatal("CreateRole returned nil")
	}
	if created.RoleID.IsZero() {
		t.Fatal("CreateRole returned zero RoleID")
	}
	if created.Label != "admin" {
		t.Errorf("Label = %q, want %q", created.Label, "admin")
	}
	if created.Permissions != `{"read":true,"write":true,"delete":true}` {
		t.Errorf("Permissions = %q, want %q", created.Permissions, `{"read":true,"write":true,"delete":true}`)
	}

	// --- Get ---
	got, err := d.GetRole(created.RoleID)
	if err != nil {
		t.Fatalf("GetRole: %v", err)
	}
	if got == nil {
		t.Fatal("GetRole returned nil")
	}
	if got.RoleID != created.RoleID {
		t.Errorf("GetRole RoleID = %v, want %v", got.RoleID, created.RoleID)
	}
	if got.Label != created.Label {
		t.Errorf("GetRole Label = %q, want %q", got.Label, created.Label)
	}
	if got.Permissions != created.Permissions {
		t.Errorf("GetRole Permissions = %q, want %q", got.Permissions, created.Permissions)
	}

	// --- List ---
	list, err := d.ListRoles()
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if list == nil {
		t.Fatal("ListRoles returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListRoles len = %d, want 1", len(*list))
	}
	if (*list)[0].RoleID != created.RoleID {
		t.Errorf("ListRoles[0].RoleID = %v, want %v", (*list)[0].RoleID, created.RoleID)
	}

	// --- Count: now 1 ---
	count, err = d.CountRoles()
	if err != nil {
		t.Fatalf("CountRoles after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountRoles after create = %d, want 1", *count)
	}

	// --- Update ---
	_, err = d.UpdateRole(ctx, ac, UpdateRoleParams{
		Label:       "editor",
		Permissions: `{"read":true,"write":true}`,
		RoleID:      created.RoleID,
	})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetRole(created.RoleID)
	if err != nil {
		t.Fatalf("GetRole after update: %v", err)
	}
	if updated.Label != "editor" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "editor")
	}
	if updated.Permissions != `{"read":true,"write":true}` {
		t.Errorf("updated Permissions = %q, want %q", updated.Permissions, `{"read":true,"write":true}`)
	}

	// --- Delete ---
	err = d.DeleteRole(ctx, ac, created.RoleID)
	if err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetRole(created.RoleID)
	if err == nil {
		t.Fatal("GetRole after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountRoles()
	if err != nil {
		t.Fatalf("CountRoles after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountRoles after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_Role_MultipleRecords verifies multiple roles can
// coexist and that deleting one does not affect the others.
func TestDatabase_CRUD_Role_MultipleRecords(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	labels := []string{"role_alpha", "role_beta", "role_gamma"}
	perms := []string{`["read"]`, `["write"]`, `["delete"]`}
	ids := make([]types.RoleID, len(labels))

	for i, label := range labels {
		r, err := d.CreateRole(ctx, ac, CreateRoleParams{
			Label:       label,
			Permissions: perms[i],
		})
		if err != nil {
			t.Fatalf("CreateRole(%s): %v", label, err)
		}
		ids[i] = r.RoleID
	}

	count, err := d.CountRoles()
	if err != nil {
		t.Fatalf("CountRoles: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountRoles = %d, want 3", *count)
	}

	list, err := d.ListRoles()
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(*list) != 3 {
		t.Fatalf("ListRoles len = %d, want 3", len(*list))
	}

	// Delete the middle one
	err = d.DeleteRole(ctx, ac, ids[1])
	if err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	count, err = d.CountRoles()
	if err != nil {
		t.Fatalf("CountRoles after delete: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountRoles after delete = %d, want 2", *count)
	}
}
