package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// =============================================================================
// ADMIN ROUTES CONSTRUCTORS
// =============================================================================

// AdminRoutesFetchCmd creates a command to fetch all admin routes.
func AdminRoutesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminRoutesFetchMsg{} }
}

// AdminRoutesSetCmd creates a command to set the admin routes list.
func AdminRoutesSetCmd(routes []db.AdminRoutes) tea.Cmd {
	return func() tea.Msg { return AdminRoutesSet{AdminRoutes: routes} }
}

// CreateAdminRouteFromDialogCmd creates a command to create an admin route from dialog input.
func CreateAdminRouteFromDialogCmd(title, slug string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminRouteFromDialogRequestMsg{Title: title, Slug: slug}
	}
}

// UpdateAdminRouteFromDialogCmd creates a command to update an admin route from dialog input.
func UpdateAdminRouteFromDialogCmd(routeID, title, slug, originalSlug string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminRouteFromDialogRequestMsg{
			RouteID:      routeID,
			Title:        title,
			Slug:         slug,
			OriginalSlug: originalSlug,
		}
	}
}

// DeleteAdminRouteCmd creates a command to delete an admin route.
func DeleteAdminRouteCmd(adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminRouteRequestMsg{AdminRouteID: adminRouteID} }
}

// ShowDeleteAdminRouteDialogCmd creates a command to show the delete admin route dialog.
func ShowDeleteAdminRouteDialogCmd(adminRouteID types.AdminRouteID, title string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminRouteDialogMsg{AdminRouteID: adminRouteID, Title: title}
	}
}

// =============================================================================
// ADMIN DATATYPES CONSTRUCTORS
// =============================================================================

// AdminAllDatatypesFetchCmd creates a command to fetch all admin datatypes.
func AdminAllDatatypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminAllDatatypesFetchMsg{} }
}

// AdminAllDatatypesSetCmd creates a command to set the admin datatypes list.
func AdminAllDatatypesSetCmd(datatypes []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg { return AdminAllDatatypesSet{AdminAllDatatypes: datatypes} }
}

// AdminDatatypeFieldsFetchCmd creates a command to fetch fields for an admin datatype.
func AdminDatatypeFieldsFetchCmd(adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg { return AdminDatatypeFieldsFetchMsg{AdminDatatypeID: adminDatatypeID} }
}

// AdminDatatypeFieldsSetCmd creates a command to set the admin datatype fields list.
func AdminDatatypeFieldsSetCmd(fields []db.AdminFields) tea.Cmd {
	return func() tea.Msg { return AdminDatatypeFieldsSet{Fields: fields} }
}

// CreateAdminDatatypeFromDialogCmd creates a command to create an admin datatype from dialog input.
func CreateAdminDatatypeFromDialogCmd(name, label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminDatatypeFromDialogRequestMsg{Name: name, Label: label, Type: dtype, ParentID: parentID}
	}
}

// UpdateAdminDatatypeFromDialogCmd creates a command to update an admin datatype from dialog input.
func UpdateAdminDatatypeFromDialogCmd(adminDatatypeID, name, label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminDatatypeFromDialogRequestMsg{
			AdminDatatypeID: adminDatatypeID,
			Name:            name,
			Label:           label,
			Type:            dtype,
			ParentID:        parentID,
		}
	}
}

// DeleteAdminDatatypeCmd creates a command to delete an admin datatype.
func DeleteAdminDatatypeCmd(adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminDatatypeRequestMsg{AdminDatatypeID: adminDatatypeID} }
}

// ShowDeleteAdminDatatypeDialogCmd creates a command to show the delete admin datatype dialog.
func ShowDeleteAdminDatatypeDialogCmd(adminDatatypeID types.AdminDatatypeID, label string, hasChildren bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminDatatypeDialogMsg{
			AdminDatatypeID: adminDatatypeID,
			Label:           label,
			HasChildren:     hasChildren,
		}
	}
}

// =============================================================================
// ADMIN FIELDS CONSTRUCTORS
// =============================================================================

// CreateAdminFieldFromDialogCmd creates a command to create an admin field from dialog input.
func CreateAdminFieldFromDialogCmd(name, label, fieldType string, adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminFieldFromDialogRequestMsg{
			Name:            name,
			Label:           label,
			Type:            fieldType,
			AdminDatatypeID: adminDatatypeID,
		}
	}
}

