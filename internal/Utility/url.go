package utility

import (
	"net/http"
	"net/url"
)

func ReadCookie(c *http.Cookie) string {
	err := c.Valid()
	if err != nil {
		return c.Valid().Error()
	}
	cv := c.Value
	v, err := url.QueryUnescape(cv)
	if err != nil {
		return err.Error()
	}
	return v
}
