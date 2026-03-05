package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// FetchUpdate signals a data fetch operation.
type FetchUpdate struct{}

// NewFetchUpdate returns a command that creates a FetchUpdate message.
func NewFetchUpdate() tea.Cmd {
	return func() tea.Msg {
		return FetchUpdate{}
	}
}

// systemDatatypeTypes are datatype types that should not appear in the child picker.
// Matches the admin panel's SYSTEM_TYPES in block-editor-src/cache.js.
var systemDatatypeTypes = map[string]bool{
	"_root":        true,
	"_nested_root": true,
	"_system_log":  true,
	"_reference":   true,
}

// filterChildDatatypes filters datatypes using the same 3-category logic as the
// admin panel block editor (cache.js fetchDatatypesGrouped):
//  1. Descendants of the root datatype
//  2. Children of _collection datatypes
//  3. _global datatypes and their children
//
// System types (_root, _nested_root, _system_log, _reference) are excluded.
func filterChildDatatypes(all []db.Datatypes, rootDatatypeID types.DatatypeID) []db.Datatypes {
	// Build lookup maps
	byID := make(map[types.DatatypeID]db.Datatypes, len(all))
	childrenOf := make(map[types.DatatypeID][]db.Datatypes)
	for _, dt := range all {
		byID[dt.DatatypeID] = dt
		pid := dt.ParentID
		if pid.Valid {
			childrenOf[pid.ID] = append(childrenOf[pid.ID], dt)
		}
	}

	// Recursively collect descendants, excluding system types
	var collectDescendants func(parentID types.DatatypeID) []db.Datatypes
	collectDescendants = func(parentID types.DatatypeID) []db.Datatypes {
		var result []db.Datatypes
		for _, kid := range childrenOf[parentID] {
			if systemDatatypeTypes[kid.Type] {
				continue
			}
			result = append(result, kid)
			result = append(result, collectDescendants(kid.DatatypeID)...)
		}
		return result
	}

	seen := make(map[types.DatatypeID]bool)
	var filtered []db.Datatypes
	addUnique := func(dts []db.Datatypes) {
		for _, dt := range dts {
			if !seen[dt.DatatypeID] {
				seen[dt.DatatypeID] = true
				filtered = append(filtered, dt)
			}
		}
	}

	// Category 1: Descendants of root datatype
	if _, ok := byID[rootDatatypeID]; ok {
		addUnique(collectDescendants(rootDatatypeID))
	}

	// Category 2: Children of _collection datatypes
	for _, dt := range all {
		if dt.Type == "_collection" {
			addUnique(collectDescendants(dt.DatatypeID))
		}
	}

	// Category 3: _global datatypes and their children
	for _, dt := range all {
		if dt.Type == string(types.DatatypeTypeGlobal) && !systemDatatypeTypes[dt.Type] {
			if !seen[dt.DatatypeID] {
				seen[dt.DatatypeID] = true
				filtered = append(filtered, dt)
			}
			addUnique(collectDescendants(dt.DatatypeID))
		}
	}

	return filtered
}
