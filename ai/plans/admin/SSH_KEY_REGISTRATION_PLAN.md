# SSH Key Registration Redesign

## Philosophy

Any valid SSH key grants full admin-level read/write access to the CMS binary. This is by design — if you can SSH in, you're trusted. Modula does not implement SSH key allowlisting; it relies on users maintaining their hosting environments. If someone can place an SSH key on your server, your system is already compromised. We recommend [devops](https://github.com/hegner123/devops) for managing a secure VPS for ModulaCMS deployments.

Registration is purely organizational: it links an SSH key to a user profile for audit trails and identity tracking. It is never a gate.

## Current Behavior (Problems)

1. **Unregistered SSH keys are blocked** — `SSHAuthenticationMiddleware` sets `authenticated=false` for unknown fingerprints, and the TUI gates all access behind a mandatory provisioning form.
2. **Forced account creation** — the provisioning flow creates a new user regardless of whether the person already has an account on the site. There is no way to link a key to an existing user.
3. **No opt-out** — the form is mandatory. You cannot use the CMS without completing it.
4. **Bcrypt cost mismatch** — SSH provisioning uses `bcrypt.DefaultCost` (10) directly instead of `auth.HashPassword()` which uses cost 12. SSH-created users get weaker password hashes than bootstrap users.
5. **Ad-hoc DB connections** — `ProvisionSSHUser` calls `db.ConfigDB(*m.Config)` instead of using the injected `m.DB` driver, bypassing connection pooling.
6. **Hollow audit trails** — unregistered sessions use `UserID("")` in audit context, and the IP is hardcoded as "cli" instead of using the actual SSH client IP.

## Proposed Behavior

1. **All SSH keys get full admin access immediately** — registered or not.
2. **Registration is opt-in** — triggered from a menu item in the TUI when the user wants to link their key to a profile.
3. **Login-first flow** — registration starts by prompting for credentials to link to an existing account. Cancel falls back to a choice: try login again or create a new account.

### Flow

```
SSH with any valid key
    ↓
Full admin CMS access (registered or not)
    ↓
User selects "Register SSH Key" from menu
    ↓
Login form (email + password)
    ├─ Success → link key to that user → success dialog → done
    └─ Cancel / Fail
        ↓
    Choice dialog: "Try Again" | "Create New Account" | "Cancel"
        ├─ Try Again → back to login form
        ├─ Create New Account → registration form (username, name, email, password)
        │       ↓
        │   Create user → link key to new user → success dialog → done
        └─ Cancel → dismiss, back to CMS
```

## Changes

### 1. SSH Authentication Middleware

**File:** `internal/middleware/ssh_middleware.go`

#### SSHAuthenticationMiddleware

Unregistered keys must be granted full access. Change the "key not registered" branch:

**Before:**
```go
ctx.SetValue("needs_provisioning", true)
ctx.SetValue("authenticated", false)
```

**After:**
```go
ctx.SetValue("ssh_key_registered", false)
ctx.SetValue("authenticated", true)
```

The `needs_provisioning` context key is replaced with `ssh_key_registered` (a boolean describing state, not prescribing behavior). Authentication is always true for valid SSH keys.

For registered keys, also set:
```go
ctx.SetValue("ssh_key_registered", true)
```

**DB error handling:** The current code treats any error from `GetUserBySSHFingerprint` as "not registered" — including connection failures and timeouts. Under the new model this is acceptable: the user still gets full access (no automatic registration). But distinguish the error for logging:

```go
user, err := dbc.GetUserBySSHFingerprint(fingerprint)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        utility.DefaultLogger.Finfo("SSH key not registered: %s", fingerprint)
    } else {
        utility.DefaultLogger.Error("DB error looking up SSH key", err, "fingerprint", fingerprint)
    }
    ctx.SetValue("ssh_key_registered", false)
    ctx.SetValue("authenticated", true)
    next(s)
    return
}
```

**Store remote address in context** for downstream audit use:

```go
ctx.SetValue("ssh_remote_addr", s.RemoteAddr().String())
```

#### SSHAuthorizationMiddleware

