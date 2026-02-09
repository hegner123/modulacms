package definitions

import (
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

// Install creates all fields, datatypes, and junction records from a SchemaDefinition.
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

	// Phase 1: Create all fields (no dependencies)
	fieldIDMap := make(map[string]types.FieldID, len(def.Fields))
	for key, fieldDef := range def.Fields {
		created := driver.CreateField(db.CreateFieldParams{
			FieldID:      types.NewFieldID(),
			Label:        fieldDef.Label,
			Data:         fieldDef.Data,
			Validation:   types.EmptyJSON,
			UIConfig:     types.EmptyJSON,
			Type:         fieldDef.Type,
			AuthorID:     author,
			DateCreated:  now,
			DateModified: now,
		})
		if created.FieldID.IsZero() {
			return InstallResult{}, fmt.Errorf("definitions: field %q created with zero ID", key)
		}
		fieldIDMap[key] = created.FieldID
		result.Fields++
	}

	// Phase 2: Create datatypes — roots first, then children iteratively
	datatypeIDMap := make(map[string]types.DatatypeID, len(def.Datatypes))

	// Create root datatypes (no parent)
	for _, key := range def.RootKeys {
		dt := def.Datatypes[key]
		created := driver.CreateDatatype(db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			Label:        dt.Label,
			Type:         dt.Type,
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

	// Iteratively resolve children until all datatypes are created.
	// Each pass creates datatypes whose parent is already in datatypeIDMap.
	// Guard against circular dependencies with a max-iterations check.
	maxPasses := len(def.Datatypes)
	for pass := range maxPasses {
		progress := false
		for key, dt := range def.Datatypes {
			if _, done := datatypeIDMap[key]; done {
				continue
			}

			// Find if this datatype is a child of any already-created datatype
			parentKey := ""
			for candidateKey, candidateDT := range def.Datatypes {
				if _, parentDone := datatypeIDMap[candidateKey]; !parentDone {
					continue
				}
				for _, childRef := range candidateDT.ChildRefs {
					if childRef == key {
						parentKey = candidateKey
						break
					}
				}
				if parentKey != "" {
					break
				}
			}

			if parentKey == "" {
				continue
			}

			parentDatatypeID := datatypeIDMap[parentKey]
			created := driver.CreateDatatype(db.CreateDatatypeParams{
				DatatypeID: types.NewDatatypeID(),
				ParentID: types.NullableContentID{
					ID:    types.ContentID(parentDatatypeID),
					Valid: true,
				},
				Label:        dt.Label,
				Type:         dt.Type,
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
		if !progress && pass > 0 {
			return InstallResult{}, fmt.Errorf("definitions: circular dependency detected, %d datatypes unresolved", len(def.Datatypes)-len(datatypeIDMap))
		}
	}

	if len(datatypeIDMap) != len(def.Datatypes) {
		return InstallResult{}, fmt.Errorf("definitions: %d datatypes could not be resolved", len(def.Datatypes)-len(datatypeIDMap))
	}

	// Phase 3: Create junction records (datatype <-> field links)
	for key, dt := range def.Datatypes {
		datatypeID := datatypeIDMap[key]
		for _, fieldRef := range dt.FieldRefs {
			fieldID := fieldIDMap[fieldRef]
			created := driver.CreateDatatypeField(db.CreateDatatypeFieldParams{
				ID:         string(types.NewDatatypeFieldID()),
				DatatypeID: datatypeID,
				FieldID:    fieldID,
			})
			if created.ID == "" {
				return InstallResult{}, fmt.Errorf("definitions: junction link for datatype %q field %q created with empty ID", key, fieldRef)
			}
			result.JunctionLinks++
		}
	}

	return result, nil
}
