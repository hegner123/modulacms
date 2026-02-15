package media

import (
	"errors"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Constants sanity checks
//
// These tests guard against accidental changes to constants that affect
// upload limits and S3 configuration. If someone changes MaxUploadSize from
// 10MB to 1MB, this test will catch it.
// ---------------------------------------------------------------------------

func TestConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      any
		want     any
		describe string
	}{
		{
			name:     "MaxUploadSize is 10MB",
			got:      MaxUploadSize,
			want:     int(10 << 20),
			describe: "10MB file size limit",
		},
		{
			name:     "MaxImageWidth is 10000",
			got:      MaxImageWidth,
			want:     10000,
			describe: "10k pixel width limit",
		},
		{
			name:     "MaxImageHeight is 10000",
			got:      MaxImageHeight,
			want:     10000,
			describe: "10k pixel height limit",
		},
		{
			name:     "MaxImagePixels is 50 million",
			got:      MaxImagePixels,
			want:     50000000,
			describe: "50 megapixel total limit",
		},
		{
			name:     "DefaultS3Region",
			got:      DefaultS3Region,
			want:     "us-southeast-1",
			describe: "default S3 region string",
		},
		{
			name:     "TempDirPrefix",
			got:      TempDirPrefix,
			want:     "modulacms-media",
			describe: "temp directory prefix for media processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.describe, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error type tests
// ---------------------------------------------------------------------------

// TestDuplicateMediaError verifies the error message format and that the error
// satisfies the error interface and is detectable via errors.As.
func TestDuplicateMediaError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      DuplicateMediaError
		wantMsg  string
		wantName string
	}{
		{
			name:     "standard filename",
			err:      DuplicateMediaError{Name: "photo.png"},
			wantMsg:  "duplicate entry found for photo.png",
			wantName: "photo.png",
		},
		{
			name:     "empty filename",
			err:      DuplicateMediaError{Name: ""},
			wantMsg:  "duplicate entry found for ",
			wantName: "",
		},
		{
			name:     "filename with spaces",
			err:      DuplicateMediaError{Name: "my photo (1).png"},
			wantMsg:  "duplicate entry found for my photo (1).png",
			wantName: "my photo (1).png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify Error() output
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}

			// Verify Name field is accessible after errors.As
			var target DuplicateMediaError
			if !errors.As(tt.err, &target) {
				t.Fatal("errors.As failed to match DuplicateMediaError")
			}
			if target.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", target.Name, tt.wantName)
			}
		})
	}
}

// TestInvalidMediaTypeError verifies the error message format includes the
// content type and the "Only images allowed" guidance.
func TestInvalidMediaTypeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         InvalidMediaTypeError
		wantContain []string
	}{
		{
			name:        "text/plain",
			err:         InvalidMediaTypeError{ContentType: "text/plain"},
			wantContain: []string{"text/plain", "Only images allowed"},
		},
		{
			name:        "application/pdf",
			err:         InvalidMediaTypeError{ContentType: "application/pdf"},
			wantContain: []string{"application/pdf", "Only images allowed"},
		},
		{
			name:        "empty content type",
			err:         InvalidMediaTypeError{ContentType: ""},
			wantContain: []string{"invalid file type", "Only images allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := tt.err.Error()
			for _, want := range tt.wantContain {
				if !strings.Contains(msg, want) {
					t.Errorf("Error() = %q, want it to contain %q", msg, want)
				}
			}

			// Verify errors.As detection
			var target InvalidMediaTypeError
			if !errors.As(tt.err, &target) {
				t.Fatal("errors.As failed to match InvalidMediaTypeError")
			}
			if target.ContentType != tt.err.ContentType {
				t.Errorf("ContentType = %q, want %q", target.ContentType, tt.err.ContentType)
			}
		})
	}
}

// TestFileTooLargeError verifies the error message includes both the actual
// file size and the maximum allowed size.
func TestFileTooLargeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     FileTooLargeError
		wantMsg string
	}{
		{
			name:    "slightly over limit",
			err:     FileTooLargeError{Size: MaxUploadSize + 1, MaxSize: MaxUploadSize},
			wantMsg: "file size 10485761 exceeds maximum 10485760",
		},
		{
			name:    "double the limit",
			err:     FileTooLargeError{Size: MaxUploadSize * 2, MaxSize: MaxUploadSize},
			wantMsg: "file size 20971520 exceeds maximum 10485760",
		},
		{
			name:    "zero size zero max",
			err:     FileTooLargeError{Size: 0, MaxSize: 0},
			wantMsg: "file size 0 exceeds maximum 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}

			// Verify errors.As and field access
			var target FileTooLargeError
			if !errors.As(tt.err, &target) {
				t.Fatal("errors.As failed to match FileTooLargeError")
			}
			if target.Size != tt.err.Size {
				t.Errorf("Size = %d, want %d", target.Size, tt.err.Size)
			}
			if target.MaxSize != tt.err.MaxSize {
				t.Errorf("MaxSize = %d, want %d", target.MaxSize, tt.err.MaxSize)
			}
		})
	}
}

// TestErrorTypes_ImplementErrorInterface is a compile-time guarantee that all
// three error types satisfy the error interface. The test function itself is
// a no-op -- the compiler does the work.
func TestErrorTypes_ImplementErrorInterface(t *testing.T) {
	t.Parallel()

	var _ error = DuplicateMediaError{}
	var _ error = InvalidMediaTypeError{}
	var _ error = FileTooLargeError{}
}

// TestValidMIMETypes verifies the validMIMETypes map contains exactly the four
// expected MIME types and no others. This guards against accidental additions
// or removals that could create security issues (allowing unexpected types) or
// break existing uploads (removing supported types).
func TestValidMIMETypes(t *testing.T) {
	t.Parallel()

	expected := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
		"image/webp": true,
	}

	// Check all expected types are present
	for mime := range expected {
		if !validMIMETypes[mime] {
			t.Errorf("expected MIME type %q to be in validMIMETypes, but it is missing", mime)
		}
	}

	// Check no unexpected types are present
	for mime := range validMIMETypes {
		if !expected[mime] {
			t.Errorf("unexpected MIME type %q found in validMIMETypes", mime)
		}
	}

	// Check the count matches
	if len(validMIMETypes) != len(expected) {
		t.Errorf("validMIMETypes has %d entries, want %d", len(validMIMETypes), len(expected))
	}
}
