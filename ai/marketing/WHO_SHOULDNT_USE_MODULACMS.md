# Who Shouldn't Use ModulaCMS

**Created:** 2026-01-16
**Purpose:** Honest assessment of who ModulaCMS is NOT for (prevents bad fits, builds trust)

---

## The Philosophy of Honesty

**Good companies say:** "Our product is for everyone!"

**Great companies say:** "Our product is perfect for X, but not for Y."

**ModulaCMS is great for:** Developers, agencies, and businesses that want to build modern web applications with complete control.

**ModulaCMS is NOT for:** People who want a no-code, click-to-install, GUI-everything solution.

---

## Who Should NOT Use ModulaCMS

### 1. Non-Technical Users Without Developer Support

**If you are:**
- A blogger who just wants to write
- A small business owner with no technical skills
- Someone who's never used a command line
- Someone without access to a developer

**Then ModulaCMS is NOT for you.**

**Why:**
```
ModulaCMS requires:
- Setting up a server (or Docker)
- Configuring a database (SQLite, MySQL, PostgreSQL)
- Building or customizing a frontend (Next.js, React, etc.)
- Understanding APIs and HTTP requests
- Using SSH (for TUI management)

Without technical skills or developer help:
→ You'll be stuck at step 1
→ Frustration, not productivity
```

**Use instead:**
- **Squarespace** - Beautiful, no-code, all-in-one
- **Wix** - Drag & drop, everything visual
- **WordPress.com** (hosted) - Click install, pick theme, write
- **Ghost** (hosted) - Simple blogging, minimal setup

**When ModulaCMS makes sense:**
- You hire a developer for initial setup
- Your agency builds it for you
- You have an in-house tech team

---

### 2. Teams That Need No-Code Content Management

**If your content team:**
- Needs a visual page builder (drag & drop)
- Can't write Markdown or HTML
- Needs WYSIWYG editors for everything
- Has zero technical skills

**Then ModulaCMS is NOT for you.**

**Why:**
```
ModulaCMS TUI (terminal interface):
- Text-based forms (not GUI)
- SSH access required
- Markdown/rich text in editors
- No visual page builders

Your content editors will:
→ Be confused by terminal UI
→ Miss drag & drop builders
→ Struggle with SSH
```

**What you need instead:**
- **WordPress** - Visual Gutenberg editor, drag & drop
- **Webflow** - Visual designer, no-code
- **Contentful** (with custom UI) - GUI content editing
- **Sanity** - Studio UI with visual editors

**When ModulaCMS works:**
- Content team uses the custom React admin panel (you build)
- Developers manage content (technical team)
- Content in Markdown is acceptable

---

### 3. Projects That Need Specific WordPress Plugins

**If your project requires:**
- WooCommerce (WordPress e-commerce)
- Yoast SEO (WordPress-specific SEO tools)
- Elementor / Divi (WordPress page builders)
- WordPress-specific integrations (20,000+ plugins)

**Then ModulaCMS is NOT for you.**

**Why:**
```
ModulaCMS has no WordPress plugin ecosystem:
- No WooCommerce equivalent
- No Yoast SEO
- No drag & drop builders
- Different architecture entirely

Migrating would mean:
→ Rebuilding features from scratch
→ Losing WordPress-specific functionality
→ Retraining team on new system
```

**Stick with WordPress if:**
- You rely on specific plugins
- Your team knows WordPress well
- Migration cost > staying cost

**ModulaCMS makes sense when:**
- You're starting fresh (greenfield project)
- You want to escape WordPress limitations
- You're willing to rebuild features your way

---

### 4. Enterprise Teams Without DevOps Capability

**If your organization:**
- Has no DevOps or infrastructure team
- Relies entirely on managed services (SaaS)
- Can't manage servers or databases
- Needs 24/7 enterprise support out of the box

**Then self-hosted ModulaCMS is NOT for you.**

**Why:**
```
ModulaCMS (self-hosted) requires:
- Server management (Linux, Docker)
- Database administration (backups, scaling)
- Monitoring and alerting
- Security updates
- Load balancing (for scale)

Without DevOps team:
→ Who manages servers?
→ Who handles incidents?
→ Who scales infrastructure?
→ Who ensures uptime?
```

