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
	DryRun              bool                      `json:"dry_run"`
	ContentDataScanned  int                       `json:"content_data_scanned"`
	ContentDataRepairs  []HealRepair              `json:"content_data_repairs"`
	ContentFieldScanned int                       `json:"content_field_scanned"`
	ContentFieldRepairs []HealRepair              `json:"content_field_repairs"`
	MissingFields       []MissingFieldReport      `json:"missing_fields"`
	DuplicateFields     []DuplicateFieldReport    `json:"duplicate_fields"`
	OrphanedFields      []OrphanedFieldReport     `json:"orphaned_fields"`
	DanglingPointers    []DanglingPointerReport   `json:"dangling_pointers"`
	OrphanedRouteRefs   []OrphanedRouteReport     `json:"orphaned_route_refs"`
	UnroutedRoots       []UnroutedRootReport      `json:"unrouted_roots"`
	RootlessContent     []RootlessContentReport   `json:"rootless_content"`
	InvalidUserRefs     []InvalidUserRefReport    `json:"invalid_user_refs"`
}

// OrphanedRouteReport records a content_data row whose route_id references
// a route that no longer exists.
type OrphanedRouteReport struct {
	ContentDataID string `json:"content_data_id"`
	RouteID       string `json:"route_id"`
	Nulled        bool   `json:"nulled"`
}

// UnroutedRootReport records a root content node (_root datatype type)
// that has no route_id set.
type UnroutedRootReport struct {
	ContentDataID string `json:"content_data_id"`
	DatatypeID    string `json:"datatype_id"`
	DatatypeName  string `json:"datatype_name"`
}

// RootlessContentReport records a content_data row assigned to a route
// that has no _root typed node. These rows are inaccessible — the content
// tree cannot be built without a root entry point.
type RootlessContentReport struct {
	ContentDataID string `json:"content_data_id"`
	RouteID       string `json:"route_id"`
	RouteSlug     string `json:"route_slug"`
	DatatypeName  string `json:"datatype_name"`
	Deleted       bool   `json:"deleted"`
}

// OrphanedFieldReport records a content_field row that references a field_id
// no longer present in its content_data's datatype.
type OrphanedFieldReport struct {
	ContentFieldID string `json:"content_field_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	Deleted        bool   `json:"deleted"`
}

// DanglingPointerReport records a tree pointer on a content_data row that
// references a content_data_id that no longer exists.
type DanglingPointerReport struct {
	ContentDataID string `json:"content_data_id"`
	Column        string `json:"column"`
	TargetID      string `json:"target_id"`
	Nulled        bool   `json:"nulled"`
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

// InvalidUserRefReport records a content row whose author_id or published_by
// references a user that does not exist. Healed by reassigning author_id to the
// user performing the heal, or nulling published_by.
type InvalidUserRefReport struct {
	Table    string `json:"table"`
	RowID    string `json:"row_id"`
	Column   string `json:"column"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Repaired bool   `json:"repaired"`
}

