# auth

OAuth authentication, user provisioning, token management, and password hashing for ModulaCMS.

## Overview

The auth package provides OAuth 2.0 integration with automatic token refresh, user provisioning from OAuth providers, CSRF protection via state parameters, PKCE verifier management, and bcrypt password hashing. It supports multiple OAuth providers including GitHub and OpenID Connect-compliant services. The package uses in-memory stores for state and PKCE verifiers with automatic cleanup of expired entries.

## Constants

```go
const defaultRoleLabel = "viewer"
const cleanupThreshold = 100
const verifierTTL = 20 * time.Minute
```

defaultRoleLabel is the role assigned to newly provisioned OAuth users. cleanupThreshold determines how many GenerateState or StoreVerifier calls trigger an automatic cleanup sweep. verifierTTL sets the lifetime for PKCE verifiers at 20 minutes.

## Types

### type Logger

```go
type Logger interface {
    Debug(message string, args ...any)
    Info(message string, args ...any)
    Warn(message string, err error, args ...any)
    Error(message string, err error, args ...any)
}
```

Logger is the logging interface consumed by TokenRefresher and UserProvisioner. Callers pass a concrete logger such as utility.Logger into NewTokenRefresher and NewUserProvisioner via their constructors. The interface requires Debug, Info, Warn, and Error methods with variadic arguments for formatted logging.

### type StateStore

```go
type StateStore struct {
    mu             sync.RWMutex
    states         map[string]time.Time
    cleanupCounter int
}
```

StateStore manages OAuth state parameters for CSRF protection. States are stored in-memory with automatic cleanup of expired entries. The global instance globalStateStore is used by GenerateState and ValidateState package functions. States expire after 20 minutes and are deleted immediately after successful validation for one-time use.

#### func (s *StateStore) Size

`func (s *StateStore) Size() int`

Size returns the current number of active states. This is primarily useful for testing and monitoring state storage. Thread-safe for concurrent access.

#### func (s *StateStore) Clear

`func (s *StateStore) Clear()`

Clear removes all states from the store and resets the cleanup counter. This should only be used in testing to reset state between test cases.

### type VerifierStore

```go
type VerifierStore struct {
    mu             sync.RWMutex
    verifiers      map[string]verifierEntry
    cleanupCounter int
}
```

VerifierStore manages PKCE verifiers for OAuth flows. Verifiers are stored in-memory and associated with state parameters for retrieval during token exchange. The global instance globalVerifierStore is used by StoreVerifier and GetVerifier package functions.

#### func (v *VerifierStore) Clear

`func (v *VerifierStore) Clear()`

Clear removes all verifiers from the store and resets the cleanup counter. This should only be used in testing to reset verifier state between test cases.

### type UserInfo

```go
type UserInfo struct {
    ProviderUserID string `json:"sub"`
    ID             int64  `json:"id"`
    Email          string `json:"email"`
    Name           string `json:"name"`
    Username       string `json:"preferred_username"`
    Login          string `json:"login"`
    AvatarURL      string `json:"avatar_url"`
}
```

UserInfo represents standardized user information retrieved from OAuth providers. This struct is provider-agnostic and maps common fields from various OAuth providers. ProviderUserID holds the OpenID Connect sub claim. ID holds GitHub numeric user ID. Username holds preferred_username for OIDC or login for GitHub. The struct handles provider-specific field mapping automatically.

### type GitHubEmail

```go
type GitHubEmail struct {
    Email      string `json:"email"`
    Primary    bool   `json:"primary"`
    Verified   bool   `json:"verified"`
    Visibility string `json:"visibility"`
}
```

GitHubEmail represents an email from GitHub's user emails endpoint. Used to fetch primary verified email addresses when the userinfo endpoint does not include email.

### type TokenRefresher

```go
type TokenRefresher struct {
    config *config.Config
    driver db.DbDriver
    log    Logger
}
```

TokenRefresher handles automatic refreshing of OAuth access tokens before expiration. It checks token expiration times and refreshes them using the OAuth provider's token endpoint when they are close to expiring. Tokens are refreshed if they expire within 5 minutes.

#### func NewTokenRefresher

`func NewTokenRefresher(log Logger, c *config.Config, driver db.DbDriver) *TokenRefresher`

NewTokenRefresher creates a new TokenRefresher with the given configuration, logger, and database driver. The logger is used for structured logging during token refresh operations.

```go
tr := NewTokenRefresher(logger, cfg, dbDriver)
err := tr.RefreshIfNeeded(userID)
```

#### func (tr *TokenRefresher) RefreshIfNeeded

`func (tr *TokenRefresher) RefreshIfNeeded(userID types.UserID) error`

RefreshIfNeeded checks if a user's OAuth token needs refreshing and refreshes it if necessary. Returns nil if refresh was successful or not needed. Returns an error if the user does not have OAuth configured or if refresh fails.

