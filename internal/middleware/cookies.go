package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
	utility "github.com/hegner123/modulacms/internal/utility"
)

// SetCookieHandler sets a cookie in the HTTP response and writes a basic response body.
// It logs the headers and bytes written for debugging purposes.
func SetCookieHandler(w http.ResponseWriter, c *http.Cookie) {
	basic := []byte("Test")
	// Set the cookie in the response header
	http.SetCookie(w, c)
	utility.DefaultLogger.Fdebug("", w.Header())
	i, err := w.Write(basic)
	if err != nil {
		return
	}
    utility.DefaultLogger.Fdebug("wrote %d bytes\n", i)
    utility.DefaultLogger.Fdebug("Cook has been set!", w)
}

// ReadCookie decodes and deserializes a cookie value into a MiddlewareCookie struct.
// It validates the cookie, base64 decodes its value, and unmarshals the JSON data.
// Returns an error if any step in the process fails.
func ReadCookie(c *http.Cookie) (*MiddlewareCookie, error) {
	k := MiddlewareCookie{}

	err := c.Valid()
	if err != nil {
		return nil, err
	}
	cv := c.Value
	b, err := base64.StdEncoding.DecodeString(cv)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &k)
	if err != nil {
		return nil, err
	}

	return &k, nil
}

// WriteCookie creates and sets a secure authentication cookie with proper security flags.
// It encodes the session data and user ID as base64-encoded JSON and applies security
// settings from the configuration (HttpOnly, Secure, SameSite).
// Returns an error if encoding or cookie creation fails.
func WriteCookie(w http.ResponseWriter, c *config.Config, sessionData string, userId types.UserID) error {
	cookie := MiddlewareCookie{
		Session: sessionData,
		UserId:  userId,
	}

	jsonData, err := json.Marshal(cookie)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)

	// Parse SameSite mode from config
	sameSite := http.SameSiteLaxMode // Default to Lax
	if c.Cookie_SameSite == "strict" {
		sameSite = http.SameSiteStrictMode
	} else if c.Cookie_SameSite == "none" {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     c.Cookie_Name,
		Value:    encoded,
		Path:     "/",
		MaxAge:   86400,             // 24 hours
		HttpOnly: true,              // Prevent JavaScript access
		Secure:   c.Cookie_Secure,   // HTTPS only (from config)
		SameSite: sameSite,          // CSRF protection (from config)
	})

	utility.DefaultLogger.Finfo("Secure cookie set for user %d", userId)
	return nil
}
