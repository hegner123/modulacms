package types

import (
	"fmt"
	"strings"
)

// DatatypeType represents the type classification of a datatype.
// Values prefixed with underscore are engine-reserved and trigger
// built-in behavior. All other values are user-defined pass-through.
type DatatypeType string

const (
	DatatypeTypeRoot       DatatypeType = "_root"
	DatatypeTypeReference  DatatypeType = "_reference"
	DatatypeTypeNestedRoot DatatypeType = "_nested_root"
	DatatypeTypeSystemLog  DatatypeType = "_system_log"
	DatatypeTypeCollection DatatypeType = "_collection"
	DatatypeTypeGlobal     DatatypeType = "_global"
	DatatypeTypePlugin     DatatypeType = "_plugin"
)

const pluginTypePrefix = "_plugin_"

// reservedTypes maps each reserved type to a description of its engine behavior.
var reservedTypes = map[DatatypeType]string{
	DatatypeTypeRoot:       "Tree entry point, one per route",
	DatatypeTypeReference:  "Triggers tree composition — resolves _id field values, attaches referenced trees as children",
	DatatypeTypeNestedRoot: "Root of a composed subtree, assigned by the engine during tree composition",
	DatatypeTypeSystemLog:  "Synthetic node injected when a reference cannot be resolved",
	DatatypeTypeCollection: "Marks content as a queryable collection; signals to clients that children support filtering",
	DatatypeTypeGlobal:     "Singleton site-wide content (menus, footers, settings); no route association, delivered via /globals endpoint",
	DatatypeTypePlugin:     "Plugin-provided content; actual types use _plugin_{name} namespace registered by plugin OnInit",
}

// IsReserved returns true if the type is engine-reserved.
func (t DatatypeType) IsReserved() bool {
	_, ok := reservedTypes[t]
	return ok
}

// IsReservedPrefix returns true if the string starts with underscore,
// which is the reserved namespace. Used to reject user-created types
// that start with underscore even if not currently in the registry.
func IsReservedPrefix(t string) bool {
	return len(t) > 0 && t[0] == '_'
}

// IsRootType returns true if the type identifies a tree root node.
// Used by the tree-building algorithm to find root nodes.
func (t DatatypeType) IsRootType() bool {
	return t == DatatypeTypeRoot || t == DatatypeTypeNestedRoot
}

// IsGlobalType returns true if the type is _global.
func (t DatatypeType) IsGlobalType() bool {
	return t == DatatypeTypeGlobal
}

// IsPluginType returns true if the type uses the _plugin_ namespace
// (e.g., "_plugin_analytics", "_plugin_seo"). The base "_plugin" type
// is the registry sentinel; actual plugin types use "_plugin_{name}".
func (t DatatypeType) IsPluginType() bool {
	return strings.HasPrefix(string(t), pluginTypePrefix)
}

// PluginName extracts the plugin name from a _plugin_{name} type.
// Returns empty string if not a plugin type.
func (t DatatypeType) PluginName() string {
	if !t.IsPluginType() {
		return ""
	}
	return string(t)[len(pluginTypePrefix):]
}

// PluginDatatypeType returns the _plugin_{name} type for a given plugin name.
func PluginDatatypeType(pluginName string) DatatypeType {
	return DatatypeType(pluginTypePrefix + pluginName)
}

// String returns the string representation of DatatypeType.
func (t DatatypeType) String() string {
	return string(t)
}

// ReservedTypes returns a copy of the registry for documentation/UI purposes.
func ReservedTypes() map[DatatypeType]string {
	out := make(map[DatatypeType]string, len(reservedTypes))
	for k, v := range reservedTypes {
		out[k] = v
	}
	return out
}

// ValidateDatatypeType validates a datatype type string.
// Rules:
//   - Empty string: error
//   - Starts with _ but not in registry: error (reserved prefix, unknown type)
//   - Starts with _ and in registry: valid reserved type
//   - Does not start with _: valid user type
func ValidateDatatypeType(t string) error {
	if t == "" {
		return fmt.Errorf("DatatypeType: cannot be empty")
	}
	if t[0] == '_' {
		dt := DatatypeType(t)
		// Allow _plugin_{name} types (validated by plugin system at load time)
		if dt.IsPluginType() {
			return nil
		}
		if !dt.IsReserved() {
			return fmt.Errorf("DatatypeType: %q uses reserved prefix '_' but is not a recognized engine type", t)
		}
	}
	return nil
}

// ValidateUserDatatypeType validates that a user-provided type is non-empty.
// Reserved-prefix types (starting with '_') are allowed — they enable system
// functionality like tree roots (_root) and nested roots (_nested_root).
func ValidateUserDatatypeType(t string) error {
	if t == "" {
		return fmt.Errorf("DatatypeType: cannot be empty")
	}
	return nil
}
