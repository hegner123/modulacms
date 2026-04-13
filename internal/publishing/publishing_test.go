package publishing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"

	_ "github.com/mattn/go-sqlite3"
)

// ===== TEST HELPERS =====

// testDB creates an isolated SQLite database with all tables for integration tests.
func testDB(t *testing.T) db.Database {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "publishing_test.db")
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

// testAuditCtx returns an AuditContext with no user.
func testAuditCtx(d db.Database) audited.AuditContext {
	return audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "test", "127.0.0.1")
}

// testAuditCtxWithUser returns an AuditContext with a real UserID.
func testAuditCtxWithUser(d db.Database, userID types.UserID) audited.AuditContext {
	return audited.Ctx(types.NodeID(d.Config.Node_ID), userID, "test", "127.0.0.1")
}

// seedData holds minimal FK-satisfying records for publishing tests.
type seedData struct {
	User     *db.Users
	Route    *db.Routes
	Datatype *db.Datatypes
	Field    *db.Fields
}

// testSeededDB creates a DB with route, datatype, field, and user for publishing tests.
func testSeededDB(t *testing.T) (db.Database, seedData) {
	t.Helper()
	d := testDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	now := types.TimestampNow()

	perm, err := d.CreatePermission(ctx, ac, db.CreatePermissionParams{Label: "test:read"})
	if err != nil {
		t.Fatalf("seed CreatePermission: %v", err)
	}
	_ = perm

	role, err := d.CreateRole(ctx, ac, db.CreateRoleParams{Label: "test-role"})
	if err != nil {
		t.Fatalf("seed CreateRole: %v", err)
	}

	user, err := d.CreateUser(ctx, ac, db.CreateUserParams{
		Username:     "publisher",
		Name:         "Publisher",
		Email:        types.Email("pub@example.com"),
		Hash:         "fakehash",
		Role:         role.RoleID.String(),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}

	acUser := testAuditCtxWithUser(d, user.UserID)

	route, err := d.CreateRoute(ctx, acUser, db.CreateRouteParams{
		Slug:         types.Slug("test-page"),
		Title:        "Test Page",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateRoute: %v", err)
	}

	datatype, err := d.CreateDatatype(ctx, acUser, db.CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "article",
		Type:         "page",
		AuthorID:     user.UserID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateDatatype: %v", err)
	}

	field, err := d.CreateField(ctx, acUser, db.CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{ID: datatype.DatatypeID, Valid: true},
		Label:        "title",
		Data:         "",
		ValidationID: types.NullableValidationID{},
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateField: %v", err)
	}

	return d, seedData{
		User:     user,
		Route:    route,
		Datatype: datatype,
		Field:    field,
	}
}

// createContentWithField creates a content data row with one content field attached.
// Returns the content data and content field.
func createContentWithField(t *testing.T, d db.Database, seed seedData) (*db.ContentData, *db.ContentFields) {
	t.Helper()
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	cd, err := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		RouteID:       types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true},
		AuthorID:      seed.User.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	cf, err := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		RootID:        types.NullableContentID{ID: cd.ContentDataID, Valid: true},
		ContentDataID: types.NullableContentID{ID: cd.ContentDataID, Valid: true},
		FieldID:       types.NullableFieldID{ID: seed.Field.FieldID, Valid: true},
		FieldValue:    "Hello World",
		Locale:        "",
		AuthorID:      seed.User.UserID,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentField: %v", err)
	}

	return cd, cf
}

// mockDispatcher records dispatched webhook events for assertions.
type mockDispatcher struct {
	mu     sync.Mutex
	events []dispatchedEvent
}

type dispatchedEvent struct {
	event string
	data  map[string]any
}

func (m *mockDispatcher) Dispatch(_ context.Context, event string, data map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, dispatchedEvent{event: event, data: data})
}

func (m *mockDispatcher) getEvents() []dispatchedEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]dispatchedEvent, len(m.events))
	copy(cp, m.events)
	return cp
}

// mockIndexer records indexing calls.
type mockIndexer struct {
	mu          sync.Mutex
	published   []db.ContentVersion
	unpublished []string
}

func (m *mockIndexer) OnPublish(snapshot *Snapshot, version db.ContentVersion) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, version)
}

func (m *mockIndexer) OnUnpublish(contentDataID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.unpublished = append(m.unpublished, contentDataID)
}

// ===== UNIT TESTS: ParseSnapshotTimestamp =====

func TestParseSnapshotTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantValid bool
		wantErr   bool
		wantTime  time.Time
	}{
		{
			name:      "empty string returns zero timestamp",
			input:     "",
			wantValid: false,
		},
		{
			name:      "null string returns zero timestamp",
			input:     "null",
			wantValid: false,
		},
		{
			name:      "valid RFC3339",
			input:     "2025-06-15T10:30:00Z",
			wantValid: true,
			wantTime:  time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:      "valid RFC3339 with offset",
			input:     "2025-06-15T10:30:00+05:00",
			wantValid: true,
			wantTime:  time.Date(2025, 6, 15, 5, 30, 0, 0, time.UTC),
		},
		{
			name:      "valid RFC3339Nano",
			input:     "2025-06-15T10:30:00.123456789Z",
			wantValid: true,
			wantTime:  time.Date(2025, 6, 15, 10, 30, 0, 123456789, time.UTC),
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "partial date",
			input:   "2025-06-15",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseSnapshotTimestamp(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if tt.wantValid && !got.Time.Equal(tt.wantTime) {
				t.Errorf("Time = %v, want %v", got.Time, tt.wantTime)
			}
		})
	}
}