Remove the `needs_provisioning` bypass. Simplify to: check `authenticated`, block if false, allow if true. Since all SSH keys set `authenticated=true`, this becomes a pass-through for SSH sessions.

### 2. TUI Middleware

**File:** `internal/tui/middleware.go`

Replace `needs_provisioning` context check with `ssh_key_registered`:

```go
if registered, ok := ctx.Value("ssh_key_registered").(bool); ok {
    m.SSHKeyRegistered = registered
}
```

Store the SSH remote address for audit context:

```go
if remoteAddr, ok := ctx.Value("ssh_remote_addr").(string); ok {
    m.SSHRemoteAddr = remoteAddr
}
```

Remove: `m.NeedsProvisioning = true` line.

### 3. Model

**File:** `internal/tui/model.go`

Replace provisioning fields:

**Remove:**
```go
NeedsProvisioning bool
```

**Add:**
```go
SSHKeyRegistered bool   // true if current SSH key is linked to a user profile
SSHRemoteAddr    string // client IP:port from SSH session, used in audit context
```

The existing `SSHFingerprint`, `SSHKeyType`, `SSHPublicKey`, and `UserID` fields stay as-is.

### 4. Audit Context for SSH Sessions

**File:** `internal/middleware/audit_helpers.go`

Add a new helper for SSH TUI operations that uses the real client IP and SSH fingerprint:

```go
// AuditContextFromSSH builds an AuditContext for SSH TUI operations.
// Uses the SSH client's remote address instead of "cli" and includes
// the SSH fingerprint for identity attribution on unregistered sessions.
func AuditContextFromSSH(c config.Config, userID types.UserID, remoteAddr string, fingerprint string) audited.AuditContext {
    // Use fingerprint as a request-ID-like identifier for unregistered sessions
    requestID := ""
    if userID == "" {
        requestID = fingerprint
    }
    return audited.Ctx(
        types.NodeID(c.Node_ID),
        userID,
        requestID,
        remoteAddr,
    )
}
```

Update all TUI command call sites that currently use `AuditContextFromCLI` to use `AuditContextFromSSH` when `m.IsSSH` is true:

```go
var ac audited.AuditContext
if m.IsSSH {
    ac = middleware.AuditContextFromSSH(*m.Config, m.UserID, m.SSHRemoteAddr, m.SSHFingerprint)
} else {
    ac = middleware.AuditContextFromCLI(*m.Config, m.UserID)
}
```

This gives unregistered SSH sessions:
- **Real client IP** instead of hardcoded "cli"
- **SSH fingerprint** in the request ID field so mutations are attributable to a specific key even without a named user

Registered SSH sessions additionally have `m.UserID` populated, so both fields are present.

**Note:** This is a cross-cutting change — the TUI has ~60+ command functions that call `AuditContextFromCLI`. To avoid touching every file, add a helper method on Model:

```go
// AuditContext returns the appropriate AuditContext for the current session type.
func (m Model) AuditContext() audited.AuditContext {
    if m.IsSSH {
        return middleware.AuditContextFromSSH(*m.Config, m.UserID, m.SSHRemoteAddr, m.SSHFingerprint)
    }
    return middleware.AuditContextFromCLI(*m.Config, m.UserID)
}
```

Then progressively replace `middleware.AuditContextFromCLI(*m.Config, m.UserID)` with `m.AuditContext()` across command files. This can be done incrementally — the old call sites still work, they just use "cli" as the IP.

### 5. View

**File:** `internal/tui/view.go`

Remove the provisioning gate entirely:

**Before:**
```go
func (m Model) View() string {
    if m.NeedsProvisioning {
        if m.FormState != nil && m.FormState.Form != nil {
            return m.FormState.Form.View()
        }
        return "Initializing user provisioning..."
    }
    return renderCMSPanelLayout(m)
}
```

**After:**
```go
func (m Model) View() string {
    return renderCMSPanelLayout(m)
}
```

### 6. Update Dispatcher

**File:** `internal/tui/update.go`

Remove the `UpdateProvisioning` gate from the main `Update` function:

**Remove:**
```go
// Handle user provisioning first if needed
if m, cmd := m.UpdateProvisioning(msg); cmd != nil {
    return m, cmd
}
```