// UpdateAdminFieldFromDialogCmd creates a command to update an admin field from dialog input.
func UpdateAdminFieldFromDialogCmd(adminFieldID, name, label, fieldType string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminFieldFromDialogRequestMsg{
			AdminFieldID: adminFieldID,
			Name:         name,
			Label:        label,
			Type:         fieldType,
		}
	}
}

// DeleteAdminFieldCmd creates a command to delete an admin field.
func DeleteAdminFieldCmd(adminFieldID types.AdminFieldID, adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return DeleteAdminFieldRequestMsg{AdminFieldID: adminFieldID}
	}
}

// ShowDeleteAdminFieldDialogCmd creates a command to show the delete admin field dialog.
func ShowDeleteAdminFieldDialogCmd(adminFieldID types.AdminFieldID, adminDatatypeID types.AdminDatatypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminFieldDialogMsg{
			AdminFieldID:    adminFieldID,
			AdminDatatypeID: adminDatatypeID,
			Label:           label,
		}
	}
}

// =============================================================================
// ADMIN CONTENT CONSTRUCTORS
// =============================================================================

// AdminContentDataFetchCmd creates a command to fetch all admin content data.
func AdminContentDataFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminContentDataFetchMsg{} }
}

// AdminContentDataSetCmd creates a command to set the admin content data list.
func AdminContentDataSetCmd(data []db.AdminContentDataTopLevel) tea.Cmd {
	return func() tea.Msg { return AdminContentDataSet{AdminContentData: data} }
}

// DeleteAdminContentCmd creates a command to delete admin content.
func DeleteAdminContentCmd(adminContentID types.AdminContentID, adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg {
		return DeleteAdminContentRequestMsg{AdminContentID: adminContentID, AdminRouteID: adminRouteID}
	}
}

// =============================================================================
// FIELD TYPES CONSTRUCTORS
// =============================================================================

// CreateFieldTypeFromDialogCmd creates a command to create a field type from dialog input.
func CreateFieldTypeFromDialogCmd(fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return CreateFieldTypeFromDialogRequestMsg{Type: fieldType, Label: label}
	}
}

// UpdateFieldTypeFromDialogCmd creates a command to update a field type from dialog input.
func UpdateFieldTypeFromDialogCmd(fieldTypeID, fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return UpdateFieldTypeFromDialogRequestMsg{
			FieldTypeID: fieldTypeID,
			Type:        fieldType,
			Label:       label,
		}
	}
}

// CreateAdminFieldTypeFromDialogCmd creates a command to create an admin field type from dialog input.
func CreateAdminFieldTypeFromDialogCmd(fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminFieldTypeFromDialogRequestMsg{Type: fieldType, Label: label}
	}
}

// UpdateAdminFieldTypeFromDialogCmd creates a command to update an admin field type from dialog input.
func UpdateAdminFieldTypeFromDialogCmd(adminFieldTypeID, fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminFieldTypeFromDialogRequestMsg{
			AdminFieldTypeID: adminFieldTypeID,
			Type:             fieldType,
			Label:            label,
		}
	}
}

// FieldTypesFetchCmd creates a command to fetch all field types.
func FieldTypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return FieldTypesFetchMsg{} }
}

// FieldTypesSetCmd creates a command to set the field types list.
func FieldTypesSetCmd(fieldTypes []db.FieldTypes) tea.Cmd {
	return func() tea.Msg { return FieldTypesSet{FieldTypes: fieldTypes} }
}

// DeleteFieldTypeCmd creates a command to delete a field type.
func DeleteFieldTypeCmd(fieldTypeID types.FieldTypeID) tea.Cmd {
	return func() tea.Msg { return DeleteFieldTypeRequestMsg{FieldTypeID: fieldTypeID} }
}

// ShowDeleteFieldTypeDialogCmd creates a command to show the delete field type dialog.
func ShowDeleteFieldTypeDialogCmd(fieldTypeID types.FieldTypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteFieldTypeDialogMsg{FieldTypeID: fieldTypeID, Label: label}
	}
}

// =============================================================================
// ADMIN FIELD TYPES CONSTRUCTORS
// =============================================================================

// AdminFieldTypesFetchCmd creates a command to fetch all admin field types.
func AdminFieldTypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminFieldTypesFetchMsg{} }
}

