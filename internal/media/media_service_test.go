package media

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// mockStore implements MediaStore for testing.
type mockStore struct {
	getByNameResult *db.Media
	getByNameErr    error
	createResult    *db.Media
	createErr       error
	deleteErr       error
	deleteCalled    bool
}

func (m *mockStore) GetMediaByName(name string) (*db.Media, error) {
	return m.getByNameResult, m.getByNameErr
}

func (m *mockStore) CreateMedia(ctx context.Context, ac audited.AuditContext, params db.CreateMediaParams) (*db.Media, error) {
	return m.createResult, m.createErr
}

func (m *mockStore) DeleteMedia(ctx context.Context, ac audited.AuditContext, id types.MediaID) error {
	m.deleteCalled = true
	return m.deleteErr
}

// pngHeader returns minimal valid PNG bytes (8-byte signature + IHDR chunk)
// that http.DetectContentType identifies as image/png.
func pngHeader() []byte {
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // "IHDR"
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, // bit depth 8, color type 2 (RGB)
		0x00, 0x00, 0x00, // compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
	}
}

// inMemoryFile wraps a bytes.Reader to satisfy multipart.File (Read + ReadAt + Seek + Close).
type inMemoryFile struct {
	*bytes.Reader
}

func (f *inMemoryFile) Close() error { return nil }

func newTestFile(data []byte) multipart.File {
	return &inMemoryFile{Reader: bytes.NewReader(data)}
}

func newTestHeader(filename string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Size:     size,
	}
}

func noopPipeline(srcFile string, dstPath string) error {
	return nil
}

func noopUploadOriginal(filePath string) (string, string, error) {
	return "http://test/original.png", "media/original.png", nil
}

func noopRollbackS3(s3Key string) {}

func testAuditCtx() audited.AuditContext {
	return audited.Ctx(types.NodeID("test-node"), types.UserID("test-user"), "req-1", "127.0.0.1")
}

