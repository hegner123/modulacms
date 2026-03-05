package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
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
func UserReassignDeleteHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	apiUserReassignDelete(w, r, c)
}

// apiUserReassignDelete reassigns all authored content from a user to another
// (defaulting to the system user), then deletes the user.
func apiUserReassignDelete(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req UserReassignDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.UserID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid user_id: %v", err), http.StatusBadRequest)
		return
	}

	// Reject deletion of the system user.
	if req.UserID == types.SystemUserID {
		http.Error(w, "cannot delete the system user", http.StatusForbidden)
		return
	}

	d := db.ConfigDB(c)

	// Validate user exists.
	_, err := d.GetUser(req.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("user not found: %v", err), http.StatusNotFound)
		return
	}

	// Resolve reassign target (default to system user).
	reassignTo := req.ReassignTo
	if reassignTo == "" || reassignTo.IsZero() {
		reassignTo = types.SystemUserID
	}

	if err := reassignTo.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid reassign_to: %v", err), http.StatusBadRequest)
		return
	}

	if req.UserID == reassignTo {
		http.Error(w, "cannot reassign to the same user being deleted", http.StatusBadRequest)
		return
	}

	// Validate target user exists.
	_, err = d.GetUser(reassignTo)
	if err != nil {
		http.Error(w, fmt.Sprintf("reassign target not found: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get counts before reassignment.
	contentCount, err := d.CountContentDataByAuthor(ctx, req.UserID)
	if err != nil {
		utility.DefaultLogger.Error("reassign-delete: failed to count content data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	datatypeCount, err := d.CountDatatypesByAuthor(ctx, req.UserID)
	if err != nil {
		utility.DefaultLogger.Error("reassign-delete: failed to count datatypes", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminContentCount, err := d.CountAdminContentDataByAuthor(ctx, req.UserID)
	if err != nil {
		utility.DefaultLogger.Error("reassign-delete: failed to count admin content data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Bulk reassign all authored rows.
	if err := d.ReassignContentDataAuthor(ctx, reassignTo, req.UserID); err != nil {
		utility.DefaultLogger.Error("reassign-delete: failed to reassign content data", err)
		http.Error(w, fmt.Sprintf("failed to reassign content data: %v", err), http.StatusInternalServerError)
		return
	}
	if err := d.ReassignDatatypeAuthor(ctx, reassignTo, req.UserID); err != nil {
		utility.DefaultLogger.Error("reassign-delete: failed to reassign datatypes", err)
		http.Error(w, fmt.Sprintf("failed to reassign datatypes: %v", err), http.StatusInternalServerError)
		return
	}
	if err := d.ReassignAdminContentDataAuthor(ctx, reassignTo, req.UserID); err != nil {
		utility.DefaultLogger.Error("reassign-delete: failed to reassign admin content data", err)
		http.Error(w, fmt.Sprintf("failed to reassign admin content data: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete the user (FK constraints now satisfied).
	ac := middleware.AuditContextFromRequest(r, c)
	if err := d.DeleteUser(ctx, ac, req.UserID); err != nil {
		utility.DefaultLogger.Error("reassign-delete: failed to delete user", err)
		http.Error(w, fmt.Sprintf("failed to delete user: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, UserReassignDeleteResponse{
		DeletedUserID:              req.UserID,
		ReassignedTo:               reassignTo,
		ContentDataReassigned:      contentCount,
		DatatypesReassigned:        datatypeCount,
		AdminContentDataReassigned: adminContentCount,
	})
}
