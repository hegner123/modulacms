package remote

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// User: SDK <-> db
// ---------------------------------------------------------------------------

// userToDb converts a SDK User to a db Users.
// Note: SDK User omits the hash field; it is set to empty string.
func userToDb(s *modula.User) db.Users {
	return db.Users{
		UserID:       types.UserID(string(s.UserID)),
		Username:     s.Username,
		Name:         s.Name,
		Email:        types.Email(string(s.Email)),
		Hash:         "",
		Role:         s.Role,
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// userFromDb converts a db Users to a SDK User.
// Note: hash is not included in the SDK User type.
func userFromDb(d db.Users) modula.User {
	return modula.User{
		UserID:       modula.UserID(string(d.UserID)),
		Username:     d.Username,
		Name:         d.Name,
		Email:        modula.Email(string(d.Email)),
		Role:         d.Role,
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// userCreateFromDb converts db CreateUserParams to SDK CreateUserParams.
// Note: db stores hash, SDK sends password; the server handles hashing.
func userCreateFromDb(d db.CreateUserParams) modula.CreateUserParams {
	return modula.CreateUserParams{
		Username: d.Username,
		Name:     d.Name,
		Email:    modula.Email(string(d.Email)),
		Password: d.Hash, // db stores the hash; SDK sends it as password for the server to validate
		Role:     d.Role,
	}
}

// userUpdateFromDb converts db UpdateUserParams to SDK UpdateUserParams.
func userUpdateFromDb(d db.UpdateUserParams) modula.UpdateUserParams {
	return modula.UpdateUserParams{
		UserID:   modula.UserID(string(d.UserID)),
		Username: d.Username,
		Name:     d.Name,
		Email:    modula.Email(string(d.Email)),
		Password: d.Hash,
		Role:     d.Role,
	}
}

// ---------------------------------------------------------------------------
// Role: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// roleToDb converts a SDK Role to a db Roles.
func roleToDb(s *modula.Role) db.Roles {
	return db.Roles{
		RoleID:          types.RoleID(string(s.RoleID)),
		Label:           s.Label,
		SystemProtected: false, // SDK does not expose this field
	}
}

// roleFromDb converts a db Roles to a SDK Role.
func roleFromDb(d db.Roles) modula.Role {
	return modula.Role{
		RoleID: modula.RoleID(string(d.RoleID)),
		Label:  d.Label,
	}
}

// ---------------------------------------------------------------------------
// Permission: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// permissionToDb converts a SDK Permission to a db Permissions.
func permissionToDb(s *modula.Permission) db.Permissions {
	return db.Permissions{
		PermissionID:    types.PermissionID(string(s.PermissionID)),
		Label:           s.Label,
		SystemProtected: false, // SDK does not expose this field
	}
}

// permissionFromDb converts a db Permissions to a SDK Permission.
func permissionFromDb(d db.Permissions) modula.Permission {
	return modula.Permission{
		PermissionID: modula.PermissionID(string(d.PermissionID)),
		Label:        d.Label,
	}
}

// ---------------------------------------------------------------------------
// RolePermission: SDK <-> db (read-only)
// ---------------------------------------------------------------------------

// rolePermissionToDb converts a SDK RolePermission to a db RolePermissions.
func rolePermissionToDb(s *modula.RolePermission) db.RolePermissions {
	return db.RolePermissions{
		ID:           types.RolePermissionID(string(s.ID)),
		RoleID:       types.RoleID(string(s.RoleID)),
		PermissionID: types.PermissionID(string(s.PermissionID)),
	}
}

// rolePermissionFromDb converts a db RolePermissions to a SDK RolePermission.
func rolePermissionFromDb(d db.RolePermissions) modula.RolePermission {
	return modula.RolePermission{
		ID:           modula.RolePermissionID(string(d.ID)),
		RoleID:       modula.RoleID(string(d.RoleID)),
		PermissionID: modula.PermissionID(string(d.PermissionID)),
	}
}

// ---------------------------------------------------------------------------
// Token: SDK <-> db (write-only: create via remote)
// ---------------------------------------------------------------------------

// tokenToDb converts a SDK Token to a db Tokens.
func tokenToDb(s *modula.Token) db.Tokens {
	return db.Tokens{
		ID:        string(s.ID),
		UserID:    nullUserID(s.UserID),
		TokenType: s.TokenType,
		Token:     s.Token,
		IssuedAt:  types.Timestamp{}, // SDK uses string for IssuedAt
		ExpiresAt: sdkTimestampToDb(s.ExpiresAt),
		Revoked:   s.Revoked,
	}
}

// tokenFromDb converts a db Tokens to a SDK Token.
func tokenFromDb(d db.Tokens) modula.Token {
	return modula.Token{
		ID:        modula.TokenID(d.ID),
		UserID:    userIDPtr(d.UserID),
		TokenType: d.TokenType,
		Token:     d.Token,
		IssuedAt:  d.IssuedAt.String(),
		ExpiresAt: dbTimestampToSdk(d.ExpiresAt),
		Revoked:   d.Revoked,
	}
}

// tokenCreateFromDb converts db CreateTokenParams to SDK CreateTokenParams.
func tokenCreateFromDb(d db.CreateTokenParams) modula.CreateTokenParams {
	return modula.CreateTokenParams{
		UserID:    userIDPtr(d.UserID),
		TokenType: d.TokenType,
		Token:     d.Token,
		IssuedAt:  d.IssuedAt.String(),
		ExpiresAt: dbTimestampToSdk(d.ExpiresAt),
		Revoked:   d.Revoked,
	}
}

// ---------------------------------------------------------------------------
// Session: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// sessionToDb converts a SDK Session to a db Sessions.
func sessionToDb(s *modula.Session) db.Sessions {
	return db.Sessions{
		SessionID:   types.SessionID(string(s.SessionID)),
		UserID:      nullUserID(s.UserID),
		DateCreated: sdkTimestampToDb(s.DateCreated),
		ExpiresAt:   sdkTimestampToDb(s.ExpiresAt),
		LastAccess:  types.Timestamp{}, // SDK uses *string, needs separate handling
		IpAddress:   dbNullStr(s.IpAddress),
		UserAgent:   dbNullStr(s.UserAgent),
		SessionData: dbNullStr(s.SessionData),
	}
}

// sessionFromDb converts a db Sessions to a SDK Session.
func sessionFromDb(d db.Sessions) modula.Session {
	// LastAccess in SDK is *string, in db is types.Timestamp
	var lastAccess *string
	if d.LastAccess.Valid {
		s := d.LastAccess.String()
		lastAccess = &s
	}
	return modula.Session{
		SessionID:   modula.SessionID(string(d.SessionID)),
		UserID:      userIDPtr(d.UserID),
		DateCreated: dbTimestampToSdk(d.DateCreated),
		ExpiresAt:   dbTimestampToSdk(d.ExpiresAt),
		LastAccess:  lastAccess,
		IpAddress:   dbStrPtr(d.IpAddress),
		UserAgent:   dbStrPtr(d.UserAgent),
		SessionData: dbStrPtr(d.SessionData),
	}
}

// ---------------------------------------------------------------------------
// UserOauth: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// userOauthToDb converts a SDK UserOauth to a db UserOauth.
func userOauthToDb(s *modula.UserOauth) db.UserOauth {
	return db.UserOauth{
		UserOauthID:         types.UserOauthID(string(s.UserOauthID)),
		UserID:              nullUserID(s.UserID),
		OauthProvider:       s.OauthProvider,
		OauthProviderUserID: s.OauthProviderUserID,
		AccessToken:         s.AccessToken,
		RefreshToken:        s.RefreshToken,
		TokenExpiresAt:      types.Timestamp{}, // SDK uses string, db uses Timestamp
		DateCreated:         sdkTimestampToDb(s.DateCreated),
	}
}

// userOauthFromDb converts a db UserOauth to a SDK UserOauth.
func userOauthFromDb(d db.UserOauth) modula.UserOauth {
	return modula.UserOauth{
		UserOauthID:         modula.UserOauthID(string(d.UserOauthID)),
		UserID:              userIDPtr(d.UserID),
		OauthProvider:       d.OauthProvider,
		OauthProviderUserID: d.OauthProviderUserID,
		AccessToken:         d.AccessToken,
		RefreshToken:        d.RefreshToken,
		TokenExpiresAt:      d.TokenExpiresAt.String(),
		DateCreated:         dbTimestampToSdk(d.DateCreated),
	}
}

// ---------------------------------------------------------------------------
// UserSshKey: SDK <-> db (write-only: create via remote)
// ---------------------------------------------------------------------------

// userSshKeyToDb converts a SDK SshKey to a db UserSshKeys.
func userSshKeyToDb(s *modula.SshKey) db.UserSshKeys {
	return db.UserSshKeys{
		SshKeyID:    string(s.SshKeyID),
		UserID:      nullUserID(s.UserID),
		PublicKey:   s.PublicKey,
		KeyType:     s.KeyType,
		Fingerprint: s.Fingerprint,
		Label:       s.Label,
		DateCreated: sdkTimestampToDb(s.DateCreated),
		LastUsed:    s.LastUsed,
	}
}

// userSshKeyFromDb converts a db UserSshKeys to a SDK SshKey.
func userSshKeyFromDb(d db.UserSshKeys) modula.SshKey {
	return modula.SshKey{
		SshKeyID:    modula.UserSshKeyID(d.SshKeyID),
		UserID:      userIDPtr(d.UserID),
		PublicKey:   d.PublicKey,
		KeyType:     d.KeyType,
		Fingerprint: d.Fingerprint,
		Label:       d.Label,
		DateCreated: dbTimestampToSdk(d.DateCreated),
		LastUsed:    d.LastUsed,
	}
}

// userSshKeyCreateFromDb converts db CreateUserSshKeyParams to SDK CreateSSHKeyParams.
func userSshKeyCreateFromDb(d db.CreateUserSshKeyParams) modula.CreateSSHKeyParams {
	return modula.CreateSSHKeyParams{
		PublicKey: d.PublicKey,
		Label:     d.Label,
	}
}
