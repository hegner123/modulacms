package middleware

import (
	"context"
	"net/http"
	"strings"

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
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

		next.ServeHTTP(w, r)

	})
}

func AuthRequest(w http.ResponseWriter, r *http.Request) (*authcontext, *db.Users) {
	var u authcontext = "authenticated"
	c := config.Env

	user, err := UserIsAuth(r, c)
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
