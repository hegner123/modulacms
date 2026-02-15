// Integration tests for the user_ssh_keys entity CRUD lifecycle.
// Uses testSeededDB (Tier 1a: SSH keys reference users via NullableUserID).
//
// Special notes:
// - No standard UpdateUserSshKey method; uses UpdateUserSshKeyLastUsed (non-audited)
//   and UpdateUserSshKeyLabel (audited) as separate update operations.
// - SshKeyID is raw string, not a typed ID.
// - Label and LastUsed are mapped from sql.NullString to plain string.
package db

import (
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_UserSshKey(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	userID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at zero ---
	count, err := d.CountUserSshKeys()
	if err != nil {
		t.Fatalf("initial CountUserSshKeys: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountUserSshKeys = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateUserSshKey(ctx, ac, CreateUserSshKeyParams{
		UserID:      userID,
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITestKeyDataHere test@example.com",
		KeyType:     "ed25519",
		Fingerprint: "SHA256:TestFingerprintAbcdef1234567890",
		Label:       "test-laptop",
		DateCreated: now,
	})
	if err != nil {
		t.Fatalf("CreateUserSshKey: %v", err)
	}
	if created == nil {
		t.Fatal("CreateUserSshKey returned nil")
	}
	if created.SshKeyID == "" {
		t.Fatal("CreateUserSshKey returned empty SshKeyID")
	}
	if created.PublicKey != "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITestKeyDataHere test@example.com" {
		t.Errorf("PublicKey = %q, want ssh-ed25519 key", created.PublicKey)
	}
	if created.KeyType != "ed25519" {
		t.Errorf("KeyType = %q, want %q", created.KeyType, "ed25519")
	}
	if created.Fingerprint != "SHA256:TestFingerprintAbcdef1234567890" {
		t.Errorf("Fingerprint = %q, want %q", created.Fingerprint, "SHA256:TestFingerprintAbcdef1234567890")
	}
	if created.Label != "test-laptop" {
		t.Errorf("Label = %q, want %q", created.Label, "test-laptop")
	}
	if !created.UserID.Valid || created.UserID.ID != seed.User.UserID {
		t.Errorf("UserID = %v, want {ID: %v, Valid: true}", created.UserID, seed.User.UserID)
	}

	// --- Get ---
	got, err := d.GetUserSshKey(created.SshKeyID)
	if err != nil {
		t.Fatalf("GetUserSshKey: %v", err)
	}
	if got == nil {
		t.Fatal("GetUserSshKey returned nil")
	}
	if got.SshKeyID != created.SshKeyID {
		t.Errorf("GetUserSshKey ID = %q, want %q", got.SshKeyID, created.SshKeyID)
	}
	if got.Fingerprint != created.Fingerprint {
		t.Errorf("GetUserSshKey Fingerprint = %q, want %q", got.Fingerprint, created.Fingerprint)
	}
	if got.Label != created.Label {
		t.Errorf("GetUserSshKey Label = %q, want %q", got.Label, created.Label)
	}

	// --- GetUserSshKeyByFingerprint ---
	gotByFP, err := d.GetUserSshKeyByFingerprint("SHA256:TestFingerprintAbcdef1234567890")
	if err != nil {
		t.Fatalf("GetUserSshKeyByFingerprint: %v", err)
	}
	if gotByFP == nil {
		t.Fatal("GetUserSshKeyByFingerprint returned nil")
	}
	if gotByFP.SshKeyID != created.SshKeyID {
		t.Errorf("GetUserSshKeyByFingerprint ID = %q, want %q", gotByFP.SshKeyID, created.SshKeyID)
	}

	// --- ListUserSshKeys (by user) ---
	list, err := d.ListUserSshKeys(userID)
	if err != nil {
		t.Fatalf("ListUserSshKeys: %v", err)
	}
	if list == nil {
		t.Fatal("ListUserSshKeys returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListUserSshKeys len = %d, want 1", len(*list))
	}
	if (*list)[0].SshKeyID != created.SshKeyID {
		t.Errorf("ListUserSshKeys[0].SshKeyID = %q, want %q", (*list)[0].SshKeyID, created.SshKeyID)
	}

	// --- Count: now 1 ---
	count, err = d.CountUserSshKeys()
	if err != nil {
		t.Fatalf("CountUserSshKeys after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountUserSshKeys after create = %d, want 1", *count)
	}

	// --- UpdateUserSshKeyLastUsed (non-audited) ---
	lastUsedTime := time.Now().UTC().Format(time.RFC3339)
	err = d.UpdateUserSshKeyLastUsed(created.SshKeyID, lastUsedTime)
	if err != nil {
		t.Fatalf("UpdateUserSshKeyLastUsed: %v", err)
	}

	gotAfterLastUsed, err := d.GetUserSshKey(created.SshKeyID)
	if err != nil {
		t.Fatalf("GetUserSshKey after UpdateLastUsed: %v", err)
	}
	if gotAfterLastUsed.LastUsed == "" {
		t.Error("LastUsed should be non-empty after UpdateUserSshKeyLastUsed")
	}

	// --- UpdateUserSshKeyLabel (audited) ---
	err = d.UpdateUserSshKeyLabel(ctx, ac, created.SshKeyID, "work-desktop")
	if err != nil {
		t.Fatalf("UpdateUserSshKeyLabel: %v", err)
	}

	gotAfterLabel, err := d.GetUserSshKey(created.SshKeyID)
	if err != nil {
		t.Fatalf("GetUserSshKey after UpdateLabel: %v", err)
	}
	if gotAfterLabel.Label != "work-desktop" {
		t.Errorf("updated Label = %q, want %q", gotAfterLabel.Label, "work-desktop")
	}
	// Verify other fields remain unchanged
	if gotAfterLabel.Fingerprint != created.Fingerprint {
		t.Errorf("Fingerprint changed after label update: %q, want %q", gotAfterLabel.Fingerprint, created.Fingerprint)
	}
	if gotAfterLabel.KeyType != created.KeyType {
		t.Errorf("KeyType changed after label update: %q, want %q", gotAfterLabel.KeyType, created.KeyType)
	}

	// --- Delete ---
	err = d.DeleteUserSshKey(ctx, ac, created.SshKeyID)
	if err != nil {
		t.Fatalf("DeleteUserSshKey: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetUserSshKey(created.SshKeyID)
	if err == nil {
		t.Fatal("GetUserSshKey after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountUserSshKeys()
	if err != nil {
		t.Fatalf("CountUserSshKeys after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountUserSshKeys after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_UserSshKey_EmptyLabel verifies that SSH keys can be
// created with an empty label (maps to sql.NullString{Valid: false}).
func TestDatabase_CRUD_UserSshKey_EmptyLabel(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	userID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	created, err := d.CreateUserSshKey(ctx, ac, CreateUserSshKeyParams{
		UserID:      userID,
		PublicKey:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC test@nolabel",
		KeyType:     "rsa",
		Fingerprint: "SHA256:NoLabelFingerprint9876543210",
		Label:       "",
		DateCreated: now,
	})
	if err != nil {
		t.Fatalf("CreateUserSshKey with empty label: %v", err)
	}
	if created == nil {
		t.Fatal("CreateUserSshKey returned nil")
	}

	got, err := d.GetUserSshKey(created.SshKeyID)
	if err != nil {
		t.Fatalf("GetUserSshKey: %v", err)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// TestDatabase_CRUD_UserSshKey_GetByFingerprint_NotFound verifies that
// GetUserSshKeyByFingerprint returns an error for a non-existent fingerprint.
func TestDatabase_CRUD_UserSshKey_GetByFingerprint_NotFound(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	_, err := d.GetUserSshKeyByFingerprint("SHA256:NonExistentFingerprint")
	if err == nil {
		t.Fatal("GetUserSshKeyByFingerprint for nonexistent: expected error, got nil")
	}
}
