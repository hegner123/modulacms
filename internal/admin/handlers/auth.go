package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/email"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// LoginPageHandler renders the login form.
func LoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If already authenticated, redirect to admin
		if middleware.AuthenticatedUser(r.Context()) != nil {
			http.Redirect(w, r, "/admin/", http.StatusFound)
			return
		}
		nextURL := r.URL.Query().Get("next")
		if nextURL == "" {
			nextURL = "/admin/"
		}
		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, pages.Login(csrfToken, utility.Version, nextURL, ""))
	}
}

// LoginSubmitHandler processes login form submissions.
func LoginSubmitHandler(mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		nextURL := r.FormValue("next")

		if nextURL == "" || !strings.HasPrefix(nextURL, "/admin") {
			nextURL = "/admin/"
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if email == "" || password == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Render(w, r, pages.Login(csrfToken, utility.Version, nextURL, "Email and password are required"))
			return
		}

		d := db.ConfigDB(*cfg)

		user, userErr := d.GetUserByEmail(types.Email(email))
		if userErr != nil || user == nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Render(w, r, pages.Login(csrfToken, utility.Version, nextURL, "Invalid credentials"))
			return
		}

		if !auth.CheckPasswordHash(password, user.Hash) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Render(w, r, pages.Login(csrfToken, utility.Version, nextURL, "Invalid credentials"))
			return
		}

		// Generate session token
		tokenBytes := make([]byte, 32)
		if _, randErr := rand.Read(tokenBytes); randErr != nil {
			utility.DefaultLogger.Error("Failed to generate session token", randErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		sessionToken := hex.EncodeToString(tokenBytes)

		// Create session in database
		expiresAt := types.NewTimestamp(time.Now().Add(24 * time.Hour))
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)
		_, sessionErr := d.CreateSession(r.Context(), ac, db.CreateSessionParams{
			UserID:      types.NullableUserID{ID: user.UserID, Valid: true},
			ExpiresAt:   expiresAt,
			SessionData: db.NewNullString(sessionToken),
			IpAddress:   db.NewNullString(r.RemoteAddr),
			UserAgent:   db.NewNullString(r.UserAgent()),
		})
		if sessionErr != nil {
			utility.DefaultLogger.Error("Failed to create session", sessionErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set cookie
		if cookieErr := middleware.WriteCookie(w, cfg, sessionToken, user.UserID); cookieErr != nil {
			utility.DefaultLogger.Error("Failed to set cookie", cookieErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, nextURL, http.StatusSeeOther)
	}
}

// LogoutHandler clears session and redirects to login.
func LogoutHandler(mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		// Clear session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     cfg.Cookie_Name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   cfg.Cookie_Secure,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	}
}

// ForgotPasswordPageHandler renders the forgot password form.
func ForgotPasswordPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, pages.ForgotPassword(csrfToken, utility.Version, "", ""))
	}
}

// ForgotPasswordSubmitHandler processes the forgot password form.
// Creates a password reset token and sends an email if the email service is configured.
// Always shows a success message regardless of whether the email exists to prevent enumeration.
func ForgotPasswordSubmitHandler(mgr *config.Manager, emailSvc *email.Service, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		addr := strings.TrimSpace(r.FormValue("email"))

		if addr == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Render(w, r, pages.ForgotPassword(csrfToken, utility.Version, "Email is required", ""))
			return
		}

		// Always show success to prevent user enumeration.
		successMsg := "If an account with that email exists, a reset link has been sent."

		cfg, err := mgr.Config()
		if err != nil {
			Render(w, r, pages.ForgotPassword(csrfToken, utility.Version, "", successMsg))
			return
		}

		user, err := driver.GetUserByEmail(types.Email(addr))
		if err != nil || user == nil {
			Render(w, r, pages.ForgotPassword(csrfToken, utility.Version, "", successMsg))
			return
		}

		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)
		userNullID := types.NullableUserID{ID: user.UserID, Valid: true}

		// Clean up existing password_reset tokens for this user.
		existingTokens, tokErr := driver.GetTokenByUserId(userNullID)
		if tokErr == nil && existingTokens != nil {
			for _, tok := range *existingTokens {
				if tok.TokenType != types.TokenTypePasswordReset {
					continue
				}
				if delErr := driver.DeleteToken(r.Context(), ac, tok.ID); delErr != nil {
					utility.DefaultLogger.Warn("failed to delete existing password reset token", delErr)
				}
			}
		}

		tokenBytes := make([]byte, 32)
		if _, randErr := rand.Read(tokenBytes); randErr != nil {
			utility.DefaultLogger.Error("failed to generate reset token", randErr)
			Render(w, r, pages.ForgotPassword(csrfToken, utility.Version, "", successMsg))
			return
		}
		tokenValue := hex.EncodeToString(tokenBytes)

		_, err = driver.CreateToken(r.Context(), ac, db.CreateTokenParams{
			UserID:    userNullID,
			TokenType: types.TokenTypePasswordReset,
			Token:     tokenValue,
			IssuedAt:  types.TimestampNow(),
			ExpiresAt: types.NewTimestamp(time.Now().UTC().Add(1 * time.Hour)),
			Revoked:   false,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create password reset token", err)
			Render(w, r, pages.ForgotPassword(csrfToken, utility.Version, "", successMsg))
			return
		}

		if emailSvc.Enabled() {
			resetLink := "https://" + r.Host + "/admin/reset-password?token=" + tokenValue
			if cfg.Password_Reset_URL != "" {
				resetLink = cfg.Password_Reset_URL + "?token=" + tokenValue
			}
			sendErr := emailSvc.Send(context.Background(), email.Message{
				To:      []email.Address{email.NewAddress(user.Name, string(user.Email))},
				Subject: "Password Reset Request",
				PlainBody: "You requested a password reset for your ModulaCMS account.\n\n" +
					"Click the link below to reset your password:\n" + resetLink + "\n\n" +
					"This link expires in 1 hour. If you did not request this, ignore this email.",
				HTMLBody: "<p>You requested a password reset for your ModulaCMS account.</p>" +
					"<p><a href=\"" + resetLink + "\">Click here to reset your password</a></p>" +
					"<p>This link expires in 1 hour. If you did not request this, ignore this email.</p>",
			})
			if sendErr != nil {
				utility.DefaultLogger.Error("failed to send password reset email", sendErr)
			}
		}

		Render(w, r, pages.ForgotPassword(csrfToken, utility.Version, "", successMsg))
	}
}

