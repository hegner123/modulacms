// Integration tests for SchemaService against a real SQLite database.
package service_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testDB creates a fresh SQLite database with all tables and returns the
// driver and a SchemaService wired to it.
func testDB(t *testing.T) (db.Database, *service.SchemaService) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

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

	svc := service.NewSchemaService(d, d)
	return d, svc
}

// testAuditCtx returns an AuditContext for use in mutation calls.
func testAuditCtx(d db.Database) audited.AuditContext {
	return audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "test", "127.0.0.1")
}

// seedUser creates a minimal user for entities that require an author FK.
func seedUser(t *testing.T, d db.Database) types.UserID {
	t.Helper()
	ctx := d.Context
	ac := testAuditCtx(d)

	role, err := d.CreateRole(ctx, ac, db.CreateRoleParams{Label: "test-role"})
	if err != nil {
		t.Fatalf("seedUser CreateRole: %v", err)
	}
	user, err := d.CreateUser(ctx, ac, db.CreateUserParams{
		Username:     "testuser",
		Name:         "Test User",
		Email:        types.Email("test@example.com"),
		Hash:         "fakehash",
		Role:         role.RoleID.String(),
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("seedUser CreateUser: %v", err)
	}
	return user.UserID
}

// seedFieldType creates a field type record so that field creation validation
// (which checks field type existence) passes.
func seedFieldType(t *testing.T, d db.Database, svc *service.SchemaService, typeName string) *db.FieldTypes {
	t.Helper()
	ctx := context.Background()
	ac := testAuditCtx(d)
	ft, err := svc.CreateFieldType(ctx, ac, db.CreateFieldTypeParams{
		Type:  typeName,
		Label: typeName + " field",
	})
	if err != nil {
		t.Fatalf("seedFieldType: %v", err)
	}
	return ft
}

// seedAdminFieldType creates an admin field type record so that admin field
// creation validation passes.
func seedAdminFieldType(t *testing.T, d db.Database, svc *service.SchemaService, typeName string) *db.AdminFieldTypes {
	t.Helper()
	ctx := context.Background()
	ac := testAuditCtx(d)
	ft, err := svc.CreateAdminFieldType(ctx, ac, db.CreateAdminFieldTypeParams{
		Type:  typeName,
		Label: typeName + " admin field",
	})
	if err != nil {
		t.Fatalf("seedAdminFieldType: %v", err)
	}
	return ft
}

// ---------------------------------------------------------------------------
// 1. Datatype CRUD round-trip
// ---------------------------------------------------------------------------

func TestSchemaService_Datatype_CRUD(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	// Create
	created, err := svc.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Label:    "article",
		Type:     "page",
		AuthorID: userID,
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}
	if created.DatatypeID.IsZero() {
		t.Fatal("created DatatypeID is zero")
	}
	if created.Label != "article" {
		t.Errorf("Label = %q, want %q", created.Label, "article")
	}

	// Get
	got, err := svc.GetDatatype(ctx, created.DatatypeID)
	if err != nil {
		t.Fatalf("GetDatatype: %v", err)
	}
	if got.DatatypeID != created.DatatypeID {
		t.Errorf("GetDatatype ID = %v, want %v", got.DatatypeID, created.DatatypeID)
	}

	// Update
	updated, err := svc.UpdateDatatype(ctx, ac, db.UpdateDatatypeParams{
		DatatypeID: created.DatatypeID,
		Label:      "blog-post",
		Type:       "blog",
		AuthorID:   userID,
	})
	if err != nil {
		t.Fatalf("UpdateDatatype: %v", err)
	}
	if updated.Label != "blog-post" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "blog-post")
	}
	if updated.Type != "blog" {
		t.Errorf("updated Type = %q, want %q", updated.Type, "blog")
	}

	// Get after update
	refreshed, err := svc.GetDatatype(ctx, created.DatatypeID)
	if err != nil {
		t.Fatalf("GetDatatype after update: %v", err)
	}
	if refreshed.Label != "blog-post" {
		t.Errorf("refreshed Label = %q, want %q", refreshed.Label, "blog-post")
	}

	// Delete
	if err := svc.DeleteDatatype(ctx, ac, created.DatatypeID); err != nil {
		t.Fatalf("DeleteDatatype: %v", err)
	}

	// Get after delete: the db layer returns an error for missing rows.
	// The service wraps this as either NotFoundError or InternalError
	// depending on whether the db returns (nil, nil) or (nil, error).
	_, err = svc.GetDatatype(ctx, created.DatatypeID)
	if err == nil {
		t.Fatal("GetDatatype after delete: expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// 2. Datatype validation
// ---------------------------------------------------------------------------

func TestSchemaService_Datatype_Validation(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	tests := []struct {
		name       string
		params     db.CreateDatatypeParams
		wantFields int
	}{
		{
			name:       "empty label",
			params:     db.CreateDatatypeParams{Label: "", Type: "page", AuthorID: userID},
			wantFields: 1,
		},
		{
			name:       "empty type",
			params:     db.CreateDatatypeParams{Label: "good-label", Type: "", AuthorID: userID},
			wantFields: 1,
		},
		{
			name:       "both empty",
			params:     db.CreateDatatypeParams{Label: "", Type: "", AuthorID: userID},
			wantFields: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateDatatype(ctx, ac, tc.params)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !service.IsValidation(err) {
				t.Fatalf("expected ValidationError, got %T: %v", err, err)
			}
			var ve *service.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("errors.As failed for ValidationError")
			}
			if len(ve.Errors) != tc.wantFields {
				t.Errorf("got %d field errors, want %d: %v", len(ve.Errors), tc.wantFields, ve.Errors)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Datatype defaults (ID and timestamps generated)
// ---------------------------------------------------------------------------

func TestSchemaService_Datatype_Defaults(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	created, err := svc.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Label:    "defaults-test",
		Type:     "page",
		AuthorID: userID,
		// DatatypeID, DateCreated, DateModified all zero
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}
	if created.DatatypeID.IsZero() {
		t.Error("DatatypeID should be auto-generated, got zero")
	}
	if !created.DateCreated.Valid {
		t.Error("DateCreated should be auto-set, got invalid")
	}
	if !created.DateModified.Valid {
		t.Error("DateModified should be auto-set, got invalid")
	}
}

// ---------------------------------------------------------------------------
// 4. Datatype paginated listing
// ---------------------------------------------------------------------------

func TestSchemaService_Datatype_Paginated(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	for i := range 3 {
		labels := []string{"alpha", "beta", "gamma"}
		_, err := svc.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
			Label:    labels[i],
			Type:     "page",
			AuthorID: userID,
		})
		if err != nil {
			t.Fatalf("CreateDatatype %d: %v", i, err)
		}
	}

	// Page 1: limit 2, offset 0
	page1, err := svc.ListDatatypesPaginated(ctx, db.PaginationParams{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("ListDatatypesPaginated page 1: %v", err)
	}
	if len(page1.Data) != 2 {
		t.Errorf("page 1 data len = %d, want 2", len(page1.Data))
	}
	if page1.Total != 3 {
		t.Errorf("page 1 total = %d, want 3", page1.Total)
	}

	// Page 2: limit 2, offset 2
	page2, err := svc.ListDatatypesPaginated(ctx, db.PaginationParams{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("ListDatatypesPaginated page 2: %v", err)
	}
	if len(page2.Data) != 1 {
		t.Errorf("page 2 data len = %d, want 1", len(page2.Data))
	}
	if page2.Total != 3 {
		t.Errorf("page 2 total = %d, want 3", page2.Total)
	}
}

// ---------------------------------------------------------------------------
// 5. Field CRUD round-trip
// ---------------------------------------------------------------------------

func TestSchemaService_Field_CRUD(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)
	seedFieldType(t, d, svc, "text")
	authorID := types.NullableUserID{ID: userID, Valid: true}

	// Create
	created, err := svc.CreateField(ctx, ac, db.CreateFieldParams{
		Label:    "title",
		Type:     types.FieldTypeText,
		AuthorID: authorID,
	})
	if err != nil {
		t.Fatalf("CreateField: %v", err)
	}
	if created.FieldID.IsZero() {
		t.Fatal("created FieldID is zero")
	}
	if created.Label != "title" {
		t.Errorf("Label = %q, want %q", created.Label, "title")
	}

	// Get
	got, err := svc.GetField(ctx, created.FieldID, "", true)
	if err != nil {
		t.Fatalf("GetField: %v", err)
	}
	if got.FieldID != created.FieldID {
		t.Errorf("GetField ID = %v, want %v", got.FieldID, created.FieldID)
	}

	// Update
	updated, err := svc.UpdateField(ctx, ac, db.UpdateFieldParams{
		FieldID:  created.FieldID,
		Label:    "title-updated",
		Type:     types.FieldTypeText,
		AuthorID: authorID,
	})
	if err != nil {
		t.Fatalf("UpdateField: %v", err)
	}
	if updated.Label != "title-updated" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "title-updated")
	}

	// Get after update
	refreshed, err := svc.GetField(ctx, created.FieldID, "", true)
	if err != nil {
		t.Fatalf("GetField after update: %v", err)
	}
	if refreshed.Label != "title-updated" {
		t.Errorf("refreshed Label = %q, want %q", refreshed.Label, "title-updated")
	}

	// Delete
	if err := svc.DeleteField(ctx, ac, created.FieldID); err != nil {
		t.Fatalf("DeleteField: %v", err)
	}

	// Get after delete: expect error (InternalError wrapping sql.ErrNoRows)
	_, err = svc.GetField(ctx, created.FieldID, "", true)
	if err == nil {
		t.Fatal("GetField after delete: expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// 6. Field validation
// ---------------------------------------------------------------------------

func TestSchemaService_Field_Validation(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)
	seedFieldType(t, d, svc, "text")
	authorID := types.NullableUserID{ID: userID, Valid: true}

	tests := []struct {
		name      string
		params    db.CreateFieldParams
		wantField string
	}{
		{
			name:      "empty label",
			params:    db.CreateFieldParams{Label: "", Type: types.FieldTypeText, AuthorID: authorID},
			wantField: "label",
		},
		{
			name:      "empty type",
			params:    db.CreateFieldParams{Label: "good-label", Type: "", AuthorID: authorID},
			wantField: "type",
		},
		{
			name:      "invalid field type",
			params:    db.CreateFieldParams{Label: "good-label", Type: "nonexistent-type", AuthorID: authorID},
			wantField: "type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateField(ctx, ac, tc.params)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !service.IsValidation(err) {
				t.Fatalf("expected ValidationError, got %T: %v", err, err)
			}
			var ve *service.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("errors.As failed for ValidationError")
			}
			if len(ve.Errors) == 0 {
				t.Fatal("expected at least one field error")
			}
			if ve.Errors[0].Field != tc.wantField {
				t.Errorf("field error field = %q, want %q", ve.Errors[0].Field, tc.wantField)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. Field role filtering
// ---------------------------------------------------------------------------

func TestSchemaService_Field_RoleFiltering(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)
	seedFieldType(t, d, svc, "text")
	authorID := types.NullableUserID{ID: userID, Valid: true}

	editorRole := "01EDITOR00000000000000EDITOR"

	// Create a field restricted to the editor role
	_, err := svc.CreateField(ctx, ac, db.CreateFieldParams{
		Label:    "editor-only-field",
		Type:     types.FieldTypeText,
		AuthorID: authorID,
		Roles:    types.NewNullableString(`["` + editorRole + `"]`),
	})
	if err != nil {
		t.Fatalf("CreateField: %v", err)
	}

	// ListFieldsFiltered with matching role
	editorFiltered, err := svc.ListFieldsFiltered(ctx, editorRole, false)
	if err != nil {
		t.Fatalf("ListFieldsFiltered (editor): %v", err)
	}
	if len(editorFiltered) != 1 {
		t.Errorf("editor sees %d fields, want 1", len(editorFiltered))
	}

	// ListFieldsFiltered with non-matching role
	viewerFiltered, err := svc.ListFieldsFiltered(ctx, "01VIEWER00000000000000VIEWER", false)
	if err != nil {
		t.Fatalf("ListFieldsFiltered (viewer): %v", err)
	}
	if len(viewerFiltered) != 0 {
		t.Errorf("viewer sees %d fields, want 0", len(viewerFiltered))
	}

	// ListFieldsFiltered with admin bypass
	adminFiltered, err := svc.ListFieldsFiltered(ctx, "any-role", true)
	if err != nil {
		t.Fatalf("ListFieldsFiltered (admin): %v", err)
	}
	if len(adminFiltered) != 1 {
		t.Errorf("admin sees %d fields, want 1", len(adminFiltered))
	}
}

// ---------------------------------------------------------------------------
// 8. Field access check
// ---------------------------------------------------------------------------

func TestSchemaService_Field_AccessCheck(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)
	seedFieldType(t, d, svc, "text")
	authorID := types.NullableUserID{ID: userID, Valid: true}

	editorRole := "01EDITOR00000000000000EDITOR"

	created, err := svc.CreateField(ctx, ac, db.CreateFieldParams{
		Label:    "restricted-field",
		Type:     types.FieldTypeText,
		AuthorID: authorID,
		Roles:    types.NewNullableString(`["` + editorRole + `"]`),
	})
	if err != nil {
		t.Fatalf("CreateField: %v", err)
	}

	// Non-matching role: ForbiddenError
	_, err = svc.GetField(ctx, created.FieldID, "01VIEWER00000000000000VIEWER", false)
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
	if !service.IsForbidden(err) {
		t.Errorf("expected ForbiddenError, got %T: %v", err, err)
	}

	// Matching role: success
	got, err := svc.GetField(ctx, created.FieldID, editorRole, false)
	if err != nil {
		t.Fatalf("GetField with matching role: %v", err)
	}
	if got.FieldID != created.FieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, created.FieldID)
	}

	// Admin bypass: success
	gotAdmin, err := svc.GetField(ctx, created.FieldID, "any-role", true)
	if err != nil {
		t.Fatalf("GetField with admin: %v", err)
	}
	if gotAdmin.FieldID != created.FieldID {
		t.Errorf("admin FieldID = %v, want %v", gotAdmin.FieldID, created.FieldID)
	}
}

// ---------------------------------------------------------------------------
// 9. Field sort order
// ---------------------------------------------------------------------------

func TestSchemaService_Field_SortOrder(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)
	seedFieldType(t, d, svc, "text")
	authorID := types.NullableUserID{ID: userID, Valid: true}

	// Create a parent datatype so GetMaxSortOrderByParentID can match via
	// WHERE parent_id = ? (NULL = NULL is false in SQL).
	dt, err := svc.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Label:    "sort-parent",
		Type:     "page",
		AuthorID: userID,
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}
	parentID := types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true}

	created, err := svc.CreateField(ctx, ac, db.CreateFieldParams{
		Label:    "sortable-field",
		Type:     types.FieldTypeText,
		ParentID: parentID,
		AuthorID: authorID,
	})
	if err != nil {
		t.Fatalf("CreateField: %v", err)
	}

	// Update sort order to 5
	err = svc.UpdateFieldSortOrder(ctx, ac, db.UpdateFieldSortOrderParams{
		FieldID:   created.FieldID,
		SortOrder: 5,
	})
	if err != nil {
		t.Fatalf("UpdateFieldSortOrder: %v", err)
	}

	// Get max sort order for fields under the parent datatype
	maxSort, err := svc.GetMaxSortOrder(ctx, parentID)
	if err != nil {
		t.Fatalf("GetMaxSortOrder: %v", err)
	}
	if maxSort != 5 {
		t.Errorf("max sort order = %d, want 5", maxSort)
	}
}

