package modula

import (
	"context"
)

// ContentTreeResource provides bulk content tree operations.
//
// The tree save endpoint (POST /api/v1/content/tree) atomically applies
// creates, deletes, and pointer-field updates to content_data nodes in a
// single HTTP round-trip. This is the preferred way to persist structural
// changes from a block editor or tree manipulation UI.
//
// # Creates
//
// New nodes are inserted with server-generated ULIDs. The caller supplies a
// client-side ID (e.g. a UUID from crypto.randomUUID) and receives an IDMap
// in the response mapping each client ID to the server-generated ULID.
// Pointer fields (ParentID, FirstChildID, etc.) may reference other new
// nodes by their client IDs — the server remaps them automatically.
//
// The server inherits RouteID from the parent content node, sets AuthorID
// from the authenticated user, and defaults Status to "draft".
//
// # Deletes
//
// Nodes listed in Deletes are removed via audited delete. Deletes are
// processed before updates so that removed nodes don't interfere with
// pointer rewiring.
//
// # Updates
//
// Each update entry specifies a content_data_id and the new values for its
// four pointer fields (parent, first_child, next_sibling, prev_sibling).
// The server fetches the existing row and preserves all non-pointer fields
// (route, datatype, author, status, dates), only overwriting the pointers
// and bumping DateModified.
//
// # Partial failures
//
// The endpoint always returns HTTP 200. Check TreeSaveResponse.Errors for
// per-node error messages. Created/Updated/Deleted counts reflect only
// successful operations.
type ContentTreeResource struct {
	http *httpClient
}

// TreeNodeCreate describes a new content_data node to insert.
// ClientID is a caller-generated identifier (e.g. UUID); the server generates
// a ULID and returns the mapping in TreeSaveResponse.IDMap.
type TreeNodeCreate struct {
	// ClientID is the caller-generated temporary ID for this node.
	// Must be unique within the request. Other nodes may reference this
	// ID in their pointer fields; the server remaps them to the real ULID.
	ClientID string `json:"client_id"`

	// DatatypeID is the datatype to assign to the new node.
	// Pass empty string for no datatype.
	DatatypeID string `json:"datatype_id"`

	// ParentID is the parent content node, or nil for a root-level node.
	ParentID *string `json:"parent_id"`

	// FirstChildID is the first child of this node, or nil.
	FirstChildID *string `json:"first_child_id"`

	// NextSiblingID is the next sibling in the linked list, or nil.
	NextSiblingID *string `json:"next_sibling_id"`

	// PrevSiblingID is the previous sibling in the linked list, or nil.
	PrevSiblingID *string `json:"prev_sibling_id"`
}

// TreeNodeUpdate describes pointer-field changes for an existing content_data node.
// Only the four tree-pointer fields are updated; all other fields (route, datatype,
// author, status, dates) are preserved from the existing row.
type TreeNodeUpdate struct {
	// ContentDataID is the ULID of the existing node to update.
	ContentDataID ContentID `json:"content_data_id"`

	// ParentID is the new parent, or nil for SQL NULL.
	ParentID *string `json:"parent_id"`

	// FirstChildID is the new first child, or nil for SQL NULL.
	FirstChildID *string `json:"first_child_id"`

	// NextSiblingID is the new next sibling, or nil for SQL NULL.
	NextSiblingID *string `json:"next_sibling_id"`

	// PrevSiblingID is the new previous sibling, or nil for SQL NULL.
	PrevSiblingID *string `json:"prev_sibling_id"`
}

// TreeSaveRequest is the request body for POST /api/v1/content/tree.
type TreeSaveRequest struct {
	// ContentID is the root content node being edited. Used to resolve
	// RouteID for newly created child nodes.
	ContentID ContentID `json:"content_id"`

	// Creates are new content_data nodes to insert. Processed first.
	// Each node's pointer fields may reference other new nodes by ClientID.
	Creates []TreeNodeCreate `json:"creates,omitempty"`

	// Updates are existing nodes whose pointer fields changed.
	// Pointer fields may reference new nodes by ClientID — the server
	// remaps them using the IDMap built during the creates phase.
	Updates []TreeNodeUpdate `json:"updates,omitempty"`

	// Deletes are content_data IDs to remove. Processed after creates
	// but before updates.
	Deletes []ContentID `json:"deletes,omitempty"`
}

// TreeSaveResponse is the response from POST /api/v1/content/tree.
type TreeSaveResponse struct {
	// Created is the number of nodes successfully created.
	Created int `json:"created"`

	// Updated is the number of nodes successfully updated.
	Updated int `json:"updated"`

	// Deleted is the number of nodes successfully deleted.
	Deleted int `json:"deleted"`

	// IDMap maps client-supplied IDs to server-generated ULIDs.
	// Only present when Creates were included in the request.
	// Use this to remap local state after a successful save.
	IDMap map[string]string `json:"id_map,omitempty"`

	// Errors contains per-node error messages for partial failures.
	// Empty when all operations succeeded.
	Errors []string `json:"errors,omitempty"`
}

// Save atomically applies tree structure changes (creates, deletes, and
// pointer updates) in a single request.
//
// This is the primary method for persisting block editor state. It replaces
// the pattern of sending N individual POST requests to /api/v1/content/batch.
//
// Example:
//
//	resp, err := client.ContentTree.Save(ctx, modula.TreeSaveRequest{
//	    ContentID: parentID,
//	    Creates: []modula.TreeNodeCreate{{
//	        ClientID:   "temp-uuid-1",
//	        DatatypeID: string(dtID),
//	        ParentID:   modula.StringPtr(string(parentID)),
//	    }},
//	    Updates: []modula.TreeNodeUpdate{{
//	        ContentDataID: existingID,
//	        FirstChildID:  modula.StringPtr("temp-uuid-1"),
//	    }},
//	    Deletes: []modula.ContentID{removedID},
//	})
func (t *ContentTreeResource) Save(ctx context.Context, req TreeSaveRequest) (*TreeSaveResponse, error) {
	var resp TreeSaveResponse
	if err := t.http.post(ctx, "/api/v1/content/tree", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StringPtr returns a pointer to s. Convenience helper for building
// TreeNodeCreate and TreeNodeUpdate pointer fields.
func StringPtr(s string) *string {
	return &s
}
