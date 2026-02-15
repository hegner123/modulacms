package config_test

import (
	"encoding/json"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// ---------------------------------------------------------------------------
// DefaultKeyMap
// ---------------------------------------------------------------------------

func TestDefaultKeyMap_AllActionsBound(t *testing.T) {
	t.Parallel()

	// Every exported Action constant must appear in the default key map
	// with at least one binding.
	actions := []config.Action{
		config.ActionQuit,
		config.ActionDismiss,
		config.ActionUp,
		config.ActionDown,
		config.ActionBack,
		config.ActionSelect,
		config.ActionNextPanel,
		config.ActionPrevPanel,
		config.ActionNew,
		config.ActionEdit,
		config.ActionDelete,
		config.ActionMove,
		config.ActionTitlePrev,
		config.ActionTitleNext,
		config.ActionPagePrev,
		config.ActionPageNext,
		config.ActionExpand,
		config.ActionCollapse,
		config.ActionReorderUp,
		config.ActionReorderDown,
		config.ActionCopy,
		config.ActionPublish,
		config.ActionArchive,
		config.ActionGoParent,
		config.ActionGoChild,
	}

	km := config.DefaultKeyMap()

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			t.Parallel()
			keys := km[action]
			if len(keys) == 0 {
				t.Errorf("DefaultKeyMap() has no bindings for action %q", action)
			}
		})
	}
}

func TestDefaultKeyMap_SpotCheckBindings(t *testing.T) {
	t.Parallel()

	// Verify specific keys that TUI handlers depend on.
	km := config.DefaultKeyMap()

	tests := []struct {
		name   string
		action config.Action
		keys   []string
	}{
		{name: "quit includes q and ctrl+c", action: config.ActionQuit, keys: []string{"q", "ctrl+c"}},
		{name: "up includes up and k", action: config.ActionUp, keys: []string{"up", "k"}},
		{name: "select includes enter", action: config.ActionSelect, keys: []string{"enter", "l", "right"}},
		{name: "dismiss is esc", action: config.ActionDismiss, keys: []string{"esc"}},
		{name: "expand includes + and =", action: config.ActionExpand, keys: []string{"+", "="}},
		{name: "reorder up uses shift+up and K", action: config.ActionReorderUp, keys: []string{"shift+up", "K"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := km[tt.action]
			if len(got) != len(tt.keys) {
				t.Fatalf("action %q: got %d keys %v, want %d keys %v",
					tt.action, len(got), got, len(tt.keys), tt.keys)
			}
			for i, k := range tt.keys {
				if got[i] != k {
					t.Errorf("action %q key[%d] = %q, want %q", tt.action, i, got[i], k)
				}
			}
		})
	}
}

func TestDefaultKeyMap_ReturnsNewInstance(t *testing.T) {
	t.Parallel()

	// Each call should return a distinct map so callers cannot mutate the
	// "default" keybindings for other callers.
	km1 := config.DefaultKeyMap()
	km2 := config.DefaultKeyMap()

	km1[config.ActionQuit] = []string{"x"}
	if len(km2[config.ActionQuit]) == 1 && km2[config.ActionQuit][0] == "x" {
		t.Error("DefaultKeyMap() returns the same map instance -- mutations leak")
	}
}

// ---------------------------------------------------------------------------
// KeyMap.Matches
// ---------------------------------------------------------------------------

func TestKeyMap_Matches(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{
		config.ActionQuit:   {"q", "ctrl+c"},
		config.ActionSelect: {"enter"},
	}

	tests := []struct {
		name   string
		key    string
		action config.Action
		want   bool
	}{
		{name: "exact match first key", key: "q", action: config.ActionQuit, want: true},
		{name: "exact match second key", key: "ctrl+c", action: config.ActionQuit, want: true},
		{name: "no match wrong key", key: "x", action: config.ActionQuit, want: false},
		{name: "no match wrong action", key: "q", action: config.ActionSelect, want: false},
		{name: "action not in map", key: "q", action: config.ActionUp, want: false},
		{name: "empty key never matches", key: "", action: config.ActionQuit, want: false},
		{name: "match single-binding action", key: "enter", action: config.ActionSelect, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := km.Matches(tt.key, tt.action)
			if got != tt.want {
				t.Errorf("Matches(%q, %q) = %v, want %v", tt.key, tt.action, got, tt.want)
			}
		})
	}
}

func TestKeyMap_Matches_EmptyMap(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{}
	if km.Matches("q", config.ActionQuit) {
		t.Error("Matches on empty KeyMap should return false")
	}
}

