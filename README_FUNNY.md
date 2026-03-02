# ModulaCMS

Born from one agency developer's slow descent into mass-produced psychic damage after years of fighting traditional CMSs.

You know the drill. Monday morning standup. The orthodontist's office website needs a before-and-after slider with a booking widget. The artisanal dog treat company wants their product page to have a "build your own treat box" configurator with real-time pricing. The divorce attorney's firm -- the one that took four months to approve a shade of blue -- now needs a live chat integration by Thursday because the partner's nephew said AI is the future. The HVAC company wants a "comfort calculator" that estimates your energy savings, and they want it to pull data from a weather API, and they want it on the homepage, and they want it yesterday, and also can you make the logo bigger?

And you're sitting there, four browser tabs into a WordPress plugin marketplace that smells like 2014, trying to jury-rig a page builder into doing things that would make its original developers file a restraining order. You've got a `functions.php` file that's three hundred lines of `add_filter` calls stacked on top of each other like a Jenga tower in an earthquake. The theme you bought for $59 has "customization options" that let you change exactly two colors and a font, and anything beyond that requires "editing the child theme" which is a phrase that has never once led to happiness. You install one plugin to fix the thing the last plugin broke and now the site takes eleven seconds to load. The client asks why.

ModulaCMS is what happens when that developer finally snaps, mass-archives every repo, and says "I'll just build the whole thing myself." And then actually does.

It's a headless CMS written in Go -- not PHP wheezing under the weight of its own backwards compatibility, not a Node.js runtime that needs 200MB of RAM to say hello, not a Python framework that runs great right up until you have two concurrent users. Go. A compiled language that starts in milliseconds, handles ten thousand concurrent connections without breaking a sweat, and compiles to a single binary with zero runtime dependencies. No interpreter. No VM. No "well first you need to install nvm to install node to install npm to install yarn to install the thing that installs the other things." One `go build`. One file. Copy it to the server. It runs. That's it. That's the deploy.

And because it's headless, your frontend team can use whatever they want -- React, Vue, Svelte, Astro, plain HTML, a fax machine, carrier pigeons -- ModulaCMS doesn't care. It serves content over HTTP and gets out of the way. Your CMS is no longer a lifestyle choice that dictates your entire tech stack. It's infrastructure. It sits there, it's fast, it scales, and it doesn't have opinions about your CSS framework.

Three concurrent servers in one process: HTTP, HTTPS with automatic Let's Encrypt certificates (so you never have to explain TLS to a project manager again), and SSH running a full terminal UI because some of us peaked when we discovered `vim` and we're not going back. A web admin panel built with HTMX and templ for the mouse-dependent. An MCP server so AI assistants can manage your content (the robots are here and they want CRUD access). SDKs in TypeScript, Go, and Swift. And an admin content system that lets you build one master admin panel and customize it per client without ever forking a repo again -- because the only thing worse than building an admin panel is building the same admin panel fifteen times.

## Who This Is For

Let's be clear about something: ModulaCMS was not built so your uncle can spin up a WordPress site for his fishing blog. It was not built for the "I watched a YouTube tutorial and now I'm a web developer" crowd. There is no one-click install from a shared hosting cPanel. There is no marketplace of themes with names like "flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor flavor" and a preview screenshot that looks nothing like what you'll actually get. This is not that.

ModulaCMS is built for agency developers doing real work for real clients -- the kind of work where you're managing thirty sites across a dozen verticals, where the content model for a healthcare provider looks nothing like the content model for a restaurant chain, where you need per-client permissioning and audit trails and deploy pipelines that don't involve FTP and a prayer. It's built for developers who have been burned by every "easy" CMS that turned into a prison the moment the client's requirements exceeded "blog with a contact form." It's built for the people who already know what they're doing and just need tools that stay out of their way and don't fall over at scale. If you need a drag-and-drop page builder and a library of stock photos, there are excellent options for you. This isn't one of them. This is the power tool aisle. Bring your own safety goggles.

Let's be real for a second though. ModulaCMS is not for everyone. It's not for the tech-illiterate. It's not for people who want to drag and drop a website into existence during their lunch break. And honestly, unless you've personally felt these pain points -- felt them in your bones, in your billable hours, in the quiet desperation of a Thursday afternoon support ticket -- you won't see the benefit. And that's fine.

Unless you've sat in front of an Advanced Custom Fields repeater with forty rows and watched the admin UI take fifteen seconds to reload while the browser tab begs for mercy like it's running Halo 2 over a 56k modem, you won't understand why a terminal-based content editor that responds in milliseconds matters. Unless you've maintained a spreadsheet of thirty client WordPress logins -- each one with a different username convention, half of them with expired passwords, a quarter of them with two-factor auth tied to someone who left the company in 2021 -- and you have to log into each one individually just to fix a typo on their About page, you won't see why SSH-based content management where one key gets you into every instance is worth building an entire TUI for. Unless you've watched a client's media library vanish into thin air because the hosting provider migrated servers and the uploads directory didn't come along for the ride, or because someone confused the staging uploads with production, or because the backup plugin only backed up the database and not `wp-content/uploads` and nobody noticed for six months, you won't understand why ModulaCMS makes S3-compatible storage the default and treats "where are my files?" as a question that should always have an immediate, concrete answer.

Unless you've spent an afternoon debugging why your frontend can't talk to your backend and the answer turned out to be that one is on HTTPS and the other is on HTTP and the browser's mixed content policy is silently eating your API calls -- no error in the console, no failed request in the network tab, just nothing, a void where your data should be -- and you end up down a rabbit hole of self-signed certificates, `mkcert`, `openssl req -x509`, browser trust stores, and a fifteen-step guide on Stack Overflow that was written for Ubuntu 16.04 and doesn't work on anything else, you won't understand why ModulaCMS ships with a built-in certificate generator and platform-specific install commands that handle the trust chain for you. One command. Certs generated. Trusted by your local machine. HTTPS works. Frontend talks to backend. Nobody cries.

If you've ever had a CORS error, you already know why `config.json` has an entire section dedicated to it. If you haven't, count your blessings and skip this paragraph. For the rest of us -- the ones who've seen `Access-Control-Allow-Origin` in their nightmares, who've added `"*"` to every header in desperation at 11 PM and still gotten blocked, who've read the MDN article on preflight requests four times and understood it less each time, who've debugged a POST that works in Postman but explodes in the browser because nobody told you that `Content-Type: application/json` triggers a preflight and your server doesn't handle OPTIONS -- ModulaCMS puts origins, methods, headers, and credentials in a flat JSON block where you can see them, edit them, and stop guessing.

These aren't hypothetical problems. These are the scars. Every feature in ModulaCMS exists because someone, somewhere, at some agency, at some point in the last decade, stared at a screen and said "there has to be a better way." If you've never said that, ModulaCMS probably isn't for you. If you've said it so many times it's become a mantra, welcome home.

## Core Values