**Use instead:**
- **Contentful** - Fully managed, enterprise SLA
- **Sanity** - Managed hosting, enterprise support
- **WordPress VIP** - Managed WordPress, enterprise-grade
- **Adobe Experience Manager** - Full enterprise solution

**When ModulaCMS works:**
- You have DevOps team (even 1-2 people)
- You're comfortable with infrastructure
- You want control over hosting
- You can build monitoring/alerting

**Or:**
- Wait for managed ModulaCMS offering (future)
- Use deployment platforms (Render, Railway, Fly.io)

---

### 5. Projects That Need Multi-Language Out of the Box

**If you need:**
- Built-in multi-language support (i18n)
- Automatic translation workflows
- Language-specific routing
- Locale management UI

**Then ModulaCMS doesn't have this built-in (yet).**

**Why:**
```
ModulaCMS today:
- No built-in i18n fields
- No automatic language routing
- No translation workflow
- Would need to build it yourself

For multi-language:
→ Create language field on datatypes
→ Build language switcher in frontend
→ Manage translations manually
```

**Better alternatives:**
- **Contentful** - Built-in localization
- **Sanity** - Document-level translations
- **Strapi** - i18n plugin
- **WordPress** - WPML, Polylang plugins

**When ModulaCMS works:**
- Single language only
- You build custom i18n solution
- Language handled in frontend (Next.js i18n)

**Future:** ModulaCMS may add i18n features, but not today.

---

### 6. Legacy System Integration Requirements

**If you must integrate with:**
- Legacy .NET applications (tight coupling)
- Legacy Java enterprise systems
- Old SOAP APIs
- Windows-only infrastructure
- Internet Explorer 11 support

**Then ModulaCMS might not fit.**

**Why:**
```
ModulaCMS is modern:
- Go backend (no .NET, Java, PHP)
- REST/GraphQL API (no SOAP)
- Runs on Linux (not Windows-first)
- Modern browser targets

Legacy integration:
→ May require middleware
→ May require adapters
→ May be complex
```

**Better for legacy:**
- **Sitecore** (.NET ecosystem)
- **Adobe Experience Manager** (Java ecosystem)
- **Umbraco** (.NET CMS)

**When ModulaCMS works:**
- Modern APIs (REST, GraphQL)
- Legacy system has API gateway
- You build integration middleware

---

### 7. Teams That Can't Customize (Limited Budget)

**If you:**
- Need everything pre-built and working
- Can't afford developer customization
- Want to use it "as-is" forever
- Have no budget for building admin panel

**Then ModulaCMS starter kit might not be enough.**

**Why:**
```
ModulaCMS philosophy:
- Core is minimal (data + API)
- Admin panel: Pre-built starter OR custom
- Frontend: Templates OR custom build
- Features: Add what you need

To get full value:
→ Customize admin panel (developer time)
→ Build custom frontend (developer time)
→ Add integrations (developer time)

Cost: Developer time (your biggest investment)
```

**Better if budget-limited:**
- **WordPress** - Themes + plugins = instant features
- **Shopify** - E-commerce out of the box
- **Squarespace** - Everything pre-built
- **Ghost** - Simple blogging, minimal setup

**When ModulaCMS makes sense:**
- You have developer budget
- You value long-term ownership over short-term speed
- Customization is an investment (compounds over time)

---

### 8. Compliance-Heavy Industries (Without Audit Budget)

**If you're in:**
- Healthcare (HIPAA compliance)
- Finance (PCI DSS, SOX)
- Government (FedRAMP, FISMA)

**And you need:**
- Pre-audited systems
- Compliance certifications
- SOC 2 Type II reports
- Vendor attestations

**Then open-source ModulaCMS needs audit work.**

**Why:**
```
ModulaCMS (open source):
- No SOC 2 report (you're self-hosting)
- No HIPAA attestation
- No compliance certifications
- You're responsible for compliance

Regulated industries:
→ Need to audit the code
→ Need to configure securely
→ Need to document controls
→ Need to prove compliance
```

**Better for compliance (out of box):**
- **Contentful** - SOC 2, HIPAA, GDPR certified
- **Adobe Experience Manager** - Enterprise compliance
- **Sitecore** - Compliance-ready

