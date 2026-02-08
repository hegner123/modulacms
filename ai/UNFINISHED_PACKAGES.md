  Completely Missing / Stub Only
  ┌─────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────┐
  │     Package     │                                          What's Missing                                          │
  ├─────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/model  │ LoadPageContent and SavePageContent are stubs returning "not implemented" (cms_model.go:53-64)   │
  ├─────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/deploy │ IssueMakeBackup() and DownloadBackup() are empty function bodies (deploy.go)                     │
  ├─────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/plugin │ Entire Lua plugin system is skeletal (~30 lines). No execution, loading, or lifecycle management │
  ├─────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/file   │ save.go is empty (package declaration only)                                                      │
  ├─────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/db     │ joins.go is empty (package declaration only)                                                     │
  └─────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────┘
  Partially Implemented
  ┌────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
  │      Package       │                                                     What's Missing                                                      │
  ├────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/transform │ Inbound parsing stubs for Sanity, Strapi, WordPress (parse_stubs.go). Outbound works fine                               │
  ├────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/router    │ Restore endpoint has no actual SQL execution (restore.go:63); media uploads hardcode author to "1" (mediaUpload.go:105) │
  ├────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/config    │ UI config loading/defaults not implemented (ui.go:130-131)                                                              │
  ├────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/cli       │ Missing JOIN queries for content data/datatypes, route selection UI (commands.go)                                       │
  ├────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/install   │ Extended wizard steps incomplete per TODO.md                                                                            │
  └────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
  Working But Needs Hardening
  ┌─────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
  │       Package       │                                                  What's Missing                                                   │
  ├─────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/middleware │ Session-based auth refactor, CSRF protection, rate limiting on SSH (TODO.md)                                      │
  ├─────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/auth       │ OAuth state validation commented out, no token refresh, hardcoded userinfo endpoint (ai/OAUTH_PRODUCTION_PLAN.md) │
  ├─────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/media      │ 23 documented bugs in MEDIA_PACKAGE_FIX_PLAN.md                                                                   │
  ├─────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ internal/utility    │ Sentry integration is a placeholder (observability.go:182)                                                        │
  └─────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
  The most critical gaps are model (can't load/save content via high-level API), plugin (entire system is skeletal), and deploy (no functional sync).
