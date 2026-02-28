# Zero-Schema Feature Implementations

**Purpose:** Explore which "infrastructure" features can be implemented using datatypes instead of new packages
**Created:** 2026-01-16
**Key Insight:** Database-backed infrastructure instead of in-memory infrastructure

---

## Philosophy Shift

**Traditional approach:**
- Cache = Redis/in-memory
- Cron = Separate scheduler process
- Webhooks = Event dispatcher + queue

**ModulaCMS creative approach:**
- Cache = Datatype with expiration
- Cron = Datatype with schedule + background process
- Webhooks = Fetch request datatype + junction table

**Trade-offs:**
- Slightly slower (database I/O vs in-memory)
- But: Persistent, queryable, manageable via TUI
- And: Good enough for 90% of use cases

---

## Pure Datatype Implementations (Zero Schema Changes)

These use **only** existing datatypes/fields system - no new tables, no code changes.

---

### ✅ 1. Cache System (via Datatype)

**Implementation:**
```sql
-- Create Cache datatype
INSERT INTO datatypes (name, slug) VALUES ('Cache Entry', 'cache_entry');

-- Create cache fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Cache Key', 'key', 'text'),           -- Unique identifier
    ('Cache Value', 'value', 'json'),       -- Cached data (JSON)
    ('Expires At', 'expires_at', 'datetime'),
    ('Created At', 'created_at', 'datetime'),
    ('Hit Count', 'hit_count', 'number'),   -- Stats
    ('Tags', 'tags', 'text');               -- Cache tags for invalidation

-- Create cache route
INSERT INTO routes (name, slug, datatype_id)
VALUES ('Cache', 'cache', <cache_entry_datatype_id>);
```

**Usage in code:**
```go
package cache

import (
    "encoding/json"
    "time"
    "github.com/hegner123/modulacms/internal/db"
)

type Cache struct {
    db db.DbDriver
}

// Get retrieves cached value if not expired
func (c *Cache) Get(key string) (interface{}, bool) {
    // Query cache route for entry with matching key
    entries := c.db.GetContentByField("cache", "key", key)
    if len(entries) == 0 {
        return nil, false
    }

    entry := entries[0]

    // Check expiration
    expiresAt := entry.GetField("expires_at")
    if time.Now().After(expiresAt) {
        // Expired - delete entry
        c.db.DeleteContent(entry.ID)
        return nil, false
    }

    // Increment hit count
    hitCount := entry.GetField("hit_count").(int)
    entry.SetField("hit_count", hitCount + 1)
    c.db.UpdateContent(entry)

    // Return cached value
    value := entry.GetField("value")
    return value, true
}

// Set stores value in cache with TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
    // Check if key exists (update vs create)
    existing := c.db.GetContentByField("cache", "key", key)

    expiresAt := time.Now().Add(ttl)
    valueJSON, _ := json.Marshal(value)

    if len(existing) > 0 {
        // Update existing
        entry := existing[0]
        entry.SetField("value", string(valueJSON))
        entry.SetField("expires_at", expiresAt)
        return c.db.UpdateContent(entry)
    }

    // Create new cache entry
    return c.db.CreateContent(&db.ContentData{
        RouteID: <cache_route_id>,
        DatatypeID: <cache_datatype_id>,
        Fields: map[string]interface{}{
            "key": key,
            "value": string(valueJSON),
            "expires_at": expiresAt,
            "created_at": time.Now(),
            "hit_count": 0,
        },
    })
}

// Delete removes cache entry
func (c *Cache) Delete(key string) error {
    entries := c.db.GetContentByField("cache", "key", key)
    for _, entry := range entries {
        c.db.DeleteContent(entry.ID)
    }
    return nil
}

// InvalidateByTags removes all cache entries with matching tags
func (c *Cache) InvalidateByTags(tags []string) error {
    for _, tag := range tags {
        entries := c.db.GetContentByField("cache", "tags", "%"+tag+"%") // LIKE query
        for _, entry := range entries {
            c.db.DeleteContent(entry.ID)
        }
    }
    return nil
}

// CleanupExpired removes expired entries (run periodically)
func (c *Cache) CleanupExpired() error {
    now := time.Now()
    // Query all cache entries where expires_at < now
    allCache := c.db.GetContentByRoute("cache")

    for _, entry := range allCache {
        expiresAt := entry.GetField("expires_at").(time.Time)
        if expiresAt.Before(now) {
            c.db.DeleteContent(entry.ID)
        }
    }
    return nil
}
```

