package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type authcontext string

func Serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Cors(w, r)

		u, user := AuthRequest(w, r)
		if u != nil {
			// Inject authenticated user information into the request context for downstream handlers
			ctx := context.WithValue(r.Context(), u, user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if strings.Contains(r.URL.Path, "api") {
			w.WriteHeader(http.StatusUnauthorized)
			msg := fmt.Sprintf("Unauthorized Request to %s", string(r.URL.Path))
			_, err := w.Write([]byte(msg))
			if err != nil {
				utility.DefaultLogger.Error("", err)
				return
			}
			return
		}

		next.ServeHTTP(w, r)

	})
}

func AuthRequest(w http.ResponseWriter, r *http.Request) (*authcontext, *db.Users) {
    if strings.Contains(r.URL.Path, "favicon.ico"){
        return nil, nil
    }
	var u authcontext = "authenticated"
	c := config.Env
	cookie, err := r.Cookie(c.Cookie_Name)
	if err != nil {
		utility.DefaultLogger.Info("cookie not found", err)
		return nil, nil
	}

	user, err := UserIsAuth(r, cookie, c)
	if err != nil {
		return nil, nil
	}
	return &u, user
}

func GetURLSegments(path string) []string {
	return strings.Split(path, "/")
}

/*
func refreshTokenIfNeeded(t string) (*db.Users, error) {
	u := db.Users{
		Email: t,
	}

	return &u, nil

}
*/