**When ModulaCMS works:**
- You have compliance team (can audit)
- You configure security controls yourself
- You document your compliance
- You accept responsibility

**Or wait for:**
- Managed ModulaCMS (future) with certifications

---

### 9. Real-Time Collaboration Features

**If you need:**
- Google Docs-style collaboration (multiple editors, same content)
- Real-time cursors and presence
- Live editing with conflict resolution
- Operational transformation

**ModulaCMS doesn't have this (yet).**

**Why:**
```
ModulaCMS today:
- Traditional save/update model
- No WebSocket real-time sync
- No collaborative editing
- Last write wins

For real-time collab:
→ Would need WebSocket layer
→ Would need conflict resolution
→ Would need presence system
→ Complex implementation
```

**Better for collaboration:**
- **Notion** - Real-time collaboration built-in
- **Sanity** - Real-time editing
- **Contentful** (Enterprise) - Live editing

**When ModulaCMS works:**
- Traditional editing workflow
- One person edits at a time
- Async collaboration acceptable

---

### 10. Projects That Need Advanced Workflow (Today)

**If you need (out of box):**
- Complex approval workflows (6-step review)
- Editorial calendar with scheduling
- Content governance (mandatory reviews)
- Automated publishing pipelines
- Advanced role-based workflows

**ModulaCMS has basic workflow only.**

**Why:**
```
ModulaCMS today:
- Basic status (draft, published, scheduled)
- Simple roles and permissions
- No complex workflow engine
- No approval chains

Advanced workflow:
→ Build it yourself (datatypes + Lua plugins)
→ Or integrate external workflow tool
→ Or wait for workflow plugins
```

**Better for complex workflow:**
- **Adobe Experience Manager** - Advanced workflow engine
- **Sitecore** - Workflow and approval chains
- **Contentful** (Enterprise) - Custom workflows

**When ModulaCMS works:**
- Simple workflow acceptable (draft → published)
- You build custom workflow (Lua plugins)
- Future: Workflow marketplace plugins

---

## Summary: Who Should NOT Use ModulaCMS

### Don't use ModulaCMS if you are:

1. ❌ **Non-technical user** without developer support
2. ❌ **Content team** that needs visual page builders
3. ❌ **WordPress power user** relying on specific plugins
4. ❌ **Enterprise** without DevOps capability
5. ❌ **Multi-language site** needing built-in i18n (today)
6. ❌ **Legacy integration** requiring .NET/Java tight coupling
7. ❌ **Budget-constrained** needing everything pre-built
8. ❌ **Compliance-heavy** without audit budget
9. ❌ **Real-time collaboration** requirements
10. ❌ **Complex workflow** requirements (today)

### Use ModulaCMS if you are:

1. ✅ **Developer** or have developer support
2. ✅ **Agency** building custom solutions for clients
3. ✅ **Growing business** wanting to escape WordPress
4. ✅ **Startup** building modern web app
5. ✅ **Enterprise** with DevOps team
6. ✅ **Tech-savvy team** comfortable with APIs
7. ✅ **Long-term thinker** valuing ownership over convenience
8. ✅ **Freedom seeker** wanting to escape vendor lock-in

---

## The Honest Comparison

| Need | ModulaCMS | Better Alternative |
|------|-----------|-------------------|
| No-code blogging | ❌ | Ghost, WordPress.com |
| Visual page builder | ❌ | Webflow, WordPress |
| WooCommerce | ❌ | WordPress + WooCommerce |
| Managed hosting | ⚠️ (DIY) | Contentful, Sanity |
| Multi-language (today) | ❌ | Contentful, Strapi |
| Real-time collaboration | ❌ | Notion, Sanity |
| Complex workflow (today) | ❌ | Adobe AEM, Sitecore |
| Compliance certifications | ❌ | Contentful (managed) |
| **Modern, headless, full control** | **✅** | **Nothing better** |
| **Custom admin panels** | **✅** | **Unique to ModulaCMS** |
| **Scales to enterprise** | **✅** | **Matches best CMSs** |
| **Zero vendor lock-in** | **✅** | **Open source advantage** |

---

## Why Honesty Matters

### Bad Fit Customer = Everyone Loses