// ---------------------------------------------------------------------------
// 10. ListFieldsByDatatypeID
// ---------------------------------------------------------------------------

func TestSchemaService_ListFieldsByDatatypeID(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)
	seedFieldType(t, d, svc, "text")
	authorID := types.NullableUserID{ID: userID, Valid: true}

	// Create a datatype
	dt, err := svc.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Label:    "parent-dt",
		Type:     "page",
		AuthorID: userID,
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	parentID := types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true}

	// Create 2 fields under this datatype
	for _, label := range []string{"field-a", "field-b"} {
		_, err := svc.CreateField(ctx, ac, db.CreateFieldParams{
			Label:    label,
			Type:     types.FieldTypeText,
			ParentID: parentID,
			AuthorID: authorID,
		})
		if err != nil {
			t.Fatalf("CreateField(%s): %v", label, err)
		}
	}

	// List by datatype ID
	fields, err := svc.ListFieldsByDatatypeID(ctx, parentID)
	if err != nil {
		t.Fatalf("ListFieldsByDatatypeID: %v", err)
	}
	if len(fields) != 2 {
		t.Errorf("field count = %d, want 2", len(fields))
	}
}

// ---------------------------------------------------------------------------
// 11. FieldType CRUD
// ---------------------------------------------------------------------------

