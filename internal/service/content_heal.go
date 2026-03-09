package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// HealReport is the result from ContentService.Heal.
type HealReport struct {
	DryRun              bool                   `json:"dry_run"`
	ContentDataScanned  int                    `json:"content_data_scanned"`
	ContentDataRepairs  []HealRepair           `json:"content_data_repairs"`
	ContentFieldScanned int                    `json:"content_field_scanned"`
	ContentFieldRepairs []HealRepair           `json:"content_field_repairs"`
	MissingFields       []MissingFieldReport   `json:"missing_fields"`
	DuplicateFields     []DuplicateFieldReport `json:"duplicate_fields"`
}

// HealRepair records a single ID repair made (or that would be made in dry-run).
type HealRepair struct {
	RowID    string `json:"row_id"`
	Column   string `json:"column"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// MissingFieldReport records a content_field row that was created (or would be
// created in dry-run mode) because the parent content_data's datatype requires
// a field that was not present.
type MissingFieldReport struct {
	ContentDataID string `json:"content_data_id"`
	FieldID       string `json:"field_id"`
	Created       bool   `json:"created"`
}

// DuplicateFieldReport records a duplicate content_field row that was deleted
// (or would be deleted in dry-run mode).
type DuplicateFieldReport struct {
	ContentFieldID string `json:"content_field_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	Deleted        bool   `json:"deleted"`
}

// Heal performs a 4-pass repair of content data:
//  1. Heal invalid ULIDs on content_data rows
//  2. Heal invalid ULIDs on content_field rows
//  3. Remove duplicate content_field rows (keep best by value/date)
//  4. Create missing content_field rows for datatypes
//
// When dryRun is true, the report describes what would be changed without modifying data.
func (s *ContentService) Heal(ctx context.Context, ac audited.AuditContext, dryRun bool) (*HealReport, error) {
	report := &HealReport{
		DryRun:              dryRun,
		ContentDataRepairs:  []HealRepair{},
		ContentFieldRepairs: []HealRepair{},
		MissingFields:       []MissingFieldReport{},
		DuplicateFields:     []DuplicateFieldReport{},
	}

	// --- Pass 1: Heal content_data rows ---
	contentRows, err := s.driver.ListContentData()
	if err != nil {
		return nil, fmt.Errorf("heal: list content_data: %w", err)
	}
	if contentRows != nil {
		report.ContentDataScanned = len(*contentRows)
		for _, row := range *contentRows {
			repairs, healed := healContentDataRow(row)
			if len(repairs) == 0 {
				continue
			}
			report.ContentDataRepairs = append(report.ContentDataRepairs, repairs...)
			if dryRun {
				continue
			}
			_, updateErr := s.driver.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: healed.ContentDataID,
				ParentID:      healed.ParentID,
				FirstChildID:  healed.FirstChildID,
				NextSiblingID: healed.NextSiblingID,
				PrevSiblingID: healed.PrevSiblingID,
				RouteID:       healed.RouteID,
				DatatypeID:    healed.DatatypeID,
				AuthorID:      healed.AuthorID,
				Status:        healed.Status,
				DateCreated:   healed.DateCreated,
				DateModified:  types.TimestampNow(),
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to update content_data %s", healed.ContentDataID), updateErr)
			}
		}
	}

	// --- Pass 2: Heal content_field rows ---
	fieldRows, err := s.driver.ListContentFields()
	if err != nil {
		return nil, fmt.Errorf("heal: list content_fields: %w", err)
	}
	if fieldRows != nil {
		report.ContentFieldScanned = len(*fieldRows)
		for _, row := range *fieldRows {
			repairs, healed := healContentFieldRow(row)
			if len(repairs) == 0 {
				continue
			}
			report.ContentFieldRepairs = append(report.ContentFieldRepairs, repairs...)
			if dryRun {
				continue
			}
			_, updateErr := s.driver.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
				ContentFieldID: healed.ContentFieldID,
				RouteID:        healed.RouteID,
				ContentDataID:  healed.ContentDataID,
				FieldID:        healed.FieldID,
				FieldValue:     healed.FieldValue,
				AuthorID:       healed.AuthorID,
				DateCreated:    healed.DateCreated,
				DateModified:   types.TimestampNow(),
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to update content_field %s", healed.ContentFieldID), updateErr)
			}
		}
	}

	// --- Pass 3: Remove duplicate content_field rows ---
	if fieldRows != nil {
		s.healDuplicateFields(ctx, ac, *fieldRows, dryRun, report)
	}

	// --- Pass 4: Create missing content_field rows ---
	if contentRows != nil {
		s.healMissingFields(ctx, ac, *contentRows, dryRun, report)
	}

	return report, nil
}

