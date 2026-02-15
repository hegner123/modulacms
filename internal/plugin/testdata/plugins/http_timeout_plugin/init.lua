-- http_timeout_plugin: Handler with infinite loop to test timeout enforcement.
--
-- Used by http_integration_test.go to verify:
--   - Context deadline enforcement during handler execution
--   - 504 HANDLER_TIMEOUT response

plugin_info = {
    name = "http_timeout_plugin",
    version = "1.0.0",
    description = "Timeout test plugin",
}

http.handle("GET", "/hang", function(req)
    while true do end
end, {public = true})

function on_init()
end
