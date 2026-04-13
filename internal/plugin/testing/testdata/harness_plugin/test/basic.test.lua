-- basic.test.lua: exercises the test module assertions and harness features.

-- setup/teardown lifecycle
local setup_called = false
local teardown_called = false

test.setup(function()
    setup_called = true
end)

test.teardown(function()
    teardown_called = true
end)

-- Passing assertions
function test_assert_true()
    test.assert(true, "true should pass")
    test.assert(1, "truthy number should pass")
    test.assert("nonempty", "truthy string should pass")
end

function test_assert_eq()
    test.assert_eq(1, 1)
    test.assert_eq("hello", "hello")
    test.assert_eq(nil, nil)
end

function test_assert_neq()
    test.assert_neq(1, 2)
    test.assert_neq("a", "b")
    test.assert_neq(nil, 1)
end

function test_assert_nil()
    test.assert_nil(nil)
end

function test_assert_not_nil()
    test.assert_not_nil(1)
    test.assert_not_nil("x")
    test.assert_not_nil(true)
end

function test_assert_error()
    test.assert_error(function()
        error("boom")
    end)
    test.assert_error(function()
        error("specific message")
    end, "specific")
end

function test_assert_contains()
    test.assert_contains("hello world", "world")
    test.assert_contains("foobar", "oob")
end

-- DB operations within a test (rolled back via SAVEPOINT)
function test_db_insert_and_query()
    db.insert("harness_items", {label = "alpha", count = 10})
    local items = db.query("harness_items", {where = {label = "alpha"}})
    test.assert_eq(1, #items)
    test.assert_eq("alpha", items[1].label)
end

-- Verify SAVEPOINT isolation: previous test's insert should be rolled back
function test_savepoint_isolation()
    local n = db.count("harness_items", {})
    test.assert_eq(0, n, "table should be empty due to savepoint rollback")
end

-- HTTP request via test.request
function test_http_ping()
    local resp = test.request("GET", "/api/v1/plugins/harness_plugin/ping")
    test.assert_eq(200, resp.status)
    test.assert_not_nil(resp.json)
    test.assert_eq(true, resp.json.pong)
end

-- HTTP POST with body
function test_http_echo()
    local resp = test.request("POST", "/api/v1/plugins/harness_plugin/echo", {
        body = '{"msg":"hi"}'
    })
    test.assert_eq(200, resp.status)
    test.assert_eq("hi", resp.json.msg)
end

-- Setup was called
function test_setup_called()
    test.assert(setup_called, "setup should have been called")
end
