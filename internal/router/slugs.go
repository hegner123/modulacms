package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/tree/core"
	"github.com/hegner123/modulacms/internal/utility"
)

// SlugHandler dispatches slug-based content delivery requests.
func SlugHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetSlugContent(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetSlugContent serves published snapshot content for public delivery,
// with a preview mode fallback that serves live draft data for authenticated users.
func apiGetSlugContent(w http.ResponseWriter, r *http.Request, c config.Config) error {
	// Check for preview mode.
	if r.URL.Query().Get("preview") == "true" {
		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "preview mode requires authentication", http.StatusForbidden)
			return fmt.Errorf("unauthenticated preview request")
		}
		w.Header().Set("X-Robots-Tag", "noindex")
		return apiGetSlugContentLive(w, r, c)
	}

	// Normal public delivery: serve from published snapshot.
	return apiGetSlugContentPublished(w, r, c)
}

// apiGetSlugContentPublished serves content from published snapshots.
// It looks up the route, finds the root content data, retrieves the published
// snapshot, deserializes it, and builds the tree for response.
func apiGetSlugContentPublished(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/content")
	if slug == "" {
		slug = "/"
	}

	// 1. Look up route by slug.
	route, err := d.GetRouteID(slug)
	if err != nil {
		utility.DefaultLogger.Error("GetRouteID failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// 2. Find the root content data for this route (the one with no parent).
	nullableRoute := types.NullableRouteID{ID: *route, Valid: true}
	contentData, err := d.ListContentDataByRoute(nullableRoute)
	if err != nil {
		utility.DefaultLogger.Error("ListContentDataByRoute failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	var rootContentDataID types.ContentID
	found := false
	for _, cd := range *contentData {
		if !cd.ParentID.Valid {
			rootContentDataID = cd.ContentDataID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "no root content data found for route", http.StatusNotFound)
		return fmt.Errorf("no root content data for slug %s", slug)
	}

	// 3. Get the published snapshot.
	version, err := d.GetPublishedSnapshot(rootContentDataID, "")
	if err != nil {
		utility.DefaultLogger.Error("GetPublishedSnapshot failed", err)
		http.Error(w, "content not published", http.StatusNotFound)
		return err
	}

	// 4. Deserialize the snapshot JSON.
	var snapshot Snapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		utility.DefaultLogger.Error("snapshot unmarshal failed", err)
		http.Error(w, "failed to read published content", http.StatusInternalServerError)
		return fmt.Errorf("unmarshal snapshot: %w", err)
	}

	// 5. Convert snapshot JSON types back to DB types for model.BuildTree.
	cdSlice, err := snapshotContentDataToSlice(snapshot.ContentData)
	if err != nil {
		utility.DefaultLogger.Error("snapshot content data conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return err
	}

	dtSlice, err := snapshotDatatypesToSlice(snapshot.Datatypes)
	if err != nil {
		utility.DefaultLogger.Error("snapshot datatypes conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return err
	}

	cfSlice, err := snapshotContentFieldsToSlice(snapshot.ContentFields)
	if err != nil {
		utility.DefaultLogger.Error("snapshot content fields conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return err
	}

	fdSlice, err := snapshotFieldsToSlice(snapshot.Fields)
	if err != nil {
		utility.DefaultLogger.Error("snapshot fields conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return err
	}

	// 6. Build the tree from snapshot data.
	root, err := model.BuildTree(utility.DefaultLogger, cdSlice, dtSlice, cfSlice, fdSlice)
	if err != nil {
		utility.DefaultLogger.Error("BuildTree error from snapshot", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// 7. Compose referenced subtrees using published snapshots.
	if root.CoreRoot != nil {
		fetcher := core.NewCachedFetcher(&SnapshotTreeFetcher{Driver: d})
		composeErr := core.ComposeTrees(r.Context(), root.CoreRoot, fetcher, core.ComposeOptions{
			MaxDepth:       c.CompositionMaxDepth(),
			MaxConcurrency: 10,
		})
		if composeErr != nil {
			utility.DefaultLogger.Warn("snapshot composition error", composeErr)
		}
		root.RebuildFromCore()
	}

	// 8. Apply format/transform the same way as the live flow.
	return applyFormatAndTransform(w, r, c, d, root)
}

// apiGetSlugContentLive serves live draft content directly from the database.
// This is the original content delivery path, now used only for preview mode.
func apiGetSlugContentLive(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/content")
	if slug == "" {
		slug = "/"
	}

	route, err := d.GetRouteID(slug)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	nullableRoute := types.NullableRouteID{ID: *route, Valid: true}
	contentData, err := d.ListContentDataByRoute(nullableRoute)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	dataSlice := *contentData

	// Fetch datatype definitions for each content data node.
	dt := []db.Datatypes{}
	for _, da := range dataSlice {
		if !da.DatatypeID.Valid {
			continue
		}
		datatype, err := d.GetDatatype(da.DatatypeID.ID)
		if err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		dt = append(dt, *datatype)
	}

	// Fetch existing content field values for this route.
	contentFields, err := d.ListContentFieldsByRoute(nullableRoute)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Build parallel slices of content fields and field definitions,
	// starting with fields that already have content values.
	var allCF []db.ContentFields
	var allFD []db.Fields
	for _, cf := range *contentFields {
		if !cf.FieldID.Valid {
			continue
		}
		field, err := d.GetField(cf.FieldID.ID)
		if err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		allCF = append(allCF, cf)
		allFD = append(allFD, *field)
	}

	// Track which (content_data_id, field_id) pairs already have values.
	type fieldKey struct{ contentDataID, fieldID string }
	existing := make(map[fieldKey]bool)
	for _, cf := range allCF {
		if cf.ContentDataID.Valid && cf.FieldID.Valid {
			existing[fieldKey{cf.ContentDataID.ID.String(), cf.FieldID.ID.String()}] = true
		}
	}

	// For each content data node, look up all schema-defined fields for its
	// datatype and add empty stubs for any that don't have content values.
	for _, da := range dataSlice {
		if !da.DatatypeID.Valid {
			continue
		}
		dtID := types.NullableDatatypeID{ID: da.DatatypeID.ID, Valid: true}
		schemaFields, err := d.ListFieldsByDatatypeID(dtID)
		if err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		if schemaFields == nil {
			continue
		}
		for _, sf := range *schemaFields {
			key := fieldKey{da.ContentDataID.String(), sf.FieldID.String()}
			if existing[key] {
				continue
			}
			stub := db.ContentFields{
				ContentDataID: types.NullableContentID{ID: da.ContentDataID, Valid: true},
				FieldID:       types.NullableFieldID{ID: sf.FieldID, Valid: true},
				RouteID:       da.RouteID,
			}
			allCF = append(allCF, stub)
			allFD = append(allFD, sf)
		}
	}

	root, err := model.BuildTree(utility.DefaultLogger, dataSlice, dt, allCF, allFD)
	if err != nil {
		utility.DefaultLogger.Error("BuildTree error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Compose referenced subtrees using live data (preview mode sees live refs).
	if root.CoreRoot != nil {
		fetcher := core.NewCachedFetcher(&core.DbTreeFetcher{Driver: d})
		composeErr := core.ComposeTrees(r.Context(), root.CoreRoot, fetcher, core.ComposeOptions{
			MaxDepth:       c.CompositionMaxDepth(),
			MaxConcurrency: 10,
		})
		if composeErr != nil {
			utility.DefaultLogger.Warn("composition error", composeErr)
		}
		root.RebuildFromCore()
	}

	return applyFormatAndTransform(w, r, c, d, root)
}

// applyFormatAndTransform applies the output format and writes the transformed
// response. Shared by both the published and live delivery paths.
func applyFormatAndTransform(w http.ResponseWriter, r *http.Request, c config.Config, d db.DbDriver, root model.Root) error {
	// Allow format override via query parameter.
	format := c.Output_Format
	if queryFormat := r.URL.Query().Get("format"); queryFormat != "" {
		if config.IsValidOutputFormat(queryFormat) {
			format = config.OutputFormat(queryFormat)
		} else {
			http.Error(w, "Invalid format parameter. Valid options: contentful, sanity, strapi, wordpress, clean, raw", http.StatusBadRequest)
			return nil
		}
	}

	// Create transform config and write response.
	transformCfg := transform.NewTransformConfig(
		format,
		c.Client_Site,
		c.Space_ID,
		d,
	)

	if err := transformCfg.TransformAndWrite(w, root); err != nil {
		utility.DefaultLogger.Error("Transform error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

///////////////////////////////
// SNAPSHOT TREE FETCHER
///////////////////////////////

// SnapshotTreeFetcher implements core.TreeFetcher by resolving content
// references via published snapshots instead of live database data.
// This ensures that public delivery only shows published content for
// referenced subtrees.
type SnapshotTreeFetcher struct {
	Driver db.DbDriver
}

// FetchAndBuildTree retrieves the published snapshot for the given content
// data ID, deserializes it, and builds the tree. If no published snapshot
// exists for the reference, it returns nil gracefully (the composition
// layer will produce a _system_log node for the missing reference).
func (f *SnapshotTreeFetcher) FetchAndBuildTree(ctx context.Context, id types.ContentID) (*core.Root, error) {
	version, err := f.Driver.GetPublishedSnapshot(id, "")
	if err != nil {
		return nil, fmt.Errorf("no published snapshot for %s: %w", id, err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot for %s: %w", id, err)
	}

	cd, err := snapshotContentDataToSlice(snapshot.ContentData)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot content data for %s: %w", id, err)
	}

	dt, err := snapshotDatatypesToSlice(snapshot.Datatypes)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot datatypes for %s: %w", id, err)
	}

	cf, err := snapshotContentFieldsToSlice(snapshot.ContentFields)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot content fields for %s: %w", id, err)
	}

	df, err := snapshotFieldsToSlice(snapshot.Fields)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot fields for %s: %w", id, err)
	}

	// Set the root's datatype type to _nested_root (same as DbTreeFetcher).
	for i, c := range cd {
		if c.ContentDataID == id {
			dt[i].Type = string(types.DatatypeTypeNestedRoot)
			break
		}
	}

	root, _, err := core.BuildTree(cd, dt, cf, df)
	return root, err
}

///////////////////////////////
// SNAPSHOT REVERSE MAPPERS
///////////////////////////////

// snapshotContentDataToSlice converts snapshot JSON content data back to
// typed db.ContentData structs for tree building.
func snapshotContentDataToSlice(items []db.ContentDataJSON) ([]db.ContentData, error) {
	result := make([]db.ContentData, len(items))
	for i, item := range items {
		ts, err := parseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] date_created: %w", i, err)
		}
		tm, err := parseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] date_modified: %w", i, err)
		}
		pubAt, err := parseSnapshotTimestamp(item.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] published_at: %w", i, err)
		}
		pubAtField, err := parseSnapshotTimestamp(item.PublishAt)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] publish_at: %w", i, err)
		}

		result[i] = db.ContentData{
			ContentDataID: types.ContentID(item.ContentDataID),
			ParentID:      snapshotNullableContentID(item.ParentID),
			FirstChildID:  snapshotNullableContentID(item.FirstChildID),
			NextSiblingID: snapshotNullableContentID(item.NextSiblingID),
			PrevSiblingID: snapshotNullableContentID(item.PrevSiblingID),
			RouteID:       parseNullableRouteID(item.RouteID),
			DatatypeID:    parseNullableDatatypeID(item.DatatypeID),
			AuthorID:      types.UserID(item.AuthorID),
			Status:        types.ContentStatus(item.Status),
			DateCreated:   ts,
			DateModified:  tm,
			PublishedAt:   pubAt,
			PublishedBy:   parseNullableUserID(item.PublishedBy),
			PublishAt:     pubAtField,
			Revision:      item.Revision,
		}
	}
	return result, nil
}

// snapshotDatatypesToSlice converts snapshot JSON datatypes back to
// typed db.Datatypes structs for tree building.
func snapshotDatatypesToSlice(items []db.DatatypeJSON) ([]db.Datatypes, error) {
	result := make([]db.Datatypes, len(items))
	for i, item := range items {
		ts, err := parseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("datatypes[%d] date_created: %w", i, err)
		}
		tm, err := parseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("datatypes[%d] date_modified: %w", i, err)
		}

		result[i] = db.Datatypes{
			DatatypeID:   types.DatatypeID(item.DatatypeID),
			ParentID:     parseNullableDatatypeID(item.ParentID),
			Name:         item.Name,
			Label:        item.Label,
			Type:         item.Type,
			AuthorID:     types.UserID(item.AuthorID),
			DateCreated:  ts,
			DateModified: tm,
		}
	}
	return result, nil
}

// snapshotContentFieldsToSlice converts snapshot content field JSON back to
// typed db.ContentFields structs for tree building.
func snapshotContentFieldsToSlice(items []SnapshotContentFieldJSON) ([]db.ContentFields, error) {
	result := make([]db.ContentFields, len(items))
	for i, item := range items {
		ts, err := parseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("content_fields[%d] date_created: %w", i, err)
		}
		tm, err := parseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("content_fields[%d] date_modified: %w", i, err)
		}

		result[i] = db.ContentFields{
			ContentFieldID: types.ContentFieldID(item.ContentFieldID),
			RouteID:        parseNullableRouteID(item.RouteID),
			ContentDataID:  snapshotNullableContentID(item.ContentDataID),
			FieldID:        parseNullableFieldID(item.FieldID),
			FieldValue:     item.FieldValue,
			AuthorID:       types.UserID(item.AuthorID),
			DateCreated:    ts,
			DateModified:   tm,
		}
	}
	return result, nil
}

