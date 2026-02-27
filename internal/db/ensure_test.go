package db

import (
	"context"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestEnsureSystemData_CreatesReferenceDatatype(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := context.Background()

	// Seed field_types so ensureFieldType doesn't fail looking up content_tree_ref.
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	_, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{Type: "content_tree_ref", Label: "Content Tree Reference"})
	if err != nil {
		t.Fatalf("seed content_tree_ref field_type: %v", err)
	}
	_, err = d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{Type: "content_tree_ref", Label: "Content Tree Reference"})
	if err != nil {
		t.Fatalf("seed content_tree_ref admin_field_type: %v", err)
	}

	// _reference should not exist yet
	_, err = d.GetDatatypeByType(string(types.DatatypeTypeReference))
	if err == nil {
		t.Fatal("expected _reference to not exist before EnsureSystemData")
	}

	// Run EnsureSystemData
	if err := EnsureSystemData(ctx, d); err != nil {
		t.Fatalf("EnsureSystemData: %v", err)
	}

	// Verify _reference datatype exists
	ref, err := d.GetDatatypeByType(string(types.DatatypeTypeReference))
	if err != nil {
		t.Fatalf("GetDatatypeByType(_reference) after ensure: %v", err)
	}
	if ref.Label != "Reference" {
		t.Errorf("label = %q, want %q", ref.Label, "Reference")
	}
	if ref.Type != string(types.DatatypeTypeReference) {
		t.Errorf("type = %q, want %q", ref.Type, string(types.DatatypeTypeReference))
	}

	// Verify Target field is linked
	links, err := d.ListDatatypeFieldByDatatypeID(ref.DatatypeID)
	if err != nil {
		t.Fatalf("ListDatatypeFieldByDatatypeID: %v", err)
	}
	if links == nil || len(*links) != 1 {
		t.Fatalf("expected 1 linked field, got %d", func() int {
			if links == nil {
				return 0
			}
			return len(*links)
		}())
	}

	// Verify the linked field is content_tree_ref type
	linkedField, err := d.GetField((*links)[0].FieldID)
	if err != nil {
		t.Fatalf("GetField for linked field: %v", err)
	}
	if linkedField.Type != types.FieldTypeContentTreeRef {
		t.Errorf("linked field type = %q, want %q", linkedField.Type, types.FieldTypeContentTreeRef)
	}
	if linkedField.Label != "Target" {
		t.Errorf("linked field label = %q, want %q", linkedField.Label, "Target")
	}
}

func TestEnsureSystemData_Idempotent(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := context.Background()

	// Seed field_types
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	_, err := d.CreateFieldType(ctx, ac, CreateFieldTypeParams{Type: "content_tree_ref", Label: "Content Tree Reference"})
	if err != nil {
		t.Fatalf("seed content_tree_ref field_type: %v", err)
	}
	_, err = d.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{Type: "content_tree_ref", Label: "Content Tree Reference"})
	if err != nil {
		t.Fatalf("seed content_tree_ref admin_field_type: %v", err)
	}

	// Run twice
	if err := EnsureSystemData(ctx, d); err != nil {
		t.Fatalf("EnsureSystemData (first call): %v", err)
	}
	if err := EnsureSystemData(ctx, d); err != nil {
		t.Fatalf("EnsureSystemData (second call): %v", err)
	}

	// Should still have exactly one _reference datatype
	all, err := d.ListDatatypes()
	if err != nil {
		t.Fatalf("ListDatatypes: %v", err)
	}
	refCount := 0
	for _, dt := range *all {
		if dt.Type == string(types.DatatypeTypeReference) {
			refCount++
		}
	}
	if refCount != 1 {
		t.Errorf("expected 1 _reference datatype after double ensure, got %d", refCount)
	}
}

func TestEnsureSystemData_CreatesFieldTypeIfMissing(t *testing.T) {
	t.Parallel()
	d, _ := testSeededDB(t)
	ctx := context.Background()

	// Do NOT seed content_tree_ref — EnsureSystemData should create it
	if err := EnsureSystemData(ctx, d); err != nil {
		t.Fatalf("EnsureSystemData: %v", err)
	}

	// Verify field_type exists
	ft, err := d.GetFieldTypeByType("content_tree_ref")
	if err != nil {
		t.Fatalf("GetFieldTypeByType(content_tree_ref) after ensure: %v", err)
	}
	if ft.Type != "content_tree_ref" {
		t.Errorf("field_type.Type = %q, want %q", ft.Type, "content_tree_ref")
	}

	// Verify admin_field_type exists
	aft, err := d.GetAdminFieldTypeByType("content_tree_ref")
	if err != nil {
		t.Fatalf("GetAdminFieldTypeByType(content_tree_ref) after ensure: %v", err)
	}
	if aft.Type != "content_tree_ref" {
		t.Errorf("admin_field_type.Type = %q, want %q", aft.Type, "content_tree_ref")
	}
}
