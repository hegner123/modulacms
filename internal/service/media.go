package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	_ "golang.org/x/image/webp"

	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/utility"
)

// ReprocessStatus tracks the progress of a bulk media reprocess job.
type ReprocessStatus struct {
	Running   bool      `json:"running"`
	Total     int       `json:"total"`
	Completed int       `json:"completed"`
	Failed    int       `json:"failed"`
	StartedAt time.Time `json:"started_at,omitempty"`
}

// MediaService manages media upload, metadata, health checks,
// orphan cleanup, and dimension presets.
type MediaService struct {
	driver db.DbDriver
	mgr    *config.Manager
	ctx    context.Context // service-scoped context, canceled on graceful shutdown

	mu              sync.RWMutex
	reprocessStatus ReprocessStatus
	restartQueued   bool
}

// NewMediaService creates a MediaService with the given dependencies.
// The ctx parameter is a service-scoped context derived from the server's root
// context, canceled on graceful shutdown. It is used for background goroutines
// (e.g., bulk media reprocessing) that must not depend on HTTP request contexts.
func NewMediaService(ctx context.Context, driver db.DbDriver, mgr *config.Manager) *MediaService {
	return &MediaService{ctx: ctx, driver: driver, mgr: mgr}
}

// UploadMediaParams holds inputs for uploading a new media file.
type UploadMediaParams struct {
	File        multipart.File
	Header      *multipart.FileHeader
	Path        string // optional S3 key prefix
	Alt         string
	Caption     string
	Description string
	DisplayName string
	FolderID    types.NullableMediaFolderID
}

// UpdateMediaMetadataParams holds inputs for updating media metadata.
type UpdateMediaMetadataParams struct {
	MediaID     types.MediaID         `json:"media_id"`
	DisplayName string                `json:"display_name"`
	Alt         string                `json:"alt"`
	Caption     string                `json:"caption"`
	Description string                `json:"description"`
	FocalX      types.NullableFloat64 `json:"focal_x"`
	FocalY      types.NullableFloat64 `json:"focal_y"`
}

// OrphanScanResult holds the result of an orphan scan.
type OrphanScanResult struct {
	TotalObjects int      `json:"total_objects"`
	TrackedKeys  int      `json:"tracked_keys"`
	OrphanedKeys []string `json:"orphaned_keys"`
	OrphanCount  int      `json:"orphan_count"`
}

// OrphanCleanupResult holds the result of orphan cleanup.
type OrphanCleanupResult struct {
	Deleted      []string `json:"deleted"`
	DeletedCount int      `json:"deleted_count"`
	Failed       []string `json:"failed"`
	FailedCount  int      `json:"failed_count"`
}

// Upload validates S3 config, constructs S3 closures, and delegates to
// media.ProcessMediaUpload for the full upload pipeline.
func (m *MediaService) Upload(ctx context.Context, ac audited.AuditContext, params UploadMediaParams) (*db.Media, error) {
	cfg, err := m.mgr.Config()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("load config: %w", err)}
	}

	// Validate S3 config is present
	if cfg.Bucket_Access_Key == "" || cfg.Bucket_Secret_Key == "" {
		return nil, NewValidationError("s3", "S3 storage must be configured for media uploads")
	}

	// Sanitize path
	mediaPath, pathErr := media.SanitizeMediaPath(params.Path)
	if pathErr != nil {
		return nil, NewValidationError("path", pathErr.Error())
	}

	// Create S3 session
	s3Session, err := newMediaS3Session(*cfg)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("S3 session: %w", err)}
	}

	acl := cfg.Bucket_Default_ACL
	if acl == "" {
		acl = "public-read"
	}
	bucketDir := cfg.Bucket_Media

	uploadOriginal := func(filePath string) (string, string, error) {
		f, fErr := os.Open(filePath)
		if fErr != nil {
			return "", "", fmt.Errorf("open file for S3 upload: %w", fErr)
		}
		defer f.Close()

		filename := filepath.Base(filePath)
		s3Key := fmt.Sprintf("%s/%s", mediaPath, filename)
		uploadURL := fmt.Sprintf("%s/%s/%s", cfg.BucketPublicURL(), bucketDir, s3Key)

		prep, prepErr := bucket.UploadPrep(s3Key, cfg.Bucket_Media, f, acl)
		if prepErr != nil {
			return "", "", fmt.Errorf("upload prep: %w", prepErr)
		}

		_, upErr := bucket.ObjectUpload(s3Session, prep)
		if upErr != nil {
			return "", "", fmt.Errorf("S3 upload: %w", upErr)
		}

		return uploadURL, s3Key, nil
	}

	rollbackS3 := func(s3Key string) {
		_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(cfg.Bucket_Media),
			Key:    aws.String(s3Key),
		})
		if delErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("rollback failed for key %s", s3Key), delErr)
		} else {
			utility.DefaultLogger.Info(fmt.Sprintf("rolled back S3 upload: %s", s3Key))
		}
	}

	pipeline := func(srcFile string, dstPath string) error {
		return media.HandleMediaUpload(srcFile, dstPath, *cfg)
	}

	row, err := media.ProcessMediaUpload(ctx, ac, params.File, params.Header, m.driver, uploadOriginal, rollbackS3, pipeline, cfg.MaxUploadSize(), params.FolderID)
	if err != nil {
		var dupErr media.DuplicateMediaError
		var sizeErr media.FileTooLargeError

		switch {
		case errors.As(err, &dupErr):
			return nil, &ConflictError{Resource: "media", Detail: dupErr.Error()}
		case errors.As(err, &sizeErr):
			return nil, NewValidationError("file", sizeErr.Error())
		default:
			return nil, fmt.Errorf("create media: %w", err)
		}
	}

	return row, nil
}

