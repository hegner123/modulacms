package router

import (
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// HealReport is the JSON response from the content heal endpoint.
type HealReport struct {
	DryRun              bool         `json:"dry_run"`
	ContentDataScanned  int          `json:"content_data_scanned"`
	ContentDataRepairs  []HealRepair `json:"content_data_repairs"`
	ContentFieldScanned int          `json:"content_field_scanned"`
	ContentFieldRepairs []HealRepair `json:"content_field_repairs"`
}

// HealRepair records a single ID repair made (or that would be made in dry-run).
type HealRepair struct {
	RowID    string `json:"row_id"`
	Column   string `json:"column"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// ContentHealHandler handles POST /api/v1/admin/content/heal.
func ContentHealHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	apiHealContent(w, r, c)
}

func apiHealContent(w http.ResponseWriter, r *http.Request, c config.Config) {
	dryRun := r.URL.Query().Get("dry_run") == "true"
	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)
	d := db.ConfigDB(c)

	report := HealReport{
		DryRun:              dryRun,
		ContentDataRepairs:  []HealRepair{},
		ContentFieldRepairs: []HealRepair{},
	}

	// --- Heal content_data rows ---
	contentRows, err := d.ListContentData()
	if err != nil {
		utility.DefaultLogger.Error("heal: failed to list content_data", err)
		http.Error(w, fmt.Sprintf("failed to list content_data: %v", err), http.StatusInternalServerError)
		return
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
			_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
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

	// --- Heal content_field rows ---
	fieldRows, err := d.ListContentFields()
	if err != nil {
		utility.DefaultLogger.Error("heal: failed to list content_fields", err)
		http.Error(w, fmt.Sprintf("failed to list content_fields: %v", err), http.StatusInternalServerError)
		return
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
			_, updateErr := d.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
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

	writeJSON(w, report)
}

// healContentDataRow validates all ID columns on a content_data row and returns
// a list of repairs plus the healed row. If no repairs are needed, repairs is empty.
func healContentDataRow(row db.ContentData) (repairs []HealRepair, healed db.ContentData) {
	healed = row
	rowID := string(row.ContentDataID)

	// ContentDataID (non-nullable PK)
	if err := row.ContentDataID.Validate(); err != nil {
		newID := types.NewContentID()
		repairs = append(repairs, HealRepair{
			RowID:    rowID,
			Column:   "content_data_id",
			OldValue: string(row.ContentDataID),
			NewValue: string(newID),
		})
		healed.ContentDataID = newID
	}

	// ParentID (nullable)
	if row.ParentID.Valid {
		if err := row.ParentID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "parent_id",
				OldValue: string(row.ParentID.ID),
				NewValue: "null",
			})
			healed.ParentID = types.NullableContentID{Valid: false}
		}
	}

	// FirstChildID (nullable)
	if row.FirstChildID.Valid {
		if err := row.FirstChildID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "first_child_id",
				OldValue: string(row.FirstChildID.ID),
				NewValue: "null",
			})
			healed.FirstChildID = types.NullableContentID{Valid: false}
		}
	}

	// NextSiblingID (nullable)
	if row.NextSiblingID.Valid {
		if err := row.NextSiblingID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "next_sibling_id",
				OldValue: string(row.NextSiblingID.ID),
				NewValue: "null",
			})
			healed.NextSiblingID = types.NullableContentID{Valid: false}
		}
	}

	// PrevSiblingID (nullable)
	if row.PrevSiblingID.Valid {
		if err := row.PrevSiblingID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "prev_sibling_id",
				OldValue: string(row.PrevSiblingID.ID),
				NewValue: "null",
			})
			healed.PrevSiblingID = types.NullableContentID{Valid: false}
		}
	}

	// RouteID (nullable)
	if row.RouteID.Valid {
		if err := row.RouteID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "route_id",
				OldValue: string(row.RouteID.ID),
				NewValue: "null",
			})
			healed.RouteID = types.NullableRouteID{Valid: false}
		}
	}

	// DatatypeID (nullable)
	if row.DatatypeID.Valid {
		if err := row.DatatypeID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "datatype_id",
				OldValue: string(row.DatatypeID.ID),
				NewValue: "null",
			})
			healed.DatatypeID = types.NullableDatatypeID{Valid: false}
		}
	}

	// AuthorID (non-nullable)
	if err := row.AuthorID.Validate(); err != nil {
		newID := types.NewUserID()
		repairs = append(repairs, HealRepair{
			RowID:    rowID,
			Column:   "author_id",
			OldValue: string(row.AuthorID),
			NewValue: string(newID),
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

	// ContentFieldID (non-nullable PK)
	if err := row.ContentFieldID.Validate(); err != nil {
		newID := types.NewContentFieldID()
		repairs = append(repairs, HealRepair{
			RowID:    rowID,
			Column:   "content_field_id",
			OldValue: string(row.ContentFieldID),
			NewValue: string(newID),
		})
		healed.ContentFieldID = newID
	}

	// RouteID (nullable)
	if row.RouteID.Valid {
		if err := row.RouteID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "route_id",
				OldValue: string(row.RouteID.ID),
				NewValue: "null",
			})
			healed.RouteID = types.NullableRouteID{Valid: false}
		}
	}

	// ContentDataID (nullable)
	if row.ContentDataID.Valid {
		if err := row.ContentDataID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "content_data_id",
				OldValue: string(row.ContentDataID.ID),
				NewValue: "null",
			})
			healed.ContentDataID = types.NullableContentID{Valid: false}
		}
	}

	// FieldID (nullable)
	if row.FieldID.Valid {
		if err := row.FieldID.ID.Validate(); err != nil {
			repairs = append(repairs, HealRepair{
				RowID:    rowID,
				Column:   "field_id",
				OldValue: string(row.FieldID.ID),
				NewValue: "null",
			})
			healed.FieldID = types.NullableFieldID{Valid: false}
		}
	}

	// AuthorID (non-nullable)
	if err := row.AuthorID.Validate(); err != nil {
		newID := types.NewUserID()
		repairs = append(repairs, HealRepair{
			RowID:    rowID,
			Column:   "author_id",
			OldValue: string(row.AuthorID),
			NewValue: string(newID),
		})
		healed.AuthorID = newID
	}

	return repairs, healed
}
