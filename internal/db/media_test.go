package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// mediaTestFixture returns a fully populated Media struct and its constituent parts.
func mediaTestFixture() (Media, types.MediaID, types.NullableUserID, types.Timestamp) {
	mediaID := types.NewMediaID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 5, 20, 14, 30, 0, 0, time.UTC))
	m := Media{
		MediaID:      mediaID,
		Name:         sql.NullString{String: "hero-banner.jpg", Valid: true},
		DisplayName:  sql.NullString{String: "Hero Banner", Valid: true},
		Alt:          sql.NullString{String: "A wide hero banner image", Valid: true},
		Caption:      sql.NullString{String: "Site header image", Valid: true},
		Description:  sql.NullString{String: "Full-width hero banner for landing page", Valid: true},
		Class:        sql.NullString{String: "image", Valid: true},
		Mimetype:     sql.NullString{String: "image/jpeg", Valid: true},
		Dimensions:   sql.NullString{String: "1920x1080", Valid: true},
		URL:          types.URL("https://cdn.example.com/images/hero-banner.jpg"),
		Srcset:       sql.NullString{String: "hero-480.jpg 480w, hero-1080.jpg 1080w", Valid: true},
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}
	return m, mediaID, authorID, ts
}

// mediaCreateParams returns a CreateMediaParams with all fields populated.
func mediaCreateParams() CreateMediaParams {
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC))
	return CreateMediaParams{
		Name:         sql.NullString{String: "document.pdf", Valid: true},
		DisplayName:  sql.NullString{String: "Annual Report", Valid: true},
		Alt:          sql.NullString{String: "PDF document", Valid: true},
		Caption:      sql.NullString{String: "Annual Report 2025", Valid: true},
		Description:  sql.NullString{String: "Company annual financial report", Valid: true},
		Class:        sql.NullString{String: "document", Valid: true},
		Mimetype:     sql.NullString{String: "application/pdf", Valid: true},
		Dimensions:   sql.NullString{Valid: false},
		URL:          types.URL("https://cdn.example.com/docs/annual-report.pdf"),
		Srcset:       sql.NullString{Valid: false},
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}
}

// mediaUpdateParams returns an UpdateMediaParams with all fields populated.
func mediaUpdateParams() UpdateMediaParams {
	mediaID := types.NewMediaID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 7, 15, 8, 45, 0, 0, time.UTC))
	return UpdateMediaParams{
		Name:         sql.NullString{String: "updated-photo.png", Valid: true},
		DisplayName:  sql.NullString{String: "Updated Photo", Valid: true},
		Alt:          sql.NullString{String: "Updated alt text", Valid: true},
		Caption:      sql.NullString{String: "Updated caption", Valid: true},
		Description:  sql.NullString{String: "Updated description", Valid: true},
		Class:        sql.NullString{String: "image", Valid: true},
		Mimetype:     sql.NullString{String: "image/png", Valid: true},
		Dimensions:   sql.NullString{String: "640x480", Valid: true},
		URL:          types.URL("https://cdn.example.com/images/updated-photo.png"),
		Srcset:       sql.NullString{String: "updated-480.png 480w", Valid: true},
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		MediaID:      mediaID,
	}
}

// --- MapStringMedia tests ---

func TestMapStringMedia_AllFields(t *testing.T) {
	t.Parallel()
	m, mediaID, authorID, ts := mediaTestFixture()

	got := MapStringMedia(m)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"MediaID", got.MediaID, mediaID.String()},
		{"Name", got.Name, "hero-banner.jpg"},
		{"DisplayName", got.DisplayName, "Hero Banner"},
		{"Alt", got.Alt, "A wide hero banner image"},
		{"Caption", got.Caption, "Site header image"},
		{"Description", got.Description, "Full-width hero banner for landing page"},
		{"Class", got.Class, "image"},
		{"Mimetype", got.Mimetype, "image/jpeg"},
		{"Dimensions", got.Dimensions, "1920x1080"},
		{"Url", got.Url, "https://cdn.example.com/images/hero-banner.jpg"},
		{"Srcset", got.Srcset, "hero-480.jpg 480w, hero-1080.jpg 1080w"},
		{"AuthorID", got.AuthorID, authorID.String()},
		{"DateCreated", got.DateCreated, ts.String()},
		{"DateModified", got.DateModified, ts.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestMapStringMedia_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value Media: NullToString returns "null" for zero-value sql.NullStrings
	// (Valid=false), non-NullString fields return their zero string representation.
	got := MapStringMedia(Media{})

	if got.MediaID != "" {
		t.Errorf("MediaID = %q, want empty string", got.MediaID)
	}
	// Zero-value sql.NullString has Valid=false, so NullToString returns "null"
	nullStringFields := []struct {
		name string
		got  string
	}{
		{"Name", got.Name},
		{"DisplayName", got.DisplayName},
		{"Alt", got.Alt},
		{"Caption", got.Caption},
		{"Description", got.Description},
		{"Class", got.Class},
		{"Mimetype", got.Mimetype},
		{"Dimensions", got.Dimensions},
		{"Srcset", got.Srcset},
	}
	for _, f := range nullStringFields {
		if f.got != "null" {
			t.Errorf("%s = %q, want %q for zero-value NullString", f.name, f.got, "null")
		}
	}
	// URL and types that use .String() directly return their zero values
	if got.Url != "" {
		t.Errorf("Url = %q, want empty string", got.Url)
	}
}

