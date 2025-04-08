package middleware

import (
	"fmt"
	"net/http"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	utility "github.com/hegner123/modulacms/internal/utility"
)

type MiddlewareCookie struct {
	Token  string `json:"token"`
	UserId int64  `json:"userId"`
}

func UserIsAuth(r *http.Request, conf config.Config) (*db.Users, error) {

	// Retrieve the cookie
	c, err := r.Cookie(conf.Cookie_Name)
	if err != nil {
		utility.DefaultLogger.Error("Error retrieving cookie:", err)
		return nil, err
	}

	// Read and parse cookie
	userCookie, err := ReadCookie(c)
	if err != nil {
		return nil, err
	}

	// Get the database instance
	dbc := db.ConfigDB(conf)

	utility.DefaultLogger.Info("userCookie ID %v\n", userCookie.UserId)

	// Retrieve tokens from the database
	tokens, err := dbc.GetTokenByUserId(userCookie.UserId)
	utility.DefaultLogger.Info("", tokens)
	if err != nil || tokens == nil || len(*tokens) == 0 {
		utility.DefaultLogger.Error("Error retrieving tokens or no tokens found:", err)
		return nil, err
	}

	// Find the Access token
	var accessToken *db.Tokens
	for _, t := range *tokens {
		if t.TokenType == "Access" {
			accessToken = &t
			break
		}
	}

	// Ensure we have a valid access token
	if accessToken == nil {
		err := fmt.Errorf("no valid access token found")
		utility.DefaultLogger.Warn("", err)
		return nil, err
	}

	// Compare tokens
	if userCookie.Token != accessToken.Token {
		err := fmt.Errorf("tokens don't match")
		utility.DefaultLogger.Warn("", err)
		return nil, err
	}
	utility.DefaultLogger.Info("Tokens match")

	// Check if token is revoked
	if accessToken.Revoked {
		utility.DefaultLogger.Info("Token revoked")
		return nil, err
	}

	// Check if token is expired
	expired := utility.TimestampLessThan(accessToken.ExpiresAt)
	if expired {
		err := fmt.Errorf("token is expired")
		return nil, err
	}

	// If everything is ok return user
	u, err := dbc.GetUser(userCookie.UserId)
	if err != nil {
		return nil, err
	}
	return u, nil
}
