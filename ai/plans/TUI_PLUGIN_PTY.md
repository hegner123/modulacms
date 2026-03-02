# TUI Plugin System: PTY-Embedded Plugin Screens

**Date:** 2026-03-01
**Status:** Draft
**Depends on:** Existing Lua plugin system (`internal/plugin/`)

---

## Motivation

The TUI is the only part of ModulaCMS that cannot be extended by plugins. Adding or modifying screens requires editing Go source code. This violates the core value of flexibility. Plugin authors should be able to ship TUI screens written in any language, using any TUI framework, without touching ModulaCMS internals.

## Approach

Embed plugin TUI processes via PTY allocation and virtual terminal emulation, following the same architectural patterns as tmux. The host TUI (Bubbletea) owns chrome (sidebar, topbar, navigation). When a plugin screen is active, the host allocates a PTY, starts the plugin binary, interprets its ANSI output through a VT emulator, and renders the result into the content area. The plugin process thinks it has a real terminal.

---

## Library Selection

| Component | Library | Rationale |
|-----------|---------|-----------|
| PTY allocation | `creack/pty` v1.1.24 | De facto standard, production-grade, 1100+ importers |
| VT emulation | `charmbracelet/x/vt` | Same team as Bubbletea, `SafeEmulator` for thread safety, `Render()` for ANSI string output, `Draw()` for v2 cell compositing, alt screen + true color + callbacks |
| Rendering | Bubbletea v1 `View()` → raw ANSI string | `vt.Render()` output returned directly from `View()` |

`taigrr/bubbleterm` (v0.1.0, Feb 2026) does exactly this but is unproven (zero importers). Reference it for implementation patterns but do not depend on it.

---

## Architecture

### Component Overview

```
internal/plugin/
    tui_process.go       # PluginProcess: PTY lifecycle, read loop, resize
    tui_manifest.go      # TUI manifest parsing and validation
    tui_api_server.go    # Unix socket JSON-RPC server for plugin data access

internal/cli/
    plugin_screen.go     # PluginScreenModel: Bubbletea model wrapping PluginProcess + VT
    plugin_input.go      # Key translation: tea.KeyMsg → ANSI bytes for PTY write
```

### Data Flow

```
Plugin Binary (any language/framework)
    ↕ (stdin/stdout/stderr via slave PTY)
PTY Pair (creack/pty)
    ↕ (master fd: raw bytes with ANSI escapes)
Read Goroutine
    ↕ (Write() bytes into emulator)
SafeEmulator (charmbracelet/x/vt)
    ↕ (Render() → ANSI string)
PluginScreenModel.View()
    ↕ (composed with host chrome)
Bubbletea Renderer
    ↕ (diff + write to terminal)
SSH Client (via Wish)
```

### Input Flow

```
SSH Client keystroke
    → Wish → Bubbletea tea.KeyMsg
    → PluginScreenModel.Update()
    → Dispatch check:
        Meta key (Ctrl-]) → return focus to host navigation
        Otherwise → translate to ANSI bytes → Write to master PTY
    → Plugin process reads from stdin (slave PTY)
```

---

## Detailed Design

### 1. Plugin Manifest Extension

Add optional `[tui]` section to plugin manifests. Plugins without this section have no TUI presence (current behavior).

```lua
-- init.lua (existing plugin manifest fields)
plugin.name = "analytics"
plugin.version = "1.0.0"
plugin.description = "Site analytics dashboard"

-- New: TUI screen registration
plugin.tui = {
    {
        name = "dashboard",
        label = "Analytics",        -- sidebar display name
        binary = "bin/analytics",   -- relative to plugin directory
        icon = "chart",             -- optional, for sidebar
        position = "after:media",   -- optional, ordering hint
    },
    -- A plugin can register multiple screens
    {
        name = "reports",
        label = "Reports",
        binary = "bin/analytics",
        args = {"--screen", "reports"},  -- subcommand pattern
    },
}
```