// GetMedia retrieves a media record by ID with NotFoundError mapping.
func (m *MediaService) GetMedia(ctx context.Context, id types.MediaID) (*db.Media, error) {
	record, err := m.driver.GetMedia(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "media", ID: string(id)}
	}
	return record, nil
}

// ListMedia returns all media records.
func (m *MediaService) ListMedia(ctx context.Context) (*[]db.Media, error) {
	return m.driver.ListMedia()
}

// ListMediaFull returns all media records with embedded author views.
func (m *MediaService) ListMediaFull(ctx context.Context) ([]db.MediaFullView, error) {
	items, err := m.driver.ListMedia()
	if err != nil {
		return nil, fmt.Errorf("list media: %w", err)
	}
	return db.AssembleMediaFullListView(m.driver, *items), nil
}

// ListMediaPaginated returns media records with pagination.
func (m *MediaService) ListMediaPaginated(ctx context.Context, limit, offset int64) (*[]db.Media, *int64, error) {
	items, err := m.driver.ListMediaPaginated(db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, nil, fmt.Errorf("list media paginated: %w", err)
	}
	total, err := m.driver.CountMedia()
	if err != nil {
		return nil, nil, fmt.Errorf("count media: %w", err)
	}
	return items, total, nil
}

// UpdateMediaMetadata fetches the existing record, overlays non-empty fields,
// validates focal point range, sets DateModified, and updates.
func (m *MediaService) UpdateMediaMetadata(ctx context.Context, ac audited.AuditContext, params UpdateMediaMetadataParams) (*db.Media, error) {
	existing, err := m.driver.GetMedia(params.MediaID)
	if err != nil {
		return nil, &NotFoundError{Resource: "media", ID: string(params.MediaID)}
	}

	// Validate focal point range [0,1]
	ve := &ValidationError{}
	if params.FocalX.Valid && (params.FocalX.Float64 < 0 || params.FocalX.Float64 > 1) {
		ve.Add("focal_x", "focal point X must be between 0 and 1")
	}
	if params.FocalY.Valid && (params.FocalY.Float64 < 0 || params.FocalY.Float64 > 1) {
		ve.Add("focal_y", "focal point Y must be between 0 and 1")
	}
	if ve.HasErrors() {
		return nil, ve
	}

	dbParams := db.UpdateMediaParams{
		MediaID:      existing.MediaID,
		Name:         existing.Name,
		DisplayName:  overlayNullString(params.DisplayName, existing.DisplayName),
		Alt:          overlayNullString(params.Alt, existing.Alt),
		Caption:      overlayNullString(params.Caption, existing.Caption),
		Description:  overlayNullString(params.Description, existing.Description),
		Class:        existing.Class,
		Mimetype:     existing.Mimetype,
		Dimensions:   existing.Dimensions,
		URL:          existing.URL,
		Srcset:       existing.Srcset,
		FocalX:       overlayFocal(params.FocalX, existing.FocalX),
		FocalY:       overlayFocal(params.FocalY, existing.FocalY),
		AuthorID:     existing.AuthorID,
		DateCreated:  existing.DateCreated,
		DateModified: types.NewTimestamp(time.Now().UTC()),
	}

	_, err = m.driver.UpdateMedia(ctx, ac, dbParams)
	if err != nil {
		return nil, fmt.Errorf("update media: %w", err)
	}

	updated, err := m.driver.GetMedia(params.MediaID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated media: %w", err)
	}
	return updated, nil
}

