package router

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// writeServiceError maps a service-layer error to the appropriate HTTP status
// code and writes the error response. It logs the error via the default logger.
//
// Mapping:
//
//	ValidationError -> 400 Bad Request
//	NotFoundError   -> 404 Not Found
//	ForbiddenError  -> 403 Forbidden
//	ConflictError   -> 409 Conflict
//	InternalError   -> 500 Internal Server Error
//	(unknown)       -> 500 Internal Server Error
func writeServiceError(w http.ResponseWriter, err error) {
	utility.DefaultLogger.Error("", err)

	switch {
	case service.IsValidation(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case service.IsNotFound(err):
		http.Error(w, err.Error(), http.StatusNotFound)
	case service.IsForbidden(err):
		http.Error(w, err.Error(), http.StatusForbidden)
	case service.IsConflict(err):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
