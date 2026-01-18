# ModulaCMS as a Distributed System

**Created:** 2026-01-16
**Purpose:** Analysis of ModulaCMS's distributed system capabilities and architecture patterns

---

## TL;DR: Yes, ModulaCMS Can Be Distributed

**Short answer:** Yes, ModulaCMS is designed to work as a distributed system.

**Requirements:**
- ✅ Use PostgreSQL or MySQL (not SQLite)
- ✅ Load balancer in front
- ✅ Database replication (optional but recommended)
- ✅ S3 for media (already distributed)
- ✅ Shared session storage (database or Redis)

**Architecture:**
```
┌─────────────────────────────────────────┐
│         Load Balancer                   │
│     (HAProxy, Nginx, AWS ALB)           │
└─────────┬───────────────────────────────┘
          │
    ┌─────┴─────┬─────────┬─────────┐
    │           │         │         │
┌───▼───┐  ┌───▼───┐ ┌───▼───┐ ┌───▼───┐
│ModCMS │  │ModCMS │ │ModCMS │ │ModCMS │
│ (1)   │  │ (2)   │ │ (3)   │ │ (4)   │
└───┬───┘  └───┬───┘ └───┬───┘ └───┬───┘
    │          │         │         │
    └──────────┴────┬────┴─────────┘
                    │
            ┌───────▼────────┐
            │  PostgreSQL    │
            │  (replicated)  │
            └────────────────┘

            ┌────────────────┐
            │   S3 Storage   │
            │    (media)     │
            └────────────────┘
```

---

## ModulaCMS Distributed System Properties

### 1. Stateless Application Servers ✅

**Why it works:**

```go
// ModulaCMS HTTP server is stateless
// No in-memory state required
// Each request is independent

func main() {
    config := loadConfig()
    router := setupRouter(config)

    // Just serve HTTP
    // No shared memory
    // No local state
    http.ListenAndServe(":8080", router)
}
```

**Key characteristics:**
- ✅ No in-memory sessions (stored in database)
- ✅ No in-memory cache (optional Redis)
- ✅ No file uploads to local disk (goes to S3)
- ✅ No local state dependencies
- ✅ Each instance can handle any request

**Result:** Horizontal scaling works! Just add more instances.

---

### 2. Shared Database State ✅

**Database options for distribution:**

#### SQLite ❌ (NOT distributed)
```
Problem: Single file database
- Can't be shared across servers
- File locking issues
- Network file systems too slow

Use case: Single server only
Not suitable for: Distributed systems
```

#### MySQL ✅ (Distributed-capable)
```
Setup: Master-slave replication
┌─────────────┐
│MySQL Master │ ← Write requests
└──────┬──────┘
       │ Replication
   ┌───┴───┬────────┐
   │       │        │
┌──▼──┐ ┌──▼──┐ ┌──▼──┐
│Slave│ │Slave│ │Slave│ ← Read requests
└─────┘ └─────┘ └─────┘

Benefits:
- Read scaling (multiple slaves)
- Geographic distribution
- Failover capability

Limitations:
- Single write master (bottleneck)
- Replication lag (eventual consistency)
```

#### PostgreSQL ✅ (Best for distribution)
```
Setup: Streaming replication
┌──────────────┐
│Postgres      │ ← Write requests
│Primary       │
└──────┬───────┘
       │ Streaming replication
   ┌───┴───┬────────┐
   │       │        │
┌──▼──┐ ┌──▼──┐ ┌──▼──┐
│Read │ │Read │ │Read │ ← Read requests
│Replica││Replica││Replica│
└─────┘ └─────┘ └─────┘

Benefits:
- Fast replication (streaming)
- Read scaling
- Automatic failover (with Patroni)
- Geographic distribution

Advanced: Multi-region
┌─────────────┐    ┌─────────────┐
│US-East      │◄──►│EU-West      │
│Primary      │    │Replica      │
└─────────────┘    └─────────────┘
```

---

### 3. Session Management ✅

**Current implementation:**
```sql
-- Sessions stored in database
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    session_data TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Why this works for distribution:**
```
Request 1 → Server A → Creates session in DB
Request 2 → Server B → Reads same session from DB
Request 3 → Server C → Reads same session from DB

All servers share session state (database)
No sticky sessions required
Load balancer can route anywhere
```

**Performance consideration:**
```
Problem: Database query on every request

Solutions:
1. Cache sessions in Redis (L1 cache)
   Request → Check Redis → If miss, check DB → Cache in Redis

2. Short-lived session tokens (reduce DB hits)
   JWT tokens (stateless) + DB validation (periodic)

