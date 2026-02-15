-- http_params_plugin: Path parameters test plugin.
--
-- Used by http_integration_test.go to verify:
--   - Path parameter extraction via req.params.id
--   - Multi-plugin coexistence

plugin_info = {
    name = "http_params_plugin",
    version = "1.0.0",
    description = "Path parameters test plugin",
}

http.handle("GET", "/items/{id}", function(req)
    return {status = 200, json = {id = req.params.id, method = req.method}}
end)

function on_init()
end