**Use cases:**
- API response caching
- Expensive query results
- Plugin output caching
- External API responses (rate limiting)
- Computed aggregations

**Advantages:**
- Persistent across restarts (unlike in-memory)
- Queryable (see all cache entries in TUI)
- Manageable (clear cache via TUI/API)
- Analytics (hit count, expiration stats)
- Tags for grouped invalidation

**Performance:**
- Database I/O on every get/set
- For small sites (<1000 req/min): perfectly fine
- For high traffic: add in-memory layer on top (two-tier cache)

**Two-tier cache:**
```go
type TwoTierCache struct {
    memory map[string]CacheEntry  // L1: In-memory
    db     *Cache                  // L2: Database
}

func (c *TwoTierCache) Get(key string) (interface{}, bool) {
    // Check memory first
    if entry, ok := c.memory[key]; ok && !entry.Expired() {
        return entry.Value, true
    }

    // Check database
    if value, ok := c.db.Get(key); ok {
        // Store in memory for next time
        c.memory[key] = CacheEntry{Value: value, ExpiresAt: time.Now().Add(5*time.Minute)}
        return value, true
    }

    return nil, false
}
```

---

### ✅ 2. Scheduled Tasks / Cron Jobs (via Datatype)

**Implementation:**
```sql
-- Create Cron Job datatype
INSERT INTO datatypes (name, slug) VALUES ('Cron Job', 'cron_job');

-- Create cron job fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Job Name', 'name', 'text'),
    ('Description', 'description', 'textarea'),
    ('Schedule', 'schedule', 'text'),         -- Cron expression: "*/5 * * * *"
    ('Action Type', 'action_type', 'select'), -- predefined / plugin / http
    ('Action Config', 'action_config', 'json'), -- Action-specific config
    ('Enabled', 'enabled', 'boolean'),
    ('Last Run', 'last_run', 'datetime'),
    ('Next Run', 'next_run', 'datetime'),
    ('Run Count', 'run_count', 'number'),
    ('Last Status', 'last_status', 'select'); -- success / error / running

-- Create cron jobs route
INSERT INTO routes (name, slug, datatype_id)
VALUES ('Cron Jobs', 'cron-jobs', <cron_job_datatype_id>);
```

**Cron job examples:**
```json
// 1. Publish scheduled content (predefined action)
{
  "name": "Publish Scheduled Content",
  "schedule": "* * * * *",  // Every minute
  "action_type": "predefined",
  "action_config": {
    "action": "publish_scheduled_content"
  },
  "enabled": true
}

// 2. Cleanup old sessions (predefined action)
{
  "name": "Cleanup Expired Sessions",
  "schedule": "0 * * * *",  // Every hour
  "action_type": "predefined",
  "action_config": {
    "action": "cleanup_sessions",
    "older_than_days": 30
  },
  "enabled": true
}

// 3. Run Lua plugin (plugin action)
{
  "name": "Generate Daily Report",
  "schedule": "0 0 * * *",  // Daily at midnight
  "action_type": "plugin",
  "action_config": {
    "plugin_name": "reports",
    "function": "generateDailyReport"
  },
  "enabled": true
}

// 4. HTTP webhook (http action)
{
  "name": "Notify External Service",
  "schedule": "0 */6 * * *",  // Every 6 hours
  "action_type": "http",
  "action_config": {
    "url": "https://api.example.com/webhook",
    "method": "POST",
    "body": {"event": "scheduled_ping"}
  },
  "enabled": true
}

// 5. Backup database (predefined action)
{
  "name": "Daily Backup",
  "schedule": "0 2 * * *",  // 2 AM daily
  "action_type": "predefined",
  "action_config": {
    "action": "backup_database"
  },
  "enabled": true
}
```

