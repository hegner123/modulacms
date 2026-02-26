package core

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// DbTreeFetcher is the concrete TreeFetcher that uses db.DbDriver to
// fetch content data descendants and build trees via core.BuildTree.
// The referenced content_data_id is treated as the root — its type is
// set to _nested_root before tree assembly.
type DbTreeFetcher struct {
	Driver db.DbDriver
}

// FetchAndBuildTree fetches the content data descendants for the given
// content data ID, fetches their datatypes and fields, sets the root
// type to _nested_root, and returns the built tree.
func (f *DbTreeFetcher) FetchAndBuildTree(ctx context.Context, id types.ContentID) (*Root, error) {
	// 1. Fetch the referenced content_data row + all its descendants.
	cdPtr, err := f.Driver.GetContentDataDescendants(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch tree for content %s: %w", id, err)
	}
	if cdPtr == nil || len(*cdPtr) == 0 {
		return nil, fmt.Errorf("no content data found for %s", id)
	}
	cd := *cdPtr

	// 2. Fetch datatypes for all content_data rows (parallel slice: dt[i] is for cd[i])
	dt, err := f.fetchDatatypesForContentData(cd)
	if err != nil {
		return nil, err
	}

	// 3. Fetch content fields + field definitions for all content_data rows
	cf, df, err := f.fetchFieldsForContentData(cd)
	if err != nil {
		return nil, err
	}

	// 4. Set the root's datatype type to _nested_root.
	for i, c := range cd {
		if c.ContentDataID == id {
			dt[i].Type = string(types.DatatypeTypeNestedRoot)
			break
		}
	}

	// 5. Build tree using the standard pipeline
	root, _, err := BuildTree(cd, dt, cf, df)
	return root, err
}

// fetchDatatypesForContentData returns a parallel slice where dt[i]
// is the datatype for cd[i]. Caches results for duplicate DatatypeIDs.
func (f *DbTreeFetcher) fetchDatatypesForContentData(cd []db.ContentData) ([]db.Datatypes, error) {
	dt := make([]db.Datatypes, len(cd))
	cache := make(map[types.DatatypeID]db.Datatypes)

	for i, c := range cd {
		if !c.DatatypeID.Valid {
			continue
		}
		dtID := c.DatatypeID.ID
		if cached, ok := cache[dtID]; ok {
			dt[i] = cached
			continue
		}
		datatype, err := f.Driver.GetDatatype(dtID)
		if err != nil {
			return nil, fmt.Errorf("fetch datatype %s for content %s: %w", dtID, c.ContentDataID, err)
		}
		cache[dtID] = *datatype
		dt[i] = *datatype
	}
	return dt, nil
}

// fetchFieldsForContentData returns parallel slices where cf[i] pairs
// with df[i]. For each content_data row, fetches field values and their
// definitions.
func (f *DbTreeFetcher) fetchFieldsForContentData(cd []db.ContentData) ([]db.ContentFields, []db.Fields, error) {
	var allCF []db.ContentFields
	var allDF []db.Fields

	// Cache field definitions to avoid repeated lookups
	fieldCache := make(map[types.FieldID]db.Fields)

	for _, c := range cd {
		nullableID := types.NullableContentID{ID: c.ContentDataID, Valid: true}
		cfPtr, err := f.Driver.ListContentFieldsByContentData(nullableID)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch content fields for %s: %w", c.ContentDataID, err)
		}
		if cfPtr == nil {
			continue
		}
		for _, cf := range *cfPtr {
			if !cf.FieldID.Valid {
				continue
			}
			fID := cf.FieldID.ID
			var fieldDef db.Fields
			if cached, ok := fieldCache[fID]; ok {
				fieldDef = cached
			} else {
				fd, err := f.Driver.GetField(fID)
				if err != nil {
					return nil, nil, fmt.Errorf("fetch field %s: %w", fID, err)
				}
				fieldCache[fID] = *fd
				fieldDef = *fd
			}
			allCF = append(allCF, cf)
			allDF = append(allDF, fieldDef)
		}
	}
	return allCF, allDF, nil
}
