-- http_blocked_headers: Handler that tries to set blocked response headers.
--
-- Used by http_integration_test.go to verify:
--   - Blocked headers (set-cookie, access-control-allow-origin) are filtered out
--   - Allowed headers (x-custom, x-plugin-id) pass through

plugin_info = {
    name = "http_blocked_headers",
    version = "1.0.0",
    description = "Blocked headers test plugin",
}

http.handle("GET", "/headers", function(req)
    return {
        status = 200,
        headers = {
            ["set-cookie"] = "evil=1",
            ["access-control-allow-origin"] = "*",
            ["x-custom"] = "allowed",
            ["x-plugin-id"] = "test",
        },
        json = {ok = true},
    }
end, {public = true})

function on_init()
end
