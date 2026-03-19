package handlers

import (
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/deploy"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// DeployPageHandler renders the deploy page showing configured environments.
func DeployPageHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			utility.DefaultLogger.Error("failed to load config for deploy page", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		envs := cfg.Deploy_Environments
		if envs == nil {
			envs = []config.DeployEnvironmentConfig{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		hasWrite := HasPermission(r, "deploy:create")

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Deploy"}`)
			Render(w, r, pages.DeployContent(envs, csrfToken, hasWrite))
			return
		}

		layout := NewAdminData(r, "Deploy")
		Render(w, r, pages.Deploy(layout, envs, csrfToken, hasWrite))
	}
}

// DeployHealthHandler tests connectivity to a configured environment.
// Returns a partial that swaps into the environment card's status area.
func DeployHealthHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			htmxError(w, r, "Environment name is required", http.StatusBadRequest)
			return
		}

		cfg, err := svc.Config()
		if err != nil {
			htmxError(w, r, "Failed to load config", http.StatusInternalServerError)
			return
		}

		health, healthErr := deploy.TestEnvConnection(r.Context(), *cfg, name)
		if healthErr != nil {
			Render(w, r, partials.DeployHealthStatus(name, nil, healthErr.Error()))
			return
		}
		Render(w, r, partials.DeployHealthStatus(name, health, ""))
	}
}

// DeployPushHandler pushes local content to a remote environment.
// Supports ?dry_run=true for validation without writing.
func DeployPushHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			htmxError(w, r, "Environment name is required", http.StatusBadRequest)
			return
		}

		cfg, err := svc.Config()
		if err != nil {
			htmxError(w, r, "Failed to load config", http.StatusInternalServerError)
			return
		}

		dryRun := r.URL.Query().Get("dry_run") == "true"
		opts := deploy.ExportOptions{} // default table set

		result, pushErr := deploy.Push(r.Context(), *cfg, svc.Driver(), name, opts, dryRun)
		if pushErr != nil {
			w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "error"}}`, "Push failed: "+pushErr.Error()))
			Render(w, r, partials.DeploySyncResult(name, "push", nil, pushErr.Error()))
			return
		}

		if result.Success {
			msg := "Push completed"
			if dryRun {
				msg = "Dry run completed"
			}
			w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "success"}}`, msg))
		}
		Render(w, r, partials.DeploySyncResult(name, "push", result, ""))
	}
}

// DeployPullHandler pulls content from a remote environment into local.
// Supports ?dry_run=true for validation without writing.
func DeployPullHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			htmxError(w, r, "Environment name is required", http.StatusBadRequest)
			return
		}

		cfg, err := svc.Config()
		if err != nil {
			htmxError(w, r, "Failed to load config", http.StatusInternalServerError)
			return
		}

		dryRun := r.URL.Query().Get("dry_run") == "true"
		opts := deploy.ExportOptions{} // default table set

		result, pullErr := deploy.Pull(r.Context(), *cfg, svc.Driver(), name, opts, false, dryRun)
		if pullErr != nil {
			w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "error"}}`, "Pull failed: "+pullErr.Error()))
			Render(w, r, partials.DeploySyncResult(name, "pull", nil, pullErr.Error()))
			return
		}

		if result.Success {
			msg := "Pull completed"
			if dryRun {
				msg = "Dry run completed"
			}
			w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "success"}}`, msg))
		}
		Render(w, r, partials.DeploySyncResult(name, "pull", result, ""))
	}
}
