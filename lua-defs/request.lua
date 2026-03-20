---@meta

--- Outbound HTTP request module. Makes requests to external domains.
--- Domains must be pre-registered at module scope and approved by admin.
---@class request
request = {}

---@class request.Opts
---@field headers? table<string, string> Custom request headers.
---@field body? string Raw request body (mutually exclusive with `json`).
---@field json? table Table to serialize as JSON body (mutually exclusive with `body`).
---@field timeout? number Request timeout in seconds.
---@field parse_json? boolean Parse response body as JSON into `response.json`.

---@class request.Response
---@field status integer HTTP status code.
---@field body string Response body as string.
---@field headers table<string, string> Response headers.
---@field json? table Parsed JSON (only if `parse_json` was true).
---@field error? string Error message if the request failed.

---@class request.RegisterOpts
---@field description? string Human-readable description of why this domain is needed.

--- Register an external domain for outbound access.
--- Must be called at MODULE SCOPE (not inside `on_init()`). Max 50 domains.
---@param domain string Domain name (e.g., `"api.example.com"`).
---@param opts? request.RegisterOpts
function request.register(domain, opts) end

--- Send an HTTP request.
--- Cannot be called in `on_init()`, `on_shutdown()`, or before-hooks.
---@param method string HTTP method.
---@param url string Full URL including scheme (e.g., `"https://api.example.com/data"`).
---@param opts? request.Opts
---@return request.Response
function request.send(method, url, opts) end

--- Send a GET request.
---@param url string
---@param opts? request.Opts
---@return request.Response
function request.get(url, opts) end

--- Send a POST request.
---@param url string
---@param opts? request.Opts
---@return request.Response
function request.post(url, opts) end

--- Send a PUT request.
---@param url string
---@param opts? request.Opts
---@return request.Response
function request.put(url, opts) end

--- Send a DELETE request.
---@param url string
---@param opts? request.Opts
---@return request.Response
function request.delete(url, opts) end

--- Send a PATCH request.
---@param url string
---@param opts? request.Opts
---@return request.Response
function request.patch(url, opts) end
