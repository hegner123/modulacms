package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test helpers ---

// newTestDatabase returns a Database with no real connection (nil).
// This is safe because mapper methods do not use the connection.
func newTestDatabase() Database {
	return Database{}
}

func newTestTimestamp() types.Timestamp {
	return types.NewTimestamp(time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC))
}

func newTestNullableUserID() types.NullableUserID {
	return types.NullableUserID{
		ID:    types.NewUserID(),
		Valid: true,
	}
}

// --- MapStringRoute tests ---

func TestMapStringRoute(t *testing.T) {
	t.Parallel()
	ts := newTestTimestamp()
	authorID := newTestNullableUserID()
	routeID := types.NewRouteID()

	route := Routes{
		RouteID:      routeID,
		Slug:         types.Slug("test-slug"),
		Title:        "Test Title",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := MapStringRoute(route)

	if got.RouteID != routeID.String() {
		t.Errorf("RouteID = %q, want %q", got.RouteID, routeID.String())
	}
	if got.Slug != "test-slug" {
		t.Errorf("Slug = %q, want %q", got.Slug, "test-slug")
	}
	if got.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Title")
	}
	if got.Status != "1" {
		t.Errorf("Status = %q, want %q", got.Status, "1")
	}
	if got.AuthorID != authorID.String() {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, authorID.String())
	}
	if got.DateCreated != ts.String() {
		t.Errorf("DateCreated = %q, want %q", got.DateCreated, ts.String())
	}
	if got.DateModified != ts.String() {
		t.Errorf("DateModified = %q, want %q", got.DateModified, ts.String())
	}
	if got.History != "" {
		t.Errorf("History = %q, want empty string", got.History)
	}
}

func TestMapStringRoute_ZeroValues(t *testing.T) {
	t.Parallel()
	route := Routes{}
	got := MapStringRoute(route)
	if got.Status != "0" {
		t.Errorf("Status = %q, want %q", got.Status, "0")
	}
}

// --- MapStringUser tests ---

func TestMapStringUser(t *testing.T) {
	t.Parallel()
	ts := newTestTimestamp()
	userID := types.NewUserID()
	email := types.Email("test@example.com")

	user := Users{
		UserID:       userID,
		Username:     "testuser",
		Name:         "Test User",
		Email:        email,
		Hash:         "hashed123",
		Role:         "admin",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := MapStringUser(user)

	if got.UserID != userID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, userID.String())
	}
	if got.Username != "testuser" {
		t.Errorf("Username = %q, want %q", got.Username, "testuser")
	}
	if got.Name != "Test User" {
		t.Errorf("Name = %q, want %q", got.Name, "Test User")
	}
	if got.Email != email.String() {
		t.Errorf("Email = %q, want %q", got.Email, email.String())
	}
	if got.Hash != "hashed123" {
		t.Errorf("Hash = %q, want %q", got.Hash, "hashed123")
	}
	if got.Role != "admin" {
		t.Errorf("Role = %q, want %q", got.Role, "admin")
	}
}

// --- MapStringPermission tests ---

func TestMapStringPermission(t *testing.T) {
	t.Parallel()
	permID := types.NewPermissionID()
	perm := Permissions{
		PermissionID: permID,
		Label:        "read_write",
	}

	got := MapStringPermission(perm)

	if got.PermissionID != permID.String() {
		t.Errorf("PermissionID = %q, want %q", got.PermissionID, permID.String())
	}
	if got.Label != "read_write" {
		t.Errorf("Label = %q, want %q", got.Label, "read_write")
	}
}

// --- MapStringSession tests ---