Everything in ModulaCMS traces back to three principles. Not the kind of "core values" you find on a corporate about page next to a stock photo of people high-fiving. Actual architectural decisions that shaped every line of code.

### Flexibility

You define the schema. Not us. Not a plugin author. Not whatever opinionated framework decided that every piece of content needs a "slug," a "featured image," and exactly one taxonomy called "categories." Your content model is yours. Datatypes, fields, tree structures, reference composition -- the building blocks are simple enough to learn in an afternoon and powerful enough to model anything a client can dream up, including the things they dream up six months after launch that "definitely weren't in the original scope."

And ModulaCMS genuinely does not care how much of it you use. Most CMSs want to be your entire world. They want you all in. They want to manage your content, your users, your auth, your media, your emails, your analytics, your deployment, your morning coffee order, and they get increasingly passive-aggressive when you try to use only part of them. ModulaCMS has no such ego. If all you ever use it for is the media upload and optimization pipeline -- S3 storage, automatic WebP conversion, responsive dimension presets, focal point cropping -- then that's your entire integration surface. Hit the media endpoints. Get optimized images. Ignore everything else. The content tree doesn't mind. The RBAC system isn't offended. The Lua plugin runtime won't send you a push notification asking if you've considered using it lately. Use one endpoint or use all 47. Build a full content-driven site or use it as a glorified image optimizer with an audit trail. ModulaCMS is a toolbox, not a religion. Take what you need and leave the rest in the drawer.

### Performance

Go isn't a trendy choice. It's a boring choice, and boring is exactly what you want from the language running your production infrastructure. Compiled to native machine code. Goroutines for concurrency without callback spaghetti. Garbage collection that doesn't stop the world for half a second every time your heap gets interesting. The result is a CMS that starts faster than most CMSs finish loading their configuration, serves thousands of concurrent requests without flinching, and runs three servers in less memory than a typical Node process uses to import its dependencies. Performance isn't a feature we bolted on. It's the foundation everything else stands on.

And critically, ModulaCMS knows what it is and knows what it isn't. It is a fast-as-hell JSON pumping machine. You ask for content, it gives you content, at a speed that will make you double-check your latency metrics because surely something got cached somewhere (it didn't -- it's just that fast). That's the job. That's the whole job. ModulaCMS is not a CDN. It is not a frontend framework. It is not a caching layer. It is not a build system. It is not an edge runtime. It is not a server-side rendering engine that also does static generation that also does incremental static regeneration that also does partial prerendering that also does -- whatever Next.js announced this week. Those are separate problems with separate solutions backed by multi-billion dollar companies with thousands of engineers whose entire job is to solve them. Cloudflare exists. Vercel exists. Fastly exists. They're very good at what they do. ModulaCMS doesn't try to be a worse version of all of them stapled together. It does one thing -- serve your content as structured data, screaming fast, over HTTP -- and lets the rest of your stack handle the rest of your stack. The UNIX philosophy didn't die, it just stopped being fashionable for a while.

### Transparency

You should know where your stuff is. This sounds obvious. It is not obvious to most CMSs. Here's a real story: a team is running a headless CMS on an Azure compute instance. They're in healthcare. The uploads need to be HIPAA-compliant, organized in a specific bucket structure, access-logged, the whole nine yards. Someone built the upload pipeline. The files are going up. Everything looks fine in the admin panel. The thumbnails render. The client is uploading patient-facing documents. Life is good.

Except nobody verified that the upload pipeline was actually connected to the storage bucket. The CMS was quietly writing files to its own local filesystem -- the local filesystem of the Azure compute unit it was running on. A compute unit you don't have SSH access to, can't browse the filesystem of, and can't pull files from. Your client's HIPAA-sensitive media library is disappearing into a sealed box in a Microsoft data center that you can see the outside of but cannot open. The person who built the pipeline either assumed the CMS was doing the right thing or didn't care enough to check. The CMS, for its part, was perfectly happy to write files into the void and tell you everything was fine. No warnings. No errors. Just a green checkmark and a lie.

You find this out when the compute instance recycles and the files vanish. The backups were also on the local filesystem. Of the same compute instance. That just recycled. The HIPAA-compliant storage bucket is empty. It was always empty. You have a meeting about this and no one makes eye contact.

ModulaCMS doesn't let this happen because ModulaCMS doesn't make storage decisions for you. You configure where media goes -- an S3 bucket, a MinIO instance, whatever S3-compatible storage you control. You configure where backups go -- local directory or S3, your choice, and you set the path. You configure the database connection string. You chose SQLite, MySQL, or PostgreSQL, and you know exactly where it lives because you put it there. Every file has a location you specified. Every backup has a path you defined. Every decision the system makes is logged in an audit trail you can query. There are no black boxes. There are no sealed compute instances eating your uploads. If you decide tomorrow that ModulaCMS isn't for you, you take everything with you -- the database is standard SQL, the media is in your bucket, the backups are ZIP files you can open with literally any computer made since 1995. No exit interview. No data hostage negotiation. No "please contact our enterprise sales team to discuss your export options."

## The Numbers

ModulaCMS doesn't boast a five-minute install from a landing page with a stock photo of a woman pointing at a laptop and smiling. It doesn't promise you'll "have your site live before your coffee gets cold." What it does offer is a cold start measured in milliseconds -- not seconds, not "please wait while we spin up your container," milliseconds -- because it's a compiled binary, not an interpreted language warming up its JIT compiler and loading fourteen thousand autoloaded class files.

The entire ModulaCMS binary is 27-29 MB. That's the whole CMS. Three servers, the admin panel, the TUI, the plugin runtime, the media pipeline, the audit system, the deploy engine, all of it. The Node.js runtime alone -- before you install a single package, before `node_modules` exists, before you've typed `npm init` -- is 80 MB. Your JavaScript runtime is three times the size of this entire CMS and it hasn't done anything yet. It's just sitting there. Breathing. Waiting for you to install Express and 847 transitive dependencies so it can serve a JSON response.

In production, ModulaCMS will realistically sit at 500 MB to 1 GB of memory for a normal deployment. That's three concurrent servers, active database connections, the permission cache, the plugin VMs, all of it. You would have to actively try to push it further -- we're talking a stress test running while a full database backup is executing while a hundred people are simultaneously SSH'd into the TUI uploading images through the media pipeline. That apocalyptic scenario caps out at 4-5 GB. That's the ceiling. That's the "everything is on fire and someone invited the entire company to upload their vacation photos at once" number. Most Node.js CMS deployments idle higher than ModulaCMS peaks under duress.

And here's the part that sounds like marketing but isn't: a random $5 shared-CPU Linux box running ModulaCMS will outperform a Vercel deployment. Not "compete with." Outperform. A single-core VPS with 1 GB of RAM, the kind of server you provision to run a hobby project and then forget about, will serve content faster than a globally distributed edge network backed by a company valued at billions of dollars. Because ModulaCMS is a compiled binary doing one thing extremely well on hardware it has all to itself, and Vercel is a serverless platform spinning up cold-start functions, routing through edge middleware, and adding latency at every layer of abstraction between your content and your user. The $5 box doesn't have a CDN. It doesn't need one. It responds before the CDN would have finished its TLS handshake.

