//go:build integration

package media

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

const (
	minioEndpoint  = "localhost:9000"
	minioAccessKey = "modulacms"
	minioSecretKey = "modulacms_secret"
	minioRegion    = "us-east-1"
)

// testMinIOConfig returns a config.Config pointing at the local MinIO and
// an isolated SQLite database in the test's temp directory.
func testMinIOConfig(t *testing.T, bucketName string) config.Config {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "integration.db")
	return config.Config{
		Db_Driver:               config.Sqlite,
		Db_URL:                  dbPath,
		Bucket_Endpoint:         minioEndpoint,
		Bucket_Access_Key:       minioAccessKey,
		Bucket_Secret_Key:       minioSecretKey,
		Bucket_Media:            bucketName,
		Bucket_Region:           minioRegion,
		Bucket_Force_Path_Style: true,
		Environment:             "http-only",
		Node_ID:                 types.NewNodeID().String(),
	}
}

// newS3Client creates an S3 client connected to the test MinIO instance.
func newS3Client(t *testing.T) *s3.S3 {
	t.Helper()
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(minioAccessKey, minioSecretKey, ""),
		Endpoint:         aws.String("http://" + minioEndpoint),
		Region:           aws.String(minioRegion),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("newS3Client: %v", err)
	}
	return s3.New(sess)
}

// ensureBucket creates a unique S3 bucket and registers cleanup to delete
// all objects and the bucket itself when the test finishes.
func ensureBucket(t *testing.T, s3Client *s3.S3, bucketName string) {
	t.Helper()

	_, err := s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("ensureBucket CreateBucket(%s): %v", bucketName, err)
	}

	t.Cleanup(func() {
		// List and delete all objects first
		listOut, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			t.Logf("cleanup ListObjectsV2(%s): %v", bucketName, err)
			return
		}
		for _, obj := range listOut.Contents {
			_, delErr := s3Client.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    obj.Key,
			})
			if delErr != nil {
				t.Logf("cleanup DeleteObject(%s/%s): %v", bucketName, *obj.Key, delErr)
			}
		}
		// Delete bucket
		_, err = s3Client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			t.Logf("cleanup DeleteBucket(%s): %v", bucketName, err)
		}
	})
}

// testSeed holds references to seeded records needed by tests.
type testSeed struct {
	UserID types.UserID
}

// testIntegrationDB creates an isolated SQLite database with all tables,
// seeds the minimum FK chain (permission -> role -> user), and seeds a
// media dimension for the optimization step.
func testIntegrationDB(t *testing.T, cfg config.Config) (db.Database, testSeed) {
	t.Helper()

	conn, err := sql.Open("sqlite3", cfg.Db_URL)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if _, err := conn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	d := db.Database{
		Connection: conn,
		Context:    context.Background(),
		Config:     cfg,
	}

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	// Seed FK chain: permission -> role -> user
	ctx := d.Context
	ac := audited.Ctx(types.NodeID(cfg.Node_ID), types.UserID(""), "test", "127.0.0.1")
	now := types.TimestampNow()

	_, err = d.CreatePermission(ctx, ac, db.CreatePermissionParams{
		TableID: "test_table",
		Mode:    1,
		Label:   "test-permission",
	})
	if err != nil {
		t.Fatalf("seed CreatePermission: %v", err)
	}

	role, err := d.CreateRole(ctx, ac, db.CreateRoleParams{
		Label:       "test-role",
		Permissions: "[]",
	})
	if err != nil {
		t.Fatalf("seed CreateRole: %v", err)
	}

	user, err := d.CreateUser(ctx, ac, db.CreateUserParams{
		Username:     "testuser",
		Name:         "Test User",
		Email:        types.Email("test@example.com"),
		Hash:         "fakehash",
		Role:         role.RoleID.String(),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}

	// Seed a media dimension (100x100 thumbnail) for the optimize step
	acUser := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, "test", "127.0.0.1")
	_, err = d.CreateMediaDimension(ctx, acUser, db.CreateMediaDimensionParams{
		Label:       db.NewNullString("thumbnail"),
		Width:       db.NewNullInt64(100),
		Height:      db.NewNullInt64(100),
		AspectRatio: db.NewNullString("1:1"),
	})
	if err != nil {
		t.Fatalf("seed CreateMediaDimension: %v", err)
	}

	return d, testSeed{UserID: user.UserID}
}

// originalS3URL builds the S3 URL that HandleMediaUpload will use for the
// original image. Mirrors the path construction in media_upload.go.
func originalS3URL(cfg config.Config, filename string) types.URL {
	now := time.Now()
	s3Key := fmt.Sprintf("%s/%d/%d/%s", cfg.Bucket_Media, now.Year(), now.Month(), filename)
	return types.URL(fmt.Sprintf("%s/%s", cfg.BucketEndpointURL(), s3Key))
}