// snapshotFieldsToSlice converts snapshot field definition JSON back to
// typed db.Fields structs for tree building.
func snapshotFieldsToSlice(items []db.FieldsJSON) ([]db.Fields, error) {
	result := make([]db.Fields, len(items))
	for i, item := range items {
		ts, err := parseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("fields[%d] date_created: %w", i, err)
		}
		tm, err := parseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("fields[%d] date_modified: %w", i, err)
		}

		sortOrder, err := strconv.ParseInt(item.SortOrder, 10, 64)
		if err != nil {
			// Default to 0 if sort order is empty or unparseable.
			sortOrder = 0
		}

		result[i] = db.Fields{
			FieldID:      types.FieldID(item.FieldID),
			ParentID:     parseNullableDatatypeID(item.ParentID),
			SortOrder:    sortOrder,
			Name:         item.Name,
			Label:        item.Label,
			Data:         item.Data,
			Validation:   item.Validation,
			UIConfig:     item.UIConfig,
			Type:         types.FieldType(item.Type),
			AuthorID:     parseNullableUserID(item.AuthorID),
			DateCreated:  ts,
			DateModified: tm,
		}
	}
	return result, nil
}

///////////////////////////////
// SNAPSHOT PARSE HELPERS
///////////////////////////////

// parseSnapshotTimestamp parses a timestamp string from a snapshot.
// Empty strings and "null" produce a zero-value (invalid) Timestamp.
func parseSnapshotTimestamp(s string) (types.Timestamp, error) {
	if s == "" || s == "null" {
		return types.Timestamp{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try RFC3339Nano as fallback.
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return types.Timestamp{}, fmt.Errorf("parse timestamp %q: %w", s, err)
		}
	}
	return types.NewTimestamp(t), nil
}