func TestMapStringSession(t *testing.T) {
	t.Parallel()
	ts := newTestTimestamp()
	sessionID := types.NewSessionID()
	userID := newTestNullableUserID()

	session := Sessions{
		SessionID:   sessionID,
		UserID:      userID,
		DateCreated:   ts,
		ExpiresAt:   ts,
		LastAccess:  sql.NullString{String: "2024-06-15T13:00:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "192.168.1.1", Valid: true},
		UserAgent:   sql.NullString{String: "Mozilla/5.0", Valid: true},
		SessionData: sql.NullString{String: "{}", Valid: true},
	}

	got := MapStringSession(session)

	if got.SessionID != sessionID.String() {
		t.Errorf("SessionID = %q, want %q", got.SessionID, sessionID.String())
	}
	if got.UserID != userID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, userID.String())
	}
	if got.LastAccess != "2024-06-15T13:00:00Z" {
		t.Errorf("LastAccess = %q, want %q", got.LastAccess, "2024-06-15T13:00:00Z")
	}
	if got.IpAddress != "192.168.1.1" {
		t.Errorf("IpAddress = %q, want %q", got.IpAddress, "192.168.1.1")
	}
	if got.UserAgent != "Mozilla/5.0" {
		t.Errorf("UserAgent = %q, want %q", got.UserAgent, "Mozilla/5.0")
	}
	if got.SessionData != "{}" {
		t.Errorf("SessionData = %q, want %q", got.SessionData, "{}")
	}
}

func TestMapStringSession_NullFields(t *testing.T) {
	t.Parallel()
	// When NullString fields are not valid, they should be empty strings
	session := Sessions{
		LastAccess:  sql.NullString{Valid: false},
		IpAddress:   sql.NullString{Valid: false},
		UserAgent:   sql.NullString{Valid: false},
		SessionData: sql.NullString{Valid: false},
	}

	got := MapStringSession(session)

	if got.LastAccess != "" {
		t.Errorf("LastAccess = %q, want empty", got.LastAccess)
	}
	if got.IpAddress != "" {
		t.Errorf("IpAddress = %q, want empty", got.IpAddress)
	}
	if got.UserAgent != "" {
		t.Errorf("UserAgent = %q, want empty", got.UserAgent)
	}
	if got.SessionData != "" {
		t.Errorf("SessionData = %q, want empty", got.SessionData)
	}
}

// --- MapStringToken tests ---

func TestMapStringToken(t *testing.T) {
	t.Parallel()
	ts := newTestTimestamp()
	userID := newTestNullableUserID()

	token := Tokens{
		ID:        "tok-123",
		UserID:    userID,
		TokenType: "access",
		Token:     "abc-def-ghi",
		IssuedAt:  "2024-06-15T12:00:00Z",
		ExpiresAt: ts,
		Revoked:   true,
	}

	got := MapStringToken(token)

	if got.ID != "tok-123" {
		t.Errorf("ID = %q, want %q", got.ID, "tok-123")
	}
	if got.UserID != userID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, userID.String())
	}
	if got.TokenType != "access" {
		t.Errorf("TokenType = %q, want %q", got.TokenType, "access")
	}
	if got.Token != "abc-def-ghi" {
		t.Errorf("Token = %q, want %q", got.Token, "abc-def-ghi")
	}
	if got.IssuedAt != "2024-06-15T12:00:00Z" {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, "2024-06-15T12:00:00Z")
	}
	if got.Revoked != "true" {
		t.Errorf("Revoked = %q, want %q", got.Revoked, "true")
	}
}

func TestMapStringToken_RevokedFalse(t *testing.T) {
	t.Parallel()
	token := Tokens{Revoked: false}
	got := MapStringToken(token)
	if got.Revoked != "false" {
		t.Errorf("Revoked = %q, want %q", got.Revoked, "false")
	}
}

// --- MapStringTable tests ---

func TestMapStringTable(t *testing.T) {
	t.Parallel()
	userID := newTestNullableUserID()
	tbl := Tables{
		ID:       "tbl-001",
		Label:    "users",
		AuthorID: userID,
	}

	got := MapStringTable(tbl)

	if got.ID != "tbl-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-001")
	}
	if got.Label != "users" {
		t.Errorf("Label = %q, want %q", got.Label, "users")
	}
	if got.AuthorID != userID.String() {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, userID.String())
	}
}

// --- SQLite MapRoute tests ---

