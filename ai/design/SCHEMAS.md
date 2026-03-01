# Data Schemas per Design Variant

Each admin panel design uses a domain-appropriate data schema built from ModulaCMS primitives: datatypes, fields (14 types), relations, and the content tree.

All schemas use the same underlying system: datatypes define structure, fields define properties (with validation/ui_config JSON), content_data holds instances in a tree, content_fields stores values, and content_relations links instances across datatypes.

---

## shadcn/ui — Developer Documentation Platform

A docs site with versioned guides, code examples, and API references. Clean, technical.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Guide | guide | _root | Documentation pages |
| Section | section | content | Guide subsections (tree children) |
| API Endpoint | api_endpoint | content | REST API reference entries |
| Code Example | code_example | content | Standalone code snippets |
| Changelog Entry | changelog_entry | content | Release notes |

### Fields

**Guide**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required, max 200 | |
| Slug | slug | required | Auto from title |
| Summary | textarea | max 500 | Card preview text |
| Body | richtext | required | Markdown-heavy toolbar |
| Version | text | required | e.g. "v2.1" |
| Category | select | required | Options: getting-started, concepts, api, recipes |
| Sections | relation | | target: Section, cardinality: many |
| Related Guides | relation | max_items: 5 | target: Guide, cardinality: many |
| Published | boolean | | |

**Section**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Heading | text | required, max 150 | |
| Body | richtext | required | |
| Code Blocks | relation | | target: Code Example, cardinality: many |
| Sort Order | number | | Display position |

**API Endpoint**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Method | select | required | Options: GET, POST, PUT, PATCH, DELETE |
| Path | text | required | e.g. "/api/v1/users/{id}" |
| Summary | textarea | required | |
| Request Body | json | | Schema definition |
| Response Body | json | | Schema definition |
| Auth Required | boolean | | |
| Deprecated | boolean | | |

**Code Example**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Language | select | required | Options: typescript, go, python, rust, bash |
| Code | textarea | required | Raw code content |
| Description | textarea | | |

**Changelog Entry**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Version | text | required | Semver |
| Date | date | required | |
| Summary | textarea | required | |
| Body | richtext | | Full release notes |
| Type | select | required | Options: major, minor, patch, security |

---

## MUI — E-Commerce Admin

An online store backend. Products, orders, customers, inventory. Dense data tables, lots of filters.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Product | product | _root | Store items |
| Product Variant | product_variant | content | Size/color variants (tree children of Product) |
| Category | category | content | Product categories (hierarchical) |
| Customer | customer | content | Buyer profiles |
| Order | order | content | Purchase records |
| Order Item | order_item | content | Line items (tree children of Order) |
| Review | review | content | Product reviews |

### Fields

**Product**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required, max 200 | |
| Slug | slug | required | |
| Description | richtext | | |
| SKU | text | required | Stock keeping unit |
| Price | number | required, min 0 | Cents (integer) |
| Compare At Price | number | min 0 | Strikethrough price |
| Featured Image | media | | |
| Gallery | relation | max_items: 20 | target: Media items |
| Category | relation | | target: Category, cardinality: one |
| Tags | text | | Comma-separated |
| Status | select | required | Options: active, draft, archived |
| Inventory Count | number | min 0 | |
| Weight | number | min 0 | Grams |
| Variants | relation | | target: Product Variant, cardinality: many |
| Reviews | relation | | target: Review, cardinality: many |

**Product Variant**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Label | text | required | e.g. "Large / Red" |
| SKU | text | required | |
| Price Override | number | min 0 | Null = use parent price |
| Inventory Count | number | min 0 | |
| Option 1 Name | text | | e.g. "Size" |
| Option 1 Value | text | | e.g. "Large" |
| Option 2 Name | text | | e.g. "Color" |
| Option 2 Value | text | | e.g. "Red" |

**Category**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Description | textarea | | |
| Image | media | | |
| Parent Category | relation | | target: Category, cardinality: one (self-referential tree) |

**Customer**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Email | email | required | |
| Phone | text | | |
| Address | textarea | | |
| Notes | textarea | | Internal notes |
| Total Spent | number | | Computed, read-only |
| Order Count | number | | Computed |
| Orders | relation | | target: Order, cardinality: many |

**Order**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Order Number | text | required | Auto-generated |
| Customer | relation | required | target: Customer, cardinality: one |
| Status | select | required | Options: pending, processing, shipped, delivered, cancelled, refunded |
| Subtotal | number | | Cents |
| Tax | number | | Cents |
| Shipping | number | | Cents |
| Total | number | | Cents |
| Shipping Address | textarea | | |
| Notes | textarea | | |
| Items | relation | | target: Order Item, cardinality: many |
| Placed At | datetime | | |