**Scheduler implementation:**
```go
package scheduler

import (
    "time"
    "github.com/robfig/cron/v3"
    "github.com/hegner123/modulacms/internal/db"
)

type Scheduler struct {
    db     db.DbDriver
    cron   *cron.Cron
    jobs   map[int]*cron.Entry  // Job ID -> cron entry
}

// Start begins the scheduler
func (s *Scheduler) Start() {
    s.cron = cron.New()
    s.jobs = make(map[int]*cron.Entry)

    // Load all enabled cron jobs from database
    s.LoadJobs()

    // Start cron scheduler
    s.cron.Start()

    // Watch for job changes (poll every minute)
    go s.WatchForChanges()
}

// LoadJobs loads all cron jobs from database
func (s *Scheduler) LoadJobs() {
    cronJobs := s.db.GetContentByRoute("cron-jobs")

    for _, job := range cronJobs {
        enabled := job.GetField("enabled").(bool)
        if !enabled {
            continue
        }

        schedule := job.GetField("schedule").(string)

        // Add to cron scheduler
        entryID, err := s.cron.AddFunc(schedule, func() {
            s.ExecuteJob(job)
        })

        if err == nil {
            s.jobs[job.ID] = &cron.Entry{ID: entryID}
        }
    }
}

// ExecuteJob runs a cron job
func (s *Scheduler) ExecuteJob(job *db.ContentData) {
    // Update status to "running"
    job.SetField("last_status", "running")
    s.db.UpdateContent(job)

    actionType := job.GetField("action_type").(string)
    actionConfig := job.GetField("action_config").(map[string]interface{})

    var err error

    switch actionType {
    case "predefined":
        err = s.ExecutePredefinedAction(actionConfig)
    case "plugin":
        err = s.ExecutePluginAction(actionConfig)
    case "http":
        err = s.ExecuteHTTPAction(actionConfig)
    }

    // Update job metadata
    now := time.Now()
    status := "success"
    if err != nil {
        status = "error"
    }

    job.SetField("last_run", now)
    job.SetField("last_status", status)
    job.SetField("run_count", job.GetField("run_count").(int) + 1)

    // Calculate next run
    schedule := job.GetField("schedule").(string)
    nextRun := s.CalculateNextRun(schedule, now)
    job.SetField("next_run", nextRun)

    s.db.UpdateContent(job)
}

// ExecutePredefinedAction runs built-in actions
func (s *Scheduler) ExecutePredefinedAction(config map[string]interface{}) error {
    action := config["action"].(string)

    switch action {
    case "publish_scheduled_content":
        return s.PublishScheduledContent()
    case "cleanup_sessions":
        days := config["older_than_days"].(int)
        return s.CleanupSessions(days)
    case "backup_database":
        return s.BackupDatabase()
    case "cleanup_cache":
        return s.CleanupExpiredCache()
    default:
        return fmt.Errorf("unknown action: %s", action)
    }
}

// ExecutePluginAction runs Lua plugin
func (s *Scheduler) ExecutePluginAction(config map[string]interface{}) error {
    pluginName := config["plugin_name"].(string)
    function := config["function"].(string)

    // Load and execute Lua plugin
    return plugin.Execute(pluginName, function, config)
}

// ExecuteHTTPAction makes HTTP request
func (s *Scheduler) ExecuteHTTPAction(config map[string]interface{}) error {
    url := config["url"].(string)
    method := config["method"].(string)
    body := config["body"]

    // Make HTTP request
    return http.Request(method, url, body)
}

// PublishScheduledContent publishes content where publish_date <= now
func (s *Scheduler) PublishScheduledContent() error {
    now := time.Now()

    // Find all content with status='scheduled' and publish_date <= now
    scheduledContent := s.db.Query(`
        SELECT * FROM content_data
        WHERE status = 'scheduled'
          AND publish_date <= ?
    `, now)

    for _, content := range scheduledContent {
        content.Status = "published"
        s.db.UpdateContent(content)
    }

    return nil
}
```

