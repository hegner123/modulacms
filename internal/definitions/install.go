package definitions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Installer is the consumer-defined interface for creating schema records.
type Installer interface {
	CreateDatatype(db.CreateDatatypeParams) (db.Datatypes, error)
	CreateField(db.CreateFieldParams) (db.Fields, error)
}

// Cleaner is the consumer-defined interface for listing and deleting
// bootstrapped records during a reinstall. Separate from Installer to
// keep Install() unchanged.
type Cleaner interface {
	GetUserByEmail(types.Email) (*db.Users, error)
	ListDatatypes() (*[]db.Datatypes, error)
	ListFields() (*[]db.Fields, error)
	ListContentData() (*[]db.ContentData, error)
	ListContentFields() (*[]db.ContentFields, error)
	ListRoutes() (*[]db.Routes, error)
	DeleteContentField(context.Context, audited.AuditContext, types.ContentFieldID) error
	DeleteContentData(context.Context, audited.AuditContext, types.ContentID) error
	DeleteField(context.Context, audited.AuditContext, types.FieldID) error
	DeleteDatatype(context.Context, audited.AuditContext, types.DatatypeID) error
	DeleteRoute(context.Context, audited.AuditContext, types.RouteID) error
}

// CleanupResult reports what was deleted during cleanup.
type CleanupResult struct {
	ContentFields int
	ContentData   int
	Fields        int
	Datatypes     int
	Routes        int
}

// InstallResult reports what was created during installation.
type InstallResult struct {
	DefinitionName string
	Datatypes      int
	Fields         int
}

// Install creates all datatypes and fields from a SchemaDefinition.
// existingDatatypes is an optional list of datatypes already in the database.
// When provided, root datatypes that match an existing record by name are
// reused instead of creating duplicates (e.g. bootstrap "page" matches
// schema "page"). Pass nil to skip this check.
func Install(driver Installer, def SchemaDefinition, authorID types.UserID, existingDatatypes []db.Datatypes) (InstallResult, error) {
	if authorID.IsZero() {
		return InstallResult{}, fmt.Errorf("definitions: authorID cannot be empty")
	}

	if err := Validate(def); err != nil {
		return InstallResult{}, err
	}

	var result InstallResult
	result.DefinitionName = def.Name

	now := types.TimestampNow()

	// Build name→ID lookup from existing datatypes for dedup.
	existingByName := make(map[string]types.DatatypeID, len(existingDatatypes))
	for _, dt := range existingDatatypes {
		if dt.Name != "" {
			existingByName[dt.Name] = dt.DatatypeID
		}
	}

	// Phase 1: Create root datatypes (no ParentRef)
	datatypeIDMap := make(map[string]types.DatatypeID, len(def.Datatypes))

	for key, dt := range def.Datatypes {
		if dt.ParentRef != "" {
			continue
		}
		// Reuse existing datatype if one matches by name.
		if existingID, found := existingByName[dt.Name]; found {
			datatypeIDMap[key] = existingID
			continue
		}
		created, createErr := driver.CreateDatatype(db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			Name:         dt.Name,
			Label:        dt.Label,
			Type:         dt.Type.String,
			AuthorID:     authorID,
			DateCreated:  now,
			DateModified: now,
		})
		if createErr != nil {
			return InstallResult{}, fmt.Errorf("definitions: datatype %q: %w", key, createErr)
		}
		if created.DatatypeID.IsZero() {
			return InstallResult{}, fmt.Errorf("definitions: datatype %q created with zero ID", key)
		}
		datatypeIDMap[key] = created.DatatypeID
		result.Datatypes++
	}

	// Phase 2: Create child datatypes iteratively until all are resolved.
	// Each pass creates datatypes whose ParentRef is already in datatypeIDMap.
	maxPasses := len(def.Datatypes)
	for range maxPasses {
		progress := false
		for key, dt := range def.Datatypes {
			if _, done := datatypeIDMap[key]; done {
				continue
			}

			parentID, parentResolved := datatypeIDMap[dt.ParentRef]
			if !parentResolved {
				continue
			}

			created, createErr := driver.CreateDatatype(db.CreateDatatypeParams{
				DatatypeID: types.NewDatatypeID(),
				ParentID: types.NullableDatatypeID{
					ID:    parentID,
					Valid: true,
				},
				Name:         dt.Name,
				Label:        dt.Label,
				Type:         dt.Type.String,
				AuthorID:     authorID,
				DateCreated:  now,
				DateModified: now,
			})
			if createErr != nil {
				return InstallResult{}, fmt.Errorf("definitions: datatype %q: %w", key, createErr)
			}
			if created.DatatypeID.IsZero() {
				return InstallResult{}, fmt.Errorf("definitions: datatype %q created with zero ID", key)
			}
			datatypeIDMap[key] = created.DatatypeID
			result.Datatypes++
			progress = true
		}

		if len(datatypeIDMap) == len(def.Datatypes) {
			break
		}
		if !progress {
			return InstallResult{}, fmt.Errorf("definitions: circular dependency detected, %d datatypes unresolved", len(def.Datatypes)-len(datatypeIDMap))
		}
	}

	if len(datatypeIDMap) != len(def.Datatypes) {
		return InstallResult{}, fmt.Errorf("definitions: %d datatypes could not be resolved", len(def.Datatypes)-len(datatypeIDMap))
	}

	// Phase 3: Create fields with parent_id linking them to their datatype
	for key, dt := range def.Datatypes {
		datatypeID := datatypeIDMap[key]
		for _, fieldDef := range dt.FieldRefs {
			data := types.EmptyJSON
			if fieldDef.Data.Valid {
				data = fieldDef.Data.String
			}

			// TODO: create validation record from fieldDef.Validation and reference by ID
			_ = fieldDef.Validation

			uiConfig, err := marshalConfig(fieldDef.UiConfig)
			if err != nil {
				return InstallResult{}, fmt.Errorf("definitions: datatype %q field %q ui_config: %w", key, fieldDef.Label, err)
			}

			created, fieldErr := driver.CreateField(db.CreateFieldParams{
				FieldID:      types.NewFieldID(),
				ParentID:     types.NullableDatatypeID{ID: datatypeID, Valid: true},
				Name:         fieldDef.Name,
				Label:        fieldDef.Label,
				Data:         data,
				ValidationID: types.NullableValidationID{},
				UIConfig:     uiConfig,
				Type:         fieldDef.Type,
				AuthorID:     types.NullableUserID{ID: authorID, Valid: true},
				DateCreated:  now,
				DateModified: now,
			})
			if fieldErr != nil {
				return InstallResult{}, fmt.Errorf("definitions: datatype %q field %q: %w", key, fieldDef.Label, fieldErr)
			}
			if created.FieldID.IsZero() {
				return InstallResult{}, fmt.Errorf("definitions: field %q created with zero ID", fieldDef.Label)
			}
			result.Fields++
		}
	}

	return result, nil
}

