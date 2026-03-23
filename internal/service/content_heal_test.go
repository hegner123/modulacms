package service_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type noopDispatcher struct{}

func (noopDispatcher) Dispatch(_ context.Context, _ string, _ map[string]any) {}

var _ publishing.WebhookDispatcher = noopDispatcher{}

func testHealDB(t *testing.T) (db.Database, *service.ContentService) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "heal_test.db")

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

	cfg := config.Config{
		Node_ID: types.NewNodeID().String(),
	}

	d := db.Database{
		Connection: conn,
		Context:    context.Background(),
		Config:     cfg,
	}

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	mgr := config.NewManager(&staticProvider{cfg: &cfg})
	if err := mgr.Load(); err != nil {
		t.Fatalf("mgr.Load: %v", err)
	}

	svc := service.NewContentService(d, mgr, noopDispatcher{})
	return d, svc
}

func testAdminHealDB(t *testing.T) (db.Database, *service.AdminContentService) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "admin_heal_test.db")

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

	cfg := config.Config{
		Node_ID: types.NewNodeID().String(),
	}

	d := db.Database{
		Connection: conn,
		Context:    context.Background(),
		Config:     cfg,
	}

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	mgr := config.NewManager(&staticProvider{cfg: &cfg})
	if err := mgr.Load(); err != nil {
		t.Fatalf("mgr.Load: %v", err)
	}

	svc := service.NewAdminContentService(d, mgr, noopDispatcher{})
	return d, svc
}

// healAC and seedUser are defined in schema_test.go (same package).

// ---------------------------------------------------------------------------
// Pass 7: Orphaned route references
// ---------------------------------------------------------------------------

func TestHeal_OrphanedRouteRefs_DryRun(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	dt, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "page",
		Label:        "Page",
		Type:         "_root",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	// Insert content_data with a fake route_id (bypassing FK) to simulate
	// the state where a route was deleted but ON DELETE SET NULL didn't fire
	// (e.g., restored from backup, cross-db migration, or manual edit).
	fakeRouteID := types.NewRouteID()
	cdID := types.NewContentID()
	now := types.TimestampNow()
	// Temporarily disable FK checks to insert an orphaned route reference.
	_, _ = d.Connection.ExecContext(ctx, "PRAGMA foreign_keys=OFF;")
	_, execErr := d.Connection.ExecContext(ctx,
		`INSERT INTO content_data (content_data_id, route_id, datatype_id, author_id, status, date_created, date_modified)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		string(cdID), string(fakeRouteID), string(dt.DatatypeID), string(userID), "draft", now.String(), now.String(),
	)
	_, _ = d.Connection.ExecContext(ctx, "PRAGMA foreign_keys=ON;")
	if execErr != nil {
		t.Fatalf("raw INSERT: %v", execErr)
	}

	report, err := svc.Heal(ctx, ac, true)
	if err != nil {
		t.Fatalf("Heal dry run: %v", err)
	}

	if len(report.OrphanedRouteRefs) != 1 {
		t.Fatalf("expected 1 orphaned route ref, got %d", len(report.OrphanedRouteRefs))
	}
	if report.OrphanedRouteRefs[0].Nulled {
		t.Error("dry run should not null the reference")
	}
}

func TestHeal_OrphanedRouteRefs_Heal(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	dt, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "page",
		Label:        "Page",
		Type:         "_root",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	// Insert with fake route_id to bypass FK cascade.
	fakeRouteID := types.NewRouteID()
	cdID := types.NewContentID()
	now := types.TimestampNow()
	// Temporarily disable FK checks to insert an orphaned route reference.
	_, _ = d.Connection.ExecContext(ctx, "PRAGMA foreign_keys=OFF;")
	_, execErr := d.Connection.ExecContext(ctx,
		`INSERT INTO content_data (content_data_id, route_id, datatype_id, author_id, status, date_created, date_modified)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		string(cdID), string(fakeRouteID), string(dt.DatatypeID), string(userID), "draft", now.String(), now.String(),
	)
	_, _ = d.Connection.ExecContext(ctx, "PRAGMA foreign_keys=ON;")
	if execErr != nil {
		t.Fatalf("raw INSERT: %v", execErr)
	}

	report, err := svc.Heal(ctx, ac, false)
	if err != nil {
		t.Fatalf("Heal: %v", err)
	}

	if len(report.OrphanedRouteRefs) != 1 {
		t.Fatalf("expected 1 orphaned route ref, got %d", len(report.OrphanedRouteRefs))
	}
	if !report.OrphanedRouteRefs[0].Nulled {
		t.Error("heal should have nulled the reference")
	}

	updated, err := d.GetContentData(cdID)
	if err != nil {
		t.Fatalf("GetContentData: %v", err)
	}
	if updated.RouteID.Valid {
		t.Errorf("route_id should be null after heal, got %s", updated.RouteID.ID)
	}
}

