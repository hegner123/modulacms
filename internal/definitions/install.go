package definitions

import (
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Installer is the consumer-defined interface for creating schema records.
type Installer interface {
	CreateDatatype(db.CreateDatatypeParams) db.Datatypes
	CreateField(db.CreateFieldParams) db.Fields
	CreateDatatypeField(db.CreateDatatypeFieldParams) db.DatatypeFields
}

// InstallResult reports what was created during installation.
type InstallResult struct {
	DefinitionName string
	Datatypes      int
	Fields         int
	JunctionLinks  int
}

// Install creates all datatypes, fields, and junction records from a SchemaDefinition.
func Install(driver Installer, def SchemaDefinition, authorID types.UserID) (InstallResult, error) {
	if authorID.IsZero() {
		return InstallResult{}, fmt.Errorf("definitions: authorID cannot be empty")
	}

	if err := Validate(def); err != nil {
		return InstallResult{}, err
	}

	var result InstallResult
	result.DefinitionName = def.Name

	now := types.TimestampNow()
	author := types.NullableUserID{ID: authorID, Valid: true}

	// Phase 1: Create root datatypes (no ParentRef)
	datatypeIDMap := make(map[string]types.DatatypeID, len(def.Datatypes))

	for key, dt := range def.Datatypes {
		if dt.ParentRef != "" {
			continue
		}
		created := driver.CreateDatatype(db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			Label:        dt.Label,
			Type:         dt.Type.String,
			AuthorID:     author,
			DateCreated:  now,
			DateModified: now,
		})
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

			created := driver.CreateDatatype(db.CreateDatatypeParams{
				DatatypeID: types.NewDatatypeID(),
				ParentID: types.NullableDatatypeID{
					ID:    parentID,
					Valid: true,
				},
				Label:        dt.Label,
				Type:         dt.Type.String,
				AuthorID:     author,
				DateCreated:  now,
				DateModified: now,
			})
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

	// Phase 3: Create fields and junction records for each datatype
	for key, dt := range def.Datatypes {
		datatypeID := datatypeIDMap[key]
		for i, fieldDef := range dt.FieldRefs {
			data := types.EmptyJSON
			if fieldDef.Data.Valid {
				data = fieldDef.Data.String
			}

			validation, err := marshalConfig(fieldDef.Validation)
			if err != nil {
				return InstallResult{}, fmt.Errorf("definitions: datatype %q field %q validation: %w", key, fieldDef.Label, err)
			}

			uiConfig, err := marshalConfig(fieldDef.UiConfig)
			if err != nil {
				return InstallResult{}, fmt.Errorf("definitions: datatype %q field %q ui_config: %w", key, fieldDef.Label, err)
			}

			created := driver.CreateField(db.CreateFieldParams{
				FieldID:      types.NewFieldID(),
				Label:        fieldDef.Label,
				Data:         data,
				Validation:   validation,
				UIConfig:     uiConfig,
				Type:         fieldDef.Type,
				AuthorID:     author,
				DateCreated:  now,
				DateModified: now,
			})
			if created.FieldID.IsZero() {
				return InstallResult{}, fmt.Errorf("definitions: field %q created with zero ID", fieldDef.Label)
			}
			result.Fields++

			junc := driver.CreateDatatypeField(db.CreateDatatypeFieldParams{
				ID:         string(types.NewDatatypeFieldID()),
				DatatypeID: datatypeID,
				FieldID:    created.FieldID,
				SortOrder:  int64(i),
			})
			if junc.ID == "" {
				return InstallResult{}, fmt.Errorf("definitions: junction link for datatype %q field %q created with empty ID", key, fieldDef.Label)
			}
			result.JunctionLinks++
		}
	}

	return result, nil
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
