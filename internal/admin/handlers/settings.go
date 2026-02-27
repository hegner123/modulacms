package handlers

import (
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// SettingsHandler renders the settings page with current configuration values.
func SettingsHandler(mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			utility.DefaultLogger.Error("failed to load config", err)
			http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
			return
		}

		layout := NewAdminData(r, "Settings")
		RenderNav(w, r, "Settings", pages.SettingsContent(cfg, layout.CSRFToken), pages.Settings(layout, cfg))
	}
}

// SettingsUpdateHandler processes settings form submissions.
// Parses form values, builds a map[string]any update, and applies via mgr.Update().
func SettingsUpdateHandler(mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		updates := make(map[string]any)

		// General settings
		setIfPresent(r, updates, "environment")
		setIfPresent(r, updates, "port")
		setIfPresent(r, updates, "ssl_port")
		setIfPresent(r, updates, "ssh_host")
		setIfPresent(r, updates, "ssh_port")
		setIfPresent(r, updates, "client_site")
		setIfPresent(r, updates, "admin_site")
		setIfPresent(r, updates, "log_path")
		setIfPresent(r, updates, "output_format")
		setIfPresent(r, updates, "node_id")
		setIfPresent(r, updates, "space_id")

		// Database settings
		setIfPresent(r, updates, "db_driver")
		setIfPresent(r, updates, "db_url")
		setIfPresent(r, updates, "db_name")
		setIfPresent(r, updates, "db_username")
		setIfPresent(r, updates, "db_password")

		// S3/Storage settings
		setIfPresent(r, updates, "bucket_region")
		setIfPresent(r, updates, "bucket_media")
		setIfPresent(r, updates, "bucket_backup")
		setIfPresent(r, updates, "bucket_endpoint")
		setIfPresent(r, updates, "bucket_access_key")
		setIfPresent(r, updates, "bucket_secret_key")
		setIfPresent(r, updates, "bucket_public_url")
		setIfPresent(r, updates, "bucket_default_acl")
		setBoolIfPresent(r, updates, "bucket_force_path_style")
		setInt64IfPresent(r, updates, "max_upload_size")

		// Auth / Cookie settings
		setIfPresent(r, updates, "auth_salt")
		setIfPresent(r, updates, "cookie_name")
		setIfPresent(r, updates, "cookie_duration")
		setBoolIfPresent(r, updates, "cookie_secure")
		setIfPresent(r, updates, "cookie_samesite")

		// OAuth settings
		setIfPresent(r, updates, "oauth_client_id")
		setIfPresent(r, updates, "oauth_client_secret")
		setIfPresent(r, updates, "oauth_provider_name")
		setIfPresent(r, updates, "oauth_redirect_url")
		setIfPresent(r, updates, "oauth_success_redirect")

		// CORS settings
		setIfPresent(r, updates, "cors_origins")
		setBoolIfPresent(r, updates, "cors_credentials")

		// Email settings
		setBoolIfPresent(r, updates, "email_enabled")
		setIfPresent(r, updates, "email_provider")
		setIfPresent(r, updates, "email_from_address")
		setIfPresent(r, updates, "email_from_name")
		setIfPresent(r, updates, "email_host")
		setIntIfPresent(r, updates, "email_port")
		setIfPresent(r, updates, "email_username")
		setIfPresent(r, updates, "email_password")
		setBoolIfPresent(r, updates, "email_tls")
		setIfPresent(r, updates, "email_api_key")
		setIfPresent(r, updates, "email_reply_to")
		setIfPresent(r, updates, "password_reset_url")

		// Plugin settings
		setBoolIfPresent(r, updates, "plugin_enabled")
		setIfPresent(r, updates, "plugin_directory")
		setIntIfPresent(r, updates, "plugin_max_vms")
		setIntIfPresent(r, updates, "plugin_timeout")

		// Observability settings
		setBoolIfPresent(r, updates, "observability_enabled")
		setIfPresent(r, updates, "observability_provider")
		setIfPresent(r, updates, "observability_dsn")
		setIfPresent(r, updates, "observability_environment")
		setIfPresent(r, updates, "observability_server_name")

		if len(updates) == 0 {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "No changes to save", "type": "info"}}`)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
			return
		}

		result, updateErr := mgr.Update(updates)
		if updateErr != nil {
			utility.DefaultLogger.Error("failed to update settings", updateErr)
			msg := "Failed to update settings"
			if !result.Valid && len(result.Errors) > 0 {
				msg = result.Errors[0]
			}
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "`+msg+`", "type": "error"}}`)
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			http.Error(w, msg, http.StatusUnprocessableEntity)
			return
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Settings saved", "type": "success"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
	}
}

// setIfPresent adds a form value to the updates map if the form field is present.
func setIfPresent(r *http.Request, updates map[string]any, key string) {
	if r.Form.Has(key) {
		updates[key] = r.FormValue(key)
	}
}

// setBoolIfPresent adds a boolean form value to the updates map.
// Checkboxes: present = true, absent = false.
func setBoolIfPresent(r *http.Request, updates map[string]any, key string) {
	// Only set if the form explicitly includes this field (via hidden companion or checkbox)
	if r.Form.Has(key) {
		updates[key] = r.FormValue(key) == "true" || r.FormValue(key) == "on"
	}
}

// setIntIfPresent adds an integer form value to the updates map.
func setIntIfPresent(r *http.Request, updates map[string]any, key string) {
	if r.Form.Has(key) {
		val := r.FormValue(key)
		if val == "" {
			updates[key] = 0
			return
		}
		if n, err := strconv.Atoi(val); err == nil {
			updates[key] = n
		}
	}
}

// setInt64IfPresent adds an int64 form value to the updates map.
func setInt64IfPresent(r *http.Request, updates map[string]any, key string) {
	if r.Form.Has(key) {
		val := r.FormValue(key)
		if val == "" {
			updates[key] = int64(0)
			return
		}
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			updates[key] = n
		}
	}
}
