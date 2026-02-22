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

// ContentDatasHandler handles CRUD operations that do not require a specific data ID.
func ContentDatasHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListContentDataPaginated(w, r, c)
		} else {
			apiListContentData(w, c)
		}
	case http.MethodPost:
		apiCreateContentData(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ContentDataHandler handles CRUD operations for specific content data items.
func ContentDataHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetContentData(w, r, c)
	case http.MethodPost:
		apiCreateContentData(w, r, c)
	case http.MethodPut:
		apiUpdateContentData(w, r, c)
	case http.MethodDelete:
		apiDeleteContentData(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetContentData handles GET requests for a single content data
func apiGetContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	contentData, err := d.GetContentData(cdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contentData)
	return nil
}

// apiListContentData handles GET requests for listing content data
func apiListContentData(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	contentDataList, err := d.ListContentData()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contentDataList)
	return nil
}

// apiCreateContentData handles POST requests to create new content data
func apiCreateContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newContentData db.CreateContentDataParams
	err := json.NewDecoder(r.Body).Decode(&newContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdContentData, err := d.CreateContentData(r.Context(), ac, newContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Append new node to parent's sibling chain
	if createdContentData.ParentID.Valid {
		ctx := r.Context()
		now := types.TimestampNow()
		parent, pErr := d.GetContentData(createdContentData.ParentID.ID)
		if pErr != nil {
			utility.DefaultLogger.Error("create: failed to fetch parent", pErr)
			http.Error(w, fmt.Sprintf("failed to fetch parent: %v", pErr), http.StatusInternalServerError)
			return pErr
		}

		if !parent.FirstChildID.Valid {
			// Parent has no children yet — set first_child_id to new node
			_, pErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: parent.ContentDataID,
				ParentID:      parent.ParentID,
				FirstChildID:  types.NullableContentID{ID: createdContentData.ContentDataID, Valid: true},
				NextSiblingID: parent.NextSiblingID,
				PrevSiblingID: parent.PrevSiblingID,
				RouteID:       parent.RouteID,
				DatatypeID:    parent.DatatypeID,
				AuthorID:      parent.AuthorID,
				Status:        parent.Status,
				DateCreated:   parent.DateCreated,
				DateModified:  now,
			})
			if pErr != nil {
				utility.DefaultLogger.Error("create: failed to set parent first_child_id", pErr)
				http.Error(w, fmt.Sprintf("failed to update parent: %v", pErr), http.StatusInternalServerError)
				return pErr
			}
		} else {
			// Walk the sibling chain to find the last sibling
			last, wErr := d.GetContentData(parent.FirstChildID.ID)
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to fetch first child", wErr)
				http.Error(w, fmt.Sprintf("failed to fetch first child: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
			for last.NextSiblingID.Valid {
				last, wErr = d.GetContentData(last.NextSiblingID.ID)
				if wErr != nil {
					utility.DefaultLogger.Error("create: failed to walk sibling chain", wErr)
					http.Error(w, fmt.Sprintf("failed to walk sibling chain: %v", wErr), http.StatusInternalServerError)
					return wErr
				}
			}
			// Link last sibling to new node
			_, wErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: last.ContentDataID,
				ParentID:      last.ParentID,
				FirstChildID:  last.FirstChildID,
				NextSiblingID: types.NullableContentID{ID: createdContentData.ContentDataID, Valid: true},
				PrevSiblingID: last.PrevSiblingID,
				RouteID:       last.RouteID,
				DatatypeID:    last.DatatypeID,
				AuthorID:      last.AuthorID,
				Status:        last.Status,
				DateCreated:   last.DateCreated,
				DateModified:  now,
			})
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to link last sibling", wErr)
				http.Error(w, fmt.Sprintf("failed to link last sibling: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
			// Set new node's prev_sibling_id to last sibling
			_, wErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: createdContentData.ContentDataID,
				ParentID:      createdContentData.ParentID,
				FirstChildID:  createdContentData.FirstChildID,
				NextSiblingID: createdContentData.NextSiblingID,
				PrevSiblingID: types.NullableContentID{ID: last.ContentDataID, Valid: true},
				RouteID:       createdContentData.RouteID,
				DatatypeID:    createdContentData.DatatypeID,
				AuthorID:      createdContentData.AuthorID,
				Status:        createdContentData.Status,
				DateCreated:   createdContentData.DateCreated,
				DateModified:  now,
			})
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to set new node prev_sibling_id", wErr)
				http.Error(w, fmt.Sprintf("failed to update new node: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
			// Re-fetch to return updated state
			createdContentData, wErr = d.GetContentData(createdContentData.ContentDataID)
			if wErr != nil {
				utility.DefaultLogger.Error("create: failed to re-fetch created node", wErr)
				http.Error(w, fmt.Sprintf("failed to re-fetch created node: %v", wErr), http.StatusInternalServerError)
				return wErr
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdContentData)
	return nil
}

// apiUpdateContentData handles PUT requests to update existing content data
func apiUpdateContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateContentData db.UpdateContentDataParams
	err := json.NewDecoder(r.Body).Decode(&updateContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateContentData(r.Context(), ac, updateContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetContentData(updateContentData.ContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated content data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteContentData handles DELETE requests for content data
func apiDeleteContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Fetch the node to repair sibling pointers before deleting
	node, err := d.GetContentData(cdID)
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
		prev, pErr := d.GetContentData(node.PrevSiblingID.ID)
		if pErr != nil {
			utility.DefaultLogger.Error("delete: failed to fetch prev sibling", pErr)
			http.Error(w, fmt.Sprintf("failed to fetch prev sibling: %v", pErr), http.StatusInternalServerError)
			return pErr
		}
		_, pErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: prev.ContentDataID,
			ParentID:      prev.ParentID,
			FirstChildID:  prev.FirstChildID,
			NextSiblingID: node.NextSiblingID,
			PrevSiblingID: prev.PrevSiblingID,
			RouteID:       prev.RouteID,
			DatatypeID:    prev.DatatypeID,
			AuthorID:      prev.AuthorID,
			Status:        prev.Status,
			DateCreated:   prev.DateCreated,
			DateModified:  now,
		})
		if pErr != nil {
			utility.DefaultLogger.Error("delete: failed to update prev sibling", pErr)
			http.Error(w, fmt.Sprintf("failed to update prev sibling: %v", pErr), http.StatusInternalServerError)
			return pErr
		}
	}

	// Repair next sibling's prev pointer
	if node.NextSiblingID.Valid {
		next, nErr := d.GetContentData(node.NextSiblingID.ID)
		if nErr != nil {
			utility.DefaultLogger.Error("delete: failed to fetch next sibling", nErr)
			http.Error(w, fmt.Sprintf("failed to fetch next sibling: %v", nErr), http.StatusInternalServerError)
			return nErr
		}
		_, nErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: next.ContentDataID,
			ParentID:      next.ParentID,
			FirstChildID:  next.FirstChildID,
			NextSiblingID: next.NextSiblingID,
			PrevSiblingID: node.PrevSiblingID,
			RouteID:       next.RouteID,
			DatatypeID:    next.DatatypeID,
			AuthorID:      next.AuthorID,
			Status:        next.Status,
			DateCreated:   next.DateCreated,
			DateModified:  now,
		})
		if nErr != nil {
			utility.DefaultLogger.Error("delete: failed to update next sibling", nErr)
			http.Error(w, fmt.Sprintf("failed to update next sibling: %v", nErr), http.StatusInternalServerError)
			return nErr
		}
	}

	// If this node is the parent's first child, update parent
	if node.ParentID.Valid {
		parent, pErr := d.GetContentData(node.ParentID.ID)
		if pErr != nil {
			utility.DefaultLogger.Error("delete: failed to fetch parent", pErr)
			http.Error(w, fmt.Sprintf("failed to fetch parent: %v", pErr), http.StatusInternalServerError)
			return pErr
		}
		if parent.FirstChildID.Valid && parent.FirstChildID.ID == cdID {
			_, pErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: parent.ContentDataID,
				ParentID:      parent.ParentID,
				FirstChildID:  node.NextSiblingID,
				NextSiblingID: parent.NextSiblingID,
				PrevSiblingID: parent.PrevSiblingID,
				RouteID:       parent.RouteID,
				DatatypeID:    parent.DatatypeID,
				AuthorID:      parent.AuthorID,
				Status:        parent.Status,
				DateCreated:   parent.DateCreated,
				DateModified:  now,
			})
			if pErr != nil {
				utility.DefaultLogger.Error("delete: failed to update parent first_child_id", pErr)
				http.Error(w, fmt.Sprintf("failed to update parent: %v", pErr), http.StatusInternalServerError)
				return pErr
			}
		}
	}

	err = d.DeleteContentData(ctx, ac, cdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListContentDataPaginated handles GET requests for listing content data with pagination.
func apiListContentDataPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListContentDataPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountContentData()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.ContentData]{
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

// ReorderContentDataRequest is the JSON body for POST /api/v1/contentdata/reorder.
type ReorderContentDataRequest struct {
	ParentID   types.NullableContentID `json:"parent_id"`
	OrderedIDs []types.ContentID       `json:"ordered_ids"`
}

// ReorderContentDataResponse is the JSON response for POST /api/v1/contentdata/reorder.
type ReorderContentDataResponse struct {
	Updated  int                     `json:"updated"`
	ParentID types.NullableContentID `json:"parent_id"`
}

// ContentDataReorderHandler handles POST requests to reorder content data siblings.
func ContentDataReorderHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	apiReorderContentData(w, r, c)
}