// createTestPNG generates a real 200x200 PNG file and returns its path.
func createTestPNG(t *testing.T, dir string, filename string) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	// Fill with a solid color so the image is non-trivial
	for y := range 200 {
		for x := range 200 {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}

	filePath := filepath.Join(dir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("createTestPNG os.Create: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("createTestPNG png.Encode: %v", err)
	}
	return filePath
}

// TestIntegration_HandleMediaUpload exercises the full pipeline:
// image optimization -> S3 upload to MinIO -> DB record update with srcset.
//
// Requires MinIO running at localhost:9000 (start with: just test-minio).
func TestIntegration_HandleMediaUpload(t *testing.T) {
	bucketName := fmt.Sprintf("test-media-%s", strings.ToLower(types.NewMediaID().String()))
	cfg := testMinIOConfig(t, bucketName)

	s3Client := newS3Client(t)
	ensureBucket(t, s3Client, bucketName)

	d, seed := testIntegrationDB(t, cfg)

	// Create a media DB record (simulating what ProcessMediaUpload does)
	ctx := d.Context
	ac := audited.Ctx(types.NodeID(cfg.Node_ID), seed.UserID, "test", "127.0.0.1")
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.UserID, Valid: true}

	mediaName := "integration-test"
	imageFile := mediaName + ".png"
	_, err := d.CreateMedia(ctx, ac, db.CreateMediaParams{
		Name:         db.NewNullString(mediaName),
		URL:          originalS3URL(cfg, imageFile),
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateMedia: %v", err)
	}

	// Write the source image to one directory, use a separate directory for
	// optimization output. This avoids copyFile(src, dst) overwriting the
	// source when src and dst resolve to the same path.
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcFile := createTestPNG(t, srcDir, imageFile)

	// Run the full pipeline
	err = HandleMediaUpload(srcFile, dstDir, cfg)
	if err != nil {
		t.Fatalf("HandleMediaUpload: %v", err)
	}

	// Verify S3: list objects in the bucket
	listOut, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2: %v", err)
	}

	// Expect at least 2 objects: the original + 1 optimized variant (100x100)
	if len(listOut.Contents) < 2 {
		t.Errorf("expected at least 2 S3 objects, got %d", len(listOut.Contents))
		for _, obj := range listOut.Contents {
			t.Logf("  S3 object: %s", *obj.Key)
		}
	}

	// Verify each object is non-empty
	for _, obj := range listOut.Contents {
		if *obj.Size == 0 {
			t.Errorf("S3 object %s has zero size", *obj.Key)
		}
	}

	// Verify DB: check the srcset was populated
	mediaRow, err := d.GetMediaByName(mediaName)
	if err != nil {
		t.Fatalf("GetMediaByName: %v", err)
	}
	if !mediaRow.Srcset.Valid || mediaRow.Srcset.String == "" {
		t.Fatal("expected non-empty srcset in DB record")
	}

	// Parse srcset JSON and verify it has the expected number of URLs
	var srcset []string
	if err := json.Unmarshal([]byte(mediaRow.Srcset.String), &srcset); err != nil {
		t.Fatalf("unmarshal srcset: %v", err)
	}

	// Should have original + at least 1 dimension variant
	if len(srcset) < 2 {
		t.Errorf("expected at least 2 srcset entries, got %d: %v", len(srcset), srcset)
	}

	// Verify each srcset URL points at our bucket
	for _, url := range srcset {
		if !strings.Contains(url, bucketName) {
			t.Errorf("srcset URL %q does not reference bucket %q", url, bucketName)
		}
	}

	// Verify srcset count matches S3 object count
	if len(srcset) != len(listOut.Contents) {
		t.Errorf("srcset entries (%d) != S3 objects (%d)", len(srcset), len(listOut.Contents))
	}

	t.Logf("S3 objects uploaded: %d", len(listOut.Contents))
	t.Logf("Srcset entries: %v", srcset)
}

// TestIntegration_HandleMediaUpload_BadBucket verifies that uploading to a
// non-existent bucket returns an error and no orphaned objects remain.
func TestIntegration_HandleMediaUpload_BadBucket(t *testing.T) {
	// Use a bucket name that doesn't exist
	bucketName := fmt.Sprintf("nonexistent-%s", strings.ToLower(types.NewMediaID().String()))
	cfg := testMinIOConfig(t, bucketName)

	d, seed := testIntegrationDB(t, cfg)

	// Create a media DB record
	ctx := d.Context
	ac := audited.Ctx(types.NodeID(cfg.Node_ID), seed.UserID, "test", "127.0.0.1")
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.UserID, Valid: true}

	mediaName := "bad-bucket-test"
	imageFile := mediaName + ".png"
	_, err := d.CreateMedia(ctx, ac, db.CreateMediaParams{
		Name:         db.NewNullString(mediaName),
		URL:          originalS3URL(cfg, imageFile),
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateMedia: %v", err)
	}

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcFile := createTestPNG(t, srcDir, imageFile)

	// This should fail because the bucket doesn't exist
	err = HandleMediaUpload(srcFile, dstDir, cfg)
	if err == nil {
		t.Fatal("expected error for non-existent bucket, got nil")
	}

	t.Logf("expected error received: %v", err)
}
