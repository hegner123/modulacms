# Headless Architecture Insight: Frontend Does Infrastructure

**Purpose:** Explore how pairing ModulaCMS with Next.js eliminates the need for most backend packages
**Created:** 2026-01-16
**Key Insight:** Headless CMS = Data layer. Next.js = Application layer. Stop building features that belong in the frontend!

---

## The Paradigm Shift

### Traditional CMS Architecture (WordPress, Drupal, Umbraco)

```
┌─────────────────────────────────────────┐
│         Monolithic CMS                  │
├─────────────────────────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ Frontend Rendering (PHP/Razor)      │ │
│ ├─────────────────────────────────────┤ │
│ │ Cache Layer (Redis/Memcached)       │ │
│ ├─────────────────────────────────────┤ │
│ │ Search (Elasticsearch)              │ │
│ ├─────────────────────────────────────┤ │
│ │ Image Processing                    │ │
│ ├─────────────────────────────────────┤ │
│ │ Email System                        │ │
│ ├─────────────────────────────────────┤ │
│ │ Background Jobs                     │ │
│ ├─────────────────────────────────────┤ │
│ │ Form Builder                        │ │
│ ├─────────────────────────────────────┤ │
│ │ Comments System                     │ │
│ ├─────────────────────────────────────┤ │
│ │ SEO Management                      │ │
│ ├─────────────────────────────────────┤ │
│ │ Content Storage (Database)          │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

**Why they need all this:**
Because the CMS does EVERYTHING - rendering, caching, processing, storage.

---

### Headless CMS Architecture (ModulaCMS + Next.js)

```
┌─────────────────────────┐         ┌─────────────────────────┐
│     Next.js Frontend    │         │   ModulaCMS Backend     │
├─────────────────────────┤         ├─────────────────────────┤
│ SSR/SSG/ISR Rendering  │         │                         │
│ Built-in Cache         │         │ Datatypes (Structure)   │
│ Image Optimization     │  HTTP   │ Fields (Schema)         │
│ Server Actions (Forms) │  API    │ Content Data (Tree)     │
│ API Routes (Comments)  │ ◄─────► │ Content Fields (Values) │
│ Middleware (Auth)      │         │ Media (S3)              │
│ Search Integration     │         │ Users & Sessions        │
│ Email (Resend)         │         │ Routes (Multi-site)     │
│ Analytics (Vercel)     │         │                         │
│ SEO Metadata           │         │ TUI Management (SSH)    │
│ Webhooks (Vercel Cron) │         │ Plugin System (Lua)     │
└─────────────────────────┘         └─────────────────────────┘
```

**Key insight:**
- **ModulaCMS** = Data storage + schema definition + content management
- **Next.js** = Everything else (rendering, caching, processing, integrations)

---

## What Next.js Provides Out of the Box

### 1. Caching (No Need for Backend Cache)

**Next.js built-in caching:**
```javascript
// Automatic caching with revalidation
async function getBlogPosts() {
  const res = await fetch('https://cms.example.com/api/posts', {
    next: { revalidate: 60 } // Cache for 60 seconds
  })
  return res.json()
}

// Incremental Static Regeneration (ISR)
export async function generateStaticParams() {
  const posts = await getPosts()
  return posts.map((post) => ({ slug: post.slug }))
}

// Static page with ISR
export const revalidate = 3600 // Revalidate every hour
```

**Edge caching (Vercel/Cloudflare):**
```javascript
export const runtime = 'edge' // Runs on edge network
```

**Data cache:**
```javascript
import { unstable_cache } from 'next/cache'

const getCachedPosts = unstable_cache(
  async () => getPosts(),
  ['posts-cache'],
  { revalidate: 3600, tags: ['posts'] }
)
```

**ModulaCMS doesn't need:**
- Cache datatype
- Redis integration
- Cache invalidation logic
- Cache cleanup jobs

**Why:** Next.js caches API responses automatically. Content changes? Revalidate with `revalidatePath()` or `revalidateTag()`.

---

### 2. Image Optimization (No Need for Backend Processing)

**Next.js Image component:**
```javascript
import Image from 'next/image'

// Automatic optimization, responsive images, lazy loading
<Image
  src="https://s3.amazonaws.com/bucket/image.jpg"
  alt="Product"
  width={800}
  height={600}
  sizes="(max-width: 768px) 100vw, 50vw"
  quality={85}
  priority={false}