**If we sell ModulaCMS to wrong customer:**
- ❌ Customer frustrated (can't use it)
- ❌ Bad reviews (not their fault, wrong tool)
- ❌ Support burden (trying to force square peg in round hole)
- ❌ Churn (they leave, tell others it's bad)
- ❌ Brand damage (ModulaCMS gets bad reputation)

**Better to say NO:**
- ✅ Customer finds right tool (happy elsewhere)
- ✅ Good reputation (honest, helpful)
- ✅ No support burden (no wrong-fit customers)
- ✅ Trust building (we care about fit, not sales)
- ✅ Right customers only (who will succeed)

### The Filter Effect

**Being honest about who shouldn't use ModulaCMS:**
- Filters OUT bad fits (saves everyone time)
- Filters IN good fits (who will succeed)
- Builds trust (we're honest, not salesy)
- Sets expectations (no surprises)
- Creates advocates (right people, right tool)

**Example:**
```
❌ Bad marketing: "ModulaCMS is for everyone!"
   → Everyone tries it
   → Wrong people frustrated
   → Bad reviews

✅ Good marketing: "ModulaCMS is for developers building modern apps"
   → Developers try it
   → Right fit
   → Great reviews, word of mouth
```

---

## How to Position This

### On Website / Docs

**Section: "Is ModulaCMS Right for You?"**

**✅ ModulaCMS is perfect if:**
- You're a developer or have developer support
- You want to build custom admin panels in React
- You're escaping WordPress or another CMS
- You value long-term ownership over short-term convenience
- You need a CMS that scales from small to enterprise

**❌ ModulaCMS might not fit if:**
- You need a no-code, visual page builder
- You rely on specific WordPress plugins
- You need enterprise compliance certifications (today)
- You have no DevOps capability (and need managed hosting)

**🤔 Not sure? Ask yourself:**
- Do we have developers? (If no → ModulaCMS is harder)
- Are we comfortable with APIs? (If no → steep learning curve)
- Do we want to build our own admin UI? (If yes → ModulaCMS is perfect!)

---

### In Sales Conversations

**When prospect asks: "Should we use ModulaCMS?"**

**Ask them:**
1. Do you have developers on your team?
2. Are you building a modern web application?
3. Do you want control over the admin interface?
4. Can you handle infrastructure (or use managed services)?
5. Are you comfortable with APIs and headless architecture?

**If 4+ YES:**
- ✅ "Yes, ModulaCMS is perfect for you!"

**If 2-3 YES:**
- ⚠️ "ModulaCMS could work, but consider [alternative] too"

**If 0-1 YES:**
- ❌ "I recommend [WordPress/Contentful/etc.] instead. Here's why..."

**Being honest builds trust** → They'll come back when they're ready

---

## The Long Game

### Today (Being Honest)

**We say:**
- ModulaCMS isn't for everyone
- No built-in i18n (yet)
- No managed hosting (yet)
- No real-time collaboration (yet)
- Requires technical skills

**Result:**
- Right customers (developers, agencies, tech-savvy teams)
- High satisfaction (tool matches needs)
- Great word of mouth (advocates)
- Room to grow (add features over time)

### Future (Expanding Addressable Market)

**As ModulaCMS adds features:**
- ✅ i18n support → multi-language sites welcome
- ✅ Managed hosting → enterprises without DevOps welcome
- ✅ Visual admin builder → less technical teams welcome
- ✅ Workflow engine → complex approval chains welcome
- ✅ Compliance certifications → regulated industries welcome

**But always:**
- Stay honest about fit
- Don't over-promise
- Guide people to right tool (even if not us)

---

## Conclusion: Strength Through Honesty

**Weak companies:** "Our product is for everyone! Buy now!"

**Strong companies:** "Our product is for X. If you're Y, use this instead."

**ModulaCMS position:**
> "We're the best CMS for developers and agencies building modern web applications.
>
> If you need no-code or visual builders, we're not there yet (use Webflow).
>
> If you rely on WordPress plugins, stay on WordPress.
>
> But if you want complete control, modern tech stack, and a CMS that grows with you from small business to enterprise without vendor lock-in...
>
> **We're the only CMS that can do that.**"

**Honesty = Trust = Long-term success**

---

**Last Updated:** 2026-01-16
