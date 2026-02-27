package modula

import (
	"context"
	"encoding/json"
	"fmt"
)

// ContentReorderRequest is the JSON body for POST /api/v1/contentdata/reorder.
type ContentReorderRequest struct {
	ParentID   *ContentID  `json:"parent_id"`
	OrderedIDs []ContentID `json:"ordered_ids"`
}

// ContentReorderResponse is the JSON response from POST /api/v1/contentdata/reorder.
type ContentReorderResponse struct {
	Updated  int        `json:"updated"`
	ParentID *ContentID `json:"parent_id"`
}

// ContentMoveRequest is the JSON body for POST /api/v1/contentdata/move.
type ContentMoveRequest struct {
	NodeID      ContentID  `json:"node_id"`
	NewParentID *ContentID `json:"new_parent_id"`
	Position    int        `json:"position"`
}

// ContentMoveResponse is the JSON response from POST /api/v1/contentdata/move.
type ContentMoveResponse struct {
	NodeID      ContentID  `json:"node_id"`
	OldParentID *ContentID `json:"old_parent_id"`
	NewParentID *ContentID `json:"new_parent_id"`
	Position    int        `json:"position"`
}

// ContentReorderResource provides content data reorder and move operations.
type ContentReorderResource struct {
	http *httpClient
}

// Reorder atomically reorders sibling content data nodes under a parent.
func (r *ContentReorderResource) Reorder(ctx context.Context, req ContentReorderRequest) (*ContentReorderResponse, error) {
	var result ContentReorderResponse
	if err := r.http.post(ctx, "/api/v1/contentdata/reorder", req, &result); err != nil {
		return nil, fmt.Errorf("reorder content: %w", err)
	}
	return &result, nil
}

// Move moves a content data node to a new parent at a given position.
func (r *ContentReorderResource) Move(ctx context.Context, req ContentMoveRequest) (*ContentMoveResponse, error) {
	var result ContentMoveResponse
	if err := r.http.post(ctx, "/api/v1/contentdata/move", req, &result); err != nil {
		return nil, fmt.Errorf("move content: %w", err)
	}
	return &result, nil
}

// AdminContentReorderRequest is the JSON body for POST /api/v1/admincontentdatas/reorder.
type AdminContentReorderRequest struct {
	ParentID   *AdminContentID  `json:"parent_id"`
	OrderedIDs []AdminContentID `json:"ordered_ids"`
}

// AdminContentReorderResponse is the JSON response from POST /api/v1/admincontentdatas/reorder.
type AdminContentReorderResponse struct {
	Updated  int             `json:"updated"`
	ParentID *AdminContentID `json:"parent_id"`
}

// AdminContentMoveRequest is the JSON body for POST /api/v1/admincontentdatas/move.
type AdminContentMoveRequest struct {
	NodeID      AdminContentID  `json:"node_id"`
	NewParentID *AdminContentID `json:"new_parent_id"`
	Position    int             `json:"position"`
}

// AdminContentMoveResponse is the JSON response from POST /api/v1/admincontentdatas/move.
type AdminContentMoveResponse struct {
	NodeID      AdminContentID  `json:"node_id"`
	OldParentID *AdminContentID `json:"old_parent_id"`
	NewParentID *AdminContentID `json:"new_parent_id"`
	Position    int             `json:"position"`
}

// AdminContentReorderResource provides admin content data reorder and move operations.
type AdminContentReorderResource struct {
	http *httpClient
}

// Reorder atomically reorders sibling admin content data nodes under a parent.
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

// Move moves an admin content data node to a new parent at a given position.
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
