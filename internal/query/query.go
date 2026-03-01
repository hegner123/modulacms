package query

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// QueryParams holds the parameters for a content query.
type QueryParams struct {
	DatatypeName string
	Filters      []Filter
	Sort         SortSpec
	Limit        int64
	Offset       int64
	Locale       string
	Status       string // "" = published only (default)
}

// QueryItem is a single content item with its flattened field values.
type QueryItem struct {
	ContentData db.ContentData
	Fields      map[string]string
	FieldTypes  map[string]types.FieldType
}

// QueryResult holds the paginated query output.
type QueryResult struct {
	Items    []QueryItem
	Datatype db.Datatypes
	Total    int64
	Limit    int64
	Offset   int64
}

// Execute runs the content query pipeline:
// 1. Resolve datatype by name
// 2. Fetch all content data for that datatype
// 3. Filter by status
// 4. Batch-fetch content fields
// 5. Fetch field schema definitions
// 6. Build indexes
// 7. Apply filters, sort, paginate
func Execute(ctx context.Context, driver db.DbDriver, params QueryParams) (QueryResult, error) {
	// 1. Resolve datatype by name.
	datatype, err := driver.GetDatatypeByName(params.DatatypeName)
	if err != nil {
		return QueryResult{}, fmt.Errorf("datatype %q not found: %w", params.DatatypeName, err)
	}

	// 2. Fetch all content data for this datatype.
	allContent, err := driver.ListContentDataByDatatypeID(datatype.DatatypeID)
	if err != nil {
		return QueryResult{}, fmt.Errorf("list content data: %w", err)
	}

	// 3. Filter by status (default: published only).
	status := types.ContentStatusPublished
	if params.Status != "" {
		status = types.ContentStatus(params.Status)
	}
	var filtered []db.ContentData
	for _, cd := range *allContent {
		if cd.Status == status {
			filtered = append(filtered, cd)
		}
	}

	if len(filtered) == 0 {
		return QueryResult{
			Items:    nil,
			Datatype: *datatype,
			Total:    0,
			Limit:    clampLimit(params.Limit),
			Offset:   clampOffset(params.Offset),
		}, nil
	}

	// 4. Collect IDs and batch-fetch content fields.
	ids := make([]types.ContentID, len(filtered))
	for i, cd := range filtered {
		ids[i] = cd.ContentDataID
	}
	allFields, err := driver.ListContentFieldsByContentDataIDs(ctx, ids, params.Locale)
	if err != nil {
		return QueryResult{}, fmt.Errorf("batch content fields: %w", err)
	}

	// 5. Fetch field schema definitions for this datatype.
	dtID := types.NullableDatatypeID{ID: datatype.DatatypeID, Valid: true}
	schemaFields, err := driver.ListFieldsByDatatypeID(dtID)
	if err != nil {
		return QueryResult{}, fmt.Errorf("list fields: %w", err)
	}

	// 6. Build indexes.
	// nameIndex: field_id -> field name
	// typeIndex: field name -> field type
	nameIndex := make(map[string]string)
	typeIndex := make(map[string]types.FieldType)
	if schemaFields != nil {
		for _, f := range *schemaFields {
			nameIndex[f.FieldID.String()] = f.Name
			typeIndex[f.Name] = f.Type
		}
	}

	// fieldIndex: content_data_id -> map[field_name]field_value
	fieldIndex := make(map[string]map[string]string)
	if allFields != nil {
		for _, cf := range *allFields {
			if !cf.ContentDataID.Valid || !cf.FieldID.Valid {
				continue
			}
			cdID := cf.ContentDataID.ID.String()
			fieldName := nameIndex[cf.FieldID.ID.String()]
			if fieldName == "" {
				continue
			}
			if fieldIndex[cdID] == nil {
				fieldIndex[cdID] = make(map[string]string)
			}
			fieldIndex[cdID][fieldName] = cf.FieldValue
		}
	}

	// 7. Build QueryItems.
	items := make([]QueryItem, 0, len(filtered))
	for _, cd := range filtered {
		fields := fieldIndex[cd.ContentDataID.String()]
		if fields == nil {
			fields = make(map[string]string)
		}
		items = append(items, QueryItem{
			ContentData: cd,
			Fields:      fields,
			FieldTypes:  typeIndex,
		})
	}

	// 8. Apply filters.
	items = applyFilters(items, params.Filters, typeIndex)

	// 9. Sort.
	ApplySort(items, params.Sort, typeIndex)

	// 10. Count total before pagination.
	total := int64(len(items))

	// 11. Paginate.
	items, limit, offset := paginate(items, params.Limit, params.Offset)

	return QueryResult{
		Items:    items,
		Datatype: *datatype,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}
