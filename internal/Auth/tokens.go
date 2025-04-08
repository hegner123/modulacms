package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	config "github.com/hegner123/modulacms/internal/config"
)

type TokenType string

const (
	REFRESHTOKEN TokenType = "refresh_token"
	ACCESSTOKEN  TokenType = "access_token"
)

func GenerateAccessToken(userID int64, sessionID int64, c config.Config) (string, error) {
	if c.Token_Secret == "" {
		return "", fmt.Errorf("token generation failed: missing token secret")
	}

	jwtSecret := []byte(c.Token_Secret)
	
	// Define token expiration time
	expirationTime := time.Now().Add(15 * time.Minute)

	// Create the JWT claims
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
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID int64, sessionID int64, c config.Config) (string, error) {
	if c.Token_Secret == "" {
		return "", fmt.Errorf("token generation failed: missing token secret")
	}

	jwtSecret := []byte(c.Token_Secret)
	
	// Define token expiration time (8 days)
	expirationTime := time.Now().Add(192 * time.Hour)

	// Create the JWT claims
	claims := jwt.MapClaims{
		"type":       REFRESHTOKEN,
		"session_id": sessionID,
		"user_id":    userID,
		"exp":        expirationTime.Unix(),
		"iat":        time.Now().Unix(),
	}

	// Create token with the specified signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

func ReadToken(tokenString string, c config.Config) (jwt.MapClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token validation failed: empty token provided")
	}
	
	if c.Token_Secret == "" {
		return nil, fmt.Errorf("token validation failed: missing token secret")
	}
	
	jwtSecret := []byte(c.Token_Secret)
	
	// Parse the token using the jwtSecret
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Verify that the token's signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("token validation failed: unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Assert that the token's claims are of type jwt.MapClaims and verify token validity
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("token validation failed: invalid or expired token")
}