/>
```

**What Next.js does automatically:**
- Resize images on-demand
- Convert to WebP/AVIF
- Responsive srcset generation
- Lazy loading
- Blur placeholder
- Edge caching

**ModulaCMS only needs:**
- Store original images in S3
- Return image URLs in API
- Optional: Store image metadata (width, height, alt)

**ModulaCMS doesn't need:**
- Advanced image transformations
- Multiple dimension presets
- On-demand resizing
- Format conversion

**Why:** Next.js does this automatically at the edge.

---

### 3. Search (No Need for Backend Search)

**Client-side search (small datasets):**
```javascript
import { useMemo, useState } from 'react'

function SearchPosts({ posts }) {
  const [query, setQuery] = useState('')

  const filtered = useMemo(() => {
    return posts.filter(post =>
      post.title.toLowerCase().includes(query.toLowerCase()) ||
      post.body.toLowerCase().includes(query.toLowerCase())
    )
  }, [posts, query])

  return (
    <>
      <input onChange={e => setQuery(e.target.value)} />
      {filtered.map(post => <Post key={post.id} {...post} />)}
    </>
  )
}
```

**Third-party search (large datasets):**
```javascript
// Algolia integration
import algoliasearch from 'algoliasearch/lite'

const searchClient = algoliasearch('APP_ID', 'API_KEY')
const index = searchClient.initIndex('posts')

// Search on user input
const results = await index.search(query)
```

**Webhook to update search index:**
```javascript
// Next.js API route: /api/webhooks/content-updated
export async function POST(request) {
  const { content } = await request.json()

  // Update Algolia index
  await index.saveObject({
    objectID: content.id,
    title: content.title,
    body: content.body,
    url: content.url
  })

  return Response.json({ success: true })
}
```

**ModulaCMS doesn't need:**
- Full-text search (SQLite FTS, PostgreSQL FTS)
- Search indexing
- Search API endpoints

**Why:** Frontend handles search directly, or integrates with search service (Algolia, Meilisearch).

---

### 4. Forms (No Need for Backend Form Builder)

**Next.js Server Actions:**
```javascript
// app/contact/actions.js
'use server'

import { z } from 'zod'

const schema = z.object({
  name: z.string().min(2),
  email: z.string().email(),
  message: z.string().min(10)
})

export async function submitContactForm(formData) {
  // Validate
  const validated = schema.parse({
    name: formData.get('name'),
    email: formData.get('email'),
    message: formData.get('message')
  })

  // Save to ModulaCMS
  await fetch('https://cms.example.com/api/form-submissions', {
    method: 'POST',
    body: JSON.stringify({
      datatype: 'contact_form',
      fields: validated
    })
  })

  // Send email (Resend)
  await resend.emails.send({
    from: 'forms@example.com',
    to: 'admin@example.com',
    subject: 'New Contact Form Submission',
    html: `<p>Name: ${validated.name}</p>...`
  })

  return { success: true }
}

// app/contact/page.jsx
import { submitContactForm } from './actions'

export default function ContactPage() {
  return (
    <form action={submitContactForm}>
      <input name="name" required />
      <input name="email" type="email" required />
      <textarea name="message" required />
      <button type="submit">Submit</button>
    </form>
  )
}
```

**ModulaCMS only needs:**
- Store form submissions as content_data
- Provide API endpoint to create submissions

**ModulaCMS doesn't need:**
- Form builder UI
- Form validation logic
- Form rendering engine
- Email sending

**Why:** Next.js handles forms with Server Actions. Email sent via Resend/SendGrid.

---

### 5. Comments (No Need for Backend Comments System)

**Next.js API Routes + ModulaCMS storage:**
```javascript
// app/api/comments/route.js
export async function GET(request) {
  const { searchParams } = new URL(request.url)
  const postId = searchParams.get('postId')

  // Fetch from ModulaCMS
  const comments = await fetch(
    `https://cms.example.com/api/content?datatype=comment&parent_id=${postId}`
  ).then(r => r.json())

  return Response.json(comments)
}

export async function POST(request) {
  const { postId, author, body } = await request.json()

  // Validate
  if (!body || body.length < 5) {
    return Response.json({ error: 'Comment too short' }, { status: 400 })
  }

  // Save to ModulaCMS
  await fetch('https://cms.example.com/api/content', {
    method: 'POST',
    body: JSON.stringify({
      datatype: 'comment',
      parent_id: postId,
      fields: {
        author,
        body,
        status: 'pending' // Moderation
      }
    })
  })

  return Response.json({ success: true })
}

// app/blog/[slug]/page.jsx
import Comments from './comments'