**Advantages:**
- User-configurable (create jobs via TUI/API)
- No code deployment for new jobs
- Audit trail (last run, run count, status)
- Plugin extensibility (custom actions)
- Manageable (enable/disable jobs, view logs)

**Use cases:**
- Publish scheduled content
- Cleanup expired sessions
- Database backups
- Generate reports
- Sync with external systems
- Cache invalidation
- Send digest emails

---

### ✅ 3. Email Queue (via Datatype)

**Implementation:**
```sql
-- Create Email Queue datatype
INSERT INTO datatypes (name, slug) VALUES ('Email Queue', 'email_queue');

-- Create email fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('To', 'to', 'email'),
    ('From', 'from', 'email'),
    ('Subject', 'subject', 'text'),
    ('Body HTML', 'body_html', 'richtext'),
    ('Body Plain', 'body_plain', 'textarea'),
    ('Status', 'status', 'select'),        -- pending / sending / sent / failed
    ('Attempts', 'attempts', 'number'),
    ('Last Attempt', 'last_attempt', 'datetime'),
    ('Error Message', 'error', 'textarea'),
    ('Send At', 'send_at', 'datetime'),    -- Scheduled send
    ('Sent At', 'sent_at', 'datetime');

-- Create email queue route
INSERT INTO routes (name, slug, datatype_id)
VALUES ('Email Queue', 'email-queue', <email_queue_datatype_id>);
```

**Email worker process:**
```go
package email

import (
    "time"
    "net/smtp"
)

type EmailWorker struct {
    db     db.DbDriver
    config *config.Config
}

// Start begins processing email queue
func (w *EmailWorker) Start() {
    ticker := time.NewTicker(30 * time.Second)

    for range ticker.C {
        w.ProcessQueue()
    }
}

// ProcessQueue sends pending emails
func (w *EmailWorker) ProcessQueue() {
    now := time.Now()

    // Get pending emails where send_at <= now
    pendingEmails := w.db.Query(`
        SELECT * FROM content_data cd
        JOIN content_fields cf ON cd.id = cf.content_data_id
        WHERE cd.route_id = ?
          AND cf.field_id = ? AND cf.field_value = 'pending'
          AND (send_at IS NULL OR send_at <= ?)
        LIMIT 100
    `, <email_queue_route_id>, <status_field_id>, now)

    for _, email := range pendingEmails {
        w.SendEmail(email)
    }
}

// SendEmail sends a single email
func (w *EmailWorker) SendEmail(email *db.ContentData) {
    // Update status to "sending"
    email.SetField("status", "sending")
    w.db.UpdateContent(email)

    // Extract email fields
    to := email.GetField("to").(string)
    from := email.GetField("from").(string)
    subject := email.GetField("subject").(string)
    bodyHTML := email.GetField("body_html").(string)

    // Send via SMTP
    err := w.SendViaSMTP(to, from, subject, bodyHTML)

    // Update status
    now := time.Now()
    attempts := email.GetField("attempts").(int) + 1

    if err != nil {
        status := "failed"
        if attempts < 3 {
            status = "pending"  // Retry
        }

        email.SetField("status", status)
        email.SetField("error", err.Error())
    } else {
        email.SetField("status", "sent")
        email.SetField("sent_at", now)
    }

    email.SetField("attempts", attempts)
    email.SetField("last_attempt", now)
    w.db.UpdateContent(email)
}
```

