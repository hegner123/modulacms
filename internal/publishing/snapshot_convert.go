package publishing

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// SnapshotContentDataToSlice converts snapshot JSON content data back to
// typed db.ContentData structs for tree building.
func SnapshotContentDataToSlice(items []db.ContentDataJSON) ([]db.ContentData, error) {
	result := make([]db.ContentData, len(items))
	for i, item := range items {
		ts, err := ParseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] date_created: %w", i, err)
		}
		tm, err := ParseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] date_modified: %w", i, err)
		}
		pubAt, err := ParseSnapshotTimestamp(item.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] published_at: %w", i, err)
		}
		pubAtField, err := ParseSnapshotTimestamp(item.PublishAt)
		if err != nil {
			return nil, fmt.Errorf("content_data[%d] publish_at: %w", i, err)
		}

		result[i] = db.ContentData{
			ContentDataID: types.ContentID(item.ContentDataID),
			ParentID:      SnapshotNullableContentID(item.ParentID),
			FirstChildID:  SnapshotNullableContentID(item.FirstChildID),
			NextSiblingID: SnapshotNullableContentID(item.NextSiblingID),
			PrevSiblingID: SnapshotNullableContentID(item.PrevSiblingID),
			RouteID:       ParseNullableRouteID(item.RouteID),
			DatatypeID:    ParseNullableDatatypeID(item.DatatypeID),
			AuthorID:      types.UserID(item.AuthorID),
			Status:        types.ContentStatus(item.Status),
			DateCreated:   ts,
			DateModified:  tm,
			PublishedAt:   pubAt,
			PublishedBy:   ParseNullableUserID(item.PublishedBy),
			PublishAt:     pubAtField,
			Revision:      item.Revision,
		}
	}
	return result, nil
}

// SnapshotDatatypesToSlice converts snapshot JSON datatypes back to
// typed db.Datatypes structs for tree building.
func SnapshotDatatypesToSlice(items []db.DatatypeJSON) ([]db.Datatypes, error) {
	result := make([]db.Datatypes, len(items))
	for i, item := range items {
		ts, err := ParseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("datatypes[%d] date_created: %w", i, err)
		}
		tm, err := ParseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("datatypes[%d] date_modified: %w", i, err)
		}

		result[i] = db.Datatypes{
			DatatypeID:   types.DatatypeID(item.DatatypeID),
			ParentID:     ParseNullableDatatypeID(item.ParentID),
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

// SnapshotContentFieldsToSlice converts snapshot content field JSON back to
// typed db.ContentFields structs for tree building.
func SnapshotContentFieldsToSlice(items []SnapshotContentFieldJSON) ([]db.ContentFields, error) {
	result := make([]db.ContentFields, len(items))
	for i, item := range items {
		ts, err := ParseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("content_fields[%d] date_created: %w", i, err)
		}
		tm, err := ParseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("content_fields[%d] date_modified: %w", i, err)
		}

		result[i] = db.ContentFields{
			ContentFieldID: types.ContentFieldID(item.ContentFieldID),
			RouteID:        ParseNullableRouteID(item.RouteID),
			ContentDataID:  SnapshotNullableContentID(item.ContentDataID),
			FieldID:        ParseNullableFieldID(item.FieldID),
			FieldValue:     item.FieldValue,
			Locale:         item.Locale,
			AuthorID:       types.UserID(item.AuthorID),
			DateCreated:    ts,
			DateModified:   tm,
		}
	}
	return result, nil
}

// SnapshotFieldsToSlice converts snapshot field definition JSON back to
// typed db.Fields structs for tree building.
func SnapshotFieldsToSlice(items []db.FieldsJSON) ([]db.Fields, error) {
	result := make([]db.Fields, len(items))
	for i, item := range items {
		ts, err := ParseSnapshotTimestamp(item.DateCreated)
		if err != nil {
			return nil, fmt.Errorf("fields[%d] date_created: %w", i, err)
		}
		tm, err := ParseSnapshotTimestamp(item.DateModified)
		if err != nil {
			return nil, fmt.Errorf("fields[%d] date_modified: %w", i, err)
		}

		sortOrder, err := strconv.ParseInt(item.SortOrder, 10, 64)
		if err != nil {
			// Default to 0 if sort order is empty or unparseable.
			sortOrder = 0
		}

		translatable := item.Translatable == "1" || item.Translatable == "true"

		var roles types.NullableString
		if item.Roles != "" {
			roles = types.NewNullableString(item.Roles)
		}

		result[i] = db.Fields{
			FieldID:      types.FieldID(item.FieldID),
			ParentID:     ParseNullableDatatypeID(item.ParentID),
			SortOrder:    sortOrder,
			Name:         item.Name,
			Label:        item.Label,
			Data:         item.Data,
			Validation:   item.Validation,
			UIConfig:     item.UIConfig,
			Type:         types.FieldType(item.Type),
			Translatable: translatable,
			Roles:        roles,
			AuthorID:     ParseNullableUserID(item.AuthorID),
			DateCreated:  ts,
			DateModified: tm,
		}
	}
	return result, nil
}

// ParseSnapshotTimestamp parses a timestamp string from a snapshot.
// Empty strings and "null" produce a zero-value (invalid) Timestamp.
func ParseSnapshotTimestamp(s string) (types.Timestamp, error) {
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

// SnapshotNullableContentID creates a NullableContentID from a snapshot string.
// Empty strings and "null" produce an invalid (unset) nullable.
func SnapshotNullableContentID(s string) types.NullableContentID {
	if s == "" || s == "null" {
		return types.NullableContentID{}
	}
	return types.NullableContentID{ID: types.ContentID(s), Valid: true}
}

// ParseNullableRouteID creates a NullableRouteID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func ParseNullableRouteID(s string) types.NullableRouteID {
	if s == "" || s == "null" {
		return types.NullableRouteID{}
	}
	return types.NullableRouteID{ID: types.RouteID(s), Valid: true}
}

// ParseNullableDatatypeID creates a NullableDatatypeID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func ParseNullableDatatypeID(s string) types.NullableDatatypeID {
	if s == "" || s == "null" {
		return types.NullableDatatypeID{}
	}
	return types.NullableDatatypeID{ID: types.DatatypeID(s), Valid: true}
}

// ParseNullableFieldID creates a NullableFieldID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func ParseNullableFieldID(s string) types.NullableFieldID {
	if s == "" || s == "null" {
		return types.NullableFieldID{}
	}
	return types.NullableFieldID{ID: types.FieldID(s), Valid: true}
}

// ParseNullableUserID creates a NullableUserID from a string.
// Empty strings and "null" produce an invalid (unset) nullable.
func ParseNullableUserID(s string) types.NullableUserID {
	if s == "" || s == "null" {
		return types.NullableUserID{}
	}
	return types.NullableUserID{ID: types.UserID(s), Valid: true}
}
