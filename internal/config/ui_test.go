package config_test

import (
	"reflect"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// ---------------------------------------------------------------------------
// DefaultStyle -- verify all fields are populated
// ---------------------------------------------------------------------------

func TestDefaultStyle_AllFieldsPopulated(t *testing.T) {
	t.Parallel()

	s := config.DefaultStyle

	// Use reflection to iterate over every field in Color and verify it is
	// not nil. This catches new fields added to Color that are forgotten
	// in DefaultStyle initialization.
	v := reflect.ValueOf(s)
	ty := v.Type()

	for i := range ty.NumField() {
		field := ty.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			t.Parallel()
			fv := v.Field(i)
			if fv.IsNil() {
				t.Errorf("DefaultStyle.%s is nil (not initialized)", field.Name)
			}
		})
	}
}

func TestDefaultStyle_FieldCount(t *testing.T) {
	t.Parallel()

	// Guard against adding Color fields without updating DefaultStyle.
	// When this test breaks, the developer must add the field to DefaultStyle
	// and update this count.
	v := reflect.ValueOf(config.DefaultStyle)
	got := v.NumField()
	want := 22 // Primary, PrimaryBG, Secondary, SecondaryBG, Tertiary, TertiaryBG,
	// Accent, AccentBG, Accent2, Accent2BG, Active, ActiveBG,
	// Status1, Status1BG, Status2, Status2BG, Status3, Status3BG,
	// PrimaryBorder, AdminAccent, Warn, WarnBG

	if got != want {
		t.Errorf("Color struct has %d fields, want %d -- update DefaultStyle and this test", got, want)
	}
}

// ---------------------------------------------------------------------------
// Color struct -- JSON tags
// ---------------------------------------------------------------------------

func TestColor_HasJSONTags(t *testing.T) {
	t.Parallel()

	// Every field in Color should have a json tag so it can be serialized
	// in config files. This catches fields added without tags.
	ty := reflect.TypeOf(config.Color{})

	for i := range ty.NumField() {
		field := ty.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			t.Parallel()
			tag := field.Tag.Get("json")
			if tag == "" {
				t.Errorf("Color.%s has no json tag", field.Name)
			}
		})
	}
}
