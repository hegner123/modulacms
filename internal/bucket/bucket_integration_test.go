//go:build integration

package bucket

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	minioEndpoint  = "localhost:9000"
	minioAccessKey = "modula"
	minioSecretKey = "modula_secret"
	minioRegion    = "us-east-1"
)

// newTestS3Client creates an S3 client connected to the local MinIO instance.
func newTestS3Client(t *testing.T) *s3.S3 {
	t.Helper()
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(minioAccessKey, minioSecretKey, ""),
		Endpoint:         aws.String("http://" + minioEndpoint),
		Region:           aws.String(minioRegion),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("newTestS3Client: %v", err)
	}
	return s3.New(sess)
}

// uniqueBucketName returns a lowercase bucket name unique to the test.
func uniqueBucketName(t *testing.T, prefix string) string {
	t.Helper()
	// Use test name, sanitized to S3 bucket naming rules.
	name := fmt.Sprintf("%s-%d", prefix, os.Getpid())
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}

// cleanupBucket deletes all objects in the bucket, then deletes the bucket.
func cleanupBucket(t *testing.T, svc *s3.S3, bucketName string) {
	t.Helper()
	t.Cleanup(func() {
		listOut, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			t.Logf("cleanup ListObjectsV2(%s): %v", bucketName, err)
			return
		}
		for _, obj := range listOut.Contents {
			_, delErr := svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    obj.Key,
			})
			if delErr != nil {
				t.Logf("cleanup DeleteObject(%s/%s): %v", bucketName, *obj.Key, delErr)
			}
		}
		_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			t.Logf("cleanup DeleteBucket(%s): %v", bucketName, err)
		}
	})
}

// TestIntegration_GetBucket verifies that GetBucket returns a working S3 client
// by listing buckets on the MinIO instance.
func TestIntegration_GetBucket(t *testing.T) {
	creds := S3Credentials{
		AccessKey:      minioAccessKey,
		SecretKey:      minioSecretKey,
		URL:            "http://" + minioEndpoint,
		Region:         minioRegion,
		ForcePathStyle: true,
	}

	svc, err := creds.GetBucket()
	if err != nil {
		t.Fatalf("GetBucket() error = %v", err)
	}

	// Verify the client can talk to MinIO.
	out, err := svc.ListBuckets(nil)
	if err != nil {
		t.Fatalf("ListBuckets() error = %v", err)
	}
	if out == nil {
		t.Fatal("ListBuckets() returned nil output")
	}
}

// TestIntegration_EnsureBucket tests creating a new bucket and idempotent re-creation.
func TestIntegration_EnsureBucket(t *testing.T) {
	svc := newTestS3Client(t)
	bucketName := uniqueBucketName(t, "ensure")
	cleanupBucket(t, svc, bucketName)

	// First call creates the bucket.
	if err := EnsureBucket(svc, bucketName); err != nil {
		t.Fatalf("EnsureBucket() first call error = %v", err)
	}

	// Verify the bucket exists.
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("HeadBucket() after EnsureBucket: %v", err)
	}

	// Second call should be a no-op (bucket already exists).
	if err := EnsureBucket(svc, bucketName); err != nil {
		t.Fatalf("EnsureBucket() idempotent call error = %v", err)
	}
}

// TestIntegration_SetPublicReadPolicy verifies that a public-read policy
// is applied to a bucket.
func TestIntegration_SetPublicReadPolicy(t *testing.T) {
	svc := newTestS3Client(t)
	bucketName := uniqueBucketName(t, "policy")
	cleanupBucket(t, svc, bucketName)

	if err := EnsureBucket(svc, bucketName); err != nil {
		t.Fatalf("EnsureBucket(): %v", err)
	}

	if err := SetPublicReadPolicy(svc, bucketName); err != nil {
		t.Fatalf("SetPublicReadPolicy() error = %v", err)
	}

	// Verify the policy was set by reading it back.
	policyOut, err := svc.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("GetBucketPolicy() error = %v", err)
	}
	if policyOut.Policy == nil || *policyOut.Policy == "" {
		t.Fatal("GetBucketPolicy() returned empty policy")
	}

	policy := *policyOut.Policy
	if !strings.Contains(policy, "s3:GetObject") {
		t.Errorf("policy missing s3:GetObject action: %s", policy)
	}
	if !strings.Contains(policy, bucketName) {
		t.Errorf("policy missing bucket name %q: %s", bucketName, policy)
	}
	if !strings.Contains(policy, `"Effect":"Allow"`) {
		t.Errorf("policy missing Allow effect: %s", policy)
	}
}

