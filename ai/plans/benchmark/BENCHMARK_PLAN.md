# ModulaCMS Benchmark Plan

Date: 2026-03-10
Status: Not started
Repository: `modulacms-benchmarks` (sibling repo, not in core)

## Goal

Produce credible, reproducible performance comparisons between ModulaCMS and competing headless CMS platforms. Measure content delivery latency, throughput, and resource efficiency under identical conditions.

---

## Competitors

| CMS | Language | Why Include |
|-----|----------|-------------|
| **Strapi** | Node.js | Most popular open-source headless CMS |
| **Directus** | Node.js | Strong REST+GraphQL competitor |
| **Payload CMS** | Node.js | Fastest-growing Node.js CMS |

All support PostgreSQL. All have Docker images. Skip hosted platforms (Sanity, Contentful) — we're benchmarking software, not infrastructure. WordPress is excluded — its REST API is a bolt-on to a monolithic PHP application designed for server-rendered pages. Benchmarking a purpose-built headless CMS against it produces results that are obvious and unimpressive. If an "industry baseline" is needed later, it can be added as an optional appendix.

---

## What to Measure

### Primary Metrics

| Metric | Tool | Notes |
|--------|------|-------|
| Latency (p50, p95, p99) | k6 | Per-endpoint, per-concurrency level |
| Latency variance | k6 | IQR or stddev across runs — required for statistical significance |
| Throughput (req/s) | k6 | At saturation on fixed resources |
| Memory (idle + under load) | cAdvisor | Time-series at 500ms intervals, exported to JSON. RSS excluding shared library pages |
| CPU (under load) | cAdvisor | Time-series at 500ms intervals, percentage of allocated cores |
| Cold start time | Script | Container start to first successful content query (not just health check — see below) |
| Response payload size | k6 | Bytes per response (same logical content) |

### Secondary Metrics

| Metric | Tool | Notes |
|--------|------|-------|
| Error rate under load | k6 | 4xx/5xx at high concurrency |
| Latency degradation curve | k6 | How p99 changes from 1 to 500 concurrent users |
| Container image size | `docker images` | Compressed size |

### Resource Monitoring

**Do not use `docker stats`.** It provides point-in-time snapshots at ~1s intervals with no aggregation control. Instead:

- Run **cAdvisor** as a sidecar container during all benchmark runs
- Export metrics at 500ms intervals to JSON files per CMS per scenario
- Report: mean, p50, p95, max for both memory RSS and CPU percentage
- Memory measurements must distinguish container RSS from shared library pages

---

## Output Format

ModulaCMS supports 6 output formats via the transform layer: `raw`, `clean`, `contentful`, `sanity`, `strapi`, `wordpress`. Benchmarks use **`clean` format** for all ModulaCMS requests. Rationale:

- `raw` has the least overhead but returns internal structure (snapshot JSON) that is not comparable to what competitors return
- `clean` applies field resolution and tree formatting without mapping to another CMS's schema, making it the closest equivalent to Strapi/Directus/Payload native output
- Using a competitor-named format (`strapi`, `contentful`) would add transform overhead that competitors don't pay

The chosen format must be documented in the results metadata. If reviewers challenge the choice, the benchmark suite should support re-running with any format via a config flag.

---

## Snapshot Architecture Disclosure

ModulaCMS serves published content from pre-computed JSON snapshots stored in the database. Tree assembly, field resolution, and reference composition happen at **publish time**, not at request time. The delivery path deserializes the snapshot and composes referenced subtrees (which are also snapshots).

Competitors (Strapi, Directus, Payload) assemble responses from normalized database rows at **request time** — JOINs, relation resolution, and serialization happen per-request.

This is a legitimate architectural advantage, not an unfair test condition. However, the methodology must:

1. **Acknowledge this difference explicitly** in published results
2. **Report the publish cost separately** — time to publish all seeded content (one-time, not per-request)
3. **Frame the comparison honestly** — ModulaCMS trades write-time work for read-time speed. The benchmark measures read performance, which is the dominant production workload for a headless CMS

---

## Endpoints to Benchmark

Map equivalent operations across all CMS platforms:

### Content Delivery (Public, Read-Only)

| Scenario | ModulaCMS Endpoint | What It Tests |
|----------|-------------------|---------------|
| **Single item by slug** | `GET /api/v1/content/{slug}` | Core delivery path. Published snapshot lookup, field resolution, tree composition |
| **Collection query** | `GET /api/v1/query/{datatype}?limit=20` | Paginated listing with type filtering |
| **Collection + sort** | `GET /api/v1/query/{datatype}?limit=20&sort=created_at` | Sorted paginated listing |
| **Nested tree** | `GET /api/v1/content/{slug}` (deep tree) | Tree assembly with composed references — ModulaCMS's architectural advantage |
| **Globals** | `GET /api/v1/globals` | Multiple global content trees in one response |