**Queue email for sending:**
```go
// QueueEmail adds email to send queue
func QueueEmail(to, subject, body string, sendAt time.Time) error {
    return db.CreateContent(&db.ContentData{
        RouteID: <email_queue_route_id>,
        DatatypeID: <email_queue_datatype_id>,
        Fields: map[string]interface{}{
            "to": to,
            "from": config.Get("email_from"),
            "subject": subject,
            "body_html": body,
            "status": "pending",
            "attempts": 0,
            "send_at": sendAt,
        },
    })
}

// Send immediately
QueueEmail("user@example.com", "Welcome", "<h1>Welcome!</h1>", time.Now())

// Send at specific time
QueueEmail("user@example.com", "Reminder", "Your appointment is tomorrow",
    time.Now().Add(24 * time.Hour))
```

**Advantages:**
- Persistent queue (survives restarts)
- Retry logic (attempts counter)
- Scheduled sending (send_at field)
- Audit trail (sent_at, attempts, errors)
- Manageable (view/retry failed emails in TUI)

---

## Minimal Schema Additions (Single Table)

These need **one new table** but integrate with existing datatypes system.

---

### 🔶 4. Webhooks (Fetch Request + Junction Table)

**Schema addition:**
```sql
-- Create fetch_requests table
CREATE TABLE fetch_requests (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    method VARCHAR(10) DEFAULT 'POST',  -- GET, POST, PUT, DELETE
    headers JSON,                       -- {"Authorization": "Bearer token"}
    body_template TEXT,                 -- Go template with content data
    timeout INTEGER DEFAULT 30,         -- Seconds
    retry_count INTEGER DEFAULT 3,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create route_webhooks junction table
CREATE TABLE route_webhooks (
    id INTEGER PRIMARY KEY,
    route_id INTEGER NOT NULL,
    fetch_request_id INTEGER NOT NULL,
    trigger_event VARCHAR(50) NOT NULL,  -- create, update, delete, publish
    condition_json JSON,                 -- Optional: conditions for firing
    FOREIGN KEY (route_id) REFERENCES routes(id),
    FOREIGN KEY (fetch_request_id) REFERENCES fetch_requests(id)
);

-- Create webhook_deliveries log table (optional but recommended)
CREATE TABLE webhook_deliveries (
    id INTEGER PRIMARY KEY,
    fetch_request_id INTEGER NOT NULL,
    route_id INTEGER NOT NULL,
    content_data_id INTEGER,
    trigger_event VARCHAR(50),
    request_body TEXT,
    response_status INTEGER,
    response_body TEXT,
    error TEXT,
    delivered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (fetch_request_id) REFERENCES fetch_requests(id)
);
```

**Example webhook configurations:**
```sql
-- 1. Notify external API when content is published
INSERT INTO fetch_requests (name, url, method, headers, body_template) VALUES
('Notify CRM on Publish', 'https://api.crm.com/webhooks/content', 'POST',
'{"Authorization": "Bearer secret123"}',
'{"event": "content_published", "id": {{.ID}}, "title": "{{.Title}}", "url": "{{.URL}}"}');

-- Link to "Blog Posts" route
INSERT INTO route_webhooks (route_id, fetch_request_id, trigger_event)
VALUES (<blog_posts_route_id>, <fetch_request_id>, 'publish');

-- 2. Invalidate CDN cache on content update
INSERT INTO fetch_requests (name, url, method, headers) VALUES
('Cloudflare Cache Purge', 'https://api.cloudflare.com/client/v4/zones/ZONE_ID/purge_cache', 'POST',
'{"Authorization": "Bearer cf_token"}',
'{"files": ["{{.URL}}"]}');

INSERT INTO route_webhooks (route_id, fetch_request_id, trigger_event)
VALUES (<pages_route_id>, <fetch_request_id>, 'update');

-- 3. Send to search index when created
INSERT INTO fetch_requests (name, url, method, body_template) VALUES
('Algolia Index', 'https://APPID.algolia.net/1/indexes/content/batch', 'POST',
'{"requests": [{"action": "addObject", "body": {"objectID": "{{.ID}}", "title": "{{.Title}}", "content": "{{.Body}}"}}]}');

INSERT INTO route_webhooks (route_id, fetch_request_id, trigger_event)
VALUES (<all_routes_id>, <fetch_request_id>, 'create');
```