Think about that for a second. Go is the language that runs behind every microservice powering every Node server on the planet. Kubernetes? Go. Docker? Go. The API gateways, the load balancers, the service meshes, the infrastructure that your infrastructure depends on -- it's all Go. Every time your Node.js application makes a request, it's passing through layers of Go services that are routing, load-balancing, health-checking, and orchestrating at speeds your application layer will never touch. That entire ecosystem of Go services exists because the companies that run the internet at scale tried other languages first and then rewrote everything in Go when the traffic got real.

ModulaCMS puts that same language -- the one trusted to power the backbone of the modern internet -- directly behind your content API. No microservice chain. No thirty-server relay race. When Netflix loads your watch list, that request bounces through a constellation of services -- API gateway, auth service, user profile service, recommendation engine, content metadata service, availability service, A/B testing service, analytics service -- each one adding latency, each one a potential point of failure, each one maintained by a different team with a different deployment schedule and a different opinion about error handling. That architecture makes sense when you're Netflix and you have thousands of engineers and billions of requests. It does not make sense when you're serving blog posts.

ModulaCMS is one binary. One process. One endpoint. Your request comes in, hits the handler, queries the database, and the response is on its way back before a microservice architecture would have finished resolving its first service discovery lookup. The speed that Netflix spent hundreds of millions of dollars and thousands of engineer-hours achieving across a distributed system -- you get that from a single Go binary answering content queries. Not because ModulaCMS is doing anything clever. Because it's not doing anything stupid.

## The Content Model

Every CMS you've ever worked with has the same dirty secret: the content model is a lie. Under the hood it's one god-table -- `wp_posts`, `nodes`, whatever they're calling it this decade -- and every content type, every custom field, every taxonomy is just metadata duct-taped onto the side of it. Want a blog post? That's a row in the posts table. Want a product? Also a row in the posts table. Want a landing page with twelve custom sections? Believe it or not, posts table. You end up with a single table doing the work of thirty, joined to a `postmeta` key-value dumping ground with four million rows where every query is a crime against indexing and your DBA has started leaving passive-aggressive sticky notes on your monitor.

ModulaCMS takes a fundamentally different approach. Five core tables and you can model anything:

**Datatypes** are your building blocks -- "Blog Post," "Product," "Hero Section," "Testimonial Card," whatever you need. They define what kinds of content exist. **Fields** define the details -- a title, a price, a body of rich text, an image reference, a boolean toggle. Fields get assigned to datatypes, so a "Blog Post" might have a title, author, body, and featured image while a "Product" has a name, price, description, and SKU. **Content data** is an instance of a datatype -- not a row crammed into a god-table, an actual typed instance. "This specific blog post. This specific product." **Content fields** hold the values for that instance. And **routes** tie it all together: a route assigns content data to a URL path, and now you've got a tree. A real tree, with parent-child relationships and sibling ordering, not a flat list with a `menu_order` column and a dream.

The tree uses sibling pointers for O(1) navigation and reordering:

- `parent_id` -- who's your daddy (node)
- `first_child_id` -- the favorite child
- `next_sibling_id` / `prev_sibling_id` -- the doubly-linked sibling rivalry

The tree build algorithm has four phases: create nodes, assign hierarchy, resolve orphans (because data is messy), and reorder by stored sibling pointers with circular reference detection (because your data is *really* messy).

And then it gets interesting: tree composition means you can embed trees inside other trees. Your landing page tree can reference your testimonial tree, which references your client logo tree. The `ComposeTrees()` function recursively fetches and embeds referenced trees up to 10 levels deep using concurrent goroutine resolution via errgroup. A `CachedFetcher` prevents duplicate tree fetches within a single request. Broken references produce `_system_log` nodes instead of errors because crashing is rude. Change the testimonials once, every page that references them updates. It's content reuse that actually works, not "shortcodes pointing to other shortcodes and if someone deletes the source shortcode the whole page renders `[testimonial_slider id=404]` in plain text to your client's customers."

Content lifecycle: **draft** -> **pending** -> **published** -> **archived** (a.k.a. the five stages of content grief, minus bargaining)

**The admin variants** (`admin_content_data`, `admin_content_fields`, `admin_datatypes`, `admin_datatype_fields`, `admin_fields`, `admin_routes`) are where it gets unhinged in the best way: the admin panel itself is CMS-managed content. The same content tree, datatypes, and fields that power your client's website also power the admin UI that manages it.

This is the "one admin panel to rule them all" play. If you're an agency, you build one master admin panel project. Client A needs a media library, a blog editor, and an SEO dashboard? Turn those on. Client B just needs a product catalog? Turn on only that. Client B calls six months later and says "actually we want the blog editor too"? Flip a switch. Same admin panel, same codebase, per-client feature toggles managed as content. No more maintaining fifteen forked versions of the same white-labeled admin. No more "wait, which client has the fix for that one bug?" No more merging upstream changes into a graveyard of divergent repos. One project. Per-client configuration. It's a CMS for your CMS, and it might be the most agency-brained feature ever built.

Now, we just told you that you can model literally anything with five tables and infinite schema flexibility, and we can feel the decision paralysis setting in from here. You're staring at an empty datatypes list thinking "I can build anything" with the same energy as standing in front of a blank canvas with every paint color ever manufactured. Total freedom is exhilarating for about forty-five seconds and then it's just terrifying.

We know. Decision overload is real, and telling a developer "you can do anything" is sometimes just a fancy way of saying "you're on your own, good luck." So ModulaCMS ships with predefined datatype schemas -- blog post, page, product, navigation, FAQ, testimonial, the greatest hits of content modeling -- ready to go out of the box. Pick one, and you've got a working content type with sensible fields in seconds, not hours of whiteboarding "what should a blog post even have" while your project manager hovers.

But here's where it's different from every other CMS that offers starter templates: you're not locked in. At all. These aren't sacred immutable templates baked into the core that you have to "override" through some arcane theming layer. They're just datatypes and fields, same as anything you'd build yourself. Don't need the excerpt field on your blog post? Remove it. Want to add a custom "mood" dropdown because your client's content team insists on tagging every article with a vibe? Add it. Want to gut the entire predefined schema and rebuild it from scratch because you have opinions? Go for it. The starter schemas are a suggestion, not a sentence. They get you to "something works" in minutes, and then they get out of your way.

Five tables. Model anything. A blog, a product catalog, a healthcare portal, a restaurant menu, a government document archive, whatever the cleaning company is doing this week. Add a field to a datatype and existing content doesn't break. Remove a field and nothing explodes. Rename, restructure, reorganize -- the schema flexes because it was designed to flex, not because you've tortured a `VARCHAR(255)` column into holding JSON that holds YAML that holds a prayer.

