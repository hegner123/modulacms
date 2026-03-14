package search

import (
	"sort"
	"strings"
)

// ExtractSnippet finds the best excerpt from a SearchDocument for the given
// query terms. The returned snippet is at most maxLen characters and is
// centred on the densest cluster of matching terms.
func ExtractSnippet(doc SearchDocument, queryTerms []string, maxLen int) string {
	// Build concatenated plain text from all fields, alphabetical order.
	keys := make([]string, 0, len(doc.Fields))
	for k := range doc.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		stripped := StripHTML(doc.Fields[k])
		stripped = strings.TrimSpace(stripped)
		if stripped != "" {
			parts = append(parts, stripped)
		}
	}
	text := strings.Join(parts, " ... ")

	if text == "" {
		return ""
	}

	offsets := TokenizeWithOffsets(text)
	if len(offsets) == 0 {
		return truncateToWordBoundary(text, 0, maxLen)
	}

	// Build lowercase query term set.
	querySet := make(map[string]bool, len(queryTerms))
	for _, qt := range queryTerms {
		querySet[strings.ToLower(qt)] = true
	}

	if len(querySet) == 0 {
		return truncateToWordBoundary(text, 0, maxLen)
	}

	// Find the best window of terms fitting within maxLen characters.
	bestStart := 0
	bestEnd := 0
	bestScore := -1.0

	for i := range offsets {
		// Expand window from i forward while staying within maxLen chars.
		windowByteStart := offsets[i].Start
		distinct := make(map[string]bool)
		j := i
		for j < len(offsets) {
			windowByteEnd := offsets[j].End
			span := windowByteEnd - windowByteStart
			if span > maxLen && j > i {
				break
			}
			if querySet[offsets[j].Term] {
				distinct[offsets[j].Term] = true
			}
			j++
		}
		// j is now one past the last included term.
		lastIdx := j - 1
		spread := offsets[lastIdx].End - offsets[i].Start
		score := float64(len(distinct))*10.0 + 1.0/float64(spread+1)

		if score > bestScore {
			bestScore = score
			bestStart = i
			bestEnd = lastIdx
		}
	}

	// If no query terms were found at all, return from the beginning.
	foundAny := false
	for _, off := range offsets {
		if querySet[off.Term] {
			foundAny = true
			break
		}
	}
	if !foundAny {
		return truncateToWordBoundary(text, 0, maxLen)
	}

	// Determine byte range for the best window.
	winByteStart := offsets[bestStart].Start
	winByteEnd := offsets[bestEnd].End

	// Try to expand the window to use the full maxLen budget, centering on
	// the matched region.
	currentLen := winByteEnd - winByteStart
	remaining := maxLen - currentLen
	if remaining > 0 {
		expandBefore := remaining / 2
		expandAfter := remaining - expandBefore

		newStart := winByteStart - expandBefore
		if newStart < 0 {
			expandAfter += -newStart
			newStart = 0
		}
		newEnd := winByteEnd + expandAfter
		if newEnd > len(text) {
			leftover := newEnd - len(text)
			newEnd = len(text)
			newStart -= leftover
			if newStart < 0 {
				newStart = 0
			}
		}
		winByteStart = newStart
		winByteEnd = newEnd
	}

	// Expand to word boundaries (don't cut mid-word).
	winByteStart = expandToWordBoundaryLeft(text, winByteStart)
	winByteEnd = expandToWordBoundaryRight(text, winByteEnd)

	// Clamp.
	if winByteStart < 0 {
		winByteStart = 0
	}
	if winByteEnd > len(text) {
		winByteEnd = len(text)
	}

	snippet := text[winByteStart:winByteEnd]
	snippet = strings.TrimSpace(snippet)

	// Add ellipsis prefix/suffix.
	prefix := ""
	suffix := ""
	if winByteStart > 0 {
		prefix = "..."
	}
	if winByteEnd < len(text) {
		suffix = "..."
	}

	return prefix + snippet + suffix
}

// truncateToWordBoundary returns a snippet starting at byteStart of at most
// maxLen characters from text, trimmed to a word boundary, with ellipsis
// suffix if truncated.
func truncateToWordBoundary(text string, byteStart int, maxLen int) string {
	if byteStart >= len(text) {
		return ""
	}

	end := byteStart + maxLen
	if end >= len(text) {
		sub := text[byteStart:]
		sub = strings.TrimSpace(sub)
		if byteStart > 0 {
			return "..." + sub
		}
		return sub
	}

	// Walk back to a space boundary.
	end = expandToWordBoundaryLeft(text, end)
	if end <= byteStart {
		// If we can't find a word boundary, just take maxLen bytes.
		end = byteStart + maxLen
		if end > len(text) {
			end = len(text)
		}
	}

	sub := strings.TrimSpace(text[byteStart:end])
	prefix := ""
	suffix := ""
	if byteStart > 0 {
		prefix = "..."
	}
	if end < len(text) {
		suffix = "..."
	}
	return prefix + sub + suffix
}

// expandToWordBoundaryLeft moves pos leftward to the nearest whitespace
// boundary so we don't cut mid-word.
func expandToWordBoundaryLeft(text string, pos int) int {
	if pos <= 0 || pos >= len(text) {
		return pos
	}
	// If we're already at a word boundary (current char is space or previous
	// char is space), no adjustment needed.
	if text[pos] == ' ' || (pos > 0 && text[pos-1] == ' ') {
		return pos
	}
	// Walk backward to find a space.
	for i := pos; i > 0; i-- {
		if text[i-1] == ' ' {
			return i
		}
	}
	return 0
}

// expandToWordBoundaryRight moves pos rightward to the end of the current word.
func expandToWordBoundaryRight(text string, pos int) int {
	if pos >= len(text) {
		return len(text)
	}
	// If we're at a space, we're at a boundary.
	if text[pos] == ' ' {
		return pos
	}
	// Walk forward to find the next space or end.
	for i := pos; i < len(text); i++ {
		if text[i] == ' ' {
			return i
		}
	}
	return len(text)
}