// ===== UNIT TESTS: Nullable ID parsers =====

func TestSnapshotNullableContentID(t *testing.T) {
	t.Parallel()

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		got := SnapshotNullableContentID("")
		if got.Valid {
			t.Error("empty string should produce invalid nullable")
		}
	})

	t.Run("null string", func(t *testing.T) {
		t.Parallel()
		got := SnapshotNullableContentID("null")
		if got.Valid {
			t.Error("null string should produce invalid nullable")
		}
	})

	t.Run("valid ULID", func(t *testing.T) {
		t.Parallel()
		id := types.NewContentID()
		got := SnapshotNullableContentID(string(id))
		if !got.Valid {
			t.Error("valid ULID should produce valid nullable")
		}
		if got.ID != id {
			t.Errorf("ID = %v, want %v", got.ID, id)
		}
	})
}

func TestParseNullableRouteID(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		if ParseNullableRouteID("").Valid {
			t.Error("expected invalid")
		}
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		if ParseNullableRouteID("null").Valid {
			t.Error("expected invalid")
		}
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		id := types.NewRouteID()
		got := ParseNullableRouteID(string(id))
		if !got.Valid || got.ID != id {
			t.Errorf("got %v, want valid with ID %v", got, id)
		}
	})
}

func TestParseNullableDatatypeID(t *testing.T) {
	t.Parallel()

	if ParseNullableDatatypeID("").Valid {
		t.Error("empty should be invalid")
	}
	if ParseNullableDatatypeID("null").Valid {
		t.Error("null should be invalid")
	}

	id := types.NewDatatypeID()
	got := ParseNullableDatatypeID(string(id))
	if !got.Valid || got.ID != id {
		t.Errorf("got %v, want valid with ID %v", got, id)
	}
}

func TestParseNullableFieldID(t *testing.T) {
	t.Parallel()

	if ParseNullableFieldID("").Valid {
		t.Error("empty should be invalid")
	}
	if ParseNullableFieldID("null").Valid {
		t.Error("null should be invalid")
	}

	id := types.NewFieldID()
	got := ParseNullableFieldID(string(id))
	if !got.Valid || got.ID != id {
		t.Errorf("got %v, want valid with ID %v", got, id)
	}
}

func TestParseNullableUserID(t *testing.T) {
	t.Parallel()

	if ParseNullableUserID("").Valid {
		t.Error("empty should be invalid")
	}
	if ParseNullableUserID("null").Valid {
		t.Error("null should be invalid")
	}

	id := types.NewUserID()
	got := ParseNullableUserID(string(id))
	if !got.Valid || got.ID != id {
		t.Errorf("got %v, want valid with ID %v", got, id)
	}
}

func TestParseNullableValidationID(t *testing.T) {
	t.Parallel()

	if ParseNullableValidationID("").Valid {
		t.Error("empty should be invalid")
	}
	if ParseNullableValidationID("null").Valid {
		t.Error("null should be invalid")
	}

	id := types.NewValidationID()
	got := ParseNullableValidationID(string(id))
	if !got.Valid || got.ID != id {
		t.Errorf("got %v, want valid with ID %v", got, id)
	}
}

// ===== UNIT TESTS: MapSnapshotContentFieldJSON =====

func TestMapSnapshotContentFieldJSON(t *testing.T) {
	t.Parallel()
	now := types.TimestampNow()
	cf := db.ContentFields{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		ContentDataID:  types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:        types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:     "test value",
		Locale:         "en",
		AuthorID:       types.NewUserID(),
		DateCreated:    now,
		DateModified:   now,
	}

	got := MapSnapshotContentFieldJSON(cf)

	if got.ContentFieldID != cf.ContentFieldID.String() {
		t.Errorf("ContentFieldID = %q, want %q", got.ContentFieldID, cf.ContentFieldID.String())
	}
	if got.RouteID != cf.RouteID.ID.String() {
		t.Errorf("RouteID = %q, want %q", got.RouteID, cf.RouteID.ID.String())
	}
	if got.ContentDataID != cf.ContentDataID.ID.String() {
		t.Errorf("ContentDataID = %q, want %q", got.ContentDataID, cf.ContentDataID.ID.String())
	}
	if got.FieldID != cf.FieldID.ID.String() {
		t.Errorf("FieldID = %q, want %q", got.FieldID, cf.FieldID.ID.String())
	}
	if got.FieldValue != "test value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "test value")
	}
	if got.Locale != "en" {
		t.Errorf("Locale = %q, want %q", got.Locale, "en")
	}
	if got.AuthorID != cf.AuthorID.String() {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, cf.AuthorID.String())
	}
}