// DeleteMedia fetches the record, extracts S3 keys from URL+srcset, deletes
// S3 objects (best-effort), then deletes the DB record.
// Does NOT handle reference cleanup (clean_refs) — that stays in the handler layer.
func (m *MediaService) DeleteMedia(ctx context.Context, ac audited.AuditContext, id types.MediaID) error {
	record, err := m.driver.GetMedia(id)
	if err != nil {
		return &NotFoundError{Resource: "media", ID: string(id)}
	}

	cfg, err := m.mgr.Config()
	if err != nil {
		return &InternalError{Err: fmt.Errorf("load config: %w", err)}
	}

	// Extract and delete S3 objects
	s3Keys := extractMediaS3Keys(record, *cfg)
	if len(s3Keys) > 0 {
		s3Session, s3Err := newMediaS3Session(*cfg)
		if s3Err != nil {
			utility.DefaultLogger.Warn("S3 session for media delete failed, proceeding with DB delete", s3Err)
		} else {
			for _, key := range s3Keys {
				_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
					Bucket: aws.String(cfg.Bucket_Media),
					Key:    aws.String(key),
				})
				if delErr != nil {
					utility.DefaultLogger.Warn("failed to delete S3 object", delErr, "key", key)
				}
			}
		}
	}

	// Delete DB record
	if err := m.driver.DeleteMedia(ctx, ac, id); err != nil {
		return fmt.Errorf("delete media record: %w", err)
	}
	return nil
}

// MediaHealth scans for orphaned S3 objects that have no corresponding DB record.
func (m *MediaService) MediaHealth(ctx context.Context) (*OrphanScanResult, error) {
	cfg, err := m.mgr.Config()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("load config: %w", err)}
	}

	s3Session, err := newMediaS3Session(*cfg)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("S3 session: %w", err)}
	}

	return findOrphanedMediaKeys(m.driver, s3Session, *cfg)
}

// MediaCleanup scans for and deletes orphaned S3 objects.
func (m *MediaService) MediaCleanup(ctx context.Context) (*OrphanCleanupResult, error) {
	cfg, err := m.mgr.Config()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("load config: %w", err)}
	}

	s3Session, err := newMediaS3Session(*cfg)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("S3 session: %w", err)}
	}

	scanResult, err := findOrphanedMediaKeys(m.driver, s3Session, *cfg)
	if err != nil {
		return nil, err
	}

	var deleted, failed []string
	for _, key := range scanResult.OrphanedKeys {
		_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(cfg.Bucket_Media),
			Key:    aws.String(key),
		})
		if delErr != nil {
			utility.DefaultLogger.Warn("failed to delete orphaned object", delErr, "key", key)
			failed = append(failed, key)
		} else {
			deleted = append(deleted, key)
		}
	}

	return &OrphanCleanupResult{
		Deleted:      deleted,
		DeletedCount: len(deleted),
		Failed:       failed,
		FailedCount:  len(failed),
	}, nil
}

// --- Media Dimensions ---

// ListMediaDimensions returns all media dimension presets.
func (m *MediaService) ListMediaDimensions(ctx context.Context) (*[]db.MediaDimensions, error) {
	return m.driver.ListMediaDimensions()
}

// GetMediaDimension retrieves a media dimension by ID with NotFoundError mapping.
func (m *MediaService) GetMediaDimension(ctx context.Context, id string) (*db.MediaDimensions, error) {
	dim, err := m.driver.GetMediaDimension(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "media_dimension", ID: id}
	}
	return dim, nil
}

