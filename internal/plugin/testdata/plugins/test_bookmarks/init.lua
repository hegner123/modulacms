-- test_bookmarks: Phase 1 integration test plugin for ModulaCMS
--
-- Exercises the full Phase 1 surface area inside on_init():
--   - db.define_table (multiple tables, all 7 column types, indexes)
--   - db.insert / db.query / db.query_one / db.update / db.delete
--   - db.count / db.exists
--   - db.transaction (commit and deliberate rollback)
--   - db.ulid / db.timestamp
--   - require (sandboxed loader from lib/)
--   - log.info / log.warn / log.debug
--   - on_shutdown lifecycle
--
-- Every operation verifies its result and calls error() on mismatch.
-- If on_init() completes without error, all assertions passed.

local validators = require("validators")

plugin_info = {
    name        = "test_bookmarks",
    version     = "1.0.0",
    description = "Phase 1 integration test plugin",
    author      = "ModulaCMS Test Suite",
}

function on_init()
    ---------------------------------------------------------------------------
    -- Schema: two tables covering all 7 abstract column types
    ---------------------------------------------------------------------------

    -- Table 1: collections (text, integer, boolean, json)
    db.define_table("collections", {
        columns = {
            {name = "name",        type = "text",    not_null = true},
            {name = "description", type = "text"},
            {name = "sort_order",  type = "integer", not_null = true, default = 0},
            {name = "is_public",   type = "boolean", not_null = true, default = 0},
            {name = "metadata",    type = "json"},
        },
        indexes = {
            {columns = {"name"}},
            {columns = {"is_public", "sort_order"}},
        },
    })

    -- Table 2: bookmarks (real, timestamp, blob + FK reference to collections)
    db.define_table("bookmarks", {
        columns = {
            {name = "collection_id", type = "text",      not_null = true},
            {name = "url",           type = "text",      not_null = true},
            {name = "title",         type = "text",      not_null = true},
            {name = "rating",        type = "real"},
            {name = "notes",         type = "blob"},
            {name = "visited_at",    type = "timestamp"},
        },
        indexes = {
            {columns = {"collection_id"}},
            {columns = {"url"}},
        },
    })

    log.info("Schema created: collections, bookmarks")

    ---------------------------------------------------------------------------
    -- Helpers: db.ulid() and db.timestamp()
    ---------------------------------------------------------------------------

    local col_id = db.ulid()
    local now = db.timestamp()

    if type(col_id) ~= "string" or col_id == "" then
        error("db.ulid() returned invalid value: " .. tostring(col_id))
    end
    if type(now) ~= "string" or now == "" then
        error("db.timestamp() returned invalid value: " .. tostring(now))
    end

    log.debug("Generated ULID", {id = col_id})
    log.debug("Generated timestamp", {ts = now})

    ---------------------------------------------------------------------------
    -- INSERT
    ---------------------------------------------------------------------------

    db.insert("collections", {
        id = col_id,
        name = "Dev Resources",
        sort_order = 1,
        is_public = 1,
        metadata = '{"theme":"dark"}',
    })

    log.info("Inserted collection", {id = col_id})

    ---------------------------------------------------------------------------
    -- QUERY: list with where clause
    ---------------------------------------------------------------------------

    local cols = db.query("collections", {where = {name = "Dev Resources"}})
    if #cols ~= 1 then
        error("expected 1 collection from query, got " .. #cols)
    end

    ---------------------------------------------------------------------------
    -- QUERY_ONE: hit
    ---------------------------------------------------------------------------

    local col = db.query_one("collections", {where = {id = col_id}})
    if not col then
        error("query_one returned nil for known collection " .. col_id)
    end

    ---------------------------------------------------------------------------
    -- QUERY_ONE: miss
    ---------------------------------------------------------------------------

    local miss = db.query_one("collections", {where = {id = "nonexistent"}})
    if miss ~= nil then
        error("query_one should return nil for missing row")
    end

    ---------------------------------------------------------------------------
    -- COUNT
    ---------------------------------------------------------------------------

    local n = db.count("collections", {})
    if n ~= 1 then
        error("expected count 1, got " .. tostring(n))
    end

    ---------------------------------------------------------------------------
    -- EXISTS: hit and miss
    ---------------------------------------------------------------------------

    local found = db.exists("collections", {where = {id = col_id}})
    if not found then
        error("exists returned false for known collection")
    end

    local not_found = db.exists("collections", {where = {id = "nonexistent"}})
    if not_found then
        error("exists returned true for missing collection")
    end

    ---------------------------------------------------------------------------
    -- INSERT bookmarks using require'd validator
    ---------------------------------------------------------------------------

    local urls = {
        {url = "https://go.dev/doc/",              title = "Go Documentation"},
        {url = "https://www.lua.org/manual/5.1/",  title = "Lua 5.1 Reference"},
        {url = "not-a-url",                        title = "Invalid"},
        {url = "",                                  title = "Empty"},
    }

    local inserted = 0
    for _, entry in ipairs(urls) do
        if validators.is_valid_url(entry.url) then
            db.insert("bookmarks", {
                collection_id = col_id,
                url   = entry.url,
                title = validators.trim(entry.title),
                rating = 4.5,
            })
            inserted = inserted + 1
        else
            log.warn("Skipped invalid URL", {url = entry.url})
        end
    end

    if inserted ~= 2 then
        error("expected 2 bookmark inserts, got " .. inserted)
    end

    ---------------------------------------------------------------------------
    -- QUERY: list all bookmarks (no where clause)
    ---------------------------------------------------------------------------

    local all_bookmarks = db.query("bookmarks", {order_by = "title", limit = 50})
    if #all_bookmarks ~= 2 then
        error("expected 2 bookmarks from query, got " .. #all_bookmarks)
    end

    ---------------------------------------------------------------------------
    -- UPDATE
    ---------------------------------------------------------------------------

    db.update("bookmarks", {
        set   = {rating = 5.0, visited_at = now},
        where = {url = "https://go.dev/doc/"},
    })

    local updated = db.query_one("bookmarks", {where = {url = "https://go.dev/doc/"}})
    if not updated then
        error("updated bookmark not found")
    end

    ---------------------------------------------------------------------------
    -- DELETE
    ---------------------------------------------------------------------------

    db.delete("bookmarks", {where = {url = "https://www.lua.org/manual/5.1/"}})

    local after_delete = db.count("bookmarks", {})
    if after_delete ~= 1 then
        error("expected 1 bookmark after delete, got " .. tostring(after_delete))
    end

    ---------------------------------------------------------------------------
    -- TRANSACTION: successful commit
    ---------------------------------------------------------------------------

    local ok, err = db.transaction(function()
        db.insert("bookmarks", {
            collection_id = col_id,
            url   = "https://pkg.go.dev/",
            title = "Go Packages",
        })
        db.insert("bookmarks", {
            collection_id = col_id,
            url   = "https://github.com/yuin/gopher-lua",
            title = "GopherLua",
        })
    end)

    if not ok then
        error("transaction commit failed: " .. tostring(err))
    end

    local post_tx = db.count("bookmarks", {})
    if post_tx ~= 3 then
        error("expected 3 bookmarks after committed tx, got " .. tostring(post_tx))
    end

    log.info("Transaction commit verified", {count = post_tx})

    ---------------------------------------------------------------------------
    -- TRANSACTION: deliberate rollback via error()
    ---------------------------------------------------------------------------

    local ok2, err2 = db.transaction(function()
        db.insert("bookmarks", {
            collection_id = col_id,
            url   = "https://should-not-persist.example.com/",
            title = "Rollback Test",
        })
        error("deliberate rollback")
    end)

    if ok2 then
        error("expected transaction to fail but it succeeded")
    end

    log.info("Transaction rollback verified", {err = err2})

    local post_rollback = db.count("bookmarks", {})
    if post_rollback ~= 3 then
        error("rollback failed: bookmark count is " .. tostring(post_rollback) .. ", expected 3")
    end

    ---------------------------------------------------------------------------
    -- Second collection to exercise multi-row queries
    ---------------------------------------------------------------------------

    local col2_id = db.ulid()
    db.insert("collections", {
        id = col2_id,
        name = "Reading List",
        description = "Articles to read later",
        sort_order = 2,
        is_public = 0,
    })

    local total_cols = db.count("collections", {})
    if total_cols ~= 2 then
        error("expected 2 collections, got " .. tostring(total_cols))
    end

    local public_cols = db.query("collections", {where = {is_public = 1}})
    if #public_cols ~= 1 then
        error("expected 1 public collection, got " .. #public_cols)
    end

    ---------------------------------------------------------------------------
    -- Done
    ---------------------------------------------------------------------------

    log.info("All Phase 1 checks passed", {
        collections = total_cols,
        bookmarks   = post_rollback,
    })
end

function on_shutdown()
    log.info("test_bookmarks shutting down")
end