export default async function BlogPost({ params }) {
  const post = await getPost(params.slug)
  return (
    <>
      <article>{post.body}</article>
      <Comments postId={post.id} />
    </>
  )
}
```

**ModulaCMS only needs:**
- Store comments as content_data
- Tree structure for threaded replies
- API to create/read comments

**ModulaCMS doesn't need:**
- Comment rendering UI
- Spam detection (use Akismet in Next.js)
- Notification emails (Next.js sends)

---

### 6. Email (No Need for Backend Email System)

**Next.js + Resend/SendGrid:**
```javascript
// app/api/send-email/route.js
import { Resend } from 'resend'

const resend = new Resend(process.env.RESEND_API_KEY)

export async function POST(request) {
  const { to, subject, html } = await request.json()

  await resend.emails.send({
    from: 'noreply@example.com',
    to,
    subject,
    html
  })

  return Response.json({ success: true })
}

// Send welcome email on user registration
async function onUserRegister(user) {
  await fetch('/api/send-email', {
    method: 'POST',
    body: JSON.stringify({
      to: user.email,
      subject: 'Welcome!',
      html: `<h1>Welcome ${user.name}!</h1>`
    })
  })
}
```

**ModulaCMS doesn't need:**
- SMTP configuration
- Email templates
- Email queue
- Delivery tracking

**Why:** Next.js integrates with email services directly. No need for backend email system.

---

### 7. Background Jobs (No Need for Backend Queue)

**Vercel Cron Jobs:**
```javascript
// app/api/cron/publish-scheduled/route.js
export async function GET(request) {
  // Verify cron secret
  if (request.headers.get('authorization') !== `Bearer ${process.env.CRON_SECRET}`) {
    return Response.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // Fetch scheduled content from ModulaCMS
  const scheduled = await fetch(
    'https://cms.example.com/api/content?status=scheduled&publish_date_lte=now'
  ).then(r => r.json())

  // Publish each item
  for (const item of scheduled) {
    await fetch(`https://cms.example.com/api/content/${item.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ status: 'published' })
    })

    // Revalidate page
    revalidatePath(`/blog/${item.slug}`)
  }

  return Response.json({ published: scheduled.length })
}

// vercel.json
{
  "crons": [{
    "path": "/api/cron/publish-scheduled",
    "schedule": "* * * * *" // Every minute
  }]
}
```

**For heavy processing (image optimization, imports):**
```javascript
// Use queue service (Inngest, QStash, Trigger.dev)
import { Inngest } from 'inngest'

const inngest = new Inngest({ id: 'my-app' })

// Define job
export const optimizeImage = inngest.createFunction(
  { id: 'optimize-image' },
  { event: 'image.uploaded' },
  async ({ event }) => {
    // Process image
    const optimized = await optimizeImage(event.data.imageUrl)
    // Upload back to S3
    await uploadToS3(optimized)
  }
)

// Trigger job
await inngest.send({
  name: 'image.uploaded',
  data: { imageUrl: 'https://s3...' }
})
```

**ModulaCMS doesn't need:**
- Job queue system
- Worker processes
- Retry logic
- Job status tracking

**Why:** Vercel Cron for scheduled tasks, or use external queue service (Inngest, QStash).

---

### 8. Webhooks (No Need for Backend Webhooks)

**Next.js API Routes as webhook receivers:**
```javascript
// app/api/webhooks/cms-updated/route.js
export async function POST(request) {
  const { event, content } = await request.json()

  switch (event) {
    case 'content.published':
      // Revalidate page
      revalidatePath(`/blog/${content.slug}`)
      break

    case 'content.deleted':
      // Clear cache
      revalidateTag('posts')
      break

    case 'media.uploaded':
      // Optimize image via job queue
      await inngest.send({ name: 'image.uploaded', data: content })
      break
  }

  return Response.json({ success: true })
}
```

**Outgoing webhooks (notify external services):**
```javascript
// When content published, notify external service
async function onContentPublished(content) {
  await fetch('https://external-api.com/webhook', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ event: 'content_published', content })
  })
}
```

**ModulaCMS doesn't need:**
- Webhook delivery system
- Webhook queue
- Retry logic
- Webhook templates

**Why:** Next.js handles webhook logic. ModulaCMS just needs to call Next.js webhook URL when events happen.

---

### 9. SEO (No Need for Backend SEO Management)

**Next.js Metadata API:**
```javascript
// app/blog/[slug]/page.jsx
export async function generateMetadata({ params }) {
  const post = await getPost(params.slug)

  return {
    title: post.meta_title || post.title,
    description: post.meta_description,
    openGraph: {
      title: post.meta_title || post.title,
      description: post.meta_description,
      images: [post.og_image],
      type: 'article',
      publishedTime: post.published_at,
      authors: [post.author]
    },
    twitter: {
      card: 'summary_large_image',
      title: post.meta_title || post.title,
      description: post.meta_description,
      images: [post.og_image]
    },
    alternates: {
      canonical: post.canonical_url || `https://example.com/blog/${post.slug}`
    },
    robots: post.robots || 'index, follow'
  }
}
```

**Automatic sitemap generation:**
```javascript
// app/sitemap.js
export default async function sitemap() {
  const posts = await getPosts()

  return posts.map(post => ({
    url: `https://example.com/blog/${post.slug}`,
    lastModified: post.updated_at,
    changeFrequency: 'weekly',
    priority: 0.8
  }))
}
```

**ModulaCMS only needs:**
- Store SEO fields (meta_title, meta_description, og_image)
- Return SEO data in API

**ModulaCMS doesn't need:**
- SEO management UI (Next.js renders)
- Sitemap generation (Next.js generates)
- Robots.txt management (Next.js handles)
- Structured data (Next.js outputs)

---

### 10. Analytics (No Need for Backend Analytics)

**Client-side analytics:**
```javascript
// app/layout.jsx
import { GoogleAnalytics } from '@next/third-parties/google'

export default function RootLayout({ children }) {
  return (
    <html>
      <body>{children}</body>
      <GoogleAnalytics gaId="G-XXXXXXXXXX" />
    </html>
  )
}
```

**Vercel Analytics (built-in):**
```javascript
import { Analytics } from '@vercel/analytics/react'

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  )
}
```

**Server-side analytics (for bot tracking):**
```javascript
// middleware.js
import { track } from '@vercel/analytics/server'

