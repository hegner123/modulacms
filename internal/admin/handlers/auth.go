package handlers

import (
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