### Equivalent Endpoints per Competitor

| Scenario | Strapi | Directus | Payload |
|----------|--------|----------|---------|
| Single item | `GET /api/{type}?filters[slug][$eq]={slug}&populate=*` | `GET /items/{collection}?filter[slug][_eq]={slug}&fields=*` | `GET /api/{collection}?where[slug][equals]={slug}&depth=2` |
| Collection | `GET /api/{type}?pagination[pageSize]=20` | `GET /items/{collection}?limit=20` | `GET /api/{collection}?limit=20` |
| Nested | `GET /api/{type}?filters[slug][$eq]={slug}&populate=deep` | `GET /items/{collection}?filter[slug][_eq]={slug}&fields=*.*.*` | `GET /api/{collection}?where[slug][equals]={slug}&depth=5` |

### Endpoint Mapping Coordination

All CMS platforms are queried by **slug**, not by internal ID. Each seed script assigns deterministic slugs to content items (e.g., `blog-post-0001` through `blog-post-10000`). The k6 `endpoints.js` module maps each CMS to its slug-based query pattern. This ensures every CMS is asked for "the same content" without maintaining a cross-CMS ID mapping table.

---

## Test Data Specification

Every CMS gets seeded with identical logical content:

### Data Shape

| Entity | Count | Purpose |
|--------|-------|---------|
| Datatypes | 5 | Blog Post, Page, Product, Author, Category |
| Fields per datatype | 8-12 | Text, rich text, number, boolean, date, image ref, relation |
| Content items (Blog Post) | 10,000 | Collection query benchmarks (must be large enough for serialization cost to appear) |
| Content items (Page) | 50 | Slug-based delivery benchmarks |
| Content tree depth | 4 levels | Nested tree benchmarks (Page > Section > Block > Element) |
| Content items (Product) | 5,000 | Secondary collection benchmarks |
| Content items (Author) | 100 | Relation resolution benchmarks |
| Content items (Category) | 20 | Relation resolution benchmarks |
| Media items | 100 | Reference fields (not benchmarking upload/download) |

### Reference Topology (Nested Tree Scenario)

The nested tree test data has a fixed, documented topology:

```
Page (root)
├── Section A
│   ├── Block 1 (2 _reference fields → Author, Category)
│   ├── Block 2 (1 _reference field → Product)
│   └── Block 3 (0 references)
├── Section B
│   ├── Block 4 (1 _reference field → Author)
│   └── Block 5 (2 _reference fields → Category, Product)
└── Section C
    └── Block 6 (1 _reference field → Author)
```

- **4 levels deep:** Page > Section > Block > Element (Blocks contain inline Elements)
- **7 composed reference lookups** per page request (each `_reference` field triggers a `GetPublishedSnapshot` call)
- **Total nodes per tree:** 10 (1 root + 3 sections + 6 blocks)
- This topology is intentionally modest — it represents a real-world page, not a stress test. The concurrency ramp scenario tests throughput at scale.

All 50 Page items share this exact topology so any page slug produces the same structural workload.

### Seed Requirements

- Every CMS must have the same number of items with the same field structure
- Relations between entities must be equivalent (e.g., Blog Post references Author)
- Seeding happens before benchmarks start, not during
- Seed scripts must be idempotent (run twice = same state)
- **Seed order:** Categories and Authors first (no dependencies), then Products, then Blog Posts and Pages (which reference the others)
- **ModulaCMS publish step:** After creating all content, the seed script must publish every content item. The delivery path serves from published snapshots, not live data. Publish time is recorded separately and reported in results as "write-time cost."
- **Deterministic slugs:** All items use predictable slugs (`blog-post-0001`, `page-0001`, `product-0001`) for cross-CMS endpoint mapping

### Data Equivalence Verification

Before running benchmarks, a verification step confirms that seeded data produces equivalent responses:

1. **Field count check:** Request the same item from each CMS, count the number of fields in the response. Must match within +/- 2 (metadata fields like `createdAt` may differ).
2. **Relation resolution check:** Request an item with relations, confirm referenced entities are populated (not just IDs).
3. **Nested depth check:** Request the nested tree item, confirm the response contains 4 levels of nesting with populated children.
4. **Collection count check:** Request the first page of blog posts, confirm 20 items returned with expected fields.
5. **Payload size recording:** Record response byte size for each scenario from each CMS. Include in results — size differences are part of the comparison (some CMS return more metadata).

