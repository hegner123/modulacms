package utility

import (
	"testing"
)

// ============================================================
// TestFuzzyScore — subsequence detection
// ============================================================

func TestFuzzyScore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		query  string
		target string
		wantOK bool
	}{
		{name: "exact match", query: "foo", target: "foo", wantOK: true},
		{name: "case insensitive", query: "FOO", target: "foo", wantOK: true},
		{name: "prefix match", query: "con", target: "Config", wantOK: true},
		{name: "scattered subsequence", query: "pgcnf", target: "PostgresConfig", wantOK: true},
		{name: "not a subsequence", query: "xyz", target: "abc", wantOK: false},
		{name: "empty query always matches", query: "", target: "anything", wantOK: true},
		{name: "query longer than target", query: "toolong", target: "too", wantOK: false},
		{name: "single char match", query: "a", target: "alpha", wantOK: true},
		{name: "single char no match", query: "z", target: "alpha", wantOK: false},
		{name: "both empty", query: "", target: "", wantOK: true},
		{name: "empty target non-empty query", query: "a", target: "", wantOK: false},
		{name: "unicode match", query: "caf", target: "Café Latte", wantOK: true},
		{name: "repeated chars", query: "aa", target: "abracadabra", wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, ok := FuzzyScore(tt.query, tt.target)
			if ok != tt.wantOK {
				t.Errorf("FuzzyScore(%q, %q) ok = %v, want %v", tt.query, tt.target, ok, tt.wantOK)
			}
		})
	}
}

// ============================================================
// TestFuzzyScoreRanking — scoring invariants (relative order)
// ============================================================

func TestFuzzyScoreRanking(t *testing.T) {
	t.Parallel()

	t.Run("exact beats prefix", func(t *testing.T) {
		t.Parallel()
		exact, _, okA := FuzzyScore("config", "config")
		prefix, _, okB := FuzzyScore("config", "configuration")
		if !okA || !okB {
			t.Fatal("both should match")
		}
		if exact <= prefix {
			t.Errorf("exact (%d) should beat prefix (%d)", exact, prefix)
		}
	})

	t.Run("prefix beats scattered", func(t *testing.T) {
		t.Parallel()
		prefix, _, okA := FuzzyScore("con", "Config")
		scattered, _, okB := FuzzyScore("con", "discontinuous")
		if !okA || !okB {
			t.Fatal("both should match")
		}
		if prefix <= scattered {
			t.Errorf("prefix (%d) should beat scattered (%d)", prefix, scattered)
		}
	})

	t.Run("consecutive beats spread", func(t *testing.T) {
		t.Parallel()
		consec, _, okA := FuzzyScore("abc", "abcxyz")
		spread, _, okB := FuzzyScore("abc", "axbxcx")
		if !okA || !okB {
			t.Fatal("both should match")
		}
		if consec <= spread {
			t.Errorf("consecutive (%d) should beat spread (%d)", consec, spread)
		}
	})

	t.Run("word boundary beats mid-word", func(t *testing.T) {
		t.Parallel()
		boundary, _, okA := FuzzyScore("mc", "MyConfig")
		midWord, _, okB := FuzzyScore("mc", "romcom")
		if !okA || !okB {
			t.Fatal("both should match")
		}
		if boundary <= midWord {
			t.Errorf("boundary (%d) should beat mid-word (%d)", boundary, midWord)
		}
	})

	t.Run("camelCase boundary", func(t *testing.T) {
		t.Parallel()
		camel, _, okA := FuzzyScore("pc", "PostgresConfig")
		noCamel, _, okB := FuzzyScore("pc", "alpacca")
		if !okA || !okB {
			t.Fatal("both should match")
		}
		if camel <= noCamel {
			t.Errorf("camelCase boundary (%d) should beat no boundary (%d)", camel, noCamel)
		}
	})

	t.Run("separator boundary", func(t *testing.T) {
		t.Parallel()
		sep, _, okA := FuzzyScore("pc", "postgres_config")
		noSep, _, okB := FuzzyScore("pc", "alpacca")
		if !okA || !okB {
			t.Fatal("both should match")
		}
		if sep <= noSep {
			t.Errorf("separator boundary (%d) should beat no boundary (%d)", sep, noSep)
		}
	})

	t.Run("first char bonus", func(t *testing.T) {
		t.Parallel()
		first, _, okA := FuzzyScore("a", "alpha")
		notFirst, _, okB := FuzzyScore("a", "beta")
		if !okA || !okB {
			t.Fatal("both should match")
		}
		if first <= notFirst {
			t.Errorf("first char (%d) should beat non-first (%d)", first, notFirst)
		}
	})
}

// ============================================================
// TestFuzzyScoreIndices — matched positions
// ============================================================

