package modula

import (
	"context"
	"encoding/json"
	"fmt"
)

// ContentReorderRequest is the JSON body for POST /api/v1/contentdata/reorder.
//
// It atomically rewrites the sibling-pointer linked list for all children of
// ParentID so they appear in the order specified by OrderedIDs. The server
// updates each node's next_sibling_id and prev_sibling_id pointers, and sets
// the parent's first_child_id to OrderedIDs[0].
type ContentReorderRequest struct {
	// ParentID is the parent node whose children are being reordered.
	// Pass nil to reorder root-level (parentless) nodes.
	ParentID *ContentID `json:"parent_id"`

	// OrderedIDs is the desired display order of the sibling nodes.
	// All IDs must be existing children of ParentID; the server returns
	// an error if any ID is missing or belongs to a different parent.
	OrderedIDs []ContentID `json:"ordered_ids"`
}

// ContentReorderResponse is the JSON response from POST /api/v1/contentdata/reorder.
type ContentReorderResponse struct {
	// Updated is the number of sibling-pointer fields that were modified.
	Updated int `json:"updated"`

	// ParentID echoes back the parent node whose children were reordered.
	// Nil when root-level nodes were reordered.
	ParentID *ContentID `json:"parent_id"`
}

// ContentMoveRequest is the JSON body for POST /api/v1/contentdata/move.
//
// It detaches a node from its current position in the sibling linked list,
// re-links the surrounding siblings, and inserts it under NewParentID at the
// given Position. All sibling pointers (next_sibling_id, prev_sibling_id) and
// the parent's first_child_id are updated atomically.
type ContentMoveRequest struct {
	// NodeID is the content node to relocate.
	NodeID ContentID `json:"node_id"`

	// NewParentID is the destination parent. Pass nil to move the node to
	// the root level (no parent).
	NewParentID *ContentID `json:"new_parent_id"`

	// Position is the zero-based index among the new parent's children
	// where the node should be inserted. Use 0 to make it the first child.
	Position int `json:"position"`
}

// ContentMoveResponse is the JSON response from POST /api/v1/contentdata/move.
type ContentMoveResponse struct {
	// NodeID echoes the moved node's ID.
	NodeID ContentID `json:"node_id"`

	// OldParentID is the parent the node was detached from. Nil if the node
	// was previously at the root level.
	OldParentID *ContentID `json:"old_parent_id"`

	// NewParentID is the parent the node was attached to. Nil if the node
	// was moved to the root level.
	NewParentID *ContentID `json:"new_parent_id"`

	// Position is the final zero-based index of the node among its new siblings.
	Position int `json:"position"`
}

// ContentReorderResource provides content tree reorder and move operations for
// public content data nodes. These operations manipulate the sibling-pointer
// linked list (next_sibling_id, prev_sibling_id) and parent-child pointers
// (parent_id, first_child_id) that define the content tree's ordering.
//
// Use [ContentReorderResource.Reorder] to change the display order of siblings
// under a common parent, and [ContentReorderResource.Move] to relocate a node
// to a different parent and position.
//
// Access this resource via [Client].ContentReorder:
//
//	resp, err := client.ContentReorder.Reorder(ctx, req)
//
// For admin content data nodes, use [AdminContentReorderResource] via
// [Client].AdminContentReorder instead.
type ContentReorderResource struct {
	http *httpClient
}

// Reorder atomically rewrites the sibling-pointer linked list so that the
// children of the specified parent appear in the order given by
// [ContentReorderRequest].OrderedIDs. All pointer updates happen in a single
// database transaction.
//
// Returns an [*ApiError] if any of the ordered IDs do not belong to the
// specified parent, or if the authenticated user lacks the content:update
// permission.
func (r *ContentReorderResource) Reorder(ctx context.Context, req ContentReorderRequest) (*ContentReorderResponse, error) {
	var result ContentReorderResponse
	if err := r.http.post(ctx, "/api/v1/contentdata/reorder", req, &result); err != nil {
		return nil, fmt.Errorf("reorder content: %w", err)
	}
	return &result, nil
}

// Move detaches a content data node from its current position in the tree and
// inserts it under a new parent at the specified zero-based position. The
// server atomically updates sibling pointers in both the old and new locations.
//
// Moving a node to its own current position is a no-op that still returns a
// successful response.
//
// Returns an [*ApiError] if the node does not exist or the authenticated user
// lacks the content:update permission.
func (r *ContentReorderResource) Move(ctx context.Context, req ContentMoveRequest) (*ContentMoveResponse, error) {
	var result ContentMoveResponse
	if err := r.http.post(ctx, "/api/v1/contentdata/move", req, &result); err != nil {
		return nil, fmt.Errorf("move content: %w", err)
	}
	return &result, nil
}