The verification script outputs a pass/fail report. Benchmarks do not run if verification fails.

---

## Environment Constraints

### Hardware Isolation

All CMS containers run with identical resource limits:

```yaml
deploy:
  resources:
    limits:
      cpus: "2.0"
      memory: "2G"
    reservations:
      cpus: "1.0"
      memory: "1G"
```

### Database Isolation

Each CMS gets its **own PostgreSQL 17 container** with identical resource limits:

```yaml
# Per-CMS PostgreSQL container
deploy:
  resources:
    limits:
      cpus: "1.0"
      memory: "1G"
    reservations:
      cpus: "0.5"
      memory: "512M"
```

Rationale: A shared PostgreSQL instance allows cross-contamination through shared buffer cache, connection pools, and I/O scheduling. Separate containers ensure one CMS's database behavior cannot affect another's measurements.

### Connection Pool Normalization

Each CMS's PostgreSQL connection pool is configured to the same size:

| Setting | Value | Notes |
|---------|-------|-------|
| Max connections | 20 | Normalized across all CMS platforms |
| Min connections | 5 | Prevents cold-pool overhead differences |
| Connection timeout | 5s | Consistent timeout behavior |

Per-CMS configuration:
- **ModulaCMS:** `db_max_open_conns` / `db_max_idle_conns` in `modula.config.json`
- **Strapi:** `database.pool.min` / `database.pool.max` in `database.js`
- **Directus:** `DB_POOL_MIN` / `DB_POOL_MAX` env vars
- **Payload:** `pool.min` / `pool.max` in database adapter config

If a CMS does not support explicit pool configuration, document the default and note it in results.

### Other Constraints

- **No Redis/caching layers** — measure the CMS, not the cache
- **No CDN/reverse proxy** — direct container access
- **Host network disabled** — use Docker bridge networking
- **Production mode only** — all CMS must run in production mode (not dev/debug)
- **cAdvisor sidecar** — runs alongside all CMS containers for resource monitoring

### Runtime Rules

- 10-second warm-up period (requests sent, results discarded)
- 5 runs per scenario, report median with IQR (interquartile range) across runs
- Readiness-based cool-down between scenarios: wait until CMS container CPU drops below 5% for 5 consecutive seconds (replaces arbitrary 30-second timer)
- Health check must pass before starting each scenario
- No other workloads on the benchmark host during runs

### Cold Start Definition

"Cold start" is measured as: container `docker start` timestamp to first **successful content query response** (not health check). Specifically:

1. Start the CMS container
2. Poll `GET /api/v1/content/page-0001` (or equivalent) every 100ms
3. Record the timestamp of the first HTTP 200 with a valid JSON body
4. Cold start = (first successful content response timestamp) - (container start timestamp)

This ensures the measurement includes migrations, bootstrap data seeding, and server initialization — not just "process alive."

---

## k6 Test Scenarios

### Scenario 1: Single Item Delivery

```
Stages: 10s ramp to 50 VUs → 60s hold → 10s ramp down
Endpoint: Single page by slug (rotates through 50 page slugs)
Metrics: p50, p95, p99 latency; req/s; IQR across 5 runs
```

### Scenario 2: Collection Query

```
Stages: 10s ramp to 50 VUs → 60s hold → 10s ramp down
Endpoint: Paginated collection (20 items, sorted by created_at)
Metrics: p50, p95, p99 latency; req/s; IQR across 5 runs
```

### Scenario 3: Nested Tree

```
Stages: 10s ramp to 50 VUs → 60s hold → 10s ramp down
Endpoint: 4-level deep page tree with 7 composed references (see topology above)
Metrics: p50, p95, p99 latency; req/s; IQR across 5 runs
```

### Scenario 4: Concurrency Ramp

```
Stages: 30s ramp to 10 VUs → 30s hold → 30s ramp to 50 → 30s hold → 30s ramp to 100 → 30s hold → 30s ramp to 200 → 30s hold → 30s ramp to 500 → 60s hold → 30s ramp down
Endpoint: Mixed (70% single item, 20% collection, 10% nested)
Purpose: Find the breaking point — where does p99 spike, where do errors appear
```

### Scenario 5: Sustained Load

```
Duration: 5 minutes at 100 VUs
Endpoint: Mixed workload
Purpose: Memory leak detection, GC pressure, connection pool exhaustion
```

### Results Format

All scenarios output raw k6 JSON summary files (`--summary-export`). These are committed to `results/` for reproducibility. Anyone can re-analyze the raw data independently.

---

## Repository Structure

