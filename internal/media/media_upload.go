package media

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	bucket "github.com/hegner123/modulacms/internal/bucket"
	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	utility "github.com/hegner123/modulacms/internal/utility"
	_ "golang.org/x/image/webp"
)

// HandleMediaUpload optimizes and uploads media files to S3, then updates the database with the resulting srcset.
func HandleMediaUpload(srcFile string, dstPath string, c config.Config) error {
	d := db.ConfigDB(c)
	bucketDir := c.Bucket_Media
	now := time.Now()
	year := now.Year()
	month := now.Month()

	filename := filepath.Base(srcFile)
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Step 1a: Fetch media record to get focal point
	utility.DefaultLogger.Debug(fmt.Sprintf("Fetching media record for: %s", baseName))
	rowPtr, err := d.GetMediaByName(baseName)
	if err != nil {
		return fmt.Errorf("failed to get media record: %w", err)
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

	// Step 1c: Optimize images
	optimized, err := OptimizeUpload(srcFile, dstPath, d, focalPoint)
	if err != nil {
		return fmt.Errorf("optimization failed: %w", err)
	}

	// Step 2: Setup S3 session
	s3Creds := bucket.S3Credentials{
		AccessKey:      c.Bucket_Access_Key,
		SecretKey:      c.Bucket_Secret_Key,
		URL:            c.BucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}

	s3Session, err := s3Creds.GetBucket()
	if err != nil {
		return fmt.Errorf("S3 session failed: %w", err)
	}

	// Step 3: Upload ALL to S3 (track successes for rollback)
	srcset := []string{}
	uploadedKeys := []string{} // Track for rollback

	// Get ACL from config or use default
	acl := c.Bucket_Default_ACL
	if acl == "" {
		acl = "public-read" // Default for backwards compatibility
	}

	for _, fullPath := range *optimized {
		file, err := os.Open(fullPath)
		if err != nil {
			rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
			return fmt.Errorf("failed to open optimized file: %w", err)
		}

		filename := filepath.Base(fullPath)
		s3Key := fmt.Sprintf("%s/%d/%d/%s", bucketDir, year, month, filename)
		uploadPath := fmt.Sprintf("%s/%s", c.BucketEndpointURL(), s3Key)

		prep, err := bucket.UploadPrep(s3Key, c.Bucket_Media, file, acl)
		if err != nil {
			file.Close()
			rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
			return fmt.Errorf("upload prep failed: %w", err)
		}

		_, err = bucket.ObjectUpload(s3Session, prep)
		file.Close()
		if err != nil {
			rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
			return fmt.Errorf("S3 upload failed: %w", err)
		}

		uploadedKeys = append(uploadedKeys, s3Key)
		srcset = append(srcset, uploadPath)
	}

	// Step 4: All uploads succeeded - update database
	srcsetJSON, err := json.Marshal(srcset)
	if err != nil {
		rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
		return fmt.Errorf("failed to marshal srcset: %w", err)
	}

	params := MapMediaParams(row)
	params.Srcset = db.StringToNullString(string(srcsetJSON))

	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(c.Node_ID), types.UserID(""), "", "system")
	_, err = d.UpdateMedia(ctx, ac, params)
	if err != nil {
		rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
		return fmt.Errorf("database update failed: %w", err)
	}

	return nil
}

// rollbackS3Uploads deletes uploaded files from S3 on failure
func rollbackS3Uploads(s3Session *s3.S3, bucketName string, keys []string) {
	for _, key := range keys {
		_, err := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("rollback failed for key %s", key), err)
		} else {
			utility.DefaultLogger.Info(fmt.Sprintf("rolled back S3 upload: %s", key))
		}
	}
}

// MapMediaParams converts a Media record to UpdateMediaParams, updating the modification timestamp.
func MapMediaParams(a db.Media) db.UpdateMediaParams {
	return db.UpdateMediaParams{
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
		DateCreated:  a.DateCreated,
		DateModified: types.TimestampNow(),
		MediaID:      a.MediaID,
	}
}
