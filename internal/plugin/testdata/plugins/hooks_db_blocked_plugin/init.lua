-- hooks_db_blocked_plugin: Tests M1 - db.* calls blocked in before-hooks.
-- The before_create hook tries to call db.query which should be rejected.

plugin_info = {
    name        = "hooks_db_blocked_plugin",
    version     = "1.0.0",
    description = "DB blocked in before-hook test",
}

hooks.on("before_create", "content_data", function(data)
    -- This should raise an error because db.* is blocked in before-hooks.
    db.query("some_table")
end)

function on_init()
    log.info("hooks_db_blocked_plugin initialized")
end