**Validation rules:**
- `binary` must exist and be executable
- `name` must be `[a-z0-9_]`, unique within the plugin
- `label` max 32 characters
- Maximum 4 TUI screens per plugin
- Binary must be a regular file (no symlinks to outside plugin directory)

### 2. PluginProcess (internal/plugin/tui_process.go)

Manages the lifecycle of a single plugin TUI process.

```go
type PluginProcess struct {
    pluginName string
    screenName string
    cmd        *exec.Cmd
    masterFd   *os.File           // PTY master
    vt         *vt.SafeEmulator   // thread-safe VT emulator
    apiSocket  string             // Unix socket path for plugin API
    width      int
    height     int
    exited     bool
    exitErr    error
    mu         sync.Mutex

    // Callback: send Bubbletea message when new output available
    onOutput   func()
}
```

**Lifecycle methods:**

```go
// Start launches the plugin binary attached to a PTY.
func (p *PluginProcess) Start(ctx context.Context) error

// Resize updates PTY dimensions and VT emulator.
// Atomic: both must update in the same call to prevent desync.
func (p *PluginProcess) Resize(cols, rows int)

// WriteInput sends raw bytes to the plugin process via PTY master.
func (p *PluginProcess) WriteInput(b []byte) error

// Render returns the current VT screen as an ANSI string.
func (p *PluginProcess) Render() string

// Stop sends SIGTERM, waits up to 3s, then SIGKILL.
func (p *PluginProcess) Stop() error

// Exited returns true if the process has terminated.
func (p *PluginProcess) Exited() bool
```

**Start implementation:**

```go
func (p *PluginProcess) Start(ctx context.Context) error {
    p.vt = vt.NewSafeEmulator(p.width, p.height)

    // Set up callbacks for terminal events
    p.vt.SetCallbacks(vt.Callbacks{
        Title: func(title string) {
            // Could update sidebar label dynamically
        },
    })

    p.cmd = exec.CommandContext(ctx, p.binaryPath, p.args...)
    p.cmd.Env = append(os.Environ(),
        fmt.Sprintf("MODULA_PLUGIN_WIDTH=%d", p.width),
        fmt.Sprintf("MODULA_PLUGIN_HEIGHT=%d", p.height),
        fmt.Sprintf("MODULA_PLUGIN_API=unix://%s", p.apiSocket),
        "TERM=xterm-256color",
    )

    var err error
    p.masterFd, err = pty.StartWithSize(p.cmd, &pty.Winsize{
        Rows: uint16(p.height),
        Cols: uint16(p.width),
    })
    if err != nil {
        return fmt.Errorf("start plugin PTY: %w", err)
    }

    // Read goroutine: master PTY → VT emulator
    go p.readLoop()

    // Wait goroutine: detect process exit
    go p.waitLoop()

    return nil
}
```

**Read loop:**

Following tmux's pattern of a single reader per pane. The goroutine reads from the master fd and feeds bytes into the VT emulator. When new output arrives, it notifies Bubbletea to re-render.

```go
func (p *PluginProcess) readLoop() {
    buf := make([]byte, 4096)
    for {
        n, err := p.masterFd.Read(buf)
        if n > 0 {
            p.vt.Write(buf[:n])  // SafeEmulator handles locking
            if p.onOutput != nil {
                p.onOutput()  // trigger Bubbletea re-render
            }
        }
        if err != nil {
            return  // EOF or error, process exited
        }
    }
}
```

**Resize:**

Atomic resize following tmux's pattern: update VT emulator dimensions AND PTY size in one call.

```go
func (p *PluginProcess) Resize(cols, rows int) {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.width = cols
    p.height = rows

    // 1. Resize VT emulator (reflows internal grid)
    p.vt.Resize(cols, rows)

    // 2. Resize PTY → kernel sends SIGWINCH to child process
    pty.Setsize(p.masterFd, &pty.Winsize{
        Rows: uint16(rows),
        Cols: uint16(cols),
    })
}
```

### 3. PluginScreenModel (internal/cli/plugin_screen.go)

