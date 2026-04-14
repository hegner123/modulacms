package bucket

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	config "github.com/hegner123/modulacms/internal/config"
)

func TestGetS3Creds(t *testing.T) {
	cfg := &config.Config{
		Bucket_Access_Key:       "test-access",
		Bucket_Secret_Key:       "test-secret",
		Bucket_Endpoint:         "minio:9000",
		Bucket_Region:           "eu-west-1",
		Bucket_Force_Path_Style: true,
		Bucket_Force_HTTP:       true,
		Environment:             config.EnvLocal,
	}

	creds := GetS3Creds(cfg)

	if creds.AccessKey != "test-access" {
		t.Errorf("AccessKey = %q, want %q", creds.AccessKey, "test-access")
	}
	if creds.SecretKey != "test-secret" {
		t.Errorf("SecretKey = %q, want %q", creds.SecretKey, "test-secret")
	}
	if creds.URL != "http://minio:9000" {
		t.Errorf("URL = %q, want %q", creds.URL, "http://minio:9000")
	}
	if creds.Region != "eu-west-1" {
		t.Errorf("Region = %q, want %q", creds.Region, "eu-west-1")
	}
	if !creds.ForcePathStyle {
		t.Error("ForcePathStyle = false, want true")
	}
}

func TestGetS3Creds_EmptyEndpoint(t *testing.T) {
	cfg := &config.Config{
		Bucket_Access_Key: "key",
		Bucket_Secret_Key: "secret",
		Bucket_Endpoint:   "",
		Environment:       config.EnvLocal,
	}

	creds := GetS3Creds(cfg)

	if creds.URL != "" {
		t.Errorf("URL = %q, want empty string for empty endpoint", creds.URL)
	}
}

func TestGetBucket_DefaultRegion(t *testing.T) {
	creds := S3Credentials{
		AccessKey:      "fake-key",
		SecretKey:      "fake-secret",
		URL:            "http://localhost:9000",
		Region:         "",
		ForcePathStyle: true,
	}

	svc, err := creds.GetBucket()
	if err != nil {
		t.Fatalf("GetBucket() error = %v", err)
	}
	if svc == nil {
		t.Fatal("GetBucket() returned nil client")
	}

	got := aws.StringValue(svc.Config.Region)
	if got != "us-east-1" {
		t.Errorf("default region = %q, want %q", got, "us-east-1")
	}
}

func TestGetBucket_ExplicitRegion(t *testing.T) {
	creds := S3Credentials{
		AccessKey:      "fake-key",
		SecretKey:      "fake-secret",
		URL:            "http://localhost:9000",
		Region:         "ap-southeast-1",
		ForcePathStyle: true,
	}

	svc, err := creds.GetBucket()
	if err != nil {
		t.Fatalf("GetBucket() error = %v", err)
	}

	got := aws.StringValue(svc.Config.Region)
	if got != "ap-southeast-1" {
		t.Errorf("region = %q, want %q", got, "ap-southeast-1")
	}
}

func TestGetBucket_ForcePathStyle(t *testing.T) {
	tests := []struct {
		name           string
		forcePathStyle bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := S3Credentials{
				AccessKey:      "fake-key",
				SecretKey:      "fake-secret",
				URL:            "http://localhost:9000",
				Region:         "us-east-1",
				ForcePathStyle: tt.forcePathStyle,
			}

			svc, err := creds.GetBucket()
			if err != nil {
				t.Fatalf("GetBucket() error = %v", err)
			}

			got := aws.BoolValue(svc.Config.S3ForcePathStyle)
			if got != tt.forcePathStyle {
				t.Errorf("S3ForcePathStyle = %v, want %v", got, tt.forcePathStyle)
			}
		})
	}
}

