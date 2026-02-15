-- reload_plugin: test fixture for hot reload watcher tests.
-- Version is read from a file-scope variable so tests can verify version
-- changes after reload.
--
-- Used by watcher_test.go to verify:
--   - File change detection triggers reload
--   - Version bump is reflected after reload
--   - Debounce prevents mid-save partial reloads

plugin_info = {
    name        = "reload_plugin",
    version     = "1.0.0",
    description = "Test fixture for hot reload",
}

function on_init()
    log.info("reload_plugin initialized v" .. plugin_info.version)
end

function on_shutdown()
    log.info("reload_plugin shutting down")
end
