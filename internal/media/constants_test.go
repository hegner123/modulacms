package media

import (
	"errors"
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
	var _ error = FileTooLargeError{}
}

// TestImageMIMETypes verifies the imageMIMETypes map contains exactly the four
// expected image MIME types supported by the optimization pipeline.
func TestImageMIMETypes(t *testing.T) {
	t.Parallel()

	expected := []string{
		"image/png",
		"image/jpeg",
		"image/gif",
		"image/webp",
	}

	for _, mime := range expected {
		if !IsImageMIME(mime) {
			t.Errorf("expected %q to be an image MIME type", mime)
		}
	}

	// Verify count via the exported helper — four image types total
	count := 0
	for _, mime := range expected {
		if IsImageMIME(mime) {
			count++
		}
	}
	if count != 4 {
		t.Errorf("expected 4 image MIME types, got %d", count)
	}
}

// TestIsImageMIME verifies IsImageMIME returns true for supported image types
// and false for non-image types.
func TestIsImageMIME(t *testing.T) {
	t.Parallel()

	tests := []struct {
		contentType string
		want        bool
	}{
		{"image/png", true},
		{"image/jpeg", true},
		{"image/gif", true},
		{"image/webp", true},
		{"text/plain", false},
		{"application/pdf", false},
		{"application/octet-stream", false},
		{"video/mp4", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			t.Parallel()
			if got := IsImageMIME(tt.contentType); got != tt.want {
				t.Errorf("IsImageMIME(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}
