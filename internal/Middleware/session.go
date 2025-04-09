package middleware

import (
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type MiddlewareCookie struct {
	Session string `json:"session"`
	UserId  int64  `json:"userId"`
}

func UserIsAuth(r *http.Request, cookie *http.Cookie, conf config.Config) (*db.Users, error) {
	userCookie, err := ReadCookie(cookie)
	if err != nil {
		return nil, err
	}

	dbc := db.ConfigDB(conf)

	utility.DefaultLogger.Info("userCookie ID %v\n", userCookie.UserId)

	session, err := dbc.GetSessionByUserId(userCookie.UserId)
	utility.DefaultLogger.Info("", session)
	if err != nil || session == nil {
		utility.DefaultLogger.Error("Error retrieving session or no sessions found:", err)
		return nil, err
	}
	if userCookie.Session != session.SessionData.String {
		err := fmt.Errorf("sessions don't match")
		utility.DefaultLogger.Warn("", err)
		return nil, err
	}

	expired := utility.TimestampLessThan(session.ExpiresAt.String)
	if expired {
		err := fmt.Errorf("session is expired")
		return nil, err
	}

	u, err := dbc.GetUser(userCookie.UserId)
	if err != nil {
		return nil, err
	}
	return u, nil
}
