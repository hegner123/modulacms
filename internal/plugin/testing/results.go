package testing

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Failure records a single assertion failure within a test function.
type Failure struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	IsError bool   `json:"error,omitempty"` // true for Lua runtime errors (not assertions)
}

// TestResult records the outcome of a single test_* function.
type TestResult struct {
	File       string    `json:"file"`
	Test       string    `json:"test"`
	Passed     bool      `json:"passed"`
	DurationMs int64     `json:"duration_ms"`
	Assertions int       `json:"assertions"`
	Failures   []Failure `json:"failures"`
}

// Report aggregates results across all test files.
type Report struct {
	PluginName string
	Results    []TestResult
	TotalTime  time.Duration
}

// Passed returns the count of passing tests.
func (r *Report) Passed() int {
	n := 0
	for _, tr := range r.Results {
		if tr.Passed {
			n++
		}
	}
	return n
}

// Failed returns the count of failing tests.
func (r *Report) Failed() int {
	n := 0
	for _, tr := range r.Results {
		if !tr.Passed {
			n++
		}
	}
	return n
}

// ExitCode returns 0 if all passed, 1 if any failed.
func (r *Report) ExitCode() int {
	if r.Failed() > 0 {
		return 1
	}
	return 0
}

// WriteHuman writes human-readable test output to w.
func (r *Report) WriteHuman(w io.Writer, verbose bool) {
	fmt.Fprintf(w, "%s plugin tests\n", r.PluginName)

	currentFile := ""
	for _, tr := range r.Results {
		if tr.File != currentFile {
			currentFile = tr.File
			fmt.Fprintf(w, "\n  %s\n", currentFile)
		}

		if tr.Passed {
			fmt.Fprintf(w, "    \u2713 %s (%dms)\n", tr.Test, tr.DurationMs)
			if verbose {
				for _, f := range tr.Failures {
					fmt.Fprintf(w, "        PASS: %s (%s:%d)\n", f.Message, tr.File, f.Line)
				}
			}
		} else {
			fmt.Fprintf(w, "    \u2717 %s (%dms)\n", tr.Test, tr.DurationMs)
			for _, f := range tr.Failures {
				if f.IsError {
					fmt.Fprintf(w, "        ERROR: %s (%s:%d)\n", f.Message, tr.File, f.Line)
				} else {
					fmt.Fprintf(w, "        FAIL: %s (%s:%d)\n", f.Message, tr.File, f.Line)
				}
			}
		}
	}

	fmt.Fprintf(w, "\n%d passed, %d failed (%dms)\n", r.Passed(), r.Failed(), r.TotalTime.Milliseconds())
}

// WriteNDJSON writes one JSON object per line per test result to w.
func (r *Report) WriteNDJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, tr := range r.Results {
		if err := enc.Encode(tr); err != nil {
			return fmt.Errorf("encode result: %w", err)
		}
	}
	return nil
}
