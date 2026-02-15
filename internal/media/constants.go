package media

// Media operation limits and configuration defaults.
const (
	// File size limits
	MaxUploadSize = 10 << 20 // 10 MB

	// Image dimension limits
	MaxImageWidth  = 10000    // 10k pixels
	MaxImageHeight = 10000    // 10k pixels
	MaxImagePixels = 50000000 // 50 megapixels

	// S3 configuration
	DefaultS3Region = "us-southeast-1"

	// Temp directory prefix
	TempDirPrefix = "modulacms-media"
)