// CreateMediaDimension validates inputs and creates a new media dimension preset.
func (m *MediaService) CreateMediaDimension(ctx context.Context, ac audited.AuditContext, params db.CreateMediaDimensionParams) (*db.MediaDimensions, error) {
	ve := &ValidationError{}
	if !params.Width.Valid || params.Width.Int64 <= 0 {
		ve.Add("width", "width must be positive")
	}
	if !params.Height.Valid || params.Height.Int64 <= 0 {
		ve.Add("height", "height must be positive")
	}
	if !params.Label.Valid || params.Label.String == "" {
		ve.Add("label", "label is required")
	}
	if ve.HasErrors() {
		return nil, ve
	}

	created, err := m.driver.CreateMediaDimension(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create media dimension: %w", err)
	}
	return created, nil
}

// UpdateMediaDimension validates inputs and updates a media dimension preset.
func (m *MediaService) UpdateMediaDimension(ctx context.Context, ac audited.AuditContext, params db.UpdateMediaDimensionParams) (*db.MediaDimensions, error) {
	ve := &ValidationError{}
	if !params.Width.Valid || params.Width.Int64 <= 0 {
		ve.Add("width", "width must be positive")
	}
	if !params.Height.Valid || params.Height.Int64 <= 0 {
		ve.Add("height", "height must be positive")
	}
	if ve.HasErrors() {
		return nil, ve
	}

	_, err := m.driver.UpdateMediaDimension(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("update media dimension: %w", err)
	}

	updated, err := m.driver.GetMediaDimension(params.MdID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated media dimension: %w", err)
	}
	return updated, nil
}

// DeleteMediaDimension deletes a media dimension preset.
func (m *MediaService) DeleteMediaDimension(ctx context.Context, ac audited.AuditContext, id string) error {
	if err := m.driver.DeleteMediaDimension(ctx, ac, id); err != nil {
		return fmt.Errorf("delete media dimension: %w", err)
	}
	return nil
}

// GetReprocessStatus returns the current bulk reprocess job status.
func (m *MediaService) GetReprocessStatus() ReprocessStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.reprocessStatus
}

// TriggerReprocess starts a background bulk reprocess of all media variants.
// If a reprocess is already running, it queues a restart after the current run
// finishes. Returns true if a new run was started, false if queued.
func (m *MediaService) TriggerReprocess() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.reprocessStatus.Running {
		m.restartQueued = true
		return false
	}

	go m.runReprocessAll()
	return true
}

// runReprocessAll executes the bulk reprocess loop. It runs in a background
// goroutine using the service-scoped context. After each run, if restartQueued
// is set, it resets and starts another run to pick up dimension changes.
func (m *MediaService) runReprocessAll() {
	for {
		m.executeReprocessRun()

		m.mu.Lock()
		if m.restartQueued {
			m.restartQueued = false
			m.mu.Unlock()
			continue
		}
		m.mu.Unlock()
		return
	}
}

// executeReprocessRun performs a single pass over all image media records,
// reprocessing their variants with current dimension presets.
func (m *MediaService) executeReprocessRun() {
	// System-level audit context (no HTTP request user).
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "media-reprocess-worker", "127.0.0.1")

	allMedia, err := m.driver.ListMedia()
	if err != nil {
		utility.DefaultLogger.Error("reprocess: failed to list media", err)
		return
	}
	if allMedia == nil {
		return
	}

	// Filter to images only
	var imageMedia []db.Media
	for _, record := range *allMedia {
		if media.IsImageMIME(record.Mimetype.String) {
			imageMedia = append(imageMedia, record)
		}
	}

	m.mu.Lock()
	m.reprocessStatus = ReprocessStatus{
		Running:   true,
		Total:     len(imageMedia),
		Completed: 0,
		Failed:    0,
		StartedAt: time.Now().UTC(),
	}
	m.mu.Unlock()

	for _, record := range imageMedia {
		// Check for graceful shutdown between records
		select {
		case <-m.ctx.Done():
			m.mu.Lock()
			m.reprocessStatus.Running = false
			m.mu.Unlock()
			utility.DefaultLogger.Info("reprocess: stopped by shutdown", "completed", m.reprocessStatus.Completed)
			return
		default:
		}

		if err := m.ReprocessMediaVariants(m.ctx, ac, record.MediaID); err != nil {
			utility.DefaultLogger.Warn("reprocess: failed for media", err, "media_id", string(record.MediaID))
			m.mu.Lock()
			m.reprocessStatus.Failed++
			m.mu.Unlock()
		}

		m.mu.Lock()
		m.reprocessStatus.Completed++
		m.mu.Unlock()
	}

	m.mu.Lock()
	m.reprocessStatus.Running = false
	m.mu.Unlock()

	utility.DefaultLogger.Info("reprocess: bulk run complete",
		"total", len(imageMedia),
		"completed", m.reprocessStatus.Completed,
		"failed", m.reprocessStatus.Failed)
}