// Heal performs a 10-pass repair of content data:
//  1. Heal invalid ULIDs on content_data rows
//  2. Heal invalid ULIDs on content_field rows
//  3. Remove duplicate content_field rows (keep best by value/date)
//  4. Create missing content_field rows for datatypes
//  5. Remove orphaned content_field rows
//  6. Null out dangling tree pointers (parent, first_child, next/prev_sibling)
//  7. Null out route_id references pointing to deleted routes
//  8. Report root content nodes with no route_id
//  9. Delete content on routes that have no _root node (inaccessible trees)
//  10. Reassign content with invalid user refs to the healing user
//
// When dryRun is true, the report describes what would be changed without modifying data.
func (s *ContentService) Heal(ctx context.Context, ac audited.AuditContext, dryRun bool) (*HealReport, error) {
	report := &HealReport{
		DryRun:              dryRun,
		ContentDataRepairs:  []HealRepair{},
		ContentFieldRepairs: []HealRepair{},
		MissingFields:       []MissingFieldReport{},
		DuplicateFields:     []DuplicateFieldReport{},
		OrphanedFields:      []OrphanedFieldReport{},
		DanglingPointers:    []DanglingPointerReport{},
		OrphanedRouteRefs:   []OrphanedRouteReport{},
		UnroutedRoots:       []UnroutedRootReport{},
		RootlessContent:     []RootlessContentReport{},
		InvalidUserRefs:     []InvalidUserRefReport{},
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

	// --- Pass 5: Remove orphaned content_field rows ---
	if contentRows != nil && fieldRows != nil {
		s.healOrphanedFields(ctx, ac, *contentRows, *fieldRows, dryRun, report)
	}

	// --- Pass 6: Null out dangling tree pointers ---
	if contentRows != nil {
		s.healDanglingPointers(ctx, ac, *contentRows, dryRun, report)
	}

	// --- Pass 7: Null out orphaned route references ---
	if contentRows != nil {
		s.healOrphanedRouteRefs(ctx, ac, *contentRows, dryRun, report)
	}

	// --- Pass 8: Report unrouted root nodes ---
	if contentRows != nil {
		s.reportUnroutedRoots(*contentRows, report)
	}

	// --- Pass 9: Delete content on routes with no _root node ---
	if contentRows != nil {
		s.healRootlessContent(ctx, ac, *contentRows, dryRun, report)
	}

	// --- Pass 10: Reassign content with invalid user refs ---
	if contentRows != nil {
		s.healInvalidUserRefs(ctx, ac, *contentRows, fieldRows, dryRun, report)
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

// healOrphanedFields finds content_field rows that reference a field_id
// no longer present in the content_data's datatype, and removes them.
func (s *ContentService) healOrphanedFields(ctx context.Context, ac audited.AuditContext, contentRows []db.ContentData, fieldRows []db.ContentFields, dryRun bool, report *HealReport) {
	// Build a map of content_data_id → datatype_id
	contentDatatypes := make(map[types.ContentID]types.DatatypeID, len(contentRows))
	for _, row := range contentRows {
		if row.DatatypeID.Valid {
			contentDatatypes[row.ContentDataID] = row.DatatypeID.ID
		}
	}

	// Cache datatype → field IDs
	datatypeFields := make(map[types.DatatypeID]map[types.FieldID]bool)

	for _, cf := range fieldRows {
		if !cf.ContentDataID.Valid || !cf.FieldID.Valid {
			continue
		}
		dtID, hasDT := contentDatatypes[cf.ContentDataID.ID]
		if !hasDT {
			continue
		}

		validFields, cached := datatypeFields[dtID]
		if !cached {
			fl, flErr := s.driver.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: dtID, Valid: true})
			if flErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to list fields for datatype %s", dtID), flErr)
				continue
			}
			validFields = make(map[types.FieldID]bool)
			if fl != nil {
				for _, f := range *fl {
					validFields[f.FieldID] = true
				}
			}
			datatypeFields[dtID] = validFields
		}

		if validFields[cf.FieldID.ID] {
			continue
		}

		entry := OrphanedFieldReport{
			ContentFieldID: string(cf.ContentFieldID),
			ContentDataID:  string(cf.ContentDataID.ID),
			FieldID:        string(cf.FieldID.ID),
			Deleted:        !dryRun,
		}
		if !dryRun {
			delErr := s.driver.DeleteContentField(ctx, ac, cf.ContentFieldID)
			if delErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to delete orphaned content_field %s", cf.ContentFieldID), delErr)
				entry.Deleted = false
			}
		}
		report.OrphanedFields = append(report.OrphanedFields, entry)
	}
}

