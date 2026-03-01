package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign computes an HMAC-SHA256 signature of payload using the given secret.
// Returns a hex-encoded string.
func Sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks that signature matches the HMAC-SHA256 of payload for the given secret.
func Verify(secret, signature string, payload []byte) bool {
	expected := Sign(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}
