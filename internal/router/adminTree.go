package router

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminTreeHandler handles GET requests for admin content trees by slug.
func AdminTreeHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminTreeContent(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetAdminTreeContent(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/tree/")
	if slug == "" {
		http.Error(w, "slug is required", http.StatusBadRequest)
		return nil
	}

	route, err := d.GetAdminRoute(types.Slug(slug))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "admin route not found: "+slug, http.StatusNotFound)
			return err
		}
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	routeID := types.NullableAdminRouteID{ID: route.AdminRouteID, Valid: true}

	// Fetch content data + datatypes in one JOIN query (replaces N+1 pattern)
	joinedData, err := d.ListAdminContentDataWithDatatypeByRoute(routeID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Split JOIN rows into parallel slices for BuildAdminTree
	var filteredData []db.AdminContentData
	var dt []db.AdminDatatypes
	for _, row := range *joinedData {
		filteredData = append(filteredData, db.AdminContentData{
			AdminContentDataID: row.AdminContentDataID,
			ParentID:           row.ParentID,
			FirstChildID:       row.FirstChildID,
			NextSiblingID:      row.NextSiblingID,
			PrevSiblingID:      row.PrevSiblingID,
			AdminRouteID:       row.AdminRouteID,
			AdminDatatypeID:    row.AdminDatatypeID,
			AuthorID:           row.AuthorID,
			Status:             row.Status,
			DateCreated:        row.DateCreated,
			DateModified:       row.DateModified,
		})
		dt = append(dt, db.AdminDatatypes{
			AdminDatatypeID: row.DtAdminDatatypeID,
			ParentID:        row.DtParentID,
			Label:           row.DtLabel,
			Type:            row.DtType,
			AuthorID:        row.DtAuthorID,
			DateCreated:     row.DtDateCreated,
			DateModified:    row.DtDateModified,
		})
	}

	// Fetch content fields + field definitions in one JOIN query (replaces N+1 pattern)
	joinedFields, err := d.ListAdminContentFieldsWithFieldByRoute(routeID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Split JOIN rows into parallel slices for BuildAdminTree
	var filteredFields []db.AdminContentFields
	var fd []db.AdminFields
	for _, row := range *joinedFields {
		filteredFields = append(filteredFields, db.AdminContentFields{
			AdminContentFieldID: row.AdminContentFieldID,
			AdminRouteID:        row.AdminRouteID,
			AdminContentDataID:  row.AdminContentDataID,
			AdminFieldID:        row.AdminFieldID,
			AdminFieldValue:     row.AdminFieldValue,
			AuthorID:            row.AuthorID,
			DateCreated:         row.DateCreated,
			DateModified:        row.DateModified,
		})
		fd = append(fd, db.AdminFields{
			AdminFieldID: row.FAdminFieldID,
			ParentID:     row.FParentID,
			Label:        row.FLabel,
			Data:         row.FData,
			Validation:   row.FValidation,
			UIConfig:     row.FUIConfig,
			Type:         row.FType,
			AuthorID:     row.FAuthorID,
			DateCreated:  row.FDateCreated,
			DateModified: row.FDateModified,
		})
	}

	root, err := model.BuildAdminTree(utility.DefaultLogger, filteredData, dt, filteredFields, fd)
	if err != nil {
		utility.DefaultLogger.Error("BuildAdminTree error", err)
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
