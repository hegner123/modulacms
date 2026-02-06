package cli

// FieldInputEntry describes a registered field input type.
type FieldInputEntry struct {
	Key         string             // Type identifier stored in DB (e.g., "text", "textarea")
	Label       string             // Display name in selector (e.g., "Text", "Textarea")
	Description string             // Short description for help text
	NewBubble   func() FieldBubble // Factory that creates a new bubble instance
}

// fieldInputRegistry is the internal registry slice.
// Order determines display order in the type selector carousel.
var fieldInputRegistry []FieldInputEntry

// RegisterFieldInput adds a field input type to the registry.
// Call during init() or startup -- not goroutine-safe.
func RegisterFieldInput(entry FieldInputEntry) {
	fieldInputRegistry = append(fieldInputRegistry, entry)
}

// FieldInputTypes returns a copy of all registered field input entries.
func FieldInputTypes() []FieldInputEntry {
	out := make([]FieldInputEntry, len(fieldInputRegistry))
	copy(out, fieldInputRegistry)
	return out
}

// FieldInputTypeIndex returns the index of the given key in the registry,
// or 0 if not found (defaults to first entry).
func FieldInputTypeIndex(key string) int {
	for i, e := range fieldInputRegistry {
		if e.Key == key {
			return i
		}
	}
	return 0
}
