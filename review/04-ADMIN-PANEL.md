# Admin Panel Review

## Overview

97 TypeScript/TSX files, ~12,900 lines. React 19 + TypeScript 5.9. Embedded in the Go binary via `//go:embed dist/*`. Served at `/admin/` with SPA routing fallback.

## Tech Stack Assessment

| Choice | Verdict | Why |
|--------|---------|-----|
| React 19 | Good | Latest stable, no experimental features misused |
| TypeScript 5.9 strict | Good | Real type safety, not just decoration |
| TanStack Router | Good | File-based routing with type-safe params |
| TanStack React Query | Good | Server state management without Redux boilerplate |
| React Hook Form + Zod | Good | Validated forms without ceremony |
| shadcn/ui + Radix | Good | Accessible primitives, customizable, not a heavy framework |
| Tailwind CSS v4 | Good | Utility-first, consistent design tokens |
| Vite 7 | Good | Fast dev server, efficient production builds |
| dnd-kit | Good | Modern drag-drop, works with React 18+ |

No questionable dependencies. No state management library beyond React Query. No CSS-in-JS overhead. This is a well-chosen stack.

## What Solves a Real Problem

**The block editor** is the centerpiece. Hierarchical content editing with drag-drop reordering, nested blocks, inline field editing, and dirty state tracking. `useBlockEditorState()` manages local edits as a `Record<contentDataId, Record<fieldId, value>>`, enabling per-block save without full-page submission. This is the kind of editor that makes a CMS usable.

**SDK proxy pattern** for auth switching is clever. Two SDK clients (session-based, API key-based) behind a Proxy object. All code imports the proxy, and the active client switches at runtime during login/logout. No component rewrites needed when auth mode changes.

**Embedded distribution** means the admin panel ships with the Go binary. No separate deployment, no CORS issues for the admin interface, no CDN required. Users get a working admin panel the moment they run the binary.

## What Is Good

**Component architecture** follows best practices:
- Small, single-responsibility components (BlockCard handles display, not list management)
- Custom hooks for reusable logic (useBlockEditorState, useMediaUpload)
- Proper memoization (useCallback for handlers, useMemo for derived state)
- Early returns over nested conditionals

**Data management** via TanStack Query:
- One query file per resource (auth, content, datatypes, fields, media, etc.)
- Hierarchical query keys in a central file
- Mutations invalidate parent keys for automatic refresh
- 5-minute stale time, 1 retry default

**Media library** is fully functional: folder navigation with breadcrumb, file upload with progress, search within directories, grid view with thumbnails. Handles S3 path conventions correctly.

**Authentication flow** supports both session cookies and API key modes, with graceful fallback and re-authentication.

## What Is Bad

### Zero Tests
No test framework installed. No unit tests, no integration tests, no component tests. For a production admin panel with complex state management (block editor, auth switching, media uploads), this is a significant gap. The block editor alone has enough state logic to warrant thorough testing.

### Incomplete Pages
Several routes exist in the sidebar but may be partially implemented:
- SSH keys management
- Plugin detail views
- Import form specifics
- Settings page depth

These should either be completed or removed from navigation to avoid dead ends.

### No Rich Text Editor Implementation
A richtext field type is referenced but the actual editor component implementation is unclear. This is a critical feature for a CMS - content editors expect WYSIWYG text editing.

## What Is Extra

The relationship between the admin panel and the TUI creates some feature overlap. Both can manage content, datatypes, fields, routes, users, media, and plugins. The admin panel does it with a better UX (drag-drop, visual tree, inline editing). The TUI does it over SSH. Having both is defensible as they serve different access patterns, but maintaining feature parity across both interfaces is a significant ongoing cost.

## Code Quality

**Strengths:**
- TypeScript strict mode throughout
- Branded IDs from @modulacms/types used in all API calls
- Zod validation on forms
- Error boundaries for crash recovery
- Accessible components (semantic HTML, ARIA attributes, keyboard navigation)
- Dark mode with OKLCH color variables

**Minor issues:**
- ESLint config is basic (type-aware rules not enabled)
- No component documentation or Storybook
- Some query files could use better error state handling in the UI

## Recommendations

1. **Add Vitest + React Testing Library** and test the block editor state management
2. **Complete or remove stub pages** from navigation
3. **Implement rich text editor** with a library like TipTap or Plate
4. **Enable type-aware ESLint rules** for deeper TypeScript checking
5. **Add loading skeletons** instead of just spinners for better perceived performance