// ---------------------------------------------------------------------------
// Pass 8: Unrouted roots
// ---------------------------------------------------------------------------

func TestHeal_UnroutedRoots(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	dt, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "page",
		Label:        "Page",
		Type:         "_root",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	// Root node with no route and no parent.
	_, err = d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		DatatypeID:   types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true},
		AuthorID:     userID,
		Status:       types.ContentStatusDraft,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	report, err := svc.Heal(ctx, ac, true)
	if err != nil {
		t.Fatalf("Heal: %v", err)
	}

	if len(report.UnroutedRoots) != 1 {
		t.Fatalf("expected 1 unrouted root, got %d", len(report.UnroutedRoots))
	}
	if report.UnroutedRoots[0].DatatypeName != "page" {
		t.Errorf("expected datatype name 'page', got %q", report.UnroutedRoots[0].DatatypeName)
	}
}

func TestHeal_UnroutedRoots_ChildNodesIgnored(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	// Non-root datatype — should NOT be reported even with no route and no parent.
	dt, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "section",
		Label:        "Section",
		Type:         "section",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	_, err = d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		DatatypeID:   types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true},
		AuthorID:     userID,
		Status:       types.ContentStatusDraft,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	report, err := svc.Heal(ctx, ac, true)
	if err != nil {
		t.Fatalf("Heal: %v", err)
	}

	if len(report.UnroutedRoots) != 0 {
		t.Errorf("expected 0 unrouted roots for non-root type, got %d", len(report.UnroutedRoots))
	}
}

// ---------------------------------------------------------------------------
// Pass 9: Rootless content
// ---------------------------------------------------------------------------

func TestHeal_RootlessContent_DryRun(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	route, err := d.CreateRoute(ctx, ac, db.CreateRouteParams{
		Slug:         "broken",
		Title:        "Broken",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateRoute: %v", err)
	}

	// A _reference type — NOT a root.
	dt, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "menu-ref",
		Label:        "Menu Reference",
		Type:         "_reference_menu",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	_, err = d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: route.RouteID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true},
		AuthorID:     userID,
		Status:       types.ContentStatusDraft,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	report, err := svc.Heal(ctx, ac, true)
	if err != nil {
		t.Fatalf("Heal dry run: %v", err)
	}

	if len(report.RootlessContent) != 1 {
		t.Fatalf("expected 1 rootless content, got %d", len(report.RootlessContent))
	}
	if report.RootlessContent[0].RouteSlug != "broken" {
		t.Errorf("expected route slug 'broken', got %q", report.RootlessContent[0].RouteSlug)
	}
	if report.RootlessContent[0].Deleted {
		t.Error("dry run should not delete")
	}
}

func TestHeal_RootlessContent_Heal(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	route, err := d.CreateRoute(ctx, ac, db.CreateRouteParams{
		Slug:         "orphaned",
		Title:        "Orphaned",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateRoute: %v", err)
	}

	dt, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "menu-ref",
		Label:        "Menu Reference",
		Type:         "_reference_menu",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	field, err := d.CreateField(ctx, ac, db.CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true},
		Label:        "Ref ID",
		Type:         "_id",
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateField: %v", err)
	}

	cd, err := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: route.RouteID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true},
		AuthorID:     userID,
		Status:       types.ContentStatusDraft,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	_, err = d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: route.RouteID, Valid: true},
		ContentDataID: types.NullableContentID{ID: cd.ContentDataID, Valid: true},
		FieldID:       types.NullableFieldID{ID: field.FieldID, Valid: true},
		FieldValue:    "",
		AuthorID:      userID,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentField: %v", err)
	}

	// Heal should delete both content_data and content_fields.
	report, err := svc.Heal(ctx, ac, false)
	if err != nil {
		t.Fatalf("Heal: %v", err)
	}

	if len(report.RootlessContent) != 1 {
		t.Fatalf("expected 1 rootless content, got %d", len(report.RootlessContent))
	}
	if !report.RootlessContent[0].Deleted {
		t.Error("heal should have deleted the rootless content")
	}

	// Verify content_data is gone.
	_, getErr := d.GetContentData(cd.ContentDataID)
	if getErr == nil {
		t.Error("content_data should have been deleted")
	}

	// Verify content_fields are gone.
	fields, _ := d.ListContentFieldsByContentData(types.NullableContentID{ID: cd.ContentDataID, Valid: true})
	if fields != nil && len(*fields) > 0 {
		t.Errorf("content_fields should have been deleted, got %d", len(*fields))
	}
}

