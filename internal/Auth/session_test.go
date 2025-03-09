package auth

import (
	"fmt"
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)

func TestCreateSessionTokens(t *testing.T) {
	verbose := false
	c := config.LoadConfig(&verbose, "")
    pk, err :=CreateSessionTokens(1, c)
    if err!=nil {
        t.Fatal(err)
    }
    fmt.Printf("Access Token:\n%v\n",pk.AccessToken)
    fmt.Printf("Refresh Token:\n%v\n",pk.AccessToken)

}
