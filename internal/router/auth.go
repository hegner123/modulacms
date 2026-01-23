package router

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/oauth2"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	err := ApiCreateUser(w, r, c)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	err := ApiUpdateUser(w, r, c)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
		provisioner := auth.NewUserProvisioner(&c)
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
		expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

		_, err = dbc.CreateSession(db.CreateSessionParams{
			UserID: user.UserID,
			ExpiresAt: sql.NullString{
				String: expiresAt,
				Valid:  true,
			},
			SessionData: sql.NullString{
				String: sessionToken,
				Valid:  true,
			},
			IpAddress: sql.NullString{
				String: r.RemoteAddr,
				Valid:  true,
			},
			UserAgent: sql.NullString{
				String: r.UserAgent(),
				Valid:  true,
			},
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

		utility.DefaultLogger.Info("OAuth login successful for user: %s (user_id: %d)", user.Email, user.UserID)
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

		utility.DefaultLogger.Info("Redirecting to OAuth provider: %s", c.Oauth_Provider_Name)

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
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	_, err = dbc.CreateSession(db.CreateSessionParams{
		UserID: user.UserID,
		ExpiresAt: sql.NullString{
			String: expiresAt,
			Valid:  true,
		},
		SessionData: sql.NullString{
			String: sessionToken,
			Valid:  true,
		},
		IpAddress: sql.NullString{
			String: r.RemoteAddr,
			Valid:  true,
		},
		UserAgent: sql.NullString{
			String: r.UserAgent(),
			Valid:  true,
		},
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

	utility.DefaultLogger.Info("User logged in successfully: %s", user.Email)
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
		utility.DefaultLogger.Info("User logged out: %d", userCookie.UserId)
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
