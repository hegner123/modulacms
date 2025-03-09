package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	config "github.com/hegner123/modulacms/internal/Config"
)

type TokenType string

const (
	REFRESHTOKEN TokenType = "refresh_token"
	ACCESSTOKEN  TokenType = "access_token"
)

func GenerateAccessToken(userID int64, sessionID int64, c config.Config) (string, error) {
	jwtSecret := []byte(c.Token_Secret)
	// Define token expiration time
	expirationTime := time.Now().Add(15 * time.Minute)


	// Create the JWT claims. Standard claims include "exp" (expiration time) and "iat" (issued at).
	// Additional claims can be added as needed.
	claims := jwt.MapClaims{
		"type":       "access",
		"session_id": sessionID,
		"user_id":    userID,
		"exp":        expirationTime.Unix(),
		"iat":        time.Now().Unix(),
	}

	// Create token with the specified signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key and return it as a string
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID int64, sessionID int64, c config.Config) (string, error) {
	jwtSecret := []byte(c.Token_Secret)
	// Define token expiration time
	expirationTime := time.Now().Add(192 * time.Hour)

	// Create the JWT claims. Standard claims include "exp" (expiration time) and "iat" (issued at).
	// Additional claims can be added as needed.
	claims := jwt.MapClaims{
		"type":       REFRESHTOKEN,
		"session_id": sessionID,
		"user_id":    userID,
		"exp":        expirationTime.Unix(),
		"iat":        time.Now().Unix(),
	}

	// Create token with the specified signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key and return it as a string
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ReadToken(tokenString string, c config.Config) (jwt.MapClaims, error) {
	jwtSecret := []byte(c.Token_Secret)
	// Parse the token using the jwtSecret.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Verify that the token's signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	// Assert that the token's claims are of type jwt.MapClaims and verify token validity.
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
