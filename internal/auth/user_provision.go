package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/oauth2"
)

// UserInfo represents the standardized user information retrieved from OAuth providers.
// This struct is provider-agnostic and maps common fields from various OAuth providers.
type UserInfo struct {
	ProviderUserID string `json:"sub"`                // OpenID Connect standard
	ID             int64  `json:"id"`                 // GitHub user ID (numeric)
	Email          string `json:"email"`              // User email address
	Name           string `json:"name"`               // Full name
	Username       string `json:"preferred_username"` // Preferred username
	Login          string `json:"login"`              // GitHub-specific username
	AvatarURL      string `json:"avatar_url"`         // Profile picture URL
}

// UserProvisioner handles user creation and OAuth account linking.
// It provides a unified interface for provisioning users from various OAuth providers.
type UserProvisioner struct {
	config *config.Config
	driver db.DbDriver
}

// NewUserProvisioner creates a new UserProvisioner with the given configuration.
func NewUserProvisioner(c *config.Config) *UserProvisioner {
	return &UserProvisioner{
		config: c,
		driver: db.ConfigDB(*c),
	}
}

// FetchUserInfo retrieves user information from the OAuth provider's userinfo endpoint.
// It uses the authenticated HTTP client to make the request and returns standardized UserInfo.
func (up *UserProvisioner) FetchUserInfo(client *http.Client) (*UserInfo, error) {
	userInfoURL := up.config.Oauth_Endpoint[config.OauthUserInfoURL]
	if userInfoURL == "" {
		return nil, fmt.Errorf("oauth_userinfo_url not configured")
	}

	utility.DefaultLogger.Info("Fetching user info from: %s", userInfoURL)

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("userinfo request failed: %s - %s", resp.Status, body)
		utility.DefaultLogger.Error("Userinfo request failed", err)
		return nil, err
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %w", err)
	}

	// Handle provider-specific fields
	// GitHub uses "login" instead of "preferred_username"
	if userInfo.Username == "" && userInfo.Login != "" {
		userInfo.Username = userInfo.Login
	}

	// GitHub uses "id" (number) instead of "sub" (string)
	// Convert GitHub's numeric ID to string for ProviderUserID
	if userInfo.ProviderUserID == "" && userInfo.ID != 0 {
		userInfo.ProviderUserID = fmt.Sprintf("%d", userInfo.ID)
	}

	utility.DefaultLogger.Info("Fetched user info - Email: %s, Username: %s, ProviderID: %s",
		userInfo.Email, userInfo.Username, userInfo.ProviderUserID)

	// If email is missing, try to fetch from GitHub's /user/emails endpoint
	if userInfo.Email == "" {
		utility.DefaultLogger.Info("Email not in user info, fetching from /user/emails endpoint")
		email, err := up.fetchGitHubEmail(client)
		if err != nil {
			utility.DefaultLogger.Warn("Failed to fetch email from /user/emails", err)
		} else {
			userInfo.Email = email
		}
	}

	// Validate required fields
	if userInfo.Email == "" {
		return nil, fmt.Errorf("email not provided by OAuth provider - you may need to request 'user:email' scope or make your email public on GitHub")
	}

	return &userInfo, nil
}