func TestProcessMediaUpload(t *testing.T) {
	validPNG := pngHeader()
	defaultMedia := &db.Media{
		MediaID:     types.NewMediaID(),
		Name:        sql.NullString{String: "test.png", Valid: true},
		AuthorID:    types.NullableUserID{ID: types.UserID("test-user"), Valid: true},
		DateCreated: types.TimestampNow(),
	}

	tests := []struct {
		name           string
		fileData       []byte
		header         *multipart.FileHeader
		store          *mockStore
		uploadOriginal UploadOriginalFunc
		pipeline       UploadPipelineFunc
		wantErr        bool
		wantErrType    any
	}{
		{
			name:     "valid image upload",
			fileData: validPNG,
			header:   newTestHeader("test.png", int64(len(validPNG))),
			store: &mockStore{
				getByNameErr: errors.New("not found"),
				createResult: defaultMedia,
			},
			uploadOriginal: noopUploadOriginal,
			pipeline:       noopPipeline,
			wantErr:        false,
		},
		{
			name:     "duplicate filename",
			fileData: validPNG,
			header:   newTestHeader("existing.png", int64(len(validPNG))),
			store: &mockStore{
				getByNameResult: defaultMedia,
				getByNameErr:    nil,
			},
			uploadOriginal: noopUploadOriginal,
			pipeline:       noopPipeline,
			wantErr:        true,
			wantErrType:    DuplicateMediaError{},
		},
		{
			name:     "invalid MIME type",
			fileData: []byte("not an image, just plain text content"),
			header:   newTestHeader("readme.txt", 37),
			store: &mockStore{
				getByNameErr: errors.New("not found"),
			},
			uploadOriginal: noopUploadOriginal,
			pipeline:       noopPipeline,
			wantErr:        true,
			wantErrType:    InvalidMediaTypeError{},
		},
		{
			name:     "file too large",
			fileData: validPNG,
			header:   newTestHeader("huge.png", MaxUploadSize+1),
			store: &mockStore{
				getByNameErr: errors.New("not found"),
			},
			uploadOriginal: noopUploadOriginal,
			pipeline:       noopPipeline,
			wantErr:        true,
			wantErrType:    FileTooLargeError{},
		},
		{
			name:     "S3 original upload failure",
			fileData: validPNG,
			header:   newTestHeader("test.png", int64(len(validPNG))),
			store: &mockStore{
				getByNameErr: errors.New("not found"),
			},
			uploadOriginal: func(filePath string) (string, string, error) {
				return "", "", errors.New("S3 connection refused")
			},
			pipeline: noopPipeline,
			wantErr:  true,
		},
		{
			name:     "DB create failure rolls back S3",
			fileData: validPNG,
			header:   newTestHeader("test.png", int64(len(validPNG))),
			store: &mockStore{
				getByNameErr: errors.New("not found"),
				createErr:    errors.New("db connection lost"),
			},
			uploadOriginal: noopUploadOriginal,
			pipeline:       noopPipeline,
			wantErr:        true,
		},
		{
			name:     "pipeline failure rolls back S3 and DB",
			fileData: validPNG,
			header:   newTestHeader("test.png", int64(len(validPNG))),
			store: &mockStore{
				getByNameErr: errors.New("not found"),
				createResult: defaultMedia,
			},
			uploadOriginal: noopUploadOriginal,
			pipeline: func(srcFile string, dstPath string) error {
				return errors.New("optimization failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := newTestFile(tt.fileData)
			ctx := context.Background()
			ac := testAuditCtx()

			result, err := ProcessMediaUpload(ctx, ac, file, tt.header, tt.store, tt.uploadOriginal, noopRollbackS3, tt.pipeline)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrType != nil {
					switch tt.wantErrType.(type) {
					case DuplicateMediaError:
						var target DuplicateMediaError
						if !errors.As(err, &target) {
							t.Errorf("expected DuplicateMediaError, got %T: %v", err, err)
						}
					case InvalidMediaTypeError:
						var target InvalidMediaTypeError
						if !errors.As(err, &target) {
							t.Errorf("expected InvalidMediaTypeError, got %T: %v", err, err)
						}
					case FileTooLargeError:
						var target FileTooLargeError
						if !errors.As(err, &target) {
							t.Errorf("expected FileTooLargeError, got %T: %v", err, err)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.MediaID != defaultMedia.MediaID {
				t.Errorf("expected media ID %s, got %s", defaultMedia.MediaID, result.MediaID)
			}
		})
	}
}

// TestProcessMediaUpload_AuditContextUserID verifies the created record uses
// the UserID from the audit context, not a hardcoded value.
func TestProcessMediaUpload_AuditContextUserID(t *testing.T) {
	validPNG := pngHeader()
	expectedUserID := types.UserID("real-user-42")

	var capturedParams db.CreateMediaParams
	store := &mockStore{
		getByNameErr: errors.New("not found"),
	}
	// Override CreateMedia to capture params
	capturingStore := &capturingMockStore{
		mockStore: store,
		onCreate: func(params db.CreateMediaParams) {
			capturedParams = params
		},
	}

	ac := audited.Ctx(types.NodeID("node"), expectedUserID, "req", "127.0.0.1")
	file := newTestFile(validPNG)
	header := newTestHeader("test.png", int64(len(validPNG)))

	_, err := ProcessMediaUpload(context.Background(), ac, file, header, capturingStore, noopUploadOriginal, noopRollbackS3, noopPipeline)
	// Pipeline writes to temp which is fine; we only care about the params check
	if err != nil {
		// The result may error from the pipeline in a temp dir, but we can still verify params were captured
		if capturedParams.AuthorID.ID != expectedUserID {
			t.Errorf("expected AuthorID %s, got %s", expectedUserID, capturedParams.AuthorID.ID)
		}
		return
	}

	if capturedParams.AuthorID.ID != expectedUserID {
		t.Errorf("expected AuthorID %s, got %s", expectedUserID, capturedParams.AuthorID.ID)
	}
}

// capturingMockStore wraps mockStore and captures CreateMedia params.
type capturingMockStore struct {
	mockStore *mockStore
	onCreate  func(db.CreateMediaParams)
}

func (c *capturingMockStore) GetMediaByName(name string) (*db.Media, error) {
	return c.mockStore.GetMediaByName(name)
}

func (c *capturingMockStore) CreateMedia(ctx context.Context, ac audited.AuditContext, params db.CreateMediaParams) (*db.Media, error) {
	c.onCreate(params)
	result := &db.Media{
		MediaID:      types.NewMediaID(),
		Name:         params.Name,
		AuthorID:     params.AuthorID,
		DateCreated:  params.DateCreated,
		DateModified: params.DateModified,
	}
	return result, nil
}

func (c *capturingMockStore) DeleteMedia(ctx context.Context, ac audited.AuditContext, id types.MediaID) error {
	return nil
}

// Verify multipart.File interface satisfaction at compile time.
var _ multipart.File = (*inMemoryFile)(nil)
var _ io.ReaderAt = (*inMemoryFile)(nil)

// jpegHeader returns minimal bytes that http.DetectContentType identifies as image/jpeg.
func jpegHeader() []byte {
	return []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}
}

// gifHeader returns minimal bytes that http.DetectContentType identifies as image/gif.
func gifHeader() []byte {
	return []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61} // GIF89a
}

// webpHeader returns minimal bytes that http.DetectContentType identifies as
// image/webp. The Go sniffer requires "RIFF" + 4 size bytes + "WEBPVP" to
// match the masked signature pattern.
func webpHeader() []byte {
	return []byte{
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x00, 0x00, 0x00, 0x00, // file size (masked out by sniffer)
		0x57, 0x45, 0x42, 0x50, // "WEBP"
		0x56, 0x50,             // "VP" (required by sniffer mask)
	}
}

// TestProcessMediaUpload_AcceptsAllValidMIMETypes verifies that PNG, JPEG, GIF,
// and WebP content types are accepted by the MIME validation step.
func TestProcessMediaUpload_AcceptsAllValidMIMETypes(t *testing.T) {
	t.Parallel()

	defaultMedia := &db.Media{
		MediaID:     types.NewMediaID(),
		Name:        sql.NullString{String: "test", Valid: true},
		AuthorID:    types.NullableUserID{ID: types.UserID("test-user"), Valid: true},
		DateCreated: types.TimestampNow(),
	}

	tests := []struct {
		name     string
		fileData []byte
		filename string
	}{
		{
			name:     "JPEG accepted",
			fileData: jpegHeader(),
			filename: "photo.jpg",
		},
		{
			name:     "GIF accepted",
			fileData: gifHeader(),
			filename: "animation.gif",
		},
		{
			name:     "WebP accepted",
			fileData: webpHeader(),
			filename: "modern.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &mockStore{
				getByNameErr: errors.New("not found"),
				createResult: defaultMedia,
			}
			file := newTestFile(tt.fileData)
			header := newTestHeader(tt.filename, int64(len(tt.fileData)))

			result, err := ProcessMediaUpload(
				context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, noopRollbackS3, noopPipeline,
			)
			if err != nil {
				t.Fatalf("expected no error for %s, got: %v", tt.name, err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
		})
	}
}

// errReader is a multipart.File implementation whose Read always fails.
// Used to test the file read error path in ProcessMediaUpload.
type errReader struct {
	readErr error
	seekErr error
}

func (e *errReader) Read(p []byte) (int, error) {
	if e.readErr != nil {
		return 0, e.readErr
	}
	return 0, io.EOF
}

func (e *errReader) ReadAt(p []byte, off int64) (int, error) {
	return e.Read(p)
}

func (e *errReader) Seek(offset int64, whence int) (int64, error) {
	if e.seekErr != nil {
		return 0, e.seekErr
	}
	return 0, nil
}

func (e *errReader) Close() error { return nil }

// Compile-time check that errReader satisfies multipart.File.
var _ multipart.File = (*errReader)(nil)

// TestProcessMediaUpload_ReadError verifies that a file whose Read fails
// returns a wrapped "read file header" error.
func TestProcessMediaUpload_ReadError(t *testing.T) {
	t.Parallel()
	file := &errReader{readErr: errors.New("disk failure")}
	header := newTestHeader("test.png", 100)
	store := &mockStore{getByNameErr: errors.New("not found")}

	_, err := ProcessMediaUpload(
		context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, noopRollbackS3, noopPipeline,
	)
	if err == nil {
		t.Fatal("expected error from Read failure, got nil")
	}
	if !strings.Contains(err.Error(), "read file header") {
		t.Errorf("expected error to contain 'read file header', got: %v", err)
	}
}

// seekFailReader reads successfully but Seek always fails. This exercises the
// seek error path after MIME detection reads the first 512 bytes.
type seekFailReader struct {
	*bytes.Reader
}

func (s *seekFailReader) Close() error { return nil }
func (s *seekFailReader) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek not supported")
}

var _ multipart.File = (*seekFailReader)(nil)

// TestProcessMediaUpload_SeekError verifies that a seek failure after MIME
// detection returns a wrapped "seek file" error.
func TestProcessMediaUpload_SeekError(t *testing.T) {
	t.Parallel()
	// Use valid PNG data so the Read succeeds and MIME detection passes,
	// then the Seek back to 0 will fail.
	file := &seekFailReader{Reader: bytes.NewReader(pngHeader())}
	header := newTestHeader("test.png", int64(len(pngHeader())))
	store := &mockStore{getByNameErr: errors.New("not found")}

	_, err := ProcessMediaUpload(
		context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, noopRollbackS3, noopPipeline,
	)
	if err == nil {
		t.Fatal("expected error from Seek failure, got nil")
	}
	if !strings.Contains(err.Error(), "seek file") {
		t.Errorf("expected error to contain 'seek file', got: %v", err)
	}
}

// TestProcessMediaUpload_EmptyUserID verifies that passing an empty UserID in
// the audit context results in AuthorID.Valid == false in the create params.
func TestProcessMediaUpload_EmptyUserID(t *testing.T) {
	t.Parallel()
	validPNG := pngHeader()
	var capturedParams db.CreateMediaParams

	store := &capturingMockStore{
		mockStore: &mockStore{getByNameErr: errors.New("not found")},
		onCreate: func(params db.CreateMediaParams) {
			capturedParams = params
		},
	}

	// Empty UserID -- simulates unauthenticated/system context
	ac := audited.Ctx(types.NodeID("node"), types.UserID(""), "req", "127.0.0.1")
	file := newTestFile(validPNG)
	header := newTestHeader("test.png", int64(len(validPNG)))

	// We don't care about the pipeline result here, just the captured params
	_, _ = ProcessMediaUpload(context.Background(), ac, file, header, store, noopUploadOriginal, noopRollbackS3, noopPipeline)

	if capturedParams.AuthorID.Valid {
		t.Errorf("expected AuthorID.Valid=false for empty UserID, got true (ID=%s)", capturedParams.AuthorID.ID)
	}
}

// TestProcessMediaUpload_FilenamePassedToStore verifies the filename from the
// multipart header is used for both the duplicate check and the DB record name.
func TestProcessMediaUpload_FilenamePassedToStore(t *testing.T) {
	t.Parallel()
	validPNG := pngHeader()
	expectedFilename := "my-special-image.png"

	var getByNameArg string
	var createNameArg sql.NullString

	store := &recordingMockStore{
		onGetByName: func(name string) (*db.Media, error) {
			getByNameArg = name
			return nil, errors.New("not found")
		},
		onCreate: func(ctx context.Context, ac audited.AuditContext, params db.CreateMediaParams) (*db.Media, error) {
			createNameArg = params.Name
			return &db.Media{
				MediaID:     types.NewMediaID(),
				Name:        params.Name,
				DateCreated: types.TimestampNow(),
			}, nil
		},
	}

	file := newTestFile(validPNG)
	header := newTestHeader(expectedFilename, int64(len(validPNG)))

	_, _ = ProcessMediaUpload(context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, noopRollbackS3, noopPipeline)

	if getByNameArg != expectedFilename {
		t.Errorf("GetMediaByName called with %q, want %q", getByNameArg, expectedFilename)
	}
	if !createNameArg.Valid || createNameArg.String != expectedFilename {
		t.Errorf("CreateMedia name = %+v, want {String:%q Valid:true}", createNameArg, expectedFilename)
	}
}

// recordingMockStore provides full callback control over all MediaStore methods.
type recordingMockStore struct {
	onGetByName func(name string) (*db.Media, error)
	onCreate    func(ctx context.Context, ac audited.AuditContext, params db.CreateMediaParams) (*db.Media, error)
	onDelete    func(ctx context.Context, ac audited.AuditContext, id types.MediaID) error
}

func (r *recordingMockStore) GetMediaByName(name string) (*db.Media, error) {
	return r.onGetByName(name)
}

func (r *recordingMockStore) CreateMedia(ctx context.Context, ac audited.AuditContext, params db.CreateMediaParams) (*db.Media, error) {
	return r.onCreate(ctx, ac, params)
}

func (r *recordingMockStore) DeleteMedia(ctx context.Context, ac audited.AuditContext, id types.MediaID) error {
	if r.onDelete != nil {
		return r.onDelete(ctx, ac, id)
	}
	return nil
}

// TestProcessMediaUpload_ExactSizeBoundary verifies that a file whose size is
// exactly MaxUploadSize is accepted (the check is strictly greater-than).
func TestProcessMediaUpload_ExactSizeBoundary(t *testing.T) {
	t.Parallel()
	validPNG := pngHeader()
	defaultMedia := &db.Media{
		MediaID:     types.NewMediaID(),
		Name:        sql.NullString{String: "boundary.png", Valid: true},
		DateCreated: types.TimestampNow(),
	}

	store := &mockStore{
		getByNameErr: errors.New("not found"),
		createResult: defaultMedia,
	}
	file := newTestFile(validPNG)
	// Size exactly at the limit -- should NOT trigger FileTooLargeError
	header := newTestHeader("boundary.png", MaxUploadSize)

	result, err := ProcessMediaUpload(
		context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, noopRollbackS3, noopPipeline,
	)
	if err != nil {
		t.Fatalf("file at exact MaxUploadSize should be accepted, got error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestProcessMediaUpload_PipelineReceivesCorrectPaths verifies that the pipeline
// function receives the temp file path and temp directory.
func TestProcessMediaUpload_PipelineReceivesCorrectPaths(t *testing.T) {
	t.Parallel()
	validPNG := pngHeader()
	defaultMedia := &db.Media{
		MediaID:     types.NewMediaID(),
		Name:        sql.NullString{String: "pipeline-test.png", Valid: true},
		DateCreated: types.TimestampNow(),
	}

	store := &mockStore{
		getByNameErr: errors.New("not found"),
		createResult: defaultMedia,
	}

	var capturedSrcFile, capturedDstPath string
	pipeline := func(srcFile string, dstPath string) error {
		capturedSrcFile = srcFile
		capturedDstPath = dstPath
		return nil
	}

	file := newTestFile(validPNG)
	header := newTestHeader("pipeline-test.png", int64(len(validPNG)))

	_, err := ProcessMediaUpload(
		context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, noopRollbackS3, pipeline,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The srcFile should end with the original filename
	if !strings.HasSuffix(capturedSrcFile, "pipeline-test.png") {
		t.Errorf("srcFile %q should end with 'pipeline-test.png'", capturedSrcFile)
	}
	// The dstPath should be a parent directory of srcFile
	if !strings.HasPrefix(capturedSrcFile, capturedDstPath) {
		t.Errorf("srcFile %q should be inside dstPath %q", capturedSrcFile, capturedDstPath)
	}
}

// TestProcessMediaUpload_DBCreateFailureRollsBackS3 verifies that when DB create
// fails after the original is uploaded to S3, the S3 upload is rolled back.
func TestProcessMediaUpload_DBCreateFailureRollsBackS3(t *testing.T) {
	t.Parallel()
	validPNG := pngHeader()

	store := &mockStore{
		getByNameErr: errors.New("not found"),
		createErr:    errors.New("db connection lost"),
	}

	var rolledBackKey string
	rollback := func(s3Key string) {
		rolledBackKey = s3Key
	}

	file := newTestFile(validPNG)
	header := newTestHeader("test.png", int64(len(validPNG)))

	_, err := ProcessMediaUpload(
		context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, rollback, noopPipeline,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if rolledBackKey != "media/original.png" {
		t.Errorf("expected rollback key %q, got %q", "media/original.png", rolledBackKey)
	}
}

// TestProcessMediaUpload_PipelineFailureRollsBackS3AndDB verifies that when
// the pipeline fails, both the S3 original and DB record are rolled back.
func TestProcessMediaUpload_PipelineFailureRollsBackS3AndDB(t *testing.T) {
	t.Parallel()
	validPNG := pngHeader()
	defaultMedia := &db.Media{
		MediaID:     types.NewMediaID(),
		Name:        sql.NullString{String: "test.png", Valid: true},
		DateCreated: types.TimestampNow(),
	}

	store := &mockStore{
		getByNameErr: errors.New("not found"),
		createResult: defaultMedia,
	}

	var rolledBackKey string
	rollback := func(s3Key string) {
		rolledBackKey = s3Key
	}

	file := newTestFile(validPNG)
	header := newTestHeader("test.png", int64(len(validPNG)))

	_, err := ProcessMediaUpload(
		context.Background(), testAuditCtx(), file, header, store, noopUploadOriginal, rollback,
		func(srcFile string, dstPath string) error {
			return errors.New("optimization failed")
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if rolledBackKey != "media/original.png" {
		t.Errorf("expected rollback key %q, got %q", "media/original.png", rolledBackKey)
	}
	if !store.deleteCalled {
		t.Error("expected DeleteMedia to be called for DB rollback")
	}
}
