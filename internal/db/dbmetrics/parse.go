package dbmetrics

import (
	"strings"
)

// QueryInfo holds the extracted operation and table name from a SQL query.
type QueryInfo struct {
	Operation string // "select", "insert", "update", "delete", "pragma", "create", "alter", "drop", "other"
	Table     string // primary table name, empty if unparseable
}

// ParseQuery extracts the SQL operation and primary table name from a raw query string.
// It uses token-based parsing (no regex). The parser is designed to be fast since it
// runs on every query.
func ParseQuery(raw string) QueryInfo {
	tokens := tokenize(raw)
	if len(tokens) == 0 {
		return QueryInfo{Operation: "other"}
	}

	// Uppercase the first token to determine operation
	first := strings.ToUpper(tokens[0])

	// Handle WITH (CTE): skip through to the terminal statement
	if first == "WITH" {
		tokens = skipCTE(tokens)
		if len(tokens) == 0 {
			return QueryInfo{Operation: "other"}
		}
		first = strings.ToUpper(tokens[0])
	}

	switch first {
	case "SELECT":
		return QueryInfo{
			Operation: "select",
			Table:     findTableAfterKeyword(tokens, "FROM"),
		}
	case "INSERT":
		return QueryInfo{
			Operation: "insert",
			Table:     findTableAfterKeyword(tokens, "INTO"),
		}
	case "UPDATE":
		return QueryInfo{
			Operation: "update",
			Table:     extractToken(tokens, 1),
		}
	case "DELETE":
		return QueryInfo{
			Operation: "delete",
			Table:     findTableAfterKeyword(tokens, "FROM"),
		}
	case "CREATE":
		return QueryInfo{
			Operation: "create",
			Table:     findTableAfterDDLKeyword(tokens),
		}
	case "ALTER":
		return QueryInfo{
			Operation: "alter",
			Table:     findTableAfterDDLKeyword(tokens),
		}
	case "DROP":
		return QueryInfo{
			Operation: "drop",
			Table:     findTableAfterDDLKeyword(tokens),
		}
	case "PRAGMA":
		return QueryInfo{
			Operation: "pragma",
			Table:     extractPragmaName(tokens),
		}
	default:
		return QueryInfo{Operation: "other"}
	}
}

// tokenize splits a SQL string into tokens, stripping comments and handling
// quoted identifiers. It is intentionally simple and fast.
func tokenize(raw string) []string {
	s := stripComments(raw)
	var tokens []string
	i := 0
	n := len(s)

	for i < n {
		// Skip whitespace
		if isSpace(s[i]) {
			i++
			continue
		}

		// Quoted identifier: backtick
		if s[i] == '`' {
			end := indexByte(s, '`', i+1)
			if end < 0 {
				end = n
			}
			tokens = append(tokens, s[i+1:end])
			i = end + 1
			continue
		}

		// Quoted identifier: double-quote
		if s[i] == '"' {
			end := indexByte(s, '"', i+1)
			if end < 0 {
				end = n
			}
			tokens = append(tokens, s[i+1:end])
			i = end + 1
			continue
		}

		// Single-quoted string literal: skip entirely (not a token we care about)
		if s[i] == '\'' {
			end := indexByte(s, '\'', i+1)
			if end < 0 {
				end = n
			}
			i = end + 1
			continue
		}

		// Parenthesized expression: skip as a single token (for things like
		// "table_info(users)" or "IF NOT EXISTS")
		if s[i] == '(' {
			depth := 1
			j := i + 1
			for j < n && depth > 0 {
				if s[j] == '(' {
					depth++
				} else if s[j] == ')' {
					depth--
				}
				j++
			}
			// Extract content inside parens as a token if it looks like a
			// function argument (e.g., PRAGMA table_info(users))
			inner := strings.TrimSpace(s[i+1 : j-1])
			if inner != "" {
				tokens = append(tokens, "("+inner+")")
			}
			i = j
			continue
		}

		// Regular token: delimited by whitespace, parens, or comma
		j := i
		for j < n && !isSpace(s[j]) && s[j] != '(' && s[j] != ')' && s[j] != ',' && s[j] != ';' {
			j++
		}
		if j > i {
			tokens = append(tokens, s[i:j])
		}
		i = j
		if i < n && (s[i] == ',' || s[i] == ';') {
			i++
		}
	}

	return tokens
}

