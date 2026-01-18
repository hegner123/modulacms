package middleware

import (
	"github.com/charmbracelet/ssh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// PublicKeyHandler is the SSH public key authentication callback.
// This validates that the key is structurally valid and allows the connection through.
// Actual authentication and authorization happens in the middleware chain.
func PublicKeyHandler(c *config.Config) func(ssh.Context, ssh.PublicKey) bool {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		// Log the attempt
		fingerprint := FingerprintSHA256(key)
		utility.DefaultLogger.Finfo("Public key presented: %s (type: %s)", fingerprint, key.Type())

		// Allow all valid keys through - authentication happens in middleware
		return true
	}
}
