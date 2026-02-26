package core

import (
	"context"
	"sync"

	"github.com/hegner123/modulacms/internal/db/types"
)

// CachedFetcher wraps a TreeFetcher with a per-request cache.
// Multiple goroutines within a single composition call can safely
// read and write concurrently. The cache is created per HTTP request
// — no cross-request sharing, no stale data.
type CachedFetcher struct {
	inner TreeFetcher
	mu    sync.Mutex
	cache map[types.ContentID]*Root
}

// NewCachedFetcher creates a CachedFetcher wrapping the given inner fetcher.
func NewCachedFetcher(inner TreeFetcher) *CachedFetcher {
	return &CachedFetcher{
		inner: inner,
		cache: make(map[types.ContentID]*Root),
	}
}

// FetchAndBuildTree returns a cached result if available, otherwise
// delegates to the inner fetcher and caches the result.
func (c *CachedFetcher) FetchAndBuildTree(ctx context.Context, id types.ContentID) (*Root, error) {
	c.mu.Lock()
	if root, ok := c.cache[id]; ok {
		c.mu.Unlock()
		return root, nil
	}
	c.mu.Unlock()

	root, err := c.inner.FetchAndBuildTree(ctx, id)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[id] = root
	c.mu.Unlock()
	return root, nil
}
