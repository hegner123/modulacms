package modula

import "context"

// HealRepair records a single ID repair made (or that would be made in dry-run
// mode). Each repair describes the correction of one malformed ULID value in a
// specific row and column.
type HealRepair struct {
	// RowID is the primary key of the row containing the malformed ID.
	RowID string `json:"row_id"`

	// Column is the database column name that held the malformed value
	// (e.g. "parent_id", "first_child_id").
	Column string `json:"column"`

	// OldValue is the malformed ID string that was found.
	OldValue string `json:"old_value"`

	// NewValue is the corrected ID string. In dry-run mode this is the value
	// that would be written; in live mode it has already been persisted.
	NewValue string `json:"new_value"`
}

// HealReport is the response from POST /api/v1/admin/content/heal. It
// summarizes the scan results for both content_data and content_field tables,
// including how many rows were scanned and which repairs were made (or would
// be made in dry-run mode).
type HealReport struct {
	// DryRun is true when the heal was a preview-only scan. When true, the
	// repairs listed in ContentDataRepairs and ContentFieldRepairs describe
	// changes that would be made but have not been persisted.
	DryRun bool `json:"dry_run"`

	// ContentDataScanned is the total number of content_data rows examined.
	ContentDataScanned int `json:"content_data_scanned"`

	// ContentDataRepairs lists every malformed ID found (and fixed, unless
	// DryRun is true) in the content_data table.
	ContentDataRepairs []HealRepair `json:"content_data_repairs"`

	// ContentFieldScanned is the total number of content_field rows examined.
	ContentFieldScanned int `json:"content_field_scanned"`

	// ContentFieldRepairs lists every malformed ID found (and fixed, unless
	// DryRun is true) in the content_field table.
	ContentFieldRepairs []HealRepair `json:"content_field_repairs"`
}

// ContentHealResource provides the content tree healing endpoint for scanning
// and repairing malformed ULID values in the content_data and content_field
// tables. This is an administrative maintenance operation typically used after
// data migrations or to recover from corruption.
//
// Access this resource via [Client].ContentHeal:
//
//	// Preview repairs without writing
//	report, err := client.ContentHeal.Heal(ctx, true)
//
//	// Apply repairs
//	report, err := client.ContentHeal.Heal(ctx, false)
type ContentHealResource struct {
	http *httpClient
}

// Heal scans all content_data and content_field rows for malformed ULID values
// in ID and pointer columns, and either repairs them or reports what would be
// repaired.
//
// When dryRun is true, the server performs a read-only scan and populates the
// repair lists in the returned [HealReport] without modifying any data. This
// is recommended as a first step before committing repairs.
//
// When dryRun is false, the server writes corrected values to the database.
// Each repair is recorded in the audit log.
//
// Returns an [*ApiError] if the authenticated user lacks admin privileges.
func (h *ContentHealResource) Heal(ctx context.Context, dryRun bool) (*HealReport, error) {
	path := "/api/v1/admin/content/heal"
	if dryRun {
		path += "?dry_run=true"
	}
	var result HealReport
	if err := h.http.post(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