// AdminFieldTypesSetCmd creates a command to set the admin field types list.
func AdminFieldTypesSetCmd(adminFieldTypes []db.AdminFieldTypes) tea.Cmd {
	return func() tea.Msg { return AdminFieldTypesSet{AdminFieldTypes: adminFieldTypes} }
}

// DeleteAdminFieldTypeCmd creates a command to delete an admin field type.
func DeleteAdminFieldTypeCmd(adminFieldTypeID types.AdminFieldTypeID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminFieldTypeRequestMsg{AdminFieldTypeID: adminFieldTypeID} }
}

// ShowDeleteAdminFieldTypeDialogCmd creates a command to show the delete admin field type dialog.
func ShowDeleteAdminFieldTypeDialogCmd(adminFieldTypeID types.AdminFieldTypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminFieldTypeDialogMsg{AdminFieldTypeID: adminFieldTypeID, Label: label}
	}
}

// =============================================================================
// VALIDATION CONSTRUCTORS
// =============================================================================

// ValidationsFetchCmd creates a command to fetch all validations.
func ValidationsFetchCmd() tea.Cmd {
	return func() tea.Msg { return ValidationsFetchMsg{} }
}

// ValidationsSetCmd creates a command to set the validations list.
func ValidationsSetCmd(validations []db.Validation) tea.Cmd {
	return func() tea.Msg { return ValidationsSet{Validations: validations} }
}

// DeleteValidationCmd creates a command to delete a validation.
func DeleteValidationCmd(validationID types.ValidationID) tea.Cmd {
	return func() tea.Msg { return DeleteValidationRequestMsg{ValidationID: validationID} }
}

// ShowDeleteValidationDialogCmd creates a command to show the delete validation dialog.
func ShowDeleteValidationDialogCmd(validationID types.ValidationID, name string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteValidationDialogMsg{ValidationID: validationID, Name: name}
	}
}

// CreateValidationFromDialogCmd creates a command to create a validation from dialog input.
func CreateValidationFromDialogCmd(name, description string) tea.Cmd {
	return func() tea.Msg {
		return CreateValidationFromDialogRequestMsg{Name: name, Description: description}
	}
}

// UpdateValidationFromDialogCmd creates a command to update a validation from dialog input.
func UpdateValidationFromDialogCmd(validationID, name, description string) tea.Cmd {
	return func() tea.Msg {
		return UpdateValidationFromDialogRequestMsg{
			ValidationID: validationID,
			Name:         name,
			Description:  description,
		}
	}
}

// ShowEditValidationDialogCmd creates a command to show the edit validation dialog.
func ShowEditValidationDialogCmd(v db.Validation) tea.Cmd {
	return func() tea.Msg { return ShowEditValidationDialogMsg{Validation: v} }
}

// =============================================================================
// ADMIN VALIDATION CONSTRUCTORS
// =============================================================================

// AdminValidationsFetchCmd creates a command to fetch all admin validations.
func AdminValidationsFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminValidationsFetchMsg{} }
}

// AdminValidationsSetCmd creates a command to set the admin validations list.
func AdminValidationsSetCmd(adminValidations []db.AdminValidation) tea.Cmd {
	return func() tea.Msg { return AdminValidationsSet{AdminValidations: adminValidations} }
}

// DeleteAdminValidationCmd creates a command to delete an admin validation.
func DeleteAdminValidationCmd(adminValidationID types.AdminValidationID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminValidationRequestMsg{AdminValidationID: adminValidationID} }
}

// ShowDeleteAdminValidationDialogCmd creates a command to show the delete admin validation dialog.
func ShowDeleteAdminValidationDialogCmd(adminValidationID types.AdminValidationID, name string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminValidationDialogMsg{AdminValidationID: adminValidationID, Name: name}
	}
}

// CreateAdminValidationFromDialogCmd creates a command to create an admin validation from dialog input.
func CreateAdminValidationFromDialogCmd(name, description string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminValidationFromDialogRequestMsg{Name: name, Description: description}
	}
}

// UpdateAdminValidationFromDialogCmd creates a command to update an admin validation from dialog input.
func UpdateAdminValidationFromDialogCmd(adminValidationID, name, description string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminValidationFromDialogRequestMsg{
			AdminValidationID: adminValidationID,
			Name:              name,
			Description:       description,
		}
	}
}

