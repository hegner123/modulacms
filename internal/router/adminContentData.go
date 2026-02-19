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

// AdminContentDatas handles CRUD operations that do not require a specific user ID.
func AdminContentDatasHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			err := apiListAdminContentDataPaginated(w, r, c)
			if err != nil {
				return
			}
		} else {
			err := apiListAdminContentData(w, r, c)
			if err != nil {
				return
			}
		}
	case http.MethodPost:
		err := apiCreateAdminContentData(w, r, c)
		if err != nil {
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func AdminContentDataHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodPost:
		err := apiCreateAdminContentData(w, r, c)
		if err != nil {
			return
		}
	case http.MethodPut:
		err := apiUpdateAdminContentData(w, r, c)
		if err != nil {
			return
		}
	case http.MethodDelete:
		err := apiDeleteAdminContentData(w, r, c)
		if err != nil {
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListAdminContentData handles GET requests for listing admin content data
func apiListAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	if r == nil {
		err := fmt.Errorf("request error")
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err

	}

	adminContentDataList, err := d.ListAdminContentData()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(adminContentDataList)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiCreateAdminContentData handles POST requests to create new admin content data
func apiCreateAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newAdminContentData db.CreateAdminContentDataParams
	err := json.NewDecoder(r.Body).Decode(&newAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdAdminContentData, err := d.CreateAdminContentData(r.Context(), ac, newAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiUpdateAdminContentData handles PUT requests to update existing admin content data
func apiUpdateAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateAdminContentData db.UpdateAdminContentDataParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	updateAdminContentData.DateModified = types.TimestampNow()

	oldData, err := d.GetAdminContentData(updateAdminContentData.AdminContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	_, err = json.Marshal(oldData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateAdminContentData(r.Context(), ac, updateAdminContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetAdminContentData(updateAdminContentData.AdminContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated admin content data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteAdminContentData handles DELETE requests for admin content data
func apiDeleteAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	acdID := types.AdminContentID(q)
	if err := acdID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteAdminContentData(r.Context(), ac, acdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListAdminContentDataPaginated handles GET requests for listing admin content data with pagination
func apiListAdminContentDataPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListAdminContentDataPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountAdminContentData()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.AdminContentData]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// ReorderAdminContentDataRequest is the JSON body for POST /api/v1/admincontentdatas/reorder.
type ReorderAdminContentDataRequest struct {
	ParentID   types.NullableAdminContentID `json:"parent_id"`
	OrderedIDs []types.AdminContentID       `json:"ordered_ids"`
}

// ReorderAdminContentDataResponse is the JSON response for POST /api/v1/admincontentdatas/reorder.
type ReorderAdminContentDataResponse struct {
	Updated  int                          `json:"updated"`
	ParentID types.NullableAdminContentID `json:"parent_id"`
}

// AdminContentDataReorderHandler handles POST requests to reorder admin content data siblings.
func AdminContentDataReorderHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	apiReorderAdminContentData(w, r, c)
}

// apiReorderAdminContentData atomically reorders sibling admin content data nodes under a parent.
func apiReorderAdminContentData(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ReorderAdminContentDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.OrderedIDs) == 0 {
		http.Error(w, "ordered_ids must not be empty", http.StatusBadRequest)
		return
	}

	// Validate each ID and reject duplicates
	seen := make(map[string]struct{}, len(req.OrderedIDs))
	for _, id := range req.OrderedIDs {
		if err := id.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid admin_content_data_id %s: %v", id, err), http.StatusBadRequest)
			return
		}
		s := string(id)
		if _, exists := seen[s]; exists {
			http.Error(w, fmt.Sprintf("duplicate id: %s", id), http.StatusBadRequest)
			return
		}
		seen[s] = struct{}{}
	}

	d := db.ConfigDB(c)

	// Fetch all nodes and verify parent ownership
	nodes := make([]db.AdminContentData, 0, len(req.OrderedIDs))
	for _, id := range req.OrderedIDs {
		node, err := d.GetAdminContentData(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("node not found: %s", id), http.StatusBadRequest)
			return
		}
		if node.ParentID != req.ParentID {
			http.Error(w, fmt.Sprintf("node %s does not belong to parent %s", id, req.ParentID), http.StatusBadRequest)
			return
		}
		nodes = append(nodes, *node)
	}

	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)
	now := types.TimestampNow()

	// Update parent's first_child_id if parent is non-null
	if req.ParentID.Valid {
		parent, err := d.GetAdminContentData(types.AdminContentID(req.ParentID.ID))
		if err != nil {
			utility.DefaultLogger.Error("reorder: failed to fetch parent", err)
			http.Error(w, fmt.Sprintf("failed to fetch parent: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
			AdminContentDataID: parent.AdminContentDataID,
			ParentID:           parent.ParentID,
			FirstChildID:       types.NullableAdminContentID{ID: req.OrderedIDs[0], Valid: true},
			NextSiblingID:      parent.NextSiblingID,
			PrevSiblingID:      parent.PrevSiblingID,
			AdminRouteID:       parent.AdminRouteID,
			AdminDatatypeID:    parent.AdminDatatypeID,
			AuthorID:           parent.AuthorID,
			Status:             parent.Status,
			DateCreated:        parent.DateCreated,
			DateModified:       now,
		})
		if err != nil {
			utility.DefaultLogger.Error("reorder: failed to update parent first_child_id", err)
			http.Error(w, fmt.Sprintf("failed to update parent: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Update sibling pointers for each node
	lastIdx := len(req.OrderedIDs) - 1
	for i, node := range nodes {
		var prevSibling types.NullableAdminContentID
		var nextSibling types.NullableAdminContentID

		if i > 0 {
			prevSibling = types.NullableAdminContentID{ID: req.OrderedIDs[i-1], Valid: true}
		}
		if i < lastIdx {
			nextSibling = types.NullableAdminContentID{ID: req.OrderedIDs[i+1], Valid: true}
		}

		_, err := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
			AdminContentDataID: node.AdminContentDataID,
			ParentID:           node.ParentID,
			FirstChildID:       node.FirstChildID,
			NextSiblingID:      nextSibling,
			PrevSiblingID:      prevSibling,
			AdminRouteID:       node.AdminRouteID,
			AdminDatatypeID:    node.AdminDatatypeID,
			AuthorID:           node.AuthorID,
			Status:             node.Status,
			DateCreated:        node.DateCreated,
			DateModified:       now,
		})
		if err != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("reorder: failed to update node %s", node.AdminContentDataID), err)
			http.Error(w, fmt.Sprintf("failed to update node %s: %v", node.AdminContentDataID, err), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  len(req.OrderedIDs),
		ParentID: req.ParentID,
	})
}