func TestDatabase_MapRoute(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	routeID := types.NewRouteID()
	authorID := newTestNullableUserID()

	input := mdb.Routes{
		RouteID:      routeID,
		Slug:         types.Slug("my-route"),
		Title:        "My Route",
		Status:       2,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapRoute(input)

	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.Slug != types.Slug("my-route") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("my-route"))
	}
	if got.Title != "My Route" {
		t.Errorf("Title = %q, want %q", got.Title, "My Route")
	}
	if got.Status != 2 {
		t.Errorf("Status = %d, want %d", got.Status, 2)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapCreateRouteParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()

	input := CreateRouteParams{
		Slug:         types.Slug("auto-id"),
		Title:        "Auto ID Route",
		Status:       1,
		AuthorID:     newTestNullableUserID(),
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateRouteParams(input)

	if got.RouteID.IsZero() {
		t.Fatal("expected non-zero RouteID to be generated when input is empty")
	}
	if got.Slug != types.Slug("auto-id") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("auto-id"))
	}
	if got.Title != "Auto ID Route" {
		t.Errorf("Title = %q, want %q", got.Title, "Auto ID Route")
	}
}

func TestDatabase_MapCreateRouteParams_AlwaysGeneratesID(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()

	input := CreateRouteParams{
		Slug:         types.Slug("explicit-id"),
		Title:        "Explicit ID Route",
		Status:       0,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateRouteParams(input)

	if got.RouteID.IsZero() {
		t.Fatal("expected non-zero RouteID to be generated")
	}
}

func TestDatabase_MapUpdateRouteParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()

	input := UpdateRouteParams{
		Slug:         types.Slug("updated-slug"),
		Title:        "Updated Title",
		Status:       3,
		AuthorID:     newTestNullableUserID(),
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("original-slug"),
	}

	got := d.MapUpdateRouteParams(input)

	if got.Slug != types.Slug("updated-slug") {
		t.Errorf("Slug = %v, want %v", got.Slug, types.Slug("updated-slug"))
	}
	if got.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated Title")
	}
	if got.Status != 3 {
		t.Errorf("Status = %d, want %d", got.Status, 3)
	}
	if got.Slug_2 != types.Slug("original-slug") {
		t.Errorf("Slug_2 = %v, want %v", got.Slug_2, types.Slug("original-slug"))
	}
}

// --- SQLite MapUser tests ---

func TestDatabase_MapUser(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	userID := types.NewUserID()

	input := mdb.Users{
		UserID:       userID,
		Username:     "jdoe",
		Name:         "John Doe",
		Email:        types.Email("jdoe@example.com"),
		Hash:         "argon2hash",
		Roles:        "editor",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapUser(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "jdoe" {
		t.Errorf("Username = %q, want %q", got.Username, "jdoe")
	}
	// Note: mdb.Users has field Roles, but the wrapper has field Role
	if got.Role != "editor" {
		t.Errorf("Role = %q, want %q", got.Role, "editor")
	}
}

func TestDatabase_MapCreateUserParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()

	input := CreateUserParams{
		Username: "newuser",
		Name:     "New User",
		Email:    types.Email("new@example.com"),
		Hash:     "hash",
		Role:     "viewer",
	}

	got := d.MapCreateUserParams(input)

	if got.UserID.IsZero() {
		t.Fatal("expected non-zero UserID to be generated")
	}
	if got.Username != "newuser" {
		t.Errorf("Username = %q, want %q", got.Username, "newuser")
	}
	// Verify the Role -> Roles field mapping
	if got.Roles != "viewer" {
		t.Errorf("Roles = %q, want %q", got.Roles, "viewer")
	}
}

func TestDatabase_MapUpdateUserParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	userID := types.NewUserID()

	input := UpdateUserParams{
		Username: "updateduser",
		Name:     "Updated User",
		Email:    types.Email("updated@example.com"),
		Hash:     "newhash",
		Role:     "admin",
		UserID:   userID,
	}

	got := d.MapUpdateUserParams(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "updateduser" {
		t.Errorf("Username = %q, want %q", got.Username, "updateduser")
	}
	if got.Roles != "admin" {
		t.Errorf("Roles = %q, want %q", got.Roles, "admin")
	}
}

// --- SQLite MapPermission tests ---

func TestDatabase_MapPermission(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	permID := types.NewPermissionID()

	input := mdb.Permissions{
		PermissionID: permID,
		Label:        "full_access",
	}

	got := d.MapPermission(input)

	if got.PermissionID != permID {
		t.Errorf("PermissionID = %v, want %v", got.PermissionID, permID)
	}
	if got.Label != "full_access" {
		t.Errorf("Label = %q, want %q", got.Label, "full_access")
	}
}

func TestDatabase_MapCreatePermissionParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	input := CreatePermissionParams{
		Label: "read",
	}

	got := d.MapCreatePermissionParams(input)

	if got.PermissionID.IsZero() {
		t.Fatal("expected non-zero PermissionID to be generated")
	}
	if got.Label != "read" {
		t.Errorf("Label = %q, want %q", got.Label, "read")
	}
}

