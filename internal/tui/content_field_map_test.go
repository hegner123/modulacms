package tui

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ============================================================
// MapContentFieldsToDisplay
// ============================================================

func TestMapContentFieldsToDisplay(t *testing.T) {
	t.Parallel()

	fieldID1 := types.FieldID("01FIELD000000000000000001")
	fieldID2 := types.FieldID("01FIELD000000000000000002")
	contentID := types.ContentID("01CONTENT00000000000000A")
	cfID1 := types.ContentFieldID("01CF00000000000000000001")
	cfID2 := types.ContentFieldID("01CF00000000000000000002")

	fieldDefs := []db.Fields{
		{
			FieldID:      fieldID1,
			Label:        "Title",
			Type:         "text",
			ValidationID: types.NullableValidationID{},
			Data:         `{}`,
		},
		{
			FieldID:      fieldID2,
			Label:        "Body",
			Type:         "richtext",
			ValidationID: types.NullableValidationID{},
			Data:         `{"format":"markdown"}`,
		},
	}

	t.Run("maps fields with matching definitions", func(t *testing.T) {
		t.Parallel()
		contentFields := []db.ContentFields{
			{
				ContentFieldID: cfID1,
				ContentDataID:  types.NullableContentID{ID: contentID, Valid: true},
				FieldID:        types.NullableFieldID{ID: fieldID1, Valid: true},
				FieldValue:     "Hello World",
			},
			{
				ContentFieldID: cfID2,
				ContentDataID:  types.NullableContentID{ID: contentID, Valid: true},
				FieldID:        types.NullableFieldID{ID: fieldID2, Valid: true},
				FieldValue:     "Some body text",
			},
		}

		result := MapContentFieldsToDisplay(contentFields, fieldDefs)
		displays := result[contentID]
		if len(displays) != 2 {
			t.Fatalf("expected 2 displays, got %d", len(displays))
		}

		// Verify first field
		if displays[0].Label != "Title" {
			t.Errorf("displays[0].Label = %q, want %q", displays[0].Label, "Title")
		}
		if displays[0].Type != "text" {
			t.Errorf("displays[0].Type = %q, want %q", displays[0].Type, "text")
		}
		if displays[0].Value != "Hello World" {
			t.Errorf("displays[0].Value = %q, want %q", displays[0].Value, "Hello World")
		}
		if displays[0].ValidationJSON != "" {
			t.Errorf("displays[0].ValidationJSON = %q, want empty (validation is now FK-based)", displays[0].ValidationJSON)
		}
	})

	t.Run("skips fields with invalid ContentDataID", func(t *testing.T) {
		t.Parallel()
		contentFields := []db.ContentFields{
			{
				ContentFieldID: cfID1,
				ContentDataID:  types.NullableContentID{Valid: false},
				FieldID:        types.NullableFieldID{ID: fieldID1, Valid: true},
				FieldValue:     "orphan",
			},
		}

		result := MapContentFieldsToDisplay(contentFields, fieldDefs)
		if len(result) != 0 {
			t.Errorf("expected empty result for invalid ContentDataID, got %d entries", len(result))
		}
	})

	t.Run("skips fields with invalid FieldID", func(t *testing.T) {
		t.Parallel()
		contentFields := []db.ContentFields{
			{
				ContentFieldID: cfID1,
				ContentDataID:  types.NullableContentID{ID: contentID, Valid: true},
				FieldID:        types.NullableFieldID{Valid: false},
				FieldValue:     "orphan",
			},
		}

		result := MapContentFieldsToDisplay(contentFields, fieldDefs)
		if len(result) != 0 {
			t.Errorf("expected empty result for invalid FieldID, got %d entries", len(result))
		}
	})

	t.Run("field without matching definition gets empty label and type", func(t *testing.T) {
		t.Parallel()
		unknownFieldID := types.FieldID("01FIELD0000000000UNKNOWN")
		contentFields := []db.ContentFields{
			{
				ContentFieldID: cfID1,
				ContentDataID:  types.NullableContentID{ID: contentID, Valid: true},
				FieldID:        types.NullableFieldID{ID: unknownFieldID, Valid: true},
				FieldValue:     "value without def",
			},
		}

		result := MapContentFieldsToDisplay(contentFields, fieldDefs)
		displays := result[contentID]
		if len(displays) != 1 {
			t.Fatalf("expected 1 display, got %d", len(displays))
		}
		if displays[0].Label != "" {
			t.Errorf("Label = %q, want empty for unmatched field", displays[0].Label)
		}
		if displays[0].Type != "" {
			t.Errorf("Type = %q, want empty for unmatched field", displays[0].Type)
		}
		if displays[0].Value != "value without def" {
			t.Errorf("Value = %q, want %q", displays[0].Value, "value without def")
		}
	})

	t.Run("empty inputs return empty map", func(t *testing.T) {
		t.Parallel()
		result := MapContentFieldsToDisplay(nil, nil)
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d entries", len(result))
		}
	})

	t.Run("groups by ContentDataID", func(t *testing.T) {
		t.Parallel()
		contentID2 := types.ContentID("01CONTENT00000000000000B")
		contentFields := []db.ContentFields{
			{
				ContentFieldID: cfID1,
				ContentDataID:  types.NullableContentID{ID: contentID, Valid: true},
				FieldID:        types.NullableFieldID{ID: fieldID1, Valid: true},
				FieldValue:     "v1",
			},
			{
				ContentFieldID: cfID2,
				ContentDataID:  types.NullableContentID{ID: contentID2, Valid: true},
				FieldID:        types.NullableFieldID{ID: fieldID2, Valid: true},
				FieldValue:     "v2",
			},
		}

		result := MapContentFieldsToDisplay(contentFields, fieldDefs)
		if len(result) != 2 {
			t.Fatalf("expected 2 content groups, got %d", len(result))
		}
		if len(result[contentID]) != 1 {
			t.Errorf("contentID group has %d items, want 1", len(result[contentID]))
		}
		if len(result[contentID2]) != 1 {
			t.Errorf("contentID2 group has %d items, want 1", len(result[contentID2]))
		}
	})
}
