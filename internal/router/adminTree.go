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
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/utility"
)

// TreeSaveRequest is the JSON body for POST /api/v1/content/tree.
type TreeSaveRequest struct {
	AdminContentID types.AdminContentID   `json:"content_id"`
	Creates        []TreeNodeCreate       `json:"creates"`
	Updates        []TreeNodeUpdate       `json:"updates"`
	Deletes        []types.AdminContentID `json:"deletes"`
}

// TreeNodeCreate describes a new content_data node to insert.
// ClientID is the client-generated UUID; the server generates a ULID and returns
// the mapping in TreeSaveResponse.IDMap.
type TreeNodeCreate struct {
	ClientID      string  `json:"client_id"`
	DatatypeID    string  `json:"datatype_id"`
	ParentID      *string `json:"parent_id"`
	FirstChildID  *string `json:"first_child_id"`
	NextSiblingID *string `json:"next_sibling_id"`
	PrevSiblingID *string `json:"prev_sibling_id"`
}

// TreeNodeUpdate describes pointer-field changes for a single content_data node.
type TreeNodeUpdate struct {
	AdminContentDataID types.AdminContentID `json:"content_data_id"`
	ParentID           *string              `json:"parent_id"`
	FirstChildID       *string              `json:"first_child_id"`
	NextSiblingID      *string              `json:"next_sibling_id"`
	PrevSiblingID      *string              `json:"prev_sibling_id"`
}

// TreeSaveResponse summarises the results of a tree save operation.
type TreeSaveResponse struct {
	Created int               `json:"created"`
	Updated int               `json:"updated"`
	Deleted int               `json:"deleted"`
	IDMap   map[string]string `json:"id_map,omitempty"`
	Errors  []string          `json:"errors,omitempty"`
}

// parseNullableAdminContentID converts a *string from JSON into a NullableAdminContentID.
// nil pointer means SQL NULL; non-nil string is parsed as a AdminContentID.
func parseNullableAdminContentID(s *string) (types.NullableAdminContentID, error) {
	if s == nil {
		return types.NullableAdminContentID{Valid: false}, nil
	}
	id := types.AdminContentID(*s)
	if err := id.Validate(); err != nil {
		return types.NullableAdminContentID{}, fmt.Errorf("invalid content_id %q: %w", *s, err)
	}
	return types.NullableAdminContentID{ID: id, Valid: true}, nil
}

// AdminTreeHandler handles GET requests for admin content trees by slug.
func AdminTreeHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminTreeContent(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetAdminTreeContent(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	d := svc.Driver()

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

	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return err
	}

	// Allow format override via query parameter
	format := cfg.Output_Format
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
		cfg.Client_Site,
		cfg.Space_ID,
		d,
	)

	if err := transformCfg.TransformAndWrite(w, root); err != nil {
		utility.DefaultLogger.Error("Transform error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
