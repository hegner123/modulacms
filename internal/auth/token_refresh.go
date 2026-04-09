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

// NewTokenRefresher creates a new TokenRefresher with the given configuration, logger, and database driver.
func NewTokenRefresher(log Logger, c *config.Config, driver db.DbDriver) *TokenRefresher {
	return &TokenRefresher{
		config: c,
		driver: driver,
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
		tr.log.Debug("no OAuth record for user %s", userID)
		return nil
	}

	if userOauth == nil {
		return nil // User doesn't use OAuth
	}

	// Check if token has no expiry (GitHub tokens don't expire)
	if !userOauth.TokenExpiresAt.Valid || userOauth.TokenExpiresAt.IsZero() {
		tr.log.Debug("Token for user %s has no expiry (long-lived token)", userID)
		return nil // No refresh needed for long-lived tokens
	}

	expiresAt := userOauth.TokenExpiresAt.Time

	// Check if token expires within 5 minutes
	if time.Until(expiresAt) > 5*time.Minute {
		tr.log.Debug("Token for user %s still valid for %s", userID, time.Until(expiresAt))
		return nil // Token still valid
	}

	tr.log.Info("Token for user %s expiring soon, refreshing...", userID)

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

	tr.log.Info("Token refreshed successfully for user %s", userID)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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
	var expiresAt types.Timestamp
	if !token.Expiry.IsZero() {
		expiresAt = types.NewTimestamp(token.Expiry)
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
		tr.log.Error("failed to update tokens in database", err)
		return err
	}

	tr.log.Debug("Updated tokens for user_oauth_id %s, new expiry: %s", userOauthID, expiresAt)
	return nil
}