**Implementation:**
```go
package webhooks

import (
    "bytes"
    "encoding/json"
    "net/http"
    "text/template"
    "time"
)

type WebhookManager struct {
    db db.DbDriver
}

// TriggerWebhooks fires all webhooks for an event
func (w *WebhookManager) TriggerWebhooks(routeID int, event string, content *db.ContentData) {
    // Get all webhooks for this route + event
    webhooks := w.db.Query(`
        SELECT fr.* FROM fetch_requests fr
        JOIN route_webhooks rw ON fr.id = rw.fetch_request_id
        WHERE rw.route_id = ?
          AND rw.trigger_event = ?
          AND fr.active = TRUE
    `, routeID, event)

    // Fire each webhook
    for _, webhook := range webhooks {
        go w.DeliverWebhook(webhook, content, event)
    }
}

// DeliverWebhook makes HTTP request
func (w *WebhookManager) DeliverWebhook(webhook *FetchRequest, content *db.ContentData, event string) {
    // Parse body template
    tmpl, err := template.New("body").Parse(webhook.BodyTemplate)
    if err != nil {
        w.LogDelivery(webhook, content, event, 0, "", err.Error())
        return
    }

    // Execute template with content data
    var bodyBuf bytes.Buffer
    err = tmpl.Execute(&bodyBuf, content)
    if err != nil {
        w.LogDelivery(webhook, content, event, 0, "", err.Error())
        return
    }

    // Create HTTP request
    req, err := http.NewRequest(webhook.Method, webhook.URL, &bodyBuf)
    if err != nil {
        w.LogDelivery(webhook, content, event, 0, "", err.Error())
        return
    }

    // Add headers
    for key, value := range webhook.Headers {
        req.Header.Set(key, value)
    }

    // Make request with timeout
    client := &http.Client{Timeout: time.Duration(webhook.Timeout) * time.Second}
    resp, err := client.Do(req)

    if err != nil {
        // Retry logic
        if webhook.RetryCount > 0 {
            time.Sleep(5 * time.Second)
            w.DeliverWebhook(webhook, content, event)  // Recursive retry
        } else {
            w.LogDelivery(webhook, content, event, 0, "", err.Error())
        }
        return
    }
    defer resp.Body.Close()

    // Read response
    respBody := new(bytes.Buffer)
    respBody.ReadFrom(resp.Body)

    // Log delivery
    w.LogDelivery(webhook, content, event, resp.StatusCode, respBody.String(), "")
}

// LogDelivery records webhook delivery
func (w *WebhookManager) LogDelivery(webhook *FetchRequest, content *db.ContentData, event string, status int, respBody, error string) {
    w.db.Exec(`
        INSERT INTO webhook_deliveries
        (fetch_request_id, route_id, content_data_id, trigger_event, response_status, response_body, error)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, webhook.ID, content.RouteID, content.ID, event, status, respBody, error)
}
```

**Integration with content operations:**
```go
// In content creation
func CreateContent(content *db.ContentData) error {
    err := db.InsertContent(content)
    if err != nil {
        return err
    }

    // Trigger webhooks
    webhooks.TriggerWebhooks(content.RouteID, "create", content)

    return nil
}

// In content update
func UpdateContent(content *db.ContentData) error {
    err := db.UpdateContent(content)
    if err != nil {
        return err
    }

    // Trigger webhooks
    webhooks.TriggerWebhooks(content.RouteID, "update", content)

    return nil
}

