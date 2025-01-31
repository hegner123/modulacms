package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

func TestTokenValidation(t *testing.T) {
	database := "get_tests.db"

	dbc := db.GetDb(db.Database{Src: database})
	if dbc.Connection == nil {
		t.Fatal("Database connection is nil")
	}

	// Generate expiration time
	newExpiresTime, _ := utility.TokenExpiredTime()

	// Update Refresh token
	_, err := db.UpdateToken(dbc.Connection, dbc.Context, mdb.UpdateTokenParams{
		Token:     "c958c8ab-6ede-4c6b-be2c-759f39a8e8aa",
		IssuedAt:  utility.TimestampS(),
		ExpiresAt: newExpiresTime,
		Revoked:   db.Nb(false),
		ID:        1,
	})
	if err != nil {
		t.Fatalf("Failed to update Refresh token: %v", err)
	}

	// Update Access token
	_, err = db.UpdateToken(dbc.Connection, dbc.Context, mdb.UpdateTokenParams{
		Token:     "e6082ffa-7a47-4bc1-8c69-46ca4f6276ab",
		IssuedAt:  utility.TimestampS(),
		ExpiresAt: newExpiresTime,
		Revoked:   db.Nb(false),
		ID:        2,
	})
	if err != nil {
		t.Fatalf("Failed to update Access token: %v", err)
	}

	// Create a test request
	r := httptest.NewRequest("GET", "https://modulacms.com", nil)

	// Create a test cookie (raw JSON encoding, as expected)
	ct := MiddlewareCookie{Token: "e6082ffa-7a47-4bc1-8c69-46ca4f6276ab", UserId: 1}
	cj, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Set cookie
	cookie := &http.Cookie{
		Name:  "modula_token",
		Value: url.QueryEscape(string(cj)), // Retaining QueryEscape as per your original code
	}
	r.AddCookie(cookie)

	// Validate the cookie retrieval
	_, err = r.Cookie("modula_token")
	if err != nil {
		t.Fatalf("Failed to retrieve cookie: %v", err)
	}

	// Run authentication check
	res := UserIsAuth(r, database)
	if !res {
		t.Fatal("Token validation failed: expected true, got false")
	}

}
