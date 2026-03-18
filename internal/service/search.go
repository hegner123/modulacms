package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/search"
)

// SearchService wraps the search.Service engine and provides validation,
// error mapping, and a consistent API for handlers.
type SearchService struct {
	engine *search.Service
}

// NewSearchService creates a SearchService wrapping the given search engine.
// Pass nil if search is disabled; all methods degrade gracefully.
func NewSearchService(engine *search.Service) *SearchService {
	return &SearchService{engine: engine}
}

// Available reports whether the search engine is initialized and ready.
func (s *SearchService) Available() bool {
	return s.engine != nil
}

// Search performs a full-text search. The last query term is treated as a
// prefix by default for typeahead behavior. Set prefix=false for exact matching.
func (s *SearchService) Search(ctx context.Context, query string, prefix bool, opts search.SearchOptions) (*search.SearchResponse, error) {
	if s.engine == nil {
		return nil, &InternalError{Err: fmt.Errorf("search is not enabled")}
	}

	if query == "" {
		ve := &ValidationError{}
		ve.Add("q", "query must not be empty")
		return nil, ve
	}

	if opts.Limit < 0 {
		ve := &ValidationError{}
		ve.Add("limit", "limit must not be negative")
		return nil, ve
	}

	if opts.Offset < 0 {
		ve := &ValidationError{}
		ve.Add("offset", "offset must not be negative")
		return nil, ve
	}

	var resp search.SearchResponse
	if prefix {
		resp = s.engine.SearchWithPrefix(query, opts)
	} else {
		resp = s.engine.Search(query, opts)
	}
	return &resp, nil
}

// Rebuild drops the current index and rebuilds it from the database.
func (s *SearchService) Rebuild(ctx context.Context) (*search.IndexStats, error) {
	if s.engine == nil {
		return nil, &InternalError{Err: fmt.Errorf("search is not enabled")}
	}

	if err := s.engine.Rebuild(); err != nil {
		return nil, fmt.Errorf("rebuild search index: %w", err)
	}

	stats := s.engine.Stats()
	return &stats, nil
}

// Stats returns summary statistics about the current search index.
func (s *SearchService) Stats(ctx context.Context) (*search.IndexStats, error) {
	if s.engine == nil {
		return nil, &InternalError{Err: fmt.Errorf("search is not enabled")}
	}

	stats := s.engine.Stats()
	return &stats, nil
}

// OnPublish forwards a publish event to the search engine for async indexing.
func (s *SearchService) OnPublish(snapshot *publishing.Snapshot, version db.ContentVersion) {
	if s.engine == nil {
		return
	}
	s.engine.OnPublish(snapshot, version)
}

// OnUnpublish forwards an unpublish event to remove content from the index.
func (s *SearchService) OnUnpublish(contentDataID string) {
	if s.engine == nil {
		return
	}
	s.engine.OnUnpublish(contentDataID)
}
