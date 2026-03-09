package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// UserReassignDeleteRequest is the JSON body for POST /api/v1/users/reassign-delete.
type UserReassignDeleteRequest struct {
	UserID     types.UserID `json:"user_id"`
	ReassignTo types.UserID `json:"reassign_to"`
}

// UserReassignDeleteResponse is the JSON response for POST /api/v1/users/reassign-delete.
type UserReassignDeleteResponse struct {
	DeletedUserID              types.UserID `json:"deleted_user_id"`
	ReassignedTo               types.UserID `json:"reassigned_to"`
	ContentDataReassigned      int64        `json:"content_data_reassigned"`
	DatatypesReassigned        int64        `json:"datatypes_reassigned"`
	AdminContentDataReassigned int64        `json:"admin_content_data_reassigned"`
}

// UserReassignDeleteHandler handles POST /api/v1/users/reassign-delete.
func UserReassignDeleteHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req UserReassignDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	c, cfgErr := svc.Config()
	if cfgErr != nil {
		service.HandleServiceError(w, r, cfgErr)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	result, err := svc.Users.ReassignDelete(r.Context(), ac, service.ReassignDeleteInput{
		UserID:     req.UserID,
		ReassignTo: req.ReassignTo,
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, UserReassignDeleteResponse{
		DeletedUserID:              result.DeletedUserID,
		ReassignedTo:               result.ReassignedTo,
		ContentDataReassigned:      result.ContentDataReassigned,
		DatatypesReassigned:        result.DatatypesReassigned,
		AdminContentDataReassigned: result.AdminContentDataReassigned,
	})
}