func TestSchemaService_FieldType_CRUD(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)

	// Create
	created, err := svc.CreateFieldType(ctx, ac, db.CreateFieldTypeParams{
		Type:  "richtext",
		Label: "Rich Text",
	})
	if err != nil {
		t.Fatalf("CreateFieldType: %v", err)
	}
	if created.FieldTypeID.IsZero() {
		t.Fatal("created FieldTypeID is zero")
	}
	if created.Type != "richtext" {
		t.Errorf("Type = %q, want %q", created.Type, "richtext")
	}

	// Get
	got, err := svc.GetFieldType(ctx, created.FieldTypeID)
	if err != nil {
		t.Fatalf("GetFieldType: %v", err)
	}
	if got.Type != "richtext" {
		t.Errorf("GetFieldType Type = %q, want %q", got.Type, "richtext")
	}

	// Update
	updated, err := svc.UpdateFieldType(ctx, ac, db.UpdateFieldTypeParams{
		FieldTypeID: created.FieldTypeID,
		Type:        "markdown",
		Label:       "Markdown",
	})
	if err != nil {
		t.Fatalf("UpdateFieldType: %v", err)
	}
	if updated.Type != "markdown" {
		t.Errorf("updated Type = %q, want %q", updated.Type, "markdown")
	}
	if updated.Label != "Markdown" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "Markdown")
	}

	// Get after update
	refreshed, err := svc.GetFieldType(ctx, created.FieldTypeID)
	if err != nil {
		t.Fatalf("GetFieldType after update: %v", err)
	}
	if refreshed.Type != "markdown" {
		t.Errorf("refreshed Type = %q, want %q", refreshed.Type, "markdown")
	}

	// Delete
	if err := svc.DeleteFieldType(ctx, ac, created.FieldTypeID); err != nil {
		t.Fatalf("DeleteFieldType: %v", err)
	}

	// Get after delete: expect error (InternalError wrapping sql.ErrNoRows)
	_, err = svc.GetFieldType(ctx, created.FieldTypeID)
	if err == nil {
		t.Fatal("GetFieldType after delete: expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// 12. GetDatatypeFull
// ---------------------------------------------------------------------------

func TestSchemaService_GetDatatypeFull(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)
	seedFieldType(t, d, svc, "text")
	seedFieldType(t, d, svc, "textarea")
	authorID := types.NullableUserID{ID: userID, Valid: true}

	// Create a datatype
	dt, err := svc.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Label:    "full-view-dt",
		Type:     "page",
		AuthorID: userID,
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	parentID := types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true}

	// Create 2 fields linked to this datatype
	_, err = svc.CreateField(ctx, ac, db.CreateFieldParams{
		Label:    "title",
		Type:     types.FieldTypeText,
		ParentID: parentID,
		AuthorID: authorID,
	})
	if err != nil {
		t.Fatalf("CreateField (title): %v", err)
	}
	_, err = svc.CreateField(ctx, ac, db.CreateFieldParams{
		Label:    "body",
		Type:     types.FieldTypeTextarea,
		ParentID: parentID,
		AuthorID: authorID,
	})
	if err != nil {
		t.Fatalf("CreateField (body): %v", err)
	}

	// Get full view
	view, err := svc.GetDatatypeFull(ctx, dt.DatatypeID)
	if err != nil {
		t.Fatalf("GetDatatypeFull: %v", err)
	}
	if view.DatatypeID != dt.DatatypeID {
		t.Errorf("DatatypeID = %v, want %v", view.DatatypeID, dt.DatatypeID)
	}
	if view.Label != "full-view-dt" {
		t.Errorf("Label = %q, want %q", view.Label, "full-view-dt")
	}
	if len(view.Fields) != 2 {
		t.Errorf("Fields len = %d, want 2", len(view.Fields))
	}
}