// GitHubEmail represents an email from GitHub's /user/emails endpoint
type GitHubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// fetchGitHubEmail fetches the primary email from GitHub's /user/emails endpoint
func (up *UserProvisioner) fetchGitHubEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", fmt.Errorf("failed to fetch emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("emails request failed: %s - %s", resp.Status, body)
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode emails: %w", err)
	}

	// Find primary verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			utility.DefaultLogger.Info("Found primary verified email: %s", email.Email)
			return email.Email, nil
		}
	}

	// Fallback to first verified email
	for _, email := range emails {
		if email.Verified {
			utility.DefaultLogger.Info("Using first verified email: %s", email.Email)
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}

// ProvisionUser creates a new user or links OAuth to an existing user.
// It follows this logic:
// 1. Check if OAuth provider+userID already exists → update tokens, return user
// 2. Check if email already exists → link OAuth, return user
// 3. Create new user → link OAuth, return user
func (up *UserProvisioner) ProvisionUser(
	userInfo *UserInfo,
	token *oauth2.Token,
	provider string,
) (*db.Users, error) {
	utility.DefaultLogger.Info("Starting user provisioning for provider: %s", provider)

	// Validate required fields
	if userInfo.Email == "" {
		err := fmt.Errorf("email is required from OAuth provider")
		utility.DefaultLogger.Error("Provisioning failed: no email", err)
		return nil, err
	}

	// Use sub as provider user ID, or fall back to email
	providerUserID := userInfo.ProviderUserID
	if providerUserID == "" {
		providerUserID = userInfo.Email
		utility.DefaultLogger.Warn("OAuth provider didn't provide 'sub', using email as provider ID", nil)
	}

	utility.DefaultLogger.Info("Provisioning user - Email: %s, ProviderUserID: %s", userInfo.Email, providerUserID)

	// Check if user already linked via OAuth
	existingOauth, err := up.driver.GetUserOauthByProviderID(provider, providerUserID)
	if err == nil && existingOauth != nil {
		utility.DefaultLogger.Info("Existing OAuth link found for %s:%s", provider, providerUserID)

		// User exists, update tokens
		err = up.updateTokens(existingOauth.UserOauthID, token)
		if err != nil {
			utility.DefaultLogger.Warn("Failed to update tokens: %v", err)
		}

		// Return existing user
		return up.driver.GetUser(existingOauth.UserID.ID)
	}

	// Check if user exists by email
	existingUser, err := up.driver.GetUserByEmail(types.Email(userInfo.Email))
	if err == nil && existingUser != nil {
		utility.DefaultLogger.Info("Found existing user by email: %s", userInfo.Email)
		// Link OAuth to existing user
		return up.linkOAuthToUser(existingUser, userInfo, token, provider, providerUserID)
	}

	// Create new user
	utility.DefaultLogger.Info("Creating new user for: %s", userInfo.Email)
	return up.createNewUser(userInfo, token, provider, providerUserID)
}

// createNewUser creates a new user account with OAuth linking.
func (up *UserProvisioner) createNewUser(
	userInfo *UserInfo,
	token *oauth2.Token,
	provider string,
	providerUserID string,
) (*db.Users, error) {
	// Generate username if not provided
	username := userInfo.Username
	if username == "" {
		username = userInfo.Email
	}

	// Set name, default to email if not provided
	name := userInfo.Name
	if name == "" {
		name = username
	}

	// Look up the viewer role by label
	viewerRoleID := ""
	roles, err := up.driver.ListRoles()
	if err == nil && roles != nil {
		for _, r := range *roles {
			if r.Label == "viewer" {
				viewerRoleID = r.RoleID.String()
				break
			}
		}
	}
	if viewerRoleID == "" {
		return nil, fmt.Errorf("failed to find viewer role")
	}

	// Create user (no password for OAuth users)
	now := types.TimestampNow()
	user, err := up.driver.CreateUser(db.CreateUserParams{
		Username:     username,
		Name:         name,
		Email:        types.Email(userInfo.Email),
		Hash:         "", // OAuth users don't have passwords
		Role:         viewerRoleID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		utility.DefaultLogger.Error("Failed to create user", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Link OAuth
	// Handle tokens without expiry (GitHub)
	expiresAt := ""
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Format(time.RFC3339)
	}
	_, err = up.driver.CreateUserOauth(db.CreateUserOauthParams{
		UserID:              types.NullableUserID{ID: user.UserID, Valid: true},
		OauthProvider:       provider,
		OauthProviderUserID: providerUserID,
		AccessToken:         token.AccessToken,
		RefreshToken:        token.RefreshToken,
		TokenExpiresAt:      expiresAt,
		DateCreated:         types.TimestampNow(),
	})
	if err != nil {
		utility.DefaultLogger.Error("Failed to link OAuth", err)
		return nil, fmt.Errorf("failed to link OAuth: %w", err)
	}

	utility.DefaultLogger.Info("Created new user via OAuth: %s (user_id: %d)", userInfo.Email, user.UserID)
	return user, nil
}

// linkOAuthToUser links OAuth credentials to an existing user account.
func (up *UserProvisioner) linkOAuthToUser(
	user *db.Users,
	userInfo *UserInfo,
	token *oauth2.Token,
	provider string,
	providerUserID string,
) (*db.Users, error) {
	// Handle tokens without expiry (GitHub)
	expiresAt := ""
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Format(time.RFC3339)
	}

	_, err := up.driver.CreateUserOauth(db.CreateUserOauthParams{
		UserID:              types.NullableUserID{ID: user.UserID, Valid: true},
		OauthProvider:       provider,
		OauthProviderUserID: providerUserID,
		AccessToken:         token.AccessToken,
		RefreshToken:        token.RefreshToken,
		TokenExpiresAt:      expiresAt,
		DateCreated:         types.TimestampNow(),
	})
	if err != nil {
		utility.DefaultLogger.Error("Failed to link OAuth to existing user", err)
		return nil, fmt.Errorf("failed to link OAuth: %w", err)
	}

	utility.DefaultLogger.Info("Linked OAuth to existing user: %s (user_id: %d)", userInfo.Email, user.UserID)
	return user, nil
}

// updateTokens updates the OAuth tokens for an existing user_oauth record.
func (up *UserProvisioner) updateTokens(userOauthID types.UserOauthID, token *oauth2.Token) error {
	// Handle tokens without expiry (GitHub)
	expiresAt := ""
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Format(time.RFC3339)
	}

	_, err := up.driver.UpdateUserOauth(db.UpdateUserOauthParams{
		UserOauthID:    userOauthID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		TokenExpiresAt: expiresAt,
	})

	if err != nil {
		return err
	}

	utility.DefaultLogger.Debug("Updated OAuth tokens for user_oauth_id: %s", userOauthID)
	return nil
}
