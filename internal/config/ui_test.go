package config_test

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// ---------------------------------------------------------------------------
// DefaultStyle -- verify all fields are populated
// ---------------------------------------------------------------------------

func TestDefaultStyle_AllFieldsPopulated(t *testing.T) {
	t.Parallel()

	s := config.DefaultStyle

	// Use reflection to iterate over every field in Color and verify it is
	// not the zero value of CompleteAdaptiveColor. This catches new fields
	// added to Color that are forgotten in DefaultStyle initialization.
	v := reflect.ValueOf(s)
	ty := v.Type()

	zero := reflect.ValueOf(lipgloss.CompleteAdaptiveColor{})

	for i := range ty.NumField() {
		field := ty.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			t.Parallel()
			fv := v.Field(i)
			if reflect.DeepEqual(fv.Interface(), zero.Interface()) {
				t.Errorf("DefaultStyle.%s is zero-value (not initialized)", field.Name)
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
	want := 21 // Primary, PrimaryBG, Secondary, SecondaryBG, Tertiary, TertiaryBG,
	// Accent, AccentBG, Accent2, Accent2BG, Active, ActiveBG,
	// Status1, Status1BG, Status2, Status2BG, Status3, Status3BG,
	// PrimaryBorder, Warn, WarnBG

	if got != want {
		t.Errorf("Color struct has %d fields, want %d -- update DefaultStyle and this test", got, want)
	}
}

// ---------------------------------------------------------------------------
// DefaultStyle -- spot-check specific semantic pairs
// ---------------------------------------------------------------------------

func TestDefaultStyle_PrimaryPair(t *testing.T) {
	t.Parallel()

	s := config.DefaultStyle

	// In dark mode, primary text is white; in light mode, black.
	if s.Primary.Dark.TrueColor != "#FFFFFF" {
		t.Errorf("Primary.Dark.TrueColor = %q, want %q", s.Primary.Dark.TrueColor, "#FFFFFF")
	}
	if s.Primary.Light.TrueColor != "#000000" {
		t.Errorf("Primary.Light.TrueColor = %q, want %q", s.Primary.Light.TrueColor, "#000000")
	}
}

func TestDefaultStyle_AccentPair(t *testing.T) {
	t.Parallel()

	s := config.DefaultStyle

	// Accent uses purple in both light and dark modes.
	if s.Accent.Dark.TrueColor != "#6612e3" {
		t.Errorf("Accent.Dark.TrueColor = %q, want %q", s.Accent.Dark.TrueColor, "#6612e3")
	}
	if s.Accent.Light.TrueColor != "#6612e3" {
		t.Errorf("Accent.Light.TrueColor = %q, want %q", s.Accent.Light.TrueColor, "#6612e3")
	}
}

func TestDefaultStyle_WarnPair(t *testing.T) {
	t.Parallel()

	s := config.DefaultStyle

	// Warn uses orange for both themes.
	if s.Warn.Dark.TrueColor != "#F75C03" {
		t.Errorf("Warn.Dark.TrueColor = %q, want %q", s.Warn.Dark.TrueColor, "#F75C03")
	}
	if s.Warn.Light.TrueColor != "#F75C03" {
		t.Errorf("Warn.Light.TrueColor = %q, want %q", s.Warn.Light.TrueColor, "#F75C03")
	}

	// Warn background is white in both themes.
	if s.WarnBG.Dark.TrueColor != "#FFFFFF" {
		t.Errorf("WarnBG.Dark.TrueColor = %q, want %q", s.WarnBG.Dark.TrueColor, "#FFFFFF")
	}
	if s.WarnBG.Light.TrueColor != "#FFFFFF" {
		t.Errorf("WarnBG.Light.TrueColor = %q, want %q", s.WarnBG.Light.TrueColor, "#FFFFFF")
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
