package testing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	stdtesting "testing"

	plugintesting "github.com/hegner123/modulacms/internal/plugin/testing"
)

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *stdtesting.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata")
}

func pluginDir(t *stdtesting.T, name string) string {
	t.Helper()
	return filepath.Join(testdataDir(t), name)
}

// --- NewHarness ---

func TestNewHarness_ValidPlugin(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()
}

func TestNewHarness_InvalidPlugin(t *stdtesting.T) {
	_, err := plugintesting.NewHarness(pluginDir(t, "nonexistent"), plugintesting.HarnessOpts{})
	if err == nil {
		t.Fatal("expected error for nonexistent plugin directory")
	}
}

// --- DiscoverTests ---

func TestDiscoverTests_FindsFiles(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	files, err := h.DiscoverTests()
	if err != nil {
		t.Fatalf("DiscoverTests: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 test files, got %d: %v", len(files), files)
	}

	// Should be sorted alphabetically
	if files[0] != "basic.test.lua" {
		t.Errorf("expected first file basic.test.lua, got %s", files[0])
	}
	if files[1] != "failures.test.lua" {
		t.Errorf("expected second file failures.test.lua, got %s", files[1])
	}
}

func TestDiscoverTests_NoTestDir(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "no_tests_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	_, err = h.DiscoverTests()
	if err == nil {
		t.Fatal("expected error for plugin with no test directory")
	}
	if !strings.Contains(err.Error(), "no test/ directory") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- RunAll: passing tests ---

func TestRunAll_PassingTests(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	ctx := context.Background()
	report := h.RunAll(ctx, []string{"basic.test.lua"})

	if report.Failed() != 0 {
		var buf bytes.Buffer
		report.WriteHuman(&buf, false)
		t.Fatalf("expected all tests to pass, got %d failures:\n%s", report.Failed(), buf.String())
	}

	if report.Passed() == 0 {
		t.Fatal("expected at least one passing test")
	}

	if report.ExitCode() != 0 {
		t.Errorf("expected exit code 0, got %d", report.ExitCode())
	}
}

// --- RunAll: failing tests ---

func TestRunAll_FailingTests(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	ctx := context.Background()
	report := h.RunAll(ctx, []string{"failures.test.lua"})

	if report.Failed() != 3 {
		var buf bytes.Buffer
		report.WriteHuman(&buf, false)
		t.Fatalf("expected 3 failing tests, got %d:\n%s", report.Failed(), buf.String())
	}

	if report.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", report.ExitCode())
	}

	// Verify each failure has an assertion message
	for _, r := range report.Results {
		if r.Passed {
			t.Errorf("test %s should have failed", r.Test)
			continue
		}
		if len(r.Failures) == 0 {
			t.Errorf("test %s failed but has no failure records", r.Test)
		}
	}
}

// --- RunAll: SAVEPOINT isolation ---

func TestRunAll_SavepointIsolation(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	ctx := context.Background()
	// basic.test.lua includes test_db_insert_and_query followed by
	// test_savepoint_isolation which checks the insert was rolled back.
	report := h.RunAll(ctx, []string{"basic.test.lua"})

	for _, r := range report.Results {
		if r.Test == "test_savepoint_isolation" {
			if !r.Passed {
				var msgs []string
				for _, f := range r.Failures {
					msgs = append(msgs, f.Message)
				}
				t.Fatalf("savepoint isolation test failed: %s", strings.Join(msgs, "; "))
			}
			return
		}
	}
	t.Fatal("test_savepoint_isolation not found in results")
}

// --- Filter ---

func TestRunAll_Filter(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{
		Filter: "http",
	})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	ctx := context.Background()
	report := h.RunAll(ctx, []string{"basic.test.lua"})

	for _, r := range report.Results {
		if !strings.Contains(r.Test, "http") {
			t.Errorf("filter should exclude %s", r.Test)
		}
	}
	if report.Passed() == 0 {
		t.Fatal("expected at least one test matching filter 'http'")
	}
}

// --- Report ---

func TestReport_WriteHuman(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	ctx := context.Background()
	report := h.RunAll(ctx, []string{"basic.test.lua"})

	var buf bytes.Buffer
	report.WriteHuman(&buf, false)
	output := buf.String()

	if !strings.Contains(output, "harness_plugin") {
		t.Error("human output missing plugin name")
	}
	if !strings.Contains(output, "basic.test.lua") {
		t.Error("human output missing test file name")
	}
	if !strings.Contains(output, "passed") {
		t.Error("human output missing pass count")
	}
}

func TestReport_WriteHumanVerbose(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	ctx := context.Background()
	report := h.RunAll(ctx, []string{"failures.test.lua"})

	var buf bytes.Buffer
	report.WriteHuman(&buf, true)
	output := buf.String()

	if !strings.Contains(output, "FAIL:") {
		t.Error("verbose output should contain FAIL: lines")
	}
}