This means schema changes are just data. Which means syncing schema changes between environments is just syncing data. Which means deploying from local to dev to staging to production is no longer a white-knuckle `wp db export | wp db import` pipe dream where you hold your breath and hope the serialized PHP in your options table didn't get corrupted by a find-and-replace on the domain name. You export. You import. The schema version is hash-validated. The payload is integrity-checked. Dry-run tells you exactly what will change before anything changes. It's boring. It's reliable. After years of deployment anxiety, boring and reliable might bring you to tears.

## The Plugin System

ModulaCMS took one look at the WordPress plugin ecosystem and had a religious experience. Not the good kind. The kind where you see the mass of `wp_options` rows injected by plugins you uninstalled three years ago, the jQuery versions fighting each other like gladiators in your `<head>` tag, the premium SEO plugin that quietly phones home with your entire sitemap, the contact form widget that -- for reasons known only to God and a developer in 2011 -- has full write access to your users table. The entire model is "here's the keys to everything, we trust you, please don't burn the house down." And then the house burns down. Every single time. A plugin updates, your site goes white-screen, and you're in an emergency SSH session deactivating plugins by renaming folders like a bomb disposal technician cutting wires.

ModulaCMS uses an embedded Lua runtime via gopher-lua, and the philosophy is the polar opposite: your plugins run in sandboxed virtual machines and can only do what you explicitly allow them to do. A plugin declares what it is. It declares what it wants -- which tables it needs, which content hooks it wants to intercept, which HTTP routes it wants to register. And then it asks you for approval. You review the request. You approve or deny. The plugin gets access to its own isolated tables and nothing else. It cannot touch core CMS tables. It cannot touch other plugins' tables. It cannot silently inject ads into your footer. It cannot quietly log user data to a third-party server. It cannot do anything you didn't specifically say "yes" to. If it misbehaves anyway -- infinite loop, runaway memory, repeated failures -- a three-state circuit breaker (Closed -> Open -> Half-Open with probe recovery) trips it automatically and shuts it down. No white screen. No emergency SSH session. No bomb disposal. Just a notification that says "this plugin broke itself" and a button to re-enable it when the author fixes their code.

Now, "single binary" is a beautiful phrase until the part of your brain that's been burned by closed ecosystems kicks in and whispers: "but what if I need it to do something it doesn't do?" Fair. Every monolith you've ever worked with has had that moment where you hit a wall and the answer was "sorry, that's not supported, here's a feature request form that feeds directly into a black hole." The fear is real: a single binary sounds like a single point of rigidity.

The plugin system is the answer to that fear, and it's not a half-measure. Plugins get their own database tables -- not some key-value junk drawer, real tables with real columns, real indexes, real query performance. The plugin query builder generates proper SQL with parameterized values and validated identifiers. Your plugin's data lives in the same database engine as the rest of the CMS, benefits from the same connection pooling, the same transaction guarantees, the same backup pipeline. If you're on PostgreSQL, your plugin tables are PostgreSQL tables. If you need a composite index on three columns for a complex lookup, you define it in your schema and you get it. The five-table content architecture handles the content modeling; the plugin system handles everything else. Need a custom analytics pipeline? Plugin. Need a webhook dispatcher? Plugin. Need a client-specific workflow that doesn't fit any standard CMS pattern because the client's business process was designed by a committee that couldn't agree on lunch? Plugin. The single binary doesn't limit what ModulaCMS can do. It limits what ModulaCMS has to do before you start extending it.

### What Plugins Can Do

- **Database** -- query builder with safe identifier validation, isolated per-plugin tables. Your plugin gets its own sandbox. No touching the core tables.
- **Content Hooks** -- before/after hooks on create, update, delete, publish, archive. Intercept content lifecycle events like a content bouncer.
- **Pipeline Registry** -- DB-backed, in-memory pipeline registry keyed by `table.operation`. Uses a build-then-swap pattern for lock-free reads. Your plugin hooks are fast.
- **HTTP Routes** -- register custom endpoints with an admin approval workflow. This isn't the Wild West.
- **Logging** -- structured logging that integrates with the CMS slog logger, not `print("here")`.

### Safety

