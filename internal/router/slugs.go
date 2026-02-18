package router

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/utility"
)

// SessionsHandler handles CRUD operations that do not require a specific user ID.
func SlugHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetSlugContent(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetSlugContent(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	route, err := d.GetRouteID(r.URL.Path)
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

	// Allow format override via query parameter
	format := c.Output_Format
	if queryFormat := r.URL.Query().Get("format"); queryFormat != "" {
		if config.IsValidOutputFormat(queryFormat) {
			format = config.OutputFormat(queryFormat)
		} else {
			http.Error(w, "Invalid format parameter. Valid options: contentful, sanity, strapi, wordpress, clean, raw", http.StatusBadRequest)
			return nil
		}
	}

	// Create transform config and write response
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
