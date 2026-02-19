package modula

import "context"

// AuthResource provides authentication operations.
type AuthResource struct {
	http *httpClient
}

// Login authenticates a user with email and password.
func (a *AuthResource) Login(ctx context.Context, params LoginParams) (*LoginResponse, error) {
	var result LoginResponse
	if err := a.http.post(ctx, "/api/v1/auth/login", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Logout ends the current session.
func (a *AuthResource) Logout(ctx context.Context) error {
	return a.http.post(ctx, "/api/v1/auth/logout", nil, nil)
}

// Me returns the currently authenticated user.
func (a *AuthResource) Me(ctx context.Context) (*User, error) {
	var result User
	if err := a.http.get(ctx, "/api/v1/auth/me", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Register creates a new user account.
func (a *AuthResource) Register(ctx context.Context, params CreateUserParams) (*User, error) {
	var result User
	if err := a.http.post(ctx, "/api/v1/auth/register", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ResetPassword initiates a password reset.
func (a *AuthResource) ResetPassword(ctx context.Context, params ResetPasswordParams) error {
	return a.http.post(ctx, "/api/v1/auth/reset", params, nil)
}