// healDuplicateFields groups content_fields by (content_data_id, field_id) and
// deletes duplicates, keeping the row with a non-empty field_value (preferring
// the most recently modified).
func (s *ContentService) healDuplicateFields(ctx context.Context, ac audited.AuditContext, fieldRows []db.ContentFields, dryRun bool, report *HealReport) {
	type cfKey struct {
		contentDataID string
		fieldID       string
	}
	groups := make(map[cfKey][]db.ContentFields)
	for _, row := range fieldRows {
		if !row.ContentDataID.Valid || !row.FieldID.Valid {
			continue
		}
		k := cfKey{
			contentDataID: string(row.ContentDataID.ID),
			fieldID:       string(row.FieldID.ID),
		}
		groups[k] = append(groups[k], row)
	}
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		// Pick the best row to keep: prefer non-empty field_value, then latest date_modified.
		keepIdx := 0
		for i := 1; i < len(group); i++ {
			keepHasValue := group[keepIdx].FieldValue != ""
			curHasValue := group[i].FieldValue != ""
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
				ContentFieldID: string(row.ContentFieldID),
				ContentDataID:  string(row.ContentDataID.ID),
				FieldID:        string(row.FieldID.ID),
				Deleted:        !dryRun,
			}
			if !dryRun {
				delErr := s.driver.DeleteContentField(ctx, ac, row.ContentFieldID)
				if delErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to delete duplicate content_field %s", row.ContentFieldID), delErr)
					entry.Deleted = false
				}
			}
			report.DuplicateFields = append(report.DuplicateFields, entry)
		}
	}
}

// healMissingFields checks each content_data row with a datatype and creates
// any missing content_field rows.
func (s *ContentService) healMissingFields(ctx context.Context, ac audited.AuditContext, contentRows []db.ContentData, dryRun bool, report *HealReport) {
	for _, row := range contentRows {
		if !row.DatatypeID.Valid {
			continue
		}
		fieldList, dtErr := s.driver.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: row.DatatypeID.ID, Valid: true})
		if dtErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to list fields for %s", row.DatatypeID.ID), dtErr)
			continue
		}
		if fieldList == nil || len(*fieldList) == 0 {
			continue
		}

		existingFields, efErr := s.driver.ListContentFieldsByContentData(types.NullableContentID{ID: row.ContentDataID, Valid: true})
		if efErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to list content fields for %s", row.ContentDataID), efErr)
			continue
		}

		existingFieldIDs := make(map[types.FieldID]bool)
		if existingFields != nil {
			for _, cf := range *existingFields {
				if cf.FieldID.Valid {
					existingFieldIDs[cf.FieldID.ID] = true
				}
			}
		}

		now := types.TimestampNow()
		for _, field := range *fieldList {
			if existingFieldIDs[field.FieldID] {
				continue
			}
			entry := MissingFieldReport{
				ContentDataID: string(row.ContentDataID),
				FieldID:       string(field.FieldID),
				Created:       !dryRun,
			}
			if !dryRun {
				_, cfErr := s.driver.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					RouteID:       row.RouteID,
					ContentDataID: types.NullableContentID{ID: row.ContentDataID, Valid: true},
					FieldID:       types.NullableFieldID{ID: field.FieldID, Valid: true},
					FieldValue:    "",
					AuthorID:      row.AuthorID,
					DateCreated:   now,
					DateModified:  now,
				})
				if cfErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to create missing content field for %s field %s", row.ContentDataID, field.FieldID), cfErr)
					entry.Created = false
				}
			}
			report.MissingFields = append(report.MissingFields, entry)
		}
	}
}

