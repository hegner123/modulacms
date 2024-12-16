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

	"golang.org/x/oauth2"
)

func handleAuth(form url.Values) {
	db, ctx, err := getDb(Database{ })
	if err != nil {
		logError("failed to : ", err)

	}
	user := dbGetUserByEmail(db, ctx, form.Get("email"))
	requestHash := authMakeHash(form.Get("hash"), "modulacms")
	if compareHashes(user.Hash, requestHash) {
	}
}

func authMakeHash(data, salt string) string {
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

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the auth cookie exists
		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate the token (in a real app, use a secure method like JWT validation)
		if cookie.Value != "valid_token_example" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Proceed to the next handler if authorized
		next.ServeHTTP(w, r)
	})
}

func oauthSettings() {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		Scopes:       []string{"SCOPE1", "SCOPE2"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://provider.com/o/oauth2/auth",
			TokenURL: "https://provider.com/o/oauth2/token",
		},
	}
	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	client.Get("...")
}
