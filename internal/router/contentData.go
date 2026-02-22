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

// MoveContentDataRequest is the JSON body for POST /api/v1/contentdata/move.
type MoveContentDataRequest struct {
	NodeID      types.ContentID         `json:"node_id"`
	NewParentID types.NullableContentID `json:"new_parent_id"`
	Position    int                     `json:"position"`
}

// MoveContentDataResponse is the JSON response for POST /api/v1/contentdata/move.
type MoveContentDataResponse struct {
	NodeID      types.ContentID         `json:"node_id"`
	OldParentID types.NullableContentID `json:"old_parent_id"`
	NewParentID types.NullableContentID `json:"new_parent_id"`
	Position    int                     `json:"position"`
}

// ContentDataMoveHandler handles POST requests to move content data to a new parent.
func ContentDataMoveHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	apiMoveContentData(w, r, c)
}

// apiMoveContentData moves a content data node to a new parent at a given position.
func apiMoveContentData(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req MoveContentDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.NodeID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid node_id: %v", err), http.StatusBadRequest)
		return
	}
	if req.NewParentID.Valid {
		if err := req.NewParentID.ID.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid new_parent_id: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Position < 0 {
		http.Error(w, "position must be >= 0", http.StatusBadRequest)
		return
	}

	d := db.ConfigDB(c)

	// Fetch the node being moved
	node, err := d.GetContentData(req.NodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("node not found: %v", err), http.StatusBadRequest)
		return
	}

	// Cycle detection: walk from new_parent_id up the parent chain
	if req.NewParentID.Valid {
		cursor := req.NewParentID.ID
		for {
			if cursor == req.NodeID {
				http.Error(w, "cannot move a node under its own descendant", http.StatusBadRequest)
				return
			}
			ancestor, aErr := d.GetContentData(cursor)
			if aErr != nil {
				http.Error(w, fmt.Sprintf("failed to verify ancestry: %v", aErr), http.StatusInternalServerError)
				return
			}
			if !ancestor.ParentID.Valid {
				break
			}
			cursor = ancestor.ParentID.ID
		}
	}

	oldParentID := node.ParentID
	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)
	now := types.TimestampNow()

	// --- Unlink from old parent ---

	// Repair prev sibling's next pointer
	if node.PrevSiblingID.Valid {
		prev, pErr := d.GetContentData(node.PrevSiblingID.ID)
		if pErr != nil {
			http.Error(w, fmt.Sprintf("failed to fetch prev sibling: %v", pErr), http.StatusInternalServerError)
			return
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
			http.Error(w, fmt.Sprintf("failed to update prev sibling: %v", pErr), http.StatusInternalServerError)
			return
		}
	}

	// Repair next sibling's prev pointer
	if node.NextSiblingID.Valid {
		next, nErr := d.GetContentData(node.NextSiblingID.ID)
		if nErr != nil {
			http.Error(w, fmt.Sprintf("failed to fetch next sibling: %v", nErr), http.StatusInternalServerError)
			return
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
			http.Error(w, fmt.Sprintf("failed to update next sibling: %v", nErr), http.StatusInternalServerError)
			return
		}
	}

	// Update old parent's first_child_id if this node was the first child
	if node.ParentID.Valid {
		parent, pErr := d.GetContentData(node.ParentID.ID)
		if pErr != nil {
			http.Error(w, fmt.Sprintf("failed to fetch old parent: %v", pErr), http.StatusInternalServerError)
			return
		}
		if parent.FirstChildID.Valid && parent.FirstChildID.ID == req.NodeID {
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
				http.Error(w, fmt.Sprintf("failed to update old parent: %v", pErr), http.StatusInternalServerError)
				return
			}
		}
	}

	// --- Insert at position in new parent ---

	// Clear node's sibling pointers before reinsertion
	var newPrev types.NullableContentID
	var newNext types.NullableContentID

	if !req.NewParentID.Valid {
		// Moving to root level — just clear pointers and parent
		_, err = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: node.ContentDataID,
			ParentID:      req.NewParentID,
			FirstChildID:  node.FirstChildID,
			NextSiblingID: newNext,
			PrevSiblingID: newPrev,
			RouteID:       node.RouteID,
			DatatypeID:    node.DatatypeID,
			AuthorID:      node.AuthorID,
			Status:        node.Status,
			DateCreated:   node.DateCreated,
			DateModified:  now,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to update moved node: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		newParent, npErr := d.GetContentData(req.NewParentID.ID)
		if npErr != nil {
			http.Error(w, fmt.Sprintf("new parent not found: %v", npErr), http.StatusBadRequest)
			return
		}

		if !newParent.FirstChildID.Valid || req.Position == 0 {
			// Insert as first child
			oldFirstChildID := newParent.FirstChildID
			newNext = oldFirstChildID

			// Update new parent's first_child_id
			_, npErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: newParent.ContentDataID,
				ParentID:      newParent.ParentID,
				FirstChildID:  types.NullableContentID{ID: req.NodeID, Valid: true},
				NextSiblingID: newParent.NextSiblingID,
				PrevSiblingID: newParent.PrevSiblingID,
				RouteID:       newParent.RouteID,
				DatatypeID:    newParent.DatatypeID,
				AuthorID:      newParent.AuthorID,
				Status:        newParent.Status,
				DateCreated:   newParent.DateCreated,
				DateModified:  now,
			})
			if npErr != nil {
				http.Error(w, fmt.Sprintf("failed to update new parent: %v", npErr), http.StatusInternalServerError)
				return
			}

			// Update old first child's prev pointer
			if oldFirstChildID.Valid {
				oldFirst, ofErr := d.GetContentData(oldFirstChildID.ID)
				if ofErr != nil {
					http.Error(w, fmt.Sprintf("failed to fetch old first child: %v", ofErr), http.StatusInternalServerError)
					return
				}
				_, ofErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
					ContentDataID: oldFirst.ContentDataID,
					ParentID:      oldFirst.ParentID,
					FirstChildID:  oldFirst.FirstChildID,
					NextSiblingID: oldFirst.NextSiblingID,
					PrevSiblingID: types.NullableContentID{ID: req.NodeID, Valid: true},
					RouteID:       oldFirst.RouteID,
					DatatypeID:    oldFirst.DatatypeID,
					AuthorID:      oldFirst.AuthorID,
					Status:        oldFirst.Status,
					DateCreated:   oldFirst.DateCreated,
					DateModified:  now,
				})
				if ofErr != nil {
					http.Error(w, fmt.Sprintf("failed to update old first child prev: %v", ofErr), http.StatusInternalServerError)
					return
				}
			}
		} else {
			// Walk to position-1 to find the node to insert after
			current, wErr := d.GetContentData(newParent.FirstChildID.ID)
			if wErr != nil {
				http.Error(w, fmt.Sprintf("failed to walk new parent children: %v", wErr), http.StatusInternalServerError)
				return
			}
			for i := 0; i < req.Position-1 && current.NextSiblingID.Valid; i++ {
				current, wErr = d.GetContentData(current.NextSiblingID.ID)
				if wErr != nil {
					http.Error(w, fmt.Sprintf("failed to walk sibling chain: %v", wErr), http.StatusInternalServerError)
					return
				}
			}

			// Insert after current
			newPrev = types.NullableContentID{ID: current.ContentDataID, Valid: true}
			newNext = current.NextSiblingID

			// Update current's next pointer to the moved node
			_, wErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: current.ContentDataID,
				ParentID:      current.ParentID,
				FirstChildID:  current.FirstChildID,
				NextSiblingID: types.NullableContentID{ID: req.NodeID, Valid: true},
				PrevSiblingID: current.PrevSiblingID,
				RouteID:       current.RouteID,
				DatatypeID:    current.DatatypeID,
				AuthorID:      current.AuthorID,
				Status:        current.Status,
				DateCreated:   current.DateCreated,
				DateModified:  now,
			})
			if wErr != nil {
				http.Error(w, fmt.Sprintf("failed to update insert-after node: %v", wErr), http.StatusInternalServerError)
				return
			}

			// Update the node after current's prev pointer
			if newNext.Valid {
				afterNode, anErr := d.GetContentData(newNext.ID)
				if anErr != nil {
					http.Error(w, fmt.Sprintf("failed to fetch node after insertion point: %v", anErr), http.StatusInternalServerError)
					return
				}
				_, anErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
					ContentDataID: afterNode.ContentDataID,
					ParentID:      afterNode.ParentID,
					FirstChildID:  afterNode.FirstChildID,
					NextSiblingID: afterNode.NextSiblingID,
					PrevSiblingID: types.NullableContentID{ID: req.NodeID, Valid: true},
					RouteID:       afterNode.RouteID,
					DatatypeID:    afterNode.DatatypeID,
					AuthorID:      afterNode.AuthorID,
					Status:        afterNode.Status,
					DateCreated:   afterNode.DateCreated,
					DateModified:  now,
				})
				if anErr != nil {
					http.Error(w, fmt.Sprintf("failed to update after-insertion node: %v", anErr), http.StatusInternalServerError)
					return
				}
			}
		}

		// Update the moved node itself
		_, err = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: node.ContentDataID,
			ParentID:      req.NewParentID,
			FirstChildID:  node.FirstChildID,
			NextSiblingID: newNext,
			PrevSiblingID: newPrev,
			RouteID:       node.RouteID,
			DatatypeID:    node.DatatypeID,
			AuthorID:      node.AuthorID,
			Status:        node.Status,
			DateCreated:   node.DateCreated,
			DateModified:  now,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to update moved node: %v", err), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, MoveContentDataResponse{
		NodeID:      req.NodeID,
		OldParentID: oldParentID,
		NewParentID: req.NewParentID,
		Position:    req.Position,
	})
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
