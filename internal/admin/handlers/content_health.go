package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/layouts"
	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// ContentHealthHandler renders the content health page.
func ContentHealthHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := CSRFTokenFromContext(r.Context())
		content := pages.ContentHealthContent(csrfToken)
		fullPage := pages.ContentHealth(NewAdminData(r, "Content Health"), csrfToken)
		RenderNav(w, r, "Content Health", content, fullPage)
	}
}

// ContentHealthCheckHandler runs a dry-run heal and renders the report.
func ContentHealthCheckHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		scope := r.FormValue("scope")
		dryRun := r.FormValue("action") != "heal"

		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}
		ac := middleware.AuditContextFromRequest(r, *cfg)

		switch scope {
		case "public":
			report, err := svc.Content.Heal(r.Context(), ac, dryRun)
			if err != nil {
				service.HandleServiceError(w, r, err)
				return
			}
			actionLabel := "Check"
			if !dryRun {
				actionLabel = "Heal"
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Public content healed", "type": "success"}}`)
			}
			Render(w, r, partials.ContentHealthReport("Public Content", actionLabel, partials.HealthReportData{
				DryRun:              report.DryRun,
				ContentDataScanned:  report.ContentDataScanned,
				ContentFieldScanned: report.ContentFieldScanned,
				ContentDataRepairs:  len(report.ContentDataRepairs),
				ContentFieldRepairs: len(report.ContentFieldRepairs),
				MissingFields:       len(report.MissingFields),
				DuplicateFields:     len(report.DuplicateFields),
				OrphanedFields:      len(report.OrphanedFields),
				DanglingPointers:    len(report.DanglingPointers),
				OrphanedRouteRefs:   len(report.OrphanedRouteRefs),
				UnroutedRoots:       len(report.UnroutedRoots),
				RootlessContent:     len(report.RootlessContent),
			}))

		case "admin":
			report, err := svc.AdminContent.Heal(r.Context(), ac, dryRun)
			if err != nil {
				service.HandleServiceError(w, r, err)
				return
			}
			actionLabel := "Check"
			if !dryRun {
				actionLabel = "Heal"
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin content healed", "type": "success"}}`)
			}
			Render(w, r, partials.ContentHealthReport("Admin Content", actionLabel, partials.HealthReportData{
				DryRun:              report.DryRun,
				ContentDataScanned:  report.ContentDataScanned,
				ContentFieldScanned: report.ContentFieldScanned,
				ContentDataRepairs:  len(report.ContentDataRepairs),
				ContentFieldRepairs: len(report.ContentFieldRepairs),
				MissingFields:       len(report.MissingFields),
				DuplicateFields:     len(report.DuplicateFields),
				OrphanedFields:      len(report.OrphanedFields),
				DanglingPointers:    len(report.DanglingPointers),
				OrphanedRouteRefs:   len(report.OrphanedRouteRefs),
				UnroutedRoots:       len(report.UnroutedRoots),
				RootlessContent:     len(report.RootlessContent),
			}))

		case "all":
			pubReport, pubErr := svc.Content.Heal(r.Context(), ac, dryRun)
			if pubErr != nil {
				service.HandleServiceError(w, r, pubErr)
				return
			}
			adminReport, adminErr := svc.AdminContent.Heal(r.Context(), ac, dryRun)
			if adminErr != nil {
				service.HandleServiceError(w, r, adminErr)
				return
			}
			actionLabel := "Check"
			if !dryRun {
				actionLabel = "Heal"
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "All content healed", "type": "success"}}`)
			}
			pubData := partials.HealthReportData{
				DryRun:              pubReport.DryRun,
				ContentDataScanned:  pubReport.ContentDataScanned,
				ContentFieldScanned: pubReport.ContentFieldScanned,
				ContentDataRepairs:  len(pubReport.ContentDataRepairs),
				ContentFieldRepairs: len(pubReport.ContentFieldRepairs),
				MissingFields:       len(pubReport.MissingFields),
				DuplicateFields:     len(pubReport.DuplicateFields),
				OrphanedFields:      len(pubReport.OrphanedFields),
				DanglingPointers:    len(pubReport.DanglingPointers),
				OrphanedRouteRefs:   len(pubReport.OrphanedRouteRefs),
				UnroutedRoots:       len(pubReport.UnroutedRoots),
				RootlessContent:     len(pubReport.RootlessContent),
			}
			adminData := partials.HealthReportData{
				DryRun:              adminReport.DryRun,
				ContentDataScanned:  adminReport.ContentDataScanned,
				ContentFieldScanned: adminReport.ContentFieldScanned,
				ContentDataRepairs:  len(adminReport.ContentDataRepairs),
				ContentFieldRepairs: len(adminReport.ContentFieldRepairs),
				MissingFields:       len(adminReport.MissingFields),
				DuplicateFields:     len(adminReport.DuplicateFields),
				OrphanedFields:      len(adminReport.OrphanedFields),
				DanglingPointers:    len(adminReport.DanglingPointers),
				OrphanedRouteRefs:   len(adminReport.OrphanedRouteRefs),
				UnroutedRoots:       len(adminReport.UnroutedRoots),
				RootlessContent:     len(adminReport.RootlessContent),
			}
			Render(w, r, partials.ContentHealthReportCombined(actionLabel, pubData, adminData))

		default:
			http.Error(w, "Invalid scope", http.StatusBadRequest)
		}
	}
}

// ContentHealthPageData builds the layout data for the content health page.
func ContentHealthPageData(r *http.Request) layouts.AdminData {
	return NewAdminData(r, "Content Health")
}
