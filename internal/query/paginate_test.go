package query

import (
	"testing"
)

func TestClampLimit(t *testing.T) {
	tests := []struct {
		input int64
		want  int64
	}{
		{0, DefaultLimit},
		{-5, DefaultLimit},
		{10, 10},
		{100, 100},
		{200, MaxLimit},
	}
	for _, tt := range tests {
		got := clampLimit(tt.input)
		if got != tt.want {
			t.Errorf("clampLimit(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestClampOffset(t *testing.T) {
	tests := []struct {
		input int64
		want  int64
	}{
		{0, 0},
		{-1, 0},
		{5, 5},
	}
	for _, tt := range tests {
		got := clampOffset(tt.input)
		if got != tt.want {
			t.Errorf("clampOffset(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestPaginate(t *testing.T) {
	items := make([]QueryItem, 50)
	for i := range items {
		items[i] = QueryItem{Fields: map[string]string{}}
	}

	t.Run("normal", func(t *testing.T) {
		result, limit, offset := paginate(items, 10, 0)
		if len(result) != 10 {
			t.Errorf("len = %d, want 10", len(result))
		}
		if limit != 10 || offset != 0 {
			t.Errorf("limit=%d offset=%d", limit, offset)
		}
	})

	t.Run("offset beyond length", func(t *testing.T) {
		result, _, _ := paginate(items, 10, 100)
		if result != nil {
			t.Errorf("expected nil, got %d items", len(result))
		}
	})

	t.Run("end of data", func(t *testing.T) {
		result, _, _ := paginate(items, 10, 45)
		if len(result) != 5 {
			t.Errorf("len = %d, want 5", len(result))
		}
	})

	t.Run("zero limit uses default", func(t *testing.T) {
		result, limit, _ := paginate(items, 0, 0)
		if limit != DefaultLimit {
			t.Errorf("limit = %d, want %d", limit, DefaultLimit)
		}
		if int64(len(result)) != DefaultLimit {
			t.Errorf("len = %d, want %d", len(result), DefaultLimit)
		}
	})

	t.Run("max limit capping", func(t *testing.T) {
		_, limit, _ := paginate(items, 500, 0)
		if limit != MaxLimit {
			t.Errorf("limit = %d, want %d", limit, MaxLimit)
		}
	})
}
