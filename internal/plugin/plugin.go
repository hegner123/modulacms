// Package plugin implements the Modula Lua plugin system.
//
// The plugin system extends Modula with runtime Lua plugins that can define
// database tables, execute queries, and integrate with the CMS lifecycle.
//
// Architecture:
//
//	Manager (discovery, loading, lifecycle, shutdown)
//	  -> VMPool (per-plugin, channel-based pool of sandboxed lua.LState)
//	       -> ApplySandbox (safe stdlib subset, stripped globals)
//	       -> RegisterPluginRequire (sandboxed module loader)
//	       -> RegisterDBAPI (db.* Lua module via query builder)
//	       -> RegisterCoreAPI (core.* Lua module for gated core table access)
//	       -> RegisterLogAPI (log.* Lua module)
//	       -> FreezeModule (read-only proxy for db/log/core)
//
// Entry point: NewManager() creates the manager; LoadAll() discovers and loads
// plugins; Shutdown() gracefully stops all plugins.
//
// Plugin-owned tables are prefixed with plugin_<plugin_name>_ and validated via
// the query builder's identifier validation. Plugins cannot access other
// plugins' tables.
//
// Core CMS tables (content_data, users, media, etc.) are accessible via the
// core.* Lua module, gated by three layers: a hardcoded table whitelist,
// per-table read/write policies, and the plugin's approved_access stored in
// the plugins DB table. See CoreTableAPI for details.
package plugin