// apiReorderContentData atomically reorders sibling content data nodes under a parent.
func apiReorderContentData(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ReorderContentDataRequest
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
			http.Error(w, fmt.Sprintf("invalid content_data_id %s: %v", id, err), http.StatusBadRequest)
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
	nodes := make([]db.ContentData, 0, len(req.OrderedIDs))
	for _, id := range req.OrderedIDs {
		node, err := d.GetContentData(id)
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
		parent, err := d.GetContentData(types.ContentID(req.ParentID.ID))
		if err != nil {
			utility.DefaultLogger.Error("reorder: failed to fetch parent", err)
			http.Error(w, fmt.Sprintf("failed to fetch parent: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: parent.ContentDataID,
			ParentID:      parent.ParentID,
			FirstChildID:  types.NullableContentID{ID: req.OrderedIDs[0], Valid: true},
			NextSiblingID: parent.NextSiblingID,
			PrevSiblingID: parent.PrevSiblingID,
			RouteID:       parent.RouteID,
			DatatypeID:    parent.DatatypeID,
			AuthorID:      parent.AuthorID,
			Status:        parent.Status,
			DateCreated:   parent.DateCreated,
			DateModified:  now,
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
		var prevSibling types.NullableContentID
		var nextSibling types.NullableContentID

		if i > 0 {
			prevSibling = types.NullableContentID{ID: req.OrderedIDs[i-1], Valid: true}
		}
		if i < lastIdx {
			nextSibling = types.NullableContentID{ID: req.OrderedIDs[i+1], Valid: true}
		}

		_, err := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: node.ContentDataID,
			ParentID:      node.ParentID,
			FirstChildID:  node.FirstChildID,
			NextSiblingID: nextSibling,
			PrevSiblingID: prevSibling,
			RouteID:       node.RouteID,
			DatatypeID:    node.DatatypeID,
			AuthorID:      node.AuthorID,
			Status:        node.Status,
			DateCreated:   node.DateCreated,
			DateModified:  now,
		})
		if err != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("reorder: failed to update node %s", node.ContentDataID), err)
			http.Error(w, fmt.Sprintf("failed to update node %s: %v", node.ContentDataID, err), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, ReorderContentDataResponse{
		Updated:  len(req.OrderedIDs),
		ParentID: req.ParentID,
	})
}