- Operation counting per VM (default 1000 ops, so your infinite loop doesn't become our infinite loop)
- Timeouts on everything (2 seconds per hook, 5 seconds per event, 0 seconds of patience for runaway scripts)
- Plugins can't access core tables or other plugins' tables (good fences make good plugins)
- **Multi-instance coordination**: a `Coordinator` polls the plugins DB table and reconciles local state. When Instance A disables a plugin, Instance B finds out on the next tick. No gossip protocol needed, just a database and patience.
- **9 named metrics** (`plugin.http.requests`, `plugin.hook.duration_ms`, `plugin.circuit_breaker.trip`, etc.) integrated with the observability layer. You can watch your plugins misbehave in real-time.

### Plugin CLI

```bash
modula plugin list               # What have we got?
modula plugin init my-plugin     # Scaffold a new plugin (we give you a template)
modula plugin validate ./path    # Check if your plugin is structurally sound
modula plugin reload my-plugin   # Hot-reload without restart
modula plugin enable my-plugin   # Let it run
modula plugin disable my-plugin  # Never mind

modula pipeline list             # See all registered pipelines
modula pipeline show <table>     # Pipelines for a specific table
modula pipeline enable <id>      # Enable a pipeline
modula pipeline disable <id>     # Disable a pipeline
```

Example plugins live in `examples/plugins/` -- a `hello_world` starter and a `task_tracker` with custom DB tables and shared Lua library modules.

## The Configuration Question

Every CMS has settings. What separates the professionals from the masochists is where they put them. WordPress scatters yours across seventeen different admin screens like an Easter egg hunt designed by a sadist. Your site title? That's under Settings > General. Your permalink structure? Settings > Permalinks. Your homepage display? Settings > Reading. Reading. The setting that controls what your entire website shows when someone visits the front page lives under a menu called "Reading," sandwiched between "posts per page" and "search engine visibility," because apparently in 2003 someone decided that "what does the homepage show" is a reading comprehension issue. Want to change your SMTP settings? That's a plugin. Want to configure CORS headers? That's a different plugin. Want to set up cron jobs? That's `wp-config.php`, which is a PHP file sitting in your web root with your database password in it, and also it controls whether WordPress auto-updates itself, and also it defines your authentication salts, and also half the constants in it are cargo-culted from a Stack Overflow answer from 2009 and nobody remembers why they're there.

ModulaCMS puts everything in one `config.json` file:

| Category | What It Controls |
|----------|-----------------|
| **Database** | `db_driver`, `db_url`, `db_name`, `db_user`, `db_password` |
| **Server** | `port`, `ssl_port`, `ssh_port`, `environment`, `cert_dir` |
| **Auth** | `auth_salt`, `cookie_*`, OAuth secrets |
| **S3 Storage** | `bucket_endpoint`, `bucket_media`, `bucket_backup`, access keys |
| **Email** | `email_enabled`, `email_provider` (SMTP/SendGrid/SES/Postmark), sender config |
| **CORS** | `cors_origins`, `cors_methods`, `cors_headers`, `cors_credentials` |
| **Output** | `output_format` (contentful/sanity/strapi/wordpress/clean/raw) |
| **Plugins** | `plugin_enabled`, `plugin_directory`, `plugin_max_vms`, `plugin_timeout` |
| **Deploy** | `deploy_environments` (name, URL, API key per environment) |
| **Observability** | `observability_enabled`, `observability_provider` (Sentry/Datadog/New Relic) |

One place. One format. Every text editor and every programming language on earth can read it. No admin panel treasure hunt. No PHP constants. No database rows pretending to be configuration. You can read it, diff it, version-control it, template it with environment variables (`${VAR}` or `${VAR:-default}`), and update it at runtime through the admin panel or API with hot-reload support. When you change your email provider, the sender swaps atomically. When you adjust CORS rules, they take effect immediately. When something needs a restart, it tells you. Like an adult.

## Scaling

Here's the thing every CMS gets wrong about scaling: they make you choose upfront. WordPress is for small sites until it isn't, and then you migrate to something "enterprise." Contentful is enterprise from day one and charges you accordingly, even when you're three people with twelve pages and a dream. Strapi is fine until your traffic spikes and then you discover that "self-hosted headless CMS" and "handles load gracefully" are two different promises. Every CMS is built for a specific size of company, and the moment you outgrow it -- or the moment you realize you're overpaying for it -- you're looking at a migration. Another migration. The third one this decade. Each one taking months and costing more than the last.

ModulaCMS doesn't make you choose because ModulaCMS works at every scale. Day one, you're a small agency. You spin up the pre-built admin panel, buy a cheap theme, connect a frontend, and you're live. One binary on one server. SQLite for the database because it's Tuesday and you don't feel like configuring Postgres. Total infrastructure cost: that same $5 VPS we keep talking about. Your client is happy. Your boss is happy. You're slightly less unhappy than usual.

Year two, the agency is growing. You've got fifteen clients, traffic is real, and SQLite isn't cutting it anymore. You change one string in `config.json` from `"sqlite"` to `"postgres"`, point it at a managed database, and you're done. No migration tool. No export-import prayer ritual. No downtime. The same binary, the same admin panel, the same everything -- just a different database engine underneath.

Year five, you're running a distributed network. Multiple ModulaCMS instances behind a load balancer, spread across regions. The audit trail's hybrid logical clocks handle distributed ordering. The plugin coordinator syncs plugin state across instances via the database. The deploy sync system pushes content between environments with schema validation and integrity checks. The permission cache refreshes independently per instance with lock-free reads. None of this was bolted on. None of this required a "enterprise edition" upgrade. It was all there from the first commit, waiting for you to need it, not charging you a dime while you didn't.

From a single binary on a budget VPS to a globally distributed fleet -- same codebase, same config format, same admin panel, same API, same SDKs. The only CMS that grows with you instead of growing out from under you.

## The WordPress Question

And here's the question. The one that comes up at every conference, every Slack thread, every agency happy hour after someone's had two beers and wants to start something: "But does ModulaCMS do everything WordPress does? Is it feature-complete compared to WordPress?"

No. And it's the healthiest no you'll ever hear.

ModulaCMS does not do everything WordPress does because doing everything WordPress does is the problem. WordPress is a twenty-year-old PHP application that started as a blogging platform, became a page builder, became an e-commerce platform, became a learning management system, became a membership site, became a real estate listing engine, became a restaurant ordering system, became a patient intake portal, and somewhere along the way forgot how to do the one thing a headless CMS actually needs to do: serve JSON, stupid fast.

WordPress has spent two decades bolting on every feature anyone has ever asked for, and the result is an application that can technically do everything and does none of it particularly well. It's a Swiss Army knife where every blade is slightly dull and the corkscrew gave you tetanus. It has a REST API that was added as an afterthought in version 4.7, returns responses shaped like it's 2005, and performs about as well as you'd expect from a system that has to boot an entire PHP application, load a theme, initialize a plugin ecosystem, connect to MySQL, run through seventeen `apply_filters` chains, serialize a response through a legacy compatibility layer, and then -- finally, heroically -- hand you a JSON object that a Go binary would have served in the time it took PHP to open the database connection.

ModulaCMS doesn't have a twenty-year legacy to maintain. It doesn't have backwards compatibility with plugins written during the Obama administration. It doesn't need to make sure that the block editor still works alongside the classic editor alongside the customizer alongside the full site editor alongside whatever new editing paradigm ships next year. It does one thing. It serves structured content over HTTP at speeds that make WordPress's REST API look like it's sending responses by carrier pigeon. That's not a dig at the thousands of brilliant people who've contributed to WordPress. It's just what happens when you start from scratch in 2024 with a compiled language and a clear scope instead of inheriting two decades of accumulated ambition.

## And Just To Be Nice

Unless you've ever tried to onboard a new developer onto a project and the setup instructions were a three-page Google Doc that started with "first install PHP 8.1 (not 8.2, that breaks the thing)" and ended with "if the database seeder fails, ask Marcus, he knows the workaround" -- and Marcus left the company -- you won't appreciate that ModulaCMS ships with a full suite of Docker Compose files for every database backend. SQLite, MySQL, PostgreSQL, full stack, infrastructure only, per-database stacks, all of them. Getting started is `docker compose up`. Getting a new developer started is `docker compose up`. Getting your CI/CD pipeline started is `docker compose up`. It's `docker compose up` all the way down. No three-page Google Doc. No Marcus.

And speaking of not being locked in -- your email provider is yours to choose. SMTP, SendGrid, AWS SES, Postmark. If it sends emails, ModulaCMS can connect to it and send emails through it. Not "here's our proprietary email integration that only works with Mailchimp and costs extra." Not "email is a premium add-on, please upgrade to the Business tier." You point ModulaCMS at whatever email service you already use -- the one your ops team already trusts, the one that's already in your infrastructure budget, the one that already has your domain's SPF and DKIM records configured -- and it works. And if you switch providers, you change the config at runtime and the sender swaps atomically without a restart. Because your CMS changing email providers should not be a deployment event.

And just to be nice, we threw in OAuth for free. Not a vendored wrapper that only works with three providers and breaks the moment someone asks for Okta. Full spec-compliant OAuth that works with any OAuth provider -- Google, GitHub, Azure, Okta, Auth0, Keycloak, your company's cursed homegrown identity server that Dave built in 2017 and no one's been brave enough to replace. If it speaks OAuth, ModulaCMS speaks back. Because after everything you've been through, the least we can do is make sure your users never have to remember another password.

---

## Quick Start

- Go 1.24+ with CGO enabled (for SQLite, the dependency that keeps on giving)
- Linux or macOS (Windows users: we see you, we love you, we're sorry)
- [just](https://github.com/casey/just) as the build runner, because Makefiles are write-only code

```bash
# Build and run (auto-creates config.json because we're not animals)
just run

# Or build a local binary
just dev
./modula-x86 serve

# Interactive setup wizard for people who read instructions
./modula serve --wizard
```

On first run, ModulaCMS generates a config, creates the database schema, bootstraps three roles (admin, editor, and "viewer" which is a polite way to say "look but don't touch"), and logs a random admin password. Please write it down. We're not going to tell you again.

**Default Ports:**

| Server | Port | Emotional Support Level |
|--------|------|------------------------|
| HTTP   | 8080 | Casual                 |
| HTTPS  | 8443 | Responsible Adult      |
| SSH    | 2222 | Peak Performance       |

Connect to the TUI: `ssh localhost -p 2222` (and feel like a hacker in front of your non-technical friends)

## Build & Development

```bash
just dev              # Build local binary with version info because `go build` was too easy
just run              # Build and run, for the impatient
just run-admin        # Hot-reload dev server via air (for admin panel work)
just dev-admin        # Build with live static assets, no embed (dev mode)
just build            # Production binary for when you're feeling serious
just check            # Compile-check without producing artifacts (commitment issues)
just clean            # Remove build artifacts (digital decluttering)
just vendor           # Update vendor directory (hoard your dependencies locally like a dragon)
```

## Code Generation

This project has more code generators than some projects have code.

```bash
just sqlc             # Regenerate type-safe SQL code (layer 1: robot writes SQL)
just dbgen            # Regenerate entity wrapper files (layer 2: robot writes Go on top of robot SQL)
just dbgen-entity Users  # Just one entity, for the focused
just dbgen-verify     # CI check: are the robots up to date?
just admin generate   # Regenerate templ Go code from .templ files
just admin bundle     # Bundle the block editor via esbuild
```

The codegen pipeline: you write SQL schemas -> sqlc generates type-safe query code in three packages -> dbgen generates application-layer wrapper methods on all three database structs -> templ generates Go code from HTML templates. At this point the robots are doing most of the typing and you're just supervising.

## Testing

```bash
just test             # Run all tests (pray)
just coverage         # Tests with coverage report (find out how brave you really are)
just lint             # Run all linters (prepare to feel inadequate)

# Single package or test
go test -v ./internal/db              # Just the database, ma'am
go test -v ./internal/db -run TestSpecificName   # I know exactly what I broke

# S3 integration tests (requires MinIO, a.k.a. "S3 at home")
just test-minio       # Start MinIO container
just test-integration # Run integration tests against fake S3
just test-minio-down  # Stop MinIO, free your RAM

# Cross-backend integration tests (for the truly paranoid)
just test-integration-db   # MySQL + Postgres integration tests
```

## Docker

```bash
just docker-up        # Full stack. Everything. The whole circus.
just docker-dev       # Rebuild just the CMS container (speed run)
just docker-infra     # Infrastructure only, for the "I'll run Go locally" purists
just docker-down      # Stop containers, keep data (optimistic)
just docker-reset     # Stop containers AND delete data (nihilistic)
```

Can't decide on a database? Try them all:

```bash
just docker-sqlite-up     # For prototyping and vibes
just docker-mysql-up      # For legacy compatibility and regret
just docker-postgres-up   # For making your DBA proud
```

## Architecture

### Runtime

The `serve` command starts three servers sharing a single database connection because resource sharing builds character:

```
HTTP  (default :8080)  --+
HTTPS (default :8443)  --+-- stdlib ServeMux -- Middleware Gauntlet -- Handlers -- DbDriver
SSH   (default :2222)  --+   Charmbracelet Wish -- Bubbletea TUI --------------- DbDriver
                              Admin Panel (HTMX + templ) ----------------------- DbDriver
```

Graceful shutdown: first SIGINT says "wrap it up." Second SIGINT says "I SAID WRAP IT UP." Then everything dies. Shutdown order: HTTP servers, plugin system, database. It's like closing a restaurant -- customers out first, then the kitchen.

### Request Flow

```
Client Request
  -> Request ID (you are now a number, not a person)
  -> Logging (we're watching)
  -> CORS (are you even allowed to be here?)
  -> Authentication (prove it)
  -> Rate Limiting (calm down, it's just an API)
  -> Permission Injection (what are you allowed to touch?)
  -> Route Handler (finally, someone does actual work)
  -> DbDriver Interface (the universal translator)
  -> Database-specific wrapper (the one that speaks the local dialect)
  -> sqlc-generated queries (robot-written SQL, better than yours)
```

### Tri-Database

One codebase, three databases, zero regrets:

1. **SQL schemas** define tables per dialect because databases can't agree on anything, ever, in the history of computing
2. **sqlc** generates type-safe Go code so we never write SQL by hand and then pretend we didn't forget a comma
3. **dbgen** generates application-layer wrappers because writing the same method three times with slightly different NULL handling is no one's idea of fun
4. **`DbDriver` interface** (~150 methods) provides the contract -- it's basically a prenup between your app and your database
5. **Wrapper structs** convert between sqlc types and application types, handling the fact that SQLite thinks everything is an int64 (bless its heart)

Switch databases by changing one string in `config.json`. It's like changing the engine in your car while it's parked, which is the responsible way to do it.

### RBAC

| Role | Permissions | Vibe |
|------|-------------|------|
| **admin** | 47 (all) | "I am the law" |
| **editor** | 28 | "I can break most things" |
| **viewer** | 3 | "I'm just here to look" |

The `PermissionCache` uses build-then-swap for lock-free reads and refreshes every 60 seconds. System-protected roles can't be deleted because we learned the hard way what happens when someone deletes the admin role "to see what happens."

### Audit Trail

Every database mutation records:
- What changed (old and new values, in JSON, because we're thorough)
- Who did it (user ID, so there's no "wasn't me")
- Where they did it from (IP address, so there's really no "wasn't me")
- When (hybrid logical clock timestamps, for distributed ordering and courtroom drama)

Think of it as git blame, but for your entire database, with causal ordering across multiple instances.

## Admin Panel

A full web admin UI built with HTMX + templ because we believe the server should do the rendering and the browser should do the browsing.

- **templ** -- type-safe Go HTML templates that compile to Go code. It's like JSX but the compiler actually helps you.
- **HTMX** -- all interactions are HTML-over-the-wire swaps. `HX-Request` distinguishes partial vs full page renders. No JSON. No state management library. No existential crisis about whether to use Redux or Zustand.
- **Light DOM Web Components** -- 9 custom `mcms-*` elements (dialog, data-table, field-renderer, media-picker, tree-nav, toast, confirm, search, file-input) that enhance HTML without Shadow DOM because we want CSS to actually work.
- **Block Editor** -- built from scratch in vanilla JavaScript, bundled via esbuild. Supports `text`, `heading`, `image`, and `container` blocks with nesting up to 8 levels deep. Uses the same sibling-pointer tree structure as the CMS content model. Has drag-and-drop, validation, caching, and its own state management. It's basically a tiny CMS inside your CMS.
- **CSRF** -- double-submit cookie pattern. Not optional.

27 full-page templ components and 23 HTMX swap partials covering everything: content editing with tree navigation, datatype and field management, media browsing, user admin, role/permission management, route configuration, plugin management, session management, token management, SSH key management, import, audit logs, settings, and a dashboard. There's also a forgot-password page because humans are humans.

```bash
just admin generate       # Regenerate templ Go code
just admin watch          # Watch .templ files for changes
just admin bundle         # Bundle block editor via esbuild
just admin bundle-watch   # Watch and rebundle on change
just admin verify         # Verify everything is up-to-date (CI)
```

## Terminal UIs

Yes, UIs, plural. There are two.

**Classic TUI** (`internal/tui/`) -- the original 40+ file Bubbletea TUI accessible over SSH. 26+ screens for managing everything. Elm Architecture (Model-Update-View). Focus system for keyboard input routing. Custom form dialogs. Async commands for database operations. It's a whole application living inside your terminal.

**New TUI** (`internal/tui/`) -- a three-panel layout:

```
+----------------+------------------------+------------------+
|   Tree Panel   |    Content Panel       |   Route Panel    |
|     (25%)      |       (50%)            |    (remainder)   |
|                |                        |                  |
|  Navigate      |  Edit content          |  Configure       |
|  content tree  |  fields and data       |  URL slugs       |
+----------------+------------------------+------------------+
|  Header: [New] [Save] [Copy] [Duplicate] [Export]          |
|  Status: Focused panel + key hints                         |
+------------------------------------------------------------+
```

Uses `Composite()` for layered overlay rendering, lipgloss `RoundedBorder` with accent focus colors, and `TuiMiddleware()` for SSH session integration. It's like tmux but with opinions.

## MCP Server

Because in 2026, if your CMS can't be operated by an AI assistant, does it even exist?

The `mcp/` directory is a standalone Go module that wraps the ModulaCMS Go SDK into a Model Context Protocol server. It compiles to a `modula-mcp` binary that AI assistants (Claude, etc.) can connect to, with 19 tool categories covering the entire CMS API -- content, admin content, schema, routes, media, users, RBAC, config, deploy sync, health, import, sessions, tokens, SSH keys, OAuth, tables, and plugins.

```bash
just mcp-build        # Build the MCP server binary
just mcp-install      # Install to /usr/local/bin/modula-mcp
```

Two environment variables: `MODULACMS_URL` and `MODULACMS_API_KEY`. That's it. The robot is ready.

## Deploy Sync

For when you have more than one CMS instance and you need them to agree on reality.

```bash
modula deploy export --file backup.json           # Export content to a file
modula deploy import backup.json --dry-run        # See what would change (smart)
modula deploy import backup.json                  # Do the thing (brave)
modula deploy push production --dry-run           # Preview remote push
modula deploy push production                     # Push to remote
modula deploy pull staging                        # Pull from remote
modula deploy snapshot list                       # Manage import snapshots
modula deploy snapshot restore <id>               # Restore a snapshot
modula deploy env list                            # List environments
modula deploy env test production                 # Test connectivity before you push
```

Schema version validation (SHA256 of column layout), payload hash verification, user reference mapping across instances, gzip compression for files over 1GB, and dry-run support. It's like `terraform plan` for your content.

## SDKs

Three SDKs, three languages, zero external dependencies, one shared philosophy: your HTTP client library doesn't need 47 transitive dependencies.

### TypeScript

A pnpm workspace monorepo with three packages:

| Package | npm | Purpose |
|---------|-----|---------|
| `@modulacms/types` | Shared types, 30 branded IDs, enums | The foundation |
| `@modulacms/sdk` | Read-only content delivery | For your frontend |
| `@modulacms/admin-sdk` | Full admin CRUD | For your admin tools |

```typescript
import { ModulaClient } from "@modulacms/sdk"

const cms = new ModulaClient({
  baseUrl: "https://cms.example.com",
  defaultFormat: "clean",  // because "raw" is for sushi, not JSON
})

const page = await cms.getPage("blog/hello-world")
// Congratulations, you have content
```

Dual ESM + CommonJS builds via tsup. TypeScript 5.7+. Node 18+. Full type declarations. Zero dependencies.

### Go

```go
import modula "github.com/hegner123/modulacms/sdks/go"

client, err := modula.NewClient(modula.ClientConfig{
    BaseURL: "https://cms.example.com",
    APIKey:  "your-api-key",
})
// Yes, you have to check err. This is Go. We check errors here.

users, err := client.Users.ListPaginated(ctx, modula.PaginationParams{Limit: 20})
```

Generic `Resource[Entity, CreateParams, UpdateParams, ID]` pattern. 23+ typed resource endpoints. Branded ID types. `IsNotFound()` and `IsUnauthorized()` error helpers. Zero dependencies.

### Swift

```swift
import Modula

let client = try ModulaClient(config: ClientConfig(
    baseURL: "https://cms.example.com",
    apiKey: "your-api-key"
))

let page = try await client.content.getPage(slug: "blog/hello-world", format: "clean")
// Your Apple device now has content. Tim Cook would be proud.
```

Async/await throughout. `Sendable` types for actor isolation. 30 branded ID types via `ResourceID` protocol. iOS 16+, macOS 13+, tvOS 16+, watchOS 9+. Swift 5.9+. Zero dependencies.

### iOS App

There's an entire Xcode project at `ios/ModulaCMS Mobile/` with test targets and UI test targets. It consumes the Swift SDK. We don't talk about scope creep here, we call it "platform coverage."

```bash
just sdk-install      # pnpm install (TypeScript workspace)
just sdk-build        # Build all TypeScript packages
just sdk-test         # Run all SDK tests
just sdk-typecheck    # Typecheck all TypeScript packages
just sdk-go-test      # Run Go SDK tests
just sdk-swift-build  # Build Swift SDK
just sdk-swift-test   # Run Swift SDK tests
```

## API

All endpoints are prefixed with `/api/v1/` and follow standard REST conventions. Content delivery uses slug-based routing at `/api/v1/content/{slug}`. The `format` query parameter controls response structure: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, or `raw`.

```
POST   /api/v1/auth/login          # Prove you're you
POST   /api/v1/auth/logout         # Forget you were you
GET    /api/v1/auth/me             # Who am I? (existential)
POST   /api/v1/auth/register       # Become someone
POST   /api/v1/auth/reset          # Forgot who you are
GET    /api/v1/auth/oauth/login    # Let any OAuth provider vouch for you
GET    /api/v1/auth/oauth/callback # The OAuth dance, step 2
```

There are approximately 47 REST endpoints and honestly listing them all here would make this README longer than some novels. The admin panel, the TUI, the MCP server, and all three SDKs all talk to the same API. It's REST. It's paginated. It's permission-checked. Content CRUD, batch operations, admin variants, schema management, media upload with health checks and orphan cleanup, full RBAC management, import from 5 CMS formats, deploy sync, plugin and pipeline management, configuration with field metadata, session/token/SSH key management. You know the drill.

## CLI

```
modula serve              Start all three servers (the full experience)
modula serve --wizard     Guided setup (recommended for first dates with the CMS)
modula install            Installation wizard
modula install --yes      "I trust the defaults" speedrun
modula tui                Terminal UI without starting HTTP servers (introvert mode)
modula db init            Initialize database
modula db wipe            Drop all tables (requires nerves of steel)
modula backup create      Create backup (responsible adult behavior)
modula backup restore     Restore backup (for when you weren't a responsible adult)
modula config show        Print config (redacted, we're not that trusting)
modula config validate    Check if your config makes sense
modula config set         Update a config field
modula cert generate      Self-signed certs (for development, please)
modula cert check         Verify your certs are still valid
modula deploy export      Export content to a file
modula deploy import      Import content from a file
modula deploy push        Push content to a remote environment
modula deploy pull        Pull content from a remote environment
modula deploy snapshot    Manage import snapshots
modula deploy env         Manage deploy environments
modula plugin list        See what plugins you have
modula plugin init        Scaffold a new plugin
modula plugin validate    Check your plugin's structure
modula plugin reload      Hot-reload a plugin
modula plugin enable      Let it run
modula plugin disable     Time out
modula pipeline list      See registered pipelines
modula pipeline show      Pipelines for a specific table
modula pipeline enable    Enable a pipeline
modula pipeline disable   Disable a pipeline
modula update check       Is there a new version? (hope)
modula update install     Install the new version (courage)
modula version            Existential information about your binary
```

## Production

```bash
just deploy           # SSH-based deploy with health check + automatic rollback
just status           # Check production container status
just logs             # Tail production CMS logs
just rollback         # Roll back to previous Docker image (for bad days)
```

The deploy process pushes, health-checks, and rolls back automatically if the health check fails. Because deploying on a Friday afternoon should at least be recoverable.

## Project Structure

```
cmd/                          CLI commands (the front door)
mcp/                          MCP server for AI assistants (the robot door)
ios/                          iOS app (yes, really)
internal/
  admin/                      Web admin panel (HTML over the wire)
  admin/handlers/             Admin page handlers and render helpers
  admin/layouts/              templ layouts (base, admin, auth)
  admin/pages/                27 templ full-page components
  admin/partials/             23 templ HTMX swap targets
  admin/components/           Shared UI (sidebar, topbar, icon)
  admin/static/               CSS, JS, web components, block editor
  auth/                       Authentication (the bouncer)
  backup/                     Backup/restore (the insurance policy)
  bucket/                     S3 client abstraction (the storage valet)
  cli/                        Classic Bubbletea TUI (40+ files of terminal beauty)
  config/                     Configuration (the settings drawer)
  db/                         DbDriver interface, ~150 methods (the universal adapter)
  db/types/                   Typed IDs, enums, field configs (the compiler's anger management)
  db/audited/                 Audit trail (the surveillance system)
  db-sqlite/, db-mysql/,
  db-psql/                    sqlc-generated code (DO NOT EDIT or the robots get upset)
  definitions/                CMS format definitions (Contentful, Sanity, Strapi, WordPress)
  deploy/                     Deploy sync system (the content teleporter)
  email/                      Email service with 4 backends (the messenger)
  install/                    Setup wizard (the welcoming committee)
  media/                      Image optimization pipeline (the beauty filter)
  middleware/                  CORS, rate limiting, sessions, RBAC (the obstacle course)
  model/                      Domain structs (the blueprints)
  plugin/                     Lua plugins, pipeline registry, coordinator, circuit breaker, metrics
  router/                     HTTP route registration (the traffic cop)
  transform/                  Format transformers (the shapeshifters)
  tree/                       Content tree operations (the family tree therapist)
  tree/core/                  Shared tree algorithms: build, compose, fetch, traverse, mutate
  tui/                        New three-panel TUI (the upgrade)
  update/                     Self-update mechanism (the self-improvement plan)
  utility/                    Logging, version info, helpers (the junk drawer, but organized)
tools/
  dbgen/                      DB entity codegen tool (the second robot)
  transform_bootstrap/        Bootstrap data transformer
  transform_cud/              Create/Update/Delete audited method transformer
sql/
  schema/                     27 schema directories, 6 files each, 3 dialects (162 SQL files)
  sqlc.yml                    sqlc configuration (the first robot's instructions)
sdks/
  typescript/                 For the "everything is JavaScript" crowd
  go/                         For the "I like explicit error handling" crowd
  swift/                      For the "I own multiple Apple devices" crowd
examples/
  plugins/                    Example plugins (hello_world + task_tracker)
deploy/
  docker/                     Container configs (it works on my machine, shipped)
```

## Data Model

27 schema directories define the full entity model across three SQL dialects. Each directory has 6 files (schema + queries for SQLite, MySQL, PostgreSQL). That's 162 SQL files maintained in parallel. We're fine. Everything's fine.

| Entity Group | Tables | Emotional Weight |
|-------------|--------|-----------------|
| **Content** | content_data, content_fields, content_relations | The whole point |
| **Admin Content** | admin_content_data, admin_content_fields, admin_routes, admin_datatypes, admin_fields, admin_datatype_fields | The CMS for your CMS |
| **Schema** | datatypes, fields, datatype_fields | The point's skeleton |
| **Media** | media, media_dimensions | The pretty stuff |
| **Routing** | routes, admin_routes | How people find the point |
| **Users & Auth** | users, roles, permissions, role_permissions, tokens, user_oauth, sessions, user_ssh_keys | Who's allowed to touch the point |
| **System** | backups, change_events, tables, plugins, pipelines | The plumbing |

All entity IDs are 26-character ULIDs wrapped in distinct Go types. The type system prevents you from mixing up a `ContentID` and a `UserID` at compile time. This has prevented approximately one million bugs. (Citation needed, but it feels true.)

## CI/CD

Two GitHub Actions workflows:

- **Go** (`.github/workflows/go.yml`) -- runs on Go source changes (excludes `sdks/**`), tests with libwebp-dev on Ubuntu
- **SDKs** (`.github/workflows/sdks.yml`) -- runs on SDK changes, tests TypeScript (pnpm + Vitest), Go SDK, and Swift SDK (SPM build + test)

## Contributing

PRs welcome. Please include tests. "It works on my machine" is not a test strategy.

Note: there are two layers of codegen (sqlc + dbgen) and a template compiler (templ) and a JS bundler (esbuild). Run `just sqlc`, `just dbgen`, `just admin generate`, and `just admin bundle` before submitting. We know. We're sorry. The robots demand tribute.

## License

See [LICENSE](LICENSE) for the legal stuff that protects us both.

---

*Built with Go, spite, and an unreasonable amount of ambition for a single binary.*
