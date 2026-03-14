package search

import (
	"context"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/utility"
)

// indexEvent represents a search index mutation triggered by a publish or
// unpublish operation.
type indexEvent struct {
	kind      string // "publish" or "unpublish"
	snapshot  *publishing.Snapshot
	version   db.ContentVersion
	contentID string
}

// Service manages the search index lifecycle: loading from disk, accepting
// publish/unpublish events, and persisting changes with debounced writes.
type Service struct {
	index     *Index
	config    SearchConfig
	driver    db.DbDriver
	updates   chan indexEvent
	done      chan struct{}
	saveTimer *time.Timer
	stopped   bool
}

// NewService creates a Service that uses the given database driver and search
// configuration. Call Start to load the index and begin processing events.
func NewService(driver db.DbDriver, cfg SearchConfig) *Service {
	return &Service{
		config:  cfg,
		driver:  driver,
		updates: make(chan indexEvent, 256),
		done:    make(chan struct{}),
	}
}

// Start loads the search index from disk (or rebuilds it from the database)
// and starts the background event processing goroutine. Errors during load
// or rebuild are logged but not returned; the service falls back to an empty
// index so search remains available.
func (s *Service) Start(ctx context.Context) error {
	idx, loadErr := Load(s.config.IndexPath, s.config)
	if loadErr != nil {
		utility.DefaultLogger.Info("search index load failed, rebuilding", loadErr)
		var rebuildErr error
		idx, rebuildErr = s.rebuild()
		if rebuildErr != nil {
			utility.DefaultLogger.Error("search index rebuild failed, starting empty", rebuildErr)
			idx = NewIndex(s.config)
		}
	}
	s.index = idx

	go s.processLoop(ctx)

	return nil
}

// Stop shuts down the event processing goroutine and flushes the index to
// disk. It is safe to call multiple times.
func (s *Service) Stop() error {
	if s.stopped {
		return nil
	}
	s.stopped = true
	close(s.updates)
	<-s.done
	return nil
}

// processLoop reads index events from the updates channel and applies them.
// A 5-second debounce timer batches disk writes after bursts of updates.
func (s *Service) processLoop(ctx context.Context) {
	defer close(s.done)
	for {
		select {
		case evt, ok := <-s.updates:
			if !ok {
				// Channel closed (Stop called), flush remaining work.
				s.flushToDisk()
				return
			}
			s.handleEvent(evt)
			s.resetSaveTimer()
		case <-ctx.Done():
			// Context cancelled, drain pending events then flush.
			s.drainPending()
			s.flushToDisk()
			return
		}
	}
}

// handleEvent applies a single index mutation.
func (s *Service) handleEvent(evt indexEvent) {
	switch evt.kind {
	case "publish":
		docs := BuildDocuments(evt.snapshot, evt.version, s.config)
		contentDataID := ""
		if len(docs) > 0 {
			contentDataID = docs[0].ContentDataID
		}
		if contentDataID != "" {
			s.index.RemoveByContentID(contentDataID)
		}
		for _, doc := range docs {
			s.index.Add(doc)
		}
	case "unpublish":
		s.index.RemoveByContentID(evt.contentID)
	}
}

// resetSaveTimer schedules a disk flush 5 seconds from now, cancelling any
// previously scheduled flush.
func (s *Service) resetSaveTimer() {
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}
	s.saveTimer = time.AfterFunc(5*time.Second, func() {
		s.flushToDisk()
	})
}

// flushToDisk persists the current index to disk at the configured path.
func (s *Service) flushToDisk() {
	if s.index == nil {
		return
	}
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}
	if err := s.index.Save(s.config.IndexPath); err != nil {
		utility.DefaultLogger.Error("failed to save search index", err)
	}
}

// drainPending processes any events remaining in the updates channel buffer
// without blocking.
func (s *Service) drainPending() {
	for {
		select {
		case evt, ok := <-s.updates:
			if !ok {
				return
			}
			s.handleEvent(evt)
		default:
			return
		}
	}
}

// OnPublish enqueues a publish event for asynchronous indexing. The snapshot
// and version are used to build search documents for the published content.
func (s *Service) OnPublish(snapshot *publishing.Snapshot, version db.ContentVersion) {
	if s.stopped {
		return
	}
	s.updates <- indexEvent{
		kind:     "publish",
		snapshot: snapshot,
		version:  version,
	}
}

// OnUnpublish enqueues an unpublish event that removes all indexed documents
// for the given content data ID.
func (s *Service) OnUnpublish(contentDataID string) {
	if s.stopped {
		return
	}
	s.updates <- indexEvent{
		kind:      "unpublish",
		contentID: contentDataID,
	}
}

// Search performs a full-text search against the index. Returns an empty
// response if the index has not been loaded.
func (s *Service) Search(query string, opts SearchOptions) SearchResponse {
	if s.index == nil {
		return SearchResponse{Query: query}
	}
	return s.index.Search(query, opts)
}

// SearchWithPrefix performs a search where the last query term is treated as
// a prefix. Returns an empty response if the index has not been loaded.
func (s *Service) SearchWithPrefix(query string, opts SearchOptions) SearchResponse {
	if s.index == nil {
		return SearchResponse{Query: query}
	}
	return s.index.SearchWithPrefix(query, opts)
}

// Rebuild drops the current index and rebuilds it from the database. The
// rebuilt index is flushed to disk immediately.
func (s *Service) Rebuild() error {
	idx, err := s.rebuild()
	if err != nil {
		return err
	}
	s.index = idx
	s.flushToDisk()
	return nil
}

// rebuild creates a new index populated from all published content versions.
//
// TODO: When ListAllPublishedContentVersions is added to DbDriver, call it
// here to load all published snapshots and index them. For now this returns
// an empty index.
//
//	versions, err := s.driver.ListAllPublishedContentVersions(context.Background())
//	for _, v := range versions {
//	    snapshot := deserializeSnapshot(v.Snapshot)
//	    docs := BuildDocuments(snapshot, v, s.config)
//	    for _, doc := range docs { idx.Add(doc) }
//	}
func (s *Service) rebuild() (*Index, error) {
	idx := NewIndex(s.config)
	return idx, nil
}

// Stats returns summary statistics about the current search index.
func (s *Service) Stats() IndexStats {
	if s.index == nil {
		return IndexStats{}
	}
	return s.index.Stats()
}