func TestMapStringMedia_NullFields(t *testing.T) {
	t.Parallel()
	// Media with all NullString fields explicitly invalid (Valid=false).
	// NullToString returns the literal string "null" for invalid NullStrings,
	// so the internal String value ("should-not-appear") must NOT leak through.
	m := Media{
		MediaID:     types.NewMediaID(),
		Name:        sql.NullString{String: "should-not-appear", Valid: false},
		DisplayName: sql.NullString{String: "should-not-appear", Valid: false},
		Alt:         sql.NullString{String: "should-not-appear", Valid: false},
		Caption:     sql.NullString{String: "should-not-appear", Valid: false},
		Description: sql.NullString{String: "should-not-appear", Valid: false},
		Class:       sql.NullString{String: "should-not-appear", Valid: false},
		Mimetype:    sql.NullString{String: "should-not-appear", Valid: false},
		Dimensions:  sql.NullString{String: "should-not-appear", Valid: false},
		Srcset:      sql.NullString{String: "should-not-appear", Valid: false},
		AuthorID:    types.NullableUserID{Valid: false},
	}

	got := MapStringMedia(m)

	// NullToString returns "null" for invalid sql.NullString values.
	// The key invariant: the underlying String value must not appear.
	nullFields := []struct {
		name string
		got  string
	}{
		{"Name", got.Name},
		{"DisplayName", got.DisplayName},
		{"Alt", got.Alt},
		{"Caption", got.Caption},
		{"Description", got.Description},
		{"Class", got.Class},
		{"Mimetype", got.Mimetype},
		{"Dimensions", got.Dimensions},
		{"Srcset", got.Srcset},
	}

	for _, f := range nullFields {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			if f.got != "null" {
				t.Errorf("%s = %q, want %q for invalid NullString", f.name, f.got, "null")
			}
			if f.got == "should-not-appear" {
				t.Errorf("%s leaked the underlying String value through an invalid NullString", f.name)
			}
		})
	}
}

func TestMapStringMedia_NullAuthorID(t *testing.T) {
	t.Parallel()
	// AuthorID with Valid=false should convert without panic
	m := Media{
		AuthorID: types.NullableUserID{Valid: false},
	}
	got := MapStringMedia(m)
	// Should produce whatever NullableUserID.String() returns for invalid
	if got.AuthorID != m.AuthorID.String() {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, m.AuthorID.String())
	}
}

// --- SQLite Database.MapMedia tests ---