// ReprocessMediaVariants re-generates cropped image variants using the current
// focal point and re-uploads them to S3, then updates the media record's srcset.
// Non-image media returns nil immediately.
func (m *MediaService) ReprocessMediaVariants(ctx context.Context, ac audited.AuditContext, mediaID types.MediaID) error {
	record, err := m.driver.GetMedia(mediaID)
	if err != nil {
		return &NotFoundError{Resource: "media", ID: string(mediaID)}
	}

	// Only process images
	if !media.IsImageMIME(record.Mimetype.String) {
		return nil
	}

	cfg, err := m.mgr.Config()
	if err != nil {
		return &InternalError{Err: fmt.Errorf("load config: %w", err)}
	}

	s3Session, err := newMediaS3Session(*cfg)
	if err != nil {
		return &InternalError{Err: fmt.Errorf("S3 session: %w", err)}
	}

	// Extract the S3 key from the media URL
	endpointPrefix := cfg.BucketPublicURL() + "/" + cfg.Bucket_Media + "/"
	originalKey := strings.TrimPrefix(string(record.URL), endpointPrefix)
	if originalKey == "" || originalKey == string(record.URL) {
		return fmt.Errorf("reprocess: could not extract S3 key from URL %s", record.URL)
	}

	// Create temp directory for downloaded original and generated variants
	tmpDir, err := os.MkdirTemp("", media.TempDirPrefix)
	if err != nil {
		return fmt.Errorf("reprocess: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the original file from S3
	resp, err := s3Session.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket_Media),
		Key:    aws.String(originalKey),
	})
	if err != nil {
		return fmt.Errorf("reprocess: download original from S3: %w", err)
	}
	defer resp.Body.Close()

	originalFilename := filepath.Base(originalKey)
	localPath := filepath.Join(tmpDir, originalFilename)
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("reprocess: create local file: %w", err)
	}
	if _, err := io.Copy(localFile, resp.Body); err != nil {
		localFile.Close()
		return fmt.Errorf("reprocess: write local file: %w", err)
	}
	if err := localFile.Close(); err != nil {
		return fmt.Errorf("reprocess: close local file: %w", err)
	}

	// Decode image headers to get bounds for focal point conversion
	var focalPoint *image.Point
	if record.FocalX.Valid && record.FocalY.Valid {
		headerFile, openErr := os.Open(localPath)
		if openErr != nil {
			return fmt.Errorf("reprocess: open for header decode: %w", openErr)
		}
		imgConfig, _, decErr := image.DecodeConfig(headerFile)
		headerFile.Close()
		if decErr != nil {
			return fmt.Errorf("reprocess: decode image config: %w", decErr)
		}
		imgBounds := image.Rect(0, 0, imgConfig.Width, imgConfig.Height)
		focalPoint = media.FocalPointToPixels(record.FocalX, record.FocalY, imgBounds)
	}

	// Generate optimized variants
	optimized, err := media.OptimizeUpload(localPath, tmpDir, m.driver, focalPoint)
	if err != nil {
		return fmt.Errorf("reprocess: optimization failed: %w", err)
	}
	if optimized == nil || len(*optimized) == 0 {
		return nil // no variants generated (source too small or no dimensions configured)
	}

	// Delete old variant S3 objects (from existing srcset)
	oldKeys := extractMediaS3Keys(record, *cfg)
	// Remove the original key — we only want to delete variants, not the original
	originalKeyNorm := originalKey
	for _, oldKey := range oldKeys {
		if oldKey == originalKeyNorm {
			continue
		}
		_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(cfg.Bucket_Media),
			Key:    aws.String(oldKey),
		})
		if delErr != nil {
			utility.DefaultLogger.Warn("reprocess: failed to delete old variant", delErr, "key", oldKey)
		}
	}

	// Upload new variants to S3
	mediaPath := media.MediaPathFromURL(string(record.URL), endpointPrefix)
	acl := cfg.Bucket_Default_ACL
	if acl == "" {
		acl = "public-read"
	}

	srcset := []string{}
	uploadedKeys := []string{}

	for _, fullPath := range *optimized {
		f, openErr := os.Open(fullPath)
		if openErr != nil {
			rollbackVariantUploads(s3Session, cfg.Bucket_Media, uploadedKeys)
			return fmt.Errorf("reprocess: open variant file: %w", openErr)
		}

		variantFilename := filepath.Base(fullPath)
		s3Key := fmt.Sprintf("%s/%s", mediaPath, variantFilename)
		uploadURL := fmt.Sprintf("%s/%s/%s", cfg.BucketPublicURL(), cfg.Bucket_Media, s3Key)

		prep, prepErr := bucket.UploadPrep(s3Key, cfg.Bucket_Media, f, acl)
		if prepErr != nil {
			f.Close()
			rollbackVariantUploads(s3Session, cfg.Bucket_Media, uploadedKeys)
			return fmt.Errorf("reprocess: upload prep: %w", prepErr)
		}

		_, upErr := bucket.ObjectUpload(s3Session, prep)
		f.Close()
		if upErr != nil {
			rollbackVariantUploads(s3Session, cfg.Bucket_Media, uploadedKeys)
			return fmt.Errorf("reprocess: S3 upload: %w", upErr)
		}

		uploadedKeys = append(uploadedKeys, s3Key)
		srcset = append(srcset, uploadURL)
	}

	// Update srcset in database
	srcsetJSON, err := json.Marshal(srcset)
	if err != nil {
		rollbackVariantUploads(s3Session, cfg.Bucket_Media, uploadedKeys)
		return fmt.Errorf("reprocess: marshal srcset: %w", err)
	}

	dbParams := media.MapMediaParams(*record)
	dbParams.Srcset = db.NewNullString(string(srcsetJSON))

	if _, err := m.driver.UpdateMedia(ctx, ac, dbParams); err != nil {
		rollbackVariantUploads(s3Session, cfg.Bucket_Media, uploadedKeys)
		return fmt.Errorf("reprocess: update media record: %w", err)
	}

	return nil
}

