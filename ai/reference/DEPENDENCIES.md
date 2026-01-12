# DEPENDENCIES.md

Complete Dependency Reference for ModulaCMS

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/DEPENDENCIES.md`
**Created:** 2026-01-12
**Purpose:** Documents every dependency in go.mod, explains why each was chosen, lists alternatives considered, and provides version management guidance.

---

## Overview

ModulaCMS has 18 direct dependencies and 37 indirect dependencies (total: 55 packages). This document explains the purpose of each dependency, why it was chosen over alternatives, and important version considerations.

**Go version:** 1.23.0
**Toolchain:** go1.24.2

---

## Critical Dependencies

These dependencies are essential to core functionality. Removing them would require significant architectural changes.

### Database Drivers

#### github.com/mattn/go-sqlite3 v1.14.24

**Purpose:** SQLite database driver for Go

**Why chosen:**
- Most mature and widely-used SQLite driver for Go
- Supports all SQLite features including foreign keys, transactions, and prepared statements
- Actively maintained with regular security updates
- CGO-based (compiles SQLite C library) for maximum performance

**Alternatives considered:**
- `modernc.org/sqlite` (pure Go, no CGO) - Rejected due to slightly lower performance and less battle-testing
- `crawshaw.io/sqlite` (low-level bindings) - Rejected due to more complex API

**Trade-offs:**
- **Requires CGO:** Cross-compilation is more complex, build times are slower
- **Platform-specific builds:** Binaries are platform-dependent
- **Worth it:** Performance and maturity outweigh CGO complexity

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`
- Development and small deployment database engine

**Documentation:** https://github.com/mattn/go-sqlite3

---

#### github.com/go-sql-driver/mysql v1.9.0

**Purpose:** MySQL database driver for Go

**Why chosen:**
- Official MySQL driver recommended by the Go community
- Pure Go implementation (no CGO required for MySQL)
- Excellent performance and connection pooling support
- Supports all modern MySQL features (prepared statements, transactions, SSL/TLS)

**Alternatives considered:**
- `github.com/siddontang/go-mysql` - Rejected, less mature for production use
- `gorm.io/driver/mysql` - Rejected, adds unnecessary ORM layer

**Trade-offs:**
- None significant - this is the de facto standard

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`
- Production database engine

**Documentation:** https://github.com/go-sql-driver/mysql

---

#### github.com/lib/pq v1.10.9

**Purpose:** PostgreSQL database driver for Go

**Why chosen:**
- Most mature and widely-used PostgreSQL driver for Go
- Pure Go implementation (no CGO required)
- Supports all PostgreSQL features including LISTEN/NOTIFY, COPY, and advanced types
- Excellent error handling and diagnostics

**Alternatives considered:**
- `github.com/jackc/pgx` - Considered, but pq is more battle-tested and stable
- `gorm.io/driver/postgres` - Rejected, adds unnecessary ORM layer

**Trade-offs:**
- None significant - this is the de facto standard

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`
- Enterprise database engine

**Documentation:** https://github.com/lib/pq

---

#### github.com/sqlc-dev/pqtype v0.3.0

**Purpose:** PostgreSQL type definitions for sqlc code generation

**Why chosen:**
- Required by sqlc for PostgreSQL-specific types (arrays, JSON, UUID, etc.)
- Provides proper Go type mappings for PostgreSQL's rich type system
- Maintained by the sqlc team for compatibility

**Alternatives considered:**
- None - this is a required companion package for sqlc + PostgreSQL

**Trade-offs:**
- Additional dependency, but unavoidable for PostgreSQL support

