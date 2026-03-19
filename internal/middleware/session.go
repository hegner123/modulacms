package middleware

import (
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// MiddlewareCookie represents the structure of authentication cookies used by the middleware.
// It contains session identifier and user ID information.
type MiddlewareCookie struct {
	Session string       `json:"session"`
	UserId  types.UserID `json:"userId"`
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

	session, err := dbc.GetSessionByToken(userCookie.Session)
	if err != nil || session == nil {
		utility.DefaultLogger.Ferror("Error retrieving session by token:", err)
		return nil, fmt.Errorf("session not found")
	}

	// Verify the session belongs to the cookie's user
	if session.UserID.String() != userCookie.UserId.String() {
		utility.DefaultLogger.Fwarn("session user mismatch: cookie=%s session=%s", fmt.Errorf("cookie=%s session=%s", userCookie.UserId, session.UserID))
		return nil, fmt.Errorf("session user mismatch")
	}

	expired := utility.TimestampLessThan(session.ExpiresAt.String())
	if expired {
		return nil, fmt.Errorf("session is expired")
	}

	// Check and refresh OAuth tokens if needed
	refresher := auth.NewTokenRefresher(utility.DefaultLogger, c, dbc)
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
