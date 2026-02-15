-- http_public_plugin: Public routes test plugin (no CMS session required).
--
-- Used by http_integration_test.go to verify:
--   - Public routes work without authentication
--   - Multi-plugin coexistence with http_plugin

plugin_info = {
    name = "http_public_plugin",
    version = "1.0.0",
    description = "Public routes test plugin",
}

http.handle("POST", "/webhook", function(req)
    return {status = 200, json = {received = true, body = req.json}}
end, {public = true})

http.handle("GET", "/status", function(req)
    return {status = 200, json = {status = "ok"}}
end, {public = true})

function on_init()
end