func TestReport_WriteNDJSON(t *stdtesting.T) {
	h, err := plugintesting.NewHarness(pluginDir(t, "harness_plugin"), plugintesting.HarnessOpts{})
	if err != nil {
		t.Fatalf("NewHarness: %v", err)
	}
	defer h.Close()

	ctx := context.Background()
	report := h.RunAll(ctx, []string{"basic.test.lua"})

	var buf bytes.Buffer
	if err := report.WriteNDJSON(&buf); err != nil {
		t.Fatalf("WriteNDJSON: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(report.Results) {
		t.Fatalf("expected %d NDJSON lines, got %d", len(report.Results), len(lines))
	}

	for i, line := range lines {
		var result plugintesting.TestResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			t.Fatalf("line %d: invalid JSON: %v", i, err)
		}
		if result.File != "basic.test.lua" {
			t.Errorf("line %d: expected file basic.test.lua, got %s", i, result.File)
		}
	}
}

// --- MockRequestEngine ---

func TestMockRequestEngine_MatchingRule(t *stdtesting.T) {
	m := &plugintesting.MockRequestEngine{}
	m.AddRule("GET", "api.example.com/data", map[string]any{
		"status": 200,
		"json":   map[string]any{"ok": true},
	})

	resp, err := m.Execute(context.Background(), "test", "GET", "https://api.example.com/data?q=1", plugintesting.MockOutboundOpts())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	status, ok := resp["status"]
	if !ok || status != 200 {
		t.Errorf("expected status 200, got %v", status)
	}
}

func TestMockRequestEngine_NoMatchingRule(t *stdtesting.T) {
	m := &plugintesting.MockRequestEngine{}

	_, err := m.Execute(context.Background(), "test", "GET", "https://api.example.com/nope", plugintesting.MockOutboundOpts())
	if err == nil {
		t.Fatal("expected error for unmatched request")
	}
	if !strings.Contains(err.Error(), "no rule matched") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockRequestEngine_ClearRules(t *stdtesting.T) {
	m := &plugintesting.MockRequestEngine{}
	m.AddRule("GET", "api.example.com/data", map[string]any{"status": 200})

	m.ClearRules()

	_, err := m.Execute(context.Background(), "test", "GET", "https://api.example.com/data", plugintesting.MockOutboundOpts())
	if err == nil {
		t.Fatal("expected error after ClearRules")
	}
}

func TestMockRequestEngine_MethodCaseInsensitive(t *stdtesting.T) {
	m := &plugintesting.MockRequestEngine{}
	m.AddRule("get", "api.example.com/data", map[string]any{"status": 200})

	resp, err := m.Execute(context.Background(), "test", "GET", "https://api.example.com/data", plugintesting.MockOutboundOpts())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp["status"] != 200 {
		t.Errorf("expected status 200, got %v", resp["status"])
	}
}

func TestMockRequestEngine_FirstRuleWins(t *stdtesting.T) {
	m := &plugintesting.MockRequestEngine{}
	m.AddRule("GET", "api.example.com/data", map[string]any{"status": 200})
	m.AddRule("GET", "api.example.com/data", map[string]any{"status": 404})

	resp, err := m.Execute(context.Background(), "test", "GET", "https://api.example.com/data", plugintesting.MockOutboundOpts())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp["status"] != 200 {
		t.Errorf("expected first rule (200) to win, got %v", resp["status"])
	}
}

func TestMockRequestEngine_JsonBody(t *stdtesting.T) {
	m := &plugintesting.MockRequestEngine{}
	m.AddRule("POST", "api.example.com/submit", map[string]any{
		"status": 201,
		"json":   map[string]any{"id": "abc"},
	})

	resp, err := m.Execute(context.Background(), "test", "POST", "https://api.example.com/submit", plugintesting.MockOutboundOpts())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	body, ok := resp["body"].(string)
	if !ok || body == "" {
		t.Error("expected JSON body to be serialized into body field")
	}
	if !strings.Contains(body, "abc") {
		t.Errorf("body should contain 'abc', got: %s", body)
	}
}

func TestMockRequestEngine_PlainBody(t *stdtesting.T) {
	m := &plugintesting.MockRequestEngine{}
	m.AddRule("GET", "api.example.com/text", map[string]any{
		"status": 200,
		"body":   "plain text response",
	})

	resp, err := m.Execute(context.Background(), "test", "GET", "https://api.example.com/text", plugintesting.MockOutboundOpts())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp["body"] != "plain text response" {
		t.Errorf("expected plain text body, got: %v", resp["body"])
	}
}