// healDanglingPointers checks each content_data row's tree pointers
// (parent_id, first_child_id, next_sibling_id, prev_sibling_id) and nulls
// out any that reference a content_data_id that does not exist.
func (s *ContentService) healDanglingPointers(ctx context.Context, ac audited.AuditContext, contentRows []db.ContentData, dryRun bool, report *HealReport) {
	// Build a set of all existing content_data IDs.
	existing := make(map[types.ContentID]bool, len(contentRows))
	for _, row := range contentRows {
		existing[row.ContentDataID] = true
	}

	type pointer struct {
		column string
		id     types.ContentID
		valid  bool
	}

	for _, row := range contentRows {
		ptrs := []pointer{
			{"parent_id", row.ParentID.ID, row.ParentID.Valid},
			{"first_child_id", row.FirstChildID.ID, row.FirstChildID.Valid},
			{"next_sibling_id", row.NextSiblingID.ID, row.NextSiblingID.Valid},
			{"prev_sibling_id", row.PrevSiblingID.ID, row.PrevSiblingID.Valid},
		}

		var danglingCols []string
		for _, p := range ptrs {
			if !p.valid {
				continue
			}
			if existing[p.id] {
				continue
			}
			danglingCols = append(danglingCols, p.column)
			report.DanglingPointers = append(report.DanglingPointers, DanglingPointerReport{
				ContentDataID: string(row.ContentDataID),
				Column:        p.column,
				TargetID:      string(p.id),
				Nulled:        !dryRun,
			})
		}

		if len(danglingCols) == 0 || dryRun {
			continue
		}

		// Build the update with dangling pointers nulled out.
		healed := row
		for _, col := range danglingCols {
			switch col {
			case "parent_id":
				healed.ParentID = types.NullableContentID{Valid: false}
			case "first_child_id":
				healed.FirstChildID = types.NullableContentID{Valid: false}
			case "next_sibling_id":
				healed.NextSiblingID = types.NullableContentID{Valid: false}
			case "prev_sibling_id":
				healed.PrevSiblingID = types.NullableContentID{Valid: false}
			}
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
			utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to null dangling pointers on content_data %s", healed.ContentDataID), updateErr)
			// Mark the entries as not nulled.
			for i := len(report.DanglingPointers) - len(danglingCols); i < len(report.DanglingPointers); i++ {
				report.DanglingPointers[i].Nulled = false
			}
		}
	}
}

// healOrphanedRouteRefs finds content_data rows whose route_id references
// a route that no longer exists and nulls the reference.
func (s *ContentService) healOrphanedRouteRefs(ctx context.Context, ac audited.AuditContext, contentRows []db.ContentData, dryRun bool, report *HealReport) {
	routes, err := s.driver.ListRoutes()
	if err != nil {
		utility.DefaultLogger.Error("heal: failed to list routes", err)
		return
	}

	validRoutes := make(map[types.RouteID]bool)
	if routes != nil {
		for _, r := range *routes {
			validRoutes[r.RouteID] = true
		}
	}

	for _, row := range contentRows {
		if !row.RouteID.Valid {
			continue
		}
		if validRoutes[row.RouteID.ID] {
			continue
		}

		entry := OrphanedRouteReport{
			ContentDataID: string(row.ContentDataID),
			RouteID:       string(row.RouteID.ID),
			Nulled:        !dryRun,
		}

		if !dryRun {
			_, updateErr := s.driver.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: row.ContentDataID,
				ParentID:      row.ParentID,
				FirstChildID:  row.FirstChildID,
				NextSiblingID: row.NextSiblingID,
				PrevSiblingID: row.PrevSiblingID,
				RouteID:       types.NullableRouteID{Valid: false},
				DatatypeID:    row.DatatypeID,
				AuthorID:      row.AuthorID,
				Status:        row.Status,
				DateCreated:   row.DateCreated,
				DateModified:  types.TimestampNow(),
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to null orphaned route_id on content_data %s", row.ContentDataID), updateErr)
				entry.Nulled = false
			}
		}

		report.OrphanedRouteRefs = append(report.OrphanedRouteRefs, entry)
	}
}

