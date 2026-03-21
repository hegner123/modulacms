package definitions

import (
	"testing"
)

func TestRegistryAllDefinitionsRegistered(t *testing.T) {
	expected := []string{
		"modula-default",
		"contentful-starter",
		"sanity-starter",
		"strapi-starter",
		"tailwind-marketing",
		"wordpress-blog",
	}

	names := Names()
	if len(names) != len(expected) {
		t.Fatalf("expected %d definitions, got %d: %v", len(expected), len(names), names)
	}

	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected names[%d] = %q, got %q", i, name, names[i])
		}
	}
}

func TestRegistryGetReturnsCorrect(t *testing.T) {
	def, ok := Get("wordpress-blog")
	if !ok {
		t.Fatal("Get(\"wordpress-blog\") returned false")
	}
	if def.Label != "WordPress Blog" {
		t.Errorf("expected label %q, got %q", "WordPress Blog", def.Label)
	}
	if def.Format != "wordpress" {
		t.Errorf("expected format %q, got %q", "wordpress", def.Format)
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	_, ok := Get("does-not-exist")
	if ok {
		t.Error("Get(\"does-not-exist\") should return false")
	}
}

func TestRegistryNamesSorted(t *testing.T) {
	names := Names()
	if len(names) == 0 {
		t.Fatal("no names returned")
	}
	// First entry is always modula-default
	if names[0] != "modula-default" {
		t.Errorf("expected names[0] = %q, got %q", "modula-default", names[0])
	}
	// Remaining entries are alphabetically sorted
	for i := 2; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("names not sorted: %q before %q", names[i-1], names[i])
		}
	}
}

func TestRegistryListMatchesNames(t *testing.T) {
	names := Names()
	defs := List()
	if len(defs) != len(names) {
		t.Fatalf("List() returned %d items, Names() returned %d", len(defs), len(names))
	}
	for i, def := range defs {
		if def.Name != names[i] {
			t.Errorf("List()[%d].Name = %q, Names()[%d] = %q", i, def.Name, i, names[i])
		}
	}
}

func TestRegistryAllDefinitionsValid(t *testing.T) {
	for _, def := range List() {
		t.Run(def.Name, func(t *testing.T) {
			if err := Validate(def); err != nil {
				t.Errorf("Validate(%q) failed: %v", def.Name, err)
			}
		})
	}
}

func TestRegistryDuplicatePanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()

	Register(SchemaDefinition{
		Name: "modula-default", // already registered
	})
}