func TestDatabase_MapMedia_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	mediaID := types.NewMediaID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC))

	input := mdb.Media{
		MediaID:      mediaID,
		Name:         sql.NullString{String: "photo.jpg", Valid: true},
		DisplayName:  sql.NullString{String: "My Photo", Valid: true},
		Alt:          sql.NullString{String: "A photo", Valid: true},
		Caption:      sql.NullString{String: "Caption text", Valid: true},
		Description:  sql.NullString{String: "Description text", Valid: true},
		Class:        sql.NullString{String: "image", Valid: true},
		Mimetype:     sql.NullString{String: "image/jpeg", Valid: true},
		Dimensions:   sql.NullString{String: "800x600", Valid: true},
		URL:          types.URL("https://example.com/photo.jpg"),
		Srcset:       sql.NullString{String: "photo-480.jpg 480w", Valid: true},
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
	if got.DisplayName.String != "My Photo" || !got.DisplayName.Valid {
		t.Errorf("DisplayName = %v, want valid 'My Photo'", got.DisplayName)
	}
	if got.Alt.String != "A photo" || !got.Alt.Valid {
		t.Errorf("Alt = %v, want valid 'A photo'", got.Alt)
	}
	if got.Caption.String != "Caption text" || !got.Caption.Valid {
		t.Errorf("Caption = %v, want valid 'Caption text'", got.Caption)
	}
	if got.Description.String != "Description text" || !got.Description.Valid {
		t.Errorf("Description = %v, want valid 'Description text'", got.Description)
	}
	if got.Class.String != "image" || !got.Class.Valid {
		t.Errorf("Class = %v, want valid 'image'", got.Class)
	}
	if got.Mimetype.String != "image/jpeg" || !got.Mimetype.Valid {
		t.Errorf("Mimetype = %v, want valid 'image/jpeg'", got.Mimetype)
	}
	if got.Dimensions.String != "800x600" || !got.Dimensions.Valid {
		t.Errorf("Dimensions = %v, want valid '800x600'", got.Dimensions)
	}
	if string(got.URL) != "https://example.com/photo.jpg" {
		t.Errorf("URL = %v, want %v", got.URL, "https://example.com/photo.jpg")
	}
	if got.Srcset.String != "photo-480.jpg 480w" || !got.Srcset.Valid {
		t.Errorf("Srcset = %v, want valid 'photo-480.jpg 480w'", got.Srcset)
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

func TestDatabase_MapMedia_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapMedia(mdb.Media{})

	if got.MediaID != "" {
		t.Errorf("MediaID = %v, want zero value", got.MediaID)
	}
	if got.Name.Valid {
		t.Errorf("Name.Valid = true, want false for zero value")
	}
	if got.URL != "" {
		t.Errorf("URL = %v, want zero value", got.URL)
	}
}

func TestDatabase_MapMedia_NullSrcset(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.Media{
		MediaID: types.NewMediaID(),
		Srcset:  sql.NullString{Valid: false},
	}

	got := d.MapMedia(input)

	if got.Srcset.Valid {
		t.Error("Srcset should be invalid/null")
	}
}

// --- SQLite Database.MapCreateMediaParams tests ---

func TestDatabase_MapCreateMediaParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := mediaCreateParams()

	got := d.MapCreateMediaParams(params)

	// MapCreateMediaParams always generates a new MediaID
	if got.MediaID.IsZero() {
		t.Fatal("expected non-zero MediaID to be generated")
	}
	if got.Name != params.Name {
		t.Errorf("Name = %v, want %v", got.Name, params.Name)
	}
	if got.DisplayName != params.DisplayName {
		t.Errorf("DisplayName = %v, want %v", got.DisplayName, params.DisplayName)
	}
	if got.Alt != params.Alt {
		t.Errorf("Alt = %v, want %v", got.Alt, params.Alt)
	}
	if got.Caption != params.Caption {
		t.Errorf("Caption = %v, want %v", got.Caption, params.Caption)
	}
	if got.Description != params.Description {
		t.Errorf("Description = %v, want %v", got.Description, params.Description)
	}
	if got.Class != params.Class {
		t.Errorf("Class = %v, want %v", got.Class, params.Class)
	}
	if got.Mimetype != params.Mimetype {
		t.Errorf("Mimetype = %v, want %v", got.Mimetype, params.Mimetype)
	}
	if got.Dimensions != params.Dimensions {
		t.Errorf("Dimensions = %v, want %v", got.Dimensions, params.Dimensions)
	}
	if got.URL != params.URL {
		t.Errorf("URL = %v, want %v", got.URL, params.URL)
	}
	if got.Srcset != params.Srcset {
		t.Errorf("Srcset = %v, want %v", got.Srcset, params.Srcset)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
	if got.DateCreated != params.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, params.DateCreated)
	}
	if got.DateModified != params.DateModified {
		t.Errorf("DateModified = %v, want %v", got.DateModified, params.DateModified)
	}
}

func TestDatabase_MapCreateMediaParams_AlwaysGeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := mediaCreateParams()

	// Call twice and verify unique IDs each time
	got1 := d.MapCreateMediaParams(params)
	got2 := d.MapCreateMediaParams(params)

	if got1.MediaID.IsZero() {
		t.Fatal("first call: expected non-zero MediaID")
	}
	if got2.MediaID.IsZero() {
		t.Fatal("second call: expected non-zero MediaID")
	}
	if got1.MediaID == got2.MediaID {
		t.Error("two calls produced identical MediaIDs; each should be unique")
	}
}

func TestDatabase_MapCreateMediaParams_ZeroInput(t *testing.T) {
	t.Parallel()
	d := Database{}
	// Even with zero CreateMediaParams, a new MediaID is generated
	got := d.MapCreateMediaParams(CreateMediaParams{})

	if got.MediaID.IsZero() {
		t.Fatal("expected non-zero MediaID even with zero-value input")
	}
}

