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

// OrphanedRouteReport records a content_data row whose route_id references
// a route that no longer exists.
type OrphanedRouteReport struct {
	ContentDataID string `json:"content_data_id"`
	RouteID       string `json:"route_id"`
	Nulled        bool   `json:"nulled"`
}

// UnroutedRootReport records a root content node that has no route_id set.
type UnroutedRootReport struct {
	ContentDataID string `json:"content_data_id"`
	DatatypeID    string `json:"datatype_id"`
	DatatypeName  string `json:"datatype_name"`
}

// RootlessContentReport records a content_data row assigned to a route
// that has no _root typed node.
type RootlessContentReport struct {
	ContentDataID string `json:"content_data_id"`
	RouteID       string `json:"route_id"`
	RouteSlug     string `json:"route_slug"`
	DatatypeName  string `json:"datatype_name"`
	Deleted       bool   `json:"deleted"`
}

// InvalidUserRefReport records a content row whose author_id or published_by
// references a user that does not exist.
type InvalidUserRefReport struct {
	Table    string `json:"table"`
	RowID    string `json:"row_id"`
	Column   string `json:"column"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Repaired bool   `json:"repaired"`
}

// DuplicatePublishedReport records a content_data_id+locale group that had more
// than one published version.
type DuplicatePublishedReport struct {
	ContentDataID string `json:"content_data_id"`
	Locale        string `json:"locale"`
	Count         int    `json:"count"`
	KeptVersionID string `json:"kept_version_id"`
	Repaired      bool   `json:"repaired"`
}

// HealReport is the response from POST /api/v1/admin/content/heal. It
// summarizes the scan results for content_data, content_field, and
// content_versions tables, including repairs and structural issues.
type HealReport struct {
	DryRun              bool                       `json:"dry_run"`
	ContentDataScanned  int                        `json:"content_data_scanned"`
	ContentDataRepairs  []HealRepair               `json:"content_data_repairs"`
	ContentFieldScanned int                        `json:"content_field_scanned"`
	ContentFieldRepairs []HealRepair               `json:"content_field_repairs"`
	MissingFields       []MissingFieldReport       `json:"missing_fields"`
	DuplicateFields     []DuplicateFieldReport     `json:"duplicate_fields"`
	OrphanedFields      []OrphanedFieldReport      `json:"orphaned_fields"`
	DanglingPointers    []DanglingPointerReport    `json:"dangling_pointers"`
	OrphanedRouteRefs   []OrphanedRouteReport      `json:"orphaned_route_refs"`
	UnroutedRoots       []UnroutedRootReport       `json:"unrouted_roots"`
	RootlessContent     []RootlessContentReport    `json:"rootless_content"`
	InvalidUserRefs     []InvalidUserRefReport     `json:"invalid_user_refs"`
	VersionsScanned     int                        `json:"versions_scanned"`
	DuplicatePublished  []DuplicatePublishedReport `json:"duplicate_published"`
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
