# Plan: `_global` + `_plugin` Reserved Types

## Summary

Add two new reserved datatype types to the engine: `_global` for singleton site-wide content (menus, footers, dialogs, site settings) and `_plugin` as a namespaced prefix for plugin-provided content nodes.

## Steps

1. Register `_global` and `_plugin` in `internal/db/types/reserved.go` with `IsPluginType()` helper validating `_plugin_` prefix against loaded plugins
2. SQL: Add `_global` to TopLevel query filters alongside `_root`; add delivery query for globals (no route, grouped by datatype)
3. DB wrapper: `ListGlobals()` method on DbDriver interface
4. Router: `/globals` public endpoint; include globals in content delivery responses
5. Plugin integration: plugin `OnInit` registers `_plugin_{name}` type; engine delegates to plugin render hook during tree build
6. Admin/CLI: surface globals as distinct section