// --- SQLite Database.MapUpdateMediaParams tests ---

func TestDatabase_MapUpdateMediaParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := mediaUpdateParams()

	got := d.MapUpdateMediaParams(params)

	if got.MediaID != params.MediaID {
		t.Errorf("MediaID = %v, want %v", got.MediaID, params.MediaID)
	}
	if got.Name != params.Name {
		t.Errorf("Name = %v, want %v", got.Name, params.Name)
	}
	if got.DisplayName != params.DisplayName {
		t.Errorf("DisplayName = %v, want %v", got.DisplayName, params.DisplayName)
	}
	if got.Alt != params.Alt {
		t.Errorf("Alt = %v, want %v", got.Alt, params.Alt)
	}
	if got.Caption != params.Caption {
		t.Errorf("Caption = %v, want %v", got.Caption, params.Caption)
	}
	if got.Description != params.Description {
		t.Errorf("Description = %v, want %v", got.Description, params.Description)
	}
	if got.Class != params.Class {
		t.Errorf("Class = %v, want %v", got.Class, params.Class)
	}
	if got.Mimetype != params.Mimetype {
		t.Errorf("Mimetype = %v, want %v", got.Mimetype, params.Mimetype)
	}
	if got.Dimensions != params.Dimensions {
		t.Errorf("Dimensions = %v, want %v", got.Dimensions, params.Dimensions)
	}
	if got.URL != params.URL {
		t.Errorf("URL = %v, want %v", got.URL, params.URL)
	}
	if got.Srcset != params.Srcset {
		t.Errorf("Srcset = %v, want %v", got.Srcset, params.Srcset)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
	if got.DateCreated != params.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, params.DateCreated)
	}
	if got.DateModified != params.DateModified {
		t.Errorf("DateModified = %v, want %v", got.DateModified, params.DateModified)
	}
}

func TestDatabase_MapUpdateMediaParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateMediaParams(UpdateMediaParams{})

	if got.MediaID != "" {
		t.Errorf("MediaID = %v, want zero value", got.MediaID)
	}
	if got.Name.Valid {
		t.Errorf("Name.Valid = true, want false for zero value")
	}
}

// --- MySQL MysqlDatabase.MapMedia tests ---

func TestMysqlDatabase_MapMedia_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	mediaID := types.NewMediaID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 5, 11, 0, 0, 0, time.UTC))

	input := mdbm.Media{
		MediaID:      mediaID,
		Name:         sql.NullString{String: "mysql-photo.jpg", Valid: true},
		DisplayName:  sql.NullString{String: "MySQL Photo", Valid: true},
		Alt:          sql.NullString{String: "MySQL alt", Valid: true},
		Caption:      sql.NullString{String: "MySQL caption", Valid: true},
		Description:  sql.NullString{String: "MySQL description", Valid: true},
		Class:        sql.NullString{String: "image", Valid: true},
		Mimetype:     sql.NullString{String: "image/jpeg", Valid: true},
		Dimensions:   sql.NullString{String: "1024x768", Valid: true},
		URL:          types.URL("https://mysql.example.com/photo.jpg"),
		Srcset:       sql.NullString{String: "mysql-480.jpg 480w", Valid: true},
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapMedia(input)

	if got.MediaID != mediaID {
		t.Errorf("MediaID = %v, want %v", got.MediaID, mediaID)
	}
	if got.Name.String != "mysql-photo.jpg" || !got.Name.Valid {
		t.Errorf("Name = %v, want valid 'mysql-photo.jpg'", got.Name)
	}
	if got.DisplayName.String != "MySQL Photo" || !got.DisplayName.Valid {
		t.Errorf("DisplayName = %v, want valid 'MySQL Photo'", got.DisplayName)
	}
	if string(got.URL) != "https://mysql.example.com/photo.jpg" {
		t.Errorf("URL = %v, want %v", got.URL, "https://mysql.example.com/photo.jpg")
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestMysqlDatabase_MapMedia_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapMedia(mdbm.Media{})

	if got.MediaID != "" {
		t.Errorf("MediaID = %v, want zero value", got.MediaID)
	}
	if got.URL != "" {
		t.Errorf("URL = %v, want zero value", got.URL)
	}
}

// --- MySQL MysqlDatabase.MapCreateMediaParams tests ---

func TestMysqlDatabase_MapCreateMediaParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := mediaCreateParams()

	got := d.MapCreateMediaParams(params)

	if got.MediaID.IsZero() {
		t.Fatal("expected non-zero MediaID to be generated")
	}
	if got.Name != params.Name {
		t.Errorf("Name = %v, want %v", got.Name, params.Name)
	}
	if got.DisplayName != params.DisplayName {
		t.Errorf("DisplayName = %v, want %v", got.DisplayName, params.DisplayName)
	}
	if got.URL != params.URL {
		t.Errorf("URL = %v, want %v", got.URL, params.URL)
	}
	if got.Mimetype != params.Mimetype {
		t.Errorf("Mimetype = %v, want %v", got.Mimetype, params.Mimetype)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
	if got.DateCreated != params.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, params.DateCreated)
	}
}