// reportUnroutedRoots finds root content nodes (_root datatype type) that have
// no route_id set. These are report-only (no automatic repair since the correct
// route cannot be inferred).
func (s *ContentService) reportUnroutedRoots(contentRows []db.ContentData, report *HealReport) {
	// Build a cache of datatype info for root-type checking.
	dtCache := make(map[types.DatatypeID]*db.Datatypes)

	for _, row := range contentRows {
		// Only check rows with no route and a datatype.
		if row.RouteID.Valid || !row.DatatypeID.Valid {
			continue
		}
		// Only check root nodes (no parent).
		if row.ParentID.Valid {
			continue
		}

		dt, cached := dtCache[row.DatatypeID.ID]
		if !cached {
			fetched, fetchErr := s.driver.GetDatatype(row.DatatypeID.ID)
			if fetchErr != nil {
				continue
			}
			dt = fetched
			dtCache[row.DatatypeID.ID] = dt
		}
		if dt == nil {
			continue
		}

		dtType := types.DatatypeType(dt.Type)
		if !dtType.IsRootType() {
			continue
		}

		report.UnroutedRoots = append(report.UnroutedRoots, UnroutedRootReport{
			ContentDataID: string(row.ContentDataID),
			DatatypeID:    string(row.DatatypeID.ID),
			DatatypeName:  dt.Name,
		})
	}
}

// healRootlessContent finds routes that have content_data rows but no _root
// typed node. Content on such routes is inaccessible — the tree cannot be
// built. In heal mode, deletes the orphaned content and its fields.
func (s *ContentService) healRootlessContent(ctx context.Context, ac audited.AuditContext, contentRows []db.ContentData, dryRun bool, report *HealReport) {
	// Group content by route_id
	byRoute := make(map[types.RouteID][]db.ContentData)
	for _, row := range contentRows {
		if !row.RouteID.Valid {
			continue
		}
		byRoute[row.RouteID.ID] = append(byRoute[row.RouteID.ID], row)
	}

	// Cache datatype lookups
	dtCache := make(map[types.DatatypeID]*db.Datatypes)

	// Build route slug lookup
	routes, routeErr := s.driver.ListRoutes()
	slugMap := make(map[types.RouteID]string)
	if routeErr == nil && routes != nil {
		for _, r := range *routes {
			slugMap[r.RouteID] = string(r.Slug)
		}
	}

	for routeID, rows := range byRoute {
		// Check if any row on this route has a _root type
		hasRoot := false
		for _, row := range rows {
			if !row.DatatypeID.Valid {
				continue
			}
			dt, cached := dtCache[row.DatatypeID.ID]
			if !cached {
				fetched, err := s.driver.GetDatatype(row.DatatypeID.ID)
				if err != nil {
					continue
				}
				dt = fetched
				dtCache[row.DatatypeID.ID] = dt
			}
			if dt != nil && types.DatatypeType(dt.Type).IsRootType() {
				hasRoot = true
				break
			}
		}
		if hasRoot {
			continue
		}

		// This route has content but no root node — all content is inaccessible
		slug := slugMap[routeID]
		for _, row := range rows {
			dtName := ""
			if row.DatatypeID.Valid {
				if dt, ok := dtCache[row.DatatypeID.ID]; ok && dt != nil {
					dtName = dt.Name
				}
			}

			entry := RootlessContentReport{
				ContentDataID: string(row.ContentDataID),
				RouteID:       string(routeID),
				RouteSlug:     slug,
				DatatypeName:  dtName,
				Deleted:       !dryRun,
			}

			if !dryRun {
				// Delete content fields first
				fields, _ := s.driver.ListContentFieldsByContentData(types.NullableContentID{ID: row.ContentDataID, Valid: true})
				if fields != nil {
					for _, cf := range *fields {
						delErr := s.driver.DeleteContentField(ctx, ac, cf.ContentFieldID)
						if delErr != nil {
							utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to delete content_field %s for rootless content %s", cf.ContentFieldID, row.ContentDataID), delErr)
						}
					}
				}
				// Delete the content data row
				delErr := s.driver.DeleteContentData(ctx, ac, row.ContentDataID)
				if delErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to delete rootless content_data %s", row.ContentDataID), delErr)
					entry.Deleted = false
				}
			}

			report.RootlessContent = append(report.RootlessContent, entry)
		}
	}
}

