package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminContentDatas handles CRUD operations that do not require a specific user ID.
func AdminContentDatasHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		err := apiListAdminContentData(w, r, c)
		if err != nil {
			return
		}
	case http.MethodPost:
		err := apiCreateAdminContentData(w, r, c)
		if err != nil {
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func AdminContentDataHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodPost:
		err := apiCreateAdminContentData(w, r, c)
		if err != nil {
			return
		}
	case http.MethodPut:
		err := apiUpdateAdminContentData(w, r, c)
		if err != nil {
			return
		}
	case http.MethodDelete:
		err := apiDeleteAdminContentData(w, r, c)
		if err != nil {
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListAdminContentData handles GET requests for listing admin content data
func apiListAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	if r == nil {
		err := fmt.Errorf("request error")
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err

	}

	adminContentDataList, err := d.ListAdminContentData()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(adminContentDataList)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiCreateAdminContentData handles POST requests to create new admin content data
func apiCreateAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var newAdminContentData db.CreateAdminContentDataParams
	err = json.NewDecoder(r.Body).Decode(&newAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	createdAdminContentData := d.CreateAdminContentData(newAdminContentData)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiUpdateAdminContentData handles PUT requests to update existing admin content data
func apiUpdateAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var updateAdminContentData db.UpdateAdminContentDataParams
	err = json.NewDecoder(r.Body).Decode(&updateAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	updateAdminContentData.DateModified = types.TimestampNow()

	oldData, err := d.GetAdminContentData(updateAdminContentData.AdminContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	_, err = json.Marshal(oldData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	updatedAdminContentData, err := d.UpdateAdminContentData(updateAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(updatedAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiDeleteAdminContentData handles DELETE requests for admin content data
func apiDeleteAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	q := r.URL.Query().Get("q")
	acdID := types.AdminContentID(q)
	if err := acdID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	err = d.DeleteAdminContentData(acdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