**Version constraints:** Must match sqlc compatibility requirements

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-psql/` (sqlc-generated code)

**Documentation:** https://github.com/sqlc-dev/pqtype

---

### Charmbracelet TUI Ecosystem

ModulaCMS uses the Charmbracelet ecosystem for its SSH-based TUI. This is a cohesive set of libraries designed to work together.

#### github.com/charmbracelet/bubbletea v1.3.4

**Purpose:** TUI framework implementing Elm Architecture (Model-Update-View)

**Why chosen:**
- Modern, well-designed architecture for complex interactive applications
- Excellent state management (no scattered mutations)
- Active development and strong community
- Used by popular tools like lazygit, glow, and gh (GitHub CLI)
- Message-based architecture makes testing easier

**Alternatives considered:**
- `github.com/rivo/tview` - Rejected, widget-based approach is less flexible
- `github.com/jroimartin/gocui` - Rejected, lower-level and more complex
- `github.com/gdamore/tcell` - Rejected, too low-level (Bubbletea uses this internally)

**Trade-offs:**
- Learning curve for Elm Architecture pattern
- Worth it: Results in maintainable, testable TUI code

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (entire TUI implementation)

**Documentation:** https://github.com/charmbracelet/bubbletea

---

#### github.com/charmbracelet/bubbles v0.21.0

**Purpose:** Pre-built TUI components (lists, text inputs, viewports, spinners)

**Why chosen:**
- Official companion library to Bubbletea
- Provides production-ready, accessible components
- Consistent styling and behavior
- Saves development time on common UI patterns

**Alternatives considered:**
- Build everything from scratch - Rejected, waste of time
- Use tview widgets - Rejected, not compatible with Bubbletea

**Trade-offs:**
- None - these are optional helpers that speed up development

**Version constraints:** Must be compatible with Bubbletea version

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (list views, text inputs, spinners)

**Documentation:** https://github.com/charmbracelet/bubbles

---

#### github.com/charmbracelet/lipgloss v1.1.0

**Purpose:** Style definitions and layout for terminal UIs

**Why chosen:**
- Declarative styling (like CSS for terminals)
- Supports colors, borders, padding, margins, alignment
- Works seamlessly with Bubbletea
- Makes TUI styling maintainable

**Alternatives considered:**
- Manual ANSI escape codes - Rejected, unmaintainable
- `github.com/fatih/color` - Rejected, too limited (only colors, no layout)

**Trade-offs:**
- None - essential for professional-looking TUIs

**Version constraints:** Must be compatible with Bubbletea version

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (all rendering functions)

**Documentation:** https://github.com/charmbracelet/lipgloss

---

#### github.com/charmbracelet/huh v0.6.0

**Purpose:** Form library with validation, theming, and accessibility

**Why chosen:**
- Declarative form definitions
- Built-in validation
- Keyboard navigation and accessibility
- Consistent with Charmbracelet ecosystem
- Supports text inputs, selections, confirmations, multi-selects

**Alternatives considered:**
- Build forms from scratch using Bubbles components - Rejected, too much work
- `github.com/AlecAivazis/survey` - Rejected, not Bubbletea-compatible

**Trade-offs:**
- None - essential for user input in TUI

**Version constraints:** Must be compatible with Bubbletea version

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (content editing forms, installation wizard)
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/install/` (installation forms)

**Documentation:** https://github.com/charmbracelet/huh

---

#### github.com/charmbracelet/log v0.4.1

**Purpose:** Structured logging with levels, key-value pairs, and formatting

**Why chosen:**
- Consistent with Charmbracelet ecosystem
- Beautiful, readable log output
- Structured logging (key-value pairs, not printf)
- Supports log levels (Debug, Info, Warn, Error, Fatal)
- Works well in both TUI and daemon modes

**Alternatives considered:**
- `github.com/sirupsen/logrus` - Rejected, heavier and not as pretty
- `go.uber.org/zap` - Rejected, overkill for CMS needs
- Standard library `log` - Rejected, too basic

**Trade-offs:**
- None - excellent balance of features and simplicity

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/logger.go`
- Used throughout the application

**Documentation:** https://github.com/charmbracelet/log

---

#### github.com/charmbracelet/ssh v0.0.0-20250213143314-8712ec3ff3ef

**Purpose:** SSH server library (wrapper around golang.org/x/crypto/ssh)

**Why chosen:**
- Built specifically for Bubbletea TUI applications over SSH
- Simplifies SSH server setup
- Handles authentication, sessions, and terminal emulation
- Maintained by Charmbracelet for their ecosystem

**Alternatives considered:**
- Raw `golang.org/x/crypto/ssh` - Rejected, too low-level and complex
- `github.com/gliderlabs/ssh` - Rejected, less integrated with Bubbletea

**Trade-offs:**
- Uses bleeding-edge version (commit hash, not semver tag)
- Risk: Potential instability from unreleased code
- Worth it: Needed for latest Bubbletea compatibility

**Version constraints:** Must match Wish and Bubbletea requirements

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go` (SSH server initialization)