func TestMapAdminSnapshotContentFieldJSON(t *testing.T) {
	t.Parallel()
	now := types.TimestampNow()
	cf := db.AdminContentFields{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		AdminFieldID:        types.NullableAdminFieldID{ID: types.NewAdminFieldID(), Valid: true},
		AdminFieldValue:     "admin value",
		Locale:              "de",
		AuthorID:            types.NewUserID(),
		DateCreated:         now,
		DateModified:        now,
	}

	got := MapAdminSnapshotContentFieldJSON(cf)

	if got.AdminContentFieldID != cf.AdminContentFieldID.String() {
		t.Errorf("AdminContentFieldID = %q, want %q", got.AdminContentFieldID, cf.AdminContentFieldID.String())
	}
	if got.AdminFieldValue != "admin value" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "admin value")
	}
	if got.Locale != "de" {
		t.Errorf("Locale = %q, want %q", got.Locale, "de")
	}
}

// ===== UNIT TESTS: SnapshotContentDataToSlice =====

func TestSnapshotContentDataToSlice(t *testing.T) {
	t.Parallel()

	t.Run("roundtrip preserves IDs and status", func(t *testing.T) {
		t.Parallel()
		cdID := types.NewContentID()
		parentID := types.NewContentID()
		routeID := types.NewRouteID()
		dtID := types.NewDatatypeID()
		authorID := types.NewUserID()
		now := time.Now().UTC().Truncate(time.Second)

		items := []db.ContentDataJSON{
			{
				ContentDataID: string(cdID),
				ParentID:      string(parentID),
				RouteID:       string(routeID),
				DatatypeID:    string(dtID),
				AuthorID:      string(authorID),
				Status:        string(types.ContentStatusDraft),
				DateCreated:   now.Format(time.RFC3339),
				DateModified:  now.Format(time.RFC3339),
				PublishedAt:   "",
				PublishAt:     "",
			},
		}

		result, err := SnapshotContentDataToSlice(items)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}

		got := result[0]
		if got.ContentDataID != cdID {
			t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, cdID)
		}
		if !got.ParentID.Valid || got.ParentID.ID != types.ContentID(parentID) {
			t.Errorf("ParentID = %v, want valid %v", got.ParentID, parentID)
		}
		if got.Status != types.ContentStatusDraft {
			t.Errorf("Status = %v, want draft", got.Status)
		}
	})

	t.Run("empty nullables produce invalid fields", func(t *testing.T) {
		t.Parallel()
		now := time.Now().UTC().Format(time.RFC3339)
		items := []db.ContentDataJSON{
			{
				ContentDataID: string(types.NewContentID()),
				ParentID:      "",
				FirstChildID:  "null",
				RouteID:       "",
				DatatypeID:    "null",
				AuthorID:      string(types.NewUserID()),
				Status:        "draft",
				DateCreated:   now,
				DateModified:  now,
				PublishedAt:   "null",
				PublishAt:     "",
			},
		}

		result, err := SnapshotContentDataToSlice(items)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := result[0]
		if got.ParentID.Valid {
			t.Error("ParentID should be invalid for empty string")
		}
		if got.FirstChildID.Valid {
			t.Error("FirstChildID should be invalid for null string")
		}
		if got.DatatypeID.Valid {
			t.Error("DatatypeID should be invalid for null string")
		}
	})

	t.Run("invalid timestamp returns error", func(t *testing.T) {
		t.Parallel()
		items := []db.ContentDataJSON{
			{
				ContentDataID: string(types.NewContentID()),
				DateCreated:   "bad-date",
				DateModified:  "2025-01-01T00:00:00Z",
			},
		}

		_, err := SnapshotContentDataToSlice(items)
		if err == nil {
			t.Fatal("expected error for invalid timestamp")
		}
	})
}

// ===== UNIT TESTS: SnapshotDatatypesToSlice =====

func TestSnapshotDatatypesToSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid roundtrip", func(t *testing.T) {
		t.Parallel()
		dtID := types.NewDatatypeID()
		now := time.Now().UTC().Truncate(time.Second)

		items := []db.DatatypeJSON{
			{
				DatatypeID:   string(dtID),
				ParentID:     "",
				Name:         "article",
				Label:        "Article",
				Type:         "page",
				AuthorID:     string(types.NewUserID()),
				DateCreated:  now.Format(time.RFC3339),
				DateModified: now.Format(time.RFC3339),
			},
		}

		result, err := SnapshotDatatypesToSlice(items)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		if result[0].DatatypeID != dtID {
			t.Errorf("DatatypeID = %v, want %v", result[0].DatatypeID, dtID)
		}
		if result[0].Label != "Article" {
			t.Errorf("Label = %q, want %q", result[0].Label, "Article")
		}
	})
}

// ===== UNIT TESTS: SnapshotContentFieldsToSlice =====