// In content publish
func PublishContent(content *db.ContentData) error {
    content.Status = "published"
    err := db.UpdateContent(content)
    if err != nil {
        return err
    }

    // Trigger webhooks
    webhooks.TriggerWebhooks(content.RouteID, "publish", content)

    return nil
}
```

**Advantages:**
- Flexible (any HTTP endpoint)
- Event-driven (create, update, delete, publish)
- Template-based bodies (dynamic data)
- Retry logic
- Delivery logs (debugging, compliance)
- Manageable via TUI (add/edit/disable webhooks)

**Use cases:**
- Notify external services (CRM, analytics)
- CDN cache invalidation
- Search index updates (Algolia, Elasticsearch)
- Social media posting
- Slack/Discord notifications
- Email service triggers

---

### 🔶 5. Background Jobs Queue (via Table)

**Schema addition:**
```sql
CREATE TABLE job_queue (
    id INTEGER PRIMARY KEY,
    job_type VARCHAR(100) NOT NULL,      -- image_optimize, email_send, import_data
    payload JSON NOT NULL,                -- Job-specific data
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    priority INTEGER DEFAULT 0,           -- Higher = more important
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    scheduled_at TIMESTAMP,               -- Run at specific time
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_job_queue_status ON job_queue(status, priority DESC);
```

**Worker implementation:**
```go
package jobs

type JobWorker struct {
    db       db.DbDriver
    handlers map[string]JobHandler
}

type JobHandler func(payload map[string]interface{}) error

// RegisterHandler adds job handler
func (w *JobWorker) RegisterHandler(jobType string, handler JobHandler) {
    w.handlers[jobType] = handler
}

// Start begins processing jobs
func (w *JobWorker) Start(workerCount int) {
    for i := 0; i < workerCount; i++ {
        go w.ProcessJobs()
    }
}

// ProcessJobs polls and executes jobs
func (w *JobWorker) ProcessJobs() {
    ticker := time.NewTicker(5 * time.Second)

    for range ticker.C {
        // Get next pending job
        job := w.GetNextJob()
        if job == nil {
            continue
        }

        // Process job
        w.ExecuteJob(job)
    }
}

// GetNextJob retrieves highest priority pending job
func (w *JobWorker) GetNextJob() *Job {
    now := time.Now()

    // Atomic claim: update status to processing and return job
    result := w.db.QueryRow(`
        UPDATE job_queue
        SET status = 'processing', started_at = ?
        WHERE id = (
            SELECT id FROM job_queue
            WHERE status = 'pending'
              AND (scheduled_at IS NULL OR scheduled_at <= ?)
            ORDER BY priority DESC, created_at ASC
            LIMIT 1
        )
        RETURNING *
    `, now, now)

    return result
}

// ExecuteJob runs job handler
func (w *JobWorker) ExecuteJob(job *Job) {
    handler, exists := w.handlers[job.JobType]
    if !exists {
        w.FailJob(job, fmt.Errorf("no handler for job type: %s", job.JobType))
        return
    }

    // Execute handler
    err := handler(job.Payload)

    if err != nil {
        // Retry or fail
        job.Attempts++
        if job.Attempts >= job.MaxAttempts {
            w.FailJob(job, err)
        } else {
            // Retry with exponential backoff
            retryDelay := time.Duration(job.Attempts*job.Attempts) * time.Minute
            job.ScheduledAt = time.Now().Add(retryDelay)
            job.Status = "pending"
            w.db.Update(job)
        }
    } else {
        // Success
        w.CompleteJob(job)
    }
}

// Queue job for processing
func QueueJob(jobType string, payload map[string]interface{}) error {
    return db.Exec(`
        INSERT INTO job_queue (job_type, payload, status)
        VALUES (?, ?, 'pending')
    `, jobType, json.Marshal(payload))
}
```

**Example job handlers:**
```go
// Register job handlers
worker := &JobWorker{db: db, handlers: make(map[string]JobHandler)}

// Image optimization
worker.RegisterHandler("image_optimize", func(payload map[string]interface{}) error {
    mediaID := payload["media_id"].(int)
    return media.OptimizeImage(mediaID)
})

// Email sending
worker.RegisterHandler("email_send", func(payload map[string]interface{}) error {
    to := payload["to"].(string)
    subject := payload["subject"].(string)
    body := payload["body"].(string)
    return email.Send(to, subject, body)
})

// Data import
worker.RegisterHandler("import_data", func(payload map[string]interface{}) error {
    csvPath := payload["csv_path"].(string)
    return importer.ImportCSV(csvPath)
})

// Start workers
worker.Start(5)  // 5 concurrent workers
```

**Queue jobs:**
```go
// Queue image optimization
QueueJob("image_optimize", map[string]interface{}{
    "media_id": 123,
})

// Queue email with priority
db.Exec(`
    INSERT INTO job_queue (job_type, payload, priority)
    VALUES ('email_send', '{"to":"user@example.com","subject":"Urgent"}', 10)
`)

// Schedule job for future
db.Exec(`
    INSERT INTO job_queue (job_type, payload, scheduled_at)
    VALUES ('backup_database', '{}', ?)
`, time.Now().Add(24*time.Hour))
```

---

## Performance Comparison

| Feature | Traditional | Database-Backed | Performance Impact |
|---------|------------|-----------------|-------------------|
| **Cache** | Redis (in-memory) | SQLite/PostgreSQL | 10-100x slower, but persistent |
| **Cron Jobs** | System cron | Database poll | Minimal (runs infrequently) |
| **Webhooks** | Event queue (Redis) | Database triggers | Slight delay (async anyway) |
| **Email Queue** | Redis queue | Database table | Acceptable (email is async) |
| **Background Jobs** | Redis/RabbitMQ | Database table | Acceptable for low-moderate volume |

**When database-backed is fine:**
- Small to medium sites (<10,000 req/day)
- Low-frequency operations (cron, webhooks)
- Async operations (email, jobs)

**When to upgrade to traditional:**
- High traffic (>100,000 req/day)
- High-frequency cache hits (>1000/sec)
- Real-time requirements (<100ms latency)

---

## Summary: What Works Without Schema Changes

### ✅ Zero Schema Changes (Just Datatypes)
1. **SEO Fields** - Add fields to datatypes
2. **Menus** - Use tree structure
3. **Comments** - Use tree structure
4. **Redirects** - Redirects as datatype
5. **Form Builder** - Forms + submissions as datatypes
6. **Categories** - Use tree structure
7. **Related Content** - Reference fields
8. **Cache** - Cache entry datatype (with expiration)
9. **Cron Jobs** - Cron job datatype (with schedule)
10. **Email Queue** - Email queue datatype
11. **Workflow States** - Add columns to content_data

### 🔶 Minimal Schema (One Table)
12. **Webhooks** - fetch_requests + route_webhooks tables
13. **Tags (Many-to-Many)** - content_tags junction table
14. **Background Jobs** - job_queue table
15. **Audit Trail** - audit_log table

### ❌ Cannot Use Schema (Need Real Packages)
16. **High-performance cache** - Must be in-memory (Redis)
17. **Video processing** - External FFmpeg/service
18. **Advanced image transforms** - On-demand processing
19. **Real-time features** - WebSockets, push notifications

---

## Recommendations

**Start with datatypes:**
1. Implement cache, cron, email queue as datatypes
2. Test with real workload
3. Monitor performance

**Upgrade when needed:**
- If cache queries slow (>50ms): Add Redis
- If job queue grows (>1000 jobs): Add Redis queue
- If webhooks slow: Add async delivery queue

**Advantages of database-backed:**
- Persistent (survives restarts)
- Queryable (view in TUI)
- Manageable (CRUD via API)
- Simple (no external services)
- Good enough for 90% of use cases

The flexible schema is more powerful than it first appears!

---

**Last Updated:** 2026-01-16
