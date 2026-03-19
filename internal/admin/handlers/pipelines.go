package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// PipelinesListHandler renders the pipelines viewer page.
// Fetches dry-run results (active chains) from the plugin registry.
// Supports optional ?table= query parameter to filter by table name.
func PipelinesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chains, err := svc.Plugins.DryRunPipelines(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list pipeline chains", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Collect distinct table names for the filter dropdown.
		tableSet := make(map[string]bool)
		for _, c := range chains {
			tableSet[c.Table] = true
		}
		tables := make([]string, 0, len(tableSet))
		for t := range tableSet {
			tables = append(tables, t)
		}

		// Apply table filter if provided.
		filterTable := r.URL.Query().Get("table")
		var filtered []service.PipelineChain
		if filterTable != "" {
			for _, c := range chains {
				if c.Table == filterTable {
					filtered = append(filtered, c)
				}
			}
		} else {
			filtered = chains
		}

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Pipelines"}`)
			Render(w, r, pages.PipelinesListContent(filtered, tables, filterTable))
			return
		}

		layout := NewAdminData(r, "Pipelines")
		Render(w, r, pages.PipelinesList(layout, filtered, tables, filterTable))
	}
}

// PipelineDetailHandler renders the detail view for a single pipeline chain.
// The chain key is passed as the path parameter {key} in the format "table.phase_operation".
func PipelineDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if key == "" {
			http.NotFound(w, r)
			return
		}

		chains, err := svc.Plugins.DryRunPipelines(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list pipeline chains", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var found *service.PipelineChain
		for i := range chains {
			if chains[i].Key == key {
				found = &chains[i]
				break
			}
		}

		if found == nil {
			http.NotFound(w, r)
			return
		}

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Pipeline: `+key+`"}`)
			Render(w, r, pages.PipelineDetailContent(found))
			return
		}

		layout := NewAdminData(r, "Pipeline: "+key)
		Render(w, r, pages.PipelineDetail(layout, found))
	}
}
