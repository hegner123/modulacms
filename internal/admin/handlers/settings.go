package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// SettingsHandler renders the settings page with current configuration values.
func SettingsHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			utility.DefaultLogger.Error("failed to load config", err)
			http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
			return
		}

		var searchStatus pages.SearchIndexStatus
		if stats, statsErr := svc.Search.Stats(r.Context()); statsErr == nil {
			searchStatus = pages.SearchIndexStatus{
				Available: true,
				Documents: stats.Documents,
				Terms:     stats.Terms,
				MemBytes:  stats.MemEstimate,
			}
		}

		layout := NewAdminData(r, "Settings")
		RenderNav(w, r, "Settings",
			pages.SettingsContent(cfg, layout.CSRFToken, searchStatus),
			pages.Settings(layout, cfg, searchStatus))
	}
}

// SearchRebuildHandler handles POST /admin/settings/search/rebuild.
// Rebuilds the full-text search index and returns a status partial.
func SearchRebuildHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := svc.Search.Rebuild(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to rebuild search index", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to rebuild search index", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Search index rebuilt", "type": "success"}}`)
		Render(w, r, partials.SearchRebuildResult(stats.Documents, stats.Terms))
	}
}

// SettingsUpdateHandler processes settings form submissions.
// Parses form values, builds a map[string]any update, and applies via svc.Manager().Update().
func SettingsUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		updates := make(map[string]any)

		// General settings
		setIfPresent(r, updates, "environment")
		setIfPresent(r, updates, "os")
		setIfPresent(r, updates, "port")
		setIfPresent(r, updates, "ssl_port")
		setIfPresent(r, updates, "cert_dir")
		setIfPresent(r, updates, "ssh_host")
		setIfPresent(r, updates, "ssh_port")
		setIfPresent(r, updates, "client_site")
		setIfPresent(r, updates, "admin_site")
		setIfPresent(r, updates, "log_path")
		setIfPresent(r, updates, "output_format")
		setIfPresent(r, updates, "custom_style_path")
		setIfPresent(r, updates, "node_id")
		setIfPresent(r, updates, "space_id")
		setJSONIfPresent(r, updates, "environment_hosts")
		setJSONIfPresent(r, updates, "options")

		// Database settings
		setIfPresent(r, updates, "db_driver")
		setIfPresent(r, updates, "db_url")
		setIfPresent(r, updates, "db_name")
		setIfPresent(r, updates, "db_username")
		setIfPresent(r, updates, "db_password")

		// Remote settings
		setIfPresent(r, updates, "remote_url")
		setIfPresent(r, updates, "remote_api_key")

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

		// Backup settings
		setIfPresent(r, updates, "backup_option")
		setSliceIfPresent(r, updates, "backup_paths")

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
		setSliceIfPresent(r, updates, "oauth_scopes")
		setOauthEndpointIfPresent(r, updates)

		// CORS settings
		setSliceIfPresent(r, updates, "cors_origins")
		setSliceIfPresent(r, updates, "cors_methods")
		setSliceIfPresent(r, updates, "cors_headers")
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
		setIfPresent(r, updates, "email_api_endpoint")
		setIfPresent(r, updates, "email_reply_to")
		setIfPresent(r, updates, "password_reset_url")
		setIfPresent(r, updates, "email_aws_access_key_id")
		setIfPresent(r, updates, "email_aws_secret_access_key")

		// Content settings
		setIntIfPresent(r, updates, "composition_max_depth")
		setIntIfPresent(r, updates, "publish_schedule_interval")
		setIntIfPresent(r, updates, "version_max_per_content")
		setSliceIfPresent(r, updates, "richtext_toolbar")

		// Plugin settings
		setBoolIfPresent(r, updates, "plugin_enabled")
		setBoolIfPresent(r, updates, "plugin_hot_reload")
		setIfPresent(r, updates, "plugin_directory")
		setIntIfPresent(r, updates, "plugin_max_vms")
		setIntIfPresent(r, updates, "plugin_timeout")
		setIntIfPresent(r, updates, "plugin_max_ops")
		setIntIfPresent(r, updates, "plugin_rate_limit")
		setIntIfPresent(r, updates, "plugin_max_routes")
		setIntIfPresent(r, updates, "plugin_max_failures")
		setInt64IfPresent(r, updates, "plugin_max_request_body")
		setInt64IfPresent(r, updates, "plugin_max_response_body")
		setIntIfPresent(r, updates, "plugin_db_max_open_conns")
		setIntIfPresent(r, updates, "plugin_db_max_idle_conns")
		setIfPresent(r, updates, "plugin_db_conn_max_lifetime")
		setIfPresent(r, updates, "plugin_reset_interval")
		setIfPresent(r, updates, "plugin_sync_interval")
		setSliceIfPresent(r, updates, "plugin_trusted_proxies")

		// Plugin hooks
		setIntIfPresent(r, updates, "plugin_hook_reserve_vms")
		setIntIfPresent(r, updates, "plugin_hook_max_ops")
		setIntIfPresent(r, updates, "plugin_hook_timeout_ms")
		setIntIfPresent(r, updates, "plugin_hook_event_timeout_ms")
		setIntIfPresent(r, updates, "plugin_hook_max_consecutive_aborts")
		setIntIfPresent(r, updates, "plugin_hook_max_concurrent_after")

		// Plugin HTTP requests
		setIntIfPresent(r, updates, "plugin_request_timeout")
		setIntIfPresent(r, updates, "plugin_request_rate_limit")
		setIntIfPresent(r, updates, "plugin_request_global_rate")
		setInt64IfPresent(r, updates, "plugin_request_max_response")
		setInt64IfPresent(r, updates, "plugin_request_max_body")
		setIntIfPresent(r, updates, "plugin_request_cb_failures")
		setIntIfPresent(r, updates, "plugin_request_cb_reset")
		setBoolIfPresent(r, updates, "plugin_request_allow_local")

		// Observability settings
		setBoolIfPresent(r, updates, "observability_enabled")
		setBoolIfPresent(r, updates, "observability_debug")
		setIfPresent(r, updates, "observability_provider")
		setIfPresent(r, updates, "observability_dsn")
		setIfPresent(r, updates, "observability_environment")
		setIfPresent(r, updates, "observability_release")
		setIfPresent(r, updates, "observability_server_name")
		setIfPresent(r, updates, "observability_flush_interval")
		setFloatIfPresent(r, updates, "observability_sample_rate")
		setFloatIfPresent(r, updates, "observability_traces_rate")
		setBoolIfPresent(r, updates, "observability_send_pii")
		setJSONIfPresent(r, updates, "observability_tags")

		// Update settings
		setBoolIfPresent(r, updates, "update_auto_enabled")
		setBoolIfPresent(r, updates, "update_notify_only")
		setIfPresent(r, updates, "update_check_interval")
		setIfPresent(r, updates, "update_channel")

		// Deploy settings
		setIfPresent(r, updates, "deploy_snapshot_dir")
		setJSONIfPresent(r, updates, "deploy_environments")

		// MCP settings
		setBoolIfPresent(r, updates, "mcp_enabled")
		setIfPresent(r, updates, "mcp_api_key")

		// Search settings
		setBoolIfPresent(r, updates, "search_enabled")
		setIfPresent(r, updates, "search_path")

		// i18n settings
		setBoolIfPresent(r, updates, "i18n_enabled")
		setIfPresent(r, updates, "i18n_default_locale")

		// Webhook settings
		setBoolIfPresent(r, updates, "webhook_enabled")
		setBoolIfPresent(r, updates, "webhook_allow_http")
		setIntIfPresent(r, updates, "webhook_timeout")
		setIntIfPresent(r, updates, "webhook_max_retries")
		setIntIfPresent(r, updates, "webhook_workers")
		setIntIfPresent(r, updates, "webhook_delivery_retention_days")

		// Keybindings
		setJSONIfPresent(r, updates, "keybindings")

		// Log every key being submitted
		utility.DefaultLogger.Info("settings update: collected updates", "count", len(updates))
		for k, v := range updates {
			utility.DefaultLogger.Info("settings update: field", "key", k, "type", fmt.Sprintf("%T", v), "value", fmt.Sprintf("%v", v))
		}

		if len(updates) == 0 {
			utility.DefaultLogger.Info("settings update: no changes to save")
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "No changes to save", "type": "info"}}`)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
			return
		}

		result, updateErr := svc.Manager().Update(updates)
		utility.DefaultLogger.Info("settings update: Manager().Update returned",
			"valid", result.Valid,
			"errors", fmt.Sprintf("%v", result.Errors),
			"warnings", fmt.Sprintf("%v", result.Warnings),
			"restart_required", fmt.Sprintf("%v", result.RestartRequired),
			"updateErr", fmt.Sprintf("%v", updateErr),
		)
		if updateErr != nil {
			utility.DefaultLogger.Error("failed to update settings", updateErr)
			msg := "Failed to update settings"
			if !result.Valid && len(result.Errors) > 0 {
				msg = result.Errors[0]
			}
			utility.DefaultLogger.Info("settings update: returning 422", "msg", msg)
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
// Checkboxes use a hidden input (value="false") + checkbox (value="true").
// When checked, both values are sent — the last value is the checkbox.
// When unchecked, only the hidden "false" is sent.
func setBoolIfPresent(r *http.Request, updates map[string]any, key string) {
	vals := r.Form[key]
	if len(vals) == 0 {
		return
	}
	last := vals[len(vals)-1]
	updates[key] = last == "true" || last == "on"
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

// setFloatIfPresent adds a float64 form value to the updates map.
func setFloatIfPresent(r *http.Request, updates map[string]any, key string) {
	if r.Form.Has(key) {
		val := r.FormValue(key)
		if val == "" {
			updates[key] = float64(0)
			return
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			updates[key] = f
		}
	}
}

// setJSONIfPresent parses a JSON textarea value and adds the parsed value
// to the updates map. Skips null/empty values to preserve existing config.
// Invalid JSON is silently ignored.
func setJSONIfPresent(r *http.Request, updates map[string]any, key string) {
	if !r.Form.Has(key) {
		return
	}
	raw := strings.TrimSpace(r.FormValue(key))
	if raw == "" || raw == "null" || raw == "{}" || raw == "[]" {
		return
	}
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		updates[key] = parsed
	}
}

// setOauthEndpointIfPresent builds the oauth_endpoint map from individual
// form fields (oauth_auth_url, oauth_token_url, oauth_userinfo_url).
func setOauthEndpointIfPresent(r *http.Request, updates map[string]any) {
	hasAny := r.Form.Has("oauth_auth_url") || r.Form.Has("oauth_token_url") || r.Form.Has("oauth_userinfo_url")
	if !hasAny {
		return
	}
	endpoints := map[string]string{}
	if r.Form.Has("oauth_auth_url") {
		v := strings.TrimSpace(r.FormValue("oauth_auth_url"))
		if v != "" {
			endpoints[string(config.OauthAuthURL)] = v
		}
	}
	if r.Form.Has("oauth_token_url") {
		v := strings.TrimSpace(r.FormValue("oauth_token_url"))
		if v != "" {
			endpoints[string(config.OauthTokenURL)] = v
		}
	}
	if r.Form.Has("oauth_userinfo_url") {
		v := strings.TrimSpace(r.FormValue("oauth_userinfo_url"))
		if v != "" {
			endpoints[string(config.OauthUserInfoURL)] = v
		}
	}
	updates["oauth_endpoint"] = endpoints
}

// setSliceIfPresent parses a comma-separated textarea value into a []string
// and adds it to the updates map. Empty input results in an empty slice.
func setSliceIfPresent(r *http.Request, updates map[string]any, key string) {
	if r.Form.Has(key) {
		raw := r.FormValue(key)
		if raw == "" {
			updates[key] = []string{}
			return
		}
		parts := strings.Split(raw, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		updates[key] = result
	}
}