// ResetPasswordPageHandler renders the password reset form.
// Requires a valid token query parameter.
func ResetPasswordPageHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := CSRFTokenFromContext(r.Context())
		token := r.URL.Query().Get("token")

		if token == "" {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "Invalid or missing reset token.", ""))
			return
		}

		tok, err := driver.GetTokenByTokenValue(token)
		if err != nil || tok == nil || tok.TokenType != types.TokenTypePasswordReset || tok.Revoked {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "Invalid or expired reset token.", ""))
			return
		}

		if !tok.ExpiresAt.Valid || time.Now().UTC().After(tok.ExpiresAt.Time) {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "This reset link has expired. Please request a new one.", ""))
			return
		}

		Render(w, r, pages.ResetPassword(csrfToken, utility.Version, token, "", ""))
	}
}

// ResetPasswordSubmitHandler processes the password reset form.
// Validates the token, updates the password, and revokes the token.
func ResetPasswordSubmitHandler(mgr *config.Manager, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		token := r.FormValue("token")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if token == "" {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "Invalid or missing reset token.", ""))
			return
		}

		if password == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, token, "Password is required.", ""))
			return
		}

		if len(password) < 8 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, token, "Password must be at least 8 characters.", ""))
			return
		}

		if password != confirmPassword {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, token, "Passwords do not match.", ""))
			return
		}

		tok, err := driver.GetTokenByTokenValue(token)
		if err != nil || tok == nil || tok.TokenType != types.TokenTypePasswordReset || tok.Revoked {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "Invalid or expired reset token.", ""))
			return
		}

		if !tok.ExpiresAt.Valid || time.Now().UTC().After(tok.ExpiresAt.Time) {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "This reset link has expired. Please request a new one.", ""))
			return
		}

		if !tok.UserID.Valid {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "Invalid reset token.", ""))
			return
		}

		user, err := driver.GetUser(tok.UserID.ID)
		if err != nil || user == nil {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "User not found.", ""))
			return
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			utility.DefaultLogger.Error("failed to hash password", err)
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, token, "An error occurred. Please try again.", ""))
			return
		}

		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, token, "An error occurred. Please try again.", ""))
			return
		}

		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		_, err = driver.UpdateUser(r.Context(), ac, db.UpdateUserParams{
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
			Render(w, r, pages.ResetPassword(csrfToken, utility.Version, token, "An error occurred. Please try again.", ""))
			return
		}

		// Revoke the used token.
		_, revokeErr := driver.UpdateToken(r.Context(), ac, db.UpdateTokenParams{
			ID:        tok.ID,
			Token:     tok.Token,
			IssuedAt:  tok.IssuedAt,
			ExpiresAt: tok.ExpiresAt,
			Revoked:   true,
		})
		if revokeErr != nil {
			utility.DefaultLogger.Warn("failed to revoke used password reset token", revokeErr)
		}

		Render(w, r, pages.ResetPassword(csrfToken, utility.Version, "", "", "Your password has been reset. You can now sign in."))
	}
}