// AdminContentReorderRequest is the JSON body for POST /api/v1/admincontentdatas/reorder.
//
// It is the admin-content equivalent of [ContentReorderRequest], operating on
// admin content data nodes rather than public content data nodes.
type AdminContentReorderRequest struct {
	// ParentID is the admin content parent whose children are being reordered.
	// Pass nil to reorder root-level admin content nodes.
	ParentID *AdminContentID `json:"parent_id"`

	// OrderedIDs is the desired display order of the admin content sibling nodes.
	OrderedIDs []AdminContentID `json:"ordered_ids"`
}

// AdminContentReorderResponse is the JSON response from POST /api/v1/admincontentdatas/reorder.
type AdminContentReorderResponse struct {
	// Updated is the number of sibling-pointer fields that were modified.
	Updated int `json:"updated"`

	// ParentID echoes back the parent node whose children were reordered.
	ParentID *AdminContentID `json:"parent_id"`
}

// AdminContentMoveRequest is the JSON body for POST /api/v1/admincontentdatas/move.
//
// It is the admin-content equivalent of [ContentMoveRequest], operating on
// admin content data nodes rather than public content data nodes.
type AdminContentMoveRequest struct {
	// NodeID is the admin content node to relocate.
	NodeID AdminContentID `json:"node_id"`

	// NewParentID is the destination parent. Pass nil to move the node to
	// the root level.
	NewParentID *AdminContentID `json:"new_parent_id"`

	// Position is the zero-based index among the new parent's children
	// where the node should be inserted.
	Position int `json:"position"`
}

// AdminContentMoveResponse is the JSON response from POST /api/v1/admincontentdatas/move.
type AdminContentMoveResponse struct {
	// NodeID echoes the moved admin content node's ID.
	NodeID AdminContentID `json:"node_id"`

	// OldParentID is the parent the node was detached from. Nil if the node
	// was previously at the root level.
	OldParentID *AdminContentID `json:"old_parent_id"`

	// NewParentID is the parent the node was attached to. Nil if the node
	// was moved to the root level.
	NewParentID *AdminContentID `json:"new_parent_id"`

	// Position is the final zero-based index of the node among its new siblings.
	Position int `json:"position"`
}

// AdminContentReorderResource provides reorder and move operations for admin
// content data nodes. It is the admin-content counterpart of
// [ContentReorderResource] and operates on the admin content tree rather than
// the public content tree.
//
// Access this resource via [Client].AdminContentReorder:
//
//	resp, err := client.AdminContentReorder.Reorder(ctx, req)
type AdminContentReorderResource struct {
	http *httpClient
}

// Reorder atomically rewrites the sibling-pointer linked list so that the
// admin content children of the specified parent appear in the order given by
// [AdminContentReorderRequest].OrderedIDs.
//
// The response is decoded through [json.RawMessage] because the server uses
// NullableContentID, which serializes differently from the SDK's typed IDs.
//
// Returns an [*ApiError] if the authenticated user lacks the required
// admin permission.
func (r *AdminContentReorderResource) Reorder(ctx context.Context, req AdminContentReorderRequest) (*AdminContentReorderResponse, error) {
	// The response comes back as raw JSON since the server uses NullableContentID
	// which serializes differently from our SDK types. Decode via json.RawMessage first.
	var raw json.RawMessage
	if err := r.http.post(ctx, "/api/v1/admincontentdatas/reorder", req, &raw); err != nil {
		return nil, fmt.Errorf("reorder admin content: %w", err)
	}
	var result AdminContentReorderResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode reorder response: %w", err)
	}
	return &result, nil
}

// Move detaches an admin content data node from its current position and
// inserts it under a new parent at the specified zero-based position. The
// server atomically updates sibling pointers in both the source and
// destination locations.
//
// Returns an [*ApiError] if the node does not exist or the authenticated user
// lacks the required admin permission.
func (r *AdminContentReorderResource) Move(ctx context.Context, req AdminContentMoveRequest) (*AdminContentMoveResponse, error) {
	var raw json.RawMessage
	if err := r.http.post(ctx, "/api/v1/admincontentdatas/move", req, &raw); err != nil {
		return nil, fmt.Errorf("move admin content: %w", err)
	}
	var result AdminContentMoveResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode move response: %w", err)
	}
	return &result, nil
}
