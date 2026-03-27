package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/email"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/oauth2"
)

// AuthService handles authentication business logic including login, registration,
// password resets, and OAuth flows.
type AuthService struct {
	driver   db.DbDriver
	mgr      *config.Manager
	emailSvc *email.Service
}

// NewAuthService creates an AuthService with the given dependencies.
func NewAuthService(driver db.DbDriver, mgr *config.Manager, emailSvc *email.Service) *AuthService {
	return &AuthService{driver: driver, mgr: mgr, emailSvc: emailSvc}
}

// LoginInput contains credentials for password-based authentication.
type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

// LoginResult contains the authenticated user and session token.
type LoginResult struct {
	User         *db.Users
	SessionToken string
}

// RegisterInput contains fields for user self-registration.
type RegisterInput struct {
	Username string
	Name     string
	Email    types.Email
	Password string
}

// PasswordResetRequestInput contains the email for a password reset request.
type PasswordResetRequestInput struct {
	Email string
}

// PasswordResetConfirmInput contains the token and new password for confirming a reset.
type PasswordResetConfirmInput struct {
	Token    string
	Password string
}

// OAuthCallbackInput contains the authorization code and PKCE verifier from the OAuth callback.
type OAuthCallbackInput struct {
	Code      string
	Verifier  string
	IPAddress string
	UserAgent string
}

// OAuthCallbackResult contains the provisioned user and session token from OAuth login.
type OAuthCallbackResult struct {
	User         *db.Users
	SessionToken string
}