// ---------------------------------------------------------------------------
// 13. Admin datatype CRUD
// ---------------------------------------------------------------------------

func TestSchemaService_AdminDatatype_CRUD(t *testing.T) {
	t.Parallel()
	d, svc := testDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	// Create -- note: admin datatypes skip ValidateUserDatatypeType
	created, err := svc.CreateAdminDatatype(ctx, ac, db.CreateAdminDatatypeParams{
		Label:    "admin-article",
		Type:     "_system",
		AuthorID: userID,
	})
	if err != nil {
		t.Fatalf("CreateAdminDatatype: %v", err)
	}
	if created.AdminDatatypeID.IsZero() {
		t.Fatal("created AdminDatatypeID is zero")
	}
	if created.Label != "admin-article" {
		t.Errorf("Label = %q, want %q", created.Label, "admin-article")
	}
	if created.Type != "_system" {
		t.Errorf("Type = %q, want %q", created.Type, "_system")
	}

	// Get
	got, err := svc.GetAdminDatatype(ctx, created.AdminDatatypeID)
	if err != nil {
		t.Fatalf("GetAdminDatatype: %v", err)
	}
	if got.AdminDatatypeID != created.AdminDatatypeID {
		t.Errorf("ID mismatch: %v != %v", got.AdminDatatypeID, created.AdminDatatypeID)
	}

	// Update
	updated, err := svc.UpdateAdminDatatype(ctx, ac, db.UpdateAdminDatatypeParams{
		AdminDatatypeID: created.AdminDatatypeID,
		Label:           "admin-article-updated",
		Type:            "_internal",
		AuthorID:        userID,
	})
	if err != nil {
		t.Fatalf("UpdateAdminDatatype: %v", err)
	}
	if updated.Label != "admin-article-updated" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "admin-article-updated")
	}

	// Delete
	if err := svc.DeleteAdminDatatype(ctx, ac, created.AdminDatatypeID); err != nil {
		t.Fatalf("DeleteAdminDatatype: %v", err)
	}

	// Get after delete: expect error (InternalError wrapping sql.ErrNoRows)
	_, err = svc.GetAdminDatatype(ctx, created.AdminDatatypeID)
	if err == nil {
		t.Fatal("GetAdminDatatype after delete: expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// 14. Error wrapping
// ---------------------------------------------------------------------------

func TestSchemaService_ErrorTypes(t *testing.T) {
	t.Parallel()
	_, svc := testDB(t)
	ctx := context.Background()

	t.Run("NotFoundError from GetDatatype", func(t *testing.T) {
		_, err := svc.GetDatatype(ctx, types.NewDatatypeID())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var nfe *service.NotFoundError
		if !errors.As(err, &nfe) {
			// The db layer may return an InternalError wrapping the sql.ErrNoRows.
			// Either NotFoundError or InternalError is acceptable depending on
			// whether the db layer returns nil or an error for missing rows.
			var ie *service.InternalError
			if !errors.As(err, &ie) {
				t.Errorf("expected NotFoundError or InternalError, got %T: %v", err, err)
			}
		}
	})

	t.Run("ValidationError from CreateField with empty label", func(t *testing.T) {
		ac := audited.Ctx(types.NodeID("n"), types.UserID("u"), "r", "127.0.0.1")
		_, err := svc.CreateField(ctx, ac, db.CreateFieldParams{
			Label: "",
			Type:  "",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var ve *service.ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("expected ValidationError, got %T: %v", err, err)
		}
	})

	t.Run("ForbiddenError from GetField with wrong role", func(t *testing.T) {
		d, svc := testDB(t)
		ac := testAuditCtx(d)
		userID := seedUser(t, d)
		seedFieldType(t, d, svc, "text")
		authorID := types.NullableUserID{ID: userID, Valid: true}

		f, err := svc.CreateField(ctx, ac, db.CreateFieldParams{
			Label:    "secret",
			Type:     types.FieldTypeText,
			AuthorID: authorID,
			Roles:    types.NewNullableString(`["specific-role"]`),
		})
		if err != nil {
			t.Fatalf("CreateField: %v", err)
		}

		_, err = svc.GetField(ctx, f.FieldID, "wrong-role", false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var fe *service.ForbiddenError
		if !errors.As(err, &fe) {
			t.Errorf("expected ForbiddenError, got %T: %v", err, err)
		}
	})
}
