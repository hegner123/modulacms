package modula

import "context"

// HealRepair records a single ID repair made (or that would be made in dry-run).
type HealRepair struct {
	RowID    string `json:"row_id"`
	Column   string `json:"column"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// HealReport is the response from POST /api/v1/admin/content/heal.
type HealReport struct {
	DryRun              bool         `json:"dry_run"`
	ContentDataScanned  int          `json:"content_data_scanned"`
	ContentDataRepairs  []HealRepair `json:"content_data_repairs"`
	ContentFieldScanned int          `json:"content_field_scanned"`
	ContentFieldRepairs []HealRepair `json:"content_field_repairs"`
}

// ContentHealResource provides the content tree healing endpoint.
type ContentHealResource struct {
	http *httpClient
}

// Heal scans content_data and content_field rows for malformed IDs and repairs them.
// Pass dryRun=true to preview repairs without writing changes.
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
