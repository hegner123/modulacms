package router

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// apiDownloadMedia generates a pre-signed S3 URL with Content-Disposition: attachment
// and redirects the client to it. The CMS never proxies file bytes.
func apiDownloadMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	rawID := r.PathValue("id")
	mediaID := types.MediaID(rawID)
	if err := mediaID.Validate(); err != nil {
		http.Error(w, "invalid media ID", http.StatusBadRequest)
		return
	}

	m, err := svc.Media.GetMedia(r.Context(), mediaID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	filename := filenameFromMedia(m)

	s3Key := extractS3Key(string(m.URL), *c)
	if s3Key == "" {
		http.Error(w, "unable to resolve storage key", http.StatusInternalServerError)
		return
	}

	creds := bucket.GetS3Creds(c)
	s3Client, err := creds.GetBucket()
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
		return
	}

	disposition := fmt.Sprintf(`attachment; filename="%s"`, sanitizeFilename(filename))

	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(c.Bucket_Media),
		Key:                        aws.String(s3Key),
		ResponseContentDisposition: aws.String(disposition),
	})

	presignedURL, err := req.Presign(15 * time.Minute)
	if err != nil {
		http.Error(w, "failed to generate download URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, presignedURL, http.StatusFound)
}

// filenameFromMedia returns the best available filename for download.
// Priority: display_name > name > last segment of URL.
func filenameFromMedia(m *db.Media) string {
	if m.DisplayName.Valid && m.DisplayName.String != "" {
		return m.DisplayName.String
	}
	if m.Name.Valid && m.Name.String != "" {
		return m.Name.String
	}
	u := string(m.URL)
	if idx := strings.LastIndex(u, "/"); idx >= 0 {
		return u[idx+1:]
	}
	return "download"
}

// extractS3Key strips the public URL prefix and bucket name to recover the S3 object key.
func extractS3Key(storedURL string, c config.Config) string {
	prefix := c.BucketPublicURL() + "/" + c.Bucket_Media + "/"
	if strings.HasPrefix(storedURL, prefix) {
		return storedURL[len(prefix):]
	}
	prefix = c.BucketEndpointURL() + "/" + c.Bucket_Media + "/"
	if strings.HasPrefix(storedURL, prefix) {
		return storedURL[len(prefix):]
	}
	return ""
}

// sanitizeFilename removes characters that are unsafe in Content-Disposition headers.
func sanitizeFilename(name string) string {
	r := strings.NewReplacer(`"`, "'", "\n", "", "\r", "")
	return r.Replace(name)
}