func TestSnapshotContentFieldsToSlice(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	cfID := types.NewContentFieldID()
	routeID := types.NewRouteID()
	cdID := types.NewContentID()
	fieldID := types.NewFieldID()

	items := []SnapshotContentFieldJSON{
		{
			ContentFieldID: string(cfID),
			RouteID:        string(routeID),
			ContentDataID:  string(cdID),
			FieldID:        string(fieldID),
			FieldValue:     "snapshot value",
			Locale:         "en",
			AuthorID:       string(types.NewUserID()),
			DateCreated:    now.Format(time.RFC3339),
			DateModified:   now.Format(time.RFC3339),
		},
	}

	result, err := SnapshotContentFieldsToSlice(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}

	got := result[0]
	if got.ContentFieldID != cfID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, cfID)
	}
	if got.FieldValue != "snapshot value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "snapshot value")
	}
	if got.Locale != "en" {
		t.Errorf("Locale = %q, want %q", got.Locale, "en")
	}
}

// ===== UNIT TESTS: SnapshotFieldsToSlice =====

func TestSnapshotFieldsToSlice(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	fieldID := types.NewFieldID()

	t.Run("basic roundtrip", func(t *testing.T) {
		t.Parallel()
		items := []db.FieldsJSON{
			{
				FieldID:      string(fieldID),
				ParentID:     "",
				SortOrder:    "5",
				Name:         "title",
				Label:        "Title",
				Data:         "",
				ValidationID: "",
				UIConfig:     "{}",
				Type:         string(types.FieldTypeText),
				Translatable: "true",
				Roles:        "",
				AuthorID:     string(types.NewUserID()),
				DateCreated:  now.Format(time.RFC3339),
				DateModified: now.Format(time.RFC3339),
			},
		}

		result, err := SnapshotFieldsToSlice(items)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := result[0]
		if got.FieldID != fieldID {
			t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
		}
		if got.SortOrder != 5 {
			t.Errorf("SortOrder = %d, want 5", got.SortOrder)
		}
		if !got.Translatable {
			t.Error("Translatable = false, want true")
		}
		if got.Type != types.FieldTypeText {
			t.Errorf("Type = %v, want %v", got.Type, types.FieldTypeText)
		}
	})

	t.Run("translatable=1 parses as true", func(t *testing.T) {
		t.Parallel()
		items := []db.FieldsJSON{
			{
				FieldID:      string(types.NewFieldID()),
				SortOrder:    "0",
				Type:         string(types.FieldTypeText),
				Translatable: "1",
				DateCreated:  now.Format(time.RFC3339),
				DateModified: now.Format(time.RFC3339),
			},
		}

		result, err := SnapshotFieldsToSlice(items)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result[0].Translatable {
			t.Error("Translatable = false for '1', want true")
		}
	})

	t.Run("unparseable sort order defaults to 0", func(t *testing.T) {
		t.Parallel()
		items := []db.FieldsJSON{
			{
				FieldID:      string(types.NewFieldID()),
				SortOrder:    "not-a-number",
				Type:         string(types.FieldTypeText),
				DateCreated:  now.Format(time.RFC3339),
				DateModified: now.Format(time.RFC3339),
			},
		}

		result, err := SnapshotFieldsToSlice(items)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[0].SortOrder != 0 {
			t.Errorf("SortOrder = %d, want 0 for unparseable input", result[0].SortOrder)
		}
	})
}

// ===== UNIT TESTS: IsRevisionConflict =====

func TestIsRevisionConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"unrelated error", errors.New("database timeout"), false},
		{"conflict prefix", fmt.Errorf("conflict: content was modified"), true},
		{"conflict with details", fmt.Errorf("conflict: revision 1 -> 2"), true},
		{"almost conflict", errors.New("conflicting data"), false},
		{"short string", errors.New("conf"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsRevisionConflict(tt.err)
			if got != tt.want {
				t.Errorf("IsRevisionConflict(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ===== UNIT TESTS: Snapshot JSON roundtrip =====

func TestSnapshot_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	snapshot := Snapshot{
		ContentData: []db.ContentDataJSON{
			{ContentDataID: "cd1", Status: "draft", DateCreated: "2025-01-01T00:00:00Z", DateModified: "2025-01-01T00:00:00Z"},
		},
		Datatypes: []db.DatatypeJSON{
			{DatatypeID: "dt1", Label: "Article", Type: "page", DateCreated: "2025-01-01T00:00:00Z", DateModified: "2025-01-01T00:00:00Z"},
		},
		ContentFields: []SnapshotContentFieldJSON{
			{ContentFieldID: "cf1", FieldValue: "test", DateCreated: "2025-01-01T00:00:00Z", DateModified: "2025-01-01T00:00:00Z"},
		},
		Fields: []db.FieldsJSON{
			{FieldID: "f1", Label: "Title", Type: "text", DateCreated: "2025-01-01T00:00:00Z", DateModified: "2025-01-01T00:00:00Z"},
		},
		Route:         SnapshotRoute{RouteID: "r1", Slug: "test", Title: "Test"},
		SchemaVersion: 1,
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Snapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", decoded.SchemaVersion)
	}
	if len(decoded.ContentData) != 1 {
		t.Fatalf("ContentData len = %d, want 1", len(decoded.ContentData))
	}
	if decoded.ContentData[0].ContentDataID != "cd1" {
		t.Errorf("ContentDataID = %q, want %q", decoded.ContentData[0].ContentDataID, "cd1")
	}
	if decoded.Route.Slug != "test" {
		t.Errorf("Route.Slug = %q, want %q", decoded.Route.Slug, "test")
	}
	if len(decoded.ContentFields) != 1 || decoded.ContentFields[0].FieldValue != "test" {
		t.Error("ContentFields roundtrip mismatch")
	}
}

func TestAdminSnapshot_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	snapshot := AdminSnapshot{
		ContentData: []db.ContentDataJSON{
			{ContentDataID: "acd1", Status: "draft", DateCreated: "2025-01-01T00:00:00Z", DateModified: "2025-01-01T00:00:00Z"},
		},
		ContentFields: []AdminSnapshotContentFieldJSON{
			{AdminContentFieldID: "acf1", AdminFieldValue: "admin test"},
		},
		Route:         AdminSnapshotRoute{AdminRouteID: "ar1", Slug: "admin", Title: "Admin"},
		SchemaVersion: 1,
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded AdminSnapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Route.AdminRouteID != "ar1" {
		t.Errorf("AdminRouteID = %q, want %q", decoded.Route.AdminRouteID, "ar1")
	}
	if decoded.ContentFields[0].AdminFieldValue != "admin test" {
		t.Errorf("AdminFieldValue = %q, want %q", decoded.ContentFields[0].AdminFieldValue, "admin test")
	}
}

// ===== INTEGRATION TESTS: BuildSnapshot =====

func TestBuildSnapshot(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, cf := createContentWithField(t, d, seed)

	snapshot, err := BuildSnapshot(d, d.Context, cd.ContentDataID, "")
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	if snapshot.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", snapshot.SchemaVersion)
	}

	// Content data slice should contain the root.
	if len(snapshot.ContentData) == 0 {
		t.Fatal("ContentData is empty")
	}
	if snapshot.ContentData[0].ContentDataID != string(cd.ContentDataID) {
		t.Errorf("ContentData[0].ID = %q, want %q", snapshot.ContentData[0].ContentDataID, cd.ContentDataID)
	}

	// Datatype should be captured.
	if len(snapshot.Datatypes) == 0 {
		t.Fatal("Datatypes is empty")
	}
	if snapshot.Datatypes[0].DatatypeID != string(seed.Datatype.DatatypeID) {
		t.Errorf("Datatypes[0].ID = %q, want %q", snapshot.Datatypes[0].DatatypeID, seed.Datatype.DatatypeID)
	}

	// Content field should be captured.
	if len(snapshot.ContentFields) == 0 {
		t.Fatal("ContentFields is empty")
	}
	if snapshot.ContentFields[0].FieldValue != "Hello World" {
		t.Errorf("ContentFields[0].FieldValue = %q, want %q", snapshot.ContentFields[0].FieldValue, "Hello World")
	}
	if snapshot.ContentFields[0].ContentFieldID != cf.ContentFieldID.String() {
		t.Errorf("ContentFields[0].ContentFieldID = %q, want %q", snapshot.ContentFields[0].ContentFieldID, cf.ContentFieldID.String())
	}

	// Field definition should be captured.
	if len(snapshot.Fields) == 0 {
		t.Fatal("Fields is empty")
	}
	if snapshot.Fields[0].FieldID != string(seed.Field.FieldID) {
		t.Errorf("Fields[0].FieldID = %q, want %q", snapshot.Fields[0].FieldID, seed.Field.FieldID)
	}

	// Route should be captured.
	if snapshot.Route.RouteID != seed.Route.RouteID.String() {
		t.Errorf("Route.RouteID = %q, want %q", snapshot.Route.RouteID, seed.Route.RouteID.String())
	}
	if snapshot.Route.Slug != "test-page" {
		t.Errorf("Route.Slug = %q, want %q", snapshot.Route.Slug, "test-page")
	}
}

func TestBuildSnapshot_NonexistentRoot(t *testing.T) {
	t.Parallel()
	d := testDB(t)
	d.CreateAllTables()

	_, err := BuildSnapshot(d, d.Context, types.NewContentID(), "")
	if err == nil {
		t.Fatal("expected error for nonexistent root")
	}
}

// ===== INTEGRATION TESTS: PublishContent =====

func TestPublishContent(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	dispatcher := &mockDispatcher{}
	indexer := &mockIndexer{}

	version, err := PublishContent(
		d.Context, d, cd.ContentDataID, "", seed.User.UserID,
		ac, 0, false, dispatcher, indexer,
	)
	if err != nil {
		t.Fatalf("PublishContent: %v", err)
	}

	// Version should be created with correct metadata.
	if version == nil {
		t.Fatal("returned nil version")
	}
	if version.VersionNumber != 1 {
		t.Errorf("VersionNumber = %d, want 1", version.VersionNumber)
	}
	if !version.Published {
		t.Error("Published = false, want true")
	}
	if version.Trigger != "publish" {
		t.Errorf("Trigger = %q, want %q", version.Trigger, "publish")
	}
	if version.ContentDataID != cd.ContentDataID {
		t.Errorf("ContentDataID = %v, want %v", version.ContentDataID, cd.ContentDataID)
	}

	// Snapshot should be valid JSON.
	var snapshot Snapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		t.Fatalf("snapshot is not valid JSON: %v", err)
	}
	if snapshot.SchemaVersion != 1 {
		t.Errorf("snapshot SchemaVersion = %d, want 1", snapshot.SchemaVersion)
	}

	// Content data should be updated to published status.
	updated, err := d.GetContentData(cd.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData after publish: %v", err)
	}
	if updated.Status != types.ContentStatusPublished {
		t.Errorf("Status = %v, want published", updated.Status)
	}

	// Webhook should have been dispatched.
	events := dispatcher.getEvents()
	if len(events) != 1 {
		t.Fatalf("dispatcher events = %d, want 1", len(events))
	}
	if events[0].event != "content.published" {
		t.Errorf("event = %q, want %q", events[0].event, "content.published")
	}

	// Indexer should have been called.
	if len(indexer.published) != 1 {
		t.Fatalf("indexer published = %d, want 1", len(indexer.published))
	}
}

func TestPublishContent_SecondPublishIncrementsVersion(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	v1, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("first publish: %v", err)
	}

	v2, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("second publish: %v", err)
	}

	if v2.VersionNumber != v1.VersionNumber+1 {
		t.Errorf("second VersionNumber = %d, want %d", v2.VersionNumber, v1.VersionNumber+1)
	}

	// Only the latest version should be published.
	oldVersion, err := d.GetContentVersion(v1.ContentVersionID)
	if err != nil {
		t.Fatalf("GetContentVersion v1: %v", err)
	}
	if oldVersion.Published {
		t.Error("first version should be unpublished after second publish")
	}
}

