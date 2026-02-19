package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// TablesHandler handles CRUD operations that do not require a specific user ID.
func TablesHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListTables(w, c)
	case http.MethodPost:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TableHandler handles CRUD operations for specific user items.
func TableHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetTable(w, r, c)
	case http.MethodPut:
		apiUpdateTables(w, r, c)
	case http.MethodDelete:
		apiDeleteTable(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetTable handles GET requests for a single table
func apiGetTable(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	tId := r.URL.Query().Get("q")
	if tId == "" {
		err := fmt.Errorf("missing table ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	table, err := d.GetTable(tId)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(table)
	return nil
}

// apiListTables handles GET requests for listing tables
func apiListTables(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	tables, err := d.ListTables()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tables)
	return nil
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
func apiUpdateTables(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateTable db.UpdateTableParams
	err := json.NewDecoder(r.Body).Decode(&updateTable)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateTable(r.Context(), ac, updateTable)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetTable(updateTable.ID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated table", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteTable handles DELETE requests for tables
func apiDeleteTable(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	tId := r.URL.Query().Get("q")
	if tId == "" {
		err := fmt.Errorf("missing table ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteTable(r.Context(), ac, tId)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
