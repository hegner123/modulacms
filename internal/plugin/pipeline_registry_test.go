package plugin

// Tests for PipelineRegistry: build, lookup (Before/After/Get), sorting,
// reload, disabled-row filtering, and parseConfigJSON edge cases.

import (
	"sync"
	"testing"
)

func TestNewPipelineRegistry_IsEmpty(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()

	if !r.IsEmpty() {
		t.Error("new registry should be empty")
	}
}

func TestBuild_PopulatesChains(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{
			PipelineID: "p1",
			PluginName: "seo",
			Handler:    "on_before_create",
			TableName:  "posts",
			Operation:  "before_create",
			Priority:   10,
			Enabled:    true,
			Config:     `{"key":"val"}`,
		},
	})

	if r.IsEmpty() {
		t.Fatal("registry should not be empty after Build with enabled rows")
	}

	entries := r.Before("posts", "create")
	if len(entries) != 1 {
		t.Fatalf("Before(posts, create) len = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.PipelineID != "p1" {
		t.Errorf("PipelineID = %q, want %q", e.PipelineID, "p1")
	}
	if e.PluginName != "seo" {
		t.Errorf("PluginName = %q, want %q", e.PluginName, "seo")
	}
	if e.Handler != "on_before_create" {
		t.Errorf("Handler = %q, want %q", e.Handler, "on_before_create")
	}
	if e.Priority != 10 {
		t.Errorf("Priority = %d, want %d", e.Priority, 10)
	}
	if !e.Enabled {
		t.Error("Enabled = false, want true")
	}
	if e.Config == nil || e.Config["key"] != "val" {
		t.Errorf("Config = %v, want map with key=val", e.Config)
	}
}

func TestBuild_FiltersDisabledRows(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{
			PipelineID: "p1",
			PluginName: "seo",
			TableName:  "posts",
			Operation:  "before_create",
			Priority:   10,
			Enabled:    false,
		},
		{
			PipelineID: "p2",
			PluginName: "cache",
			TableName:  "posts",
			Operation:  "after_create",
			Priority:   20,
			Enabled:    true,
		},
	})

	if r.Before("posts", "create") != nil {
		t.Error("disabled row should not appear in Before lookup")
	}

	entries := r.After("posts", "create")
	if len(entries) != 1 {
		t.Fatalf("After(posts, create) len = %d, want 1", len(entries))
	}
	if entries[0].PluginName != "cache" {
		t.Errorf("PluginName = %q, want %q", entries[0].PluginName, "cache")
	}
}

func TestBuild_AllDisabledProducesEmptyRegistry(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", Enabled: false},
		{PipelineID: "p2", TableName: "posts", Operation: "after_create", Enabled: false},
	})

	if !r.IsEmpty() {
		t.Error("registry should be empty when all rows are disabled")
	}
}

func TestBuild_NilRows(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build(nil)

	if !r.IsEmpty() {
		t.Error("Build(nil) should leave registry empty")
	}
}

func TestBuild_EmptyRows(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{})

	if !r.IsEmpty() {
		t.Error("Build(empty) should leave registry empty")
	}
}

func TestBefore_ReturnsNilForMissingKey(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Enabled: true},
	})

	if got := r.Before("posts", "update"); got != nil {
		t.Errorf("Before(posts, update) = %v, want nil", got)
	}
	if got := r.Before("pages", "create"); got != nil {
		t.Errorf("Before(pages, create) = %v, want nil", got)
	}
}

func TestAfter_ReturnsNilForMissingKey(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "after_delete", PluginName: "audit", Enabled: true},
	})

	if got := r.After("posts", "create"); got != nil {
		t.Errorf("After(posts, create) = %v, want nil", got)
	}
}

func TestGet_ReturnsExactOperationKey(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true},
		{PipelineID: "p2", TableName: "posts", Operation: "after_create", PluginName: "cache", Priority: 20, Enabled: true},
	})

	before := r.Get("posts", "before_create")
	if len(before) != 1 {
		t.Fatalf("Get(posts, before_create) len = %d, want 1", len(before))
	}
	if before[0].PluginName != "seo" {
		t.Errorf("PluginName = %q, want %q", before[0].PluginName, "seo")
	}

	after := r.Get("posts", "after_create")
	if len(after) != 1 {
		t.Fatalf("Get(posts, after_create) len = %d, want 1", len(after))
	}
	if after[0].PluginName != "cache" {
		t.Errorf("PluginName = %q, want %q", after[0].PluginName, "cache")
	}
}

