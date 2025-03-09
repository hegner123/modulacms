package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

func SetCookieHandler(w http.ResponseWriter, c *http.Cookie) {
	/*
				expiration := time.Now().Add(24 * time.Hour)
		        // Create a new cookie
		        cookie := http.Cookie{
		                Name:     "myCookie",
		                Value:    "cookieValue",
		                Expires:  expiration,
		                Path:     "/",       // Cookie is accessible on all paths
		                HttpOnly: true,      // Not accessible via JavaScript
		                Secure:   false,     // Set true if using HTTPS
		            }
	*/

	basic := []byte("Test")
	// Set the cookie in the response header
	http.SetCookie(w, c)
	fmt.Println(w.Header())
	i, err := w.Write(basic)
	if err != nil {
		return
	}
	fmt.Printf("wrote %d bytes\n", i)

	fmt.Fprintln(w, "Cookie has been set!")
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
