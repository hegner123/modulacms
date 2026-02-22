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

	// Append new node to parent's sibling chain
	if createdAdminContentData.ParentID.Valid {
		ctx := r.Context()
		now := types.TimestampNow()
		parent, pErr := d.GetAdminContentData(createdAdminContentData.ParentID.ID)
		if pErr != nil {
			utility.DefaultLogger.Error("create: failed to fetch parent", pErr)
			http.Error(w, fmt.Sprintf("failed to fetch parent: %v", pErr), http.StatusInternalServerError)
			return pErr
		}

		if !parent.FirstChildID.Valid {
			// Parent has no children yet — set first_child_id to new node
			_, pErr = d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: parent.AdminContentDataID,
				ParentID:           parent.ParentID,
				FirstChildID:       types.NullableAdminContentID{ID: createdAdminContentData.AdminContentDataID, Valid: true},
				NextSiblingID:      parent.NextSiblingID,
				PrevSiblingID:      parent.PrevSiblingID,
				AdminRouteID:       parent.AdminRouteID,
				AdminDatatypeID:    parent.AdminDatatypeID,
				AuthorID:           parent.AuthorID,
				Status:             parent.Status,
				DateCreated:        parent.DateCreated,
				DateModified:       now,
			})
			if pErr != nil {
				utility.DefaultLogger.Error("create: failed to set parent first_child_id", pErr)
				http.Error(w, fmt.Sprintf("failed to update parent: %v", pErr), http.StatusInternalServerError)
				return pErr
			}
		} else {
			// Walk the sibling chain to find the last sibling
			last, wErr := d.GetAdminContentData(parent.FirstChildID.ID)
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to fetch first child", wErr)
				http.Error(w, fmt.Sprintf("failed to fetch first child: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
			for last.NextSiblingID.Valid {
				last, wErr = d.GetAdminContentData(last.NextSiblingID.ID)
				if wErr != nil {
					utility.DefaultLogger.Error("create: failed to walk sibling chain", wErr)
					http.Error(w, fmt.Sprintf("failed to walk sibling chain: %v", wErr), http.StatusInternalServerError)
					return wErr
				}
			}
			// Link last sibling to new node
			_, wErr = d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: last.AdminContentDataID,
				ParentID:           last.ParentID,
				FirstChildID:       last.FirstChildID,
				NextSiblingID:      types.NullableAdminContentID{ID: createdAdminContentData.AdminContentDataID, Valid: true},
				PrevSiblingID:      last.PrevSiblingID,
				AdminRouteID:       last.AdminRouteID,
				AdminDatatypeID:    last.AdminDatatypeID,
				AuthorID:           last.AuthorID,
				Status:             last.Status,
				DateCreated:        last.DateCreated,
				DateModified:       now,
			})
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to link last sibling", wErr)
				http.Error(w, fmt.Sprintf("failed to link last sibling: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
			// Set new node's prev_sibling_id to last sibling
			_, wErr = d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: createdAdminContentData.AdminContentDataID,
				ParentID:           createdAdminContentData.ParentID,
				FirstChildID:       createdAdminContentData.FirstChildID,
				NextSiblingID:      createdAdminContentData.NextSiblingID,
				PrevSiblingID:      types.NullableAdminContentID{ID: last.AdminContentDataID, Valid: true},
				AdminRouteID:       createdAdminContentData.AdminRouteID,
				AdminDatatypeID:    createdAdminContentData.AdminDatatypeID,
				AuthorID:           createdAdminContentData.AuthorID,
				Status:             createdAdminContentData.Status,
				DateCreated:        createdAdminContentData.DateCreated,
				DateModified:       now,
			})
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to set new node prev_sibling_id", wErr)
				http.Error(w, fmt.Sprintf("failed to update new node: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
			// Re-fetch to return updated state
			createdAdminContentData, wErr = d.GetAdminContentData(createdAdminContentData.AdminContentDataID)
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to re-fetch created node", wErr)
				http.Error(w, fmt.Sprintf("failed to re-fetch created node: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
		}
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

	// Fetch the node to repair sibling pointers before deleting
	node, err := d.GetAdminContentData(acdID)
	if err != nil {
		utility.DefaultLogger.Error("delete: failed to fetch node", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	ctx := r.Context()
	now := types.TimestampNow()

	// Repair prev sibling's next pointer
	if node.PrevSiblingID.Valid {
		prev, pErr := d.GetAdminContentData(node.PrevSiblingID.ID)
		if pErr != nil {
			utility.DefaultLogger.Error("delete: failed to fetch prev sibling", pErr)
			http.Error(w, fmt.Sprintf("failed to fetch prev sibling: %v", pErr), http.StatusInternalServerError)
			return pErr
		}
		_, pErr = d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
			AdminContentDataID: prev.AdminContentDataID,
			ParentID:           prev.ParentID,
			FirstChildID:       prev.FirstChildID,
			NextSiblingID:      node.NextSiblingID,
			PrevSiblingID:      prev.PrevSiblingID,
			AdminRouteID:       prev.AdminRouteID,
			AdminDatatypeID:    prev.AdminDatatypeID,
			AuthorID:           prev.AuthorID,
			Status:             prev.Status,
			DateCreated:        prev.DateCreated,
			DateModified:       now,
		})
		if pErr != nil {
			utility.DefaultLogger.Error("delete: failed to update prev sibling", pErr)
			http.Error(w, fmt.Sprintf("failed to update prev sibling: %v", pErr), http.StatusInternalServerError)
			return pErr
		}
	}

	// Repair next sibling's prev pointer
	if node.NextSiblingID.Valid {
		next, nErr := d.GetAdminContentData(node.NextSiblingID.ID)
		if nErr != nil {
			utility.DefaultLogger.Error("delete: failed to fetch next sibling", nErr)
			http.Error(w, fmt.Sprintf("failed to fetch next sibling: %v", nErr), http.StatusInternalServerError)
			return nErr
		}
		_, nErr = d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
			AdminContentDataID: next.AdminContentDataID,
			ParentID:           next.ParentID,
			FirstChildID:       next.FirstChildID,
			NextSiblingID:      next.NextSiblingID,
			PrevSiblingID:      node.PrevSiblingID,
			AdminRouteID:       next.AdminRouteID,
			AdminDatatypeID:    next.AdminDatatypeID,
			AuthorID:           next.AuthorID,
			Status:             next.Status,
			DateCreated:        next.DateCreated,
			DateModified:       now,
		})
		if nErr != nil {
			utility.DefaultLogger.Error("delete: failed to update next sibling", nErr)
			http.Error(w, fmt.Sprintf("failed to update next sibling: %v", nErr), http.StatusInternalServerError)
			return nErr
		}
	}

	// If this node is the parent's first child, update parent
	if node.ParentID.Valid {
		parent, pErr := d.GetAdminContentData(node.ParentID.ID)
		if pErr != nil {
			utility.DefaultLogger.Error("delete: failed to fetch parent", pErr)
			http.Error(w, fmt.Sprintf("failed to fetch parent: %v", pErr), http.StatusInternalServerError)
			return pErr
		}
		if parent.FirstChildID.Valid && parent.FirstChildID.ID == acdID {
			_, pErr = d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: parent.AdminContentDataID,
				ParentID:           parent.ParentID,
				FirstChildID:       node.NextSiblingID,
				NextSiblingID:      parent.NextSiblingID,
				PrevSiblingID:      parent.PrevSiblingID,
				AdminRouteID:       parent.AdminRouteID,
				AdminDatatypeID:    parent.AdminDatatypeID,
				AuthorID:           parent.AuthorID,
				Status:             parent.Status,
				DateCreated:        parent.DateCreated,
				DateModified:       now,
			})
			if pErr != nil {
				utility.DefaultLogger.Error("delete: failed to update parent first_child_id", pErr)
				http.Error(w, fmt.Sprintf("failed to update parent: %v", pErr), http.StatusInternalServerError)
				return pErr
			}
		}
	}

	err = d.DeleteAdminContentData(ctx, ac, acdID)
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
