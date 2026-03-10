package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminFieldTypeHandler handles CRUD operations for specific admin field type items.
func AdminFieldTypeHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminFieldType(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminFieldType(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminFieldType(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminFieldTypeHandler handles CRUD operations for specific admin field type items.
func AdminFieldTypeHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminFieldType(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminFieldType(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminFieldType(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiDeleteAdminFieldType handles DELETE requests for admin field types.
func apiDeleteAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err = svc.Schema.DeleteAdminFieldType(r.Context(), ac, aftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiDeleteAdminFieldType handles DELETE requests for admin field types.
func apiDeleteAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err = svc.Schema.DeleteAdminFieldType(r.Context(), ac, aftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiDeleteAdminFieldType handles DELETE requests for admin field types.
func apiDeleteAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err = svc.Schema.DeleteAdminFieldType(r.Context(), ac, aftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiDeleteAdminFieldType handles DELETE requests for admin field types.
func apiDeleteAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err = svc.Schema.DeleteAdminFieldType(r.Context(), ac, aftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiDeleteAdminFieldType handles DELETE requests for admin field types.
func apiDeleteAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err = svc.Schema.DeleteAdminFieldType(r.Context(), ac, aftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