**Order Item**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Product | relation | required | target: Product, cardinality: one |
| Variant | relation | | target: Product Variant, cardinality: one |
| Quantity | number | required, min 1 | |
| Unit Price | number | required | Cents at time of purchase |
| Total | number | | quantity * unit_price |

**Review**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Author Name | text | required | |
| Rating | number | required, min 1, max 5 | |
| Title | text | | |
| Body | textarea | | |
| Verified Purchase | boolean | | |
| Date | datetime | | |

---

## Tailwind UI — SaaS Dashboard

A multi-tenant SaaS platform. Projects, teams, billing, usage analytics. Card-heavy layout with charts.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Organization | organization | _root | Tenant accounts |
| Project | project | content | Workspaces within an org |
| Team Member | team_member | content | Org users with roles |
| Subscription | subscription | content | Billing plan |
| Invoice | invoice | content | Payment records |
| Usage Record | usage_record | content | Metered usage events |
| Announcement | announcement | content | Platform-wide notices |

### Fields

**Organization**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required, max 100 | |
| Slug | slug | required | URL namespace |
| Logo | media | | |
| Plan | select | required | Options: free, starter, pro, enterprise |
| Owner | relation | required | target: Team Member, cardinality: one |
| Members | relation | | target: Team Member, cardinality: many |
| Projects | relation | | target: Project, cardinality: many |
| Created At | datetime | | |

**Project**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Description | textarea | max 500 | |
| Status | select | required | Options: active, paused, archived |
| Environment | select | required | Options: development, staging, production |
| API Key | text | | Auto-generated, masked display |
| Region | select | | Options: us-east, us-west, eu-west, ap-southeast |
| Created At | datetime | | |
| Last Active | datetime | | |

**Team Member**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Email | email | required | |
| Role | select | required | Options: owner, admin, member, viewer |
| Avatar | media | | |
| Joined At | datetime | | |
| Last Login | datetime | | |

**Subscription**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Plan | select | required | Options: free, starter, pro, enterprise |
| Status | select | required | Options: active, past_due, cancelled, trialing |
| Current Period Start | datetime | | |
| Current Period End | datetime | | |
| Monthly Price | number | | Cents |
| Payment Method | select | | Options: card, invoice, wire |
| Auto Renew | boolean | | |

**Invoice**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Number | text | required | |
| Amount | number | required | Cents |
| Status | select | required | Options: draft, open, paid, void, uncollectible |
| Period Start | date | | |
| Period End | date | | |
| Due Date | date | | |
| Paid At | datetime | | |
| PDF | media | | |

**Usage Record**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Metric | select | required | Options: api_calls, storage_gb, bandwidth_gb, compute_hours |
| Value | number | required | |
| Timestamp | datetime | required | |
| Project | relation | | target: Project, cardinality: one |

**Announcement**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Body | richtext | | |
| Type | select | required | Options: info, warning, maintenance, feature |
| Active | boolean | | |
| Start Date | datetime | | |
| End Date | datetime | | |

---

## Sentry — Error Monitoring Platform

Issue tracking, error events, alert rules, project configuration. Paper feel, muted tones, information-dense.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Project | project | _root | Monitored applications |
| Issue | issue | content | Grouped error events |
| Event | event | content | Individual error occurrences (tree children of Issue) |
| Alert Rule | alert_rule | content | Notification triggers |
| Release | release | content | Deployed versions |
| Team | team | content | Groups of members |
| Environment | environment | content | Deployment targets |

### Fields

**Project**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Platform | select | required | Options: javascript, python, go, rust, java, ruby, php, swift, kotlin, csharp |
| DSN | text | | Auto-generated, read-only |
| Default Environment | relation | | target: Environment, cardinality: one |
| Teams | relation | | target: Team, cardinality: many |
| Alert Rules | relation | | target: Alert Rule, cardinality: many |

**Issue**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | Error message summary |
| Culprit | text | | Function/file that caused it |
| Level | select | required | Options: fatal, error, warning, info, debug |
| Status | select | required | Options: unresolved, resolved, ignored, muted |
| Priority | select | | Options: critical, high, medium, low |
| First Seen | datetime | | |
| Last Seen | datetime | | |
| Event Count | number | | |
| Affected Users | number | | |
| Assigned To | text | | User/team name |
| Tags | json | | Key-value pairs |
| Project | relation | required | target: Project, cardinality: one |

