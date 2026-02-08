// Package plugin implements the ModulaCMS Lua plugin system.
//
// The plugin system extends ModulaCMS with runtime Lua plugins that can define
// database tables, execute queries, and integrate with the CMS lifecycle.
//
// Architecture:
//
//	Manager (discovery, loading, lifecycle, shutdown)
//	  -> VMPool (per-plugin, channel-based pool of sandboxed lua.LState)
//	       -> ApplySandbox (safe stdlib subset, stripped globals)
//	       -> RegisterPluginRequire (sandboxed module loader)
//	       -> RegisterDBAPI (db.* Lua module via query builder)
//	       -> RegisterLogAPI (log.* Lua module)
//	       -> FreezeModule (read-only proxy for db/log)
//
// Entry point: NewManager() creates the manager; LoadAll() discovers and loads
// plugins; Shutdown() gracefully stops all plugins.
//
// All plugin tables are prefixed with plugin_<plugin_name>_ and validated via
// the query builder's identifier validation. Plugins cannot access core CMS
// tables or other plugins' tables.
package plugin