export function middleware(request) {
  track('PageView', { path: request.nextUrl.pathname })
  return NextResponse.next()
}
```

**ModulaCMS doesn't need:**
- Analytics integration
- Event tracking
- Dashboard
- Reports

**Why:** Analytics happens entirely on frontend.

---

## What ModulaCMS ACTUALLY Needs

After removing all the features Next.js handles, ModulaCMS's scope becomes beautifully simple:

### Core Responsibilities

```
┌─────────────────────────────────────────┐
│          ModulaCMS Core                 │
├─────────────────────────────────────────┤
│ 1. Schema Definition                    │
│    - Datatypes (define structure)       │
│    - Fields (define field types)        │
│    - Routes (multi-site support)        │
│                                         │
│ 2. Content Storage                      │
│    - Content Data (tree structure)      │
│    - Content Fields (field values)      │
│    - Tree operations (O(1) lookups)     │
│                                         │
│ 3. Media Management                     │
│    - S3 integration                     │
│    - Upload handling                    │
│    - Media metadata                     │
│                                         │
│ 4. User Management                      │
│    - Users & roles                      │
│    - Sessions & authentication          │
│    - Permissions                        │
│                                         │
│ 5. Content API                          │
│    - REST API (read/write content)      │
│    - Tree queries                       │
│    - Field queries                      │
│                                         │
│ 6. Management Interface                 │
│    - TUI (SSH-based content editing)    │
│    - Content CRUD operations            │
│    - Media uploads                      │
│                                         │
│ 7. Extensibility                        │
│    - Lua plugin system                  │
│    - Plugin API                         │
└─────────────────────────────────────────┘
```

### Database Schema (All You Need)

```sql
-- Schema definition
CREATE TABLE datatypes (...);
CREATE TABLE fields (...);
CREATE TABLE datatypes_fields (...);  -- Junction
CREATE TABLE routes (...);

-- Content storage
CREATE TABLE content_data (...);      -- Tree structure
CREATE TABLE content_fields (...);    -- Field values

-- Media
CREATE TABLE media (...);
CREATE TABLE media_dimensions (...);

-- Users
CREATE TABLE users (...);
CREATE TABLE roles (...);
CREATE TABLE permissions (...);
CREATE TABLE sessions (...);
CREATE TABLE user_oauth (...);