3. Session replication
   Write-through cache (Redis) + DB backup
```

---

### 4. Media Storage (S3) ✅ Already Distributed

**Current implementation:**
```go
// Media already goes to S3
// S3 is distributed by design
// CDN-ready (CloudFront, Cloudflare)

upload → ModulaCMS (any instance) → S3
retrieve → CDN → S3
```

**Geographic distribution:**
```
         Cloudflare CDN (Global)
                  ↓
              S3 Bucket
                  ↓
         ┌────────┼────────┐
         │        │        │
     US-East  EU-West  Asia-Pacific
```

**Why it works:**
- ✅ No local file storage
- ✅ Any instance can upload
- ✅ Any instance can reference
- ✅ CDN handles distribution
- ✅ Globally accessible

---

### 5. Database Transactions and Tree Integrity ⚠️

**Challenge: Tree structure with concurrent updates**

```sql
-- Tree uses sibling pointers
CREATE TABLE content_data (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    first_child_id INTEGER,
    next_sibling_id INTEGER,
    prev_sibling_id INTEGER,
    -- ...
);
```

**Potential race conditions:**

#### Scenario 1: Concurrent sibling insertion
```
Server A: Insert node as first child
Server B: Insert node as first child (same parent)

Without proper locking:
→ Both read parent.first_child_id = NULL
→ Both insert as first child
→ Broken tree! (parent has two first children)
```

#### Solution: Database transactions
```go
func InsertAsFirstChild(parentID, newNodeID int) error {
    tx := db.Begin()
    defer tx.Rollback()

    // Lock parent row (FOR UPDATE)
    parent := tx.QueryRow(`
        SELECT first_child_id FROM content_data
        WHERE id = ?
        FOR UPDATE
    `, parentID)

    oldFirstChild := parent.FirstChildID

    // Update new node
    tx.Exec(`
        UPDATE content_data
        SET next_sibling_id = ?
        WHERE id = ?
    `, oldFirstChild, newNodeID)

    // Update parent
    tx.Exec(`
        UPDATE content_data
        SET first_child_id = ?
        WHERE id = ?
    `, newNodeID, parentID)

    return tx.Commit()
}
```

**Database guarantees:**
- ✅ PostgreSQL: Row-level locking (FOR UPDATE)
- ✅ MySQL: Row-level locking (InnoDB)
- ✅ ACID transactions
- ✅ Isolation levels (READ COMMITTED, SERIALIZABLE)

**Result:** Tree integrity maintained even with concurrent writes.

---

### 6. Cache Coherency Across Instances

**Challenge: Cache invalidation**

```
Server A: Updates content (ID 123)
Server A: Invalidates cache (ID 123)
Server B: Still has old cached content (ID 123) ← Problem!
```

**Solution 1: Database-backed cache (no coherency issue)**
```go
// Cache stored in database or Redis
// All servers read from same cache
// Invalidation affects all servers

cache.Set("content:123", data)  // Writes to Redis
cache.Delete("content:123")     // Deletes from Redis (all servers see it)
```

**Solution 2: Cache invalidation pubsub**
```go
// Server A updates content
db.Update(content)

// Server A publishes invalidation message
redis.Publish("cache:invalidate", "content:123")

// All servers subscribe
redis.Subscribe("cache:invalidate", func(msg) {
    localCache.Delete(msg) // Each server invalidates local cache
})
```

**Solution 3: Time-based expiration (eventual consistency)**
```go
// Cache with short TTL
cache.Set("content:123", data, 60*time.Second) // 60 second TTL

// Stale cache for max 60 seconds (acceptable for most use cases)
```

**Recommendation:**
- For content (reads >> writes): Short TTL cache (30-60s)
- For critical data (user sessions): Redis shared cache
- For high-traffic: Two-tier cache (local + Redis)

---

### 7. Load Balancing Strategies

#### Strategy 1: Round Robin (Simplest)
```
Request 1 → Server A
Request 2 → Server B
Request 3 → Server C
Request 4 → Server A
...
```

**Works for ModulaCMS:** ✅ (stateless servers, sessions in DB)

---

#### Strategy 2: Least Connections
```
Server A: 10 active connections
Server B: 5 active connections  ← Send here
Server C: 15 active connections

