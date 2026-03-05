package modula

import (
	"context"
	"fmt"
)

// UserCompositeResource provides composite user operations such as
// reassigning a user's owned content to another user and then deleting
// the original user.
type UserCompositeResource struct {
	http *httpClient
}

// ReassignDelete reassigns all content owned by the specified user to
// another user (or the system default) and then deletes the original user.
func (u *UserCompositeResource) ReassignDelete(ctx context.Context, params UserReassignDeleteParams) (*UserReassignDeleteResponse, error) {
	var result UserReassignDeleteResponse
	if err := u.http.post(ctx, "/api/v1/users/reassign-delete", params, &result); err != nil {
		return nil, fmt.Errorf("user reassign-delete: %w", err)
	}
	return &result, nil
}