// healInvalidUserRefs finds content_data and content_fields rows whose
// author_id or published_by references a user that does not exist. For
// author_id (NOT NULL), reassigns to the user performing the heal. For
// published_by (nullable), nulls it out.
func (s *ContentService) healInvalidUserRefs(ctx context.Context, ac audited.AuditContext, contentRows []db.ContentData, fieldRows *[]db.ContentFields, dryRun bool, report *HealReport) {
	// Build set of valid user IDs.
	users, err := s.driver.ListUsers()
	if err != nil {
		utility.DefaultLogger.Error("heal: failed to list users for user ref check", err)
		return
	}
	validUsers := make(map[types.UserID]bool)
	if users != nil {
		for _, u := range *users {
			validUsers[u.UserID] = true
		}
	}

	healerID := ac.UserID

	// Check content_data author_id and published_by.
	for _, row := range contentRows {
		if !validUsers[row.AuthorID] {
			entry := InvalidUserRefReport{
				Table:    "content_data",
				RowID:    string(row.ContentDataID),
				Column:   "author_id",
				OldValue: string(row.AuthorID),
				NewValue: string(healerID),
				Repaired: !dryRun,
			}
			if !dryRun {
				_, updateErr := s.driver.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
					ContentDataID: row.ContentDataID,
					ParentID:      row.ParentID,
					FirstChildID:  row.FirstChildID,
					NextSiblingID: row.NextSiblingID,
					PrevSiblingID: row.PrevSiblingID,
					RouteID:       row.RouteID,
					DatatypeID:    row.DatatypeID,
					AuthorID:      healerID,
					Status:        row.Status,
					DateCreated:   row.DateCreated,
					DateModified:  types.TimestampNow(),
				})
				if updateErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to reassign author on content_data %s", row.ContentDataID), updateErr)
					entry.Repaired = false
				}
			}
			report.InvalidUserRefs = append(report.InvalidUserRefs, entry)
		}

		if row.PublishedBy.Valid && !validUsers[row.PublishedBy.ID] {
			entry := InvalidUserRefReport{
				Table:    "content_data",
				RowID:    string(row.ContentDataID),
				Column:   "published_by",
				OldValue: string(row.PublishedBy.ID),
				NewValue: "",
				Repaired: !dryRun,
			}
			if !dryRun {
				updateErr := s.driver.UpdateContentDataPublishMeta(ctx, db.UpdateContentDataPublishMetaParams{
					ContentDataID: row.ContentDataID,
					Status:        row.Status,
					PublishedAt:   row.PublishedAt,
					PublishedBy:   types.NullableUserID{},
					DateModified:  types.TimestampNow(),
				})
				if updateErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to null published_by on content_data %s", row.ContentDataID), updateErr)
					entry.Repaired = false
				}
			}
			report.InvalidUserRefs = append(report.InvalidUserRefs, entry)
		}
	}

	// Check content_fields author_id.
	if fieldRows == nil {
		return
	}
	for _, row := range *fieldRows {
		if validUsers[row.AuthorID] {
			continue
		}
		entry := InvalidUserRefReport{
			Table:    "content_fields",
			RowID:    string(row.ContentFieldID),
			Column:   "author_id",
			OldValue: string(row.AuthorID),
			NewValue: string(healerID),
			Repaired: !dryRun,
		}
		if !dryRun {
			_, updateErr := s.driver.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
				ContentFieldID: row.ContentFieldID,
				RouteID:        row.RouteID,
				ContentDataID:  row.ContentDataID,
				FieldID:        row.FieldID,
				FieldValue:     row.FieldValue,
				AuthorID:       healerID,
				DateCreated:    row.DateCreated,
				DateModified:   types.TimestampNow(),
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("heal: failed to reassign author on content_field %s", row.ContentFieldID), updateErr)
				entry.Repaired = false
			}
		}
		report.InvalidUserRefs = append(report.InvalidUserRefs, entry)
	}
}
