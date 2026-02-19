package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediasHandler handles CRUD operations that do not require a specific media ID.
func MediasHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListMediaPaginated(w, r, c)
		} else {
			apiListMedia(w, c)
		}
	case http.MethodPost:
		apiCreateMedia(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MediaHandler handles CRUD operations for specific media items.
func MediaHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetMedia(w, r, c)
	case http.MethodPut:
		apiUpdateMedia(w, r, c)
	case http.MethodDelete:
		apiDeleteMedia(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetMedia handles GET requests for a single media item
func apiGetMedia(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	mID := types.MediaID(q)
	if err := mID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	media, err := d.GetMedia(mID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(media)
	return nil
}

// apiListMedia handles GET requests for listing media items
func apiListMedia(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	mediaList, err := d.ListMedia()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mediaList)
	return nil
}

// apiCreateMedia handles POST requests to upload and create a new media item.
// Accepts multipart form with a "file" field. Delegates validation, S3 upload,
// DB creation, and optimization pipeline to media.ProcessMediaUpload.
func apiCreateMedia(w http.ResponseWriter, r *http.Request, c config.Config) {
	err := r.ParseMultipartForm(c.MaxUploadSize())
	if err != nil {
		utility.DefaultLogger.Error("parse form", err)
		http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utility.DefaultLogger.Error("parse file", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	// Set up S3 session once for both uploadOriginal and rollback
	s3Creds := bucket.S3Credentials{
		AccessKey:      c.Bucket_Access_Key,
		SecretKey:      c.Bucket_Secret_Key,
		URL:            c.BucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}

	s3Session, err := s3Creds.GetBucket()
	if err != nil {
		utility.DefaultLogger.Error("S3 session", err)
		http.Error(w, "Storage service unavailable", http.StatusInternalServerError)
		return
	}

	// Read optional path parameter for S3 key organization
	mediaPath, pathErr := media.SanitizeMediaPath(r.PostFormValue("path"))
	if pathErr != nil {
		utility.DefaultLogger.Error("invalid media path", pathErr)
		http.Error(w, pathErr.Error(), http.StatusBadRequest)
		return
	}

	acl := c.Bucket_Default_ACL
	if acl == "" {
		acl = "public-read"
	}

	bucketDir := c.Bucket_Media

	uploadOriginal := func(filePath string) (string, string, error) {
		f, err := os.Open(filePath)
		if err != nil {
			return "", "", fmt.Errorf("open file for S3 upload: %w", err)
		}
		defer f.Close()

		filename := filepath.Base(filePath)
		s3Key := fmt.Sprintf("%s/%s", mediaPath, filename)
		uploadURL := fmt.Sprintf("%s/%s/%s", c.BucketPublicURL(), bucketDir, s3Key)

		prep, err := bucket.UploadPrep(s3Key, c.Bucket_Media, f, acl)
		if err != nil {
			return "", "", fmt.Errorf("upload prep: %w", err)
		}

		_, err = bucket.ObjectUpload(s3Session, prep)
		if err != nil {
			return "", "", fmt.Errorf("S3 upload: %w", err)
		}

		return uploadURL, s3Key, nil
	}

	rollbackS3 := func(s3Key string) {
		_, err := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(c.Bucket_Media),
			Key:    aws.String(s3Key),
		})
		if err != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("rollback failed for key %s", s3Key), err)
		} else {
			utility.DefaultLogger.Info(fmt.Sprintf("rolled back S3 upload: %s", s3Key))
		}
	}

	pipeline := func(srcFile string, dstPath string) error {
		return media.HandleMediaUpload(srcFile, dstPath, c)
	}

	row, err := media.ProcessMediaUpload(r.Context(), ac, file, header, d, uploadOriginal, rollbackS3, pipeline, c.MaxUploadSize())
	if err != nil {
		var dupErr media.DuplicateMediaError
		var sizeErr media.FileTooLargeError

		switch {
		case errors.As(err, &dupErr):
			utility.DefaultLogger.Error("duplicate media", err)
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.As(err, &sizeErr):
			utility.DefaultLogger.Error("file too large", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			utility.DefaultLogger.Error("create media", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(row)
}

// apiUpdateMedia handles PUT requests to update an existing media item
func apiUpdateMedia(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateMedia db.UpdateMediaParams
	err := json.NewDecoder(r.Body).Decode(&updateMedia)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateMedia(r.Context(), ac, updateMedia)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetMedia(updateMedia.MediaID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated media", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteMedia handles DELETE requests for media items.
// Deletes all S3 objects (original + optimized variants) then removes the DB record.
func apiDeleteMedia(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	mID := types.MediaID(q)
	if err := mID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid media ID", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Fetch record to get S3 keys before deleting
	record, err := d.GetMedia(mID)
	if err != nil {
		utility.DefaultLogger.Error("media not found", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return err
	}

	// Collect all S3 keys: original URL + srcset variants
	// URL format: {BucketPublicURL}/{Bucket_Media}/{year}/{month}/{file}
	// S3 key format: {year}/{month}/{file} (bucket name is separate)
	endpointPrefix := c.BucketPublicURL() + "/" + c.Bucket_Media + "/"
	var s3Keys []string

	if string(record.URL) != "" {
		key := strings.TrimPrefix(string(record.URL), endpointPrefix)
		s3Keys = append(s3Keys, key)
	}

	if record.Srcset.Valid && record.Srcset.String != "" {
		var srcsetURLs []string
		if jsonErr := json.Unmarshal([]byte(record.Srcset.String), &srcsetURLs); jsonErr == nil {
			for _, u := range srcsetURLs {
				key := strings.TrimPrefix(u, endpointPrefix)
				s3Keys = append(s3Keys, key)
			}
		}
	}

	// Delete S3 objects
	if len(s3Keys) > 0 {
		s3Creds := bucket.S3Credentials{
			AccessKey:      c.Bucket_Access_Key,
			SecretKey:      c.Bucket_Secret_Key,
			URL:            c.BucketEndpointURL(),
			Region:         c.Bucket_Region,
			ForcePathStyle: c.Bucket_Force_Path_Style,
		}

		s3Session, s3Err := s3Creds.GetBucket()
		if s3Err != nil {
			utility.DefaultLogger.Error("S3 session for media delete", s3Err)
			http.Error(w, "Storage service unavailable", http.StatusInternalServerError)
			return s3Err
		}

		for _, key := range s3Keys {
			_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(c.Bucket_Media),
				Key:    aws.String(key),
			})
			if delErr != nil {
				utility.DefaultLogger.Warn("failed to delete S3 object", delErr, "key", key)
			}
		}
	}

	// Delete DB record
	ac := middleware.AuditContextFromRequest(r, c)
	err = d.DeleteMedia(r.Context(), ac, mID)
	if err != nil {
		utility.DefaultLogger.Error("delete media record", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// mediaOrphanResult holds the result of scanning for orphaned S3 objects.
type mediaOrphanResult struct {
	TotalObjects int
	TrackedKeys  int
	OrphanedKeys []string
}

// findOrphanedMediaKeys compares all S3 objects in the media bucket against
// URLs and srcset entries in the database, returning any untracked keys.
func findOrphanedMediaKeys(d db.DbDriver, s3Session *s3.S3, c config.Config) (*mediaOrphanResult, error) {
	mediaList, err := d.ListMedia()
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

	return &mediaOrphanResult{
		TotalObjects: totalObjects,
		TrackedKeys:  len(knownKeys),
		OrphanedKeys: orphanedKeys,
	}, nil
}

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

// MediaHealthHandler checks for orphaned files in the media S3 bucket that have
// no corresponding database record.
func MediaHealthHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	s3Session, err := newMediaS3Session(c)
	if err != nil {
		utility.DefaultLogger.Error("S3 session for health check", err)
		http.Error(w, "Storage service unavailable", http.StatusInternalServerError)
		return
	}

	result, err := findOrphanedMediaKeys(d, s3Session, c)
	if err != nil {
		utility.DefaultLogger.Error("media health check", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		TotalObjects int      `json:"total_objects"`
		TrackedKeys  int      `json:"tracked_keys"`
		OrphanedKeys []string `json:"orphaned_keys"`
		OrphanCount  int      `json:"orphan_count"`
	}{
		TotalObjects: result.TotalObjects,
		TrackedKeys:  result.TrackedKeys,
		OrphanedKeys: result.OrphanedKeys,
		OrphanCount:  len(result.OrphanedKeys),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MediaCleanupHandler deletes orphaned files from the media S3 bucket that have
// no corresponding database record.
func MediaCleanupHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	s3Session, err := newMediaS3Session(c)
	if err != nil {
		utility.DefaultLogger.Error("S3 session for cleanup", err)
		http.Error(w, "Storage service unavailable", http.StatusInternalServerError)
		return
	}

	result, err := findOrphanedMediaKeys(d, s3Session, c)
	if err != nil {
		utility.DefaultLogger.Error("media cleanup scan", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var deleted []string
	var failed []string

	for _, key := range result.OrphanedKeys {
		_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(c.Bucket_Media),
			Key:    aws.String(key),
		})
		if delErr != nil {
			utility.DefaultLogger.Warn("failed to delete orphaned object", delErr, "key", key)
			failed = append(failed, key)
		} else {
			deleted = append(deleted, key)
		}
	}

	response := struct {
		Deleted      []string `json:"deleted"`
		DeletedCount int      `json:"deleted_count"`
		Failed       []string `json:"failed"`
		FailedCount  int      `json:"failed_count"`
	}{
		Deleted:      deleted,
		DeletedCount: len(deleted),
		Failed:       failed,
		FailedCount:  len(failed),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// apiListMediaPaginated handles GET requests for listing media with pagination.
func apiListMediaPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListMediaPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountMedia()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.Media]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}