func TestMysqlDatabase_MapCreateMediaParams_AlwaysGeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := mediaCreateParams()

	got1 := d.MapCreateMediaParams(params)
	got2 := d.MapCreateMediaParams(params)

	if got1.MediaID == got2.MediaID {
		t.Error("two calls produced identical MediaIDs; each should be unique")
	}
}

// --- MySQL MysqlDatabase.MapUpdateMediaParams tests ---

func TestMysqlDatabase_MapUpdateMediaParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := mediaUpdateParams()

	got := d.MapUpdateMediaParams(params)

	if got.MediaID != params.MediaID {
		t.Errorf("MediaID = %v, want %v", got.MediaID, params.MediaID)
	}
	if got.Name != params.Name {
		t.Errorf("Name = %v, want %v", got.Name, params.Name)
	}
	if got.URL != params.URL {
		t.Errorf("URL = %v, want %v", got.URL, params.URL)
	}
	if got.Mimetype != params.Mimetype {
		t.Errorf("Mimetype = %v, want %v", got.Mimetype, params.Mimetype)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
}

// --- PostgreSQL PsqlDatabase.MapMedia tests ---

func TestPsqlDatabase_MapMedia_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	mediaID := types.NewMediaID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 28, 16, 0, 0, 0, time.UTC))

	input := mdbp.Media{
		MediaID:      mediaID,
		Name:         sql.NullString{String: "psql-photo.jpg", Valid: true},
		DisplayName:  sql.NullString{String: "PostgreSQL Photo", Valid: true},
		Alt:          sql.NullString{String: "Psql alt", Valid: true},
		Caption:      sql.NullString{String: "Psql caption", Valid: true},
		Description:  sql.NullString{String: "Psql description", Valid: true},
		Class:        sql.NullString{String: "image", Valid: true},
		Mimetype:     sql.NullString{String: "image/jpeg", Valid: true},
		Dimensions:   sql.NullString{String: "2560x1440", Valid: true},
		URL:          types.URL("https://psql.example.com/photo.jpg"),
		Srcset:       sql.NullString{String: "psql-480.jpg 480w, psql-1440.jpg 1440w", Valid: true},
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapMedia(input)

	if got.MediaID != mediaID {
		t.Errorf("MediaID = %v, want %v", got.MediaID, mediaID)
	}
	if got.Name.String != "psql-photo.jpg" || !got.Name.Valid {
		t.Errorf("Name = %v, want valid 'psql-photo.jpg'", got.Name)
	}
	if got.DisplayName.String != "PostgreSQL Photo" || !got.DisplayName.Valid {
		t.Errorf("DisplayName = %v, want valid 'PostgreSQL Photo'", got.DisplayName)
	}
	if string(got.URL) != "https://psql.example.com/photo.jpg" {
		t.Errorf("URL = %v, want %v", got.URL, "https://psql.example.com/photo.jpg")
	}
	if got.Srcset.String != "psql-480.jpg 480w, psql-1440.jpg 1440w" || !got.Srcset.Valid {
		t.Errorf("Srcset = %v, want valid srcset string", got.Srcset)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestPsqlDatabase_MapMedia_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapMedia(mdbp.Media{})

	if got.MediaID != "" {
		t.Errorf("MediaID = %v, want zero value", got.MediaID)
	}
	if got.URL != "" {
		t.Errorf("URL = %v, want zero value", got.URL)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateMediaParams tests ---

func TestPsqlDatabase_MapCreateMediaParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := mediaCreateParams()

	got := d.MapCreateMediaParams(params)

	if got.MediaID.IsZero() {
		t.Fatal("expected non-zero MediaID to be generated")
	}
	if got.Name != params.Name {
		t.Errorf("Name = %v, want %v", got.Name, params.Name)
	}
	if got.DisplayName != params.DisplayName {
		t.Errorf("DisplayName = %v, want %v", got.DisplayName, params.DisplayName)
	}
	if got.URL != params.URL {
		t.Errorf("URL = %v, want %v", got.URL, params.URL)
	}
	if got.Mimetype != params.Mimetype {
		t.Errorf("Mimetype = %v, want %v", got.Mimetype, params.Mimetype)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
	if got.DateCreated != params.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, params.DateCreated)
	}
}

func TestPsqlDatabase_MapCreateMediaParams_AlwaysGeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := mediaCreateParams()

	got1 := d.MapCreateMediaParams(params)
	got2 := d.MapCreateMediaParams(params)

	if got1.MediaID == got2.MediaID {
		t.Error("two calls produced identical MediaIDs; each should be unique")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateMediaParams tests ---

func TestPsqlDatabase_MapUpdateMediaParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := mediaUpdateParams()

	got := d.MapUpdateMediaParams(params)

	if got.MediaID != params.MediaID {
		t.Errorf("MediaID = %v, want %v", got.MediaID, params.MediaID)
	}
	if got.Name != params.Name {
		t.Errorf("Name = %v, want %v", got.Name, params.Name)
	}
	if got.URL != params.URL {
		t.Errorf("URL = %v, want %v", got.URL, params.URL)
	}
	if got.Mimetype != params.Mimetype {
		t.Errorf("Mimetype = %v, want %v", got.Mimetype, params.Mimetype)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
}

// --- Cross-database mapper consistency ---
// All three MapMedia implementations receive equivalent input and must produce
// identical wrapper Media structs.

func TestCrossDatabaseMapMedia_Consistency(t *testing.T) {
	t.Parallel()
	mediaID := types.NewMediaID()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC))

	name := sql.NullString{String: "cross-db.jpg", Valid: true}
	displayName := sql.NullString{String: "Cross DB", Valid: true}
	alt := sql.NullString{String: "Cross DB alt", Valid: true}
	caption := sql.NullString{String: "Cross DB caption", Valid: true}
	desc := sql.NullString{String: "Cross DB description", Valid: true}
	class := sql.NullString{String: "image", Valid: true}
	mime := sql.NullString{String: "image/jpeg", Valid: true}
	dims := sql.NullString{String: "100x100", Valid: true}
	url := types.URL("https://cross-db.example.com/image.jpg")
	srcset := sql.NullString{String: "cross-480.jpg 480w", Valid: true}

	sqliteInput := mdb.Media{
		MediaID: mediaID, Name: name, DisplayName: displayName, Alt: alt,
		Caption: caption, Description: desc, Class: class, Mimetype: mime,
		Dimensions: dims, URL: url, Srcset: srcset, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.Media{
		MediaID: mediaID, Name: name, DisplayName: displayName, Alt: alt,
		Caption: caption, Description: desc, Class: class, Mimetype: mime,
		Dimensions: dims, URL: url, Srcset: srcset, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.Media{
		MediaID: mediaID, Name: name, DisplayName: displayName, Alt: alt,
		Caption: caption, Description: desc, Class: class, Mimetype: mime,
		Dimensions: dims, URL: url, Srcset: srcset, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapMedia(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapMedia(mysqlInput)
	psqlResult := PsqlDatabase{}.MapMedia(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateMediaParams consistency ---
// All three always generate new IDs; verify the common fields are preserved.

func TestCrossDatabaseMapCreateMediaParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	params := mediaCreateParams()

	sqliteResult := Database{}.MapCreateMediaParams(params)
	mysqlResult := MysqlDatabase{}.MapCreateMediaParams(params)
	psqlResult := PsqlDatabase{}.MapCreateMediaParams(params)

	if sqliteResult.MediaID.IsZero() {
		t.Error("SQLite: expected non-zero generated MediaID")
	}
	if mysqlResult.MediaID.IsZero() {
		t.Error("MySQL: expected non-zero generated MediaID")
	}
	if psqlResult.MediaID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated MediaID")
	}

	// Each call should generate a unique ID
	if sqliteResult.MediaID == mysqlResult.MediaID {
		t.Error("SQLite and MySQL generated the same MediaID -- each call should be unique")
	}
	if sqliteResult.MediaID == psqlResult.MediaID {
		t.Error("SQLite and PostgreSQL generated the same MediaID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewMediaCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("media-node-1"),
		RequestID: "media-req-123",
		IP:        "10.0.0.1",
	}
	params := mediaCreateParams()

	cmd := Database{}.NewMediaCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	p, ok := cmd.Params().(CreateMediaParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateMediaParams", cmd.Params())
	}
	if p.Name != params.Name {
		t.Errorf("Params().Name = %v, want %v", p.Name, params.Name)
	}
	if p.URL != params.URL {
		t.Errorf("Params().URL = %v, want %v", p.URL, params.URL)
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewMediaCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	mediaID := types.NewMediaID()
	cmd := NewMediaCmd{}

	row := mdb.Media{MediaID: mediaID}
	got := cmd.GetID(row)
	if got != string(mediaID) {
		t.Errorf("GetID() = %q, want %q", got, string(mediaID))
	}
}

func TestUpdateMediaCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := mediaUpdateParams()

	cmd := Database{}.UpdateMediaCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	if cmd.GetID() != string(params.MediaID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(params.MediaID))
	}
	p, ok := cmd.Params().(UpdateMediaParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateMediaParams", cmd.Params())
	}
	if p.Name != params.Name {
		t.Errorf("Params().Name = %v, want %v", p.Name, params.Name)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteMediaCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	mediaID := types.NewMediaID()

	cmd := Database{}.DeleteMediaCmd(ctx, ac, mediaID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	if cmd.GetID() != string(mediaID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(mediaID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewMediaCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-media-node"),
		RequestID: "mysql-media-req",
		IP:        "192.168.1.1",
	}
	params := mediaCreateParams()

	cmd := MysqlDatabase{}.NewMediaCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	p, ok := cmd.Params().(CreateMediaParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateMediaParams", cmd.Params())
	}
	if p.Name != params.Name {
		t.Errorf("Params().Name = %v, want %v", p.Name, params.Name)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewMediaCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	mediaID := types.NewMediaID()
	cmd := NewMediaCmdMysql{}

	row := mdbm.Media{MediaID: mediaID}
	got := cmd.GetID(row)
	if got != string(mediaID) {
		t.Errorf("GetID() = %q, want %q", got, string(mediaID))
	}
}

func TestUpdateMediaCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := mediaUpdateParams()

	cmd := MysqlDatabase{}.UpdateMediaCmd(ctx, ac, params)

	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	if cmd.GetID() != string(params.MediaID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(params.MediaID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateMediaParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateMediaParams", cmd.Params())
	}
	if p.Name != params.Name {
		t.Errorf("Params().Name = %v, want %v", p.Name, params.Name)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteMediaCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	mediaID := types.NewMediaID()

	cmd := MysqlDatabase{}.DeleteMediaCmd(ctx, ac, mediaID)

	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	if cmd.GetID() != string(mediaID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(mediaID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Audited Command Accessor tests ---

func TestNewMediaCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-media-node"),
		RequestID: "psql-media-req",
		IP:        "172.16.0.1",
	}
	params := mediaCreateParams()

	cmd := PsqlDatabase{}.NewMediaCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	p, ok := cmd.Params().(CreateMediaParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateMediaParams", cmd.Params())
	}
	if p.Name != params.Name {
		t.Errorf("Params().Name = %v, want %v", p.Name, params.Name)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewMediaCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	mediaID := types.NewMediaID()
	cmd := NewMediaCmdPsql{}

	row := mdbp.Media{MediaID: mediaID}
	got := cmd.GetID(row)
	if got != string(mediaID) {
		t.Errorf("GetID() = %q, want %q", got, string(mediaID))
	}
}

func TestUpdateMediaCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := mediaUpdateParams()

	cmd := PsqlDatabase{}.UpdateMediaCmd(ctx, ac, params)

	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	if cmd.GetID() != string(params.MediaID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(params.MediaID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateMediaParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateMediaParams", cmd.Params())
	}
	if p.Name != params.Name {
		t.Errorf("Params().Name = %v, want %v", p.Name, params.Name)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteMediaCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	mediaID := types.NewMediaID()

	cmd := PsqlDatabase{}.DeleteMediaCmd(ctx, ac, mediaID)

	if cmd.TableName() != "media" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media")
	}
	if cmd.GetID() != string(mediaID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(mediaID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- Cross-database audited command consistency ---

func TestAuditedMediaCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateMediaParams{}
	updateParams := UpdateMediaParams{}
	mediaID := types.NewMediaID()

	// SQLite
	sqliteCreate := Database{}.NewMediaCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateMediaCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteMediaCmd(ctx, ac, mediaID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewMediaCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateMediaCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteMediaCmd(ctx, ac, mediaID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewMediaCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateMediaCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteMediaCmd(ctx, ac, mediaID)

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", sqliteCreate.TableName()},
		{"SQLite Update", sqliteUpdate.TableName()},
		{"SQLite Delete", sqliteDelete.TableName()},
		{"MySQL Create", mysqlCreate.TableName()},
		{"MySQL Update", mysqlUpdate.TableName()},
		{"MySQL Delete", mysqlDelete.TableName()},
		{"PostgreSQL Create", psqlCreate.TableName()},
		{"PostgreSQL Update", psqlUpdate.TableName()},
		{"PostgreSQL Delete", psqlDelete.TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "media" {
				t.Errorf("TableName() = %q, want %q", c.name, "media")
			}
		})
	}
}

func TestAuditedMediaCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateMediaParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewMediaCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewMediaCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewMediaCmd(ctx, ac, createParams).Recorder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recorder == nil {
				t.Fatalf("%s recorder is nil", tt.name)
			}
		})
	}
}

// --- Audited Command GetID consistency ---
// Verify GetID returns the correct identifier across all database variants.

func TestAuditedMediaCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	mediaID := types.NewMediaID()

	t.Run("UpdateCmd GetID returns MediaID", func(t *testing.T) {
		t.Parallel()
		params := UpdateMediaParams{MediaID: mediaID}

		sqliteCmd := Database{}.UpdateMediaCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateMediaCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateMediaCmd(ctx, ac, params)

		wantID := string(mediaID)
		if sqliteCmd.GetID() != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), wantID)
		}
		if mysqlCmd.GetID() != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), wantID)
		}
		if psqlCmd.GetID() != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), wantID)
		}
	})

	t.Run("DeleteCmd GetID returns media ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteMediaCmd(ctx, ac, mediaID)
		mysqlCmd := MysqlDatabase{}.DeleteMediaCmd(ctx, ac, mediaID)
		psqlCmd := PsqlDatabase{}.DeleteMediaCmd(ctx, ac, mediaID)

		wantID := string(mediaID)
		if sqliteCmd.GetID() != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), wantID)
		}
		if mysqlCmd.GetID() != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), wantID)
		}
		if psqlCmd.GetID() != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), wantID)
		}
	})

	t.Run("CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		testMediaID := types.NewMediaID()

		sqliteCmd := NewMediaCmd{}
		mysqlCmd := NewMediaCmdMysql{}
		psqlCmd := NewMediaCmdPsql{}

		wantID := string(testMediaID)

		sqliteRow := mdb.Media{MediaID: testMediaID}
		mysqlRow := mdbm.Media{MediaID: testMediaID}
		psqlRow := mdbp.Media{MediaID: testMediaID}

		if sqliteCmd.GetID(sqliteRow) != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), wantID)
		}
		if mysqlCmd.GetID(mysqlRow) != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), wantID)
		}
		if psqlCmd.GetID(psqlRow) != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), wantID)
		}
	})
}