**Event**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Event ID | text | required | Unique event identifier |
| Message | text | | |
| Stacktrace | json | | Frames array |
| Breadcrumbs | json | | Ordered list of user actions |
| Tags | json | | Key-value pairs |
| Context | json | | Device, OS, browser, runtime |
| User | json | | id, email, ip_address |
| Environment | text | | |
| Release | text | | Version string |
| Timestamp | datetime | required | |
| Level | select | required | Options: fatal, error, warning, info, debug |

**Alert Rule**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Conditions | json | required | When to trigger |
| Actions | json | required | What to do (email, Slack, webhook) |
| Frequency | number | | Minutes between alerts |
| Enabled | boolean | | |
| Project | relation | required | target: Project, cardinality: one |
| Environment | relation | | target: Environment, cardinality: one |

**Release**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Version | text | required | Semver or commit SHA |
| Project | relation | required | target: Project, cardinality: one |
| Environment | relation | | target: Environment, cardinality: one |
| Deploy Date | datetime | | |
| Commit Count | number | | |
| New Issues | number | | Issues first seen in this release |
| Resolved Issues | number | | Issues resolved by this release |
| Status | select | | Options: active, archived |

**Team**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Members | json | | Array of user references |

**Environment**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | e.g. "production", "staging" |
| URL | url | | Base URL for this environment |
| Is Production | boolean | | |

---

## Vercel — Deployment Platform

Projects, deployments, domains, environment variables, logs. Monochrome, minimal, modern.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Project | project | _root | Deployed applications |
| Deployment | deployment | content | Build + deploy events |
| Domain | domain | content | Custom domains |
| Environment Variable | env_var | content | Config key-value pairs |
| Log Entry | log_entry | content | Runtime/build logs |
| Integration | integration | content | Connected services |
| Edge Config | edge_config | content | Edge runtime configuration |

### Fields

**Project**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Framework | select | | Options: nextjs, remix, astro, sveltekit, nuxt, vite, other |
| Repository | url | | Git repo URL |
| Production Branch | text | | Default: main |
| Root Directory | text | | Monorepo path |
| Build Command | text | | e.g. "npm run build" |
| Output Directory | text | | e.g. ".next" |
| Node Version | select | | Options: 18.x, 20.x, 22.x |
| Domains | relation | | target: Domain, cardinality: many |
| Env Vars | relation | | target: Environment Variable, cardinality: many |
| Created At | datetime | | |
| Last Deployed | datetime | | |

**Deployment**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| URL | url | | Unique deployment URL |
| Status | select | required | Options: queued, building, ready, error, cancelled |
| Environment | select | required | Options: production, preview, development |
| Source | select | | Options: git, cli, api, redeploy |
| Git Branch | text | | |
| Git Commit SHA | text | | |
| Git Commit Message | text | | |
| Build Duration | number | | Seconds |
| Functions Count | number | | Serverless functions deployed |
| Size | number | | Bytes |
| Created At | datetime | | |
| Ready At | datetime | | |

**Domain**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | e.g. "example.com" |
| Type | select | required | Options: production, preview, redirect |
| Verified | boolean | | DNS verification status |
| SSL Status | select | | Options: pending, active, error |
| Redirect Target | url | | For redirect domains |
| Git Branch | text | | Branch for preview domains |
| Added At | datetime | | |

**Environment Variable**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Key | text | required | Variable name |
| Value | text | required | Variable value (masked in UI) |
| Target | select | required | Options: production, preview, development |
| Sensitive | boolean | | Mask value in logs |
| Created At | datetime | | |

**Log Entry**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Message | text | required | |
| Level | select | required | Options: info, warn, error |
| Source | select | | Options: build, runtime, edge, static |
| Timestamp | datetime | required | |
| Request Path | text | | For runtime logs |
| Status Code | number | | HTTP status |
| Duration | number | | Milliseconds |
| Region | text | | Edge region |

**Integration**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Provider | select | required | Options: github, gitlab, bitbucket, slack, datadog, sentry |
| Status | select | required | Options: active, inactive, error |
| Config | json | | Provider-specific settings |
| Installed At | datetime | | |

**Edge Config**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Items | json | required | Key-value store content |
| Size | number | | Bytes |
| Read Regions | json | | Array of edge regions |
| Last Modified | datetime | | |

---

## Vue — Blog / Magazine