// snapshotNullableContentID creates a NullableContentID from a snapshot string.
// Empty strings and "null" produce an invalid (unset) nullable.
func snapshotNullableContentID(s string) types.NullableContentID {
	if s == "" || s == "null" {
		return types.NullableContentID{}
	}
	return types.NullableContentID{ID: types.ContentID(s), Valid: true}
}

// parseNullableRouteID creates a NullableRouteID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func parseNullableRouteID(s string) types.NullableRouteID {
	if s == "" || s == "null" {
		return types.NullableRouteID{}
	}
	return types.NullableRouteID{ID: types.RouteID(s), Valid: true}
}

// parseNullableDatatypeID creates a NullableDatatypeID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func parseNullableDatatypeID(s string) types.NullableDatatypeID {
	if s == "" || s == "null" {
		return types.NullableDatatypeID{}
	}
	return types.NullableDatatypeID{ID: types.DatatypeID(s), Valid: true}
}

// parseNullableFieldID creates a NullableFieldID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func parseNullableFieldID(s string) types.NullableFieldID {
	if s == "" || s == "null" {
		return types.NullableFieldID{}
	}
	return types.NullableFieldID{ID: types.FieldID(s), Valid: true}
}

// parseNullableUserID creates a NullableUserID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func parseNullableUserID(s string) types.NullableUserID {
	if s == "" || s == "null" {
		return types.NullableUserID{}
	}
	return types.NullableUserID{ID: types.UserID(s), Valid: true}
}