func TestPublishContent_WithLocale(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	// Create content data.
	cd, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true},
		AuthorID:     seed.User.UserID,
		Status:       types.ContentStatusDraft,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	// Create content field with locale "en".
	_, err = d.CreateContentField(d.Context, ac, db.CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		RootID:        types.NullableContentID{ID: cd.ContentDataID, Valid: true},
		ContentDataID: types.NullableContentID{ID: cd.ContentDataID, Valid: true},
		FieldID:       types.NullableFieldID{ID: seed.Field.FieldID, Valid: true},
		FieldValue:    "English Title",
		Locale:        "en",
		AuthorID:      seed.User.UserID,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentField: %v", err)
	}

	dispatcher := &mockDispatcher{}
	version, err := PublishContent(d.Context, d, cd.ContentDataID, "en", seed.User.UserID, ac, 0, false, dispatcher, nil)
	if err != nil {
		t.Fatalf("PublishContent with locale: %v", err)
	}

	if version.Locale != "en" {
		t.Errorf("Locale = %q, want %q", version.Locale, "en")
	}

	// Should dispatch both content.published and locale.published events.
	events := dispatcher.getEvents()
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}
	eventNames := map[string]bool{}
	for _, e := range events {
		eventNames[e.event] = true
	}
	if !eventNames["content.published"] {
		t.Error("missing content.published event")
	}
	if !eventNames["locale.published"] {
		t.Error("missing locale.published event")
	}
}

func TestPublishContent_NilDispatcherAndIndexer(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Should not panic with nil dispatcher and indexer.
	_, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent with nil dispatcher/indexer: %v", err)
	}
}

// ===== INTEGRATION TESTS: UnpublishContent =====

func TestUnpublishContent(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish first.
	_, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent: %v", err)
	}

	dispatcher := &mockDispatcher{}
	indexer := &mockIndexer{}

	err = UnpublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, dispatcher, indexer)
	if err != nil {
		t.Fatalf("UnpublishContent: %v", err)
	}

	// Content should be draft.
	updated, err := d.GetContentData(cd.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData after unpublish: %v", err)
	}
	if updated.Status != types.ContentStatusDraft {
		t.Errorf("Status = %v, want draft", updated.Status)
	}

	// Webhook should have fired.
	events := dispatcher.getEvents()
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].event != "content.unpublished" {
		t.Errorf("event = %q, want %q", events[0].event, "content.unpublished")
	}

	// Indexer should have been called.
	if len(indexer.unpublished) != 1 {
		t.Fatalf("indexer unpublished = %d, want 1", len(indexer.unpublished))
	}
	if indexer.unpublished[0] != cd.ContentDataID.String() {
		t.Errorf("unpublished ID = %q, want %q", indexer.unpublished[0], cd.ContentDataID.String())
	}
}

// ===== INTEGRATION TESTS: PruneExcessVersions =====

func TestPruneExcessVersions(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish 5 times to create 5 versions.
	for range 5 {
		_, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
		if err != nil {
			t.Fatalf("PublishContent: %v", err)
		}
	}

	count, err := d.CountContentVersionsByContent(cd.ContentDataID)
	if err != nil {
		t.Fatalf("CountContentVersionsByContent: %v", err)
	}
	if *count != 5 {
		t.Fatalf("version count = %d, want 5", *count)
	}

	// Prune to keep 3.
	PruneExcessVersions(d, cd.ContentDataID, "", 3)

	// Wait briefly for async goroutine (PruneExcessVersions is synchronous in this call).
	count, err = d.CountContentVersionsByContent(cd.ContentDataID)
	if err != nil {
		t.Fatalf("CountContentVersionsByContent after prune: %v", err)
	}
	if *count != 3 {
		t.Errorf("version count after prune = %d, want 3", *count)
	}
}