Classic publishing platform. Articles, authors, categories, tags, series. Vue-green accent.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Article | article | _root | Published posts |
| Author | author | content | Writers |
| Category | category | content | Topic categories (hierarchical) |
| Tag | tag | content | Flat taxonomy labels |
| Series | series | content | Multi-part article collections |
| Page | page | content | Static pages (about, contact) |

### Fields

**Article**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required, max 200 | |
| Slug | slug | required | |
| Excerpt | textarea | max 300 | |
| Body | richtext | required | Full toolbar |
| Featured Image | media | | |
| Author | relation | required | target: Author, cardinality: one |
| Category | relation | required | target: Category, cardinality: one |
| Tags | relation | max_items: 10 | target: Tag, cardinality: many |
| Series | relation | | target: Series, cardinality: one |
| Series Position | number | min 1 | Order within series |
| Published Date | datetime | | |
| Reading Time | number | | Minutes (computed or manual) |
| SEO Title | text | max 60 | |
| SEO Description | textarea | max 160 | |
| Featured | boolean | | Homepage feature |

**Author**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Bio | textarea | max 500 | |
| Avatar | media | | |
| Email | email | | |
| Website | url | | |
| Social Twitter | text | | Handle |
| Social GitHub | text | | Handle |

**Category**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Description | textarea | | |
| Color | text | | Hex color for category badge |
| Parent | relation | | target: Category, cardinality: one (self-referential) |

**Tag**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |

**Series**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Slug | slug | required | |
| Description | richtext | | |
| Cover Image | media | | |
| Status | select | required | Options: ongoing, complete |

**Page**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Slug | slug | required | |
| Body | richtext | required | |
| Template | select | | Options: default, full-width, sidebar |

---

## React — Project Management Tool

Kanban boards, tasks, sprints, team workload. Component-rich, state-heavy.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Workspace | workspace | _root | Top-level container |
| Board | board | content | Kanban board |
| Column | column | content | Board lane (tree children of Board) |
| Task | task | content | Work items (tree children of Column) |
| Sprint | sprint | content | Time-boxed iterations |
| Label | label | content | Color-coded tags |
| Comment | comment | content | Task discussion (tree children of Task) |

### Fields

**Workspace**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Description | textarea | | |
| Icon | text | | Emoji or icon name |
| Boards | relation | | target: Board, cardinality: many |
| Members | json | | Array of user IDs + roles |

**Board**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Description | textarea | | |
| Type | select | required | Options: kanban, scrum, backlog |
| Columns | relation | | target: Column, cardinality: many |
| Default Assignee | text | | |
| Sprint | relation | | target: Sprint, cardinality: one |

**Column**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | e.g. "To Do", "In Progress", "Done" |
| Color | text | | Hex color |
| WIP Limit | number | min 0 | Max tasks allowed |
| Is Done Column | boolean | | Marks completion |

**Task**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required, max 300 | |
| Description | richtext | | |
| Status | select | required | Options: open, in_progress, review, done, blocked |
| Priority | select | required | Options: urgent, high, medium, low, none |
| Assignee | text | | User reference |
| Labels | relation | max_items: 5 | target: Label, cardinality: many |
| Due Date | date | | |
| Estimate | number | min 0 | Story points |
| Subtasks | relation | | target: Task, cardinality: many (self-referential) |
| Attachments | relation | max_items: 10 | target: Media |
| Sprint | relation | | target: Sprint, cardinality: one |
| Created At | datetime | | |
| Completed At | datetime | | |

**Sprint**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | e.g. "Sprint 14" |
| Goal | textarea | | |
| Start Date | date | required | |
| End Date | date | required | |
| Status | select | required | Options: planning, active, complete |
| Velocity | number | | Completed points |

**Label**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | e.g. "Bug", "Feature", "Tech Debt" |
| Color | text | required | Hex color |

**Comment**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Body | richtext | required | |
| Author | text | required | |
| Created At | datetime | | |
| Edited | boolean | | |

---

## Solid — Real-time Analytics Dashboard

Metrics, dashboards, data sources, alerts. Reactive, performance-focused.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Dashboard | dashboard | _root | Metrics display |
| Widget | widget | content | Chart/stat/table (tree children of Dashboard) |
| Data Source | data_source | content | External data connections |
| Metric | metric | content | Named measurement |
| Alert | alert | content | Threshold notifications |
| Report | report | content | Scheduled exports |

### Fields