func TestGet_ReturnsNilForMissingKey(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()

	if got := r.Get("posts", "before_create"); got != nil {
		t.Errorf("Get on empty registry = %v, want nil", got)
	}
}

func TestSortOrder_PriorityAscending(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p3", TableName: "posts", Operation: "before_create", PluginName: "c_plugin", Priority: 30, Enabled: true},
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "a_plugin", Priority: 10, Enabled: true},
		{PipelineID: "p2", TableName: "posts", Operation: "before_create", PluginName: "b_plugin", Priority: 20, Enabled: true},
	})

	entries := r.Before("posts", "create")
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}

	want := []struct {
		name     string
		priority int
	}{
		{"a_plugin", 10},
		{"b_plugin", 20},
		{"c_plugin", 30},
	}

	for i, w := range want {
		if entries[i].PluginName != w.name {
			t.Errorf("entries[%d].PluginName = %q, want %q", i, entries[i].PluginName, w.name)
		}
		if entries[i].Priority != w.priority {
			t.Errorf("entries[%d].Priority = %d, want %d", i, entries[i].Priority, w.priority)
		}
	}
}

func TestSortOrder_TiebreakByPluginName(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p2", TableName: "posts", Operation: "before_create", PluginName: "zebra", Priority: 10, Enabled: true},
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "alpha", Priority: 10, Enabled: true},
		{PipelineID: "p3", TableName: "posts", Operation: "before_create", PluginName: "mike", Priority: 10, Enabled: true},
	})

	entries := r.Before("posts", "create")
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}

	wantNames := []string{"alpha", "mike", "zebra"}
	for i, name := range wantNames {
		if entries[i].PluginName != name {
			t.Errorf("entries[%d].PluginName = %q, want %q", i, entries[i].PluginName, name)
		}
	}
}

func TestSortOrder_MixedPriorityAndName(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "t", Operation: "before_create", PluginName: "beta", Priority: 5, Enabled: true},
		{PipelineID: "p2", TableName: "t", Operation: "before_create", PluginName: "alpha", Priority: 5, Enabled: true},
		{PipelineID: "p3", TableName: "t", Operation: "before_create", PluginName: "gamma", Priority: 1, Enabled: true},
	})

	entries := r.Before("t", "create")
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}

	// gamma (priority 1) first, then alpha and beta (both priority 5, alpha < beta)
	wantNames := []string{"gamma", "alpha", "beta"}
	for i, name := range wantNames {
		if entries[i].PluginName != name {
			t.Errorf("entries[%d].PluginName = %q, want %q", i, entries[i].PluginName, name)
		}
	}
}

func TestReload_ReplacesExistingChains(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()

	// First load
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true},
		{PipelineID: "p2", TableName: "posts", Operation: "after_delete", PluginName: "audit", Priority: 5, Enabled: true},
	})

	if r.Before("posts", "create") == nil {
		t.Fatal("before_create should exist after first Build")
	}
	if r.After("posts", "delete") == nil {
		t.Fatal("after_delete should exist after first Build")
	}

	// Reload with different data -- old chains should be replaced entirely.
	r.Reload([]PipelineRow{
		{PipelineID: "p3", TableName: "pages", Operation: "before_update", PluginName: "validator", Priority: 1, Enabled: true},
	})

	// Old keys should be gone.
	if got := r.Before("posts", "create"); got != nil {
		t.Errorf("old key Before(posts, create) should be nil after Reload, got %v", got)
	}
	if got := r.After("posts", "delete"); got != nil {
		t.Errorf("old key After(posts, delete) should be nil after Reload, got %v", got)
	}

	// New key should be present.
	entries := r.Before("pages", "update")
	if len(entries) != 1 {
		t.Fatalf("Before(pages, update) len = %d, want 1", len(entries))
	}
	if entries[0].PluginName != "validator" {
		t.Errorf("PluginName = %q, want %q", entries[0].PluginName, "validator")
	}
}

func TestReload_WithEmptyRowsClearsRegistry(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Enabled: true},
	})

	if r.IsEmpty() {
		t.Fatal("registry should not be empty after Build")
	}

	r.Reload([]PipelineRow{})

	if !r.IsEmpty() {
		t.Error("Reload with empty rows should clear the registry")
	}
}

