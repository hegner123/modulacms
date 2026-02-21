// Integration tests for the datatype entity CRUD lifecycle.
// Uses testSeededDB (Tier 1b: requires user for author_id FK).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Datatype(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := seed.User.UserID

	// --- Count: starts at 1 (seed datatype) ---
	count, err := d.CountDatatypes()
	if err != nil {
		t.Fatalf("initial CountDatatypes: %v", err)
	}
	if *count != 1 {
		t.Fatalf("initial CountDatatypes = %d, want 1", *count)
	}

	// --- Create ---
	created, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "crud-test-datatype",
		Type:         "article",
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateDatatype: %v", err)
	}
	if created == nil {
		t.Fatal("CreateDatatype returned nil")
	}
	if created.DatatypeID.IsZero() {
		t.Fatal("CreateDatatype returned zero DatatypeID")
	}
	if created.Label != "crud-test-datatype" {
		t.Errorf("Label = %q, want %q", created.Label, "crud-test-datatype")
	}
	if created.Type != "article" {
		t.Errorf("Type = %q, want %q", created.Type, "article")
	}

	// --- Get ---
	got, err := d.GetDatatype(created.DatatypeID)
	if err != nil {
		t.Fatalf("GetDatatype: %v", err)
	}
	if got == nil {
		t.Fatal("GetDatatype returned nil")
	}
	if got.DatatypeID != created.DatatypeID {
		t.Errorf("GetDatatype ID = %v, want %v", got.DatatypeID, created.DatatypeID)
	}
	if got.Label != created.Label {
		t.Errorf("GetDatatype Label = %q, want %q", got.Label, created.Label)
	}
	if got.Type != created.Type {
		t.Errorf("GetDatatype Type = %q, want %q", got.Type, created.Type)
	}

	// --- List ---
	list, err := d.ListDatatypes()
	if err != nil {
		t.Fatalf("ListDatatypes: %v", err)
	}
	if list == nil {
		t.Fatal("ListDatatypes returned nil")
	}
	if len(*list) != 2 {
		t.Fatalf("ListDatatypes len = %d, want 2", len(*list))
	}

	// --- Count: now 2 ---
	count, err = d.CountDatatypes()
	if err != nil {
		t.Fatalf("CountDatatypes after create: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountDatatypes after create = %d, want 2", *count)
	}

	// --- Update ---
	updatedNow := types.TimestampNow()
	updateResult, err := d.UpdateDatatype(ctx, ac, UpdateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		Label:        "crud-test-datatype-updated",
		Type:         "blog",
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: updatedNow,
		DatatypeID:   created.DatatypeID,
	})
	if err != nil {
		t.Fatalf("UpdateDatatype: %v", err)
	}
	// UpdateDatatype returns a success message on success
	if updateResult == nil {
		t.Error("UpdateDatatype returned nil message, expected success message")
	}

	// --- Get after update ---
	updated, err := d.GetDatatype(created.DatatypeID)
	if err != nil {
		t.Fatalf("GetDatatype after update: %v", err)
	}
	if updated.Label != "crud-test-datatype-updated" {
		t.Errorf("updated Label = %q, want %q", updated.Label, "crud-test-datatype-updated")
	}
	if updated.Type != "blog" {
		t.Errorf("updated Type = %q, want %q", updated.Type, "blog")
	}

	// --- Delete (only our created datatype, not the seed) ---
	err = d.DeleteDatatype(ctx, ac, created.DatatypeID)
	if err != nil {
		t.Fatalf("DeleteDatatype: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetDatatype(created.DatatypeID)
	if err == nil {
		t.Fatal("GetDatatype after delete: expected error, got nil")
	}

	// --- Count: back to 1 (seed datatype remains) ---
	count, err = d.CountDatatypes()
	if err != nil {
		t.Fatalf("CountDatatypes after delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountDatatypes after delete = %d, want 1", *count)
	}
}

// TestDatabase_CRUD_Datatype_ListDatatypesRoot tests that root datatypes
// (Type = 'ROOT') are returned correctly.
func TestDatabase_CRUD_Datatype_ListDatatypesRoot(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := seed.User.UserID

	// Seed datatype has Type="page", so it does NOT appear in root list.
	// ListDatatypesRoot filters by WHERE type = 'ROOT'.
	roots, err := d.ListDatatypesRoot()
	if err != nil {
		t.Fatalf("ListDatatypesRoot: %v", err)
	}
	if roots == nil {
		t.Fatal("ListDatatypesRoot returned nil")
	}
	if len(*roots) != 0 {
		t.Fatalf("ListDatatypesRoot len = %d, want 0 (no ROOT type exists)", len(*roots))
	}

	// Create a ROOT type datatype
	rootDT, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "root-datatype",
		Type:         "ROOT",
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateDatatype (ROOT): %v", err)
	}

	roots, err = d.ListDatatypesRoot()
	if err != nil {
		t.Fatalf("ListDatatypesRoot after create: %v", err)
	}
	if len(*roots) != 1 {
		t.Fatalf("ListDatatypesRoot len = %d, want 1", len(*roots))
	}
	if (*roots)[0].DatatypeID != rootDT.DatatypeID {
		t.Errorf("root DatatypeID = %v, want %v", (*roots)[0].DatatypeID, rootDT.DatatypeID)
	}
}

// TestDatabase_CRUD_Datatype_ListDatatypeChildren tests listing child datatypes.
func TestDatabase_CRUD_Datatype_ListDatatypeChildren(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := seed.User.UserID

	// Create a child datatype under the seed datatype
	child, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{
		DatatypeID: types.NewDatatypeID(),
		ParentID: types.NullableDatatypeID{
			ID:    seed.Datatype.DatatypeID,
			Valid: true,
		},
		Label:        "child-datatype",
		Type:         "section",
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateDatatype (child): %v", err)
	}

	// List children of the seed datatype
	children, err := d.ListDatatypeChildren(seed.Datatype.DatatypeID)
	if err != nil {
		t.Fatalf("ListDatatypeChildren: %v", err)
	}
	if children == nil {
		t.Fatal("ListDatatypeChildren returned nil")
	}
	if len(*children) != 1 {
		t.Fatalf("ListDatatypeChildren len = %d, want 1", len(*children))
	}
	if (*children)[0].DatatypeID != child.DatatypeID {
		t.Errorf("child DatatypeID = %v, want %v", (*children)[0].DatatypeID, child.DatatypeID)
	}
	if (*children)[0].Label != "child-datatype" {
		t.Errorf("child Label = %q, want %q", (*children)[0].Label, "child-datatype")
	}

	// List children of a datatype that has no children
	noChildren, err := d.ListDatatypeChildren(child.DatatypeID)
	if err != nil {
		t.Fatalf("ListDatatypeChildren (no children): %v", err)
	}
	if noChildren == nil {
		t.Fatal("ListDatatypeChildren (no children) returned nil")
	}
	if len(*noChildren) != 0 {
		t.Fatalf("ListDatatypeChildren (no children) len = %d, want 0", len(*noChildren))
	}
}