func TestKeyMap_Matches_NilMap(t *testing.T) {
	t.Parallel()

	var km config.KeyMap
	if km.Matches("q", config.ActionQuit) {
		t.Error("Matches on nil KeyMap should return false")
	}
}

func TestKeyMap_Matches_EmptyKeySlice(t *testing.T) {
	t.Parallel()

	// Action exists in map but has no keys bound.
	km := config.KeyMap{
		config.ActionQuit: {},
	}
	if km.Matches("q", config.ActionQuit) {
		t.Error("Matches should return false when action has empty key slice")
	}
}

// ---------------------------------------------------------------------------
// KeyMap.Merge
// ---------------------------------------------------------------------------

func TestKeyMap_Merge_OverridesExisting(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{
		config.ActionQuit:   {"q", "ctrl+c"},
		config.ActionSelect: {"enter"},
	}

	overrides := config.KeyMap{
		config.ActionQuit: {"x"},
	}

	km.Merge(overrides)

	// Quit should now have only "x"
	if len(km[config.ActionQuit]) != 1 || km[config.ActionQuit][0] != "x" {
		t.Errorf("after Merge, ActionQuit = %v, want [x]", km[config.ActionQuit])
	}

	// Select should be untouched
	if len(km[config.ActionSelect]) != 1 || km[config.ActionSelect][0] != "enter" {
		t.Errorf("after Merge, ActionSelect = %v, want [enter]", km[config.ActionSelect])
	}
}

func TestKeyMap_Merge_AddsNewAction(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{
		config.ActionQuit: {"q"},
	}

	overrides := config.KeyMap{
		config.ActionNew: {"n", "N"},
	}

	km.Merge(overrides)

	if len(km[config.ActionNew]) != 2 {
		t.Fatalf("after Merge, ActionNew = %v, want [n N]", km[config.ActionNew])
	}
	if km[config.ActionNew][0] != "n" || km[config.ActionNew][1] != "N" {
		t.Errorf("after Merge, ActionNew = %v, want [n N]", km[config.ActionNew])
	}
}

func TestKeyMap_Merge_EmptyOverrides(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{
		config.ActionQuit: {"q"},
	}

	km.Merge(config.KeyMap{})

	if len(km[config.ActionQuit]) != 1 || km[config.ActionQuit][0] != "q" {
		t.Errorf("Merge with empty overrides changed map: %v", km)
	}
}

func TestKeyMap_Merge_NilOverrides(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{
		config.ActionQuit: {"q"},
	}

	km.Merge(nil)

	if len(km[config.ActionQuit]) != 1 || km[config.ActionQuit][0] != "q" {
		t.Errorf("Merge with nil overrides changed map: %v", km)
	}
}

// ---------------------------------------------------------------------------
// KeyMap.HintString
// ---------------------------------------------------------------------------

func TestKeyMap_HintString(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{
		config.ActionQuit:   {"q", "ctrl+c"},
		config.ActionSelect: {"enter"},
	}

	tests := []struct {
		name   string
		action config.Action
		want   string
	}{
		{name: "returns first key when multiple bound", action: config.ActionQuit, want: "q"},
		{name: "returns only key when single bound", action: config.ActionSelect, want: "enter"},
		{name: "returns ? for unbound action", action: config.ActionUp, want: "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := km.HintString(tt.action)
			if got != tt.want {
				t.Errorf("HintString(%q) = %q, want %q", tt.action, got, tt.want)
			}
		})
	}
}

func TestKeyMap_HintString_EmptyKeySlice(t *testing.T) {
	t.Parallel()

	km := config.KeyMap{
		config.ActionQuit: {},
	}

	got := km.HintString(config.ActionQuit)
	if got != "?" {
		t.Errorf("HintString for action with empty keys = %q, want %q", got, "?")
	}
}

func TestKeyMap_HintString_NilMap(t *testing.T) {
	t.Parallel()

	var km config.KeyMap
	got := km.HintString(config.ActionQuit)
	if got != "?" {
		t.Errorf("HintString on nil KeyMap = %q, want %q", got, "?")
	}
}

// ---------------------------------------------------------------------------
// KeyMap JSON round-trip
// ---------------------------------------------------------------------------