// --- SQLite MapRole tests ---

func TestDatabase_MapRole(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	roleID := types.NewRoleID()

	input := mdb.Roles{
		RoleID: roleID,
		Label:  "admin",
	}

	got := d.MapRole(input)

	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
	if got.Label != "admin" {
		t.Errorf("Label = %q, want %q", got.Label, "admin")
	}
}

// --- SQLite MapSession tests ---

func TestDatabase_MapSession(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	sessionID := types.NewSessionID()
	userID := newTestNullableUserID()

	input := mdb.Sessions{
		SessionID:   sessionID,
		UserID:      userID,
		DateCreated:   ts,
		ExpiresAt:   ts,
		LastAccess:  sql.NullString{String: "2024-06-15T12:00:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "10.0.0.1", Valid: true},
		UserAgent:   sql.NullString{String: "curl/7.79", Valid: true},
		SessionData: sql.NullString{String: `{"foo":"bar"}`, Valid: true},
	}

	got := d.MapSession(input)

	if got.SessionID != sessionID {
		t.Errorf("SessionID = %v, want %v", got.SessionID, sessionID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.LastAccess.String != "2024-06-15T12:00:00Z" || !got.LastAccess.Valid {
		t.Errorf("LastAccess = %v, want valid 2024-06-15T12:00:00Z", got.LastAccess)
	}
	if got.IpAddress.String != "10.0.0.1" || !got.IpAddress.Valid {
		t.Errorf("IpAddress = %v, want valid 10.0.0.1", got.IpAddress)
	}
}

// --- SQLite MapToken tests ---

func TestDatabase_MapToken(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	userID := newTestNullableUserID()

	input := mdb.Tokens{
		ID:        "tok-abc",
		UserID:    userID,
		TokenType: "refresh",
		Tokens:    "refresh-token-value",
		IssuedAt:  "2024-06-15T12:00:00Z",
		ExpiresAt: ts,
		Revoked:   false,
	}

	got := d.MapToken(input)

	if got.ID != "tok-abc" {
		t.Errorf("ID = %q, want %q", got.ID, "tok-abc")
	}
	// Note: mdb.Tokens has Tokens field, wrapper has Token field
	if got.Token != "refresh-token-value" {
		t.Errorf("Token = %q, want %q", got.Token, "refresh-token-value")
	}
	if got.TokenType != "refresh" {
		t.Errorf("TokenType = %q, want %q", got.TokenType, "refresh")
	}
	if got.Revoked != false {
		t.Errorf("Revoked = %v, want false", got.Revoked)
	}
}

// --- SQLite MapTable tests ---

func TestDatabase_MapTable(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	userID := newTestNullableUserID()

	input := mdb.Tables{
		ID:       "tbl-xyz",
		Label:    "content_data",
		AuthorID: userID,
	}

	got := d.MapTable(input)

	if got.ID != "tbl-xyz" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-xyz")
	}
	if got.Label != "content_data" {
		t.Errorf("Label = %q, want %q", got.Label, "content_data")
	}
	if got.AuthorID != userID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, userID)
	}
}

// --- Audited Command Accessor tests ---

