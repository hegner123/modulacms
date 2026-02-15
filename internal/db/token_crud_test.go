// Integration tests for the token entity CRUD lifecycle.
// Uses testSeededDB (Tier 1a: tokens reference users via NullableUserID).
package db

import (
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Token(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	userID := types.NullableUserID{ID: seed.User.UserID, Valid: true}
	expiresAt := types.NewTimestamp(time.Now().Add(24 * time.Hour))

	// --- Count: starts at zero ---
	count, err := d.CountTokens()
	if err != nil {
		t.Fatalf("initial CountTokens: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountTokens = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateToken(ctx, ac, CreateTokenParams{
		UserID:    userID,
		TokenType: "refresh",
		Token:     "test-token-value-abc123",
		IssuedAt:  time.Now().UTC().Format(time.RFC3339),
		ExpiresAt: expiresAt,
		Revoked:   false,
	})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if created == nil {
		t.Fatal("CreateToken returned nil")
	}
	if created.ID == "" {
		t.Fatal("CreateToken returned empty ID")
	}
	if created.TokenType != "refresh" {
		t.Errorf("TokenType = %q, want %q", created.TokenType, "refresh")
	}
	if created.Token != "test-token-value-abc123" {
		t.Errorf("Token = %q, want %q", created.Token, "test-token-value-abc123")
	}
	if !created.UserID.Valid || created.UserID.ID != seed.User.UserID {
		t.Errorf("UserID = %v, want {ID: %v, Valid: true}", created.UserID, seed.User.UserID)
	}
	if created.Revoked {
		t.Errorf("Revoked = true, want false")
	}

	// --- Get ---
	got, err := d.GetToken(created.ID)
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if got == nil {
		t.Fatal("GetToken returned nil")
	}
	if got.ID != created.ID {
		t.Errorf("GetToken ID = %q, want %q", got.ID, created.ID)
	}
	if got.Token != created.Token {
		t.Errorf("GetToken Token = %q, want %q", got.Token, created.Token)
	}
	if got.TokenType != created.TokenType {
		t.Errorf("GetToken TokenType = %q, want %q", got.TokenType, created.TokenType)
	}

	// --- GetTokenByTokenValue ---
	gotByValue, err := d.GetTokenByTokenValue("test-token-value-abc123")
	if err != nil {
		t.Fatalf("GetTokenByTokenValue: %v", err)
	}
	if gotByValue == nil {
		t.Fatal("GetTokenByTokenValue returned nil")
	}
	if gotByValue.ID != created.ID {
		t.Errorf("GetTokenByTokenValue ID = %q, want %q", gotByValue.ID, created.ID)
	}

	// --- GetTokenByUserId ---
	gotByUser, err := d.GetTokenByUserId(userID)
	if err != nil {
		t.Fatalf("GetTokenByUserId: %v", err)
	}
	if gotByUser == nil {
		t.Fatal("GetTokenByUserId returned nil")
	}
	if len(*gotByUser) != 1 {
		t.Fatalf("GetTokenByUserId len = %d, want 1", len(*gotByUser))
	}
	if (*gotByUser)[0].ID != created.ID {
		t.Errorf("GetTokenByUserId[0].ID = %q, want %q", (*gotByUser)[0].ID, created.ID)
	}

	// --- List ---
	list, err := d.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if list == nil {
		t.Fatal("ListTokens returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListTokens len = %d, want 1", len(*list))
	}
	if (*list)[0].ID != created.ID {
		t.Errorf("ListTokens[0].ID = %q, want %q", (*list)[0].ID, created.ID)
	}

	// --- Count: now 1 ---
	count, err = d.CountTokens()
	if err != nil {
		t.Fatalf("CountTokens after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountTokens after create = %d, want 1", *count)
	}

	// --- Update ---
	newExpires := types.NewTimestamp(time.Now().Add(48 * time.Hour))
	_, err = d.UpdateToken(ctx, ac, UpdateTokenParams{
		ID:        created.ID,
		Token:     "updated-token-value-xyz789",
		IssuedAt:  created.IssuedAt,
		ExpiresAt: newExpires,
		Revoked:   true,
	})
	if err != nil {
		t.Fatalf("UpdateToken: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetToken(created.ID)
	if err != nil {
		t.Fatalf("GetToken after update: %v", err)
	}
	if updated.Token != "updated-token-value-xyz789" {
		t.Errorf("updated Token = %q, want %q", updated.Token, "updated-token-value-xyz789")
	}
	if !updated.Revoked {
		t.Errorf("updated Revoked = false, want true")
	}

	// --- Delete ---
	err = d.DeleteToken(ctx, ac, created.ID)
	if err != nil {
		t.Fatalf("DeleteToken: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetToken(created.ID)
	if err == nil {
		t.Fatal("GetToken after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountTokens()
	if err != nil {
		t.Fatalf("CountTokens after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountTokens after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_Token_GetByTokenValue_NotFound verifies that
// GetTokenByTokenValue returns an error for a non-existent token value.
func TestDatabase_CRUD_Token_GetByTokenValue_NotFound(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	_, err := d.GetTokenByTokenValue("nonexistent-token-value")
	if err == nil {
		t.Fatal("GetTokenByTokenValue for nonexistent value: expected error, got nil")
	}
}
