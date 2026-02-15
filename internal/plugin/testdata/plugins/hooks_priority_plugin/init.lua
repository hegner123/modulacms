-- hooks_priority_plugin: Tests priority ordering (M6).
-- Registers multiple hooks with different priorities to verify ordering.

plugin_info = {
    name        = "hooks_priority_plugin",
    version     = "1.0.0",
    description = "Priority ordering test plugin",
}

-- Priority 50 should run first.
hooks.on("before_update", "content_data", function(data)
    log.info("priority_50 hook fired")
end, { priority = 50 })

-- Priority 200 should run after priority 50.
hooks.on("before_update", "content_data", function(data)
    log.info("priority_200 hook fired")
end, { priority = 200 })

-- Default priority (100) should run between 50 and 200.
hooks.on("before_update", "content_data", function(data)
    log.info("priority_100 hook fired")
end)

function on_init()
    log.info("hooks_priority_plugin initialized")
end
