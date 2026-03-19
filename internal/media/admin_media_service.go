// admin_media_service.go mirrors media_service.go for the admin content system.
// It defines the AdminMediaStore interface, ProcessAdminMediaUpload pipeline,
// HandleAdminMediaUpload optimization pipeline, and MapAdminMediaParams helper.
//
// Admin media uses its own S3 bucket configuration (with fallback to the shared
// media bucket) via config.AdminBucket*() methods.
package media

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"

	bucket "github.com/hegner123/modulacms/internal/bucket"
	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	utility "github.com/hegner123/modulacms/internal/utility"
)

// AdminMediaStore is the consumer-defined interface for admin media persistence.
// All three DB drivers (Database, MysqlDatabase, PsqlDatabase) satisfy this implicitly.
type AdminMediaStore interface {
	GetAdminMediaByName(name string) (*db.AdminMedia, error)
	CreateAdminMedia(ctx context.Context, ac audited.AuditContext, params db.CreateAdminMediaParams) (*db.AdminMedia, error)
	DeleteAdminMedia(ctx context.Context, ac audited.AuditContext, id types.AdminMediaID) error
}

// ProcessAdminMediaUpload validates, uploads the original to S3, persists to DB, and
// conditionally runs the image optimization pipeline for supported image types.
//
// This mirrors ProcessMediaUpload but targets the admin_media table and accepts
// NullableAdminMediaFolderID for folder placement.
//
// Flow:
//  1. Validate size and detect MIME type
//  2. Check for duplicate filename
//  3. Write uploaded file to temp directory
//  4. Upload original to S3 via uploadOriginal callback
//  5. Create admin media DB record (with URL from step 4)
//  6. Run optimization/upload pipeline (images only)
//
// Rollback: S3 original is deleted if DB create or pipeline fails.
// DB record is deleted if pipeline fails after DB create succeeds.
func ProcessAdminMediaUpload(
	ctx context.Context,
	ac audited.AuditContext,
	file multipart.File,
	header *multipart.FileHeader,
	store AdminMediaStore,
	uploadOriginal UploadOriginalFunc,
	rollbackS3 RollbackS3Func,
	pipeline UploadPipelineFunc,
	maxUploadSize int64,
	folderID types.NullableAdminMediaFolderID,
) (*db.AdminMedia, error) {
	// Step 1: Validate size
	if header.Size > maxUploadSize {
		return nil, FileTooLargeError{Size: header.Size, MaxSize: maxUploadSize}
	}

	// Detect MIME type from first 512 bytes
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read file header: %w", err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("seek file: %w", err)
	}

	contentType := http.DetectContentType(buffer)

	// Step 2: Deduplicate filename -- append -1, -2, etc. if name exists
	filename := header.Filename
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for suffix := 1; suffix <= 100; suffix++ {
		_, nameErr := store.GetAdminMediaByName(filename)
		if nameErr != nil {
			break // name is available
		}
		filename = fmt.Sprintf("%s-%d%s", base, suffix, ext)
	}
	header.Filename = filename

	// Step 3: Write uploaded file to temp directory
	tmp, err := os.MkdirTemp("", TempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	dstPath := filepath.Join(tmp, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("create destination file: %w", err)
	}

	written, copyErr := io.Copy(dst, file)
	if copyErr != nil {
		dst.Close()
		return nil, fmt.Errorf("copy file: %w", copyErr)
	}

	if err := dst.Close(); err != nil {
		return nil, fmt.Errorf("flush temp file: %w", err)
	}

	utility.DefaultLogger.Info("admin media temp file written", "path", dstPath, "bytes", written)

	// Step 4: Upload original to S3
	originalURL, originalKey, err := uploadOriginal(dstPath)
	if err != nil {
		return nil, fmt.Errorf("upload original to S3: %w", err)
	}

	// Step 5: Create admin media DB record with URL and mimetype
	params := db.CreateAdminMediaParams{
		Name:         db.NewNullString(header.Filename),
		Mimetype:     db.NewNullString(contentType),
		URL:          types.URL(originalURL),
		AuthorID:     types.NullableUserID{ID: ac.UserID, Valid: ac.UserID != ""},
		FolderID:     folderID,
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	utility.DefaultLogger.Info("CreateAdminMedia params", "folder_id_valid", params.FolderID.Valid, "folder_id", params.FolderID.ID)

	row, err := store.CreateAdminMedia(ctx, ac, params)
	if err != nil {
		rollbackS3(originalKey)
		return nil, fmt.Errorf("create admin media record: %w", err)
	}

	// Step 6: Run optimization/upload pipeline (images only)
	if IsImageMIME(contentType) {
		if err := pipeline(dstPath, tmp); err != nil {
			rollbackS3(originalKey)
			deleteErr := store.DeleteAdminMedia(ctx, ac, row.AdminMediaID)
			if deleteErr != nil {
				utility.DefaultLogger.Error("failed to rollback admin media record", deleteErr)
			}
			return nil, fmt.Errorf("admin media pipeline: %w", err)
		}
	}

	return row, nil
}

// HandleAdminMediaUpload optimizes and uploads admin media files to S3, then
// updates the admin_media database record with the resulting srcset.
//
// This mirrors HandleMediaUpload but uses AdminBucket*() config methods for
// S3 credentials and the admin bucket name.
func HandleAdminMediaUpload(srcFile string, dstPath string, c config.Config) error {
	d := db.ConfigDB(c)
	bucketDir := c.AdminBucketMedia()

	filename := filepath.Base(srcFile)

	// Step 1a: Fetch admin media record to get focal point
	utility.DefaultLogger.Debug(fmt.Sprintf("Fetching admin media record for: %s", filename))
	rowPtr, err := d.GetAdminMediaByName(filename)
	if err != nil {
		return fmt.Errorf("failed to get admin media record: %w", err)
	}
	row := *rowPtr

	// Step 1b: Decode image headers to get bounds for focal point conversion
	var focalPoint *image.Point
	if row.FocalX.Valid && row.FocalY.Valid {
		headerFile, err := os.Open(srcFile)
		if err != nil {
			return fmt.Errorf("failed to open file for header decode: %w", err)
		}
		imgConfig, _, err := image.DecodeConfig(headerFile)
		headerFile.Close()
		if err != nil {
			return fmt.Errorf("failed to decode image config: %w", err)
		}
		imgBounds := image.Rect(0, 0, imgConfig.Width, imgConfig.Height)
		focalPoint = FocalPointToPixels(row.FocalX, row.FocalY, imgBounds)
	}

	// Step 1c: Optimize images (reuses shared dimension presets)
	optimized, err := OptimizeUpload(srcFile, dstPath, d, focalPoint)
	if err != nil {
		return fmt.Errorf("optimization failed: %w", err)
	}

	// Step 2: Setup S3 session using admin bucket credentials
	s3Creds := bucket.S3Credentials{
		AccessKey:      c.AdminBucketAccessKey(),
		SecretKey:      c.AdminBucketSecretKey(),
		URL:            c.AdminBucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}

	s3Session, err := s3Creds.GetBucket()
	if err != nil {
		return fmt.Errorf("S3 session failed: %w", err)
	}

	// Step 3: Upload ALL to S3 (track successes for rollback)
	srcset := []string{}
	uploadedKeys := []string{}

	// Derive the S3 path prefix from the original upload URL
	endpointPrefix := c.AdminBucketPublicURL() + "/" + bucketDir + "/"
	mediaPath := MediaPathFromURL(string(row.URL), endpointPrefix)

	// Get ACL from config or use default
	acl := c.Bucket_Default_ACL
	if acl == "" {
		acl = "public-read" // Default for backwards compatibility
	}

	for _, fullPath := range *optimized {
		file, err := os.Open(fullPath)
		if err != nil {
			rollbackS3Uploads(s3Session, bucketDir, uploadedKeys)
			return fmt.Errorf("failed to open optimized file: %w", err)
		}

		filename := filepath.Base(fullPath)
		s3Key := fmt.Sprintf("%s/%s", mediaPath, filename)
		uploadPath := fmt.Sprintf("%s/%s/%s", c.AdminBucketPublicURL(), bucketDir, s3Key)

		prep, err := bucket.UploadPrep(s3Key, bucketDir, file, acl)
		if err != nil {
			file.Close()
			rollbackS3Uploads(s3Session, bucketDir, uploadedKeys)
			return fmt.Errorf("upload prep failed: %w", err)
		}

		_, err = bucket.ObjectUpload(s3Session, prep)
		file.Close()
		if err != nil {
			rollbackS3Uploads(s3Session, bucketDir, uploadedKeys)
			return fmt.Errorf("S3 upload failed: %w", err)
		}

		uploadedKeys = append(uploadedKeys, s3Key)
		srcset = append(srcset, uploadPath)
	}

	// Step 4: All uploads succeeded - update database
	srcsetJSON, err := json.Marshal(srcset)
	if err != nil {
		rollbackS3Uploads(s3Session, bucketDir, uploadedKeys)
		return fmt.Errorf("failed to marshal srcset: %w", err)
	}

	updateParams := MapAdminMediaParams(row)
	updateParams.Srcset = db.NewNullString(string(srcsetJSON))

	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(c.Node_ID), types.UserID(""), "", "system")
	_, err = d.UpdateAdminMedia(ctx, ac, updateParams)
	if err != nil {
		rollbackS3Uploads(s3Session, bucketDir, uploadedKeys)
		return fmt.Errorf("database update failed: %w", err)
	}

	return nil
}

// MapAdminMediaParams converts an AdminMedia record to UpdateAdminMediaParams,
// updating the modification timestamp. Mirrors MapMediaParams for the admin table.
func MapAdminMediaParams(a db.AdminMedia) db.UpdateAdminMediaParams {
	return db.UpdateAdminMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		URL:          a.URL,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		FocalX:       a.FocalX,
		FocalY:       a.FocalY,
		AuthorID:     a.AuthorID,
		FolderID:     a.FolderID,
		DateCreated:  a.DateCreated,
		DateModified: types.TimestampNow(),
		AdminMediaID: a.AdminMediaID,
	}
}