// stripComments removes SQL line comments (--) and block comments (/* */).
func stripComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	n := len(s)

	for i < n {
		// Line comment
		if i+1 < n && s[i] == '-' && s[i+1] == '-' {
			for i < n && s[i] != '\n' {
				i++
			}
			if i < n {
				b.WriteByte(' ')
				i++ // skip \n
			}
			continue
		}

		// Block comment
		if i+1 < n && s[i] == '/' && s[i+1] == '*' {
			i += 2
			for i+1 < n {
				if s[i] == '*' && s[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			if i >= n {
				break
			}
			b.WriteByte(' ')
			continue
		}

		b.WriteByte(s[i])
		i++
	}

	return b.String()
}

// findTableAfterKeyword finds the first token after the given keyword (case-insensitive).
func findTableAfterKeyword(tokens []string, keyword string) string {
	for i, t := range tokens {
		if strings.EqualFold(t, keyword) && i+1 < len(tokens) {
			return cleanTableName(tokens[i+1])
		}
	}
	return ""
}

// findTableAfterDDLKeyword finds the table name after TABLE or INDEX in DDL statements.
// Handles patterns like "CREATE TABLE IF NOT EXISTS foo" and "CREATE INDEX idx ON foo".
func findTableAfterDDLKeyword(tokens []string) string {
	for i, t := range tokens {
		upper := strings.ToUpper(t)
		if upper == "TABLE" || upper == "INDEX" {
			// Skip optional IF NOT EXISTS
			j := i + 1
			for j < len(tokens) && isReservedDDLWord(tokens[j]) {
				j++
			}
			if j < len(tokens) {
				name := cleanTableName(tokens[j])
				// For INDEX, if the next meaningful word is ON, the table is after ON
				if upper == "INDEX" {
					// tokens[j] is the index name; look for ON
					for k := j + 1; k < len(tokens); k++ {
						if strings.EqualFold(tokens[k], "ON") && k+1 < len(tokens) {
							return cleanTableName(tokens[k+1])
						}
					}
					return name
				}
				return name
			}
		}
	}
	return ""
}

// extractPragmaName extracts the pragma name from tokens like ["PRAGMA", "table_info(users)"].
func extractPragmaName(tokens []string) string {
	if len(tokens) < 2 {
		return ""
	}
	name := tokens[1]
	// Remove parenthesized arguments: "table_info(users)" -> "table_info"
	if idx := strings.IndexByte(name, '('); idx >= 0 {
		name = name[:idx]
	}
	// Handle case where pragma name and parens are separate tokens: ["PRAGMA", "table_info", "(users)"]
	return cleanTableName(name)
}

// extractToken returns the cleaned token at index, or "" if out of bounds.
func extractToken(tokens []string, index int) string {
	if index >= len(tokens) {
		return ""
	}
	return cleanTableName(tokens[index])
}

// cleanTableName strips schema prefixes, alias suffixes, and quote characters
// from a raw table token.
func cleanTableName(raw string) string {
	// Remove surrounding parens if present (from tokenizer)
	if len(raw) > 1 && raw[0] == '(' && raw[len(raw)-1] == ')' {
		raw = raw[1 : len(raw)-1]
	}
	// Strip schema prefix (schema.table -> table)
	if dotIdx := strings.LastIndexByte(raw, '.'); dotIdx >= 0 {
		raw = raw[dotIdx+1:]
	}
	// Strip backticks and double-quotes (already handled by tokenizer in most
	// cases, but be defensive)
	raw = strings.Trim(raw, "`\"")
	// Return lowercase
	return strings.ToLower(raw)
}

// skipCTE skips through a WITH ... AS (...) clause to find the terminal statement.
func skipCTE(tokens []string) []string {
	// WITH name AS (...), name AS (...) SELECT/INSERT/UPDATE/DELETE ...
	i := 1
	for i < len(tokens) {
		upper := strings.ToUpper(tokens[i])
		switch upper {
		case "SELECT", "INSERT", "UPDATE", "DELETE":
			return tokens[i:]
		}
		i++
	}
	return nil
}

// isReservedDDLWord returns true for words that appear between TABLE/INDEX and the
// actual name in DDL statements (IF, NOT, EXISTS).
func isReservedDDLWord(token string) bool {
	upper := strings.ToUpper(token)
	return upper == "IF" || upper == "NOT" || upper == "EXISTS"
}

// isSpace returns true for ASCII whitespace characters.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// indexByte returns the index of the first occurrence of c in s starting at start,
// or -1 if not found.
func indexByte(s string, c byte, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