// TestIntegration_ObjectUpload verifies the full upload flow:
// UploadPrep -> ObjectUpload -> verify object exists in S3.
func TestIntegration_ObjectUpload(t *testing.T) {
	svc := newTestS3Client(t)
	bucketName := uniqueBucketName(t, "upload")
	cleanupBucket(t, svc, bucketName)

	if err := EnsureBucket(svc, bucketName); err != nil {
		t.Fatalf("EnsureBucket(): %v", err)
	}

	// Create a temp file with known content.
	tmp, err := os.CreateTemp(t.TempDir(), "upload-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	content := "integration test content"
	if _, writeErr := tmp.WriteString(content); writeErr != nil {
		t.Fatalf("WriteString: %v", writeErr)
	}
	if _, seekErr := tmp.Seek(0, 0); seekErr != nil {
		t.Fatalf("Seek: %v", seekErr)
	}
	defer tmp.Close()

	// Prepare and execute upload.
	uploadKey := "test/upload.txt"
	input, err := UploadPrep(uploadKey, bucketName, tmp, "public-read")
	if err != nil {
		t.Fatalf("UploadPrep() error = %v", err)
	}

	_, err = ObjectUpload(svc, input)
	if err != nil {
		t.Fatalf("ObjectUpload() error = %v", err)
	}

	// Verify the object exists and has the right size.
	headOut, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uploadKey),
	})
	if err != nil {
		t.Fatalf("HeadObject() error = %v", err)
	}

	expectedSize := int64(len(content))
	if got := aws.Int64Value(headOut.ContentLength); got != expectedSize {
		t.Errorf("ContentLength = %d, want %d", got, expectedSize)
	}
}

// TestIntegration_ObjectUpload_PNG verifies upload with content-type detection.
func TestIntegration_ObjectUpload_PNG(t *testing.T) {
	svc := newTestS3Client(t)
	bucketName := uniqueBucketName(t, "upload-png")
	cleanupBucket(t, svc, bucketName)

	if err := EnsureBucket(svc, bucketName); err != nil {
		t.Fatalf("EnsureBucket(): %v", err)
	}

	// Create a temp file with PNG-like content.
	tmp, err := os.CreateTemp(t.TempDir(), "upload-*.png")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, writeErr := tmp.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}); writeErr != nil {
		t.Fatalf("Write: %v", writeErr)
	}
	if _, seekErr := tmp.Seek(0, 0); seekErr != nil {
		t.Fatalf("Seek: %v", seekErr)
	}
	defer tmp.Close()

	uploadKey := "images/test.png"
	input, err := UploadPrep(uploadKey, bucketName, tmp, "public-read")
	if err != nil {
		t.Fatalf("UploadPrep() error = %v", err)
	}

	_, err = ObjectUpload(svc, input)
	if err != nil {
		t.Fatalf("ObjectUpload() error = %v", err)
	}

	headOut, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uploadKey),
	})
	if err != nil {
		t.Fatalf("HeadObject() error = %v", err)
	}

	// Verify content type was set from the upload key extension.
	if got := aws.StringValue(headOut.ContentType); got != "image/png" {
		t.Errorf("ContentType = %q, want %q", got, "image/png")
	}
}

// TestIntegration_PrintBuckets verifies PrintBuckets does not panic.
// PrintBuckets logs output via utility.DefaultLogger, so we verify it completes
// without error rather than capturing stdout.
func TestIntegration_PrintBuckets(t *testing.T) {
	svc := newTestS3Client(t)

	// Create a bucket so there's at least one to list.
	bucketName := uniqueBucketName(t, "print")
	cleanupBucket(t, svc, bucketName)

	if err := EnsureBucket(svc, bucketName); err != nil {
		t.Fatalf("EnsureBucket(): %v", err)
	}

	// PrintBuckets calls Fatal on error, so reaching the next line means success.
	PrintBuckets(svc)
}
