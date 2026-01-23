package media

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	bucket "github.com/hegner123/modulacms/internal/bucket"
	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	utility "github.com/hegner123/modulacms/internal/utility"
)

//srcFile is the source file
//dstPath is the local destination for optimized files before uploading
func HandleMediaUpload(srcFile string, dstPath string, c config.Config) error {
	d := db.ConfigDB(c)
	bucketDir := c.Bucket_Media
	now := time.Now()
	year := now.Year()
	month := now.Month()

	filename := filepath.Base(srcFile)
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Step 1: Optimize images
	optimized, err := OptimizeUpload(srcFile, dstPath, c)
	if err != nil {
		return fmt.Errorf("optimization failed: %w", err)
	}

	// Step 2: Setup S3 session
	s3Creds := bucket.S3Credentials{
		AccessKey: c.Bucket_Access_Key,
		SecretKey: c.Bucket_Secret_Key,
		URL:       c.Bucket_Endpoint,
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
		uploadPath := fmt.Sprintf("https://%s/%s", c.Bucket_Endpoint, s3Key)

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

	utility.DefaultLogger.Debug(fmt.Sprintf("Fetching media record for: %s", baseName))
	rowPtr, err := d.GetMediaByName(baseName)
	if err != nil {
		rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
		return fmt.Errorf("failed to get media record: %w", err)
	}

	row := *rowPtr
	params := MapMediaParams(row)
	params.Srcset = db.StringToNullString(string(srcsetJSON))

	_, err = d.UpdateMedia(params)
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
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: types.TimestampNow(),
		MediaID:      a.MediaID,
	}
}
