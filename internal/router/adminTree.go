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

	routeIDStr := route.AdminRouteID.String()

	contentData, err := d.ListAdminContentDataByRoute(routeIDStr)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Filter to entries with valid datatype IDs and fetch corresponding datatypes.
	// contentData and dt must stay in 1:1 correspondence for BuildAdminTree.
	var filteredData []db.AdminContentData
	var dt []db.AdminDatatypes
	for _, da := range *contentData {
		if !da.AdminDatatypeID.Valid {
			continue
		}
		datatype, err := d.GetAdminDatatypeById(da.AdminDatatypeID.ID)
		if err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		filteredData = append(filteredData, da)
		dt = append(dt, *datatype)
	}

	contentFields, err := d.ListAdminContentFieldsByRoute(routeIDStr)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Filter to entries with valid field IDs and fetch corresponding fields.
	// filteredFields and fd must stay in 1:1 correspondence for BuildAdminTree.
	var filteredFields []db.AdminContentFields
	var fd []db.AdminFields
	for _, cf := range *contentFields {
		if !cf.AdminFieldID.Valid {
			continue
		}
		field, err := d.GetAdminField(cf.AdminFieldID.ID)
		if err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		filteredFields = append(filteredFields, cf)
		fd = append(fd, *field)
	}

	root := model.BuildAdminTree(filteredData, dt, filteredFields, fd)

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
