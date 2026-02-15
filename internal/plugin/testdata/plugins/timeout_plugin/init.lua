-- timeout_plugin: test fixture for sandbox timeout enforcement.
-- on_init contains an infinite loop. The VM context deadline should
-- cancel execution before this completes.

plugin_info = {
    name        = "timeout_plugin",
    version     = "1.0.0",
    description = "Tests execution timeout enforcement",
}

function on_init()
    local i = 0
    while true do
        i = i + 1
    end
end
