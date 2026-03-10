# ModulaCMS Product Philosophy

## Core Values

Every feature in ModulaCMS exists because an agency developer hit a specific pain point in production. Not from market research. Not from competitor analysis. From scars. Everything traces back to three principles.

### Flexibility

You define the schema. Not us. Not a plugin author. Not whatever opinionated framework decided that every piece of content needs a "slug," a "featured image," and exactly one taxonomy called "categories."

The core of ModulaCMS is five database tables: **routes**, **datatypes**, **fields**, **content_data**, and **content_fields**. These five tables can model anything. A real estate listing system. An event calendar. A recipe database. A product catalog. A healthcare document portal. A restaurant menu. Whatever the cleaning company is doing this week. The same schema handles all of them.

Datatypes and fields are configured at runtime — not in code, not in migration files, not at deploy time. When a client asks for a new content type six months after launch ("this definitely wasn't in the original scope"), the editor creates it through the TUI, admin panel, or API. No developer needed. No deploy needed. No downtime. No risk.

**Adding a field never breaks existing content.** Existing content simply doesn't have a value for the new field yet. No migration, no null backfill, no breaking changes. The API returns what exists. The client renders what it gets.

**Removing a field never explodes anything.** Clean separation between schema and data means the content continues to exist even as the shape around it changes.

**Schema changes are just data.** Which means syncing schema changes between environments is just syncing data. No white-knuckle `wp db export | wp db import` pipe dream where you hold your breath and hope the serialized PHP didn't get corrupted by a find-and-replace on the domain name. You export. You import. The schema version is hash-validated. The payload is integrity-checked. Dry-run tells you exactly what will change before anything changes.

This is the answer to every agency developer who has ever built the same custom WordPress plugin three times for three different clients with three different data structures. Same work, different shapes. ModulaCMS makes the shapes runtime-configurable so you build the tooling once.

And ModulaCMS genuinely does not care how much of it you use. If all you ever use is the media upload and optimization pipeline — S3 storage, automatic WebP conversion, responsive dimension presets, focal point cropping — then that's your entire integration surface. Hit the media endpoints. Get optimized images. Ignore everything else. Use one endpoint or use all of them. ModulaCMS is a toolbox, not a religion. Take what you need and leave the rest in the drawer.

### Performance

Go isn't a trendy choice. It's a boring choice, and boring is exactly what you want from the language running your production infrastructure. Compiled to native machine code. Goroutines for concurrency without callback spaghetti. The result is a CMS that starts faster than most CMSs finish loading their configuration, serves thousands of concurrent requests without flinching, and runs three servers in less memory than a typical Node process uses to import its dependencies.

The entire ModulaCMS binary is 27-29 MB. That's the whole CMS — three servers, the admin panel, the TUI, the plugin runtime, the media pipeline, the audit system, the deploy engine, all of it. The Node.js runtime alone — before you install a single package — is 80 MB.

And critically, ModulaCMS knows what it is and knows what it isn't. It is a fast-as-hell JSON pumping machine. You ask for content, it gives you content, at a speed that will make you double-check your latency metrics because surely something got cached somewhere. It didn't — it's just that fast. That's the job. That's the whole job.

ModulaCMS is not a CDN. It is not a frontend framework. It is not a caching layer. It is not a server-side rendering engine. Those are separate problems with separate solutions backed by companies whose entire job is to solve them. Cloudflare exists. Vercel exists. Fastly exists. ModulaCMS doesn't try to be a worse version of all of them stapled together. It does one thing — serve your content as structured data, screaming fast, over HTTP — and lets the rest of your stack handle the rest of your stack.

Content assembly, tree traversal, field resolution, and collection queries are all handled server-side in Go. Every query is as fast as a hand-optimized SQL query from any other CMS. The client gets a single, complete JSON response — no N+1 queries, no waterfall requests, no client-side assembly.

#### Deterministic JSON with flexible composition

When content is published, the entire tree — content_data, datatypes, content_fields, fields, and route metadata — is snapshotted as frozen JSON in a `content_versions` row. This snapshot is what gets served to clients. The JSON is deterministic: it's exactly what was published, not a live query that might return different results depending on what else changed in the database since publish.

But content trees can reference other content trees via reference datatypes. A landing page can reference a testimonials section, which references a client logo carousel. These references are resolved at runtime — each referenced tree's own published snapshot is fetched and embedded into the response. A cache layer prevents duplicate fetches within a single request.

The result:
- Your landing page's tree is frozen at publish time — deterministic, predictable, cacheable
- The testimonials section it references always reflects the latest published version of the testimonials tree
- Change the testimonials once, every page that references them gets the update — without republishing the pages themselves
- The client gets a single, complete JSON response with all referenced content already embedded

This is content reuse that actually works. Not shortcodes pointing to other shortcodes where deleting the source renders `[testimonial_slider id=404]` in plain text to your client's customers. The base tree is frozen. The references are live. Broken references produce system log nodes instead of crashing. The response is always complete, always valid, always fast.

### Transparency

You should know where your stuff is. This sounds obvious. It is not obvious to most CMSs.

ModulaCMS doesn't make storage decisions for you. You configure where media goes — an S3 bucket, a MinIO instance, whatever S3-compatible storage you control. You configure where backups go — local directory or S3, your choice, and you set the path. You configure the database connection string. You chose SQLite, MySQL, or PostgreSQL, and you know exactly where it lives because you put it there. Every file has a location you specified. Every backup has a path you defined. Every mutation the system makes is logged in an audit trail you can query.

There are no black boxes. If you decide tomorrow that ModulaCMS isn't for you, you take everything with you — the database is standard SQL, the media is in your bucket, the backups are ZIP files you can open with literally any computer. No exit interview. No data hostage negotiation.

