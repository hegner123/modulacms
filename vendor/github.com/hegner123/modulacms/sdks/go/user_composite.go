package modula

import (
	"context"
	"fmt"
)

// UserCompositeResource provides composite user operations that span multiple
// tables atomically. Use this instead of separate delete operations when a user
// owns content that must be preserved.
// It is accessed via [Client].UserComposite.
type UserCompositeResource struct {
	http *httpClient
}

// ReassignDelete atomically reassigns all content owned by the specified user to
// another user (identified in params) and then deletes the original user.
// This is the safe way to remove a user who has authored content, as it prevents
// orphaned content records. The operation is transactional: if reassignment fails,
// the user is not deleted.
// Returns a [*UserReassignDeleteResponse] summarizing how many items were reassigned.
func (u *UserCompositeResource) ReassignDelete(ctx context.Context, params UserReassignDeleteParams) (*UserReassignDeleteResponse, error) {
	var result UserReassignDeleteResponse
	if err := u.http.post(ctx, "/api/v1/users/reassign-delete", params, &result); err != nil {
		return nil, fmt.Errorf("user reassign-delete: %w", err)
	}
	return &result, nil
}