func TestGetBucket_Endpoint(t *testing.T) {
	creds := S3Credentials{
		AccessKey:      "fake-key",
		SecretKey:      "fake-secret",
		URL:            "http://custom-endpoint:9000",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	svc, err := creds.GetBucket()
	if err != nil {
		t.Fatalf("GetBucket() error = %v", err)
	}

	got := aws.StringValue(svc.Config.Endpoint)
	if got != "http://custom-endpoint:9000" {
		t.Errorf("Endpoint = %q, want %q", got, "http://custom-endpoint:9000")
	}
}

func TestUploadPrep(t *testing.T) {
	// Minimal valid 1x1 PNG (67 bytes) so the test is self-contained and
	// does not depend on fixture files that may be gitignored.
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, // 8-bit RGB
		0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54, // IDAT chunk
		0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00, 0x00,
		0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc, 0x33,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, // IEND chunk
		0xae, 0x42, 0x60, 0x82,
	}

	tmp := filepath.Join(t.TempDir(), "test.png")
	if err := os.WriteFile(tmp, pngData, 0o644); err != nil {
		t.Fatalf("write temp png: %v", err)
	}

	f, err := os.Open(tmp)
	if err != nil {
		t.Fatalf("open temp png: %v", err)
	}
	defer f.Close()

	expectedSize := int64(len(pngData))

	input, err := UploadPrep("media/images/photo.png", "my-bucket", f, "public-read")
	if err != nil {
		t.Fatalf("UploadPrep() error = %v", err)
	}

	if got := aws.StringValue(input.Bucket); got != "my-bucket" {
		t.Errorf("Bucket = %q, want %q", got, "my-bucket")
	}
	if got := aws.StringValue(input.Key); got != "media/images/photo.png" {
		t.Errorf("Key = %q, want %q", got, "media/images/photo.png")
	}
	if got := aws.StringValue(input.ACL); got != "public-read" {
		t.Errorf("ACL = %q, want %q", got, "public-read")
	}
	if got := aws.Int64Value(input.ContentLength); got != expectedSize {
		t.Errorf("ContentLength = %d, want %d", got, expectedSize)
	}
	if input.ContentType == nil {
		t.Error("ContentType is nil, want image/png")
	} else if got := aws.StringValue(input.ContentType); got != "image/png" {
		t.Errorf("ContentType = %q, want %q", got, "image/png")
	}
	if input.Body == nil {
		t.Error("Body is nil")
	}
}

func TestUploadPrep_ContentTypes(t *testing.T) {
	// Create a temp file to use as the data source.
	tmp, err := os.CreateTemp(t.TempDir(), "upload-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, writeErr := tmp.WriteString("test content"); writeErr != nil {
		t.Fatalf("write temp file: %v", writeErr)
	}
	if _, seekErr := tmp.Seek(0, 0); seekErr != nil {
		t.Fatalf("seek temp file: %v", seekErr)
	}
	defer tmp.Close()

	tests := []struct {
		uploadPath  string
		wantType    string
		wantTypeNil bool
	}{
		{"file.jpg", "image/jpeg", false},
		{"file.jpeg", "image/jpeg", false},
		{"file.css", "text/css; charset=utf-8", false},
		{"file.js", "text/javascript; charset=utf-8", false},
		{"file.html", "text/html; charset=utf-8", false},
		{"file.json", "application/json", false},
		{"file.pdf", "application/pdf", false},
		{"file.svg", "image/svg+xml", false},
		{"no-extension", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.uploadPath, func(t *testing.T) {
			if _, seekErr := tmp.Seek(0, 0); seekErr != nil {
				t.Fatalf("seek: %v", seekErr)
			}

			input, prepErr := UploadPrep(tt.uploadPath, "bucket", tmp, "private")
			if prepErr != nil {
				t.Fatalf("UploadPrep() error = %v", prepErr)
			}

			if tt.wantTypeNil {
				if input.ContentType != nil {
					t.Errorf("ContentType = %q, want nil", aws.StringValue(input.ContentType))
				}
				return
			}

			if input.ContentType == nil {
				t.Fatalf("ContentType is nil, want %q", tt.wantType)
			}
			if got := aws.StringValue(input.ContentType); got != tt.wantType {
				t.Errorf("ContentType = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestUploadPrep_ClosedFile(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "closed-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmp.Close()

	f, err := os.Open(tmp.Name())
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	// Close immediately so Stat still works but verifies the file handle
	// is in a usable state at the time of the call.
	f.Close()

	_, err = UploadPrep("path.png", "bucket", f, "private")
	if err == nil {
		t.Error("UploadPrep() with closed file: expected error, got nil")
	}
}
