// Integration tests for the user entity CRUD lifecycle.
// Uses testSeededDB (Tier 1a: needs Role seed for user creation).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_User(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	// --- Count: seed has 1 user ---
	count, err := d.CountUsers()
	if err != nil {
		t.Fatalf("initial CountUsers: %v", err)
	}
	if *count != 1 {
		t.Fatalf("initial CountUsers = %d, want 1", *count)
	}

	// --- Create ---
	created, err := d.CreateUser(ctx, ac, CreateUserParams{
		Username:     "cruduser",
		Name:         "CRUD Test User",
		Email:        types.Email("crud@example.com"),
		Hash:         "crudhash",
		Role:         seed.Role.RoleID.String(),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created == nil {
		t.Fatal("CreateUser returned nil")
	}
	if created.UserID.IsZero() {
		t.Fatal("CreateUser returned zero UserID")
	}
	if created.Username != "cruduser" {
		t.Errorf("Username = %q, want %q", created.Username, "cruduser")
	}
	if created.Name != "CRUD Test User" {
		t.Errorf("Name = %q, want %q", created.Name, "CRUD Test User")
	}
	if created.Email != types.Email("crud@example.com") {
		t.Errorf("Email = %q, want %q", created.Email, "crud@example.com")
	}
	if created.Hash != "crudhash" {
		t.Errorf("Hash = %q, want %q", created.Hash, "crudhash")
	}
	if created.Role != seed.Role.RoleID.String() {
		t.Errorf("Role = %q, want %q", created.Role, seed.Role.RoleID.String())
	}

	// --- Get ---
	got, err := d.GetUser(created.UserID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got == nil {
		t.Fatal("GetUser returned nil")
	}
	if got.UserID != created.UserID {
		t.Errorf("GetUser ID = %v, want %v", got.UserID, created.UserID)
	}
	if got.Username != created.Username {
		t.Errorf("GetUser Username = %q, want %q", got.Username, created.Username)
	}
	if got.Email != created.Email {
		t.Errorf("GetUser Email = %q, want %q", got.Email, created.Email)
	}

	// --- GetUserByEmail ---
	gotByEmail, err := d.GetUserByEmail(types.Email("crud@example.com"))
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if gotByEmail == nil {
		t.Fatal("GetUserByEmail returned nil")
	}
	if gotByEmail.UserID != created.UserID {
		t.Errorf("GetUserByEmail ID = %v, want %v", gotByEmail.UserID, created.UserID)
	}

	// --- List ---
	list, err := d.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if list == nil {
		t.Fatal("ListUsers returned nil")
	}
	if len(*list) != 2 {
		t.Fatalf("ListUsers len = %d, want 2", len(*list))
	}

	// --- Count: now 2 ---
	count, err = d.CountUsers()
	if err != nil {
		t.Fatalf("CountUsers after create: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountUsers after create = %d, want 2", *count)
	}

	// --- Update ---
	updatedNow := types.TimestampNow()
	_, err = d.UpdateUser(ctx, ac, UpdateUserParams{
		UserID:       created.UserID,
		Username:     "cruduser-updated",
		Name:         "Updated CRUD User",
		Email:        types.Email("crud-updated@example.com"),
		Hash:         "updatedhash",
		Role:         seed.Role.RoleID.String(),
		DateCreated:  created.DateCreated,
		DateModified: updatedNow,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetUser(created.UserID)
	if err != nil {
		t.Fatalf("GetUser after update: %v", err)
	}
	if updated.Username != "cruduser-updated" {
		t.Errorf("updated Username = %q, want %q", updated.Username, "cruduser-updated")
	}
	if updated.Name != "Updated CRUD User" {
		t.Errorf("updated Name = %q, want %q", updated.Name, "Updated CRUD User")
	}
	if updated.Email != types.Email("crud-updated@example.com") {
		t.Errorf("updated Email = %q, want %q", updated.Email, "crud-updated@example.com")
	}
	if updated.Hash != "updatedhash" {
		t.Errorf("updated Hash = %q, want %q", updated.Hash, "updatedhash")
	}

	// --- Delete ---
	err = d.DeleteUser(ctx, ac, created.UserID)
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetUser(created.UserID)
	if err == nil {
		t.Fatal("GetUser after delete: expected error, got nil")
	}

	// --- Count: back to 1 (seed user remains) ---
	count, err = d.CountUsers()
	if err != nil {
		t.Fatalf("CountUsers after delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountUsers after delete = %d, want 1", *count)
	}
}

// TestDatabase_CRUD_User_GetByEmail_NotFound verifies that GetUserByEmail
// returns an error for a non-existent email.
func TestDatabase_CRUD_User_GetByEmail_NotFound(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	_, err := d.GetUserByEmail(types.Email("nonexistent@example.com"))
	if err == nil {
		t.Fatal("GetUserByEmail for nonexistent email: expected error, got nil")
	}
}
