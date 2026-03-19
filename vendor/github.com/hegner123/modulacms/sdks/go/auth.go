package modula

import "context"

// AuthResource provides authentication operations including login, logout,
// registration, and password reset flows. It is accessed via [Client].Auth.
//
// All methods except [AuthResource.Register] and [AuthResource.RequestPasswordReset]
// require an authenticated session or API key.
type AuthResource struct {
	http *httpClient
}

// Login authenticates a user with email and password, returning a session token
// on success. The returned [LoginResponse] contains both the authenticated [User]
// and a session token that can be used for subsequent requests.
// Returns an [*ApiError] with status 401 if the credentials are invalid.
func (a *AuthResource) Login(ctx context.Context, params LoginParams) (*LoginResponse, error) {
	var result LoginResponse
	if err := a.http.post(ctx, "/api/v1/auth/login", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Logout ends the current authenticated session, invalidating the session token.
// After logout, the API key or a new login is required for further requests.
func (a *AuthResource) Logout(ctx context.Context) error {
	return a.http.post(ctx, "/api/v1/auth/logout", nil, nil)
}

// Me returns the currently authenticated user's profile.
// Returns an [*ApiError] with status 401 if no valid session or API key is present.
func (a *AuthResource) Me(ctx context.Context) (*User, error) {
	var result User
	if err := a.http.get(ctx, "/api/v1/auth/me", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Register creates a new user account with the given parameters.
// The new user is assigned the default "viewer" role. Only admins can
// assign elevated roles after registration via the Users resource.
// Returns the created [User] on success.
func (a *AuthResource) Register(ctx context.Context, params CreateUserParams) (*User, error) {
	var result User
	if err := a.http.post(ctx, "/api/v1/auth/register", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RequestPasswordReset sends a password reset email to the given address.
// Always succeeds (returns 200) regardless of whether the email exists,
// to prevent user enumeration.
func (a *AuthResource) RequestPasswordReset(ctx context.Context, params RequestPasswordResetParams) (*MessageResponse, error) {
	var result MessageResponse
	if err := a.http.post(ctx, "/api/v1/auth/request-password-reset", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfirmPasswordReset validates a reset token and sets the new password.
// The token must be one previously issued by [AuthResource.RequestPasswordReset].
// Returns an [*ApiError] with status 400 if the token is invalid or expired.
func (a *AuthResource) ConfirmPasswordReset(ctx context.Context, params ConfirmPasswordResetParams) (*MessageResponse, error) {
	var result MessageResponse
	if err := a.http.post(ctx, "/api/v1/auth/confirm-password-reset", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