// Login validates credentials, creates a session, and returns the user with a session token.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	if input.Email == "" || input.Password == "" {
		return nil, &ValidationError{Errors: []FieldError{
			{Field: "credentials", Message: "email and password are required"},
		}}
	}

	user, err := s.driver.GetUserByEmail(types.Email(input.Email))
	if err != nil {
		return nil, &UnauthorizedError{Message: "invalid credentials"}
	}

	if !auth.CheckPasswordHash(input.Password, user.Hash) {
		return nil, &UnauthorizedError{Message: "invalid credentials"}
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	expiresAt := types.NewTimestamp(time.Now().Add(24 * time.Hour))

	cfg, cfgErr := s.mgr.Config()
	if cfgErr != nil {
		return nil, fmt.Errorf("load config for session creation: %w", cfgErr)
	}
	ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, "", input.IPAddress)

	_, err = s.driver.CreateSession(ctx, ac, db.CreateSessionParams{
		UserID:      types.NullableUserID{ID: user.UserID, Valid: true},
		ExpiresAt:   expiresAt,
		SessionData: db.NewNullString(sessionToken),
		IpAddress:   db.NewNullString(input.IPAddress),
		UserAgent:   db.NewNullString(input.UserAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &LoginResult{User: user, SessionToken: sessionToken}, nil
}

// Register creates a new user with the viewer role.
func (s *AuthService) Register(ctx context.Context, ac audited.AuditContext, input RegisterInput) (*db.Users, error) {
	if input.Password == "" {
		return nil, NewValidationError("password", "password is required")
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	viewerRole, err := s.driver.GetRoleByLabel("viewer")
	if err != nil {
		return nil, fmt.Errorf("get viewer role: %w", err)
	}

	user, err := s.driver.CreateUser(ctx, ac, db.CreateUserParams{
		Username: input.Username,
		Name:     input.Name,
		Email:    input.Email,
		Hash:     hash,
		Role:     string(viewerRole.RoleID),
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// RequestPasswordReset generates a reset token and sends an email.
// It silently succeeds if the email is not found to prevent user enumeration.
func (s *AuthService) RequestPasswordReset(ctx context.Context, ac audited.AuditContext, input PasswordResetRequestInput) error {
	if input.Email == "" {
		return nil
	}

	user, err := s.driver.GetUserByEmail(types.Email(input.Email))
	if err != nil || user == nil {
		return nil
	}

	userNullID := types.NullableUserID{ID: user.UserID, Valid: true}

	// Clean up existing password_reset tokens for this user.
	existingTokens, err := s.driver.GetTokenByUserId(userNullID)
	if err == nil && existingTokens != nil {
		for _, tok := range *existingTokens {
			if tok.TokenType != types.TokenTypePasswordReset {
				continue
			}
			if delErr := s.driver.DeleteToken(ctx, ac, tok.ID); delErr != nil {
				utility.DefaultLogger.Warn("failed to delete existing password reset token", delErr)
			}
		}
	}

	// Generate 32-byte random token, hex-encoded.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}
	tokenValue := hex.EncodeToString(tokenBytes)
	hashedToken := utility.HashToken(tokenValue)

	_, err = s.driver.CreateToken(ctx, ac, db.CreateTokenParams{
		UserID:    userNullID,
		TokenType: types.TokenTypePasswordReset,
		Token:     hashedToken,
		IssuedAt:  types.TimestampNow(),
		ExpiresAt: types.NewTimestamp(time.Now().UTC().Add(1 * time.Hour)),
		Revoked:   false,
	})
	if err != nil {
		return fmt.Errorf("create password reset token: %w", err)
	}

	// Send the reset email if email service is enabled and reset URL is configured.
	cfg, cfgErr := s.mgr.Config()
	if cfgErr != nil {
		return fmt.Errorf("load config for password reset: %w", cfgErr)
	}

	if s.emailSvc.Enabled() && cfg.Password_Reset_URL != "" {
		resetLink := cfg.Password_Reset_URL + "?token=" + tokenValue
		sendErr := s.emailSvc.Send(context.Background(), email.Message{
			To:      []email.Address{email.NewAddress(user.Name, string(user.Email))},
			Subject: "Password Reset Request",
			PlainBody: "You requested a password reset for your Modula account.\n\n" +
				"Click the link below to reset your password:\n" + resetLink + "\n\n" +
				"This link expires in 1 hour. If you did not request this, ignore this email.",
			HTMLBody: "<p>You requested a password reset for your Modula account.</p>" +
				"<p><a href=\"" + resetLink + "\">Click here to reset your password</a></p>" +
				"<p>This link expires in 1 hour. If you did not request this, ignore this email.</p>",
		})
		if sendErr != nil {
			utility.DefaultLogger.Error("failed to send password reset email", sendErr)
			// Still succeed -- the token was created, the email just failed to send.
		}
	}

	return nil
}

// ConfirmPasswordReset validates a reset token and sets the new password.
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, ac audited.AuditContext, input PasswordResetConfirmInput) error {
	if input.Token == "" || input.Password == "" {
		return NewValidationError("credentials", "token and password are required")
	}

	tok, err := s.driver.GetTokenByTokenValue(utility.HashToken(input.Token))
	if err != nil || tok == nil {
		return &UnauthorizedError{Message: "invalid or expired reset token"}
	}

	if tok.TokenType != types.TokenTypePasswordReset {
		return &UnauthorizedError{Message: "invalid or expired reset token"}
	}

	if tok.Revoked {
		return &UnauthorizedError{Message: "invalid or expired reset token"}
	}

	if !tok.ExpiresAt.Valid || time.Now().UTC().After(tok.ExpiresAt.Time) {
		return &UnauthorizedError{Message: "invalid or expired reset token"}
	}

	if !tok.UserID.Valid {
		return &UnauthorizedError{Message: "invalid reset token"}
	}

	user, err := s.driver.GetUser(tok.UserID.ID)
	if err != nil || user == nil {
		return &NotFoundError{Resource: "user", ID: string(tok.UserID.ID)}
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = s.driver.UpdateUser(ctx, ac, db.UpdateUserParams{
		UserID:       user.UserID,
		Username:     user.Username,
		Name:         user.Name,
		Email:        user.Email,
		Hash:         hash,
		Role:         user.Role,
		DateCreated:  user.DateCreated,
		DateModified: types.NewTimestamp(time.Now().UTC()),
	})
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	// Revoke the used token.
	_, revokeErr := s.driver.UpdateToken(ctx, ac, db.UpdateTokenParams{
		ID:        tok.ID,
		Token:     tok.Token,
		IssuedAt:  tok.IssuedAt,
		ExpiresAt: tok.ExpiresAt,
		Revoked:   true,
	})
	if revokeErr != nil {
		utility.DefaultLogger.Warn("failed to revoke used password reset token", revokeErr)
	}

	// Clean up any other password_reset tokens for this user.
	userNullID := types.NullableUserID{ID: user.UserID, Valid: true}
	existingTokens, tokErr := s.driver.GetTokenByUserId(userNullID)
	if tokErr == nil && existingTokens != nil {
		for _, t := range *existingTokens {
			if t.TokenType != types.TokenTypePasswordReset || t.ID == tok.ID {
				continue
			}
			if delErr := s.driver.DeleteToken(ctx, ac, t.ID); delErr != nil {
				utility.DefaultLogger.Warn("failed to delete leftover password reset token", delErr)
			}
		}
	}

	return nil
}

// GetOAuthAuthURL generates the OAuth authorization URL with PKCE and state parameters.
func (s *AuthService) GetOAuthAuthURL() (string, error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return "", fmt.Errorf("load config for OAuth: %w", err)
	}

	if cfg.Oauth_Client_Id == "" || cfg.Oauth_Redirect_URL == "" {
		return "", &ValidationError{Errors: []FieldError{
			{Field: "oauth", Message: "OAuth not configured"},
		}}
	}

	state, err := auth.GenerateState()
	if err != nil {
		return "", fmt.Errorf("generate OAuth state: %w", err)
	}

	verifier := oauth2.GenerateVerifier()
	auth.StoreVerifier(state, verifier)

	conf := s.oauthConfig(cfg)
	authURL := conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	return authURL, nil
}

// HandleOAuthCallback exchanges the authorization code for a token, provisions the user,
// and creates a session.
func (s *AuthService) HandleOAuthCallback(ctx context.Context, input OAuthCallbackInput) (*OAuthCallbackResult, error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("load config for OAuth callback: %w", err)
	}

	conf := s.oauthConfig(cfg)

	token, err := conf.Exchange(ctx, input.Code, oauth2.VerifierOption(input.Verifier))
	if err != nil {
		return nil, fmt.Errorf("OAuth token exchange: %w", err)
	}

	client := conf.Client(ctx, token)

	provisioner := auth.NewUserProvisioner(utility.DefaultLogger, cfg, s.driver)
	userInfo, err := provisioner.FetchUserInfo(client)
	if err != nil {
		return nil, fmt.Errorf("fetch OAuth user info: %w", err)
	}

	provider := cfg.Oauth_Provider_Name
	if provider == "" {
		provider = "oauth"
	}

	user, err := provisioner.ProvisionUser(userInfo, token, provider)
	if err != nil {
		return nil, fmt.Errorf("provision OAuth user: %w", err)
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	expiresAt := types.NewTimestamp(time.Now().Add(24 * time.Hour))
	ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, "", input.IPAddress)

	_, err = s.driver.CreateSession(ctx, ac, db.CreateSessionParams{
		UserID:      types.NullableUserID{ID: user.UserID, Valid: true},
		ExpiresAt:   expiresAt,
		SessionData: db.NewNullString(sessionToken),
		IpAddress:   db.NewNullString(input.IPAddress),
		UserAgent:   db.NewNullString(input.UserAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("create OAuth session: %w", err)
	}

	return &OAuthCallbackResult{User: user, SessionToken: sessionToken}, nil
}

// oauthConfig builds an oauth2.Config from the application config.
func (s *AuthService) oauthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.Oauth_Client_Id,
		ClientSecret: cfg.Oauth_Client_Secret,
		Scopes:       cfg.Oauth_Scopes,
		RedirectURL:  cfg.Oauth_Redirect_URL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.Oauth_Endpoint[config.OauthAuthURL],
			TokenURL: cfg.Oauth_Endpoint[config.OauthTokenURL],
		},
	}
}

// generateSessionToken creates a cryptographically secure random session token.
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