// ShowEditAdminValidationDialogCmd creates a command to show the edit admin validation dialog.
func ShowEditAdminValidationDialogCmd(v db.AdminValidation) tea.Cmd {
	return func() tea.Msg { return ShowEditAdminValidationDialogMsg{AdminValidation: v} }
}

// =============================================================================
// ADMIN CONTENT TREE CONSTRUCTORS
// =============================================================================

// ReloadAdminContentTreeCmd fetches admin content data and builds a tree.
func ReloadAdminContentTreeCmd(cfg *config.Config, adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		routeFilter := types.NullableAdminRouteID{ID: adminRouteID, Valid: true}

		rows, err := d.ListAdminContentDataWithDatatypeByRoute(routeFilter)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to load admin content: " + err.Error()}
		}

		var cd []db.AdminContentData
		var dt []db.AdminDatatypes
		dtSeen := make(map[types.AdminDatatypeID]bool)

		if rows != nil {
			for _, r := range *rows {
				cd = append(cd, db.AdminContentData{
					AdminContentDataID: r.AdminContentDataID,
					ParentID:           r.ParentID,
					FirstChildID:       r.FirstChildID,
					NextSiblingID:      r.NextSiblingID,
					PrevSiblingID:      r.PrevSiblingID,
					AdminRouteID:       r.AdminRouteID,
					AdminDatatypeID:    r.AdminDatatypeID,
					AuthorID:           r.AuthorID,
					Status:             r.Status,
					DateCreated:        r.DateCreated,
					DateModified:       r.DateModified,
				})
				if !r.DtAdminDatatypeID.IsZero() && !dtSeen[r.DtAdminDatatypeID] {
					dtSeen[r.DtAdminDatatypeID] = true
					dt = append(dt, db.AdminDatatypes{
						AdminDatatypeID: r.DtAdminDatatypeID,
						ParentID:        r.DtParentID,
						Label:           r.DtLabel,
						Type:            r.DtType,
						AuthorID:        r.DtAuthorID,
						DateCreated:     r.DtDateCreated,
						DateModified:    r.DtDateModified,
					})
				}
			}
		}

		cfRows, cfErr := d.ListAdminContentFieldsWithFieldByRoute(routeFilter)
		if cfErr != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to load admin content fields: " + cfErr.Error()}
		}

		var cf []db.AdminContentFields
		var df []db.AdminFields
		dfSeen := make(map[types.AdminFieldID]bool)

		if cfRows != nil {
			for _, r := range *cfRows {
				cf = append(cf, db.AdminContentFields{
					AdminContentFieldID: r.AdminContentFieldID,
					AdminRouteID:        r.AdminRouteID,
					AdminContentDataID:  r.AdminContentDataID,
					AdminFieldID:        r.AdminFieldID,
					AdminFieldValue:     r.AdminFieldValue,
					AuthorID:            r.AuthorID,
					DateCreated:         r.DateCreated,
					DateModified:        r.DateModified,
				})
				if r.AdminFieldID.Valid && !dfSeen[r.FAdminFieldID] {
					dfSeen[r.FAdminFieldID] = true
					df = append(df, db.AdminFields{
						AdminFieldID: r.FAdminFieldID,
						ParentID:     r.FParentID,
						Label:        r.FLabel,
						Data:         r.FData,
						ValidationID: r.FValidationID,
						UIConfig:     r.FUIConfig,
						Type:         r.FType,
						AuthorID:     r.FAuthorID,
						DateCreated:  r.FDateCreated,
						DateModified: r.FDateModified,
					})
				}
			}
		}

		root := tree.NewRoot()
		stats, treeErr := root.LoadFromAdminData(cd, dt, cf, df)
		if treeErr != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to build admin tree: " + treeErr.Error()}
		}

		return AdminTreeLoadedMsg{RootNode: root, Stats: stats}
	}
}

