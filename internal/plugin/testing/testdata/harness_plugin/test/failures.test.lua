-- failures.test.lua: tests that deliberately fail assertions.
-- The Go test verifies these produce the expected failures in the report.

function test_assert_false()
    test.assert(false, "deliberate failure")
end

function test_assert_eq_mismatch()
    test.assert_eq(1, 2, "numbers differ")
end

function test_assert_nil_non_nil()
    test.assert_nil("not nil", "should be nil")
end