func TestNewRouteCmd_Accessors(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID: types.NewUserID(),
	}
	params := CreateRouteParams{
		Slug:         types.Slug("test"),
		Title:        "Test",
		Status:       1,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := d.NewRouteCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.Params() == nil {
		t.Error("Params() returned nil")
	}
	if cmd.Connection() != nil {
		// We constructed with nil Connection, so it should be nil
		t.Error("Connection() should be nil for test database")
	}
	if cmd.Recorder() == nil {
		t.Error("Recorder() returned nil")
	}
}

func TestUpdateRouteCmd_Accessors(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID: types.NewUserID(),
	}
	params := UpdateRouteParams{
		Slug:         types.Slug("updated"),
		Title:        "Updated",
		Status:       2,
		DateCreated:  ts,
		DateModified: ts,
		Slug_2:       types.Slug("original"),
	}

	cmd := d.UpdateRouteCmd(ctx, ac, params)

	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.GetID() != "original" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "original")
	}
}

func TestDeleteRouteCmd_Accessors(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ctx := context.Background()
	ac := audited.AuditContext{}
	routeID := types.NewRouteID()

	cmd := d.DeleteRouteCmd(ctx, ac, routeID)

	if cmd.TableName() != "routes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "routes")
	}
	if cmd.GetID() != string(routeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(routeID))
	}
}

func TestNewUserCmd_Accessors(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := CreateUserParams{
		Username: "cmd-test",
		Name:     "Cmd Test",
		Email:    types.Email("cmd@test.com"),
	}

	cmd := d.NewUserCmd(ctx, ac, params)

	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.Recorder() == nil {
		t.Error("Recorder() returned nil")
	}
}

func TestDeleteUserCmd_Accessors(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ctx := context.Background()
	ac := audited.AuditContext{}
	userID := types.NewUserID()

	cmd := d.DeleteUserCmd(ctx, ac, userID)

	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.GetID() != string(userID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(userID))
	}
}

func TestNewPermissionCmd_Accessors(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ctx := context.Background()
	ac := audited.AuditContext{}
	params := CreatePermissionParams{
		Label: "read",
	}

	cmd := d.NewPermissionCmd(ctx, ac, params)

	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
}

func TestNewSessionCmd_Accessors(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ctx := context.Background()
	ac := audited.AuditContext{}
	params := CreateSessionParams{
		UserID: newTestNullableUserID(),
	}

	cmd := d.NewSessionCmd(ctx, ac, params)

	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
}

// --- Audited Command GetID tests for mdb types ---

func TestNewRouteCmd_GetID(t *testing.T) {
	t.Parallel()
	routeID := types.NewRouteID()
	cmd := NewRouteCmd{}

	row := mdb.Routes{RouteID: routeID}
	got := cmd.GetID(row)
	if got != string(routeID) {
		t.Errorf("GetID() = %q, want %q", got, string(routeID))
	}
}

func TestNewUserCmd_GetID(t *testing.T) {
	t.Parallel()
	userID := types.NewUserID()
	cmd := NewUserCmd{}

	row := mdb.Users{UserID: userID}
	got := cmd.GetID(row)
	if got != string(userID) {
		t.Errorf("GetID() = %q, want %q", got, string(userID))
	}
}

func TestUpdateUserCmd_GetID(t *testing.T) {
	t.Parallel()
	userID := types.NewUserID()
	cmd := UpdateUserCmd{
		params: UpdateUserParams{UserID: userID},
	}

	got := cmd.GetID()
	if got != string(userID) {
		t.Errorf("GetID() = %q, want %q", got, string(userID))
	}
}

func TestDeletePermissionCmd_GetID(t *testing.T) {
	t.Parallel()
	permID := types.NewPermissionID()
	cmd := DeletePermissionCmd{id: permID}

	got := cmd.GetID()
	if got != permID.String() {
		t.Errorf("GetID() = %q, want %q", got, permID.String())
	}
}

// --- MapCreateTokenParams tests ---

