package publishing

import (
	"github.com/hegner123/modulacms/internal/db"
)

// Snapshot holds the serialized content tree data for a published version.
// It stores the raw parallel slices that can be fed back into model.BuildTree
// to reconstruct the content tree on delivery.
type Snapshot struct {
	ContentData   []db.ContentDataJSON       `json:"content_data"`
	Datatypes     []db.DatatypeJSON          `json:"datatypes"`
	ContentFields []SnapshotContentFieldJSON `json:"content_fields"`
	Fields        []db.FieldsJSON            `json:"fields"`
	Route         SnapshotRoute              `json:"route"`
	SchemaVersion int                        `json:"schema_version"`
}

// SnapshotRoute holds route metadata at the time of publish.
type SnapshotRoute struct {
	RouteID string `json:"route_id"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
}

// SnapshotContentFieldJSON is a string-based representation of a content field
// for snapshot serialization. The existing ContentFieldsJSON type is deprecated
// and uses int64 IDs which are incompatible with the ULID-based typed IDs.
type SnapshotContentFieldJSON struct {
	ContentFieldID string `json:"content_field_id"`
	RouteID        string `json:"route_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	FieldValue     string `json:"field_value"`
	Locale         string `json:"locale"`
	AuthorID       string `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
}

// MapSnapshotContentFieldJSON converts a ContentFields to its JSON representation.
func MapSnapshotContentFieldJSON(a db.ContentFields) SnapshotContentFieldJSON {
	return SnapshotContentFieldJSON{
		ContentFieldID: a.ContentFieldID.String(),
		RouteID:        a.RouteID.String(),
		ContentDataID:  a.ContentDataID.String(),
		FieldID:        a.FieldID.String(),
		FieldValue:     a.FieldValue,
		Locale:         a.Locale,
		AuthorID:       a.AuthorID.String(),
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}

// AdminSnapshot holds the serialized admin content tree data for a published version.
type AdminSnapshot struct {
	ContentData   []db.ContentDataJSON            `json:"content_data"`
	Datatypes     []db.DatatypeJSON               `json:"datatypes"`
	ContentFields []AdminSnapshotContentFieldJSON `json:"content_fields"`
	Fields        []db.FieldsJSON                 `json:"fields"`
	Route         AdminSnapshotRoute              `json:"route"`
	SchemaVersion int                             `json:"schema_version"`
}

// AdminSnapshotRoute holds admin route metadata at the time of publish.
type AdminSnapshotRoute struct {
	AdminRouteID string `json:"admin_route_id"`
}

// AdminSnapshotContentFieldJSON is a string-based representation of an admin content field.
type AdminSnapshotContentFieldJSON struct {
	AdminContentFieldID string `json:"admin_content_field_id"`
	AdminRouteID        string `json:"admin_route_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	Locale              string `json:"locale"`
	AuthorID            string `json:"author_id"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
}

// MapAdminSnapshotContentFieldJSON converts an AdminContentFields to its JSON representation.
func MapAdminSnapshotContentFieldJSON(a db.AdminContentFields) AdminSnapshotContentFieldJSON {
	return AdminSnapshotContentFieldJSON{
		AdminContentFieldID: a.AdminContentFieldID.String(),
		AdminRouteID:        a.AdminRouteID.String(),
		AdminContentDataID:  a.AdminContentDataID.String(),
		AdminFieldID:        a.AdminFieldID.String(),
		AdminFieldValue:     a.AdminFieldValue,
		Locale:              a.Locale,
		AuthorID:            a.AuthorID.String(),
		DateCreated:         a.DateCreated.String(),
		DateModified:        a.DateModified.String(),
	}
}

// RestoreResult holds the result of a content restore operation.
type RestoreResult struct {
	FieldsRestored int
	UnmappedFields []string
}
