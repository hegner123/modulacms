package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// DatatypesHandler handles CRUD operations that do not require a specific datatype ID.
func DatatypesHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListDatatypesPaginated(w, r, c)
		} else {
			apiListDatatypes(w, c)
		}
	case http.MethodPost:
		apiCreateDatatype(w, r, c)
	case http.MethodDelete:
		apiDeleteDatatype(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DatatypeHandler handles CRUD operations for specific datatype items.
func DatatypeHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetDatatype(w, r, c)
	case http.MethodPut:
		apiUpdateDatatype(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DatatypeFullHandler handles requests for the composed datatype+fields view.
func DatatypeFullHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetDatatypeFull(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetDatatypeFull handles GET requests for a datatype with all field definitions.
func apiGetDatatypeFull(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	dtID, err := types.ParseDatatypeID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	view, err := db.AssembleDatatypeFullView(d, dtID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
	return nil
}

// apiGetDatatype handles GET requests for a single datatype
func apiGetDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	dId, err := types.ParseDatatypeID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	datatype, err := d.GetDatatype(dId)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datatype)
	return nil
}

// apiListDatatypes handles GET requests for listing datatypes
func apiListDatatypes(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	datatypes, err := d.ListDatatypes()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datatypes)
	return nil
}

// apiCreateDatatype handles POST requests to create a new datatype
func apiCreateDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newDatatype db.CreateDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&newDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if newDatatype.DatatypeID.IsZero() {
		newDatatype.DatatypeID = types.NewDatatypeID()
	}
	now := types.NewTimestamp(time.Now().UTC())
	if !newDatatype.DateCreated.Valid {
		newDatatype.DateCreated = now
	}
	if !newDatatype.DateModified.Valid {
		newDatatype.DateModified = now
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdDatatype, err := d.CreateDatatype(r.Context(), ac, newDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func apiUpdateDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateDatatype db.UpdateDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&updateDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	updatedDatatype, err := d.UpdateDatatype(r.Context(), ac, updateDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedDatatype)
	return nil
}

// apiDeleteDatatype handles DELETE requests for datatypes
func apiDeleteDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("id")
	dtID, err := types.ParseDatatypeID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err = d.DeleteDatatype(r.Context(), ac, dtID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	res := fmt.Sprintf("Deleted %s", dtID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	return nil
}

// apiListDatatypesPaginated handles GET requests for listing datatypes with pagination.
func apiListDatatypesPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListDatatypesPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountDatatypes()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.Datatypes]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}
