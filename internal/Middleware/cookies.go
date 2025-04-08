package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	utility "github.com/hegner123/modulacms/internal/utility"
)

func SetCookieHandler(w http.ResponseWriter, c *http.Cookie) {

	basic := []byte("Test")
	// Set the cookie in the response header
	http.SetCookie(w, c)
	utility.DefaultLogger.Debug("", w.Header())
	i, err := w.Write(basic)
	if err != nil {
		return
	}
    utility.DefaultLogger.Debug("wrote %d bytes\n", i)
    utility.DefaultLogger.Fdebug("Cook has been set!", w)
}

func ReadCookie(c *http.Cookie) (*MiddlewareCookie, error) {
	k := MiddlewareCookie{}

	err := c.Valid()
	if err != nil {
		return nil, err
	}
	cv := c.Value
	b, err := base64.StdEncoding.DecodeString(cv)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &k)
	if err != nil {
		return nil, err
	}

	return &k, nil
}