Next request → Server B
```

**Better for:** Long-running requests, WebSocket (if added)

---

#### Strategy 3: IP Hash (Sticky sessions)
```
User from IP 1.2.3.4 → Always Server A
User from IP 5.6.7.8 → Always Server B
```

**Not needed for ModulaCMS** (sessions in DB, no local state)

---

#### Strategy 4: Geographic routing
```
User in US → US-East cluster
User in EU → EU-West cluster
User in Asia → Asia-Pacific cluster
```

**Best for:** Global distribution, low latency

---

### 8. Deployment Architectures

#### Architecture 1: Single Region (HA)
```
┌─────────────────────────────────────┐
│         Load Balancer               │
│      (HAProxy / Nginx)              │
└──────────┬──────────────────────────┘
           │
    ┌──────┴──────┬─────────┐
    │             │         │
┌───▼───┐    ┌───▼───┐ ┌───▼───┐
│ModCMS │    │ModCMS │ │ModCMS │
│  (1)  │    │  (2)  │ │  (3)  │
└───┬───┘    └───┬───┘ └───┬───┘
    │            │         │
    └────────────┴────┬────┘
                      │
              ┌───────▼────────┐
              │  PostgreSQL    │
              │  Primary       │
              └───────┬────────┘
                      │
              ┌───────▼────────┐
              │  PostgreSQL    │
              │  Standby       │
              └────────────────┘

Benefits:
- High availability (3+ servers)
- Database failover (primary → standby)
- Simple to manage
- Low latency (single region)

Use case: Medium traffic (1k-100k req/day)
```

---

#### Architecture 2: Multi-Region (Global)
```
┌──────────────────────────────────────────────┐
│         Cloudflare / AWS Global Accelerator │
│            (Global Load Balancer)            │
└───────┬──────────────────────┬───────────────┘
        │                      │
    ┌───▼───────┐         ┌───▼───────┐
    │  US-East  │         │  EU-West  │
    │  Region   │         │  Region   │
    └───┬───────┘         └───┬───────┘
        │                      │
┌───────▼────────┐     ┌───────▼────────┐
│ Load Balancer  │     │ Load Balancer  │
└───────┬────────┘     └───────┬────────┘
        │                      │
  ┌─────┴─────┐          ┌─────┴─────┐
┌─▼─┐ ┌─▼─┐ ┌─▼─┐      ┌─▼─┐ ┌─▼─┐ ┌─▼─┐
│MC │ │MC │ │MC │      │MC │ │MC │ │MC │
│(1)│ │(2)│ │(3)│      │(4)│ │(5)│ │(6)│
└─┬─┘ └─┬─┘ └─┬─┘      └─┬─┘ └─┬─┘ └─┬─┘
  │     │     │            │     │     │
  └─────┴──┬──┘            └─────┴──┬──┘
           │                        │
    ┌──────▼─────┐           ┌──────▼─────┐
    │PostgreSQL  │◄─────────►│PostgreSQL  │
    │ Primary    │  Streaming│ Replica    │
    │ (US-East)  │  Repl.    │ (EU-West)  │
    └────────────┘           └────────────┘

Benefits:
- Global low latency
- Geographic redundancy
- Read scaling per region
- Disaster recovery

Use case: Global traffic (100k+ req/day)
```

---

#### Architecture 3: Active-Active Multi-Master (Advanced)
```
┌──────────────────────────────────────────────┐
│              Global Load Balancer            │
└───────┬──────────────────────┬───────────────┘
        │                      │
    ┌───▼───────┐         ┌───▼───────┐
    │  US-East  │         │  EU-West  │
    └───┬───────┘         └───┬───────┘
        │                      │
┌───────▼────────┐     ┌───────▼────────┐
│ModulaCMS       │     │ModulaCMS       │
│Cluster (1-3)   │     │Cluster (4-6)   │
└───────┬────────┘     └───────┬────────┘
        │                      │
┌───────▼────────┐     ┌───────▼────────┐
│PostgreSQL      │◄───►│PostgreSQL      │
│Multi-Master    │  ↕  │Multi-Master    │
│(US-East)       │     │(EU-West)       │
└────────────────┘     └────────────────┘
  Bi-directional replication

Tools: PostgreSQL BDR, Citus, CockroachDB

Benefits:
- Write scaling (multiple write regions)
- Zero RPO (no data loss)
- Active-active (both regions serve traffic)

Challenges:
- Conflict resolution (concurrent updates)
- Complex setup
- Expensive

