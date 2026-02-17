Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Plan: @modulacms/plugin-sdk + Go UI Metadata

 Context

 ModulaCMS has a mature Lua plugin system (Phases 1-4 complete) with DB, HTTP routes, lifecycle hooks, and admin management. Plugins can expose REST APIs but have no way to declare or serve
 admin UIs. There is no web frontend in this repo — it's a headless backend.

 This plan adds a @modulacms/plugin-sdk TypeScript package that plugin authors install to build Web Component-based UIs, plus Go-side changes to support UI metadata in the plugin manifest
 and API responses. Plugin authors call definePlugin(), bundle the result, and serve it via their Lua HTTP routes. External consuming apps (admin panels, dashboards) load the bundle and
 mount the custom element.

 Part 1: TypeScript SDK Package

 New package: sdks/typescript/plugin-sdk/

 Files to create:

 sdks/typescript/plugin-sdk/
   package.json
   tsconfig.json
   tsup.config.ts
   vitest.config.ts
   src/
     index.ts           # Public exports
     define-plugin.ts   # definePlugin() + custom element class
     context.ts         # PluginContext factory, API client
     configure.ts       # Module-level configure() for API key injection
     theme.ts           # Theme tokens, CSS injection
     errors.ts          # PluginError, PluginApiError, type guards
     json.ts            # JsonValue type and serialization guard
     types.ts           # All public type definitions

 Files to modify:

 - sdks/typescript/pnpm-workspace.yaml — add 'plugin-sdk' entry

 package.json

 Zero runtime dependencies (no @modulacms/types — there is no type overlap: no branded IDs, entity types,
 or pagination). Plugin API responses are opaque — consumers provide their own response types via generics.
 Same structure as @modulacms/admin-sdk package.json, but with no dependencies field:
 - name: @modulacms/plugin-sdk
 - version: 0.1.0
 - type: module
 - ESM+CJS dual exports with .d.ts/.d.cts
 - sideEffects: false (correct: definePlugin is consumer-invoked, not called at import time)
 - engines.node: >=18
 - Scripts: build (tsup), typecheck (tsc --noEmit), test (vitest run), clean
 - No dependencies field (zero runtime dependencies)
 - devDependencies: jsdom (required for vitest jsdom environment; not in the workspace root)

 Note: typescript, tsup, and vitest are already workspace-root devDependencies (inherited).
 Only jsdom needs to be added as a package-level devDependency. Use the same version constraint
 style as the workspace root (e.g., "jsdom": "^26.0.0" or latest stable).

 Browser compatibility target: Custom Elements v1 (Chrome 67+, Firefox 63+, Safari 10.1+)
 for the core custom element lifecycle. However, AbortSignal.any() (used for signal composition)
 requires Chrome 116+, Firefox 124+, Safari 17.4+. This narrows the effective minimum. If broader
 support is needed, implement a manual AbortController-based composition that listens for abort on
 both signals instead of using AbortSignal.any().
 No polyfills included; consumers targeting older browsers must provide their own.
 Target bundle size: <5 KB minified+gzipped (zero dependencies, no framework).

 tsconfig.json / tsup.config.ts

 Identical to admin-sdk. Extends ../tsconfig.base.json (already has "lib": ["ES2022", "DOM"]).

 vitest.config.ts

 Same as admin-sdk but with environment: "jsdom" for DOM API access (HTMLElement, customElements, document).

 Core API: definePlugin(definition)

 type PluginDefinition = {
   tag: string                                          // custom element tag (must contain hyphen)
   setup: (ctx: PluginContext, el: HTMLElement) => void | Promise<void>   // one-time init
   mount: (ctx: PluginContext, el: HTMLElement) => void | Promise<void>   // mount UI into element (called once)
   destroy?: (el: HTMLElement) => void                  // cleanup on disconnect
 }

 definePlugin() internally:
 1. Validates tag (must start with lowercase letter, contain a hyphen, lowercase alphanumeric + hyphens only,
    max 64 chars, not a reserved HTML name, not already registered).
    If tag is already registered, throw with message:
    "Tag '<tag>' already registered. Hot-reload is not supported — reload the page to re-register."
 2. Creates a class extending HTMLElement
 3. Calls customElements.define(tag, cls)

 Reserved HTML custom element names (reject these): annotation-xml, color-profile, font-face,
 font-face-src, font-face-name, font-face-format, font-face-uri, missing-glyph.

 Note: Tag validation rules are duplicated in Go ValidateManifest. Both implementations must stay
 in sync — see "Tag Validation Shared Rules" section below.

 Custom Element Lifecycle (Light DOM, no Shadow DOM)

 - connectedCallback → read attributes → validate required attributes → build context → inject theme → setup(ctx, el) → mount(ctx, el)
 - disconnectedCallback → destroy(el) → clear theme properties → clear innerHTML → release refs
 - Guard against double-init with _initialized flag
 - Check this.isConnected after each await (handles disconnect during async setup)
 - Elements are single-use: once disconnectedCallback fires, the element cannot be re-mounted.
   Re-appending a previously disconnected element to the DOM is a no-op (the _initialized guard
   prevents re-initialization and the AbortSignal is already aborted). Plugin authors who need
   to move elements must create a new instance. This is documented in the SDK README.

 CSS Isolation Strategy (Light DOM)

 Light DOM means no browser-enforced style encapsulation. Plugin CSS can leak into the host page
 and vice versa. The SDK mitigates this with a convention-based approach:

 1. All CSS custom properties use the --mcms- namespace (theme tokens). No collisions with host styles.
 2. Plugin authors MUST scope all selectors to the custom element tag. The SDK injects a
    data-mcms-plugin attribute on the host element during connectedCallback. Plugin authors should
    use this for selector scoping: [data-mcms-plugin="my_plugin"] .my-class { ... }
 3. The SDK sets CSS containment on the host element: contain: layout style (prevents layout and
    counter-reset leakage). contain: paint is intentionally omitted to allow overflow (dropdowns, tooltips).
 4. Documentation and the starter template (future) will recommend CSS Modules or tag-prefixed
    class names (e.g., mcms-my-plugin__header). BEM with the tag name as block is the recommended
    convention.

 This is a convention, not enforcement. Plugin marketplace publication (future) can lint for
 unscoped selectors as a quality gate.

 Context Object

 Built from HTML attributes on the host element:

 <mcms-my-plugin
   base-url="https://cms.example.com"
   plugin-name="my_plugin"
   auth-user-id="01HXK..."
   auth-username="admin"
   auth-role="admin"
   theme='{"colors":{"primary":"oklch(0.7 0.15 250)"}}'
 />

 Required attributes: base-url and plugin-name MUST be present on the host element. If either is
 missing, the context factory throws a PluginError during connectedCallback with message:
   "Missing required attribute '<attr-name>' on <tag>. Plugin cannot initialize."
 This error is caught by the lifecycle error handler and rendered as the [role="alert"] fallback.

 type PluginContext = {
   pluginName: string
   baseUrl: string
   api: PluginApiClient        // scoped fetch: paths prefixed with /api/v1/plugins/<name>
   auth: PluginAuth | null     // from auth-* attributes, null if absent
   theme: ThemeTokens          // OKLCH color tokens + spacing + radii + typography
   onNavigate: (handler: (path: string) => void) => () => void  // subscribe
   navigate: (path: string) => void                              // dispatch
   element: HTMLElement
   signal: AbortSignal         // aborted on disconnectedCallback — pass to fetch or listen for cleanup
 }

 /** JSON-serializable value. Prevents passing functions, symbols, or circular structures. */
 type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue }

 /**
  * Scoped API client for plugin HTTP routes.
  *
  * Generic parameter T on get/post/put is a trust-based cast — the SDK does NOT validate
  * response shapes at runtime. Consumers are responsible for verifying that API responses
  * match T. For runtime validation, use raw() and validate the Response body yourself.
  */
 type PluginApiClient = {
   /** GET request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
   get: <T = unknown>(path: string, params?: Record<string, string>, signal?: AbortSignal) => Promise<T>
   /** POST request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
   post: <T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal) => Promise<T>
   /** PUT request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
   put: <T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal) => Promise<T>
   /** PATCH request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
   patch: <T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal) => Promise<T>
   /** DELETE request. Throws PluginApiError on non-2xx. Returns void — error info only via thrown PluginApiError. */
   del: (path: string, params?: Record<string, string>, signal?: AbortSignal) => Promise<void>
   /** Raw fetch. Does NOT throw on non-2xx — caller must check response.ok. */
   raw: (path: string, init?: RequestInit) => Promise<Response>
 }

 Auth strategy: if configure() has been called with an API key → Authorization: Bearer header.
 Otherwise → credentials: 'include' for cookies.

 There is NO api-key DOM attribute. API keys must not be exposed in the DOM where any script on
 the page can read them. The only way to inject an API key is via the configure() module-level API.

 configure() API — module-level export for injecting API keys:

 type PluginConfig = {
   apiKey: string
 }

 function configure(config: PluginConfig): void
 function resetConfig(): void

 File: src/configure.ts. Stores the API key in a module-scoped closure (not exported). The context
 factory checks this closure; if no key is set, falls back to cookie auth (credentials: 'include').

 Resolution order: configure() closure > cookie auth (credentials: 'include').

 resetConfig() clears the module-scoped closure. Exported for:
 - Test isolation (call in afterEach to prevent cross-test leakage)
 - Multi-tenant scenarios where the host app switches between tenants with different API keys

 PluginApiClient.raw() behavior: raw() does NOT throw on non-2xx responses — it returns the raw Response
 object. The caller is responsible for checking response.ok. This differs from get/post/put/patch/del
 which throw PluginApiError on non-2xx. Document this distinction in the type definition via JSDoc comment.

 Signal composition: When a per-request signal is provided to get/post/put/patch/del, it is composed
 with the context signal via AbortSignal.any([ctx.signal, requestSignal]). Either signal aborting
 cancels the request. This means the context signal always has authority to cancel (element removed
 from DOM), while per-request signals add additional cancellation triggers (e.g., user-initiated cancel).
 raw() does not compose signals — the caller manages their own signal via RequestInit.

 Navigation via CustomEvents: auto-namespaced per plugin name. Event names use the pattern
 mcms-navigate:<plugin-name> (outbound, bubbles) and mcms-navigate-to:<plugin-name> (inbound from host).
 Plugin names are unique (enforced on install), so no cross-plugin event leakage.
 The context factory reads plugin-name from the host element attribute and bakes it into event dispatch/listen.

 Theme Contract

 Default OKLCH tokens injected as CSS custom properties on the host element:

 ┌────────────┬────────────────────────────────────────────────────────────────────────────┬──────────────────────┐
 │  Category  │                                 Properties                                 │       Example        │
 ├────────────┼────────────────────────────────────────────────────────────────────────────┼──────────────────────┤
 │ Colors     │ --mcms-color-primary, -surface, -border, -text, -error, -success, -warning │ oklch(0.65 0.15 250) │
 ├────────────┼────────────────────────────────────────────────────────────────────────────┼──────────────────────┤
 │ Spacing    │ --mcms-spacing-xs through -2xl                                             │ 0.25rem to 3rem      │
 ├────────────┼────────────────────────────────────────────────────────────────────────────┼──────────────────────┤
 │ Radii      │ --mcms-radius-sm, -md, -lg, -full                                          │ 0.25rem to 9999px    │
 ├────────────┼────────────────────────────────────────────────────────────────────────────┼──────────────────────┤
 │ Typography │ --mcms-font-family, --mcms-font-size-*, --mcms-line-height                 │ system-ui, ...       │
 └────────────┴────────────────────────────────────────────────────────────────────────────┴──────────────────────┘

 Host overrides via theme JSON attribute (recursive deep-merged with defaults) or mcms-theme-changed CustomEvent.
 Deep-merge semantics: recursive merge per nested key. A partial colors override (e.g., only "primary")
 merges into defaults — unspecified tokens (surface, border, text, etc.) retain their default values.
 Only explicitly provided keys are overwritten.

 Malformed theme JSON: If the theme attribute is present but contains invalid JSON, log a
 console.warn with: "Invalid theme JSON on <tag>, using defaults: <error.message>" and fall
 back to DEFAULT_THEME. Do not throw — the plugin should still mount with default theme tokens.

 Error Handling

 - Required attribute missing → throw PluginError("Missing required attribute '<name>' on <tag>. Plugin cannot initialize.")
 - Setup/mount errors → dispatch mcms-error CustomEvent (bubbles, composed) + render minimal [role="alert"] fallback + console.error
 - Destroy errors → log + dispatch event, but still clean up DOM
 - API errors → throw PluginApiError with _tag, status, message
 - Exported type guards: isPluginError(), isPluginApiError()

 Public Exports

 // Runtime
 export { definePlugin } from './define-plugin.js'
 export { configure, resetConfig } from './configure.js'
 export { isPluginError, isPluginApiError } from './errors.js'
 export { DEFAULT_THEME } from './theme.js'

 // Types
 export type { PluginDefinition, PluginContext, PluginApiClient, PluginAuth, ThemeTokens, PluginConfig }
 export type { PluginError, PluginApiError }
 export type { JsonValue } from './json.js'

 Tests (jsdom environment)

 | File                   | Coverage                                                                                             |
 |------------------------|------------------------------------------------------------------------------------------------------|
 | define-plugin.test.ts  | Registration, lifecycle sequencing, async setup/mount, double-init guard, disconnect during async, signal aborted on disconnect, re-append after disconnect is no-op, data-mcms-plugin attribute set, CSS containment applied |
 | context.test.ts        | Attribute reading, required attr validation (throws on missing base-url/plugin-name), API client path prefixing, configure() auth header vs cookie credentials, navigation events, AbortSignal cancellation, AbortSignal.any composition with per-request signal, patch method |
 | configure.test.ts      | Module-level configure(), resetConfig() clears key, cookie fallback when no key, reset between tests |
 | theme.test.ts          | CSS property injection/removal, default theme, JSON attribute merge, theme-changed event              |
 | errors.test.ts         | Error event dispatch, fallback UI rendering, type guards, destroy-error cleanup                       |
 | json.test.ts           | JsonValue type acceptance (primitives, arrays, objects), rejection of non-serializable values at type level |

 Mock fetch with vi.stubGlobal. Helper mountPlugin(tag, attrs) to create + append + return cleanup.

 ---
 Part 2: Go-Side Changes

 Extend PluginInfo struct

 File: internal/plugin/manager.go

 Add new struct and field:

 type PluginUIInfo struct {
     Tag      string // custom element tag (e.g., "mcms-task-tracker")
     Bundle   string // opaque metadata: informational path to JS bundle (not validated, resolved, or served by Go)
     HasAdmin bool   // whether plugin provides admin UI
 }

 No JSON struct tags on PluginUIInfo — this is an internal struct like PluginInfo. JSON
 serialization happens through handler-local structs (pluginUIJSON) in cli_commands.go.
 Do not add json tags here.

 // Add to PluginInfo:
 UI *PluginUIInfo // nil if plugin has no UI declaration

 Extract UI metadata in ExtractManifest

 File: internal/plugin/manager.go (after the dependencies extraction block in ExtractManifest)

 Read optional ui table from plugin_info:
 - ui.tag → string (required if ui table present)
 - ui.bundle → string (optional)
 - ui.has_admin → bool (optional)
 - Only attach UI if tag is non-empty

 Add UI validation in ValidateManifest

 File: internal/plugin/manager.go (extend ValidateManifest)

 When info.UI != nil:
 - Tag must start with a lowercase ASCII letter
 - Tag must contain at least one hyphen (Web Components requirement)
 - Tag must be lowercase alphanumeric + hyphens only
 - Tag max 64 characters
 - Tag must not be a reserved HTML custom element name (annotation-xml, color-profile, font-face,
   font-face-src, font-face-name, font-face-format, font-face-uri, missing-glyph)
 - Bundle is opaque metadata (string). Go does NOT validate the path, resolve it, or use it to
   read/serve files. It is surfaced in API responses for consuming admin panels to display.
   No path traversal check is needed because Go never acts on this value.
 - Route warning: After a plugin reaches StateRunning (end of loadPlugin in manager.go, after
   line `inst.State = StateRunning`), if inst.Info.UI is non-nil and the Manager's HTTPBridge
   is non-nil, filter bridge.ListRoutes() for routes where PluginName matches. If zero routes
   found, log a warning via utility.DefaultLogger.Warn:
   `plugin %q declares UI tag %q but has no registered HTTP routes to serve the bundle`
   This does not block loading — the plugin may register routes dynamically or rely on external
   serving.

 Tag Validation Shared Rules

 The following tag rules are enforced in both Go (ValidateManifest) and TypeScript (definePlugin).
 Both implementations MUST stay in sync. When modifying, update both and add matching test cases.
 - Must start with a lowercase ASCII letter [a-z]
 - Must contain at least one hyphen
 - Only lowercase alphanumeric [a-z0-9] and hyphens allowed
 - Max 64 characters
 - Must not be a reserved HTML custom element name (8 names listed above)

 Shared test fixture: testdata/tag-validation-cases.json

 Canonical location: testdata/tag-validation-cases.json at the project root (alongside go.mod).
 This is a NEW top-level testdata/ directory at the project root, NOT the existing
 internal/plugin/testdata/ directory used for Lua plugin fixtures.

 Both test suites reference this single file:
 - Go tests: os.ReadFile("../../testdata/tag-validation-cases.json") from internal/plugin/
 - TypeScript tests: fs.readFileSync("../../../../testdata/tag-validation-cases.json") from plugin-sdk/src/
   (src/ → plugin-sdk/ → typescript/ → sdks/ → project root)

 Format: array of { tag: string, valid: boolean, reason?: string }. Both test suites iterate
 this file to verify identical validation behavior. When adding a new rule, add test cases to this
 file first, then update both implementations until all cases pass.

 Fixture content (create this file with these initial cases):
 [
   { "tag": "mcms-tasks",          "valid": true,  "reason": "valid: lowercase with hyphen" },
   { "tag": "my-plugin",           "valid": true,  "reason": "valid: simple two-part tag" },
   { "tag": "a-b",                 "valid": true,  "reason": "valid: minimal tag" },
   { "tag": "mcms-task-tracker-2", "valid": true,  "reason": "valid: digits allowed after first char" },
   { "tag": "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-1234", "valid": true, "reason": "valid: exactly 64 characters" },
   { "tag": "nohyphen",            "valid": false, "reason": "no hyphen" },
   { "tag": "1-starts-with-digit", "valid": false, "reason": "starts with digit, not lowercase letter" },
   { "tag": "-starts-with-hyphen", "valid": false, "reason": "starts with hyphen, not lowercase letter" },
   { "tag": "My-Plugin",           "valid": false, "reason": "uppercase characters not allowed" },
   { "tag": "my_plugin",           "valid": false, "reason": "underscores not allowed" },
   { "tag": "my plugin",           "valid": false, "reason": "spaces not allowed" },
   { "tag": "",                    "valid": false, "reason": "empty string" },
   { "tag": "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-12345", "valid": false, "reason": "65 characters, exceeds max 64" },
   { "tag": "annotation-xml",      "valid": false, "reason": "reserved HTML custom element name" },
   { "tag": "color-profile",       "valid": false, "reason": "reserved HTML custom element name" },
   { "tag": "font-face",           "valid": false, "reason": "reserved HTML custom element name" },
   { "tag": "font-face-src",       "valid": false, "reason": "reserved HTML custom element name" },
   { "tag": "font-face-name",      "valid": false, "reason": "reserved HTML custom element name" },
   { "tag": "font-face-format",    "valid": false, "reason": "reserved HTML custom element name" },
   { "tag": "font-face-uri",       "valid": false, "reason": "reserved HTML custom element name" },
   { "tag": "missing-glyph",       "valid": false, "reason": "reserved HTML custom element name" }
 ]

 CI enforcement: .github/workflows/sdks.yml adds a step that runs both Go and TypeScript tag
 validation tests whenever testdata/tag-validation-cases.json is modified. The Go CI workflow
 (.github/workflows/go.yml) also runs these tests on changes to testdata/.

 Surface UI in API responses

 File: internal/plugin/cli_commands.go

 PluginListHandler — add HasAdminUI bool field to pluginJSON.
 Prerequisite: PluginInfo must already have the UI field added (see "Extend PluginInfo struct" above).
 HasAdminUI bool `json:"has_admin_ui,omitempty"`
 // Set from: inst.Info.UI != nil && inst.Info.UI.HasAdmin

 PluginInfoHandler — add nested UI field to the locally-scoped infoJSON struct inside PluginInfoHandler
 (this struct is declared inside the handler function, not at package level):
 type pluginUIJSON struct {
     Tag      string `json:"tag"`
     Bundle   string `json:"bundle,omitempty"`
     HasAdmin bool   `json:"has_admin"`
 }
 // Add to the existing infoJSON struct: UI *pluginUIJSON `json:"ui,omitempty"`
 // Prerequisite: PluginInfo must already have the UI field added (see "Extend PluginInfo struct" above).

 Go tests

 File: internal/plugin/manager_test.go

 - TestExtractManifest_WithUI — valid ui table extraction
 - TestExtractManifest_UIWithoutTag — ui table ignored when tag empty
 - TestValidateManifest_UITagNoHyphen — validation rejects tag without hyphen
 - TestValidateManifest_UITagStartsWithDigit — validation rejects tag starting with non-letter
 - TestValidateManifest_UITagReservedName — validation rejects reserved HTML element names
 - TestValidateManifest_UITagSharedFixture — iterates testdata/tag-validation-cases.json

 File: internal/plugin/cli_commands_test.go

 - Extend TestPluginInfoHandler_Found to verify ui JSON field when UI is set

 ---
 Part 3: TypeScript Admin SDK Type Update

 File: sdks/typescript/modulacms-admin-sdk/src/types/plugins.ts

 Add:
 export type PluginUIInfo = {
   tag: string
   bundle?: string
   has_admin: boolean
 }

 Extend PluginInfo: add ui?: PluginUIInfo
 Extend PluginListItem: add has_admin_ui?: boolean

 Note: The base tsconfig has exactOptionalPropertyTypes: true. Consumers who construct PluginInfo
 objects in tests must either include ui with a valid PluginUIInfo value or omit the property
 entirely — they cannot write ui: undefined. This is minor test friction, not a breaking change.

 File: sdks/typescript/modulacms-admin-sdk/src/index.ts

 Add PluginUIInfo to the plugin type re-exports.

 ---
 Design Decisions

 1. No route validation on install: If a plugin declares ui.tag but registers no HTTP routes to serve the
    bundle, the Go install phase logs a warning. It does not block installation — the plugin may register
    routes dynamically or rely on external serving. Routes are exposed at install time for admin approval;
    plugin authors are responsible for registering routes that serve their bundle.

 2. API versioning: Versioned API means old API versions continue to exist while new ones are deployed.
    No breaking changes to existing endpoints — only deprecation warnings. Plugin bundles targeting
    /api/v1/plugins/ will continue to work when v2 is introduced. The plugin-sdk package follows semver;
    v1 API surface will always exist with sufficient deprecation window for adoption of new versions.

 3. Hot-reload not supported: customElements.define() is irreversible per the Web Components spec. If a
    plugin bundle is reloaded, the tag is already registered. definePlugin() throws with a clear message
    directing the user to reload the page. This is a known limitation, not a bug.

 4. jsdom test limitations: jsdom fires connectedCallback synchronously, unlike real browsers (microtask).
    The isConnected-after-await tests verify the guard code exists but cannot exercise the actual race
    condition. Manual browser testing is recommended for lifecycle edge cases.

 5. mount vs render naming: The mount callback is called exactly once during connectedCallback. It is NOT
    re-invoked on theme changes or other updates. The name "mount" signals single-invocation semantics
    (consistent with Web Component lifecycle). Plugins that need to react to theme changes should subscribe
    via ctx.element.addEventListener('mcms-theme-changed', handler) inside setup and manage their own
    re-rendering. This avoids framework-like abstractions that would increase SDK complexity.

 6. AbortSignal on disconnect: The context exposes a signal (AbortSignal) that is aborted when
    disconnectedCallback fires. Plugin authors should pass ctx.signal to fetch calls and long-running
    operations inside setup/mount. This prevents resource leaks from in-flight requests when the
    element is removed from the DOM. The API client methods compose per-request signals with the
    context signal via AbortSignal.any() — either aborting cancels the request.

 7. API responses are opaque: Plugin API responses are typed as generic T with no runtime validation.
    The SDK does not depend on @modulacms/types because plugin HTTP routes return plugin-defined
    payloads, not CMS entities. Consumers are responsible for their own response types and validation.

 8. No api-key DOM attribute: API keys are injected exclusively via the configure() module-level API.
    DOM attributes are readable by any script on the page, making them an XSS vector for credential
    harvesting. The configure() closure keeps the key in JavaScript memory only.

 9. CSS isolation is convention-based: Light DOM is used for slot/form participation and simpler
    interop with host frameworks. CSS containment (layout, style) is applied programmatically.
    Plugin authors must scope their selectors (BEM with tag name, or data-mcms-plugin attribute).
    Future plugin marketplace can enforce scoping via automated linting.

 10. Tag uniqueness: Element tags are derived from plugin names (e.g., plugin "task_tracker" uses tag
     "mcms-task-tracker"). Plugin names are unique at the database level (enforced on install), so
     cross-plugin tag collisions cannot occur through normal installation.

 11. Bundle URL discovery: The Go API returns ui.bundle as opaque metadata. The actual served URL
     depends on Lua HTTP routes registered by the plugin, which are exposed at install time for
     admin approval. The consuming admin panel resolves bundle URLs via the plugin's registered
     routes, not the bundle metadata field.

 12. Theme and dark mode: The SDK provides a mcms-theme-changed CustomEvent hook. The host
     application (admin panel) is responsible for dispatching theme change events. Plugin authors
     are responsible for handling theme transitions. A future plugin marketplace can set standards
     for theme compliance as a publication requirement.

 13. Elements are single-use: Once disconnectedCallback fires, the element cannot be re-mounted.
     The _initialized guard and aborted signal prevent re-initialization. This avoids complex
     state management for re-entrant lifecycle and matches the typical Web Component usage pattern.

 14. Bundle path is opaque: The ui.bundle field in the Go manifest is informational metadata.
     Go never reads, resolves, or serves the file at this path. No path validation is performed
     because the value has no security-relevant use on the server side.

 ---
 Implementation Order

 1. Shared test fixture — create testdata/tag-validation-cases.json at project root
 2. Package scaffold — create plugin-sdk/ directory, config files, workspace entry
 3. SDK source — json.ts → types.ts → errors.ts → configure.ts → theme.ts → context.ts → define-plugin.ts → index.ts
 4. SDK tests — all test files consuming shared fixture, verify passing
 5. Go changes — PluginUIInfo struct, ExtractManifest, ValidateManifest, API handlers, Go tests (consuming shared fixture)
 6. Admin SDK types — PluginUIInfo type, PluginInfo/PluginListItem extensions
 7. CI update — add shared fixture trigger to both workflows

 Steps 3-4 and 5-6 can run concurrently (after step 1 completes).

 No justfile changes needed — existing `just sdk-build` runs `pnpm build` at the workspace root,
 which picks up new workspace packages automatically once pnpm-workspace.yaml is updated (step 2).
 Similarly, `just sdk-test` and `just sdk-install` work unchanged.

 Verification

 1. cd sdks/typescript && pnpm install && pnpm build — all packages build including plugin-sdk
 2. cd sdks/typescript/plugin-sdk && pnpm test — all jsdom tests pass
 3. cd sdks/typescript && pnpm typecheck — all packages typecheck
 4. go test -v ./internal/plugin/... — Go tests pass including new UI metadata tests
 5. just check — compile check passes
