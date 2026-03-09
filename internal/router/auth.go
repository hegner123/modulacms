package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// LoginHandler handles password-based authentication.
// It validates credentials, creates a session, and sets an HTTP-only cookie.
func LoginHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		utility.DefaultLogger.Error("Failed to decode login request", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := svc.Auth.Login(r.Context(), service.LoginInput{
		Email:     credentials.Email,
		Password:  credentials.Password,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	if err := middleware.WriteCookie(w, cfg, result.SessionToken, result.User.UserID); err != nil {
		utility.DefaultLogger.Error("Failed to set cookie", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"user_id":    result.User.UserID,
		"email":      result.User.Email,
		"username":   result.User.Username,
		"created_at": result.User.DateCreated,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	utility.DefaultLogger.Info("User logged in successfully:", result.User.Email)
}

// LogoutHandler clears the session cookie and invalidates the session.
func LogoutHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie(cfg.Cookie_Name)
	if err != nil {
		// No cookie, already logged out.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
		return
	}

	userCookie, err := middleware.ReadCookie(cookie)
	if err == nil && userCookie != nil {
		utility.DefaultLogger.Info("User logged out:", userCookie.UserId)
	}

	// Clear the cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Cookie_Name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// MeHandler returns information about the currently authenticated user.
func MeHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie(cfg.Cookie_Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	user, err := middleware.UserIsAuth(r, cookie, cfg)
	if err != nil {
		utility.DefaultLogger.Error("Session validation failed", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

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

// RegisterHandler handles user self-registration.
func RegisterHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req struct {
		Username string      `json:"username"`
		Name     string      `json:"name"`
		Email    types.Email `json:"email"`
		Password string      `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	ac := middleware.AuditContextFromRequest(r, *cfg)
	user, err := svc.Auth.Register(r.Context(), ac, service.RegisterInput{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// ResetPasswordHandler delegates to the user update flow.
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	apiUpdateUser(w, r, svc)
}

// RequestPasswordResetHandler initiates a token-based password reset flow.
// Always returns 200 regardless of whether the email exists to prevent user enumeration.
func RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	ac := middleware.AuditContextFromRequest(r, *cfg)
	err = svc.Auth.RequestPasswordReset(r.Context(), ac, service.PasswordResetRequestInput{
		Email: req.Email,
	})
	if err != nil {
		// Log but do not expose the error to prevent user enumeration.
		utility.DefaultLogger.Error("password reset request failed", err)
	}

	// Always return success.
	successMsg := map[string]string{"message": "If an account with that email exists, a reset link has been sent."}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(successMsg)
}

// ConfirmPasswordResetHandler validates a password reset token and sets the new password.
func ConfirmPasswordResetHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	ac := middleware.AuditContextFromRequest(r, *cfg)
	err = svc.Auth.ConfirmPasswordReset(r.Context(), ac, service.PasswordResetConfirmInput{
		Token:    req.Token,
		Password: req.Password,
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password has been reset successfully."})
}

// OauthInitiateHandler starts the OAuth flow with PKCE and state parameter for CSRF protection.
func OauthInitiateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authURL, err := svc.Auth.GetOAuthAuthURL()
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		utility.DefaultLogger.Info("Redirecting to OAuth provider:", cfg.Oauth_Provider_Name)
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

// OauthCallbackHandler handles the OAuth provider's redirect with state validation and PKCE.
func OauthCallbackHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			utility.DefaultLogger.Error("Missing code parameter", nil)
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		state := r.URL.Query().Get("state")
		if err := auth.ValidateState(state); err != nil {
			utility.DefaultLogger.Error("State validation failed", err)
			http.Error(w, "Invalid or expired state", http.StatusBadRequest)
			return
		}

		verifier, err := auth.GetVerifier(state)
		if err != nil {
			utility.DefaultLogger.Error("Verifier retrieval failed", err)
			http.Error(w, "Invalid session", http.StatusBadRequest)
			return
		}

		result, err := svc.Auth.HandleOAuthCallback(r.Context(), service.OAuthCallbackInput{
			Code:      code,
			Verifier:  verifier,
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
		if err != nil {
			utility.DefaultLogger.Error("OAuth callback failed", err)
			http.Error(w, "OAuth login failed", http.StatusInternalServerError)
			return
		}

		cfg, err := svc.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		if err := middleware.WriteCookie(w, cfg, result.SessionToken, result.User.UserID); err != nil {
			utility.DefaultLogger.Error("Cookie creation failed", err)
			http.Error(w, "Cookie creation failed", http.StatusInternalServerError)
			return
		}

		redirectURL := cfg.Oauth_Success_Redirect
		if redirectURL == "" {
			redirectURL = "/"
		}

		utility.DefaultLogger.Info("OAuth login successful for user:", result.User.Email, "user_id:", result.User.UserID)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}
