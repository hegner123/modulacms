-- harness_plugin: test fixture for plugin/testing harness tests.
--
-- Registers an HTTP route and a hook so the harness tests can exercise
-- test.request() and test.fire_hook().

plugin_info = {
    name        = "harness_plugin",
    version     = "1.0.0",
    description = "Fixture plugin for harness Go tests",
}

http.handle("GET", "/ping", function(req)
    return {status = 200, json = {pong = true}}
end)

http.handle("POST", "/echo", function(req)
    return {status = 200, json = req.json}
end)

hooks.on("before_create", "content_data", function(data)
    log.info("harness_plugin before_create fired")
end)

hooks.on("after_create", "content_data", function(data)
    log.info("harness_plugin after_create fired")
end)

function on_init()
    db.define_table("harness_items", {
        columns = {
            {name = "label", type = "text", not_null = true},
            {name = "count", type = "integer", not_null = true, default = 0},
        },
    })
    log.info("harness_plugin initialized")
end
