package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminHealReport is the result from AdminContentService.Heal.
type AdminHealReport struct {
	DryRun              bool                        `json:"dry_run"`
	ContentDataScanned  int                         `json:"content_data_scanned"`
	ContentDataRepairs  []HealRepair                `json:"content_data_repairs"`
	ContentFieldScanned int                         `json:"content_field_scanned"`
	ContentFieldRepairs []HealRepair                `json:"content_field_repairs"`
	MissingFields       []MissingFieldReport        `json:"missing_fields"`
	DuplicateFields     []DuplicateFieldReport      `json:"duplicate_fields"`
	OrphanedFields      []AdminOrphanedFieldReport  `json:"orphaned_fields"`
}

// AdminOrphanedFieldReport records a content_field row that references a field_id
// no longer present in its content_data's datatype.
type AdminOrphanedFieldReport struct {
	ContentFieldID string `json:"content_field_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	Deleted        bool   `json:"deleted"`
}

// Heal performs a 5-pass repair of admin content data:
//  1. Heal invalid ULIDs on admin_content_data rows
//  2. Heal invalid ULIDs on admin_content_field rows
//  3. Remove duplicate admin_content_field rows (keep best by value/date)
//  4. Create missing admin_content_field rows for datatypes
//  5. Remove orphaned admin_content_field rows (field no longer in datatype)
//
// When dryRun is true, the report describes what would be changed without modifying data.
func (s *AdminContentService) Heal(ctx context.Context, ac audited.AuditContext, dryRun bool) (*AdminHealReport, error) {
	report := &AdminHealReport{
		DryRun:              dryRun,
		ContentDataRepairs:  []HealRepair{},
		ContentFieldRepairs: []HealRepair{},
		MissingFields:       []MissingFieldReport{},
		DuplicateFields:     []DuplicateFieldReport{},
		OrphanedFields:      []AdminOrphanedFieldReport{},
	}

	// --- Pass 1: Heal admin_content_data rows ---
	contentRows, err := s.driver.ListAdminContentData()
	if err != nil {
		return nil, fmt.Errorf("admin heal: list admin_content_data: %w", err)
	}
	if contentRows != nil {
		report.ContentDataScanned = len(*contentRows)
		for _, row := range *contentRows {
			repairs, healed := healAdminContentDataRow(row)
			if len(repairs) == 0 {
				continue
			}
			report.ContentDataRepairs = append(report.ContentDataRepairs, repairs...)
			if dryRun {
				continue
			}
			_, updateErr := s.driver.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: healed.AdminContentDataID,
				ParentID:           healed.ParentID,
				FirstChildID:       healed.FirstChildID,
				NextSiblingID:      healed.NextSiblingID,
				PrevSiblingID:      healed.PrevSiblingID,
				AdminRouteID:       healed.AdminRouteID,
				AdminDatatypeID:    healed.AdminDatatypeID,
				AuthorID:           healed.AuthorID,
				Status:             healed.Status,
				DateCreated:        healed.DateCreated,
				DateModified:       types.TimestampNow(),
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to update admin_content_data %s", healed.AdminContentDataID), updateErr)
			}
		}
	}

	// --- Pass 2: Heal admin_content_field rows ---
	fieldRows, err := s.driver.ListAdminContentFields()
	if err != nil {
		return nil, fmt.Errorf("admin heal: list admin_content_fields: %w", err)
	}
	if fieldRows != nil {
		report.ContentFieldScanned = len(*fieldRows)
		for _, row := range *fieldRows {
			repairs, healed := healAdminContentFieldRow(row)
			if len(repairs) == 0 {
				continue
			}
			report.ContentFieldRepairs = append(report.ContentFieldRepairs, repairs...)
			if dryRun {
				continue
			}
			_, updateErr := s.driver.UpdateAdminContentField(ctx, ac, db.UpdateAdminContentFieldParams{
				AdminContentFieldID: healed.AdminContentFieldID,
				AdminRouteID:        healed.AdminRouteID,
				AdminContentDataID:  healed.AdminContentDataID,
				AdminFieldID:        healed.AdminFieldID,
				AdminFieldValue:     healed.AdminFieldValue,
				AuthorID:            healed.AuthorID,
				DateCreated:         healed.DateCreated,
				DateModified:        types.TimestampNow(),
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to update admin_content_field %s", healed.AdminContentFieldID), updateErr)
			}
		}
	}

	// --- Pass 3: Remove duplicate admin_content_field rows ---
	if fieldRows != nil {
		s.healAdminDuplicateFields(ctx, ac, *fieldRows, dryRun, report)
	}

	// --- Pass 4: Create missing admin_content_field rows ---
	if contentRows != nil {
		s.healAdminMissingFields(ctx, ac, *contentRows, dryRun, report)
	}

	// --- Pass 5: Remove orphaned admin_content_field rows ---
	if contentRows != nil && fieldRows != nil {
		s.healAdminOrphanedFields(ctx, ac, *contentRows, *fieldRows, dryRun, report)
	}

	return report, nil
}

// healAdminDuplicateFields groups admin_content_fields by (content_data_id, field_id) and
// deletes duplicates, keeping the row with a non-empty field_value (preferring
// the most recently modified).
func (s *AdminContentService) healAdminDuplicateFields(ctx context.Context, ac audited.AuditContext, fieldRows []db.AdminContentFields, dryRun bool, report *AdminHealReport) {
	type cfKey struct {
		contentDataID string
		fieldID       string
	}
	groups := make(map[cfKey][]db.AdminContentFields)
	for _, row := range fieldRows {
		if !row.AdminContentDataID.Valid || !row.AdminFieldID.Valid {
			continue
		}
		k := cfKey{
			contentDataID: string(row.AdminContentDataID.ID),
			fieldID:       string(row.AdminFieldID.ID),
		}
		groups[k] = append(groups[k], row)
	}
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		keepIdx := 0
		for i := 1; i < len(group); i++ {
			keepHasValue := group[keepIdx].AdminFieldValue != ""
			curHasValue := group[i].AdminFieldValue != ""
			if curHasValue && !keepHasValue {
				keepIdx = i
				continue
			}
			if curHasValue == keepHasValue && group[i].DateModified.Time.After(group[keepIdx].DateModified.Time) {
				keepIdx = i
			}
		}
		for i, row := range group {
			if i == keepIdx {
				continue
			}
			entry := DuplicateFieldReport{
				ContentFieldID: string(row.AdminContentFieldID),
				ContentDataID:  string(row.AdminContentDataID.ID),
				FieldID:        string(row.AdminFieldID.ID),
				Deleted:        !dryRun,
			}
			if !dryRun {
				delErr := s.driver.DeleteAdminContentField(ctx, ac, row.AdminContentFieldID)
				if delErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to delete duplicate admin_content_field %s", row.AdminContentFieldID), delErr)
					entry.Deleted = false
				}
			}
			report.DuplicateFields = append(report.DuplicateFields, entry)
		}
	}
}

// healAdminMissingFields checks each admin_content_data row with a datatype and
// creates any missing admin_content_field rows.
func (s *AdminContentService) healAdminMissingFields(ctx context.Context, ac audited.AuditContext, contentRows []db.AdminContentData, dryRun bool, report *AdminHealReport) {
	for _, row := range contentRows {
		if !row.AdminDatatypeID.Valid {
			continue
		}
		fieldList, dtErr := s.driver.ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID{ID: row.AdminDatatypeID.ID, Valid: true})
		if dtErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to list fields for %s", row.AdminDatatypeID.ID), dtErr)
			continue
		}
		if fieldList == nil || len(*fieldList) == 0 {
			continue
		}

		existingFields, efErr := s.driver.ListAdminContentFieldsByContentData(types.NullableAdminContentID{ID: row.AdminContentDataID, Valid: true})
		if efErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to list content fields for %s", row.AdminContentDataID), efErr)
			continue
		}

		existingFieldIDs := make(map[types.AdminFieldID]bool)
		if existingFields != nil {
			for _, cf := range *existingFields {
				if cf.AdminFieldID.Valid {
					existingFieldIDs[cf.AdminFieldID.ID] = true
				}
			}
		}

		now := types.TimestampNow()
		for _, field := range *fieldList {
			if existingFieldIDs[field.AdminFieldID] {
				continue
			}
			entry := MissingFieldReport{
				ContentDataID: string(row.AdminContentDataID),
				FieldID:       string(field.AdminFieldID),
				Created:       !dryRun,
			}
			if !dryRun {
				_, cfErr := s.driver.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
					AdminRouteID:       row.AdminRouteID,
					AdminContentDataID: types.NullableAdminContentID{ID: row.AdminContentDataID, Valid: true},
					AdminFieldID:       types.NullableAdminFieldID{ID: field.AdminFieldID, Valid: true},
					AdminFieldValue:    "",
					AuthorID:           row.AuthorID,
					DateCreated:        now,
					DateModified:       now,
				})
				if cfErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to create missing admin content field for %s field %s", row.AdminContentDataID, field.AdminFieldID), cfErr)
					entry.Created = false
				}
			}
			report.MissingFields = append(report.MissingFields, entry)
		}
	}
}

// healAdminOrphanedFields finds admin_content_field rows that reference a field_id
// no longer present in the content_data's datatype, and removes them.
func (s *AdminContentService) healAdminOrphanedFields(ctx context.Context, ac audited.AuditContext, contentRows []db.AdminContentData, fieldRows []db.AdminContentFields, dryRun bool, report *AdminHealReport) {
	// Build a map of content_data_id → datatype_id
	contentDatatypes := make(map[types.AdminContentID]types.AdminDatatypeID, len(contentRows))
	for _, row := range contentRows {
		if row.AdminDatatypeID.Valid {
			contentDatatypes[row.AdminContentDataID] = row.AdminDatatypeID.ID
		}
	}

	// Cache datatype → field IDs
	datatypeFields := make(map[types.AdminDatatypeID]map[types.AdminFieldID]bool)

	for _, cf := range fieldRows {
		if !cf.AdminContentDataID.Valid || !cf.AdminFieldID.Valid {
			continue
		}
		dtID, hasDT := contentDatatypes[cf.AdminContentDataID.ID]
		if !hasDT {
			continue
		}

		validFields, cached := datatypeFields[dtID]
		if !cached {
			fl, flErr := s.driver.ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID{ID: dtID, Valid: true})
			if flErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to list fields for datatype %s", dtID), flErr)
				continue
			}
			validFields = make(map[types.AdminFieldID]bool)
			if fl != nil {
				for _, f := range *fl {
					validFields[f.AdminFieldID] = true
				}
			}
			datatypeFields[dtID] = validFields
		}

		if validFields[cf.AdminFieldID.ID] {
			continue
		}

		entry := AdminOrphanedFieldReport{
			ContentFieldID: string(cf.AdminContentFieldID),
			ContentDataID:  string(cf.AdminContentDataID.ID),
			FieldID:        string(cf.AdminFieldID.ID),
			Deleted:        !dryRun,
		}
		if !dryRun {
			delErr := s.driver.DeleteAdminContentField(ctx, ac, cf.AdminContentFieldID)
			if delErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("admin heal: failed to delete orphaned admin_content_field %s", cf.AdminContentFieldID), delErr)
				entry.Deleted = false
			}
		}
		report.OrphanedFields = append(report.OrphanedFields, entry)
	}
}

// healAdminContentDataRow validates all ID columns on an admin_content_data row.
func healAdminContentDataRow(row db.AdminContentData) (repairs []HealRepair, healed db.AdminContentData) {
	healed = row
	rowID := string(row.AdminContentDataID)

	if err := row.AdminContentDataID.Validate(); err != nil {
		newID := types.NewAdminContentID()
		repairs = append(repairs, HealRepair{
			RowID: rowID, Column: "admin_content_data_id",
			OldValue: string(row.AdminContentDataID), NewValue: string(newID),
		})
		healed.AdminContentDataID = newID
	}

	if row.ParentID.Valid {
		if err := row.ParentID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "parent_id",
				OldValue: string(row.ParentID.ID), NewValue: "null",
			})
			healed.ParentID = types.NullableAdminContentID{Valid: false}
		}
	}

	if row.FirstChildID.Valid {
		if err := row.FirstChildID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "first_child_id",
				OldValue: string(row.FirstChildID.ID), NewValue: "null",
			})
			healed.FirstChildID = types.NullableAdminContentID{Valid: false}
		}
	}

	if row.NextSiblingID.Valid {
		if err := row.NextSiblingID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "next_sibling_id",
				OldValue: string(row.NextSiblingID.ID), NewValue: "null",
			})
			healed.NextSiblingID = types.NullableAdminContentID{Valid: false}
		}
	}

	if row.PrevSiblingID.Valid {
		if err := row.PrevSiblingID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "prev_sibling_id",
				OldValue: string(row.PrevSiblingID.ID), NewValue: "null",
			})
			healed.PrevSiblingID = types.NullableAdminContentID{Valid: false}
		}
	}

	if row.AdminRouteID.Valid {
		if err := row.AdminRouteID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "admin_route_id",
				OldValue: string(row.AdminRouteID.ID), NewValue: "null",
			})
			healed.AdminRouteID = types.NullableAdminRouteID{Valid: false}
		}
	}

	if row.AdminDatatypeID.Valid {
		if err := row.AdminDatatypeID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "admin_datatype_id",
				OldValue: string(row.AdminDatatypeID.ID), NewValue: "null",
			})
			healed.AdminDatatypeID = types.NullableAdminDatatypeID{Valid: false}
		}
	}

	if err := row.AuthorID.Validate(); err != nil {
		newID := types.NewUserID()
		repairs = append(repairs, HealRepair{
			RowID: rowID, Column: "author_id",
			OldValue: string(row.AuthorID), NewValue: string(newID),
		})
		healed.AuthorID = newID
	}

	return repairs, healed
}

// healAdminContentFieldRow validates all ID columns on an admin_content_field row.
func healAdminContentFieldRow(row db.AdminContentFields) (repairs []HealRepair, healed db.AdminContentFields) {
	healed = row
	rowID := string(row.AdminContentFieldID)

	if err := row.AdminContentFieldID.Validate(); err != nil {
		newID := types.NewAdminContentFieldID()
		repairs = append(repairs, HealRepair{
			RowID: rowID, Column: "admin_content_field_id",
			OldValue: string(row.AdminContentFieldID), NewValue: string(newID),
		})
		healed.AdminContentFieldID = newID
	}

	if row.AdminRouteID.Valid {
		if err := row.AdminRouteID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "admin_route_id",
				OldValue: string(row.AdminRouteID.ID), NewValue: "null",
			})
			healed.AdminRouteID = types.NullableAdminRouteID{Valid: false}
		}
	}

	if row.AdminContentDataID.Valid {
		if err := row.AdminContentDataID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "admin_content_data_id",
				OldValue: string(row.AdminContentDataID.ID), NewValue: "null",
			})
			healed.AdminContentDataID = types.NullableAdminContentID{Valid: false}
		}
	}

	if row.AdminFieldID.Valid {
		if err := row.AdminFieldID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "admin_field_id",
				OldValue: string(row.AdminFieldID.ID), NewValue: "null",
			})
			healed.AdminFieldID = types.NullableAdminFieldID{Valid: false}
		}
	}

	if err := row.AuthorID.Validate(); err != nil {
		newID := types.NewUserID()
		repairs = append(repairs, HealRepair{
			RowID: rowID, Column: "author_id",
			OldValue: string(row.AuthorID), NewValue: string(newID),
		})
		healed.AuthorID = newID
	}

	return repairs, healed
}