Add SSH key registration message types to the main type switch. The registration flow is dialog-based — it uses the existing overlay/dialog system, not a blocking state machine.

### 7. Menu

**File:** `internal/tui/menus.go`

No new page. Add a menu item to the homepage that dispatches `StartSSHKeyRegistrationMsg`. The flow runs entirely through the overlay/dialog system.

```go
// SSH key registration (only shown for unregistered SSH keys)
if m.IsSSH && !m.SSHKeyRegistered {
    pages = append(pages, NewPage(ACTIONSPAGE, "Register SSH Key"))
}
```

This goes in the "Power user" section of `HomepageMenuInit`, after `ACTIONSPAGE`. Selecting it triggers the login dialog overlay instead of navigating to a page.

### 8. Registration Flow (Rewrite)

**File:** `internal/tui/ssh_key_registration.go` (new, replaces `user_provisioning.go`)

#### Messages

```go
// StartSSHKeyRegistrationMsg triggers the login-first registration flow.
type StartSSHKeyRegistrationMsg struct{}

// SSHKeyLoginResultMsg is the result of attempting to log in.
type SSHKeyLoginResultMsg struct {
    UserID types.UserID
    Error  error
}

// SSHKeyRegistrationCompleteMsg is sent when key registration finishes.
type SSHKeyRegistrationCompleteMsg struct {
    UserID types.UserID
    Error  error
}

// SSHKeyRegistrationChoiceMsg is sent when the user picks from the retry/create/cancel choice.
type SSHKeyRegistrationChoiceMsg struct {
    Choice string // "login", "create", "cancel"
}
```

#### Login Form

A huh form with email and password fields. On completion, dispatches a command that:
1. Looks up the user by email via `m.DB.GetUserByEmail()`
2. Verifies the password with `auth.CheckPasswordHash()` (not direct bcrypt)
3. On success: registers the SSH key to that user via `m.DB.CreateUserSshKey()`, returns `SSHKeyRegistrationCompleteMsg`
4. On failure: returns `SSHKeyLoginResultMsg` with error

All DB operations use `m.DB`, not `db.ConfigDB()`.

#### Choice Dialog

Shown after login failure or cancel. Three options via the existing dialog system:
- **Try Again** → re-show login form
- **Create New Account** → show the registration form
- **Cancel** → dismiss, return to CMS

#### Registration Form

Same fields as current provisioning: username, name, email, password, confirm. On completion:
1. Hash password with `auth.HashPassword()` (cost 12, not direct bcrypt)
2. Create user with **viewer** role — SSH access grants admin to the TUI binary, but the user account persists across all interfaces (HTTP API, admin panel). The viewer role is intentionally restrictive for web-facing access. Admin promotion is a separate, explicit action.
3. Register SSH key to that user
4. Return `SSHKeyRegistrationCompleteMsg`

#### Success Handling

On `SSHKeyRegistrationCompleteMsg` with no error:
- Set `m.SSHKeyRegistered = true`
- Set `m.UserID = msg.UserID`
- Show success dialog: "SSH key registered to [user email]"

### 9. Update Handler for Registration

**File:** `internal/tui/update_ssh_registration.go` (new, replaces `update_provisioning.go`)

Handle the registration messages. The flow uses the overlay/dialog system:
- `StartSSHKeyRegistrationMsg` → push login form as overlay
- Login form completion → dispatch login command
- `SSHKeyLoginResultMsg` with error → show choice dialog
- `SSHKeyRegistrationChoiceMsg` "login" → push login form again
- `SSHKeyRegistrationChoiceMsg` "create" → push registration form as overlay
- Registration form completion → dispatch create user + register key command
- `SSHKeyRegistrationCompleteMsg` → show success dialog, update model state

### 10. Role Assignment

**Current:** provisioning creates users with the `viewer` role.

**New:** keep the `viewer` role. SSH access grants full admin to the TUI binary regardless of role. The user account is used across HTTP API and admin panel where RBAC applies. Creating an admin account via SSH self-registration would grant unintended admin access to the web admin panel. Admin promotion is a deliberate action by an existing admin, not an automatic side effect of SSH registration.

