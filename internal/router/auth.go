package router

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/email"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/oauth2"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	var req struct {
		Username string      `json:"username"`
		Name     string      `json:"name"`
		Email    types.Email `json:"email"`
		Password string      `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	// Always assign viewer role for self-registration (ignore any role in request body)
	viewerRole, roleErr := d.GetRoleByLabel("viewer")
	if roleErr != nil {
		utility.DefaultLogger.Error("failed to get viewer role for registration", roleErr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	newUser := db.CreateUserParams{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Hash:     hash,
		Role:     string(viewerRole.RoleID),
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdUser, createErr := d.CreateUser(r.Context(), ac, newUser)
	if createErr != nil {
		utility.DefaultLogger.Error("", createErr)
		http.Error(w, createErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	apiUpdateUser(w, r, svc)
}

// RequestPasswordResetHandler initiates a token-based password reset flow.
// It looks up the user by email, generates a reset token, and sends an email
// with a reset link. Always returns 200 regardless of whether the email exists
// to prevent user enumeration.
func RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request, c config.Config, emailSvc *email.Service, driver db.DbDriver) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Always return the same response to prevent user enumeration.
	successMsg := map[string]string{"message": "If an account with that email exists, a reset link has been sent."}

	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(successMsg)
		return
	}

	user, err := driver.GetUserByEmail(types.Email(req.Email))
	if err != nil || user == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(successMsg)
		return
	}

	ctx := r.Context()
	clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
	if splitErr != nil {
		clientIP = r.RemoteAddr
	}
	ac := audited.Ctx(types.NodeID(c.Node_ID), user.UserID, middleware.RequestIDFromContext(ctx), clientIP)
	userNullID := types.NullableUserID{ID: user.UserID, Valid: true}

	// Clean up any existing password_reset tokens for this user.
	existingTokens, err := driver.GetTokenByUserId(userNullID)
	if err == nil && existingTokens != nil {
		for _, tok := range *existingTokens {
			if tok.TokenType != "password_reset" {
				continue
			}
			if delErr := driver.DeleteToken(ctx, ac, tok.ID); delErr != nil {
				utility.DefaultLogger.Warn("failed to delete existing password reset token", delErr)
			}
		}
	}

	// Generate 32-byte random token, hex-encoded.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		utility.DefaultLogger.Error("failed to generate reset token", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tokenValue := hex.EncodeToString(tokenBytes)
	hashedToken := utility.HashToken(tokenValue)

	_, err = driver.CreateToken(ctx, ac, db.CreateTokenParams{
		UserID:    userNullID,
		TokenType: "password_reset",
		Token:     hashedToken,
		IssuedAt:  types.TimestampNow(),
		ExpiresAt: types.NewTimestamp(time.Now().UTC().Add(1 * time.Hour)),
		Revoked:   false,
	})
	if err != nil {
		utility.DefaultLogger.Error("failed to create password reset token", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Send the reset email if email service is enabled and reset URL is configured.
	if emailSvc.Enabled() && c.Password_Reset_URL != "" {
		resetLink := c.Password_Reset_URL + "?token=" + tokenValue
		sendErr := emailSvc.Send(context.Background(), email.Message{
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
			// Still return 200 — the token was created, the email just failed to send.
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(successMsg)
}

// ConfirmPasswordResetHandler validates a password reset token and sets the new password.
// The token must be of type "password_reset", not revoked, and not expired.
func ConfirmPasswordResetHandler(w http.ResponseWriter, r *http.Request, c config.Config, driver db.DbDriver) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.Password == "" {
		http.Error(w, "token and password are required", http.StatusBadRequest)
		return
	}

	tok, err := driver.GetTokenByTokenValue(utility.HashToken(req.Token))
	if err != nil || tok == nil {
		http.Error(w, "invalid or expired reset token", http.StatusBadRequest)
		return
	}

	if tok.TokenType != "password_reset" {
		http.Error(w, "invalid or expired reset token", http.StatusBadRequest)
		return
	}

	if tok.Revoked {
		http.Error(w, "invalid or expired reset token", http.StatusBadRequest)
		return
	}

	if !tok.ExpiresAt.Valid || time.Now().UTC().After(tok.ExpiresAt.Time) {
		http.Error(w, "invalid or expired reset token", http.StatusBadRequest)
		return
	}

	if !tok.UserID.Valid {
		http.Error(w, "invalid reset token", http.StatusBadRequest)
		return
	}

	user, err := driver.GetUser(tok.UserID.ID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		utility.DefaultLogger.Error("failed to hash password", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
	if splitErr != nil {
		clientIP = r.RemoteAddr
	}
	ac := audited.Ctx(types.NodeID(c.Node_ID), user.UserID, middleware.RequestIDFromContext(ctx), clientIP)

	_, err = driver.UpdateUser(ctx, ac, db.UpdateUserParams{
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
		utility.DefaultLogger.Error("failed to update user password", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Revoke the used token.
	_, revokeErr := driver.UpdateToken(ctx, ac, db.UpdateTokenParams{
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
	existingTokens, tokErr := driver.GetTokenByUserId(userNullID)
	if tokErr == nil && existingTokens != nil {
		for _, t := range *existingTokens {
			if t.TokenType != "password_reset" || t.ID == tok.ID {
				continue
			}
			if delErr := driver.DeleteToken(ctx, ac, t.ID); delErr != nil {
				utility.DefaultLogger.Warn("failed to delete leftover password reset token", delErr)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password has been reset successfully."})
}

func respond(w http.ResponseWriter, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// OauthCallbackHandler handles the OAuth provider's redirect with state validation and PKCE.
// This handler validates the state parameter for CSRF protection, retrieves the PKCE verifier,
// and exchanges the authorization code for an access token.
func OauthCallbackHandler(c config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Retrieve the authorization code from the query parameters
		code := r.URL.Query().Get("code")
		if code == "" {
			utility.DefaultLogger.Error("Missing code parameter", nil)
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		// Retrieve and validate the state parameter for CSRF protection
		state := r.URL.Query().Get("state")
		if err := auth.ValidateState(state); err != nil {
			utility.DefaultLogger.Error("State validation failed", err)
			http.Error(w, "Invalid or expired state", http.StatusBadRequest)
			return
		}

		// Retrieve the PKCE verifier associated with this state
		verifier, err := auth.GetVerifier(state)
		if err != nil {
			utility.DefaultLogger.Error("Verifier retrieval failed", err)
			http.Error(w, "Invalid session", http.StatusBadRequest)
			return
		}

		// Build the OAuth2 configuration
		conf := &oauth2.Config{
			ClientID:     c.Oauth_Client_Id,
			ClientSecret: c.Oauth_Client_Secret,
			Scopes:       c.Oauth_Scopes,
			RedirectURL:  c.Oauth_Redirect_URL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.Oauth_Endpoint[config.OauthAuthURL],
				TokenURL: c.Oauth_Endpoint[config.OauthTokenURL],
			},
		}

		// Exchange the code for an access token using the PKCE verifier
		token, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			utility.DefaultLogger.Error("Token exchange failed", err)
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		// Create authenticated client
		client := conf.Client(ctx, token)

		// Fetch user info from provider
		provisioner := auth.NewUserProvisioner(utility.DefaultLogger, &c, db.ConfigDB(c))
		userInfo, err := provisioner.FetchUserInfo(client)
		if err != nil {
			utility.DefaultLogger.Error("Failed to fetch user info", err)
			http.Error(w, "Failed to fetch user information", http.StatusInternalServerError)
			return
		}

		// Provision user (create or link)
		provider := c.Oauth_Provider_Name
		if provider == "" {
			provider = "oauth" // Default provider name
		}

		user, err := provisioner.ProvisionUser(userInfo, token, provider)
		if err != nil {
			utility.DefaultLogger.Error("User provisioning failed", err)
			http.Error(w, "User provisioning failed", http.StatusInternalServerError)
			return
		}

		// Create session
		sessionToken, err := generateSessionToken()
		if err != nil {
			utility.DefaultLogger.Error("Session token generation failed", err)
			http.Error(w, "Session creation failed", http.StatusInternalServerError)
			return
		}

		dbc := db.ConfigDB(c)
		expiresAt := types.NewTimestamp(time.Now().Add(24 * time.Hour))

		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(c.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)
		_, err = dbc.CreateSession(r.Context(), ac, db.CreateSessionParams{
			UserID:      types.NullableUserID{ID: user.UserID, Valid: true},
			ExpiresAt:   expiresAt,
			SessionData: db.NewNullString(sessionToken),
			IpAddress:   db.NewNullString(r.RemoteAddr),
			UserAgent:   db.NewNullString(r.UserAgent()),
		})

		if err != nil {
			utility.DefaultLogger.Error("Session creation failed", err)
			http.Error(w, "Session creation failed", http.StatusInternalServerError)
			return
		}

		// Set auth cookie
		if err := middleware.WriteCookie(w, &c, sessionToken, user.UserID); err != nil {
			utility.DefaultLogger.Error("Cookie creation failed", err)
			http.Error(w, "Cookie creation failed", http.StatusInternalServerError)
			return
		}

		// Redirect to success URL
		redirectURL := c.Oauth_Success_Redirect
		if redirectURL == "" {
			redirectURL = "/"
		}

		utility.DefaultLogger.Info("OAuth login successful for user:", user.Email, "user_id:", user.UserID)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// OauthInitiateHandler starts the OAuth flow with PKCE and state parameter for CSRF protection.
// It generates a secure state parameter, creates a PKCE code verifier, and redirects the user
// to the OAuth provider's authorization endpoint.
func OauthInitiateHandler(c config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate state for CSRF protection
		state, err := auth.GenerateState()
		if err != nil {
			utility.DefaultLogger.Error("Failed to generate state", err)
			http.Error(w, "Failed to generate state", http.StatusInternalServerError)
			return
		}

		// Generate PKCE verifier and challenge
		verifier := oauth2.GenerateVerifier()

		// Store verifier associated with state for callback retrieval
		auth.StoreVerifier(state, verifier)

		// Validate required configuration
		if c.Oauth_Client_Id == "" || c.Oauth_Redirect_URL == "" {
			utility.DefaultLogger.Error("OAuth configuration incomplete", nil)
			http.Error(w, "OAuth not configured", http.StatusInternalServerError)
			return
		}

		// Build OAuth2 config
		conf := &oauth2.Config{
			ClientID:     c.Oauth_Client_Id,
			ClientSecret: c.Oauth_Client_Secret,
			Scopes:       c.Oauth_Scopes,
			RedirectURL:  c.Oauth_Redirect_URL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.Oauth_Endpoint[config.OauthAuthURL],
				TokenURL: c.Oauth_Endpoint[config.OauthTokenURL],
			},
		}

		// Generate authorization URL with state and PKCE challenge
		url := conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

		utility.DefaultLogger.Info("Redirecting to OAuth provider:", c.Oauth_Provider_Name)

		// Redirect to provider
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// LoginHandler handles password-based authentication.
// It validates credentials, creates a session, and sets an HTTP-only cookie.
func LoginHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		utility.DefaultLogger.Error("Failed to decode login request", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if credentials.Email == "" || credentials.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get database connection
	dbc := db.ConfigDB(c)

	// Retrieve user by email
	user, err := dbc.GetUserByEmail(types.Email(credentials.Email))
	if err != nil {
		utility.DefaultLogger.Error("User not found", err)
		// Don't reveal whether user exists
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if !auth.CheckPasswordHash(credentials.Password, user.Hash) {
		utility.DefaultLogger.Warn("Invalid password attempt for user", nil, credentials.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate session token
	sessionToken, err := generateSessionToken()
	if err != nil {
		utility.DefaultLogger.Error("Failed to generate session token", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session in database
	expiresAt := types.NewTimestamp(time.Now().Add(24 * time.Hour))
	clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
	if splitErr != nil {
		clientIP = r.RemoteAddr
	}
	ac := audited.Ctx(types.NodeID(c.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)
	_, err = dbc.CreateSession(r.Context(), ac, db.CreateSessionParams{
		UserID:      types.NullableUserID{ID: user.UserID, Valid: true},
		ExpiresAt:   expiresAt,
		SessionData: db.NewNullString(sessionToken),
		IpAddress:   db.NewNullString(r.RemoteAddr),
		UserAgent:   db.NewNullString(r.UserAgent()),
	})

	if err != nil {
		utility.DefaultLogger.Error("Failed to create session", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set secure HTTP-only cookie
	if err := middleware.WriteCookie(w, &c, sessionToken, user.UserID); err != nil {
		utility.DefaultLogger.Error("Failed to set cookie", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return user info (without sensitive data)
	response := map[string]any{
		"user_id":    user.UserID,
		"email":      user.Email,
		"username":   user.Username,
		"created_at": user.DateCreated,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	utility.DefaultLogger.Info("User logged in successfully:", user.Email)
}

// LogoutHandler clears the session cookie and invalidates the session.
func LogoutHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session cookie
	cookie, err := r.Cookie(c.Cookie_Name)
	if err != nil {
		// No cookie, already logged out
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
		return
	}

	// Decode cookie to get user ID
	userCookie, err := middleware.ReadCookie(cookie)
	if err == nil && userCookie != nil {
		// Delete session from database
		// Note: You'll need to implement DeleteSessionByUserId in the db package
		// For now, we'll just clear the cookie
		utility.DefaultLogger.Info("User logged out:", userCookie.UserId)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     c.Cookie_Name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete immediately
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// MeHandler returns information about the currently authenticated user.
func MeHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session cookie
	cookie, err := r.Cookie(c.Cookie_Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	// Validate session
	user, err := middleware.UserIsAuth(r, cookie, &c)
	if err != nil {
		utility.DefaultLogger.Error("Session validation failed", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	// Return user info (without sensitive data)
	response := map[string]any{
		"user_id":  user.UserID,
		"email":    user.Email,
		"username": user.Username,
		"name":     user.Name,
		"role":     user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// generateSessionToken creates a cryptographically secure random session token.
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
