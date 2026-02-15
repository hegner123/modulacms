-- hooks_abort_plugin: Tests before-hook abort behavior.
-- The before_create hook always raises an error to abort the mutation.

plugin_info = {
    name        = "hooks_abort_plugin",
    version     = "1.0.0",
    description = "Hook abort test plugin",
}

hooks.on("before_create", "content_data", function(data)
    error("validation failed: title is required")
end)

function on_init()
    log.info("hooks_abort_plugin initialized")
end
