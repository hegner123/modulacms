package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// TablesHandler handles CRUD operations that do not require a specific user ID.
func TablesHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListTables(w, r, svc)
	case http.MethodPost:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TableHandler handles CRUD operations for specific user items.
func TableHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetTable(w, r, svc)
	case http.MethodPut:
		apiUpdateTables(w, r, svc)
	case http.MethodDelete:
		apiDeleteTable(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetTable handles GET requests for a single table
func apiGetTable(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	tID := r.URL.Query().Get("q")
	if tID == "" {
		http.Error(w, "missing table ID", http.StatusBadRequest)
		return
	}

	table, err := svc.Tables.GetTable(r.Context(), tID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(table)
}

// apiListTables handles GET requests for listing tables
func apiListTables(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	tables, err := svc.Tables.ListTables(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tables)
}

/*
// apiCreateTable handles POST requests to create a new table
func apiCreateTable(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var label string
	err = r.ParseForm()
	if err != nil {
		utility.DefaultLogger.Error("", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	createdTable, err := d.CreateTable(newTable)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTable)
	return nil
}
*/

// apiUpdateTables handles PUT requests to update an existing table
func apiUpdateTables(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.UpdateTableParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.Tables.UpdateTable(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// apiDeleteTable handles DELETE requests for tables
func apiDeleteTable(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	tID := r.URL.Query().Get("q")
	if tID == "" {
		http.Error(w, "missing table ID", http.StatusBadRequest)
		return
	}

	c, cfgErr := svc.Config()
	if cfgErr != nil {
		service.HandleServiceError(w, r, cfgErr)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.Tables.DeleteTable(r.Context(), ac, tID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