// --- Edge cases ---

func TestUpdateMediaCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	// When MediaID is empty, GetID should return empty string
	params := UpdateMediaParams{MediaID: ""}

	sqliteCmd := Database{}.UpdateMediaCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateMediaCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateMediaCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteMediaCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.MediaID("")

	sqliteCmd := Database{}.DeleteMediaCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteMediaCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteMediaCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestNewMediaCmd_GetID_EmptyMediaID(t *testing.T) {
	t.Parallel()
	// CreateCmd.GetID extracts from a row; an empty MediaID in the row should return ""
	sqliteCmd := NewMediaCmd{}
	if sqliteCmd.GetID(mdb.Media{}) != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID(mdb.Media{}))
	}

	mysqlCmd := NewMediaCmdMysql{}
	if mysqlCmd.GetID(mdbm.Media{}) != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID(mdbm.Media{}))
	}

	psqlCmd := NewMediaCmdPsql{}
	if psqlCmd.GetID(mdbp.Media{}) != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID(mdbp.Media{}))
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.Media]  = NewMediaCmd{}
	_ audited.UpdateCommand[mdb.Media]  = UpdateMediaCmd{}
	_ audited.DeleteCommand[mdb.Media]  = DeleteMediaCmd{}
	_ audited.CreateCommand[mdbm.Media] = NewMediaCmdMysql{}
	_ audited.UpdateCommand[mdbm.Media] = UpdateMediaCmdMysql{}
	_ audited.DeleteCommand[mdbm.Media] = DeleteMediaCmdMysql{}
	_ audited.CreateCommand[mdbp.Media] = NewMediaCmdPsql{}
	_ audited.UpdateCommand[mdbp.Media] = UpdateMediaCmdPsql{}
	_ audited.DeleteCommand[mdbp.Media] = DeleteMediaCmdPsql{}
)