**Documentation:** https://github.com/charmbracelet/ssh

---

#### github.com/charmbracelet/wish v1.4.6

**Purpose:** SSH middleware for Bubbletea applications

**Why chosen:**
- Connects SSH server to Bubbletea TUI applications
- Provides middleware for logging, authentication, terminal setup
- Essential for serving TUI over SSH
- Part of Charmbracelet ecosystem

**Alternatives considered:**
- None - this is the only solution for Bubbletea + SSH

**Trade-offs:**
- None - required for SSH functionality

**Version constraints:** Must be compatible with Bubbletea and charmbracelet/ssh

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go` (SSH server middleware)
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (TUI middleware)

**Documentation:** https://github.com/charmbracelet/wish

---

### Storage and Media

#### github.com/aws/aws-sdk-go v1.55.5

**Purpose:** AWS SDK for S3-compatible object storage

**Why chosen:**
- Industry standard for S3 API interaction
- Supports AWS S3 and all S3-compatible providers (Linode, DigitalOcean, Backblaze, etc.)
- Comprehensive API coverage
- Excellent error handling and retry logic
- Well-documented and battle-tested

**Alternatives considered:**
- `github.com/minio/minio-go` - Considered, but AWS SDK is more widely known
- Direct HTTP calls to S3 API - Rejected, too much work to handle signing, errors, etc.

**Trade-offs:**
- Large dependency (AWS SDK is big)
- Worth it: Saves enormous development time and handles edge cases

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/bucket/` (S3 operations)
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/media/` (media uploads)

**Documentation:** https://docs.aws.amazon.com/sdk-for-go/

---

#### golang.org/x/image v0.22.0

**Purpose:** Image encoding/decoding and manipulation

**Why chosen:**
- Official Go extended library for image processing
- Supports PNG, JPEG, GIF, WebP, BMP
- Provides draw package for scaling, cropping, and compositing
- Pure Go implementation (no CGO required)
- High-quality scaling algorithms

**Alternatives considered:**
- `github.com/disintegration/imaging` - Rejected, adds another dependency layer
- `github.com/nfnt/resize` - Rejected, less feature-complete
- ImageMagick via exec - Rejected, requires external binary

**Trade-offs:**
- None - excellent balance of features and performance

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/media/media_optomize.go` (image scaling, cropping)

**Documentation:** https://pkg.go.dev/golang.org/x/image

---

### Authentication and Security

#### golang.org/x/oauth2 v0.29.0

**Purpose:** OAuth 2.0 client implementation

**Why chosen:**
- Official Go extended library for OAuth
- Supports standard OAuth 2.0 and OpenID Connect
- Works with any OAuth provider (Azure AD, Okta, Google, GitHub, etc.)
- Handles token refresh, expiration, and storage
- Well-tested and secure

**Alternatives considered:**
- `github.com/coreos/go-oidc` - Considered for OpenID Connect specifically
- Manual OAuth implementation - Rejected, security-critical code should use vetted libraries

**Trade-offs:**
- None - this is the standard library for OAuth in Go

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/auth/` (OAuth flow)
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/config/config.go` (OAuth configuration)

**Documentation:** https://pkg.go.dev/golang.org/x/oauth2

---

#### golang.org/x/crypto v0.37.0

**Purpose:** Cryptographic functions and Let's Encrypt autocert

**Why chosen:**
- Official Go extended library for cryptography
- Provides `acme/autocert` for Let's Encrypt certificate management
- SSH server support via `golang.org/x/crypto/ssh`
- Secure password hashing (bcrypt, argon2)
- Up-to-date with latest security standards