A Bubbletea `tea.Model` that wraps `PluginProcess` and integrates with the host TUI's navigation system.

```go
type PluginScreenModel struct {
    process    *plugin.PluginProcess
    pluginName string
    screenName string
    focused    bool      // true when plugin owns input
    width      int       // content area width
    height     int       // content area height
    err        error     // set if process crashed
}
```

**Messages:**

```go
// PluginOutputMsg signals new output is available from a plugin process
type PluginOutputMsg struct {
    PluginName string
    ScreenName string
}

// PluginExitedMsg signals a plugin process has terminated
type PluginExitedMsg struct {
    PluginName string
    ScreenName string
    Err        error
}

// PluginStartMsg requests starting a plugin screen
type PluginStartMsg struct {
    PluginName string
    ScreenName string
}
```

**Update:**

```go
func (m PluginScreenModel) Update(msg tea.Msg) (PluginScreenModel, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        // Recalculate content area (subtract chrome)
        contentW := msg.Width - sidebarWidth
        contentH := msg.Height - topbarHeight - statusbarHeight
        m.width = contentW
        m.height = contentH
        if m.process != nil {
            m.process.Resize(contentW, contentH)
        }

    case tea.KeyMsg:
        if !m.focused {
            return m, nil  // host handles navigation
        }

        // Meta key: return focus to host
        if msg.Type == tea.KeyCtrlBackslash {
            m.focused = false
            return m, nil
        }

        // Forward everything else to plugin process
        ansiBytes := translateKey(msg)
        if len(ansiBytes) > 0 {
            m.process.WriteInput(ansiBytes)
        }

    case PluginOutputMsg:
        // New output available, Bubbletea will call View()
        return m, listenForOutput(m.process)

    case PluginExitedMsg:
        m.err = msg.Err
    }

    return m, nil
}
```

**View:**

```go
func (m PluginScreenModel) View() string {
    if m.err != nil {
        return renderPluginError(m.pluginName, m.err, m.width, m.height)
    }
    if m.process == nil {
        return renderPluginLoading(m.pluginName, m.width, m.height)
    }

    // VT emulator → ANSI string → Bubbletea
    screen := m.process.Render()

    // Add focus indicator border
    if m.focused {
        return lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("62")).
            Width(m.width).
            Height(m.height).
            Render(screen)
    }
    return screen
}
```

### 4. Input Translation (internal/cli/plugin_input.go)

Translate Bubbletea key messages into ANSI escape sequences for writing to the PTY. This is the inverse of what tmux's `tty-keys.c` does.

```go
var keyMap = map[tea.KeyType][]byte{
    tea.KeyUp:        []byte("\x1b[A"),
    tea.KeyDown:      []byte("\x1b[B"),
    tea.KeyRight:     []byte("\x1b[C"),
    tea.KeyLeft:      []byte("\x1b[D"),
    tea.KeyHome:      []byte("\x1b[H"),
    tea.KeyEnd:       []byte("\x1b[F"),
    tea.KeyPgUp:      []byte("\x1b[5~"),
    tea.KeyPgDown:    []byte("\x1b[6~"),
    tea.KeyDelete:    []byte("\x1b[3~"),
    tea.KeyBackspace: []byte{0x7f},
    tea.KeyTab:       []byte{0x09},
    tea.KeyEnter:     []byte{0x0d},
    tea.KeyEscape:    []byte{0x1b},
    tea.KeyF1:        []byte("\x1bOP"),
    tea.KeyF2:        []byte("\x1bOQ"),
    tea.KeyF3:        []byte("\x1bOR"),
    tea.KeyF4:        []byte("\x1bOS"),
    tea.KeyF5:        []byte("\x1b[15~"),
    // ... F6-F12
    tea.KeyCtrlA:     []byte{0x01},
    tea.KeyCtrlB:     []byte{0x02},
    tea.KeyCtrlC:     []byte{0x03},
    // ... Ctrl-D through Ctrl-Z
}

func translateKey(msg tea.KeyMsg) []byte {
    if msg.Type == tea.KeyRunes {
        return []byte(string(msg.Runes))
    }
    if b, ok := keyMap[msg.Type]; ok {
        return b
    }
    return nil
}
```

