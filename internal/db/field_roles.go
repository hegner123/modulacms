package db

import "encoding/json"

// FilterFieldsByRole returns only the fields accessible to the given role.
// Admin bypasses all restrictions. NULL roles (not set) means unrestricted.
// Invalid JSON in roles is treated as restricted (fail-closed).
func FilterFieldsByRole(fields []Fields, roleID string, isAdmin bool) []Fields {
	if isAdmin {
		return fields
	}
	result := make([]Fields, 0, len(fields))
	for _, f := range fields {
		if !f.Roles.Valid {
			result = append(result, f)
			continue
		}
		if roleContains(f.Roles.String, roleID) {
			result = append(result, f)
		}
	}
	return result
}

// FilterAdminFieldsByRole returns only the admin fields accessible to the given role.
// Admin bypasses all restrictions. NULL roles (not set) means unrestricted.
// Invalid JSON in roles is treated as restricted (fail-closed).
func FilterAdminFieldsByRole(fields []AdminFields, roleID string, isAdmin bool) []AdminFields {
	if isAdmin {
		return fields
	}
	result := make([]AdminFields, 0, len(fields))
	for _, f := range fields {
		if !f.Roles.Valid {
			result = append(result, f)
			continue
		}
		if roleContains(f.Roles.String, roleID) {
			result = append(result, f)
		}
	}
	return result
}

// FilterFieldViewsByRole returns only the DatatypeFieldView entries accessible to the given role.
func FilterFieldViewsByRole(fields []DatatypeFieldView, roleID string, isAdmin bool) []DatatypeFieldView {
	if isAdmin {
		return fields
	}
	result := make([]DatatypeFieldView, 0, len(fields))
	for _, f := range fields {
		if !f.Roles.Valid {
			result = append(result, f)
			continue
		}
		if roleContains(f.Roles.String, roleID) {
			result = append(result, f)
		}
	}
	return result
}

// FilterFieldsWithSortOrderByRole returns only the FieldWithSortOrderRow entries accessible to the given role.
func FilterFieldsWithSortOrderByRole(fields []FieldWithSortOrderRow, roleID string, isAdmin bool) []FieldWithSortOrderRow {
	if isAdmin {
		return fields
	}
	result := make([]FieldWithSortOrderRow, 0, len(fields))
	for _, f := range fields {
		if !f.Roles.Valid {
			result = append(result, f)
			continue
		}
		if roleContains(f.Roles.String, roleID) {
			result = append(result, f)
		}
	}
	return result
}

// roleContains checks if roleID appears in a JSON array string.
// Returns false for invalid JSON (fail-closed).
func roleContains(jsonArray string, roleID string) bool {
	var roles []string
	if err := json.Unmarshal([]byte(jsonArray), &roles); err != nil {
		return false
	}
	for _, r := range roles {
		if r == roleID {
			return true
		}
	}
	return false
}

// IsFieldAccessible checks if a single field is accessible to the given role.
func IsFieldAccessible(f Fields, roleID string, isAdmin bool) bool {
	if isAdmin {
		return true
	}
	if !f.Roles.Valid {
		return true
	}
	return roleContains(f.Roles.String, roleID)
}
