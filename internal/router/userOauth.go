package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// UserOauthsHandler handles CRUD operations that do not require a specific user ID.
func UserOauthsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPost:
		ApiCreateUserOauth(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserHandler handles CRUD operations for specific user items.
func UserOauthHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPut:
		ApiUpdateUserOauth(w, r, c)
	case http.MethodDelete:
		ApiDeleteUserOauth(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ApiCreateUserOauth handles POST requests to create a new user OAuth connection
func ApiCreateUserOauth(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var newUserOauth db.CreateUserOauthParams
	err = json.NewDecoder(r.Body).Decode(&newUserOauth)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	createdUserOauth, err := d.CreateUserOauth(newUserOauth)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUserOauth)
	return nil
}

// ApiUpdateUserOauth handles PUT requests to update an existing user OAuth connection
func ApiUpdateUserOauth(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var updateUserOauth db.UpdateUserOauthParams
	err = json.NewDecoder(r.Body).Decode(&updateUserOauth)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	updatedUserOauth, err := d.UpdateUserOauth(updateUserOauth)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUserOauth)
	return nil
}

// ApiDeleteUserOauth handles DELETE requests for user OAuth connections
func ApiDeleteUserOauth(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	q := r.URL.Query().Get("q")
	uoID := types.UserOauthID(q)
	if err := uoID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	err = d.DeleteUserOauth(uoID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