Use case: Mission-critical, global (1M+ req/day)
```

---

## Distributed System Challenges & Solutions

### Challenge 1: CAP Theorem

**CAP Theorem:** Can only have 2 of 3:
- **C**onsistency (all nodes see same data)
- **A**vailability (system always responds)
- **P**artition tolerance (works despite network splits)

**ModulaCMS choice: CP (Consistency + Partition Tolerance)**

```
Why CP?
- Content must be consistent (can't have two versions)
- Partition tolerance required (multi-region)
- Trade-off: Availability (if DB down, system down)

Alternative (AP):
- High availability
- Eventual consistency
- Risk: Users see stale content
```

**Implementation:**
```go
// Strong consistency (PostgreSQL SERIALIZABLE)
tx := db.Begin(sql.LevelSerializable)

// vs Eventual consistency (async replication)
tx := db.Begin(sql.LevelReadCommitted)
// Accept replication lag (stale reads possible)
```

---

### Challenge 2: Database Replication Lag

**Problem:**
```
Primary DB: Content updated (v2)
Replica DB: Still has old content (v1) ← 100ms lag

User reads from replica: Sees old content!
```

**Solutions:**

#### Solution 1: Read-your-writes consistency
```go
// After write, read from primary (not replica)
db.Update(content) // Write to primary

// Read from primary (not replica)
content := db.QueryPrimary("SELECT * FROM content_data WHERE id = ?", id)
```

#### Solution 2: Session affinity to primary
```go
// User who just wrote → Route to primary for next N requests
if recentWrite(userID) {
    return primary.Query(...)
} else {
    return replica.Query(...)
}
```

#### Solution 3: Accept eventual consistency
```go
// Most content reads can tolerate 100ms lag
// Trade consistency for performance
content := replica.Query(...) // May be slightly stale
```

**Recommendation:**
- Critical data (user sessions): Read from primary
- Public content: Read from replicas (eventual consistency OK)
- Recent writes: Read-your-writes consistency

---

### Challenge 3: Distributed Transactions

**Problem:**
```
Update content in US-East DB
Update related content in EU-West DB
One succeeds, one fails → Inconsistent state!
```

**Solutions:**

#### Solution 1: Single database (recommended)
```
- Don't distribute writes across regions
- One primary database (writes)
- Replicas for reads only
- Simpler, more reliable
```

#### Solution 2: Two-phase commit (complex)
```
Coordinator: "Prepare to commit"
US-East DB: "Ready"
EU-West DB: "Ready"
Coordinator: "Commit!"
Both: Commit

Problem: Coordinator can fail (blocking)
```

#### Solution 3: Saga pattern
```
Step 1: Update US-East → Success
Step 2: Update EU-West → Fail
Compensation: Rollback US-East update

Eventually consistent
Requires compensation logic
```

**ModulaCMS recommendation:** Single primary database, replicas for reads.

---

### Challenge 4: Clock Synchronization

**Problem:**
```
Server A clock: 2024-01-16 10:00:00
Server B clock: 2024-01-16 09:59:55 (5 seconds behind)

Content created on A: created_at = 10:00:00
Content created on B: created_at = 09:59:55

Order by created_at → Wrong order!
```

**Solutions:**

#### Solution 1: NTP (Network Time Protocol)
```bash
# All servers sync with NTP
ntpd -q -g
# Keeps clocks within milliseconds
```

#### Solution 2: Database-generated timestamps
```sql
-- Don't use application time
-- Use database time (single source of truth)
CREATE TABLE content_data (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- DB time
);
```

#### Solution 3: Hybrid logical clocks
```
HLC = (physical_time, logical_counter)
Guarantees: later events have later HLC
Used by: CockroachDB, Spanner
```

**ModulaCMS recommendation:** Database-generated timestamps (simple, reliable).

---

## Deployment Examples

### Example 1: Kubernetes Deployment

```yaml
# modulacms-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: modulacms
spec:
  replicas: 3  # 3 instances
  selector:
    matchLabels:
      app: modulacms
  template:
    metadata:
      labels:
        app: modulacms
    spec:
      containers:
      - name: modulacms
        image: modulacms/modulacms:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: postgres-primary.default.svc.cluster.local
        - name: DB_NAME
          value: modulacms
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
# Service + Load Balancer
apiVersion: v1
kind: Service
metadata:
  name: modulacms
spec:
  type: LoadBalancer
  selector:
    app: modulacms
  ports:
  - port: 80
    targetPort: 8080
```

**Result:** 3 ModulaCMS instances behind Kubernetes load balancer.

---

### Example 2: Docker Compose (Development)

```yaml
# docker-compose.yml
version: '3.8'

services:
  modulacms-1:
    image: modulacms/modulacms:latest
    environment:
      DB_HOST: postgres
      DB_NAME: modulacms
      DB_USER: postgres
      DB_PASSWORD: password
    depends_on:
      - postgres
      - redis

  modulacms-2:
    image: modulacms/modulacms:latest
    environment:
      DB_HOST: postgres
      DB_NAME: modulacms
      DB_USER: postgres
      DB_PASSWORD: password
    depends_on:
      - postgres
      - redis

  modulacms-3:
    image: modulacms/modulacms:latest
    environment:
      DB_HOST: postgres
      DB_NAME: modulacms
      DB_USER: postgres
      DB_PASSWORD: password
    depends_on:
      - postgres
      - redis

  nginx:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - modulacms-1
      - modulacms-2
      - modulacms-3

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: modulacms
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7
    volumes:
      - redis-data:/data

volumes:
  postgres-data:
  redis-data:
```

```nginx
# nginx.conf
upstream modulacms {
    server modulacms-1:8080;
    server modulacms-2:8080;
    server modulacms-3:8080;
}

server {
    listen 80;

    location / {
        proxy_pass http://modulacms;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

### Example 3: AWS Multi-Region

```
CloudFormation / Terraform:

# US-East-1
- ALB (Application Load Balancer)
- ECS Fargate (3 tasks)
- RDS PostgreSQL (primary)
- ElastiCache Redis
- S3 (media)

# EU-West-1
- ALB
- ECS Fargate (3 tasks)
- RDS PostgreSQL (read replica)
- ElastiCache Redis
- S3 (media - replicated)

# Route 53 (global DNS)
- Geolocation routing
  - US traffic → US-East-1
  - EU traffic → EU-West-1
```

---

## Performance Benchmarks (Estimated)

### Single Server
```
Setup: 1 ModulaCMS instance, PostgreSQL
Capacity: 1,000 req/min (60k req/hour)
Latency: 50ms avg
```

### Load Balanced (3 servers)
```
Setup: 3 ModulaCMS instances, PostgreSQL
Capacity: 3,000 req/min (180k req/hour)
Latency: 50ms avg
```

### Multi-Region (6 servers)
```
Setup: 3 instances US + 3 instances EU, PostgreSQL replication
Capacity: 6,000 req/min (360k req/hour)
Latency: 20-30ms avg (geo-distributed)
```

### Enterprise (20+ servers)
```
Setup: Load balanced clusters, PostgreSQL multi-master
Capacity: 100,000+ req/min (6M+ req/hour)
Latency: <20ms avg
```

---

## Monitoring Distributed ModulaCMS

### Key Metrics

**Application:**
- Request rate (requests/sec per instance)
- Response time (p50, p95, p99)
- Error rate (4xx, 5xx)
- Active connections

**Database:**
- Query latency
- Connection pool usage
- Replication lag (replica behind primary)
- Transaction throughput

**Infrastructure:**
- CPU usage per instance
- Memory usage per instance
- Network I/O
- Disk I/O

### Tools

**Metrics:** Prometheus + Grafana
**Logging:** ELK stack (Elasticsearch, Logstash, Kibana)
**Tracing:** Jaeger, OpenTelemetry
**Alerting:** PagerDuty, OpsGenie

---

## Conclusion: ModulaCMS Distributed System Readiness

### ✅ Ready Today

1. **Stateless servers** - Horizontal scaling works
2. **Shared database** - PostgreSQL/MySQL replication
3. **S3 media** - Already distributed
4. **Session management** - Database-backed sessions
5. **Load balancing** - Works out of box
6. **Multi-region** - Deploy anywhere

### ⚠️ Needs Configuration

1. **Database replication** - Set up read replicas
2. **Cache layer** - Add Redis for performance
3. **Load balancer** - Configure HAProxy, Nginx, or cloud LB
4. **Monitoring** - Add Prometheus, Grafana

### 🔮 Future Enhancements

1. **Built-in cache invalidation** - Pubsub for multi-instance cache
2. **Read replica routing** - Automatic read/write splitting
3. **Multi-master support** - Active-active regions
4. **Distributed tracing** - OpenTelemetry integration

---

## Recommendation

**For most use cases:**
```
Start: Single server (1k-10k req/day)
↓
Grow: Load balanced 3 servers (10k-100k req/day)
↓
Scale: Multi-region (100k-1M req/day)
↓
Enterprise: Active-active multi-master (1M+ req/day)
```

**ModulaCMS can handle all of these stages with the same core binary.**

**The sophistication lies in how unsophisticated it is:**
- Simple: Stateless Go binary
- Powerful: Scales to enterprise
- Flexible: Deploy anywhere (single server → global)

---

**Last Updated:** 2026-01-16