func TestMultipleTablesAndOperations(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true},
		{PipelineID: "p2", TableName: "posts", Operation: "after_create", PluginName: "cache", Priority: 5, Enabled: true},
		{PipelineID: "p3", TableName: "pages", Operation: "before_update", PluginName: "validator", Priority: 1, Enabled: true},
		{PipelineID: "p4", TableName: "pages", Operation: "after_delete", PluginName: "audit", Priority: 20, Enabled: true},
	})

	tests := []struct {
		name    string
		table   string
		op      string
		method  string // "before", "after", or "get"
		wantLen int
		wantPl  string // expected first plugin name
	}{
		{"posts before_create", "posts", "create", "before", 1, "seo"},
		{"posts after_create", "posts", "create", "after", 1, "cache"},
		{"pages before_update", "pages", "update", "before", 1, "validator"},
		{"pages after_delete", "pages", "delete", "after", 1, "audit"},
		{"posts before_update miss", "posts", "update", "before", 0, ""},
		{"pages after_create miss", "pages", "create", "after", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var entries []PipelineEntry
			switch tt.method {
			case "before":
				entries = r.Before(tt.table, tt.op)
			case "after":
				entries = r.After(tt.table, tt.op)
			case "get":
				entries = r.Get(tt.table, tt.op)
			}

			if tt.wantLen == 0 {
				if entries != nil {
					t.Errorf("expected nil, got %d entries", len(entries))
				}
				return
			}

			if len(entries) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(entries), tt.wantLen)
			}
			if entries[0].PluginName != tt.wantPl {
				t.Errorf("PluginName = %q, want %q", entries[0].PluginName, tt.wantPl)
			}
		})
	}
}

func TestParseConfigJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantKey string
		wantVal any
	}{
		{"empty string", "", true, "", nil},
		{"empty object", "{}", true, "", nil},
		{"valid json", `{"timeout":30}`, false, "timeout", float64(30)},
		{"nested json", `{"db":{"host":"localhost"}}`, false, "db", map[string]any{"host": "localhost"}},
		{"string value", `{"mode":"strict"}`, false, "mode", "strict"},
		{"invalid json", `{not valid json}`, true, "", nil},
		{"array not object", `[1,2,3]`, true, "", nil},
		{"bare string", `hello`, true, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseConfigJSON(tt.input)

			if tt.wantNil {
				if got != nil {
					t.Errorf("parseConfigJSON(%q) = %v, want nil", tt.input, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("parseConfigJSON(%q) = nil, want non-nil", tt.input)
			}

			val, ok := got[tt.wantKey]
			if !ok {
				t.Fatalf("result missing key %q, got %v", tt.wantKey, got)
			}

			// For nested maps, compare via sprint since reflect.DeepEqual
			// works but is heavier than needed for test diagnostics.
			switch expected := tt.wantVal.(type) {
			case map[string]any:
				gotMap, ok := val.(map[string]any)
				if !ok {
					t.Errorf("value for key %q is %T, want map[string]any", tt.wantKey, val)
					return
				}
				for k, v := range expected {
					if gotMap[k] != v {
						t.Errorf("nested key %q = %v, want %v", k, gotMap[k], v)
					}
				}
			default:
				if val != tt.wantVal {
					t.Errorf("value for key %q = %v (%T), want %v (%T)", tt.wantKey, val, val, tt.wantVal, tt.wantVal)
				}
			}
		})
	}
}

func TestConfigPropagation(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true, Config: `{"max_title":60}`},
		{PipelineID: "p2", TableName: "posts", Operation: "before_create", PluginName: "validate", Priority: 20, Enabled: true, Config: ""},
		{PipelineID: "p3", TableName: "posts", Operation: "before_create", PluginName: "log", Priority: 30, Enabled: true, Config: "{}"},
	})

	entries := r.Before("posts", "create")
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}

	// First entry has config.
	if entries[0].Config == nil || entries[0].Config["max_title"] != float64(60) {
		t.Errorf("entries[0].Config = %v, want map with max_title=60", entries[0].Config)
	}

	// Second and third entries have nil config (empty string and "{}").
	if entries[1].Config != nil {
		t.Errorf("entries[1].Config = %v, want nil for empty config", entries[1].Config)
	}
	if entries[2].Config != nil {
		t.Errorf("entries[2].Config = %v, want nil for {} config", entries[2].Config)
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true},
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Concurrent reads should not race with each other.
	for range goroutines {
		go func() {
			defer wg.Done()
			r.Before("posts", "create")
		}()
		go func() {
			defer wg.Done()
			r.After("posts", "create")
		}()
		go func() {
			defer wg.Done()
			r.IsEmpty()
		}()
	}

	wg.Wait()
}

