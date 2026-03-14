package search

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// TermOffset holds a lowercased term and its byte offsets in the original text.
type TermOffset struct {
	Term  string
	Start int
	End   int
}

// StripHTML removes HTML tags and decodes common HTML entities using a
// character-by-character state machine. Tags are replaced with spaces.
func StripHTML(html string) string {
	var b strings.Builder
	b.Grow(len(html))

	inTag := false

	for i := 0; i < len(html); {
		ch := html[i]

		if inTag {
			if ch == '>' {
				inTag = false
				b.WriteByte(' ')
			}
			i++
			continue
		}

		if ch == '<' {
			inTag = true
			i++
			continue
		}

		if ch == '&' {
			entity, advance := decodeEntity(html[i:])
			if advance > 0 {
				b.WriteString(entity)
				i += advance
				continue
			}
		}

		// Normal character — may be multi-byte UTF-8
		_, size := utf8.DecodeRuneInString(html[i:])
		b.WriteString(html[i : i+size])
		i += size
	}

	return b.String()
}

// decodeEntity attempts to match a known HTML entity at the start of s.
// Returns the decoded string and the number of bytes consumed, or ("", 0)
// if no known entity matches.
func decodeEntity(s string) (string, int) {
	entities := [5]struct {
		encoded string
		decoded string
	}{
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
		{"&#39;", "'"},
	}

	for _, ent := range entities {
		if len(s) >= len(ent.encoded) && s[:len(ent.encoded)] == ent.encoded {
			return ent.decoded, len(ent.encoded)
		}
	}

	return "", 0
}

// Tokenize strips HTML, lowercases the result, then splits on non-letter,
// non-digit boundaries. Returns terms in order.
func Tokenize(text string) []string {
	plain := StripHTML(text)

	var terms []string
	var current strings.Builder

	for _, r := range strings.ToLower(plain) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				terms = append(terms, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		terms = append(terms, current.String())
	}

	return terms
}

// TokenizeAndFilter tokenizes text and removes stop words, returning the
// filtered terms and their original 0-based positions from the full
// Tokenize output.
func TokenizeAndFilter(text string, stopWords map[string]bool) (terms []string, positions []int) {
	all := Tokenize(text)

	for i, term := range all {
		if !stopWords[term] {
			terms = append(terms, term)
			positions = append(positions, i)
		}
	}

	return terms, positions
}

// TokenizeWithOffsets returns terms with their byte offsets in the original
// text. Each Term is lowercased; Start and End are byte positions in text.
func TokenizeWithOffsets(text string) []TermOffset {
	var offsets []TermOffset
	var current strings.Builder
	tokenStart := -1

	for i, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if tokenStart < 0 {
				tokenStart = i
			}
			current.WriteRune(unicode.ToLower(r))
		} else {
			if current.Len() > 0 {
				offsets = append(offsets, TermOffset{
					Term:  current.String(),
					Start: tokenStart,
					End:   i,
				})
				current.Reset()
				tokenStart = -1
			}
		}
	}

	if current.Len() > 0 {
		offsets = append(offsets, TermOffset{
			Term:  current.String(),
			Start: tokenStart,
			End:   len(text),
		})
	}

	return offsets
}