**Dashboard**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Slug | slug | required | |
| Description | textarea | | |
| Layout | json | | Grid positions for widgets |
| Refresh Interval | number | min 5 | Seconds |
| Time Range | select | | Options: 1h, 6h, 24h, 7d, 30d, custom |
| Is Default | boolean | | Show on login |
| Widgets | relation | | target: Widget, cardinality: many |
| Shared With | json | | User/team IDs |

**Widget**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Type | select | required | Options: line_chart, bar_chart, pie_chart, stat_card, table, heatmap, gauge |
| Metric | relation | required | target: Metric, cardinality: one |
| Data Source | relation | required | target: Data Source, cardinality: one |
| Query | json | | Aggregation/filter config |
| Thresholds | json | | Color bands for values |
| Position | json | | {x, y, w, h} grid coords |
| Refresh Override | number | | Per-widget interval |

**Data Source**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Type | select | required | Options: postgresql, mysql, clickhouse, prometheus, elasticsearch, http_api |
| Connection | json | required | Host, port, credentials (encrypted) |
| Status | select | | Options: connected, error, disabled |
| Last Checked | datetime | | |
| Query Timeout | number | min 1 | Seconds |

**Metric**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Unit | text | | e.g. "ms", "%", "req/s", "GB" |
| Data Source | relation | required | target: Data Source, cardinality: one |
| Query | text | required | SQL or PromQL |
| Aggregation | select | | Options: sum, avg, min, max, count, p50, p95, p99 |
| Description | textarea | | |

**Alert**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Metric | relation | required | target: Metric, cardinality: one |
| Condition | select | required | Options: above, below, equals, change_pct |
| Threshold | number | required | |
| Duration | number | required | Minutes sustained before firing |
| Severity | select | required | Options: critical, warning, info |
| Notify | json | | Channels: email, slack, webhook |
| Enabled | boolean | | |
| Last Fired | datetime | | |

**Report**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Dashboard | relation | required | target: Dashboard, cardinality: one |
| Schedule | select | required | Options: daily, weekly, monthly |
| Format | select | required | Options: pdf, csv, png |
| Recipients | json | | Email addresses |
| Enabled | boolean | | |
| Last Sent | datetime | | |

---

## Astro — Portfolio / Agency Website

Client projects, case studies, team bios, services. Islands architecture suits mixed static/dynamic content.

### Datatypes

| Datatype | Name | Type | Purpose |
|----------|------|------|---------|
| Case Study | case_study | _root | Client project showcases |
| Service | service | content | Offered capabilities |
| Team Member | team_member | content | Staff profiles |
| Testimonial | testimonial | content | Client quotes |
| Page | page | content | Static pages |
| FAQ | faq | content | Question/answer pairs |

### Fields

**Case Study**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Slug | slug | required | |
| Client Name | text | required | |
| Client Logo | media | | |
| Excerpt | textarea | max 200 | Card summary |
| Body | richtext | required | |
| Hero Image | media | required | |
| Gallery | relation | max_items: 20 | Media references |
| Services Used | relation | | target: Service, cardinality: many |
| Team | relation | | target: Team Member, cardinality: many |
| Testimonial | relation | | target: Testimonial, cardinality: one |
| Industry | select | | Options: tech, finance, healthcare, education, retail, media, nonprofit |
| Results | json | | Array of {metric, value, description} |
| Published Date | date | | |
| Featured | boolean | | |
| SEO Title | text | max 60 | |
| SEO Description | textarea | max 160 | |

**Service**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Tagline | text | max 100 | |
| Description | richtext | | |
| Icon | text | | Icon name or SVG |
| Featured Image | media | | |
| Pricing Model | select | | Options: project, hourly, retainer, custom |
| Starting Price | number | | |

**Team Member**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Name | text | required | |
| Slug | slug | required | |
| Role | text | required | Job title |
| Bio | richtext | | |
| Photo | media | required | |
| Email | email | | |
| LinkedIn | url | | |
| Sort Order | number | | Display position |

**Testimonial**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Quote | textarea | required, max 500 | |
| Author Name | text | required | |
| Author Title | text | | Job title |
| Author Company | text | | |
| Author Photo | media | | |
| Rating | number | min 1, max 5 | |
| Featured | boolean | | |

**Page**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Title | text | required | |
| Slug | slug | required | |
| Body | richtext | required | |
| Template | select | | Options: default, landing, contact, about |
| SEO Title | text | max 60 | |
| SEO Description | textarea | max 160 | |

**FAQ**
| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| Question | text | required | |
| Answer | richtext | required | |
| Category | select | | Options: general, services, pricing, process |
| Sort Order | number | | |