func TestConcurrentReadAndReload(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()

	rows1 := []PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true},
	}
	rows2 := []PipelineRow{
		{PipelineID: "p2", TableName: "pages", Operation: "after_update", PluginName: "cache", Priority: 5, Enabled: true},
	}

	r.Build(rows1)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines + 1)

	// One writer doing Reload.
	go func() {
		defer wg.Done()
		for range goroutines {
			r.Reload(rows2)
			r.Reload(rows1)
		}
	}()

	// Many concurrent readers. No panics or data races is the success criterion.
	for range goroutines {
		go func() {
			defer wg.Done()
			r.Before("posts", "create")
			r.After("pages", "update")
			r.IsEmpty()
			r.Get("posts", "before_create")
		}()
	}

	wg.Wait()
}

func TestBefore_ConstructsCorrectKey(t *testing.T) {
	t.Parallel()

	// Verify Before("mytable", "update") looks up "mytable.before_update"
	// by building with that exact operation string.
	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "mytable", Operation: "before_update", PluginName: "v", Enabled: true},
	})

	if entries := r.Before("mytable", "update"); len(entries) != 1 {
		t.Errorf("Before(mytable, update) len = %d, want 1", len(entries))
	}

	// Get with full operation string should also work.
	if entries := r.Get("mytable", "before_update"); len(entries) != 1 {
		t.Errorf("Get(mytable, before_update) len = %d, want 1", len(entries))
	}
}

func TestAfter_ConstructsCorrectKey(t *testing.T) {
	t.Parallel()

	// Verify After("mytable", "delete") looks up "mytable.after_delete".
	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "mytable", Operation: "after_delete", PluginName: "a", Enabled: true},
	})

	if entries := r.After("mytable", "delete"); len(entries) != 1 {
		t.Errorf("After(mytable, delete) len = %d, want 1", len(entries))
	}

	// Get with full operation string should also work.
	if entries := r.Get("mytable", "after_delete"); len(entries) != 1 {
		t.Errorf("Get(mytable, after_delete) len = %d, want 1", len(entries))
	}
}

func TestIsEmpty_FalseAfterBuild(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "t", Operation: "before_create", PluginName: "x", Enabled: true},
	})

	if r.IsEmpty() {
		t.Error("IsEmpty() = true after Build with enabled row, want false")
	}
}

func TestIsEmpty_TrueAfterReloadWithEmpty(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "t", Operation: "before_create", PluginName: "x", Enabled: true},
	})

	r.Reload(nil)

	if !r.IsEmpty() {
		t.Error("IsEmpty() = false after Reload(nil), want true")
	}
}

// --- DryRun / ListKeys tests ---

func TestListKeys_Empty(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	keys := r.ListKeys()
	if len(keys) != 0 {
		t.Errorf("ListKeys on empty registry = %v, want empty", keys)
	}
}

func TestListKeys_Sorted(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Enabled: true},
		{PipelineID: "p2", TableName: "media", Operation: "after_delete", PluginName: "audit", Enabled: true},
		{PipelineID: "p3", TableName: "posts", Operation: "after_create", PluginName: "cache", Enabled: true},
	})

	keys := r.ListKeys()
	if len(keys) != 3 {
		t.Fatalf("ListKeys len = %d, want 3", len(keys))
	}

	wantKeys := []string{"media.after_delete", "posts.after_create", "posts.before_create"}
	for i, want := range wantKeys {
		if keys[i] != want {
			t.Errorf("keys[%d] = %q, want %q", i, keys[i], want)
		}
	}
}

