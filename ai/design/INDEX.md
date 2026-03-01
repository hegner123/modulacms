# ModulaCMS Admin Panel Designs

Planning directory for alternative admin panel implementations. Each design is a complete admin UI for ModulaCMS, targeting the same backend API but with different visual identity, component library, or framework.

## Shared Specs

- [FEATURES.md](FEATURES.md) — Complete feature spec (screens, interactions, permissions, UI patterns)
- [SCHEMAS.md](SCHEMAS.md) — Domain-appropriate data schemas per design variant

## Component Libraries

| Design | Directory | Description |
|--------|-----------|-------------|
| shadcn/ui | `shadcn/` | Radix-based composable components — *Developer Documentation Platform* |
| MUI | `mui/` | Material Design system — *E-Commerce Admin* |

## Custom Design Directions

| Design | Directory | Description |
|--------|-----------|-------------|
| Tailwind UI | `tailwind/` | Tailwind UI component catalog — *SaaS Dashboard (multi-tenant)* |
| Sentry | `sentry/` | Paper feel, muted tones — *Error Monitoring Platform* |
| Vercel | `vercel/` | Black/white, minimal, modern — *Deployment Platform* |

## Framework-Specific

| Design | Directory | Description |
|--------|-----------|-------------|
| Vue | `vue/` | Vue 3 + Composition API — *Blog / Magazine* |
| React | `react/` | React 18+ — *Project Management Tool* |
| Solid | `solid/` | SolidJS, fine-grained reactivity — *Real-time Analytics Dashboard* |
| Astro | `astro/` | Islands architecture — *Portfolio / Agency Website* |

## Structure

Each design directory will contain:

```
{design}/
  DESIGN.md       -- Visual direction, color palette, typography, spacing
  COMPONENTS.md   -- Component inventory and mapping to CMS features
  SCREENS.md      -- Screen-by-screen specs (layout, interactions, states)
  DECISIONS.md    -- Trade-offs, deviations from reference, rationale
```

## Status

| Design | Status |
|--------|--------|
| shadcn/ui | Not started |
| MUI | Not started |
| Tailwind UI | Not started |
| Sentry | Not started |
| Vercel | Not started |
| Vue | Not started |
| React | Not started |
| Solid | Not started |
| Astro | Not started |
