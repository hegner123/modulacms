package types

import (
	"fmt"
	"strings"
)

// DatatypeType represents the type classification of a datatype.
// Values prefixed with underscore are engine-reserved and trigger
// built-in behavior. All other values are user-defined pass-through.
//
// Reserved types support optional suffixes separated by underscore:
// "_reference" and "_reference_menu" both trigger reference behavior.
// The suffix is metadata for the admin panel (e.g. filtering _id field
// dropdowns) and does not change engine behavior.
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

// reservedBases lists the base reserved type strings that support suffixes.
// The suffix separator is underscore: _global_menu has base "_global" and suffix "menu".
var reservedBases = []string{
	string(DatatypeTypeRoot),
	string(DatatypeTypeReference),
	string(DatatypeTypeNestedRoot),
	string(DatatypeTypeSystemLog),
	string(DatatypeTypeCollection),
	string(DatatypeTypeGlobal),
	string(DatatypeTypePlugin),
}

// reservedTypes maps each reserved base type to a description of its engine behavior.
var reservedTypes = map[DatatypeType]string{
	DatatypeTypeRoot:       "Tree entry point, one per route",
	DatatypeTypeReference:  "Triggers tree composition — resolves _id field values, attaches referenced trees as children",
	DatatypeTypeNestedRoot: "Root of a composed subtree, assigned by the engine during tree composition",
	DatatypeTypeSystemLog:  "Synthetic node injected when a reference cannot be resolved",
	DatatypeTypeCollection: "Marks content as a queryable collection; signals to clients that children support filtering",
	DatatypeTypeGlobal:     "Singleton site-wide content (menus, footers, settings); no route association, delivered via /globals endpoint",
	DatatypeTypePlugin:     "Plugin-provided content; actual types use _plugin_{name} namespace registered by plugin OnInit",
}

// BaseType returns the engine base of a DatatypeType by stripping any suffix.
// "_reference_menu" -> "_reference", "_global" -> "_global", "page" -> "page".
func (t DatatypeType) BaseType() DatatypeType {
	s := string(t)
	if len(s) == 0 || s[0] != '_' {
		return t
	}
	for _, base := range reservedBases {
		if s == base {
			return DatatypeType(base)
		}
		if len(s) > len(base) && s[:len(base)] == base && s[len(base)] == '_' {
			return DatatypeType(base)
		}
	}
	return t
}

// Suffix returns the suffix portion after the base reserved type.
// "_reference_menu" -> "menu", "_global" -> "", "page" -> "".
func (t DatatypeType) Suffix() string {
	s := string(t)
	base := string(t.BaseType())
	if len(s) > len(base)+1 && s[len(base)] == '_' {
		return s[len(base)+1:]
	}
	return ""
}

// IsReserved returns true if the type matches a known reserved base,
// with or without a suffix. "_reference_menu" is reserved.
func (t DatatypeType) IsReserved() bool {
	base := t.BaseType()
	_, ok := reservedTypes[base]
	return ok
}

// IsReservedPrefix returns true if the string starts with underscore,
// which is the reserved namespace. Used to reject user-created types
// that start with underscore even if not currently in the registry.
func IsReservedPrefix(t string) bool {
	return len(t) > 0 && t[0] == '_'
}

// IsRootType returns true if the type identifies a tree root node.
// Matches _root and _nested_root, with or without suffixes.
func (t DatatypeType) IsRootType() bool {
	base := t.BaseType()
	return base == DatatypeTypeRoot || base == DatatypeTypeNestedRoot
}

// IsReferenceType returns true if the base type is _reference.
// Matches _reference, _reference_menu, etc.
func (t DatatypeType) IsReferenceType() bool {
	return t.BaseType() == DatatypeTypeReference
}

// IsCollectionType returns true if the base type is _collection.
func (t DatatypeType) IsCollectionType() bool {
	return t.BaseType() == DatatypeTypeCollection
}

// IsGlobalType returns true if the base type is _global.
// Matches _global, _global_menu, etc.
func (t DatatypeType) IsGlobalType() bool {
	return t.BaseType() == DatatypeTypeGlobal
}

// IsSystemLogType returns true if the base type is _system_log.
func (t DatatypeType) IsSystemLogType() bool {
	return t.BaseType() == DatatypeTypeSystemLog
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
//   - Starts with _ and base matches a reserved type (with or without suffix): valid
//   - Starts with _ and is a plugin type: valid
//   - Starts with _ but base is not recognized: error
//   - Does not start with _: valid user type
func ValidateDatatypeType(t string) error {
	if t == "" {
		return fmt.Errorf("DatatypeType: cannot be empty")
	}
	if t[0] == '_' {
		dt := DatatypeType(t)
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

// --- Field type prefix helpers ---

const fieldTypeIDRefPrefix = "_id"

// IsIDRefType returns true if the field type starts with _id.
// Matches _id, _id_menu, etc.
func (t FieldType) IsIDRefType() bool {
	s := string(t)
	if s == fieldTypeIDRefPrefix {
		return true
	}
	return len(s) > len(fieldTypeIDRefPrefix) &&
		s[:len(fieldTypeIDRefPrefix)] == fieldTypeIDRefPrefix &&
		s[len(fieldTypeIDRefPrefix)] == '_'
}

// IDRefSuffix returns the suffix portion of an _id field type.
// "_id_menu" -> "menu", "_id" -> "".
func (t FieldType) IDRefSuffix() string {
	s := string(t)
	if len(s) > len(fieldTypeIDRefPrefix)+1 && s[len(fieldTypeIDRefPrefix)] == '_' {
		return s[len(fieldTypeIDRefPrefix)+1:]
	}
	return ""
}
