package router

import (
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/query"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// QueryHandler handles content query requests by datatype name.
func QueryHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	datatypeName := r.PathValue("datatype")
	if datatypeName == "" {
		http.Error(w, "datatype name is required", http.StatusBadRequest)
		return
	}

	params := parseQueryParams(r, datatypeName)

	result, err := query.Execute(r.Context(), d, params)
	if err != nil {
		utility.DefaultLogger.Error("query execute failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	transformer := query.DefaultTransformer()
	data, err := transformer.TransformResultToJSON(result)
	if err != nil {
		utility.DefaultLogger.Error("query transform failed", err)
		http.Error(w, "failed to transform query result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func parseQueryParams(r *http.Request, datatypeName string) query.QueryParams {
	qp := r.URL.Query()

	var limit int64
	if l := qp.Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}

	var offset int64
	if o := qp.Get("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 64); err == nil {
			offset = parsed
		}
	}

	filters := query.ParseFilters(qp)

	return query.QueryParams{
		DatatypeName: datatypeName,
		Filters:      filters,
		Sort:         query.ParseSort(qp.Get("sort")),
		Limit:        limit,
		Offset:       offset,
		Locale:       qp.Get("locale"),
		Status:       qp.Get("status"),
	}
}