// ReprocessAdminMediaVariants re-generates cropped image variants for an admin
// media record using the current focal point and re-uploads them to S3, then
// updates the admin media record's srcset. Non-image media returns nil immediately.
func (m *MediaService) ReprocessAdminMediaVariants(ctx context.Context, ac audited.AuditContext, adminMediaID types.AdminMediaID) error {
	record, err := m.driver.GetAdminMedia(adminMediaID)
	if err != nil {
		return &NotFoundError{Resource: "admin_media", ID: string(adminMediaID)}
	}

	if !media.IsImageMIME(record.Mimetype.String) {
		return nil
	}

	cfg, err := m.mgr.Config()
	if err != nil {
		return &InternalError{Err: fmt.Errorf("load config: %w", err)}
	}

	// Admin media uses its own bucket config with fallback to shared bucket
	bucketName := cfg.AdminBucketMedia()

	s3Creds := bucket.S3Credentials{
		AccessKey:      cfg.Bucket_Access_Key,
		SecretKey:      cfg.Bucket_Secret_Key,
		URL:            cfg.AdminBucketEndpointURL(),
		Region:         cfg.Bucket_Region,
		ForcePathStyle: cfg.Bucket_Force_Path_Style,
	}
	s3Session, err := s3Creds.GetBucket()
	if err != nil {
		return &InternalError{Err: fmt.Errorf("S3 session: %w", err)}
	}

	// Extract the S3 key from the admin media URL
	endpointPrefix := cfg.AdminBucketPublicURL() + "/" + bucketName + "/"
	originalKey := strings.TrimPrefix(string(record.URL), endpointPrefix)
	if originalKey == "" || originalKey == string(record.URL) {
		return fmt.Errorf("reprocess admin: could not extract S3 key from URL %s", record.URL)
	}

	// Create temp directory for downloaded original and generated variants
	tmpDir, err := os.MkdirTemp("", media.TempDirPrefix)
	if err != nil {
		return fmt.Errorf("reprocess admin: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the original file from S3
	resp, err := s3Session.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(originalKey),
	})
	if err != nil {
		return fmt.Errorf("reprocess admin: download original from S3: %w", err)
	}
	defer resp.Body.Close()

	originalFilename := filepath.Base(originalKey)
	localPath := filepath.Join(tmpDir, originalFilename)
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("reprocess admin: create local file: %w", err)
	}
	if _, err := io.Copy(localFile, resp.Body); err != nil {
		localFile.Close()
		return fmt.Errorf("reprocess admin: write local file: %w", err)
	}
	if err := localFile.Close(); err != nil {
		return fmt.Errorf("reprocess admin: close local file: %w", err)
	}

	// Decode image headers to get bounds for focal point conversion
	var focalPoint *image.Point
	if record.FocalX.Valid && record.FocalY.Valid {
		headerFile, openErr := os.Open(localPath)
		if openErr != nil {
			return fmt.Errorf("reprocess admin: open for header decode: %w", openErr)
		}
		imgConfig, _, decErr := image.DecodeConfig(headerFile)
		headerFile.Close()
		if decErr != nil {
			return fmt.Errorf("reprocess admin: decode image config: %w", decErr)
		}
		imgBounds := image.Rect(0, 0, imgConfig.Width, imgConfig.Height)
		focalPoint = media.FocalPointToPixels(record.FocalX, record.FocalY, imgBounds)
	}

	// Generate optimized variants
	optimized, err := media.OptimizeUpload(localPath, tmpDir, m.driver, focalPoint)
	if err != nil {
		return fmt.Errorf("reprocess admin: optimization failed: %w", err)
	}
	if optimized == nil || len(*optimized) == 0 {
		return nil
	}

	// Delete old variant S3 objects (from existing srcset)
	if record.Srcset.Valid && record.Srcset.String != "" {
		var oldSrcsetURLs []string
		if jsonErr := json.Unmarshal([]byte(record.Srcset.String), &oldSrcsetURLs); jsonErr == nil {
			for _, u := range oldSrcsetURLs {
				oldKey := strings.TrimPrefix(u, endpointPrefix)
				if oldKey == originalKey {
					continue
				}
				_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(oldKey),
				})
				if delErr != nil {
					utility.DefaultLogger.Warn("reprocess admin: failed to delete old variant", delErr, "key", oldKey)
				}
			}
		}
	}

	// Upload new variants to S3
	mediaPath := media.MediaPathFromURL(string(record.URL), endpointPrefix)
	acl := cfg.Bucket_Default_ACL
	if acl == "" {
		acl = "public-read"
	}

	srcset := []string{}
	uploadedKeys := []string{}

	for _, fullPath := range *optimized {
		f, openErr := os.Open(fullPath)
		if openErr != nil {
			rollbackVariantUploads(s3Session, bucketName, uploadedKeys)
			return fmt.Errorf("reprocess admin: open variant file: %w", openErr)
		}

		variantFilename := filepath.Base(fullPath)
		s3Key := fmt.Sprintf("%s/%s", mediaPath, variantFilename)
		uploadURL := fmt.Sprintf("%s/%s/%s", cfg.AdminBucketPublicURL(), bucketName, s3Key)

		prep, prepErr := bucket.UploadPrep(s3Key, bucketName, f, acl)
		if prepErr != nil {
			f.Close()
			rollbackVariantUploads(s3Session, bucketName, uploadedKeys)
			return fmt.Errorf("reprocess admin: upload prep: %w", prepErr)
		}

		_, upErr := bucket.ObjectUpload(s3Session, prep)
		f.Close()
		if upErr != nil {
			rollbackVariantUploads(s3Session, bucketName, uploadedKeys)
			return fmt.Errorf("reprocess admin: S3 upload: %w", upErr)
		}

		uploadedKeys = append(uploadedKeys, s3Key)
		srcset = append(srcset, uploadURL)
	}

	// Update srcset in database
	srcsetJSON, err := json.Marshal(srcset)
	if err != nil {
		rollbackVariantUploads(s3Session, bucketName, uploadedKeys)
		return fmt.Errorf("reprocess admin: marshal srcset: %w", err)
	}

	dbParams := media.MapAdminMediaParams(*record)
	dbParams.Srcset = db.NewNullString(string(srcsetJSON))

	if _, err := m.driver.UpdateAdminMedia(ctx, ac, dbParams); err != nil {
		rollbackVariantUploads(s3Session, bucketName, uploadedKeys)
		return fmt.Errorf("reprocess admin: update admin media record: %w", err)
	}

	return nil
}

