package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// EnsureSystemData is an idempotent startup function that ensures system-level
// data exists in the database. It is safe to call on every boot — it checks
// for the existence of each resource before creating it.
//
// Currently ensures:
//   - The "content_tree_ref" field type exists in both field_types and admin_field_types
//   - The "_reference" system datatype exists with its "Target" field linked
func EnsureSystemData(ctx context.Context, driver DbDriver) error {
	if err := ensureFieldType(ctx, driver); err != nil {
		return fmt.Errorf("ensureFieldType: %w", err)
	}
	if err := ensureReferenceDatatype(ctx, driver); err != nil {
		return fmt.Errorf("ensureReferenceDatatype: %w", err)
	}
	return nil
}

// ensureFieldType checks that "content_tree_ref" exists in both field_types
// and admin_field_types tables, creating it if missing.
func ensureFieldType(ctx context.Context, driver DbDriver) error {
	const ftType = "content_tree_ref"
	const ftLabel = "Content Tree Reference"

	// Check field_types
	_, err := driver.GetFieldTypeByType(ftType)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !isNotFound(err) {
			return fmt.Errorf("check field_types for %q: %w", ftType, err)
		}
		systemUserID, userErr := findSystemUserID(driver)
		if userErr != nil {
			return fmt.Errorf("find system user for field_type seed: %w", userErr)
		}
		ac := audited.Ctx(types.NewNodeID(), systemUserID, "ensure-system-data", "system")
		_, err = driver.CreateFieldType(ctx, ac, CreateFieldTypeParams{Type: ftType, Label: ftLabel})
		if err != nil {
			return fmt.Errorf("create field_type %q: %w", ftType, err)
		}
		utility.DefaultLogger.Info("Created missing field_type", "type", ftType)
	}

	// Check admin_field_types
	_, err = driver.GetAdminFieldTypeByType(ftType)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !isNotFound(err) {
			return fmt.Errorf("check admin_field_types for %q: %w", ftType, err)
		}
		systemUserID, userErr := findSystemUserID(driver)
		if userErr != nil {
			return fmt.Errorf("find system user for admin_field_type seed: %w", userErr)
		}
		ac := audited.Ctx(types.NewNodeID(), systemUserID, "ensure-system-data", "system")
		_, err = driver.CreateAdminFieldType(ctx, ac, CreateAdminFieldTypeParams{Type: ftType, Label: ftLabel})
		if err != nil {
			return fmt.Errorf("create admin_field_type %q: %w", ftType, err)
		}
		utility.DefaultLogger.Info("Created missing admin_field_type", "type", ftType)
	}

	return nil
}

// ensureReferenceDatatype checks that a "_reference" datatype exists,
// creating it along with its "Target" field and junction link if missing.
func ensureReferenceDatatype(ctx context.Context, driver DbDriver) error {
	_, err := driver.GetDatatypeByType(string(types.DatatypeTypeReference))
	if err == nil {
		return nil // already exists
	}
	if !errors.Is(err, sql.ErrNoRows) && !isNotFound(err) {
		return fmt.Errorf("check _reference datatype: %w", err)
	}

	systemUserID, userErr := findSystemUserID(driver)
	if userErr != nil {
		return fmt.Errorf("find system user for _reference datatype: %w", userErr)
	}
	ac := audited.Ctx(types.NewNodeID(), systemUserID, "ensure-system-data", "system")

	// Create _reference datatype
	refDatatype, err := driver.CreateDatatype(ctx, ac, CreateDatatypeParams{
		ParentID:     types.NullableDatatypeID{},
		Name:         "reference",
		Label:        "Reference",
		Type:         string(types.DatatypeTypeReference),
		AuthorID:     systemUserID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("create _reference datatype: %w", err)
	}
	utility.DefaultLogger.Info("Creating missing _reference system datatype", "datatype_id", refDatatype.DatatypeID)

	// Create "Target" field linked to _reference datatype via parent_id
	_, err = driver.CreateField(ctx, ac, CreateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: refDatatype.DatatypeID, Valid: true},
		SortOrder:    0,
		Name:         "target",
		Label:        "Target",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeContentTreeRef,
		AuthorID:     types.NullableUserID{Valid: true, ID: systemUserID},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("create _reference Target field: %w", err)
	}

	return nil
}

// EnsurePublishPermission checks that the "content:publish" permission exists
// and is assigned to the admin role. This is idempotent — safe to call on every boot.
// For fresh installs (where CreateBootstrapData already includes "content:publish"),
// this is a no-op. For upgrades from older versions, this backfills the permission.
func EnsurePublishPermission(ctx context.Context, driver DbDriver) error {
	const label = "content:publish"

	// Check if the permission already exists
	perms, err := driver.ListPermissions()
	if err != nil {
		return fmt.Errorf("list permissions: %w", err)
	}
	if perms != nil {
		for _, p := range *perms {
			if p.Label == label {
				return nil // already exists
			}
		}
	}

	// Permission doesn't exist — create it
	systemUserID, userErr := findSystemUserID(driver)
	if userErr != nil {
		return fmt.Errorf("find system user for publish permission: %w", userErr)
	}
	ac := audited.Ctx(types.NewNodeID(), systemUserID, "ensure-publish-permission", "system")

	perm, err := driver.CreatePermission(ctx, ac, CreatePermissionParams{
		Label:           label,
		SystemProtected: true,
	})
	if err != nil {
		return fmt.Errorf("create permission %q: %w", label, err)
	}
	utility.DefaultLogger.Info("Created missing permission", "label", label)

	// Assign to admin role
	roles, err := driver.ListRoles()
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}
	if roles != nil {
		for _, r := range *roles {
			if r.Label == "admin" {
				_, assignErr := driver.CreateRolePermission(ctx, ac, CreateRolePermissionParams{
					RoleID:       r.RoleID,
					PermissionID: perm.PermissionID,
				})
				if assignErr != nil {
					return fmt.Errorf("assign %q to admin role: %w", label, assignErr)
				}
				utility.DefaultLogger.Info("Assigned permission to admin role", "label", label)
				break
			}
		}
	}

	return nil
}

// findSystemUserID returns the UserID of the "system" user.
func findSystemUserID(driver DbDriver) (types.UserID, error) {
	users, err := driver.ListUsers()
	if err != nil {
		return types.UserID(""), fmt.Errorf("list users: %w", err)
	}
	if users == nil {
		return types.UserID(""), fmt.Errorf("no users found")
	}
	for _, u := range *users {
		if u.Username == "system" {
			return u.UserID, nil
		}
	}
	// Fall back to the first user if "system" not found
	if len(*users) > 0 {
		return (*users)[0].UserID, nil
	}
	return types.UserID(""), fmt.Errorf("no users found")
}

// isNotFound checks if an error message indicates a "not found" condition.
// Some wrapper methods wrap sql.ErrNoRows in fmt.Errorf, losing the sentinel.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for i := 0; i <= len(msg)-9; i++ {
		if msg[i:i+9] == "no rows i" {
			return true
		}
	}
	return false
}