// ReloadAdminContentTreeByRootIDCmd loads an admin content tree by root_id for standalone/global content.
func ReloadAdminContentTreeByRootIDCmd(cfg *config.Config, rootID types.AdminContentID) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		rootFilter := types.NullableAdminContentID{ID: rootID, Valid: true}

		rows, err := d.ListAdminContentDataWithDatatypeByRootID(rootFilter)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to load admin content by root_id: " + err.Error()}
		}

		var cd []db.AdminContentData
		var dt []db.AdminDatatypes
		dtSeen := make(map[types.AdminDatatypeID]bool)

		if rows != nil {
			for _, r := range *rows {
				cd = append(cd, db.AdminContentData{
					AdminContentDataID: r.AdminContentDataID,
					ParentID:           r.ParentID,
					FirstChildID:       r.FirstChildID,
					NextSiblingID:      r.NextSiblingID,
					PrevSiblingID:      r.PrevSiblingID,
					AdminRouteID:       r.AdminRouteID,
					RootID:             r.RootID,
					AdminDatatypeID:    r.AdminDatatypeID,
					AuthorID:           r.AuthorID,
					Status:             r.Status,
					DateCreated:        r.DateCreated,
					DateModified:       r.DateModified,
				})
				if !r.DtAdminDatatypeID.IsZero() && !dtSeen[r.DtAdminDatatypeID] {
					dtSeen[r.DtAdminDatatypeID] = true
					dt = append(dt, db.AdminDatatypes{
						AdminDatatypeID: r.DtAdminDatatypeID,
						ParentID:        r.DtParentID,
						Label:           r.DtLabel,
						Type:            r.DtType,
						AuthorID:        r.DtAuthorID,
						DateCreated:     r.DtDateCreated,
						DateModified:    r.DtDateModified,
					})
				}
			}
		}

		cfRows, cfErr := d.ListAdminContentFieldsByRootID(rootFilter)
		if cfErr != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to load admin content fields by root_id: " + cfErr.Error()}
		}

		var cf []db.AdminContentFields
		var df []db.AdminFields
		dfSeen := make(map[types.AdminFieldID]bool)

		if cfRows != nil {
			cf = *cfRows
			for _, contentField := range cf {
				if contentField.AdminFieldID.Valid && !dfSeen[contentField.AdminFieldID.ID] {
					dfSeen[contentField.AdminFieldID.ID] = true
					field, fieldErr := d.GetAdminField(contentField.AdminFieldID.ID)
					if fieldErr == nil && field != nil {
						df = append(df, *field)
					}
				}
			}
		}

		root := tree.NewRoot()
		stats, treeErr := root.LoadFromAdminData(cd, dt, cf, df)
		if treeErr != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to build admin tree: " + treeErr.Error()}
		}

		return AdminTreeLoadedMsg{RootNode: root, Stats: stats}
	}
}

// LoadAdminContentFieldsCmd fetches admin content fields for display.
// When adminDatatypeID is valid, it merges content field values with the
// canonical field list from the datatype (showing all fields including empty ones).
// When adminDatatypeID is not valid, it returns only fields that have stored values
// by using the join query.
func LoadAdminContentFieldsCmd(cfg *config.Config, adminContentDataID types.AdminContentID, adminDatatypeID types.NullableAdminDatatypeID, locale string) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		contentFilter := types.NullableAdminContentID{ID: adminContentDataID, Valid: true}

		if adminDatatypeID.Valid {
			// Full mode: merge content values with canonical field list
			contentFields, err := d.ListAdminContentFieldsByContentDataAndLocale(contentFilter, locale)
			if err != nil {
				return ActionResultMsg{Title: "Error", Message: "Failed to load admin content fields: " + err.Error()}
			}

			canonicalFields, fieldErr := d.ListAdminFieldsByDatatypeID(adminDatatypeID)
			if fieldErr != nil {
				return ActionResultMsg{Title: "Error", Message: "Failed to load admin fields: " + fieldErr.Error()}
			}

			valueMap := make(map[types.AdminFieldID]db.AdminContentFields)
			if contentFields != nil {
				for _, cf := range *contentFields {
					if cf.AdminFieldID.Valid {
						valueMap[cf.AdminFieldID.ID] = cf
					}
				}
			}

			var display []AdminContentFieldDisplay
			if canonicalFields != nil {
				for _, f := range *canonicalFields {
					fd := AdminContentFieldDisplay{
						AdminFieldID:   f.AdminFieldID,
						Label:          f.Label,
						Type:           string(f.Type),
						ValidationJSON: "", // TODO: resolve from validation table
						DataJSON:       f.Data,
					}
					if cf, ok := valueMap[f.AdminFieldID]; ok {
						fd.AdminContentFieldID = cf.AdminContentFieldID
						fd.Value = cf.AdminFieldValue
					}
					display = append(display, fd)
				}
			}

			return AdminLoadContentFieldsMsg{Fields: display}
		}

		// Lightweight mode: return only fields with stored values via join
		rows, err := d.ListAdminContentFieldsWithFieldByContentData(contentFilter)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to load admin content fields: " + err.Error()}
		}

		var display []AdminContentFieldDisplay
		if rows != nil {
			for _, r := range *rows {
				display = append(display, AdminContentFieldDisplay{
					AdminContentFieldID: r.AdminContentFieldID,
					AdminFieldID:        r.FAdminFieldID,
					Label:               r.FLabel,
					Type:                string(r.FType),
					Value:               r.AdminFieldValue,
					ValidationJSON:      "", // TODO: resolve from validation table
					DataJSON:            r.FData,
				})
			}
		}

		return AdminLoadContentFieldsMsg{Fields: display}
	}
}