### 11. Cleanup

**Delete:**
- `internal/tui/update_provisioning.go` — replaced by `update_ssh_registration.go`
- `internal/tui/user_provisioning.go` — replaced by `ssh_key_registration.go`

**Remove from source:**
- `NeedsProvisioning` field from Model
- `needs_provisioning` context key from middleware
- Provisioning gate in `view.go`
- `UpdateProvisioning` call in `update.go`
- Direct `bcrypt.GenerateFromPassword` call (use `auth.HashPassword()`)
- `db.ConfigDB(*m.Config)` call in registration (use `m.DB`)

**Keep:**
- `SSHFingerprint`, `SSHKeyType`, `SSHPublicKey` fields on Model
- All DB methods for SSH keys (`CreateUserSshKey`, `GetUserSshKeyByFingerprint`, etc.)
- `FingerprintSHA256` helper in middleware

## Edge Cases

### Key already registered to another user

If the middleware finds the key via `GetUserBySSHFingerprint`, it sets `ssh_key_registered=true` and populates `m.UserID`. The "Register SSH Key" menu item is hidden. This is correct — the key is already linked. Re-registration is not supported; to move a key between users, delete it first via the users admin screen.

### Concurrent sessions with same unregistered key

Two terminals, same unregistered key. Both get full admin. One runs registration. The other session's in-memory `SSHKeyRegistered` and `UserID` do not update — they continue working without a linked identity. Audit records for the unregistered session still have the fingerprint. This is acceptable; the next SSH connection with that key will pick up the registration.

### DB error during fingerprint lookup

`GetUserBySSHFingerprint` may fail due to connection errors, not just "key not found." The middleware distinguishes these for logging but treats both as `ssh_key_registered=false, authenticated=true`. The user gets full access regardless. No automatic registration is triggered — registration is always opt-in.

## Files Summary

| File | Action |
|------|--------|
| `internal/middleware/ssh_middleware.go` | Modify — grant full access to unregistered keys, replace `needs_provisioning` with `ssh_key_registered`, log DB errors distinctly, store `ssh_remote_addr` |
| `internal/middleware/audit_helpers.go` | Modify — add `AuditContextFromSSH` helper |
| `internal/tui/middleware.go` | Modify — read `ssh_key_registered` and `ssh_remote_addr`, set `SSHKeyRegistered` and `SSHRemoteAddr` |
| `internal/tui/model.go` | Modify — replace `NeedsProvisioning` with `SSHKeyRegistered`, add `SSHRemoteAddr`, add `AuditContext()` method |
| `internal/tui/view.go` | Modify — remove provisioning gate |
| `internal/tui/update.go` | Modify — remove `UpdateProvisioning` gate, add registration message handlers |
| `internal/tui/ssh_key_registration.go` | New — login form, choice dialog, registration form, commands |
| `internal/tui/update_ssh_registration.go` | New — message handler for registration flow |
| `internal/tui/update_provisioning.go` | Delete |
| `internal/tui/user_provisioning.go` | Delete |
| `internal/tui/menus.go` | Modify — add "Register SSH Key" menu item for unregistered SSH sessions |

## Testing

- SSH with unregistered key → full CMS access, no form shown
- SSH with registered key → full CMS access, `UserID` populated
- Trigger "Register SSH Key" → login form appears
- Login with valid credentials → key linked, success dialog
- Login with invalid credentials → choice dialog shown
- Login with invalid email → error message, not crash
- Choice "Try Again" → login form re-shown
- Choice "Create New Account" → registration form shown
- Complete registration → user created with viewer role, key linked, success dialog
- Choice "Cancel" → dialog dismissed, back to CMS
- After registration, `SSHKeyRegistered` is true, "Register SSH Key" menu item hidden
- Audit records for unregistered session contain SSH fingerprint and real client IP
- Audit records for registered session contain user ID, fingerprint, and real client IP
- DB connection failure during fingerprint lookup → user still gets full CMS access, error logged
- Password hashing uses `auth.HashPassword()` (cost 12)
- Login verification uses `auth.CheckPasswordHash()`
