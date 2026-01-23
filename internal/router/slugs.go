package router

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
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
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	route, err := d.GetRouteID(r.URL.Path)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	contentData, err := d.ListContentDataByRoute(*route)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	dataSlice := *contentData
	dt := []db.Datatypes{}
	fd := []db.Fields{}
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
	contentFields, err := d.ListContentFieldsByRoute(*route)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	for _, cf := range *contentFields {
		field, err := d.GetField(cf.FieldID)
		if err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		fd = append(fd, *field)
	}
	root := model.BuildTree(*contentData, dt, *contentFields, fd)

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