// FetchAdminContentForEditCmd fetches admin content fields for editing in a form dialog.
func FetchAdminContentForEditCmd(cfg *config.Config, adminContentID types.AdminContentID, adminDatatypeID types.AdminDatatypeID, adminRouteID types.AdminRouteID, title string, locale string) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		contentFilter := types.NullableAdminContentID{ID: adminContentID, Valid: true}
		contentFields, err := d.ListAdminContentFieldsByContentDataAndLocale(contentFilter, locale)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to fetch admin content fields: " + err.Error()}
		}

		dtFilter := types.NullableAdminDatatypeID{ID: adminDatatypeID, Valid: true}
		canonicalFields, fieldErr := d.ListAdminFieldsByDatatypeID(dtFilter)
		if fieldErr != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to fetch admin fields: " + fieldErr.Error()}
		}

		// Build map of existing values
		valueMap := make(map[types.AdminFieldID]db.AdminContentFields)
		if contentFields != nil {
			for _, cf := range *contentFields {
				if cf.AdminFieldID.Valid {
					valueMap[cf.AdminFieldID.ID] = cf
				}
			}
		}

		var fields []ExistingAdminContentField
		if canonicalFields != nil {
			for _, f := range *canonicalFields {
				ef := ExistingAdminContentField{
					AdminFieldID:   f.AdminFieldID,
					Label:          f.Label,
					Type:           string(f.Type),
					ValidationJSON: "", // TODO: resolve from validation table
					DataJSON:       f.Data,
				}
				if cf, ok := valueMap[f.AdminFieldID]; ok {
					ef.AdminContentFieldID = cf.AdminContentFieldID
					ef.Value = cf.AdminFieldValue
				}
				fields = append(fields, ef)
			}
		}

		return ShowEditAdminContentFormDialogMsg{
			AdminContentID:  adminContentID,
			AdminDatatypeID: adminDatatypeID,
			AdminRouteID:    adminRouteID,
			Fields:          fields,
		}
	}
}

// FetchAdminRootDatatypesCmd fetches root-level admin datatypes (those without parents).
func FetchAdminRootDatatypesCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		all, err := d.ListAdminDatatypes()
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: "Failed to load admin datatypes: " + err.Error()}
		}

		var roots []db.AdminDatatypes
		if all != nil {
			for _, dt := range *all {
				if !dt.ParentID.Valid || dt.ParentID.ID.IsZero() {
					roots = append(roots, dt)
				}
			}
		}

		return AdminRootDatatypesFetchResultsMsg{RootDatatypes: roots}
	}
}

// =============================================================================
// ADMIN CONTENT SIMPLE CMD CONSTRUCTORS
// =============================================================================

// AdminReorderSiblingCmd creates a command to reorder admin content siblings.
func AdminReorderSiblingCmd(adminContentID types.AdminContentID, adminRouteID types.AdminRouteID, direction string) tea.Cmd {
	return func() tea.Msg {
		return AdminReorderSiblingRequestMsg{AdminContentID: adminContentID, AdminRouteID: adminRouteID, Direction: direction}
	}
}

// AdminCopyContentCmd creates a command to copy admin content.
func AdminCopyContentCmd(sourceID types.AdminContentID, adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg {
		return AdminCopyContentRequestMsg{SourceID: sourceID, AdminRouteID: adminRouteID}
	}
}

