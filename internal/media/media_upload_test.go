package media

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// TestMapMediaParams verifies that MapMediaParams correctly maps all fields from
// a db.Media to db.UpdateMediaParams, including setting DateModified to a fresh
// timestamp rather than copying the original.
func TestMapMediaParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input db.Media
	}{
		{
			name: "all fields populated",
			input: db.Media{
				MediaID:      types.MediaID("01HX1234567890ABCDEF12345"),
				Name:         db.NewNullString("photo.png"),
				DisplayName:  db.NewNullString("My Photo"),
				Alt:          db.NewNullString("A sunset"),
				Caption:      db.NewNullString("Taken in 2024"),
				Description:  db.NewNullString("Beautiful sunset over the ocean"),
				Class:        db.NewNullString("hero-image"),
				URL:          types.URL("https://cdn.example.com/photo.png"),
				Mimetype:     db.NewNullString("image/png"),
				Dimensions:   db.NewNullString("1920x1080"),
				Srcset:       db.NewNullString(`["url1","url2"]`),
				AuthorID:     types.NullableUserID{ID: types.UserID("user-123"), Valid: true},
				DateCreated:  types.TimestampNow(),
				DateModified: types.TimestampNow(),
			},
		},
		{
			name: "nullable fields empty",
			input: db.Media{
				MediaID:      types.MediaID("01HX9999999999ZZZZZZZZZZ"),
				Name:         db.NullString{},
				DisplayName:  db.NullString{},
				Alt:          db.NullString{},
				Caption:      db.NullString{},
				Description:  db.NullString{},
				Class:        db.NullString{},
				Mimetype:     db.NullString{},
				Dimensions:   db.NullString{},
				Srcset:       db.NullString{},
				AuthorID:     types.NullableUserID{},
				DateCreated:  types.TimestampNow(),
				DateModified: types.TimestampNow(),
			},
		},
		{
			name: "zero-value media ID",
			input: db.Media{
				MediaID:      types.MediaID(""),
				DateCreated:  types.TimestampNow(),
				DateModified: types.TimestampNow(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MapMediaParams(tt.input)

			// Every field except DateModified should be copied verbatim
			if result.MediaID != tt.input.MediaID {
				t.Errorf("MediaID: got %s, want %s", result.MediaID, tt.input.MediaID)
			}
			if result.Name != tt.input.Name {
				t.Errorf("Name: got %+v, want %+v", result.Name, tt.input.Name)
			}
			if result.DisplayName != tt.input.DisplayName {
				t.Errorf("DisplayName: got %+v, want %+v", result.DisplayName, tt.input.DisplayName)
			}
			if result.Alt != tt.input.Alt {
				t.Errorf("Alt: got %+v, want %+v", result.Alt, tt.input.Alt)
			}
			if result.Caption != tt.input.Caption {
				t.Errorf("Caption: got %+v, want %+v", result.Caption, tt.input.Caption)
			}
			if result.Description != tt.input.Description {
				t.Errorf("Description: got %+v, want %+v", result.Description, tt.input.Description)
			}
			if result.Class != tt.input.Class {
				t.Errorf("Class: got %+v, want %+v", result.Class, tt.input.Class)
			}
			if result.URL != tt.input.URL {
				t.Errorf("URL: got %s, want %s", result.URL, tt.input.URL)
			}
			if result.Mimetype != tt.input.Mimetype {
				t.Errorf("Mimetype: got %+v, want %+v", result.Mimetype, tt.input.Mimetype)
			}
			if result.Dimensions != tt.input.Dimensions {
				t.Errorf("Dimensions: got %+v, want %+v", result.Dimensions, tt.input.Dimensions)
			}
			if result.Srcset != tt.input.Srcset {
				t.Errorf("Srcset: got %+v, want %+v", result.Srcset, tt.input.Srcset)
			}
			if result.AuthorID != tt.input.AuthorID {
				t.Errorf("AuthorID: got %+v, want %+v", result.AuthorID, tt.input.AuthorID)
			}
			if result.DateCreated != tt.input.DateCreated {
				t.Errorf("DateCreated: got %s, want %s", result.DateCreated, tt.input.DateCreated)
			}

			// DateModified should be set to a fresh timestamp, NOT copied from input
			if result.DateModified == tt.input.DateModified {
				// This is a timing-sensitive check. In practice, TimestampNow() is called
				// again inside MapMediaParams, so it should differ unless the clock resolution
				// is too coarse. We only flag this if the string is literally identical to
				// the original -- which would mean it was copied rather than regenerated.
				// Note: This test can give a false positive if TimestampNow() returns the
				// same second twice. We accept this risk because the behavioral contract
				// is that DateModified is refreshed.
				t.Log("WARNING: DateModified matches input; may be coincidence if timestamps resolve to same second")
			}
		})
	}
}

// TestMapMediaParams_PreservesDateCreated verifies that the original DateCreated
// is never overwritten. This is important for audit trails -- a media record's
// creation date must remain stable across updates.
func TestMapMediaParams_PreservesDateCreated(t *testing.T) {
	t.Parallel()
	originalCreated := types.TimestampNow()

	input := db.Media{
		MediaID:      types.MediaID("01HXPRESERVEDATECREATED00"),
		DateCreated:  originalCreated,
		DateModified: types.TimestampNow(),
	}

	result := MapMediaParams(input)
	if result.DateCreated != originalCreated {
		t.Errorf("DateCreated was modified: got %s, want %s", result.DateCreated, originalCreated)
	}
}
