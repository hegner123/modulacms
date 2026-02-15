-- http_middleware_plugin: Middleware enrichment test plugin.
--
-- Used by http_integration_test.go to verify:
--   - Middleware can enrich the request table with custom fields
--   - Handler can read middleware-set fields

plugin_info = {
    name = "http_middleware_plugin",
    version = "1.0.0",
    description = "Middleware enrichment test plugin",
}

http.use(function(req)
    req.custom_field = "enriched_value"
    req.request_count = 1
    return nil
end)

http.handle("GET", "/check", function(req)
    return {status = 200, json = {
        custom_field = req.custom_field,
        request_count = req.request_count,
    }}
end, {public = true})

function on_init()
end
