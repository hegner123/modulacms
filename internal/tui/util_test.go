package tui

import "testing"

// ============================================================
// Required
// ============================================================

func TestRequired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "non-empty string", input: "hello", wantErr: false},
		{name: "single char", input: "x", wantErr: false},
		{name: "whitespace only", input: "   ", wantErr: false}, // not trimmed, passes
		{name: "empty string", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Required(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Required(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// InitTables
// ============================================================

func TestInitTables(t *testing.T) {
	t.Parallel()

	t.Run("creates identity map", func(t *testing.T) {
		t.Parallel()
		tables := []string{"users", "content", "media"}
		got := InitTables(tables)
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
		for _, name := range tables {
			if got[name] != name {
				t.Errorf("got[%q] = %q, want %q", name, got[name], name)
			}
		}
	})

	t.Run("empty slice returns empty map", func(t *testing.T) {
		t.Parallel()
		got := InitTables([]string{})
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}
	})

	t.Run("nil slice returns empty map", func(t *testing.T) {
		t.Parallel()
		got := InitTables(nil)
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}
	})

	t.Run("duplicates collapse", func(t *testing.T) {
		t.Parallel()
		got := InitTables([]string{"a", "a", "b"})
		if len(got) != 2 {
			t.Errorf("len = %d, want 2", len(got))
		}
	})
}
