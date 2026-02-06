package config

import "encoding/json"

// Action represents a semantic keybinding action in the TUI.
type Action string

const (
	ActionQuit      Action = "quit"
	ActionDismiss   Action = "dismiss"
	ActionUp        Action = "up"
	ActionDown      Action = "down"
	ActionBack      Action = "back"
	ActionSelect    Action = "select"
	ActionNextPanel Action = "next_panel"
	ActionPrevPanel Action = "prev_panel"
	ActionNew       Action = "new"
	ActionEdit      Action = "edit"
	ActionDelete    Action = "delete"
	ActionMove      Action = "move"
	ActionTitlePrev Action = "title_prev"
	ActionTitleNext Action = "title_next"
	ActionPagePrev  Action = "page_prev"
	ActionPageNext  Action = "page_next"
	ActionExpand      Action = "expand"
	ActionCollapse    Action = "collapse"
	ActionReorderUp   Action = "reorder_up"
	ActionReorderDown Action = "reorder_down"
	ActionCopy        Action = "copy"
	ActionPublish     Action = "publish"
	ActionArchive     Action = "archive"
	ActionGoParent    Action = "go_parent"
	ActionGoChild     Action = "go_child"
)

// KeyMap maps semantic actions to one or more key strings (as reported by
// bubbletea's KeyMsg.String()).
type KeyMap map[Action][]string

// DefaultKeyMap returns the built-in keybindings that match the original
// hardcoded values in the control handlers.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		ActionQuit:      {"q", "ctrl+c"},
		ActionDismiss:   {"esc"},
		ActionUp:        {"up", "k"},
		ActionDown:      {"down", "j"},
		ActionBack:      {"h", "left", "backspace"},
		ActionSelect:    {"enter", "l", "right"},
		ActionNextPanel: {"tab"},
		ActionPrevPanel: {"shift+tab"},
		ActionNew:       {"n"},
		ActionEdit:      {"e"},
		ActionDelete:    {"d"},
		ActionMove:      {"m"},
		ActionTitlePrev: {"shift+left"},
		ActionTitleNext: {"shift+right"},
		ActionPagePrev:  {"left"},
		ActionPageNext:  {"right"},
		ActionExpand:      {"+", "="},
		ActionCollapse:    {"-", "_"},
		ActionReorderUp:   {"shift+up", "K"},
		ActionReorderDown: {"shift+down", "J"},
		ActionCopy:        {"c"},
		ActionPublish:     {"p"},
		ActionArchive:     {"a"},
		ActionGoParent:    {"g"},
		ActionGoChild:     {"G"},
	}
}

// Matches returns true when key is bound to the given action.
func (km KeyMap) Matches(key string, action Action) bool {
	for _, k := range km[action] {
		if k == key {
			return true
		}
	}
	return false
}

// Merge replaces bindings in km with those from overrides. Actions not
// present in overrides keep their current bindings.
func (km KeyMap) Merge(overrides KeyMap) {
	for action, keys := range overrides {
		km[action] = keys
	}
}

// HintString returns the first key bound to action, suitable for display
// in the status bar. Returns "?" if the action has no bindings.
func (km KeyMap) HintString(action Action) string {
	keys := km[action]
	if len(keys) == 0 {
		return "?"
	}
	return keys[0]
}

// MarshalJSON implements json.Marshaler so that only user-specified
// overrides are written to JSON (the full map is serialised as-is).
func (km KeyMap) MarshalJSON() ([]byte, error) {
	raw := make(map[string][]string, len(km))
	for action, keys := range km {
		raw[string(action)] = keys
	}
	return json.Marshal(raw)
}

// UnmarshalJSON implements json.Unmarshaler, reading string-keyed maps
// into the Action-keyed KeyMap.
func (km *KeyMap) UnmarshalJSON(data []byte) error {
	var raw map[string][]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*km = make(KeyMap, len(raw))
	for action, keys := range raw {
		(*km)[Action(action)] = keys
	}
	return nil
}