**Alternatives considered:**
- Standard library `crypto` - Rejected, doesn't include autocert or SSH
- Third-party autocert libraries - Rejected, golang.org/x/crypto is the standard

**Trade-offs:**
- None - essential for security

**Version constraints:** Must stay current for security patches

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go` (autocert for HTTPS)
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/auth/` (password hashing)
- SSH server (underlying implementation)

**Documentation:** https://pkg.go.dev/golang.org/x/crypto

---

### Plugin System

#### github.com/yuin/gopher-lua v1.1.1

**Purpose:** Lua VM for Go (plugin scripting)

**Why chosen:**
- Pure Go implementation of Lua 5.1
- No CGO required (cross-platform friendly)
- Good performance for embedded scripting
- Can call Go functions from Lua scripts
- Sandboxing capabilities for security

**Alternatives considered:**
- `github.com/Shopify/go-lua` - Rejected, less mature and fewer features
- JavaScript via `github.com/robertkrimen/otto` - Rejected, JS is overkill for plugins
- Python via `github.com/sbinet/go-python` - Rejected, requires CGO and Python runtime
- WebAssembly (wazero) - Considered for future, but Lua is simpler for users

**Trade-offs:**
- Lua 5.1 (not latest Lua 5.4) - Acceptable, 5.1 is widely known
- Performance is good but not native - Acceptable for plugin use case

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/plugin/plugin.go` (plugin execution)

**Documentation:** https://github.com/yuin/gopher-lua

---

### Text Processing

#### github.com/muesli/reflow v0.3.0

**Purpose:** Text wrapping and reflow for terminal UIs

**Why chosen:**
- Handles word wrapping at specified widths
- Supports ANSI escape codes (preserves styling during wrap)
- Essential for responsive terminal layouts
- Works seamlessly with Lipgloss

**Alternatives considered:**
- Manual text wrapping - Rejected, complex to handle ANSI codes correctly
- Standard library `text/tabwriter` - Rejected, doesn't handle wrapping

**Trade-offs:**
- None - essential for TUI text rendering

**Version constraints:** None (uses latest stable)

**Usage locations:**
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (text rendering in TUI)

**Documentation:** https://github.com/muesli/reflow

---

#### github.com/muesli/termenv v0.16.0

**Purpose:** Terminal environment detection (color support, capabilities)

**Why chosen:**
- Detects terminal color support (truecolor, 256-color, 16-color, monochrome)
- Provides color profile information
- Handles terminal-specific quirks
- Essential for cross-terminal compatibility

**Alternatives considered:**
- Manual terminal detection - Rejected, too complex and error-prone
- `github.com/mattn/go-isatty` alone - Rejected, only detects TTY, not capabilities

**Trade-offs:**
- None - essential for robust terminal support

**Version constraints:** None (uses latest stable)

**Usage locations:**
- Used indirectly by Lipgloss and Bubbletea
- `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (terminal setup)

**Documentation:** https://github.com/muesli/termenv

---

## Indirect Dependencies

These dependencies are pulled in automatically by the direct dependencies above. Understanding them helps with debugging and security auditing.

### Authentication and Cryptography

#### filippo.io/edwards25519 v1.1.0

**Purpose:** Edwards25519 elliptic curve implementation

**Why present:** Used by golang.org/x/crypto for SSH key generation (Ed25519 keys)

**Notes:** Ed25519 is the modern standard for SSH keys (faster and more secure than RSA)

---

### Terminal and UI Components

#### github.com/atotto/clipboard v0.1.4

**Purpose:** Cross-platform clipboard access

**Why present:** Used by Charmbracelet components for copy/paste functionality

---

#### github.com/aymanbagabas/go-osc52/v2 v2.0.1

**Purpose:** OSC 52 escape sequence support (terminal clipboard integration)

**Why present:** Used by Charmbracelet components for clipboard operations over SSH

**Notes:** OSC 52 allows clipboard access in remote terminal sessions

---

#### github.com/catppuccin/go v0.2.0

**Purpose:** Catppuccin color palette

**Why present:** Used by Charmbracelet components for theming

