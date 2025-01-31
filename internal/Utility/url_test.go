package utility

import (
	"net/http"
	"testing"
)

func TestReadCookie(t *testing.T) {
	v := "%7B%22token%22%3A%22e6082ffa-7a47-4bc1-8c69-46ca4f6276ab%22%2C%22userId%22%3A1%7D"
	e := `{"token":"e6082ffa-7a47-4bc1-8c69-46ca4f6276ab","userId":1}`
	c := http.Cookie{Name: "testCookie", Value: v}
	rc := ReadCookie(&c)
	if rc != e {
		t.Fatal(rc, e)
	}

}