### 5. Plugin API Server (internal/plugin/tui_api_server.go)

Each plugin TUI process gets a Unix domain socket for accessing CMS data. This separates the data channel (socket) from the display channel (PTY).

```go
type PluginAPIServer struct {
    socketPath string
    listener   net.Listener
    driver     db.DbDriver
    pluginName string
}
```

**Protocol:** JSON-RPC 2.0 over Unix socket.

**Methods exposed:**

| Method | Description |
|--------|-------------|
| `content.list` | List content entries (with pagination, filtering) |
| `content.get` | Get single content entry by ID |
| `content.fields` | Get content fields for an entry |
| `datatypes.list` | List datatypes |
| `datatypes.get` | Get single datatype |
| `fields.list` | List field definitions |
| `media.list` | List media entries |
| `media.get` | Get single media entry |
| `users.current` | Get the authenticated user for this session |
| `config.get` | Get plugin-specific config values |
| `db.query` | Query plugin-owned tables (same sandbox as Lua API) |
| `db.execute` | Insert/update/delete on plugin-owned tables |

**Access control:** Same rules as the existing Lua plugin system. Plugin-owned tables have full access. Core tables use the existing three-tier whitelist (read+write, read-only, blocked columns).

**Plugin client libraries (future):** Provide thin client SDKs in Go, Python, and TypeScript that wrap the socket protocol. Plugin authors import the SDK and call typed methods instead of raw JSON-RPC.

### 6. Host TUI Integration

**Sidebar registration:**

During plugin loading, the manager reads `plugin.tui` entries from each manifest and registers them as navigation targets. The sidebar gets a "Plugins" section with entries for each registered screen.

```
┌─────────────────────────────────────┐
│ ModulaCMS                           │
├──────────┬──────────────────────────┤
│ Content  │                          │
│ Media    │  [Plugin Content Area]   │
│ Types    │                          │
│ Fields   │  Plugin process renders  │
│ Routes   │  here via PTY + VT       │
│ Users    │                          │
│ ──────── │                          │
│ Plugins  │                          │
│ > Analyt.│                          │
│   Reports│                          │
│ ──────── │                          │
│ Settings │                          │
└──────────┴──────────────────────────┘
```

**Focus model:**

Two focus states, following tmux's modal approach:

| State | Sidebar | Content Area | Status Bar |
|-------|---------|-------------|------------|
| **Host focused** | Navigable (arrow keys) | Plugin visible but not receiving input | Shows `Enter: focus plugin` |
| **Plugin focused** | Dimmed | Plugin receives all input | Shows `Ctrl-\: return to nav` |

