package tui

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// MapContentFieldsToDisplay maps raw DB content fields + field definitions
// to display structs, keyed by ContentDataID. Joins content field values with
// field definitions by FieldID to resolve labels, types, and metadata.
// Only fields with stored values are included (suitable for read-only preview).
func MapContentFieldsToDisplay(
	contentFields []db.ContentFields,
	fieldDefs []db.Fields,
) map[types.ContentID][]ContentFieldDisplay {
	// Build field def lookup: fieldID -> Fields
	defMap := make(map[types.FieldID]db.Fields, len(fieldDefs))
	for _, f := range fieldDefs {
		defMap[f.FieldID] = f
	}

	// Group content fields by ContentDataID, resolving labels from defs
	result := make(map[types.ContentID][]ContentFieldDisplay)
	for _, cf := range contentFields {
		if !cf.ContentDataID.Valid || !cf.FieldID.Valid {
			continue
		}
		cid := cf.ContentDataID.ID
		d := ContentFieldDisplay{
			ContentFieldID: cf.ContentFieldID,
			FieldID:        cf.FieldID.ID,
			Value:          cf.FieldValue,
		}
		if def, ok := defMap[cf.FieldID.ID]; ok {
			d.Label = def.Label
			d.Type = string(def.Type)
			d.ValidationJSON = "" // TODO: resolve from validation table
			d.DataJSON = def.Data
		}
		result[cid] = append(result[cid], d)
	}

	return result
}

// MapAdminContentFieldsToDisplay maps raw admin DB content fields + field definitions
// to display structs, keyed by AdminContentDataID.
func MapAdminContentFieldsToDisplay(
	contentFields []db.AdminContentFields,
	fieldDefs []db.AdminFields,
) map[types.AdminContentID][]AdminContentFieldDisplay {
	defMap := make(map[types.AdminFieldID]db.AdminFields, len(fieldDefs))
	for _, f := range fieldDefs {
		defMap[f.AdminFieldID] = f
	}

	result := make(map[types.AdminContentID][]AdminContentFieldDisplay)
	for _, cf := range contentFields {
		if !cf.AdminContentDataID.Valid || !cf.AdminFieldID.Valid {
			continue
		}
		cid := cf.AdminContentDataID.ID
		d := AdminContentFieldDisplay{
			AdminContentFieldID: cf.AdminContentFieldID,
			AdminFieldID:        cf.AdminFieldID.ID,
			Value:               cf.AdminFieldValue,
		}
		if def, ok := defMap[cf.AdminFieldID.ID]; ok {
			d.Label = def.Label
			d.Type = string(def.Type)
			d.ValidationJSON = "" // TODO: resolve from validation table
			d.DataJSON = def.Data
		}
		result[cid] = append(result[cid], d)
	}

	return result
}
