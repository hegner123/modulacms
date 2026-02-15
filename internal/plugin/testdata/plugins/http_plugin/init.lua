-- http_plugin: Task tracker test plugin with authenticated routes and middleware.
--
-- Used by http_integration_test.go to verify:
--   - Authenticated GET/POST routes
--   - Middleware enrichment (processed_by field)
--   - Multi-plugin coexistence
--   - CRUD response shape

plugin_info = {
    name = "http_plugin",
    version = "1.0.0",
    description = "Task tracker test plugin",
}

-- In-memory task storage (per-VM, resets on snapshot restore).
local tasks = {}
local next_id = 1

http.use(function(req)
    req.processed_by = "http_plugin_middleware"
    return nil
end)

http.handle("GET", "/tasks", function(req)
    return {status = 200, json = {tasks = tasks, processed_by = req.processed_by}}
end)

http.handle("POST", "/tasks", function(req)
    local task = {id = next_id, title = req.json and req.json.title or "untitled"}
    next_id = next_id + 1
    tasks[#tasks + 1] = task
    return {status = 201, json = task}
end)

function on_init()
end
