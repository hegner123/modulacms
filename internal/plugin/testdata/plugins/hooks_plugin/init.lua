-- hooks_plugin: Basic hook registration for testing Phase 3.
-- Registers before_create and after_create hooks on content_data.

plugin_info = {
    name        = "hooks_plugin",
    version     = "1.0.0",
    description = "Hook test plugin",
}

-- Register hooks at module scope (before on_init, per phase guard).
hooks.on("before_create", "content_data", function(data)
    log.info("before_create fired for " .. tostring(data._table))
end)

hooks.on("after_create", "content_data", function(data)
    log.info("after_create fired for " .. tostring(data._table))
end)

function on_init()
    log.info("hooks_plugin initialized")
end
