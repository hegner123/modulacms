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

// TreeSaveRequest is the JSON body for POST /api/v1/content/tree.
type TreeSaveRequest struct {
	ContentID types.ContentID   `json:"content_id"`
	Creates   []TreeNodeCreate  `json:"creates"`
	Updates   []TreeNodeUpdate  `json:"updates"`
	Deletes   []types.ContentID `json:"deletes"`
}

// TreeNodeCreate describes a new content_data node to insert.
// ClientID is the client-generated UUID; the server generates a ULID and returns
// the mapping in TreeSaveResponse.IDMap.
type TreeNodeCreate struct {
	ClientID      string  `json:"client_id"`
	DatatypeID    string  `json:"datatype_id"`
	ParentID      *string `json:"parent_id"`
	FirstChildID  *string `json:"first_child_id"`
	NextSiblingID *string `json:"next_sibling_id"`
	PrevSiblingID *string `json:"prev_sibling_id"`
}

// TreeNodeUpdate describes pointer-field changes for a single content_data node.
type TreeNodeUpdate struct {
	ContentDataID types.ContentID `json:"content_data_id"`
	ParentID      *string         `json:"parent_id"`
	FirstChildID  *string         `json:"first_child_id"`
	NextSiblingID *string         `json:"next_sibling_id"`
	PrevSiblingID *string         `json:"prev_sibling_id"`
}

// TreeSaveResponse summarises the results of a tree save operation.
type TreeSaveResponse struct {
	Created int               `json:"created"`
	Updated int               `json:"updated"`
	Deleted int               `json:"deleted"`
	IDMap   map[string]string `json:"id_map,omitempty"`
	Errors  []string          `json:"errors,omitempty"`
}

// parseNullableContentID converts a *string from JSON into a NullableContentID.
// nil pointer means SQL NULL; non-nil string is parsed as a ContentID.
func parseNullableContentID(s *string) (types.NullableContentID, error) {
	if s == nil {
		return types.NullableContentID{Valid: false}, nil
	}
	id := types.ContentID(*s)
	if err := id.Validate(); err != nil {
		return types.NullableContentID{}, fmt.Errorf("invalid content_id %q: %w", *s, err)
	}
	return types.NullableContentID{ID: id, Valid: true}, nil
}

// remapPointer replaces a client UUID in a pointer string with its server ULID
// if a mapping exists. Returns the (possibly remapped) pointer unchanged otherwise.
func remapPointer(ptr *string, idMap map[string]string) *string {
	if ptr == nil {
		return nil
	}
	if mapped, ok := idMap[*ptr]; ok {
		return &mapped
	}
	return ptr
}

