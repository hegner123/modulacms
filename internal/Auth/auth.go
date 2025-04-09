package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)


// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	// Use cost of 12 (which is a good balance of security and performance)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// Legacy SHA-256 hash function - kept for backward compatibility
func AuthMakeHash(data, salt string) string {
	// This is the legacy hash function, kept for backward compatibility
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

var Verifier string

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

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
	utility.DefaultLogger.Info("Visit the URL for the auth dialog:\n %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
        utility.DefaultLogger.Error("",err)
	}
	tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(Verifier))
	if err != nil {
        utility.DefaultLogger.Error("",err)
	}

	client := conf.Client(ctx, tok)
	_, err = client.Get("...")
	if err != nil {
		utility.DefaultLogger.Error("failed to", err)
	}
}
