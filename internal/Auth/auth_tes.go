package auth

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
)

func TestOauthSettings(t *testing.T){
    verboseFlag := false
    c:=config.LoadConfig(&verboseFlag,"")
    OauthSettings(c)

}
