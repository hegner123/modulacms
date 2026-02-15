-- invalid_bad_name: test fixture for manager discovery tests.
-- plugin_info.name contains spaces and uppercase, violating the
-- [a-z0-9_] / max 32 chars constraint. Manager should reject.

plugin_info = {
    name        = "Has Spaces And CAPS",
    version     = "1.0.0",
    description = "Invalid plugin name",
}