func TestHeal_RootlessContent_RoutesWithRootAreSkipped(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	route, err := d.CreateRoute(ctx, ac, db.CreateRouteParams{
		Slug:         "healthy",
		Title:        "Healthy",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateRoute: %v", err)
	}

	rootDT, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "page",
		Label:        "Page",
		Type:         "_root",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype (root): %v", err)
	}

	childDT, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "section",
		Label:        "Section",
		Type:         "section",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype (child): %v", err)
	}

	// Create root + child on same route — should NOT be flagged.
	_, err = d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: route.RouteID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: rootDT.DatatypeID, Valid: true},
		AuthorID:     userID,
		Status:       types.ContentStatusDraft,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentData (root): %v", err)
	}

	_, err = d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: route.RouteID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: childDT.DatatypeID, Valid: true},
		AuthorID:     userID,
		Status:       types.ContentStatusDraft,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentData (child): %v", err)
	}

	report, err := svc.Heal(ctx, ac, true)
	if err != nil {
		t.Fatalf("Heal: %v", err)
	}

	if len(report.RootlessContent) != 0 {
		t.Errorf("expected 0 rootless content for healthy route, got %d", len(report.RootlessContent))
	}
}

// ---------------------------------------------------------------------------
// Healthy content produces clean report
// ---------------------------------------------------------------------------

func TestHeal_HealthyContent_NoIssues(t *testing.T) {
	t.Parallel()
	d, svc := testHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	route, err := d.CreateRoute(ctx, ac, db.CreateRouteParams{
		Slug:         "blog",
		Title:        "Blog",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateRoute: %v", err)
	}

	dt, err := d.CreateDatatype(ctx, ac, db.CreateDatatypeParams{
		Name:         "blog-root",
		Label:        "Blog Root",
		Type:         "_root",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}

	_, err = d.CreateContentData(ctx, ac, db.CreateContentDataParams{
		RouteID:      types.NullableRouteID{ID: route.RouteID, Valid: true},
		DatatypeID:   types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true},
		AuthorID:     userID,
		Status:       types.ContentStatusDraft,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	report, err := svc.Heal(ctx, ac, true)
	if err != nil {
		t.Fatalf("Heal: %v", err)
	}

	if len(report.OrphanedRouteRefs) != 0 {
		t.Errorf("expected 0 orphaned route refs, got %d", len(report.OrphanedRouteRefs))
	}
	if len(report.UnroutedRoots) != 0 {
		t.Errorf("expected 0 unrouted roots, got %d", len(report.UnroutedRoots))
	}
	if len(report.RootlessContent) != 0 {
		t.Errorf("expected 0 rootless content, got %d", len(report.RootlessContent))
	}
}

// ---------------------------------------------------------------------------
// Admin heal — pass 9 rootless content
// ---------------------------------------------------------------------------

func TestAdminHeal_RootlessContent_DryRun(t *testing.T) {
	t.Parallel()
	d, svc := testAdminHealDB(t)
	ctx := context.Background()
	ac := testAuditCtx(d)
	userID := seedUser(t, d)

	route, err := d.CreateAdminRoute(ctx, ac, db.CreateAdminRouteParams{
		Slug:         "admin-broken",
		Title:        "Admin Broken",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateAdminRoute: %v", err)
	}

	dt, err := d.CreateAdminDatatype(ctx, ac, db.CreateAdminDatatypeParams{
		Name:         "admin-ref",
		Label:        "Admin Ref",
		Type:         "_reference",
		AuthorID:     userID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateAdminDatatype: %v", err)
	}

	_, err = d.CreateAdminContentData(ctx, ac, db.CreateAdminContentDataParams{
		AdminRouteID:    types.NullableAdminRouteID{ID: route.AdminRouteID, Valid: true},
		AdminDatatypeID: types.NullableAdminDatatypeID{ID: dt.AdminDatatypeID, Valid: true},
		AuthorID:        userID,
		Status:          types.ContentStatusDraft,
		DateCreated:     types.TimestampNow(),
		DateModified:    types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("CreateAdminContentData: %v", err)
	}

	report, err := svc.Heal(ctx, ac, true)
	if err != nil {
		t.Fatalf("Admin Heal dry run: %v", err)
	}

	if len(report.RootlessContent) != 1 {
		t.Fatalf("expected 1 admin rootless content, got %d", len(report.RootlessContent))
	}
	if report.RootlessContent[0].Deleted {
		t.Error("dry run should not delete")
	}
}
