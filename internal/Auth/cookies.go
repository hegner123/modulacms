package auth

import (
	"net/http"
	"time"

	config "github.com/hegner123/modulacms/internal/config"
)

func CreateAuthCookie(value string, c config.Config) (*http.Cookie, error) {
	var k http.Cookie
	d, err := time.ParseDuration(c.Cookie_Duration)
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(d)
	k.Name = c.Cookie_Name
	k.Expires = expires
	k.HttpOnly = true
	k.Value = value
	k.Path = "/"
	k.Secure = true
	k.SameSite = http.SameSiteNoneMode
	return &k, nil
}