// rollbackVariantUploads deletes uploaded variant files from S3 on failure.
func rollbackVariantUploads(s3Session *s3.S3, bucketName string, keys []string) {
	for _, key := range keys {
		_, err := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			utility.DefaultLogger.Warn("reprocess rollback: failed to delete S3 object", err, "key", key)
		}
	}
}

// --- Private Helpers ---

// newMediaS3Session creates an S3 session from the config.
func newMediaS3Session(c config.Config) (*s3.S3, error) {
	s3Creds := bucket.S3Credentials{
		AccessKey:      c.Bucket_Access_Key,
		SecretKey:      c.Bucket_Secret_Key,
		URL:            c.BucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}
	return s3Creds.GetBucket()
}

// extractMediaS3Keys parses URL + srcset JSON to collect all S3 keys for a media record.
func extractMediaS3Keys(record *db.Media, c config.Config) []string {
	endpointPrefix := c.BucketPublicURL() + "/" + c.Bucket_Media + "/"
	var keys []string

	if string(record.URL) != "" {
		key := strings.TrimPrefix(string(record.URL), endpointPrefix)
		keys = append(keys, key)
	}

	if record.Srcset.Valid && record.Srcset.String != "" {
		var srcsetURLs []string
		if jsonErr := json.Unmarshal([]byte(record.Srcset.String), &srcsetURLs); jsonErr == nil {
			for _, u := range srcsetURLs {
				key := strings.TrimPrefix(u, endpointPrefix)
				keys = append(keys, key)
			}
		}
	}

	return keys
}