-- Optional (minimal additions)
CREATE TABLE content_tags (...);      -- For many-to-many tags
CREATE TABLE webhooks (...);          -- Webhook configs
```

**That's it.** No cache tables, no job queues, no email templates, no form builders.

---

## Architecture Diagram

```
┌──────────────────────────────────────────────────────┐
│                  User's Browser                      │
└────────────────────────┬─────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────┐
│              Next.js Application                     │
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐   │
│  │   Pages    │  │   Server   │  │    API     │   │
│  │  (SSR/SSG) │  │  Actions   │  │   Routes   │   │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘   │
│        │               │               │           │
│  ┌─────▼───────────────▼───────────────▼─────┐    │
│  │         Content API Client                 │    │
│  │  (Fetch from ModulaCMS)                    │    │
│  └─────┬──────────────────────────────────────┘    │
│        │                                            │
│  ┌─────▼──────────┐  ┌─────────────────────────┐  │
│  │  Built-in      │  │  Third-party Services   │  │
│  │  Features:     │  │  - Resend (email)       │  │
│  │  - Image opt   │  │  - Algolia (search)     │  │
│  │  - Caching     │  │  - Inngest (jobs)       │  │
│  │  - SEO         │  │  - Vercel Analytics     │  │
│  └────────────────┘  └─────────────────────────┘  │
└────────────────────┬─────────────────────────────┘
                     │
                     │ HTTPS API
                     │
                     ▼
┌──────────────────────────────────────────────────────┐
│              ModulaCMS Backend                       │
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌────────────────────────────────────────────┐    │
│  │         Content API Endpoints              │    │
│  │  GET    /api/content                       │    │
│  │  POST   /api/content                       │    │
│  │  PATCH  /api/content/:id                   │    │
│  │  DELETE /api/content/:id                   │    │
│  │  GET    /api/datatypes                     │    │
│  │  GET    /api/routes                        │    │
│  │  POST   /api/media                         │    │
│  └────────────────┬───────────────────────────┘    │
│                   │                                 │
│  ┌────────────────▼───────────────────────────┐    │
│  │         Database Layer                     │    │
│  │  - Datatypes, Fields                       │    │
│  │  - Content Data (tree structure)           │    │
│  │  - Content Fields (values)                 │    │
│  │  - Users, Sessions, Permissions            │    │
│  │  - Media metadata                          │    │
│  └────────────────┬───────────────────────────┘    │
│                   │                                 │
│  ┌────────────────▼───────────────────────────┐    │
│  │    SQLite / MySQL / PostgreSQL             │    │
│  └────────────────────────────────────────────┘    │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │         S3 Storage (Media)                  │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │         TUI (SSH Management)                │   │
│  │  - Content editing                          │   │
│  │  - Datatype management                      │   │
│  │  - User management                          │   │
│  └─────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────┘
```

---

## What This Means for ModulaCMS Development

### Stop Building These (Next.js Handles Them)

1. ❌ Cache system (Next.js automatic caching)
2. ❌ Search indexing (Client-side or Algolia)
3. ❌ Image transformations (Next.js Image component)
4. ❌ Form builder (Server Actions)
5. ❌ Email system (Resend/SendGrid)
6. ❌ Background jobs (Vercel Cron or Inngest)
7. ❌ Webhook delivery (Next.js API routes)
8. ❌ SEO management UI (Next.js Metadata API)
9. ❌ Analytics (Vercel Analytics)
10. ❌ Comment rendering (Next.js components)

### Focus Development On These

1. ✅ **Rock-solid content API** (REST or GraphQL)
2. ✅ **Tree structure optimization** (O(1) operations)
3. ✅ **Multi-database support** (SQLite, MySQL, PostgreSQL)
4. ✅ **S3 integration** (reliable media storage)
5. ✅ **Flexible schema** (datatypes/fields system)
6. ✅ **TUI excellence** (best SSH content editing experience)
7. ✅ **Plugin system** (Lua extensibility)
8. ✅ **Security** (authentication, sessions, permissions)
9. ✅ **Performance** (fast queries, efficient tree operations)
10. ✅ **Developer experience** (great API docs, SDK)

---

## Benefits of This Architecture

### For ModulaCMS

1. **Simpler codebase** - Focus on data, not infrastructure
2. **Faster development** - No need to build features Next.js provides
3. **Easier maintenance** - Fewer packages, fewer bugs
4. **Better performance** - Edge caching, CDN, built-in optimizations
5. **Future-proof** - Next.js evolves, ModulaCMS stays simple

### For Users

1. **Best of both worlds** - Modern frontend + flexible CMS
2. **Performance** - Edge rendering, ISR, automatic optimization
3. **Flexibility** - Any React framework (Next.js, Remix, Gatsby)
4. **Developer experience** - Modern tooling, TypeScript, hot reload
5. **Scalability** - Edge deployment, automatic scaling

### For the Ecosystem

1. **Clear separation** - Data layer vs application layer
2. **Interoperability** - Works with any frontend framework
3. **Standards-based** - HTTP API, no proprietary protocols
4. **Composability** - Mix and match services (Algolia, Resend, etc.)

---

## Example: Full Blog Implementation

### ModulaCMS Setup (One-time)

```sql
-- Create Post datatype
INSERT INTO datatypes (name, slug) VALUES ('Post', 'post');