**Notes:** Catppuccin is a popular pastel color scheme

---

#### github.com/charmbracelet/colorprofile v0.2.3

**Purpose:** Color profile detection for terminals

**Why present:** Used by Lipgloss to determine color capabilities

---

#### github.com/charmbracelet/keygen v0.5.1

**Purpose:** SSH key generation

**Why present:** Used by charmbracelet/ssh for generating host keys

---

#### github.com/charmbracelet/x/* (multiple packages)

**Purpose:** Experimental Charmbracelet utilities

**Why present:** Internal utilities used across Charmbracelet ecosystem

**Notes:** These packages are in the `x` namespace (experimental/extended)

**Packages:**
- `x/ansi` - ANSI escape code handling
- `x/cellbuf` - Terminal cell buffer manipulation
- `x/conpty` - Windows ConPTY support
- `x/errors` - Enhanced error handling
- `x/exp/strings` - Experimental string utilities
- `x/input` - Input event handling
- `x/term` - Terminal management
- `x/termios` - Unix terminal I/O settings
- `x/windows` - Windows terminal support

---

#### github.com/creack/pty v1.1.21

**Purpose:** Cross-platform PTY (pseudo-terminal) support

**Why present:** Used by charmbracelet/ssh for terminal emulation

**Notes:** Essential for SSH server terminal handling

---

#### github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f

**Purpose:** Windows console input handling

**Why present:** Used by Bubbletea for Windows terminal support

---

#### github.com/lucasb-eyer/go-colorful v1.2.0

**Purpose:** Color space conversions and manipulation

**Why present:** Used by Lipgloss for color operations

---

#### github.com/mattn/go-isatty v0.0.20

**Purpose:** Detects if file descriptor is a terminal (TTY)

**Why present:** Used by Charmbracelet components to detect terminal vs pipe

---

#### github.com/mattn/go-localereader v0.0.1

**Purpose:** Locale-aware input reading

**Why present:** Used by Bubbletea for internationalization support

---

#### github.com/mattn/go-runewidth v0.0.16

**Purpose:** Unicode character width calculation for terminal rendering

**Why present:** Used by Lipgloss and Bubbletea for proper text alignment

**Notes:** Essential for rendering CJK (Chinese, Japanese, Korean) characters correctly

---

#### github.com/mitchellh/hashstructure/v2 v2.0.2

**Purpose:** Generate hashes from Go structs

**Why present:** Used by Charmbracelet components for caching and change detection

---

#### github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6

**Purpose:** ANSI escape sequence parsing and manipulation

**Why present:** Used by reflow and Lipgloss for styled text handling

---

#### github.com/muesli/cancelreader v0.2.2

**Purpose:** Cancelable reader implementation

**Why present:** Used by Bubbletea for interruptible input reading

---

#### github.com/rivo/uniseg v0.4.7

**Purpose:** Unicode text segmentation (grapheme clusters, word boundaries)

**Why present:** Used by go-runewidth for proper Unicode handling

**Notes:** Handles complex Unicode like emoji and combining characters

---

#### github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e

**Purpose:** Terminal capability database access

**Why present:** Used by termenv for terminal feature detection

---

### SSH Server

#### github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be

**Purpose:** Shell-like lexical analysis (parse shell commands)

**Why present:** Used by charmbracelet/ssh for command parsing

---

### AWS SDK Dependencies

#### github.com/jmespath/go-jmespath v0.4.0

**Purpose:** JMESPath query language for JSON

**Why present:** Used by AWS SDK for filtering and querying API responses

---

### Utilities

#### github.com/dustin/go-humanize v1.0.1

**Purpose:** Human-readable formatting (file sizes, time durations, numbers)

**Why present:** Used by Charmbracelet components for user-friendly display

**Example:** "2.5 MB" instead of "2621440 bytes"

---

#### github.com/go-logfmt/logfmt v0.6.0

**Purpose:** Logfmt format parsing and encoding

**Why present:** Used by charmbracelet/log for structured logging

---

### Go Extended Libraries

#### golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56

**Purpose:** Experimental Go features