// ContentTreeSaveHandler handles POST /api/v1/content/tree — bulk tree
// creates, pointer updates, and deletes in a single request.
//
// Creates use a two-phase approach because CreateContentData generates its own
// ULID internally. Phase 1a inserts rows with NULL pointers and collects the
// server-generated IDs. Phase 1b updates those rows with the correct remapped
// pointers. This ensures all new nodes exist before any cross-references are set.
func ContentTreeSaveHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req TreeSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if req.ContentID.IsZero() {
		http.Error(w, "content_id is required", http.StatusBadRequest)
		return
	}

	if len(req.Creates) == 0 && len(req.Updates) == 0 && len(req.Deletes) == 0 {
		http.Error(w, "at least one create, update, or delete is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)
	d := db.ConfigDB(c)
	now := types.TimestampNow()
	nullPointer := types.NullableContentID{Valid: false}

	resp := TreeSaveResponse{}

	// Resolve routeID and authorID for new blocks.
	var parentRouteID types.NullableRouteID
	parentContent, err := d.GetContentData(req.ContentID)
	if err != nil {
		resp.Errors = append(resp.Errors, fmt.Sprintf("get parent %s: %v", req.ContentID, err))
	} else {
		parentRouteID = parentContent.RouteID
	}

	var authorID types.UserID
	if user := middleware.AuthenticatedUser(ctx); user != nil {
		authorID = user.UserID
	}

	// --- Phase 1a: Create rows with NULL pointers, collect server IDs ---
	idMap := make(map[string]string, len(req.Creates))
	createdNodes := make([]struct {
		clientID string
		serverID types.ContentID
		create   TreeNodeCreate
	}, 0, len(req.Creates))

	for _, cr := range req.Creates {
		if cr.ClientID == "" {
			resp.Errors = append(resp.Errors, "create: missing client_id")
			continue
		}

		var datatypeID types.NullableDatatypeID
		if cr.DatatypeID != "" {
			dtID := types.DatatypeID(cr.DatatypeID)
			if validateErr := dtID.Validate(); validateErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s: invalid datatype_id: %v", cr.ClientID, validateErr))
				continue
			}
			datatypeID = types.NullableDatatypeID{ID: dtID, Valid: true}
		}

		created, createErr := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
			ParentID:      nullPointer,
			FirstChildID:  nullPointer,
			NextSiblingID: nullPointer,
			PrevSiblingID: nullPointer,
			RouteID:       parentRouteID,
			DatatypeID:    datatypeID,
			AuthorID:      authorID,
			Status:        types.ContentStatusDraft,
			DateCreated:   now,
			DateModified:  now,
		})
		if createErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("tree-save: create %s failed", cr.ClientID), createErr)
			resp.Errors = append(resp.Errors, fmt.Sprintf("create %s: %v", cr.ClientID, createErr))
			continue
		}

		serverID := created.ContentDataID
		idMap[cr.ClientID] = serverID.String()
		createdNodes = append(createdNodes, struct {
			clientID string
			serverID types.ContentID
			create   TreeNodeCreate
		}{clientID: cr.ClientID, serverID: serverID, create: cr})
		resp.Created++
	}

	if len(idMap) > 0 {
		resp.IDMap = idMap
	}

	// --- Phase 1b: Update newly created rows with remapped pointers ---
	for _, cn := range createdNodes {
		parentID, parseErr := parseNullableContentID(remapPointer(cn.create.ParentID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
			continue
		}
		firstChildID, parseErr := parseNullableContentID(remapPointer(cn.create.FirstChildID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
			continue
		}
		nextSiblingID, parseErr := parseNullableContentID(remapPointer(cn.create.NextSiblingID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
			continue
		}
		prevSiblingID, parseErr := parseNullableContentID(remapPointer(cn.create.PrevSiblingID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
			continue
		}

		// Skip the update if all pointers are NULL (nothing to set).
		if !parentID.Valid && !firstChildID.Valid && !nextSiblingID.Valid && !prevSiblingID.Valid {
			continue
		}

		existing, getErr := d.GetContentData(cn.serverID)
		if getErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointer update: %v", cn.clientID, getErr))
			continue
		}

		_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: cn.serverID,
			ParentID:      parentID,
			FirstChildID:  firstChildID,
			NextSiblingID: nextSiblingID,
			PrevSiblingID: prevSiblingID,
			RouteID:       existing.RouteID,
			DatatypeID:    existing.DatatypeID,
			AuthorID:      existing.AuthorID,
			Status:        existing.Status,
			DateCreated:   existing.DateCreated,
			DateModified:  now,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("tree-save: create %s pointer update failed", cn.clientID), updateErr)
			resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, updateErr))
		}
	}

	// --- Phase 2: Deletes ---
	for _, id := range req.Deletes {
		if validateErr := id.Validate(); validateErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("invalid delete id %s: %v", id, validateErr))
			continue
		}
		if deleteErr := d.DeleteContentData(ctx, ac, id); deleteErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("tree-save: delete %s failed", id), deleteErr)
			resp.Errors = append(resp.Errors, fmt.Sprintf("delete %s: %v", id, deleteErr))
			continue
		}
		resp.Deleted++
	}

	// --- Phase 3: Updates ---
	// Fetch existing row, merge pointer fields (with ID remapping), write back.
	for _, upd := range req.Updates {
		if validateErr := upd.ContentDataID.Validate(); validateErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("invalid update id %s: %v", upd.ContentDataID, validateErr))
			continue
		}

		existing, getErr := d.GetContentData(upd.ContentDataID)
		if getErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("get %s: %v", upd.ContentDataID, getErr))
			continue
		}

		parentID, parseErr := parseNullableContentID(remapPointer(upd.ParentID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
			continue
		}
		firstChildID, parseErr := parseNullableContentID(remapPointer(upd.FirstChildID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
			continue
		}
		nextSiblingID, parseErr := parseNullableContentID(remapPointer(upd.NextSiblingID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
			continue
		}
		prevSiblingID, parseErr := parseNullableContentID(remapPointer(upd.PrevSiblingID, idMap))
		if parseErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
			continue
		}

		_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: existing.ContentDataID,
			ParentID:      parentID,
			FirstChildID:  firstChildID,
			NextSiblingID: nextSiblingID,
			PrevSiblingID: prevSiblingID,
			RouteID:       existing.RouteID,
			DatatypeID:    existing.DatatypeID,
			AuthorID:      existing.AuthorID,
			Status:        existing.Status,
			DateCreated:   existing.DateCreated,
			DateModified:  now,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("tree-save: update %s failed", upd.ContentDataID), updateErr)
			resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, updateErr))
			continue
		}
		resp.Updated++
	}

	writeJSON(w, resp)
}