func TestDatabase_MapCreateTokenParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	userID := newTestNullableUserID()

	input := CreateTokenParams{
		UserID:    userID,
		TokenType: "access",
		Token:     "my-token",
		IssuedAt:  "2024-06-15T12:00:00Z",
		ExpiresAt: ts,
		Revoked:   false,
	}

	got := d.MapCreateTokenParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.TokenType != "access" {
		t.Errorf("TokenType = %q, want %q", got.TokenType, "access")
	}
	// Note: wrapper Token maps to mdb Tokens field
	if got.Tokens != "my-token" {
		t.Errorf("Tokens = %q, want %q", got.Tokens, "my-token")
	}
}

func TestDatabase_MapUpdateTokenParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()

	input := UpdateTokenParams{
		Token:     "updated-token",
		IssuedAt:  "2024-06-15T12:00:00Z",
		ExpiresAt: ts,
		Revoked:   true,
		ID:        "tok-001",
	}

	got := d.MapUpdateTokenParams(input)

	if got.Tokens != "updated-token" {
		t.Errorf("Tokens = %q, want %q", got.Tokens, "updated-token")
	}
	if got.ID != "tok-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tok-001")
	}
	if got.Revoked != true {
		t.Errorf("Revoked = %v, want true", got.Revoked)
	}
}

// --- MapCreateTableParams tests ---

func TestDatabase_MapCreateTableParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()

	input := CreateTableParams{
		Label: "my_table",
	}

	got := d.MapCreateTableParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.Label != "my_table" {
		t.Errorf("Label = %q, want %q", got.Label, "my_table")
	}
}

func TestDatabase_MapUpdateTableParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()

	input := UpdateTableParams{
		Label: "renamed_table",
		ID:    "tbl-001",
	}

	got := d.MapUpdateTableParams(input)

	if got.Label != "renamed_table" {
		t.Errorf("Label = %q, want %q", got.Label, "renamed_table")
	}
	if got.ID != "tbl-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-001")
	}
}

// --- MapCreateSessionParams tests ---

func TestDatabase_MapCreateSessionParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	userID := newTestNullableUserID()

	input := CreateSessionParams{
		UserID:      userID,
		DateCreated:   ts,
		ExpiresAt:   ts,
		LastAccess:  sql.NullString{String: "2024-06-15T12:00:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "127.0.0.1", Valid: true},
		UserAgent:   sql.NullString{String: "test-agent", Valid: true},
		SessionData: sql.NullString{String: "{}", Valid: true},
	}

	got := d.MapCreateSessionParams(input)

	if got.SessionID.IsZero() {
		t.Fatal("expected non-zero SessionID to be generated")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.IpAddress.String != "127.0.0.1" {
		t.Errorf("IpAddress = %v, want 127.0.0.1", got.IpAddress)
	}
}

// --- MapCreateRoleParams tests ---

func TestDatabase_MapCreateRoleParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()

	input := CreateRoleParams{
		Label: "editor",
	}

	got := d.MapCreateRoleParams(input)

	if got.RoleID.IsZero() {
		t.Fatal("expected non-zero RoleID to be generated")
	}
	if got.Label != "editor" {
		t.Errorf("Label = %q, want %q", got.Label, "editor")
	}
}

func TestDatabase_MapUpdateRoleParams(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	roleID := types.NewRoleID()

	input := UpdateRoleParams{
		Label:  "super-admin",
		RoleID: roleID,
	}

	got := d.MapUpdateRoleParams(input)

	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
	if got.Label != "super-admin" {
		t.Errorf("Label = %q, want %q", got.Label, "super-admin")
	}
}

// --- MapMedia tests ---

