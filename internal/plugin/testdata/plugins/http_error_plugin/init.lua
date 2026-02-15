-- http_error_plugin: Handler that raises error() to test error handling.
--
-- Used by http_integration_test.go to verify:
--   - Lua error() in handler produces 500 HANDLER_ERROR
--   - Generic error message sent to client (not the Lua error string)

plugin_info = {
    name = "http_error_plugin",
    version = "1.0.0",
    description = "Error handler test plugin",
}

http.handle("GET", "/fail", function(req)
    error("deliberate test error")
end, {public = true})

function on_init()
end