**Why present:** Used by various dependencies for experimental stdlib features

---

#### golang.org/x/net v0.34.0

**Purpose:** Extended networking libraries

**Why present:** Used by golang.org/x/oauth2 for HTTP client enhancements

---

#### golang.org/x/sync v0.13.0

**Purpose:** Additional synchronization primitives

**Why present:** Used by AWS SDK and other concurrent operations

---

#### golang.org/x/sys v0.32.0

**Purpose:** System calls and OS-specific operations

**Why present:** Used by many dependencies for platform-specific code

---

#### golang.org/x/text v0.24.0

**Purpose:** Text processing (encoding, transformations, localization)

**Why present:** Used by various dependencies for Unicode handling

---

### Configuration

#### gopkg.in/yaml.v2 v2.4.0

**Purpose:** YAML parsing and encoding

**Why present:** Used by various dependencies for configuration files

**Notes:** ModulaCMS uses JSON for config, but dependencies may use YAML

---

## Dependency Management Strategy

### Version Pinning

**Current approach:** No version pinning (uses latest stable releases)

**Reasoning:**
- Receive security updates automatically
- Benefit from bug fixes
- Stay compatible with latest Go versions

**Risk mitigation:**
- Pin versions if breaking changes occur
- Test thoroughly after `go get -u` updates
- Monitor release notes for major dependencies

### Security Updates

**Critical dependencies to monitor:**
1. `golang.org/x/crypto` - Security patches
2. `github.com/mattn/go-sqlite3` - SQLite CVEs
3. `github.com/go-sql-driver/mysql` - MySQL driver vulnerabilities
4. `github.com/lib/pq` - PostgreSQL driver vulnerabilities
5. `github.com/aws/aws-sdk-go` - AWS SDK security issues

**Update frequency:**
- Security patches: Immediately
- Bug fixes: Within 1 week
- Feature updates: As needed

### Dependency Audit

**Run regularly:**
```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Or use Go's built-in vulnerability checking
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Check for outdated dependencies
go list -u -m all
```

### Minimal Dependency Philosophy

**Guidelines for adding new dependencies:**
1. **Is it necessary?** Can we implement this ourselves in <100 lines?
2. **Is it maintained?** Last commit within 6 months, active issues, responsive maintainers?
3. **Is it widely used?** >1000 stars, used by other major projects?
4. **Is it stable?** Semver v1.0+, not pre-release?
5. **Does it pull in many dependencies?** Check transitive dependency count

**Red flags:**
- Unmaintained (no commits in >1 year)
- Pulls in >10 transitive dependencies
- Pre-v1.0 with frequent breaking changes
- Only one maintainer with no bus factor
- Large binary size impact (check with `go build -ldflags="-w -s"`)

---

## Update Checklist

**Before running `go get -u`:**

1. **Read release notes** for major dependencies
2. **Check for breaking changes** in changelogs
3. **Backup go.mod and go.sum** (`cp go.mod go.mod.backup`)
4. **Update in isolation** (one dependency at a time for major versions)
5. **Run full test suite** (`make test`)
6. **Test manually** with `--cli` mode
7. **Check build size** (`ls -lh modulacms-x86`)
8. **Commit with detailed message** explaining what was updated and why

**After update:**
```bash
# Verify all dependencies are properly downloaded
go mod download

# Tidy up go.mod
go mod tidy

# Verify checksums
go mod verify

# Run tests
make test

# Check for vulnerabilities
govulncheck ./...
```

---

## Build Tags and CGO

### CGO Dependencies

**Only one CGO dependency:** `github.com/mattn/go-sqlite3`

**Building without CGO:**
```bash
# This will fail because sqlite3 requires CGO
CGO_ENABLED=0 go build ./cmd
# Error: build constraints exclude all Go files

# To build without SQLite (hypothetical - not currently supported):
# Would need build tags to exclude SQLite code
```

**Cross-compilation with CGO:**
```bash
# Building for Linux AMD64 from macOS
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build ./cmd

# Requires cross-compilation toolchain for target platform
```

**Recommendation:** Keep SQLite as only CGO dependency. If adding new dependencies, prefer pure Go implementations.

