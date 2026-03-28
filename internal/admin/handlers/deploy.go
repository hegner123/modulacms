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

// findDeployEnv looks up a deploy environment by name from the config.
func findDeployEnv(envs []config.DeployEnvironmentConfig, name string) (config.DeployEnvironmentConfig, bool) {
	for _, env := range envs {
		if env.Name == name {
			return env, true
		}
	}
	return config.DeployEnvironmentConfig{}, false
}

// resolveDeployTables reads table selection from the form (custom dialog) or
// falls back to the environment's configured default tables.
func resolveDeployTables(r *http.Request, env config.DeployEnvironmentConfig) deploy.ExportOptions {
	if err := r.ParseForm(); err == nil {
		if formTables := r.Form["tables"]; len(formTables) > 0 {
			return deploy.ExportOptions{Tables: deploy.ResolveTables(formTables)}
		}
	}
	return deploy.ExportOptions{Tables: deploy.TablesForEnv(env)}
}

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
// If form contains "tables" values (from custom dialog), those are used.
// Otherwise falls back to the environment's configured tables or content-only default.
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

		env, found := findDeployEnv(cfg.Deploy_Environments, name)
		if !found {
			htmxError(w, r, "Unknown deploy environment", http.StatusBadRequest)
			return
		}

		dryRun := r.URL.Query().Get("dry_run") == "true"
		opts := resolveDeployTables(r, env)

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
// If form contains "tables" values (from custom dialog), those are used.
// Otherwise falls back to the environment's configured tables or content-only default.
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

		env, found := findDeployEnv(cfg.Deploy_Environments, name)
		if !found {
			htmxError(w, r, "Unknown deploy environment", http.StatusBadRequest)
			return
		}

		dryRun := r.URL.Query().Get("dry_run") == "true"
		opts := resolveDeployTables(r, env)

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
