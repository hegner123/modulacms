package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log"
	"mime/multipart"

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
	"golang.org/x/oauth2"
)

func HandleAuthForm(conf config.Config, form *multipart.Form) (bool, *db.Users, error) {
	ue := form.Value["email"][0]
	up := form.Value["password"][0]
	if ue == "" || up == "" {
		err := fmt.Errorf("email: %s, password: %s\n", ue, up)
		return false, nil, err
	}
	dbc := db.ConfigDB(conf)

	user, err := dbc.GetUserByEmail(ue)
	if err != nil {
		utility.LogError("failed to : ", err)
        return false, nil, err
	}
	requestHash := AuthMakeHash(up, config.Env.Auth_Salt)
	hashMatch := compareHashes(user.Hash, requestHash)
	return hashMatch, user, nil
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
