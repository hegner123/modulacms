package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

type MiddlewareCookie struct {
	Token  string `json:"token"`
	UserId int64  `json:"userId"`
}

func ScanCookie(i []byte) (*MiddlewareCookie, error) {
	c := MiddlewareCookie{}
	err := json.Unmarshal(i, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil

}
func UserIsAuth(r *http.Request, dbSrc string) bool {
	var buf bytes.Buffer

	// Retrieve the cookie
	c, err := r.Cookie("modula_token")
	if err != nil {
		fmt.Println("Error retrieving cookie:", err)
		return false
	}

	// Read and parse cookie
	rc := utility.ReadCookie(c)
	_, err = buf.WriteString(rc)
	if err != nil {
		fmt.Println("Error writing cookie data:", err)
		return false
	}

	userCookie, err := ScanCookie(buf.Bytes())
	if err != nil {
		fmt.Println("Error scanning cookie:", err)
		return false
	}

	// Get the database instance
	var dbc db.Database
	if dbSrc != "" {
		dbc = db.GetDb(db.Database{Src: dbSrc})
	} else {
		dbc = db.GetDb(db.Database{})
	}

	// Ensure the database connection is valid
	if dbc.Connection == nil {
		fmt.Println("Database connection is nil")
		return false
	}

	// Retrieve tokens from the database
	tokens, err := db.GetTokenByUserId(dbc.Connection, dbc.Context, userCookie.UserId)
	if err != nil || tokens == nil || len(*tokens) == 0 {
		fmt.Println("Error retrieving tokens or no tokens found:", err)
		return false
	}

	// Find the Access token
	var accessToken *mdb.Tokens
	for _, t := range *tokens {
		if t.TokenType == "Access" {
			accessToken = &t
			break
		}
	}

	// Ensure we have a valid access token
	if accessToken == nil {
		fmt.Println("No valid Access token found")
		return false
	}

	// Compare tokens
	if userCookie.Token != accessToken.Token {
		fmt.Println("Tokens don't match")
		return false
	}

	// Check if token is revoked
	if accessToken.Revoked.Bool {
		fmt.Println("Token revoked")
		return false
	}

	// Check if token is expired
	expired := utility.TimestampLessThan(accessToken.ExpiresAt)
    return !expired
}
