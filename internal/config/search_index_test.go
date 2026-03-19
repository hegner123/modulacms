package config

import "testing"

func TestBuildSearchIndex_CoversAllRegistryFields(t *testing.T) {
	index := BuildSearchIndex()
	if len(index) != len(FieldRegistry) {
		t.Fatalf("index has %d entries, FieldRegistry has %d", len(index), len(FieldRegistry))
	}

	for i, entry := range index {
		reg := FieldRegistry[i]
		if entry.Key != reg.JSONKey {
			t.Errorf("entry %d: key %q != registry key %q", i, entry.Key, reg.JSONKey)
		}
		if entry.Label == "" {
			t.Errorf("entry %q: empty label", entry.Key)
		}
		if entry.Category == "" {
			t.Errorf("entry %q: empty category", entry.Key)
		}
		if entry.CategoryLabel == "" {
			t.Errorf("entry %q: empty category_label", entry.Key)
		}
		if entry.Description == "" {
			t.Errorf("entry %q: empty description", entry.Key)
		}
	}
}

func TestBuildSearchIndex_HelpTextPopulated(t *testing.T) {
	index := BuildSearchIndex()

	// Spot-check a few well-known fields that have help text in HELP_TEXT.md.
	checks := map[string]struct {
		wantHelpContains   string
		wantDefaultContains string
	}{
		"port":       {wantHelpContains: "HTTP server listens on", wantDefaultContains: ":8080"},
		"db_driver":  {wantHelpContains: "database backend", wantDefaultContains: "sqlite"},
		"plugin_enabled": {wantHelpContains: "Lua plugin system", wantDefaultContains: "false"},
		"email_provider": {wantHelpContains: "email sending backend", wantDefaultContains: ""},
	}

	byKey := make(map[string]SearchIndexEntry, len(index))
	for _, e := range index {
		byKey[e.Key] = e
	}

	for key, check := range checks {
		entry, ok := byKey[key]
		if !ok {
			t.Errorf("missing index entry for %q", key)
			continue
		}
		if check.wantHelpContains != "" && !contains(entry.HelpText, check.wantHelpContains) {
			t.Errorf("%q: help_text %q does not contain %q", key, entry.HelpText, check.wantHelpContains)
		}
		if check.wantDefaultContains != "" && !contains(entry.Default, check.wantDefaultContains) {
			t.Errorf("%q: default %q does not contain %q", key, entry.Default, check.wantDefaultContains)
		}
	}
}

func TestBuildSearchIndex_NoDuplicateKeys(t *testing.T) {
	index := BuildSearchIndex()
	seen := make(map[string]bool, len(index))
	for _, e := range index {
		if seen[e.Key] {
			t.Errorf("duplicate key: %q", e.Key)
		}
		seen[e.Key] = true
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
