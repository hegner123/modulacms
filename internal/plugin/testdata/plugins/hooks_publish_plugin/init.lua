-- hooks_publish_plugin: Tests M12 publish/archive status transition hooks.
-- Registers before_publish, after_publish, before_archive, and after_archive
-- hooks on content_data to verify that DetectStatusTransition fires them.

plugin_info = {
    name        = "hooks_publish_plugin",
    version     = "1.0.0",
    description = "Publish/archive detection test plugin",
}

hooks.on("before_publish", "content_data", function(data)
    log.info("before_publish fired for " .. tostring(data._table))
end)

hooks.on("after_publish", "content_data", function(data)
    log.info("after_publish fired for " .. tostring(data._table))
end)

hooks.on("before_archive", "content_data", function(data)
    log.info("before_archive fired for " .. tostring(data._table))
end)

hooks.on("after_archive", "content_data", function(data)
    log.info("after_archive fired for " .. tostring(data._table))
end)

function on_init()
    log.info("hooks_publish_plugin initialized")
end
