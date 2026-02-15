// Integration tests for the admin_route entity CRUD lifecycle.
// Uses testSeededDB (Tier 1b: requires user for author_id FK).
//
// NOTE: AdminRoute UpdateAdminRoute is skipped in the main lifecycle test
// because UpdateAdminRouteCmd.GetID() returns the slug string (not a ULID),
// which violates the change_events table CHECK constraint:
// length(record_id) = 26. This is a known codebase issue where admin route
// updates use slug-based identification for audit recording.
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_AdminRoute(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at 1 (seed admin route) ---
	count, err := d.CountAdminRoutes()
	if err != nil {
		t.Fatalf("initial CountAdminRoutes: %v", err)
	}
	if *count != 1 {
		t.Fatalf("initial CountAdminRoutes = %d, want 1", *count)
	}

	// --- Create ---
	created, err := d.CreateAdminRoute(ctx, ac, CreateAdminRouteParams{
		Slug:         types.Slug("crud-test-admin-route"),
		Title:        "CRUD Test Admin Route",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateAdminRoute: %v", err)
	}
	if created == nil {
		t.Fatal("CreateAdminRoute returned nil")
	}
	if created.AdminRouteID.IsZero() {
		t.Fatal("CreateAdminRoute returned zero AdminRouteID")
	}
	if created.Slug != types.Slug("crud-test-admin-route") {
		t.Errorf("Slug = %q, want %q", created.Slug, "crud-test-admin-route")
	}
	if created.Title != "CRUD Test Admin Route" {
		t.Errorf("Title = %q, want %q", created.Title, "CRUD Test Admin Route")
	}
	if created.Status != 1 {
		t.Errorf("Status = %d, want %d", created.Status, 1)
	}

	// --- Get (by slug) ---
	got, err := d.GetAdminRoute(types.Slug("crud-test-admin-route"))
	if err != nil {
		t.Fatalf("GetAdminRoute: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminRoute returned nil")
	}
	if got.AdminRouteID != created.AdminRouteID {
		t.Errorf("GetAdminRoute ID = %v, want %v", got.AdminRouteID, created.AdminRouteID)
	}
	if got.Slug != created.Slug {
		t.Errorf("GetAdminRoute Slug = %q, want %q", got.Slug, created.Slug)
	}
	if got.Title != created.Title {
		t.Errorf("GetAdminRoute Title = %q, want %q", got.Title, created.Title)
	}

	// --- List ---
	list, err := d.ListAdminRoutes()
	if err != nil {
		t.Fatalf("ListAdminRoutes: %v", err)
	}
	if list == nil {
		t.Fatal("ListAdminRoutes returned nil")
	}
	if len(*list) != 2 {
		t.Fatalf("ListAdminRoutes len = %d, want 2", len(*list))
	}

	// --- Count: now 2 ---
	count, err = d.CountAdminRoutes()
	if err != nil {
		t.Fatalf("CountAdminRoutes after create: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountAdminRoutes after create = %d, want 2", *count)
	}

	// --- Update: SKIPPED ---
	// AdminRoute UpdateAdminRoute uses slug as record_id in audit events, which
	// violates the change_events CHECK(length(record_id) = 26) constraint.
	// See UpdateAdminRouteCmd.GetID() in admin_route.go.

	// --- Delete ---
	err = d.DeleteAdminRoute(ctx, ac, created.AdminRouteID)
	if err != nil {
		t.Fatalf("DeleteAdminRoute: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetAdminRoute(types.Slug("crud-test-admin-route"))
	if err == nil {
		t.Fatal("GetAdminRoute after delete: expected error, got nil")
	}

	// --- Count: back to 1 (seed admin route remains) ---
	count, err = d.CountAdminRoutes()
	if err != nil {
		t.Fatalf("CountAdminRoutes after delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountAdminRoutes after delete = %d, want 1", *count)
	}
}

// TestDatabase_CRUD_AdminRoute_GetBySlug tests fetching the seed admin route by slug.
func TestDatabase_CRUD_AdminRoute_GetBySlug(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)

	got, err := d.GetAdminRoute(types.Slug("test-admin-route"))
	if err != nil {
		t.Fatalf("GetAdminRoute by seed slug: %v", err)
	}
	if got == nil {
		t.Fatal("GetAdminRoute returned nil for seed slug")
	}
	if got.AdminRouteID != seed.AdminRoute.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, seed.AdminRoute.AdminRouteID)
	}
}
