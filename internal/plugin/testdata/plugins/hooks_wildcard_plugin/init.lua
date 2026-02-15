-- hooks_wildcard_plugin: Tests wildcard hook registration.
-- Registers a wildcard before_create hook that fires for all tables.

plugin_info = {
    name        = "hooks_wildcard_plugin",
    version     = "1.0.0",
    description = "Wildcard hook test plugin",
}

hooks.on("before_create", "*", function(data)
    log.info("wildcard before_create fired for " .. tostring(data._table))
end)

hooks.on("after_update", "*", function(data)
    log.info("wildcard after_update fired for " .. tostring(data._table))
end)

function on_init()
    log.info("hooks_wildcard_plugin initialized")
end
