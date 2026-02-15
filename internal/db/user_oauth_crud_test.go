// Integration tests for the user_oauth entity CRUD lifecycle.
// Uses testSeededDB (Tier 1a: user_oauth references users via NullableUserID).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_UserOauth(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	userID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at zero ---
	count, err := d.CountUserOauths()
	if err != nil {
		t.Fatalf("initial CountUserOauths: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountUserOauths = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateUserOauth(ctx, ac, CreateUserOauthParams{
		UserID:              userID,
		OauthProvider:       "github",
		OauthProviderUserID: "gh-user-12345",
		AccessToken:         "access-token-abc",
		RefreshToken:        "refresh-token-xyz",
		TokenExpiresAt:      "2026-12-31T23:59:59Z",
		DateCreated:         now,
	})
	if err != nil {
		t.Fatalf("CreateUserOauth: %v", err)
	}
	if created == nil {
		t.Fatal("CreateUserOauth returned nil")
	}
	if created.UserOauthID.IsZero() {
		t.Fatal("CreateUserOauth returned zero UserOauthID")
	}
	if created.OauthProvider != "github" {
		t.Errorf("OauthProvider = %q, want %q", created.OauthProvider, "github")
	}
	if created.OauthProviderUserID != "gh-user-12345" {
		t.Errorf("OauthProviderUserID = %q, want %q", created.OauthProviderUserID, "gh-user-12345")
	}
	if created.AccessToken != "access-token-abc" {
		t.Errorf("AccessToken = %q, want %q", created.AccessToken, "access-token-abc")
	}
	if created.RefreshToken != "refresh-token-xyz" {
		t.Errorf("RefreshToken = %q, want %q", created.RefreshToken, "refresh-token-xyz")
	}
	if created.TokenExpiresAt != "2026-12-31T23:59:59Z" {
		t.Errorf("TokenExpiresAt = %q, want %q", created.TokenExpiresAt, "2026-12-31T23:59:59Z")
	}
	if !created.UserID.Valid || created.UserID.ID != seed.User.UserID {
		t.Errorf("UserID = %v, want {ID: %v, Valid: true}", created.UserID, seed.User.UserID)
	}

	// --- Get ---
	got, err := d.GetUserOauth(created.UserOauthID)
	if err != nil {
		t.Fatalf("GetUserOauth: %v", err)
	}
	if got == nil {
		t.Fatal("GetUserOauth returned nil")
	}
	if got.UserOauthID != created.UserOauthID {
		t.Errorf("GetUserOauth ID = %v, want %v", got.UserOauthID, created.UserOauthID)
	}
	if got.OauthProvider != created.OauthProvider {
		t.Errorf("GetUserOauth OauthProvider = %q, want %q", got.OauthProvider, created.OauthProvider)
	}
	if got.AccessToken != created.AccessToken {
		t.Errorf("GetUserOauth AccessToken = %q, want %q", got.AccessToken, created.AccessToken)
	}

	// --- GetUserOauthByUserId ---
	gotByUser, err := d.GetUserOauthByUserId(userID)
	if err != nil {
		t.Fatalf("GetUserOauthByUserId: %v", err)
	}
	if gotByUser == nil {
		t.Fatal("GetUserOauthByUserId returned nil")
	}
	if gotByUser.UserOauthID != created.UserOauthID {
		t.Errorf("GetUserOauthByUserId ID = %v, want %v", gotByUser.UserOauthID, created.UserOauthID)
	}

	// --- GetUserOauthByProviderID ---
	gotByProvider, err := d.GetUserOauthByProviderID("github", "gh-user-12345")
	if err != nil {
		t.Fatalf("GetUserOauthByProviderID: %v", err)
	}
	if gotByProvider == nil {
		t.Fatal("GetUserOauthByProviderID returned nil")
	}
	if gotByProvider.UserOauthID != created.UserOauthID {
		t.Errorf("GetUserOauthByProviderID ID = %v, want %v", gotByProvider.UserOauthID, created.UserOauthID)
	}

	// --- List ---
	list, err := d.ListUserOauths()
	if err != nil {
		t.Fatalf("ListUserOauths: %v", err)
	}
	if list == nil {
		t.Fatal("ListUserOauths returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListUserOauths len = %d, want 1", len(*list))
	}
	if (*list)[0].UserOauthID != created.UserOauthID {
		t.Errorf("ListUserOauths[0].UserOauthID = %v, want %v", (*list)[0].UserOauthID, created.UserOauthID)
	}

	// --- Count: now 1 ---
	count, err = d.CountUserOauths()
	if err != nil {
		t.Fatalf("CountUserOauths after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountUserOauths after create = %d, want 1", *count)
	}

	// --- Update (only AccessToken, RefreshToken, TokenExpiresAt) ---
	_, err = d.UpdateUserOauth(ctx, ac, UpdateUserOauthParams{
		UserOauthID:    created.UserOauthID,
		AccessToken:    "updated-access-token",
		RefreshToken:   "updated-refresh-token",
		TokenExpiresAt: "2027-06-30T12:00:00Z",
	})
	if err != nil {
		t.Fatalf("UpdateUserOauth: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetUserOauth(created.UserOauthID)
	if err != nil {
		t.Fatalf("GetUserOauth after update: %v", err)
	}
	if updated.AccessToken != "updated-access-token" {
		t.Errorf("updated AccessToken = %q, want %q", updated.AccessToken, "updated-access-token")
	}
	if updated.RefreshToken != "updated-refresh-token" {
		t.Errorf("updated RefreshToken = %q, want %q", updated.RefreshToken, "updated-refresh-token")
	}
	if updated.TokenExpiresAt != "2027-06-30T12:00:00Z" {
		t.Errorf("updated TokenExpiresAt = %q, want %q", updated.TokenExpiresAt, "2027-06-30T12:00:00Z")
	}
	// Verify immutable fields remain unchanged
	if updated.OauthProvider != "github" {
		t.Errorf("updated OauthProvider = %q, want %q (should be unchanged)", updated.OauthProvider, "github")
	}
	if updated.OauthProviderUserID != "gh-user-12345" {
		t.Errorf("updated OauthProviderUserID = %q, want %q (should be unchanged)", updated.OauthProviderUserID, "gh-user-12345")
	}

	// --- Delete ---
	err = d.DeleteUserOauth(ctx, ac, created.UserOauthID)
	if err != nil {
		t.Fatalf("DeleteUserOauth: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetUserOauth(created.UserOauthID)
	if err == nil {
		t.Fatal("GetUserOauth after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountUserOauths()
	if err != nil {
		t.Fatalf("CountUserOauths after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountUserOauths after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_UserOauth_GetByProviderID_NotFound verifies that
// GetUserOauthByProviderID returns an error for a non-existent provider/user combo.
func TestDatabase_CRUD_UserOauth_GetByProviderID_NotFound(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	_, err := d.GetUserOauthByProviderID("nonexistent-provider", "nonexistent-user")
	if err == nil {
		t.Fatal("GetUserOauthByProviderID for nonexistent combo: expected error, got nil")
	}
}