// healContentDataRow validates all ID columns on a content_data row and returns
// a list of repairs plus the healed row. If no repairs are needed, repairs is empty.
func healContentDataRow(row db.ContentData) (repairs []HealRepair, healed db.ContentData) {
	healed = row
	rowID := string(row.ContentDataID)

	if err := row.ContentDataID.Validate(); err != nil {
		newID := types.NewContentID()
		repairs = append(repairs, HealRepair{
			RowID: rowID, Column: "content_data_id",
			OldValue: string(row.ContentDataID), NewValue: string(newID),
		})
		healed.ContentDataID = newID
	}

	if row.ParentID.Valid {
		if err := row.ParentID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "parent_id",
				OldValue: string(row.ParentID.ID), NewValue: "null",
			})
			healed.ParentID = types.NullableContentID{Valid: false}
		}
	}

	if row.FirstChildID.Valid {
		if err := row.FirstChildID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "first_child_id",
				OldValue: string(row.FirstChildID.ID), NewValue: "null",
			})
			healed.FirstChildID = types.NullableContentID{Valid: false}
		}
	}

	if row.NextSiblingID.Valid {
		if err := row.NextSiblingID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "next_sibling_id",
				OldValue: string(row.NextSiblingID.ID), NewValue: "null",
			})
			healed.NextSiblingID = types.NullableContentID{Valid: false}
		}
	}

	if row.PrevSiblingID.Valid {
		if err := row.PrevSiblingID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "prev_sibling_id",
				OldValue: string(row.PrevSiblingID.ID), NewValue: "null",
			})
			healed.PrevSiblingID = types.NullableContentID{Valid: false}
		}
	}

	if row.RouteID.Valid {
		if err := row.RouteID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "route_id",
				OldValue: string(row.RouteID.ID), NewValue: "null",
			})
			healed.RouteID = types.NullableRouteID{Valid: false}
		}
	}

	if row.DatatypeID.Valid {
		if err := row.DatatypeID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "datatype_id",
				OldValue: string(row.DatatypeID.ID), NewValue: "null",
			})
			healed.DatatypeID = types.NullableDatatypeID{Valid: false}
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

// healContentFieldRow validates all ID columns on a content_field row and returns
// a list of repairs plus the healed row. If no repairs are needed, repairs is empty.
func healContentFieldRow(row db.ContentFields) (repairs []HealRepair, healed db.ContentFields) {
	healed = row
	rowID := string(row.ContentFieldID)

	if err := row.ContentFieldID.Validate(); err != nil {
		newID := types.NewContentFieldID()
		repairs = append(repairs, HealRepair{
			RowID: rowID, Column: "content_field_id",
			OldValue: string(row.ContentFieldID), NewValue: string(newID),
		})
		healed.ContentFieldID = newID
	}

	if row.RouteID.Valid {
		if err := row.RouteID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "route_id",
				OldValue: string(row.RouteID.ID), NewValue: "null",
			})
			healed.RouteID = types.NullableRouteID{Valid: false}
		}
	}

	if row.ContentDataID.Valid {
		if err := row.ContentDataID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "content_data_id",
				OldValue: string(row.ContentDataID.ID), NewValue: "null",
			})
			healed.ContentDataID = types.NullableContentID{Valid: false}
		}
	}

	if row.FieldID.Valid {
		if err := row.FieldID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID: rowID, Column: "field_id",
				OldValue: string(row.FieldID.ID), NewValue: "null",
			})
			healed.FieldID = types.NullableFieldID{Valid: false}
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
