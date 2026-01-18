package middleware

import (
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// MiddlewareCookie represents the structure of authentication cookies used by the middleware.
// It contains session identifier and user ID information.
type MiddlewareCookie struct {
	Session string `json:"session"`
	UserId  int64  `json:"userId"`
}

// UserIsAuth validates a user's authentication status based on the provided cookie.
// It verifies that the session in the cookie matches the one in the database,
// checks if the session is still valid (not expired), and retrieves the user data.
// Returns the user object if authentication is successful, or an error if any validation fails.
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error) {
	userCookie, err := ReadCookie(cookie)
	if err != nil {
		return nil, err
	}

	dbc := db.ConfigDB(*c)

	utility.DefaultLogger.Fdebug("userCookie ID %v\n", userCookie.UserId)

	session, err := dbc.GetSessionByUserId(userCookie.UserId)
	utility.DefaultLogger.Fdebug("", session)
	if err != nil || session == nil {
		utility.DefaultLogger.Ferror("Error retrieving session or no sessions found:", err)
		return nil, err
	}
	utility.DefaultLogger.Fdebug("Cookie session: %s", userCookie.Session)
	utility.DefaultLogger.Fdebug("DB session: %s", session.SessionData.String)
	if userCookie.Session != session.SessionData.String {
		err := fmt.Errorf("sessions don't match")
		utility.DefaultLogger.Fwarn("", err)
		return nil, err
	}

	expired := utility.TimestampLessThan(session.ExpiresAt.String)
	if expired {
		err := fmt.Errorf("session is expired")
		return nil, err
	}

	// Check and refresh OAuth tokens if needed
	refresher := auth.NewTokenRefresher(c)
	if err := refresher.RefreshIfNeeded(userCookie.UserId); err != nil {
		utility.DefaultLogger.Fwarn("Token refresh warning: %v", err)
		// Don't fail auth if refresh fails - token might still be valid
		// This is especially important for non-OAuth users
	}

	u, err := dbc.GetUser(userCookie.UserId)
	if err != nil {
		return nil, err
	}
	return u, nil
}