func TestDryRun_BeforeAndAfter(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true},
		{PipelineID: "p2", TableName: "posts", Operation: "after_create", PluginName: "cache", Priority: 5, Enabled: true},
	})

	results := r.DryRun("posts", "create")
	if len(results) != 2 {
		t.Fatalf("DryRun results len = %d, want 2", len(results))
	}

	// Before should come first.
	if results[0].Phase != "before" {
		t.Errorf("results[0].Phase = %q, want 'before'", results[0].Phase)
	}
	if results[0].Table != "posts" {
		t.Errorf("results[0].Table = %q, want 'posts'", results[0].Table)
	}
	if results[0].Operation != "create" {
		t.Errorf("results[0].Operation = %q, want 'create'", results[0].Operation)
	}
	if len(results[0].Entries) != 1 {
		t.Errorf("results[0].Entries len = %d, want 1", len(results[0].Entries))
	}

	if results[1].Phase != "after" {
		t.Errorf("results[1].Phase = %q, want 'after'", results[1].Phase)
	}
	if len(results[1].Entries) != 1 {
		t.Errorf("results[1].Entries len = %d, want 1", len(results[1].Entries))
	}
}

func TestDryRun_NoMatch(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Enabled: true},
	})

	results := r.DryRun("pages", "delete")
	if len(results) != 0 {
		t.Errorf("DryRun for non-existent table/op = %d results, want 0", len(results))
	}
}

func TestDryRun_OnlyBefore(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Enabled: true},
	})

	results := r.DryRun("posts", "create")
	if len(results) != 1 {
		t.Fatalf("DryRun results len = %d, want 1", len(results))
	}
	if results[0].Phase != "before" {
		t.Errorf("Phase = %q, want 'before'", results[0].Phase)
	}
}

func TestDryRun_OnlyAfter(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "after_create", PluginName: "cache", Enabled: true},
	})

	results := r.DryRun("posts", "create")
	if len(results) != 1 {
		t.Fatalf("DryRun results len = %d, want 1", len(results))
	}
	if results[0].Phase != "after" {
		t.Errorf("Phase = %q, want 'after'", results[0].Phase)
	}
}

func TestDryRunAll_MultipleChains(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "posts", Operation: "before_create", PluginName: "seo", Priority: 10, Enabled: true},
		{PipelineID: "p2", TableName: "posts", Operation: "after_create", PluginName: "cache", Priority: 5, Enabled: true},
		{PipelineID: "p3", TableName: "media", Operation: "before_delete", PluginName: "audit", Priority: 1, Enabled: true},
	})

	results := r.DryRunAll()
	if len(results) != 3 {
		t.Fatalf("DryRunAll results len = %d, want 3", len(results))
	}

	// Should be sorted by key: media.before_delete, posts.after_create, posts.before_create
	expected := []struct {
		table string
		op    string
		phase string
	}{
		{"media", "delete", "before"},
		{"posts", "create", "after"},
		{"posts", "create", "before"},
	}

	for i, want := range expected {
		if results[i].Table != want.table {
			t.Errorf("results[%d].Table = %q, want %q", i, results[i].Table, want.table)
		}
		if results[i].Operation != want.op {
			t.Errorf("results[%d].Operation = %q, want %q", i, results[i].Operation, want.op)
		}
		if results[i].Phase != want.phase {
			t.Errorf("results[%d].Phase = %q, want %q", i, results[i].Phase, want.phase)
		}
	}
}

func TestDryRunAll_Empty(t *testing.T) {
	t.Parallel()

	r := NewPipelineRegistry()
	results := r.DryRunAll()
	if len(results) != 0 {
		t.Errorf("DryRunAll on empty registry = %d results, want 0", len(results))
	}
}

func TestDryRunAll_DotsInTableName(t *testing.T) {
	t.Parallel()

	// Table names with dots should be handled correctly via LastIndex.
	r := NewPipelineRegistry()
	r.Build([]PipelineRow{
		{PipelineID: "p1", TableName: "schema.posts", Operation: "before_create", PluginName: "seo", Enabled: true},
	})

	results := r.DryRunAll()
	if len(results) != 1 {
		t.Fatalf("DryRunAll len = %d, want 1", len(results))
	}
	if results[0].Table != "schema.posts" {
		t.Errorf("Table = %q, want 'schema.posts'", results[0].Table)
	}
	if results[0].Operation != "create" {
		t.Errorf("Operation = %q, want 'create'", results[0].Operation)
	}
	if results[0].Phase != "before" {
		t.Errorf("Phase = %q, want 'before'", results[0].Phase)
	}
}