```
modulacms-benchmarks/
  README.md
  justfile                              # Orchestration commands
  docker-compose.infra.yml              # cAdvisor + shared network

  cms/
    modulacms/
      docker-compose.yml                # CMS + its own PostgreSQL + resource limits
      seed.sh                           # Seed via ModulaCMS REST API + publish all content
      config.json                       # Connection pool, port config
    strapi/
      Dockerfile                        # Strapi with PostgreSQL config
      docker-compose.yml                # CMS + its own PostgreSQL + resource limits
      seed.js                           # Schema creation (admin API) + data seeding (content API)
      database.js                       # Connection pool config
    directus/
      docker-compose.yml                # CMS + its own PostgreSQL + resource limits
      seed.js                           # Schema creation + data seeding via Directus API
    payload/
      Dockerfile                        # Payload with seed script baked in
      docker-compose.yml                # CMS + its own PostgreSQL + resource limits
      seed.ts                           # Seed via Payload REST API (not Local API — see note)

  k6/
    lib/
      config.js                         # Shared constants (VU counts, durations, slug lists)
      endpoints.js                      # Per-CMS slug-based endpoint mapping
      helpers.js                        # Response validation, metric tagging
    scenarios/
      single-item.js
      collection-query.js
      nested-tree.js
      concurrency-ramp.js
      sustained-load.js

  scripts/
    run.sh                              # Full benchmark orchestration
    run-single.sh                       # Benchmark one CMS
    seed.sh                             # Seed one CMS (dispatches to cms/{name}/seed.*)
    verify.sh                           # Data equivalence verification (runs before benchmarks)
    cold-start.sh                       # Cold start measurement
    collect-metrics.sh                  # cAdvisor metric export
    compare.sh                          # Generate comparison table from results

  results/
    .gitkeep                            # Committed baseline results go here
    README.md                           # How to interpret results, methodology disclosure

  analysis/
    charts.py                           # Generate comparison charts (matplotlib)
    templates/
      report.md.tmpl                    # Markdown report template
```

### Payload CMS Seeding

Payload's Local API requires importing the config and running in the same Node.js process, which means running inside the container. To keep seeding consistent across all CMS platforms (external script hitting REST API), use **Payload's REST API** for seeding instead. The seed script:

1. Authenticates via `POST /api/users/login`
2. Creates collections via Payload's admin config (must be defined in code, not API — bake into Dockerfile)
3. Seeds content via `POST /api/{collection}` REST endpoints

Collection schemas are defined in the Payload config file baked into the Docker image. Only data seeding happens at runtime via REST.

---

## Phases

### Phase 1 — Infrastructure & ModulaCMS Baseline

**Scope:** Get ModulaCMS benchmarking itself in isolation. No competitors yet.

- Set up repo with Docker Compose (PostgreSQL container + ModulaCMS container + cAdvisor)
- Write seed script using ModulaCMS REST API:
  - Create datatypes and fields (schema setup)
  - Create content items in dependency order (Categories/Authors → Products → Blog Posts/Pages)
  - Publish all content items (delivery serves from snapshots)
  - Record publish timing
- Write data equivalence verification script (field counts, relation checks, nesting depth)
- Write k6 scenarios for all 5 test types against ModulaCMS endpoints
- Write `run-single.sh` orchestration (start, seed, verify, warm up, benchmark, collect cAdvisor metrics, teardown)
- Write cold start measurement script
- Establish baseline numbers with IQR across 5 runs
- Document hardware spec and how to reproduce

**Deliverable:** `just bench modulacms` runs the full suite and outputs JSON results + cAdvisor metrics.

**Size:** 3-4 sessions. (Seed script with schema creation, content creation in dependency order, and publish step is a full session. k6 scenarios with response validation and mixed-workload weighting are another. Orchestration and debugging are a third.)

### Phase 2 — Add Strapi

**Scope:** First competitor comparison. Strapi is the most popular and will surface any methodology problems.

- Write Strapi Dockerfile with PostgreSQL config, connection pool settings, and production mode
- Write seed script:
  - Create content types via Strapi admin API (schema definition)
  - Seed data via Strapi content API (same item counts, same field values)
  - Assign deterministic slugs to all items
- Map ModulaCMS k6 scenarios to Strapi slug-based endpoints in `endpoints.js`
- Run data equivalence verification between ModulaCMS and Strapi responses
- Run both benchmarks on same hardware, same data shape
- Compare results with IQR — check if differences exceed variance
- Identify any methodology bias and adjust if needed

**Deliverable:** Side-by-side ModulaCMS vs Strapi numbers for all 5 scenarios with statistical context.

**Size:** 2-3 sessions. (Strapi schema creation via admin API is fiddly. The first competitor also requires tuning the verification and comparison scripts.)

