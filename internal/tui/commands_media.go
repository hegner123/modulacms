package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediaUploader is a consumer-defined interface satisfied by RemoteDriver.
// The TUI type-asserts the DbDriver to this interface for remote media uploads.
type MediaUploader interface {
	UploadMedia(ctx context.Context, filePath string) (*db.Media, error)
}

// MediaProgressUploader extends MediaUploader with progress callback support.
type MediaProgressUploader interface {
	MediaUploader
	UploadMediaWithProgress(ctx context.Context, filePath string, progressFn func(bytesSent int64, total int64)) (*db.Media, error)
}

// waitForMsg returns a tea.Cmd that blocks until a message arrives on ch.
func waitForMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

// HandleMediaUpload runs the media upload pipeline asynchronously.
// In remote mode, the file is sent to the server via the SDK with progress.
// In local mode, the existing optimize+S3 pipeline runs.
func (m Model) HandleMediaUpload(msg MediaUploadStartMsg) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	isRemote := m.IsRemote
	return func() tea.Msg {
		filename := filepath.Base(msg.FilePath)
		baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

		logger.Finfo(fmt.Sprintf("Starting media upload: %s", filename))

		// Remote mode: upload via SDK with progress channel
		if isRemote {
			d := db.ConfigDB(*cfg)
			progressCh := make(chan tea.Msg, 1)

			// Try progress-aware uploader first, fall back to basic
			if pu, ok := d.(MediaProgressUploader); ok {
				go func() {
					progressFn := func(bytesSent int64, total int64) {
						select {
						case progressCh <- MediaUploadProgressMsg{
							BytesSent:  bytesSent,
							Total:      total,
							ProgressCh: progressCh,
						}:
						default: // don't block if channel is full
						}
					}
					_, err := pu.UploadMediaWithProgress(context.Background(), msg.FilePath, progressFn)
					if err != nil {
						logger.Ferror("Remote media upload failed", err)
						progressCh <- ActionResultMsg{
							Title:   "Upload Error",
							Message: fmt.Sprintf("Upload failed: %v", err),
							IsError: true,
						}
						return
					}
					logger.Finfo(fmt.Sprintf("Media uploaded remotely: %s", baseName))
					progressCh <- MediaUploadedMsg{Name: baseName}
				}()
				return <-progressCh
			}

			// Fall back to basic uploader without progress
			uploader, ok := d.(MediaUploader)
			if !ok {
				return ActionResultMsg{
					Title:   "Upload Error",
					Message: "Remote driver does not support media upload",
					IsError: true,
				}
			}
			_, err := uploader.UploadMedia(context.Background(), msg.FilePath)
			if err != nil {
				logger.Ferror("Remote media upload failed", err)
				return ActionResultMsg{
					Title:   "Upload Error",
					Message: fmt.Sprintf("Upload failed: %v", err),
					IsError: true,
				}
			}
			logger.Finfo(fmt.Sprintf("Media uploaded remotely: %s", baseName))
			return MediaUploadedMsg{Name: baseName}
		}

		// Local mode: existing pipeline
		// Step 1: Create placeholder DB record
		_, err := media.CreateMedia(baseName, *cfg)
		if err != nil {
			logger.Ferror("Failed to create media record", err)
			return ActionResultMsg{
				Title:   "Upload Error",
				Message: fmt.Sprintf("Failed to create media record: %v", err),
			}
		}

		// Step 2: Create temp directory for optimized files
		tmpDir, err := os.MkdirTemp("", media.TempDirPrefix)
		if err != nil {
			logger.Ferror("Failed to create temp directory", err)
			return ActionResultMsg{
				Title:   "Upload Error",
				Message: fmt.Sprintf("Failed to create temp directory: %v", err),
			}
		}
		defer os.RemoveAll(tmpDir)

		// Step 3: Run upload pipeline (optimize -> S3 upload -> DB update)
		if err := media.HandleMediaUpload(msg.FilePath, tmpDir, *cfg); err != nil {
			logger.Ferror("Media upload failed", err)
			return ActionResultMsg{
				Title:   "Upload Error",
				Message: fmt.Sprintf("Upload failed: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Media uploaded successfully: %s", baseName))
		return MediaUploadedMsg{Name: baseName}
	}
}
