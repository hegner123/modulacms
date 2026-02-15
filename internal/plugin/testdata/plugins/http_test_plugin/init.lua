-- http_test_plugin: Test plugin for HTTP bridge tests
--
-- Registers both authenticated and public routes plus middleware
-- for testing the full HTTPBridge dispatch pipeline.

plugin_info = {
    name = "http_test",
    version = "1.0.0",
    description = "Test plugin for HTTP bridge tests",
}

-- Authenticated route (default): requires CMS session
http.handle("GET", "/hello", function(req)
    return {status = 200, json = {message = "hello"}}
end)

-- Authenticated POST route that echoes JSON body
http.handle("POST", "/echo", function(req)
    return {status = 200, json = req.json}
end)

-- Public route: no CMS session required
http.handle("POST", "/webhook", function(req)
    return {status = 200, json = {received = true}}
end, {public = true})

-- Route with path parameters
http.handle("GET", "/items/{id}", function(req)
    return {status = 200, json = {id = req.params.id}}
end)

-- Route that returns custom headers (some blocked, some allowed)
http.handle("GET", "/headers", function(req)
    return {
        status = 200,
        headers = {
            ["set-cookie"] = "evil=1",
            ["x-custom"] = "allowed",
        },
        json = {ok = true},
    }
end)

function on_init()
    -- no-op for HTTP tests
end
