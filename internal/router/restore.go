package router

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/config"
)

// RestoreHandler handles recieving DB restore commands
func RestoreHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		GetRestoreConfirm(w, r, c)
	case http.MethodPost:
		CreateRestoreCMD(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetRestoreConfirm(w http.ResponseWriter, r *http.Request, c config.Config) {
	hash := backup.RestoreHash
	params := r.URL.Query()
	confirm := params.Get("hash")
	if hash == backup.RestoreConfirmHash(confirm) {
		w.WriteHeader(http.StatusOK)
	}
	w.WriteHeader(401)
}

func CreateRestoreCMD(w http.ResponseWriter, r *http.Request, c config.Config) {
	cmd := backup.RestoreCMD{}

	err := json.NewDecoder(r.Body).Decode(&cmd)
	if err != nil {
		http.Error(w, "json decode error", http.StatusInternalServerError)
		return
	}

	u := url.URL{
		Scheme: "https",
		Host:   cmd.Origin,
		Path:   "/restore",
	}

	params := url.Values{}
	params.Add("hash", string(cmd.Hash))

	u.RawQuery = params.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		http.Error(w, "cannot confirm restore", http.StatusInternalServerError)
		return
	}
	if res.Status != "200 OK" {
		http.Error(w, "cannot confirm restore", http.StatusInternalServerError)
		return
	}
	// TODO db function to execute sql file at path
}
