package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// DatatypesHandler handles CRUD operations that do not require a specific datatype ID.
func DatatypesHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListDatatypesPaginated(w, r, svc)
		} else {
			apiListDatatypes(w, r, svc)
		}
	case http.MethodPost:
		apiCreateDatatype(w, r, svc)
	case http.MethodDelete:
		apiDeleteDatatype(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DatatypeHandler handles CRUD operations for specific datatype items.
func DatatypeHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetDatatype(w, r, svc)
	case http.MethodPut:
		apiUpdateDatatype(w, r, svc)
	case http.MethodDelete:
		apiDeleteDatatype(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DatatypeFullHandler handles requests for the composed datatype+fields view.
func DatatypeFullHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetDatatypeFull(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetDatatypeFull handles GET requests for a datatype with all field definitions.
func apiGetDatatypeFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	dtID, err := types.ParseDatatypeID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	view, err := svc.Schema.GetDatatypeFull(r.Context(), dtID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
	return nil
}

// apiGetDatatype handles GET requests for a single datatype
func apiGetDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	dId, err := types.ParseDatatypeID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	datatype, err := svc.Schema.GetDatatype(r.Context(), dId)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datatype)
	return nil
}

// apiListDatatypes handles GET requests for listing datatypes
func apiListDatatypes(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	datatypes, err := svc.Schema.ListDatatypes(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datatypes)
	return nil
}

// apiCreateDatatype handles POST requests to create a new datatype
func apiCreateDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var newDatatype db.CreateDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&newDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	createdDatatype, err := svc.Schema.CreateDatatype(r.Context(), ac, newDatatype)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiUpdateDatatype handles PUT requests to update an existing datatype
func apiUpdateDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var updateDatatype db.UpdateDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&updateDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := svc.Schema.UpdateDatatype(r.Context(), ac, updateDatatype)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteDatatype handles DELETE requests for datatypes.
// When cascade=true, deletes all content using the datatype first.
func apiDeleteDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	dtID, err := types.ParseDatatypeID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	cascade := r.URL.Query().Get("cascade") == "true"

	if cascade {
		d := svc.Driver()
		ctx := r.Context()

		contentList, listErr := d.ListContentDataByDatatypeID(dtID)
		if listErr != nil {
			utility.DefaultLogger.Error("cascade delete: failed to list content by datatype", listErr)
			http.Error(w, listErr.Error(), http.StatusInternalServerError)
			return listErr
		}

		contentDeleted := 0
		if contentList != nil {
			for _, cd := range *contentList {
				if delErr := deleteContentWithSiblingRepair(ctx, ac, d, cd.ContentDataID); delErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("cascade delete: failed to delete content %s", cd.ContentDataID), delErr)
					http.Error(w, fmt.Sprintf("failed to delete content %s: %v", cd.ContentDataID, delErr), http.StatusInternalServerError)
					return delErr
				}
				contentDeleted++
			}
		}

		if delErr := svc.Schema.DeleteDatatype(ctx, ac, dtID); delErr != nil {
			writeServiceError(w, delErr)
			return delErr
		}

		writeJSON(w, DatatypeCascadeDeleteResponse{
			DeletedDatatypeID: dtID,
			ContentDeleted:    contentDeleted,
			Errors:            make([]string, 0),
		})
		return nil
	}

	err = svc.Schema.DeleteDatatype(r.Context(), ac, dtID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}
	res := fmt.Sprintf("Deleted %s", dtID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	return nil
}

// apiListDatatypesPaginated handles GET requests for listing datatypes with pagination.
func apiListDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// DatatypesFullListHandler handles requests for the datatype list with field counts.
func DatatypesFullListHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListDatatypesFull(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func apiListDatatypesFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	views, err := svc.Schema.ListDatatypesFull(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(views)
}

// DatatypeCascadeDeleteResponse is the JSON response for DELETE with cascade=true.
type DatatypeCascadeDeleteResponse struct {
	DeletedDatatypeID types.DatatypeID `json:"deleted_datatype_id"`
	ContentDeleted    int              `json:"content_deleted"`
	Errors            []string         `json:"errors"`
}