When the user navigates to a plugin screen entry and presses Enter:
1. Host starts the plugin process (if not already running)
2. Content area shows the plugin's VT output
3. User is in host-focused mode (can still navigate sidebar)
4. Pressing Enter again switches to plugin-focused mode
5. `Ctrl-\` returns to host-focused mode

**Page registration:**

```go
// In page initialization, after plugin loading
for _, p := range mgr.TUIScreens() {
    pageIdx := allocatePluginPageIndex(p.PluginName, p.ScreenName)
    m.PageMap[pageIdx] = Page{
        Name:   p.Label,
        Index:  pageIdx,
        Type:   PluginPage,
        Plugin: p,
    }
    m.Pages = insertAfter(m.Pages, p.Position, m.PageMap[pageIdx])
}
```

### 7. Process Lifecycle

**Start strategies:**

| Strategy | When | Behavior |
|----------|------|----------|
| **Lazy** | User navigates to plugin screen | Start process on first visit, show loading state |
| **Eager** | Plugin has `eager = true` in manifest | Start at TUI boot, pre-warm for instant display |

Default is lazy. Plugin authors opt into eager start for screens that need warm-up time (data loading, connection setup).

**Backgrounding:**

When the user navigates away from a plugin screen:

| Option | Behavior | Memory | State |
|--------|----------|--------|-------|
| **Keep alive** (default) | Process continues running, output buffered | Higher | Preserved |
| **Suspend** | SIGTSTP to process, SIGCONT on return | Same | Preserved |
| **Kill** | SIGTERM, restart on return | Lowest | Lost |

Plugin manifest controls this via `lifecycle` field:
```lua
plugin.tui = {
    {
        name = "dashboard",
        label = "Analytics",
        binary = "bin/analytics",
        lifecycle = "keep_alive",  -- "keep_alive" | "suspend" | "restart"
    },
}
```

**Crash recovery:**

When a plugin process exits unexpectedly:
1. Read goroutine detects EOF on master PTY
2. `PluginExitedMsg` sent to Bubbletea
3. Content area shows error state:
   ```
   Plugin "analytics" exited unexpectedly.
   Exit code: 1

   [R] Restart    [B] Back to navigation
   ```
4. Circuit breaker: if a plugin crashes 3 times within 60 seconds, disable its TUI screen and show a persistent error. Admin must re-enable via CLI or admin panel.

### 8. SSH Context Considerations

The host TUI runs over SSH via Wish. Plugin processes run locally on the server. Key implications:

- Plugins cannot assume a local user is present
- Plugins cannot open GUI windows or access the client's filesystem
- The authenticated SSH user's identity is passed to the plugin via the API socket (not env vars, which could be forged by a malicious plugin)
- Terminal capabilities (color depth, Unicode support) come from the SSH client. The host detects these via Wish and passes them to the plugin via env vars (`COLORTERM`, `LANG`)

### 9. Security

**Binary sandboxing:**

Plugin binaries run as the same OS user as ModulaCMS. For stronger isolation:
- Future: support running plugin binaries in a container or namespace (Linux only)
- Immediate: validate that the binary path is within the plugin directory (no path traversal)
- The plugin API socket is the only data access channel — the binary cannot connect to the CMS database directly (it doesn't have credentials)

**Resource limits:**

| Resource | Limit | Enforcement |
|----------|-------|-------------|
| CPU | Not enforced (OS scheduling) | Monitor via process metrics |
| Memory | Not enforced initially | Future: cgroup limits on Linux |
| Output rate | 1 MB/s from PTY | Read loop throttle with backpressure |
| API calls | 100/s per socket | Rate limiter on API server |
| Disk | Plugin directory only (by convention) | Future: namespace/chroot |
| Network | Unrestricted | Future: network namespace |

**Output rate limiting (flow control):**

Following tmux's pattern: if the VT emulator's pending output exceeds a threshold, pause reading from the master PTY. This prevents a fast-producing plugin from consuming unbounded memory in the VT buffer.

```go
const maxPendingBytes = 1 << 20  // 1 MB

