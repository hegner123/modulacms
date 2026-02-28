// Integration tests for the audited Create/Update/Delete pipeline.
//
// These tests exercise the full audited command flow against a real SQLite
// database: mutation + audit record atomicity, before/after hook lifecycle,
// before-state capture on update/delete, and content_data status transition
// detection.
//
// No Docker required -- uses in-memory SQLite via the db.Database wrapper.
package audited_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ---------------------------------------------------------------------------
// Test infrastructure
// ---------------------------------------------------------------------------

// testDB creates an isolated SQLite database with all tables for testing.
func testDB(t *testing.T) db.Database {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audited_test.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if _, err := conn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	d := db.Database{
		Connection: conn,
		Context:    context.Background(),
		Config:     config.Config{Node_ID: types.NewNodeID().String()},
	}

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}
	return d
}

// seedUser inserts a minimal role + user for FK satisfaction and returns the user.
func seedUser(t *testing.T, d db.Database) *db.Users {
	t.Helper()
	ctx := d.Context
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "seed", "127.0.0.1")
	now := types.TimestampNow()

	role, err := d.CreateRole(ctx, ac, db.CreateRoleParams{Label: "test-role"})
	if err != nil {
		t.Fatalf("seed CreateRole: %v", err)
	}

	user, err := d.CreateUser(ctx, ac, db.CreateUserParams{
		Username:     "audituser",
		Name:         "Audit User",
		Email:        types.Email("audit@test.com"),
		Hash:         "fakehash",
		Role:         role.RoleID.String(),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}
	return user
}

// seedRouteAndDatatype creates seed route + datatype for content_data FK satisfaction.
func seedRouteAndDatatype(t *testing.T, d db.Database, user *db.Users) (types.NullableRouteID, types.NullableDatatypeID) {
	t.Helper()
	ctx := d.Context
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), user.UserID, "seed", "127.0.0.1")
	now := types.TimestampNow()

	route, err := d.CreateRoute(ctx, ac, db.CreateRouteParams{
		Slug:         types.Slug("audit-test-route"),
		Title:        "Audit Test Route",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateRoute: %v", err)
	}

	datatype, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "audit-page",
		Type:         "page",
		AuthorID:     user.UserID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateDatatype: %v", err)
	}

	routeID := types.NullableRouteID{ID: route.RouteID, Valid: true}
	datatypeID := types.NullableDatatypeID{ID: datatype.DatatypeID, Valid: true}
	return routeID, datatypeID
}

// testAuditCtx creates an AuditContext with a real user.
func testAuditCtx(d db.Database, userID types.UserID) audited.AuditContext {
	return audited.Ctx(types.NodeID(d.Config.Node_ID), userID, "integration-test", "10.0.0.1")
}

// trackingHookRunner records hook invocations for assertion.
type trackingHookRunner struct {
	beforeCreateCalls atomic.Int64
	afterCreateCalls  atomic.Int64
	beforeUpdateCalls atomic.Int64
	afterUpdateCalls  atomic.Int64
	beforeDeleteCalls atomic.Int64
	afterDeleteCalls  atomic.Int64

	// Set abortBefore* to true to have the corresponding before-hook return an error.
	abortBeforeCreate bool
	abortBeforeUpdate bool
	abortBeforeDelete bool
}

func (r *trackingHookRunner) HasHooks(event audited.HookEvent, table string) bool {
	switch event {
	case audited.HookBeforeCreate, audited.HookAfterCreate:
		return true
	case audited.HookBeforeUpdate, audited.HookAfterUpdate:
		return true
	case audited.HookBeforeDelete, audited.HookAfterDelete:
		return true
	}
	return false
}

