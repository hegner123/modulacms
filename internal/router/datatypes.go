package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// DatatypesHandler handles CRUD operations that do not require a specific datatype ID.
func DatatypesHandler(w http.ResponseWriter, r *http.Request, c config.Config, svc *service.Registry) {
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
func DatatypeHandler(w http.ResponseWriter, r *http.Request, c config.Config, svc *service.Registry) {
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
func DatatypeFullHandler(w http.ResponseWriter, r *http.Request, c config.Config, svc *service.Registry) {
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

// apiDeleteDatatype handles DELETE requests for datatypes
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