// AdminTogglePublishCmd creates a command to toggle admin content publish status.
func AdminTogglePublishCmd(adminContentID types.AdminContentID, adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg {
		return AdminTogglePublishRequestMsg{AdminContentID: adminContentID, AdminRouteID: adminRouteID}
	}
}

// AdminListVersionsCmd creates a command to list versions for admin content.
func AdminListVersionsCmd(adminContentID types.AdminContentID, adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg {
		return AdminListVersionsRequestMsg{AdminContentID: adminContentID, AdminRouteID: adminRouteID}
	}
}

// AdminMoveContentCmd creates a command to move admin content.
func AdminMoveContentCmd(sourceID types.AdminContentID, targetID types.AdminContentID, adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg {
		return AdminMoveContentRequestMsg{SourceID: sourceID, TargetID: targetID, AdminRouteID: adminRouteID}
	}
}

// =============================================================================
// ADMIN CONTENT DIALOG-TRIGGERING CMD CONSTRUCTORS
// =============================================================================

// ShowDeleteAdminContentDialogCmd creates a command to show the delete admin content dialog.
func ShowDeleteAdminContentDialogCmd(adminContentID types.AdminContentID, adminRouteID types.AdminRouteID, name string, hasChildren bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminContentDialogMsg{
			AdminContentID: adminContentID,
			AdminRouteID:   adminRouteID,
			ContentName:    name,
			HasChildren:    hasChildren,
		}
	}
}

// ShowPublishAdminContentDialogCmd creates a command to show the publish admin content dialog.
func ShowPublishAdminContentDialogCmd(adminContentID types.AdminContentID, adminRouteID types.AdminRouteID, name string, isPublished bool) tea.Cmd {
	return func() tea.Msg {
		return ShowPublishAdminContentDialogMsg{
			AdminContentID: adminContentID,
			AdminRouteID:   adminRouteID,
			Name:           name,
			IsPublished:    isPublished,
		}
	}
}

// ShowRestoreAdminVersionDialogCmd creates a command to show the restore admin version dialog.
func ShowRestoreAdminVersionDialogCmd(adminContentID types.AdminContentID, versionID types.AdminContentVersionID, adminRouteID types.AdminRouteID, versionNumber int64) tea.Cmd {
	return func() tea.Msg {
		return ShowRestoreAdminVersionDialogMsg{
			AdminContentID: adminContentID,
			VersionID:      versionID,
			AdminRouteID:   adminRouteID,
			VersionNumber:  versionNumber,
		}
	}
}

// ShowDeleteAdminContentFieldDialogCmd creates a command to show the delete admin content field dialog.
func ShowDeleteAdminContentFieldDialogCmd(adminContentFieldID types.AdminContentFieldID, adminContentID types.AdminContentID, adminRouteID types.AdminRouteID, adminDatatypeID types.NullableAdminDatatypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminContentFieldDialogMsg{
			AdminContentFieldID: adminContentFieldID,
			AdminContentID:      adminContentID,
			AdminRouteID:        adminRouteID,
			AdminDatatypeID:     adminDatatypeID,
			Label:               label,
		}
	}
}

// ShowMoveAdminContentDialogCmd creates a command to show the move admin content dialog.
func ShowMoveAdminContentDialogCmd(node *tree.Node, adminRouteID types.AdminRouteID, targets []ParentOption) tea.Cmd {
	return func() tea.Msg {
		return ShowMoveAdminContentDialogMsg{
			SourceNode:   node,
			AdminRouteID: adminRouteID,
			Targets:      targets,
		}
	}
}

// AdminFetchFieldsForFormCmd fetches admin fields for a datatype, then emits
// AdminBuildContentFormMsg so the dialog system can build the creation form.
func AdminFetchFieldsForFormCmd(d db.DbDriver, adminDatatypeID types.AdminDatatypeID, adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg {
		if d == nil {
			return ActionResultMsg{Title: "Error", Message: "Database not available"}
		}
		dtFilter := types.NullableAdminDatatypeID{ID: adminDatatypeID, Valid: true}
		fields, err := d.ListAdminFieldsByDatatypeID(dtFilter)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to fetch fields: %v", err)}
		}
		var fieldList []db.AdminFields
		if fields != nil {
			fieldList = *fields
		}
		return AdminBuildContentFormMsg{
			AdminDatatypeID: adminDatatypeID,
			AdminRouteID:    adminRouteID,
			Fields:          fieldList,
		}
	}
}