-- Create fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Title', 'title', 'text'),
    ('Slug', 'slug', 'text'),
    ('Body', 'body', 'richtext'),
    ('Featured Image', 'featured_image', 'image'),
    ('Meta Title', 'meta_title', 'text'),
    ('Meta Description', 'meta_description', 'textarea'),
    ('Published Date', 'published_date', 'datetime');

-- Link fields to datatype
INSERT INTO datatypes_fields (datatype_id, field_id, position) VALUES
    (<post_datatype_id>, <title_field_id>, 1),
    (<post_datatype_id>, <slug_field_id>, 2),
    ...;

-- Create Blog route
INSERT INTO routes (name, slug, datatype_id)
VALUES ('Blog', 'blog', <post_datatype_id>);
```

### Next.js Implementation

```javascript
// app/blog/page.jsx - Blog listing
export default async function BlogPage() {
  const posts = await fetch('https://cms.example.com/api/content?route=blog')
    .then(r => r.json())

  return (
    <div>
      <h1>Blog</h1>
      {posts.map(post => (
        <article key={post.id}>
          <h2><Link href={`/blog/${post.slug}`}>{post.title}</Link></h2>
          <time>{post.published_date}</time>
        </article>
      ))}
    </div>
  )
}

// app/blog/[slug]/page.jsx - Single post
export async function generateMetadata({ params }) {
  const post = await getPost(params.slug)
  return {
    title: post.meta_title || post.title,
    description: post.meta_description,
    openGraph: {
      images: [post.featured_image]
    }
  }
}

export default async function BlogPost({ params }) {
  const post = await getPost(params.slug)

  return (
    <article>
      <h1>{post.title}</h1>
      <Image src={post.featured_image} width={1200} height={630} />
      <time>{post.published_date}</time>
      <div dangerouslySetInnerHTML={{ __html: post.body }} />
    </article>
  )
}

// Static generation
export async function generateStaticParams() {
  const posts = await getPosts()
  return posts.map(post => ({ slug: post.slug }))
}

// ISR: Revalidate every hour
export const revalidate = 3600

// Helper function
async function getPost(slug) {
  return fetch(`https://cms.example.com/api/content?route=blog&slug=${slug}`)
    .then(r => r.json())
    .then(posts => posts[0])
}
```

**That's it!**
- ModulaCMS: Stores data, provides API
- Next.js: Renders, caches, optimizes, serves

---

## Migration Path for Existing Features

If you've already built backend features, migrate them to Next.js:

### 1. Cache → Next.js Cache
```javascript
// Before (ModulaCMS cache datatype)
const cached = await cache.Get('posts-list')

// After (Next.js fetch cache)
const posts = await fetch('...', { next: { revalidate: 60 } })
```

### 2. Cron Jobs → Vercel Cron
```javascript
// Before (ModulaCMS cron datatype)
// Cron job in database

// After (Next.js API route + vercel.json)
export async function GET() {
  // Publish scheduled content
}
```

### 3. Email Queue → Resend
```javascript
// Before (ModulaCMS email queue datatype)
await emailQueue.Add({ to, subject, body })

// After (Next.js + Resend)
await resend.emails.send({ to, subject, html: body })
```

### 4. Webhooks → Next.js API Routes
```javascript
// Before (ModulaCMS fetch_requests table)
// Webhook configs in database

// After (Next.js middleware triggers webhook)
await fetch(webhookUrl, { method: 'POST', body: JSON.stringify(event) })
```

---

## Conclusion: ModulaCMS is a Data API

**ModulaCMS's job:**
- Define schema (datatypes, fields)
- Store data (content_data, content_fields, tree structure)
- Provide API (REST endpoints)
- Manage content (TUI)
- Handle media (S3 integration)
- Authenticate users (sessions, OAuth)

**Everything else** is the frontend's job.

This is the **right architecture** for a headless CMS in 2026. Don't build features that belong in Next.js!

---

**Last Updated:** 2026-01-16
