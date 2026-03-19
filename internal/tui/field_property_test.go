package tui

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

func TestFieldPropertiesFromField_Count(t *testing.T) {
	f := db.Fields{
		FieldID:  "f1",
		Label:    "Title",
		Name:     "title",
		Type:     "text",
		Data:     "{}",
		UIConfig: "{}",
	}
	props := FieldPropertiesFromField(f)
	if len(props) != 13 {
		t.Fatalf("expected 13 properties, got %d", len(props))
	}
}

func TestFieldPropertiesFromField_EditableOrder(t *testing.T) {
	f := db.Fields{FieldID: "f1", Label: "Title", Name: "title", Type: "text"}
	props := FieldPropertiesFromField(f)

	// First 9 should be editable
	for i := 0; i < 9; i++ {
		if !props[i].Editable {
			t.Errorf("expected props[%d] (%s) to be editable", i, props[i].Key)
		}
	}
	// Last 4 should be read-only
	for i := 9; i < 13; i++ {
		if props[i].Editable {
			t.Errorf("expected props[%d] (%s) to be read-only", i, props[i].Key)
		}
	}
}

func TestFieldPropertiesFromField_Values(t *testing.T) {
	f := db.Fields{
		FieldID:      "f1",
		Label:        "Hero",
		Name:         "hero",
		Type:         "group",
		SortOrder:    3,
		Data:         `{"key":"value"}`,
		ValidationID: types.NullableValidationID{},
		UIConfig:     `{"widget":"textarea"}`,
		Translatable: true,
		Roles:        types.NullableString{String: "editor", Valid: true},
	}
	props := FieldPropertiesFromField(f)

	checks := map[string]string{
		"Label":        "Hero",
		"Name":         "hero",
		"Type":         "group",
		"Sort Order":   "3",
		"Data":         `{"key":"value"}`,
		"Validation ID": "null",
		"UI Config":    `{"widget":"textarea"}`,
		"Translatable": "true",
		"Roles":        "editor",
		"ID":           "f1",
	}
	for _, p := range props {
		if expected, ok := checks[p.Key]; ok {
			if p.Value != expected {
				t.Errorf("property %q: expected %q, got %q", p.Key, expected, p.Value)
			}
		}
	}
}

func TestFieldPropertiesFromField_UIConfigNone(t *testing.T) {
	f := db.Fields{FieldID: "f1", UIConfig: "{}"}
	props := FieldPropertiesFromField(f)
	for _, p := range props {
		if p.Key == "UI Config" {
			if p.Value != "(none)" {
				t.Errorf("expected (none) for empty UIConfig, got %q", p.Value)
			}
			return
		}
	}
	t.Fatal("UI Config property not found")
}

func TestFieldPropertiesFromAdminField_ID(t *testing.T) {
	f := db.AdminFields{AdminFieldID: "af1", Label: "Test"}
	props := FieldPropertiesFromAdminField(f)
	for _, p := range props {
		if p.Key == "ID" {
			if p.Value != "af1" {
				t.Errorf("expected admin field ID af1, got %q", p.Value)
			}
			return
		}
	}
	t.Fatal("ID property not found")
}

func TestCompactJSON_Long(t *testing.T) {
	long := `{"very_long_key":"very_long_value_that_exceeds_the_forty_character_limit"}`
	result := compactJSON(long)
	if len(result) > 40 {
		t.Errorf("expected truncated to 40 chars, got %d", len(result))
	}
	if result[len(result)-3:] != "..." {
		t.Errorf("expected ellipsis suffix, got %q", result)
	}
}

func TestNullableStringDisplay(t *testing.T) {
	if nullableStringDisplay("", false) != "(none)" {
		t.Error("invalid nullable should be (none)")
	}
	if nullableStringDisplay("test", true) != "test" {
		t.Error("valid nullable should show value")
	}
}