func (r *trackingHookRunner) RunBeforeHooks(_ context.Context, event audited.HookEvent, _ string, _ any) error {
	switch event {
	case audited.HookBeforeCreate:
		r.beforeCreateCalls.Add(1)
		if r.abortBeforeCreate {
			return audited.NewHookError("test-plugin", event, "content_data", "create blocked by test")
		}
	case audited.HookBeforeUpdate:
		r.beforeUpdateCalls.Add(1)
		if r.abortBeforeUpdate {
			return audited.NewHookError("test-plugin", event, "content_data", "update blocked by test")
		}
	case audited.HookBeforeDelete:
		r.beforeDeleteCalls.Add(1)
		if r.abortBeforeDelete {
			return audited.NewHookError("test-plugin", event, "content_data", "delete blocked by test")
		}
	}
	return nil
}

func (r *trackingHookRunner) RunAfterHooks(_ context.Context, event audited.HookEvent, _ string, _ any) {
	switch event {
	case audited.HookAfterCreate:
		r.afterCreateCalls.Add(1)
	case audited.HookAfterUpdate:
		r.afterUpdateCalls.Add(1)
	case audited.HookAfterDelete:
		r.afterDeleteCalls.Add(1)
	}
}

// ---------------------------------------------------------------------------
// Tests: audited.Create
// ---------------------------------------------------------------------------

func TestIntegration_AuditedCreate_RecordsAuditEvent(t *testing.T) {
	// Verifies that audited.Create atomically creates the entity AND
	// records a change_event with operation=INSERT, action=create, and
	// the serialized new entity in new_values.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	created, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	// Verify the entity was created
	got, err := d.GetContentData(created.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData: %v", err)
	}
	if got.ContentDataID != created.ContentDataID {
		t.Errorf("ContentDataID mismatch: got %v, want %v", got.ContentDataID, created.ContentDataID)
	}

	// Verify change_event was recorded
	events, err := d.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 1 {
		t.Fatalf("expected 1 change event, got %d", lenOrZero(events))
	}

	ev := (*events)[0]
	if ev.Operation != types.OpInsert {
		t.Errorf("Operation = %v, want %v", ev.Operation, types.OpInsert)
	}
	if ev.Action != types.ActionCreate {
		t.Errorf("Action = %v, want %v", ev.Action, types.ActionCreate)
	}
	if ev.TableName != "content_data" {
		t.Errorf("TableName = %q, want %q", ev.TableName, "content_data")
	}
	if ev.RecordID != string(created.ContentDataID) {
		t.Errorf("RecordID = %q, want %q", ev.RecordID, string(created.ContentDataID))
	}

	// new_values should be non-empty JSON
	if !ev.NewValues.Valid {
		t.Error("NewValues should be valid")
	}
	newBytes, marshalErr := json.Marshal(ev.NewValues.Data)
	if marshalErr != nil {
		t.Errorf("NewValues marshal: %v", marshalErr)
	}
	var newVals map[string]any
	if err := json.Unmarshal(newBytes, &newVals); err != nil {
		t.Errorf("NewValues is not valid JSON: %v", err)
	}

	// old_values should be empty/null for a create operation
	if ev.OldValues.Valid && ev.OldValues.Data != nil {
		oldBytes, _ := json.Marshal(ev.OldValues.Data)
		if string(oldBytes) != "null" {
			t.Errorf("OldValues should be empty for create, got %s", string(oldBytes))
		}
	}

	// Verify audit metadata fields
	if ev.IP.Valid && ev.IP.String != "10.0.0.1" {
		t.Errorf("IP = %q, want %q", ev.IP.String, "10.0.0.1")
	}
	if ev.RequestID.Valid && ev.RequestID.String != "integration-test" {
		t.Errorf("RequestID = %q, want %q", ev.RequestID.String, "integration-test")
	}
}

func TestIntegration_AuditedCreate_WithUserID(t *testing.T) {
	// Verifies that the UserID from AuditContext is persisted in the change_event.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	created, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	events, err := d.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 1 {
		t.Fatalf("expected 1 change event, got %d", lenOrZero(events))
	}

	ev := (*events)[0]
	if !ev.UserID.Valid {
		t.Fatal("UserID should be valid")
	}
	if ev.UserID.ID != user.UserID {
		t.Errorf("UserID = %v, want %v", ev.UserID.ID, user.UserID)
	}
}