---

## Alternatives We Chose Not To Use

### ORM Libraries

**Not using:**
- GORM
- Ent
- SQLBoiler

**Why:** sqlc provides type safety without ORM overhead. ORMs abstract away SQL, but ModulaCMS needs control over queries for performance (tree operations, lazy loading).

### Web Frameworks

**Not using:**
- Gin
- Echo
- Fiber
- Chi

**Why:** Standard library `net/http` is sufficient. Frameworks add unnecessary complexity and dependencies for a CMS serving simple HTTP endpoints.

### Template Engines

**Not using:**
- html/template
- pongo2
- jet

**Why:** ModulaCMS is headless. The client builds their own frontend. No server-side rendering needed.

### Testing Frameworks

**Not using:**
- testify
- goconvey
- ginkgo

**Why:** Standard library `testing` package is sufficient. These frameworks add sugar but not essential value.

---

## Dependency License Compliance

**All dependencies use permissive licenses:**
- **MIT License:** Majority of dependencies
- **Apache 2.0:** AWS SDK
- **BSD 3-Clause:** Go extended libraries (golang.org/x/*)

**No GPL dependencies:** No copyleft requirements for ModulaCMS binary distribution.

**License audit:**
```bash
# Install go-licenses
go install github.com/google/go-licenses@latest

# Check all licenses
go-licenses report ./cmd --template=licenses.tpl
```

---

## Future Dependency Considerations

### Potential Additions

**Monitoring/Observability:**
- `github.com/prometheus/client_golang` - Metrics export
- Consider: Only if users request it

**Internationalization:**
- `golang.org/x/text/message` - Already indirectly included
- Consider: If building admin UI features

**Caching:**
- `github.com/patrickmn/go-cache` - In-memory cache
- Consider: For content tree caching in high-traffic scenarios

**Validation:**
- `github.com/go-playground/validator` - Struct validation
- Consider: If form validation becomes more complex

### Potential Removals

**None currently.** All dependencies are actively used and provide clear value.

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - Database driver usage
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md` - Charmbracelet ecosystem usage
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/PLUGIN_ARCHITECTURE.md` - Lua plugin system

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MEDIA_PACKAGE.md` - Image processing and AWS SDK
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI framework usage
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/AUTH_PACKAGE.md` - OAuth implementation

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Dependency mocking strategies

**Reference:**
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Build commands and conventions

---

## Quick Reference

### Direct Dependencies by Category

**Database (4):**
- `github.com/mattn/go-sqlite3` - SQLite driver (CGO)
- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/sqlc-dev/pqtype` - PostgreSQL types for sqlc

**TUI/CLI (7):**
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - TUI styling
- `github.com/charmbracelet/huh` - Forms
- `github.com/charmbracelet/log` - Logging
- `github.com/charmbracelet/ssh` - SSH server
- `github.com/charmbracelet/wish` - SSH middleware

**Storage (1):**
- `github.com/aws/aws-sdk-go` - S3-compatible storage

**Security (2):**
- `golang.org/x/crypto` - Cryptography, autocert, SSH
- `golang.org/x/oauth2` - OAuth 2.0 client

**Media (1):**
- `golang.org/x/image` - Image processing

**Plugins (1):**
- `github.com/yuin/gopher-lua` - Lua VM

**Text Processing (2):**
- `github.com/muesli/reflow` - Text wrapping
- `github.com/muesli/termenv` - Terminal detection

### Common Update Commands

```bash
# Update all dependencies to latest
go get -u ./...

# Update specific dependency
go get -u github.com/charmbracelet/bubbletea@latest

# Update to specific version
go get github.com/charmbracelet/bubbletea@v1.3.4

# Update Go toolchain
go get toolchain@go1.24.2

# Clean up unused dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Security Check Commands

```bash
# Check for known vulnerabilities
govulncheck ./...

# Audit dependencies with nancy
go list -json -m all | nancy sleuth

# Check for outdated versions
go list -u -m all

# List all dependencies
go mod graph
```

---

**Last Updated:** 2026-01-12