// Reinstall deletes all system-authored records from a previous install, then
// runs Install to recreate them from the definition. User-created records and
// system datatypes (type starting with "_") are preserved.
func Reinstall(ctx context.Context, cleaner Cleaner, installer Installer, def SchemaDefinition, authorID types.UserID) (InstallResult, error) {
	if authorID.IsZero() {
		return InstallResult{}, fmt.Errorf("definitions: authorID cannot be empty")
	}

	// Look up system user to identify bootstrapped records.
	systemUser, err := cleaner.GetUserByEmail(types.Email("system@modula.local"))
	if err != nil {
		return InstallResult{}, fmt.Errorf("definitions: get system user: %w", err)
	}
	systemUserID := systemUser.UserID

	ac := audited.Ctx(types.NewNodeID(), authorID, "reinstall", "system")

	allDatatypes, err := cleaner.ListDatatypes()
	if err != nil {
		return InstallResult{}, fmt.Errorf("definitions: list datatypes: %w", err)
	}

	var cleanup CleanupResult

	// Step 1: Delete system-authored content_fields.
	contentFields, err := cleaner.ListContentFields()
	if err != nil {
		return InstallResult{}, fmt.Errorf("definitions: list content_fields: %w", err)
	}
	if contentFields != nil {
		for _, cf := range *contentFields {
			if cf.AuthorID == systemUserID {
				if delErr := cleaner.DeleteContentField(ctx, ac, cf.ContentFieldID); delErr != nil {
					return InstallResult{}, fmt.Errorf("definitions: delete content_field %s: %w", cf.ContentFieldID, delErr)
				}
				cleanup.ContentFields++
			}
		}
	}

	// Step 2: Delete system-authored content_data.
	contentData, err := cleaner.ListContentData()
	if err != nil {
		return InstallResult{}, fmt.Errorf("definitions: list content_data: %w", err)
	}
	if contentData != nil {
		for _, cd := range *contentData {
			if cd.AuthorID == systemUserID {
				if delErr := cleaner.DeleteContentData(ctx, ac, cd.ContentDataID); delErr != nil {
					return InstallResult{}, fmt.Errorf("definitions: delete content_data %s: %w", cd.ContentDataID, delErr)
				}
				cleanup.ContentData++
			}
		}
	}

	// Step 3: Delete system-authored fields.
	fields, err := cleaner.ListFields()
	if err != nil {
		return InstallResult{}, fmt.Errorf("definitions: list fields: %w", err)
	}
	if fields != nil {
		for _, f := range *fields {
			if !f.AuthorID.Valid || f.AuthorID.ID != systemUserID {
				continue
			}
			if delErr := cleaner.DeleteField(ctx, ac, f.FieldID); delErr != nil {
				return InstallResult{}, fmt.Errorf("definitions: delete field %s: %w", f.FieldID, delErr)
			}
			cleanup.Fields++
		}
	}

	// Step 5: Delete system-authored datatypes.
	if allDatatypes != nil {
		for _, dt := range *allDatatypes {
			if dt.AuthorID == systemUserID {
				if delErr := cleaner.DeleteDatatype(ctx, ac, dt.DatatypeID); delErr != nil {
					return InstallResult{}, fmt.Errorf("definitions: delete datatype %s: %w", dt.DatatypeID, delErr)
				}
				cleanup.Datatypes++
			}
		}
	}

	// Step 6: Delete system-authored routes.
	routes, err := cleaner.ListRoutes()
	if err != nil {
		return InstallResult{}, fmt.Errorf("definitions: list routes: %w", err)
	}
	if routes != nil {
		for _, r := range *routes {
			if r.AuthorID.Valid && r.AuthorID.ID == systemUserID {
				if delErr := cleaner.DeleteRoute(ctx, ac, r.RouteID); delErr != nil {
					return InstallResult{}, fmt.Errorf("definitions: delete route %s: %w", r.RouteID, delErr)
				}
				cleanup.Routes++
			}
		}
	}

	// Now install fresh. Reinstall cleans up first, so no existing datatypes to dedup.
	return Install(installer, def, authorID, nil)
}

// marshalConfig marshals a config struct to JSON. Returns EmptyJSON for zero-value structs.
func marshalConfig(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	s := string(b)
	if s == "{}" || s == "null" {
		return types.EmptyJSON, nil
	}
	return s, nil
}
