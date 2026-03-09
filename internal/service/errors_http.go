package service

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/utility"
)

// HandleServiceError maps service-layer errors to HTTP responses.
// For HTMX requests (HX-Request header present), errors are returned via
// HX-Trigger toast events instead of JSON bodies.
func HandleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	isHTMX := r.Header.Get("HX-Request") == "true"

	switch {
	case IsNotFound(err):
		if isHTMX {
			writeHTMXError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSONError(w, http.StatusNotFound, err.Error())

	case IsValidation(err):
		var ve *ValidationError
		ok := extractValidation(err, &ve)
		if ok && isHTMX {
			writeHTMXError(w, http.StatusUnprocessableEntity, ve.Error())
			return
		}
		if ok {
			writeJSONValidation(w, ve)
			return
		}
		writeJSONError(w, http.StatusUnprocessableEntity, err.Error())

	case IsConflict(err):
		if isHTMX {
			writeHTMXError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSONError(w, http.StatusConflict, err.Error())

	case IsForbidden(err):
		if isHTMX {
			writeHTMXError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusForbidden, err.Error())

	case IsUnauthorized(err):
		if isHTMX {
			writeHTMXError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeJSONError(w, http.StatusUnauthorized, err.Error())

	default:
		utility.DefaultLogger.Error("internal service error", err,
			"path", r.URL.Path,
			"method", r.Method,
		)
		if isHTMX {
			writeHTMXError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func writeJSONValidation(w http.ResponseWriter, ve *ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error":  ve.Error(),
		"fields": ve.Errors,
	})
}

func writeHTMXError(w http.ResponseWriter, status int, message string) {
	escaped, err := json.Marshal(message)
	if err != nil {
		escaped = []byte(`"internal error"`)
	}
	w.Header().Set("HX-Trigger", `{"showToast": {"message": `+string(escaped)+`, "type": "error"}}`)
	w.WriteHeader(status)
}

func extractValidation(err error, target **ValidationError) bool {
	var ve *ValidationError
	if IsValidation(err) {
		// errors.As through the chain
		for unwrap := err; unwrap != nil; {
			if v, ok := unwrap.(*ValidationError); ok {
				*target = v
				return true
			}
			if u, ok := unwrap.(interface{ Unwrap() error }); ok {
				unwrap = u.Unwrap()
			} else {
				break
			}
		}
		// IsValidation returned true but we couldn't extract — use a synthetic one
		*target = &ValidationError{Errors: []FieldError{{Field: "unknown", Message: err.Error()}}}
		_ = ve
		return true
	}
	return false
}
