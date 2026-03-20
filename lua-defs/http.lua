---@meta

--- HTTP routing module. Registers plugin HTTP routes and middleware.
--- All routes are served under `/api/v1/plugins/<plugin_name>/`.
--- Must be called at MODULE SCOPE (not inside `on_init()`).
---@class http
http = {}

---@alias http.Method
---| "GET"
---| "POST"
---| "PUT"
---| "DELETE"
---| "PATCH"

---@class http.Request
---@field method string HTTP method (e.g., `"GET"`).
---@field path string Full URL path.
---@field body string Raw request body as string.
---@field client_ip string Client IP address (proxy-aware).
---@field headers table<string, string> Request headers (lowercase keys).
---@field query table<string, string> Query parameters.
---@field params table<string, string> Path parameters from `{param}` segments.
---@field json table|nil Parsed JSON body (only when Content-Type is application/json).

---@class http.Response
---@field status? integer HTTP status code (default 200).
---@field json? table Table to serialize as JSON response.
---@field body? string Raw string body (used only if `json` is nil).
---@field headers? table<string, string> Response headers (security headers blocked).

---@class http.RouteOpts
---@field public? boolean If true, bypass CMS authentication (default false).

--- Register an HTTP route handler.
--- Max 50 routes per plugin. Path must start with `/`, max 256 chars.
--- Supports Go 1.22+ path parameters via `{param}` syntax.
---@param method http.Method
---@param path string Route path (e.g., `"/tasks/{id}"`).
---@param handler fun(req: http.Request): http.Response
---@param opts? http.RouteOpts
function http.handle(method, path, handler, opts) end

--- Register middleware that runs for ALL plugin routes.
--- Call `next()` to pass to the next middleware or handler.
---@param middleware fun(req: http.Request, res: http.Response, next: fun())
function http.use(middleware) end