// findOrphanedMediaKeys compares all S3 objects against DB records and returns untracked keys.
func findOrphanedMediaKeys(driver db.DbDriver, s3Session *s3.S3, c config.Config) (*OrphanScanResult, error) {
	mediaList, err := driver.ListMedia()
	if err != nil {
		return nil, fmt.Errorf("list media: %w", err)
	}

	endpointPrefix := c.BucketPublicURL() + "/" + c.Bucket_Media + "/"
	knownKeys := make(map[string]bool)

	for _, m := range *mediaList {
		if string(m.URL) != "" {
			knownKeys[strings.TrimPrefix(string(m.URL), endpointPrefix)] = true
		}
		if m.Srcset.Valid && m.Srcset.String != "" {
			var srcsetURLs []string
			if jsonErr := json.Unmarshal([]byte(m.Srcset.String), &srcsetURLs); jsonErr == nil {
				for _, u := range srcsetURLs {
					knownKeys[strings.TrimPrefix(u, endpointPrefix)] = true
				}
			}
		}
	}

	var orphanedKeys []string
	var totalObjects int

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.Bucket_Media),
	}

	for {
		result, listErr := s3Session.ListObjectsV2(listInput)
		if listErr != nil {
			return nil, fmt.Errorf("list bucket objects: %w", listErr)
		}

		for _, obj := range result.Contents {
			totalObjects++
			key := aws.StringValue(obj.Key)
			if !knownKeys[key] {
				orphanedKeys = append(orphanedKeys, key)
			}
		}

		if !aws.BoolValue(result.IsTruncated) {
			break
		}
		listInput.ContinuationToken = result.NextContinuationToken
	}

	return &OrphanScanResult{
		TotalObjects: totalObjects,
		TrackedKeys:  len(knownKeys),
		OrphanedKeys: orphanedKeys,
		OrphanCount:  len(orphanedKeys),
	}, nil
}

// overlayNullString returns a new NullString from val if non-empty, otherwise keeps existing.
func overlayNullString(val string, existing db.NullString) db.NullString {
	trimmed := strings.TrimSpace(val)
	if trimmed != "" {
		return db.NewNullString(trimmed)
	}
	return existing
}

// overlayFocal returns newVal if valid, otherwise keeps existing.
func overlayFocal(newVal, existing types.NullableFloat64) types.NullableFloat64 {
	if newVal.Valid {
		return newVal
	}
	return existing
}