One `modula.config.json` file with every setting in one place — database, server, auth, S3, email, CORS, output format, plugins, deploy environments, observability. One file. Every text editor on earth can read it. No admin panel treasure hunt across seventeen screens. No PHP constants. No database rows pretending to be configuration. You can read it, diff it, version-control it, template it with environment variables, and update it at runtime with hot-reload support.

## Data Authority, Not Behavior Authority

This is the architectural principle that connects all three values.

The CMS is a **data authority** — it stores, organizes, resolves, and serves structured content. It is not a **behavior authority** — it never assumes how clients render, route, or respond to that data.

**The CMS answers: "What is at this URL?"**

**The client decides: "What do I do about it?"**

### How this works in practice

A typical ModulaCMS client is a Next.js, Nuxt, or SvelteKit app with a catch-all route like `[...slug]`. Every page request passes through that route, calls the CMS API with the URL path, and renders whatever comes back. The CMS is the routing brain. The client is the renderer.

When the CMS receives a request for `/old-blog`, it might respond:

```json
{
  "type": "redirect",
  "source": "/old-blog",
  "target": "/blog",
  "status": 301,
  "preserve_query": true
}
```

The CMS does not return an HTTP 301. It returns structured data describing a redirect. The client framework decides how to execute it. A Next.js app calls `redirect()`. A mobile app navigates. A static site generator records it in its output config. The CMS made no assumptions about the consumer.

When the CMS receives a request for `/blog`, it does the hard data work — slug resolution, tree traversal, field assembly, status filtering — and returns the complete content structure as JSON. The client gets a fast, complete response and renders it however it wants.

### What the CMS owns vs what the client owns

The CMS has enough baked-in behavior to handle complex data operations — tree assembly, collection queries, publish workflows, media optimization — but never forces that behavior on the client.

| CMS owns (data + assembly) | Client owns (behavior) |
|---|---|
| Slug → content tree resolution | Rendering, routing, page layout |
| Redirect mappings and data | Whether and how to execute redirects |
| Content tree assembly + field resolution | Component mapping and display |
| Publish/draft status, scheduled publishing | Preview mode, access gates, UI treatment |
| Media optimization and srcset metadata | Image display, lazy loading, responsive strategy |
| Collection queries and pagination | How results are displayed and navigated |
| Locale resolution and available translations | Language picker, locale routing, fallback display |

### What we deliberately don't do

These are intentional boundaries, not missing features:

- **No HTML rendering** — Content is structured data. The CMS never generates markup.
- **No HTTP redirects** — The CMS returns redirect data. The client executes it.
- **No caching headers** — Caching strategy belongs to the client's hosting layer.
- **No error pages** — The CMS returns error data. The client renders error UI.
- **No analytics** — The CMS doesn't track page views or user behavior.
- **No sitemap generation** — The CMS provides data to build a sitemap. The client generates the XML.

Each boundary exists because crossing it means the CMS is making assumptions about how the client works. Every assumption is a future constraint.

## REST API + SDKs

Every CMS capability is accessible through the REST API. The API returns JSON with consistent structure, typed IDs, and predictable pagination. There is no hidden behavior.

The SDKs (TypeScript, Go, Swift) provide opt-in convenience:

- **Type safety** — Typed responses, branded IDs, autocompletion
- **Common patterns** — Redirect handling, tree traversal, content assembly, pagination
- **Error handling** — Typed errors with helper methods (`IsNotFound()`, `IsUnauthorized()`)

**Don't use the SDK** and you get raw data with full control. **Use the SDK** and you get convenience without losing access to the underlying API. Both are fully supported, first-class ways to use ModulaCMS.

The SDK is the answer to "but you're making the developer do more work." No — we're giving them the choice. The SDK does the work if they want it to. The API is there if they don't. We never force a pattern on you. We never hide behavior behind magic.

## Scaling Without Migration

ModulaCMS doesn't make you choose your scale upfront.

Day one, you're a small agency. One binary on one server. SQLite for the database. Total infrastructure cost: a $5 VPS. Year two, traffic is real and SQLite isn't cutting it. Change one string in `modula.config.json` from `"sqlite"` to `"postgres"`, point it at a managed database, done. No migration tool. No export-import ritual. Same binary, same admin panel, same everything — just a different database engine underneath.

Year five, you're running a distributed network. Multiple instances behind a load balancer. The audit trail's hybrid logical clocks handle distributed ordering. The plugin coordinator syncs state across instances. The permission cache refreshes independently per instance with lock-free reads. None of this was bolted on. It was all there from the first commit, waiting for you to need it.

Same codebase, same config format, same API, same SDKs — from a single binary on a budget VPS to a globally distributed fleet. The only CMS that grows with you instead of growing out from under you.

## Contributing: The Decision Framework

When building new features for ModulaCMS, apply this test:

1. **Does it produce or organize data?** → CMS core. Store it, index it, serve it through the API.
2. **Does it interpret data into client-side behavior?** → SDK. Provide a convenience method that developers can opt into.
3. **Can it work within the existing datatypes/fields model?** → Prefer that over new tables. Adding a new content type should never require a schema migration for end users.
4. **Does it make assumptions about the consumer?** → Reconsider. The consumer might be a browser, a mobile app, a static site generator, a CLI tool, or something that doesn't exist yet.

If you're unsure, default to exposing data. It's always possible to add an SDK convenience method later. It's much harder to remove baked-in behavior that clients depend on.

And remember: ModulaCMS doesn't care how much of it you use. Every feature should be independently useful. No feature should require another feature as a prerequisite unless there's a genuine data dependency.