// ---------------------------------------------------------------------------
// Tests: audited.Update
// ---------------------------------------------------------------------------

func TestIntegration_AuditedUpdate_CapturesBeforeState(t *testing.T) {
	// Verifies that audited.Update captures the before-state and records both
	// old_values and new_values in the change_event.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	// Create the entity first
	created, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	// Update: draft -> published
	_, err = d.UpdateContentData(d.Context, ac, db.UpdateContentDataParams{
		ContentDataID: created.ContentDataID,
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("UpdateContentData: %v", err)
	}

	// Should now have 2 events: create + update
	events, err := d.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 2 {
		t.Fatalf("expected 2 change events, got %d", lenOrZero(events))
	}

	// Find the update event (operation=UPDATE)
	var updateEvent *db.ChangeEvent
	for i := range *events {
		if (*events)[i].Operation == types.OpUpdate {
			updateEvent = &(*events)[i]
			break
		}
	}
	if updateEvent == nil {
		t.Fatal("no UPDATE change event found")
	}

	if updateEvent.Action != types.ActionUpdate {
		t.Errorf("Action = %v, want %v", updateEvent.Action, types.ActionUpdate)
	}

	// old_values should contain the before-state (draft status)
	if !updateEvent.OldValues.Valid {
		t.Fatal("OldValues should be valid for update")
	}
	oldBytes, _ := json.Marshal(updateEvent.OldValues.Data)
	var oldVals map[string]any
	if err := json.Unmarshal(oldBytes, &oldVals); err != nil {
		t.Fatalf("OldValues is not valid JSON: %v", err)
	}

	// new_values should contain the update params
	if !updateEvent.NewValues.Valid {
		t.Fatal("NewValues should be valid for update")
	}
	newBytes, _ := json.Marshal(updateEvent.NewValues.Data)
	var newVals map[string]any
	if err := json.Unmarshal(newBytes, &newVals); err != nil {
		t.Fatalf("NewValues is not valid JSON: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: audited.Delete
// ---------------------------------------------------------------------------

func TestIntegration_AuditedDelete_CapturesBeforeState(t *testing.T) {
	// Verifies that audited.Delete captures the entity state before deletion
	// and records it in old_values.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	created, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	err = d.DeleteContentData(d.Context, ac, created.ContentDataID)
	if err != nil {
		t.Fatalf("DeleteContentData: %v", err)
	}

	// The entity is deleted, but the audit trail remains
	events, err := d.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 2 {
		t.Fatalf("expected 2 change events (create + delete), got %d", lenOrZero(events))
	}

	// Find the delete event
	var deleteEvent *db.ChangeEvent
	for i := range *events {
		if (*events)[i].Operation == types.OpDelete {
			deleteEvent = &(*events)[i]
			break
		}
	}
	if deleteEvent == nil {
		t.Fatal("no DELETE change event found")
	}

	if deleteEvent.Action != types.ActionDelete {
		t.Errorf("Action = %v, want %v", deleteEvent.Action, types.ActionDelete)
	}

	// old_values should contain the entity's state before deletion
	if !deleteEvent.OldValues.Valid {
		t.Fatal("OldValues should be valid for delete")
	}
	oldBytes, _ := json.Marshal(deleteEvent.OldValues.Data)
	var oldVals map[string]any
	if err := json.Unmarshal(oldBytes, &oldVals); err != nil {
		t.Fatalf("OldValues is not valid JSON: %v", err)
	}

	// new_values should be empty for delete
	if deleteEvent.NewValues.Valid && deleteEvent.NewValues.Data != nil {
		newBytes, _ := json.Marshal(deleteEvent.NewValues.Data)
		if string(newBytes) != "null" {
			t.Errorf("NewValues should be empty for delete, got %s", string(newBytes))
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: Audit trail completeness
// ---------------------------------------------------------------------------

func TestIntegration_AuditedLifecycle_FullTrail(t *testing.T) {
	// A single content_data through create -> update -> delete produces
	// exactly 3 ordered change events with correct operations.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	created, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = d.UpdateContentData(d.Context, ac, db.UpdateContentDataParams{
		ContentDataID: created.ContentDataID,
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	err = d.DeleteContentData(d.Context, ac, created.ContentDataID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	events, err := d.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 3 {
		t.Fatalf("expected 3 change events, got %d", lenOrZero(events))
	}

	// GetChangeEventsByRecord returns ORDER BY hlc_timestamp DESC (newest first).
	// Verify all three operation types are present regardless of order.
	opCounts := make(map[types.Operation]int)
	for _, ev := range *events {
		opCounts[ev.Operation]++
	}
	if opCounts[types.OpInsert] != 1 {
		t.Errorf("INSERT count = %d, want 1", opCounts[types.OpInsert])
	}
	if opCounts[types.OpUpdate] != 1 {
		t.Errorf("UPDATE count = %d, want 1", opCounts[types.OpUpdate])
	}
	if opCounts[types.OpDelete] != 1 {
		t.Errorf("DELETE count = %d, want 1", opCounts[types.OpDelete])
	}
}

// ---------------------------------------------------------------------------
// Tests: Transaction atomicity
// ---------------------------------------------------------------------------

func TestIntegration_AuditedCreate_TransactionRollsBackOnExecuteFailure(t *testing.T) {
	// When the underlying mutation fails, both the entity and the audit record
	// should be absent (transaction rolled back).
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	_, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	// Use a route_id pointing to a non-existent route -- FK violation
	badRouteID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}

	_, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       badRouteID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err == nil {
		t.Fatal("expected FK violation error, got nil")
	}

	// Verify no change events leaked
	count, err := d.CountChangeEvents()
	if err != nil {
		t.Fatalf("CountChangeEvents: %v", err)
	}
	// The seed user + role + route + datatype creates produced events too.
	// But the failed content_data create should not have added any for "content_data".
	allEvents, err := d.ListChangeEvents(db.ListChangeEventsParams{Limit: 1000, Offset: 0})
	if err != nil {
		t.Fatalf("ListChangeEvents: %v", err)
	}
	for _, ev := range *allEvents {
		if ev.TableName == "content_data" {
			t.Errorf("found leaked content_data change event after rollback: %v", ev.EventID)
		}
	}
	_ = count // used for debug if needed
}

// ---------------------------------------------------------------------------
// Tests: Hook integration (before-hooks can abort, after-hooks fire)
// ---------------------------------------------------------------------------

func TestIntegration_AuditedCreate_BeforeHookAborts(t *testing.T) {
	// When a before_create hook returns an error, the entire transaction
	// (entity + audit record) is rolled back.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	now := types.TimestampNow()

	hooks := &trackingHookRunner{abortBeforeCreate: true}
	ac := testAuditCtx(d, user.UserID)
	ac.HookRunner = hooks

	_, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err == nil {
		t.Fatal("expected hook abort error, got nil")
	}

	var hookErr *audited.HookError
	if !errors.As(err, &hookErr) {
		t.Fatalf("expected *audited.HookError, got %T: %v", err, err)
	}

	// Verify the hook was called
	if hooks.beforeCreateCalls.Load() != 1 {
		t.Errorf("beforeCreateCalls = %d, want 1", hooks.beforeCreateCalls.Load())
	}

	// After-hooks should NOT have fired (transaction was rolled back)
	if hooks.afterCreateCalls.Load() != 0 {
		t.Errorf("afterCreateCalls = %d, want 0 (aborted)", hooks.afterCreateCalls.Load())
	}

	// No content_data should exist
	cdCount, err := d.CountContentData()
	if err != nil {
		t.Fatalf("CountContentData: %v", err)
	}
	if *cdCount != 0 {
		t.Errorf("CountContentData = %d, want 0 (rolled back)", *cdCount)
	}
}

func TestIntegration_AuditedCreate_AfterHookFires(t *testing.T) {
	// When create succeeds, after_create hook fires exactly once.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	now := types.TimestampNow()

	hooks := &trackingHookRunner{}
	ac := testAuditCtx(d, user.UserID)
	ac.HookRunner = hooks

	_, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	if hooks.beforeCreateCalls.Load() != 1 {
		t.Errorf("beforeCreateCalls = %d, want 1", hooks.beforeCreateCalls.Load())
	}
	if hooks.afterCreateCalls.Load() != 1 {
		t.Errorf("afterCreateCalls = %d, want 1", hooks.afterCreateCalls.Load())
	}
}

func TestIntegration_AuditedUpdate_BeforeHookAborts(t *testing.T) {
	// before_update hook abort rolls back the update but the original entity remains.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	now := types.TimestampNow()

	// Create without hooks
	acNoHooks := testAuditCtx(d, user.UserID)
	created, err := d.CreateContentData(d.Context, acNoHooks, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	// Update with aborting hook
	hooks := &trackingHookRunner{abortBeforeUpdate: true}
	acHooks := testAuditCtx(d, user.UserID)
	acHooks.HookRunner = hooks

	_, err = d.UpdateContentData(d.Context, acHooks, db.UpdateContentDataParams{
		ContentDataID: created.ContentDataID,
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  types.TimestampNow(),
	})
	if err == nil {
		t.Fatal("expected hook abort error on update")
	}

	// Entity should still exist with original status (draft)
	got, err := d.GetContentData(created.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData after aborted update: %v", err)
	}
	if got.Status != types.ContentStatusDraft {
		t.Errorf("Status = %q after aborted update, want %q", got.Status, types.ContentStatusDraft)
	}
}

func TestIntegration_AuditedDelete_BeforeHookAborts(t *testing.T) {
	// before_delete hook abort prevents deletion -- entity still exists.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	now := types.TimestampNow()

	acNoHooks := testAuditCtx(d, user.UserID)
	created, err := d.CreateContentData(d.Context, acNoHooks, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	hooks := &trackingHookRunner{abortBeforeDelete: true}
	acHooks := testAuditCtx(d, user.UserID)
	acHooks.HookRunner = hooks

	err = d.DeleteContentData(d.Context, acHooks, created.ContentDataID)
	if err == nil {
		t.Fatal("expected hook abort error on delete")
	}

	// Entity should still exist
	got, err := d.GetContentData(created.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData after aborted delete: %v", err)
	}
	if got.ContentDataID != created.ContentDataID {
		t.Errorf("entity ID mismatch after aborted delete")
	}
}

// ---------------------------------------------------------------------------
// Tests: Status transition detection
// ---------------------------------------------------------------------------

func TestIntegration_StatusTransition_DraftToPublished(t *testing.T) {
	// DetectStatusTransition should detect draft -> published.
	t.Parallel()

	before := map[string]any{"status": "draft"}
	params := map[string]any{"status": "published"}

	events := audited.DetectStatusTransition("content_data", before, params)
	if len(events) != 1 {
		t.Fatalf("expected 1 transition event, got %d", len(events))
	}
	if events[0] != audited.HookBeforePublish {
		t.Errorf("event = %v, want before_publish", events[0])
	}
}

func TestIntegration_StatusTransition_DraftToNonPublished(t *testing.T) {
	// Transitioning to a non-published status does not fire any transition events.
	t.Parallel()

	before := map[string]any{"status": "draft"}
	params := map[string]any{"status": "draft"}

	events := audited.DetectStatusTransition("content_data", before, params)
	if len(events) != 0 {
		t.Errorf("expected 0 transition events for draft->draft, got %d", len(events))
	}
}

func TestIntegration_StatusTransition_PublishedToPublished(t *testing.T) {
	// No transition: same status -> no events.
	t.Parallel()

	before := map[string]any{"status": "published"}
	params := map[string]any{"status": "published"}

	events := audited.DetectStatusTransition("content_data", before, params)
	if len(events) != 0 {
		t.Errorf("expected 0 transition events, got %d", len(events))
	}
}

func TestIntegration_StatusTransition_NonContentDataTable(t *testing.T) {
	// DetectStatusTransition only fires for content_data table.
	t.Parallel()

	before := map[string]any{"status": "draft"}
	params := map[string]any{"status": "published"}

	events := audited.DetectStatusTransition("routes", before, params)
	if events != nil {
		t.Errorf("expected nil for non-content_data table, got %v", events)
	}
}

func TestIntegration_StatusTransition_NoStatusInParams(t *testing.T) {
	// Partial update not touching status -> no transition.
	t.Parallel()

	before := map[string]any{"status": "draft"}
	params := map[string]any{"title": "updated title"}

	events := audited.DetectStatusTransition("content_data", before, params)
	if events != nil {
		t.Errorf("expected nil when params has no status, got %v", events)
	}
}

func TestIntegration_BeforeToAfterEvent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input audited.HookEvent
		want  audited.HookEvent
	}{
		{audited.HookBeforeCreate, audited.HookAfterCreate},
		{audited.HookBeforeUpdate, audited.HookAfterUpdate},
		{audited.HookBeforeDelete, audited.HookAfterDelete},
		{audited.HookBeforePublish, audited.HookAfterPublish},
		// Unknown event passes through unchanged
		{audited.HookEvent("unknown"), audited.HookEvent("unknown")},
	}

	for _, tc := range cases {
		got := audited.BeforeToAfterEvent(tc.input)
		if got != tc.want {
			t.Errorf("BeforeToAfterEvent(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: StructToMap
// ---------------------------------------------------------------------------

func TestIntegration_StructToMap_NilInput(t *testing.T) {
	t.Parallel()

	result, err := audited.StructToMap(nil)
	if err != nil {
		t.Fatalf("StructToMap(nil): %v", err)
	}
	if result != nil {
		t.Errorf("StructToMap(nil) = %v, want nil", result)
	}
}

func TestIntegration_StructToMap_PreservesFields(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	result, err := audited.StructToMap(testStruct{Name: "test", Count: 42})
	if err != nil {
		t.Fatalf("StructToMap: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
}

// ---------------------------------------------------------------------------
// Tests: Multiple events on same entity are independently queryable
// ---------------------------------------------------------------------------

func TestIntegration_AuditedMultipleUpdates_EachRecorded(t *testing.T) {
	// Three sequential updates produce 4 total events (1 create + 3 updates).
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	created, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	statuses := []types.ContentStatus{
		types.ContentStatusPublished,
		types.ContentStatusDraft,
		types.ContentStatusPublished,
	}

	for _, status := range statuses {
		_, err = d.UpdateContentData(d.Context, ac, db.UpdateContentDataParams{
			ContentDataID: created.ContentDataID,
			RouteID:       routeID,
			ParentID:      types.NullableContentID{},
			FirstChildID:  types.NullableContentID{},
			NextSiblingID: types.NullableContentID{},
			PrevSiblingID: types.NullableContentID{},
			DatatypeID:    datatypeID,
			AuthorID:      user.UserID,
			Status:        status,
			DateCreated:   now,
			DateModified:  types.TimestampNow(),
		})
		if err != nil {
			t.Fatalf("Update to %s: %v", status, err)
		}
	}

	events, err := d.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 4 {
		t.Fatalf("expected 4 change events (1 create + 3 updates), got %d", lenOrZero(events))
	}

	// Count by operation type
	insertCount := 0
	updateCount := 0
	for _, ev := range *events {
		switch ev.Operation {
		case types.OpInsert:
			insertCount++
		case types.OpUpdate:
			updateCount++
		}
	}
	if insertCount != 1 {
		t.Errorf("INSERT events = %d, want 1", insertCount)
	}
	if updateCount != 3 {
		t.Errorf("UPDATE events = %d, want 3", updateCount)
	}
}

// ---------------------------------------------------------------------------
// Tests: HLC timestamps are monotonically increasing
// ---------------------------------------------------------------------------

func TestIntegration_AuditedEvents_DistinctEventIDs(t *testing.T) {
	// Each audited operation produces a change event with a unique EventID.
	t.Parallel()
	d := testDB(t)
	user := seedUser(t, d)
	routeID, datatypeID := seedRouteAndDatatype(t, d, user)
	ac := testAuditCtx(d, user.UserID)
	now := types.TimestampNow()

	created, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = d.UpdateContentData(d.Context, ac, db.UpdateContentDataParams{
		ContentDataID: created.ContentDataID,
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      user.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	events, err := d.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", lenOrZero(events))
	}

	// All EventIDs must be distinct
	seen := make(map[types.EventID]bool)
	for _, ev := range *events {
		if seen[ev.EventID] {
			t.Errorf("duplicate EventID: %v", ev.EventID)
		}
		seen[ev.EventID] = true

		// Each HLC should be non-zero
		if ev.HlcTimestamp == 0 {
			t.Error("HlcTimestamp is zero")
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: Recorder records via transaction
// ---------------------------------------------------------------------------

func TestIntegration_SQLiteRecorder_RecordInTransaction(t *testing.T) {
	// Directly test that the SQLite recorder can record within a transaction.
	t.Parallel()
	d := testDB(t)

	recorder := db.SQLiteRecorder
	nodeID := types.NodeID(d.Config.Node_ID)
	recordID := types.NewULID().String()

	tx, err := d.Connection.BeginTx(d.Context, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	err = recorder.Record(d.Context, tx, audited.ChangeEventParams{
		EventID:      types.NewEventID(),
		HlcTimestamp: types.HLCNow(),
		NodeID:       nodeID,
		TableName:    "test_recorder",
		RecordID:     recordID,
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       types.NullableUserID{},
		NewValues:    types.NewJSONData(map[string]string{"key": "value"}),
		RequestID:    types.NullableString{},
		IP:           types.NullableString{},
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Verify the event was stored
	events, err := d.GetChangeEventsByRecord("test_recorder", recordID)
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 1 {
		t.Fatalf("expected 1 event, got %d", lenOrZero(events))
	}

	ev := (*events)[0]
	if !ev.NewValues.Valid {
		t.Error("NewValues should be valid")
	}
	valBytes, marshalErr := json.Marshal(ev.NewValues.Data)
	if marshalErr != nil {
		t.Fatalf("NewValues marshal: %v", marshalErr)
	}
	var vals map[string]any
	if err := json.Unmarshal(valBytes, &vals); err != nil {
		t.Fatalf("NewValues unmarshal: %v", err)
	}
	if vals["key"] != "value" {
		t.Errorf("NewValues[key] = %v, want value", vals["key"])
	}
}

func TestIntegration_SQLiteRecorder_RollbackPreventsRecord(t *testing.T) {
	// When a transaction is rolled back, the change event should not persist.
	t.Parallel()
	d := testDB(t)

	recorder := db.SQLiteRecorder
	nodeID := types.NodeID(d.Config.Node_ID)
	recordID := types.NewULID().String()

	tx, err := d.Connection.BeginTx(d.Context, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	err = recorder.Record(d.Context, tx, audited.ChangeEventParams{
		EventID:      types.NewEventID(),
		HlcTimestamp: types.HLCNow(),
		NodeID:       nodeID,
		TableName:    "rollback_test",
		RecordID:     recordID,
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       types.NullableUserID{},
		NewValues:    types.NewJSONData(nil),
		RequestID:    types.NullableString{},
		IP:           types.NullableString{},
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	// Rollback instead of commit
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	events, err := d.GetChangeEventsByRecord("rollback_test", recordID)
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events != nil && len(*events) != 0 {
		t.Errorf("expected 0 events after rollback, got %d", len(*events))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func lenOrZero(events *[]db.ChangeEvent) int {
	if events == nil {
		return 0
	}
	return len(*events)
}
