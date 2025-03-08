package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	middleware "github.com/hegner123/modulacms/internal/Middleware"
	utility "github.com/hegner123/modulacms/internal/Utility"
	"golang.org/x/oauth2"
)

func HandleAuth(conf config.Config, form url.Values) bool {
	dbc := db.ConfigDB(conf)

	user, err := dbc.GetUser(1)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	requestHash := AuthMakeHash(form.Get("hash"), "modulacms")
	return compareHashes(user.Hash, requestHash)
}

func AuthMakeHash(data, salt string) string {
	input := data + salt
	hash := sha256.Sum256([]byte(input))

	return hex.EncodeToString(hash[:])
}

func compareHashes(hash1, hash2 string) bool {
	hash1Bytes, err1 := hex.DecodeString(hash1)
	hash2Bytes, err2 := hex.DecodeString(hash2)

	if err1 != nil || err2 != nil || len(hash1Bytes) != len(hash2Bytes) {
		return false
	}
	return subtle.ConstantTimeCompare(hash1Bytes, hash2Bytes) == 1
}

func AuthMiddleware(next http.Handler, conf config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the auth cookie exists
		cookie, err := r.Cookie("modula_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		_, err = middleware.UserIsAuth(r, conf)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Proceed to the next handler if authorized
		next.ServeHTTP(w, r)

	})
}

var Verifier string

func OauthSettings(c config.Config) {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     c.Oauth_Client_Id,
		ClientSecret: c.Oauth_Client_Secret,
		Scopes:       c.Oauth_Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.Oauth_Endpoint[config.OauthAuthURL],
			TokenURL: c.Oauth_Endpoint[config.OauthAuthURL],
		},
	}
	Verifier := oauth2.GenerateVerifier()

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(Verifier))
	fmt.Printf("Visit the URL for the auth dialog:\n %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(Verifier))
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	_, err = client.Get("...")
	if err != nil {
		utility.LogError("failed to : ", err)
	}
}
