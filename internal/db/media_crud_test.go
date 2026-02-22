// Integration tests for the media entity CRUD lifecycle.
// Uses testSeededDB (Tier 1b: requires user for author_id FK).
package db

import (
	"database/sql"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Media(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// --- Count: starts at 0 (no seed media) ---
	count, err := d.CountMedia()
	if err != nil {
		t.Fatalf("initial CountMedia: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountMedia = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         NullString{sql.NullString{String: "test-image", Valid: true}},
		DisplayName:  NullString{sql.NullString{String: "Test Image", Valid: true}},
		Alt:          NullString{sql.NullString{String: "A test image", Valid: true}},
		Caption:      NullString{sql.NullString{String: "Test caption", Valid: true}},
		Description:  NullString{sql.NullString{String: "Test description", Valid: true}},
		Class:        NullString{sql.NullString{String: "hero", Valid: true}},
		Mimetype:     NullString{sql.NullString{String: "image/jpeg", Valid: true}},
		Dimensions:   NullString{sql.NullString{String: "1920x1080", Valid: true}},
		URL:          types.URL("https://example.com/test.jpg"),
		Srcset:       NullString{sql.NullString{String: "test-sm.jpg 480w, test-lg.jpg 1920w", Valid: true}},
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateMedia: %v", err)
	}
	if created == nil {
		t.Fatal("CreateMedia returned nil")
	}
	if created.MediaID.IsZero() {
		t.Fatal("CreateMedia returned zero MediaID")
	}
	if !created.Name.Valid || created.Name.String != "test-image" {
		t.Errorf("Name = %v, want {test-image, true}", created.Name)
	}
	if !created.DisplayName.Valid || created.DisplayName.String != "Test Image" {
		t.Errorf("DisplayName = %v, want {Test Image, true}", created.DisplayName)
	}
	if created.URL != types.URL("https://example.com/test.jpg") {
		t.Errorf("URL = %q, want %q", created.URL, "https://example.com/test.jpg")
	}

	// --- Get ---
	got, err := d.GetMedia(created.MediaID)
	if err != nil {
		t.Fatalf("GetMedia: %v", err)
	}
	if got == nil {
		t.Fatal("GetMedia returned nil")
	}
	if got.MediaID != created.MediaID {
		t.Errorf("GetMedia ID = %v, want %v", got.MediaID, created.MediaID)
	}
	if !got.Name.Valid || got.Name.String != "test-image" {
		t.Errorf("GetMedia Name = %v, want {test-image, true}", got.Name)
	}
	if got.URL != created.URL {
		t.Errorf("GetMedia URL = %q, want %q", got.URL, created.URL)
	}
	if !got.Alt.Valid || got.Alt.String != "A test image" {
		t.Errorf("GetMedia Alt = %v, want {A test image, true}", got.Alt)
	}
	if !got.Mimetype.Valid || got.Mimetype.String != "image/jpeg" {
		t.Errorf("GetMedia Mimetype = %v, want {image/jpeg, true}", got.Mimetype)
	}

	// --- List ---
	list, err := d.ListMedia()
	if err != nil {
		t.Fatalf("ListMedia: %v", err)
	}
	if list == nil {
		t.Fatal("ListMedia returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListMedia len = %d, want 1", len(*list))
	}

	// --- Count: now 1 ---
	count, err = d.CountMedia()
	if err != nil {
		t.Fatalf("CountMedia after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountMedia after create = %d, want 1", *count)
	}

	// --- Update ---
	updatedNow := types.TimestampNow()
	updateResult, err := d.UpdateMedia(ctx, ac, UpdateMediaParams{
		Name:         NullString{sql.NullString{String: "test-image-updated", Valid: true}},
		DisplayName:  NullString{sql.NullString{String: "Updated Test Image", Valid: true}},
		Alt:          NullString{sql.NullString{String: "An updated test image", Valid: true}},
		Caption:      NullString{sql.NullString{String: "Updated caption", Valid: true}},
		Description:  NullString{sql.NullString{String: "Updated description", Valid: true}},
		Class:        NullString{sql.NullString{String: "banner", Valid: true}},
		Mimetype:     NullString{sql.NullString{String: "image/png", Valid: true}},
		Dimensions:   NullString{sql.NullString{String: "3840x2160", Valid: true}},
		URL:          types.URL("https://example.com/test-updated.png"),
		Srcset:       NullString{sql.NullString{String: "test-updated-sm.png 480w", Valid: true}},
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: updatedNow,
		MediaID:      created.MediaID,
	})
	if err != nil {
		t.Fatalf("UpdateMedia: %v", err)
	}
	// UpdateMedia returns a success message on success
	if updateResult == nil {
		t.Error("UpdateMedia returned nil message, expected success message")
	}

	// --- Get after update ---
	updated, err := d.GetMedia(created.MediaID)
	if err != nil {
		t.Fatalf("GetMedia after update: %v", err)
	}
	if !updated.Name.Valid || updated.Name.String != "test-image-updated" {
		t.Errorf("updated Name = %v, want {test-image-updated, true}", updated.Name)
	}
	if !updated.DisplayName.Valid || updated.DisplayName.String != "Updated Test Image" {
		t.Errorf("updated DisplayName = %v, want {Updated Test Image, true}", updated.DisplayName)
	}
	if updated.URL != types.URL("https://example.com/test-updated.png") {
		t.Errorf("updated URL = %q, want %q", updated.URL, "https://example.com/test-updated.png")
	}
	if !updated.Mimetype.Valid || updated.Mimetype.String != "image/png" {
		t.Errorf("updated Mimetype = %v, want {image/png, true}", updated.Mimetype)
	}

	// --- Delete ---
	err = d.DeleteMedia(ctx, ac, created.MediaID)
	if err != nil {
		t.Fatalf("DeleteMedia: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetMedia(created.MediaID)
	if err == nil {
		t.Fatal("GetMedia after delete: expected error, got nil")
	}

	// --- Count: back to 0 ---
	count, err = d.CountMedia()
	if err != nil {
		t.Fatalf("CountMedia after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountMedia after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_Media_GetMediaByName tests fetching media by name.
func TestDatabase_CRUD_Media_GetMediaByName(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	_, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         NullString{sql.NullString{String: "lookup-by-name", Valid: true}},
		DisplayName:  NullString{},
		Alt:          NullString{},
		Caption:      NullString{},
		Description:  NullString{},
		Class:        NullString{},
		Mimetype:     NullString{},
		Dimensions:   NullString{},
		URL:          types.URL("https://example.com/lookup.jpg"),
		Srcset:       NullString{},
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateMedia: %v", err)
	}

	got, err := d.GetMediaByName("lookup-by-name")
	if err != nil {
		t.Fatalf("GetMediaByName: %v", err)
	}
	if got == nil {
		t.Fatal("GetMediaByName returned nil")
	}
	if !got.Name.Valid || got.Name.String != "lookup-by-name" {
		t.Errorf("Name = %v, want {lookup-by-name, true}", got.Name)
	}
	if got.URL != types.URL("https://example.com/lookup.jpg") {
		t.Errorf("URL = %q, want %q", got.URL, "https://example.com/lookup.jpg")
	}
}

// TestDatabase_CRUD_Media_GetMediaByURL tests fetching media by URL.
func TestDatabase_CRUD_Media_GetMediaByURL(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	targetURL := types.URL("https://example.com/by-url-test.png")
	_, err := d.CreateMedia(ctx, ac, CreateMediaParams{
		Name:         NullString{sql.NullString{String: "url-lookup-media", Valid: true}},
		DisplayName:  NullString{},
		Alt:          NullString{},
		Caption:      NullString{},
		Description:  NullString{},
		Class:        NullString{},
		Mimetype:     NullString{},
		Dimensions:   NullString{},
		URL:          targetURL,
		Srcset:       NullString{},
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateMedia: %v", err)
	}

	got, err := d.GetMediaByURL(targetURL)
	if err != nil {
		t.Fatalf("GetMediaByURL: %v", err)
	}
	if got == nil {
		t.Fatal("GetMediaByURL returned nil")
	}
	if got.URL != targetURL {
		t.Errorf("URL = %q, want %q", got.URL, targetURL)
	}
	if !got.Name.Valid || got.Name.String != "url-lookup-media" {
		t.Errorf("Name = %v, want {url-lookup-media, true}", got.Name)
	}
}
