// Black-box tests for ParsePaginationParams and HasPaginationParams.
//
// These are pure functions that only depend on *http.Request query parameters,
// making them ideal for table-driven unit tests with no external dependencies.
package router_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/hegner123/modulacms/internal/router"
)

func TestParsePaginationParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		wantLimit  int64
		wantOffset int64
	}{
		{
			name:       "no params uses defaults",
			query:      "",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "limit only",
			query:      "limit=20",
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "offset only",
			query:      "offset=10",
			wantLimit:  50,
			wantOffset: 10,
		},
		{
			name:       "both params",
			query:      "limit=25&offset=100",
			wantLimit:  25,
			wantOffset: 100,
		},
		{
			name:       "limit at max boundary",
			query:      "limit=1000",
			wantLimit:  1000,
			wantOffset: 0,
		},
		{
			// Limits above 1000 are capped to 1000
			name:       "limit above max is capped",
			query:      "limit=1001",
			wantLimit:  1000,
			wantOffset: 0,
		},
		{
			name:       "very large limit is capped",
			query:      "limit=999999",
			wantLimit:  1000,
			wantOffset: 0,
		},
		{
			// Negative limit is ignored; default 50 is kept
			name:       "negative limit uses default",
			query:      "limit=-1",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			// Zero limit is ignored (n > 0 check); default 50 is kept
			name:       "zero limit uses default",
			query:      "limit=0",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			// Negative offset is ignored; default 0 is kept
			name:       "negative offset uses default",
			query:      "offset=-5",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			// Zero offset is valid (n >= 0 check)
			name:       "zero offset is valid",
			query:      "offset=0",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "non-numeric limit uses default",
			query:      "limit=abc",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "non-numeric offset uses default",
			query:      "offset=xyz",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "empty limit value uses default",
			query:      "limit=",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "empty offset value uses default",
			query:      "offset=",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			// Fractional values fail ParseInt and are ignored
			name:       "fractional limit uses default",
			query:      "limit=10.5",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "limit=1 is valid minimum",
			query:      "limit=1",
			wantLimit:  1,
			wantOffset: 0,
		},
		{
			// Extra unrelated params are ignored
			name:       "unrelated params are ignored",
			query:      "limit=30&offset=5&sort=name&order=asc",
			wantLimit:  30,
			wantOffset: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &http.Request{URL: &url.URL{RawQuery: tt.query}}
			got := router.ParsePaginationParams(r)

			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", got.Limit, tt.wantLimit)
			}
			if got.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", got.Offset, tt.wantOffset)
			}
		})
	}
}

func TestHasPaginationParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{
			name:  "no params",
			query: "",
			want:  false,
		},
		{
			name:  "limit only",
			query: "limit=10",
			want:  true,
		},
		{
			name:  "offset only",
			query: "offset=0",
			want:  true,
		},
		{
			name:  "both params",
			query: "limit=10&offset=5",
			want:  true,
		},
		{
			// Has("limit") is true even with empty value
			name:  "limit with empty value",
			query: "limit=",
			want:  true,
		},
		{
			// Has("offset") is true even with empty value
			name:  "offset with empty value",
			query: "offset=",
			want:  true,
		},
		{
			// Unrelated params do not trigger pagination detection
			name:  "unrelated params only",
			query: "sort=name&order=asc",
			want:  false,
		},
		{
			name:  "limit among other params",
			query: "sort=name&limit=10&order=asc",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &http.Request{URL: &url.URL{RawQuery: tt.query}}
			got := router.HasPaginationParams(r)

			if got != tt.want {
				t.Errorf("HasPaginationParams = %v, want %v", got, tt.want)
			}
		})
	}
}
