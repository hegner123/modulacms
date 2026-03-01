package query

// DefaultLimit is the default number of items per page.
const DefaultLimit int64 = 20

// MaxLimit is the maximum number of items per page.
const MaxLimit int64 = 100

// clampLimit ensures limit is within bounds.
func clampLimit(limit int64) int64 {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// clampOffset ensures offset is non-negative.
func clampOffset(offset int64) int64 {
	if offset < 0 {
		return 0
	}
	return offset
}

// paginate returns a slice of items for the given offset and limit.
// Returns the clamped limit, offset, and the sliced result.
func paginate(items []QueryItem, limit, offset int64) ([]QueryItem, int64, int64) {
	limit = clampLimit(limit)
	offset = clampOffset(offset)

	total := int64(len(items))
	if offset >= total {
		return nil, limit, offset
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return items[offset:end], limit, offset
}