func TestFuzzyScoreIndices(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		query       string
		target      string
		wantIndices []int
	}{
		{name: "consecutive prefix", query: "abc", target: "abcdef", wantIndices: []int{0, 1, 2}},
		{name: "alternating", query: "ace", target: "abcde", wantIndices: []int{0, 2, 4}},
		{name: "skip first", query: "bd", target: "abcde", wantIndices: []int{1, 3}},
		{name: "empty query", query: "", target: "abc", wantIndices: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, matched, ok := FuzzyScore(tt.query, tt.target)
			if !ok {
				t.Fatalf("FuzzyScore(%q, %q) should match", tt.query, tt.target)
			}
			if len(matched) != len(tt.wantIndices) {
				t.Fatalf("matched len = %d, want %d", len(matched), len(tt.wantIndices))
			}
			for i, idx := range matched {
				if idx != tt.wantIndices[i] {
					t.Errorf("matched[%d] = %d, want %d", i, idx, tt.wantIndices[i])
				}
			}
		})
	}
}

// ============================================================
// TestFuzzyMatch — string slice matching
// ============================================================

func TestFuzzyMatch(t *testing.T) {
	t.Parallel()

	t.Run("empty query returns all in order", func(t *testing.T) {
		t.Parallel()
		candidates := []string{"bravo", "alpha", "charlie"}
		results := FuzzyMatch("", candidates)
		if len(results) != 3 {
			t.Fatalf("len = %d, want 3", len(results))
		}
		for i, r := range results {
			if r.Index != i {
				t.Errorf("results[%d].Index = %d, want %d", i, r.Index, i)
			}
			if r.Score != 0 {
				t.Errorf("results[%d].Score = %d, want 0", i, r.Score)
			}
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		t.Parallel()
		results := FuzzyMatch("xyz", []string{"alpha", "bravo"})
		if len(results) != 0 {
			t.Errorf("len = %d, want 0", len(results))
		}
	})

	t.Run("nil candidates returns empty", func(t *testing.T) {
		t.Parallel()
		results := FuzzyMatch("a", nil)
		if len(results) != 0 {
			t.Errorf("len = %d, want 0", len(results))
		}
	})

	t.Run("sorted by score descending", func(t *testing.T) {
		t.Parallel()
		candidates := []string{"discontinuous", "PostgresConfig", "Config"}
		results := FuzzyMatch("con", candidates)
		if len(results) != 3 {
			t.Fatalf("len = %d, want 3", len(results))
		}
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Errorf("results not sorted: [%d].Score=%d > [%d].Score=%d",
					i, results[i].Score, i-1, results[i-1].Score)
			}
		}
		// Config should rank first (prefix match)
		if results[0].Index != 2 {
			t.Errorf("best match Index = %d, want 2 (Config)", results[0].Index)
		}
	})
}

// ============================================================
// TestFuzzyFind — generic multi-field search
// ============================================================

func TestFuzzyFind(t *testing.T) {
	t.Parallel()

	type item struct {
		Name  string
		Label string
	}

	t.Run("multi-field picks best score", func(t *testing.T) {
		t.Parallel()
		items := []item{
			{Name: "xyz", Label: "PostgresConfig"},
		}
		results := FuzzyFind("pgcnf", items, func(it item) []string {
			return []string{it.Name, it.Label}
		})
		if len(results) != 1 {
			t.Fatalf("len = %d, want 1", len(results))
		}
		if results[0].Score <= 0 {
			t.Errorf("score = %d, want > 0", results[0].Score)
		}
	})

	t.Run("empty query returns all in order", func(t *testing.T) {
		t.Parallel()
		items := []item{
			{Name: "bravo", Label: "B"},
			{Name: "alpha", Label: "A"},
		}
		results := FuzzyFind("", items, func(it item) []string {
			return []string{it.Name, it.Label}
		})
		if len(results) != 2 {
			t.Fatalf("len = %d, want 2", len(results))
		}
		if results[0].Index != 0 || results[1].Index != 1 {
			t.Errorf("order not preserved: [0].Index=%d, [1].Index=%d", results[0].Index, results[1].Index)
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		t.Parallel()
		items := []item{{Name: "alpha", Label: "A"}}
		results := FuzzyFind("xyz", items, func(it item) []string {
			return []string{it.Name, it.Label}
		})
		if len(results) != 0 {
			t.Errorf("len = %d, want 0", len(results))
		}
	})

	t.Run("sorted by score descending", func(t *testing.T) {
		t.Parallel()
		items := []item{
			{Name: "discontinuous", Label: "Unrelated"},
			{Name: "conf", Label: "config File"},
			{Name: "postgres_config", Label: "DB Config"},
		}
		results := FuzzyFind("con", items, func(it item) []string {
			return []string{it.Name, it.Label}
		})
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Errorf("results not sorted: [%d].Score=%d > [%d].Score=%d",
					i, results[i].Score, i-1, results[i-1].Score)
			}
		}
	})
}
