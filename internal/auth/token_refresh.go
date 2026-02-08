package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"golang.org/x/oauth2"
)

// TokenRefresher handles automatic refreshing of OAuth access tokens before expiration.
// It checks token expiration times and refreshes them using the OAuth provider's
// token endpoint when they are close to expiring.
type TokenRefresher struct {
	config *config.Config
	driver db.DbDriver
	log    Logger
}

// NewTokenRefresher creates a new TokenRefresher with the given configuration and logger.
func NewTokenRefresher(log Logger, c *config.Config) *TokenRefresher {
	return &TokenRefresher{
		config: c,
		driver: db.ConfigDB(*c),
		log:    log,
	}
}

// RefreshIfNeeded checks if a user's OAuth token needs refreshing and refreshes it if necessary.
// Tokens are refreshed if they expire within 5 minutes.
// Returns nil if refresh was successful or not needed.
// Returns an error if the user doesn't have OAuth configured or if refresh fails.
func (tr *TokenRefresher) RefreshIfNeeded(userID types.UserID) error {
	// Get user's OAuth record
	userOauth, err := tr.driver.GetUserOauthByUserId(types.NullableUserID{Valid: true, ID: userID})
	if err != nil {
		// User doesn't use OAuth - this is not an error
		tr.log.Debug("No OAuth record for user %d", userID)
		return nil
	}

	if userOauth == nil {
		return nil // User doesn't use OAuth
	}

	// Check if token has no expiry (GitHub tokens don't expire)
	if userOauth.TokenExpiresAt == "" {
		tr.log.Debug("Token for user %d has no expiry (long-lived token)", userID)
		return nil // No refresh needed for long-lived tokens
	}

	// Parse expiration time
	expiresAt, err := time.Parse(time.RFC3339, userOauth.TokenExpiresAt)
	if err != nil {
		tr.log.Warn("Failed to parse token expiration for user", err, userID)
		// Try parsing as alternative formats
		expiresAt, err = time.Parse("2006-01-02 15:04:05", userOauth.TokenExpiresAt)
		if err != nil {
			return fmt.Errorf("failed to parse expiration: %w", err)
		}
	}

	// Double-check if token has no expiry
	if expiresAt.IsZero() || expiresAt.Year() == 1 {
		tr.log.Debug("Token for user %d has no expiry (long-lived token)", userID)
		return nil // No refresh needed for long-lived tokens
	}

	// Check if token expires within 5 minutes
	if time.Until(expiresAt) > 5*time.Minute {
		tr.log.Debug("Token for user %d still valid for %s", userID, time.Until(expiresAt))
		return nil // Token still valid
	}

	tr.log.Info("Token for user %d expiring soon, refreshing...", userID)

	// Refresh the token
	newToken, err := tr.refreshToken(userOauth)
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Update database with new tokens
	err = tr.updateTokens(userOauth.UserOauthID, newToken)
	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	tr.log.Info("Token refreshed successfully for user %d", userID)
	return nil
}

// refreshToken performs the actual OAuth token refresh using the refresh token.
func (tr *TokenRefresher) refreshToken(userOauth *db.UserOauth) (*oauth2.Token, error) {
	// Build OAuth config
	conf := &oauth2.Config{
		ClientID:     tr.config.Oauth_Client_Id,
		ClientSecret: tr.config.Oauth_Client_Secret,
		Endpoint: oauth2.Endpoint{
			TokenURL: tr.config.Oauth_Endpoint[config.OauthTokenURL],
		},
	}

	// Create token source from refresh token
	token := &oauth2.Token{
		RefreshToken: userOauth.RefreshToken,
		AccessToken:  userOauth.AccessToken,
	}

	// Refresh the token
	ctx := context.Background()
	tokenSource := conf.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		tr.log.Error("Token refresh API call failed", err)
		return nil, err
	}

	return newToken, nil
}

// updateTokens updates the user_oauth record with new token information.
func (tr *TokenRefresher) updateTokens(userOauthID types.UserOauthID, token *oauth2.Token) error {
	// Handle tokens without expiry (GitHub)
	expiresAt := ""
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Format(time.RFC3339)
	}

	// Update in database
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(tr.config.Node_ID), types.UserID(""), "token-refresh", "system")
	_, err := tr.driver.UpdateUserOauth(ctx, ac, db.UpdateUserOauthParams{
		UserOauthID:    userOauthID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		TokenExpiresAt: expiresAt,
	})

	if err != nil {
		tr.log.Error("Failed to update tokens in database", err)
		return err
	}

	tr.log.Debug("Updated tokens for user_oauth_id %s, new expiry: %s", userOauthID, expiresAt)
	return nil
}
