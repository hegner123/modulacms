package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestFilterFieldsByRole(t *testing.T) {
	adminRole := "01ADMIN000000000000000ADMIN"
	editorRole := "01EDITOR00000000000000EDITOR"
	viewerRole := "01VIEWER00000000000000VIEWER"

	makeField := func(label string, roles types.NullableString) Fields {
		return Fields{
			FieldID: types.NewFieldID(),
			Label:   label,
			Roles:   roles,
		}
	}

	unrestricted := makeField("unrestricted", types.NullableString{})
	editorOnly := makeField("editor_only", types.NewNullableString(`["`+editorRole+`"]`))
	editorAndViewer := makeField("editor_and_viewer", types.NewNullableString(`["`+editorRole+`","`+viewerRole+`"]`))
	emptyArray := makeField("empty_array", types.NewNullableString(`[]`))
	invalidJSON := makeField("invalid_json", types.NewNullableString(`not valid json`))

	allFields := []Fields{unrestricted, editorOnly, editorAndViewer, emptyArray, invalidJSON}

	tests := []struct {
		name     string
		roleID   string
		isAdmin  bool
		expected []string
	}{
		{
			name:     "admin bypasses all restrictions",
			roleID:   adminRole,
			isAdmin:  true,
			expected: []string{"unrestricted", "editor_only", "editor_and_viewer", "empty_array", "invalid_json"},
		},
		{
			name:     "editor sees unrestricted and editor-allowed fields",
			roleID:   editorRole,
			isAdmin:  false,
			expected: []string{"unrestricted", "editor_only", "editor_and_viewer"},
		},
		{
			name:     "viewer sees unrestricted and viewer-allowed fields",
			roleID:   viewerRole,
			isAdmin:  false,
			expected: []string{"unrestricted", "editor_and_viewer"},
		},
		{
			name:     "unknown role sees only unrestricted fields",
			roleID:   "01UNKNOWN0000000000000UNKNOWN",
			isAdmin:  false,
			expected: []string{"unrestricted"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FilterFieldsByRole(allFields, tc.roleID, tc.isAdmin)
			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d fields, got %d", len(tc.expected), len(result))
			}
			for i, f := range result {
				if f.Label != tc.expected[i] {
					t.Errorf("field[%d]: expected label %q, got %q", i, tc.expected[i], f.Label)
				}
			}
		})
	}
}

func TestRoleContains(t *testing.T) {
	tests := []struct {
		name      string
		jsonArray string
		roleID    string
		expected  bool
	}{
		{"matching role", `["role1","role2"]`, "role1", true},
		{"non-matching role", `["role1","role2"]`, "role3", false},
		{"empty array", `[]`, "role1", false},
		{"invalid JSON", `not json`, "role1", false},
		{"empty string", ``, "role1", false},
		{"single role match", `["role1"]`, "role1", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := roleContains(tc.jsonArray, tc.roleID)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestIsFieldAccessible(t *testing.T) {
	editorRole := "01EDITOR00000000000000EDITOR"

	tests := []struct {
		name     string
		field    Fields
		roleID   string
		isAdmin  bool
		expected bool
	}{
		{
			name:     "admin always accessible",
			field:    Fields{Roles: types.NewNullableString(`["other"]`)},
			roleID:   editorRole,
			isAdmin:  true,
			expected: true,
		},
		{
			name:     "null roles accessible to all",
			field:    Fields{},
			roleID:   editorRole,
			isAdmin:  false,
			expected: true,
		},
		{
			name:     "matching role accessible",
			field:    Fields{Roles: types.NewNullableString(`["` + editorRole + `"]`)},
			roleID:   editorRole,
			isAdmin:  false,
			expected: true,
		},
		{
			name:     "non-matching role blocked",
			field:    Fields{Roles: types.NewNullableString(`["other"]`)},
			roleID:   editorRole,
			isAdmin:  false,
			expected: false,
		},
		{
			name:     "invalid JSON blocked",
			field:    Fields{Roles: types.NewNullableString(`broken`)},
			roleID:   editorRole,
			isAdmin:  false,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsFieldAccessible(tc.field, tc.roleID, tc.isAdmin)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
