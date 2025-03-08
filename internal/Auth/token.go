package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

func GenerateToken(secretKey []byte) (string, error) {
	// Define token claims. You can include custom claims as needed.
	claims := jwt.MapClaims{
		"sub": "user_id",                             // subject: e.g., user identifier
		"exp": time.Now().Add(72 * time.Hour).Unix(), // expiration time (72 hours)
		"iat": time.Now().Unix(),                     // issued at time
		// add any other claims you need here
	}

	// Create a new token with the specified claims and signing method.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key.
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