Tokens are refreshed if they expire within 5 minutes. Long-lived tokens without expiry such as GitHub personal access tokens are not refreshed. Returns nil if the user does not use OAuth authentication.

Returns an error if token parsing fails or the OAuth provider token endpoint returns an error.

### type UserProvisioner

```go
type UserProvisioner struct {
    config *config.Config
    driver db.DbDriver
    log    Logger
}
```

UserProvisioner handles user creation and OAuth account linking. It provides a unified interface for provisioning users from various OAuth providers including GitHub and OpenID Connect services.

#### func NewUserProvisioner

`func NewUserProvisioner(log Logger, c *config.Config, driver db.DbDriver) *UserProvisioner`

NewUserProvisioner creates a new UserProvisioner with the given configuration, logger, and database driver. The logger is used for structured logging during user provisioning.

```go
up := NewUserProvisioner(logger, cfg, dbDriver)
user, err := up.ProvisionUser(userInfo, token, "github")
```

#### func (up *UserProvisioner) FetchUserInfo

`func (up *UserProvisioner) FetchUserInfo(client *http.Client) (*UserInfo, error)`

FetchUserInfo retrieves user information from the OAuth provider's userinfo endpoint. It uses the authenticated HTTP client to make the request and returns standardized UserInfo. The client must be configured with a valid OAuth access token.

Handles provider-specific field mapping. GitHub uses login instead of preferred_username and numeric id instead of sub. If email is missing, attempts to fetch from GitHub's user emails endpoint.

Returns an error if the userinfo URL is not configured, the HTTP request fails, the response status is not 200 OK, JSON decoding fails, or email is not provided by the provider.

#### func (up *UserProvisioner) ProvisionUser

`func (up *UserProvisioner) ProvisionUser(userInfo *UserInfo, token *oauth2.Token, provider string) (*db.Users, error)`

ProvisionUser creates a new user or links OAuth to an existing user. Returns the user record after provisioning.

Provisioning logic follows these steps:
1. Check if OAuth provider and user ID already exist. If found, update tokens and return user.
2. Check if email already exists. If found, link OAuth to existing user and return user.
3. Create new user with default viewer role, link OAuth, and return user.

Returns an error if email is required but not provided, the viewer role cannot be found, user creation fails, or OAuth linking fails.

```go
user, err := up.ProvisionUser(userInfo, oauthToken, "github")
if err != nil {
    return fmt.Errorf("provisioning failed: %w", err)
}
```

## Password Hashing

### func HashPassword

`func HashPassword(password string) (string, error)`

HashPassword creates a bcrypt hash of the password using cost factor 12. Returns the hash string or an error if the password is empty or exceeds 72 bytes.

Returns an error if password is empty or if password length exceeds 72 bytes. The 72-byte limit is a bcrypt requirement.

```go
hash, err := HashPassword("userpassword123")
if err != nil {
    return fmt.Errorf("hash failed: %w", err)
}
```

### func CheckPasswordHash

`func CheckPasswordHash(password, hash string) bool`

CheckPasswordHash compares a plaintext password with a bcrypt hash. Returns true if the password matches the hash, false otherwise. Never returns an error, always returns a boolean.

```go
if CheckPasswordHash(inputPassword, storedHash) {
    // password is correct
}
```

## OAuth State Management

### func GenerateState

`func GenerateState() (string, error)`

GenerateState creates a new cryptographically secure state parameter using 32 random bytes encoded as base64 URL encoding. The state is valid for 20 minutes and can only be used once.

Returns the state string or an error if random generation fails. Automatically triggers cleanup of expired states every 100 calls.

```go
state, err := GenerateState()
if err != nil {
    return fmt.Errorf("state generation failed: %w", err)
}
```

### func ValidateState

`func ValidateState(state string) error`

ValidateState verifies that a state parameter is valid and not expired. States can only be used once and are deleted after successful validation.

Returns an error if the state parameter is empty, invalid, expired, or already used. This function must be called during the OAuth callback to prevent CSRF attacks.

```go
if err := ValidateState(callbackState); err != nil {
    return fmt.Errorf("CSRF check failed: %w", err)
}
```

## PKCE Verifier Management

### func StoreVerifier

`func StoreVerifier(state, verifier string)`

StoreVerifier associates a PKCE verifier with a state parameter. The verifier can be retrieved later using the state as a key during token exchange. Verifiers expire after 20 minutes. Automatically triggers cleanup of expired verifiers every 100 calls.

```go
StoreVerifier(state, pkceVerifier)
```

### func GetVerifier

`func GetVerifier(state string) (string, error)`

GetVerifier retrieves the PKCE verifier associated with a state parameter. The verifier is deleted after retrieval for one-time use.

Returns an error if the state is not found or the verifier has expired. Always deletes the verifier entry after retrieval regardless of expiration status.

```go
verifier, err := GetVerifier(callbackState)
if err != nil {
    return fmt.Errorf("verifier not found: %w", err)
}
```
