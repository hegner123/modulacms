// Integration tests for the session entity CRUD lifecycle.
// Uses testSeededDB (Tier 1a: sessions reference users via NullableUserID).
package db

import (
	"database/sql"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Session(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	later := types.TimestampNow()

	userID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at zero ---
	count, err := d.CountSessions()
	if err != nil {
		t.Fatalf("initial CountSessions: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountSessions = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateSession(ctx, ac, CreateSessionParams{
		UserID:      userID,
		CreatedAt:   now,
		ExpiresAt:   later,
		LastAccess:  sql.NullString{String: "2026-01-01T00:00:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "192.168.1.1", Valid: true},
		UserAgent:   sql.NullString{String: "test-agent/1.0", Valid: true},
		SessionData: sql.NullString{String: "{}", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if created == nil {
		t.Fatal("CreateSession returned nil")
	}
	if created.SessionID.IsZero() {
		t.Fatal("CreateSession returned zero SessionID")
	}
	if !created.UserID.Valid || created.UserID.ID != seed.User.UserID {
		t.Errorf("UserID = %v, want {ID: %v, Valid: true}", created.UserID, seed.User.UserID)
	}
	if !created.IpAddress.Valid || created.IpAddress.String != "192.168.1.1" {
		t.Errorf("IpAddress = %v, want {192.168.1.1, true}", created.IpAddress)
	}
	if !created.UserAgent.Valid || created.UserAgent.String != "test-agent/1.0" {
		t.Errorf("UserAgent = %v, want {test-agent/1.0, true}", created.UserAgent)
	}
	if !created.SessionData.Valid || created.SessionData.String != "{}" {
		t.Errorf("SessionData = %v, want {{}, true}", created.SessionData)
	}

	// --- Get ---
	got, err := d.GetSession(created.SessionID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got == nil {
		t.Fatal("GetSession returned nil")
	}
	if got.SessionID != created.SessionID {
		t.Errorf("GetSession ID = %v, want %v", got.SessionID, created.SessionID)
	}
	if got.IpAddress.String != created.IpAddress.String {
		t.Errorf("GetSession IpAddress = %q, want %q", got.IpAddress.String, created.IpAddress.String)
	}

	// --- GetSessionByUserId ---
	gotByUser, err := d.GetSessionByUserId(userID)
	if err != nil {
		t.Fatalf("GetSessionByUserId: %v", err)
	}
	if gotByUser == nil {
		t.Fatal("GetSessionByUserId returned nil")
	}
	if gotByUser.SessionID != created.SessionID {
		t.Errorf("GetSessionByUserId ID = %v, want %v", gotByUser.SessionID, created.SessionID)
	}

	// --- List ---
	list, err := d.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if list == nil {
		t.Fatal("ListSessions returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListSessions len = %d, want 1", len(*list))
	}
	if (*list)[0].SessionID != created.SessionID {
		t.Errorf("ListSessions[0].SessionID = %v, want %v", (*list)[0].SessionID, created.SessionID)
	}

	// --- Count: now 1 ---
	count, err = d.CountSessions()
	if err != nil {
		t.Fatalf("CountSessions after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountSessions after create = %d, want 1", *count)
	}

	// --- Update ---
	updatedNow := types.TimestampNow()
	_, err = d.UpdateSession(ctx, ac, UpdateSessionParams{
		SessionID:   created.SessionID,
		UserID:      userID,
		CreatedAt:   created.CreatedAt,
		ExpiresAt:   updatedNow,
		LastAccess:  sql.NullString{String: "2026-02-01T00:00:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "10.0.0.1", Valid: true},
		UserAgent:   sql.NullString{String: "test-agent/2.0", Valid: true},
		SessionData: sql.NullString{String: "{\"updated\": true}", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateSession: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetSession(created.SessionID)
	if err != nil {
		t.Fatalf("GetSession after update: %v", err)
	}
	if !updated.IpAddress.Valid || updated.IpAddress.String != "10.0.0.1" {
		t.Errorf("updated IpAddress = %v, want {10.0.0.1, true}", updated.IpAddress)
	}
	if !updated.UserAgent.Valid || updated.UserAgent.String != "test-agent/2.0" {
		t.Errorf("updated UserAgent = %v, want {test-agent/2.0, true}", updated.UserAgent)
	}
	if !updated.SessionData.Valid || updated.SessionData.String != "{\"updated\": true}" {
		t.Errorf("updated SessionData = %v, want {{\"updated\": true}, true}", updated.SessionData)
	}

	// --- Delete ---
	err = d.DeleteSession(ctx, ac, created.SessionID)
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetSession(created.SessionID)
	if err == nil {
		t.Fatal("GetSession after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountSessions()
	if err != nil {
		t.Fatalf("CountSessions after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountSessions after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_Session_NullOptionalFields verifies that sessions can
// be created with null optional fields (LastAccess, IpAddress, UserAgent, SessionData).
func TestDatabase_CRUD_Session_NullOptionalFields(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	userID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	created, err := d.CreateSession(ctx, ac, CreateSessionParams{
		UserID:      userID,
		CreatedAt:   now,
		ExpiresAt:   now,
		LastAccess:  sql.NullString{},
		IpAddress:   sql.NullString{},
		UserAgent:   sql.NullString{},
		SessionData: sql.NullString{},
	})
	if err != nil {
		t.Fatalf("CreateSession with null optionals: %v", err)
	}
	if created == nil {
		t.Fatal("CreateSession returned nil")
	}

	got, err := d.GetSession(created.SessionID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.LastAccess.Valid {
		t.Errorf("LastAccess should be invalid (null), got %v", got.LastAccess)
	}
	if got.IpAddress.Valid {
		t.Errorf("IpAddress should be invalid (null), got %v", got.IpAddress)
	}
	if got.UserAgent.Valid {
		t.Errorf("UserAgent should be invalid (null), got %v", got.UserAgent)
	}
	if got.SessionData.Valid {
		t.Errorf("SessionData should be invalid (null), got %v", got.SessionData)
	}
}