func TestKeyMap_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := config.KeyMap{
		config.ActionQuit:   {"q", "ctrl+c"},
		config.ActionSelect: {"enter"},
		config.ActionUp:     {"up", "k"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var decoded config.KeyMap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	// Verify all actions survived the round trip
	for action, wantKeys := range original {
		gotKeys := decoded[action]
		if len(gotKeys) != len(wantKeys) {
			t.Errorf("action %q: got %d keys, want %d", action, len(gotKeys), len(wantKeys))
			continue
		}
		for i, k := range wantKeys {
			if gotKeys[i] != k {
				t.Errorf("action %q key[%d] = %q, want %q", action, i, gotKeys[i], k)
			}
		}
	}

	// Verify no extra actions appeared
	if len(decoded) != len(original) {
		t.Errorf("decoded map has %d entries, want %d", len(decoded), len(original))
	}
}

func TestKeyMap_JSONRoundTrip_EmptyMap(t *testing.T) {
	t.Parallel()

	original := config.KeyMap{}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var decoded config.KeyMap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("decoded empty KeyMap has %d entries, want 0", len(decoded))
	}
}

func TestKeyMap_UnmarshalJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	var km config.KeyMap
	err := json.Unmarshal([]byte(`{not valid`), &km)
	if err == nil {
		t.Fatal("UnmarshalJSON should return error for invalid JSON")
	}
}

func TestKeyMap_UnmarshalJSON_WrongType(t *testing.T) {
	t.Parallel()

	// JSON is valid but not the expected shape (string instead of object)
	var km config.KeyMap
	err := json.Unmarshal([]byte(`"not an object"`), &km)
	if err == nil {
		t.Fatal("UnmarshalJSON should return error for non-object JSON")
	}
}

func TestKeyMap_MarshalJSON_ProducesValidJSON(t *testing.T) {
	t.Parallel()

	km := config.DefaultKeyMap()
	data, err := json.Marshal(km)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	if !json.Valid(data) {
		t.Error("MarshalJSON output is not valid JSON")
	}
}

func TestKeyMap_JSONRoundTrip_DefaultKeyMap(t *testing.T) {
	t.Parallel()

	original := config.DefaultKeyMap()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var decoded config.KeyMap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if len(decoded) != len(original) {
		t.Fatalf("decoded has %d actions, want %d", len(decoded), len(original))
	}

	// Every action in the default map should survive the round trip with
	// identical bindings.
	for action, wantKeys := range original {
		gotKeys := decoded[action]
		if len(gotKeys) != len(wantKeys) {
			t.Errorf("action %q: got %d keys, want %d", action, len(gotKeys), len(wantKeys))
			continue
		}
		for i, k := range wantKeys {
			if gotKeys[i] != k {
				t.Errorf("action %q key[%d] = %q, want %q", action, i, gotKeys[i], k)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Action constants
// ---------------------------------------------------------------------------

func TestActionConstants_StringValues(t *testing.T) {
	t.Parallel()

	// These string values are used as JSON keys and in keybinding configs.
	// Changing them is a breaking change for user config files.
	tests := []struct {
		name string
		got  config.Action
		want string
	}{
		{name: "quit", got: config.ActionQuit, want: "quit"},
		{name: "dismiss", got: config.ActionDismiss, want: "dismiss"},
		{name: "up", got: config.ActionUp, want: "up"},
		{name: "down", got: config.ActionDown, want: "down"},
		{name: "back", got: config.ActionBack, want: "back"},
		{name: "select", got: config.ActionSelect, want: "select"},
		{name: "next_panel", got: config.ActionNextPanel, want: "next_panel"},
		{name: "prev_panel", got: config.ActionPrevPanel, want: "prev_panel"},
		{name: "new", got: config.ActionNew, want: "new"},
		{name: "edit", got: config.ActionEdit, want: "edit"},
		{name: "delete", got: config.ActionDelete, want: "delete"},
		{name: "move", got: config.ActionMove, want: "move"},
		{name: "title_prev", got: config.ActionTitlePrev, want: "title_prev"},
		{name: "title_next", got: config.ActionTitleNext, want: "title_next"},
		{name: "page_prev", got: config.ActionPagePrev, want: "page_prev"},
		{name: "page_next", got: config.ActionPageNext, want: "page_next"},
		{name: "expand", got: config.ActionExpand, want: "expand"},
		{name: "collapse", got: config.ActionCollapse, want: "collapse"},
		{name: "reorder_up", got: config.ActionReorderUp, want: "reorder_up"},
		{name: "reorder_down", got: config.ActionReorderDown, want: "reorder_down"},
		{name: "copy", got: config.ActionCopy, want: "copy"},
		{name: "publish", got: config.ActionPublish, want: "publish"},
		{name: "archive", got: config.ActionArchive, want: "archive"},
		{name: "go_parent", got: config.ActionGoParent, want: "go_parent"},
		{name: "go_child", got: config.ActionGoChild, want: "go_child"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.got) != tt.want {
				t.Errorf("Action constant = %q, want %q", tt.got, tt.want)
			}
		})
	}
}
