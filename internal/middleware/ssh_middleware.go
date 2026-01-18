package middleware

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// SSHAuthenticationMiddleware validates SSH keys and populates session context.
// This should run early in the middleware chain to set up authentication state.
func SSHAuthenticationMiddleware(c *config.Config) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			ctx := s.Context()

			// Get the public key from the session
			if s.PublicKey() == nil {
				utility.DefaultLogger.Fwarn("No public key provided in SSH session", nil)
				next(s)
				return
			}

			fingerprint := FingerprintSHA256(s.PublicKey())
			utility.DefaultLogger.Finfo("SSH auth attempt with fingerprint: %s", fingerprint)

			// Look up user by SSH key fingerprint
			dbc := db.ConfigDB(*c)
			user, err := dbc.GetUserBySSHFingerprint(fingerprint)

			if err != nil {
				// Key not registered - mark for provisioning
				utility.DefaultLogger.Finfo("SSH key not registered, needs provisioning. Fingerprint: %s", fingerprint)
				ctx.SetValue("needs_provisioning", true)
				ctx.SetValue("ssh_fingerprint", fingerprint)
				ctx.SetValue("ssh_key_type", s.PublicKey().Type())
				ctx.SetValue("ssh_public_key", string(s.PublicKey().Marshal()))
				ctx.SetValue("authenticated", false)
				next(s)
				return
			}

			// User found - update last used timestamp
			sshKey, err := dbc.GetUserSshKeyByFingerprint(fingerprint)
			if err == nil && sshKey != nil {
				lastUsed := time.Now().Format(time.RFC3339)
				_ = dbc.UpdateUserSshKeyLastUsed(sshKey.SshKeyID, lastUsed)
			}

			// Store user in context
			ctx.SetValue("user_id", user.UserID)
			ctx.SetValue("user", user)
			ctx.SetValue("authenticated", true)
			ctx.SetValue("needs_provisioning", false)

			utility.DefaultLogger.Finfo("SSH authentication successful for user: %s (user_id: %d)", user.Email, user.UserID)
			next(s)
		}
	}
}

// SSHAuthorizationMiddleware ensures the user is authenticated before proceeding.
// This can be used to protect specific endpoints or require authentication.
func SSHAuthorizationMiddleware(c *config.Config) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			ctx := s.Context()

			// Check if user needs provisioning - allow through for provisioning flow
			if needsProvisioning, ok := ctx.Value("needs_provisioning").(bool); ok && needsProvisioning {
				utility.DefaultLogger.Finfo("User needs provisioning, allowing through")
				next(s)
				return
			}

			// Check authentication
			if authenticated, ok := ctx.Value("authenticated").(bool); !ok || !authenticated {
				utility.DefaultLogger.Fwarn("Unauthorized SSH access attempt", nil)
				wish.Println(s, "Authentication required")
				return
			}

			next(s)
		}
	}
}

// SSHRateLimitMiddleware limits connection attempts per IP.
// This prevents brute force attacks on SSH keys.
func SSHRateLimitMiddleware(c *config.Config) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			// TODO: Implement rate limiting logic
			// For now, just pass through
			next(s)
		}
	}
}

// SSHSessionLoggingMiddleware logs SSH session details.
func SSHSessionLoggingMiddleware(c *config.Config) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			remoteAddr := s.RemoteAddr().String()
			user := s.User()

			utility.DefaultLogger.Finfo("SSH session started", "user", user, "remote", remoteAddr)

			// Call next handler
			next(s)

			utility.DefaultLogger.Finfo("SSH session ended", "user", user, "remote", remoteAddr)
		}
	}
}

// FingerprintSHA256 generates a SHA256 fingerprint from an SSH public key.
// This matches the format used by modern SSH clients (SHA256:...).
func FingerprintSHA256(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	b64hash := base64.StdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf("SHA256:%s", b64hash)
}

// ParseSSHPublicKey parses an SSH public key string (e.g., from authorized_keys format)
// and returns the key type and fingerprint.
func ParseSSHPublicKey(publicKeyStr string) (keyType string, fingerprint string, err error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyStr))
	if err != nil {
		return "", "", fmt.Errorf("failed to parse SSH public key: %w", err)
	}

	keyType = pubKey.Type()
	fingerprint = FingerprintSHA256(pubKey)
	return keyType, fingerprint, nil
}
