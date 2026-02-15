-- failing_plugin: test fixture for circuit breaker tests.
-- on_init raises an error to simulate a plugin that fails during initialization.
--
-- Used by recovery_test.go and cli_commands_test.go to verify:
--   - Circuit breaker trips after consecutive failures
--   - Admin can reset the circuit breaker via enable endpoint
--   - Failed plugins are visible in the plugin list

plugin_info = {
    name        = "failing_plugin",
    version     = "1.0.0",
    description = "Test fixture for circuit breaker",
}

function on_init()
    error("intentional failure for testing")
end