func TestPruneExcessVersions_ZeroCap(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish 3 times.
	for range 3 {
		_, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
		if err != nil {
			t.Fatalf("PublishContent: %v", err)
		}
	}

	// Zero cap means unlimited: no pruning should occur.
	PruneExcessVersions(d, cd.ContentDataID, "", 0)

	count, err := d.CountContentVersionsByContent(cd.ContentDataID)
	if err != nil {
		t.Fatalf("CountContentVersionsByContent: %v", err)
	}
	if *count != 3 {
		t.Errorf("version count = %d, want 3 (no pruning for cap=0)", *count)
	}
}

func TestPruneExcessVersions_NegativeCap(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	_, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent: %v", err)
	}

	// Negative cap should be treated like zero (disabled).
	PruneExcessVersions(d, cd.ContentDataID, "", -1)

	count, err := d.CountContentVersionsByContent(cd.ContentDataID)
	if err != nil {
		t.Fatalf("CountContentVersionsByContent: %v", err)
	}
	if *count != 1 {
		t.Errorf("version count = %d, want 1 (no pruning for cap=-1)", *count)
	}
}

func TestPruneExcessVersions_UnderCap(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish 2 times, cap is 5.
	for range 2 {
		_, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
		if err != nil {
			t.Fatalf("PublishContent: %v", err)
		}
	}

	PruneExcessVersions(d, cd.ContentDataID, "", 5)

	count, err := d.CountContentVersionsByContent(cd.ContentDataID)
	if err != nil {
		t.Fatalf("CountContentVersionsByContent: %v", err)
	}
	if *count != 2 {
		t.Errorf("version count = %d, want 2 (under cap, no pruning)", *count)
	}
}

// ===== INTEGRATION TESTS: PublishContent with retention cap =====

func TestPublishContent_WithRetentionCap(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish 5 times with cap=0 to avoid async prune goroutines fighting SQLite locks.
	for range 5 {
		_, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
		if err != nil {
			t.Fatalf("PublishContent: %v", err)
		}
	}

	count, err := d.CountContentVersionsByContent(cd.ContentDataID)
	if err != nil {
		t.Fatalf("CountContentVersionsByContent before prune: %v", err)
	}
	if *count != 5 {
		t.Fatalf("version count = %d, want 5", *count)
	}

	// Prune synchronously to cap of 3.
	PruneExcessVersions(d, cd.ContentDataID, "", 3)

	count, err = d.CountContentVersionsByContent(cd.ContentDataID)
	if err != nil {
		t.Fatalf("CountContentVersionsByContent after prune: %v", err)
	}
	if *count != 3 {
		t.Errorf("version count = %d, want 3 (retention cap)", *count)
	}
}

// ===== INTEGRATION TESTS: PublishContent publishAll =====

func TestPublishContent_PublishAll(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	// Create parent content.
	parent, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true},
		AuthorID:     seed.User.UserID,
		Status:       types.ContentStatusDraft,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateContentData parent: %v", err)
	}

	// Create child content.
	child, err := d.CreateContentData(d.Context, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		ParentID:     types.NullableContentID{ID: parent.ContentDataID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true},
		AuthorID:     seed.User.UserID,
		Status:       types.ContentStatusDraft,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateContentData child: %v", err)
	}

	// Link parent to child.
	err = d.UpdateContentDataWithRevision(d.Context, db.UpdateContentDataWithRevisionParams{
		RouteID:       types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{ID: child.ContentDataID, Valid: true},
		DatatypeID:    types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true},
		AuthorID:      seed.User.UserID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
		ContentDataID: parent.ContentDataID,
		Revision:      parent.Revision,
	})
	if err != nil {
		t.Fatalf("UpdateContentDataWithRevision: %v", err)
	}

	// Create content field on parent so snapshot has data.
	_, err = d.CreateContentField(d.Context, ac, db.CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: seed.Route.RouteID, Valid: true},
		RootID:        types.NullableContentID{ID: parent.ContentDataID, Valid: true},
		ContentDataID: types.NullableContentID{ID: parent.ContentDataID, Valid: true},
		FieldID:       types.NullableFieldID{ID: seed.Field.FieldID, Valid: true},
		FieldValue:    "Parent Title",
		AuthorID:      seed.User.UserID,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentField: %v", err)
	}

	// Publish with publishAll=true.
	_, err = PublishContent(d.Context, d, parent.ContentDataID, "", seed.User.UserID, ac, 0, true, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent publishAll: %v", err)
	}

	// Parent should be published.
	parentAfter, err := d.GetContentData(parent.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData parent: %v", err)
	}
	if parentAfter.Status != types.ContentStatusPublished {
		t.Errorf("parent status = %v, want published", parentAfter.Status)
	}

	// Child should also be published.
	childAfter, err := d.GetContentData(child.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData child: %v", err)
	}
	if childAfter.Status != types.ContentStatusPublished {
		t.Errorf("child status = %v, want published", childAfter.Status)
	}
}