func (p *PluginProcess) readLoop() {
    buf := make([]byte, 4096)
    for {
        // Backpressure: pause if VT buffer is large
        for p.vt.PendingBytes() > maxPendingBytes {
            time.Sleep(10 * time.Millisecond)
        }

        n, err := p.masterFd.Read(buf)
        if n > 0 {
            p.vt.Write(buf[:n])
            if p.onOutput != nil {
                p.onOutput()
            }
        }
        if err != nil {
            return
        }
    }
}
```

---

## Implementation Phases

### Phase 1: PTY + VT Foundation

**Goal:** Embed a single hardcoded binary in the TUI content area.

1. Add `creack/pty` and `charmbracelet/x/vt` to vendor
2. Implement `PluginProcess` (start, read loop, resize, stop, render)
3. Implement `PluginScreenModel` (Update, View, focus toggle)
4. Implement `translateKey` for the key map
5. Add a test page that launches a known binary (e.g., `htop` or a simple Bubbletea demo) in the content area
6. Verify: resize, input forwarding, alt screen, color rendering, process exit handling

**Deliverable:** A working proof of concept where navigating to a test page shows an embedded TUI process.

### Phase 2: Manifest + Navigation

**Goal:** Plugin manifests declare TUI screens, which appear in the sidebar.

1. Add `tui` field to plugin manifest parsing and validation
2. Register plugin screens as navigation targets during plugin loading
3. Add plugin screen entries to the sidebar under a "Plugins" section
4. Implement lazy start (start process on first navigation)
5. Implement keep-alive behavior (process persists when navigating away)
6. Implement crash recovery UI and circuit breaker

**Deliverable:** Plugins can declare TUI screens in their manifest and users can navigate to them from the sidebar.

### Phase 3: Plugin API Socket

**Goal:** Plugin processes can access CMS data.

1. Implement `PluginAPIServer` with JSON-RPC over Unix socket
2. Expose content, datatypes, fields, media, users, and config methods
3. Enforce same access control rules as Lua plugin system
4. Plugin-owned table access via sandboxed query builder
5. Pass socket path to plugin process via `MODULA_PLUGIN_API` env var
6. Implement a Go client SDK package that plugins can import

**Deliverable:** Plugin TUI processes can query CMS data through a typed API.

### Phase 4: Polish + Production Hardening

**Goal:** Production-ready with security and operational controls.

1. Output rate limiting (flow control / backpressure)
2. Process metrics (CPU, memory, uptime) visible in plugin detail page
3. Admin controls: force-stop a plugin TUI process from admin panel or CLI
4. Eager start option in manifest
5. Suspend/restart lifecycle options
6. SSH context propagation (user identity, terminal capabilities)
7. Documentation: plugin author guide for building TUI screens
8. Example plugin: a Bubbletea-based analytics dashboard

### Future

- Python and TypeScript client SDKs for the plugin API socket
- Container-based isolation for plugin binaries (Linux)
- Bubbletea v2 integration: switch from `Render()` string to `Draw()` cell compositing
- Plugin-to-plugin communication via named channels on the API socket
- Mouse event forwarding (depends on Bubbletea mouse support in SSH context)

---

## File Map

```
internal/plugin/
    tui_process.go         # PluginProcess: PTY lifecycle, read loop, VT emulator
    tui_process_test.go    # Tests with a simple echo binary
    tui_manifest.go        # TUI manifest field parsing and validation
    tui_manifest_test.go
    tui_api_server.go      # Unix socket JSON-RPC server
    tui_api_server_test.go
    tui_api_methods.go     # RPC method implementations
    tui_api_client.go      # Go client SDK for plugin authors

internal/cli/
    plugin_screen.go       # PluginScreenModel: Bubbletea model
    plugin_input.go        # Key translation: tea.KeyMsg → ANSI bytes
    plugin_input_test.go   # Translation table tests
```

---

## Open Questions

1. **Meta key choice:** `Ctrl-\` is proposed for returning focus to host navigation. Alternatives: `Ctrl-]` (conflicts with telnet), `Esc Esc` (double-tap, 300ms timeout like tmux escape-time), or configurable in user settings. What feels right?

2. **Multiple concurrent plugins:** Can two plugin screens run simultaneously (split view)? Phase 1-4 assume one active plugin screen at a time. Split view would require a layout manager similar to tmux's layout tree. Worth considering but likely a separate effort.

3. **Plugin API authentication:** The Unix socket is per-process and file-permission restricted. Is this sufficient, or should the socket protocol include a token handshake? A token would prevent a rogue process from connecting to another plugin's socket.

4. **Bubbletea v2 timeline:** `charmbracelet/x/vt` has first-class v2 support via `Draw()`. When ModulaCMS upgrades to Bubbletea v2, the PluginScreenModel switches from `Render()` string to `Draw()` cell compositing for better performance. No architectural changes needed — just swap the View implementation.

5. **Plugin binary distribution:** How do plugin authors distribute compiled binaries for multiple architectures (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)? The manifest could support architecture-specific binary paths, or plugins could use Go's cross-compilation to ship fat plugin directories.
