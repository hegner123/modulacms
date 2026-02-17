╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Forms Plugin Plan

 Context

 ModulaCMS needs its first production-ready plugin: a form builder with submissions and webhook delivery queuing. This spans two deliverables:

 1. Lua plugin (plugins/forms/) — tables, routes, validation, webhook queue population
 2. Web components package (sdks/typescript/forms-components/) — three embeddable components for rendering, viewing entries, and building forms

 Webhook HTTP delivery is out of scope — a separate Go module will process the queue later. The Lua plugin only inserts queue rows.

 API Constraint: This plugin uses ONLY the currently available db.* API surface (db.query, db.query_one, db.insert, db.update, db.delete, db.count, db.exists, db.transaction, db.define_table, db.ulid, db.timestamp, db.timestamp_ago). No dependency on unimplemented APIs (db.query_advanced, db.insert_many, db.create_index are NOT used).

 Prerequisites (must be completed before plugin implementation)

 1. Fix db API sort direction support. The Lua db.query opts accept order_by as a string, but
    the Go query builder's ValidColumnName() rejects spaces — so order_by = "created_at DESC"
    fails validation. The Go SelectParams struct has a Desc bool field that is never exposed to
    Lua. Fix: in db_api.go, add a "desc" boolean parameter to the Lua opts table and wire it to
    SelectParams.Desc. After the fix, Lua usage becomes:
      db.query("form_entries", { order_by = "created_at", desc = true, limit = 20 })
    This also fixes the broken pattern in examples/plugins/task_tracker/init.lua and the
    incorrect documentation in ai/lua/PLUGIN_API.md.

 2. Fix db API comparison operators in where clauses. The Go buildWhere() function only generates
    equality (column = value) and IS NULL conditions. The forms plugin needs "greater than"
    comparisons for offset-based export pagination (WHERE id > last_id). Fix: support operator
    maps in where values. When a where value is a Lua table with a single operator key, generate
    the corresponding SQL operator:
      db.query("form_entries", { where = { form_id = id, id = { gt = last_id } } })
    Supported operators for v1: gt (>), gte (>=), lt (<), lte (<=). This is the minimum needed
    for cursor-based export and can be extended later.

 3. Add db.timestamp_ago(seconds) helper. The Lua sandbox has no os.time(), no date parsing
    library, and no way to perform arithmetic on RFC3339 timestamp strings. The forms plugin
    needs to compare timestamps for rate limit window expiry (is count_reset_at older than 1
    hour?). Fix: in db_api.go, register a "timestamp_ago" function that takes an integer
    (seconds) and returns an RFC3339 string representing (now - N seconds) in UTC. Lua usage:
      local one_hour_ago = db.timestamp_ago(3600)
      if form.count_reset_at < one_hour_ago then -- lexicographic comparison works for UTC RFC3339
    This is safe because RFC3339 with zero-padded UTC fields sorts lexicographically. The Go
    implementation is trivial: time.Now().UTC().Add(-time.Duration(n) * time.Second).Format(time.RFC3339)

 ---
 Part 1: Lua Plugin (plugins/forms/)

 File Structure

 plugins/forms/
   init.lua              -- manifest, route registration, lifecycle, schema (db.define_table calls)
   lib/
     validators.lua      -- validation for forms, fields, submissions, webhooks
     forms.lua           -- form CRUD handlers
     fields.lua          -- field CRUD + reorder handlers
     entries.lua         -- submit, list, get, delete, export + webhook queue
     webhooks.lua        -- webhook config CRUD handlers
     utils.lua           -- pagination, error response helpers

 Tables (5)

 All auto-prefixed as plugin_forms_* with auto-columns id, created_at, updated_at.

 1. forms — form definitions

 ┌─────────────────┬─────────┬────────────────────────────────────────────────────┐
 │     Column      │  Type   │                    Constraints                     │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ name            │ text    │ not_null                                           │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ description     │ text    │                                                    │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ submit_label    │ text    │ not_null, default "Submit"                         │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ success_message │ text    │ not_null, default "Thank you for your submission." │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ redirect_url    │ text    │                                                    │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ enabled         │ boolean │ not_null, default 1                                │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ version         │ integer │ not_null, default 1                                │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ captcha_config  │ json    │                                                    │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ captcha_secret  │ text    │                                                    │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ rate_limit      │ integer │ not_null, default 100                              │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ sub_count       │ integer │ not_null, default 0                                │
 ├─────────────────┼─────────┼────────────────────────────────────────────────────┤
 │ count_reset_at  │ text    │ not_null, default ""                               │
 └─────────────────┴─────────┴────────────────────────────────────────────────────┘

 captcha_config: Optional JSON for CAPTCHA integration. Example: {"provider": "recaptcha", "site_key": "..."}. Web components read provider/site_key from captcha_config. null = no CAPTCHA.
 captcha_secret: CAPTCHA secret key stored as a separate column, NOT inside captcha_config JSON. This separation ensures the public endpoint (GET /forms/{id}/public) can return captcha_config directly without stripping nested keys. The secret is only used server-side for CAPTCHA verification.
 rate_limit: Max submissions per hour for this form (across all IPs). Default 100. Enforced via the sub_count/count_reset_at counter (see Anti-spam design decision below).
 sub_count: Rolling submission counter for rate limiting. Incremented on each successful submit. Reset to 0 when count_reset_at is more than 1 hour old.
 count_reset_at: Timestamp of last counter reset. Compared in Lua against db.timestamp() to determine if the window has expired. Empty string on creation; set to current timestamp on first submit.

 Indexes: name, enabled, created_at

 2. form_fields — field definitions per form

 ┌──────────────────┬─────────┬──────────────────────────────────────────────────────────────┐
 │      Column      │  Type   │                         Constraints                          │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ form_id          │ text    │ not_null, FK -> plugin_forms_forms CASCADE                   │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ name             │ text    │ not_null                                                     │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ label            │ text    │ not_null                                                     │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ field_type       │ text    │ not_null                                                     │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ placeholder      │ text    │                                                              │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ default_value    │ text    │                                                              │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ help_text        │ text    │                                                              │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ required         │ boolean │ not_null, default 0                                          │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ validation_rules │ json    │                                                              │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ options          │ json    │                                                              │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ position         │ integer │ not_null, default 0                                          │
 ├──────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ config           │ json    │                                                              │
 └──────────────────┴─────────┴──────────────────────────────────────────────────────────────┘

 Indexes: form_id, (form_id, position), (form_id, name) UNIQUE

 3. form_entries — submissions

 ┌──────────────┬─────────┬──────────────────────────────────────────────────────┐
 │    Column    │  Type   │                     Constraints                      │
 ├──────────────┼─────────┼──────────────────────────────────────────────────────┤
 │ form_id      │ text    │ not_null, FK -> plugin_forms_forms CASCADE           │
 ├──────────────┼─────────┼──────────────────────────────────────────────────────┤
 │ form_version │ integer │ not_null                                             │
 ├──────────────┼─────────┼──────────────────────────────────────────────────────┤
 │ data         │ json    │ not_null                                             │
 ├──────────────┼─────────┼──────────────────────────────────────────────────────┤
 │ client_ip    │ text    │                                                      │
 ├──────────────┼─────────┼──────────────────────────────────────────────────────┤
 │ user_agent   │ text    │                                                      │
 ├──────────────┼─────────┼──────────────────────────────────────────────────────┤
 │ status       │ text    │ not_null, default "submitted"                        │
 └──────────────┴─────────┴──────────────────────────────────────────────────────┘

 Indexes: form_id, (form_id, status), created_at, status

 4. form_webhooks — webhook configs per form

 ┌─────────┬─────────┬──────────────────────────────────────────────────────────────┐
 │ Column  │  Type   │                         Constraints                          │
 ├─────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ form_id │ text    │ not_null, FK -> plugin_forms_forms CASCADE                   │
 ├─────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ url     │ text    │ not_null                                                     │
 ├─────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ method  │ text    │ not_null, default "POST"                                     │
 ├─────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ headers │ json    │                                                              │
 ├─────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ events  │ text    │ not_null                                                     │
 ├─────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ active  │ boolean │ not_null, default 1                                          │
 ├─────────┼─────────┼──────────────────────────────────────────────────────────────┤
 │ secret  │ text    │                                                              │
 └─────────┴─────────┴──────────────────────────────────────────────────────────────┘

 events: Comma-delimited text, NOT JSON. Example: "entry.created,entry.deleted". Matching is done in Lua via string.find after fetching active webhooks for the form. This avoids LIKE queries on JSON columns which are fragile and not portable across SQLite/MySQL/PostgreSQL.
 Valid event names: entry.created, entry.updated, entry.deleted, form.deleted

 Indexes: form_id, (form_id, active)

 5. webhook_queue — delivery queue (Go module processes this)

 ┌───────────────┬───────────┬──────────────────────────────────────────────────────────────────┐
 │    Column     │   Type    │                           Constraints                            │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ webhook_id    │ text      │ not_null, FK -> plugin_forms_form_webhooks SET NULL              │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ entry_id      │ text      │ not_null, FK -> plugin_forms_form_entries SET NULL               │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ event         │ text      │ not_null                                                         │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ payload       │ json      │ not_null                                                         │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ status        │ text      │ not_null, default "pending"                                      │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ attempts      │ integer   │ not_null, default 0                                              │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ next_retry_at │ text      │                                                                  │
 ├───────────────┼───────────┼──────────────────────────────────────────────────────────────────┤
 │ last_error    │ text      │                                                                  │
 └───────────────┴───────────┴──────────────────────────────────────────────────────────────────┘

 Indexes: status, (status, next_retry_at), webhook_id, entry_id

 Routes (21 total, under /api/v1/plugins/forms/)

 Public (2):

 ┌────────┬────────────────────┬─────────────────────────┬───────────────────────────────────────────┐
 │ Method │        Path        │         Handler         │                Description                │
 ├────────┼────────────────────┼─────────────────────────┼───────────────────────────────────────────┤
 │ GET    │ /forms/{id}/public │ entries.get_public_form │ Get enabled form + fields                 │
 ├────────┼────────────────────┼─────────────────────────┼───────────────────────────────────────────┤
 │ POST   │ /forms/{id}/submit │ entries.submit          │ Submit entry (validates, queues webhooks) │
 └────────┴────────────────────┴─────────────────────────┴───────────────────────────────────────────┘

 Admin (19):

 ┌────────┬─────────────────────────────────┬─────────────────────┬──────────────────────────────────────────────────────────────────┐
 │ Method │              Path               │       Handler       │                          Description                           │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /forms                          │ forms.list          │ List forms (paginated)                                           │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ POST   │ /forms                          │ forms.create        │ Create form                                                      │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /forms/{id}                     │ forms.get           │ Get form + fields                                                │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ PUT    │ /forms/{id}                     │ forms.update        │ Update form (bumps version, requires version in body)            │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ DELETE │ /forms/{id}                     │ forms.delete        │ Delete form (cascade)                                            │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /forms/{id}/fields              │ fields.list         │ List fields by position                                          │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ POST   │ /forms/{id}/fields              │ fields.create       │ Add field (bumps version)                                        │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ PUT    │ /fields/{id}                    │ fields.update       │ Update field (bumps version)                                     │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ DELETE │ /fields/{id}                    │ fields.delete       │ Delete field (bumps version)                                     │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ POST   │ /forms/{id}/fields/reorder      │ fields.reorder      │ Reorder via {field_ids: [...], version: N}                       │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /forms/{id}/entries             │ entries.list        │ List entries (paginated)                                         │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /entries/{id}                   │ entries.get         │ Get single entry                                                 │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ DELETE │ /entries/{id}                   │ entries.delete      │ Delete entry (queues webhooks)                                   │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /forms/{id}/export              │ entries.export      │ Export entries as JSON (max 10,000 rows, cursor via ?after=)      │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /forms/{id}/webhooks            │ webhooks.list       │ List webhooks                                                    │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ POST   │ /forms/{id}/webhooks            │ webhooks.create     │ Create webhook                                                   │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ PUT    │ /webhooks/{id}                  │ webhooks.update     │ Update webhook                                                   │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ DELETE │ /webhooks/{id}                  │ webhooks.delete     │ Delete webhook                                                   │
 ├────────┼─────────────────────────────────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ GET    │ /forms/{id}/webhooks/queue      │ webhooks.queue_info │ Queue depth + recent failures for this form's webhooks           │
 └────────┴─────────────────────────────────┴─────────────────────┴──────────────────────────────────────────────────────────────────┘

 Key Design Decisions

 - Form versioning: forms.version increments on every form or field mutation. form_entries.form_version captures which version was submitted against.
 - Optimistic concurrency: PUT /forms/{id}, PUT /fields/{id}, POST /forms/{id}/fields/reorder all require a "version" field in the request body matching the current forms.version. If the version does not match, return 409 Conflict with {error: "version conflict", current_version: N}. This prevents silent overwrites when two admins edit concurrently.
 - Field types: text, textarea, email, number, tel, url, date, time, datetime, select, radio, checkbox, hidden, file
 - Password field: DEFERRED. Requires a crypto API (crypto.bcrypt_hash) exposed from the Go host. No bcrypt or hashing functions are available in the Lua sandbox. Will be added when the crypto API ships.
 - Submission validation: Checks required fields, email format, URL format, number type, select/radio option membership, min_length/max_length from validation_rules JSON.
 - Webhook event matching: Query all active webhooks for the form via db.query("form_webhooks", {where = {form_id = id, active = 1}}), then filter in Lua using string.find on the comma-delimited events column. Webhook count per form is small (typically 1-5), so Lua-side filtering is efficient. No LIKE queries on JSON columns.
 - Webhook queue population: On submit, match webhooks in Lua, then insert each queue row individually inside a db.transaction. At typical webhook counts (1-5 per form), individual inserts are negligible. The entire submit flow (insert entry + insert queue rows) is wrapped in a single db.transaction for atomicity.
 - Sort direction: All db.query calls use the desc boolean parameter (see Prerequisites). Example:
   db.query("form_entries", { where = { form_id = id }, order_by = "created_at", desc = true, limit = 20 })
   This replaces the broken order_by = "created_at DESC" pattern from the task_tracker example.
 - Delete ordering: Entry and webhook delete handlers queue webhook events BEFORE performing the delete. webhook_queue FKs use SET NULL (not CASCADE) so queue rows survive source deletion. The Go queue processor must handle NULL webhook_id/entry_id gracefully (skip or mark as undeliverable).
 - Form deletion strategy: Form delete does NOT iterate entries to queue per-entry webhook events. This avoids the 1,000-op-per-request limit (plugin_max_ops) which would be exceeded by any form with more than ~300 entries and active webhooks. Instead, form delete: (1) fetches active webhooks for the form, (2) queues a single "form.deleted" event per active webhook with the form metadata as payload, (3) performs the delete (CASCADE removes entries, fields, webhooks). Webhook consumers that need to react to bulk entry loss should subscribe to "form.deleted". This caps the operation count to: 1 form fetch + 1 webhook query + N queue inserts (N = active webhook count, typically 1-5) + 1 delete = under 10 operations regardless of entry count.
 - Export: GET /forms/{id}/export uses db.query with the hard cap of 10,000 rows per page. Uses cursor-based pagination via the comparison operator prerequisite: WHERE id > after_id (using the gt operator map). Clients pass ?after=<last_entry_id> to fetch the next page. Response: {items: [...], count: N, after: "last_id_or_null", has_more: bool}. Each page is a separate request/response, avoiding memory accumulation in Lua. ULID IDs are lexicographically sortable, so id ordering is chronological.
 - Pagination: ?limit= (1-100, default 20) and ?offset= (default 0). Response: {items, total, limit, offset}
 - Error format: {error: "message"} consistently across all endpoints
 - Boolean handling: Explicitly convert true->1, false->0 for SQLite. Read-back checks both == 1 and == true.
 - Anti-spam: Public submit endpoint includes:
   1. Honeypot validation: A hidden field (_hp) rendered by consumers must be empty. If filled, return 200 with standard success response (do not reveal detection to bots).
   2. Per-IP rate limiting from the plugin HTTP bridge.
   3. Per-form submission throttle via counter columns: The forms table has sub_count (integer) and count_reset_at (timestamp text). On submit, the handler fetches the form row (already needed for enabled check + version). In Lua, it compares count_reset_at against db.timestamp_ago(3600) using lexicographic string comparison (valid for zero-padded UTC RFC3339 strings). If the window has expired (count_reset_at < one_hour_ago or count_reset_at == ""), it resets sub_count to 0 and updates count_reset_at to db.timestamp() (single db.update). If sub_count >= rate_limit, return 429 {error: "submission limit exceeded"}. On successful submit, increment sub_count via db.update (setting sub_count = form.sub_count + 1). Example:
       local one_hour_ago = db.timestamp_ago(3600)
       if form.count_reset_at == "" or form.count_reset_at < one_hour_ago then
           db.update("forms", form.id, { sub_count = 0, count_reset_at = db.timestamp() })
           form.sub_count = 0
       end
       if form.sub_count >= form.rate_limit then
           return { status = 429, json = { error = "submission limit exceeded" } }
       end
   This avoids comparison operators in WHERE clauses — all checks happen in Lua after fetching the form row. Trade-off: under high concurrency, two requests may read the same sub_count and both increment to the same value, causing the counter to underreport by 1. This is acceptable for anti-spam — it means slightly more submissions get through, not fewer.
   4. CAPTCHA integration point: forms.captcha_config JSON field + forms.captcha_secret text column. Web components read provider/site_key from captcha_config for client-side widget rendering. Server-side CAPTCHA validation is NOT available in v1 — the Go host does not yet expose an HTTP client helper for external API calls. If captcha_config is set, the submit handler logs a warning ("CAPTCHA configured but server-side validation unavailable") and allows the submission. The schema fields are present for forward compatibility. null captcha_config = no CAPTCHA required.
 - Public endpoint secret stripping: GET /forms/{id}/public MUST exclude the captcha_secret column from the response. The handler returns captcha_config as-is (it contains only provider and site_key) and omits captcha_secret entirely. Webhook secrets (form_webhooks.secret) are never exposed on any public endpoint.
 - File field handling: file field type accepts base64-encoded file data in the JSON submission body. Max 5MB per file field (enforced via validation_rules.max_file_size, default 5MB). Max request body size is configurable via CMS admin (plugin_max_request_body in config). Storage note: base64 data is stored directly in the form_entries.data JSON column. For high-volume forms with file uploads, this will grow the database. Future integration with the CMS media system (internal/media/) is planned but out of scope for v1.
 - FK table references: All foreign key ref_table values use the fully prefixed table name (e.g., plugin_forms_forms, plugin_forms_form_webhooks, plugin_forms_form_entries). The plugin prefix is auto-applied to table names in Lua code but FK references must be explicit. This is reflected in the table definitions above.
 - Reorder validation: POST /forms/{id}/fields/reorder requires:
   1. The field_ids array must contain exactly all field IDs belonging to the specified form (no partial reorders).
   2. All IDs must belong to the specified form (prevents cross-form field manipulation).
   3. No duplicate IDs in the array.
   4. A version field matching the current forms.version (optimistic concurrency).
   Violations return 400 with a specific error message.
 - Error handling convention: The db.* API has two distinct error paths:
   (a) Recoverable errors (bad table name, SQL execution failure): returns nil, errmsg (2 values).
   (b) Unrecoverable errors (op limit exceeded, missing VM context): raises a Lua error via
       error() that terminates the handler entirely. These cannot be caught with nil checks.
   All db.query and db.query_one calls MUST check for nil return values to handle path (a). Pattern:
     local result, err = db.query("form_entries", opts)
     if not result then
         return { status = 500, json = { error = "query failed: " .. (err or "unknown") } }
     end
   This prevents "attempt to get length of a nil value" crashes when SQL queries fail. Path (b)
   errors are handled by the plugin runtime's recover mechanism and produce automatic 500 responses.
   Inside db.transaction callbacks, db.* errors propagate as Lua errors that trigger automatic
   rollback. Individual error checks on db.insert/db.update are NOT needed inside transactions —
   if any operation fails, the entire callback unwinds and the transaction rolls back. Do NOT
   attempt to return HTTP responses from inside a transaction callback.
 - Webhook queue monitoring: GET /forms/{id}/webhooks/queue returns {pending: N, failed: N, recent_failures: [...]} by querying webhook_queue rows linked to this form's webhooks. Allows operators to monitor queue health without direct database access.

 Reference Pattern

 - examples/plugins/task_tracker/init.lua — follows the same module-scope route registration, require() pattern, and handler conventions

 Known Limitations (v1)

 - No password field type (requires crypto API from Go host)
 - Export capped at 10,000 rows per request (use cursor pagination via ?after= for larger datasets)
 - No form version diff/rollback (version is tracked but only the current state is stored)
 - No optimistic concurrency on entry delete (low contention endpoint, acceptable for v1)
 - No CAPTCHA validation server-side until Go host exposes an HTTP client helper for external API calls. The captcha_config schema field is present for forward compatibility; the submit handler logs a warning if captcha_config is set but validation cannot be performed.
 - Plugin removal leaves 5 tables in the database. Orphaned tables can be dropped via the admin cleanup API. Data dependencies (entries referencing forms, queue referencing webhooks) are handled by CASCADE/SET NULL FKs, so table drops will succeed in any order.
 - Plaintext secret storage: Webhook secrets (form_webhooks.secret) and CAPTCHA secret keys (forms.captcha_secret) are stored as plaintext. Anyone with database read access (backups, admin queries) can read them. Remediation: when the crypto API ships (same dependency as the password field type), add encryption-at-rest using a server-managed key. Until then, operators should restrict database access and treat backups as sensitive.
 - Form deletion does not queue per-entry webhook events. Only a single "form.deleted" event is queued per active webhook. This is a trade-off to stay within the 1,000-op-per-request limit (plugin_max_ops). Webhook consumers that need entry-level granularity should export entries before deleting a form.
 - Per-form rate limit counter (sub_count) may undercount by 1 under concurrent submissions due to read-then-write without atomic increment. Acceptable for anti-spam purposes.

 ---
 Part 2: Web Components (sdks/typescript/forms-components/)

 Package: @modulacms/forms

 Three web components distributed as a single npm package. Zero runtime dependencies (vanilla HTMLElement subclasses). Shadow DOM with CSS custom properties and ::part() for style
 encapsulation.

 File Structure

 sdks/typescript/forms-components/
   package.json
   tsconfig.json
   tsup.config.ts
   vitest.config.ts
   src/
     index.ts                          -- re-exports + customElements.define()
     types.ts                          -- FormDefinition, FormFieldDefinition, FormEntry, etc.
     api-client.ts                     -- fetch-based client for /api/v1/plugins/forms/*
     utils/
       dom.ts                          -- shadow DOM helpers, html tagged template
       validation.ts                   -- client-side field validation engine
       state.ts                        -- lightweight reactive state container
       events.ts                       -- typed CustomEvent dispatch
       styles.ts                       -- shared CSS custom property definitions
       aria.ts                         -- accessibility helpers
       drag.ts                         -- drag-and-drop for form builder
     components/
       modulacms-form.ts               -- <modulacms-form>
       modulacms-entries.ts            -- <modulacms-entries>
       modulacms-form-builder.ts       -- <modulacms-form-builder>
     styles/
       base.css.ts                     -- base CSS as template literal
       form.css.ts                     -- form renderer styles
       entries.css.ts                  -- entries viewer styles
       builder.css.ts                  -- form builder styles
     tests/
       api-client.test.ts
       validation.test.ts
       modulacms-form.test.ts
       modulacms-entries.test.ts
       modulacms-form-builder.test.ts

 Component 1: <modulacms-form> — Form Renderer

 Attributes: form-id, api-url, api-key, submit-label, success-message, redirect-url, loading-text

 Methods: reset(), validate(), submit(), setFieldValue(name, value), getFieldValue(name)

 Events: form:loaded, form:submit (cancelable), form:success, form:error, field:change, field:validate

 CSS Parts: form, field, label, input, help-text, error, submit, loading, success, error-state

 CSS Custom Properties: --modulacms-font-family, --modulacms-primary-color, --modulacms-error-color, --modulacms-border-color, --modulacms-field-gap, --modulacms-input-padding,
 --modulacms-border-radius, --modulacms-button-*, etc.

 Component 2: <modulacms-entries> — Entries Viewer

 Attributes: form-id, api-url, api-key, page-size, sortable, filterable, export-enabled

 Methods: refresh(), goToPage(n), exportEntries(format), setFilter(filter), clearFilters(), setSort(field, dir)

 Events: entries:loaded, entry:select, entries:page-change, entries:export

 CSS Parts: table, thead, th, tbody, tr, td, pagination, page-button, export-button, filter-input, empty-state, loading, error-state

 Component 3: <modulacms-form-builder> — Form Builder

 Attributes: form-id, api-url, api-key, auto-save

 Methods: save(), addField(type), removeField(index), moveField(from, to), getDefinition(), setDefinition(def)

 Events: builder:loaded, builder:save, builder:change, field:add, field:remove, field:reorder

 CSS Parts: builder, toolbar, field-palette, field-type-button, canvas, field-item, field-handle, field-config, config-input, preview, save-button, loading, error-state

 Build Output

 ┌────────┬──────────────────────────────┬──────────────────────────────┐
 │ Format │             File             │            Usage             │
 ├────────┼──────────────────────────────┼──────────────────────────────┤
 │ ESM    │ dist/index.js                │ import '@modulacms/forms'    │
 ├────────┼──────────────────────────────┼──────────────────────────────┤
 │ CJS    │ dist/index.cjs               │ require('@modulacms/forms')  │
 ├────────┼──────────────────────────────┼──────────────────────────────┤
 │ IIFE   │ dist/modulacms-forms.iife.js │ <script src="..."> CDN usage │
 └────────┴──────────────────────────────┴──────────────────────────────┘

 Workspace Integration

 Modify: sdks/typescript/pnpm-workspace.yaml — add 'forms-components'

 No CI changes needed — .github/workflows/sdks.yml already triggers on sdks/typescript/**.

 ---
 Implementation Order

 Phase 0: db API Prerequisite Fixes (Go, in internal/plugin/db_api.go and internal/db/query_builder.go)

 0a. Add "desc" boolean parameter to Lua db.query opts — wire to SelectParams.Desc
 0b. Add comparison operator support to buildWhere — parse Lua table values with gt/gte/lt/lte keys
 0c. Add db.timestamp_ago(seconds) — returns RFC3339 string of (now - N seconds) in UTC
 0d. Fix task_tracker example and PLUGIN_API.md docs to use the new desc parameter
 0e. Verify with tests: db.query with order_by + desc, db.query with where gt operator, db.timestamp_ago returns valid RFC3339 and sorts correctly

 Phase 1: Lua Plugin Foundation

 1. lib/utils.lua — no deps, used by everything
 2. lib/validators.lua — validation engine
 3. on_init() schema — 5 db.define_table calls with indexes (inside init.lua on_init function, following task_tracker pattern)

 Phase 2: Lua Plugin Route Handlers

 4. lib/forms.lua — form CRUD (with optimistic concurrency on update)
 5. lib/fields.lua — field CRUD + reorder + version bumping + reorder validation
 6. lib/webhooks.lua — webhook config CRUD + queue_info handler
 7. lib/entries.lua — submit (with transaction, anti-spam counter, webhook queue), list, get, delete, export (with cursor pagination via gt operator)
 8. init.lua — manifest, route registration, lifecycle

 Phase 3: Web Components Foundation

 9. Package setup: package.json, tsconfig.json, tsup.config.ts, vitest.config.ts
 10. src/types.ts — all TypeScript type definitions
 11. src/api-client.ts — forms plugin API client
 12. src/utils/ — state, validation, dom, events, styles, aria

 Phase 4: Web Components Implementation

 13. src/components/modulacms-form.ts + src/styles/form.css.ts
 14. src/components/modulacms-entries.ts + src/styles/entries.css.ts
 15. src/utils/drag.ts + src/components/modulacms-form-builder.ts + src/styles/builder.css.ts
 16. src/index.ts — auto-registration + re-exports

 Phase 5: Integration

 17. Update pnpm-workspace.yaml
 18. Tests for API client, validation, components

 ---
 Verification

 1. Lua plugin: Enable plugins in config.json, start server, verify plugin loads: just run then modulacms plugin info forms
 2. Route approval: modulacms plugin approve forms --all-routes
 3. Smoke test: curl the public form endpoint, submit an entry, verify webhook queue row created
 4. Web components: cd sdks/typescript && pnpm install && pnpm -r build && pnpm -r test
 5. IIFE: Open a test HTML file that loads dist/modulacms-forms.iife.js and renders <modulacms-form>

 Testing Areas (require coverage before production)

 Validation engine:
 - Required field enforcement (present, non-empty, non-nil)
 - Email format validation (valid and invalid patterns)
 - URL format validation
 - Number type coercion and rejection of non-numeric input
 - Select/radio option membership (value in allowed options set)
 - min_length / max_length boundary conditions (exact boundary, off-by-one)
 - File field: base64 decoding, max_file_size enforcement
 - Honeypot detection (filled _hp returns 200 but does not create entry)

 Submit-to-queue flow:
 - Submit with 0 active webhooks (no queue rows created)
 - Submit with N active webhooks (N queue rows created, correct payload/event)
 - Submit with mixed active/inactive webhooks (only active ones queued)
 - Event filtering: webhook subscribed to "entry.created" only gets entry.created events
 - Transaction atomicity: if queue insert fails, entry insert is also rolled back

 Delete ordering:
 - Delete entry: webhook queue rows created BEFORE entry is deleted
 - Delete form: single "form.deleted" event queued per active webhook, then CASCADE delete
 - Queue rows survive source deletion (SET NULL FK verified)
 - Form delete with 0 active webhooks: no queue rows, delete succeeds
 - Form delete with N active webhooks: N queue rows with form.deleted event

 Optimistic concurrency:
 - PUT /forms/{id} with correct version succeeds
 - PUT /forms/{id} with stale version returns 409
 - Concurrent field mutations both attempt same version, one succeeds, one gets 409

 Anti-spam:
 - Per-form rate_limit enforcement via sub_count/count_reset_at counter
 - Counter reset when window expires (count_reset_at > 1 hour old)
 - 429 returned when sub_count >= rate_limit
 - sub_count incremented on successful submit
 - Honeypot field filled: 200 returned, no entry created
 - Honeypot field empty: normal processing

 Export:
 - Form with <10,000 entries: full export, has_more: false, after: null
 - Form with >10,000 entries: first 10,000 returned, has_more: true, after: last entry id
 - ?after=<id> follow-up request returns next page (uses gt operator on id column)

 Error handling:
 - db.query returning nil triggers 500 response (not Lua crash)
 - All handler paths check db operation results before using them

 Tests use SQLite databases in testdb/ following the existing test pattern.

 ---
 Critical Reference Files

 - examples/plugins/task_tracker/init.lua — primary Lua plugin pattern reference
 - internal/plugin/db_api.go — db.* API implementation details
 - internal/plugin/http_api.go — http.handle implementation details
 - sdks/typescript/modulacms-admin-sdk/package.json — package.json structure to replicate
 - sdks/typescript/modulacms-admin-sdk/tsup.config.ts — build config to replicate
 - sdks/typescript/pnpm-workspace.yaml — must add new package
