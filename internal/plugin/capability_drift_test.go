package plugin

import (
	"testing"
)

func TestCompareCapabilities_NoDrift(t *testing.T) {
	caps := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10},
		{Table: "content", Op: "after_create", Handler: "notify", Priority: 20},
	}
	result := compareCapabilities(caps, caps)
	if result != nil {
		t.Errorf("expected nil for identical capabilities, got %d entries", len(result))
	}
}

func TestCompareCapabilities_Added(t *testing.T) {
	current := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10},
		{Table: "content", Op: "after_create", Handler: "notify", Priority: 20},
	}
	stored := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10},
	}
	result := compareCapabilities(current, stored)
	if len(result) != 1 {
		t.Fatalf("expected 1 drift entry, got %d", len(result))
	}
	if result[0].Kind != "added" {
		t.Errorf("expected kind 'added', got %q", result[0].Kind)
	}
	if result[0].Current == nil {
		t.Error("expected Current to be non-nil for added entry")
	}
	if result[0].Previous != nil {
		t.Error("expected Previous to be nil for added entry")
	}
	if result[0].Current.Op != "after_create" {
		t.Errorf("expected added op 'after_create', got %q", result[0].Current.Op)
	}
}

func TestCompareCapabilities_Removed(t *testing.T) {
	current := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10},
	}
	stored := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10},
		{Table: "content", Op: "after_create", Handler: "notify", Priority: 20},
	}
	result := compareCapabilities(current, stored)
	if len(result) != 1 {
		t.Fatalf("expected 1 drift entry, got %d", len(result))
	}
	if result[0].Kind != "removed" {
		t.Errorf("expected kind 'removed', got %q", result[0].Kind)
	}
	if result[0].Current != nil {
		t.Error("expected Current to be nil for removed entry")
	}
	if result[0].Previous == nil {
		t.Error("expected Previous to be non-nil for removed entry")
	}
}

func TestCompareCapabilities_Changed(t *testing.T) {
	current := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate_v2", Priority: 5},
	}
	stored := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10},
	}
	result := compareCapabilities(current, stored)
	if len(result) != 1 {
		t.Fatalf("expected 1 drift entry, got %d", len(result))
	}
	if result[0].Kind != "changed" {
		t.Errorf("expected kind 'changed', got %q", result[0].Kind)
	}
	if result[0].Current == nil || result[0].Previous == nil {
		t.Fatal("expected both Current and Previous to be non-nil for changed entry")
	}
	if result[0].Current.Handler != "validate_v2" {
		t.Errorf("expected current handler 'validate_v2', got %q", result[0].Current.Handler)
	}
	if result[0].Previous.Handler != "validate" {
		t.Errorf("expected previous handler 'validate', got %q", result[0].Previous.Handler)
	}
}

func TestCompareCapabilities_Mixed(t *testing.T) {
	current := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate_v2", Priority: 5}, // changed
		{Table: "media", Op: "before_create", Handler: "resize", Priority: 10},       // added
	}
	stored := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10}, // changed
		{Table: "content", Op: "after_delete", Handler: "cleanup", Priority: 100},  // removed
	}
	result := compareCapabilities(current, stored)
	if len(result) != 3 {
		t.Fatalf("expected 3 drift entries, got %d", len(result))
	}

	counts := map[string]int{}
	for _, d := range result {
		counts[d.Kind]++
	}
	if counts["added"] != 1 {
		t.Errorf("expected 1 added, got %d", counts["added"])
	}
	if counts["removed"] != 1 {
		t.Errorf("expected 1 removed, got %d", counts["removed"])
	}
	if counts["changed"] != 1 {
		t.Errorf("expected 1 changed, got %d", counts["changed"])
	}
}

func TestCompareCapabilities_BothEmpty(t *testing.T) {
	result := compareCapabilities(nil, nil)
	if result != nil {
		t.Errorf("expected nil for both empty, got %d entries", len(result))
	}

	result = compareCapabilities([]PluginCapability{}, []PluginCapability{})
	if result != nil {
		t.Errorf("expected nil for both empty slices, got %d entries", len(result))
	}
}

func TestCompareCapabilities_OneEmpty(t *testing.T) {
	caps := []PluginCapability{
		{Table: "content", Op: "before_create", Handler: "validate", Priority: 10},
	}

	// All added.
	result := compareCapabilities(caps, nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 drift entry for all added, got %d", len(result))
	}
	if result[0].Kind != "added" {
		t.Errorf("expected 'added', got %q", result[0].Kind)
	}

	// All removed.
	result = compareCapabilities(nil, caps)
	if len(result) != 1 {
		t.Fatalf("expected 1 drift entry for all removed, got %d", len(result))
	}
	if result[0].Kind != "removed" {
		t.Errorf("expected 'removed', got %q", result[0].Kind)
	}
}

func TestCompareCapabilities_NilVsEmptySlice(t *testing.T) {
	// nil vs empty slice should be identical (no drift).
	result := compareCapabilities(nil, []PluginCapability{})
	if result != nil {
		t.Errorf("expected nil for nil vs empty, got %d entries", len(result))
	}

	result = compareCapabilities([]PluginCapability{}, nil)
	if result != nil {
		t.Errorf("expected nil for empty vs nil, got %d entries", len(result))
	}
}