// ===== INTEGRATION TESTS: RestoreContent =====

func TestRestoreContent(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish v1 with "Hello World".
	v1, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent v1: %v", err)
	}

	// Update the field value.
	nullableID := types.NullableContentID{ID: cd.ContentDataID, Valid: true}
	fields, err := d.ListContentFieldsByContentData(nullableID)
	if err != nil {
		t.Fatalf("ListContentFieldsByContentData: %v", err)
	}
	if fields == nil || len(*fields) == 0 {
		t.Fatal("no content fields found")
	}
	cf := (*fields)[0]
	_, err = d.UpdateContentField(d.Context, ac, db.UpdateContentFieldParams{
		RouteID:        cf.RouteID,
		ContentDataID:  cf.ContentDataID,
		FieldID:        cf.FieldID,
		FieldValue:     "Updated Value",
		AuthorID:       seed.User.UserID,
		DateCreated:    cf.DateCreated,
		DateModified:   types.TimestampNow(),
		ContentFieldID: cf.ContentFieldID,
	})
	if err != nil {
		t.Fatalf("UpdateContentField: %v", err)
	}

	// Publish v2 with "Updated Value".
	_, err = PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent v2: %v", err)
	}

	// Restore to v1.
	result, err := RestoreContent(d.Context, d, cd.ContentDataID, v1.ContentVersionID, seed.User.UserID, ac)
	if err != nil {
		t.Fatalf("RestoreContent: %v", err)
	}

	if result.FieldsRestored != 1 {
		t.Errorf("FieldsRestored = %d, want 1", result.FieldsRestored)
	}
	if len(result.UnmappedFields) != 0 {
		t.Errorf("UnmappedFields = %v, want empty", result.UnmappedFields)
	}

	// Field value should be back to v1.
	fieldsAfter, err := d.ListContentFieldsByContentData(nullableID)
	if err != nil {
		t.Fatalf("ListContentFieldsByContentData after restore: %v", err)
	}
	if (*fieldsAfter)[0].FieldValue != "Hello World" {
		t.Errorf("FieldValue after restore = %q, want %q", (*fieldsAfter)[0].FieldValue, "Hello World")
	}

	// Content should be reset to draft.
	updated, err := d.GetContentData(cd.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData after restore: %v", err)
	}
	if updated.Status != types.ContentStatusDraft {
		t.Errorf("Status after restore = %v, want draft", updated.Status)
	}
}

func TestRestoreContent_VersionMismatch(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd1, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish content 1.
	v1, err := PublishContent(d.Context, d, cd1.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent: %v", err)
	}

	// Try to restore with a different content_data_id.
	fakeID := types.NewContentID()
	_, err = RestoreContent(d.Context, d, fakeID, v1.ContentVersionID, seed.User.UserID, ac)
	if err == nil {
		t.Fatal("expected error for version belonging to different content_data_id")
	}
}

func TestRestoreContent_UnmappedFields(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	cd, _ := createContentWithField(t, d, seed)
	ac := testAuditCtxWithUser(d, seed.User.UserID)

	// Publish v1 with existing field.
	v1, err := PublishContent(d.Context, d, cd.ContentDataID, "", seed.User.UserID, ac, 0, false, nil, nil)
	if err != nil {
		t.Fatalf("PublishContent: %v", err)
	}

	// Delete the content field (simulates schema drift).
	nullableID := types.NullableContentID{ID: cd.ContentDataID, Valid: true}
	fields, err := d.ListContentFieldsByContentData(nullableID)
	if err != nil {
		t.Fatalf("ListContentFieldsByContentData: %v", err)
	}
	if fields == nil || len(*fields) == 0 {
		t.Fatal("no content fields found")
	}
	err = d.DeleteContentField(d.Context, ac, (*fields)[0].ContentFieldID)
	if err != nil {
		t.Fatalf("DeleteContentField: %v", err)
	}

	// Restore to v1: should report unmapped fields.
	result, err := RestoreContent(d.Context, d, cd.ContentDataID, v1.ContentVersionID, seed.User.UserID, ac)
	if err != nil {
		t.Fatalf("RestoreContent: %v", err)
	}

	if result.FieldsRestored != 0 {
		t.Errorf("FieldsRestored = %d, want 0", result.FieldsRestored)
	}
	if len(result.UnmappedFields) != 1 {
		t.Errorf("UnmappedFields len = %d, want 1", len(result.UnmappedFields))
	}
}

// ===== INTEGRATION TESTS: RestoreResult =====

func TestRestoreResult_ZeroValue(t *testing.T) {
	t.Parallel()
	r := RestoreResult{}
	if r.FieldsRestored != 0 {
		t.Errorf("FieldsRestored zero value = %d, want 0", r.FieldsRestored)
	}
	if r.UnmappedFields != nil {
		t.Errorf("UnmappedFields zero value = %v, want nil", r.UnmappedFields)
	}
}
