// Integration tests for the media_dimension entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0: no FK dependencies).
package db

import (
	"testing"
)

func TestDatabase_CRUD_MediaDimension(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// --- Count: starts at zero ---
	count, err := d.CountMediaDimensions()
	if err != nil {
		t.Fatalf("initial CountMediaDimensions: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountMediaDimensions = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateMediaDimension(ctx, ac, CreateMediaDimensionParams{
		Label:       NewNullString("Thumbnail"),
		Width:       NewNullInt64(1920),
		Height:      NewNullInt64(1080),
		AspectRatio: NewNullString("16:9"),
	})
	if err != nil {
		t.Fatalf("CreateMediaDimension: %v", err)
	}
	if created == nil {
		t.Fatal("CreateMediaDimension returned nil")
	}
	if created.MdID == "" {
		t.Fatal("CreateMediaDimension returned empty MdID")
	}
	if !created.Label.Valid || created.Label.String != "Thumbnail" {
		t.Errorf("Label = %v, want valid 'Thumbnail'", created.Label)
	}
	if !created.Width.Valid || created.Width.Int64 != 1920 {
		t.Errorf("Width = %v, want valid 1920", created.Width)
	}
	if !created.Height.Valid || created.Height.Int64 != 1080 {
		t.Errorf("Height = %v, want valid 1080", created.Height)
	}
	if !created.AspectRatio.Valid || created.AspectRatio.String != "16:9" {
		t.Errorf("AspectRatio = %v, want valid '16:9'", created.AspectRatio)
	}

	// --- Get ---
	got, err := d.GetMediaDimension(created.MdID)
	if err != nil {
		t.Fatalf("GetMediaDimension: %v", err)
	}
	if got == nil {
		t.Fatal("GetMediaDimension returned nil")
	}
	if got.MdID != created.MdID {
		t.Errorf("GetMediaDimension MdID = %q, want %q", got.MdID, created.MdID)
	}
	if got.Label != created.Label {
		t.Errorf("GetMediaDimension Label = %v, want %v", got.Label, created.Label)
	}
	if got.Width != created.Width {
		t.Errorf("GetMediaDimension Width = %v, want %v", got.Width, created.Width)
	}
	if got.Height != created.Height {
		t.Errorf("GetMediaDimension Height = %v, want %v", got.Height, created.Height)
	}

	// --- List ---
	list, err := d.ListMediaDimensions()
	if err != nil {
		t.Fatalf("ListMediaDimensions: %v", err)
	}
	if list == nil {
		t.Fatal("ListMediaDimensions returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListMediaDimensions len = %d, want 1", len(*list))
	}
	if (*list)[0].MdID != created.MdID {
		t.Errorf("ListMediaDimensions[0].MdID = %q, want %q", (*list)[0].MdID, created.MdID)
	}

	// --- Count: now 1 ---
	count, err = d.CountMediaDimensions()
	if err != nil {
		t.Fatalf("CountMediaDimensions after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountMediaDimensions after create = %d, want 1", *count)
	}

	// --- Update ---
	_, err = d.UpdateMediaDimension(ctx, ac, UpdateMediaDimensionParams{
		Label:       NewNullString("Banner"),
		Width:       NewNullInt64(2560),
		Height:      NewNullInt64(1440),
		AspectRatio: NewNullString("16:9"),
		MdID:        created.MdID,
	})
	if err != nil {
		t.Fatalf("UpdateMediaDimension: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetMediaDimension(created.MdID)
	if err != nil {
		t.Fatalf("GetMediaDimension after update: %v", err)
	}
	if !updated.Label.Valid || updated.Label.String != "Banner" {
		t.Errorf("updated Label = %v, want valid 'Banner'", updated.Label)
	}
	if !updated.Width.Valid || updated.Width.Int64 != 2560 {
		t.Errorf("updated Width = %v, want valid 2560", updated.Width)
	}
	if !updated.Height.Valid || updated.Height.Int64 != 1440 {
		t.Errorf("updated Height = %v, want valid 1440", updated.Height)
	}

	// --- Delete ---
	err = d.DeleteMediaDimension(ctx, ac, created.MdID)
	if err != nil {
		t.Fatalf("DeleteMediaDimension: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetMediaDimension(created.MdID)
	if err == nil {
		t.Fatal("GetMediaDimension after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountMediaDimensions()
	if err != nil {
		t.Fatalf("CountMediaDimensions after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountMediaDimensions after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_MediaDimension_NullFields verifies that creating
// a media dimension with all-null optional fields works correctly.
func TestDatabase_CRUD_MediaDimension_NullFields(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	created, err := d.CreateMediaDimension(ctx, ac, CreateMediaDimensionParams{
		Label:       NullString{},
		Width:       NullInt64{},
		Height:      NullInt64{},
		AspectRatio: NullString{},
	})
	if err != nil {
		t.Fatalf("CreateMediaDimension with nulls: %v", err)
	}
	if created == nil {
		t.Fatal("CreateMediaDimension returned nil")
	}
	if created.MdID == "" {
		t.Fatal("CreateMediaDimension returned empty MdID")
	}

	got, err := d.GetMediaDimension(created.MdID)
	if err != nil {
		t.Fatalf("GetMediaDimension: %v", err)
	}
	if got.Label.Valid {
		t.Errorf("Label.Valid = true, want false for null field")
	}
	if got.Width.Valid {
		t.Errorf("Width.Valid = true, want false for null field")
	}
	if got.Height.Valid {
		t.Errorf("Height.Valid = true, want false for null field")
	}
	if got.AspectRatio.Valid {
		t.Errorf("AspectRatio.Valid = true, want false for null field")
	}
}