func TestDatabase_MapMedia(t *testing.T) {
	t.Parallel()
	d := newTestDatabase()
	ts := newTestTimestamp()
	mediaID := types.NewMediaID()
	authorID := newTestNullableUserID()

	input := mdb.Media{
		MediaID:      mediaID,
		Name:         sql.NullString{String: "photo.jpg", Valid: true},
		DisplayName:  sql.NullString{String: "My Photo", Valid: true},
		Alt:          sql.NullString{String: "A photo", Valid: true},
		Caption:      sql.NullString{String: "Caption text", Valid: true},
		Description:  sql.NullString{String: "Description", Valid: true},
		Class:        sql.NullString{String: "image", Valid: true},
		Mimetype:     sql.NullString{String: "image/jpeg", Valid: true},
		Dimensions:   sql.NullString{String: "800x600", Valid: true},
		URL:          types.URL("https://example.com/photo.jpg"),
		Srcset:       sql.NullString{Valid: false},
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapMedia(input)

	if got.MediaID != mediaID {
		t.Errorf("MediaID = %v, want %v", got.MediaID, mediaID)
	}
	if got.Name.String != "photo.jpg" || !got.Name.Valid {
		t.Errorf("Name = %v, want valid 'photo.jpg'", got.Name)
	}
	if string(got.URL) != "https://example.com/photo.jpg" {
		t.Errorf("URL = %v, want %v", got.URL, "https://example.com/photo.jpg")
	}
	if got.Srcset.Valid {
		t.Error("Srcset should be invalid/null")
	}
}

// --- Command Table Names are consistent across database types ---

func TestAuditedCommand_TableNames(t *testing.T) {
	t.Parallel()

	// Verify that all command types for the same entity return the same table name
	tests := []struct {
		name     string
		commands []fmt.Stringer
		want     string
	}{
		{
			name: "route commands",
			want: "routes",
		},
		{
			name: "user commands",
			want: "users",
		},
		{
			name: "permission commands",
			want: "permissions",
		},
		{
			name: "session commands",
			want: "sessions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newTestDatabase()
			ctx := context.Background()
			ac := audited.AuditContext{}

			switch tt.want {
			case "routes":
				cmd1 := d.NewRouteCmd(ctx, ac, CreateRouteParams{})
				cmd2 := d.UpdateRouteCmd(ctx, ac, UpdateRouteParams{})
				cmd3 := d.DeleteRouteCmd(ctx, ac, types.NewRouteID())
				if cmd1.TableName() != tt.want || cmd2.TableName() != tt.want || cmd3.TableName() != tt.want {
					t.Errorf("inconsistent table names: create=%q update=%q delete=%q",
						cmd1.TableName(), cmd2.TableName(), cmd3.TableName())
				}
			case "users":
				cmd1 := d.NewUserCmd(ctx, ac, CreateUserParams{})
				cmd2 := d.UpdateUserCmd(ctx, ac, UpdateUserParams{})
				cmd3 := d.DeleteUserCmd(ctx, ac, types.NewUserID())
				if cmd1.TableName() != tt.want || cmd2.TableName() != tt.want || cmd3.TableName() != tt.want {
					t.Errorf("inconsistent table names: create=%q update=%q delete=%q",
						cmd1.TableName(), cmd2.TableName(), cmd3.TableName())
				}
			case "permissions":
				cmd1 := d.NewPermissionCmd(ctx, ac, CreatePermissionParams{})
				cmd2 := d.UpdatePermissionCmd(ctx, ac, UpdatePermissionParams{})
				cmd3 := d.DeletePermissionCmd(ctx, ac, types.NewPermissionID())
				if cmd1.TableName() != tt.want || cmd2.TableName() != tt.want || cmd3.TableName() != tt.want {
					t.Errorf("inconsistent table names: create=%q update=%q delete=%q",
						cmd1.TableName(), cmd2.TableName(), cmd3.TableName())
				}
			case "sessions":
				cmd1 := d.NewSessionCmd(ctx, ac, CreateSessionParams{})
				cmd2 := d.UpdateSessionCmd(ctx, ac, UpdateSessionParams{})
				cmd3 := d.DeleteSessionCmd(ctx, ac, types.NewSessionID())
				if cmd1.TableName() != tt.want || cmd2.TableName() != tt.want || cmd3.TableName() != tt.want {
					t.Errorf("inconsistent table names: create=%q update=%q delete=%q",
						cmd1.TableName(), cmd2.TableName(), cmd3.TableName())
				}
			}
		})
	}
}
