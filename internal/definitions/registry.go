package definitions

import "fmt"

var registry = map[string]SchemaDefinition{}

// Register adds a SchemaDefinition to the global registry.
// Panics on duplicate name — this is a programming error caught at init time.
func Register(def SchemaDefinition) {
	if _, exists := registry[def.Name]; exists {
		panic(fmt.Sprintf("definitions: duplicate schema name %q", def.Name))
	}
	registry[def.Name] = def
}

// Get returns a registered SchemaDefinition by name and whether it was found.
func Get(name string) (SchemaDefinition, bool) {
	def, ok := registry[name]
	return def, ok
}

// List returns all registered definitions sorted by name.
func List() []SchemaDefinition {
	names := Names()
	result := make([]SchemaDefinition, len(names))
	for i, name := range names {
		result[i] = registry[name]
	}
	return result
}

// Names returns all registered definition names in sorted order.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	// Insertion sort — small N, avoids sort import
	for i := 1; i < len(names); i++ {
		key := names[i]
		j := i - 1
		for j >= 0 && names[j] > key {
			names[j+1] = names[j]
			j--
		}
		names[j+1] = key
	}
	return names
}
