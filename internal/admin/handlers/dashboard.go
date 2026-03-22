package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// DashboardHandler renders the admin dashboard with system status overview.
func DashboardHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			utility.DefaultLogger.Error("dashboard: failed to load config", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := pages.DashboardData{
			Environment:  cfg.Environment,
			DbDriver:     string(cfg.Db_Driver),
			OutputFormat: string(cfg.Output_Format),
			Version:      utility.Version,
		}

		// Server ports
		data.HTTPPort = cfg.Port
		data.SSLPort = cfg.SSL_Port
		data.SSHPort = cfg.SSH_Port

		// Feature toggles
		data.SSLEnabled = cfg.SSL_Port != "" && cfg.Cert_Dir != ""
		data.SSHEnabled = cfg.SSH_Port != ""
		data.S3Enabled = cfg.Bucket_Endpoint != "" && cfg.Bucket_Media != ""
		data.OAuthEnabled = cfg.Oauth_Client_Id != "" && cfg.Oauth_Provider_Name != ""
		data.OAuthProvider = cfg.Oauth_Provider_Name
		data.EmailEnabled = cfg.Email_Enabled
		data.EmailProvider = string(cfg.Email_Provider)
		data.PluginsEnabled = cfg.Plugin_Enabled
		data.ObservabilityEnabled = cfg.Observability_Enabled
		data.ObservabilityProvider = cfg.Observability_Provider
		data.SearchEnabled = cfg.Search_Enabled
		data.WebhooksEnabled = cfg.Webhook_Enabled
		data.I18nEnabled = cfg.I18n_Enabled
		data.I18nLocale = cfg.I18n_Default_Locale
		data.MCPEnabled = cfg.MCP_Enabled
		data.AutoUpdateEnabled = cfg.Update_Auto_Enabled
		data.UpdateChannel = cfg.Update_Channel
		data.CORSEnabled = len(cfg.Cors_Origins) > 0
		data.DeployEnabled = len(cfg.Deploy_Environments) > 0
		data.DeployTargets = len(cfg.Deploy_Environments)

		// Plugin details (only if enabled)
		if cfg.Plugin_Enabled && svc.Plugins != nil {
			plugins, pluginErr := svc.Plugins.List(r.Context())
			if pluginErr == nil {
				data.Plugins = plugins
				for _, p := range plugins {
					switch p.State {
					case "running":
						data.PluginsRunning++
					case "failed":
						data.PluginsFailed++
					case "stopped", "disabled":
						data.PluginsStopped++
					}
				}
			}
		}

		layout := NewAdminData(r, "Dashboard")
		RenderNav(w, r, "Dashboard", pages.DashboardContent(data), pages.Dashboard(layout, data))
	}
}