### Phase 3 — Add Remaining Competitors

**Scope:** Directus and Payload CMS. These follow the pattern established in Phase 2.

- Per-CMS: Dockerfile, docker-compose.yml (with dedicated PostgreSQL), seed script, endpoint mapping
- Payload: Collection schemas baked into Dockerfile, data seeding via REST API
- Directus: Schema creation via Directus API, data seeding via items API
- Run data equivalence verification for each new CMS
- Run all 4 CMS platforms through the full suite

**Deliverable:** Full 4-way comparison table.

**Size:** 2-3 sessions (parallelizable — one CMS per agent).

### Phase 4 — Analysis & Reporting

**Scope:** Turn raw numbers into publishable results.

- Build comparison charts (latency distributions with error bars, throughput bars, resource usage time-series)
- Write methodology disclosure (snapshot architecture, output format choice, connection pool normalization)
- Write analysis highlighting where ModulaCMS wins and where it doesn't
- Include publish-time cost for ModulaCMS alongside read-time results
- Generate reproducibility guide (exact commands, hardware specs, Docker versions, image tags)
- Create a summary table suitable for README/marketing

**Deliverable:** `results/` directory with charts, tables, raw k6 JSON, and narrative analysis.

**Size:** 1-2 sessions.

### Phase 5 — CI Integration (Optional)

**Scope:** Run ModulaCMS-only benchmarks on every release to detect regressions.

- GitHub Actions workflow triggered on release tags
- Runs Phase 1 suite (ModulaCMS only, no competitors)
- Stores results as release artifacts
- Compares against previous release baseline
- Alerts if p95 latency increases >10% or throughput drops >10%

**Deliverable:** `.github/workflows/benchmark.yml` in the benchmark repo.

**Size:** 1 session.

---

## Fairness Checklist

Before publishing any results, verify:

- [ ] All CMS containers have identical CPU/memory limits
- [ ] Each CMS has its own PostgreSQL container with identical resource limits
- [ ] Connection pool sizes are normalized across all CMS platforms (or defaults are documented)
- [ ] All seed data has the same number of items, fields, and relations
- [ ] Data equivalence verification script passes for all CMS platforms
- [ ] ModulaCMS output format is documented and justified (`clean`)
- [ ] Snapshot architecture is disclosed in methodology
- [ ] ModulaCMS publish-time cost is reported separately
- [ ] Warm-up period is applied to all platforms
- [ ] Each scenario runs 5+ times with median and IQR reported
- [ ] No CMS has a caching layer the others don't (acknowledge ModulaCMS snapshots)
- [ ] All CMS run in production mode
- [ ] Container image versions are pinned (not `:latest`)
- [ ] Results include exact Docker image tags, k6 version, cAdvisor version, host hardware
- [ ] Response payload sizes are documented per CMS per scenario
- [ ] Raw k6 JSON output is committed for independent re-analysis

---

## Expected Advantages

Based on ModulaCMS architecture, expect strong results in:

1. **Nested tree delivery** — Go composes trees server-side from published snapshots. Competitors resolve relations and assemble trees from normalized rows at request time.
2. **Cold start** — Single Go binary vs Node.js boot + ORM initialization + migration checks.
3. **Memory under load** — Go's memory model vs V8 heap pressure.
4. **Tail latency (p99)** — Go's GC is designed for low-pause-time. Node.js event loop can stall under high allocation rates.
5. **Sustained load** — No GC pauses accumulating, no event loop saturation.

Expect parity or slight disadvantage in:

1. **Simple single-item GET** — Database is the bottleneck, all platforms are equivalent.
2. **Collection queries** — Database-bound. ORM overhead in Node.js might show slightly, but margins will be small at low item counts. At 10K items, serialization cost should become visible.

---

## Anti-Patterns to Avoid

- **Don't cherry-pick scenarios.** Run the full suite, publish the full suite.
- **Don't compare SQLite ModulaCMS vs PostgreSQL competitors.** Use PostgreSQL for all.
- **Don't disable competitor features.** If Strapi does schema validation on reads, that's part of its cost.
- **Don't over-tune ModulaCMS.** Run with default config. If tuning is done, document it and offer equivalent tuning for competitors.
- **Don't benchmark dev mode.** All CMS must run in production mode.
- **Don't claim results without variance data.** Overlapping IQRs mean the difference is not significant.
- **Don't hide the snapshot advantage.** Disclose it in methodology. Let readers evaluate whether write-time work for read-time speed is a fair trade.
- **Don't publish until methodology is peer-reviewed.** Have someone reproduce the results independently.