// =============================================================================
// ADMIN CONTENT FIELD DIALOG CMDS
// =============================================================================

// ShowEditAdminSingleFieldDialogMsg requests showing an edit dialog for a single admin content field.
type ShowEditAdminSingleFieldDialogMsg struct {
	Field           AdminContentFieldDisplay
	AdminContentID  types.AdminContentID
	AdminRouteID    types.AdminRouteID
	AdminDatatypeID types.NullableAdminDatatypeID
}

// ShowEditAdminSingleFieldDialogCmd creates a command to show the edit admin single field dialog.
func ShowEditAdminSingleFieldDialogCmd(cf AdminContentFieldDisplay, adminContentID types.AdminContentID, adminRouteID types.AdminRouteID, adminDatatypeID types.NullableAdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowEditAdminSingleFieldDialogMsg{
			Field:           cf,
			AdminContentID:  adminContentID,
			AdminRouteID:    adminRouteID,
			AdminDatatypeID: adminDatatypeID,
		}
	}
}

// ShowAddAdminContentFieldDialogMsg requests showing a picker to add admin content fields.
type ShowAddAdminContentFieldDialogMsg struct {
	Options         []huh.Option[string]
	AdminContentID  types.AdminContentID
	AdminRouteID    types.AdminRouteID
	AdminDatatypeID types.NullableAdminDatatypeID
}

// ShowAddAdminContentFieldDialogCmd creates a command to show the add admin content field picker.
func ShowAddAdminContentFieldDialogCmd(options []huh.Option[string], adminContentID types.AdminContentID, adminRouteID types.AdminRouteID, adminDatatypeID types.NullableAdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowAddAdminContentFieldDialogMsg{
			Options:         options,
			AdminContentID:  adminContentID,
			AdminRouteID:    adminRouteID,
			AdminDatatypeID: adminDatatypeID,
		}
	}
}

// =============================================================================
// DIALOG-TRIGGERING CONSTRUCTORS (relocated from admin_controls.go)
// =============================================================================

// ShowEditAdminRouteDialogCmd creates a command to show the edit admin route dialog.
func ShowEditAdminRouteDialogCmd(route db.AdminRoutes) tea.Cmd {
	return func() tea.Msg { return ShowEditAdminRouteDialogMsg{Route: route} }
}

// ShowAdminFormDialogCmd creates a command to show an admin form dialog (e.g., new datatype).
func ShowAdminFormDialogCmd(action FormDialogAction, title string, parents []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowAdminFormDialogMsg{Action: action, Title: title, Parents: parents}
	}
}

// ShowEditAdminDatatypeDialogCmd creates a command to show the edit admin datatype dialog.
func ShowEditAdminDatatypeDialogCmd(dt db.AdminDatatypes, allDatatypes []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowEditAdminDatatypeDialogMsg{Datatype: dt, Parents: allDatatypes}
	}
}

// ShowEditAdminFieldDialogCmd creates a command to show the edit admin field dialog.
func ShowEditAdminFieldDialogCmd(field db.AdminFields) tea.Cmd {
	return func() tea.Msg { return ShowEditAdminFieldDialogMsg{Field: field} }
}

// ShowEditFieldTypeDialogCmd creates a command to show the edit field type dialog.
func ShowEditFieldTypeDialogCmd(ft db.FieldTypes) tea.Cmd {
	return func() tea.Msg { return ShowEditFieldTypeDialogMsg{FieldType: ft} }
}

// ShowEditAdminFieldTypeDialogCmd creates a command to show the edit admin field type dialog.
func ShowEditAdminFieldTypeDialogCmd(ft db.AdminFieldTypes) tea.Cmd {
	return func() tea.Msg { return ShowEditAdminFieldTypeDialogMsg{AdminFieldType: ft} }
}

// CreateAdminRouteWithContentCmd creates a command to create an admin route with initial content.
func CreateAdminRouteWithContentCmd(title, slug, adminDatatypeID string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminRouteWithContentRequestMsg{
			Title:           title,
			Slug:            slug,
			AdminDatatypeID: adminDatatypeID,
		}
	}
}
